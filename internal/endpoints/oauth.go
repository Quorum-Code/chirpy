package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (cfg *ApiConfig) PostToken(resp http.ResponseWriter, req *http.Request) {
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

	acc, err := cfg.Db.OAuth2Password(email, password)

	if err != nil {
		fmt.Println("OAuth2 failed")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Marshal into json
	data, err := json.Marshal(acc)
	if err != nil {
		fmt.Println("failed to marshal OAuth2Access")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write response
	resp.WriteHeader(http.StatusAccepted)
	resp.Write(data)
}
