# Project Context & AI Agent Orientation: LACI DOSCOM

> **Note for AI Coding Assistants & LLMs:**
> Read this file first to understand the architecture, domain logic, constraints, and file organization of this repository before making changes.

---

## 1. Project Overview & Core Philosophy

**LACI (Layanan Arsip & Cek Informasi Surat)** adalah platform internal berbasis web milik **DOSCOM (Dinus Open Source Community)** yang mendigitalisasi pembuatan surat organisasi, integrasi template Google Docs, pelacakan alur persetujuan fisik (Kanban Board), dan manajemen anggota organisasi.

### Core Principles (Ponytail Protocol)
* **The Ladder:** YAGNI > Existing utils/patterns > Standard library first > Native platform feature > Installed dependencies > Minimal code.
* **Zero Bloat:** Hindari abstraksi spekulatif, factory kosong, wrapper delegasi satu baris, atau dependensi berlebih jika pustaka standar mencukupi.
* **Shortest Working Diff:** Deletion over addition, boring over clever, root-cause guards over wide cascading refactors.
* **Anti-Slop UI (Floema Aesthetic):** Antarmuka editorial bersih, tipografi presisi, hindari card-inside-card soup, responsive mobile-first, dan transisi halus.

---

## 2. Tech Stack Architecture

