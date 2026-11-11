<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-xl font-semibold">节点连接管理</h2>
        <p class="text-xs text-gray-400 mt-0.5">
          敏感凭据（API Token、SSH 密码与私钥口令）均保存在系统 Keyring 中，本地配置文件仅存连接元数据。
        </p>
      </div>
      <n-button type="primary" @click="openCreate">新增节点 Profile</n-button>
    </div>

    <n-grid :cols="2" :x-gap="16" :y-gap="16" responsive="screen" item-responsive>
      <n-grid-item v-for="p in conn.profiles" :key="p.id" span="2 m:1">
        <n-card size="small" :title="p.name">
          <template #header-extra>
            <n-tag
              v-if="conn.status.connected && conn.status.profile_id === p.id"
              type="success"
              size="small"
            >
              已连接 · {{ conn.status.latency_ms }} ms
            </n-tag>
            <n-tag v-else size="small" :bordered="false">
              {{ p.mode === 'ssh_tunnel' ? 'SSH 隧道' : 'HTTPS 直连' }}
            </n-tag>
          </template>

          <div class="space-y-1.5 text-xs text-gray-300">
            <template v-if="p.mode === 'ssh_tunnel'">
              <div>SSH 目标：<span class="mono">{{ p.ssh_user }}@{{ p.ssh_host }}:{{ p.ssh_port }}</span></div>
              <div>远端服务：<span class="mono">{{ p.remote_bind_host }}:{{ p.remote_panel_port }}</span>（认证：{{ p.ssh_auth_type }}）</div>
              <div v-if="p.ssh_fingerprint" class="truncate">
                TOFU 指纹：<span class="mono text-gray-400">{{ p.ssh_fingerprint }}</span>
              </div>
            </template>
            <template v-else>
              <div>直连地址：<span class="mono">{{ p.direct_url }}</span></div>
            </template>
            <div class="flex gap-2 pt-1">
              <n-tag size="tiny" :type="p.has_api_token ? 'success' : 'warning'">
                {{ p.has_api_token ? 'API Token 已保存' : '缺少 Token' }}
              </n-tag>
            </div>
          </div>

          <template #action>
            <div class="flex justify-end gap-2">
              <n-button
                size="small"
                type="primary"
                :loading="conn.connecting"
                @click="connectNode(p.id)"
              >
                连接
              </n-button>
              <n-button size="small" secondary @click="openEdit(p)">编辑</n-button>
              <n-popconfirm @positive-click="removeNode(p.id)">
                <template #trigger>
                  <n-button size="small" type="error" quaternary>删除</n-button>
                </template>
                确定删除节点「{{ p.name }}」及其本地凭据？
              </n-popconfirm>
            </div>
          </template>
        </n-card>
      </n-grid-item>
    </n-grid>

    <n-empty v-if="conn.profiles.length === 0" description="暂无服务器节点，点击右上角新增" />

    <n-modal v-model:show="showModal" preset="card" :title="editingId ? '编辑节点 Profile' : '新增节点 Profile'" class="w-[620px]">
      <n-form label-placement="left" label-width="130" size="small">
        <n-form-item label="节点名称" required>
          <n-input v-model:value="form.name" placeholder="例如：东京中转机 / 美西落地机" />
        </n-form-item>

        <n-form-item label="连接模式">
          <n-radio-group v-model:value="form.mode">
            <n-radio-button value="ssh_tunnel">内置 SSH 隧道（推荐）</n-radio-button>
            <n-radio-button value="direct">HTTPS / 本地直连</n-radio-button>
          </n-radio-group>
        </n-form-item>

        <template v-if="form.mode === 'ssh_tunnel'">
          <n-form-item label="SSH 主机" required>
            <n-input v-model:value="form.ssh_host" placeholder="VPS 公网 IP 或域名" />
          </n-form-item>
          <n-form-item label="SSH 端口 / 用户">
            <div class="flex gap-2 w-full">
              <n-input-number v-model:value="form.ssh_port" :min="1" :max="65535" class="w-32" />
              <n-input v-model:value="form.ssh_user" placeholder="root" class="flex-1" />
            </div>
          </n-form-item>
          <n-form-item label="SSH 认证方式">
            <n-select
              v-model:value="form.ssh_auth_type"
              :options="[
                { label: '本地私钥文件 (~/.ssh/id_ed25519)', value: 'key_file' },
                { label: '粘贴私钥 PEM 内容', value: 'key_pem' },
                { label: 'SSH 密码登录', value: 'password' },
                { label: '系统 SSH Agent (SSH_AUTH_SOCK)', value: 'agent' }
              ]"
            />
          </n-form-item>
          <n-form-item v-if="form.ssh_auth_type === 'key_file'" label="私钥文件路径">
            <n-input v-model:value="form.ssh_key_path" placeholder="~/.ssh/id_ed25519（留空自动探测）" />
          </n-form-item>
          <n-form-item v-if="form.ssh_auth_type === 'key_pem'" label="私钥 PEM">
            <n-input
              v-model:value="secrets.ssh_key_pem"
              type="textarea"
              :rows="3"
              :placeholder="form.has_ssh_key_pem ? '已保存（留空保持不变）' : '-----BEGIN OPENSSH PRIVATE KEY-----'"
            />
          </n-form-item>
          <n-form-item v-if="form.ssh_auth_type === 'key_file' || form.ssh_auth_type === 'key_pem'" label="私钥口令 (可选)">
            <n-input
              v-model:value="secrets.ssh_key_passphrase"
              type="password"
              show-password-on="click"
              :placeholder="form.has_ssh_key_passphrase ? '已保存（留空保持不变）' : '无口令请留空'"
            />
          </n-form-item>
          <n-form-item v-if="form.ssh_auth_type === 'password'" label="SSH 密码">
            <n-input
              v-model:value="secrets.ssh_password"
              type="password"
              show-password-on="click"
              :placeholder="form.has_ssh_password ? '已保存（留空保持不变）' : '输入 SSH 登录密码'"
            />
          </n-form-item>
          <n-form-item label="远端监听地址">
            <div class="flex gap-2 w-full">
              <n-input v-model:value="form.remote_bind_host" placeholder="127.0.0.1" class="flex-1" />
              <n-input-number v-model:value="form.remote_panel_port" :min="1" :max="65535" class="w-32" />
            </div>
          </n-form-item>
          <n-form-item v-if="form.ssh_fingerprint" label="TOFU 指纹">
            <div class="flex items-center gap-2 w-full">
              <n-input :value="form.ssh_fingerprint" readonly class="mono" />
              <n-button size="small" @click="form.ssh_fingerprint = ''">重置</n-button>
            </div>
          </n-form-item>
        </template>

        <template v-else>
          <n-form-item label="面板服务 URL" required>
            <n-input v-model:value="form.direct_url" placeholder="https://panel.example.com" />
          </n-form-item>
          <n-form-item label="允许明文 HTTP">
            <n-switch v-model:value="form.allow_insecure_http" />
            <span class="text-xs text-gray-400 ml-2">仅建议测试环境开启，公网推荐 HTTPS 或 SSH 隧道</span>
          </n-form-item>
        </template>

        <n-form-item label="API Bearer Token" :required="!form.has_api_token">
          <n-input
            v-model:value="secrets.api_token"
            type="password"
            show-password-on="click"
            :placeholder="form.has_api_token ? '已保存在 Keyring（留空保持不变）' : 'xrp_...'"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="save">保存</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import {
  desktopApi,
  formatError,
  type ProfileSecrets,
  type ServerProfile
} from '@/api/client'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const showModal = ref(false)
const saving = ref(false)
const editingId = ref('')

