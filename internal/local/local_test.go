package local_test

import (
	"testing"

	"xray-panel-desktop/internal/local"
)

func TestLocalWGKeyGenerationAndDerivation(t *testing.T) {
	kp, err := local.GenerateWGKeyPair()
	if err != nil {
		t.Fatalf("GenerateWGKeyPair: %v", err)
	}
	pub, err := local.DeriveWGPublicKey(kp.PrivateKey)
	if err != nil {
		t.Fatalf("DeriveWGPublicKey: %v", err)
	}
	if pub != kp.PublicKey {
		t.Fatalf("derived pubkey %q != %q", pub, kp.PublicKey)
	}
}

func TestValidateDirectURL(t *testing.T) {
	res, err := local.ValidateDirectURL("panel.example.com")
	if err != nil {
		t.Fatalf("ValidateDirectURL: %v", err)
	}
	if res.NormalizedURL != "https://panel.example.com" || res.RequiresHTTPSWarn {
		t.Fatalf("unexpected HTTPS result: %+v", res)
	}

	plainPublic, err := local.ValidateDirectURL("http://panel.example.com:8082/")
	if err != nil {
		t.Fatalf("ValidateDirectURL plain public: %v", err)
	}
	if !plainPublic.RequiresHTTPSWarn || plainPublic.NormalizedURL != "http://panel.example.com:8082" {
		t.Fatalf("expected HTTPS warning on public plaintext HTTP: %+v", plainPublic)
	}

	plainLocal, err := local.ValidateDirectURL("http://127.0.0.1:8082")
	if err != nil {
		t.Fatalf("ValidateDirectURL loopback: %v", err)
	}
	if plainLocal.RequiresHTTPSWarn || !plainLocal.IsLoopback {
		t.Fatalf("loopback HTTP should not trigger public HTTPS warning: %+v", plainLocal)
	}
}
