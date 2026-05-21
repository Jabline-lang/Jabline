import CodeBlock from '../components/common/CodeBlock'
import DocsSidebar from '../components/common/DocsSidebar'

const SIDEBAR = [
  { id: 'structs',      label: 'Structs & Methods' },
  { id: 'generics',     label: 'Generics' },
  { id: 'concurrency',  label: 'Concurrency' },
  { id: 'async-await',  label: 'Async / Await' },
  { id: 'modules',      label: 'Module System' },
  { id: 'type-system',  label: 'Type System' },
  { id: 'operators',    label: 'Operators' },
  { id: 'testing',      label: 'Testing' },
  { id: 'cli',          label: 'CLI Reference' },
  { id: 'lsp',          label: 'LSP & Editor' },
  { id: 'formatter',    label: 'Formatter' },
]

const TYPE_TABLE = [
  ['int', '42', '64-bit signed integer (default)'],
  ['int8 · int16 · int32 · int64', 'int32(255)', 'Sized signed integers'],
  ['uint8 · uint16 · uint32 · uint64', 'uint8(200)', 'Unsigned integers'],
  ['float32 · float64', '3.14', 'IEEE 754 floating point'],
  ['string', '"hello"', 'Immutable UTF-8 string'],
  ['bool', 'true', 'Boolean (true / false)'],
  ['null', 'null', 'Null / nil value'],
  ['Array<T>', 'Array<int>()', 'Generic ordered collection'],
  ['Hash', '{k: v}', 'Dynamic key-value map'],
]

const CLI_CMDS = [
  { cmd: 'jabline run <file.jb>', desc: 'Run a program directly' },
  { cmd: 'jabline run -e "<code>"', desc: 'Execute inline code string' },
  { cmd: 'jabline build <file.jb>', desc: 'Compile to native binary' },
  { cmd: 'jabline test', desc: 'Run all *_test.jb files' },
  { cmd: 'jabline fmt <path>', desc: 'Format source code' },
  { cmd: 'jabline fmt --check', desc: 'Check formatting (CI mode)' },
  { cmd: 'jabline debug <file.jb>', desc: 'Interactive debugger' },
  { cmd: 'jabline lsp', desc: 'Start LSP server (stdio)' },
  { cmd: 'jabline repl', desc: 'Interactive REPL session' },
  { cmd: 'jabline get <pkg>', desc: 'Install a package' },
  { cmd: 'jabline install', desc: 'Install all dependencies' },
  { cmd: 'jabline uninstall <pkg>', desc: 'Remove a package' },
]

