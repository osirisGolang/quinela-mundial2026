# Plan de Desarrollo: Quinela Mundial 2026

## Tecnologías

- **Frontend**: Wails + Quasar 2 + Vue 3 (Composition API)
- **Backend**: Go (Gin / Echo)
- **BD**: SQLite (embebida vía Go)
- **UI/UX**: Quasar components + Vite

## Estructura del Proyecto

```
quinela-mundial2026/
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── internal/
│   │   ├── database/     # Migraciones, conexión SQLite
│   │   ├── models/       # Structs (Team, Match, Group, Prediction, User)
│   │   ├── handlers/     # HTTP handlers (API REST)
│   │   ├── services/     # Lógica de negocio
│   │   └── middleware/   # Auth, CORS, logging
│   └── data/
│       └── quinela.db
│
└── frontend/
    ├── wails.json
    ├── package.json
    ├── src/
    │   ├── App.vue
    │   ├── main.js
    │   ├── router/
    │   │   └── index.js
    │   ├── stores/
    │   │   └── quinela.js        # Pinia store
    │   ├── components/
    │   │   ├── GroupTable.vue      # Tabla de posiciones de grupo
    │   │   ├── MatchCard.vue       # Tarjeta de partido
    │   │   ├── KnockoutBracket.vue # Árbol de eliminatorias
    │   │   └── PredictionForm.vue  # Formulario de pronóstico
│   │   ├── views/
│   │   │   ├── GroupsView.vue      # Vista de fase de grupos (A-L)
│   │   │   ├── ScheduleView.vue    # Vista de horarios por país (paginada)
│   │   │   ├── ResultsInput.vue    # Vista unificada: juegos + tabla posiciones
│   │   │   ├── KnockoutView.vue    # Vista de eliminatorias
│   │   │   ├── StandingsView.vue   # Tabla de posiciones general
│   │   │   └── AdminView.vue       # Admin (cargar resultados)
    │   └── assets/
    └── public/
```

## Formato del Mundial 2026

- **48 selecciones**, 12 grupos (A–L) de 4 equipos cada uno
- Pasan a 16avos: **32 equipos** (1.° y 2.° de cada grupo + 8 mejores 3.°)
- 16avos → Octavos → Cuartos → Semis → Final
- **(H)** = País sede (Canadá, México, Estados Unidos)

---

## Fases de Desarrollo

### Fase 1 — Inicialización del proyecto
- Crear proyecto Wails: `wails init -t vue`
- Instalar Quasar en el frontend: `quasar add quasar`
- Configurar Vue Router y Pinia
- Inicializar módulo Go en `backend/` con `go mod init`
- Agregar dependencias Go (Gin, SQLite driver `modernc.org/sqlite`, JWT, bcrypt)
- Crear estructura de directorios
- Crear script SQL de migración `migrations/001_init.sql` con las 48 selecciones, 12 grupos, 72 partidos de fase de grupos y 16 partidos de 16avos predefinidos

### Fase 2 — Modelos de datos y BD
- Definir modelos Go:
  ```go
  type Team struct {
      ID        uint      `db:"id"`
      Name      string    `db:"name"`        // Nombre oficial en español
      ShortCode string    `db:"short_code"`   // Código 3 letras (MEX, RSA, etc.)
      GroupID   uint      `db:"group_id"`
      IsHost    bool      `db:"is_host"`
      FlagURL   string    `db:"flag_url"`     // URL de bandera
  }

  type Group struct {
      ID    uint   `db:"id"`
      Name  string `db:"name"` // "A", "B", ... "L"
  }

  type Match struct {
      ID           uint      `db:"id"`
      MatchNumber  int       `db:"match_number"`  // Número FIFA (1-104)
      Stage        string    `db:"stage"`         // "group", "r32", "r16", "qf", "sf", "f", "third"
      GroupID      *uint     `db:"group_id"`      // NULL para eliminatorias
      LocalTeamID  uint      `db:"local_team_id"`
      VisitorTeamID uint     `db:"visitor_team_id"`
      MatchDate    time.Time `db:"match_date"`
      Venue        string    `db:"venue"`
      City         string    `db:"city"`
      LocalScore   *int      `db:"local_score"`
      VisitorScore *int      `db:"visitor_score"`
      Status       string    `db:"status"` // "pending", "live", "finished"
  }

  type Prediction struct {
      ID            uint `db:"id"`
      UserID        uint `db:"user_id"`
      MatchID       uint `db:"match_id"`
      LocalScore    int  `db:"local_score"`
      VisitorScore  int  `db:"visitor_score"`
      Points        int  `db:"points"` // Calculado al cargar resultado
  }

  type User struct {
      ID           uint   `db:"id"`
      Username     string `db:"username"`
      PasswordHash string `db:"password_hash"`
      IsAdmin      bool   `db:"is_admin"`
  }

  type Standings struct {
      UserID      uint   `db:"user_id"`
      Username    string `db:"username"`
      TotalPoints int    `db:"total_points"`
      ExactScore  int    `db:"exact_score"`
      ResultOnly  int    `db:"result_only"`
  }
  ```
- Crear migraciones SQLite automáticas (al iniciar el programa)
- Poblar BD con datos FIFA reales (48 equipos, 72 partidos de grupos, 16 partidos de 16avos)

### Fase 3 — API REST (backend)
- `GET  /api/groups` — lista de grupos con sus 4 equipos + partidos (para tabs)
- `GET  /api/groups/:id/standings` — tabla de posiciones de un grupo calculada desde resultados
- `GET  /api/groups/:id/matches` — partidos de un grupo específico
- `GET  /api/matches?stage=group` — todos los partidos de fase de grupos
- `GET  /api/matches?stage=r32` — partidos de 16avos
- `GET  /api/matches?stage=r16` — partidos de octavos
- `GET  /api/matches?stage=qf` — cuartos de final
- `GET  /api/matches?stage=sf` — semifinales
- `GET  /api/matches?stage=f` — final
- `GET  /api/matches?stage=third` — tercer lugar
- `POST /api/predictions` — enviar pronóstico
- `GET  /api/predictions/:userId` — ver todos los pronósticos de un usuario
- `PUT  /api/predictions/:matchId` — actualizar pronóstico (si aún no inicia)
- `POST /api/results` — admin: guardar resultado (recalcula standings y puntos de usuarios automáticamente)
- `GET  /api/results/:matchId` — obtener resultado de un partido
- `GET  /api/standings` — tabla de posiciones de la quiniela (todos los usuarios)
- `GET  /api/standings/:userId` — puntaje detallado de un usuario
- `POST /api/auth/register` — registro de usuario
- `GET  /api/auth/login` — login (devuelve JWT)
- `GET  /api/schedule/:iso2` — devuelve los partidos de un equipo específico filtrado por código ISO-2
- `GET  /api/admin/users` — admin: lista de usuarios
- `DELETE /api/admin/users/:id` — admin: eliminar usuario

### Fase 4 — Frontend: Fase de Grupos (Quasar)
- **12 tabs** (uno por grupo A–L) usando `q-tabs` y `q-tab-panels`
- Cada tab muestra:
  - **Tabla de posiciones** (`GroupTable.vue`) calculada automáticamente desde los resultados de los partidos:
    - Columnas: `# Pos`, `Equipo` (con bandera), `PJ` (Partidos Jugados), `G` (Ganados), `E` (Empatados), `P` (Perdidos), `GF` (Goles a Favor), `GC` (Goles en Contra), `DG` (Diferencia de Goles), `Pts` (Puntos)
    - Ordenamiento: Pts DESC → DG DESC → GF DESC → Nombre ASC
    - Indicador **(H)** junto a países sede
    - Fila verde (`bg-green-1`) para 1.° y 2.° lugar
    - Fila amarilla (`bg-amber-1`) para repechaje (3.° lugar)
    - Fila gris claro para 4.° lugar
    - Al cargar resultados de partidos, la tabla se actualiza en tiempo real
  - **3 jornadas** con tarjetas de partidos (`q-card`) usando `MatchCard.vue`
    - Cada tarjeta muestra: fecha, hora, equipos con flags, marcador (si ya se jugó), o inputs numéricos (si es pronóstico)
- Componente `GroupTable.vue`: tabla de posiciones calculada dinámicamente desde `matches`
- Componente `MatchCard.vue`: muestra partido con:
  - Bandera + nombre local vs visitante
  - Fecha y hora (zona horaria local)
  - Estadio y ciudad
  - Marcador real o input de pronóstico
  - Indicador de partido jugado vs pendiente vs en vivo
- **Vista de Horarios por País** (`ScheduleView.vue`):
  - Selector `q-select` con las 48 selecciones (en español, con bandera)
  - Al seleccionar un país, muestra los partidos donde ese país es local o visitante
  - Tabla `q-table` con los partidos del país seleccionado:
    - Columnas: #, Fecha, Hora, Local, Visitante, Estadio, Ciudad, Fase
    - Paginación: **10 registros por página**
    - Selector de filas por página: `q-select` con opciones `[5, 10, 20, 50, 100]`
    - Navegación: `q-pagination` para avanzar/retroceder páginas
    - Ordenamiento por columna (fecha, hora)
  - Endpoint: `GET /api/schedule/:iso2` — devuelve los partidos de ese equipo
  - El selector muestra las 48 selecciones con nombre en español y código ISO
- Validación: no se puede modificar pronóstico después de `match_date`
- Filtro para ver solo partidos sin jugar / en juego / finalizados

### Fase 5 — Frontend: Eliminatorias
- Componente `KnockoutBracket.vue`: árbol visual SVG o flexbox
- Vista de 16avos: 8 partidos lado a lado
- Vista de octavos, cuartos, semis, final: progresión visual
- Pronóstico de eliminatorias: marcador + selección de quién avanza (input adicional)
- Para 16avos y octavos, los cruces muestran "1.° Grupo X" y "2.° Grupo Y" hasta que se determinen los clasificados; luego muestran el nombre real
- Bloqueo automático al iniciar cada ronda

### Fase 6 — Sistema de Puntaje
- **3 pts** — marcador exacto (score exacto)
- **1 pt** — resultado acertado (ganador o empate, sin marcador exacto)
- **0 pts** — resultado incorrecto
- **Bonus eliminatorias**: en 16avos, octavos y cuartos, el puntaje se multiplica x2
- Semifinal y final: puntaje se multiplica x3
- Tabla de posiciones general en `StandingsView.vue` con `q-table` ordenada por Pts

### Fase 7 — Autenticación y Multijugador
- Login/registro con `q-input` y `q-btn`
- JWT almacenado en localStorage
- Pinia store con estado de usuario autenticado
- Ver mis pronósticos vs resultados reales (overlay en cada partido)

### Fase 8 — Admin y Resultados

Panel admin protegido (ruta con guard) con una vista unificada que combina la **tabla de juegos por grupo** y la **tabla de posiciones**, ambas sincronizadas en tiempo real.

### Fase 9 — Vista Unificada: Juegos y Tabla de Posiciones

Vista principal del admin (`ResultsView.vue`) que contiene en una sola pantalla la **tabla de juegos por grupo** (con casillas editables para goles) y la **tabla de posiciones por grupo** (calculada automáticamente). Un botón **"Actualizar Tablas"** recalcula y muestra las estadísticas actualizadas.

#### Estructura de la Vista

