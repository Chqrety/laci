# Panduan Pembuatan Template Google Docs untuk Sekretaris (LACI)

Panduan praktis bagi Sekretaris (Sekre) DOSCOM untuk menyiapkan dokumen template surat di Google Docs agar dapat di-generate otomatis secara sempurna oleh sistem **LACI**.

---

## 1. Format Penamaan File Template di Google Drive

Simpan dokumen template di Google Drive dalam format **Google Docs** (bukan file `.docx` mentah). Berikan nama file dengan awalan nomor urut 2 digit:

| Nama File di Google Drive | Nomor Urut Otomatis | Perihal Otomatis |
| :--- | :---: | :--- |
| `01_Surat Undangan` | `01` | Surat Undangan |
| `02_Surat Peminjaman Gedung` | `02` | Surat Peminjaman Gedung |
| `03_Surat Permohonan Delegasi` | `03` | Surat Permohonan Delegasi |
| `04_Surat Tugas` | `04` | Surat Tugas |

> [!NOTE]
> Awalan angka (misal `01_`) otomatis dibaca LACI sebagai nomor urut awal dan nama sisanya otomatis mengisi field **Perihal**. Di aplikasi, nomor urut dan perihal ini tetap bisa Anda edit manual jika diperlukan.

---

## 2. Aturan Tag Variabel `{{placeholder}}`

LACI mengganti teks di dalam tanda kurung kurawal ganda `{{nama_variabel}}` dengan data yang diisi melalui form. 
- Gunakan huruf kecil dan garis bawah (`_`).
- Tidak boleh ada spasi di dalam nama variabel (contoh: `{{nama_kegiatan}}`, BUKAN `{{nama kegiatan}}`).

---

## 3. Daftar Tag Standar yang Didukung Otomatis

Form **Buat Surat** di LACI telah mengenali tag-tag berikut secara otomatis:

### A. Nomor, Tanggal Terbit, & Identitas Surat
| Tag Placeholder | Letak di Surat | Sumber Data | Contoh Hasil Render di Dokumen |
| :--- | :--- | :--- | :--- |
| `{{nomor_surat}}` | Kiri Atas | Segmented Form | `01/DOSCOM/OPREC-2026/X/2026` atau `01/DOSCOM/X/2026` |
| `{{tanggal_surat}}` | **Kanan Atas (Titimangsa)** | **Otomatis Hari Ini (`new Date`)** | `25 Oktober 2026` |
| `{{perihal}}` | Kiri Atas | Nama template / Form | `Surat Undangan` |

> [!TIP]
> - **Tanggal Surat (`{{tanggal_surat}}`):** Otomatis diisi tanggal hari ini saat surat digenerate. Sekre **tidak perlu input manual**.
> - **Tujuan Dokumen ("Kepada Yth. ..."):** Dibuat **statis** langsung di dalam dokumen template masing-masing (misal: *Kepada Yth. Kepala UPT Komputer*), sehingga Sekre tidak perlu mengetik tujuan secara berulang di form.

### B. Detail Pelaksanaan / Acara / Peminjaman (Input Manual)
| Tag Placeholder | Deskripsi | Contoh Hasil Render di Dokumen |
| :--- | :--- | :--- |
| `{{nama_kegiatan}}` | Nama agenda atau acara | `Penutupan DOSCOM University 2026` |
| `{{hari_tanggal}}` | Format tanggal kegiatan bahasa Indonesia | `Senin, 20 Oktober 2026` *(1 Hari)* <br>atau<br> `Senin, 20 Oktober 2026 s.d. Selasa, 21 Oktober 2026` *(Rentang Hari)* |
| `{{waktu}}` | Jam pelaksanaan agenda | `08:00 WIB - Selesai` |
| `{{tempat}}` | Lokasi pelaksanaan agenda | `Gedung H Lantai 7 UDINUS` |

### C. Narahubung (Contact Person)
Format standar yang direkomendasikan adalah meletakkan label `Nama` dan `No.Hp` di template Google Docs secara terpisah:

```text
Adapun narahubung (ketua pelaksana) yang dapat dihubungi:
Nama   : {{nama_narahubung}}
No.Hp  : {{nohp_narahubung}}
```

