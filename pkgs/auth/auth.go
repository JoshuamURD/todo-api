package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService is an interface that defines the methods for the authentication service
type AuthService interface {
	// Authenticate creates authentication state for a user and handles the response
	Authenticate(ctx context.Context, userID string, w http.ResponseWriter) (*AuthResponse, error)
	// Refresh updates the authentication state
	RefreshAuth(ctx context.Context, refreshToken string) (*AuthResponse, error)
}

// AuthClaims represents generic authentication claims
type AuthClaims struct {
	UserID string
	// Add other generic claims as needed
}

// RSAKeys holds the public and private keys for JWT signing
type RSAKeys struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// JWTAuthService implements AuthService using JWT
type JWTAuthService struct {
	keys RSAKeys
}

// JWTClaims struct is used to store the JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// NewJWTAuthService creates a new JWT authentication service with RSA keys
func NewJWTAuthService(privateKey *rsa.PrivateKey) *JWTAuthService {
	return &JWTAuthService{
		keys: RSAKeys{
			privateKey: privateKey,
			publicKey:  &privateKey.PublicKey,
		},
	}
}

// Authenticate generates JWTs, both access and refresh tokens
// It takes a context and a user ID and returns the access and refresh tokens
func (j *JWTAuthService) Authenticate(ctx context.Context, userID string, w http.ResponseWriter) (*AuthResponse, error) {
	// Generate access token (short-lived)
	accessClaims := JWTClaims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := j.generateToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (longer-lived)
	refreshClaims := JWTClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken, err := j.generateToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Set only refresh token in HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth/refresh", // Restrict to refresh endpoint
	})

	// Return access token in response body
	return &AuthResponse{
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	}, nil
}

// generateToken helper function to create signed tokens
func (j *JWTAuthService) generateToken(claims JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if j.keys.privateKey == nil {
		return "", errors.New("private key is not initialized")
	}

	// Sign the token with the private key
	return token.SignedString(j.keys.privateKey)
}

// Validate validates the JWT token and returns the claims
func (j *JWTAuthService) Validate(tokenString string) (*JWTClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.keys.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// Refresh generates a new access token using a valid refresh token
func (j *JWTAuthService) RefreshAuth(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	claims, err := j.Validate(refreshToken)
	if err != nil {
		return nil, err
	}

	// Ensure the token is a refresh token
	if claims.Type != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Generate new access token
	accessClaims := JWTClaims{
		UserID: claims.UserID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := j.generateToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	authResponse := AuthResponse{
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	}

	return &authResponse, nil
}
