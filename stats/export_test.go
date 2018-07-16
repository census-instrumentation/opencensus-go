package stats

import (
	"context"
	"testing"

	"reflect"

	"go.opencensus.io/tag"
)

type Capture struct {
	Data []Data
}

func (c *Capture) ExportStats(data Data) {
	c.Data = append(c.Data, data)
}

func TestExporter(t *testing.T) {
	name := "blah"
	value := "blah"
	key, _ := tag.NewKey(name)
	ctx, _ := tag.New(context.Background(), tag.Insert(key, value))

	e := &Capture{}
	RegisterExporter(e)
	defer UnregisterExporter(e)

	measure := Int64("name", "description", "kg")
	m := measure.M(123)
	Record(ctx, m)

	if got := len(e.Data); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
	if got := e.Data[0].Measurements; !reflect.DeepEqual(got, []Measurement{m}) {
		t.Errorf("got %v, want %v", got, []Measurement{m})
	}
	if got, ok := e.Data[0].Tags.Value(key); !ok || got != value {
		t.Errorf("got %v, want %v", got, value)
	}
}
