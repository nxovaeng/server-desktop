import type {
  APITokenDTO,
  CreateTokenRequest,
  CreateTokenResponse,
  CreateUserRequest,
  DashboardResponse,
  DeleteResponse,
  DomainDTO,
  GeoDataInfo,
  ImportCertRequest,
  ImportPresetResponse,
  InboundDTO,
  InboundListResponse,
  InboundWGClientConfigResponse,
  LogsResponse,
  OutboundDTO,
  OutboundListResponse,
  OutboundTestResult,
  ParsedWireGuardConfig,
  RoutingRuleDTO,
  RoutingRuleListResponse,
  ScanCertificatesResponse,
  ScanImportCertificatesResponse,
  SettingsDTO,
  TokenDeleteResponse,
  TokenListResponse,
  UpdateSettingsRequest,
  UpdateUserRequest,
  UpdateWGServerRequest,
  UpsertDomainRequest,
  UpsertInboundRequest,
  UpsertOutboundRequest,
  UpsertRoutingRuleRequest,
  UpsertWGPeerRequest,
  UserDTO,
  UserListResponse,
  WGKeyPair,
  WGPeerDTO,
  WGPeerExportConfigResponse,
  WGPeerListResponse,
  WGServerConfigDTO,
  WGServiceActionResponse,
  WGServiceStatus,
  WGSetupEnvResponse,
  XrayApplyResponse,
  XrayConfigResponse,
  XrayRestartResponse,
  XrayStatusResponse
} from './types'

export type ConnectionMode = 'ssh_tunnel' | 'direct'
export type SSHAuthType = 'key_file' | 'key_pem' | 'password' | 'agent'

export interface ServerProfile {
  id: string
  name: string
  mode: ConnectionMode
  direct_url?: string
  allow_insecure_http?: boolean
  ssh_host?: string
  ssh_port?: number
  ssh_user?: string
  ssh_auth_type?: SSHAuthType
  ssh_key_path?: string
  ssh_fingerprint?: string
  remote_bind_host?: string
  remote_panel_port?: number
  has_api_token: boolean
  has_ssh_password: boolean
  has_ssh_key_pem: boolean
  has_ssh_key_passphrase: boolean
  created_at: string
  updated_at: string
  last_connected_at?: string | null
}

export interface ProfileSecrets {
  api_token?: string
  ssh_password?: string
  ssh_key_pem?: string
  ssh_key_passphrase?: string
}

export interface ProfileListResult {
  profiles: ServerProfile[]
  active_profile_id: string
}

export interface ConnectionStatus {
  connected: boolean
  profile_id: string
  profile_name: string
  mode: string
  latency_ms: number
  ssh_fingerprint?: string
  server_version?: string
  token_name?: string
  server_public_host?: string
  sub_domain?: string
  sub_path?: string
  panel_title?: string
  warning?: string
}

export interface URLCheckResult {
  normalized_url: string
  is_plaintext_http: boolean
  is_loopback: boolean
  requires_https_warn: boolean
}

interface WailsAppBinding {
  ListProfiles(): Promise<ProfileListResult>
  SaveProfile(profile: Partial<ServerProfile>, secrets: ProfileSecrets): Promise<ServerProfile>
  DeleteProfile(id: string): Promise<void>
  ConnectProfile(id: string, trustTOFU: boolean): Promise<ConnectionStatus>
  Disconnect(): Promise<void>
  GetConnectionStatus(): Promise<ConnectionStatus>
  TestLatency(): Promise<number>
  GenerateLocalWGKeys(): Promise<WGKeyPair>
  DeriveLocalWGPublicKey(privKey: string): Promise<string>
  ValidateDirectURL(raw: string): Promise<URLCheckResult>
  SaveTextFile(defaultFilename: string, content: string): Promise<string>

  VerifyAuth(): Promise<Record<string, unknown>>
  ListTokens(): Promise<TokenListResponse>
  CreateToken(req: CreateTokenRequest): Promise<CreateTokenResponse>
  DeleteToken(id: string, purge: boolean): Promise<TokenDeleteResponse>

  GetDashboard(): Promise<DashboardResponse>

  ListUsers(query: string): Promise<UserListResponse>
  CreateUser(req: CreateUserRequest): Promise<UserDTO>
  UpdateUser(id: string, req: UpdateUserRequest): Promise<UserDTO>
  DeleteUser(id: string): Promise<DeleteResponse>
  ToggleUser(id: string): Promise<UserDTO>
  ResetUserTraffic(id: string): Promise<UserDTO>

  ListInbounds(): Promise<InboundListResponse>
  CreateInbound(req: UpsertInboundRequest): Promise<InboundDTO>
  UpdateInbound(id: string, req: UpsertInboundRequest): Promise<InboundDTO>
  DeleteInbound(id: string): Promise<DeleteResponse>
  ToggleInbound(id: string): Promise<InboundDTO>
  GetInboundWGClient(id: string): Promise<InboundWGClientConfigResponse>

