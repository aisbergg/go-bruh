package ctxerror

import (
	"fmt"
	"testing"
)

var (
	benchContextSink Context
	benchTagsSink    Tags
)

func buildLinearChain(b *testing.B, depth int, shared bool) error {
	if depth < 1 {
		depth = 1
	}

	context := make(map[string]any, depth*2)
	context["k0"] = 0
	cur := New("node-0").
		SetContext("req", context).
		SetTag("k0", "0")

	for i := 1; i < depth; i++ {
		msg := fmt.Sprintf("node-%d", i)
		cur = Wrap(cur, msg).
			SetContext("req", map[string]any{fmt.Sprintf("k%d", i): i}).
			SetTag(fmt.Sprintf("k%d", i), fmt.Sprintf("%d", i))
	}

	return cur
}

func BenchmarkGetContextAndTags(b *testing.B) {
	err := buildLinearChain(b, 10, false)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		benchContextSink = GetContext(err)
		benchTagsSink = GetTags(err)
	}
}

func BenchmarkGeneralUsage(b *testing.B) {
	for b.Loop() {
		base := New("base").
			SetContext("req", map[string]any{"id": "1"}).
			SetTag("op", "base").
			SetTag("status", "500")
		wrapped := Wrap(base, "left").
			SetContext("req", map[string]any{"side": "left"})
		benchContextSink = GetContext(wrapped)
	}
}
