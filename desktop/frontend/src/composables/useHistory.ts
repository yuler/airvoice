import { ref } from 'vue'

interface HistoryEntry {
  id: number
  content: string
  device: string
  createdAt: string
}

export function useHistory() {
  const entries = ref<HistoryEntry[]>([])
  const loading = ref(false)

  async function loadHistory(limit = 50) {
    loading.value = true
    try {
      entries.value = await window.go.main.App.GetHistory(limit)
    } catch (e) {
      console.error('Failed to load history:', e)
    } finally {
      loading.value = false
    }
  }

  async function clearHistory() {
    try {
      await window.go.main.App.ClearHistory()
      entries.value = []
    } catch (e) {
      console.error('Failed to clear history:', e)
    }
  }

  return { entries, loading, loadHistory, clearHistory }
}
