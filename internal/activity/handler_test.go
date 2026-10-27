package activity

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

func (m *MockService) LogActivity(ctx context.Context, arg LogParams) {
	m.Called(ctx, arg)
}

func (m *MockService) ListProjectActivities(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, query ListActivityLogsQuery) ([]ActivityLogResponse, *response.PaginationMeta, error) {
	args := m.Called(ctx, projectID, userID, query)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var meta *response.PaginationMeta
	if args.Get(1) != nil {
		meta = args.Get(1).(*response.PaginationMeta)
	}
	return args.Get(0).([]ActivityLogResponse), meta, args.Error(2)
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

func TestHandler_ListProjectActivities_Success(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	projectID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	expectedRes := []ActivityLogResponse{
		{
			ID:         uuid.New().String(),
			ProjectID:  projectID.String(),
			UserID:     testUserID.String(),
			UserName:   "Khoirir",
			UserEmail:  "khoirir@example.com",
			Action:     "task_created",
			EntityType: "task",
			EntityID:   uuid.New().String(),
			Details:    "Created task 'Setup DB'",
			CreatedAt:  time.Now(),
		},
	}

	expectedMeta := &response.PaginationMeta{
		CurrentPage: 1,
		PerPage:     10,
		TotalItems:  1,
		TotalPages:  1,
	}

	mockService.On("ListProjectActivities", mock.Anything, projectID, testUserID, ListActivityLogsQuery{}).
		Return(expectedRes, expectedMeta, nil)

	url := "/api/v1/projects/" + projectID.String() + "/activities"
	w := performRequest(router, http.MethodGet, url)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Activity logs retrieved successfully", res.Message)
	assert.NotNil(t, res.Meta)
	mockService.AssertExpectations(t)
}

func TestHandler_ListProjectActivities_InvalidProjectID(t *testing.T) {
	mockService := new(MockService)
	testUserID := uuid.New()
	router := setupTestRouter(mockService, testUserID)

	w := performRequest(router, http.MethodGet, "/api/v1/projects/invalid-uuid/activities")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ListProjectActivities")
}