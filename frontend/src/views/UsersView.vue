<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-xl font-semibold">用户管理</h2>
      <div class="flex gap-2">
        <n-input
          v-model:value="search"
          placeholder="搜索用户名或邮箱..."
          clearable
          size="small"
          class="w-56"
          @keyup.enter="load"
          @clear="load"
        />
        <n-button size="small" @click="load" :disabled="!conn.status.connected">搜索</n-button>
        <n-button size="small" type="primary" :disabled="!conn.status.connected" @click="openCreate">
          新增用户
        </n-button>
      </div>
    </div>

    <n-data-table :columns="columns" :data="users" :loading="loading" size="small" />

    <n-modal v-model:show="showModal" preset="card" :title="editingId ? '编辑用户' : '新增用户'" class="w-[540px]">
      <n-form label-placement="left" label-width="110" size="small">
        <n-form-item label="用户名" required>
          <n-input v-model:value="form.name" placeholder="用于识别及订阅文件名" />
        </n-form-item>
        <n-form-item label="邮箱">
          <n-input v-model:value="form.email" placeholder="可选" />
        </n-form-item>
        <n-form-item label="UUID">
          <div class="flex gap-2 w-full">
            <n-input v-model:value="form.uuid" placeholder="留空自动生成 UUIDv4" class="mono" />
            <n-button size="small" @click="randomUUID">随机</n-button>
          </div>
        </n-form-item>
        <n-form-item label="流量限额 (GB)">
          <n-input-number v-model:value="limitGB" :min="0" class="w-full" />
        </n-form-item>
        <n-form-item label="到期时间">
          <n-date-picker
            v-model:value="expiryTs"
            type="date"
            clearable
            class="w-full"
            placeholder="留空永久有效"
          />
        </n-form-item>
        <n-form-item label="WG 公钥 (可选)">
          <n-input v-model:value="form.wg_pub_key" class="mono" placeholder="用于 WireGuard 入站多客户端" />
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="form.note" type="textarea" :rows="2" />
        </n-form-item>
        <n-form-item label="启用状态">
          <n-switch v-model:value="form.enabled" />
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="save">保存</n-button>
        </div>
      </template>
    </n-modal>

    <n-modal v-model:show="showQRModal" preset="card" :title="`订阅链接 - ${qrUser?.name}`" class="w-[460px]">
      <div v-if="qrUser" class="space-y-3">
        <div class="flex justify-center p-4 bg-white rounded">
          <img v-if="qrDataUrl" :src="qrDataUrl" alt="Subscription QR Code" class="w-52 h-52" />
        </div>
        <div class="space-y-2 text-xs">
          <div v-for="(url, fmt) in subUrlEntries" :key="fmt" class="flex items-center gap-2">
            <n-tag size="tiny" class="w-16 justify-center uppercase">{{ fmt }}</n-tag>
            <n-input :value="url" readonly size="tiny" class="mono flex-1" />
            <n-button size="tiny" @click="copyUrl(url)">复制</n-button>
          </div>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useMessage, type DataTableColumns } from 'naive-ui'
import QRCode from 'qrcode'
import { copyToClipboard, desktopApi, formatBytes, formatError } from '@/api/client'
import type { UserDTO } from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const users = ref<UserDTO[]>([])
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const showModal = ref(false)
const editingId = ref('')
const limitGB = ref(0)
const expiryTs = ref<number | null>(null)

const form = reactive({
  name: '',
  email: '',
  uuid: '',
  note: '',
  wg_pub_key: '',
  enabled: true
})

const showQRModal = ref(false)
const qrUser = ref<UserDTO | null>(null)
const qrDataUrl = ref('')

const subUrlEntries = computed((): Record<string, string> => {
  if (!qrUser.value) return {}
  return {
    base64: qrUser.value.sub_urls.base64,
    plain: qrUser.value.sub_urls.plain,
    json: qrUser.value.sub_urls.json,
    yaml: qrUser.value.sub_urls.yaml
  }
})

