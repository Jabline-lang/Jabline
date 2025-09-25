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
	cache   map[string]*object.Environment
	paths   []string
	loading map[string]bool // track modules being loaded to prevent circular imports
}

// NewModuleSystem creates a new module system with enhanced caching
func NewModuleSystem() *ModuleSystem {
	return &ModuleSystem{
		cache: make(map[string]*object.Environment),
		paths: []string{
			"./",
			"./modules/",
			"./lib/",
			"./node_modules/", // npm-style modules
		},
	}
}

// Global module system instance
var moduleSystem = NewModuleSystem()

// evalImportStatement evaluates an import statement with full syntax support
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

	switch is.ImportType {
	case ast.IMPORT_SIDE_EFFECT:
		// import "module" - just load for side effects
		return NULL

	case ast.IMPORT_DEFAULT:
		// import defaultExport from "module"
		defaultValue, exists := moduleEnv.Get("default")
		if !exists {
			return newError("module '%s' has no default export", moduleName)
		}
		env.Set(is.DefaultImport.Value, defaultValue)

	case ast.IMPORT_NAMED:
		// import { name1, name2 as alias } from "module"
		for _, item := range is.NamedImports {
			originalName := item.Name.Value
			importName := originalName

			// Use alias if provided
			if item.Alias != nil {
				importName = item.Alias.Value
			}

			value, exists := moduleEnv.Get(originalName)
			if !exists {
				return newError("'%s' is not exported by module '%s'", originalName, moduleName)
			}

			env.Set(importName, value)
		}

	case ast.IMPORT_NAMESPACE:
		// import * as namespace from "module"
		moduleObj := &object.Hash{
			Pairs: make(map[object.HashKey]object.HashPair),
		}

		// Convert all exports to hash object
		for name, value := range moduleEnv.GetStore() {
			// Skip built-ins and private variables
			if isBuiltinFunction(name) || name[0] == '_' {
				continue
			}

			key := &object.String{Value: name}
			hashKey := key.HashKey()
			moduleObj.Pairs[hashKey] = object.HashPair{
				Key:   key,
				Value: value,
			}
		}

		env.Set(is.NamespaceAlias.Value, moduleObj)

	case ast.IMPORT_MIXED:
		// import defaultExport, { name1, name2 } from "module"
		// Handle default import
		if is.DefaultImport != nil {
			defaultValue, exists := moduleEnv.Get("default")
			if !exists {
				return newError("module '%s' has no default export", moduleName)
			}
			env.Set(is.DefaultImport.Value, defaultValue)
		}

		// Handle named imports
		for _, item := range is.NamedImports {
			originalName := item.Name.Value
			importName := originalName

			if item.Alias != nil {
				importName = item.Alias.Value
			}

			value, exists := moduleEnv.Get(originalName)
			if !exists {
				return newError("'%s' is not exported by module '%s'", originalName, moduleName)
			}

			env.Set(importName, value)
		}
	}

	return NULL
}

