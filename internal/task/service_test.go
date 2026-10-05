package task

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

type MockTaskRepository struct {
	mock.Mock
}

func(m *MockTaskRepository) CreateTask(ctx context.Context, arg db.CreateTaskParams) (db.Task, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Task), args.Error(1)
}

func(m *MockTaskRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (db.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Task), args.Error(1)
}

func (m *MockTaskRepository) ListTaskByProject(ctx context.Context, projectID uuid.UUID) ([]db.Task, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]db.Task), args.Error(1)
}

func (m *MockTaskRepository) UpdateTask(ctx context.Context, arg db.UpdateTaskParams) (db.Task, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Task), args.Error(1)
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (db.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Project), args.Error(1)
}

func TestCreateTask_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	userID := uuid.New()
	projectID := uuid.New()
	assigneeID := uuid.New()
	assigneeStr := assigneeID.String()
	dueDate := time.Now().Add(24 * time.Hour)
	now := time.Now()

	req := CreateTaskRequest{
		Title: "Task title",
		Description: "Task description",
		Status: "in_progress",
		Priority: "high",
		AssignedTo: &assigneeStr,
		DueDate: &dueDate,
	}

	mockProject := db.Project{
		ID: projectID,
		OwnerID: userID,
	}

	mockTask := db.Task{
		ID: uuid.New(),
		ProjectID: projectID,
		Title: req.Title,
		Description: pgtype.Text{String: req.Description, Valid: true},
		Status: req.Status,
		Priority: req.Priority,
		AssignedTo: pgtype.UUID{Bytes: assigneeID, Valid: true},
		DueDate: pgtype.Timestamptz{Time: dueDate, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockTaskRepo.On("CreateTask", mock.Anything, mock.MatchedBy(func(arg db.CreateTaskParams) bool {
		return arg.Title == req.Title && 
			arg.Status == req.Status &&
			arg.Priority == req.Priority &&
			arg.AssignedTo.Valid && uuid.UUID(arg.AssignedTo.Bytes) == assigneeID
	})).Return(mockTask, nil)

	res, err := service.CreateTask(context.Background(), projectID, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, mockTask.ID.String(), res.ID)
	assert.Equal(t, req.Title, res.Title)
	assert.Equal(t, req.Status, res.Status)
	assert.Equal(t, req.Priority, res.Priority)
	assert.Equal(t, &assigneeStr, res.AssignedTo)
	mockProjectRepo.AssertExpectations(t)
	mockTaskRepo.AssertExpectations(t)
}

func TestCreateTask_Success_DefaultValues(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	userID := uuid.New()
	projectID := uuid.New()

	req := CreateTaskRequest{
		Title: "Task title",
	}

	mockProject := db.Project{
		ID: projectID,
		OwnerID: userID,
	}

	mockTask := db.Task{
		ID: uuid.New(),
		ProjectID: projectID,
		Title: req.Title,
		Status: "todo",
		Priority: "medium",
	}

	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockTaskRepo.On("CreateTask", mock.Anything, mock.MatchedBy(func(arg db.CreateTaskParams) bool {
		return arg.Title == req.Title &&
			arg.Status == "todo" &&
			arg.Priority == "medium" &&
			!arg.AssignedTo.Valid &&
			!arg.DueDate.Valid
	})).Return(mockTask, nil)

	res, err := service.CreateTask(context.Background(), projectID, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "todo", res.Status)
	assert.Equal(t, "medium", res.Priority)
}

func TestCreateTask_ProjectNotFound(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	projectID := uuid.New()
	userID := uuid.New()

	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(db.Project{}, errors.New("sql: no rowa in result set"))

	res, err := service.CreateTask(context.Background(), projectID, userID, CreateTaskRequest{Title: "Title"})

	assert.Error(t, err)
	assert.Nil(t, res)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, 404, appErr.Code)
}

func TestCreateTask_UnauthorizedProject(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	projectID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(db.Project{
			ID: projectID,
			OwnerID: otherUserID,
		}, nil)

	res, err := service.CreateTask(context.Background(), projectID, userID, CreateTaskRequest{Title: "Title"})

	assert.Error(t, err)
	assert.Nil(t, res)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, 403, appErr.Code)
}

