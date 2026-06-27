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

interface WailsApp {
  GetQRCode(): Promise<string>
  GetConnectionStatus(): Promise<ConnectionStatus>
}

interface Window {
  go: {
    main: {
      App: WailsApp
    }
  }
}
