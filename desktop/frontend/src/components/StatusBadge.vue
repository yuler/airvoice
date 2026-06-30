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

const statusLabel = computed(() => {
  if (isConnected.value) return t('status.connectedShort')
  if (isConnecting.value) return t('status.connecting')
  return t('status.disconnected')
})

const connectionAddress = computed(() => {
  if (status.value.host && status.value.port) {
    return `${status.value.host}:${status.value.port}`
  }
  if (status.value.deviceName) {
    return status.value.deviceName
  }
  return ''
})
</script>

<template>
  <div class="flex w-full flex-col items-center gap-4">
    <div v-if="isConnected" class="flex h-16 w-16 items-center justify-center rounded-full bg-status-success">
      <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M5 13l4 4L19 7"/>
      </svg>
    </div>

    <div class="text-center">
      <h1 class="text-2xl font-semibold tracking-tight text-primary-text">
        {{ isConnected ? t('status.connectedTitle') : t('status.waitingTitle') }}
      </h1>
    </div>

    <div class="flex w-full items-center gap-2 rounded-xl border border-border-default bg-bg-secondary px-4 py-3">
      <span class="h-2 w-2 shrink-0 rounded-full" :class="badgeClass" />
      <span class="shrink-0 text-sm text-secondary-text">{{ statusLabel }}</span>
      <template v-if="isConnected && connectionAddress">
        <span class="text-muted-text">·</span>
        <span class="min-w-0 truncate font-mono text-sm text-muted-text">{{ connectionAddress }}</span>
      </template>
    </div>
  </div>
</template>
