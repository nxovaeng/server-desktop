<template>
  <n-layout has-sider class="h-screen">
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="208"
      show-trigger
      v-model:collapsed="collapsed"
    >
      <div class="h-14 flex items-center px-4 border-b border-subtle font-semibold text-base tracking-wide truncate">
        <ShieldCheck class="w-5 h-5 text-accent mr-2 shrink-0" />
        <span v-if="!collapsed">{{ conn.status.panel_title || 'Xray Panel' }}</span>
      </div>
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="20"
        :options="menuOptions"
        :value="activeKey"
        @update:value="onSelectMenu"
      />
    </n-layout-sider>

    <n-layout class="flex flex-col h-screen">
      <n-layout-header bordered class="h-14 px-4 flex items-center justify-between gap-3 shrink-0">
        <div class="flex items-center gap-2 min-w-0">
          <n-select
            v-model:value="selectedProfileId"
            :options="profileOptions"
            placeholder="选择服务器节点..."
            class="w-56"
            size="small"
          />
          <n-button
            v-if="!conn.status.connected || conn.status.profile_id !== selectedProfileId"
            type="primary"
            size="small"
            :loading="conn.connecting"
            :disabled="!selectedProfileId"
            @click="handleConnect"
          >
            连接
          </n-button>
          <n-button
            v-else
            type="warning"
            secondary
            size="small"
            @click="handleDisconnect"
          >
            断开
          </n-button>

          <n-button
            v-if="conn.status.connected"
            size="small"
            quaternary
            @click="handlePing"
            title="测速延迟"
          >
            <template #icon>
              <Activity class="w-4 h-4" />
            </template>
            {{ conn.status.latency_ms }} ms
          </n-button>
        </div>

        <div class="flex items-center gap-2 shrink-0">
          <n-button size="small" quaternary circle @click="theme.toggle()" :title="theme.isDark ? '切换到浅色' : '切换到深色'">
            <template #icon>
              <Sun v-if="theme.isDark" class="w-4 h-4" />
              <Moon v-else class="w-4 h-4" />
            </template>
          </n-button>
          <n-tag v-if="conn.status.connected" type="success" size="small" round>
            {{ conn.status.mode === 'ssh_tunnel' ? 'SSH 隧道' : '直连' }} · {{ conn.status.token_name || 'token' }}
          </n-tag>
          <n-tag v-else type="default" size="small" round>
            未连接
          </n-tag>
          <n-tag v-if="conn.status.server_version" size="small" :bordered="false">
            v{{ conn.status.server_version }}
          </n-tag>
        </div>
      </n-layout-header>

      <n-layout-content class="p-4 overflow-auto flex-1">
        <n-alert
          v-if="conn.status.warning"
          type="warning"
          class="mb-4"
          closable
        >
          {{ conn.status.warning }}
        </n-alert>

        <n-alert
          v-if="!conn.status.connected && route.name !== 'profiles'"
          type="info"
          title="尚未连接到远程 xray-panel-server"
          class="mb-4"
        >
          请先在顶部下拉框选择节点并点击「连接」，或前往
          <n-button text type="primary" @click="router.push('/profiles')">节点连接配置</n-button>
          添加远程服务器。
        </n-alert>

        <div class="mx-auto w-full max-w-[1560px]">
          <router-view :key="conn.status.profile_id || 'disconnected'" />
        </div>
      </n-layout-content>
    </n-layout>
  </n-layout>

  <n-modal
    :show="Boolean(conn.tofuChallenge)"
    preset="dialog"
    type="warning"
    title="首次连接 SSH 主机指纹确认 (TOFU)"
    positive-text="信任并保存指纹"
    negative-text="取消"
    @positive-click="handleConfirmTOFU"
    @negative-click="conn.cancelTOFU()"
    @close="conn.cancelTOFU()"
  >
    <div v-if="conn.tofuChallenge" class="space-y-2 text-sm">
      <p>
        首次连接节点 <strong>{{ conn.tofuChallenge.profileName }}</strong>，请核对远程服务器 SSH 公钥 SHA-256 指纹：
      </p>
      <div class="p-2.5 surface-code mono text-xs break-all select-all">
        {{ conn.tofuChallenge.fingerprint }}
      </div>
      <p class="text-xs text-muted">
        确认保存后，后续若主机密钥发生变更将自动拦截并告警，防止中间人攻击。
      </p>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, type MenuOption } from 'naive-ui'
