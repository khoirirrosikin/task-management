package notification

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	UserID uuid.UUID
	Title string
	Message string
}

type ListNotificationsQuery struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

type NotificationResponse struct {
	ID string `json:"id"`
	UserID string `json:"user_id"`
	Title string `json:"title"`
	Message string `json:"message"`
	IsRead bool `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}