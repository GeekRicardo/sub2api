<script setup lang="ts">
// Sample data for demonstration
const sampleRequest = {
  model: 'gpt-4-turbo',
  messages: [
    {
      role: 'system',
      content: '你是一个专业的AI助手，擅长回答技术问题。你可以使用 search_documents 工具来搜索相关文档以提供更准确的答案。'
    },
    {
      role: 'user',
      content: '请解释什么是 RAG（Retrieval-Augmented Generation）技术？'
    },
    {
      role: 'assistant',
      content: null,
      tool_calls: [
        {
          id: 'call_xyz123',
          type: 'function',
          function: {
            name: 'search_documents',
            arguments: '{"query":"RAG技术原理","limit":3}'
          }
        }
      ]
    },
    {
      role: 'tool',
      tool_call_id: 'call_xyz123',
      name: 'search_documents',
      content: '[{"title":"RAG技术概述","content":"RAG结合了检索和生成两个步骤..."},{"title":"RAG应用场景","content":"RAG广泛应用于问答系统..."}]'
    },
    {
      role: 'assistant',
      content: 'RAG（Retrieval-Augmented Generation，检索增强生成）是一种结合了信息检索和文本生成的AI技术架构。\n\n根据检索到的文档，RAG 的核心特点包括：\n\n1. **检索阶段**：从知识库中检索相关信息\n2. **生成阶段**：基于检索结果生成答案\n3. **减少幻觉**：通过真实数据源验证内容\n4. **知识更新**：无需重训练即可更新知识库'
    },
    {
      role: 'user',
      content: '能否给出一个具体的实现示例？'
    }
  ],
  temperature: 0.7,
  max_tokens: 2048,
  top_p: 1,
  frequency_penalty: 0,
  presence_penalty: 0,
  stream: true,
  tools: [
    {
      type: 'function',
      function: {
        name: 'search_documents',
        description: '搜索相关文档',
        parameters: {
          type: 'object',
          properties: {
            query: {
              type: 'string',
              description: '搜索关键词'
            },
            limit: {
              type: 'number',
              description: '返回结果数量'
            }
          },
          required: ['query']
        }
      }
    }
  ]
}

const sampleResponse = {
  id: 'chatcmpl-abc123xyz',
  object: 'chat.completion',
  created: 1712345678,
  model: 'gpt-4-turbo-2024-04-09',
  choices: [
    {
      index: 0,
      message: {
        role: 'assistant',
        content: '当然！下面是一个使用 Python 实现 RAG 的简单示例：\n\n```python\nfrom langchain.embeddings import OpenAIEmbeddings\nfrom langchain.vectorstores import Chroma\nfrom langchain.chains import RetrievalQA\nfrom langchain.llms import OpenAI\n\n# 1. 创建向量数据库\nembeddings = OpenAIEmbeddings()\nvectorstore = Chroma.from_documents(\n    documents=documents,\n    embedding=embeddings\n)\n\n# 2. 创建检索器\nretriever = vectorstore.as_retriever(\n    search_kwargs={"k": 3}\n)\n\n# 3. 创建 RAG 链\nqa_chain = RetrievalQA.from_chain_type(\n    llm=OpenAI(),\n    retriever=retriever,\n    return_source_documents=True\n)\n\n# 4. 查询\nresult = qa_chain({"query": "什么是RAG?"})\nprint(result["result"])\n```\n\n这个示例展示了 RAG 的三个核心步骤：文档向量化、相似度检索和生成答案。',
        tool_calls: [
          {
            id: 'call_code_example',
            type: 'function',
            function: {
              name: 'generate_code_snippet',
              arguments: '{"language":"python","topic":"RAG implementation"}'
            }
          }
        ]
      },
      finish_reason: 'tool_calls'
    }
  ],
  usage: {
    prompt_tokens: 892,
    completion_tokens: 445,
    total_tokens: 1337
  }
}

