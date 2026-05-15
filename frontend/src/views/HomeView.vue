<template>
  <q-page class="home-page">

    <!-- Hero con imagen de fondo -->
    <div class="hero-section">
      <div class="hero-overlay" />

      <!-- Logo / trofeo flotante -->
      <div class="trophy-badge">
        <q-icon name="emoji_events" size="2.5rem" color="amber-3" />
      </div>

      <!-- Imagen principal Unica.jpg -->
      <div class="hero-image-wrap">
        <img :src="unicaImg" class="hero-image" alt="Quiniela Mundial 2026" />
        <div class="hero-glow" />
      </div>

      <!-- Textos sobre el hero -->
      <div class="hero-text q-mt-md">
        <div class="world-cup-label">
          <span class="wc-dot" /><span>FIFA WORLD CUP 2026™</span><span class="wc-dot" />
        </div>
        <h1 class="quiniela-title">QUINIELA<br /><span class="title-accent">MUNDIAL</span></h1>
        <p class="quiniela-subtitle">🇺🇸 Estados Unidos &nbsp;·&nbsp; 🇲🇽 México &nbsp;·&nbsp; 🇨🇦 Canadá</p>
      </div>
    </div>

    <!-- Separador decorativo -->
    <div class="divider-strip">
      <span v-for="i in 12" :key="i" class="strip-flag">⚽</span>
    </div>

    <!-- Botones de navegación -->
    <div class="nav-section q-px-md">
      <div class="row justify-center q-gutter-md">
        <q-btn
          v-for="item in navItems" :key="item.to"
          :to="item.to"
          :icon="item.icon"
          :label="item.label"
          class="nav-btn"
          :style="`--btn-color: ${item.color}`"
          unelevated no-caps
        />
      </div>
    </div>

    <!-- Botón de contacto / desarrollador -->
    <div class="contact-trigger-wrap q-pb-xl q-pt-md">
      <button class="contact-trigger-btn" @click="showContact = true">
        <span class="ctb-glow" />
        <span class="ctb-icon">🚀</span>
        <span class="ctb-text">
          <span class="ctb-top">¿Tienes un proyecto en mente?</span>
          <span class="ctb-sub">Hablemos — desarrollo a medida</span>
        </span>
        <span class="ctb-arrow">›</span>
      </button>
    </div>

    <!-- ═══════════════════ DIÁLOGO DE CONTACTO ═══════════════════ -->
    <q-dialog v-model="showContact" :maximized="false" transition-show="slide-up" transition-hide="slide-down">
      <div class="contact-card">

        <!-- Botón X -->
        <button class="cc-close" @click="showContact = false" aria-label="Cerrar">
          <q-icon name="close" size="1.1rem" />
        </button>

        <!-- Header degradado -->
        <div class="cc-header">
          <div class="cc-badge">
            <span class="cc-badge-icon">⚡</span>
          </div>
          <p class="cc-greeting">Hola, soy el desarrollador</p>
          <h2 class="cc-headline">Convierte tu idea<br/><span class="cc-hl-accent">en software real</span></h2>
          <p class="cc-pitch">
            Gracias por usar mi Quiniela. Fue desarrollada en
            <strong>Go</strong> buscando rendimiento y estabilidad.
          </p>
        </div>

        <!-- Servicios -->
        <div class="cc-services">
          <div class="cc-service" v-for="s in services" :key="s.icon">
            <span class="cc-svc-icon">{{ s.icon }}</span>
            <div>
              <div class="cc-svc-title">{{ s.title }}</div>
              <div class="cc-svc-desc">{{ s.desc }}</div>
            </div>
          </div>
        </div>

        <!-- Mensaje cierre -->
        <p class="cc-cta-text">
          ¿Tienes un proyecto en mente en el que pueda ayudarte?
          <strong>Escríbeme</strong> y lo conversamos sin compromiso.
        </p>
         <p class="cc-footer-note">osiris.remolina.golang@gmail.com</p>

        <!-- Acciones -->
        <div class="cc-actions">
          <a
            href="mailto:osiris.remolina.golang@gmail.com?subject=Proyecto%20de%20software&body=Hola%20Osiris%2C%20me%20interesa%20comentarte%20sobre%20mi%20proyecto..."
            class="cc-btn-contact"
          >
            <q-icon name="mail" size="1.1rem" class="q-mr-xs" />
            Contactar ahora
          </a>
          <button class="cc-btn-continue" @click="showContact = false">
            Seguir explorando
          </button>
        </div>

        <!-- Nota final -->
        <p class="cc-footer-note">osiris.remolina.golang@gmail.com</p>

      </div>
    </q-dialog>

  </q-page>
