package accounts

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type Service interface {
	List(ctx context.Context, userID string, includeArchived *bool) ([]AccountResponse, error)
	Get(ctx context.Context, id, userID string) (*AccountResponse, error)
	Create(ctx context.Context, userID string, req *CreateAccountRequest) (*AccountResponse, error)
	Update(ctx context.Context, id, userID string, req *UpdateAccountRequest) (*AccountResponse, error)
	Delete(ctx context.Context, id, userID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, userID string, includeArchived *bool) ([]AccountResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	rows, err := s.repo.ListAccounts(ctx, uid, includeArchived)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list accounts", "error", err)
		return nil, httpx.NewInternalError("could not list accounts")
	}

	result := make([]AccountResponse, len(rows))
	for i, r := range rows {
		result[i] = accountResponse(r)
	}
	return result, nil
}

func (s *service) Get(ctx context.Context, id, userID string) (*AccountResponse, error) {
	accID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid account ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	row, err := s.repo.GetAccountByID(ctx, accID, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get account", "id", id, "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("account not found")
		}
		return nil, httpx.NewInternalError("could not retrieve account")
	}

	resp := accountDetailResponse(*row)
	return &resp, nil
}

func (s *service) Create(ctx context.Context, userID string, req *CreateAccountRequest) (*AccountResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	var instID pgtype.UUID
	if req.InstitutionID != nil && *req.InstitutionID != "" {
		instID, err = common.StringToUUID(*req.InstitutionID)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid institution ID")
		}
	}

	currBal, err := common.StringToNumeric(req.CurrentBalance)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid current balance format")
	}

	var availBal pgtype.Numeric
	if req.AvailableBalance != nil {
		availBal, err = common.StringToNumeric(*req.AvailableBalance)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid available balance format")
		}
	}

	var credLimit pgtype.Numeric
	if req.CreditLimit != nil {
		credLimit, err = common.StringToNumeric(*req.CreditLimit)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid credit limit format")
		}
	}

	var intRate pgtype.Numeric
	if req.InterestRate != nil {
		intRate, err = common.StringToNumeric(*req.InterestRate)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid interest rate format")
		}
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	created, err := s.repo.CreateAccount(ctx, sqlc.CreateAccountParams{
		UserID:           uid,
		InstitutionID:    instID,
		Name:             req.Name,
		AccountType:      req.AccountType,
		AccountSubtype:   req.AccountSubtype,
		Currency:         currency,
		CurrentBalance:   currBal,
		AvailableBalance: availBal,
		CreditLimit:      credLimit,
		InterestRate:     intRate,
		SourceType:       "manual",
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create account", "error", err)
		return nil, httpx.NewInternalError("could not create account")
	}

	// Fetch with joined institution details
	return s.Get(ctx, common.UUIDToString(created.ID), userID)
}

func (s *service) Update(ctx context.Context, id, userID string, req *UpdateAccountRequest) (*AccountResponse, error) {
	accID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid account ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	var instID pgtype.UUID
	if req.InstitutionID != nil && *req.InstitutionID != "" {
		instID, err = common.StringToUUID(*req.InstitutionID)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid institution ID")
		}
	}

	var currBal pgtype.Numeric
	if req.CurrentBalance != nil {
		currBal, err = common.StringToNumeric(*req.CurrentBalance)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid current balance")
		}
	}

	var availBal pgtype.Numeric
	if req.AvailableBalance != nil {
		availBal, err = common.StringToNumeric(*req.AvailableBalance)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid available balance")
		}
	}

	var credLimit pgtype.Numeric
	if req.CreditLimit != nil {
		credLimit, err = common.StringToNumeric(*req.CreditLimit)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid credit limit")
		}
	}

	var intRate pgtype.Numeric
	if req.InterestRate != nil {
		intRate, err = common.StringToNumeric(*req.InterestRate)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid interest rate")
		}
	}

	_, err = s.repo.UpdateAccount(ctx, sqlc.UpdateAccountParams{
		ID:               accID,
		UserID:           uid,
		InstitutionID:    instID,
		Name:             req.Name,
		AccountType:      req.AccountType,
		AccountSubtype:   req.AccountSubtype,
		Currency:         req.Currency,
		CurrentBalance:   currBal,
		AvailableBalance: availBal,
		CreditLimit:      credLimit,
		InterestRate:     intRate,
		Archive:          req.Archive,
		Unarchive:        req.Unarchive,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to update account", "id", id, "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("account not found")
		}
		return nil, httpx.NewInternalError("could not update account")
	}

	return s.Get(ctx, id, userID)
}

func (s *service) Delete(ctx context.Context, id, userID string) error {
	accID, err := common.StringToUUID(id)
	if err != nil {
		return httpx.NewBadRequestError("invalid account ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return httpx.NewBadRequestError("invalid user ID")
	}

	if err := s.repo.DeleteAccount(ctx, accID, uid); err != nil {
		slog.ErrorContext(ctx, "failed to delete account", "id", id, "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return httpx.NewNotFoundError("account not found")
		}
		return httpx.NewInternalError("could not delete account")
	}
	return nil
}
