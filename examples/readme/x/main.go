package main

import (
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	// error creation
	err := bruh.New("unexpected end of file")
	// error wrapping
	err = bruh.Wrapf(err, "reading file '%s'", "example.txt")
	// using a custom error type
	err = WrapUserError(err, "12345", "retrieving user data")

	// error formatting
	fmt.Println(bruh.String(err))
	// using a different format
	fmt.Println(bruh.StringFormat(err, bruh.GoPanicFormatter))

	// testing for specific errors
	var userErr *UserError
	if bruh.As(err, &userErr) {
		fmt.Printf("User ID: %s\n", userErr.UserID)
	}
}

type UserError struct {
	bruh.Err
	UserID string
}

func WrapUserError(err error, userID, message string) *UserError {
	return &UserError{
		Err:    *bruh.WrapSkip(err, 1, message),
		UserID: userID,
	}
}
