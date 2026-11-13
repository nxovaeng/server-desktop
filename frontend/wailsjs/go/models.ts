export namespace apiclient {
	
	export class APINginxResponse {
	    enabled: boolean;
	    domain: string;
	    direct_url: string;
	    config_path: string;
	    matched_cert_domain: string;
	    reloaded: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new APINginxResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.domain = source["domain"];
	        this.direct_url = source["direct_url"];
	        this.config_path = source["config_path"];
	        this.matched_cert_domain = source["matched_cert_domain"];
	        this.reloaded = source["reloaded"];
	        this.message = source["message"];
	    }
	}
	export class APITokenDTO {
	    id: string;
	    name: string;
	    token_prefix: string;
	    // Go type: time
	    expires_at?: any;
	    revoked: boolean;
	    // Go type: time
	    last_used_at?: any;
	    last_ip: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new APITokenDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.token_prefix = source["token_prefix"];
	        this.expires_at = this.convertValues(source["expires_at"], null);
	        this.revoked = source["revoked"];
	        this.last_used_at = this.convertValues(source["last_used_at"], null);
	        this.last_ip = source["last_ip"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AuthVerifyResponse {
	    valid: boolean;
	    token_id: string;
	    token_name: string;
	    token_prefix: string;
	    version: string;
	    server_public_host: string;
	    sub_domain: string;
	    sub_path: string;
	    panel_title: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthVerifyResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.token_id = source["token_id"];
	        this.token_name = source["token_name"];
	        this.token_prefix = source["token_prefix"];
	        this.version = source["version"];
	        this.server_public_host = source["server_public_host"];
	        this.sub_domain = source["sub_domain"];
	        this.sub_path = source["sub_path"];
	        this.panel_title = source["panel_title"];
	    }
	}
	export class CFTunnelStatus {
	    is_linux: boolean;
	    installed: boolean;
	    binary_path: string;
	    version: string;
	    service_installed: boolean;
	    active: boolean;
	    enabled: boolean;
	    status: string;
	    panel_listen: string;
	    sub_domain: string;
	    sub_path: string;
	    recommended_service: string;
	    recommended_path_rule: string;
	
	    static createFrom(source: any = {}) {
	        return new CFTunnelStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_linux = source["is_linux"];
	        this.installed = source["installed"];
	        this.binary_path = source["binary_path"];
	        this.version = source["version"];
	        this.service_installed = source["service_installed"];
	        this.active = source["active"];
	        this.enabled = source["enabled"];
	        this.status = source["status"];
	        this.panel_listen = source["panel_listen"];
	        this.sub_domain = source["sub_domain"];
	        this.sub_path = source["sub_path"];
	        this.recommended_service = source["recommended_service"];
	        this.recommended_path_rule = source["recommended_path_rule"];
	    }
	}
	export class CFTunnelActionResponse {
	    action: string;
	    success: boolean;
	    status: CFTunnelStatus;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new CFTunnelActionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.success = source["success"];
	        this.status = this.convertValues(source["status"], CFTunnelStatus);
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CFTunnelInstallResponse {
	    installed: boolean;
	    binary_path: string;
	    version: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new CFTunnelInstallResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.binary_path = source["binary_path"];
	        this.version = source["version"];
	        this.message = source["message"];
	    }
	}
	
	export class CertificateInfo {
	    domain: string;
	    domains: string[];
	    cert_path: string;
	    key_path: string;
	    // Go type: time
	    expiry_date: any;
	    days_to_expiry: number;
	    status: string;
	    issuer: string;
	    is_wildcard: boolean;
	    exists: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CertificateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.domains = source["domains"];
	        this.cert_path = source["cert_path"];
	        this.key_path = source["key_path"];
	        this.expiry_date = this.convertValues(source["expiry_date"], null);
	        this.days_to_expiry = source["days_to_expiry"];
	        this.status = source["status"];
	        this.issuer = source["issuer"];
	        this.is_wildcard = source["is_wildcard"];
	        this.exists = source["exists"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CreateTokenRequest {
	    name: string;
	    expires_days?: number;
	
	    static createFrom(source: any = {}) {
	        return new CreateTokenRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.expires_days = source["expires_days"];
	    }
	}
	export class CreateTokenResponse {
	    token: APITokenDTO;
	    plaintext_token: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateTokenResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = this.convertValues(source["token"], APITokenDTO);
	        this.plaintext_token = source["plaintext_token"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CreateUserRequest {
	    name: string;
	    email?: string;
	    uuid?: string;
	    traffic_limit_bytes: number;
	    // Go type: time
	    expiry_date?: any;
	    expire_days?: number;
	    enabled?: boolean;
	    note?: string;
	    wg_pub_key?: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateUserRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.email = source["email"];
	        this.uuid = source["uuid"];
	        this.traffic_limit_bytes = source["traffic_limit_bytes"];
	        this.expiry_date = this.convertValues(source["expiry_date"], null);
	        this.expire_days = source["expire_days"];
	        this.enabled = source["enabled"];
	        this.note = source["note"];
	        this.wg_pub_key = source["wg_pub_key"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DashboardCounts {
	    total: number;
	    active: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardCounts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.active = source["active"];
	    }
	}
	export class DashboardServices {
	    xray_active: boolean;
	    xray_status: string;
	    wireguard_installed: boolean;
	    wireguard_active: boolean;
	    wireguard_interface: string;
	    wireguard_online_peers: number;
	    wireguard_total_peers: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardServices(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.xray_active = source["xray_active"];
	        this.xray_status = source["xray_status"];
	        this.wireguard_installed = source["wireguard_installed"];
	        this.wireguard_active = source["wireguard_active"];
	        this.wireguard_interface = source["wireguard_interface"];
	        this.wireguard_online_peers = source["wireguard_online_peers"];
	        this.wireguard_total_peers = source["wireguard_total_peers"];
	    }
	}
	export class DashboardSystemMetrics {
	    hostname: string;
	    os: string;
	    platform: string;
	    uptime_seconds: number;
	    go_version: string;
	    cpu_model: string;
	    cpu_cores: number;
	    cpu_usage_percent: number;
	    mem_total_bytes: number;
	    mem_used_bytes: number;
	    mem_free_bytes: number;
	    mem_percent: number;
	    disk_total_bytes: number;
	    disk_used_bytes: number;
	    disk_free_bytes: number;
	    disk_percent: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardSystemMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostname = source["hostname"];
	        this.os = source["os"];
	        this.platform = source["platform"];
	        this.uptime_seconds = source["uptime_seconds"];
	        this.go_version = source["go_version"];
	        this.cpu_model = source["cpu_model"];
	        this.cpu_cores = source["cpu_cores"];
	        this.cpu_usage_percent = source["cpu_usage_percent"];
	        this.mem_total_bytes = source["mem_total_bytes"];
	        this.mem_used_bytes = source["mem_used_bytes"];
	        this.mem_free_bytes = source["mem_free_bytes"];
	        this.mem_percent = source["mem_percent"];
	        this.disk_total_bytes = source["disk_total_bytes"];
	        this.disk_used_bytes = source["disk_used_bytes"];
	        this.disk_free_bytes = source["disk_free_bytes"];
	        this.disk_percent = source["disk_percent"];
	    }
	}
	export class DashboardTraffic {
	    users_traffic_used_bytes: number;
	    monthly_net_sent_bytes: number;
	    monthly_net_recv_bytes: number;
	    monthly_net_total_bytes: number;
	    baseline_month: string;
	
	    static createFrom(source: any = {}) {
	        return new DashboardTraffic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.users_traffic_used_bytes = source["users_traffic_used_bytes"];
	        this.monthly_net_sent_bytes = source["monthly_net_sent_bytes"];
	        this.monthly_net_recv_bytes = source["monthly_net_recv_bytes"];
	        this.monthly_net_total_bytes = source["monthly_net_total_bytes"];
	        this.baseline_month = source["baseline_month"];
	    }
	}
	export class DashboardResponse {
	    users: DashboardCounts;
	    inbounds: DashboardCounts;
	    outbounds: DashboardCounts;
	    domains: DashboardCounts;
	    routing_rules: DashboardCounts;
	    traffic: DashboardTraffic;
	    system: DashboardSystemMetrics;
	    services: DashboardServices;
	
	    static createFrom(source: any = {}) {
	        return new DashboardResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.users = this.convertValues(source["users"], DashboardCounts);
	        this.inbounds = this.convertValues(source["inbounds"], DashboardCounts);
	        this.outbounds = this.convertValues(source["outbounds"], DashboardCounts);
	        this.domains = this.convertValues(source["domains"], DashboardCounts);
	        this.routing_rules = this.convertValues(source["routing_rules"], DashboardCounts);
	        this.traffic = this.convertValues(source["traffic"], DashboardTraffic);
	        this.system = this.convertValues(source["system"], DashboardSystemMetrics);
	        this.services = this.convertValues(source["services"], DashboardServices);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class DeleteResponse {
	    id: string;
	    deleted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeleteResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.deleted = source["deleted"];
	    }
	}
	export class DomainDTO {
	    id: string;
	    domain: string;
	    type: string;
	    is_wildcard: boolean;
	    cert_path: string;
	    key_path: string;
	    enabled: boolean;
	    has_cert: boolean;
	    // Go type: time
	    cert_expiry_date?: any;
	    days_to_expiry?: number;
	    cert_issuer: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new DomainDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.domain = source["domain"];
	        this.type = source["type"];
	        this.is_wildcard = source["is_wildcard"];
	        this.cert_path = source["cert_path"];
	        this.key_path = source["key_path"];
	        this.enabled = source["enabled"];
	        this.has_cert = source["has_cert"];
	        this.cert_expiry_date = this.convertValues(source["cert_expiry_date"], null);
	        this.days_to_expiry = source["days_to_expiry"];
	        this.cert_issuer = source["cert_issuer"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DomainListResponse {
	    domains: DomainDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new DomainListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domains = this.convertValues(source["domains"], DomainDTO);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DomainSummaryDTO {
	    id: string;
	    domain: string;
	    type: string;
	    is_wildcard: boolean;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DomainSummaryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.domain = source["domain"];
	        this.type = source["type"];
	        this.is_wildcard = source["is_wildcard"];
	        this.enabled = source["enabled"];
	    }
	}
	export class GeoDataInfo {
	    geosite_available: boolean;
	    geoip_available: boolean;
	    geosite_tags: string[];
	    geoip_codes: string[];
	
	    static createFrom(source: any = {}) {
	        return new GeoDataInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.geosite_available = source["geosite_available"];
	        this.geoip_available = source["geoip_available"];
	        this.geosite_tags = source["geosite_tags"];
	        this.geoip_codes = source["geoip_codes"];
	    }
	}
	export class ImportCertRequest {
	    domain: string;
	    cert_path: string;
	    key_path: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportCertRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.cert_path = source["cert_path"];
	        this.key_path = source["key_path"];
	    }
	}
	export class RoutingRuleDTO {
	    id: string;
	    name: string;
	    type: string;
	    inbound_tag: string;
	    domains: string;
	    ips: string;
	    geosite_tags: string;
	    geoip_codes: string;
	    protocols: string;
	    outbound_tag: string;
	    priority: number;
	    enabled: boolean;
	    remark: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new RoutingRuleDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.inbound_tag = source["inbound_tag"];
	        this.domains = source["domains"];
	        this.ips = source["ips"];
	        this.geosite_tags = source["geosite_tags"];
	        this.geoip_codes = source["geoip_codes"];
	        this.protocols = source["protocols"];
	        this.outbound_tag = source["outbound_tag"];
	        this.priority = source["priority"];
	        this.enabled = source["enabled"];
	        this.remark = source["remark"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ImportPresetResponse {
	    preset: string;
	    imported_count: number;
	    skipped_count: number;
	    rules: RoutingRuleDTO[];
	
	    static createFrom(source: any = {}) {
	        return new ImportPresetResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preset = source["preset"];
	        this.imported_count = source["imported_count"];
	        this.skipped_count = source["skipped_count"];
	        this.rules = this.convertValues(source["rules"], RoutingRuleDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InboundDTO {
	    id: string;
	    tag: string;
	    protocol: string;
	    transport: string;
	    port: number;
	    listen: string;
	    domain_id: string;
	    domain?: DomainSummaryDTO;
	    path: string;
	    service_name: string;
	    host: string;
	    mode: string;
	    actual_domain: string;
	    custom_sni: string;
	    connect_domain: string;
	    has_wg_secret_key: boolean;
	    wg_public_key: string;
	    wg_peer_pub_key: string;
	    wg_mtu: number;
	    wg_local_ip: string;
	    exclude_from_sub: boolean;
	    enabled: boolean;
	    use_uds: boolean;
	    socket_path: string;
	    remark: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new InboundDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tag = source["tag"];
	        this.protocol = source["protocol"];
	        this.transport = source["transport"];
	        this.port = source["port"];
	        this.listen = source["listen"];
	        this.domain_id = source["domain_id"];
	        this.domain = this.convertValues(source["domain"], DomainSummaryDTO);
	        this.path = source["path"];
	        this.service_name = source["service_name"];
	        this.host = source["host"];
	        this.mode = source["mode"];
	        this.actual_domain = source["actual_domain"];
	        this.custom_sni = source["custom_sni"];
	        this.connect_domain = source["connect_domain"];
	        this.has_wg_secret_key = source["has_wg_secret_key"];
	        this.wg_public_key = source["wg_public_key"];
	        this.wg_peer_pub_key = source["wg_peer_pub_key"];
	        this.wg_mtu = source["wg_mtu"];
	        this.wg_local_ip = source["wg_local_ip"];
	        this.exclude_from_sub = source["exclude_from_sub"];
	        this.enabled = source["enabled"];
	        this.use_uds = source["use_uds"];
	        this.socket_path = source["socket_path"];
	        this.remark = source["remark"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InboundListResponse {
	    inbounds: InboundDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new InboundListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inbounds = this.convertValues(source["inbounds"], InboundDTO);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InboundWGClientConfigResponse {
	    inbound_id: string;
	    inbound_tag: string;
	    server_public_key: string;
	    server_host: string;
	    port: number;
	    endpoint: string;
	    peer_ip: string;
	    mtu: number;
	    wg_conf: string;
	
	    static createFrom(source: any = {}) {
	        return new InboundWGClientConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inbound_id = source["inbound_id"];
	        this.inbound_tag = source["inbound_tag"];
	        this.server_public_key = source["server_public_key"];
	        this.server_host = source["server_host"];
	        this.port = source["port"];
	        this.endpoint = source["endpoint"];
	        this.peer_ip = source["peer_ip"];
	        this.mtu = source["mtu"];
	        this.wg_conf = source["wg_conf"];
	    }
	}
	export class IssueCertRequest {
	    domain: string;
	    include_wildcard: boolean;
	    type?: string;
	    provider: string;
	    cf_token?: string;
	    cf_key?: string;
	    cf_email?: string;
	    cf_zone_id?: string;
	    lua_key?: string;
	    lua_email?: string;
	    ca_server?: string;
	    email?: string;
	    force: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IssueCertRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.include_wildcard = source["include_wildcard"];
	        this.type = source["type"];
	        this.provider = source["provider"];
	        this.cf_token = source["cf_token"];
	        this.cf_key = source["cf_key"];
	        this.cf_email = source["cf_email"];
	        this.cf_zone_id = source["cf_zone_id"];
	        this.lua_key = source["lua_key"];
	        this.lua_email = source["lua_email"];
	        this.ca_server = source["ca_server"];
	        this.email = source["email"];
	        this.force = source["force"];
	    }
	}
	export class IssueCertResponse {
	    domain: DomainDTO;
	    provider: string;
	    ca_server: string;
	    reused_existing: boolean;
	    output: string;
	
	    static createFrom(source: any = {}) {
	        return new IssueCertResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = this.convertValues(source["domain"], DomainDTO);
	        this.provider = source["provider"];
	        this.ca_server = source["ca_server"];
	        this.reused_existing = source["reused_existing"];
	        this.output = source["output"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LogsResponse {
	    service: string;
	    lines: number;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new LogsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.service = source["service"];
	        this.lines = source["lines"];
	        this.content = source["content"];
	    }
	}
	export class OutboundDTO {
	    id: string;
	    tag: string;
	    type: string;
	    server: string;
	    port: number;
	    username: string;
	    has_password: boolean;
	    has_wg_secret_key: boolean;
	    wg_client_pub_key: string;
	    wg_public_key: string;
	    wg_reserved: string;
	    wg_local_ipv4: string;
	    wg_local_ipv6: string;
	    wg_mtu: number;
	    wg_dns: string;
	    has_trojan_password: boolean;
	    trojan_sni: string;
	    trojan_network: string;
	    uuid: string;
	    flow: string;
	    security: string;
	    network: string;
	    header_type: string;
	    request_host: string;
	    path: string;
	    service_name: string;
	    tls: boolean;
	    tls_server_name: string;
	    tls_alpn: string;
	    reality: boolean;
	    reality_pubkey: string;
	    reality_short_id: string;
	    reality_sni: string;
	    enabled: boolean;
	    priority: number;
	    remark: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new OutboundDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tag = source["tag"];
	        this.type = source["type"];
	        this.server = source["server"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.has_password = source["has_password"];
	        this.has_wg_secret_key = source["has_wg_secret_key"];
	        this.wg_client_pub_key = source["wg_client_pub_key"];
	        this.wg_public_key = source["wg_public_key"];
	        this.wg_reserved = source["wg_reserved"];
	        this.wg_local_ipv4 = source["wg_local_ipv4"];
	        this.wg_local_ipv6 = source["wg_local_ipv6"];
	        this.wg_mtu = source["wg_mtu"];
	        this.wg_dns = source["wg_dns"];
	        this.has_trojan_password = source["has_trojan_password"];
	        this.trojan_sni = source["trojan_sni"];
	        this.trojan_network = source["trojan_network"];
	        this.uuid = source["uuid"];
	        this.flow = source["flow"];
	        this.security = source["security"];
	        this.network = source["network"];
	        this.header_type = source["header_type"];
	        this.request_host = source["request_host"];
	        this.path = source["path"];
	        this.service_name = source["service_name"];
	        this.tls = source["tls"];
	        this.tls_server_name = source["tls_server_name"];
	        this.tls_alpn = source["tls_alpn"];
	        this.reality = source["reality"];
	        this.reality_pubkey = source["reality_pubkey"];
	        this.reality_short_id = source["reality_short_id"];
	        this.reality_sni = source["reality_sni"];
	        this.enabled = source["enabled"];
	        this.priority = source["priority"];
	        this.remark = source["remark"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OutboundListResponse {
	    outbounds: OutboundDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new OutboundListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.outbounds = this.convertValues(source["outbounds"], OutboundDTO);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OutboundTestResult {
	    success: boolean;
	    message: string;
	    latency_ms?: number;
	    endpoint?: string;
	    outbound_ip?: string;
	    location?: string;
	
	    static createFrom(source: any = {}) {
	        return new OutboundTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.latency_ms = source["latency_ms"];
	        this.endpoint = source["endpoint"];
	        this.outbound_ip = source["outbound_ip"];
	        this.location = source["location"];
	    }
	}
	export class PVPNActionResponse {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new PVPNActionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class PVPNConfigResponse {
	    socks_port: number;
	    socks_host: string;
	    data_dir: string;
	    session_path: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new PVPNConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.socks_port = source["socks_port"];
	        this.socks_host = source["socks_host"];
	        this.data_dir = source["data_dir"];
	        this.session_path = source["session_path"];
	        this.message = source["message"];
	    }
	}
	export class PVPNConnectResponse {
	    success: boolean;
	    server: string;
	    country?: string;
	    outbound_tag?: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new PVPNConnectResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.server = source["server"];
	        this.country = source["country"];
	        this.outbound_tag = source["outbound_tag"];
	        this.message = source["message"];
	    }
	}
	export class PVPNCountryDTO {
	    code: string;
	    name: string;
	    count: number;
	    avg_load: number;
	
	    static createFrom(source: any = {}) {
	        return new PVPNCountryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.count = source["count"];
	        this.avg_load = source["avg_load"];
	    }
	}
	export class PVPNCountriesResponse {
	    countries: PVPNCountryDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PVPNCountriesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.countries = this.convertValues(source["countries"], PVPNCountryDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class PVPNLoginResponse {
	    success: boolean;
	    message: string;
	    cert_expires_at: number;
	    auth_mode: string;
	    session_path?: string;
	
	    static createFrom(source: any = {}) {
	        return new PVPNLoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.cert_expires_at = source["cert_expires_at"];
	        this.auth_mode = source["auth_mode"];
	        this.session_path = source["session_path"];
	    }
	}
	export class PVPNServerDTO {
	    name: string;
	    entry_ip: string;
	    exit_ip: string;
	    country: string;
	    tier: number;
	    load: number;
	    score: number;
	    status: number;
	
	    static createFrom(source: any = {}) {
	        return new PVPNServerDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.entry_ip = source["entry_ip"];
	        this.exit_ip = source["exit_ip"];
	        this.country = source["country"];
	        this.tier = source["tier"];
	        this.load = source["load"];
	        this.score = source["score"];
	        this.status = source["status"];
	    }
	}
	export class PVPNServersResponse {
	    servers: PVPNServerDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PVPNServersResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.servers = this.convertValues(source["servers"], PVPNServerDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PVPNStatusResponse {
	    installed: boolean;
	    service_status: string;
	    running: boolean;
	    server: string;
	    country: string;
	    ip: string;
	    port: number;
	    logged_in: boolean;
	    auth_mode: string;
	    cert_expires_at: number;
	    total_servers: number;
	    tier: number;
	    outbound_exists: boolean;
	    outbound_tag: string;
	    api_endpoint: string;
	    session_path?: string;
	
	    static createFrom(source: any = {}) {
	        return new PVPNStatusResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.service_status = source["service_status"];
	        this.running = source["running"];
	        this.server = source["server"];
	        this.country = source["country"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.logged_in = source["logged_in"];
	        this.auth_mode = source["auth_mode"];
	        this.cert_expires_at = source["cert_expires_at"];
	        this.total_servers = source["total_servers"];
	        this.tier = source["tier"];
	        this.outbound_exists = source["outbound_exists"];
	        this.outbound_tag = source["outbound_tag"];
	        this.api_endpoint = source["api_endpoint"];
	        this.session_path = source["session_path"];
	    }
	}
	export class ParsedWireGuardConfig {
	    private_key: string;
	    client_public_key: string;
	    public_key: string;
	    endpoint: string;
	    server: string;
	    port: number;
	    address: string;
	    local_ipv4: string;
	    local_ipv6: string;
	    dns: string;
	    mtu: number;
	
	    static createFrom(source: any = {}) {
	        return new ParsedWireGuardConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.private_key = source["private_key"];
	        this.client_public_key = source["client_public_key"];
	        this.public_key = source["public_key"];
	        this.endpoint = source["endpoint"];
	        this.server = source["server"];
	        this.port = source["port"];
	        this.address = source["address"];
	        this.local_ipv4 = source["local_ipv4"];
	        this.local_ipv6 = source["local_ipv6"];
	        this.dns = source["dns"];
	        this.mtu = source["mtu"];
	    }
	}
	
	export class RoutingRuleListResponse {
	    rules: RoutingRuleDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new RoutingRuleListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rules = this.convertValues(source["rules"], RoutingRuleDTO);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ScanCertificatesResponse {
	    cert_dir: string;
	    certificates: CertificateInfo[];
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new ScanCertificatesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cert_dir = source["cert_dir"];
	        this.certificates = this.convertValues(source["certificates"], CertificateInfo);
	        this.count = source["count"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ScanImportCertificatesResponse {
	    cert_dir: string;
	    imported_count: number;
	    updated_count: number;
	    skipped_count: number;
	    errors: string[];
	    domains: DomainDTO[];
	
	    static createFrom(source: any = {}) {
	        return new ScanImportCertificatesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cert_dir = source["cert_dir"];
	        this.imported_count = source["imported_count"];
	        this.updated_count = source["updated_count"];
	        this.skipped_count = source["skipped_count"];
	        this.errors = source["errors"];
	        this.domains = this.convertValues(source["domains"], DomainDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SettingsDTO {
	    panel_title: string;
	    sub_domain: string;
	    sub_path: string;
	    sub_nginx_enabled: boolean;
	    api_domain: string;
	    api_nginx_enabled: boolean;
	    server_public_host: string;
	    xray_log_level: string;
	    enable_traffic_stats: boolean;
	    enable_sniffing: boolean;
	    default_traffic_limit: number;
	    default_expire_days: number;
	    direct_domain_strategy: string;
	    cn_ip_fallback: string;
	    monthly_net_baseline_sent: string;
	    monthly_net_baseline_recv: string;
	    monthly_net_baseline_month: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.panel_title = source["panel_title"];
	        this.sub_domain = source["sub_domain"];
	        this.sub_path = source["sub_path"];
	        this.sub_nginx_enabled = source["sub_nginx_enabled"];
	        this.api_domain = source["api_domain"];
	        this.api_nginx_enabled = source["api_nginx_enabled"];
	        this.server_public_host = source["server_public_host"];
	        this.xray_log_level = source["xray_log_level"];
	        this.enable_traffic_stats = source["enable_traffic_stats"];
	        this.enable_sniffing = source["enable_sniffing"];
	        this.default_traffic_limit = source["default_traffic_limit"];
	        this.default_expire_days = source["default_expire_days"];
	        this.direct_domain_strategy = source["direct_domain_strategy"];
	        this.cn_ip_fallback = source["cn_ip_fallback"];
	        this.monthly_net_baseline_sent = source["monthly_net_baseline_sent"];
	        this.monthly_net_baseline_recv = source["monthly_net_baseline_recv"];
	        this.monthly_net_baseline_month = source["monthly_net_baseline_month"];
	    }
	}
	export class SpeedtestStatus {
	    enabled: boolean;
	    domain: string;
	    sub_path: string;
	    file_exists: boolean;
	    file_size: number;
	    url: string;
	    config_path: string;
	
	    static createFrom(source: any = {}) {
	        return new SpeedtestStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.domain = source["domain"];
	        this.sub_path = source["sub_path"];
	        this.file_exists = source["file_exists"];
	        this.file_size = source["file_size"];
	        this.url = source["url"];
	        this.config_path = source["config_path"];
	    }
	}
	export class SubscriptionNginxResponse {
	    enabled: boolean;
	    domain: string;
	    sub_path: string;
	    config_path: string;
	    matched_cert_domain: string;
	    reloaded: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SubscriptionNginxResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.domain = source["domain"];
	        this.sub_path = source["sub_path"];
	        this.config_path = source["config_path"];
	        this.matched_cert_domain = source["matched_cert_domain"];
	        this.reloaded = source["reloaded"];
	        this.message = source["message"];
	    }
	}
	export class TokenDeleteResponse {
	    id: string;
	    revoked: boolean;
	    deleted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TokenDeleteResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.revoked = source["revoked"];
	        this.deleted = source["deleted"];
	    }
	}
	export class TokenListResponse {
	    tokens: APITokenDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new TokenListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tokens = this.convertValues(source["tokens"], APITokenDTO);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateSettingsRequest {
	    panel_title?: string;
	    sub_domain?: string;
	    sub_path?: string;
	    sub_nginx_enabled?: boolean;
	    api_domain?: string;
	    api_nginx_enabled?: boolean;
	    server_public_host?: string;
	    xray_log_level?: string;
	    enable_traffic_stats?: boolean;
	    enable_sniffing?: boolean;
	    default_traffic_limit?: number;
	    default_expire_days?: number;
	    direct_domain_strategy?: string;
	    cn_ip_fallback?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateSettingsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.panel_title = source["panel_title"];
	        this.sub_domain = source["sub_domain"];
	        this.sub_path = source["sub_path"];
	        this.sub_nginx_enabled = source["sub_nginx_enabled"];
	        this.api_domain = source["api_domain"];
	        this.api_nginx_enabled = source["api_nginx_enabled"];
	        this.server_public_host = source["server_public_host"];
	        this.xray_log_level = source["xray_log_level"];
	        this.enable_traffic_stats = source["enable_traffic_stats"];
	        this.enable_sniffing = source["enable_sniffing"];
	        this.default_traffic_limit = source["default_traffic_limit"];
	        this.default_expire_days = source["default_expire_days"];
	        this.direct_domain_strategy = source["direct_domain_strategy"];
	        this.cn_ip_fallback = source["cn_ip_fallback"];
	    }
	}
	export class UpdateSpeedtestRequest {
	    domain?: string;
	    sub_path?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateSpeedtestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.sub_path = source["sub_path"];
	    }
	}
	export class UpdateUserRequest {
	    name: string;
	    email: string;
	    uuid: string;
	    traffic_limit_bytes: number;
	    // Go type: time
	    expiry_date?: any;
	    enabled?: boolean;
	    note: string;
	    wg_pub_key: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateUserRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.email = source["email"];
	        this.uuid = source["uuid"];
	        this.traffic_limit_bytes = source["traffic_limit_bytes"];
	        this.expiry_date = this.convertValues(source["expiry_date"], null);
	        this.enabled = source["enabled"];
	        this.note = source["note"];
	        this.wg_pub_key = source["wg_pub_key"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateWGServerRequest {
	    interface_name?: string;
	    listen_port?: number;
	    address?: string;
	    mtu?: number;
	    private_key?: string;
	    enable_nat?: boolean;
	    post_up?: string;
	    post_down?: string;
	    enabled?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateWGServerRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.interface_name = source["interface_name"];
	        this.listen_port = source["listen_port"];
	        this.address = source["address"];
	        this.mtu = source["mtu"];
	        this.private_key = source["private_key"];
	        this.enable_nat = source["enable_nat"];
	        this.post_up = source["post_up"];
	        this.post_down = source["post_down"];
	        this.enabled = source["enabled"];
	    }
	}
	export class UploadCertRequest {
	    domain: string;
	    type?: string;
	    cert_pem: string;
	    key_pem: string;
	    enabled?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UploadCertRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.type = source["type"];
	        this.cert_pem = source["cert_pem"];
	        this.key_pem = source["key_pem"];
	        this.enabled = source["enabled"];
	    }
	}
	export class UpsertDomainRequest {
	    domain: string;
	    type?: string;
	    is_wildcard: boolean;
	    cert_path?: string;
	    key_path?: string;
	    enabled?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpsertDomainRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.type = source["type"];
	        this.is_wildcard = source["is_wildcard"];
	        this.cert_path = source["cert_path"];
	        this.key_path = source["key_path"];
	        this.enabled = source["enabled"];
	    }
	}
	export class UpsertInboundRequest {
	    tag?: string;
	    protocol?: string;
	    transport?: string;
	    port?: number;
	    listen?: string;
	    domain_id?: string;
	    path?: string;
	    service_name?: string;
	    host?: string;
	    mode?: string;
	    custom_sni?: string;
	    connect_domain?: string;
	    wg_secret_key?: string;
	    wg_public_key?: string;
	    wg_peer_pub_key?: string;
	    wg_mtu?: number;
	    wg_local_ip?: string;
	    exclude_from_sub: boolean;
	    use_uds?: boolean;
	    enabled?: boolean;
	    random_port?: boolean;
	    random_path?: boolean;
	    remark?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpsertInboundRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag = source["tag"];
	        this.protocol = source["protocol"];
	        this.transport = source["transport"];
	        this.port = source["port"];
	        this.listen = source["listen"];
	        this.domain_id = source["domain_id"];
	        this.path = source["path"];
	        this.service_name = source["service_name"];
	        this.host = source["host"];
	        this.mode = source["mode"];
	        this.custom_sni = source["custom_sni"];
	        this.connect_domain = source["connect_domain"];
	        this.wg_secret_key = source["wg_secret_key"];
	        this.wg_public_key = source["wg_public_key"];
	        this.wg_peer_pub_key = source["wg_peer_pub_key"];
	        this.wg_mtu = source["wg_mtu"];
	        this.wg_local_ip = source["wg_local_ip"];
	        this.exclude_from_sub = source["exclude_from_sub"];
	        this.use_uds = source["use_uds"];
	        this.enabled = source["enabled"];
	        this.random_port = source["random_port"];
	        this.random_path = source["random_path"];
	        this.remark = source["remark"];
	    }
	}
	export class UpsertOutboundRequest {
	    tag?: string;
	    type: string;
	    server?: string;
	    port?: number;
	    username?: string;
	    password?: string;
	    wg_secret_key?: string;
	    wg_client_pub_key?: string;
	    wg_public_key?: string;
	    wg_reserved?: string;
	    wg_local_ipv4?: string;
	    wg_local_ipv6?: string;
	    wg_mtu?: number;
	    wg_dns?: string;
	    trojan_password?: string;
	    trojan_sni?: string;
	    trojan_network?: string;
	    uuid?: string;
	    flow?: string;
	    security?: string;
	    network?: string;
	    header_type?: string;
	    request_host?: string;
	    path?: string;
	    service_name?: string;
	    tls: boolean;
	    tls_server_name?: string;
	    tls_alpn?: string;
	    reality: boolean;
	    reality_pubkey?: string;
	    reality_short_id?: string;
	    reality_sni?: string;
	    enabled?: boolean;
	    priority: number;
	    remark?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpsertOutboundRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag = source["tag"];
	        this.type = source["type"];
	        this.server = source["server"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.wg_secret_key = source["wg_secret_key"];
	        this.wg_client_pub_key = source["wg_client_pub_key"];
	        this.wg_public_key = source["wg_public_key"];
	        this.wg_reserved = source["wg_reserved"];
	        this.wg_local_ipv4 = source["wg_local_ipv4"];
	        this.wg_local_ipv6 = source["wg_local_ipv6"];
	        this.wg_mtu = source["wg_mtu"];
	        this.wg_dns = source["wg_dns"];
	        this.trojan_password = source["trojan_password"];
	        this.trojan_sni = source["trojan_sni"];
	        this.trojan_network = source["trojan_network"];
	        this.uuid = source["uuid"];
	        this.flow = source["flow"];
	        this.security = source["security"];
	        this.network = source["network"];
	        this.header_type = source["header_type"];
	        this.request_host = source["request_host"];
	        this.path = source["path"];
	        this.service_name = source["service_name"];
	        this.tls = source["tls"];
	        this.tls_server_name = source["tls_server_name"];
	        this.tls_alpn = source["tls_alpn"];
	        this.reality = source["reality"];
	        this.reality_pubkey = source["reality_pubkey"];
	        this.reality_short_id = source["reality_short_id"];
	        this.reality_sni = source["reality_sni"];
	        this.enabled = source["enabled"];
	        this.priority = source["priority"];
	        this.remark = source["remark"];
	    }
	}
	export class UpsertRoutingRuleRequest {
	    name: string;
	    type: string;
	    inbound_tag?: string;
	    domains?: string;
	    ips?: string;
	    geosite_tags?: string;
	    geoip_codes?: string;
	    protocols?: string;
	    outbound_tag: string;
	    priority: number;
	    enabled?: boolean;
	    remark?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpsertRoutingRuleRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.inbound_tag = source["inbound_tag"];
	        this.domains = source["domains"];
	        this.ips = source["ips"];
	        this.geosite_tags = source["geosite_tags"];
	        this.geoip_codes = source["geoip_codes"];
	        this.protocols = source["protocols"];
	        this.outbound_tag = source["outbound_tag"];
	        this.priority = source["priority"];
	        this.enabled = source["enabled"];
	        this.remark = source["remark"];
	    }
	}
	export class UpsertWGPeerRequest {
	    name: string;
	    public_key?: string;
	    private_key?: string;
	    preshared_key?: string;
	    allowed_ips?: string;
	    endpoint?: string;
	    persistent_keepalive?: number;
	    enabled?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpsertWGPeerRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.public_key = source["public_key"];
	        this.private_key = source["private_key"];
	        this.preshared_key = source["preshared_key"];
	        this.allowed_ips = source["allowed_ips"];
	        this.endpoint = source["endpoint"];
	        this.persistent_keepalive = source["persistent_keepalive"];
	        this.enabled = source["enabled"];
	    }
	}
	export class UserSubURLs {
	    base64: string;
	    plain: string;
	    json: string;
	    yaml: string;
	    relative_path: string;
	
	    static createFrom(source: any = {}) {
	        return new UserSubURLs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.base64 = source["base64"];
	        this.plain = source["plain"];
	        this.json = source["json"];
	        this.yaml = source["yaml"];
	        this.relative_path = source["relative_path"];
	    }
	}
	export class UserDTO {
	    id: string;
	    uuid: string;
	    name: string;
	    email: string;
	    traffic_limit_bytes: number;
	    traffic_used_bytes: number;
	    remaining_traffic_bytes: number;
	    // Go type: time
	    traffic_reset_at?: any;
	    // Go type: time
	    expiry_date?: any;
	    remaining_days: number;
	    enabled: boolean;
	    is_active: boolean;
	    sub_path: string;
	    sub_urls: UserSubURLs;
	    note: string;
	    wg_pub_key: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new UserDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.uuid = source["uuid"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.traffic_limit_bytes = source["traffic_limit_bytes"];
	        this.traffic_used_bytes = source["traffic_used_bytes"];
	        this.remaining_traffic_bytes = source["remaining_traffic_bytes"];
	        this.traffic_reset_at = this.convertValues(source["traffic_reset_at"], null);
	        this.expiry_date = this.convertValues(source["expiry_date"], null);
	        this.remaining_days = source["remaining_days"];
	        this.enabled = source["enabled"];
	        this.is_active = source["is_active"];
	        this.sub_path = source["sub_path"];
	        this.sub_urls = this.convertValues(source["sub_urls"], UserSubURLs);
	        this.note = source["note"];
	        this.wg_pub_key = source["wg_pub_key"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UserListResponse {
	    users: UserDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new UserListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.users = this.convertValues(source["users"], UserDTO);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class WGKeyPair {
	    private_key: string;
	    public_key: string;
	
	    static createFrom(source: any = {}) {
	        return new WGKeyPair(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.private_key = source["private_key"];
	        this.public_key = source["public_key"];
	    }
	}
	export class WGPeerRuntimeStats {
	    public_key: string;
	    endpoint: string;
	    allowed_ips: string;
	    // Go type: time
	    latest_handshake?: any;
	    handshake_ago: string;
	    transfer_rx: number;
	    transfer_tx: number;
	    transfer_rx_human: string;
	    transfer_tx_human: string;
	    is_online: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WGPeerRuntimeStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.public_key = source["public_key"];
	        this.endpoint = source["endpoint"];
	        this.allowed_ips = source["allowed_ips"];
	        this.latest_handshake = this.convertValues(source["latest_handshake"], null);
	        this.handshake_ago = source["handshake_ago"];
	        this.transfer_rx = source["transfer_rx"];
	        this.transfer_tx = source["transfer_tx"];
	        this.transfer_rx_human = source["transfer_rx_human"];
	        this.transfer_tx_human = source["transfer_tx_human"];
	        this.is_online = source["is_online"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WGPeerDTO {
	    id: number;
	    name: string;
	    public_key: string;
	    has_private_key: boolean;
	    has_preshared_key: boolean;
	    allowed_ips: string;
	    endpoint: string;
	    persistent_keepalive: number;
	    enabled: boolean;
	    runtime: WGPeerRuntimeStats;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new WGPeerDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.public_key = source["public_key"];
	        this.has_private_key = source["has_private_key"];
	        this.has_preshared_key = source["has_preshared_key"];
	        this.allowed_ips = source["allowed_ips"];
	        this.endpoint = source["endpoint"];
	        this.persistent_keepalive = source["persistent_keepalive"];
	        this.enabled = source["enabled"];
	        this.runtime = this.convertValues(source["runtime"], WGPeerRuntimeStats);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WGPeerExportConfigResponse {
	    peer_id: number;
	    peer_name: string;
	    server_host: string;
	    endpoint: string;
	    wg_conf: string;
	    xray_json: string;
	    has_private_key: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WGPeerExportConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.peer_id = source["peer_id"];
	        this.peer_name = source["peer_name"];
	        this.server_host = source["server_host"];
	        this.endpoint = source["endpoint"];
	        this.wg_conf = source["wg_conf"];
	        this.xray_json = source["xray_json"];
	        this.has_private_key = source["has_private_key"];
	    }
	}
	export class WGPeerListResponse {
	    peers: WGPeerDTO[];
	    next_available_ip: string;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new WGPeerListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.peers = this.convertValues(source["peers"], WGPeerDTO);
	        this.next_available_ip = source["next_available_ip"];
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class WGServerConfigDTO {
	    id: number;
	    interface_name: string;
	    has_private_key: boolean;
	    public_key: string;
	    listen_port: number;
	    address: string;
	    mtu: number;
	    enable_nat: boolean;
	    post_up: string;
	    post_down: string;
	    enabled: boolean;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new WGServerConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.interface_name = source["interface_name"];
	        this.has_private_key = source["has_private_key"];
	        this.public_key = source["public_key"];
	        this.listen_port = source["listen_port"];
	        this.address = source["address"];
	        this.mtu = source["mtu"];
	        this.enable_nat = source["enable_nat"];
	        this.post_up = source["post_up"];
	        this.post_down = source["post_down"];
	        this.enabled = source["enabled"];
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WGServiceStatus {
	    is_linux: boolean;
	    is_installed: boolean;
	    is_active: boolean;
	    is_enabled: boolean;
	    interface_name: string;
	    listen_port: number;
	    public_key: string;
	    address: string;
	    total_peers: number;
	    active_peers: number;
	    online_peers: number;
	    port_conflict: boolean;
	    conflict_inbound: string;
	    ipv4_forwarding: boolean;
	    ipv6_forwarding: boolean;
	    enable_nat: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WGServiceStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_linux = source["is_linux"];
	        this.is_installed = source["is_installed"];
	        this.is_active = source["is_active"];
	        this.is_enabled = source["is_enabled"];
	        this.interface_name = source["interface_name"];
	        this.listen_port = source["listen_port"];
	        this.public_key = source["public_key"];
	        this.address = source["address"];
	        this.total_peers = source["total_peers"];
	        this.active_peers = source["active_peers"];
	        this.online_peers = source["online_peers"];
	        this.port_conflict = source["port_conflict"];
	        this.conflict_inbound = source["conflict_inbound"];
	        this.ipv4_forwarding = source["ipv4_forwarding"];
	        this.ipv6_forwarding = source["ipv6_forwarding"];
	        this.enable_nat = source["enable_nat"];
	    }
	}
	export class WGServiceActionResponse {
	    action: string;
	    success: boolean;
	    status: WGServiceStatus;
	
	    static createFrom(source: any = {}) {
	        return new WGServiceActionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.success = source["success"];
	        this.status = this.convertValues(source["status"], WGServiceStatus);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class WGSetupEnvResponse {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new WGSetupEnvResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class XrayApplyResponse {
	    applied: boolean;
	    method: string;
	    config_len: number;
	    changed: boolean;
	    warnings?: string[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new XrayApplyResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.applied = source["applied"];
	        this.method = source["method"];
	        this.config_len = source["config_len"];
	        this.changed = source["changed"];
	        this.warnings = source["warnings"];
	        this.message = source["message"];
	    }
	}
	export class XrayConfigResponse {
	    config: {[key: string]: any};
	    raw: string;
	    current_hash?: string;
	    applied_hash?: string;
	    dirty?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new XrayConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = source["config"];
	        this.raw = source["raw"];
	        this.current_hash = source["current_hash"];
	        this.applied_hash = source["applied_hash"];
	        this.dirty = source["dirty"];
	    }
	}
	export class XrayRestartResponse {
	    restarted: boolean;
	    status: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new XrayRestartResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.restarted = source["restarted"];
	        this.status = source["status"];
	        this.active = source["active"];
	    }
	}
	export class XrayStatusResponse {
	    status: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new XrayStatusResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.active = source["active"];
	    }
	}

}

export namespace local {
	
	export class URLCheckResult {
	    normalized_url: string;
	    is_plaintext_http: boolean;
	    is_loopback: boolean;
	    requires_https_warn: boolean;
	
	    static createFrom(source: any = {}) {
	        return new URLCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.normalized_url = source["normalized_url"];
	        this.is_plaintext_http = source["is_plaintext_http"];
	        this.is_loopback = source["is_loopback"];
	        this.requires_https_warn = source["requires_https_warn"];
	    }
	}
	export class WGKeyPair {
	    private_key: string;
	    public_key: string;
	
	    static createFrom(source: any = {}) {
	        return new WGKeyPair(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.private_key = source["private_key"];
	        this.public_key = source["public_key"];
	    }
	}

}

export namespace main {
	
	export class ProfileListResult {
	    profiles: store.ServerProfile[];
	    active_profile_id: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profiles = this.convertValues(source["profiles"], store.ServerProfile);
	        this.active_profile_id = source["active_profile_id"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace session {
	
	export class ConnectionStatus {
	    connected: boolean;
	    profile_id: string;
	    profile_name: string;
	    mode: string;
	    latency_ms: number;
	    ssh_fingerprint?: string;
	    server_version?: string;
	    token_name?: string;
	    server_public_host?: string;
	    sub_domain?: string;
	    sub_path?: string;
	    panel_title?: string;
	    warning?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.profile_id = source["profile_id"];
	        this.profile_name = source["profile_name"];
	        this.mode = source["mode"];
	        this.latency_ms = source["latency_ms"];
	        this.ssh_fingerprint = source["ssh_fingerprint"];
	        this.server_version = source["server_version"];
	        this.token_name = source["token_name"];
	        this.server_public_host = source["server_public_host"];
	        this.sub_domain = source["sub_domain"];
	        this.sub_path = source["sub_path"];
	        this.panel_title = source["panel_title"];
	        this.warning = source["warning"];
	    }
	}
	export class BackendUpdateResult {
	    success: boolean;
	    exit_code: number;
	    output: string;
	    error?: string;
	    status: ConnectionStatus;
	
	    static createFrom(source: any = {}) {
	        return new BackendUpdateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.exit_code = source["exit_code"];
	        this.output = source["output"];
	        this.error = source["error"];
	        this.status = this.convertValues(source["status"], ConnectionStatus);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace store {
	
	export class ProfileSecrets {
	    api_token?: string;
	    ssh_password?: string;
	    ssh_key_pem?: string;
	    ssh_key_passphrase?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileSecrets(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_token = source["api_token"];
	        this.ssh_password = source["ssh_password"];
	        this.ssh_key_pem = source["ssh_key_pem"];
	        this.ssh_key_passphrase = source["ssh_key_passphrase"];
	    }
	}
	export class ServerProfile {
	    id: string;
	    name: string;
	    mode: string;
	    direct_url?: string;
	    allow_insecure_http?: boolean;
	    ssh_host?: string;
	    ssh_port?: number;
	    ssh_user?: string;
	    ssh_auth_type?: string;
	    ssh_key_path?: string;
	    ssh_fingerprint?: string;
	    remote_bind_host?: string;
	    remote_panel_port?: number;
	    has_api_token: boolean;
	    has_ssh_password: boolean;
	    has_ssh_key_pem: boolean;
	    has_ssh_key_passphrase: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    // Go type: time
	    last_connected_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new ServerProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.mode = source["mode"];
	        this.direct_url = source["direct_url"];
	        this.allow_insecure_http = source["allow_insecure_http"];
	        this.ssh_host = source["ssh_host"];
	        this.ssh_port = source["ssh_port"];
	        this.ssh_user = source["ssh_user"];
	        this.ssh_auth_type = source["ssh_auth_type"];
	        this.ssh_key_path = source["ssh_key_path"];
	        this.ssh_fingerprint = source["ssh_fingerprint"];
	        this.remote_bind_host = source["remote_bind_host"];
	        this.remote_panel_port = source["remote_panel_port"];
	        this.has_api_token = source["has_api_token"];
	        this.has_ssh_password = source["has_ssh_password"];
	        this.has_ssh_key_pem = source["has_ssh_key_pem"];
	        this.has_ssh_key_passphrase = source["has_ssh_key_passphrase"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.last_connected_at = this.convertValues(source["last_connected_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