const form = reactive<Partial<ServerProfile>>({
  name: '',
  mode: 'ssh_tunnel',
  direct_url: '',
  allow_insecure_http: false,
  ssh_host: '',
  ssh_port: 22,
  ssh_user: 'root',
  ssh_auth_type: 'key_file',
  ssh_key_path: '',
  ssh_fingerprint: '',
  remote_bind_host: '127.0.0.1',
  remote_panel_port: 8082,
  has_api_token: false,
  has_ssh_password: false,
  has_ssh_key_pem: false,
  has_ssh_key_passphrase: false
})

const secrets = reactive<ProfileSecrets>({
  api_token: '',
  ssh_password: '',
  ssh_key_pem: '',
  ssh_key_passphrase: ''
})

function resetForm() {
  editingId.value = ''
  Object.assign(form, {
    id: '',
    name: '',
    mode: 'ssh_tunnel',
    direct_url: '',
    allow_insecure_http: false,
    ssh_host: '',
    ssh_port: 22,
    ssh_user: 'root',
    ssh_auth_type: 'key_file',
    ssh_key_path: '',
    ssh_fingerprint: '',
    remote_bind_host: '127.0.0.1',
    remote_panel_port: 8082,
    has_api_token: false,
    has_ssh_password: false,
    has_ssh_key_pem: false,
    has_ssh_key_passphrase: false
  })
  Object.assign(secrets, {
    api_token: '',
    ssh_password: '',
    ssh_key_pem: '',
    ssh_key_passphrase: ''
  })
}

function openCreate() {
  resetForm()
  showModal.value = true
}

function openEdit(p: ServerProfile) {
  resetForm()
  editingId.value = p.id
  Object.assign(form, { ...p })
  showModal.value = true
}

async function save() {
  saving.value = true
  try {
    await desktopApi.saveProfile({ ...form }, { ...secrets })
    await conn.loadProfiles()
    showModal.value = false
    message.success('节点配置已保存')
  } catch (err) {
    message.error('保存失败: ' + formatError(err))
  } finally {
    saving.value = false
  }
}

async function removeNode(id: string) {
  try {
    await desktopApi.deleteProfile(id)
    await conn.loadProfiles()
    await conn.syncStatus()
    message.success('已删除节点')
  } catch (err) {
    message.error('删除失败: ' + formatError(err))
  }
}

async function connectNode(id: string) {
  try {
    const st = await conn.connect(id, false)
    if (st) {
      message.success(`已连接 ${st.profile_name} (${st.latency_ms} ms)`)
    }
  } catch (err) {
    message.error('连接失败: ' + formatError(err))
  }
}
</script>