```
┌─────────────────────────────────────────────────────────┐
│  Resultados — Fase de Grupos                             │
│  [Grupo A ▼]                      [🔄 Actualizar Tablas]  │
├─────────────────────────────────────────────────────────┤
│  ● Tabla de Juegos del Grupo                            │
│  ┌────┬──────────────────┬────┬──────────┬─────────┐     │
│  │ #  │ Equipo Local     │ GF │ GE │ GF │ Equipo  │Fecha│     │
│  │    │ ⚑ name          │[  ]│ - │[  ]│ Visit.  │     │     │
│  └────┴──────────────────┴────┴──────────┴─────────┘     │
│  ● Tabla de Posiciones del Grupo                        │
│  ┌────┬──────────┬────┬──┬──┬──┬────┬────┬────┬────┐     │
│  │Pos.│Selección │ PJ │ G│ E│ P│ GF │ GC │ DG │Pts│     │
│  │ 1  │⚑ México │ 2  │1 │0 │1 │ 3  │ 2  │ +1 │ 3 │     │
│  └────┴──────────┴────┴──┴──┴──┴────┴────┴────┴────┘     │
│  [Grupo B] [Grupo C] ... [Grupo L]                     │
└─────────────────────────────────────────────────────────┘
```

#### Componente `ResultsView.vue` — Vista unificada de juegos y posiciones

```vue
<template>
  <q-page padding>
    <div class="row items-center justify-between q-mb-md">
      <div class="text-h5">Resultados — Fase de Grupos</div>
      <div class="row q-gutter-sm items-center">
        <q-select
          v-model="selectedGroup"
          :options="groupOptions"
          option-value="id"
          option-label="name"
          emit-value
          map-options
          label="Grupo"
          dense
          outlined
          style="min-width: 120px;"
          @update:model-value="loadGroupData"
        />
        <q-btn
          color="primary"
          icon="refresh"
          label="Actualizar Tablas"
          @click="refreshAll"
          :loading="loading"
        >
          <q-tooltip>Recalcular tabla de posiciones con los resultados cargados</q-tooltip>
        </q-btn>
      </div>
    </div>

    <!-- Selector de tabs por grupo -->
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
        v-for="group in groups"
        :key="group.id"
        :name="group.name"
        style="padding: 0;"
      >
        <div class="q-pa-md">
          <!-- ── Tabla de Juegos del Grupo ── -->
          <div class="text-subtitle1 text-weight-bold q-mb-sm text-primary">
            <q-icon name="sports_soccer" class="q-mr-xs" />
            Juegos del Grupo {{ group.name }}
          </div>

          <q-table
            :rows="group.matches"
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
              <q-tr class="bg-primary text-white text-weight-bold">
                <q-th v-for="col in props.cols" :key="col.name" :props="props">
                  {{ col.label }}
                </q-th>
              </q-tr>
            </template>

            <template #body="props">
              <q-tr :props="props" :class="props.row.status === 'finished' ? 'bg-green-1' : ''">
                <!-- # Partido -->
                <q-td key="match_number" :props="props" class="text-center">
                  <q-badge color="primary" :label="`#${props.row.match_number}`" />
                </q-td>

                <!-- Fecha y Hora -->
                <q-td key="datetime" :props="props">
                  <div class="text-caption text-weight-medium">
                    <q-icon name="event" size="xs" class="q-mr-xs" />
                    {{ formatDate(props.row.match_date) }}
                  </div>
                  <div class="text-caption text-grey-7">
                    <q-icon name="schedule" size="xs" class="q-mr-xs" />
                    {{ formatTime(props.row.match_date) }}
                  </div>
                </q-td>

                <!-- Equipo Local -->
                <q-td key="local" :props="props">
                  <div class="row items-center no-wrap">
                    <img
                      :src="`https://flagcdn.com/w40/${props.row.local_iso2}.png`"
                      width="24" height="16"
                      style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;"
                      class="q-mr-sm"
                    />
                    <span class="text-weight-medium">{{ props.row.local_team }}</span>
                  </div>
                </q-td>

                <!-- Goles Local (input editable si no está finalizado) -->
                <q-td key="local_score" :props="props" class="text-center">
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
                    :disable="saving === props.row.id"
                    @update:model-value="markDirty(props.row.id)"
                  />
                </q-td>

                <!-- Separador -->
                <q-td key="sep" :props="props" class="text-center text-h5 text-grey-5">
                  —
                </q-td>

                <!-- Goles Visitante (input editable si no está finalizado) -->
                <q-td key="visitor_score" :props="props" class="text-center">
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
                    :disable="saving === props.row.id"
                    @update:model-value="markDirty(props.row.id)"
                  />
                </q-td>

                <!-- Equipo Visitante -->
                <q-td key="visitor" :props="props">
                  <div class="row items-center no-wrap">
                    <span class="text-weight-medium">{{ props.row.visitor_team }}</span>
                    <img
                      :src="`https://flagcdn.com/w40/${props.row.visitor_iso2}.png`"
                      width="24" height="16"
                      style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;"
                      class="q-ml-sm"
                    />
                  </div>
                </q-td>

                <!-- Estadio / Ciudad -->
                <q-td key="venue" :props="props">
                  <div class="text-caption">{{ props.row.venue }}</div>
                  <div class="text-caption text-grey-6">{{ props.row.city }}</div>
                </q-td>

                <!-- Estado / Botón guardar -->
                <q-td key="action" :props="props" class="text-center">
                  <div v-if="props.row.status === 'finished'" class="text-positive">
                    <q-icon name="check_circle" size="sm" />
                    <q-tooltip>Partido finalizado</q-tooltip>
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
                  >
                    <q-tooltip>Guardar resultado</q-tooltip>
                  </q-btn>
                  <div v-else class="text-grey-4">
                    <q-icon name="radio_button_unchecked" size="sm" />
                  </div>
                </q-td>
              </q-tr>
            </template>
          </q-table>

          <!-- ── Tabla de Posiciones del Grupo ── -->
          <div class="text-subtitle1 text-weight-bold q-mb-sm q-mt-md text-primary">
            <q-icon name="leaderboard" class="q-mr-xs" />
            Tabla de Posiciones — Grupo {{ group.name }}
          </div>

          <q-table
            :rows="groupStandings[group.id] || []"
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
              <q-tr class="bg-primary text-white text-weight-bold">
                <q-th v-for="col in props.cols" :key="col.name" :props="props" class="text-center">
                  {{ col.label }}
                </q-th>
              </q-tr>
            </template>

            <template #body="props">
              <q-tr :props="props" :class="getRowClass(props.row.position)">
                <!-- Posición -->
                <q-td key="position" :props="props" class="text-center">
                  <q-badge
                    :color="positionColor(props.row.position)"
                    text-color="white"
                    :label="props.row.position"
                    class="text-weight-bold"
                    style="min-width: 26px;"
                  />
                </q-td>

                <!-- Selección -->
                <q-td key="team" :props="props">
                  <div class="row items-center no-wrap q-gutter-xs">
                    <img
                      :src="`https://flagcdn.com/w40/${props.row.iso2}.png`"
                      width="26" height="17"
                      style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;"
                    />
                    <span class="text-weight-medium">{{ props.row.team_name }}</span>
                    <q-badge v-if="props.row.is_host" color="amber-10" text-color="black" label="(H)" />
                  </div>
                </q-td>

                <!-- PJ -->
                <q-td key="played" :props="props" class="text-center">
                  <q-badge color="grey-7" :label="props.row.played" />
                </q-td>

                <!-- G -->
                <q-td key="won" :props="props" class="text-center">
                  <q-badge color="positive" :label="props.row.won" />
                </q-td>

                <!-- E -->
                <q-td key="drawn" :props="props" class="text-center">
                  <q-badge color="warning" text-color="black" :label="props.row.drawn" />
                </q-td>

                <!-- P -->
                <q-td key="lost" :props="props" class="text-center">
                  <q-badge color="negative" :label="props.row.lost" />
                </q-td>

                <!-- GF -->
                <q-td key="gf" :props="props" class="text-center text-weight-medium">
                  {{ props.row.goals_for }}
                </q-td>

                <!-- GC -->
                <q-td key="ga" :props="props" class="text-center">
                  {{ props.row.goals_against }}
                </q-td>

                <!-- DG -->
                <q-td key="gd" :props="props" class="text-center">
                  <span :class="{
                    'text-positive': props.row.goal_diff > 0,
                    'text-negative': props.row.goal_diff < 0,
                    'text-grey': props.row.goal_diff === 0
                  }">
                    {{ props.row.goal_diff > 0 ? '+' : '' }}{{ props.row.goal_diff }}
                  </span>
                </q-td>

                <!-- Pts -->
                <q-td key="pts" :props="props" class="text-center">
                  <q-badge
                    color="primary"
                    text-color="white"
                    :label="props.row.points"
                    class="text-weight-bold"
                    style="min-width: 32px; font-size: 14px;"
                  />
                </q-td>
              </q-tr>
            </template>
          </q-table>

          <!-- Leyenda de colores -->
          <div class="row q-mt-sm q-gutter-x-md text-caption">
            <div class="row items-center q-gutter-xs">
              <div style="width: 16px; height: 16px; background: #e8f5e9; border: 1px solid #ccc; border-radius: 2px;"></div>
              <span>Clasifica a octavos (1.° y 2.°)</span>
            </div>
            <div class="row items-center q-gutter-xs">
              <div style="width: 16px; height: 16px; background: #fff8e1; border: 1px solid #ccc; border-radius: 2px;"></div>
              <span>Repechaje (3.°)</span>
            </div>
            <div class="row items-center q-gutter-xs">
              <div class="text-positive"><q-icon name="check_circle" size="sm" /> Finalizado</div>
            </div>
            <div class="text-grey-5 row items-center q-gutter-xs">
              <q-icon name="radio_button_unchecked" size="sm" /> Pendiente
            </div>
          </div>
        </div>
      </q-tab-panel>
    </q-tab-panels>

    <!-- Botones globales -->
    <div class="row justify-between items-center q-mt-lg q-pa-md bg-grey-2 rounded-borders">
      <div class="text-caption text-grey-7">
        <q-icon name="info" size="xs" />
        {{ totalDirty }} partido(s) con resultados pendientes de guardar
      </div>
      <div class="row q-gutter-sm">
        <q-btn
          color="primary"
          icon="refresh"
          label="Actualizar Tablas"
          @click="refreshAll"
          :loading="loading"
        />
        <q-btn
          color="positive"
          icon="save"
          label="Guardar Todos"
          @click="saveAll"
          :loading="savingAll"
          :disable="totalDirty === 0"
        />
        <q-btn
          color="grey-7"
          icon="undo"
          label="Cancelar Cambios"
          flat
          @click="cancelChanges"
          :disable="totalDirty === 0"
        />
      </div>
    </div>

    <!-- Notificación -->
    <q-banner v-if="totalDirty > 0" class="q-mt-sm" type="warning" inline-actions>
      <template #avatar>
        <q-icon name="edit" color="warning" />
      </template>
      Hay {{ totalDirty }} resultado(s) pendiente(s). Haz clic en "Guardar" o "Actualizar Tablas" para confirmar.
    </q-banner>
  </q-page>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { api } from 'src/boot/axios'
import { useQuasar } from 'quasar'

const $q = useQuasar()
const tab = ref('A')
const groups = ref([])
const groupStandings = reactive({})
const scores = reactive({})
const dirtyRows = ref(new Set())
const saving = ref(null)
const savingAll = ref(false)
const loading = ref(false)
const selectedGroup = ref(null)

const groupOptions = computed(() =>
  groups.value.map(g => ({ id: g.id, name: g.name }))
)

const totalDirty = computed(() => dirtyRows.value.size)

const matchColumns = [
  { name: 'match_number', label: '#', field: 'match_number', align: 'center', style: 'width: 50px;' },
  { name: 'datetime', label: 'Fecha / Hora', field: 'match_date', align: 'left' },
  { name: 'local', label: 'Equipo Local', field: 'local_team', align: 'left', sortable: true },
  { name: 'local_score', label: 'Goles', field: 'id', align: 'center', style: 'width: 70px;' },
  { name: 'sep', label: '', field: 'id', align: 'center', style: 'width: 20px;' },
  { name: 'visitor_score', label: 'Goles', field: 'id', align: 'center', style: 'width: 70px;' },
  { name: 'visitor', label: 'Equipo Visitante', field: 'visitor_team', align: 'left', sortable: true },
  { name: 'venue', label: 'Estadio / Ciudad', field: 'venue', align: 'left' },
  { name: 'action', label: 'Estado', field: 'id', align: 'center', style: 'width: 60px;' },
]

