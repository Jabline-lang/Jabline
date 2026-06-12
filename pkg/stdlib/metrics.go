package stdlib

import (
	"fmt"
	"jabline/pkg/object"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type metricCounter struct {
	name   string
	help   string
	labels []string
	values map[string]int64 // labels key -> count
}

type metricGauge struct {
	name   string
	help   string
	labels []string
	values map[string]int64
}

type metricHistogram struct {
	name    string
	help    string
	buckets []float64
	labels  []string
	values  map[string]map[int]int64 // labels key -> bucket index -> count
	count   map[string]int64
	sum     map[string]float64
}

var (
	metricsMu     sync.RWMutex
	counters      = make(map[string]*metricCounter)
	gauges        = make(map[string]*metricGauge)
	histograms    = make(map[string]*metricHistogram)
	httpRequests  int64
	httpDuration  atomic.Int64
	activeGoroutines atomic.Int64
)

var MetricsBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"metrics_counter", &object.Builtin{Fn: metricsCounter}},
	{"metrics_inc", &object.Builtin{Fn: metricsInc}},
	{"metrics_gauge", &object.Builtin{Fn: metricsGauge}},
	{"metrics_gauge_set", &object.Builtin{Fn: metricsGaugeSet}},
	{"metrics_gauge_inc", &object.Builtin{Fn: metricsGaugeInc}},
	{"metrics_gauge_dec", &object.Builtin{Fn: metricsGaugeDec}},
	{"metrics_histogram", &object.Builtin{Fn: metricsHistogram}},
	{"metrics_observe", &object.Builtin{Fn: metricsObserve}},
	{"metrics_text", &object.Builtin{Fn: metricsText}},
	{"metrics_inc_http", &object.Builtin{Fn: metricsIncHTTP}},
	{"metrics_inc_goroutine", &object.Builtin{Fn: metricsIncGoroutine}},
	{"metrics_dec_goroutine", &object.Builtin{Fn: metricsDecGoroutine}},
}

func init() {
	NativeModuleRegistry["_metrics"] = MetricsBuiltins
	NativeModulePrefixes["_metrics"] = "metrics_"
}

func metricsCounter(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("metrics_counter expects (name, help)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("metrics_counter: name must be STRING")
	}
	help, ok := args[1].(*object.String)
	if !ok {
		return newError("metrics_counter: help must be STRING")
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	if _, exists := counters[name.Value]; exists {
		return &object.Null{}
	}
	counter := &metricCounter{
		name:   name.Value,
		help:   help.Value,
		values: make(map[string]int64),
	}
	if len(args) >= 3 {
		if arr, ok := args[2].(*object.Array); ok {
			for _, el := range arr.Elements {
				if s, ok := el.(*object.String); ok {
					counter.labels = append(counter.labels, s.Value)
				}
			}
		}
	}
	counters[name.Value] = counter
	return &object.Null{}
}

func metricsInc(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("metrics_inc expects (name, [labels...])")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("metrics_inc: name must be STRING")
	}
	metricsMu.Lock()
	c, exists := counters[name.Value]
	metricsMu.Unlock()
	if !exists {
		return newError("metrics_inc: counter '%s' not registered", name.Value)
	}
	labelsKey := labelsKeyFromArgs(args[1:])
	metricsMu.Lock()
	c.values[labelsKey]++
	metricsMu.Unlock()
	return &object.Null{}
}

