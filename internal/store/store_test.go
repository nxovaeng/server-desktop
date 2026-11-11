package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"xray-panel-desktop/internal/store"
)

func TestProfileStoreSeparatesMetadataAndSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "profiles.json")
	vault := store.NewMemoryVault()
	ps := store.NewProfileStore(profilePath, vault)

	saved, err := ps.Save(store.ServerProfile{
		Name:        "Tokyo VPS",
		Mode:        store.ModeSSHTunnel,
		SSHHost:     "1.2.3.4",
		SSHPort:     22,
		SSHUser:     "root",
		SSHAuthType: store.SSHAuthKeyPEM,
	}, &store.ProfileSecrets{
		APIToken:         "xrp_secret_token_123456",
		SSHKeyPEM:        "-----BEGIN PRIVATE KEY-----\nSECRET\n-----END PRIVATE KEY-----",
		SSHKeyPassphrase: "my-passphrase",
	})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if !saved.HasAPIToken || !saved.HasSSHKeyPEM || !saved.HasSSHKeyPassphrase || saved.HasSSHPassword {
		t.Fatalf("unexpected secret presence flags: %+v", saved)
	}

	// Verify JSON on disk contains NO plaintext secrets
	raw, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	diskContent := string(raw)
	for _, secret := range []string{"xrp_secret_token_123456", "BEGIN PRIVATE KEY", "my-passphrase"} {
		if strings.Contains(diskContent, secret) {
			t.Fatalf("plaintext secret %q leaked to disk JSON: %s", secret, diskContent)
		}
	}

	// Update profile leaving secrets empty -> must preserve existing secrets in vault
	_, err = ps.Save(store.ServerProfile{
		ID:          saved.ID,
		Name:        "Tokyo VPS (Renamed)",
		Mode:        store.ModeSSHTunnel,
		SSHHost:     "1.2.3.4",
		SSHPort:     2222,
		SSHUser:     "ubuntu",
		SSHAuthType: store.SSHAuthKeyPEM,
	}, &store.ProfileSecrets{})
	if err != nil {
		t.Fatalf("Update Save failed: %v", err)
	}

	got, gotSec, err := ps.Get(saved.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != "Tokyo VPS (Renamed)" || got.SSHPort != 2222 {
		t.Fatalf("unexpected updated metadata: %+v", got)
	}
	if gotSec.APIToken != "xrp_secret_token_123456" || gotSec.SSHKeyPassphrase != "my-passphrase" {
		t.Fatalf("secrets were not preserved on blank update: %+v", gotSec)
	}

	// Update TOFU fingerprint
	if err := ps.UpdateSSHFingerprint(saved.ID, "SHA256:abc123"); err != nil {
		t.Fatalf("UpdateSSHFingerprint: %v", err)
	}
	got, _, _ = ps.Get(saved.ID)
	if got.SSHFingerprint != "SHA256:abc123" || got.LastConnectedAt == nil {
		t.Fatalf("expected persisted fingerprint and timestamp, got %+v", got)
	}

	// Delete removes profile and secrets
	if err := ps.Delete(saved.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, _, err := ps.Get(saved.ID); err == nil {
		t.Fatalf("expected error getting deleted profile")
	}
	if _, err := vault.Get("xray-panel-desktop", saved.ID+":api_token"); err == nil {
		t.Fatalf("expected keyring secret to be deleted")
	}
}
