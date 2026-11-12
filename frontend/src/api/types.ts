// TypeScript API contract interfaces aligned with xray-panel-server/api/openapi.yaml

export type ApiErrorCode =
  | 'UNAUTHORIZED'
  | 'RATE_LIMITED'
  | 'VALIDATION_ERROR'
  | 'NOT_FOUND'
  | 'CONFLICT'
  | 'FORBIDDEN'
  | 'XRAY_VALIDATION_FAILED'
  | 'SYSTEM_COMMAND_FAILED'
  | 'INTERNAL_ERROR'

export interface ApiErrorDetail {
  code: ApiErrorCode
  message: string
  details?: Record<string, unknown>
}

export interface ApiEnvelope<T> {
  success: boolean
  data?: T
  error?: ApiErrorDetail
}

export interface DeleteResponse {
  id: string
  deleted: boolean
}

export interface HealthResponse {
  status: string
  version: string
}

export interface AuthVerifyResponse {
  valid: boolean
  token_id: string
  token_name: string
  token_prefix: string
  version: string
  server_public_host: string
  sub_domain: string
  sub_path: string
  panel_title: string
}

export interface APITokenDTO {
  id: string
  name: string
  token_prefix: string
  expires_at: string | null
  revoked: boolean
  last_used_at: string | null
  last_ip: string
  created_at: string
}

export interface TokenListResponse {
  tokens: APITokenDTO[]
  total: number
}

export interface CreateTokenRequest {
  name: string
  expires_days?: number | null
}

export interface CreateTokenResponse {
  token: APITokenDTO
  plaintext_token: string
}

export interface TokenDeleteResponse {
  id: string
  revoked: boolean
  deleted: boolean
}

export interface DashboardCounts {
  total: number
  active: number
}

export interface DashboardTraffic {
  users_traffic_used_bytes: number
  monthly_net_sent_bytes: number
  monthly_net_recv_bytes: number
  monthly_net_total_bytes: number
  baseline_month: string
}

export interface DashboardSystemMetrics {
  hostname: string
  os: string
  platform: string
  uptime_seconds: number
  go_version: string
  cpu_model: string
  cpu_cores: number
  cpu_usage_percent: number
  mem_total_bytes: number
  mem_used_bytes: number
  mem_free_bytes: number
  mem_percent: number
  disk_total_bytes: number
  disk_used_bytes: number
  disk_free_bytes: number
  disk_percent: number
}

export interface DashboardServices {
  xray_active: boolean
  xray_status: string
  wireguard_installed: boolean
  wireguard_active: boolean
  wireguard_interface: string
  wireguard_online_peers: number
  wireguard_total_peers: number
}

export interface DashboardResponse {
  users: DashboardCounts
  inbounds: DashboardCounts
  outbounds: DashboardCounts
  domains: DashboardCounts
  routing_rules: DashboardCounts
  traffic: DashboardTraffic
  system: DashboardSystemMetrics
  services: DashboardServices
}

export interface UserSubURLs {
  base64: string
  plain: string
  json: string
  yaml: string
  relative_path: string
}

export interface UserDTO {
  id: string
  uuid: string
  name: string
  email: string
  traffic_limit_bytes: number
  traffic_used_bytes: number
  remaining_traffic_bytes: number
  traffic_reset_at: string | null
  expiry_date: string | null
  remaining_days: number
  enabled: boolean
  is_active: boolean
  sub_path: string
  sub_urls: UserSubURLs
  note: string
  wg_pub_key: string
  created_at: string
  updated_at: string
}

export interface UserListResponse {
  users: UserDTO[]
  total: number
}

export interface CreateUserRequest {
  name: string
  email?: string
  uuid?: string
  traffic_limit_bytes?: number
  expiry_date?: string | null
  expire_days?: number | null
  enabled?: boolean
  note?: string
  wg_pub_key?: string
}

export interface UpdateUserRequest {
  name: string
  email?: string
  uuid?: string
  traffic_limit_bytes?: number
  expiry_date?: string | null
  enabled?: boolean
  note?: string
  wg_pub_key?: string
}

export type InboundProtocol = 'vless' | 'trojan' | 'wireguard'
export type InboundTransport = 'xhttp' | 'grpc' | 'ws' | 'raw'
export type DomainType = 'direct' | 'cdn'

