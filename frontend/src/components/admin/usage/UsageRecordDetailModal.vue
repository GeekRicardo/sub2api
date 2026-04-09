<template>
  <BaseDialog :show="show" :title="title" width="ultra-wide" @close="close">
    <div class="min-h-[72vh] bg-[#f7f8fb] p-4 dark:bg-dark-950 md:p-6">
      <div v-if="loading" class="flex h-[60vh] items-center justify-center">
        <div class="flex flex-col items-center gap-3">
          <div class="h-9 w-9 animate-spin rounded-full border-b-2 border-primary-600" />
          <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
        </div>
      </div>

      <div v-else-if="detail" class="space-y-5">
        <div class="rounded-3xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <span class="inline-flex items-center rounded-full bg-primary-50 px-3 py-1 text-xs font-black uppercase tracking-wide text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">
                  Usage Record
                </span>
                <span class="inline-flex items-center rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300">
                  {{ requestTypeLabel(detail) }}
                </span>
              </div>
              <div class="mt-3 font-mono text-sm text-gray-700 dark:text-gray-200 break-all">{{ detail.request_id || `#${detail.id}` }}</div>
            </div>

            <div class="grid gap-2 text-right text-sm">
              <div class="text-gray-500 dark:text-gray-400">{{ formatDateTime(detail.created_at) }}</div>
              <div class="font-semibold text-gray-900 dark:text-white">{{ detail.user?.email || `#${detail.user_id}` }}</div>
              <div class="text-gray-500 dark:text-gray-400">{{ displayModel(detail) }}</div>
            </div>
          </div>

          <div class="mt-5 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            <div v-for="item in summaryCards" :key="item.label" class="rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-800">
              <div class="text-[11px] font-bold uppercase tracking-wide text-gray-400">{{ item.label }}</div>
              <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white break-words">{{ item.value }}</div>
            </div>
          </div>
        </div>

        <div class="grid gap-5 xl:grid-cols-2">
          <section class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-4 dark:border-dark-700">
              <div>
                <div class="text-xs font-black uppercase tracking-[0.2em] text-gray-400">请求</div>
                <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ payloadMeta(detail.request_body_bytes, detail.request_body_truncated, true) }}
                </div>
              </div>
              <button class="btn btn-secondary btn-sm" @click="copyPayload(detail.request_body || '', '请求体')">复制</button>
            </div>

            <div class="grid gap-4 p-5">
              <div class="rounded-2xl border border-gray-100 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-950">
                <div class="mb-2 text-xs font-bold uppercase tracking-wide text-gray-400">元数据</div>
                <dl class="grid gap-2 text-sm md:grid-cols-2">
                  <div>
                    <dt class="text-gray-400">入站端点</dt>
                    <dd class="mt-1 break-all font-medium text-gray-900 dark:text-white">{{ detail.inbound_endpoint || '-' }}</dd>
                  </div>
                  <div>
                    <dt class="text-gray-400">User-Agent</dt>
                    <dd class="mt-1 break-all font-medium text-gray-900 dark:text-white">{{ detail.user_agent || '-' }}</dd>
                  </div>
                  <div>
                    <dt class="text-gray-400">账号</dt>
                    <dd class="mt-1 break-all font-medium text-gray-900 dark:text-white">{{ detail.account?.name || `#${detail.account_id}` }}</dd>
                  </div>
                  <div>
                    <dt class="text-gray-400">分组</dt>
                    <dd class="mt-1 break-all font-medium text-gray-900 dark:text-white">{{ detail.group?.name || (detail.group_id != null ? `#${detail.group_id}` : '-') }}</dd>
                  </div>
                </dl>
              </div>

              <div class="overflow-hidden rounded-2xl border border-gray-100 dark:border-dark-700">
                <div class="flex items-center justify-between bg-slate-950 px-4 py-3 text-slate-100">
                  <div class="text-xs font-black uppercase tracking-[0.2em]">Request Body</div>
                  <div class="text-[11px] text-slate-400">{{ detail.request_body_truncated ? '截断预览' : '完整预览' }}</div>
                </div>
                <pre class="max-h-[480px] overflow-auto bg-[#0b1020] p-4 text-xs leading-6 text-slate-100"><code>{{ prettyJSON(detail.request_body) }}</code></pre>
              </div>
            </div>
          </section>

          <section class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-4 dark:border-dark-700">
              <div>
                <div class="text-xs font-black uppercase tracking-[0.2em] text-gray-400">返回</div>
                <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ payloadMeta(detail.response_body_bytes, detail.response_body_truncated, false) }}
                </div>
              </div>
              <button class="btn btn-secondary btn-sm" @click="copyPayload(detail.response_body || '', '响应体')">复制</button>
            </div>

            <div class="grid gap-4 p-5">
              <div class="grid gap-3 md:grid-cols-3">
                <div class="rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-800">
                  <div class="text-[11px] font-bold uppercase tracking-wide text-gray-400">耗时</div>
                  <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ formatDuration(detail.duration_ms) }}</div>
                </div>
                <div class="rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-800">
                  <div class="text-[11px] font-bold uppercase tracking-wide text-gray-400">首 Token</div>
                  <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ formatDuration(detail.first_token_ms) }}</div>
                </div>
                <div class="rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-800">
                  <div class="text-[11px] font-bold uppercase tracking-wide text-gray-400">上游端点</div>
                  <div class="mt-1 break-all text-sm font-semibold text-gray-900 dark:text-white">{{ detail.upstream_endpoint || '-' }}</div>
                </div>
              </div>

              <div class="rounded-2xl border border-gray-100 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-950">
                <div class="mb-2 text-xs font-bold uppercase tracking-wide text-gray-400">用量与计费</div>
                <dl class="grid gap-2 text-sm md:grid-cols-2">
                  <div><dt class="text-gray-400">输入 Tokens</dt><dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ number(detail.input_tokens) }}</dd></div>
                  <div><dt class="text-gray-400">输出 Tokens</dt><dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ number(detail.output_tokens) }}</dd></div>
                  <div><dt class="text-gray-400">缓存写入</dt><dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ number(detail.cache_creation_tokens) }}</dd></div>
                  <div><dt class="text-gray-400">缓存读取</dt><dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ number(detail.cache_read_tokens) }}</dd></div>
                  <div><dt class="text-gray-400">用户计费</dt><dd class="mt-1 font-medium text-emerald-600 dark:text-emerald-400">${{ money(detail.actual_cost) }}</dd></div>
                  <div><dt class="text-gray-400">原始成本</dt><dd class="mt-1 font-medium text-gray-900 dark:text-white">${{ money(detail.total_cost) }}</dd></div>
                </dl>
              </div>

              <div class="overflow-hidden rounded-2xl border border-gray-100 dark:border-dark-700">
                <div class="flex items-center justify-between bg-slate-950 px-4 py-3 text-slate-100">
                  <div class="text-xs font-black uppercase tracking-[0.2em]">Response Body</div>
                  <div class="text-[11px] text-slate-400">{{ detail.response_body_truncated ? '截断预览' : '完整预览' }}</div>
                </div>
                <pre class="max-h-[480px] overflow-auto bg-[#0b1020] p-4 text-xs leading-6 text-slate-100"><code>{{ prettyJSON(detail.response_body) }}</code></pre>
              </div>
            </div>
          </section>
        </div>
      </div>

      <div v-else class="flex h-[60vh] items-center justify-center text-sm text-gray-500 dark:text-gray-400">
        暂无详情
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminUsageAPI, type AdminUsageLogDetail } from '@/api/admin/usage'
import type { AdminUsageLog } from '@/types'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import { resolveUsageRequestType } from '@/utils/usageRequestType'

