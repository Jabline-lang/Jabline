import CodeBlock from '../components/common/CodeBlock'
import DocsSidebar from '../components/common/DocsSidebar'

const MODULES = [
  {
    id: 'http',
    name: 'std/net/http',
    desc: 'HTTP server, router, middleware, and client',
    color: 'text-blue-400',
    code: `import * as http from "std/net/http";

let server = http.createServer();

server.use(http.cors());           // CORS middleware
server.use(http.logger());         // Request logger

server.get("/", fn(req) {
    return http.response(200, "Hello!");
});

server.get("/users/:id", fn(req) {
    let id = req.params.id;
    return http.json(200, { id: id, name: "Alice" });
});

server.post("/users", fn(req) {
    let body = req.json();
    return http.json(201, { status: "created", data: body });
});

server.listen(8080);
echo("Listening on :8080");`,
    fns: ['createServer()', 'response(status, body)', 'json(status, obj)', 'cors()', 'logger()', 'get(path, handler)', 'post/put/delete(path, handler)', 'listen(port)'],
  },
  {
    id: 'json',
    name: 'std/encoding/json',
    desc: 'JSON serialization and deserialization',
    color: 'text-yellow-400',
    code: `import { parse, stringify, prettify } from "std/encoding/json";

// Parse JSON string to object
let user = parse('{"name":"Alice","age":30}');
echo(user.name);  // Alice

// Stringify object to JSON
let json = stringify({ version: "0.6.0", active: true });
echo(json);  // {"version":"0.6.0","active":true}

// Pretty-print with indent
let pretty = prettify({ data: [1, 2, 3] }, 2);
echo(pretty);
// {
//   "data": [1, 2, 3]
// }`,
    fns: ['parse(str)', 'stringify(obj)', 'prettify(obj, indent)', 'isValid(str)', 'merge(a, b)'],
  },
  {
    id: 'crypto',
    name: 'std/crypto',
    desc: 'Hashing, encryption, UUIDs, random bytes',
    color: 'text-purple-400',
    code: `import * as crypto from "std/crypto";

let hash    = crypto.sha256("hello world");
let md5     = crypto.md5("data");
let uuid    = crypto.uuid();
let rand    = crypto.randomBytes(32);
let randInt = crypto.randomInt(1, 100);

// HMAC signing
let sig = crypto.hmacSha256("secret-key", "message");

// AES encryption
let encrypted = crypto.aesEncrypt("key-32-bytes-long", "plaintext");
let decrypted = crypto.aesDecrypt("key-32-bytes-long", encrypted);

// bcrypt (password hashing)
let hashed   = crypto.bcrypt("password", 12);
let verified = crypto.bcryptVerify("password", hashed);`,
    fns: ['sha256(data)', 'md5(data)', 'uuid()', 'randomBytes(n)', 'randomInt(min, max)', 'hmacSha256(key, msg)', 'aesEncrypt/aesDecrypt(key, data)', 'bcrypt(pass, cost)', 'bcryptVerify(pass, hash)'],
  },
  {
    id: 'db',
    name: 'std/db',
    desc: 'SQLite database operations',
    color: 'text-green-400',
    code: `import * as db from "std/db";

let conn = db.open("app.db");

// Create table
conn.exec(\`
    CREATE TABLE IF NOT EXISTS users (
        id    INTEGER PRIMARY KEY AUTOINCREMENT,
        name  TEXT NOT NULL,
        email TEXT UNIQUE
    )
\`);

// Insert
conn.exec(\`INSERT INTO users (name, email) VALUES (?, ?)\`,
    ["Alice", "alice@example.com"]);

// Query
let rows = conn.query("SELECT * FROM users WHERE id = ?", [1]);
for row in rows {
    echo(\`\${row.id}: \${row.name} — \${row.email}\`);
}

// Transactions
conn.transaction(fn() {
    conn.exec("INSERT INTO users (name) VALUES ('Bob')");
    conn.exec("INSERT INTO users (name) VALUES ('Charlie')");
});

conn.close();`,
    fns: ['open(path)', 'exec(sql, params?)', 'query(sql, params?)', 'queryOne(sql, params?)', 'transaction(fn)', 'close()'],
  },
  {
    id: 'io',
    name: 'std/sys/io',
    desc: 'File system operations',
    color: 'text-orange-400',
    code: `import * as io from "std/sys/io";

// Read / write files
let content = io.readFile("data.txt");
io.writeFile("output.txt", content);
io.appendFile("log.txt", "new line\n");

// File metadata
let exists = io.fileExists("config.json");
let size   = io.fileSize("data.bin");
let info   = io.stat("main.jb");

// Directory operations
io.mkdir("dist", { recursive: true });
let files = io.readDir("./src");
io.remove("temp.txt");
io.rename("old.txt", "new.txt");

// Copy / move
io.copy("src/main.jb", "dist/main.jb");

// Path utilities
let abs  = io.absPath("./main.jb");
let base = io.basename("/home/user/file.txt");  // file.txt
let ext  = io.extension("main.jb");             // .jb`,
    fns: ['readFile(path)', 'writeFile(path, data)', 'appendFile(path, data)', 'fileExists(path)', 'mkdir(path)', 'readDir(path)', 'remove(path)', 'copy(src, dst)', 'basename(path)', 'extension(path)'],
  },
  {
    id: 'time',
    name: 'std/time/datetime',
    desc: 'Date, time, and duration utilities',
    color: 'text-cyan-400',
    code: `import * as time from "std/time/datetime";

let now       = time.now();
let formatted = time.format(now, "YYYY-MM-DD HH:mm:ss");
let unix      = time.unix(now);

// Parsing
let date = time.parse("2026-01-15", "YYYY-MM-DD");

// Arithmetic
let tomorrow  = time.addDays(now, 1);
let nextWeek  = time.addDays(now, 7);
let nextMonth = time.addMonths(now, 1);
let nextYear  = time.addYears(now, 1);

// Duration
let start = time.now();
doWork();
let elapsed = time.since(start);
echo(\`Took \${time.milliseconds(elapsed)}ms\`);

// Sleep
sleep(1000);  // 1 second`,
    fns: ['now()', 'format(t, fmt)', 'parse(str, fmt)', 'unix(t)', 'addDays/addMonths/addYears(t, n)', 'since(t)', 'milliseconds(d)', 'sleep(ms)'],
  },
  {
    id: 'strings',
    name: 'std/strings',
    desc: 'String manipulation and pattern matching',
    color: 'text-pink-400',
    code: `import * as str from "std/strings";

let s = "Hello, World!";

echo(str.toUpper(s));           // HELLO, WORLD!
echo(str.toLower(s));           // hello, world!
echo(str.trim("  hi  "));       // "hi"
echo(str.trimLeft("  hi"));     // "hi"
echo(str.replace(s, "World", "Jabline"));

let parts = str.split("a,b,c", ",");  // ["a","b","c"]
let joined = str.join(parts, " | ");  // "a | b | c"

echo(str.contains(s, "World"));   // true
echo(str.startsWith(s, "Hello")); // true
echo(str.endsWith(s, "!"));       // true
echo(str.indexOf(s, "World"));    // 7
echo(str.length(s));              // 13
echo(str.substring(s, 7, 12));    // World

// Regex
let matches = str.match(s, /\w+/g);
let cleaned = str.replaceRegex(s, /[^a-z]/gi, "");`,
    fns: ['toUpper/toLower(s)', 'trim/trimLeft/trimRight(s)', 'split(s, sep)', 'join(arr, sep)', 'replace(s, from, to)', 'contains/startsWith/endsWith(s, sub)', 'indexOf(s, sub)', 'length(s)', 'substring(s, start, end)', 'match(s, regex)', 'replaceRegex(s, re, repl)'],
  },
  {
    id: 'math',
    name: 'std/math',
    desc: 'Mathematical functions and constants',
    color: 'text-red-400',
    code: `import * as math from "std/math";

// Constants
echo(math.PI);   // 3.141592653589793
echo(math.E);    // 2.718281828459045
echo(math.PHI);  // 1.618033988749895

// Basic functions
echo(math.abs(-5));        // 5
echo(math.sqrt(16));       // 4
echo(math.pow(2, 10));     // 1024
echo(math.log(math.E));    // 1
echo(math.log2(1024));     // 10
echo(math.log10(1000));    // 3

// Rounding
echo(math.floor(3.7));  // 3
echo(math.ceil(3.2));   // 4
echo(math.round(3.5));  // 4

// Trig
echo(math.sin(math.PI / 2));  // 1
echo(math.cos(0));             // 1

// Min / Max
echo(math.min(3, 7, 1, 9));  // 1
echo(math.max(3, 7, 1, 9));  // 9`,
    fns: ['PI · E · PHI', 'abs(n)', 'sqrt(n)', 'pow(base, exp)', 'log/log2/log10(n)', 'floor/ceil/round(n)', 'sin/cos/tan(n)', 'min/max(...nums)'],
  },
]

