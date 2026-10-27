package activity

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	projectActivities := rg.Group("/projects/:id/activities")
	projectActivities.Use(authMiddleware)
	{
		projectActivities.GET("", h.ListProjectActivities)
	}
}

func (h *Handler) ListProjectActivities(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", nil)
		return
	}

	var query ListActivityLogsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation error", response.FormatValidationError(err))
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, meta, err := h.service.ListProjectActivities(c.Request.Context(), projectID, userID, query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Activity logs retrieved successfully", res, meta)
}