package stdlib

import (
	"jabline/pkg/object"
	"strings"
)

// GetNativeModule returns a Hash object containing the builtins for a given module name.
// It returns nil if the module is not found.
func GetNativeModule(name string) *object.Hash {
	var builtins []struct {
		Name   string
		Object object.Object
	}
	var prefix string

	switch name {
	case "_math":
		builtins = MathBuiltins
		prefix = "math_"
	case "_os":
		builtins = OSBuiltins
		prefix = "os_"
	case "_io":
		builtins = IOBuiltins
		prefix = "io_"
	case "_fs":
		builtins = IOBuiltins
		prefix = "io_"
	case "_encoding":
		builtins = EncodingBuiltins
		prefix = "encoding_"
	case "_json":
		builtins = JSONBuiltins
		prefix = "json_"
	case "_http":
		builtins = HTTPBuiltins
		prefix = "http_"
	case "_strings":
		builtins = StringBuiltins
		prefix = "strings_"
	case "_crypto":
		builtins = CryptoBuiltins
		prefix = "crypto_"
	case "_time":
		builtins = TimeBuiltins
		prefix = "time_"
	case "_types":
		builtins = TypesBuiltins
		prefix = "to_" // Functions are named toInt8, toUint32, etc.
	default:
		return nil
	}

	pairs := make(map[object.HashKey]object.HashPair)
	for _, b := range builtins {
		// Clean up names using the explicit prefix
		cleanName := b.Name
		cleanName = strings.TrimPrefix(b.Name, prefix)

		key := &object.String{Value: cleanName}
		pairs[key.HashKey()] = object.HashPair{Key: key, Value: b.Object}
	}

	return &object.Hash{Pairs: pairs}
}
