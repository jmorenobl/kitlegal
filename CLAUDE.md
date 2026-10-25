# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Qué es kitlegal

`kitlegal` es un conjunto de **skills agénticas** para consultar fuentes legales públicas españolas (BOE, PLACSP, BDNS, BORME, EUR-Lex…) y llevar un asunto ante cualquier administración, para despachos y para la ciudadanía, apoyadas en un binario Go multicall que les da **herramientas deterministas**. El producto son las skills; el binario es su herramienta (constitución, principio VIII; ADR 0008). Todo es genérico para cualquier materia y cualquier municipio de España (principios IX y X; ADR 0009 y 0044), y se valida primero en lo administrativo y en la Comunidad de Madrid con Leganés.

## Dónde estamos

Este fichero no lleva el estado del proyecto, que caduca con cada hito: se mira en el repositorio.

- **Hito en curso**: `ls docs/roadmap/in-progress/`; si está vacío, no hay ninguno lanzado. El siguiente es el primero de `docs/roadmap/todo/` por su prefijo `NNN-`.
- **Lo entregado**: `docs/HISTORIAL.md`, una viñeta por hito con sus cifras de cierre, y la historia del workflow, de la constitución y del orden del roadmap. Al fusionar un hito, su viñeta se escribe allí y en ningún otro sitio.
- **El rumbo**: `docs/ROADMAP.md` y el ADR 0044 (fases: la base legal genérica, lo administrativo sobre la base, despachos y asesorías; las materias, al backlog).
- **Versiones**: las releases, `git tag --sort=-v:refname`; el workflow `hito`, el campo `version` de `.specify/workflows/hito/workflow.yml`; la constitución, su última línea; el modelo que decide las evals, `MODELO_DE_EVALS` en `.github/workflows/evals.yml`.
- **Agentes soportados**: solo lo que se ha probado; los demás se documentan como «sin probar» (`README.md`, «Otras instalaciones, sin probar»).

## Cómo se trabaja

**Método de trabajo: spec-kit por hito.** El proyecto se implementa hito a hito (`docs/ROADMAP.md`, con un fichero por hito en `docs/roadmap/`: `done/`, `in-progress/`, `todo/` en orden y `backlog/`) con el workflow `hito` de spec-kit (`.specify/workflows/hito/workflow.yml`, documentado en `docs/WORKFLOW.md`). La persona está solo en los extremos (ADR 0018): escribe la entrada —el fichero del hito en `docs/roadmap/`, que mueve a `docs/roadmap/in-progress/` para lanzarlo— y lee el informe final, que llega como cuerpo de la propuesta de cambio; en medio no hay pausas humanas. La constitución `.specify/memory/constitution.md` recoge principios, restricciones y el «Criterio de decisión autónoma» que rige `clarify` y los gates automáticos: siempre la mejor solución sin atajos; lo no especificado no se implementa; ante alcance, frontera humana, privacidad, TOS o decisiones cerradas, la lectura conservadora (no se implementa) registrada como supuesto. Ningún modelo detiene el run: solo el rechazo de entrada del spec y las causas mayores que detecta el código (DAG bloqueado, dependencia externa inaccesible, credencial ausente, presupuesto de tiempo). Artefactos por hito en `specs/NNN-hN-slug/`. Lanzar con `scripts/hito.sh H<n>`, que supervisa el run y toma el candado de sesión única. La primera tarea escribe la suite de aceptación, que queda congelada; los datos externos los graba un paso sin modelo; una tarea que no sale va a cuarentena; el cierre (push, propuesta de cambio, CI y evals sobre la cabeza) lo hace el workflow tras la revisión final, y fusionar lo decide siempre una persona (protección de `main` en GitHub y permisos; ADR 0018 y 0021), y el agente lo ejecuta solo a petición expresa en una sesión interactiva (ADR 0020). Detalle en `docs/WORKFLOW.md`.

