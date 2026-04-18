<script setup lang="ts">
import { computed, ref } from 'vue'
import RdJsonNode from './RdJsonNode.vue'

interface Props {
  data: unknown
  maxDepth?: number
  showCopy?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxDepth: 3,
  showCopy: true
})

const copied = ref(false)

const rawJson = computed(() => {
  try {
    return JSON.stringify(props.data, null, 2)
  } catch {
    return String(props.data)
  }
})

async function copy() {
  try {
    await navigator.clipboard.writeText(rawJson.value)
    copied.value = true
    setTimeout(() => (copied.value = false), 1500)
  } catch {
    /* ignore */
  }
}
</script>

<template>
  <div class="rd-json-viewer">
    <div v-if="showCopy" class="rd-json-header">
      <button class="rd-json-copy" @click="copy">
        <svg v-if="!copied" width="12" height="12" viewBox="0 0 16 16" fill="none">
          <path d="M4 4h8v8H4V4zm1 1v6h6V5H5z" fill="currentColor" />
          <path d="M2 2h8v1H3v7H2V2z" fill="currentColor" />
        </svg>
        <svg v-else width="12" height="12" viewBox="0 0 16 16" fill="none">
          <path d="M13 4L6 11 3 8" stroke="currentColor" stroke-width="2" fill="none" />
        </svg>
        {{ copied ? 'Copied' : 'Copy' }}
      </button>
    </div>
    <div class="rd-json-tree">
      <RdJsonNode :data="data" :depth="0" :max-depth="maxDepth" />
    </div>
  </div>
</template>

<style scoped>
.rd-json-viewer {
  border: 1px solid rgb(226 232 240);
  border-radius: 6px;
  overflow: hidden;
  background: white;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
}

:global(.dark) .rd-json-viewer {
  background: rgb(30 41 59);
  border-color: rgb(51 65 85);
}

.rd-json-header {
  display: flex;
  justify-content: flex-end;
  padding: 0.375rem 0.625rem;
  background: rgb(248 250 252);
  border-bottom: 1px solid rgb(226 232 240);
}

:global(.dark) .rd-json-header {
  background: rgb(15 23 42);
  border-color: rgb(51 65 85);
}

.rd-json-copy {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.1875rem 0.5rem;
  background: transparent;
  border: 1px solid rgb(226 232 240);
  border-radius: 4px;
  color: rgb(100 116 139);
  font-size: 0.6875rem;
  font-family: inherit;
  cursor: pointer;
}

.rd-json-copy:hover {
  color: rgb(15 23 42);
}

:global(.dark) .rd-json-copy {
  border-color: rgb(51 65 85);
  color: rgb(148 163 184);
}

:global(.dark) .rd-json-copy:hover {
  color: rgb(241 245 249);
}

.rd-json-tree {
  padding: 0.625rem 0.75rem;
  overflow-x: auto;
  max-height: 640px;
  overflow-y: auto;
  line-height: 1.6;
  color: rgb(15 23 42);
}

:global(.dark) .rd-json-tree {
  color: rgb(241 245 249);
}

.rd-json-tree :deep(.rd-json-row) {
  margin-left: 1.25rem;
  padding: 0.0625rem 0;
}

.rd-json-tree :deep(.rd-json-key) {
  color: #8b5cf6;
  font-weight: 500;
  margin-right: 0.375rem;
}

.rd-json-tree :deep(.rd-json-str) {
  color: #059669;
  word-break: break-word;
}

.rd-json-tree :deep(.rd-json-num) {
  color: #d97706;
}

.rd-json-tree :deep(.rd-json-bool) {
  color: #2563eb;
  font-weight: 600;
}

.rd-json-tree :deep(.rd-json-null) {
  color: #94a3b8;
  font-style: italic;
}

.rd-json-tree :deep(.rd-json-bracket) {
  color: rgb(100 116 139);
}

.rd-json-tree :deep(.rd-json-toggle) {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-left: 0.25rem;
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  font-family: inherit;
  font-size: 0.6875rem;
  padding: 0;
}

.rd-json-tree :deep(.rd-json-caret) {
  font-size: 0.625rem;
}

.rd-json-tree :deep(.rd-json-count) {
  font-style: italic;
}

:global(.dark) .rd-json-tree :deep(.rd-json-key) {
  color: #a78bfa;
}

:global(.dark) .rd-json-tree :deep(.rd-json-str) {
  color: #34d399;
}

:global(.dark) .rd-json-tree :deep(.rd-json-num) {
  color: #fbbf24;
}
</style>
