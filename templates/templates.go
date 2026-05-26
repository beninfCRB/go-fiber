package templates

import (
	_ "embed"
)

//go:embed email_verification.html
var EmailVerificationHTML string
