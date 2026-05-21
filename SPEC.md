# Jabline Language Specification v0.6.0

## 1. Lexical Structure

### 1.1 Character Set
Source code is UTF-8 encoded text.

### 1.2 Comments
```ebnf
comment = line_comment | block_comment ;
line_comment = "//" { character } ;
block_comment = "/*" { character } "*/" ;
```

### 1.3 Identifiers
```ebnf
identifier = (letter | "_") { letter | digit | "_" } ;
letter = "A" … "Z" | "a" … "z" ;
digit = "0" … "9" ;
```

### 1.4 Keywords
```
echo      let       const     null      if        else
for       while     fn        return    struct    break
continue  in        switch    case      default   try
catch     throw     async     await     import    export
from      true      false     enum      as        retry
service   spawn     interface match     meter     trace
```

### 1.5 Built-in Type Keywords
```
int       int8      int16     int32     int64
uint8     uint16    uint32    uint64
float     float32   float64
string    bool
```

### 1.6 Literals

```ebnf
literal = int_literal | float_literal | string_literal
        | template_literal | bool_literal | null_literal ;

int_literal = decimal_literal | hex_literal | octal_literal | binary_literal ;
decimal_literal = digit { digit } ;
hex_literal = "0" ("x" | "X") hex_digit { hex_digit } ;
octal_literal = "0" ("o" | "O") octal_digit { octal_digit } ;
binary_literal = "0" ("b" | "B") binary_digit { binary_digit } ;

float_literal = digits "." digits [ "e" | "E" [ "+" | "-" ] digits ] ;

string_literal = double_quoted | single_quoted ; (* unicode with escape sequences *)
double_quoted = '"' { unicode_char | escape } '"' ;

template_literal = "`" { template_char | "${" expression "}" } "`" ;

bool_literal = "true" | "false" ;
null_literal = "null" ;
```

### 1.7 Operators

| Category   | Tokens                                                                  |
|------------|-------------------------------------------------------------------------|
| Arithmetic | `+` `-` `*` `/` `%` `++` `--`                                          |
| Assignment | `=` `+=` `-=` `*=` `/=`                                                |
| Bitwise    | `&` `\|` `^` `~` `<<` `>>`                                             |
| Logical    | `!` `&&` `\|\|`                                                        |
| Comparison | `<` `>` `<=` `>=` `==` `!=`                                            |
| Other      | `?` `=>` `<-` `??` `?.` `\|>`                                          |

### 1.8 Delimiters
```
(  )  {  }  [  ]  ,  ;  .  :
```

### 1.9 Precedence (lowest to highest)

| Level              | Operators                          | Associativity |
|--------------------|------------------------------------|---------------|
| pipe               | `\|>`                              | left          |
| nullish coalescing | `??`                               | left          |
| ternary            | `? ... :`                          | right         |
| channel send       | `<-`                               | left          |
| logical OR         | `\|\|`                             | left          |
| logical AND        | `&&`                               | left          |
| equals             | `==` `!=`                          | left          |
| less/greater       | `<` `>` `<=` `>=`                  | left          |
| sum                | `+` `-` `\|` `^`                   | left          |
| product            | `*` `/` `%` `&` `<<` `>>`          | left          |
| prefix             | `!` `-` `~` `<-` `await` `spawn`   | n/a           |
| call               | `(...)` `{...}` (struct literal)    | left          |
| index              | `.` `[...]`                        | left          |
| optional chaining  | `?.`                               | left          |
| postfix            | `++` `--`                          | n/a           |

---

## 2. Types

### 2.1 Built-in Types
```
int       int8     int16     int32     int64     (* signed integers *)
uint8     uint16   uint32    uint64              (* unsigned integers *)
float     float32  float64                       (* floating point *)
string                                          (* text *)
bool                                            (* boolean *)
null                                            (* null type *)
```

### 2.2 Type Expressions
```ebnf
type_expr = ( builtin_type | identifier ) [ "<" type_expr { "," type_expr } ">" ] ;

(* Examples *)
string          (* simple type *)
Array<int>      (* generic type *)
Result<string, int>  (* multi-parameter generic *)
```

### 2.3 Type Parameters (in declarations)
```ebnf
type_params = identifier { "," identifier } ;

(* Example *)
fn identity<T>(value: T): T { return value; }
struct Pair<T, U> { first: T, second: U }
```

---

## 3. Complete EBNF Grammar

