<script setup lang="ts">
import { computed, ref } from 'vue'
import AiInspectorIcon from './AiInspectorIcon.vue'
import AiInspectorJsonViewer from './AiInspectorJsonViewer.vue'

interface Message {
  role: 'system' | 'user' | 'assistant' | 'tool'
  content?: string | any[] | null
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
  input?: any
}

interface Props {
  message: Message
  index: number
  searchQuery?: string
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: ''
})

const viewMode = ref<'markdown' | 'json'>('markdown')

// 提取文本内容
const textContent = computed(() => {
  const content = props.message.content
  if (typeof content === 'string') {
    return content
  }
  if (Array.isArray(content)) {
    const textItems = content.filter((item: any) => item.type === 'text')
    return textItems.map((item: any) => item.text).join('\n\n')
  }
  return ''
})

const hasTextContent = computed(() => {
  return textContent.value && textContent.value.trim().length > 0
})

// 提取其他内容（图片、工具调用等）
const otherContent = computed(() => {
  const content = props.message.content
  if (!Array.isArray(content)) return []
  return content.filter((item: any) => item.type !== 'text')
})

const roleConfig = computed(() => {
  const configs = {
    system: { emoji: '⚙️', label: 'system', bgClass: 'bg-purple-50 dark:bg-purple-950/20' },
    user: { emoji: '👤', label: 'user', bgClass: 'bg-blue-50 dark:bg-blue-950/20' },
    assistant: { emoji: '🤖', label: 'assistant', bgClass: 'bg-slate-50 dark:bg-slate-950/20' },
    tool: { emoji: '🔧', label: 'tool', bgClass: 'bg-orange-50 dark:bg-orange-950/20' }
  }
  return configs[props.message.role] || configs.user
})

const getToolName = (toolCall: ToolCall): string => {
  return toolCall.function?.name || toolCall.name || 'unknown'
}

const getToolArgs = (toolCall: ToolCall): any => {
  if (toolCall.input) return toolCall.input
  const args = toolCall.function?.arguments || toolCall.arguments
  if (typeof args === 'string') {
    try {
      return JSON.parse(args)
    } catch {
      return args
    }
  }
  return args
}

