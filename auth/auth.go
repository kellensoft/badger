package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"encoding/json"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	store       *Store
	emailSender *EmailSender
}

type contextKey string

const userIDKey contextKey = "userID"

func NewAuth(store *Store) *Auth {
	return &Auth{
		store:       store,
		emailSender: NewEmailSender(),
	}
}

func (a *Auth) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Only POST allowed")
			return
		}
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		if username == "" || email == "" || password == "" {
			writeError(w, http.StatusBadRequest, "Missing fields")
			return
		}
		if err := a.Signup(username, email, password); err != nil {
			writeError(w, http.StatusBadRequest, "Signup failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"message": "Signup successful"})
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Only POST allowed")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "" || password == "" {
			writeError(w, http.StatusBadRequest, "Missing fields")
			return
		}
		accessToken, refreshToken, err := a.Login(username, password)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Login failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Only POST allowed")
			return
		}
		refreshToken := r.FormValue("refresh_token")
		if refreshToken == "" {
			writeError(w, http.StatusBadRequest, "Missing refresh token")
			return
		}
		newAccessToken, err := a.Refresh(refreshToken)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Refresh failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"access_token": newAccessToken,
		})
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Only POST allowed")
			return
		}
		token := extractBearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}
		if err := a.Logout(token); err != nil {
			writeError(w, http.StatusBadRequest, "Logout failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Logged out successfully",
		})
	})

	return mux
}

func (a *Auth) Signup(username, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := a.store.CreateUser(username, email, string(hashedPassword)); err != nil {
		return err
	}

	// Send welcome email (non-blocking)
	go func() {
		_ = a.emailSender.Send(email, "Welcome to GopherIt!", "Thanks for signing up!")
	}()

	return nil
}

func (a *Auth) Login(username, password string) (accessToken, refreshToken string, err error) {
	user, err := a.store.FindUserByUsername(username)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, accessExpiresAt, err := CreateAccessToken()
	if err != nil {
		return "", "", err
	}
	refreshToken, refreshExpiresAt, err := CreateRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err := a.store.CreateAccessToken(user.ID, accessToken, accessExpiresAt); err != nil {
		return "", "", err
	}
	if err := a.store.CreateRefreshToken(user.ID, refreshToken, refreshExpiresAt); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (a *Auth) Refresh(refreshToken string) (newAccessToken string, err error) {
	userID, err := a.store.FindUserIDByRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	newAccessToken, accessExpiresAt, err := CreateAccessToken()
	if err != nil {
		return "", err
	}

	if err := a.store.CreateAccessToken(userID, newAccessToken, accessExpiresAt); err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (a *Auth) Logout(accessToken string) error {
	userID, err := a.store.FindUserIDByAccessToken(accessToken)
	if err != nil {
		return errors.New("invalid token")
	}

	return a.store.DeleteAllTokensForUser(userID)
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := a.store.FindUserIDByAccessToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Inject user ID into context
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func UserIDFromContext(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(userIDKey).(int)
	return id, ok
}