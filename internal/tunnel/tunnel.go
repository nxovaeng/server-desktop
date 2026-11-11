package tunnel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// AuthMethod specifies SSH authentication mechanism.
type AuthMethod string

const (
	AuthKeyFile  AuthMethod = "key_file"
	AuthKeyPEM   AuthMethod = "key_pem"
	AuthPassword AuthMethod = "password"
	AuthAgent    AuthMethod = "agent"
)

// Config defines parameters to establish an in-process SSH tunnel without local listening ports.
type Config struct {
	Host                string
	Port                int
	User                string
	AuthMethod          AuthMethod
	KeyPath             string
	KeyPEM              string
	KeyPassphrase       string
	Password            string
	ExpectedFingerprint string
	TrustOnFirstUse     bool
	ConnectTimeout      time.Duration
	KeepaliveInterval   time.Duration
	OnFirstFingerprint  func(fingerprint string)
	OnDisconnect        func(err error)
}

// TOFURequiredError is returned when connecting to a host with no stored fingerprint and TrustOnFirstUse=false.
type TOFURequiredError struct {
	Host        string
	Fingerprint string
}

func (e *TOFURequiredError) Error() string {
	return fmt.Sprintf("unknown SSH host key for %s (fingerprint: %s)", e.Host, e.Fingerprint)
}

// HostKeyMismatchError is returned when the server's SSH host key does not match the stored TOFU fingerprint.
type HostKeyMismatchError struct {
	Host     string
	Expected string
	Actual   string
}

func (e *HostKeyMismatchError) Error() string {
	return fmt.Sprintf("SSH host key mismatch for %s: expected %s, got %s", e.Host, e.Expected, e.Actual)
}

// Tunnel manages an in-process ssh.Client with keepalive monitoring and DialContext transport.
// It binds zero local TCP ports.
type Tunnel struct {
	cfg       Config
	mu        sync.RWMutex
	client    *ssh.Client
	agentConn net.Conn
	stopCh    chan struct{}
	closed    bool
	serverFP  string
	lastRTTMs int64
}

// Connect establishes the SSH session, enforces TOFU host key verification, and starts 20s keepalives.
func Connect(cfg Config) (*Tunnel, error) {
	if cfg.Host == "" {
		return nil, errors.New("ssh host is required")
	}
	if cfg.Port <= 0 {
		cfg.Port = 22
	}
	if cfg.User == "" {
		cfg.User = "root"
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 10 * time.Second
	}
	if cfg.KeepaliveInterval <= 0 {
		cfg.KeepaliveInterval = 20 * time.Second
	}

	authMethods, agentConn, err := buildAuthMethods(cfg)
	if err != nil {
		return nil, err
	}

	t := &Tunnel{
		cfg:       cfg,
		agentConn: agentConn,
		stopCh:    make(chan struct{}),
	}

	var tofuErr error
	hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		fp := ssh.FingerprintSHA256(key)
		t.mu.Lock()
		t.serverFP = fp
		t.mu.Unlock()

		if cfg.ExpectedFingerprint == "" {
			if !cfg.TrustOnFirstUse {
				tofuErr = &TOFURequiredError{Host: cfg.Host, Fingerprint: fp}
				return tofuErr
			}
			if cfg.OnFirstFingerprint != nil {
				cfg.OnFirstFingerprint(fp)
			}
			return nil
		}

		if fp != cfg.ExpectedFingerprint {
			tofuErr = &HostKeyMismatchError{
				Host:     cfg.Host,
				Expected: cfg.ExpectedFingerprint,
				Actual:   fp,
			}
			return tofuErr
		}
		return nil
	}

	sshCfg := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         cfg.ConnectTimeout,
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	start := time.Now()
	client, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		if agentConn != nil {
			_ = agentConn.Close()
		}
		if tofuErr != nil {
			return nil, tofuErr
		}
		return nil, fmt.Errorf("ssh dial %s failed: %w", addr, err)
	}

	t.client = client
	t.lastRTTMs = time.Since(start).Milliseconds()
	go t.keepaliveLoop()

	return t, nil
}

