<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()

function toggleLanguage() {
  locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
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
