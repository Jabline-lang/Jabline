import { useState, useEffect } from 'react'
import FeatureCard from '../components/common/FeatureCard'
import CodeBlock from '../components/common/CodeBlock'

const HERO_SNIPPETS = [
  {
    label: 'Hello World',
    code: `fn main() {
    let name = "World";
    echo(\`Hello, \${name}!\`);
}`,
  },
  {
    label: 'Concurrency',
    code: `fn main() {
    let ch = make_chan();
    
    spawn fn() {
        send(ch, compute());
    }();
    
    let result = recv(ch);
    echo(result);
}`,
  },
  {
    label: 'HTTP Server',
    code: `import * as http from "std/net/http";

fn main() {
    let server = http.createServer();
    
    server.get("/", fn(req) {
        return http.response(200, "🚀 Jabline!");
    });
    
    server.listen(3000);
}`,
  },
]

const FEATURES = [
  {
    icon: '⚡',
    title: 'Bytecode VM',
    description: 'Compiles to JBVM bytecode — faster than tree-walking interpreters, simpler than LLVM.',
  },
  {
    icon: '🔀',
    title: 'Native Concurrency',
    description: 'CSP model with spawn & channels. Goroutine-style parallelism built into the language.',
  },
  {
    icon: '🔒',
    title: 'Static Type System',
    description: 'AOT type checker catches errors before runtime. Generics, structs, and type inference.',
  },
  {
    icon: '📦',
    title: 'Complete Toolchain',
    description: 'LSP, formatter, test runner, debugger, REPL, and package manager — all built in.',
  },
  {
    icon: '☁️',
    title: 'Cloud-Native',
    description: 'std/net/http, std/db, std/crypto, std/encoding/json — production modules included.',
  },
  {
    icon: '🎯',
    title: 'Zero Config',
    description: 'One binary. No runtime dependencies. Build and ship anywhere.',
  },
]

const STATS = [
  { label: 'Version', value: 'v0.6.0' },
  { label: 'License', value: 'MIT' },
  { label: 'Written in', value: 'Go' },
  { label: 'Platforms', value: 'Win · Mac · Linux' },
]