export default function Stdlib({ setCurrentPage }) {
  return (
    <div className="pt-20 min-h-screen">
      <div className="border-b border-jb-800 bg-jb-950/40">
        <div className="container-wide py-12">
          <span className="accent-badge mb-4 inline-block">Standard Library</span>
          <h1 className="text-jb-50 font-black mb-3">Standard Library</h1>
          <p className="text-jb-400 text-xl max-w-xl">
            Production-ready modules included with every Jabline installation. No external packages needed for common tasks.
          </p>
        </div>
      </div>

      <div className="container-wide py-12">
        <div className="flex gap-12">
          <DocsSidebar sections={MODULES.map(m => ({ id: m.id, label: m.name }))} />

          <main className="flex-1 min-w-0 space-y-16">
            {MODULES.map(mod => (
              <section key={mod.id} id={mod.id} className="scroll-mt-24">
                <div className="flex items-center gap-3 mb-2">
                  <h2 className={`font-bold text-2xl ${mod.color}`}>{mod.name}</h2>
                </div>
                <p className="text-jb-400 mb-2">{mod.desc}</p>
                <div className="h-px bg-jb-800 mb-8" />

                <CodeBlock code={mod.code} language="jabline" />

                {/* Function quick-ref */}
                <div className="mt-4 p-4 jb-card">
                  <p className="text-jb-500 text-xs uppercase tracking-widest mb-3 font-semibold">Functions</p>
                  <div className="flex flex-wrap gap-2">
                    {mod.fns.map(fn => (
                      <code key={fn} className="text-xs bg-jb-850 border border-jb-700 text-accent px-2 py-1 rounded">
                        {fn}
                      </code>
                    ))}
                  </div>
                </div>
              </section>
            ))}
          </main>
        </div>
      </div>
    </div>
  )
}
