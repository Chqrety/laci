# 📋 Technical Specification & System Architecture Review
### Proyek: Laci DOSCOM (Sistem Manajemen & Pelacakan Surat)
**Dokumen untuk:** Project Manager (PM) & Tech Lead (TL)  
**Status:** Ready for Review (v1.0)  
**Organisasi:** Dinus Open Source Community (DOSCOM)

---

## 1. Executive Summary & Problem Framing

### 1.1 Latar Belakang & Kebutuhan Bisnis
Dalam operasional organisasi mahasiswa seperti DOSCOM, persuratan resmi (surat permohonan, peminjaman ruangan, delegasi, dan undangan kegiatan) sering mengalami *bottleneck*:
* Keberadaan surat fisik/digital tidak terpantau posisinya (apakah masih di Sekretaris, menunggu persetujuan Ketua, atau sudah dikirim ke Dosen/Birokrat).
* Penomoran surat manual rentan dobel antar pengurus/divisi.
* Belum adanya sistem terpusat yang ringan untuk monitoring surat dari smartphone saat pengurus sedang berada di kampus.

### 1.2 Solusi Produk: Laci DOSCOM
Laci dirancang secara harfiah sebagai "laci arsip digital" yang gesit:
1. **Visual Kanban Tracking:** Memantau alur tahapan surat (`Draft` $\to$ `Review` $\to$ `Signed` $\to$ `Sent`) secara *real-time*.
2. **On-Device OCR Auto-Fill:** Ekstraksi nomor surat, perihal, dan proker otomatis via kamera smartphone dalam < 1 detik.
3. **Role-Based Access Control:** Membatasi wewenang Sekretaris (`sekre`) sebagai pengelola master arsip/proker dan Humas (`humas`) sebagai pelaksana lapangan.
4. **Mobile-First High Performance:** Didesain seringan kilat untuk perangkat *low-end* tanpa framework validasi berat.

---

## 2. Arsitektur Teknis Sistem

```
[ Browser / Mobile Client ]
            │
            ▼ (Port 5173 / Port 80 Nginx)
┌────────────────────────────────────────────────────────┐
│  Frontend Client (Vue 3 + Vite + Tailwind CSS v4)      │
│  - Route-level Async Chunk Splitting                   │
│  - On-Demand Tesseract.js WASM Web Worker              │
│  - Zero-Bloat Form (Native Vue v-model + Central Toast)│
└───────────────────────────┬────────────────────────────┘
                            │ /api/* (Reverse Proxy)
                            ▼
┌────────────────────────────────────────────────────────┐
│  Backend Service (Golang / Gin Engine)                 │
│  - JWT Bearer Authentication Context                   │
│  - Role Guard Middleware (Sekre vs Humas)              │
│  - Database Auto-Migration via GORM                    │
└───────────────────────────┬────────────────────────────┘
                            │ TCP Connection :5432
                            ▼
┌────────────────────────────────────────────────────────┐
│  Database (PostgreSQL 15 Alpine)                       │
│  - Tables: users, surat, log_trackings, prokers        │
│  - Relational Integrity & Case-Insensitive Uniqueness  │
└────────────────────────────────────────────────────────┘
```

---

## 3. Matriks Hak Akses & Keamanan (RBAC)

| Modul / Endpoint | Endpoint Path | Sekretaris (`sekre`) | Humas (`humas`) | Catatan Keamanan |
| :--- | :--- | :---: | :---: | :--- |
| **Login & Auth** | `/login`, `/auth/me` | ✅ | ✅ | Token JWT berlaku dengan context claims role. |
| **Kanban Board** | `/letters` (GET, PUT) | ✅ | ✅ | Humas dan Sekre dapat memperbarui status surat. |
| **Arsip Surat (Hapus)** | `/letters/:id` (DELETE) | ✅ | ❌ (403) | Penghapusan arsip hanya dapat dilakukan Sekre. |
| **Kelola Proker** | `/programs` (CRUD) | ✅ | ❌ (403) | Master kode penomoran surat dikunci untuk Sekre. |
| **Kelola Pengurus** | `/users` (CRUD, Import) | ✅ | ❌ (403) | CRUD user, reset password, dan validasi duplikat. |
| **Ganti Username/Sandi** | `/profile/*` | ✅ | ✅ | Setiap user mengelola kredensial mandiri. |

