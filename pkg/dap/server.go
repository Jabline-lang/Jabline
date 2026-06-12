package dap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"jabline/pkg/symbol"
	"jabline/pkg/vm"
)

// DAP message types
type Message struct {
	Seq         int64            `json:"seq"`
	Type        string           `json:"type"`
	Command     string           `json:"command,omitempty"`
	RequestSeq  int64            `json:"request_seq,omitempty"`
	Success     bool             `json:"success,omitempty"`
	Event       string           `json:"event,omitempty"`
	Body        json.RawMessage  `json:"body,omitempty"`
	Arguments   json.RawMessage  `json:"arguments,omitempty"`
	Message     string           `json:"message,omitempty"`
}

// DebugAdapter bridges DAP with the Jabline VM debugger
type DebugAdapter struct {
	reader      *bufio.Reader
	writer      io.Writer
	seq         int64
	vm          *vm.VM
	debugSess   *vm.DebugSession
	source      string        // source code content
	sourcePath  string        // original file path
	sourceLines []string
	breakpoints map[int]int // line -> breakpointId
	running     bool
	done        chan struct{}

	symbolTable   *symbol.SymbolTable // top-level symbol table from compiler
	pauseFlag     int32               // atomic flag for pause request (fallback)
	lastException error               // last VM exception caught
}

func NewDebugAdapter(r io.Reader, w io.Writer) *DebugAdapter {
	return &DebugAdapter{
		reader:      bufio.NewReader(r),
		writer:      w,
		breakpoints: make(map[int]int),
		done:        make(chan struct{}),
	}
}

func (da *DebugAdapter) Run() error {
	for {
		msg, err := da.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		switch msg.Type {
		case "request":
			da.handleRequest(msg)
		}
	}
}

func (da *DebugAdapter) nextSeq() int64 {
	da.seq++
	return da.seq
}

func (da *DebugAdapter) sendMessage(msg Message) {
	data, _ := json.Marshal(msg)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	da.writer.Write([]byte(header))
	da.writer.Write(data)
}

func (da *DebugAdapter) sendResponse(request Message, body interface{}) {
	var raw json.RawMessage
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	da.sendMessage(Message{
		Type:       "response",
		RequestSeq: request.Seq,
		Command:    request.Command,
		Success:    true,
		Seq:        da.nextSeq(),
		Body:       raw,
	})
}

func (da *DebugAdapter) sendErrorResponse(request Message, errMsg string) {
	da.sendMessage(Message{
		Type:       "response",
		RequestSeq: request.Seq,
		Command:    request.Command,
		Success:    false,
		Message:    errMsg,
		Seq:        da.nextSeq(),
	})
}

func (da *DebugAdapter) sendEvent(event string, body interface{}) {
	var raw json.RawMessage
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	da.sendMessage(Message{
		Type:  "event",
		Event: event,
		Seq:   da.nextSeq(),
		Body:  raw,
	})
}

func (da *DebugAdapter) sendOutput(category, output string) {
	da.sendEvent("output", map[string]interface{}{
		"category": category,
		"output":   output,
	})
}

func (da *DebugAdapter) readMessage() (Message, error) {
	// Read Content-Length header
	line, err := da.reader.ReadString('\n')
	if err != nil {
		return Message{}, err
	}
	line = strings.TrimSpace(line)

	var length int
	if strings.HasPrefix(line, "Content-Length:") {
		length, err = strconv.Atoi(strings.TrimSpace(line[15:]))
		if err != nil {
			return Message{}, fmt.Errorf("invalid content length: %s", line)
		}
	} else {
		return Message{}, fmt.Errorf("expected Content-Length header, got: %s", line)
	}

	// Read blank line
	da.reader.ReadString('\n')

	// Read JSON body
	body := make([]byte, length)
	_, err = io.ReadFull(da.reader, body)
	if err != nil {
		return Message{}, fmt.Errorf("read body error: %w", err)
	}

	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return Message{}, fmt.Errorf("parse error: %w", err)
	}
	return msg, nil
}

