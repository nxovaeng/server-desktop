package tunnel_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"xray-panel-desktop/internal/apiclient"
	"xray-panel-desktop/internal/tunnel"
)

type directTCPIPData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

func startMockSSHServer(t *testing.T, clientPub ssh.PublicKey) (addr string, hostFP string, closeFn func()) {
	t.Helper()

	_, hostPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostPriv)
	if err != nil {
		t.Fatal(err)
	}
	hostFP = ssh.FingerprintSHA256(hostSigner.PublicKey())

	cfg := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if string(key.Marshal()) == string(clientPub.Marshal()) {
				return &ssh.Permissions{}, nil
			}
			return nil, errors.New("unauthorized public key")
		},
		KeyboardInteractiveCallback: func(c ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
			answers, err := challenge("", "", []string{"Password:"}, []bool{false})
			if err != nil {
				return nil, err
			}
			if len(answers) == 1 && answers[0] == "test-password" {
				return &ssh.Permissions{}, nil
			}
			return nil, errors.New("wrong password")
		},
	}
	cfg.AddHostKey(hostSigner)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	go func() {
		for {
			rawConn, err := ln.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				sshConn, chans, reqs, err := ssh.NewServerConn(c, cfg)
				if err != nil {
					return
				}
				defer sshConn.Close()

				go func() {
					for req := range reqs {
						if req.WantReply {
							_ = req.Reply(true, nil)
						}
					}
				}()

				for newChan := range chans {
					if newChan.ChannelType() != "direct-tcpip" {
						_ = newChan.Reject(ssh.UnknownChannelType, "unsupported")
						continue
					}
					var d directTCPIPData
					if err := ssh.Unmarshal(newChan.ExtraData(), &d); err != nil {
						_ = newChan.Reject(ssh.ConnectionFailed, err.Error())
						continue
					}
					targetConn, err := net.Dial("tcp", net.JoinHostPort(d.DestAddr, strconv.Itoa(int(d.DestPort))))
					if err != nil {
						_ = newChan.Reject(ssh.ConnectionFailed, err.Error())
						continue
					}
					ch, chReqs, err := newChan.Accept()
					if err != nil {
						_ = targetConn.Close()
						continue
					}
					go ssh.DiscardRequests(chReqs)
					go func() {
						_, _ = io.Copy(ch, targetConn)
						_ = ch.CloseWrite()
					}()
					go func() {
						_, _ = io.Copy(targetConn, ch)
						_ = targetConn.Close()
					}()
				}
			}(rawConn)
		}
	}()

	return ln.Addr().String(), hostFP, func() {
		close(stopCh)
		_ = ln.Close()
		wg.Wait()
	}
}

func generateED25519KeyPEM(t *testing.T) (string, ssh.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pemBlock, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		t.Fatal(err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(pemBlock)), sshPub
}

