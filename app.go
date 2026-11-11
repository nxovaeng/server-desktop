package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"xray-panel-desktop/internal/apiclient"
	"xray-panel-desktop/internal/local"
	"xray-panel-desktop/internal/session"
	"xray-panel-desktop/internal/store"
)

// App exposes Wails Go bindings to the Vue 3 frontend.
type App struct {
	ctx      context.Context
	profiles *store.ProfileStore
	session  *session.Manager
}

// NewApp creates a new App instance.
func NewApp() *App {
	cfgDir, err := store.DefaultConfigDir()
	if err != nil {
		cfgDir = filepath.Join(os.TempDir(), "xray-panel-desktop")
	}
	ps := store.NewProfileStore(filepath.Join(cfgDir, "profiles.json"), newOSKeyringVault())
	return &App{
		profiles: ps,
		session:  session.NewManager(ps),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.session.Disconnect()
}

func (a *App) callCtx() (context.Context, context.CancelFunc) {
	return a.timeoutCtx(25 * time.Second)
}

func (a *App) timeoutCtx(d time.Duration) (context.Context, context.CancelFunc) {
	base := a.ctx
	if base == nil {
		base = context.Background()
	}
	return context.WithTimeout(base, d)
}

func (a *App) client() (*apiclient.Client, error) {
	return a.session.Client()
}

// ===== Profile & Session Bindings =====

type ProfileListResult struct {
	Profiles        []store.ServerProfile `json:"profiles"`
	ActiveProfileID string                `json:"active_profile_id"`
}

func (a *App) ListProfiles() (*ProfileListResult, error) {
	items, activeID, err := a.profiles.List()
	if err != nil {
		return nil, err
	}
	return &ProfileListResult{Profiles: items, ActiveProfileID: activeID}, nil
}

func (a *App) SaveProfile(profile store.ServerProfile, secrets store.ProfileSecrets) (*store.ServerProfile, error) {
	return a.profiles.Save(profile, &secrets)
}

func (a *App) DeleteProfile(id string) error {
	if st := a.session.Status(); st.Connected && st.ProfileID == id {
		a.session.Disconnect()
	}
	return a.profiles.Delete(id)
}

func (a *App) ConnectProfile(id string, trustTOFU bool) (*session.ConnectionStatus, error) {
	ctx, cancel := a.callCtx()
	defer cancel()
	return a.session.Connect(ctx, id, trustTOFU)
}

func (a *App) Disconnect() {
	a.session.Disconnect()
}

func (a *App) GetConnectionStatus() session.ConnectionStatus {
	return a.session.Status()
}

func (a *App) TestLatency() (int64, error) {
	ctx, cancel := a.callCtx()
	defer cancel()
	return a.session.TestLatency(ctx)
}

// ===== Local Helpers & Native File Dialog =====

func (a *App) GenerateLocalWGKeys() (*local.WGKeyPair, error) {
	return local.GenerateWGKeyPair()
}

func (a *App) DeriveLocalWGPublicKey(privKey string) (string, error) {
	return local.DeriveWGPublicKey(privKey)
}

func (a *App) ValidateDirectURL(raw string) (*local.URLCheckResult, error) {
	return local.ValidateDirectURL(raw)
}

func (a *App) SaveTextFile(defaultFilename, content string) (string, error) {
	if a.ctx == nil {
		return "", nil
	}
	selected, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Title:           "导出配置文件",
	})
	if err != nil || selected == "" {
		return "", err
	}
	if err := os.WriteFile(selected, []byte(content), 0600); err != nil {
		return "", err
	}
	return selected, nil
}

// ===== Auth & Tokens =====

func (a *App) VerifyAuth() (apiclient.AuthVerifyResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.AuthVerifyResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.VerifyAuth(ctx)
}

func (a *App) ListTokens() (apiclient.TokenListResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.TokenListResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ListTokens(ctx)
}

func (a *App) CreateToken(req apiclient.CreateTokenRequest) (apiclient.CreateTokenResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.CreateTokenResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.CreateToken(ctx, req)
}