func (da *DebugAdapter) handleRequest(msg Message) {
	switch msg.Command {
	case "initialize":
		da.handleInitialize(msg)
	case "launch":
		da.handleLaunch(msg)
	case "setBreakpoints":
		da.handleSetBreakpoints(msg)
	case "setExceptionBreakpoints":
		da.handleSetExceptionBreakpoints(msg)
	case "configurationDone":
		da.handleConfigurationDone(msg)
	case "threads":
		da.handleThreads(msg)
	case "stackTrace":
		da.handleStackTrace(msg)
	case "scopes":
		da.handleScopes(msg)
	case "variables":
		da.handleVariables(msg)
	case "continue":
		da.handleContinue(msg)
	case "next":
		da.handleNext(msg)
	case "stepIn":
		da.handleStepIn(msg)
	case "stepOut":
		da.handleStepOut(msg)
	case "pause":
		da.handlePause(msg)
	case "evaluate":
		da.handleEvaluate(msg)
	case "exceptionInfo":
		da.handleExceptionInfo(msg)
	case "disconnect":
		da.handleDisconnect(msg)
	default:
		da.sendResponse(msg, nil)
	}
}

func (da *DebugAdapter) handleInitialize(msg Message) {
	da.sendResponse(msg, map[string]interface{}{
		"supportsConfigurationDoneRequest": true,
		"supportsSetVariable":              false,
		"supportsConditionalBreakpoints":   false,
		"supportsHitConditionalBreakpoints": false,
		"supportsEvaluateForHovers":        true,
		"supportsStepBack":                 false,
		"supportsRestartFrame":             false,
		"supportsSteppingGranularity":      false,
		"supportsGotoTargetsRequest":       false,
		"supportsCompletionsRequest":       false,
		"supportsBreakpointLocationsRequest": false,
		"supportsTerminateRequest":         true,
	})
}

func (da *DebugAdapter) handleLaunch(msg Message) {
	var args struct {
		Program string `json:"program"`
		Args    []string `json:"args"`
	}
	if msg.Arguments != nil {
		json.Unmarshal(msg.Arguments, &args)
	}

	programPath := args.Program
	if programPath == "" {
		da.sendErrorResponse(msg, "missing 'program' argument")
		return
	}

	source, err := os.ReadFile(programPath)
	if err != nil {
		da.sendErrorResponse(msg, fmt.Sprintf("cannot read program: %s", err))
		return
	}
	da.source = string(source)
	da.sourcePath = programPath
	da.sourceLines = strings.Split(strings.ReplaceAll(da.source, "\r\n", "\n"), "\n")

	l := lexer.New(da.source)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		da.sendErrorResponse(msg, fmt.Sprintf("parse errors: %v", p.Errors()))
		return
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		da.sendErrorResponse(msg, fmt.Sprintf("compile error: %s", err))
		return
	}

	bc := comp.Bytecode()
	machine := vm.New(bc.Instructions, bc.Constants, programPath)
	da.vm = machine
	da.symbolTable = bc.SymbolTable

	// Create debug session in DAP mode
	sess := vm.NewDebugSession(machine, da.source)
	sess.SetStepping(false) // don't step without explicit DAP request
	sess.PauseCh = make(chan struct{}, 1)
	sess.ResumeCh = make(chan vm.DebugAction, 1)
	sess.OnPause = func(line int) {
		da.sendEvent("stopped", map[string]interface{}{
			"reason":           "breakpoint",
			"threadId":         1,
			"text":             fmt.Sprintf("paused at line %d", line),
			"allThreadsStopped": true,
		})
	}
	da.debugSess = sess
	machine.Debug = sess

	// Set breakpoints from initial config
	for line := range da.breakpoints {
		sess.Breakpoints[line] = true
	}

	da.sendResponse(msg, nil)
	da.sendEvent("initialized", nil)
}

func (da *DebugAdapter) handleSetBreakpoints(msg Message) {
	var args struct {
		Source struct {
			Path string `json:"path"`
		} `json:"source"`
		Breakpoints []struct {
			Line int `json:"line"`
		} `json:"breakpoints"`
	}
	if msg.Arguments != nil {
		json.Unmarshal(msg.Arguments, &args)
	}

	da.breakpoints = make(map[int]int)
	for i, bp := range args.Breakpoints {
		da.breakpoints[bp.Line] = i + 1
		if da.debugSess != nil {
			da.debugSess.Breakpoints[bp.Line] = true
		}
	}

	// Build verified breakpoints response
	bps := make([]map[string]interface{}, len(args.Breakpoints))
	for i, bp := range args.Breakpoints {
		bps[i] = map[string]interface{}{
			"id":       i + 1,
			"verified": true,
			"line":     bp.Line,
		}
	}

	da.sendResponse(msg, map[string]interface{}{
		"breakpoints": bps,
	})
}

