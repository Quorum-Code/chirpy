package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Quorum-Code/chirpy/internal"
)

func main() {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("../../."))
	apiCfg := apiConfig{}
	handler := http.StripPrefix("/app", fileServer)

	db, err := internal.NewDB("./database.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	apiCfg.db = *db

	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/metrics", getMetricsHandler(&apiCfg))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("/api/reset", apiCfg.middlewareMetricsReset())
	mux.HandleFunc("GET /admin/metrics", adminMetricsHandler(&apiCfg))
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("POST /api/chirps", postChirpsHandler(&apiCfg))
	mux.HandleFunc("GET /api/chirps", getChirpsHandler(&apiCfg))
	mux.HandleFunc("GET /api/chirps/{chirpID}", getChirpByIDHandler(&apiCfg))
	mux.HandleFunc("POST /api/users", postUserHandler(&apiCfg))
	mux.HandleFunc("POST /api/login", postLoginHandler(&apiCfg))

	corsMux := internal.MiddlewareCors(mux)
	server := http.Server{Addr: ":8000", Handler: corsMux}
	server.ListenAndServe()
}

type apiConfig struct {
	fileserverHits int
	db             internal.DB
}

func postLoginHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Pass  string `json:"password"`
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		p := parameters{}
		err := decoder.Decode(&p)
		if err != nil {
			resp.WriteHeader(400)
			resp.Write([]byte("unparseable body"))
			return
		}

		user, ok := cfg.db.ValidLogin(p.Email, p.Pass)

		if !ok {
			resp.WriteHeader(401)
			resp.Write([]byte("incorrect login information"))
			return
		} else {
			dat, err := json.Marshal(user)
			if err != nil {
				resp.WriteHeader(500)
				resp.Write([]byte("something went wrong while loging in, please try again"))
				return
			}

			resp.WriteHeader(200)
			resp.Write(dat)
			return
		}
	}
}

func postUserHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Pass  string `json:"password"`
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		p := parameters{}
		err := decoder.Decode(&p)
		if err != nil {
			resp.WriteHeader(400)
			resp.Write([]byte("unparseable body"))
			return
		}

		user, err := cfg.db.CreateUser(p.Email, p.Pass)
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte("something went wrong while creating the user"))
			return
		}

		dat, err := json.Marshal(user)
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte("something went wrong while decoding the user"))
			return
		}

		resp.WriteHeader(201)
		resp.Write(dat)
	}
}

func getChirpByIDHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		id, err := strconv.Atoi(req.PathValue("chirpID"))
		if err != nil {
			resp.WriteHeader(400)
			resp.Write([]byte("end must be int"))
			return
		}

		chirp, err := cfg.db.GetChirp(id)
		if err != nil {
			resp.WriteHeader(404)
			resp.Write([]byte(err.Error()))
			return
		}

		dat, err := json.Marshal(chirp)
		if err != nil {
			resp.WriteHeader(400)
			resp.Write([]byte(err.Error()))
			return
		}

		resp.WriteHeader(200)
		resp.Write(dat)
	}
}

func postChirpsHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		p := parameters{}
		err := decoder.Decode(&p)
		if err != nil {
			resp.WriteHeader(400)
			return
		}

		chirp, err := cfg.db.CreateChirp(p.Body)
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))
			return
		}

		resp.WriteHeader(201)
		dat, err := json.Marshal(chirp)
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))
			return
		}

		resp.Write(dat)
	}
}

func getChirpsHandler(cfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		chirps, err := cfg.db.GetChirps()
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))
			return
		}

		dat, err := json.Marshal(chirps)
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))
			return
		}

		resp.WriteHeader(200)
		resp.Write(dat)
	}
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
