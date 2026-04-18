<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import RdDetailPanel from '@/components/request-details/RdDetailPanel.vue'
import { adminUsageAPI, type AdminUsageLogDetail, type AdminUsageStatsResponse } from '@/api/admin/usage'
import type { AdminUsageLog, UsageRequestType } from '@/types'
import { useAppStore } from '@/stores'
import { useI18n } from 'vue-i18n'

const appStore = useAppStore()
const router = useRouter()
const route = useRoute()
const { t } = useI18n()

function formatLocalDate(d: Date) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function getDefaultRange() {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return { start: formatLocalDate(start), end: formatLocalDate(end) }
}

const defaultRange = getDefaultRange()

// ───── URL 同步的筛选状态 ─────
const startDate = ref((route.query.start_date as string) || defaultRange.start)
const endDate = ref((route.query.end_date as string) || defaultRange.end)
const requestFilter = ref<'all' | UsageRequestType>(((route.query.type as string) || 'all') as any)
const modelFilter = ref((route.query.model as string) || '')
const sessionFilter = ref((route.query.session_id as string) || '')
const searchQuery = ref((route.query.q as string) || '')

const selectedRequestId = ref<number | null>(
  route.query.request_id ? Number(route.query.request_id) : null
)

// ───── 数据 ─────
const requests = ref<AdminUsageLog[]>([])
const selectedDetail = ref<AdminUsageLogDetail | null>(null)
const stats = ref<AdminUsageStatsResponse | null>(null)
const loading = ref(false)
const detailLoading = ref(false)

let listAbort: AbortController | null = null

const modalOpen = computed(() => selectedRequestId.value != null)

const visibleRequests = computed(() => {
  let list = requests.value
  if (sessionFilter.value) {
    list = list.filter(r => (r as any).session_id === sessionFilter.value)
  }
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(r =>
      (r.request_id || '').toLowerCase().includes(q) ||
      (r.model || '').toLowerCase().includes(q) ||
      (r.upstream_endpoint || '').toLowerCase().includes(q) ||
      (r.user?.email || '').toLowerCase().includes(q)
    )
  }
  return list
})

// ───── URL 同步 ─────
function syncUrl() {
  const query: Record<string, string> = {}
  if (startDate.value !== defaultRange.start) query.start_date = startDate.value
  if (endDate.value !== defaultRange.end) query.end_date = endDate.value
  if (requestFilter.value !== 'all') query.type = requestFilter.value
  if (modelFilter.value) query.model = modelFilter.value
  if (sessionFilter.value) query.session_id = sessionFilter.value
  if (searchQuery.value) query.q = searchQuery.value
  if (selectedRequestId.value) query.request_id = String(selectedRequestId.value)
  router.replace({ query })
}

// ───── 加载 ─────
async function loadRequests() {
  listAbort?.abort()
  listAbort = new AbortController()
  try {
    const response = await adminUsageAPI.list(
      {
        page: 1,
        page_size: 100,
        start_date: startDate.value,
        end_date: endDate.value,
        request_type: requestFilter.value === 'all' ? undefined : requestFilter.value,
        model: modelFilter.value || undefined,
        exact_total: false
      },
      { signal: listAbort.signal }
    )
    requests.value = response.items || []
    if (selectedRequestId.value) loadDetail()
  } catch (error: any) {
    if (error?.name === 'AbortError' || error?.code === 'ERR_CANCELED') return
    appStore.showError(error?.message || t('admin.requestDetails.failedToLoad'))
  }
}

async function loadDetail() {
  if (!selectedRequestId.value) {
    selectedDetail.value = null
    return
  }
  detailLoading.value = true
  try {
    selectedDetail.value = await adminUsageAPI.getById(selectedRequestId.value)
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.requestDetails.failedToLoad'))
  } finally {
    detailLoading.value = false
  }
}

