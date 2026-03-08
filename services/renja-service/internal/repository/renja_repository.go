package repository

import "eplan/services/renja-service/internal/models"

type RenjaRepository interface {
	List() ([]models.Renja, error)
}

type InMemoryRenjaRepository struct{}

func (r InMemoryRenjaRepository) List() ([]models.Renja, error) {
	return []models.Renja{{ID: 1, Code: "RNJ-2026", Name: "Renja 2026", Status: "DRAFT"}}, nil
}
