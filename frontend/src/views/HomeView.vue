<template>
  <q-page class="flex flex-center q-pa-md">
    <div class="text-center">
      <q-icon name="emoji_events" size="5em" color="amber-7" class="q-mb-md" />
      <h1 class="text-h3 text-primary text-weight-bold q-mb-xs">Quinela Mundial 2026</h1>
      <p class="text-h6 text-grey-7 q-mb-xl">Copa Mundial de la FIFA — Estados Unidos, México y Canadá 2026</p>
      <div class="row justify-center q-gutter-md">
        <q-btn color="primary" size="lg" to="/groups" label="Ver Grupos" icon="groups" unelevated />
        <q-btn color="secondary" size="lg" to="/schedule" label="Calendario" icon="calendar_today" unelevated />
        <q-btn color="positive" size="lg" to="/standings" label="Tabla Quiniela" icon="leaderboard" unelevated />
        <q-btn color="accent" size="lg" to="/results" label="Ingresar Resultados" icon="scoreboard" unelevated />
      </div>
    </div>
  </q-page>
</template>

<script setup>
</script>
