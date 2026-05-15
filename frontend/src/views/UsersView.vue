<template>
  <q-page class="q-pa-md">

    <!-- Access guard -->
    <div v-if="!auth.isAdmin" class="flex flex-center column q-pa-xl">
      <q-icon name="lock" size="5em" color="negative" />
      <p class="text-h6 text-negative q-mt-md">Acceso restringido — Solo administradores</p>
      <q-btn color="primary" label="Ir al inicio" to="/" unelevated class="q-mt-md" />
    </div>

    <template v-else>
      <!-- Header -->
      <div class="row items-center q-mb-md">
        <div class="col">
          <h4 class="q-my-none text-primary">
            <q-icon name="manage_accounts" class="q-mr-sm" />
            Gestión de Usuarios
          </h4>
          <div class="text-grey-7 text-caption">Solo visible para administradores</div>
        </div>
        <div class="col-auto">
          <q-btn
            color="primary"
            icon="person_add"
            label="Nuevo Usuario"
            unelevated
            @click="openCreateDialog"
          />
        </div>
      </div>

      <!-- Stats bar -->
      <div class="row q-gutter-md q-mb-md">
        <q-card flat bordered class="stat-card">
          <q-card-section class="text-center q-pa-sm">
            <div class="text-h4 text-primary text-weight-bold">{{ users.length }}</div>
            <div class="text-caption text-grey-7">Total usuarios</div>
          </q-card-section>
        </q-card>
        <q-card flat bordered class="stat-card">
          <q-card-section class="text-center q-pa-sm">
            <div class="text-h4 text-positive text-weight-bold">{{ users.filter(u=>!u.locked && u.pred_count>0).length }}</div>
            <div class="text-caption text-grey-7">Con pronósticos</div>
          </q-card-section>
        </q-card>
        <q-card flat bordered class="stat-card">
          <q-card-section class="text-center q-pa-sm">
            <div class="text-h4 text-negative text-weight-bold">{{ users.filter(u=>u.locked).length }}</div>
            <div class="text-caption text-grey-7">Compromisos cerrados</div>
          </q-card-section>
        </q-card>
        <q-card flat bordered class="stat-card">
          <q-card-section class="text-center q-pa-sm">
            <div class="text-h4 text-amber-8 text-weight-bold">{{ users.filter(u=>u.is_admin).length }}</div>
            <div class="text-caption text-grey-7">Administradores</div>
          </q-card-section>
        </q-card>
      </div>

      <!-- Users table -->
      <q-table
        :rows="users"
        :columns="columns"
        row-key="id"
        flat
        bordered
        :loading="loading"
        :filter="search"
        :pagination="{ rowsPerPage: 15 }"
        rows-per-page-label="Filas por página"
      >
        <template #top-left>
          <q-input
            v-model="search"
            dense
            outlined
            placeholder="Buscar usuario..."
            clearable
            style="min-width: 220px"
          >
            <template #prepend><q-icon name="search" /></template>
          </q-input>
        </template>

        <!-- Username -->
        <template #body-cell-username="props">
          <q-td :props="props">
            <div class="row items-center no-wrap q-gutter-xs">
              <q-avatar size="28px" :color="props.row.is_admin ? 'amber-8' : 'primary'" text-color="white" font-size="12px">
                {{ props.row.username.charAt(0).toUpperCase() }}
              </q-avatar>
              <span class="text-weight-medium">{{ props.row.username }}</span>
            </div>
          </q-td>
        </template>

        <!-- Role badge -->
        <template #body-cell-is_admin="props">
          <q-td :props="props" class="text-center">
            <q-badge
              :color="props.row.is_admin ? 'amber-8' : 'blue-grey-5'"
              :label="props.row.is_admin ? 'Admin' : 'Usuario'"
              :text-color="props.row.is_admin ? 'black' : 'white'"
            />
          </q-td>
        </template>

        <!-- Predictions count -->
        <template #body-cell-pred_count="props">
          <q-td :props="props" class="text-center">
            <q-chip
              :color="props.row.pred_count > 0 ? 'positive' : 'grey-4'"
              :text-color="props.row.pred_count > 0 ? 'white' : 'grey-7'"
              dense
              size="sm"
            >
              {{ props.row.pred_count }} / 72
            </q-chip>
          </q-td>
        </template>

        <!-- Points -->
        <template #body-cell-points="props">
          <q-td :props="props" class="text-center">
            <span class="text-weight-bold text-primary">{{ props.row.points }}</span>
          </q-td>
        </template>

        <!-- Lock status -->
        <template #body-cell-locked="props">
          <q-td :props="props" class="text-center">
            <q-icon
              :name="props.row.locked ? 'lock' : 'lock_open'"
              :color="props.row.locked ? 'negative' : 'positive'"
              size="sm"
            >
              <q-tooltip>{{ props.row.locked ? 'Compromiso cerrado' : 'Compromiso abierto' }}</q-tooltip>
            </q-icon>
          </q-td>
        </template>

        <!-- Actions -->
        <template #body-cell-actions="props">
          <q-td :props="props" class="text-center">
            <q-btn-group flat>
              <q-btn flat round dense icon="edit" color="primary" size="sm" @click="openEditDialog(props.row)">
                <q-tooltip>Editar usuario</q-tooltip>
              </q-btn>
              <q-btn flat round dense icon="key" color="orange-7" size="sm" @click="openPasswordDialog(props.row)">
                <q-tooltip>Cambiar contraseña</q-tooltip>
              </q-btn>
            
              <q-btn
                flat round dense icon="delete" color="negative" size="sm"
                :disable="props.row.id === auth.user.id"
                @click="openDeleteDialog(props.row)"
              >
                <q-tooltip>{{ props.row.id === auth.user.id ? 'No puedes eliminarte a ti mismo' : 'Eliminar usuario' }}</q-tooltip>
              </q-btn>
            </q-btn-group>
          </q-td>
        </template>
      </q-table>
    </template>

    <!-- ─── Dialog: Create / Edit user ────────────────────────── -->
    <q-dialog v-model="showEditDialog" persistent>
      <q-card style="min-width: 420px;">
        <q-card-section class="bg-primary text-white row items-center">
          <q-icon :name="editMode === 'create' ? 'person_add' : 'edit'" size="sm" class="q-mr-sm" />
          <span class="text-h6">{{ editMode === 'create' ? 'Nuevo Usuario' : 'Editar Usuario' }}</span>
          <q-space />
          <q-btn flat round dense icon="close" v-close-popup />
        </q-card-section>

        <q-card-section class="q-gutter-md q-pt-lg">
          <q-input
            v-model="editForm.username"
            label="Nombre de usuario"
            outlined dense
            :rules="[v => !!v || 'Requerido', v => v.length >= 3 || 'Mínimo 3 caracteres']"
          >
            <template #prepend><q-icon name="person" /></template>
          </q-input>

          <q-input
            v-if="editMode === 'create'"
            v-model="editForm.password"
            label="Contraseña"
            outlined dense
            type="password"
            :rules="[v => !!v || 'Requerido', v => v.length >= 4 || 'Mínimo 4 caracteres']"
          >
            <template #prepend><q-icon name="lock" /></template>
          </q-input>

          <q-toggle
            v-model="editForm.is_admin"
            label="Rol Administrador"
            color="amber-8"
            left-label
          />
        </q-card-section>

        <q-card-actions align="right" class="q-pa-md q-pt-none">
          <q-btn flat label="Cancelar" v-close-popup />
          <q-btn
            unelevated
            :color="editMode === 'create' ? 'primary' : 'positive'"
            :label="editMode === 'create' ? 'Crear Usuario' : 'Guardar Cambios'"
            :loading="saving"
            @click="saveUser"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- ─── Dialog: Change password ────────────────────────────── -->
    <q-dialog v-model="showPasswordDialog" persistent>
      <q-card style="min-width: 380px;">
        <q-card-section class="bg-orange-7 text-white row items-center">
          <q-icon name="key" size="sm" class="q-mr-sm" />
          <span class="text-h6">Cambiar Contraseña — {{ passwordTarget?.username }}</span>
          <q-space />
          <q-btn flat round dense icon="close" v-close-popup />
        </q-card-section>

        <q-card-section class="q-pt-lg">
          <q-input
            v-model="newPassword"
            label="Nueva contraseña"
            outlined dense
            :type="showNewPwd ? 'text' : 'password'"
            :rules="[v => !!v || 'Requerido', v => v.length >= 4 || 'Mínimo 4 caracteres']"
          >
            <template #prepend><q-icon name="lock" /></template>
            <template #append>
              <q-icon :name="showNewPwd ? 'visibility_off' : 'visibility'" class="cursor-pointer" @click="showNewPwd = !showNewPwd" />
            </template>
          </q-input>
        </q-card-section>

        <q-card-actions align="right" class="q-pa-md q-pt-none">
          <q-btn flat label="Cancelar" v-close-popup />
          <q-btn unelevated color="orange-7" label="Cambiar Contraseña" :loading="saving" @click="savePassword" />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- ─── Dialog: Delete confirmation ───────────────────────── -->
    <q-dialog v-model="showDeleteDialog" persistent>
      <q-card style="min-width: 360px;">
        <q-card-section class="row items-center">
          <q-avatar icon="delete_forever" color="negative" text-color="white" />
          <span class="q-ml-sm text-h6">¿Eliminar usuario?</span>
        </q-card-section>
        <q-card-section class="q-pt-none">
          Se eliminará el usuario <strong>{{ deleteTarget?.username }}</strong> junto con
          todos sus pronósticos. <span class="text-negative">Esta acción es irreversible.</span>
        </q-card-section>
        <q-card-actions align="right">
          <q-btn flat label="Cancelar" v-close-popup />
          <q-btn unelevated color="negative" label="Eliminar" :loading="deleting" @click="deleteUser" />
        </q-card-actions>
      </q-card>
    </q-dialog>

  </q-page>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useQuasar } from 'quasar'

