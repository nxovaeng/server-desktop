package local

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// WGKeyPair holds a generated Curve25519 key pair.
type WGKeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// GenerateWGKeyPair generates a Curve25519 keypair locally without contacting the server.
func GenerateWGKeyPair() (*WGKeyPair, error) {
	privKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return &WGKeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(privKey.Bytes()),
		PublicKey:  base64.StdEncoding.EncodeToString(privKey.PublicKey().Bytes()),
	}, nil
}

// DeriveWGPublicKey derives the Curve25519 public key from a base64 private key.
func DeriveWGPublicKey(privKeyB64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(privKeyB64))
	if err != nil {
		return "", fmt.Errorf("invalid base64 private key: %w", err)
	}
	privKey, err := ecdh.X25519().NewPrivateKey(raw)
	if err != nil {
		return "", fmt.Errorf("invalid Curve25519 private key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(privKey.PublicKey().Bytes()), nil
}

// URLCheckResult describes a validated direct connection URL.
type URLCheckResult struct {
	NormalizedURL     string `json:"normalized_url"`
	IsPlaintextHTTP   bool   `json:"is_plaintext_http"`
	IsLoopback        bool   `json:"is_loopback"`
	RequiresHTTPSWarn bool   `json:"requires_https_warn"`
}

// ValidateDirectURL normalizes and inspects a direct panel server URL, flagging non-TLS public HTTP.
func ValidateDirectURL(raw string) (*URLCheckResult, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, errors.New("server URL is required")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	u, err := url.Parse(trimmed)
	if err != nil || u.Host == "" {
		return nil, errors.New("invalid server URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	host := u.Hostname()
	isLoopback := host == "localhost"
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		isLoopback = true
	}

	normalized := strings.TrimRight(u.String(), "/")
	isPlain := u.Scheme == "http"
	return &URLCheckResult{
		NormalizedURL:     normalized,
		IsPlaintextHTTP:   isPlain,
		IsLoopback:        isLoopback,
		RequiresHTTPSWarn: isPlain && !isLoopback,
	}, nil
}
