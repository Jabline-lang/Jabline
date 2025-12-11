package lsp

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {

	workspaceStore.UpdateDocument(params.TextDocument.URI, params.TextDocument.Text, context)
	return nil
}

func textDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {

	if len(params.ContentChanges) > 0 {
		change, ok := params.ContentChanges[0].(protocol.TextDocumentContentChangeEvent)
		if ok {

			workspaceStore.UpdateDocument(params.TextDocument.URI, change.Text, context)
		}
	}
	return nil
}
