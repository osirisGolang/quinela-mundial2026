<template>
  <q-page class="q-pa-md">
    <h4 class="text-primary">Calendario por Seleccion</h4>

    <div class="row q-mb-md items-center">
      <q-select
        v-model="selectedTeam"
        :options="filteredTeamOptions"
        option-value="iso2"
        option-label="name"
        emit-value
        map-options
        use-input
        input-debounce="0"
        label="Seleccionar Equipo"
        outlined
        dense
        clearable
        style="min-width: 350px;"
        class="col-12 col-md-6"
        @filter="filterTeams"
        @clear="selectedTeam = null"
      >
        <template #option="{ itemProps, opt }">
          <q-item v-bind="itemProps">
            <q-item-section avatar>
              <img :src="`https://flagcdn.com/w40/${opt.iso2}.png`" width="28" height="19" style="border-radius:2px;border:1px solid #ccc" />
            </q-item-section>
            <q-item-section>
              <q-item-label>{{ opt.name }}</q-item-label>
              <q-item-label caption>Grupo {{ opt.group }}</q-item-label>
            </q-item-section>
          </q-item>
        </template>
        <template #selected-item="{ opt }">
          <div class="row items-center no-wrap q-gutter-xs">
            <img :src="`https://flagcdn.com/w40/${opt.iso2}.png`" width="28" height="19" style="border-radius:2px;border:1px solid #ccc" />
            <span>{{ opt.name }}</span>
          </div>
        </template>
        <template #no-option>
          <q-item>
            <q-item-section class="text-grey">No se encontró el equipo</q-item-section>
          </q-item>
        </template>
      </q-select>

      <q-space />

      <div class="text-caption text-grey-7 q-pr-md">
        {{ filteredMatches.length }} partidos encontrados
      </div>
    </div>

    <q-table
      :rows="paginatedMatches"
      :columns="columns"
      row-key="id"
      flat
      bordered
      :pagination="pagination"
      @update:pagination="val => pagination = val"
      class="matches-table"
    >
      <template #header="props">
        <q-tr class="bg-primary text-white">
          <q-th v-for="col in props.cols" :key="col.name" :props="props" class="text-center">
            {{ col.label }}
          </q-th>
        </q-tr>
      </template>

      <template #body="props">
        <q-tr>
          <q-td class="text-center">
            <q-badge color="primary" :label="`#${props.row.match_number}`" />
          </q-td>
          <q-td class="text-center text-caption">
            {{ formatDate(props.row.match_date) }}
          </q-td>
          <q-td class="text-center text-caption">
            {{ formatTime(props.row.match_date) }}
          </q-td>
          <q-td class="text-center">
            <div class="row items-center justify-end no-wrap">
              <span class="text-weight-medium">{{ props.row.local_team }}</span>
              <img :src="`https://flagcdn.com/w40/${props.row.local_iso2}.png`" width="28" height="19" style="border-radius:2px;border:1px solid #ccc" class="q-ml-sm" />
            </div>
          </q-td>
          <q-td class="text-center text-h6">
            <span v-if="props.row.local_score !== null" class="text-weight-bold">{{ props.row.local_score }}</span>
            <span v-else class="text-grey-5">-</span>
            <span class="text-grey-5 q-mx-sm">:</span>
            <span v-if="props.row.visitor_score !== null" class="text-weight-bold">{{ props.row.visitor_score }}</span>
            <span v-else class="text-grey-5">-</span>
          </q-td>
          <q-td class="text-center">
            <div class="row items-center no-wrap">
              <img :src="`https://flagcdn.com/w40/${props.row.visitor_iso2}.png`" width="28" height="19" style="border-radius:2px;border:1px solid #ccc" class="q-mr-sm" />
              <span class="text-weight-medium">{{ props.row.visitor_team }}</span>
            </div>
          </q-td>
          <q-td class="text-center">
            <div class="text-caption text-weight-medium">{{ props.row.venue }}</div>
            <div class="text-caption text-grey-6">{{ props.row.city }}</div>
          </q-td>
          <q-td class="text-center">
            <q-badge :color="stageColor(props.row.stage)" text-color="white" :label="stageLabel(props.row.stage)" />
          </q-td>
        </q-tr>
      </template>
    </q-table>

    <div v-if="filteredMatches.length > pagination.rowsPerPage" class="row justify-center q-mt-md">
      <q-pagination
        v-model="pagination.page"
        :max="Math.ceil(filteredMatches.length / pagination.rowsPerPage)"
        :max-pages="7"
        boundary-numbers
      />
    </div>

    <div v-if="!selectedTeam && !loading" class="text-center q-pa-xl text-grey-6">
      <q-icon name="sports_soccer" size="4em" class="q-mb-md" />
      <div class="text-h6">Selecciona una seleccion para ver su calendario</div>
    </div>

    <div v-if="loading" class="text-center q-pa-xl">
      <q-spinner color="primary" size="3em" />
    </div>

    <div v-if="selectedTeam && filteredMatches.length === 0 && !loading" class="text-center q-pa-xl text-grey-6">
      <q-icon name="search_off" size="4em" class="q-mb-md" />
      <div class="text-h6">No hay partidos para esta seleccion</div>
    </div>
  </q-page>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import axios from 'axios'

