package jpm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	ModFileName    = "jabline.toml"
	LockFileName   = "jabline.lock"
	DefaultVersion = "0.1.0"
)

type ProjectConfig struct {
	Project         ProjectMetadata   `toml:"project"`
	Dependencies    map[string]string `toml:"dependencies"`
	DevDependencies map[string]string `toml:"dev-dependencies,omitempty"`
}

type ProjectMetadata struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
}

// InitProject creates a new Jabline project with default files in the specified path.
func InitProject(projectName string, targetPath string, template string) error {
	if targetPath != "" && targetPath != "." {
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	config := ProjectConfig{
		Project: ProjectMetadata{
			Name:        projectName,
			Version:     DefaultVersion,
			Description: "A new Jabline project",
		},
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
	}

	configPath := filepath.Join(targetPath, ModFileName)
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", ModFileName, err)
	}

	switch template {
	case "hello", "":
		if err := writeTemplateFile(targetPath, "main.jb", `echo("Hello World!");
`); err != nil {
			return err
		}
	case "http-server":
		if err := writeTemplateFile(targetPath, "main.jb", `// Cloud-native HTTP server with health checks and metrics
import "config"
import "health"
import "log"
import "metrics"
import "http_advanced"

// Load configuration from env vars with defaults
let port = 8080
let name = config.env("SERVICE_NAME", "my-service")
let level = config.env("LOG_LEVEL", "info")
log.set_level(level)
log.info("Starting " + name, {port: port, version: "1.0.0"})

// Register health checks
health.set_liveness(fn() { return true })
health.set_readiness(fn() {
    // Check dependencies here (DB, Redis, etc.)
    return true
})

// Register Prometheus metrics
metrics.counter("http_requests_total", "Total HTTP requests", ["method", "path"])
metrics.counter("http_errors_total", "Total HTTP errors", ["method"])
metrics.histogram("http_request_duration_ms", "Request duration in ms", [5, 10, 25, 50, 100, 250, 500, 1000], ["method"])

// Setup HTTP mux with health and metrics endpoints
http_advanced.mux()
http_advanced.healthz_handler()
http_advanced.readyz_handler()
http_advanced.metrics_handler()

// Mount API routes
http_advanced.mount("GET /api/hello", fn(req) {
    metrics.inc("http_requests_total", "GET", "/api/hello")
    return {status: 200, body: '{"message": "Hello from Jabline!"}', headers: {"Content-Type": "application/json"}}
})

http_advanced.mount("POST /api/echo", fn(req) {
    metrics.inc("http_requests_total", "POST", "/api/echo")
    return {status: 200, body: req.body, headers: {"Content-Type": "application/octet-stream"}}
})

// Start server
log.info("Starting HTTP server on port " + to_str(port))
http_serve(port, fn(req) {
    metrics.inc("http_requests_total", req.method, req.path)
    return {status: 404, body: '{"error": "not found"}', headers: {"Content-Type": "application/json"}}
})
`); err != nil {
			return err
		}
		writeTemplateFile(targetPath, "Dockerfile", `FROM jabline/jabline:latest
WORKDIR /app
COPY . .
EXPOSE 8080
CMD ["run", "main.jb"]
`)
	case "api":
		if err := writeTemplateFile(targetPath, "main.jb", `// REST API with CORS middleware and structured logging
import "json"
import "http_advanced"
import "log"
import "metrics"
import "config"

log.set_level(config.env("LOG_LEVEL", "info"))
let port = to_int(config.env("PORT", "8080"))
log.info("Starting API server", {port: port})

// Register metrics
metrics.counter("api_requests", "API requests", ["method", "endpoint"])

// Setup with health checks
http_advanced.mux()
http_advanced.healthz_handler()
http_advanced.readyz_handler()
http_advanced.metrics_handler()

// CORS middleware wrapper
let cors = http_advanced.cors()

// Users resource
let users = [
    {id: 1, name: "Alice", email: "alice@example.com"},
    {id: 2, name: "Bob", email: "bob@example.com"},
]

http_advanced.mount("GET /api/users", cors(fn(req) {
    metrics.inc("api_requests", "GET", "/api/users")
    return {status: 200, body: json.encode(users), headers: {"Content-Type": "application/json"}}
}))

http_advanced.mount("GET /api/health", fn(req) {
    return {status: 200, body: '{"status": "ok", "service": "jabline-api"}'}
})

http_advanced.mount("POST /api/echo", cors(fn(req) {
    metrics.inc("api_requests", "POST", "/api/echo")
    return {status: 200, body: req.body, headers: {"Content-Type": "application/json"}}
}))

log.info("Serving on port " + to_str(port))
http_serve(port, fn(req) {
    return {status: 404, body: '{"error": "not found"}'}
})
`); err != nil {
			return err
		}
	case "worker":
		if err := writeTemplateFile(targetPath, "main.jb", `// Concurrent worker with backpressure, structured concurrency, and metrics
import "sync"
import "parallel"
import "metrics"
import "log"
import "context"
import "config"

log.set_level(config.env("LOG_LEVEL", "info"))
let numWorkers = to_int(config.env("NUM_WORKERS", "4"))
let queueSize = to_int(config.env("QUEUE_SIZE", "100"))
log.info("Starting worker pool", {workers: numWorkers, queueSize: queueSize})

// Register metrics
metrics.gauge("tasks_processed", "Tasks processed", ["status"])
metrics.gauge("active_workers", "Active workers")
metrics.counter("tasks_total", "Total tasks")

// Create worker pool
parallel.pool("workers", numWorkers)

// Work function
let processTask = fn(task) {
    metrics.gauge_inc("active_workers")
    metrics.inc("tasks_total")
    log.info("Processing task", {id: task.id, payload: task.data})
    // Simulate work
    let result = task.data * 2
    metrics.gauge_set("tasks_processed", 1, "completed")
    metrics.gauge_dec("active_workers")
    return result
}

// Submit tasks with timeout
let results = []
for (let i = 0; i < 10; i++) {
    let task = {id: i, data: i * 100}
    with_timeout(5000, fn(ctx) {
        parallel.pool_submit("workers", processTask, task)
    })
}

// Wait for completion
let output = parallel.pool_await("workers")
log.info("All tasks completed", {results: output})
echo("Results: " + json.encode(output))
`); err != nil {
			return err
		}
	case "microservice":
		if err := writeTemplateFile(targetPath, "main.jb", `// Production-ready microservice with config, health, metrics, and graceful shutdown
import "config"
import "health"
import "log"
import "metrics"
import "http_advanced"
import "json"

// Configuration
let cfg = {
    port: to_int(config.env("PORT", "8080")),
    name: config.env("SERVICE_NAME", "jabline-svc"),
    version: config.env("VERSION", "1.0.0"),
    env: config.env("ENV", "development"),
}
log.set_level(config.env("LOG_LEVEL", "info"))

log.info("Initializing " + cfg.name, {version: cfg.version, env: cfg.env})

// Health checks
health.set_liveness(fn() { return true })
health.set_readiness(fn() {
    // Extend this to check DB/Redis connectivity
    return {healthy: true, checks: {database: "ok", redis: "ok"}}
})

// Metrics
metrics.counter("svc_requests", "Service requests", ["method", "endpoint", "status"])
metrics.histogram("svc_duration_ms", "Request duration", [1, 5, 10, 25, 50, 100, 250, 500, 1000])
metrics.gauge("svc_connections", "Active connections")

// Router
http_advanced.mux()
http_advanced.healthz_handler()
http_advanced.readyz_handler()
http_advanced.metrics_handler()
http_advanced.static("/static", "./public")

// Root endpoint
http_advanced.mount("GET /", fn(req) {
    let info = {
        service: cfg.name,
        version: cfg.version,
        env: cfg.env,
    }
    return {status: 200, body: json.encode(info), headers: {"Content-Type": "application/json"}}
})

// Health check detail
http_advanced.mount("GET /api/v1/status", fn(req) {
    let liveness = health.liveness()
    let readiness = health.readiness()
    return {
        status: 200,
        body: json.encode({liveness: liveness, readiness: readiness, uptime: "running"}),
        headers: {"Content-Type": "application/json"},
    }
})

log.info(cfg.name + " listening on port " + to_str(cfg.port))
http_serve(cfg.port, fn(req) {
    return {status: 404, body: '{"error": "endpoint not found"}', headers: {"Content-Type": "application/json"}}
})
`); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown template %q (use: hello, http-server, api, worker, microservice)", template)
	}

	// Create .gitignore
	gitignorePath := filepath.Join(targetPath, ".gitignore")
	gitignoreContent := `# Jabline binaries
*.exe
jabline
jabline_debug

# Dependency directory
lib/

# Local cache
.jb_cache/
`
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	return nil
}

