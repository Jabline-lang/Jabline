package lsp

import (
	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"jabline/pkg/parser"
	"sync"
	"strconv"
	"regexp"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var errorRegex = regexp.MustCompile(`line (\d+), column (\d+): (.*)`)

type DocumentSemanticInfo struct {
	Program     *ast.Program
	SymbolTable *SymbolTable
	URI         string
}

type WorkspaceSymbolStore struct {
	Documents map[string]*DocumentSemanticInfo
	Mutex     sync.RWMutex
	
	analyzingDocuments map[string]bool
	analysisMutex sync.Mutex
}

var workspaceStore = &WorkspaceSymbolStore{
	Documents: make(map[string]*DocumentSemanticInfo),
	analyzingDocuments: make(map[string]bool),
}

func (ws *WorkspaceSymbolStore) AddAnalyzingDocument(uri string) {
	ws.analysisMutex.Lock()
	defer ws.analysisMutex.Unlock()
	ws.analyzingDocuments[uri] = true
}

func (ws *WorkspaceSymbolStore) RemoveAnalyzingDocument(uri string) {
	ws.analysisMutex.Lock()
	defer ws.analysisMutex.Unlock()
	delete(ws.analyzingDocuments, uri)
}

func (ws *WorkspaceSymbolStore) IsAnalyzingDocument(uri string) bool {
	ws.analysisMutex.Lock()
	defer ws.analysisMutex.Unlock()
	return ws.analyzingDocuments[uri]
}


func (ws *WorkspaceSymbolStore) UpdateDocument(uri string, content string, context *glsp.Context) {

	l := lexer.New(content)
	p := parser.New(l)
	program := p.ParseProgram()

	sa := NewSemanticAnalyzer(program, ws, uri)
	sa.Analyze()

	ws.Mutex.Lock()
	ws.Documents[uri] = &DocumentSemanticInfo{
		Program:     program,
		SymbolTable: sa.Symbols,
		URI:         uri,
	}
	ws.Mutex.Unlock()

	var diagnostics []protocol.Diagnostic

	for _, errStr := range p.Errors() {
		matches := errorRegex.FindStringSubmatch(errStr)
		if len(matches) == 4 {
			line, _ := strconv.Atoi(matches[1])
			col, _ := strconv.Atoi(matches[2])
			msg := matches[3]

			lineIndex := uint32(line - 1)
			colIndex := uint32(col - 1)

			diagnostic := protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: lineIndex, Character: colIndex},
					End:   protocol.Position{Line: lineIndex, Character: colIndex + 1},
				},
				Severity: ptr(protocol.DiagnosticSeverityError),
				Source:   ptr(lsName),
				Message:  msg,
			}
			diagnostics = append(diagnostics, diagnostic)
		} else {
			diagnostic := protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 1},
				},
				Severity: ptr(protocol.DiagnosticSeverityError),
				Source:   ptr(lsName),
				Message:  errStr,
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	for _, errStr := range sa.Errors {
		matches := errorRegex.FindStringSubmatch(errStr)
		if len(matches) == 4 {
			line, _ := strconv.Atoi(matches[1])
			col, _ := strconv.Atoi(matches[2])
			msg := matches[3]

			lineIndex := uint32(line - 1)
			colIndex := uint32(col - 1)

			diagnostic := protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: lineIndex, Character: colIndex},
					End:   protocol.Position{Line: lineIndex, Character: colIndex + 1},
				},
				Severity: ptr(protocol.DiagnosticSeverityError),
				Source:   ptr(lsName),
				Message:  msg,
			}
			diagnostics = append(diagnostics, diagnostic)
		} else {
			diagnostic := protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 1},
				},
				Severity: ptr(protocol.DiagnosticSeverityError),
				Source:   ptr(lsName),
				Message:  errStr,
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}

	go context.Notify("textDocument/publishDiagnostics", protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}