func (a *App) DeleteToken(id string, purge bool) (apiclient.TokenDeleteResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.TokenDeleteResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DeleteToken(ctx, id, purge)
}

// ===== Dashboard =====

func (a *App) GetDashboard() (apiclient.DashboardResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DashboardResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetDashboard(ctx)
}

// ===== Users =====

func (a *App) ListUsers(query string) (apiclient.UserListResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.UserListResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ListUsers(ctx, query)
}

func (a *App) GetUser(id string) (apiclient.UserDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.UserDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetUser(ctx, id)
}

func (a *App) CreateUser(req apiclient.CreateUserRequest) (apiclient.UserDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.UserDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.CreateUser(ctx, req)
}

func (a *App) UpdateUser(id string, req apiclient.UpdateUserRequest) (apiclient.UserDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.UserDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UpdateUser(ctx, id, req)
}

func (a *App) DeleteUser(id string) (apiclient.DeleteResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DeleteResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DeleteUser(ctx, id)
}

func (a *App) ToggleUser(id string) (apiclient.UserDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.UserDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ToggleUser(ctx, id)
}

func (a *App) ResetUserTraffic(id string) (apiclient.UserDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.UserDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ResetUserTraffic(ctx, id)
}

// ===== Inbounds =====

func (a *App) ListInbounds() (apiclient.InboundListResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.InboundListResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ListInbounds(ctx)
}

func (a *App) GetInbound(id string) (apiclient.InboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.InboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetInbound(ctx, id)
}

func (a *App) CreateInbound(req apiclient.UpsertInboundRequest) (apiclient.InboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.InboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.CreateInbound(ctx, req)
}

func (a *App) UpdateInbound(id string, req apiclient.UpsertInboundRequest) (apiclient.InboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.InboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UpdateInbound(ctx, id, req)
}

func (a *App) DeleteInbound(id string) (apiclient.DeleteResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DeleteResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DeleteInbound(ctx, id)
}

func (a *App) ToggleInbound(id string) (apiclient.InboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.InboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ToggleInbound(ctx, id)
}

func (a *App) GetInboundWGClient(id string) (apiclient.InboundWGClientConfigResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.InboundWGClientConfigResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetInboundWGClient(ctx, id)
}

// ===== Outbounds =====

func (a *App) ListOutbounds() (apiclient.OutboundListResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.OutboundListResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ListOutbounds(ctx)
}

func (a *App) GetOutbound(id string) (apiclient.OutboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.OutboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetOutbound(ctx, id)
}

func (a *App) CreateOutbound(req apiclient.UpsertOutboundRequest) (apiclient.OutboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.OutboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.CreateOutbound(ctx, req)
}

func (a *App) UpdateOutbound(id string, req apiclient.UpsertOutboundRequest) (apiclient.OutboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.OutboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UpdateOutbound(ctx, id, req)
}

func (a *App) DeleteOutbound(id string) (apiclient.DeleteResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DeleteResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DeleteOutbound(ctx, id)
}

func (a *App) ToggleOutbound(id string) (apiclient.OutboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.OutboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ToggleOutbound(ctx, id)
}

