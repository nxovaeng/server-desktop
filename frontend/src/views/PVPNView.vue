<template>
  <div class="space-y-4">
    <!-- 顶部标题与操作栏 -->
    <div class="flex items-center justify-between flex-wrap gap-2">
      <div>
        <div class="flex items-center gap-2">
          <h2 class="text-xl font-semibold">Proton 出口代理</h2>
          <n-tag size="small" type="primary" secondary>Go 原生内嵌引擎</n-tag>
          <n-tag size="small" :type="status?.running ? 'success' : 'default'">
            {{ status?.running ? 'WireGuard 运行中' : '空闲待命中' }}
          </n-tag>
        </div>
        <div class="text-xs text-muted mt-1">
          服务端原生集成 Proton VPN 游客协议，基于 WireGuard 映射至本地 SOCKS5 (127.0.0.1:{{ status?.port || 1080 }}) 出口
        </div>
      </div>
      <div class="flex gap-2">
        <n-button size="small" @click="loadData" :disabled="!conn.status.connected" :loading="loading">
          刷新状态
        </n-button>
      </div>
    </div>

    <!-- 状态概览网格 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
      <!-- 卡片 1: 运行状态与出口地区 -->
      <n-card size="small">
        <div class="text-xs text-muted mb-1">出口节点与地区</div>
        <div class="flex items-center gap-2">
          <n-tag size="small" :type="status?.running ? 'success' : 'default'">
            {{ status?.running ? '已连接' : '未连接' }}
          </n-tag>
          <span class="text-xs text-strong font-medium truncate">
            {{ status?.server && status.server !== 'None' ? status.server : '未选定节点' }}
          </span>
        </div>
        <div class="mt-2 text-[11px] text-muted flex items-center justify-between">
          <span>出口国家：</span>
          <span class="mono text-strong text-accent">
            {{ status?.country ? getCountryLabel(status.country) : (status?.running ? '自动获取中' : '无') }}
          </span>
        </div>
      </n-card>

      <!-- 卡片 2: 出口公网 IP -->
      <n-card size="small">
        <div class="text-xs text-muted mb-1">探测出口 IP (SOCKS5 出口)</div>
        <div class="mono text-sm font-semibold text-strong truncate">
          {{ status?.ip || 'N/A' }}
        </div>
        <div class="mt-2 text-[11px] text-muted flex items-center justify-between">
          <span>本地监听：</span>
          <div class="flex items-center gap-1">
            <span class="mono">127.0.0.1:{{ status?.port || 1080 }}</span>
            <n-button size="tiny" text type="primary" @click="handleCopyAddr">复制</n-button>
            <n-button size="tiny" text type="warning" @click="openPortModal">修改</n-button>
          </div>
        </div>
      </n-card>

      <!-- 卡片 3: Proton 凭证与 24h 动态轮换 -->
      <n-card size="small">
        <div class="text-xs text-muted mb-1">Proton 凭证与 24h 轮换</div>
        <div class="flex items-center gap-2 text-xs">
          <n-tag size="tiny" :type="status?.logged_in ? 'info' : 'default'">
            {{ status?.logged_in ? 'Guest 访客已登录' : '未登录 (待初始化)' }}
          </n-tag>
          <span class="text-secondary text-[11px]">
            {{ getExpiryCountdown(status?.cert_expires_at) }}
          </span>
        </div>
        <div class="mt-2 text-[11px] text-muted flex items-center justify-between">
          <span class="truncate">到期：<span class="mono">{{ formatExpiry(status?.cert_expires_at) }}</span></span>
          <n-button
            size="tiny"
            type="primary"
            secondary
            :disabled="!conn.status.connected || loggingIn"
            :loading="loggingIn"
            @click="handleLogin"
          >
            {{ status?.logged_in ? '刷新凭据' : '立即登录' }}
          </n-button>
        </div>
      </n-card>

      <!-- 卡片 4: Xray 链路对齐 -->
      <n-card size="small">
        <div class="text-xs text-muted mb-1">Xray 出站关联</div>
        <div class="flex items-center gap-1.5 text-xs">
          <n-tag size="small" :type="status?.outbound_exists ? 'success' : 'warning'">
            {{ status?.outbound_exists ? '✔ proton 出站已就绪' : '⚠ 未创建出站' }}
          </n-tag>
          <span class="mono text-[11px] text-secondary">tag: {{ status?.outbound_tag || 'proton' }}</span>
        </div>
        <div class="mt-2 flex items-center justify-between">
          <span class="text-[11px] text-muted">指向 SOCKS5 :{{ status?.port || 1080 }}</span>
          <n-button
            v-if="!status?.outbound_exists"
            size="tiny"
            type="primary"
            secondary
            :loading="ensuringOutbound"
            @click="handleEnsureOutbound"
          >
            一键补全出站
          </n-button>
          <router-link v-else to="/routing" class="text-[11px] text-accent hover:underline">
            配置分流规则 →
          </router-link>
        </div>
      </n-card>
    </div>

    <!-- 主体区域：切换出口地区 & 原生服务状态 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- 左侧 2 列：切换出口地区 -->
      <div class="lg:col-span-2 space-y-4">
        <n-card title="切换出口地区 (Switch Exit Region)" size="small">
          <div class="space-y-4">
            <div class="text-xs text-muted">
              从 Proton VPN 免费节点中选择目标出口，服务端将自动维护 WireGuard 会话并秒级建立连接。
            </div>

            <!-- 国家选择下拉框 -->
            <n-form-item label="目标出口国家 / 地区" :show-feedback="false">
              <n-select
                v-model:value="selectedCountry"
                filterable
                placeholder="搜索或选择出口国家 (如 美国 US、日本 JP、荷兰 NL)"
                :options="countryOptions"
                :loading="loadingCountries"
                @update:value="onCountryChange"
              />
            </n-form-item>

            <!-- 常用快捷选择徽章 -->
            <div class="flex items-center gap-1.5 flex-wrap text-xs">
              <span class="text-muted mr-1">快捷地区：</span>
              <n-button
                v-for="c in quickCountries"
                :key="c.code"
                size="tiny"
                :type="selectedCountry === c.code ? 'primary' : 'default'"
                :secondary="selectedCountry === c.code"
                @click="quickSelectCountry(c.code)"
              >
                {{ c.label }}
              </n-button>
            </div>

            <!-- 节点选择模式 -->
            <div class="pt-2 border-t border-subtle space-y-3">
              <div class="flex items-center justify-between flex-wrap gap-2 text-xs">
                <span class="font-medium text-secondary">节点选取策略：</span>
                <n-radio-group v-model:value="connectMode" size="small">
                  <n-radio-button value="auto">自动选择最低负载节点 (推荐)</n-radio-button>
                  <n-radio-button value="manual">手动指定具体服务器</n-radio-button>
                </n-radio-group>
              </div>

              <!-- 手动指定具体服务器下拉框 -->
              <div v-if="connectMode === 'manual'" class="space-y-2">
                <n-form-item label="具体服务器节点" :show-feedback="false">
                  <n-select
                    v-model:value="selectedServer"
                    filterable
                    placeholder="选择具体服务器节点 (按实时负载从低到高排序)"
                    :options="serverOptions"
                    :loading="loadingServers"
                  />
                </n-form-item>
              </div>
            </div>

            <!-- 操作按钮组 -->
            <div class="pt-2 border-t border-subtle flex items-center justify-between flex-wrap gap-2">
              <div class="flex items-center gap-2">
                <n-switch v-model:value="autoEnsureOutbound" size="small" />
                <span class="text-xs text-muted">连接时自动确保 Xray SOCKS5 出站已创建</span>
              </div>

              <div class="flex gap-2">
                <n-button
                  size="small"
                  type="primary"
                  :disabled="!conn.status.connected || !selectedCountry || connecting"
                  :loading="connecting"
                  @click="handleConnect"
                >
                  连接所选地区
                </n-button>
                <n-button
                  size="small"
                  type="warning"
                  secondary
                  :disabled="!conn.status.connected || !status?.running || disconnecting"
                  :loading="disconnecting"
                  @click="handleDisconnect"
                >
                  断开代理
                </n-button>
              </div>
            </div>

            <!-- 提示信息 -->
            <n-alert type="info" size="small">
              <div class="space-y-1 text-xs">
                <div>• <strong>凭据与会话落盘</strong>：访客身份凭证与 WireGuard 私钥持久化于服务端 <code class="mono">{{ status?.session_path || '/opt/xray-panel/proton/session.json' }}</code>（权限 0600）。</div>
                <div>• <strong>登录时机与手动触发</strong>：拉取节点或首次建立连接时自动按需登录；亦可随时在上方卡片点击「刷新凭据 / 立即登录」手动重新获取访客身份与全新 24 小时证书。</div>
                <div>• <strong>24 小时动态轮换机制</strong>：Proton 官方证书有效期固定为 24 小时，服务端后台看门狗每 5 分钟轮询，在剩余有效期不足 2 小时时静默申请新证书并热重载 WireProxy，免去人工干预。</div>
                <div>• <strong>链路对齐与自愈</strong>：可在修改本地 SOCKS5 端口后自动同步 Xray <span class="mono text-accent">proton</span> 出站规则；每 30 秒健康检测并在节点故障时秒级故障转移。</div>
              </div>
            </n-alert>
          </div>
        </n-card>
      </div>

      <!-- 右侧 1 列：原生引擎与环境就绪状态 -->
      <div class="space-y-4">
        <n-card title="原生引擎与 WireProxy 状态" size="small">
          <div class="space-y-3 text-xs">
            <div class="flex items-center justify-between">
              <span class="text-muted">集成架构：</span>
              <span class="text-strong text-accent">Go 原生内嵌引擎</span>
            </div>

            <div class="flex items-center justify-between">
              <span class="text-muted">WireProxy 运行状态：</span>
              <n-tag size="small" :type="status?.running ? 'success' : 'default'">
                {{ status?.running ? '运行中 (Active)' : '空闲 / 未连接' }}
              </n-tag>
            </div>

            <div class="flex items-center justify-between">
              <span class="text-muted">底层二进制就绪：</span>
              <n-tag size="small" :type="status?.installed ? 'success' : 'warning'">
                {{ status?.installed ? '✔ wireproxy 已就绪' : '⚠ 未检测到二进制' }}
              </n-tag>
            </div>

            <div class="flex items-center justify-between">
              <span class="text-muted">可用服务器总数：</span>
              <span class="mono text-strong">{{ countries.reduce((acc, c) => acc + c.count, 0) || status?.total_servers || '已同步' }} 节点</span>
            </div>

            <div class="flex items-center justify-between">
              <span class="text-muted">凭据落盘位置：</span>
              <span class="mono text-[11px] text-strong truncate max-w-[170px]" :title="status?.session_path || '/opt/xray-panel/proton/session.json'">
                {{ status?.session_path || '/opt/xray-panel/proton/session.json' }}
              </span>
            </div>

            <!-- 控制按钮 -->
            <div class="pt-3 border-t border-subtle flex flex-wrap gap-2">
              <n-button
                size="small"
                :disabled="!conn.status.connected || !status?.installed"
                :loading="serviceActionLoading === 'restart'"
                @click="handleServiceAction('restart')"
              >
                重启代理
              </n-button>
              <n-button
                size="small"
                secondary
                type="warning"
                :disabled="!conn.status.connected || !status?.running"
                :loading="serviceActionLoading === 'stop'"
                @click="handleServiceAction('stop')"
              >
                停止代理
              </n-button>
              <n-button
                size="small"
                secondary
                type="primary"
                :disabled="!conn.status.connected"
                :loading="loadingCountries"
                @click="loadCountries"
              >
                重新拉取节点
              </n-button>
            </div>

            <!-- 一键部署 wireproxy 引擎 -->
            <div v-if="!status?.installed" class="pt-3 border-t border-subtle space-y-2">
              <div class="text-xs text-muted">
                服务端尚未检测到静态单二进制 wireproxy。点击下方按钮可在远端一键自动下载并部署至系统，无需任何 Python 或依赖环境。
              </div>
              <n-popconfirm @positive-click="handleInstall">
                <template #trigger>
                  <n-button size="small" type="primary" :loading="installing">
                    一键部署 wireproxy 静态引擎
                  </n-button>
                </template>
                确认在服务端自动下载并安装静态单二进制 wireproxy？
              </n-popconfirm>
            </div>
          </div>
        </n-card>

        <!-- 快速配置提示卡 -->
        <n-card size="small" title="分流集成说明">
          <div class="text-xs text-muted space-y-2">
            <div>本地 SOCKS5 地址：<code class="mono text-strong">127.0.0.1:{{ status?.port || 1080 }}</code></div>
            <div>Xray 推荐出站标签：<code class="mono text-strong">proton</code></div>
            <div class="text-[11px] leading-relaxed">
              可以在「路由规则」中添加域名规则（如 OpenAI / Netflix / 目标 IP 范围）分流至 <span class="mono text-accent">proton</span>，无需全局走 VPN。
            </div>
          </div>
        </n-card>
      </div>
    </div>

    <!-- 修改 SOCKS5 监听端口弹窗 -->
    <n-modal
      v-model:show="showPortModal"
      preset="card"
      title="配置 Proton SOCKS5 监听端口"
      size="small"
      style="max-width: 420px"
    >
      <div class="space-y-4">
        <div class="text-xs text-muted">
          修改 WireProxy 本地 SOCKS5 监听端口。保存后将自动更新服务端系统设置、更新 Xray 的 proton 出站端口，并热重启 WireProxy 代理进程。
        </div>
        <n-form-item label="SOCKS5 监听端口 (1-65535)" :show-feedback="false">
          <n-input-number
            v-model:value="customPort"
            :min="1"
            :max="65535"
            placeholder="默认 1080"
            class="w-full"
          />
        </n-form-item>
        <div class="flex justify-end gap-2 pt-2">
          <n-button size="small" @click="showPortModal = false">取消</n-button>
          <n-button
            size="small"
            type="primary"
            :loading="savingPort"
            @click="handleSavePort"
          >
            保存并应用
          </n-button>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { copyToClipboard, desktopApi, formatError } from '@/api/client'
