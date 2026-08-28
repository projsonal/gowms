package task

import (
	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type Filter struct {
	Status string

	AssignedTo uint
}

type Repository interface {
	List(p utils.PaginationParams, f Filter) ([]model.Task, int64, error)
	FindByID(id uint) (*model.Task, error)
	Create(t *model.Task) error
	Update(t *model.Task) error
	Delete(id uint) error

	CountByStatus(assignedTo uint, status string) (int64, error)
	CountOverdue(assignedTo uint) (int64, error)
}