Skills de agente instaladas (no son parte del producto): el conjunto `samber/cc-skills-golang` vive en `.agents/skills/` (registro en `skills-lock.json`); `.claude/skills/` son symlinks a ese directorio. Spec-kit tiene dos integraciones: `claude` (por defecto, skills en `.claude/skills/`; la única que ejecuta el workflow `hito`) y `agy` para Antigravity (skills `speckit-*` en `.agents/skills/`; `AGENTS.md` es un symlink a este fichero). `specify integration status` da error `unsafe-multi-install` por diseño: las dos no comparten ficheros. Tras `specify integration use`, restaurar el bit de ejecución con `git checkout -- .specify/scripts/bash`. Al escribir Go, `golang-how-to` orquesta el resto (`golang-cli`, `golang-project-layout`, `golang-testing`, `golang-database`…). Para la visibilidad de la web (`docs/SEO.md`): el servidor MCP `gsc` de `.mcp.json` lee Google Search Console con una clave en `~/.config/kitlegal/` (nunca en el repositorio), la skill `seo-semanal` (`.agents/skills/`, enlace en `.claude/skills/`) da el informe semanal, y el plugin `claude-seo` (marketplace en `.claude/settings.json`) hace las auditorías.

Antes de escribir código, lee los documentos semilla de `refs/` en este orden:

1. `refs/mapa-sistema-legal-skills.md` — sistema legal español, fuentes con semáforo de automatizabilidad (🟢 API / 🟡 scraping / 🔴 requiere identidad humana) y catálogo de skills.
2. `refs/kitlegal-estructura-y-ecosistema.md` — estructura del monorepo, convenciones del multicall, packs por vertical, distribución (skills, plugin, MCP, librería Go).
3. `refs/kitlegal-grafo.md` — grafo legal sobre ELI en SQLite (mundo público + asunto privado), alimentación por uso, `graph check` y reglas de anomalías.
4. `refs/boe.py` — solo para H4: copia congelada (2026-05-14) de la skill `boe-fiscal` en Python, el fuente que se porta a `internal/source/boe`. Material de lectura, no código del producto: no se mantiene y no entra en ningún gate. Tampoco se ejecuta dentro del repositorio: la única excepción es la pausa `[datos]` que deriva las referencias del diff de aceptación, donde un guion de un solo uso, fuera del repositorio y sin versionar, importa este fichero sin modificarlo y le sustituye la red por las grabaciones. Ni el producto, ni los tests, ni `make ci` necesitan Python.

`refs/00-README.md` resume el estado, el primer hito y las decisiones cerradas. Ante conflicto, la constitución y `docs/ROADMAP.md` prevalecen sobre `refs/`.

## Lo que ya se usa

`kitlegal boe articulo BOE-A-2015-10565 a21` funciona en Go con caché SQLite (H4) y la skill genérica `boe-legislacion` (H5) consulta y cita cualquier norma consolidada del BOE con el binario; `kitlegal territorio resolver Leganés` sitúa la pregunta en su territorio y `legal-core` (H6) empieza por ahí. Desde H19 cualquiera lo instala sin clonar (`curl … | sh`, Homebrew, Scoop, `.deb`/`.rpm`) y pone las skills en su proyecto con `kitlegal skills install`. `refs/boe.py` fue el patrón de H4 y hoy es solo lectura. La skill `boe-fiscal` (Python, fuera de este repositorio) se migra con la vertical fiscal, la primera candidata de la fase 3, cuando el uso lo pida: deja de ser una skill y pasa a ser el paquete fiscal sobre las skills de capacidad, con sus consultas actuales como evals (ADR 0044).

## Comandos

El `Makefile` es la única superficie de invocación; `make help` lista los objetivos (`build`, `test`, `lint`, `ci`, `hooks`, `skills-sync`, `release`…). Tooling:

- Módulo `github.com/jmorenobl/kitlegal`, CLI con Kong (`github.com/alecthomas/kong`), SQLite sin cgo (`modernc.org/sqlite`), release con goreleaser.
- Tests unitarios offline contra fixtures grabados en `testdata/<fuente>/`. Graba el paso `grabar_datos` del workflow, sin modelo, con `KITLEGAL_RECORD=1` y solo fuentes con fila revisada en `docs/SOURCES.md` de `main` (ADR 0018); de los tests, solo `scripts/verify-sources.sh` (CI nightly) toca la red.
- `scripts/skills-sync.sh` regenera `skills/*/references/*.md` desde `data/*.yaml` y los symlinks de `scripts/`.

## Explorar el código con CodeGraph

