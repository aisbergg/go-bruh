//go:build !(bruh.maxstackdepth12 || bruh.maxstackdepth24 || bruh.maxstackdepth48)

package bruh

// MAX_STACK_DEPTH defines the maximum number of stack frames to capture per error.
// If a function call stack exceeds this depth, the excess frames are truncated.
// This is generally not an issue, as the library merges stack traces across the error chain during serialization.
// To ensure full stack trace reconstruction, wrap errors from deeply nested calls to maintain stack frame overlap.
const MAX_STACK_DEPTH = 32 //nolint:revive
