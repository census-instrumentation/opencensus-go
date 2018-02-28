package opentracer

import (
	"reflect"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.opencensus.io/trace"
)

// Span represents an active, un-finished span in the OpenTracing system.
//
// Spans are created by the Tracer interface.
type Span struct {
	tracer *Tracer
	ocSpan *trace.Span

	mutex   sync.Mutex  // mutex guards baggage and tags; because you never know
	baggage []log.Field // baggage contains local baggage
	tags    []log.Field // array of tags
}

func (s *Span) mergeFields(fields ...log.Field) []log.Field {
	var merged []log.Field
	merged = upsert(merged, s.baggage...)
	merged = upsert(merged, s.tags...)
	merged = upsert(merged, fields...)
	return merged
}

func (s *Span) logFields(fields ...log.Field) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tracer.logger.LogFields(s.mergeFields(fields...)...)
}

func (s *Span) logFieldsTime(t time.Time, fields ...log.Field) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tracer.logger.LogFieldsTime(t, s.mergeFields(fields...)...)
}

// Sets the end timestamp and finalizes Span state.
//
// With the exception of calls to Context() (which are always allowed),
// Finish() must be the last call made to any span instance, and to do
// otherwise leads to undefined behavior.
func (s *Span) Finish() {
	s.ocSpan.End()
}

// FinishWithOptions is like Finish() but with explicit control over
// timestamps and log data.
func (s *Span) FinishWithOptions(opts opentracing.FinishOptions) {
	for _, record := range opts.LogRecords {
		var t = record.Timestamp
		if t.IsZero() {
			t = time.Now()
		}
		s.logFieldsTime(record.Timestamp, record.Fields...)
	}

	if len(opts.BulkLogData) > 0 {
		s.logFields(log.String("message", "BulkLogData is deprecated.  Please use LogRecords instead."))
	}

	s.ocSpan.End()
}

// ForeachBaggageItem grants access to all baggage items stored in the
// SpanContext.
// The handler function will be called for each baggage key/value pair.
// The ordering of items is not guaranteed.
//
// The bool return value indicates if the handler wants to continue iterating
// through the rest of the baggage items; for example if the handler is trying to
// find some baggage item by pattern matching the name, it can return false
// as soon as the item is found to stop further iterations.
func (s *Span) ForeachBaggageItem(handler func(k, v string) bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, field := range s.baggage {
		if ok := handler(field.Key(), field.Value().(string)); !ok {
			return
		}
	}
}

// Context() yields the SpanContext for this Span. Note that the return
// value of Context() is still valid after a call to Span.Finish(), as is
// a call to Span.Context() after a call to Span.Finish().
func (s *Span) Context() opentracing.SpanContext {
	return s
}

// Sets or changes the operation name.
func (s *Span) SetOperationName(operationName string) opentracing.Span {
	s.logFields(log.String("message", "SetOperationName is not supported.  operationName is immutable."))
	return s
}

// Adds a tag to the span.
//
// If there is a pre-existing tag set for `key`, it is overwritten.
//
// Tag values can be numeric types, strings, or bools. The behavior of
// other tag value types is undefined at the OpenTracing level. If a
// tracing system does not know how to handle a particular value type, it
// may ignore the tag, but shall not panic.
func (s *Span) SetTag(key string, value interface{}) opentracing.Span {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if isValidTagValue(value) {
		for _, field := range makeLogFields(key, value) {
			s.tags = upsert(s.tags, field)
		}
	}

	return s
}

// LogFields is an efficient and type-checked way to record key:value
// logging data about a Span, though the programming interface is a little
// more verbose than LogKV(). Here's an example:
//
//    span.LogFields(
//        log.String("event", "soft error"),
//        log.String("type", "cache timeout"),
//        log.Int("waited.millis", 1500))
//
// Also see Span.FinishWithOptions() and FinishOptions.BulkLogData.
func (s *Span) LogFields(fields ...log.Field) {
	s.logFields(fields...)
}

