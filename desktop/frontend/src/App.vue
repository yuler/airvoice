<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConnection } from './composables/useConnection'
import Header from './components/Header.vue'
import QRCode from './components/QRCode.vue'
import StatusBadge from './components/StatusBadge.vue'
import LogPanel from './components/LogPanel.vue'
import DisconnectButton from './components/DisconnectButton.vue'

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
  <div class="w-full h-full flex flex-col bg-bg-primary">
    <Header />
    
    <main class="flex-1 flex flex-col items-center justify-center px-6">
      <template v-if="isConnected">
        <StatusBadge class="mb-6" />
        <LogPanel class="mb-6" />
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
