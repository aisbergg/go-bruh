package main

import (
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/ctxerror"
)

func main() {
	err := saveOrder("ord_42", 0)
	if err != nil {
		fmt.Println(err.Error())

		tags := ctxerror.GetTags(err)
		ctx := ctxerror.GetContext(err)

		for k, v := range tags {
			fmt.Printf("tag %s=%s\n", k, v)
		}
		for k, v := range ctx {
			fmt.Printf("context %s=%v\n", k, v)
		}
	}

	// Custom errors can implement Context and Tags methods to provide context and tags to ctxerror. This allows you to integrate with existing error types without wrapping them in *ctxerror.Err.
	customErr := &CustomError1{msg: "custom error occurred"}
	err = ctxerror.Wrap(customErr, "operation failed").SetTag("op", "custom_op")
	fmt.Println(err.Error())
	fmt.Printf("tags: %v\n", ctxerror.GetTags(err))
	fmt.Printf("context: %v\n", ctxerror.GetContext(err))

	// Alternatively, custom errors can implement AppendContext and AppendTags for append-style population of context and tags.
	customErr2 := &CustomError2{msg: "another custom error"}
	err = ctxerror.Wrap(customErr2, "operation failed").SetTag("op", "custom_op_2")
	fmt.Println(err.Error())
	fmt.Printf("tags: %v\n", ctxerror.GetTags(err))
	fmt.Printf("context: %v\n", ctxerror.GetContext(err))
}

func saveOrder(orderID string, itemCount int) error {
	if err := validateOrder(itemCount); err != nil {
		return ctxerror.Wrap(err, "saving order").
			SetTag("op", "save_order").
			SetContext("order", map[string]any{"id": orderID, "items": itemCount})
	}
	return nil
}

func validateOrder(itemCount int) error {
	if itemCount <= 0 {
		return ctxerror.Wrap(GlobalErr, "validating order").
			SetContext("validation", map[string]any{"itemCount": itemCount, "min": 1})
	}
	return nil
}

// Global errors or errors that are shared among multiple error chains need to
// be unshared, which creates a private metadata copy for upstream errors.
var GlobalErr = ctxerror.New("global error").SetTag("locale", "de").Unshare()

// Custom Error with context ---------------------------------------------------

type CustomError1 struct {
	msg string
}

func (ce *CustomError1) Error() string {
	return ce.msg
}

// Context allows `ctxerror` to extract context from this error type.
func (ce *CustomError1) Context() map[string]map[string]any {
	return map[string]map[string]any{
		"req": {
			"trace_id": 1,
		},
	}
}

// Tags allows `ctxerror` to extract tags from this error type.
func (ce *CustomError1) Tags() map[string]string {
	return map[string]string{
		"kind": "custom",
	}
}

type CustomError2 struct {
	msg string
}

func (ce *CustomError2) Error() string {
	return ce.msg
}

// AppendContext implements append-style population of caller-provided context map.
func (ce *CustomError2) AppendContext(context map[string]map[string]any) {
	context["req"] = map[string]any{"trace_id": 1}
}

// AppendTags implements append-style population of caller-provided tags map.
func (ce *CustomError2) AppendTags(tags map[string]string) {
	tags["kind"] = "custom"
}
