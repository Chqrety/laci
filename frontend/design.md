# Design System & UI/UX Guidelines: LACI DOSCOM

Dokumen ini merupakan **Source of Truth** untuk identitas visual, arsitektur antarmuka, dan komponen desain aplikasi **LACI (Pelacakan Surat DOSCOM)**. Seluruh pedoman dirangkum murni berdasarkan implementasi kode aktual pada codebase Vue 3, Tailwind CSS v4, dan font lokal Satoshi.

---

## 1. Filosofi & Arah Desain (Floema Architectural Minimalist)

LACI mengusung gaya **Floema Architectural**:
- **Tenang & Utilitarian**: Panel data tidak berisik, layout renggang dengan napas visual yang cukup (*white space* lapang).
- **Tactile & Responsif**: Menggunakan radius membulat halus (*pill* & *squircle*), elevasi lembut (*subtle shadows*), dan *micro-interactions* instan (180ms ease).
- **Identitas Khas DOSCOM**: Kombinasi warna biru langit (*Sky Blue / DOSCOM Tech*) sebagai warna navigasi utama dan aksen oranye (*DOSCOM Orange*) untuk titik aksi krusial.
- **Ambient Lighting**: Garis grid arsitektural halus (`64px x 64px`, opacity 2.5%–3%) dengan blur bola warna di sudut latar belakang (`blur-2xl` & `blur-3xl`).

---

## 2. Palet Warna (Color Palette)

### 2.1 Warna Utama (Brand Colors)
| Token / Nama | Hex / Nilai | Tailwind Class | Penggunaan Utama |
|---|---|---|---|
| **DOSCOM Sky (Primary)** | `#0284C7` / `#0A84DC` | `bg-sky-600`, `text-sky-600`, `border-sky-500` | Header brand mark, tombol utama, tautan aktif, border hover. |
| **Sky Dark (Primary Hover)** | `#0369A1` | `hover:bg-sky-700`, `text-sky-700` | State hover tombol primer, teks link aktif, avatar badge. |
| **Sky Soft (Primary Tint)** | `#F0F9FF` / `#E0F2FE` | `bg-sky-50`, `border-sky-100`, `text-sky-700` | Badge status arsip/cetak, tag divisi aktif, background avatar. |
| **DOSCOM Orange (Accent)** | `#F97316` | `bg-orange-500`, `text-orange-500` | Aksen tombol submit auth, status humas, highlight kegiatan. |
| **Orange Dark (Accent Hover)**| `#EA580C` | `hover:bg-orange-600` | State hover tombol aksi oranye. |
| **Orange Soft (Accent Tint)** | `#FFF7ED` | `bg-orange-50`, `border-orange-200` | Badge highlight proker, ambient glow latar bawah. |

### 2.2 Warna Latar & Permukaan (Neutrals & Surfaces)
| Token / Nama | Hex / Nilai | Tailwind Class | Penggunaan Utama |
|---|---|---|---|
| **Canvas Background** | `#F8FAFC` | `bg-[#F8FAFC]` (`slate-50`) | Latar belakang seluruh halaman aplikasi. |
| **Card / Panel Surface** | `#FFFFFF` | `bg-white`, `bg-white/95` | Kontainer kartu kanban, form sheet, modal, dan tabel. |
| **Card Surface Muted** | `#F1F5F9` | `bg-slate-100/80`, `bg-zinc-50` | Input form field, pill background navigasi tab. |
| **Border Neutral** | `#E2E8F0` / `#0000000D` | `border-slate-200/90`, `border-black/5` | Border kartu, pembatas tabel, divider list. |
| **Text Primary** | `#0F172A` / `#18181B` | `text-slate-900`, `text-zinc-900` | Heading utama, judul surat, nama akun. |
| **Text Secondary / Muted** | `#64748B` / `#71717A` | `text-slate-500`, `text-zinc-500` | Subheading, label keterangan, placeholder, ikon non-aktif. |
| **Text Tertiary / Subtle** | `#94A3B8` / `#A1A1AA` | `text-slate-400`, `text-zinc-400` | Nomor urut tabel, timestamp, teks footer hak cipta. |

### 2.3 Status Tracking & Semantic Badges
| Kategori Status | Background | Text & Border | Makna / Kolom Kanban |
|---|---|---|---|
| **Standby / Netral** | `bg-zinc-100` | `text-zinc-700 border-zinc-200` | Draft surat, antrean awal. |
| **Cetak / Diproses** | `bg-sky-50` | `text-sky-700 border-sky-200` | Surat dalam proses cetak/penggandaan. |
| **TTD Lapis / Internal** | `bg-amber-50` | `text-amber-700 border-amber-200` | Menunggu paraf sekretaris / TTD Ketum. |
| **TTD Eksternal / Pejabat** | `bg-purple-50` | `text-purple-700 border-purple-200` | Menunggu TTD Pembina, Koormawa, Dekanat. |
| **Selesai / Approved** | `bg-emerald-50` | `text-emerald-700 border-emerald-200` | Surat tuntas ditandatangani dan diarsipkan. |
| **Error / Ditolak** | `bg-rose-50` | `text-rose-700 border-rose-200` | Validasi gagal, surat dibatalkan/dihapus. |

