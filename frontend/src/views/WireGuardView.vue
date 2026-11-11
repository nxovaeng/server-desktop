<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between flex-wrap gap-2">
      <h2 class="text-xl font-semibold">WireGuard 内核组网</h2>
      <div class="flex gap-2 flex-wrap">
        <n-button size="small" @click="load" :disabled="!conn.status.connected" :loading="loading">刷新</n-button>
        <n-button size="small" secondary :disabled="!conn.status.connected" @click="openServerModal">
          服务端配置
        </n-button>
        <n-button size="small" type="primary" :disabled="!conn.status.connected" @click="openCreatePeer">
          新增客户端 Peer
        </n-button>
      </div>
    </div>

    <n-alert
      v-if="wgStatus?.port_conflict"
      type="error"
      title="端口冲突警告"
    >
      WireGuard 监听端口 {{ wgStatus.listen_port }} 与 Xray 入站 [{{ wgStatus.conflict_inbound }}] 发生 UDP 端口冲突，请修改服务端监听端口或停用冲突入站。
    </n-alert>

    <n-alert
      v-if="wgStatus && wgStatus.is_linux && (!wgStatus.ipv4_forwarding)"
      type="warning"
      title="内核 IPv4 转发未开启"
    >
      检测到远程主机尚未开启 IPv4 Forwarding，客户端接入后可能无法转发流量；可点击下方「初始化环境 / 开启转发」修复。
    </n-alert>

    <n-card v-if="wgStatus" size="small" title="服务状态与运维">
      <div class="space-y-3">
        <div class="flex flex-wrap gap-2 items-center">
          <n-tag size="small" :type="wgStatus.is_installed ? 'success' : 'error'">
            {{ wgStatus.is_installed ? 'wg / wg-quick 已安装' : '未安装 WireGuard 工具' }}
          </n-tag>
          <n-tag size="small" :type="wgStatus.is_active ? 'success' : 'warning'">
            {{ wgStatus.is_active ? `运行中 (${wgStatus.interface_name})` : `已停止 (${wgStatus.interface_name})` }}
          </n-tag>
          <n-tag size="small" :type="wgStatus.is_enabled ? 'info' : 'default'">
            开机自启：{{ wgStatus.is_enabled ? '已开启' : '未开启' }}
          </n-tag>
          <n-tag size="small" :type="wgStatus.ipv4_forwarding ? 'success' : 'warning'">
            IPv4 转发：{{ wgStatus.ipv4_forwarding ? '已开启' : '未开启' }}
          </n-tag>
          <n-tag size="small" :type="wgStatus.ipv6_forwarding ? 'success' : 'default'">
            IPv6 转发：{{ wgStatus.ipv6_forwarding ? '已开启' : '未开启' }}
          </n-tag>
          <n-tag size="small" :type="wgStatus.enable_nat ? 'info' : 'default'">
            NAT：{{ wgStatus.enable_nat ? '已启用' : '未启用' }}
          </n-tag>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 gap-3 text-xs text-gray-300">
          <div>
            <span class="text-gray-400">监听端口：</span>
            <span class="mono">{{ wgStatus.listen_port }}/udp</span>
          </div>
          <div>
            <span class="text-gray-400">内网网段：</span>
            <span class="mono">{{ wgStatus.address }}</span>
          </div>
          <div>
            <span class="text-gray-400">在线 Peer：</span>
            <span>{{ wgStatus.online_peers }} / {{ wgStatus.active_peers }} 启用 (共 {{ wgStatus.total_peers }})</span>
          </div>
          <div class="truncate" :title="wgStatus.public_key">
            <span class="text-gray-400">服务端公钥：</span>
            <span class="mono">{{ wgStatus.public_key || '-' }}</span>
          </div>
        </div>

        <div class="flex flex-wrap gap-2 pt-1">
          <n-button
            size="tiny"
            type="info"
            secondary
            :loading="settingUpEnv"
            @click="setupEnv"
          >
            初始化环境 / 安装与开启转发
          </n-button>
          <n-button
            size="tiny"
            type="success"
            secondary
            :loading="actionLoading === 'start'"
            @click="runServiceAction('start')"
          >
            启动
          </n-button>
          <n-button
            size="tiny"
            type="warning"
            secondary
            :loading="actionLoading === 'restart'"
            @click="runServiceAction('restart')"
          >
            重启
          </n-button>
          <n-button
            size="tiny"
            type="error"
            secondary
            :loading="actionLoading === 'stop'"
            @click="runServiceAction('stop')"
          >
            停止
          </n-button>
          <n-button
            v-if="!wgStatus.is_enabled"
            size="tiny"
            :loading="actionLoading === 'enable'"
            @click="runServiceAction('enable')"
          >
            开启自启
          </n-button>
          <n-button
            v-else
            size="tiny"
            :loading="actionLoading === 'disable'"
            @click="runServiceAction('disable')"
          >
            取消自启
          </n-button>
        </div>
      </div>
    </n-card>

    <n-data-table :columns="columns" :data="peers" :loading="loading" size="small" />

    <n-modal v-model:show="showServerModal" preset="card" title="WireGuard 服务端接口配置" class="w-[620px]">
      <n-form label-placement="left" label-width="125" size="small">
        <div class="grid grid-cols-2 gap-3">
          <n-form-item label="接口名称">
            <n-input v-model:value="serverForm.interface_name" class="mono" placeholder="wg0" />
          </n-form-item>
          <n-form-item label="UDP 监听端口">
            <n-input-number v-model:value="serverForm.listen_port" :min="1024" :max="65535" class="w-full" />
          </n-form-item>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <n-form-item label="服务端虚拟网段">
            <n-input v-model:value="serverForm.address" class="mono" placeholder="10.66.66.1/24" />
          </n-form-item>
          <n-form-item label="MTU">
            <n-input-number v-model:value="serverForm.mtu" :min="576" :max="1500" class="w-full" />
          </n-form-item>
        </div>
        <n-form-item label="服务端私钥">
          <div class="flex gap-2 w-full">
            <n-input
              v-model:value="serverForm.private_key"
              type="password"
              show-password-on="click"
              class="mono"
              :placeholder="serverHasPrivKey ? '已保存（留空保持不变）' : '点击右侧生成'"
            />
            <n-button size="small" @click="genServerKeys">生成新密钥</n-button>
          </div>
        </n-form-item>
        <n-form-item label="服务端公钥">
          <n-input v-model:value="serverForm.public_key" class="mono" readonly />
        </n-form-item>
        <n-form-item label="自动 NAT 伪装">
          <n-switch v-model:value="serverForm.enable_nat" />
          <span class="text-xs text-gray-400 ml-2">开启后为流量转发配置 iptables MASQUERADE</span>
        </n-form-item>
        <n-form-item label="PostUp 脚本">
          <n-input v-model:value="serverForm.post_up" class="mono" placeholder="留空按 NAT 策略自动填充" />
        </n-form-item>
        <n-form-item label="PostDown 脚本">
          <n-input v-model:value="serverForm.post_down" class="mono" placeholder="留空按 NAT 策略自动填充" />
        </n-form-item>
        <n-form-item label="启用服务配置">
          <n-switch v-model:value="serverForm.enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button size="small" @click="showServerModal = false">取消</n-button>
          <n-button size="small" type="primary" :loading="savingServer" @click="saveServer">保存并同步</n-button>
        </div>
      </template>
    </n-modal>

    <n-modal v-model:show="showPeerModal" preset="card" :title="editingPeerId ? '编辑 WireGuard Peer' : '新增 WireGuard Peer'" class="w-[620px]">
      <n-form label-placement="left" label-width="125" size="small">
        <n-form-item label="客户端名称" required>
          <n-input v-model:value="peerForm.name" placeholder="例如：laptop-home" />
        </n-form-item>
        <n-form-item label="虚拟 IP (AllowedIPs)">
          <div class="flex gap-2 w-full">
            <n-input v-model:value="peerForm.allowed_ips" class="mono" placeholder="10.66.66.2/32" />
            <n-button v-if="nextAvailableIP" size="small" @click="peerForm.allowed_ips = nextAvailableIP">
              填入建议 IP
            </n-button>
          </div>
        </n-form-item>
        <n-form-item label="客户端私钥">
          <div class="flex gap-2 w-full">
            <n-input
              v-model:value="peerForm.private_key"
              type="password"
              show-password-on="click"
              class="mono"
              :placeholder="peerHasPrivKey ? '已保存（留空保持不变）' : '可选；仅填写公钥时可留空'"
            />
            <n-button size="small" @click="genPeerKeys">本地生成密钥对</n-button>
          </div>
        </n-form-item>
        <n-form-item label="客户端公钥" required>
          <div class="flex gap-2 w-full">
            <n-input v-model:value="peerForm.public_key" class="mono" placeholder="Curve25519 公钥" />
            <n-button v-if="peerForm.private_key" size="small" @click="derivePeerPubKey">从私钥推导</n-button>
          </div>
        </n-form-item>
        <n-form-item label="PresharedKey">
          <n-input
            v-model:value="peerForm.preshared_key"
            type="password"
            show-password-on="click"
            class="mono"
            :placeholder="peerHasPSK ? '已保存（留空保持不变）' : '可选，额外对称密钥'"
          />
        </n-form-item>
        <div class="grid grid-cols-2 gap-3">
          <n-form-item label="Endpoint (可选)">
            <n-input v-model:value="peerForm.endpoint" class="mono" placeholder="覆盖服务器地址:端口" />
          </n-form-item>
          <n-form-item label="Keepalive (秒)">
            <n-input-number v-model:value="peerForm.persistent_keepalive" :min="0" :max="600" class="w-full" />
          </n-form-item>
        </div>
        <n-form-item label="启用 Peer">
          <n-switch v-model:value="peerForm.enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button size="small" @click="showPeerModal = false">取消</n-button>
          <n-button size="small" type="primary" :loading="savingPeer" @click="savePeer">保存并下发</n-button>
        </div>
      </template>
    </n-modal>

    <n-modal v-model:show="showExportModal" preset="card" :title="`导出 Peer 配置 - ${exportCfg?.peer_name || ''}`" class="w-[660px]">
      <div v-if="exportCfg" class="space-y-3">
        <div class="text-xs text-gray-400 flex flex-wrap gap-4">
          <span>Server Host：<span class="mono text-gray-200">{{ exportCfg.server_host || '-' }}</span></span>
          <span>Endpoint：<span class="mono text-gray-200">{{ exportCfg.endpoint || '-' }}</span></span>
        </div>

        <n-alert v-if="!exportCfg.has_private_key" type="warning" size="small">
          该 Peer 创建时未在服务端托管私钥，导出的配置中私钥为占位符，请替换为客户端本地保存的私钥后使用。
        </n-alert>

        <n-tabs type="line" animated size="small">
          <n-tab-pane name="wg" tab="wg0.conf (标准 WireGuard)">
            <div class="space-y-2">
              <pre class="p-3 rounded bg-black/40 border border-white/10 mono text-xs overflow-auto max-h-72">{{ exportCfg.wg_conf }}</pre>
              <div class="flex justify-end gap-2">
                <n-button size="small" @click="exportToFile(`${exportCfg.peer_name}.conf`, exportCfg.wg_conf)">
                  保存 .conf 文件
                </n-button>
                <n-button size="small" type="primary" @click="copyConf(exportCfg.wg_conf)">
                  复制 wg0.conf
                </n-button>
              </div>
            </div>
          </n-tab-pane>
          <n-tab-pane name="xray" tab="Xray Outbound JSON">
            <div class="space-y-2">
              <pre class="p-3 rounded bg-black/40 border border-white/10 mono text-xs overflow-auto max-h-72">{{ exportCfg.xray_json }}</pre>
              <div class="flex justify-end gap-2">
                <n-button size="small" @click="exportToFile(`${exportCfg.peer_name}-xray.json`, exportCfg.xray_json)">
                  保存 .json 文件
                </n-button>
                <n-button size="small" type="primary" @click="copyConf(exportCfg.xray_json)">
                  复制 Xray JSON
                </n-button>
              </div>
            </div>
          </n-tab-pane>
        </n-tabs>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref, watch } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useMessage, type DataTableColumns } from 'naive-ui'
