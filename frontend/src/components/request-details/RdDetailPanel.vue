<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import type { AdminUsageLogDetail } from '@/api/admin/usage'
import RdJsonTree from './RdJsonTree.vue'
import RdMessage from './RdMessage.vue'
import RdMarkdown from './RdMarkdown.vue'
import { extractSessionId } from '@/utils/rdSession'
import { formatDateTime } from '@/utils/format'

interface Props {
  detail: AdminUsageLogDetail | null
  loading?: boolean
  activeSessionId?: string
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  activeSessionId: ''
})

const emit = defineEmits<{
  (e: 'filter-session', sessionId: string): void
  (e: 'close'): void
}>()

type Section = 'input' | 'output' | 'messages' | 'stream' | 'metadata'
const section = ref<Section>('messages')
const sectionView = ref<Record<string, 'parsed' | 'json'>>({})

function parseJsonSafe(raw: unknown): unknown {
  if (raw == null) return null
  if (typeof raw !== 'string') return raw
  try { return JSON.parse(raw) } catch { return raw }
}

const requestBody = computed(() => parseJsonSafe(props.detail?.request_body))
const responseBody = computed(() => parseJsonSafe(props.detail?.response_body))

const messages = computed<any[]>(() => {
  const rb: any = requestBody.value
  return Array.isArray(rb?.messages) ? rb.messages : []
})

const streamChunks = computed<any[]>(() => {
  const raw = String(props.detail?.response_body || '')
  if (!raw.trim()) return []
  // 判断是否 SSE：含有 "data:" 行
  if (!/(^|\n)data:/.test(raw)) return []
  const blocks = raw.split(/\n\s*\n/).filter(Boolean)
  return blocks.map((block, idx) => {
    const lines = block.split('\n')
    const eventLine = lines.find(l => l.startsWith('event:'))
    const dataLines = lines.filter(l => l.startsWith('data:'))
    const raw = dataLines.map(l => l.slice(5).trim()).join('\n')
    let parsed: any
    try { parsed = raw ? JSON.parse(raw) : undefined } catch { parsed = undefined }
    return {
      index: idx + 1,
      event: eventLine ? eventLine.slice(6).trim() : '',
      raw: block,
      parsed
    }
  })
})

const sessionId = computed(() => extractSessionId(requestBody.value as any) || '')

const lastUserMessage = computed(() => {
  const list = messages.value
  for (let i = list.length - 1; i >= 0; i--) {
    const m = list[i]
    if (m?.role !== 'user') continue
    if (typeof m.content === 'string' && m.content.trim()) return m
    if (Array.isArray(m.content) && m.content.some((c: any) => c?.type === 'text' && c.text?.trim())) return m
  }
  for (let i = list.length - 1; i >= 0; i--) {
    if (list[i]?.role === 'user') return list[i]
  }
  return null
})

const finalAssistantMessage = computed(() => {
  const body: any = responseBody.value
  if (!body) return null
  if (body?.choices?.[0]?.message) return body.choices[0].message
  if (Array.isArray(body?.content) && body.role) return body
  return null
})

const tokensSummary = computed(() => {
  const d = props.detail
  if (!d) return null
  const input = Number(d.input_tokens || 0)
  const output = Number(d.output_tokens || 0)
  const cacheRead = Number(d.cache_read_tokens || 0)
  const cacheCreate = Number(d.cache_creation_tokens || 0)
  const total = input + output + cacheRead + cacheCreate
  return { input, output, cacheRead, cacheCreate, total }
})

const costSummary = computed(() => {
  const d = props.detail
  if (!d) return null
  return {
    total: Number(d.total_cost || 0),
    actual: Number(d.actual_cost || 0)
  }
})

const isStream = computed(() => {
  if (props.detail?.stream) return true
  return streamChunks.value.length > 0
})

