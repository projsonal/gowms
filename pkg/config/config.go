package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App           AppConfig
	DB            DBConfig
	JWT           JWTConfig
	TOTP          TOTPConfig
	CORS          CORSConfig
	Storage       StorageConfig
	Captcha       CaptchaConfig
	HumanCheck    HumanCheckConfig
	BotCheck      BotCheckConfig
	WhatsApp      WhatsAppConfig
	WAOTP         WAOTPConfig
	SMS           SMSConfig
	PasswordReset PasswordResetConfig
	GeoIP         GeoIPConfig
	Swagger       SwaggerConfig
	SelfUpdate    SelfUpdateConfig
}

type AppConfig struct {
	Name           string
	Env            string
	Host           string
	Port           string
	TrustedProxies []string
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

type HumanCheckConfig struct {
	Secret          string
	TTLMinutes      int
	MinDelaySeconds int
}

type BotCheckConfig struct {
	Enabled       bool
	Secret        string
	WindowMinutes int
}

type WhatsAppConfig struct {
	Driver      string
	APIURL      string
	APIKey      string
	Sender      string
	SessionPath string
}

type WAOTPConfig struct {
	Secret     string
	TTLMinutes int
}

type SMSConfig struct {
	APIURL string
	APIKey string
	Sender string
}

type PasswordResetConfig struct {
	Secret     string
	TTLMinutes int
}

type GeoIPConfig struct {
	Enabled bool
	BaseURL string
}

type SwaggerConfig struct {
	Enabled       bool
	BasicAuthUser string
	BasicAuthPass string
}

type SelfUpdateConfig struct {
	Enabled     bool
	GitHubOwner string
	GitHubRepo  string
	ScriptPath  string
	WorkDir     string
	ServiceName string
	StatusPath  string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name:           getEnv("APP_NAME", "stockrsd"),
			Env:            getEnv("APP_ENV", "development"),
			Host:           getEnv("APP_HOST", ""),
			Port:           getEnv("APP_PORT", "8080"),
			TrustedProxies: getEnvList("TRUSTED_PROXIES", []string{"127.0.0.1/32", "::1/128"}),
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

			AllowedOrigins: splitAndTrim(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, http://10.11.12.173:3000")),
		},
		Storage: StorageConfig{
			Path: getEnv("STORAGE_PATH", "./storage"),
		},
		Captcha: CaptchaConfig{
			Secret:     getEnv("CAPTCHA_SECRET", "change-me-captcha-secret"),
			TTLMinutes: getEnvAsInt("CAPTCHA_TTL_MINUTES", 5),
		},
		HumanCheck: HumanCheckConfig{
			Secret:          getEnv("HUMANCHECK_SECRET", "change-me-humancheck-secret"),
			TTLMinutes:      getEnvAsInt("HUMANCHECK_TTL_MINUTES", 10),
			MinDelaySeconds: getEnvAsInt("HUMANCHECK_MIN_DELAY_SECONDS", 2),
		},
		BotCheck: BotCheckConfig{
			Enabled:       getEnvAsBool("BOTCHECK_ENABLED", false),
			Secret:        getEnv("BOTCHECK_SECRET", "change-me-botcheck-secret"),
			WindowMinutes: getEnvAsInt("BOTCHECK_WINDOW_MINUTES", 60),
		},
		WhatsApp: WhatsAppConfig{
			Driver:      getEnv("WHATSAPP_DRIVER", "gateway"),
			APIURL:      getEnv("WHATSAPP_API_URL", ""),
			APIKey:      getEnv("WHATSAPP_API_KEY", ""),
			Sender:      getEnv("WHATSAPP_SENDER", "stockrsd"),
			SessionPath: getEnv("WHATSMEOW_SESSION_PATH", "./var/whatsmeow-session.db"),
		},
		WAOTP: WAOTPConfig{
			Secret:     getEnv("WA_OTP_SECRET", "change-me-wa-otp-secret"),
			TTLMinutes: getEnvAsInt("WA_OTP_TTL_MINUTES", 5),
		},
		SMS: SMSConfig{
			APIURL: getEnv("SMS_API_URL", ""),
			APIKey: getEnv("SMS_API_KEY", ""),
			Sender: getEnv("SMS_SENDER", "stockrsd"),
		},
		PasswordReset: PasswordResetConfig{
			Secret:     getEnv("PASSWORD_RESET_SECRET", "change-me-password-reset-secret"),
			TTLMinutes: getEnvAsInt("PASSWORD_RESET_TTL_MINUTES", 10),
		},
		GeoIP: GeoIPConfig{
			Enabled: getEnvAsBool("GEOIP_ENABLED", false),
			BaseURL: getEnv("GEOIP_BASE_URL", "http://ip-api.com"),
		},
		Swagger: SwaggerConfig{
			Enabled:       getEnvAsBool("SWAGGER_ENABLED", getEnv("APP_ENV", "development") != "production"),
			BasicAuthUser: getEnv("SWAGGER_BASIC_AUTH_USER", ""),
			BasicAuthPass: getEnv("SWAGGER_BASIC_AUTH_PASS", ""),
		},
		SelfUpdate: SelfUpdateConfig{
			Enabled:     getEnvAsBool("AUTO_UPDATE_ENABLED", false),
			GitHubOwner: getEnv("AUTO_UPDATE_GITHUB_OWNER", "projsonal"),
			GitHubRepo:  getEnv("AUTO_UPDATE_GITHUB_REPO", "gowms"),
			ScriptPath:  getEnv("AUTO_UPDATE_SCRIPT_PATH", "./deploy/scripts/self-update.sh"),
			WorkDir:     getEnv("AUTO_UPDATE_WORKDIR", "."),
			ServiceName: getEnv("AUTO_UPDATE_SERVICE_NAME", "gowms-backend"),
			StatusPath:  getEnv("AUTO_UPDATE_STATUS_PATH", "./var/run/self-update-status.json"),
		},
	}

	validateProductionSecrets(cfg)
	return cfg
}

var weakSecretDefaults = map[string]string{
	"JWT_ACCESS_SECRET":  "access-secret",
	"JWT_REFRESH_SECRET": "refresh-secret",
	"CAPTCHA_SECRET":     "change-me-captcha-secret",
	"HUMANCHECK_SECRET":  "change-me-humancheck-secret",
	"BOTCHECK_SECRET":    "change-me-botcheck-secret",
}

func validateProductionSecrets(cfg *Config) {
	if cfg.App.Env != "production" {
		return
	}
	actual := map[string]string{
		"JWT_ACCESS_SECRET":  cfg.JWT.AccessSecret,
		"JWT_REFRESH_SECRET": cfg.JWT.RefreshSecret,
		"CAPTCHA_SECRET":     cfg.Captcha.Secret,
		"HUMANCHECK_SECRET":  cfg.HumanCheck.Secret,
		"BOTCHECK_SECRET":    cfg.BotCheck.Secret,
	}
	var insecure []string
	for key, value := range actual {
		if value == weakSecretDefaults[key] || len(value) < 24 {
			insecure = append(insecure, key)
		}
	}
	if len(insecure) > 0 {
		log.Fatalf(
			"config: APP_ENV=production tapi env var berikut masih pakai nilai contoh/terlalu pendek (min 24 karakter): %s — set nilai acak yang kuat sebelum deploy",
			strings.Join(insecure, ", "),
		)
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
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
