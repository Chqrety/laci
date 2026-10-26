# Rencana Revisi 2: Refactoring Arsitektur Backend (Pragmatic Layered Architecture)

Dokumen ini merumuskan blueprint restrukturisasi arsitektur backend Go pada repositori **LACI DOSCOM** agar lebih modular, mudah diuji (*testable*), memiliki pemisahan tanggung jawab yang tegas (*Separation of Concerns*), namun tetap mematuhi prinsip **Ponytail Protocol (Zero Bloat, YAGNI, No Over-Engineering)**.

---

## 1. Latar Belakang & Masalah Saat Ini

Saat ini, seluruh logika backend LACI terpusat di dalam folder `controllers/` (*Fat Controller / Smart UI Pattern*):
1. **Campur Aduk Tanggung Jawab:** Controller Gin (`*gin.Context`) melakukan HTTP binding, validasi otorisasi, logika bisnis (enkripsi bcrypt, token aktivasi, pemanggilan SMTP, HTTP request ke Google Apps Script), sekaligus mengeksekusi query database GORM secara langsung.
2. **Sulit Diuji (*Untestable*):** Logika bisnis tidak dapat diuji secara mandiri (unit test) tanpa harus memalsukan (*mock*) engine HTTP Gin dan basis data fisik PostgreSQL sekaligus.
3. **Duplikasi Query & Alur:** Pola transaksi GORM dan query pencarian berulang di beberapa controller tanpa abstraksi akses data terpusat.

---

## 2. Pilihan Arsitektur: Formal Enterprise vs. Pragmatic Lean

| Aspek | Formal 4-Layer (Enterprise "Java-in-Go") | **Pragmatic Lean Layered (Gaya LACI - Ponytail)** |
| :--- | :--- | :--- |
| **Pemisahan Lapisan** | 4 Lapisan: Handler $\rightarrow$ Usecase $\rightarrow$ Repository $\rightarrow$ Entity | **3 Lapisan Logika + Shared Models:** Handler $\rightarrow$ Service $\rightarrow$ Repository |
| **Penggunaan Interface** | Wajib interface di setiap struct (misal: `UserService` interface + `userService` struct) | **Concrete Structs First.** Interface hanya dibuat jika komponen benar-benar membutuhkan mocking di unit test. |
| **DTO (Data Transfer Object)** | Membuat 3-4 struct terpisah (`UserRequest`, `UserDTO`, `UserEntity`, `UserResponse`) dengan fungsi mapping manual | **Shared Models & Lightweight DTO.** Struct domain langsung digunakan jika bentuk datanya identik, hanya buat DTO untuk payload input/output khusus. |
| **Dependency Injection** | Menggunakan framework kompleks (Google Wire, Uber Dig, dsb.) | **Pure Manual Constructor Injection (`NewXxx`)** sederhana, transparan, dan mudah dibaca tanpa code-generation tambahan. |
| **File Sprawl (Jumlah Berkas)** | Sangat banyak berkas kecil (30+ berkas baru hanya untuk deklarasi interface & mock) | **Fokus & Ringkas.** Berkas dikelompokkan per domain fungsional: `auth`, `user`, `surat`, `template`, `proker`. |

---

## 3. Desain Lapisan LACI (Pragmatic 3-Tier)

```
[ HTTP Request (Client / Browser) ]
                │
                ▼
┌────────────────────────────────────────────────────────┐
│ 1. Delivery / Handler Layer (backend/handlers/)        │
│    - Dependensi: Framework Gin (*gin.Context)         │
│    - Tanggung Jawab:                                   │
│      * Binding JSON / Query / Param dari HTTP         │
│      * Mengambil JWT claims dari context               │
│      * Memanggil Service terkait                       │
│      * Memetakan error ke HTTP Status Code (4xx / 5xx) │
│      * Mengembalikan respons format JSON               │
└───────────────────────┬────────────────────────────────┘
                        │
                        ▼
┌────────────────────────────────────────────────────────┐
│ 2. Service / Business Logic Layer (backend/services/)   │
│    - Dependensi: Murni Go Standar + Repositories        │
│    - BEBAS dari referensi *gin.Context                │
│    - Tanggung Jawab:                                   │
│      * Validasi aturan bisnis (suksesi Ketum, dsb)     │
│      * Hashing kata sandi & pembuatan token crypto     │
│      * Orchestrator SMTP email & Google Apps Script    │
│      * Eksekusi transaksi bisnis lintas repositori     │
└───────────────────────┬────────────────────────────────┘
                        │
                        ▼
┌────────────────────────────────────────────────────────┐
│ 3. Repository / Data Access Layer (backend/repositories)│
│    - Dependensi: GORM (*gorm.DB)                      │
│    - Tanggung Jawab:                                   │
│      * Operasi CRUD murni ke PostgreSQL                │
│      * Preload relasi, sorting, filtering, & pagination│
│      * Transaksi database (*gorm.DB transaction)       │
└───────────────────────┬────────────────────────────────┘
                        │
                        ▼
           [ Database PostgreSQL 15 ]
```