async function loadStats() {
  try {
    stats.value = await adminUsageAPI.getStats({
      start_date: startDate.value,
      end_date: endDate.value,
      request_type: requestFilter.value === 'all' ? undefined : requestFilter.value
    })
  } catch { /* ignore */ }
}

async function refreshAll() {
  loading.value = true
  try {
    await Promise.all([loadRequests(), loadStats()])
  } finally {
    loading.value = false
  }
}

// ───── 事件 ─────
function onDateRangeChange(range: { startDate: string; endDate: string }) {
  startDate.value = range.startDate
  endDate.value = range.endDate
  syncUrl()
  refreshAll()
}

function onFilterChange() {
  syncUrl()
  refreshAll()
}

function onSelectRequest(id: number) {
  selectedRequestId.value = id
  syncUrl()
  loadDetail()
}

function onFilterSession(sid: string) {
  sessionFilter.value = sid
  syncUrl()
  closeModal()
}

function clearSessionFilter() {
  sessionFilter.value = ''
  syncUrl()
}

function closeModal() {
  selectedRequestId.value = null
  selectedDetail.value = null
  syncUrl()
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && modalOpen.value) closeModal()
}

function onPopState() {
  startDate.value = (route.query.start_date as string) || defaultRange.start
  endDate.value = (route.query.end_date as string) || defaultRange.end
  requestFilter.value = ((route.query.type as string) || 'all') as any
  modelFilter.value = (route.query.model as string) || ''
  sessionFilter.value = (route.query.session_id as string) || ''
  searchQuery.value = (route.query.q as string) || ''
  const qid = route.query.request_id ? Number(route.query.request_id) : null
  if (qid !== selectedRequestId.value) {
    selectedRequestId.value = qid
    loadDetail()
  }
}

watch(() => route.query, onPopState)

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
  refreshAll()
})

onUnmounted(() => {
  listAbort?.abort()
  document.removeEventListener('keydown', onKeydown)
})

// ───── 派生/展示 ─────
const statCards = computed(() => {
  if (!stats.value) return []
  const totalTokens =
    Number(stats.value.total_input_tokens || 0) +
    Number(stats.value.total_output_tokens || 0) +
    Number(stats.value.total_cache_tokens || 0)
  return [
    { label: t('admin.requestDetails.stats.requests'), value: fmtNum(stats.value.total_requests), hint: t('admin.requestDetails.stats.requestsHint') },
    { label: t('admin.requestDetails.stats.tokens'), value: fmtNum(totalTokens), hint: t('admin.requestDetails.stats.tokensHint') },
    { label: t('admin.requestDetails.stats.cost'), value: `$${Number(stats.value.total_cost || 0).toFixed(4)}`, hint: t('admin.requestDetails.stats.costHint') },
    { label: t('admin.requestDetails.stats.duration'), value: `${Math.round(stats.value.average_duration_ms || 0)}ms`, hint: t('admin.requestDetails.stats.durationHint') }
  ]
})

function fmtNum(n: number) { return Number(n || 0).toLocaleString() }

function formatTimeCell(iso: string | null | undefined): string {
  if (!iso) return '-'
  const d = new Date(iso)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60_000) return 'just now'
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`
  return d.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function formatDuration(ms: number | null | undefined): string {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(2)}s`
  const m = Math.floor(s / 60)
  return `${m}m ${Math.round(s - m * 60)}s`
}

function tokensFor(r: AdminUsageLog): string {
  const input = Number(r.input_tokens || 0)
  const output = Number(r.output_tokens || 0)
  if (!input && !output) return '—'
  return `${fmtNum(input)} → ${fmtNum(output)}`
}

function rowStatus(r: AdminUsageLog): 'success' | 'error' | 'processing' {
  if ((r.output_tokens || 0) > 0) return 'success'
  if (r.duration_ms == null) return 'processing'
  return 'success'
}
</script>

