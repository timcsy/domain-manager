package models

import "errors"

var (
	// Domain errors
	ErrInvalidDomainName  = errors.New("invalid domain name")
	ErrInvalidServiceName = errors.New("invalid service name")
	ErrInvalidPort        = errors.New("invalid port number")
	ErrDomainNotFound     = errors.New("domain not found")
	ErrDomainExists       = errors.New("domain already exists")

	// Auth errors
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUnauthorized       = errors.New("unauthorized")

	// Diagnostic Log errors
	ErrDiagnosticLogNotFound = errors.New("diagnostic log not found")

	// General errors
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInternalError = errors.New("internal server error")
)
