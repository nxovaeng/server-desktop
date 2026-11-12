<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between flex-wrap gap-2">
      <h2 class="text-xl font-semibold">路由规则</h2>
      <div class="flex gap-2 flex-wrap">
        <n-button size="small" @click="load" :disabled="!conn.status.connected">刷新</n-button>
        <n-button size="small" secondary :disabled="!conn.status.connected" @click="refreshGeoData" :loading="geoLoading">
          刷新 GeoData
        </n-button>
        <n-button size="small" type="primary" :disabled="!conn.status.connected" @click="openCreate">
          新增规则
        </n-button>
      </div>
    </div>

    <n-card size="small" :bordered="true">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-2 flex-wrap">
          <span class="text-xs text-muted">快速导入预设：</span>
          <n-button
            size="tiny"
            secondary
            type="info"
            :disabled="!conn.status.connected"
            :loading="importingPreset === 'china-direct'"
            @click="importPreset('china-direct')"
          >
            国内直连 (china-direct)
          </n-button>
          <n-button
            size="tiny"
            secondary
            type="warning"
            :disabled="!conn.status.connected"
            :loading="importingPreset === 'warp-china'"
            @click="importPreset('warp-china')"
          >
            WARP 代理中国 (warp-china)
          </n-button>
          <n-button
            size="tiny"
            secondary
            type="warning"
            :disabled="!conn.status.connected"
            :loading="importingPreset === 'warp-streaming'"
            @click="importPreset('warp-streaming')"
          >
            WARP 流媒体解锁 (warp-streaming)
          </n-button>
        </div>

        <div class="flex items-center gap-2 text-xs">
          <n-tag size="tiny" :type="geoData?.geosite_available ? 'success' : 'default'">
            GeoSite: {{ geoData?.geosite_available ? `${geoData.geosite_tags.length} 标签` : '未加载' }}
          </n-tag>
          <n-tag size="tiny" :type="geoData?.geoip_available ? 'success' : 'default'">
            GeoIP: {{ geoData?.geoip_available ? `${geoData.geoip_codes.length} 代码` : '未加载' }}
          </n-tag>
        </div>
      </div>
    </n-card>

    <n-data-table :columns="columns" :data="sortedRules" :loading="loading" size="small" />

    <n-modal v-model:show="showModal" preset="card" :title="editingId ? '编辑路由规则' : '新增路由规则'" class="w-[680px]">
      <n-form label-placement="left" label-width="115" size="small">
        <n-form-item label="规则名称" required>
          <n-input v-model:value="form.name" placeholder="例如：国内网站直连" />
        </n-form-item>

        <div class="grid grid-cols-2 gap-3">
          <n-form-item label="匹配类型" required>
            <n-select v-model:value="form.type" :options="ruleTypeOptions" />
          </n-form-item>
          <n-form-item label="优先级">
            <n-input-number v-model:value="form.priority" :min="1" :max="9999" class="w-full" />
          </n-form-item>
        </div>

        <n-form-item v-if="form.type === 'inbound'" label="来源入站标签" required>
          <n-select
            v-model:value="form.inbound_tag"
            :options="inboundOptions"
            filterable
            tag
            clearable
            placeholder="选择入站 Tag 或手动输入"
          />
        </n-form-item>

        <n-form-item v-else-if="form.type === 'domain'" label="匹配域名" required>
          <div class="w-full space-y-1">
            <n-input
              v-model:value="form.domains"
              type="textarea"
              :rows="2"
              placeholder="逗号分隔，如: domain:google.com,full:example.com,keyword:ads"
            />
            <div class="text-xs text-muted">支持 Xray 域名语法：domain: / full: / keyword: / regexp:，多个用英文逗号分隔</div>
          </div>
        </n-form-item>

        <n-form-item v-else-if="form.type === 'ip'" label="匹配 IP / CIDR" required>
          <div class="w-full space-y-1">
            <n-input
              v-model:value="form.ips"
              type="textarea"
              :rows="2"
              placeholder="逗号分隔，如: 192.168.0.0/16,1.1.1.1"
            />
            <div class="text-xs text-muted">支持单 IP 与 CIDR 网段，多个用英文逗号分隔</div>
          </div>
        </n-form-item>

        <template v-else-if="form.type === 'geosite'">
          <n-form-item label="GeoSite 标签" required>
            <div class="w-full space-y-2">
              <n-input
                v-model:value="form.geosite_tags"
                placeholder="逗号分隔，如: cn,geolocation-cn,category-ads-all"
              />
              <div class="p-2.5 rounded surface-code border border-subtle space-y-2">
                <div class="flex items-center justify-between gap-2">
                  <n-input
                    v-model:value="geoFilter"
                    size="tiny"
                    clearable
                    placeholder="搜索 GeoSite 标签（点击添加到上方）..."
                    class="flex-1"
                  />
                  <span class="text-xs text-muted shrink-0">
                    匹配 {{ filteredGeoSiteTags.length }} / {{ availableGeoSiteTags.length }}
                  </span>
                </div>
                <div class="max-h-36 overflow-y-auto flex flex-wrap gap-1.5">
                  <n-tag
                    v-for="tag in filteredGeoSiteTags.slice(0, 120)"
                    :key="tag"
                    size="tiny"
                    checkable
                    :checked="isCSVItemSelected(form.geosite_tags, tag)"
                    @update:checked="() => toggleCSVItem('geosite_tags', tag)"
                    class="cursor-pointer mono"
                  >
                    {{ tag }}
                  </n-tag>
                  <span v-if="filteredGeoSiteTags.length > 120" class="text-xs text-muted self-center">
                    仅展示前 120 条，请输入关键词缩小范围
                  </span>
                </div>
              </div>
            </div>
          </n-form-item>
        </template>

        <template v-else-if="form.type === 'geoip'">
          <n-form-item label="GeoIP 代码" required>
            <div class="w-full space-y-2">
              <n-input v-model:value="form.geoip_codes" placeholder="逗号分隔，如: cn,private,us" />
              <div class="p-2.5 rounded surface-code border border-subtle space-y-2">
                <div class="flex items-center justify-between gap-2">
                  <n-input
                    v-model:value="geoFilter"
                    size="tiny"
                    clearable
                    placeholder="搜索 GeoIP 国家/区域代码（点击添加到上方）..."
                    class="flex-1"
                  />
                  <span class="text-xs text-muted shrink-0">
                    匹配 {{ filteredGeoIPCodes.length }} / {{ availableGeoIPCodes.length }}
                  </span>
                </div>
                <div class="max-h-36 overflow-y-auto flex flex-wrap gap-1.5">
                  <n-tag
                    v-for="code in filteredGeoIPCodes.slice(0, 120)"
                    :key="code"
                    size="tiny"
                    checkable
                    :checked="isCSVItemSelected(form.geoip_codes, code)"
                    @update:checked="() => toggleCSVItem('geoip_codes', code)"
                    class="cursor-pointer mono"
                  >
                    {{ code }}
                  </n-tag>
                </div>
              </div>
            </div>
          </n-form-item>
        </template>

        <n-form-item v-else-if="form.type === 'protocol'" label="应用层协议" required>
          <div class="w-full space-y-2">
            <n-input v-model:value="form.protocols" placeholder="逗号分隔，如: bittorrent,http,tls" />
            <div class="flex flex-wrap gap-1.5">
              <n-tag
                v-for="proto in ['bittorrent', 'http', 'tls', 'quic']"
                :key="proto"
                size="tiny"
                checkable
                :checked="isCSVItemSelected(form.protocols, proto)"
                @update:checked="() => toggleCSVItem('protocols', proto)"
                class="cursor-pointer mono"
              >
                {{ proto }}
              </n-tag>
            </div>
          </div>
        </n-form-item>

        <n-form-item label="目标出站 Tag" required>
          <n-select
            v-model:value="form.outbound_tag"
            :options="outboundOptions"
            filterable
            tag
            placeholder="选择目标出站（如 direct / block / warp）"
          />
        </n-form-item>

        <n-form-item label="启用规则">
          <n-switch v-model:value="form.enabled" />
        </n-form-item>

        <n-form-item label="备注">
          <n-input v-model:value="form.remark" placeholder="可选说明" />
        </n-form-item>
      </n-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <n-button size="small" @click="showModal = false">取消</n-button>
          <n-button size="small" type="primary" :loading="saving" @click="save">保存</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, watch } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, useMessage, type DataTableColumns } from 'naive-ui'
