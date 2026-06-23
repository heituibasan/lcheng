<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { NCard, NButton, NEmpty, NScrollbar } from 'naive-ui'
import { useAppStatus } from '../composables/useAppStatus'
import { GetLogs } from '../../wailsjs/go/main/App'

const { running, refresh } = useAppStatus()
const logs = ref<string[]>([])
const autoScroll = ref(true)

const load = async () => {
  if (!running.value) { logs.value = []; return }
  logs.value = await GetLogs()
}

let timer: number
onMounted(async () => {
  await refresh()
  await load()
  timer = window.setInterval(load, 1500)
})
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <div class="page">
    <n-card title="运行日志" :bordered="false">
      <template #header-extra>
        <n-button size="small" @click="load">刷新</n-button>
      </template>
      <n-empty v-if="!running" description="核心未运行" />
      <n-scrollbar v-else class="log-scroll">
        <pre class="log-box">{{ logs.join('\n') || '暂无日志...' }}</pre>
      </n-scrollbar>
    </n-card>
  </div>
</template>

<style scoped>
.log-scroll {
  max-height: min(600px, calc(100dvh - 180px));
}
.log-box {
  margin: 0;
  padding: 12px;
  font-family: Consolas, 'Cascadia Code', monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
