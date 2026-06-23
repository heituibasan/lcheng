import { createApp } from 'vue'
import naive from 'naive-ui'
import { loader } from '@guolao/vue-monaco-editor'
import * as monaco from 'monaco-editor'
import App from './App.vue'
import router from './router'
import './style.css'

loader.config({ monaco })

const editableSelector = 'input, textarea, [contenteditable="true"], .monaco-editor, .allow-text-select'

const blockBrowserContextMenu = (event: MouseEvent) => {
  const target = event.target
  if (target instanceof Element && target.closest(editableSelector)) {
    return
  }
  event.preventDefault()
}

const blockTextSelection = (event: Event) => {
  const target = event.target
  if (!(target instanceof Element)) {
    event.preventDefault()
    return
  }
  if (target.closest(editableSelector)) {
    return
  }
  event.preventDefault()
}

document.addEventListener('contextmenu', blockBrowserContextMenu, { capture: true })
document.addEventListener('selectstart', blockTextSelection, { capture: true })

const app = createApp(App)
app.use(naive)
app.use(router)
app.mount('#app')