const standingsColumns = [
  { name: 'position', label: '#', field: 'position', align: 'center', style: 'width: 40px;' },
  { name: 'team', label: 'Selección', field: 'team_name', align: 'left', sortable: true },
  { name: 'played', label: 'PJ', field: 'played', align: 'center', style: 'width: 40px;' },
  { name: 'won', label: 'G', field: 'won', align: 'center', style: 'width: 35px;' },
  { name: 'drawn', label: 'E', field: 'drawn', align: 'center', style: 'width: 35px;' },
  { name: 'lost', label: 'P', field: 'lost', align: 'center', style: 'width: 35px;' },
  { name: 'gf', label: 'GF', field: 'goals_for', align: 'center', style: 'width: 40px;' },
  { name: 'ga', label: 'label', field: 'goals_against', align: 'center', style: 'width: 40px;' },
  { name: 'gd', label: 'DG', field: 'goal_diff', align: 'center', style: 'width: 45px;' },
  { name: 'pts', label: 'Pts', field: 'points', align: 'center', sortable: true, style: 'width: 45px;' },
]

onMounted(async () => {
  await loadAllGroups()
})

const loadAllGroups = async () => {
  loading.value = true
  const { data } = await api.get('/api/groups')
  groups.value = data
  selectedGroup.value = data[0]?.id

  data.forEach(g => {
    scores[g.id] = {}
    g.matches.forEach(m => {
      scores[m.id] = {
        local_score: m.local_score ?? 0,
        visitor_score: m.visitor_score ?? 0,
      }
    })
  })

  await calculateAllStandings()
  loading.value = false
}

const loadGroupData = async (groupId) => {
  // Already loaded in loadAllGroups, just recalculate
  await calculateAllStandings()
}

const calculateAllStandings = async () => {
  const { data } = await api.get('/api/groups')
  groups.value = data

  data.forEach(g => {
    const stats = calculateGroupStandings(g.matches, g.teams)
    groupStandings[g.id] = stats
  })
}

const calculateGroupStandings = (matches, teams) => {
  const stats = {}
  teams.forEach(t => {
    stats[t.id] = {
      team_id: t.id,
      team_name: t.name,
      iso2: t.iso2,
      is_host: t.is_host,
      played: 0, won: 0, drawn: 0, lost: 0,
      goals_for: 0, goals_against: 0,
      goal_diff: 0, points: 0,
    }
  })

  matches
    .filter(m => m.status === 'finished' && m.local_score !== null)
    .forEach(m => {
      const loc = stats[m.local_team_id]
      const vis = stats[m.visitor_team_id]
      if (!loc || !vis) return

      loc.played++
      vis.played++
      loc.goals_for += m.local_score
      loc.goals_against += m.visitor_score
      vis.goals_for += m.visitor_score
      vis.goals_against += m.local_score

      if (m.local_score > m.visitor_score) {
        loc.won++; loc.points += 3; vis.lost++
      } else if (m.local_score < m.visitor_score) {
        vis.won++; vis.points += 3; loc.lost++
      } else {
        loc.drawn++; vis.drawn++; loc.points++; vis.points++
      }
    })

  Object.values(stats).forEach(s => {
    s.goal_diff = s.goals_for - s.goals_against
  })

  const sorted = Object.values(stats).sort((a, b) => {
    if (b.points !== a.points) return b.points - a.points
    if (b.goal_diff !== a.goal_diff) return b.goal_diff - a.goal_diff
    if (b.goals_for !== a.goals_for) return b.goals_for - a.goals_for
    return a.team_name.localeCompare(b.team_name)
  })

  sorted.forEach((s, i) => { s.position = i + 1 })
  return sorted
}

const markDirty = (matchId) => {
  dirtyRows.value = new Set([...dirtyRows.value, matchId])
}

const saveResult = async (match) => {
  saving.value = match.id
  try {
    await api.post('/api/results', {
      match_id: match.id,
      local_score: scores[match.id].local_score,
      visitor_score: scores[match.id].visitor_score,
    })
    match.status = 'finished'
    match.local_score = scores[match.id].local_score
    match.visitor_score = scores[match.id].visitor_score
    dirtyRows.value = new Set([...dirtyRows.value].filter(id => id !== match.id))
    await calculateAllStandings()
    $q.notify({ type: 'positive', message: 'Resultado guardado y tablas actualizadas', icon: 'check' })
  } catch (e) {
    $q.notify({ type: 'negative', message: 'Error al guardar', icon: 'error' })
  } finally {
    saving.value = null
  }
}

const saveAll = async () => {
  savingAll.value = true
  const promises = [...dirtyRows.value].map(async (matchId) => {
    const match = groups.value.flatMap(g => g.matches).find(m => m.id === matchId)
    if (match) await saveResult(match)
  })
  await Promise.all(promises)
  savingAll.value = false
}

const refreshAll = async () => {
  loading.value = true
  await calculateAllStandings()
  $q.notify({ type: 'info', message: 'Tablas de posiciones actualizadas', icon: 'refresh' })
  loading.value = false
}

const cancelChanges = () => {
  groups.value.forEach(g => g.matches.forEach(m => {
    scores[m.id] = {
      local_score: m.local_score ?? 0,
      visitor_score: m.visitor_score ?? 0,
    }
  }))
  dirtyRows.value = new Set()
}

const formatDate = (d) => new Date(d).toLocaleDateString('es-ES', {
  weekday: 'short', day: 'numeric', month: 'short', year: 'numeric'
})
const formatTime = (d) => new Date(d).toLocaleTimeString('es-ES', {
  hour: '2-digit', minute: '2-digit'
})

const positionColor = (pos) => pos <= 2 ? 'positive' : pos === 3 ? 'warning' : 'grey-6'
const getRowClass = (pos) => pos <= 2 ? 'bg-green-1' : pos === 3 ? 'bg-amber-1' : 'bg-grey-2'
</script>

<style scoped>
.rounded-borders { border-radius: 8px; }
.matches-table .q-table__top { display: none; }
.standings-table .q-table__top { display: none; }
</style>
```

#### Flujo de datos completo

```
1. Admin abre vista ResultsView.vue
2. Se cargan todos los grupos con partidos y equipos
3. scores{} se inicializa con los marcadores actuales (o 0 si no hay)
4. calculateAllStandings() calcula tabla de posiciones con resultados existentes
5. Admin escribe goles en los inputs → markDirty()
6. Admin hace clic en:
   a) 💾 Guardar (individual) → POST /api/results → recalcula standings localmente
   b) 💾 Guardar Todos → guarda todos los pendientes → recalcula standings
   c) 🔄 Actualizar Tablas → recalcula standings con datos actuales de la BD
