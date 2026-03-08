package service

import (
	"eplan/services/performance-service/internal/models"
	"eplan/services/performance-service/internal/repository"
)

type PerformanceService struct {
	repo repository.PerformanceRepository
}

func NewPerformanceService(repo repository.PerformanceRepository) PerformanceService {
	return PerformanceService{repo: repo}
}

func (s PerformanceService) List() ([]models.Performance, error) {
	return s.repo.List()
}
