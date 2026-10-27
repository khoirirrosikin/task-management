package task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/response"
)

type Service interface {
	CreateTask(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, req CreateTaskRequest) (*TaskResponse, error)
	GetTaskByID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*TaskResponse, error)
	ListTaskByProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, filter TaskFilterQuery) ([]TaskResponse, *response.PaginationMeta, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, req UpdateTaskRequest) (*TaskResponse, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error

	AddComment(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, req CreateCommentRequest) (*CommentResponse, error)
	ListComments(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) ([]CommentResponse, error)
	DeleteComment(ctx context.Context, taskID uuid.UUID, commentID uuid.UUID, userID uuid.UUID) error
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

func (s *service) CreateTask(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, req CreateTaskRequest) (*TaskResponse, error) {
	project, err := s.verifyProjectAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = "todo"
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	assignedTo, err := parseAssignedTo(req.AssignedTo)
	if err != nil {
		return nil, err
	}

	if err := s.validateAssignee(ctx, project, assignedTo); err != nil {
		return nil, err
	}

	arg := db.CreateTaskParams{
		ProjectID: projectID,
		Title:     req.Title,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  req.Description != "",
		},
		Status:     status,
		Priority:   priority,
		AssignedTo: assignedTo,
		DueDate:    parseDueDate(req.DueDate),
	}

	task, err := s.repo.CreateTask(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to create task: %w", err)
	}

	return toTaskResponse(task), nil
}

func (s *service) GetTaskByID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*TaskResponse, error) {
	task, _, err := s.getTaskAndVerifyAccess(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}

	return toTaskResponse(task), nil
}

func (s *service) ListTaskByProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, filter TaskFilterQuery) ([]TaskResponse, *response.PaginationMeta, error) {
	if _, err := s.verifyProjectAccess(ctx, projectID, userID); err != nil {
		return nil, nil, err
	}

	assignedTo, err := parseAssignedTo(filter.AssignedTo)
	if err != nil {
		return nil, nil, err
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	status := pgtype.Text{
		String: filter.Status,
		Valid: filter.Status != "",
	}
	priority := pgtype.Text{
		String: filter.Priority,
		Valid: filter.Priority != "",
	}
	search := pgtype.Text{
		String: filter.Search,
		Valid: filter.Search != "",
	}
	dueBefore := parseDueDate(filter.DueBefore)

	totalItems, err := s.repo.CountTasksByProject(ctx, db.CountTasksByProjectParams{
		ProjectID: projectID,
		Status: status,
		Priority: priority,
		AssignedTo: assignedTo,
		DueBefore: dueBefore,
		Search: search,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to count tasks: %w", err)
	}

	tasks, err := s.repo.ListTaskByProject(ctx, db.ListTaskByProjectParams{
		ProjectID: projectID,
		Limit: int32(limit),
		Offset: int32(offset),
		Status: status,
		Priority: priority,
		AssignedTo: assignedTo,
		DueBefore: dueBefore,
		Search: search,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to list tasks: %w", err)
	}

	responses := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = *toTaskResponse(t)
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = (int(totalItems) + limit -1) / limit
	}

	meta := &response.PaginationMeta{
		CurrentPage: page,
		PerPage: limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}

	return responses, meta, nil
}

func (s *service) UpdateTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, req UpdateTaskRequest) (*TaskResponse, error) {
	existingTask, project, err := s.getTaskAndVerifyAccess(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}

	arg := db.UpdateTaskParams{
		ID:          taskID,
		Title:       existingTask.Title,
		Description: existingTask.Description,
		Status:      existingTask.Status,
		Priority:    existingTask.Priority,
		AssignedTo:  existingTask.AssignedTo,
		DueDate:     existingTask.DueDate,
	}

	if req.Title != "" {
		arg.Title = req.Title
	}

	if req.Description != "" {
		arg.Description = pgtype.Text{String: req.Description, Valid: true}
	}

	if req.Status != "" {
		arg.Status = req.Status
	}

	if req.Priority != "" {
		arg.Priority = req.Priority
	}

	if req.AssignedTo != nil {
		assignedTo, err := parseAssignedTo(req.AssignedTo)
		if err != nil {
			return nil, err
		}
		if err := s.validateAssignee(ctx, project, assignedTo); err != nil {
			return nil, err
		}
		arg.AssignedTo = assignedTo
	}

	if req.DueDate != nil {
		arg.DueDate = parseDueDate(req.DueDate)
	}

	updated, err := s.repo.UpdateTask(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to update task: %w", err)
	}

	return toTaskResponse(updated), nil
}

func (s *service) DeleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	if _, _, err := s.getTaskAndVerifyAccess(ctx, taskID, userID); err != nil {
		return err
	}

	if err := s.repo.DeleteTask(ctx, taskID); err != nil {
		return fmt.Errorf("Failed to delete task: %w", err)
	}

	return nil
}

