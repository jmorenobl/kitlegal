<!-- Propuesta de cambio de H5. La escribió la tarea de cierre (T029, intento 3) con la medida hecha sobre 536359c (`feat(H5): T032`) el 2026-09-15; las tareas de plataforma (T030 y T031) añaden lo que digan la integración continua, Codecov, la prueba de red y la ejecución de cierre, y lo comprobable del supuesto S8. Sustituye a la medida del intento 2 sobre be901f7, cuya verificación cayó en un test intermitente de internal/httpx que T032 arregló (gates/tarea-T029.md, gates/tarea-T032.md). -->

## Objetivo

«Que un agente en Claude Code sepa consultar y citar cualquier norma consolidada del BOE (administrativa,
contratación, local, fiscal…), y dejar el andamiaje con el que se miden todas las skills siguientes»
(`docs/ROADMAP.md` §4, H5). H4 dejó el applet `boe`: la herramienta determinista. **H5 es el primer hito que entrega
una skill**, `boe-legislacion`, y el andamiaje que compartirán todas las que vengan: una tabla de normas como única
fuente de verdad, lo generado comprobado en `make ci`, la instalación con una orden, el formato común de eval y un
job de evals con Claude Code que no deja llegar ninguna petición a la red de una fuente.

Tres compromisos, y los tres se comprueban sin red, sin modelo y sin Python:

1. **La skill es el protocolo de `boe-fiscal` generalizado a cualquier materia.** `SKILL.md` (159 líneas) identifica
   la norma, resuelve `BOE-A-…`, lee índice y bloques con `scripts/boe`, responde citando `[BOE-A-…, bloque …]`,
   distingue ley y reglamento y señala la variación autonómica; sus reglas prohíben concluir «no existe», inventar
   contenido legal, presentar como vigente lo derogado y actuar en nombre de nadie (FR-001 a FR-015, SC-004).
2. **Nada generado diverge de su fuente de verdad.** `references/normas.md` sale de `data/normas.yaml`; la tabla de
   comandos, de `--describe`; los enlaces de `scripts/`, de `metadata.kitlegal-applets`. `make skills-sync` lo escribe
   y `make skills-check`, dentro de `make ci`, lo regenera en memoria y falla nombrando la skill y el fichero, la
   entrada o el enlace (FR-030 a FR-044, SC-005). Los identificadores de la tabla se verifican contra búsquedas grabadas
   del BOE (SC-008).
3. **La skill se mide con evals comparadas por identificador.** Doce evals (diez positivas de materias distintas, una
   de ellas reproducción de una consulta de `boe-fiscal`, y dos de no activación) validadas contra `schemas/eval.yaml.json`;
   lo grabado basta para cada una y `make ci` lo comprueba con `--offline` sobre una caché preparada (SC-010). El job
   `evals` abre una sesión de Claude Code por eval bajo `strace`, con la red cerrada salvo la del modelo y sin Python
   en el runner, y un informe sin modelo juzga activación, comandos, citas y conexiones (FR-070 a FR-082, SC-012).

Hito de **skill** (principio VIII, ADR 0012): la base genérica de la que `boe-fiscal` y las demás verticales se
especializan desde el backlog. El binario distribuido no cambia.

## Alcance

Frente a `main`, en `536359c` (`feat(H5): T032`, la cabeza de la rama y el último commit que toca código, datos o
documentación): 360 ficheros, 26 551 líneas añadidas y 77 retiradas, en 34 commits. Fuera de `specs/`: 307 ficheros,
20 295 añadidas y 77 retiradas. Por árboles:

- **`skills/boe-legislacion/`, la skill** (3 ficheros): `SKILL.md` (frontmatter con `name`, `description` y
  `metadata.kitlegal-applets: boe` y `metadata.kitlegal-referencias: normas`; protocolo en cinco pasos; cómo se cita;
  la región de comandos generada; cinco reglas), `references/normas.md` (generado, con la cabecera «generado desde
  data/normas.yaml, no editar») y `scripts/boe -> ../../../bin/instalado/kitlegal`.
- **`data/normas.yaml`**, diez normas indexadas por identificador (LPAC, LCSP, LRBRL, LGT, TRLRHL, LIRPF, LRJSP,
  LTAIBG, CE, ET), cada una con título, rango, abreviatura y materias, sin campo `vertical`; y sus dos esquemas nuevos,
  `schemas/normas.yaml.json` (con el `enum` de `rango` copiado de las búsquedas grabadas) y `schemas/eval.yaml.json`.
- **`evals/boe-legislacion/`** (12 ficheros): `01-lpac-articulo-21` … `10-et-vacaciones` (positivas) y
  `11-no-activa-programacion`, `12-no-activa-acuerdo-entre-amigos`; la 06 lleva `reproduce: boe-fiscal`.
- **`internal/skills`, paquete nuevo de herramienta** (9 ficheros de producto, 9 de test, 4 guiones `testscript`):
  `skill.go`, `frontmatter.go`, `esquemas.go` (el lector común de YAML, que rechaza claves repetidas), `normas.go`,
  `referencias.go`, `comandos.go`, `enlaces.go`, `sincronia.go`, `doc.go`; `instalacion_test.go` con `TestInstalacion`
  (etiqueta `integration`) sobre `testdata/script/instalar*.txtar`.
- **`internal/evals`, paquete nuevo de herramienta** (11 ficheros de producto, 12 de test): `formato.go`,
  `conjunto.go`, `consultas.go`, `grabaciones.go`, `preparar.go`, `trazas.go`, `sesion.go`, `citas.go`, `juzgar.go`,
  `informe.go`, `doc.go`; los arneses `grabacion_test.go` (etiqueta `grabacion`) y `job_test.go` (etiqueta `evals`);
  y **195 ficheros de sesiones, trazas e informes sintéticos** en `testdata/sesiones/` (`leer-sesion`, `leer-trazas`,
  `informe`), fijados en tres tareas `[datos]`.
- **`internal/app/skills_test.go`**: `TestSkillsDelRepositorio` y `TestTablaDeComandosCoincideConLaGramatica`, el
  único fichero nuevo en el paquete del applet (usa `describir` sin exportar nada del kernel).
- **Datos grabados de H5** bajo `testdata/evals/`: el manifiesto `grabaciones.json` y 32 respuestas de la API de
  Legislación Consolidada en `boe.legislacion-consolidada/`, grabadas por una persona en la pausa de T008 (commit
  `1f4feea`); ninguna con el nombre de una de H4.
- **Guiones** (4, modo `100755`): `scripts/skills-sync.sh`, `scripts/instalar-skills.sh`, `scripts/grabar-evals.sh`
  y `scripts/evals.sh`. **Flujo**: `.github/workflows/evals.yml` (`name: evals`; manual con la entrada
  `prueba_de_red`, semanal sobre la rama principal y por las etiquetas `evals` y `evals-prueba-de-red`;
  `ubuntu-24.04`; Claude Code 2.1.270 y `strace`; el paso «Retirar Python del runner»; modelo
  `claude-haiku-4-5-20251001`; `CLAUDE_CODE_OAUTH_TOKEN`; la sesión, con `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0`).
- **Ficheros de H0-H4 tocados**, y ninguno más (nueve): `Makefile` (`install` enlaza las skills; `skills-sync` real;
  `skills-check` nuevo, en `ci`, que pasa de nueve a diez controles; `evals` nuevo, fuera de `ci`); `.golangci.yml`
  (la etiqueta `evals` en `run.build-tags` y `comando`, `comandos`, `defectos`, `legislativo` y `patrones` en
  `misspell.ignore-rules`, cada una con su motivo); `go.mod` (solo `go.yaml.in/yaml/v3` pasa de indirecta a directa);
  `docs/PENDIENTES.md` (se retiran las tres entradas «En H5»); `README.md`, `CONTRIBUTING.md` y `CHANGELOG.md`; y,
  por T032, `internal/httpx/reintentos_test.go` e `internal/httpx/ritmo_test.go` (41 líneas añadidas y 27 retiradas,
  solo en la medida de las llegadas y sus comentarios: ni el decorador de ritmo, ni el de reintentos, ni el cubo, ni la
  cadena del cliente, ni otro test).
- **Nuevo fuera de esos árboles**: `.agents/.gitattributes` (una línea).
- **Artefactos del hito**: `specs/006-h5-skill-boe-legislacion/` (53 ficheros: spec con sus clarificaciones, plan,
  research D1-D23 y V1-V64, data-model, seis contratos, quickstart, tasks, checklist y `gates/` con las notas de
  T009, T011, T012, T029 y T032).

**Sin cambios**, como exige el spec: `internal/core`, `internal/cli`, `internal/cache`, `internal/source/boe` y sus
esquemas, `docs/SOURCES.md`, el código de producto de `internal/httpx` (T032 solo toca dos de sus tests),
`internal/app` salvo el test nuevo, `.claude/skills/`, `skills-lock.json` y el contenido de `.agents/skills/`
(escenario 11: `git diff --stat main...HEAD` sobre esos tres últimos está vacío). `go.sum` y `codecov.yml` no aparecen
en el diff.

**Fuera de alcance** (spec, *Fuera de alcance*): el campo `vertical`, `boe-fiscal` y su migración (backlog
«Verticales»); la tabla completa de leyes vertebrales, `legal-core` y `territorio` (H6); `graph check` (H7); el applet
`cita` y `cita-verificada` (H8); verbos nuevos de `boe`, filtro por departamento y campo ELI; `kitlegal skills …`,
packs, plugin, MCP, release; respuesta esperada textual en las evals; juez de modelo como gate; cambios en la fuente o
en el kernel; grabar dentro del bucle de implementación o del job; ADR nuevo (el plan no se aparta de ninguna
decisión existente). Puntos 4, 7, 8 y 12 de la Definition of Done: no aplican; el 11, solo en lo que prohíbe (ningún
caso especial de municipio en la skill, los datos ni las evals).

## Dependencias (constitución §V)

