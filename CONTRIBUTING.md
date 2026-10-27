# Panduan Kontribusi - LACI DOSCOM

Terima kasih atas minat Anda untuk berkontribusi pada **LACI (Layanan Administrasi & Catatan Informasi)**! Proyek ini dikembangkan secara terbuka oleh dan untuk anggota **Dinus Open Source Community (DOSCOM)** serta komunitas sumber terbuka secara luas.

Setiap bentuk kontribusi—mulai dari perbaikan bug kecil, dokumentasi, pengujian, hingga implementasi fitur baru—sangat dihargai.

---

## 1. Alur Kerja Kontribusi (Git Workflow)

1. **Fork Repositori**: Buat salinan repositori ini ke akun GitHub Anda.
2. **Clone & Buat Branch Baru**:
   - Seluruh kontribusi fitur atau perbaikan bug **wajib diarahkan ke branch `dev`** (bukan `main`).
   - Berikan nama branch yang deskriptif:
     - `feat/nama-fitur` untuk penambahan fitur baru.
     - `fix/nama-bug` untuk perbaikan bug.
     - `docs/pembaruan-panduan` untuk perubahan dokumentasi.
3. **Konvensi Pesan Commit (Conventional Commits)**:
   Gunakan format commit terstandar:
   - `feat(scope): deskripsi singkat fitur baru`
   - `fix(scope): perbaikan bug spesifik`
   - `refactor(scope): penyederhanaan atau restrukturisasi kode`
   - `docs(scope): penambahan atau pembaruan dokumentasi`
   - `chore(scope): pemeliharaan dependensi atau konfigurasi alat`
4. **Buka Pull Request (PR)**:
   - Targetkan basis PR ke branch `dev`.
   - Isi form PR sesuai dengan templat yang telah disediakan di [.github/PULL_REQUEST_TEMPLATE.md](.github/PULL_REQUEST_TEMPLATE.md).

---

## 2. Prinsip Rekayasa Perangkat Lunak (Design Philosophy)

Dalam ekosistem DOSCOM, kami mengutamakan kode yang dapat dirawat dalam jangka panjang dengan prinsip:

- **Lean & YAGNI (You Aren't Gonna Need It):** Jangan membuat lapisan abstraksi yang tidak diperlukan. Pilih solusi paling sederhana yang langsung menyelesaikan masalah.
- **Standar Library Terlebih Dahulu:** Maksimalkan kapabilitas bawaan Go standard library dan fitur native Web API sebelum menambahkan pustaka pihak ketiga.
- **Shortest Working Diff:** Perubahan yang bersih, terfokus, dan minimal lebih disukai daripada refaktor besar yang menyentuh banyak modul yang tidak relevan.
- **Bebas Ketergantungan Tersembunyi:** Jangan ada *hardcoded domain*, kredensial rahasia, atau tautan *localhost* statis pada berkas produksi.

---

## 3. Lingkungan Pengembangan Lokal

Untuk menjalankan dan menguji kode di lingkungan lokal:

```bash
# Menggunakan Docker (Rekomendasi - Hot Reload Go Air & Vite HMR)
docker compose -f docker-compose.dev.yml up --build

# Verifikasi Backend (Golang)
cd backend
go vet ./...
go build -buildvcs=false -o nul .

# Verifikasi Frontend (Vue 3 + TypeScript)
cd frontend
npm run typecheck
```

Pastikan kedua verifikasi di atas menghasilkan status *exit code 0* (tanpa galat) sebelum mengirimkan Pull Request.

---

## 4. Pelaporan Masalah & Diskusi

- Untuk mendiskusikan ide baru atau melaporkan bug, silakan buat tiket baru di [GitHub Issues](https://github.com/doscom/laci/issues).
- Gunakan bahasa yang sopan, santun, dan konstruktif sesuai dengan [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
- Jika Anda menemukan celah keamanan sensitif, mohon ikuti prosedur pelaporan di [SECURITY.md](SECURITY.md).
