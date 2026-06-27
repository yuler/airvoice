<script setup lang="ts">
import QRCode from './components/QRCode.vue'
import StatusBadge from './components/StatusBadge.vue'
</script>

<template>
  <div class="min-h-screen bg-bg-primary text-text-primary flex flex-col">
    <header class="p-4 border-b border-border-default">
      <h1 class="text-xl font-semibold">Airvoice</h1>
    </header>

    <main class="flex-1 flex flex-col items-center justify-center p-4">
      <QRCode />
      <StatusBadge class="mt-4" />
    </main>
  </div>
</template>
