<script setup lang="ts">
import { computed, ref } from 'vue'
import RdJsonTree from './RdJsonTree.vue'
import RdMarkdown from './RdMarkdown.vue'
import RdToolBlock from './RdToolBlock.vue'

interface AnyMessage {
  role?: string
  content?: unknown
  tool_calls?: Array<Record<string, any>>
  tool_call_id?: string
  [k: string]: unknown
}

interface Props {
  message: AnyMessage
  index?: number
  /** 强制展示的 role 名（用于独立 tool_result 等场景） */
  roleLabel?: string
  /** 顶部 id 文本（如 tool_use_id） */
  headerId?: string
  /** 默认 parsed/json */
  defaultView?: 'parsed' | 'json'
  /** 默认是否折叠 */
  defaultCollapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  index: 0,
  roleLabel: '',
  headerId: '',
  defaultView: 'parsed',
  defaultCollapsed: false
})

const view = ref<'parsed' | 'json'>(props.defaultView)
const collapsed = ref(props.defaultCollapsed)

const role = computed(() => props.roleLabel || props.message.role || 'unknown')

const textContent = computed(() => {
  const c = props.message.content
  if (typeof c === 'string') return c
  if (Array.isArray(c)) {
    return c
      .filter((i: any) => i?.type === 'text' && i.text)
      .map((i: any) => i.text)
      .join('\n\n')
  }
  return ''
})

const hasText = computed(() => !!textContent.value && textContent.value.trim().length > 0)

interface Extra {
  kind: 'call' | 'result' | 'image'
  name?: string
  id?: string
  data?: unknown
  raw?: unknown
}

const extras = computed<Extra[]>(() => {
  const list: Extra[] = []
  const c = props.message.content
  if (Array.isArray(c)) {
    c.forEach((item: any) => {
      if (!item) return
      if (item.type === 'image_url' || item.type === 'image') {
        list.push({ kind: 'image' })
      } else if (item.type === 'tool_use') {
        list.push({ kind: 'call', name: item.name, id: item.id, data: item.input, raw: item })
      } else if (item.type === 'tool_result') {
        let data = item.content
        if (typeof data === 'string') {
          try { data = JSON.parse(data) } catch { /* keep string */ }
        }
        list.push({ kind: 'result', name: `result:${item.tool_use_id || ''}`, id: item.tool_use_id, data, raw: item })
      }
    })
  }
  if (Array.isArray(props.message.tool_calls)) {
    props.message.tool_calls.forEach((tc: any) => {
      let args: any = tc.function?.arguments ?? tc.arguments ?? '{}'
      try { args = typeof args === 'string' ? JSON.parse(args) : args } catch { /* keep */ }
      list.push({
        kind: 'call',
        name: tc.function?.name || tc.name,
        id: tc.id,
        data: args,
        raw: tc
      })
    })
  }
  return list
})

const hasBody = computed(() => hasText.value || extras.value.length > 0)

const preview = computed(() => {
  const first = textContent.value.split('\n')[0] || ''
  return first.slice(0, 120)
})
</script>

<template>
  <div class="rd-msg" :class="[`rd-msg-${role}`, { 'rd-msg-collapsed': collapsed }]">
    <div class="rd-msg-header" @click="collapsed = !collapsed">
      <button class="rd-msg-collapse" @click.stop="collapsed = !collapsed">
        <svg width="10" height="10" viewBox="0 0 16 16" fill="none" class="rd-msg-caret">
          <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
      <span class="rd-msg-role">{{ role }}</span>
      <span v-if="headerId" class="rd-msg-id">{{ headerId }}</span>
      <span v-if="collapsed && preview" class="rd-msg-preview">
        {{ preview }}{{ textContent.length > 120 ? '…' : '' }}
      </span>
      <div v-if="hasBody && !collapsed" class="rd-msg-tabs" @click.stop>
        <button class="rd-msg-tab" :class="{ active: view === 'parsed' }" @click="view = 'parsed'">Parsed</button>
        <button class="rd-msg-tab" :class="{ active: view === 'json' }" @click="view = 'json'">JSON</button>
      </div>
    </div>

    <div v-if="!collapsed" class="rd-msg-body">
      <template v-if="hasBody && view === 'parsed'">
        <RdMarkdown v-if="hasText" :text="textContent" />
        <div v-if="extras.length" class="rd-msg-extras">
          <template v-for="(ex, i) in extras" :key="i">
            <div v-if="ex.kind === 'image'" class="rd-msg-image-ph">[image attachment]</div>
            <RdToolBlock v-else :kind="ex.kind" :name="ex.name" :id="ex.id" :data="ex.data" :raw="ex.raw" />
          </template>
        </div>
      </template>
      <template v-else-if="hasBody && view === 'json'">
        <div class="rd-msg-json-wrap">
          <RdJsonTree :data="message" :max-depth="3" :show-copy="false" />
        </div>
      </template>
      <div v-else class="rd-msg-json-wrap">
        <RdJsonTree :data="message" :max-depth="3" :show-copy="false" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.rd-msg {
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 6px;
  margin-bottom: 0.625rem;
  overflow: hidden;
}