---

## 3. Tipografi (Typography)

Seluruh aplikasi menggunakan font **Satoshi** lokal dalam format modern `.woff2` (`Satoshi-Variable.woff2` dan fallback `Satoshi-Medium.woff2`).

```css
@font-face {
  font-family: 'Satoshi';
  src: url('./assets/fonts/Satoshi-Variable.woff2') format('woff2-variations'),
       url('./assets/fonts/Satoshi-Medium.woff2') format('woff2');
  font-weight: 300 900;
  font-display: swap;
  font-style: normal;
}
```

### 3.1 Skala Tipografi
- **Headings (h1–h6)**:
  - Font: `Satoshi`
  - Weight: `font-medium` (500) hingga `font-semibold` (600)
  - Tracking: `tracking-tight` (`letter-spacing: -0.025em`)
  - Leading: `leading-snug` (1.35) atau `leading-none`
- **Body & Teks Utama**:
  - Font: `Satoshi`, `font-normal` (400) atau `font-medium` (500)
  - Ukuran umum: `text-xs` (12px) untuk antarmuka rapat/dashboard, `text-sm` (14px) untuk form teks.
- **Identitas Khusus (Monospace)**:
  - Font: `font-mono` (`ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas`)
  - Penggunaan: Nomor surat (`015/DOSCOM/MAKRAB/X/2026`), NIM (`A11.2022.14000`), tanggal ISO, tag role kode.

---

## 4. Sistem Bentuk & Radius Sudut (Shapes & Radii)

Untuk mempertahankan kesan *tactile* dan ergonomis:
- **Pill Shape (`rounded-full`)**:
  - Tombol aksi utama (*CTA*).
  - Kontainer tab navigasi header (`nav.bg-slate-100/80.rounded-full`).
  - Filter chip status dan badge role user (`SEKRE`, `HUMAS`).
- **Super-Smooth Rounded Cards (`rounded-3xl` / 24px)**:
  - Kontainer utama kartu profil (`Profile.vue`).
  - Panel form modal popup (`KanbanModalForm.vue`).
  - Kontainer tabel data (`SuratManagement.vue`, `UserManagement.vue`).
  - Kotak dialog otentikasi login.
- **Standard Card & Inputs (`rounded-2xl` / 16px)**:
  - Kartu individual kanban board (`KanbanCard.vue`).
  - Seluruh elemen `<input>`, `<select>`, dan `<textarea>`.
  - Alert notification banner (`bg-emerald-50`, `bg-rose-50`).
- **Micro Badges (`rounded-lg` / 8px)**:
  - Tag nomor surat kecil di dalam kartu kanban.
  - Tag label counter statistik.

---

## 5. Komponen Desain Standar

### 5.1 Tombol (Button Styles)
```html
<!-- Primary Button (Sky Blue) -->
<button class="px-5 py-2.5 rounded-full text-xs font-bold text-white bg-sky-600 hover:bg-sky-700 active:scale-[0.98] shadow-xs transition-all disabled:opacity-50 cursor-pointer">
  Simpan Data
</button>

<!-- Accent CTA Button (DOSCOM Orange Pill) -->
<button class="px-8 py-3.5 rounded-full text-sm font-bold text-white bg-orange-500 hover:bg-orange-600 active:scale-[0.98] shadow-sm hover:shadow transition-all cursor-pointer">
  Konfirmasi & Masuk →
</button>

<!-- Secondary / Outline Button -->
<button class="px-4 py-2 rounded-full text-xs font-semibold bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 shadow-xs active:scale-[0.98] transition-all cursor-pointer">
  Muat Ulang
</button>

<!-- Danger / Destructive Button -->
<button class="px-4 py-2 rounded-full text-xs font-bold bg-rose-600 text-white hover:bg-rose-700 active:scale-[0.98] transition-all cursor-pointer">
  Hapus Pengurus
</button>
```

### 5.2 Form Input
Semua input form memiliki *padding* nyaman, sudut `rounded-2xl`, latar *subtle* abu-abu terang, dan efek cincin fokus biru DOSCOM:
```html
<div>
  <label class="block text-xs font-medium text-zinc-700 mb-1">Nama Pengurus</label>
  <input
    type="text"
    placeholder="Masukkan nama lengkap"
    class="w-full px-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
  />
  <p class="text-[11px] text-zinc-400 mt-1">Gunakan nama resmi organisasi.</p>
</div>
```

### 5.3 Kartu Kanban (`KanbanCard.vue`)
- Menggunakan `rounded-2xl`, `border-slate-200/90`, dan efek `hover:-translate-y-0.5 hover:border-sky-400`.
- Menyertakan nomor surat monospaced di pojok kiri atas, judul perihal semi-bold dengan `line-clamp-2`, PIC dengan penanda dot berwarna, serta indikator lampiran berkas jika arsip sudah dipindai.

