package helper

import (
	"bytes"
	"html/template"
	"backend/templates"
)

// VerificationPageData menampung data yang akan di-render ke dalam template HTML verifikasi email.
type VerificationPageData struct {
	IsSuccess   bool
	StatusClass string
	Title       string
	Subtitle    string
	ButtonText  string
	ButtonUrl   string
}

// Inisialisasi template verifikasi sekali saat aplikasi dimulai agar performa render optimal.
var verificationTemplate = template.Must(template.New("email_verification").Parse(templates.EmailVerificationHTML))

// GetEmailVerificationHTML mengembalikan halaman HTML interaktif dan responsif untuk status verifikasi email.
func GetEmailVerificationHTML(isSuccess bool, errMsg string) string {
	var data VerificationPageData
	data.IsSuccess = isSuccess

	if isSuccess {
		data.StatusClass = "success"
		data.Title = "Email Berhasil Diverifikasi!"
		data.Subtitle = "Akun Anda telah diaktifkan sepenuhnya. Silakan masuk kembali menggunakan akun Anda untuk mulai menjelajah platform kami."
		data.ButtonText = "Masuk Ke Aplikasi"
		data.ButtonUrl = "http://localhost:3000/login" // Default URL login
	} else {
		data.StatusClass = "failure"
		data.Title = "Verifikasi Email Gagal"
		if errMsg == "" {
			data.Subtitle = "Tautan verifikasi yang Anda gunakan tidak valid, telah kedaluwarsa, atau sudah pernah digunakan sebelumnya."
		} else {
			data.Subtitle = "Kesalahan: " + errMsg + ". Silakan ajukan ulang permintaan verifikasi atau hubungi layanan bantuan kami."
		}
		data.ButtonText = "Hubungi Dukungan"
		data.ButtonUrl = "mailto:support@example.com"
	}

	var buf bytes.Buffer
	if err := verificationTemplate.Execute(&buf, data); err != nil {
		// Fallback sederhana jika rendering template gagal
		return "<h1>" + data.Title + "</h1><p>" + data.Subtitle + "</p>"
	}

	return buf.String()
}
