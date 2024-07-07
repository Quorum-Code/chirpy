package webserver

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Quorum-Code/chirpy/internal"
	"github.com/Quorum-Code/chirpy/internal/database"
	"github.com/Quorum-Code/chirpy/internal/endpoints"

	"github.com/flowchartsman/swaggerui"
	"github.com/joho/godotenv"
)

//go:embed spec/chirpy.yml
var spec []byte
var root = "../../."

var ChirpyFolder = ".chirpy"
var DatabaseFile = "database.json"
var TestingDatabaseFile = "database-testing.json"
var TestingDatabasePath = "./test/data/database-testing.json"

func StartServer(cfg ServerConfig) *http.Server {
	fmt.Println("starting web server")

	// Load env variables
	godotenv.Load("../../.env")

	// Create server
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(root))
	apiCfg := endpoints.ApiConfig{}

	if cfg.IsDebug {
		// Load empty database
		apiCfg.Db = *database.InitCleanDB()
	} else {
		var path string
		if cfg.IsTesting {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return nil
			}

			path = filepath.Join(cwd, TestingDatabaseFile)
		} else {
			// Get user directory
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}

			// join path
			path = filepath.Join(home, ChirpyFolder, DatabaseFile)
		}

		// Get database file
		reader, err := os.Open(path)
		if err != nil {
			dir, _ := os.Getwd()

			fmt.Printf("%s, couldn't open db file, %s", err.Error(), dir)
			return nil
		}

		// Initialize database
		db, err := database.InitDB(reader, path)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		apiCfg.Db = *db
	}

	// Index url handler
	mux.HandleFunc("/", apiCfg.IndexHandler)

	// Signup handlers
	mux.HandleFunc("POST /api/signup", apiCfg.PostSignupHandler)

	// OAuth handlers
	mux.HandleFunc("POST /oauth/token", apiCfg.PostToken)

	// Chirps handlers
	mux.HandleFunc("POST /chirps", apiCfg.PostChirp)
	mux.HandleFunc("GET /chirps", apiCfg.GetChirps)
	mux.HandleFunc("GET /chirps/{chirpID}", apiCfg.GetChirpByID)
	mux.HandleFunc("PUT /chirps/{chirpID}", apiCfg.PutChirp)
	mux.HandleFunc("DELETE /chirps/{chirpID}", apiCfg.DeleteChirp)

	// Alias to remove "/app"
	handler := http.StripPrefix("/app", fileServer)
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
	if spec != nil {
		mux.Handle("/swagger/", http.StripPrefix("/swagger", swaggerui.Handler(spec)))
		fmt.Println("serving SwaggerUI at: localhost:8000/swagger/")
	}

	// Final server setup
	corsMux := internal.MiddlewareCors(mux)
	server := http.Server{Addr: ":8000", Handler: corsMux}

	// Start server
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println("ERROR: ", err)
		}
	}()

	return &server
}

// func getServerSpecs() ([]byte, error) {
// 	// User home
// 	userhome, err := os.UserHomeDir()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Join
// 	specpath := filepath.Join(userhome, ChirpyFolder, SpecYML)

// 	// Get file
// 	file, err := os.Open(specpath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	// File size to allocate byte slice capacity
// 	stat, err := file.Stat()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Read file into buffer
// 	buffer := make([]byte, stat.Size())
// 	_, err = bufio.NewReader(file).Read(buffer)
// 	if err != nil && err != io.EOF {
// 		return nil, err
// 	}

// 	return buffer, nil
// }
