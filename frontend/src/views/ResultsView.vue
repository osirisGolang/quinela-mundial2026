<template>
  <q-page class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div class="text-h5 text-primary">Resultados - Fase de Grupos</div>
      <q-btn
        color="primary"
        icon="refresh"
        label="Actualizar Tablas"
        @click="refreshAll"
        :loading="loading"
      />
    </div>

    <q-tabs
      v-model="tab"
      class="text-primary"
      active-color="primary"
      indicator-color="primary"
      align="justify"
      narrow-indicator
    >
      <q-tab name="A" label="Grupo A" />
      <q-tab name="B" label="Grupo B" />
      <q-tab name="C" label="Grupo C" />
      <q-tab name="D" label="Grupo D" />
      <q-tab name="E" label="Grupo E" />
      <q-tab name="F" label="Grupo F" />
      <q-tab name="G" label="Grupo G" />
      <q-tab name="H" label="Grupo H" />
      <q-tab name="I" label="Grupo I" />
      <q-tab name="J" label="Grupo J" />
      <q-tab name="K" label="Grupo K" />
      <q-tab name="L" label="Grupo L" />
    </q-tabs>

    <q-tab-panels v-model="tab" animated class="q-mt-md">
      <q-tab-panel
        v-for="groupName in groups"
        :key="groupName"
        :name="groupName"
      >
        <div v-if="loading" class="text-center q-pa-md">
          <q-spinner color="primary" size="3em" />
        </div>
        <div v-else class="row q-col-gutter-md">
          <div class="col-12 col-lg-7">
            <div class="text-subtitle1 text-weight-bold q-mb-sm text-primary">
              <q-icon name="sports_soccer" class="q-mr-xs" />
              Juegos del Grupo {{ groupName }}
            </div>
            <q-table
              :rows="getGroupMatches(groupName)"
              :columns="matchColumns"
              row-key="id"
              flat
              bordered
              dense
              :pagination="{ rowsPerPage: 0 }"
              hide-pagination
              class="matches-table q-mb-lg"
            >
              <template #header="props">
                <q-tr class="bg-primary text-white">
                  <q-th v-for="col in props.cols" :key="col.name" :props="props" class="text-center">
                    {{ col.label }}
                  </q-th>
                </q-tr>
              </template>
              <template #body="props">
                <q-tr :class="props.row.status === 'finished' ? 'bg-green-1' : ''">
                  <q-td class="text-center">
                    <q-badge color="primary" :label="`#${props.row.match_number}`" />
                  </q-td>
                  <q-td>
                    <div class="text-caption text-weight-medium">
                      <q-icon name="event" size="xs" class="q-mr-xs" />
                      {{ formatDate(props.row.match_date) }}
                    </div>
                    <div class="text-caption text-grey-7">
                      <q-icon name="schedule" size="xs" class="q-mr-xs" />
                      {{ formatTime(props.row.match_date) }}
                    </div>
                  </q-td>
                  <q-td>
                    <div class="row items-center no-wrap">
                      <img :src="`https://flagcdn.com/w40/${props.row.local_iso2}.png`" width="24" height="16" style="border-radius:2px;border:1px solid #ccc" class="q-mr-sm" />
                      <span class="text-weight-medium">{{ props.row.local_team }}</span>
                    </div>
                  </q-td>
                  <q-td class="text-center">
                    <div v-if="props.row.status === 'finished'" class="text-h6 text-weight-bold">
                      {{ props.row.local_score }}
                    </div>
                    <q-input
                      v-else
                      v-model.number="scores[props.row.id].local_score"
                      type="number"
                      min="0"
                      dense
                      outlined
                      style="width: 55px;"
                      @update:model-value="markDirty(props.row.id)"
                    />
                  </q-td>
                  <q-td class="text-center text-h5 text-grey-5">-</q-td>
                  <q-td class="text-center">
                    <div v-if="props.row.status === 'finished'" class="text-h6 text-weight-bold">
                      {{ props.row.visitor_score }}
                    </div>
                    <q-input
                      v-else
                      v-model.number="scores[props.row.id].visitor_score"
                      type="number"
                      min="0"
                      dense
                      outlined
                      style="width: 55px;"
                      @update:model-value="markDirty(props.row.id)"
                    />
                  </q-td>
                  <q-td>
                    <div class="row items-center no-wrap">
                      <span class="text-weight-medium">{{ props.row.visitor_team }}</span>
                      <img :src="`https://flagcdn.com/w40/${props.row.visitor_iso2}.png`" width="24" height="16" style="border-radius:2px;border:1px solid #ccc" class="q-ml-sm" />
                    </div>
                  </q-td>
                  <q-td>
                    <div class="text-caption">{{ props.row.venue }}</div>
                    <div class="text-caption text-grey-6">{{ props.row.city }}</div>
                  </q-td>
                  <q-td class="text-center">
                    <div v-if="props.row.status === 'finished'">
                      <q-icon name="check_circle" color="positive" size="sm" />
                    </div>
                    <q-btn
                      v-else-if="dirtyRows.has(props.row.id)"
                      color="positive"
                      icon="save"
                      size="sm"
                      round
                      dense
                      @click="saveResult(props.row)"
                      :loading="saving === props.row.id"
                    />
                    <div v-else>
                      <q-icon name="radio_button_unchecked" color="grey-4" size="sm" />
                    </div>
                  </q-td>
                </q-tr>
              </template>
            </q-table>
          </div>

          <div class="col-12 col-lg-5">
            <div class="text-subtitle1 text-weight-bold q-mb-sm text-primary">
              <q-icon name="leaderboard" class="q-mr-xs" />
              Tabla de Posiciones - Grupo {{ groupName }}
            </div>
            <q-table
              :rows="computedStandings[groupName] || []"
              :columns="standingsColumns"
              row-key="team_id"
              flat
              bordered
              dense
              :pagination="{ rowsPerPage: 0 }"
              hide-pagination
              class="standings-table"
            >
              <template #header="props">
                <q-tr class="bg-primary text-white">
                  <q-th v-for="col in props.cols" :key="col.name" :props="props" class="text-center">
                    {{ col.label }}
                  </q-th>
                </q-tr>
              </template>
              <template #body="props">
                <q-tr :class="getRowClass(props.row.position)">
                  <q-td class="text-center">
                    <q-badge
                      :color="positionColor(props.row.position)"
                      text-color="white"
                      :label="props.row.position"
                      class="text-weight-bold"
                      style="min-width: 26px;"
                    />
                  </q-td>
                  <q-td>
                    <div class="row items-center no-wrap q-gutter-xs">
                      <img :src="`https://flagcdn.com/w40/${props.row.iso2}.png`" width="26" height="17" style="border-radius:2px;border:1px solid #ccc" />
                      <span class="text-weight-medium">{{ props.row.team_name }}</span>
                      <q-badge v-if="props.row.is_host" color="amber-10" text-color="black" label="(H)" />
                    </div>
                  </q-td>
                  <q-td class="text-center">{{ props.row.pj }}</q-td>
                  <q-td class="text-center">{{ props.row.g }}</q-td>
                  <q-td class="text-center">{{ props.row.e }}</q-td>
                  <q-td class="text-center">{{ props.row.p }}</q-td>
                  <q-td class="text-center">{{ props.row.gf }}</q-td>
                  <q-td class="text-center">{{ props.row.gc }}</q-td>
                  <q-td class="text-center">{{ props.row.dg }}</q-td>
                  <q-td class="text-center text-weight-bold">{{ props.row.pts }}</q-td>
                </q-tr>
              </template>
            </q-table>
          </div>
        </div>
      </q-tab-panel>
    </q-tab-panels>
  </q-page>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import axios from 'axios'

