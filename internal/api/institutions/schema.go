package institutions

import (
	"time"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type InstitutionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	LogoURL     *string   `json:"logo_url,omitempty"`
	Website     *string   `json:"website,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateInstitutionRequest struct {
	Name        string  `json:"name"         validate:"required,min=1,max=100"`
	DisplayName string  `json:"display_name" validate:"required,min=1,max=100"`
	LogoURL     *string `json:"logo_url"     validate:"omitempty,url"`
	Website     *string `json:"website"      validate:"omitempty,url"`
}

type CreateInstitutionInput struct {
	Body CreateInstitutionRequest `body:"required"`
}

type GetInstitutionInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

type ListInstitutionsInput struct{}

func institutionResponse(inst sqlc.Institution) InstitutionResponse {
	return InstitutionResponse{
		ID:          common.UUIDToString(inst.ID),
		Name:        inst.Name,
		DisplayName: inst.DisplayName,
		LogoURL:     inst.LogoUrl,
		Website:     inst.Website,
		CreatedAt:   inst.CreatedAt.Time,
		UpdatedAt:   inst.UpdatedAt.Time,
	}
}
