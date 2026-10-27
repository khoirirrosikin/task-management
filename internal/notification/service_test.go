package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateNotification(ctx context.Context, arg db.CreateNotificationParams) (db.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Notification), args.Error(1)
}

func (m *MockRepository) ListNotificationsByUser(ctx context.Context, arg db.ListNotificationsByUserParams) ([]db.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Notification), args.Error(1)
}

func (m *MockRepository) CountNotificationsByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) MarkNotificationAsRead(ctx context.Context, arg db.MarkNotificationAsReadParams) (db.Notification, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Notification), args.Error(1)
}

func (m *MockRepository) MarkAllNotificationsAsRead(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestWorker_EnqueueAndProcess(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, 10)

	userID := uuid.New()
	event := Event{
		UserID:  userID,
		Title:   "New Task Assigned",
		Message: "You have been assigned to task 'Setup DB'",
	}

	processed := make(chan struct{})
	mockRepo.On("CreateNotification", mock.Anything, db.CreateNotificationParams{
		UserID:  event.UserID,
		Title:   event.Title,
		Message: event.Message,
	}).Run(func(args mock.Arguments) {
		close(processed)
	}).Return(db.Notification{
		ID:      uuid.New(),
		UserID:  event.UserID,
		Title:   event.Title,
		Message: event.Message,
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go service.StartWorker(ctx)

	service.Enqueue(event)

	select {
	case <-processed:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for worker to process notification event")
	}

	mockRepo.AssertExpectations(t)
}

func TestListNotifications_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, 10)

	userID := uuid.New()
	now := time.Now()

	mockNotifications := []db.Notification{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "Task Assigned",
			Message:   "Message 1",
			IsRead:    false,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	mockRepo.On("CountNotificationsByUser", mock.Anything, userID).Return(int64(1), nil)
	mockRepo.On("ListNotificationsByUser", mock.Anything, db.ListNotificationsByUserParams{
		UserID: userID,
		Limit:  10,
		Offset: 0,
	}).Return(mockNotifications, nil)

	res, meta, err := service.ListNotifications(context.Background(), userID, ListNotificationsQuery{})

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "Task Assigned", res[0].Title)
	assert.NotNil(t, meta)
	assert.Equal(t, int64(1), meta.TotalItems)
	assert.Equal(t, 1, meta.TotalPages)
	mockRepo.AssertExpectations(t)
}

func TestMarkAsRead_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, 10)

	userID := uuid.New()
	notifID := uuid.New()
	now := time.Now()

	mockRepo.On("MarkNotificationAsRead", mock.Anything, db.MarkNotificationAsReadParams{
		ID:     notifID,
		UserID: userID,
	}).Return(db.Notification{
		ID:        notifID,
		UserID:    userID,
		Title:     "Task Assigned",
		Message:   "Message",
		IsRead:    true,
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}, nil)

	res, err := service.MarkAsRead(context.Background(), notifID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.IsRead)
	mockRepo.AssertExpectations(t)
}

func TestMarkAsRead_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, 10)

	userID := uuid.New()
	notifID := uuid.New()

	mockRepo.On("MarkNotificationAsRead", mock.Anything, db.MarkNotificationAsReadParams{
		ID:     notifID,
		UserID: userID,
	}).Return(db.Notification{}, pgx.ErrNoRows)

	res, err := service.MarkAsRead(context.Background(), notifID, userID)

	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}

func TestMarkAllAsRead_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, 10)

	userID := uuid.New()

	mockRepo.On("MarkAllNotificationsAsRead", mock.Anything, userID).Return(nil)

	err := service.MarkAllAsRead(context.Background(), userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}