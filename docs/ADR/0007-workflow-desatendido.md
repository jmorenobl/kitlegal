# 0007 · Workflow desatendido: unión de motivos, jueces en paralelo, publicación de la rama y pausa humana acotada

- **Estado**: aceptada
- **Fecha**: 2026-09-12
- **Hito**: transversal (workflow `hito` 1.6.0, antes de H2)

## Contexto y problema

El run de H1 (`ad3d19d6`, 2026-09-11) duró 11 h de reloj. Los pasos con modelo sumaron 7,3 h; la
batería `make ci` por tarea, 4 min en total; y **3,4 h fueron huecos entre una parada y la reanudación a
mano**. El run se detuvo cuatro veces, y solo una de ellas era una decisión que la constitución reserva a
una persona:

| Parada | Causa | Era humana |
|---|---|---|
| `juez_spec` | Fable devolvió «monthly spend limit»; el motor de spec-kit no reintenta ni cambia de modelo | No |
| `commit_artefactos_plan` | El paso shell salió con 1; al reanudar pasó limpio | No |
| `siguiente_tarea` (T021) | Tarea que exige `git push` y una propuesta de cambio, denegados por `.claude/settings.json`; quemó tres intentos antes de parar | Sí, por política propia; y predecible desde `tasks.md` |
| `ci_final` | El juez A aprobó y el juez B rechazó con 13 motivos; la regla «desacuerdo → humano» paró el run | No: los 13 motivos eran corregibles y el corrector los aplicó sin discutir ninguno |

La revisión final aportó además dos datos sobre los propios jueces. En la ronda 1, A (`sonnet@max`, 3 min)
no encontró nada y B (`opus@max`, 26 min) encontró 13 defectos reales: el desacuerdo medía la profundidad
de cada juez, no la ambigüedad del código. En la ronda 2, con las correcciones aplicadas, la situación se
invirtió: A rechazó con cinco motivos nuevos (tres de ellos sustantivos) y B aprobó. «Ambos aprueban» no
era un criterio de parada estable, y «desacuerdo → humano» convertía cada ronda en una parada.

Por último, las dos pausas `[datos]` de H1 fueron por material de test sintético (código Go y guiones
`.txtar` bajo `internal/app/testdata/`), que no es verdad terreno de ninguna fuente. La constitución (Capa 3)
reserva a la persona «toda modificación de ficheros existentes en `testdata/` y `schemas/` y toda grabación
de fixtures»; el workflow pausaba ante cualquier fichero bajo cualquier `testdata/`, más de lo que la
constitución pide.

## Opciones consideradas

**Sobre la revisión final.**

1. Mantener «desacuerdo → humano». Rechazada: en H1 habría exigido una persona en cada ronda, y la
   evidencia dice que el desacuerdo describe a los jueces.
2. Un solo juez profundo. Rechazada: se pierde el punto ciego cruzado que justificó dos prompts distintos
   (uno adversarial) y dos familias de modelo.
3. **Corrector sobre la unión de motivos** cuando cualquiera de los dos rechaza con motivos corregibles;
   humano solo si algún motivo es no corregible (alcance, frontera humana, privacidad, TOS, anomalías,
   decisión cerrada) o se agotan las rondas. Los dos jueces corren **en paralelo** (`fan-out` del motor,
   `max_concurrency: 2`) sin ver el veredicto del otro, con la misma capacidad de modelo y la diversidad
   puesta en el prompt y en la familia. Elegida.

**Sobre las paradas que no son decisiones.**

1. Añadir reintentos al motor de spec-kit. No está en nuestra mano: el motor solo ofrece
   `continue_on_error` y `on_reject: retry` en gates.
2. Un agente supervisor con criterio propio que «resuelva lo que se pare». Rechazada tal cual: sería una
   cuarta capa sin rúbrica que podría decidir lo que la constitución reserva a una persona.
