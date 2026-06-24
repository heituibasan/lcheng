<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import {
  NCard, NGrid, NGi, NStatistic, NButton, NSpace, NTag, NSelect, NSwitch,
  NInputNumber, useMessage,
} from 'naive-ui'
import { useAppStatus, formatBytes, formatSpeed } from '../composables/useAppStatus'
import { useAppTheme } from '../composables/useAppTheme'
import type { ThemeMode } from '../types'
import {
  GetConfigs, GetTraffic, GetSettings, SaveSettings,
  SetTunEnabled, SetAllowLan, SetLaunchAtLogin,
} from '../../wailsjs/go/main/App'
import type { AppSettings } from '../types'

const message = useMessage()
const { status, loading, running, connected, refresh, connect, disconnect, setMode } = useAppStatus()
const { themeMode, setThemeMode } = useAppTheme()

const mode = ref('rule')
const trafficUp = ref(0)
const trafficDown = ref(0)
const trafficUpTotal = ref(0)
const trafficDownTotal = ref(0)

const form = ref<AppSettings>({
  autoStartCore: false,
  autoSystemProxy: true,
  systemProxyEnabled: false,
  mixedPort: 7890,
  tunEnabled: false,
  allowLan: false,
  logLevel: 'info',
  subscriptionUserAgent: 'clash-verge/v1.0',
  activeProfileId: 'default',
  proxySelections: {},
  startMinimized: false,
  launchAtLogin: false,
  themeMode: 'system',
})

const modeOptions = [
  { label: '规则模式', value: 'rule' },
  { label: '全局模式', value: 'global' },
  { label: '直连模式', value: 'direct' },
]

const logLevels = [
  { label: '静默', value: 'silent' },
  { label: '错误', value: 'error' },
  { label: '警告', value: 'warning' },
  { label: '信息', value: 'info' },
  { label: '调试', value: 'debug' },
]

const themeOptions = [
  { label: '跟随系统', value: 'system' },
  { label: '浅色', value: 'light' },
  { label: '深色', value: 'dark' },
]

let portTimer: number