const API = 'http://localhost:8080'
const auth = useAuthStore()
const $q = useQuasar()

const users   = ref([])
const loading = ref(false)
const search  = ref('')

// Edit dialog
const showEditDialog = ref(false)
const editMode = ref('create')  // 'create' | 'edit'
const editTarget = ref(null)
const editForm = ref({ username: '', password: '', is_admin: false })
const saving = ref(false)

// Password dialog
const showPasswordDialog = ref(false)
const passwordTarget = ref(null)
const newPassword = ref('')
const showNewPwd = ref(false)

// Delete dialog
const showDeleteDialog = ref(false)
const deleteTarget = ref(null)
const deleting = ref(false)

const columns = [
  { name: 'id',        label: 'ID',          field: 'id',         align: 'center', sortable: true, style: 'width:50px' },
  { name: 'username',  label: 'Usuario',      field: 'username',   align: 'left',   sortable: true },
  { name: 'is_admin',  label: 'Rol',          field: 'is_admin',   align: 'center', sortable: true },
  { name: 'pred_count',label: 'Pronósticos',  field: 'pred_count', align: 'center', sortable: true },
  { name: 'points',    label: 'Puntos',       field: 'points',     align: 'center', sortable: true },
  { name: 'locked',    label: 'Compromiso',   field: 'locked',     align: 'center', sortable: true },
  { name: 'actions',   label: 'Acciones',     field: 'actions',    align: 'center', sortable: false },
]