const pills = computed(() => {
  const d = props.detail
  if (!d) return [] as Array<{ label?: string; value: string; cls?: string }>
  const out: Array<{ label?: string; value: string; cls?: string }> = []
  out.push({ label: 'Time', value: formatDateTime(d.created_at) })
  if (d.duration_ms != null) out.push({ label: 'Latency', value: formatDuration(d.duration_ms), cls: 'mono' })
  if (d.first_token_ms != null) out.push({ label: 'TTFT', value: formatDuration(d.first_token_ms), cls: 'mono' })
  if (d.model) out.push({ value: d.model, cls: 'accent mono' })
  if (tokensSummary.value && tokensSummary.value.total > 0) {
    const t = tokensSummary.value
    out.push({
      value: `${fmtNum(t.input)} → ${fmtNum(t.output)} (Σ ${fmtNum(t.total)})`,
      cls: 'mono'
    })
  }
  if (costSummary.value && costSummary.value.total > 0) {
    out.push({ label: 'Cost', value: `$${costSummary.value.total.toFixed(6)}`, cls: 'mono' })
  }
  if (isStream.value) out.push({ value: 'stream' })
  return out
})

function fmtNum(n: number) { return Number(n || 0).toLocaleString() }

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(2)}s`
  const m = Math.floor(s / 60)
  const r = Math.round(s - m * 60)
  return `${m}m ${r}s`
}

const sections = computed(() => {
  const d = props.detail
  if (!d) return []
  return [
    { id: 'input' as Section, label: 'Input', count: null as number | null, show: true },
    { id: 'output' as Section, label: 'Output', count: null, show: !!responseBody.value || streamChunks.value.length > 0 },
    { id: 'messages' as Section, label: 'Messages', count: messages.value.length, show: messages.value.length > 0 },
    { id: 'stream' as Section, label: 'Stream', count: streamChunks.value.length, show: streamChunks.value.length > 0 },
    { id: 'metadata' as Section, label: 'Metadata', count: null, show: true }
  ].filter(s => s.show)
})

const contentEl = ref<HTMLElement | null>(null)

async function scrollIntoPosition() {
  // 等两帧：第一帧等 Vue 渲染，第二帧等子组件（JSON 树 / Markdown）布局
  await nextTick()
  await new Promise((r) => requestAnimationFrame(() => r(null)))
  const el = contentEl.value
  if (!el) return
  if (section.value === 'messages') {
    el.scrollTop = el.scrollHeight
  } else {
    el.scrollTop = 0
  }
}

watch(() => props.detail?.id, async () => {
  if (!props.detail) return
  if (messages.value.length > 0) section.value = 'messages'
  else if (responseBody.value) section.value = 'output'
  else section.value = 'input'
  scrollIntoPosition()
})

watch(section, () => { scrollIntoPosition() }, { flush: 'post' })

function viewModeOf(key: string, defaultView: 'parsed' | 'json' = 'parsed') {
  return sectionView.value[key] ?? defaultView
}
function setViewMode(key: string, v: 'parsed' | 'json') {
  sectionView.value = { ...sectionView.value, [key]: v }
}

function onCopyId() {
  const id = props.detail?.request_id
  if (!id) return
  navigator.clipboard.writeText(id).catch(() => {})
}

function onFilterSession() {
  if (sessionId.value) emit('filter-session', sessionId.value)
}

function scrollToTop() {
  if (contentEl.value) contentEl.value.scrollTo({ top: 0, behavior: 'smooth' })
}
function scrollToBottom() {
  if (contentEl.value) contentEl.value.scrollTo({ top: contentEl.value.scrollHeight, behavior: 'smooth' })
}

const streamView = ref<'text' | 'json'>('text')
const accumulatedText = computed(() => {
  let t = ''
  streamChunks.value.forEach(c => {
    const p = c.parsed
    if (!p) return
    if (p.choices?.[0]?.delta?.content) t += p.choices[0].delta.content
    else if (p.type === 'content_block_delta' && p.delta?.type === 'text_delta') t += p.delta.text
    else if (typeof p.delta === 'string') t += p.delta
  })
  return t
})
</script>

<template>
  <div class="rd-detail">
    <div v-if="!detail && !loading" class="rd-detail-empty">
      <p>从左侧选择一条请求查看详情</p>
    </div>

    <div v-else-if="loading" class="rd-detail-empty">
      <div class="rd-spinner"></div>
      <span>加载中...</span>
    </div>

    <template v-else-if="detail">
      <div class="rd-detail-header">
        <div class="rd-detail-header-left">
          <span class="rd-detail-label">Trace</span>
          <span class="rd-detail-id">{{ detail.request_id || `#${detail.id}` }}</span>
          <button class="rd-icon-btn" title="Copy ID" @click="onCopyId">
            <svg width="12" height="12" viewBox="0 0 16 16" fill="none">
              <path d="M4 4h8v8H4V4zm1 1v6h6V5H5z" fill="currentColor" />
              <path d="M2 2h8v1H3v7H2V2z" fill="currentColor" />
            </svg>
          </button>
        </div>
        <div class="rd-detail-header-actions">
          <button class="rd-icon-btn" title="Close" @click="emit('close')">
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
              <path d="M12 4L4 12M4 4l8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
          </button>
        </div>
      </div>

      <div class="rd-detail-meta">
        <span v-for="(p, i) in pills" :key="i" class="rd-pill" :class="p.cls?.split(' ').map(c => `rd-pill-${c}`).join(' ')">
          <span v-if="p.label" class="rd-pill-label">{{ p.label }}:</span>
          <span>{{ p.value }}</span>
        </span>
        <span v-if="sessionId" class="rd-pill rd-pill-session">
          <span class="rd-pill-label">Session:</span>
          <span class="rd-pill-mono" :title="sessionId">
            {{ sessionId.length > 10 ? sessionId.slice(0, 8) + '…' : sessionId }}
          </span>
          <button
            class="rd-pill-action"
            :class="{ active: activeSessionId === sessionId }"
            :title="activeSessionId === sessionId ? 'Session filter active' : 'Filter by this session'"
            @click="onFilterSession"
          >
            <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
              <path d="M2 3h12l-5 6v4l-2 1V9L2 3z" fill="currentColor" />
            </svg>
          </button>
        </span>
      </div>

      <div class="rd-detail-body">
        <aside class="rd-nav">
          <button
            v-for="s in sections"
            :key="s.id"
            class="rd-nav-item"
            :class="{ active: section === s.id }"
            @click="section = s.id"
          >
            <span>{{ s.label }}</span>
            <span v-if="s.count != null" class="rd-nav-count">{{ s.count }}</span>
          </button>
        </aside>

        <main ref="contentEl" class="rd-content">
          <!-- INPUT: 最新用户消息 -->
          <template v-if="section === 'input'">
            <div class="rd-section-header">
              <h3 class="rd-section-title">Latest User Input</h3>
              <span v-if="lastUserMessage" class="rd-section-hint">
                turn {{ messages.indexOf(lastUserMessage) + 1 }} / {{ messages.length }}
              </span>
            </div>
            <RdMessage v-if="lastUserMessage" :message="lastUserMessage" :index="0" />
            <div v-else class="rd-card">
              <div class="rd-card-header">Request Body</div>
              <RdJsonTree :data="requestBody" :max-depth="3" />
            </div>
          </template>

          <!-- OUTPUT: assistant 回复 -->
          <template v-else-if="section === 'output'">
            <div class="rd-section-header">
              <h3 class="rd-section-title">Response Output</h3>
              <div class="rd-tabs">
                <button class="rd-tab" :class="{ active: viewModeOf('output') === 'parsed' }" @click="setViewMode('output', 'parsed')">Parsed</button>
                <button class="rd-tab" :class="{ active: viewModeOf('output') === 'json' }" @click="setViewMode('output', 'json')">JSON</button>
              </div>
            </div>
            <template v-if="viewModeOf('output') === 'parsed'">
              <RdMessage v-if="finalAssistantMessage" :message="finalAssistantMessage" :index="0" />
              <div v-else-if="accumulatedText" class="rd-card">
                <div class="rd-card-header">Accumulated from stream</div>
                <RdMarkdown :text="accumulatedText" />
              </div>
              <div v-else-if="responseBody" class="rd-card">
                <div class="rd-card-header">Response Body</div>
                <RdJsonTree :data="responseBody" :max-depth="3" :show-copy="false" />
              </div>
              <div v-else class="rd-empty"><span>(no response)</span></div>
            </template>
            <template v-else>
              <RdJsonTree :data="responseBody ?? detail.response_body" :max-depth="3" />
            </template>
          </template>

          <!-- MESSAGES: 完整对话 -->
          <template v-else-if="section === 'messages'">
            <div class="rd-section-header">
              <h3 class="rd-section-title">
                Messages <span class="rd-section-badge">{{ messages.length }}</span>
              </h3>
              <div class="rd-scroll-controls">
                <button class="rd-scroll-btn" title="滚动到顶部" @click="scrollToTop">
                  <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                    <path d="M8 3L3 8h3v5h4V8h3L8 3z" fill="currentColor" />
                  </svg>
                </button>
                <button class="rd-scroll-btn" title="滚动到底部" @click="scrollToBottom">
                  <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                    <path d="M8 13l5-5h-3V3H6v5H3l5 5z" fill="currentColor" />
                  </svg>
                </button>
              </div>
            </div>
            <div class="rd-messages">
              <RdMessage v-for="(m, i) in messages" :key="i" :message="m" :index="i" />
            </div>
          </template>

          <!-- STREAM -->
          <template v-else-if="section === 'stream'">
            <div class="rd-section-header">
              <h3 class="rd-section-title">
                Stream <span class="rd-section-badge">{{ streamChunks.length }}</span>
              </h3>
              <div class="rd-tabs">
                <button class="rd-tab" :class="{ active: streamView === 'text' }" @click="streamView = 'text'">Accumulated</button>
                <button class="rd-tab" :class="{ active: streamView === 'json' }" @click="streamView = 'json'">Chunks</button>
              </div>
            </div>
            <div v-if="streamView === 'text'" class="rd-card">
              <div class="rd-card-header">Accumulated Text</div>
              <div class="rd-stream-text">{{ accumulatedText || '(no text content)' }}</div>
            </div>
            <div v-else class="rd-stream-chunks">
              <div v-for="chunk in streamChunks" :key="chunk.index" class="rd-stream-chunk">
                <span class="rd-stream-chunk-idx">#{{ chunk.index }}</span>
                <div class="rd-stream-chunk-body">
                  <div v-if="chunk.event" class="rd-stream-event">{{ chunk.event }}</div>
                  <RdJsonTree v-if="chunk.parsed" :data="chunk.parsed" :max-depth="2" :show-copy="false" />
                  <pre v-else class="rd-stream-raw">{{ chunk.raw }}</pre>
                </div>
              </div>
            </div>
          </template>

          <!-- METADATA -->
          <template v-else-if="section === 'metadata'">
            <div class="rd-section-header">
              <h3 class="rd-section-title">Metadata</h3>
            </div>
            <div class="rd-card">
              <table class="rd-kv">
                <tbody>
                  <tr><th>Request ID</th><td class="mono">{{ detail.request_id || `#${detail.id}` }}</td></tr>
                  <tr v-if="sessionId">
                    <th>Session ID</th>
                    <td>
                      <span class="rd-pill-mono" :title="sessionId">{{ sessionId }}</span>
                      <button
                        class="rd-btn rd-btn-sm"
                        :class="{ 'rd-btn-accent': activeSessionId === sessionId }"
                        style="margin-left: 0.5rem;"
                        @click="onFilterSession"
                      >
                        <svg width="11" height="11" viewBox="0 0 16 16" fill="none">
                          <path d="M2 3h12l-5 6v4l-2 1V9L2 3z" fill="currentColor" />
                        </svg>
                        {{ activeSessionId === sessionId ? 'Filtered' : 'Filter this session' }}
                      </button>
                    </td>
                  </tr>
                  <tr><th>Created</th><td>{{ formatDateTime(detail.created_at) }}</td></tr>
                  <tr v-if="detail.user?.email"><th>User</th><td>{{ detail.user.email }}</td></tr>
                  <tr v-if="detail.api_key?.name"><th>API Key</th><td class="mono">{{ detail.api_key.name }}</td></tr>
                  <tr v-if="detail.account?.name"><th>Account</th><td class="mono">{{ detail.account.name }}</td></tr>
                  <tr v-if="detail.model"><th>Model</th><td class="mono">{{ detail.model }}</td></tr>
                  <tr v-if="detail.upstream_model"><th>Upstream Model</th><td class="mono">{{ detail.upstream_model }}</td></tr>
                  <tr v-if="detail.upstream_endpoint"><th>Upstream</th><td class="mono">{{ detail.upstream_endpoint }}</td></tr>
                  <tr v-if="detail.inbound_endpoint"><th>Inbound</th><td class="mono">{{ detail.inbound_endpoint }}</td></tr>
                  <tr><th>Stream</th><td>{{ detail.stream ? 'yes' : 'no' }}</td></tr>
                  <tr v-if="detail.duration_ms != null"><th>Duration</th><td class="mono">{{ detail.duration_ms }} ms</td></tr>
                  <tr v-if="detail.first_token_ms != null"><th>TTFT</th><td class="mono">{{ detail.first_token_ms }} ms</td></tr>
                  <tr v-if="tokensSummary && tokensSummary.input"><th>Input Tokens</th><td class="mono">{{ fmtNum(tokensSummary.input) }}</td></tr>
                  <tr v-if="tokensSummary && tokensSummary.output"><th>Output Tokens</th><td class="mono">{{ fmtNum(tokensSummary.output) }}</td></tr>
                  <tr v-if="tokensSummary && tokensSummary.cacheRead"><th>Cache Read</th><td class="mono">{{ fmtNum(tokensSummary.cacheRead) }}</td></tr>
                  <tr v-if="tokensSummary && tokensSummary.cacheCreate"><th>Cache Creation</th><td class="mono">{{ fmtNum(tokensSummary.cacheCreate) }}</td></tr>
                  <tr v-if="costSummary && costSummary.total"><th>Total Cost</th><td class="mono">${{ costSummary.total.toFixed(6) }}</td></tr>
                  <tr v-if="costSummary && costSummary.actual"><th>Actual Cost</th><td class="mono">${{ costSummary.actual.toFixed(6) }}</td></tr>
                  <tr v-if="detail.ip_address"><th>Client IP</th><td class="mono">{{ detail.ip_address }}</td></tr>
                  <tr v-if="detail.user_agent"><th>User Agent</th><td class="mono">{{ detail.user_agent }}</td></tr>
                </tbody>
              </table>
            </div>
          </template>
        </main>
      </div>
    </template>
  </div>
