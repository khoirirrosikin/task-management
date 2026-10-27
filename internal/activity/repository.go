package activity

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/database/db"
)

type Repository interface {
	CreateActivityLog(ctx context.Context, arg db.CreateActivityLogParams) (db.ActivityLog, error)
	ListActivityLogsByProject(ctx context.Context, arg db.ListActivityLogsByProjectParams) ([]db.ListActivityLogsByProjectRow, error)
	CountActivityLogsByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) CreateActivityLog(ctx context.Context, arg db.CreateActivityLogParams) (db.ActivityLog, error) {
	return r.queries.CreateActivityLog(ctx, arg)
}

func (r *repository) ListActivityLogsByProject(ctx context.Context, arg db.ListActivityLogsByProjectParams) ([]db.ListActivityLogsByProjectRow, error) {
	return r.queries.ListActivityLogsByProject(ctx, arg)
}

func (r *repository) CountActivityLogsByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	return r.queries.CountActivityLogsByProject(ctx, projectID)
}