* **Backend:** Golang 1.22+ (Pragmatic Lean Layered: Handlers $\rightarrow$ Services $\rightarrow$ Repositories, `Gin Gonic`, `GORM`, `PostgreSQL 15`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`).
* **Frontend:** Vue 3 Composition API (`<script setup lang="ts">`), `Tailwind CSS v4`, `Radix-Vue` (shadcn-vue style), `Axios`, `Vue Router`.
* **Automation Bridge:** Serverless Google Apps Script (`gas/`) sebagai jembatan nir-biaya (*zero-cost webhook*) ke Google Drive & Google Docs API.
* **DevOps:** Docker & Docker Compose (Dual target: `docker-compose.dev.yml` untuk Air hot-reload & Vite HMR, `docker-compose.yml` untuk Production).

---

## 3. Directory Layout & Key Locations

```
laci/
├── .env.example             # Template konfigurasi environment tunggal di root
├── .env.dev.example         # Template konfigurasi environment khusus dev docker
├── docker-compose.dev.yml   # Docker compose development (Air + Vite HMR)
├── docker-compose.yml       # Docker compose production (Alpine binary + Nginx)
├── README.md                # Dokumentasi publik proyek
├── CONTEXT.md               # File panduan konteks agen AI (dokumen ini)
├── docs/                    # Dokumentasi produk & teknis
│   ├── PRD.md               # Master PRD v2 (Auth, Kanban, Generator, Role Matrix)
│   ├── revisi-2-arsitektur-backend.md # Blueprint arsitektur berlapis backend
│   └── PANDUAN_SEKRE_TEMPLATE_SURAT.md # Panduan format Google Docs untuk Sekretaris
├── gas/                     # Google Apps Script code (Web App Webhook)
│   └── Code.gs              # Logic duplicate, replace {{placeholders}}, export PDF
├── backend/
│   ├── config/              # database.go (GORM connection & idempoten migrations)
│   ├── repositories/        # Data Access Layer (GORM): user, surat, proker, template
│   ├── services/            # Business Logic Layer (Clean Go): auth, user, surat, proker, template
│   ├── handlers/            # Transport/Delivery Layer (Gin HTTP): auth, user, surat, proker, template
│   ├── models/              # models.go (User, Surat, LogTracking, Proker, SuratTemplate, AppSetting)
│   ├── routes/              # setup.go (Route registration, dependency wiring, static asset hosting)
│   ├── seeders/             # seeder.go (SeedCore: Genesis Ketum & Default Proker)
│   ├── utils/               # email.go (SMTP notification cards), token.go (crypto hex token)
│   ├── main.go              # Entrypoint backend, port binding, auto-seeder boot
│   └── Dockerfile.dev       # Hot-reload container menggunakan cosmtrek/air
└── frontend/
    ├── src/
    │   ├── views/           # KanbanBoard, SuratManagement, UserManagement, Login, Profile, etc.
    │   ├── components/
    │   │   ├── molecules/   # CardDetail (detail modal), KanbanFilter
    │   │   ├── organisms/   # KanbanCard, KanbanModalForm, SuratGeneratorModal, TemplateSetupModal
    │   │   └── ui/          # Primitif shadcn-vue (button, input, badge, dialog, select, toast)
    │   ├── services/
    │   │   └── api.ts       # Axios instance, interceptors, authService, suratService, prokerService
    │   ├── composables/     # useKanbanDragDrop, useDocumentScanner
    │   └── router/          # index.ts (Role guard sekre/ketum, public auth routes)
    └── Dockerfile.dev       # Vite container development
```

---

## 4. Role & Authority Model

Role dan divisi disatukan pada atribut `divisi` di tabel `users`:

1. **`ketum` (Ketua Umum - Superadmin / God Mode):**
   - Mengelola semua modul: Surat, Template, Proker, dan Manajemen Pengurus.
   - Satu-satunya role yang berwenang mengangkat Ketua Umum baru melalui fitur **Serah Terima Jabatan**.
2. **`sekre` (Sekretaris - Administrator Operasional):**
   - Mengelola Surat, Master Template GDocs, Proker, dan Pengurus (Humas, Pengurus, sesama Sekre).
   - **Dilarang** membuat atau mempromosikan akun menjadi `ketum`.
3. **`humas` & `pengurus` (Operator Alur):**
   - Melihat dan memindahkan alur kartu surat di Kanban Board.
   - Menambahkan catatan instruksi di kartu surat.
   - **Proteksi Berkas:** Hanya diizinkan mengunggah foto arsip fisik ketika status surat sudah mencapai `Selesai`.
4. **`ex-ketum` (Demisioner Ketua Umum):**
   - Memiliki hak akses sama dengan pengurus biasa.
   - Didapatkan murni secara otomatis saat Ketua Umum aktif melakukan serah terima jabatan.

---

## 5. Core Business Lifecycles

### A. Auth v2 & Genesis Ketum
* Tidak ada password bawaan / hardcoded password.
* **Genesis Ketum:** Backend membaca konfigurasi `.env` (`INIT_KETUM_*`) saat boot pertama. Jika belum ada akun `ketum`, dibuat akun non-aktif dan tautan aktivasi mandiri dikirim via email / dicetak ke log terminal.
* **Self-Activation:** Pengguna baru menerima email aktivasi dengan token kriptografis 24 jam (`/activate?token=...`) untuk membuat password rahasia sendiri.
* **Password Reset:** Alur pemulihan akun 1 jam dengan respons generik untuk mencegah kebocoran data pengguna (*anti-enumeration*).

### B. Kanban Persuratan (8 Tahapan Resmi)
* Urutan tahapan: `Standby` → `Cetak` → `TTD Dasar` → `TTD Ketum` → `TTD Pembina` → `Paraf Koormawa` → `TTD Tertinggi` → `Selesai`.
* Setiap perubahan kolom secara otomatis mencatat riwayat ke tabel `log_trackings`.

### C. Generator Surat (Google Drive & Google Apps Script)
* Format placeholder di Google Docs: `{{nama_variabel}}`.
* **Sync Template:** Memindai folder Google Drive via GAS untuk membaca seluruh template dan mendeteksi placeholder secara dinamis.
* **Generate:** GAS menyalin template, mengganti placeholder dengan data form input, menyimpan ke folder output, dan mengembalikan URL Google Docs serta URL PDF siap unduh.

---

## 6. Environment & Running Workflows

Semua konfigurasi terpusat di satu berkas `.env` pada root repositori:

```bash
# Menjalankan ekosistem development (Hot-reload Go Air & Vite HMR)
docker compose -f docker-compose.dev.yml up --build

# Menjalankan production (Zero-Setup / Tanpa Konfigurasi Wajib)
docker compose up -d --build
```

* **Web UI (Dev):** `http://localhost:5173`
* **Backend API (Dev):** `http://localhost:8080`
* **PostgreSQL (Dev):** `localhost:5432`
* **Production Web:** Port 80 (Buka via IP Server atau Domain apa saja)