import { desktopApi, formatError } from '@/api/client'
import type { GeoDataInfo, InboundDTO, OutboundDTO, RoutingRuleDTO, RuleType } from '@/api/types'
import { useConnectionStore } from '@/stores/connection'

const conn = useConnectionStore()
const message = useMessage()

const rules = ref<RoutingRuleDTO[]>([])
const inbounds = ref<InboundDTO[]>([])
const outbounds = ref<OutboundDTO[]>([])
const geoData = ref<GeoDataInfo | null>(null)
const loading = ref(false)
const geoLoading = ref(false)
const saving = ref(false)
const importingPreset = ref('')
const showModal = ref(false)
const editingId = ref('')
const geoFilter = ref('')

const fallbackGeoSiteTags = [
  'cn',
  'geolocation-cn',
  'geolocation-!cn',
  'category-ads-all',
  'category-porn',
  'google',
  'youtube',
  'netflix',
  'disney',
  'openai',
  'anthropic',
  'github',
  'telegram',
  'apple',
  'microsoft',
  'spotify',
  'tiktok'
]

const fallbackGeoIPCodes = ['cn', 'private', 'us', 'jp', 'sg', 'hk', 'tw', 'kr', 'de', 'gb', 'cloudflare']

const ruleTypeOptions: Array<{ label: string; value: RuleType }> = [
  { label: 'GeoSite 域名集合', value: 'geosite' },
  { label: 'GeoIP 国家/网段', value: 'geoip' },
  { label: '自定义域名 (Domain)', value: 'domain' },
  { label: '自定义 IP / CIDR', value: 'ip' },
  { label: '来源入站 (Inbound)', value: 'inbound' },
  { label: '嗅探协议 (Protocol)', value: 'protocol' }
]

