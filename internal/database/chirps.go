package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
)

type Chirp struct {
	Body     string `json:"body"`
	Id       int    `json:"id"`
	AuthorId int    `json:"author_id"`
}

func (db *DB) UserPutChirp(req *http.Request, chirpID int) error {
	// Get chirp
	chirp, err := db.GetChirp(chirpID)
	if err != nil {
		return err
	}

	// Get caller auths
	auth, err := RequestToToken(req)
	if err != nil {
		return err
	}
	authID, err := strconv.Atoi(auth.Claim.Subject)
	if err != nil {
		return err
	}

	// Only allow author or admin to edit chirps
	if chirp.AuthorId != authID {
		return ErrNotAuthorized
	}

	// Get updated chirp body from request
	type body struct {
		ChirpBody string `json:"chirpBody"`
	}
	decoder := json.NewDecoder(req.Body)
	b := body{}
	err = decoder.Decode(&b)
	if err != nil {
		return errors.New("bad request")
	}

	// Write to database
	db.mux.Lock()
	defer db.mux.Unlock()

	chirp.Body = b.ChirpBody
	db.database.Chirps[chirp.Id] = chirp

	go db.writeDB()

	return nil
}

func (db *DB) UserDeleteChirp(req *http.Request, chirpID int) error {
	// Get auth claims
	auth, err := RequestToToken(req)
	if err != nil {
		return err
	}
	userID, err := strconv.Atoi(auth.Claim.Subject)
	if err != nil {
		return err
	}

	// Get chirp
	chirp, err := db.GetChirp(chirpID)
	if err != nil {
		return err
	}

	// Check chirpID userID match
	if chirp.AuthorId != userID {
		return errors.New("unauthorized")
	}

	// Delete chirp
	db.DeleteChirp(chirpID)

	return nil
}

func (db *DB) DeleteChirp(cid int) {
	db.mux.Lock()
	defer db.mux.Unlock()

	delete(db.database.Chirps, cid)
	go db.writeDB()
}

func (db *DB) CreateChirp(id int, body string) (Chirp, error) {
	if len(body) > 140 {
		return Chirp{}, errors.New("chirp is too long")
	}

	if db.database.Chirps == nil {
		return Chirp{}, errors.New("chirps map is nil")
	}

	db.mux.Lock()
	defer db.mux.Unlock()

	chirp := Chirp{Body: body, Id: db.database.NextCID, AuthorId: id}
	db.database.NextCID++
	db.database.Chirps[chirp.Id] = chirp

	go db.writeDB()

	return chirp, nil
}

func (db *DB) GetChirpsByUserID(id int) ([]Chirp, error) {
	err := db.loadDB()
	if err != nil {
		return nil, err
	}

	db.mux.Lock()
	defer db.mux.Unlock()

	keys := []int{}
	for k := range db.database.Chirps {
		chirp := db.database.Chirps[k]
		if chirp.AuthorId == id {
			keys = append(keys, k)
		}
	}

	sort.Ints(keys)
	chirps := []Chirp{}
	for v := range keys {
		chirps = append(chirps, db.database.Chirps[keys[v]])
	}

	return chirps, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	err := db.loadDB()

	db.mux.Lock()
	defer db.mux.Unlock()

	if err != nil {
		return nil, err
	}

	keys := []int{}
	for k := range db.database.Chirps {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	chirps := []Chirp{}
	for v := range keys {
		chirps = append(chirps, db.database.Chirps[keys[v]])
	}

	return chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	if db.database.Chirps == nil {
		return Chirp{}, errors.New("chirp map is nil")
	}

	chirp, ok := db.database.Chirps[id]
	if !ok {
		fmt.Printf("OK: %t\n", ok)
		fmt.Println(db.database.Chirps)
		return Chirp{}, ErrChirpNotFound
	}

	return chirp, nil
}