**Una dependencia directa nueva, ya presente en el grafo: `go.yaml.in/yaml/v3 v3.0.5`.** `go.mod` solo la pasa de
`// indirect` a directa; `go.sum` no cambia y ningún módulo entra en el grafo (research V31). Justificación
(*Complexity Tracking* del plan): hay que leer YAML (`data/normas.yaml`, las evals y el frontmatter de `SKILL.md`), y
el lector común del hito necesita `yaml.Node` con líneas y `(*yaml.Node).Decode` con su comprobación de claves
repetidas (V43). Es el mismo código que `gopkg.in/yaml.v3`, la entrada literal de la lista de §V, mantenido por la
organización YAML desde que `go-yaml` quedó sin mantenimiento; hoy congelado como legado con solo correcciones de
seguridad (V58). Alternativas rechazadas: *`gopkg.in/yaml.v3`*, literalmente en la lista pero sin mantenimiento y con
un módulo más en `go.sum`; *`go.yaml.in/yaml/v4`*, ya en el grafo por `internal/cli`, pero sin ninguna versión
estable (solo candidatas), con la API y el texto del error de clave repetida cambiados entre la candidata del grafo y
la última, y con una sola versión para las herramientas y para el binario, de modo que subirla para unas movería la
del otro (V58, D8). Solo la usan `internal/skills` e `internal/evals`, que el binario distribuido no enlaza:
`TestDependenciasDelBinario` y `TestArquitectura` siguen sin cambios y en verde.

**Herramientas fuera del módulo, solo en el job** (*Complexity Tracking*): **Claude Code 2.1.270**, porque FR-070
pide ejecutar las evals con Claude Code; **`strace`**, porque FR-072 pide cada invocación con su código y sus
conexiones, que la herramienta Bash de Claude Code no da (V9) y que ni ganchos ni transcript aportan (V13); el runner
**`ubuntu-24.04`**, porque `strace` es de Linux y la imagen trae lo que el paso de retirada de Python necesita (V56,
V57). Alternativa rechazada: `anthropics/claude-code-action`, no verificable en local y sin traza. El **secreto
`CLAUDE_CODE_OAUTH_TOKEN`** es el token de la suscripción de Claude que genera `claude setup-token` (el proyecto no usa
clave de API de pago por uso; spec, *Assumptions*) y las **etiquetas `evals` y `evals-prueba-de-red`** disparan el job
sobre la rama del hito antes de fusionar (FR-070, D13), porque `gh workflow run --ref` no está documentado para un
fichero que aún no está en la rama principal (V14). Ninguna entra en `make ci`; el secreto y las etiquetas los da de
alta una persona y T030 comprueba que existen.

**Sin cambio de versión** en las herramientas de control de H0-H4 (`golangci-lint`, `govulncheck`, `gitleaks`,
`lefthook`, `santhosh-tekuri/jsonschema/v6`, `testscript`). `.golangci.yml` no añade ninguna exclusión: la etiqueta
`evals` mete `job_test.go` bajo el lint (V38), y las cinco palabras de `misspell` responden cada una a una palabra que
el código tiene que escribir suelta por contrato (D20).

## Controles añadidos

Los 22 controles de la tabla del plan («Controles mecánicos que este hito añade o toca») están en el árbol; 21 en
`make ci` y el último, el job de evals con modelo, fuera de él por diseño. Lo que pasa a ser mecánico, con el escenario
del quickstart que lo demuestra:

- **`make skills-check`, nuevo y dentro de `make ci`**, que pasa de nueve a diez controles: una sola invocación de
  `go test` con `TestSkillsDelRepositorio`, `TestNormasDelRepositorio`, `TestEvalsDelRepositorio` y
  `TestIdentificadoresDeLasNormas` sobre `internal/app`, `internal/skills` e `internal/evals` (escenario 1).
- **Frontmatter del estándar Agent Skills** (`/frontmatter`, `TestValidarFrontmatter`) y **`SKILL.md` < 300 líneas**
  (`/trescientas-lineas`, `TestContarLineas`): un `name` con mayúscula falla nombrando el defecto y el directorio; 300
  líneas fallan con «tiene 300 líneas (máximo 299)» (escenario 4).
- **Deriva de `references/`, de la tabla de comandos y de los enlaces de `scripts/`** (`/skills`,
  `/referencia-editada`, `/datos-sin-regenerar`, `/describe-cambiado`, `/enlaces`; `TestRegenerarYComparar`,
  `TestRenderizarTabla`, `TestSustituirRegion`, `TestEnlacesEsperados`): una línea a mano en `normas.md`, un título
  cambiado en `data/normas.yaml` o una ayuda de verbo cambiada sin regenerar → `contenido-distinto` nombrando el
  fichero (escenario 3); un enlace ausente, con otro destino o sobrante → `enlace-ausente`,
  `enlace-con-otro-destino (apunta a …)`, `enlace-sobrante` (escenario 4).
- **Regeneración determinista e idempotente** (`/regenerar-dos-veces`; `make skills-sync` dos veces sobre un clon
  limpio no cambia nada, escenario 2) y **tabla contra gramática** (`TestTablaDeComandosCoincideConLaGramatica`, que
  invoca cada sintaxis generada contra `--describe`).
- **Normas contra esquema e identificadores contra búsqueda grabada** (`TestLeerNormas`, `TestEsquemaDeNormas`,
  `TestNormasDelRepositorio`, `TestIdentificadoresDeLasNormas`): `vertical` → «campo no declarado: vertical»;
  `BOE-A-15-10565` → «identificador con otra forma» (escenario 5); un título que no coincide con la búsqueda grabada
  → «el título no coincide con la búsqueda grabada: …» (escenario 3, SC-008). **Gramáticas iguales a las de `boe`**
  (`TestGramaticasCoincidenConBoe`).
- **Formato y conjunto de evals** (`TestLeerEval`, `TestEsquemaDeEval`, `TestLeerConjunto`, `TestConjuntoDeEvals`,
  `TestEvalsDelRepositorio/formato`, `/conjunto`, `/normas-conocidas`): sin `pregunta` → «missing property
  'pregunta'» nombrando el fichero, y el conjunto deja de tener diez positivas y la del art. 21 (escenario 6).
- **Lo grabado basta para cada eval y cada sesión se prepara con las consultas de todas** (FR-074, FR-075;
  `TestEvalsDelRepositorio/grabado`, `TestPrepararYComprobar`, `TestPrepararDirectorioDeSesion`): retirar los
  metadatos de la LPAC → falla nombrando la eval y el comando `boe articulo BOE-A-2015-10565 a21`, primero al preparar
  y después al comprobar con `--offline` (escenario 6, SC-010).
- **Manifiesto bien formado y grabaciones sin solape con H4** (`TestManifiestoDeGrabaciones`,
  `TestGrabacionesSinSolape`): dos entradas con el mismo `titulo_empieza_por` → «entrada 1 … y entrada 2 …: la misma
  norma repetida, con el mismo prefijo», también desde `TestIdentificadoresDeLasNormas` (escenario 6).
- **Comparación mecánica** (`TestInterpretarInvocacion`, `TestExtraerCitas`, `TestJuzgar` con 18 subtests): una cita
  de otro bloque o de otra norma no satisface (`sc-009-otro-bloque`, `sc-009-otra-norma`); un bloque leído con código
  4 o sin código no cuenta; `--describe` y `--dry-run` no satisfacen; una sesión sin terminar no pasa ni en las de no
  activación (escenario 7, SC-009).
- **Lectura de sesión, trazas e informe** (`TestLeerSesion` 15, `TestLeerTrazas` 21, `TestLeerTrazasSinFicheros` 2,
  `TestInforme` 13, `TestEscribirInformeSinSusEntradas` 6): hilos por `clone`, `clone3` y de un hilo, ficheros sin
  origen, líneas de señal, sesiones cortadas por el tope con y sin llamada interrumpida, conexiones públicas en curso
  en IPv4 e IPv6, códigos 124 y 137, transcripts sin `result` o con `is_error`, cabecera del informe byte a byte,
  motivos de la raíz en su orden, y ningún informe escrito si una entrada no se puede leer (escenario 7).
- **`make install`** (`TestInstalacion`, etiqueta `integration`, cuatro guiones `testscript`; `make test-integration`
  la nombra): instala el binario, crea `bin/instalado/kitlegal` y enlaza la skill; repetirlo no cambia nada; una
  entrada ajena en conflicto termina en 2 y no se toca (escenario 8, SC-006).
- **Sin instrucciones de evals en la skill** (`/sin-instrucciones-de-evals`; escenario 9, FR-077) y **normas
  nombradas en `SKILL.md`** con entrada en datos (`/normas-nombradas`, FR-020).
- **Lint de los ficheros etiquetados**: `run.build-tags: [integration, fuentes, grabacion, evals]`.
- **Sin Python** en la skill, los datos, las evals, los esquemas, los guiones, el `Makefile` y el flujo (escenario 10,
  SC-007); en el job, el paso «Retirar Python del runner» y la comprobación como root antes de la primera sesión, cuya
  evidencia registran T030 y T031.
- **`.agents/.gitattributes`**: `skills/** linguist-vendored linguist-generated`, que git aplica exactamente a
  `.agents/skills/` y a nada más (escenario 11, `git check-attr`).
- **La medida del ritmo de `internal/httpx` con la cota exacta del limitador** (T032;
  `internal/httpx/reintentos_test.go`, `internal/httpx/ritmo_test.go`): cada llegada al servidor se mide contra un
  instante tomado antes de la primera petición de la operación y no puede adelantarse a su turno —la llegada k, k
  intervalos tras el comienzo, con el permiso del sitio en el turno cero; el intento n de los reintentos, n
  intervalos—, sin holgura. Sustituye a la comparación de cada par de llegadas consecutivas con «un intervalo menos
  una holgura» (10 ms y 25 ms), que exigía más que el contrato del limitador —fija turnos, no llegadas— y cayó de forma
  intermitente en la verificación del intento 2 de T029 (85,39 ms frente a 90 ms). Comprobado sobre una copia
  desechable con mutantes (`gates/tarea-T032.md`): el despacho tardío (una implementación correcta) hacía caer la
  aserción antigua 10 de 10 y pasa con la nueva; los reintentos por debajo del ritmo, la ráfaga de dos tokens y la
  ausencia de espera caen con la nueva. Es el único cambio de H5 fuera de lo que enumera el spec, y no toca código de
  producto.
- Los de H0-H4 (R1-R5, formato, `-race`, `govulncheck`, `gosec`, `gitleaks`, `go mod verify`, `tidy -diff`, CodeQL,
  `schema-check`, `TestDependenciasDelBinario`, `TestArquitectura`) siguen sin exclusiones nuevas.
- **Evals con modelo** (`.github/workflows/evals.yml` → `make evals SKILL=boe-legislacion`): fuera de `make ci`; la
  ejecuta la plataforma por etiqueta, a mano o cada semana (T030, T031).

## Evidencia

