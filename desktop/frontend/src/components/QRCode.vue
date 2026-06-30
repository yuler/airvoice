<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetPairingLink, GetQRCode, RefreshPairing } from '../../wailsjs/go/main/App'
import { useToast } from '../composables/useToast'
import RefreshIcon from './icons/RefreshIcon.vue'

const qrCodeData = ref<string>('')
const pairingLink = ref<string>('')
const error = ref<string>('')
const refreshing = ref(false)
const { t } = useI18n()
const { show } = useToast()

async function loadQRCode() {
  try {
    qrCodeData.value = await GetQRCode()
    pairingLink.value = await GetPairingLink()
  } catch (e) {
    error.value = t('qr.error')
    console.error(e)
  }
}

async function refreshPairing() {
  if (refreshing.value) return

  refreshing.value = true
  error.value = ''

  try {
    await RefreshPairing()
    await loadQRCode()
    show(t('qr.refreshed'), 'success')
  } catch (e) {
    error.value = t('qr.error')
    show(t('qr.refreshError'), 'error')
    console.error(e)
  } finally {
    refreshing.value = false
  }
}

onMounted(() => {
  loadQRCode()
  const runtime = (window as any).runtime
  if (runtime && runtime.EventsOn) {
    runtime.EventsOn('server_restarted', loadQRCode)
  }
})

onUnmounted(() => {
  const runtime = (window as any).runtime
  if (runtime && runtime.EventsOff) {
    runtime.EventsOff('server_restarted', loadQRCode)
  }
})
</script>

<template>
  <div class="flex flex-col items-center gap-6 w-full">
    <div class="text-center">
      <h2 class="text-2xl font-semibold tracking-tight text-primary-text">{{ t('pair.title') }}</h2>
      <p class="text-base text-secondary-text mt-2">{{ t('pair.subtitle') }}</p>
    </div>
    
    <div v-if="error" class="text-status-error text-sm">{{ error }}</div>
    <div v-else-if="qrCodeData" class="w-52 h-52 rounded-xl flex items-center justify-center p-4 overflow-hidden" style="background-color: var(--color-bg-primary); border: 1px solid var(--color-border-default);">
      <img :src="qrCodeData" alt="QR Code" class="block w-full h-full" style="image-rendering: pixelated;" />
    </div>
    <div v-else class="text-secondary-text">{{ t('qr.loading') }}</div>
    
    <button
      type="button"
      :disabled="refreshing"
      @click="refreshPairing"
      class="w-11 h-11 flex items-center justify-center rounded-full border border-border-default text-primary-text hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
    >
      <RefreshIcon :class="{ 'animate-spin': refreshing }" />
    </button>
  </div>
</template>