Si hay un índice (`.codegraph/`, local y fuera de git; se crea con `codegraph init`), úsalo antes de recorrer el código con grep y lecturas: `codegraph_explore` (o `codegraph explore "<pregunta>"` desde la terminal y en subagentes) para «cómo funciona X» y los flujos entre applets, kernel y `internal/`; `codegraph callers`, `callees` e `impact <símbolo>` antes de cambiar una firma o un contrato (`Applet`, `Resultado`, `Procedencia`); `codegraph affected <ficheros>` para saber qué tests correr. No sustituye a leer el fichero que vas a editar ni a `docs/`, `data/` y `specs/`, que son YAML y Markdown. `codegraph.json` deja fuera `.agents/skills/` y rebaja `specs/`, `evidencias/`, `scripts/`, `testdata/` y `refs/` en el ranking.

## Decisiones ya tomadas (no reabrir)

Cada una, en una línea; su porqué y su detalle, en el ADR que la cita. Ninguna se reabre en spec, plan ni clarify (constitución).

**Producto y alcance**

- **Skills primero** (ADR 0008): cada hito entrega o mejora una skill medible con evals, o protege las existentes. Solo va a Go lo que exige determinismo o verificabilidad (fuentes, parseo, caché, fechas y plazos, ids, hashes, grafo); una herramienta que ninguna skill usa no se construye. Se planifica de fuera adentro y se implementa de dentro afuera.
- **Componente legal general** (ADR 0044): cualquier materia, para la ciudadanía y los profesionales. Fase 1, la base legal genérica; fase 2, lo administrativo sobre la base; fase 3, despachos y asesorías; las materias, al backlog, nunca como fase. La entrada de un hito es su fichero de `docs/roadmap/`; lanzarlo es moverlo a `in-progress/` en `main`; el orden de `todo/` es el prefijo `NNN-` y el id `H<n>` no cambia nunca.
- **Genericidad por materia** (principio X; ADR 0044): ningún caso especial para una materia en código, `data/` ni protocolo de skill. Herramientas y skills por capacidad (normativa, jurisprudencia, plazos, territorio y destinatarios, documentos); lo que varía por materia son datos y referencias, y una materia entra como paquete sin tocar código ni protocolo. Lo que el paquete necesita y la base no tiene es un hito de la base, genérico, validado con dos implementaciones.
- **Base antes que vertical** (ADR 0012, precisado por el 0044): una vertical (fiscal, laboral, mercantil…) es una materia empaquetada para un profesional —normas con `vertical:` en `data/normas.yaml`, una referencia por skill de capacidad, cobertura y evals—, no una skill ni herramientas propias. Un pack (`packs/<pack>/pack.yaml`) llega con la segunda vertical. Entran en la fase 3 en el orden de `docs/USO.md`; la fiscal, con la migración de `boe-fiscal`, es la primera candidata.
- **Cobertura en positivo** (principio II; ADR 0044): se cubre lo declarado, por capacidad y materia, y lo demás no. Fuera de lo declarado, la skill lo dice con la forma fija, cita la norma que haya leído y no da de memoria ningún plazo, órgano ni vía, ni aplica los de otra materia; lo que el agente busque con las herramientas genéricas va marcado como no verificado. Lo que la ley regula aparte (DA 1ª.2 LPAC; art. 112.4) se declara no cubierto con la norma que lo rige.
- **Genericidad territorial** (principio IX; ADR 0009 y 0027): ningún caso especial para un municipio en código, `data/` ni skills. Datos de registros nacionales (INE, DIR3); lo territorial, por comunidad o boletín en `data/territorio/` y `data/boletines/`; los municipios de cada asunto van en su `.kitlegal/config.yaml`. Validación en la Comunidad de Madrid (Leganés); otros territorios, bajo demanda. Sin cobertura completa, ninguna skill concluye «no existe».
- **Frontera humana** (ADR 0004): solo fuentes públicas. Presentar, notificar o cualquier acción con identidad termina en «fichero listo para firmar» y exit 6; nunca un POST a una sede.
- **CENDOJ** (ADR 0036): no se consulta de forma automática ni se sortea su CAPTCHA, por ningún medio. La persona busca con su navegador lo que le prepara la skill y trae el documento; el binario solo comprueba, sin red, que es el pedido. Ninguna sentencia se da por existente ni se cita sin su documento delante. Nunca búsqueda automática por materia o texto libre, ni descarga, ni ingesta.
- **Sin telemetría** (ADR 0027): nada de preguntas, asuntos ni datos de uso a un servidor propio, tampoco con consentimiento.
- Reglas invariantes en todas las skills: nunca inventar contenido legal; cada afirmación con cita resuelta por `cita`; distinguir ley y reglamento; señalar la variación autonómica; declarar no cubierto lo no declarado, sin plazos, órganos ni vías de memoria; no aplicar el procedimiento común a lo que la ley regula aparte.