```ebnf
program = { statement } ;

statement = let_stmt
          | const_stmt
          | return_stmt
          | echo_stmt
          | while_stmt
          | for_stmt
          | foreach_stmt
          | function_stmt
          | async_function_stmt
          | struct_stmt
          | interface_stmt
          | break_stmt
          | continue_stmt
          | try_stmt
          | retry_stmt
          | service_stmt
          | throw_stmt
          | switch_stmt
          | import_stmt
          | export_stmt
          | enum_stmt
          | match_stmt
          | meter_stmt
          | trace_stmt
          | assign_stmt
          | expr_stmt
          | ";" ;

block = "{" { statement } "}" ;

(* ---------- Variable Declarations ---------- *)

let_stmt = "let" identifier [ ":" type_expr ] "=" expression ";" ;
const_stmt = "const" identifier [ ":" type_expr ] "=" expression ";" ;

(* ---------- Return ---------- *)

return_stmt = "return" [ expression ] ";" ;

(* ---------- Echo (Print) ---------- *)

echo_stmt = "echo" "(" expression_list ")" ";" ;

(* ---------- Control Flow ---------- *)

while_stmt = "while" "(" expression ")" block ;

for_stmt = "for" "(" [ for_init ] ";" [ expression ] ";" [ for_update ] ")" block ;
for_init = let_stmt | assign_stmt | expr_stmt ;
for_update = assign_stmt | expr_stmt ;

foreach_stmt = "for" "(" identifier "in" expression ")" block ;

if_expr = "if" "(" expression ")" block [ "else" ( if_expr | block ) ] ;

switch_stmt = "switch" "(" expression ")" "{" { case_clause } [ default_clause ] "}" ;
case_clause = "case" expression ":" { statement } ;
default_clause = "default" ":" { statement } ;

match_stmt = "match" "(" expression ")" "{" { match_case } "}" ;
match_case = ( "case" pattern | "default" ) ":" { statement } ;
pattern = identifier         (* type match *)
         | array_literal     (* structural match *)
         | expression ;      (* equality match *)

(* ---------- Functions ---------- *)

function_stmt = "fn" [ "(" identifier identifier ")" ]           (* receiver *)
                identifier [ "<" type_params ">" ]                (* name + generics *)
                "(" func_params ")" [ ":" type_expr ]            (* params + return *)
                block ;                                          (* body *)

async_function_stmt = "async" "fn" identifier
                      [ "<" type_params ">" ]
                      "(" func_params ")" [ ":" type_expr ]
                      block ;

function_literal = "fn" [ "<" type_params ">" ]
                   "(" func_params ")" [ ":" type_expr ]
                   block ;

async_function_literal = "async" "fn"
                         "(" func_params ")" [ ":" type_expr ]
                         block ;

arrow_function = ( identifier | "(" func_params ")" )
                 [ ":" type_expr ] "=>" expression ;

func_params = [ parameter { "," parameter } ] ;
parameter = identifier [ ":" type_expr ] ;

(* ---------- Structs ---------- *)

struct_stmt = "struct" identifier [ "<" type_params ">" ]
              "{" [ struct_field { "," struct_field } ] "}" ;
struct_field = identifier ":" type_expr ;

struct_literal = expression "{"
                 [ identifier ":" expression { "," identifier ":" expression } ]
                 "}" ;

(* ---------- Interfaces ---------- *)

interface_stmt = "interface" identifier [ "<" type_params ">" ]
                 "{" [ method_sig { "," method_sig } ] "}" ;
method_sig = identifier "(" func_params ")" [ ":" type_expr ] ;

(* ---------- Enums ---------- *)

enum_stmt = "enum" identifier "{" identifier { "," identifier } "}" ;

(* ---------- Services ---------- *)

service_stmt = "service" identifier "{"
               { identifier ":" expression ";"
               | function_stmt }
               "}" ;

(* ---------- Error Handling ---------- *)

try_stmt = "try" block [ "catch" [ "(" identifier ")" ] block ] ;
throw_stmt = "throw" expression ";" ;

retry_stmt = "retry" "(" expression ")" block
             [ "catch" [ "(" identifier ")" ] block ] ;

(* ---------- Concurrency ---------- *)

spawn_expr = "spawn" call_expression ;
await_expr = "await" expression ;

(* ---------- Telemetry ---------- *)

meter_stmt = "meter" expression [ "++" | "--" ] ";" ;
trace_stmt = "trace" expression block ;

(* ---------- Modules ---------- *)

import_stmt = "import" import_spec "from" string_literal ;

import_spec = identifier                            (* default *)
            | "{" import_items "}"                  (* named *)
            | "*" "as" identifier                   (* namespace *)
            | identifier "," "{" import_items "}"   (* default + named *)
            | string_literal [ "as" identifier ]    (* side-effect + alias *)
            ;

import_items = import_item { "," import_item } ;
import_item = identifier [ "as" identifier ] ;

export_stmt = "export" ( export_spec ) ;
export_spec = statement                         (* inline export *)
             | "default" statement              (* default export *)
             | "{" export_items "}"             (* named exports *)
             | "{" export_items "}" "from" string_literal  (* re-export *)
             | "*" [ "as" identifier ] "from" string_literal (* namespace re-export *)
             ;

export_items = export_item { "," export_item } ;
export_item = identifier [ "as" identifier ] ;

(* ---------- Assignment ---------- *)

assign_stmt = ( index_expr | identifier )
              ( "=" | "+=" | "-=" | "*=" | "/=" )
              expression ";" ;

(* ---------- Expression Statement ---------- *)

expr_stmt = expression ";" ;

(* ---------- Primary Expressions ---------- *)

expression = pipe_expr ;

pipe_expr = nullish_coalescing_expr { "|>" nullish_coalescing_expr } ;
nullish_coalescing_expr = ternary_expr { "??" ternary_expr } ;
ternary_expr = logical_or_expr [ "?" expression ":" ternary_expr ] ;
logical_or_expr = logical_and_expr { "||" logical_and_expr } ;
logical_and_expr = equals_expr { "&&" equals_expr } ;
equals_expr = comparison_expr { ( "==" | "!=" ) comparison_expr } ;
comparison_expr = sum_expr { ( "<" | ">" | "<=" | ">=" ) sum_expr } ;
sum_expr = product_expr { ( "+" | "-" | "|" | "^" ) product_expr } ;
product_expr = prefix_expr { ( "*" | "/" | "%" | "&" | "<<" | ">>" ) prefix_expr } ;

prefix_expr = ( "!" | "-" | "~" | "<-" ) postfix_expr
            | await_expr
            | spawn_expr
            | postfix_expr ;

postfix_expr = call_expr { ( "++" | "--" ) } ;

call_expr = primary_expr
            { "(" expression_list ")"                     (* function call *)
            | "." identifier                              (* property access *)
            | "[" expression "]"                          (* index access *)
            | "?." identifier                             (* optional chaining *)
            | "{" struct_literal_fields "}"               (* struct literal *)
            | "<-" expression                             (* channel send *)
            } ;

primary_expr = identifier
             | literal
             | if_expr
             | function_literal
             | async_function_literal
             | arrow_function
             | "(" expression ")"
             | "(" func_params ")" [ ":" type_expr ] "=>" expression
             | "[" [ expression_list ] "]"
             | "{" [ key_value_pairs ] "}"
             | type_keyword [ "(" expression ")" ]     (* type cast *)
             | identifier "<" type_expr { "," type_expr } ">"  (* generic instantiation *)
             ;

expression_list = expression { "," expression } ;
key_value_pairs = ( expression ":" expression ) { "," ( expression ":" expression ) } ;
```

