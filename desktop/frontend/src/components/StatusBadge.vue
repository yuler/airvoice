<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConnection } from '../composables/useConnection'

const { status } = useConnection()
const { t } = useI18n()

const badgeClass = computed(() => {
  switch (status.value.state) {
    case 'connected':
      return 'bg-status-success'
    case 'connecting':
    case 'waiting':
      return 'bg-status-warning'
    case 'disconnected':
    default:
      return 'bg-text-muted'
  }
})

const statusText = computed(() => {
  switch (status.value.state) {
    case 'connected':
      return t('status.connected', { device: status.value.deviceName })
    case 'connecting':
      return t('status.connecting')
    case 'waiting':
      return t('status.waiting')
    case 'disconnected':
    default:
      return t('status.disconnected')
  }
})
</script>

<template>
  <div class="flex items-center gap-2 px-3 py-2 rounded-full border border-border-default">
    <span class="w-2 h-2 rounded-full" :class="badgeClass"></span>
    <span class="text-sm text-text-secondary">{{ statusText }}</span>
  </div>
</template>
