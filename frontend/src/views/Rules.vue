<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  NCard, NDataTable, NEmpty, NInput, NButton, NSpace, useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { GetConfigRulesPage } from '../../wailsjs/go/main/App'

interface RuleItem {
  type: string
  payload: string
  proxy: string
}

const message = useMessage()
const rules = ref<RuleItem[]>([])
const keyword = ref('')
const debouncedKeyword = ref('')
const loading = ref(false)
const loaded = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 100

let debounceTimer: number

watch(keyword, (v) => {
  clearTimeout(debounceTimer)
  debounceTimer = window.setTimeout(() => {
    debouncedKeyword.value = v
    page.value = 1
    if (loaded.value) load()
  }, 300)
})

const pageCount = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

const columns: DataTableColumns<RuleItem> = [
  { title: '类型', key: 'type', width: 140 },
  { title: '匹配项', key: 'payload', ellipsis: { tooltip: true } },
  { title: '策略', key: 'proxy', width: 160 },
]

const load = async () => {
  loading.value = true
  try {
    const offset = (page.value - 1) * pageSize
    const data = await GetConfigRulesPage(offset, pageSize, debouncedKeyword.value)
    rules.value = (data.rules as RuleItem[]) || []
    total.value = data.total || 0
    loaded.value = true
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    loading.value = false
  }
}

watch(page, () => {
  if (loaded.value) load()
})
</script>

<template>
  <div class="page">
    <n-card title="分流规则" :bordered="false">
      <template #header-extra>
        <div class="page-toolbar">
          <n-input v-model:value="keyword" placeholder="搜索规则" clearable class="toolbar-input" />
          <n-button type="primary" :loading="loading" @click="load">
            {{ loaded ? '刷新' : '加载规则' }}
          </n-button>
        </div>
      </template>

      <n-empty v-if="!loaded" description="规则从当前 YAML 配置读取，点击「加载规则」查看（不会阻塞程序）" />

      <template v-else>
        <p class="summary">共 {{ total }} 条规则，当前显示第 {{ page }} / {{ pageCount }} 页</p>
        <n-data-table
          :columns="columns"
          :data="rules"
          :loading="loading"
          :bordered="false"
          size="small"
          :max-height="520"
          :scroll-x="720"
          virtual-scroll
        />
        <n-space justify="center" style="margin-top: 12px">
          <n-button size="small" :disabled="page <= 1" @click="page--">上一页</n-button>
          <n-button size="small" :disabled="page >= pageCount" @click="page++">下一页</n-button>
        </n-space>
      </template>
    </n-card>
  </div>
</template>

<style scoped>
.summary { margin: 0 0 12px; font-size: 13px; color: var(--n-text-color-3); }
.toolbar-input { width: 220px; max-width: 100%; }
</style>
