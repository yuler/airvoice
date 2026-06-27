/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface WailsApp {
  GetQRCode(): Promise<string>
  GetConnectionStatus(): Promise<string>
}

interface Window {
  go: {
    main: {
      App: WailsApp
    }
  }
}
