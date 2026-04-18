<script setup lang="ts">
import { ref } from 'vue'
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

withDefaults(defineProps<Props>(), {
  name: '',
  id: null,
  raw: undefined
})

const expanded = ref(true)
const view = ref<'parsed' | 'json'>('parsed')

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
        <button class="rd-tool-tab" :class="{ active: view === 'parsed' }" @click.stop="view = 'parsed'">Parsed</button>
        <button class="rd-tool-tab" :class="{ active: view === 'json' }" @click.stop="view = 'json'">JSON</button>
      </div>
    </div>
    <div v-if="expanded" class="rd-tool-body">
      <template v-if="view === 'parsed'">
        <div v-if="data == null" class="rd-tool-empty">(empty)</div>
        <RdMarkdown v-else-if="typeof data === 'string'" :text="data" plain />
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
</style>
