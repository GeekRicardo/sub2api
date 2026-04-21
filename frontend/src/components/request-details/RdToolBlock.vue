<script setup lang="ts">
import { computed, ref } from 'vue'
import RdJsonTree from './RdJsonTree.vue'
import RdMarkdown from './RdMarkdown.vue'

interface Props {
  kind: 'call' | 'result'
  name?: string
  id?: string | null
  /** 解析后的数据（object 用 JSON 树，string 用行号纯文本） */
  data: unknown
  /** 原始项（用于 JSON 视图） */
  raw?: unknown
}

const props = withDefaults(defineProps<Props>(), {
  name: '',
  id: null,
  raw: undefined
})

const expanded = ref(true)
type ToolView = 'markdown' | 'text' | 'json'
// tool_call 的 data 通常是 object 参数，用 Text 的 KV 表格更直观；
// tool_result 多为自然语言输出，默认 Markdown 渲染更好。
const view = ref<ToolView>(props.kind === 'call' ? 'text' : 'markdown')

// tool_result 常见形态是 [{ type: 'text', text: '...' }, ...] 的纯文本数组。
// Parsed 视图下应直接把 text 拼接后当纯文本展示，而不是展示 JSON 树。
const textArrayContent = computed<string | null>(() => {
  const arr = props.data
  if (!Array.isArray(arr) || arr.length === 0) return null
  const texts: string[] = []
  for (const item of arr) {
    if (!item || typeof item !== 'object') return null
    if ((item as any).type !== 'text') return null
    const t = (item as any).text
    if (typeof t !== 'string') return null
    texts.push(t)
  }
  return texts.join('\n')
})

// tool_call 的 input / 任意对象参数：渲染为两列 key-value 表格。
// 只有 data 是 plain object（非数组）时启用。
const kvEntries = computed<Array<[string, unknown]> | null>(() => {
  const d = props.data
  if (d == null || typeof d !== 'object' || Array.isArray(d)) return null
  const entries = Object.entries(d as Record<string, unknown>)
  return entries.length > 0 ? entries : null
})

function valueKind(v: unknown): 'string' | 'scalar' | 'null' | 'object' {
  if (v === null || v === undefined) return 'null'
  if (typeof v === 'string') return 'string'
  if (typeof v === 'number' || typeof v === 'boolean') return 'scalar'
  return 'object'
}

// Markdown 视图的文本来源：
//   - 字符串 / text 数组：直接作为 markdown 渲染
//   - 对象 / 数组：包装到 ```json 代码块中，既可读又保留结构
const markdownSource = computed<string>(() => {
  const d = props.data
  if (d == null) return ''
  if (typeof d === 'string') return d
  if (textArrayContent.value !== null) return textArrayContent.value
  try {
    return '```json\n' + JSON.stringify(d, null, 2) + '\n```'
  } catch {
    return String(d)
  }
})

function toggle(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (target.closest('.rd-tool-tabs')) return
  expanded.value = !expanded.value
}
</script>

<template>
  <div class="rd-tool" :class="[`rd-tool-${kind}`, { 'rd-tool-expanded': expanded }]">
    <div class="rd-tool-header" @click="toggle">
      <div class="rd-tool-title">
        <svg width="10" height="10" viewBox="0 0 16 16" fill="none" class="rd-tool-caret">
          <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <span class="rd-tool-kind">{{ kind === 'result' ? 'Tool Result' : 'Tool Call' }}</span>
        <span v-if="name" class="rd-tool-name">{{ name }}</span>
        <span v-if="id" class="rd-tool-id">{{ id }}</span>
      </div>
      <div v-if="expanded" class="rd-tool-tabs">
        <button class="rd-tool-tab" :class="{ active: view === 'markdown' }" @click.stop="view = 'markdown'">Markdown</button>
        <button class="rd-tool-tab" :class="{ active: view === 'text' }" @click.stop="view = 'text'">Text</button>
        <button class="rd-tool-tab" :class="{ active: view === 'json' }" @click.stop="view = 'json'">JSON</button>
      </div>
    </div>
    <div v-if="expanded" class="rd-tool-body">
      <template v-if="view === 'markdown'">
        <div v-if="data == null" class="rd-tool-empty">(empty)</div>
        <RdMarkdown v-else :text="markdownSource" />
      </template>
      <template v-else-if="view === 'text'">
        <div v-if="data == null" class="rd-tool-empty">(empty)</div>
        <RdMarkdown v-else-if="typeof data === 'string'" :text="data" plain />
        <RdMarkdown v-else-if="textArrayContent !== null" :text="textArrayContent" plain />
        <div v-else-if="kvEntries" class="rd-tool-kv">
          <div v-for="[k, v] in kvEntries" :key="k" class="rd-tool-kv-row">
            <div class="rd-tool-kv-key">{{ k }}</div>
            <div class="rd-tool-kv-value" :class="`rd-tool-kv-${valueKind(v)}`">
              <template v-if="valueKind(v) === 'null'">
                <span class="rd-tool-kv-lit">null</span>
              </template>
              <RdMarkdown v-else-if="valueKind(v) === 'string'" :text="v as string" plain />
              <span v-else-if="valueKind(v) === 'scalar'" class="rd-tool-kv-lit">{{ v }}</span>
              <RdJsonTree v-else :data="v" :max-depth="3" :show-copy="false" />
            </div>
          </div>
        </div>
        <RdJsonTree v-else :data="data" :max-depth="4" :show-copy="false" />
      </template>
      <template v-else>
        <RdJsonTree :data="raw ?? data" :max-depth="4" :show-copy="false" />
      </template>
    </div>
  </div>