const tab = ref('A')
const groups = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L']
const loading = ref(false)
const saving = ref(null)
const dirtyRows = ref(new Set())
const scores = reactive({})
const teamsData = ref({})
const matchesData = ref([])

const groupIndex = { A: 1, B: 2, C: 3, D: 4, E: 5, F: 6, G: 7, H: 8, I: 9, J: 10, K: 11, L: 12 }

const matchColumns = [
  { name: 'match_number', label: '#', align: 'center' },
  { name: 'datetime', label: 'Fecha/Hora', align: 'left' },
  { name: 'local', label: 'Local', align: 'left' },
  { name: 'local_score', label: 'GF', align: 'center' },
  { name: 'sep', label: '', align: 'center' },
  { name: 'visitor_score', label: 'GC', align: 'center' },
  { name: 'visitor', label: 'Visitante', align: 'left' },
  { name: 'venue', label: 'Estadio/Ciudad', align: 'left' },
  { name: 'action', label: '', align: 'center' }
]

const standingsColumns = [
  { name: 'position', label: '#', align: 'center' },
  { name: 'team', label: 'Seleccion', align: 'left' },
  { name: 'pj', label: 'PJ', align: 'center' },
  { name: 'g', label: 'G', align: 'center' },
  { name: 'e', label: 'E', align: 'center' },
  { name: 'p', label: 'P', align: 'center' },
  { name: 'gf', label: 'GF', align: 'center' },
  { name: 'gc', label: 'GC', align: 'center' },
  { name: 'dg', label: 'DG', align: 'center' },
  { name: 'pts', label: 'Pts', align: 'center' }
]

