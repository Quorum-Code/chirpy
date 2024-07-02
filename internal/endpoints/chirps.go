package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Quorum-Code/chirpy/internal/database"
)

func (cfg *ApiConfig) PostChirp(resp http.ResponseWriter, req *http.Request) {
	type body struct {
		ChirpBody string `json:"chirpBody"`
	}

	// Check valid user
	auth, err := database.RequestToToken(req)
	if err != nil {
		resp.WriteHeader(http.StatusUnauthorized)
		resp.Write([]byte(err.Error()))
		return
	}
	if auth.Claim.Issuer != "chirpy-access" {
		resp.WriteHeader(http.StatusUnauthorized)
		resp.Write([]byte("no chirpy-access"))
		return
	}

	// Get body values
	decoder := json.NewDecoder(req.Body)
	b := body{}
	err = decoder.Decode(&b)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}

	// Get userID
	userID, err := strconv.Atoi(auth.Claim.Subject)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}

	// Create chirp
	_, err = cfg.Db.CreateChirp(userID, b.ChirpBody)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	// Return 200
	resp.WriteHeader(http.StatusOK)
}

func (cfg *ApiConfig) GetChirps(resp http.ResponseWriter, req *http.Request) {
	fmt.Println("getChirp")
}

func (cfg *ApiConfig) GetChirpByID(resp http.ResponseWriter, req *http.Request) {
	// Parse request
	cid, err := strconv.Atoi(req.PathValue("chirpID"))
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Find chirp
	chirp, err := cfg.Db.GetChirp(cid)
	if err != nil {
		if err == database.ErrChirpNotFound {
			resp.WriteHeader(http.StatusNotFound)
			return
		} else {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Get chirp info as json
	dat, err := json.MarshalIndent(&chirp, "", "  ")
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return chirp data and status
	resp.Write(dat)
	resp.WriteHeader(http.StatusAccepted)
}

func (cfg *ApiConfig) PutChirp(resp http.ResponseWriter, req *http.Request) {
	cid, err := strconv.Atoi(req.PathValue("chirpID"))
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Update chirp
	err = cfg.Db.UserPutChirp(req, cid)
	if err != nil {
		if err == database.ErrNotAuthorized {
			resp.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Return 201
	resp.WriteHeader(http.StatusCreated)
}

// Handles request to delete chirp by ID
func (cfg *ApiConfig) DeleteChirp(resp http.ResponseWriter, req *http.Request) {
	// Get chirpID
	cid, err := strconv.Atoi(req.PathValue("chirpID"))
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try to delete chirp
	err = cfg.Db.UserDeleteChirp(req, cid)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

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

	authData, err := database.RequestToToken(req)
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
		Body string `json:"chirpBody"`
	}

	decoder := json.NewDecoder(req.Body)
	p := parameters{}
	err := decoder.Decode(&p)
	if err != nil {
		fmt.Println("Bad body")

		buf := new(strings.Builder)
		_, err := io.Copy(buf, req.Body)
		if err != nil {
			fmt.Println(err)
			resp.WriteHeader(500)
			return
		}
		fmt.Println(buf.String())

		resp.WriteHeader(400)
		return
	}

	authData, err := database.RequestToToken(req)
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
	var chirps []database.Chirp

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

// func ValidateChirpHandler(resp http.ResponseWriter, req *http.Request) {
// 	type parameters struct {
// 		Body string `json:"body"`
// 	}

// 	decoder := json.NewDecoder(req.Body)
// 	params := parameters{}
// 	err := decoder.Decode(&params)
// 	if err != nil {
// 		fmt.Printf("Error decoding parameters: %s", err)
// 		resp.WriteHeader(500)
// 	}

// 	type returnVals struct {
// 		Cleaned_Body string `json:"cleaned_body"`
// 	}

// 	//valid := false
// 	if len(params.Body) <= 140 {
// 		//valid = true
// 		resp.WriteHeader(200)

// 	} else {
// 		resp.WriteHeader(400)
// 	}

// 	respBody := returnVals{
// 		Cleaned_Body: internal.StripProfane(params.Body),
// 	}
// 	fmt.Println(params.Body)
// 	dat, err := json.Marshal(respBody)
// 	if err != nil {
// 		fmt.Printf("Error marshalling JSON: %s", err)
// 		resp.WriteHeader(500)
// 		return
// 	}

// 	resp.Header().Set("Content-Type", "application/json")
// 	resp.Write(dat)
// }