7. Tabla de posiciones se actualiza mostrando: PJ, G, E, P, GF, GC, DG, Pts
8. Fila verde para 1.°-2.°, fila amarilla para 3.°, fila gris para 4.°
```

#### Endpoint `POST /api/results` — Recalcula standings automáticamente

```go
func SaveResult(c *gin.Context) {
    var input struct {
        MatchID       uint `json:"match_id" binding:"required"`
        LocalScore    int  `json:"local_score" binding:"min=0"`
        VisitorScore  int  `json:"visitor_score" binding:"min=0"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 1. Actualizar partido
    _, err := db.Exec(`
        UPDATE matches
        SET local_score = ?, visitor_score = ?, status = 'finished'
        WHERE id = ?
    `, input.LocalScore, input.VisitorScore, input.MatchID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Error updating match"})
        return
    }

    // 2. Recalcular standings del grupo (para devolverlos actualizados)
    groupStandings := recalculateGroupStandings(input.MatchID)

    // 3. Recalcular puntos de TODOS los usuarios para este partido
    recalculateUserPoints(input.MatchID)

    c.JSON(200, gin.H{
        "message":  "ok",
        "standings": groupStandings,
    })
}

func recalculateGroupStandings(matchID uint) []map[string]interface{} {
    // Obtener match → groupID → teams → all finished matches of group
    // Calcular PJ, G, E, P, GF, GC, DG, Pts
    // Ordenar por Pts → DG → GF → Nombre
    // Devolver array ordenado con posición
}

func recalculateUserPoints(matchID uint) {
    // Para cada predicción de ese partido:
    //   - Comparar con resultado real
    //   - Si score exacto → 3 pts
    //   - Si solo resultado correcto → 1 pt
    //   - Si incorrecto → 0 pts
    //   - Multiplicadores para eliminatorias
    //   - Actualizar tabla predictions.points
}
```

> El componente `ResultsInput.vue` fue reemplazado por la vista unificada `ResultsView.vue` que combina tabla de juegos + tabla de posiciones en una sola pantalla.
          <template #header="props">
            <q-tr :props="props" class="bg-grey-2 text-weight-bold">
              <q-th v-for="col in props.cols" :key="col.name" :props="props">
                {{ col.label }}
              </q-th>
            </q-tr>
          </template>

          <template #body="props">
            <q-tr :props="props">
              <!-- # Partido -->
              <q-td key="match_number" :props="props" class="text-center">
                <q-badge color="primary" :label="props.row.match_number" />
              </q-td>

              <!-- Fecha -->
              <q-td key="date" :props="props">
                <div class="text-caption">
                  <q-icon name="event" size="xs" class="q-mr-xs" />
                  {{ formatDate(props.row.match_date) }}
                </div>
                <div class="text-caption text-grey-7">
                  <q-icon name="schedule" size="xs" class="q-mr-xs" />
                  {{ formatTime(props.row.match_date) }}
                </div>
              </q-td>

              <!-- Equipo Local -->
              <q-td key="local" :props="props">
                <div class="row items-center no-wrap">
                  <img
                    :src="`https://flagcdn.com/w40/${props.row.local_iso2}.png`"
                    width="28" height="18"
                    style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;"
                    class="q-mr-sm"
                  />
                  <span>{{ props.row.local_team }}</span>
                </div>
              </q-td>

              <!-- Goles Local -->
              <q-td key="local_score" :props="props" class="text-center">
                <q-input
                  v-model.number="results[props.row.id].local_score"
                  type="number"
                  min="0"
                  dense
                  outlined
                  style="width: 60px;"
                  :disable="props.row.status === 'finished'"
                  @update:model-value="markDirty(props.row.id)"
                />
              </q-td>

              <!-- Separador -->
              <q-td key="separator" :props="props" class="text-center text-h6 text-grey-5">
                -
              </q-td>

              <!-- Goles Visitante -->
              <q-td key="visitor_score" :props="props" class="text-center">
                <q-input
                  v-model.number="results[props.row.id].visitor_score"
                  type="number"
                  min="0"
                  dense
                  outlined
                  style="width: 60px;"
                  :disable="props.row.status === 'finished'"
                  @update:model-value="markDirty(props.row.id)"
                />
              </q-td>

              <!-- Equipo Visitante -->
              <q-td key="visitor" :props="props">
                <div class="row items-center no-wrap">
                  <span>{{ props.row.visitor_team }}</span>
                  <img
                    :src="`https://flagcdn.com/w40/${props.row.visitor_iso2}.png`"
                    width="28" height="18"
                    style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;"
                    class="q-ml-sm"
                  />
                </div>
              </q-td>

              <!-- Estadio -->
              <q-td key="venue" :props="props">
                <div class="text-caption">{{ props.row.venue }}</div>
                <div class="text-caption text-grey-7">{{ props.row.city }}</div>
              </q-td>

              <!-- Estado / Guardar -->
              <q-td key="actions" :props="props" class="text-center">
                <div v-if="props.row.status === 'finished'" class="text-positive">
                  <q-icon name="check_circle" size="sm" />
                  <span class="text-caption q-ml-xs">Guardado</span>
                </div>
                <div v-else-if="dirtyRows.has(props.row.id)">
                  <q-btn
                    color="positive"
                    icon="save"
                    size="sm"
                    dense
                    @click="saveResult(props.row)"
                    :loading="saving === props.row.id"
                  >
                    <q-tooltip>Guardar resultado</q-tooltip>
                  </q-btn>
                </div>
                <div v-else class="text-grey-5">
                  <q-icon name="pending" size="sm" />
                </div>
              </q-td>
            </q-tr>
          </template>
        </q-table>
      </q-tab-panel>
    </q-tab-panels>

    <!-- Botón guardar todos -->
    <div class="row justify-end q-mt-lg q-gutter-md">
      <q-btn
        color="primary"
        label="Guardar Todos"
        icon="save"
        @click="saveAll"
        :loading="savingAll"
        :disable="dirtyRows.size === 0"
      />
      <q-btn
        color="grey-7"
        label="Cancelar"
        flat
        @click="resetAll"
        :disable="dirtyRows.size === 0"
      />
    </div>

    <!-- Resumen de cambios -->
    <q-banner v-if="dirtyRows.size > 0" class="q-mt-md" type="warning" inline-actions>
      <template #avatar>
        <q-icon name="edit" color="warning" />
      </template>
      {{ dirtyRows.size }} partido(s) con resultados pendientes de guardar
    </q-banner>
  </q-page>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { api } from 'src/boot/axios'
import { useQuasar } from 'quasar'

const $q = useQuasar()
const tab = ref('A')
const groups = ref([])
const results = reactive({})
const dirtyRows = ref(new Set())
const saving = ref(null)
const savingAll = ref(false)

const columns = [
  { name: 'match_number', label: '#', field: 'match_number', align: 'center' },
  { name: 'date', label: 'Fecha / Hora', field: 'match_date', align: 'left' },
  { name: 'local', label: 'Local', field: 'local_team', align: 'left' },
  { name: 'local_score', label: 'Goles', field: 'id', align: 'center' },
  { name: 'separator', label: '', field: 'id', align: 'center' },
  { name: 'visitor_score', label: 'Goles', field: 'id', align: 'center' },
  { name: 'visitor', label: 'Visitante', field: 'visitor_team', align: 'left' },
  { name: 'venue', label: 'Estadio / Ciudad', field: 'venue', align: 'left' },
  { name: 'actions', label: 'Acción', field: 'id', align: 'center' },
]

onMounted(async () => {
  const { data } = await api.get('/api/groups')
  groups.value = data

  // Inicializar resultados con valores actuales o 0
  data.forEach(g => g.matches.forEach(m => {
    results[m.id] = {
      local_score: m.local_score ?? 0,
      visitor_score: m.visitor_score ?? 0,
    }
  }))
})

const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleDateString('es-ES', {
    weekday: 'short', day: 'numeric', month: 'short', year: 'numeric'
  })
}

const formatTime = (dateStr) => {
  return new Date(dateStr).toLocaleTimeString('es-ES', {
    hour: '2-digit', minute: '2-digit'
  })
}

const markDirty = (matchId) => {
  dirtyRows.value = new Set([...dirtyRows.value, matchId])
}

const saveResult = async (match) => {
  saving.value = match.id
  try {
    await api.post('/api/results', {
      match_id: match.id,
      local_score: results[match.id].local_score,
      visitor_score: results[match.id].visitor_score,
    })
    match.status = 'finished'
    match.local_score = results[match.id].local_score
    match.visitor_score = results[match.id].visitor_score
    dirtyRows.value = new Set([...dirtyRows.value].filter(id => id !== match.id))
    $q.notify({ type: 'positive', message: 'Resultado guardado', icon: 'check' })
  } catch (e) {
    $q.notify({ type: 'negative', message: 'Error al guardar', icon: 'error' })
  } finally {
    saving.value = null
  }
}

const saveAll = async () => {
  savingAll.value = true
  const promises = [...dirtyRows.value].map(async (matchId) => {
    const match = groups.value.flatMap(g => g.matches).find(m => m.id === matchId)
    if (match) await saveResult(match)
  })
  await Promise.all(promises)
  savingAll.value = false
}

const resetAll = () => {
  dirtyRows.value = new Set()
  groups.value.forEach(g => g.matches.forEach(m => {
    results[m.id] = {
      local_score: m.local_score ?? 0,
      visitor_score: m.visitor_score ?? 0,
    }
  }))
}
</script>

<style scoped>
.rounded-borders {
  border-radius: 8px;
}
</style>
```

#### Características de la tabla de resultados:

- **12 tabs** (uno por grupo A-L), igual que la vista de grupos
- **Tabla editable** por grupo con:
  - `#` — número de partido (badge azul)
  - **Fecha y hora** — icono de calendario + reloj
  - **Equipo Local** — bandera + nombre
  - **Goles Local** — `q-input type="number"` editable (min=0)
  - **Separador** — guión central
  - **Goles Visitante** — `q-input type="number"` editable
  - **Equipo Visitante** — nombre + bandera
  - **Estadio / Ciudad** — nombre del estadio y ciudad
  - **Acción** — botón guardar (solo aparece si hay cambios pendientes)
- **Estado visual**: icono verde con "Guardado" cuando ya fue guardado, icono pendiente cuando no
- **Notificaciones** Quasar (`$q.notify`) al guardar o al haber errores
- **Botón "Guardar Todos"** para guardar todos los cambios pendientes en una sola operación
- Los campos se deshabilitan automáticamente cuando `status === 'finished'`
- Contador de cambios pendientes con `q-banner` de advertencia
- **Botón "Cancelar"** para revertir cambios no guardados

#### Endpoint `POST /api/results` — Handler Go:

```go
func SaveResult(c *gin.Context) {
    var input struct {
        MatchID       uint `json:"match_id" binding:"required"`
        LocalScore    int  `json:"local_score" binding:"min=0"`
        VisitorScore int  `json:"visitor_score" binding:"min=0"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 1. Actualizar marcador del partido
    _, err := db.Exec(`
        UPDATE matches
        SET local_score = ?, visitor_score = ?, status = 'finished'
        WHERE id = ?
    `, input.LocalScore, input.VisitorScore, input.MatchID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Error updating match"})
        return
    }

    // 2. Recalcular puntos de TODOS los usuarios para este partido
    recalculatePointsForMatch(input.MatchID)

    c.JSON(200, gin.H{"message": "ok"})
}

