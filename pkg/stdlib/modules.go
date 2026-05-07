package stdlib

import (
	"jabline/pkg/object"
	"strings"
)

// Executor is injected by the VM to allow running closures from stdlib.
// Defined here so it's always available regardless of build tags.
var Executor object.VMExecutor

// NativeModuleRegistry maps module names (e.g. "_db") to their builtins.
// Modules register themselves via their own init() functions based on build tags,
// enabling Tree-Shaking: only modules actually imported are linked into the binary.
var NativeModuleRegistry = make(map[string][]struct {
	Name   string
	Object object.Object
})

// NativeModulePrefixes maps module names to the prefix they use for their functions
var NativeModulePrefixes = map[string]string{
	"_math":      "math_",
	"_os":        "os_",
	"_io":        "io_",
	"_fs":        "io_",
	"_encoding":  "encoding_",
	"_json":      "json_",
	"_http":      "http_",
	"_strings":   "strings_",
	"_crypto":    "crypto_",
	"_time":      "time_",
	"_types":     "to_",
	"_runtime":   "runtime",
	"_websocket": "ws",
	"_env":       "env",
	"_db":        "db_",
}

func init() {
	NativeModuleRegistry["_math"] = MathBuiltins
	NativeModuleRegistry["_os"] = OSBuiltins
	NativeModuleRegistry["_io"] = IOBuiltins
	NativeModuleRegistry["_fs"] = IOBuiltins
	NativeModuleRegistry["_encoding"] = EncodingBuiltins
	NativeModuleRegistry["_json"] = JSONBuiltins
	NativeModuleRegistry["_strings"] = StringBuiltins
	NativeModuleRegistry["_crypto"] = CryptoBuiltins
	NativeModuleRegistry["_time"] = TimeBuiltins
	NativeModuleRegistry["_types"] = TypesBuiltins
	NativeModuleRegistry["_runtime"] = RuntimeBuiltins
	NativeModuleRegistry["_env"] = EnvBuiltins
}

// GetNativeModule returns a Hash object containing the builtins for a given module name.
// It returns nil if the module is not found or was not linked (Tree-Shaking).
func GetNativeModule(name string) *object.Hash {
	builtins, exists := NativeModuleRegistry[name]
	if !exists {
		return nil
	}

	prefix := NativeModulePrefixes[name]

	pairs := make(map[object.HashKey]object.HashPair)
	for _, b := range builtins {
		// Clean up names using the explicit prefix
		cleanName := strings.TrimPrefix(b.Name, prefix)

		key := &object.String{Value: cleanName}
		pairs[key.HashKey()] = object.HashPair{Key: key, Value: b.Object}
	}

	return &object.Hash{Pairs: pairs}
}
