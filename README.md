# 🗄️ Laci DOSCOM

Halo Sobat DOSCOM! 👋 Welcome ke repositori **Laci**.
Kalau kamu anak baru di divisi _development_ atau sekadar mau _contribute_, _you are in the right place, bro!_

## 🤔 Laci Tuh Apa Sih?

Singkatnya, **Laci** adalah aplikasi web buat bantuin Sekretaris dan pengurus DOSCOM biar nggak pusing ngurusin surat-menyurat dan program kerja (Proker). Di sini kita bisa:

- Pantau status surat pakai **Kanban Board** yang _smooth_.
- Filter daftar surat dengan UI _rounded_ kekinian ala iOS.
- Ngatur klasifikasi surat dan proker tanpa ribet.

Kita bikin ini se-ringan mungkin. _No_ UI _slop_, _no lag_ (bahkan di Firefox sekalipun), pokoknya _blazing fast_! ⚡

## 🛠️ Senjata Kita (Tech Stack)

Biar _stay relevant_, kita pakai _stack_ modern:

- **Frontend:** Vue 3 (Composition API) + Tailwind CSS + Vite
- **Backend:** Golang (Kecil, cepat, dan _powerful_)
- **Database:** PostgreSQL 15
- **DevOps:** Docker & Docker Compose

## 🚀 Cara Gas Lokal (Quick Start)

Paling gampang nge-jalanin repo ini tuh pakai **Docker**. Nggak perlu instal Node.js, Go, atau Postgres secara manual.

### Langkah-langkah

1. **Clone dulu reponya:**

   ```bash
   git clone https://github.com/doscom/laci.git
   cd laci
   ```

2. **Mantra sakti Docker:**

   ```bash
   cp .env.example .env
   docker compose up -d --build
   ```

   _(Tungguin aja, Docker lagi sibuk nge-build Vue & Go di belakang layar)._

3. **Cekidot!**
   Buka browser lu dan gas ke 👉 **`http://localhost:5173`**

> **💡 Kenapa port 5173?**
> Secara _default_, web server berjalan di port `80`. Tapi karena kebanyakan anak DOSCOM pakai **XAMPP** (yang membajak port 80 untuk Apache), kita _mapping_ frontend Laci ke port `5173` biar bebas hambatan dan terasa seperti _environment dev_ Vue asli!

## 💻 Panduan Buat Ngedit (Dev Mode)

Kalau lu pengen ngoprek kodenya langsung dan butuh _Hot-Reload_:

1. **Nyalain Database-nya doang:**

   ```bash
   docker compose up db -d
   ```

2. **Nyalain Backend (Golang):**
   Masuk ke folder `backend` dan ketik:

   ```bash
   go mod tidy
   go run .
   ```

   _(Backend bakal jalan di port `8080`)_

3. **Nyalain Frontend (Vue 3):**
   Masuk folder `frontend` dan ketik:
   ```bash
   cp .env.example .env
   npm ci
   npm run dev
   ```

URL integrasi backend diatur melalui `VITE_API_URL`, tanpa hostname atau port bawaan di kode frontend. Lihat [panduan environment frontend](frontend/README.md) untuk konfigurasi lokal, Docker, dan production.

## 🧪 Inject Dummy Data (Buat Testing)

Aplikasi kosong emang nggak asik. Kalo lu butuh _dummy data_ (50 biji), sikat _query_ ini! Masuk ke kontainer database dulu:

```bash
docker compose exec db psql -U root -d laci_doscom
```

Lalu, _copy-paste_ SQL barbar ini:

```sql
TRUNCATE TABLE surat CASCADE;

INSERT INTO surat (id, nomor_surat, perihal, pic_nama, status_saat_ini)
SELECT
    gen_random_uuid(),
    LPAD(i::text, 3, '0') || '/DOSCOM/DU/X/2026',
    'Surat Permohonan Dummy Ke-' || i,
    (ARRAY['Budi', 'Citra', 'Dimas', 'Eka', 'Gilang', 'Rina', 'Siti', 'Bagas'])[floor(random() * 8 + 1)],
    (ARRAY['Standby', 'Cetak', 'TTD Lapis 1', 'TTD Ketum', 'TTD Pembina', 'Paraf Koormawa', 'TTD Tertinggi', 'Selesai'])[floor(random() * 8 + 1)]
FROM generate_series(1, 50) AS s(i);
```

Ketik `\q` buat keluar, lalu _refresh_ browser lu!

## 🤝 Rules of Play (Kontribusi)

Kita sangat _open_ buat PR (Pull Request)!

1. Bikin _branch_ baru (misal: `feat/kanban-animation`).
2. _Commit_ dengan pesan yang jelas.
3. Bikin _Pull Request_ ke _branch_ `main`.

**Dibuat dengan 💻, ☕, dan 💖 oleh divisi Pemro DOSCOM.**
