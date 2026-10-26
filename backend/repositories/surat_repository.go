package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"laci-backend/models"
)

type SuratRepository struct {
	db *gorm.DB
}

func NewSuratRepository(db *gorm.DB) *SuratRepository {
	return &SuratRepository{db: db}
}

func (r *SuratRepository) FindAllWithLogs() ([]models.Surat, error) {
	var list []models.Surat
	if err := r.db.Preload("Logs").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *SuratRepository) FindByID(id uuid.UUID) (*models.Surat, error) {
	var surat models.Surat
	if err := r.db.First(&surat, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &surat, nil
}

func (r *SuratRepository) CreateWithLog(surat *models.Surat, log *models.LogTracking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(surat).Error; err != nil {
			return err
		}
		if err := tx.Create(log).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *SuratRepository) UpdateWithLog(surat *models.Surat, updates map[string]interface{}, log *models.LogTracking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(surat).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Create(log).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *SuratRepository) UpdateArsipURL(id uuid.UUID, arsipURL string) error {
	return r.db.Model(&models.Surat{}).Where("id = ?", id).Update("arsip_url", arsipURL).Error
}

func (r *SuratRepository) DeleteWithLogs(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var surat models.Surat
		if err := tx.First(&surat, "id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Where("surat_id = ?", id).Delete(&models.LogTracking{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&surat).Error; err != nil {
			return err
		}
		return nil
	})
}
