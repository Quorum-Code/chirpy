package endpointhandlers

import "net/http"

func (cfg *ApiConfig) HealthzHandler(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	resp.Write([]byte("OK"))
}
