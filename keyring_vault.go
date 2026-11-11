package main

import (
	"errors"

	"github.com/zalando/go-keyring"

	"xray-panel-desktop/internal/store"
)

// osKeyringVault implements store.SecretVault backed by the native OS credential store
// (Windows Credential Manager, macOS Keychain, Linux Secret Service).
type osKeyringVault struct{}

func newOSKeyringVault() store.SecretVault {
	return &osKeyringVault{}
}

func (v *osKeyringVault) Get(service, account string) (string, error) {
	val, err := keyring.Get(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", store.ErrSecretNotFound
	}
	return val, err
}

func (v *osKeyringVault) Set(service, account, secret string) error {
	return keyring.Set(service, account, secret)
}

func (v *osKeyringVault) Delete(service, account string) error {
	err := keyring.Delete(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
