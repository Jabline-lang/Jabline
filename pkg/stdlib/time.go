package stdlib

import (
	"jabline/pkg/object"
	"time"
)

var TimeBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"now", &object.Builtin{Fn: timeNow}},
	{"unix", &object.Builtin{Fn: timeUnix}},
	{"sleep", &object.Builtin{Fn: timeSleep}},
}

func timeNow(args ...object.Object) object.Object {
	return &object.Integer{Value: time.Now().Unix()}
}

func timeUnix(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	arg, ok := args[0].(*object.Integer)
	if !ok {
		return newError("argument to `unix` must be INTEGER, got %s", args[0].Type())
	}
	// This would return a helper for formatting usually, but for now we return the same or more info.
	// We'll use this primarily in datetime.jb to breakdown.
	t := time.Unix(arg.Value, 0)

	pairs := make(map[object.HashKey]object.HashPair)

	add := func(k string, v int64) {
		ks := &object.String{Value: k}
		pairs[ks.HashKey()] = object.HashPair{Key: ks, Value: &object.Integer{Value: v}}
	}

	add("year", int64(t.Year()))
	add("month", int64(t.Month()))
	add("day", int64(t.Day()))
	add("hour", int64(t.Hour()))
	add("minute", int64(t.Minute()))
	add("second", int64(t.Second()))
	add("weekday", int64(t.Weekday()))

	return &object.Hash{Pairs: pairs}
}

func timeSleep(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	ms, ok := args[0].(*object.Integer)
	if !ok {
		return newError("argument to `sleep` must be INTEGER (ms), got %s", args[0].Type())
	}
	time.Sleep(time.Duration(ms.Value) * time.Millisecond)
	return &object.Null{}
}
