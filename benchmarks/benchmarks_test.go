package benchmarks_test

import (
	"fmt"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
	pkgerrors "github.com/pkg/errors"
	"github.com/rotisserie/eris"

	emperror "emperror.dev/errors"
)

var (
	rootErrorMessage = "error"
	wrapErrorMessage = "wrap"
)

var (
	global any
	cases  = []struct{ layers int }{{1}, {10}, {100}}
)

func wrapBruh(layers int) error {
	err := bruh.New(rootErrorMessage)
	for i := 0; i < layers-1; i++ {
		err = bruh.Wrap(err, wrapErrorMessage)
	}
	return err
}

func wrapPkgErrors(layers int) error {
	err := pkgerrors.New(rootErrorMessage)
	for i := 0; i < layers-1; i++ {
		err = pkgerrors.Wrap(err, wrapErrorMessage)
	}
	return err
}

func wrapEris(layers int) error {
	err := eris.New(rootErrorMessage)
	for i := 0; i < layers-1; i++ {
		err = eris.Wrap(err, wrapErrorMessage)
	}
	return err
}

func wrapEmperror(layers int) error {
	err := emperror.New(rootErrorMessage)
	for i := 0; i < layers-1; i++ {
		err = emperror.Wrap(err, wrapErrorMessage)
	}
	return err
}

func BenchmarkCompareWrap(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("pkg=bruh/layers=%v", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapBruh(tc.layers)
			}
			b.StopTimer()
			global = err
		})

		b.Run(fmt.Sprintf("pkg=pkgerrors/layers=%v", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapPkgErrors(tc.layers)
			}
			b.StopTimer()
			global = err
		})

		b.Run(fmt.Sprintf("pkg=eris/layers=%v", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapEris(tc.layers)
			}
			b.StopTimer()
			global = err
		})

		b.Run(fmt.Sprintf("pkg=emperror/layers=%v", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapEmperror(tc.layers)
			}
			b.StopTimer()
			global = err
		})
	}
}

func BenchmarkCompareFormatMessageOnly(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("pkg=bruh/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapBruh(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = bruh.Message(err)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("pkg=pkgerrors/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapPkgErrors(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprint(err)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("pkg=eris/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapEris(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = eris.ToString(err, false)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("pkg=emperror/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapEmperror(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprintf("%s", err)
			}
			b.StopTimer()
			global = str
		})
	}
}

func BenchmarkCompareFormatTrace(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("pkg=bruh/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapBruh(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = bruh.StringFormat(err, bruh.BruhFormatter)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("pkg=bruh-stacked/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapBruh(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = bruh.StringFormat(err, bruh.BruhStackedFormatter)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("pkg=pkgerrors/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapPkgErrors(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprintf("%+v", err)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("pkg=eris/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapEris(tc.layers)
			format := eris.StringFormat{
				Options: eris.FormatOptions{
					WithTrace:    true,
					WithExternal: true,
				},
				MsgStackSep:  "\n",
				PreStackSep:  "    ",
				StackElemSep: ":",
				ErrorSep:     "\n",
			}
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = eris.ToCustomString(err, format)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("pkg=emperror/layers=%v", tc.layers), func(b *testing.B) {
			err := wrapEmperror(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprintf("%+v", err)
			}
			b.StopTimer()
			global = str
		})
	}
}

func BenchmarkMessage(b *testing.B) {
	err := wrapBruh(20)
	var str string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		str = bruh.Message(err)
	}
	b.StopTimer()
	global = str
}

func BenchmarkFormatBruh(b *testing.B) {
	err := wrapBruh(20)
	var str string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		str = bruh.StringFormat(err, bruh.BruhFormatter)
	}
	b.StopTimer()
	global = str
}

func BenchmarkFormatBruhStacked(b *testing.B) {
	err := wrapBruh(20)
	var str string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		str = bruh.StringFormat(err, bruh.BruhStackedFormatter)
	}
	b.StopTimer()
	global = str
}