<template>
  <AppLayout>
    <div class="rd-page">
      <!-- Header -->
      <header class="rd-page-header">
        <div>
          <h2 class="rd-page-title">{{ t('admin.requestDetails.title') }}</h2>
          <p class="rd-page-subtitle">{{ t('admin.requestDetails.description') }}</p>
        </div>
        <div class="rd-page-actions">
          <button class="rd-action-btn" @click="refreshAll" :disabled="loading">
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
              <path d="M13.65 2.35A7.95 7.95 0 008 0C3.58 0 .01 3.58.01 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08c-.82 2.33-3.04 4-5.65 4a6 6 0 01-6-6c0-3.31 2.69-6 6-6 1.66 0 3.14.69 4.22 1.78L9 7h7V0l-2.35 2.35z" fill="currentColor" />
            </svg>
            刷新
          </button>
        </div>
      </header>

      <!-- Stats -->
      <div v-if="stats" class="rd-stats">
        <div v-for="card in statCards" :key="card.label" class="rd-stat-card">
          <div class="rd-stat-label" :title="card.hint">{{ card.label }}</div>
          <div class="rd-stat-value">{{ card.value }}</div>
        </div>
      </div>

      <!-- Toolbar -->
      <div class="rd-toolbar">
        <DateRangePicker
          v-model:start-date="startDate"
          v-model:end-date="endDate"
          @change="onDateRangeChange"
        />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索 request_id / model / upstream / 用户..."
          class="rd-input"
          @input="syncUrl"
        >
        <select v-model="requestFilter" class="rd-select" @change="onFilterChange">
          <option value="all">{{ t('admin.requestDetails.allTypes') }}</option>
          <option value="sync">Sync</option>
          <option value="stream">Stream</option>
          <option value="ws_v2">WS</option>
        </select>
        <input
          v-model="modelFilter"
          type="text"
          :placeholder="t('admin.requestDetails.modelPlaceholder')"
          class="rd-input rd-input-sm"
          @change="onFilterChange"
        >
        <span class="rd-trace-count">{{ visibleRequests.length }} {{ t('admin.requestDetails.requestCountSuffix') }}</span>
      </div>

      <!-- Active filter chip -->
      <div v-if="sessionFilter" class="rd-active-filters">
        <span class="rd-chip">
          <span class="rd-chip-label">Session</span>
          <span class="rd-chip-value" :title="sessionFilter">{{ sessionFilter }}</span>
          <button class="rd-chip-close" @click="clearSessionFilter">
            <svg width="10" height="10" viewBox="0 0 16 16" fill="none">
              <path d="M12 4L4 12M4 4l8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
          </button>
        </span>
      </div>

      <!-- Table list -->
      <div class="rd-table-wrapper">
        <div v-if="loading" class="rd-empty">
          <div class="rd-spinner"></div>
          <span>加载中...</span>
        </div>
        <div v-else-if="visibleRequests.length === 0" class="rd-empty">
          <span>{{ t('admin.requestDetails.empty') }}</span>
        </div>
        <table v-else class="rd-table">
          <thead>
            <tr>
              <th style="width: 130px;">Time</th>
              <th>Request ID</th>
              <th>Model</th>
              <th>Upstream</th>
              <th style="width: 96px;">Latency</th>
              <th style="width: 140px;">Tokens</th>
              <th>User</th>
              <th style="width: 96px;">Status</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="req in visibleRequests"
              :key="req.id"
              :class="{ active: selectedRequestId === req.id }"
              @click="onSelectRequest(req.id)"
            >
              <td class="rd-cell-time">{{ formatTimeCell(req.created_at) }}</td>
              <td class="rd-cell-id" :title="req.request_id">
                {{ req.request_id || `#${req.id}` }}
                <span v-if="req.stream" class="rd-tag">stream</span>
              </td>
              <td class="rd-cell-mono">{{ req.model || '—' }}</td>
              <td class="rd-cell-mono rd-cell-muted">{{ req.upstream_endpoint || '—' }}</td>
              <td class="rd-cell-num">{{ formatDuration(req.duration_ms) }}</td>
              <td class="rd-cell-num">{{ tokensFor(req) }}</td>
              <td class="rd-cell-muted">{{ req.user?.email || '—' }}</td>
              <td>
                <span class="rd-status" :class="`rd-status-${rowStatus(req)}`">
                  {{ rowStatus(req) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Modal -->
      <Teleport to="body">
        <div v-if="modalOpen" class="rd-modal">
          <div class="rd-modal-overlay" @click="closeModal"></div>
          <div class="rd-modal-container">
            <RdDetailPanel
              :detail="selectedDetail"
              :loading="detailLoading"
              :active-session-id="sessionFilter"
              @filter-session="onFilterSession"
              @close="closeModal"
            />
          </div>
        </div>
      </Teleport>
    </div>
  </AppLayout>
</template>

<style scoped>
.rd-page {
  min-height: 100vh;
  padding: 1rem 1.25rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  background: rgb(248 250 252);
}

:global(.dark) .rd-page {
  background: rgb(15 23 42);
}

.rd-page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgb(226 232 240);
}

:global(.dark) .rd-page-header { border-bottom-color: rgb(51 65 85); }

.rd-page-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0 0 0.25rem;
  letter-spacing: -0.01em;
  color: rgb(15 23 42);
}