export default function Home({ setCurrentPage }) {
  const [snippetIdx, setSnippetIdx] = useState(0)
  const [copied, setCopied] = useState(false)
  const installCmd = 'go install github.com/Jabline-lang/Jabline@latest'

  // Rotate hero snippets
  useEffect(() => {
    const id = setInterval(() => {
      setSnippetIdx(i => (i + 1) % HERO_SNIPPETS.length)
    }, 4000)
    return () => clearInterval(id)
  }, [])

  const handleCopyInstall = async () => {
    await navigator.clipboard.writeText(installCmd).catch(() => {})
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <>
      {/* ── HERO ─────────────────────────────────────────────── */}
      <section className="relative min-h-screen flex flex-col justify-center overflow-hidden">
        {/* Radial glow background */}
        <div className="absolute inset-0 hero-glow pointer-events-none" />
        {/* Grid pattern overlay */}
        <div
          className="absolute inset-0 opacity-[0.03] pointer-events-none"
          style={{
            backgroundImage: 'linear-gradient(#f7a41d 1px, transparent 1px), linear-gradient(90deg, #f7a41d 1px, transparent 1px)',
            backgroundSize: '48px 48px',
          }}
        />

        <div className="container-wide relative z-10 pt-32 pb-20">
          <div className="grid lg:grid-cols-2 gap-16 items-center">
            {/* Left: Text */}
            <div>
              <div className="animate-fade-in">
                <span className="accent-badge mb-6 inline-block">
                  v0.6.0 — Production Ready
                </span>
              </div>

              <h1
                className="animate-slide-up delay-100 font-black tracking-tight mb-6"
                style={{ lineHeight: '1.05' }}
              >
                <span className="text-jb-50">Fast.</span>{' '}
                <span className="text-jb-50">Expressive.</span>
                <br />
                <span className="gradient-text">Cloud-Native.</span>
              </h1>

              <p className="animate-slide-up delay-200 text-jb-300 text-xl leading-relaxed mb-8 max-w-lg">
                Jabline is a compiled, cloud-native language with a custom bytecode VM. 
                Built for speed, concurrency, and a complete developer experience.
              </p>

              {/* Install command */}
              <div className="animate-slide-up delay-300 mb-8">
                <p className="text-jb-500 text-xs font-mono uppercase tracking-widest mb-2">Install</p>
                <div className="flex items-center gap-2 bg-jb-850 border border-jb-700 rounded-lg px-4 py-3 max-w-md group">
                  <span className="text-accent font-mono text-sm select-none">$</span>
                  <code className="text-jb-200 font-mono text-sm flex-1 truncate">{installCmd}</code>
                  <button
                    onClick={handleCopyInstall}
                    className="text-jb-500 hover:text-accent transition-colors flex-shrink-0"
                    aria-label="Copy install command"
                  >
                    {copied ? (
                      <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" fill="none" viewBox="0 0 24 24" stroke="#22d3a5" strokeWidth="2.5">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7"/>
                      </svg>
                    ) : (
                      <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                        <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
                      </svg>
                    )}
                  </button>
                </div>
              </div>

              {/* CTA Buttons */}
              <div className="animate-slide-up delay-400 flex flex-wrap gap-3">
                <button
                  onClick={() => setCurrentPage('guide')}
                  className="jb-btn jb-btn-primary text-base px-6 py-3"
                >
                  Get Started
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2.5">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7"/>
                  </svg>
                </button>
                <button
                  onClick={() => setCurrentPage('playground')}
                  className="jb-btn jb-btn-secondary text-base px-6 py-3"
                >
                  ▶ Try Playground
                </button>
                <a
                  href="https://github.com/Jabline-lang/Jabline"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="jb-btn jb-btn-ghost text-base px-6 py-3"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
                  </svg>
                  GitHub
                </a>
              </div>
            </div>

            {/* Right: Animated code snippets */}
            <div className="animate-fade-in delay-300 hidden lg:block">
              <div className="relative">
                {/* Glow effect behind code block */}
                <div className="absolute -inset-4 bg-accent/5 rounded-2xl blur-xl" />
                <div className="relative">
                  {/* Snippet selector tabs */}
                  <div className="flex gap-1 mb-2">
                    {HERO_SNIPPETS.map((s, i) => (
                      <button
                        key={i}
                        onClick={() => setSnippetIdx(i)}
                        className={`px-3 py-1.5 rounded-md text-xs font-mono transition-all ${
                          i === snippetIdx
                            ? 'bg-accent/15 text-accent border border-accent/30'
                            : 'text-jb-400 hover:text-jb-200'
                        }`}
                      >
                        {s.label}
                      </button>
                    ))}
                  </div>
                  <CodeBlock
                    key={snippetIdx}
                    code={HERO_SNIPPETS[snippetIdx].code}
                    language="jabline"
                    filename={`example.jb`}
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Stats bar */}
          <div className="mt-16 pt-8 border-t border-jb-800">
            <div className="flex flex-wrap gap-8 justify-center sm:justify-start">
              {STATS.map(stat => (
                <div key={stat.label} className="text-center sm:text-left">
                  <div className="text-jb-50 font-mono font-bold text-lg">{stat.value}</div>
                  <div className="text-jb-500 text-xs uppercase tracking-widest mt-0.5">{stat.label}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* ── FEATURES ─────────────────────────────────────────── */}
      <section className="py-24 relative">
        <div className="container-max">
          <div className="text-center mb-16">
            <span className="accent-badge mb-4 inline-block">Features</span>
            <h2 className="text-jb-50 font-bold mb-4">Everything you need</h2>
            <p className="text-jb-400 text-lg max-w-xl mx-auto">
              Jabline ships with a complete toolchain out of the box — no ecosystem fragmentation.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {FEATURES.map((feature, idx) => (
              <div key={idx} className="animate-fade-in" style={{ animationDelay: `${idx * 80}ms` }}>
                <FeatureCard
                  icon={feature.icon}
                  title={feature.title}
                  description={feature.description}
                  accent
                />
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── CODE PREVIEW SECTION ─────────────────────────────── */}
      <section className="py-24 bg-jb-950/50">
        <div className="container-max">
          <div className="grid lg:grid-cols-2 gap-16 items-center">
            <div>
              <span className="accent-badge mb-4 inline-block">Syntax</span>
              <h2 className="text-jb-50 font-bold mb-6">Familiar yet powerful</h2>
              <p className="text-jb-400 leading-relaxed mb-6">
                Jabline takes the best from modern languages: type inference from Go, 
                expressive syntax from JavaScript, and CSP concurrency from Erlang.
              </p>
              <ul className="space-y-3">
                {[
                  'Static typing with full type inference',
                  'Generics with type-safe data structures',
                  'First-class functions and closures',
                  'Template literals and pipe operator',
                  'Optional chaining and nullish coalescing',
                ].map(item => (
                  <li key={item} className="flex items-center gap-3 text-jb-300 text-sm">
                    <span className="w-5 h-5 rounded-full bg-accent/15 border border-accent/30 flex items-center justify-center flex-shrink-0">
                      <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" fill="none" viewBox="0 0 24 24" stroke="#f7a41d" strokeWidth="3">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7"/>
                      </svg>
                    </span>
                    {item}
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <CodeBlock
                code={`struct User {
    id:     int,
    name:   string,
    email:  string,
    active: bool
}

fn (u User) greet(): string {
    return \`Hello, \${u.name}!\`;
}

fn main() {
    let users = Array<User>();
    
    users.push(User {
        id: 1, name: "Alice",
        email: "alice@example.com",
        active: true
    });
    
    // Filter active users
    for user in users {
        if (user.active) {
            echo(user.greet());
        }
    }
}`}
                language="jabline"
                filename="users.jb"
              />
            </div>
          </div>
        </div>
      </section>

      {/* ── CONCURRENCY SHOWCASE ─────────────────────────────── */}
      <section className="py-24">
        <div className="container-max">
          <div className="grid lg:grid-cols-2 gap-16 items-center">
            <div className="order-2 lg:order-1">
              <CodeBlock
                code={`import * as http from "std/net/http";
import * as json from "std/encoding/json";

fn main() {
    let server = http.createServer();
    let ch = make_chan();
    
    // Background worker pool
    for (let i = 0; i < 4; i++) {
        spawn fn() {
            while (true) {
                let task = recv(ch);
                processTask(task);
            }
        }();
    }
    
    server.post("/tasks", fn(req) {
        let body = json.parse(req.body);
        send(ch, body);
        return http.response(202, "Accepted");
    });
    
    server.listen(8080);
    echo("Workers ready ✓");
}`}
                language="jabline"
                filename="server.jb"
              />
            </div>
            <div className="order-1 lg:order-2">
              <span className="accent-badge mb-4 inline-block">Concurrency</span>
              <h2 className="text-jb-50 font-bold mb-6">Parallelism without the pain</h2>
              <p className="text-jb-400 leading-relaxed mb-6">
                Jabline's CSP model makes concurrent code readable and safe. 
                Spawn goroutines, communicate via channels, and build scalable backends effortlessly.
              </p>
              <div className="grid grid-cols-2 gap-4">
                {[
                  { label: 'spawn', desc: 'Lightweight goroutines' },
                  { label: 'make_chan()', desc: 'Buffered channels' },
                  { label: 'send/recv', desc: 'Safe message passing' },
                  { label: 'async/await', desc: 'Promise-based async' },
                ].map(item => (
                  <div key={item.label} className="jb-card p-4">
                    <code className="text-accent text-sm font-mono block mb-1">{item.label}</code>
                    <span className="text-jb-400 text-xs">{item.desc}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── TOOLCHAIN ────────────────────────────────────────── */}
      <section className="py-24 bg-jb-950/50">
        <div className="container-max">
          <div className="text-center mb-16">
            <span className="accent-badge mb-4 inline-block">Toolchain</span>
            <h2 className="text-jb-50 font-bold mb-4">One binary, everything included</h2>
          </div>
          <div className="jb-code-block">
            <div className="code-header">
              <div className="code-dots">
                <span className="dot dot-red" /><span className="dot dot-yellow" /><span className="dot dot-green" />
              </div>
              <span className="text-jb-500 text-xs font-mono">terminal</span>
            </div>
            <div className="p-6 font-mono text-sm space-y-2">
              {[
                { cmd: 'jabline run hello.jb',   desc: '# Run a program' },
                { cmd: 'jabline build hello.jb',  desc: '# Compile to binary' },
                { cmd: 'jabline test',             desc: '# Run *_test.jb files' },
                { cmd: 'jabline fmt ./...',         desc: '# Format all code' },
                { cmd: 'jabline debug hello.jb',   desc: '# Interactive debugger' },
                { cmd: 'jabline lsp',               desc: '# Start LSP server' },
                { cmd: 'jabline repl',              desc: '# Interactive REPL' },
                { cmd: 'jabline get pkg@1.0.0',    desc: '# Install a package' },
              ].map((row, i) => (
                <div key={i} className="flex items-center gap-4">
                  <span className="text-accent">$</span>
                  <span className="text-jb-100">{row.cmd}</span>
                  <span className="text-jb-600 ml-auto hidden sm:inline">{row.desc}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* ── CTA ──────────────────────────────────────────────── */}
      <section className="py-32 relative overflow-hidden">
        <div className="absolute inset-0 hero-glow pointer-events-none opacity-50" />
        <div className="container-narrow text-center relative z-10">
          <h2 className="text-jb-50 font-black mb-6" style={{ fontSize: 'clamp(2rem, 5vw, 3.5rem)' }}>
            Ready to build?
          </h2>
          <p className="text-jb-400 text-xl mb-10 max-w-lg mx-auto">
            Join the Jabline community and start shipping cloud-native applications today.
          </p>
          <div className="flex flex-wrap gap-4 justify-center">
            <button
              onClick={() => setCurrentPage('guide')}
              className="jb-btn jb-btn-primary text-lg px-8 py-4 animate-glow"
            >
              Start Learning
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2.5">
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7"/>
              </svg>
            </button>
            <button
              onClick={() => setCurrentPage('playground')}
              className="jb-btn jb-btn-secondary text-lg px-8 py-4"
            >
              ▶ Open Playground
            </button>
          </div>
        </div>
      </section>
    </>
  )
}
