<template>
  <q-page class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div>
        <div class="text-h5 text-primary text-weight-bold">
          <q-icon name="leaderboard" class="q-mr-xs" />
          Tabla General de la Quiniela
        </div>
        <div class="text-caption text-grey-6 q-mt-xs" v-if="lastUpdated">
          <q-icon name="schedule" size="xs" class="q-mr-xs" />
          Actualizado: {{ lastUpdated }} — se recarga automáticamente
        </div>
      </div>
      <div class="row items-center q-gutter-sm">
        <q-badge color="green-7" :label="`${matchesFinished} partidos finalizados`" />
        <q-btn
          color="primary"
          icon="refresh"
          label="Actualizar"
          outline
          dense
          @click="loadStandings"
          :loading="loading"
        />
      </div>
    </div>

    <div v-if="loading && standings.length === 0" class="text-center q-pa-xl">
      <q-spinner color="primary" size="3em" />
      <div class="text-grey-6 q-mt-md">Cargando tabla de posiciones...</div>
    </div>

    <q-table
      v-else
      :rows="standings"
      :columns="columns"
      row-key="user_id"
      flat
      bordered
      :pagination="{ rowsPerPage: 50 }"
      :loading="loading"
    >
      <template #header="props">
        <q-tr class="bg-primary text-white text-weight-bold">
          <q-th v-for="col in props.cols" :key="col.name" :props="props" class="text-center">
            {{ col.label }}
          </q-th>
        </q-tr>
      </template>

      <template #body="props">
        <q-tr :props="props" :class="rowClass(props.rowIndex)">
          <!-- Posición -->
          <q-td class="text-center">
            <q-badge
              :color="podiumColor(props.rowIndex + 1)"
              text-color="white"
              :label="props.rowIndex + 1"
              class="text-weight-bold"
              style="min-width: 28px; font-size: 13px;"
            />
          </q-td>

          <!-- Usuario -->
          <q-td>
            <div class="row items-center no-wrap q-gutter-sm">
              <q-avatar size="32px" :color="avatarColor(props.rowIndex + 1)" text-color="white" class="text-weight-bold">
                {{ props.row.username.charAt(0).toUpperCase() }}
              </q-avatar>
              <span class="text-weight-medium">{{ props.row.username }}</span>
              <q-badge v-if="props.rowIndex === 0 && props.row.total_points > 0" color="amber-8" icon="emoji_events" label="Líder" />
            </div>
          </q-td>

          <!-- Puntos totales -->
          <q-td class="text-center">
            <q-badge
              color="primary"
              text-color="white"
              :label="props.row.total_points"
              class="text-weight-bold"
              style="min-width: 40px; font-size: 15px;"
            />
          </q-td>

          <!-- Marcador exacto (3 pts) -->
          <q-td class="text-center">
            <div class="row items-center justify-center q-gutter-xs">
              <q-icon name="gps_fixed" color="positive" size="xs" />
              <span class="text-positive text-weight-bold">{{ props.row.exact_score }}</span>
            </div>
          </q-td>

          <!-- Solo resultado (1 pt) -->
          <q-td class="text-center">
            <div class="row items-center justify-center q-gutter-xs">
              <q-icon name="check_circle_outline" color="warning" size="xs" />
              <span class="text-warning text-weight-medium">{{ props.row.result_only }}</span>
            </div>
          </q-td>

          <!-- Incorrectos (0 pts) -->
          <q-td class="text-center">
            <div class="row items-center justify-center q-gutter-xs">
              <q-icon name="cancel" color="negative" size="xs" />
              <span class="text-grey-6">{{ props.row.wrong }}</span>
            </div>
          </q-td>

          <!-- Pronósticos ingresados -->
          <q-td class="text-center">
            <q-linear-progress
              :value="props.row.pred_count / 72"
              color="blue-4"
              track-color="grey-3"
              rounded
              style="height: 8px; width: 80px; margin: 0 auto;"
            />
            <div class="text-caption text-grey-7 q-mt-xs">{{ props.row.pred_count }}/72</div>
          </q-td>
        </q-tr>
      </template>

      <template #no-data>
        <div class="full-width text-center q-pa-xl text-grey-6">
          <q-icon name="sentiment_neutral" size="4em" class="q-mb-md" />
          <div class="text-h6">Aún no hay pronósticos registrados</div>
          <div class="text-caption">¡Sé el primero en registrar tus pronósticos!</div>
        </div>
      </template>
    </q-table>

    <!-- Leyenda del sistema de puntos -->
    <q-card flat bordered class="q-mt-lg">
      <q-card-section>
        <div class="text-subtitle1 text-weight-bold q-mb-sm">
          <q-icon name="info" class="q-mr-xs text-primary" />
          Sistema de Puntuación
        </div>
        <div class="row q-gutter-md">
          <div class="row items-center q-gutter-xs">
            <q-badge color="positive" label="3 pts" />
            <span class="text-grey-7">Marcador exacto</span>
          </div>
          <div class="row items-center q-gutter-xs">
            <q-badge color="warning" text-color="black" label="1 pt" />
            <span class="text-grey-7">Resultado correcto (sin marcador exacto)</span>
          </div>
          <div class="row items-center q-gutter-xs">
            <q-badge color="grey-5" label="0 pts" />
            <span class="text-grey-7">Pronóstico incorrecto</span>
          </div>
          <div class="row items-center q-gutter-xs">
            <q-icon name="info_outline" color="blue-4" />
            <span class="text-grey-7">Desempate: marcadores exactos → resultados correctos → nombre</span>
          </div>
        </div>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import axios from 'axios'

