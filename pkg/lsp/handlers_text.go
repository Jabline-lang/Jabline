package lsp

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type debounceEntry struct {
	timer *time.Timer
	mu    sync.Mutex
}

var (
	debounceTimers   = make(map[string]*debounceEntry)
	debounceMu       sync.Mutex
	debounceDuration = 250 * time.Millisecond
)

func getDebounceEntry(uri string) *debounceEntry {
	debounceMu.Lock()
	defer debounceMu.Unlock()
	entry, ok := debounceTimers[uri]
	if !ok {
		entry = &debounceEntry{}
		debounceTimers[uri] = entry
	}
	return entry
}

func removeDebounceEntry(uri string) {
	debounceMu.Lock()
	defer debounceMu.Unlock()
	if entry, ok := debounceTimers[uri]; ok {
		entry.mu.Lock()
		if entry.timer != nil {
			entry.timer.Stop()
		}
		entry.mu.Unlock()
		delete(debounceTimers, uri)
	}
}

func debounceUpdateDocument(uri string, content string, context *glsp.Context) {
	entry := getDebounceEntry(uri)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	if entry.timer != nil {
		entry.timer.Stop()
	}
	entry.timer = time.AfterFunc(debounceDuration, func() {
		workspaceStore.UpdateDocument(uri, content, context)
	})
}

func textDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	workspaceStore.UpdateDocument(params.TextDocument.URI, params.TextDocument.Text, context)
	logger.Info("textDocumentDidOpen: stored " + params.TextDocument.URI)
	return nil
}

func textDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}

	var text string
	for _, raw := range params.ContentChanges {
		switch c := raw.(type) {
		case protocol.TextDocumentContentChangeEvent:
			text = c.Text
		case protocol.TextDocumentContentChangeEventWhole:
			text = c.Text
		default:
			b, err := json.Marshal(raw)
			if err == nil {
				var m map[string]interface{}
				if json.Unmarshal(b, &m) == nil {
					if t, ok := m["text"].(string); ok && t != "" {
						text = t
					}
				}
			}
		}
	}

	if text == "" {
		return nil
	}

	debounceUpdateDocument(params.TextDocument.URI, text, context)
	return nil
}

func textDocumentDidClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	removeDebounceEntry(params.TextDocument.URI)
	workspaceStore.RemoveDocument(params.TextDocument.URI)
	logger.Info("textDocumentDidClose: removed " + params.TextDocument.URI)
	return nil
}

func textDocumentDidSave(context *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	if params.Text != nil {
		workspaceStore.UpdateDocument(params.TextDocument.URI, *params.Text, context)
		logger.Info("textDocumentDidSave: updated " + params.TextDocument.URI)
	}
	return nil
}
