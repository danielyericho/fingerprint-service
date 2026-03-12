# fingerprint-service

Service berbasis Go untuk Windows yang mengintegrasikan **ZKFinger SDK** (ZKTeco) via COM/ActiveX. Service mengembalikan data fingerprint dalam bentuk string dan menyediakan pencocokan 1:1 (verify) dan 1:N (identify) menggunakan fitur SDK; data template untuk matching dikirim oleh pemanggil (tidak ada database di dalam service).

**Target:** Windows only. **Pendekatan:** Go + go-ole (COM).

## Persyaratan

- **Windows** (driver ZKFinger dan sensor terpasang)
- **Go 1.17+** (disarankan 1.21+)
- **ZKFinger SDK / driver** terinstal (setup dari folder Fingerprint; komponen COM/OCX terdaftar, mis. ZKFPEngX / Biokey.ocx)

## Instalasi driver

1. Jalankan setup dari folder SDK (mis. `C:\Users\...\Fingerprint\en\` atau `chs\`).
2. Pastikan sensor fingerprint terhubung dan dikenali di Device Manager.
3. Demo C++ (Demo.exe) dari SDK bisa dipakai untuk memastikan sensor berfungsi.

## Build

Pastikan **GOROOT** mengarah ke instalasi Go yang benar (path std lib harus valid, mis. `GOROOT\src\encoding\json`, bukan `GOROOT\src\src\...`). Jika build gagal dengan "package X is not in std", perbaiki instalasi Go atau set `GOROOT` ke folder Go yang benar. Kemudian:

```bash
cd D:\Education\programming\golang\fingerprint-service
go mod tidy
go build -o bin/server.exe ./cmd/server
go build -o bin/cli.exe ./cmd/cli
```

Build hanya bermakna di Windows (package `internal/zkfp` memakai go-ole dan API Windows).

## Menjalankan server

```bash
bin\server.exe -addr :8080
```

Endpoint:

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET/POST | `/capture` | Baca satu jari dari hardware; mengembalikan `template9` dan `template10` (string). Query: `timeout_sec` (default 30). |
| POST | `/enroll` | Enroll (beberapa kali tekan). Query: `presses` (default 3), `timeout_sec` (default 60). Return template string. |
| POST | `/verify` | Verifikasi 1:1. Body JSON: `registered_template`, `verification_template`, optional `do_learning`. Return `match`, `score`. |
| POST | `/identify` | Identifikasi 1:N. Body JSON: `templates` (array of `{ id, template9, template10 }`), `verification_template`. Return `matched_id` (-1 jika tidak ada), `score`, `processed`. |
| GET | `/health` | Health check. |

## CLI (tes tanpa HTTP)

```bash
# Baca satu jari (return template string)
bin\cli.exe -cmd capture -timeout 30

# Enroll 3 kali tekan
bin\cli.exe -cmd enroll -presses 3 -timeout 60

# Verifikasi 1:1 (butuh dua template string)
bin\cli.exe -cmd verify -reg "<registered_template>" -ver "<verification_template>"

# Identifikasi 1:N (butuh file JSON templates + verification template)
bin\cli.exe -cmd identify -templates templates.json -ver "<verification_template>"
```

Format `templates.json` untuk identify:

```json
[
  { "id": 1, "template9": "...", "template10": "" },
  { "id": 2, "template9": "...", "template10": "..." }
]
```

## Arsitektur ringkas

- **Baca:** Hardware → COM (ZKFPEngX) → service → response JSON berisi template string. Pemanggil (mis. HMS) menyimpan string ke DB mereka.
- **Verify/Identify:** Pemanggil mengirim template (dari DB mereka) di request; service memanggil SDK (VerFingerFromStr, IdentificationFromStrInFPCacheDB) dan mengembalikan hasil match.

Tidak ada database di dalam service ini.

## Struktur repo

- `cmd/server` — HTTP API server
- `cmd/cli` — CLI untuk capture, enroll, verify, identify
- `internal/zkfp` — Wrapper COM (go-ole) untuk ZKFPEngX; build tag `windows`
- `internal/api` — HTTP handler dan implementasi service
- `pkg/fingerprint` — Tipe request/response dan interface service
- `sdk-reference/` — Salinan referensi SDK (header, sample C++); driver/installer tetap diinstall dari paket asli

## Catatan

- **Message pump:** COM/ActiveX membutuhkan message loop Windows agar event OnCapture/OnEnroll bisa terkirim. Implementasi saat ini memakai polling + message loop singkat; untuk production dengan event penuh mungkin perlu event sink.
- **ProgID:** Default `ZKFPEngX.ZKFPEngX`. Jika setup SDK mendaftarkan CLSID dengan nama lain, sesuaikan di kode (NewEngine(progID)).
- **Go 1.17:** `go.mod` memakai `go 1.17`; tetap kompatibel dengan Go yang lebih baru.
