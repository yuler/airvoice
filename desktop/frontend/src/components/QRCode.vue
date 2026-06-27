<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const qrCodeData = ref<string>('')
const error = ref<string>('')
const { t } = useI18n()

async function loadQRCode() {
  try {
    qrCodeData.value = await window.go.main.App.GetQRCode()
  } catch (e) {
    error.value = 'Failed to generate QR code'
    console.error(e)
  }
}

onMounted(loadQRCode)
</script>

<template>
  <div class="flex flex-col items-center p-4">
    <div v-if="error" class="text-status-error text-sm">{{ error }}</div>
    <div v-else-if="qrCodeData" class="bg-white p-4 rounded-lg">
      <img :src="qrCodeData" alt="QR Code" class="w-48 h-48" />
    </div>
    <div v-else class="text-text-secondary">{{ t('qr.loading') }}</div>
    <p class="text-text-muted text-xs mt-2">{{ t('qr.scan') }}</p>
  </div>
</template>
