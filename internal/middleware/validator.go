package middleware

import "github.com/go-playground/validator/v10"

// Validate is the shared validator instance for all handlers.
var Validate = validator.New()
