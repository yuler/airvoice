<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useHistory } from '../composables/useHistory'

const { entries, loading, loadHistory, searchHistory, clearHistory } = useHistory()
const { t } = useI18n()
const query = ref('')

let debounceTimer: ReturnType<typeof setTimeout> | null = null

onMounted(() => loadHistory())

function onSearch() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    if (query.value.trim()) {
      searchHistory(query.value.trim())
    } else {
      loadHistory()
    }
  }, 300)
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="flex items-center justify-between px-4 py-2 border-b border-border-default">
      <h2 class="text-sm font-medium text-text-secondary">{{ t('history.title') }}</h2>
      <button
        v-if="entries.length > 0"
        @click="clearHistory"
        class="text-xs text-text-muted hover:text-status-error"
      >
        {{ t('history.clear') }}
      </button>
    </div>

    <div class="px-4 py-2 border-b border-border-default">
      <input
        v-model="query"
        @input="onSearch"
        type="text"
        :placeholder="t('history.search')"
        class="w-full px-3 py-1.5 text-sm bg-bg-primary border border-border-default rounded-md text-text-primary placeholder:text-text-muted"
      />
    </div>

    <div class="flex-1 overflow-y-auto">
      <div v-if="loading" class="p-4 text-center text-text-muted">{{ t('history.loading') }}</div>
      <div v-else-if="entries.length === 0" class="p-4 text-center text-text-muted">
        {{ t('history.empty') }}
      </div>
      <div
        v-for="entry in entries"
        :key="entry.id"
        class="px-4 py-3 border-b border-border-default hover:bg-bg-secondary"
      >
        <p class="text-sm text-text-primary truncate">{{ entry.content }}</p>
        <div class="flex items-center gap-2 mt-1">
          <span class="text-xs text-text-muted">{{ entry.device }}</span>
          <span class="text-xs text-text-muted">{{ formatTime(entry.createdAt) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
