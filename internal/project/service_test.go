package project

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Project), args.Error(1)
}

func (m *MockRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (db.Project, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Project), args.Error(1)
}

func (m *MockRepository) ListProjectsByOwner(ctx context.Context, ownerID uuid.UUID) ([]db.Project, error) {
	args := m.Called(ctx, ownerID)
	return args.Get(0).([]db.Project), args.Error(1)
}

func (m *MockRepository) UpdateProject(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Project), args.Error(1)
}

func (m *MockRepository) DeleteProject(ctx context.Context, arg db.DeleteProjectParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRepository) AddProjectMember(ctx context.Context, arg db.AddProjectMemberParams) (db.ProjectMember, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ProjectMember), args.Error(1)
}

func (m *MockRepository) GetProjectMember(ctx context.Context, arg db.GetProjectMemberParams) (db.ProjectMember, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ProjectMember), args.Error(1)
}

func (m *MockRepository) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]db.ListProjectMembersRow, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]db.ListProjectMembersRow), args.Error(1)
}

func (m *MockRepository) RemoveProjectMember(ctx context.Context, arg db.RemoveProjectMemberParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockRepository) UpdateProjectMemberRole(ctx context.Context, arg db.UpdateProjectMemberRoleParams) (db.ProjectMember, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ProjectMember), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func TestCreateProject_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	ownerID := uuid.New()
	projectID := uuid.New()
	now := time.Now()

	req := CreateProjectRequest{
		Name:        "Test project",
		Description: "Test description",
	}

	mockProject := db.Project{
		ID:          projectID,
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: true},
		OwnerID:     ownerID,
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}

	mockRepo.On("CreateProject", mock.Anything, mock.MatchedBy(func(arg db.CreateProjectParams) bool {
		return arg.Name == req.Name && arg.OwnerID == ownerID
	})).Return(mockProject, nil)

	res, err := service.CreateProject(context.Background(), ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, projectID.String(), res.ID)
	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, req.Description, res.Description)
	assert.Equal(t, ownerID.String(), res.OwnerID)
	mockRepo.AssertExpectations(t)
}

func TestCreateProject_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	ownerID := uuid.New()
	req := CreateProjectRequest{
		Name: "Test project",
	}

	mockRepo.On("CreateProject", mock.Anything, mock.Anything).
		Return(db.Project{}, errors.New("Database error"))

	res, err := service.CreateProject(context.Background(), ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Failed to create project")
	mockRepo.AssertExpectations(t)
}

