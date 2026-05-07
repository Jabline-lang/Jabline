package fmt

import (
	"bytes"
	gofmt "fmt"
	"jabline/pkg/ast"
	"sort"
	"strings"
)

// Format returns a well-formatted string of the given AST node.
// It is safe to call on any ast.Node, including a full *ast.Program.
func Format(node ast.Node) string {
	var out bytes.Buffer
	formatNode(&out, node, 0)
	return strings.TrimRight(out.String(), "\n")
}

// ─── Core dispatcher ───────────────────────────────────────────────────────────

func formatNode(out *bytes.Buffer, node ast.Node, indent int) {
	if node == nil {
		return
	}

	switch n := node.(type) {

	// ─── Top-level program ──────────────────────────────────────────────────
	case *ast.Program:
		// stmtCategory returns a grouping category for blank-line separation:
		//   0 = import
		//   1 = let / const (variable declarations)
		//   2 = struct / interface / enum / function / service (heavy declarations)
		//   3 = everything else (echo, expressions, control flow, etc.)
		stmtCategory := func(s ast.Statement) int {
			switch s.(type) {
			case *ast.ImportStatement:
				return 0
			case *ast.LetStatement, *ast.ConstStatement:
				return 1
			case *ast.FunctionStatement, *ast.AsyncFunctionStatement,
				*ast.StructStatement, *ast.InterfaceStatement,
				*ast.EnumStatement, *ast.ServiceStatement:
				return 2
			default:
				return 3
			}
		}

		lastCat := -1
		for i, stmt := range n.Statements {
			cat := stmtCategory(stmt)
			heavy := isHeavy(stmt)

			if i > 0 {
				// Always blank line before/after heavy declarations
				if heavy || isHeavy(n.Statements[i-1]) {
					out.WriteString("\n")
				} else if cat != lastCat {
					// Blank line whenever the statement category changes
					out.WriteString("\n")
				}
			}

			formatNode(out, stmt, indent)
			out.WriteString("\n")
			lastCat = cat
		}

	// ─── Statements ─────────────────────────────────────────────────────────
	case *ast.ExpressionStatement:
		writeIndent(out, indent)
		formatNode(out, n.Expression, indent)
		if !exprEndsWithBlock(n.Expression) {
			out.WriteString(";")
		}

	case *ast.LetStatement:
		writeIndent(out, indent)
		out.WriteString("let ")
		out.WriteString(n.Name.Value)
		if n.Type != nil {
			out.WriteString(": ")
			out.WriteString(n.Type.String())
		}
		out.WriteString(" = ")
		formatNode(out, n.Value, indent)
		if !exprEndsWithBlock(n.Value) {
			out.WriteString(";")
		}

	case *ast.ConstStatement:
		writeIndent(out, indent)
		out.WriteString("const ")
		out.WriteString(n.Name.Value)
		if n.Type != nil {
			out.WriteString(": ")
			out.WriteString(n.Type.String())
		}
		out.WriteString(" = ")
		formatNode(out, n.Value, indent)
		if !exprEndsWithBlock(n.Value) {
			out.WriteString(";")
		}

	case *ast.AssignmentStatement:
		writeIndent(out, indent)
		formatNode(out, n.Left, indent)
		out.WriteString(" = ")
		formatNode(out, n.Value, indent)
		out.WriteString(";")

	case *ast.ReturnStatement:
		writeIndent(out, indent)
		out.WriteString("return")
		if n.ReturnValue != nil {
			out.WriteString(" ")
			formatNode(out, n.ReturnValue, indent)
		}
		out.WriteString(";")

	case *ast.EchoStatement:
		writeIndent(out, indent)
		out.WriteString("echo(")
		for i, val := range n.Values {
			formatNode(out, val, indent)
			if i < len(n.Values)-1 {
				out.WriteString(", ")
			}
		}
		out.WriteString(");")

	case *ast.BreakStatement:
		writeIndent(out, indent)
		out.WriteString("break;")

	case *ast.ContinueStatement:
		writeIndent(out, indent)
		out.WriteString("continue;")

	case *ast.ThrowStatement:
		writeIndent(out, indent)
		out.WriteString("throw ")
		formatNode(out, n.Value, indent)
		out.WriteString(";")

	// ─── Block ──────────────────────────────────────────────────────────────
	case *ast.BlockStatement:
		out.WriteString(" {\n")
		lastWasHeavy := false
		for i, stmt := range n.Statements {
			heavy := isHeavy(stmt)
			if i > 0 && (lastWasHeavy || heavy) {
				out.WriteString("\n")
			}
			formatNode(out, stmt, indent+1)
			out.WriteString("\n")
			lastWasHeavy = heavy
		}
		writeIndent(out, indent)
		out.WriteString("}")

	// ─── Control Flow ────────────────────────────────────────────────────────
	case *ast.IfExpression:
		writeIndent(out, indent)
		formatIf(out, n, indent)

	case *ast.WhileStatement:
		writeIndent(out, indent)
		out.WriteString("while (")
		formatNode(out, n.Condition, 0)
		out.WriteString(")")
		formatNode(out, n.Body, indent)

	case *ast.ForStatement:
		writeIndent(out, indent)
		out.WriteString("for (")
		if n.Init != nil {
			// Init is usually a LetStatement or AssignmentStatement — emit without indent/semicolons
			formatForInit(out, n.Init, indent)
		}
		out.WriteString("; ")
		if n.Condition != nil {
			formatNode(out, n.Condition, 0)
		}
		out.WriteString("; ")
		if n.Update != nil {
			formatForInit(out, n.Update, indent)
		}
		out.WriteString(")")
		formatNode(out, n.Body, indent)

	case *ast.ForEachStatement:
		writeIndent(out, indent)
		out.WriteString("for (")
		out.WriteString(n.Variable.Value)
		out.WriteString(" in ")
		formatNode(out, n.Iterable, 0)
		out.WriteString(")")
		formatNode(out, n.Body, indent)

	case *ast.SwitchStatement:
		writeIndent(out, indent)
		out.WriteString("switch (")
		formatNode(out, n.Expression, 0)
		out.WriteString(") {\n")
		for _, c := range n.Cases {
			writeIndent(out, indent+1)
			out.WriteString("case ")
			formatNode(out, c.Value, 0)
			out.WriteString(":\n")
			for _, stmt := range c.Statements {
				formatNode(out, stmt, indent+2)
				out.WriteString("\n")
			}
		}
		if n.DefaultCase != nil {
			writeIndent(out, indent+1)
			out.WriteString("default:\n")
			for _, stmt := range n.DefaultCase.Statements {
				formatNode(out, stmt, indent+2)
				out.WriteString("\n")
			}
		}
		writeIndent(out, indent)
		out.WriteString("}")

	case *ast.MatchStatement:
		writeIndent(out, indent)
		out.WriteString("match (")
		formatNode(out, n.Expression, 0)
		out.WriteString(") {\n")
		for _, c := range n.Cases {
			writeIndent(out, indent+1)
			if c.IsDefault {
				out.WriteString("default:\n")
			} else {
				out.WriteString("case ")
				formatNode(out, c.Pattern, 0)
				out.WriteString(":\n")
			}
			for _, stmt := range c.Statements {
				formatNode(out, stmt, indent+2)
				out.WriteString("\n")
			}
		}
		writeIndent(out, indent)
		out.WriteString("}")

	case *ast.TryStatement:
		writeIndent(out, indent)
		out.WriteString("try")
		formatNode(out, n.TryBlock, indent)
		if n.CatchBlock != nil {
			out.WriteString(" catch")
			if n.CatchParam != nil {
				out.WriteString(" (")
				out.WriteString(n.CatchParam.Value)
				out.WriteString(")")
			}
			formatNode(out, n.CatchBlock, indent)
		}

	case *ast.RetryStatement:
		writeIndent(out, indent)
		out.WriteString("retry (")
		formatNode(out, n.Attempts, 0)
		out.WriteString(")")
		formatNode(out, n.RetryBlock, indent)
		if n.CatchBlock != nil {
			out.WriteString(" catch")
			if n.CatchParam != nil {
				out.WriteString(" (")
				out.WriteString(n.CatchParam.Value)
				out.WriteString(")")
			}
			formatNode(out, n.CatchBlock, indent)
		}

	// ─── Functions ───────────────────────────────────────────────────────────
	case *ast.FunctionStatement:
		writeIndent(out, indent)
		out.WriteString("fn ")
		if n.ReceiverName != nil && n.ReceiverType != nil {
			out.WriteString("(")
			out.WriteString(n.ReceiverName.Value)
			out.WriteString(" ")
			out.WriteString(n.ReceiverType.Value)
			out.WriteString(") ")
		}
		out.WriteString(n.Name.Value)
		formatTypeParams(out, n.TypeParameters)
		formatParamsFull(out, n.Parameters)
		if n.ReturnType != nil {
			out.WriteString(": ")
			out.WriteString(n.ReturnType.String())
		}
		formatNode(out, n.Body, indent)

	case *ast.AsyncFunctionStatement:
		writeIndent(out, indent)
		out.WriteString("async fn ")
		out.WriteString(n.Name.Value)
		formatTypeParams(out, n.TypeParameters)
		formatParamsFull(out, n.Parameters)
		if n.ReturnType != nil {
			out.WriteString(": ")
			out.WriteString(n.ReturnType.String())
		}
		formatNode(out, n.Body, indent)

	case *ast.FunctionLiteral:
		out.WriteString("fn")
		formatTypeParams(out, n.TypeParameters)
		formatParamsFull(out, n.Parameters)
		if n.ReturnType != nil {
			out.WriteString(": ")
			out.WriteString(n.ReturnType.String())
		}
		formatNode(out, n.Body, indent)

	case *ast.AsyncFunctionLiteral:
		out.WriteString("async fn")
		formatTypeParams(out, n.TypeParameters)
		formatParamsFull(out, n.Parameters)
		if n.ReturnType != nil {
			out.WriteString(": ")
			out.WriteString(n.ReturnType.String())
		}
		formatNode(out, n.Body, indent)

	case *ast.ArrowFunction:
		if len(n.Parameters) == 1 {
			out.WriteString(n.Parameters[0].Value)
		} else {
			formatParamsFull(out, n.Parameters)
		}
		if n.ReturnType != nil {
			out.WriteString(": ")
			out.WriteString(n.ReturnType.String())
		}
		out.WriteString(" => ")
		formatNode(out, n.Body, indent)

	// ─── Expressions ─────────────────────────────────────────────────────────
	case *ast.CallExpression:
		formatNode(out, n.Function, indent)
		if len(n.TypeArguments) > 0 {
			out.WriteString("[")
			for i, ta := range n.TypeArguments {
				out.WriteString(ta.String())
				if i < len(n.TypeArguments)-1 {
					out.WriteString(", ")
				}
			}
			out.WriteString("]")
		}
		out.WriteString("(")
		// Multiline for big argument lists (>3 args or any arg is a function/block)
		if shouldMultilineArgs(n.Arguments) {
			out.WriteString("\n")
			for i, arg := range n.Arguments {
				writeIndent(out, indent+1)
				formatNode(out, arg, indent+1)
				if i < len(n.Arguments)-1 {
					out.WriteString(",")
				}
				out.WriteString("\n")
			}
			writeIndent(out, indent)
		} else {
			for i, arg := range n.Arguments {
				formatNode(out, arg, indent)
				if i < len(n.Arguments)-1 {
					out.WriteString(", ")
				}
			}
		}
		out.WriteString(")")

	case *ast.InfixExpression:
		formatNode(out, n.Left, indent)
		out.WriteString(" ")
		out.WriteString(n.Operator)
		out.WriteString(" ")
		formatNode(out, n.Right, indent)

	case *ast.PrefixExpression:
		out.WriteString(n.Operator)
		formatNode(out, n.Right, indent)

	case *ast.PostfixExpression:
		formatNode(out, n.Left, indent)
		out.WriteString(n.Operator)

	case *ast.IndexExpression:
		// The Jabline parser stores dot-access (obj.prop) as IndexExpression
		// where Index is always a StringLiteral with the property name.
		// We always re-emit using dot notation since the parser only accepts IDENT after DOT.
		formatNode(out, n.Left, 0)
		out.WriteString(".")
		// Index is always a StringLiteral from the parser — emit just the identifier value
		if strLit, ok := n.Index.(*ast.StringLiteral); ok {
			out.WriteString(strLit.Value)
		} else {
			formatNode(out, n.Index, 0)
		}

	case *ast.ArrayIndexExpression:
		// Bracket access: left[index]
		formatNode(out, n.Left, 0)
		out.WriteString("[")
		formatNode(out, n.Index, 0)
		out.WriteString("]")

	case *ast.InstantiatedExpression:
		formatNode(out, n.Left, 0)
		out.WriteString("[")
		for i, ta := range n.TypeArguments {
			out.WriteString(ta.String())
			if i < len(n.TypeArguments)-1 {
				out.WriteString(", ")
			}
		}
		out.WriteString("]")

	case *ast.TernaryExpression:
		formatNode(out, n.Condition, indent)
		out.WriteString(" ? ")
		formatNode(out, n.TrueValue, indent)
		out.WriteString(" : ")
		formatNode(out, n.FalseValue, indent)

	case *ast.NullishCoalescingExpression:
		formatNode(out, n.Left, indent)
		out.WriteString(" ?? ")
		formatNode(out, n.Right, indent)

	case *ast.OptionalChainingExpression:
		formatNode(out, n.Left, indent)
		out.WriteString("?.")
		formatNode(out, n.Right, indent)

	case *ast.AwaitExpression:
		out.WriteString("await ")
		formatNode(out, n.Value, indent)

	case *ast.SpawnExpression:
		out.WriteString("spawn ")
		formatNode(out, n.Call, indent)

	// ─── Literals ────────────────────────────────────────────────────────────
	case *ast.Identifier:
		out.WriteString(n.Value)

	case *ast.IntegerLiteral:
		out.WriteString(gofmt.Sprintf("%d", n.Value))

	case *ast.FloatLiteral:
		out.WriteString(n.Token.Literal)

	case *ast.StringLiteral:
		out.WriteString("\"")
		out.WriteString(escapeString(n.Value))
		out.WriteString("\"")

	case *ast.Boolean:
		if n.Value {
			out.WriteString("true")
		} else {
			out.WriteString("false")
		}

	case *ast.Null:
		out.WriteString("null")

	case *ast.TemplateLiteral:
		out.WriteString("`")
		for i, part := range n.Parts {
			out.WriteString(part)
			if i < len(n.Expressions) {
				out.WriteString("${")
				formatNode(out, n.Expressions[i], 0)
				out.WriteString("}")
			}
		}
		out.WriteString("`")

	case *ast.ArrayLiteral:
		if len(n.Elements) == 0 {
			out.WriteString("[]")
		} else if shouldMultilineElements(n.Elements) {
			out.WriteString("[\n")
			for i, el := range n.Elements {
				writeIndent(out, indent+1)
				formatNode(out, el, indent+1)
				if i < len(n.Elements)-1 {
					out.WriteString(",")
				}
				out.WriteString("\n")
			}
			writeIndent(out, indent)
			out.WriteString("]")
		} else {
			out.WriteString("[")
			for i, el := range n.Elements {
				formatNode(out, el, indent)
				if i < len(n.Elements)-1 {
					out.WriteString(", ")
				}
			}
			out.WriteString("]")
		}

	case *ast.HashLiteral:
		if len(n.Pairs) == 0 {
			out.WriteString("{}")
		} else {
			out.WriteString("{\n")
			// Sort keys for deterministic output
			type kv struct {
				k, v ast.Expression
			}
			pairs := make([]kv, 0, len(n.Pairs))
			for k, v := range n.Pairs {
				pairs = append(pairs, kv{k, v})
			}
			sort.Slice(pairs, func(i, j int) bool {
				return pairs[i].k.String() < pairs[j].k.String()
			})
			for i, p := range pairs {
				writeIndent(out, indent+1)
				formatNode(out, p.k, indent+1)
				out.WriteString(": ")
				formatNode(out, p.v, indent+1)
				if i < len(pairs)-1 {
					out.WriteString(",")
				}
				out.WriteString("\n")
			}
			writeIndent(out, indent)
			out.WriteString("}")
		}

	// ─── Structs & Interfaces ────────────────────────────────────────────────
	case *ast.StructStatement:
		writeIndent(out, indent)
		out.WriteString("struct ")
		out.WriteString(n.Name.Value)
		formatTypeParams(out, n.TypeParameters)
		out.WriteString(" {\n")
		// Sort fields for deterministic output
		keys := sortedStringKeys(n.Fields)
		for i, k := range keys {
			writeIndent(out, indent+1)
			out.WriteString(k)
			out.WriteString(": ")
			out.WriteString(n.Fields[k].String())
			if i < len(keys)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		writeIndent(out, indent)
		out.WriteString("}")

	case *ast.StructLiteral:
		out.WriteString(n.Name.String())
		out.WriteString(" {\n")
		keys := sortedStringKeysExpr(n.Fields)
		for i, k := range keys {
			writeIndent(out, indent+1)
			out.WriteString(k)
			out.WriteString(": ")
			formatNode(out, n.Fields[k], indent+1)
			if i < len(keys)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		writeIndent(out, indent)
		out.WriteString("}")

	case *ast.InterfaceStatement:
		writeIndent(out, indent)
		out.WriteString("interface ")
		out.WriteString(n.Name.Value)
		formatTypeParams(out, n.TypeParameters)
		out.WriteString(" {\n")
		methKeys := make([]string, 0, len(n.Methods))
		for k := range n.Methods {
			methKeys = append(methKeys, k)
		}
		sort.Strings(methKeys)
		for _, name := range methKeys {
			sig := n.Methods[name]
			writeIndent(out, indent+1)
			out.WriteString(name)
			out.WriteString("(")
			for i, p := range sig.Parameters {
				out.WriteString(p.Value)
				if i < len(sig.Parameters)-1 {
					out.WriteString(", ")
				}
			}
			out.WriteString(")")
			if sig.ReturnType != nil {
				out.WriteString(": ")
				out.WriteString(sig.ReturnType.String())
			}
			out.WriteString("\n")
		}
		writeIndent(out, indent)
		out.WriteString("}")

	// ─── Enums ───────────────────────────────────────────────────────────────
	case *ast.EnumStatement:
		writeIndent(out, indent)
		out.WriteString("enum ")
		out.WriteString(n.Name.Value)
		out.WriteString(" {\n")
		for i, val := range n.Values {
			writeIndent(out, indent+1)
			out.WriteString(val.Value)
			if i < len(n.Values)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		writeIndent(out, indent)
		out.WriteString("}")

	// ─── Modules ─────────────────────────────────────────────────────────────
	case *ast.ImportStatement:
		writeIndent(out, indent)
		out.WriteString(n.String())
		out.WriteString(";")

	case *ast.ExportStatement:
		writeIndent(out, indent)
		formatExport(out, n, indent)

	case *ast.ReExportStatement:
		writeIndent(out, indent)
		out.WriteString(n.String())
		out.WriteString(";")

	// ─── Services ──────────────────────────────────────────────────────────
	case *ast.ServiceStatement:
		writeIndent(out, indent)
		out.WriteString("service ")
		out.WriteString(n.Name.Value)
		out.WriteString(" {\n")
		fieldKeys := sortedStringKeysExpr(n.Fields)
		for _, k := range fieldKeys {
			writeIndent(out, indent+1)
			out.WriteString(k)
			out.WriteString(": ")
			formatNode(out, n.Fields[k], indent+1)
			out.WriteString("\n")
		}
		for _, method := range n.Methods {
			formatNode(out, method, indent+1)
			out.WriteString("\n")
		}
		writeIndent(out, indent)
		out.WriteString("}")

	// ─── Telemetry ──────────────────────────────────────────────────────────
	case *ast.MeterStatement:
		writeIndent(out, indent)
		out.WriteString("meter ")
		formatNode(out, n.Name, 0)
		out.WriteString(n.Operator)
		out.WriteString(";")

	case *ast.TraceStatement:
		writeIndent(out, indent)
		out.WriteString("trace ")
		formatNode(out, n.Name, 0)
		formatNode(out, n.Body, indent)

	// ─── Fallback ───────────────────────────────────────────────────────────
	default:
		out.WriteString(node.String())
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

// formatIf handles if / else if / else chains cleanly without extra indentation.
func formatIf(out *bytes.Buffer, n *ast.IfExpression, indent int) {
	out.WriteString("if (")
	formatNode(out, n.Condition, 0)
	out.WriteString(")")
	formatNode(out, n.Consequence, indent)

	if n.Alternative == nil {
		return
	}

	// Check for else-if chain: { ExpressionStatement{ IfExpression } }
	if len(n.Alternative.Statements) == 1 {
		if exprStmt, ok := n.Alternative.Statements[0].(*ast.ExpressionStatement); ok {
			if innerIf, ok := exprStmt.Expression.(*ast.IfExpression); ok {
				out.WriteString(" else ")
				formatIf(out, innerIf, indent)
				return
			}
		}
	}

	out.WriteString(" else")
	formatNode(out, n.Alternative, indent)
}

// formatExport handles all export variants properly (declaration, default, list, etc.)
func formatExport(out *bytes.Buffer, n *ast.ExportStatement, indent int) {
	switch n.ExportType {
	case ast.EXPORT_DECLARATION:
		out.WriteString("export ")
		if n.IsDefault {
			out.WriteString("default ")
		}
		if n.Statement != nil {
			formatNode(out, n.Statement, indent)
		}
	case ast.EXPORT_DEFAULT:
		out.WriteString("export default ")
		if n.Statement != nil {
			formatNode(out, n.Statement, indent)
		} else {
			out.WriteString(";")
		}
	case ast.EXPORT_LIST:
		out.WriteString("export { ")
		for i, item := range n.ExportList {
			out.WriteString(item.String())
			if i < len(n.ExportList)-1 {
				out.WriteString(", ")
			}
		}
		out.WriteString(" };")
	case ast.EXPORT_ALL:
		out.WriteString("export * from ")
		out.WriteString(n.ModuleName.String())
		out.WriteString(";")
	case ast.EXPORT_NAMED_FROM:
		out.WriteString("export { ")
		for i, item := range n.ExportList {
			out.WriteString(item.String())
			if i < len(n.ExportList)-1 {
				out.WriteString(", ")
			}
		}
		out.WriteString(" } from ")
		out.WriteString(n.ModuleName.String())
		out.WriteString(";")
	case ast.EXPORT_ALL_AS:
		out.WriteString("export * as ")
		out.WriteString(n.NamespaceAlias.String())
		out.WriteString(" from ")
		out.WriteString(n.ModuleName.String())
		out.WriteString(";")
	default:
		out.WriteString(n.String())
	}
}

// formatForInit emits a for-loop init/update without block prefix or trailing newline.
func formatForInit(out *bytes.Buffer, stmt ast.Statement, indent int) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		out.WriteString("let ")
		out.WriteString(s.Name.Value)
		if s.Type != nil {
			out.WriteString(": ")
			out.WriteString(s.Type.String())
		}
		out.WriteString(" = ")
		formatNode(out, s.Value, indent)
	case *ast.AssignmentStatement:
		formatNode(out, s.Left, indent)
		out.WriteString(" = ")
		formatNode(out, s.Value, indent)
	case *ast.ExpressionStatement:
		formatNode(out, s.Expression, indent)
	default:
		out.WriteString(stmt.String())
	}
}

// formatTypeParams writes generic type parameter lists like [T, U].
func formatTypeParams(out *bytes.Buffer, params []*ast.Identifier) {
	if len(params) == 0 {
		return
	}
	out.WriteString("[")
	for i, p := range params {
		out.WriteString(p.Value)
		if i < len(params)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString("]")
}

// formatParamsFull writes a function parameter list including type annotations.
func formatParamsFull(out *bytes.Buffer, params []*ast.Identifier) {
	out.WriteString("(")
	for i, p := range params {
		out.WriteString(p.Value)
		if i < len(params)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(")")
}

// writeIndent writes n levels of 4-space indentation.
func writeIndent(out *bytes.Buffer, indent int) {
	out.WriteString(strings.Repeat("    ", indent))
}

// escapeString escapes special characters inside string literals.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}

// sortedStringKeys returns sorted keys of a map[string]*TypeExpression.
func sortedStringKeys(m map[string]*ast.TypeExpression) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedStringKeysExpr returns sorted keys of a map[string]Expression.
func sortedStringKeysExpr(m map[string]ast.Expression) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// isSimpleLiteral returns true for primitive literals that are fine inline.
func isSimpleLiteral(e ast.Expression) bool {
	switch e.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.Boolean, *ast.Null, *ast.Identifier:
		return true
	}
	return false
}

// shouldMultilineArgs decides if call arguments span multiple lines.
// Only true when any arg itself contains a block/function. Never for simple literals.
func shouldMultilineArgs(args []ast.Expression) bool {
	for _, a := range args {
		switch a.(type) {
		case *ast.FunctionLiteral, *ast.AsyncFunctionLiteral, *ast.ArrowFunction:
			return true
		}
	}
	return false
}

// shouldMultilineElements decides if array literal elements span multiple lines.
// Only expand when elements are complex (nested structures/functions), never for simple literals.
func shouldMultilineElements(elems []ast.Expression) bool {
	for _, e := range elems {
		if !isSimpleLiteral(e) {
			switch e.(type) {
			case *ast.HashLiteral, *ast.ArrayLiteral, *ast.FunctionLiteral:
				return true
			}
		}
	}
	return false
}

// isHeavy returns true for multi-line constructs that need surrounding blank lines.
func isHeavy(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.FunctionStatement,
		*ast.AsyncFunctionStatement,
		*ast.StructStatement,
		*ast.InterfaceStatement,
		*ast.EnumStatement,
		*ast.SwitchStatement,
		*ast.MatchStatement,
		*ast.TryStatement,
		*ast.RetryStatement,
		*ast.ServiceStatement,
		*ast.TraceStatement,
		*ast.WhileStatement,
		*ast.ForStatement,
		*ast.ForEachStatement:
		_ = n
		return true
	case *ast.LetStatement:
		return exprEndsWithBlock(n.Value)
	case *ast.ConstStatement:
		return exprEndsWithBlock(n.Value)
	case *ast.ExpressionStatement:
		return exprEndsWithBlock(n.Expression)
	case *ast.ExportStatement:
		if n.Statement != nil {
			return isHeavy(n.Statement)
		}
	}
	return false
}

// exprEndsWithBlock returns true if an expression ends with a `}` block.
func exprEndsWithBlock(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.IfExpression,
		*ast.FunctionLiteral,
		*ast.AsyncFunctionLiteral,
		*ast.BlockStatement,
		*ast.MatchStatement,
		*ast.TraceStatement:
		_ = n
		return true
	case *ast.WhileStatement,
		*ast.ForStatement,
		*ast.ForEachStatement,
		*ast.SwitchStatement,
		*ast.TryStatement:
		_ = n
		return true
	case *ast.InfixExpression:
		return exprEndsWithBlock(n.Right)
	case *ast.CallExpression:
		if len(n.Arguments) > 0 {
			return exprEndsWithBlock(n.Arguments[len(n.Arguments)-1])
		}
	}
	return false
}
