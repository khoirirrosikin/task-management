package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/response"
)

type Service interface {
	CreateProject(ctx context.Context, ownerID uuid.UUID, req CreateProjectRequest) (*ProjectResponse, error)
	GetProjectByID(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (*ProjectResponse, error)
	ListProjects(ctx context.Context, ownerID uuid.UUID) ([]ProjectResponse, error)
	UpdateProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID) error

	AddMember(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, req AddMemberRequest) (*MemberResponse, error)
	ListMembers(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]MemberResponse, error)
	RemoveMember(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, memberID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, memberID uuid.UUID, req UpdateMemberRoleRequest) (*MemberResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateProject(ctx context.Context, ownerID uuid.UUID, req CreateProjectRequest) (*ProjectResponse, error) {
	arg := db.CreateProjectParams{
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid: req.Description != "",
		},
		OwnerID: ownerID,
	}

	project, err := s.repo.CreateProject(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to create project: %w", err)
	}

	return toProjectResponse(project), nil
}

func (s *service) GetProjectByID(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (*ProjectResponse, error) {
	project, err := s.getProjectAndVerifyOwner(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	return toProjectResponse(project), nil
}

func (s *service) ListProjects(ctx context.Context, ownerID uuid.UUID) ([]ProjectResponse, error) {
	projects, err := s.repo.ListProjectsByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("Failed to list projects: %w", err)
	}

	res := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		res = append(res, *toProjectResponse(p))
	}

	return res, nil
}

func (s *service) UpdateProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, req UpdateProjectRequest) (*ProjectResponse, error) {
	if _, err := s.getProjectAndVerifyOwner(ctx, projectID, ownerID); err != nil {
		return nil, err
	}

	arg := db.UpdateProjectParams{
		ID: projectID,
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid: req.Description != "",
		},
		OwnerID: ownerID,
	}

	updated, err := s.repo.UpdateProject(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to update project")
	}

	return toProjectResponse(updated), nil
}

func (s *service) DeleteProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID) error {
	if _, err := s.getProjectAndVerifyOwner(ctx, projectID, ownerID); err != nil {
		return err
	}

	arg := db.DeleteProjectParams{
		ID: projectID,
		OwnerID: ownerID,
	}

	if err := s.repo.DeleteProject(ctx, arg); err != nil {
		return fmt.Errorf("Failed to delete project: %w", err)
	}

	return nil
}

func(s *service) AddMember(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, req AddMemberRequest) (*MemberResponse, error) {
	if _, err := s.getProjectAndVerifyOwner(ctx, projectID, ownerID); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.ErrNotFound("User with this email not found")
		}
		return nil, fmt.Errorf("Failed to find user by email: %w", err)
	}

	if user.ID == ownerID {
		return nil, response.ErrBadRequest("Cannot add yourself as a member")
	}

	_, err = s.repo.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID: user.ID,
	})
	if err == nil {
		return nil, response.ErrBadRequest("User is already a member of this project")
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	pm, err := s.repo.AddProjectMember(ctx,
		db.AddProjectMemberParams{
			ProjectID: projectID,
			UserID: user.ID,
			Role: role,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to add project member: %w", err)
	}

	return toMemberResponse(pm, user.Name, user.Email), nil
}

func (s *service) ListMembers(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]MemberResponse, error) {
	if _, err := s.getProjectAndVerifyAccess(ctx, projectID, userID); err != nil {
		return nil, err
	}

	members, err := s.repo.ListProjectMembers(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("Failed to list project members: %w", err)
	}

	res := make([]MemberResponse, len(members))
	for i, m := range members {
		res[i] = MemberResponse{
			ID: m.ID.String(),
			ProjectID: m.ProjectID.String(),
			UserID: m.UserID.String(),
			UserName: m.UserName,
			UserEmail: m.UserEmail,
			Role: m.Role,
			JoinedAt: m.JoinedAt.Time,
		}
	}

	return res, nil
}

func (s *service) RemoveMember(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, memberID uuid.UUID) error {
	p, err := s.getProjectAndVerifyOwner(ctx, projectID, ownerID)
	if err != nil {
		return err
	}

	if memberID == p.OwnerID {
		return response.ErrBadRequest("Cannot remove project owner from project members")
	}

	_, err = s.repo.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID: memberID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.ErrNotFound("Member not found in this project")
		}
		return fmt.Errorf("Failed to find project member: %w", err)
	}

	if err := s.repo.RemoveProjectMember(ctx, db.RemoveProjectMemberParams{
		ProjectID: projectID,
		UserID: memberID,
	}); err != nil {
		return fmt.Errorf("Failed to remove project member: %w", err)
	}

	return nil
}

func (s *service) UpdateMemberRole(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID, memberID uuid.UUID, req UpdateMemberRoleRequest) (*MemberResponse, error) {
	p, err := s.getProjectAndVerifyOwner(ctx, projectID, ownerID)
	if err != nil {
		return nil, err
	}

	if memberID == p.OwnerID {
		return nil, response.ErrBadRequest("Cannot change role of project owner")
	}

	pm, err := s.repo.UpdateProjectMemberRole(ctx, db.UpdateProjectMemberRoleParams{
		ProjectID: projectID,
		UserID: memberID,
		Role: req.Role,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.ErrNotFound("Member not found in this project")
		}
		return nil, fmt.Errorf("Failed to update project member role: %w", err)
	}

	return toMemberResponse(pm, "", ""), nil

}

func (s *service) getProjectAndVerifyOwner(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID) (db.Project, error) {
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return db.Project{}, response.ErrNotFound("Project not found")
	}

	if project.OwnerID != ownerID {
		return db.Project{}, response.ErrForbidden("Unauthorized to access this project")
	}

	return project, nil
}

func (s *service) getProjectAndVerifyAccess(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (db.Project, error) {
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return db.Project{}, response.ErrNotFound("Project not found")
	}

	if project.OwnerID == userID {
		return project, nil
	}

	_, err = s.repo.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID: userID,
	})
	if err != nil {
		return db.Project{}, response.ErrForbidden("Unauthorized to access this project")
	}

	return project, nil
}

func toProjectResponse(p db.Project) *ProjectResponse {
	var desc string
	if p.Description.Valid {
		desc = p.Description.String
	}

	return &ProjectResponse{
		ID: p.ID.String(),
		Name: p.Name,
		Description: desc,
		OwnerID: p.OwnerID.String(),
		CreatedAt: p.CreatedAt.Time,
		UpdatedAt: p.UpdatedAt.Time,
	}
}

func toMemberResponse(pm db.ProjectMember, userName, userEmail string) *MemberResponse {
	return &MemberResponse{
		ID: pm.ID.String(),
		ProjectID: pm.ProjectID.String(),
		UserID: pm.UserID.String(),
		UserName: userName,
		UserEmail: userEmail,
		Role: pm.Role,
		JoinedAt: pm.JoinedAt.Time,
	}
}