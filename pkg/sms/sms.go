// Package sms mengirim kode OTP lewat gateway SMS. Strukturnya sengaja
// dibuat mirip pkg/wa (WhatsApp) supaya kedua kanal pengiriman OTP
// (WhatsApp & SMS) punya bentuk pemakaian yang konsisten di controller.
package sms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Sender mengirim kode OTP lewat SMS ke nomor HP tertentu.
type Sender interface {
	SendOTP(phoneNumber, code string) error
}

type Client struct {
	apiURL     string
	apiKey     string
	senderName string
	http       *http.Client
}

func NewClient(apiURL, apiKey, senderName string) *Client {
	return &Client{
		apiURL:     apiURL,
		apiKey:     apiKey,
		senderName: senderName,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

type sendRequest struct {
	Target  string `json:"target"`
	Message string `json:"message"`
	Sender  string `json:"sender,omitempty"`
}

func (c *Client) SendOTP(phoneNumber, code string) error {
	if c.apiURL == "" {
		return errors.New("sms: SMS_API_URL belum dikonfigurasi")
	}

	body, err := json.Marshal(sendRequest{
		Target:  phoneNumber,
		Message: fmt.Sprintf("Kode verifikasi StockRSD Anda: %s (berlaku singkat, jangan bagikan ke siapa pun).", code),
		Sender:  c.senderName,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("sms: gagal menghubungi gateway: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("sms: gateway membalas status %d", resp.StatusCode)
	}
	return nil
}
