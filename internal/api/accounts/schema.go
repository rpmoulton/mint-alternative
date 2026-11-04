package accounts

import (
	"time"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type AccountResponse struct {
	ID                     string     `json:"id"`
	UserID                 string     `json:"user_id"`
	InstitutionID          *string    `json:"institution_id,omitempty"`
	InstitutionName        *string    `json:"institution_name,omitempty"`
	InstitutionDisplayName *string    `json:"institution_display_name,omitempty"`
	InstitutionLogoURL     *string    `json:"institution_logo_url,omitempty"`
	Name                   string     `json:"name"`
	AccountType            string     `json:"account_type"`
	AccountSubtype         string     `json:"account_subtype"`
	Currency               string     `json:"currency"`
	CurrentBalance         string     `json:"current_balance"`
	AvailableBalance       *string    `json:"available_balance,omitempty"`
	CreditLimit            *string    `json:"credit_limit,omitempty"`
	InterestRate           *string    `json:"interest_rate,omitempty"`
	SourceType             string     `json:"source_type"`
	ArchivedAt             *time.Time `json:"archived_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type CreateAccountRequest struct {
	InstitutionID    *string `json:"institution_id"    validate:"omitempty,uuid"`
	Name             string  `json:"name"              validate:"required,min=1,max=100"`
	AccountType      string  `json:"account_type"      validate:"required,oneof=depository credit loan investment other_asset other_liability"`
	AccountSubtype   string  `json:"account_subtype"   validate:"omitempty,max=50"`
	Currency         string  `json:"currency"          validate:"omitempty,len=3"`
	CurrentBalance   string  `json:"current_balance"   validate:"required"`
	AvailableBalance *string `json:"available_balance" validate:"omitempty"`
	CreditLimit      *string `json:"credit_limit"      validate:"omitempty"`
	InterestRate     *string `json:"interest_rate"     validate:"omitempty"`
}

type UpdateAccountRequest struct {
	InstitutionID    *string `json:"institution_id"    validate:"omitempty,uuid"`
	Name             *string `json:"name"              validate:"omitempty,min=1,max=100"`
	AccountType      *string `json:"account_type"      validate:"omitempty,oneof=depository credit loan investment other_asset other_liability"`
	AccountSubtype   *string `json:"account_subtype"   validate:"omitempty,max=50"`
	Currency         *string `json:"currency"          validate:"omitempty,len=3"`
	CurrentBalance   *string `json:"current_balance"   validate:"omitempty"`
	AvailableBalance *string `json:"available_balance" validate:"omitempty"`
	CreditLimit      *string `json:"credit_limit"      validate:"omitempty"`
	InterestRate     *string `json:"interest_rate"     validate:"omitempty"`
	Archive          *bool   `json:"archive"           validate:"omitempty"`
	Unarchive        *bool   `json:"unarchive"         validate:"omitempty"`
}

type CreateAccountInput struct {
	Body CreateAccountRequest `body:"required"`
}

type UpdateAccountInput struct {
	ID   string               `path:"id"      validate:"required,uuid"`
	Body UpdateAccountRequest `body:"required"`
}

type GetAccountInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

type DeleteAccountInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

type ListAccountsInput struct {
	IncludeArchived *bool `query:"include_archived" validate:"omitempty"`
}

func accountResponse(row sqlc.ListAccountsByUserRow) AccountResponse {
	var instID *string
	if row.InstitutionID.Valid {
		s := common.UUIDToString(row.InstitutionID)
		instID = &s
	}

	var availBal *string
	if row.AvailableBalance.Valid {
		s := common.NumericToString(row.AvailableBalance)
		availBal = &s
	}

	var credLimit *string
	if row.CreditLimit.Valid {
		s := common.NumericToString(row.CreditLimit)
		credLimit = &s
	}

	var intRate *string
	if row.InterestRate.Valid {
		s := common.NumericToString(row.InterestRate)
		intRate = &s
	}

	var archivedAt *time.Time
	if row.ArchivedAt.Valid {
		t := row.ArchivedAt.Time
		archivedAt = &t
	}

	return AccountResponse{
		ID:                     common.UUIDToString(row.ID),
		UserID:                 common.UUIDToString(row.UserID),
		InstitutionID:          instID,
		InstitutionName:        row.InstitutionName,
		InstitutionDisplayName: row.InstitutionDisplayName,
		InstitutionLogoURL:     row.InstitutionLogoUrl,
		Name:                   row.Name,
		AccountType:            row.AccountType,
		AccountSubtype:         row.AccountSubtype,
		Currency:               row.Currency,
		CurrentBalance:         common.NumericToString(row.CurrentBalance),
		AvailableBalance:       availBal,
		CreditLimit:            credLimit,
		InterestRate:           intRate,
		SourceType:             row.SourceType,
		ArchivedAt:             archivedAt,
		CreatedAt:              row.CreatedAt.Time,
		UpdatedAt:              row.UpdatedAt.Time,
	}
}

func accountDetailResponse(row sqlc.GetAccountByIDRow) AccountResponse {
	var instID *string
	if row.InstitutionID.Valid {
		s := common.UUIDToString(row.InstitutionID)
		instID = &s
	}

	var availBal *string
	if row.AvailableBalance.Valid {
		s := common.NumericToString(row.AvailableBalance)
		availBal = &s
	}

	var credLimit *string
	if row.CreditLimit.Valid {
		s := common.NumericToString(row.CreditLimit)
		credLimit = &s
	}

	var intRate *string
	if row.InterestRate.Valid {
		s := common.NumericToString(row.InterestRate)
		intRate = &s
	}

	var archivedAt *time.Time
	if row.ArchivedAt.Valid {
		t := row.ArchivedAt.Time
		archivedAt = &t
	}

	return AccountResponse{
		ID:                     common.UUIDToString(row.ID),
		UserID:                 common.UUIDToString(row.UserID),
		InstitutionID:          instID,
		InstitutionName:        row.InstitutionName,
		InstitutionDisplayName: row.InstitutionDisplayName,
		InstitutionLogoURL:     row.InstitutionLogoUrl,
		Name:                   row.Name,
		AccountType:            row.AccountType,
		AccountSubtype:         row.AccountSubtype,
		Currency:               row.Currency,
		CurrentBalance:         common.NumericToString(row.CurrentBalance),
		AvailableBalance:       availBal,
		CreditLimit:            credLimit,
		InterestRate:           intRate,
		SourceType:             row.SourceType,
		ArchivedAt:             archivedAt,
		CreatedAt:              row.CreatedAt.Time,
		UpdatedAt:              row.UpdatedAt.Time,
	}
}
