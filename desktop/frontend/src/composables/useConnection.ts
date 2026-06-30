import { ref, onMounted, onUnmounted } from 'vue'

interface ConnectionStatus {
  state: 'disconnected' | 'connecting' | 'connected' | 'waiting'
  deviceName: string
  host: string
  port: number
}

const status = ref<ConnectionStatus>({
  state: 'disconnected',
  deviceName: '',
  host: '',
  port: 7383,
})

let listenerCount = 0

export function useConnection() {
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
    if (listenerCount === 0) {
      const runtime = (window as any).runtime
      if (runtime && runtime.EventsOn) {
        runtime.EventsOn('status_changed', (newStatus: ConnectionStatus) => {
          status.value = newStatus
        })
      }
    }
    listenerCount++
  })

  onUnmounted(() => {
    listenerCount--
    if (listenerCount <= 0) {
      listenerCount = 0
      const runtime = (window as any).runtime
      if (runtime && runtime.EventsOff) {
        runtime.EventsOff('status_changed')
      }
    }
  })

  async function disconnect() {
    try {
      await (window as any).go.main.App.StopServer()
    } catch (e) {
      console.error('Failed to disconnect:', e)
    }
  }

  return { status, disconnect }
}
