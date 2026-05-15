import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'

import '@quasar/extras/material-icons/material-icons.css'
import 'quasar/src/css/index.sass'

import { Quasar, Notify, Dialog } from 'quasar'
import App from './App.vue'
import { useAuthStore } from './stores/auth'

import HomeView from './views/HomeView.vue'
import ScheduleView from './views/ScheduleView.vue'
import GroupsView from './views/GroupsView.vue'
import ResultsView from './views/ResultsView.vue'
import StandingsView from './views/StandingsView.vue'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import PredictionsView from './views/PredictionsView.vue'
import UsersView from './views/UsersView.vue'
import AdminPredictionsView from './views/AdminPredictionsView.vue'

const routes = [
  { path: '/', component: HomeView },
  { path: '/schedule', component: ScheduleView },
  { path: '/groups', component: GroupsView },
  { path: '/results', component: ResultsView, meta: { requiresAdmin: true } },
  { path: '/standings', component: StandingsView },
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/predictions', component: PredictionsView, meta: { requiresAuth: true } },
  { path: '/users', component: UsersView, meta: { requiresAdmin: true } },
  { path: '/admin-predictions', component: AdminPredictionsView, meta: { requiresAdmin: true } },
]

const pinia = createPinia()
const router = createRouter({ history: createWebHistory(), routes })

const app = createApp(App)
app.use(pinia)
app.use(router)
app.use(Quasar, {
  plugins: { Notify, Dialog },
  config: { notify: { position: 'top-right', timeout: 3000 } }
})

// Guard registrado DESPUÉS de instalar pinia, para poder usar el store
router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.user) return next('/login')
  if (to.meta.requiresAdmin && !auth.isAdmin) return next('/')
  next()
})

app.mount('#app')
