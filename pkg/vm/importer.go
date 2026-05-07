package vm

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"jabline/pkg/stdlib"
	"jabline/pkg/symbol"
)

type ModuleLoader struct {
	cache   map[string]*object.Hash
	paths   []string
	embedFS fs.FS // Optional embedded filesystem
}

func NewModuleLoader() *ModuleLoader {
	return NewModuleLoaderWithEmbed(nil)
}

func NewModuleLoaderWithEmbed(embedFS fs.FS) *ModuleLoader {
	cwd, _ := os.Getwd()

	paths := []string{
		cwd,
		filepath.Join(cwd, "lib"),
		filepath.Join(cwd, "internal", "embedded", "modules"),
	}

	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		paths = append(paths, filepath.Join(exeDir, "modules"))
		paths = append(paths, exeDir)
	}

	return &ModuleLoader{
		cache:   make(map[string]*object.Hash),
		paths:   paths,
		embedFS: embedFS,
	}
}

func (ml *ModuleLoader) Load(name string) (*object.Hash, error) {
	originalName := name
	name = strings.TrimPrefix(name, "std/")

	// Check for native modules (e.g. "_strings" -> Go builtin)
	if strings.HasPrefix(name, "_") {
		if nativeMod := stdlib.GetNativeModule(name); nativeMod != nil {
			return nativeMod, nil
		}
	}

	// Determine the file path key for caching
	cacheKey := name
	if filepath.Ext(cacheKey) == "" {
		cacheKey += ".jb"
	}

	if module, ok := ml.cache[cacheKey]; ok {
		// Circular dependency check
		if module == nil {
			return nil, fmt.Errorf("import error: circular dependency detected for module '%s'", name)
		}
		return module, nil
	}

	// Mark as resolving to detect circular dependencies
	ml.cache[cacheKey] = nil

	// 1. Try embedded FS first (bundled into the binary)
	if ml.embedFS != nil {
		content, err := fs.ReadFile(ml.embedFS, filepath.ToSlash(cacheKey))
		if err == nil {
			// Use a trusted filename so embedded stdlib modules can import _private natives.
			// ops_module.go allows imports from paths containing "internal/embedded".
			trustedName := "internal/embedded/modules/" + filepath.ToSlash(cacheKey)
			res, compileErr := ml.compileContent(string(content), originalName, name, trustedName)
			if compileErr != nil {
				delete(ml.cache, cacheKey)
			}
			return res, compileErr
		}
	}

	// 2. Fallback to OS filesystem
	absPath, err := ml.resolvePath(name)
	if err != nil {
		delete(ml.cache, cacheKey)
		return nil, err
	}

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		delete(ml.cache, cacheKey)
		return nil, fmt.Errorf("failed to read module '%s': %s", absPath, err)
	}

	res, compileErr := ml.compileContent(string(content), originalName, name, absPath)
	if compileErr != nil {
		delete(ml.cache, cacheKey)
	}
	return res, compileErr
}

func (ml *ModuleLoader) compileContent(source, originalName, name, cacheKey string) (*object.Hash, error) {
	l := lexer.New(source)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in module '%s': %v", originalName, p.Errors())
	}

	comp := compiler.New()
	err := comp.Compile(prog)
	if err != nil {
		return nil, fmt.Errorf("compilation error in module '%s': %s", originalName, err)
	}

	bytecode := comp.Bytecode()

	// Use cacheKey as filename for the module VM.
	// For embedded stdlib modules, cacheKey is "internal/embedded/modules/..." which
	// allows the security check in ops_module.go to permit importing _private natives.
	moduleVM := NewWithLoader(bytecode.Instructions, bytecode.Constants, cacheKey, ml)

	err = moduleVM.Run()
	if err != nil {
		return nil, fmt.Errorf("runtime error in module '%s': %s", name, err)
	}

	exports := make(map[object.HashKey]object.HashPair)
	for symName, sym := range bytecode.SymbolTable.GetStore() {
		if sym.Scope == symbol.GlobalScope && sym.IsExported {
			if sym.Index < len(moduleVM.globals) {
				val := moduleVM.globals[sym.Index]
				if val != nil {
					key := &object.String{Value: symName}
					exports[key.HashKey()] = object.HashPair{Key: key, Value: val}
				}
			}
		}
	}

	moduleHash := &object.Hash{Pairs: exports}
	ml.cache[cacheKey] = moduleHash

	return moduleHash, nil
}

func (ml *ModuleLoader) resolvePath(name string) (string, error) {
	filename := name
	if filepath.Ext(filename) == "" {
		filename += ".jb"
	}

	if filepath.IsAbs(filename) || strings.HasPrefix(filename, ".") {
		abs, err := filepath.Abs(filename)
		if err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs, nil
			}
		}
		return "", fmt.Errorf("module not found at '%s'", filename)
	}

	for _, path := range ml.paths {
		fullPath := filepath.Join(path, filename)
		if info, err := os.Stat(fullPath); err == nil {
			if info.IsDir() {
				// Try dir/main.jb
				mainPath := filepath.Join(fullPath, "main.jb")
				if _, err := os.Stat(mainPath); err == nil {
					return filepath.Abs(mainPath)
				}
			} else {
				return filepath.Abs(fullPath)
			}
		}

		// Also try without .jb extension if it was added automatically and it's a directory
		if strings.HasSuffix(filename, ".jb") {
			nameOnly := strings.TrimSuffix(filename, ".jb")
			dirPath := filepath.Join(path, nameOnly)
			if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
				mainPath := filepath.Join(dirPath, "main.jb")
				if _, err := os.Stat(mainPath); err == nil {
					return filepath.Abs(mainPath)
				}
			}
		}
	}

	return "", fmt.Errorf("module '%s' not found in paths %v", name, ml.paths)
}