func TestCreateTask_InvalidAssigneeUUID(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	projectID := uuid.New()
	userID := uuid.New()
	invalidUUID := "not-a-uuid"

	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(db.Project{
			ID: projectID,
			OwnerID: userID,
		}, nil)

	res, err := service.CreateTask(context.Background(), projectID, userID, 
		CreateTaskRequest{
			Title: "Title",
			AssignedTo: &invalidUUID,
		},
	)

	assert.Error(t, err)
	assert.Nil(t, res)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, 400, appErr.Code)
}

func TestGetTaskByID_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	taskID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()

	mockTask := db.Task{
		ID: taskID,
		ProjectID: projectID,
		Title: "task",
	}

	mockProject := db.Project{
		ID: projectID,
		OwnerID: userID,
	}

	mockTaskRepo.On("GetTaskByID", mock.Anything, taskID).
		Return(mockTask, nil)
	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(mockProject, nil)

	res, err := service.GetTaskByID(context.Background(), taskID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, taskID.String(), res.ID)
}

func TestGetTaskByID_NotFound(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	taskID := uuid.New()
	userID := uuid.New()

	mockTaskRepo.On("GetTaskByID", mock.Anything, taskID).
		Return(db.Task{}, errors.New("Not Found"))

	res, err := service.GetTaskByID(context.Background(), taskID, userID)

	assert.Error(t, err)
	assert.Nil(t, res)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, 404, appErr.Code)
}

func TestListTaskByProject_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	projectID := uuid.New()
	userID := uuid.New()

	mockProject := db.Project{
		ID: projectID,
		OwnerID: userID,
	}

	mockTasks := []db.Task{
		{
			ID: uuid.New(), ProjectID: projectID, Title: "Task 1",
		},
		{
			ID: uuid.New(), ProjectID: projectID, Title: "Task 2",
		},
	}

	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(mockProject, nil)

	mockTaskRepo.On("ListTaskByProject", mock.Anything, projectID).
		Return(mockTasks, nil)
	

	res, err := service.ListTaskByProject(context.Background(), projectID, userID)

	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestUpdateTask_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	taskID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()

	existingTask := db.Task{
		ID: taskID,
		ProjectID: projectID,
		Title: "Old title",
		Status: "todo",
		Priority: "medium",
	}

	mockProject := db.Project{ID: projectID, OwnerID: userID}

	newTitle := "Updated title"
	newStatus := "completed"
	req := UpdateTaskRequest{
		Title: newTitle,
		Status: newStatus,
	}

	updatedTask := db.Task{
		ID: taskID,
		ProjectID: projectID,
		Title: newTitle,
		Status: newStatus,
		Priority: "medium",
	}

	mockTaskRepo.On("GetTaskByID", mock.Anything, taskID).Return(existingTask, nil)
	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockTaskRepo.On("UpdateTask", mock.Anything, mock.MatchedBy(func(arg db.UpdateTaskParams) bool{
		return arg.ID == taskID &&
			arg.Title == newTitle &&
			arg.Status == newStatus
	})).Return(updatedTask, nil)

	res, err := service.UpdateTask(context.Background(), taskID, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, newTitle, res.Title)
	assert.Equal(t, newStatus, res.Status)
	mockProjectRepo.AssertExpectations(t)
	mockTaskRepo.AssertExpectations(t)
}

func TestDeleteTask_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockProjectRepo := new(MockProjectRepository)
	service := NewService(mockTaskRepo, mockProjectRepo)

	taskID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()

	mockTask := db.Task{ID: taskID, ProjectID: projectID}
	mockProject := db.Project{ID: projectID, OwnerID: userID}

	mockTaskRepo.On("GetTaskByID", mock.Anything, taskID).Return(mockTask, nil)
	mockProjectRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockTaskRepo.On("DeleteTask", mock.Anything, taskID).Return(nil)

	err := service.DeleteTask(context.Background(), taskID, userID)

	assert.NoError(t, err)
	mockTaskRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
}