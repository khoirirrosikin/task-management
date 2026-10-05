package project

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/database/db"
)

type Repository interface {
	CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (db.Project, error)
	ListProjectsByOwner(ctx context.Context, ownerID uuid.UUID) ([]db.Project, error)
	UpdateProject(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error)
	DeleteProject(ctx context.Context, arg db.DeleteProjectParams) error

	AddProjectMember(ctx context.Context, arg db.AddProjectMemberParams) (db.ProjectMember, error)
	GetProjectMember(ctx context.Context, arg db.GetProjectMemberParams) (db.ProjectMember, error)
	ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]db.ListProjectMembersRow, error)
	RemoveProjectMember(ctx context.Context, arg db.RemoveProjectMemberParams) error
	UpdateProjectMemberRole(ctx context.Context, arg db.UpdateProjectMemberRoleParams) (db.ProjectMember, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)

}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{
		queries: queries,
	}
}

func (r *repository) CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error) {
	return r.queries.CreateProject(ctx, arg)
}

func (r *repository) GetProjectByID(ctx context.Context, id uuid.UUID) (db.Project, error) {
	return r.queries.GetProjectByID(ctx, id)
}

func (r *repository) ListProjectsByOwner(ctx context.Context, ownerID uuid.UUID) ([]db.Project, error) {
	return r.queries.ListProjectsByOwner(ctx, ownerID)
}

func (r *repository) UpdateProject(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error) {
	return r.queries.UpdateProject(ctx, arg)
}

func (r *repository) DeleteProject(ctx context.Context, arg db.DeleteProjectParams) error {
	return r.queries.DeleteProject(ctx, arg)
}

func (r *repository) AddProjectMember(ctx context.Context, arg db.AddProjectMemberParams) (db.ProjectMember, error) {
	return r.queries.AddProjectMember(ctx, arg)
}

func (r *repository) GetProjectMember(ctx context.Context, arg db.GetProjectMemberParams) (db.ProjectMember, error) {
	return r.queries.GetProjectMember(ctx, arg)
}

func (r *repository) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]db.ListProjectMembersRow, error) {
	return r.queries.ListProjectMembers(ctx, projectID)
}

func (r *repository) RemoveProjectMember(ctx context.Context, arg db.RemoveProjectMemberParams) error {
	return r.queries.RemoveProjectMember(ctx, arg)
}

func (r *repository) UpdateProjectMemberRole(ctx context.Context, arg db.UpdateProjectMemberRoleParams) (db.ProjectMember, error) {
	return r.queries.UpdateProjectMemberRole(ctx, arg)
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}