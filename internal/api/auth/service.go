package auth

import (
	"context"
	"log/slog"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
	"github.com/rpmoulton/mint-alternative/pkg/argon2id"
)

type Service interface {
	Register(ctx context.Context, user *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, credentials *LoginRequest) (*AuthResponse, error)
	Me(ctx context.Context, userID string) (*UserResponse, error)
	Refresh(ctx context.Context, email, userID string) (*jwt.Tokens, error)
}

type service struct {
	repo  Repository
	token jwt.Service
	argon *argon2id.Config
}

func NewService(repo Repository) Service {
	return &service{repo: repo, token: jwt.GetService(), argon: argon2id.DefaultConfig()}
}

func (s *service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := s.argon.HashPassword(req.Password)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash user password", "error", err)
		return nil, httpx.NewInternalError("could not register user")
	}

	currency := req.PreferredCurrency
	if currency == "" {
		currency = "USD"
	}
	tz := req.Timezone
	if tz == "" {
		tz = "UTC"
	}

	user, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		Email:             req.Email,
		PasswordHash:      hashedPassword,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		PreferredCurrency: currency,
		Timezone:          tz,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create user", "error", err)
		mapped := db.WrapDBError(ctx, err)
		if db.IsAlreadyExistsError(mapped) {
			return nil, httpx.NewConflictError("a user with this email already exists")
		}
		return nil, httpx.NewInternalError("could not register user")
	}

	userResp := userResponse(*user)
	tokens, err := s.token.IssueTokenPair(userResp.ID, user.Email, "customer")
	if err != nil {
		slog.ErrorContext(ctx, "failed to issue tokens", "error", err)
		return nil, httpx.NewInternalError("could not issue tokens")
	}

	return &AuthResponse{
		User:   userResp,
		Tokens: *tokens,
	}, nil
}

func (s *service) Login(ctx context.Context, credentials *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, credentials.Email)
	if err != nil {
		slog.WarnContext(ctx, "login email lookup failed", "error", err)
		return nil, httpx.NewUnauthorizedError("invalid credentials")
	}

	match, err := s.argon.ComparePasswordAndHash(credentials.Password, user.PasswordHash)
	if err != nil {
		slog.ErrorContext(ctx, "failed to verify password", "error", err)
		return nil, httpx.NewInternalError("could not authenticate")
	}
	if !match {
		return nil, httpx.NewUnauthorizedError("invalid credentials")
	}

	userResp := userResponse(*user)
	tokens, err := s.token.IssueTokenPair(userResp.ID, user.Email, "customer")
	if err != nil {
		slog.ErrorContext(ctx, "failed to issue tokens", "error", err)
		return nil, httpx.NewInternalError("could not issue tokens")
	}

	return &AuthResponse{
		User:   userResp,
		Tokens: *tokens,
	}, nil
}

func (s *service) Me(ctx context.Context, userID string) (*UserResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	user, err := s.repo.FindUserByID(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load current user", "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("user does not exist")
		}
		return nil, httpx.NewInternalError("could not load user")
	}

	resp := userResponse(*user)
	return &resp, nil
}

func (s *service) Refresh(ctx context.Context, email, userID string) (*jwt.Tokens, error) {
	tokens, err := s.token.IssueTokenPair(userID, email, "customer")
	if err != nil {
		slog.WarnContext(ctx, "refresh tokens failed", "error", err)
		return nil, httpx.NewUnauthorizedError("could not refresh token")
	}
	return tokens, nil
}
