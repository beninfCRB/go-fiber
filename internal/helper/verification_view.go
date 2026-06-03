package helper

import (
	"backend/templates"
	"bytes"
	"html/template"
	"strings"
)

type VerificationPageData struct {
	IsSuccess   bool
	StatusClass string
	Title       string
	Subtitle    string
	ButtonText  string
	ButtonUrl   string
}

var verificationTemplate = template.Must(template.New("email_verification").Parse(templates.EmailVerificationHTML))

func GetEmailVerificationHTML(isSuccess bool, errMsg string, appURL string) string {
	var data VerificationPageData
	data.IsSuccess = isSuccess

	// Normalisasi URL: hapus trailing slash lalu tambahkan /login
	baseURL := strings.TrimRight(appURL, "/")
	data.ButtonUrl = baseURL + "/login"

	if isSuccess {
		data.StatusClass = "success"
		data.Title = "Email Berhasil Diverifikasi!"
		data.Subtitle = "Akun Anda telah diaktifkan sepenuhnya. Silakan masuk kembali menggunakan akun Anda untuk mulai menjelajah platform kami."
		data.ButtonText = "Masuk ke Akun"
	} else {
		data.StatusClass = "failure"
		data.Title = "Verifikasi Email Gagal"
		data.ButtonText = "Kembali ke Beranda"
		if errMsg == "" {
			data.Subtitle = "Tautan verifikasi yang Anda gunakan tidak valid, telah kedaluwarsa, atau sudah pernah digunakan sebelumnya."
		} else {
			data.Subtitle = "Kesalahan: " + errMsg + ". Silakan ajukan ulang permintaan verifikasi atau hubungi layanan bantuan kami."
		}
	}

	var buf bytes.Buffer
	if err := verificationTemplate.Execute(&buf, data); err != nil {
		// Fallback sederhana jika rendering template gagal
		return "<h1>" + data.Title + "</h1><p>" + data.Subtitle + "</p>"
	}

	return buf.String()
}
