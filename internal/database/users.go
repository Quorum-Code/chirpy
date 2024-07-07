package database

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) GetUserById(id int) (*User, bool) {
	user, ok := db.database.Users[id]
	return &user, ok
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