func metricsGauge(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("metrics_gauge expects (name, help)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("metrics_gauge: name must be STRING")
	}
	help, ok := args[1].(*object.String)
	if !ok {
		return newError("metrics_gauge: help must be STRING")
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	if _, exists := gauges[name.Value]; exists {
		return &object.Null{}
	}
	g := &metricGauge{
		name:   name.Value,
		help:   help.Value,
		values: make(map[string]int64),
	}
	if len(args) >= 3 {
		if arr, ok := args[2].(*object.Array); ok {
			for _, el := range arr.Elements {
				if s, ok := el.(*object.String); ok {
					g.labels = append(g.labels, s.Value)
				}
			}
		}
	}
	gauges[name.Value] = g
	return &object.Null{}
}

func metricsGaugeSet(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("metrics_gauge_set expects (name, value)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("metrics_gauge_set: name must be STRING")
	}
	val, ok := args[1].(*object.Integer)
	if !ok {
		return newError("metrics_gauge_set: value must be INTEGER")
	}
	metricsMu.RLock()
	g, exists := gauges[name.Value]
	metricsMu.RUnlock()
	if !exists {
		return newError("metrics_gauge_set: gauge '%s' not registered", name.Value)
	}
	labelsKey := labelsKeyFromArgs(args[2:])
	metricsMu.Lock()
	g.values[labelsKey] = val.Value
	metricsMu.Unlock()
	return &object.Null{}
}

func metricsGaugeInc(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("metrics_gauge_inc expects (name)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("metrics_gauge_inc: name must be STRING")
	}
	metricsMu.RLock()
	g, exists := gauges[name.Value]
	metricsMu.RUnlock()
	if !exists {
		return newError("metrics_gauge_inc: gauge '%s' not registered", name.Value)
	}
	labelsKey := labelsKeyFromArgs(args[1:])
	metricsMu.Lock()
	g.values[labelsKey]++
	metricsMu.Unlock()
	return &object.Null{}
}

func metricsGaugeDec(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("metrics_gauge_dec expects (name)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("metrics_gauge_dec: name must be STRING")
	}
	metricsMu.RLock()
	g, exists := gauges[name.Value]
	metricsMu.RUnlock()
	if !exists {
		return newError("metrics_gauge_dec: gauge '%s' not registered", name.Value)
	}
	labelsKey := labelsKeyFromArgs(args[1:])
	metricsMu.Lock()
	g.values[labelsKey]--
	metricsMu.Unlock()
	return &object.Null{}
}

func metricsHistogram(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("metrics_histogram expects (name, help, buckets_array)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("metrics_histogram: name must be STRING")
	}
	help, ok := args[1].(*object.String)
	if !ok {
		return newError("metrics_histogram: help must be STRING")
	}
	bucketsArr, ok := args[2].(*object.Array)
	if !ok {
		return newError("metrics_histogram: buckets must be ARRAY")
	}
	var buckets []float64
	for _, el := range bucketsArr.Elements {
		if i, ok := el.(*object.Integer); ok {
			buckets = append(buckets, float64(i.Value))
		} else if f, ok := el.(*object.Float); ok {
			buckets = append(buckets, f.Value)
		}
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	if _, exists := histograms[name.Value]; exists {
		return &object.Null{}
	}
	h := &metricHistogram{
		name:    name.Value,
		help:    help.Value,
		buckets: buckets,
		values:  make(map[string]map[int]int64),
		count:   make(map[string]int64),
		sum:     make(map[string]float64),
	}
	if len(args) >= 4 {
		if arr, ok := args[3].(*object.Array); ok {
			var labels []string
			for _, el := range arr.Elements {
				if s, ok := el.(*object.String); ok {
					labels = append(labels, s.Value)
				}
			}
			h.labels = labels
		}
	}
	histograms[name.Value] = h
	return &object.Null{}
}

func metricsObserve(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("metrics_observe expects (name, value)")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("metrics_observe: name must be STRING")
	}
	var val float64
	switch v := args[1].(type) {
	case *object.Integer:
		val = float64(v.Value)
	case *object.Float:
		val = v.Value
	default:
		return newError("metrics_observe: value must be INTEGER or FLOAT")
	}
	metricsMu.RLock()
	h, exists := histograms[name.Value]
	metricsMu.RUnlock()
	if !exists {
		return newError("metrics_observe: histogram '%s' not registered", name.Value)
	}
	labelsKey := labelsKeyFromArgs(args[2:])
	metricsMu.Lock()
	h.sum[labelsKey] += val
	h.count[labelsKey]++
	if h.values[labelsKey] == nil {
		h.values[labelsKey] = make(map[int]int64)
	}
	for i, b := range h.buckets {
		if val <= b {
			h.values[labelsKey][i]++
		}
	}
	metricsMu.Unlock()
	return &object.Null{}
}

func metricsText(args ...object.Object) object.Object {
	var sb strings.Builder
	metricsMu.RLock()
	defer metricsMu.RUnlock()
	
	now := time.Now().UnixMilli()
	
	// Counters
	names := make([]string, 0, len(counters))
	for n := range counters {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		c := counters[n]
		fmt.Fprintf(&sb, "# HELP %s %s\n", c.name, c.help)
		fmt.Fprintf(&sb, "# TYPE %s counter\n", c.name)
		if len(c.labels) == 0 {
			total := int64(0)
			for _, v := range c.values {
				total += v
			}
			fmt.Fprintf(&sb, "%s %d %d\n", c.name, total, now)
		} else {
			for lk, v := range c.values {
				sb.WriteString(formatLabels(c.name, c.labels, lk, v, now))
			}
		}
	}
	
	// Gauges
	for _, n := range names[:0] {
		_ = n
	}
	gNames := make([]string, 0, len(gauges))
	for n := range gauges {
		gNames = append(gNames, n)
	}
	sort.Strings(gNames)
	for _, n := range gNames {
		g := gauges[n]
		fmt.Fprintf(&sb, "# HELP %s %s\n", g.name, g.help)
		fmt.Fprintf(&sb, "# TYPE %s gauge\n", g.name)
		if len(g.labels) == 0 {
			total := int64(0)
			for _, v := range g.values {
				total += v
			}
			fmt.Fprintf(&sb, "%s %d %d\n", g.name, total, now)
		} else {
			for lk, v := range g.values {
				sb.WriteString(formatLabels(g.name, g.labels, lk, v, now))
			}
		}
	}
	
	// Histograms
	hNames := make([]string, 0, len(histograms))
	for n := range histograms {
		hNames = append(hNames, n)
	}
	sort.Strings(hNames)
	for _, n := range hNames {
		h := histograms[n]
		fmt.Fprintf(&sb, "# HELP %s %s\n", h.name, h.help)
		fmt.Fprintf(&sb, "# TYPE %s histogram\n", h.name)
		if len(h.labels) == 0 {
			emptyKey := ""
			writeHistogram(&sb, h, emptyKey, now)
		} else {
			for lk := range h.values {
				writeHistogram(&sb, h, lk, now)
			}
		}
	}
	
	// Go runtime info
	fmt.Fprintf(&sb, "# HELP jabline_goroutines Number of active goroutines\n")
	fmt.Fprintf(&sb, "# TYPE jabline_goroutines gauge\n")
	fmt.Fprintf(&sb, "jabline_goroutines %d %d\n", activeGoroutines.Load(), now)
	
	fmt.Fprintf(&sb, "# HELP jabline_http_requests_total Total HTTP requests\n")
	fmt.Fprintf(&sb, "# TYPE jabline_http_requests_total counter\n")
	fmt.Fprintf(&sb, "jabline_http_requests_total %d %d\n", atomic.LoadInt64(&httpRequests), now)
	
	fmt.Fprintf(&sb, "# HELP jabline_uptime_seconds Uptime in seconds\n")
	fmt.Fprintf(&sb, "# TYPE jabline_uptime_seconds gauge\n")
	fmt.Fprintf(&sb, "jabline_uptime_seconds %d %d\n", int64(time.Since(startTime).Seconds()), now)
	
	return &object.String{Value: sb.String()}
}

func writeHistogram(sb *strings.Builder, h *metricHistogram, labelsKey string, now int64) {
	prefix := h.name
	labelStr := ""
	if labelsKey != "" && len(h.labels) > 0 {
		parts := strings.Split(labelsKey, ",")
		var lp []string
		for i, l := range h.labels {
			if i < len(parts) {
				lp = append(lp, fmt.Sprintf("%s=\"%s\"", l, parts[i]))
			}
		}
		labelStr = "{" + strings.Join(lp, ",") + "}"
	}
	
	totalCount := h.count[labelsKey]
	totalSum := h.sum[labelsKey]
	
	for i, b := range h.buckets {
		count := h.values[labelsKey][i]
		fmt.Fprintf(sb, "%s_bucket%s{le=\"%g\"} %d %d\n", prefix, labelStr, b, count, now)
	}
	fmt.Fprintf(sb, "%s_bucket%s{le=\"+Inf\"} %d %d\n", prefix, labelStr, totalCount, now)
	fmt.Fprintf(sb, "%s_count%s %d %d\n", prefix, labelStr, totalCount, now)
	fmt.Fprintf(sb, "%s_sum%s %g %d\n", prefix, labelStr, totalSum, now)
}

func formatLabels(name string, labels []string, labelsKey string, value int64, now int64) string {
	if labelsKey == "" {
		return fmt.Sprintf("%s %d %d\n", name, value, now)
	}
	parts := strings.Split(labelsKey, ",")
	var lp []string
	for i, l := range labels {
		if i < len(parts) {
			lp = append(lp, fmt.Sprintf("%s=\"%s\"", l, parts[i]))
		}
	}
	return fmt.Sprintf("%s{%s} %d %d\n", name, strings.Join(lp, ","), value, now)
}

func labelsKeyFromArgs(args []object.Object) string {
	if len(args) == 0 {
		return ""
	}
	var parts []string
	for _, a := range args {
		if s, ok := a.(*object.String); ok {
			parts = append(parts, s.Value)
		} else if i, ok := a.(*object.Integer); ok {
			parts = append(parts, fmt.Sprintf("%d", i.Value))
		} else {
			parts = append(parts, a.Inspect())
		}
	}
	return strings.Join(parts, ",")
}

func metricsIncHTTP(args ...object.Object) object.Object {
	atomic.AddInt64(&httpRequests, 1)
	return &object.Null{}
}

func metricsIncGoroutine(args ...object.Object) object.Object {
	activeGoroutines.Add(1)
	return &object.Null{}
}

func metricsDecGoroutine(args ...object.Object) object.Object {
	activeGoroutines.Add(-1)
	return &object.Null{}
}
