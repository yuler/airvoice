import { ref, onMounted, onUnmounted } from 'vue'

interface ConnectionStatus {
  state: 'disconnected' | 'connecting' | 'connected' | 'waiting'
  deviceName: string
  host: string
  port: number
}

export function useConnection() {
  const status = ref<ConnectionStatus>({
    state: 'disconnected',
    deviceName: '',
    host: '',
    port: 7383,
  })

  let interval: ReturnType<typeof setInterval> | null = null

  async function fetchStatus() {
    try {
      const result = await window.go.main.App.GetConnectionStatus()
      status.value = result
    } catch (e) {
      console.error('Failed to fetch status:', e)
    }
  }

  onMounted(() => {
    fetchStatus()
    interval = setInterval(fetchStatus, 1000)
  })

  onUnmounted(() => {
    if (interval) clearInterval(interval)
  })

  return { status }
}
