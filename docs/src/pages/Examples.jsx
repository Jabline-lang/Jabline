import { useState } from 'react'
import CodeBlock from '../components/common/CodeBlock'

const EXAMPLES = [
  {
    id: 'web-server',
    title: 'REST API Server',
    category: 'Cloud',
    desc: 'Full HTTP REST server with routing, JSON responses, and CORS.',
    code: `import * as http from "std/net/http";
import * as json from "std/encoding/json";

struct User {
    id:    int,
    name:  string,
    email: string
}

let store = Array<User>();
let nextId = 1;

fn main() {
    let server = http.createServer();
    server.use(http.cors());
    server.use(http.logger());

    // GET /users
    server.get("/users", fn(req) {
        return http.json(200, store);
    });

    // GET /users/:id
    server.get("/users/:id", fn(req) {
        let id = int(req.params.id);
        for user in store {
            if (user.id == id) {
                return http.json(200, user);
            }
        }
        return http.json(404, { error: "Not found" });
    });

    // POST /users
    server.post("/users", fn(req) {
        let body = req.json();
        store.push(User {
            id:    nextId,
            name:  body.name,
            email: body.email
        });
        nextId++;
        return http.json(201, { status: "created" });
    });

    // DELETE /users/:id
    server.delete("/users/:id", fn(req) {
        let id = int(req.params.id);
        store = store.filter(fn(u) { return u.id != id; });
        return http.json(200, { status: "deleted" });
    });

    server.listen(3000);
    echo("API running on :3000");
}`,
  },
  {
    id: 'concurrency',
    title: 'Parallel Processing',
    category: 'Concurrency',
    desc: 'Spawn multiple workers and collect results via channels.',
    code: `import * as time from "std/time/datetime";

fn processItem(id: int, input: string, results) {
    // Simulate work
    sleep(100);
    let output = \`[\${id}] Processed: \${input}\`;
    send(results, output);
}

fn main() {
    let items   = ["alpha", "beta", "gamma", "delta", "epsilon"];
    let results = make_chan();
    let start   = time.now();

    // Spawn one worker per item
    for (let i = 0; i < len(items); i++) {
        let item = items[i];
        let idx  = i + 1;
        spawn processItem(idx, item, results);
    }

    // Collect all results
    let outputs = Array<string>();
    for (let i = 0; i < len(items); i++) {
        outputs.push(recv(results));
    }

    let elapsed = time.since(start);
    echo(\`Processed \${len(items)} items in \${time.milliseconds(elapsed)}ms\`);

    for out in outputs {
        echo(out);
    }
}`,
  },
  {
    id: 'data-processing',
    title: 'Data Pipeline',
    category: 'Patterns',
    desc: 'Filter, map, and aggregate structured data with generics.',
    code: `import * as json from "std/encoding/json";

struct Order {
    id:       int,
    customer: string,
    amount:   float64,
    status:   string
}

fn totalRevenue(orders: Array<Order>): float64 {
    let total = 0.0;
    for o in orders {
        if (o.status == "completed") {
            total = total + o.amount;
        }
    }
    return total;
}

fn groupByStatus(orders: Array<Order>): Hash {
    let groups = {};
    for o in orders {
        if (!groups[o.status]) {
            groups[o.status] = Array<Order>();
        }
        groups[o.status].push(o);
    }
    return groups;
}

fn main() {
    let orders = [
        Order { id: 1, customer: "Alice", amount: 120.00, status: "completed" },
        Order { id: 2, customer: "Bob",   amount: 85.50,  status: "pending"   },
        Order { id: 3, customer: "Carol", amount: 200.00, status: "completed" },
        Order { id: 4, customer: "Dave",  amount: 45.00,  status: "cancelled" },
    ];

    echo(\`Total revenue: $\${totalRevenue(orders)}\`);

    let groups = groupByStatus(orders);
    for (status, items) in groups {
        echo(\`\${status}: \${len(items)} orders\`);
    }
}`,
  },
  {
    id: 'cli-tool',
    title: 'CLI Tool',
    category: 'Cloud',
    desc: 'Build a command-line tool with argument parsing and file operations.',
    code: `import * as io   from "std/sys/io";
import * as json from "std/encoding/json";

struct Config {
    output: string,
    format: string,
    verbose: bool
}

fn parseArgs(args: Array<string>): Config {
    let cfg = Config {
        output:  "output.json",
        format:  "json",
        verbose: false
    };

    for (let i = 0; i < len(args); i++) {
        switch (args[i]) {
            case "--output": cfg.output  = args[i + 1];
            case "--format": cfg.format  = args[i + 1];
            case "--verbose": cfg.verbose = true;
        }
    }

    return cfg;
}

fn processFiles(dir: string, cfg: Config) {
    let files = io.readDir(dir);
    let results = Array<Hash>();

    for file in files {
        if (cfg.verbose) { echo(\`Processing: \${file}\`); }
        let content = io.readFile(\`\${dir}/\${file}\`);
        results.push({ file: file, size: len(content) });
    }

    io.writeFile(cfg.output, json.prettify(results, 2));
    echo(\`Done! Results saved to \${cfg.output}\`);
}

fn main() {
    let cfg = parseArgs(argv);
    processFiles("./data", cfg);
}`,
  },
  {
    id: 'http-client',
    title: 'API Client',
    category: 'Cloud',
    desc: 'Async HTTP client with error handling and retry logic.',
    code: `import * as http from "std/net/http";
import * as json from "std/encoding/json";

struct ApiClient {
    baseUrl: string,
    token:   string,
    retries: int
}

async fn (c ApiClient) get(path: string): Hash {
    let url = c.baseUrl + path;

    for (let attempt = 0; attempt <= c.retries; attempt++) {
        try {
            let resp = await http.get(url, {
                headers: { "Authorization": "Bearer " + c.token }
            });

            if (resp.status == 200) {
                return json.parse(resp.body);
            }

            if (resp.status == 429) {  // Rate limited
                sleep(1000 * attempt);
                continue;
            }

            throw \`HTTP \${resp.status}: \${resp.body}\`;
        } catch (err) {
            if (attempt == c.retries) { throw err; }
            echo(\`Attempt \${attempt + 1} failed, retrying...\`);
        }
    }
}

async fn main() {
    let client = ApiClient {
        baseUrl: "https://api.github.com",
        token:   "your-token-here",
        retries: 3
    };

    let user = await client.get("/users/Jabline-lang");
    echo(\`Name: \${user.name}\`);
    echo(\`Repos: \${user.public_repos}\`);
}`,
  },
]

