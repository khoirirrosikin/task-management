package notification

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/response"
)

type Service interface {
	Enqueue(event Event)
	StartWorker(ctx context.Context)
	ListNotifications(ctx context.Context, userID uuid.UUID, query ListNotificationsQuery) ([]NotificationResponse, *response.PaginationMeta, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) (*NotificationResponse, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	repo Repository
	eventChan chan Event
}

func NewService(repo Repository, bufferSize int) Service {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &service{
		repo: repo,
		eventChan: make(chan Event, bufferSize),
	}
}

func (s *service) Enqueue(event Event) {
	select {
		case s.eventChan <- event:
		default:
			log.Printf("Notification queue is full, dropping notification for user %s", event.UserID)
	}
}

func (s *service) StartWorker(ctx context.Context) {
	log.Println("Notification background worker started")
	for {
		select {
			case <-ctx.Done():
				log.Println("Notification background worker stopping...")
				return
			case event, ok := <-s.eventChan:
				if !ok {
					return
				}
				s.processEvent(event)
		}
	}
}

func (s *service) processEvent(event Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.repo.CreateNotification(ctx, db.CreateNotificationParams{
		UserID: event.UserID,
		Title: event.Title,
		Message: event.Message,
	})
	if err != nil {
		log.Printf("Failed to save notification for user %s: %v", event.UserID, err)
	}
}

func (s *service) ListNotifications(ctx context.Context, userID uuid.UUID, query ListNotificationsQuery) ([]NotificationResponse, *response.PaginationMeta, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	totalItems, err := s.repo.CountNotificationsByUser(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to count notifications: %w", err)
	}

	notifications, err := s.repo.ListNotificationsByUser(ctx, db.ListNotificationsByUserParams{
		UserID: userID,
		Limit: int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	res := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		res[i] = *toNotificationResponse(n)
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = (int(totalItems) + limit - 1) / limit
	}

	meta := &response.PaginationMeta{
		CurrentPage: page,
		PerPage: limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return res, meta, nil
}

func (s *service) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) (*NotificationResponse, error) {
	n, err := s.repo.MarkNotificationAsRead(ctx, db.MarkNotificationAsReadParams{
		ID: notificationID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.ErrNotFound("Notification not found")
		}
		return nil, fmt.Errorf("Failed to mark notification as read: %w", err)
	}
	return toNotificationResponse(n), nil
}

func (s *service) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.MarkAllNotificationsAsRead(ctx, userID); err != nil {
		return fmt.Errorf("Failed to mark all notifications as read %w", err)
	}
	return nil
}

func toNotificationResponse(n db.Notification) *NotificationResponse {
	return &NotificationResponse{
		ID: n.ID.String(),
		UserID: n.UserID.String(),
		Title: n.Title,
		Message: n.Message,
		IsRead: n.IsRead,
		CreatedAt: n.CreatedAt.Time,
	}
}