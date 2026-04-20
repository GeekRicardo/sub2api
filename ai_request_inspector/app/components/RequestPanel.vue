<script setup lang="ts">
interface Props {
  data: any
  title: string
  icon: string
  searchQuery?: string
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: ''
})

const viewMode = ref<'formatted' | 'raw'>('formatted')
const copied = ref(false)

// Check if data has messages array (request format)
const hasMessages = computed(() => {
  return Array.isArray(props.data?.messages) && props.data.messages.length > 0
})

// Check if data has choices with messages (response format)
const hasChoicesMessages = computed(() => {
  return Array.isArray(props.data?.choices) && 
         props.data.choices.some((choice: any) => choice.message)
})

// Extract messages from choices
const choicesMessages = computed(() => {
  if (!hasChoicesMessages.value) return []
  return props.data.choices
    .filter((choice: any) => choice.message)
    .map((choice: any) => choice.message)
})

// Get other parameters (excluding messages) for table view
const otherParams = computed(() => {
  if (!props.data) return []
  
  const result: { key: string; value: any; type: string }[] = []
  const excludeKeys = hasMessages.value ? ['messages'] : hasChoicesMessages.value ? ['choices'] : []
  
  Object.entries(props.data).forEach(([key, value]) => {
    if (excludeKeys.includes(key)) return
    
    const type = Array.isArray(value) ? 'array' : typeof value
    result.push({ key, value, type })
  })
  
  // If response format, add choices metadata (without messages)
  if (hasChoicesMessages.value && props.data.choices) {
    const choicesMeta = props.data.choices.map((choice: any, idx: number) => ({
      index: choice.index ?? idx,
      finish_reason: choice.finish_reason
    }))
    result.push({ key: 'finish_reasons', value: choicesMeta, type: 'array' })
  }
  
  return result
})

const rawJson = computed(() => {
  try {
    return JSON.stringify(props.data, null, 2)
  } catch {
    return ''
  }
})

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(rawJson.value)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  } catch (e) {
    console.error('复制失败', e)
  }
}

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

const formatValue = (value: any): string => {
  if (value === null) return 'null'
  if (value === undefined) return 'undefined'
  if (typeof value === 'object') {
    return JSON.stringify(value, null, 2)
  }
  return String(value)
}

const isSimpleValue = (value: any): boolean => {
  if (value === null || value === undefined) return true
  if (typeof value === 'object') return false
  return true
}

const getTypeColor = (type: string): string => {
  const colors: Record<string, string> = {
    'string': 'blue',
    'number': 'green',
    'boolean': 'orange',
    'array': 'purple',
    'object': 'indigo'
  }
  return colors[type] || 'neutral'
}
</script>