const ruleTypeLabels: Record<RuleType, string> = {
  inbound: '入站',
  domain: '域名',
  ip: 'IP',
  geosite: 'GeoSite',
  geoip: 'GeoIP',
  protocol: '协议'
}

const form = reactive({
  name: '',
  type: 'geosite' as RuleType,
  inbound_tag: '',
  domains: '',
  ips: '',
  geosite_tags: '',
  geoip_codes: '',
  protocols: '',
  outbound_tag: 'direct',
  priority: 100,
  enabled: true,
  remark: ''
})

const sortedRules = computed(() =>
  [...rules.value].sort((a, b) => a.priority - b.priority)
)

const inboundOptions = computed(() =>
  inbounds.value.map((ib) => ({
    label: `${ib.tag} (${ib.protocol.toUpperCase()})`,
    value: ib.tag
  }))
)

const outboundOptions = computed(() => {
  const seen = new Set<string>()
  const opts: Array<{ label: string; value: string }> = []
  const defaults: Array<{ label: string; value: string }> = [
    { label: 'direct (直连出站)', value: 'direct' },
    { label: 'block (黑洞拦截)', value: 'block' }
  ]
  for (const d of defaults) {
    seen.add(d.value)
    opts.push(d)
  }
  for (const ob of outbounds.value) {
    if (!seen.has(ob.tag)) {
      seen.add(ob.tag)
      opts.push({ label: `${ob.tag} (${ob.type})`, value: ob.tag })
    }
  }
  return opts
})

const availableGeoSiteTags = computed(() =>
  geoData.value?.geosite_tags && geoData.value.geosite_tags.length > 0
    ? geoData.value.geosite_tags
    : fallbackGeoSiteTags
)

const availableGeoIPCodes = computed(() =>
  geoData.value?.geoip_codes && geoData.value.geoip_codes.length > 0
    ? geoData.value.geoip_codes
    : fallbackGeoIPCodes
)

const filteredGeoSiteTags = computed(() => {
  const q = geoFilter.value.trim().toLowerCase()
  if (!q) return availableGeoSiteTags.value
  return availableGeoSiteTags.value.filter((t) => t.toLowerCase().includes(q))
})

const filteredGeoIPCodes = computed(() => {
  const q = geoFilter.value.trim().toLowerCase()
  if (!q) return availableGeoIPCodes.value
  return availableGeoIPCodes.value.filter((c) => c.toLowerCase().includes(q))
})

function summarizeRuleTarget(r: RoutingRuleDTO): string {
  switch (r.type) {
    case 'inbound':
      return r.inbound_tag || '-'
    case 'domain':
      return r.domains || '-'
    case 'ip':
      return r.ips || '-'
    case 'geosite':
      return r.geosite_tags || '-'
    case 'geoip':
      return r.geoip_codes || '-'
    case 'protocol':
      return r.protocols || '-'
    default:
      return '-'
  }
}

const columns: DataTableColumns<RoutingRuleDTO> = [
  { title: '优先级', key: 'priority', width: 75 },
  { title: '规则名称', key: 'name', width: 170 },
  {
    title: '类型',
    key: 'type',
    width: 90,
    render: (r) => h(NTag, { size: 'small', bordered: false }, () => ruleTypeLabels[r.type] || r.type)
  },
  {
    title: '匹配条件',
    key: 'target',
    ellipsis: { tooltip: true },
    render: (r) => h('span', { class: 'mono text-xs' }, summarizeRuleTarget(r))
  },
  {
    title: '目标出站',
    key: 'outbound_tag',
    width: 130,
    render: (r) => {
      const t = r.outbound_tag === 'block' ? 'error' : r.outbound_tag === 'direct' ? 'success' : 'info'
      return h(NTag, { size: 'small', type: t }, () => r.outbound_tag)
    }
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
        h(NButton, { size: 'tiny', onClick: () => toggle(r.id) }, () => (r.enabled ? '停用' : '启用')),
        h(
          NPopconfirm,
          { onPositiveClick: () => remove(r.id) },
          {
            trigger: () => h(NButton, { size: 'tiny', type: 'error', quaternary: true }, () => '删除'),
            default: () => `确认删除路由规则 ${r.name}？`
          }
        )
      ])
  }
]

