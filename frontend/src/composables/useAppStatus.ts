import { ref, computed } from 'vue'
import type { AppStatus } from '../types'
import {
  GetStatus,
  Connect,
  Disconnect,
  PatchMode,
  SetSystemProxy,
} from '../../wailsjs/go/main/App'

const status = ref<AppStatus>({
  running: false,
  configPath: '',
  controller: '',
  mixedPort: 7890,
  systemProxyEnabled: false,
  systemProxyServer: '',
  connected: false,
  version: '',
})

const loading = ref(false)

export function useAppStatus() {
  const refresh = async () => {
    status.value = await GetStatus()
  }

  const connect = async () => {
    loading.value = true
    try {
      await Connect()
      await refresh()
    } finally {
      loading.value = false
    }
  }

  const disconnect = async () => {
    loading.value = true
    try {
      await Disconnect()
      await refresh()
    } finally {
      loading.value = false
    }
  }

  const toggleSystemProxy = async (enabled: boolean) => {
    await SetSystemProxy(enabled)
    await refresh()
  }

  const setMode = async (mode: string) => {
    await PatchMode(mode)
    await refresh()
  }

  const running = computed(() => status.value.running)
  const connected = computed(() => status.value.connected)

  return {
    status,
    loading,
    running,
    connected,
    refresh,
    connect,
    disconnect,
    toggleSystemProxy,
    setMode,
  }
}

export function formatBytes(n: number) {
  if (!n) return '0 B'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  if (n < 1024 * 1024 * 1024) return `${(n / 1024 / 1024).toFixed(1)} MB`
  return `${(n / 1024 / 1024 / 1024).toFixed(2)} GB`
}

export function formatSpeed(n: number) {
  return `${formatBytes(n)}/s`
}
