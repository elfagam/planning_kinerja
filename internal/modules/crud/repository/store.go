package repository

import (
	"errors"

	"e-plan-ai/internal/modules/crud/domain"
)

var ErrNotFound = errors.New("data not found")

type Store interface {
	List(resource string, filter domain.ListFilter) ([]domain.Record, int64, error)
	Create(resource string, payload domain.Payload) (domain.Record, error)
	Get(resource string, id int64) (domain.Record, error)
	Update(resource string, id int64, payload domain.Payload) (domain.Record, error)
	Delete(resource string, id int64) error
}