:global(.dark) .rd-msg {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-msg-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4375rem 0.75rem;
  background: rgb(248 250 252);
  border-bottom: 1px solid rgb(226 232 240);
  cursor: pointer;
}

.rd-msg-collapsed .rd-msg-header {
  border-bottom: none;
}

:global(.dark) .rd-msg-header {
  background: rgb(15 23 42);
  border-bottom-color: rgb(51 65 85);
}

.rd-msg-collapse {
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

.rd-msg-collapse:hover {
  background: rgb(241 245 249);
  color: rgb(15 23 42);
}

:global(.dark) .rd-msg-collapse:hover {
  background: rgb(51 65 85);
  color: rgb(241 245 249);
}

.rd-msg-caret {
  transition: transform 0.15s;
  transform: rotate(90deg);
}

.rd-msg-collapsed .rd-msg-caret {
  transform: rotate(0deg);
}

.rd-msg-role {
  display: inline-flex;
  align-items: center;
  gap: 0.3125rem;
  font-size: 0.71875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgb(100 116 139);
}

.rd-msg-role::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.rd-msg-user .rd-msg-role { color: rgb(20 184 166); }
.rd-msg-assistant .rd-msg-role { color: rgb(16 185 129); }
.rd-msg-system .rd-msg-role { color: rgb(139 92 246); }
.rd-msg-tool .rd-msg-role { color: rgb(245 158 11); }

.rd-msg-id {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.6875rem;
  color: rgb(148 163 184);
}

.rd-msg-preview {
  flex: 1;
  color: rgb(100 116 139);
  font-size: 0.8125rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
  margin-left: 0.5rem;
}

.rd-msg-tabs {
  display: inline-flex;
  padding: 2px;
  background: white;
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
  margin-left: auto;
}

:global(.dark) .rd-msg-tabs {
  background: rgb(15 23 42);
  border-color: rgb(51 65 85);
}

.rd-msg-tab {
  padding: 0.125rem 0.4375rem;
  font-size: 0.6875rem;
  font-weight: 500;
  color: rgb(100 116 139);
  background: transparent;
  border: none;
  border-radius: 3px;
  cursor: pointer;
}

.rd-msg-tab:hover {
  color: rgb(15 23 42);
}

:global(.dark) .rd-msg-tab:hover {
  color: rgb(241 245 249);
}

.rd-msg-tab.active {
  background: rgb(248 250 252);
  color: rgb(15 23 42);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

:global(.dark) .rd-msg-tab.active {
  background: rgb(51 65 85);
  color: rgb(241 245 249);
}

.rd-msg-body {
  padding: 0;
}

.rd-msg-extras {
  padding: 0.5rem 0.75rem;
}

.rd-msg-image-ph {
  padding: 0.5rem 0.75rem;
  font-size: 0.8125rem;
  color: rgb(100 116 139);
  background: rgb(248 250 252);
  border-radius: 4px;
  margin-bottom: 0.5rem;
}

.rd-msg-json-wrap {
  padding: 0.5rem 0.75rem;
}

.rd-msg-json-wrap :deep(.rd-json-viewer) {
  max-height: 480px;
}
</style>