:global(.dark) .rd-page-title { color: rgb(241 245 249); }

.rd-page-subtitle {
  margin: 0;
  font-size: 0.8125rem;
  color: rgb(100 116 139);
}

.rd-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.4375rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgb(15 23 42);
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 6px;
  cursor: pointer;
}

.rd-action-btn:hover:not(:disabled) { background: rgb(248 250 252); border-color: rgb(203 213 225); }
.rd-action-btn:disabled { opacity: 0.5; cursor: not-allowed; }

:global(.dark) .rd-action-btn {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
  color: rgb(241 245 249);
}

:global(.dark) .rd-action-btn:hover:not(:disabled) {
  background: rgb(15 23 42);
  border-color: rgb(71 85 105);
}

.rd-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 0.75rem;
}

.rd-stat-card {
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 6px;
  padding: 0.75rem 0.875rem;
}

:global(.dark) .rd-stat-card {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-stat-label {
  font-size: 0.6875rem;
  font-weight: 500;
  color: rgb(100 116 139);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 0.25rem;
}

.rd-stat-value {
  font-size: 1.25rem;
  font-weight: 600;
  color: rgb(15 23 42);
  font-variant-numeric: tabular-nums;
}

:global(.dark) .rd-stat-value { color: rgb(241 245 249); }

.rd-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}

.rd-input, .rd-select {
  height: 32px;
  padding: 0 0.625rem;
  font-size: 0.8125rem;
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
  color: rgb(15 23 42);
  outline: none;
}

.rd-input:focus, .rd-select:focus { border-color: rgb(20 184 166); }

.rd-input { min-width: 280px; flex: 1; max-width: 420px; }
.rd-input-sm { min-width: 160px; max-width: 200px; flex: 0 0 auto; }

:global(.dark) .rd-input, :global(.dark) .rd-select {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
  color: rgb(241 245 249);
}

.rd-trace-count {
  margin-left: auto;
  font-size: 0.75rem;
  color: rgb(100 116 139);
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
}

.rd-active-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  padding: 0.5rem 0.625rem;
  background: rgb(240 253 250);
  border: 1px solid rgb(94 234 212);
  border-radius: 6px;
}

:global(.dark) .rd-active-filters {
  background: rgba(20, 184, 166, 0.08);
  border-color: rgb(15 118 110);
}

.rd-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.1875rem 0.375rem 0.1875rem 0.5rem;
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
  font-size: 0.75rem;
}

