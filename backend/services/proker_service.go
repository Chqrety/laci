package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"laci-backend/models"
	"laci-backend/repositories"
)

var (
	ErrProkerNotFound  = errors.New("Proker not found")
	ErrProkerNameEmpty = errors.New("nama_proker cannot be empty")
	ErrProkerNameTaken = errors.New("Nama proker sudah digunakan")
)

type ProkerService struct {
	prokerRepo *repositories.ProkerRepository
}

func NewProkerService(prokerRepo *repositories.ProkerRepository) *ProkerService {
	return &ProkerService{prokerRepo: prokerRepo}
}

func (s *ProkerService) GetAllProkers() ([]models.Proker, error) {
	return s.prokerRepo.FindAll()
}

func (s *ProkerService) CreateProker(namaProker, deskripsi string) (*models.Proker, error) {
	trimmed := strings.TrimSpace(namaProker)
	if trimmed == "" {
		return nil, ErrProkerNameEmpty
	}

	proker := models.Proker{
		ID:         uuid.New(),
		NamaProker: trimmed,
		Deskripsi:  strings.TrimSpace(deskripsi),
	}

	if err := s.prokerRepo.Create(&proker); err != nil {
		return nil, errors.New("Proker already exists or failed to save")
	}

	return &proker, nil
}

func (s *ProkerService) UpdateProker(id uuid.UUID, namaProker, deskripsi string) (*models.Proker, error) {
	proker, err := s.prokerRepo.FindByID(id)
	if err != nil {
		return nil, ErrProkerNotFound
	}

	trimmed := strings.TrimSpace(namaProker)
	if trimmed == "" {
		return nil, ErrProkerNameEmpty
	}

	count, err := s.prokerRepo.CountByNameExcludingID(trimmed, id)
	if err == nil && count > 0 {
		return nil, ErrProkerNameTaken
	}

	proker.NamaProker = trimmed
	proker.Deskripsi = strings.TrimSpace(deskripsi)

	if err := s.prokerRepo.Update(proker); err != nil {
		return nil, err
	}

	return proker, nil
}

func (s *ProkerService) DeleteProker(id uuid.UUID) error {
	_, err := s.prokerRepo.FindByID(id)
	if err != nil {
		return ErrProkerNotFound
	}
	return s.prokerRepo.Delete(id)
}
