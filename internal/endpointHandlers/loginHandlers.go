package endpointhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
