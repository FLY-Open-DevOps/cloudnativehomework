package module2

import (
	"encoding/json"
	"log"
	"module2/server"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	const (
		version       = "v1"
		myHeaderKey   = "MY_HEADER"
		myHeaderValue = "hello"
	)
	// start server
	os.Setenv(server.VERSION_ENV, version)
	go server.Serve(":8080")
	time.Sleep(time.Second)

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
	if v := resp.Header.Get(server.VERSION_ENV); v != version {
		log.Fatalf("expect HEADER %s is %s, but got %s", server.VERSION_ENV, version, v)
	}
	if v := resp.Header.Get(myHeaderKey); v != myHeaderValue {
		log.Fatalf("expect HEADER %s is %s, but got %s", myHeaderKey, myHeaderValue, v)
	}

	// print response
	respJson, err := json.MarshalIndent(resp.Header, "", "\t")
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("\nStatusCode is %d\nResponse Header is %s", resp.StatusCode, respJson)
}
