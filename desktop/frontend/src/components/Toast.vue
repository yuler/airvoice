<script setup lang="ts">
import { useToast, type ToastType } from '../composables/useToast'

const { toasts, dismiss } = useToast()

function dotClass(type: ToastType) {
  switch (type) {
    case 'success':
      return 'bg-status-success'
    case 'error':
      return 'bg-status-error'
    default:
      return 'bg-accent-blue'
  }
}
</script>

<template>
  <div
    class="absolute inset-x-0 top-0 z-50 flex flex-col gap-2 px-6 pt-6 pointer-events-none"
    aria-live="polite"
  >
    <TransitionGroup name="toast">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast-item pointer-events-auto"
        role="status"
        @click="dismiss(toast.id)"
      >
        <span class="toast-dot" :class="dotClass(toast.type)" />
        <span class="toast-message">{{ toast.message }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>
