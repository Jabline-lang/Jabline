import { useState, useEffect, useRef, useCallback } from 'react'
import Editor from 'react-simple-code-editor'
import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'

// Register javascript for jabline
if (!hljs.getLanguage('javascript')) {
  hljs.registerLanguage('javascript', javascript)
}

/* ── Snippet Library ─────────────────────────────────────────── */
const SNIPPETS = [
  {
    id: 'hello',
    label: 'Hello World',
    category: 'Basics',
    code: `fn main() {
    let name = "Jabline";
    echo(\`Hello from \${name}! 🚀\`);
    
    let version = "0.6.0";
    echo(\`Version: \${version}\`);
    echo("Welcome to the playground!");
}`,
    output: `Hello from Jabline! 🚀
Version: 0.6.0
Welcome to the playground!`,
  },
  {
    id: 'variables',
    label: 'Variables & Types',
    category: 'Basics',
    code: `fn main() {
    // Type inference
    let name   = "Alice";
    let age    = 30;
    let pi     = 3.14159;
    let active = true;
    
    // Explicit types
    let count: int     = 100;
    let price: float64 = 19.99;
    
    // Constants
    const MAX = 1000;
    
    echo(name);
    echo(age);
    echo(pi);
    echo(active);
    echo(\`count = \${count}, price = \${price}\`);
    echo(\`MAX = \${MAX}\`);
}`,
    output: `Alice
30
3.14159
true
count = 100, price = 19.99
MAX = 1000`,
  },
  {
    id: 'functions',
    label: 'Functions',
    category: 'Basics',
    code: `fn add(a: int, b: int): int {
    return a + b;
}

fn greet(name: string): string {
    return \`Hello, \${name}!\`;
}

let double = (x) => x * 2;

fn makeCounter() {
    let count = 0;
    return fn() {
        count = count + 1;
        return count;
    };
}

fn main() {
    echo(add(3, 7));
    echo(greet("World"));
    echo(double(21));
    
    let counter = makeCounter();
    echo(counter());
    echo(counter());
    echo(counter());
}`,
    output: `10
Hello, World!
42
1
2
3`,
  },
  {
    id: 'structs',
    label: 'Structs & Methods',
    category: 'Types',
    code: `struct Point {
    x: float64,
    y: float64
}

fn (p Point) toString(): string {
    return \`(\${p.x}, \${p.y})\`;
}

fn (p Point) distanceTo(other: Point): float64 {
    let dx = p.x - other.x;
    let dy = p.y - other.y;
    return math.sqrt(dx*dx + dy*dy);
}

struct Rectangle {
    topLeft:     Point,
    bottomRight: Point
}

fn (r Rectangle) area(): float64 {
    let w = r.bottomRight.x - r.topLeft.x;
    let h = r.bottomRight.y - r.topLeft.y;
    return w * h;
}

fn main() {
    let a = Point { x: 0.0, y: 0.0 };
    let b = Point { x: 3.0, y: 4.0 };
    
    echo(a.toString());
    echo(b.toString());
    echo(\`Distance: \${a.distanceTo(b)}\`);
    
    let rect = Rectangle {
        topLeft:     Point { x: 0.0, y: 0.0 },
        bottomRight: Point { x: 5.0, y: 3.0 }
    };
    echo(\`Area: \${rect.area()}\`);
}`,
    output: `(0, 0)
(3, 4)
Distance: 5
Area: 15`,
  },
  {
    id: 'generics',
    label: 'Generics',
    category: 'Types',
    code: `fn identity<T>(value: T): T {
    return value;
}

fn map<T, U>(arr: Array<T>, f: fn(T): U): Array<U> {
    let result = Array<U>();
    for item in arr {
        result.push(f(item));
    }
    return result;
}

fn filter<T>(arr: Array<T>, pred: fn(T): bool): Array<T> {
    let result = Array<T>();
    for item in arr {
        if (pred(item)) { result.push(item); }
    }
    return result;
}

struct Box<T> {
    value: T,
    label: string
}

fn main() {
    echo(identity(42));
    echo(identity("hello"));
    
    let nums = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
    let doubled = map(nums, (x) => x * 2);
    let evens   = filter(nums, (x) => x % 2 == 0);
    
    echo(\`Doubled: \${doubled}\`);
    echo(\`Evens: \${evens}\`);
    
    let box = Box { value: 42, label: "answer" };
    echo(\`\${box.label}: \${box.value}\`);
}`,
    output: `42
hello
Doubled: [2, 4, 6, 8, 10, 12, 14, 16, 18, 20]
Evens: [2, 4, 6, 8, 10]
answer: 42`,
  },
  {
    id: 'concurrency',
    label: 'Concurrency',
    category: 'Advanced',
    code: `fn worker(id: int, jobs, results) {
    let job = recv(jobs);
    let result = job * job;
    echo(\`Worker \${id}: \${job}² = \${result}\`);
    send(results, result);
}

fn main() {
    let jobs    = make_chan();
    let results = make_chan();
    
    // Spawn 5 workers
    for (let i = 1; i <= 5; i++) {
        spawn worker(i, jobs, results);
    }
    
    // Send jobs
    let inputs = [3, 5, 7, 11, 13];
    for n in inputs {
        send(jobs, n);
    }
    
    // Collect and sum results
    let total = 0;
    for (let i = 0; i < 5; i++) {
        total = total + recv(results);
    }
    
    echo(\`Sum of squares: \${total}\`);
}`,
    output: `Worker 1: 3² = 9
Worker 2: 5² = 25
Worker 3: 7² = 49
Worker 4: 11² = 121
Worker 5: 13² = 169
Sum of squares: 373`,
  },
  {
    id: 'http-server',
    label: 'HTTP Server',
    category: 'Cloud',
    code: `import * as http from "std/net/http";
import * as json from "std/encoding/json";

struct User {
    id:    int,
    name:  string,
    email: string
}

let users = Array<User>();

fn main() {
    let server = http.createServer();
    server.use(http.cors());
    server.use(http.logger());
    
    server.get("/users", fn(req) {
        return http.json(200, users);
    });
    
    server.post("/users", fn(req) {
        let body = req.json();
        users.push(User {
            id:    len(users) + 1,
            name:  body.name,
            email: body.email
        });
        return http.json(201, { status: "created" });
    });
    
    server.delete("/users/:id", fn(req) {
        let id = int(req.params.id);
        users = users.filter(fn(u) { return u.id != id; });
        return http.json(200, { status: "deleted" });
    });
    
    server.listen(3000);
    echo("✓ API running on http://localhost:3000");
    echo("  GET  /users      → list users");
    echo("  POST /users      → create user");
    echo("  DEL  /users/:id  → delete user");
}`,
    output: `✓ API running on http://localhost:3000
  GET  /users      → list users
  POST /users      → create user
  DEL  /users/:id  → delete user`,
  },
  {
    id: 'json',
    label: 'JSON & Encoding',
    category: 'Cloud',
    code: `import { parse, stringify, prettify } from "std/encoding/json";

fn main() {
    // Parse
    let raw = '{"name":"Alice","age":30,"skills":["Go","Jabline"]}';
    let user = parse(raw);
    
    echo(\`Name:  \${user.name}\`);
    echo(\`Age:   \${user.age}\`);
    echo(\`Skills: \${user.skills}\`);
    
    // Stringify
    let config = {
        version:  "0.6.0",
        features: ["generics", "concurrency", "LSP"],
        cloud:    true
    };
    
    echo(stringify(config));
    
    // Pretty-print
    let pretty = prettify({ data: [1, 2, 3] }, 2);
    echo(pretty);
}`,
    output: `Name:  Alice
Age:   30
Skills: ["Go","Jabline"]
{"version":"0.6.0","features":["generics","concurrency","LSP"],"cloud":true}
{
  "data": [
    1,
    2,
    3
  ]
}`,
  },
]