Medida por T029 (intento 3) el 2026-09-15 (entre las 04:28 y las 04:35, hora de Madrid) sobre **536359c**
(`feat(H5): T032`, la cabeza de la rama), ejecutando `make ci` y después `quickstart.md` entero salvo §12
(plataforma), con cada orden tal cual, desde los prerrequisitos hasta la limpieza, en una sesión desatendida. El
`quickstart.md` es el que corrigió el intento 1 de la misma tarea (`gates/tarea-T029.md`: `wc` sobre un fichero de la
carpeta temporal está bloqueado por Claude Code fuera del directorio de trabajo, también como orden sola; pasa a
`rtk proxy wc -l`, con una sonda nueva en los prerrequisitos; ningún test, patrón ni filtro cambió), ya confirmado en
el árbol. **Ninguna orden pidió aprobación** y las tres sondas de los prerrequisitos dieron lo esperado (`código 0`,
`sin HOME temporal todavía`, `1 /tmp/kitlegal-quickstart-h5/repo/.agents/.gitattributes`). Los escenarios 1, 2, 7, 9,
10 y 11 se lanzaron a la vez (no comparten ficheros: el 2 trabaja en el clon y los demás solo leen el árbol) y después
3, 4, 5, 6 y 8 en orden sobre el clon, con una sonda de `git status --porcelain` del clon antes de cada uno, que dio
siempre vacío.

| Escenario | Resultado |
|---|---|
| `make ci` | `0 issues.`; los doce paquetes en `ok` en los dos perfiles (`-race -shuffle=on` y `-race -tags=integration`), `internal/httpx` incluido (97,2 %); `govulncheck` «No vulnerabilities found»; `schema-check` y `skills-check` en `ok`; `gitleaks` «no leaks found»; `go mod verify` en la raíz y los cuatro módulos de herramienta; `tidy -diff`; `ci: todos los controles en verde`; `código 0` |
| Prerrequisitos | `go1.27.1 darwin/arm64`, git 2.50.1, GNU Make 3.81, rama `h5-skill-boe-legislacion`, ninguna línea de `git status` fuera del directorio del feature, clon creado sin mensajes, tres sondas en lo esperado |
| 1 · skill y controles | `make skills-check` en `ok` para `internal/app`, `internal/skills` e `internal/evals`; **159** líneas; el enlace da `../../../bin/instalado/kitlegal`; solo `SKILL.md`, `references/normas.md` y `scripts/boe`; la cabecera «generado desde data/normas.yaml, no editar», `1` vez |
| 2 · `skills-sync` idempotente | Las dos regeneraciones en 0 (`ok internal/app`); los dos `git status` del clon, solo `fin del estado`; en el `Makefile`, `skills-sync` solo en la lista de objetivos (línea 54) y en su regla (114-116), sin ningún anuncio de H5 |
| 3 · deriva de lo generado | Ver abajo (tres `código 2`) |
| 4 · defectos de la skill | Ver abajo (cinco `código 2`); en el árbol real, los once subtests de `TestSkillsDelRepositorio` en `PASS` (`skills`, `datos-sin-regenerar`, `referencia-editada`, `describe-cambiado`, `trescientas-lineas`, `normas-nombradas`, `sin-instrucciones-de-evals`, `regenerar-dos-veces`, `frontmatter`, `enlaces`, `region`), 35 líneas `PASS` con los anidados y ninguna `FAIL` |
| 5 · normas y esquema | Ver abajo (dos `código 2`); `TestLeerNormas` (25 subtests, entre ellos `sin-titulo`, `sin-rango`, `sin-materias`, `identificador-repetido-con-titulos-distintos`, `identificador-repetido-por-un-alias`, `con-vertical`, `identificador-con-otra-forma`, `clave-de-fusion`), `TestEsquemaDeNormas` (`compila`, `rangos-grabados`) y `TestRenderizarNormas` (4) en `PASS` |
| 6 · evals y lo grabado | Los doce ficheros de `contracts/evals-y-grabaciones.md` §2, en orden; `06-irpf-rendimientos-del-trabajo.yaml` listado con `reproduce: boe-fiscal`; ver abajo (tres `código 2`); los ocho tests en `PASS` con todos sus subtests, entre ellos `TestManifiestoDeGrabaciones/clave-repetida`, `/prefijo-repetido`, `/prefijo-de-otro-prefijo`, `/repositorio`, `TestPrepararYComprobar/sin-metadatos-de-un-bloque-esperado`, `/sin-indice-de-una-norma`, `TestLeerConjunto/con-mal-formadas`, `/entradas-que-no-son-evals`, `TestEvalsDelRepositorio/conjunto`, `/normas-conocidas`, `TestPrepararDirectorioDeSesion/eval-normal`, `/prueba-de-red`, `/eval-inexistente`, `/eval-mal-formada`, `/con-faltas` |
| 7 · comparación mecánica | Todo en `PASS`, ningún `FAIL`, con cada subtest que nombra la guía; por test: `TestJuzgar` 18, `TestLeerSesion` 15, `TestLeerTrazas` 21, `TestInforme` 13, `TestInterpretarInvocacion` 12, `TestExtraerCitas` 5, `TestLeerTrazasSinFicheros` 2, `TestEscribirInformeSinSusEntradas` 6 (este último incluye `informe-json-no-se-puede-escribir`) |
| 8 · `make install` | Ver abajo (`código 0`, `código 0`, `código 2`); los cuatro guiones de `TestInstalacion` en `PASS` (`instalar`, `instalar-de-nuevo`, `instalar-con-conflicto`, `instalar-sin-gobin`) |
| 9 · protocolo, cita y reglas | Las cinco claves del frontmatter (`name`, `description`, `metadata`, `kitlegal-applets`, `kitlegal-referencias`); los seis elementos del protocolo, cada uno en su paso o regla (líneas 30, 36, 54, 95, 99 y 101, más la `description` en la 10); la cita de ejemplo `[BOE-A-2015-10565, bloque a21]` en la línea 112; la búsqueda de `KITLEGAL_CACHE_DIR`, evals, job, GitHub Actions y nombres de modelo, solo `fin de la búsqueda`. SC-004, leyendo `## Reglas`: la 1 es FR-006 (no concluir que algo no existe), la 2 FR-010 (nunca inventar contenido legal), la 3 FR-011 (trasladar la vigencia y el carácter informativo) y la 4 FR-015 (nunca actuar en nombre de nadie): 4 de 4; la 5 es FR-012 (ningún caso especial para un territorio) |
| 10 · sin Python | Solo `fin de la búsqueda` sobre `skills`, `data`, `evals`, los dos esquemas, los cuatro guiones, el `Makefile` y `evals.yml` |
| 11 · directorios, atributos y pendientes | Ver la salida completa abajo: todo lo esperado |
| Limpieza | La carpeta desaparece (`test ! -e`) y el estado del árbol da solo `fin del estado` |

**Escenarios negativos, sobre el clon desechable** (`git clone --quiet --branch h5-skill-boe-legislacion .
/tmp/kitlegal-quickstart-h5/repo`; cada caso rompe el clon, ejecuta `make -C … skills-check` y lo restaura con
`git checkout`; el árbol de trabajo no cambia y la limpieza borra el clon):

- **3.a** Una fila `| a mano | | | | |` al final de `references/normas.md`: `código 2`;
  `TestSkillsDelRepositorio/skills` falla con `[boe-legislacion: references/normas.md: contenido-distinto]`.
- **3.b** «de 1 de octubre» → «de 2 de octubre» en el título de la LPAC de `data/normas.yaml`: `código 2`; el mismo
  fallo de `references/normas.md: contenido-distinto` y `TestIdentificadoresDeLasNormas` con
  «`BOE-A-2015-10565: el título no coincide con la búsqueda grabada: Ley 39/2015, de 1 de octubre, del Procedimiento
  Administrativo Común de las Administraciones Públicas.`» (SC-008).
- **3.c** «Devuelve el texto vigente de un bloque» → «Devuelve el texto de un bloque» en `internal/app/boe.go` sin
  regenerar: `código 2`; `[boe-legislacion: SKILL.md: contenido-distinto]`.
- **4.a** `SKILL.md` rellenado hasta 300 líneas (`rtk proxy wc -l` da `300`): `código 2`;
  `[boe-legislacion: SKILL.md tiene 300 líneas (máximo 299)]`.
- **4.b** `name: Boe-Legislacion`: `código 2`; `[boe-legislacion: name "Boe-Legislacion" con caracteres que no son
  a-z, 0-9 ni - boe-legislacion: name "Boe-Legislacion" distinto del nombre del directorio]`.
- **4.c** `scripts/boe` borrado: `código 2`; `[boe-legislacion: scripts/boe: enlace-ausente]`.
- **4.d** `scripts/boe -> ../../../bin/kitlegal`: `código 2`; `[boe-legislacion: scripts/boe: enlace-con-otro-destino
  (apunta a ../../../bin/kitlegal)]`.
- **4.e** `scripts/cita` añadido: `código 2`; `[boe-legislacion: scripts/cita: enlace-sobrante]`.
- **5.a** `vertical: fiscal` en la LPAC: `código 2`; `[boe-legislacion: data/normas.yaml: BOE-A-2015-10565: campo no
  declarado: vertical]`, también en `TestSkillsDelRepositorio/normas-nombradas`, `TestNormasDelRepositorio`,
  `TestEvalsDelRepositorio` y `TestIdentificadoresDeLasNormas`.
- **5.b** `BOE-A-2015-10565` → `BOE-A-15-10565`: `código 2`; `[boe-legislacion: data/normas.yaml: BOE-A-15-10565:
  identificador con otra forma]` y los mismos cuatro tests.
- **6.a** `01-lpac-articulo-21.yaml` sin `pregunta`: `código 2`; `TestEvalsDelRepositorio/formato` con
  `[01-lpac-articulo-21.yaml: línea 2: missing property 'pregunta']` y «"11" is not greater than or equal to "12"»;
  `/conjunto` con «hay 9 evals positivas … y el conjunto lleva exactamente 10», «LPAC (normas con esa abreviatura:
  BOE-A-2015-10565)» sin positiva que la cite y «ninguna eval pregunta exactamente «¿qué dice el art. 21 de la Ley
  39/2015?»».
- **6.b** Sin los metadatos grabados de `BOE-A-2015-10565` (fichero de H4): `código 2`; `TestEvalsDelRepositorio/grabado`
  nombra `01-lpac-articulo-21.yaml`, «el comando esperado boe articulo BOE-A-2015-10565 a21», «la cita esperada
  BOE-A-2015-10565 a21 sin su bloque» y «la norma BOE-A-2015-10565 sin metadatos», primero con código 1 al preparar
  («no hay grabación de GET …/metadatos») y después con código 4 con `--offline` («no hay ninguna entrada vigente»)
  (SC-010).
