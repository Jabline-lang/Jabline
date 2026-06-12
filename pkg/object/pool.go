package object

import "sync"

var (
	integerPool = sync.Pool{
		New: func() interface{} {
			return &Integer{}
		},
	}

	stringPool = sync.Pool{
		New: func() interface{} {
			return &String{}
		},
	}

	booleanPool = sync.Pool{
		New: func() interface{} {
			return &Boolean{}
		},
	}

	// Singleton instances for true/false to avoid any allocation for booleans
	// when possible, though the pool is still useful if we dynamically create them
	// without going through constants.
	TrueObj  = &Boolean{Value: true}
	FalseObj = &Boolean{Value: false}
)

// NewInteger creates an Integer object or reuses one from the pool
func NewInteger(value int64) *Integer {
	obj := integerPool.Get().(*Integer)
	obj.Value = value
	return obj
}

// ReleaseInteger returns an Integer to the pool safely
func ReleaseInteger(obj *Integer) {
	// Only release if it's safe (user must ensure no dangling pointers)
	integerPool.Put(obj)
}

// NewString creates a String object or reuses one from the pool
func NewString(value string) *String {
	obj := stringPool.Get().(*String)
	obj.Value = value
	return obj
}

// ReleaseString returns a String to the pool safely
func ReleaseString(obj *String) {
	obj.Value = "" // Prevent memory leaks from holding large strings in pool
	stringPool.Put(obj)
}

// NewBoolean returns the singleton True/False objects if possible.
// If not, it uses the pool. (In Jabline, Booleans are mostly singletons)
func NewBoolean(value bool) *Boolean {
	if value {
		return TrueObj
	}
	return FalseObj
}

// ReleaseObject safely determines if an object is poolable and releases it.
// Call this ONLY on intermediate results that you are 100% sure are not referenced elsewhere.
func ReleaseObject(obj Object) {
	if obj == nil {
		return
	}
	switch o := obj.(type) {
	case *Integer:
		ReleaseInteger(o)
	case *String:
		ReleaseString(o)
	// We don't release Booleans because they are singletons now (TrueObj / FalseObj)
	}
}
