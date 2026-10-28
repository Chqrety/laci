package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"laci-backend/models"
	"laci-backend/repositories"
)

var (
	ErrInstansiNotFound  = errors.New("Instansi tidak ditemukan")
	ErrInstansiNameEmpty = errors.New("Nama instansi wajib diisi")
	ErrInstansiNameTaken = errors.New("Nama instansi sudah terdaftar")
)

type InstansiService struct {
	instansiRepo *repositories.InstansiRepository
}

func NewInstansiService(instansiRepo *repositories.InstansiRepository) *InstansiService {
	return &InstansiService{instansiRepo: instansiRepo}
}

func (s *InstansiService) GetAllInstansi() ([]models.Instansi, error) {
	return s.instansiRepo.FindAll()
}

func (s *InstansiService) CreateInstansi(namaInstansi, kategori, keterangan string) (*models.Instansi, error) {
	trimmedNama := strings.TrimSpace(namaInstansi)
	if trimmedNama == "" {
		return nil, ErrInstansiNameEmpty
	}

	trimmedKategori := strings.TrimSpace(kategori)
	if trimmedKategori == "" {
		trimmedKategori = "Internal Kampus"
	}

	if count, err := s.instansiRepo.CountByNameExcludingID(trimmedNama, uuid.Nil); err == nil && count > 0 {
		return nil, ErrInstansiNameTaken
	}

	now := time.Now()
	ins := models.Instansi{
		ID:           uuid.New(),
		NamaInstansi: trimmedNama,
		Kategori:     trimmedKategori,
		Keterangan:   strings.TrimSpace(keterangan),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.instansiRepo.Create(&ins); err != nil {
		return nil, err
	}

	return &ins, nil
}

func (s *InstansiService) UpdateInstansi(id uuid.UUID, namaInstansi, kategori, keterangan string) (*models.Instansi, error) {
	ins, err := s.instansiRepo.FindByID(id)
	if err != nil {
		return nil, ErrInstansiNotFound
	}

	trimmedNama := strings.TrimSpace(namaInstansi)
	if trimmedNama == "" {
		return nil, ErrInstansiNameEmpty
	}

	trimmedKategori := strings.TrimSpace(kategori)
	if trimmedKategori == "" {
		trimmedKategori = ins.Kategori
	}

	if count, err := s.instansiRepo.CountByNameExcludingID(trimmedNama, id); err == nil && count > 0 {
		return nil, ErrInstansiNameTaken
	}

	ins.NamaInstansi = trimmedNama
	ins.Kategori = trimmedKategori
	ins.Keterangan = strings.TrimSpace(keterangan)
	ins.UpdatedAt = time.Now()

	if err := s.instansiRepo.Update(ins); err != nil {
		return nil, err
	}

	return ins, nil
}

func (s *InstansiService) DeleteInstansi(id uuid.UUID) error {
	_, err := s.instansiRepo.FindByID(id)
	if err != nil {
		return ErrInstansiNotFound
	}
	return s.instansiRepo.Delete(id)
}
