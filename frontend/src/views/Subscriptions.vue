<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import {
  NCard, NButton, NSpace, NInput, NGrid, NGi, NTag, NDropdown,
  NModal, useMessage, useDialog, NEmpty,
} from 'naive-ui'
import type { DropdownOption } from 'naive-ui'
import {
  ListConfigs, DownloadConfigFromURL, ImportConfigAsNew,
  SelectConfig, DeleteConfig, RefreshConfigFromURL, RefreshAllConfigs,
  GetConfigContent, SaveConfigContent, TestConfig, RenameConfig,
} from '../../wailsjs/go/main/App'
import YamlEditor from '../components/YamlEditor.vue'

interface ConfigEntry {
  id: string
  name: string
  sourceUrl: string
  updatedAt: number
  isActive: boolean
  proxyCount: number
}

const message = useMessage()
const dialog = useDialog()
const urlInput = ref('')
const configs = ref<ConfigEntry[]>([])
const loading = ref(false)
const editorVisible = ref(false)
const editing = ref<ConfigEntry | null>(null)
const editorContent = ref('')
const saving = ref(false)
const testing = ref(false)
const contextTarget = ref<ConfigEntry | null>(null)
const dropdownX = ref(0)
const dropdownY = ref(0)
const showDropdown = ref(false)

const prefersDark = window.matchMedia('(prefers-color-scheme: dark)')
const isDark = ref(prefersDark.matches)
prefersDark.addEventListener('change', (e) => { isDark.value = e.matches })

const hasUrlConfigs = computed(() => configs.value.some((c) => c.sourceUrl))

const formatTime = (ts: number) => (ts ? new Date(ts * 1000).toLocaleString() : '-')

const load = async () => {
  configs.value = await ListConfigs()
}

const downloadFromURL = async () => {
  const url = urlInput.value.trim()
  if (!url) {
    message.warning('请输入配置链接')
    return
  }
  loading.value = true
  try {
    await DownloadConfigFromURL(url)
    urlInput.value = ''
    message.success('配置已下载')
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    loading.value = false
  }
}

const importLocal = async () => {
  loading.value = true
  try {
    const entry = await ImportConfigAsNew()
    if (!entry?.id) return
    message.success('配置已导入')
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    loading.value = false
  }
}

const updateAll = async () => {
  loading.value = true
  try {
    await RefreshAllConfigs()
    message.success('全部配置已更新')
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    loading.value = false
  }
}

