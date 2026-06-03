# Manual de Usuario — Quiniela Mundial 2026

## Tabla de Contenidos

1. [Introducción](#introducción)
2. [Registro e Inicio de Sesión](#registro-e-inicio-de-sesión)
3. [Pantalla Principal](#pantalla-principal)
4. [Fase de Grupos](#fase-de-grupos)
5. [Horarios por País](#horarios-por-país)
6. [Pronósticos](#pronósticos)
7. [Eliminatorias](#eliminatorias)
8. [Tabla de Posiciones](#tabla-de-posiciones)
9. [Sistema de Puntaje](#sistema-de-puntaje)
10. [Panel de Administrador](#panel-de-administrador)
11. [Preguntas Frecuentes](#preguntas-frecuentes)

---

## Introducción

**Quiniela Mundial 2026** es una aplicación de escritorio para pronosticar los resultados del FIFA World Cup 2026. Compite con otros jugadores prediciendo marcadores de los 104 partidos del torneo y acumula puntos según la precisión de tus pronósticos.

El Mundial 2026 cuenta con:
- **48 selecciones** clasificadas
- **12 grupos** (A–L) de 4 equipos cada uno
- **104 partidos** en total: 72 de fase de grupos + 32 de eliminatorias
- Sedes en **Canadá, México y Estados Unidos** (marcados con **(H)**)

---

## Registro e Inicio de Sesión

### Crear una cuenta nueva

1. Abre la aplicación y haz clic en **"Registrarse"**.
2. Ingresa un **nombre de usuario** único.
3. Ingresa y confirma tu **contraseña**.
4. Haz clic en **"Crear cuenta"**.

> **Nota:** El nombre de usuario es público y aparecerá en la tabla de posiciones general.

### Iniciar sesión

1. Ingresa tu **nombre de usuario** y **contraseña**.
2. Haz clic en **"Iniciar sesión"**.
3. Tu sesión se mantiene activa hasta que cierres la aplicación o hagas clic en **"Cerrar sesión"**.

---

## Pantalla Principal

Una vez dentro, verás el menú de navegación con las siguientes secciones:

| Sección | Descripción |
|---|---|
| **Grupos** | Tabla de posiciones y partidos de cada grupo (A–L) |
| **Horarios** | Partidos filtrados por selección |
| **Pronósticos** | Ingresa y consulta tus predicciones |
| **Eliminatorias** | Árbol de llaves desde 16avos hasta la Final |
| **Posiciones** | Ranking general de todos los jugadores |
| **Resultados** | *(Solo admin)* Carga de resultados reales |

---

## Fase de Grupos

La vista de grupos muestra los **12 grupos (A–L)** con sus posiciones y partidos.

### Navegar entre grupos

- Usa las **pestañas A, B, C … L** en la parte superior para cambiar de grupo.

### Tabla de posiciones del grupo

Cada grupo muestra una tabla con las siguientes columnas:

| Columna | Significado |
|---|---|
| **#** | Posición actual |
| **Equipo** | Nombre y bandera de la selección |
| **PJ** | Partidos Jugados |
| **G** | Ganados |
| **E** | Empatados |
| **P** | Perdidos |
| **GF** | Goles a Favor |
| **GC** | Goles en Contra |
| **DG** | Diferencia de Goles |
| **Pts** | Puntos |

**Colores de la tabla:**

- 🟢 **Verde** — 1.° y 2.° lugar (clasifican directamente a 16avos)
- 🟡 **Amarillo** — 3.° lugar (posible repechaje entre los 8 mejores terceros)
- ⬜ **Gris** — 4.° lugar (eliminado)

### Partidos del grupo

Debajo de la tabla de posiciones se muestran los **3 partidos** del grupo organizados por jornada. Cada tarjeta de partido muestra:

- Fecha, hora y ciudad del partido
- Equipos (local vs visitante) con sus banderas
- Marcador real (si el partido ya se jugó)
- Estado: **Pendiente** / **En juego** / **Finalizado**

---

## Horarios por País

Esta vista te permite ver todos los partidos de una selección específica.

### Cómo usar

1. Ve a la sección **"Horarios"**.
2. Selecciona una **selección** en el menú desplegable (las 48 selecciones están disponibles en español).
3. Se mostrará una tabla con todos los partidos de ese país, incluyendo:
   - Número de partido
   - Fecha y hora local
   - Equipos (local y visitante)
   - Estadio y ciudad
   - Fase del torneo

### Paginación

- La tabla muestra **10 registros por página** por defecto.
- Puedes cambiar la cantidad a 5, 10, 20, 50 o 100 registros usando el selector de filas.
- Usa los botones de navegación **◀ ▶** para cambiar de página.
- Las columnas se pueden ordenar haciendo clic en el encabezado.

---

## Pronósticos

### Ingresar un pronóstico

1. Ve a la sección **"Pronósticos"** o navega a la vista de grupos.
2. Busca el partido que deseas pronosticar.
3. Ingresa el marcador que predices en los **campos de goles** (local y visitante).
4. Haz clic en **"Guardar"** o **"Enviar pronóstico"**.

> ⚠️ **Importante:** Solo puedes ingresar o modificar un pronóstico **antes de que inicie el partido**. Una vez que el partido comienza, el pronóstico queda bloqueado.

### Modificar un pronóstico

- Puedes cambiar tu predicción tantas veces como quieras **hasta la hora de inicio del partido**.
- Después del inicio, el pronóstico queda fijo y no puede modificarse.

### Ver mis pronósticos

- En la vista de pronósticos puedes ver todos tus pronósticos comparados con los resultados reales ya disponibles.
- Los partidos que aún no han iniciado muestran tu predicción actual.
- Los partidos terminados muestran el resultado real junto a tu predicción y los puntos obtenidos.

---

## Eliminatorias

La vista de eliminatorias muestra el **árbol de llaves** desde 16avos de final hasta la Gran Final.

### Fases de eliminatorias

| Fase | Partidos |
|---|---|
| 16avos de Final | 32 equipos — 16 partidos |
| Octavos de Final | 16 equipos — 8 partidos |
| Cuartos de Final | 8 equipos — 4 partidos |
| Semifinales | 4 equipos — 2 partidos |
| Tercer Lugar | 2 partidos — 1 partido |
| Final | 2 equipos — 1 partido |

### Clasificados

- Hasta que no se definan los resultados de la fase previa, los cruces muestran **"1.° Grupo X"** o **"2.° Grupo Y"**.
- Una vez que una selección clasifica, aparece su nombre y bandera real en el árbol.

### Pronósticos en eliminatorias

- Para partidos de eliminatorias debes ingresar:
  1. El **marcador** que predices (en 90 minutos).
  2. El **equipo ganador** que avanza (en caso de empate al 90').
- El pronóstico se bloquea automáticamente al inicio de cada partido.

---

## Tabla de Posiciones

La sección **"Posiciones"** muestra el ranking general de todos los jugadores de la quiniela.

### Columnas de la tabla

| Columna | Descripción |
|---|---|
| **#** | Posición en el ranking |
| **Jugador** | Nombre de usuario |
| **Pts** | Puntos totales acumulados |
| **Exactos** | Cantidad de marcadores exactos acertados |
| **Resultados** | Cantidad de resultados acertados (sin marcador exacto) |

La tabla se ordena automáticamente por **puntos totales** de mayor a menor y se actualiza cada vez que el administrador carga un nuevo resultado.

---

## Sistema de Puntaje

### Puntaje base por partido

| Acierto | Puntos |
|---|---|
| **Marcador exacto** (ej: predijiste 2-1 y fue 2-1) | **3 puntos** |
| **Resultado correcto** (ej: predijiste 2-1 y fue 1-0, pero ganó el mismo equipo) | **1 punto** |
| **Resultado incorrecto** | **0 puntos** |

### Multiplicadores por fase eliminatoria

| Fase | Multiplicador |
|---|---|
| Fase de Grupos | x1 (sin multiplicador) |
| 16avos de Final | x2 |
| Octavos de Final | x2 |
| Cuartos de Final | x2 |
| Semifinales | x3 |
| Tercer Lugar | x3 |
| Final | x3 |

**Ejemplo:** Si predices el marcador exacto en una Semifinal → 3 pts × 3 = **9 puntos**.

---

## Panel de Administrador

> Esta sección es **exclusiva para usuarios con rol de administrador**.

### Acceso

El panel de administrador es accesible mediante una ruta protegida. Los usuarios normales no pueden ver ni acceder a esta sección.

### Cargar resultados

1. Ve a la sección **"Resultados"**.
2. Selecciona el **grupo** usando el menú desplegable.
3. Ingresa los goles de cada partido en las casillas correspondientes.
4. Haz clic en **"Actualizar Tablas"** para recalcular automáticamente:
   - La tabla de posiciones del grupo.
   - Los puntos de todos los jugadores de la quiniela.

> Los puntos de todos los participantes se recalculan automáticamente en cuanto se guarda un resultado.

### Gestión de usuarios

- El administrador puede consultar la lista de todos los usuarios registrados.
- Puede eliminar usuarios en caso necesario.

---

## Preguntas Frecuentes

**¿Puedo cambiar mi pronóstico después de haberlo guardado?**
Sí, siempre que el partido **no haya iniciado**. Una vez que comienza el partido, el pronóstico queda bloqueado.

**¿Qué pasa si no ingresé pronóstico para un partido?**
Los partidos sin pronóstico registrado reciben **0 puntos**, sin importar el resultado.

**¿Los puntos se actualizan en tiempo real?**
Los puntos se actualizan cada vez que el administrador carga el resultado de un partido. La tabla de posiciones refleja el estado más reciente disponible.

**¿Qué selecciones son consideradas locales (sede)?**
Las selecciones marcadas con **(H)** son los países anfitriones: **Canadá**, **México** y **Estados Unidos**.

**¿Cuántos partidos hay en total para pronosticar?**
El Mundial 2026 tiene **104 partidos** en total: 72 de fase de grupos y 32 de eliminatorias (incluyendo el partido por el tercer lugar).

**¿Cómo se desempata en la tabla de posiciones de la quiniela?**
En caso de empate en puntos totales, el desempate se resuelve por: mayor cantidad de marcadores exactos y luego por mayor cantidad de resultados acertados.

---

*Versión del manual: 1.0 — Mundial 2026*
