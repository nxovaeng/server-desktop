package session_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"xray-panel-desktop/internal/apiclient"
	"xray-panel-desktop/internal/session"
	"xray-panel-desktop/internal/store"
)

func TestSessionManagerDirectSwitchAndValidation(t *testing.T) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health":
			_ = json.NewEncoder(w).Encode(apiclient.HealthResponse{Status: "ok", Version: "1.0.0"})
		case "/api/v1/auth/verify":
			_ = json.NewEncoder(w).Encode(apiclient.Envelope[apiclient.AuthVerifyResponse]{
				Success: true,
				Data: apiclient.AuthVerifyResponse{
					Valid: true, TokenID: "t1", TokenName: "work", Version: "1.0.0",
					ServerPublicHost: "node1.example.com", SubPath: "/d", PanelTitle: "Node 1",
				},
			})
		}
	}))
	defer srv1.Close()

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health":
			_ = json.NewEncoder(w).Encode(apiclient.HealthResponse{Status: "ok", Version: "1.1.0"})
		case "/api/v1/auth/verify":
			_ = json.NewEncoder(w).Encode(apiclient.Envelope[apiclient.AuthVerifyResponse]{
				Success: true,
				Data: apiclient.AuthVerifyResponse{
					Valid: true, TokenID: "t2", TokenName: "mac", Version: "1.1.0",
					ServerPublicHost: "node2.example.com", SubPath: "/vip", PanelTitle: "Node 2",
				},
			})
		}
	}))
	defer srv2.Close()

	vault := store.NewMemoryVault()
	ps := store.NewProfileStore(filepath.Join(t.TempDir(), "profiles.json"), vault)

	p1, err := ps.Save(store.ServerProfile{
		Name: "Node 1", Mode: store.ModeDirect, DirectURL: srv1.URL,
	}, &store.ProfileSecrets{APIToken: "xrp_tok_1"})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := ps.Save(store.ServerProfile{
		Name: "Node 2", Mode: store.ModeDirect, DirectURL: srv2.URL,
	}, &store.ProfileSecrets{APIToken: "xrp_tok_2"})
	if err != nil {
		t.Fatal(err)
	}

	// Non-loopback plaintext HTTP must be blocked unless AllowInsecureHTTP=true
	pPlain, err := ps.Save(store.ServerProfile{
		Name: "Unsafe Public", Mode: store.ModeDirect, DirectURL: "http://example.com:8082",
	}, &store.ProfileSecrets{APIToken: "xrp_tok_3"})
	if err != nil {
		t.Fatal(err)
	}

	mgr := session.NewManager(ps)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := mgr.Connect(ctx, pPlain.ID, false); err == nil {
		t.Fatalf("expected plaintext public HTTP direct connection to be blocked")
	}

	st1, err := mgr.Connect(ctx, p1.ID, false)
	if err != nil || !st1.Connected || st1.ServerVersion != "1.0.0" || st1.PanelTitle != "Node 1" {
		t.Fatalf("Connect p1 failed: %+v err=%v", st1, err)
	}

	if rtt, err := mgr.TestLatency(ctx); err != nil || rtt < 0 {
		t.Fatalf("TestLatency failed: rtt=%d err=%v", rtt, err)
	}

	// Switch to Node 2
	st2, err := mgr.Connect(ctx, p2.ID, false)
	if err != nil || !st2.Connected || st2.ServerVersion != "1.1.0" || st2.SubPath != "/vip" {
		t.Fatalf("Switch to p2 failed: %+v err=%v", st2, err)
	}

	_, activeID, _ := ps.List()
	if activeID != p2.ID {
		t.Fatalf("expected active profile %s, got %s", p2.ID, activeID)
	}

	mgr.Disconnect()
	if mgr.Status().Connected {
		t.Fatalf("expected disconnected after Disconnect()")
	}
}
