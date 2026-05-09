import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const API = 'http://localhost:8080'

// Limpiar cualquier dato previo en localStorage al arrancar
localStorage.removeItem('user')
localStorage.removeItem('token')

export const useAuthStore = defineStore('auth', () => {
  // Sin persistencia: la sesión solo dura mientras la app está abierta
  const user = ref(null)
  const token = ref(null)

  const isAdmin = computed(() => {
    const val = user.value?.is_admin
    return val === true || val === 1 || val === '1' || val === 'true'
  })

  async function login(username, password) {
    const res = await fetch(`${API}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Error al iniciar sesión')
    }
    const data = await res.json()
    user.value = data.user
    token.value = data.token
    return data
  }

  async function register(username, password) {
    const res = await fetch(`${API}/api/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Error al registrarse')
    }
    return await res.json()
  }

  function logout() {
    user.value = null
    token.value = null
  }

  return { user, token, isAdmin, login, register, logout }
})
