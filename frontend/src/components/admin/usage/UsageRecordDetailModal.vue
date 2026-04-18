<template>
  <BaseDialog :show="show" :title="title" width="ultra-wide" @close="close">
    <UsageRecordInspector
      :detail="detail"
      :loading="loading"
      empty-message="暂无详情"
    />
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import UsageRecordInspector from './UsageRecordInspector.vue'
import { adminUsageAPI, type AdminUsageLogDetail } from '@/api/admin/usage'
import type { AdminUsageLog } from '@/types'
import { useAppStore } from '@/stores'

const props = defineProps<{
  show: boolean
  row: AdminUsageLog | null
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const appStore = useAppStore()

const loading = ref(false)
const detail = ref<AdminUsageLogDetail | null>(null)

const title = computed(() => {
  const requestId = detail.value?.request_id || props.row?.request_id
  return requestId ? `使用记录详情 · ${requestId}` : '使用记录详情'
})

function close() {
  emit('update:show', false)
}

async function loadDetail() {
  if (!props.show || !props.row?.id) return
  loading.value = true
  try {
    detail.value = await adminUsageAPI.getById(props.row.id)
  } catch (error: any) {
    detail.value = null
    appStore.showError(error?.message || '加载使用记录详情失败')
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.row?.id] as const,
  ([show]) => {
    if (!show) {
      detail.value = null
      return
    }
    loadDetail()
  },
  { immediate: true }
)
</script>