3. **Un supervisor determinista** (`scripts/hito.sh`) con una lista cerrada de acciones: ante límite de
   uso o error de la API, cambia de familia de modelo en los roles afectados y reanuda; ante un fallo
   transitorio de commit o de sesión, reanuda una vez; ante un `check_*`, `guardian_*`, `leer_*`,
   intentos agotados o `ci_final`, se detiene y explica; ante un gate, se detiene. Nunca edita
   artefactos, veredictos ni permisos. Elegida.

**Sobre la plataforma remota.**

1. Mantener `git push` y `gh` fuera del alcance de Claude. Rechazada: toda tarea de evidencia en la
   propuesta de cambio termina en un prerrequisito humano (T014 de H0, T021 de H1) y el hito nunca queda
   publicado sin una persona.
2. **Permitir empujar la rama del hito y crear o leer la propuesta**, y seguir reservando a la persona la
   fusión, el push a `main`, los push forzados, los borrados y las etiquetas. Dos capas: los permisos de
   `.claude/settings.json` filtran las órdenes por prefijo, y `scripts/lefthook/pre-push/guardia-push.sh` (gancho `pre-push` de
   lefthook, con `use_stdin`) rechaza esas referencias vean lo que vean los permisos. Las tareas que
   necesitan la plataforma llevan `[plataforma]`, van al final, y si los permisos no están concedidos el
   workflow pausa antes de gastar intentos. Al final del hito, `publicar_rama` empuja la rama y abre la
   propuesta si no existe. Elegida.

**Sobre la pausa `[datos]`.**

1. Pausar ante cualquier fichero bajo `testdata/`. Rechazada: excede la constitución y detiene el run por
   material que los jueces finales ya revisan (fixtures retocados, tests vacíos).
2. **Pausar exactamente por lo que dice la Capa 3**: modificación de ficheros existentes bajo `testdata/`
   o `schemas/`, o ficheros nuevos en el territorio de fixtures grabados y esquemas (`testdata/` de raíz,
   `internal/source/**`, `schemas/`). Material de test nuevo fuera de ese territorio sigue exigiendo la
   etiqueta `[datos]` y el guardián de diff, pero no pausa. Elegida.

## Decisión

Workflow `hito` 1.6.0, constitución 1.3.0 (Capa 2: «cualquier rechazo corregible de cualquiera de los dos
jueces va al corrector con la unión de los motivos; solo un motivo no corregible o el agotamiento de las
rondas escala a humano»; Flujo §4: «empujar la rama del hito y abrir la propuesta de cambio lo hace el
workflow»). Detalle operativo en `docs/WORKFLOW.md`.

Además, los autores de `spec`, `plan` y `tasks` reciben la instrucción de comprobar su artefacto contra
la rúbrica del juez que lo evaluará: en H1 los tres gates fueron rechazados exactamente una vez y
aprobados tras corregir; si el autor se autoevalúa, el juez sigue juzgando igual y la ronda extra puede
desaparecer sin bajar el listón.

## Consecuencias

- Un run sin decisiones humanas pendientes termina con la rama empujada y la propuesta abierta; fusionar
  sigue siendo una acción humana, mecánicamente garantizada por el gancho `pre-push`.
- Las paradas que quedan son las de la constitución: clarificaciones escaladas, motivos no corregibles,
  gates `[datos]` con material existente o fixtures grabados, rutas sensibles, y el modo supervisado.
- La revisión final puede dar hasta tres veredictos y dos correcciones antes de escalar; el coste de una
  ronda es el del juez más lento más el corrector, no la suma de los dos jueces.
- Quien relance un juez a mano usa `scripts/paso.sh revision_juez_a|b <hito>`; `scripts/coste-run.sh`
  atribuye las sesiones paralelas por familia de modelo.
- El cambio de permisos en `.claude/settings.json` es una decisión de la persona y se aplica a mano; hasta
  entonces una tarea `[plataforma]` pausa el run en lugar de fallar tres veces.
