package opentracer

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func TestUpsert(t *testing.T) {
	var (
		a  = log.String("a", "apple")
		b  = log.String("b", "boy")
		c  = log.String("c", "cat")
		a2 = log.String("a", "adam")
		a3 = log.Int64("a", 1)
	)

	t.Run("nil", func(t *testing.T) {
		var actual = upsert(nil, a)
		if expected := []log.Field{a}; !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %v; got %v", expected, actual)
		}
	})

	t.Run("append", func(t *testing.T) {
		var actual = upsert([]log.Field{a}, b)
		if expected := []log.Field{a, b}; !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %v; got %v", expected, actual)
		}
	})

	t.Run("unchanged", func(t *testing.T) {
		var actual = upsert([]log.Field{a}, a)
		if expected := []log.Field{a}; !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %v; got %v", expected, actual)
		}
	})

	t.Run("replace same type", func(t *testing.T) {
		var actual = upsert([]log.Field{a}, a2)
		if expected := []log.Field{a2}; !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %v; got %v", expected, actual)
		}
	})

	t.Run("replace different type type", func(t *testing.T) {
		var actual = upsert([]log.Field{a}, a3)
		if expected := []log.Field{a3}; !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %v; got %v", expected, actual)
		}
	})

	t.Run("replace from head", func(t *testing.T) {
		var actual = upsert([]log.Field{a, b, c}, a2)
		if expected := []log.Field{b, c, a2}; !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %v; got %v", expected, actual)
		}
	})

	t.Run("upsert multiple", func(t *testing.T) {
		var actual = upsert([]log.Field{a, b}, a2, c)
		if expected := []log.Field{b, a2, c}; !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %v; got %v", expected, actual)
		}
	})
}

type logCapturer struct {
	logs [][]log.Field
}

func (l *logCapturer) LogFields(fields ...log.Field) {
	l.logs = append(l.logs, fields)
}

func (l *logCapturer) LogFieldsTime(t time.Time, fields ...log.Field) {
	l.logs = append(l.logs, fields)
}

func setup() *logCapturer {
	var (
		logger logCapturer
		tracer = New(&logger)
	)
	opentracing.SetGlobalTracer(tracer)

	return &logger
}