function getGroupMatches(groupName) {
  const idx = groupIndex[groupName]
  return matchesData.value.filter(m => m.group_id === idx)
}

const computedStandings = computed(() => {
  const result = {}
  for (const g of groups) {
    const idx = groupIndex[g]
    const teamList = teamsData.value[g] || []
    const teamStats = {}
    for (const t of teamList) {
      teamStats[t.id] = { pj: 0, g: 0, e: 0, p: 0, gf: 0, gc: 0 }
    }
    const groupMatches = matchesData.value.filter(m => m.group_id === idx)
    for (const m of groupMatches) {
      if (m.local_score === null || m.visitor_score === null) continue
      const ls = m.local_score, vs = m.visitor_score
      if (teamStats[m.local_team_id]) {
        teamStats[m.local_team_id].pj++
        teamStats[m.local_team_id].gf += ls
        teamStats[m.local_team_id].gc += vs
        if (ls > vs) teamStats[m.local_team_id].g++
        else if (ls === vs) teamStats[m.local_team_id].e++
        else teamStats[m.local_team_id].p++
      }
      if (teamStats[m.visitor_team_id]) {
        teamStats[m.visitor_team_id].pj++
        teamStats[m.visitor_team_id].gf += vs
        teamStats[m.visitor_team_id].gc += ls
        if (vs > ls) teamStats[m.visitor_team_id].g++
        else if (vs === ls) teamStats[m.visitor_team_id].e++
        else teamStats[m.visitor_team_id].p++
      }
    }
    const rows = teamList.map(t => ({
      team_id: t.id,
      team_name: t.name,
      iso2: t.iso2,
      is_host: t.is_host,
      pj: teamStats[t.id]?.pj || 0,
      g: teamStats[t.id]?.g || 0,
      e: teamStats[t.id]?.e || 0,
      p: teamStats[t.id]?.p || 0,
      gf: teamStats[t.id]?.gf || 0,
      gc: teamStats[t.id]?.gc || 0,
      dg: (teamStats[t.id]?.gf || 0) - (teamStats[t.id]?.gc || 0),
      pts: (teamStats[t.id]?.g || 0) * 3 + (teamStats[t.id]?.e || 0)
    }))
    rows.sort((a, b) => {
      if (b.pts !== a.pts) return b.pts - a.pts
      if (b.dg !== a.dg) return b.dg - a.dg
      if (b.gf !== a.gf) return b.gf - a.gf
      return a.team_name.localeCompare(b.team_name)
    })
    rows.forEach((r, i) => r.position = i + 1)
    result[g] = rows
  }
  return result
})

function markDirty(id) {
  dirtyRows.value.add(id)
}

async function saveResult(match) {
  const ls = scores[match.id]?.local_score
  const vs = scores[match.id]?.visitor_score
  if (ls === null || ls === undefined || vs === null || vs === undefined) return
  try {
    saving.value = match.id
    await axios.post('http://localhost:8080/api/results', {
      match_id: match.id,
      local_score: ls,
      visitor_score: vs
    })
    await refreshAll()
    dirtyRows.value.delete(match.id)
  } catch (e) {
    console.error('Error saving result:', e)
  } finally {
    saving.value = null
  }
}

async function refreshAll() {
  try {
    loading.value = true
    const [groupsRes, matchesRes] = await Promise.all([
      axios.get('http://localhost:8080/api/groups'),
      axios.get('http://localhost:8080/api/matches?stage=group')
    ])
    teamsData.value = groupsRes.data
    matchesData.value = matchesRes.data
    for (const m of matchesRes.data) {
      if (!scores[m.id]) {
        scores[m.id] = { local_score: m.local_score ?? '', visitor_score: m.visitor_score ?? '' }
      }
    }
  } catch (e) {
    console.error('Error refreshing:', e)
  } finally {
    loading.value = false
  }
}

function getRowClass(pos) {
  if (pos <= 2) return 'bg-green-1'
  if (pos === 3) return 'bg-amber-1'
  return ''
}

function positionColor(pos) {
  if (pos <= 2) return 'positive'
  if (pos === 3) return 'warning'
  return 'grey'
}

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

onMounted(() => {
  refreshAll()
})
</script>