# Kebijakan Keamanan - LACI DOSCOM

Kami di **Dinus Open Source Community (DOSCOM)** memprioritaskan keamanan data dan keandalan sistem pelacakan surat organisasi ini.

---

## 1. Versi yang Didukung

Pembaruan keamanan aktif hanya diberikan pada versi utama cabang pengembangan:

| Versi | Didukung? |
| ----- | --------- |
| `main` (Production Release v2.x) | :white_check_mark: Ya |
| `dev` (Active Development)       | :white_check_mark: Ya |
| Legacy v1.x                      | :x: Tidak |

---

## 2. Pelaporan Kerentanan (Responsible Disclosure)

Jika Anda menemukan potensi celah keamanan (seperti kebocoran kredensial, *broken authentication*, *IDOR*, *SQL injection*, atau *CORS misconfiguration*):

1. **JANGAN membuat GitHub Issue publik.**
2. Laporkan temuan Anda secara privat melalui email ke: `doscom.go@gmail.com` atau hubungi Tim Teknis Pengurus DOSCOM secara langsung.
3. Cantumkan rincian berikut dalam laporan Anda:
   - Deskripsi kerentanan dan potensi dampaknya.
   - Langkah-langkah reproduksi (Proof of Concept / PoC).
   - Saran mitigasi atau patch (jika ada).

Tim Pengurus akan merespons dalam waktu 1x24 jam kerja untuk memvalidasi temuan dan menyiapkan perbaikan terenkripsi sebelum mempublikasikan release patch ke publik.
