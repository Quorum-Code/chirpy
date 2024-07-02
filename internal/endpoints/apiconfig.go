package endpoints

import (
	"github.com/Quorum-Code/chirpy/internal/database"
)

type ApiConfig struct {
	FileserverHits int
	Db             database.DB
}
