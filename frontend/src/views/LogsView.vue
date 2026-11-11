<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between flex-wrap gap-2">
      <h2 class="text-xl font-semibold">服务日志</h2>
      <div class="flex items-center gap-2 flex-wrap">
        <n-select
          v-model:value="service"
          size="small"
          class="w-40"
          :options="[
            { label: 'Xray Core (xray)', value: 'xray' },
            { label: '控制面 (panel)', value: 'panel' },
            { label: 'WireGuard (wg0)', value: 'wireguard' },
            { label: 'Nginx', value: 'nginx' }
          ]"
          @update:value="load"
        />
        <n-select
          v-model:value="lines"
          size="small"
          class="w-28"
          :options="[
            { label: '100 行', value: 100 },
            { label: '300 行', value: 300 },
            { label: '500 行', value: 500 },
            { label: '1000 行', value: 1000 }
          ]"
          @update:value="load"
        />
        <n-select
          v-model:value="levelFilter"
          size="small"
          class="w-28"
          :options="[
            { label: '全部级别', value: 'ALL' },
            { label: 'ERROR', value: 'ERROR' },
            { label: 'WARN', value: 'WARN' },
            { label: 'INFO', value: 'INFO' }
          ]"
        />
        <n-input
          v-model:value="keyword"
          size="small"
          clearable
          class="w-48"
          placeholder="关键字过滤..."
        />
        <div class="flex items-center gap-1.5 px-2 text-xs text-gray-300">
          <n-switch v-model:value="autoRefresh" size="small" />
          <span>自动刷新 (3s)</span>
        </div>
        <n-button size="small" :disabled="!conn.status.connected" :loading="loading" @click="load">
          刷新
        </n-button>
        <n-button size="small" :disabled="!filteredText" @click="copyLogs">复制</n-button>
        <n-button size="small" type="primary" secondary :disabled="!filteredText" @click="exportLogs">
          导出日志
        </n-button>
      </div>
    </div>

    <n-card size="small">
      <div class="text-xs text-gray-400 mb-2 flex items-center justify-between">
        <span>
          服务：<span class="mono text-gray-200">{{ service }}</span> · 显示
          <span class="mono text-gray-200">{{ filteredLines.length }}</span> /
          <span class="mono">{{ totalLines }}</span> 行
        </span>
        <span v-if="lastUpdated" class="mono">最近拉取：{{ lastUpdated }}</span>
      </div>
      <pre
        class="p-3 rounded bg-black/50 border border-white/10 mono text-xs overflow-auto h-[520px] whitespace-pre-wrap break-all leading-relaxed"
      >{{ filteredText || (loading ? '正在拉取日志...' : '暂无匹配日志内容') }}</pre>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { copyToClipboard, desktopApi, formatError } from '@/api/client'
import { useConnectionStore } from '@/stores/connection'

type LogService = 'xray' | 'panel' | 'wireguard' | 'nginx'
type LevelFilter = 'ALL' | 'ERROR' | 'WARN' | 'INFO'

const conn = useConnectionStore()
const message = useMessage()

const service = ref<LogService>('xray')
const lines = ref(300)
const levelFilter = ref<LevelFilter>('ALL')
const keyword = ref('')
const rawContent = ref('')
const lastUpdated = ref('')
const loading = ref(false)
const autoRefresh = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

const allLines = computed(() => {
  if (!rawContent.value) return []
  return rawContent.value.split('\n')
})

const totalLines = computed(() => allLines.value.length)

const filteredLines = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  const lvl = levelFilter.value
  return allLines.value.filter((line) => {
    if (lvl !== 'ALL') {
      const upper = line.toUpperCase()
      if (lvl === 'WARN') {
        if (!upper.includes('WARN') && !upper.includes('[WARNING]')) return false
      } else if (!upper.includes(lvl)) {
        return false
      }
    }
    if (kw && !line.toLowerCase().includes(kw)) {
      return false
    }
    return true
  })
})

const filteredText = computed(() => filteredLines.value.join('\n'))

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const res = await desktopApi.getLogs(service.value, lines.value)
    rawContent.value = res.content || ''
    lastUpdated.value = new Date().toLocaleTimeString()
  } catch (err) {
    message.error('读取日志失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

async function copyLogs() {
  await copyToClipboard(filteredText.value)
  message.success('日志已复制到剪贴板')
}

async function exportLogs() {
  try {
    const filename = `${service.value}-${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.log`
    const saved = await desktopApi.saveTextFile(filename, filteredText.value)
    if (saved) {
      message.success(`已保存至 ${saved}`)
    }
  } catch (err) {
    message.error('导出日志失败: ' + formatError(err))
  }
}

function stopTimer() {
  if (timer !== null) {
    clearInterval(timer)
    timer = null
  }
}

watch(autoRefresh, (enabled) => {
  stopTimer()
  if (enabled) {
    load()
    timer = setInterval(() => {
      if (conn.status.connected && !loading.value) {
        load()
      }
    }, 3000)
  }
})

watch(
  () => conn.status.profile_id,
  (id) => {
    if (id && conn.status.connected) {
      load()
    }
  }
)

onMounted(load)
onBeforeUnmount(stopTimer)
</script>
