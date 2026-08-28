# Catatan Debugging Build

## Yang sudah dicek dan AMAN (bukan penyebab error)

1. **Syntax code** — semua file `.go` di-scan dengan `gofmt`; tidak ada
   syntax error. 4 file kecil hanya butuh format ulang (whitespace/newline),
   sudah diperbaiki di commit ini:
   - `internal/repositories/notification/notification_repos.go`
   - `internal/controller/asset_gudang/ping_controller.go`
   - `internal/controller/dashboard/struct.go`
   - `pkg/reportexport/docs.go`

2. **Duplicate declaration** — dicek semua package untuk func/type yang
   ke-declare dua kali dalam package yang sama. Hanya ditemukan satu
   "duplikat" semu: `detachProcess` di `launch_update_windows.go` dan
   `launch_update_unix.go` — ini AMAN karena keduanya saling eksklusif
   lewat Go build tags (`_windows.go` suffix dan `//go:build !windows`).

3. **go.sum** — sudah lengkap untuk semua dependency termasuk
   `go.mau.fi/whatsmeow`, `github.com/mdp/qrterminal/v3`, dan
   `modernc.org/sqlite`. Komentar lama di `go.mod` yang bilang dependensi
   ini "belum ter-resolve" sudah usang — sudah dibersihkan/diperbarui.

## Kemungkinan penyebab utama "gagal build"

`go.mod` mensyaratkan `go 1.26.5` — ini BUKAN kesalahan penulisan, tapi
konsekuensi nyata dari salah satu dependency langsung:

- `github.com/xuri/excelize/v2 v2.11.0` mensyaratkan `go >= 1.25.0` di
  `go.mod`-nya sendiri.

Kode project ini sendiri **tidak** memakai fitur bahasa khusus Go 1.26
(sudah dicek: tidak ada pemakaian `simd/archsimd`, `crypto/hpke`,
`runtime/secret`, atau sintaks `new(expr)` ala Go 1.26).

### Skenario paling mungkin

Kalau `go build ./...` gagal dengan pesan seperti:

```
go: go.mod requires go >= 1.26.5 (running go X.Y.Z; GOTOOLCHAIN=...)
```

atau kalau `GOTOOLCHAIN=auto` mencoba download otomatis lalu gagal dengan
error jaringan/403/timeout ke `proxy.golang.org` — itu tandanya:

1. Go lokal kamu versinya lebih lama dari 1.25/1.26, **dan**
2. mesin/server/CI tempat build tidak punya akses internet ke
   `proxy.golang.org` untuk auto-download toolchain yang benar.

### Cara memperbaiki (pilih salah satu)

**A. Install Go 1.26.5+ secara manual** (paling simpel kalau ada internet):
```bash
# unduh dari https://go.dev/dl/ sesuai OS/arch, lalu:
go version   # pastikan >= 1.26.5
go build ./...
```

**B. Vendor dependency-nya** (kalau server build tidak ada akses internet
sama sekali) — jalankan ini SEKALI di mesin yang ada internet & Go 1.26+:
```bash
go mod vendor
git add vendor
git commit -m "chore: vendor dependencies for offline build"
```
Setelah itu build di server offline pakai:
```bash
go build -mod=vendor ./...
```

**C. Kalau tidak butuh fitur WhatsApp via whatsmeow** (driver default
project ini adalah "gateway" HTTP, bukan whatsmeow), dan ingin mengurangi
dependency berat, hapus:
- `pkg/wa/whatsmeow_sender.go`
- `cmd/whatsapp-pair/main.go`

lalu jalankan `go mod tidy` untuk membersihkan `go.mod`/`go.sum` dari
dependency `whatsmeow`, `modernc.org/sqlite`, dan `qrterminal` yang jadi
tidak terpakai.

## Kalau masih gagal setelah ini

Tolong jalankan dan kirim hasil lengkapnya:
```bash
go version
go env GOTOOLCHAIN GOPROXY
go build ./... 2>&1
```
Pesan error persis (bukan ringkasan) sangat membantu untuk diagnosis lebih
lanjut — banyak error Go yang terlihat mirip tapi akar masalahnya beda
(missing go.sum entry, CGO, mismatched GOOS/GOARCH, dll).
