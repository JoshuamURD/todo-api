package auth

import (
	"context"
	"joshuamURD/go-auth-api/pkgs/models"
)

type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (models.User, error)
}
