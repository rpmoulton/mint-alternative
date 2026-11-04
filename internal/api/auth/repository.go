package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*sqlc.User, error)
	FindUserByID(ctx context.Context, id pgtype.UUID) (*sqlc.User, error)
	CreateUser(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error)
}

type repository struct {
	query *sqlc.Queries
}

func NewRepository(pg db.DB) Repository {
	return &repository{
		query: pg.RootStore().Queries(),
	}
}

func (r *repository) FindUserByEmail(ctx context.Context, email string) (*sqlc.User, error) {
	user, err := r.query.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find user by email %q: %w", email, err)
	}
	return &user, nil
}

func (r *repository) FindUserByID(ctx context.Context, id pgtype.UUID) (*sqlc.User, error) {
	user, err := r.query.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find user by id %v: %w", id, err)
	}
	return &user, nil
}

func (r *repository) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error) {
	user, err := r.query.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}
