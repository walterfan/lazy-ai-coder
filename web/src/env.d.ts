/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  readonly VITE_API_PROTOCOL?: string
  readonly VITE_API_HOST?: string
  readonly VITE_API_PORT?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

// Extend Window interface to include Bootstrap
interface Window {
  bootstrap: typeof import('bootstrap');
}
