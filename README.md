# fingerprint-service

Service berbasis Go untuk Windows yang mengintegrasikan **ZKFinger SDK** (ZKTeco) via COM/ActiveX. Service mengembalikan data fingerprint dalam bentuk string dan menyediakan pencocokan 1:1 (verify) dan 1:N (identify) menggunakan fitur SDK; data template untuk matching dikirim oleh pemanggil (tidak ada database di dalam service).

**Target:** Windows 7 32-bit ke atas (termasuk Win10/Win11 via WOW64). **Pendekatan:** Go + go-ole (COM).

## Persyaratan

- **Windows 7 32-bit** atau lebih baru (Win10/Win11 64-bit menjalankan binary 32-bit via WOW64)
- **Go 1.17.x** (untuk build dari sumber; sudah diverifikasi jalan di Win7 32-bit)
- **ZKFinger SDK / driver** terinstal (setup dari folder Fingerprint; komponen COM/OCX terdaftar, mis. ZKFPEngX / Biokey.ocx)

## Instalasi driver

1. Jalankan setup dari folder SDK (mis. `C:\Users\...\Fingerprint\en\` atau `chs\`).
2. Pastikan sensor fingerprint terhubung dan dikenali di Device Manager.
3. Demo C++ (Demo.exe) dari SDK bisa dipakai untuk memastikan sensor berfungsi.

## Build

Pastikan **GOROOT** mengarah ke instalasi Go yang benar (path std lib harus valid, mis. `GOROOT\src\encoding\json`, bukan `GOROOT\src\src\...`). Jika build gagal dengan "package X is not in std", perbaiki instalasi Go atau set `GOROOT` ke folder Go yang benar.

### Build dengan script (disarankan)

Script `build.sh` mem-build **binary 32-bit** (`GOARCH=386`) — satu executable untuk Win7 32-bit dan Win10/Win11 64-bit — menyematkan build version, lalu membuat ZIP installer di `dist/`.

```bash
cd D:\Education\programming\golang\fingerprint-service
chmod +x build.sh   # sekali saja di Git Bash / WSL
./build.sh
```

Opsi version tanpa prompt interaktif:

```bash
./build.sh v1.0
# atau
VERSION=v1.0 ./build.sh
```

Ubah lokasi output ZIP (default: `dist/` di root proyek):

```bash
OUTPUT_DIR=/c/Users/Dream/Downloads ./build.sh v1.0
```

Cek version setelah build:

```bash
bin\fingerprint-service.exe -version
bin\cli.exe -version
```

### Build manual

```bash
cd D:\Education\programming\golang\fingerprint-service
go mod tidy
set GOOS=windows
set GOARCH=386
go build -ldflags "-X main.version=v1.0" -o bin\fingerprint-service.exe ./cmd/server
go build -ldflags "-X main.version=v1.0" -o bin\cli.exe ./cmd/cli
```

Build hanya bermakna di Windows (package `internal/zkfp` memakai go-ole dan API Windows).

## Menjalankan server (console)

```bash
bin\fingerprint-service.exe -addr :8083
```

## Menjalankan sebagai Windows Service

Service name (internal): **FingerprintQubuService**  
Display name: **Fingerprint Qubu Service**

Aplikasi mendukung dua mode:

1. **Native Windows SCM** — jika dijalankan langsung oleh Service Control Manager, proses mendeteksi konteks service dan memakai `golang.org/x/sys/windows/svc`.
2. **NSSM wrapper** — disarankan untuk instalasi production (mirip `orbita-lock-door`), dengan script di folder `bin\`.

### Instalasi via NSSM (disarankan)

Sebelum menjalankan `install.bat`, pastikan folder `bin\` berisi:

- `fingerprint-service.exe` — binary hasil build
- `nssm.exe` — service wrapper

Jika `nssm.exe` belum ada, jalankan dulu:

```bash
cd bin
download-nssm.bat
```

Script itu akan mencoba unduh otomatis (URL CI NSSM atau `winget install NSSM.NSSM`). Alternatif manual:

```powershell
winget install NSSM.NSSM
copy "$env:LOCALAPPDATA\Microsoft\WinGet\Links\nssm.exe" bin\nssm.exe
```

Jalankan sebagai **Administrator**:

```bash
cd bin
install.bat
```

Script lain:

| Script | Fungsi |
|--------|--------|
| `install.bat` | Install service via NSSM, konfigurasi log rotasi, lalu start |
| `start.bat` | Start service yang sudah terinstall |
| `stop.bat` | Stop service (graceful Ctrl+C dulu) |
| `uninstall.bat` | Stop dan hapus service |

### Log service

NSSM menangkap stdout/stderr ke:

| File | Isi |
|------|-----|
| `bin\logs\fingerprint.err.log` | Log utama (startup, shutdown, error) |
| `bin\logs\fingerprint.out.log` | Output stdout (jarang dipakai) |

Live-tail di PowerShell:

```powershell
Get-Content -Path .\bin\logs\fingerprint.err.log -Wait -Tail 50
```

### Catatan service

- Jalankan script install/start/stop/uninstall sebagai **Administrator**.
- Akses sensor fingerprint/COM sering membutuhkan hak admin atau service account yang sesuai.
- Saat dijalankan sebagai service, working directory otomatis dipindah ke folder executable agar path relatif konsisten.

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

- `cmd/server` — HTTP API server (termasuk dukungan Windows Service)
- `bin/` — Script install/start/stop/uninstall service (NSSM) dan output binary
- `cmd/cli` — CLI untuk capture, enroll, verify, identify
- `internal/zkfp` — Wrapper COM (go-ole) untuk ZKFPEngX; build tag `windows`
- `internal/api` — HTTP handler dan implementasi service
- `pkg/fingerprint` — Tipe request/response dan interface service
- `sdk-reference/` — Salinan referensi SDK (header, sample C++); driver/installer tetap diinstall dari paket asli

## Catatan

- **Message pump:** COM/ActiveX membutuhkan message loop Windows agar event OnCapture/OnEnroll bisa terkirim. Implementasi saat ini memakai polling + message loop singkat; untuk production dengan event penuh mungkin perlu event sink.
- **ProgID:** Default `ZKFPEngX.ZKFPEngX`. Jika setup SDK mendaftarkan CLSID dengan nama lain, sesuaikan di kode (NewEngine(progID)).
- **Go 1.17:** `go.mod` memakai `go 1.17`; build 32-bit (`GOARCH=386`) agar satu binary jalan di Win7 32-bit dan Win10/11 64-bit.
