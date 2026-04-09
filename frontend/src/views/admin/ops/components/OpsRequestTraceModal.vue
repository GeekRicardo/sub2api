<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useClipboard } from '@/composables/useClipboard'
import { opsAPI, type OpsErrorDetail, type OpsRequestDetail, type OpsSuccessRequestDetail } from '@/api/admin/ops'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import { resolvePrimaryResponseBody } from '../utils/errorDetailResponse'

type TabKey = 'overview' | 'request' | 'response'

const props = defineProps<{
  show: boolean
  row: OpsRequestDetail | null
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const activeTab = ref<TabKey>('overview')
const successDetail = ref<OpsSuccessRequestDetail | null>(null)
const errorDetail = ref<OpsErrorDetail | null>(null)
const tabs: Array<{ key: TabKey; label: string }> = [
  { key: 'overview', label: '概览' },
  { key: 'request', label: '请求参数' },
  { key: 'response', label: '返回值' }
]

const rowKind = computed(() => props.row?.kind ?? 'success')
const detailTitle = computed(() => {
  if (!props.row) return '请求详情'
  const prefix = props.row.kind === 'error' ? '错误请求详情' : '成功请求详情'
  return props.row.request_id ? `${prefix} · ${props.row.request_id}` : prefix
})

const isSuccess = computed(() => rowKind.value === 'success')

const requestBody = computed(() => {
  if (isSuccess.value) return successDetail.value?.request_body || ''
  return errorDetail.value?.request_body || ''
})

const requestBodyTruncated = computed(() => {
  if (isSuccess.value) return Boolean(successDetail.value?.request_body_truncated)
  return Boolean(errorDetail.value?.request_body_truncated)
})

const requestBodyBytes = computed(() => {
  if (isSuccess.value) return successDetail.value?.request_body_bytes ?? null
  return errorDetail.value?.request_body_bytes ?? null
})

const responseBody = computed(() => {
  if (isSuccess.value) return successDetail.value?.response_body || ''
  const errorType = String(errorDetail.value?.phase || '').toLowerCase() === 'upstream' ? 'upstream' : 'request'
  return resolvePrimaryResponseBody(errorDetail.value, errorType)
})

const responseBodyTruncated = computed(() => {
  if (isSuccess.value) return Boolean(successDetail.value?.response_body_truncated)
  return false
})

const responseBodyBytes = computed(() => {
  if (isSuccess.value) return successDetail.value?.response_body_bytes ?? null
  return null
})

const statusLabel = computed(() => {
  if (isSuccess.value) return '200'
  return String(errorDetail.value?.status_code ?? props.row?.status_code ?? '—')
})

const statusClass = computed(() => {
  if (isSuccess.value) return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  const code = errorDetail.value?.status_code ?? 0
  if (code >= 500) return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (code >= 400) return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200'
})

const summaryItems = computed(() => {
  if (isSuccess.value) {
    const detail = successDetail.value
    if (!detail) return []
    return [
      { label: '时间', value: formatDateTime(detail.created_at) },
      { label: '平台', value: detail.platform || 'unknown' },
      { label: '模型', value: renderModel(detail.requested_model, detail.upstream_model, detail.model) },
      { label: '耗时', value: formatDuration(detail.duration_ms) },
      { label: '端点', value: detail.inbound_endpoint || '—' },
      { label: '上游端点', value: detail.upstream_endpoint || '—' },
      { label: '用户', value: detail.user_email || asText(detail.user_id) },
      { label: '账号', value: detail.account_name || asText(detail.account_id) },
      { label: '分组', value: detail.group_name || asText(detail.group_id) },
      { label: '请求类型', value: formatRequestType(detail.request_type) },
      { label: '流式', value: detail.stream ? '是' : '否' },
      { label: 'User-Agent', value: detail.user_agent || '—' }
    ]
  }

  const detail = errorDetail.value
  if (!detail) return []
  return [
    { label: '时间', value: formatDateTime(detail.created_at) },
    { label: '平台', value: detail.platform || 'unknown' },
    { label: '模型', value: renderModel(detail.requested_model, detail.upstream_model, detail.model) },
    { label: '耗时', value: formatDuration(props.row?.duration_ms) },
    { label: '端点', value: detail.inbound_endpoint || detail.request_path || '—' },
    { label: '上游端点', value: detail.upstream_endpoint || '—' },
    { label: '用户', value: detail.user_email || asText(detail.user_id) },
    { label: '账号', value: detail.account_name || asText(detail.account_id) },
    { label: '分组', value: detail.group_name || asText(detail.group_id) },
    { label: '请求类型', value: formatRequestType(detail.request_type) },
    { label: '阶段', value: detail.phase || '—' },
    { label: '错误信息', value: detail.message || '—' }
  ]
})

const usageItems = computed(() => {
  if (!isSuccess.value || !successDetail.value) return []
  const detail = successDetail.value
  return [
    { label: '输入 Tokens', value: String(detail.input_tokens ?? 0) },
    { label: '输出 Tokens', value: String(detail.output_tokens ?? 0) },
    { label: '缓存写入', value: String(detail.cache_creation_tokens ?? 0) },
    { label: '缓存读取', value: String(detail.cache_read_tokens ?? 0) },
    { label: '总成本', value: Number(detail.total_cost ?? 0).toFixed(4) },
    { label: '实际成本', value: Number(detail.actual_cost ?? 0).toFixed(4) }
  ]
})

function asText(value?: number | null): string {
  return value == null ? '—' : String(value)
}

function renderModel(requested?: string | null, upstream?: string | null, model?: string | null): string {
  const requestedModel = String(requested || model || '').trim()
  const upstreamModel = String(upstream || '').trim()
  if (requestedModel && upstreamModel && requestedModel !== upstreamModel) {
    return `${requestedModel} → ${upstreamModel}`
  }
  return requestedModel || upstreamModel || '—'
}

function formatDuration(value?: number | null): string {
  return typeof value === 'number' ? `${value} ms` : '—'
}

function formatRequestType(value?: number | null): string {
  switch (value) {
    case 1: return 'sync'
    case 2: return 'stream'
    case 3: return 'ws'
    default: return 'unknown'
  }
}

function close() {
  emit('update:show', false)
}

function resetState() {
  successDetail.value = null
  errorDetail.value = null
  activeTab.value = 'overview'
}

function prettyJSON(raw?: string): string {
  const text = String(raw || '').trim()
  if (!text) return '暂无数据'
  try {
    return JSON.stringify(JSON.parse(text), null, 2)
  } catch {
    return text
  }
}

async function copyRaw(raw: string, label: string) {
  const ok = await copyToClipboard(raw, `${label}已复制`)
  if (!ok) appStore.showWarning(t('common.copyFailed'))
}

async function fetchDetail() {
  if (!props.show || !props.row?.detail_id) return
  loading.value = true
  try {
    if (props.row.kind === 'success') {
      successDetail.value = await opsAPI.getSuccessRequestDetail(props.row.detail_id)
      errorDetail.value = null
    } else {
      errorDetail.value = await opsAPI.getRequestErrorDetail(props.row.detail_id)
      successDetail.value = null
    }
  } catch (err: any) {
    resetState()
    appStore.showError(err?.message || '加载请求详情失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.row?.detail_id, props.row?.kind] as const,
  ([show]) => {
    if (!show) {
      resetState()
      return
    }
    fetchDetail()
  },
  { immediate: true }
)
</script>

<template>
  <BaseDialog :show="show" :title="detailTitle" width="full" @close="close">
    <div class="min-h-[72vh] bg-gradient-to-b from-slate-50 to-white p-6 dark:from-dark-950 dark:to-dark-900">
      <div v-if="loading" class="flex h-[60vh] items-center justify-center">
        <div class="flex flex-col items-center gap-3">
          <div class="h-9 w-9 animate-spin rounded-full border-b-2 border-primary-600" />
          <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
        </div>
      </div>

      <div v-else-if="!row" class="flex h-[60vh] items-center justify-center text-sm text-gray-500 dark:text-gray-400">
        暂无请求详情
      </div>

      <div v-else class="space-y-6">
        <div class="overflow-hidden rounded-3xl bg-white/90 shadow-sm ring-1 ring-black/5 backdrop-blur dark:bg-dark-800/90 dark:ring-white/10">
          <div class="flex flex-wrap items-start justify-between gap-4 border-b border-gray-100 px-6 py-5 dark:border-dark-700">
            <div class="space-y-2">
              <div class="flex flex-wrap items-center gap-2">
                <span class="rounded-full px-3 py-1 text-xs font-black uppercase tracking-wide" :class="row.kind === 'success' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'">
                  {{ row.kind === 'success' ? 'Success' : 'Error' }}
                </span>
                <span class="rounded-full px-3 py-1 text-xs font-black" :class="statusClass">
                  HTTP {{ statusLabel }}
                </span>
              </div>
              <div class="break-all font-mono text-sm text-gray-700 dark:text-gray-200">
                {{ row.request_id || '无 request_id' }}
              </div>
            </div>

            <div class="flex flex-wrap gap-2">
              <button class="btn btn-secondary btn-sm" @click="activeTab = 'request'">请求参数</button>
              <button class="btn btn-secondary btn-sm" @click="activeTab = 'response'">返回值</button>
            </div>
          </div>

          <div class="grid gap-4 px-6 py-6 lg:grid-cols-[1.3fr,0.7fr]">
            <div class="grid gap-3 sm:grid-cols-2">
              <div
                v-for="item in summaryItems"
                :key="item.label"
                class="rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-900"
              >
                <div class="text-[11px] font-bold uppercase tracking-wide text-gray-400">{{ item.label }}</div>
                <div class="mt-1 break-words text-sm font-medium text-gray-900 dark:text-white">{{ item.value }}</div>
              </div>
            </div>

            <div class="rounded-2xl bg-slate-900 p-4 text-slate-50 shadow-inner">
              <div class="text-xs font-black uppercase tracking-widest text-slate-400">Payload 摘要</div>
              <div class="mt-3 grid gap-3 sm:grid-cols-2">
                <div class="rounded-2xl bg-white/5 p-3">
                  <div class="text-[11px] uppercase tracking-wide text-slate-400">请求体</div>
                  <div class="mt-1 text-lg font-black">{{ requestBodyBytes ?? 0 }}</div>
                  <div class="text-xs text-slate-400">bytes {{ requestBodyTruncated ? '· 已截断' : '' }}</div>
                </div>
                <div class="rounded-2xl bg-white/5 p-3">
                  <div class="text-[11px] uppercase tracking-wide text-slate-400">响应体</div>
                  <div class="mt-1 text-lg font-black">{{ responseBodyBytes ?? 0 }}</div>
                  <div class="text-xs text-slate-400">bytes {{ responseBodyTruncated ? '· 已截断' : '' }}</div>
                </div>
              </div>

              <div v-if="usageItems.length" class="mt-4 grid gap-2 sm:grid-cols-2">
                <div v-for="item in usageItems" :key="item.label" class="rounded-xl bg-white/5 px-3 py-2">
                  <div class="text-[11px] uppercase tracking-wide text-slate-400">{{ item.label }}</div>
                  <div class="mt-1 text-sm font-bold">{{ item.value }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="overflow-hidden rounded-3xl bg-white shadow-sm ring-1 ring-black/5 dark:bg-dark-800 dark:ring-white/10">
          <div class="flex flex-wrap items-center gap-2 border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              class="rounded-full px-4 py-2 text-sm font-bold transition"
              :class="activeTab === tab.key ? 'bg-primary-600 text-white shadow-sm' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
              @click="activeTab = tab.key"
            >
              {{ tab.label }}
            </button>
          </div>

          <div v-if="activeTab === 'overview'" class="grid gap-4 px-6 py-6 md:grid-cols-2">
            <div class="rounded-2xl border border-gray-100 p-4 dark:border-dark-700">
              <div class="text-xs font-black uppercase tracking-wide text-gray-400">请求预览</div>
              <pre class="mt-3 max-h-[320px] overflow-auto rounded-2xl bg-gray-50 p-4 text-xs text-gray-800 dark:bg-dark-900 dark:text-gray-100"><code>{{ prettyJSON(requestBody) }}</code></pre>
            </div>
            <div class="rounded-2xl border border-gray-100 p-4 dark:border-dark-700">
              <div class="text-xs font-black uppercase tracking-wide text-gray-400">响应预览</div>
              <pre class="mt-3 max-h-[320px] overflow-auto rounded-2xl bg-gray-50 p-4 text-xs text-gray-800 dark:bg-dark-900 dark:text-gray-100"><code>{{ prettyJSON(responseBody) }}</code></pre>
            </div>
          </div>

          <div v-else class="px-6 py-6">
            <div class="mb-3 flex items-center justify-between gap-3">
              <div>
                <div class="text-xs font-black uppercase tracking-wide text-gray-400">
                  {{ activeTab === 'request' ? '请求参数' : '返回值' }}
                </div>
                <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ activeTab === 'request'
                    ? (requestBodyTruncated ? '已做脱敏与截断存储' : '已做脱敏存储')
                    : (responseBodyTruncated ? '响应已按预览长度截断' : '客户端可见响应预览') }}
                </div>
              </div>
              <button
                class="btn btn-secondary btn-sm"
                @click="copyRaw(activeTab === 'request' ? requestBody : responseBody, activeTab === 'request' ? '请求参数' : '返回值')"
              >
                复制内容
              </button>
            </div>

            <pre class="max-h-[520px] overflow-auto rounded-3xl bg-slate-950 p-5 text-xs text-slate-100"><code>{{ prettyJSON(activeTab === 'request' ? requestBody : responseBody) }}</code></pre>
          </div>
        </div>
      </div>
    </div>
  </BaseDialog>
</template>
