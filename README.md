# 🗄️ Laci DOSCOM

Halo Sobat DOSCOM! 👋 Selamat datang di repositori resmi **Laci DOSCOM** (Layanan Arsip & Cek Informasi Surat).
Repositori ini terbuka untuk seluruh pengurus dan anggota yang ingin berkontribusi dalam digitalisasi sistem persuratan organisasi kita.

> 📚 **Dokumentasi Lengkap:**
> - Ingin panduan arsitektur & konteks untuk AI coding assistant? 👉 Baca [**CONTEXT.md**](CONTEXT.md)
> - Ingin membaca spesifikasi produk & matriks otoritas lengkap? 👉 Baca [**Master PRD v2**](docs/PRD.md)
> - Panduan format template Google Docs untuk Sekretaris? 👉 Baca [**Panduan Template Surat Sekre**](docs/PANDUAN_SEKRE_TEMPLATE_SURAT.md)
> - Panduan teknis arsitektur backend Go & API? 👉 Baca [**Backend README**](backend/README.md)
> - Panduan teknis frontend Vue 3 & styling? 👉 Baca [**Frontend README**](frontend/README.md)

---

## 🤔 Laci Tuh Apa Sih?

Secara harfiah, **Laci** adalah tempat kita menyimpan dan mengorganisir dokumen. Di DOSCOM, **Laci** adalah aplikasi web internal untuk membantu Sekretaris dan seluruh pengurus mengelola, men-generate, serta memantau status alur surat-menyurat dan program kerja organisasi tanpa ribet.

Di Laci v2, kita bisa:

- 📋 **Pantau Status Surat (Kanban Board):** Papan Kanban 8 tahapan resmi (dari `Standby`, `Cetak`, rangkaian tanda tangan, hingga `Selesai`) dengan pencatatan audit log otomatis.
- ⚡ **Generator Surat Otomatis (Google Docs & GAS):** Sekretaris cukup menaruh template di Google Drive; sistem otomatis mendeteksi variabel `{{placeholder}}` dan meng-generate dokumen Google Docs serta PDF resmi secara instan tanpa biaya.
- 🔒 **Zero-Knowledge Auth v2 & Hierarki Tegas:** Tanpa password bawaan (*no hardcoded password*). Anggota mengaktivasi akun mandiri via tautan email resmi. Hak akses terbagi rapi antara **Ketua Umum (Superadmin)**, **Sekretaris (Admin Operasional)**, dan **Humas / Pengurus (Operator Alur)**.
- 🔄 **Suksesi Kepemimpinan Sah:** Serah Terima Jabatan Ketua Umum terintegrasi dengan validasi keamanan kata sandi aktif.
- 👥 **Impor Anggota Massal:** Impor data pengurus dalam hitungan detik via file CSV atau link Google Spreadsheet publik.
- ⚡ **Super Ringan & Responsif:** Desain modern mobile-first yang nyaman dibuka langsung dari browser smartphone pengurus.

---

## 🛠️ Senjata Kita (Tech Stack)

- **Frontend:** Vue 3 (Composition API) + Tailwind CSS v4 + Vite + Radix-Vue (shadcn-vue style)
- **Backend:** Golang 1.22+ (Gin Gonic + GORM)
- **Database:** PostgreSQL 15 Alpine
- **Automation Bridge:** Google Apps Script Webhook (Serverless GDocs & GDrive Bridge)
- **DevOps:** Docker & Docker Compose (Single root `.env`)

---

## 🚀 Cara Menjalankan Proyek (Quick Start)

Paling mudah menjalankan proyek ini adalah menggunakan **Docker Compose**, tanpa perlu instalasi Go atau PostgreSQL manual di komputer:

### 1. Clone Repositori & Siapkan Konfigurasi

```bash
git clone https://github.com/Chqrety/laci.git
cd laci
```

Buat berkas `.env` di root repositori dengan menyalin template yang disediakan:

```bash
cp .env.example .env
```

Sesuaikan konfigurasi SMTP email (Gmail App Password) dan data awal Ketua Umum (`INIT_KETUM_*`) jika diperlukan.

---

### 2. Mode Development (Hot-Reload Otomatis)

Gunakan mode ini saat aktif mengembangkan fitur. Frontend (Vite HMR) dan Backend (Golang Air) otomatis melakukan live-reload setiap kali kode disimpan:

```bash
docker compose -f docker-compose.dev.yml up --build
```

- **Aplikasi Web (Dev):** `http://localhost:5173`
- **Backend API (Dev):** `http://localhost:8080`
- **PostgreSQL Database:** `localhost:5432` (User: `root`, DB: `laci_dev_db`)

> [!TIP]
> Saat pertama kali dijalankan, sistem otomatis menjalankan migrasi skema dan seeder inti (akun perdana Ketua Umum & master Proker default). Tautan aktivasi akun Ketua Umum akan dicetak langsung di log terminal backend!

---

### 3. Mode Production (Build & Deploy)

Gunakan mode ini untuk simulasi production atau deployment ke server VPS (Frontend Nginx statis & Backend binary Alpine ringan):

```bash
docker compose up -d --build
```

- **Aplikasi Web (Prod):** `http://localhost` (Port 80)
- **Backend API (Internal):** Port 8080 di-reverse proxy oleh Nginx

---

### 4. Opsi Alternatif: Local Dev Native (Tanpa Docker)

Jika ingin menjalankan service secara native di mesin lokal:

**A. Jalankan Backend (Golang):**
```bash
cd backend
go run .
# Backend berjalan di http://localhost:8080
```

**B. Jalankan Frontend (Vue 3):**
```bash
cd frontend
npm install
npm run dev
# Frontend berjalan di http://localhost:5173
```

---

## 👥 Kontribusi

Tertarik menambahkan fitur baru atau merapikan tampilan?
1. Fork repositori ini.
2. Buat branch fitur baru (`git checkout -b feat/nama-fitur`).
3. Pastikan kode mengikuti prinsip **Ponytail Protocol** (minimalis, tanpa bloat, gunakan library standar/eksis).
4. Commit dan buat Pull Request ke branch `dev`.