func (s *service) AddComment(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, req CreateCommentRequest) (*CommentResponse, error) {
	if _, _, err := s.getTaskAndVerifyAccess(ctx, taskID, userID); err != nil {
		return nil, err
	}

	comment, err := s.repo.CreateComment(ctx, db.CreateCommentParams{
		TaskID:  taskID,
		UserID:  userID,
		Content: req.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create comment: %w", err)
	}

	return &CommentResponse{
		ID:        comment.ID.String(),
		TaskID:    comment.TaskID.String(),
		UserID:    comment.UserID.String(),
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Time,
	}, nil
}

func (s *service) ListComments(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) ([]CommentResponse, error) {
	if _, _, err := s.getTaskAndVerifyAccess(ctx, taskID, userID); err != nil {
		return nil, err
	}

	comments, err := s.repo.ListCommentsByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("Failed to list comments: %w", err)
	}

	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = CommentResponse{
			ID:        c.ID.String(),
			TaskID:    c.TaskID.String(),
			UserID:    c.UserID.String(),
			UserName:  c.UserName,
			UserEmail: c.UserEmail,
			Content:   c.Content,
			CreatedAt: c.CreatedAt.Time,
		}
	}

	return responses, nil
}

func (s *service) DeleteComment(ctx context.Context, taskID uuid.UUID, commentID uuid.UUID, userID uuid.UUID) error {
	_, project, err := s.getTaskAndVerifyAccess(ctx, taskID, userID)
	if err != nil {
		return err
	}

	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil || comment.TaskID != taskID {
		return response.ErrNotFound("Comment not found")
	}

	if comment.UserID != userID && project.OwnerID != userID {
		return response.ErrForbidden("Unauthorized to delete this comment")
	}

	if err := s.repo.DeleteComment(ctx, commentID); err != nil {
		return fmt.Errorf("Failed to delete comment: %w", err)
	}

	return nil
}

func (s *service) verifyProjectAccess(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (db.Project, error) {
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return db.Project{}, response.ErrNotFound("Project not found")
	}

	if project.OwnerID == userID {
		return project, nil
	}

	_, err = s.projectRepo.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return db.Project{}, response.ErrForbidden("Unauthorized to access this project")
	}

	return project, nil
}

func (s *service) getTaskAndVerifyAccess(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (db.Task, db.Project, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return db.Task{}, db.Project{}, response.ErrNotFound("Task not found")
	}

	project, err := s.verifyProjectAccess(ctx, task.ProjectID, userID)
	if err != nil {
		return db.Task{}, db.Project{}, err
	}

	return task, project, nil
}

func (s *service) validateAssignee(ctx context.Context, project db.Project, assignedTo pgtype.UUID) error {
	if !assignedTo.Valid {
		return nil
	}

	assigneeID := uuid.UUID(assignedTo.Bytes)
	if assigneeID == project.OwnerID {
		return nil
	}

	_, err := s.projectRepo.GetProjectMember(ctx, db.GetProjectMemberParams{
		ProjectID: project.ID,
		UserID:    assigneeID,
	})
	if err != nil {
		return response.ErrBadRequest("Assignee is not a member of this project")
	}

	return nil
}

func toTaskResponse(t db.Task) *TaskResponse {
	var desc string
	if t.Description.Valid {
		desc = t.Description.String
	}

	var assignedTo *string
	if t.AssignedTo.Valid {
		idStr := uuid.UUID(t.AssignedTo.Bytes).String()
		assignedTo = &idStr
	}

	var dueDate *time.Time
	if t.DueDate.Valid {
		d := t.DueDate.Time
		dueDate = &d
	}

	return &TaskResponse{
		ID:          t.ID.String(),
		ProjectID:   t.ProjectID.String(),
		Title:       t.Title,
		Description: desc,
		Status:      t.Status,
		Priority:    t.Priority,
		AssignedTo:  assignedTo,
		DueDate:     dueDate,
		CreatedAt:   t.CreatedAt.Time,
		UpdatedAt:   t.UpdatedAt.Time,
	}
}

func parseAssignedTo(assignedTo *string) (pgtype.UUID, error) {
	if assignedTo == nil || *assignedTo == "" {
		return pgtype.UUID{Valid: false}, nil
	}

	parsedUUID, err := uuid.Parse(*assignedTo)
	if err != nil {
		return pgtype.UUID{}, response.ErrBadRequest("Invalid assigned_to UUID")
	}

	return pgtype.UUID{Bytes: parsedUUID, Valid: true}, nil
}

func parseDueDate(dueDate *time.Time) pgtype.Timestamptz {
	if dueDate == nil {
		return pgtype.Timestamptz{Valid: false}
	}

	return pgtype.Timestamptz{Time: *dueDate, Valid: true}
}
