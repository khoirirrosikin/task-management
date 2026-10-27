package task

import "time"

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,min=2,max=200"`
	Description string     `json:"description"`
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress completed"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	AssignedTo  *string    `json:"assigned_to" binding:"omitempty,uuid"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       string     `json:"title" binding:"omitempty,min=2,max=200"`
	Description string     `json:"description"`
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress completed"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	AssignedTo  *string    `json:"assigned_to" binding:"omitempty,uuid"`
	DueDate     *time.Time `json:"due_date"`
}

type TaskResponse struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssignedTo  *string    `json:"assigned_to"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type CommentResponse struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	UserEmail string    `json:"user_email,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskFilterQuery struct {
	Search string `form:"search" binding:"omitempty,max=100"`
	Status string `form:"status" binding:"omitempty,oneof=todo in_progress completed"`
	Priority string `form:"priority" binding:"omitempty,oneof=low medium high urgent"`
	AssignedTo *string `form:"assigned_to" binding:"omitempty,uuid"`
	DueBefore *time.Time `form:"due_before" time_format:"2006-01-02T15:04:05Z07:00"`
	Page int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}