func (da *DebugAdapter) handleSetExceptionBreakpoints(msg Message) {
	da.sendResponse(msg, nil)
}

func (da *DebugAdapter) handleConfigurationDone(msg Message) {
	da.sendResponse(msg, nil)
	if da.vm != nil && !da.running {
		da.running = true
		go da.runProgram()
	}
}

func (da *DebugAdapter) runProgram() {
	// Hook debug session into VM execution
	// The VM runs in a goroutine, debug session pauses via channels
	da.sendOutput("stdout", "Program started\n")

	err := da.vm.Run()
	if err != nil {
		da.lastException = err
		errMsg := err.Error()
		// Check if this was a user-requested pause via context cancellation
		if strings.Contains(errMsg, "cancelled") && atomic.LoadInt32(&da.pauseFlag) != 0 {
			atomic.StoreInt32(&da.pauseFlag, 0)
			da.sendEvent("stopped", map[string]interface{}{
				"reason":           "pause",
				"threadId":         1,
				"text":             "paused by user",
				"allThreadsStopped": true,
			})
			return
		}
		da.sendEvent("stopped", map[string]interface{}{
			"reason":           "exception",
			"threadId":         1,
			"text":             errMsg,
			"allThreadsStopped": true,
		})
		return
	}

	da.sendEvent("terminated", nil)
	da.sendEvent("exited", map[string]interface{}{
		"exitCode": 0,
	})
	close(da.done)
}

func (da *DebugAdapter) handleThreads(msg Message) {
	da.sendResponse(msg, map[string]interface{}{
		"threads": []map[string]interface{}{
			{"id": 1, "name": "main"},
		},
	})
}

func (da *DebugAdapter) handleStackTrace(msg Message) {
	if da.vm == nil {
		da.sendResponse(msg, map[string]interface{}{
			"stackFrames": []map[string]interface{}{},
			"totalFrames": 0,
		})
		return
	}

	frames := []map[string]interface{}{}
	// Get call stack from VM
	for i := da.vm.FramesIndex() - 1; i >= 0; i-- {
		frm := da.vm.FrameAt(i)
		name := "<main>"
		fCl := frm.Cl()
		if fCl != nil && fCl.Fn.Name != "" {
			name = fCl.Fn.Name
		}
		line := 0
		if fCl.Fn.SourceMap != nil {
			if pos, ok := fCl.Fn.SourceMap[frm.IP()]; ok {
				line = pos.Line
			}
		}
		sourcePath := da.sourcePath

		frames = append(frames, map[string]interface{}{
			"id":      i,
			"name":    name,
			"line":    line,
			"column":  0,
			"source":  map[string]interface{}{
				"path":  sourcePath,
				"name":  name,
			},
		})
	}

	da.sendResponse(msg, map[string]interface{}{
		"stackFrames": frames,
		"totalFrames": len(frames),
	})
}

func (da *DebugAdapter) handleScopes(msg Message) {
	da.sendResponse(msg, map[string]interface{}{
		"scopes": []map[string]interface{}{
			{
				"name":               "Locals",
				"presentationHint":   "locals",
				"variablesReference": 1000,
				"expensive":          false,
			},
			{
				"name":               "Globals",
				"presentationHint":   "globals",
				"variablesReference": 1001,
				"expensive":          false,
			},
		},
	})
}

