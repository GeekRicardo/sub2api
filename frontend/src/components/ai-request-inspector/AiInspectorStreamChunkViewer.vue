<script setup lang="ts">
import { computed, ref } from 'vue'
import AiInspectorIcon from './AiInspectorIcon.vue'
import AiInspectorJsonViewer from './AiInspectorJsonViewer.vue'

interface Chunk {
  id: number
  timestamp: string
  raw: string
  parsed?: any
  type?: string
}

interface Props {
  chunks: Chunk[]
  viewMode: 'raw' | 'parsed'
  searchQuery?: string
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: ''
})

const selectedChunk = ref<number | null>(null)

const highlightText = (text: string, query: string): string => {
  if (!query) return escapeHtml(text)
  const escapedQuery = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escapedQuery})`, 'gi')
  return escapeHtml(text).replace(regex, '<span class="search-highlight">$1</span>')
}

const escapeHtml = (text: string): string => {
  return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

const getChunkTypeIcon = (type?: string): string => {
  switch (type) {
    case 'text-delta': return 'i-lucide-type'
    case 'tool-call': return 'i-lucide-wrench'
    case 'tool-result': return 'i-lucide-check-circle'
    case 'finish': return 'i-lucide-flag'
    case 'error': return 'i-lucide-alert-circle'
    default: return 'i-lucide-code'
  }
}

const getChunkTypeColor = (type?: string): string => {
  switch (type) {
    case 'text-delta': return 'text-sky-400'
    case 'tool-call': return 'text-amber-400'
    case 'tool-result': return 'text-emerald-400'
    case 'finish': return 'text-violet-400'
    case 'error': return 'text-red-400'
    default: return 'text-muted'
  }
}

const assembledContent = computed(() => {
  let text = ''
  const toolCalls: any[] = []
  for (const chunk of props.chunks) {
    if (chunk.parsed) {
      if (chunk.type === 'text-delta' && chunk.parsed.textDelta) text += chunk.parsed.textDelta
      if (chunk.type === 'tool-call') toolCalls.push(chunk.parsed)
    }
  }
  return { text, toolCalls }
})
</script>

<template>
  <div class="flex h-full flex-col">
    <div v-if="viewMode === 'parsed'" class="border-b border-default bg-muted/30 p-4">
      <div class="mb-3 flex items-center gap-2">
        <AiInspectorIcon name="i-lucide-layers" class="size-4 text-primary" />
        <span class="text-sm font-medium text-default">组装后内容</span>
      </div>

      <div v-if="assembledContent.text" class="mb-3 rounded-lg bg-elevated p-3">
        <p class="whitespace-pre-wrap text-sm leading-relaxed text-default">{{ assembledContent.text }}</p>
      </div>

      <div v-if="assembledContent.toolCalls.length > 0" class="space-y-2">
        <div class="mb-2 text-xs text-muted">工具调用</div>
        <div
          v-for="(tool, index) in assembledContent.toolCalls"
          :key="index"
          class="rounded-lg bg-elevated p-3"
        >
          <div class="mb-2 flex items-center gap-2">
            <AiInspectorIcon name="i-lucide-wrench" class="size-4 text-amber-400" />
            <span class="text-sm font-medium text-default">{{ tool.toolName || 'Unknown Tool' }}</span>
          </div>
          <AiInspectorJsonViewer :data="tool.args || tool" :search-query="searchQuery" />
        </div>
      </div>

      <div v-if="!assembledContent.text && assembledContent.toolCalls.length === 0" class="text-sm text-muted">
        暂无可组装的内容
      </div>
    </div>

    <div class="custom-scrollbar flex-1 overflow-auto">
      <div class="divide-y divide-muted/30">
        <div v-for="chunk in chunks" :key="chunk.id" class="group">
          <div
            class="flex cursor-pointer items-center gap-3 px-4 py-2 transition-colors hover:bg-muted/30"
            :class="{ 'bg-muted/40': selectedChunk === chunk.id }"
            @click="selectedChunk = selectedChunk === chunk.id ? null : chunk.id"
          >
            <AiInspectorIcon :name="selectedChunk === chunk.id ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'" class="size-4 text-muted" />

            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <AiInspectorIcon :name="getChunkTypeIcon(chunk.type)" :class="['size-4 shrink-0', getChunkTypeColor(chunk.type)]" />
                <span class="font-mono text-xs text-muted">Chunk #{{ chunk.id }}</span>
                <span v-if="chunk.type" class="rounded-full bg-default/50 px-2 py-1 text-[11px] text-muted">{{ chunk.type }}</span>
              </div>
            </div>

            <span class="shrink-0 text-xs text-dimmed">{{ chunk.timestamp }}</span>
          </div>

          <div v-if="selectedChunk === chunk.id" class="bg-muted/20 px-4 pb-4">
            <div class="mt-2 inline-flex overflow-hidden rounded-lg border border-default">
              <button class="px-3 py-1.5 text-xs font-medium text-default">原始数据</button>
            </div>
            <div class="custom-scrollbar mt-3 overflow-x-auto rounded-lg bg-elevated p-3">
              <pre class="whitespace-pre-wrap break-all font-mono text-xs text-muted" v-html="highlightText(chunk.raw, searchQuery)" />
            </div>
            <div class="mt-3">
              <div class="mb-2 text-xs text-muted">解析结果</div>
              <div class="custom-scrollbar overflow-x-auto rounded-lg bg-elevated p-3">
                <AiInspectorJsonViewer v-if="chunk.parsed" :data="chunk.parsed" :search-query="searchQuery" />
                <span v-else class="text-xs text-muted">无法解析</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="chunks.length === 0" class="flex h-full flex-col items-center justify-center py-12 text-muted">
        <AiInspectorIcon name="i-lucide-inbox" class="mb-3 size-12 opacity-50" />
        <span class="text-sm">暂无流式数据</span>
      </div>
    </div>
  </div>
</template>
