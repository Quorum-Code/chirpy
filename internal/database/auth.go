package database

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthData struct {
	Token *jwt.Token
	Claim Claims
}

type Claims struct {
	Issuer string `json:"iss"`
	jwt.RegisteredClaims
}

func RequestToToken(req *http.Request) (AuthData, error) {
	t := req.Header.Get("Authorization")

	// Remove "Bearer" prefix
	split := strings.Split(t, " ")
	if len(split) > 1 {
		t = split[1]
	}

	var claims Claims
	token, err := jwt.ParseWithClaims(t, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		fmt.Printf("Authorization: %s\n", t)
		return AuthData{}, err
	}

	return AuthData{Token: token, Claim: claims}, nil
}

type OAuth2Access struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (db *DB) OAuth2Password(email string, password string) (OAuth2Access, error) {
	// Check that credentials valid
	user, ok := db.ValidLogin(email, password)
	if !ok {
		return OAuth2Access{}, errors.New("invalid credentials")
	}

	// Environment jwtSecret
	jwtSecret := os.Getenv("JWT_SECRET")

	// Generate access token
	ca := jwt.RegisteredClaims{Issuer: "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add((time.Hour)).UTC()),
		Subject:   strconv.Itoa(user.Id)}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, ca)

	// Sign access token
	ats, err := tk.SignedString([]byte(jwtSecret))
	if err != nil {
		return OAuth2Access{}, err
	}

	// Generate refresh token
	cr := jwt.RegisteredClaims{Issuer: "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(time.Hour * 24 * 60)).UTC()),
		Subject:   strconv.Itoa(user.Id)}
	rtk := jwt.NewWithClaims(jwt.SigningMethodHS256, cr)

	// Sign refresh token
	rts, err := rtk.SignedString([]byte(jwtSecret))
	if err != nil {
		return OAuth2Access{}, err
	}

	return OAuth2Access{AccessToken: ats, RefreshToken: rts, TokenType: "Bearer", ExpiresIn: 3600}, nil
}

func (db *DB) IsPolkaKey(key string) bool {
	db.mux.RLock()
	defer db.mux.RUnlock()

	return key == db.polkaApiKey
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
