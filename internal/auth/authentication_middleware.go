package auth

import (
	"log"

	"github.com/g0ulartleo/mirante-alerts/internal/config"
	"github.com/labstack/echo/v4"
)

func AuthenticationMiddleware() echo.MiddlewareFunc {
	authConfig, err := config.LoadAuthConfig()
	if err != nil {
		log.Printf("Error loading auth config: %v", err)
		return nil
	}

	oauthService, err := NewOAuthService(authConfig)
	if err != nil {
		log.Printf("Error creating OAuth service: %v", err)
		return nil
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return OAuthMiddleware(authConfig, oauthService)(next)(c)
		}
	}
}
