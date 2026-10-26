package repositories

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"laci-backend/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsernameOrEmailOrNIM(identifier string) (*models.User, error) {
	var user models.User
	clean := strings.TrimSpace(identifier)
	err := r.db.Where("LOWER(username) = LOWER(?) OR LOWER(nim) = LOWER(?) OR LOWER(email) = LOWER(?)", clean, clean, clean).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	clean := strings.ToLower(strings.TrimSpace(email))
	if err := r.db.Where("LOWER(email) = ? AND is_active = true", clean).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByActivationToken(token string, now time.Time) (*models.User, error) {
	var user models.User
	err := r.db.Where("activation_token = ? AND activation_token_expires_at > ?", token, now).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByResetToken(token string, now time.Time) (*models.User, error) {
	var user models.User
	err := r.db.Where("reset_password_token = ? AND reset_password_expires_at > ?", token, now).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	if err := r.db.Order("nama ASC, username ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) CountByUsernameExcludingID(username string, excludeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("LOWER(username) = LOWER(?) AND id != ?", strings.TrimSpace(username), excludeID).Count(&count).Error
	return count, err
}

func (r *UserRepository) CountByNIMExcludingID(nim string, excludeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("LOWER(nim) = LOWER(?) AND id != ?", strings.TrimSpace(nim), excludeID).Count(&count).Error
	return count, err
}

func (r *UserRepository) CountByNamaExcludingID(nama string, excludeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("LOWER(nama) = LOWER(?) AND id != ?", strings.TrimSpace(nama), excludeID).Count(&count).Error
	return count, err
}

func (r *UserRepository) ExistsByUsernameOrNIM(username, nim string) bool {
	var count int64
	r.db.Model(&models.User{}).Where("username = ? OR (nim != '' AND nim = ?)", strings.TrimSpace(username), strings.TrimSpace(nim)).Count(&count)
	return count > 0
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Updates(user *models.User, updates map[string]interface{}) error {
	return r.db.Model(user).Updates(updates).Error
}

func (r *UserRepository) UpdateColumn(user *models.User, column string, value interface{}) error {
	return r.db.Model(user).Update(column, value).Error
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *UserRepository) DeleteBatch(ids []uuid.UUID) (int64, error) {
	res := r.db.Where("id IN ?", ids).Delete(&models.User{})
	return res.RowsAffected, res.Error
}

func (r *UserRepository) TransferKetum(oldKetumID, targetUserID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", oldKetumID).Update("divisi", "ex-ketum").Error; err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).Where("id = ?", targetUserID).Update("divisi", "ketum").Error; err != nil {
			return err
		}
		return nil
	})
}