  ListOutbounds(): Promise<OutboundListResponse>
  CreateOutbound(req: UpsertOutboundRequest): Promise<OutboundDTO>
  UpdateOutbound(id: string, req: UpsertOutboundRequest): Promise<OutboundDTO>
  DeleteOutbound(id: string): Promise<DeleteResponse>
  ToggleOutbound(id: string): Promise<OutboundDTO>
  TestOutbound(id: string): Promise<OutboundTestResult>
  ImportOutbound(link: string): Promise<OutboundDTO>
  ParseWireGuardOutbound(cfg: string): Promise<ParsedWireGuardConfig>
  GenerateOutboundWGKeys(): Promise<WGKeyPair>

  ListRouting(): Promise<RoutingRuleListResponse>
  CreateRouting(req: UpsertRoutingRuleRequest): Promise<RoutingRuleDTO>
  UpdateRouting(id: string, req: UpsertRoutingRuleRequest): Promise<RoutingRuleDTO>
  DeleteRouting(id: string): Promise<DeleteResponse>
  ToggleRouting(id: string): Promise<RoutingRuleDTO>
  ImportRoutingPreset(preset: string): Promise<ImportPresetResponse>
  GetGeoData(): Promise<GeoDataInfo>

  ListDomains(): Promise<{ domains: DomainDTO[]; total: number }>
  CreateDomain(req: UpsertDomainRequest): Promise<DomainDTO>
  UpdateDomain(id: string, req: UpsertDomainRequest): Promise<DomainDTO>
  DeleteDomain(id: string): Promise<DeleteResponse>
  ScanCerts(): Promise<ScanCertificatesResponse>
  ScanImportCerts(): Promise<ScanImportCertificatesResponse>
  ImportSingleCert(req: ImportCertRequest): Promise<DomainDTO>

  GetWGStatus(): Promise<WGServiceStatus>
  SetupWGEnv(): Promise<WGSetupEnvResponse>
  WGServiceAction(action: string): Promise<WGServiceActionResponse>
  GetWGServer(): Promise<WGServerConfigDTO>
  UpdateWGServer(req: UpdateWGServerRequest): Promise<WGServerConfigDTO>
  ListWGPeers(): Promise<WGPeerListResponse>
  CreateWGPeer(req: UpsertWGPeerRequest): Promise<WGPeerDTO>
  UpdateWGPeer(id: number, req: UpsertWGPeerRequest): Promise<WGPeerDTO>
  DeleteWGPeer(id: number): Promise<DeleteResponse>
  ToggleWGPeer(id: number): Promise<WGPeerDTO>
  ExportWGPeerConfig(id: number): Promise<WGPeerExportConfigResponse>
  GenerateWGKeys(): Promise<WGKeyPair>

  GetXrayStatus(): Promise<XrayStatusResponse>
  RestartXray(): Promise<XrayRestartResponse>
  GetXrayConfig(): Promise<XrayConfigResponse>
  ApplyXrayConfig(hot: boolean): Promise<XrayApplyResponse>
  GetSettings(): Promise<SettingsDTO>
  UpdateSettings(req: UpdateSettingsRequest): Promise<SettingsDTO>
  RestartPanelAndReconnect(): Promise<ConnectionStatus>
  GetLogs(serviceName: string, lines: number): Promise<LogsResponse>
}

declare global {
  interface Window {
    go?: {
      main?: {
        App?: WailsAppBinding
      }
    }
  }
}

export function getApp(): WailsAppBinding {
  const app = window.go?.main?.App
  if (!app) {
    throw new Error('Wails runtime bridge (window.go.main.App) is unavailable')
  }
  return app
}

export function formatError(err: unknown): string {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return String(err)
}

export function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 2)} ${units[i]}`
}

export function formatUptime(seconds: number): string {
  if (!seconds || seconds < 60) return `${Math.max(0, Math.floor(seconds || 0))} 秒`
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days} 天 ${hours} 小时`
  if (hours > 0) return `${hours} 小时 ${mins} 分钟`
  return `${mins} 分钟`
}

export async function copyToClipboard(text: string): Promise<void> {
  await navigator.clipboard.writeText(text)
}

