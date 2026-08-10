package routes

import (
	"context"
	"log"

	"github.com/projsonal/gowms/pkg/config"
	"github.com/projsonal/gowms/pkg/wa"
)

type errSender struct{ reason string }

func (e errSender) SendOTP(string, string) error {
	return errWithPrefix("whatsapp belum siap: ", e.reason)
}

func errWithPrefix(prefix, msg string) error {
	return &prefixedError{prefix: prefix, msg: msg}
}

type prefixedError struct {
	prefix string
	msg    string
}

func (e *prefixedError) Error() string { return e.prefix + e.msg }

func buildWhatsAppSender(cfg *config.Config) wa.Sender {
	if cfg.WhatsApp.Driver == "whatsmeow" {
		sender, err := wa.NewWhatsmeowSender(context.Background(), cfg.WhatsApp.SessionPath)
		if err != nil {
			log.Printf("PERINGATAN: WHATSAPP_DRIVER=whatsmeow tapi gagal disiapkan (%v) — "+
				"pengiriman OTP WhatsApp akan SELALU GAGAL sampai ini diperbaiki. "+
				"Jalankan `go run ./cmd/whatsapp-pair` untuk memasangkan sesi.", err)
			return errSender{reason: err.Error()}
		}
		return sender
	}

	if cfg.WhatsApp.APIURL == "" {
		log.Println("PERINGATAN: WHATSAPP_API_URL kosong (driver=gateway) — pengiriman OTP " +
			"WhatsApp akan SELALU GAGAL sampai diisi, atau set WHATSAPP_DRIVER=whatsmeow " +
			"untuk pakai sesi WhatsApp sendiri (lihat pkg/wa/whatsmeow_sender.go).")
	}
	return wa.NewClient(cfg.WhatsApp.APIURL, cfg.WhatsApp.APIKey, cfg.WhatsApp.Sender)
}
