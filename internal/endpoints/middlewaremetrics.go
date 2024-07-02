package endpoints

import (
	"fmt"
	"net/http"
)

func adminMetricHTML() string {
	return `<html>

	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>`
}

func (cfg *ApiConfig) AdminMetricsHandler(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf(adminMetricHTML(), cfg.FileserverHits)))
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) GetMetricsHandler(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf("Hits: %d", cfg.FileserverHits)))
}

func (cfg *ApiConfig) MiddlewareMetricsReset(resp http.ResponseWriter, req *http.Request) {
	cfg.FileserverHits = 0
	resp.WriteHeader(200)
}