- **6.c** El `titulo_empieza_por` de la primera entrada del manifiesto copiado en la segunda: `código 2`;
  `TestIdentificadoresDeLasNormas` con «manifiesto de grabación: entrada 1 ("Ley 39/2015,") y entrada 2 ("Ley
  39/2015,"): la misma norma repetida, con el mismo prefijo».
- **8** `make -C … install` con `HOME`, `GOBIN`, `GOENV=off` y `GOPROXY=off` temporales: `código 0`, con
  «`instalar-skills: boe-legislacion → /private/tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion`» y
  «`instalar-skills: kitlegal → /tmp/kitlegal-quickstart-h5/gobin/kitlegal`» (el binario compilado con
  `-X main.commit=536359ce…`); los dos `readlink` dan esas rutas; `scripts/boe articulo BOE-A-2015-10565 a21 --describe`
  por el enlace de la skill instalada da `"title": "boe articulo"`; la segunda instalación `código 0` y en
  `~/.claude/skills` temporal solo `boe-legislacion`; con un directorio ajeno `boe-legislacion` en otro `HOME`,
  `código 2` con «`instalar-skills: conflicto: /tmp/kitlegal-quickstart-h5/otro-home/.claude/skills/boe-legislacion ya
  existe y no es un enlace a /private/tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion; no se modifica`», y «la
  entrada en conflicto sigue siendo un directorio» (SC-006).

Los subtests de `TestSkillsDelRepositorio` que copian el árbol (`datos-sin-regenerar`, `describe-cambiado`,
`referencia-editada`, `regenerar-dos-veces`, `enlaces`, `region`, `frontmatter`, `trescientas-lineas`) fallan también
en 3, 4 y 5, porque copian el defecto y esperan un árbol limpio; no contradice lo esperado.

**Escenario 11, la salida tal cual** (FR-083 a FR-085, SC-011):

```text
.agents/skills/golang-how-to/SKILL.md: linguist-vendored: set
.agents/skills/golang-how-to/SKILL.md: linguist-generated: set
skills/boe-legislacion/SKILL.md: linguist-vendored: unspecified
skills/boe-legislacion/SKILL.md: linguist-generated: unspecified
.agents/.gitattributes: linguist-vendored: unspecified
.agents/.gitattributes: linguist-generated: unspecified
fin del diff
0
fin del recuento
151:| `skills/` | **El producto que se distribuye**: las skills de kitlegal, hoy `boe-legislacion` | sí |
152:| `.agents/skills/` | Skills de agente vendorizadas para trabajar en este repositorio —las de Go de `samber/cc-skills-golang`—, registradas con su origen y su huella en el registro de bloqueo `skills-lock.json`. Se versionan tal cual y no se editan; `.agents/.gitattributes` las marca como vendorizadas y generadas, para que no cuenten en las estadísticas de lenguaje del repositorio ni se desplieguen en los diffs de las propuestas de cambio | no |
153:| `.claude/skills/` | Lo que carga Claude Code al trabajar en el repositorio: un enlace a cada skill de `.agents/skills/` más las skills de spec-kit, con las que se prepara cada hito | no |
157:`make install` —la orden de [Construir e instalar](#construir-e-instalar)— enlaza cada skill de `skills/` en
README.md:29:- **`make install`** instala el binario y enlaza la skill en el directorio personal de skills de Claude Code;
README.md:30:  **`make skills-sync`** regenera lo que se deriva de los datos y del binario; **`make skills-check`**, dentro de
README.md:31:  `make ci`, falla si algo diverge; y **`make evals`** mide la skill con Claude Code desde el job de evals.
README.md:135:make install     # instala el binario en el directorio de binarios de Go ($GOBIN, o $HOME/go/bin) y enlaza las skills en el directorio personal de skills de Claude Code (~/.claude/skills)
README.md:139:`-trimpath`. Qué enlaza `make install`, y cómo, está en la sección siguiente.
README.md:157:`make install` —la orden de [Construir e instalar](#construir-e-instalar)— enlaza cada skill de `skills/` en
README.md:167:### Lo generado: `make skills-sync` y `make skills-check`
README.md:174:- **`make skills-sync`** las regenera y escribe en el árbol; dos ejecuciones seguidas no cambian nada. Se ejecuta
README.md:176:- **`make skills-check`** las regenera en memoria y las compara con el árbol sin escribir nada; comprueba además el
README.md:208:Cada fichero se valida contra el esquema `schemas/eval.yaml.json` dentro de `make ci` (`make skills-check`): una
README.md:215:`make evals SKILL=<skill>` ejecuta las evals de una skill con Claude Code: una sesión por eval, con la skill
README.md:216:instalada por `make install`, y un informe que juzga cada sesión sin modelo —si activó la skill, si hizo con éxito
README.md:248:`make install` nunca imprime los valores por defecto del código (`dev`, `none`, `unknown`); si los
README.md:282:| `make test-integration` | Tests con la etiqueta de compilación `integration`, con detector de carreras: los que dependen del entorno (permisos del sistema de ficheros, dos procesos, la instalación de las skills con `make install` en un directorio personal temporal), siempre dentro de directorios temporales | sí |
README.md:285:| `make skills-check` | Comprueba sin red, sin modelo y sin escribir nada el frontmatter y el límite de líneas de cada `SKILL.md`; las derivas de las referencias, de la tabla de comandos y de los enlaces de `scripts/`; la tabla de normas contra su esquema y sus identificadores; y el formato y el conjunto de las evals y lo grabado que necesitan | sí |
README.md:290:| `make skills-sync` | **Regenera** las referencias, la tabla de comandos de `SKILL.md` y los enlaces de `scripts/` de cada skill: escribe en el árbol, y por eso no forma parte de `ci` | no |
README.md:294:| `make evals` | Ejecuta las evals de una skill (`SKILL=<skill>`) en sesiones con modelo de Claude Code; lo lanza el job de evals | no |
README.md:299:[`CONTRIBUTING.md`](CONTRIBUTING.md). `make evals` tampoco está en `ci`, aunque no pide nada a ninguna
CONTRIBUTING.md:101:| Tests con la etiqueta `integration` (dependen del entorno: permisos, dos procesos, la instalación de las skills con `make install` en un directorio personal temporal) | `make test-integration` | sí |
CONTRIBUTING.md:104:| Skills, datos y evals, sin red, sin modelo y sin escribir nada: frontmatter y límite de líneas de cada `SKILL.md`; derivas de las referencias, de la tabla de comandos y de los enlaces de `scripts/`; tabla de normas contra su esquema y sus identificadores; formato y conjunto de evals y lo grabado que necesitan | `make skills-check` | sí |
CONTRIBUTING.md:105:| Regeneración de lo que se deriva de cada skill (referencias, tabla de comandos de `SKILL.md`, enlaces de `scripts/`) | `make skills-sync` | no — escribe en el árbol |
CONTRIBUTING.md:111:| Evals de una skill con Claude Code (`scripts/evals.sh`; Linux con `strace`, como root o con `sudo`) | `make evals` | no — sesiones con modelo y credencial, fuera de `make ci`; las lanza el job de evals |
CONTRIBUTING.md:170:`make test-e2e`, `make test-integration`, `make schema-check` y `make skills-sync` ya no están en esta
CONTRIBUTING.md:174:H4 la tercera compara `schemas/` con lo que emite `--describe` (sección siguiente); y desde H5, `make skills-sync`
CONTRIBUTING.md:202:directorios llamados `skills`, qué hace `make install` y qué se genera está en el [`README.md`](README.md#skills);
CONTRIBUTING.md:207:cambiar su entrada o su salida, se ejecuta `make skills-sync` y lo regenerado va en el mismo cambio:
CONTRIBUTING.md:208:`make skills-check`, dentro de `make ci`, lo regenera en memoria y falla nombrando la skill y el fichero o el enlace
CONTRIBUTING.md:241:`make evals SKILL=<skill>` ejecuta `scripts/evals.sh` y no forma parte de `make ci`: sus sesiones usan un modelo,
CHANGELOG.md:25:única fuente de verdad, lo generado comprobado en `make ci`, la instalación con `make install` y el formato
CHANGELOG.md:41:  versionado. `make build` y `make install` compilan sin cgo, con `-trimpath` e inyectando los datos de
CHANGELOG.md:220:- **Job de evals**: el flujo `evals` (`.github/workflows/evals.yml`) ejecuta `make evals SKILL=boe-legislacion` a
CHANGELOG.md:225:- **`make skills-sync` real**: deja de anunciar que las skills llegan en H5 y regenera, desde `data/*.yaml` y desde
CHANGELOG.md:228:- **`make skills-check`**, dentro de `make ci`: regenera todo eso en memoria y lo compara con el árbol sin escribir
CHANGELOG.md:232:- **`make evals`** (`scripts/evals.sh <skill>`), fuera de `make ci`: una sesión de Claude Code por eval, bajo
CHANGELOG.md:233:  `strace` y con la red cerrada salvo la del modelo, con la skill instalada por `make install` y el binario
CHANGELOG.md:324:- **`make install` enlaza las skills.** Tras el `go install` de H0, sin cambios, `scripts/instalar-skills.sh` enlaza
README.md:2
CONTRIBUTING.md:1
CHANGELOG.md:1
fin del formato
README.md:3
CONTRIBUTING.md:2
CHANGELOG.md:2
fin del job
1
fin de diez
0
fin de nueve
1
fin de skills-check en README
1
fin de skills-sync en README
1
fin de evals en README
1
fin de test-integration en README
1
fin de install en README
0
fin de make help en README
1
fin de skills-check en CONTRIBUTING
1
fin de skills-sync en CONTRIBUTING
1
fin de evals en CONTRIBUTING
1
fin de integration en CONTRIBUTING
0
fin de la tabla posterior
1
fin del párrafo de CONTRIBUTING
1
fin de la introducción del CHANGELOG
```

Es decir: los dos atributos `set` solo bajo `.agents/skills/` y `unspecified` en la skill del producto y en el propio
`.gitattributes`; ningún cambio en `.claude/skills`, `skills-lock.json` ni `.agents/skills` frente a `main`; `0`
entradas «En H5» en `docs/PENDIENTES.md`; los tres directorios explicados en `README.md` (líneas 151-153); las cuatro
órdenes en los tres documentos; «formato común de eval» 2/1/1 y «job de evals» 3/2/2; «diez controles» `1` y «nueve
controles» `0`; las cuatro filas de `README.md` y el comentario de `make install`, `1`, y la frase de `make help` con
`skills-sync`, `0`; las cuatro filas de `CONTRIBUTING.md`, `1`, su tabla posterior `0` y su párrafo `1`; la entrada de
H5 en la introducción del `CHANGELOG.md`, `1`. Que GitHub excluya `.agents/skills/` de las estadísticas y lo pliegue en
los diffs es el supuesto S8 (*Pendientes*).

**Cobertura**, remedida por T035 el 2026-09-15 (hacia las 06:35, hora de Madrid) con `go clean -testcache` y
`make ci` (código 0, `ci: todos los controles en verde`) sobre el árbol del commit `feat(H5): T035`, cuya base es
`234c6f0` (`feat(H5): T034`), con el `coverage.out` y el
`coverage-integration.out` que deja ese `make ci` (`go tool cover -func` para el total; para cada árbol, la suma de
sentencias del perfil, que es lo que Codecov mide, contando cada bloque una vez):

| Umbral | Exigido | Perfil unitario (`make test`) | Unión de los dos perfiles (lo que Codecov une) |
|---|---|---|---|
| Global (`codecov/project`) | ≥ 70 % | **96,8 %** (`-func`; 5677/5866 sentencias, 96,8 %) | **97,3 %** (`-func` del perfil de integración; 5706/5866, 97,3 %) |
| `internal/core/**` (componente `internal_core`) | ≥ 85 % | **90,1 %** (73/81) | **90,1 %** |
| `internal/cli/**` (componente `internal_cli`) | ≥ 90 % | **98,6 %** (348/353) | **98,6 %** |
| `internal/skills` (nuevo; T035) | — | 98,5 % (1169/1187) | 98,5 % |
| `internal/evals` (nuevo; T034) | — | 99,0 % (1428/1443) | 99,0 % |
| `internal/app` (con sus ejemplos de e2e, 0/32 sentencias cubiertas) | — | 87,8 % (445/507; 93,7 % el paquete solo) | 87,8 % |
| `internal/source/boe` | — | 99,4 % (875/880) | 99,4 % |
| `internal/cache` | — | 90,7 % (478/527) | 96,2 % (507/527) con `-tags=integration` |
| `internal/httpx` (T032 solo cambia tests) | — | 97,2 % (792/815) | igual |
| `internal/render` | — | 95,8 % (69/72) | igual |

`internal/core` no gana ni pierde sentencias: sus 81 siguen siendo las de `internal/core/schema` (`error.go`,
`huella.go`, `sobre.go`), y `internal/app` conserva las 507 de H4: H5 no toca el dominio ni el applet. Frente a la
medida de T029 sobre `536359c` cambian dos árboles: `internal/evals`, con T034, de 94,8 % (1365/1440) a 99,0 %
(1428/1443), con tres sentencias más (el rango del código final de la traza y la lectura del esquema de eval por su
ruta); e `internal/skills`, con T035, de 95,8 % (1134/1184) a 98,5 % (1169/1187), con tres sentencias más en total (la
lectura del esquema de normas por su ruta, el identificador de una norma como texto de su clave, el cierre de
`properties` leído como un token más y un único directorio creado por arreglo). Los demás árboles dan las mismas
cifras. En los seis ficheros de cada paquete que `codecov/patch` midió en rojo quedan 16 bloques sin cubrir en
`internal/evals`, de una línea cada uno (antes 82 líneas), y 18 en `internal/skills`, de una sentencia cada uno (antes
53 líneas), cada uno con su motivo en `gates/tarea-T034.md` y en `gates/tarea-T035.md`. **Ningún umbral se rebaja**:
`codecov.yml` no aparece en el diff frente a `main`. `codecov/patch` (`target: auto`) lo leyó T030 en la plataforma
(intentos 2 y 3, 2026-09-15, cabezas `417635e` y `857ec46`, con las mismas cifras porque T036 y T037 no tocan ningún
fichero Go): **98,42 % del diff frente al objetivo de 94,70 %, en verde**, con 33 líneas sin cubrir, las justificadas
en `gates/tarea-T034.md` y `gates/tarea-T035.md`; `codecov/project` 96,30 % (`gates/evidencia-plataforma.md`).

**Sin ninguna supresión nueva**: `0` líneas `//nolint` y `0` `t.Skip` añadidas en ficheros `.go` frente a `main`.
`gosec` se resuelve sin supresiones (`filepath.Clean`, `0o600`, lectura y escritura por auxiliares distintos, programas
y argumentos de `exec` como constantes).

**Sin red y sin tocar lo protegido**: ninguna tarea ejecutó `KITLEGAL_RECORD`, `scripts/grabar-evals.sh`,
`make evals`, `scripts/evals.sh` ni `make verify-sources`; el manifiesto, las grabaciones de H5, los dos esquemas, los
guiones `testscript` y las 195 sesiones sintéticas los fijó una persona en las pausas `[datos]` (T001, T008, T009, T018,
T020, T021, T022), como exige la obligación 2 del plan; las grabaciones de H4 no cambian.

**Aviso de método.** El envoltorio de terminal de esta máquina reescribe la salida de `go test`, `git status
--porcelain` y `git diff`. Todo lo de arriba está medido con el paso directo (`rtk proxy`), que es la forma en que el
quickstart escribe cada orden. Cuando la salida de un escenario negativo superó el tamaño que la herramienta muestra,
el veredicto (`código`, mensajes de fallo y líneas `PASS`) se leyó con `rtk proxy grep` sobre la copia que la
herramienta guarda; nada se redirigió a ficheros.

## Decisiones

Las de diseño están en research.md (D1-D23) con alternativas y motivo; aquí, las que un revisor necesita para leer
esta propuesta, y las tomadas durante la implementación.

- **Herramientas como tests, no como binarios** (D2, D7): `internal/skills` e `internal/evals` no los enlaza el
  binario; `make skills-check` y `scripts/skills-sync.sh` ejecutan `go test` (`-args -regenerar-skills` para
  escribir), como `schema-check`, `verify-sources` y `grabar-fixtures`. La tabla de comandos sale de `describir` desde
  `internal/app/skills_test.go`, sin exportar nada del kernel. Alternativas rechazadas: `package main` de herramienta
  (excepciones de `forbidigo` y contra `golang-project-layout`); exportar `app.Describir` (cambio del kernel que el hito
  no necesita).
- **La skill declara lo que se genera** (D4): `metadata.kitlegal-applets` y `metadata.kitlegal-referencias`. Deducirlo
  de `scripts/` no detectaría un enlace borrado (FR-036, FR-042).
- **`scripts/boe -> ../../../bin/instalado/kitlegal`** (D5): `make install` crea `bin/instalado/kitlegal` como enlace a
  lo que `go list -f '{{.Target}}'` da, para que la skill ejecute el binario **instalado** y no la última construcción
  (`bin/kitlegal`).
- **Normas indexadas por identificador, con el lector común que rechaza claves repetidas** (D8): `data/normas.yaml`, las
  evals y el frontmatter pasan por `esquemas.go` sobre `yaml.Node`; un alias o una clave de fusión que repitan una clave
  también fallan (`TestLeerNormas/clave-de-fusion`, `/identificador-repetido-por-un-alias`). El `enum` de `rango` es el
  conjunto exacto de `rango.texto` de las doce búsquedas grabadas (nueve valores; `gates/tarea-T009.md`).
- **Verificación offline de identificadores con el propio applet** (D9): `TestIdentificadoresDeLasNormas` reproduce en
  proceso cada búsqueda del manifiesto sobre `httpx.Replay` y exige identificador y título; también falla si una norma no
  tiene entrada en el manifiesto o sale en la búsqueda de otra entrada.
- **Grabación en la pausa de un manifiesto bajo `testdata/` de la raíz** (D11): es la única ubicación fuera de la fuente
  en la que un fichero nuevo pausa al workflow; el arnés siembra desde H4 y solo graba lo que falta. La primera grabación
  se cortó con el buscador del BOE caído; la persona la completó antes de reanudar (`1809a56`: el arnés siembra desde
  H4 con la grabación activa; `1f4feea`: 32 respuestas nuevas). Ningún título, identificador ni bloque de `data/normas.yaml`
  o de las evals se escribió de memoria: cada valor se copió de la respuesta grabada (`gates/tarea-T011.md`,
  `gates/tarea-T012.md`); la constante sintética con el título de la LRBRL se corrigió al título grabado («Reguladora»
  con mayúscula y sin punto final).
- **Evals comparadas mecánicamente** (D10, D12): activación, respuesta, modelo y versión del transcript `stream-json`;
  código de la sesión de un fichero que el guion escribe siempre; invocaciones y conexiones de una traza `strace -ff`
  de la sesión entera, con atribución por hilos. Una eval cuya sesión no terminó no pasa; una invocación fuera de lo
  grabado no hace fallar la eval por sí sola (FR-076), pero una petición llegada a la red de una fuente hace fallar el
  veredicto global (SC-012).
- **Garantía de red del job** (D16): proxy del entorno hacia `127.0.0.1:9`, cerrado y comprobado, más la traza como
  observación independiente; y **sin Python** (D17): búsqueda como root en todo el sistema de ficheros salvo `/proc` y
  `/sys`, retirada de lo encontrado y la misma búsqueda antes de la primera sesión, registrada en `sin-python.txt` y
  copiada byte a byte al informe, sin consultar ni purgar paquetes (T033, más abajo).
- **Prueba de red por etiqueta o entrada `prueba_de_red`**, no como eval: añadir una eval rompería las diez positivas de
  FR-062. **La ejecución se identifica por el último evento `labeled` y `workflowName`**, nunca como «la última de la
  lista» (quickstart §12; supuesto S12).
- **`.agents/.gitattributes` anidado** (D18): el guardián del workflow no admite `.gitattributes` en la raíz y git
  aplica igual un fichero anidado (V33, V35); cambiar el extractor es proceso, fuera de la rama del hito.
- **Etiqueta `evals` en `run.build-tags` y cinco palabras en `misspell.ignore-rules`** (D20): cada palabra responde a
  una que el código escribe suelta por contrato (`comandos` es la clave del formato; `legislativo`, parte de «Real
  Decreto Legislativo» que una expresión casa tal cual…); las otras veintiuna de V42 no entran y el código no las
  escribe sueltas.
- **El esquema de eval es exacto al contrato** (`0c24e15`, T001): se retiró un `maxLength` de `reproduce` que el
  contrato no fija.
- **Cada orden del quickstart se ejecuta tal cual en la sesión desatendida** (T029): el intento 1 se detuvo en la
  segunda orden del escenario 4 porque `wc` sobre un fichero de la carpeta temporal está bloqueado por Claude Code
  fuera del directorio de trabajo, también como orden sola; la guía pasa a `rtk proxy wc -l`, la tabla de formas gana la
  fila y una sonda de los prerrequisitos comprueba la forma en cada ejecución. No se sustituyó la orden por otra a
  criterio del ejecutor, como manda la tarea (`gates/tarea-T029.md`).
- **Un test intermitente no se reintenta hasta el verde: se arregla de raíz en una tarea nueva antes del cierre**
  (T032, procedimiento de H3/T018 y H4/T037): la verificación del intento 2 de T029 cayó una vez en
  `internal/httpx` › `TestReintentosDosErroresYUnAcierto` (85,39 ms frente a 90 ms), un fichero de H2 que H5 no había
  tocado. La causa es de la medida, no del ritmo: el limitador (`x/time/rate`, cubo de un token) fija los **turnos**
  anclados al primero, y la llegada al servidor añade a cada turno un despacho que no es igual en todas (un despertar
  tardío bajo `-race` con todos los paquetes en paralelo, la conexión que la primera abre y las demás reutilizan), así
  que dos llegadas consecutivas pueden acercarse por debajo de «intervalo menos holgura» sin que el ritmo haya fallado.
  T032 cambia las dos aserciones (reintentos y ritmo) a la cota que el limitador sí garantiza —cada llegada, medida
  desde antes de la primera petición, no se adelanta a su turno— y retira las dos holguras; el decorador, el cubo y la
  cadena no cambian, y los mutantes de `gates/tarea-T032.md` demuestran que la cota nueva distingue una implementación
  correcta con despacho tardío de las cuatro incorrectas. Como las rutas congeladas de T029 no admitían el arreglo y su
  propia línea prohíbe tocar ficheros fuera del directorio del feature, T029 volvió a `[ ]` y este intento 3 repitió la
  guía entera sobre el commit de T032.
- **La retirada de Python no consulta ni purga paquetes** (T033; research D17 y V60): la prueba de red del intento 1 de
  T030 (ejecución 34922606273, `gates/prueba-de-red.md`) se detuvo en «Retirar Python del runner» con código 100. En
  `ubuntu-24.04` el filtro del paso elegía 110 paquetes de Python, entre ellos `python3`, `python3-minimal`,
  `python3-apt`, `python3-debconf` y `python3-netplan`, de los que dependen paquetes del sistema, y `apt-get purge -y
  --auto-remove` terminaba con «pkgProblemResolver::Resolve generated breaks» (`shim-signed`, `grub-efi-amd64-signed`,
  `grub2-common`) antes de buscar nada; V57 había comprobado la purga en un contenedor en el que ningún paquete del
  sistema dependía de Python. El paso ya no usa `dpkg-query` ni `apt-get`: la garantía de FR-073 y FR-081 es la que ya
  daba la búsqueda —como root en todo el sistema de ficheros, retirada de cada ruta con la regla del prefijo y de lo
  usado, y la búsqueda final—, sin cambiar la orden de búsqueda, esa regla, lo usado, las marcas, `set -euo pipefail` ni
  la comprobación 3 de `scripts/evals.sh`. Comprobado en contenedores desechables de `ubuntu:24.04` sin red, con el paso
  extraído del contrato: con un paquete falso esencial que depende de dos de Python, el paso anterior termina con 100 y
  el mismo mensaje sin buscar, y el nuevo con 0, retira el intérprete y su biblioteca y deja la comprobación 3 en
  `resultado: ninguno`; los casos de V56 dan con el paso nuevo lo que anota V60. Los paquetes siguen registrados en dpkg
  y pierden lo que la búsqueda encuentra, también sus ficheros del registro con esos nombres; ningún paso posterior usa
  dpkg ni apt. Alternativas rechazadas: forzar `apt` (se llevaría lo que depende de Python, hasta la cadena de arranque);
  `dpkg --purge --force-depends` (el `prerm` de `python3` ejecuta `py3clean`, que necesita el intérprete: retirado este
  antes, falla con 127 y deja `python3` a medias, así que el resultado depende del orden); y retirar los ficheros de las
  listas de dpkg (no retira ningún intérprete que la búsqueda no encuentre y añade como supuestos el formato de esas
  listas y sus desvíos).
- **Ramas del diff sin cubrir: tests con su condición real, dos reestructuraciones y lo demás justificado** (T034,
  `gates/tarea-T034.md`): `codecov/patch` salió en rojo en el intento 1 de T030. En `internal/evals`, cada rama
  alcanzable tiene un test con entradas literales en `t.TempDir()` y fallos de disco por la estructura del directorio,
  no por permisos. Dos errores que solo podía dar una constante se hicieron comprobables: el código de
  `+++ exited with N +++` se toma con todas sus cifras y se exige de 0 a 255 (antes `[0-9]{1,3}` hacía imposible el
  error de conversión y aceptaba de 256 a 999; data-model §9, regla 5), y el esquema de eval se compila desde una ruta
  (`compilarEsquemaDeEval`). Quedan sin cubrir, cada uno con su motivo, los errores del registro de applets, de las
  banderas globales, del esquema publicado y de la codificación del informe (constantes que vigilan sus propios tests),
  y seis lecturas, escrituras o retiradas de lo que la propia función acaba de listar o crear (solo fallan por permisos
  o carreras). Alternativa rechazada: pasar el registro y la gramática como parámetros para que un test inyecte uno que
  falla, que fuerza la rama con un doble y deja las propagaciones igual.
- **La sesión sin el aislamiento de subprocesos de Claude Code** (T036; research V12, V61 y D13; plan §VII): la prueba
  de red del intento 2 de T030 (ejecución 34930222593) mostró que, con `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1`, Claude Code
  2.1.270 de Linux no arranca sin `bubblewrap`, y el binario fuerza además con esa variable el modo de permisos
  `default`. T036 comprobó en contenedores de `ubuntu:24.04`, con un token inválido y un servidor local en lugar de la
  API —sin modelo y sin red salvo para instalar paquetes—, lo que haría falta para mantenerla: con `bubblewrap` y
  `--allowedTools Skill Bash` las sesiones arrancan, pero la primera orden de Bash exige `socat` y, con él, corre dentro
  de `bwrap` con un espacio de nombres de PID propio, cuyos `clone` devuelven números que no son los de los ficheros de
  la traza (`LeerTrazas`: «dos líneas crean t.50»), y con el disco de solo lectura salvo `/home`, `/tmp` y otros cinco
  directorios. Es la incompatibilidad con la traza que preveía la línea de T036, así que la sesión fija
  `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0` y conserva `--permission-mode bypassPermissions`, sin instalar nada más en el
  runner. Claude Code sigue sin pasar el token al entorno de las órdenes y de los ganchos, y se asume que una orden del
  mismo usuario podría leerlo en `/proc`, del entorno de `claude` (runner desechable, `contents: read`, disparo solo
  desde el repositorio, token revocable). Alternativas rechazadas: `bwrap` con la traza adaptada (`strace
  --decode-pids=pidns`), que cambia las formas de data-model §9 antes del último intento, ejecuta la skill con el disco
  de solo lectura (lo que D16 rechaza del sandbox) y depende de espacios de nombres de usuario que AppArmor puede
  restringir en Ubuntu 24.04; `processWrapper` y el descriptor del token, sin documentar; el sandbox de Bash, por D16.
  La misma comprobación encontró que `strace -s 4096` corta el argv de la instantánea de shell que Claude Code crea
  antes de la primera orden de Bash, lo que dejaba ilegible toda sesión que ejecute Bash con cualquier valor de la
  variable: lo arregla T037 (siguiente punto), antes de T030.
- **Trazas enteras: `strace -s 131072`** (T037; research V61 y V62; contrato del job §3.2; data-model §9, regla 5): la
  orden de la sesión pasa de `-s 4096` a `-s 131072`, el tamaño máximo de un argumento de `execve` en Linux con páginas
  de 4 KiB (`MAX_ARG_STRLEN`, con el nulo final), y la lectura de la traza no cambia: un argumento cortado sigue
  haciéndola ilegible, porque el argv de una invocación cortada no se puede comparar. Comprobado en contenedores de
  `ubuntu:24.04` con `strace` 6.8 y `--network none`: un argumento de 131 071 octetos sale entero y `LeerTrazas` lo lee
  (con `-s 4096`, cortado e ilegible), y uno de 131 072 da `E2BIG`; con Claude Code 2.1.270 en las condiciones del caso
  4 de T036 (V61), el bloque de la sesión de `scripts/evals.sh` deja la instantánea de shell entera (6 987 octetos) y
  una traza legible con las dos invocaciones de `a9998` (código 5 con tres `connect` locales y código 4 sin
  conexiones), cuando el bloque anterior la dejaba ilegible en esa línea. Queda un límite: `strace` corta con `...]` una
  lista de más de 131 072 argumentos, que el núcleo admite, y esa sesión saldría ilegible, sin ocultar la invocación
  (S4). Alternativas rechazadas: admitir en `LeerTrazas` cadenas cortadas, que compararía argv incompletos con los
  comandos esperados; y `-v`, que tampoco corta la lista pero escribe el entorno de cada `execve` con sus valores, y con
  él `CLAUDE_CODE_OAUTH_TOKEN` desde la `execve` de `claude`.
- **`LeerTrazas` admite el relleno de alineación de `strace`** (T038 y T039; research V53, V63 y S4; data-model §9,
  regla 5; contrato del job §4): la prueba de red del intento 3 de T030 (ejecución 34936425178) dejó ilegibles las
  trece sesiones en la línea `vfork()` + 33 espacios + `= <pid>` con la que Claude Code 2.1.270 de x86_64 crea los
  procesos de sus órdenes: tras el paréntesis de cierre, `strace` escribe un espacio y, si la llamada no llega a la
  columna 40 (`-a 40`, su valor por defecto), el relleno hasta ella, y `formaDeLlamada` exigía un solo espacio. Ahora
  admite ahí uno o más espacios y nada más: el resultado sigue siendo obligatorio, `? ERRNO` solo vale en una sesión
  cortada y cualquier otra línea hace ilegible la traza. Lo fijan el caso `proceso-por-vfork` de `TestLeerTrazas`, con
  la línea real en las trazas sintéticas (T038), y `TestLeerLlamadaConRelleno` sobre literales: la línea del runner
  (47 octetos, con el `=` en la columna 41) y la misma con un solo espacio se leen como una `vfork` con resultado 11494
  que crea ese proceso, y sin `=`, sin resultado, sin espacio o con un tabulador la línea sigue siendo ilegible; los
  dos, en rojo con el lector anterior (el caso, con el motivo del runner) y en verde con el nuevo. Sobre una copia
  desechable del árbol, cada variante de la expresión cae en un subtest: la de un solo espacio, en `proceso-por-vfork`
  y `relleno-del-runner`; la de cero o más espacios, en `sin-espacio`; y la de cualquier blanco, en `tabulador`. Comprobado en un
  contenedor de `ubuntu:24.04` con `strace` 6.8 y sin red (V63): `getppid()`, añadida solo al filtro de la sonda, sale
  con 31 espacios y el `=` en la columna 41, como la línea del runner, y `LeerTrazas` rechaza esa traza porque
  `getppid` no es del filtro; con el filtro de la sesión, `clone(child_stack=NULL, flags=SIGCHLD)` (38 columnas) sale
  con dos espacios, y `LeerTrazas` lee la traza con la invocación del proceso que crea, código 5 y tres `connect`
  locales. Alternativas rechazadas: `strace -a 1` en la orden de la sesión, que quitaría el relleno pero cambia el
  contrato del job §3.2, obliga a repetir en contenedores V61 y V62 y deja el lector dependiendo de una opción para leer
  el formato por defecto de `strace`, el que reproducen las trazas sintéticas, sin añadir ninguna garantía, porque el
  relleno no pierde información; admitir cualquier blanco (`\s+`), una forma que `strace` no escribe; e ignorar las
  líneas que no casan, que dejaría hilos sin atribuir y `red` vacío en falso (FR-076).
- **Protocolo de lectura de bloques y forma de la cita frente a lo grabado** (T040; research D23 y V64; contrato de la
  skill §2.3, §2.4 y §3): la prueba de red del intento 4 de T030 (ejecución 34941499481) leyó las trece trazas y dio
  `fallo` por seis positivas, sin ningún defecto del job. Tres causas eran del protocolo: `scripts/boe articulos` falla
  entero en cuanto falla un bloque, y en 03 y 08 el modelo pidió el bloque esperado con un vecino no grabado y, con la
  regla 2, se rindió sin pedirlo por separado (la 03 ejecutó tal cual el ejemplo de `articulos` del paso 3); en 02
  compuso `a118` del número del artículo, y en la Ley 9/2017 el artículo 118 es `a1-30`; y en 10 citó con el nombre de
  la norma dentro de los corchetes. `SKILL.md` conserva sus cinco pasos y sus cinco reglas y ahora dice que el id se
  copia de la entrada del índice y nunca se compone del número; que los bloques se leen de uno en uno con `articulo`, y
  `articulos` solo cuando hacen falta varios a la vez y todos salen del índice (sin ejemplo con ids seguidos); que,
  ante un 4 o un 5 de una orden con varios bloques, se pide cada uno por separado antes de decir qué no se pudo
  consultar (también en la regla 2); y que dentro de los corchetes no va nada más que el identificador y el id, con
  `[Ley 39/2015, BOE-A-2015-10565, bloque a21]` como forma que no vale. No nombra evals, job ni modelos, la región
  generada no cambia y la skill queda en 171 líneas. No hay comprobación local con un modelo (FR-044): la evidencia es
  el intento siguiente de T030 y la ejecución de cierre de T031. Alternativas rechazadas: grabar los vecinos que pidió
  el modelo (persigue cada sesión y deja lo grabado sin regla; FR-074); admitir texto delante del identificador en la
  extracción de citas (cambia la forma fija de FR-008 para tolerar un desvío); contar un bloque leído con 4 o 5
  (FR-072, FR-076); cambiar el modelo de las sesiones (clarificación del spec, D13); y que `articulos` devuelva los
  bloques que resuelve (contrato del verbo de H4). Las evals 05 y 07, que fallan por el artículo elegido y no por el
  protocolo, siguen pendientes de la decisión de la persona (*Pendientes*).
- **Ningún ADR nuevo**: el plan no se aparta de ninguna decisión existente (skills sin código, `data/` como fuente de
  verdad, `scripts/` como symlinks al binario, ADR 0012).

## Pendientes

- **Supuesto S8 (research D22), pendiente de la fusión**: que GitHub excluya de las estadísticas de lenguaje
  los ficheros con `linguist-vendored` y pliegue en los diffs los que llevan `linguist-generated`, también desde un
  `.gitattributes` anidado. Lo comprobable sin plataforma está arriba (escenario 11: `git check-attr` da `set` para los
  dos atributos solo bajo `.agents/skills/`). **Lo comprobable con la propuesta de cambio abierta** (T030, intentos 1 a 4,
  2026-09-15, #27, cabezas `6d68c21`, `417635e`, `857ec46` y `537e5d6`): la API de ficheros de la propuesta
  (`gh api --paginate 'repos/jmorenobl/kitlegal/pulls/27/files?per_page=100' --jq '.[].filename'`) lista 360, 366,
  366 y 368 ficheros respectivamente y, bajo `.agents/`, solo `.agents/.gitattributes` en los cuatro casos; ningún fichero de `.agents/skills/` está en el diff, así que el
  plegado no se puede observar aquí (y `gh pr diff 27 --name-only` tampoco sirve: la plataforma lo rechaza con
  `HTTP 406 … the diff exceeded the maximum number of files (300)`); se verá en la primera propuesta que toque
  `.agents/skills/`. Las estadísticas de la rama principal antes de fusionar (`gh api repos/jmorenobl/kitlegal/languages`)
  son `{"Go":1579928,"Shell":138689,"Python":65011,"PowerShell":35337,"Makefile":9505}` (la misma lectura en los cuatro
  intentos), y los ficheros versionados de
  `.agents/skills/` suman 13 666 bytes de `.go` y 0 de `.py`, `.ps1` y `.sh` (`git ls-files -z ".agents/skills/*.<ext>"`
  con `xargs -0 cat` y `wc -c`): tras la fusión, S8 se cumple si la cifra de `Go` deja de contar esos bytes, lo que se
  compara con la suma de los `.go` versionados con y sin `.agents/skills/` en el commit de fusión.
- **Plataforma, intento 1 de T030 (2026-09-15)**: `ci` en verde (ejecución 34922178932, 4 m 35 s) y `codecov/project`
  (94,22 %), `internal/core` (90,36 %) e `internal/cli` (98,09 %) en verde; **`codecov/patch` en rojo**, 93,54 % del diff
  frente al objetivo `auto` de 94,70 % (`gates/evidencia-plataforma.md`), que arreglan T034 y T035 con tests de las
  ramas sin cubrir, sin tocar ningún umbral (hechas: en sus seis ficheros quedan 16 bloques de una línea en
  `internal/evals` y 18 de una sentencia en `internal/skills`, justificados en `gates/tarea-T034.md` y
  `gates/tarea-T035.md`); y la **prueba de red** (ejecución 34922606273) se detuvo en «Retirar Python del
  runner» con código 100: `apt-get purge` no puede retirar los 110 paquetes de Python de `ubuntu-24.04` porque de ellos
  dependen paquetes del sistema, que arregló T033. Los detalles de ese intento están en las versiones de
  `gates/evidencia-plataforma.md`, `gates/prueba-de-red.md` y `gates/tarea-T030.md` del commit `c4c7613`.
- **Plataforma, intento 2 de T030 (2026-09-15, cabeza `417635e`)**: `ci` en verde (ejecución 34929854868, 4 m 49 s) y
  los cuatro estados de Codecov en verde, `codecov/patch` incluido (98,42 % frente a 94,70 %; `codecov/project`
  96,30 %, `internal/core` 90,36 %, `internal/cli` 98,09 %; `gates/evidencia-plataforma.md`). La **prueba de red**
  (ejecución 34930222593) pasa el paso «Retirar Python del runner» (921 rutas, 1 m 47 s, `búsqueda tras retirar:
  ninguno`) y llega al informe, pero las trece sesiones terminan con código 1 al arrancar Claude Code: con
  `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1`, el binario 2.1.270 de Linux exige `bubblewrap`, que la imagen no trae, y
  además fuerza el modo de permisos a `default` (`gates/prueba-de-red.md`, `gates/tarea-T030.md`). T036 lo comprobó en
  contenedores y lo resolvió con `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0`, porque el aislamiento con `bwrap` deja la traza
  sin atribuir (*Decisiones*); la comprobación encontró además el corte del argv de la instantánea de shell, que
  arregló T037 (`strace -s 131072`; *Decisiones*). T030 siguió sin marcar; su intento 3, tras T036 y T037, repitió la
  prueba de red.
- **Plataforma, intento 3 de T030 (2026-09-15, cabeza `857ec46`)**: `ci` en verde (ejecución 34935998623, 4 m 30 s) y
  los cuatro estados de Codecov en verde con las mismas cifras del intento 2 (T036 y T037 no tocan ningún fichero Go;
  `gates/evidencia-plataforma.md`). La **prueba de red** (ejecución 34936425178) pasa el paso «Retirar Python del
  runner» (921 rutas, 3 m 7 s, `búsqueda tras retirar: ninguno`) y llega al informe, y esta vez **las trece sesiones
  arrancan, se autentican, terminan con código 0 entre 7 y 40 s y responden** (`Modelos de las sesiones:
  claude-haiku-4-5-20251001`, `Versiones de Claude Code: 2.1.270`, salida de error vacía en las trece, las diez
  positivas con la skill activada y las dos de no activación sin ella): los dos motivos del intento 2 están resueltos.
  Pero las trece salen con `sesión ilegible` por la misma línea del fichero del hilo principal de `claude`: en x86_64,
  Claude Code crea los procesos de sus órdenes con `vfork`, y `strace` escribe esa llamada, más corta que su columna
  de alineación, como `vfork()` + 33 espacios + `= <pid>`, una forma que `LeerTrazas` no admite porque exige un solo
  espacio antes de `= ` y ninguna sonda de research (todas en arm64, donde `vfork` no existe como llamada) vio el
  relleno (`gates/prueba-de-red.md` §4, `gates/tarea-T030.md`). Ninguna invocación se leyó, así que el informe no
  dice qué consultaron las sesiones 05, 06 y 10, cuyas respuestas dicen que la fuente devolvió código 5 (una consulta
  fuera de lo grabado): lo dirá el intento siguiente. Arreglado en **T038** (de datos: el caso `proceso-por-vfork` de
  las trazas sintéticas con la línea real) y **T039** (`LeerTrazas` admite el relleno; data-model §9, contrato §4,
  research V53, V63 y S4; *Decisiones*), antes de T030: la forma está comprobada con la línea del runner en los tests
  y con `strace` 6.8 en contenedores (V63), y que las trece trazas reales se lean con ella lo dirá el intento siguiente
  de T030. T030 sigue sin marcar, con sus tres intentos agotados.
- **Plataforma, intento 4 de T030 (2026-09-15, cabeza `537e5d6`, el tercero que cuenta el workflow tras la ampliación
  de la supervisión)**: `ci` en verde (ejecución 34941036417, 4 m 44 s) y los cuatro estados de Codecov en verde con las
  mismas cifras de los intentos 2 y 3 (T039 no añade ninguna línea sin cubrir; `gates/evidencia-plataforma.md`). La
  **prueba de red** (ejecución 34941499481) pasa el paso «Retirar Python del runner» (las mismas 921 rutas, 2 m 55 s,
  `búsqueda tras retirar: ninguno`), llega al informe y, con T038 y T039, **las trece trazas se leen enteras**: la prueba
  de red cumple todo lo que SC-012 espera (las dos filas de `a9998` con 5 y 4, la conexión `127.0.0.1:9` de clase
  `local` atribuida a la invocación sin `--offline`, ninguna sesión ilegible ni cortada, ninguna petición a la red de una
  fuente, la sesión de prueba de red igual que la 01), pero el veredicto es `fallo` porque **seis positivas no pasan**
  (02, 03, 05, 07 y 08 por `comando ausente` y `cita ausente`; 10 solo por `cita ausente`), todas por lo que el modelo
  hizo frente a lo grabado y ninguna por el job: en 03 y 08 pidió con `articulos` el bloque esperado junto a un vecino
  no grabado, el verbo (todo o nada) terminó con 5 y el modelo se rindió sin pedir el bloque por separado; en 02 compuso
  `a118` del número del artículo, y en la Ley 9/2017 el artículo 118 es `a1-30`; en 05 y 07 eligió otro artículo (`a2`,
  `a140`), porque el índice de la fuente solo da «Artículo N» sin rúbrica y encontrar un artículo por su materia depende
  de lo que el modelo sabe; y en 10 citó con el nombre de la norma dentro de los corchetes (`gates/prueba-de-red.md` §3.3
  y §4, `gates/tarea-T030.md`). Arreglado en **T040** (el protocolo de la skill: el id se copia de la entrada del índice
  y nunca se compone del número; los bloques de uno en uno, y cada uno por separado si una orden con varios termina con
  4 o 5; nada más que identificador e id dentro de los corchetes; *Decisiones*, research D23 y V64), antes de T030: las
  medidas de la caché preparada están comprobadas en local (V64), y que el protocolo reforzado dé las positivas 02, 03,
  08 y 10 lo dirán el intento siguiente de T030 y la ejecución de cierre de T031, porque ningún control local ejecuta un
  modelo (FR-044). Para 05 y 07 no hay arreglo en el protocolo, y la decisión queda para la persona (abajo, *Lo humano
  que queda*). T030 sigue sin marcar.
- **Supuestos de plataforma (research D22), pendientes de T030 (prueba de red, `gates/prueba-de-red.md`)**. Lo que el
  intento 4 dejó (`gates/prueba-de-red.md` §5): S12 se cumple en (1), (2), (3), (5) y (6), con (4) sin ejercer; S2 y
  S7 se cumplen en todo lo ejercido (el job no instala `bubblewrap` ni `socat`); S9 se cumple en las trece sesiones
  (entre 6 y 38 s, ninguna cortada); S10 se cumple en el 0 y en su propagación; S1, S5, S6 y S11 se cumplen en lo
  ejercido (el job de la rama corre con el secreto, las sesiones se autentican y aceptan el modelo, la skill se activa
  donde debe, Bash ejecuta el binario y el modelo llega a la API con el proxy que rechaza para todo lo demás);
  **S4 se cumple**: las trece trazas del runner se leen enteras (la `vfork()` con relleno de V63, las `execve` enteras
  de V62, los hilos de Go y de Claude Code), la conexión `local` de `a9998` se atribuye a su invocación y las 26
  invocaciones del binario tienen código y conexiones coherentes. Los supuestos: S1
  (`pull_request` con `types: [labeled]` ejecuta el fichero del job de la rama con los secretos del repositorio), S2 (lo
  que trae `ubuntu-24.04`: `sudo -n`, `strace`, `node`, `npm`, `timeout`, findutils y coreutils; la línea
  `búsqueda:` de la retirada), S4 (formato de `strace -ff` en el runner x86_64 con Claude Code: la conexión `127.0.0.1:9` de
  clase `local` de la invocación `a9998` y ninguna sesión no cortada con traza ilegible), S5 (Claude Code en `-p` carga
  `~/.claude/skills`, y Bash hereda el proxy y `KITLEGAL_CACHE_DIR` y corre sin sandbox), S6
  (`CLAUDE_CODE_OAUTH_TOKEN` autentica y acepta el modelo), S7 (la retirada de Python: qué trae la imagen, que se puede buscar y retirar como root y que no
  rompe el job), S10 (`timeout --kill-after=10s 240s` y los códigos 124 y 137; el `codigo_de_la_sesion` 0 de las
  terminadas), S11 (Claude Code solo necesita `api.anthropic.com`) y S12 (identificar la ejecución por el último evento
  `labeled`, el formato de los instantes y `workflowName`). **Pendientes de T031 (ejecución de cierre,
  `gates/evals-cierre.md` y `gates/aceptacion.md`)**: S9 (una sesión de Haiku con ≤ 30 turnos cabe en 240 s; cada sesión
  con código 124 o 137 se anota como evidencia en contra), el veredicto `aprobado` con 10 de 10 positivas y las dos de
  no activación, ninguna petición llegada a la red de una fuente, las tres líneas de `sin_python` y la lista de ficheros
  cambiados entre el commit evaluado y la cabeza solo bajo el directorio del hito (SC-001 a SC-003, FR-080 a FR-082).
  S3 (cada búsqueda del manifiesto devuelve su norma y existen los bloques) ya lo comprobó la persona en la pausa de
  T008 y lo vigila `make ci` desde entonces.
- **T030 `[plataforma]`, cuatro intentos sin marcar**: los cuatro publicaron la rama, abrieron esta propuesta con este
  fichero como cuerpo, leyeron `ci` y los cuatro estados de Codecov (`codecov/project`, `internal/core`,
  `internal/cli` y `codecov/patch` con objetivo `auto`, todos en verde sobre `417635e`, `857ec46` y `537e5d6`) en
  `gates/evidencia-plataforma.md` y comprobaron el secreto y las dos etiquetas; cada prueba de red descubrió un
  defecto que solo la plataforma podía mostrar (la purga de paquetes, T033; `bubblewrap` y el modo de permisos, T036,
  y el argv cortado, T037; el relleno de `vfork()`, T038 y T039; el protocolo de lectura de bloques y la forma de la
  cita frente a lo grabado, T040). Con el intento 4 el job queda comprobado de punta a punta (S4 se cumple); queda
  repetir la prueba de red sobre la cabeza con T040 (y con lo que la persona decida de las evals 05 y 07), en un
  intento que concede una persona, porque el contador de la tarea vuelve a estar agotado. **T031 `[plataforma]`,
  la última**: la ejecución de cierre y la aceptación. Si la
  plataforma descubre un defecto, el arreglo va en una tarea nueva antes de T031 (plan, obligación 1). Ninguna tarea
  fusiona, empuja a `main`, fuerza ni etiqueta (ADR 0007).
- **Lo humano que queda**: el secreto `CLAUDE_CODE_OAUTH_TOKEN` (token de `claude setup-token`) y las etiquetas
  `evals` y `evals-prueba-de-red` ya están dados de alta (comprobados por T030); la pausa del workflow por las rutas sensibles (`schemas/`, `testdata/`); la revisión
  (`/code-review`, `/security-review`) y la fusión. **Y una decisión antes del intento siguiente de T030** (detalle en
  `gates/tarea-T030.md`, «Lo que decide la persona»): las evals 05 (artículo 59 del TRLRHL) y 07 (artículo 25 de la
  LRJSP) dependen de que el modelo sepa el número del artículo, porque el índice de la fuente no tiene rúbricas y en el
  job no se puede explorar (todo bloque no grabado responde 5); con el modelo de gama económica que fija la
  clarificación del spec, la 05 falló en los intentos 3 y 4 y la 07 en el 4, y T040 no las arregla. Opciones: dejarlas y
  repetir; que sus preguntas nombren el artículo, como la 01 y la 09 (tarea nueva sobre las dos evals y la tabla del
  contrato de evals §2; la recomendada, porque mide lo que el job puede medir); cambiar el modelo de las sesiones
  (reabre la clarificación del spec y D13); o una herramienta que encuentre artículos por materia dentro de una norma
  (la raíz, para el backlog: `docs/USO.md`).
- **`docs/PENDIENTES.md`** queda con sus dos entradas ajenas a H5 (absorber `refs/` cuando existan
  `docs/ARCHITECTURE.md` y `docs/SOURCES.md`; plegar `specs/*/gates/` cuando molesten). Las tres de H5 se retiraron
  como resueltas (FR-083, FR-084, FR-085).
- **Para los hitos siguientes**, ya en el roadmap y en el spec: la tabla completa de leyes vertebrales, `legal-core` y
  `territorio` (H6); las operaciones de grafo de `boe` y `graph check` sobre lo consultado (H7); resolver cada cita con
  `cita` (H8); `boe-fiscal` como primera vertical sobre esta base, con el campo `vertical` (backlog «Verticales», ADR
  0012); la robustez de `scripts/` frente a zip, Windows o marketplace y la versión de la skill (backlog
  «Distribución»).
- **Lo que no se hizo, a propósito**: ni `make evals`, ni `scripts/evals.sh`, ni `scripts/grabar-evals.sh`, ni
  `make verify-sources`, ni `KITLEGAL_RECORD`, ni §12 del quickstart: todo lo que toca red o plataforma queda para T030 y
  T031 o para una persona. Ninguna grabación se rehízo y ninguna fecha de revisión la escribió el ejecutor.