const CATEGORIES = ['All', ...new Set(EXAMPLES.map(e => e.category))]

export default function Examples({ setCurrentPage, onTryCode }) {
  const [activeCategory, setActiveCategory] = useState('All')

  const filtered = activeCategory === 'All'
    ? EXAMPLES
    : EXAMPLES.filter(e => e.category === activeCategory)

  return (
    <div className="pt-20 min-h-screen">
      <div className="border-b border-jb-800 bg-jb-950/40">
        <div className="container-wide py-12">
          <span className="accent-badge mb-4 inline-block">Examples</span>
          <h1 className="text-jb-50 font-black mb-3">Practical Examples</h1>
          <p className="text-jb-400 text-xl max-w-xl">
            Real-world Jabline code demonstrating common patterns and use cases.
          </p>
        </div>
      </div>

      <div className="container-max py-12">
        {/* Category filter */}
        <div className="flex gap-2 flex-wrap mb-10">
          {CATEGORIES.map(cat => (
            <button
              key={cat}
              onClick={() => setActiveCategory(cat)}
              className={`px-4 py-1.5 rounded-full text-sm font-medium transition-all border ${
                activeCategory === cat
                  ? 'bg-accent/15 text-accent border-accent/40'
                  : 'text-jb-400 border-jb-700 hover:border-jb-500 hover:text-jb-200'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>

        {/* Examples */}
        <div className="space-y-12">
          {filtered.map(example => (
            <div key={example.id} id={example.id} className="scroll-mt-24">
              <div className="flex items-start justify-between gap-4 mb-4">
                <div>
                  <div className="flex items-center gap-3 mb-1">
                    <h2 className="text-jb-50 font-bold text-xl">{example.title}</h2>
                    <span className="text-xs px-2 py-0.5 rounded border border-jb-700 text-jb-400 font-mono">
                      {example.category}
                    </span>
                  </div>
                  <p className="text-jb-400">{example.desc}</p>
                </div>
                <button
                  onClick={() => onTryCode(example.code)}
                  className="jb-btn jb-btn-ghost text-sm py-1.5 px-3 flex-shrink-0 border border-jb-700 hover:border-accent/40 hover:text-accent"
                >
                  ▶ Try in Playground
                </button>
              </div>
              <CodeBlock code={example.code} language="jabline" />
            </div>
          ))}
        </div>

        {/* CTA */}
        <div className="mt-16 p-8 jb-card text-center border-accent/20">
          <h3 className="text-jb-100 font-bold text-xl mb-3">Want to run these examples?</h3>
          <p className="text-jb-400 mb-6">
            Try them directly in the browser playground or install Jabline locally.
          </p>
          <div className="flex gap-3 justify-center flex-wrap">
            <button onClick={() => setCurrentPage('playground')} className="jb-btn jb-btn-primary">
              ▶ Open Playground
            </button>
            <button onClick={() => setCurrentPage('guide')} className="jb-btn jb-btn-secondary">
              Installation Guide
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