function splitCSVList(raw: string): string[] {
  return raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
}

function isCSVItemSelected(raw: string, item: string): boolean {
  return splitCSVList(raw).includes(item)
}

function toggleCSVItem(field: 'geosite_tags' | 'geoip_codes' | 'protocols', item: string) {
  const list = splitCSVList(form[field])
  const idx = list.indexOf(item)
  if (idx >= 0) {
    list.splice(idx, 1)
  } else {
    list.push(item)
  }
  form[field] = list.join(',')
}

async function load() {
  if (!conn.status.connected) return
  loading.value = true
  try {
    const [ruleRes, ibRes, obRes, geo] = await Promise.all([
      desktopApi.listRouting(),
      desktopApi.listInbounds(),
      desktopApi.listOutbounds(),
      conn.getGeoData(false)
    ])
    rules.value = ruleRes.rules
    inbounds.value = ibRes.inbounds
    outbounds.value = obRes.outbounds
    geoData.value = geo
  } catch (err) {
    message.error('加载路由规则失败: ' + formatError(err))
  } finally {
    loading.value = false
  }
}

async function refreshGeoData() {
  if (!conn.status.connected) return
  geoLoading.value = true
  try {
    geoData.value = await conn.getGeoData(true)
    message.success('GeoData 列表已刷新')
  } catch (err) {
    message.error('刷新 GeoData 失败: ' + formatError(err))
  } finally {
    geoLoading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  geoFilter.value = ''
  Object.assign(form, {
    name: '',
    type: 'geosite' as RuleType,
    inbound_tag: '',
    domains: '',
    ips: '',
    geosite_tags: '',
    geoip_codes: '',
    protocols: '',
    outbound_tag: 'direct',
    priority: 100,
    enabled: true,
    remark: ''
  })
  showModal.value = true
}

function openEdit(r: RoutingRuleDTO) {
  editingId.value = r.id
  geoFilter.value = ''
  Object.assign(form, {
    name: r.name,
    type: r.type,
    inbound_tag: r.inbound_tag,
    domains: r.domains,
    ips: r.ips,
    geosite_tags: r.geosite_tags,
    geoip_codes: r.geoip_codes,
    protocols: r.protocols,
    outbound_tag: r.outbound_tag,
    priority: r.priority,
    enabled: r.enabled,
    remark: r.remark
  })
  showModal.value = true
}

async function save() {
  if (!form.name.trim() || !form.outbound_tag.trim()) {
    message.warning('请填写规则名称与目标出站')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      type: form.type,
      inbound_tag: form.type === 'inbound' ? form.inbound_tag.trim() : '',
      domains: form.type === 'domain' ? form.domains.trim() : '',
      ips: form.type === 'ip' ? form.ips.trim() : '',
      geosite_tags: form.type === 'geosite' ? form.geosite_tags.trim() : '',
      geoip_codes: form.type === 'geoip' ? form.geoip_codes.trim() : '',
      protocols: form.type === 'protocol' ? form.protocols.trim() : '',
      outbound_tag: form.outbound_tag.trim(),
      priority: form.priority,
      enabled: form.enabled,
      remark: form.remark.trim()
    }
    if (editingId.value) {
      await desktopApi.updateRouting(editingId.value, payload)
    } else {
      await desktopApi.createRouting(payload)
    }
    showModal.value = false
    message.success('路由规则已保存')
    await load()
  } catch (err) {
    message.error('保存失败: ' + formatError(err))
  } finally {
    saving.value = false
  }
}

async function toggle(id: string) {
  try {
    await desktopApi.toggleRouting(id)
    await load()
  } catch (err) {
    message.error('切换状态失败: ' + formatError(err))
  }
}

async function remove(id: string) {
  try {
    await desktopApi.deleteRouting(id)
    message.success('规则已删除')
    await load()
  } catch (err) {
    message.error('删除失败: ' + formatError(err))
  }
}

async function importPreset(preset: 'china-direct' | 'warp-china' | 'warp-streaming') {
  importingPreset.value = preset
  try {
    const res = await desktopApi.importRoutingPreset(preset)
    message.success(`预设 ${res.preset}：新增 ${res.imported_count} 条，跳过已存在 ${res.skipped_count} 条`)
    rules.value = res.rules
  } catch (err) {
    message.error('导入预设失败: ' + formatError(err))
  } finally {
    importingPreset.value = ''
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
