<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const qrCodeData = ref<string>('')
const pairingLink = ref<string>('')
const error = ref<string>('')
const copied = ref(false)
const { t } = useI18n()

async function loadQRCode() {
  try {
    qrCodeData.value = await window.go.main.App.GetQRCode()
    pairingLink.value = await window.go.main.App.GetPairingLink()
  } catch (e) {
    error.value = 'Failed to generate QR code'
    console.error(e)
  }
}

async function copyLink() {
  if (!pairingLink.value) return
  try {
    await navigator.clipboard.writeText(pairingLink.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch (e) {
    console.error('Failed to copy:', e)
  }
}

onMounted(() => {
  loadQRCode()
  const runtime = (window as any).runtime
  if (runtime && runtime.EventsOn) {
    runtime.EventsOn('server_restarted', loadQRCode)
  }
})
</script>

<template>
  <div class="flex flex-col items-center p-4">
    <div v-if="error" class="text-status-error text-sm">{{ error }}</div>
    <div v-else-if="qrCodeData" class="bg-white p-4 rounded-lg">
      <img :src="qrCodeData" alt="QR Code" class="w-48 h-48" />
    </div>
    <div v-else class="text-text-secondary">{{ t('qr.loading') }}</div>
    <p class="text-text-muted text-xs mt-2">{{ t('qr.scan') }}</p>
    <button
      v-if="pairingLink"
      @click="copyLink"
      class="mt-2 px-3 py-1 text-xs rounded-full border border-border-default text-text-secondary hover:text-text-primary transition-colors"
    >
      {{ copied ? t('qr.copied') : t('qr.copyLink') }}
    </button>
  </div>
</template>