func TestGetProjectByID_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	now := time.Now()

	mockProject := db.Project{
		ID:          projectID,
		Name:        "Test project",
		Description: pgtype.Text{String: "Test decription", Valid: true},
		OwnerID:     ownerID,
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(mockProject, nil)

	res, err := service.GetProjectByID(context.Background(), projectID, ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, projectID.String(), res.ID)
	assert.Equal(t, "Test project", res.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetProjectByID_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()

	mockRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(db.Project{}, errors.New("Not found"))

	res, err := service.GetProjectByID(context.Background(), projectID, ownerID)

	assert.Error(t, err)
	assert.Nil(t, res)

	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Project not found", appErr.Message)
	mockRepo.AssertExpectations(t)
}

func TestGetProjectByID_Unauthorized(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	mockProject := db.Project{
		ID:      projectID,
		Name:    "Test project owner",
		OwnerID: ownerID,
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(mockProject, nil)

	res, err := service.GetProjectByID(context.Background(), projectID, otherUserID)

	assert.Error(t, err)
	assert.Nil(t, res)

	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusForbidden, appErr.Code)
	assert.Equal(t, "Unauthorized to access this project", appErr.Message)
	mockRepo.AssertExpectations(t)
}

func TestListProjects_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	ownerID := uuid.New()
	now := time.Now()

	mockProjects := []db.Project{
		{
			ID:        uuid.New(),
			Name:      "Test project 1",
			OwnerID:   ownerID,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			ID:        uuid.New(),
			Name:      "Test project 2",
			OwnerID:   ownerID,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	mockRepo.On("ListProjectsByOwner", mock.Anything, ownerID).
		Return(mockProjects, nil)

	res, err := service.ListProjects(context.Background(), ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Test project 1", res[0].Name)
	assert.Equal(t, "Test project 2", res[1].Name)
	assert.Len(t, res, 2)
	mockRepo.AssertExpectations(t)
}

func TestListProjects_Empty(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	ownerID := uuid.New()

	mockRepo.On("ListProjectsByOwner", mock.Anything, ownerID).
		Return([]db.Project{}, nil)

	res, err := service.ListProjects(context.Background(), ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 0)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProject_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	now := time.Now()

	req := UpdateProjectRequest{
		Name:        "Test project update",
		Description: "Test description update",
	}

	existing := db.Project{
		ID:      projectID,
		Name:    "Test project",
		OwnerID: ownerID,
	}

	updated := db.Project{
		ID:          projectID,
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: true},
		OwnerID:     ownerID,
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).
		Return(existing, nil)
	mockRepo.On("UpdateProject", mock.Anything, mock.MatchedBy(func(arg db.UpdateProjectParams) bool {
		return arg.ID == projectID && arg.Name == req.Name && arg.OwnerID == ownerID
	})).Return(updated, nil)

	res, err := service.UpdateProject(context.Background(), projectID, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, req.Description, res.Description)
	mockRepo.AssertExpectations(t)
}

func TestDeleteProject_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()

	existing := db.Project{
		ID:      projectID,
		Name:    "To Delete",
		OwnerID: ownerID,
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(existing, nil)
	mockRepo.On("DeleteProject", mock.Anything, db.DeleteProjectParams{
		ID:      projectID,
		OwnerID: ownerID,
	}).Return(nil)

	err := service.DeleteProject(context.Background(), projectID, ownerID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAddMember_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	memberUserID := uuid.New()
	now := time.Now()

	req := AddMemberRequest{
		Email: "collab@example.com",
		Role:  "member",
	}

	mockProject := db.Project{
		ID:      projectID,
		Name:    "Collab project",
		OwnerID: ownerID,
	}

	mockUser := db.User{
		ID:    memberUserID,
		Name:  "Collaborator",
		Email: req.Email,
	}

	mockMember := db.ProjectMember{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    memberUserID,
		Role:      req.Role,
		JoinedAt:  pgtype.Timestamptz{Time: now, Valid: true},
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(mockUser, nil)
	mockRepo.On("GetProjectMember", mock.Anything, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    memberUserID,
	}).Return(db.ProjectMember{}, pgx.ErrNoRows)
	mockRepo.On("AddProjectMember", mock.Anything, db.AddProjectMemberParams{
		ProjectID: projectID,
		UserID: memberUserID,
		Role: req.Role,
	}).Return(mockMember, nil)

	res, err := service.AddMember(context.Background(), projectID, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, memberUserID.String(), res.UserID)
	assert.Equal(t, req.Role, res.Role)
	mockRepo.AssertExpectations(t)
}

func TestAddMember_CannotAddSelf(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	projectID := uuid.New()
	ownerID := uuid.New()

	req := AddMemberRequest{
		Email: "owner@example.com",
		Role:  "member",
	}

	mockProject := db.Project{
		ID:      projectID,
		Name:    "Owner Project",
		OwnerID: ownerID,
	}
	mockUser := db.User{
		ID:    ownerID,
		Name:  "Owner",
		Email: req.Email,
	}
	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(mockUser, nil)

	res, err := service.AddMember(context.Background(), projectID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, res)

	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestAddMember_AlreadyMember(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	memberUserID := uuid.New()

	req := AddMemberRequest{
		Email: "collab@example.com",
		Role:  "member",
	}
	mockProject := db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}
	mockUser := db.User{
		ID:    memberUserID,
		Name:  "Collab",
		Email: req.Email,
	}
	mockExistingMember := db.ProjectMember{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    memberUserID,
		Role:      "member",
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(mockUser, nil)
	mockRepo.On("GetProjectMember", mock.Anything, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    memberUserID,
	}).Return(mockExistingMember, nil)

	res, err := service.AddMember(context.Background(), projectID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestListMembers_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	now := time.Now()

	mockProject := db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}

	mockMembers := []db.ListProjectMembersRow{
		{
			ID:        uuid.New(),
			ProjectID: projectID,
			UserID:    ownerID,
			UserName:  "Owner User",
			UserEmail: "owner@example.com",
			Role:      "owner",
			JoinedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			ID:        uuid.New(),
			ProjectID: projectID,
			UserID:    uuid.New(),
			UserName:  "Member User",
			UserEmail: "member@example.com",
			Role:      "member",
			JoinedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockRepo.On("ListProjectMembers", mock.Anything, projectID).Return(mockMembers, nil)

	res, err := service.ListMembers(context.Background(), projectID, ownerID)

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "Owner User", res[0].UserName)
	mockRepo.AssertExpectations(t)
}

func TestListMembers_Forbidden(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	nonMemberID := uuid.New()

	mockProject := db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockRepo.On("GetProjectMember", mock.Anything, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    nonMemberID,
	}).Return(db.ProjectMember{}, pgx.ErrNoRows)

	res, err := service.ListMembers(context.Background(), projectID, nonMemberID)

	assert.Error(t, err)
	assert.Nil(t, res)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusForbidden, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUpdateMemberRole_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	memberUserID := uuid.New()
	now := time.Now()

	req := UpdateMemberRoleRequest{
		Role: "owner",
	}

	mockProject := db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}

	mockUpdatedMember := db.ProjectMember{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    memberUserID,
		Role:      req.Role,
		JoinedAt:  pgtype.Timestamptz{Time: now, Valid: true},
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockRepo.On("UpdateProjectMemberRole", mock.Anything, db.UpdateProjectMemberRoleParams{
		ProjectID: projectID,
		UserID:    memberUserID,
		Role:      req.Role,
	}).Return(mockUpdatedMember, nil)

	res, err := service.UpdateMemberRole(context.Background(), projectID, ownerID, memberUserID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Role, res.Role)
	mockRepo.AssertExpectations(t)
}

func TestUpdateMemberRole_CannotUpdateOwner(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()

	req := UpdateMemberRoleRequest{
		Role: "member",
	}

	mockProject := db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)

	res, err := service.UpdateMemberRole(context.Background(), projectID, ownerID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestRemoveMember_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()
	memberUserID := uuid.New()

	mockProject := db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}

	mockMember := db.ProjectMember{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    memberUserID,
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)
	mockRepo.On("GetProjectMember", mock.Anything, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    memberUserID,
	}).Return(mockMember, nil)
	mockRepo.On("RemoveProjectMember", mock.Anything, db.RemoveProjectMemberParams{
		ProjectID: projectID,
		UserID:    memberUserID,
	}).Return(nil)

	err := service.RemoveMember(context.Background(), projectID, ownerID, memberUserID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRemoveMember_CannotRemoveOwner(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	projectID := uuid.New()
	ownerID := uuid.New()

	mockProject := db.Project{
		ID:      projectID,
		OwnerID: ownerID,
	}

	mockRepo.On("GetProjectByID", mock.Anything, projectID).Return(mockProject, nil)

	err := service.RemoveMember(context.Background(), projectID, ownerID, ownerID)

	assert.Error(t, err)
	var appErr *response.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	mockRepo.AssertExpectations(t)
}


