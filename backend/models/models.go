package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NIM                   string    `gorm:"type:varchar(50);index" json:"nim"`
	Nama                  string    `gorm:"type:varchar(150);not null;default:''" json:"nama"`
	Email                 string    `gorm:"type:varchar(150);index" json:"email"`
	Username              string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password              string    `gorm:"type:varchar(255);not null" json:"-"`
	DisplayPassword       string    `gorm:"type:varchar(255);not null;default:''" json:"display_password,omitempty"`
	Divisi                string    `gorm:"type:varchar(100);not null;default:'humas'" json:"divisi"`
	IsPasswordDefault     bool      `gorm:"type:boolean;not null;default:true" json:"is_password_default"`
	HasSeenPasswordPrompt bool      `gorm:"type:boolean;not null;default:false" json:"has_seen_password_prompt"`
}

func (User) TableName() string {
	return "users"
}

type Surat struct {
	ID            uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NomorSurat    string        `gorm:"type:varchar(255);not null" json:"nomor_surat"`
	Perihal       string        `gorm:"type:varchar(255);not null" json:"perihal"`
	Tujuan        string        `gorm:"type:varchar(255);default:''" json:"tujuan"`
	StatusSaatIni string        `gorm:"type:varchar(100);not null" json:"status_saat_ini"`
	PICNama       string        `gorm:"type:varchar(100);not null" json:"pic_nama"`
	ArsipURL      *string       `gorm:"type:text" json:"arsip_url"`
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
