package activity

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/response"
)

type Service interface {
	LogActivity(ctx context.Context, arg LogParams)
	ListProjectActivities(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, query ListActivityLogsQuery) ([]ActivityLogResponse, *response.PaginationMeta, error)
}

type ProjectRepository interface {
	GetProjectByID(ctx context.Context, id uuid.UUID) (db.Project, error)
	GetProjectMember(ctx context.Context, arg db.GetProjectMemberParams) (db.ProjectMember, error)
}

type service struct {
	repo        Repository
	projectRepo ProjectRepository
}

func NewService(repo Repository, projectRepo ProjectRepository) Service {
	return &service{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

func (s *service) LogActivity(ctx context.Context, arg LogParams) {
	_, err := s.repo.CreateActivityLog(ctx, db.CreateActivityLogParams{
		ProjectID:  arg.ProjectID,
		UserID:     arg.UserID,
		Action:     arg.Action,
		EntityType: arg.EntityType,
		EntityID:   arg.EntityID,
		Details:    arg.Details,
	})
	if err != nil {
		log.Printf("Failed to record activity log for project %s: %v", arg.ProjectID, err)
	}
}

func (s *service) ListProjectActivities(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, query ListActivityLogsQuery) ([]ActivityLogResponse, *response.PaginationMeta, error) {
	if err := s.verifyProjectAccess(ctx, projectID, userID); err != nil {
		return nil, nil, err
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	totalItems, err := s.repo.CountActivityLogsByProject(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to count activity logs: %w", err)
	}

	logs, err := s.repo.ListActivityLogsByProject(ctx, db.ListActivityLogsByProjectParams{
		ProjectID: projectID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to list activity logs: %w", err)
	}

	res := make([]ActivityLogResponse, len(logs))
	for i, l := range logs {
		res[i] = ActivityLogResponse{
			ID:         l.ID.String(),
			ProjectID:  l.ProjectID.String(),
			UserID:     l.UserID.String(),
			UserName:   l.UserName,
			UserEmail:  l.UserEmail,
			Action:     l.Action,
			EntityType: l.EntityType,
			EntityID:   l.EntityID.String(),
			Details:    l.Details,
			CreatedAt:  l.CreatedAt.Time,
		}
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = (int(totalItems) + limit - 1) / limit
	}

	meta := &response.PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
	}

	return res, meta, nil
}

func (s *service) verifyProjectAccess(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error {
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return response.ErrNotFound("Project not found")
	}

	if project.OwnerID == userID {
		return nil
	}

	_, err = s.projectRepo.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return response.ErrForbidden("Unauthorized to access this project")
	}

	return nil
}