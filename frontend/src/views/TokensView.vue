<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between flex-wrap gap-2">
      <h2 class="text-xl font-semibold">API Token 凭证管理</h2>
      <div class="flex gap-2">
        <n-button size="small" @click="load" :disabled="!conn.status.connected" :loading="loading">刷新</n-button>
        <n-button size="small" type="primary" :disabled="!conn.status.connected" @click="openCreate">
          签发新 Token
        </n-button>
      </div>
    </div>

    <n-alert type="info" size="small">
      服务端采用静态 SHA-256 摘要存储 <span class="mono">xrp_...</span> Bearer Token，明文仅在签发时展示一次。当前正在使用的 Token 名称：
      <span class="mono font-medium">{{ conn.status.token_name || '-' }}</span>
    </n-alert>

    <n-data-table :columns="columns" :data="tokens" :loading="loading" size="small" />

    <n-modal v-model:show="showCreateModal" preset="card" title="签发 API Token" class="w-[460px]">
      <n-form label-placement="left" label-width="110" size="small">
        <n-form-item label="Token 名称" required>
          <n-input v-model:value="createForm.name" placeholder="例如：macbook-work / ci-deploy" />
        </n-form-item>
        <n-form-item label="有效天数">
          <n-input-number
            v-model:value="createForm.expires_days"
            :min="0"
            :max="3650"
            class="w-full"
            placeholder="0 或留空表示永不过期"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button size="small" @click="showCreateModal = false">取消</n-button>
          <n-button size="small" type="primary" :loading="creating" @click="createToken">签发</n-button>
        </div>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showSecretModal"
      preset="card"
      title="明文 Token 已生成（仅展示一次）"
      :mask-closable="false"
      class="w-[560px]"
    >
      <div v-if="createdResp" class="space-y-3">
        <n-alert type="warning" size="small">
          请立即复制并保存下方 API Token！关闭弹窗后服务端仅保留哈希值，无法再次查看明文。
        </n-alert>
        <div class="text-xs text-muted">
          名称：<span class="text-strong font-medium">{{ createdResp.token.name }}</span> · 过期时间：
          <span class="mono">{{ createdResp.token.expires_at ? createdResp.token.expires_at.slice(0, 10) : '永不过期' }}</span>
        </div>
        <div class="p-3 rounded surface-code border border-subtle mono text-sm break-all select-all text-accent">
          {{ createdResp.plaintext_token }}
        </div>
        <div class="flex justify-end gap-2">
          <n-button size="small" type="primary" @click="copyPlaintext">复制明文 Token</n-button>
          <n-button size="small" @click="showSecretModal = false">我已妥善保存</n-button>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref, watch } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useMessage, type DataTableColumns } from 'naive-ui'
import { copyToClipboard, desktopApi, formatError } from '@/api/client'
import type { APITokenDTO, CreateTokenResponse } from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const tokens = ref<APITokenDTO[]>([])
const loading = ref(false)
const creating = ref(false)

const showCreateModal = ref(false)
const createForm = reactive({
  name: '',
  expires_days: null as number | null
})

const showSecretModal = ref(false)
const createdResp = ref<CreateTokenResponse | null>(null)

function tokenStatusTag(t: APITokenDTO) {
  if (t.revoked) {
    return h(NTag, { size: 'small', type: 'error' }, () => '已吊销')
  }
  if (t.expires_at && new Date(t.expires_at).getTime() <= Date.now()) {
    return h(NTag, { size: 'small', type: 'warning' }, () => '已过期')
  }
  return h(NTag, { size: 'small', type: 'success' }, () => '有效')
}

const columns: DataTableColumns<APITokenDTO> = [
  {
    title: '名称',
    key: 'name',
    width: 170,
    render: (t) =>
      h(NSpace, { size: 6, align: 'center' }, () => [
        h('span', { class: 'font-medium' }, t.name),
        conn.status.token_name === t.name
          ? h(NTag, { size: 'tiny', type: 'info' }, () => '当前会话')
          : null
      ])
  },
  {
    title: '前缀',
    key: 'token_prefix',
    width: 140,
    render: (t) => h('span', { class: 'mono text-xs text-secondary' }, `${t.token_prefix}...`)
  },
  {
    title: '状态',
    key: 'status',
    width: 85,
    render: (t) => tokenStatusTag(t)
  },
  {
    title: '过期时间',
    key: 'expires_at',
    width: 120,
    render: (t) => (t.expires_at ? t.expires_at.slice(0, 10) : '永不过期')
  },
  {
    title: '最近使用 / IP',
    key: 'last_used',
    render: (t) => {
      if (!t.last_used_at) return '尚未使用'
      const dt = t.last_used_at.replace('T', ' ').slice(0, 19)
      return `${dt}${t.last_ip ? ` (${t.last_ip})` : ''}`
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 165,
    render: (t) =>
      h(NSpace, { size: 'small' }, () => [
        !t.revoked
          ? h(
              NPopconfirm,
              { onPositiveClick: () => deleteToken(t.id, false) },
              {
                trigger: () => h(NButton, { size: 'tiny', type: 'warning', secondary: true }, () => '吊销'),
                default: () => `确认吊销 Token「${t.name}」？吊销后立即失效，如当前会话使用此 Token 将需更换。`
              }
            )
          : null,
        h(
          NPopconfirm,
          { onPositiveClick: () => deleteToken(t.id, true) },
          {
            trigger: () => h(NButton, { size: 'tiny', type: 'error', quaternary: true }, () => '删除'),
            default: () => `确认永久删除 Token「${t.name}」记录？`
          }
        )
      ])
  }
]

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const res = await desktopApi.listTokens()
    tokens.value = res.tokens
  } catch (err) {
    message.error('加载 API Token 失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  createForm.name = ''
  createForm.expires_days = null
  showCreateModal.value = true
}

async function createToken() {
  if (!createForm.name.trim()) {
    message.warning('请填写 Token 名称')
    return
  }
  creating.value = true
  try {
    const days = createForm.expires_days && createForm.expires_days > 0 ? createForm.expires_days : null
    const res = await desktopApi.createToken({
      name: createForm.name.trim(),
      expires_days: days
    })
    showCreateModal.value = false
    createdResp.value = res
    showSecretModal.value = true
    await load()
  } catch (err) {
    message.error('签发 Token 失败: ' + formatError(err))
  } finally {
    creating.value = false
  }
}

async function copyPlaintext() {
  if (!createdResp.value) return
  await copyToClipboard(createdResp.value.plaintext_token)
  message.success('明文 Token 已复制到剪贴板')
}

async function deleteToken(id: string, purge: boolean) {
  try {
    await desktopApi.deleteToken(id, purge)
    message.success(purge ? 'Token 已永久删除' : 'Token 已吊销')
    await load()
  } catch (err) {
    message.error('操作失败: ' + formatError(err))
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
