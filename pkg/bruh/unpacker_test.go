package bruh_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// XXX: test other fields as well
func TestUnpack(t *testing.T) {
	// t.Parallel()
	tests := []struct {
		name      string
		err       error
		unpackAll bool
		exp       bruh.UnpackedError
	}{
		{
			name: "Nil",
			err:  nil,
			exp:  bruh.UnpackedError{},
		},
		{
			name: "Single",
			err:  bruh.New("root error"),
			exp: bruh.UnpackedError{
				{
					Msg: "root error",
				},
			},
		},
		{
			name: "SingleGlobal",
			err:  globalErr,
			exp: bruh.UnpackedError{
				{
					Msg: "global error",
				},
			},
		},
		{
			name: "WrappedNil",
			err:  bruh.Wrap(bruh.Wrap(nil, "additional context"), "even more context"),
			exp:  bruh.UnpackedError{},
		},
		{
			name: "Wrapped",
			err:  bruh.Wrap(bruh.Wrap(bruh.New("root error"), "additional context"), "even more context"),
			exp: bruh.UnpackedError{
				{
					Msg: "even more context",
				},
				{
					Msg: "additional context",
				},
				{
					Msg: "root error",
				},
			},
		},
		{
			name: "WrappedGlobal",
			err:  bruh.Wrap(bruh.Wrap(globalErr, "additional context"), "even more context"),
			exp: bruh.UnpackedError{
				{
					Msg: "even more context",
				},
				{
					Msg: "additional context",
				},
				{
					Msg: "global error",
				},
			},
		},
		{
			name: "WrappedExternalError",
			err:  bruh.Wrap(fmt.Errorf("some external error: %w", errors.New("root cause")), "external wrapper"),
			exp: bruh.UnpackedError{
				{
					Msg: "external wrapper",
				},
				{
					Msg: "some external error: root cause",
				},
			},
		},
		{
			name:      "WrappedExternalErrorUnpackAll",
			err:       bruh.Wrap(fmt.Errorf("some external error: %w", errors.New("root cause")), "external wrapper"),
			unpackAll: true,
			exp: bruh.UnpackedError{
				{
					Msg: "external wrapper",
				},
				{
					Msg: "some external error",
				},
				{
					Msg: "root cause",
				},
			},
		},
		{
			name:      "WrappedMixedError",
			err:       bruh.Wrap(fmt.Errorf("some external error: %w", bruh.Wrap(errors.New("root cause"), "additional context")), "even more context"),
			unpackAll: true,
			exp: bruh.UnpackedError{
				{
					Msg: "even more context",
				},
				{
					Msg: "some external error",
				},
				{
					Msg: "additional context",
				},
				{
					Msg: "root cause",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatter := func(b []byte, unpacker *bruh.Unpacker) []byte {
				upkErr := unpacker.Unpack()
				if !isUnpackedErrorEqual(tc.exp, upkErr) {
					expMsgs := make([]string, 0, len(tc.exp))
					for _, msg := range tc.exp {
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

			bruh.StringFormat(tc.err, formatter, tc.unpackAll)
			// upk := bruh.Unpack(tc.err, tc.unpackAll)
			// // assert.Equal(tt.exp, upk)
			// if !isUnpackedErrorEqual(tc.exp, upk) {
			// 	expMsgs := make([]string, 0, len(tc.exp))
			// 	for _, msg := range tc.exp {
			// 		expMsgs = append(expMsgs, msg.Msg)
			// 	}
			// 	actMsgs := make([]string, 0, len(upk))
			// 	for _, msg := range upk {
			// 		actMsgs = append(actMsgs, msg.Msg)
			// 	}
			// 	t.Errorf("expected %#v, got %#v", expMsgs, actMsgs)
			// }
		})
	}
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
