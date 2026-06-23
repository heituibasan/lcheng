<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  NCard, NButton, NSpace, NModal, NInput, NEmpty, useMessage, useDialog,
} from 'naive-ui'
import {
  ListProfiles, CreateProfile, DeleteProfile, ActivateProfile,
} from '../../wailsjs/go/main/App'
import type { ProfileItem } from '../types'

const message = useMessage()
const dialog = useDialog()
const items = ref<ProfileItem[]>([])
const showAdd = ref(false)
const newName = ref('')

const load = async () => {
  items.value = await ListProfiles()
}

const create = async () => {
  if (!newName.value.trim()) return
  try {
    await CreateProfile(newName.value.trim())
    showAdd.value = false
    newName.value = ''
    message.success('档案已创建')
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const activate = async (id: string) => {
  try {
    await ActivateProfile(id)
    message.success('已切换档案')
    await load()
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const remove = (item: ProfileItem) => {
  dialog.warning({
    title: '删除档案',
    content: `确定删除「${item.name}」？`,
    positiveText: '删除',
    onPositiveClick: async () => {
      await DeleteProfile(item.id)
      message.success('已删除')
      await load()
    },
  })
}

onMounted(load)
</script>

<template>
  <div class="page">
    <n-card title="配置档案" :bordered="false">
      <template #header-extra>
        <n-button type="primary" @click="showAdd = true">新建档案</n-button>
      </template>

      <n-empty v-if="items.length === 0" description="暂无配置档案" />

      <n-space v-else vertical>
        <n-card v-for="item in items" :key="item.id" size="small">
          <n-space justify="space-between" align="center">
            <div>
              <strong>{{ item.name }}</strong>
              <p class="meta">更新于 {{ new Date(item.updatedAt * 1000).toLocaleString() }}</p>
            </div>
            <n-space>
              <n-button size="small" type="primary" @click="activate(item.id)">启用</n-button>
              <n-button v-if="item.id !== 'default'" size="small" @click="remove(item)">删除</n-button>
            </n-space>
          </n-space>
        </n-card>
      </n-space>
    </n-card>

    <n-modal v-model:show="showAdd" preset="dialog" title="新建档案" positive-text="创建" @positive-click="create">
      <n-input v-model:value="newName" placeholder="档案名称" />
    </n-modal>
  </div>
</template>

<style scoped>
.page { max-width: 700px; }
.meta { margin: 4px 0 0; font-size: 12px; color: var(--n-text-color-3); }
</style>