func recalculatePointsForMatch(matchID uint) {
    // Obtener partido
    // Obtener predicciones de ese partido
    // Para cada predicción, comparar con resultado real y asignar puntos
    // 3 pts marcador exacto, 1 pt solo resultado, 0 pts incorrecto
    // Multiplicadores para eliminatorias
}
```

---

## Grupos del Mundial 2026 (datos oficiales FIFA/Wikipedia)

### Grupo A — México (sede)
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | México (H) | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Sudáfrica | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Corea del Sur | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | República Checa | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo B — Canadá (sede)
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Canadá (H) | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Bosnia y Herzegovina | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Catar | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Suiza | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo C
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Brasil | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Marruecos | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Haití | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Escocia | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo D — Estados Unidos (sede)
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Estados Unidos (H) | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Paraguay | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Australia | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Türkiye | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo E
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Alemania | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Curazao | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Costa de Marfil | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Ecuador | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo F
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Países Bajos | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Japón | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Suecia | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Túnez | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo G
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Bélgica | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Egipto | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Iran | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Nueva Zelanda | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo H
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | España | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Cabo Verde | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Arabia Saudita | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Uruguay | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo I
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Francia | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Senegal | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Irak | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Noruega | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo J
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Argentina | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Algeria | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Austria | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Jordania | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo K
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Portugal | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | RD Congo | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Uzbekistán | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Colombia | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

### Grupo L
| Pos | Equipo | PJ | G | E | P | GF | GC | DG | Pts | Clasificación |
|-----|--------|----|---|---|---|----|----|----|-----|----------------|
| 1 | Inglaterra | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 2 | Croacia | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Clasifica a octavos |
| 3 | Ghana | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | Repechaje |
| 4 | Panamá | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | — |

---

## Partidos de Fase de Grupos (72 partidos)

### Jornada 1

| # | Fecha | Hora | Local | Visitante | Estadio | Ciudad |
|---|-------|------|-------|----------|---------|--------|
| 1 | 11 jun 2026 | 13:00 UTC-6 | México | Sudáfrica | Estadio Azteca | Ciudad de México |
| 2 | 11 jun 2026 | 20:00 UTC-6 | Corea del Sur | República Checa | Estadio Akron | Zapopan |
| 3 | 12 jun 2026 | 15:00 UTC-4 | Canadá | Bosnia y Herzegovina | BMO Field | Toronto |
| 4 | 12 jun 2026 | 18:00 UTC-7 | Estados Unidos | Paraguay | SoFi Stadium | Inglewood |
| 5 | 13 jun 2026 | 21:00 UTC-4 | Haiti | Escocia | Gillette Stadium | Foxborough |
| 6 | 13 jun 2026 | 21:00 UTC-7 | Australia | Türkiye | BC Place | Vancouver |
| 7 | 13 jun 2026 | 18:00 UTC-4 | Brasil | Marruecos | MetLife Stadium | East Rutherford |
| 8 | 13 jun 2026 | 12:00 UTC-7 | Catar | Suiza | Levi's Stadium | Santa Clara |
| 9 | 14 jun 2026 | 19:00 UTC-4 | Costa de Marfil | Ecuador | Lincoln Financial Field | Filadelfia |
| 10 | 14 jun 2026 | 12:00 UTC-5 | Alemania | Curazao | NRG Stadium | Houston |
| 11 | 14 jun 2026 | 15:00 UTC-5 | Países Bajos | Japón | AT&T Stadium | Arlington |
| 12 | 14 jun 2026 | 20:00 UTC-6 | Suecia | Túnez | Estadio BBVA | Guadalupe |
| 13 | 15 jun 2026 | 18:00 UTC-4 | Arabia Saudita | Uruguay | Hard Rock Stadium | Miami Gardens |
| 14 | 15 jun 2026 | 12:00 UTC-4 | España | Cabo Verde | Mercedes-Benz Stadium | Atlanta |
| 15 | 15 jun 2026 | 18:00 UTC-7 | Iran | Nueva Zelanda | SoFi Stadium | Inglewood |
| 16 | 15 jun 2026 | 12:00 UTC-7 | Bélgica | Egipto | Lumen Field | Seattle |
| 17 | 16 jun 2026 | 20:00 UTC-5 | Francia | Senegal | Arrowhead Stadium | Kansas City |
| 18 | 16 jun 2026 | 21:00 UTC-7 | Austria | Jordania | Levi's Stadium | Santa Clara |
| 19 | 16 jun 2026 | 20:00 UTC-5 | Argentina | Algeria | Arrowhead Stadium | Kansas City |
| 20 | 16 jun 2026 | 21:00 UTC-7 | Irak | Noruega | Levi's Stadium | Santa Clara |
| 21 | 17 jun 2026 | 19:00 UTC-4 | Ghana | Panamá | BMO Field | Toronto |
| 22 | 17 jun 2026 | 15:00 UTC-5 | Inglaterra | Croacia | AT&T Stadium | Arlington |
| 23 | 17 jun 2026 | 12:00 UTC-5 | Portugal | RD Congo | NRG Stadium | Houston |
| 24 | 17 jun 2026 | 20:00 UTC-6 | Uzbekistán | Colombia | Estadio Azteca | Ciudad de México |

### Jornada 2

| # | Fecha | Hora | Local | Visitante | Estadio | Ciudad |
|---|-------|------|-------|----------|---------|--------|
| 25 | 18 jun 2026 | 12:00 UTC-4 | República Checa | Sudáfrica | Mercedes-Benz Stadium | Atlanta |
| 26 | 18 jun 2026 | 12:00 UTC-7 | Suiza | Bosnia y Herzegovina | SoFi Stadium | Inglewood |
| 27 | 18 jun 2026 | 15:00 UTC-7 | Canadá | Catar | BC Place | Vancouver |
| 28 | 18 jun 2026 | 19:00 UTC-6 | México | Corea del Sur | Estadio Akron | Zapopan |
| 29 | 19 jun 2026 | 20:30 UTC-4 | Brasil | Haiti | Lincoln Financial Field | Filadelfia |
| 30 | 19 jun 2026 | 18:00 UTC-4 | Escocia | Marruecos | Gillette Stadium | Foxborough |
| 31 | 19 jun 2026 | 20:00 UTC-7 | Türkiye | Paraguay | Levi's Stadium | Santa Clara |
| 32 | 19 jun 2026 | 12:00 UTC-7 | Estados Unidos | Australia | Lumen Field | Seattle |
| 33 | 20 jun 2026 | 16:00 UTC-4 | Alemania | Costa de Marfil | BMO Field | Toronto |
| 34 | 20 jun 2026 | 19:00 UTC-5 | Ecuador | Curazao | Arrowhead Stadium | Kansas City |
| 35 | 20 jun 2026 | 12:00 UTC-5 | Países Bajos | Suecia | NRG Stadium | Houston |
| 36 | 20 jun 2026 | 22:00 UTC-6 | Túnez | Japón | Estadio BBVA | Guadalupe |
| 37 | 21 jun 2026 | 18:00 UTC-4 | Uruguay | Cabo Verde | Hard Rock Stadium | Miami Gardens |
| 38 | 21 jun 2026 | 12:00 UTC-4 | España | Arabia Saudita | Mercedes-Benz Stadium | Atlanta |
| 39 | 21 jun 2026 | 12:00 UTC-7 | Bélgica | Iran | SoFi Stadium | Inglewood |
| 40 | 21 jun 2026 | 18:00 UTC-7 | Nueva Zelanda | Egipto | BC Place | Vancouver |
| 41 | 21 jun 2026 | 18:00 UTC-4 | Noruega | Senegal | Hard Rock Stadium | Miami Gardens |
| 42 | 21 jun 2026 | 20:00 UTC-4 | Francia | Irak | Hard Rock Stadium | Miami Gardens |
| 43 | 22 jun 2026 | 12:00 UTC-5 | Argentina | Austria | AT&T Stadium | Arlington |
| 44 | 22 jun 2026 | 20:00 UTC-7 | Jordania | Algeria | Levi's Stadium | Santa Clara |
| 45 | 23 jun 2026 | 16:00 UTC-4 | Inglaterra | Ghana | Gillette Stadium | Foxborough |
| 46 | 23 jun 2026 | 19:00 UTC-4 | Panamá | Croacia | BMO Field | Toronto |
| 47 | 23 jun 2026 | 12:00 UTC-5 | Portugal | Uzbekistán | NRG Stadium | Houston |
| 48 | 23 jun 2026 | 20:00 UTC-6 | Colombia | RD Congo | Estadio Akron | Zapopan |

### Jornada 3

| # | Fecha | Hora | Local | Visitante | Estadio | Ciudad |
|---|-------|------|-------|----------|---------|--------|
| 49 | 24 jun 2026 | 18:00 UTC-4 | Escocia | Brasil | Hard Rock Stadium | Miami Gardens |
| 50 | 24 jun 2026 | 18:00 UTC-4 | Marruecos | Haiti | Mercedes-Benz Stadium | Atlanta |
| 51 | 24 jun 2026 | 12:00 UTC-7 | Suiza | Canadá | BC Place | Vancouver |
| 52 | 24 jun 2026 | 12:00 UTC-7 | Bosnia y Herzegovina | Catar | Lumen Field | Seattle |
| 53 | 24 jun 2026 | 19:00 UTC-6 | República Checa | México | Estadio Azteca | Ciudad de México |
| 54 | 24 jun 2026 | 19:00 UTC-6 | Sudáfrica | Corea del Sur | Estadio BBVA | Guadalupe |
| 55 | 25 jun 2026 | 16:00 UTC-4 | Curazao | Costa de Marfil | Lincoln Financial Field | Filadelfia |
| 56 | 25 jun 2026 | 16:00 UTC-4 | Ecuador | Alemania | MetLife Stadium | East Rutherford |
| 57 | 25 jun 2026 | 18:00 UTC-5 | Japón | Suecia | AT&T Stadium | Arlington |
| 58 | 25 jun 2026 | 18:00 UTC-5 | Túnez | Países Bajos | Arrowhead Stadium | Kansas City |
| 59 | 25 jun 2026 | 19:00 UTC-7 | Türkiye | Estados Unidos | SoFi Stadium | Inglewood |
| 60 | 25 jun 2026 | 19:00 UTC-7 | Paraguay | Australia | Levi's Stadium | Santa Clara |
| 61 | 26 jun 2026 | 20:00 UTC-4 | Noruega | Francia | Hard Rock Stadium | Miami Gardens |
| 62 | 26 jun 2026 | 20:00 UTC-4 | Senegal | Irak | Hard Rock Stadium | Miami Gardens |
| 63 | 26 jun 2026 | 20:00 UTC-7 | Egipto | Iran | Lumen Field | Seattle |
| 64 | 26 jun 2026 | 20:00 UTC-7 | Nueva Zelanda | Bélgica | BC Place | Vancouver |
| 65 | 26 jun 2026 | 19:00 UTC-5 | Cabo Verde | Arabia Saudita | NRG Stadium | Houston |
| 66 | 26 jun 2026 | 18:00 UTC-6 | Uruguay | España | Estadio Akron | Zapopan |
| 67 | 27 jun 2026 | 17:00 UTC-4 | Panamá | Inglaterra | MetLife Stadium | East Rutherford |
| 68 | 27 jun 2026 | 17:00 UTC-4 | Croacia | Ghana | Lincoln Financial Field | Filadelfia |
| 69 | 27 jun 2026 | 21:00 UTC-5 | Algeria | Austria | Arrowhead Stadium | Kansas City |
| 70 | 27 jun 2026 | 21:00 UTC-5 | Jordania | Argentina | AT&T Stadium | Arlington |
| 71 | 27 jun 2026 | 19:30 UTC-4 | Colombia | Portugal | Hard Rock Stadium | Miami Gardens |
| 72 | 27 jun 2026 | 19:30 UTC-4 | RD Congo | Uzbekistán | Mercedes-Benz Stadium | Atlanta |

---

## Cruces de 16avos de Final (Round of 32)

> Los cruces dependen de qué terceros de grupo se clasifican. Se usan las tablas de repechaje de la FIFA.

| Partido | Fecha | Hora | Local | Visitante | Estadio | Ciudad |
|---------|-------|------|-------|----------|---------|--------|
| 73 | 28 jun 2026 | 12:00 UTC-7 | 2.° Grupo A | 2.° Grupo B | SoFi Stadium | Inglewood |
| 74 | 29 jun 2026 | 12:00 UTC-5 | 1.° Grupo E | 3.° Grupo A/B/C/D/F | Gillette Stadium | Foxborough |
| 75 | 29 jun 2026 | 19:00 UTC-6 | 1.° Grupo F | 2.° Grupo C | Estadio BBVA | Guadalupe |
| 76 | 29 jun 2026 | 12:00 UTC-5 | 1.° Grupo C | 2.° Grupo F | NRG Stadium | Houston |
| 77 | 29 jun 2026 | 16:30 UTC-4 | 1.° Grupo I | 3.° Grupo C/D/F/G/H | Gillette Stadium | Foxborough |
| 78 | 30 jun 2026 | 12:00 UTC-5 | 1.° Grupo A | 1.° Grupo B | NRG Stadium | Houston |
| 79 | 30 jun 2026 | 12:00 UTC-5 | 1.° Grupo G | 3.° Grupo C/E/F/H/I | NRG Stadium | Houston |
| 80 | 30 jun 2026 | 20:00 UTC-6 | 1.° Grupo D | 3.° Grupo B/E/F/I/J | Estadio Akron | Zapopan |
| 81 | 1 jul 2026 | 12:00 UTC-7 | 2.° Grupo B | 2.° Grupo D | Lumen Field | Seattle |
| 82 | 1 jul 2026 | 12:00 UTC-7 | 2.° Grupo G | 2.° Grupo H | Lumen Field | Seattle |
| 83 | 2 jul 2026 | 12:00 UTC-7 | 2.° Grupo K | 2.° Grupo L | BC Place | Vancouver |
| 84 | 2 jul 2026 | 12:00 UTC-7 | 1.° Grupo H | 3.° Grupo E/H/I/J/K | SoFi Stadium | Inglewood |
| 85 | 3 jul 2026 | 12:00 UTC-4 | 1.° Grupo K | 3.° Grupo D/E/I/J/L | Arrowhead Stadium | Kansas City |
| 86 | 3 jul 2026 | 12:00 UTC-5 | 2.° Grupo H | 2.° Grupo J | AT&T Stadium | Arlington |
| 87 | 3 jul 2026 | 20:00 UTC-4 | 1.° Grupo J | 3.° Grupo D/E/I/J/L | Hard Rock Stadium | Miami Gardens |
| 88 | 3 jul 2026 | 20:00 UTC-5 | 2.° Grupo D | 2.° Grupo E | AT&T Stadium | Arlington |

---

## Código ISO y Banderas

Para las banderas se usan códigos ISO 3166-1 alpha-3. La tabla de equipos para la BD:

| Código | Nombre en español | Grupo |
|--------|-------------------|-------|
| MEX | México | A |
| RSA | Sudáfrica | A |
| KOR | Corea del Sur | A |
| CZE | República Checa | A |
| CAN | Canadá | B |
| BIH | Bosnia y Herzegovina | B |
| QAT | Catar | B |
| SUI | Suiza | B |
| BRA | Brasil | C |
| MAR | Marruecos | C |
| HAI | Hait | C |
| SCO | Escocia | C |
| USA | Estados Unidos | D |
| PAR | Paraguay | D |
| AUS | Australia | D |
| TUR | Türkiye | D |
| GER | Alemania | E |
| CUW | Curazao | E |
| CIV | Costa de Marfil | E |
| ECU | Ecuador | E |
| NED | Países Bajos | F |
| JPN | Japón | F |
| SWE | Suecia | F |
| TUN | Túnez | F |
| BEL | Bélgica | G |
| EGY | Egipto | G |
| IRN | Iran | G |
| NZL | Nueva Zelanda | G |
| ESP | España | H |
| CPV | Cabo Verde | H |
| KSA | Arabia Saudita | H |
| URU | Uruguay | H |
| FRA | Francia | I |
| SEN | Senegal | I |
| IRQ | Irak | I |
| NOR | Noruega | I |
| ARG | Argentina | J |
| ALG | Algeria | J |
| AUT | Austria | J |
| JOR | Jordania | J |
| POR | Portugal | K |
| COD | RD Congo | K |
| UZB | Uzbekistán | K |
| COL | Colombia | K |
| ENG | Inglaterra | L |
| CRO | Croacia | L |
| GHA | Ghana | L |
| PAN | Panamá | L |

## Notas de Implementación

- Las **48 selecciones** se insertan en la migración inicial de la BD
- Los **72 partidos de grupos** y **16 partidos de 16avos** se insertan como datos seed
- Los campos `local_score` y `visitor_score` de 16avos son `NULL` hasta que se jugaron las fases previas
- Las **URLs de banderas** usan `https://flagcdn.com/w40/{iso2}.png` con códigos ISO 3166-1 alpha-2
- Para los terceros de grupo que pasan a 16avos, se calcula dinámicamente según ranking de la FIFA
- El componente Quasar `q-table` se usa para todas las tablas con sorting automático
- Los inputs de pronóstico son `q-input type="number"` con min=0 y validación required
- El locking de partidos se maneja con `v-if="new Date() < new Date(match.match_date)"`

