package endpointhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type parameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *ApiConfig) PostUserHandler(resp http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Pass  string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	p := parameters{}
	err := decoder.Decode(&p)
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("unparseable body"))
		return
	}

	user, err := cfg.Db.CreateUser(p.Email, p.Pass)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("something went wrong while creating the user"))
		return
	}

	dat, err := json.Marshal(user)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte("something went wrong while decoding the user"))
		return
	}

	resp.WriteHeader(201)
	resp.Write(dat)
}

func (cfg *ApiConfig) PutUsersHandler(resp http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	p := parameters{}
	err := decoder.Decode(&p)
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("unable to parse body"))
		return
	}

	t := req.Header.Get("Authorization")
	split := strings.Split(t, " ")
	if len(split) > 1 {
		t = split[1]
	}

	type CustomClaim struct {
		Issuer string `json:"iss"`
		jwt.RegisteredClaims
	}

	var claims CustomClaim
	token, err := jwt.ParseWithClaims(t, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		fmt.Println(err.Error())
		resp.WriteHeader(401)
		resp.Write([]byte("unauthorized token"))
		return
	}

	fmt.Printf("ISSUER: %s\n", claims.Issuer)
	if claims.Issuer != "chirpy-access" {
		resp.WriteHeader(401)
		resp.Write([]byte("not access token"))
		return
	}

	if !token.Valid {
		resp.WriteHeader(401)
		resp.Write([]byte("unauthorized"))
		return
	}

	idString, err := token.Claims.GetSubject()
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("error getting subject"))
		return
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("bad id"))
	}

	user, err := cfg.Db.UpdateUser(id, p.Email, p.Password)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(err.Error()))
		return
	}

	dat, err := json.Marshal(user)
	if err != nil {
		dat = []byte("Warning: unable to decode user")
	}

	resp.WriteHeader(200)
	resp.Write(dat)
}
