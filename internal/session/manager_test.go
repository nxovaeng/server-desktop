package session_test

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
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

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

func TestManagerUpdateBackendViaSSH(t *testing.T) {
	// 1. Direct-mode connection must reject SSH update
	directSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/health":
			_ = json.NewEncoder(w).Encode(apiclient.HealthResponse{Status: "ok", Version: "1.0.0"})
		case "/api/v1/auth/verify":
			_ = json.NewEncoder(w).Encode(apiclient.Envelope[apiclient.AuthVerifyResponse]{
				Success: true,
				Data:    apiclient.AuthVerifyResponse{Valid: true, TokenID: "t1", TokenName: "direct", Version: "1.0.0"},
			})
		}
	}))
	defer directSrv.Close()

	vault := store.NewMemoryVault()
	ps := store.NewProfileStore(filepath.Join(t.TempDir(), "profiles.json"), vault)
	pDirect, err := ps.Save(store.ServerProfile{Name: "Direct", Mode: store.ModeDirect, DirectURL: directSrv.URL}, &store.ProfileSecrets{APIToken: "xrp_direct"})
	if err != nil {
		t.Fatal(err)
	}

	mgr := session.NewManager(ps)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := mgr.Connect(ctx, pDirect.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.UpdateBackendViaSSH(ctx); err == nil {
		t.Fatalf("expected direct connection mode to reject SSH backend update")
	}
	mgr.Disconnect()

	// 2. SSH-tunnel mode executes remote install.sh update and refreshes ServerVersion
	var version atomic.Value
	version.Store("1.0.0")

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := version.Load().(string)
		switch r.URL.Path {
		case "/api/v1/health":
			_ = json.NewEncoder(w).Encode(apiclient.HealthResponse{Status: "ok", Version: v})
		case "/api/v1/auth/verify":
			_ = json.NewEncoder(w).Encode(apiclient.Envelope[apiclient.AuthVerifyResponse]{
				Success: true,
				Data: apiclient.AuthVerifyResponse{
					Valid: true, TokenID: "t-ssh", TokenName: "ssh", Version: v, PanelTitle: "Panel",
				},
			})
		}
	}))
	defer apiSrv.Close()

	parsedURL, _ := url.Parse(apiSrv.URL)
	apiHost, apiPortStr, _ := net.SplitHostPort(parsedURL.Host)
	apiPort, _ := strconv.Atoi(apiPortStr)

	_, hostPriv, _ := ed25519.GenerateKey(rand.Reader)
	hostSigner, _ := ssh.NewSignerFromKey(hostPriv)
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	pemBlock, _ := ssh.MarshalPrivateKey(priv, "")
	clientKeyPEM := string(pem.EncodeToMemory(pemBlock))
	clientPub, _ := ssh.NewPublicKey(pub)

	var failNext atomic.Bool
	failNext.Store(true)
	var seenCmd atomic.Value

	sshCfg := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if string(key.Marshal()) == string(clientPub.Marshal()) {
				return &ssh.Permissions{}, nil
			}
			return nil, errors.New("unauthorized")
		},
	}
	sshCfg.AddHostKey(hostSigner)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			raw, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				sshConn, chans, reqs, err := ssh.NewServerConn(c, sshCfg)
				if err != nil {
					return
				}
				defer sshConn.Close()
				go ssh.DiscardRequests(reqs)
				for nc := range chans {
					switch nc.ChannelType() {
					case "direct-tcpip":
						var d struct {
							DestAddr   string
							DestPort   uint32
							OriginAddr string
							OriginPort uint32
						}
						_ = ssh.Unmarshal(nc.ExtraData(), &d)
						tgt, err := net.Dial("tcp", net.JoinHostPort(d.DestAddr, strconv.Itoa(int(d.DestPort))))
						if err != nil {
							_ = nc.Reject(ssh.ConnectionFailed, err.Error())
							continue
						}
						ch, chReqs, _ := nc.Accept()
						go ssh.DiscardRequests(chReqs)
						go func() { _, _ = io.Copy(ch, tgt); _ = ch.CloseWrite() }()
						go func() { _, _ = io.Copy(tgt, ch); _ = tgt.Close() }()
					case "session":
						ch, chReqs, _ := nc.Accept()
						go func() {
							for req := range chReqs {
								if req.Type == "exec" {
									var p struct{ Command string }
									_ = ssh.Unmarshal(req.Payload, &p)
									seenCmd.Store(p.Command)
									_ = req.Reply(true, nil)
									var exitCode uint32
									if failNext.Swap(false) {
										_, _ = io.WriteString(ch, "\x1b[31m[ERR]\x1b[0m download failed\n")
										exitCode = 1
									} else {
										version.Store("1.2.3")
										_, _ = io.WriteString(ch, "\x1b[32m[OK]\x1b[0m xray-panel-server 升级完成\n")
										exitCode = 0
									}
									var exitMsg [4]byte
									binary.BigEndian.PutUint32(exitMsg[:], exitCode)
									_, _ = ch.SendRequest("exit-status", false, exitMsg[:])
									_ = ch.CloseWrite()
									_ = ch.Close()
								}
							}
						}()
					default:
						_ = nc.Reject(ssh.UnknownChannelType, "unsupported")
					}
				}
			}(raw)
		}
	}()

	sshHost, sshPortStr, _ := net.SplitHostPort(ln.Addr().String())
	sshPort, _ := strconv.Atoi(sshPortStr)

	pSSH, err := ps.Save(store.ServerProfile{
		Name:            "SSH Node",
		Mode:            store.ModeSSHTunnel,
		SSHHost:         sshHost,
		SSHPort:         sshPort,
		SSHUser:         "root",
		SSHAuthType:     "key_pem",
		RemoteBindHost:  apiHost,
		RemotePanelPort: apiPort,
	}, &store.ProfileSecrets{
		APIToken:  "xrp_ssh",
		SSHKeyPEM: clientKeyPEM,
	})
	if err != nil {
		t.Fatal(err)
	}

	st, err := mgr.Connect(ctx, pSSH.ID, true)
	if err != nil || st.ServerVersion != "1.0.0" {
		t.Fatalf("SSH Connect failed: %+v err=%v", st, err)
	}

	// First run returns remote failure output with ANSI stripped
	failRes, err := mgr.UpdateBackendViaSSH(ctx)
	if err != nil || failRes.Success || failRes.ExitCode != 1 || failRes.Output != "[ERR] download failed" {
		t.Fatalf("expected structured remote failure with ANSI stripped, got res=%+v err=%v", failRes, err)
	}

	// Second run succeeds, verifies install.sh update command, and refreshes ServerVersion to 1.2.3
	okRes, err := mgr.UpdateBackendViaSSH(ctx)
	if err != nil || !okRes.Success || okRes.ExitCode != 0 || okRes.Output != "[OK] xray-panel-server 升级完成" {
		t.Fatalf("expected SSH update success, got res=%+v err=%v", okRes, err)
	}
	if okRes.Status.ServerVersion != "1.2.3" || mgr.Status().ServerVersion != "1.2.3" {
		t.Fatalf("expected refreshed ServerVersion 1.2.3, got res=%s mgr=%s", okRes.Status.ServerVersion, mgr.Status().ServerVersion)
	}
	cmdStr, _ := seenCmd.Load().(string)
	if !strings.Contains(cmdStr, session.DefaultBackendUpdateScriptURL) || !strings.Contains(cmdStr, "update") {
		t.Fatalf("expected update command to invoke install.sh update, got: %s", cmdStr)
	}
}
