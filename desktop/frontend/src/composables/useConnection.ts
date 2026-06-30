import { ref, onMounted } from 'vue'

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
    const runtime = (window as any).runtime
    if (runtime && runtime.EventsOn) {
      runtime.EventsOn('status_changed', (newStatus: ConnectionStatus) => {
        status.value = newStatus
      })
    }
  })

  return { status }
}
