package lsp

// BuiltinDoc holds the signature and description of a builtin function.
type BuiltinDoc struct {
	Signature   string
	Description string
}

// BuiltinDocs maps every builtin name to its documentation.
var BuiltinDocs = map[string]BuiltinDoc{
	// Core
	"echo": {
		Signature:   "fn echo(...values): null",
		Description: "Prints all arguments to stdout separated by spaces, followed by a newline.",
	},
	"len": {
		Signature:   "fn len(value): int",
		Description: "Returns the length of a string, array, or hash.",
	},
	"type": {
		Signature:   "fn type(value): string",
		Description: "Returns the runtime type name of the value as a string.",
	},
	"toString": {
		Signature:   "fn toString(value): string",
		Description: "Converts any value to its string representation.",
	},
	"parseInt": {
		Signature:   "fn parseInt(s: string): int",
		Description: "Parses a decimal string into an integer. Returns an error on failure.",
	},
	"parseFloat": {
		Signature:   "fn parseFloat(s: string): float",
		Description: "Parses a string into a floating-point number. Returns an error on failure.",
	},
	"input": {
		Signature:   "fn input(prompt?: string): string",
		Description: "Reads a line from stdin. Optional prompt is printed first.",
	},
	"is_error": {
		Signature:   "fn is_error(value): bool",
		Description: "Returns true if the value is an Error object.",
	},
	"Error": {
		Signature:   "fn Error(message: string): Error",
		Description: "Creates a new Error object with the given message.",
	},
	"panic": {
		Signature:   "fn panic(message: string): never",
		Description: "Terminates execution immediately with the given message.",
	},
	// Arrays
	"push": {
		Signature:   "fn push(arr: Array, value): Array",
		Description: "Returns a new array with the value appended at the end.",
	},
	"pop": {
		Signature:   "fn pop(arr: Array): any",
		Description: "Removes and returns the last element of the array.",
	},
	"first": {
		Signature:   "fn first(arr: Array): any",
		Description: "Returns the first element of the array, or null if empty.",
	},
	"last": {
		Signature:   "fn last(arr: Array): any",
		Description: "Returns the last element of the array, or null if empty.",
	},
	"rest": {
		Signature:   "fn rest(arr: Array): Array",
		Description: "Returns a new array with all elements except the first.",
	},
	// Hashes
	"keys": {
		Signature:   "fn keys(hash: Hash): Array",
		Description: "Returns an array of all keys in the hash.",
	},
	"values": {
		Signature:   "fn values(hash: Hash): Array",
		Description: "Returns an array of all values in the hash.",
	},
	"set": {
		Signature:   "fn set(obj: Hash|Array, key, value): any",
		Description: "Sets a key-value pair in a hash, or an index in an array.",
	},
	// Numeric constructors
	"int8": {
		Signature:   "fn int8(value): int8",
		Description: "Converts a numeric value to int8 (8-bit signed integer).",
	},
	"int16": {
		Signature:   "fn int16(value): int16",
		Description: "Converts a numeric value to int16 (16-bit signed integer).",
	},
	"int32": {
		Signature:   "fn int32(value): int32",
		Description: "Converts a numeric value to int32 (32-bit signed integer).",
	},
	"int64": {
		Signature:   "fn int64(value): int64",
		Description: "Converts a numeric value to int64 (64-bit signed integer).",
	},
	"uint8": {
		Signature:   "fn uint8(value): uint8",
		Description: "Converts a numeric value to uint8 (8-bit unsigned integer).",
	},
	"uint16": {
		Signature:   "fn uint16(value): uint16",
		Description: "Converts a numeric value to uint16 (16-bit unsigned integer).",
	},
	"uint32": {
		Signature:   "fn uint32(value): uint32",
		Description: "Converts a numeric value to uint32 (32-bit unsigned integer).",
	},
	"uint64": {
		Signature:   "fn uint64(value): uint64",
		Description: "Converts a numeric value to uint64 (64-bit unsigned integer).",
	},
	"float32": {
		Signature:   "fn float32(value): float32",
		Description: "Converts a numeric value to float32 (32-bit floating point).",
	},
	"float64": {
		Signature:   "fn float64(value): float64",
		Description: "Converts a numeric value to float64 (64-bit floating point).",
	},
	// Concurrency
	"cancel": {
		Signature:   "fn cancel(process): null",
		Description: "Cancels a running async process/channel.",
	},
	// JSON globals
	"parse": {
		Signature:   "fn parse(json: string): any",
		Description: "Parses a JSON string into a Jabline value (hash, array, string, int, bool, null).",
	},
	"stringify": {
		Signature:   "fn stringify(value): string",
		Description: "Serializes a Jabline value to a JSON string.",
	},
}

// SnippetDoc holds a label, detail, and snippet body for completion items.
type SnippetDoc struct {
	Label       string
	Detail      string
	InsertText  string
}

// KeywordSnippets contains snippet completions for Jabline keywords and control structures.
var KeywordSnippets = []SnippetDoc{
	{
		Label:      "fn",
		Detail:     "Function declaration",
		InsertText: "fn ${1:name}(${2:params}) {\n\t${3}\n}",
	},
	{
		Label:      "async fn",
		Detail:     "Async function declaration",
		InsertText: "async fn ${1:name}(${2:params}) {\n\t${3}\n}",
	},
	{
		Label:      "let",
		Detail:     "Variable declaration",
		InsertText: "let ${1:name} = ${2:value};",
	},
	{
		Label:      "const",
		Detail:     "Constant declaration",
		InsertText: "const ${1:name} = ${2:value};",
	},
	{
		Label:      "if",
		Detail:     "If statement",
		InsertText: "if (${1:condition}) {\n\t${2}\n}",
	},
	{
		Label:      "if else",
		Detail:     "If / else statement",
		InsertText: "if (${1:condition}) {\n\t${2}\n} else {\n\t${3}\n}",
	},
	{
		Label:      "for",
		Detail:     "For loop",
		InsertText: "for (let ${1:i} = 0; ${1:i} < ${2:len}; ${1:i} = ${1:i} + 1) {\n\t${3}\n}",
	},
	{
		Label:      "foreach",
		Detail:     "For-each loop",
		InsertText: "for (let ${1:item} in ${2:collection}) {\n\t${3}\n}",
	},
	{
		Label:      "while",
		Detail:     "While loop",
		InsertText: "while (${1:condition}) {\n\t${2}\n}",
	},
	{
		Label:      "return",
		Detail:     "Return statement",
		InsertText: "return ${1:value};",
	},
	{
		Label:      "struct",
		Detail:     "Struct definition",
		InsertText: "struct ${1:Name} {\n\t${2:field}: ${3:type}\n}",
	},
	{
		Label:      "interface",
		Detail:     "Interface definition",
		InsertText: "interface ${1:Name} {\n\t${2:method}(${3:params}): ${4:type}\n}",
	},
	{
		Label:      "try",
		Detail:     "Try / catch block",
		InsertText: "try {\n\t${1}\n} catch (${2:err}) {\n\t${3}\n}",
	},
	{
		Label:      "import",
		Detail:     "Import statement",
		InsertText: "import * as ${1:alias} from \"${2:module}\";",
	},
	{
		Label:      "import named",
		Detail:     "Named import",
		InsertText: "import { ${1:name} } from \"${2:module}\";",
	},
	{
		Label:      "echo",
		Detail:     "Print to stdout",
		InsertText: "echo(${1});",
	},
}
