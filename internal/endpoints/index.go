package endpoints

import (
	"fmt"
	"net/http"
)

func (cfg *ApiConfig) IndexHandler(resp http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		fmt.Println(req.URL.Path)
		resp.WriteHeader(404)
		resp.Write([]byte("NOT OK"))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte("OK"))
}
