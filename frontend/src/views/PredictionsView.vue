<template>
  <q-page class="q-pa-md">

    <!-- Header con info del usuario -->
    <div class="row items-center q-mb-md no-print">
      <div class="col">
        <h4 class="q-my-none text-primary">
          <q-icon name="how_to_vote" class="q-mr-sm" />
          Mis Pronósticos — Fase de Grupos
        </h4>
        <div class="text-grey-7 q-mt-xs">
          Copa Mundial FIFA 2026 — 72 partidos en 12 grupos
        </div>
      </div>
      <div class="col-auto">
        <q-chip
          color="primary"
          text-color="white"
          icon="badge"
          size="lg"
          class="q-pa-md"
        >
          Código: <strong class="q-ml-xs">{{ auth.user?.username?.toUpperCase() }}</strong>
        </q-chip>
      </div>
    </div>

    <!-- Banner de estado de bloqueo -->
    <q-banner
      v-if="locked"
      class="bg-negative text-white q-mb-md rounded-borders no-print"
      dense
    >
      <template #avatar>
        <q-icon name="lock" />
      </template>
      <strong>COMPROMISO CERRADO</strong> — Tus pronósticos están bloqueados desde el {{ lockedAt }}.
      Ya no es posible modificarlos.
    </q-banner>

    <q-banner
      v-if="!locked && !loading"
      class="bg-info text-white q-mb-md rounded-borders no-print"
      dense
    >
      <template #avatar>
        <q-icon name="info" />
      </template>
      Ingresa el marcador que crees que quedará en cada partido. Cuando termines, haz clic en
      <strong>"Cerrar Compromiso"</strong> para bloquear tus pronósticos definitivamente.
    </q-banner>

    <!-- Cabecera de impresión (solo visible al imprimir) -->
    <div class="print-only print-header">
      <h2 class="text-center">Quinela Mundial 2026</h2>
      <p class="text-center">
        <strong>Usuario:</strong> {{ auth.user?.username?.toUpperCase() }} &nbsp;|&nbsp;
        <strong>Código:</strong> {{ auth.user?.username?.toUpperCase() }} &nbsp;|&nbsp;
        <strong>Fecha impresión:</strong> {{ printDate }}
      </p>
      <hr />
    </div>

    <!-- Cargando -->
    <div v-if="loading" class="text-center q-pa-xl">
      <q-spinner color="primary" size="4em" />
      <p class="q-mt-md text-grey-7">Cargando partidos...</p>
    </div>

    <!-- Tabs por grupo -->
    <template v-if="!loading">
      <q-tabs
        v-model="activeGroup"
        class="text-primary no-print"
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
        dense
      >
        <q-tab v-for="g in groupNames" :key="g" :name="g" :label="`Grupo ${g}`" />
      </q-tabs>

      <q-tab-panels v-model="activeGroup" animated class="no-print">
        <q-tab-panel v-for="g in groupNames" :key="g" :name="g" class="q-pa-none q-pt-md">
          <GroupMatchTable
            :group-name="g"
            :matches="matchesByGroup[g]"
            :locked="locked"
            @update-score="updateScore"
          />
        </q-tab-panel>
      </q-tab-panels>

      <!-- Vista de impresión: todos los grupos visibles -->
      <div class="print-only">
        <div v-for="g in groupNames" :key="'print-'+g" class="print-group">
          <h3>Grupo {{ g }}</h3>
          <GroupMatchTable
            :group-name="g"
            :matches="matchesByGroup[g]"
            :locked="true"
          />
        </div>
      </div>
    </template>

    <!-- Botones de acción -->
    <div v-if="!loading" class="row q-gutter-md q-mt-lg no-print">
      <q-btn
        v-if="!locked"
        color="primary"
        icon="save"
        label="Guardar / Modificar"
        :loading="saving"
        @click="savePredictions"
        size="lg"
        unelevated
      />

      <q-btn
        color="teal"
        icon="print"
        label="Imprimir PDF"
        @click="printPDF"
        size="lg"
        outline
      />

      <!-- Excel export -->
      <q-btn
        color="green-8"
        icon="download"
        label="Descargar Excel"
        @click="exportExcel"
        size="lg"
        outline
      >
        <q-tooltip>Descarga un archivo Excel con todos los partidos para completar offline</q-tooltip>
      </q-btn>

      <!-- Excel import -->
      <q-btn
        v-if="!locked"
        color="orange-8"
        icon="upload_file"
        label="Importar Excel"
        @click="triggerImport"
        size="lg"
        outline
        :loading="importing"
      >
        <q-tooltip>Sube el archivo Excel con tus pronósticos ya completados</q-tooltip>
      </q-btn>
      <!-- Hidden file input -->
      <input
        ref="fileInput"
        type="file"
        accept=".xlsx,.xls"
        style="display:none"
        @change="importExcel"
      />

      <q-btn
        v-if="!locked"
        color="negative"
        icon="lock"
        label="Cerrar Compromiso"
        @click="confirmLock = true"
        size="lg"
        unelevated
      />
    </div>

    <!-- Diálogo de confirmación de cierre -->
    <q-dialog v-model="confirmLock" persistent>
      <q-card style="min-width: 380px;">
        <q-card-section class="row items-center">
          <q-avatar icon="lock" color="negative" text-color="white" />
          <span class="q-ml-sm text-h6">¿Cerrar Compromiso?</span>
        </q-card-section>

        <q-card-section class="q-pt-none">
          <p>
            Al cerrar el compromiso, tus pronósticos quedarán
            <strong>bloqueados definitivamente</strong>. No podrás
            modificarlos después.
          </p>
          <p class="text-negative">Esta acción es irreversible.</p>
        </q-card-section>

        <q-card-actions align="right">
          <q-btn flat label="Cancelar" color="grey" v-close-popup />
          <q-btn
            unelevated
            color="negative"
            icon="lock"
            label="Sí, cerrar compromiso"
            :loading="locking"
            @click="lockPredictions"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

  </q-page>
