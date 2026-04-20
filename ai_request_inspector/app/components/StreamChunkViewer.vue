<script setup lang="ts">
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
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
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
      if (chunk.type === 'text-delta' && chunk.parsed.textDelta) {
        text += chunk.parsed.textDelta
      }
      if (chunk.type === 'tool-call') {
        toolCalls.push(chunk.parsed)
      }
    }
  }
  
  return { text, toolCalls }
})
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Assembled Content Preview -->
    <div v-if="viewMode === 'parsed'" class="border-b border-default p-4 bg-muted/30">
      <div class="flex items-center gap-2 mb-3">
        <UIcon name="i-lucide-layers" class="size-4 text-primary" />
        <span class="text-sm font-medium text-default">组装后内容</span>
      </div>
      
      <div v-if="assembledContent.text" class="bg-elevated rounded-lg p-3 mb-3">
        <p class="text-sm text-default whitespace-pre-wrap leading-relaxed">{{ assembledContent.text }}</p>
      </div>
      
      <div v-if="assembledContent.toolCalls.length > 0" class="space-y-2">
        <div class="text-xs text-muted mb-2">工具调用</div>
        <div 
          v-for="(tool, index) in assembledContent.toolCalls" 
          :key="index"
          class="bg-elevated rounded-lg p-3"
        >
          <div class="flex items-center gap-2 mb-2">
            <UIcon name="i-lucide-wrench" class="size-4 text-amber-400" />
            <span class="text-sm font-medium text-default">{{ tool.toolName || 'Unknown Tool' }}</span>
          </div>
          <JsonViewer :data="tool.args || tool" :search-query="searchQuery" />
        </div>
      </div>
      
      <div v-if="!assembledContent.text && assembledContent.toolCalls.length === 0" class="text-muted text-sm">
        暂无可组装的内容
      </div>
    </div>
    
    <!-- Chunk List -->
    <div class="flex-1 overflow-auto custom-scrollbar">
      <div class="divide-y divide-muted/30">
        <div 
          v-for="chunk in chunks" 
          :key="chunk.id"
          class="group"
        >
          <!-- Chunk Header -->
          <div 
            class="flex items-center gap-3 px-4 py-2 cursor-pointer hover:bg-muted/30 transition-colors"
            :class="{ 'bg-muted/40': selectedChunk === chunk.id }"
            @click="selectedChunk = selectedChunk === chunk.id ? null : chunk.id"
          >
            <UIcon 
              :name="selectedChunk === chunk.id ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'" 
              class="size-4 text-muted"
            />
            
            <div class="flex items-center gap-2 min-w-0 flex-1">
              <UIcon 
                :name="getChunkTypeIcon(chunk.type)" 
                :class="['size-4 shrink-0', getChunkTypeColor(chunk.type)]"
              />
              <span class="text-xs font-mono text-muted">Chunk #{{ chunk.id }}</span>
              <UBadge 
                v-if="chunk.type" 
                :label="chunk.type" 
                size="xs" 
                variant="subtle"
                color="neutral"
              />
            </div>
            
            <span class="text-xs text-dimmed shrink-0">{{ chunk.timestamp }}</span>
          </div>
          
          <!-- Chunk Content -->
          <div v-if="selectedChunk === chunk.id" class="px-4 pb-4 bg-muted/20">
            <UTabs
              :items="[
                { label: '原始数据', slot: 'raw', icon: 'i-lucide-file-text' },
                { label: '解析结果', slot: 'parsed', icon: 'i-lucide-braces' }
              ]"
              default-value="raw"
              class="mt-2"
            >
              <template #raw>
                <div class="mt-3 bg-elevated rounded-lg p-3 overflow-x-auto custom-scrollbar">
                  <pre class="text-xs font-mono text-muted whitespace-pre-wrap break-all" v-html="highlightText(chunk.raw, searchQuery)" />
                </div>
              </template>
              
              <template #parsed>
                <div class="mt-3 bg-elevated rounded-lg p-3 overflow-x-auto custom-scrollbar">
                  <JsonViewer 
                    v-if="chunk.parsed" 
                    :data="chunk.parsed" 
                    :search-query="searchQuery"
                  />
                  <span v-else class="text-xs text-muted">无法解析</span>
                </div>
              </template>
            </UTabs>
          </div>
        </div>
      </div>
      
      <div v-if="chunks.length === 0" class="flex flex-col items-center justify-center h-full text-muted py-12">
        <UIcon name="i-lucide-inbox" class="size-12 mb-3 opacity-50" />
        <span class="text-sm">暂无流式数据</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-highlight {
  background-color: #fcd34d;
  color: #111827;
  padding: 0 2px;
  border-radius: 2px;
}

:global(.dark) .search-highlight {
  background-color: #78350f;
  color: #f9fafb;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background-color: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: #d1d5db;
  border-radius: 3px;
}

:global(.dark) .custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: #374151;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background-color: #9ca3af;
}

:global(.dark) .custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background-color: #4b5563;
}
</style>
