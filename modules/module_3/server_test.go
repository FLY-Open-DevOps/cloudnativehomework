package module2

import (
	"log"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	const (
		version       = "v1"
		myHeaderKey   = "MY_HEADER"
		myHeaderValue = "hello"
	)

	// build request
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/healthz", nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	req.Header.Set(myHeaderKey, myHeaderValue)

	// handle response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("expect status code  is %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}
