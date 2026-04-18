<template>
  <Icon :name="resolvedName" :size="size" :stroke-width="strokeWidth" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  name: string
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  strokeWidth?: number
}>(), {
  size: 'md',
  strokeWidth: 1.5
})

const iconMap: Record<string, string> = {
  'i-lucide-play': 'play',
  'i-lucide-search': 'search',
  'i-lucide-x': 'x',
  'i-lucide-send': 'arrowUp',
  'i-lucide-message-square': 'chatBubble',
  'i-lucide-radio': 'bolt',
  'i-lucide-brain-circuit': 'brain',
  'i-lucide-hash': 'grid',
  'i-lucide-clock': 'clock',
  'i-lucide-timer': 'clock',
  'i-lucide-cpu': 'cpu',
  'i-lucide-coins': 'dollar',
  'i-lucide-layout': 'grid',
  'i-lucide-file-json': 'document',
  'i-lucide-copy': 'copy',
  'i-lucide-check': 'check',
  'i-lucide-messages-square': 'chatBubble',
  'i-lucide-settings-2': 'cog',
  'i-lucide-file-x': 'document',
  'i-lucide-function-square': 'calculator',
  'i-lucide-wrench': 'cog',
  'i-lucide-check-circle': 'checkCircle',
  'i-lucide-flag': 'badge',
  'i-lucide-alert-circle': 'exclamationCircle',
  'i-lucide-code': 'terminal',
  'i-lucide-layers': 'cube',
  'i-lucide-list': 'menu',
  'i-lucide-braces': 'document',
  'i-lucide-chevron-right': 'chevronRight',
  'i-lucide-chevron-down': 'chevronDown'
}

const resolvedName = computed(() => {
  return (iconMap[props.name] || 'sparkles') as any
})
</script>