// LogKV is a concise, readable way to record key:value logging data about
// a Span, though unfortunately this also makes it less efficient and less
// type-safe than LogFields(). Here's an example:
//
//    span.LogKV(
//        "event", "soft error",
//        "type", "cache timeout",
//        "waited.millis", 1500)
//
// For LogKV (as opposed to LogFields()), the parameters must appear as
// key-value pairs, like
//
//    span.LogKV(key1, val1, key2, val2, key3, val3, ...)
//
// The keys must all be strings. The values may be strings, numeric types,
// bools, Go error instances, or arbitrary structs.
//
// (Note to implementors: consider the log.InterleavedKVToFields() helper)
func (s *Span) LogKV(alternatingKeyValues ...interface{}) {
	s.logFields(makeLogFields(alternatingKeyValues...)...)
}

// SetBaggageItem sets a key:value pair on this Span and its SpanContext
// that also propagates to descendants of this Span.
//
// SetBaggageItem() enables powerful functionality given a full-stack
// opentracing integration (e.g., arbitrary application data from a mobile
// app can make it, transparently, all the way into the depths of a storage
// system), and with it some powerful costs: use this feature with care.
//
// IMPORTANT NOTE #1: SetBaggageItem() will only propagate baggage items to
// *future* causal descendants of the associated Span.
//
// IMPORTANT NOTE #2: Use this thoughtfully and with care. Every key and
// value is copied into every local *and remote* child of the associated
// Span, and that can add up to a lot of network and cpu overhead.
//
// Returns a reference to this Span for chaining.
func (s *Span) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.baggage = upsert(s.baggage, log.String(restrictedKey, value))
	return s
}

// Gets the value for a baggage item given its key. Returns the empty string
// if the value isn't found in this Span.
func (s *Span) BaggageItem(restrictedKey string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, item := range s.baggage {
		if item.Key() == restrictedKey {
			return item.Value().(string)
		}
	}
	return ""
}

// Provides access to the Tracer that created this Span.
func (s *Span) Tracer() opentracing.Tracer {
	return s.tracer
}

// Deprecated: use LogFields or LogKV
func (s *Span) LogEvent(event string) {
	s.logFields(log.String("message", "deprecated LogEvent called"))
}

// Deprecated: use LogFields or LogKV
func (s *Span) LogEventWithPayload(event string, payload interface{}) {
	s.logFields(log.String("message", "deprecated LogEventWithPayload called"))
}

// Deprecated: use LogFields or LogKV
func (s *Span) Log(data opentracing.LogData) {
	s.logFields(log.String("message", "deprecated Log called"))
}

// isValidTagValue returns true if value is numeric, string, or bool
func isValidTagValue(value interface{}) bool {
	switch value.(type) {
	case bool:
		return true
	case string:
		return true
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}

// makeLogFields converts key value pairs into log.Field values
// odd number values will be dropped
func makeLogFields(kvs ...interface{}) []log.Field {
	var fields []log.Field

	for i, length := 0, len(kvs); i < length; i += 2 {
		var (
			key, keyOk = kvs[i].(string)
			value      = kvs[i+1]
		)
		if !keyOk {
			continue
		}

		switch v := value.(type) {
		case bool:
			fields = append(fields, log.Bool(key, v))
		case string:
			fields = append(fields, log.String(key, v))
		case int:
			fields = append(fields, log.Int(key, v))
		case int8:
			fields = append(fields, log.Int(key, int(v)))
		case int16:
			fields = append(fields, log.Int(key, int(v)))
		case int32:
			fields = append(fields, log.Int32(key, v))
		case int64:
			fields = append(fields, log.Int64(key, v))
		case uint:
			fields = append(fields, log.Uint32(key, uint32(v)))
		case uint8:
			fields = append(fields, log.Uint32(key, uint32(v)))
		case uint16:
			fields = append(fields, log.Uint32(key, uint32(v)))
		case uint32:
			fields = append(fields, log.Uint32(key, v))
		case uint64:
			fields = append(fields, log.Uint64(key, v))
		case float32:
			fields = append(fields, log.Float32(key, v))
		case float64:
			fields = append(fields, log.Float64(key, v))
		}
	}

	return fields
}

// upsert updates an existing field if changed or appends to the list
// if this is a new field
func upsert(original []log.Field, fields ...log.Field) []log.Field {
loop:
	for _, field := range fields {
		for i, length := 0, len(original); i < length; i++ {
			if f := original[i]; f.Key() == field.Key() {
				if reflect.DeepEqual(f.Value(), field.Value()) {
					continue loop

				} else {
					original = append(original[:i], original[i+1:]...)
					original = append(original, field)
					continue loop
				}
			}
		}

		original = append(original, field)
	}

	return original
}
