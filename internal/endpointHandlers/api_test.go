package endpointhandlers

import (
	"fmt"
	"net/http"
	"testing"
)

type FakeRespWriter struct {
}

func (f *FakeRespWriter) Header() http.Header {
	return nil
}

func (f *FakeRespWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (f *FakeRespWriter) WriteHeader(statusCode int) {
	return
}

func TestFunc(t *testing.T) {
	fmt.Println("ran a test")

	a := 1 + 2

	if a != 3 {
		t.Error("oh no")
	}

	c := ApiConfig{}
	resp := FakeRespWriter{}

	c.PostChirp(&resp, &http.Request{})
}
