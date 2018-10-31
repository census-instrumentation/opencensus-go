package resource

import (
	"context"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := context.Background()
	res := &Resource{
		Type: "t1",
		Labels: map[string]string{
			"l1": "v1",
		},
	}
	ctx = NewContext(ctx, res)

	ctxRes, ok := FromContext(ctx)

	if !ok {
		t.Fatalf("Expected resource")
	}

	if got, want := ctxRes.Type, "t1"; got != want {
		t.Fatalf("ctxRes.Type = %s; want %s", got, want)
	}
}
