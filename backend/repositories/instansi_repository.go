package repositories

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"laci-backend/models"
)

type InstansiRepository struct {
	db *gorm.DB
}

func NewInstansiRepository(db *gorm.DB) *InstansiRepository {
	return &InstansiRepository{db: db}
}

func (r *InstansiRepository) FindAll() ([]models.Instansi, error) {
	var list []models.Instansi
	if err := r.db.Order("nama_instansi asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *InstansiRepository) FindByID(id uuid.UUID) (*models.Instansi, error) {
	var ins models.Instansi
	if err := r.db.First(&ins, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ins, nil
}

func (r *InstansiRepository) CountByNameExcludingID(name string, excludeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Instansi{}).Where("LOWER(nama_instansi) = ? AND id != ?", strings.ToLower(strings.TrimSpace(name)), excludeID).Count(&count).Error
	return count, err
}

func (r *InstansiRepository) Create(ins *models.Instansi) error {
	return r.db.Create(ins).Error
}

func (r *InstansiRepository) Update(ins *models.Instansi) error {
	return r.db.Save(ins).Error
}

func (r *InstansiRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Instansi{}, "id = ?", id).Error
}
