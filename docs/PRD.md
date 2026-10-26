# Product Requirement Document (PRD): LACI DOSCOM v2

- **Produk:** LACI (Layanan Arsip & Cek Informasi Surat)
- **Organisasi:** DOSCOM (Dinus Open Source Community)
- **Status:** Release Candidate (v2.0 Architecture)
- **Filosofi Arsitektur:** Ponytail Protocol (Zero Bloat, Zero Cost, Serverless Webhook, Lean Engineering)

---

## 1. Ringkasan Eksekutif & Tujuan Sistem

### 1.1 Latar Belakang
Persuratan organisasi mahasiswa sering kali menghadapi kendala klasik:
1. Pembuatan surat keluar dilakukan secara manual dengan menduplikasi file Google Docs satu per satu dan mengedit bagian kecil secara berulang.
2. Pemantauan alur persetujuan fisik (tanda tangan Ketum, Pembina, Koormawa) tersebar di chat personal tanpa pelacakan transparan.
3. Arsitektur autentikasi lama (v1) menyimpan sandi secara terbaca (*plaintext/display_password*) dengan password bawaan seragam (`123`), dan belum memiliki pemisahan hak akses hierarkis yang merefleksikan tata kelola organisasi riil.

### 1.2 Tujuan Utama (v2)
1. **Zero-Knowledge Security:** Meniadakan password default dan menghapus kolom `display_password`. Sistem menerapkan *self-activation* via tautan email resmi.
2. **Bootstrap Genesis Ketum:** Inisialisasi mandiri akun Ketua Umum perdana dari konfigurasi lingkungan (`.env`) tanpa celah pembuatan akun Ketum liar oleh admin biasa.
3. **Suksesi Jabatan Sah:** Pergantian Ketua Umum hanya dapat dilakukan melalui mekanisme resmi Serah Terima Jabatan dengan verifikasi kata sandi aktif Ketum saat ini.
4. **Generator Surat GDocs Otomatis:** Integrasi nir-biaya (*zero-cost*) dengan Google Apps Script (GAS) untuk menyinkronkan seluruh template dari folder Google Drive dan meng-generate dokumen secara instan berbasis variabel `{{placeholder}}`.
5. **Pelacakan Alur Kanban Responsif:** Papan Kanban 8 tahapan terstruktur dengan audit log otomatis dan pembatasan unggah bukti fisik (hanya saat status `Selesai`).

---

## 2. Matriks Otoritas & Peran Pengguna (Role-Based Access)

Sistem menggunakan kesatuan atribut peran dan divisi (`divisi`) dengan tingkatan otoritas berikut:

| Modul / Kemampuan | Ketua Umum (`ketum`) | Sekretaris (`sekre`) | Humas (`humas`) | Pengurus (`pengurus`) | Ex-Ketua Umum (`ex-ketum`) |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Tingkat Otoritas** | **Superadmin (God Mode)** | **Operational Admin** | **Operator Alur** | **Operator Alur** | **Demisioner / Anggota** |
| **Kanban Board (Lihat & Geser Kartu)** | Ya | Ya | Ya | Ya | Ya |
| **Buat Surat Baru (Manual / Generator)** | Ya | Ya | Ya | Ya | Ya |
| **Tambah Catatan Instruksi di Kartu** | Ya | Ya | Ya | Ya | Ya |
| **Unggah Bukti Arsip Fisik** | Kapan saja | Kapan saja | **Hanya saat "Selesai"** | **Hanya saat "Selesai"** | **Hanya saat "Selesai"** |
| **Kelola Master Template GDocs & Sync** | Ya | Ya | Tidak | Tidak | Tidak |
| **Kelola Master Program Kerja (Proker)** | Ya | Ya | Tidak | Tidak | Tidak |
| **Hapus Arsip Surat** | Ya | Ya | Tidak | Tidak | Tidak |
| **Tambah Pengurus (Humas & Pengurus)** | Ya | Ya | Tidak | Tidak | Tidak |
| **Tambah Sesama Sekretaris (`sekre`)** | Ya | Ya | Tidak | Tidak | Tidak |
| **Serah Terima Jabatan (`ketum`)** | **Ya (Khusus Ketum)** | **DILARANG** | Tidak | Tidak | Tidak |
| **Ubah Profil & Password Sendiri** | Ya | Ya | Ya | Ya | Ya |