</template>

<style scoped>
.rd-detail {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 8px;
  overflow: hidden;
}

:global(.dark) .rd-detail {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-detail-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  gap: 0.75rem;
  color: rgb(100 116 139);
  font-size: 0.875rem;
}

.rd-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid rgb(226 232 240);
  border-top-color: rgb(20 184 166);
  border-radius: 50%;
  animation: rd-spin 0.8s linear infinite;
}

@keyframes rd-spin { to { transform: rotate(360deg); } }

.rd-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.625rem 1rem;
  border-bottom: 1px solid rgb(226 232 240);
  min-height: 44px;
  flex-shrink: 0;
}

:global(.dark) .rd-detail-header {
  border-bottom-color: rgb(51 65 85);
}

.rd-detail-header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.rd-detail-label {
  font-size: 0.6875rem;
  font-weight: 600;
  color: rgb(100 116 139);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 0.1875rem 0.5rem;
  background: rgb(248 250 252);
  border-radius: 4px;
}

:global(.dark) .rd-detail-label {
  background: rgb(15 23 42);
}

.rd-detail-id {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.8125rem;
  font-weight: 500;
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rd-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  padding: 0;
  background: transparent;
  border: none;
  color: rgb(148 163 184);
  cursor: pointer;
  border-radius: 4px;
}

