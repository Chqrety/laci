package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password string    `gorm:"type:varchar(255);not null" json:"-"`
	Role     string    `gorm:"type:varchar(50);not null" json:"role"`
}

func (User) TableName() string {
	return "users"
}

type Surat struct {
	ID            uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NomorSurat    string        `gorm:"type:varchar(255);not null" json:"nomor_surat"`
	Perihal       string        `gorm:"type:varchar(255);not null" json:"perihal"`
	StatusSaatIni string        `gorm:"type:varchar(100);not null" json:"status_saat_ini"`
	PICNama       string        `gorm:"type:varchar(100);not null" json:"pic_nama"`
	ArsipURL      *string       `gorm:"type:text" json:"arsip_url"`
	FileArsip     string        `gorm:"type:text" json:"file_arsip"`
	Catatan       string        `gorm:"type:text" json:"catatan"`
	ProkerID      *uuid.UUID    `gorm:"type:uuid" json:"proker_id,omitempty"`
	Logs          []LogTracking `gorm:"foreignKey:SuratID" json:"logs,omitempty"`
}

func (Surat) TableName() string {
	return "surat"
}

type LogTracking struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SuratID     uuid.UUID `gorm:"type:uuid;not null;index" json:"surat_id"`
	StatusBaru  string    `gorm:"type:varchar(100);not null" json:"status_baru"`
	CatatanLog  string    `gorm:"type:text" json:"catatan_log"`
	WaktuUpdate time.Time `gorm:"type:timestamp;not null" json:"waktu_update"`
}

func (LogTracking) TableName() string {
	return "log_tracking"
}

type Proker struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NamaProker string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"nama_proker"`
	Deskripsi  string    `gorm:"type:text" json:"deskripsi"`
}

func (Proker) TableName() string {
	return "proker"
}
