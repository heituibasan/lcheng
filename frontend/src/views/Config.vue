<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NCard, NInput, NButton, NSpace, useMessage } from 'naive-ui'
import {
  GetConfig, SaveConfig, TestConfig, ImportConfigFromFile, ExportConfigToFile,
} from '../../wailsjs/go/main/App'

const message = useMessage()
const content = ref('')
const saving = ref(false)
const testing = ref(false)

const load = async () => { content.value = await GetConfig() }

const save = async () => {
  saving.value = true
  try {
    await SaveConfig(content.value)
    message.success('配置已保存并热重载')
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    saving.value = false
  }
}

const test = async () => {
  testing.value = true
  try {
    await TestConfig(content.value)
    message.success('配置校验通过')
  } catch (e: unknown) {
    message.error(String(e))
  } finally {
    testing.value = false
  }
}

const importConfig = async () => {
  try {
    const result = await ImportConfigFromFile()
    if (!result) return
    content.value = result
    message.success('配置已导入')
  } catch (e: unknown) {
    message.error(String(e))
  }
}

const exportConfig = async () => {
  try {
    await ExportConfigToFile()
    message.success('配置已导出')
  } catch (e: unknown) {
    message.error(String(e))
  }
}

onMounted(load)
</script>

<template>
  <div class="page">
    <n-card title="YAML 配置" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-button @click="importConfig">导入</n-button>
          <n-button @click="exportConfig">导出</n-button>
          <n-button :loading="testing" @click="test">校验</n-button>
          <n-button type="primary" :loading="saving" @click="save">保存</n-button>
        </n-space>
      </template>
      <n-input v-model:value="content" type="textarea" :rows="26" class="editor" />
      <p class="hint">完全兼容 Clash / Mihomo YAML。修改后点击保存即可热重载，无需重启核心。</p>
    </n-card>
  </div>
</template>

<style scoped>
.editor { font-family: Consolas, 'Cascadia Code', monospace; font-size: 13px; }
.hint { margin-top: 12px; font-size: 12px; color: var(--n-text-color-3); }
</style>
