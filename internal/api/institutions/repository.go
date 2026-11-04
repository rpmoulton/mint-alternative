package institutions

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type Repository interface {
	ListInstitutions(ctx context.Context) ([]sqlc.Institution, error)
	GetInstitutionByID(ctx context.Context, id pgtype.UUID) (*sqlc.Institution, error)
	CreateInstitution(ctx context.Context, params sqlc.CreateInstitutionParams) (*sqlc.Institution, error)
}

type repository struct {
	query *sqlc.Queries
}

func NewRepository(pg db.DB) Repository {
	return &repository{
		query: pg.RootStore().Queries(),
	}
}

func (r *repository) ListInstitutions(ctx context.Context) ([]sqlc.Institution, error) {
	insts, err := r.query.ListInstitutions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list institutions: %w", err)
	}
	return insts, nil
}

func (r *repository) GetInstitutionByID(ctx context.Context, id pgtype.UUID) (*sqlc.Institution, error) {
	inst, err := r.query.GetInstitutionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get institution %v: %w", id, err)
	}
	return &inst, nil
}

func (r *repository) CreateInstitution(ctx context.Context, params sqlc.CreateInstitutionParams) (*sqlc.Institution, error) {
	inst, err := r.query.CreateInstitution(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create institution: %w", err)
	}
	return &inst, nil
}
