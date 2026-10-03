package task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/project"
	"github.com/khoirirrosikin/task-management/internal/response"
)

type Service interface {
	CreateTask(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, req CreateTaskRequest) (*TaskResponse, error)
	GetTaskByID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*TaskResponse, error)
	ListTaskByProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]TaskResponse, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, req UpdateTaskRequest) (*TaskResponse, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
}

type service struct {
	repo Repository
	projectRepo project.Repository
}

func NewService(repo Repository, projectRepo project.Repository) Service {
	return &service{
		repo: repo,
		projectRepo: projectRepo,
	}
}

func (s *service) CreateTask(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, req CreateTaskRequest) (*TaskResponse, error) {
	if _, err := s.verifyProjectOwner(ctx, projectID, userID); err != nil {
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

	arg := db.CreateTaskParams{
		ProjectID: projectID,
		Title: req.Title,
		Description: pgtype.Text{
			String: req.Description,
			Valid: req.Description != "",
		},
		Status: status,
		Priority: priority,
		AssignedTo: assignedTo,
		DueDate: parseDueDate(req.DueDate),
	}

	task, err := s.repo.CreateTask(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to create task: %w", err)
	}

	return toTaskResponse(task), nil
}

func (s *service) GetTaskByID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*TaskResponse, error) {
	task, err := s.getTaskAndVerifyAccess(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}

	return toTaskResponse(task), nil
}

func (s *service) ListTaskByProject(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) ([]TaskResponse, error) {
	if _, err := s.verifyProjectOwner(ctx, projectID, userID); err != nil {
		return nil, err
	}

	tasks, err := s.repo.ListTaskByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("Failed to list tasks: %w", err)
	}

	responses := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = *toTaskResponse(t)
	}

	return responses, nil
}

func (s *service) UpdateTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, req UpdateTaskRequest) (*TaskResponse, error) {
	existingTask, err := s.getTaskAndVerifyAccess(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}

	arg := db.UpdateTaskParams{
		ID: taskID,
		Title: existingTask.Title,
		Description: existingTask.Description,
		Status: existingTask.Status,
		Priority: existingTask.Priority,
		AssignedTo: existingTask.AssignedTo,
		DueDate: existingTask.DueDate,
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
	if _, err := s.getTaskAndVerifyAccess(ctx, taskID, userID); err != nil {
		return err
	}

	if err := s.repo.DeleteTask(ctx, taskID); err != nil {
		return fmt.Errorf("Failed to delete task: %w", err)
	}

	return nil
}

func (s *service) verifyProjectOwner(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (db.Project, error) {
	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return db.Project{}, response.ErrNotFound("Project not found")
	}

	if project.OwnerID != userID {
		return db.Project{}, response.ErrForbidden("Unauthorized to access this project")
	}

	return project, nil
}

func (s *service) getTaskAndVerifyAccess(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (db.Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return db.Task{}, response.ErrNotFound("Task not found")
	}

	if _, err := s.verifyProjectOwner(ctx, task.ProjectID, userID); err != nil {
		return db.Task{}, err
	}

	return task, nil
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
		ID: t.ID.String(),
		ProjectID: t.ProjectID.String(),
		Title: t.Title,
		Description: desc,
		Status: t.Status,
		Priority: t.Priority,
		AssignedTo: assignedTo,
		DueDate: dueDate,
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
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