func (a *App) TestOutbound(id string) (apiclient.OutboundTestResult, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.OutboundTestResult{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.TestOutbound(ctx, id)
}

func (a *App) ImportOutbound(link string) (apiclient.OutboundDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.OutboundDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ImportOutbound(ctx, link)
}

func (a *App) ParseWireGuardOutbound(cfg string) (apiclient.ParsedWireGuardConfig, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.ParsedWireGuardConfig{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ParseWireGuardOutbound(ctx, cfg)
}

func (a *App) GenerateOutboundWGKeys() (apiclient.WGKeyPair, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGKeyPair{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GenerateOutboundWGKeys(ctx)
}

// ===== Routing =====

func (a *App) ListRouting() (apiclient.RoutingRuleListResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.RoutingRuleListResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ListRouting(ctx)
}

func (a *App) GetRouting(id string) (apiclient.RoutingRuleDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.RoutingRuleDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetRouting(ctx, id)
}

func (a *App) CreateRouting(req apiclient.UpsertRoutingRuleRequest) (apiclient.RoutingRuleDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.RoutingRuleDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.CreateRouting(ctx, req)
}

func (a *App) UpdateRouting(id string, req apiclient.UpsertRoutingRuleRequest) (apiclient.RoutingRuleDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.RoutingRuleDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UpdateRouting(ctx, id, req)
}

func (a *App) DeleteRouting(id string) (apiclient.DeleteResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DeleteResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DeleteRouting(ctx, id)
}

func (a *App) ToggleRouting(id string) (apiclient.RoutingRuleDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.RoutingRuleDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ToggleRouting(ctx, id)
}

func (a *App) ImportRoutingPreset(preset string) (apiclient.ImportPresetResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.ImportPresetResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ImportRoutingPreset(ctx, preset)
}

func (a *App) GetGeoData() (apiclient.GeoDataInfo, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.GeoDataInfo{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetGeoData(ctx)
}

// ===== Domains =====

func (a *App) ListDomains() (apiclient.DomainListResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DomainListResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ListDomains(ctx)
}

func (a *App) GetDomain(id string) (apiclient.DomainDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DomainDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetDomain(ctx, id)
}

func (a *App) CreateDomain(req apiclient.UpsertDomainRequest) (apiclient.DomainDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DomainDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.CreateDomain(ctx, req)
}

func (a *App) UpdateDomain(id string, req apiclient.UpsertDomainRequest) (apiclient.DomainDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DomainDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UpdateDomain(ctx, id, req)
}

func (a *App) DeleteDomain(id string) (apiclient.DeleteResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DeleteResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DeleteDomain(ctx, id)
}

func (a *App) ScanCerts() (apiclient.ScanCertificatesResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.ScanCertificatesResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ScanCerts(ctx)
}

func (a *App) ScanImportCerts() (apiclient.ScanImportCertificatesResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.ScanImportCertificatesResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ScanImportCerts(ctx)
}

func (a *App) ImportSingleCert(req apiclient.ImportCertRequest) (apiclient.DomainDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DomainDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ImportSingleCert(ctx, req)
}

func (a *App) UploadDomainCert(req apiclient.UploadCertRequest) (apiclient.DomainDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DomainDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UploadDomainCert(ctx, req)
}

func (a *App) IssueDomainCert(req apiclient.IssueCertRequest) (apiclient.IssueCertResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.IssueCertResponse{}, err
	}
	ctx, cancel := a.timeoutCtx(180 * time.Second)
	defer cancel()
	return c.IssueDomainCert(ctx, req)
}

// ===== WireGuard =====

func (a *App) GetWGStatus() (apiclient.WGServiceStatus, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGServiceStatus{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetWGStatus(ctx)
}

func (a *App) SetupWGEnv() (apiclient.WGSetupEnvResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGSetupEnvResponse{}, err
	}
	base := a.ctx
	if base == nil {
		base = context.Background()
	}
	ctx, cancel := context.WithTimeout(base, 120*time.Second)
	defer cancel()
	return c.SetupWGEnv(ctx)
}

func (a *App) WGServiceAction(action string) (apiclient.WGServiceActionResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGServiceActionResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.WGServiceAction(ctx, action)
}

func (a *App) GetWGServer() (apiclient.WGServerConfigDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGServerConfigDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetWGServer(ctx)
}

func (a *App) UpdateWGServer(req apiclient.UpdateWGServerRequest) (apiclient.WGServerConfigDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGServerConfigDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UpdateWGServer(ctx, req)
}

func (a *App) ListWGPeers() (apiclient.WGPeerListResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGPeerListResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ListWGPeers(ctx)
}

func (a *App) GetWGPeer(id uint) (apiclient.WGPeerDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGPeerDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetWGPeer(ctx, id)
}

func (a *App) CreateWGPeer(req apiclient.UpsertWGPeerRequest) (apiclient.WGPeerDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGPeerDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.CreateWGPeer(ctx, req)
}

func (a *App) UpdateWGPeer(id uint, req apiclient.UpsertWGPeerRequest) (apiclient.WGPeerDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGPeerDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UpdateWGPeer(ctx, id, req)
}

func (a *App) DeleteWGPeer(id uint) (apiclient.DeleteResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.DeleteResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DeleteWGPeer(ctx, id)
}

func (a *App) ToggleWGPeer(id uint) (apiclient.WGPeerDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGPeerDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ToggleWGPeer(ctx, id)
}

func (a *App) ExportWGPeerConfig(id uint) (apiclient.WGPeerExportConfigResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGPeerExportConfigResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ExportWGPeerConfig(ctx, id)
}

func (a *App) GenerateWGKeys() (apiclient.WGKeyPair, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.WGKeyPair{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GenerateWGKeys(ctx)
}

// ===== Xray, Settings, Panel & Logs =====

func (a *App) GetXrayStatus() (apiclient.XrayStatusResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.XrayStatusResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetXrayStatus(ctx)
}

func (a *App) RestartXray() (apiclient.XrayRestartResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.XrayRestartResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.RestartXray(ctx)
}

func (a *App) GetXrayConfig() (apiclient.XrayConfigResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.XrayConfigResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetXrayConfig(ctx)
}

func (a *App) ApplyXrayConfig(hot bool) (apiclient.XrayApplyResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.XrayApplyResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ApplyXrayConfig(ctx, hot)
}

func (a *App) GetSettings() (apiclient.SettingsDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.SettingsDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetSettings(ctx)
}

func (a *App) UpdateSettings(req apiclient.UpdateSettingsRequest) (apiclient.SettingsDTO, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.SettingsDTO{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.UpdateSettings(ctx, req)
}

func (a *App) RestartPanelAndReconnect() (*session.ConnectionStatus, error) {
	base := a.ctx
	if base == nil {
		base = context.Background()
	}
	ctx, cancel := context.WithTimeout(base, 40*time.Second)
	defer cancel()
	return a.session.RestartPanelAndReconnect(ctx)
}

func (a *App) GetLogs(serviceName string, lines int) (apiclient.LogsResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.LogsResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetLogs(ctx, serviceName, lines)
}

// ===== Subscription Nginx & Cloudflare Tunnel =====

func (a *App) ApplySubscriptionNginx() (apiclient.SubscriptionNginxResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.SubscriptionNginxResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ApplySubscriptionNginx(ctx)
}

func (a *App) DisableSubscriptionNginx() (apiclient.SubscriptionNginxResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.SubscriptionNginxResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DisableSubscriptionNginx(ctx)
}

func (a *App) ApplyAPINginx() (apiclient.APINginxResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.APINginxResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.ApplyAPINginx(ctx)
}

func (a *App) DisableAPINginx() (apiclient.APINginxResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.APINginxResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.DisableAPINginx(ctx)
}

func (a *App) GetCFTunnelStatus() (apiclient.CFTunnelStatus, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.CFTunnelStatus{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.GetCFTunnelStatus(ctx)
}

func (a *App) InstallCFTunnel() (apiclient.CFTunnelInstallResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.CFTunnelInstallResponse{}, err
	}
	ctx, cancel := a.timeoutCtx(120 * time.Second)
	defer cancel()
	return c.InstallCFTunnel(ctx)
}

func (a *App) SetupCFTunnel(tunnelToken string) (apiclient.CFTunnelActionResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.CFTunnelActionResponse{}, err
	}
	ctx, cancel := a.timeoutCtx(60 * time.Second)
	defer cancel()
	return c.SetupCFTunnel(ctx, tunnelToken)
}

func (a *App) CFTunnelServiceAction(action string) (apiclient.CFTunnelActionResponse, error) {
	c, err := a.client()
	if err != nil {
		return apiclient.CFTunnelActionResponse{}, err
	}
	ctx, cancel := a.callCtx()
	defer cancel()
	return c.CFTunnelServiceAction(ctx, action)
}
