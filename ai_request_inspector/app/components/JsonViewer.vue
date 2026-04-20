<script setup lang="ts">
interface Props {
  data: any
  searchQuery?: string
  level?: number
  path?: string
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: '',
  level: 0,
  path: ''
})

const emit = defineEmits<{
  pathClick: [path: string]
}>()

const collapsedKeys = ref<Set<string>>(new Set())

const isObject = (val: any): val is object => val !== null && typeof val === 'object' && !Array.isArray(val)
const isArray = (val: any): val is any[] => Array.isArray(val)
const isPrimitive = (val: any) => !isObject(val) && !isArray(val)

const getType = (val: any): string => {
  if (val === null) return 'null'
  if (val === undefined) return 'undefined'
  if (isArray(val)) return 'array'
  if (isObject(val)) return 'object'
  return typeof val
}

const toggleCollapse = (key: string) => {
  const fullPath = props.path ? `${props.path}.${key}` : key
  if (collapsedKeys.value.has(fullPath)) {
    collapsedKeys.value.delete(fullPath)
  } else {
    collapsedKeys.value.add(fullPath)
  }
}

const isCollapsed = (key: string): boolean => {
  const fullPath = props.path ? `${props.path}.${key}` : key
  return collapsedKeys.value.has(fullPath)
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
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

const formatValue = (val: any): string => {
  if (val === null) return 'null'
  if (val === undefined) return 'undefined'
  if (typeof val === 'string') return `"${val}"`
  return String(val)
}

const getValueClass = (val: any): string => {
  const type = getType(val)
  switch (type) {
    case 'string': return 'json-string'
    case 'number': return 'json-number'
    case 'boolean': return 'json-boolean'
    case 'null':
    case 'undefined': return 'json-null'
    default: return ''
  }
}

const getPreview = (val: any): string => {
  if (isArray(val)) {
    return `Array(${val.length})`
  }
  if (isObject(val)) {
    const keys = Object.keys(val)
    if (keys.length === 0) return '{}'
    if (keys.length <= 3) {
      return `{ ${keys.join(', ')} }`
    }
    return `{ ${keys.slice(0, 3).join(', ')}, ... }`
  }
  return ''
}

const indentStyle = computed(() => ({
  paddingLeft: `${props.level * 16}px`
}))
</script>

<template>
  <div class="font-mono text-sm">
    <template v-if="isPrimitive(data)">
      <span :class="getValueClass(data)" v-html="highlightText(formatValue(data), searchQuery)" />
    </template>
    
    <template v-else-if="isArray(data)">
      <div v-for="(item, index) in data" :key="index" class="relative">
        <div 
          class="flex items-start gap-1 py-0.5 hover:bg-muted/50 rounded cursor-pointer group"
          :style="indentStyle"
          @click="toggleCollapse(String(index))"
        >
          <UIcon 
            v-if="!isPrimitive(item)"
            :name="isCollapsed(String(index)) ? 'i-lucide-chevron-right' : 'i-lucide-chevron-down'" 
            class="size-4 text-muted shrink-0 mt-0.5 opacity-60 group-hover:opacity-100 transition-opacity"
          />
          <span v-else class="size-4 shrink-0" />
          
          <span class="json-bracket">[</span>
          <span class="json-number" v-html="highlightText(String(index), searchQuery)" />
          <span class="json-bracket">]</span>
          <span class="text-muted mx-1">:</span>
          
          <template v-if="isPrimitive(item)">
            <span :class="getValueClass(item)" v-html="highlightText(formatValue(item), searchQuery)" />
          </template>
          <template v-else-if="isCollapsed(String(index))">
            <span class="text-muted">{{ getPreview(item) }}</span>
          </template>
        </div>
        
        <template v-if="!isPrimitive(item) && !isCollapsed(String(index))">
          <JsonViewer 
            :data="item" 
            :search-query="searchQuery" 
            :level="level + 1"
            :path="path ? `${path}.${index}` : String(index)"
            @path-click="emit('pathClick', $event)"
          />
        </template>
      </div>
    </template>
    
    <template v-else-if="isObject(data)">
      <div v-for="(value, key) in data" :key="key" class="relative">
        <div 
          class="flex items-start gap-1 py-0.5 hover:bg-muted/50 rounded cursor-pointer group"
          :style="indentStyle"
          @click="toggleCollapse(String(key))"
        >
          <UIcon 
            v-if="!isPrimitive(value)"
            :name="isCollapsed(String(key)) ? 'i-lucide-chevron-right' : 'i-lucide-chevron-down'" 
            class="size-4 text-muted shrink-0 mt-0.5 opacity-60 group-hover:opacity-100 transition-opacity"
          />
          <span v-else class="size-4 shrink-0" />
          
          <span class="json-key" v-html="highlightText(String(key), searchQuery)" />
          <span class="text-muted mx-1">:</span>
          
          <template v-if="isPrimitive(value)">
            <span :class="getValueClass(value)" v-html="highlightText(formatValue(value), searchQuery)" />
          </template>
          <template v-else-if="isCollapsed(String(key))">
            <span class="text-muted">{{ getPreview(value) }}</span>
          </template>
        </div>
        
        <template v-if="!isPrimitive(value) && !isCollapsed(String(key))">
          <JsonViewer 
            :data="value" 
            :search-query="searchQuery" 
            :level="level + 1"
            :path="path ? `${path}.${key}` : String(key)"
            @path-click="emit('pathClick', $event)"
          />
        </template>
      </div>
    </template>
  </div>
</template>

<style scoped>
.json-key {
  color: #0ea5e9;
}

.json-string {
  color: #22c55e;
}

.json-number {
  color: #f97316;
}

.json-boolean {
  color: #a855f7;
}

.json-null {
  color: #6b7280;
}

.json-bracket {
  color: #6b7280;
}

.search-highlight {
  background-color: #fcd34d;
  color: #111827;
  padding: 0 2px;
  border-radius: 2px;
}

:global(.dark) .json-key {
  color: #38bdf8;
}

:global(.dark) .json-string {
  color: #4ade80;
}

:global(.dark) .json-number {
  color: #fb923c;
}

:global(.dark) .json-boolean {
  color: #c084fc;
}

:global(.dark) .json-null {
  color: #9ca3af;
}

:global(.dark) .search-highlight {
  background-color: #78350f;
  color: #f9fafb;
}
</style>
