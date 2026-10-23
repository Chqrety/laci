# 🗄️ Laci DOSCOM

Halo Sobat DOSCOM! 👋 Selamat datang di repositori resmi **Laci DOSCOM**.
Repositori ini terbuka untuk seluruh pengurus dan anggota yang ingin berkontribusi dalam digitalisasi sistem persuratan organisasi kita.

> 📚 **Dokumentasi Lanjutan:**
>
> - Ingin review arsitektur teknis lengkap untuk PM / Tech Lead? 👉 Baca [**Technical Review PM & Tech Lead**](docs/PM_TECHLEAD_REVIEW.md)
> - Tertarik oprek arsitektur backend Go & API? 👉 Baca [**Backend README**](backend/README.md)
> - Tertarik eksplorasi frontend Vue 3 & styling? 👉 Baca [**Frontend README**](frontend/README.md)

---

## 🤔 Laci Tuh Apa Sih?

Secara harfiah, **Laci** adalah tempat kita menyimpan dan mengorganisir dokumen. Di DOSCOM, **Laci** adalah aplikasi web internal untuk membantu Sekretaris dan pengurus mengelola serta memantau status alur surat-menyurat dan program kerja organisasi tanpa ribet.

Di sini kita bisa:

- 📋 **Pantau Status Surat:** Kanban Board responsif untuk melihat apakah surat masih berstatus Draft, Review, Signed, atau Sent.
- 📷 **Scan Dokumen Kilat:** Ambil foto surat fisik via kamera HP, dan sistem akan membaca nomor surat, perihal, serta proker secara instan (< 1 detik).
- 🔒 **Role-Based Access:** Akses aman terpisah antara pengurus **Sekretaris** dan **Humas**.
- ⚡ **Super Ringan di HP:** Didesain seringan mungkin tanpa lag, bahkan di browser smartphone hemat daya sekalipun.

---

## 🛠️ Senjata Kita (Tech Stack)

- **Frontend:** Vue 3 (Composition API) + Tailwind CSS v4 + Vite + Radix-Vue
- **Backend:** Golang (Gin Gonic + GORM)
- **Database:** PostgreSQL 15 Alpine
- **DevOps:** Docker & Docker Compose

---

## 🚀 Cara Gas Lokal (Quick Start via Docker)

Paling mudah menjalankan proyek ini adalah menggunakan **Docker Compose**, tanpa perlu instalasi Go atau PostgreSQL manual:

1. **Clone repositori:**

   ```bash
   git clone https://github.com/doscom/laci.git
   cd laci
   ```

2. **Jalankan container stack:**

   ```bash
   cp .env.example .env
   docker compose up -d --build
   ```

3. **Buka di Browser:**
   Buka browser kamu dan akses 👉 **`http://localhost:5173`**

> 💡 **Kenapa Port 5173?**
> Secara umum web server production memakai port `80`. Namun karena banyak pengurus DOSCOM yang menjalankan **XAMPP** lokal (yang memakai port 80 untuk Apache), Laci di-mapping ke port `5173` agar bebas bentrok dan langsung terasa seperti environment development modern!

---

## 💻 Panduan Ngoding Mandiri (Local Dev Mode)

Jika kamu ingin mengembangkan fitur dengan _Hot-Reload_:

1. **Nyalakan Database PostgreSQL saja:**

   ```bash
   docker compose up db -d
   ```

2. **Jalankan Backend (Golang):**

   ```bash
   cd backend
   go run .
   # Backend berjalan di http://localhost:8080
   ```

3. **Jalankan Frontend (Vue 3):**
   ```bash
   cd frontend
   cp .env.example .env
   npm ci
   npm run dev
   # Frontend berjalan di http://localhost:5173
   ```

---

## 🧪 Inject Dummy Data (Buat Testing)

Aplikasi kosong tentu sepi untuk ditest. Kamu bisa memasukkan **50 data surat simulasi** secara instan ke database lokal:

1. **Masuk ke container database:**

   ```bash
   docker compose exec db psql -U root -d laci_doscom
   ```

2. **Jalankan query SQL berikut:**

   ```sql
   TRUNCATE TABLE log_trackings, surat CASCADE;

   INSERT INTO surat (id, nomor_surat, perihal, pic_nama, status_saat_ini)
   SELECT
       gen_random_uuid(),
       LPAD(i::text, 3, '0') || '/DOSCOM/DU/X/2026',
       'Surat Permohonan Kegiatan Ke-' || i,
       'Pengurus DOSCOM ' || i,
       CASE (i % 4)
           WHEN 0 THEN 'draft'
           WHEN 1 THEN 'review'
           WHEN 2 THEN 'signed'
           ELSE 'sent'
       END
   FROM generate_series(1, 50) AS i;
   ```

---

## 🤝 Kontribusi

Tertarik menambahkan fitur atau memperbaiki bug?

1. Fork repositori ini.
2. Buat branch fitur baru (`git checkout -b fitu-keren-kamu`).
3. Pastikan `npm run build` di frontend dan `go vet ./...` di backend lulus tanpa error.
4. Buat Pull Request dengan penjelasan singkat yang jelas.

Salam Open Source! 🐧 **DOSCOM — Memasyarakatkan Open Source dan Meng-Open Source-kan Masyarakat.**