:global(.dark) .rd-chip {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-chip-label { color: rgb(100 116 139); font-weight: 500; }

.rd-chip-value {
  color: rgb(15 23 42);
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
}

:global(.dark) .rd-chip-value { color: rgb(241 245 249); }

.rd-chip-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  padding: 0;
  background: transparent;
  border: none;
  color: rgb(148 163 184);
  cursor: pointer;
  border-radius: 3px;
}

.rd-chip-close:hover { background: rgb(248 250 252); color: rgb(239 68 68); }

:global(.dark) .rd-chip-close:hover { background: rgb(51 65 85); }

/* Table */
.rd-table-wrapper {
  border: 1px solid rgb(226 232 240);
  border-radius: 6px;
  overflow: hidden;
  background: white;
  min-height: 320px;
}

:global(.dark) .rd-table-wrapper {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.rd-table thead th {
  text-align: left;
  padding: 0.625rem 0.875rem;
  font-size: 0.6875rem;
  font-weight: 600;
  color: rgb(100 116 139);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  background: rgb(248 250 252);
  border-bottom: 1px solid rgb(226 232 240);
  white-space: nowrap;
}

:global(.dark) .rd-table thead th {
  background: rgb(15 23 42);
  border-bottom-color: rgb(51 65 85);
}

.rd-table tbody tr {
  border-bottom: 1px solid rgb(226 232 240);
  cursor: pointer;
  transition: background 0.1s;
}

.rd-table tbody tr:last-child { border-bottom: none; }

.rd-table tbody tr:hover {
  background: rgb(248 250 252);
}

:global(.dark) .rd-table tbody tr { border-bottom-color: rgb(51 65 85); }
:global(.dark) .rd-table tbody tr:hover { background: rgb(15 23 42); }

.rd-table tbody tr.active {
  background: rgb(240 253 250);
}

:global(.dark) .rd-table tbody tr.active {
  background: rgba(20, 184, 166, 0.08);
}

.rd-table td {
  padding: 0.625rem 0.875rem;
  vertical-align: middle;
  color: rgb(15 23 42);
}

:global(.dark) .rd-table td { color: rgb(241 245 249); }

.rd-cell-time {
  color: rgb(100 116 139);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.rd-cell-id {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rd-cell-mono {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
}

.rd-cell-muted { color: rgb(100 116 139); }

.rd-cell-num {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.rd-tag {
  display: inline-flex;
  align-items: center;
  padding: 0.0625rem 0.3125rem;
  font-size: 0.625rem;
  font-weight: 500;
  color: rgb(13 148 136);
  background: rgb(240 253 250);
  border: 1px solid rgb(94 234 212);
  border-radius: 3px;
  margin-left: 0.375rem;
}

:global(.dark) .rd-tag {
  background: rgba(20, 184, 166, 0.1);
  border-color: rgb(15 118 110);
  color: rgb(94 234 212);
}

.rd-status {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.125rem 0.4375rem;
  font-size: 0.6875rem;
  font-weight: 500;
  border-radius: 4px;
  text-transform: capitalize;
  border: 1px solid transparent;
}

.rd-status::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.rd-status-success {
  color: rgb(16 185 129);
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.25);
}

.rd-status-error {
  color: rgb(239 68 68);
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.25);
}

.rd-status-processing {
  color: rgb(245 158 11);
  background: rgba(245, 158, 11, 0.08);
  border-color: rgba(245, 158, 11, 0.25);
}

.rd-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  gap: 0.75rem;
  color: rgb(100 116 139);
  font-size: 0.8125rem;
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

/* Modal */
.rd-modal {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.rd-modal-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(2px);
}

:global(.dark) .rd-modal-overlay {
  background: rgba(0, 0, 0, 0.6);
}

.rd-modal-container {
  position: relative;
  width: min(1440px, 96vw);
  height: min(92vh, 100%);
  display: flex;
  flex-direction: column;
}

@media (max-width: 1024px) {
  .rd-modal-container {
    width: 98vw;
    height: 96vh;
  }
}
</style>
