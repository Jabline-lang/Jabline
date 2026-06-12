package vm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DebugSession holds the debugger's runtime state.
type DebugSession struct {
	Breakpoints map[int]bool // set of source line numbers with active breakpoints
	vm          *VM
	sourceLines []string // raw lines of the source file for contextual printing
	scanner     *bufio.Scanner
	stepping    bool // if true, pause on every new line
	lastLine    int  // last source line seen, to avoid re-pausing on same line

	// DAP integration fields (set by DAP server, nil in REPL mode)
	OnPause func(line int)        // called when debugger pauses
	PauseCh chan struct{}         // nil in REPL mode; when set, OnInstruction waits here for resume
	ResumeCh chan DebugAction     // channel to signal resume action
	PauseRequested bool           // set to true to pause on next instruction
}

func NewDebugSession(vm *VM, source string) *DebugSession {
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	return &DebugSession{
		Breakpoints: make(map[int]bool),
		vm:          vm,
		sourceLines: lines,
		scanner:     bufio.NewScanner(os.Stdin),
		stepping:    true, // start in step mode
		lastLine:    -1,
	}
}

// OnInstruction is called by the VM on every instruction cycle.
// It checks if we should pause (step or breakpoint) and launches the REPL.
func (ds *DebugSession) SetStepping(v bool) {
	ds.stepping = v
}

func (ds *DebugSession) OnInstruction(vm *VM) {
	currentLine := ds.CurrentLine()
	if currentLine == 0 {
		return
	}

	shouldPause := false
	if ds.stepping && currentLine != ds.lastLine {
		shouldPause = true
	}
	if ds.Breakpoints[currentLine] {
		shouldPause = true
	}
	if ds.PauseRequested {
		shouldPause = true
		ds.PauseRequested = false
	}

	if !shouldPause {
		return
	}

	ds.lastLine = currentLine

	if ds.PauseCh != nil && ds.ResumeCh != nil {
		// DAP mode: notify and wait for client action
		if ds.OnPause != nil {
			ds.OnPause(currentLine)
		}
		// Signal that we paused
		ds.PauseCh <- struct{}{}
		// Wait for resume action
		action := <-ds.ResumeCh
		switch action {
		case ActionStep:
			ds.stepping = true
		case ActionContinue:
			ds.stepping = false
		}
	} else {
		// REPL mode
		action := ds.REPL(ds.scanner)
		switch action {
		case ActionStep:
			ds.stepping = true
		case ActionContinue:
			ds.stepping = false
		}
	}
}

// CurrentLine returns the source line number at the current instruction pointer.
func (ds *DebugSession) CurrentLine() int {
	frm := ds.vm.currentFrame()
	if frm.cl.Fn.SourceMap == nil {
		return 0
	}
	pos, ok := frm.cl.Fn.SourceMap[frm.ip]
	if !ok {
		return 0
	}
	return pos.Line
}

// PrintContext prints the source code around the current line.
func (ds *DebugSession) PrintContext(window int) {
	line := ds.CurrentLine()
	if line == 0 || len(ds.sourceLines) == 0 {
		fmt.Println("[debugger] No source context available.")
		return
	}

	start := line - window - 1
	if start < 0 {
		start = 0
	}
	end := line + window
	if end > len(ds.sourceLines) {
		end = len(ds.sourceLines)
	}

	for i := start; i < end; i++ {
		lineNum := i + 1
		marker := "   "
		if lineNum == line {
			marker = "→  "
		}
		if ds.Breakpoints[lineNum] {
			marker = " ● "
		}
		if lineNum == line && ds.Breakpoints[lineNum] {
			marker = "→● "
		}
		fmt.Printf("\033[90m%3d\033[0m %s %s\n", lineNum, marker, ds.sourceLines[i])
	}
}

// PrintLocals prints variables in the current stack frame.
func (ds *DebugSession) PrintLocals() {
	frm := ds.vm.currentFrame()
	base := frm.basePointer
	top := ds.vm.sp

	if top <= base {
		fmt.Println("[locals] (vacío)")
		return
	}

	fmt.Println("[locals]")
	for i := base; i < top; i++ {
		obj := ds.vm.stack[i]
		if obj != nil {
			fmt.Printf("  [%d] %s = %s\n", i-base, obj.Type(), obj.Inspect())
		}
	}
}

// PrintStack prints the current call stack.
func (ds *DebugSession) PrintStack() {
	fmt.Println("[call stack]")
	for i := ds.vm.framesIndex - 1; i >= 0; i-- {
		frm := ds.vm.frames[i]
		name := "<main>"
		if frm.cl != nil && frm.cl.Fn.Name != "" {
			name = frm.cl.Fn.Name
		}
		line := 0
		if frm.cl.Fn.SourceMap != nil {
			if pos, ok := frm.cl.Fn.SourceMap[frm.ip]; ok {
				line = pos.Line
			}
		}
		fmt.Printf("  #%d  def %s()  línea %d\n", i, name, line)
	}
}

// REPL starts the interactive debugger prompt (called when a breakpoint is hit or stepping).
func (ds *DebugSession) REPL(scanner *bufio.Scanner) DebugAction {
	ds.PrintContext(3)
	fmt.Printf("\n\033[36m(jabline-debug) \033[0m")

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(input)
		if len(parts) == 0 {
			fmt.Printf("\033[36m(jabline-debug) \033[0m")
			continue
		}

		switch parts[0] {
		case "n", "next", "s", "step":
			return ActionStep

		case "c", "continue":
			return ActionContinue

		case "b", "break":
			if len(parts) < 2 {
				fmt.Println("Uso: b <número_de_línea>")
			} else {
				lineNum, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Println("Número de línea inválido.")
				} else {
					ds.Breakpoints[lineNum] = true
					fmt.Printf("Breakpoint añadido en línea %d\n", lineNum)
				}
			}

		case "d", "delete":
			if len(parts) < 2 {
				fmt.Println("Uso: d <número_de_línea>")
			} else {
				lineNum, _ := strconv.Atoi(parts[1])
				delete(ds.Breakpoints, lineNum)
				fmt.Printf("Breakpoint en línea %d eliminado\n", lineNum)
			}

		case "p", "print":
			// Simply dump locals for now; a named lookup would require symbol table access
			ds.PrintLocals()

		case "locals":
			ds.PrintLocals()

		case "stack":
			ds.PrintStack()

		case "list":
			ds.PrintContext(5)

		case "q", "quit", "exit":
			fmt.Println("Saliendo del depurador.")
			os.Exit(0)

		case "h", "help":
			fmt.Println(debugHelp)

		default:
			fmt.Printf("Comando desconocido: '%s'. Escribe 'h' para ayuda.\n", parts[0])
		}

		fmt.Printf("\033[36m(jabline-debug) \033[0m")
	}
	return ActionContinue
}

// DebugAction signals to the Run loop what to do next.
type DebugAction int

const (
	ActionStep     DebugAction = iota
	ActionContinue DebugAction = iota
)

const debugHelp = `Comandos disponibles:
  n / next      — Ejecutar la siguiente instrucción (step)
  c / continue  — Continuar hasta el próximo breakpoint
  b <línea>     — Poner breakpoint en la línea indicada
  d <línea>     — Eliminar breakpoint en la línea indicada
  p / locals    — Ver variables locales del frame actual
  stack         — Ver el call stack
  list          — Ver el código fuente alrededor de la línea actual
  q / quit      — Salir del depurador
  h / help      — Mostrar esta ayuda`
