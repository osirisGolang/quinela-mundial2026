<template>
  <q-layout view="hHh lpR fFf">
    <q-header elevated class="bg-primary text-white">
      <q-toolbar>
        <q-btn flat dense round icon="menu" @click="drawer = !drawer" />
        <q-toolbar-title class="text-weight-bold">
          <q-icon name="sports_soccer" class="q-mr-xs" />
          Quinela Mundial 2026
        </q-toolbar-title>
        <div v-if="auth.user" class="row items-center q-gutter-sm">
          <q-badge color="amber-10" text-color="black" :label="auth.user.username" />
          <q-btn flat dense round icon="logout" @click="logout">
            <q-tooltip>Cerrar sesión</q-tooltip>
          </q-btn>
        </div>
        <div v-else class="row q-gutter-sm">
          <q-btn flat dense to="/login" label="Iniciar Sesión" icon="login" />
          <q-btn unelevated color="white" text-color="primary" to="/register" label="Registrarse" />
        </div>
      </q-toolbar>
    </q-header>

    <q-drawer v-model="drawer" show-if-above bordered>
      <q-list padding>
        <q-item-label header class="text-primary text-weight-bold">
          <q-icon name="emoji_events" class="q-mr-xs" />
          Navegación
        </q-item-label>

        <q-item clickable v-ripple to="/" exact>
          <q-item-section avatar>
            <q-icon name="home" />
          </q-item-section>
          <q-item-section>Inicio</q-item-section>
        </q-item>

        <q-item clickable v-ripple to="/groups">
          <q-item-section avatar>
            <q-icon name="groups" />
          </q-item-section>
          <q-item-section>Grupos</q-item-section>
        </q-item>

        <q-item clickable v-ripple to="/schedule">
          <q-item-section avatar>
            <q-icon name="calendar_today" />
          </q-item-section>
          <q-item-section>Calendario</q-item-section>
        </q-item>

        <q-separator class="q-my-sm" />

        <q-item-label header class="text-grey-7">Quiniela</q-item-label>

        <q-item clickable v-ripple to="/standings">
          <q-item-section avatar>
            <q-icon name="leaderboard" />
          </q-item-section>
          <q-item-section>Tabla General</q-item-section>
        </q-item>

        <q-item v-if="auth.user" clickable v-ripple to="/predictions">
          <q-item-section avatar>
            <q-icon name="how_to_vote" />
          </q-item-section>
          <q-item-section>Mis Pronósticos</q-item-section>
        </q-item>

        <q-separator class="q-my-sm" />

        <q-item-label header class="text-grey-7">Admin</q-item-label>

        <q-item v-if="auth.isAdmin" clickable v-ripple to="/results">
          <q-item-section avatar>
            <q-icon name="scoreboard" />
          </q-item-section>
          <q-item-section>Ingresar Resultados</q-item-section>
        </q-item>

        <q-item v-if="auth.isAdmin" clickable v-ripple to="/users">
          <q-item-section avatar>
            <q-icon name="manage_accounts" color="amber-8" />
          </q-item-section>
          <q-item-section class="text-weight-medium">Gestión de Usuarios</q-item-section>
        </q-item>
        <q-item v-if="auth.isAdmin" clickable v-ripple to="/admin-predictions">
          <q-item-section avatar><q-icon name="table_view" color="primary" /></q-item-section>
          <q-item-section class="text-weight-medium">Pronósticos de Usuarios</q-item-section>
          <q-item-section side>
            <q-badge color="amber-8" label="Admin" text-color="black" />
          </q-item-section>
        </q-item>

        <q-separator class="q-my-sm" />

        <q-item v-if="!auth.user" clickable v-ripple to="/login">
          <q-item-section avatar>
            <q-icon name="login" />
          </q-item-section>
          <q-item-section>Iniciar Sesión</q-item-section>
        </q-item>

        <q-item v-if="auth.user" clickable v-ripple @click="logout">
          <q-item-section avatar>
            <q-icon name="logout" color="negative" />
          </q-item-section>
          <q-item-section class="text-negative">Cerrar Sesión</q-item-section>
        </q-item>
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const drawer = ref(false)
const router = useRouter()
const auth = useAuthStore()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>
