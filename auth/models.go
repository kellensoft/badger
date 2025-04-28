package auth

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`  // Don't JSON expose password hashes!
}

type AccessToken struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp
	CreatedAt int64  `json:"created_at"`
}

type RefreshToken struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp
	CreatedAt int64  `json:"created_at"`
}
