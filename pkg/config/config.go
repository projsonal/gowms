// Package config memuat seluruh konfigurasi aplikasi StockRSD dari
// environment variable (.env), dipakai oleh cmd/main.go sebagai composition root.
package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config merepresentasikan seluruh konfigurasi aplikasi.
type Config struct {
	App      AppConfig
	DB       DBConfig
	JWT      JWTConfig
	TOTP     TOTPConfig
	CORS     CORSConfig
	Storage  StorageConfig
	Captcha  CaptchaConfig
	BotCheck BotCheckConfig
	WhatsApp WhatsAppConfig
	WAOTP    WAOTPConfig
	GeoIP    GeoIPConfig
}

type AppConfig struct {
	Name string
	Env  string
	Host string
	Port string
}

func (a AppConfig) ListenAddress() string {
	if a.Host == "" {
		return ":" + a.Port
	}
	if strings.Contains(a.Host, ":") {
		return "[" + a.Host + "]:" + a.Port
	}
	return a.Host + ":" + a.Port
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpiryMinutes int
	RefreshExpiryDays   int
}

type TOTPConfig struct {
	Issuer string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type StorageConfig struct {
	Path string
}

type CaptchaConfig struct {
	Secret     string
	TTLMinutes int
}

type BotCheckConfig struct {
	Secret        string
	WindowMinutes int
}

type WhatsAppConfig struct {
	APIURL string
	APIKey string
	Sender string
}

type WAOTPConfig struct {
	Secret     string
	TTLMinutes int
}

type GeoIPConfig struct {
	Enabled bool
	BaseURL string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "WMS-RSD"),
			Env:  getEnv("APP_ENV", "development"),
			Host: getEnv("APP_HOST", ""),
			Port: getEnv("APP_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "t4mp4n"),
			Name:     getEnv("DB_NAME", "stockrsd"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			AccessSecret:        getEnv("JWT_ACCESS_SECRET", "access-secret"),
			RefreshSecret:       getEnv("JWT_REFRESH_SECRET", "refresh-secret"),
			AccessExpiryMinutes: getEnvAsInt("JWT_ACCESS_EXPIRY_MINUTES", 15),
			RefreshExpiryDays:   getEnvAsInt("JWT_REFRESH_EXPIRY_DAYS", 7),
		},
		TOTP: TOTPConfig{
			Issuer: getEnv("TOTP_ISSUER", "WMS - RSD"),
		},
		CORS: CORSConfig{

			AllowedOrigins: splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		},
		Storage: StorageConfig{
			Path: getEnv("STORAGE_PATH", ""),
		},
		Captcha: CaptchaConfig{
			Secret:     getEnv("CAPTCHA_SECRET", "change-me-captcha-secret"),
			TTLMinutes: getEnvAsInt("CAPTCHA_TTL_MINUTES", 5),
		},
		BotCheck: BotCheckConfig{
			Secret:        getEnv("BOTCHECK_SECRET", "change-me-botcheck-secret"),
			WindowMinutes: getEnvAsInt("BOTCHECK_WINDOW_MINUTES", 60),
		},
		WhatsApp: WhatsAppConfig{
			APIURL: getEnv("WHATSAPP_API_URL", ""),
			APIKey: getEnv("WHATSAPP_API_KEY", ""),
			Sender: getEnv("WHATSAPP_SENDER", "wms-RSD"),
		},
		WAOTP: WAOTPConfig{
			Secret:     getEnv("WA_OTP_SECRET", "change-me-wa-otp-secret"),
			TTLMinutes: getEnvAsInt("WA_OTP_TTL_MINUTES", 5),
		},
		GeoIP: GeoIPConfig{
			Enabled: getEnvAsBool("GEOIP_ENABLED", false),
			BaseURL: getEnv("GEOIP_BASE_URL", "http://ip-api.com"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
