package endpoints

import (
	"io"
	"net/http"
	"net/url"
)

type MockRespWriter struct {
}

func (w *MockRespWriter) Header() http.Header {
	return nil
}

func (w *MockRespWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *MockRespWriter) WriteHeader(statusCode int) {
}

type MockRequest struct {
	Method string
	URL    *url.URL

	Proto      string
	ProtoMajor int
	ProtoMinor int

	Header http.Header

	Body io.ReadCloser
}

func (r *MockRequest) GetBody(io.ReadCloser, error) {

}
