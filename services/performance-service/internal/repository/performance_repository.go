package repository

import "eplan/services/performance-service/internal/models"

type PerformanceRepository interface {
	List() ([]models.Performance, error)
}

type InMemoryPerformanceRepository struct{}

func (r InMemoryPerformanceRepository) List() ([]models.Performance, error) {
	return []models.Performance{{ID: 1, IndicatorName: "BOR", Target: 75, Realization: 70}}, nil
}
