<script setup lang="ts">
import { computed } from 'vue'
import AiInspectorIcon from './AiInspectorIcon.vue'

interface Message {
  role: 'system' | 'user' | 'assistant' | 'tool'
  content?: string | null
  name?: string
  tool_calls?: ToolCall[]
  tool_call_id?: string
}

interface ToolCall {
  id?: string
  type?: string
  function?: {
    name: string
    arguments: string
  }
  name?: string
  arguments?: string
}

interface Props {
  messages: Message[]
  searchQuery?: string
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: ''
})

const searchQuery = computed(() => props.searchQuery || '')

const getRoleConfig = (role: string) => {
  const configs: Record<string, { label: string; chip: string }> = {
    system: { label: '系统', chip: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300' },
    user: { label: '用户', chip: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300' },
    assistant: { label: '助手', chip: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' },
    tool: { label: '工具', chip: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300' }
  }
  return configs[role] || configs.user
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

const formatJson = (obj: any): string => {
  try {
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(obj)
  }
}

const parseToolArguments = (args: any): any => {
  if (typeof args === 'string') {
    try {
      return JSON.parse(args)
    } catch {
      return args
    }
  }
  return args
}

const parseToolContent = (content: string | null | undefined): any => {
  if (!content) return null
  try {
    return JSON.parse(content)
  } catch {
    return content
  }
}

const getToolName = (toolCall: ToolCall): string => {
  return toolCall.function?.name || toolCall.name || 'unknown'
}

const getToolArgs = (toolCall: ToolCall): any => {
  const args = toolCall.function?.arguments || toolCall.arguments
  return parseToolArguments(args)
}
</script>

<template>
  <div class="space-y-4">
    <div
      v-for="(message, index) in messages"
      :key="index"
      class="overflow-hidden rounded-lg border border-default bg-white transition-all hover:border-primary/40 dark:bg-gray-950"
      :class="[
        message.role === 'system' && 'border-l-4 border-l-purple-500',
        message.role === 'user' && 'border-l-4 border-l-blue-500',
        message.role === 'assistant' && 'border-l-4 border-l-green-500',
        message.role === 'tool' && 'border-l-4 border-l-orange-500'
      ]"
    >
      <div
        class="flex items-start justify-between px-5 py-4"
        :class="[
          message.role === 'system' && 'bg-purple-50/50 dark:bg-purple-950/20',
          message.role === 'user' && 'bg-blue-50/50 dark:bg-blue-950/20',
          message.role === 'assistant' && 'bg-green-50/50 dark:bg-green-950/20',
          message.role === 'tool' && 'bg-orange-50/50 dark:bg-orange-950/20'
        ]"
      >
        <div class="flex items-start gap-3">
          <div class="mt-0.5 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-default text-xs font-bold text-muted">
            {{ index + 1 }}
          </div>

          <div>
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-full px-2.5 py-1 text-xs font-semibold" :class="getRoleConfig(message.role).chip">
                {{ getRoleConfig(message.role).label }}
              </span>

              <span v-if="message.role === 'tool' && message.name" class="rounded bg-default/50 px-2 py-1 font-mono text-xs text-muted">
                {{ message.name }}
              </span>

              <code v-if="message.tool_call_id" class="rounded bg-default/50 px-2 py-1 font-mono text-xs text-muted">
                {{ message.tool_call_id }}
              </code>

              <span
                v-if="message.tool_calls && message.tool_calls.length > 0"
                class="rounded-full bg-orange-100 px-2 py-1 text-xs font-medium text-orange-700 dark:bg-orange-900/30 dark:text-orange-300"
              >
                调用 {{ message.tool_calls.length }} 个工具
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-4 border-t border-default px-5 py-4">
        <div v-if="message.content" class="prose prose-sm max-w-none dark:prose-invert">
          <div class="whitespace-pre-wrap break-words text-sm leading-relaxed text-default" v-html="highlightText(message.content, searchQuery)" />
        </div>

        <div v-if="message.tool_calls && message.tool_calls.length > 0">
          <div class="mb-3 text-xs font-semibold uppercase tracking-wider text-orange-700 dark:text-orange-400">工具调用</div>
          <div class="space-y-3">
            <div
              v-for="(toolCall, tcIndex) in message.tool_calls"
              :key="tcIndex"
              class="overflow-hidden rounded-lg border border-orange-200 bg-orange-50 dark:border-orange-800 dark:bg-orange-950/30"
            >
              <div class="flex items-center gap-2 bg-orange-100/60 px-3 py-2.5 dark:bg-orange-900/40">
                <AiInspectorIcon name="i-lucide-function-square" class="size-4 text-orange-600 dark:text-orange-400" />
                <span class="font-mono text-sm font-semibold text-default">{{ getToolName(toolCall) }}</span>
                <code v-if="toolCall.id" class="rounded bg-white px-1.5 py-0.5 font-mono text-xs text-muted dark:bg-gray-950">
                  {{ toolCall.id }}
                </code>
              </div>

              <div class="bg-white px-3 py-3 dark:bg-gray-950/50">
                <div class="mb-2 text-xs font-semibold uppercase tracking-wider text-muted">参数</div>
                <div class="custom-scrollbar max-h-32 overflow-auto rounded bg-muted/20 p-3">
                  <pre class="whitespace-pre-wrap break-all font-mono text-xs leading-relaxed text-muted" v-html="highlightText(formatJson(getToolArgs(toolCall)), searchQuery)" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="message.role === 'tool' && message.content">
          <div class="mb-3 text-xs font-semibold uppercase tracking-wider text-orange-700 dark:text-orange-400">返回结果</div>
          <div class="custom-scrollbar max-h-48 overflow-auto rounded-lg border border-default/50 bg-muted/20 p-3">
            <pre class="whitespace-pre-wrap break-all font-mono text-xs leading-relaxed text-muted" v-html="highlightText(formatJson(parseToolContent(message.content)), searchQuery)" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
