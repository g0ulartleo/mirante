package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type OAuthService struct {
	config      *config.AuthConfig
	oauthConfig *oauth2.Config
}

type UserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    string `json:"id"`
}

type Claims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    string `json:"id"`
	jwt.RegisteredClaims
}

func NewOAuthService(authConfig *config.AuthConfig) (*OAuthService, error) {
	if !authConfig.OAuth.Enabled {
		return nil, fmt.Errorf("OAuth is not enabled")
	}

	clientID := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")

	if clientID == "" {
		return nil, fmt.Errorf("OAUTH_CLIENT_ID environment variable is required")
	}

	if clientSecret == "" {
		return nil, fmt.Errorf("OAUTH_CLIENT_SECRET environment variable is required")
	}

	var oauthConfig *oauth2.Config
	switch authConfig.OAuth.Provider {
	case "google":
		oauthConfig = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  authConfig.OAuth.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	case "github":
		oauthConfig = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  authConfig.OAuth.RedirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", authConfig.OAuth.Provider)
	}

	return &OAuthService{
		config:      authConfig,
		oauthConfig: oauthConfig,
	}, nil
}

func (s *OAuthService) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state)
}

func (s *OAuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.oauthConfig.Exchange(ctx, code)
}

func (s *OAuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := s.oauthConfig.Client(ctx, token)

	var userInfoURL string
	switch s.config.OAuth.Provider {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		userInfoURL = "https://api.github.com/user"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", s.config.OAuth.Provider)
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	if s.config.OAuth.Provider == "github" && userInfo.Email == "" {
		if err := s.getGitHubEmail(client, &userInfo); err != nil {
			return nil, fmt.Errorf("failed to get GitHub email: %w", err)
		}
	}

	return &userInfo, nil
}

func (s *OAuthService) getGitHubEmail(client *http.Client, userInfo *UserInfo) error {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return err
	}

	for _, email := range emails {
		if email.Primary {
			userInfo.Email = email.Email
			return nil
		}
	}

	return fmt.Errorf("no primary email found")
}

func (s *OAuthService) IsEmailAllowed(email string) bool {
	return s.config.OAuth.IsEmailAllowed(email)
}

func (s *OAuthService) GenerateJWT(userInfo *UserInfo) (string, error) {
	jwtSecret := os.Getenv("OAUTH_JWT_SECRET")
	if jwtSecret == "" {
		return "", fmt.Errorf("OAUTH_JWT_SECRET environment variable is required")
	}

	if err := validateJWTSecret(jwtSecret); err != nil {
		return "", fmt.Errorf("JWT secret validation failed: %w", err)
	}

	duration, err := time.ParseDuration(s.config.OAuth.SessionTimeout)
	if err != nil {
		duration = 24 * time.Hour
	}

	claims := Claims{
		Email: userInfo.Email,
		Name:  userInfo.Name,
		ID:    userInfo.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mirante",
			Subject:   userInfo.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func (s *OAuthService) ValidateJWT(tokenString string) (*Claims, error) {
	jwtSecret := os.Getenv("OAUTH_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("OAUTH_JWT_SECRET environment variable is required")
	}

	if err := validateJWTSecret(jwtSecret); err != nil {
		return nil, fmt.Errorf("JWT secret validation failed: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if !s.IsEmailAllowed(claims.Email) {
			return nil, fmt.Errorf("email no longer allowed: %s", claims.Email)
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func validateJWTSecret(secret string) error {
	if len(secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}
	return nil
}
