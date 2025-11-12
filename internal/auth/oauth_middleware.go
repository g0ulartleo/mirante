package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/g0ulartleo/mirante-alerts/internal/config"
	"github.com/labstack/echo/v4"
)

func OAuthMiddleware(authConfig *config.AuthConfig, oauthService *OAuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !authConfig.OAuth.Enabled {
				return APIKeyAuthMiddleware(authConfig.APIKey)(next)(c)
			}

			var tokenString string

			authHeader := c.Request().Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			}

			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
			}

			claims, err := oauthService.ValidateJWT(tokenString)
			if err != nil {
				log.Printf("JWT validation error: %v", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authentication token")
			}

			c.Set("user", claims)
			c.Set("user_email", claims.Email)

			return next(c)
		}
	}
}

func APIKeyAuthMiddleware(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if apiKey == "" {
				log.Printf("No API key configured for authentication")
				return echo.NewHTTPError(http.StatusInternalServerError, "No authentication method configured")
			}

			requestAPIKey := strings.TrimSpace(c.Request().Header.Get("X-API-Key"))
			if requestAPIKey == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "API key required")
			}

			if requestAPIKey != apiKey {
				log.Printf("Invalid API key attempt from %s", c.RealIP())
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
			}

			return next(c)
		}
	}
}
