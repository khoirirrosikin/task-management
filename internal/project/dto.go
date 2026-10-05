package project

import "time"

type CreateProjectRequest struct {
	Name string `json:"name" binding:"required,min=2,max=150"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name string `json:"name" binding:"required,min=2,max=150"`
	Description string `json:"description"`
}

type ProjectResponse struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	OwnerID string `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role string `json:"role" binding:"omitempty,oneof=owner member"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=owner member"`
}

type MemberResponse struct {
	ID string `json:"id"`
	ProjectID string `json:"project_id"`
	UserID string `json:"user_id"`
	UserName string `json:"user_name"`
	UserEmail string `json:"user_email"`
	Role string `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}