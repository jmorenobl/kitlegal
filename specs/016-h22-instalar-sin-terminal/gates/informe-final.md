# Informe del hito H22 · Instalar sin terminal: la extensión de escritorio con el servidor y el plugin de Claude con las skills (adelantado; ADR 0035)

Generado por el workflow `hito` el 2026-10-02T17:45:38Z, sobre `574380c` de `016-h22-instalar-sin-terminal`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/016-h22-instalar-sin-terminal/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `574380c` (medición 2): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 10 umbrales, **2 sin cumplir o solo publicados** (`expresiones_prohibidas:claude-haiku-4-5-20251001:orden`, `expresiones_prohibidas:claude-haiku-4-5-20251001:herramienta`); legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 4 rondas en 2 ciclos (el primero juzga el hito; cada uno de los siguientes, lo que cambió después del último veredicto).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 7 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 17 commits; 82 files changed, 25677 insertions(+), 194 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- plan: el valor de los textos de la ficha, que el spec deja al plan (FR-015) → nombre visible `kitlegal`; descripción corta, la que goreleaser ya da al cask, al bucket y a los paquetes, «Tu asistente de IA responde con la ley vigente del BOE y la cita exacta» (71 caracteres); un párrafo largo que dice qué hace, que todo corre en el equipo y que hace falta también el plugin (data-model §5); y autoría `kitlegal`, sin nombre de persona ni correo. Es lo que la persona ve en la ficha de la extensión y… (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: el nombre del catálogo, que el hito no fija → `kitlegal-plugins`, el del repositorio; el plugin se instala como `kitlegal@kitlegal-plugins`. Su `owner` es la autoría de los textos (research D5; data-model §6).
- plan: «El repositorio MUST llevar la plantilla de `.claude-plugin/marketplace.json`» (FR-030) → la plantilla es código, `internal/empaquetado/catalogo.go`, con tres valores que cambian, y no un fichero JSON con marcadores: los textos tienen que salir de su único sitio (FR-015), y un JSON los escribiría otra vez o pediría rellenarlo en shell, en dos flujos y sin test. El catálogo de cada etiqueta lo escribe `empaquetar catalogo`, que es también lo que valida `claude plugin validate` en la CI (res… (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: el README parte «Instalar» en tres apartados → «Instalar sin terminal» (nuevo, la oficial), «Instalar con la terminal» (el de hoy, con otro título) y «Otras instalaciones, sin probar»; el enlace de CONTRIBUTING a `README.md#instalar` pasa al segundo (contracts/documentacion.md §1).
- clarify Q3 (criterio c): Después de `humo`, y solo si sale en verde. `release.yml` pasa a tener tres trabajos: `publicar`, `humo` y uno nuevo que depende de `humo` y escribe `.claude-plugin/marketplace.json` en `jmorenobl/kitlegal-plugins`, con la `version` de la etiqueta sin la `v`, la dirección de `kitlegal-plugin.zip` en la release de esa etiqueta y el `sha256` que da para ese fichero el `checksums.txt` publicado. `PUBLISHER_TOKEN` lo ven dos pasos y ninguno más: el de goreleaser en `publicar… (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- plan: «Aceptación e2e» → no aplica: el hito no añade ni cambia ningún comportamiento del binario (FR-001, FR-080), así que no hay tarea `[aceptacion]` ni guiones congelados. Hacen de aceptación las subpruebas nuevas de `TestSnapshot` y `claude plugin validate` en el trabajo `snapshot` de la propuesta de cambio (SC-001), y la prueba humana de SC-002 (research D13).
- plan: el esquema oficial de la versión `0.3` del manifiesto (Clarifications, pregunta 2) → se leyó en local, del paquete `@anthropic-ai/mcpb` 2.1.2 de la caché de `npx` del equipo, solo para comprobar la forma del manifiesto del plan, que lo cumple; no se copia al repositorio ni se valida contra él en ningún control (research V13, D18).
- T006: el primer párrafo del README presenta kitlegal como «skills para Claude Code, Codex, Antigravity y cualquier agente que siga el estándar Agent Skills», y contracts/documentacion.md no lo nombra ni entre lo que cambia ni entre lo que deja de decirse → no se toca: no dice «oficial» ni «soportada» de ninguno, que es como el contrato concreta el MUST NOT de FR-051, y lo que está probado y lo que no lo dicen los tres apartados de instalación. Tampoco se tocan las frases de H19 de CONTRIBUTING s… (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: contracts/documentacion.md §1.3 dice qué se queda en «El servidor MCP», y dos frases de hoy dejan de ser ciertas con la extensión —«Necesita el programa instalado, y se declara una vez en cada agente» y que las skills «se instalan igual, con `kitlegal skills install`»— → la primera pasa a decir que en la app de escritorio de Claude el servidor lo trae la extensión y que en Claude Code necesita el programa y se declara una vez, y la segunda pierde su cola; lo demás del apartado queda como e… (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: «lo que la persona ve» en el paso dos y en la web y el móvil, que nadie ha probado con las piezas de una release (ADR 0035, «Pendiente de verificar», 3 y 5) → del paso dos el README da el menú y el modo, que es lo probado a mano el 2026-10-01 y el 2026-10-02, y no describe ninguna pantalla más; del marketplace dice que no está probado que la app lo admita y que, si no, se suba el zip; y la línea `⚠ SIN CONSULTA AL BOE:` va entera, como la llevan las dos `SKILL.md`, bajo la condición «si la… (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q1 (criterio d, conservadora): Los archivos de la release siguen siendo los seis de hoy y no se publica ninguno más: el binario universal de macOS solo viaja dentro de `kitlegal.mcpb`. `make snapshot-check` y `humo` leen de `server/kitlegal` sus arquitecturas, fallan si no son exactamente dos —una `amd64` y una `arm64`, reconocidas por su tipo de CPU y no por su posición— y comparan cada una con el binario `kitlegal` de `kitlegal_darwin_amd64.tar.gz` y de `kitlegal_darwin_arm64.tar.gz`: … (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q2 (criterio d, conservadora): El esquema oficial de la versión `0.3` no puede llegar a `testdata/` dentro del run, y el run ni lo versiona ni lo sustituye por uno propio: no hay tarea `[datos]`, `testdata/` no se toca y `docs/SOURCES.md` queda sin cambios. El control de FR-061 comprueba el manifiesto del snapshot con una comprobación propia del repositorio: que se declara de la versión `0.3`; que lleva cada campo de FR-012 con su valor y ninguno que FR-012 no nombre, y por tanto tampoco… (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- reparar_cierre: la medición del cierre sobre `4b350fa` da 1 de 54 respuestas del modelo que decide sin la skill activada en el modo herramienta (umbral 0; los otros nueve umbrales se cumplen y `legal-core` aprueba): la sesión 01-01 leyó el art. 21 de la Ley 39/2015 con `boe_articulo` sin activar `boe-legislacion` y sin llamar a `graph_check`, y el spec deja las dos skills «byte a byte» («Relación con H19 y H21», «Fuera de alcance», «Assumptions») → cambia la `description` de `boe-legislacion` (v… (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_revision: la frase que la reparación del cierre añadió a la `description` de `boe-legislacion`, «Actívala también antes de llamar a sus herramientas (boe_articulo…)», puede leerse como todas las herramientas de kitlegal —también `territorio_resolver`, que es de `legal-core`, y las de `graph`—, y la medición solo pide la activación antes de leer una norma con `boe_articulo` (revisión final, ronda 3, [i]; FR-003 de H7.4) → lectura conservadora: la frase queda «Actívala antes de llamar a … (entera en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-10-02T15:34:28Z · «fix(H22): motivos de la revisión final»: make ci sigue en rojo; lo verán el paso siguiente y el informe final.
- 2026-10-02T16:41:50Z · La revisión final abre el ciclo 2 para juzgar lo que cambió fuera de gates/ después del veredicto sobre a9c64f9: 41ff279 «fix(H22): cierre en la plataforma».

- Observaciones de los jueces, por debajo del umbral y sin corregir: 12 en `spec-r1.json`, 8 en `plan-r1.json`, 11 en `tasks-r1.json`, 10 en `revision-a-r4.json`, 9 en `revision-b-r4.json`.

### Internos (16)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (1), T002 (2), T003 (2), T004 (3), T005 (1), T007 (1), barrido (2), plan (4). Enteras en `specs/016-h22-instalar-sin-terminal/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `574380c`; decide `claude-sonnet-5-5` con 2 de 3; 21 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |

Respuestas con alguna expresión prohibida, en las evals que activan la skill (`expresiones_prohibidas_por_modelo`):

| Modelo | Con alguna | Respuestas | Porcentaje |
|---|---|---|---|
| `claude-sonnet-5-5 (orden)` | 0 | 54 | 0,0 % |
| `claude-haiku-4-5-20251001 (orden)` | 0 | 30 | 0,0 % |
| `claude-sonnet-5-5 (herramienta)` | 0 | 54 | 0,0 % |
| `claude-haiku-4-5-20251001 (herramienta)` | 0 | 30 | 0,0 % |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `expresiones_prohibidas:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 5,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_no_leida:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `expresiones_prohibidas:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 5,0 % | sí | no: solo se publica, no es un control |
| `expresiones_prohibidas:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 5,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_no_leida:claude-sonnet-5-5:herramienta` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `expresiones_prohibidas:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 5,0 % | sí | no: solo se publica, no es un control |
| `duracion_de_las_sesiones:orden` | 441 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 448 | ≤ 900 | sí | sí |

**legal-core**: aprobado sobre `574380c`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |

Umbrales: el job no publica ninguno para esta skill.

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

## 4. Revisión que la constitución reserva a la persona

Nada: el hito no toca fuentes, anomalías, adaptadores, fixtures, esquemas, grabaciones ni evidencias.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T002(hecha) | — | comprobado por su control | `ci:internal/arch_test.go:TestElBinarioNoEnlazaElPaso`, en make ci (verde) |
| FR-002 | T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-003 | T005(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-004 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezasReproducibles`, en make ci (verde) |
| FR-005 | T002(hecha) T003(hecha) | — | tareas hechas | — |
| FR-006 | T003(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-010 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-011 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-012 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-013 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-014 | T001(hecha) T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/ejecutar_test.go:TestEjecutar`, en make ci (verde); `ci:internal/app/herramientas_test.go:TestHerramientasAnunciadas`, en make ci (verde) |
| FR-015 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/textos_test.go:TestDescripcionCorta`, en make ci (verde) |
| FR-016 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestIcono`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-017 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-020 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/ejecutar_test.go:TestEjecutar`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-021 | T002(hecha) | — | tareas hechas | — |
| FR-022 | T002(hecha) | — | tareas hechas | — |
| FR-023 | T004(hecha) | — | tareas hechas | — |
| FR-030 | T002(hecha) T004(hecha) | — | tareas hechas | — |
| FR-031 | T002(hecha) T005(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-032 | T005(hecha) | — | tareas hechas | — |
| FR-040 | T005(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-041 | T005(hecha) | — | tareas hechas | — |
| FR-042 | T005(hecha) | — | tareas hechas | — |
| FR-043 | T005(hecha) | — | tareas hechas | — |
| FR-050 | T006(hecha) | — | tareas hechas | — |
| FR-051 | T006(hecha) | — | tareas hechas | — |
| FR-052 | T006(hecha) | — | tareas hechas | — |
| FR-053 | T006(hecha) T007(hecha) | — | tareas hechas | — |
| FR-060 | T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-061 | T001(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/ejecutar_test.go:TestEjecutar`, en make ci (verde); `ci:internal/app/herramientas_test.go:TestHerramientasAnunciadas`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-062 | T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-063 | T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/ejecutar_test.go:TestEjecutar`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-064 | T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-065 | T004(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-066 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/textos_test.go:TestDescripcionCorta`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestIcono`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| FR-067 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezasReproducibles`, en make ci (verde) |
| FR-068 | T003(hecha) T004(hecha) T005(hecha) T007(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| FR-069 | T001(hecha) T002(hecha) T003(hecha) T007(hecha) | — | tareas hechas | — |
| FR-070 | T002(hecha) T007(hecha) | — | tareas hechas | — |
| FR-080 | T001(hecha) T007(hecha) | — | tareas hechas | — |
| SC-001 | T007(hecha) | — | tareas hechas | — |
| SC-002 | T007(hecha) | — | tareas hechas | — |
| SC-003 | T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| SC-004 | T001(hecha) T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/ejecutar_test.go:TestEjecutar`, en make ci (verde); `ci:internal/app/herramientas_test.go:TestHerramientasAnunciadas`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| SC-005 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| SC-006 | T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-007 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/ejecutar_test.go:TestEjecutar`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| SC-008 | T004(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde) |
| SC-009 | T002(hecha) T003(hecha) | — | comprobado por su control | `ci:internal/empaquetado/textos_test.go:TestDescripcionCorta`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestIcono`, en make ci (verde); `ci:internal/empaquetado/piezas_test.go:TestPiezas`, en make ci (verde) |
| SC-010 | T002(hecha) | — | comprobado por su control | `ci:internal/empaquetado/piezas_test.go:TestPiezasReproducibles`, en make ci (verde) |
| SC-011 | T003(hecha) T005(hecha) T007(hecha) | — | tareas hechas | — |
| SC-012 | T006(hecha) | — | tareas hechas | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`874d502`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `a9c64f9` fix(H22): motivos de la revisión final (make ci en rojo): lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/empaquetado/catalogo.go`, `internal/empaquetado/catalogo_test.go`, `internal/empaquetado/export_test.go`, `contracts/paso.md`.
- `d882a2c` docs(H22): veredictos de la revisión final: solo registros de `gates/`.
- `4b350fa` docs(H22): registros del run: solo registros de `gates/`.
- `41ff279` fix(H22): cierre en la plataforma: lo vio la ronda 3 (juez A: aprobado; juez B: rechazado). Toca `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `plan.md`, `research.md`.
- `b960a9b` fix(H22): motivos de la revisión final: lo vio la ronda 4 (juez A: aprobado; juez B: aprobado). Toca `CHANGELOG.md`, `skills/boe-legislacion/SKILL.md`, `plan.md`, `research.md`.
- `6cca537` docs(H22): veredictos de la revisión final: solo registros de `gates/`.
- `574380c` docs(H22): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/016-h22-instalar-sin-terminal/quickstart.md`. Suite de aceptación congelada: `specs/016-h22-instalar-sin-terminal/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `4849fe7b`: 8 h 2 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 4849fe7b`.
