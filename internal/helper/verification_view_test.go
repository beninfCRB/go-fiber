package helper

import (
	"strings"
	"testing"
)


func TestGetEmailVerificationHTML(t *testing.T) {
	t.Run("Success Case", func(t *testing.T) {
		html := GetEmailVerificationHTML(true, "")
		
		if !strings.Contains(html, "Email Berhasil Diverifikasi!") {
			t.Errorf("expected success title not found in HTML")
		}
		if !strings.Contains(html, "checkmark__check") {
			t.Errorf("expected success checkmark icon not found in HTML")
		}
		if !strings.Contains(html, "#0acf97") {
			t.Errorf("expected success teal color theme not found in HTML")
		}
	})

	t.Run("Failure Case with custom error", func(t *testing.T) {
		errorMsg := "Token kadaluwarsa"
		html := GetEmailVerificationHTML(false, errorMsg)
		
		if !strings.Contains(html, "Verifikasi Email Gagal") {
			t.Errorf("expected failure title not found in HTML")
		}
		if !strings.Contains(html, "Kesalahan: Token kadaluwarsa") {
			t.Errorf("expected custom error message not found in HTML")
		}
		if !strings.Contains(html, "#ff4757") {
			t.Errorf("expected failure rose color theme not found in HTML")
		}
	})
}
