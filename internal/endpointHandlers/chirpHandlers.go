package endpointhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Quorum-Code/chirpy/internal"
)

func (cfg *ApiConfig) PostChirpsHandler(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	p := parameters{}
	err := decoder.Decode(&p)
	if err != nil {
		resp.WriteHeader(400)
		return
	}

	chirp, err := cfg.Db.CreateChirp(p.Body)
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

func (cfg *ApiConfig) GetChirpsHandler(resp http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.Db.GetChirps()
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

func (cfg *ApiConfig) GetChirpByIDHandler(resp http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("chirpID"))
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("end must be int"))
		return
	}

	chirp, err := cfg.Db.GetChirp(id)
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

func ValidateChirpHandler(resp http.ResponseWriter, req *http.Request) {
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