export interface DomainSummaryDTO {
  id: string
  domain: string
  type: DomainType
  is_wildcard: boolean
  enabled: boolean
}

export interface InboundDTO {
  id: string
  tag: string
  protocol: InboundProtocol
  transport: InboundTransport
  port: number
  listen: string
  domain_id: string
  domain: DomainSummaryDTO | null
  path: string
  service_name: string
  host: string
  mode: string
  actual_domain: string
  custom_sni: string
  connect_domain: string
  has_wg_secret_key: boolean
  wg_public_key: string
  wg_peer_pub_key: string
  wg_mtu: number
  wg_local_ip: string
  exclude_from_sub: boolean
  enabled: boolean
  use_uds: boolean
  socket_path: string
  remark: string
  created_at: string
  updated_at: string
}

export interface InboundListResponse {
  inbounds: InboundDTO[]
  total: number
}

export interface UpsertInboundRequest {
  tag?: string
  protocol?: InboundProtocol
  transport?: InboundTransport
  port?: number
  listen?: string
  domain_id?: string
  path?: string
  service_name?: string
  host?: string
  mode?: string
  custom_sni?: string
  connect_domain?: string
  wg_secret_key?: string
  wg_public_key?: string
  wg_peer_pub_key?: string
  wg_mtu?: number
  wg_local_ip?: string
  exclude_from_sub?: boolean
  use_uds?: boolean | null
  enabled?: boolean | null
  random_port?: boolean
  random_path?: boolean
  remark?: string
}

export interface InboundWGClientConfigResponse {
  inbound_id: string
  inbound_tag: string
  server_public_key: string
  server_host: string
  port: number
  endpoint: string
  peer_ip: string
  mtu: number
  wg_conf: string
}

export type OutboundType =
  | 'direct'
  | 'socks5'
  | 'wireguard'
  | 'trojan'
  | 'blackhole'
  | 'vless'
  | 'vmess'

export interface OutboundDTO {
  id: string
  tag: string
  type: OutboundType
  server: string
  port: number
  username: string
  has_password: boolean
  has_wg_secret_key: boolean
  wg_client_pub_key: string
  wg_public_key: string
  wg_reserved: string
  wg_local_ipv4: string
  wg_local_ipv6: string
  wg_mtu: number
  wg_dns: string
  has_trojan_password: boolean
  trojan_sni: string
  trojan_network: string
  uuid: string
  flow: string
  security: string
  network: string
  header_type: string
  request_host: string
  path: string
  service_name: string
  tls: boolean
  tls_server_name: string
  tls_alpn: string
  reality: boolean
  reality_pubkey: string
  reality_short_id: string
  reality_sni: string
  enabled: boolean
  priority: number
  remark: string
  created_at: string
  updated_at: string
}

export interface OutboundListResponse {
  outbounds: OutboundDTO[]
  total: number
}

export interface UpsertOutboundRequest {
  tag?: string
  type: OutboundType
  server?: string
  port?: number
  username?: string
  password?: string
  wg_secret_key?: string
  wg_client_pub_key?: string
  wg_public_key?: string
  wg_reserved?: string
  wg_local_ipv4?: string
  wg_local_ipv6?: string
  wg_mtu?: number
  wg_dns?: string
  trojan_password?: string
  trojan_sni?: string
  trojan_network?: string
  uuid?: string
  flow?: string
  security?: string
  network?: string
  header_type?: string
  request_host?: string
  path?: string
  service_name?: string
  tls?: boolean
  tls_server_name?: string
  tls_alpn?: string
  reality?: boolean
  reality_pubkey?: string
  reality_short_id?: string
  reality_sni?: string
  enabled?: boolean | null
  priority?: number
  remark?: string
}

export interface ParsedWireGuardConfig {
  private_key: string
  client_public_key: string
  public_key: string
  endpoint: string
  server: string
  port: number
  address: string
  local_ipv4: string
  local_ipv6: string
  dns: string
  mtu: number
}

export interface OutboundTestResult {
  success: boolean
  message: string
  latency_ms?: number
  endpoint?: string
  outbound_ip?: string
  location?: string
}

export interface WGKeyPair {
  private_key: string
  public_key: string
}

export type RuleType = 'inbound' | 'domain' | 'ip' | 'geosite' | 'geoip' | 'protocol'

