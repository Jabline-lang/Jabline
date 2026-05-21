import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { exec } from 'child_process'
import fs from 'fs'
import path from 'path'
import { promisify } from 'util'

const execAsync = promisify(exec)

// Custom Vite plugin to handle backend tasks for the docs
function jablineDocsApi() {
  return {
    name: 'jabline-docs-api',
    configureServer(server) {
      server.middlewares.use(async (req, res, next) => {
        
        // ── 1. JBVM Execution API ──────────────────────────────────────────
        if (req.url === '/api/run' && req.method === 'POST') {
          let body = ''
          req.on('data', chunk => { body += chunk.toString() })
          req.on('end', async () => {
            try {
              let code = JSON.parse(body).code || ''
              
              if (code.includes('fn main()') && !code.includes('main();')) {
                code += '\nmain();\n'
              }
              
              // Write to temp file
              const tmpFile = path.join(process.cwd(), 'temp_playground.jb')
              fs.writeFileSync(tmpFile, code)
              
              // Run via Jabline
              // timeout: 5000ms prevents infinite loops
              let output = ''
              let isError = false
              try {
                const { stdout, stderr } = await execAsync(`jabline run temp_playground.jb`, { timeout: 5000 })
                output = stdout || stderr
              } catch (err) {
                isError = true
                output = err.stdout || err.stderr || err.message
              }
              
              // Cleanup
              if (fs.existsSync(tmpFile)) fs.unlinkSync(tmpFile)
              
              res.setHeader('Content-Type', 'application/json')
              res.end(JSON.stringify({ output: output.trim(), isError }))
            } catch (err) {
              res.statusCode = 500
              res.end(JSON.stringify({ error: err.message }))
            }
          })
          return
        }

        // ── 2. Blog API ───────────────────────────────────────────────────
        const blogFile = path.join(process.cwd(), 'src', 'data', 'blogs.json')
        
        // Initialize if not exists
        if (!fs.existsSync(blogFile)) {
          const defaultBlogs = [
            {
              id: 1,
              title: 'Introducing Jabline v0.6.0',
              excerpt: 'The biggest release yet — AOT type checker, object pool GC optimization, full LSP support, and a production-ready HTTP server.',
              date: '2026-05-08',
              category: 'Release',
              author: 'Jabline Team',
              readTime: '5 min',
              featured: true,
            }
          ]
          fs.mkdirSync(path.dirname(blogFile), { recursive: true })
          fs.writeFileSync(blogFile, JSON.stringify(defaultBlogs, null, 2))
        }

        if (req.url.startsWith('/api/blogs')) {
          res.setHeader('Content-Type', 'application/json')
          
          if (req.method === 'GET') {
            const data = fs.readFileSync(blogFile, 'utf8')
            res.end(data)
            return
          }
          
          if (req.method === 'POST') {
            let body = ''
            req.on('data', chunk => { body += chunk.toString() })
            req.on('end', () => {
              const newBlog = JSON.parse(body)
              const blogs = JSON.parse(fs.readFileSync(blogFile, 'utf8'))
              newBlog.id = Date.now()
              blogs.unshift(newBlog) // Add to top
              fs.writeFileSync(blogFile, JSON.stringify(blogs, null, 2))
              res.end(JSON.stringify(newBlog))
            })
            return
          }
          
          if (req.method === 'DELETE') {
            const id = parseInt(req.url.split('/').pop(), 10)
            const blogs = JSON.parse(fs.readFileSync(blogFile, 'utf8'))
            const filtered = blogs.filter(b => b.id !== id)
            fs.writeFileSync(blogFile, JSON.stringify(filtered, null, 2))
            res.end(JSON.stringify({ success: true }))
            return
          }
        }
        
        next()
      })
    }
  }
}

export default defineConfig({
  plugins: [react(), jablineDocsApi()],
  server: {
    port: 3000,
    open: true
  },
  build: {
    outDir: 'dist',
    sourcemap: false
  }
})
