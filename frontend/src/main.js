import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'

import '@quasar/extras/material-icons/material-icons.css'
import 'quasar/src/css/index.sass'

import { Quasar } from 'quasar'
import App from './App.vue'

import HomeView from './views/HomeView.vue'
import ScheduleView from './views/ScheduleView.vue'
import GroupsView from './views/GroupsView.vue'
import ResultsView from './views/ResultsView.vue'

const routes = [
  { path: '/', component: HomeView },
  { path: '/schedule', component: ScheduleView },
  { path: '/groups', component: GroupsView },
  { path: '/results', component: ResultsView },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

const pinia = createPinia()
const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(Quasar, { plugins: {} })

app.mount('#app')