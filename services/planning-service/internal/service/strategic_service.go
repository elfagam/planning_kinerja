package service

import (
	"eplan/services/planning-service/internal/models"
	"eplan/services/planning-service/internal/repository"
)

type StrategicService struct {
	repo repository.StrategicRepository
}

func NewStrategicService(repo repository.StrategicRepository) StrategicService {
	return StrategicService{repo: repo}
}

func (s StrategicService) ListByType(planType string) ([]models.StrategicPlan, error) {
	return s.repo.ListByType(planType)
}
