package lsp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

const lsName = "jabline-lsp"

var version = "0.1.0"
var handler protocol.Handler
var logger commonlog.Logger

// customHandler wraps protocol.Handler and supports methods not in LSP 3.16 (e.g. InlayHint).
type customHandler struct {
	protocol *protocol.Handler
	custom   map[string]func(*glsp.Context) (any, bool, bool, error)
}

func (h *customHandler) Handle(context *glsp.Context) (any, bool, bool, error) {
	if handlerFn, ok := h.custom[context.Method]; ok {
		return handlerFn(context)
	}
	return h.protocol.Handle(context)
}

func NewServer() *server.Server {
	// Configure logging to file
	logFile := filepath.Join(os.TempDir(), "jabline-lsp.log")
	commonlog.Configure(2, &logFile)
	logger = commonlog.GetLogger(lsName)

	handler = protocol.Handler{
		Initialize:                    withRecovery("Initialize", initialize),
		Initialized:                   withRecoveryError("Initialized", initialized),
		Shutdown:                      withRecoveryNoParams("Shutdown", shutdown),
		SetTrace:                      withRecoveryError("SetTrace", setTrace),
		TextDocumentDidOpen:           withRecoveryError("TextDocumentDidOpen", textDocumentDidOpen),
		TextDocumentDidChange:         withRecoveryError("TextDocumentDidChange", textDocumentDidChange),
		TextDocumentDidClose:          withRecoveryError("TextDocumentDidClose", textDocumentDidClose),
		TextDocumentDidSave:           withRecoveryError("TextDocumentDidSave", textDocumentDidSave),
		TextDocumentHover:             withRecovery("TextDocumentHover", textDocumentHover),
		TextDocumentDefinition:        withRecovery("TextDocumentDefinition", textDocumentDefinition),
		TextDocumentTypeDefinition:    withRecovery("TextDocumentTypeDefinition", textDocumentTypeDefinition),
		TextDocumentImplementation:    withRecovery("TextDocumentImplementation", textDocumentImplementation),
		TextDocumentDocumentSymbol:    withRecovery("TextDocumentDocumentSymbol", textDocumentDocumentSymbol),
		TextDocumentCompletion:        withRecovery("TextDocumentCompletion", textDocumentCompletion),
		CompletionItemResolve:         withRecovery("CompletionItemResolve", textDocumentCompletionResolve),
		TextDocumentSignatureHelp:     withRecovery("TextDocumentSignatureHelp", textDocumentSignatureHelp),
		TextDocumentReferences:        withRecovery("TextDocumentReferences", textDocumentReferences),
		TextDocumentRename:            withRecovery("TextDocumentRename", textDocumentRename),
		TextDocumentPrepareRename:     withRecovery("TextDocumentPrepareRename", textDocumentPrepareRename),
		TextDocumentFormatting:        withRecovery("TextDocumentFormatting", textDocumentFormatting),
		TextDocumentCodeAction:        withRecovery("TextDocumentCodeAction", textDocumentCodeAction),
		TextDocumentFoldingRange:      withRecovery("TextDocumentFoldingRange", textDocumentFoldingRange),
		TextDocumentSemanticTokensFull: withRecovery("TextDocumentSemanticTokensFull", textDocumentSemanticTokensFull),
		WorkspaceSymbol:               withRecovery("WorkspaceSymbol", workspaceSymbol),
		TextDocumentCodeLens:          withRecovery("TextDocumentCodeLens", textDocumentCodeLens),
		TextDocumentPrepareCallHierarchy: withRecovery("TextDocumentPrepareCallHierarchy", textDocumentPrepareCallHierarchy),
		CallHierarchyIncomingCalls:    withRecovery("CallHierarchyIncomingCalls", callHierarchyIncomingCalls),
		CallHierarchyOutgoingCalls:    withRecovery("CallHierarchyOutgoingCalls", callHierarchyOutgoingCalls),
	}

	wrapped := &customHandler{
		protocol: &handler,
		custom: map[string]func(*glsp.Context) (any, bool, bool, error){
			"textDocument/inlayHint": withRecoveryInlayHint,
		},
	}

	return server.NewServer(wrapped, lsName, true)
}

