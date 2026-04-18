<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  text: string
  /** 纯文本模式：不解析 markdown 语法（用于 tool 输出等） */
  plain?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  plain: false
})

function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

function applyInline(html: string): string {
  html = html.replace(/`([^`\n]+)`/g, '<code>$1</code>')
  html = html.replace(/\*\*([^\n*]+?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/(^|[\s(])\*([^*\n]+?)\*(?=[\s.,)!?:]|$)/g, '$1<em>$2</em>')
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>')
  return html
}

function renderMarkdownLine(line: string): string {
  if (!line) return '&nbsp;'
  const lead = line.match(/^(\s+)/)
  const leadStr = lead ? lead[0] : ''
  const rest = line.slice(leadStr.length)
  let html = escapeHtml(rest)

  const hMatch = html.match(/^(#{1,6})\s+(.+)$/)
  if (hMatch) {
    const level = hMatch[1].length
    html = `<span class="rd-md-h rd-md-h${level}">${applyInline(hMatch[2])}</span>`
  } else {
    const bullet = html.match(/^([-*+])\s+(.+)$/)
    const order = html.match(/^(\d+)\.\s+(.+)$/)
    if (bullet) {
      html = `<span class="rd-md-bullet">•</span> ${applyInline(bullet[2])}`
    } else if (order) {
      html = `<span class="rd-md-bullet">${order[1]}.</span> ${applyInline(order[2])}`
    } else {
      html = applyInline(html)
    }
  }

  const pad = leadStr.replace(/\t/g, '    ').replace(/ /g, '&nbsp;')
  return `${pad}${html}`
}

interface Row {
  num: number
  kind: 'text' | 'code' | 'fence' | 'plain'
  html: string
}

const rows = computed<Row[]>(() => {
  const lines = (props.text || '').split('\n')
  const result: Row[] = []
  let inCode = false
  lines.forEach((line, i) => {
    const num = i + 1
    if (props.plain) {
      result.push({ num, kind: 'plain', html: escapeHtml(line) || '&nbsp;' })
      return
    }
    const fence = line.match(/^\s*```(\w*)\s*$/)
    if (fence) {
      inCode = !inCode
      result.push({ num, kind: 'fence', html: `<span class="rd-md-fence-txt">${escapeHtml(line)}</span>` })
      return
    }
    if (inCode) {
      result.push({ num, kind: 'code', html: `<code>${escapeHtml(line) || '&nbsp;'}</code>` })
      return
    }
    result.push({ num, kind: 'text', html: renderMarkdownLine(line) })
  })
  return result
})
</script>

<template>
  <div class="rd-md">
    <div v-for="row in rows" :key="row.num" class="rd-md-row" :class="`rd-md-row-${row.kind}`">
      <span class="rd-md-gutter">{{ row.num }}</span>
      <span class="rd-md-line" v-html="row.html"></span>
    </div>
  </div>
</template>

<style scoped>
.rd-md {
  display: block;
  background: white;
  overflow: hidden;
  line-height: 1.6;
  border-radius: 0;
}

:global(.dark) .rd-md {
  background: rgb(30 41 59);
}

.rd-md-row {
  display: flex;
  align-items: stretch;
  min-height: 1.6em;
  width: 100%;
}

.rd-md-row:hover {
  background: rgb(248 250 252);
}

:global(.dark) .rd-md-row:hover {
  background: rgb(15 23 42);
}

.rd-md-gutter {
  flex-shrink: 0;
  width: 3rem;
  padding: 0.125rem 0.5rem 0.125rem 0.375rem;
  text-align: right;
  color: rgb(148 163 184);
  background: rgb(248 250 252);
  border-right: 1px solid rgb(203 213 225);
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.6875rem;
  font-variant-numeric: tabular-nums;
  user-select: none;
  line-height: 1.6;
}

:global(.dark) .rd-md-gutter {
  color: rgb(100 116 139);
  background: rgb(15 23 42);
  border-right-color: rgb(71 85 105);
}

.rd-md-line {
  flex: 1 1 0;
  min-width: 0;
  max-width: 100%;
  padding: 0.125rem 0.875rem 0.125rem 0.875rem;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
  font-size: 0.875rem;
  line-height: 1.6;
  color: rgb(15 23 42);
}

:global(.dark) .rd-md-line {
  color: rgb(241 245 249);
}

.rd-md-row-code,
.rd-md-row-fence,
.rd-md-row-plain {
  background: rgb(248 250 252);
}

:global(.dark) .rd-md-row-code,
:global(.dark) .rd-md-row-fence,
:global(.dark) .rd-md-row-plain {
  background: rgb(15 23 42);
}

.rd-md-row-code .rd-md-line,
.rd-md-row-fence .rd-md-line,
.rd-md-row-plain .rd-md-line {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.8125rem;
}

.rd-md-row-code:hover,
.rd-md-row-fence:hover,
.rd-md-row-plain:hover {
  background: rgb(241 245 249);
}

:global(.dark) .rd-md-row-code:hover,
:global(.dark) .rd-md-row-fence:hover,
:global(.dark) .rd-md-row-plain:hover {
  background: rgb(30 41 59);
}

.rd-md-row :deep(code) {
  background: rgb(241 245 249);
  padding: 0.0625rem 0.25rem;
  border-radius: 3px;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 0.8125em;
}

.rd-md-row-code :deep(code),
.rd-md-row-fence :deep(code) {
  background: none;
  padding: 0;
  font-family: inherit;
  font-size: inherit;
}

:global(.dark) .rd-md-row :deep(code) {
  background: rgb(51 65 85);
  color: rgb(226 232 240);
}

.rd-md-row :deep(a) {
  color: rgb(20 184 166);
  text-decoration: none;
}

.rd-md-row :deep(a:hover) {
  text-decoration: underline;
}

.rd-md-row :deep(.rd-md-h) {
  display: inline-block;
  font-weight: 600;
}

.rd-md-row :deep(.rd-md-h1) {
  font-size: 1.0625rem;
}

.rd-md-row :deep(.rd-md-h2) {
  font-size: 1rem;
}

.rd-md-row :deep(.rd-md-h3) {
  font-size: 0.9375rem;
}

.rd-md-row :deep(.rd-md-bullet) {
  display: inline-block;
  min-width: 1rem;
  color: rgb(20 184 166);
  font-weight: 600;
}

.rd-md-row :deep(.rd-md-fence-txt) {
  color: rgb(148 163 184);
  font-family: inherit;
}

.rd-md-row :deep(strong) {
  font-weight: 600;
}
</style>
