<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConnection } from '../composables/useConnection'

const { status } = useConnection()
const { t } = useI18n()

const isRelay = computed(() => status.value.mode === 'relay')
const isConnected = computed(() => status.value.state === 'connected')
const isWarning = computed(() =>
  status.value.state === 'connecting' ||
  status.value.state === 'waiting' ||
  (isRelay.value && status.value.state === 'disconnected')
)
const isError = computed(() => status.value.state === 'error')

const badgeClass = computed(() => {
  if (isConnected.value) return 'bg-status-success'
  if (isError.value) return 'bg-status-error'
  if (isWarning.value) return 'bg-status-warning'
  return 'bg-status-neutral'
})

const statusLabel = computed(() => {
  if (isConnected.value) {
    return isRelay.value ? t('status.connectedRelay') : t('status.connectedLan')
  }
  if (status.value.state === 'waiting') {
    return t('status.waiting')
  }
  if (status.value.state === 'connecting') {
    return isRelay.value ? t('status.relayConnecting') : t('status.connecting')
  }
  if (isError.value) {
    return status.value.error || t('status.relayError')
  }
  if (isRelay.value) {
    return t('status.relayReconnecting')
  }
  return t('status.disconnected')
})

const connectionAddress = computed(() => {
  if (isRelay.value) {
    if (status.value.deviceName && status.value.host) {
      return `${status.value.deviceName} · ${status.value.host}`
    }
    return status.value.host || status.value.deviceName || ''
  }
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
