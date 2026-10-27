package activity

import (
	"time"

	"github.com/google/uuid"
)

type LogParams struct {
	ProjectID  uuid.UUID
	UserID     uuid.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	Details    string
}

type ListActivityLogsQuery struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

type ActivityLogResponse struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserEmail  string    `json:"user_email"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Details    string    `json:"details"`
	CreatedAt  time.Time `json:"created_at"`
}