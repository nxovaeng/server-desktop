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
          </div>
        </n-card>

        <n-card title="Xray config.json 实时预览" size="small">
          <template #header-extra>
            <div class="flex gap-2">
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
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { copyToClipboard, desktopApi, formatBytes, formatError } from '@/api/client'
import type { DirectDomainStrategy, XrayLogLevel, XrayStatusResponse } from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const loading = ref(false)
const saving = ref(false)
const applying = ref(false)
const hotReloading = ref(false)
const restartingXray = ref(false)
const loadingConfig = ref(false)

const xrayStatus = ref<XrayStatusResponse | null>(null)
const xrayRaw = ref('')
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
    const [settings, st, cfg] = await Promise.all([
      desktopApi.getSettings(),
      desktopApi.getXrayStatus(),
      desktopApi.getXrayConfig()
    ])
    Object.assign(form, {
      panel_title: settings.panel_title,
      sub_domain: settings.sub_domain,
      sub_path: settings.sub_path,
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
