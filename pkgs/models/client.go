package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	Email          string
	Verified       bool
	FailedAttempts int
	Locked         bool
	HashedPassword string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