---

## 4. Standard Library

### 4.1 Built-in Functions (always available)
```
len(x)           -> int       (* length of array, string, or hash *)
type(x)          -> string    (* type name of value *)
toString(x)      -> string    (* convert to string *)
parseInt(s)      -> int       (* parse integer from string *)
parseFloat(s)    -> float     (* parse float from string *)
echo(...)        -> void      (* print values to stdout *)
push(arr, item)  -> void      (* append to array *)
pop(arr)         -> any       (* remove and return last element *)
rest(arr)        -> array     (* array without first element *)
first(arr)       -> any       (* first element *)
last(arr)        -> any       (* last element *)
keys(hash)       -> array     (* keys of hash *)
values(hash)     -> array     (* values of hash *)
is_error(x)      -> bool      (* check if value is an error *)
Error(msg)       -> error     (* create error object *)
panic(msg)       -> void      (* panic with message *)
input()          -> string    (* read line from stdin *)
cancel(proc)     -> void      (* cancel a process *)
make_chan()      -> channel   (* create a channel *)
send(ch, val)    -> void      (* send value to channel *)
recv(ch)         -> any       (* receive value from channel *)
int8(x)          -> int8      (* numeric conversion *)
int16(x)         -> int16
int32(x)         -> int32
int64(x)         -> int64
uint8(x)         -> uint8
uint16(x)        -> uint16
uint32(x)        -> uint32
uint64(x)        -> uint64
float32(x)       -> float32
float64(x)       -> float64   (* same as float *)
gc()             -> void      (* trigger garbage collection *)
memoryStats()    -> hash      (* memory statistics *)
```

### 4.2 Standard Modules
```
import * as http from "std/net/http"
import { parse } from "std/encoding/json"
import * as crypto from "std/crypto"
import * as fs from "std/sys/io"
import { describe, it, assertEqual } from "std/testing/assert"
```

See the full stdlib documentation at `docs/` for the complete module catalog.

---

## 5. Execution Model

1. **Lexing** — Source text is tokenized into a stream of tokens.
2. **Parsing** — A Pratt (top-down operator precedence) parser builds an AST.
3. **Type Checking** — Optional AOT type checker validates type constraints.
4. **Compilation** — The AST is compiled to bytecode for the Jabline Virtual Machine (JBVM).
   - Constant folding optimizes compile-time evaluable expressions.
   - Dead code elimination prunes unreachable branches.
5. **Execution** — The JBVM executes the bytecode on a stack-based architecture.
   - Generational approach: interpreted bytecode (no JIT).
   - Concurrency via `spawn`/channels (CSP model) and `async`/`await`.
   - FFI via `purego` for calling C dynamic libraries.