const applySettings = async () => {
  try {
    await SaveSettings(form.value)
    await refresh()
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const loadExtra = async () => {
  const settings = await GetSettings()
  form.value = {
    ...settings,
    themeMode: (settings.themeMode === 'light' || settings.themeMode === 'dark' || settings.themeMode === 'system')
      ? settings.themeMode
      : 'system',
  }
  themeMode.value = form.value.themeMode

  if (!running.value) return
  try {
    const configs = await GetConfigs()
    mode.value = (configs.mode as string) || 'rule'
    const t = await GetTraffic()
    trafficUp.value = t.upload || 0
    trafficDown.value = t.download || 0
    trafficUpTotal.value = t.upTotal || 0
    trafficDownTotal.value = t.downTotal || 0
  } catch { /* ignore */ }
}

const toggleProxy = async () => {
  try {
    if (connected.value) {
      await disconnect()
      message.success('已断开代理')
    } else {
      await connect()
      message.success('代理已连接')
    }
    await loadExtra()
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const onModeChange = async (value: string) => {
  try {
    await setMode(value)
  } catch (e: unknown) {
    message.error(String(e))
    await loadExtra()
  }
}

const onAutoStart = async (v: boolean) => {
  form.value.autoStartCore = v
  await applySettings()
}

const onAutoProxy = async (v: boolean) => {
  form.value.autoSystemProxy = v
  await applySettings()
}

const onLaunchAtLogin = async (v: boolean) => {
  form.value.launchAtLogin = v
  try {
    await SetLaunchAtLogin(v)
    message.success(v ? '已开启开机启动' : '已关闭开机启动')
  } catch (e: unknown) {
    message.error(String(e))
    form.value.launchAtLogin = !v
  }
}

const onPortChange = (v: number | null) => {
  form.value.mixedPort = v || 7890
  clearTimeout(portTimer)
  portTimer = window.setTimeout(applySettings, 400)
}

const onTun = async (v: boolean) => {
  form.value.tunEnabled = v
  try {
    await SetTunEnabled(v)
  } catch (e: unknown) {
    message.error(String(e))
    form.value.tunEnabled = !v
  }
}

const onAllowLan = async (v: boolean) => {
  form.value.allowLan = v
  try {
    await SetAllowLan(v)
  } catch (e: unknown) {
    message.error(String(e))
    form.value.allowLan = !v
  }
}

const onLogLevel = async (v: string) => {
  form.value.logLevel = v
  await applySettings()
}

const onThemeMode = async (v: ThemeMode) => {
  form.value.themeMode = v
  await setThemeMode(v)
}

let timer: number
onMounted(async () => {
  await refresh()
  await loadExtra()
  timer = window.setInterval(loadExtra, 2000)
})
onUnmounted(() => {
  clearInterval(timer)
  clearTimeout(portTimer)
})
</script>

<template>
  <div class="page page-narrow">
    <!-- 连接区 -->
    <n-card :bordered="false" class="hero-card">
      <div class="hero">
        <div>
          <n-tag :type="connected ? 'success' : 'default'" round size="large">
            {{ connected ? '代理已连接' : '代理未连接' }}
          </n-tag>
          <h2>{{ connected ? '流量已通过代理' : '点击连接开始代理' }}</h2>
          <p v-if="status.version" class="sub">{{ status.version }} · 端口 {{ status.mixedPort }}</p>
        </div>
        <n-button
          :type="connected ? 'error' : 'primary'"
          size="large"
          class="connect-btn"
          :loading="loading"
          @click="toggleProxy"
        >
          {{ connected ? '断开' : '一键连接' }}
        </n-button>
      </div>
    </n-card>

    <!-- 流量统计 -->
    <n-grid cols="1 s:2 l:4" responsive="screen" :x-gap="12" :y-gap="12" class="section">
      <n-gi><n-card size="small" :bordered="false"><n-statistic label="上传" :value="formatSpeed(trafficUp)" /></n-card></n-gi>
      <n-gi><n-card size="small" :bordered="false"><n-statistic label="下载" :value="formatSpeed(trafficDown)" /></n-card></n-gi>
      <n-gi><n-card size="small" :bordered="false"><n-statistic label="累计上传" :value="formatBytes(trafficUpTotal)" /></n-card></n-gi>
      <n-gi><n-card size="small" :bordered="false"><n-statistic label="累计下载" :value="formatBytes(trafficDownTotal)" /></n-card></n-gi>
    </n-grid>

    <!-- 控制 + 设置 -->
    <n-grid cols="1 m:2" responsive="screen" :x-gap="16" :y-gap="16" class="section">
      <n-gi>
        <n-card title="连接控制" size="small" :bordered="false">
          <n-space vertical :size="16">
            <div class="setting-row">
              <span>出站模式</span>
              <n-select
                v-model:value="mode"
                :options="modeOptions"
                :disabled="!running"
                style="width: 140px"
                size="small"
                @update:value="onModeChange"
              />
            </div>
            <div class="setting-row">
              <span>核心状态</span>
              <n-tag :type="running ? 'success' : 'default'" size="small">{{ running ? '运行中' : '已停止' }}</n-tag>
            </div>
          </n-space>
        </n-card>
      </n-gi>

      <n-gi>
        <n-card title="高级选项" size="small" :bordered="false">
          <n-space vertical :size="16">
            <div class="setting-row">
              <span>TUN 模式</span>
              <n-switch :value="form.tunEnabled" @update:value="onTun" />
            </div>
            <div class="setting-row">
              <span>允许局域网</span>
              <n-switch :value="form.allowLan" @update:value="onAllowLan" />
            </div>
            <div class="setting-row">
              <span>日志级别</span>
              <n-select
                :value="form.logLevel"
                :options="logLevels"
                size="small"
                style="width: 120px"
                @update:value="onLogLevel"
              />
            </div>
          </n-space>
        </n-card>
      </n-gi>
    </n-grid>

    <n-card title="启动选项" size="small" :bordered="false" class="section startup-card">
      <n-space vertical :size="16">
        <div class="setting-row">
          <span>主题</span>
          <n-select
            :value="themeMode"
            :options="themeOptions"
            size="small"
            style="width: 140px"
            @update:value="onThemeMode"
          />
        </div>
        <div class="setting-row">
          <span>开机启动</span>
          <n-switch :value="form.launchAtLogin" @update:value="onLaunchAtLogin" />
        </div>
        <div class="setting-row">
          <span>启动时自动连接</span>
          <n-switch :value="form.autoStartCore" @update:value="onAutoStart" />
        </div>
        <div class="setting-row">
          <span>连接时启用系统代理</span>
          <n-switch :value="form.autoSystemProxy" @update:value="onAutoProxy" />
        </div>
        <div class="setting-row">
          <span>混合端口</span>
          <n-input-number
            :value="form.mixedPort"
            :min="1024"
            :max="65535"
            size="small"
            style="width: 120px"
            @update:value="onPortChange"
          />
        </div>
      </n-space>
    </n-card>
  </div>
</template>

<style scoped>
.section { margin-top: 16px; }
.startup-card { max-width: 100%; }
.hero-card {
  background: linear-gradient(135deg, rgba(24,160,88,0.1), rgba(32,128,240,0.06));
}
.hero {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}
.hero h2 { margin: 10px 0 4px; font-size: 20px; font-weight: 600; }
.sub { margin: 0; font-size: 12px; color: var(--n-text-color-3); }
.connect-btn { min-width: 120px; height: 44px; }
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
  flex-wrap: wrap;
}

@media (max-width: 640px) {
  .hero {
    flex-direction: column;
    align-items: stretch;
  }

  .connect-btn {
    width: 100%;
  }
}
</style>
