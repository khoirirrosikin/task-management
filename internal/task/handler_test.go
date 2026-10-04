package task

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateTask(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, req CreateTaskRequest) (*TaskResponse, error) {
	args := m.Called(ctx, projectID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TaskResponse), args.Error(1)
}

func (m *MockService) GetTaskByID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*TaskResponse, error) {
	args := m.Called(ctx, taskID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TaskResponse), args.Error(1)
}

func (m *MockService) ListTaskByProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]TaskResponse, error) {
	args := m.Called(ctx, projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TaskResponse), args.Error(1)
}

func (m *MockService) UpdateTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, req UpdateTaskRequest) (*TaskResponse, error) {
	args := m.Called(ctx, taskID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TaskResponse), args.Error(1)
}

func (m *MockService) DeleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, taskID, userID)
	return args.Error(0)
}

func setupTestRouter(service Service, testUserID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := NewHandler(service)

	mockAuth := func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	}

	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1, mockAuth)

	return router
}

func performRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		switch b := body.(type) {
		case []byte:
			bodyReader = bytes.NewBuffer(b)
		default:
			jsonBytes, _ := json.Marshal(body)
			bodyReader = bytes.NewBuffer(jsonBytes)
		}
	}

	req, _ := http.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestHandler_CreateTask_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := CreateTaskRequest{
		Title:       "Title",
		Description: "Description",
		Priority:    "high",
		Status:      "todo",
	}

	expectedResponse := &TaskResponse{
		ID:          uuid.New().String(),
		ProjectID:   projectID.String(),
		Title:       reqBody.Title,
		Description: reqBody.Description,
		Priority:    reqBody.Priority,
		Status:      reqBody.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("CreateTask", mock.Anything, projectID, testUserID, reqBody).
		Return(expectedResponse, nil)

	url := "/api/v1/projects/" + projectID.String() + "/tasks"
	w := performRequest(router, http.MethodPost, url, reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Task created successfully", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_CreateTask_InvalidProjectID(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := "invalid-uuid"
	router := setupTestRouter(mockService, testUserID)

	reqBody := CreateTaskRequest{
		Title: "Title",
	}

	url := "/api/v1/projects/" + projectID + "/tasks"
	w := performRequest(router, http.MethodPost, url, reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid project ID", res.Message)
	mockService.AssertNotCalled(t, "CreateTask")
}

func TestHandler_CreateTask_ValidationError(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := CreateTaskRequest{
		Title: "",
	}

	url := "/api/v1/projects/" + projectID.String() + "/tasks"
	w := performRequest(router, http.MethodPost, url, reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation error", res.Message)
	mockService.AssertNotCalled(t, "CreateTask")
}

func TestHandler_ListTasks_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	expectedResponse := []TaskResponse{
		{
			ID:          uuid.New().String(),
			ProjectID:   projectID.String(),
			Title:       "Task 1",
			Description: "Description 1",
			Priority:    "high",
			Status:      "todo",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			ProjectID:   projectID.String(),
			Title:       "Task 2",
			Description: "Description 2",
			Priority:    "medium",
			Status:      "in_progress",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockService.On("ListTaskByProject", mock.Anything, projectID, testUserID).
		Return(expectedResponse, nil)

	url := "/api/v1/projects/" + projectID.String() + "/tasks"
	w := performRequest(router, http.MethodGet, url, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Tasks retrieved successfully", res.Message)
	assert.Len(t, res.Data, len(expectedResponse))
	mockService.AssertExpectations(t)
}

func TestHandler_ListTasks_InvalidProjectID(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := "invalid-uuid"
	router := setupTestRouter(mockService, testUserID)

	mockService.On("ListTaskByProject", mock.Anything, projectID, testUserID).
		Return(nil, errors.New("invalid project ID"))

	url := "/api/v1/projects/" + projectID + "/tasks"
	w := performRequest(router, http.MethodGet, url, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid project ID", res.Message)
	mockService.AssertNotCalled(t, "ListTaskByProject")
}

func TestHandler_GetTaskByID_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	expectedResponse := &TaskResponse{
		ID:          taskID.String(),
		ProjectID:   uuid.New().String(),
		Title:       "Task",
		Description: "Description",
		Priority:    "high",
		Status:      "todo",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("GetTaskByID", mock.Anything, taskID, testUserID).
		Return(expectedResponse, nil)

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodGet, url, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Task retrieved successfully", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_GetTaskByID_InvalidTaskID(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := "invalid-uuid"
	router := setupTestRouter(mockService, testUserID)

	url := "/api/v1/tasks/" + taskID
	w := performRequest(router, http.MethodGet, url, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid task ID", res.Message)
	mockService.AssertNotCalled(t, "GetTaskByID")
}

func TestHandler_GetTaskByID_NotFound(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("GetTaskByID", mock.Anything, taskID, testUserID).
		Return(nil, response.ErrNotFound("Task not found"))

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodGet, url, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Task not found", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateTask_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := UpdateTaskRequest{
		Title:       "Title updated",
		Description: "Description updated",
		Priority:    "high",
		Status:      "completed",
	}

	expectedResponse := &TaskResponse{
		ID:          taskID.String(),
		Title:       reqBody.Title,
		Description: reqBody.Description,
		Priority:    reqBody.Priority,
		Status:      reqBody.Status,
		UpdatedAt:   time.Now(),
	}

	mockService.On("UpdateTask", mock.Anything, taskID, testUserID, reqBody).
		Return(expectedResponse, nil)

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodPut, url, reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Task updated successfully", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateTask_InvalidTaskID(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := "invalid-uuid"
	router := setupTestRouter(mockService, testUserID)

	reqBody := UpdateTaskRequest{
		Title:       "Title updated",
		Description: "Description updated",
		Priority:    "high",
		Status:      "completed",
	}

	url := "/api/v1/tasks/" + taskID
	w := performRequest(router, http.MethodPut, url, reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid task ID", res.Message)
	mockService.AssertNotCalled(t, "UpdateTask")
}

func TestHandler_UpdateTask_ValidationError(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := UpdateTaskRequest{
		Status:      "invalid_status",
	}

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodPut, url, reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation error", res.Message)
	mockService.AssertNotCalled(t, "UpdateTask")
}

func TestHandler_UpdateTask_NotFound(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := UpdateTaskRequest{
		Title:      "Title updated",
	}

	mockService.On("UpdateTask", mock.Anything, taskID, testUserID, reqBody).
		Return(nil, response.ErrNotFound("Task not found"))

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodPut, url, reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Task not found", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateTask_Unauthorized(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := UpdateTaskRequest{
		Title:      "Title updated",
	}

	mockService.On("UpdateTask", mock.Anything, taskID, testUserID, reqBody).
		Return(nil, response.ErrUnauthorized("Unauthorized to access this task"))

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodPut, url, reqBody)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Unauthorized to access this task", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_deleteTask_Success(t *testing.T){
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("DeleteTask", mock.Anything, taskID, testUserID).Return(nil)

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodDelete, url, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Task deleted successfully", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_deleteTask_InvalidTaskID(t *testing.T){
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := "invalid-uuid"
	router := setupTestRouter(mockService, testUserID)

	url := "/api/v1/tasks/" + taskID
	w := performRequest(router, http.MethodDelete, url, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Invalid task ID", res.Message)
	mockService.AssertNotCalled(t, "DeleteTask")
}

func TestHandler_deleteTask_Unauthorized(t *testing.T){
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("DeleteTask", mock.Anything, taskID, testUserID).
		Return(response.ErrUnauthorized("Unauthorized to delete this task"))

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodDelete, url, nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Unauthorized to delete this task", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_deleteTask_NotFound(t *testing.T){
	mockService := new(MockService)
	testUserID := uuid.New()
	taskID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("DeleteTask", mock.Anything, taskID, testUserID).
		Return(response.ErrNotFound("Task not found"))

	url := "/api/v1/tasks/" + taskID.String()
	w := performRequest(router, http.MethodDelete, url, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Task not found", res.Message)
	mockService.AssertExpectations(t)
}