.rd-icon-btn:hover {
  background: rgb(241 245 249);
  color: rgb(15 23 42);
}

:global(.dark) .rd-icon-btn:hover {
  background: rgb(51 65 85);
  color: rgb(241 245 249);
}

.rd-detail-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.375rem;
  padding: 0.625rem 1rem;
  border-bottom: 1px solid rgb(226 232 240);
  flex-shrink: 0;
}

:global(.dark) .rd-detail-meta {
  border-bottom-color: rgb(51 65 85);
}

.rd-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.3125rem;
  padding: 0.1875rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: rgb(15 23 42);
  background: rgb(248 250 252);
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
  line-height: 1.3;
  white-space: nowrap;
}

:global(.dark) .rd-pill {
  background: rgb(15 23 42);
  border-color: rgb(51 65 85);
  color: rgb(241 245 249);
}

.rd-pill-label {
  color: rgb(100 116 139);
  font-weight: 400;
}

.rd-pill-mono {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.71875rem;
}

.rd-pill-accent {
  color: rgb(13 148 136);
  border-color: rgb(94 234 212);
  background: rgb(240 253 250);
}

:global(.dark) .rd-pill-accent {
  color: rgb(94 234 212);
  border-color: rgb(15 118 110);
  background: rgba(20, 184, 166, 0.1);
}