func writeTemplateFile(targetPath, name, content string) error {
	fp := filepath.Join(targetPath, name)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return os.WriteFile(fp, []byte(content), 0644)
	}
	return nil
}

func LoadProject() (*ProjectConfig, error) {
	data, err := os.ReadFile(ModFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", ModFileName, err)
	}

	var config ProjectConfig
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", ModFileName, err)
	}

	if config.Dependencies == nil {
		config.Dependencies = make(map[string]string)
	}
	if config.DevDependencies == nil {
		config.DevDependencies = make(map[string]string)
	}

	return &config, nil
}

func SaveProject(config *ProjectConfig) error {
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	err = os.WriteFile(ModFileName, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", ModFileName, err)
	}

	return nil
}

// DevDependencies holds development-only dependencies.
type DevDependency struct {
	URL string `toml:"url"`
}

// AddDevDependency adds a development dependency.
func (c *ProjectConfig) AddDevDependency(url string) {
	if c.DevDependencies == nil {
		c.DevDependencies = make(map[string]string)
	}
	name := filepath.Base(url)
	name = strings.TrimSuffix(name, ".git")
	c.DevDependencies[name] = url
}

// GetDefaultProjectName returns the name of the current directory if not specified.
func GetDefaultProjectName() string {
	dir, err := os.Getwd()
	if err != nil {
		return "my-project"
	}
	return filepath.Base(dir)
}

// AddDependency adds a dependency to the project config.
func (c *ProjectConfig) AddDependency(url string) {
	if c.Dependencies == nil {
		c.Dependencies = make(map[string]string)
	}
	// Use the last part of the URL as the package name for now
	name := filepath.Base(url)
	name = strings.TrimSuffix(name, ".git")
	c.Dependencies[name] = url
}

// LoadPackageConfig loads a package's jabline.toml from the lib directory.
func LoadPackageConfig(libDir, name string) (*ProjectConfig, error) {
	path := filepath.Join(libDir, name, ModFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s for %q: %w", ModFileName, name, err)
	}
	var config ProjectConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s for %q: %w", ModFileName, name, err)
	}
	if config.Dependencies == nil {
		config.Dependencies = make(map[string]string)
	}
	return &config, nil
}

// RemoveDependency removes a dependency by name.
func (c *ProjectConfig) RemoveDependency(name string) bool {
	if c.Dependencies == nil {
		return false
	}
	if _, ok := c.Dependencies[name]; ok {
		delete(c.Dependencies, name)
		return true
	}
	return false
}
