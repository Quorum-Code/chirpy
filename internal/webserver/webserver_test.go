package webserver

import (
	"testing"
)

func TestServer(t *testing.T) {
	server := StartServer(ServerConfig{
		IsTesting: true,
	})

	if server == nil {
		t.Error("server failed to start")
	}
}
