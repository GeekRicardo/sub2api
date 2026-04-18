<template>
  <div class="relative font-mono text-xs leading-6">
    <button
      v-if="level === 0 && showCopyButton"
      type="button"
      class="absolute right-0 top-0 rounded-md p-1 text-slate-400 transition hover:bg-white/10 hover:text-white"
      title="Copy JSON"
      @click.stop="copyJson"
    >
      <Icon :name="copied ? 'check' : 'copy'" size="xs" />
    </button>

    <template v-if="normalizedType === 'null'">
      <span class="text-slate-500">null</span>
    </template>

    <template v-else-if="normalizedType === 'boolean'">
      <span class="font-medium text-indigo-400">{{ String(data) }}</span>
    </template>

    <template v-else-if="normalizedType === 'number'">
      <span class="text-emerald-400">{{ data as number }}</span>
    </template>

    <template v-else-if="normalizedType === 'string'">
      <span class="text-violet-300">
        <template v-if="expandedString || !isStringTruncated">
          {{ renderString(data as string) }}
        </template>
        <template v-else>
          {{ renderString((data as string).slice(0, stringMaxLength)) }}<span class="text-slate-500">...</span>
        </template>
        <button
          v-if="isStringTruncated"
          type="button"
          class="ml-2 rounded-md bg-white/5 px-2 py-1 text-[11px] text-slate-300 transition hover:bg-white/10"
          @click.stop="expandedString = !expandedString"
        >
          {{ expandedString ? 'Show less' : `Show ${(data as string).length - stringMaxLength} more` }}
        </button>
      </span>
    </template>

    <template v-else-if="normalizedType === 'array'">
      <div>
        <button type="button" class="inline-flex items-center gap-1 text-slate-300" @click="expanded = !expanded">
          <Icon :name="expanded ? 'chevronDown' : 'chevronRight'" size="xs" />
          <span>[</span>
          <span v-if="!expanded" class="text-[11px] italic text-slate-500">...{{ (data as unknown[]).length }} items</span>
          <span v-if="!expanded">]</span>
        </button>
        <div v-if="expanded" class="ml-4 border-l border-slate-700 pl-3">
          <div v-for="(item, index) in data as unknown[]" :key="index">
            <UsageJsonRenderer :data="item" :level="level + 1" :show-copy-button="false" />
            <span v-if="index < (data as unknown[]).length - 1" class="text-slate-500">,</span>
          </div>
        </div>
        <div v-if="expanded" class="ml-3 text-slate-300">]</div>
      </div>
    </template>

    <template v-else>
      <div>
        <button type="button" class="inline-flex items-center gap-1 text-slate-300" @click="expanded = !expanded">
          <Icon :name="expanded ? 'chevronDown' : 'chevronRight'" size="xs" />
          <span>{</span>
          <span v-if="!expanded" class="text-[11px] italic text-slate-500">...{{ objectEntries.length }} properties</span>
          <span v-if="!expanded">}</span>
        </button>
        <div v-if="expanded" class="ml-4 border-l border-slate-700 pl-3">
          <div v-for="([key, value], index) in objectEntries" :key="key">
            <span class="font-medium text-sky-300">"{{ key }}"</span><span class="text-slate-500">: </span>
            <UsageJsonRenderer :data="value" :level="level + 1" :show-copy-button="false" />
            <span v-if="index < objectEntries.length - 1" class="text-slate-500">,</span>
          </div>
        </div>
        <div v-if="expanded" class="ml-3 text-slate-300">}</div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'

defineOptions({ name: 'UsageJsonRenderer' })

const props = withDefaults(defineProps<{
  data: unknown
  level?: number
  showCopyButton?: boolean
  isExpanded?: boolean
}>(), {
  level: 0,
  showCopyButton: true,
  isExpanded: true
})

const expanded = ref(props.isExpanded)
const expandedString = ref(false)
const copied = ref(false)
const stringMaxLength = 10000

const normalizedType = computed(() => {
  if (props.data === null || props.data === undefined) return 'null'
  if (Array.isArray(props.data)) return 'array'
  return typeof props.data
})

const objectEntries = computed(() => {
  if (normalizedType.value !== 'object') return []
  return Object.entries(props.data as Record<string, unknown>)
})

const isStringTruncated = computed(() => {
  return normalizedType.value === 'string' && String(props.data).length > stringMaxLength
})

function renderString(value: string) {
  return props.level === 0 ? value : `"${value.replace(/\n/g, '\\n')}"`
}

async function copyJson() {
  try {
    await navigator.clipboard.writeText(JSON.stringify(props.data, null, 2))
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch {
    copied.value = false
  }
}
</script>
