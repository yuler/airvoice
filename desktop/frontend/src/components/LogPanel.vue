<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const logs = ref<string[]>([])
const logContainer = ref<HTMLElement>()

function addLog(message: string) {
  const time = new Date().toLocaleTimeString('en-US', { hour12: false })
  logs.value.push(`[${time}] ${message}`)
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
}

onMounted(() => {
  addLog(t('log.connected'))
  addLog(t('log.handshake'))
  addLog(t('log.listening'))
})

onUnmounted(() => {
  logs.value = []
})
</script>

<template>
  <div 
    ref="logContainer"
    class="w-full h-32 p-3 bg-bg-primary border border-border-default rounded-xl overflow-y-auto font-mono text-xs leading-relaxed"
  >
    <div v-for="(log, index) in logs" :key="index" class="text-muted-text">
      {{ log }}
    </div>
  </div>
</template>