function authHeaders() {
  return {
    'Content-Type': 'application/json',
    'Authorization': auth.token || '',
  }
}

async function fetchUsers() {
  loading.value = true
  try {
    const res = await fetch(`${API}/api/users`, {
      headers: authHeaders(),
    })
    if (!res.ok) throw new Error((await res.json()).error)
    users.value = await res.json() || []
  } catch (e) {
    $q.notify({ type: 'negative', message: e.message || 'Error cargando usuarios' })
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editMode.value = 'create'
  editTarget.value = null
  editForm.value = { username: '', password: '', is_admin: false }
  showEditDialog.value = true
}

function openEditDialog(user) {
  editMode.value = 'edit'
  editTarget.value = user
  editForm.value = { username: user.username, password: '', is_admin: user.is_admin }
  showEditDialog.value = true
}

async function saveUser() {
  if (!editForm.value.username || editForm.value.username.length < 3) return
  if (editMode.value === 'create' && (!editForm.value.password || editForm.value.password.length < 4)) return

  saving.value = true
  try {
    if (editMode.value === 'create') {
      const res = await fetch(`${API}/api/users`, {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({
          username: editForm.value.username,
          password: editForm.value.password,
        }),
      })
      if (!res.ok) throw new Error((await res.json()).error)
      const newUser = await res.json()
      // Set admin role if needed
      if (editForm.value.is_admin) {
        await fetch(`${API}/api/users/${newUser.id}`, {
          method: 'PUT',
          headers: authHeaders(),
          body: JSON.stringify({ is_admin: true }),
        })
      }
      $q.notify({ type: 'positive', message: `Usuario "${editForm.value.username}" creado` })
    } else {
      const payload = {
        username: editForm.value.username,
        is_admin: editForm.value.is_admin,
      }
      const res = await fetch(`${API}/api/users/${editTarget.value.id}`, {
        method: 'PUT',
        headers: authHeaders(),
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw new Error((await res.json()).error)
      $q.notify({ type: 'positive', message: 'Usuario actualizado' })
    }
    showEditDialog.value = false
    await fetchUsers()
  } catch (e) {
    $q.notify({ type: 'negative', message: e.message || 'Error guardando usuario' })
  } finally {
    saving.value = false
  }
}

function openPasswordDialog(user) {
  passwordTarget.value = user
  newPassword.value = ''
  showNewPwd.value = false
  showPasswordDialog.value = true
}

async function savePassword() {
  if (!newPassword.value || newPassword.value.length < 4) return
  saving.value = true
  try {
    const res = await fetch(`${API}/api/users/${passwordTarget.value.id}`, {
      method: 'PUT',
      headers: authHeaders(),
      body: JSON.stringify({ password: newPassword.value }),
    })
    if (!res.ok) throw new Error((await res.json()).error)
    $q.notify({ type: 'positive', message: `Contraseña de "${passwordTarget.value.username}" actualizada` })
    showPasswordDialog.value = false
  } catch (e) {
    $q.notify({ type: 'negative', message: e.message || 'Error al cambiar contraseña' })
  } finally {
    saving.value = false
  }
}

async function toggleLock(user) {
  const newLocked = !user.locked
  const action = newLocked ? 'cerrar' : 'abrir'
  try {
    const res = await fetch(`${API}/api/users/${user.id}`, {
      method: 'PUT',
      headers: authHeaders(),
      body: JSON.stringify({ locked: newLocked }),
    })
    if (!res.ok) throw new Error((await res.json()).error)
    $q.notify({
      type: newLocked ? 'warning' : 'positive',
      message: `Compromiso de "${user.username}" ${newLocked ? 'cerrado' : 'abierto'}`,
    })
    await fetchUsers()
  } catch (e) {
    $q.notify({ type: 'negative', message: `Error al ${action} compromiso: ${e.message}` })
  }
}

async function toggleAdmin(user) {
  const newAdmin = !user.is_admin
  try {
    const res = await fetch(`${API}/api/users/${user.id}`, {
      method: 'PUT',
      headers: authHeaders(),
      body: JSON.stringify({ is_admin: newAdmin }),
    })
    if (!res.ok) throw new Error((await res.json()).error)
    $q.notify({
      type: 'positive',
      message: `"${user.username}" ${newAdmin ? 'ahora es administrador' : 'ya no es administrador'}`,
    })
    await fetchUsers()
  } catch (e) {
    $q.notify({ type: 'negative', message: `Error al cambiar rol: ${e.message}` })
  }
}

function openDeleteDialog(user) {
  deleteTarget.value = user
  showDeleteDialog.value = true
}

async function deleteUser() {
  deleting.value = true
  try {
    const res = await fetch(`${API}/api/users/${deleteTarget.value.id}`, {
      method: 'DELETE',
      headers: authHeaders(),
    })
    if (!res.ok) throw new Error((await res.json()).error)
    $q.notify({ type: 'positive', message: `Usuario "${deleteTarget.value.username}" eliminado` })
    showDeleteDialog.value = false
    await fetchUsers()
  } catch (e) {
    $q.notify({ type: 'negative', message: e.message || 'Error al eliminar usuario' })
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  if (auth.isAdmin) fetchUsers()
})
</script>

<style scoped>
.stat-card {
  min-width: 130px;
  border-radius: 8px;
}
</style>