<template>
  <div class="h-full flex flex-col bg-default">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-default bg-muted/30">
      <div class="flex items-center gap-2">
        <UIcon :name="icon" class="size-5 text-primary" />
        <h3 class="font-medium text-default">{{ title }}</h3>
      </div>
      
      <div class="flex items-center gap-2">
        <div class="inline-flex rounded-lg border border-default overflow-hidden">
          <UButton
            :variant="viewMode === 'formatted' ? 'solid' : 'ghost'"
            size="xs"
            icon="i-lucide-layout"
            class="rounded-none"
            @click="viewMode = 'formatted'"
          >
            格式化
          </UButton>
          <UButton
            :variant="viewMode === 'raw' ? 'solid' : 'ghost'"
            size="xs"
            icon="i-lucide-file-json"
            class="rounded-none"
            @click="viewMode = 'raw'"
          >
            原始 JSON
          </UButton>
        </div>
        
        <UButton
          variant="ghost"
          size="xs"
          :icon="copied ? 'i-lucide-check' : 'i-lucide-copy'"
          :color="copied ? 'success' : 'neutral'"
          @click="copyToClipboard"
        >
          {{ copied ? '已复制' : '复制' }}
        </UButton>
      </div>
    </div>
    
    <!-- Content -->
    <div class="flex-1 overflow-auto custom-scrollbar p-6">
      <!-- Formatted View -->
      <template v-if="viewMode === 'formatted'">
        <template v-if="data">
          <!-- Messages Section -->
          <div v-if="hasMessages || hasChoicesMessages">
            <div class="flex items-center gap-3 mb-6">
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-messages-square" class="size-6 text-primary" />
                <div>
                  <h4 class="text-base font-semibold text-default">Messages</h4>
                  <p class="text-xs text-muted">{{ hasMessages ? data.messages.length : choicesMessages.length }} 条消息</p>
                </div>
              </div>
            </div>
            
            <MessageList 
              :messages="hasMessages ? data.messages : choicesMessages" 
              :search-query="searchQuery"
            />
          </div>
          
          <!-- Other Parameters Table -->
          <div v-if="otherParams.length > 0" class="mt-10">
            <div class="mb-6">
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-settings-2" class="size-6 text-muted" />
                <h4 class="text-base font-semibold text-default">请求参数</h4>
              </div>
            </div>
            
            <div class="border border-default rounded-lg overflow-hidden bg-muted/20">
              <table class="w-full">
                <thead>
                  <tr class="border-b border-default bg-muted/40">
                    <th class="px-5 py-3 text-left text-xs font-semibold text-muted uppercase tracking-wider">参数</th>
                    <th class="px-5 py-3 text-left text-xs font-semibold text-muted uppercase tracking-wider w-20">类型</th>
                    <th class="px-5 py-3 text-left text-xs font-semibold text-muted uppercase tracking-wider">值</th>
                  </tr>
                </thead>
                <tbody>
                  <tr 
                    v-for="(param, index) in otherParams" 
                    :key="param.key"
                    :class="[
                      'border-b border-default hover:bg-muted/40 transition-colors',
                      index === otherParams.length - 1 ? 'border-b-0' : ''
                    ]"
                  >
                    <!-- Parameter Name -->
                    <td class="px-5 py-4">
                      <code class="text-sm font-mono text-default bg-elevated/50 px-2.5 py-1 rounded">
                        {{ param.key }}
                      </code>
                    </td>
                    
                    <!-- Type Badge -->
                    <td class="px-5 py-4">
                      <UBadge 
                        :label="param.type" 
                        :color="getTypeColor(param.type)"
                        variant="subtle" 
                        size="sm" 
                      />
                    </td>
                    
                    <!-- Value -->
                    <td class="px-5 py-4">
                      <template v-if="isSimpleValue(param.value)">
                        <code 
                          class="text-sm font-mono"
                          :class="[
                            typeof param.value === 'boolean' ? 'text-orange-600 dark:text-orange-400' :
                            typeof param.value === 'number' ? 'text-green-600 dark:text-green-400' :
                            typeof param.value === 'string' ? 'text-blue-600 dark:text-blue-400' :
                            param.value === null ? 'text-muted italic' :
                            'text-default'
                          ]"
                        >
                          {{ formatValue(param.value) }}
                        </code>
                      </template>
                      
                      <template v-else>
                        <details class="group">
                          <summary class="cursor-pointer text-sm text-muted hover:text-default transition-colors inline-flex items-center gap-1.5 select-none">
                            <UIcon name="i-lucide-chevron-right" class="size-4 transition-transform group-open:rotate-90" />
                            <span>{{ Array.isArray(param.value) ? `数组 (${param.value.length})` : `对象 (${Object.keys(param.value).length})` }}</span>
                          </summary>
                          <div class="mt-3 bg-elevated rounded-lg p-4 border border-default max-h-72 overflow-auto custom-scrollbar">
                            <pre 
                              class="text-xs font-mono text-muted whitespace-pre-wrap break-all leading-relaxed"
                              v-html="highlightText(formatValue(param.value), searchQuery)"
                            />
                          </div>
                        </details>
                      </template>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          
          <!-- No Data -->
          <div v-if="!hasMessages && !hasChoicesMessages && otherParams.length === 0" class="flex flex-col items-center justify-center py-12 text-muted">
            <UIcon name="i-lucide-file-x" class="size-12 mb-3 opacity-50" />
            <span class="text-sm">暂无数据</span>
          </div>
        </template>
        
        <div v-else class="flex flex-col items-center justify-center h-full text-muted">
          <UIcon name="i-lucide-file-x" class="size-12 mb-3 opacity-50" />
          <span class="text-sm">暂无数据</span>
        </div>
      </template>
      
      <!-- Raw JSON View -->
      <template v-else>
        <div class="bg-elevated rounded-lg p-4 overflow-x-auto border border-default">
          <pre 
            v-if="rawJson" 
            class="text-xs font-mono text-muted whitespace-pre-wrap break-all leading-relaxed"
            v-html="highlightText(rawJson, searchQuery)"
          />
          <span v-else class="text-muted text-sm">暂无数据</span>
        </div>
      </template>
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