.rd-pill-session { gap: 0.25rem; }

.rd-pill-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  padding: 0;
  background: transparent;
  border: none;
  color: rgb(148 163 184);
  cursor: pointer;
  border-radius: 3px;
}

.rd-pill-action:hover,
.rd-pill-action.active {
  background: rgb(240 253 250);
  color: rgb(13 148 136);
}

:global(.dark) .rd-pill-action:hover,
:global(.dark) .rd-pill-action.active {
  background: rgba(20, 184, 166, 0.15);
  color: rgb(94 234 212);
}

.rd-detail-body {
  display: flex;
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.rd-nav {
  width: 200px;
  flex-shrink: 0;
  padding: 0.5rem;
  border-right: 1px solid rgb(226 232 240);
  overflow-y: auto;
}

:global(.dark) .rd-nav {
  border-right-color: rgb(51 65 85);
}

.rd-nav-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.4375rem 0.625rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgb(100 116 139);
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-bottom: 0.125rem;
  text-align: left;
}

.rd-nav-item:hover {
  background: rgb(248 250 252);
  color: rgb(15 23 42);
}

:global(.dark) .rd-nav-item:hover {
  background: rgb(15 23 42);
  color: rgb(241 245 249);
}

.rd-nav-item.active {
  background: rgb(240 253 250);
  color: rgb(13 148 136);
}

