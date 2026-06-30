<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from './components/QRCode.vue'
import StatusBadge from './components/StatusBadge.vue'
import HistoryList from './components/HistoryList.vue'
import SettingsPanel from './components/SettingsPanel.vue'
import LanguageSwitch from './components/LanguageSwitch.vue'

const { locale } = useI18n()

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
  <div class="min-h-screen bg-bg-primary text-text-primary flex flex-col">
    <header class="p-4 border-b border-border-default flex items-center justify-between">
      <h1 class="text-xl font-semibold">Airvoice</h1>
      <div class="flex items-center gap-2">
        <LanguageSwitch />
        <SettingsPanel />
      </div>
    </header>

    <main class="flex-1 flex flex-col items-center justify-center p-4">
      <QRCode />
      <StatusBadge class="mt-4" />
    </main>

    <footer class="flex-1 border-t border-border-default">
      <HistoryList />
    </footer>
  </div>
</template>