</template>

<script setup>
import unicaImg from '../assets/images/Unica.jpg'
import { ref } from 'vue'

const showContact = ref(false)

const navItems = [
  { to: '/groups',     icon: 'groups',         label: 'Ver Grupos',           color: '#1565C0' },
  { to: '/schedule',   icon: 'calendar_today',  label: 'Calendario',           color: '#00695C' },
  { to: '/predictions',icon: 'how_to_vote',     label: 'Mis Pronósticos',      color: '#6A1B9A' },
  { to: '/standings',  icon: 'leaderboard',     label: 'Tabla Quiniela',       color: '#E65100' },
  { to: '/results',    icon: 'scoreboard',      label: 'Ingresar Resultados',  color: '#B71C1C' },
]

const services = [
  { icon: '🖥️',  title: 'Software de escritorio',  desc: 'Apps nativas de alto rendimiento con Go + Wails' },
  { icon: '🌐',  title: 'APIs & Servidores',         desc: 'Backends REST rápidos, seguros y escalables' },
  { icon: '📊',  title: 'Herramientas de gestión',   desc: 'Reportes, dashboards y módulos empresariales' },
]
</script>

<style scoped>
/* ── Página ── */
.home-page {
  background: #0a0f1e;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* ── Hero ── */
.hero-section {
  position: relative;
  width: 100%;
  max-width: 700px;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 28px;
}

.hero-overlay {
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at 50% 0%, rgba(255,215,0,0.12) 0%, transparent 70%);
  pointer-events: none;
}

