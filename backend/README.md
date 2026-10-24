# ⚙️ Laci Backend Service

REST API Service untuk **Laci DOSCOM** dibangun menggunakan bahasa pemrograman **Golang** dengan framework **Gin**, ORM **GORM**, dan database **PostgreSQL 15**.

---

## 🛠️ Tech Stack & Dependencies

* **Language:** Golang `1.22+`
* **HTTP Framework:** [Gin Gonic](https://github.com/gin-gonic/gin)
* **ORM & Database:** [GORM](https://gorm.io/) (PostgreSQL Driver)
* **Auth & Security:** JWT (`golang-jwt/jwt/v5`), bcrypt (`golang.org/x/crypto`)
* **Utilities:** `google/uuid`, `godotenv`

---

## 🗄️ Database Schema & Relational Design

Sistem mengandalkan database relasional terstruktur dengan auto-migration:

```
┌──────────────┐       1:N       ┌──────────────────┐
│    surat     ├────────────────►│  log_trackings   │
│  (ID: UUID)  │                 │   (ID: UUID)     │
└──────────────┘                 └──────────────────┘
       ▲
       │ (References proker code in Nomor Surat)
┌──────┴───────┐
│   prokers    │
│  (ID: UUID)  │
└──────────────┘

┌──────────────┐
│    users     │ (NIM, Username, Password, Divisi: sekre/humas)
│  (ID: UUID)  │
└──────────────┘
```

### 1. `users` (Tabel Pengurus)
* `id` (UUID, Primary Key)
* `nim` (varchar(50), Indexed) — Nomor Induk Mahasiswa pengurus.
* `nama` (varchar(255), Not Null)
* `email` (varchar(255))
* `username` (varchar(100), Unique Index, Not Null) — Auto-sanitized atau custom.
* `password` (varchar(255), Not Null) — Bcrypt hash string.
* `display_password` (text) — Password awal default (dapat dilihat Sekre sebelum dirotasi).
* `divisi` (varchar(100), Default: 'humas') — Enum terbatas: `sekre` atau `humas`.
* `is_password_default` (boolean, Default: true) — Menandai apakah user wajib rotasi sandi.
* `has_seen_password_prompt` (boolean, Default: false)

### 2. `surat` (Tabel Arsip Surat)
* `id` (UUID, Primary Key)
* `nomor_surat` (varchar(255), Unique, Not Null) — Format: `[No]/DOSCOM/[Proker]/[Romawi]/[Tahun]`.
* `perihal` (text, Not Null)
* `tujuan` (varchar(255))
* `status_saat_ini` (varchar(100)) — Contoh: `draft`, `review`, `signed`, `sent`.
* `pic_nama` (varchar(255))
* `catatan` (text)
* `arsip_url` (text) — Path berkas scan atau dokumen arsip.

### 3. `log_trackings` (Riwayat Perubahan Status Surat)
* `id` (UUID, Primary Key)
* `surat_id` (UUID, Foreign Key `surat.id` ON DELETE CASCADE)
* `status_baru` (varchar(100))
* `catatan_log` (text)
* `waktu_update` (timestamp)

### 4. `prokers` (Master Program Kerja)
* `id` (UUID, Primary Key)
* `nama_proker` (varchar(255), Unique, Not Null)
* `deskripsi` (text)

---

## 📡 REST API Reference

Semua response sukses mengembalikan JSON standar HTTP status code `200` atau `201`. Error mengembalikan format `{ "error": "keterangan error" }` dengan status code yang sesuai (`400`, `401`, `403`, `404`, `500`).

### 1. Authentication (`/auth`)
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `POST` | `/login` | Public | Login via Username/NIM & Password. Mengembalikan JWT token. |
| `GET` | `/auth/me` | Bearer | Mendapatkan identitas dan hak akses user yang sedang login. |
| `PUT` | `/profile/change-username` | Bearer | Mengubah username pengguna yang login (min. 3 karakter, tanpa spasi). |
| `PUT` | `/auth/change-password` | Bearer | Memperbarui password akun (min. 6 karakter). |
| `PUT` | `/auth/dismiss-password-prompt` | Bearer | Menutup prompt wajib ganti password awal. |

### 2. Manajemen Surat (`/letters`)
| Method | Endpoint | Role | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/letters` | Bearer | Mengambil seluruh daftar surat beserta relasi log tracking. |
| `POST` | `/letters` | Bearer | Menambahkan surat baru. |
| `PUT` | `/letters/:id` | Bearer | Memperbarui atribut surat atau memindahkan status tahapan. |
| `DELETE` | `/letters/:id` | `sekre` | Menghapus arsip surat secara permanen. |

### 3. Manajemen Program Kerja (`/programs`)
| Method | Endpoint | Role | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/programs` | Bearer | Mengambil seluruh master proker terurut alfabetis. |
| `POST` | `/programs` | `sekre` | Menambahkan master program kerja baru. |
| `PUT` | `/programs/:id` | `sekre` | Mengubah nama atau deskripsi proker. |
| `DELETE` | `/programs/:id` | `sekre` | Menghapus proker. |

### 4. Manajemen Pengurus (`/users`)
| Method | Endpoint | Role | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/users` | `sekre` | Mengambil seluruh data pengurus DOSCOM. |
| `POST` | `/users` | `sekre` | Menambah pengurus manual (validasi unik NIM, Nama, dan Username). |
| `PUT` | `/users/:id` | `sekre` | Memperbarui biodata atau mereset password pengurus. |
| `DELETE` | `/users/:id` | `sekre` | Menghapus akun pengurus. |
| `POST` | `/users/import` | `sekre` | Import massal pengurus via file CSV. |

---

## 🏃 Local Setup & Development

### 1. Prasyarat
* Go `1.22+` atau `1.26`
* PostgreSQL `15+`

### 2. Konfigurasi Environment (`backend/.env`)
Salin file `.env.example`:
```env
PORT=8080
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=root
DB_PASSWORD=password
DB_NAME=laci_doscom
DB_SSLMODE=disable
JWT_SECRET=supersecretjwtkeyforlaci
APP_URL=http://localhost:5173

# Konfigurasi SMTP (Opsional untuk pengiriman kredensial)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
```

### 3. Eksekusi
```bash
# Download dependencies
go mod tidy

# Jalankan backend service
go run .

# Menjalankan Database Seeder inti kebutuhan real (Users & Proker)
go run cmd/seed/main.go

# Atau jika menjalankan via Docker Compose:
docker compose exec backend ./laci-seed
```