</template>

<script setup>
import { ref, computed, onMounted, defineComponent, h } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useQuasar } from 'quasar'

const API = 'http://localhost:8080'
const auth = useAuthStore()
const router = useRouter()
const $q = useQuasar()

const loading = ref(true)
const saving = ref(false)
const importing = ref(false)
const locking = ref(false)
const confirmLock = ref(false)
const locked = ref(false)
const lockedAt = ref('')
const activeGroup = ref('A')
const allMatches = ref([])
const fileInput = ref(null)

const groupNames = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L']

const printDate = computed(() => new Date().toLocaleDateString('es-ES', {
  day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit'
}))

const matchesByGroup = computed(() => {
  const result = {}
  for (const g of groupNames) result[g] = []
  for (const m of allMatches.value) {
    if (result[m.group_name]) result[m.group_name].push(m)
  }
  return result
})

function formatDate(raw) {
  if (!raw) return ''
  const d = new Date(raw.replace(' ', 'T'))
  return d.toLocaleDateString('es-ES', {
    day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit'
  })
}

function updateScore(matchId, field, value) {
  const m = allMatches.value.find(m => m.match_id === matchId)
  if (!m) return
  if (field === 'local') {
    m.pred_local_score = value === '' || value === null ? null : parseInt(value)
  } else {
    m.pred_visitor_score = value === '' || value === null ? null : parseInt(value)
  }
}

async function fetchMatches() {
  const userId = auth.user?.id
  if (!userId) return

  try {
    const [matchRes, lockRes] = await Promise.all([
      fetch(`${API}/api/predictions/user/${userId}`),
      fetch(`${API}/api/predictions/lock/${userId}`)
    ])
    const matches = await matchRes.json()
    const lockData = await lockRes.json()

    allMatches.value = (matches || []).map(m => ({
      ...m,
      pred_local_score: m.pred_local_score ?? null,
      pred_visitor_score: m.pred_visitor_score ?? null,
    }))
    locked.value = lockData.locked === true
    if (lockData.locked_at) lockedAt.value = lockData.locked_at
  } catch (e) {
    $q.notify({ type: 'negative', message: 'Error al cargar los partidos' })
  } finally {
    loading.value = false
  }
}

