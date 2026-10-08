package seeders

import (
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

type Proker struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NamaProker string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"nama_proker"`
	Deskripsi  string    `gorm:"type:text" json:"deskripsi"`
}

func (Proker) TableName() string {
	return "proker"
}

type Surat struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NomorSurat    string     `gorm:"type:varchar(255);not null" json:"nomor_surat"`
	Perihal       string     `gorm:"type:varchar(255);not null" json:"perihal"`
	StatusSaatIni string     `gorm:"type:varchar(100);not null" json:"status_saat_ini"`
	PICNama       string     `gorm:"type:varchar(100);not null" json:"pic_nama"`
	ArsipURL      *string    `gorm:"type:text" json:"arsip_url"`
	FileArsip     string     `gorm:"type:text" json:"file_arsip"`
	Catatan       string     `gorm:"type:text" json:"catatan"`
	ProkerID      *uuid.UUID `gorm:"type:uuid" json:"proker_id,omitempty"`
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

func RunAllSeeders(db *gorm.DB) {
	seedUsers(db)
	seedProker(db)
	seedSurat(db)
	seedLogTracking(db)
}

func seedUsers(db *gorm.DB) {
	var sekreCount int64
	db.Model(&User{}).Where("username = ?", "sekre").Count(&sekreCount)
	if sekreCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash default password: %v", err)
		}
		defaultUser := User{
			ID:       uuid.New(),
			Username: "sekre",
			Password: string(hashedPassword),
			Role:     "sekre",
		}
		if err := db.Create(&defaultUser).Error; err != nil {
			log.Printf("Warning: failed to seed default admin user: %v", err)
		}
	}

	var humasCount int64
	db.Model(&User{}).Where("username = ?", "humas").Count(&humasCount)
	if humasCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash default password: %v", err)
		}
		humasUser := User{
			ID:       uuid.New(),
			Username: "humas",
			Password: string(hashedPassword),
			Role:     "humas",
		}
		if err := db.Create(&humasUser).Error; err != nil {
			log.Printf("Warning: failed to seed default user 'humas': %v", err)
		}
	}
}

func seedProker(db *gorm.DB) {
	defaultProkers := []Proker{
		{NamaProker: "MAKRAB", Deskripsi: "Malam Keakraban"},
		{NamaProker: "OPREC", Deskripsi: "Open Recruitment"},
		{NamaProker: "OOTS", Deskripsi: "Open Order T-Shirt"},
		{NamaProker: "DST", Deskripsi: "Doscom Software Training"},
		{NamaProker: "DU", Deskripsi: "Doscom University"},
		{NamaProker: "RP", Deskripsi: "Release Party"},
	}

	for _, p := range defaultProkers {
		var proker Proker
		p.ID = uuid.New()
		if err := db.Where(Proker{NamaProker: p.NamaProker}).FirstOrCreate(&proker, p).Error; err != nil {
			log.Printf("Warning: failed to seed proker %s: %v", p.NamaProker, err)
		}
	}
}

func seedSurat(db *gorm.DB) {
	var prokers []Proker
	db.Find(&prokers)

	prokerMap := make(map[string]uuid.UUID)
	for _, p := range prokers {
		prokerMap[p.NamaProker] = p.ID
	}

	getProkerID := func(nama string) *uuid.UUID {
		if id, exists := prokerMap[nama]; exists {
			return &id
		}
		if len(prokers) > 0 {
			return &prokers[0].ID
		}
		return nil
	}

	dummyList := []struct {
		Nomor   string
		Perihal string
		Proker  string
		Status  string
		PIC     string
		Catatan string
	}{
		{
			Nomor:   "015/DOSCOM/MAKRAB/X/2026",
			Perihal: "Permohonan Peminjaman Ruang Gedung H",
			Proker:  "MAKRAB",
			Status:  "Standby",
			PIC:     "Ahmad Fauzi",
			Catatan: "Menunggu lampiran denah ruang untuk kegiatan Malam Keakraban",
		},
		{
			Nomor:   "018/DOSCOM/OPREC/X/2026",
			Perihal: "Surat Izin Penyelenggaraan Acara",
			Proker:  "OPREC",
			Status:  "Cetak",
			PIC:     "Nabila Putri",
			Catatan: "Disposisi disetujui, siap cetak fisik untuk ditandatangani",
		},
		{
			Nomor:   "021/DOSCOM/DST/X/2026",
			Perihal: "Peminjaman Proyektor dan Sound System",
			Proker:  "DST",
			Status:  "TTD Lapis 1",
			PIC:     "Dimas Arya",
			Catatan: "Menunggu tanda tangan sekretaris umum DOSCOM",
		},
		{
			Nomor:   "022/DOSCOM/DU/X/2026",
			Perihal: "Permohonan SK Kepengurusan",
			Proker:  "DU",
			Status:  "TTD Ketum",
			PIC:     "Siti Rahma",
			Catatan: "Menunggu tanda tangan dan validasi Ketua Umum DOSCOM",
		},
		{
			Nomor:   "025/DOSCOM/RP/X/2026",
			Perihal: "Surat Undangan Pembicara",
			Proker:  "RP",
			Status:  "TTD Pembina",
			PIC:     "Rian Pratama",
			Catatan: "Sudah disetujui Ketum, saat ini menunggu tanda tangan Pembina DOSCOM",
		},
		{
			Nomor:   "028/DOSCOM/OOTS/X/2026",
			Perihal: "Permohonan Izin Stand Bazaar Kampus",
			Proker:  "OOTS",
			Status:  "Paraf Koormawa",
			PIC:     "Fajar Santoso",
			Catatan: "Telah diajukan ke BEM, sedang proses paraf Koordinator Kemahasiswaan",
		},
		{
			Nomor:   "030/DOSCOM/MAKRAB/X/2026",
			Perihal: "Permohonan Rekomendasi Kegiatan Luar Kampus",
			Proker:  "MAKRAB",
			Status:  "TTD Tertinggi",
			PIC:     "Lestari Widya",
			Catatan: "Menunggu tanda tangan Dekan / Wakil Rektor III",
		},
		{
			Nomor:   "032/DOSCOM/DST/X/2026",
			Perihal: "Surat Tugas Instruktur Workshop",
			Proker:  "DST",
			Status:  "Selesai",
			PIC:     "Bayu Saputra",
			Catatan: "Semua tanda tangan lengkap dan surat fisik telah diarsipkan",
		},
	}

	for _, d := range dummyList {
		var s Surat
		newSurat := Surat{
			ID:            uuid.New(),
			NomorSurat:    d.Nomor,
			Perihal:       d.Perihal,
			StatusSaatIni: d.Status,
			PICNama:       d.PIC,
			Catatan:       d.Catatan,
			ProkerID:      getProkerID(d.Proker),
		}

		if err := db.Where(Surat{NomorSurat: d.Nomor}).FirstOrCreate(&s, newSurat).Error; err != nil {
			log.Printf("Warning: failed to seed surat %s: %v", d.Nomor, err)
			continue
		}
	}
}

func seedLogTracking(db *gorm.DB) {
	statusHierarchy := []string{
		"Standby",
		"Cetak",
		"TTD Lapis 1",
		"TTD Ketum",
		"TTD Pembina",
		"Paraf Koormawa",
		"TTD Tertinggi",
		"Selesai",
	}

	statusNoteMap := map[string]string{
		"Standby":        "Surat dibuat dan masuk ke sistem tracking",
		"Cetak":          "Surat telah dicetak fisik dan disiapkan untuk penandatanganan",
		"TTD Lapis 1":    "Surat telah ditandatangani oleh Sekretaris Umum",
		"TTD Ketum":      "Surat telah ditandatangani oleh Ketua Umum DOSCOM",
		"TTD Pembina":    "Surat telah ditandatangani oleh Pembina DOSCOM",
		"Paraf Koormawa": "Surat telah diparaf oleh Koordinator Kemahasiswaan",
		"TTD Tertinggi":  "Surat telah ditandatangani oleh Pimpinan Tertinggi Kampus",
		"Selesai":        "Surat selesai diproses, diarsipkan, dan siap didistribusikan",
	}

	var allSurat []Surat
	if err := db.Find(&allSurat).Error; err != nil {
		log.Printf("Warning: failed to query surat for log tracking seeder: %v", err)
		return
	}

	for _, s := range allSurat {
		targetIdx := -1
		for idx, st := range statusHierarchy {
			if strings.EqualFold(st, s.StatusSaatIni) {
				targetIdx = idx
				break
			}
		}

		if targetIdx == -1 {
			targetIdx = 0
		}

		for i := 0; i <= targetIdx; i++ {
			stepStatus := statusHierarchy[i]
			note := statusNoteMap[stepStatus]
			if i == targetIdx && s.Catatan != "" {
				note = s.Catatan
			}

			offsetHours := time.Duration((targetIdx-i)*14 + 2)
			logTime := time.Now().Add(-offsetHours * time.Hour)

			newLog := LogTracking{
				ID:          uuid.New(),
				SuratID:     s.ID,
				StatusBaru:  stepStatus,
				CatatanLog:  note,
				WaktuUpdate: logTime,
			}

			var existing LogTracking
			if err := db.Where("surat_id = ? AND status_baru = ?", s.ID, stepStatus).FirstOrCreate(&existing, newLog).Error; err != nil {
				log.Printf("Warning: failed to seed log tracking for surat %s status %s: %v", s.NomorSurat, stepStatus, err)
			}
		}
	}
}
