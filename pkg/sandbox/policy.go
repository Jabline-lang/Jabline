package sandbox

import (
	"fmt"
	"strings"
)

// Level defines the sandbox restriction level.
type Level int

const (
	LevelNone       Level = iota // No restrictions (current default)
	LevelSecure                  // Blocks FFI, restricts filesystem to CWD, read-only network
	LevelRestrictive             // Blocks FFI + network, restricts filesystem to specific paths
	LevelIsolated                // Blocks FFI + network + filesystem + spawn. Only pure computation.
)

func (l Level) String() string {
	switch l {
	case LevelNone:
		return "none"
	case LevelSecure:
		return "secure"
	case LevelRestrictive:
		return "restrictive"
	case LevelIsolated:
		return "isolated"
	default:
		return "unknown"
	}
}

func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "none":
		return LevelNone, nil
	case "secure":
		return LevelSecure, nil
	case "restrictive":
		return LevelRestrictive, nil
	case "isolated":
		return LevelIsolated, nil
	default:
		return LevelNone, fmt.Errorf("unknown sandbox level: %q (valid: none, secure, restrictive, isolated)", s)
	}
}

// Permission is a granular capability within a sandbox level.
type Permission int

const (
	PermFFI             Permission = iota // Load shared libraries and call native code
	PermFileRead                           // Read files from the filesystem
	PermFileWrite                          // Write files to the filesystem
	PermFileDelete                         // Delete files
	PermNetworkConnect                     // Make outbound network connections
	PermNetworkListen                      // Listen on network ports
	PermSpawn                              // Spawn concurrent processes
	PermEnvRead                            // Read environment variables
	PermEnvWrite                           // Write environment variables
	PermExec                               // Execute external commands
	PermStdlibAdvanced                     // Access advanced stdlib (crypto, db, etc.)
	PermModuleUnsafe                       // Import modules from unsafe sources
	PermTelemetry                          // Access telemetry (meter, trace)
	PermDebug                              // Access debugger
)

func (p Permission) String() string {
	names := map[Permission]string{
		PermFFI:             "ffi",
		PermFileRead:        "file_read",
		PermFileWrite:       "file_write",
		PermFileDelete:      "file_delete",
		PermNetworkConnect:  "network_connect",
		PermNetworkListen:   "network_listen",
		PermSpawn:           "spawn",
		PermEnvRead:         "env_read",
		PermEnvWrite:        "env_write",
		PermExec:            "exec",
		PermStdlibAdvanced:  "stdlib_advanced",
		PermModuleUnsafe:    "module_unsafe",
		PermTelemetry:       "telemetry",
		PermDebug:           "debug",
	}
	if name, ok := names[p]; ok {
		return name
	}
	return "unknown"
}

// Policy defines a complete sandbox configuration.
type Policy struct {
	// Level is the base restriction level.
	Level Level

	// AllowedPaths is a whitelist of filesystem paths that can be accessed.
	// Empty means no filesystem access (at restrictive/isolated).
	AllowedPaths []string

	// DeniedPaths overrides AllowedPaths for specific paths.
	DeniedPaths []string

	// AllowedHosts is a whitelist of network hosts/domains that can be connected to.
	// Empty means no network access (at restrictive level).
	AllowedHosts []string

	// DeniedHosts overrides AllowedHosts.
	DeniedHosts []string

	// AllowedEnvs is a whitelist of environment variables that can be read.
	// Empty means all env vars are readable (at secure level) or none (at restrictive+).
	AllowedEnvs []string

	// Overrides allows fine-grained permission toggling.
	// If set, these override the level-based defaults.
	Overrides map[Permission]bool

	// CustomError is a custom error message shown when a policy violation occurs.
	CustomError string
}

