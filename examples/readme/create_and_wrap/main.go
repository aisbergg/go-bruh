package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// define a global error; you can also create root errors with `bruh.New`
// elsewhere as needed
var ErrInternalServer = bruh.New("error internal server")

func main() {
	url := "https://foo-bar.local/resource/dog.jpg"
	if _, err := Get(url); err != nil {
		err = bruh.Wrap(err, "getting an image of a dog")
		fmt.Fprintf(os.Stderr, "%s\n", bruh.String(err))
		os.Exit(1)
	}
}

func Get(url string) (*http.Response, error) {
	client := http.Client{Timeout: 300 * time.Millisecond}
	res, err := client.Get(url)
	if err == nil && res.StatusCode != http.StatusOK {
		// create a root error with a formatted message
		err = bruh.Errorf("GET \"%s\" failed with status code %d", url, res.StatusCode)
	}
	if err != nil {
		// wrap the error and add context
		return nil, bruh.Wrap(err, "requesting remote resource")
	}
	return res, nil
}
