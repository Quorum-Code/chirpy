package webserver

import (
	"testing"
)

func TestServer(t *testing.T) {
	server := StartServer(false)

	if server == nil {
		t.Error("server failed to start")
	}
}
