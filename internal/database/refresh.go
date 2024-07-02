package database

func (db *DB) AddRefreshToken(refreshToken string) {
	db.mux.Lock()
	defer db.mux.Unlock()

	db.database.RefreshTokens[refreshToken] = true
}

func (db *DB) RevokeRefreshToken(refreshToken string) {
	db.mux.Lock()
	defer db.mux.Unlock()

	delete(db.database.RefreshTokens, refreshToken)
}

func (db *DB) IsValidRefreshToken(refreshToken string) bool {
	_, ok := db.database.RefreshTokens[refreshToken]
	return ok
}
