<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between flex-wrap gap-2">
      <h2 class="text-xl font-semibold">域名与证书管理</h2>
      <div class="flex gap-2 flex-wrap">
        <n-button size="small" @click="load" :disabled="!conn.status.connected">刷新</n-button>
        <n-button size="small" secondary :disabled="!conn.status.connected" :loading="scanning" @click="openScanModal">
          扫描本地证书
        </n-button>
        <n-button size="small" secondary :disabled="!conn.status.connected" @click="openUploadModal">
          上传证书 (PEM)
        </n-button>
        <n-button size="small" type="info" secondary :disabled="!conn.status.connected" @click="openIssueModal">
          API 申请证书 (CF / LuaDNS)
        </n-button>
        <n-button size="small" type="primary" :disabled="!conn.status.connected" @click="openCreate">
          新增域名
        </n-button>
      </div>
    </div>

    <n-data-table :columns="columns" :data="domains" :loading="loading" size="small" />

    <n-modal v-model:show="showModal" preset="card" :title="editingId ? '编辑域名' : '新增域名'" class="w-[620px]">
      <n-form label-placement="left" label-width="130" size="small">
        <n-form-item label="域名" required>
          <n-input v-model:value="form.domain" placeholder="example.com 或 *.example.com" />
        </n-form-item>
        <n-form-item label="类型" required>
          <n-select
            v-model:value="form.type"
            :options="[
              { label: '直连域名 (direct - Nginx TLS)', value: 'direct' },
              { label: 'CDN 优选/反代 (cdn)', value: 'cdn' },
              { label: 'Reality 伪装 (reality)', value: 'reality' }
            ]"
          />
        </n-form-item>
        <n-form-item label="通配符证书">
          <n-switch v-model:value="form.is_wildcard" />
          <span class="text-xs text-gray-400 ml-2">包含 *.example.com 泛解析</span>
        </n-form-item>

        <template v-if="form.type === 'direct' || form.type === 'cdn'">
          <n-form-item label="证书路径 (Cert)" required>
            <n-input
              v-model:value="form.cert_path"
              class="mono"
              placeholder="/root/.acme.sh/example.com_ecc/fullchain.cer"
            />
          </n-form-item>
          <n-form-item label="私钥路径 (Key)" required>
            <n-input
              v-model:value="form.key_path"
              class="mono"
              placeholder="/root/.acme.sh/example.com_ecc/example.com.key"
            />
          </n-form-item>
        </template>

        <template v-else>
          <n-form-item label="伪装域名 (SNI)">
            <n-input v-model:value="form.server_name" placeholder="www.microsoft.com" />
          </n-form-item>
          <n-form-item label="uTLS 指纹">
            <n-select
              v-model:value="form.fingerprint"
              filterable
              tag
              clearable
              :options="[
                { label: 'chrome', value: 'chrome' },
                { label: 'firefox', value: 'firefox' },
                { label: 'safari', value: 'safari' },
                { label: 'ios', value: 'ios' },
                { label: 'android', value: 'android' },
                { label: 'edge', value: 'edge' },
                { label: 'random', value: 'random' }
              ]"
            />
          </n-form-item>
          <n-form-item label="Short ID">
            <n-input v-model:value="form.short_id" class="mono" placeholder="16 进制短 ID（如 a1b2c3d4）" />
          </n-form-item>
          <n-form-item label="Reality 私钥">
            <n-input
              v-model:value="form.private_key"
              type="password"
              show-password-on="click"
              class="mono"
              :placeholder="hasExistingPrivateKey ? '已保存（留空保持不变）' : '可选，X25519 私钥'"
            />
          </n-form-item>
          <n-form-item label="Reality 公钥">
            <n-input v-model:value="form.public_key" class="mono" placeholder="可选，客户端共享公钥" />
          </n-form-item>
        </template>

        <n-form-item label="启用">
          <n-switch v-model:value="form.enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button size="small" @click="showModal = false">取消</n-button>
          <n-button size="small" type="primary" :loading="saving" @click="save">保存</n-button>
        </div>
      </template>
    </n-modal>

    <n-modal v-model:show="showScanModal" preset="card" title="扫描服务器证书" class="w-[760px]">
      <div v-if="scanResult" class="space-y-3">
        <div class="flex items-center justify-between flex-wrap gap-2">
          <div class="text-xs text-gray-400">
            扫描目录：<span class="mono text-gray-200">{{ scanResult.cert_dir }}</span> · 发现 {{ scanResult.count }} 组证书
          </div>
          <n-button
            size="small"
            type="primary"
            secondary
            :disabled="scanResult.count === 0"
            :loading="batchImporting"
            @click="batchImportCerts"
          >
            一键全部导入 / 更新
          </n-button>
        </div>

        <n-data-table :columns="scanColumns" :data="scanResult.certificates" size="small" :max-height="380" />
      </div>
    </n-modal>

    <n-modal v-model:show="showUploadModal" preset="card" title="上传 TLS 证书与私钥 (PEM)" class="w-[660px]">
      <n-form label-placement="left" label-width="120" size="small">
        <div class="grid grid-cols-2 gap-3">
          <n-form-item label="绑定域名" required>
            <n-input v-model:value="uploadForm.domain" class="mono" placeholder="example.com 或 *.example.com" />
          </n-form-item>
          <n-form-item label="域名类型">
            <n-select
              v-model:value="uploadForm.type"
              :options="[
                { label: '直连域名 (direct)', value: 'direct' },
                { label: 'CDN 反代 (cdn)', value: 'cdn' }
              ]"
            />
          </n-form-item>
        </div>

        <n-form-item label="证书内容 (PEM)" required>
          <div class="w-full space-y-1.5">
            <input
              type="file"
              accept=".pem,.cer,.crt,.txt"
              class="text-xs text-gray-300"
              @change="(e) => onPickPEMFile(e, 'cert')"
            />
            <n-input
              v-model:value="uploadForm.cert_pem"
              type="textarea"
              :rows="5"
              class="mono text-xs"
              placeholder="-----BEGIN CERTIFICATE-----&#10;..."
            />
          </div>
        </n-form-item>

        <n-form-item label="私钥内容 (KEY)" required>
          <div class="w-full space-y-1.5">
            <input
              type="file"
              accept=".key,.pem,.txt"
              class="text-xs text-gray-300"
              @change="(e) => onPickPEMFile(e, 'key')"
            />
            <n-input
              v-model:value="uploadForm.key_pem"
              type="textarea"
              :rows="4"
              class="mono text-xs"
              placeholder="-----BEGIN PRIVATE KEY----- / -----BEGIN EC PRIVATE KEY-----&#10;..."
            />
          </div>
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button size="small" @click="showUploadModal = false">取消</n-button>
          <n-button size="small" type="primary" :loading="uploading" @click="submitUpload">校验并上传保存</n-button>
        </div>
      </template>
    </n-modal>

    <n-modal v-model:show="showIssueModal" preset="card" title="通过 DNS API 自动申请 ACME 证书" class="w-[660px]">
      <n-form label-placement="left" label-width="135" size="small">
        <n-alert type="info" size="small" class="mb-3">
          服务端将通过 DNS-01 自动完成验证与签发（支持泛域名），并将凭据交由 acme.sh 本地托管用于后续自动续期（面板数据库不落盘您的 API 密钥）。
        </n-alert>

        <div class="grid grid-cols-2 gap-3">
          <n-form-item label="申请域名" required>
            <n-input v-model:value="issueForm.domain" class="mono" placeholder="example.com 或 *.example.com" />
          </n-form-item>
          <n-form-item label="域名类型">
            <n-select
              v-model:value="issueForm.type"
              :options="[
                { label: '直连域名 (direct)', value: 'direct' },
                { label: 'CDN 反代 (cdn)', value: 'cdn' }
              ]"
            />
          </n-form-item>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <n-form-item label="同时申请泛域名">
            <n-switch v-model:value="issueForm.include_wildcard" />
            <span class="text-xs text-gray-400 ml-2">包含 *.域名</span>
          </n-form-item>
          <n-form-item label="强制重签">
            <n-switch v-model:value="issueForm.force" />
          </n-form-item>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <n-form-item label="DNS 服务商" required>
            <n-select
              v-model:value="issueForm.provider"
              :options="[
                { label: 'Cloudflare (dns_cf)', value: 'cf' },
                { label: 'LuaDNS (dns_lua)', value: 'luadns' }
              ]"
            />
          </n-form-item>
          <n-form-item label="CA 机构">
            <n-select
              v-model:value="issueForm.ca_server"
              :options="[
                { label: 'Let\'s Encrypt (推荐，免 EAB)', value: 'letsencrypt' },
                { label: 'ZeroSSL (需填邮箱)', value: 'zerossl' }
              ]"
            />
          </n-form-item>
        </div>

        <n-form-item label="账户/续期邮箱" :required="issueForm.ca_server === 'zerossl'">
          <n-input v-model:value="issueForm.email" placeholder="Let's Encrypt 可选；ZeroSSL 必填" />
        </n-form-item>

        <template v-if="issueForm.provider === 'cf'">
          <n-form-item label="CF 鉴权方式">
            <n-radio-group v-model:value="cfAuthMode" size="small">
              <n-radio-button value="token">Scoped API Token (推荐)</n-radio-button>
              <n-radio-button value="key">Global API Key + Email</n-radio-button>
            </n-radio-group>
          </n-form-item>
          <n-form-item v-if="cfAuthMode === 'token'" label="CF_Token" required>
            <n-input
              v-model:value="issueForm.cf_token"
              type="password"
              show-password-on="click"
              class="mono"
              placeholder="需具备 Zone -> DNS -> Edit 权限"
            />
          </n-form-item>
          <template v-else>
            <n-form-item label="CF_Email" required>
              <n-input v-model:value="issueForm.cf_email" class="mono" placeholder="Cloudflare 登录邮箱" />
            </n-form-item>
            <n-form-item label="CF_Key" required>
              <n-input
                v-model:value="issueForm.cf_key"
                type="password"
                show-password-on="click"
                class="mono"
                placeholder="Cloudflare Global API Key"
              />
            </n-form-item>
          </template>
          <n-form-item label="CF_Zone_ID (可选)">
            <n-input v-model:value="issueForm.cf_zone_id" class="mono" placeholder="单域名 Token 若限制查询域列表可填" />
          </n-form-item>
        </template>

        <template v-else>
          <n-form-item label="LuaDNS 邮箱" required>
            <n-input v-model:value="issueForm.lua_email" class="mono" placeholder="LUA_Email" />
          </n-form-item>
          <n-form-item label="LuaDNS API Key" required>
            <n-input
              v-model:value="issueForm.lua_key"
              type="password"
              show-password-on="click"
              class="mono"
              placeholder="LUA_Key"
            />
          </n-form-item>
        </template>

        <div v-if="issueOutput" class="mt-2">
          <div class="text-xs text-gray-400 mb-1">签发日志：</div>
          <pre class="p-2.5 rounded bg-black/50 border border-white/10 mono text-xs max-h-40 overflow-auto whitespace-pre-wrap">{{ issueOutput }}</pre>
        </div>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button size="small" @click="showIssueModal = false">关闭</n-button>
          <n-button size="small" type="primary" :loading="issuing" @click="submitIssue">立即申请并绑定</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref, watch } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useMessage, type DataTableColumns } from 'naive-ui'