const sampleChunks = [
  {
    id: 1,
    timestamp: '00:00.000',
    raw: 'data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","created":1712345678,"model":"gpt-4-turbo","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}',
    parsed: { type: 'start', role: 'assistant' },
    type: 'text-delta'
  },
  {
    id: 2,
    timestamp: '00:00.120',
    raw: 'data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"RAG"},"finish_reason":null}]}',
    parsed: { textDelta: 'RAG' },
    type: 'text-delta'
  },
  {
    id: 3,
    timestamp: '00:00.240',
    raw: 'data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"（Retrieval-Augmented Generation，"},"finish_reason":null}]}',
    parsed: { textDelta: '（Retrieval-Augmented Generation，' },
    type: 'text-delta'
  },
  {
    id: 4,
    timestamp: '00:00.360',
    raw: 'data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"检索增强生成）"},"finish_reason":null}]}',
    parsed: { textDelta: '检索增强生成）' },
    type: 'text-delta'
  },
  {
    id: 5,
    timestamp: '00:00.480',
    raw: 'data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"是一种结合了信息检索和文本生成的AI技术架构。"},"finish_reason":null}]}',
    parsed: { textDelta: '是一种结合了信息检索和文本生成的AI技术架构。' },
    type: 'text-delta'
  },
  {
    id: 6,
    timestamp: '00:01.200',
    raw: 'data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_xyz","type":"function","function":{"name":"search_documents","arguments":"{\\\"query\\\":\\\"RAG技术原理\\\"}"}}]},"finish_reason":null}]}',
    parsed: { toolName: 'search_documents', args: { query: 'RAG技术原理' } },
    type: 'tool-call'
  },
  {
    id: 7,
    timestamp: '00:02.500',
    raw: 'data: {"id":"chatcmpl-abc123","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}',
    parsed: { finishReason: 'stop' },
    type: 'finish'
  },
  {
    id: 8,
    timestamp: '00:02.520',
    raw: 'data: [DONE]',
    parsed: null,
    type: 'finish'
  }
]

const activeTab = ref('request')
const searchQuery = ref('')
const streamViewMode = ref<'raw' | 'parsed'>('parsed')
const isStreaming = ref(false)
const currentChunks = ref<typeof sampleChunks>([])

// Simulate streaming
const simulateStream = async () => {
  isStreaming.value = true
  currentChunks.value = []
  
  for (const chunk of sampleChunks) {
    await new Promise(resolve => setTimeout(resolve, 300))
    currentChunks.value.push(chunk)
  }
  
  isStreaming.value = false
}

// Load demo data
onMounted(() => {
  currentChunks.value = sampleChunks
})

const tabs = [
  { label: '请求参数', slot: 'request', icon: 'i-lucide-send' },
  { label: '响应结果', slot: 'response', icon: 'i-lucide-message-square' },
  { label: '流式响应', slot: 'stream', icon: 'i-lucide-radio' }
]

const metadata = computed(() => ({
  requestId: 'req_abc123xyz789',
  timestamp: '2024-04-09 14:32:18',
  duration: '2.52s',
  model: sampleRequest.model,
  tokens: sampleResponse.usage.total_tokens
}))
</script>

