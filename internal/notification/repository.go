package notification

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/database/db"
)

type Repository interface {
	CreateNotification(ctx context.Context, arg db.CreateNotificationParams) (db.Notification, error)
	ListNotificationsByUser(ctx context.Context, arg db.ListNotificationsByUserParams) ([]db.Notification, error)
	CountNotificationsByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkNotificationAsRead(ctx context.Context, arg db.MarkNotificationAsReadParams) (db.Notification, error)
	MarkAllNotificationsAsRead(ctx context.Context, userID uuid.UUID) error
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) CreateNotification(ctx context.Context, arg db.CreateNotificationParams) (db.Notification, error) {
	return r.queries.CreateNotification(ctx, arg)
}

func (r *repository) ListNotificationsByUser(ctx context.Context, arg db.ListNotificationsByUserParams) ([]db.Notification, error) {
	return r.queries.ListNotificationsByUser(ctx, arg)
}

func (r *repository) CountNotificationsByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.CountNotificationsByUser(ctx, userID)
}

func (r *repository) MarkNotificationAsRead(ctx context.Context, arg db.MarkNotificationAsReadParams) (db.Notification, error) {
	return r.queries.MarkNotificationAsRead(ctx, arg)
}

func (r *repository) MarkAllNotificationsAsRead(ctx context.Context, userID uuid.UUID) error {
	return r.queries.MarkAllNotificationsAsRead(ctx, userID)
}