package metric

import (
	"testing"
)

func TestRegistry_AddProducer(t *testing.T) {
	r := NewRegistry()
	m1 := &Metric{
		Descriptor: &Descriptor{
			Name: "test",
			Unit: UnitDimensionless,
		},
	}
	remove := r.AddProducer(&constProducer{m1})
	if got, want := len(r.Read()), 1; got != want {
		t.Fatal("Expected to read a single metric")
	}
	remove()
	if got, want := len(r.Read()), 0; got != want {
		t.Fatal("Expected to read no metrics")
	}
}

type constProducer []*Metric

func (cp constProducer) Read() []*Metric {
	return cp
}