import { desktopApi, formatError } from '@/api/client'
import type {
  CertificateInfo,
  DomainDTO,
  DomainType,
  ScanCertificatesResponse
} from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const domains = ref<DomainDTO[]>([])
const loading = ref(false)
const saving = ref(false)
const scanning = ref(false)
const batchImporting = ref(false)
const importingPath = ref('')

const showModal = ref(false)
const editingId = ref('')
const hasExistingPrivateKey = ref(false)

const showScanModal = ref(false)
const scanResult = ref<ScanCertificatesResponse | null>(null)

const showUploadModal = ref(false)
const uploading = ref(false)
const uploadForm = reactive({
  domain: '',
  type: 'direct' as 'direct' | 'cdn',
  cert_pem: '',
  key_pem: ''
})

const showIssueModal = ref(false)
const issuing = ref(false)
const issueOutput = ref('')
const cfAuthMode = ref<'token' | 'key'>('token')
const issueForm = reactive({
  domain: '',
  include_wildcard: true,
  type: 'direct' as 'direct' | 'cdn',
  provider: 'cf' as 'cf' | 'luadns',
  cf_token: '',
  cf_key: '',
  cf_email: '',
  cf_zone_id: '',
  lua_key: '',
  lua_email: '',
  ca_server: 'letsencrypt' as 'letsencrypt' | 'zerossl',
  email: '',
  force: false
})

