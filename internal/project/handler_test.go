package project

import (
	"bytes"
	"context"
	"encoding/json"
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

func (m *MockService) CreateProject(ctx context.Context, ownerID uuid.UUID, req CreateProjectRequest) (*ProjectResponse, error) {
	args := m.Called(ctx, ownerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProjectResponse), args.Error(1)
}
func (m *MockService) GetProjectByID(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (*ProjectResponse, error) {
	args := m.Called(ctx, projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProjectResponse), args.Error(1)
}

func (m *MockService) ListProjects(ctx context.Context, ownerID uuid.UUID) ([]ProjectResponse, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ProjectResponse), args.Error(1)
}

func (m *MockService) UpdateProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error) {
	args := m.Called(ctx, projectID, ownerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProjectResponse), args.Error(1)
}

func (m *MockService) DeleteProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID) error {
	args := m.Called(ctx, projectID, ownerID)
	return args.Error(0)
}

func (m *MockService) AddMember(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, req AddMemberRequest) (*MemberResponse, error) {
	args := m.Called(ctx, projectID, ownerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MemberResponse), args.Error(1)
}

func (m *MockService) ListMembers(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]MemberResponse, error) {
	args := m.Called(ctx, projectID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]MemberResponse), args.Error(1)
}

func (m *MockService) RemoveMember(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, memberID uuid.UUID) error {
	args := m.Called(ctx, projectID, ownerID, memberID)
	return args.Error(0)
}

func (m *MockService) UpdateMemberRole(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, memberID uuid.UUID, req UpdateMemberRoleRequest) (*MemberResponse, error) {
	args := m.Called(ctx, projectID, ownerID, memberID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MemberResponse), args.Error(1)
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

func TestHandler_CreateProject_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := CreateProjectRequest{
		Name:        "Test project",
		Description: "Test project description",
	}

	expectedRes := &ProjectResponse{
		ID:          uuid.New().String(),
		Name:        reqBody.Name,
		Description: reqBody.Description,
		OwnerID:     testUserID.String(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("CreateProject", mock.Anything, testUserID, reqBody).
		Return(expectedRes, nil)

	w := performRequest(router, http.MethodPost, "/api/v1/projects", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Project created successfully", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_CreateProject_ValidationError(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := CreateProjectRequest{
		Name: "",
	}

	w := performRequest(router, http.MethodPost, "/api/v1/projects", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Validation error", res.Message)
	mockService.AssertNotCalled(t, "CreateProject")
}

func TestHandler_ListProjects_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	expectedRes := []ProjectResponse{
		{
			ID:          uuid.New().String(),
			Name:        "Test project 1",
			Description: "Test description 1",
		},
		{
			ID:          uuid.New().String(),
			Name:        "Test project 2",
			Description: "Test description 2",
		},
	}

	mockService.On("ListProjects", mock.Anything, testUserID).
		Return(expectedRes, nil)

	w := performRequest(router, http.MethodGet, "/api/v1/projects", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Len(t, res.Data, len(expectedRes))

	mockService.AssertExpectations(t)
}

func TestHandler_GetProjectByID_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	expectedRes := &ProjectResponse{
		ID:          projectID.String(),
		Name:        "Test project",
		Description: "Test project description",
		OwnerID:     testUserID.String(),
	}

	mockService.On("GetProjectByID", mock.Anything, projectID, testUserID).
		Return(expectedRes, nil)

	w := performRequest(router, http.MethodGet, "/api/v1/projects/"+projectID.String(), nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Project retrieved successfully", res.Message)
	assert.NotNil(t, res.Data)
	mockService.AssertExpectations(t)
}

func TestHandler_GetProjectByID_NotFound(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("GetProjectByID", mock.Anything, projectID, testUserID).
		Return(nil, response.ErrNotFound("Project not found"))

	w := performRequest(router, http.MethodGet, "/api/v1/projects/"+projectID.String(), nil)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Project not found", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_GetProjectByID_Unauthorized(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("GetProjectByID", mock.Anything, projectID, testUserID).
		Return(nil, response.ErrUnauthorized("Unauthorized to access this project"))

	w := performRequest(router, http.MethodGet, "/api/v1/projects/"+projectID.String(), nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Unauthorized to access this project", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateProject_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	reqBody := UpdateProjectRequest{
		Name:        "Test project updated",
		Description: "Test project description updated",
	}

	expectedRes := &ProjectResponse{
		ID:          projectID.String(),
		Name:        reqBody.Name,
		Description: reqBody.Description,
		OwnerID:     testUserID.String(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("UpdateProject", mock.Anything, projectID, testUserID, reqBody).
		Return(expectedRes, nil)

	w := performRequest(router, http.MethodPut, "/api/v1/projects/"+projectID.String(), reqBody)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Project updated successfully", res.Message)
	assert.NotNil(t, res.Data)
	mockService.AssertExpectations(t)
}

func TestHandler_DeleteProject_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("DeleteProject", mock.Anything, projectID, testUserID).
		Return(nil)

	w := performRequest(router, http.MethodDelete, "/api/v1/projects/"+projectID.String(), nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Project deleted successfully", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_DeleteProject_Unauthorized(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("DeleteProject", mock.Anything, projectID, testUserID).
		Return(response.ErrUnauthorized("Unauthorized to access this project"))

	w := performRequest(router, http.MethodDelete, "/api/v1/projects/"+projectID.String(), nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Unauthorized to access this project", res.Message)
	mockService.AssertExpectations(t)
}