> [!NOTE]
> Divisi `ex-ketum` tidak dapat dipilih secara manual pada formulir tambah pengurus. Divisi ini hanya didapatkan secara eksklusif ketika Ketua Umum aktif menyerahkan mandat kepemimpinannya kepada pengurus lain melalui menu **Serah Terima Jabatan**.

---

## 3. Spesifikasi Arsitektur Modul

### 3.1 Bootstrap Genesis Ketum & Alur Autentikasi v2
1. **Genesis Ketum Initialization:**
   - Saat backend melakukan *booting* pertama kali, seeder memeriksa keberadaan akun dengan `divisi = 'ketum'`.
   - Jika belum ada, sistem membaca `INIT_KETUM_NIM`, `INIT_KETUM_NAMA`, `INIT_KETUM_EMAIL`, dan `INIT_KETUM_USERNAME` dari `.env`.
   - Akun dibuat dengan status `is_active = false` dan password acak terkunci, lalu memicu pengiriman email aktivasi mandiri bertenggat 24 jam.
   - Tautan aktivasi dicetak ke log backend sebagai fallback jika SMTP belum terkonfigurasi.
2. **Aktivasi Akun Mandiri:**
   - Pengguna membuka tautan aktivasi di peramban (`/activate?token=...`).
   - Token divalidasi ke backend. Pengguna membuat password baru (minimal 6 karakter).
   - Akun menjadi aktif (`is_active = true`), token dibersihkan, dan pengguna dapat langsung login.
3. **Pemulihan Kata Sandi (*Forgot Password*):**
   - Pengguna memasukkan email terdaftar di `/forgot-password`.
   - Token reset berlaku 1 jam dikirimkan ke email terdaftar.
   - Sistem merespons dengan pesan netral untuk mencegah *user enumeration attack*.
4. **Serah Terima Jabatan Ketua Umum:**
   - Ketua Umum aktif memilih pengurus penerima mandat dan memverifikasi kata sandi aktifnya saat ini.
   - Dalam satu transaksi basis data:
     - Akun penerima dinaikkan menjadi `divisi = 'ketum'`.
     - Akun Ketua Umum lama diturunkan statusnya menjadi `divisi = 'ex-ketum'`.

---

### 3.2 Generator Surat Sekre (Integrasi Google Drive & Apps Script)
1. **Prinsip Bebas Biaya (*Zero-Cost Bridge*):**
   - Berjalan serverless menggunakan Web App Google Apps Script (GAS) tanpa perlu Google Cloud Console berbayar.
   - Database LACI hanya menyimpan metadata, URL Google Docs, dan URL PDF hasil generate; penyimpanan berkas fisik berada di Google Drive organisasi.
2. **Alur Sinkronisasi Template (Batch Folder Discovery):**
   - Sekretaris menaruh seluruh template dokumen Google Docs pada 1 Folder Master di Google Drive.
   - Di menu *Setup Template*, Sekretaris menempelkan link Folder Drive dan mengklik **Sync Template**.
   - Skrip GAS memindai seluruh dokumen, mengekstrak variabel bermotif `{{nama_variabel}}`, dan memperbarui basis data LACI secara otomatis.
3. **Alur Generate Surat:**
   - Pengguna memilih jenis template surat dari daftar hasil sinkronisasi.
   - Antarmuka otomatis merender formulir dinamis sesuai variabel yang ditemukan di template.
   - Saat disubmit, GAS menduplikasi file master, mengganti seluruh placeholder, menyimpannya di folder output yang ditentukan, dan mengembalikan tautan Google Docs serta tautan unduh instan PDF.
   - Sistem mencatat surat baru ke papan Kanban dengan status `Standby`.

---

### 3.3 Pelacakan Alur Surat (Kanban Board)
1. **8 Tahapan Resmi DOSCOM:**
   - `Standby` (Surat dibuat, menunggu dicetak)
   - `Cetak` (Berkas fisik telah dicetak)
   - `TTD Dasar` (Tanda tangan panitia pelaksana/PIC)
   - `TTD Ketum` (Tanda tangan Ketua Umum)
   - `TTD Pembina` (Tanda tangan Dosen Pembina Komunitas)
   - `Paraf Koormawa` (Paraf Koordinator Kemahasiswaan)
   - `TTD Tertinggi` (Tanda tangan Wakil Dekan / Pejabat Fakultas)
   - `Selesai` (Surat tuntas & siap diarsipkan)