import { copyToClipboard, desktopApi, formatError } from '@/api/client'
import type {
  WGPeerDTO,
  WGPeerExportConfigResponse,
  WGServerConfigDTO,
  WGServiceStatus
} from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const wgStatus = ref<WGServiceStatus | null>(null)
const serverCfg = ref<WGServerConfigDTO | null>(null)
const peers = ref<WGPeerDTO[]>([])
const nextAvailableIP = ref('')

const loading = ref(false)
const settingUpEnv = ref(false)
const actionLoading = ref('')
const savingServer = ref(false)
const savingPeer = ref(false)

const showServerModal = ref(false)
const serverHasPrivKey = ref(false)
const serverForm = reactive({
  interface_name: 'wg0',
  listen_port: 51820,
  address: '10.66.66.1/24',
  mtu: 1420,
  private_key: '',
  public_key: '',
  enable_nat: true,
  post_up: '',
  post_down: '',
  enabled: true
})

const showPeerModal = ref(false)
const editingPeerId = ref<number | null>(null)
const peerHasPrivKey = ref(false)
const peerHasPSK = ref(false)
const peerForm = reactive({
  name: '',
  allowed_ips: '',
  private_key: '',
  public_key: '',
  preshared_key: '',
  endpoint: '',
  persistent_keepalive: 25,
  enabled: true
})

