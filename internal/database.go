package internal

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"
)

type DB struct {
	path     string
	database *Database
	mux      *sync.RWMutex
	nextId   int
}

type Chirp struct {
	Body string `json:"body"`
	Id   int    `json:"id"`
}

type Database struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := DB{database: &Database{Chirps: make(map[int]Chirp)}, mux: &sync.RWMutex{}, path: path, nextId: 1}

	err := db.ensureDB()
	if err != nil {
		return nil, err
	}

	err = db.loadDB()
	if err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	if len(body) > 140 {
		return Chirp{}, errors.New("chirp is too long")
	}

	if db.database.Chirps == nil {
		return Chirp{}, errors.New("chirps map is nil")
	}

	db.mux.Lock()
	defer db.mux.Unlock()

	chirp := Chirp{Body: body, Id: db.nextId}
	db.database.Chirps[chirp.Id] = chirp
	db.nextId++

	go db.writeDB()

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	err := db.loadDB()
	if err != nil {
		return nil, err
	}

	db.mux.Lock()
	defer db.mux.Unlock()

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

func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	if err != nil {
		dat, err := json.Marshal(db.database)
		if err != nil {
			return err
		}
		os.WriteFile(db.path, dat, 0777)
	}
	return nil
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
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := json.Marshal(db.database)
	if err != nil {
		return err
	}

	os.WriteFile(db.path, dat, 0777)

	return nil
}