## Tabla de Banderas — Códigos FIFA a ISO 3166-1 alpha-2

Los códigos de flagcdn usan ISO alpha-2 (2 letras), NO los códigos FIFA de 3 letras.

| FIFA (3L) | Nombre en español | ISO-2 | URL bandera |
|-----------|-------------------|-------|-------------|
| MEX | México | mx | `https://flagcdn.com/w40/mx.png` |
| RSA | Sudáfrica | za | `https://flagcdn.com/w40/za.png` |
| KOR | Corea del Sur | kr | `https://flagcdn.com/w40/kr.png` |
| CZE | República Checa | cz | `https://flagcdn.com/w40/cz.png` |
| CAN | Canadá | ca | `https://flagcdn.com/w40/ca.png` |
| BIH | Bosnia y Herzegovina | ba | `https://flagcdn.com/w40/ba.png` |
| QAT | Catar | qa | `https://flagcdn.com/w40/qa.png` |
| SUI | Suiza | ch | `https://flagcdn.com/w40/ch.png` |
| BRA | Brasil | br | `https://flagcdn.com/w40/br.png` |
| MAR | Marruecos | ma | `https://flagcdn.com/w40/ma.png` |
| HAI | Hait | ht | `https://flagcdn.com/w40/ht.png` |
| SCO | Escocia | gb-sct | `https://flagcdn.com/w40/gb-sct.png` |
| USA | Estados Unidos | us | `https://flagcdn.com/w40/us.png` |
| PAR | Paraguay | py | `https://flagcdn.com/w40/py.png` |
| AUS | Australia | au | `https://flagcdn.com/w40/au.png` |
| TUR | Türkiye | tr | `https://flagcdn.com/w40/tr.png` |
| GER | Alemania | de | `https://flagcdn.com/w40/de.png` |
| CUW | Curazao | cw | `https://flagcdn.com/w40/cw.png` |
| CIV | Costa de Marfil | ci | `https://flagcdn.com/w40/ci.png` |
| ECU | Ecuador | ec | `https://flagcdn.com/w40/ec.png` |
| NED | Países Bajos | nl | `https://flagcdn.com/w40/nl.png` |
| JPN | Japón | jp | `https://flagcdn.com/w40/jp.png` |
| SWE | Suecia | se | `https://flagcdn.com/w40/se.png` |
| TUN | Túnez | tn | `https://flagcdn.com/w40/tn.png` |
| BEL | Bélgica | be | `https://flagcdn.com/w40/be.png` |
| EGY | Egipto | eg | `https://flagcdn.com/w40/eg.png` |
| IRN | Iran | ir | `https://flagcdn.com/w40/ir.png` |
| NZL | Nueva Zelanda | nz | `https://flagcdn.com/w40/nz.png` |
| ESP | España | es | `https://flagcdn.com/w40/es.png` |
| CPV | Cabo Verde | cv | `https://flagcdn.com/w40/cv.png` |
| KSA | Arabia Saudita | sa | `https://flagcdn.com/w40/sa.png` |
| URU | Uruguay | uy | `https://flagcdn.com/w40/uy.png` |
| FRA | Francia | fr | `https://flagcdn.com/w40/fr.png` |
| SEN | Senegal | sn | `https://flagcdn.com/w40/sn.png` |
| IRQ | Irak | iq | `https://flagcdn.com/w40/iq.png` |
| NOR | Noruega | no | `https://flagcdn.com/w40/no.png` |
| ARG | Argentina | ar | `https://flagcdn.com/w40/ar.png` |
| ALG | Algeria | dz | `https://flagcdn.com/w40/dz.png` |
| AUT | Austria | at | `https://flagcdn.com/w40/at.png` |
| JOR | Jordania | jo | `https://flagcdn.com/w40/jo.png` |
| POR | Portugal | pt | `https://flagcdn.com/w40/pt.png` |
| COD | RD Congo | cd | `https://flagcdn.com/w40/cd.png` |
| UZB | Uzbekistán | uz | `https://flagcdn.com/w40/uz.png` |
| COL | Colombia | co | `https://flagcdn.com/w40/co.png` |
| ENG | Inglaterra | gb-eng | `https://flagcdn.com/w40/gb-eng.png` |
| CRO | Croacia | hr | `https://flagcdn.com/w40/hr.png` |
| GHA | Ghana | gh | `https://flagcdn.com/w40/gh.png` |
| PAN | Panamá | pa | `https://flagcdn.com/w40/pa.png` |

### Modelo de Equipo con Bandera

```go
type Team struct {
    ID        uint   `db:"id"`
    Name      string `db:"name"`
    ShortCode string `db:"short_code"` // Código FIFA 3 letras (MEX, RSA, etc.)
    ISO2      string `db:"iso2"`        // Código ISO 3166-1 alpha-2 (mx, za, etc.)
    GroupID   uint   `db:"group_id"`
    IsHost    bool   `db:"is_host"`
}
```

### Seed Data de Equipos (Go)

```go
var teams = []struct {
    Name      string
    ShortCode string
    ISO2      string
    GroupID   uint
    IsHost    bool
}{
    {"México", "MEX", "mx", 1, true},
    {"Sudáfrica", "RSA", "za", 1, false},
    {"Corea del Sur", "KOR", "kr", 1, false},
    {"República Checa", "CZE", "cz", 1, false},
    {"Canadá", "CAN", "ca", 2, true},
    {"Bosnia y Herzegovina", "BIH", "ba", 2, false},
    {"Catar", "QAT", "qa", 2, false},
    {"Suiza", "SUI", "ch", 2, false},
    {"Brasil", "BRA", "br", 3, false},
    {"Marruecos", "MAR", "ma", 3, false},
    {"Haití", "HAI", "ht", 3, false},
    {"Escocia", "SCO", "gb-sct", 3, false},
    {"Estados Unidos", "USA", "us", 4, true},
    {"Paraguay", "PAR", "py", 4, false},
    {"Australia", "AUS", "au", 4, false},
    {"Türkiye", "TUR", "tr", 4, false},
    {"Alemania", "GER", "de", 5, false},
    {"Curazao", "CUW", "cw", 5, false},
    {"Costa de Marfil", "CIV", "ci", 5, false},
    {"Ecuador", "ECU", "ec", 5, false},
    {"Países Bajos", "NED", "nl", 6, false},
    {"Japón", "JPN", "jp", 6, false},
    {"Suecia", "SWE", "se", 6, false},
    {"Túnez", "TUN", "tn", 6, false},
    {"Bélgica", "BEL", "be", 7, false},
    {"Egipto", "EGY", "eg", 7, false},
    {"Irán", "IRN", "ir", 7, false},
    {"Nueva Zelanda", "NZL", "nz", 7, false},
    {"España", "ESP", "es", 8, false},
    {"Cabo Verde", "CPV", "cv", 8, false},
    {"Arabia Saudita", "KSA", "sa", 8, false},
    {"Uruguay", "URU", "uy", 8, false},
    {"Francia", "FRA", "fr", 9, false},
    {"Senegal", "SEN", "sn", 9, false},
    {"Irak", "IRQ", "iq", 9, false},
    {"Noruega", "NOR", "no", 9, false},
    {"Argentina", "ARG", "ar", 10, false},
    {"Argelia", "ALG", "dz", 10, false},
    {"Austria", "AUT", "at", 10, false},
    {"Jordania", "JOR", "jo", 10, false},
    {"Portugal", "POR", "pt", 11, false},
    {"RD Congo", "COD", "cd", 11, false},
    {"Uzbekistán", "UZB", "uz", 11, false},
    {"Colombia", "COL", "co", 11, false},
    {"Inglaterra", "ENG", "gb-eng", 12, false},
    {"Croacia", "CRO", "hr", 12, false},
    {"Ghana", "GHA", "gh", 12, false},
    {"Panamá", "PAN", "pa", 12, false},
}
```

### Componente Quasar — Mostrar Banderas en Todas las Vistas

#### 1. Selector de país con bandera (`q-select`)

```vue
<q-select
  v-model="selectedTeam"
  :options="teamOptions"
  option-value="code"
  option-label="label"
  emit-value
  map-options
  label="Selecciona un país"
  outlined
  use-input
  input-debounce="300"
  @filter="filterTeams"
  @update:model-value="loadSchedule"
>
  <template #option="{ itemProps, opt }">
    <q-item v-bind="itemProps">
      <q-item-section avatar>
        <img :src="`https://flagcdn.com/w40/${opt.iso2}.png`" width="30" height="20"
             style="object-fit: cover; border-radius: 2px; border: 1px solid #ddd;" />
      </q-item-section>
      <q-item-section>
        <q-item-label>{{ opt.name }}</q-item-label>
        <q-item-label caption>Grupo {{ opt.group }}</q-item-label>
      </q-item-section>
    </q-item>
  </template>
  <template #selected-item="{ opt }">
    <div class="row items-center no-wrap">
      <img :src="`https://flagcdn.com/w40/${opt.iso2}.png`" width="22" height="15"
           style="object-fit: cover; border-radius: 2px; border: 1px solid #ddd;" class="q-mr-xs" />
      {{ opt.name }}
    </div>
  </template>
</q-select>
```

#### 2. Tabla de grupo con bandera en columna Equipo (`GroupTable.vue`)

```vue
<q-table
  :rows="teams"
  :columns="columns"
  row-key="id"
  flat
  bordered
>
  <template #body-cell-team="props">
    <q-td :props="props">
      <div class="row items-center no-wrap q-gutter-xs">
        <img
          :src="`https://flagcdn.com/w40/${props.row.iso2}.png`"
          width="28"
          height="18"
          style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;"
        />
        <span class="text-weight-medium">{{ props.value }}</span>
        <q-badge v-if="props.row.is_host" color="amber-10" text-color="black" label="(H)" class="q-ml-xs" />
      </div>
    </q-td>
  </template>

  <template #body-cell-classification="props">
    <q-td :props="props">
      <q-badge
        :color="props.value === 'Clasifica a octavos' ? 'positive' : props.value === 'Repechaje' ? 'warning' : 'grey-6'"
        :label="props.value"
      />
    </q-td>
  </template>

  <template #body-cell-pos="props">
    <q-td :props="props" class="text-center">
      <q-badge
        color="primary"
        :label="props.value"
        class="text-weight-bold"
        style="min-width: 24px;"
      />
    </q-td>
  </template>
</q-table>