.trophy-badge {
  background: linear-gradient(135deg, #b8860b, #ffd700, #b8860b);
  border-radius: 50%;
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 0 24px rgba(255,215,0,0.7), 0 0 60px rgba(255,215,0,0.25);
  margin-bottom: 18px;
  animation: pulse-glow 2.5s ease-in-out infinite alternate;
}

@keyframes pulse-glow {
  from { box-shadow: 0 0 18px rgba(255,215,0,0.6), 0 0 40px rgba(255,215,0,0.2); }
  to   { box-shadow: 0 0 36px rgba(255,215,0,1),   0 0 80px rgba(255,215,0,0.4); }
}

/* ── Imagen ── */
.hero-image-wrap {
  position: relative;
  width: 88%;
  max-width: 580px;
}

.hero-image {
  width: 100%;
  border-radius: 20px;
  display: block;
  border: 3px solid rgba(255, 215, 0, 0.55);
  box-shadow:
    0 0 0 1px rgba(255,215,0,0.2),
    0 8px 40px rgba(0,0,0,0.7),
    0 0 80px rgba(255,215,0,0.15);
}

.hero-glow {
  position: absolute;
  inset: 0;
  border-radius: 20px;
  box-shadow: inset 0 -60px 60px rgba(10,15,30,0.85);
  pointer-events: none;
}

/* ── Texto Hero ── */
.hero-text {
  text-align: center;
  padding: 0 16px;
  margin-top: 10px;
}

.world-cup-label {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-family: 'Roboto Condensed', sans-serif;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.25em;
  color: #ffd700;
  text-transform: uppercase;
  margin-bottom: 8px;
}
.wc-dot {
  width: 5px; height: 5px;
  background: #ffd700;
  border-radius: 50%;
  display: inline-block;
}

.quiniela-title {
  font-family: 'Roboto Condensed', 'Arial Narrow', sans-serif;
  font-size: clamp(2.8rem, 9vw, 5rem);
  font-weight: 900;
  line-height: 0.9;
  color: #ffffff;
  letter-spacing: -0.02em;
  text-shadow: 0 4px 20px rgba(0,0,0,0.5);
  margin: 0 0 10px;
}
.title-accent {
  color: #ffd700;
  text-shadow: 0 0 30px rgba(255,215,0,0.6);
}

.quiniela-subtitle {
  font-size: 0.85rem;
  color: rgba(255,255,255,0.6);
  letter-spacing: 0.05em;
  margin: 0;
}

/* ── Separador ── */
.divider-strip {
  width: 100%;
  background: linear-gradient(90deg, #0a0f1e, #1a2540, #0a0f1e);
  border-top: 1px solid rgba(255,215,0,0.2);
  border-bottom: 1px solid rgba(255,215,0,0.2);
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 8px 0;
  font-size: 1rem;
  opacity: 0.5;
  margin: 20px 0 10px;
  overflow: hidden;
}

/* ── Botones ── */
.nav-section {
  width: 100%;
  max-width: 680px;
}

.nav-btn {
  font-family: 'Roboto Condensed', sans-serif;
  font-weight: 700;
  font-size: 0.85rem;
  letter-spacing: 0.06em;
  padding: 12px 22px;
  border-radius: 12px;
  background: var(--btn-color) !important;
  color: #fff !important;
  border: 1px solid rgba(255,255,255,0.15) !important;
  box-shadow: 0 4px 20px rgba(0,0,0,0.4), inset 0 1px 0 rgba(255,255,255,0.1);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.nav-btn:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 28px rgba(0,0,0,0.5), 0 0 20px color-mix(in srgb, var(--btn-color) 60%, transparent);
}

.nav-btn:active {
  transform: translateY(0);
}

/* ══════════════════════════════════════
   BOTÓN TRIGGER DE CONTACTO
══════════════════════════════════════ */
.contact-trigger-wrap {
  width: 100%;
  max-width: 680px;
  display: flex;
  justify-content: center;
  padding: 0 16px;
}

.contact-trigger-btn {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  max-width: 480px;
  background: linear-gradient(135deg, #1a1f3a 0%, #0f1628 100%);
  border: 1px solid rgba(255,215,0,0.35);
  border-radius: 16px;
  padding: 14px 20px;
  cursor: pointer;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  box-shadow: 0 4px 30px rgba(0,0,0,0.5), 0 0 0 0 rgba(255,215,0,0);
}
.contact-trigger-btn:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 36px rgba(0,0,0,0.6), 0 0 24px rgba(255,215,0,0.2);
  border-color: rgba(255,215,0,0.7);
}
.contact-trigger-btn:hover .ctb-glow {
  opacity: 1;
}

.ctb-glow {
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at 0% 50%, rgba(255,215,0,0.08) 0%, transparent 70%);
  opacity: 0;
  transition: opacity 0.3s;
  pointer-events: none;
}

.ctb-icon {
  font-size: 1.6rem;
  flex-shrink: 0;
  filter: drop-shadow(0 0 8px rgba(255,215,0,0.6));
}

.ctb-text {
  flex: 1;
  text-align: left;
}
.ctb-top {
  display: block;
  font-size: 0.88rem;
  font-weight: 700;
  color: #ffd700;
  letter-spacing: 0.03em;
}
.ctb-sub {
  display: block;
  font-size: 0.72rem;
  color: rgba(255,255,255,0.45);
  margin-top: 2px;
}

.ctb-arrow {
  font-size: 1.6rem;
  color: rgba(255,215,0,0.5);
  font-weight: 300;
  transition: transform 0.2s;
}
.contact-trigger-btn:hover .ctb-arrow {
  transform: translateX(4px);
  color: #ffd700;
}

/* ══════════════════════════════════════
   DIÁLOGO / CARD DE CONTACTO
══════════════════════════════════════ */
.contact-card {
  position: relative;
  width: 92vw;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
  background: #0d1225;
  border-radius: 24px;
  border: 1px solid rgba(255,215,0,0.25);
  box-shadow:
    0 0 0 1px rgba(255,255,255,0.04),
    0 24px 80px rgba(0,0,0,0.8),
    0 0 60px rgba(255,215,0,0.06);
  overflow: hidden;
}

/* Botón X */
.cc-close {
  position: absolute;
  top: 14px;
  right: 14px;
  z-index: 10;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid rgba(255,255,255,0.12);
  background: rgba(255,255,255,0.06);
  color: rgba(255,255,255,0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
}
.cc-close:hover {
  background: rgba(255,100,100,0.2);
  color: #ff6b6b;
  border-color: rgba(255,100,100,0.4);
}

/* Header */
.cc-header {
  background: linear-gradient(160deg, #1a2a5e 0%, #0e1830 60%, #0d1225 100%);
  padding: 36px 28px 24px;
  text-align: center;
  border-bottom: 1px solid rgba(255,215,0,0.1);
}

.cc-badge {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #b8860b, #ffd700);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 14px;
  box-shadow: 0 0 20px rgba(255,215,0,0.5);
  animation: pulse-glow 2.5s ease-in-out infinite alternate;
}
.cc-badge-icon { font-size: 1.6rem; }

.cc-greeting {
  font-size: 0.72rem;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: #ffd700;
  margin: 0 0 6px;
  opacity: 0.8;
}

.cc-headline {
  font-size: clamp(1.5rem, 5vw, 2rem);
  font-weight: 900;
  color: #fff;
  line-height: 1.1;
  margin: 0 0 14px;
  font-family: 'Roboto Condensed', sans-serif;
}
.cc-hl-accent {
  color: #ffd700;
  text-shadow: 0 0 20px rgba(255,215,0,0.5);
}

.cc-pitch {
  font-size: 0.82rem;
  color: rgba(255,255,255,0.55);
  line-height: 1.55;
  margin: 0;
}
.cc-pitch strong { color: rgba(255,255,255,0.85); }

/* Servicios */
.cc-services {
  padding: 20px 24px 4px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cc-service {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: rgba(255,255,255,0.03);
  border: 1px solid rgba(255,255,255,0.06);
  border-radius: 12px;
  padding: 11px 14px;
  transition: background 0.2s;
}
.cc-service:hover {
  background: rgba(255,215,0,0.04);
  border-color: rgba(255,215,0,0.2);
}

.cc-svc-icon { font-size: 1.3rem; flex-shrink: 0; margin-top: 1px; }
.cc-svc-title {
  font-size: 0.83rem;
  font-weight: 700;
  color: #fff;
  margin-bottom: 2px;
}
.cc-svc-desc {
  font-size: 0.72rem;
  color: rgba(255,255,255,0.45);
  line-height: 1.4;
}

/* CTA text */
.cc-cta-text {
  text-align: center;
  font-size: 0.8rem;
  color: rgba(255,255,255,0.5);
  padding: 16px 28px 0;
  line-height: 1.5;
  margin: 0;
}
.cc-cta-text strong { color: #ffd700; }

/* Acciones */
.cc-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 20px 24px 8px;
}

.cc-btn-contact {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 14px;
  border-radius: 12px;
  background: linear-gradient(135deg, #b8860b 0%, #ffd700 50%, #b8860b 100%);
  background-size: 200% auto;
  color: #0a0f1e;
  font-weight: 800;
  font-size: 0.9rem;
  letter-spacing: 0.04em;
  text-decoration: none;
  border: none;
  cursor: pointer;
  transition: background-position 0.4s ease, transform 0.15s ease, box-shadow 0.2s ease;
  box-shadow: 0 4px 20px rgba(255,215,0,0.35);
}
.cc-btn-contact:hover {
  background-position: right center;
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(255,215,0,0.55);
}

.cc-btn-continue {
  padding: 12px;
  border-radius: 12px;
  background: transparent;
  border: 1px solid rgba(255,255,255,0.12);
  color: rgba(255,255,255,0.45);
  font-size: 0.82rem;
  cursor: pointer;
  transition: background 0.2s, color 0.2s, border-color 0.2s;
}
.cc-btn-continue:hover {
  background: rgba(255,255,255,0.05);
  color: rgba(255,255,255,0.8);
  border-color: rgba(255,255,255,0.25);
}

/* Nota email */
.cc-footer-note {
  text-align: center;
  font-size: 0.68rem;
  color: rgba(255,255,255,0.2);
  padding: 0 24px 20px;
  letter-spacing: 0.05em;
  margin: 0;
}
</style>

