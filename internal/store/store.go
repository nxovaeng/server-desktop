package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConnectionMode defines how the desktop app connects to xray-panel-server.
type ConnectionMode string

const (
	ModeSSHTunnel ConnectionMode = "ssh_tunnel"
	ModeDirect    ConnectionMode = "direct"
)

// SSHAuthType defines how SSH authentication is performed.
type SSHAuthType string

const (
	SSHAuthKeyFile  SSHAuthType = "key_file"
	SSHAuthKeyPEM   SSHAuthType = "key_pem"
	SSHAuthPassword SSHAuthType = "password"
	SSHAuthAgent    SSHAuthType = "agent"
)

// ServerProfile holds non-sensitive metadata for a remote xray-panel-server node.
// Sensitive credentials (API token, SSH password, key PEM, passphrase) are stored in SecretVault.
type ServerProfile struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Mode              ConnectionMode `json:"mode"`
	DirectURL         string         `json:"direct_url,omitempty"`
	AllowInsecureHTTP bool           `json:"allow_insecure_http,omitempty"`

	SSHHost         string      `json:"ssh_host,omitempty"`
	SSHPort         int         `json:"ssh_port,omitempty"`
	SSHUser         string      `json:"ssh_user,omitempty"`
	SSHAuthType     SSHAuthType `json:"ssh_auth_type,omitempty"`
	SSHKeyPath      string      `json:"ssh_key_path,omitempty"`
	SSHFingerprint  string      `json:"ssh_fingerprint,omitempty"`
	RemoteBindHost  string      `json:"remote_bind_host,omitempty"`
	RemotePanelPort int         `json:"remote_panel_port,omitempty"`

	HasAPIToken         bool       `json:"has_api_token"`
	HasSSHPassword      bool       `json:"has_ssh_password"`
	HasSSHKeyPEM        bool       `json:"has_ssh_key_pem"`
	HasSSHKeyPassphrase bool       `json:"has_ssh_key_passphrase"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	LastConnectedAt     *time.Time `json:"last_connected_at,omitempty"`
}

// ProfileSecrets holds sensitive credentials stored in OS Keyring.
type ProfileSecrets struct {
	APIToken         string `json:"api_token,omitempty"`
	SSHPassword      string `json:"ssh_password,omitempty"`
	SSHKeyPEM        string `json:"ssh_key_pem,omitempty"`
	SSHKeyPassphrase string `json:"ssh_key_passphrase,omitempty"`
}

// SecretVault abstracts OS Keyring storage for testability and cross-platform support.
type SecretVault interface {
	Get(service, account string) (string, error)
	Set(service, account, secret string) error
	Delete(service, account string) error
}

// ErrSecretNotFound indicates no secret was found in the vault.
var ErrSecretNotFound = errors.New("secret not found")

const keyringServiceName = "xray-panel-desktop"

// ProfileStore manages server profiles on disk and secrets in SecretVault.
type ProfileStore struct {
	mu       sync.RWMutex
	filePath string
	vault    SecretVault
}

// DefaultConfigDir returns the OS-specific user config directory for xray-panel-desktop.
func DefaultConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "xray-panel-desktop"), nil
}

// NewProfileStore creates a ProfileStore backed by filePath and vault.
func NewProfileStore(filePath string, vault SecretVault) *ProfileStore {
	return &ProfileStore{
		filePath: filePath,
		vault:    vault,
	}
}

type persistedFile struct {
	ActiveProfileID string          `json:"active_profile_id"`
	Profiles        []ServerProfile `json:"profiles"`
}

// List returns all saved server profiles and the last active profile ID.
func (s *ProfileStore) List() ([]ServerProfile, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.readFileLocked()
	if err != nil {
		return nil, "", err
	}
	for i := range data.Profiles {
		s.populateSecretFlags(&data.Profiles[i])
	}
	return data.Profiles, data.ActiveProfileID, nil
}

// Get returns a single profile and its associated secrets from the vault.
func (s *ProfileStore) Get(id string) (*ServerProfile, *ProfileSecrets, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := s.readFileLocked()
	if err != nil {
		return nil, nil, err
	}

	for _, p := range data.Profiles {
		if p.ID == id {
			s.populateSecretFlags(&p)
			sec := s.loadSecrets(p.ID)
			return &p, sec, nil
		}
	}
	return nil, nil, fmt.Errorf("profile %q not found", id)
}

// Save creates or updates a profile and writes non-empty secrets into SecretVault.
// Passing empty string for a secret on update preserves any previously stored value.
func (s *ProfileStore) Save(profile ServerProfile, secrets *ProfileSecrets) (*ServerProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile.Name = strings.TrimSpace(profile.Name)
	if profile.Name == "" {
		return nil, errors.New("profile name is required")
	}
	if profile.Mode == "" {
		profile.Mode = ModeSSHTunnel
	}
	if profile.Mode == ModeSSHTunnel {
		if strings.TrimSpace(profile.SSHHost) == "" {
			return nil, errors.New("ssh_host is required in ssh_tunnel mode")
		}
		if profile.SSHPort <= 0 {
			profile.SSHPort = 22
		}
		if strings.TrimSpace(profile.SSHUser) == "" {
			profile.SSHUser = "root"
		}
		if profile.SSHAuthType == "" {
			profile.SSHAuthType = SSHAuthKeyFile
		}
		if strings.TrimSpace(profile.RemoteBindHost) == "" {
			profile.RemoteBindHost = "127.0.0.1"
		}
		if profile.RemotePanelPort <= 0 {
			profile.RemotePanelPort = 8082
		}
	} else if profile.Mode == ModeDirect {
		if strings.TrimSpace(profile.DirectURL) == "" {
			return nil, errors.New("direct_url is required in direct mode")
		}
	} else {
		return nil, fmt.Errorf("unsupported connection mode: %s", profile.Mode)
	}

	data, err := s.readFileLocked()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if profile.ID == "" {
		profile.ID = uuid.New().String()
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now

	found := false
	for i := range data.Profiles {
		if data.Profiles[i].ID == profile.ID {
			if profile.CreatedAt.IsZero() {
				profile.CreatedAt = data.Profiles[i].CreatedAt
			}
			if profile.SSHFingerprint == "" {
				profile.SSHFingerprint = data.Profiles[i].SSHFingerprint
			}
			data.Profiles[i] = profile
			found = true
			break
		}
	}
	if !found {
		if profile.CreatedAt.IsZero() {
			profile.CreatedAt = now
		}
		data.Profiles = append(data.Profiles, profile)
	}

	if secrets != nil {
		if err := s.saveSecrets(profile.ID, secrets); err != nil {
			return nil, err
		}
	}

	if data.ActiveProfileID == "" {
		data.ActiveProfileID = profile.ID
	}

	if err := s.writeFileLocked(data); err != nil {
		return nil, err
	}

	s.populateSecretFlags(&profile)
	return &profile, nil
}

// Delete removes a profile and all its stored secrets from the vault.
func (s *ProfileStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFileLocked()
	if err != nil {
		return err
	}

	filtered := make([]ServerProfile, 0, len(data.Profiles))
	found := false
	for _, p := range data.Profiles {
		if p.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, p)
	}
	if !found {
		return fmt.Errorf("profile %q not found", id)
	}

	data.Profiles = filtered
	if data.ActiveProfileID == id {
		data.ActiveProfileID = ""
		if len(filtered) > 0 {
			data.ActiveProfileID = filtered[0].ID
		}
	}

	s.deleteSecrets(id)
	return s.writeFileLocked(data)
}

// SetActive records the currently selected profile ID.
func (s *ProfileStore) SetActive(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFileLocked()
	if err != nil {
		return err
	}
	data.ActiveProfileID = id
	return s.writeFileLocked(data)
}

// UpdateSSHFingerprint persists a verified TOFU SSH host key fingerprint for a profile.
func (s *ProfileStore) UpdateSSHFingerprint(id, fingerprint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFileLocked()
	if err != nil {
		return err
	}
	for i := range data.Profiles {
		if data.Profiles[i].ID == id {
			data.Profiles[i].SSHFingerprint = fingerprint
			now := time.Now().UTC()
			data.Profiles[i].LastConnectedAt = &now
			data.Profiles[i].UpdatedAt = now
			return s.writeFileLocked(data)
		}
	}
	return fmt.Errorf("profile %q not found", id)
}

func (s *ProfileStore) secretKey(profileID, field string) string {
	return profileID + ":" + field
}

func (s *ProfileStore) populateSecretFlags(p *ServerProfile) {
	if s.vault == nil {
		return
	}
	p.HasAPIToken = s.hasSecret(p.ID, "api_token")
	p.HasSSHPassword = s.hasSecret(p.ID, "ssh_password")
	p.HasSSHKeyPEM = s.hasSecret(p.ID, "ssh_key_pem")
	p.HasSSHKeyPassphrase = s.hasSecret(p.ID, "ssh_key_passphrase")
}

func (s *ProfileStore) hasSecret(profileID, field string) bool {
	val, err := s.vault.Get(keyringServiceName, s.secretKey(profileID, field))
	return err == nil && val != ""
}

func (s *ProfileStore) loadSecrets(profileID string) *ProfileSecrets {
	sec := &ProfileSecrets{}
	if s.vault == nil {
		return sec
	}
	sec.APIToken, _ = s.vault.Get(keyringServiceName, s.secretKey(profileID, "api_token"))
	sec.SSHPassword, _ = s.vault.Get(keyringServiceName, s.secretKey(profileID, "ssh_password"))
	sec.SSHKeyPEM, _ = s.vault.Get(keyringServiceName, s.secretKey(profileID, "ssh_key_pem"))
	sec.SSHKeyPassphrase, _ = s.vault.Get(keyringServiceName, s.secretKey(profileID, "ssh_key_passphrase"))
	return sec
}

func (s *ProfileStore) saveSecrets(profileID string, sec *ProfileSecrets) error {
	if s.vault == nil {
		return nil
	}
	pairs := []struct {
		field string
		val   string
	}{
		{"api_token", sec.APIToken},
		{"ssh_password", sec.SSHPassword},
		{"ssh_key_pem", sec.SSHKeyPEM},
		{"ssh_key_passphrase", sec.SSHKeyPassphrase},
	}
	for _, p := range pairs {
		if p.val != "" {
			if err := s.vault.Set(keyringServiceName, s.secretKey(profileID, p.field), p.val); err != nil {
				return fmt.Errorf("failed to store %s in keyring: %w", p.field, err)
			}
		}
	}
	return nil
}

func (s *ProfileStore) deleteSecrets(profileID string) {
	if s.vault == nil {
		return
	}
	for _, field := range []string{"api_token", "ssh_password", "ssh_key_pem", "ssh_key_passphrase"} {
		_ = s.vault.Delete(keyringServiceName, s.secretKey(profileID, field))
	}
}

func (s *ProfileStore) readFileLocked() (*persistedFile, error) {
	raw, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return &persistedFile{Profiles: []ServerProfile{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var data persistedFile
	if len(strings.TrimSpace(string(raw))) == 0 {
		return &persistedFile{Profiles: []ServerProfile{}}, nil
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if data.Profiles == nil {
		data.Profiles = []ServerProfile{}
	}
	return &data, nil
}

func (s *ProfileStore) writeFileLocked(data *persistedFile) error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.filePath + ".tmp"
	if err := os.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.filePath)
}

// MemoryVault is an in-memory SecretVault implementation for unit tests.
type MemoryVault struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryVault() *MemoryVault {
	return &MemoryVault{data: make(map[string]string)}
}

func (v *MemoryVault) Get(service, account string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	val, ok := v.data[service+"\x00"+account]
	if !ok {
		return "", ErrSecretNotFound
	}
	return val, nil
}

func (v *MemoryVault) Set(service, account, secret string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.data[service+"\x00"+account] = secret
	return nil
}

func (v *MemoryVault) Delete(service, account string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.data, service+"\x00"+account)
	return nil
}