func (da *DebugAdapter) handleVariables(msg Message) {
	if da.vm == nil {
		da.sendResponse(msg, map[string]interface{}{
			"variables": []map[string]interface{}{},
		})
		return
	}

	var args struct {
		VariablesReference int `json:"variablesReference"`
	}
	if msg.Arguments != nil {
		json.Unmarshal(msg.Arguments, &args)
	}

	switch args.VariablesReference {
	case 1000:
		// Locals scope: show stack values for the current frame
		frm := da.vm.CurrentFrame()
		base := frm.BasePointer()
		top := da.vm.SP()
		cl := frm.Cl()

		// Build name->index map from the function's symbol table
		localNames := make(map[int]string)
		if cl != nil && cl.Fn != nil && cl.Fn.SymTable != nil {
			if st, ok := cl.Fn.SymTable.(*symbol.SymbolTable); ok {
				for name, sym := range st.GetStore() {
					if sym.Scope == symbol.LocalScope || sym.Scope == symbol.FunctionScope {
						localNames[sym.Index] = name
					}
				}
			}
		}

		vars := []map[string]interface{}{}
		for i := base; i < top; i++ {
			obj := da.vm.StackAt(i)
			if obj != nil {
				name, ok := localNames[i-base]
				if !ok {
					name = fmt.Sprintf("[%d]", i-base)
				}
				vars = append(vars, map[string]interface{}{
					"name":  name,
					"value": obj.Inspect(),
					"type":  string(obj.Type()),
				})
			}
		}
		if len(vars) == 0 {
			vars = append(vars, map[string]interface{}{
				"name":  "(empty)",
				"value": "",
				"type":  "",
			})
		}
		da.sendResponse(msg, map[string]interface{}{
			"variables": vars,
		})

	case 1001:
		// Globals scope: show global variables
		vars := []map[string]interface{}{}
		if da.symbolTable != nil {
			// Build a name->index map from globals in the symbol table
			globalIdx := make(map[int]string)
			for name, sym := range da.symbolTable.GetStore() {
				if sym.Scope == symbol.GlobalScope {
					globalIdx[sym.Index] = name
				}
			}
			for idx := 0; idx < da.vm.GlobalsLen(); idx++ {
				obj := da.vm.GlobalAt(idx)
				if obj != nil && obj.Type() != "" {
					name, ok := globalIdx[idx]
					if !ok {
						name = fmt.Sprintf("global[%d]", idx)
					}
					vars = append(vars, map[string]interface{}{
						"name":  name,
						"value": obj.Inspect(),
						"type":  string(obj.Type()),
					})
				}
			}
		} else {
			// Fallback: iterate all globals without names
			for idx := 0; idx < da.vm.GlobalsLen(); idx++ {
				obj := da.vm.GlobalAt(idx)
				if obj != nil && obj.Type() != "" {
					vars = append(vars, map[string]interface{}{
						"name":  fmt.Sprintf("global[%d]", idx),
						"value": obj.Inspect(),
						"type":  string(obj.Type()),
					})
				}
			}
		}
		if len(vars) == 0 {
			vars = append(vars, map[string]interface{}{
				"name":  "(empty)",
				"value": "",
				"type":  "",
			})
		}
		da.sendResponse(msg, map[string]interface{}{
			"variables": vars,
		})

	default:
		da.sendResponse(msg, map[string]interface{}{
			"variables": []map[string]interface{}{},
		})
	}
}

func (da *DebugAdapter) handleContinue(msg Message) {
	if da.debugSess != nil {
		da.debugSess.SetStepping(false)
		// Signal the VM to resume (unblock OnInstruction)
		if da.debugSess.ResumeCh != nil {
			select {
			case da.debugSess.ResumeCh <- vm.ActionContinue:
			default:
			}
		}
	}
	da.sendResponse(msg, map[string]interface{}{
		"allThreadsContinued": true,
	})
}

func (da *DebugAdapter) handleNext(msg Message) {
	if da.debugSess != nil {
		da.debugSess.SetStepping(true)
		// Signal the VM to resume with step
		if da.debugSess.ResumeCh != nil {
			select {
			case da.debugSess.ResumeCh <- vm.ActionStep:
			default:
			}
		}
	}
	da.sendResponse(msg, nil)
}

func (da *DebugAdapter) handleStepIn(msg Message) {
	// StepIn works like next in a line-level debugger: executes the next instruction
	if da.debugSess != nil {
		da.debugSess.SetStepping(true)
		if da.debugSess.ResumeCh != nil {
			select {
			case da.debugSess.ResumeCh <- vm.ActionStep:
			default:
			}
		}
	}
	da.sendResponse(msg, nil)
}

func (da *DebugAdapter) handleStepOut(msg Message) {
	// StepOut: continue until the current function returns (set stepping to false,
	// which means run until the next breakpoint)
	if da.debugSess != nil {
		da.debugSess.SetStepping(false)
		if da.debugSess.ResumeCh != nil {
			select {
			case da.debugSess.ResumeCh <- vm.ActionContinue:
			default:
			}
		}
	}
	da.sendResponse(msg, nil)
}

