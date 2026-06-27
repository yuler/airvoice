/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ConnectionStatus {
  state: 'disconnected' | 'connecting' | 'connected' | 'waiting'
  deviceName: string
  host: string
  port: number
}

interface HistoryEntry {
  id: number
  content: string
  device: string
  createdAt: string
}

interface WailsApp {
  GetQRCode(): Promise<string>
  GetConnectionStatus(): Promise<ConnectionStatus>
  GetHistory(limit: number): Promise<HistoryEntry[]>
  ClearHistory(): Promise<void>
}

interface Window {
  go: {
    main: {
      App: WailsApp
    }
  }
}