:global(.dark) .rd-nav-item.active {
  background: rgba(20, 184, 166, 0.1);
  color: rgb(94 234 212);
}

.rd-nav-count {
  font-size: 0.6875rem;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  color: rgb(148 163 184);
  padding: 0.0625rem 0.3125rem;
  background: rgb(241 245 249);
  border-radius: 3px;
}

:global(.dark) .rd-nav-count {
  background: rgb(51 65 85);
}

.rd-nav-item.active .rd-nav-count {
  background: rgba(20, 184, 166, 0.15);
  color: rgb(13 148 136);
}

:global(.dark) .rd-nav-item.active .rd-nav-count {
  color: rgb(94 234 212);
}

.rd-content {
  flex: 1;
  overflow-y: auto;
  padding: 1rem 1.25rem;
}

.rd-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.875rem;
  gap: 0.75rem;
}

.rd-section-title {
  font-size: 0.9375rem;
  font-weight: 600;
  margin: 0;
  color: rgb(15 23 42);
}

:global(.dark) .rd-section-title { color: rgb(241 245 249); }

.rd-section-hint {
  font-size: 0.75rem;
  color: rgb(148 163 184);
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
}

.rd-scroll-controls {
  display: inline-flex;
  gap: 0.25rem;
}

.rd-scroll-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
  color: rgb(100 116 139);
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.rd-scroll-btn:hover {
  background: rgb(240 253 250);
  color: rgb(13 148 136);
  border-color: rgb(94 234 212);
}

:global(.dark) .rd-scroll-btn {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
  color: rgb(148 163 184);
}

:global(.dark) .rd-scroll-btn:hover {
  background: rgba(20, 184, 166, 0.08);
  color: rgb(94 234 212);
  border-color: rgb(15 118 110);
}

