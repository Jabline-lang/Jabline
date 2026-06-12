package lsp

import (
	"fmt"
	"jabline/pkg/ast"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentCodeLens(context *glsp.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	workspaceStore.Mutex.RLock()
	docInfo, ok := workspaceStore.Documents[params.TextDocument.URI]
	workspaceStore.Mutex.RUnlock()

	if !ok || docInfo == nil || docInfo.SymbolTable == nil {
		return nil, nil
	}

	var lenses []protocol.CodeLens

	var walkScope func(s *Scope)
	walkScope = func(s *Scope) {
		for _, sym := range s.Symbols {
			if sym.Definition == nil {
				continue
			}
			switch sym.Definition.(type) {
			case *ast.FunctionStatement, *ast.AsyncFunctionStatement,
				*ast.StructStatement, *ast.InterfaceStatement,
				*ast.EnumStatement, *ast.ServiceStatement:

				refCount := len(sym.References)
				title := fmt.Sprintf("%d reference", refCount)
				if refCount != 1 {
					title += "s"
				}

				dt := declTokenFromSymbol(sym)
				if dt == nil {
					continue
				}
				start, end := tokToPos(*dt)

				lenses = append(lenses, protocol.CodeLens{
					Range: protocol.Range{Start: start, End: end},
					Command: &protocol.Command{
						Title:     title,
						Command:   "editor.action.showReferences",
						Arguments: []any{params.TextDocument.URI, start},
					},
				})
			}
		}
		for _, child := range s.Children {
			walkScope(child)
		}
	}

	walkScope(docInfo.SymbolTable.RootScope)
	return lenses, nil
}
