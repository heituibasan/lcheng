export interface AppStatus {
  running: boolean
  configPath: string
  controller: string
  mixedPort: number
  systemProxyEnabled: boolean
  systemProxyServer: string
  connected: boolean
  version: string
}

export interface ProxyGroup {
  name: string
  type: string
  now: string
  all: string[]
}

export interface ConnectionItem {
  id: string
  metadata: {
    host: string
    network: string
    type: string
    sourceIP: string
    destinationIP: string
    destinationPort: string
  }
  upload: number
  download: number
  start: string
  chains: string[]
  rule: string
}

export interface SubscriptionItem {
  id: string
  name: string
  url: string
  updatedAt: number
  trafficUsed?: string
  trafficTotal?: string
  expireAt?: string
}

export interface ProfileItem {
  id: string
  name: string
  filename: string
  updatedAt: number
}

export type ThemeMode = 'system' | 'light' | 'dark'

export interface AppSettings {
  autoStartCore: boolean
  autoSystemProxy: boolean
  systemProxyEnabled: boolean
  mixedPort: number
  tunEnabled: boolean
  allowLan: boolean
  logLevel: string
  subscriptionUserAgent: string
  activeProfileId: string
  proxySelections: Record<string, Record<string, string>>
  startMinimized: boolean
  launchAtLogin: boolean
  themeMode: ThemeMode
}

export interface RuleItem {
  type: string
  payload: string
  proxy: string
}
