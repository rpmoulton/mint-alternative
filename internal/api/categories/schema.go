package categories

import (
	"time"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type CategoryResponse struct {
	ID               string    `json:"id"`
	UserID           *string   `json:"user_id,omitempty"`
	ParentCategoryID *string   `json:"parent_category_id,omitempty"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	Icon             *string   `json:"icon,omitempty"`
	Color            *string   `json:"color,omitempty"`
	IsSystem         bool      `json:"is_system"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	ParentCategoryID *string `json:"parent_category_id" validate:"omitempty,uuid"`
	Name             string  `json:"name"               validate:"required,min=1,max=100"`
	Type             string  `json:"type"               validate:"required,oneof=income expense transfer"`
	Icon             *string `json:"icon"               validate:"omitempty,max=50"`
	Color            *string `json:"color"              validate:"omitempty,max=50"`
}

type UpdateCategoryRequest struct {
	ParentCategoryID *string `json:"parent_category_id" validate:"omitempty,uuid"`
	Name             *string `json:"name"               validate:"omitempty,min=1,max=100"`
	Type             *string `json:"type"               validate:"omitempty,oneof=income expense transfer"`
	Icon             *string `json:"icon"               validate:"omitempty,max=50"`
	Color            *string `json:"color"              validate:"omitempty,max=50"`
}

type CreateCategoryInput struct {
	Body CreateCategoryRequest `body:"required"`
}

type UpdateCategoryInput struct {
	ID   string                `path:"id"      validate:"required,uuid"`
	Body UpdateCategoryRequest `body:"required"`
}

type GetCategoryInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

type DeleteCategoryInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

type ListCategoriesInput struct{}

func categoryResponse(cat sqlc.Category) CategoryResponse {
	var uid *string
	isSystem := true
	if cat.UserID.Valid {
		s := common.UUIDToString(cat.UserID)
		uid = &s
		isSystem = false
	}

	var parentID *string
	if cat.ParentCategoryID.Valid {
		s := common.UUIDToString(cat.ParentCategoryID)
		parentID = &s
	}

	return CategoryResponse{
		ID:               common.UUIDToString(cat.ID),
		UserID:           uid,
		ParentCategoryID: parentID,
		Name:             cat.Name,
		Type:             cat.Type,
		Icon:             cat.Icon,
		Color:            cat.Color,
		IsSystem:         isSystem,
		CreatedAt:        cat.CreatedAt.Time,
		UpdatedAt:        cat.UpdatedAt.Time,
	}
}