const selectConfig = async (item: ConfigEntry) => {
  if (item.isActive) return
  try {
    await SelectConfig(item.id)
    message.success(`已选择「${item.name}」`)
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const openEditor = async (item: ConfigEntry) => {
  editing.value = item
  editorContent.value = await GetConfigContent(item.id)
  editorVisible.value = true
}

const saveEditor = async () => {
  if (!editing.value) return
  saving.value = true
  try {
    await SaveConfigContent(editing.value.id, editorContent.value)
    message.success('配置已保存')
    editorVisible.value = false
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    saving.value = false
  }
}

const testEditor = async () => {
  testing.value = true
  try {
    await TestConfig(editorContent.value)
    message.success('校验通过')
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    testing.value = false
  }
}

const refreshOne = async (item: ConfigEntry) => {
  loading.value = true
  try {
    await RefreshConfigFromURL(item.id)
    message.success('配置已更新')
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    loading.value = false
  }
}

const remove = (item: ConfigEntry) => {
  dialog.warning({
    title: '删除配置',
    content: `确定删除「${item.name}」？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      await DeleteConfig(item.id)
      message.success('已删除')
      await load()
    },
  })
}

const rename = (item: ConfigEntry) => {
  const name = ref(item.name)
  dialog.create({
    title: '重命名',
    content: () => h(NInput, {
      class: 'allow-text-select',
      value: name.value,
      onUpdateValue: (v: string) => { name.value = v },
      placeholder: '配置名称',
    }),
    positiveText: '保存',
    onPositiveClick: async () => {
      await RenameConfig(item.id, name.value.trim() || item.name)
      await load()
    },
  })
}

const menuOptions = computed<DropdownOption[]>(() => {
  const item = contextTarget.value
  if (!item) return []
  const opts: DropdownOption[] = [
    { label: '编辑', key: 'edit' },
    { label: '选择', key: 'select', disabled: item.isActive },
    { label: '重命名', key: 'rename' },
  ]
  if (item.sourceUrl) {
    opts.push({ label: '更新配置', key: 'refresh' })
  }
  if (item.id !== 'default') {
    opts.push({ type: 'divider', key: 'd1' })
    opts.push({ label: '删除', key: 'delete' })
  }
  return opts
})

const onMenuSelect = (key: string) => {
  showDropdown.value = false
  const item = contextTarget.value
  if (!item) return
  switch (key) {
    case 'edit': openEditor(item); break
    case 'select': selectConfig(item); break
    case 'rename': rename(item); break
    case 'refresh': refreshOne(item); break
    case 'delete': remove(item); break
  }
}

const onContextMenu = (e: MouseEvent, item: ConfigEntry) => {
  e.preventDefault()
  e.stopPropagation()
  contextTarget.value = item
  dropdownX.value = e.clientX
  dropdownY.value = e.clientY
  showDropdown.value = true
}

onMounted(load)
</script>

<template>
  <div class="page">
    <n-card title="代理与配置" :bordered="false">
      <n-space vertical :size="20">
        <div class="toolbar">
          <n-input
            v-model:value="urlInput"
            class="url-input allow-text-select"
            placeholder="粘贴配置/订阅链接"
            clearable
            @keyup.enter="downloadFromURL"
          />
          <n-space :size="12">
            <n-button type="primary" :loading="loading" @click="downloadFromURL">下载</n-button>
            <n-button :loading="loading" @click="importLocal">导入</n-button>
            <n-button v-if="hasUrlConfigs" :loading="loading" @click="updateAll">更新配置</n-button>
          </n-space>
        </div>

        <n-empty v-if="configs.length === 0" description="暂无配置" />

        <n-grid v-else cols="1 s:2 m:3" :x-gap="16" :y-gap="16">
          <n-gi v-for="item in configs" :key="item.id">
            <n-card
              size="small"
              hoverable
              class="config-card app-context-menu"
              :class="{ active: item.isActive }"
              @click="selectConfig(item)"
              @contextmenu="onContextMenu($event, item)"
              @dblclick.stop="openEditor(item)"
            >
              <div class="card-head">
                <strong class="name">{{ item.name }}</strong>
                <n-tag v-if="item.isActive" type="success" size="small">当前</n-tag>
              </div>
              <p class="meta">{{ item.proxyCount }} 个节点</p>
              <p class="source">{{ item.sourceUrl || '本地配置' }}</p>
              <p class="time">{{ formatTime(item.updatedAt) }}</p>
              <n-space v-if="item.sourceUrl" size="small" style="margin-top: 8px" @click.stop>
                <n-button size="tiny" :loading="loading" @click="refreshOne(item)">更新</n-button>
              </n-space>
            </n-card>
          </n-gi>
        </n-grid>
      </n-space>
    </n-card>

    <n-dropdown
      :show="showDropdown"
      :options="menuOptions"
      :x="dropdownX"
      :y="dropdownY"
      placement="bottom-start"
      trigger="manual"
      @clickoutside="showDropdown = false"
      @select="onMenuSelect"
    />

    <n-modal
      v-model:show="editorVisible"
      preset="card"
      :title="editing ? `编辑配置 - ${editing.name}` : '编辑配置'"
      class="editor-modal"
      :mask-closable="false"
    >
      <YamlEditor v-model="editorContent" :dark="isDark" height="calc(80vh - 160px)" />
      <template #footer>
        <n-space justify="end">
          <n-button @click="editorVisible = false">取消</n-button>
          <n-button :loading="testing" @click="testEditor">校验</n-button>
          <n-button type="primary" :loading="saving" @click="saveEditor">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}
.url-input { flex: 1 1 280px; min-width: 0; max-width: 100%; }
.config-card { cursor: pointer; transition: border-color 0.2s; }
.config-card.active {
  border: 1px solid var(--n-color-target);
  box-shadow: 0 0 0 1px var(--n-color-target);
}
.card-head { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.name { font-size: 15px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.meta { margin: 8px 0 4px; font-size: 13px; color: var(--n-text-color-2); }
.source {
  margin: 0;
  font-size: 12px;
  color: var(--n-text-color-3);
  word-break: break-all;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.time { margin: 6px 0 0; font-size: 11px; color: var(--n-text-color-3); }
:global(.editor-modal) {
  width: 92vw !important;
  max-width: 1200px !important;
}
</style>
