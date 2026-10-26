# ⚙️ Laci Backend Service

REST API Service untuk **Laci DOSCOM** dibangun menggunakan bahasa pemrograman **Golang** dengan framework **Gin**, ORM **GORM**, dan database **PostgreSQL 15**.

---

## 🛠️ Tech Stack & Dependencies

* **Language:** Golang `1.22+`
* **HTTP Framework:** [Gin Gonic](https://github.com/gin-gonic/gin)
* **ORM & Database:** [GORM](https://gorm.io/) (PostgreSQL Driver)
* **Auth & Security:** JWT (`golang-jwt/jwt/v5`), bcrypt (`golang.org/x/crypto`)
* **Utilities:** `google/uuid`, `godotenv`
* **Mailer:** Standard library `net/smtp` dengan TLS / STARTTLS

---

## 🗄️ Database Schema & Relational Design

Sistem mengandalkan database relasional PostgreSQL dengan auto-migration dan migrasi idempoten:

```
┌──────────────┐       1:N       ┌──────────────────┐
│    surat     ├────────────────►│   log_tracking   │
│  (ID: UUID)  │                 │    (ID: UUID)    │
└──────┬───────┘                 └──────────────────┘
       │ N:1
       ▼
┌──────────────┐
│    proker    │
│  (ID: UUID)  │
└──────────────┘

┌──────────────┐                 ┌──────────────────┐
│    users     │                 │ surat_templates  │
│  (ID: UUID)  │                 │    (ID: UUID)    │
└──────────────┘                 └──────────────────┘

┌──────────────┐
│ app_settings │ (Key-Value: gas_webhook_url, default_output_folder_url)
└──────────────┘
```

### 1. `users` (Tabel Pengurus)
* `id` (UUID, Primary Key)
* `nim` (varchar(50), Indexed) — Nomor Induk Mahasiswa pengurus.
* `nama` (varchar(150), Not Null)
* `email` (varchar(150), Indexed)
* `username` (varchar(100), Unique Index, Not Null) — Auto-sanitized atau kustom.
* `password` (varchar(255), Not Null) — Bcrypt hash string.
* `divisi` (varchar(100), Default: 'humas') — Role: `ketum`, `sekre`, `humas`, `pengurus`, `ex-ketum`.
* `is_active` (boolean, Default: false) — Menandai status aktivasi akun via email.
* `activation_token` (varchar(255), Indexed)
* `activation_token_expires_at` (timestamp)
* `reset_password_token` (varchar(255), Indexed)
* `reset_password_expires_at` (timestamp)

### 2. `surat` (Tabel Arsip Surat)
* `id` (UUID, Primary Key)
* `nomor_surat` (varchar(255), Not Null) — Format: `[No]/DOSCOM/[Proker]/[Romawi]/[Tahun]`.
* `perihal` (varchar(255), Not Null)
* `tujuan` (varchar(255))
* `status_saat_ini` (varchar(100), Not Null) — 8 tahapan resmi alur Kanban.
* `pic_nama` (varchar(100), Not Null)
* `catatan` (text)
* `arsip_url` (text) — Path file fisik lokal atau URL PDF Google Drive.
* `g_docs_url` (text) — URL file Google Docs yang dapat diedit di Drive.
* `proker_id` (UUID, Foreign Key opsional ke `proker.id`)

### 3. `log_tracking` (Riwayat Perubahan Status Surat)
* `id` (UUID, Primary Key)
* `surat_id` (UUID, Indexed)
* `status_baru` (varchar(100), Not Null)
* `catatan_log` (text)
* `waktu_update` (timestamp, Not Null)

### 4. `proker` (Master Program Kerja)
* `id` (UUID, Primary Key)
* `nama_proker` (varchar(255), Unique Index, Not Null)
* `deskripsi` (text)

### 5. `surat_templates` (Master Template Google Docs)
* `id` (UUID, Primary Key)
* `nama_template` (varchar(255), Not Null)
* `g_doc_id` (varchar(255), Unique Index, Not Null)
* `placeholders` (text) — Disimpan sebagai JSON array string `["nama_kegiatan", "tanggal"]`.
* `is_active` (boolean, Default: true)

### 6. `app_settings` (Konfigurasi Aplikasi)
* `key` (varchar(100), Primary Key)
* `value` (text, Not Null)

---

## 📡 REST API Reference

Semua response sukses mengembalikan JSON standar HTTP status code `200` atau `201`. Error mengembalikan format `{ "error": "keterangan error" }` dengan status code yang sesuai (`400`, `401`, `403`, `404`, `500`).

### 1. Autentikasi & Profil Pengguna
| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :---: | :--- |
| `POST` | `/login` | Public | Login via Username/NIM/Email & Password. Mengembalikan JWT. |
| `GET` | `/auth/verify-token` | Public | Verifikasi keabsahan token aktivasi akun atau token reset kata sandi. |
| `POST` | `/auth/activate` | Public | Aktivasi akun perdana dan penetapan kata sandi mandiri. |
| `POST` | `/auth/forgot-password` | Public | Permintaan tautan pemulihan kata sandi via email. |
| `POST` | `/auth/reset-password` | Public | Penetapan kata sandi baru menggunakan token reset. |
| `GET` | `/profile` | Bearer | Mengambil data profil akun pengguna yang sedang login. |
| `PUT` | `/profile/change-username` | Bearer | Mengubah username sendiri (min. 3 karakter, tanpa spasi). |
| `PUT` | `/profile/change-password` | Bearer | Mengubah password akun sendiri dengan verifikasi password lama. |

### 2. Manajemen Pengurus (`/users`)
| Method | Endpoint | Hak Akses | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/users` | Bearer | Mengambil daftar seluruh anggota/pengurus. |
| `POST` | `/users` | `sekre` / `ketum` | Mendaftarkan akun anggota baru & memicu email aktivasi. |
| `PUT` | `/users/:id` | `sekre` / `ketum` | Mengubah informasi anggota atau divisi. |
| `DELETE` | `/users/:id` | `sekre` / `ketum` | Menghapus akun anggota tertentu. |
| `DELETE` | `/users/batch` | `sekre` / `ketum` | Menghapus banyak akun anggota sekaligus. |
| `POST` | `/users/import-csv` | `sekre` / `ketum` | Impor data anggota massal via berkas CSV. |
| `POST` | `/users/import-spreadsheet`| `sekre` / `ketum` | Impor langsung dari tautan publik Google Sheets. |
| `POST` | `/users/transfer-ketum` | `ketum` | Serah Terima Jabatan resmi Ketua Umum. |

### 3. Persuratan & Kanban (`/letters`)
| Method | Endpoint | Hak Akses | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/letters` | Bearer | Mengambil seluruh daftar surat beserta relasi log tracking. |
| `POST` | `/letters` | Bearer | Menambahkan surat baru ke papan alur (status awal: Standby). |
| `PUT` | `/letters/:id` | Bearer | Mengubah status kolom alur atau memperbarui catatan instruksi. |
| `POST` | `/letters/:id/upload` | Bearer | Mengunggah foto bukti fisik arsip (hanya saat status Selesai untuk non-sekre). |
| `DELETE` | `/letters/:id` | `sekre` / `ketum` | Menghapus arsip surat secara permanen. |

### 4. Generator Surat & Template Google Docs
| Method | Endpoint | Hak Akses | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/letters/templates` | Bearer | Mengambil daftar template surat aktif. |
| `POST` | `/letters/templates/sync` | `sekre` / `ketum` | Sinkronisasi folder master Google Drive via Google Apps Script. |
| `POST` | `/letters/generate` | Bearer | Generate dokumen Google Docs & PDF dari template surat. |
| `GET` | `/settings` | `sekre` / `ketum` | Mengambil konfigurasi URL webhook GAS & default folder output. |
| `PUT` | `/settings` | `sekre` / `ketum` | Memperbarui konfigurasi URL webhook GAS & folder output. |

### 5. Master Program Kerja (`/programs`)
| Method | Endpoint | Hak Akses | Deskripsi |
| :--- | :--- | :---: | :--- |
| `GET` | `/programs` | Public / Bearer | Mengambil seluruh master program kerja terurut alfabetis. |
| `POST` | `/programs` | `sekre` / `ketum` | Menambahkan program kerja baru. |
| `PUT` | `/programs/:id` | `sekre` / `ketum` | Memperbarui nama atau deskripsi program kerja. |
| `DELETE` | `/programs/:id` | `sekre` / `ketum` | Menghapus program kerja. |
