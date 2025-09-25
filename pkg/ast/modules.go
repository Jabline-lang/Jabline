package ast

import (
	"jabline/pkg/token"
)

// ImportItem represents a single imported item with optional alias
type ImportItem struct {
	Name  *Identifier // original name
	Alias *Identifier // alias if using "as", nil otherwise
}

func (ii *ImportItem) String() string {
	if ii.Alias != nil {
		return ii.Name.String() + " as " + ii.Alias.String()
	}
	return ii.Name.String()
}

// ImportType represents the type of import
type ImportType int

const (
	IMPORT_DEFAULT     ImportType = iota // import defaultExport from "module"
	IMPORT_NAMED                         // import { name1, name2 } from "module"
	IMPORT_NAMESPACE                     // import * as namespace from "module"
	IMPORT_SIDE_EFFECT                   // import "module"
	IMPORT_MIXED                         // import defaultExport, { name1, name2 } from "module"
)

// ImportStatement represents all types of import statements
type ImportStatement struct {
	Token          token.Token    // 'import'
	ImportType     ImportType     // type of import
	ModuleName     *StringLiteral // module path
	DefaultImport  *Identifier    // for default imports
	NamedImports   []*ImportItem  // for named imports
	NamespaceAlias *Identifier    // for namespace imports (import * as alias)
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	switch is.ImportType {
	case IMPORT_SIDE_EFFECT:
		return "import " + is.ModuleName.String()

	case IMPORT_DEFAULT:
		return "import " + is.DefaultImport.String() + " from " + is.ModuleName.String()

	case IMPORT_NAMED:
		out := "import { "
		for i, item := range is.NamedImports {
			if i > 0 {
				out += ", "
			}
			out += item.String()
		}
		out += " } from " + is.ModuleName.String()
		return out

	case IMPORT_NAMESPACE:
		return "import * as " + is.NamespaceAlias.String() + " from " + is.ModuleName.String()

	case IMPORT_MIXED:
		out := "import " + is.DefaultImport.String() + ", { "
		for i, item := range is.NamedImports {
			if i > 0 {
				out += ", "
			}
			out += item.String()
		}
		out += " } from " + is.ModuleName.String()
		return out
	}
	return "import"
}

// ExportType represents the type of export
type ExportType int

const (
	EXPORT_DECLARATION ExportType = iota // export let/const/fn/etc.
	EXPORT_DEFAULT                       // export default expression
	EXPORT_LIST                          // export { name1, name2 }
	EXPORT_ALL                           // export * from "module"
	EXPORT_NAMED_FROM                    // export { name1, name2 } from "module"
	EXPORT_ALL_AS                        // export * as namespace from "module"
)

// ExportItem represents a single exported item with optional alias
type ExportItem struct {
	Name  *Identifier // original name
	Alias *Identifier // export alias if using "as", nil otherwise
}

func (ei *ExportItem) String() string {
	if ei.Alias != nil {
		return ei.Name.String() + " as " + ei.Alias.String()
	}
	return ei.Name.String()
}

// ExportStatement represents all types of export statements
type ExportStatement struct {
	Token          token.Token    // 'export'
	ExportType     ExportType     // type of export
	Statement      Statement      // for EXPORT_DECLARATION and EXPORT_DEFAULT
	ExportList     []*ExportItem  // for EXPORT_LIST and EXPORT_NAMED_FROM
	ModuleName     *StringLiteral // for re-exports (export ... from "module")
	NamespaceAlias *Identifier    // for EXPORT_ALL_AS
	IsDefault      bool           // true for default exports
}

func (es *ExportStatement) statementNode()       {}
func (es *ExportStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExportStatement) String() string {
	switch es.ExportType {
	case EXPORT_DECLARATION:
		out := "export "
		if es.IsDefault {
			out += "default "
		}
		if es.Statement != nil {
			out += es.Statement.String()
		}
		return out

	case EXPORT_DEFAULT:
		return "export default " + es.Statement.String()

	case EXPORT_LIST:
		out := "export { "
		for i, item := range es.ExportList {
			if i > 0 {
				out += ", "
			}
			out += item.String()
		}
		out += " }"
		return out

	case EXPORT_ALL:
		return "export * from " + es.ModuleName.String()

	case EXPORT_NAMED_FROM:
		out := "export { "
		for i, item := range es.ExportList {
			if i > 0 {
				out += ", "
			}
			out += item.String()
		}
		out += " } from " + es.ModuleName.String()
		return out

	case EXPORT_ALL_AS:
		return "export * as " + es.NamespaceAlias.String() + " from " + es.ModuleName.String()
	}
	return "export"
}

// ModuleResolver represents module resolution information
type ModuleResolver struct {
	BasePath    string
	ModulePaths []string
	Extensions  []string
}

// Module represents a module with its metadata and exports
type Module struct {
	Name          string
	Path          string
	Exports       map[string]interface{}
	DefaultExport interface{}
	Dependencies  []string
	Resolved      bool
}

// ReExportStatement represents re-export statements
type ReExportStatement struct {
	Token      token.Token
	ModuleName *StringLiteral
	ExportList []*ExportItem // nil for export *
	Alias      *Identifier   // for export * as alias
}

func (res *ReExportStatement) statementNode()       {}
func (res *ReExportStatement) TokenLiteral() string { return res.Token.Literal }
func (res *ReExportStatement) String() string {
	if res.Alias != nil {
		return "export * as " + res.Alias.String() + " from " + res.ModuleName.String()
	}
	if len(res.ExportList) == 0 {
		return "export * from " + res.ModuleName.String()
	}

	out := "export { "
	for i, item := range res.ExportList {
		if i > 0 {
			out += ", "
		}
		out += item.String()
	}
	out += " } from " + res.ModuleName.String()
	return out
}
