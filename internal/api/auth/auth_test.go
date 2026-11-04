package auth

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/config"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
	"github.com/rpmoulton/mint-alternative/pkg/argon2id"
)

type mockAuthRepo struct {
	usersByEmail map[string]*sqlc.User
	usersByID    map[string]*sqlc.User
}

func (m *mockAuthRepo) FindUserByEmail(ctx context.Context, email string) (*sqlc.User, error) {
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

func (m *mockAuthRepo) FindUserByID(ctx context.Context, id pgtype.UUID) (*sqlc.User, error) {
	idStr := string(id.Bytes[:])
	u, ok := m.usersByID[idStr]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return u, nil
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error) {
	u := &sqlc.User{
		ID:                pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
		Email:             params.Email,
		PasswordHash:      params.PasswordHash,
		FirstName:         params.FirstName,
		LastName:          params.LastName,
		PreferredCurrency: params.PreferredCurrency,
		Timezone:          params.Timezone,
	}
	m.usersByEmail[params.Email] = u
	m.usersByID[string(u.ID.Bytes[:])] = u
	return u, nil
}

func TestAuthService_PasswordHashingAndVerification(t *testing.T) {
	cfg := argon2id.DefaultConfig()
	password := "SecretPass123!"

	hash, err := cfg.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	match, err := cfg.ComparePasswordAndHash(password, hash)
	if err != nil || !match {
		t.Fatalf("expected password to match hash, match=%v, err=%v", match, err)
	}

	wrongMatch, _ := cfg.ComparePasswordAndHash("WrongPassword!", hash)
	if wrongMatch {
		t.Fatalf("expected wrong password to not match")
	}
}

func TestAuthService_RegistrationAndLogin(t *testing.T) {
	err := jwt.InitJWT(&config.Jwt{
		PubKeyPath:        "../../../cert/jwt-pub.pem",
		PvtKeyPath:        "../../../cert/jwt-pvt.pem",
		AccessExpiryTime:  300,
		RefreshExpiryTime: 1800,
	})
	if err != nil {
		t.Fatalf("failed to init jwt: %v", err)
	}

	repo := &mockAuthRepo{
		usersByEmail: make(map[string]*sqlc.User),
		usersByID:    make(map[string]*sqlc.User),
	}
	svc := NewService(repo)

	ctx := context.Background()
	regReq := &RegisterRequest{
		Email:             "test@example.com",
		Password:          "SecurePass123",
		FirstName:         "John",
		LastName:          "Doe",
		PreferredCurrency: "USD",
		Timezone:          "America/New_York",
	}

	authResp, err := svc.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	if authResp.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", authResp.User.Email)
	}
	if authResp.Tokens.AccessToken == "" {
		t.Fatalf("expected access token to be present")
	}

	// Test login with valid credentials
	loginResp, err := svc.Login(ctx, &LoginRequest{
		Email:    "test@example.com",
		Password: "SecurePass123",
	})
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	if loginResp.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", loginResp.User.Email)
	}

	// Test login with invalid credentials
	_, err = svc.Login(ctx, &LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword999",
	})
	if err == nil {
		t.Fatalf("expected error for invalid password")
	}
}
