package vm

import (
	"fmt"

	"jabline/pkg/object"
	"jabline/pkg/sandbox"
)

// CheckPermission panics via error return if the VM's sandbox policy denies the permission.
func (vm *VM) CheckPermission(perm sandbox.Permission, context string) error {
	if vm.Sandbox == nil {
		return nil
	}
	if !vm.Sandbox.Allowed(perm) {
		return &sandbox.ViolationError{
			Permission: perm,
			Context:    context,
			Message:    vm.Sandbox.CustomError,
		}
	}
	return nil
}

// CheckFilePath verifies that a filesystem path is allowed by the sandbox.
func (vm *VM) CheckFilePath(path string, write bool) error {
	if vm.Sandbox == nil {
		return nil
	}
	perm := sandbox.PermFileRead
	if write {
		perm = sandbox.PermFileWrite
	}
	if !vm.Sandbox.Allowed(perm) {
		return &sandbox.ViolationError{
			Permission: perm,
			Context:    path,
		}
	}
	if !vm.Sandbox.PathAllowed(path) {
		return &sandbox.ViolationError{
			Permission: perm,
			Context:    fmt.Sprintf("path not in allowed set: %s", path),
		}
	}
	return nil
}

// CheckHost verifies that a network host is allowed by the sandbox.
func (vm *VM) CheckHost(host string) error {
	if vm.Sandbox == nil {
		return nil
	}
	if !vm.Sandbox.HostAllowed(host) {
		return &sandbox.ViolationError{
			Permission: sandbox.PermNetworkConnect,
			Context:    host,
		}
	}
	return nil
}

// CheckEnv verifies that an environment variable is allowed by the sandbox.
func (vm *VM) CheckEnv(name string) error {
	if vm.Sandbox == nil {
		return nil
	}
	if !vm.Sandbox.EnvAllowed(name) {
		return &sandbox.ViolationError{
			Permission: sandbox.PermEnvRead,
			Context:    fmt.Sprintf("env: %s", name),
		}
	}
	return nil
}

// SetSandboxLevel configures the VM's sandbox to a named level.
func (vm *VM) SetSandboxLevel(levelStr string) error {
	level, err := sandbox.ParseLevel(levelStr)
	if err != nil {
		return err
	}
	vm.Sandbox = sandbox.DefaultPolicy(level)
	return nil
}

// SandboxGetErrorObj returns an Error object for a sandbox violation.
// This is used by builtins that detect violations.
func SandboxGetErrorObj(err error) *object.Error {
	return &object.Error{Message: err.Error()}
}

// SandboxToObject converts a sandbox.Level to a Jabline Integer object.
func SandboxToObject(level sandbox.Level) object.Object {
	return &object.Integer{Value: int64(level)}
}