// Markdown 渲染
const renderMarkdown = (text: string): string => {
  let html = escapeHtml(text)

  // 代码块
  html = html.replace(/```(\w+)?\n([\s\S]*?)```/g, (_match, lang, code) => {
    return `<pre class="markdown-code-block"><code class="language-${lang || 'text'}">${code.trim()}</code></pre>`
  })

  // 行内代码
  html = html.replace(/`([^`]+)`/g, '<code class="markdown-inline-code">$1</code>')

  // 标题
  html = html.replace(/^### (.+)$/gm, '<h3 class="markdown-h3">$1</h3>')
  html = html.replace(/^## (.+)$/gm, '<h2 class="markdown-h2">$1</h2>')
  html = html.replace(/^# (.+)$/gm, '<h1 class="markdown-h1">$1</h1>')

  // 粗体
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/__(.+?)__/g, '<strong>$1</strong>')

  // 斜体
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')
  html = html.replace(/_(.+?)_/g, '<em>$1</em>')

  // 链接
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>')

  // 无序列表
  html = html.replace(/^\* (.+)$/gm, '<li>$1</li>')
  html = html.replace(/^- (.+)$/gm, '<li>$1</li>')
  html = html.replace(/(<li>.*<\/li>\n?)+/g, '<ul class="markdown-list">$&</ul>')

  // 换行
  html = html.replace(/\n\n/g, '</p><p class="markdown-p">')
  html = html.replace(/\n/g, '<br>')

  return `<div class="markdown-content"><p class="markdown-p">${html}</p></div>`
}

const escapeHtml = (text: string): string => {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}
</script>

<template>
  <div class="message-bubble" :class="[message.role, roleConfig.bgClass]">
    <div class="message-avatar">{{ roleConfig.emoji }}</div>
    <div class="message-content-wrapper">
      <div class="message-header">
        <div class="message-role-label">{{ roleConfig.label }}</div>
        <div v-if="hasTextContent" class="message-view-toggle">
          <button
            class="view-toggle-btn"
            :class="{ active: viewMode === 'markdown' }"
            @click="viewMode = 'markdown'"
          >
            <AiInspectorIcon name="i-lucide-file-text" size="xs" />
            Markdown
          </button>
          <button
            class="view-toggle-btn"
            :class="{ active: viewMode === 'json' }"
            @click="viewMode = 'json'"
          >
            <AiInspectorIcon name="i-lucide-braces" size="xs" />
            JSON
          </button>
        </div>
      </div>

      <!-- Markdown 视图 -->
      <div v-if="hasTextContent && viewMode === 'markdown'" class="message-content-markdown" v-html="renderMarkdown(textContent)" />

      <!-- JSON 视图 -->
      <div v-if="hasTextContent && viewMode === 'json'" class="message-content-json">
        <AiInspectorJsonViewer :data="message" :search-query="searchQuery" />
      </div>

      <!-- 如果没有文本内容，直接显示 JSON -->
      <div v-if="!hasTextContent" class="message-content-json">
        <AiInspectorJsonViewer :data="message" :search-query="searchQuery" />
      </div>

      <!-- 其他内容（图片、工具调用等） -->
      <div v-if="otherContent.length > 0" class="message-other-content">
        <div v-for="(item, idx) in otherContent" :key="idx">
          <div v-if="item.type === 'image_url'" class="message-content-box">
            🖼️ [Image]
          </div>
          <div v-else-if="item.type === 'tool_use'" class="tool-call-card">
            <div class="tool-call-header">
              <div class="tool-call-title">
                <span class="tool-call-icon">🔧</span>
                {{ item.name }}
                <span v-if="item.id" class="tool-call-id">{{ item.id }}</span>
              </div>
              <span class="tool-call-expand">▼</span>
            </div>
            <div class="tool-call-body">
              <AiInspectorJsonViewer :data="item.input" :search-query="searchQuery" />
            </div>
          </div>
          <div v-else-if="item.type === 'tool_result'" class="tool-call-card tool-result-card">
            <div class="tool-call-header">
              <div class="tool-call-title">
                <span class="tool-call-icon">✅</span>
                Tool Result: {{ item.tool_use_id }}
              </div>
              <span class="tool-call-expand">▼</span>
            </div>
            <div class="tool-call-body">
              <AiInspectorJsonViewer :data="item.content" :search-query="searchQuery" />
            </div>
          </div>
        </div>
      </div>

      <!-- OpenAI tool calls -->
      <div v-if="message.tool_calls && message.tool_calls.length > 0" class="message-tool-calls">
        <div v-for="(tc, tcIdx) in message.tool_calls" :key="tcIdx" class="tool-call-card">
          <div class="tool-call-header">
            <div class="tool-call-title">
              <span class="tool-call-icon">🔧</span>
              {{ getToolName(tc) }}
              <span v-if="tc.id" class="tool-call-id">{{ tc.id }}</span>
            </div>
            <span class="tool-call-expand">▼</span>
          </div>
          <div class="tool-call-body">
            <AiInspectorJsonViewer :data="getToolArgs(tc)" :search-query="searchQuery" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-bubble {
  display: flex;
  gap: 1rem;
  padding: 1.5rem;
  border-radius: 0.75rem;
  animation: messageSlideIn 0.3s ease-out;
}

@keyframes messageSlideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-bubble.user {
  flex-direction: row-reverse;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), rgba(37, 99, 235, 0.05));
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.5);
}

.message-bubble.user .message-avatar {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
}

.message-content-wrapper {
  flex: 1;
  min-width: 0;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.message-role-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
}

.message-view-toggle {
  display: flex;
  gap: 0.25rem;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 0.5rem;
  padding: 0.125rem;
}

.view-toggle-btn {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s;
}

.view-toggle-btn.active {
  background: #3b82f6;
  color: white;
}

.message-content-markdown {
  line-height: 1.6;
  color: var(--color-text-default);
}

.markdown-content {
  font-size: 0.875rem;
}

.markdown-p {
  margin: 0.5rem 0;
}

.markdown-code-block {
  background: rgba(0, 0, 0, 0.05);
  border-radius: 0.5rem;
  padding: 1rem;
  margin: 0.75rem 0;
  overflow-x: auto;
}

.markdown-inline-code {
  background: rgba(0, 0, 0, 0.05);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.875em;
}

.markdown-h1, .markdown-h2, .markdown-h3 {
  margin: 1rem 0 0.5rem;
  font-weight: 600;
}

.markdown-list {
  margin: 0.5rem 0;
  padding-left: 1.5rem;
}

.tool-call-card {
  margin-top: 0.75rem;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 0.5rem;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.5);
}

.tool-call-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: rgba(0, 0, 0, 0.02);
  cursor: pointer;
  user-select: none;
}

.tool-call-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.tool-call-icon {
  font-size: 1rem;
}

.tool-call-id {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  font-family: 'JetBrains Mono', monospace;
}

.tool-call-expand {
  color: var(--color-text-muted);
  transition: transform 0.2s;
}

.tool-call-card:not(.expanded) .tool-call-body {
  display: none;
}

.tool-call-card.expanded .tool-call-expand {
  transform: rotate(180deg);
}

.tool-call-body {
  padding: 1rem;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

.tool-result-card {
  border-color: rgba(16, 185, 129, 0.2);
}

.tool-result-card .tool-call-header {
  background: rgba(16, 185, 129, 0.05);
}
</style>
