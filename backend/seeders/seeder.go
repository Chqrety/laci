package seeders

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"laci-backend/config"
	"laci-backend/models"
	"laci-backend/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedCore menjalankan seeder kebutuhan dasar (Genesis Ketum, Proker & Instansi)
func SeedCore(db *gorm.DB) {
	SeedGenesisKetum(db)
	SeedProker(db)
	SeedInstansi(db)
}

// SeedGenesisKetum menginisialisasi Ketua Umum pertama jika belum ada satupun akun ketum di database.
// Menggunakan Solusi 1 Opsi 1A: akun dibuat inactive dan otomatis dikirimkan email aktivasi untuk membuat password sendiri.
func SeedGenesisKetum(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Where("LOWER(divisi) = 'ketum'").Count(&count)
	if count > 0 {
		return
	}

	nim := strings.TrimSpace(os.Getenv("INIT_KETUM_NIM"))
	nama := strings.TrimSpace(os.Getenv("INIT_KETUM_NAMA"))
	email := strings.TrimSpace(os.Getenv("INIT_KETUM_EMAIL"))
	username := strings.TrimSpace(os.Getenv("INIT_KETUM_USERNAME"))

	if username == "" {
		log.Println("[GENESIS KETUM] Variabel INIT_KETUM_USERNAME belum disetel di environment. Inisialisasi dilewati.")
		return
	}

	var existing models.User
	if err := db.Where("username = ? OR (email != '' AND email = ?)", username, email).First(&existing).Error; err == nil {
		// Jika username/email sudah ada tapi bukan ketum, promosikan menjadi ketum
		db.Model(&existing).Update("divisi", "ketum")
		log.Printf("[GENESIS KETUM] Akun '%s' dipromosikan menjadi Ketua Umum.", existing.Username)
		return
	}

	activationToken := utils.GenerateToken(32)
	tokenExpiry := time.Now().Add(24 * time.Hour)

	lockedPassword, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)

	genesisKetum := models.User{
		ID:                       uuid.New(),
		NIM:                      nim,
		Nama:                     nama,
		Email:                    email,
		Username:                 username,
		Password:                 string(lockedPassword),
		Divisi:                   "ketum",
		IsActive:                 false,
		ActivationToken:          &activationToken,
		ActivationTokenExpiresAt: &tokenExpiry,
	}

	if err := db.Create(&genesisKetum).Error; err != nil {
		log.Printf("[GENESIS KETUM ERROR] Gagal membuat akun Ketua Umum pertama: %v", err)
		return
	}

	appURL := strings.TrimRight(config.GetEnv("APP_URL", ""), "/")
	var activationURL string
	if appURL != "" {
		activationURL = fmt.Sprintf("%s/activate?token=%s", appURL, activationToken)
	} else {
		activationURL = fmt.Sprintf("/activate?token=%s", activationToken)
	}

	log.Printf("================================================================================")
	log.Printf("[GENESIS KETUM] Akun Ketua Umum perdana berhasil disiapkan!")
	log.Printf("Username       : %s", username)
	log.Printf("Email          : %s", email)
	log.Printf("Token Aktivasi : %s", activationToken)
	if appURL != "" {
		log.Printf("Tautan Aktivasi: %s", activationURL)
	} else {
		log.Printf("Tautan Aktivasi: %s (buka via domain/IP server)", activationURL)
	}
	log.Printf("================================================================================")

	if email != "" {
		utils.SendActivationEmail(email, nama, username, activationToken)
	}
}

func SeedProker(db *gorm.DB) {
	defaultProkers := []models.Proker{
		{NamaProker: "MAKRAB", Deskripsi: "Malam Keakraban"},
		{NamaProker: "OPREC", Deskripsi: "Open Recruitment"},
		{NamaProker: "OOTS", Deskripsi: "Opensource On The School"},
		{NamaProker: "DST", Deskripsi: "Doscom Sharing Time"},
		{NamaProker: "DU", Deskripsi: "Doscom University"},
		{NamaProker: "RP", Deskripsi: "Release Party"},
	}

	for _, p := range defaultProkers {
		var proker models.Proker
		p.ID = uuid.New()
		if err := db.Where(models.Proker{NamaProker: p.NamaProker}).FirstOrCreate(&proker, p).Error; err != nil {
			log.Printf("Warning: failed to seed proker %s: %v", p.NamaProker, err)
		}
	}
}

func SeedInstansi(db *gorm.DB) {
	defaultInstansi := []models.Instansi{
		{NamaInstansi: "Dekan Fakultas Teknologi Cerdas", Kategori: "Fakultas", Keterangan: "Persetujuan izin kegiatan resmi DOSCOM"},
		{NamaInstansi: "Tata Usaha Fakultas Teknologi Cerdas", Kategori: "Fakultas", Keterangan: "Peminjaman ruangan kampus, sarpras, dan media promosi"},
		{NamaInstansi: "Kepala Biro Umum", Kategori: "Universitas", Keterangan: "Perizinan gedung eksternal dan parkir kampus"},
		{NamaInstansi: "Wakil Rektor II Bidang Umum dan Keuangan", Kategori: "Rektorat", Keterangan: "Fasilitas kendaraan operasional universitas"},
		{NamaInstansi: "Dosen Pengampu Mata Kuliah", Kategori: "Akademik", Keterangan: "Dispensasi izin perkuliahan panitia dan delegasi"},
		{NamaInstansi: "Kepala Poliklinik", Kategori: "Unit Kampus", Keterangan: "Bantuan obat-obatan dan P3K kegiatan"},
		{NamaInstansi: "Biro Kemahasiswaan (BIMA)", Kategori: "Universitas", Keterangan: "Permohonan plakat kenang-kenangan dan pendamping kegiatan"},
		{NamaInstansi: "Bapak/Ibu Kepala Sekolah", Kategori: "Sekolah Mitra", Keterangan: "Surat pengantar sosialisasi dan kunjungan sekolah"},
		{NamaInstansi: "Mitra Sponsor / Perusahaan", Kategori: "Eksternal", Keterangan: "Kerjasama pendanaan dan sponsorship kegiatan"},
		{NamaInstansi: "Tamu Undangan / Ormawa Kampus", Kategori: "Ormawa / Mitra", Keterangan: "Undangan menghadiri acara resmi DOSCOM"},
		{NamaInstansi: "Pembicara / Pemateri Kegiatan", Kategori: "Akademik", Keterangan: "Berita acara honorarium dan pemateri kegiatan"},
	}

	for _, ins := range defaultInstansi {
		var existing models.Instansi
		ins.ID = uuid.New()
		now := time.Now()
		ins.CreatedAt = now
		ins.UpdatedAt = now
		if err := db.Where("nama_instansi = ?", ins.NamaInstansi).FirstOrCreate(&existing, ins).Error; err != nil {
			log.Printf("Warning: failed to seed instansi %s: %v", ins.NamaInstansi, err)
		}
	}
}
