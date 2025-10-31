package ast

import (
	"jabline/pkg/token"
)

type ImportItem struct {
	Name  *Identifier
	Alias *Identifier
}

func (ii *ImportItem) String() string {
	if ii.Alias != nil {
		return ii.Name.String() + " as " + ii.Alias.String()
	}
	return ii.Name.String()
}

type ImportType int

const (
	IMPORT_DEFAULT ImportType = iota
	IMPORT_NAMED
	IMPORT_NAMESPACE
	IMPORT_SIDE_EFFECT
	IMPORT_MIXED
)

type ImportStatement struct {
	Token          token.Token
	ImportType     ImportType
	ModuleName     *StringLiteral
	DefaultImport  *Identifier
	NamedImports   []*ImportItem
	NamespaceAlias *Identifier
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

type ExportType int

const (
	EXPORT_DECLARATION ExportType = iota
	EXPORT_DEFAULT
	EXPORT_LIST
	EXPORT_ALL
	EXPORT_NAMED_FROM
	EXPORT_ALL_AS
)

type ExportItem struct {
	Name  *Identifier
	Alias *Identifier
}

func (ei *ExportItem) String() string {
	if ei.Alias != nil {
		return ei.Name.String() + " as " + ei.Alias.String()
	}
	return ei.Name.String()
}

type ExportStatement struct {
	Token          token.Token
	ExportType     ExportType
	Statement      Statement
	ExportList     []*ExportItem
	ModuleName     *StringLiteral
	NamespaceAlias *Identifier
	IsDefault      bool
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

type ModuleResolver struct {
	BasePath    string
	ModulePaths []string
	Extensions  []string
}

type Module struct {
	Name          string
	Path          string
	Exports       map[string]interface{}
	DefaultExport interface{}
	Dependencies  []string
	Resolved      bool
}

type ReExportStatement struct {
	Token      token.Token
	ModuleName *StringLiteral
	ExportList []*ExportItem
	Alias      *Identifier
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
