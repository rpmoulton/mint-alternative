package transactions

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type Repository interface {
	ListTransactions(ctx context.Context, params sqlc.ListTransactionsParams) ([]sqlc.ListTransactionsRow, error)
	CountTransactions(ctx context.Context, params sqlc.CountTransactionsParams) (int64, error)
	GetTransactionByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.GetTransactionByIDRow, error)
	CreateTransaction(ctx context.Context, params sqlc.CreateTransactionParams) (*sqlc.Transaction, error)
	UpdateTransaction(ctx context.Context, params sqlc.UpdateTransactionParams) (*sqlc.Transaction, error)
	DeleteTransaction(ctx context.Context, id, userID pgtype.UUID) error
	GetAccountByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.GetAccountByIDRow, error)
	UpdateAccountBalance(ctx context.Context, id, userID pgtype.UUID, amountDelta pgtype.Numeric) (*sqlc.Account, error)
}

type repository struct {
	query *sqlc.Queries
}

func NewRepository(pg db.DB) Repository {
	return &repository{
		query: pg.RootStore().Queries(),
	}
}

func (r *repository) ListTransactions(ctx context.Context, params sqlc.ListTransactionsParams) ([]sqlc.ListTransactionsRow, error) {
	rows, err := r.query.ListTransactions(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	return rows, nil
}

func (r *repository) CountTransactions(ctx context.Context, params sqlc.CountTransactionsParams) (int64, error) {
	count, err := r.query.CountTransactions(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("count transactions: %w", err)
	}
	return count, nil
}

func (r *repository) GetTransactionByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.GetTransactionByIDRow, error) {
	row, err := r.query.GetTransactionByID(ctx, sqlc.GetTransactionByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get transaction %v: %w", id, err)
	}
	return &row, nil
}

func (r *repository) CreateTransaction(ctx context.Context, params sqlc.CreateTransactionParams) (*sqlc.Transaction, error) {
	tx, err := r.query.CreateTransaction(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}
	return &tx, nil
}

func (r *repository) UpdateTransaction(ctx context.Context, params sqlc.UpdateTransactionParams) (*sqlc.Transaction, error) {
	tx, err := r.query.UpdateTransaction(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}
	return &tx, nil
}

func (r *repository) DeleteTransaction(ctx context.Context, id, userID pgtype.UUID) error {
	err := r.query.DeleteTransaction(ctx, sqlc.DeleteTransactionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete transaction %v: %w", id, err)
	}
	return nil
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

func (r *repository) UpdateAccountBalance(ctx context.Context, id, userID pgtype.UUID, amountDelta pgtype.Numeric) (*sqlc.Account, error) {
	acc, err := r.query.UpdateAccountBalance(ctx, sqlc.UpdateAccountBalanceParams{
		ID:          id,
		UserID:      userID,
		AmountDelta: amountDelta,
	})
	if err != nil {
		return nil, fmt.Errorf("update account balance: %w", err)
	}
	return &acc, nil
}
