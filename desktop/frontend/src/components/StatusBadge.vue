<script setup lang="ts">
import { computed } from 'vue'
import { useConnection } from '../composables/useConnection'

const { status } = useConnection()

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
      return `Connected to ${status.value.deviceName}`
    case 'connecting':
      return 'Connecting...'
    case 'waiting':
      return 'Waiting for device'
    case 'disconnected':
    default:
      return 'Disconnected'
  }
})
</script>

<template>
  <div class="flex items-center gap-2 px-3 py-2 rounded-full border border-border-default">
    <span class="w-2 h-2 rounded-full" :class="badgeClass"></span>
    <span class="text-sm text-text-secondary">{{ statusText }}</span>
  </div>
</template>
