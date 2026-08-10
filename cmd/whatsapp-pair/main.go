package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "modernc.org/sqlite"
)

func main() {
	_ = godotenv.Load()

	sessionPath := os.Getenv("WHATSMEOW_SESSION_PATH")
	if sessionPath == "" {
		sessionPath = "./var/whatsmeow-session.db"
	}

	ctx := context.Background()
	dbLog := waLog.Stdout("whatsmeow/db", "INFO", true)
	container, err := sqlstore.New(ctx, "sqlite", "file:"+sessionPath+"?_pragma=foreign_keys(1)", dbLog)
	if err != nil {
		log.Fatalf("gagal membuka session store di %s: %v", sessionPath, err)
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		log.Fatalf("gagal membaca device: %v", err)
	}

	client := whatsmeow.NewClient(device, waLog.Stdout("whatsmeow", "INFO", true))

	if client.Store.ID != nil {
		fmt.Println("Sudah ada sesi WhatsApp yang dipasangkan sebelumnya di", sessionPath)
		fmt.Println("Kalau mau pasang ulang (mis. device di-unlink dari HP), hapus dulu file tsb lalu jalankan ulang.")
		return
	}

	qrChan, _ := client.GetQRChannel(ctx)
	if err := client.Connect(); err != nil {
		log.Fatalf("gagal connect: %v", err)
	}

	for evt := range qrChan {
		switch evt.Event {
		case "code":
			fmt.Println("Scan QR ini pakai WhatsApp -> Perangkat Tertaut -> Tautkan Perangkat:")
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
		case "success":
			fmt.Println("Berhasil dipasangkan! Sesi tersimpan di", sessionPath)
			fmt.Println("Set WHATSAPP_DRIVER=whatsmeow di .env lalu jalankan server seperti biasa.")
		case "timeout":
			log.Fatal("waktu scan QR habis, jalankan ulang perintah ini untuk coba lagi")
		}
	}
}
