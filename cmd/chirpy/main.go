package main

import (
	_ "embed"

	"fmt"
	"net/http"

	"github.com/Quorum-Code/chirpy/internal"
	endpointhandlers "github.com/Quorum-Code/chirpy/internal/endpoints"
	"github.com/flowchartsman/swaggerui"
	"github.com/joho/godotenv"
)

//go:embed spec/chirpy.yml
var spec []byte

func main() {
	godotenv.Load("../../.env")

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("../../."))
	apiCfg := endpointhandlers.ApiConfig{}
	handler := http.StripPrefix("/app", fileServer)

	db, err := internal.NewDB("./database.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	apiCfg.Db = *db

	mux.HandleFunc("/", apiCfg.IndexHandler)

	mux.HandleFunc("POST /api/signup", apiCfg.PostSignupHandler)

	mux.HandleFunc("POST /oauth/token", apiCfg.PostToken)

	mux.HandleFunc("POST /chirps", apiCfg.PostChirp)
	mux.HandleFunc("GET /chirps", apiCfg.GetChirps)
	mux.HandleFunc("PUT /chirps/{chirpID}", apiCfg.PutChirp)
	mux.HandleFunc("DELETE /chirps/{chirpID}", apiCfg.DeleteChirp)

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
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.DeleteChirpsHandler)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.PostPolkaWebhook)

	mux.Handle("/swagger/", http.StripPrefix("/swagger", swaggerui.Handler(spec)))

	corsMux := internal.MiddlewareCors(mux)
	server := http.Server{Addr: ":8000", Handler: corsMux}

	fmt.Println("serving SwaggerUI at: localhost:8000/swagger/")
	server.ListenAndServe()
}
