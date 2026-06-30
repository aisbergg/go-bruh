package bruh_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestUnpack(t *testing.T) {
	t.Parallel()
	assertUnpack := func(name string, err error, unpackAll bool, exp bruh.UnpackedError) {
		t.Run(name, func(t *testing.T) {
			formatter := func(b []byte, unpacker *bruh.Unpacker) []byte {
				upkErr := unpacker.Unpack()
				if !isUnpackedErrorEqual(exp, upkErr) {
					expMsgs := make([]string, 0, len(exp))
					for _, msg := range exp {
						expMsgs = append(expMsgs, msg.Msg)
					}
					actMsgs := make([]string, 0, len(upkErr))
					for _, msg := range upkErr {
						actMsgs = append(actMsgs, msg.Msg)
					}
					t.Errorf("expected %#v, got %#v", expMsgs, actMsgs)
				}
				return b
			}

			bruh.StringFormat(err, formatter, unpackAll)
		})
	}

	assertUnpack("Nil", nil, false, bruh.UnpackedError{})
	assertUnpack("Single", bruh.New("root error"), false, bruh.UnpackedError{{Msg: "root error"}})
	assertUnpack("SingleGlobal", globalErr, false, bruh.UnpackedError{{Msg: "global error"}})
	assertUnpack(
		"WrappedNil",
		bruh.Wrap(bruh.Wrap(nil, "additional context"), "even more context"),
		false,
		bruh.UnpackedError{},
	)
	assertUnpack(
		"Wrapped",
		bruh.Wrap(bruh.Wrap(bruh.New("root error"), "additional context"), "even more context"),
		false,
		bruh.UnpackedError{{Msg: "even more context"}, {Msg: "additional context"}, {Msg: "root error"}},
	)
	assertUnpack(
		"WrappedGlobal",
		bruh.Wrap(bruh.Wrap(globalErr, "additional context"), "even more context"),
		false,
		bruh.UnpackedError{{Msg: "even more context"}, {Msg: "additional context"}, {Msg: "global error"}},
	)
	assertUnpack(
		"WrappedExternalError",
		bruh.Wrap(fmt.Errorf("some external error: %w", errors.New("root cause")), "external wrapper"),
		false,
		bruh.UnpackedError{{Msg: "external wrapper"}, {Msg: "some external error: root cause"}},
	)
	assertUnpack(
		"WrappedExternalErrorUnpackAll",
		bruh.Wrap(fmt.Errorf("some external error: %w", errors.New("root cause")), "external wrapper"),
		true,
		bruh.UnpackedError{{Msg: "external wrapper"}, {Msg: "some external error"}, {Msg: "root cause"}},
	)
	assertUnpack(
		"WrappedMixedError",
		bruh.Wrap(
			fmt.Errorf("some external error: %w", bruh.Wrap(errors.New("root cause"), "additional context")),
			"even more context",
		),
		true,
		bruh.UnpackedError{
			{Msg: "even more context"},
			{Msg: "some external error"},
			{Msg: "additional context"},
			{Msg: "root cause"},
		},
	)
}

func isUnpackedErrorEqual(a, b []bruh.UnpackedElement) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Msg != b[i].Msg {
			return false
		}
	}

	return true
}
