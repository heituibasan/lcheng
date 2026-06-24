<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import {
  NCard, NCollapse, NCollapseItem, NButton, NSpace,
  NTag, NEmpty, NInput, useMessage,
} from 'naive-ui'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { useAppStatus } from '../composables/useAppStatus'
import {
  GetProxies, SelectProxy, DelayTestGroup,
  GetConfigProxies, GetActiveConfigID,
} from '../../wailsjs/go/main/App'
import type { ProxyGroup } from '../types'

const message = useMessage()
const { running, refresh } = useAppStatus()
const groups = ref<ProxyGroup[]>([])
const delays = ref<Record<string, number>>({})
const testing = ref(false)
const keyword = ref('')
const proxyCount = ref(0)

const selectableGroups = computed(() => {
  const list = groups.value.filter((g) =>
    ['select', 'url-test', 'fallback', 'load-balance', 'selector', 'urltest', 'loadbalance'].includes(g.type.toLowerCase())
  )
  if (!keyword.value) return list
  const kw = keyword.value.toLowerCase()
  return list.filter((g) => g.name.toLowerCase().includes(kw) || g.all.some((p) => p.toLowerCase().includes(kw)))
})

const parseLiveGroups = (data: Record<string, unknown>) => {
  const proxies = data.proxies as Record<string, Record<string, unknown>>
  if (!proxies) { groups.value = []; return }

  groups.value = Object.entries(proxies)
    .filter(([, item]) => {
      const t = ((item.type as string) || '').toLowerCase()
      return ['selector', 'urltest', 'fallback', 'loadbalance', 'select', 'url-test', 'load-balance'].some((x) => t.includes(x.replace('-', '')))
    })
    .map(([name, item]) => ({
      name,
      type: ((item.type as string) || '').toLowerCase(),
      now: (item.now as string) || '',
      all: (item.all as string[]) || [],
    }))
}

const loadPreview = async () => {
  try {
    const id = await GetActiveConfigID()
    const data = await GetConfigProxies(id)
    proxyCount.value = data.proxyCount || 0
    groups.value = (data.groups as ProxyGroup[]) || []
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const loadProxies = async () => {
  if (running.value) {
    try {
      const data = await GetProxies()
      parseLiveGroups(data as Record<string, unknown>)
      return
    } catch {
      // fallback to preview
    }
  }
  await loadPreview()
}

const selectProxy = async (group: string, proxy: string) => {
  try {
    await SelectProxy(group, proxy)
    const target = groups.value.find((g) => g.name === group)
    if (target) target.now = proxy
    message.success(`已选择 ${group} → ${proxy}`)
    if (running.value) {
      await loadProxies()
    }
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const testGroup = async (group: string) => {
  testing.value = true
  try {
    const res = await DelayTestGroup(group)
    delays.value = { ...delays.value, ...(res as Record<string, number>) }
    message.success(`${group} 延迟测试完成`)
    await refresh()
    await loadProxies()
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    testing.value = false
  }
}

const testAll = async () => {
  testing.value = true
  try {
    for (const group of selectableGroups.value) {
      const res = await DelayTestGroup(group.name)
      delays.value = { ...delays.value, ...(res as Record<string, number>) }
    }
    message.success('全部延迟测试完成')
    await refresh()
    await loadProxies()
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    testing.value = false
  }
}

const delayColor = (delay: number) => {
  if (!delay) return 'default'
  if (delay < 200) return 'success'
  if (delay < 500) return 'warning'
  return 'error'
}

const selectedProxy = (group: ProxyGroup) => group.now || group.all[0] || ''

let offSelected: () => void
let offUpdated: () => void
let offProxySelected: () => void

onMounted(async () => {
  await refresh()
  await loadProxies()
  offSelected = EventsOn('config:selected', () => loadProxies())
  offUpdated = EventsOn('config:updated', () => loadProxies())
  offProxySelected = EventsOn('proxy:selected', () => loadProxies())
})

onUnmounted(() => {
  offSelected?.()
  offUpdated?.()
  offProxySelected?.()
})

watch(running, loadProxies)
</script>

<template>
  <div class="page">
    <n-card title="代理节点" :bordered="false">
      <template #header-extra>
        <div class="page-toolbar">
          <n-input v-model:value="keyword" placeholder="搜索节点/组" clearable class="toolbar-input" />
          <n-button :loading="testing" @click="testAll">全部测速</n-button>
          <n-button @click="loadProxies">刷新</n-button>
        </div>
      </template>

      <n-empty v-if="selectableGroups.length === 0" description="当前配置暂无代理组，请先在「代理与配置」下载或导入配置" />

      <n-collapse v-else>
        <n-collapse-item v-for="group in selectableGroups" :key="group.name" :title="group.name" :name="group.name">
          <template #header-extra>
            <n-space>
              <n-tag size="small" type="info">{{ group.now || group.all[0] || '-' }}</n-tag>
              <n-button size="tiny" :loading="testing" @click.stop="testGroup(group.name)">测速</n-button>
            </n-space>
          </template>
          <div class="proxy-grid">
            <button
              v-for="proxy in group.all"
              :key="proxy"
              type="button"
              class="proxy-item"
              :class="{ active: selectedProxy(group) === proxy }"
              @click="selectProxy(group.name, proxy)"
            >
              <span class="proxy-name">{{ proxy }}</span>
              <n-tag
                v-if="delays[proxy]"
                :type="delayColor(delays[proxy])"
                size="small"
                class="proxy-delay"
              >
                {{ delays[proxy] }} ms
              </n-tag>
            </button>
          </div>
        </n-collapse-item>
      </n-collapse>

      <p v-if="proxyCount > 0" class="summary">共 {{ proxyCount }} 个节点</p>
    </n-card>
  </div>
</template>

<style scoped>
.summary { margin: 12px 0 0; font-size: 12px; color: var(--n-text-color-3); }
.toolbar-input { width: 200px; max-width: 100%; }

.proxy-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(148px, 1fr));
  gap: 10px;
  padding: 4px 0;
}

.proxy-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 6px;
  min-height: 52px;
  padding: 10px 14px;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  background: var(--n-color);
  color: var(--n-text-color);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition: border-color 0.15s, background-color 0.15s, box-shadow 0.15s;
}

.proxy-item:hover {
  border-color: var(--n-color-target);
  background: var(--n-color-hover);
}

.proxy-item.active {
  border-color: var(--n-color-target);
  background: rgba(24, 160, 88, 0.1);
  box-shadow: 0 0 0 1px var(--n-color-target);
}

.proxy-name {
  width: 100%;
  line-height: 1.4;
  word-break: break-all;
}

.proxy-delay {
  flex-shrink: 0;
}
</style>