</template>

<style scoped>
.rd-tool {
  background: rgb(248 250 252);
  border: 1px solid rgb(226 232 240);
  border-left: 2px solid rgb(245 158 11);
  border-radius: 4px;
  overflow: hidden;
  margin-top: 0.5rem;
}

.rd-tool-result {
  border-left-color: rgb(20 184 166);
}

:global(.dark) .rd-tool {
  background: rgb(15 23 42);
  border-color: rgb(51 65 85);
}

.rd-tool-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.4375rem 0.625rem;
  cursor: pointer;
  user-select: none;
  font-size: 0.8125rem;
}

.rd-tool-header:hover {
  background: rgb(241 245 249);
}

:global(.dark) .rd-tool-header:hover {
  background: rgb(30 41 59);
}

.rd-tool-title {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  min-width: 0;
  flex: 1;
}

.rd-tool-caret {
  color: rgb(148 163 184);
  transition: transform 0.15s;
}

.rd-tool-expanded .rd-tool-caret {
  transform: rotate(90deg);
}

.rd-tool-kind {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgb(245 158 11);
}

.rd-tool-result .rd-tool-kind {
  color: rgb(20 184 166);
}

.rd-tool-name {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-weight: 500;
  color: rgb(15 23 42);
}

:global(.dark) .rd-tool-name {
  color: rgb(241 245 249);
}

.rd-tool-id {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.6875rem;
  color: rgb(148 163 184);
}

.rd-tool-tabs {
  display: inline-flex;
  padding: 2px;
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
}

:global(.dark) .rd-tool-tabs {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-tool-tab {
  padding: 0.125rem 0.4375rem;
  font-size: 0.6875rem;
  font-weight: 500;
  color: rgb(100 116 139);
  background: transparent;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.1s;
}

.rd-tool-tab:hover {
  color: rgb(15 23 42);
}

:global(.dark) .rd-tool-tab:hover {
  color: rgb(241 245 249);
}

.rd-tool-tab.active {
  background: rgb(248 250 252);
  color: rgb(15 23 42);
}

:global(.dark) .rd-tool-tab.active {
  background: rgb(51 65 85);
  color: rgb(241 245 249);
}

.rd-tool-body {
  border-top: 1px solid rgb(226 232 240);
  background: white;
}

:global(.dark) .rd-tool-body {
  background: rgb(30 41 59);
  border-top-color: rgb(51 65 85);
}

.rd-tool-empty {
  padding: 0.625rem 0.75rem;
  color: rgb(148 163 184);
  font-size: 0.8125rem;
}

.rd-tool-body :deep(.rd-json-viewer) {
  border: none;
  border-radius: 0;
}

.rd-tool-body :deep(.rd-md) {
  border-radius: 0;
}

/* KV 表格视图：两列 grid，key 列按内容自适应但不超过上限，超出后换行 */
.rd-tool-kv {
  display: grid;
  /* fit-content(MAX)：max-content 与 MAX 取较小值；超过 MAX 则强制换行 */
  grid-template-columns: fit-content(220px) minmax(0, 1fr);
  font-size: 0.8125rem;
  background: rgb(226 232 240);
  gap: 1px;
}

:global(.dark) .rd-tool-kv {
  background: rgb(51 65 85);
}

.rd-tool-kv-row {
  display: contents;
}

.rd-tool-kv-key {
  padding: 0.4375rem 0.75rem;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-weight: 500;
  color: rgb(37 99 235);
  background: rgb(248 250 252);
  word-break: break-word;
  overflow-wrap: anywhere;
  line-height: 1.55;
}

:global(.dark) .rd-tool-kv-key {
  background: rgb(15 23 42);
  color: rgb(96 165 250);
}

.rd-tool-kv-value {
  padding: 0.4375rem 0.75rem;
  min-width: 0;
  background: white;
  line-height: 1.55;
  overflow-wrap: anywhere;
  word-break: break-word;
}

:global(.dark) .rd-tool-kv-value {
  background: rgb(30 41 59);
}

.rd-tool-kv-value.rd-tool-kv-string,
.rd-tool-kv-value.rd-tool-kv-object {
  padding: 0;
}

.rd-tool-kv-value :deep(.rd-md),
.rd-tool-kv-value :deep(.rd-json-viewer) {
  background: transparent;
  max-height: 360px;
}

.rd-tool-kv-lit {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  color: rgb(22 163 74);
}

.rd-tool-kv-value.rd-tool-kv-null .rd-tool-kv-lit {
  color: rgb(148 163 184);
}
</style>
