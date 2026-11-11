<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-semibold">入站配置</h2>
      <div class="flex gap-2">
        <n-button size="small" @click="load" :disabled="!conn.status.connected">刷新</n-button>
        <n-button size="small" type="primary" :disabled="!conn.status.connected" @click="openCreate">
          新增入站
        </n-button>
      </div>
    </div>

    <n-data-table :columns="columns" :data="inbounds" :loading="loading" size="small" />

    <n-modal v-model:show="showModal" preset="card" :title="editingId ? '编辑入站' : '新增入站'" class="w-[620px]">
      <n-form label-placement="left" label-width="120" size="small">
        <n-form-item label="标签 (Tag)">
          <n-input v-model:value="form.tag" placeholder="留空自动生成" />
        </n-form-item>
        <n-form-item label="协议">
          <n-select
            v-model:value="form.protocol"
            :options="[
              { label: 'VLESS', value: 'vless' },
              { label: 'Trojan', value: 'trojan' },
              { label: 'WireGuard (中转入站)', value: 'wireguard' }
            ]"
          />
        </n-form-item>

        <template v-if="form.protocol !== 'wireguard'">
          <n-form-item label="监听方式">
            <n-switch v-model:value="form.use_uds" />
            <span class="text-xs text-gray-400 ml-2">{{ form.use_uds ? 'UDS 内存套接字 (推荐)' : 'TCP 本地端口' }}</span>
          </n-form-item>
          <n-form-item v-if="!form.use_uds" label="TCP 端口">
            <div class="flex gap-2 w-full">
              <n-input-number v-model:value="form.port" :min="0" :max="65535" class="flex-1" placeholder="0 自动生成" />
              <n-button size="small" @click="form.port = Math.floor(Math.random() * 50000) + 10000">随机</n-button>
            </div>
          </n-form-item>
          <n-form-item label="传输协议">
            <n-select
              v-model:value="form.transport"
              :options="[
                { label: 'XHTTP (推荐)', value: 'xhttp' },
                { label: 'gRPC', value: 'grpc' },
                { label: 'WebSocket', value: 'ws' }
              ]"
            />
          </n-form-item>
          <n-form-item v-if="form.transport !== 'grpc'" label="路径">
            <div class="flex gap-2 w-full">
              <n-input v-model:value="form.path" placeholder="留空自动生成随机路径" />
              <n-button size="small" @click="randomPath">随机</n-button>
            </div>
          </n-form-item>
          <n-form-item v-if="form.transport !== 'grpc'" label="Host (可选)">
            <n-input v-model:value="form.host" />
          </n-form-item>
          <n-form-item v-if="form.transport === 'grpc'" label="gRPC ServiceName">
            <n-input v-model:value="form.service_name" placeholder="grpc-service" />
          </n-form-item>
          <n-form-item label="关联域名 (Nginx)">
            <n-select v-model:value="form.domain_id" :options="domainOptions" clearable placeholder="无（免 Nginx / 隧道模式）" />
          </n-form-item>
          <n-form-item v-if="!form.domain_id" label="自定义 SNI">
            <n-input v-model:value="form.custom_sni" placeholder="纯隧道模式（如 CF Tunnel）填写实际域名" />
          </n-form-item>
          <n-form-item label="连接目标域名">
            <n-input v-model:value="form.connect_domain" placeholder="可选，CDN 优选域名" />
          </n-form-item>
          <n-form-item label="排除出订阅">
            <n-switch v-model:value="form.exclude_from_sub" />
          </n-form-item>
        </template>

        <template v-else>
          <n-form-item label="UDP 监听端口">
            <div class="flex gap-2 w-full">
              <n-input-number v-model:value="form.port" :min="0" :max="65535" class="flex-1" />
              <n-button size="small" @click="form.port = Math.floor(Math.random() * 10000) + 51820">随机</n-button>
            </div>
          </n-form-item>
          <n-form-item label="本端虚拟 IP">
            <n-input v-model:value="form.wg_local_ip" placeholder="10.0.0.1/32" />
          </n-form-item>
          <n-form-item label="私钥 (SecretKey)" :required="!editingId">
            <div class="flex gap-2 w-full">
              <n-input
                v-model:value="form.wg_secret_key"
                type="password"
                show-password-on="click"
                :placeholder="hasExistingWGSecret ? '已保存（留空保持不变）' : '点击右侧生成'"
              />
              <n-button size="small" @click="genWGKeys">生成密钥对</n-button>
            </div>
          </n-form-item>
          <n-form-item label="公钥 (PublicKey)">
            <n-input v-model:value="form.wg_public_key" class="mono" />
          </n-form-item>
          <n-form-item label="对端公钥 (可选)">
            <n-input v-model:value="form.wg_peer_pub_key" class="mono" />
          </n-form-item>
          <n-form-item label="MTU">
            <n-input-number v-model:value="form.wg_mtu" :min="576" :max="1500" class="w-full" />
          </n-form-item>
        </template>

        <n-form-item label="备注">
          <n-input v-model:value="form.remark" type="textarea" :rows="2" />
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="save">保存</n-button>
        </div>
      </template>
    </n-modal>

    <n-modal v-model:show="showWGModal" preset="card" title="WireGuard 客户端配置参考" class="w-[560px]">
      <div v-if="wgClientCfg" class="space-y-3">
        <div class="text-xs text-gray-400">
          Endpoint：<span class="mono text-gray-200">{{ wgClientCfg.endpoint }}</span>
        </div>
        <pre class="p-3 rounded bg-black/40 border border-white/10 mono text-xs overflow-auto">{{ wgClientCfg.wg_conf }}</pre>
        <div class="flex justify-end">
          <n-button size="small" type="primary" @click="copyConf(wgClientCfg.wg_conf)">复制配置</n-button>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useMessage, type DataTableColumns } from 'naive-ui'
