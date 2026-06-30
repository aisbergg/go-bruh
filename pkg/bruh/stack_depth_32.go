//go:build bruh.maxerrorstackdepth32

package bruh

// MaxErrorStackDepth defines the maximum number of stack frames to capture per error.
// If a function call stack exceeds this depth, the excess frames are truncated.
// This is generally not an issue, as the library merges stack traces across the error chain during serialization.
// To ensure full stack trace reconstruction, wrap errors from deeply nested calls to maintain stack frame overlap.
const MaxErrorStackDepth = 32 //nolint:revive
