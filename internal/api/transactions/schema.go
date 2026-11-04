package transactions

import (
	"time"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type TransactionResponse struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	AccountID             string    `json:"account_id"`
	AccountName           string    `json:"account_name"`
	CategoryID            *string   `json:"category_id,omitempty"`
	CategoryName          *string   `json:"category_name,omitempty"`
	CategoryColor         *string   `json:"category_color,omitempty"`
	CategoryIcon          *string   `json:"category_icon,omitempty"`
	CategoryType          *string   `json:"category_type,omitempty"`
	Amount                string    `json:"amount"`
	Currency              string    `json:"currency"`
	TransactionDate       string    `json:"transaction_date"`
	PostedDate            *string   `json:"posted_date,omitempty"`
	MerchantName          *string   `json:"merchant_name,omitempty"`
	Description           string    `json:"description"`
	Notes                 *string   `json:"notes,omitempty"`
	TransactionType       string    `json:"transaction_type"`
	Status                string    `json:"status"`
	SourceType            string    `json:"source_type"`
	ExternalTransactionID *string   `json:"external_transaction_id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CreateTransactionRequest struct {
	AccountID             string  `json:"account_id"              validate:"required,uuid"`
	CategoryID            *string `json:"category_id"             validate:"omitempty,uuid"`
	Amount                string  `json:"amount"                  validate:"required"`
	Currency              string  `json:"currency"                validate:"omitempty,len=3"`
	TransactionDate       string  `json:"transaction_date"        validate:"required"` // YYYY-MM-DD
	PostedDate            *string `json:"posted_date"             validate:"omitempty"`
	MerchantName          *string `json:"merchant_name"           validate:"omitempty,max=100"`
	Description           string  `json:"description"             validate:"required,min=1,max=255"`
	Notes                 *string `json:"notes"                   validate:"omitempty,max=500"`
	TransactionType       string  `json:"transaction_type"        validate:"required,oneof=debit credit"`
	Status                string  `json:"status"                  validate:"omitempty,oneof=pending posted"`
	SourceType            string  `json:"source_type"             validate:"omitempty,oneof=manual imported"`
	ExternalTransactionID *string `json:"external_transaction_id" validate:"omitempty,max=100"`
}

type UpdateTransactionRequest struct {
	AccountID       *string `json:"account_id"       validate:"omitempty,uuid"`
	CategoryID      *string `json:"category_id"      validate:"omitempty,uuid"`
	Amount          *string `json:"amount"           validate:"omitempty"`
	Currency        *string `json:"currency"         validate:"omitempty,len=3"`
	TransactionDate *string `json:"transaction_date" validate:"omitempty"`
	PostedDate      *string `json:"posted_date"      validate:"omitempty"`
	MerchantName    *string `json:"merchant_name"    validate:"omitempty,max=100"`
	Description     *string `json:"description"      validate:"omitempty,min=1,max=255"`
	Notes           *string `json:"notes"            validate:"omitempty,max=500"`
	TransactionType *string `json:"transaction_type" validate:"omitempty,oneof=debit credit"`
	Status          *string `json:"status"           validate:"omitempty,oneof=pending posted"`
}

type ListTransactionsInput struct {
	AccountID       *string `query:"account_id"       validate:"omitempty,uuid"`
	CategoryID      *string `query:"category_id"      validate:"omitempty,uuid"`
	TransactionType *string `query:"transaction_type" validate:"omitempty,oneof=debit credit"`
	StartDate       *string `query:"start_date"       validate:"omitempty"`
	EndDate         *string `query:"end_date"         validate:"omitempty"`
	Search          *string `query:"search"           validate:"omitempty"`
	Limit           *int32  `query:"limit"            validate:"omitempty,min=1,max=100"`
	Offset          *int32  `query:"offset"           validate:"omitempty,min=0"`
}

type CreateTransactionInput struct {
	Body CreateTransactionRequest `body:"required"`
}

type UpdateTransactionInput struct {
	ID   string                   `path:"id"      validate:"required,uuid"`
	Body UpdateTransactionRequest `body:"required"`
}

type GetTransactionInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

type DeleteTransactionInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

func transactionResponse(row sqlc.ListTransactionsRow) TransactionResponse {
	var catID *string
	if row.CategoryID.Valid {
		s := common.UUIDToString(row.CategoryID)
		catID = &s
	}

	var postedDate *string
	if row.PostedDate.Valid {
		s := common.DateToString(row.PostedDate)
		postedDate = &s
	}

	return TransactionResponse{
		ID:                    common.UUIDToString(row.ID),
		UserID:                common.UUIDToString(row.UserID),
		AccountID:             common.UUIDToString(row.AccountID),
		AccountName:           row.AccountName,
		CategoryID:            catID,
		CategoryName:          row.CategoryName,
		CategoryColor:         row.CategoryColor,
		CategoryIcon:          row.CategoryIcon,
		CategoryType:          row.CategoryType,
		Amount:                common.NumericToString(row.Amount),
		Currency:              row.Currency,
		TransactionDate:       common.DateToString(row.TransactionDate),
		PostedDate:            postedDate,
		MerchantName:          row.MerchantName,
		Description:           row.Description,
		Notes:                 row.Notes,
		TransactionType:       row.TransactionType,
		Status:                row.Status,
		SourceType:            row.SourceType,
		ExternalTransactionID: row.ExternalTransactionID,
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
	}
}

func transactionDetailResponse(row sqlc.GetTransactionByIDRow) TransactionResponse {
	var catID *string
	if row.CategoryID.Valid {
		s := common.UUIDToString(row.CategoryID)
		catID = &s
	}

	var postedDate *string
	if row.PostedDate.Valid {
		s := common.DateToString(row.PostedDate)
		postedDate = &s
	}

	return TransactionResponse{
		ID:                    common.UUIDToString(row.ID),
		UserID:                common.UUIDToString(row.UserID),
		AccountID:             common.UUIDToString(row.AccountID),
		AccountName:           row.AccountName,
		CategoryID:            catID,
		CategoryName:          row.CategoryName,
		CategoryColor:         row.CategoryColor,
		CategoryIcon:          row.CategoryIcon,
		CategoryType:          row.CategoryType,
		Amount:                common.NumericToString(row.Amount),
		Currency:              row.Currency,
		TransactionDate:       common.DateToString(row.TransactionDate),
		PostedDate:            postedDate,
		MerchantName:          row.MerchantName,
		Description:           row.Description,
		Notes:                 row.Notes,
		TransactionType:       row.TransactionType,
		Status:                row.Status,
		SourceType:            row.SourceType,
		ExternalTransactionID: row.ExternalTransactionID,
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
	}
}