const form = reactive({
  domain: '',
  type: 'direct' as DomainType,
  is_wildcard: false,
  server_name: '',
  fingerprint: 'chrome',
  short_id: '',
  private_key: '',
  public_key: '',
  cert_path: '',
  key_path: '',
  enabled: true
})

function certStatusTag(d: DomainDTO) {
  if (d.type === 'reality') {
    return h(NTag, { size: 'small', type: 'info', bordered: false }, () => d.server_name ? `SNI: ${d.server_name}` : 'Reality 模式')
  }
  if (!d.has_cert) {
    return h(NTag, { size: 'small', type: 'error' }, () => '未检测到有效证书')
  }
  const days = d.days_to_expiry ?? 0
  const dateStr = d.cert_expiry_date ? d.cert_expiry_date.slice(0, 10) : ''
  const tagType = days <= 0 ? 'error' : days <= 30 ? 'warning' : 'success'
  return h(NTag, { size: 'small', type: tagType }, () => `${dateStr} (剩 ${days} 天)`)
}

const columns: DataTableColumns<DomainDTO> = [
  {
    title: '域名',
    key: 'domain',
    width: 210,
    render: (r) =>
      h(NSpace, { size: 6, align: 'center' }, () => [
        h('span', { class: 'mono font-medium' }, r.domain),
        r.is_wildcard ? h(NTag, { size: 'tiny', type: 'info' }, () => '通配符') : null
      ])
  },
  {
    title: '类型',
    key: 'type',
    width: 90,
    render: (r) => {
      const label = r.type === 'direct' ? '直连' : r.type === 'cdn' ? 'CDN' : 'Reality'
      return h(NTag, { size: 'small', bordered: false }, () => label)
    }
  },
  {
    title: '证书 / 到期状态',
    key: 'expiry',
    width: 200,
    render: (r) => certStatusTag(r)
  },
  {
    title: '证书路径 / 签发者',
    key: 'issuer',
    ellipsis: { tooltip: true },
    render: (r) => r.cert_issuer || r.cert_path || '-'
  },
  {
    title: '状态',
    key: 'enabled',
    width: 75,
    render: (r) => h(NTag, { size: 'small', type: r.enabled ? 'success' : 'default' }, () => (r.enabled ? '启用' : '停用'))
  },
  {
    title: '操作',
    key: 'actions',
    width: 210,
    render: (r) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'tiny', onClick: () => openEdit(r) }, () => '编辑'),
        h(NButton, { size: 'tiny', onClick: () => toggle(r) }, () => (r.enabled ? '停用' : '启用')),
        h(
          NPopconfirm,
          { onPositiveClick: () => remove(r.id) },
          {
            trigger: () => h(NButton, { size: 'tiny', type: 'error', quaternary: true }, () => '删除'),
            default: () => `确认删除域名 ${r.domain}？`
          }
        )
      ])
  }
]

