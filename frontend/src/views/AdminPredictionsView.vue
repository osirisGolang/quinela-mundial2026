<template>
  <q-page class="q-pa-md">
    <div class="row items-center q-mb-md">
      <div class="col">
        <div class="text-h5 text-primary text-weight-bold">
          <q-icon name="how_to_vote" class="q-mr-xs" />
          Pronósticos de todos los usuarios
        </div>
        <div class="text-caption text-grey-6">Vista de administrador — todos los grupos</div>
      </div>
      <div class="col-auto row q-gutter-sm items-center">
        <q-select
          v-model="selectedGroup"
          :options="groupOptions"
          label="Grupo"
          outlined dense clearable
          style="min-width:110px"
        />
        <q-btn icon="refresh" color="primary" outline dense @click="load" :loading="loading" label="Actualizar" />
      </div>
    </div>

    <div v-if="loading" class="text-center q-pa-xl">
      <q-spinner color="primary" size="3em" />
    </div>

    <template v-else-if="users.length && filteredMatches.length">
      <!-- Una tarjeta por partido -->
      <q-card
        v-for="match in filteredMatches"
        :key="match.id"
        flat bordered
        class="q-mb-md"
        :class="match.status === 'finished' ? 'bg-green-1' : ''"
      >
        <!-- Header del partido -->
        <q-card-section class="q-py-sm bg-primary text-white">
          <div class="row items-center no-wrap q-gutter-sm">
            <q-badge color="white" text-color="primary" :label="`#${match.match_number}`" class="text-weight-bold" />
            <div class="row items-center no-wrap q-gutter-xs">
              <img :src="`https://flagcdn.com/w40/${match.local_iso2}.png`" width="22" height="15" style="border-radius:2px" />
              <span class="text-weight-bold">{{ match.local_team }}</span>
            </div>
            <div class="text-weight-bold q-px-xs">
              <span v-if="match.status === 'finished'" class="text-amber-3">
                {{ match.real_local }} : {{ match.real_visitor }}
              </span>
              <span v-else class="text-white-6">vs</span>
            </div>
            <div class="row items-center no-wrap q-gutter-xs">
              <span class="text-weight-bold">{{ match.visitor_team }}</span>
              <img :src="`https://flagcdn.com/w40/${match.visitor_iso2}.png`" width="22" height="15" style="border-radius:2px" />
            </div>
            <q-space />
            <q-badge :color="match.status === 'finished' ? 'positive' : 'grey-5'"
              :label="match.status === 'finished' ? 'Finalizado' : 'Pendiente'" />
            <span class="text-caption text-white-7">Grupo {{ match.group_name }} · {{ formatDate(match.match_date) }}</span>
          </div>
        </q-card-section>

        <!-- Pronósticos de cada usuario -->
        <q-card-section class="q-py-sm">
          <div class="row q-gutter-sm flex-wrap">
            <div
              v-for="user in users"
              :key="user.id"
              class="pred-chip"
              :class="chipClass(match, user.username)"
            >
              <div class="text-caption text-weight-bold text-grey-7">{{ user.username }}</div>
              <div v-if="match.predictions[user.username]?.local !== null" class="text-h6 text-weight-bold">
                {{ match.predictions[user.username].local }} - {{ match.predictions[user.username].visitor }}
              </div>
              <div v-else class="text-grey-4 text-caption">—</div>
              <div v-if="match.status === 'finished' && match.predictions[user.username]?.local !== null">
                <q-badge
                  :color="pointsColor(match.predictions[user.username]?.points)"
                  :label="pointsLabel(match.predictions[user.username]?.points)"
                  class="text-weight-bold"
                />
              </div>
            </div>
          </div>
        </q-card-section>
      </q-card>
    </template>

    <div v-else-if="!loading" class="text-center q-pa-xl text-grey-6">
      <q-icon name="sentiment_neutral" size="4em" class="q-mb-md" />
      <div class="text-h6">No hay datos disponibles</div>
    </div>
  </q-page>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const users = ref([])
const matches = ref([])
const selectedGroup = ref(null)

const groupOptions = ['A','B','C','D','E','F','G','H','I','J','K','L']

const filteredMatches = computed(() => {
  if (!selectedGroup.value) return matches.value
  return matches.value.filter(m => m.group_name === selectedGroup.value)
})

async function load() {
  loading.value = true
  try {
    const { data } = await axios.get('http://localhost:8080/api/admin/predictions', {
      headers: { Authorization: auth.token }
    })
    users.value = data.users || []
    matches.value = (data.matches || []).map(m => ({
      id: m.id,
      match_number: m.match_number,
      group_name: m.group_name,
      local_team: m.local_team,
      local_iso2: m.local_iso2,
      visitor_team: m.visitor_team,
      visitor_iso2: m.visitor_iso2,
      match_date: m.match_date,
      real_local: m.real_local,
      real_visitor: m.real_visitor,
      status: m.status,
      predictions: m.predictions || {}
    }))
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function chipClass(match, username) {
  const pred = match.predictions[username]
  if (!pred || pred.local === null) return 'chip-empty'
  if (match.status !== 'finished') return 'chip-pending'
  if (pred.points === 3) return 'chip-exact'
  if (pred.points === 1) return 'chip-result'
  return 'chip-wrong'
}

function pointsColor(pts) {
  if (pts === 3) return 'positive'
  if (pts === 1) return 'warning'
  return 'negative'
}

function pointsLabel(pts) {
  if (pts === 3) return '🎯 3 pts'
  if (pts === 1) return '✅ 1 pt'
  return '❌ 0 pts'
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const [date] = dateStr.split(' ')
  const [y, m, d] = date.split('-')
  return `${d}/${m}/${y}`
}

onMounted(() => load())
</script>

<style scoped>
.pred-chip {
  min-width: 90px;
  padding: 6px 10px;
  border-radius: 8px;
  text-align: center;
  border: 1px solid #e0e0e0;
  background: #fafafa;
}
.chip-exact   { background: #e8f5e9; border-color: #4caf50; }
.chip-result  { background: #fff8e1; border-color: #ffc107; }
.chip-wrong   { background: #fce4ec; border-color: #f44336; }
.chip-pending { background: #e3f2fd; border-color: #2196f3; }
.chip-empty   { background: #f5f5f5; border-color: #e0e0e0; opacity: 0.5; }
</style>
