<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NButton, NSpace, useMessage } from 'naive-ui'
import { GetAppInfo, CheckForUpdates, OpenWebsite, OpenHelpDocs } from '../../wailsjs/go/main/App'

const message = useMessage()
const checking = ref(false)
const info = ref({
  name: '绿橙',
  version: '-',
  websiteUrl: '',
  helpDocsUrl: '',
})

const loadInfo = async () => {
  info.value = await GetAppInfo()
}

const checkUpdate = async () => {
  checking.value = true
  try {
    const result = await CheckForUpdates()
    message.success(result.message || '当前已是最新版本')
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    checking.value = false
  }
}

const openWebsite = () => {
  OpenWebsite()
}

const openHelpDocs = () => {
  OpenHelpDocs()
}

onMounted(loadInfo)
</script>

<template>
  <div class="page page-narrow">
    <n-card title="关于" :bordered="false">
      <div class="about-head">
        <div class="logo">绿橙</div>
        <div>
          <h2>{{ info.name }}</h2>
          <p class="version">版本 {{ info.version }}</p>
        </div>
      </div>

      <n-space vertical :size="12" class="actions">
        <n-button block :loading="checking" @click="checkUpdate">检查更新</n-button>
        <n-button block @click="openWebsite">官网</n-button>
        <n-button block @click="openHelpDocs">帮助文档</n-button>
      </n-space>
    </n-card>
  </div>
</template>

<style scoped>
.about-head {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.logo {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #18a058, #2080f0);
}
.about-head h2 {
  margin: 0 0 4px;
  font-size: 20px;
  font-weight: 600;
}
.version {
  margin: 0;
  font-size: 13px;
  color: var(--n-text-color-3);
}
.actions {
  max-width: 320px;
}
</style>
