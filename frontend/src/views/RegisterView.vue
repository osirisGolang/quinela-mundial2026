<template>
  <q-page class="flex flex-center bg-grey-1">
    <q-card style="min-width: 380px;" class="q-pa-lg shadow-2">
      <q-card-section class="text-center q-pb-none">
        <q-icon name="sports_soccer" size="3em" color="primary" />
        <div class="text-h5 text-primary text-weight-bold q-mt-sm">Crear Cuenta</div>
        <div class="text-subtitle2 text-grey-6">Únete a la quiniela del Mundial 2026</div>
      </q-card-section>

      <q-card-section>
        <q-form @submit.prevent="onSubmit" class="q-gutter-md">
          <q-input
            v-model="username"
            label="Usuario"
            outlined
            dense
            autocomplete="username"
            :rules="[
              val => !!val || 'El usuario es requerido',
              val => val.length >= 3 || 'Mínimo 3 caracteres',
              val => /^[a-zA-Z0-9_]+$/.test(val) || 'Solo letras, números y guión bajo'
            ]"
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
            autocomplete="new-password"
            :rules="[
              val => !!val || 'La contraseña es requerida',
              val => val.length >= 6 || 'Mínimo 6 caracteres'
            ]"
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

          <q-input
            v-model="confirmPassword"
            label="Confirmar Contraseña"
            outlined
            dense
            :type="showPwd ? 'text' : 'password'"
            autocomplete="new-password"
            :rules="[
              val => !!val || 'Confirma tu contraseña',
              val => val === password || 'Las contraseñas no coinciden'
            ]"
          >
            <template #prepend>
              <q-icon name="lock_outline" />
            </template>
          </q-input>

          <q-banner v-if="error" type="negative" class="rounded-borders" dense>
            <template #avatar><q-icon name="error" /></template>
            {{ error }}
          </q-banner>

          <q-banner v-if="success" type="positive" class="rounded-borders" dense>
            <template #avatar><q-icon name="check_circle" /></template>
            ¡Cuenta creada exitosamente! Redirigiendo...
          </q-banner>

          <q-btn
            type="submit"
            color="primary"
            label="Registrarse"
            icon="person_add"
            class="full-width"
            size="md"
            :loading="loading"
            unelevated
          />
        </q-form>
      </q-card-section>

      <q-card-section class="text-center q-pt-none">
        <span class="text-grey-7">¿Ya tienes cuenta? </span>
        <router-link to="/login" class="text-primary text-weight-medium">Inicia sesión aquí</router-link>
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
const confirmPassword = ref('')
const showPwd = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref(false)

async function onSubmit() {
  error.value = ''
  if (password.value !== confirmPassword.value) {
    error.value = 'Las contraseñas no coinciden'
    return
  }
  loading.value = true
  try {
    await auth.register(username.value, password.value)
    success.value = true
    setTimeout(() => router.push('/login'), 1500)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>
