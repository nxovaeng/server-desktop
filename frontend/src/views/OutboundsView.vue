<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-semibold">出站配置</h2>
      <div class="flex gap-2">
        <n-button size="small" @click="load" :disabled="!conn.status.connected">刷新</n-button>
        <n-button size="small" secondary :disabled="!conn.status.connected" @click="showImportModal = true">
          导入分享链接
        </n-button>
        <n-button size="small" type="primary" :disabled="!conn.status.connected" @click="openCreate">
          新增出站
        </n-button>
      </div>
    </div>

    <n-data-table :columns="columns" :data="outbounds" :loading="loading" size="small" />

    <n-modal v-model:show="showModal" preset="card" :title="editingId ? '编辑出站' : '新增出站'" class="w-[640px]">
      <n-form label-placement="left" label-width="130" size="small">
        <n-form-item label="标签 (Tag)">
          <n-input v-model:value="form.tag" placeholder="留空自动生成" />
        </n-form-item>
        <n-form-item label="协议类型" required>
          <n-select
            v-model:value="form.type"
            :options="[
              { label: 'SOCKS5 代理', value: 'socks5' },
              { label: 'WireGuard (WARP / 中转落地)', value: 'wireguard' },
              { label: 'Trojan', value: 'trojan' },
              { label: 'VLESS', value: 'vless' },
              { label: 'VMess', value: 'vmess' }
            ]"
          />
        </n-form-item>

        <n-form-item label="服务器地址 / 端口">
          <div class="flex gap-2 w-full">
            <n-input v-model:value="form.server" placeholder="IP 或域名" class="flex-1" />
            <n-input-number v-model:value="form.port" :min="1" :max="65535" class="w-32" />
          </div>
        </n-form-item>

        <template v-if="form.type === 'socks5'">
          <n-form-item label="用户名 (可选)">
            <n-input v-model:value="form.username" />
          </n-form-item>
          <n-form-item label="密码 (可选)">
            <n-input
              v-model:value="form.password"
              type="password"
              show-password-on="click"
              :placeholder="flags.has_password ? '已保存（留空保持不变）' : ''"
            />
          </n-form-item>
        </template>

        <template v-else-if="form.type === 'wireguard'">
          <n-form-item label="导入 .conf">
            <div class="w-full space-y-1.5">
              <n-input v-model:value="wgRawConf" type="textarea" :rows="3" placeholder="粘贴 WireGuard 客户端配置文件..." />
              <n-button size="tiny" secondary @click="parseWGConf">解析配置并填充</n-button>
            </div>
          </n-form-item>
          <n-form-item label="私钥 (SecretKey)">
            <div class="flex gap-2 w-full">
              <n-input
                v-model:value="form.wg_secret_key"
                type="password"
                show-password-on="click"
                :placeholder="flags.has_wg_secret_key ? '已保存（留空保持不变）' : '粘贴或生成'"
              />
              <n-button size="small" @click="genWGKeys">生成密钥对</n-button>
            </div>
          </n-form-item>
          <n-form-item label="本地公钥">
            <n-input v-model:value="form.wg_client_pub_key" class="mono" readonly />
          </n-form-item>
          <n-form-item label="对端服务器公钥">
            <n-input v-model:value="form.wg_public_key" class="mono" />
          </n-form-item>
          <n-form-item label="本地 IPv4 / IPv6">
            <div class="flex gap-2 w-full">
              <n-input v-model:value="form.wg_local_ipv4" placeholder="172.16.0.2/32" />
              <n-input v-model:value="form.wg_local_ipv6" placeholder="可选 IPv6" />
            </div>
          </n-form-item>
          <n-form-item label="Reserved / MTU">
            <div class="flex gap-2 w-full">
              <n-input v-model:value="form.wg_reserved" placeholder="[0,0,0]" />
              <n-input-number v-model:value="form.wg_mtu" :min="576" :max="1500" class="w-32" />
            </div>
          </n-form-item>
        </template>

        <template v-else-if="form.type === 'trojan'">
          <n-form-item label="Trojan 密码">
            <n-input
              v-model:value="form.trojan_password"
              type="password"
              show-password-on="click"
              :placeholder="flags.has_trojan_password ? '已保存（留空保持不变）' : ''"
            />
          </n-form-item>
          <n-form-item label="SNI / 传输">
            <div class="flex gap-2 w-full">
              <n-input v-model:value="form.trojan_sni" placeholder="example.com" class="flex-1" />
              <n-select
                v-model:value="form.trojan_network"
                class="w-36"
                :options="[{ label: 'TCP', value: 'tcp' }, { label: 'WS', value: 'ws' }, { label: 'gRPC', value: 'grpc' }]"
              />
            </div>
          </n-form-item>
        </template>

        <template v-else>
          <n-form-item label="UUID">
            <n-input v-model:value="form.uuid" class="mono" />
          </n-form-item>
          <n-form-item label="传输 / Path">
            <div class="flex gap-2 w-full">
              <n-select
                v-model:value="form.network"
                class="w-32"
                :options="[{ label: 'TCP', value: 'tcp' }, { label: 'WS', value: 'ws' }, { label: 'gRPC', value: 'grpc' }, { label: 'XHTTP', value: 'xhttp' }]"
              />
              <n-input v-model:value="form.path" placeholder="/path" />
            </div>
          </n-form-item>
          <n-form-item label="TLS / SNI">
            <div class="flex items-center gap-2 w-full">
              <n-switch v-model:value="form.tls" />
              <n-input v-model:value="form.tls_server_name" placeholder="TLS SNI" class="flex-1" />
            </div>
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

    <n-modal v-model:show="showImportModal" preset="card" title="导入代理分享链接" class="w-[520px]">
      <div class="space-y-3">
        <n-input
          v-model:value="shareLink"
          type="textarea"
          :rows="4"
          placeholder="粘贴 vless://, vmess:// 或 trojan:// 链接"
        />
        <div class="flex justify-end gap-2">
          <n-button @click="showImportModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="importLink">解析并导入</n-button>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useMessage, type DataTableColumns } from 'naive-ui'
