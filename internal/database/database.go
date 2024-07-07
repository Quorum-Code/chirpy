package database

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
)

type DB struct {
	path     string
	database *Database
	mux      *sync.RWMutex

	JWT_SECRET  string
	polkaApiKey string
}

type Database struct {
	NextUID       int             `json:"nextuid"`
	NextCID       int             `json:"nextcid"`
	Chirps        map[int]Chirp   `json:"chirps"`
	Users         map[int]User    `json:"users"`
	RefreshTokens map[string]bool `json:"refresh_tokens"`
	Hashes        map[int][]byte  `json:"hashes"`
}

var ErrChirpNotFound error = errors.New("chirp not found")
var ErrUserNotFound error = errors.New("user not found")

type User struct {
	Email       string `json:"email"`
	Id          int    `json:"id"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

// Initialize db from io.Reader
func InitDB(reader io.Reader) (*DB, error) {
	database := &Database{
		Chirps:        make(map[int]Chirp),
		NextCID:       1,
		Users:         make(map[int]User),
		NextUID:       1,
		Hashes:        make(map[int][]byte),
		RefreshTokens: make(map[string]bool),
	}

	db := DB{
		database:    database,
		mux:         &sync.RWMutex{},
		path:        "./database.json",
		polkaApiKey: os.Getenv("POLKA_SECRET"),
		JWT_SECRET:  os.Getenv("JWT_SECRET"),
	}

	err := db.loadDB()
	if err != nil {
		return nil, err
	}

	return &db, nil
}

// Initialize empty db
func InitCleanDB() *DB {
	database := &Database{
		Chirps:        make(map[int]Chirp),
		NextCID:       1,
		Users:         make(map[int]User),
		NextUID:       1,
		Hashes:        make(map[int][]byte),
		RefreshTokens: make(map[string]bool),
	}

	db := DB{
		database:    database,
		mux:         &sync.RWMutex{},
		path:        "",
		polkaApiKey: os.Getenv("POLKA_SECRET"),
		JWT_SECRET:  os.Getenv("JWT_SECRET"),
	}

	return &db
}

func (db *DB) loadDB() error {
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := os.ReadFile(db.path)
	if err != nil {
		return err
	}

	database := Database{}
	err = json.Unmarshal(dat, &database)
	if err != nil {
		return err
	}
	db.database = &database

	return nil
}

func (db *DB) writeDB() error {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dat, err := json.MarshalIndent(db.database, "", "  ")
	if err != nil {
		return err
	}

	os.WriteFile(db.path, dat, 0777)

	return nil
}
