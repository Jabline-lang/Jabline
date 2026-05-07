package lsp

import (
	"encoding/json"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	workspaceStore.UpdateDocument(params.TextDocument.URI, params.TextDocument.Text, context)
	logger.Info("textDocumentDidOpen: stored " + params.TextDocument.URI)
	return nil
}

func textDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}

	// vscode-languageclient v9 with full-document sync sends either:
	//   TextDocumentContentChangeEvent      { range, rangeLength, text }
	//   TextDocumentContentChangeEventWhole { text }  ← most common for full sync
	// We try typed assertions first, then fall back to raw JSON parsing.
	for _, raw := range params.ContentChanges {
		switch c := raw.(type) {
		case protocol.TextDocumentContentChangeEvent:
			workspaceStore.UpdateDocument(params.TextDocument.URI, c.Text, context)
			return nil
		case protocol.TextDocumentContentChangeEventWhole:
			workspaceStore.UpdateDocument(params.TextDocument.URI, c.Text, context)
			return nil
		default:
			// Last-resort: marshal back to JSON and extract "text" field
			b, err := json.Marshal(raw)
			if err == nil {
				var m map[string]interface{}
				if json.Unmarshal(b, &m) == nil {
					if text, ok := m["text"].(string); ok && text != "" {
						workspaceStore.UpdateDocument(params.TextDocument.URI, text, context)
						return nil
					}
				}
			}
		}
	}

	return nil
}