import {
  Activity,
  ArrowDownUp,
  FileText,
  Globe,
  KeyRound,
  LayoutDashboard,
  Moon,
  Network,
  Route as RouteIcon,
  Server,
  Settings,
  Shield,
  ShieldCheck,
  Sun,
  Users
} from 'lucide-vue-next'
import { useConnectionStore } from '@/stores/connection'
import { useThemeStore } from '@/stores/theme'
import { formatError } from '@/api/client'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const conn = useConnectionStore()
const theme = useThemeStore()

const collapsed = ref(false)
const selectedProfileId = ref('')

const renderIcon = (icon: unknown) => () => h(icon as never, { class: 'w-4 h-4' })

const menuOptions: MenuOption[] = [
  { label: '节点连接', key: '/profiles', icon: renderIcon(Server) },
  { label: '仪表盘', key: '/dashboard', icon: renderIcon(LayoutDashboard) },
  { label: '用户管理', key: '/users', icon: renderIcon(Users) },
  { label: '入站配置', key: '/inbounds', icon: renderIcon(ArrowDownUp) },
  { label: '出站配置', key: '/outbounds', icon: renderIcon(Network) },
  { label: '路由规则', key: '/routing', icon: renderIcon(RouteIcon) },
  { label: '域名证书', key: '/domains', icon: renderIcon(Globe) },
  { label: 'WireGuard', key: '/wireguard', icon: renderIcon(Shield) },
  { label: '应用配置', key: '/settings', icon: renderIcon(Settings) },
  { label: '系统日志', key: '/logs', icon: renderIcon(FileText) },
  { label: 'API 令牌', key: '/tokens', icon: renderIcon(KeyRound) }
]

const activeKey = computed(() => route.path)

const profileOptions = computed(() =>
  conn.profiles.map((p) => ({
    label: `${p.name} (${p.mode === 'ssh_tunnel' ? p.ssh_host : p.direct_url})`,
    value: p.id
  }))
)

watch(
  () => conn.activeProfileId,
  (val) => {
    if (val) selectedProfileId.value = val
  },
  { immediate: true }
)

// Token 失效（401）统一提示：任一 desktopApi 调用返回 401 时，
// connection store 的 handleUnauthorized 会置未连接并递增 authErrorSeq。
watch(
  () => conn.authErrorSeq,
  () => {
    if (conn.authError) {
      message.error(conn.authError)
      conn.clearAuthError()
    }
  }
)

function onSelectMenu(key: string) {
  router.push(key)
}

async function handleConnect() {
  if (!selectedProfileId.value) return
  try {
    const st = await conn.connect(selectedProfileId.value, false)
    if (st) {
      message.success(`已连接到 ${st.profile_name} (${st.latency_ms} ms)`)
      if (route.path === '/profiles') {
        router.push('/dashboard')
      }
    }
  } catch (err) {
    message.error('连接失败: ' + formatError(err))
  }
}

async function handleConfirmTOFU() {
  try {
    const st = await conn.confirmTOFU()
    if (st) {
      message.success(`已信任指纹并连接到 ${st.profile_name}`)
    }
  } catch (err) {
    message.error('连接失败: ' + formatError(err))
  }
}

async function handleDisconnect() {
  await conn.disconnect()
  message.info('已断开服务器连接')
}

async function handlePing() {
  try {
    const ms = await conn.testLatency()
    message.success(`当前连接延迟: ${ms} ms`)
  } catch (err) {
    message.error('测速失败: ' + formatError(err))
  }
}

onMounted(async () => {
  try {
    await conn.loadProfiles()
    await conn.syncStatus()
    if (conn.profiles.length === 0) {
      router.push('/profiles')
    }
  } catch {
    // Allow running outside Wails during static inspection
  }
})
</script>