// DefaultPolicy returns the default policy for a given level.
func DefaultPolicy(level Level) *Policy {
	p := &Policy{
		Level:     level,
		Overrides: make(map[Permission]bool),
	}

	switch level {
	case LevelNone:
		p.AllowedPaths = []string{"/"}
		p.AllowedHosts = []string{"*"}
		for _, perm := range allPermissions() {
			p.Overrides[perm] = true
		}

	case LevelSecure:
		p.AllowedPaths = []string{"."}
		p.AllowedHosts = []string{"*"}
		p.AllowedEnvs = []string{"PATH", "HOME", "USER", "JABLINE_*"}
		for _, perm := range allPermissions() {
			p.Overrides[perm] = true
		}
		p.Overrides[PermFFI] = false
		p.Overrides[PermExec] = false
		p.Overrides[PermEnvWrite] = false
		p.Overrides[PermFileDelete] = false

	case LevelRestrictive:
		p.AllowedPaths = []string{"."}
		p.AllowedHosts = nil
		for _, perm := range allPermissions() {
			p.Overrides[perm] = false
		}
		p.Overrides[PermFileRead] = true
		p.Overrides[PermEnvRead] = true
		p.Overrides[PermSpawn] = true
		p.Overrides[PermTelemetry] = true

	case LevelIsolated:
		for _, perm := range allPermissions() {
			p.Overrides[perm] = false
		}
	}

	return p
}

// Merge applies overrides from another policy on top of this one.
func (p *Policy) Merge(other *Policy) {
	if other == nil {
		return
	}
	if other.Level > p.Level {
		p.Level = other.Level
	}
	p.AllowedPaths = append(p.AllowedPaths, other.AllowedPaths...)
	p.DeniedPaths = append(p.DeniedPaths, other.DeniedPaths...)
	p.AllowedHosts = append(p.AllowedHosts, other.AllowedHosts...)
	p.DeniedHosts = append(p.DeniedHosts, other.DeniedHosts...)
	p.AllowedEnvs = append(p.AllowedEnvs, other.AllowedEnvs...)
	for k, v := range other.Overrides {
		p.Overrides[k] = v
	}
	if other.CustomError != "" {
		p.CustomError = other.CustomError
	}
}

// Allowed checks if a specific permission is granted.
func (p *Policy) Allowed(perm Permission) bool {
	if v, ok := p.Overrides[perm]; ok {
		return v
	}
	return false
}

func allPermissions() []Permission {
	return []Permission{
		PermFFI, PermFileRead, PermFileWrite, PermFileDelete,
		PermNetworkConnect, PermNetworkListen,
		PermSpawn, PermEnvRead, PermEnvWrite, PermExec,
		PermStdlibAdvanced, PermModuleUnsafe,
		PermTelemetry, PermDebug,
	}
}

// ViolationError is returned when a policy is violated.
type ViolationError struct {
	Permission Permission
	Message    string
	Context    string
}

func (e *ViolationError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("sandbox violation: %s (permission: %s)", e.Message, e.Permission)
	}
	return fmt.Sprintf("sandbox violation: permission %s denied. context: %s", e.Permission, e.Context)
}

// PathAllowed checks if a given filesystem path is allowed by the policy.
func (p *Policy) PathAllowed(path string) bool {
	if !p.Allowed(PermFileRead) && !p.Allowed(PermFileWrite) {
		return false
	}

	for _, denied := range p.DeniedPaths {
		if strings.HasPrefix(path, denied) {
			return false
		}
	}

	if len(p.AllowedPaths) == 0 {
		return false
	}

	for _, allowed := range p.AllowedPaths {
		if allowed == "/" || allowed == "*" {
			return true
		}
		if strings.HasPrefix(path, allowed) {
			return true
		}
	}

	return false
}

// HostAllowed checks if a given network host is allowed by the policy.
func (p *Policy) HostAllowed(host string) bool {
	if !p.Allowed(PermNetworkConnect) && !p.Allowed(PermNetworkListen) {
		return false
	}

	for _, denied := range p.DeniedHosts {
		if denied == host || strings.HasSuffix(host, "."+denied) {
			return false
		}
	}

	if len(p.AllowedHosts) == 0 {
		return false
	}

	for _, allowed := range p.AllowedHosts {
		if allowed == "*" {
			return true
		}
		if allowed == host || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}

	return false
}

// EnvAllowed checks if a given environment variable is allowed by the policy.
func (p *Policy) EnvAllowed(name string) bool {
	if !p.Allowed(PermEnvRead) {
		return false
	}
	if len(p.AllowedEnvs) == 0 {
		return true
	}
	for _, allowed := range p.AllowedEnvs {
		if allowed == name {
			return true
		}
		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(name, prefix) {
				return true
			}
		}
	}
	return false
}