async function savePredictions() {
  if (locked.value) return
  saving.value = true

  const userId = auth.user?.id
  let saved = 0
  let errors = 0

  for (const m of allMatches.value) {
    const ls = m.pred_local_score
    const vs = m.pred_visitor_score
    if (ls === null || ls === undefined || vs === null || vs === undefined) continue

    try {
      const res = await fetch(`${API}/api/predictions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
          match_id: m.match_id,
          local_score: ls,
          visitor_score: vs,
        })
      })
      if (res.ok) saved++
      else errors++
    } catch {
      errors++
    }
  }

  saving.value = false

  if (errors === 0) {
    $q.notify({ type: 'positive', message: `${saved} pronósticos guardados correctamente` })
  } else {
    $q.notify({ type: 'warning', message: `${saved} guardados, ${errors} errores` })
  }
}

async function lockPredictions() {
  locking.value = true
  try {
    // First save all current predictions
    await savePredictions()

    const res = await fetch(`${API}/api/predictions/lock`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: auth.user?.id })
    })
    if (res.ok) {
      locked.value = true
      lockedAt.value = new Date().toLocaleString('es-ES')
      confirmLock.value = false
      $q.notify({ type: 'positive', message: 'Compromiso cerrado. Tus pronósticos están bloqueados.' })
    } else {
      $q.notify({ type: 'negative', message: 'Error al cerrar el compromiso' })
    }
  } catch {
    $q.notify({ type: 'negative', message: 'Error de conexión' })
  } finally {
    locking.value = false
  }
}

function printPDF() {
  window.print()
}

function exportExcel() {
  const userId = auth.user?.id
  const username = auth.user?.username || 'usuario'
  const url = `${API}/api/predictions/export?user_id=${userId}&username=${encodeURIComponent(username)}`
  const a = document.createElement('a')
  a.href = url
  a.download = `quinela_mundial2026_${username.toLowerCase()}.xlsx`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

function triggerImport() {
  fileInput.value?.click()
}

async function importExcel(event) {
  const file = event.target.files?.[0]
  if (!file) return

  importing.value = true
  const formData = new FormData()
  formData.append('file', file)
  formData.append('user_id', String(auth.user?.id))

  try {
    const res = await fetch(`${API}/api/predictions/import`, {
      method: 'POST',
      body: formData,
    })
    const data = await res.json()
    if (res.ok) {
      $q.notify({
        type: 'positive',
        message: data.message || `${data.saved} pronósticos importados`,
        caption: data.skipped ? `${data.skipped} partidos sin completar omitidos` : '',
        timeout: 5000,
      })
      // Reload data to reflect imported predictions
      loading.value = true
      await fetchMatches()
    } else {
      $q.notify({ type: 'negative', message: data.error || 'Error al importar' })
    }
  } catch (e) {
    $q.notify({ type: 'negative', message: 'Error de conexión al importar' })
  } finally {
    importing.value = false
    // Reset file input so the same file can be re-selected
    if (fileInput.value) fileInput.value.value = ''
  }
}

// --- Inline component for group match table ---
const GroupMatchTable = defineComponent({
  name: 'GroupMatchTable',
  props: {
    groupName: String,
    matches: Array,
    locked: Boolean,
  },
  emits: ['update-score'],
  setup(props, { emit }) {
    function flagUrl(iso2) {
      return `https://flagcdn.com/24x18/${iso2?.toLowerCase()}.png`
    }
    function formatDate(raw) {
      if (!raw) return ''
      const d = new Date(raw.replace(' ', 'T'))
      return d.toLocaleDateString('es-ES', {
        day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit'
      })
    }
    function scoreClass(match) {
      if (match.pred_local_score === null || match.pred_visitor_score === null) return ''
      if (match.points === 3) return 'text-positive text-weight-bold'
      if (match.points === 1) return 'text-warning text-weight-bold'
      if (match.match_status === 'finished') return 'text-negative'
      return ''
    }
    return () => {
      const rows = (props.matches || []).map(m => {
        const ls = m.pred_local_score
        const vs = m.pred_visitor_score

        return h('tr', { key: m.match_id, class: 'match-row' }, [
          // Partido #
          h('td', { class: 'text-center text-caption col-num' }, `#${m.match_number}`),
          // Fecha
          h('td', { class: 'text-caption col-date' }, formatDate(m.match_date)),
          // Local team
          h('td', { class: 'text-right col-team' }, [
            h('span', { class: 'team-name q-mr-xs' }, m.local_team),
            h('img', { src: flagUrl(m.local_iso2), class: 'flag-img', alt: m.local_iso2 }),
          ]),
          // Score local input
          h('td', { class: 'text-center col-score' },
            props.locked
              ? h('span', { class: `score-display ${scoreClass(m)}` }, ls !== null && ls !== undefined ? ls : '-')
              : h('input', {
                  type: 'number', min: 0, max: 99,
                  value: ls !== null && ls !== undefined ? ls : '',
                  class: 'score-input',
                  onInput: (e) => emit('update-score', m.match_id, 'local', e.target.value === '' ? null : parseInt(e.target.value)),
                })
          ),
          // Separator
          h('td', { class: 'text-center col-sep text-weight-bold' }, '-'),
          // Score visitor input
          h('td', { class: 'text-center col-score' },
            props.locked
              ? h('span', { class: `score-display ${scoreClass(m)}` }, vs !== null && vs !== undefined ? vs : '-')
              : h('input', {
                  type: 'number', min: 0, max: 99,
                  value: vs !== null && vs !== undefined ? vs : '',
                  class: 'score-input',
                  onInput: (e) => emit('update-score', m.match_id, 'visitor', e.target.value === '' ? null : parseInt(e.target.value)),
                })
          ),
          // Visitor team
          h('td', { class: 'col-team' }, [
            h('img', { src: flagUrl(m.visitor_iso2), class: 'flag-img q-mr-xs', alt: m.visitor_iso2 }),
            h('span', { class: 'team-name' }, m.visitor_team),
          ]),
          // Points badge
          h('td', { class: 'text-center col-pts' },
            m.match_status === 'finished' && (m.pred_local_score !== null)
              ? h('span', {
                  class: 'pts-badge ' + (m.points === 3 ? 'pts-3' : m.points === 1 ? 'pts-1' : 'pts-0')
                }, `${m.points ?? 0}pts`)
              : h('span', { class: 'text-grey-5 text-caption' }, m.pred_local_score !== null ? '—' : '')
          ),
          // City
          h('td', { class: 'text-caption text-grey-6 col-city' }, m.city),
        ])
      })

      return h('div', { class: 'group-table-wrapper' }, [
        h('table', { class: 'match-table full-width' }, [
          h('thead', {}, [
            h('tr', { class: 'table-header' }, [
              h('th', { class: 'col-num' }, '#'),
              h('th', { class: 'col-date' }, 'Fecha'),
              h('th', { class: 'col-team text-right' }, 'Local'),
              h('th', { class: 'col-score' }, 'G'),
              h('th', { class: 'col-sep' }, ''),
              h('th', { class: 'col-score' }, 'G'),
              h('th', { class: 'col-team' }, 'Visitante'),
              h('th', { class: 'col-pts' }, 'Pts'),
              h('th', { class: 'col-city' }, 'Ciudad'),
            ])
          ]),
          h('tbody', {}, rows),
        ])
      ])
    }
  }
})

