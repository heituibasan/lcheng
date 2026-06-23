import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Proxies from '../views/Proxies.vue'
import Connections from '../views/Connections.vue'
import Subscriptions from '../views/Subscriptions.vue'
import Rules from '../views/Rules.vue'
import Logs from '../views/Logs.vue'
import About from '../views/About.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'dashboard', component: Dashboard },
  { path: '/proxies', name: 'proxies', component: Proxies },
  { path: '/subscriptions', name: 'subscriptions', component: Subscriptions },
  { path: '/config', redirect: '/subscriptions' },
  { path: '/profiles', redirect: '/subscriptions' },
  { path: '/settings', redirect: '/' },
  { path: '/rules', name: 'rules', component: Rules },
  { path: '/connections', name: 'connections', component: Connections },
  { path: '/logs', name: 'logs', component: Logs },
  { path: '/about', name: 'about', component: About },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