func (da *DebugAdapter) handlePause(msg Message) {
	if da.vm != nil {
		// Try cooperative pause via debug session first
		if da.debugSess != nil {
			da.debugSess.PauseRequested = true
		}
		// Fallback: cancel the VM context to force interruption
		atomic.StoreInt32(&da.pauseFlag, 1)
		da.vm.Cancel()
	}
	da.sendResponse(msg, nil)
}

func (da *DebugAdapter) handleEvaluate(msg Message) {
	var args struct {
		Expression string `json:"expression"`
		FrameID    int    `json:"frameId"`
		Context    string `json:"context"`
	}
	if msg.Arguments != nil {
		json.Unmarshal(msg.Arguments, &args)
	}

	if da.vm == nil || da.debugSess == nil {
		da.sendErrorResponse(msg, "no active debug session")
		return
	}

	expr := strings.TrimSpace(args.Expression)
	if expr == "" {
		da.sendErrorResponse(msg, "empty expression")
		return
	}

	// Try to resolve as a variable name in locals first, then globals
	result := da.lookupVariable(expr)
	if result != "" {
		da.sendResponse(msg, map[string]interface{}{
			"result": result,
			"type":   "string",
		})
		return
	}

	da.sendErrorResponse(msg, fmt.Sprintf("cannot evaluate: '%s'", expr))
}

// lookupVariable attempts to find a variable by name in the current frame's locals or globals.
func (da *DebugAdapter) lookupVariable(name string) string {
	if da.vm == nil {
		return ""
	}

	frm := da.vm.CurrentFrame()
	base := frm.BasePointer()
	top := da.vm.SP()
	cl := frm.Cl()

	// Check locals using symbol table
	if cl != nil && cl.Fn != nil && cl.Fn.SymTable != nil {
		if st, ok := cl.Fn.SymTable.(*symbol.SymbolTable); ok {
			sym, found := st.Resolve(name)
			if found {
				switch sym.Scope {
				case symbol.LocalScope, symbol.FunctionScope:
					idx := base + sym.Index
					if idx < top {
						obj := da.vm.StackAt(idx)
						if obj != nil {
							return obj.Inspect()
						}
					}
				case symbol.GlobalScope:
					obj := da.vm.GlobalAt(sym.Index)
					if obj != nil {
						return obj.Inspect()
					}
				case symbol.FreeScope:
					if sym.Index < len(cl.Free) {
						obj := cl.Free[sym.Index]
						if obj != nil {
							return obj.Inspect()
						}
					}
				case symbol.BuiltinScope:
					return "<builtin>"
				}
			}
		}
	}

	// Fallback: check globals via top-level symbol table
	if da.symbolTable != nil {
		sym, found := da.symbolTable.Resolve(name)
		if found && sym.Scope == symbol.GlobalScope {
			obj := da.vm.GlobalAt(sym.Index)
			if obj != nil {
				return obj.Inspect()
			}
		}
	}

	return ""
}

func (da *DebugAdapter) handleExceptionInfo(msg Message) {
	var args struct {
		ThreadID int `json:"threadId"`
	}
	if msg.Arguments != nil {
		json.Unmarshal(msg.Arguments, &args)
	}

	if da.lastException == nil {
		da.sendErrorResponse(msg, "no exception available")
		return
	}

	errMsg := da.lastException.Error()
	exceptionType := "RuntimeError"
	stackTrace := ""

	// Try to extract structured info from RuntimeError
	if re, ok := da.lastException.(*vm.RuntimeError); ok {
		stackTrace = re.Error()
		for _, cf := range re.StackTrace {
			stackTrace += fmt.Sprintf("\n  at %s (%s:%d:%d)", cf.Function, cf.File, cf.Line, cf.Column)
		}
	}

	da.sendResponse(msg, map[string]interface{}{
		"exceptionId":   exceptionType,
		"breakMode":     "unhandled",
		"description":   errMsg,
		"details": map[string]interface{}{
			"message":   errMsg,
			"typeName":  exceptionType,
			"stackTrace": stackTrace,
		},
	})
}

func (da *DebugAdapter) handleDisconnect(msg Message) {
	da.sendResponse(msg, nil)
	close(da.done)
}
