<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()

async function toggleLanguage() {
  const newLocale = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  locale.value = newLocale
  try {
    const settings = await window.go.main.App.GetSettings()
    settings.language = newLocale
    await window.go.main.App.SaveSettings(settings)
  } catch (e) {
    console.error('Failed to save language setting:', e)
  }
}
</script>

<template>
  <button
    @click="toggleLanguage"
    class="px-2 py-1 text-xs text-text-secondary hover:text-text-primary border border-border-default rounded"
  >
    {{ locale === 'zh-CN' ? 'EN' : '中文' }}
  </button>
</template>
