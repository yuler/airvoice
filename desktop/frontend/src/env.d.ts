/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ConnectionStatus {
  state: 'disconnected' | 'connecting' | 'connected' | 'waiting' | 'error'
  mode?: 'lan' | 'relay'
  deviceName: string
  host: string
  port: number
  error?: string
}

interface HistoryEntry {
  id: number
  content: string
  device: string
  createdAt: string
}

interface Settings {
  token?: string
  relayKey?: string
  port: number
  autoStart: boolean
  language: string
  connectionMode: 'lan' | 'relay'
  relayUrl: string
  relayAuthKey: string
}

interface WailsApp {
  GetQRCode(): Promise<string>
  GetPairingLink(): Promise<string>
  RefreshPairing(): Promise<void>
  GetConnectionStatus(): Promise<ConnectionStatus>
  DisconnectClients(): Promise<void>
  GetHistory(limit: number): Promise<HistoryEntry[]>
  SearchHistory(query: string, limit: number): Promise<HistoryEntry[]>
  ClearHistory(): Promise<void>
  GetSettings(): Promise<Settings>
  SaveSettings(settings: Settings): Promise<void>
}

interface Window {
  go: {
    main: {
      App: WailsApp
    }
  }
}
