package transactions

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type Service interface {
	List(ctx context.Context, userID string, input *ListTransactionsInput) ([]TransactionResponse, *httpx.Pagination, error)
	Get(ctx context.Context, id, userID string) (*TransactionResponse, error)
	Create(ctx context.Context, userID string, req *CreateTransactionRequest) (*TransactionResponse, error)
	Update(ctx context.Context, id, userID string, req *UpdateTransactionRequest) (*TransactionResponse, error)
	Delete(ctx context.Context, id, userID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, userID string, input *ListTransactionsInput) ([]TransactionResponse, *httpx.Pagination, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, nil, httpx.NewBadRequestError("invalid user ID")
	}

	var accID pgtype.UUID
	if input.AccountID != nil && *input.AccountID != "" {
		accID, err = common.StringToUUID(*input.AccountID)
		if err != nil {
			return nil, nil, httpx.NewBadRequestError("invalid account ID")
		}
	}

	var catID pgtype.UUID
	if input.CategoryID != nil && *input.CategoryID != "" {
		catID, err = common.StringToUUID(*input.CategoryID)
		if err != nil {
			return nil, nil, httpx.NewBadRequestError("invalid category ID")
		}
	}

	var startDate pgtype.Date
	if input.StartDate != nil && *input.StartDate != "" {
		startDate, err = common.StringToDate(*input.StartDate)
		if err != nil {
			return nil, nil, httpx.NewBadRequestError("invalid start_date")
		}
	}

	var endDate pgtype.Date
	if input.EndDate != nil && *input.EndDate != "" {
		endDate, err = common.StringToDate(*input.EndDate)
		if err != nil {
			return nil, nil, httpx.NewBadRequestError("invalid end_date")
		}
	}

	limit := int32(50)
	if input.Limit != nil && *input.Limit > 0 {
		limit = *input.Limit
	}

	offset := int32(0)
	if input.Offset != nil && *input.Offset >= 0 {
		offset = *input.Offset
	}

	countParams := sqlc.CountTransactionsParams{
		UserID:          uid,
		AccountID:       accID,
		CategoryID:      catID,
		TransactionType: input.TransactionType,
		StartDate:       startDate,
		EndDate:         endDate,
		Search:          input.Search,
	}

	total, err := s.repo.CountTransactions(ctx, countParams)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count transactions", "error", err)
		return nil, nil, httpx.NewInternalError("could not count transactions")
	}

	rows, err := s.repo.ListTransactions(ctx, sqlc.ListTransactionsParams{
		UserID:          uid,
		AccountID:       accID,
		CategoryID:      catID,
		TransactionType: input.TransactionType,
		StartDate:       startDate,
		EndDate:         endDate,
		Search:          input.Search,
		LimitVal:        limit,
		OffsetVal:       offset,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to list transactions", "error", err)
		return nil, nil, httpx.NewInternalError("could not list transactions")
	}

	result := make([]TransactionResponse, len(rows))
	for i, r := range rows {
		result[i] = transactionResponse(r)
	}

	pagination := &httpx.Pagination{
		Limit:   limit,
		Records: total,
		HasPrev: offset > 0,
		HasNext: int64(offset+limit) < total,
	}

	return result, pagination, nil
}

func (s *service) Get(ctx context.Context, id, userID string) (*TransactionResponse, error) {
	txID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid transaction ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	row, err := s.repo.GetTransactionByID(ctx, txID, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get transaction", "id", id, "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("transaction not found")
		}
		return nil, httpx.NewInternalError("could not retrieve transaction")
	}

	resp := transactionDetailResponse(*row)
	return &resp, nil
}

func (s *service) Create(ctx context.Context, userID string, req *CreateTransactionRequest) (*TransactionResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}
	accID, err := common.StringToUUID(req.AccountID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid account ID")
	}

	// Verify account ownership
	acc, err := s.repo.GetAccountByID(ctx, accID, uid)
	if err != nil {
		return nil, httpx.NewBadRequestError("account not found or not owned by user")
	}

	var catID pgtype.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		catID, err = common.StringToUUID(*req.CategoryID)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid category ID")
		}
	}

	txDate, err := common.StringToDate(req.TransactionDate)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid transaction date")
	}

	var postedDate pgtype.Date
	if req.PostedDate != nil && *req.PostedDate != "" {
		postedDate, err = common.StringToDate(*req.PostedDate)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid posted date")
		}
	}

	// Exact decimal parsing & sign normalization
	decAmount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid amount format")
	}

	txType := req.TransactionType
	if txType == "debit" && decAmount.IsPositive() {
		decAmount = decAmount.Neg()
	} else if txType == "credit" && decAmount.IsNegative() {
		decAmount = decAmount.Abs()
	}

	amountNumeric, err := common.StringToNumeric(decAmount.StringFixed(4))
	if err != nil {
		return nil, httpx.NewInternalError("failed to format amount")
	}

	status := req.Status
	if status == "" {
		status = "posted"
	}
	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = "manual"
	}
	currency := req.Currency
	if currency == "" {
		currency = acc.Currency
	}

	created, err := s.repo.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		UserID:                uid,
		AccountID:             accID,
		CategoryID:            catID,
		Amount:                amountNumeric,
		Currency:              currency,
		TransactionDate:       txDate,
		PostedDate:            postedDate,
		MerchantName:          req.MerchantName,
		Description:           req.Description,
		Notes:                 req.Notes,
		TransactionType:       txType,
		Status:                status,
		SourceType:            sourceType,
		ExternalTransactionID: req.ExternalTransactionID,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create transaction", "error", err)
		return nil, httpx.NewInternalError("could not create transaction")
	}

	// Update account balance based on asset vs liability rules
	balanceDelta := decAmount
	if acc.AccountType == "credit" || acc.AccountType == "loan" || acc.AccountType == "other_liability" {
		balanceDelta = decAmount.Neg()
	}
	deltaNumeric, _ := common.StringToNumeric(balanceDelta.StringFixed(4))
	_, _ = s.repo.UpdateAccountBalance(ctx, accID, uid, deltaNumeric)

	return s.Get(ctx, common.UUIDToString(created.ID), userID)
}