### 5.4 Tabel Data & Panel Kontainer
- Dibungkus dalam kontainer `bg-white/95 border border-slate-200/90 rounded-3xl shadow-xs overflow-hidden`.
- Header tabel: `bg-black/[0.02] text-zinc-500 font-semibold px-5 py-4 text-xs`.
- Baris tabel: `hover:bg-black/[0.02] transition-colors divide-y divide-black/5`.

---

## 6. Animasi & Transisi Interaksi (Interaction Transitions)

Semua transisi diatur secara global pada `frontend/src/style.css` agar bebas dari *layout-shift* (*jank*):

1. **Micro-interactions Elemen Interaktif (180ms)**:
   ```css
   button, input, select, textarea, a {
     transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform;
     transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
     transition-duration: 180ms;
   }
   ```
2. **Page View Transition (`.page-fade-*`)**:
   - Durasi: 220ms ease-out (masuk), 150ms ease-in (keluar).
   - Efek: Perpindahan halaman dengan fade halus dan pergeseran vertikal `translateY(6px)` ke `0px`.
3. **Modal Dialog Pop (`.modal-fade-*`)**:
   - *Decoupled Backdrop & Panel*: Backdrop hanya memudarkan opasitas (0 ke 1) tanpa skala, sedangkan dialog panel bagian dalam (`.modal-panel`) melakukan transisi `scale(0.96) translateY(6px)` ke `scale(1) translateY(0)` dengan kurva `cubic-bezier(0.16, 1, 0.3, 1)`.
4. **Drawer Slide-over Sheet (`.drawer-*`)**:
   - Lembar detail samping kanan meluncur dari tepi luar layar (`translateX(100%)` ke `0`) dalam durasi 280ms kurva luwes `cubic-bezier(0.16, 1, 0.3, 1)`.

---

## 7. Aturan Penerapan untuk Fitur Baru (Rules for Contributors)

1. **Wajib Menggunakan Font Satoshi**: Hindari font CDN pihak ketiga atau default browser sans-serif; selalu gunakan font stack `font-family: 'Satoshi'`.
2. **Gunakan Pill Buttons untuk Aksi Utama**: Tombol navigasi dan aksi selalu menggunakan `rounded-full`.
3. **Hindari Tombol Sudut Kaku (Sharp Angles)**: Gunakan minimal `rounded-2xl` untuk input dan `rounded-3xl` untuk card/modal utama.
4. **Jaga Kontras & Aksesibilitas**: Pastikan teks di atas latar putih memakai minimal `text-slate-700` atau `text-zinc-800`.
5. **Pertahankan Transisi Halus**: Selalu bungkus modal dengan `<Transition name="modal-fade">` dan tambahkan kelas `.modal-panel` pada kontainer dialog intinya.

---

## 8. Arsitektur UI & Form Validation

Untuk menjaga konsistensi antarmuka, keandalan data, serta pengalaman pengguna yang seragam di seluruh aplikasi, wajib mengikuti arsitektur komponen dan validasi form berikut:

### 8.1 Komponen UI Utama (Shadcn-Vue Mandatory)
- Semua pengembangan elemen antarmuka (seperti `Button`, `Input`, `Textarea`, `Badge`, `Label`, `Dialog`/Modal, `Select`, dan `Toast`) **WAJIB** menggunakan komponen terpusat dari `@/components/ui/*` (`shadcn-vue`).
- **Dilarang** membuat komponen form mentah dari nol atau menulis tag HTML dasar secara manual jika komponen Shadcn-Vue yang sesuai sudah tersedia di dalam proyek.

### 8.2 Form State & Schema Validation (Vee-Validate + Zod)
- **Ekosistem Validasi**: Seluruh manajemen form wajib menggunakan ekosistem `vee-validate`, `@vee-validate/zod`, dan `zod` untuk pendefinisian skema validasi deklaratif.
- **Tanpa State Error Manual**: Dilarang mengelola state error secara manual (contoh: `const error = ref('')`, `const isInvalid = ref(false)`) atau mengandalkan validasi HTML native browser.
- **Atribut Form**: Wajib menyertakan atribut `novalidate` pada setiap tag `<form>` untuk menonaktifkan tooltip bawaan browser dan memastikan konsistensi interaksi visual kustom.

### 8.3 Error Handling Visual (<FormMessage>)
- Pesan kesalahan validasi wajib dirender secara otomatis menggunakan komponen `<FormMessage />` dari Shadcn-Vue.
- Input yang berada dalam kondisi error otomatis mendapatkan border berwarna merah (`border-destructive` / `border-rose-500`) dan pesan teks peringatan merah di bawah input, sepenuhnya selaras dengan token warna semantik DOSCOM di atas.

