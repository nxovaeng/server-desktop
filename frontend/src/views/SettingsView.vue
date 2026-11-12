<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between flex-wrap gap-2">
      <h2 class="text-xl font-semibold">系统设置与 Xray 配置</h2>
      <div class="flex gap-2">
        <n-button size="small" @click="load" :disabled="!conn.status.connected" :loading="loading">刷新</n-button>
        <n-button size="small" type="primary" :disabled="!conn.status.connected" :loading="saving" @click="save">
          保存设置
        </n-button>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <n-card title="服务端参数 (Server Mode)" size="small">
        <n-form label-placement="left" label-width="145" size="small">
          <n-form-item label="面板显示名称">
            <n-input v-model:value="form.panel_title" placeholder="Xray Panel" />
          </n-form-item>
          <n-form-item label="公网出口地址">
            <n-input
              v-model:value="form.server_public_host"
              class="mono"
              placeholder="服务器公网 IP 或默认域名（用于订阅与 WG Endpoint）"
            />
          </n-form-item>
          <n-form-item label="订阅域名 (Sub Domain)">
            <n-input v-model:value="form.sub_domain" class="mono" placeholder="sub.example.com（可留空回落公网地址）" />
          </n-form-item>
          <n-form-item label="订阅路径前缀">
            <n-input v-model:value="form.sub_path" class="mono" placeholder="/d（保存后动态热生效）" />
          </n-form-item>
          <n-form-item label="API 直连域名">
            <n-input v-model:value="form.api_domain" class="mono" placeholder="api.example.com（用于 HTTPS / CDN 直连控制面）" />
          </n-form-item>
          <n-form-item label="Xray 日志级别">
            <n-select
              v-model:value="form.xray_log_level"
              :options="[
                { label: 'debug', value: 'debug' },
                { label: 'info', value: 'info' },
                { label: 'warning (推荐)', value: 'warning' },
                { label: 'error', value: 'error' },
                { label: 'none', value: 'none' }
              ]"
            />
          </n-form-item>
          <n-form-item label="直连出站域名策略">
            <n-select
              v-model:value="form.direct_domain_strategy"
              :options="[
                { label: 'UseIPv4 (优先 IPv4)', value: 'UseIPv4' },
                { label: 'UseIP (解析后直连)', value: 'UseIP' },
                { label: 'UseIPv6 (优先 IPv6)', value: 'UseIPv6' },
                { label: 'AsIs (保持域名)', value: 'AsIs' }
              ]"
            />
          </n-form-item>
          <div class="grid grid-cols-2 gap-3">
            <n-form-item label="默认用户流量 (GB)">
              <n-input-number v-model:value="defaultLimitGB" :min="0" class="w-full" placeholder="0 = 无限" />
            </n-form-item>
            <n-form-item label="默认有效天数">
              <n-input-number v-model:value="form.default_expire_days" :min="0" class="w-full" placeholder="0 = 永不过期" />
            </n-form-item>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <n-form-item label="流量统计 (Stats)">
              <n-switch v-model:value="form.enable_traffic_stats" />
            </n-form-item>
            <n-form-item label="协议嗅探 (Sniffing)">
              <n-switch v-model:value="form.enable_sniffing" />
            </n-form-item>
          </div>
        </n-form>

        <div
          v-if="baselineInfo.month"
          class="mt-2 pt-2 border-t border-white/10 text-xs text-gray-400 flex flex-wrap gap-4"
        >
          <span>当月网卡基线月份：<span class="mono text-gray-200">{{ baselineInfo.month }}</span></span>
          <span>Rx Baseline：<span class="mono text-gray-200">{{ formatBytes(Number(baselineInfo.recv) || 0) }}</span></span>
          <span>Tx Baseline：<span class="mono text-gray-200">{{ formatBytes(Number(baselineInfo.sent) || 0) }}</span></span>
        </div>
      </n-card>

      <div class="space-y-4">
        <n-card title="Xray 与面板进程控制" size="small">
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <div class="text-sm text-gray-300">
                Xray 服务状态：
                <n-tag size="small" :type="xrayStatus?.active ? 'success' : 'error'">
                  {{ xrayStatus ? `${xrayStatus.active ? '运行中' : '未运行'} (${xrayStatus.status})` : '未知' }}
                </n-tag>
              </div>
              <n-button size="tiny" quaternary @click="refreshXrayStatus" :disabled="!conn.status.connected">
                检查状态
              </n-button>
            </div>

            <div class="flex flex-wrap gap-2">
              <n-button
                size="small"
                type="primary"
                secondary
                :disabled="!conn.status.connected"
                :loading="applying"
                @click="applyConfig(false)"
              >
                校验并应用 Xray 配置
              </n-button>
              <n-button
                size="small"
                type="info"
                secondary
                :disabled="!conn.status.connected"
                :loading="hotReloading"
                @click="applyConfig(true)"
              >
                热重载配置 (Hot)
              </n-button>
              <n-button
                size="small"
                type="warning"
                secondary
                :disabled="!conn.status.connected"
                :loading="restartingXray"
                @click="restartXray"
              >
                重启 Xray 服务
              </n-button>
            </div>

            <div class="pt-3 border-t border-white/10 flex items-center justify-between flex-wrap gap-2">
              <div class="text-xs text-gray-400">
                远程重启 <span class="mono">xray-panel-server</span> 进程，客户端将自动等待并重建 SSH/API 连接。
              </div>
              <n-popconfirm @positive-click="restartPanel">
                <template #trigger>
                  <n-button
                    size="small"
                    type="error"
                    secondary
                    :disabled="!conn.status.connected"
                    :loading="conn.connecting"
                  >
                    重启服务端并重连
                  </n-button>
                </template>
                确认远程重启 xray-panel-server？当前连接将短暂中断并自动重连。
              </n-popconfirm>
            </div>

            <div class="pt-3 border-t border-white/10 space-y-2">
              <div class="flex items-center justify-between flex-wrap gap-2">
                <div class="text-xs text-gray-400">
                  经当前 SSH 隧道执行远端 <span class="mono">install.sh update</span> 在线更新后端 API（仅 SSH 隧道模式可用，非
                  <span class="mono">root</span> 需免密 <span class="mono">sudo</span>）。
                  <span v-if="conn.status.server_version" class="ml-1 text-gray-300">
                    当前版本：<span class="mono text-emerald-400">v{{ conn.status.server_version }}</span>
                  </span>
                </div>
                <n-popconfirm @positive-click="updateBackendViaSSH">
                  <template #trigger>
                    <n-button
                      size="small"
                      type="primary"
                      secondary
                      :disabled="!conn.status.connected || conn.status.mode !== 'ssh_tunnel' || updatingBackend"
                      :loading="updatingBackend"
                    >
                      {{ conn.status.mode === 'ssh_tunnel' ? 'SSH 在线更新后端' : 'SSH 在线更新（需 SSH 隧道）' }}
                    </n-button>
                  </template>
                  确认通过当前 SSH 隧道在线下载安装最新 xray-panel-server Release 并重启服务？
                </n-popconfirm>
              </div>
              <div v-if="backendUpdateLog" class="space-y-1">
                <div class="flex items-center justify-between text-[11px] text-gray-400">
                  <span>远端 SSH 更新日志</span>
                  <n-button size="tiny" quaternary @click="backendUpdateLog = ''">收起</n-button>
                </div>
                <pre class="p-2.5 rounded bg-black/40 border border-white/10 mono text-xs overflow-auto max-h-56 whitespace-pre-wrap">{{ backendUpdateLog }}</pre>
              </div>
            </div>
          </div>
        </n-card>

        <n-card title="Xray config.json 实时预览" size="small">
          <template #header-extra>
            <div class="flex gap-2 items-center">
              <n-tag v-if="xrayDirty === true" type="warning" size="small">有未应用更改</n-tag>
              <n-tag v-else-if="xrayDirty === false" type="success" size="small">配置与运行中一致</n-tag>
              <n-button size="tiny" @click="loadXrayConfig" :disabled="!conn.status.connected" :loading="loadingConfig">
                生成预览
              </n-button>
              <n-button size="tiny" :disabled="!xrayRaw" @click="copyConfig">复制</n-button>
              <n-button size="tiny" type="primary" secondary :disabled="!xrayRaw" @click="exportConfig">
                导出 JSON
              </n-button>
            </div>
          </template>
          <pre
            v-if="xrayRaw"
            class="p-3 rounded bg-black/40 border border-white/10 mono text-xs overflow-auto max-h-96"
          >{{ xrayRaw }}</pre>
          <div v-else class="text-xs text-gray-400 py-4 text-center">
            点击右上角「生成预览」查看当前配置生成的 Xray config.json
          </div>
        </n-card>
      </div>
    </div>

    <n-card title="域名反代与公网暴露（API HTTPS/CDN · 订阅 · Cloudflare Tunnel）" size="small">
      <n-tabs type="line" animated size="small">
        <n-tab-pane name="api" tab="控制面 API 域名反代 (HTTPS / CDN)">
          <div class="space-y-3 text-xs">
            <div class="flex items-center justify-between flex-wrap gap-2">
              <div class="flex items-center gap-2 flex-wrap">
                <n-tag size="small" :type="form.api_nginx_enabled ? 'success' : 'default'">
                  {{ form.api_nginx_enabled ? 'API Nginx 反代：已启用' : 'API Nginx 反代：未启用' }}
                </n-tag>
                <span class="text-gray-300">
                  直连地址：
                  <span class="mono text-emerald-400">{{ form.api_domain ? `https://${form.api_domain}` : '(未填写 api_domain)' }}</span>
                </span>
                <n-button
                  v-if="form.api_domain"
                  size="tiny"
                  quaternary
                  @click="copyText(`https://${form.api_domain}`, '直连 API URL')"
                >
                  复制直连 URL
                </n-button>
              </div>
              <div class="flex gap-2">
                <n-button
                  size="small"
                  type="primary"
                  secondary
                  :disabled="!conn.status.connected"
                  :loading="applyingAPINginx"
                  @click="enableAPINginx"
                >
                  保存域名并开启 API 反代
                </n-button>
                <n-button
                  v-if="form.api_nginx_enabled"
                  size="small"
                  type="warning"
                  secondary
                  :disabled="!conn.status.connected"
                  :loading="disablingAPINginx"
                  @click="disableAPINginx"
                >
                  停用 API 反代
                </n-button>
              </div>
            </div>
            <n-alert type="info" size="small">
              生成仅放行 <span class="mono">/api/v1/</span> 的 Nginx HTTPS 反代（自动关闭 SSE 缓冲以支持实时日志，根路径 <span class="mono">/</span> 返回 404）；支持过 Cloudflare CDN（推荐上传 Cloudflare Origin CA 证书或签发 ACME 证书并设 SSL 为 <span class="mono">Full (Strict)</span>），开启后可在桌面端新增节点时选择「HTTPS / 本地直连」直接接入。如与订阅或入站共用同一域名，会自动合并于同一份 Nginx 站点配置。
            </n-alert>
            <div v-if="apiNginxResult" class="p-2.5 rounded bg-black/40 border border-white/10 text-xs text-gray-300">
              {{ apiNginxResult }}
            </div>
          </div>
        </n-tab-pane>

        <n-tab-pane name="nginx" tab="订阅域名 Nginx 反代">
          <div class="space-y-3 text-xs">
            <div class="flex items-center justify-between flex-wrap gap-2">
              <div class="flex items-center gap-2 flex-wrap">
                <n-tag size="small" :type="form.sub_nginx_enabled ? 'success' : 'default'">
                  {{ form.sub_nginx_enabled ? '订阅 Nginx 反代：已启用' : '订阅 Nginx 反代：未启用' }}
                </n-tag>
                <span class="text-gray-300">
                  当前目标：<span class="mono">{{ form.sub_domain || '(未填写 sub_domain)' }}</span> · 仅放行前缀
                  <span class="mono text-emerald-400">{{ form.sub_path || '/d' }}/</span>
                </span>
              </div>
              <div class="flex gap-2">
                <n-button
                  size="small"
                  type="primary"
                  secondary
                  :disabled="!conn.status.connected"
                  :loading="applyingSubNginx"
                  @click="enableSubNginx"
                >
                  保存域名并应用订阅 Nginx
                </n-button>
                <n-button
                  v-if="form.sub_nginx_enabled"
                  size="small"
                  type="warning"
                  secondary
                  :disabled="!conn.status.connected"
                  :loading="disablingSubNginx"
                  @click="disableSubNginx"
                >
                  停用反代
                </n-button>
              </div>
            </div>
            <n-alert type="info" size="small">
              仅代理 <span class="mono">{{ form.sub_path || '/d' }}/</span> 到面板回环地址（<span class="mono">/api/v1/*</span> 在公网访问返回 404）；会自动从「域名证书管理」中精确匹配或通配符匹配 TLS 证书。如订阅域名同时绑定了 Xray 入站，也会自动合并保留于同一份站点配置中。
            </n-alert>
            <div v-if="subNginxResult" class="p-2.5 rounded bg-black/40 border border-white/10 text-xs text-gray-300">
              {{ subNginxResult }}
            </div>
          </div>
        </n-tab-pane>

        <n-tab-pane name="cf" tab="Cloudflare Tunnel (免开放 443 与服务器证书)">
          <div class="space-y-3 text-xs">
            <div class="flex items-center justify-between flex-wrap gap-2">
              <div class="flex items-center gap-2 flex-wrap">
                <n-tag size="small" :type="cfStatus?.installed ? 'success' : 'default'">
                  {{ cfStatus?.installed ? 'cloudflared 已安装' : '未安装 cloudflared' }}
                </n-tag>
                <n-tag size="small" :type="cfStatus?.active ? 'success' : 'warning'">
                  服务状态：{{ cfStatus?.status || '未知' }}
                </n-tag>
                <span v-if="cfStatus?.version" class="mono text-gray-400">{{ cfStatus.version }}</span>
              </div>
              <div class="flex gap-2 flex-wrap">
                <n-button
                  size="small"
                  type="info"
                  secondary
                  :disabled="!conn.status.connected"
                  :loading="installingCF"
                  @click="installCF"
                >
                  一键安装 cloudflared
                </n-button>
                <n-button size="small" quaternary :disabled="!conn.status.connected" @click="refreshCFStatus">
                  刷新状态
                </n-button>
              </div>
            </div>

            <div class="flex items-center gap-2 flex-wrap">
              <n-input
                v-model:value="cfToken"
                type="password"
                show-password-on="click"
                size="small"
                class="mono flex-1 min-w-[280px]"
                placeholder="粘贴 Cloudflare Zero Trust Tunnel Token（或完整 cloudflared service install ... 命令）"
              />
              <n-button
                size="small"
                type="primary"
                :disabled="!conn.status.connected"
                :loading="settingUpCF"
                @click="setupCF"
              >
                配置并启动 Tunnel
              </n-button>
              <n-button
                size="small"
                :disabled="!conn.status.connected || !cfStatus?.service_installed"
                :loading="cfActionLoading === 'restart'"
                @click="cfAction('restart')"
              >
                重启
              </n-button>
              <n-button
                size="small"
                type="warning"
                secondary
                :disabled="!conn.status.connected || !cfStatus?.service_installed"
                :loading="cfActionLoading === 'stop'"
                @click="cfAction('stop')"
              >
                停止
              </n-button>
              <n-popconfirm @positive-click="cfAction('uninstall')">
                <template #trigger>
                  <n-button
                    size="small"
                    type="error"
                    quaternary
                    :disabled="!conn.status.connected || !cfStatus?.service_installed"
                    :loading="cfActionLoading === 'uninstall'"
                  >
                    卸载服务
                  </n-button>
                </template>
                确认从服务端卸载 cloudflared 系统服务？
              </n-popconfirm>
            </div>

            <div v-if="cfStatus" class="p-3 rounded bg-black/40 border border-white/10 space-y-1.5 text-gray-300">
              <div class="font-medium text-gray-200">Cloudflare Zero Trust 控制台 Public Hostname 推荐映射：</div>
              <div class="flex flex-wrap gap-4">
                <span>Subdomain / Domain：<span class="mono text-emerald-300">{{ form.sub_domain || 'sub.example.com' }}</span></span>
                <span>
                  Path (强烈建议限制仅暴露订阅)：
                  <span class="mono text-emerald-300">{{ cfStatus.recommended_path_rule }}</span>
                </span>
                <span>
                  Service：
                  <span class="mono text-emerald-300">{{ cfStatus.recommended_service }}</span>
                </span>
              </div>
              <div class="text-gray-400">
                仅将 <span class="mono">{{ cfStatus.recommended_path_rule }}</span> 指向 <span class="mono">{{ cfStatus.recommended_service }}</span> 即可让客户端访问订阅链接，同时保证 <span class="mono">/api/v1/*</span> 不暴露在公网。
              </div>
            </div>
          </div>
        </n-tab-pane>
      </n-tabs>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { copyToClipboard, desktopApi, formatBytes, formatError } from '@/api/client'
import type { CFTunnelStatus, DirectDomainStrategy, XrayLogLevel, XrayStatusResponse } from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const loading = ref(false)
const saving = ref(false)
const applying = ref(false)
const hotReloading = ref(false)
const restartingXray = ref(false)
const updatingBackend = ref(false)
const backendUpdateLog = ref('')
const loadingConfig = ref(false)
const applyingSubNginx = ref(false)
const disablingSubNginx = ref(false)
const subNginxResult = ref('')
const applyingAPINginx = ref(false)
const disablingAPINginx = ref(false)
const apiNginxResult = ref('')

const cfStatus = ref<CFTunnelStatus | null>(null)
const cfToken = ref('')
const installingCF = ref(false)
const settingUpCF = ref(false)
const cfActionLoading = ref('')

const xrayStatus = ref<XrayStatusResponse | null>(null)
const xrayRaw = ref('')
// 脏状态：true=有未应用更改，false=与运行中一致，null=未知（后端未返回 current_hash 时不展示）
const xrayDirty = ref<boolean | null>(null)
const defaultLimitGB = ref(0)

const baselineInfo = reactive({
  month: '',
  sent: '',
  recv: ''
})

const form = reactive({
  panel_title: 'Xray Panel',
  sub_domain: '',
  sub_path: '/d',
  sub_nginx_enabled: false,
  api_domain: '',
  api_nginx_enabled: false,
  server_public_host: '',
  xray_log_level: 'warning' as XrayLogLevel,
  direct_domain_strategy: 'UseIPv4' as DirectDomainStrategy,
  enable_traffic_stats: true,
  enable_sniffing: true,
  default_expire_days: 30
})

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const [settings, st, cfg, cf] = await Promise.all([
      desktopApi.getSettings(),
      desktopApi.getXrayStatus(),
      desktopApi.getXrayConfig(),
      desktopApi.getCFTunnelStatus()
    ])
    Object.assign(form, {
      panel_title: settings.panel_title,
      sub_domain: settings.sub_domain,
      sub_path: settings.sub_path,
      sub_nginx_enabled: settings.sub_nginx_enabled,
      api_domain: settings.api_domain,
      api_nginx_enabled: settings.api_nginx_enabled,
      server_public_host: settings.server_public_host,
      xray_log_level: settings.xray_log_level,
      direct_domain_strategy: settings.direct_domain_strategy,
      enable_traffic_stats: settings.enable_traffic_stats,
      enable_sniffing: settings.enable_sniffing,
      default_expire_days: settings.default_expire_days
    })
    defaultLimitGB.value = Math.round((settings.default_traffic_limit || 0) / (1024 * 1024 * 1024))
    baselineInfo.month = settings.monthly_net_baseline_month
    baselineInfo.sent = settings.monthly_net_baseline_sent
    baselineInfo.recv = settings.monthly_net_baseline_recv
    xrayStatus.value = st
    xrayRaw.value = cfg.raw
    xrayDirty.value = cfg.current_hash !== undefined ? !!cfg.dirty : null
    cfStatus.value = cf
  } catch (err) {
    message.error('加载设置失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    const updated = await desktopApi.updateSettings({
      panel_title: form.panel_title.trim(),
      sub_domain: form.sub_domain.trim(),
      sub_path: form.sub_path.trim(),
      api_domain: form.api_domain.trim(),
      server_public_host: form.server_public_host.trim(),
      xray_log_level: form.xray_log_level,
      direct_domain_strategy: form.direct_domain_strategy,
      enable_traffic_stats: form.enable_traffic_stats,
      enable_sniffing: form.enable_sniffing,
      default_traffic_limit: Math.max(0, Math.round(defaultLimitGB.value || 0)) * 1024 * 1024 * 1024,
      default_expire_days: Math.max(0, Math.round(form.default_expire_days || 0))
    })
    conn.status.panel_title = updated.panel_title
    conn.status.sub_domain = updated.sub_domain
    conn.status.sub_path = updated.sub_path
    conn.status.server_public_host = updated.server_public_host
    message.success('系统设置已保存')
  } catch (err) {
    message.error('保存设置失败: ' + formatError(err))
  } finally {
    saving.value = false
  }
}

async function refreshXrayStatus() {
  try {
    xrayStatus.value = await desktopApi.getXrayStatus()
  } catch (err) {
    message.error('获取 Xray 状态失败: ' + formatError(err))
  }
}

async function loadXrayConfig() {
  loadingConfig.value = true
  try {
    const res = await desktopApi.getXrayConfig()
    xrayRaw.value = res.raw
    xrayDirty.value = res.current_hash !== undefined ? !!res.dirty : null
  } catch (err) {
    message.error('生成 Xray 配置失败: ' + formatError(err))
  } finally {
    loadingConfig.value = false
  }
}

async function applyConfig(hot: boolean) {
  if (hot) {
    hotReloading.value = true
  } else {
    applying.value = true
  }
  try {
    const res = await desktopApi.applyXrayConfig(hot)
    message.success(res.message || (hot ? '热重载成功' : '已校验并应用 Xray 配置'))
    await Promise.all([refreshXrayStatus(), loadXrayConfig()])
  } catch (err) {
    message.error('应用配置失败: ' + formatError(err))
  } finally {
    applying.value = false
    hotReloading.value = false
  }
}

async function restartXray() {
  restartingXray.value = true
  try {
    const res = await desktopApi.restartXray()
    xrayStatus.value = { active: res.active, status: res.status }
    message.success(`Xray 服务已重启 (${res.status})`)
  } catch (err) {
    message.error('重启 Xray 失败: ' + formatError(err))
  } finally {
    restartingXray.value = false
  }
}

async function restartPanel() {
  try {
    const st = await conn.restartPanelAndReconnect()
    message.success(`面板已重启并重新连接 (${st.latency_ms} ms)`)
    await load()
  } catch (err) {
    message.error('重启或重连失败: ' + formatError(err))
  }
}

async function updateBackendViaSSH() {
  updatingBackend.value = true
  backendUpdateLog.value = ''
  try {
    const res = await conn.updateBackendViaSSH()
    backendUpdateLog.value = res.output || ''
    if (!res.success) {
      message.error(res.error || 'SSH 在线更新失败')
      return
    }
    if (res.error) {
      message.warning(res.error)
    } else {
      const ver = res.status?.server_version ? `v${res.status.server_version}` : '已刷新'
      message.success(`后端 API 在线更新完成 (${ver} · ${res.status?.latency_ms ?? 0} ms)`)
    }
    if (res.status?.connected) {
      await load()
    }
  } catch (err) {
    message.error('SSH 在线更新失败: ' + formatError(err))
  } finally {
    updatingBackend.value = false
  }
}

async function copyConfig() {
  await copyToClipboard(xrayRaw.value)
  message.success('config.json 已复制到剪贴板')
}

async function exportConfig() {
  try {
    const path = await desktopApi.saveTextFile('config.json', xrayRaw.value)
    if (path) {
      message.success(`已保存至 ${path}`)
    }
  } catch (err) {
    message.error('导出失败: ' + formatError(err))
  }
}

async function enableSubNginx() {
  if (!form.sub_domain.trim()) {
    message.warning('请先在左侧表单填写订阅域名 (Sub Domain)')
    return
  }
  applyingSubNginx.value = true
  subNginxResult.value = ''
  try {
    const updated = await desktopApi.updateSettings({
      sub_domain: form.sub_domain.trim(),
      sub_path: form.sub_path.trim()
    })
    conn.status.sub_domain = updated.sub_domain
    conn.status.sub_path = updated.sub_path
    const res = await desktopApi.applySubscriptionNginx()
    form.sub_nginx_enabled = res.enabled
    subNginxResult.value = `${res.message} · 匹配证书：${res.matched_cert_domain} · 配置文件：${res.config_path}`
    message.success(res.message)
  } catch (err) {
    message.error('应用订阅 Nginx 失败: ' + formatError(err))
  } finally {
    applyingSubNginx.value = false
  }
}

async function disableSubNginx() {
  disablingSubNginx.value = true
  try {
    const res = await desktopApi.disableSubscriptionNginx()
    form.sub_nginx_enabled = res.enabled
    subNginxResult.value = res.message
    message.success(res.message)
  } catch (err) {
    message.error('停用失败: ' + formatError(err))
  } finally {
    disablingSubNginx.value = false
  }
}

async function copyText(text: string, label: string) {
  await copyToClipboard(text)
  message.success(`${label}已复制到剪贴板`)
}

async function enableAPINginx() {
  if (!form.api_domain.trim()) {
    message.warning('请先在左侧表单填写 API 直连域名')
    return
  }
  applyingAPINginx.value = true
  apiNginxResult.value = ''
  try {
    await desktopApi.updateSettings({
      api_domain: form.api_domain.trim()
    })
    const res = await desktopApi.applyAPINginx()
    form.api_nginx_enabled = res.enabled
    apiNginxResult.value = `${res.message} · 直连 URL：${res.direct_url} · 匹配证书：${res.matched_cert_domain} · 配置文件：${res.config_path}`
    message.success(res.message)
  } catch (err) {
    message.error('应用 API Nginx 反代失败: ' + formatError(err))
  } finally {
    applyingAPINginx.value = false
  }
}

async function disableAPINginx() {
  disablingAPINginx.value = true
  try {
    const res = await desktopApi.disableAPINginx()
    form.api_nginx_enabled = res.enabled
    apiNginxResult.value = res.message
    message.success(res.message)
  } catch (err) {
    message.error('停用失败: ' + formatError(err))
  } finally {
    disablingAPINginx.value = false
  }
}

async function refreshCFStatus() {
  try {
    cfStatus.value = await desktopApi.getCFTunnelStatus()
  } catch (err) {
    message.error('查询 cloudflared 状态失败: ' + formatError(err))
  }
}

async function installCF() {
  installingCF.value = true
  try {
    const res = await desktopApi.installCFTunnel()
    message.success(`${res.message} (${res.binary_path})`)
    await refreshCFStatus()
  } catch (err) {
    message.error('安装 cloudflared 失败: ' + formatError(err))
  } finally {
    installingCF.value = false
  }
}

async function setupCF() {
  if (!cfToken.value.trim()) {
    message.warning('请粘贴 Cloudflare Tunnel Token')
    return
  }
  settingUpCF.value = true
  try {
    const res = await desktopApi.setupCFTunnel(cfToken.value.trim())
    cfStatus.value = res.status
    cfToken.value = ''
    message.success(res.message)
  } catch (err) {
    message.error('配置 Tunnel 失败: ' + formatError(err))
  } finally {
    settingUpCF.value = false
  }
}

async function cfAction(action: 'start' | 'stop' | 'restart' | 'uninstall') {
  cfActionLoading.value = action
  try {
    const res = await desktopApi.cfTunnelServiceAction(action)
    cfStatus.value = res.status
    message.success(res.message)
  } catch (err) {
    message.error(`执行 ${action} 失败: ` + formatError(err))
  } finally {
    cfActionLoading.value = ''
  }
}

watch(
  () => conn.status.profile_id,
  (id) => {
    if (id && conn.status.connected) {
      load()
    }
  }
)

onMounted(load)
</script>