| Tag Placeholder | Deskripsi | Contoh Hasil Render di Dokumen |
| :--- | :--- | :--- |
| `{{nama_narahubung}}` | Nama narahubung (Ketua Pelaksana/CP) | `Fujikawa Shinichi` |
| `{{nohp_narahubung}}` | Nomor HP / WhatsApp aktif | `0896 5354 9412` |
| `{{narahubung}}` | Format 2 baris otomatis (Nama & No.Hp) | `Nama   : Fujikawa Shinichi`<br>`No.Hp  : 0896 5354 9412` |

### D. Tanda Tangan Pengurus (Ketum & Penandatangan Dasar)
| Tag Placeholder | Posisi di Dokumen | Keterangan |
| :--- | :---: | :--- |
| `{{nama_ketum}}` | **Kiri** | Nama lengkap Ketua Umum (diambil dari akun berdivisi `ketum`) |
| `{{nim_ketum}}` | **Kiri** | NIM Ketua Umum |
| `{{nama_dasar}}` | **Kanan** | Nama Penandatangan Dasar (Ketua Pelaksana / Pembuat surat) |
| `{{nim_dasar}}` | **Kanan** | NIM Penandatangan Dasar |

> [!NOTE]
> Tag alias berikut juga otomatis didukung:
> - `{{nama_tingkat_1}}` / `{{nim_tingkat_1}}`
> - `{{nama_lapis_1}}` / `{{nim_lapis_1}}`
> - `{{nama_ketupel}}` / `{{nim_ketupel}}` (kompatibel untuk template lama)

> [!IMPORTANT]
> **Tanda Tangan Pembina & Wakil Rektor III:** Tulis **statis** langsung di template dokumen (Nama, NIDN/NIP, dan jabatannya), karena pejabat pembina dan WR III tidak berganti-ganti setiap surat.

---

## 4. Contoh Penataan Format di Google Docs

Salin struktur layout berikut ke dalam file Google Docs Anda:

```text
Nomor   : {{nomor_surat}}                                   Semarang, {{tanggal_surat}}
Lampiran: -
Perihal : {{perihal}}

Kepada Yth.
Ketua Ormawa Universitas Dian Nuswantoro
di Tempat

Dengan hormat,
Sehubungan dengan diselenggarakannya agenda {{nama_kegiatan}}, kami selaku pengurus DOSCOM bermaksud mengundang Saudara/i untuk hadir pada:

  Hari, Tanggal : {{hari_tanggal}}
  Waktu         : {{waktu}}
  Tempat        : {{tempat}}
Demikian surat permohonan ini kami sampaikan, atas perhatiannya kami ucapkan terima kasih.

Adapun narahubung (ketua pelaksana) yang dapat dihubungi:
Nama   : {{nama_narahubung}}
No.Hp  : {{nohp_narahubung}}


Hormat kami,
Pengurus DOSCOM Periode 2026/2027

   Ketua Umum,                              Ketua Pelaksana / Penandatangan Dasar,



{{nama_ketum}}                           {{nama_dasar}}
NIM. {{nim_ketum}}                       NIM. {{nim_dasar}}


                            Mengetahui,
              Pembina UKM DOSCOM Universitas Dian Nuswantoro



                        (Nama Pembina Statis)
                        NPP. 0686.11.1990.001
```

---

## 5. Menambahkan Variabel Kustom Tambahan

Jika suatu template membutuhkan variabel di luar daftar di atas (misal `{{ruangan}}`, `{{nama_barang}}`, atau `{{alasan}}`):
1. Cukup ketik `{{ruangan}}` di dokumen template Google Docs Anda.
2. Saat Anda memilih template tersebut di LACI, form akan secara otomatis membuat input baru **"RUANGAN"** di bagian bawah. Sekre tinggal mengisi nilainya.

---

## 6. Langkah Sinkronisasi & Testing di LACI

1. Letakkan seluruh file Google Docs di folder Google Drive template Anda.
2. Buka aplikasi **LACI** $\rightarrow$ menu **Surat**.
3. Klik tombol **Setup Template**:
   - Pastikan **Web App URL (Google Apps Script)** sudah terisi.
   - Masukkan **Link Folder Template** Google Drive.
   - Masukkan **Link Folder Output** (tempat file hasil generate disimpan).
   - Klik **Simpan Pengaturan**, lalu klik **Sinkronkan Template Sekarang**.
4. Klik tombol **+ Buat Surat**:
   - Pilih template yang ingin diuji.
   - Periksa format nomor surat, tanggal, narahubung, dan penandatangan.
   - Klik **Generate Surat**.
   - Klik link **Buka Google Docs** dan **Download PDF** untuk memverifikasi dokumen hasil generate.
