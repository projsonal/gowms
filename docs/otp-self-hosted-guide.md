# Panduan: OTP Self-Hosted (Next.js → Backend API → OTP Service → WhatsApp/SMS Provider)

Diagram yang kamu gambar sudah **persis** menggambarkan arsitektur yang dipakai di project ini —
dokumen ini memetakan tiap kotak diagram ke file/kode aslinya, supaya jelas di mana logikanya
dan bagaimana menggantinya kalau perlu.

```
Next.js  →  Backend API Route  →  OTP Service  →  WhatsApp/SMS Provider
   |                                                        |
   └──────────────── user masukkan kode ke form ◄───────────┘
```

## 1. Next.js (frontend) — memicu permintaan kode

File: `src/component/auth/ForgotPasswordStep.tsx`, `RegisterPhoneVerifyStep.tsx`, dll.

```ts
// src/lib/api/auth.ts
requestPasswordReset: (payload) =>
  apiClient.post('/auth/password/forgot', payload, { skipAuth: true }),
```

Next.js **tidak pernah** bicara langsung ke WhatsApp/SMS provider. Semua lewat backend Go —
supaya API key gateway tidak pernah terekspos ke browser.

## 2. Backend API Route — endpoint yang menerima permintaan

File: `internal/controller/auth/auth_controller.go` (fungsi `ForgotPassword`,
`RequestRegisterOTP`, dll).

Tugasnya:
1. Validasi siapa yang minta (user ada? nomor HP ada?).
2. Panggil **OTP Service** untuk generate kode.
3. Panggil **Provider** (WhatsApp/SMS) untuk kirim kode itu.
4. Balas ke frontend token referensi (`otp_token`) — BUKAN kodenya sendiri.

```go
code, otpToken, err := h.resetOTPSvc.Generate()   // (3) OTP Service
...
sendErr = h.smsSender.SendOTP(u.PhoneNumber, code) // (4) Provider
```

## 3. OTP Service — generate & verifikasi kode

File: `pkg/otp/otp.go` + `pkg/otp/store.go`.

Ini bagian yang **kamu buat sendiri** (bukan library pihak ketiga) — logikanya simpel & sudah
ada di project:

```go
// pkg/otp/otp.go (ringkasan)
func (s *Service) Generate() (code, token string, err error) {
    code = randomDigits(6)                 // "482913"
    token = randomToken()                  // referensi acak, dikirim ke frontend
    s.store.Save(token, hash(code), s.ttl)  // simpan HASH kodenya, bukan kode mentah
    return code, token, nil
}

func (s *Service) Verify(token, code string) error {
    saved, err := s.store.Get(token)
    if err != nil { return ErrExpired }
    if saved.used { return ErrAlreadyUsed }   // cegah kode dipakai ulang
    if !compareHash(saved.hash, code) { return ErrWrongCode }
    s.store.MarkUsed(token)
    return nil
}
```

Poin penting kalau kamu mau bikin ulang dari nol:
- **Jangan simpan kode mentah** — simpan hash-nya (mis. SHA-256), sama seperti password.
- **TTL pendek** (project ini pakai 5 menit, lihat `WA_OTP_TTL_MINUTES`).
- **Tandai "sudah dipakai"** setelah verifikasi sukses sekali — mencegah kode yang sama dipakai
  berkali-kali (replay attack).
- **Rate-limit** permintaan generate kode per-IP/per-user (lihat
  `internal/middleware/rate_limit_middleware.go`) — tanpa ini, endpoint OTP bisa dipakai spam
  SMS/WA ke nomor orang lain (dan menghabiskan kuota kamu).

## 4. WhatsApp/SMS Provider — pengiriman sungguhan

Ini bagian yang paling sering bikin bingung karena **dua pilihan berbeda arsitektur**:

### Opsi A — Gateway pihak ketiga (prabayar/berlangganan)
File: `pkg/wa/wa.go`, `pkg/sms/sms.go`.

Cukup HTTP POST ke API gateway (Fonnte, Wablas, Zenziva, Twilio, dst) dengan API key mereka.
Simpel, tapi **berbayar per pesan/bulan** — ini yang kamu ingin hindari.

### Opsi B — Self-hosted, GRATIS, pakai nomor WhatsApp sendiri
File: `pkg/wa/whatsmeow_sender.go` + `cmd/whatsapp-pair/main.go`.

Ini **jawaban untuk "saya ingin buat sendiri"**. Caranya:

1. **Sekali saja**, pasangkan nomor WhatsApp kamu sebagai "device" (persis seperti WhatsApp Web):
   ```bash
   go get go.mau.fi/whatsmeow@latest
   go get github.com/mdp/qrterminal/v3@latest
   go get modernc.org/sqlite@latest
   go mod tidy
   go run ./cmd/whatsapp-pair
   ```
   Scan QR yang muncul di terminal pakai HP (WhatsApp → Perangkat Tertaut → Tautkan Perangkat).
   Sesi login tersimpan di `./var/whatsmeow-session.db`.

2. Set di `.env`:
   ```
   WHATSAPP_DRIVER=whatsmeow
   ```

3. Selesai. Server otomatis connect ulang pakai sesi tersimpan setiap kali start — **tidak perlu
   scan ulang**, tidak ada biaya per pesan, tidak ada API key pihak ketiga.

Kenapa ini bisa gratis: `whatsmeow` adalah implementasi protokol WhatsApp Web di Go — persis
seperti kamu buka `web.whatsapp.com` di browser dan scan QR, tapi dijalankan oleh server kamu.
Pesan yang terkirim akan tampak "dari" nomor WhatsApp kamu sendiri, sama seperti kirim manual.

**Trade-off yang perlu kamu tahu:**
- Nomor itu harus tetap aktif/online (server harus tetap jalan) — kalau logout dari WhatsApp
  di HP, sesi ini ikut putus dan perlu pairing ulang.
- Ini "sesi personal", bukan WhatsApp Business API resmi — cocok untuk skala kecil-menengah
  (OTP internal tim/perusahaan), tapi WhatsApp bisa membatasi nomor yang mengirim pesan otomatis
  dalam volume besar. Untuk skala besar (ribuan OTP/hari), gateway resmi (Opsi A) atau WhatsApp
  Business API tetap lebih aman dari risiko pemblokiran.

### SMS gratis?
Untuk SMS **tidak ada** ekuivalen "self-hosted gratis" seperti whatsmeow — SMS selalu lewat
jaringan operator seluler yang berbayar. Kalau anggaran jadi masalah, sarankan pengguna pakai
opsi WhatsApp (gratis lewat whatsmeow) sebagai default, dan SMS sebagai fallback opsional saja.

## Ringkasan keputusan konfigurasi

| Situasi | Setting `.env` |
|---|---|
| Belum ada budget gateway sama sekali | `WHATSAPP_DRIVER=whatsmeow` + jalankan `cmd/whatsapp-pair` sekali |
| Sudah langganan gateway (Fonnte/dst) | `WHATSAPP_DRIVER=gateway` + isi `WHATSAPP_API_URL`/`WHATSAPP_API_KEY` |
| Butuh SMS juga | Isi `SMS_API_URL`/`SMS_API_KEY` (tidak ada opsi gratis untuk SMS) |

Cara cek konfigurasi mana yang aktif: lihat **log saat server startup** — ada baris
`PERINGATAN: ...` eksplisit kalau ada channel yang belum siap (lihat
`internal/routes/wa_driver.go`).
