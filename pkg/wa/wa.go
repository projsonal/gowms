package wa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

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
		return errors.New("whatsapp: WHATSAPP_API_URL belum dikonfigurasi")
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
		return fmt.Errorf("whatsapp: gagal menghubungi gateway: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("whatsapp: gateway membalas status %d", resp.StatusCode)
	}
	return nil
}