func (s *service) Update(ctx context.Context, id, userID string, req *UpdateTransactionRequest) (*TransactionResponse, error) {
	txID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid transaction ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	// Fetch existing
	existing, err := s.repo.GetTransactionByID(ctx, txID, uid)
	if err != nil {
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("transaction not found")
		}
		return nil, httpx.NewInternalError("could not load transaction")
	}

	var accID pgtype.UUID
	if req.AccountID != nil && *req.AccountID != "" {
		accID, err = common.StringToUUID(*req.AccountID)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid account ID")
		}
	}

	var catID pgtype.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		catID, err = common.StringToUUID(*req.CategoryID)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid category ID")
		}
	}

	var txDate pgtype.Date
	if req.TransactionDate != nil && *req.TransactionDate != "" {
		txDate, err = common.StringToDate(*req.TransactionDate)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid transaction date")
		}
	}

	var postedDate pgtype.Date
	if req.PostedDate != nil && *req.PostedDate != "" {
		postedDate, err = common.StringToDate(*req.PostedDate)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid posted date")
		}
	}

	var amountNumeric pgtype.Numeric
	if req.Amount != nil {
		decAmount, err := decimal.NewFromString(*req.Amount)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid amount format")
		}
		txType := existing.TransactionType
		if req.TransactionType != nil {
			txType = *req.TransactionType
		}
		if txType == "debit" && decAmount.IsPositive() {
			decAmount = decAmount.Neg()
		} else if txType == "credit" && decAmount.IsNegative() {
			decAmount = decAmount.Abs()
		}
		amountNumeric, _ = common.StringToNumeric(decAmount.StringFixed(4))
	}

	_, err = s.repo.UpdateTransaction(ctx, sqlc.UpdateTransactionParams{
		ID:              txID,
		UserID:          uid,
		AccountID:       accID,
		CategoryID:      catID,
		Amount:          amountNumeric,
		Currency:        req.Currency,
		TransactionDate: txDate,
		PostedDate:      postedDate,
		MerchantName:    req.MerchantName,
		Description:     req.Description,
		Notes:           req.Notes,
		TransactionType: req.TransactionType,
		Status:          req.Status,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to update transaction", "id", id, "error", err)
		return nil, httpx.NewInternalError("could not update transaction")
	}

	return s.Get(ctx, id, userID)
}

func (s *service) Delete(ctx context.Context, id, userID string) error {
	txID, err := common.StringToUUID(id)
	if err != nil {
		return httpx.NewBadRequestError("invalid transaction ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return httpx.NewBadRequestError("invalid user ID")
	}

	// Fetch transaction to revert account balance change
	existing, err := s.repo.GetTransactionByID(ctx, txID, uid)
	if err != nil {
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return httpx.NewNotFoundError("transaction not found")
		}
		return httpx.NewInternalError("could not find transaction")
	}

	acc, err := s.repo.GetAccountByID(ctx, existing.AccountID, uid)
	if err == nil && acc != nil {
		existingDec, _ := decimal.NewFromString(common.NumericToString(existing.Amount))
		revertDelta := existingDec.Neg()
		if acc.AccountType == "credit" || acc.AccountType == "loan" || acc.AccountType == "other_liability" {
			revertDelta = existingDec
		}
		revertNumeric, _ := common.StringToNumeric(revertDelta.StringFixed(4))
		_, _ = s.repo.UpdateAccountBalance(ctx, existing.AccountID, uid, revertNumeric)
	}

	if err := s.repo.DeleteTransaction(ctx, txID, uid); err != nil {
		slog.ErrorContext(ctx, "failed to delete transaction", "id", id, "error", err)
		return httpx.NewInternalError("could not delete transaction")
	}
	return nil
}