import type { PVPNCountryDTO, PVPNServerDTO, PVPNStatusResponse } from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const loading = ref(false)
const loadingCountries = ref(false)
const loadingServers = ref(false)
const connecting = ref(false)
const disconnecting = ref(false)
const ensuringOutbound = ref(false)
const installing = ref(false)
const serviceActionLoading = ref('')
const loggingIn = ref(false)

const showPortModal = ref(false)
const customPort = ref<number | null>(1080)
const savingPort = ref(false)

const status = ref<PVPNStatusResponse | null>(null)
const countries = ref<PVPNCountryDTO[]>([])
const servers = ref<PVPNServerDTO[]>([])

const selectedCountry = ref('US')
const selectedServer = ref('')
const connectMode = ref<'auto' | 'manual'>('auto')
const autoEnsureOutbound = ref(true)

// 常用快捷国家
const quickCountries = [
  { code: 'US', label: '🇺🇸 美国 (US)' },
  { code: 'JP', label: '🇯🇵 日本 (JP)' },
  { code: 'NL', label: '🇳🇱 荷兰 (NL)' },
  { code: 'CH', label: '🇨🇭 瑞士 (CH)' },
  { code: 'SG', label: '🇸🇬 新加坡 (SG)' },
  { code: 'GB', label: '🇬🇧 英国 (GB)' },
  { code: 'DE', label: '🇩🇪 德国 (DE)' }
]

