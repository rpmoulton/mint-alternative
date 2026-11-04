package institutions

import (
	"context"
	"log/slog"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type Service interface {
	List(ctx context.Context) ([]InstitutionResponse, error)
	Get(ctx context.Context, id string) (*InstitutionResponse, error)
	Create(ctx context.Context, req *CreateInstitutionRequest) (*InstitutionResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) ([]InstitutionResponse, error) {
	insts, err := s.repo.ListInstitutions(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list institutions", "error", err)
		return nil, httpx.NewInternalError("could not list institutions")
	}

	result := make([]InstitutionResponse, len(insts))
	for i, inst := range insts {
		result[i] = institutionResponse(inst)
	}
	return result, nil
}

func (s *service) Get(ctx context.Context, id string) (*InstitutionResponse, error) {
	uid, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid institution ID")
	}

	inst, err := s.repo.GetInstitutionByID(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get institution", "id", id, "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("institution not found")
		}
		return nil, httpx.NewInternalError("could not retrieve institution")
	}

	resp := institutionResponse(*inst)
	return &resp, nil
}

func (s *service) Create(ctx context.Context, req *CreateInstitutionRequest) (*InstitutionResponse, error) {
	inst, err := s.repo.CreateInstitution(ctx, sqlc.CreateInstitutionParams{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		LogoUrl:     req.LogoURL,
		Website:     req.Website,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create institution", "error", err)
		return nil, httpx.NewInternalError("could not create institution")
	}

	resp := institutionResponse(*inst)
	return &resp, nil
}