// DialContext opens a TCP stream through the SSH tunnel to remote addr (e.g. "127.0.0.1:8082").
func (t *Tunnel) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	t.mu.RLock()
	client := t.client
	closed := t.closed
	t.mu.RUnlock()

	if closed || client == nil {
		return nil, errors.New("ssh tunnel is closed")
	}

	type dialRes struct {
		conn net.Conn
		err  error
	}
	ch := make(chan dialRes, 1)
	go func() {
		c, err := client.Dial(network, addr)
		ch <- dialRes{conn: c, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.conn, r.err
	}
}

// Fingerprint returns the remote server's SHA-256 SSH host key fingerprint.
func (t *Tunnel) Fingerprint() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.serverFP
}

// LatencyMs returns the latest measured SSH RTT in milliseconds.
func (t *Tunnel) LatencyMs() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastRTTMs
}

// IsConnected reports whether the underlying SSH session is active.
func (t *Tunnel) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return !t.closed && t.client != nil
}

// Close tears down the keepalive loop, SSH connection, and any SSH agent socket.
func (t *Tunnel) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	close(t.stopCh)
	client := t.client
	agentConn := t.agentConn
	t.client = nil
	t.agentConn = nil
	t.mu.Unlock()

	var firstErr error
	if client != nil {
		if err := client.Close(); err != nil {
			firstErr = err
		}
	}
	if agentConn != nil {
		_ = agentConn.Close()
	}
	return firstErr
}

func (t *Tunnel) keepaliveLoop() {
	ticker := time.NewTicker(t.cfg.KeepaliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.mu.RLock()
			client := t.client
			closed := t.closed
			t.mu.RUnlock()
			if closed || client == nil {
				return
			}

			start := time.Now()
			_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				_ = t.Close()
				if t.cfg.OnDisconnect != nil {
					t.cfg.OnDisconnect(err)
				}
				return
			}

			t.mu.Lock()
			t.lastRTTMs = time.Since(start).Milliseconds()
			t.mu.Unlock()
		}
	}
}

func buildAuthMethods(cfg Config) ([]ssh.AuthMethod, net.Conn, error) {
	switch cfg.AuthMethod {
	case AuthPassword:
		if cfg.Password == "" {
			return nil, nil, errors.New("ssh password is required")
		}
		return []ssh.AuthMethod{ssh.Password(cfg.Password)}, nil, nil

	case AuthKeyPEM:
		if strings.TrimSpace(cfg.KeyPEM) == "" {
			return nil, nil, errors.New("ssh private key PEM is required")
		}
		signer, err := parsePrivateKey([]byte(cfg.KeyPEM), cfg.KeyPassphrase)
		if err != nil {
			return nil, nil, err
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil, nil

	case AuthKeyFile:
		path := expandHome(cfg.KeyPath)
		if path == "" {
			home, _ := os.UserHomeDir()
			for _, candidate := range []string{"id_ed25519", "id_ecdsa", "id_rsa"} {
				p := filepath.Join(home, ".ssh", candidate)
				if _, err := os.Stat(p); err == nil {
					path = p
					break
				}
			}
		}
		if path == "" {
			return nil, nil, errors.New("ssh private key file path not provided and no default key found")
		}
		pemBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read ssh key %s: %w", path, err)
		}
		signer, err := parsePrivateKey(pemBytes, cfg.KeyPassphrase)
		if err != nil {
			return nil, nil, err
		}
		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil, nil

	case AuthAgent:
		sock := os.Getenv("SSH_AUTH_SOCK")
		if sock == "" {
			return nil, nil, errors.New("SSH_AUTH_SOCK environment variable is not set")
		}
		conn, err := net.Dial("unix", sock)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to SSH agent (%s): %w", sock, err)
		}
		ag := agent.NewClient(conn)
		return []ssh.AuthMethod{ssh.PublicKeysCallback(ag.Signers)}, conn, nil

	default:
		return nil, nil, fmt.Errorf("unsupported ssh auth method: %s", cfg.AuthMethod)
	}
}

func parsePrivateKey(pemBytes []byte, passphrase string) (ssh.Signer, error) {
	if passphrase != "" {
		signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("failed to parse encrypted ssh key: %w", err)
		}
		return signer, nil
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ssh key: %w", err)
	}
	return signer, nil
}

func expandHome(p string) string {
	p = strings.TrimSpace(p)
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
