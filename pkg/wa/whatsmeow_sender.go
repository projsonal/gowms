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

	_ "modernc.org/sqlite"
)

const whatsmeowSendTimeout = 15 * time.Second

type WhatsmeowSender struct {
	client *whatsmeow.Client
}

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

func (s *WhatsmeowSender) Close() {
	if s.client != nil {
		s.client.Disconnect()
	}
}

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