---

## 4. Evaluasi Keputusan Arsitektur & Kinerja (Tech Lead Highlights)

### 4.1 Pemangkasan Zod & Vee-Validate (Zero-Bloat Architecture)
* **Temuan Awal:** Library `zod`, `vee-validate`, dan `@vee-validate/zod` sebelumnya hanya dipakai pada 1 field username di `Profile.vue`, namun menambah bobot bundle hingga ~100 kB dan 7 file wrapper komponen.
* **Keputusan:** Dihapus total dari dependensi. Form dikembalikan ke native Vue `v-model` + validasi standar HTML5 (`required`, `minlength`).
* **Hasil:** Ukuran chunk profil terpangkas **93%** (dari 103 kB menjadi 7.1 kB). Beban komputasi parsing JS di CPU mobile hilang.

### 4.2 Optimasi OCR Dokumen di Smartphone
* **Tantangan:** Foto resolusi tinggi kamera HP (12–48 MP) jika langsung diproses oleh Tesseract WASM membutuhkan waktu 30–45 detik dan menghabiskan RAM hingga browser *crash*.
* **Solusi:**
  1. *Canvas Pre-processing:* Foto di-downscale ke dimensi proporsional maksimal 1500px dan dikonversi ke grayscale kontras tinggi.
  2. *Lazy Dynamic Import:* `tesseract.js` diimpor dinamis saat tombol scan ditekan.
  3. *Lifecycle Termination:* Worker langsung di-terminate (`worker.terminate()`) segera setelah teks terbaca.
* **Hasil:** Waktu pembacaan teks turun menjadi **< 1 detik**, RAM browser HP langsung dibebaskan.

### 4.3 Hirarki Stacking Context & Teleport
* Mencegah bug visual dropdown tertutup dialog di Radix-Vue:
  * Modal Dialog: `z-[100]`
  * Toaster Notification: `z-[110]`
  * Select Content: `z-[120]`
  * Mobile Detail Drawer: Ditranslasikan keluar via `<Teleport to="body">` dengan *body scroll lock* otomatis.

---

## 5. Skema Relasi Database (PostgreSQL 15)

```
[users] 
  ├── id (UUID, PK)
  ├── nim (varchar(50), indexed)
  ├── nama (varchar(255))
  ├── username (varchar(100), unique)
  ├── password (bcrypt hash)
  └── divisi (varchar(100): 'sekre' | 'humas')

[surat] 
  ├── id (UUID, PK)
  ├── nomor_surat (varchar(255), unique)
  ├── perihal (text)
  ├── status_saat_ini (varchar(100))
  ├── pic_nama (varchar(255))
  └── file_surat (varchar(255))
        │
        └───► 1:N ───► [log_trackings]
                        ├── id (UUID, PK)
                        ├── surat_id (UUID, FK -> surat.id CASCADE)
                        ├── status_sebelumnya (varchar(100))
                        ├── status_baru (varchar(100))
                        └── created_at (timestamp)

[prokers]
  ├── id (UUID, PK)
  ├── nama_proker (varchar(255), unique)
  └── deskripsi (text)
```

---

## 6. Project Roadmap & Next Action Items

* [x] **Sprint 1 — Core Architecture:** Setup Gin + GORM, PostgreSQL, Vue 3, Tailwind v4, Docker Compose.
* [x] **Sprint 2 — Kanban & Workflow:** Board alur surat, drawer detail tracking, perbaikan stacking dialog & select.
* [x] **Sprint 3 — OCR & Mobile Performance:** Canvas pre-processing scanner, lazy routes, pencabutan Zod/Vee-Validate bloat.
* [x] **Sprint 4 — User Management & RBAC:** Restriksi divisi (`sekre`/`humas`), validasi dobel NIM/Nama di database, rotasi password.
* [ ] **Sprint 5 (Next) — Template Engine & Otomasi Surat:** Generator PDF surat resmi DOSCOM langsung dari form input web.
* [ ] **Sprint 6 (Next) — Public Verification:** Halaman publik verifikasi keaslian surat via QR code / tracking ID.
