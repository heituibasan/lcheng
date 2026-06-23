import { ref, computed, watch, onMounted, provide, inject, type InjectionKey, type ComputedRef, type Ref } from 'vue'
import { darkTheme, type GlobalTheme } from 'naive-ui'
import { GetSettings, SaveSettings } from '../../wailsjs/go/main/App'
import {
  WindowSetDarkTheme,
  WindowSetLightTheme,
  WindowSetSystemDefaultTheme,
} from '../../wailsjs/runtime/runtime'
import type { ThemeMode } from '../types'

export interface AppThemeContext {
  themeMode: Ref<ThemeMode>
  isDark: ComputedRef<boolean>
  theme: ComputedRef<GlobalTheme | null>
  setThemeMode: (mode: ThemeMode) => Promise<void>
  toggleLightDark: () => Promise<void>
  loadTheme: () => Promise<void>
}

const themeKey: InjectionKey<AppThemeContext> = Symbol('appTheme')

const prefersDark = window.matchMedia('(prefers-color-scheme: dark)')

function syncNativeWindowTheme(mode: ThemeMode, dark: boolean) {
  try {
    if (mode === 'system') {
      WindowSetSystemDefaultTheme()
      return
    }
    if (dark) {
      WindowSetDarkTheme()
    } else {
      WindowSetLightTheme()
    }
  } catch {
    /* browser dev */
  }
}

export function provideAppTheme(): AppThemeContext {
  const themeMode = ref<ThemeMode>('system')
  const systemDark = ref(prefersDark.matches)

  prefersDark.addEventListener('change', (event) => {
    systemDark.value = event.matches
  })

  const isDark = computed(() => {
    if (themeMode.value === 'dark') return true
    if (themeMode.value === 'light') return false
    return systemDark.value
  })

  const theme = computed<GlobalTheme | null>(() => (isDark.value ? darkTheme : null))

  const loadTheme = async () => {
    try {
      const settings = await GetSettings()
      const mode = settings.themeMode
      if (mode === 'light' || mode === 'dark' || mode === 'system') {
        themeMode.value = mode
      }
    } catch {
      /* ignore */
    }
    syncNativeWindowTheme(themeMode.value, isDark.value)
  }

  const setThemeMode = async (mode: ThemeMode) => {
    themeMode.value = mode
    syncNativeWindowTheme(mode, isDark.value)
    try {
      const settings = await GetSettings()
      settings.themeMode = mode
      await SaveSettings(settings)
    } catch {
      /* ignore */
    }
  }

  const toggleLightDark = async () => {
    await setThemeMode(isDark.value ? 'light' : 'dark')
  }

  watch(isDark, (dark) => {
    syncNativeWindowTheme(themeMode.value, dark)
  })

  onMounted(loadTheme)

  const context: AppThemeContext = {
    themeMode,
    isDark,
    theme,
    setThemeMode,
    toggleLightDark,
    loadTheme,
  }

  provide(themeKey, context)
  return context
}

export function useAppTheme(): AppThemeContext {
  const context = inject(themeKey)
  if (!context) {
    throw new Error('useAppTheme must be used within provideAppTheme')
  }
  return context
}
