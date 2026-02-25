package vm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"jabline/pkg/compiler"
	"jabline/pkg/lexer"
	"jabline/pkg/object"
	"jabline/pkg/parser"
	"jabline/pkg/stdlib" // Add this import
	"jabline/pkg/symbol"
)

type ModuleLoader struct {
	cache map[string]*object.Hash
	paths []string
}

func NewModuleLoader() *ModuleLoader {
	cwd, _ := os.Getwd()
	return &ModuleLoader{
		cache: make(map[string]*object.Hash),
		paths: []string{
			cwd,
			filepath.Join(cwd, "modules"),
			filepath.Join(cwd, "lib"),
		},
	}
}

func (ml *ModuleLoader) Load(name string) (*object.Hash, error) {
	// Check for native modules mapping (e.g. "modules/strings" -> "_strings")
	var nativeName string
	if strings.HasPrefix(name, "modules/") {
		nativeName = "_" + strings.TrimPrefix(name, "modules/")
	} else if strings.HasPrefix(name, "_") {
		// Allow direct access to "_" prefixed native modules if needed
		nativeName = name
	}

	if nativeName != "" {
		if nativeMod := stdlib.GetNativeModule(nativeName); nativeMod != nil {
			return nativeMod, nil
		}
		// If it has modules/ prefix but not found in stdlib, it might be an error or a physical folder
		// but typically we expect them in stdlib.
	}

	absPath, err := ml.resolvePath(name)
	if err != nil {
		return nil, err
	}

	if module, ok := ml.cache[absPath]; ok {
		return module, nil
	}

	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module '%s': %s", absPath, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in module '%s': %v", name, p.Errors())
	}

	comp := compiler.New()
	err = comp.Compile(prog)
	if err != nil {
		return nil, fmt.Errorf("compilation error in module '%s': %s", name, err)
	}

	bytecode := comp.Bytecode()

	moduleVM := NewWithLoader(bytecode.Instructions, bytecode.Constants, absPath, ml)

	err = moduleVM.Run()
	if err != nil {
		return nil, fmt.Errorf("runtime error in module '%s': %s", name, err)
	}

	exports := make(map[object.HashKey]object.HashPair)
	for name, sym := range bytecode.SymbolTable.GetStore() { // Renamed 'symbol' to 'sym'
		if sym.Scope == symbol.GlobalScope && sym.IsExported { // Only export marked symbols
			if sym.Index < len(moduleVM.globals) {
				val := moduleVM.globals[sym.Index]
				if val != nil {
					key := &object.String{Value: name}
					exports[key.HashKey()] = object.HashPair{Key: key, Value: val}
				}
			}
		}
	}

	moduleHash := &object.Hash{Pairs: exports}

	ml.cache[absPath] = moduleHash

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
		if _, err := os.Stat(fullPath); err == nil {
			return filepath.Abs(fullPath)
		}
	}

	return "", fmt.Errorf("module '%s' not found in paths %v", name, ml.paths)
}
