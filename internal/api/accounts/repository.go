package accounts

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type Repository interface {
	ListAccounts(ctx context.Context, userID pgtype.UUID, includeArchived *bool) ([]sqlc.ListAccountsByUserRow, error)
	GetAccountByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.GetAccountByIDRow, error)
	CreateAccount(ctx context.Context, params sqlc.CreateAccountParams) (*sqlc.Account, error)
	UpdateAccount(ctx context.Context, params sqlc.UpdateAccountParams) (*sqlc.Account, error)
	DeleteAccount(ctx context.Context, id, userID pgtype.UUID) error
}

type repository struct {
	query *sqlc.Queries
}

func NewRepository(pg db.DB) Repository {
	return &repository{
		query: pg.RootStore().Queries(),
	}
}

func (r *repository) ListAccounts(ctx context.Context, userID pgtype.UUID, includeArchived *bool) ([]sqlc.ListAccountsByUserRow, error) {
	rows, err := r.query.ListAccountsByUser(ctx, sqlc.ListAccountsByUserParams{
		UserID:          userID,
		IncludeArchived: includeArchived,
	})
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return rows, nil
}

func (r *repository) GetAccountByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.GetAccountByIDRow, error) {
	row, err := r.query.GetAccountByID(ctx, sqlc.GetAccountByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get account %v: %w", id, err)
	}
	return &row, nil
}

func (r *repository) CreateAccount(ctx context.Context, params sqlc.CreateAccountParams) (*sqlc.Account, error) {
	acc, err := r.query.CreateAccount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return &acc, nil
}

func (r *repository) UpdateAccount(ctx context.Context, params sqlc.UpdateAccountParams) (*sqlc.Account, error) {
	acc, err := r.query.UpdateAccount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}
	return &acc, nil
}

func (r *repository) DeleteAccount(ctx context.Context, id, userID pgtype.UUID) error {
	err := r.query.DeleteAccount(ctx, sqlc.DeleteAccountParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}
