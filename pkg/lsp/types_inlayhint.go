package lsp

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// InlayHintKind represents LSP 3.17 InlayHintKind (not in glsp's protocol_3_16).
type InlayHintKind int

const (
	InlayHintKindType      InlayHintKind = 1
	InlayHintKindParameter InlayHintKind = 2
)

// InlayHintParams represents LSP 3.17 InlayHintParams.
type InlayHintParams struct {
	TextDocument protocol.TextDocumentIdentifier `json:"textDocument"`
	Range        protocol.Range                  `json:"range"`
}

// InlayHint represents LSP 3.17 InlayHint.
type InlayHint struct {
	Position     protocol.Position `json:"position"`
	Label        string            `json:"label"`
	Kind         *InlayHintKind    `json:"kind,omitempty"`
	PaddingLeft  *bool             `json:"paddingLeft,omitempty"`
	PaddingRight *bool             `json:"paddingRight,omitempty"`
}
