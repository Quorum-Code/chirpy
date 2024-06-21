package internal

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path     string
	database *Database
	mux      *sync.RWMutex

	JWT_SECRET  string
	polkaApiKey string
}

type Chirp struct {
	Body     string `json:"body"`
	Id       int    `json:"id"`
	AuthorId int    `json:"author_id"`
}

type User struct {
	Email       string `json:"email"`
	Id          int    `json:"id"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type Database struct {
	Users         map[int]User    `json:"users"`
	NextUID       int             `json:"nextuid"`
	Chirps        map[int]Chirp   `json:"chirps"`
	NextCID       int             `json:"nextcid"`
	Hashes        map[int][]byte  `json:"hashes"`
	RefreshTokens map[string]bool `json:"refresh_tokens"`
}

func NewDB(path string) (*DB, error) {
	database := &Database{Chirps: make(map[int]Chirp), NextCID: 1, Users: make(map[int]User), NextUID: 1}
	database.Hashes = make(map[int][]byte)
	database.RefreshTokens = make(map[string]bool)

	db := DB{database: database,
		mux:         &sync.RWMutex{},
		path:        path,
		polkaApiKey: os.Getenv("POLKA_SECRET"),
		JWT_SECRET:  os.Getenv("JWT_SECRET")}

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	var err error
	if *dbg {
		fmt.Println("Starting in Debug Mode")
		err = db.freshEnsureDB()
	} else {
		fmt.Println("Starting in Normal Mode")
		err = db.ensureDB()
	}
	if err != nil {
		return nil, err
	}

	err = db.loadDB()
	if err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *DB) IsPolkaKey(key string) bool {
	db.mux.RLock()
	defer db.mux.RUnlock()

	return key == db.polkaApiKey
}

func (db *DB) AddRefreshToken(refreshToken string) {
	db.mux.Lock()
	defer db.mux.Unlock()

	db.database.RefreshTokens[refreshToken] = true
}

func (db *DB) RevokeRefreshToken(refreshToken string) {
	db.mux.Lock()
	defer db.mux.Unlock()

	delete(db.database.RefreshTokens, refreshToken)
}

func (db *DB) IsValidRefreshToken(refreshToken string) bool {
	_, ok := db.database.RefreshTokens[refreshToken]
	return ok
}

func (db *DB) ValidLogin(email string, pass string) (User, bool) {
	user, ok := db.getUserByEmail(email)
	if !ok {
		return User{}, false
	}

	_, ok = db.database.Hashes[user.Id]
	if !ok {
		return User{}, false
	}

	err := bcrypt.CompareHashAndPassword(db.database.Hashes[user.Id], []byte(pass))
	if err != nil {
		return User{}, false
	} else {
		return user, true
	}
}

func (db *DB) GetUserById(id int) (*User, bool) {
	user, ok := db.database.Users[id]
	return &user, ok
}

func (db *DB) IsEmailUsed(email string) bool {
	db.mux.RLock()
	defer db.mux.RUnlock()

	for _, user := range db.database.Users {
		if user.Email == email {
			return true
		}
	}
	return false
}

func (db *DB) getUserByEmail(email string) (User, bool) {
	for _, val := range db.database.Users {
		if val.Email == email {
			return val, true
		}
	}

	return User{}, false
}

func (db *DB) CreateUser(email string, pass string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user := User{Email: email, Id: db.database.NextUID, IsChirpyRed: false}
	db.database.NextUID++
	db.database.Users[user.Id] = user

	db.database.Hashes[user.Id] = hash

	go db.writeDB()

	return user, nil
}

func (db *DB) UpdateUser(id int, email string, pass string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user := User{Email: email, Id: id}

	db.database.Users[user.Id] = user
	db.database.Hashes[user.Id] = hash

	return user, nil
}

func (db *DB) UpgradeUser(id int) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	user, ok := db.GetUserById(id)
	if !ok {
		return errors.New("no user found with id")
	}

	user.IsChirpyRed = true

	db.database.Users[id] = *user

	return nil
}

func (db *DB) DeleteChirp(cid int) {
	delete(db.database.Chirps, cid)
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
	if db.database.Chirps == nil {
		return Chirp{}, errors.New("chirp map is nil")
	}

	chirp, ok := db.database.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("ID not found in map")
	}

	return chirp, nil
}

func (db *DB) freshEnsureDB() error {
	dat, err := json.Marshal(db.database)
	if err != nil {
		return err
	}

	os.WriteFile(db.path, dat, 0777)
	return nil
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
	dat, err := json.Marshal(db.database)
	if err != nil {
		return err
	}

	os.WriteFile(db.path, dat, 0777)

	return nil
}
