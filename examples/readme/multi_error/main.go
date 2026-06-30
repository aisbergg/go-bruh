package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aisbergg/go-bruh/pkg/multierror"
)

func main() {
	err := validateUser("", -4, "not-an-email")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func validateUser(name string, age int, email string) error {
	me := multierror.New("validating user payload", multierror.Options{})

	if name == "" {
		me.Add(errors.New("name is required"))
	}
	if age < 0 {
		me.Add(fmt.Errorf("age must be >= 0, got %d", age))
	}
	if !strings.Contains(email, "@") {
		me.Add(fmt.Errorf("email %q is invalid", email))
	}

	return me.ErrorOrNil()
}
