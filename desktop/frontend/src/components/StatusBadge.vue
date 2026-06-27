<script setup lang="ts">
import { ref, onMounted } from 'vue'

const status = ref<string>('disconnected')

async function loadStatus() {
  try {
    status.value = await window.go.main.App.GetConnectionStatus()
  } catch (e) {
    console.error(e)
  }
}

onMounted(loadStatus)
</script>

<template>
  <div class="flex items-center gap-2 text-sm">
    <span
      class="w-2 h-2 rounded-full"
      :class="{
        'bg-status-success': status === 'connected',
        'bg-status-warning': status === 'connecting',
        'bg-status-error': status === 'disconnected',
      }"
    />
    <span class="text-text-secondary capitalize">{{ status }}</span>
  </div>
</template>
