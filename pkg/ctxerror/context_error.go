// Package ctxerror provides errors that can carry structured metadata (context and
// tags). The data model matches Sentry's contexts/tags and can be uploaded to
// Sentry.
//
// Use context for richer, grouped data (e.g. request info, payload fragments)
// and tags for small string key/value attributes (e.g. operation, region).
//
// By default, metadata is shared across an error chain to minimize allocations.
// For errors that are reused (e.g. global errors), call (*Err).Unshare() to
// ensure they maintain a private copy of their metadata.
package ctxerror

import (
	"fmt"
	"maps"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// ContextDefaultMapSize is the default size for context maps. It is used to
// avoid unnecessary allocations when adding context to an error. You can change
// this value if you expect to have more or less context groups and tags.
var ContextDefaultMapSize = 12

// Context holds metadata attached to an error chain. Use Context for richer,
// grouped data (for example request information or payload fragments). Context
// is intended as easy extra data for structured logging and error catchers
// (e.g. Sentry).
type Context = map[string]map[string]any

// Tags is a small set of string key/value attributes attached to an error
// chain. Use Tags for lightweight attributes with low cardinality (for example
// `operation` or `region`). Tags are intended for structured logging and error
// catchers (e.g. Sentry).
type Tags = map[string]string

// ModifiableContextErr is an error that can modify its context.
type ModifiableContextErr interface {
	error
	SetContext(string, map[string]any) ModifiableContextErr
	SetContexts(Context) ModifiableContextErr
	SetTag(string, string) ModifiableContextErr
	SetTags(Tags) ModifiableContextErr
	Unshare() ModifiableContextErr
}

var _ ModifiableContextErr = (*Err)(nil)

// Err is an error type that allows attaching structured metadata, known as
// context and tags. This is useful for adding supplementary information to
// errors, which can then be retrieved for logging or sent to error tracking
// services like Sentry.
//
// The attached context and tags are not included in the error message returned
// by the Error() method. They can be accessed using the [GetContext] and
// [GetTags] functions.
//
// By default, Err uses a shared metadata model. This means that metadata is
// shared across a single error chain to optimize performance by reducing
// allocations. When you add context or tags to an error, that metadata becomes
// part of a common pool for the entire chain.
//
// If you intend to reuse an error (for example, a global error variable) across
// different error chains, you should call Unshare() on it. This creates a
// private copy of the metadata, preventing modifications in one chain from
// affecting another.
type Err struct {
	bruh.Err
	context Context
	tags    Tags

	// shared indicates whether this error uses shared metadata maps found
	// earlier in the chain. Default is true. When set to false the error will
	// create a private deep-copy of the first context/tags it can find when the
	// private accessor is called.
	shared bool
}

// New creates a new context-aware error [Err] with the provided message and
// stack trace. Shared metadata is the default.
func New(msg string) ModifiableContextErr {
	return &Err{Err: *bruh.NewSkip(1, msg), shared: true}
}

// NewSkip behaves like [New] but skips the given number of callers when
// creating a stack trace. You should only use this if you are implementing a
// new error type on top of [Err].

// Errorf creates a new context-aware error [Err] with a formatted message.
func Errorf(format string, args ...any) ModifiableContextErr {
	return &Err{Err: *bruh.NewSkip(1, fmt.Sprintf(format, args...)), shared: true}
}

// Wrap wraps the given error by creating new context-aware error [Err] with
// the provided message and stack trace. Shared metadata is the default.
func Wrap(err error, msg string) ModifiableContextErr {
	if err == nil {
		return nil
	}
	return &Err{Err: *bruh.WrapSkip(err, 1, msg), shared: true}
}

// Wrapf wraps the given error by creating a new context-aware error [Err] with
// a formatted message and stack trace. Shared metadata is the default.
func Wrapf(err error, format string, args ...any) ModifiableContextErr {
	if err == nil {
		return nil
	}
	return &Err{Err: *bruh.WrapSkip(err, 1, fmt.Sprintf(format, args...)), shared: true}
}

// privateContext returns the context of the error without merging it with the
// context of the wrapped errors. It is used internally to implement the
// [Context] method and should not be used directly. If no context was set, nil
// is returned.
func (e *Err) privateContext() Context {
	// Ensure shared/common context map is initialized.
	ctx := e.initContext()
	if !e.shared {
		if ctx == nil {
			ctx = make(Context, ContextDefaultMapSize)
			return ctx
		}
		nc := make(Context, len(ctx))
		for g, m := range ctx {
			nm := make(map[string]any, len(m))
			maps.Copy(nm, m)
			nc[g] = nm
		}
		return nc
	}
	return ctx
}

func (e *Err) privateTags() Tags {
	tags := e.initTags()
	if !e.shared {
		if tags == nil {
			tags = make(Tags, ContextDefaultMapSize)
			return tags
		}
		tagsCopy := make(Tags, len(tags))
		maps.Copy(tagsCopy, tags)
		return tagsCopy
	}
	return tags
}

// SetContext adds the given key-value pairs to the context of the error.
func (e *Err) SetContext(key string, value map[string]any) ModifiableContextErr {
	if e == nil {
		return nil
	}
	ctx := e.initContext()
	if ctx == nil {
		ctx = make(Context, ContextDefaultMapSize)
		e.context = ctx
	}
	// group does not exist, add the whole group
	if _, groupExists := ctx[key]; !groupExists {
		ctx[key] = value
		return e
	}
	// merge same group context
	maps.Copy(ctx[key], value)
	return e
}

// SetContexts adds all key-value pairs from `context` to the error context.
func (e *Err) SetContexts(context Context) ModifiableContextErr {
	if e == nil {
		return nil
	}
	ctx := e.initContext()
	if ctx == nil {
		e.context = context
		return e
	}
	for key, value := range context {
		if _, groupExists := ctx[key]; !groupExists {
			ctx[key] = value
			continue
		}
		maps.Copy(ctx[key], value)
	}
	return e
}

// SetTag adds a single key-value pair to the error tags.
func (e *Err) SetTag(key, value string) ModifiableContextErr {
	if e == nil {
		return nil
	}
	tags := e.initTags()
	if tags == nil {
		tags = make(Tags, ContextDefaultMapSize)
		e.tags = tags
	}
	tags[key] = value
	return e
}

// SetTags adds all key-value pairs from tags to the error tags.
func (e *Err) SetTags(tags Tags) ModifiableContextErr {
	if e == nil {
		return nil
	}
	existingTags := e.initTags()
	if existingTags == nil {
		e.tags = tags
		return e
	}
	maps.Copy(existingTags, tags)
	return e
}

// Unshare makes the error keep its own private copies of context and tags.
// Call this when you plan to make an error global or reuse it across
// independent wrappers. It deep-copies the maps and marks the error as
// unshared (shared=false).
func (e *Err) Unshare() ModifiableContextErr {
	if e == nil {
		return nil
	}
	// mark as unshared; private deep-copy will occur lazily on first access
	// through privateContext/privateTags.
	e.shared = false
	return e
}

// initContext initializes the context of the error. It tries to get the
// common context map instance and merge any other context maps found in the
// chain into it. Mirrors initTags behaviour for tags.
func (e *Err) initContext() Context {
	if e == nil {
		return nil
	}
	ctx := e.context
	if ctx != nil {
		return ctx
	}

	// locate first initialized context in wrapped chain
	depthToUnwrap := 0
	allocationRequired := false
	for err := bruh.Unwrap(e); err != nil; err = bruh.Unwrap(err) {
		if cerr, ok := err.(privateContexter); ok {
			pct := cerr.privateContext()
			if pct != nil {
				ctx = pct
				break
			}
		} else if _, ok := err.(contexter); ok {
			allocationRequired = true
		}
		depthToUnwrap++
	}
	if ctx == nil {
		if !allocationRequired {
			return nil
		}
		ctx = make(Context, ContextDefaultMapSize)
	}
	e.context = ctx

	// merge other context maps into the common one
	for err := bruh.Unwrap(e); err != nil && depthToUnwrap > 0; err = bruh.Unwrap(err) {
		depthToUnwrap--
		if terr, ok := err.(privateContexter); ok {
			tctx := terr.privateContext()
			if tctx == nil {
				continue
			}
			for g, m := range tctx {
				if _, ok := ctx[g]; !ok {
					ctx[g] = m
					continue
				}
				maps.Copy(ctx[g], m)
			}
		} else if terr, ok := err.(contexter); ok {
			innerCtx := terr.Context()
			for g, m := range innerCtx {
				if _, ok := ctx[g]; !ok {
					ctx[g] = m
					continue
				}
				maps.Copy(ctx[g], m)
			}
		}
	}
	return ctx
}

// initTags initializes the tags of the error. It tries to get the common tags
// map instance and merge any other tags maps found in the chain into it.
func (e *Err) initTags() Tags {
	if e == nil {
		return nil
	}
	tags := e.tags
	if tags != nil {
		return tags
	}
	// get initialized tags
	depthToUnwrap := 0
	allocationRequired := false
	for err := bruh.Unwrap(e); err != nil; err = bruh.Unwrap(err) {
		if cerr, ok := err.(privateContexter); ok {
			ptags := cerr.privateTags()
			if ptags != nil {
				tags = ptags
				break
			}
		} else if _, ok := err.(tagsAppender); ok {
			allocationRequired = true
		} else if _, ok := err.(tagser); ok {
			allocationRequired = true
		}
		depthToUnwrap++
	}
	if tags == nil {
		if !allocationRequired {
			return nil
		}
		tags = make(Tags, ContextDefaultMapSize)
	}
	e.tags = tags

	// merge other tags maps into the common one
	for err := bruh.Unwrap(e); err != nil && depthToUnwrap > 0; err = bruh.Unwrap(err) {
		depthToUnwrap--
		if terr, ok := err.(tagsAppender); ok {
			terr.AppendTags(tags)
		} else if terr, ok := err.(tagser); ok {
			innerTags := terr.Tags()
			maps.Copy(tags, innerTags)
		}
	}
	return tags
}

// -----------------------------------------------------------------------------
//
// Convenience Functions
//
// -----------------------------------------------------------------------------

// GetContext returns the context of the given error chain. If there are
// multiple context maps in the chain, they will be merged together. If there
// are duplicate keys, the value of the outermost error will be used. If no
// context was set, an empty context map is returned. Do not modify the returned
// context directly.
func GetContext(err error) Context {
	if err == nil {
		return make(Context)
	}
	var ctx Context

	// find first initialized context
	depthToUnwrap := 0
	allocationRequired := false
	for uerr := err; uerr != nil; uerr = bruh.Unwrap(uerr) {
		if cerr, ok := uerr.(privateContexter); ok {
			pct := cerr.privateContext()
			if pct != nil {
				ctx = pct
				break
			}
		} else if _, ok := uerr.(contextAppender); ok {
			allocationRequired = true
		} else if _, ok := uerr.(contexter); ok {
			allocationRequired = true
		}
		depthToUnwrap++
	}
	if ctx == nil {
		if !allocationRequired {
			return make(Context)
		}
		ctx = make(Context, ContextDefaultMapSize)
	}

	// merge other context maps into the common one
	for uerr := bruh.Unwrap(err); uerr != nil && depthToUnwrap > 0; uerr = bruh.Unwrap(uerr) {
		depthToUnwrap--
		if terr, ok := uerr.(contextAppender); ok {
			terr.AppendContext(ctx)
		} else if terr, ok := uerr.(contexter); ok {
			innerCtx := terr.Context()
			for g, m := range innerCtx {
				if _, ok := ctx[g]; !ok {
					ctx[g] = m
					continue
				}
				maps.Copy(ctx[g], m)
			}
		}
	}
	return ctx
}

// GetTags returns the tags of the given error chain. If there are multiple tags maps in the chain, they will be merged together. If there are duplicate keys,
// the value of the outermost error will be used. If no tags were set, an empty
// tags map is returned. Do not modify the returned tags directly.
func GetTags(err error) Tags {
	if err == nil {
		return make(Tags)
	}
	var tags Tags

	// get initialized tags
	depthToUnwrap := 0
	allocationRequired := false
	for uerr := err; uerr != nil; uerr = bruh.Unwrap(uerr) {
		if cerr, ok := uerr.(privateContexter); ok {
			ptags := cerr.privateTags()
			if ptags != nil {
				tags = ptags
				break
			}
		} else if _, ok := uerr.(tagsAppender); ok {
			allocationRequired = true
		} else if _, ok := uerr.(tagser); ok {
			allocationRequired = true
		}
		depthToUnwrap++
	}
	if tags == nil {
		if !allocationRequired {
			return make(Tags)
		}
		tags = make(Tags, ContextDefaultMapSize)
	}

	// merge other tags maps into the common one
	for uerr := bruh.Unwrap(err); uerr != nil && depthToUnwrap > 0; uerr = bruh.Unwrap(uerr) {
		depthToUnwrap--
		if terr, ok := uerr.(tagsAppender); ok {
			terr.AppendTags(tags)
		} else if terr, ok := uerr.(tagser); ok {
			innerTags := terr.Tags()
			maps.Copy(tags, innerTags)
		}
	}
	return tags
}

type contexter interface {
	Context() Context
}

type contextAppender interface {
	AppendContext(context Context)
}

type tagser interface {
	Tags() Tags
}

type tagsAppender interface {
	AppendTags(tags Tags)
}

type privateContexter interface {
	error
	privateContext() Context
	privateTags() Tags
}
