package main

import (
	"fmt"
	"net/http"

	"github.com/Quorum-Code/chirpy/internal"
	endpointhandlers "github.com/Quorum-Code/chirpy/internal/endpointHandlers"
	"github.com/joho/godotenv"
)

func main() {

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("../../."))
	apiCfg := endpointhandlers.ApiConfig{}
	handler := http.StripPrefix("/app", fileServer)

	db, err := internal.NewDB("./database.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	apiCfg.Db = *db

	mux.Handle("/app/*", apiCfg.MiddlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/metrics", apiCfg.GetMetricsHandler)
	mux.HandleFunc("GET /api/healthz", apiCfg.HealthzHandler)
	mux.HandleFunc("/api/reset", apiCfg.MiddlewareMetricsReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.AdminMetricsHandler)
	mux.HandleFunc("POST /api/validate_chirp", endpointhandlers.ValidateChirpHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.PostChirpsHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirpByIDHandler)
	mux.HandleFunc("POST /api/users", apiCfg.PostUserHandler)
	mux.HandleFunc("POST /api/login", apiCfg.PostLoginHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.PutUsersHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.PostRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.PostRevoke)

	godotenv.Load("../../.env")

	corsMux := internal.MiddlewareCors(mux)
	server := http.Server{Addr: ":8000", Handler: corsMux}
	server.ListenAndServe()
}
