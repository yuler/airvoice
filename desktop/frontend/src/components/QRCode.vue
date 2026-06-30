<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import RefreshIcon from './icons/RefreshIcon.vue'

const qrCodeData = ref<string>('')
const pairingLink = ref<string>('')
const error = ref<string>('')
const { t } = useI18n()

async function loadQRCode() {
  try {
    qrCodeData.value = await window.go.main.App.GetQRCode()
    pairingLink.value = await window.go.main.App.GetPairingLink()
  } catch (e) {
    error.value = t('qr.error')
    console.error(e)
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
    <div v-else-if="qrCodeData" class="w-full p-6 bg-bg-primary border border-border-default rounded-xl flex items-center justify-center">
      <img :src="qrCodeData" alt="QR Code" class="block max-w-full" style="width: 200px; height: 200px;" />
    </div>
    <div v-else class="text-secondary-text">{{ t('qr.loading') }}</div>
    
    <button
      @click="loadQRCode"
      class="w-11 h-11 flex items-center justify-center rounded-full border border-border-default text-primary-text hover:bg-bg-secondary transition-colors"
    >
      <RefreshIcon />
    </button>
  </div>
</template>
