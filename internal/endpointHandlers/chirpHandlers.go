package endpointhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Quorum-Code/chirpy/internal"
)

func (cfg *ApiConfig) GetChirpsByAuthor(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(400)
}

func (cfg *ApiConfig) DeleteChirpsHandler(resp http.ResponseWriter, req *http.Request) {
	cid, err := strconv.Atoi(req.PathValue("chirpID"))
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("end must be int"))
		return
	}

	authData, err := internal.RequestToToken(req)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(err.Error()))
		return
	}

	chirp, err := cfg.Db.GetChirp(cid)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(err.Error()))
		return
	}

	uid, err := strconv.Atoi(authData.Claim.Subject)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(err.Error()))
		return
	}

	if uid == chirp.AuthorId {
		cfg.Db.DeleteChirp(chirp.Id)
		resp.WriteHeader(200)
		resp.Write([]byte("chirp deleted"))
		return
	}

	resp.WriteHeader(403)
	resp.Write([]byte("not your chirp"))
}

func (cfg *ApiConfig) PostChirpsHandler(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	// verify authorization

	decoder := json.NewDecoder(req.Body)
	p := parameters{}
	err := decoder.Decode(&p)
	if err != nil {
		resp.WriteHeader(400)
		return
	}

	authData, err := internal.RequestToToken(req)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(err.Error()))
		return
	}

	subject := authData.Claim.Subject
	id, err := strconv.Atoi(subject)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(err.Error()))
		return
	}

	chirp, err := cfg.Db.CreateChirp(id, p.Body)
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
	sid := req.URL.Query().Get("author_id")
	id, err := strconv.Atoi(sid)
	var chirps []internal.Chirp

	if err != nil {
		// no id
		chirps, err = cfg.Db.GetChirps()
		if err != nil {
			resp.WriteHeader(400)
			resp.Write([]byte(err.Error()))
			return
		}
	} else {
		// user id
		chirps, err = cfg.Db.GetChirpsByUserID(id)
		if err != nil {
			resp.WriteHeader(400)
			resp.Write([]byte(err.Error()))
			return
		}
	}

	sort := req.URL.Query().Get("sort")
	if sort == "desc" {
		// reverse chirps slice
		for i, j := 0, len(chirps)-1; i < j; i, j = i+1, j-1 {
			chirps[i], chirps[j] = chirps[j], chirps[i] //reverse the slice
		}
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
