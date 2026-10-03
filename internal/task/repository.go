package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/database/db"
)

type Repository interface {
	CreateTask(ctx context.Context, arg db.CreateTaskParams) (db.Task, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (db.Task, error)
	ListTaskByProject(ctx context.Context, projectID uuid.UUID) ([]db.Task, error)
	UpdateTask(ctx context.Context, arg db.UpdateTaskParams) (db.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	queries *db.Queries
}

func NewRepository(query *db.Queries) Repository {
	return &repository{
		queries: query,
	}
}

func (r *repository) CreateTask(ctx context.Context, arg db.CreateTaskParams) (db.Task, error) {
	return r.queries.CreateTask(ctx, arg)
}

func (r *repository) GetTaskByID(ctx context.Context, id uuid.UUID) (db.Task, error) {
	return r.queries.GetTaskByID(ctx, id)
}

func (r *repository) ListTaskByProject(ctx context.Context, projectID uuid.UUID) ([]db.Task, error) {
	return r.queries.ListTaskByProject(ctx, projectID)
}

func (r *repository) UpdateTask(ctx context.Context, arg db.UpdateTaskParams) (db.Task, error) {
	return r.queries.UpdateTask(ctx, arg)
}

func (r *repository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteTask(ctx, id)
}