import { desktopApi, formatError } from '@/api/client'
import type { OutboundDTO, OutboundType } from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const outbounds = ref<OutboundDTO[]>([])
const loading = ref(false)
const saving = ref(false)
const testingId = ref('')
const showModal = ref(false)
const showImportModal = ref(false)
const editingId = ref('')
const shareLink = ref('')
const wgRawConf = ref('')

const flags = reactive({
  has_password: false,
  has_wg_secret_key: false,
  has_trojan_password: false
})

const form = reactive({
  tag: '',
  type: 'socks5' as OutboundType,
  server: '',
  port: 1080,
  username: '',
  password: '',
  wg_secret_key: '',
  wg_client_pub_key: '',
  wg_public_key: '',
  wg_reserved: '',
  wg_local_ipv4: '',
  wg_local_ipv6: '',
  wg_mtu: 1420,
  trojan_password: '',
  trojan_sni: '',
  trojan_network: 'tcp',
  uuid: '',
  network: 'tcp',
  path: '',
  tls: false,
  tls_server_name: '',
  remark: ''
})

const columns: DataTableColumns<OutboundDTO> = [
  { title: '标签', key: 'tag', width: 160 },
  { title: '类型', key: 'type', width: 110, render: (r) => r.type.toUpperCase() },
  { title: '节点地址', key: 'endpoint', render: (r) => (r.server ? `${r.server}:${r.port}` : '-') },
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
        h(
          NButton,
          { size: 'tiny', type: 'info', secondary: true, loading: testingId.value === r.id, onClick: () => testDial(r.id) },
          () => '真拨测'
        ),
        h(NButton, { size: 'tiny', onClick: () => openEdit(r) }, () => '编辑'),
        h(NButton, { size: 'tiny', onClick: () => toggle(r.id) }, () => (r.enabled ? '停用' : '启用')),
        h(
          NPopconfirm,
          { onPositiveClick: () => remove(r.id) },
          {
            trigger: () => h(NButton, { size: 'tiny', type: 'error', quaternary: true }, () => '删除'),
            default: () => `确认删除出站 ${r.tag}？`
          }
        )
      ])
  }
]

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const res = await desktopApi.listOutbounds()
    outbounds.value = res.outbounds
  } catch (err) {
    message.error('加载出站失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  Object.assign(flags, { has_password: false, has_wg_secret_key: false, has_trojan_password: false })
  Object.assign(form, {
    tag: '',
    type: 'socks5',
    server: '',
    port: 1080,
    username: '',
    password: '',
    wg_secret_key: '',
    wg_client_pub_key: '',
    wg_public_key: '',
    wg_reserved: '',
    wg_local_ipv4: '',
    wg_local_ipv6: '',
    wg_mtu: 1420,
    trojan_password: '',
    trojan_sni: '',
    trojan_network: 'tcp',
    uuid: '',
    network: 'tcp',
    path: '',
    tls: false,
    tls_server_name: '',
    remark: ''
  })
  showModal.value = true
}

function openEdit(o: OutboundDTO) {
  editingId.value = o.id
  flags.has_password = o.has_password
  flags.has_wg_secret_key = o.has_wg_secret_key
  flags.has_trojan_password = o.has_trojan_password
  Object.assign(form, {
    ...o,
    password: '',
    wg_secret_key: '',
    trojan_password: ''
  })
  showModal.value = true
}

async function parseWGConf() {
  if (!wgRawConf.value.trim()) return
  try {
    const parsed = await desktopApi.parseWireGuardOutbound(wgRawConf.value)
    form.wg_secret_key = parsed.private_key
    form.wg_client_pub_key = parsed.client_public_key
    form.wg_public_key = parsed.public_key
    form.server = parsed.server
    if (parsed.port > 0) form.port = parsed.port
    form.wg_local_ipv4 = parsed.local_ipv4
    form.wg_local_ipv6 = parsed.local_ipv6
    if (parsed.mtu > 0) form.wg_mtu = parsed.mtu
    message.success('WireGuard 配置解析成功')
  } catch (err) {
    message.error('解析失败: ' + formatError(err))
  }
}

async function genWGKeys() {
  const kp = await desktopApi.generateLocalWGKeys()
  form.wg_secret_key = kp.private_key
  form.wg_client_pub_key = kp.public_key
}

async function save() {
  saving.value = true
  try {
    if (editingId.value) {
      await desktopApi.updateOutbound(editingId.value, { ...form })
    } else {
      await desktopApi.createOutbound({ ...form })
    }
    showModal.value = false
    message.success('出站已保存')
    await load()
  } catch (err) {
    message.error('保存失败: ' + formatError(err))
  } finally {
    saving.value = false
  }
}

async function importLink() {
  saving.value = true
  try {
    await desktopApi.importOutbound(shareLink.value)
    showImportModal.value = false
    shareLink.value = ''
    message.success('导入成功')
    await load()
  } catch (err) {
    message.error('导入失败: ' + formatError(err))
  } finally {
    saving.value = false
  }
}

async function testDial(id: string) {
  testingId.value = id
  try {
    const res = await desktopApi.testOutbound(id)
    if (res.success) {
      message.success(`拨测成功 · ${res.latency_ms ?? 0} ms · 出口 IP: ${res.outbound_ip || '-'} (${res.location || '-'})`)
    } else {
      message.error(`拨测失败: ${res.message}`)
    }
  } catch (err) {
    message.error('拨测异常: ' + formatError(err))
  } finally {
    testingId.value = ''
  }
}

async function toggle(id: string) {
  await desktopApi.toggleOutbound(id)
  await load()
}

async function remove(id: string) {
  await desktopApi.deleteOutbound(id)
  message.success('已删除出站')
  await load()
}

onMounted(load)
</script>
