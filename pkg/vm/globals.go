package vm

import (
	"jabline/pkg/object"
	"sync"
)

// GlobalStore is a thread-safe store for global variables.
// It replaces the raw []object.Object slice to prevent data races
// when multiple VM forks (goroutines) share the same globals.
type GlobalStore struct {
	mu      sync.RWMutex
	values  []object.Object
}

// NewGlobalStore creates a GlobalStore with the given capacity.
func NewGlobalStore(size int) *GlobalStore {
	return &GlobalStore{
		values: make([]object.Object, size),
	}
}

// GlobalStoreFromSlice wraps an existing slice in a GlobalStore.
// Used during hot reload or state transfer.
func GlobalStoreFromSlice(vals []object.Object) *GlobalStore {
	cp := make([]object.Object, len(vals))
	copy(cp, vals)
	return &GlobalStore{values: cp}
}

// Get reads a global variable at the given index. Thread-safe.
func (g *GlobalStore) Get(index int) object.Object {
	g.mu.RLock()
	v := g.values[index]
	g.mu.RUnlock()
	return v
}

// Set writes a value to a global variable at the given index. Thread-safe.
func (g *GlobalStore) Set(index int, val object.Object) {
	g.mu.Lock()
	g.values[index] = val
	g.mu.Unlock()
}

// Snapshot returns a point-in-time copy of all globals.
// Used when a fork needs its own isolated globals.
func (g *GlobalStore) Snapshot() []object.Object {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cp := make([]object.Object, len(g.values))
	copy(cp, g.values)
	return cp
}

// Len returns the number of slots in the store.
func (g *GlobalStore) Len() int {
	return len(g.values)
}

// GlobalsLen returns the number of global slots in the VM.
func (vm *VM) GlobalsLen() int {
	return vm.globals.Len()
}

// GlobalAt returns the value at the given global index, or nil if out of range.
func (vm *VM) GlobalAt(idx int) object.Object {
	if idx < 0 || idx >= vm.globals.Len() {
		return nil
	}
	return vm.globals.Get(idx)
}
