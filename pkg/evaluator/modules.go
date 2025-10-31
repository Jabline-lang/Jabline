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

type ModuleSystem struct {
	cache   map[string]*object.Environment
	paths   []string
	loading map[string]bool
}

func NewModuleSystem() *ModuleSystem {
	return &ModuleSystem{
		cache: make(map[string]*object.Environment),
		paths: []string{
			"./",
			"./modules/",
			"./lib/",
			"./node_modules/",
		},
	}
}

var moduleSystem = NewModuleSystem()

func evalImportStatement(is *ast.ImportStatement, env *object.Environment) object.Object {
	if is == nil || is.ModuleName == nil {
		return newError("invalid import statement")
	}

	moduleName := is.ModuleName.Value

	moduleEnv, err := moduleSystem.LoadModule(moduleName)
	if err != nil {
		return newError("failed to load module '%s': %s", moduleName, err.Error())
	}

	switch is.ImportType {
	case ast.IMPORT_SIDE_EFFECT:
		return NULL

	case ast.IMPORT_DEFAULT:
		defaultValue, exists := moduleEnv.Get("default")
		if !exists {
			return newError("module '%s' has no default export", moduleName)
		}
		env.Set(is.DefaultImport.Value, defaultValue)

	case ast.IMPORT_NAMED:
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

	case ast.IMPORT_NAMESPACE:
		moduleObj := &object.Hash{
			Pairs: make(map[object.HashKey]object.HashPair),
		}

		for name, value := range moduleEnv.GetStore() {
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
		if is.DefaultImport != nil {
			defaultValue, exists := moduleEnv.Get("default")
			if !exists {
				return newError("module '%s' has no default export", moduleName)
			}
			env.Set(is.DefaultImport.Value, defaultValue)
		}

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

func evalExportStatement(es *ast.ExportStatement, env *object.Environment) object.Object {
	if es == nil {
		return newError("invalid export statement")
	}

	switch es.ExportType {
	case ast.EXPORT_DECLARATION:
		if es.Statement == nil {
			return newError("export declaration missing statement")
		}

		result := Eval(es.Statement, env)
		if isError(result) {
			return result
		}

		if es.IsDefault {
			env.Set("default", result)
		}

		return result

	case ast.EXPORT_DEFAULT:
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

			if exportName != originalName {
				env.Set(exportName, value)
			}
		}

		return NULL

	case ast.EXPORT_ALL:
		if es.ModuleName == nil {
			return newError("export * missing module name")
		}

		reExportedModule, err := moduleSystem.LoadModule(es.ModuleName.Value)
		if err != nil {
			return newError("failed to load module '%s' for re-export: %s", es.ModuleName.Value, err.Error())
		}

		for name, value := range reExportedModule.GetStore() {
			if name != "default" && !isBuiltinFunction(name) && name[0] != '_' {
				env.Set(name, value)
			}
		}

		return NULL

	case ast.EXPORT_NAMED_FROM:
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
		if es.ModuleName == nil || es.NamespaceAlias == nil {
			return newError("export * as namespace missing module name or alias")
		}

		reExportedModule, err := moduleSystem.LoadModule(es.ModuleName.Value)
		if err != nil {
			return newError("failed to load module '%s' for namespace re-export: %s", es.ModuleName.Value, err.Error())
		}

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

func (ms *ModuleSystem) LoadModule(moduleName string) (*object.Environment, error) {
	modulePath, err := ms.ResolveModulePath(moduleName)
	if err != nil {
		return nil, err
	}

	if cachedEnv, exists := ms.cache[modulePath]; exists {
		return cachedEnv, nil
	}

	if ms.isLoading(modulePath) {
		return nil, fmt.Errorf("circular import detected for module '%s'", moduleName)
	}

	ms.setLoading(modulePath, true)
	defer ms.setLoading(modulePath, false)

	content, err := ioutil.ReadFile(modulePath)
	if err != nil {
		return nil, fmt.Errorf("could not read module file: %s", err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in module: %v", p.Errors())
	}

	moduleEnv := newModuleEnvironment()

	oldWd, _ := os.Getwd()
	moduleDir := filepath.Dir(modulePath)
	if moduleDir != "" && moduleDir != "." {
		os.Chdir(moduleDir)
		defer os.Chdir(oldWd)
	}

	result := Eval(program, moduleEnv)
	if isError(result) {
		return nil, fmt.Errorf("runtime error in module: %s", result.(*object.Error).Message)
	}

	ms.cache[modulePath] = moduleEnv

	return moduleEnv, nil
}

func (ms *ModuleSystem) findModule(moduleName string) (string, error) {
	extensions := []string{".jb", ".jabline", ""}

	for _, ext := range extensions {
		testName := moduleName
		if ext != "" && !strings.HasSuffix(moduleName, ext) {
			testName = moduleName + ext
		}

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

func (ms *ModuleSystem) AddModulePath(path string) {
	ms.paths = append(ms.paths, path)
}

func (ms *ModuleSystem) ClearCache() {
	ms.cache = make(map[string]*object.Environment)
	ms.loading = make(map[string]bool)
}

func (ms *ModuleSystem) isLoading(path string) bool {
	if ms.loading == nil {
		ms.loading = make(map[string]bool)
	}
	return ms.loading[path]
}

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

func (ms *ModuleSystem) GetLoadedModules() []string {
	var modules []string
	for name := range ms.cache {
		modules = append(modules, name)
	}
	return modules
}

func newModuleEnvironment() *object.Environment {
	env := object.NewEnvironment()

	for name, builtin := range builtins {
		env.Set(name, builtin)
	}

	return env
}

func isBuiltinFunction(name string) bool {
	_, exists := builtins[name]
	return exists
}

func GetExportedValues(env *object.Environment) map[string]object.Object {
	exported := make(map[string]object.Object)

	for name, value := range env.GetStore() {
		if !isBuiltinFunction(name) && name[0] != '_' {
			exported[name] = value
		}
	}

	return exported
}

func (ms *ModuleSystem) ResolveModulePath(moduleName string) (string, error) {
	if strings.HasPrefix(moduleName, "./") || strings.HasPrefix(moduleName, "../") {
		return ms.resolveRelativePath(moduleName)
	}

	return ms.findModule(moduleName)
}

func (ms *ModuleSystem) resolveRelativePath(moduleName string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current working directory: %s", err)
	}

	extensions := []string{".jb", ""}

	for _, ext := range extensions {
		testName := moduleName
		if ext != "" && !strings.HasSuffix(moduleName, ext) {
			testName = moduleName + ext
		}

		absolutePath := filepath.Join(cwd, testName)

		if _, err := os.Stat(absolutePath); err == nil {
			return absolutePath, nil
		}
	}

	return "", fmt.Errorf("relative module '%s' not found", moduleName)
}

type ModuleCache struct {
	modules map[string]*object.Environment
	loading map[string]bool
}

func NewModuleCache() *ModuleCache {
	return &ModuleCache{
		modules: make(map[string]*object.Environment),
		loading: make(map[string]bool),
	}
}

func (mc *ModuleCache) GetModule(path string) *object.Environment {
	return mc.modules[path]
}

func (mc *ModuleCache) SetModule(path string, env *object.Environment) {
	mc.modules[path] = env
	delete(mc.loading, path)
}

func (mc *ModuleCache) IsLoading(path string) bool {
	return mc.loading[path]
}

func (mc *ModuleCache) SetLoading(path string) {
	mc.loading[path] = true
}

func (mc *ModuleCache) ClearLoading(path string) {
	delete(mc.loading, path)
}
