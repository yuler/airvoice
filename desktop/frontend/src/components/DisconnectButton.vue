<script setup lang="ts">
import { useConnection } from '../composables/useConnection'
import LinkOffIcon from './icons/LinkOffIcon.vue'

const { disconnect } = useConnection()
</script>

<template>
  <button
    @click="disconnect"
    class="w-11 h-11 flex items-center justify-center rounded-full border-2 border-status-error text-status-error hover:bg-status-error hover:text-white transition-colors"
  >
    <LinkOffIcon />
  </button>
</template>