.rd-section-badge {
  font-size: 0.6875rem;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  color: rgb(100 116 139);
  padding: 0.0625rem 0.375rem;
  background: rgb(248 250 252);
  border-radius: 3px;
  margin-left: 0.375rem;
}

:global(.dark) .rd-section-badge {
  background: rgb(15 23 42);
  color: rgb(148 163 184);
}

.rd-tabs {
  display: inline-flex;
  padding: 2px;
  background: rgb(248 250 252);
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
}

:global(.dark) .rd-tabs {
  background: rgb(15 23 42);
  border-color: rgb(51 65 85);
}

.rd-tab {
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: rgb(100 116 139);
  background: transparent;
  border: none;
  border-radius: 3px;
  cursor: pointer;
}

.rd-tab:hover {
  color: rgb(15 23 42);
}

:global(.dark) .rd-tab:hover {
  color: rgb(241 245 249);
}

.rd-tab.active {
  background: white;
  color: rgb(15 23 42);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

:global(.dark) .rd-tab.active {
  background: rgb(51 65 85);
  color: rgb(241 245 249);
}

.rd-card {
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 6px;
  overflow: hidden;
  margin-bottom: 0.75rem;
}

:global(.dark) .rd-card {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-card-header {
  padding: 0.5rem 0.75rem;
  background: rgb(248 250 252);
  border-bottom: 1px solid rgb(226 232 240);
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(100 116 139);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

:global(.dark) .rd-card-header {
  background: rgb(15 23 42);
  border-bottom-color: rgb(51 65 85);
}

.rd-card :deep(.rd-json-viewer) {
  border: none;
  border-radius: 0;
}

.rd-card :deep(.rd-md) {
  border-radius: 0;
}

.rd-empty {
  padding: 2rem;
  text-align: center;
  color: rgb(148 163 184);
  font-size: 0.875rem;
}

.rd-messages {
  display: flex;
  flex-direction: column;
}

.rd-stream-text {
  padding: 0.875rem 1rem;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.875rem;
  line-height: 1.65;
}

.rd-stream-chunks {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.rd-stream-chunk {
  display: flex;
  gap: 0.625rem;
  padding: 0.5rem 0.625rem;
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
  font-size: 0.75rem;
}

:global(.dark) .rd-stream-chunk {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-stream-chunk-idx {
  flex-shrink: 0;
  color: rgb(148 163 184);
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-weight: 600;
  min-width: 2.5rem;
}

.rd-stream-chunk-body {
  flex: 1;
  min-width: 0;
}

.rd-stream-event {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.6875rem;
  font-weight: 600;
  color: rgb(245 158 11);
  margin-bottom: 0.1875rem;
  text-transform: uppercase;
}

.rd-stream-raw {
  margin: 0;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.75rem;
  color: rgb(100 116 139);
}

.rd-kv {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.rd-kv th,
.rd-kv td {
  padding: 0.4375rem 0.75rem;
  text-align: left;
  border-bottom: 1px solid rgb(226 232 240);
  vertical-align: top;
}

:global(.dark) .rd-kv th,
:global(.dark) .rd-kv td {
  border-bottom-color: rgb(51 65 85);
}

.rd-kv tr:last-child th,
.rd-kv tr:last-child td {
  border-bottom: none;
}

.rd-kv th {
  width: 160px;
  font-weight: 500;
  color: rgb(100 116 139);
  font-size: 0.75rem;
}

.rd-kv td.mono {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  word-break: break-word;
}

.rd-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  line-height: 1;
  color: rgb(15 23 42);
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
  cursor: pointer;
}

.rd-btn:hover { background: rgb(248 250 252); }

:global(.dark) .rd-btn {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
  color: rgb(241 245 249);
}

:global(.dark) .rd-btn:hover { background: rgb(15 23 42); }

.rd-btn-sm {
  padding: 0.25rem 0.5rem;
  font-size: 0.6875rem;
}

.rd-btn-accent {
  color: rgb(13 148 136);
  border-color: rgb(94 234 212);
  background: rgb(240 253 250);
}

:global(.dark) .rd-btn-accent {
  color: rgb(94 234 212);
  border-color: rgb(15 118 110);
  background: rgba(20, 184, 166, 0.1);
}
</style>
