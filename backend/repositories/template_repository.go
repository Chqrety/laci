package repositories

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"laci-backend/config"
	"laci-backend/models"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) FindActiveTemplates() ([]models.SuratTemplate, error) {
	var templates []models.SuratTemplate
	if err := r.db.Where("is_active = ?", true).Order("nama_template ASC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *TemplateRepository) FindTemplateByID(id uuid.UUID) (*models.SuratTemplate, error) {
	var template models.SuratTemplate
	if err := r.db.First(&template, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) UpsertTemplates(templates []models.SuratTemplate) ([]models.SuratTemplate, error) {
	var saved []models.SuratTemplate
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range templates {
			var existing models.SuratTemplate
			tx.Where("g_doc_id = ?", item.GDocID).Limit(1).Find(&existing)
			if existing.ID != uuid.Nil {
				existing.NamaTemplate = item.NamaTemplate
				existing.Placeholders = item.Placeholders
				existing.IsActive = true
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				saved = append(saved, existing)
			} else {
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
				saved = append(saved, item)
			}
		}
		return nil
	})
	return saved, err
}

func (r *TemplateRepository) GetSetting(key, fallback string) string {
	var setting models.AppSetting
	if err := r.db.Where("key = ?", key).First(&setting).Error; err == nil {
		if strings.TrimSpace(setting.Value) != "" {
			return setting.Value
		}
	}
	return config.GetEnv(key, fallback)
}

func (r *TemplateRepository) SetSetting(key, value string) error {
	setting := models.AppSetting{Key: key, Value: value}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&setting).Error
}

func (r *TemplateRepository) GetAllSettings() (map[string]string, error) {
	var list []models.AppSetting
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for _, s := range list {
		res[s.Key] = s.Value
	}
	return res, nil
}
