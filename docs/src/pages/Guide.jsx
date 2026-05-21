import CodeBlock from '../components/common/CodeBlock'
import DocsSidebar from '../components/common/DocsSidebar'

const SIDEBAR = [
  { id: 'installation',   label: '1. Installation' },
  { id: 'first-program',  label: '2. First Program' },
  { id: 'variables',      label: '3. Variables & Types' },
  { id: 'functions',      label: '4. Functions' },
  { id: 'control-flow',   label: '5. Control Flow' },
  { id: 'arrays-maps',    label: '6. Arrays & Maps' },
  { id: 'error-handling', label: '7. Error Handling' },
  { id: 'imports',        label: '8. Imports & Modules' },
]

export default function Guide({ setCurrentPage }) {
  return (
    <div className="pt-20 min-h-screen">
      {/* Page header */}
      <div className="border-b border-jb-800 bg-jb-950/40">
        <div className="container-wide py-12">
          <span className="accent-badge mb-4 inline-block">Getting Started</span>
          <h1 className="text-jb-50 font-black mb-3">Quick Start Guide</h1>
          <p className="text-jb-400 text-xl max-w-xl">
            Install Jabline and write your first program in under 5 minutes.
          </p>
        </div>
      </div>

      <div className="container-wide py-12">
        <div className="flex gap-12">
          <DocsSidebar sections={SIDEBAR} />

          {/* Content */}
          <main className="flex-1 min-w-0 space-y-16">

            {/* ── 1. Installation ─────────────────────────── */}
            <section id="installation" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">1. Installation</h2>
              <div className="h-px bg-jb-800 mb-8" />

              <p className="text-jb-400 mb-6">
                Jabline is distributed as a single binary. You can install it via <code className="text-accent">go install</code> or by building from source.
              </p>

              <h3 className="text-jb-200 font-semibold mb-3 text-base">Prerequisites</h3>
              <div className="jb-card p-4 mb-6">
                <div className="flex items-center gap-3">
                  <span className="w-2 h-2 rounded-full bg-accent" />
                  <span className="text-jb-300 text-sm">Go 1.21 or higher — <a href="https://go.dev/dl/" target="_blank" rel="noopener noreferrer" className="text-accent">go.dev/dl</a></span>
                </div>
              </div>

              <h3 className="text-jb-200 font-semibold mb-3 text-base">Option A — Go Install (recommended)</h3>
              <CodeBlock
                code={`go install github.com/Jabline-lang/Jabline@latest`}
                language="bash"
              />

              <h3 className="text-jb-200 font-semibold mb-3 text-base mt-6">Option B — Build from Source</h3>
              <CodeBlock
                code={`git clone https://github.com/Jabline-lang/Jabline.git
cd Jabline

# Linux / macOS
go build -o jabline .
sudo mv jabline /usr/local/bin/

# Windows (PowerShell)
go build -o jabline.exe .`}
                language="bash"
              />

              <h3 className="text-jb-200 font-semibold mb-3 text-base mt-6">Verify Installation</h3>
              <CodeBlock code={`jabline --version\n# jabline v0.6.0`} language="bash" />
            </section>

            {/* ── 2. First Program ────────────────────────── */}
            <section id="first-program" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">2. First Program</h2>
              <div className="h-px bg-jb-800 mb-8" />

              <p className="text-jb-400 mb-4">
                Create a file called <code className="text-accent bg-jb-850 px-1.5 py-0.5 rounded text-sm">hello.jb</code>:
              </p>
              <CodeBlock
                code={`fn main() {
    let name = "World";
    echo(\`Hello, \${name}!\`);
    
    // Numbers and math
    let x = 10;
    let y = 32;
    echo(\`\${x} + \${y} = \${x + y}\`);
}`}
                language="jabline"
                filename="hello.jb"
              />

              <p className="text-jb-400 mt-6 mb-4">Run it:</p>
              <CodeBlock code={`jabline run hello.jb\n# Hello, World!\n# 10 + 32 = 42`} language="bash" />

              <p className="text-jb-400 mt-6 mb-4">Run code inline without a file:</p>
              <CodeBlock code={`jabline run -e 'echo("Hello inline!")'`} language="bash" />

              <p className="text-jb-400 mt-6 mb-4">Start the interactive REPL:</p>
              <CodeBlock code={`jabline repl\n# Jabline REPL v0.6.0\n# Type .exit to quit\n> echo("hello from REPL")\nhello from REPL`} language="bash" />
            </section>

            {/* ── 3. Variables & Types ─────────────────────── */}
            <section id="variables" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">3. Variables & Types</h2>
              <div className="h-px bg-jb-800 mb-8" />

              <p className="text-jb-400 mb-6">
                Jabline uses <code className="text-accent text-sm">let</code> for variables (mutable) and <code className="text-accent text-sm">const</code> for constants.
                The type checker infers types from values — no need to annotate everything.
              </p>

              <CodeBlock
                code={`// Type inference
let name    = "Alice";    // string
let age     = 30;         // int
let price   = 9.99;       // float64
let active  = true;       // bool

// Explicit type annotations
let count:  int     = 100;
let weight: float64 = 72.5;
let data:   string  = "payload";

// Constants (immutable)
const PI       = 3.14159265;
const MAX_SIZE = 1024;
const APP_NAME = "MyApp";

// Multiple assignment
let x, y = 10, 20;`}
                language="jabline"
              />

              <div className="mt-8 overflow-x-auto">
                <table className="jb-table">
                  <thead>
                    <tr>
                      <th>Type</th>
                      <th>Example</th>
                      <th>Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    {[
                      ['int', '42', '64-bit signed integer'],
                      ['int8 · int16 · int32 · int64', 'int32(255)', 'Sized integers'],
                      ['uint8 · uint16 · uint32 · uint64', 'uint8(200)', 'Unsigned integers'],
                      ['float32 · float64', '3.14', 'IEEE 754 floating point'],
                      ['string', '"hello"', 'UTF-8 string'],
                      ['bool', 'true / false', 'Boolean value'],
                      ['null', 'null', 'Null/nil value'],
                      ['Array<T>', 'Array<int>()', 'Generic ordered collection'],
                      ['Hash', '{key: val}', 'Key-value map'],
                    ].map(([type, ex, desc]) => (
                      <tr key={type}>
                        <td><code>{type}</code></td>
                        <td><code className="text-jb-300">{ex}</code></td>
                        <td>{desc}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>

            {/* ── 4. Functions ────────────────────────────── */}
            <section id="functions" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">4. Functions</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock
                code={`// Basic function
fn add(a: int, b: int): int {
    return a + b;
}

// Inferred parameter types
fn greet(name) {
    echo(\`Hello, \${name}!\`);
}

// Multiple return values
fn divmod(a: int, b: int): (int, int) {
    return (a / b, a % b);
}

// Arrow functions (lambdas)
let square = (x) => x * x;
let double = (x: int): int => x * 2;

// Higher order functions
fn apply(f, value) {
    return f(value);
}
echo(apply(square, 7)); // 49

// Closures
fn makeAdder(n: int) {
    return fn(x: int): int {
        return x + n;
    };
}
let add5 = makeAdder(5);
echo(add5(10)); // 15

// Variadic (rest params)
fn sum(...nums: int): int {
    let total = 0;
    for n in nums { total = total + n; }
    return total;
}
echo(sum(1, 2, 3, 4, 5)); // 15`}
                language="jabline"
              />
            </section>

            {/* ── 5. Control Flow ─────────────────────────── */}
            <section id="control-flow" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">5. Control Flow</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock
                code={`// ── If / else if / else ──────────────────────
let age = 20;

if (age < 13) {
    echo("Child");
} else if (age < 18) {
    echo("Teen");
} else {
    echo("Adult");
}

// ── Ternary expression ────────────────────────
let label = age >= 18 ? "adult" : "minor";

// ── Switch ────────────────────────────────────
switch (label) {
    case "adult":  echo("Can vote!");
    case "minor":  echo("Not yet.");
    default:       echo("Unknown");
}

// ── For loop (C-style) ────────────────────────
for (let i = 0; i < 5; i++) {
    echo(i);
}

// ── For-in (range over array) ─────────────────
let fruits = ["apple", "banana", "cherry"];
for fruit in fruits {
    echo(fruit);
}

// ── For-in with index ─────────────────────────
for (idx, fruit) in fruits {
    echo(\`\${idx}: \${fruit}\`);
}

// ── While ─────────────────────────────────────
let count = 0;
while (count < 3) {
    echo(\`count = \${count}\`);
    count++;
}

// ── Break / Continue ──────────────────────────
for (let i = 0; i < 10; i++) {
    if (i == 3) { continue; }
    if (i == 7) { break; }
    echo(i);
}`}
                language="jabline"
              />
            </section>

            {/* ── 6. Arrays & Maps ────────────────────────── */}
            <section id="arrays-maps" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">6. Arrays & Maps</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock
                code={`// ── Arrays ───────────────────────────────────
let nums = Array<int>();
nums.push(10);
nums.push(20);
nums.push(30);

echo(nums[0]);        // 10
echo(len(nums));      // 3

// Array literal
let colors = ["red", "green", "blue"];

// Array methods
colors.push("yellow");
let popped = colors.pop();
let idx = colors.indexOf("green"); // 1
let sliced = colors.slice(0, 2);   // ["red", "green"]

// ── Maps / Hashes ─────────────────────────────
let config = {
    host: "localhost",
    port: 3000,
    debug: true
};

echo(config.host);        // localhost
echo(config["port"]);     // 3000

config.timeout = 30;      // add new key

// Check if key exists
if (config.debug) {
    echo("Debug mode on");
}

// Iterate over a map
for (key, value) in config {
    echo(\`\${key}: \${value}\`);
}`}
                language="jabline"
              />
            </section>

            {/* ── 7. Error Handling ───────────────────────── */}
            <section id="error-handling" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">7. Error Handling</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock
                code={`// ── Try / catch / finally ────────────────────
try {
    let result = riskyOperation();
    echo(result);
} catch (err) {
    echo("Error: " + err);
} finally {
    echo("Cleanup done.");
}

// ── Throwing errors ───────────────────────────
fn divide(a: int, b: int): float64 {
    if (b == 0) {
        throw "Division by zero";
    }
    return a / b;
}

try {
    echo(divide(10, 0));
} catch (e) {
    echo(\`Caught: \${e}\`);
}

// ── Error propagation with ? operator ─────────
fn readConfig(path: string): Hash {
    let content = io.readFile(path)?; // throws on error
    return json.parse(content)?;
}

// ── Optional chaining ─────────────────────────
let city = user?.address?.city;       // null-safe
let name = user.name ?? "Anonymous";  // nullish coalescing`}
                language="jabline"
              />
            </section>

            {/* ── 8. Imports & Modules ────────────────────── */}
            <section id="imports" className="scroll-mt-24">
              <h2 className="text-jb-50 font-bold mb-2">8. Imports & Modules</h2>
              <div className="h-px bg-jb-800 mb-8" />
              <CodeBlock
                code={`// ── Import entire module ─────────────────────
import * as http   from "std/net/http";
import * as json   from "std/encoding/json";
import * as crypto from "std/crypto";
import * as io     from "std/sys/io";

// ── Named imports ─────────────────────────────
import { parse, stringify } from "std/encoding/json";
import { sha256, uuid }     from "std/crypto";

// ── Local file imports ────────────────────────
import * as utils from "./utils";
import { helper } from "./lib/helpers";

// ── Package imports (from registry) ───────────
import * as colors from "colors@1.2.0";

// ── Usage ─────────────────────────────────────
fn main() {
    let data = json.parse('{"key": "value"}');
    let id   = crypto.uuid();
    let hash = sha256("secret");
    
    let server = http.createServer();
    server.listen(8080);
}`}
                language="jabline"
              />

              <div className="mt-8 p-6 jb-card border-accent/20">
                <h3 className="text-jb-100 font-semibold mb-4">Next Steps</h3>
                <div className="grid sm:grid-cols-2 gap-3">
                  {[
                    { label: 'Full Documentation', page: 'docs', desc: 'Structs, generics, concurrency and more' },
                    { label: 'Standard Library', page: 'stdlib', desc: 'Browse all built-in modules' },
                    { label: 'Examples', page: 'examples', desc: 'Real-world code snippets' },
                    { label: 'Playground', page: 'playground', desc: 'Try Jabline in your browser' },
                  ].map(item => (
                    <button
                      key={item.page}
                      onClick={() => setCurrentPage(item.page)}
                      className="text-left p-3 rounded-lg bg-jb-850 border border-jb-700 hover:border-accent/40 transition-all group"
                    >
                      <div className="text-accent text-sm font-semibold group-hover:text-accent-light transition-colors">{item.label} →</div>
                      <div className="text-jb-500 text-xs mt-0.5">{item.desc}</div>
                    </button>
                  ))}
                </div>
              </div>
            </section>

          </main>
        </div>
      </div>
    </div>
  )
}
