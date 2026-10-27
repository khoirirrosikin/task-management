package notification

import (
	"context"
	"encoding/json"
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

func (m *MockService) Enqueue(event Event) {
	m.Called(event)
}

func (m *MockService) StartWorker(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockService) ListNotifications(ctx context.Context, userID uuid.UUID, query ListNotificationsQuery) ([]NotificationResponse, *response.PaginationMeta, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var meta *response.PaginationMeta
	if args.Get(1) != nil {
		meta = args.Get(1).(*response.PaginationMeta)
	}
	return args.Get(0).([]NotificationResponse), meta, args.Error(2)
}

func (m *MockService) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) (*NotificationResponse, error) {
	args := m.Called(ctx, notificationID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*NotificationResponse), args.Error(1)
}

func (m *MockService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
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

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestHandler_ListNotifications_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	expectedRes := []NotificationResponse{
		{
			ID:        uuid.New().String(),
			UserID:    testUserID.String(),
			Title:     "New Task Assigned",
			Message:   "You have been assigned to a task",
			IsRead:    false,
			CreatedAt: time.Now(),
		},
	}

	expectedMeta := &response.PaginationMeta{
		CurrentPage: 1,
		PerPage:     10,
		TotalItems:  1,
		TotalPages:  1,
	}

	mockService.On("ListNotifications", mock.Anything, testUserID, ListNotificationsQuery{}).
		Return(expectedRes, expectedMeta, nil)

	w := performRequest(router, http.MethodGet, "/api/v1/notifications")

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.NotNil(t, res.Meta)
	mockService.AssertExpectations(t)
}

func TestHandler_MarkAsRead_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	notifID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	expectedRes := &NotificationResponse{
		ID:        notifID.String(),
		UserID:    testUserID.String(),
		Title:     "New Task Assigned",
		Message:   "You have been assigned to a task",
		IsRead:    true,
		CreatedAt: time.Now(),
	}

	mockService.On("MarkAsRead", mock.Anything, notifID, testUserID).
		Return(expectedRes, nil)

	url := "/api/v1/notifications/" + notifID.String() + "/read"
	w := performRequest(router, http.MethodPatch, url)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Notification marked as read", res.Message)
	mockService.AssertExpectations(t)
}

func TestHandler_MarkAsRead_InvalidID(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	w := performRequest(router, http.MethodPatch, "/api/v1/notifications/invalid-uuid/read")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "MarkAsRead")
}

func TestHandler_MarkAllAsRead_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	mockService.On("MarkAllAsRead", mock.Anything, testUserID).Return(nil)

	w := performRequest(router, http.MethodPatch, "/api/v1/notifications/read-all")

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "All notifications marked as read", res.Message)
	mockService.AssertExpectations(t)
}