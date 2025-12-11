package lsp

import (
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

func NewServer() *server.Server {
	// Configure logging to file
	logFile := filepath.Join(os.TempDir(), "jabline-lsp.log")
	commonlog.Configure(2, &logFile)
	logger = commonlog.GetLogger(lsName)

	handler = protocol.Handler{
		Initialize:                 withRecovery("Initialize", initialize),
		Initialized:                withRecoveryError("Initialized", initialized),
		Shutdown:                   withRecoveryNoParams("Shutdown", shutdown),
		SetTrace:                   withRecoveryError("SetTrace", setTrace),
		TextDocumentDidOpen:        withRecoveryError("TextDocumentDidOpen", textDocumentDidOpen),
		TextDocumentDidChange:      withRecoveryError("TextDocumentDidChange", textDocumentDidChange),
		TextDocumentHover:          withRecovery("TextDocumentHover", textDocumentHover),
		TextDocumentDefinition:     withRecovery("TextDocumentDefinition", textDocumentDefinition),
		TextDocumentDocumentSymbol: withRecovery("TextDocumentDocumentSymbol", textDocumentDocumentSymbol),
		TextDocumentCompletion:     withRecovery("TextDocumentCompletion", textDocumentCompletion),
		TextDocumentSignatureHelp:  withRecovery("TextDocumentSignatureHelp", textDocumentSignatureHelp),
		TextDocumentReferences:     withRecovery("TextDocumentReferences", textDocumentReferences),
		TextDocumentRename:         withRecovery("TextDocumentRename", textDocumentRename),
	}

	return server.NewServer(&handler, lsName, true)
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
	capabilities.TextDocumentSync = protocol.TextDocumentSyncKindFull
	capabilities.HoverProvider = true
	capabilities.DefinitionProvider = true
	capabilities.DocumentSymbolProvider = true
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{".", ":"},
	}
	capabilities.SignatureHelpProvider = &protocol.SignatureHelpOptions{
		TriggerCharacters: []string{"(,"},
	}
	capabilities.ReferencesProvider = true
	capabilities.RenameProvider = true

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
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
