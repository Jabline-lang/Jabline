package vm

import (
	"testing"

	"jabline/pkg/sandbox"
)

func TestSandboxDefaultNone(t *testing.T) {
	vm := &VM{Sandbox: sandbox.DefaultPolicy(sandbox.LevelNone)}
	if err := vm.CheckPermission(sandbox.PermFFI, "test"); err != nil {
		t.Errorf("expected no error at LevelNone, got: %v", err)
	}
	if err := vm.CheckPermission(sandbox.PermFileRead, "test"); err != nil {
		t.Errorf("expected no error at LevelNone, got: %v", err)
	}
	if err := vm.CheckPermission(sandbox.PermNetworkConnect, "test"); err != nil {
		t.Errorf("expected no error at LevelNone, got: %v", err)
	}
	if err := vm.CheckPermission(sandbox.PermExec, "test"); err != nil {
		t.Errorf("expected no error at LevelNone, got: %v", err)
	}
}

func TestSandboxSecure_BlocksFFI(t *testing.T) {
	vm := &VM{Sandbox: sandbox.DefaultPolicy(sandbox.LevelSecure)}
	if err := vm.CheckPermission(sandbox.PermFFI, "test"); err == nil {
		t.Errorf("expected error for FFI at LevelSecure")
	}
	if err := vm.CheckPermission(sandbox.PermFileRead, "test"); err != nil {
		t.Errorf("expected no error for FileRead at LevelSecure, got: %v", err)
	}
	if err := vm.CheckPermission(sandbox.PermNetworkConnect, "test"); err != nil {
		t.Errorf("expected no error for Network at LevelSecure, got: %v", err)
	}
	if err := vm.CheckPermission(sandbox.PermExec, "test"); err == nil {
		t.Errorf("expected error for Exec at LevelSecure")
	}
}

func TestSandboxRestrictive_BlocksNetwork(t *testing.T) {
	vm := &VM{Sandbox: sandbox.DefaultPolicy(sandbox.LevelRestrictive)}
	if err := vm.CheckPermission(sandbox.PermFFI, "test"); err == nil {
		t.Errorf("expected error for FFI at LevelRestrictive")
	}
	if err := vm.CheckPermission(sandbox.PermNetworkConnect, "test"); err == nil {
		t.Errorf("expected error for Network at LevelRestrictive")
	}
	if err := vm.CheckPermission(sandbox.PermFileRead, "test"); err != nil {
		t.Errorf("expected no error for FileRead at LevelRestrictive, got: %v", err)
	}
	if err := vm.CheckPermission(sandbox.PermSpawn, "test"); err != nil {
		t.Errorf("expected no error for Spawn at LevelRestrictive, got: %v", err)
	}
}

func TestSandboxIsolated_BlocksAll(t *testing.T) {
	vm := &VM{Sandbox: sandbox.DefaultPolicy(sandbox.LevelIsolated)}
	perms := []sandbox.Permission{
		sandbox.PermFFI, sandbox.PermFileRead, sandbox.PermFileWrite,
		sandbox.PermNetworkConnect, sandbox.PermSpawn, sandbox.PermExec,
		sandbox.PermEnvRead, sandbox.PermTelemetry,
	}
	for _, perm := range perms {
		if err := vm.CheckPermission(perm, "test"); err == nil {
			t.Errorf("expected error for %s at LevelIsolated", perm)
		}
	}
}

func TestSandboxPathAllowed(t *testing.T) {
	p := sandbox.DefaultPolicy(sandbox.LevelRestrictive)
	p.AllowedPaths = []string{"/home", "/tmp"}
	p.DeniedPaths = []string{"/etc/"}
	if !p.PathAllowed("/home/user/file.txt") {
		t.Error("expected /home/user/file.txt to be allowed")
	}
	if p.PathAllowed("/etc/passwd") {
		t.Error("expected /etc/passwd to be denied")
	}
	if !p.PathAllowed("/tmp/foo.txt") {
		t.Error("expected /tmp/foo.txt to be allowed")
	}
	if p.PathAllowed("/var/log/syslog") {
		t.Error("expected /var/log/syslog to be denied")
	}
}

func TestSandboxHostAllowed(t *testing.T) {
	p := sandbox.DefaultPolicy(sandbox.LevelSecure)
	p.DeniedHosts = []string{"evil.com"}
	if p.HostAllowed("evil.com") {
		t.Error("expected evil.com to be denied")
	}
	if !p.HostAllowed("example.com") {
		t.Error("expected example.com to be allowed")
	}
}

func TestSandboxEnvAllowed(t *testing.T) {
	p := sandbox.DefaultPolicy(sandbox.LevelSecure)
	if !p.EnvAllowed("PATH") {
		t.Error("expected PATH to be allowed")
	}
	if p.EnvAllowed("SECRET_KEY") {
		t.Error("expected SECRET_KEY to be denied")
	}
}

func TestSandboxMergePreservesRestrictions(t *testing.T) {
	base := sandbox.DefaultPolicy(sandbox.LevelRestrictive)
	override := &sandbox.Policy{
		Overrides: map[sandbox.Permission]bool{
			sandbox.PermNetworkConnect: true,
		},
	}
	base.Merge(override)
	if !base.Allowed(sandbox.PermNetworkConnect) {
		t.Error("expected network to be allowed after merge")
	}
	if base.Allowed(sandbox.PermFFI) {
		t.Error("expected FFI to still be denied after merge")
	}
}

func TestSandboxLevelString(t *testing.T) {
	cases := []struct {
		level sandbox.Level
		want  string
	}{
		{sandbox.LevelNone, "none"},
		{sandbox.LevelSecure, "secure"},
		{sandbox.LevelRestrictive, "restrictive"},
		{sandbox.LevelIsolated, "isolated"},
	}
	for _, c := range cases {
		if got := c.level.String(); got != c.want {
			t.Errorf("Level(%d).String() = %q, want %q", c.level, got, c.want)
		}
	}
}

func TestSandboxParseLevel(t *testing.T) {
	if l, err := sandbox.ParseLevel("isolated"); err != nil || l != sandbox.LevelIsolated {
		t.Errorf("ParseLevel(isolated) = %d, %v", l, err)
	}
	if _, err := sandbox.ParseLevel("invalid"); err == nil {
		t.Error("expected error for invalid level")
	}
}
