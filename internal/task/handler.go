package task

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
	projectTasks := rg.Group("/projects/:id/tasks")
	projectTasks.Use(authMiddleware)
	{
		projectTasks.POST("", h.CreateTask)
		projectTasks.GET("", h.ListTasks)
	}

	tasks := rg.Group("/tasks")
	tasks.Use(authMiddleware)
	{
		tasks.GET("/:id", h.GetTaskByID)
		tasks.PUT("/:id", h.UpdateTask)
		tasks.DELETE("/:id", h.DeleteTask)
	}
}

func (h *Handler) CreateTask(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", nil)
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation error", response.FormatValidationError(err))
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	
	res, err := h.service.CreateTask(c.Request.Context(), projectID, userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Task created successfully", res)
}

func (h *Handler) ListTasks(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", nil)
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, err := h.service.ListTaskByProject(c.Request.Context(), projectID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Tasks retrieved successfully", res)
}

func (h *Handler) GetTaskByID(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task ID", nil)
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, err := h.service.GetTaskByID(c.Request.Context(), taskID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Task retrieved successfully", res)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task ID", nil)
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation error", response.FormatValidationError(err))
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, err := h.service.UpdateTask(c.Request.Context(), taskID, userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Task updated successfully", res)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid task ID", nil)
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.service.DeleteTask(c.Request.Context(), taskID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Task deleted successfully", nil)
}