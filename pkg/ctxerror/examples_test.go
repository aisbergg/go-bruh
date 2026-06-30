package ctxerror_test

import (
	"errors"
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/ctxerror"
)

func ExampleNew() {
	err := ctxerror.New("open file").
		SetTag("op", "open").
		SetContext("file", map[string]any{"path": "report.csv"})

	fmt.Println(err.Error())
	fmt.Println(ctxerror.GetTags(err)["op"])
	fmt.Println(ctxerror.GetContext(err)["file"]["path"])

	// Output:
	// open file
	// open
	// report.csv
}

func ExampleErrorf() {
	err := ctxerror.Errorf("open %s", "report.csv").
		SetTag("op", "open").
		SetContext("file", map[string]any{"path": "report.csv"})

	fmt.Println(err.Error())
	fmt.Println(ctxerror.GetTags(err)["op"])
	fmt.Println(ctxerror.GetContext(err)["file"]["path"])

	// Output:
	// open report.csv
	// open
	// report.csv
}

func ExampleWrap() {
	err := ctxerror.Wrap(errors.New("permission denied"), "open file").
		SetTag("op", "open").
		SetContext("file", map[string]any{"path": "report.csv"})
	fmt.Println(err.Error())
	fmt.Println(ctxerror.GetTags(err)["op"])
	fmt.Println(ctxerror.GetContext(err)["file"]["path"])

	// Output:
	// open file: permission denied
	// open
	// report.csv
}

func ExampleWrapf() {
	err := ctxerror.Wrapf(errors.New("permission denied"), "open %s", "report.csv").
		SetTag("op", "open").
		SetContext("file", map[string]any{"path": "report.csv"})
	fmt.Println(err.Error())
	fmt.Println(ctxerror.GetTags(err)["op"])
	fmt.Println(ctxerror.GetContext(err)["file"]["path"])

	// Output:
	// open report.csv: permission denied
	// open
	// report.csv
}

func ExampleGetContext() {
	base := ctxerror.New("db error").
		SetContext("req", map[string]any{"id": "r1"})
	wrapped := ctxerror.Wrap(base, "service error").
		SetContext("user", map[string]any{"id": "u1"})

	ctx := ctxerror.GetContext(wrapped)
	fmt.Println(ctx["req"]["id"], ctx["user"]["id"])

	// Output:
	// r1 u1
}

func ExampleGetTags() {
	base := ctxerror.New("db error").
		SetTag("env", "prod")
	wrapped := ctxerror.Wrap(base, "service error").
		SetTag("region", "eu")

	tags := ctxerror.GetTags(wrapped)
	fmt.Println(tags["env"], tags["region"])

	// Output:
	// prod eu
}

func ExampleErr_SetContext() {
	err := ctxerror.New("request failed").
		SetContext("req", map[string]any{"id": "r1"}).
		SetContext("user", map[string]any{"id": "u1"})
	fmt.Println(ctxerror.GetContext(err)["req"]["id"], ctxerror.GetContext(err)["user"]["id"])

	// Output:
	// r1 u1
}

func ExampleErr_SetContexts() {
	err := ctxerror.New("request failed").SetContexts(ctxerror.Context{
		"req":  {"id": "r1"},
		"user": {"id": "u1"},
	})
	context := ctxerror.GetContext(err)
	fmt.Println(context["req"]["id"], context["user"]["id"])

	// Output:
	// r1 u1
}

func ExampleErr_SetTag() {
	err := ctxerror.New("request failed").
		SetTag("env", "prod").
		SetTag("region", "eu")
	fmt.Println(ctxerror.GetTags(err)["env"], ctxerror.GetTags(err)["region"])

	// Output:
	// prod eu
}

func ExampleErr_SetTags() {
	err := ctxerror.New("request failed").
		SetTags(ctxerror.Tags{"env": "prod", "region": "eu"})
	tags := ctxerror.GetTags(err)
	fmt.Println(tags["env"], tags["region"])

	// or with pre-allocated map, so later added tags don't cause map allocations
	tagsMap := make(ctxerror.Tags, 12)
	tagsMap["env"] = "prod"
	err = ctxerror.New("request failed").
		SetTags(tagsMap)
	err = ctxerror.Wrap(err, "another error").
		SetTag("region", "eu")
	tags = ctxerror.GetTags(err)
	fmt.Println(tags["env"], tags["region"])

	// Output:
	// prod eu
	// prod eu
}

func ExampleErr_Unshare() {
	base := ctxerror.New("db").
		SetContext("req", map[string]any{"path": "/v1"}).
		Unshare()

	fmt.Println(ctxerror.GetContext(base)["req"]["path"]) // -> /v1

	// because of the call to `Unshare` wrapped has its own private copy of context
	wrapped := ctxerror.Wrap(base, "left").SetContext("req", map[string]any{"path": "/v2"})

	fmt.Println(ctxerror.GetContext(wrapped)["req"]["path"]) // -> /v2
	fmt.Println(ctxerror.GetContext(base)["req"]["path"])    // -> /v1

	// Output:
	// /v1
	// /v2
	// /v1
}
