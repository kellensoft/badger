package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"os"
	"strconv"
)

func generateRandomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func CreateAccessToken() (token string, expiresAt int64, err error) {
	token, err = generateRandomToken(32) 
	if err != nil {
		return "", 0, err
	}

	minutes := 30 
	if envMinutes := os.Getenv("ACCESS_TOKEN_EXPIRY_MINUTES"); envMinutes != "" {
		if parsed, err := strconv.Atoi(envMinutes); err == nil {
			minutes = parsed
		}
	}
	expiresAt = time.Now().Add(time.Duration(minutes) * time.Minute).Unix()
	return token, expiresAt, nil
}

func CreateRefreshToken() (token string, expiresAt int64, err error) {
	token, err = generateRandomToken(32)
	if err != nil {
		return "", 0, err
	}

	days := 30 
	if envDays := os.Getenv("REFRESH_TOKEN_EXPIRY_DAYS"); envDays != "" {
		if parsed, err := strconv.Atoi(envDays); err == nil {
			days = parsed
		}
	}
	expiresAt = time.Now().Add(time.Duration(days) * 24 * time.Hour).Unix()
	return token, expiresAt, nil
}
