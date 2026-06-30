<script setup lang="ts">
import { computed } from 'vue'
import { useConnection } from '../composables/useConnection'
import { useTheme } from '../composables/useTheme'
import SunIcon from './icons/SunIcon.vue'
import MoonIcon from './icons/MoonIcon.vue'

const { status } = useConnection()
const { theme, toggle } = useTheme()

const isConnected = computed(() => status.value.state === 'connected')
</script>

<template>
  <header class="flex items-center justify-between px-6 py-4" style="--wails-draggable: drag">
    <div class="flex items-center gap-2">
      <span 
        v-if="isConnected" 
        class="w-2 h-2 rounded-full bg-status-success"
      />
      <span class="text-lg font-semibold text-primary-text">Airvoice</span>
    </div>
    <button
      @click="toggle"
      class="w-8 h-8 flex items-center justify-center text-secondary-text hover:text-primary-text transition-colors"
      style="--wails-draggable: no-drag"
    >
      <SunIcon v-if="theme === 'dark'" />
      <MoonIcon v-else />
    </button>
  </header>
</template>
