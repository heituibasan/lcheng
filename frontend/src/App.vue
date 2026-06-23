<script setup lang="ts">
import { computed, h, ref, onMounted, onUnmounted } from 'vue'
import { RouterView, useRouter, useRoute } from 'vue-router'
import {
  NConfigProvider,
  NLayout,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NIcon,
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
  darkTheme,
  type MenuOption,
} from 'naive-ui'
import {
  SpeedometerOutline,
  GitNetworkOutline,
  PulseOutline,
  CloudDownloadOutline,
  ListOutline,
  TerminalOutline,
  InformationCircleOutline,
} from '@vicons/ionicons5'

const router = useRouter()
const route = useRoute()

const SIDEBAR_BREAKPOINT = 960
const collapsed = ref(window.innerWidth < SIDEBAR_BREAKPOINT)

const prefersDark = window.matchMedia('(prefers-color-scheme: dark)')
const isDark = ref(prefersDark.matches)
prefersDark.addEventListener('change', (e) => { isDark.value = e.matches })

const theme = computed(() => (isDark.value ? darkTheme : null))

const menuOptions: MenuOption[] = [
  { label: '控制面板', key: 'dashboard', icon: () => h(NIcon, null, { default: () => h(SpeedometerOutline) }) },
  { label: '代理节点', key: 'proxies', icon: () => h(NIcon, null, { default: () => h(GitNetworkOutline) }) },
  { label: '代理与配置', key: 'subscriptions', icon: () => h(NIcon, null, { default: () => h(CloudDownloadOutline) }) },
  { label: '分流规则', key: 'rules', icon: () => h(NIcon, null, { default: () => h(ListOutline) }) },
  { label: '活动连接', key: 'connections', icon: () => h(NIcon, null, { default: () => h(PulseOutline) }) },
  { label: '运行日志', key: 'logs', icon: () => h(NIcon, null, { default: () => h(TerminalOutline) }) },
  { label: '关于', key: 'about', icon: () => h(NIcon, null, { default: () => h(InformationCircleOutline) }) },
]

const activeKey = computed(() => route.name as string)
const onMenuUpdate = (key: string) => router.push({ name: key })

const handleResize = () => {
  if (window.innerWidth < SIDEBAR_BREAKPOINT) {
    collapsed.value = true
  }
}

onMounted(() => window.addEventListener('resize', handleResize))
onUnmounted(() => window.removeEventListener('resize', handleResize))
</script>

<template>
  <n-config-provider class="root-provider" :theme="theme">
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <n-layout has-sider class="app-layout">
            <n-layout-sider
              bordered
              collapse-mode="width"
              :collapsed-width="64"
              :width="220"
              :collapsed="collapsed"
              show-trigger
              @collapse="collapsed = true"
              @expand="collapsed = false"
            >
              <div class="brand">
                <span v-if="!collapsed">绿橙</span>
                <span v-else>绿</span>
              </div>
              <n-menu
                :value="activeKey"
                :collapsed="collapsed"
                :collapsed-width="64"
                :collapsed-icon-size="22"
                :options="menuOptions"
                @update:value="onMenuUpdate"
              />
            </n-layout-sider>

            <n-layout class="main-panel">
              <n-layout-content class="content" :native-scrollbar="false">
                <RouterView />
              </n-layout-content>
            </n-layout>
          </n-layout>
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.root-provider {
  flex: 1;
  min-width: 0;
  min-height: 0;
  width: 100%;
  display: flex;
  flex-direction: column;
}

.app-layout {
  flex: 1;
  width: 100%;
  min-width: 0;
  min-height: 0;
  height: 100%;
}

.main-panel {
  flex: 1;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.brand {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 18px;
}

.content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: auto;
  padding: 16px;
}

@media (min-width: 768px) {
  .content {
    padding: 20px 24px;
  }
}
</style>
