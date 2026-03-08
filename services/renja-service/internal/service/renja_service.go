package service

import (
	"eplan/services/renja-service/internal/models"
	"eplan/services/renja-service/internal/repository"
)

type RenjaService struct {
	repo repository.RenjaRepository
}

func NewRenjaService(repo repository.RenjaRepository) RenjaService {
	return RenjaService{repo: repo}
}

func (s RenjaService) List() ([]models.Renja, error) {
	return s.repo.List()
}
