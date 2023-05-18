package benchmarks_test

import (
	"fmt"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
	pkgerrors "github.com/pkg/errors"
	"github.com/rotisserie/eris"
)

var (
	global any
	cases  = []struct {
		layers int
	}{
		{1},
		{10},
		{100},
	}
)

func wrapPkgErrors(layers int) error {
	err := pkgerrors.New("error")
	for i := 0; i < layers; i++ {
		err = pkgerrors.Wrapf(err, "wrap %v", i)
	}
	return err
}

func wrapEris(layers int) error {
	err := eris.New("error")
	for i := 0; i < layers; i++ {
		err = eris.Wrapf(err, "wrap %v", i)
	}
	return err
}

func wrapBruh(layers int) error {
	err := bruh.New("error")
	for i := 0; i < layers; i++ {
		err = bruh.Wrapf(err, "wrap %v", i)
	}
	return err
}

func BenchmarkWrap(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("pkg errors %v layers", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapPkgErrors(tc.layers)
			}
			b.StopTimer()
			global = err
		})

		b.Run(fmt.Sprintf("eris %v layers", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapEris(tc.layers)
			}
			b.StopTimer()
			global = err
		})

		b.Run(fmt.Sprintf("bruh %v layers", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapBruh(tc.layers)
			}
			b.StopTimer()
			global = err
		})
	}
}

func BenchmarkFormatWithoutTrace(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("pkg errors %v layers", tc.layers), func(b *testing.B) {
			err := wrapPkgErrors(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprint(err)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("eris %v layers", tc.layers), func(b *testing.B) {
			err := wrapEris(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = eris.ToString(err, false)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("bruh %v layers", tc.layers), func(b *testing.B) {
			err := wrapBruh(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = bruh.ToString(err, false)
			}
			b.StopTimer()
			global = str
		})
	}
}

func BenchmarkFormatWithTrace(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("pkg errors %v layers", tc.layers), func(b *testing.B) {
			err := wrapPkgErrors(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprintf("%+v", err)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("eris %v layers", tc.layers), func(b *testing.B) {
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

		b.Run(fmt.Sprintf("bruh %v layers", tc.layers), func(b *testing.B) {
			err := wrapBruh(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = bruh.ToString(err, true)
			}
			b.StopTimer()
			global = str
		})
	}
}
