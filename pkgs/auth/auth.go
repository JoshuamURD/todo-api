package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// RSAKeys holds the public and private keys for JWT signing
type RSAKeys struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// JWTAuthService provides JWT authentication functionality
type JWTAuthService struct {
	keys RSAKeys
}

// JWTClaims struct is used to store the JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
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

// GenerateTokens generates both access and refresh tokens
// It takes a context and a user ID and returns the access and refresh tokens
func (j *JWTAuthService) GenerateTokens(ctx context.Context, userID string) (accessToken, refreshToken string, err error) {
	// Generate access token (short-lived)
	// It takes a context and a user ID and returns the access and refresh tokens
	accessClaims := JWTClaims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Generate access token with helper function
	accessToken, err = j.generateToken(accessClaims)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (longer-lived)
	refreshClaims := JWTClaims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Generate refresh token with helper function
	refreshToken, err = j.generateToken(refreshClaims)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Return the access and refresh tokens
	return accessToken, refreshToken, nil
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

// ValidateToken validates the JWT token and returns the claims
func (j *JWTAuthService) ValidateToken(tokenString string) (*JWTClaims, error) {

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

// RefreshAccessToken generates a new access token using a valid refresh token
func (j *JWTAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Ensure the token is a refresh token
	if claims.Type != "refresh" {
		return "", errors.New("invalid token type")
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

	return j.generateToken(accessClaims)
}