<template>
  <div class="min-h-screen bg-default">
    <!-- Header -->
    <header class="border-b border-default bg-elevated/50 backdrop-blur-sm sticky top-0 z-50">
      <div class="container mx-auto px-4 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
              <UIcon name="i-lucide-brain-circuit" class="size-6 text-primary" />
            </div>
            <div>
              <h1 class="text-lg font-semibold text-default">AI Request Inspector</h1>
              <p class="text-xs text-muted">模型请求与响应调试查看器</p>
            </div>
          </div>
          
          <div class="flex items-center gap-3">
            <UColorModeButton variant="ghost" />
            <UButton
              variant="outline"
              size="sm"
              icon="i-lucide-play"
              :loading="isStreaming"
              @click="simulateStream"
            >
              模拟流式
            </UButton>
          </div>
        </div>
      </div>
    </header>
    
    <!-- Metadata Bar -->
    <div class="border-b border-default bg-muted/30">
      <div class="container mx-auto px-4 py-3">
        <div class="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-hash" class="size-4 text-muted" />
            <span class="text-muted">Request ID:</span>
            <code class="font-mono text-xs text-default bg-elevated px-2 py-0.5 rounded">{{ metadata.requestId }}</code>
          </div>
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-clock" class="size-4 text-muted" />
            <span class="text-muted">时间:</span>
            <span class="text-default">{{ metadata.timestamp }}</span>
          </div>
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-timer" class="size-4 text-muted" />
            <span class="text-muted">耗时:</span>
            <span class="text-default">{{ metadata.duration }}</span>
          </div>
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-cpu" class="size-4 text-muted" />
            <span class="text-muted">模型:</span>
            <UBadge :label="metadata.model" variant="subtle" size="sm" />
          </div>
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-coins" class="size-4 text-muted" />
            <span class="text-muted">Tokens:</span>
            <span class="text-default font-mono">{{ metadata.tokens }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Search Bar -->
    <div class="border-b border-default">
      <div class="container mx-auto px-4 py-3">
        <UInput
          v-model="searchQuery"
          placeholder="搜索 JSON 内容..."
          icon="i-lucide-search"
          size="lg"
          class="max-w-md"
          :ui="{ base: 'bg-elevated' }"
        >
          <template v-if="searchQuery" #trailing>
            <UButton
              variant="ghost"
              size="xs"
              icon="i-lucide-x"
              color="neutral"
              @click="searchQuery = ''"
            />
</template>

<style scoped>
.streaming-dot {
  animation: pulse 1.4s ease-in-out infinite;
}

.streaming-dot:nth-child(2) {
  animation-delay: 0.2s;
}

.streaming-dot:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.3;
    transform: scale(0.8);
  }
  50% {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
        </UInput>
      </div>
    </div>
    
    <!-- Main Content -->
    <main class="container mx-auto px-4 py-6">
      <UTabs
        :items="tabs"
        :default-value="activeTab"
        class="w-full"
      >
        <template #request>
          <div class="mt-6 border border-default rounded-xl overflow-hidden bg-elevated shadow-sm">
            <RequestPanel
              :data="sampleRequest"
              title="请求参数"
              icon="i-lucide-send"
              :search-query="searchQuery"
            />
          </div>
        </template>
        
        <template #response>
          <div class="mt-6 border border-default rounded-xl overflow-hidden bg-elevated shadow-sm">
            <RequestPanel
              :data="sampleResponse"
              title="响应结果"
              icon="i-lucide-message-square"
              :search-query="searchQuery"
            />
          </div>
        </template>
        
        <template #stream>
          <div class="mt-6">
            <!-- Stream Controls -->
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-3">
                <div v-if="isStreaming" class="flex items-center gap-2 text-primary">
                  <span class="flex gap-1">
                    <span class="w-2 h-2 rounded-full bg-primary streaming-dot" />
                    <span class="w-2 h-2 rounded-full bg-primary streaming-dot" />
                    <span class="w-2 h-2 rounded-full bg-primary streaming-dot" />
                  </span>
                  <span class="text-sm">接收中...</span>
                </div>
                <UBadge v-else :label="`${currentChunks.length} chunks`" variant="subtle" />
              </div>
              
              <div class="inline-flex rounded-lg border border-default overflow-hidden">
                <UButton
                  :variant="streamViewMode === 'parsed' ? 'solid' : 'ghost'"
                  size="sm"
                  icon="i-lucide-layers"
                  class="rounded-none"
                  @click="streamViewMode = 'parsed'"
                >
                  组装视图
                </UButton>
                <UButton
                  :variant="streamViewMode === 'raw' ? 'solid' : 'ghost'"
                  size="sm"
                  icon="i-lucide-list"
                  class="rounded-none"
                  @click="streamViewMode = 'raw'"
                >
                  原始 Chunk
                </UButton>
              </div>
            </div>
            
            <div class="border border-default rounded-xl overflow-hidden bg-elevated shadow-sm min-h-[500px]">
              <StreamChunkViewer
                :chunks="currentChunks"
                :view-mode="streamViewMode"
                :search-query="searchQuery"
              />
            </div>
          </div>
        </template>
      </UTabs>
    </main>
    
    <!-- Footer -->
    <footer class="border-t border-default mt-12">
      <div class="container mx-auto px-4 py-6">
        <div class="flex items-center justify-between text-sm text-muted">
          <span>AI Request Inspector - 模型调试工具</span>
          <div class="flex items-center gap-4">
            <a href="#" class="hover:text-default transition-colors">文档</a>
            <a href="#" class="hover:text-default transition-colors">GitHub</a>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>