const scanColumns: DataTableColumns<CertificateInfo> = [
  {
    title: '域名',
    key: 'domain',
    width: 190,
    render: (c) =>
      h(NSpace, { size: 4, align: 'center' }, () => [
        h('span', { class: 'mono' }, c.domain),
        c.is_wildcard ? h(NTag, { size: 'tiny', type: 'info' }, () => '通配符') : null
      ])
  },
  {
    title: '有效期',
    key: 'status',
    width: 150,
    render: (c) => {
      const t = c.status === 'valid' ? 'success' : c.status === 'expiring_soon' ? 'warning' : 'error'
      return h(NTag, { size: 'tiny', type: t }, () => `${c.days_to_expiry} 天 (${c.expiry_date.slice(0, 10)})`)
    }
  },
  { title: '签发者', key: 'issuer', width: 140, ellipsis: { tooltip: true } },
  {
    title: '证书路径',
    key: 'cert_path',
    ellipsis: { tooltip: true },
    render: (c) => h('span', { class: 'mono text-xs' }, c.cert_path)
  },
  {
    title: '操作',
    key: 'op',
    width: 95,
    render: (c) =>
      h(
        NButton,
        {
          size: 'tiny',
          type: 'primary',
          secondary: true,
          loading: importingPath.value === c.cert_path,
          onClick: () => importOneCert(c)
        },
        () => '导入'
      )
  }
]

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const res = await desktopApi.listDomains()
    domains.value = res.domains
  } catch (err) {
    message.error('加载域名列表失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  hasExistingPrivateKey.value = false
  Object.assign(form, {
    domain: '',
    type: 'direct' as DomainType,
    is_wildcard: false,
    server_name: '',
    fingerprint: 'chrome',
    short_id: '',
    private_key: '',
    public_key: '',
    cert_path: '',
    key_path: '',
    enabled: true
  })
  showModal.value = true
}

function openEdit(d: DomainDTO) {
  editingId.value = d.id
  hasExistingPrivateKey.value = d.has_private_key
  Object.assign(form, {
    domain: d.domain,
    type: d.type,
    is_wildcard: d.is_wildcard,
    server_name: d.server_name,
    fingerprint: d.fingerprint || 'chrome',
    short_id: d.short_id,
    private_key: '',
    public_key: d.public_key,
    cert_path: d.cert_path,
    key_path: d.key_path,
    enabled: d.enabled
  })
  showModal.value = true
}

async function save() {
  if (!form.domain.trim()) {
    message.warning('请填写域名')
    return
  }
  saving.value = true
  try {
    const payload = {
      domain: form.domain.trim(),
      type: form.type,
      is_wildcard: form.is_wildcard,
      server_name: form.server_name.trim(),
      fingerprint: form.fingerprint,
      short_id: form.short_id.trim(),
      private_key: form.private_key.trim(),
      public_key: form.public_key.trim(),
      cert_path: form.cert_path.trim(),
      key_path: form.key_path.trim(),
      enabled: form.enabled
    }
    if (editingId.value) {
      await desktopApi.updateDomain(editingId.value, payload)
    } else {
      await desktopApi.createDomain(payload)
    }
    showModal.value = false
    message.success('域名已保存')
    await load()
  } catch (err) {
    message.error('保存失败: ' + formatError(err))
  } finally {
    saving.value = false
  }
}

async function toggle(d: DomainDTO) {
  try {
    await desktopApi.updateDomain(d.id, {
      domain: d.domain,
      type: d.type,
      is_wildcard: d.is_wildcard,
      server_name: d.server_name,
      fingerprint: d.fingerprint,
      short_id: d.short_id,
      cert_path: d.cert_path,
      key_path: d.key_path,
      enabled: !d.enabled
    })
    await load()
  } catch (err) {
    message.error('切换状态失败: ' + formatError(err))
  }
}

async function remove(id: string) {
  try {
    await desktopApi.deleteDomain(id)
    message.success('域名已删除')
    await load()
  } catch (err) {
    message.error('删除失败: ' + formatError(err))
  }
}

async function openScanModal() {
  scanning.value = true
  try {
    scanResult.value = await desktopApi.scanCerts()
    showScanModal.value = true
  } catch (err) {
    message.error('扫描证书失败: ' + formatError(err))
  } finally {
    scanning.value = false
  }
}

async function importOneCert(c: CertificateInfo) {
  importingPath.value = c.cert_path
  try {
    const targetDomain = c.domains.find((d) => d.startsWith('*.')) || c.domain
    await desktopApi.importSingleCert({
      domain: targetDomain,
      cert_path: c.cert_path,
      key_path: c.key_path
    })
    message.success(`证书 ${targetDomain} 已导入`)
    await load()
  } catch (err) {
    message.error('导入证书失败: ' + formatError(err))
  } finally {
    importingPath.value = ''
  }
}

async function batchImportCerts() {
  batchImporting.value = true
  try {
    const res = await desktopApi.scanImportCerts()
    domains.value = res.domains
    if (res.errors.length > 0) {
      message.warning(`导入 ${res.imported_count} 个，更新 ${res.updated_count} 个，跳过 ${res.skipped_count} 个，错误 ${res.errors.length} 个`)
    } else {
      message.success(`导入 ${res.imported_count} 个，更新 ${res.updated_count} 个，跳过 ${res.skipped_count} 个`)
    }
    showScanModal.value = false
  } catch (err) {
    message.error('批量导入失败: ' + formatError(err))
  } finally {
    batchImporting.value = false
  }
}

function openUploadModal() {
  Object.assign(uploadForm, {
    domain: '',
    type: 'direct' as 'direct' | 'cdn',
    cert_pem: '',
    key_pem: ''
  })
  showUploadModal.value = true
}

async function onPickPEMFile(ev: Event, target: 'cert' | 'key') {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const text = await file.text()
  if (target === 'cert') {
    uploadForm.cert_pem = text
  } else {
    uploadForm.key_pem = text
  }
}

async function submitUpload() {
  if (!uploadForm.domain.trim() || !uploadForm.cert_pem.trim() || !uploadForm.key_pem.trim()) {
    message.warning('请填写域名并上传/粘贴证书和私钥内容')
    return
  }
  uploading.value = true
  try {
    const dom = await desktopApi.uploadDomainCert({
      domain: uploadForm.domain.trim(),
      type: uploadForm.type,
      cert_pem: uploadForm.cert_pem,
      key_pem: uploadForm.key_pem
    })
    showUploadModal.value = false
    message.success(`证书 ${dom.domain} 上传并校验通过`)
    await load()
  } catch (err) {
    message.error('上传证书失败: ' + formatError(err))
  } finally {
    uploading.value = false
  }
}

function openIssueModal() {
  issueOutput.value = ''
  showIssueModal.value = true
}

async function submitIssue() {
  if (!issueForm.domain.trim()) {
    message.warning('请填写待申请证书的域名')
    return
  }
  issuing.value = true
  issueOutput.value = ''
  try {
    const res = await desktopApi.issueDomainCert({
      domain: issueForm.domain.trim(),
      include_wildcard: issueForm.include_wildcard,
      type: issueForm.type,
      provider: issueForm.provider,
      cf_token: issueForm.provider === 'cf' && cfAuthMode.value === 'token' ? issueForm.cf_token.trim() : '',
      cf_key: issueForm.provider === 'cf' && cfAuthMode.value === 'key' ? issueForm.cf_key.trim() : '',
      cf_email: issueForm.provider === 'cf' && cfAuthMode.value === 'key' ? issueForm.cf_email.trim() : '',
      cf_zone_id: issueForm.provider === 'cf' ? issueForm.cf_zone_id.trim() : '',
      lua_key: issueForm.provider === 'luadns' ? issueForm.lua_key.trim() : '',
      lua_email: issueForm.provider === 'luadns' ? issueForm.lua_email.trim() : '',
      ca_server: issueForm.ca_server,
      email: issueForm.email.trim(),
      force: issueForm.force
    })
    issueOutput.value = res.output || '签发成功'
    message.success(
      res.reused_existing
        ? `域名 ${res.domain.domain} 已有有效证书，已同步导入`
        : `域名 ${res.domain.domain} 证书申请并配置成功`
    )
    await load()
  } catch (err) {
    const msg = formatError(err)
    issueOutput.value = msg
    message.error('申请证书失败: ' + msg)
  } finally {
    issuing.value = false
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
