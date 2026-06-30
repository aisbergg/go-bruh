package multierror_test

import (
	"errors"
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/multierror"
)

type exampleCustomErr struct{}

func (exampleCustomErr) Error() string { return "custom" }

func ExampleNew() {
	me := multierror.New("batch import")
	fmt.Println(me.IsNil())

	// Output:
	// true
}

func ExampleErrorf() {
	me := multierror.Errorf(multierror.Options{}, "batch %d", 42)
	fmt.Println(me.IsNil())

	// Output:
	// true
}

func ExampleOptions_unwrapBehavior() {
	me := multierror.New("batch", multierror.Options{UnwrapBehavior: multierror.UnwrapLast})
	me.Add(errors.New("first"), errors.New("last"))
	fmt.Println(me.Unwrap())

	// Output:
	// last
}

func ExampleOptions_limitPrint() {
	me := multierror.New("batch", multierror.Options{LimitPrint: 2})
	me.Add(
		errors.New("error1"),
		errors.New("error2"),
		errors.New("error3"),
		errors.New("error4"),
		errors.New("error5"),
	)
	fmt.Println(me.Error())

	// Output:
	// batch
	//   #0: error1
	//   #1: error2
	//   ... and 3 more errors
}

func ExampleOptions_filter() {
	me := multierror.New("batch", multierror.Options{Filter: func(err error) bool {
		return err.Error() != "skip"
	}})
	me.Add(errors.New("keep"), errors.New("skip"), errors.New("keep2"))
	fmt.Println(len(me.Errors()))

	// Output:
	// 2
}

func ExampleErr_Grow() {
	me := multierror.New("batch")
	me.Grow(4)
	fmt.Println(cap(me.Errors()) >= 4)

	// Output:
	// true
}

func ExampleErr_Errors() {
	me := multierror.New("batch")
	me.Add(errors.New("e1"), errors.New("e2"))
	fmt.Println(len(me.Errors()))

	// Output:
	// 2
}

func ExampleErr_Error() {
	me := multierror.New("batch")
	me.Add(errors.New("e1"))
	fmt.Println(me.Error())

	// Output:
	// batch
	//   #0: e1
}

func ExampleErr_Message() {
	me := multierror.New("batch")
	me.Add(errors.New("e1"))
	fmt.Println(me.Message())

	// Output:
	// batch
	//   #0: e1
}

func ExampleErr_ErrorOrNil() {
	empty := multierror.New("batch")
	fmt.Println(empty.ErrorOrNil() == nil)

	nonEmpty := multierror.New("batch")
	nonEmpty.Add(errors.New("e1"))
	fmt.Println(nonEmpty.ErrorOrNil() != nil)

	// Output:
	// true
	// true
}

func ExampleErr_SingleOrNil() {
	empty := multierror.New("batch")
	fmt.Println(empty.SingleOrNil() == nil)

	one := multierror.New("batch")
	one.Add(errors.New("e1"))
	fmt.Println(one.SingleOrNil())

	many := multierror.New("batch")
	many.Add(errors.New("e1"), errors.New("e2"))
	fmt.Println(many.SingleOrNil() == many)

	// Output:
	// true
	// e1
	// true
}

func ExampleErr_Unwrap() {
	me := multierror.New("batch", multierror.Options{UnwrapBehavior: multierror.UnwrapFirst})
	me.Add(errors.New("e1"), errors.New("e2"))
	fmt.Println(me.Unwrap())

	// Output:
	// e1
}

func ExampleErr_IsNil() {
	empty := multierror.New("batch")
	fmt.Println(empty.IsNil())

	empty.Add(errors.New("e1"))
	fmt.Println(empty.IsNil())

	// Output:
	// true
	// false
}

func ExampleErr_Add() {
	me := multierror.New("batch")
	me.Add(nil, errors.New("e1"), nil, errors.New("e2"))
	fmt.Println(len(me.Errors()))

	// Output:
	// 2
}

func ExampleErr_Merge() {
	left := multierror.New("left")
	left.Add(errors.New("a"))

	right := multierror.New("right")
	right.Add(errors.New("b"), errors.New("c"))

	left.Merge(right)
	fmt.Println(len(left.Errors()))

	// Output:
	// 3
}
