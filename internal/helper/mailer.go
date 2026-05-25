package helper

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"time"
)

// Mailer adalah utilitas untuk menangani pengiriman email baik via SMTP maupun simulasi log.
type Mailer struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewMailer membuat instance baru dari Mailer dengan menerima konfigurasi SMTP secara langsung.
func NewMailer(host, port, username, password, from string) *Mailer {
	return &Mailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// SendEmail mengirimkan email. Menggunakan SMTP asli jika konfigurasi lengkap,
// atau melakukan fallback dengan menulis isi email ke file log di folder storage/logs/mails.log.
func (m *Mailer) SendEmail(to, subject, body string) error {
	// Jika konfigurasi SMTP lengkap, kirim email asli
	if m.host != "" && m.port != "" && m.username != "" && m.password != "" {
		auth := smtp.PlainAuth("", m.username, m.password, m.host)
		addr := fmt.Sprintf("%s:%s", m.host, m.port)
		msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, body))
		
		err := smtp.SendMail(addr, auth, m.from, []string{to}, msg)
		if err != nil {
			log.Printf("[MAILER ERROR] Gagal mengirim email SMTP ke %s: %v", to, err)
			// Lanjutkan ke fallback log jika SMTP gagal agar alur program tidak terputus
		} else {
			log.Printf("[MAILER] Email SMTP berhasil dikirim ke %s", to)
			return nil
		}
	}

	// Fallback: Tulis ke file log lokal di folder storage/logs
	logDir := filepath.Join(".", "storage", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori log email: %v", err)
	}

	logFile := filepath.Join(logDir, "mails.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("gagal membuka file log email: %v", err)
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logContent := fmt.Sprintf("========================================================================\n"+
		"WAKTU   : %s\n"+
		"KEPADA  : %s\n"+
		"SUBJEK  : %s\n"+
		"PESAN   :\n%s\n"+
		"========================================================================\n\n",
		timestamp, to, subject, body)

	if _, err := f.WriteString(logContent); err != nil {
		return fmt.Errorf("gagal menulis ke file log email: %v", err)
	}

	log.Printf("[MAILER SIMULATION] Email simulasi ke %s dicatat di %s", to, logFile)
	return nil
}
