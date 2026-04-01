package service

import (
	"errors"
	"fmt"
	"sync"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Done        bool   `json:"done"`
}

type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

type UpdateTaskInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

type TaskService struct {
	mu      sync.RWMutex
	tasks   map[string]Task
	counter int
}

func New() *TaskService {
	return &TaskService{
		tasks: make(map[string]Task),
	}
}

func (s *TaskService) Create(input CreateTaskInput) (Task, error) {
	if input.Title == "" {
		return Task{}, errors.New("title is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	id := fmt.Sprintf("t_%03d", s.counter)

	task := Task{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		DueDate:     input.DueDate,
		Done:        false,
	}

	s.tasks[id] = task
	return task, nil
}

func (s *TaskService) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}

	return result
}

func (s *TaskService) Get(id string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	return task, nil
}

func (s *TaskService) Update(id string, input UpdateTaskInput) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.DueDate != nil {
		task.DueDate = *input.DueDate
	}
	if input.Done != nil {
		task.Done = *input.Done
	}

	if task.Title == "" {
		return Task{}, errors.New("title is required")
	}

	s.tasks[id] = task
	return task, nil
}

func (s *TaskService) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}

	delete(s.tasks, id)
	return nil
}