export interface RoutingRuleDTO {
  id: string
  name: string
  type: RuleType
  inbound_tag: string
  domains: string
  ips: string
  geosite_tags: string
  geoip_codes: string
  protocols: string
  outbound_tag: string
  priority: number
  enabled: boolean
  remark: string
  created_at: string
  updated_at: string
}

export interface RoutingRuleListResponse {
  rules: RoutingRuleDTO[]
  total: number
}

export interface UpsertRoutingRuleRequest {
  name: string
  type: RuleType
  inbound_tag?: string
  domains?: string
  ips?: string
  geosite_tags?: string
  geoip_codes?: string
  protocols?: string
  outbound_tag: string
  priority?: number
  enabled?: boolean | null
  remark?: string
}

export interface ImportPresetResponse {
  preset: string
  imported_count: number
  skipped_count: number
  rules: RoutingRuleDTO[]
}

export interface GeoDataInfo {
  geosite_available: boolean
  geoip_available: boolean
  geosite_tags: string[]
  geoip_codes: string[]
}

export interface DomainDTO {
  id: string
  domain: string
  type: DomainType
  is_wildcard: boolean
  cert_path: string
  key_path: string
  enabled: boolean
  has_cert: boolean
  cert_expiry_date: string | null
  days_to_expiry: number | null
  cert_issuer: string
  created_at: string
  updated_at: string
}

export interface DomainListResponse {
  domains: DomainDTO[]
  total: number
}

export interface UpsertDomainRequest {
  domain: string
  type?: DomainType
  is_wildcard?: boolean
  cert_path?: string
  key_path?: string
  enabled?: boolean | null
}

export interface CertificateInfo {
  domain: string
  domains: string[]
  cert_path: string
  key_path: string
  expiry_date: string
  days_to_expiry: number
  status: 'valid' | 'expiring_soon' | 'expired' | 'unknown'
  issuer: string
  is_wildcard: boolean
  exists: boolean
}

export interface ScanCertificatesResponse {
  cert_dir: string
  certificates: CertificateInfo[]
  count: number
}

export interface ScanImportCertificatesResponse {
  cert_dir: string
  imported_count: number
  updated_count: number
  skipped_count: number
  errors: string[]
  domains: DomainDTO[]
}

export interface ImportCertRequest {
  domain: string
  cert_path: string
  key_path: string
}

export interface UploadCertRequest {
  domain: string
  type?: 'direct' | 'cdn'
  cert_pem: string
  key_pem: string
  enabled?: boolean | null
}

export interface IssueCertRequest {
  domain: string
  include_wildcard?: boolean
  type?: 'direct' | 'cdn'
  provider: 'cf' | 'luadns'
  cf_token?: string
  cf_key?: string
  cf_email?: string
  cf_zone_id?: string
  lua_key?: string
  lua_email?: string
  ca_server?: 'letsencrypt' | 'zerossl'
  email?: string
  force?: boolean
}

export interface IssueCertResponse {
  domain: DomainDTO
  provider: string
  ca_server: string
  reused_existing: boolean
  output: string
}

export interface WGServiceStatus {
  is_linux: boolean
  is_installed: boolean
  is_active: boolean
  is_enabled: boolean
  interface_name: string
  listen_port: number
  public_key: string
  address: string
  total_peers: number
  active_peers: number
  online_peers: number
  port_conflict: boolean
  conflict_inbound: string
  ipv4_forwarding: boolean
  ipv6_forwarding: boolean
  enable_nat: boolean
}

export interface WGSetupEnvResponse {
  success: boolean
  message: string
}

export interface WGServiceActionResponse {
  action: 'start' | 'stop' | 'restart' | 'enable' | 'disable'
  success: boolean
  status: WGServiceStatus
}

export interface WGServerConfigDTO {
  id: number
  interface_name: string
  has_private_key: boolean
  public_key: string
  listen_port: number
  address: string
  mtu: number
  enable_nat: boolean
  post_up: string
  post_down: string
  enabled: boolean
  updated_at: string
}

export interface UpdateWGServerRequest {
  interface_name?: string
  listen_port?: number
  address?: string
  mtu?: number
  private_key?: string
  enable_nat?: boolean | null
  post_up?: string
  post_down?: string
  enabled?: boolean | null
}