---

## 4. Pembagian Modul & Domain

Refaktor dilakukan per domain bisnis tanpa mengubah kontrak endpoint REST API yang sudah berjalan (100% kompatibel dengan frontend):

### A. Domain `Auth` & `User`
* **Repository (`UserRepository`):**
  * `FindByID`, `FindByUsernameOrEmailOrNIM`, `FindAll`, `Create`, `Update`, `Delete`, `DeleteBatch`, `CountByDivisi`.
* **Service (`AuthService` & `UserService`):**
  * `Login(identifier, password) -> (token, user, error)`
  * `ActivateAccount(token, newPassword) -> error`
  * `ForgotPassword(email) -> error`
  * `ResetPassword(token, newPassword) -> error`
  * `TransferKetum(currentKetumID, targetUserID, currentPassword) -> error`
  * `ImportCSV(reader) -> (imported, skipped, error)`
  * `ImportSpreadsheet(sheetURL) -> (imported, skipped, error)`
* **Handler (`AuthHandler` & `UserHandler`):**
  * Menangani routing `/login`, `/auth/*`, `/profile/*`, `/users/*`.

### B. Domain `Surat` & `Kanban`
* **Repository (`SuratRepository`):**
  * `FindAllWithLogs`, `FindByID`, `CreateWithInitialLog`, `UpdateWithLog`, `DeleteWithLogs`, `UpdateArsipURL`.
* **Service (`SuratService`):**
  * Validasi batas hak unggah arsip (hanya saat `Selesai` untuk non-sekre).
  * Pengelolaan pencatatan histori perubahan status surat di `log_tracking`.
  * Upload handler file fisik ke disk lokal `/uploads`.
* **Handler (`SuratHandler`):**
  * Menangani routing `/letters` dan sub-rutenya.

### C. Domain `Template` & `Google Apps Script`
* **Repository (`TemplateRepository` & `SettingRepository`):**
  * CRUD `surat_templates` dan get/set key-value pada `app_settings`.
* **Service (`TemplateService`):**
  * Komunikasi HTTP client ke webhook Google Apps Script (`sync_templates` & `generate_surat`).
  * Parsing regex ID folder Google Drive dan format nama dokumen standar.
* **Handler (`TemplateHandler`):**
  * Menangani routing `/letters/templates/*`, `/letters/generate`, `/settings`.

### D. Domain `Proker`
* **Repository (`ProkerRepository`):**
  * `FindAll`, `FindByID`, `Create`, `Update`, `Delete` (termasuk detach `proker_id` di surat).
* **Service (`ProkerService`):**
  * Validasi duplikasi nama proker dan sanitasi string.
* **Handler (`ProkerHandler`):**
  * Menangani routing `/programs`.

---

## 5. Struktur Direktori Hasil Refaktor

```
backend/
├── config/              # database.go (Koneksi GORM)
├── models/              # models.go (Struktur entitas domain)
├── repositories/        # [BARU] Data Access Layer (GORM)
│   ├── user_repository.go
│   ├── surat_repository.go
│   ├── proker_repository.go
│   └── template_repository.go
├── services/            # [BARU] Pure Business Logic Layer
│   ├── auth_service.go
│   ├── user_service.go
│   ├── surat_service.go
│   ├── proker_service.go
│   └── template_service.go
├── handlers/            # [BARU Pengganti controllers/] HTTP Transport (Gin)
│   ├── auth_handler.go
│   ├── user_handler.go
│   ├── surat_handler.go
│   ├── proker_handler.go
│   └── template_handler.go
├── routes/              # setup.go (Inisialisasi repo, service, handler & registrasi rute)
├── seeders/             # seeder.go (SeedCore)
├── utils/               # email.go, token.go
└── main.go              # Entrypoint backend
```

---

## 6. Prinsip Keselamatan Eksekusi (*Safety & Rollback Plan*)

1. **Zero Breaking Change on API Contracts:** Seluruh URL endpoint, method HTTP, format JSON request body, dan JSON response tetap identik dengan yang diharapkan frontend.
2. **Inisialisasi Bersih Tanpa Framework Eksternal:** Wiring dependensi dilakukan secara manual di `routes/setup.go` dengan fungsi inisialisasi yang lugas.
3. **Verifikasi Bertahap:** Setiap pemindahan domain langsung diuji dengan `go build` dan `go vet` untuk memastikan tidak ada circular dependency maupun *import cycle*.
4. **Penghapusan Bersih:** Folder `controllers/` lama hanya dihapus setelah seluruh fungsi controller telah berpindah 100% ke kombinasi `handlers/`, `services/`, dan `repositories/`.
