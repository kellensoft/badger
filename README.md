# badger
Go Authentication Service

## ✨ Features

- Secure password storage (bcrypt)
- Access + Refresh token system
- Modular `net/http` handlers (`Routes()`, `Middleware()`)
- Injects authenticated user ID via `context.Context`
- Environment-based configuration
- Minimal external dependencies

---

## 📦 Installation

```bash
go get github.com/kellensoft/badger
```

---

## ⚙️ Required Environment Variables

Create a `.env` file or set system env vars:

```dotenv
DATABASE_URL=auth.db

# SMTP settings (for signup welcome emails)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=yourusername
SMTP_PASSWORD=yourpassword
SMTP_FROM_EMAIL=no-reply@example.com

# Token expiration settings
ACCESS_TOKEN_EXPIRY_MINUTES=30
REFRESH_TOKEN_EXPIRY_DAYS=30
```

---

## 🚀 Usage Example

```go
package main

import (
	"log"
	"net/http"

	"github.com/kellensoft/badger/auth"
	"github.com/kellensoft/badger/config"
)

func main() {
	config.LoadEnv()

	store, err := auth.NewStore(config.GetEnv("DATABASE_URL", "auth.db"))
	if err != nil {
		log.Fatal(err)
	}

	authService := auth.NewAuth(store)

	mux := http.NewServeMux()

	// Mount auth routes under /auth/*
	mux.Handle("/auth/", http.StripPrefix("/auth", authService.Routes()))

	// Example protected route
	mux.Handle("/api/hello", authService.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserIDFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Write([]byte("Hello user ID: " + fmt.Sprint(userID)))
	})))

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

---

## 📚 Available Auth Routes

| Method | Path            | Description              |
|:-------|:-----------------|:--------------------------|
| POST   | `/auth/signup`    | Create new account         |
| POST   | `/auth/login`     | Authenticate user          |
| POST   | `/auth/refresh`   | Get a new access token     |
| POST   | `/auth/logout`    | Invalidate all user tokens |

---

## 📋 Notes

- **All responses are JSON.**
- **Access tokens** expire after 30 minutes by default.
- **Refresh tokens** expire after 30 days by default.
- **Middleware** automatically validates access tokens and injects the `user_id` into the request context.
- **SMTP settings** are required if you want welcome emails at signup (otherwise can be disabled easily).