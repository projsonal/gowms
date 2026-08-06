package geoip

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"
)

const unknownLocation = "-"
const defaultTimeout = 2 * time.Second

type Resolver interface {
	Lookup(ip string) (location string, err error)
}

type NoopResolver struct{}

func (NoopResolver) Lookup(string) (string, error) { return unknownLocation, nil }

type httpResolver struct {
	client  *http.Client
	baseURL string
}

func NewHTTPResolver(baseURL string) Resolver {
	return &httpResolver{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: baseURL,
	}
}

type ipAPIResponse struct {
	Status  string `json:"status"`
	City    string `json:"city"`
	Country string `json:"country"`
}

func (r *httpResolver) Lookup(ip string) (string, error) {
	if !isPublicIP(ip) {
		return unknownLocation, nil
	}

	resp, err := r.client.Get(r.baseURL + "/json/" + ip + "?fields=status,city,country")
	if err != nil {
		return unknownLocation, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return unknownLocation, errors.New("geoip: status HTTP tidak 200")
	}

	var body ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return unknownLocation, err
	}
	if body.Status != "success" || body.City == "" {
		return unknownLocation, nil
	}

	return body.City + ", " + body.Country, nil
}

func isPublicIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsUnspecified() && !ip.IsLinkLocalUnicast()
}
