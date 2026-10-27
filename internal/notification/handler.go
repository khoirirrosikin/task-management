package notification

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
	notifications := rg.Group("/notifications")
	notifications.Use(authMiddleware)
	{
		notifications.GET("", h.ListNotifications)
		notifications.PATCH("/read-all", h.MarkAllAsRead)
		notifications.PATCH("/:id/read", h.MarkAsRead)
	}
}

func (h *Handler) ListNotifications(c *gin.Context) {
	var query ListNotificationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation error", response.FormatValidationError(err))
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, meta, err := h.service.ListNotifications(c.Request.Context(), userID, query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Notifications retrieved successfully", res, meta)
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid notification ID", nil)
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, err := h.service.MarkAsRead(c.Request.Context(), notificationID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	
	response.Success(c, http.StatusOK, "Notification marked as read", res)
}

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "All notifications marked as read", nil)
}