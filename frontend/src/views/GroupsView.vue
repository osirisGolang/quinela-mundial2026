<template>
  <q-page class="q-pa-md">
    <h4 class="text-primary">Grupos - Fase de Grupos</h4>

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
        <div v-else>
          <q-table
            :rows="getStandings(groupName)"
            :columns="standingsColumns"
            row-key="id"
            flat
            bordered
            dense
            :pagination="{ rowsPerPage: 0 }"
            hide-pagination
            class="standings-table q-mb-lg"
          >
            <template #header="props">
              <q-tr class="bg-primary text-white">
                <q-th v-for="col in props.cols" :key="col.name" :props="props" class="text-center">
                  {{ col.label }}
                </q-th>
              </q-tr>
            </template>
            <template #body="props">
              <q-tr :props="props" :class="getRowClass(props.rowIndex)">
                <q-td key="position" :props="props" class="text-center">
                  <q-badge
                    :color="positionColor(props.rowIndex + 1)"
                    text-color="white"
                    :label="props.rowIndex + 1"
                    class="text-weight-bold"
                    style="min-width: 26px;"
                  />
                </q-td>
                <q-td key="team" :props="props">
                  <div class="row items-center no-wrap q-gutter-xs">
                    <img
                      :src="`https://flagcdn.com/w40/${props.row.iso2}.png`"
                      width="26" height="17"
                      style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;"
                    />
                    <span class="text-weight-medium">{{ props.row.name }}</span>
                    <q-badge v-if="props.row.is_host" color="amber-10" text-color="black" label="(H)" />
                  </div>
                </q-td>
                <q-td key="pj" :props="props" class="text-center">{{ getStats(props.row.id).pj }}</q-td>
                <q-td key="g" :props="props" class="text-center">{{ getStats(props.row.id).g }}</q-td>
                <q-td key="e" :props="props" class="text-center">{{ getStats(props.row.id).e }}</q-td>
                <q-td key="p" :props="props" class="text-center">{{ getStats(props.row.id).p }}</q-td>
                <q-td key="gf" :props="props" class="text-center">{{ getStats(props.row.id).gf }}</q-td>
                <q-td key="gc" :props="props" class="text-center">{{ getStats(props.row.id).gc }}</q-td>
                <q-td key="dg" :props="props" class="text-center">{{ getStats(props.row.id).dg }}</q-td>
                <q-td key="pts" :props="props" class="text-center text-weight-bold">{{ getStats(props.row.id).pts }}</q-td>
              </q-tr>
            </template>
          </q-table>

          <div class="text-subtitle1 text-weight-bold q-mb-sm text-primary">
            <q-icon name="sports_soccer" class="q-mr-xs" />
            Partidos del Grupo {{ groupName }}
          </div>

          <q-table
            :rows="getMatches(groupName)"
            :columns="matchColumns"
            row-key="id"
            flat
            bordered
            dense
            :pagination="{ rowsPerPage: 0 }"
            hide-pagination
          >
            <template #body="props">
              <q-tr>
                <q-td class="text-center">
                  <q-badge color="primary" :label="`#${props.row.match_number}`" />
                </q-td>
                <q-td>
                  <div class="row items-center no-wrap">
                    <img :src="`https://flagcdn.com/w40/${props.row.local_iso2}.png`" width="24" height="16" style="border-radius:2px;border:1px solid #ccc" class="q-mr-sm" />
                    <span class="text-weight-medium">{{ props.row.local_team }}</span>
                  </div>
                </q-td>
                <q-td class="text-center text-h6">
                  <span v-if="props.row.local_score !== null">{{ props.row.local_score }}</span>
                  <span v-else class="text-grey-5">-</span>
                  <span class="text-grey-5 mx-2">:</span>
                  <span v-if="props.row.visitor_score !== null">{{ props.row.visitor_score }}</span>
                  <span v-else class="text-grey-5">-</span>
                </q-td>
                <q-td>
                  <div class="row items-center no-wrap">
                    <span class="text-weight-medium">{{ props.row.visitor_team }}</span>
                    <img :src="`https://flagcdn.com/w40/${props.row.visitor_iso2}.png`" width="24" height="16" style="border-radius:2px;border:1px solid #ccc" class="q-ml-sm" />
                  </div>
                </q-td>
                <q-td>
                  <div class="text-caption">{{ formatDate(props.row.match_date) }}</div>
                  <div class="text-caption text-grey-6">{{ props.row.venue }}</div>
                </q-td>
              </q-tr>
            </template>
          </q-table>
        </div>
      </q-tab-panel>
    </q-tab-panels>
  </q-page>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'

const tab = ref('A')
const groups = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L']
const loading = ref(true)
const teamsData = ref({})
const matchesData = ref([])

const standingsColumns = [
  { name: 'position', label: '#', align: 'center', field: 'id' },
  { name: 'team', label: 'Seleccion', align: 'left', field: 'name' },
  { name: 'pj', label: 'PJ', align: 'center' },
  { name: 'g', label: 'G', align: 'center' },
  { name: 'e', label: 'E', align: 'center' },
  { name: 'p', label: 'P', align: 'center' },
  { name: 'gf', label: 'GF', align: 'center' },
  { name: 'gc', label: 'GC', align: 'center' },
  { name: 'dg', label: 'DG', align: 'center' },
  { name: 'pts', label: 'Pts', align: 'center' },
]

const matchColumns = [
  { name: 'match_number', label: '#', align: 'center' },
  { name: 'local', label: 'Local', align: 'left' },
  { name: 'score', label: 'Marcador', align: 'center' },
  { name: 'visitor', label: 'Visitante', align: 'left' },
  { name: 'venue', label: 'Fecha / Estadio', align: 'left' },
]

const groupIndex = { A: 1, B: 2, C: 3, D: 4, E: 5, F: 6, G: 7, H: 8, I: 9, J: 10, K: 11, L: 12 }

function getStandings(groupName) {
  return teamsData.value[groupName] || []
}

function getMatches(groupName) {
  const idx = groupIndex[groupName]
  return matchesData.value.filter(m => m.group_id === idx)
}

function getStats(teamId) {
  const teamMatches = matchesData.value.filter(
    m => m.local_team_id === teamId || m.visitor_team_id === teamId
  )
  let pj = 0, g = 0, e = 0, p = 0, gf = 0, gc = 0
  for (const m of teamMatches) {
    if (m.local_score === null || m.visitor_score === null) continue
    pj++
    const isLocal = m.local_team_id === teamId
    const gfT = isLocal ? m.local_score : m.visitor_score
    const gcT = isLocal ? m.visitor_score : m.local_score
    gf += gfT
    gc += gcT
    if (gfT > gcT) g++
    else if (gfT === gcT) e++
    else p++
  }
  return { pj, g, e, p, gf, gc, dg: gf - gc, pts: g * 3 + e }
}

function getRowClass(index) {
  if (index < 2) return 'bg-green-1'
  if (index === 2) return 'bg-amber-1'
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
  return `${parts[2]}/${parts[1]}/${parts[0]}`
}

onMounted(async () => {
  try {
    const res = await axios.get('http://localhost:8080/api/groups')
    teamsData.value = res.data
    const matchRes = await axios.get('http://localhost:8080/api/matches?stage=group')
    matchesData.value = matchRes.data
  } catch (e) {
    console.error('Error loading data:', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.mx-2 { margin-left: 0.5rem; margin-right: 0.5rem; }
</style>