import { copyToClipboard, desktopApi, formatError } from '@/api/client'
import type {
  DomainDTO,
  InboundDTO,
  InboundProtocol,
  InboundTransport,
  InboundWGClientConfigResponse
} from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const inbounds = ref<InboundDTO[]>([])
const domains = ref<DomainDTO[]>([])
const loading = ref(false)
const saving = ref(false)
const showModal = ref(false)
const editingId = ref('')
const hasExistingWGSecret = ref(false)

const showWGModal = ref(false)
const wgClientCfg = ref<InboundWGClientConfigResponse | null>(null)

const form = reactive({
  tag: '',
  protocol: 'vless' as InboundProtocol,
  transport: 'xhttp' as InboundTransport,
  port: 0,
  use_uds: true,
  path: '',
  host: '',
  service_name: '',
  domain_id: '' as string | null,
  custom_sni: '',
  connect_domain: '',
  exclude_from_sub: false,
  wg_secret_key: '',
  wg_public_key: '',
  wg_peer_pub_key: '',
  wg_local_ip: '10.0.0.1/32',
  wg_mtu: 1420,
  remark: ''
})

const domainOptions = computed(() =>
  domains.value.map((d) => ({ label: d.domain, value: d.id }))
)

const columns: DataTableColumns<InboundDTO> = [
  { title: '标签', key: 'tag', width: 150 },
  {
    title: '协议 / 传输',
    key: 'proto',
    width: 140,
    render: (r) => `${r.protocol.toUpperCase()} / ${r.transport.toUpperCase()}`
  },
  {
    title: '监听目标',
    key: 'listen',
    render: (r) => (r.use_uds ? r.socket_path : `${r.listen}:${r.port}`)
  },
  {
    title: '域名 / SNI',
    key: 'sni',
    render: (r) => r.actual_domain || r.custom_sni || (r.protocol === 'wireguard' ? '内部中转' : '-')
  },
  {
    title: '状态',
    key: 'enabled',
    width: 80,
    render: (r) => h(NTag, { size: 'small', type: r.enabled ? 'success' : 'default' }, () => (r.enabled ? '启用' : '停用'))
  },
  {
    title: '操作',
    key: 'actions',
    width: 260,
    render: (r) =>
      h(NSpace, { size: 'small' }, () => [
        r.protocol === 'wireguard'
          ? h(NButton, { size: 'tiny', type: 'info', secondary: true, onClick: () => openWGClient(r.id) }, () => '客户端配置')
          : null,
        h(NButton, { size: 'tiny', onClick: () => openEdit(r) }, () => '编辑'),
        h(NButton, { size: 'tiny', onClick: () => toggle(r.id) }, () => (r.enabled ? '停用' : '启用')),
        h(
          NPopconfirm,
          { onPositiveClick: () => remove(r.id) },
          {
            trigger: () => h(NButton, { size: 'tiny', type: 'error', quaternary: true }, () => '删除'),
            default: () => `确认删除入站 ${r.tag}？（关联的受控 Nginx 配置将同步清理）`
          }
        )
      ])
  }
]

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const [inbRes, domRes] = await Promise.all([desktopApi.listInbounds(), desktopApi.listDomains()])
    inbounds.value = inbRes.inbounds
    domains.value = domRes.domains
  } catch (err) {
    message.error('加载入站失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  hasExistingWGSecret.value = false
  Object.assign(form, {
    tag: '',
    protocol: 'vless',
    transport: 'xhttp',
    port: 0,
    use_uds: true,
    path: '',
    host: '',
    service_name: '',
    domain_id: null,
    custom_sni: '',
    connect_domain: '',
    exclude_from_sub: false,
    wg_secret_key: '',
    wg_public_key: '',
    wg_peer_pub_key: '',
    wg_local_ip: '10.0.0.1/32',
    wg_mtu: 1420,
    remark: ''
  })
  showModal.value = true
}

function openEdit(r: InboundDTO) {
  editingId.value = r.id
  hasExistingWGSecret.value = r.has_wg_secret_key
  Object.assign(form, {
    tag: r.tag,
    protocol: r.protocol,
    transport: r.transport,
    port: r.port,
    use_uds: r.use_uds,
    path: r.path,
    host: r.host,
    service_name: r.service_name,
    domain_id: r.domain_id || null,
    custom_sni: r.custom_sni,
    connect_domain: r.connect_domain,
    exclude_from_sub: r.exclude_from_sub,
    wg_secret_key: '',
    wg_public_key: r.wg_public_key,
    wg_peer_pub_key: r.wg_peer_pub_key,
    wg_local_ip: r.wg_local_ip || '10.0.0.1/32',
    wg_mtu: r.wg_mtu || 1420,
    remark: r.remark
  })
  showModal.value = true
}

function randomPath() {
  const bytes = new Uint8Array(8)
  crypto.getRandomValues(bytes)
  form.path = '/' + Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
}

async function genWGKeys() {
  const kp = await desktopApi.generateLocalWGKeys()
  form.wg_secret_key = kp.private_key
  form.wg_public_key = kp.public_key
}

async function save() {
  saving.value = true
  try {
    const payload = {
      tag: form.tag,
      protocol: form.protocol,
      transport: form.protocol === 'wireguard' ? 'raw' : form.transport,
      port: form.port,
      use_uds: form.use_uds,
      path: form.path,
      host: form.host,
      service_name: form.service_name,
      domain_id: form.domain_id || '',
      custom_sni: form.custom_sni,
      connect_domain: form.connect_domain,
      exclude_from_sub: form.exclude_from_sub,
      wg_secret_key: form.wg_secret_key,
      wg_public_key: form.wg_public_key,
      wg_peer_pub_key: form.wg_peer_pub_key,
      wg_local_ip: form.wg_local_ip,
      wg_mtu: form.wg_mtu,
      remark: form.remark
    }
    if (editingId.value) {
      await desktopApi.updateInbound(editingId.value, payload)
    } else {
      await desktopApi.createInbound(payload)
    }
    showModal.value = false
    message.success('入站已保存')
    await load()
  } catch (err) {
    message.error('保存失败: ' + formatError(err))
  } finally {
    saving.value = false
  }
}

async function toggle(id: string) {
  await desktopApi.toggleInbound(id)
  await load()
}

async function remove(id: string) {
  await desktopApi.deleteInbound(id)
  message.success('已删除入站')
  await load()
}

async function openWGClient(id: string) {
  wgClientCfg.value = await desktopApi.getInboundWGClient(id)
  showWGModal.value = true
}

async function copyConf(text: string) {
  await copyToClipboard(text)
  message.success('配置已复制')
}

onMounted(load)
</script>
