package repositories

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"laci-backend/models"
)

type ProkerRepository struct {
	db *gorm.DB
}

func NewProkerRepository(db *gorm.DB) *ProkerRepository {
	return &ProkerRepository{db: db}
}

func (r *ProkerRepository) FindAll() ([]models.Proker, error) {
	var list []models.Proker
	if err := r.db.Order("nama_proker asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ProkerRepository) FindByID(id uuid.UUID) (*models.Proker, error) {
	var proker models.Proker
	if err := r.db.First(&proker, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &proker, nil
}

func (r *ProkerRepository) CountByNameExcludingID(name string, excludeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Proker{}).Where("nama_proker = ? AND id != ?", strings.TrimSpace(name), excludeID).Count(&count).Error
	return count, err
}

func (r *ProkerRepository) Create(proker *models.Proker) error {
	return r.db.Create(proker).Error
}

func (r *ProkerRepository) Update(proker *models.Proker) error {
	return r.db.Save(proker).Error
}

func (r *ProkerRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Detach relasi proker_id dari surat
		if err := tx.Model(&models.Surat{}).Where("proker_id = ?", id).Update("proker_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Proker{}, "id = ?", id).Error
	})
}
