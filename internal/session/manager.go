package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"xray-panel-desktop/internal/apiclient"
	"xray-panel-desktop/internal/local"
	"xray-panel-desktop/internal/store"
	"xray-panel-desktop/internal/tunnel"
)

const DefaultBackendUpdateScriptURL = "https://raw.githubusercontent.com/nxovaeng/apiserver/main/scripts/install.sh"

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// BackendUpdateResult reports the outcome of upgrading xray-panel-server over the SSH tunnel.
type BackendUpdateResult struct {
	Success  bool             `json:"success"`
	ExitCode int              `json:"exit_code"`
	Output   string           `json:"output"`
	Error    string           `json:"error,omitempty"`
	Status   ConnectionStatus `json:"status"`
}

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
	mu        sync.RWMutex
	connectMu sync.Mutex
	store     *store.ProfileStore
	profile   *store.ServerProfile
	secrets   *store.ProfileSecrets
	tun       *tunnel.Tunnel
	client    *apiclient.Client
	status    ConnectionStatus
}

func NewManager(profileStore *store.ProfileStore) *Manager {
	return &Manager{store: profileStore}
}

// Connect establishes a direct or SSH-tunneled session to profileID and verifies the Bearer token.
// It is serialized with a dedicated mutex: two concurrent Connect calls run one after
// the other instead of each building a tunnel and leaking the first one.
func (m *Manager) Connect(ctx context.Context, profileID string, trustTOFU bool) (*ConnectionStatus, error) {
	m.connectMu.Lock()
	defer m.connectMu.Unlock()

	profile, secrets, err := m.store.Get(profileID)
	if err != nil {
		return nil, err
	}
	if secrets.APIToken == "" {
		return nil, errors.New("profile is missing API token")
	}

	m.disconnectLocked()

	var newTun *tunnel.Tunnel
	var baseURL string
	var dialCtx apiclient.DialContextFunc
	var warning string
	var savedFP string

	if profile.Mode == store.ModeSSHTunnel {
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

		// NOTE: the TOFU fingerprint captured in savedFP is NOT persisted here.
		// It is only written to the profile after VerifyAuth succeeds below,
		// so a failed token check cannot pin a fingerprint for a host we
		// never authenticated against.
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

	// Token verified: only now is it safe to trust the TOFU fingerprint
	// observed during the SSH handshake.
	if savedFP != "" {
		_ = m.store.UpdateSSHFingerprint(profile.ID, savedFP)
		profile.SSHFingerprint = savedFP
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
// It takes the same connectMu as Connect so a disconnect racing a connect
// cannot tear down a tunnel the connect just established.
func (m *Manager) Disconnect() {
	m.connectMu.Lock()
	defer m.connectMu.Unlock()
	m.disconnectLocked()
}

// disconnectLocked is Disconnect without the connectMu; callers must hold it.
func (m *Manager) disconnectLocked() {
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

	return m.refreshOrReconnect(waitCtx, prof.ID, client)
}

// UpdateBackendViaSSH executes the server install.sh update command over the active SSH tunnel,
// then waits for the restarted backend API to come back healthy and refreshes ServerVersion.
func (m *Manager) UpdateBackendViaSSH(ctx context.Context) (*BackendUpdateResult, error) {
	m.mu.RLock()
	client := m.client
	prof := m.profile
	tun := m.tun
	m.mu.RUnlock()

	if client == nil || prof == nil {
		return nil, errors.New("not connected to any server profile")
	}
	if prof.Mode != store.ModeSSHTunnel || tun == nil {
		return nil, errors.New("仅 SSH 隧道连接模式支持执行 SSH 在线更新")
	}
	if !tun.IsConnected() {
		return nil, errors.New("SSH 隧道已断开，请先重新连接")
	}

	rawOut, exitCode, runErr := tun.RunCommand(ctx, buildBackendUpdateCommand(DefaultBackendUpdateScriptURL))
	cleanOut := stripANSI(strings.TrimSpace(rawOut))
	if runErr != nil {
		errMsg := fmt.Sprintf("SSH 在线更新失败 (exit=%d): %v", exitCode, runErr)
		return &BackendUpdateResult{
			Success:  false,
			ExitCode: exitCode,
			Output:   cleanOut,
			Error:    errMsg,
			Status:   m.Status(),
		}, nil
	}

	waitCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	st, refreshErr := m.refreshOrReconnect(waitCtx, prof.ID, client)
	if refreshErr != nil {
		return &BackendUpdateResult{
			Success:  true,
			ExitCode: 0,
			Output:   cleanOut,
			Error:    "更新命令已执行，但等待面板重连失败: " + refreshErr.Error(),
			Status:   m.Status(),
		}, nil
	}

	return &BackendUpdateResult{
		Success:  true,
		ExitCode: 0,
		Output:   cleanOut,
		Status:   *st,
	}, nil
}

func (m *Manager) refreshOrReconnect(ctx context.Context, profileID string, client *apiclient.Client) (*ConnectionStatus, error) {
	if _, err := client.WaitForHealthy(ctx, time.Second); err == nil {
		start := time.Now()
		if verify, vErr := client.VerifyAuth(ctx); vErr == nil {
			rtt := time.Since(start).Milliseconds()
			m.mu.Lock()
			if m.profile != nil && m.profile.ID == profileID {
				m.status.Connected = true
				m.status.LatencyMs = rtt
				m.status.ServerVersion = verify.Version
				m.status.TokenName = verify.TokenName
				m.status.ServerPublicHost = verify.ServerPublicHost
				m.status.SubDomain = verify.SubDomain
				m.status.SubPath = verify.SubPath
				m.status.PanelTitle = verify.PanelTitle
			}
			st := m.status
			m.mu.Unlock()
			return &st, nil
		}
	}

	// Re-establish session (rebuilds SSH tunnel if needed)
	return m.Connect(ctx, profileID, false)
}

func buildBackendUpdateCommand(scriptURL string) string {
	return fmt.Sprintf(`set -e
TMP_SCRIPT="$(mktemp /tmp/xray-panel-install.XXXXXX.sh)"
trap 'rm -f "$TMP_SCRIPT"' EXIT
if command -v curl >/dev/null 2>&1; then
  curl -fsSL %q -o "$TMP_SCRIPT"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP_SCRIPT" %q
else
  echo "ERR: curl or wget is required" >&2
  exit 127
fi
chmod +x "$TMP_SCRIPT"
if [ "$(id -u)" -eq 0 ]; then
  bash "$TMP_SCRIPT" update
elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
  sudo -n bash "$TMP_SCRIPT" update
else
  echo "ERR: 需要 root 权限或免密 sudo 才能在线更新后端服务" >&2
  exit 1
fi`, scriptURL, scriptURL)
}

func stripANSI(s string) string {
	return ansiEscapeRegex.ReplaceAllString(s, "")
}