<script setup>
const columns = [
  { name: 'pos', label: '#', field: 'position', align: 'center', sortable: true },
  { name: 'team', label: 'Selección', field: 'name', align: 'left', sortable: true },
  { name: 'pld', label: 'PJ', field: 'played', align: 'center' },
  { name: 'w', label: 'G', field: 'won', align: 'center' },
  { name: 'd', label: 'E', field: 'drawn', align: 'center' },
  { name: 'l', label: 'P', field: 'lost', align: 'center' },
  { name: 'gf', label: 'GF', field: 'goals_for', align: 'center' },
  { name: 'ga', label: 'GC', field: 'goals_against', align: 'center' },
  { name: 'gd', label: 'DG', field: 'goal_diff', align: 'center' },
  { name: 'pts', label: 'Pts', field: 'points', align: 'center', sortable: true },
  { name: 'classification', label: 'Clasificación', field: 'classification', align: 'center' },
]
</script>
```

#### 3. Tarjeta de partido con banderas (`MatchCard.vue`)

```vue
<div class="match-card q-pa-md q-mb-sm rounded-borders" style="border: 1px solid #e0e0e0;">
  <div class="row items-center justify-between">
    <!-- Equipo Local -->
    <div class="col-4 text-right">
      <img
        :src="`https://flagcdn.com/w40/${match.local_iso2}.png`"
        width="32"
        height="22"
        style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc; vertical-align: middle;"
        class="q-mr-sm"
      />
      <span class="text-subtitle1 text-weight-medium">{{ match.local_team }}</span>
    </div>

    <!-- Marcador / Input -->
    <div class="col-2 text-center">
      <div v-if="isFinished" class="text-h6 text-weight-bold">
        {{ match.local_score }} - {{ match.visitor_score }}
      </div>
      <div v-else-if="canPredict" class="row items-center justify-center q-gutter-xs">
        <q-input
          v-model.number="prediction.local"
          type="number"
          min="0"
          dense
          outlined
          style="width: 50px;"
          :rules="[val => val >= 0 || 'Mínimo 0']"
        />
        <span class="text-h6">-</span>
        <q-input
          v-model.number="prediction.visitor"
          type="number"
          min="0"
          dense
          outlined
          style="width: 50px;"
        />
      </div>
      <div v-else class="text-grey-6 text-caption">Por jugar</div>
    </div>

    <!-- Equipo Visitante -->
    <div class="col-4 text-left">
      <span class="text-subtitle1 text-weight-medium">{{ match.visitor_team }}</span>
      <img
        :src="`https://flagcdn.com/w40/${match.visitor_iso2}.png`"
        width="32"
        height="22"
        style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc; vertical-align: middle;"
        class="q-ml-sm"
      />
    </div>
  </div>

  <!-- Info del partido -->
  <div class="row justify-center q-mt-xs text-caption text-grey-7">
    <q-icon name="event" size="xs" class="q-mr-xs" />
    {{ formatDate(match.match_date) }}
    <q-icon name="location_on" size="xs" class="q-ml-md q-mr-xs" />
    {{ match.venue }}, {{ match.city }}
  </div>
</div>
```

#### 4. Vista completa de Grupos con tabs (`GroupsView.vue`)

Vista para usuarios que muestra la tabla de posiciones y los partidos con pronósticos.

```vue
<template>
  <q-page padding>
    <div class="text-h4 q-mb-md text-center">Fase de Grupos — Mundial 2026</div>

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
      <!-- ... hasta L -->
    </q-tabs>

    <q-tab-panels v-model="tab" animated>
      <q-tab-panel v-for="group in groups" :key="group.id" :name="group.name">
        <!-- Tabla de posiciones -->
        <div class="text-h6 q-mb-sm">Tabla de Posiciones</div>
        <GroupTable :matches="group.matches" :teams="group.teams" />

        <!-- Partidos del grupo -->
        <div class="text-h6 q-mt-lg q-mb-sm">Partidos</div>
        <!-- Mostrar tabla de partidos con inputs de pronóstico -->
      </q-tab-panel>
    </q-tab-panels>
  </q-page>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import GroupTable from 'components/GroupTable.vue'
import { api } from 'src/boot/axios'

const tab = ref('A')
const groups = ref([])

onMounted(async () => {
  const { data } = await api.get('/api/groups')
  groups.value = data
})
</script>
```

> **Importante:** el endpoint `/api/groups` devuelve cada grupo con sus equipos y partidos. Cada equipo incluye `name`, `short_code` y `iso2`. Cada partido incluye `local_team`, `local_iso2`, `visitor_team`, `visitor_iso2`, `match_date`, `venue`, `city` y scores. En Quasar, las imágenes de bandera usan `object-fit: cover` con bordes redondeados de 2px para un aspecto limpio.

### Componente GroupTable.vue — Tabla de Posiciones Calculada

Este componente recibe los partidos del grupo y calcula automáticamente la tabla de posiciones en tiempo real conforme se van cargando los resultados.

```vue
<template>
  <q-table
    :rows="standings"
    :columns="columns"
    row-key="team_id"
    flat
    bordered
    :pagination="{ rowsPerPage: 0 }"
    hide-pagination
    class="standings-table"
  >
    <template #header="props">
      <q-tr :props="props" class="bg-primary text-white text-weight-bold">
        <q-th v-for="col in props.cols" :key="col.name" :props="props" :style="col.style">
          {{ col.label }}
        </q-th>
      </q-tr>
    </template>

    <template #body="props">
      <q-tr :props="props" :class="getRowClass(props.row)">
        <!-- Posición -->
        <q-td key="pos" :props="props" class="text-center">
          <q-badge
            :color="positionColor(props.row.position)"
            text-color="white"
            :label="props.row.position"
            class="text-weight-bold"
            style="min-width: 28px;"
          />
        </q-td>

        <!-- Equipo con bandera -->
        <q-td key="team" :props="props">
          <div class="row items-center no-wrap q-gutter-xs">
            <img
              :src="`https://flagcdn.com/w40/${props.row.iso2}.png`"
              width="28"
              height="18"
              style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;"
            />
            <span class="text-weight-medium">{{ props.row.team_name }}</span>
            <q-badge v-if="props.row.is_host" color="amber-10" text-color="black" label="(H)" class="q-ml-xs" />
          </div>
        </q-td>

        <!-- Partidos Jugados -->
        <q-td key="played" :props="props" class="text-center">
          <q-badge color="grey-7" :label="props.row.played" />
        </q-td>

        <!-- Ganados -->
        <q-td key="won" :props="props" class="text-center">
          <q-badge color="positive" :label="props.row.won" />
        </q-td>

        <!-- Empatados -->
        <q-td key="drawn" :props="props" class="text-center">
          <q-badge color="warning" text-color="black" :label="props.row.drawn" />
        </q-td>

        <!-- Perdidos -->
        <q-td key="lost" :props="props" class="text-center">
          <q-badge color="negative" :label="props.row.lost" />
        </q-td>

        <!-- Goles a Favor -->
        <q-td key="goals_for" :props="props" class="text-center text-weight-medium">
          {{ props.row.goals_for }}
        </q-td>

        <!-- Goles en Contra -->
        <q-td key="goals_against" :props="props" class="text-center">
          {{ props.row.goals_against }}
        </q-td>

        <!-- Diferencia de Goles -->
        <q-td key="goal_diff" :props="props" class="text-center">
          <span :class="props.row.goal_diff > 0 ? 'text-positive' : props.row.goal_diff < 0 ? 'text-negative' : 'text-grey'">
            {{ props.row.goal_diff > 0 ? '+' : '' }}{{ props.row.goal_diff }}
          </span>
        </q-td>

        <!-- Puntos -->
        <q-td key="points" :props="props" class="text-center">
          <q-badge color="primary" text-color="white" :label="props.row.points" class="text-weight-bold text-h6" style="min-width: 36px;" />
        </q-td>
      </q-tr>
    </template>
  </q-table>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  matches: {
    type: Array,
    required: true,
  },
  teams: {
    type: Array,
    required: true,
  },
})

const columns = [
  { name: 'pos', label: '#', field: 'position', align: 'center', style: 'width: 40px;' },
  { name: 'team', label: 'Selección', field: 'team_name', align: 'left', sortable: true },
  { name: 'played', label: 'PJ', field: 'played', align: 'center', style: 'width: 40px;' },
  { name: 'won', label: 'G', field: 'won', align: 'center', style: 'width: 40px;' },
  { name: 'drawn', label: 'E', field: 'drawn', align: 'center', style: 'width: 40px;' },
  { name: 'lost', label: 'P', field: 'lost', align: 'center', style: 'width: 40px;' },
  { name: 'goals_for', label: 'GF', field: 'goals_for', align: 'center', style: 'width: 40px;' },
  { name: 'goals_against', label: 'GC', field: 'goals_against', align: 'center', style: 'width: 40px;' },
  { name: 'goal_diff', label: 'DG', field: 'goal_diff', align: 'center', style: 'width: 50px;' },
  { name: 'points', label: 'Pts', field: 'points', align: 'center', sortable: true, style: 'width: 50px;' },
]

const standings = computed(() => {
  const stats = {}

  // Inicializar stats para cada equipo
  props.teams.forEach(team => {
    stats[team.id] = {
      team_id: team.id,
      team_name: team.name,
      iso2: team.iso2,
      is_host: team.is_host,
      played: 0,
      won: 0,
      drawn: 0,
      lost: 0,
      goals_for: 0,
      goals_against: 0,
      goal_diff: 0,
      points: 0,
    }
  })

  // Recorrer solo partidos finalizados
  props.matches
    .filter(m => m.status === 'finished' && m.local_score !== null && m.visitor_score !== null)
    .forEach(match => {
      const local = stats[match.local_team_id]
      const visitor = stats[match.visitor_team_id]

      if (!local || !visitor) return

      local.played++
      visitor.played++
      local.goals_for += match.local_score
      local.goals_against += match.visitor_score
      visitor.goals_for += match.visitor_score
      visitor.goals_against += match.local_score

      if (match.local_score > match.visitor_score) {
        local.won++
        local.points += 3
        visitor.lost++
      } else if (match.local_score < match.visitor_score) {
        visitor.won++
        visitor.points += 3
        local.lost++
      } else {
        local.drawn++
        local.points += 1
        visitor.drawn++
        visitor.points += 1
      }
    })

  // Calcular diferencia de goles
  Object.values(stats).forEach(s => {
    s.goal_diff = s.goals_for - s.goals_against
  })

  // Ordenar: Pts DESC → DG DESC → GF DESC → Nombre ASC
  const sorted = Object.values(stats).sort((a, b) => {
    if (b.points !== a.points) return b.points - a.points
    if (b.goal_diff !== a.goal_diff) return b.goal_diff - a.goal_diff
    if (b.goals_for !== a.goals_for) return b.goals_for - a.goals_for
    return a.team_name.localeCompare(b.team_name)
  })

  // Asignar posición
  sorted.forEach((s, i) => {
    s.position = i + 1
  })

  return sorted
})

const positionColor = (pos) => {
  if (pos <= 2) return 'positive'
  if (pos === 3) return 'warning'
  return 'grey-6'
}

const getRowClass = (row) => {
  if (row.position <= 2) return 'bg-green-1'
  if (row.position === 3) return 'bg-amber-1'
  return 'bg-grey-2'
}
</script>