export const desktopApi = {
  listProfiles: () => getApp().ListProfiles(),
  saveProfile: (profile: Partial<ServerProfile>, secrets: ProfileSecrets) =>
    getApp().SaveProfile(profile, secrets),
  deleteProfile: (id: string) => getApp().DeleteProfile(id),
  connectProfile: (id: string, trustTOFU: boolean) => getApp().ConnectProfile(id, trustTOFU),
  disconnect: () => getApp().Disconnect(),
  getConnectionStatus: () => getApp().GetConnectionStatus(),
  testLatency: () => getApp().TestLatency(),
  generateLocalWGKeys: () => getApp().GenerateLocalWGKeys(),
  deriveLocalWGPublicKey: (priv: string) => getApp().DeriveLocalWGPublicKey(priv),
  validateDirectURL: (raw: string) => getApp().ValidateDirectURL(raw),
  saveTextFile: (name: string, content: string) => getApp().SaveTextFile(name, content),

  listTokens: () => getApp().ListTokens(),
  createToken: (req: CreateTokenRequest) => getApp().CreateToken(req),
  deleteToken: (id: string, purge: boolean) => getApp().DeleteToken(id, purge),

  getDashboard: () => getApp().GetDashboard(),

  listUsers: (q = '') => getApp().ListUsers(q),
  createUser: (req: CreateUserRequest) => getApp().CreateUser(req),
  updateUser: (id: string, req: UpdateUserRequest) => getApp().UpdateUser(id, req),
  deleteUser: (id: string) => getApp().DeleteUser(id),
  toggleUser: (id: string) => getApp().ToggleUser(id),
  resetUserTraffic: (id: string) => getApp().ResetUserTraffic(id),

  listInbounds: () => getApp().ListInbounds(),
  createInbound: (req: UpsertInboundRequest) => getApp().CreateInbound(req),
  updateInbound: (id: string, req: UpsertInboundRequest) => getApp().UpdateInbound(id, req),
  deleteInbound: (id: string) => getApp().DeleteInbound(id),
  toggleInbound: (id: string) => getApp().ToggleInbound(id),
  getInboundWGClient: (id: string) => getApp().GetInboundWGClient(id),

  listOutbounds: () => getApp().ListOutbounds(),
  createOutbound: (req: UpsertOutboundRequest) => getApp().CreateOutbound(req),
  updateOutbound: (id: string, req: UpsertOutboundRequest) => getApp().UpdateOutbound(id, req),
  deleteOutbound: (id: string) => getApp().DeleteOutbound(id),
  toggleOutbound: (id: string) => getApp().ToggleOutbound(id),
  testOutbound: (id: string) => getApp().TestOutbound(id),
  importOutbound: (link: string) => getApp().ImportOutbound(link),
  parseWireGuardOutbound: (cfg: string) => getApp().ParseWireGuardOutbound(cfg),
  generateOutboundWGKeys: () => getApp().GenerateOutboundWGKeys(),

  listRouting: () => getApp().ListRouting(),
  createRouting: (req: UpsertRoutingRuleRequest) => getApp().CreateRouting(req),
  updateRouting: (id: string, req: UpsertRoutingRuleRequest) => getApp().UpdateRouting(id, req),
  deleteRouting: (id: string) => getApp().DeleteRouting(id),
  toggleRouting: (id: string) => getApp().ToggleRouting(id),
  importRoutingPreset: (preset: string) => getApp().ImportRoutingPreset(preset),
  getGeoData: () => getApp().GetGeoData(),

  listDomains: () => getApp().ListDomains(),
  createDomain: (req: UpsertDomainRequest) => getApp().CreateDomain(req),
  updateDomain: (id: string, req: UpsertDomainRequest) => getApp().UpdateDomain(id, req),
  deleteDomain: (id: string) => getApp().DeleteDomain(id),
  scanCerts: () => getApp().ScanCerts(),
  scanImportCerts: () => getApp().ScanImportCerts(),
  importSingleCert: (req: ImportCertRequest) => getApp().ImportSingleCert(req),

  getWGStatus: () => getApp().GetWGStatus(),
  setupWGEnv: () => getApp().SetupWGEnv(),
  wgServiceAction: (action: string) => getApp().WGServiceAction(action),
  getWGServer: () => getApp().GetWGServer(),
  updateWGServer: (req: UpdateWGServerRequest) => getApp().UpdateWGServer(req),
  listWGPeers: () => getApp().ListWGPeers(),
  createWGPeer: (req: UpsertWGPeerRequest) => getApp().CreateWGPeer(req),
  updateWGPeer: (id: number, req: UpsertWGPeerRequest) => getApp().UpdateWGPeer(id, req),
  deleteWGPeer: (id: number) => getApp().DeleteWGPeer(id),
  toggleWGPeer: (id: number) => getApp().ToggleWGPeer(id),
  exportWGPeerConfig: (id: number) => getApp().ExportWGPeerConfig(id),
  generateWGKeys: () => getApp().GenerateWGKeys(),

  getXrayStatus: () => getApp().GetXrayStatus(),
  restartXray: () => getApp().RestartXray(),
  getXrayConfig: () => getApp().GetXrayConfig(),
  applyXrayConfig: (hot = false) => getApp().ApplyXrayConfig(hot),
  getSettings: () => getApp().GetSettings(),
  updateSettings: (req: UpdateSettingsRequest) => getApp().UpdateSettings(req),
  restartPanelAndReconnect: () => getApp().RestartPanelAndReconnect(),
  getLogs: (serviceName: string, lines: number) => getApp().GetLogs(serviceName, lines)
}

export type { APITokenDTO }