const columns: DataTableColumns<UserDTO> = [
  { title: '用户名', key: 'name', width: 140 },
  {
    title: '流量使用',
    key: 'traffic',
    render: (row) =>
      row.traffic_limit_bytes > 0
        ? `${formatBytes(row.traffic_used_bytes)} / ${formatBytes(row.traffic_limit_bytes)}`
        : `${formatBytes(row.traffic_used_bytes)} / 无限制`
  },
  {
    title: '剩余天数',
    key: 'remaining_days',
    width: 100,
    render: (row) => (row.remaining_days < 0 ? '永久' : `${row.remaining_days} 天`)
  },
  {
    title: '状态',
    key: 'is_active',
    width: 90,
    render: (row) =>
      h(NTag, { size: 'small', type: row.is_active ? 'success' : 'error' }, () =>
        row.is_active ? '正常' : row.enabled ? '超限/过期' : '禁用'
      )
  },
  {
    title: '操作',
    key: 'actions',
    width: 310,
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(NButton, { size: 'tiny', type: 'primary', secondary: true, onClick: () => openQR(row) }, () => '订阅/二维码'),
        h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, () => '编辑'),
        h(NButton, { size: 'tiny', onClick: () => toggle(row.id) }, () => (row.enabled ? '禁用' : '启用')),
        h(NButton, { size: 'tiny', onClick: () => resetTraffic(row.id) }, () => '重置流量'),
        h(
          NPopconfirm,
          { onPositiveClick: () => remove(row.id) },
          {
            trigger: () => h(NButton, { size: 'tiny', type: 'error', quaternary: true }, () => '删除'),
            default: () => `确认删除用户 ${row.name}？`
          }
        )
      ])
  }
]

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const res = await desktopApi.listUsers(search.value)
    users.value = res.users
  } catch (err) {
    message.error('加载用户失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  Object.assign(form, { name: '', email: '', uuid: crypto.randomUUID(), note: '', wg_pub_key: '', enabled: true })
  limitGB.value = 100
  expiryTs.value = null
  showModal.value = true
}

function openEdit(u: UserDTO) {
  editingId.value = u.id
  Object.assign(form, {
    name: u.name,
    email: u.email,
    uuid: u.uuid,
    note: u.note,
    wg_pub_key: u.wg_pub_key,
    enabled: u.enabled
  })
  limitGB.value = Math.round(u.traffic_limit_bytes / (1024 * 1024 * 1024))
  expiryTs.value = u.expiry_date ? new Date(u.expiry_date).getTime() : null
  showModal.value = true
}

async function save() {
  saving.value = true
  try {
    const bytes = Math.max(0, Math.round(limitGB.value || 0)) * 1024 * 1024 * 1024
    const expiryISO = expiryTs.value ? new Date(expiryTs.value).toISOString() : null
    if (editingId.value) {
      await desktopApi.updateUser(editingId.value, {
        name: form.name,
        email: form.email,
        uuid: form.uuid,
        traffic_limit_bytes: bytes,
        expiry_date: expiryISO,
        enabled: form.enabled,
        note: form.note,
        wg_pub_key: form.wg_pub_key
      })
    } else {
      await desktopApi.createUser({
        name: form.name,
        email: form.email,
        uuid: form.uuid,
        traffic_limit_bytes: bytes,
        expiry_date: expiryISO,
        enabled: form.enabled,
        note: form.note,
        wg_pub_key: form.wg_pub_key
      })
    }
    showModal.value = false
    message.success('用户已保存')
    await load()
  } catch (err) {
    message.error('保存失败: ' + formatError(err))
  } finally {
    saving.value = false
  }
}

async function toggle(id: string) {
  try {
    await desktopApi.toggleUser(id)
    await load()
  } catch (err) {
    message.error(formatError(err))
  }
}

async function resetTraffic(id: string) {
  try {
    await desktopApi.resetUserTraffic(id)
    message.success('流量已重置')
    await load()
  } catch (err) {
    message.error(formatError(err))
  }
}

async function remove(id: string) {
  try {
    await desktopApi.deleteUser(id)
    message.success('用户已删除')
    await load()
  } catch (err) {
    message.error(formatError(err))
  }
}

async function openQR(u: UserDTO) {
  qrUser.value = u
  showQRModal.value = true
  qrDataUrl.value = await QRCode.toDataURL(u.sub_urls.base64, { width: 220, margin: 1 })
}

function randomUUID() {
  form.uuid = crypto.randomUUID()
}

async function copyUrl(url: string) {
  await copyToClipboard(url)
  message.success('订阅链接已复制')
}

onMounted(load)
</script>
