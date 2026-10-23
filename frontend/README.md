# 💻 Laci Frontend Client

Aplikasi web Single Page Application (SPA) **Laci DOSCOM** dibangun menggunakan **Vue 3 (Composition API)**, **Vite**, dan **Tailwind CSS v4** dengan fokus pada estetika minimalis (*Floema Architecture*), responsivitas mobile, dan performa tinggi.

---

## 🛠️ Tech Stack & Ekosistem

* **Framework:** [Vue 3](https://vuejs.org/) (`v3.5.42`) — Composition API `<script setup lang="ts">`
* **Bundler & Tooling:** [Vite](https://vite.dev/) (`v8.3.3`) & [Vue TSC](https://github.com/vuejs/language-tools)
* **Styling:** [Tailwind CSS v4](https://tailwindcss.com/) via `@tailwindcss/vite` (Zero config PostCSS)
* **Headless UI Primitives:** [Radix-Vue](https://www.radix-vue.com/) (`v1.9.17`)
* **Typography:** Font **Satoshi Variable** lokal (`.woff2`)
* **HTTP Client:** [Axios](https://axios-http.com/) (`v1.20.0`) dengan interceptor JWT terpusat
* **OCR Scanning Engine:** [Tesseract.js](https://tesseract.projectnaptha.com/) (`v7.0.0`) on-demand WASM worker

---

## ⚡ Prinsip Arsitektur & Optimasi Kunci

### 1. Zero-Bloat Form Architecture
* **Native Vue & HTML5:** Form input menggunakan `v-model` reaktif langsung dipadukan dengan atribut validasi native browser (`required`, `minlength`, `type="email"`).
* **Toast Notification System:** Feedback validasi dan pesan backend ditampilkan terpusat melalui komponen Toast Radix (`use-toast.ts`), menghindari boilerplate wrapper form yang berat.

### 2. High-Performance On-Device OCR Scanner
* Terletak di composable [`useDocumentScanner.ts`](src/composables/useDocumentScanner.ts).
* **Canvas Pre-processing:** Foto kamera resolusi tinggi (12–48 MP) di-downscale otomatis ke maksimal 1500px serta diubah ke mode grayscale threshold kontras tinggi sebelum diproses OCR. Memangkas waktu scan dari ~30 detik menjadi **< 1 detik**.
* **Lazy Dynamic Import & Worker Termination:** Modul `tesseract.js` (~3 MB) **tidak** dimasukkan ke initial bundle utama; hanya diunduh saat user menekan tombol scan dokumen, dan worker RAM langsung di-terminate setelah pembacaan selesai.

### 3. Route-Level Code Splitting
* Seluruh halaman di [`router/index.ts`](src/router/index.ts) diimpor secara dinamis (`component: () => import('@/views/...')`).
* Ukuran bundle awal aplikasi ditekan hingga **~320 kB (109 kB gzip)** sehingga instan dimuat di koneksi 3G/4G smartphone.

### 4. Stacking Context & Teleport Hierarchy
Hirarki `z-index` terstandarisasi untuk mencegah dropdown atau dialog tertutup elemen lain:
* Modal Dialog: `z-[100]`
* Toaster Notification: `z-[110]`
* Dropdown Select Menu: `z-[120]`
* Drawer Detail Surat: Ditranslasikan keluar via `<Teleport to="body">` dengan *body scroll lock* otomatis saat aktif di layar HP.

---

## 📁 Struktur Direktori Frontend

```text
frontend/
├── public/                 # Static assets (logo-doscom.png, favicon.ico)
├── src/
│   ├── assets/             # Font lokal Satoshi (.woff2)
│   ├── components/
│   │   ├── molecules/      # CardDetail drawer, KanbanFilter bar
│   │   ├── organisms/      # KanbanCard, KanbanModalForm, PasswordReminderModal
│   │   └── ui/             # Radix primitives: badge, button, dialog, input, label, select, textarea, toast
│   ├── composables/        # useDocumentScanner.ts, useKanbanDragDrop.ts
│   ├── lib/                # utils.ts (Tailwind cn merge helper)
│   ├── router/             # Vue Router dengan route guard & RBAC check
│   ├── services/           # api.ts (Letter, Program, User, & Auth Axios services)
│   ├── views/              # Halaman: KanbanBoard, SuratManagement, ProkerManagement, UserManagement, Login, Profile
│   ├── App.vue             # Root layout & Floema navigation bar
│   ├── main.ts             # Inisialisasi Vue application
│   └── style.css           # Tailwind v4 import & global styles
├── design.md               # Source of truth desain Floema DOSCOM
└── vite.config.ts          # Konfigurasi plugin Vite & path alias (@)
```

---

## 🚀 Panduan Eksekusi Lokal

### 1. Install Dependencies
```bash
cp .env.example .env
npm ci
```

### 2. Konfigurasi Environment (`frontend/.env`)
```env
# Alamat REST API Backend yang dapat diakses browser
VITE_API_URL=http://localhost:8080
```

### 3. Jalankan Dev Server
```bash
npm run dev
```
Aplikasi berjalan di `http://localhost:5173` dengan *hot-module-replacement* (HMR).

### 4. Production Build & Lint Type-Check
```bash
npm run build
```
Menjalankan `vue-tsc --build` untuk validasi tipe TypeScript secara menyeluruh dan meng-output bundle statis ke direktori `dist/`.