onMounted(async () => {
  if (!auth.user) {
    router.push('/login')
    return
  }
  await fetchMatches()
})
</script>

<style scoped>
/* =========================================================
   Tabla de partidos
   ========================================================= */
.group-table-wrapper {
  overflow-x: auto;
}

.match-table {
  border-collapse: collapse;
  font-size: 0.85rem;
}

.match-table th,
.match-table td {
  padding: 6px 8px;
  border-bottom: 1px solid #e0e0e0;
  white-space: nowrap;
}

.match-table .table-header th {
  background: #1976d2;
  color: white;
  font-weight: 600;
  text-align: center;
}

.match-row:hover {
  background: #f5f5f5;
}

.col-num   { width: 40px;  text-align: center; }
.col-date  { width: 160px; }
.col-team  { min-width: 140px; }
.col-score { width: 50px;  text-align: center; }
.col-sep   { width: 20px;  text-align: center; }
.col-pts   { width: 60px;  text-align: center; }
.col-city  { min-width: 120px; }

.flag-img {
  width: 24px;
  height: 18px;
  vertical-align: middle;
  border: 1px solid #ccc;
  border-radius: 2px;
}

.team-name {
  vertical-align: middle;
}

/* Inputs de score */
.score-input {
  width: 44px;
  text-align: center;
  border: 1.5px solid #1976d2;
  border-radius: 4px;
  padding: 4px 2px;
  font-size: 1rem;
  font-weight: 600;
  outline: none;
  background: #e3f2fd;
  color: #1565c0;
  transition: border-color 0.2s;
  -moz-appearance: textfield;
}

.score-input::-webkit-outer-spin-button,
.score-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.score-input:focus {
  border-color: #0d47a1;
  background: #bbdefb;
}

.score-display {
  font-size: 1.1rem;
  font-weight: 700;
  min-width: 24px;
  display: inline-block;
  text-align: center;
}

/* Points badges */
.pts-badge {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 700;
}

.pts-3 { background: #e8f5e9; color: #2e7d32; border: 1px solid #4caf50; }
.pts-1 { background: #fff8e1; color: #f57f17; border: 1px solid #ffca28; }
.pts-0 { background: #ffebee; color: #c62828; border: 1px solid #ef9a9a; }

/* =========================================================
   Print
   ========================================================= */
.print-only { display: none; }

@media print {
  .no-print { display: none !important; }
  .print-only { display: block !important; }

  .print-header h2 { font-size: 1.4rem; margin: 0; }
  .print-header p  { margin: 4px 0 8px; font-size: 0.9rem; }

  .print-group { page-break-inside: avoid; margin-bottom: 20px; }
  .print-group h3 { background: #1976d2; color: white; padding: 4px 8px; font-size: 1rem; margin: 8px 0 4px; }

  .match-table th {
    background: #1976d2 !important;
    color: white !important;
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }

  .score-input { border: 1px solid #999; background: white; }

  .pts-3 { background: #e8f5e9 !important; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  .pts-1 { background: #fff8e1 !important; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  .pts-0 { background: #ffebee !important; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
}
</style>