export default function Docs({ setCurrentPage }) {
  return (
    <div className="pt-20 min-h-screen">
      <div className="border-b border-jb-800 bg-jb-950/40">
        <div className="container-wide py-12">
          <span className="accent-badge mb-4 inline-block">Reference</span>
          <h1 className="text-jb-50 font-black mb-3">Language Documentation</h1>
          <p className="text-jb-400 text-xl max-w-xl">
            Complete reference for all Jabline language features, CLI commands, and tooling.
          </p>
        </div>
      </div>

      <div className="container-wide py-12">
        <div className="flex gap-12">
          <DocsSidebar sections={SIDEBAR} />
          <main className="flex-1 min-w-0 space-y-16">

            <section id="structs" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Structs & Methods</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock code={`struct User {
    id: int, name: string, email: string, active: bool
}

fn (u User) greet(): string {
    return \`Hello, I'm \${u.name}!\`;
}

fn (u User) isAdult(): bool { return u.age >= 18; }

let user = User { id: 1, name: "Alice", email: "alice@example.com", active: true };
echo(user.greet());   // Hello, I'm Alice!
echo(user.isAdult()); // true

// Nested structs
struct Company { name: string, ceo: User }
let company = Company { name: "AcmeCorp", ceo: user };
echo(company.ceo.name); // Alice`} language="jabline" />
            </section>

            <section id="generics" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Generics</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock code={`fn identity<T>(value: T): T { return value; }
fn first<T>(arr: Array<T>): T { return arr[0]; }

fn map<T, U>(arr: Array<T>, f: fn(T): U): Array<U> {
    let result = Array<U>();
    for item in arr { result.push(f(item)); }
    return result;
}

struct Box<T> { value: T, label: string }
fn (b Box<T>) unwrap(): T { return b.value; }

struct Pair<A, B> { first: A, second: B }

// Usage
let doubled = map([1,2,3,4,5], (x) => x * 2); // [2,4,6,8,10]
let box = Box { value: 42, label: "answer" };
echo(box.unwrap()); // 42`} language="jabline" />
            </section>

            <section id="concurrency" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Concurrency</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <p className="text-jb-400 mb-6">CSP model — goroutines communicate via channels. No shared mutable state.</p>
              <CodeBlock code={`// Basic channel
let ch = make_chan();
spawn fn() { send(ch, heavyComputation()); }();
let value = recv(ch);

// Worker pool
fn worker(id: int, jobs, results) {
    while (true) {
        let job = recv(jobs);
        if (job == null) { break; }
        send(results, job * job);
    }
}

fn main() {
    let jobs = make_chan();
    let results = make_chan();
    for (let i = 1; i <= 4; i++) { spawn worker(i, jobs, results); }
    for (let j = 1; j <= 9; j++) { send(jobs, j); }
    for (let k = 0; k < 9; k++) { echo(recv(results)); }
}

// Timeout
let result = recv_timeout(ch, 5000); // 5 second timeout
if (result == null) { echo("Timed out!"); }`} language="jabline" />
            </section>

            <section id="async-await" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Async / Await</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock code={`async fn fetchUser(id: int): Hash {
    let resp = await http.get(\`https://api.example.com/users/\${id}\`);
    return json.parse(resp.body);
}

// Parallel await
async fn main() {
    let [user, posts] = await all([
        fetchUser(1),
        fetchPosts(1)
    ]);
    echo(\`\${user.name} has \${len(posts)} posts\`);
}

// Error handling
async fn safeGet(url: string): string {
    try {
        let resp = await http.get(url);
        return resp.body;
    } catch (err) {
        echo(\`Failed: \${err}\`);
        return "";
    }
}`} language="jabline" />
            </section>

            <section id="modules" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Module System</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock code={`// Stdlib
import * as http   from "std/net/http";
import * as json   from "std/encoding/json";
import * as crypto from "std/crypto";
import * as io     from "std/sys/io";
import * as time   from "std/time/datetime";
import * as math   from "std/math";

// Named imports
import { parse, stringify } from "std/encoding/json";
import { sha256, uuid }     from "std/crypto";

// Local files
import * as utils  from "./utils";
import { Router }  from "./router";

// Registry packages
import * as colors from "colors@1.2.0";
import * as dotenv from "dotenv@0.3.0";

// Exporting
export fn add(a, b) { return a + b; }
export const VERSION = "1.0.0";
export struct Config { port: int, host: string }`} language="jabline" />
            </section>

            <section id="type-system" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Type System</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <div className="overflow-x-auto mb-6">
                <table className="jb-table">
                  <thead><tr><th>Type</th><th>Example</th><th>Description</th></tr></thead>
                  <tbody>
                    {TYPE_TABLE.map(([type, ex, desc]) => (
                      <tr key={type}>
                        <td><code>{type}</code></td>
                        <td><code className="text-jb-300 text-xs">{ex}</code></td>
                        <td>{desc}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <CodeBlock code={`let n = int(3.14);       // 3
let f = float64(42);     // 42.0
let s = string(100);     // "100"

echo(typeof(x));         // "int"
echo(x is int);          // true

// Union types
fn process(value: int | string) {
    if (value is int) { echo("Number: " + string(value)); }
    else { echo("String: " + value); }
}

// Nullable
let maybe: string? = null;
let safe = maybe ?? "default";`} language="jabline" />
            </section>

            <section id="operators" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Operators</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock code={`// Arithmetic: + - * / % ** (power)
// Comparison: == != < <= > >=
// Logical:    && || !
// Bitwise:    & | ^ ~ << >>
// Assignment: = += -= *= /= ++ --

// Pipe operator
let result = data |> transform |> validate |> format;

// Optional chaining
let city = user?.address?.city;

// Nullish coalescing
let name = user.name ?? "Anonymous";

// Spread
let arr2 = [...arr1, 4, 5];
let obj2 = {...config, port: 80};

// Template literals
let msg = \`Hello \${name}, you are \${age} years old!\`;

// Ternary
let label = age >= 18 ? "adult" : "minor";`} language="jabline" />
            </section>

            <section id="testing" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Testing</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <p className="text-jb-400 mb-4">Test files end in <code className="text-accent text-sm">_test.jb</code>. Run with <code className="text-accent text-sm">jabline test</code>.</p>
              <CodeBlock code={`// math_test.jb
import { add } from "./math";
import * as test from "std/testing";

test.describe("Math", fn() {
    test.it("adds numbers", fn() {
        test.expect(add(2, 3)).toBe(5);
        test.expect(add(-1, 1)).toBe(0);
    });

    test.it("async test", async fn() {
        let data = await fetchUser(1);
        test.expect(data.id).toBe(1);
        test.expect(data.name).not.toBeNull();
    });
});`} language="jabline" filename="math_test.jb" />
              <div className="mt-4">
                <CodeBlock code={`jabline test              # Run all tests
jabline test math_test.jb # Run specific file
jabline test --watch       # Watch mode
jabline test --coverage    # Coverage report`} language="bash" />
              </div>
            </section>

            <section id="cli" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">CLI Reference</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <div className="overflow-x-auto">
                <table className="jb-table">
                  <thead><tr><th>Command</th><th>Description</th></tr></thead>
                  <tbody>
                    {CLI_CMDS.map(({ cmd, desc }) => (
                      <tr key={cmd}><td><code>{cmd}</code></td><td>{desc}</td></tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>

            <section id="lsp" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">LSP & Editor Support</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <div className="grid sm:grid-cols-2 gap-3 mb-6">
                {['Autocomplete', 'Hover docs', 'Go to definition', 'Find references',
                  'Inline diagnostics', 'Code formatting', 'Rename symbol', 'Signature help'
                ].map(f => (
                  <div key={f} className="flex items-center gap-2 text-jb-300 text-sm">
                    <span className="text-success">✓</span>{f}
                  </div>
                ))}
              </div>
              <CodeBlock code={`// .vscode/settings.json
{
  "jabline.lspPath": "jabline",
  "editor.formatOnSave": true,
  "[jabline]": {
    "editor.defaultFormatter": "jabline.jabline-vscode"
  }
}`} language="json" filename=".vscode/settings.json" />
            </section>

            <section id="formatter" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">Formatter</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock code={`jabline fmt main.jb        # Format a file
jabline fmt ./...          # Format all files
jabline fmt --check ./...  # CI check (exit 1 if unformatted)
jabline fmt --diff main.jb # Show diff`} language="bash" />
            </section>

          </main>
        </div>
      </div>
    </div>
  )
}
