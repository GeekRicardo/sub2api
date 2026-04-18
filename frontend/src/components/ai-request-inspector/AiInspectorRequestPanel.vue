<script setup lang="ts">
import { computed, ref } from 'vue'
import AiInspectorIcon from './AiInspectorIcon.vue'
import AiInspectorJsonViewer from './AiInspectorJsonViewer.vue'
import AiInspectorMessageListLangSmith from './AiInspectorMessageListLangSmith.vue'

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

const hasMessages = computed(() => Array.isArray(props.data?.messages) && props.data.messages.length > 0)
const hasChoicesMessages = computed(() => Array.isArray(props.data?.choices) && props.data.choices.some((choice: any) => choice.message))

const choicesMessages = computed(() => {
  if (!hasChoicesMessages.value) return []
  return props.data.choices.filter((choice: any) => choice.message).map((choice: any) => choice.message)
})

const otherParams = computed(() => {
  if (!props.data) return []
  const result: { key: string; value: any; type: string }[] = []
  const excludeKeys = hasMessages.value ? ['messages'] : hasChoicesMessages.value ? ['choices'] : []
  Object.entries(props.data).forEach(([key, value]) => {
    if (excludeKeys.includes(key)) return
    const type = Array.isArray(value) ? 'array' : typeof value
    result.push({ key, value, type })
  })
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
    setTimeout(() => (copied.value = false), 2000)
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
  return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

const formatValue = (value: any): string => {
  if (value === null) return 'null'
  if (value === undefined) return 'undefined'
  if (typeof value === 'object') return JSON.stringify(value, null, 2)
  return String(value)
}

const isSimpleValue = (value: any): boolean => !(value !== null && value !== undefined && typeof value === 'object')

const getTypeColor = (type: string): string => {
  const colors: Record<string, string> = {
    string: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
    number: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
    boolean: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
    array: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
    object: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300'
  }
  return colors[type] || 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
}
</script>

<template>
  <div class="flex h-full flex-col bg-default">
    <div class="flex items-center justify-between border-b border-default bg-muted/30 px-4 py-3">
      <div class="flex items-center gap-2">
        <AiInspectorIcon :name="icon" class="size-5 text-primary" />
        <h3 class="font-medium text-default">{{ title }}</h3>
      </div>

      <div class="flex items-center gap-2">
        <div class="inline-flex overflow-hidden rounded-lg border border-default">
          <button
            class="px-3 py-1.5 text-xs transition"
            :class="viewMode === 'formatted' ? 'bg-primary text-white' : 'bg-transparent text-muted hover:text-default'"
            @click="viewMode = 'formatted'"
          >
            格式化
          </button>
          <button
            class="px-3 py-1.5 text-xs transition"
            :class="viewMode === 'raw' ? 'bg-primary text-white' : 'bg-transparent text-muted hover:text-default'"
            @click="viewMode = 'raw'"
          >
            原始 JSON
          </button>
        </div>

        <button
          class="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs text-muted transition hover:bg-muted/40 hover:text-default"
          @click="copyToClipboard"
        >
          <AiInspectorIcon :name="copied ? 'i-lucide-check' : 'i-lucide-copy'" size="xs" />
          {{ copied ? '已复制' : '复制' }}
        </button>
      </div>
    </div>

    <div class="custom-scrollbar flex-1 overflow-auto p-6">
      <template v-if="viewMode === 'formatted'">
        <template v-if="data">
          <div v-if="hasMessages || hasChoicesMessages">
            <div class="mb-6 flex items-center gap-3">
              <div class="flex items-center gap-2">
                <AiInspectorIcon name="i-lucide-messages-square" class="size-6 text-primary" />
                <div>
                  <h4 class="text-base font-semibold text-default">Messages</h4>
                  <p class="text-xs text-muted">{{ hasMessages ? data.messages.length : choicesMessages.length }} 条消息</p>
                </div>
              </div>
            </div>

            <AiInspectorMessageListLangSmith :messages="hasMessages ? data.messages : choicesMessages" :search-query="searchQuery" />
          </div>

          <div v-if="otherParams.length > 0" class="mt-10">
            <div class="mb-6">
              <div class="flex items-center gap-2">
                <AiInspectorIcon name="i-lucide-settings-2" class="size-6 text-muted" />
                <h4 class="text-base font-semibold text-default">请求参数</h4>
              </div>
            </div>

            <div class="overflow-hidden rounded-lg border border-default bg-muted/20">
              <table class="w-full">
                <thead>
                  <tr class="border-b border-default bg-muted/40">
                    <th class="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">参数</th>
                    <th class="w-20 px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">类型</th>
                    <th class="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">值</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(param, index) in otherParams"
                    :key="param.key"
                    class="border-b border-default transition-colors hover:bg-muted/40"
                    :class="index === otherParams.length - 1 ? 'border-b-0' : ''"
                  >
                    <td class="px-5 py-4">
                      <code class="rounded bg-elevated/50 px-2.5 py-1 font-mono text-sm text-default">{{ param.key }}</code>
                    </td>
                    <td class="px-5 py-4">
                      <span class="rounded-full px-2 py-1 text-xs font-medium" :class="getTypeColor(param.type)">{{ param.type }}</span>
                    </td>
                    <td class="px-5 py-4">
                      <template v-if="isSimpleValue(param.value)">
                        <code class="text-sm font-mono" :class="[
                          typeof param.value === 'boolean' ? 'text-orange-600 dark:text-orange-400' :
                          typeof param.value === 'number' ? 'text-green-600 dark:text-green-400' :
                          typeof param.value === 'string' ? 'text-blue-600 dark:text-blue-400' :
                          param.value === null ? 'italic text-muted' : 'text-default'
                        ]">
                          {{ formatValue(param.value) }}
                        </code>
                      </template>
                      <template v-else>
                        <details class="group">
                          <summary class="inline-flex cursor-pointer items-center gap-1.5 select-none text-sm text-muted transition-colors hover:text-default">
                            <AiInspectorIcon name="i-lucide-chevron-right" class="size-4 transition-transform group-open:rotate-90" />
                            <span>{{ Array.isArray(param.value) ? `数组 (${param.value.length})` : `对象 (${Object.keys(param.value).length})` }}</span>
                          </summary>
                          <div class="custom-scrollbar mt-3 max-h-72 overflow-auto rounded-lg border border-default bg-elevated p-4">
                            <AiInspectorJsonViewer :data="param.value" :search-query="searchQuery" />
                          </div>
                        </details>
                      </template>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div v-if="!hasMessages && !hasChoicesMessages && otherParams.length === 0" class="flex flex-col items-center justify-center py-12 text-muted">
            <AiInspectorIcon name="i-lucide-file-x" class="mb-3 size-12 opacity-50" />
            <span class="text-sm">暂无数据</span>
          </div>
        </template>

        <div v-else class="flex h-full flex-col items-center justify-center text-muted">
          <AiInspectorIcon name="i-lucide-file-x" class="mb-3 size-12 opacity-50" />
          <span class="text-sm">暂无数据</span>
        </div>
      </template>

      <template v-else>
        <div class="custom-scrollbar overflow-x-auto rounded-lg border border-default bg-elevated p-4">
          <pre
            v-if="rawJson"
            class="whitespace-pre-wrap break-all font-mono text-xs leading-relaxed text-muted"
            v-html="highlightText(rawJson, searchQuery)"
          />
          <span v-else class="text-sm text-muted">暂无数据</span>
        </div>
      </template>
    </div>
  </div>
</template>
