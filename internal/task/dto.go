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
	AssignedTo 	*string    `json:"assigned_to"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}