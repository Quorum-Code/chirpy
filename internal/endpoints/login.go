package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *ApiConfig) PostSignupHandler(resp http.ResponseWriter, req *http.Request) {
	// Get email
	email := req.Header.Get("email")
	if email == "" {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("No email in header"))
		return
	}

	// Get password
	password := req.Header.Get("password")
	if password == "" {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("No password in header"))
		return
	}

	// Check email not used
	if cfg.Db.IsEmailUsed(email) {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("Email already used by another account"))
		return
	}

	_, err := cfg.Db.CreateUser(email, password)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
	}

	resp.WriteHeader(http.StatusAccepted)
	resp.Write([]byte("User created"))
}

func (cfg *ApiConfig) PostLoginHandler(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Pass  string `json:"password"`
	}

	type details struct {
		Id           int    `json:"id"`
		Email        string `json:"email"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(req.Body)
	p := parameters{}
	err := decoder.Decode(&p)
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("unparseable body"))
		return
	}

	user, ok := cfg.Db.ValidLogin(p.Email, p.Pass)
	if !ok {
		resp.WriteHeader(401)
		resp.Write([]byte("incorrect login information"))
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	ca := jwt.RegisteredClaims{Issuer: "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add((time.Hour)).UTC()),
		Subject:   strconv.Itoa(user.Id)}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, ca)

	cr := jwt.RegisteredClaims{Issuer: "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(time.Hour * 24 * 60)).UTC()),
		Subject:   strconv.Itoa(user.Id)}
	rtk := jwt.NewWithClaims(jwt.SigningMethodHS256, cr)

	ts, err := tk.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Println(err.Error())
		resp.WriteHeader(402)
		return
	}

	trs, err := rtk.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Println(err.Error())
		resp.WriteHeader(402)
		return
	}
	cfg.Db.AddRefreshToken(trs)

	d := details{Id: user.Id, Email: user.Email, IsChirpyRed: user.IsChirpyRed, Token: ts, RefreshToken: trs}

	dat, err := json.Marshal(d)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte("something went wrong while tokening"))
		return
	}

	resp.WriteHeader(200)
	resp.Write(dat)
}
