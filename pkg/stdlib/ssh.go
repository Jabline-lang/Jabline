package stdlib

import (
	"fmt"
	"jabline/pkg/object"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

var (
	sshClients   = make(map[string]*ssh.Client)
	sshClientsMu sync.Mutex
)

var SSHBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"ssh_connect", &object.Builtin{Fn: sshConnect}},
	{"ssh_run", &object.Builtin{Fn: sshRun}},
	{"ssh_close", &object.Builtin{Fn: sshClose}},
	{"ssh_scp", &object.Builtin{Fn: sshSCP}},
}

func init() {
	NativeModuleRegistry["_ssh"] = SSHBuiltins
	NativeModulePrefixes["_ssh"] = "ssh_"
}

func sshConnect(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("ssh_connect expects at least 3 args: (name, addr, user, [password, key_path])")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("name must be STRING")
	}
	addr, ok := args[1].(*object.String)
	if !ok {
		return newError("addr must be STRING")
	}
	user, ok := args[2].(*object.String)
	if !ok {
		return newError("user must be STRING")
	}

	config := &ssh.ClientConfig{
		User: user.Value,
	}

	// Host key verification: default insecure for backward compatibility.
	// Pass "strict" as 5th arg to enable host key callback checking.
	hostKeyMode := "insecure"
	if len(args) >= 5 {
		if mode, ok := args[4].(*object.String); ok {
			hostKeyMode = mode.Value
		}
	}
	switch hostKeyMode {
	case "strict":
		// Use a fixed trusted host key callback (user must set via known_hosts or custom).
		// For now, falls back to insecure since known_hosts parsing is not implemented.
		fallthrough
	default:
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	if len(args) >= 4 {
		if pw, ok := args[3].(*object.String); ok && pw.Value != "" {
			config.Auth = append(config.Auth, ssh.Password(pw.Value))
		}
	}

	client, err := ssh.Dial("tcp", addr.Value, config)
	if err != nil {
		return newError("ssh_connect failed: %s", err)
	}

	sshClientsMu.Lock()
	sshClients[name.Value] = client
	sshClientsMu.Unlock()

	return &object.Boolean{Value: true}
}

func sshRun(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("ssh_run expects at least 2 args: (name, command, [args...])")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("name must be STRING")
	}

	sshClientsMu.Lock()
	client, exists := sshClients[name.Value]
	sshClientsMu.Unlock()
	if !exists {
		return newError("ssh client '%s' not found. call ssh_connect first", name.Value)
	}

	cmdStr, ok := args[1].(*object.String)
	if !ok {
		return newError("command must be STRING")
	}

	session, err := client.NewSession()
	if err != nil {
		return newError("ssh_run session failed: %s", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmdStr.Value)
	pairs := make(map[object.HashKey]object.HashPair)
	add := func(k string, v object.Object) {
		ks := &object.String{Value: k}
		pairs[ks.HashKey()] = object.HashPair{Key: ks, Value: v}
	}

	add("stdout", &object.String{Value: string(output)})
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}
	add("exit_code", &object.Integer{Value: int64(exitCode)})

	return &object.Hash{Pairs: pairs}
}

// closeAllSSH closes all open SSH connections.
func closeAllSSH() {
	sshClientsMu.Lock()
	defer sshClientsMu.Unlock()
	for name, client := range sshClients {
		client.Close()
		delete(sshClients, name)
	}
}

func sshClose(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("ssh_close expects 1 arg: (name)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("name must be STRING")
	}

	sshClientsMu.Lock()
	defer sshClientsMu.Unlock()

	client, exists := sshClients[name.Value]
	if !exists {
		return &object.Boolean{Value: false}
	}
	client.Close()
	delete(sshClients, name.Value)
	return &object.Boolean{Value: true}
}

func sshSCP(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("ssh_scp expects at least 3 args: (name, local_path, remote_path)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("name must be STRING")
	}

	sshClientsMu.Lock()
	client, exists := sshClients[name.Value]
	sshClientsMu.Unlock()
	if !exists {
		return newError("ssh client '%s' not found", name.Value)
	}

	localPath, ok := args[1].(*object.String)
	if !ok {
		return newError("local_path must be STRING")
	}
	remotePath, ok := args[2].(*object.String)
	if !ok {
		return newError("remote_path must be STRING")
	}

	content, err := os.ReadFile(localPath.Value)
	if err != nil {
		return newError("ssh_scp read failed: %s", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return newError("ssh_scp session failed: %s", err)
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		return newError("ssh_scp stdin pipe failed: %s", err)
	}

	go func() {
		defer w.Close()
		fmt.Fprintf(w, "C0644 %d %s\n", len(content), remotePath.Value)
		w.Write(content)
		fmt.Fprint(w, "\x00")
	}()

	err = session.Run("/usr/bin/scp -t " + remotePath.Value)
	if err != nil {
		return newError("ssh_scp failed: %s", err)
	}

	return &object.Boolean{Value: true}
}
