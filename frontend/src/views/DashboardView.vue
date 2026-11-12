<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-semibold">系统仪表盘</h2>
      <n-button size="small" :loading="loading" :disabled="!conn.status.connected" @click="load">
        刷新数据
      </n-button>
    </div>

    <template v-if="data">
      <n-grid :cols="4" :x-gap="12" :y-gap="12" responsive="screen" item-responsive>
        <n-grid-item span="4 m:2 l:1">
          <n-card size="small" title="代理用户">
            <n-statistic :value="data.users.active">
              <template #suffix>/ {{ data.users.total }}</template>
            </n-statistic>
          </n-card>
        </n-grid-item>
        <n-grid-item span="4 m:2 l:1">
          <n-card size="small" title="活跃入站 / 出站">
            <div class="text-2xl font-semibold">
              {{ data.inbounds.active }}/{{ data.inbounds.total }}
              <span class="text-muted text-sm mx-1">·</span>
              {{ data.outbounds.active }}/{{ data.outbounds.total }}
            </div>
          </n-card>
        </n-grid-item>
        <n-grid-item span="4 m:2 l:1">
          <n-card size="small" title="本月主机总流量">
            <div class="text-2xl font-semibold text-accent">
              {{ formatBytes(data.traffic.monthly_net_total_bytes) }}
            </div>
            <div class="text-xs text-muted mt-1">
              ↑ {{ formatBytes(data.traffic.monthly_net_sent_bytes) }} · ↓ {{ formatBytes(data.traffic.monthly_net_recv_bytes) }}
            </div>
          </n-card>
        </n-grid-item>
        <n-grid-item span="4 m:2 l:1">
          <n-card size="small" title="服务状态">
            <div class="flex flex-wrap gap-2 mt-1">
              <n-tag :type="data.services.xray_active ? 'success' : 'error'" size="small">
                Xray: {{ data.services.xray_active ? '运行中' : data.services.xray_status }}
              </n-tag>
              <n-tag
                :type="data.services.wireguard_active ? 'success' : 'default'"
                size="small"
              >
                WG: {{ data.services.wireguard_active ? `${data.services.wireguard_online_peers}/${data.services.wireguard_total_peers} 在线` : '未运行' }}
              </n-tag>
            </div>
          </n-card>
        </n-grid-item>
      </n-grid>

      <n-grid :cols="3" :x-gap="12" :y-gap="12" responsive="screen" item-responsive>
        <n-grid-item span="3 m:1">
          <n-card size="small" title="CPU 使用率">
            <n-progress
              type="line"
              :percentage="Number(data.system.cpu_usage_percent.toFixed(1))"
              indicator-placement="inside"
            />
            <div class="text-xs text-muted mt-2 truncate">
              {{ data.system.cpu_cores }} 核 · {{ data.system.cpu_model || 'CPU' }}
            </div>
          </n-card>
        </n-grid-item>
        <n-grid-item span="3 m:1">
          <n-card size="small" title="内存占用">
            <n-progress
              type="line"
              :percentage="Number(data.system.mem_percent.toFixed(1))"
              status="info"
              indicator-placement="inside"
            />
            <div class="text-xs text-muted mt-2">
              {{ formatBytes(data.system.mem_used_bytes) }} / {{ formatBytes(data.system.mem_total_bytes) }}
            </div>
          </n-card>
        </n-grid-item>
        <n-grid-item span="3 m:1">
          <n-card size="small" title="磁盘占用">
            <n-progress
              type="line"
              :percentage="Number(data.system.disk_percent.toFixed(1))"
              status="warning"
              indicator-placement="inside"
            />
            <div class="text-xs text-muted mt-2">
              {{ formatBytes(data.system.disk_used_bytes) }} / {{ formatBytes(data.system.disk_total_bytes) }}
            </div>
          </n-card>
        </n-grid-item>
      </n-grid>

      <n-card size="small" title="主机信息">
        <n-descriptions :column="3" size="small">
          <n-descriptions-item label="主机名">{{ data.system.hostname }}</n-descriptions-item>
          <n-descriptions-item label="平台">{{ data.system.platform || data.system.os }}</n-descriptions-item>
          <n-descriptions-item label="运行时间">{{ formatUptime(data.system.uptime_seconds) }}</n-descriptions-item>
          <n-descriptions-item label="用户累计流量">{{ formatBytes(data.traffic.users_traffic_used_bytes) }}</n-descriptions-item>
          <n-descriptions-item label="域名 / 路由">{{ data.domains.active }} / {{ data.routing_rules.active }}</n-descriptions-item>
          <n-descriptions-item label="Go Runtime">{{ data.system.go_version }}</n-descriptions-item>
        </n-descriptions>
      </n-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { desktopApi, formatBytes, formatError, formatUptime } from '@/api/client'
import type { DashboardResponse } from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()
const loading = ref(false)
const data = ref<DashboardResponse | null>(null)

async function load() {
  if (!conn.status.connected || loading.value) return
  loading.value = true
  try {
    data.value = await desktopApi.getDashboard()
  } catch (err) {
    message.error('加载仪表盘失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

// 30 秒自动刷新：只在已连接时请求；切页/卸载时清掉定时器
let refreshTimer: ReturnType<typeof setInterval> | null = null

function startAutoRefresh() {
  stopAutoRefresh()
  refreshTimer = setInterval(() => {
    if (conn.status.connected) void load()
  }, 30000)
}

function stopAutoRefresh() {
  if (refreshTimer !== null) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

onMounted(() => {
  void load()
  startAutoRefresh()
})
onUnmounted(stopAutoRefresh)

// 连接建立后（例如从 Profiles 页连上再切回来）立即刷一次
watch(
  () => conn.status.connected,
  (connected) => {
    if (connected) void load()
  }
)
</script>