**El binario**

- **Nombre único** `kitlegal` para repositorio, módulo, binario y directorios (`.kitlegal/` en el directorio del asunto, `~/.cache/kitlegal/` para caché y grafo del mundo).
- **Multicall**: un solo ejecutable; `os.Args[0]` o el primer argumento eligen el applet. Las skills invocan `kitlegal <applet> …` desde el `PATH` (ADR 0019).
- **Contrato de applet y resultados** (ADR 0005 y 0023): 0 hecho · 2 argumentos · 3 no encontrado · 4 fuente no disponible · 5 límite o TOS · 6 requiere identidad humana · 7 conflicto con el estado local · 1 defecto inesperado. Una verificación que encuentra algo sale con 0 y sus hallazgos en `data`; ningún hito elige códigos.
- **Sobre de salida** (ADR 0006): `{ok, fuente, url, fecha_consulta, hash, data}` en todo applet; sin `fuente`, `url`, `fecha_consulta` y `hash` no hay cita. La salida legible va en el `Resultado` (ADR 0026).
- **Convenciones de agente** en `internal/cli`: `--json`, `--timeout`, `--offline`, `--dry-run`, `--describe` (JSON Schema del que salen las herramientas MCP y la tabla de comandos de cada `SKILL.md`), `--no-graph`, `--asunto`. `internal/cli` no se extrae a librería hasta que tres applets repitan el patrón.
- **HTTP** solo por `internal/httpx` (ritmo por sitio, `robots.txt`, User-Agent `kitlegal/x.y (+https://kitlegal.es/bot)`), solo GET y HEAD. Un CAPTCHA o un bloqueo no se sortean; nada de paralelismo agresivo contra PETETE, DYCTEA o HJ del TC.
- **Grafo** (ADR 0014): ids naturales (ELI, ECLI, DIR3, NIF, código INE) en `world.db`; el del asunto (`.kitlegal/case.db`) solo guarda ids del mundo y no se sincroniza ni exporta por defecto. `Persona` nunca lleva NIF ni DNI; un dato sin `source` no entra. Las operaciones viajan en el `Resultado` y las aplica el kernel. El grafo es índice y validador, nunca fuente del texto que se cita.

**Skills, distribución y web**

- **Skills sin código** (ADR 0046 y 0019): `skills/<nombre>/` lleva `SKILL.md` —cuerpo de como mucho 11 000 caracteres, descripción en tercera persona de como mucho 900— y `references/`, plano: lo escrito a mano y lo **generado** desde `data/*.yaml`, la única fuente de verdad (cabecera `<!-- generado desde … no editar -->`); sin `scripts/`. La guía es `docs/SKILLS.md`.
- **Distribución** (ADR 0019, 0025 y 0035): el binario lleva las skills dentro y se instala con el gestor de paquetes de cada plataforma (goreleaser en cada etiqueta); `kitlegal skills install` las pone en `./.agents/skills/` (con `-g`, en `~/.agents/skills/`) y enlaza cada host que no lee ese directorio. Todo corre en el equipo de quien lo usa: el servidor MCP es local, por stdio; el remoto está aplazado y pediría un ADR. `make install` es solo el bucle de desarrollo.
- **Web** (ADR 0024 y 0034): `web/` (Astro, HTML estático) en https://kitlegal.es, fuera de `make ci` (`make web`). Ninguna cita escrita a mano: los fragmentos de `web/src/data/citas.yaml` se comprueban al construir contra los sobres del binario (`make web-citas`), y su `version` hace fallar la construcción si el BOE trae otra versión del bloque. Habla al usuario final: `/` ciudadanía, `/despachos/`, una página por consulta y el buzón `info@kitlegal.es`. Las fotografías no enseñan texto legible, marcas ni logotipos de una institución.

