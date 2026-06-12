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
	"_template": "template_",
	"_websocket": "ws",
	"_env":       "env",
	"_db":        "db_",
	"_regex":     "regex_",
}

func init() {
	NativeModuleRegistry["_math"] = MathBuiltins
	NativeModuleRegistry["_os"] = OSBuiltins
	NativeModuleRegistry["_io"] = IOBuiltins
	NativeModuleRegistry["_fs"] = IOBuiltins
	NativeModuleRegistry["_encoding"] = EncodingBuiltins
	NativeModuleRegistry["_json"] = JSONBuiltins
	NativeModuleRegistry["_strings"] = StringBuiltins
	NativeModuleRegistry["_time"] = TimeBuiltins
	NativeModuleRegistry["_types"] = TypesBuiltins
	NativeModuleRegistry["_runtime"] = RuntimeBuiltins
	NativeModuleRegistry["_env"] = EnvBuiltins
	NativeModuleRegistry["_template"] = TemplateBuiltins
	NativeModuleRegistry["_regex"] = RegexBuiltins
	NativeModuleRegistry["_csv"] = CSVBuiltins
	NativeModuleRegistry["_datetime"] = DateTimeBuiltins
	NativeModuleRegistry["_yaml"] = YAMLBuiltins
	NativeModuleRegistry["_compress"] = CompressBuiltins
	NativeModuleRegistry["_tls"] = TLSBuiltins
	NativeModuleRegistry["_sync"] = SyncBuiltins
	NativeModuleRegistry["_xml"] = XMLBuiltins

	// Now that NativeModuleRegistry is populated, build GlobalModules
	nativeModules := []string{"_strings", "_math", "_json", "_os", "_io", "_fs", "_http", "_db", "_template", "_websocket", "_regex", "_csv", "_time", "_datetime", "_yaml", "_compress", "_tls", "_sync", "_xml", "_config", "_health", "_log", "_metrics", "_parallel", "_context", "_resilience", "_http_advanced"}
	for _, modName := range nativeModules {
		if modHash := GetNativeModule(modName); modHash != nil {
			globalName := modName[1:]
			GlobalModules[globalName] = modHash
		}
	}

	// Register Global Modules in the global Registry
	for name, obj := range GlobalModules {
		Registry = append(Registry, struct {
			Name   string
			Object object.Object
		}{name, obj})
	}
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
