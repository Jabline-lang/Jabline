package stdlib

import (
	"fmt"
	"jabline/pkg/object"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"jwt_sign", &object.Builtin{Fn: jwtSign}},
	{"jwt_verify", &object.Builtin{Fn: jwtVerify}},
	{"jwt_decode", &object.Builtin{Fn: jwtDecode}},
}

func init() {
	NativeModuleRegistry["_jwt"] = JWTBuiltins
	NativeModulePrefixes["_jwt"] = "jwt_"
}

func jwtSign(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("jwt_sign expects at least 2 args: (payload_hash, secret)")
	}

	payload, ok := args[0].(*object.Hash)
	if !ok {
		return newError("payload must be HASH, got %s", args[0].Type())
	}
	secret, ok := args[1].(*object.String)
	if !ok {
		return newError("secret must be STRING, got %s", args[1].Type())
	}

	claims := jwt.MapClaims{}
	for _, pair := range payload.Pairs {
		k := ""
		if ks, ok := pair.Key.(*object.String); ok {
			k = ks.Value
		} else {
			continue
		}
		switch v := pair.Value.(type) {
		case *object.String:
			claims[k] = v.Value
		case *object.Integer:
			claims[k] = v.Value
		case *object.Float:
			claims[k] = v.Value
		case *object.Boolean:
			claims[k] = v.Value
		case *object.Null:
			claims[k] = nil
		case *object.Array:
			var arr []interface{}
			for _, el := range v.Elements {
				if s, ok := el.(*object.String); ok {
					arr = append(arr, s.Value)
				} else if i, ok := el.(*object.Integer); ok {
					arr = append(arr, i.Value)
				} else if f, ok := el.(*object.Float); ok {
					arr = append(arr, f.Value)
				} else if b, ok := el.(*object.Boolean); ok {
					arr = append(arr, b.Value)
				}
			}
			claims[k] = arr
		}
	}

	// Handle special registered claims from hash keys
	if sub, ok := claims["sub"]; ok {
		claims["sub"] = fmt.Sprint(sub)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret.Value))
	if err != nil {
		return newError("jwt_sign failed: %s", err)
	}

	return &object.String{Value: signed}
}

func jwtVerify(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("jwt_verify expects 2 args: (token, secret)")
	}
	tokenStr, ok := args[0].(*object.String)
	if !ok {
		return newError("token must be STRING, got %s", args[0].Type())
	}
	secret, ok := args[1].(*object.String)
	if !ok {
		return newError("secret must be STRING, got %s", args[1].Type())
	}

	token, err := jwt.Parse(tokenStr.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret.Value), nil
	})

	if err != nil {
		return &object.Hash{
			Pairs: map[object.HashKey]object.HashPair{
				(&object.String{Value: "valid"}).HashKey():   {Key: &object.String{Value: "valid"}, Value: &object.Boolean{Value: false}},
				(&object.String{Value: "error"}).HashKey():   {Key: &object.String{Value: "error"}, Value: &object.String{Value: err.Error()}},
			},
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return &object.Hash{
			Pairs: map[object.HashKey]object.HashPair{
				(&object.String{Value: "valid"}).HashKey(): {Key: &object.String{Value: "valid"}, Value: &object.Boolean{Value: false}},
			},
		}
	}

	pairs := make(map[object.HashKey]object.HashPair)
	add := func(k string, v object.Object) {
		ks := &object.String{Value: k}
		pairs[ks.HashKey()] = object.HashPair{Key: ks, Value: v}
	}

	add("valid", &object.Boolean{Value: true})
	for k, v := range claims {
		switch val := v.(type) {
		case string:
			add(k, &object.String{Value: val})
		case float64:
			add(k, &object.Float{Value: val})
		case int64:
			add(k, &object.Integer{Value: val})
		case bool:
			add(k, &object.Boolean{Value: val})
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
			add(k, &object.Array{Elements: elements})
		}
	}

	return &object.Hash{Pairs: pairs}
}

func jwtDecode(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("jwt_decode expects 1 arg: (token)")
	}
	tokenStr, ok := args[0].(*object.String)
	if !ok {
		return newError("token must be STRING, got %s", args[0].Type())
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenStr.Value, jwt.MapClaims{})
	if err != nil {
		return newError("jwt_decode failed: %s", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return newError("jwt_decode: invalid claims")
	}

	pairs := make(map[object.HashKey]object.HashPair)
	add := func(k string, v object.Object) {
		ks := &object.String{Value: k}
		pairs[ks.HashKey()] = object.HashPair{Key: ks, Value: v}
	}

	// Add header
	add("__header__", &object.String{Value: fmt.Sprintf("%v", token.Header)})

	for k, v := range claims {
		switch val := v.(type) {
		case string:
			add(k, &object.String{Value: val})
		case float64:
			// JWT numeric dates
			if k == "exp" || k == "nbf" || k == "iat" {
				tm := time.Unix(int64(val), 0)
				add(k, &object.String{Value: tm.Format(time.RFC3339)})
				add(k+"_unix", &object.Integer{Value: int64(val)})
			} else {
				add(k, &object.Float{Value: val})
			}
		case bool:
			add(k, &object.Boolean{Value: val})
		case []interface{}:
			elements := make([]object.Object, len(val))
			for i, el := range val {
				switch e := el.(type) {
				case string:
					elements[i] = &object.String{Value: e}
				case float64:
					elements[i] = &object.Float{Value: e}
				default:
					elements[i] = &object.String{Value: fmt.Sprint(e)}
				}
			}
			add(k, &object.Array{Elements: elements})
		default:
			add(k, &object.String{Value: fmt.Sprint(v)})
		}
	}

	return &object.Hash{Pairs: pairs}
}
