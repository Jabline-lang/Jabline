package evaluator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"jabline/pkg/ast"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
)

// ModuleSystem manages module loading and caching
type ModuleSystem struct {
	cache map[string]*object.Environment
	paths []string
}

// NewModuleSystem creates a new module system
func NewModuleSystem() *ModuleSystem {
	return &ModuleSystem{
		cache: make(map[string]*object.Environment),
		paths: []string{
			"./",
			"./modules/",
			"./lib/",
		},
	}
}

// Global module system instance
var moduleSystem = NewModuleSystem()

// evalImportStatement evaluates an import statement
func evalImportStatement(is *ast.ImportStatement, env *object.Environment) object.Object {
	if is == nil || is.ModuleName == nil {
		return newError("invalid import statement")
	}

	moduleName := is.ModuleName.Value

	// Load the module
	moduleEnv, err := moduleSystem.LoadModule(moduleName)
	if err != nil {
		return newError("failed to load module '%s': %s", moduleName, err.Error())
	}

	// If it's a selective import, only import specific items
	if len(is.ImportList) > 0 {
		for _, item := range is.ImportList {
			value, exists := moduleEnv.Get(item.Value)
			if !exists {
				return newError("'%s' is not exported by module '%s'", item.Value, moduleName)
			}
			env.Set(item.Value, value)
		}
	} else {
		// Import the entire module
		if is.Alias != nil {
			// Import as alias (import math as m)
			moduleObj := &object.Hash{
				Pairs: make(map[object.HashKey]object.HashPair),
			}

			// Convert environment to hash object
			for name, value := range moduleEnv.GetStore() {
				key := &object.String{Value: name}
				hashKey := key.HashKey()
				moduleObj.Pairs[hashKey] = object.HashPair{
					Key:   key,
					Value: value,
				}
			}

			env.Set(is.Alias.Value, moduleObj)
		} else {
			// Import all exports directly into current scope
			for name, value := range moduleEnv.GetStore() {
				env.Set(name, value)
			}
		}
	}

	return NULL
}

// evalExportStatement evaluates an export statement
func evalExportStatement(es *ast.ExportStatement, env *object.Environment) object.Object {
	if es == nil || es.Statement == nil {
		return newError("invalid export statement")
	}

	// Evaluate the statement normally
	result := Eval(es.Statement, env)
	if isError(result) {
		return result
	}

	// Mark the exported item in the environment
	// For now, we'll just evaluate it normally since the module system
	// will capture all variables in the module's environment

	return result
}

// LoadModule loads a module from the file system
func (ms *ModuleSystem) LoadModule(moduleName string) (*object.Environment, error) {
	// Check cache first
	if cachedEnv, exists := ms.cache[moduleName]; exists {
		return cachedEnv, nil
	}

	// Find the module file
	modulePath, err := ms.findModule(moduleName)
	if err != nil {
		return nil, err
	}

	// Read the module file
	content, err := ioutil.ReadFile(modulePath)
	if err != nil {
		return nil, fmt.Errorf("could not read module file: %s", err)
	}

	// Parse and evaluate the module
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in module: %v", p.Errors())
	}

	// Create a new environment for the module with built-ins
	moduleEnv := newModuleEnvironment()

	// Evaluate the module in its own environment
	result := Eval(program, moduleEnv)
	if isError(result) {
		return nil, fmt.Errorf("runtime error in module: %s", result.(*object.Error).Message)
	}

	// Cache the module environment
	ms.cache[moduleName] = moduleEnv

	return moduleEnv, nil
}

// findModule searches for a module file in the module paths
func (ms *ModuleSystem) findModule(moduleName string) (string, error) {
	// Add .jb extension if not present
	if !strings.HasSuffix(moduleName, ".jb") {
		moduleName += ".jb"
	}

	// Search in module paths
	for _, path := range ms.paths {
		fullPath := filepath.Join(path, moduleName)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("module '%s' not found in paths: %v", moduleName, ms.paths)
}

// AddModulePath adds a new path to search for modules
func (ms *ModuleSystem) AddModulePath(path string) {
	ms.paths = append(ms.paths, path)
}

// ClearCache clears the module cache
func (ms *ModuleSystem) ClearCache() {
	ms.cache = make(map[string]*object.Environment)
}

// GetLoadedModules returns the names of all loaded modules
func (ms *ModuleSystem) GetLoadedModules() []string {
	var modules []string
	for name := range ms.cache {
		modules = append(modules, name)
	}
	return modules
}

// newModuleEnvironment creates a new environment with built-in functions for modules
func newModuleEnvironment() *object.Environment {
	env := object.NewEnvironment()

	// Add all built-in functions to the module environment
	for name, builtin := range builtins {
		env.Set(name, builtin)
	}

	return env
}
