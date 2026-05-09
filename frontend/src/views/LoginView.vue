<template>
  <q-page class="flex flex-center bg-grey-1">
    <q-card style="min-width: 380px;" class="q-pa-lg shadow-2">
      <q-card-section class="text-center q-pb-none">
        <q-icon name="sports_soccer" size="3em" color="primary" />
        <div class="text-h5 text-primary text-weight-bold q-mt-sm">Quinela Mundial 2026</div>
        <div class="text-subtitle2 text-grey-6">Inicia sesión para continuar</div>
      </q-card-section>

      <q-card-section>
        <q-form @submit.prevent="onSubmit" class="q-gutter-md">
          <q-input
            v-model="username"
            label="Usuario"
            outlined
            dense
            autocomplete="username"
            :rules="[val => !!val || 'El usuario es requerido']"
          >
            <template #prepend>
              <q-icon name="person" />
            </template>
          </q-input>

          <q-input
            v-model="password"
            label="Contraseña"
            outlined
            dense
            :type="showPwd ? 'text' : 'password'"
            autocomplete="current-password"
            :rules="[val => !!val || 'La contraseña es requerida']"
          >
            <template #prepend>
              <q-icon name="lock" />
            </template>
            <template #append>
              <q-icon
                :name="showPwd ? 'visibility_off' : 'visibility'"
                class="cursor-pointer"
                @click="showPwd = !showPwd"
              />
            </template>
          </q-input>

          <q-banner v-if="error" type="negative" class="rounded-borders" dense>
            <template #avatar><q-icon name="error" /></template>
            {{ error }}
          </q-banner>

          <q-btn
            type="submit"
            color="primary"
            label="Iniciar Sesión"
            icon="login"
            class="full-width"
            size="md"
            :loading="loading"
            unelevated
          />
        </q-form>
      </q-card-section>

      <q-card-section class="text-center q-pt-none">
        <span class="text-grey-7">¿No tienes cuenta? </span>
        <router-link to="/register" class="text-primary text-weight-medium">Regístrate aquí</router-link>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const showPwd = ref(false)
const loading = ref(false)
const error = ref('')

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    router.push('/')
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>
