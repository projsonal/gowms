// Package wa (file ini) menyediakan implementasi Sender ALTERNATIF yang
// tidak butuh gateway WhatsApp berbayar (Fonnte/Wablas/Zenziva dkk).
//
// LATAR BELAKANG DEBUGGING: kalau kode OTP "tidak pernah sampai" padahal
// endpoint /auth/password/forgot membalas sukses (200) atau error jelas
// (502 "gagal mengirim kode OTP lewat WhatsApp"), akar masalahnya HAMPIR
// SELALU env var WHATSAPP_API_URL / SMS_API_URL belum diisi — wa.Client
// yang lama (lihat wa.go) memang SENGAJA gagal kalau kosong (lihat
// errors.New("whatsapp: WHATSAPP_API_URL belum dikonfigurasi")), supaya
// error-nya jelas alih-alih mengaku sukses padahal tidak mengirim apa-apa.
// Kalau tidak punya akses ke gateway berbayar, WhatsmeowSender di file ini
// adalah alternatifnya: login SEKALI pakai nomor WhatsApp sendiri (scan QR
// lewat cmd/whatsapp-pair), sesi tersimpan di file SQLite, lalu proses
// utama otomatis konek ulang pakai sesi itu setiap start — tanpa gateway
// pihak ketiga sama sekali. Ini memakai whatsmeow, package Go yang jadi
// MESIN di balik github.com/whatsauth/whatsauth (whatsauth membungkus
// whatsmeow dengan arsitektur hub/websocket terpisah untuk multi-tenant;
// untuk satu akun WA pengirim OTP saja, whatsmeow langsung lebih ringkas
// & lebih gampang di-debug).
//
// CARA PAKAI:
//  1. Jalankan sekali: `go run ./cmd/whatsapp-pair` lalu scan QR yang
//     muncul di terminal pakai HP yang jadi pengirim OTP (WhatsApp ->
//     Perangkat Tertaut -> Tautkan Perangkat). Sesi tersimpan otomatis di
//     WHATSMEOW_SESSION_PATH (default ./var/whatsmeow-session.db).
//  2. Set WHATSAPP_DRIVER=whatsmeow di .env.
//  3. Jalankan server seperti biasa — sesi yang sudah dipasangkan dipakai
//     otomatis, TIDAK perlu scan ulang tiap restart.
//
// DEPENDENSI: package ini butuh modul yang belum ada di go.mod. Jalankan
// sebelum build:
//
//	go get go.mau.fi/whatsmeow@latest
//	go get modernc.org/sqlite@latest   # driver SQLite pure-Go, tanpa CGO
//	go mod tidy
package wa

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	_ "modernc.org/sqlite" // driver database/sql "sqlite" — pure Go, tidak butuh CGO
)

const whatsmeowSendTimeout = 15 * time.Second

// WhatsmeowSender mengirim OTP lewat sesi WhatsApp Web milik akun sendiri
// (sudah dipasangkan sebelumnya lewat cmd/whatsapp-pair). Implements Sender.
type WhatsmeowSender struct {
	client *whatsmeow.Client
}

// NewWhatsmeowSender membuka sesi tersimpan di sessionPath & langsung
// connect. Kalau belum pernah dipasangkan (device belum ada), fungsi ini
// mengembalikan error yang jelas — jangan sampai server nyala tapi diam-
// diam tidak pernah bisa mengirim WhatsApp.
func NewWhatsmeowSender(ctx context.Context, sessionPath string) (*WhatsmeowSender, error) {
	dbLog := waLog.Stdout("whatsmeow/db", "WARN", true)
	container, err := sqlstore.New(ctx, "sqlite", "file:"+sessionPath+"?_pragma=foreign_keys(1)", dbLog)
	if err != nil {
		return nil, fmt.Errorf("wa: gagal membuka session store whatsmeow di %s: %w", sessionPath, err)
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("wa: gagal membaca device whatsmeow: %w", err)
	}
	if device == nil || device.ID == nil {
		return nil, fmt.Errorf(
			"wa: belum ada sesi WhatsApp yang dipasangkan di %s — jalankan `go run ./cmd/whatsapp-pair` dulu untuk scan QR",
			sessionPath,
		)
	}

	client := whatsmeow.NewClient(device, waLog.Stdout("whatsmeow", "WARN", true))
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("wa: gagal connect ke WhatsApp dengan sesi tersimpan: %w", err)
	}

	log.Println("wa: whatsmeow terhubung memakai sesi WhatsApp yang sudah dipasangkan")
	return &WhatsmeowSender{client: client}, nil
}

// Close menutup koneksi whatsmeow — panggil saat aplikasi shutdown.
func (s *WhatsmeowSender) Close() {
	if s.client != nil {
		s.client.Disconnect()
	}
}

// normalizePhoneToJID mengubah nomor format E.164 (mis. "+6281234567890")
// menjadi JID WhatsApp ("6281234567890@s.whatsapp.net").
func normalizePhoneToJID(phoneNumber string) (types.JID, error) {
	digits := strings.TrimPrefix(strings.TrimSpace(phoneNumber), "+")
	if digits == "" {
		return types.JID{}, errors.New("wa: nomor HP kosong")
	}
	return types.NewJID(digits, types.DefaultUserServer), nil
}

func buildOTPMessage(code string) *waE2E.Message {
	text := fmt.Sprintf("Kode verifikasi StockRSD Anda: %s (berlaku singkat, jangan bagikan ke siapa pun).", code)
	return &waE2E.Message{Conversation: proto.String(text)}
}

func (s *WhatsmeowSender) SendOTP(phoneNumber, code string) error {
	if s.client == nil || !s.client.IsConnected() {
		return errors.New("wa: sesi whatsmeow belum terhubung — cek apakah device masih ter-link (WhatsApp di HP -> Perangkat Tertaut)")
	}
	jid, err := normalizePhoneToJID(phoneNumber)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), whatsmeowSendTimeout)
	defer cancel()
	if _, err := s.client.SendMessage(ctx, jid, buildOTPMessage(code)); err != nil {
		return fmt.Errorf("wa: whatsmeow gagal mengirim pesan: %w", err)
	}
	return nil
}