export interface WGPeerRuntimeStats {
  public_key: string
  endpoint: string
  allowed_ips: string
  latest_handshake: string | null
  handshake_ago: string
  transfer_rx: number
  transfer_tx: number
  transfer_rx_human: string
  transfer_tx_human: string
  is_online: boolean
}

export interface WGPeerDTO {
  id: number
  name: string
  public_key: string
  has_private_key: boolean
  has_preshared_key: boolean
  allowed_ips: string
  endpoint: string
  persistent_keepalive: number
  enabled: boolean
  runtime: WGPeerRuntimeStats
  created_at: string
  updated_at: string
}

export interface WGPeerListResponse {
  peers: WGPeerDTO[]
  next_available_ip: string
  total: number
}

export interface UpsertWGPeerRequest {
  name: string
  public_key?: string
  private_key?: string
  preshared_key?: string
  allowed_ips?: string
  endpoint?: string
  persistent_keepalive?: number
  enabled?: boolean | null
}

export interface WGPeerExportConfigResponse {
  peer_id: number
  peer_name: string
  server_host: string
  endpoint: string
  wg_conf: string
  xray_json: string
  has_private_key: boolean
}

export interface XrayStatusResponse {
  status: string
  active: boolean
}

export interface XrayRestartResponse {
  restarted: boolean
  status: string
  active: boolean
}

export interface XrayConfigResponse {
  config: Record<string, unknown>
  raw: string
  // 与后端约定的脏状态字段（后端尚未部署时可能缺失，均为可选）
  current_hash?: string
  applied_hash?: string
  dirty?: boolean
}

export interface XrayApplyResponse {
  applied: boolean
  method: 'restart' | 'hot_reload'
  config_len: number
  message: string
}

export type DirectDomainStrategy = 'AsIs' | 'UseIP' | 'UseIPv4' | 'UseIPv6'
export type XrayLogLevel = 'debug' | 'info' | 'warning' | 'error' | 'none'

export interface SettingsDTO {
  panel_title: string
  sub_domain: string
  sub_path: string
  sub_nginx_enabled: boolean
  api_domain: string
  api_nginx_enabled: boolean
  server_public_host: string
  xray_log_level: XrayLogLevel
  enable_traffic_stats: boolean
  enable_sniffing: boolean
  default_traffic_limit: number
  default_expire_days: number
  direct_domain_strategy: DirectDomainStrategy
  cn_ip_fallback: string
  monthly_net_baseline_sent: string
  monthly_net_baseline_recv: string
  monthly_net_baseline_month: string
}

export interface UpdateSettingsRequest {
  panel_title?: string | null
  sub_domain?: string | null
  sub_path?: string | null
  sub_nginx_enabled?: boolean | null
  api_domain?: string | null
  api_nginx_enabled?: boolean | null
  server_public_host?: string | null
  xray_log_level?: XrayLogLevel | null
  enable_traffic_stats?: boolean | null
  enable_sniffing?: boolean | null
  default_traffic_limit?: number | null
  default_expire_days?: number | null
  direct_domain_strategy?: DirectDomainStrategy | null
  cn_ip_fallback?: string | null
}

export interface SubscriptionNginxResponse {
  enabled: boolean
  domain: string
  sub_path: string
  config_path: string
  matched_cert_domain: string
  reloaded: boolean
  message: string
}

export interface APINginxResponse {
  enabled: boolean
  domain: string
  direct_url: string
  config_path: string
  matched_cert_domain: string
  reloaded: boolean
  message: string
}

export interface CFTunnelStatus {
  is_linux: boolean
  installed: boolean
  binary_path: string
  version: string
  service_installed: boolean
  active: boolean
  enabled: boolean
  status: string
  panel_listen: string
  sub_domain: string
  sub_path: string
  recommended_service: string
  recommended_path_rule: string
}

export interface CFTunnelInstallResponse {
  installed: boolean
  binary_path: string
  version: string
  message: string
}

export interface CFTunnelActionResponse {
  action: string
  success: boolean
  status: CFTunnelStatus
  message: string
}

export interface PanelRestartResponse {
  restarting: boolean
  message: string
}

export interface LogsResponse {
  service: 'xray' | 'panel' | 'wireguard' | 'nginx'
  lines: number
  content: string
}
