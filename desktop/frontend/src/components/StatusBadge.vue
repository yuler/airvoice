<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConnection } from '../composables/useConnection'

const { status } = useConnection()
const { t } = useI18n()

const isConnected = computed(() => status.value.state === 'connected')
const isConnecting = computed(() => status.value.state === 'connecting' || status.value.state === 'waiting')

const badgeClass = computed(() => {
  if (isConnected.value) return 'bg-status-success'
  if (isConnecting.value) return 'bg-status-warning'
  return 'bg-status-neutral'
})

const statusText = computed(() => {
  if (isConnected.value) return t('status.connected', { device: status.value.deviceName })
  if (isConnecting.value) return t('status.connecting')
  return t('status.disconnected')
})
</script>

<template>
  <div class="flex flex-col items-center gap-4">
    <div v-if="isConnected" class="w-16 h-16 rounded-full bg-status-success flex items-center justify-center">
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M5 13l4 4L19 7"/>
      </svg>
    </div>
    
    <div class="text-center">
      <h1 class="text-2xl font-semibold tracking-tight text-primary-text">
        {{ isConnected ? t('status.connectedTitle') : t('status.waitingTitle') }}
      </h1>
      <p v-if="isConnected" class="text-base text-secondary-text mt-2">{{ status.deviceName }}</p>
    </div>
    
    <div class="flex items-center gap-2 px-4 py-3 bg-bg-secondary border border-border-default rounded-xl">
      <span class="w-2 h-2 rounded-full" :class="badgeClass"></span>
      <span class="text-sm text-secondary-text font-mono">{{ statusText }}</span>
    </div>
  </div>
</template>