// evalExportStatement evaluates an export statement with full syntax support
func evalExportStatement(es *ast.ExportStatement, env *object.Environment) object.Object {
	if es == nil {
		return newError("invalid export statement")
	}

	switch es.ExportType {
	case ast.EXPORT_DECLARATION:
		// export let/const/fn/etc.
		if es.Statement == nil {
			return newError("export declaration missing statement")
		}

		result := Eval(es.Statement, env)
		if isError(result) {
			return result
		}

		// If it's a default export, store under "default"
		if es.IsDefault {
			env.Set("default", result)
		}

		return result

	case ast.EXPORT_DEFAULT:
		// export default expression
		if es.Statement == nil {
			return newError("export default missing expression")
		}

		result := Eval(es.Statement, env)
		if isError(result) {
			return result
		}

		env.Set("default", result)
		return result

	case ast.EXPORT_LIST:
		// export { name1, name2 as alias }
		for _, item := range es.ExportList {
			originalName := item.Name.Value
			exportName := originalName

			if item.Alias != nil {
				exportName = item.Alias.Value
			}

			value, exists := env.Get(originalName)
			if !exists {
				return newError("'%s' is not defined", originalName)
			}

			// Create new binding with export name if different
			if exportName != originalName {
				env.Set(exportName, value)
			}
		}

		return NULL

	case ast.EXPORT_ALL:
		// export * from "module"
		if es.ModuleName == nil {
			return newError("export * missing module name")
		}

		reExportedModule, err := moduleSystem.LoadModule(es.ModuleName.Value)
		if err != nil {
			return newError("failed to load module '%s' for re-export: %s", es.ModuleName.Value, err.Error())
		}

		// Re-export all non-default exports
		for name, value := range reExportedModule.GetStore() {
			if name != "default" && !isBuiltinFunction(name) && name[0] != '_' {
				env.Set(name, value)
			}
		}

		return NULL

	case ast.EXPORT_NAMED_FROM:
		// export { name1, name2 } from "module"
		if es.ModuleName == nil {
			return newError("export from missing module name")
		}

		reExportedModule, err := moduleSystem.LoadModule(es.ModuleName.Value)
		if err != nil {
			return newError("failed to load module '%s' for re-export: %s", es.ModuleName.Value, err.Error())
		}

		for _, item := range es.ExportList {
			originalName := item.Name.Value
			exportName := originalName

			if item.Alias != nil {
				exportName = item.Alias.Value
			}

			value, exists := reExportedModule.Get(originalName)
			if !exists {
				return newError("'%s' is not exported by module '%s'", originalName, es.ModuleName.Value)
			}

			env.Set(exportName, value)
		}

		return NULL

	case ast.EXPORT_ALL_AS:
		// export * as namespace from "module"
		if es.ModuleName == nil || es.NamespaceAlias == nil {
			return newError("export * as namespace missing module name or alias")
		}

		reExportedModule, err := moduleSystem.LoadModule(es.ModuleName.Value)
		if err != nil {
			return newError("failed to load module '%s' for namespace re-export: %s", es.ModuleName.Value, err.Error())
		}

		// Create namespace object with all exports
		namespaceObj := &object.Hash{
			Pairs: make(map[object.HashKey]object.HashPair),
		}

		for name, value := range reExportedModule.GetStore() {
			if !isBuiltinFunction(name) && name[0] != '_' {
				key := &object.String{Value: name}
				hashKey := key.HashKey()
				namespaceObj.Pairs[hashKey] = object.HashPair{
					Key:   key,
					Value: value,
				}
			}
		}

		env.Set(es.NamespaceAlias.Value, namespaceObj)
		return NULL
	}

	return newError("unknown export type")
}

// LoadModule loads a module from the file system with circular import detection
func (ms *ModuleSystem) LoadModule(moduleName string) (*object.Environment, error) {
	// Resolve full path
	modulePath, err := ms.ResolveModulePath(moduleName)
	if err != nil {
		return nil, err
	}

	// Check cache first using resolved path
	if cachedEnv, exists := ms.cache[modulePath]; exists {
		return cachedEnv, nil
	}

	// Check for circular imports
	if ms.isLoading(modulePath) {
		return nil, fmt.Errorf("circular import detected for module '%s'", moduleName)
	}

	// Mark as loading
	ms.setLoading(modulePath, true)
	defer ms.setLoading(modulePath, false)

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

	// Store current working directory for relative imports
	oldWd, _ := os.Getwd()
	moduleDir := filepath.Dir(modulePath)
	if moduleDir != "" && moduleDir != "." {
		os.Chdir(moduleDir)
		defer os.Chdir(oldWd)
	}

	// Evaluate the module in its own environment
	result := Eval(program, moduleEnv)
	if isError(result) {
		return nil, fmt.Errorf("runtime error in module: %s", result.(*object.Error).Message)
	}

	// Cache the module environment using resolved path
	ms.cache[modulePath] = moduleEnv

	return moduleEnv, nil
}

