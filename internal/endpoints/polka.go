package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (cfg *ApiConfig) PostPolkaWebhook(resp http.ResponseWriter, req *http.Request) {
	type body struct {
		Event string `json:"event,omitempty"`
		Data  struct {
			UserID int `json:"user_id,omitempty"`
		} `json:"data,omitempty"`
	}

	auth := req.Header.Get("Authorization")
	apiKey := ""
	split := strings.Split(auth, " ")
	if len(split) > 1 {
		apiKey = split[1]
	}

	if !cfg.Db.IsPolkaKey(apiKey) {
		resp.WriteHeader(401)
		return
	}

	b := body{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&b)
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte("couldn't parse body"))
		return
	}

	if b.Event == "user.upgraded" {
		err := cfg.Db.UpgradeUser(b.Data.UserID)

		if err != nil {
			resp.WriteHeader(404)
			resp.Write([]byte(err.Error()))
			return
		} else {
			resp.WriteHeader(200)
			return
		}
	} else {
		resp.WriteHeader(200)
		return
	}
}
