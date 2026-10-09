package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/encourage13/todo-api/internal/model"
	"github.com/google/uuid"
)

type todoRepository interface {
	Create(ctx context.Context, t *model.Todo) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Todo, error)
	List(ctx context.Context, userID uuid.UUID) ([]model.Todo, error)
	Update(ctx context.Context, t *model.Todo) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type TodoService struct {
	store todoRepository
}

func NewTodoService(repo todoRepository) *TodoService {
	return &TodoService{store: repo}
}

type UpdateTodoInput struct {
	Title       *string
	Description *string
	Completed   *bool
}

const (
	titleMaxLength       = 200
	descriptionMaxLength = 2000
)

func (s *TodoService) Get(ctx context.Context, userID, id uuid.UUID) (*model.Todo, error) {
	todo, err := s.store.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, model.ErrTodoNotFound) {
			return nil, model.ErrTodoNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return todo, nil
}

func (s *TodoService) List(ctx context.Context, userID uuid.UUID) ([]model.Todo, error) {
	tasks, err := s.store.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get list of tasks: %w", err)
	}
	return tasks, nil
}

func (s *TodoService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	err := s.store.Delete(ctx, id, userID)
	if err != nil {
		if errors.Is(err, model.ErrTodoNotFound) {
			return model.ErrTodoNotFound
		}
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (s *TodoService) Create(ctx context.Context, userID uuid.UUID, title, description string) (*model.Todo, error) {
	title, err := validateTitle(title)
	if err != nil {
		return nil, err
	}

	description, err = validateDescription(description)
	if err != nil {
		return nil, err
	}

	var task model.Todo
	task.UserID = userID
	task.Title = title
	task.Description = description
	err = s.store.Create(ctx, &task)

	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}

	return &task, nil
}

func (s *TodoService) Update(ctx context.Context, userID, id uuid.UUID, in UpdateTodoInput) (*model.Todo, error) {
	if in.Completed == nil && in.Description == nil && in.Title == nil {
		return nil, fmt.Errorf("nothing to update: %w", model.ErrInvalidInput)
	}

	currentTask, err := s.store.GetByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, model.ErrTodoNotFound) {
			return nil, model.ErrTodoNotFound
		}
		return nil, fmt.Errorf("get task by id: %w", err)
	}

	if in.Completed != nil {
		currentTask.Completed = *in.Completed
	}
	if in.Description != nil {
		description, err := validateDescription(*in.Description)
		if err != nil {
			return nil, err
		}
		currentTask.Description = description
	}
	if in.Title != nil {
		title, err := validateTitle(*in.Title)
		if err != nil {
			return nil, err
		}
		currentTask.Title = title
	}

	if err := s.store.Update(ctx, currentTask); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return currentTask, nil
}

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > titleMaxLength || title == "" {
		return "", fmt.Errorf("%w: title must be 1-200 characters", model.ErrInvalidInput)
	}
	return title, nil
}

func validateDescription(description string) (string, error) {
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > descriptionMaxLength {
		return "", fmt.Errorf("%w: description must be 2000 characters at maximum", model.ErrInvalidInput)
	}
	return description, nil
}