<style scoped>
.standings-table {
  font-size: 14px;
}
</style>
```

#### Lógica de cálculo de posiciones:

| Campo | Descripción | Cálculo |
|-------|-------------|---------|
| **PJ** | Partidos Jugados | Contador de partidos con `status === 'finished'` |
| **G** | Partidos Ganados | Contador cuando `local_score > visitor_score` (de local) o viceversa |
| **E** | Partidos Empatados | Contador cuando `local_score === visitor_score` |
| **P** | Partidos Perdidos | Contador cuando `local_score < visitor_score` (de local) o viceversa |
| **GF** | Goles a Favor | Suma de `local_score` (como local) + `visitor_score` (como visitante) |
| **GC** | Goles en Contra | Suma de `visitor_score` (como local) + `local_score` (como visitante) |
| **DG** | Diferencia de Goles | `GF - GC` (mostrar `+` si es positivo) |
| **Pts** | Puntos | `G × 3 + E × 1 + P × 0` |

#### Ordenamiento de la tabla:

```
1. Puntos (Pts) DESC
2. Diferencia de goles (DG) DESC
3. Goles a favor (GF) DESC
4. Nombre del equipo ASC (desempate alfabético)
```

#### Endpoint `GET /api/groups/:id/standings` — Handler Go:

```go
func GetGroupStandings(c *gin.Context) {
    groupID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

    // Obtener equipos del grupo
    rows, err := db.Query(`
        SELECT id, name, iso2, is_host FROM teams WHERE group_id = ?
    `, groupID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    stats := make(map[uint]map[string]interface{})
    for rows.Next() {
        var id uint
        var name, iso2 string
        var isHost bool
        rows.Scan(&id, &name, &iso2, &isHost)
        stats[id] = map[string]interface{}{
            "id": id, "name": name, "iso2": iso2, "is_host": isHost,
            "played": 0, "won": 0, "drawn": 0, "lost": 0,
            "goals_for": 0, "goals_against": 0, "goal_diff": 0, "points": 0,
        }
    }

    // Obtener partidos finalizados del grupo
    matchRows, _ := db.Query(`
        SELECT id, local_team_id, visitor_team_id, local_score, visitor_score, status
        FROM matches WHERE group_id = ? AND status = 'finished'
    `, groupID)

    for matchRows.Next() {
        var id uint
        var localID, visitorID uint
        var localScore, visitorScore int
        var status string
        matchRows.Scan(&id, &localID, &visitorID, &localScore, &visitorScore, &status)

        local := stats[localID]
        visitor := stats[visitorID]

        local["played"] = local["played"].(int) + 1
        visitor["played"] = visitor["played"].(int) + 1
        local["goals_for"] = local["goals_for"].(int) + localScore
        local["goals_against"] = local["goals_against"].(int) + visitorScore
        visitor["goals_for"] = visitor["goals_for"].(int) + visitorScore
        visitor["goals_against"] = visitor["goals_against"].(int) + localScore

        if localScore > visitorScore {
            local["won"] = local["won"].(int) + 1
            local["points"] = local["points"].(int) + 3
            visitor["lost"] = visitor["lost"].(int) + 1
        } else if localScore < visitorScore {
            visitor["won"] = visitor["won"].(int) + 1
            visitor["points"] = visitor["points"].(int) + 3
            local["lost"] = local["lost"].(int) + 1
        } else {
            local["drawn"] = local["drawn"].(int) + 1
            local["points"] = local["points"].(int) + 1
            visitor["drawn"] = visitor["drawn"].(int) + 1
            visitor["points"] = visitor["points"].(int) + 1
        }
    }

    // Calcular DG y ordenar
    for _, s := range stats {
        gf := s["goals_for"].(int)
        gc := s["goals_against"].(int)
        s["goal_diff"] = gf - gc
    }

    sorted := sortStandings(stats)
    for i, s := range sorted {
        s["position"] = i + 1
    }

    c.JSON(200, sorted)
}

func sortStandings(stats map[uint]map[string]interface{}) []map[string]interface{} {
    sorted := make([]map[string]interface{}, 0, len(stats))
    for _, s := range stats {
        sorted = append(sorted, s)
    }
    sort.Slice(sorted, func(i, j int) bool {
        si, sj := sorted[i], sorted[j]
        if si["points"].(int) != sj["points"].(int) {
            return si["points"].(int) > sj["points"].(int)
        }
        if si["goal_diff"].(int) != sj["goal_diff"].(int) {
            return si["goal_diff"].(int) > sj["goal_diff"].(int)
        }
        if si["goals_for"].(int) != sj["goals_for"].(int) {
            return si["goals_for"].(int) > sj["goals_for"].(int)
        }
        return si["name"].(string) < sj["name"].(string)
    })
    return sorted
}
```

#### Integración con ResultsView.vue y GroupsView.vue:

```
1. Admin abre vista ResultsView.vue
2. Se cargan todos los grupos con partidos y equipos
3. scores{} se inicializa con los marcadores actuales (o 0 si no hay)
4. calculateAllStandings() calcula tabla de posiciones con resultados existentes
5. Admin escribe goles en los inputs → markDirty()
6. Admin hace clic en:
   a) 💾 Guardar (individual) → POST /api/results → recalcula standings localmente
   b) 💾 Guardar Todos → guarda todos los pendientes → recalcula standings
   c) 🔄 Actualizar Tablas → recalcula standings con datos actuales de la BD
7. Tabla de posiciones se actualiza mostrando: PJ, G, E, P, GF, GC, DG, Pts
8. Fila verde para 1.°-2.°, fila amarilla para 3.°, fila gris para 4.°
```

#### Grupos del Mundial 2026 (datos oficiales FIFA/Wikipedia)

#### Endpoint `GET /api/schedule/:teamCode` — SQL:

```sql
SELECT
  m.id,
  m.match_number,
  m.match_date,
  m.venue,
  m.city,
  m.stage,
  m.local_score,
  m.visitor_score,
  lt.name AS local_team,
  lt.iso2 AS local_iso2,
  vt.name AS visitor_team,
  vt.iso2 AS visitor_iso2,
  g.name AS group_name
FROM matches m
JOIN teams lt ON m.local_team_id = lt.id
JOIN teams vt ON m.visitor_team_id = vt.id
LEFT JOIN groups g ON m.group_id = g.id
WHERE lt.iso2 = :iso2 OR vt.iso2 = :iso2
ORDER BY m.match_date ASC
LIMIT :limit OFFSET :offset
```

> En el backend Go se usa un conteo previo (`SELECT COUNT(*)`) para calcular `rowsNumber` del paginado.

#### Selector de filas por página en `q-table`:
```vue
<q-table
  :rows-per-page-options="[5, 10, 20, 50, 100]"
  :pagination="pagination"
  @request="onRequest"
/>
```

### Componente ScheduleView.vue — Detalle de Implementación

Vista accesible desde el menú principal que permite consultar todos los horarios de la primera fase del Mundial filtrados por país.

#### Código Vue 3 + Quasar:

```vue
<template>
  <q-page padding>
    <div class="text-h5 q-mb-md">Horarios del Mundial 2026</div>

    <!-- Selector de país -->
    <q-select
      v-model="selectedTeam"
      :options="teamOptions"
      option-value="iso2"
      option-label="label"
      emit-value
      map-options
      label="Selecciona un país"
      class="q-mb-md"
      outlined
      use-input
      input-debounce="300"
      @filter="filterTeams"
      @update:model-value="loadSchedule"
    >
      <template #option="{ itemProps, opt }">
        <q-item v-bind="itemProps">
          <q-item-section avatar>
            <img :src="`https://flagcdn.com/w40/${opt.iso2}.png`" width="30" height="20"
                 style="object-fit: cover; border-radius: 3px; border: 1px solid #ddd;" />
          </q-item-section>
          <q-item-section>
            <q-item-label>{{ opt.name }}</q-item-label>
            <q-item-label caption>Grupo {{ opt.group }}</q-item-label>
          </q-item-section>
        </q-item>
      </template>
      <template #selected-item="{ opt }">
        <div class="row items-center no-wrap">
          <img :src="`https://flagcdn.com/w40/${opt.iso2}.png`" width="22" height="15"
               style="object-fit: cover; border-radius: 3px; border: 1px solid #ddd;" class="q-mr-xs" />
          {{ opt.name }}
        </div>
      </template>
    </q-select>

    <!-- Tabla de horarios con paginación -->
    <q-table
      v-if="selectedTeam"
      v-model:pagination="pagination"
      :rows="scheduleRows"
      :columns="columns"
      row-key="match_number"
      :pagination="pagination"
      :rows-per-page-options="[5, 10, 20, 50, 100]"
      @request="onRequest"
    >
      <template #body-cell-match_number="props">
        <q-td :props="props">
          <q-badge color="primary" :label="props.value" />
        </q-td>
      </template>

      <template #body-cell-local_team="props">
        <q-td :props="props">
          <div class="row items-center">
            <img :src="`https://flagcdn.com/w40/${props.row.local_iso2}.png`" width="20" height="14"
                 style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;" class="q-mr-xs" />
            {{ props.value }}
            <q-badge v-if="props.row.local_iso2 === selectedTeam" color="accent" class="q-ml-xs" label="Tú" />
          </div>
        </q-td>
      </template>

      <template #body-cell-visitor_team="props">
        <q-td :props="props">
          <div class="row items-center">
            <img :src="`https://flagcdn.com/w40/${props.row.visitor_iso2}.png`" width="20" height="14"
                 style="object-fit: cover; border-radius: 2px; border: 1px solid #ccc;" class="q-mr-xs" />
            {{ props.value }}
            <q-badge v-if="props.row.visitor_iso2 === selectedTeam" color="accent" class="q-ml-xs" label="Tú" />
          </div>
        </q-td>
      </template>

      <template #body-cell-match_date="props">
        <q-td :props="props">
          {{ formatDate(props.value) }}
        </q-td>
      </template>

      <template #body-cell-stage="props">
        <q-td :props="props">
          <q-badge :color="stageColor(props.value)" :label="stageLabel(props.value)" />
        </q-td>
      </template>

      <template #body-cell-score="props">
        <q-td :props="props">
          <span v-if="props.row.local_score !== null">
            {{ props.row.local_score }} - {{ props.row.visitor_score }}
          </span>
          <span v-else class="text-grey">Por jugar</span>
        </q-td>
      </template>
    </q-table>

    <!-- Mensaje cuando no hay selección -->
    <q-banner v-else class="q-mt-md" inline-actions type="info">
      Selecciona un país para ver sus horarios
    </q-banner>
  </q-page>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api } from 'src/boot/axios'

const selectedTeam = ref(null) // holds the iso2 code
const teamOptions = ref([])
const allTeams = ref([])
const scheduleRows = ref([])
const loading = ref(false)

const columns = [
  { name: 'match_number', label: '#', field: 'match_number', align: 'center', sortable: true },
  { name: 'match_date', label: 'Fecha', field: 'match_date', align: 'left', sortable: true },
  { name: 'local_team', label: 'Local', field: 'local_team', align: 'left', sortable: true },
  { name: 'visitor_team', label: 'Visitante', field: 'visitor_team', align: 'left', sortable: true },
  { name: 'venue', label: 'Estadio', field: 'venue', align: 'left', sortable: true },
  { name: 'city', label: 'Ciudad', field: 'city', align: 'left' },
  { name: 'stage', label: 'Fase', field: 'stage', align: 'center', sortable: true },
  { name: 'score', label: 'Marcador', field: 'match_number', align: 'center' },
]

const pagination = ref({
  page: 1,
  rowsPerPage: 10,
  sortBy: 'match_date',
  descending: false,
  rowsNumber: 0,
})

onMounted(async () => {
  const { data } = await api.get('/api/teams')
  allTeams.value = data
  teamOptions.value = data.map(t => ({
    iso2: t.iso2,
    name: t.name,
    group: t.group_name,
  }))
})

const filterTeams = (val, update) => {
  update(() => {
    const needle = val.toLowerCase()
    teamOptions.value = allTeams.value
      .filter(t => t.name.toLowerCase().includes(needle))
      .map(t => ({ iso2: t.iso2, name: t.name, group: t.group_name }))
  })
}

const loadSchedule = async (iso2) => {
  if (!iso2) return
  loading.value = true
  const { data } = await api.get(`/api/schedule/${iso2}`)
  scheduleRows.value = data
  pagination.value.rowsNumber = data.length
  loading.value = false
}

const onRequest = (props) => {
  pagination.value = props.pagination
}

const formatDate = (dateStr) => {
  const d = new Date(dateStr)
  return d.toLocaleDateString('es-ES', { weekday: 'short', day: 'numeric', month: 'short' })
}

const stageColor = (stage) => ({
  group: 'blue',
  r32: 'orange',
  r16: 'deep-orange',
  qf: 'red',
  sf: 'purple',
  f: 'green',
  third: 'grey',
}[stage] || 'grey')

const stageLabel = (stage) => ({
  group: 'Fase de Grupos',
  r32: '16avos',
  r16: 'Octavos',
  qf: 'Cuartos',
  sf: 'Semifinal',
  f: 'Final',
  third: '3er Lugar',
}[stage] || stage)
</script>
```

#### Detalle del paginado:
- **Filas por página**: selector con `[5, 10, 20, 50, 100]` opciones usando `rows-per-page-options`
- **Navegación**: componente `q-pagination` integrado en `q-table` (prop `pagination`)
- **Ordenamiento**: cada columna es `sortable: true`; default: ordenar por `match_date` ascendente
- **Total de filas**: `rowsNumber` se actualiza al cargar los datos del endpoint
- El endpoint `GET /api/schedule/:iso2` filtra en la BD: `WHERE local_team_id = ? OR visitor_team_id = ? ORDER BY match_date ASC`