package main

import (
	"fmt"
	"net/http"

	"github.com/Quorum-Code/chirpy/internal"
)

func main() {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("../../."))
	apiCfg := apiConfig{}
	handler := http.StripPrefix("/app", fileServer)

	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("/metrics", getMetricsHandler(&apiCfg))
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/reset", apiCfg.middlewareMetricsReset())

	corsMux := internal.MiddlewareCors(mux)
	server := http.Server{Addr: ":8000", Handler: corsMux}
	server.ListenAndServe()
}

type apiConfig struct {
	fileserverHits int
}

func healthzHandler(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	resp.Write([]byte("OK"))
}

func getMetricsHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits)))
	}
}

func (cfg *apiConfig) middlewareMetricsReset() func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits = 0
		resp.WriteHeader(200)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}
