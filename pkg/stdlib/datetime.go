package stdlib

import (
	"time"

	"jabline/pkg/object"
)

func init() {
	NativeModuleRegistry["_datetime"] = DateTimeBuiltins
	NativeModulePrefixes["_datetime"] = "dt_"
}

var DateTimeBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"dt_now", &object.Builtin{Fn: dtNow}},
	{"dt_parse", &object.Builtin{Fn: dtParse}},
	{"dt_format", &object.Builtin{Fn: dtFormat}},
	{"dt_unix", &object.Builtin{Fn: dtUnix}},
	{"dt_add", &object.Builtin{Fn: dtAdd}},
	{"dt_sub", &object.Builtin{Fn: dtSub}},
	{"dt_diff", &object.Builtin{Fn: dtDiff}},
	{"dt_year", &object.Builtin{Fn: dtYear}},
	{"dt_month", &object.Builtin{Fn: dtMonth}},
	{"dt_day", &object.Builtin{Fn: dtDay}},
	{"dt_hour", &object.Builtin{Fn: dtHour}},
	{"dt_minute", &object.Builtin{Fn: dtMinute}},
	{"dt_second", &object.Builtin{Fn: dtSecond}},
	{"dt_weekday", &object.Builtin{Fn: dtWeekday}},
}

func dtNow(args ...object.Object) object.Object {
	return &object.DateTime{Time: time.Now()}
}

func dtParse(args ...object.Object) object.Object {
	if len(args) != 1 && len(args) != 2 {
		return newError("dt_parse expects 1-2 arguments, got %d", len(args))
	}

	s, ok := args[0].(*object.String)
	if !ok {
		return newError("dt_parse expects string, got %s", args[0].Type())
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		time.RFC1123,
		time.RFC1123Z,
	}

	if len(args) == 2 {
		layout, ok := args[1].(*object.String)
		if !ok {
			return newError("dt_parse expects string layout, got %s", args[1].Type())
		}
		t, err := time.Parse(layout.Value, s.Value)
		if err != nil {
			return newError("dt_parse error: %s", err.Error())
		}
		return &object.DateTime{Time: t}
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s.Value); err == nil {
			return &object.DateTime{Time: t}
		}
	}

	return newError("dt_parse: unable to parse %q", s.Value)
}

func dtFormat(args ...object.Object) object.Object {
	if len(args) != 1 && len(args) != 2 {
		return newError("dt_format expects 1-2 arguments, got %d", len(args))
	}

	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_format expects datetime, got %s", args[0].Type())
	}

	layout := time.RFC3339
	if len(args) == 2 {
		l, ok := args[1].(*object.String)
		if !ok {
			return newError("dt_format expects string layout, got %s", args[1].Type())
		}
		layout = l.Value
	}

	return &object.String{Value: dt.Time.Format(layout)}
}

func dtUnix(args ...object.Object) object.Object {
	if len(args) == 0 {
		// Get current unix timestamp
		return &object.Integer{Value: time.Now().Unix()}
	}

	dt, ok := args[0].(*object.DateTime)
	if !ok {
		// Convert from unix timestamp
		sec, ok := args[0].(*object.Integer)
		if !ok {
			return newError("dt_unix expects datetime or integer, got %s", args[0].Type())
		}
		return &object.DateTime{Time: time.Unix(sec.Value, 0)}
	}

	return &object.Integer{Value: dt.Time.Unix()}
}

func dtAdd(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("dt_add expects 2 arguments, got %d", len(args))
	}

	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_add expects datetime, got %s", args[0].Type())
	}

	dur, ok := args[1].(*object.Integer)
	if !ok {
		return newError("dt_add expects integer seconds, got %s", args[1].Type())
	}

	return &object.DateTime{Time: dt.Time.Add(time.Duration(dur.Value) * time.Second)}
}

func dtSub(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("dt_sub expects 2 arguments, got %d", len(args))
	}

	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_sub expects datetime, got %s", args[0].Type())
	}

	dur, ok := args[1].(*object.Integer)
	if !ok {
		return newError("dt_sub expects integer seconds, got %s", args[1].Type())
	}

	return &object.DateTime{Time: dt.Time.Add(-time.Duration(dur.Value) * time.Second)}
}

func dtDiff(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("dt_diff expects 2 arguments, got %d", len(args))
	}

	dt1, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_diff expects datetime, got %s", args[0].Type())
	}

	dt2, ok := args[1].(*object.DateTime)
	if !ok {
		return newError("dt_diff expects datetime, got %s", args[1].Type())
	}

	return &object.Integer{Value: int64(dt1.Time.Sub(dt2.Time).Seconds())}
}

func dtYear(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("dt_year expects 1 argument, got %d", len(args))
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_year expects datetime, got %s", args[0].Type())
	}
	return &object.Integer{Value: int64(dt.Time.Year())}
}

func dtMonth(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("dt_month expects 1 argument, got %d", len(args))
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_month expects datetime, got %s", args[0].Type())
	}
	return &object.Integer{Value: int64(dt.Time.Month())}
}

func dtDay(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("dt_day expects 1 argument, got %d", len(args))
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_day expects datetime, got %s", args[0].Type())
	}
	return &object.Integer{Value: int64(dt.Time.Day())}
}

func dtHour(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("dt_hour expects 1 argument, got %d", len(args))
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_hour expects datetime, got %s", args[0].Type())
	}
	return &object.Integer{Value: int64(dt.Time.Hour())}
}

func dtMinute(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("dt_minute expects 1 argument, got %d", len(args))
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_minute expects datetime, got %s", args[0].Type())
	}
	return &object.Integer{Value: int64(dt.Time.Minute())}
}

func dtSecond(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("dt_second expects 1 argument, got %d", len(args))
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_second expects datetime, got %s", args[0].Type())
	}
	return &object.Integer{Value: int64(dt.Time.Second())}
}

func dtWeekday(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("dt_weekday expects 1 argument, got %d", len(args))
	}
	dt, ok := args[0].(*object.DateTime)
	if !ok {
		return newError("dt_weekday expects datetime, got %s", args[0].Type())
	}
	return &object.String{Value: dt.Time.Weekday().String()}
}