func TestSpan(t *testing.T) {
	t.Run("basic log", func(t *testing.T) {
		var (
			ctx    = context.Background()
			logger = setup()
			field  = log.String("message", "hello world")
		)

		span, _ := opentracing.StartSpanFromContext(ctx, "span")
		span.LogFields(field)
		span.Finish()

		if expected := 1; len(logger.logs) != expected {
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
		if expected := []log.Field{field}; !reflect.DeepEqual(expected, logger.logs[0]) {
			t.Fatalf("expected %#v; got %#v", expected, logger.logs[0])
		}
	})

	t.Run("logs tags", func(t *testing.T) {
		var (
			ctx    = context.Background()
			logger = setup()
			tag    = log.String("foo", "bar")
			field  = log.String("message", "hello world")
		)

		span, _ := opentracing.StartSpanFromContext(ctx, "span")
		span.SetTag(tag.Key(), tag.Value())
		span.LogFields(field)
		span.Finish()

		if expected := 1; len(logger.logs) != expected {
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
		if expected := []log.Field{tag, field}; !reflect.DeepEqual(expected, logger.logs[0]) {
			t.Fatalf("expected %#v; got %#v", expected, logger.logs[0])
		}
	})

	t.Run("logs baggage", func(t *testing.T) {
		var (
			ctx     = context.Background()
			logger  = setup()
			baggage = log.String("foo", "bar")
			field   = log.String("message", "hello world")
		)

		span, _ := opentracing.StartSpanFromContext(ctx, "span")
		span.SetBaggageItem(baggage.Key(), baggage.Value().(string))
		span.LogFields(field)
		span.Finish()

		if expected := 1; len(logger.logs) != expected {
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
		if expected := []log.Field{baggage, field}; !reflect.DeepEqual(expected, logger.logs[0]) {
			t.Fatalf("expected %#v; got %#v", expected, logger.logs[0])
		}
	})

	t.Run("tags override baggage", func(t *testing.T) {
		var (
			ctx     = context.Background()
			logger  = setup()
			baggage = log.String("key", "baggage")
			tag     = log.String("key", "tag")
			field   = log.String("message", "hello world")
		)

		span, _ := opentracing.StartSpanFromContext(ctx, "span")
		span.SetTag(tag.Key(), tag.Value())
		span.SetBaggageItem(baggage.Key(), baggage.Value().(string))
		span.LogFields(field)
		span.Finish()

		if expected := 1; len(logger.logs) != expected {
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
		if expected := []log.Field{tag, field}; !reflect.DeepEqual(expected, logger.logs[0]) {
			t.Fatalf("expected %#v; got %#v", expected, logger.logs[0])
		}
	})

	t.Run("fields overrides tags", func(t *testing.T) {
		var (
			ctx    = context.Background()
			logger = setup()
			tag    = log.String("key", "tag")
			field  = log.String("key", "hello world")
		)

		span, _ := opentracing.StartSpanFromContext(ctx, "span")
		span.SetTag(tag.Key(), tag.Value())
		span.LogFields(field)
		span.Finish()

		if expected := 1; len(logger.logs) != expected {
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
		if expected := []log.Field{field}; !reflect.DeepEqual(expected, logger.logs[0]) {
			t.Fatalf("expected %#v; got %#v", expected, logger.logs[0])
		}
	})

	t.Run("LogKV", func(t *testing.T) {
		var (
			ctx    = context.Background()
			logger = setup()
			field  = log.String("key", "hello world")
		)

		span, _ := opentracing.StartSpanFromContext(ctx, "span")
		span.LogKV(field.Key(), field.Value())
		span.Finish()

		if expected := 1; len(logger.logs) != expected {
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
		if expected := []log.Field{field}; !reflect.DeepEqual(expected, logger.logs[0]) {
			t.Fatalf("expected %#v; got %#v", expected, logger.logs[0])
		}
	})

	t.Run("Tracer", func(t *testing.T) {
		var (
			ctx = context.Background()
		)

		span, _ := opentracing.StartSpanFromContext(ctx, "span")
		if span.Tracer() == nil {
			t.Errorf("expected Tracer; got nil")
		}
	})

	t.Run("BaggageItem propagates to children", func(t *testing.T) {
		var (
			ctx   = context.Background()
			key   = "key"
			value = "value"
		)

		parent, ctx := opentracing.StartSpanFromContext(ctx, "parent")
		parent.SetBaggageItem(key, value)

		child, _ := opentracing.StartSpanFromContext(ctx, "child")

		if expected := value; child.BaggageItem(key) != expected {
			t.Errorf("expected %v; got %v", expected, child.BaggageItem(key))
		}
	})

	t.Run("logs tags set during span initialization", func(t *testing.T) {
		var (
			ctx     = context.Background()
			logger  = setup()
			tag     = log.String("key", "value")
			message = log.String("hello", "world")
		)

		span, ctx := opentracing.StartSpanFromContext(ctx, "span", opentracing.Tags{
			tag.Key(): tag.Value(),
		})
		span.LogFields(message)
		span.Finish()

		if expected := 1; len(logger.logs) != expected {
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
		if expected := []log.Field{tag, message}; !reflect.DeepEqual(expected, logger.logs[0]) {
			t.Errorf("expected %v; got %v", expected, logger.logs[0])
		}
	})

	t.Run("ForeachBaggageItem", func(t *testing.T) {
		var (
			ctx   = context.Background()
			key   = "key"
			value = "value"
		)

		span, ctx := opentracing.StartSpanFromContext(ctx, "span")
		span.SetBaggageItem(key, value)

		var (
			count = 0
			seen  = false
		)
		span.Context().ForeachBaggageItem(func(k, v string) bool {
			count++
			if k == key {
				seen = true
				return false
			}
			return true
		})

		if count != 1 {
			t.Errorf("expected 1 baggage item; got %v", count)
		}
		if !seen {
			t.Errorf("expected our baggage item to be seen")
		}
	})

	t.Run("ForeachBaggageItem", func(t *testing.T) {
		var (
			ctx    = context.Background()
			logger = setup()
		)

		span, ctx := opentracing.StartSpanFromContext(ctx, "span")
		span.SetOperationName("blah")

		if expected := 1; len(logger.logs) != expected {
			// no need to read message about it not being supported
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
	})
	t.Run("ForeachBaggageItem", func(t *testing.T) {
		var (
			ctx    = context.Background()
			logger = setup()
			field  = log.String("message", "blah")
		)

		span, ctx := opentracing.StartSpanFromContext(ctx, "span")
		span.FinishWithOptions(opentracing.FinishOptions{
			LogRecords: []opentracing.LogRecord{
				{
					Timestamp: time.Now(),
					Fields:    []log.Field{field},
				},
			},
		})

		if expected := 1; len(logger.logs) != expected {
			t.Fatalf("expected %v; got %v", expected, len(logger.logs))
		}
		if expected := []log.Field{field}; !reflect.DeepEqual(expected, logger.logs[0]) {
			t.Errorf("expected %v; got %v", expected, logger.logs[0])
		}
	})
}

func TestDeprecated(t *testing.T) {
	var tracer = New(nil)
	opentracing.SetGlobalTracer(tracer)

	// just invoke the deprecated methods to ensure nothing blows up,
	// we don't really care what happens
	span := opentracing.StartSpan("span")
	span.Log(opentracing.LogData{})
	span.LogEvent("blah")
	span.LogEventWithPayload("blah", nil)
}
