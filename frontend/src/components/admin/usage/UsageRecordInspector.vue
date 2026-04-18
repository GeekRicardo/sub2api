<template>
  <div class="min-h-[72vh] bg-[#0b1020] text-slate-100">
    <div v-if="loading" class="flex h-[60vh] items-center justify-center">
      <div class="flex flex-col items-center gap-3">
        <div class="h-9 w-9 animate-spin rounded-full border-b-2 border-primary-400" />
        <div class="text-sm text-slate-400">{{ t('common.loading') }}</div>
      </div>
    </div>

    <div v-else-if="detail" class="flex min-h-[72vh] flex-col">
      <div class="border-b border-slate-800 bg-[#0f172a] px-5 py-4">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-full border border-primary-500/30 bg-primary-500/10 px-3 py-1 text-[11px] font-black uppercase tracking-[0.2em] text-primary-300">
                Trace Inspector
              </span>
              <span class="rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3 py-1 text-[11px] font-black uppercase tracking-[0.2em] text-emerald-300">
                {{ requestTypeLabel(detail) }}
              </span>
            </div>
            <div class="mt-3 truncate font-mono text-sm font-semibold text-slate-100">{{ detail.request_id || `#${detail.id}` }}</div>
            <div class="mt-1 text-sm text-slate-400">{{ detail.user?.email || `#${detail.user_id}` }} · {{ displayModel(detail) }}</div>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <button
              v-for="option in renderModes"
              :key="option.value"
              type="button"
              class="rounded-full border px-3 py-1.5 text-xs font-bold uppercase tracking-wide transition"
              :class="mode === option.value
                ? 'border-primary-400 bg-primary-500 text-white'
                : 'border-slate-700 bg-slate-900 text-slate-300 hover:border-slate-500 hover:text-white'"
              @click="mode = option.value"
            >
              {{ option.label }}
            </button>
          </div>
        </div>

        <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <div v-for="item in summaryCards" :key="item.label" class="rounded-2xl border border-slate-800 bg-[#111827] px-4 py-3">
            <div class="text-[11px] font-bold uppercase tracking-wide text-slate-500">{{ item.label }}</div>
            <div class="mt-1 break-words text-sm font-semibold text-slate-100">{{ item.value }}</div>
          </div>
        </div>
      </div>

      <div class="grid flex-1 gap-px bg-slate-800 xl:grid-cols-2">
        <section class="flex min-h-[64vh] flex-col bg-[#0b1020]">
          <div class="flex items-center justify-between border-b border-slate-800 px-5 py-4">
            <div>
              <div class="text-xs font-black uppercase tracking-[0.2em] text-slate-500">Request</div>
              <div class="mt-1 text-sm text-slate-400">{{ payloadMeta(detail.request_body_bytes, detail.request_body_truncated, true) }}</div>
            </div>
            <button class="rounded-lg border border-slate-700 bg-slate-900 px-3 py-1.5 text-xs font-semibold text-slate-200 transition hover:border-slate-500 hover:text-white" @click="copyPayload(detail.request_body || '', '请求体')">复制</button>
          </div>

          <div class="flex-1 overflow-auto p-5">
            <template v-if="mode === 'ai'">
              <div class="space-y-4">
                <div class="rounded-2xl border border-slate-800 bg-[#111827] p-4">
                  <div class="mb-3 text-xs font-black uppercase tracking-[0.2em] text-slate-500">Request Meta</div>
                  <dl class="grid gap-3 text-sm md:grid-cols-2">
                    <div><dt class="text-slate-500">Inbound Endpoint</dt><dd class="mt-1 break-all font-medium text-slate-100">{{ detail.inbound_endpoint || '-' }}</dd></div>
                    <div><dt class="text-slate-500">User-Agent</dt><dd class="mt-1 break-all font-medium text-slate-100">{{ detail.user_agent || '-' }}</dd></div>
                    <div><dt class="text-slate-500">Account</dt><dd class="mt-1 break-all font-medium text-slate-100">{{ detail.account?.name || `#${detail.account_id}` }}</dd></div>
                    <div><dt class="text-slate-500">Group</dt><dd class="mt-1 break-all font-medium text-slate-100">{{ detail.group?.name || (detail.group_id != null ? `#${detail.group_id}` : '-') }}</dd></div>
                  </dl>
                </div>

                <div v-if="requestMessages.length" class="space-y-3">
                  <div v-for="(message, index) in requestMessages" :key="`${message.role}-${index}`" class="rounded-2xl border px-4 py-3" :class="message.isUser ? 'border-sky-500/20 bg-sky-500/10' : 'border-violet-500/20 bg-violet-500/10'">
                    <div class="mb-2 flex items-center gap-2 text-xs font-black uppercase tracking-[0.2em]" :class="message.isUser ? 'text-sky-300' : 'text-violet-300'">
                      <span>{{ message.role }}</span>
                    </div>
                    <div class="whitespace-pre-wrap text-sm leading-7 text-slate-100">{{ message.content }}</div>
                  </div>
                </div>

                <div v-else class="rounded-2xl border border-dashed border-slate-700 p-6 text-sm text-slate-400">
                  无法识别成对话消息，已回退到 JSON 结构。
                </div>
              </div>
            </template>

            <template v-else-if="mode === 'json'">
              <UsageJsonRenderer :data="parsedRequestBody ?? detail.request_body ?? ''" />
            </template>

            <template v-else>
              <pre class="whitespace-pre-wrap break-words text-xs leading-6 text-slate-200">{{ detail.request_body || '暂无数据' }}</pre>
            </template>
          </div>
        </section>

        <section class="flex min-h-[64vh] flex-col bg-[#0d1324]">
          <div class="flex items-center justify-between border-b border-slate-800 px-5 py-4">
            <div>
              <div class="text-xs font-black uppercase tracking-[0.2em] text-slate-500">Response</div>
              <div class="mt-1 text-sm text-slate-400">{{ payloadMeta(detail.response_body_bytes, detail.response_body_truncated, false) }}</div>
            </div>
            <button class="rounded-lg border border-slate-700 bg-slate-900 px-3 py-1.5 text-xs font-semibold text-slate-200 transition hover:border-slate-500 hover:text-white" @click="copyPayload(detail.response_body || '', '响应体')">复制</button>
          </div>

          <div class="flex-1 overflow-auto p-5">
            <template v-if="mode === 'ai'">
              <div class="space-y-4">
                <div class="grid gap-3 md:grid-cols-3">
                  <div class="rounded-2xl border border-slate-800 bg-[#111827] px-4 py-3">
                    <div class="text-[11px] font-bold uppercase tracking-wide text-slate-500">Duration</div>
                    <div class="mt-1 text-sm font-semibold text-slate-100">{{ formatDuration(detail.duration_ms) }}</div>
                  </div>
                  <div class="rounded-2xl border border-slate-800 bg-[#111827] px-4 py-3">
                    <div class="text-[11px] font-bold uppercase tracking-wide text-slate-500">First Token</div>
                    <div class="mt-1 text-sm font-semibold text-slate-100">{{ formatDuration(detail.first_token_ms) }}</div>
                  </div>
                  <div class="rounded-2xl border border-slate-800 bg-[#111827] px-4 py-3">
                    <div class="text-[11px] font-bold uppercase tracking-wide text-slate-500">Upstream Endpoint</div>
                    <div class="mt-1 break-all text-sm font-semibold text-slate-100">{{ detail.upstream_endpoint || '-' }}</div>
                  </div>
                </div>

                <div v-if="responseSseEvents.length" class="space-y-3">
                  <div
                    v-for="(event, index) in responseSseEvents"
                    :key="`${event.event}-${index}`"
                    class="rounded-2xl border border-slate-800 bg-[#111827] p-4"
                  >
                    <div class="mb-3 flex items-center justify-between gap-3">
                      <div class="rounded-full bg-slate-800 px-3 py-1 text-[11px] font-black uppercase tracking-[0.2em] text-cyan-300">{{ event.event || 'message' }}</div>
                      <div class="text-[11px] text-slate-500">event {{ index + 1 }}</div>
                    </div>
                    <UsageJsonRenderer v-if="event.parsed !== undefined" :data="event.parsed" :show-copy-button="false" />
                    <pre v-else class="whitespace-pre-wrap break-words text-xs leading-6 text-slate-200">{{ event.raw }}</pre>
                  </div>
                </div>

                <div v-else-if="responseMessages.length" class="space-y-3">
                  <div v-for="(message, index) in responseMessages" :key="`${message.role}-${index}`" class="rounded-2xl border border-emerald-500/20 bg-emerald-500/10 px-4 py-3">
                    <div class="mb-2 text-xs font-black uppercase tracking-[0.2em] text-emerald-300">{{ message.role }}</div>
                    <div class="whitespace-pre-wrap text-sm leading-7 text-slate-100">{{ message.content }}</div>
                  </div>
                </div>

                <div v-else class="rounded-2xl border border-dashed border-slate-700 p-6 text-sm text-slate-400">
                  无法识别成结构化 AI 输出，已回退到 JSON / Raw 模式查看。
                </div>
              </div>
            </template>

            <template v-else-if="mode === 'json'">
              <UsageJsonRenderer :data="parsedResponseBody ?? detail.response_body ?? ''" />
            </template>

            <template v-else>
              <pre class="whitespace-pre-wrap break-words text-xs leading-6 text-slate-200">{{ detail.response_body || '暂无数据' }}</pre>
            </template>
          </div>
        </section>
      </div>
    </div>

    <div v-else class="flex h-[60vh] items-center justify-center text-sm text-slate-400">
      {{ emptyMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import UsageJsonRenderer from './UsageJsonRenderer.vue'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import { resolveUsageRequestType } from '@/utils/usageRequestType'
import type { AdminUsageLogDetail } from '@/api/admin/usage'

type RenderMode = 'ai' | 'json' | 'raw'

type ChatPreviewMessage = {
  role: string
  content: string
  isUser: boolean
}

type SseEventPreview = {
  event: string
  raw: string
  parsed?: unknown
}

const props = withDefaults(defineProps<{
  detail: AdminUsageLogDetail | null
  loading?: boolean
  emptyMessage?: string
}>(), {
  loading: false,
  emptyMessage: '从左侧选择一条请求查看详情'
})

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const mode = ref<RenderMode>('ai')

const renderModes = [
  { value: 'ai' as const, label: 'AI' },
  { value: 'json' as const, label: 'JSON' },
  { value: 'raw' as const, label: 'RAW' }
]

const summaryCards = computed(() => {
  const d = props.detail
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

const parsedRequestBody = computed(() => parseJsonSafe(props.detail?.request_body))
const parsedResponseBody = computed(() => parseJsonSafe(props.detail?.response_body))

const requestMessages = computed(() => {
  return extractRequestMessages(parsedRequestBody.value, props.detail?.request_body || '')
})

const responseMessages = computed(() => {
  return extractResponseMessages(parsedResponseBody.value, props.detail?.response_body || '')
})

const responseSseEvents = computed(() => {
  return parseSseEvents(props.detail?.response_body || '')
})

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
  const size = typeof bytes === 'number' ? `${bytes.toLocaleString()} bytes` : '大小未知'
  if (truncated) return `${size} · ${isRequest ? '请求' : '响应'}内容过长，已截断`
  return `${size} · 原始${isRequest ? '请求' : '响应'}内容`
}

async function copyPayload(raw: string, label: string) {
  const ok = await copyToClipboard(raw, `${label}已复制`)
  if (!ok) appStore.showWarning('复制失败')
}

function parseJsonSafe(raw?: string | null): unknown | undefined {
  const text = String(raw || '').trim()
  if (!text) return undefined
  try {
    return JSON.parse(text)
  } catch {
    return undefined
  }
}

function stringifyContent(value: unknown): string {
  if (value == null) return ''
  if (typeof value === 'string') return value
  if (Array.isArray(value)) {
    return value
      .map((item) => {
        if (typeof item === 'string') return item
        if (item && typeof item === 'object') {
          const record = item as Record<string, unknown>
          if (typeof record.text === 'string') return record.text
          if (typeof record.input_text === 'string') return record.input_text
          if (typeof record.type === 'string') return `[${record.type}]`
        }
        return JSON.stringify(item)
      })
      .filter(Boolean)
      .join('\n')
  }
  if (typeof value === 'object') {
    const record = value as Record<string, unknown>
    if (typeof record.text === 'string') return record.text
    if (typeof record.content === 'string') return record.content
  }
  return JSON.stringify(value, null, 2)
}

function extractRequestMessages(parsed: unknown, raw: string): ChatPreviewMessage[] {
  if (parsed && typeof parsed === 'object') {
    const record = parsed as Record<string, unknown>
    const messages = Array.isArray(record.messages) ? record.messages : Array.isArray(record.input) ? record.input : Array.isArray(record.contents) ? record.contents : []
    const output: ChatPreviewMessage[] = []

    if (typeof record.instructions === 'string' && record.instructions.trim()) {
      output.push({ role: 'developer', content: record.instructions.trim(), isUser: false })
    }

    for (const message of messages) {
      if (!message || typeof message !== 'object') continue
      const item = message as Record<string, unknown>
      const role = String(item.role || item.author || 'user')
      const content = stringifyContent(item.content ?? item.parts ?? item.input ?? item)
      if (!content.trim()) continue
      output.push({ role, content: content.trim(), isUser: ['user', 'tool'].includes(role) })
    }

    if (output.length) return output
  }

  return raw.trim() ? [{ role: 'request', content: raw.trim(), isUser: true }] : []
}

function extractResponseMessages(parsed: unknown, raw: string): ChatPreviewMessage[] {
  if (parsed && typeof parsed === 'object') {
    const record = parsed as Record<string, unknown>
    const output: ChatPreviewMessage[] = []

    if (typeof record.output_text === 'string' && record.output_text.trim()) {
      output.push({ role: 'assistant', content: record.output_text.trim(), isUser: false })
    }

    const choices = Array.isArray(record.choices) ? record.choices : []
    for (const choice of choices) {
      if (!choice || typeof choice !== 'object') continue
      const message = (choice as Record<string, unknown>).message
      if (message && typeof message === 'object') {
        const msg = message as Record<string, unknown>
        const content = stringifyContent(msg.content)
        if (content.trim()) output.push({ role: String(msg.role || 'assistant'), content: content.trim(), isUser: false })
      }
    }

    const items = Array.isArray(record.output) ? record.output : []
    for (const item of items) {
      if (!item || typeof item !== 'object') continue
      const content = stringifyContent((item as Record<string, unknown>).content ?? item)
      if (content.trim()) output.push({ role: String((item as Record<string, unknown>).role || 'assistant'), content: content.trim(), isUser: false })
    }

    if (output.length) return output
  }

  return raw.trim() ? [{ role: 'response', content: raw.trim(), isUser: false }] : []
}

function parseSseEvents(raw: string): SseEventPreview[] {
  const text = raw.trim()
  if (!text.includes('event:') && !text.includes('data:')) return []
  const blocks = text.split(/\n\s*\n/)
  const events: SseEventPreview[] = []
  for (const block of blocks) {
    const lines = block.split('\n')
    const event = lines.find((line) => line.startsWith('event:'))?.slice(6).trim() || 'message'
    const dataLines = lines.filter((line) => line.startsWith('data:')).map((line) => line.slice(5).trim())
    const payload = dataLines.join('\n')
    if (!payload) continue
    events.push({
      event,
      raw: payload,
      parsed: parseJsonSafe(payload)
    })
  }
  return events
}
</script>
