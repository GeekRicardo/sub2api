<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  data: unknown
  depth?: number
  maxDepth?: number
}

const props = withDefaults(defineProps<Props>(), {
  depth: 0,
  maxDepth: 3
})

const collapsed = ref(props.depth >= props.maxDepth)

const kind = computed<'array' | 'object' | 'primitive'>(() => {
  const v = props.data
  if (Array.isArray(v)) return 'array'
  if (v !== null && typeof v === 'object') return 'object'
  return 'primitive'
})

const keys = computed(() => (kind.value === 'object' ? Object.keys(props.data as object) : []))
const arr = computed(() => (kind.value === 'array' ? (props.data as unknown[]) : []))

const primClass = computed(() => {
  const v = props.data
  if (v === null || v === undefined) return 'rd-json-null'
  const t = typeof v
  if (t === 'boolean') return 'rd-json-bool'
  if (t === 'number') return 'rd-json-num'
  if (t === 'string') return 'rd-json-str'
  return 'rd-json-unknown'
})

const primText = computed(() => {
  const v = props.data
  if (v === null) return 'null'
  if (v === undefined) return 'undefined'
  if (typeof v === 'string') return `"${v}"`
  return String(v)
})
</script>

<template>
  <span v-if="kind === 'primitive'" :class="primClass">{{ primText }}</span>

  <span v-else-if="kind === 'array' && arr.length === 0" class="rd-json-bracket">[]</span>

  <span v-else-if="kind === 'array'">
    <span class="rd-json-bracket">[</span>
    <button class="rd-json-toggle" @click="collapsed = !collapsed">
      <span class="rd-json-caret">{{ collapsed ? '▶' : '▼' }}</span>
      <span class="rd-json-count">{{ arr.length }} items</span>
    </button>
    <div v-if="!collapsed" class="rd-json-children">
      <div v-for="(item, i) in arr" :key="i" class="rd-json-row">
        <span class="rd-json-key">{{ i }}:</span>
        <RdJsonNode :data="item" :depth="depth + 1" :max-depth="maxDepth" />
        <span v-if="i < arr.length - 1" class="rd-json-bracket">,</span>
      </div>
    </div>
    <span class="rd-json-bracket">]</span>
  </span>

  <span v-else-if="kind === 'object' && keys.length === 0" class="rd-json-bracket">{}</span>

  <span v-else>
    <span class="rd-json-bracket">{</span>
    <button class="rd-json-toggle" @click="collapsed = !collapsed">
      <span class="rd-json-caret">{{ collapsed ? '▶' : '▼' }}</span>
      <span class="rd-json-count">{{ keys.length }} keys</span>
    </button>
    <div v-if="!collapsed" class="rd-json-children">
      <div v-for="(k, i) in keys" :key="k" class="rd-json-row">
        <span class="rd-json-key">"{{ k }}":</span>
        <RdJsonNode :data="(data as any)[k]" :depth="depth + 1" :max-depth="maxDepth" />
        <span v-if="i < keys.length - 1" class="rd-json-bracket">,</span>
      </div>
    </div>
    <span class="rd-json-bracket">}</span>
  </span>
</template>

<script lang="ts">
export default { name: 'RdJsonNode' }
</script>
