package main

import (
	"encoding/json"
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
	mux.HandleFunc("GET /api/metrics", getMetricsHandler(&apiCfg))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("/api/reset", apiCfg.middlewareMetricsReset())
	mux.HandleFunc("GET /admin/metrics", adminMetricsHandler(&apiCfg))
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	corsMux := internal.MiddlewareCors(mux)
	server := http.Server{Addr: ":8000", Handler: corsMux}
	server.ListenAndServe()
}

type apiConfig struct {
	fileserverHits int
}

func adminMetricsHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf(internal.AdminMetricHTML(), cfg.fileserverHits)))
	}
}

func validateChirpHandler(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Error decoding parameters: %s", err)
		resp.WriteHeader(500)
	}

	type returnVals struct {
		// CreatedAt    time.Time `json:"created_at"`
		// ID           int       `json:"id"`
		// Valid        bool      `json:"valid"`
		Cleaned_Body string `json:"cleaned_body"`
	}

	//valid := false
	if len(params.Body) <= 140 {
		//valid = true
		resp.WriteHeader(200)

	} else {
		resp.WriteHeader(400)
	}

	respBody := returnVals{
		// CreatedAt:    time.Now(),
		// ID:           123,
		// Valid:        valid,
		Cleaned_Body: internal.StripProfane(params.Body),
	}
	fmt.Println(params.Body)
	dat, err := json.Marshal(respBody)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s", err)
		resp.WriteHeader(500)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.Write(dat)
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
