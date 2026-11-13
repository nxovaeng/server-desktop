package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DialContextFunc matches http.Transport.DialContext (such as Tunnel.DialContext).
type DialContextFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// APIError represents a structured backend error response.
type APIError struct {
	HTTPStatus int
	Code       ErrorCode
	Message    string
	Details    map[string]interface{}
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s (HTTP %d)", e.Code, e.Message, e.HTTPStatus)
	}
	return fmt.Sprintf("%s (HTTP %d)", e.Message, e.HTTPStatus)
}

// Client is a strongly typed REST client aligned with xray-panel-server OpenAPI 3.0.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient constructs a direct HTTP/HTTPS client or tunneled client when dialCtx != nil.
func NewClient(baseURL, token string, dialCtx DialContextFunc) *Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     60 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	if dialCtx != nil {
		transport.DialContext = dialCtx
	}

	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   strings.TrimSpace(token),
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   180 * time.Second,
		},
	}
}

func doEnvelope[T any](ctx context.Context, c *Client, method, path string, reqBody interface{}) (T, error) {
	var zero T

	var bodyReader io.Reader
	if reqBody != nil {
		raw, err := json.Marshal(reqBody)
		if err != nil {
			return zero, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return zero, fmt.Errorf("create request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errEnv ErrorResponse
		if json.Unmarshal(respBytes, &errEnv) == nil && errEnv.Error.Code != "" {
			return zero, &APIError{
				HTTPStatus: resp.StatusCode,
				Code:       errEnv.Error.Code,
				Message:    errEnv.Error.Message,
				Details:    errEnv.Error.Details,
			}
		}
		return zero, &APIError{
			HTTPStatus: resp.StatusCode,
			Message:    strings.TrimSpace(string(respBytes)),
		}
	}

	var env Envelope[T]
	if err := json.Unmarshal(respBytes, &env); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	if !env.Success {
		if env.Error != nil {
			return zero, &APIError{
				HTTPStatus: resp.StatusCode,
				Code:       env.Error.Code,
				Message:    env.Error.Message,
				Details:    env.Error.Details,
			}
		}
		// A success:false envelope without error details must never be
		// treated as a successful call returning the zero value.
		return zero, &APIError{
			HTTPStatus: resp.StatusCode,
			Code:       ErrInternal,
			Message:    "backend returned success:false without error details",
		}
	}
	return env.Data, nil
}

// Health calls GET /api/v1/health (unauthenticated).
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/health", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{HTTPStatus: resp.StatusCode, Message: string(raw)}
	}
	var out HealthResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// WaitForHealthy polls GET /api/v1/health until the server responds OK or ctx/timeout expires.
func (c *Client) WaitForHealthy(ctx context.Context, interval time.Duration) (*HealthResponse, error) {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if h, err := c.Health(ctx); err == nil && h.Status == "ok" {
			return h, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// Auth & Tokens
func (c *Client) VerifyAuth(ctx context.Context) (AuthVerifyResponse, error) {
	return doEnvelope[AuthVerifyResponse](ctx, c, http.MethodGet, "/api/v1/auth/verify", nil)
}
func (c *Client) ListTokens(ctx context.Context) (TokenListResponse, error) {
	return doEnvelope[TokenListResponse](ctx, c, http.MethodGet, "/api/v1/tokens", nil)
}
func (c *Client) CreateToken(ctx context.Context, req CreateTokenRequest) (CreateTokenResponse, error) {
	return doEnvelope[CreateTokenResponse](ctx, c, http.MethodPost, "/api/v1/tokens", req)
}
func (c *Client) DeleteToken(ctx context.Context, id string, purge bool) (TokenDeleteResponse, error) {
	q := ""
	if purge {
		q = "?purge=true"
	}
	return doEnvelope[TokenDeleteResponse](ctx, c, http.MethodDelete, "/api/v1/tokens/"+url.PathEscape(id)+q, nil)
}

// Dashboard
func (c *Client) GetDashboard(ctx context.Context) (DashboardResponse, error) {
	return doEnvelope[DashboardResponse](ctx, c, http.MethodGet, "/api/v1/dashboard", nil)
}

// Users
func (c *Client) ListUsers(ctx context.Context, query string) (UserListResponse, error) {
	path := "/api/v1/users"
	if query != "" {
		path += "?q=" + url.QueryEscape(query)
	}
	return doEnvelope[UserListResponse](ctx, c, http.MethodGet, path, nil)
}
func (c *Client) GetUser(ctx context.Context, id string) (UserDTO, error) {
	return doEnvelope[UserDTO](ctx, c, http.MethodGet, "/api/v1/users/"+url.PathEscape(id), nil)
}
func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (UserDTO, error) {
	return doEnvelope[UserDTO](ctx, c, http.MethodPost, "/api/v1/users", req)
}
func (c *Client) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (UserDTO, error) {
	return doEnvelope[UserDTO](ctx, c, http.MethodPut, "/api/v1/users/"+url.PathEscape(id), req)
}
func (c *Client) DeleteUser(ctx context.Context, id string) (DeleteResponse, error) {
	return doEnvelope[DeleteResponse](ctx, c, http.MethodDelete, "/api/v1/users/"+url.PathEscape(id), nil)
}
func (c *Client) ToggleUser(ctx context.Context, id string) (UserDTO, error) {
	return doEnvelope[UserDTO](ctx, c, http.MethodPost, "/api/v1/users/"+url.PathEscape(id)+"/toggle", nil)
}
func (c *Client) ResetUserTraffic(ctx context.Context, id string) (UserDTO, error) {
	return doEnvelope[UserDTO](ctx, c, http.MethodPost, "/api/v1/users/"+url.PathEscape(id)+"/reset-traffic", nil)
}

// Inbounds
func (c *Client) ListInbounds(ctx context.Context) (InboundListResponse, error) {
	return doEnvelope[InboundListResponse](ctx, c, http.MethodGet, "/api/v1/inbounds", nil)
}
func (c *Client) GetInbound(ctx context.Context, id string) (InboundDTO, error) {
	return doEnvelope[InboundDTO](ctx, c, http.MethodGet, "/api/v1/inbounds/"+url.PathEscape(id), nil)
}
func (c *Client) CreateInbound(ctx context.Context, req UpsertInboundRequest) (InboundDTO, error) {
	return doEnvelope[InboundDTO](ctx, c, http.MethodPost, "/api/v1/inbounds", req)
}
func (c *Client) UpdateInbound(ctx context.Context, id string, req UpsertInboundRequest) (InboundDTO, error) {
	return doEnvelope[InboundDTO](ctx, c, http.MethodPut, "/api/v1/inbounds/"+url.PathEscape(id), req)
}
func (c *Client) DeleteInbound(ctx context.Context, id string) (DeleteResponse, error) {
	return doEnvelope[DeleteResponse](ctx, c, http.MethodDelete, "/api/v1/inbounds/"+url.PathEscape(id), nil)
}
func (c *Client) ToggleInbound(ctx context.Context, id string) (InboundDTO, error) {
	return doEnvelope[InboundDTO](ctx, c, http.MethodPost, "/api/v1/inbounds/"+url.PathEscape(id)+"/toggle", nil)
}
func (c *Client) GetInboundWGClient(ctx context.Context, id string) (InboundWGClientConfigResponse, error) {
	return doEnvelope[InboundWGClientConfigResponse](ctx, c, http.MethodGet, "/api/v1/inbounds/"+url.PathEscape(id)+"/wg-client", nil)
}

// Outbounds
func (c *Client) ListOutbounds(ctx context.Context) (OutboundListResponse, error) {
	return doEnvelope[OutboundListResponse](ctx, c, http.MethodGet, "/api/v1/outbounds", nil)
}
func (c *Client) GetOutbound(ctx context.Context, id string) (OutboundDTO, error) {
	return doEnvelope[OutboundDTO](ctx, c, http.MethodGet, "/api/v1/outbounds/"+url.PathEscape(id), nil)
}
func (c *Client) CreateOutbound(ctx context.Context, req UpsertOutboundRequest) (OutboundDTO, error) {
	return doEnvelope[OutboundDTO](ctx, c, http.MethodPost, "/api/v1/outbounds", req)
}
func (c *Client) UpdateOutbound(ctx context.Context, id string, req UpsertOutboundRequest) (OutboundDTO, error) {
	return doEnvelope[OutboundDTO](ctx, c, http.MethodPut, "/api/v1/outbounds/"+url.PathEscape(id), req)
}
func (c *Client) DeleteOutbound(ctx context.Context, id string) (DeleteResponse, error) {
	return doEnvelope[DeleteResponse](ctx, c, http.MethodDelete, "/api/v1/outbounds/"+url.PathEscape(id), nil)
}
func (c *Client) ToggleOutbound(ctx context.Context, id string) (OutboundDTO, error) {
	return doEnvelope[OutboundDTO](ctx, c, http.MethodPost, "/api/v1/outbounds/"+url.PathEscape(id)+"/toggle", nil)
}
func (c *Client) TestOutbound(ctx context.Context, id string) (OutboundTestResult, error) {
	return doEnvelope[OutboundTestResult](ctx, c, http.MethodPost, "/api/v1/outbounds/"+url.PathEscape(id)+"/test", nil)
}
func (c *Client) ImportOutbound(ctx context.Context, link string) (OutboundDTO, error) {
	return doEnvelope[OutboundDTO](ctx, c, http.MethodPost, "/api/v1/outbounds/import", ImportShareLinkRequest{Link: link})
}
func (c *Client) ParseWireGuardOutbound(ctx context.Context, configText string) (ParsedWireGuardConfig, error) {
	return doEnvelope[ParsedWireGuardConfig](ctx, c, http.MethodPost, "/api/v1/outbounds/parse-wireguard", ParseWireGuardConfigRequest{Config: configText})
}
func (c *Client) GenerateOutboundWGKeys(ctx context.Context) (WGKeyPair, error) {
	return doEnvelope[WGKeyPair](ctx, c, http.MethodPost, "/api/v1/outbounds/generate-wg-keys", nil)
}

// Routing
func (c *Client) ListRouting(ctx context.Context) (RoutingRuleListResponse, error) {
	return doEnvelope[RoutingRuleListResponse](ctx, c, http.MethodGet, "/api/v1/routing", nil)
}
func (c *Client) GetRouting(ctx context.Context, id string) (RoutingRuleDTO, error) {
	return doEnvelope[RoutingRuleDTO](ctx, c, http.MethodGet, "/api/v1/routing/"+url.PathEscape(id), nil)
}
func (c *Client) CreateRouting(ctx context.Context, req UpsertRoutingRuleRequest) (RoutingRuleDTO, error) {
	return doEnvelope[RoutingRuleDTO](ctx, c, http.MethodPost, "/api/v1/routing", req)
}
func (c *Client) UpdateRouting(ctx context.Context, id string, req UpsertRoutingRuleRequest) (RoutingRuleDTO, error) {
	return doEnvelope[RoutingRuleDTO](ctx, c, http.MethodPut, "/api/v1/routing/"+url.PathEscape(id), req)
}
func (c *Client) DeleteRouting(ctx context.Context, id string) (DeleteResponse, error) {
	return doEnvelope[DeleteResponse](ctx, c, http.MethodDelete, "/api/v1/routing/"+url.PathEscape(id), nil)
}
func (c *Client) ToggleRouting(ctx context.Context, id string) (RoutingRuleDTO, error) {
	return doEnvelope[RoutingRuleDTO](ctx, c, http.MethodPost, "/api/v1/routing/"+url.PathEscape(id)+"/toggle", nil)
}
func (c *Client) ImportRoutingPreset(ctx context.Context, preset string) (ImportPresetResponse, error) {
	return doEnvelope[ImportPresetResponse](ctx, c, http.MethodPost, "/api/v1/routing/preset/"+url.PathEscape(preset), nil)
}
func (c *Client) GetGeoData(ctx context.Context) (GeoDataInfo, error) {
	return doEnvelope[GeoDataInfo](ctx, c, http.MethodGet, "/api/v1/routing/geodata", nil)
}

// Domains
func (c *Client) ListDomains(ctx context.Context) (DomainListResponse, error) {
	return doEnvelope[DomainListResponse](ctx, c, http.MethodGet, "/api/v1/domains", nil)
}
func (c *Client) GetDomain(ctx context.Context, id string) (DomainDTO, error) {
	return doEnvelope[DomainDTO](ctx, c, http.MethodGet, "/api/v1/domains/"+url.PathEscape(id), nil)
}
func (c *Client) CreateDomain(ctx context.Context, req UpsertDomainRequest) (DomainDTO, error) {
	return doEnvelope[DomainDTO](ctx, c, http.MethodPost, "/api/v1/domains", req)
}
func (c *Client) UpdateDomain(ctx context.Context, id string, req UpsertDomainRequest) (DomainDTO, error) {
	return doEnvelope[DomainDTO](ctx, c, http.MethodPut, "/api/v1/domains/"+url.PathEscape(id), req)
}
func (c *Client) DeleteDomain(ctx context.Context, id string) (DeleteResponse, error) {
	return doEnvelope[DeleteResponse](ctx, c, http.MethodDelete, "/api/v1/domains/"+url.PathEscape(id), nil)
}
func (c *Client) ScanCerts(ctx context.Context) (ScanCertificatesResponse, error) {
	return doEnvelope[ScanCertificatesResponse](ctx, c, http.MethodGet, "/api/v1/domains/scan-certs", nil)
}
func (c *Client) ScanImportCerts(ctx context.Context) (ScanImportCertificatesResponse, error) {
	return doEnvelope[ScanImportCertificatesResponse](ctx, c, http.MethodPost, "/api/v1/domains/scan-import", nil)
}
func (c *Client) ImportSingleCert(ctx context.Context, req ImportCertRequest) (DomainDTO, error) {
	return doEnvelope[DomainDTO](ctx, c, http.MethodPost, "/api/v1/domains/import-cert", req)
}
func (c *Client) UploadDomainCert(ctx context.Context, req UploadCertRequest) (DomainDTO, error) {
	return doEnvelope[DomainDTO](ctx, c, http.MethodPost, "/api/v1/domains/upload-cert", req)
}
func (c *Client) IssueDomainCert(ctx context.Context, req IssueCertRequest) (IssueCertResponse, error) {
	return doEnvelope[IssueCertResponse](ctx, c, http.MethodPost, "/api/v1/domains/issue-cert", req)
}

// WireGuard
func (c *Client) GetWGStatus(ctx context.Context) (WGServiceStatus, error) {
	return doEnvelope[WGServiceStatus](ctx, c, http.MethodGet, "/api/v1/wireguard/status", nil)
}
func (c *Client) SetupWGEnv(ctx context.Context) (WGSetupEnvResponse, error) {
	return doEnvelope[WGSetupEnvResponse](ctx, c, http.MethodPost, "/api/v1/wireguard/setup-env", nil)
}
func (c *Client) WGServiceAction(ctx context.Context, action string) (WGServiceActionResponse, error) {
	return doEnvelope[WGServiceActionResponse](ctx, c, http.MethodPost, "/api/v1/wireguard/service/"+url.PathEscape(action), nil)
}
func (c *Client) GetWGServer(ctx context.Context) (WGServerConfigDTO, error) {
	return doEnvelope[WGServerConfigDTO](ctx, c, http.MethodGet, "/api/v1/wireguard/server", nil)
}
func (c *Client) UpdateWGServer(ctx context.Context, req UpdateWGServerRequest) (WGServerConfigDTO, error) {
	return doEnvelope[WGServerConfigDTO](ctx, c, http.MethodPut, "/api/v1/wireguard/server", req)
}
func (c *Client) ListWGPeers(ctx context.Context) (WGPeerListResponse, error) {
	return doEnvelope[WGPeerListResponse](ctx, c, http.MethodGet, "/api/v1/wireguard/peers", nil)
}
func (c *Client) GetWGPeer(ctx context.Context, id uint) (WGPeerDTO, error) {
	return doEnvelope[WGPeerDTO](ctx, c, http.MethodGet, fmt.Sprintf("/api/v1/wireguard/peers/%d", id), nil)
}
func (c *Client) CreateWGPeer(ctx context.Context, req UpsertWGPeerRequest) (WGPeerDTO, error) {
	return doEnvelope[WGPeerDTO](ctx, c, http.MethodPost, "/api/v1/wireguard/peers", req)
}
func (c *Client) UpdateWGPeer(ctx context.Context, id uint, req UpsertWGPeerRequest) (WGPeerDTO, error) {
	return doEnvelope[WGPeerDTO](ctx, c, http.MethodPut, fmt.Sprintf("/api/v1/wireguard/peers/%d", id), req)
}
func (c *Client) DeleteWGPeer(ctx context.Context, id uint) (DeleteResponse, error) {
	return doEnvelope[DeleteResponse](ctx, c, http.MethodDelete, fmt.Sprintf("/api/v1/wireguard/peers/%d", id), nil)
}
func (c *Client) ToggleWGPeer(ctx context.Context, id uint) (WGPeerDTO, error) {
	return doEnvelope[WGPeerDTO](ctx, c, http.MethodPost, fmt.Sprintf("/api/v1/wireguard/peers/%d/toggle", id), nil)
}
func (c *Client) ExportWGPeerConfig(ctx context.Context, id uint) (WGPeerExportConfigResponse, error) {
	return doEnvelope[WGPeerExportConfigResponse](ctx, c, http.MethodGet, fmt.Sprintf("/api/v1/wireguard/peers/%d/config", id), nil)
}
func (c *Client) GenerateWGKeys(ctx context.Context) (WGKeyPair, error) {
	return doEnvelope[WGKeyPair](ctx, c, http.MethodPost, "/api/v1/wireguard/generate-keys", nil)
}

// Xray, Settings, Panel & Logs
func (c *Client) GetXrayStatus(ctx context.Context) (XrayStatusResponse, error) {
	return doEnvelope[XrayStatusResponse](ctx, c, http.MethodGet, "/api/v1/xray/status", nil)
}
func (c *Client) RestartXray(ctx context.Context) (XrayRestartResponse, error) {
	return doEnvelope[XrayRestartResponse](ctx, c, http.MethodPost, "/api/v1/xray/restart", nil)
}
func (c *Client) GetXrayConfig(ctx context.Context) (XrayConfigResponse, error) {
	return doEnvelope[XrayConfigResponse](ctx, c, http.MethodGet, "/api/v1/xray/config", nil)
}
func (c *Client) ApplyXrayConfig(ctx context.Context, hot, force bool) (XrayApplyResponse, error) {
	var q []string
	if hot {
		q = append(q, "hot=true")
	}
	if force {
		q = append(q, "force=true")
	}
	query := ""
	if len(q) > 0 {
		query = "?" + strings.Join(q, "&")
	}
	return doEnvelope[XrayApplyResponse](ctx, c, http.MethodPost, "/api/v1/xray/apply"+query, XrayApplyRequest{Hot: hot, Force: force})
}
func (c *Client) GetSettings(ctx context.Context) (SettingsDTO, error) {
	return doEnvelope[SettingsDTO](ctx, c, http.MethodGet, "/api/v1/settings", nil)
}
func (c *Client) UpdateSettings(ctx context.Context, req UpdateSettingsRequest) (SettingsDTO, error) {
	return doEnvelope[SettingsDTO](ctx, c, http.MethodPut, "/api/v1/settings", req)
}
func (c *Client) RestartPanel(ctx context.Context) (PanelRestartResponse, error) {
	return doEnvelope[PanelRestartResponse](ctx, c, http.MethodPost, "/api/v1/panel/restart", nil)
}
func (c *Client) GetLogs(ctx context.Context, service string, lines int) (LogsResponse, error) {
	path := fmt.Sprintf("/api/v1/logs?service=%s&lines=%d", url.QueryEscape(service), lines)
	return doEnvelope[LogsResponse](ctx, c, http.MethodGet, path, nil)
}

// Subscription & API Nginx and Cloudflare Tunnel
func (c *Client) ApplySubscriptionNginx(ctx context.Context) (SubscriptionNginxResponse, error) {
	return doEnvelope[SubscriptionNginxResponse](ctx, c, http.MethodPost, "/api/v1/subscription/nginx", nil)
}
func (c *Client) DisableSubscriptionNginx(ctx context.Context) (SubscriptionNginxResponse, error) {
	return doEnvelope[SubscriptionNginxResponse](ctx, c, http.MethodDelete, "/api/v1/subscription/nginx", nil)
}
func (c *Client) ApplyAPINginx(ctx context.Context) (APINginxResponse, error) {
	return doEnvelope[APINginxResponse](ctx, c, http.MethodPost, "/api/v1/panel/nginx", nil)
}
func (c *Client) DisableAPINginx(ctx context.Context) (APINginxResponse, error) {
	return doEnvelope[APINginxResponse](ctx, c, http.MethodDelete, "/api/v1/panel/nginx", nil)
}
func (c *Client) GetSpeedtestStatus(ctx context.Context) (SpeedtestStatus, error) {
	return doEnvelope[SpeedtestStatus](ctx, c, http.MethodGet, "/api/v1/speedtest", nil)
}

func (c *Client) UpdateSpeedtest(ctx context.Context, req UpdateSpeedtestRequest) (SpeedtestStatus, error) {
	return doEnvelope[SpeedtestStatus](ctx, c, http.MethodPut, "/api/v1/speedtest", req)
}

func (c *Client) StartSpeedtest(ctx context.Context) (SpeedtestStatus, error) {
	return doEnvelope[SpeedtestStatus](ctx, c, http.MethodPost, "/api/v1/speedtest/start", nil)
}

func (c *Client) StopSpeedtest(ctx context.Context) (SpeedtestStatus, error) {
	return doEnvelope[SpeedtestStatus](ctx, c, http.MethodPost, "/api/v1/speedtest/stop", nil)
}

func (c *Client) GetCFTunnelStatus(ctx context.Context) (CFTunnelStatus, error) {
	return doEnvelope[CFTunnelStatus](ctx, c, http.MethodGet, "/api/v1/tunnels/cf/status", nil)
}
func (c *Client) InstallCFTunnel(ctx context.Context) (CFTunnelInstallResponse, error) {
	return doEnvelope[CFTunnelInstallResponse](ctx, c, http.MethodPost, "/api/v1/tunnels/cf/install", nil)
}
func (c *Client) SetupCFTunnel(ctx context.Context, token string) (CFTunnelActionResponse, error) {
	return doEnvelope[CFTunnelActionResponse](ctx, c, http.MethodPost, "/api/v1/tunnels/cf/setup", CFTunnelSetupRequest{TunnelToken: token})
}
func (c *Client) CFTunnelServiceAction(ctx context.Context, action string) (CFTunnelActionResponse, error) {
	return doEnvelope[CFTunnelActionResponse](ctx, c, http.MethodPost, "/api/v1/tunnels/cf/service/"+url.PathEscape(action), nil)
}