const standings = ref([])
const loading = ref(false)
const lastUpdated = ref('')
const matchesFinished = ref(0)
let refreshInterval = null

const columns = [
  { name: 'position', label: '#', align: 'center', sortable: false, style: 'width: 50px;' },
  { name: 'username', label: 'Participante', align: 'left', field: 'username', sortable: false },
  { name: 'total_points', label: 'Pts Total', align: 'center', field: 'total_points', sortable: false },
  { name: 'exact_score', label: '🎯 Exactos (3pts)', align: 'center', field: 'exact_score', sortable: false },
  { name: 'result_only', label: '✅ Resultado (1pt)', align: 'center', field: 'result_only', sortable: false },
  { name: 'wrong', label: '❌ Incorrectos', align: 'center', field: 'wrong', sortable: false },
  { name: 'pred_count', label: 'Pronósticos', align: 'center', field: 'pred_count', sortable: false },
]

async function loadStandings() {
  loading.value = true
  try {
    const [standRes, matchRes] = await Promise.all([
      axios.get('http://localhost:8080/api/standings'),
      axios.get('http://localhost:8080/api/matches'),
    ])
    standings.value = standRes.data || []
    matchesFinished.value = (matchRes.data || []).filter(m => m.status === 'finished').length
    const now = new Date()
    lastUpdated.value = now.toLocaleTimeString('es', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch (e) {
    console.error('Error loading standings:', e)
  } finally {
    loading.value = false
  }
}

function podiumColor(pos) {
  if (pos === 1) return 'amber-8'
  if (pos === 2) return 'grey-5'
  if (pos === 3) return 'brown-4'
  return 'blue-grey-4'
}

function avatarColor(pos) {
  if (pos === 1) return 'amber-8'
  if (pos === 2) return 'grey-6'
  if (pos === 3) return 'brown-4'
  return 'primary'
}

function rowClass(idx) {
  if (idx === 0) return 'bg-amber-1'
  if (idx === 1) return 'bg-grey-2'
  return ''
}

onMounted(() => {
  loadStandings()
  // Auto-refresh cada 30 segundos
  refreshInterval = setInterval(loadStandings, 30000)
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})
</script>