const showExportModal = ref(false)
const exportCfg = ref<WGPeerExportConfigResponse | null>(null)

const columns: DataTableColumns<WGPeerDTO> = [
  { title: '名称', key: 'name', width: 130 },
  {
    title: '虚拟 IP',
    key: 'allowed_ips',
    width: 140,
    render: (p) => h('span', { class: 'mono text-xs' }, p.allowed_ips)
  },
  {
    title: '公钥',
    key: 'public_key',
    width: 170,
    ellipsis: { tooltip: true },
    render: (p) => h('span', { class: 'mono text-xs' }, p.public_key)
  },
  {
    title: '握手 / 对端',
    key: 'runtime',
    render: (p) => {
      const rt = p.runtime
      return h(NSpace, { size: 6, align: 'center' }, () => [
        h(NTag, { size: 'tiny', type: rt?.is_online ? 'success' : 'default' }, () =>
          rt?.is_online ? '在线' : '离线'
        ),
        h('span', { class: 'text-xs text-gray-300' }, rt?.handshake_ago || '从未握手'),
        rt?.endpoint ? h('span', { class: 'mono text-xs text-gray-400' }, `(${rt.endpoint})`) : null
      ])
    }
  },
  {
    title: '流量 (Rx / Tx)',
    key: 'traffic',
    width: 160,
    render: (p) =>
      h(
        'span',
        { class: 'mono text-xs' },
        `${p.runtime?.transfer_rx_human || '0 B'} / ${p.runtime?.transfer_tx_human || '0 B'}`
      )
  },
  {
    title: '状态',
    key: 'enabled',
    width: 75,
    render: (p) => h(NTag, { size: 'small', type: p.enabled ? 'success' : 'default' }, () => (p.enabled ? '启用' : '停用'))
  },
  {
    title: '操作',
    key: 'actions',
    width: 260,
    render: (p) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'tiny', type: 'info', secondary: true, onClick: () => openExport(p.id) }, () => '导出配置'),
        h(NButton, { size: 'tiny', onClick: () => openEditPeer(p) }, () => '编辑'),
        h(NButton, { size: 'tiny', onClick: () => togglePeer(p.id) }, () => (p.enabled ? '停用' : '启用')),
        h(
          NPopconfirm,
          { onPositiveClick: () => deletePeer(p.id) },
          {
            trigger: () => h(NButton, { size: 'tiny', type: 'error', quaternary: true }, () => '删除'),
            default: () => `确认删除 WireGuard 客户端 ${p.name}？`
          }
        )
      ])
  }
]

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const [st, srv, peerRes] = await Promise.all([
      desktopApi.getWGStatus(),
      desktopApi.getWGServer(),
      desktopApi.listWGPeers()
    ])
    wgStatus.value = st
    serverCfg.value = srv
    peers.value = peerRes.peers
    nextAvailableIP.value = peerRes.next_available_ip
  } catch (err) {
    message.error('加载 WireGuard 状态失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

async function setupEnv() {
  settingUpEnv.value = true
  try {
    const res = await desktopApi.setupWGEnv()
    message.success(res.message || 'WireGuard 环境初始化完成')
    await load()
  } catch (err) {
    message.error('初始化环境失败: ' + formatError(err))
  } finally {
    settingUpEnv.value = false
  }
}

async function runServiceAction(action: 'start' | 'stop' | 'restart' | 'enable' | 'disable') {
  actionLoading.value = action
  try {
    const res = await desktopApi.wgServiceAction(action)
    wgStatus.value = res.status
    message.success(`操作 ${action} 已执行`)
  } catch (err) {
    message.error(`执行 ${action} 失败: ` + formatError(err))
  } finally {
    actionLoading.value = ''
  }
}

function openServerModal() {
  const s = serverCfg.value
  serverHasPrivKey.value = s?.has_private_key ?? false
  Object.assign(serverForm, {
    interface_name: s?.interface_name || 'wg0',
    listen_port: s?.listen_port || 51820,
    address: s?.address || '10.66.66.1/24',
    mtu: s?.mtu || 1420,
    private_key: '',
    public_key: s?.public_key || '',
    enable_nat: s?.enable_nat ?? true,
    post_up: s?.post_up || '',
    post_down: s?.post_down || '',
    enabled: s?.enabled ?? true
  })
  showServerModal.value = true
}

async function genServerKeys() {
  try {
    const kp = await desktopApi.generateLocalWGKeys()
    serverForm.private_key = kp.private_key
    serverForm.public_key = kp.public_key
  } catch (err) {
    message.error('生成密钥失败: ' + formatError(err))
  }
}

async function saveServer() {
  savingServer.value = true
  try {
    const updated = await desktopApi.updateWGServer({
      interface_name: serverForm.interface_name.trim(),
      listen_port: serverForm.listen_port,
      address: serverForm.address.trim(),
      mtu: serverForm.mtu,
      private_key: serverForm.private_key.trim(),
      enable_nat: serverForm.enable_nat,
      post_up: serverForm.post_up.trim(),
      post_down: serverForm.post_down.trim(),
      enabled: serverForm.enabled
    })
    serverCfg.value = updated
    showServerModal.value = false
    message.success('服务端配置已保存并同步')
    await load()
  } catch (err) {
    message.error('保存服务端配置失败: ' + formatError(err))
  } finally {
    savingServer.value = false
  }
}

function openCreatePeer() {
  editingPeerId.value = null
  peerHasPrivKey.value = false
  peerHasPSK.value = false
  Object.assign(peerForm, {
    name: '',
    allowed_ips: nextAvailableIP.value || '',
    private_key: '',
    public_key: '',
    preshared_key: '',
    endpoint: '',
    persistent_keepalive: 25,
    enabled: true
  })
  showPeerModal.value = true
}

function openEditPeer(p: WGPeerDTO) {
  editingPeerId.value = p.id
  peerHasPrivKey.value = p.has_private_key
  peerHasPSK.value = p.has_preshared_key
  Object.assign(peerForm, {
    name: p.name,
    allowed_ips: p.allowed_ips,
    private_key: '',
    public_key: p.public_key,
    preshared_key: '',
    endpoint: p.endpoint,
    persistent_keepalive: p.persistent_keepalive || 25,
    enabled: p.enabled
  })
  showPeerModal.value = true
}

async function genPeerKeys() {
  try {
    const kp = await desktopApi.generateLocalWGKeys()
    peerForm.private_key = kp.private_key
    peerForm.public_key = kp.public_key
  } catch (err) {
    message.error('生成密钥对失败: ' + formatError(err))
  }
}

async function derivePeerPubKey() {
  try {
    peerForm.public_key = await desktopApi.deriveLocalWGPublicKey(peerForm.private_key.trim())
  } catch (err) {
    message.error('推导公钥失败: ' + formatError(err))
  }
}

async function savePeer() {
  if (!peerForm.name.trim()) {
    message.warning('请填写客户端名称')
    return
  }
  if (!peerForm.public_key.trim() && !peerForm.private_key.trim() && !editingPeerId.value) {
    message.warning('请提供客户端公钥或点击本地生成密钥对')
    return
  }
  savingPeer.value = true
  try {
    const payload = {
      name: peerForm.name.trim(),
      public_key: peerForm.public_key.trim(),
      private_key: peerForm.private_key.trim(),
      preshared_key: peerForm.preshared_key.trim(),
      allowed_ips: peerForm.allowed_ips.trim(),
      endpoint: peerForm.endpoint.trim(),
      persistent_keepalive: peerForm.persistent_keepalive,
      enabled: peerForm.enabled
    }
    if (editingPeerId.value !== null) {
      await desktopApi.updateWGPeer(editingPeerId.value, payload)
    } else {
      await desktopApi.createWGPeer(payload)
    }
    showPeerModal.value = false
    message.success('客户端 Peer 已保存并同步')
    await load()
  } catch (err) {
    message.error('保存失败: ' + formatError(err))
  } finally {
    savingPeer.value = false
  }
}

async function togglePeer(id: number) {
  try {
    await desktopApi.toggleWGPeer(id)
    await load()
  } catch (err) {
    message.error('切换 Peer 状态失败: ' + formatError(err))
  }
}

async function deletePeer(id: number) {
  try {
    await desktopApi.deleteWGPeer(id)
    message.success('Peer 已删除')
    await load()
  } catch (err) {
    message.error('删除失败: ' + formatError(err))
  }
}

async function openExport(id: number) {
  try {
    exportCfg.value = await desktopApi.exportWGPeerConfig(id)
    showExportModal.value = true
  } catch (err) {
    message.error('导出配置失败: ' + formatError(err))
  }
}

async function copyConf(text: string) {
  await copyToClipboard(text)
  message.success('配置已复制到剪贴板')
}

async function exportToFile(filename: string, content: string) {
  try {
    const savedPath = await desktopApi.saveTextFile(filename, content)
    if (savedPath) {
      message.success(`已保存至 ${savedPath}`)
    }
  } catch (err) {
    message.error('保存文件失败: ' + formatError(err))
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