const selectedTeam = ref(null)
const loading = ref(false)
const allTeams = ref([])
const allMatches = ref([])
const filteredTeamOptions = ref([])

const pagination = ref({
  page: 1,
  rowsPerPage: 10,
  sortBy: 'match_date',
  descending: false
})

const columns = [
  { name: 'match_number', label: '#', align: 'center', field: 'match_number' },
  { name: 'date', label: 'Fecha', align: 'center', field: 'match_date' },
  { name: 'time', label: 'Hora', align: 'center', field: 'match_date' },
  { name: 'local', label: 'Local', align: 'center', field: 'local_team' },
  { name: 'score', label: 'Marcador', align: 'center' },
  { name: 'visitor', label: 'Visitante', align: 'center', field: 'visitor_team' },
  { name: 'venue', label: 'Estadio / Ciudad', align: 'center', field: 'venue' },
  { name: 'stage', label: 'Fase', align: 'center', field: 'stage' }
]

const teamOptions = computed(() => {
  return allTeams.value.map(t => ({
    id: t.id,
    name: t.name,
    iso2: t.iso2,
    group: t.group_name
  }))
})

function filterTeams(val, update) {
  update(() => {
    const needle = val.toLowerCase()
    filteredTeamOptions.value = teamOptions.value.filter(
      t => t.name.toLowerCase().includes(needle) || t.group.toLowerCase().includes(needle)
    )
  })
}

watch(allTeams, () => {
  filteredTeamOptions.value = teamOptions.value
})

const filteredMatches = computed(() => {
  if (!selectedTeam.value) return []
  const iso = selectedTeam.value
  return allMatches.value.filter(m =>
    m.local_iso2 === iso || m.visitor_iso2 === iso
  )
})

const paginatedMatches = computed(() => {
  const start = (pagination.value.page - 1) * pagination.value.rowsPerPage
  const end = start + pagination.value.rowsPerPage
  return filteredMatches.value.slice(start, end)
})

watch(selectedTeam, () => {
  pagination.value.page = 1
})

function formatDate(dateStr) {
  if (!dateStr) return ''
  const parts = dateStr.split(' ')[0].split('-')
  return parts[2] + '/' + parts[1] + '/' + parts[0]
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const parts = dateStr.split(' ')[1].split(':')
  let h = parseInt(parts[0])
  const m = parts[1]
  const ampm = h >= 12 ? 'PM' : 'AM'
  h = h % 12 || 12
  return `${h}:${m} ${ampm}`
}

function stageLabel(stage) {
  const labels = {
    group: 'Grupos',
    r32: '16vos',
    r16: 'Octavos',
    qf: 'Cuartos',
    sf: 'Semis',
    f: 'Final',
    third: '3er Lugar'
  }
  return labels[stage] || stage
}

function stageColor(stage) {
  const colors = {
    group: 'primary',
    r32: 'secondary',
    r16: 'accent',
    qf: 'deep-orange',
    sf: 'purple',
    f: 'positive',
    third: 'grey'
  }
  return colors[stage] || 'grey'
}

onMounted(async () => {
  try {
    loading.value = true
    const res = await axios.get('http://localhost:8080/api/groups')
    const teams = []
    for (const g of Object.values(res.data)) {
      teams.push(...g)
    }
    allTeams.value = teams

    const matchRes = await axios.get('http://localhost:8080/api/matches')
    allMatches.value = matchRes.data
  } catch (e) {
    console.error('Error loading data:', e)
  } finally {
    loading.value = false
  }
})
</script>