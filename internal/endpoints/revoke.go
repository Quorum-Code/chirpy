package endpoints

import (
	"net/http"
	"strings"
)

func (cfg *ApiConfig) PostRevoke(resp http.ResponseWriter, req *http.Request) {
	tk := req.Header.Get("Authorization")
	split := strings.Split(tk, " ")
	if len(split) > 1 {
		tk = split[1]
	}

	cfg.Db.RevokeRefreshToken(tk)

	resp.WriteHeader(200)
}
