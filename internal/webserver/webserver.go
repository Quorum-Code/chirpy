package webserver

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"

	"github.com/Quorum-Code/chirpy/internal"
	"github.com/Quorum-Code/chirpy/internal/database"
	"github.com/Quorum-Code/chirpy/internal/endpoints"

	"github.com/flowchartsman/swaggerui"
	"github.com/joho/godotenv"
)

//go:embed spec/chirpy.yml
var spec []byte
var root = "../../."
var dbpath = "./database.json"

func StartServer(isDebug bool) {
	fmt.Println("starting web server")

	// Load env variables
	godotenv.Load("../../.env")

	// Create server
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(root))
	apiCfg := endpoints.ApiConfig{}

	// Alias to remove "/app"
	handler := http.StripPrefix("/app", fileServer)

	var db *database.DB
	if isDebug {
		// If debug Initialize clean DB
		db = database.InitCleanDB()
	} else {
		// Get database file
		dbreader, err := os.Open(dbpath)
		if err != nil {
			fmt.Println("couldnt open db file")
			return
		}

		// Initialize database
		db, err = database.InitDB(dbreader)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	apiCfg.Db = *db

	// Index url handler
	mux.HandleFunc("/", apiCfg.IndexHandler)

	// Signup handlers
	mux.HandleFunc("POST /api/signup", apiCfg.PostSignupHandler)

	// OAuth handlers
	mux.HandleFunc("POST /oauth/token", apiCfg.PostToken)

	// Chirps handlers
	mux.HandleFunc("POST /chirps", apiCfg.PostChirp)
	mux.HandleFunc("GET /chirps", apiCfg.GetChirps)
	mux.HandleFunc("PUT /chirps/{chirpID}", apiCfg.PutChirp)
	mux.HandleFunc("DELETE /chirps/{chirpID}", apiCfg.DeleteChirp)

	mux.Handle("/app/*", apiCfg.MiddlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/metrics", apiCfg.GetMetricsHandler)
	mux.HandleFunc("GET /api/healthz", apiCfg.HealthzHandler)
	mux.HandleFunc("/api/reset", apiCfg.MiddlewareMetricsReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.AdminMetricsHandler)
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

	// Include swaggerui
	mux.Handle("/swagger/", http.StripPrefix("/swagger", swaggerui.Handler(spec)))

	// Final server setup
	corsMux := internal.MiddlewareCors(mux)
	server := http.Server{Addr: ":8000", Handler: corsMux}

	// Start server
	fmt.Println("serving SwaggerUI at: localhost:8000/swagger/")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}
}
