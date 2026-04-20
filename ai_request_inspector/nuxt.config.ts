export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ['@nuxt/ui'],
  css: ['~/assets/css/main.css'],
  compatibilityDate: '2025-01-01',
  app: {
    head: {
      title: 'AI Request Inspector',
      meta: [
        { name: 'description', content: 'AI 模型请求/响应调试查看器' }
      ]
    }
  }
})
