package repository

import "eplan/services/planning-service/internal/models"

type StrategicRepository interface {
	ListByType(planType string) ([]models.StrategicPlan, error)
}

type InMemoryStrategicRepository struct{}

func (r InMemoryStrategicRepository) ListByType(planType string) ([]models.StrategicPlan, error) {
	return []models.StrategicPlan{{ID: 1, Type: planType, Code: "PLN-001", Name: "Contoh Data Planning"}}, nil
}
