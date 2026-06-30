<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

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
  const runtime = (window as any).runtime
  if (runtime && runtime.EventsOn) {
    runtime.EventsOn('log_added', (msg: string) => {
      addLog(msg)
    })
  }
})

onUnmounted(() => {
  const runtime = (window as any).runtime
  if (runtime && runtime.EventsOff) {
    runtime.EventsOff('log_added')
  }
  logs.value = []
})
</script>

<template>
  <div 
    ref="logContainer"
    class="w-full h-32 p-3 bg-bg-primary border border-border-default rounded-xl overflow-y-auto font-mono text-xs leading-relaxed"
  >
    <div v-for="(log, index) in logs" :key="index" class="text-text-muted">
      {{ log }}
    </div>
  </div>
</template>
