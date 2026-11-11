# xray-panel-desktop

跨平台桌面控制端（Wails v2 + Go + Vue 3 + TypeScript + Pinia + Naive UI），用于在本地桌面集中管理多台远程无头 [`xray-panel-server`](../xray-panel-server) 节点。

## 核心架构与特性

- **多节点 Profile 与安全凭证存储**：
  - 仅在本地 `profiles.json` 存储连接元数据（名称、SSH 主机/端口/用户名、连接模式、TOFU 指纹等）。
  - 所有敏感凭证（API Token `xrp_...`、SSH 密码、私钥 PEM、私钥口令）均托管于操作系统原生凭据库（macOS Keychain / Windows Credential Manager / Linux Secret Service，`go-keyring`），绝不明文写入 JSON。
- **Go 进程内原生 SSH 隧道（零本地端口监听）**：
  - 通过 `golang.org/x/crypto/ssh` 的 `sshClient.Dial("tcp", remoteAddr)` 直接作为 Go `http.Transport.DialContext`，在内存内将 HTTP 请求隧道化转发至远端 `127.0.0.1:8082`。
  - 本地不开放任何监听端口，从根本上避免本机多进程/跨软件探测风险。
  - 支持 `key_file`（支持加密私钥 passphrase）、`key_pem`、`password`、`ssh-agent` 四种认证方式，内置 20s 心跳保活、TOFU 首次主机指纹确认与 MITM 密钥变更阻断。
  - 同时支持直连 HTTPS 模式（带非安全明文 HTTP 警告开关及本地 URL 校验）。
- **全功能可视化控制台**：
  - **节点配置 (`ProfilesView`)**：新增/编辑/删除节点配置、连接切换、TOFU 指纹管理。
  - **仪表盘 (`DashboardView`)**：CPU/内存/磁盘/运行时间、当月流量与服务实时状态。
  - **用户与订阅 (`UsersView`)**：用户 CRUD、流量/到期管控、订阅二维码与订阅链接一键复制。
  - **入站管理 (`InboundsView`)**：VLESS/Trojan/WireGuard 入站、UDS 与 Nginx 域名关联、本地密钥生成与 WG 客户端配置参考。
  - **出站管理 (`OutboundsView`)**：Direct / SOCKS5 / WireGuard / Trojan / VLESS / VMess 出站、分享链接及 `.conf` 解析导入、远端真连接延迟/出口 IP 测试。
  - **路由规则 (`RoutingView`)**：按优先级管理 6 类规则、一键导入预设模板、浏览远端 GeoSite / GeoIP 数据。
  - **域名与证书 (`DomainsView`)**：Direct / CDN / Reality 域名配置、扫描远端 ACME/LE 证书并支持单条或批量导入。
  - **WireGuard 内核组网 (`WireGuardView`)**：服务启停、端口冲突与 IP 转发检测、Peer 在线握手与流量监控、`wg0.conf` 与 Xray Outbound JSON 导出/另存。
  - **系统与 Xray (`SettingsView`)**：Server 模式参数、`server_public_host`、动态 `sub_path`、Xray 校验/热重载/重启、远端重启控制面并自动重连、实时导出 `config.json`。
  - **日志与 Token (`LogsView` / `TokensView`)**：分级过滤/搜索/自动刷新与日志导出；静态 `xrp_` API Token 签发、吊销与清理。

## 前置构建依赖

- Go `1.22+`
- Node.js `18+`（推荐 `20+`）与 `npm`
- [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation)（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）
- Linux 构建桌面端需安装系统 WebKit2GTK 与 GTK3 开发头文件（以及 `libsecret` 用于 `go-keyring`）

## 开发与打包命令

```bash
# 1. 拉取 Go 依赖并安装前端包（首次在联网环境执行）
go mod tidy
cd frontend && npm install && cd ..

# 2. 运行 Go 核心模块离线单元测试（store / tunnel / session / local）
go test -v ./internal/...

# 3. 前端类型检查与构建
cd frontend && npm run build && cd ..

# 4. 启动 Wails 热重载桌面开发窗口（Ubuntu 24.04 需加 -tags webkit2_41）
wails dev
# wails dev -tags webkit2_41

# 5. 构建桌面安装包 / 可执行文件（产物输出在 build/bin/）
wails build
# wails build -tags webkit2_41
```

> **提示**：使用内置 SSH 隧道连接服务端时，服务端无需额外运行隧道软件，只需确保 OpenSSH（`sshd`）允许 TCP 转发（`/etc/ssh/sshd_config` 中 `AllowTcpForwarding yes`）且能通过公钥/密码认证登录；详见 [`xray-panel-server/README.md`](../xray-panel-server/README.md#ssh-隧道接入服务端要求与排查)。

## 目录结构

```text
xray-panel-desktop/
├── app.go                 # Wails 绑定层（暴露 Profile、Tunnel、API 代理与本地工具方法）
├── main.go                # Wails v2 应用入口与嵌入前端资源
├── keyring_vault.go       # OS Keyring 凭据保险库适配器
├── wails.json             # Wails 构建配置
├── go.mod
├── internal/
│   ├── apiclient/         # 对接 xray-panel-server 的强类型 Go HTTP 客户端
│   ├── local/             # 本地 X25519 密钥对生成/推导与直连 URL 校验
│   ├── session/           # 多节点连接状态管理、延迟探测与重启重连逻辑
│   ├── store/             # profiles.json 持久化与 SecretVault 接口
│   └── tunnel/            # 进程内 SSH DialContext 隧道与 TOFU/MITM 校验
└── frontend/              # Vue 3 + TypeScript + Pinia + Naive UI + Tailwind CSS
    └── src/
        ├── api/           # Wails IPC 封装与 OpenAPI 对齐类型定义
        ├── layouts/       # 侧边栏、顶部多节点切换器与 TOFU 指纹弹窗
        ├── router/        # Hash 路由
        ├── stores/        # Pinia 连接状态与 GeoData 缓存
        └── views/         # 11 个业务视图页面
```
