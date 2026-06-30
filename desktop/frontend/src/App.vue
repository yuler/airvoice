<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConnection } from './composables/useConnection'
import Header from './components/Header.vue'
import QRCode from './components/QRCode.vue'
import StatusBadge from './components/StatusBadge.vue'
import LogPanel from './components/LogPanel.vue'
import DisconnectButton from './components/DisconnectButton.vue'
import Toast from './components/Toast.vue'

const { locale } = useI18n()
const { status } = useConnection()

const isConnected = computed(() => status.value.state === 'connected')

onMounted(async () => {
  try {
    const settings = await window.go.main.App.GetSettings()
    if (settings && settings.language) {
      locale.value = settings.language
    }
  } catch (e) {
    console.error('Failed to load language setting:', e)
  }
})
</script>

<template>
  <div class="relative w-full h-full flex flex-col bg-bg-primary">
    <Toast />
    <Header />
    
    <main class="flex-1 min-h-0 flex flex-col p-6" :class="isConnected ? 'gap-4' : 'items-center justify-center overflow-y-auto'">
      <template v-if="isConnected">
        <StatusBadge class="shrink-0" />
        <LogPanel class="min-h-0" />
      </template>
      <template v-else>
        <QRCode />
      </template>
    </main>
    
    <footer class="flex justify-center pb-6">
      <DisconnectButton v-if="isConnected" />
    </footer>
  </div>
</template>
