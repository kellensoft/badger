package auth

import (
	"database/sql"
	"errors"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := setupTables(db); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func setupTables(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS access_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`)
	return err
}

func (s *Store) CreateUser(username, email, passwordHash string) error {
	_, err := s.db.Exec(`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`, username, email, passwordHash)
	return err
}

func (s *Store) FindUserByUsername(username string) (*User, error) {
	row := s.db.QueryRow(`SELECT id, username, email, password_hash FROM users WHERE username = ?`, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *Store) CreateAccessToken(userID int, token string, expiresAt int64) error {
	_, err := s.db.Exec(`INSERT INTO access_tokens (user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?)`,
		userID, token, expiresAt, time.Now().Unix())
	return err
}

func (s *Store) CreateRefreshToken(userID int, token string, expiresAt int64) error {
	_, err := s.db.Exec(`INSERT INTO refresh_tokens (user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?)`,
		userID, token, expiresAt, time.Now().Unix())
	return err
}

func (s *Store) FindUserIDByAccessToken(token string) (int, error) {
	var userID int
	var expiresAt int64
	err := s.db.QueryRow(`SELECT user_id, expires_at FROM access_tokens WHERE token = ?`, token).Scan(&userID, &expiresAt)
	if err != nil {
		return 0, err
	}
	if time.Now().Unix() > expiresAt {
		return 0, errors.New("access token expired")
	}
	return userID, nil
}

func (s *Store) FindUserIDByRefreshToken(token string) (int, error) {
	var userID int
	var expiresAt int64
	err := s.db.QueryRow(`SELECT user_id, expires_at FROM refresh_tokens WHERE token = ?`, token).Scan(&userID, &expiresAt)
	if err != nil {
		return 0, err
	}
	if time.Now().Unix() > expiresAt {
		return 0, errors.New("refresh token expired")
	}
	return userID, nil
}

func (s *Store) DeleteAccessToken(token string) error {
	_, err := s.db.Exec(`DELETE FROM access_tokens WHERE token = ?`, token)
	return err
}

func (s *Store) DeleteRefreshToken(token string) error {
	_, err := s.db.Exec(`DELETE FROM refresh_tokens WHERE token = ?`, token)
	return err
}

func (s *Store) DeleteAllTokensForUser(userID int) error {
	_, err1 := s.db.Exec(`DELETE FROM access_tokens WHERE user_id = ?`, userID)
	_, err2 := s.db.Exec(`DELETE FROM refresh_tokens WHERE user_id = ?`, userID)

	if err1 != nil {
		return err1
	}
	return err2
}