const countryOptions = computed(() =>
  countries.value.map((c) => ({
    label: `${getFlagEmoji(c.code)} ${c.name} (${c.code}) · ${c.count} 节点 · 均载 ${c.avg_load}%`,
    value: c.code
  }))
)

const serverOptions = computed(() =>
  servers.value.map((s) => ({
    label: `${s.name} · 负载 ${s.load}% · IP: ${s.entry_ip}`,
    value: s.name
  }))
)

function getFlagEmoji(countryCode: string) {
  if (!countryCode || countryCode.length !== 2) return '🌐'
  const codePoints = countryCode
    .toUpperCase()
    .split('')
    .map((char) => 127397 + char.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

function getCountryLabel(code: string) {
  const found = countries.value.find((c) => c.code === code)
  if (found) return `${getFlagEmoji(code)} ${found.name}`
  const q = quickCountries.find((c) => c.code === code)
  if (q) return q.label
  return `${getFlagEmoji(code)} ${code}`
}

function formatExpiry(ts?: number) {
  if (!ts) return '未签发'
  const date = new Date(ts * 1000)
  return date.toLocaleString()
}

function getExpiryCountdown(ts?: number) {
  if (!ts) return '凭证未生效'
  const now = Math.floor(Date.now() / 1000)
  const diff = ts - now
  if (diff <= 0) return '已过期 (将自动重新签发)'
  const hours = Math.floor(diff / 3600)
  const mins = Math.floor((diff % 3600) / 60)
  if (hours < 2) {
    return `剩余 ${hours}h ${mins}m · 即将静默续签`
  }
  return `有效剩余 ${hours}h ${mins}m`
}

async function handleCopyAddr() {
  const port = status.value?.port || 1080
  try {
    await copyToClipboard(`127.0.0.1:${port}`)
    message.success(`已复制 SOCKS5 地址 127.0.0.1:${port}`)
  } catch (err) {
    message.error('复制失败: ' + formatError(err))
  }
}

async function loadData() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    status.value = await desktopApi.getPVPNStatus()
    if (status.value.country) {
      selectedCountry.value = status.value.country
    }
    await loadCountries()
  } catch (err) {
    message.error('加载 Proton 状态失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

async function loadCountries() {
  loadingCountries.value = true
  try {
    const res = await desktopApi.getPVPNCountries()
    countries.value = res.countries || []
    if (selectedCountry.value) {
      await loadServers(selectedCountry.value)
    }
  } catch (err) {
    console.warn('Failed to load countries:', err)
  } finally {
    loadingCountries.value = false
  }
}

async function loadServers(country: string) {
  if (!country) return
  loadingServers.value = true
  try {
    const res = await desktopApi.getPVPNServers(country)
    servers.value = (res.servers || []).sort((a, b) => a.load - b.load)
    if (servers.value.length > 0 && !selectedServer.value) {
      selectedServer.value = servers.value[0].name
    }
  } catch (err) {
    console.warn('Failed to load servers:', err)
  } finally {
    loadingServers.value = false
  }
}

function onCountryChange(val: string) {
  selectedServer.value = ''
  loadServers(val)
}

function quickSelectCountry(code: string) {
  selectedCountry.value = code
  onCountryChange(code)
}

async function handleConnect() {
  connecting.value = true
  try {
    const server = connectMode.value === 'manual' ? selectedServer.value : ''
    const res = await desktopApi.connectPVPN(selectedCountry.value, server, autoEnsureOutbound.value)
    message.success(res.message || '已成功连接 Proton 节点')
    await loadData()
  } catch (err) {
    message.error('连接失败: ' + formatError(err))
  } finally {
    connecting.value = false
  }
}

async function handleDisconnect() {
  disconnecting.value = true
  try {
    await desktopApi.disconnectPVPN()
    message.success('已断开 Proton Wireproxy 代理')
    await loadData()
  } catch (err) {
    message.error('断开失败: ' + formatError(err))
  } finally {
    disconnecting.value = false
  }
}

async function handleEnsureOutbound() {
  ensuringOutbound.value = true
  try {
    await desktopApi.ensurePVPNOutbound()
    const p = status.value?.port || 1080
    message.success(`已在 Xray 中创建 proton SOCKS5 出站 (127.0.0.1:${p})`)
    await loadData()
  } catch (err) {
    message.error('创建出站失败: ' + formatError(err))
  } finally {
    ensuringOutbound.value = false
  }
}

function openPortModal() {
  customPort.value = status.value?.port || 1080
  showPortModal.value = true
}

async function handleSavePort() {
  if (!customPort.value || customPort.value < 1 || customPort.value > 65535) {
    message.error('请输入 1 - 65535 范围内的有效端口号')
    return
  }
  savingPort.value = true
  try {
    const res = await desktopApi.updatePVPNConfig(customPort.value)
    message.success(res.message || `已成功更新 SOCKS5 监听端口为 ${customPort.value}`)
    showPortModal.value = false
    await loadData()
  } catch (err) {
    message.error('更新端口失败: ' + formatError(err))
  } finally {
    savingPort.value = false
  }
}

async function handleLogin() {
  loggingIn.value = true
  try {
    const res = await desktopApi.loginPVPN()
    message.success(res.message || 'Proton 访客登录与凭据刷新成功')
    await loadData()
  } catch (err) {
    message.error('登录/刷新凭据失败: ' + formatError(err))
  } finally {
    loggingIn.value = false
  }
}

async function handleServiceAction(action: string) {
  serviceActionLoading.value = action
  try {
    const res = await desktopApi.pvpnServiceAction(action)
    message.success(res.message)
    await loadData()
  } catch (err) {
    message.error(`执行 ${action} 失败: ` + formatError(err))
  } finally {
    serviceActionLoading.value = ''
  }
}

async function handleInstall() {
  installing.value = true
  try {
    const res = await desktopApi.installPVPN()
    message.success(res.message)
    await loadData()
  } catch (err) {
    message.error('部署 wireproxy 引擎失败: ' + formatError(err))
  } finally {
    installing.value = false
  }
}

watch(
  () => conn.status.profile_id,
  (id) => {
    if (id && conn.status.connected) {
      loadData()
    }
  }
)

onMounted(loadData)
</script>
