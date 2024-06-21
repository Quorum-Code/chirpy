package endpointhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (cfg *ApiConfig) PostTokenHandler(resp http.ResponseWriter, req *http.Request) {
	// Load the x-www-form-urlencoded data
	err := req.ParseForm()
	if err != nil {
		fmt.Println(err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check username not empty
	email := req.FormValue("username")
	if email == "" {
		fmt.Println("no username/email provided in request")
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check password not empty
	password := req.FormValue("password")
	if email == "" {
		fmt.Println("no password provided in request")
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check is valid login
	_, ok := cfg.Db.ValidLogin(email, password)
	if !ok {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Form the tokenResponse
	tr := tokenResponse{
		AccessToken:  "",
		RefreshToken: "",
		IDToken:      "",
		TokenType:    "",
		ExpiresIn:    1000,
	}

	// Marshal into json
	data, err := json.Marshal(tr)
	if err != nil {
		fmt.Println("failed to marshal tokenResponse")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write response
	resp.WriteHeader(http.StatusAccepted)
	resp.Write(data)
}

func (cfg *ApiConfig) PostRefresh(resp http.ResponseWriter, req *http.Request) {
	tk := req.Header.Get("Authorization")
	split := strings.Split(tk, " ")
	if len(split) > 1 {
		tk = split[1]
	}

	type CustomClaim struct {
		Issuer string `json:"iss"`
		jwt.RegisteredClaims
	}

	var claims CustomClaim
	token, err := jwt.ParseWithClaims(tk, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		fmt.Println(err.Error())
		resp.WriteHeader(401)
		resp.Write([]byte("unauthorized token"))
		return
	}

	// check valid
	if !token.Valid {
		resp.WriteHeader(401)
		return
	}

	// check is refresh
	if claims.Issuer != "chirpy-refresh" {
		resp.WriteHeader(401)
		return
	}

	// verify there are no revocations of this token
	if !cfg.Db.IsValidRefreshToken(tk) {
		resp.WriteHeader(401)
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	subject, err := claims.GetSubject()
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte("wrong user id"))
		return
	}

	ca := jwt.RegisteredClaims{Issuer: "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add((time.Hour)).UTC()),
		Subject:   subject}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, ca)

	ts, err := tok.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Println(err.Error())
		resp.WriteHeader(402)
		return
	}

	type details struct {
		Token string `json:"token"`
	}

	dat, err := json.Marshal(details{Token: ts})
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("something went wrong while marshaling token"))
		return
	}

	resp.WriteHeader(200)
	resp.Write(dat)
}
