package activity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateActivityLog(ctx context.Context, arg db.CreateActivityLogParams) (db.ActivityLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ActivityLog), args.Error(1)
}

func (m *MockRepository) ListActivityLogsByProject(ctx context.Context, arg db.ListActivityLogsByProjectParams) ([]db.ListActivityLogsByProjectRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.ListActivityLogsByProjectRow), args.Error(1)
}

func (m *MockRepository) CountActivityLogsByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).(int64), args.Error(1)
}

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (db.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Project), args.Error(1)
}

func (m *MockProjectRepository) GetProjectMember(ctx context.Context, arg db.GetProjectMemberParams) (db.ProjectMember, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ProjectMember), args.Error(1)
}

func TestLogActivity_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockRepo, mockProjectRepo)

	projectID := uuid.New()
	userID := uuid.New()
	entityID := uuid.New()

	params := LogParams{
		ProjectID:  projectID,
		UserID:     userID,
		Action:     "task_created",
		EntityType: "task",
		EntityID:   entityID,
		Details:    "Created task 'Setup DB'",
	}

	mockRepo.On("CreateActivityLog", mock.Anything, db.CreateActivityLogParams{
		ProjectID:  params.ProjectID,
		UserID:     params.UserID,
		Action:     params.Action,
		EntityType: params.EntityType,
		EntityID:   params.EntityID,
		Details:    params.Details,
	}).Return(db.ActivityLog{ID: uuid.New()}, nil)

	service.LogActivity(context.Background(), params)

	mockRepo.AssertExpectations(t)
}

func TestListProjectActivities_Success_Owner(t *testing.T) {
	mockRepo := new(MockRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockRepo, mockProjectRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	entityID := uuid.New()
	now := time.Now()

	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).Return(db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}, nil)

	mockRepo.On("CountActivityLogsByProject", mock.Anything, projectID).Return(int64(1), nil)
	mockRepo.On("ListActivityLogsByProject", mock.Anything, db.ListActivityLogsByProjectParams{
		ProjectID: projectID,
		Limit:     10,
		Offset:    0,
	}).Return([]db.ListActivityLogsByProjectRow{
		{
			ID:         uuid.New(),
			ProjectID:  projectID,
			UserID:     ownerID,
			UserName:   "Khoirir",
			UserEmail:  "khoirir@example.com",
			Action:     "task_created",
			EntityType: "task",
			EntityID:   entityID,
			Details:    "Created task 'Setup DB'",
			CreatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		},
	}, nil)

	res, meta, err := service.ListProjectActivities(context.Background(), projectID, ownerID, ListActivityLogsQuery{})

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "task_created", res[0].Action)
	assert.NotNil(t, meta)
	assert.Equal(t, int64(1), meta.TotalItems)
	mockProjectRepo.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestListProjectActivities_Forbidden(t *testing.T) {
	mockRepo := new(MockRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockRepo, mockProjectRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	strangerID := uuid.New()

	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).Return(db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}, nil)
	mockProjectRepo.On("GetProjectMember", mock.Anything, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    strangerID,
	}).Return(db.ProjectMember{}, errors.New("not found"))

	res, meta, err := service.ListProjectActivities(context.Background(), projectID, strangerID, ListActivityLogsQuery{})

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Nil(t, meta)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, 403, appErr.Code)
	mockRepo.AssertNotCalled(t, "ListActivityLogsByProject")
}