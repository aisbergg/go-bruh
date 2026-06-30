module github.com/aisbergg/go-bruh

go 1.25.0

// The standard Bruh package is completely third-party dependency free. The OTEL
// dependency will only be pulled and included in your project, if you import the
// helper package "github.com/aisbergg/go-bruh/pkg/ctxerror/ctxotel", Go's
// module graph pruning (https://go.dev/ref/mod#graph-pruning) makes sure of
// that.
require go.opentelemetry.io/otel v1.44.0

require github.com/cespare/xxhash/v2 v2.3.0 // indirect