func withRecoveryInlayHint(context *glsp.Context) (result any, validMethod bool, validParams bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("PANIC in InlayHint handler: %v\n%s", r, debug.Stack()))
			err = fmt.Errorf("LSP server error in InlayHint: %v", r)
			validMethod = true
		}
	}()
	var params InlayHintParams
	if err2 := json.Unmarshal(context.Params, &params); err2 != nil {
		return nil, true, false, nil
	}
	hints, err2 := textDocumentInlayHint(context, &params)
	if err2 != nil {
		return nil, true, true, err2
	}
	return hints, true, true, nil
}

func withRecovery[P any, R any](name string, f func(*glsp.Context, P) (R, error)) func(*glsp.Context, P) (R, error) {
	return func(context *glsp.Context, params P) (result R, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("PANIC in %s handler: %v\n%s", name, r, debug.Stack()))
				err = fmt.Errorf("LSP server error in %s: %v", name, r)
			}
		}()
		return f(context, params)
	}
}

func withRecoveryError[P any](name string, f func(*glsp.Context, P) error) func(*glsp.Context, P) error {
	return func(context *glsp.Context, params P) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("PANIC in %s handler: %v\n%s", name, r, debug.Stack()))
				err = fmt.Errorf("LSP server error in %s: %v", name, r)
			}
		}()
		return f(context, params)
	}
}

func withRecoveryNoParams(name string, f func(*glsp.Context) error) func(*glsp.Context) error {
	return func(context *glsp.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("PANIC in %s handler: %v\n%s", name, r, debug.Stack()))
				err = fmt.Errorf("LSP server error in %s: %v", name, r)
			}
		}()
		return f(context)
	}
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()

	syncKind := protocol.TextDocumentSyncKindFull
	capabilities.TextDocumentSync = &protocol.TextDocumentSyncOptions{
		OpenClose: ptr(true),
		Change:    &syncKind,
		Save: &protocol.SaveOptions{
			IncludeText: ptr(false),
		},
	}
	capabilities.HoverProvider = true
	capabilities.DefinitionProvider = true
	capabilities.TypeDefinitionProvider = true
	capabilities.ImplementationProvider = true
	capabilities.DocumentSymbolProvider = true
	capabilities.WorkspaceSymbolProvider = true
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{".", ":"},
		ResolveProvider:   ptr(true),
	}
	capabilities.SignatureHelpProvider = &protocol.SignatureHelpOptions{
		TriggerCharacters: []string{"(", ","},
	}
	capabilities.ReferencesProvider = true
	capabilities.RenameProvider = &protocol.RenameOptions{
		PrepareProvider: ptr(true),
	}
	capabilities.DocumentFormattingProvider = true
	capabilities.CodeActionProvider = &protocol.CodeActionOptions{
		CodeActionKinds: []protocol.CodeActionKind{
			protocol.CodeActionKindQuickFix,
			protocol.CodeActionKindRefactor,
		},
	}
	capabilities.FoldingRangeProvider = &protocol.FoldingRangeOptions{}

	capabilities.SemanticTokensProvider = &protocol.SemanticTokensOptions{
		Full: &protocol.SemanticDelta{Delta: ptr(false)},
		Legend: protocol.SemanticTokensLegend{
			TokenTypes:     semanticTokenTypes,
			TokenModifiers: semanticTokenModifiers,
		},
	}

	capabilities.CodeLensProvider = &protocol.CodeLensOptions{}
	capabilities.CallHierarchyProvider = true

	result := protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}

	// Marshal to JSON and add InlayHint (LSP 3.17, not in glsp protocol_3_16)
	data, err := json.Marshal(result)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to marshal initialize result: %v", err))
		return result, nil
	}
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		logger.Error(fmt.Sprintf("failed to unmarshal initialize result: %v", err))
		return result, nil
	}
	if caps, ok := rawMap["capabilities"].(map[string]interface{}); ok {
		caps["inlayHintProvider"] = true
	}
	return rawMap, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
