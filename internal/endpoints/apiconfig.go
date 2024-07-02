package endpoints

import "github.com/Quorum-Code/chirpy/internal"

type ApiConfig struct {
	FileserverHits int
	Db             internal.DB
}
