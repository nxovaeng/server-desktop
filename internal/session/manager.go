package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"xray-panel-desktop/internal/apiclient"
	"xray-panel-desktop/internal/local"
	"xray-panel-desktop/internal/store"
	"xray-panel-desktop/internal/tunnel"
)

// ConnectionStatus reports current connection state for the active server profile.
type ConnectionStatus struct {
	Connected        bool   `json:"connected"`
	ProfileID        string `json:"profile_id"`
	ProfileName      string `json:"profile_name"`
	Mode             string `json:"mode"`
	LatencyMs        int64  `json:"latency_ms"`
	SSHFingerprint   string `json:"ssh_fingerprint,omitempty"`
	ServerVersion    string `json:"server_version,omitempty"`
	TokenName        string `json:"token_name,omitempty"`
	ServerPublicHost string `json:"server_public_host,omitempty"`
	SubDomain        string `json:"sub_domain,omitempty"`
	SubPath          string `json:"sub_path,omitempty"`
	PanelTitle       string `json:"panel_title,omitempty"`
	Warning          string `json:"warning,omitempty"`
}

// Manager coordinates profile switching, SSH tunnel lifecycles, and API handshake verification.
type Manager struct {
	mu      sync.RWMutex
	store   *store.ProfileStore
	profile *store.ServerProfile
	secrets *store.ProfileSecrets
	tun     *tunnel.Tunnel
	client  *apiclient.Client
	status  ConnectionStatus
}

func NewManager(profileStore *store.ProfileStore) *Manager {
	return &Manager{store: profileStore}
}

// Connect establishes a direct or SSH-tunneled session to profileID and verifies the Bearer token.
func (m *Manager) Connect(ctx context.Context, profileID string, trustTOFU bool) (*ConnectionStatus, error) {
	profile, secrets, err := m.store.Get(profileID)
	if err != nil {
		return nil, err
	}
	if secrets.APIToken == "" {
		return nil, errors.New("profile is missing API token")
	}

	m.Disconnect()

	var newTun *tunnel.Tunnel
	var baseURL string
	var dialCtx apiclient.DialContextFunc
	var warning string

	if profile.Mode == store.ModeSSHTunnel {
		var savedFP string
		tunCfg := tunnel.Config{
			Host:                profile.SSHHost,
			Port:                profile.SSHPort,
			User:                profile.SSHUser,
			AuthMethod:          tunnel.AuthMethod(profile.SSHAuthType),
			KeyPath:             profile.SSHKeyPath,
			KeyPEM:              secrets.SSHKeyPEM,
			KeyPassphrase:       secrets.SSHKeyPassphrase,
			Password:            secrets.SSHPassword,
			ExpectedFingerprint: profile.SSHFingerprint,
			TrustOnFirstUse:     trustTOFU,
			OnFirstFingerprint: func(fp string) {
				savedFP = fp
			},
		}

		newTun, err = tunnel.Connect(tunCfg)
		if err != nil {
			return nil, err
		}

		if savedFP != "" {
			_ = m.store.UpdateSSHFingerprint(profile.ID, savedFP)
			profile.SSHFingerprint = savedFP
		}

		bindHost := profile.RemoteBindHost
		if bindHost == "" {
			bindHost = "127.0.0.1"
		}
		port := profile.RemotePanelPort
		if port <= 0 {
			port = 8082
		}
		baseURL = fmt.Sprintf("http://%s", net.JoinHostPort(bindHost, fmt.Sprintf("%d", port)))
		dialCtx = newTun.DialContext
	} else {
		urlCheck, err := local.ValidateDirectURL(profile.DirectURL)
		if err != nil {
			return nil, err
		}
		if urlCheck.RequiresHTTPSWarn && !profile.AllowInsecureHTTP {
			return nil, errors.New("refusing plaintext http:// connection to non-local host without explicit allow_insecure_http")
		}
		if urlCheck.RequiresHTTPSWarn {
			warning = "当前使用明文 HTTP 直连公网节点，建议切换为 SSH 隧道或 HTTPS 域名"
		}
		baseURL = urlCheck.NormalizedURL
	}

	client := apiclient.NewClient(baseURL, secrets.APIToken, dialCtx)

	start := time.Now()
	verify, err := client.VerifyAuth(ctx)
	rtt := time.Since(start).Milliseconds()
	if err != nil {
		if newTun != nil {
			_ = newTun.Close()
		}
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	if newTun != nil && newTun.LatencyMs() > 0 && rtt == 0 {
		rtt = newTun.LatencyMs()
	}

	_ = m.store.SetActive(profile.ID)

	st := ConnectionStatus{
		Connected:        true,
		ProfileID:        profile.ID,
		ProfileName:      profile.Name,
		Mode:             string(profile.Mode),
		LatencyMs:        rtt,
		SSHFingerprint:   profile.SSHFingerprint,
		ServerVersion:    verify.Version,
		TokenName:        verify.TokenName,
		ServerPublicHost: verify.ServerPublicHost,
		SubDomain:        verify.SubDomain,
		SubPath:          verify.SubPath,
		PanelTitle:       verify.PanelTitle,
		Warning:          warning,
	}

	m.mu.Lock()
	m.profile = profile
	m.secrets = secrets
	m.tun = newTun
	m.client = client
	m.status = st
	m.mu.Unlock()

	return &st, nil
}

// Disconnect closes any active SSH tunnel and clears the active session client.
func (m *Manager) Disconnect() {
	m.mu.Lock()
	tun := m.tun
	m.tun = nil
	m.client = nil
	m.profile = nil
	m.secrets = nil
	m.status = ConnectionStatus{Connected: false}
	m.mu.Unlock()

	if tun != nil {
		_ = tun.Close()
	}
}

// Client returns the active apiclient.Client or an error if not connected.
func (m *Manager) Client() (*apiclient.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.client == nil {
		return nil, errors.New("not connected to any server profile")
	}
	return m.client, nil
}

// Status returns a snapshot of the current connection status.
func (m *Manager) Status() ConnectionStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st := m.status
	if m.tun != nil && !m.tun.IsConnected() {
		st.Connected = false
	}
	return st
}

// TestLatency measures `/api/v1/health` roundtrip on the active session.
func (m *Manager) TestLatency(ctx context.Context) (int64, error) {
	client, err := m.Client()
	if err != nil {
		return 0, err
	}
	start := time.Now()
	if _, err := client.Health(ctx); err != nil {
		return 0, err
	}
	rtt := time.Since(start).Milliseconds()

	m.mu.Lock()
	m.status.LatencyMs = rtt
	m.mu.Unlock()
	return rtt, nil
}

// RestartPanelAndReconnect triggers panel restart and waits for health + tunnel recovery.
func (m *Manager) RestartPanelAndReconnect(ctx context.Context) (*ConnectionStatus, error) {
	m.mu.RLock()
	client := m.client
	prof := m.profile
	m.mu.RUnlock()

	if client == nil || prof == nil {
		return nil, errors.New("not connected to any server profile")
	}

	if _, err := client.RestartPanel(ctx); err != nil {
		return nil, err
	}

	// Allow backend 500ms restart delay to begin
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(1500 * time.Millisecond):
	}

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if _, err := client.WaitForHealthy(waitCtx, time.Second); err == nil {
		if _, err := m.TestLatency(waitCtx); err == nil {
			st := m.Status()
			return &st, nil
		}
	}

	// Re-establish session (rebuilds SSH tunnel if needed)
	return m.Connect(waitCtx, prof.ID, false)
}
