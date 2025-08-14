package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

var ErrNotExists = os.ErrNotExist

func main() {
	config, err := loadConfig("example.json")
	if err != nil {
		fmt.Println(
			clean(bruh.StringFormat(err, bruh.BruhStackedFancyFormatter(true, true, false))),
		)
		os.Exit(1)
	}
	fmt.Println(config)
}

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
	return nil, bruh.Wrapf(ErrNotExists, "reading file '%s'", path)
}

// clean replaces the paths in the formatted error output to allow for consistent testing.
func clean(s string) string {
	_, repoDir, _, _ := runtime.Caller(0)
	repoDir = filepath.Dir(filepath.Dir(filepath.Dir(repoDir))) + "/"
	s = strings.ReplaceAll(s, repoDir, "")
	return s
}
