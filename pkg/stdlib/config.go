package stdlib

import (
	"encoding/json"
	"fmt"
	"jabline/pkg/object"
	"os"
	"strings"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

var ConfigBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"config_load", &object.Builtin{Fn: configLoad}},
	{"config_get", &object.Builtin{Fn: configGet}},
	{"config_set", &object.Builtin{Fn: configSet}},
	{"config_env", &object.Builtin{Fn: configEnv}},
	{"config_file", &object.Builtin{Fn: configFile}},
}

func init() {
	NativeModuleRegistry["_config"] = ConfigBuiltins
	NativeModulePrefixes["_config"] = "config_"
}

// configLoad loads config from args: path strings or hash maps.
// config.load("app.json") or config.load({port: 8080})
func configLoad(args ...object.Object) object.Object {
	if len(args) == 0 {
		return newError("config_load expects at least 1 arg")
	}

	switch v := args[0].(type) {
	case *object.String:
		return loadConfigFile(v.Value)
	case *object.Hash:
		// Store config values directly
		configMu.Lock()
		for _, pair := range v.Pairs {
			k := ""
			if ks, ok := pair.Key.(*object.String); ok {
				k = ks.Value
			} else {
				continue
			}
			configStore[k] = pair.Value
		}
		configMu.Unlock()
		return &object.Boolean{Value: true}
	default:
		return newError("config_load expects STRING (path) or HASH (values), got %s", args[0].Type())
	}
}

func loadConfigFile(path string) object.Object {
	data, err := os.ReadFile(path)
	if err != nil {
		return newError("config_load: cannot read '%s': %s", path, err)
	}

	ext := ""
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		ext = strings.ToLower(path[idx+1:])
	}

	switch ext {
	case "json":
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			return newError("config_load: JSON parse error in '%s': %s", path, err)
		}
		flattenConfig("", raw)
		return &object.Boolean{Value: true}
	case "yaml", "yml":
		// Use JSON as intermediary for YAML
		if YAMLToJSONfn != nil {
			jsonBytes, err := YAMLToJSONfn(data)
			if err != nil {
				return newError("config_load: YAML parse error in '%s': %s", path, err)
			}
			var raw map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &raw); err != nil {
				return newError("config_load: YAML->JSON error: %s", err)
			}
			flattenConfig("", raw)
			return &object.Boolean{Value: true}
		}
		return newError("config_load: YAML support not available (import _yaml first)")
	case "toml":
		var raw map[string]interface{}
		if err := toml.Unmarshal(data, &raw); err != nil {
			return newError("config_load: TOML parse error in '%s': %s", path, err)
		}
		flattenConfig("", raw)
		return &object.Boolean{Value: true}
	default:
		// Try JSON first, then YAML
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err == nil {
			flattenConfig("", raw)
			return &object.Boolean{Value: true}
		}
		if YAMLToJSONfn != nil {
			jsonBytes, err := YAMLToJSONfn(data)
			if err == nil {
				if err := json.Unmarshal(jsonBytes, &raw); err == nil {
					flattenConfig("", raw)
					return &object.Boolean{Value: true}
				}
			}
		}
		return newError("config_load: unsupported format for '%s' (try .json or .yaml)", path)
	}
}

var (
	configMu    sync.Mutex
	configStore = make(map[string]object.Object)
	YAMLToJSONfn func([]byte) ([]byte, error)
)

func flattenConfig(prefix string, raw map[string]interface{}) {
	for k, v := range raw {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]interface{}:
			flattenConfig(key, val)
		case string:
			configStore[key] = &object.String{Value: val}
		case float64:
			if val == float64(int64(val)) {
				configStore[key] = &object.Integer{Value: int64(val)}
			} else {
				configStore[key] = &object.Float{Value: val}
			}
		case bool:
			configStore[key] = &object.Boolean{Value: val}
		case []interface{}:
			elements := make([]object.Object, len(val))
			for i, el := range val {
				switch e := el.(type) {
				case string:
					elements[i] = &object.String{Value: e}
				case float64:
					elements[i] = &object.Float{Value: e}
				case bool:
					elements[i] = &object.Boolean{Value: e}
				default:
					elements[i] = &object.String{Value: fmt.Sprint(e)}
				}
			}
			configStore[key] = &object.Array{Elements: elements}
		case nil:
			configStore[key] = &object.Null{}
		}
	}
}

func configGet(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("config_get expects (key)")
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("config_get: key must be STRING, got %s", args[0].Type())
	}

	// Check env vars first (highest priority)
	if envVal := os.Getenv(key.Value); envVal != "" {
		return &object.String{Value: envVal}
	}
	envKey := strings.ToUpper(strings.ReplaceAll(key.Value, ".", "_"))
	if envVal := os.Getenv(envKey); envVal != "" {
		return &object.String{Value: envVal}
	}

	configMu.Lock()
	val, exists := configStore[key.Value]
	configMu.Unlock()
	if !exists {
		return &object.Null{}
	}
	return val
}

func configSet(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("config_set expects (key, value)")
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("config_set: key must be STRING, got %s", args[0].Type())
	}
	configMu.Lock()
	configStore[key.Value] = args[1]
	configMu.Unlock()
	return args[1]
}

func configEnv(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("config_env expects (key, [default])")
	}
	key, ok := args[0].(*object.String)
	if !ok {
		return newError("config_env: key must be STRING, got %s", args[0].Type())
	}
	val := os.Getenv(key.Value)
	if val == "" {
		if len(args) >= 2 {
			if s, ok := args[1].(*object.String); ok {
				return &object.String{Value: s.Value}
			}
			return args[1]
		}
		return &object.Null{}
	}
	return &object.String{Value: val}
}

func configFile(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("config_file expects (path)")
	}
	path, ok := args[0].(*object.String)
	if !ok {
		return newError("config_file: path must be STRING, got %s", args[0].Type())
	}
	return loadConfigFile(path.Value)
}