func TestSSHTunnelTOFUAndInProcessHTTPClient(t *testing.T) {
	// Ensure unused binary import isn't left (or remove)
	_ = binary.BigEndian

	clientKeyPEM, clientPub := generateED25519KeyPEM(t)
	sshAddr, hostFP, stopSSH := startMockSSHServer(t, clientPub)
	defer stopSSH()

	sshHost, sshPortStr, _ := net.SplitHostPort(sshAddr)
	sshPort, _ := strconv.Atoi(sshPortStr)

	// Start mock xray-panel-server HTTP backend on random localhost port
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health":
			_ = json.NewEncoder(w).Encode(apiclient.HealthResponse{Status: "ok", Version: "1.0.0"})
		case "/api/v1/auth/verify":
			if r.Header.Get("Authorization") != "Bearer xrp_test_token" {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(apiclient.ErrorResponse{
					Success: false,
					Error:   apiclient.ErrorDetail{Code: apiclient.ErrUnauthorized, Message: "bad token"},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(apiclient.Envelope[apiclient.AuthVerifyResponse]{
				Success: true,
				Data: apiclient.AuthVerifyResponse{
					Valid:            true,
					TokenID:          "tok-1",
					TokenName:        "desktop",
					TokenPrefix:      "xrp_test_tok...",
					Version:          "1.0.0",
					ServerPublicHost: "vps.example.com",
					SubPath:          "/d",
					PanelTitle:       "Xray Panel",
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer backend.Close()

	// 1. First connection without TrustOnFirstUse must return TOFURequiredError with exact SHA-256 fingerprint
	_, err := tunnel.Connect(tunnel.Config{
		Host:            sshHost,
		Port:            sshPort,
		User:            "root",
		AuthMethod:      tunnel.AuthKeyPEM,
		KeyPEM:          clientKeyPEM,
		TrustOnFirstUse: false,
	})
	var tofuErr *tunnel.TOFURequiredError
	if !errors.As(err, &tofuErr) {
		t.Fatalf("expected TOFURequiredError, got %v", err)
	}
	if tofuErr.Fingerprint != hostFP {
		t.Fatalf("expected TOFU fingerprint %s, got %s", hostFP, tofuErr.Fingerprint)
	}

	// 2. Mismatched fingerprint must be rejected (MITM protection)
	_, err = tunnel.Connect(tunnel.Config{
		Host:                sshHost,
		Port:                sshPort,
		User:                "root",
		AuthMethod:          tunnel.AuthKeyPEM,
		KeyPEM:              clientKeyPEM,
		ExpectedFingerprint: "SHA256:definitely-wrong-fingerprint",
	})
	var mismatchErr *tunnel.HostKeyMismatchError
	if !errors.As(err, &mismatchErr) {
		t.Fatalf("expected HostKeyMismatchError, got %v", err)
	}

	// 3. Trusted connection establishes in-process tunnel and forwards HTTP over sshClient.Dial
	var capturedFP string
	tun, err := tunnel.Connect(tunnel.Config{
		Host:              sshHost,
		Port:              sshPort,
		User:              "root",
		AuthMethod:        tunnel.AuthKeyPEM,
		KeyPEM:            clientKeyPEM,
		TrustOnFirstUse:   true,
		KeepaliveInterval: 500 * time.Millisecond,
		OnFirstFingerprint: func(fp string) {
			capturedFP = fp
		},
	})
	if err != nil {
		t.Fatalf("tunnel.Connect: %v", err)
	}
	defer tun.Close()

	if capturedFP != hostFP || tun.Fingerprint() != hostFP || !tun.IsConnected() {
		t.Fatalf("unexpected tunnel state: captured=%s fp=%s connected=%v", capturedFP, tun.Fingerprint(), tun.IsConnected())
	}

	// Build apiclient routing through sshClient.DialContext (no local listening port!)
	api := apiclient.NewClient(backend.URL, "xrp_test_token", tun.DialContext)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := api.Health(ctx)
	if err != nil || health.Status != "ok" {
		t.Fatalf("Health via SSH tunnel failed: %+v err=%v", health, err)
	}

	verify, err := api.VerifyAuth(ctx)
	if err != nil || !verify.Valid || verify.TokenName != "desktop" || verify.ServerPublicHost != "vps.example.com" {
		t.Fatalf("VerifyAuth via SSH tunnel failed: %+v err=%v", verify, err)
	}

	// Invalid token should surface typed APIError
	badAPI := apiclient.NewClient(backend.URL, "xrp_wrong", tun.DialContext)
	if _, err := badAPI.VerifyAuth(ctx); err == nil {
		t.Fatalf("expected APIError for wrong token")
	} else {
		var apiErr *apiclient.APIError
		if !errors.As(err, &apiErr) || apiErr.Code != apiclient.ErrUnauthorized {
			t.Fatalf("expected ErrUnauthorized APIError, got %v", err)
		}
	}
}

func TestSSHTunnelKeyboardInteractive(t *testing.T) {
	_, clientPub := generateED25519KeyPEM(t)
	sshAddr, _, stopSSH := startMockSSHServer(t, clientPub)
	defer stopSSH()

	sshHost, sshPortStr, _ := net.SplitHostPort(sshAddr)
	sshPort, _ := strconv.Atoi(sshPortStr)

	// Correct password answers the keyboard-interactive challenge.
	tun, err := tunnel.Connect(tunnel.Config{
		Host:              sshHost,
		Port:              sshPort,
		User:              "root",
		AuthMethod:        tunnel.AuthKeyboardInteractive,
		Password:          "test-password",
		TrustOnFirstUse:   true,
		KeepaliveInterval: 500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("keyboard-interactive connect failed: %v", err)
	}
	if !tun.IsConnected() {
		t.Fatal("tunnel should be connected after keyboard-interactive auth")
	}
	_ = tun.Close()

	// Wrong password must fail authentication.
	_, err = tunnel.Connect(tunnel.Config{
		Host:            sshHost,
		Port:            sshPort,
		User:            "root",
		AuthMethod:      tunnel.AuthKeyboardInteractive,
		Password:        "wrong-password",
		TrustOnFirstUse: true,
	})
	if err == nil {
		t.Fatal("expected error for wrong keyboard-interactive password")
	}
}
