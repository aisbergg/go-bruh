package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// -----------------------------------------------------------------------------
// The custom error
// -----------------------------------------------------------------------------

// TimestampedError is a custom error that contains a timestamp.
type TimestampedError struct {
	bruh.Err
	timestamp time.Time
}

// Errorf creates a new TimestampedError with a formatted message.
func Errorf(format string, args ...any) error {
	return &TimestampedError{
		// skip is required to skip the current function and thus exclude this
		// function from the stack trace
		Err:       *bruh.ErrorfSkip(1, format, args...),
		timestamp: time.Now(),
	}
}

// Wrapf wraps an error with a formatted message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	// Easiest way to change the final error message is to change it early on.
	// Here we include the timestamp in the message.
	ts := time.Now()
	msg := fmt.Sprintf(format, args...)
	msg = fmt.Sprintf("[%s] %s", ts.Format(time.RFC3339), msg)

	return &TimestampedError{
		Err:       *bruh.WrapSkip(err, 1, msg),
		timestamp: ts,
	}
}

// Timestamp returns the timestamp of the error.
func (te *TimestampedError) Timestamp() time.Time {
	return te.timestamp
}

// -----------------------------------------------------------------------------
// Main
// -----------------------------------------------------------------------------

func main() {
	_, err := loadConfig("example.json")

	var terr *TimestampedError
	var berr *bruh.Err
	if bruh.As(err, &terr) {
		fmt.Printf(
			"timestamp error:\n  time: %s\n  msg: %s\n",
			terr.Timestamp().Format(time.RFC3339),
			terr,
		)
	} else if bruh.As(err, &berr) {
		fmt.Println("bruh error:", berr)
	} else {
		fmt.Println("other error:", err)
	}
}

// -----------------------------------------------------------------------------
// Just some functions that return errors
// -----------------------------------------------------------------------------

type Config struct {
	Interface string
	Port      int
}

func loadConfig(path string) (*Config, error) {
	fileContent, err := readFile(path)
	if err != nil {
		return nil, bruh.Wrap(err, "loading config")
	}
	var config Config
	if err := json.Unmarshal(fileContent, &config); err != nil {
		return nil, bruh.Wrap(err, "parsing config")
	}
	return &config, nil
}

func readFile(path string) ([]byte, error) {
	return nil, Wrapf(os.ErrNotExist, "reading file '%s'", path)
}
