package auth

import (
	"log"
	"net/http"

	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/labstack/echo/v4"
)

func AuthenticationMiddleware() echo.MiddlewareFunc {
	authConfig, err := config.LoadAuthConfig()
	if err != nil {
		log.Printf("Authentication config error (fail-closed): %v", err)
		return failClosedAuthMiddleware("authentication configuration error")
	}

	if !authConfig.OAuth.Enabled {
		return APIKeyAuthMiddleware(authConfig.APIKey)
	}

	oauthService, err := NewOAuthService(authConfig)
	if err != nil {
		log.Printf("OAuth service initialization error (fail-closed): %v", err)
		return failClosedAuthMiddleware("authentication service initialization error")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return OAuthMiddleware(authConfig, oauthService)(next)(c)
		}
	}
}

func failClosedAuthMiddleware(message string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusInternalServerError, message)
		}
	}
}
