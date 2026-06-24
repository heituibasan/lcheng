<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, h } from 'vue'
import {
  NCard, NDataTable, NButton, NSpace, NEmpty, useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useAppStatus, formatBytes } from '../composables/useAppStatus'
import { GetConnections, CloseAllConnections, CloseConnection } from '../../wailsjs/go/main/App'
import type { ConnectionItem } from '../types'

const message = useMessage()
const { running, refresh } = useAppStatus()
const connections = ref<ConnectionItem[]>([])
const uploadTotal = ref(0)
const downloadTotal = ref(0)
const loading = ref(false)

const columns: DataTableColumns<ConnectionItem> = [
  { title: '目标', key: 'host', render: (row) => row.metadata?.host || '-' },
  { title: '类型', key: 'type', width: 80, render: (row) => row.metadata?.type || '-' },
  { title: '上传', key: 'upload', width: 100, render: (row) => formatBytes(row.upload) },
  { title: '下载', key: 'download', width: 100, render: (row) => formatBytes(row.download) },
  { title: '链路', key: 'chains', ellipsis: { tooltip: true }, render: (row) => (row.chains || []).join(' → ') },
  { title: '规则', key: 'rule', ellipsis: { tooltip: true } },
  {
    title: '操作', key: 'action', width: 80,
    render: (row) => h(NButton, { size: 'tiny', onClick: () => closeOne(row.id) }, { default: () => '关闭' }),
  },
]

const load = async () => {
  if (!running.value) { connections.value = []; return }
  loading.value = true
  try {
    const data = await GetConnections()
    connections.value = (data.connections as ConnectionItem[]) || []
    uploadTotal.value = (data.uploadTotal as number) || 0
    downloadTotal.value = (data.downloadTotal as number) || 0
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    loading.value = false
  }
}

const closeAll = async () => {
  try {
    await CloseAllConnections()
    message.success('已关闭全部连接')
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const closeOne = async (id: string) => {
  try {
    await CloseConnection(id)
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  }
}

let timer: number
onMounted(async () => { await refresh(); await load(); timer = window.setInterval(load, 3000) })
onUnmounted(() => clearInterval(timer))
watch(running, load)
</script>

<template>
  <div class="page">
    <n-card title="活动连接" :bordered="false">
      <template #header-extra>
        <div class="page-toolbar">
          <span class="total">总上传 {{ formatBytes(uploadTotal) }} / 总下载 {{ formatBytes(downloadTotal) }}</span>
          <n-button :disabled="!running" @click="load">刷新</n-button>
          <n-button :disabled="!running" type="warning" @click="closeAll">全部关闭</n-button>
        </div>
      </template>
      <n-empty v-if="!running" description="请先连接代理" />
      <n-data-table
        v-else
        :columns="columns"
        :data="connections"
        :loading="loading"
        :bordered="false"
        size="small"
        :max-height="520"
        :scroll-x="900"
      />
    </n-card>
  </div>
</template>

<style scoped>
.total { font-size: 13px; color: var(--n-text-color-3); }
</style>
