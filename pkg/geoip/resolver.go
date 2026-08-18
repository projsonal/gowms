package geoip

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	unknownLocation       = "-"
	defaultTimeout        = 2 * time.Second
	maxResponseSize int64 = 1 << 20 // 1 MB
	geoPath               = "/json/"
	geoFields             = "status,city,country,timezone"
	statusSuccess         = "success"
)

type Resolver interface {
	Lookup(ctx context.Context, ip string) (string, error)
}

type NoopResolver struct{}

func (NoopResolver) Lookup(context.Context, string) (string, error) {
	return unknownLocation, nil
}

type httpResolver struct {
	client  *http.Client
	baseURL *url.URL
}

func NewHTTPResolver(baseURL string) (Resolver, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse geoip base url: %w", err)
	}
	return &httpResolver{
		client:  &http.Client{Timeout: defaultTimeout},
		baseURL: parsedURL,
	}, nil
}

type ipAPIResponse struct {
	Status   string `json:"status"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Timezone string `json:"timezone"`
}

func (r *httpResolver) Lookup(ctx context.Context, ip string) (string, error) {
	if !isPublicIP(ip) {
		return unknownLocation, nil
	}

	requestURL := r.buildURL(ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return unknownLocation, fmt.Errorf("create geoip request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return unknownLocation, fmt.Errorf("execute geoip request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return unknownLocation, fmt.Errorf("geoip api returned status %d", resp.StatusCode)
	}

	result, err := decodeResponse(resp)
	if err != nil {
		return unknownLocation, err
	}
	if !result.isValid() {
		return unknownLocation, nil
	}

	return formatLocation(result.City, result.Country, result.Timezone), nil
}

func (r *httpResolver) buildURL(ip string) string {
	u := *r.baseURL
	u.Path = geoPath + ip

	query := u.Query()
	query.Set("fields", geoFields)
	u.RawQuery = query.Encode()

	return u.String()
}

func decodeResponse(resp *http.Response) (*ipAPIResponse, error) {
	reader := io.LimitReader(resp.Body, maxResponseSize)

	var result ipAPIResponse
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode geoip response: %w", err)
	}
	return &result, nil
}

func (r ipAPIResponse) isValid() bool {
	return r.Status == statusSuccess && r.City != ""
}

func formatLocation(city, country, timezone string) string {
	location := city
	if country != "" {
		location = fmt.Sprintf("%s, %s", city, country)
	}
	if tz := describeTimezone(timezone); tz != "" {
		return fmt.Sprintf("%s (%s)", location, tz)
	}
	return location
}
