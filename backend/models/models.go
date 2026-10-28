package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                       uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NIM                      string     `gorm:"type:varchar(50);index" json:"nim"`
	Nama                     string     `gorm:"type:varchar(150);not null;default:''" json:"nama"`
	Email                    string     `gorm:"type:varchar(150);index" json:"email"`
	Username                 string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password                 string     `gorm:"type:varchar(255);not null" json:"-"`
	Divisi                   string     `gorm:"type:varchar(100);not null;default:'humas'" json:"divisi"`
	IsKoor                   bool       `gorm:"type:boolean;not null;default:false" json:"is_koor"`
	IsActive                 bool       `gorm:"type:boolean;not null;default:false" json:"is_active"`
	ActivationToken          *string    `gorm:"type:varchar(255);index" json:"-"`
	ActivationTokenExpiresAt *time.Time `json:"-"`
	ResetPasswordToken       *string    `gorm:"type:varchar(255);index" json:"-"`
	ResetPasswordExpiresAt   *time.Time `json:"-"`
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
	GDocsURL      *string       `gorm:"column:g_docs_url;type:text" json:"gdocs_url,omitempty"`
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

type SuratTemplate struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NamaTemplate string    `gorm:"type:varchar(255);not null" json:"nama_template"`
	GDocID       string    `gorm:"column:g_doc_id;type:varchar(255);uniqueIndex;not null" json:"gdoc_id"`
	Placeholders string    `gorm:"type:text" json:"placeholders"` // Disimpan sebagai JSON array string
	IsActive     bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
}

func (SuratTemplate) TableName() string {
	return "surat_templates"
}

type AppSetting struct {
	Key   string `gorm:"type:varchar(100);primaryKey" json:"key"`
	Value string `gorm:"type:text;not null" json:"value"`
}

func (AppSetting) TableName() string {
	return "app_settings"
}

type Instansi struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NamaInstansi string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"nama_instansi"`
	Kategori     string    `gorm:"type:varchar(100);not null;default:'Internal Kampus'" json:"kategori"`
	Keterangan   string    `gorm:"type:text" json:"keterangan"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;not null" json:"updated_at"`
}

func (Instansi) TableName() string {
	return "instansi"
}
