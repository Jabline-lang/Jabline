package sandbox

import (
	"testing"
)

func TestLevelString(t *testing.T) {
	cases := []struct {
		level Level
		want  string
	}{
		{LevelNone, "none"},
		{LevelSecure, "secure"},
		{LevelRestrictive, "restrictive"},
		{LevelIsolated, "isolated"},
		{Level(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.level.String(); got != c.want {
			t.Errorf("Level(%d).String() = %q, want %q", c.level, got, c.want)
		}
	}
}

func TestParseLevel(t *testing.T) {
	cases := []struct {
		input string
		want  Level
		ok    bool
	}{
		{"none", LevelNone, true},
		{"None", LevelNone, true},
		{"NONE", LevelNone, true},
		{"secure", LevelSecure, true},
		{"restrictive", LevelRestrictive, true},
		{"isolated", LevelIsolated, true},
		{"", LevelNone, false},
		{"invalid", LevelNone, false},
		{"strict", LevelNone, false},
	}
	for _, c := range cases {
		got, err := ParseLevel(c.input)
		if c.ok && err != nil {
			t.Errorf("ParseLevel(%q) unexpected error: %v", c.input, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ParseLevel(%q) expected error, got level %d", c.input, got)
		}
		if c.ok && got != c.want {
			t.Errorf("ParseLevel(%q) = %d, want %d", c.input, got, c.want)
		}
	}
}

func TestDefaultPolicyLevelNone(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	if p.Level != LevelNone {
		t.Errorf("Level = %d, want %d", p.Level, LevelNone)
	}
	for _, perm := range allPermissions() {
		if !p.Allowed(perm) {
			t.Errorf("expected %s to be allowed at LevelNone", perm)
		}
	}
	if !p.PathAllowed("/any/path") {
		t.Error("expected any path to be allowed at LevelNone")
	}
	if !p.HostAllowed("any-host.com") {
		t.Error("expected any host to be allowed at LevelNone")
	}
}

func TestDefaultPolicyLevelSecure(t *testing.T) {
	p := DefaultPolicy(LevelSecure)
	if p.Level != LevelSecure {
		t.Errorf("Level = %d, want %d", p.Level, LevelSecure)
	}
	// Should allow: most things
	if !p.Allowed(PermFileRead) {
		t.Error("expected FileRead allowed at LevelSecure")
	}
	if !p.Allowed(PermFileWrite) {
		t.Error("expected FileWrite allowed at LevelSecure")
	}
	if !p.Allowed(PermNetworkConnect) {
		t.Error("expected NetworkConnect allowed at LevelSecure")
	}
	// Should block: FFI, Exec, EnvWrite, FileDelete
	if p.Allowed(PermFFI) {
		t.Error("expected FFI denied at LevelSecure")
	}
	if p.Allowed(PermExec) {
		t.Error("expected Exec denied at LevelSecure")
	}
	if p.Allowed(PermEnvWrite) {
		t.Error("expected EnvWrite denied at LevelSecure")
	}
	if p.Allowed(PermFileDelete) {
		t.Error("expected FileDelete denied at LevelSecure")
	}
	// Path should be restricted to current directory
	if !p.PathAllowed(".") {
		t.Error("expected '.' to be allowed at LevelSecure")
	}
	if p.PathAllowed("/etc/passwd") {
		t.Error("expected '/etc/passwd' to be denied at LevelSecure")
	}
	// Hosts should be all allowed
	if !p.HostAllowed("example.com") {
		t.Error("expected any host to be allowed at LevelSecure")
	}
	// Env vars should have PATH, HOME, USER, JABLINE_*
	if !p.EnvAllowed("PATH") {
		t.Error("expected PATH to be allowed")
	}
	if !p.EnvAllowed("HOME") {
		t.Error("expected HOME to be allowed")
	}
	if !p.EnvAllowed("JABLINE_HOME") {
		t.Error("expected JABLINE_HOME to be allowed")
	}
	if p.EnvAllowed("SECRET_KEY") {
		t.Error("expected SECRET_KEY to be denied")
	}
}

func TestDefaultPolicyLevelRestrictive(t *testing.T) {
	p := DefaultPolicy(LevelRestrictive)
	if p.Level != LevelRestrictive {
		t.Errorf("Level = %d, want %d", p.Level, LevelRestrictive)
	}
	// Should allow: FileRead, EnvRead, Spawn, Telemetry
	if !p.Allowed(PermFileRead) {
		t.Error("expected FileRead allowed at LevelRestrictive")
	}
	if !p.Allowed(PermEnvRead) {
		t.Error("expected EnvRead allowed at LevelRestrictive")
	}
	if !p.Allowed(PermSpawn) {
		t.Error("expected Spawn allowed at LevelRestrictive")
	}
	if !p.Allowed(PermTelemetry) {
		t.Error("expected Telemetry allowed at LevelRestrictive")
	}
	// Should block: everything else
	if p.Allowed(PermFFI) {
		t.Error("expected FFI denied at LevelRestrictive")
	}
	if p.Allowed(PermFileWrite) {
		t.Error("expected FileWrite denied at LevelRestrictive")
	}
	if p.Allowed(PermFileDelete) {
		t.Error("expected FileDelete denied at LevelRestrictive")
	}
	if p.Allowed(PermNetworkConnect) {
		t.Error("expected NetworkConnect denied at LevelRestrictive")
	}
	if p.Allowed(PermNetworkListen) {
		t.Error("expected NetworkListen denied at LevelRestrictive")
	}
	if p.Allowed(PermExec) {
		t.Error("expected Exec denied at LevelRestrictive")
	}
	if p.Allowed(PermEnvWrite) {
		t.Error("expected EnvWrite denied at LevelRestrictive")
	}
	if p.Allowed(PermStdlibAdvanced) {
		t.Error("expected StdlibAdvanced denied at LevelRestrictive")
	}
	if p.Allowed(PermModuleUnsafe) {
		t.Error("expected ModuleUnsafe denied at LevelRestrictive")
	}
	if p.Allowed(PermDebug) {
		t.Error("expected Debug denied at LevelRestrictive")
	}
	// Path should be restricted to current directory
	if !p.PathAllowed("./foo.jb") {
		t.Error("expected './foo.jb' to be allowed at LevelRestrictive")
	}
	if p.PathAllowed("/etc/passwd") {
		t.Error("expected '/etc/passwd' to be denied at LevelRestrictive")
	}
	// Hosts should be none
	if p.HostAllowed("example.com") {
		t.Error("expected no hosts allowed at LevelRestrictive")
	}
}

func TestDefaultPolicyLevelIsolated(t *testing.T) {
	p := DefaultPolicy(LevelIsolated)
	if p.Level != LevelIsolated {
		t.Errorf("Level = %d, want %d", p.Level, LevelIsolated)
	}
	for _, perm := range allPermissions() {
		if p.Allowed(perm) {
			t.Errorf("expected %s to be denied at LevelIsolated", perm)
		}
	}
	if p.PathAllowed("/any/path") {
		t.Error("expected no path to be allowed at LevelIsolated")
	}
	if p.HostAllowed("any-host.com") {
		t.Error("expected no host to be allowed at LevelIsolated")
	}
}

func TestPathAllowed(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	p.AllowedPaths = []string{"/home", "/tmp"}
	p.DeniedPaths = []string{"/etc/"}

	cases := []struct {
		path  string
		allow bool
	}{
		{"/home/user/file.txt", true},
		{"/tmp/foo.txt", true},
		{"/etc/passwd", false},
		{"/var/log/syslog", false},
		{"/home", true},
		{"/tmp", true},
	}
	for _, c := range cases {
		got := p.PathAllowed(c.path)
		if got != c.allow {
			t.Errorf("PathAllowed(%q) = %v, want %v", c.path, got, c.allow)
		}
	}
}

func TestPathAllowedWildcard(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	p.AllowedPaths = []string{"/"}
	if !p.PathAllowed("/any/path") {
		t.Error("expected '/' wildcard to allow all paths")
	}
	p.AllowedPaths = []string{"*"}
	if !p.PathAllowed("/any/path") {
		t.Error("expected '*' wildcard to allow all paths")
	}
}

func TestPathAllowedNoPerms(t *testing.T) {
	p := DefaultPolicy(LevelIsolated)
	if p.PathAllowed("/tmp/foo") {
		t.Error("expected false when file permissions are denied")
	}
}

func TestPathAllowedEmptyAllowed(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	p.AllowedPaths = nil
	if p.PathAllowed("/tmp/foo") {
		t.Error("expected false when AllowedPaths is empty")
	}
}

func TestHostAllowed(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	p.DeniedHosts = []string{"evil.com", "malware.net"}

	cases := []struct {
		host  string
		allow bool
	}{
		{"example.com", true},
		{"evil.com", false},
		{"sub.evil.com", false},
		{"malware.net", false},
		{"safe.org", true},
	}
	for _, c := range cases {
		got := p.HostAllowed(c.host)
		if got != c.allow {
			t.Errorf("HostAllowed(%q) = %v, want %v", c.host, got, c.allow)
		}
	}
}

func TestHostAllowedNoPerms(t *testing.T) {
	p := DefaultPolicy(LevelIsolated)
	if p.HostAllowed("example.com") {
		t.Error("expected false when network permissions are denied")
	}
}

func TestHostAllowedEmptyAllowed(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	p.AllowedHosts = nil
	if p.HostAllowed("example.com") {
		t.Error("expected false when AllowedHosts is empty")
	}
}

func TestEnvAllowed(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	p.AllowedEnvs = []string{"PATH", "HOME", "JABLINE_*"}

	cases := []struct {
		name  string
		allow bool
	}{
		{"PATH", true},
		{"HOME", true},
		{"JABLINE_HOME", true},
		{"JABLINE_CONFIG_DIR", true},
		{"SECRET_KEY", false},
		{"API_TOKEN", false},
	}
	for _, c := range cases {
		got := p.EnvAllowed(c.name)
		if got != c.allow {
			t.Errorf("EnvAllowed(%q) = %v, want %v", c.name, got, c.allow)
		}
	}
}

func TestEnvAllowedNoReadPerm(t *testing.T) {
	p := DefaultPolicy(LevelIsolated)
	if p.EnvAllowed("PATH") {
		t.Error("expected false when EnvRead is denied")
	}
}

func TestEnvAllowedEmptyAllowed(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	p.AllowedEnvs = nil
	if !p.EnvAllowed("ANY_VAR") {
		t.Error("expected all env vars allowed when AllowedEnvs is empty")
	}
}

func TestMerge(t *testing.T) {
	base := DefaultPolicy(LevelRestrictive)
	override := &Policy{
		AllowedPaths: []string{"/extra"},
		DeniedPaths:  []string{"/secret"},
		AllowedHosts: []string{"extra.com"},
		DeniedHosts:  []string{"blocked.com"},
		AllowedEnvs:  []string{"EXTRA_VAR"},
		Overrides: map[Permission]bool{
			PermNetworkConnect: true,
		},
		CustomError: "custom",
	}
	base.Merge(override)

	if !base.Allowed(PermNetworkConnect) {
		t.Error("expected NetworkConnect allowed after merge")
	}
	if base.Allowed(PermFFI) {
		t.Error("expected FFI still denied after merge")
	}
	if !base.PathAllowed("/extra/file") {
		t.Error("expected /extra/file allowed after merge")
	}
	if base.PathAllowed("/secret/data") {
		t.Error("expected /secret/data denied after merge")
	}
	if !base.HostAllowed("extra.com") {
		t.Error("expected extra.com allowed after merge")
	}
	if base.HostAllowed("blocked.com") {
		t.Error("expected blocked.com denied after merge")
	}
	if !base.EnvAllowed("EXTRA_VAR") {
		t.Error("expected EXTRA_VAR allowed after merge")
	}
}

func TestMergeWithNil(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	p.Merge(nil)
	if !p.Allowed(PermFFI) {
		t.Error("expected no change after nil merge")
	}
}

func TestMergeStrictestLevel(t *testing.T) {
	base := DefaultPolicy(LevelSecure)
	override := DefaultPolicy(LevelRestrictive)
	base.Merge(override)
	if base.Level != LevelRestrictive {
		t.Errorf("Level = %d, want %d (strictest)", base.Level, LevelRestrictive)
	}
}

func TestAllowedUnknownPermission(t *testing.T) {
	p := DefaultPolicy(LevelNone)
	if p.Allowed(Permission(99)) {
		t.Error("expected unknown permission to be denied (not in Overrides)")
	}
	p2 := DefaultPolicy(LevelIsolated)
	if p2.Allowed(Permission(99)) {
		t.Error("expected unknown permission to be denied at Isolated")
	}
}

func TestViolationError(t *testing.T) {
	err := &ViolationError{
		Permission: PermFileRead,
		Context:    "/etc/passwd",
	}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}

	err2 := &ViolationError{
		Permission: PermFFI,
		Message:    "FFI is disabled",
		Context:    "test",
	}
	msg2 := err2.Error()
	if msg2 == "" {
		t.Error("expected non-empty error message")
	}
}

func TestPermissionString(t *testing.T) {
	cases := []struct {
		perm Permission
		want string
	}{
		{PermFFI, "ffi"},
		{PermFileRead, "file_read"},
		{PermFileWrite, "file_write"},
		{PermFileDelete, "file_delete"},
		{PermNetworkConnect, "network_connect"},
		{PermNetworkListen, "network_listen"},
		{PermSpawn, "spawn"},
		{PermEnvRead, "env_read"},
		{PermEnvWrite, "env_write"},
		{PermExec, "exec"},
		{PermStdlibAdvanced, "stdlib_advanced"},
		{PermModuleUnsafe, "module_unsafe"},
		{PermTelemetry, "telemetry"},
		{PermDebug, "debug"},
		{Permission(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.perm.String(); got != c.want {
			t.Errorf("Permission(%d).String() = %q, want %q", c.perm, got, c.want)
		}
	}
}

func TestLevelOrder(t *testing.T) {
	if LevelNone >= LevelSecure {
		t.Error("expected LevelNone < LevelSecure")
	}
	if LevelSecure >= LevelRestrictive {
		t.Error("expected LevelSecure < LevelRestrictive")
	}
	if LevelRestrictive >= LevelIsolated {
		t.Error("expected LevelRestrictive < LevelIsolated")
	}
}
