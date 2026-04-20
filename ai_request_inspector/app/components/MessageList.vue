<script setup lang="ts">
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

const getRoleConfig = (role: string) => {
  const configs: Record<string, { label: string; color: string }> = {
    system: { label: '系统', color: 'purple' },
    user: { label: '用户', color: 'blue' },
    assistant: { label: '助手', color: 'green' },
    tool: { label: '工具', color: 'orange' }
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

// 修复模板中的 Uncaught SyntaxError
const getMessageLabel = (role: string): string => {
  const labels: Record<string, string> = {
    'system': '系统',
    'user': '用户',
    'assistant': '助手',
    'tool': '工具'
  }
  return labels[role] || '消息'
}
</script>

<template>
  <div class="space-y-4">
    <!-- Message Card -->
    <div
      v-for="(message, index) in messages"
      :key="index"
      class="border border-default rounded-lg overflow-hidden bg-white dark:bg-gray-950 transition-all hover:border-primary/40"
      :class="[
        message.role === 'system' && 'border-l-4 border-l-purple-500',
        message.role === 'user' && 'border-l-4 border-l-blue-500',
        message.role === 'assistant' && 'border-l-4 border-l-green-500',
        message.role === 'tool' && 'border-l-4 border-l-orange-500'
      ]"
    >
      <!-- Message Header -->
      <div 
        class="px-5 py-4 flex items-start justify-between"
        :class="[
          message.role === 'system' && 'bg-purple-50/50 dark:bg-purple-950/20',
          message.role === 'user' && 'bg-blue-50/50 dark:bg-blue-950/20',
          message.role === 'assistant' && 'bg-green-50/50 dark:bg-green-950/20',
          message.role === 'tool' && 'bg-orange-50/50 dark:bg-orange-950/20'
        ]"
      >
        <div class="flex items-start gap-3">
          <!-- Index Badge -->
          <div class="w-7 h-7 rounded-full bg-default flex items-center justify-center text-xs font-bold text-muted flex-shrink-0 mt-0.5">
            {{ index + 1 }}
          </div>
          
          <div>
            <!-- Role and Metadata -->
            <div class="flex items-center gap-2 flex-wrap">
              <UBadge 
                :label="getRoleConfig(message.role).label" 
                :color="getRoleConfig(message.role).color"
                variant="solid"
                size="sm"
              />
              
              <!-- Tool metadata -->
              <span v-if="message.role === 'tool' && message.name" class="text-xs font-mono text-muted px-2 py-1 bg-default/50 rounded">
                {{ message.name }}
              </span>
              
              <code v-if="message.tool_call_id" class="text-xs font-mono text-muted bg-default/50 px-2 py-1 rounded">
                {{ message.tool_call_id }}
              </code>
              
              <UBadge 
                v-if="message.tool_calls && message.tool_calls.length > 0" 
                :label="`调用 ${message.tool_calls.length} 个工具`"
                color="orange"
                variant="subtle"
                size="xs"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Message Content -->
      <div class="px-5 py-4 space-y-4 border-t border-default">
        <!-- Text Content -->
        <div v-if="message.content" class="prose prose-sm dark:prose-invert max-w-none">
          <div 
            class="text-sm text-default whitespace-pre-wrap break-words leading-relaxed"
            v-html="highlightText(message.content, searchQuery)"
          />
        </div>

        <!-- Tool Calls -->
        <div v-if="message.tool_calls && message.tool_calls.length > 0">
          <div class="text-xs font-semibold text-orange-700 dark:text-orange-400 mb-3 uppercase tracking-wider">
            工具调用
          </div>
          
          <div class="space-y-3">
            <div
              v-for="(toolCall, tcIndex) in message.tool_calls"
              :key="tcIndex"
              class="border border-orange-200 dark:border-orange-800 rounded-lg bg-orange-50 dark:bg-orange-950/30 overflow-hidden"
            >
              <!-- Tool Call Header -->
              <div class="px-3 py-2.5 bg-orange-100/60 dark:bg-orange-900/40 flex items-center gap-2">
                <UIcon name="i-lucide-function-square" class="size-4 text-orange-600 dark:text-orange-400" />
                <span class="font-mono text-sm font-semibold text-default">{{ getToolName(toolCall) }}</span>
                <code v-if="toolCall.id" class="text-xs font-mono text-muted bg-white dark:bg-gray-950 px-1.5 py-0.5 rounded">
                  {{ toolCall.id }}
                </code>
              </div>
              
              <!-- Arguments -->
              <div class="px-3 py-3 bg-white dark:bg-gray-950/50">
                <div class="text-xs text-muted font-semibold mb-2 uppercase tracking-wider">参数</div>
                <div class="bg-muted/20 rounded p-3 overflow-auto max-h-32 custom-scrollbar">
                  <pre 
                    class="text-xs font-mono text-muted whitespace-pre-wrap break-all leading-relaxed"
                    v-html="highlightText(formatJson(getToolArgs(toolCall)), searchQuery)"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Tool Result -->
        <div v-if="message.role === 'tool' && message.content">
          <div class="text-xs font-semibold text-orange-700 dark:text-orange-400 mb-3 uppercase tracking-wider">
            返回结果
          </div>
          <div class="bg-muted/20 rounded-lg p-3 overflow-auto max-h-48 custom-scrollbar border border-default/50">
            <pre 
              class="text-xs font-mono text-muted whitespace-pre-wrap break-all leading-relaxed"
              v-html="highlightText(formatJson(parseToolContent(message.content)), searchQuery)"
            />
          </div>
        </div>
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
