# Frontend Laci

Frontend menggunakan Vue 3 dan Vite. Semua request API serta URL lampiran relatif dari backend memakai `VITE_API_URL`.

## Development lokal

```sh
cp .env.example .env
npm ci
npm run dev
```

Atur `VITE_API_URL` di `.env` ke URL backend yang dapat diakses browser. Contoh lokal tersedia di `.env.example`; untuk akses dari perangkat lain, gunakan alamat backend yang dapat dijangkau perangkat tersebut. Tidak ada fallback hostname atau port di kode aplikasi.

Isi `VITE_API_URL` dengan URL HTTP(S) backend yang dapat diakses browser. Restart dev server setelah mengubah `.env`.

## Build

```sh
npm run build
```

Vite memasukkan `VITE_API_URL` ke bundle **saat build**. Gunakan `.env.production`, `.env.production.local`, atau environment proses untuk nilai production; environment proses memiliki prioritas tertinggi. Setelah mengganti URL, jalankan build ulang. Variabel `VITE_*` terlihat oleh browser, sehingga hanya boleh berisi konfigurasi publik.

## Docker Compose

Jalankan dari root repositori:

```sh
cp .env.example .env
docker compose up -d --build
```

Docker Compose membaca konfigurasi dari root `.env`. `frontend/.env` digunakan untuk development lokal.

| Variabel | Penggunaan | Default Compose |
| --- | --- | --- |
| `VITE_API_URL` | Build argument frontend, digunakan browser untuk request API dan lampiran. | `/api` |
| `BACKEND_URL` | Environment runtime Nginx, alamat backend dalam jaringan Docker tanpa trailing slash. | `http://backend:8080` |

Nginx meneruskan `/api/` ke `BACKEND_URL/` dengan menghapus awalan `/api/`. Request `/api/surat` diteruskan ke `/surat`, dan lampiran `/api/uploads/...` diteruskan ke `/uploads/...`. Alamat upstream di template Nginx berasal dari environment.

Perubahan `VITE_API_URL` memerlukan rebuild image. Perubahan `BACKEND_URL` memerlukan recreate container frontend. File `.env*` lokal tidak disalin ke image build.