// findModule searches for a module file in the module paths with multiple extensions
func (ms *ModuleSystem) findModule(moduleName string) (string, error) {
	extensions := []string{".jb", ".jabline", ""}

	// Try each extension
	for _, ext := range extensions {
		testName := moduleName
		if ext != "" && !strings.HasSuffix(moduleName, ext) {
			testName = moduleName + ext
		}

		// Search in module paths
		for _, path := range ms.paths {
			fullPath := filepath.Join(path, testName)
			if _, err := os.Stat(fullPath); err == nil {
				absPath, absErr := filepath.Abs(fullPath)
				if absErr != nil {
					return fullPath, nil
				}
				return absPath, nil
			}
		}

		// Also try as directory with index.jb
		if ext == ".jb" {
			for _, path := range ms.paths {
				dirPath := filepath.Join(path, moduleName)
				indexPath := filepath.Join(dirPath, "index.jb")
				if _, err := os.Stat(indexPath); err == nil {
					absPath, absErr := filepath.Abs(indexPath)
					if absErr != nil {
						return indexPath, nil
					}
					return absPath, nil
				}
			}
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
	ms.loading = make(map[string]bool)
}

// isLoading checks if a module is currently being loaded
func (ms *ModuleSystem) isLoading(path string) bool {
	if ms.loading == nil {
		ms.loading = make(map[string]bool)
	}
	return ms.loading[path]
}

// setLoading sets the loading state for a module
func (ms *ModuleSystem) setLoading(path string, loading bool) {
	if ms.loading == nil {
		ms.loading = make(map[string]bool)
	}
	if loading {
		ms.loading[path] = true
	} else {
		delete(ms.loading, path)
	}
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

// isBuiltinFunction checks if a name is a built-in function
func isBuiltinFunction(name string) bool {
	_, exists := builtins[name]
	return exists
}

// GetExportedValues returns only the exported values from a module environment
func GetExportedValues(env *object.Environment) map[string]object.Object {
	exported := make(map[string]object.Object)

	for name, value := range env.GetStore() {
		// Skip built-ins and private variables (starting with _)
		if !isBuiltinFunction(name) && name[0] != '_' {
			exported[name] = value
		}
	}

	return exported
}

// ResolveModulePath resolves a module path based on different resolution strategies
func (ms *ModuleSystem) ResolveModulePath(moduleName string) (string, error) {
	// 1. Try relative path first if it starts with ./ or ../
	if strings.HasPrefix(moduleName, "./") || strings.HasPrefix(moduleName, "../") {
		return ms.resolveRelativePath(moduleName)
	}

	// 2. Try as absolute path with extensions
	return ms.findModule(moduleName)
}

// resolveRelativePath resolves relative module paths
func (ms *ModuleSystem) resolveRelativePath(moduleName string) (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current working directory: %s", err)
	}

	// Extensions to try
	extensions := []string{".jb", ""}

	for _, ext := range extensions {
		testName := moduleName
		if ext != "" && !strings.HasSuffix(moduleName, ext) {
			testName = moduleName + ext
		}

		// Build absolute path
		absolutePath := filepath.Join(cwd, testName)

		// Check if file exists
		if _, err := os.Stat(absolutePath); err == nil {
			return absolutePath, nil
		}
	}

	return "", fmt.Errorf("relative module '%s' not found", moduleName)
}

// ModuleCache represents the module caching system
type ModuleCache struct {
	modules map[string]*object.Environment
	loading map[string]bool // prevent circular imports
}

// NewModuleCache creates a new module cache
func NewModuleCache() *ModuleCache {
	return &ModuleCache{
		modules: make(map[string]*object.Environment),
		loading: make(map[string]bool),
	}
}

// GetModule gets a cached module or nil if not found
func (mc *ModuleCache) GetModule(path string) *object.Environment {
	return mc.modules[path]
}

// SetModule caches a module
func (mc *ModuleCache) SetModule(path string, env *object.Environment) {
	mc.modules[path] = env
	delete(mc.loading, path) // remove from loading set
}

// IsLoading checks if a module is currently being loaded
func (mc *ModuleCache) IsLoading(path string) bool {
	return mc.loading[path]
}

// SetLoading marks a module as being loaded
func (mc *ModuleCache) SetLoading(path string) {
	mc.loading[path] = true
}

// ClearLoading removes a module from the loading set
func (mc *ModuleCache) ClearLoading(path string) {
	delete(mc.loading, path)
}
