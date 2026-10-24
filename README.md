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
   Buka browser kamu dan akses 👉 **`http://localhost`** (atau `http://localhost:5173` jika menggunakan `docker-compose.dev.yml`).

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

## 🌱 Menjalankan Seeder di Docker

Setelah container berjalan (`docker compose up -d`), jalankan seeder langsung ke dalam container backend:

```bash
docker compose exec backend ./laci-seed
```


---

## 🤝 Kontribusi

Tertarik menambahkan fitur atau memperbaiki bug?

1. Fork repositori ini.
2. Buat branch fitur baru (`git checkout -b fitur-keren-kamu`).
3. Pastikan `npm run build` di frontend dan `go vet ./...` di backend lulus tanpa error.
4. Buat Pull Request dengan penjelasan singkat yang jelas.

Salam Open Source! 🐧 **DOSCOM — Memasyarakatkan Open Source dan Meng-Open Source-kan Masyarakat.**