/* ── Smarter code simulation ─────────────────────────────────── */
function simulateRun(code) {
  // 1. Try exact snippet match
  const match = SNIPPETS.find(s => s.code.trim() === code.trim())
  if (match) return { output: match.output, isError: false }

  // 2. Check for syntax errors (basic)
  const openBraces  = (code.match(/\{/g) || []).length
  const closeBraces = (code.match(/\}/g) || []).length
  const openParens  = (code.match(/\(/g) || []).length
  const closeParens = (code.match(/\)/g) || []).length
  if (openBraces !== closeBraces)  return { output: 'SyntaxError: Unmatched braces { }', isError: true }
  if (openParens !== closeParens)  return { output: 'SyntaxError: Unmatched parentheses ( )', isError: true }

  // 3. Extract echo() statements and simulate output
  const lines = []
  const echoRe = /echo\(([^)]*(?:\([^)]*\)[^)]*)*)\)/g
  let m
  while ((m = echoRe.exec(code)) !== null) {
    const arg = m[1].trim()

    // String literal
    if ((arg.startsWith('"') && arg.endsWith('"')) || (arg.startsWith("'") && arg.endsWith("'"))) {
      lines.push(arg.slice(1, -1))
      continue
    }

    // Template literal
    if (arg.startsWith('`') && arg.endsWith('`')) {
      let result = arg.slice(1, -1)
      // Resolve simple ${name} where name is a let/const variable in the code
      result = result.replace(/\$\{([^}]+)\}/g, (_, expr) => {
        const varMatch = code.match(new RegExp(`let\\s+${expr.trim()}\\s*=\\s*["'\`]?([^"'\`;\n]+)["'\`]?`))
        if (varMatch) return varMatch[1].trim().replace(/^["'`]|["'`]$/g, '')
        return `\${${expr}}`
      })
      lines.push(result)
      continue
    }

    // Number literal
    if (/^-?\d+(\.\d+)?$/.test(arg)) {
      lines.push(arg)
      continue
    }

    // Boolean literal
    if (arg === 'true' || arg === 'false') {
      lines.push(arg)
      continue
    }

    // Variable reference — try to resolve from let/const
    const varRe = new RegExp(`(?:let|const)\\s+${arg}\\s*(?::[^=]+)?=\\s*(.+?)(?:;|\\n|$)`)
    const varMatch = code.match(varRe)
    if (varMatch) {
      let val = varMatch[1].trim()
      val = val.replace(/^["'`]|["'`]$/g, '').replace(/\\n/g, '\n')
      lines.push(val)
      continue
    }

    // Fallback: show the expression
    lines.push(`[${arg}]`)
  }

  if (lines.length === 0) {
    return { output: '// No output\n// Use echo() to print values\n// Example:\n// echo("Hello, World!");', isError: false }
  }

  return { output: lines.join('\n'), isError: false }
}

/* ── Basic code formatter ────────────────────────────────────── */
function formatCode(code) {
  let indent = 0
  const lines = code.split('\n').map(line => line.trim()).filter((l, i, arr) => {
    // Remove consecutive blank lines
    return !(l === '' && arr[i - 1] === '')
  })
  return lines.map(line => {
    if (line.startsWith('}') || line.startsWith(')')) indent = Math.max(0, indent - 1)
    const result = '    '.repeat(indent) + line
    if (line.endsWith('{') || line.endsWith('(')) indent++
    return result
  }).join('\n')
}

export default function Playground({ setCurrentPage, initialCode }) {
  const [code, setCode]           = useState(initialCode || SNIPPETS[0].code)
  const [output, setOutput]       = useState('')
  const [isError, setIsError]     = useState(false)
  const [running, setRunning]     = useState(false)
  const [hasRun, setHasRun]       = useState(false)
  const [activeSnippet, setActiveSnippet] = useState(initialCode ? null : SNIPPETS[0].id)
  const [copyLabel, setCopyLabel] = useState('Share')
  const [lineCount, setLineCount] = useState(1)
  const textareaRef = useRef(null)
  const lineNumRef  = useRef(null)

  useEffect(() => {
    if (initialCode) {
      setCode(initialCode)
      setActiveSnippet(null)
      setOutput('')
      setHasRun(false)
    }
  }, [initialCode])

  useEffect(() => {
    setLineCount(code.split('\n').length)
  }, [code])

  // Sync line numbers scroll with textarea scroll
  const syncScroll = useCallback(() => {
    if (lineNumRef.current && textareaRef.current) {
      // Editor's inner textarea or pre might be scrolling
      const editorEl = textareaRef.current._input || textareaRef.current
      lineNumRef.current.scrollTop = editorEl.scrollTop
    }
  }, [])

  const handleRun = async () => {
    setRunning(true)
    setOutput('')
    setHasRun(false)

    try {
      const res = await fetch('/api/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code })
      })
      
      const data = await res.json()
      setOutput(data.output)
      setIsError(data.isError)
    } catch (err) {
      setOutput('Error connecting to backend API: ' + err.message)
      setIsError(true)
    } finally {
      setRunning(false)
      setHasRun(true)
    }
  }

  const handleLoadSnippet = (snippet) => {
    setCode(snippet.code)
    setOutput('')
    setHasRun(false)
    setIsError(false)
    setActiveSnippet(snippet.id)
    textareaRef.current?.focus()
  }

  const handleShare = () => {
    try {
      const encoded = btoa(unescape(encodeURIComponent(code)))
      const url = `${window.location.origin}${window.location.pathname}#play:${encoded}`
      navigator.clipboard.writeText(url)
    } catch {
      navigator.clipboard.writeText(code)
    }
    setCopyLabel('Copied!')
    setTimeout(() => setCopyLabel('Share'), 2000)
  }

  const handleFormat = () => {
    setCode(formatCode(code))
  }

  const handleTabKey = (e) => {
    if (e.key === 'Tab') {
      e.preventDefault()
      const el   = e.target
      const start = el.selectionStart
      const end   = el.selectionEnd
      const next  = code.substring(0, start) + '    ' + code.substring(end)
      setCode(next)
      requestAnimationFrame(() => {
        el.selectionStart = el.selectionEnd = start + 4
      })
    }
    // Ctrl+Enter or Cmd+Enter → Run
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault()
      handleRun()
    }
  }

  const categories = [...new Set(SNIPPETS.map(s => s.category))]

  return (
    <div className="min-h-screen pt-20 pb-10" style={{ background: 'var(--bg-base)' }}>
      {/* Header bar */}
      <div className="border-b border-jb-800" style={{ background: 'rgba(7,8,15,0.9)' }}>
        <div className="container-wide py-4">
          <div className="flex items-center justify-between gap-4 flex-wrap">
            <div className="flex items-center gap-3">
              <span className="text-accent text-xl">▶</span>
              <div>
                <div className="flex items-center gap-2">
                  <h1 className="text-jb-50 text-lg font-bold">Playground</h1>
                  <span className="accent-badge text-xs">Beta</span>
                </div>
                <p className="text-jb-500 text-xs font-mono hidden sm:block">
                  Ctrl+Enter to run · Tab to indent
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button onClick={handleFormat} className="jb-btn jb-btn-ghost text-xs py-1.5 px-3 border border-jb-700">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h8m-8 6h16"/>
                </svg>
                Format
              </button>
              <button onClick={handleShare} className="jb-btn jb-btn-ghost text-xs py-1.5 px-3 border border-jb-700">
                {copyLabel === 'Copied!' ? (
                  <><svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="#22d3a5" strokeWidth="2.5"><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7"/></svg> Copied!</>
                ) : (
                  <><svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2"><path strokeLinecap="round" strokeLinejoin="round" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"/></svg> Share</>
                )}
              </button>
              <button
                onClick={handleRun}
                disabled={running}
                className="jb-btn jb-btn-primary text-sm py-2 px-5"
              >
                {running ? (
                  <span className="flex items-center gap-2">
                    <span className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    Running...
                  </span>
                ) : (
                  '▶ Run'
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main layout */}
      <div className="container-wide mt-4">
        <div className="flex gap-4" style={{ height: 'calc(100vh - 160px)', minHeight: '520px' }}>

          {/* Snippet sidebar */}
          <aside className="w-48 flex-shrink-0 flex flex-col">
            <div className="jb-card overflow-hidden flex flex-col flex-1">
              <div className="px-3 py-2.5 border-b border-jb-800">
                <p className="text-jb-500 text-xs font-mono uppercase tracking-wider">Examples</p>
              </div>
              <div className="overflow-y-auto flex-1">
                {categories.map(cat => (
                  <div key={cat}>
                    <p className="text-jb-600 text-xs font-semibold uppercase tracking-widest px-3 py-2 mt-1">
                      {cat}
                    </p>
                    {SNIPPETS.filter(s => s.category === cat).map(snippet => (
                      <button
                        key={snippet.id}
                        onClick={() => handleLoadSnippet(snippet)}
                        className={`w-full text-left px-3 py-2 text-xs transition-all border-l-2 ${
                          activeSnippet === snippet.id
                            ? 'bg-accent/10 text-accent border-accent'
                            : 'text-jb-400 hover:text-jb-200 hover:bg-jb-800 border-transparent'
                        }`}
                      >
                        {snippet.label}
                      </button>
                    ))}
                  </div>
                ))}
              </div>
            </div>
          </aside>

          {/* Editor + Output column */}
          <div className="flex-1 flex flex-col gap-3 min-w-0 overflow-hidden">

            {/* Editor */}
            <div className="playground-editor flex-1 min-h-0" style={{ minHeight: '300px' }}>
              {/* Editor header */}
              <div className="code-header flex-shrink-0" style={{ background: '#04050d' }}>
                <div className="flex items-center gap-3">
                  <div className="code-dots">
                    <span className="dot dot-red" /><span className="dot dot-yellow" /><span className="dot dot-green" />
                  </div>
                  <span className="text-jb-500 text-xs font-mono">playground.jb</span>
                </div>
                <span className="text-jb-700 text-xs font-mono">{lineCount} lines</span>
              </div>

              {/* Editor body: line numbers + textarea */}
              <div className="flex flex-1 overflow-hidden" style={{ height: 'calc(100% - 41px)' }}>
                {/* Line numbers */}
                <div
                  ref={lineNumRef}
                  className="overflow-hidden flex-shrink-0 select-none"
                  style={{
                    background: '#04050d',
                    borderRight: '1px solid var(--border)',
                    padding: '1rem 0.75rem 1rem 0.5rem',
                    textAlign: 'right',
                    color: '#2c3150',
                    fontFamily: '"JetBrains Mono", monospace',
                    fontSize: '0.875rem',
                    lineHeight: '1.75',
                    minWidth: '3rem',
                  }}
                >
                  {Array.from({ length: lineCount }, (_, i) => (
                    <div key={i + 1}>{i + 1}</div>
                  ))}
                </div>

                {/* Editor component */}
                <div 
                  className="flex-1 overflow-auto playground-scroll"
                  onScroll={e => {
                    if (lineNumRef.current) lineNumRef.current.scrollTop = e.target.scrollTop
                  }}
                  style={{ background: 'transparent' }}
                >
                  <Editor
                    ref={textareaRef}
                    value={code}
                    onValueChange={code => setCode(code)}
                    highlight={code => hljs.highlight(code, { language: 'javascript' }).value}
                    padding={16}
                    style={{
                      fontFamily: '"JetBrains Mono", monospace',
                      fontSize: '0.875rem',
                      lineHeight: '1.75',
                      color: 'transparent', // The pre block handles color
                      minHeight: '100%',
                    }}
                    textareaClassName="focus:outline-none"
                    preClassName="hljs"
                  />
                </div>
              </div>
            </div>

            {/* Output panel */}
            <div className="playground-output flex-shrink-0" style={{ height: '180px' }}>
              {/* Output header */}
              <div className="code-header flex-shrink-0" style={{ background: '#02030a' }}>
                <div className="flex items-center gap-3">
                  <div className="code-dots">
                    <span className="dot dot-red" /><span className="dot dot-yellow" /><span className="dot dot-green" />
                  </div>
                  <span className="text-jb-500 text-xs font-mono">Output</span>
                </div>
                <div className="flex items-center gap-2">
                  {running && (
                    <span className="flex items-center gap-1.5 text-xs font-mono text-accent">
                      <span className="w-2 h-2 border border-accent/40 border-t-accent rounded-full animate-spin" />
                      Running...
                    </span>
                  )}
                  {hasRun && !running && !isError && (
                    <span className="flex items-center gap-1.5 text-xs font-mono text-success">
                      <span className="w-1.5 h-1.5 rounded-full bg-success" />
                      Done
                    </span>
                  )}
                  {hasRun && !running && isError && (
                    <span className="flex items-center gap-1.5 text-xs font-mono text-danger">
                      <span className="w-1.5 h-1.5 rounded-full bg-danger" />
                      Error
                    </span>
                  )}
                  {!hasRun && !running && (
                    <span className="text-jb-700 text-xs font-mono">Press ▶ Run or Ctrl+Enter</span>
                  )}
                </div>
              </div>

              {/* Output content */}
              <div className="overflow-y-auto flex-1">
                {output ? (
                  <pre className={isError ? 'error-output' : ''} style={{ padding: '0.875rem 1.25rem', margin: 0, fontFamily: '"JetBrains Mono", monospace', fontSize: '0.8125rem', lineHeight: '1.75' }}>
                    {output}
                  </pre>
                ) : (
                  <div className="flex items-center justify-center h-full">
                    <p className="text-jb-700 text-xs font-mono">
                      {running ? 'Executing...' : 'Output will appear here'}
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Note bar */}
        <div className="mt-3 px-4 py-3 rounded-lg flex items-center gap-3" style={{ background: 'rgba(61,126,255,0.05)', border: '1px solid rgba(61,126,255,0.15)' }}>
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="var(--accent-light)" strokeWidth="2" className="flex-shrink-0">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <p className="text-jb-400 text-xs">
            <span className="text-accent font-semibold">Note:</span>{' '}
            Code is now executing natively on your local Jabline VM!
          </p>
        </div>
      </div>
    </div>
  )
}