**Proceso y evals**

- **Fusionar, etiquetar y publicar los decide una persona** (ADR 0018, 0020 y 0021); el agente los ejecuta solo a petición expresa en una sesión interactiva.
- **Modelo que decide las evals** (ADR 0031): el id al que resuelve el alias `sonnet` en la versión de Claude Code del job, los dos fijados juntos en `.github/workflows/evals.yml`. Antes de lanzar un hito, `claude -p --model sonnet --output-format json 'Responde solo: ok'` dice en `modelUsage` si el alias cambió; si no da `MODELO_DE_EVALS`, el hito espera a la propuesta que suba los dos. Que el modelo nuevo no pase es un defecto de la skill.
- **Juez con modelo** (ADR 0037): lo que tiene forma o es un hecho de la sesión se comprueba con un guion; el significado lo juzga `claude-opus-5-5`, con una pregunta cerrada por clase y tres votos unánimes. Una clase solo decide si su juez se ha medido contra casos etiquetados, y la medida la repite una persona con su etiqueta. Nunca se juzga una respuesta con una lista de palabras; el juez no corre en un paso de un run.
- **Umbral de una clase del juez** (ADR 0038 y 0045): cero donde hay forma, tasa donde hay significado. La tasa solo se estrecha, y la baja una persona con la serie de cierres (`scripts/workflow/serie.sh`); lo reutilizado de una medida de `main` no cuenta.
- **Sesiones de un run** (ADR 0032): sin credenciales ni sesiones con modelo, en el sandbox de `scripts/claude-modelo.sh`, que es la garantía, no los permisos de `.claude/settings.json` ni un gancho. El token del sondeo de evals vive en un llavero bloqueado, nunca en un fichero ni en el entorno de un run, y el sondeo lo lanza una persona. La ruta de un secreto no se escribe en la memoria del proyecto.

## Arquitectura prevista (resumen)

```
skills/<skill>/             EL PRODUCTO: SKILL.md + references/ (generadas); empotradas en el binario, que las instala
packs/  data/  evals/       packs de verticales · fuente de verdad (normas, territorio, boletines, festivos, anomalías) · evals de skills
cmd/kitlegal/main.go        multicall → internal/app (registro y dispatch de applets)
internal/cli                Kong + flags globales + exit codes
internal/core/{territorio,cita,plazos,competencia,schema,ids}
internal/source/<fuente>    un adaptador por fuente: Source{Name, Fetch(ctx,req), TTL, Terms}
internal/{httpx,cache,store,graph,analysis,render,docgen}
pkg/legalkit                API pública pequeña y estable
schemas/  testdata/  mcp/  plugin/  docs/
```

Todo applet que observe identidades devuelve sus operaciones de grafo en el `Resultado` y el kernel las entrega al `GraphStore` (`internal/graph`) tras presentar la salida (`--no-graph` = store nulo; ADR 0014). Los packs (`packs/<pack>/pack.yaml`) agrupan verticales para un tipo de profesional —una gestoría pide fiscal, laboral y mercantil juntas— y llegan con la segunda vertical, por la distribución (backlog; ADR 0044).

## Pendientes

Los pendientes de **estructura del repositorio** (dónde van los fixtures, los tres directorios `skills`, los symlinks de las skills al binario, `refs/`…) están en `docs/PENDIENTES.md`, cada uno con el hito en que se decide.

## Pendiente de verificar antes de fijar en código o `data/`

- Identificadores BOE de la tabla de leyes vertebrales (algunos están de memoria; comprobar con `kitlegal boe buscar`).
- Endpoints actuales de BDNS (swagger) y nombres de los feeds Atom de PLACSP (y qué llega agregado desde plataformas autonómicas).
- Términos de uso de PETETE (DGT), DYCTEA (TEAC) y del buscador HJ del TC.

## Convenciones de idioma

Documentación, nombres de applets/verbos (`articulo`, `buscar`, `calcular`) y claves JSON (`fuente`, `fecha_consulta`) van en español. Identificadores Go siguen la convención habitual del lenguaje.
