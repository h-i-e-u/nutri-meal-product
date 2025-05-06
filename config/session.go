package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
)

var store *session.Store

func InitSession() {
    store = session.New(session.Config{
        Expiration:     24 * time.Hour,  // Session expiration
        KeyLookup:      "cookie:session_id",
        CookieSecure:   false,  // Set to true in production with HTTPS
        CookieHTTPOnly: true,
        CookieSameSite: "Lax",
    })
}

func GetStore() *session.Store {
    return store
}