2. **Audit Trail Otomatis:**
   - Setiap perpindahan kolom mencatat record baru pada tabel `log_trackings` berisi status baru, catatan mutasi, pengubah, dan stempel waktu.
3. **Proteksi Bukti Arsip Fisik:**
   - Pengurus umum (`humas`, `pengurus`, `ex-ketum`) hanya diizinkan mengunggah foto arsip fisik setelah surat berstatus `Selesai` guna mencegah arsip prematur sebelum surat ditandatangani lengkap.

---

### 3.4 Manajemen Program Kerja & Anggota
1. **Program Kerja (Proker):**
   - Master data kode proker untuk penamaan dan klasifikasi surat dinamis (misal: MAKRAB, OPREC, DST, OOTS).
2. **Impor Pengurus Cepat:**
   - **CSV Import:** Unggah berkas CSV dengan kolom `nim,nama,email,divisi`.
   - **Google Sheets Import:** Impor langsung dari URL Google Spreadsheet publik tanpa perlu unduh-unggah manual.
   - Seluruh anggota hasil impor otomatis menerima email aktivasi mandiri.

---

## 4. Spesifikasi Skema Database

```
┌─────────────────┐       1:N       ┌──────────────────┐
│      surat      ├────────────────►│  log_trackings   │
│   (ID: UUID)    │                 │   (ID: UUID)     │
└────────┬────────┘                 └──────────────────┘
         │ N:1
         ▼
┌─────────────────┐
│     prokers     │
│   (ID: UUID)    │
└─────────────────┘

┌─────────────────┐                 ┌──────────────────┐
│      users      │                 │ surat_templates  │
│   (ID: UUID)    │                 │   (ID: UUID)     │
└─────────────────┘                 └──────────────────┘

┌─────────────────┐
│  app_settings   │ (Key-Value Config: gas_webhook_url, default_output_folder_url)
└─────────────────┘
```

1. **`users`**: `id`, `nim`, `nama`, `email`, `username`, `password`, `divisi`, `is_active`, `activation_token`, `activation_token_expires_at`, `reset_password_token`, `reset_password_expires_at`.
2. **`surat`**: `id`, `nomor_surat`, `perihal`, `tujuan`, `status_saat_ini`, `pic_nama`, `catatan`, `arsip_url`, `gdocs_url`, `proker_id`.
3. **`log_trackings`**: `id`, `surat_id`, `status_baru`, `catatan_log`, `waktu_update`.
4. **`prokers`**: `id`, `nama_proker`, `deskripsi`.
5. **`surat_templates`**: `id`, `nama_template`, `gdoc_id`, `placeholders` (JSON string array), `is_active`.
6. **`app_settings`**: `key` (Primary Key), `value`.

---

## 5. Ringkasan Endpoint API

* **Autentikasi & Profil:**
  * `POST /login`
  * `GET /auth/verify-token`
  * `POST /auth/activate`
  * `POST /auth/forgot-password`
  * `POST /auth/reset-password`
  * `GET /profile` (JWT)
  * `PUT /profile/change-username` (JWT)
  * `PUT /profile/change-password` (JWT)
* **Manajemen Pengurus (Sekre & Ketum):**
  * `GET /users`, `POST /users`, `PUT /users/:id`, `DELETE /users/:id`, `DELETE /users/batch`
  * `POST /users/import-csv`, `POST /users/import-spreadsheet`
  * `POST /users/transfer-ketum` (Khusus Ketum)
* **Surat & Proker:**
  * `GET /letters`, `POST /letters`, `PUT /letters/:id`, `POST /letters/:id/upload`, `DELETE /letters/:id`
  * `GET /programs`, `POST /programs`, `PUT /programs/:id`, `DELETE /programs/:id`
* **Generator Surat & Pengaturan:**
  * `GET /letters/templates`, `POST /letters/templates/sync`, `POST /letters/generate`
  * `GET /settings`, `PUT /settings`
