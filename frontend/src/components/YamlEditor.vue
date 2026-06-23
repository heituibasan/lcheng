<script setup lang="ts">
import { ref, shallowRef, watch, computed } from 'vue'
import { VueMonacoEditor } from '@guolao/vue-monaco-editor'
import type { editor } from 'monaco-editor'

const props = withDefaults(defineProps<{
  modelValue: string
  dark?: boolean
  height?: string
}>(), {
  height: '520px',
})

const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const editorRef = shallowRef<editor.IStandaloneCodeEditor>()
const content = ref(props.modelValue)

watch(() => props.modelValue, (v) => {
  if (v !== content.value) content.value = v
})

watch(content, (v) => emit('update:modelValue', v))

const theme = computed(() => (props.dark ? 'vs-dark' : 'vs'))

const options: editor.IStandaloneEditorConstructionOptions = {
  minimap: { enabled: false },
  fontSize: 13,
  lineNumbers: 'on',
  scrollBeyondLastLine: false,
  wordWrap: 'on',
  tabSize: 2,
  automaticLayout: true,
}

const onMount = (instance: editor.IStandaloneCodeEditor) => {
  editorRef.value = instance
}
</script>

<template>
  <div class="yaml-editor">
    <VueMonacoEditor
      v-model:value="content"
      language="yaml"
      :theme="theme"
      :height="height"
      :options="options"
      @mount="onMount"
    />
  </div>
</template>

<style scoped>
.yaml-editor {
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
  overflow: hidden;
}
</style>
