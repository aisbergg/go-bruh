package bruh_test

import (
	"testing"
	"unsafe"

	fuzz "github.com/AdaLogics/go-fuzz-headers"
	"github.com/aisbergg/go-bruh/pkg/bruh"
)

type Unpacker struct {
	err       error
	upkErr    *bruh.UnpackedError
	cbdStk    *bruh.Stack
	chainLen  int
	unpackAll bool
}

func generateUnpacker(t *testing.T, data []byte) *bruh.Unpacker {
	fz := fuzz.NewConsumer(data)
	var upkErr bruh.UnpackedError
	err := fz.CreateSlice(&upkErr)
	if err != nil {
		t.Skip()
	}
	cbdStk := combinedStack(upkErr)
	unpacker := &Unpacker{
		err:       err,
		upkErr:    &upkErr,
		cbdStk:    &cbdStk,
		chainLen:  len(upkErr),
		unpackAll: false,
	}
	return (*bruh.Unpacker)(unsafe.Pointer(unpacker))
}

func combinedStack(upkErr bruh.UnpackedError) bruh.Stack {
	if len(upkErr) == 0 {
		return bruh.Stack{}
	}
	numFrames := 0
	for i := 0; i < len(upkErr); i++ {
		numFrames += len(upkErr[i].PartialStack)
	}
	combinedStack := make(bruh.Stack, 0, numFrames)
	for i := len(upkErr) - 1; i >= 0; i-- {
		combinedStack = append(combinedStack, upkErr[i].PartialStack...)
	}
	return combinedStack
}

func FuzzFormatBruh(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte, colored, sourced bool) {
		unpacker := generateUnpacker(t, data)
		_ = bruh.BruhFancyFormatter(colored, sourced)(nil, unpacker)
	})
}

func FuzzFormatBruhStacked(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte, colored, sourced, typed bool) {
		unpacker := generateUnpacker(t, data)
		_ = bruh.BruhStackedFancyFormatter(colored, sourced, typed)(nil, unpacker)
	})
}

func FuzzFormatGoPanic(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		unpacker := generateUnpacker(t, data)
		_ = bruh.GoPanicFormatter(nil, unpacker)
	})
}

func FuzzFormatPythonTraceback(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		unpacker := generateUnpacker(t, data)
		_ = bruh.BruhFormatter(nil, unpacker)
	})
}

func FuzzFormatJavaStackTrace(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		unpacker := generateUnpacker(t, data)
		_ = bruh.BruhFormatter(nil, unpacker)
	})
}
