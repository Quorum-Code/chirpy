package internal

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

	split := strings.Split(t, " ")
	if len(split) > 1 {
		t = split[1]
	}

	var claims Claims
	token, err := jwt.ParseWithClaims(t, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return AuthData{}, err
	}

	return AuthData{Token: token, Claim: claims}, nil
}