const props = defineProps<{
  show: boolean
  row: AdminUsageLog | null
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const detail = ref<AdminUsageLogDetail | null>(null)

const title = computed(() => {
  const requestId = detail.value?.request_id || props.row?.request_id
  return requestId ? `使用记录详情 · ${requestId}` : '使用记录详情'
})

const summaryCards = computed(() => {
  const d = detail.value
  if (!d) return []
  return [
    { label: '用户', value: d.user?.email || `#${d.user_id}` },
    { label: 'API Key', value: d.api_key?.name || `#${d.api_key_id}` },
    { label: '模型', value: displayModel(d) },
    { label: '时间', value: formatDateTime(d.created_at) },
    { label: '请求 ID', value: d.request_id || '-' },
    { label: 'IP', value: d.ip_address || '-' },
    { label: 'Service Tier', value: d.service_tier || '-' },
    { label: 'Reasoning', value: d.reasoning_effort || '-' }
  ]
})

function close() {
  emit('update:show', false)
}

function number(value?: number | null) {
  return Number(value || 0).toLocaleString()
}

function money(value?: number | null) {
  return Number(value || 0).toFixed(6)
}

function displayModel(d: AdminUsageLogDetail) {
  const requested = String(d.model || '').trim()
  const upstream = String(d.upstream_model || '').trim()
  if (requested && upstream && requested !== upstream) return `${requested} → ${upstream}`
  return requested || upstream || '-'
}

function requestTypeLabel(d: AdminUsageLogDetail) {
  const type = resolveUsageRequestType(d)
  if (type === 'ws_v2') return 'WS'
  if (type === 'stream') return 'Stream'
  if (type === 'sync') return 'Sync'
  return 'Unknown'
}

function formatDuration(value?: number | null) {
  if (value == null) return '-'
  if (value < 1000) return `${value} ms`
  return `${(value / 1000).toFixed(2)} s`
}

function payloadMeta(bytes?: number | null, truncated?: boolean, isRequest = true) {
  const size = typeof bytes === 'number' ? `${bytes} bytes` : '大小未知'
  if (truncated) return `${size} · 原始${isRequest ? '请求' : '响应'}内容（超长截断）`
  return `${size} · 原始${isRequest ? '请求' : '响应'}内容`
}

function prettyJSON(raw?: string | null) {
  const text = String(raw || '').trim()
  if (!text) return '暂无数据'
  try {
    return JSON.stringify(JSON.parse(text), null, 2)
  } catch {
    return text
  }
}

async function copyPayload(raw: string, label: string) {
  const ok = await copyToClipboard(raw, `${label}已复制`)
  if (!ok) appStore.showWarning('复制失败')
}

async function loadDetail() {
  if (!props.show || !props.row?.id) return
  loading.value = true
  try {
    detail.value = await adminUsageAPI.getById(props.row.id)
  } catch (error: any) {
    detail.value = null
    appStore.showError(error?.message || '加载使用记录详情失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.row?.id] as const,
  ([show]) => {
    if (!show) {
      detail.value = null
      return
    }
    loadDetail()
  },
  { immediate: true }
)
</script>
