package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	Port               string
	DBDSN              string
	JWTPrivateKey      *rsa.PrivateKey
	JWTPublicKey       *rsa.PublicKey
	JWTExpiry          time.Duration
	RefreshTokenExpiry time.Duration
	SMTP               SMTPConfig
	AppURL             string
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

	// Membaca jalur file kunci RSA
	privateKeyPath := getEnv("JWT_PRIVATE_KEY_PATH", "keys/jwt_private.pem")
	publicKeyPath := getEnv("JWT_PUBLIC_KEY_PATH", "keys/jwt_public.pem")

	// Pastikan file kunci RSA ada, jika tidak ada, buat secara otomatis
	if err := ensureRSAKeys(privateKeyPath, publicKeyPath); err != nil {
		return nil, fmt.Errorf("gagal menginisialisasi kunci RSA JWT: %v", err)
	}

	// Muat Kunci Privat RSA
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file kunci privat: %v", err)
	}
	privBlock, _ := pem.Decode(privBytes)
	if privBlock == nil || privBlock.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("format kunci privat RSA tidak valid")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("gagal mem-parsing kunci privat RSA: %v", err)
	}

	// Muat Kunci Publik RSA
	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file kunci publik: %v", err)
	}
	pubBlock, _ := pem.Decode(pubBytes)
	if pubBlock == nil || pubBlock.Type != "PUBLIC KEY" {
		return nil, errors.New("format kunci publik RSA tidak valid")
	}
	pubKeyInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("gagal mem-parsing kunci publik RSA: %v", err)
	}
	pubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("tipe kunci publik bukan *rsa.PublicKey")
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
		JWTPrivateKey:      privKey,
		JWTPublicKey:       pubKey,
		JWTExpiry:          jwtExpiry,
		RefreshTokenExpiry: refreshExpiry,
		SMTP: SMTPConfig{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: smtpUser,
			Password: smtpPass,
			From:     smtpFrom,
		},
		AppURL: getEnv("APP_URL", "http://localhost:8080"),
	}, nil
}

// ensureRSAKeys mengecek keberadaan berkas kunci RSA, jika tidak ada maka akan membuat kunci baru secara otomatis.
func ensureRSAKeys(privPath, pubPath string) error {
	_, errPriv := os.Stat(privPath)
	_, errPub := os.Stat(pubPath)

	// Jika kedua berkas sudah ada, lewati pembuatan
	if errPriv == nil && errPub == nil {
		return nil
	}

	// Pastikan direktori folder kunci ada
	privDir := filepath.Dir(privPath)
	if err := os.MkdirAll(privDir, 0755); err != nil {
		return err
	}
	pubDir := filepath.Dir(pubPath)
	if err := os.MkdirAll(pubDir, 0755); err != nil {
		return err
	}

	// Generate kunci privat RSA 2048-bit
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Simpan Kunci Privat
	privFile, err := os.OpenFile(privPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer privFile.Close()

	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privFile, privBlock); err != nil {
		return err
	}

	// Simpan Kunci Publik
	pubFile, err := os.OpenFile(pubPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer pubFile.Close()

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	if err := pem.Encode(pubFile, pubBlock); err != nil {
		return err
	}

	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
