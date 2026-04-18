<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import AiInspectorMessageBubble from './AiInspectorMessageBubble.vue'
import AiInspectorIcon from './AiInspectorIcon.vue'

interface Message {
  role: 'system' | 'user' | 'assistant' | 'tool'
  content?: string | any[] | null
  name?: string
  tool_calls?: any[]
  tool_call_id?: string
}

interface Props {
  messages: Message[]
  searchQuery?: string
}

defineProps<Props>()

const messagesContainer = ref<HTMLElement | null>(null)

const scrollToTop = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTo({ top: 0, behavior: 'smooth' })
  }
}

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTo({ top: messagesContainer.value.scrollHeight, behavior: 'smooth' })
  }
}

onMounted(() => {
  nextTick(() => {
    scrollToBottom()
  })
})

// 暴露方法给父组件
defineExpose({
  scrollToTop,
  scrollToBottom
})
</script>

<template>
  <div class="messages-section">
    <div class="messages-header">
      <h3 class="messages-title">
        <span class="section-icon">💬</span>
        对话消息 ({{ messages.length }})
      </h3>
      <div class="messages-scroll-controls">
        <button class="scroll-control-btn" @click="scrollToTop" title="滚动到顶部">
          <AiInspectorIcon name="i-lucide-arrow-up" size="sm" />
        </button>
        <button class="scroll-control-btn" @click="scrollToBottom" title="滚动到底部">
          <AiInspectorIcon name="i-lucide-arrow-down" size="sm" />
        </button>
      </div>
    </div>
    <div ref="messagesContainer" class="messages-container">
      <AiInspectorMessageBubble
        v-for="(message, index) in messages"
        :key="index"
        :message="message"
        :index="index"
        :search-query="searchQuery"
      />
    </div>
  </div>
</template>

<style scoped>
.messages-section {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.messages-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--color-border-default);
  background: var(--color-bg-muted);
}

.messages-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-default);
}

.section-icon {
  font-size: 1.25rem;
}

.messages-scroll-controls {
  display: flex;
  gap: 0.5rem;
}

.scroll-control-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  border-radius: 0.375rem;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all 0.2s;
}

.scroll-control-btn:hover {
  background: var(--color-bg-muted);
  color: #3b82f6;
  border-color: #3b82f6;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* 自定义滚动条 */
.messages-container::-webkit-scrollbar {
  width: 8px;
}

.messages-container::-webkit-scrollbar-track {
  background: transparent;
}

.messages-container::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
}

.messages-container::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.3);
}
</style>
