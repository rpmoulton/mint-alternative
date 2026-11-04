package auth

import (
	"time"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

type RegisterRequest struct {
	Email             string `json:"email"              validate:"required,email,lowercase"`
	Password          string `json:"password"           validate:"required,min=8"`
	FirstName         string `json:"first_name"         validate:"omitempty,max=100"`
	LastName          string `json:"last_name"          validate:"omitempty,max=100"`
	PreferredCurrency string `json:"preferred_currency" validate:"omitempty,len=3"`
	Timezone          string `json:"timezone"           validate:"omitempty,max=50"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email,lowercase"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterInput struct {
	Body RegisterRequest `body:"required"`
}

type LoginInput struct {
	Body LoginRequest `body:"required"`
}

type EmptyInput struct{}

// UserResponse is the client-safe projection of an auth user. Password hashes
// and other authentication internals must never be serialized.
type UserResponse struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	PreferredCurrency string    `json:"preferred_currency"`
	Timezone          string    `json:"timezone"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AuthResponse struct {
	User   UserResponse `json:"user"`
	Tokens jwt.Tokens   `json:"tokens"`
}

func userResponse(user sqlc.User) UserResponse {
	return UserResponse{
		ID:                common.UUIDToString(user.ID),
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		PreferredCurrency: user.PreferredCurrency,
		Timezone:          user.Timezone,
		CreatedAt:         user.CreatedAt.Time,
		UpdatedAt:         user.UpdatedAt.Time,
	}
}
