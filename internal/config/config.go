package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type Config struct {
	Port                string
	DBDSN               string
	JWTSecret           []byte
	JWTExpiry           time.Duration
	RefreshTokenExpiry  time.Duration
	SMTP                SMTPConfig
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func LoadConfig() (*Config, error) {
	godotenv.Load()

	jwtExpiry := 15 * time.Minute
	if v := os.Getenv("JWT_EXPIRY_MINUTES"); v != "" {
		if d, err := time.ParseDuration(v + "m"); err == nil {
			jwtExpiry = d
		}
	}

	refreshExpiry := 7 * 24 * time.Hour
	if v := os.Getenv("REFRESH_EXPIRY_DAYS"); v != "" {
		if d, err := time.ParseDuration(v + "h"); err == nil {
			refreshExpiry = d * 24
		}
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	// Validasi SMTP: Jika salah satu diisi, maka semua wajib diisi
	hasSMTP := smtpHost != "" || smtpPort != "" || smtpUser != "" || smtpPass != "" || smtpFrom != ""
	if hasSMTP {
		if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || smtpFrom == "" {
			return nil, errors.New("konfigurasi SMTP tidak lengkap: seluruh variabel SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, dan SMTP_FROM wajib diisi jika ingin mengaktifkan layanan email")
		}

		// Validasi port harus berupa angka integer
		if _, err := strconv.Atoi(smtpPort); err != nil {
			return nil, fmt.Errorf("port SMTP tidak valid: '%s' harus berupa angka integer (contoh: 587 atau 465)", smtpPort)
		}

		// Validasi format email pengirim
		if !emailRegex.MatchString(smtpFrom) {
			return nil, fmt.Errorf("email pengirim SMTP_FROM tidak valid: '%s'", smtpFrom)
		}
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBDSN:              os.Getenv("DB_DSN"),
		JWTSecret:          []byte(os.Getenv("JWT_SECRET")),
		JWTExpiry:          jwtExpiry,
		RefreshTokenExpiry: refreshExpiry,
		SMTP: SMTPConfig{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: smtpUser,
			Password: smtpPass,
			From:     smtpFrom,
		},
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}