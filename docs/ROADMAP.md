# kitlegal: roadmap de hitos

Roadmap para implementar lo descrito en `refs/`. Enfoque lean: hitos de 0,5 a 3 días, cada uno termina en algo que se puede usar (o que protege lo que ya se usa). No hay fechas; hay orden y dependencias. Lo que no está aquí no se construye hasta que un hito lo necesite (YAGNI).

Las decisiones cerradas de `CLAUDE.md` y `refs/00-README.md` no se repiten: este documento las asume.

**Qué se construye y en qué orden.** El producto son las **skills**; el binario Go es la capa de herramientas deterministas que usan (constitución, principio VIII). kitlegal es un **componente legal general**: ayuda a entender y resolver una situación legal de cualquier materia, a la ciudadanía, que quiere saber qué puede y qué no puede hacer antes de actuar, y a los profesionales, que necesitan más profundidad. El orden, desde el 2026-10-21 («Punto de partida y rumbo», abajo): consultar y citar normativa (fase 0) y el terreno y la memoria (H6 a H26.1) están hechos; ahora, **la base legal genérica**: capas, interfaces y skills bien hechas que sirven para cualquier materia y se conectan de forma genérica a las fuentes que se pueden usar (fase 1); luego **lo administrativo sobre la base: llevar un asunto ante cualquier administración** —expedientes y escritos, para un despacho o para un ciudadano— (fase 2); y después **los profesionales, despachos y asesorías**: los plazos procesales, el Constitucional y las verticales (fase 3). Las materias y todo lo demás —distribución, territorios, fuentes, profundidad del BOE, transparencia y fiscalización— son **backlog**, que se reprioriza con el uso real (ADR 0013). Todo es **genérico para cualquier municipio de España** y se valida primero en el territorio de referencia, la Comunidad de Madrid con Leganés (principio IX); los demás territorios llegan como datos y adaptadores, sin tocar las skills. Y todo es **genérico por materia**: una materia entra como datos, referencias, cobertura y evals sobre los motores de la base, sin tocar el código ni el protocolo de las skills; una **vertical** (fiscal, laboral, mercantil…) es una materia empaquetada para un tipo de profesional (ADR 0012, con lo que precise H29). El grafo llega en tres piezas: memoria pronto, asunto con los expedientes, instrumento tarde (ADR 0014).

---

## Punto de partida y rumbo (2026-10-21)

**De dónde partimos.** En `main`, con la v0.8.1, están la fase 0 y lo que era la fase 1 hasta H26.1: territorio, el grafo del mundo, la redacción a una fecha, `cita` con el cotejo de la sentencia que trae la persona, `plazos`, el servidor MCP y la instalación sin terminal, y el juez de las evals; con tres skills, `boe-legislacion`, `legal-core` y `jurisprudencia`. Por capacidades:

- **Lo que ya es genérico.** Leer y citar cualquier norma consolidada del BOE, de cualquier materia y en cualquier fecha; preparar la consulta de una sentencia de cualquier orden y cotejar el documento que trae la persona. Una familia a la que una aseguradora renovó y cobró una póliza tras el fallecimiento de la asegurada leyó y citó bien la Ley 50/1980, de Contrato de Seguro (`docs/USO.md`, 2026-10-21).
- **Lo que es administrativo por construcción.** El cómputo de plazos: dos cómputos fijados en el código, `lpac` y `contencioso`, con el sábado y el domingo siempre inhábiles, y los tipos de plazo enumerados en Go. La cobertura por materia, que se declara en negativo y respecto de la LPAC (`data/cobertura.yaml`: lo que la ley regula aparte del procedimiento común), de modo que una pregunta civil no cae en ninguna fila y nada la declara no cubierta. `territorio`, cuyo eje es municipio, comunidad, DIR3 del ayuntamiento y boletines. La descripción de `legal-core` («derecho público»). Y la fase 2 prevista, un expediente ante una administración con los escritos de la LPAC. Al caso de la póliza le faltó todo lo que rodea a la norma: los plazos, a quién reclamar y qué vía seguir.
- **Lo que no sigue las prácticas de las skills.** La especificación Agent Skills (agentskills.io) y la guía de autoría de Anthropic recomiendan un cuerpo de `SKILL.md` por debajo de unos 5000 tokens, con el detalle en referencias que se cargan cuando hacen falta, a un solo nivel, e índice en las de más de 100 líneas. El cuerpo de `boe-legislacion` tiene unos 24 400 caracteres y su descripción 1023 de los 1024 que admite la especificación; el de `legal-core`, unos 19 200; y sus `references/` son tablas de 15 a 24 líneas, porque todas se generan desde `data/` y el protocolo no tiene otro sitio que `SKILL.md`. El límite de menos de 300 líneas se cumple con líneas de 120 columnas.
- **Lo que tardan las evals.** Cualquier cambio en `skills/`, `evals/`, `data/`, `cmd/` o `internal/` mide las tres skills enteras en los dos modos; el peor caso de `boe-legislacion` pasa de cinco horas (entrada de H20).

**Cómo seguimos.** Tres fases y un backlog:

1. **Fase 1, la base legal genérica** (H29 a H38). Capas e interfaces bien hechas —fuentes, cómputo, destinatarios, documentos, cobertura— y skills con la estructura de la especificación. Se puede preguntar de cualquier materia: lo que la base cubre se responde con la norma citada; lo que no cubre se declara, y el agente lo busca con las herramientas genéricas y lo marca como no verificado. Es la limitación inicial, y ya es un avance. Cada abstracción nueva se valida con dos implementaciones reales, no con una y una promesa (§5, la regla de los tres applets). Las evals pasan antes a medir solo lo que cambia (H30), porque la fase reescribe las skills una a una.
2. **Fase 2, lo administrativo sobre la base.** Lo que ya hay pasa a ser el primer paquete de materia; después, el asunto ante cualquier administración y sus escritos, los boletines y las ordenanzas, y la doctrina del BOE. Desbloquea todo lo que se haga con la administración.
3. **Fase 3, los profesionales: despachos y asesorías.** Los plazos procesales, el Constitucional por el BOE y las verticales —fiscal, laboral, mercantil— en el orden que pida `docs/USO.md`. Son quienes más profundidad piden (ADR 0040; `docs/USO.md`, 2026-10-19).
4. **El backlog.** Las materias, una línea cada una, y los grupos de siempre. Una materia no es una fase: entra como paquete sobre los motores de la base.

Cada hito está desde entonces en su fichero, en `docs/roadmap/`, y el directorio dice si está hecho, en curso, en la cola de `todo/` o en el backlog (§4). Correspondencia con la organización anterior, sin renumerar: lo que era la fase 1 queda como «Terreno y memoria», hecha hasta H26.1; de lo que le quedaba, H8 y H28 van a la fase 2 y H9.1 y H27 a la fase 3; la fase 2 (H10 y H11) sigue siendo la 2; de la fase 3, H15 y H16 pasan a la fase 2 y H12, H13, H14, H17 y H18 a «Transparencia y fiscalización municipal», detrás de la fase 3. El ADR que fija esta visión —y sustituye en eso la frontera de la base del ADR 0027 y el orden del ADR 0040— es H29.

---

## 0. Cómo leer este roadmap

- **Fase** = épica. Al final de cada fase hay alguien que gana algo en su trabajo real: quien consulta normativa (fase 0); quien pregunta de cualquier materia y recibe la norma citada, la cobertura declarada y, donde no la hay, lo que el agente buscó marcado como tal (fase 1); quien lleva un asunto ante cualquier administración, sea un despacho o un ciudadano (fase 2); y los despachos y las asesorías, con sus plazos procesales, el Constitucional y su vertical (fase 3). Lo hecho entre la fase 0 y el 2026-10-21 está en «Terreno y memoria» (§4).
- **Hito** = PR única, rama corta desde `main` (trunk-based), squash-merge con CI verde. Un hito no se cierra sin cumplir la *Definition of Done* (§1).
- **Numerados y backlog.** Un hito tiene número y especificación desde que entra en `docs/roadmap/todo/`, y su fichero no cambia de nombre ni de número después: solo de directorio, de `todo/` a `in-progress/` y a `done/`. El orden de `todo/` es el prefijo de cada fichero. Lo demás es un backlog de candidatos sin número, agrupados por tema y sin orden entre grupos (`docs/roadmap/backlog/`); los de «Transparencia y fiscalización municipal» conservan el número que tenían como fase 3. Un candidato recibe número cuando pasa a `todo/`, y lo que decide que pase es la bitácora de uso `docs/USO.md` (§6).
- Cada hito indica: objetivo, entrega usable, alcance, controles (tests y calidad) y criterio de aceptación. A partir de H5, el objetivo se formula como lo que una skill pasa a poder resolver.
- **Lo que no es un hito.** La web (ADR 0024), la presencia en directorios de skills y de servidores MCP, un benchmark público y los acuerdos con colegios o universidades no entregan ni protegen una skill (principio VIII): no entran aquí y los decide la persona. Si necesitan código del producto, como el servidor MCP (H21, ADR 0035), ese código sí es un candidato (ADR 0027).
- Historia de la numeración: rehecha el 2026-09-11 (ADR 0008 y 0009); el 2026-09-12 la fase 0 pasó a entregar `boe-legislacion` en H5 (ADR 0012); el 2026-09-13 se reordenó por tiempo hasta el uso desde H6 y el grafo pasó a tres piezas (ADR 0013 y 0014); el 2026-09-26 H19 (release) se adelantó a la fase 1, detrás de H6, con las skills dentro del binario (ADR 0019); el 2026-09-27 la fase 2 pasó a ser genérica —un asunto ante cualquier administración, con la cobertura por materia declarada—, entró H20 (la redacción a una fecha) entre H7 y H8 con el siguiente número libre, y H8 pasó a validar textos ajenos (ADR 0027); el 2026-10-01 el servidor MCP local y su instalación sin terminal —la extensión de escritorio y el plugin— salieron del backlog de distribución como H21 y H22, por delante de H20, porque ninguna de las dos audiencias usa una terminal (ADR 0035); el 2026-10-03, comprobar que una sentencia existe por el formulario del CENDOJ salió del backlog de fuentes como H23 (ADR 0036), y H23 y H9 pasaron por delante de H20: lo que más daño hace a un despacho y lo que más busca la ciudadanía (`docs/USO.md`, 2026-10-01 y 2026-10-02); el 2026-10-04, el juez con modelo de las evals entró como H24, por delante de H23, porque el umbral de H23 mide un significado y la lista de expresiones no sirve para eso (ADR 0037; `docs/USO.md`, 2026-10-04); el 2026-10-06, al comprobar la entrada de H23 antes de lanzarlo, el juez de `jurisprudencia` salió a un hito propio, H25, detrás de H23, porque una clase solo decide con el juez medido contra respuestas reales de la skill, y la constitución (2.12.0) ganó la excepción del formulario de consulta (ADR 0036, «Enmienda»); el 2026-10-07 el buscador del CENDOJ respondió con un CAPTCHA al cliente identificado, el run de H23 se cerró sin entrega y H23 pasó a ser la skill con la comprobación hecha por la persona, sin consulta automática, y la constitución (2.13.0) volvió a GET y HEAD sin excepción (ADR 0036, «Enmienda: el buscador pide un CAPTCHA»; `docs/USO.md`, 2026-10-07). El 2026-10-19, detrás de H9, la fase 1 pasó a ordenarse por el trabajo de un despacho: H9.1 (los plazos procesales civiles y sociales) entre H20 y H8, y H27 (el TC por el BOE) y H28 (la doctrina publicada en el BOE) detrás de H8, con la fase 3 por delante de los candidatos de despacho del backlog (ADR 0040; `docs/USO.md`, 2026-10-19). El 2026-10-20 entró H26.1, delante de H9.1: las evals sirven lo que la skill lleva a leer y hacen fallar a propósito la lectura que quieren probar, porque una lectura grabada que la caché de la sesión no servía puso en rojo `legal-core` en la #175 (ADR 0043). La correspondencia con la numeración anterior está en `docs/ADR/0013-roadmap-por-tiempo-hasta-el-uso.md`. El 2026-10-21 el roadmap se reorganizó en base, administrativo y profesionales («Punto de partida y rumbo»): lo que era la fase 1 queda como «Terreno y memoria», hecha hasta H26.1; la fase 1 nueva es la base legal genérica (H29 a H38, con los siguientes números libres); la fase 2 recibe H8, H15, H16 y H28; la fase 3 nueva, despachos y asesorías, lleva H9.1, H27 y las verticales; H12, H13, H14, H17 y H18 pasan detrás, a «Transparencia y fiscalización municipal», con sus números; y las materias entran en el backlog. El ADR que lo fija es H29 (`docs/USO.md`, 2026-10-21).

---

## 1. Definition of Done (aplica a todos los hitos)

1. `make ci` en verde: `fmt`, `lint`, `test` (con `-race`), `test-integration`, `vuln`, `schema-check`.
2. Tests unitarios offline para el código nuevo; fixtures grabados en `testdata/<fuente>/` si toca red.
3. Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de los paquetes autorizados (lo vigila `depguard`/`forbidigo`, ver §3).
4. Toda salida de applet pasa la validación contra `schemas/*.json` en test, y `make schema-check` falla si `schemas/` no coincide con lo que `--describe` emite.
5. Errores tipados que mapean a un exit code estable; ningún `panic` en rutas de usuario.
6. Si cambia un comportamiento visible: test e2e (`testscript`) actualizado y `CHANGELOG.md` (sección *Unreleased*).
7. Si cambia una decisión de arquitectura: ADR nuevo en `docs/ADR/`.
8. Si toca una fuente externa: fila de `docs/SOURCES.md` actualizada (licencia, TOS, rate limit, fecha de verificación) y caso en `scripts/verify-sources.sh`.
9. Cobertura de `internal/core/**` ≥ 85 % y global ≥ 70 % (umbrales en `codecov.yml`; no son objetivo, son red de seguridad).
10. Si entrega o cambia una skill (desde H5): evals en `evals/<skill>/` (activación y respuesta con citas esperadas) escritas antes que el código y en verde; `SKILL.md` < 300 líneas con frontmatter válido; `references/` y tabla de comandos regeneradas sin drift.
11. Si tiene dimensión territorial: e2e con un municipio cubierto y otro no cubierto, con la cobertura declarada en `data`; ningún caso especial para un municipio en código ni en `data/` (los municipios concretos solo aparecen en fixtures y evals).
12. Desde H7: todo applet que observe identidades devuelve sus operaciones de grafo en el `Resultado` (ADR 0014); un test comprueba que ninguna entra sin `source` y que aplicar dos veces el mismo resultado no duplica nada. Ningún verbo devuelve texto legal leído del grafo.
13. Desde H9, si calcula recursos o plazos, o fundamenta o genera un escrito: e2e o eval con un asunto del procedimiento común y otro de un procedimiento que la ley regula aparte (DA 1ª.2 LPAC), que se declara no cubierto, con la norma que lo rige y sin los recursos ni los plazos del común (ADR 0027).

---

## 2. Arquitectura objetivo (resumen operativo)

Dos capas: las **skills** (el producto: protocolo de razonamiento, conocimiento y comandos) y el **binario** (herramientas deterministas que las skills invocan). `data/` es la única fuente de verdad de ambas.

```
  skills/<skill>/  SKILL.md (protocolo) · references/ (generadas) · empotradas en kitlegal   ← EL PRODUCTO
  packs/<vertical>/pack.yaml                                    (agrupan skills por vertical)
  data/  normas · territorio · boletines · festivos · anomalias  (fuente de verdad; skills-sync genera references/)
                │  el agente ejecuta kitlegal <applet> <verbo> … (en el PATH) y recibe el sobre JSON
                ▼
                ┌────────────────────────── cmd/kitlegal (main, multicall)
                │
        internal/app ── composition root, registro de applets (Command + Registry)
                │
        internal/cli ── Kong, flags globales, mapeo error→exit code, sobre de salida,
                │       entrega de las operaciones de grafo del Resultado al GraphStore
  ┌─────────────┼─────────────────────────────────────────────┐
  │ internal/core (dominio puro, sin I/O)                     │
  │   ids · territorio · cita · plazos · competencia · schema │
  │   define los PUERTOS: Source, Fetcher, Cache, GraphStore, │
  │   Renderer, y los tipos de operación de grafo             │
  └─────────────┬─────────────────────────────────────────────┘
                │ implementan
   ADAPTADORES: internal/source/<fuente>  (Adapter, uno por fuente)
                internal/httpx            (Decorator chain: UA → robots → ratelimit → retry → record/replay)
                internal/cache, store, graph (Repository sobre SQLite)
                internal/render           (Strategy: json | table | markdown)
```

**Qué va a Go y qué no.** Va al binario lo que exige determinismo o verificabilidad: acceso a fuentes, parseo, caché, fechas y plazos, identificadores, hashes y grafo. El razonamiento, la interpretación, la elección de fuentes y la redacción de la respuesta viven en `SKILL.md` y `references/`. Si una capacidad puede resolverla el modelo con fiabilidad a partir de una referencia generada, no se convierte en applet.

**Territorio.** Los datos de cada municipio (INE, provincia, comunidad, DIR3) salen de registros nacionales; lo que varía por territorio se configura por comunidad o por boletín en `data/territorio/` y `data/boletines/`. Los municipios de cada asunto —el del interesado y el de la sede del órgano, los que mira el art. 30.6 LPAC— viven en el `.kitlegal/config.yaml` del asunto, no el de quien usa el kit (ADR 0027). Fuera del territorio cubierto, cada applet declara la cobertura dentro de `data`.

**Materia.** Las skills base y sus herramientas cubren el procedimiento común (LPAC) y el acceso a la información (LTAIBG) en cualquier materia y ante cualquier administración. Lo que la ley regula aparte —la DA 1ª.2 LPAC (aplicación de los tributos, Seguridad Social y desempleo, sancionadores tributario, social, de tráfico y de extranjería, extranjería y asilo), las vías que sustituyen a la alzada (art. 112.2) y las reclamaciones económico-administrativas (art. 112.4)— se declara no cubierto, con la norma que lo rige, y llega con su vertical (ADR 0027). [enmienda 2026-10-21: es la frontera vigente hasta H29, que la sustituye por la cobertura en positivo por capacidad y materia (H33): se cubre lo declarado, y lo demás se declara no cubierto y el agente lo busca marcado como no verificado.]

**Grafo.** Es memoria y validador, nunca fuente del texto que se cita (ADR 0014): dice qué hay que volver a comprobar, no qué dice el artículo. El texto se cita siempre del sobre de una consulta al applet de la fuente, que la caché sirve sin red cuando está vigente.

Patrones que se usan y por qué:

| Patrón | Dónde | Motivo |
|---|---|---|
| Adapter | `internal/source/<x>` | Cada fuente expone la misma interfaz `Source`; cambiar BOE no toca PLACSP |
| Command + Registry | `internal/app` | Un applet = un comando registrado; el multicall solo despacha |
| Decorator / middleware | `internal/httpx` | Retries, rate limit, robots y grabación componen sin acoplarse |
| Strategy | `internal/render` | `--json` manda; tabla y markdown son intercambiables |
| Repository | `cache`, `store`, `graph` | El dominio no sabe que hay SQLite |
| Null Object | `graph` con `--no-graph` | Las operaciones del `Resultado` siempre se entregan; el store nulo descarta |
| Errores tipados (sentinel + `errors.As`) | `internal/cli/errors.go` | `ErrNotFound`→3, `ErrSourceDown`→4, `ErrRateLimited`→5, `ErrHumanRequired`→6 |
| Functional options | constructores de `httpx`, `cache`, `graph`, applets | Configuración explícita y testeable sin globals |
| Configuración por datos | `data/territorio/`, `data/boletines/` | Un territorio o un boletín nuevo es un YAML, no código ni una skill nueva |
| Facade | `pkg/legalkit` (backlog · distribución) | API pública pequeña sobre los internos |

Reglas de dependencia (se hacen cumplir con `depguard`, no con buena voluntad):

- `internal/core/**` no importa nada de `internal/{source,httpx,cache,store,graph,render,cli,app}`.
- Solo `internal/httpx` importa `net/http`.
- Solo `internal/{cache,store,graph}` importan `modernc.org/sqlite` y `database/sql`.
- Solo `internal/cli` y `cmd/` llaman a `os.Exit`.
- Solo `internal/render` escribe en stdout; logs (`log/slog`) siempre a stderr.
- `internal/graph` no importa `internal/source/*` ni `internal/render`: recibe operaciones, no consulta fuentes ni presenta.

---

## 3. Estrategia de calidad y controles

Qué control, con qué herramienta, y en qué hito entra. Todo lo que entra se queda como gate de CI.

| Control | Herramienta | Entra en |
|---|---|---|
| Formato | `gofumpt` + `goimports` (vía golangci-lint formatters) | H0 |
| Lint | `golangci-lint` con `.golangci.yml` estricto: `errcheck`, `govet`, `staticcheck`, `revive`, `gocritic`, `errorlint`, `exhaustive`, `nilerr`, `bodyclose`, `noctx`, `contextcheck`, `sqlclosecheck`, `rowserrcheck`, `nolintlint`, `misspell` (es/en), `gocyclo`, `dupl`, `testifylint`, `thelper`, `paralleltest`, `tparallel`, `gosec` | H0 |
| Reglas de arquitectura | `depguard` + `forbidigo` (§2) y un test `internal/arch_test.go` que recorre `go list -deps` | H1 |
| Tests unitarios | stdlib + `testify`; table-driven; `t.Parallel()`; fixtures en `testdata/`; golden files para JSON de salida | H1 |
| Grabación/replay HTTP | `httpx` con `KITLEGAL_RECORD=1` (graba a `testdata/<fuente>/`); en test, replay determinista | H2 |
| Race detector | `go test -race ./...` en CI | H0 |
| Fuzzing | `go test -fuzz` sobre parsers: `ids`, `cita`, parseo XML/Atom; corpus mínimo versionado en `testdata/fuzz/` | H6, H8, H13 |
| Property tests | `testing/quick` o `pgregory.net/rapid` en `plazos` (idempotencia, monotonía, ida y vuelta) | H9 |
| Tests de integración | `//go:build integration`, SQLite en `t.TempDir()`, `httptest.Server` como fuente | H3 |
| Tests e2e | `github.com/rogpeppe/go-internal/testscript` contra el binario compilado: `kitlegal …`, `boe …` vía symlink, exit codes, `--json`, `--offline` | H1 |
| Matriz territorial | e2e por applet con dimensión territorial: municipio cubierto (Leganés), no cubierto (Tordesillas) y, cuando aplique, foral | H6 |
| Cobertura por materia | e2e o evals con un asunto del procedimiento común y otro de un procedimiento que la ley regula aparte, declarado no cubierto (Definition of Done §1.13) | H9 |
| Tests de contrato | Cada salida validada contra `schemas/*.json` (`santhosh-tekuri/jsonschema`); `--describe` genera el schema desde H1 y `make schema-check` falla si `schemas/` está desactualizado | H4 (primer applet con fuente) |
| Grafo | Idempotencia de `Apply`; ninguna operación sin `source`; `case.db` solo con ids del mundo; ningún verbo de `graph` devuelve texto legal; `Persona` rechaza NIF/DNI por regex | H7 / H10 |
| Smoke contra red | `scripts/verify-sources.sh` en CI nightly; abre issue si una fuente cambia | H4 |
| Comprobación mecánica de skills | En `make ci`: frontmatter válido, `SKILL.md` < 300 líneas, `references/` y tabla de comandos idénticas a lo generado | H5 |
| Evals de skills | `evals/<skill>/*.yaml` (pregunta, si debe activar la skill, comandos esperados, citas esperadas por identificador); ~~job semanal con modelo barato~~ job disparado por los cambios que las evals miden, con el modelo del uso real de la skill, repeticiones y umbral, y modelos informativos que se publican sin decidir [enmienda 2026-09-16, ADR 0016: «modelo barato» no era un criterio de calidad —el proyecto usa suscripción, no API de pago por uso— y medía el peor caso; la ejecución semanal sobre la rama principal no medía ningún cambio] [enmienda 2026-09-30, ADR 0031: el modelo del uso real es el id al que resuelve el alias `sonnet` en la versión de Claude Code del job, fijados juntos —hoy `claude-sonnet-5-5` con 2.1.284—; Sonnet 5 sale del job y Haiku 4.5 sigue como límite inferior] [enmienda 2026-10-05, ADR 0037: lo que tiene forma se compara sin modelo; el significado de la respuesta, cuando no la tiene, lo juzga un modelo distinto del que decide —hoy `claude-opus-5-5`—, medido contra casos etiquetados, desde H24] | H5 |
| Vulnerabilidades | `govulncheck` en cada PR | H0 |
| SAST | `gosec` (en golangci) + CodeQL semanal | H0 |
| Secretos | `gitleaks` en pre-commit y CI | H0 |
| Cadena de suministro | `go mod verify`, `go.sum` obligatorio, Dependabot semanal, SBOM (syft) + firma (cosign) + provenance en release | H0 / H19 |
| Dependencias mínimas | `go mod tidy -diff` en CI; una dependencia nueva requiere justificación en la PR | H0 |
| Observabilidad local | `log/slog` a stderr, `--verbose`, `KITLEGAL_LOG=debug`; `--dry-run` imprime la petición sin ejecutarla | H1 |
| Pre-commit | `lefthook`: `gofumpt`, `golangci-lint run --fast`, `gitleaks`, `go mod tidy -diff` | H0 |
| Commits / versiones | Conventional Commits, SemVer, `CHANGELOG.md` a mano y notas de cada release generadas por goreleaser | H0 / H19 |
| Docs de decisiones | `docs/ADR/NNNN-*.md` (formato MADR corto) | H0 |

Dependencias fijadas (fuera de esta lista, justificar): `alecthomas/kong`, `modernc.org/sqlite`, `stretchr/testify`, `rogpeppe/go-internal` (testscript), `golang.org/x/time/rate`, `temoto/robotstxt`, `gopkg.in/yaml.v3`, `invopop/jsonschema` (generación), `santhosh-tekuri/jsonschema` (validación en tests). En el hito de distribución (backlog): `modelcontextprotocol/go-sdk`.

---

## 4. Fases y hitos

Cada hito vive en su fichero, en `docs/roadmap/` (`docs/roadmap/README.md`), y el directorio en el que está dice su estado:

- `done/` — los terminados, con el texto con que se lanzaron y sus enmiendas fechadas: `H9-…md`.
- `in-progress/` — el que se implementa ahora, como mucho uno.
- `todo/` — los que vienen, en orden: el prefijo numérico es la posición (`010-H29-…md`, `020-H30-…md`). Cambiar el orden es renombrar un fichero, con los huecos de diez en diez para no tocar los demás; el identificador (`H29`) no cambia nunca, porque lo citan los `specs/`, los ADR, el CHANGELOG, los commits y la serie de cierres.
- `backlog/` — los candidatos, sin número ni especificación: las materias, las verticales, y los grupos de distribución, territorios, fuentes, profundidad del BOE, medida de las skills, transparencia y fiscalización, y fiscalizar y vigilar. Un candidato recibe su número cuando pasa a `todo/` (§6).

`make roadmap` enseña el estado. Cada fichero lleva una cabecera con su `id`, su `fase` y, si los tiene, los hitos de los que `depende_de`, y después `# H<n> · título` y la entrada del hito: objetivo, entrega, alcance, controles y aceptación. La lista de hitos que admite el workflow se genera desde este directorio (`make roadmap-sync`) y `make ci` falla si no coincide. Lo que sigue es lo que gana cada fase al terminar; sus hitos son los ficheros con su `fase`.

### Fase 0 — Fundación: la primera skill con su herramienta (hecha; `done/`, fase `0`)

Al terminar: la skill `boe-legislacion` responde en Claude Code a consultas sobre cualquier norma consolidada del BOE con el binario Go, medida con evals; `kitlegal boe articulo BOE-A-2015-10565 a21` funciona con caché y `make install` deja binario y skill en su sitio. No hay release todavía: estaba al cierre de la fase 3 (H19), o antes si alguien distinto de quien desarrolla tenía que instalarlo (ADR 0013), y eso ocurrió al cerrar H6: H19 va ahora en la fase 1 (ADR 0019). La skill `boe-fiscal` (Python, fuera de este repo) sigue funcionando como hasta ahora y se migra como primera vertical (backlog) cuando se quiera, sin bloquear nada.

### Terreno y memoria: territorio, grafo del mundo, redacción a una fecha, citas y plazos (hecho hasta H26.1; `done/`, fase `terreno-y-memoria`)

_Era la fase 1 hasta el 2026-10-21 («Punto de partida y rumbo»). Está hecha hasta H26.1; de lo que quedaba, H8 y H28 pasaron a la fase 2 y H9.1 y H27 a la fase 3. El «Al terminar» que sigue es el de entonces._

Al terminar: `legal-core` y `cita-verificada` funcionan para cualquier municipio de España. Una cita en lenguaje natural se resuelve al texto vigente, o al vigente en la fecha que se pida, con URL ELI; un texto entero —un escrito, un dictamen, la respuesta de otro asistente— se revisa cita a cita a su fecha; y el plazo para recurrir un acto de cualquier administración se calcula con los festivos nacionales, autonómicos y locales del municipio del interesado y de la sede del órgano cuando el territorio está cubierto (Comunidad de Madrid primero), avisando explícitamente cuando no lo está, y declarando no cubierto el procedimiento que la ley regula aparte (ADR 0027). Y el binario recuerda en `~/.cache/kitlegal/world.db` cada norma, bloque, versión y órgano que ha visto, con su fuente: `graph check` dice si algo de lo consultado ha cambiado o caducado. Y, con H19 adelantado aquí, alguien distinto de quien desarrolla lo instala con dos órdenes y sin clonar; y, con H21 y H22, quien no usa una terminal lo instala con una extensión y un plugin y lo usa en la app de escritorio de Claude, con todo corriendo en su equipo (ADR 0035). Y, con H9.1, H27 y H28 (ADR 0040), un despacho cuenta un plazo procesal civil o social con el binario y cita la jurisprudencia constitucional y la doctrina que publica el BOE con su texto delante.

### Fase 1 — La base legal genérica: capas, interfaces y skills para cualquier materia (ahora; `todo/`, fase `1`)

Al terminar: se puede preguntar a kitlegal de cualquier materia. La norma que gobierna la pregunta se lee y se cita; lo que la base cubre se responde con sus herramientas; lo que no cubre se declara con el mismo «⚠ NO CUBIERTO» en todas las materias, y el agente lo busca con las herramientas genéricas y lo marca como no verificado, sin dar de memoria ningún plazo, órgano ni vía. Las skills siguen la especificación Agent Skills: un `SKILL.md` corto que guía, y el detalle —también el de cada materia— en referencias que se cargan cuando hacen falta. El código tiene puertos genéricos para las fuentes, el cómputo de plazos, los destinatarios y los documentos, cada uno validado con dos implementaciones. Y una propuesta de cambio mide solo las evals que ha podido cambiar. Lo administrativo de hoy sigue funcionando igual, ya como el primer paquete de materia; lo protegen sus evals. **Aceptación de la fase**: evals de cinco o seis materias, entre ellas el caso de la póliza (`docs/USO.md`, 2026-10-21), en las que la norma se lee y se cita, ningún plazo sale de memoria, la cobertura se declara y lo que falta se busca marcado.

Las entradas de H29 a H38 son de alto nivel: cada una se completa antes de lanzarla (§6). H38 no depende de ninguna y puede adelantarse.

### Fase 2 — Lo administrativo sobre la base: un asunto ante cualquier administración (después; `todo/`, fase `2`)

Al terminar: quien lleva un asunto —un despacho para su cliente o un ciudadano para sí— registra lo que ha presentado o le han notificado ante cualquier administración, y el kit le dice qué vence y le avisa del silencio; una solicitud de acceso a la información, un recurso de alzada o de reposición o unas alegaciones salen listos para firmar, con la fundamentación resuelta por `cita` en la redacción vigente a la fecha del acto, el órgano que indica la notificación (o `territorio`, si es un ayuntamiento) y el plazo por `plazos`; un texto ajeno se revisa cita a cita (H8); las skills leen el boletín de la comunidad y las ordenanzas del municipio (H15 y H16, validados en la Comunidad de Madrid); y la doctrina publicada en el BOE se cita con su texto (H28). El ayuntamiento es un destinatario más. Cubre el procedimiento común (LPAC) y el acceso a la información (LTAIBG) en cualquier materia; lo que la ley regula aparte se declara no cubierto y llega con su vertical (ADR 0027). Todo va sobre la base de la fase 1: lo administrativo es su primer paquete de materia, el asunto es el genérico de H37 y los escritos usan un `docgen` que no sabe de ninguna materia. Es el primer bucle completo —consultar, actuar, seguir—. Los escritos que se apoyen en ordenanzas esperan a H16. Orden: H8, H10, H11, H15, H16 y H28.

### Fase 3 — Los profesionales: despachos y asesorías (después; `todo/`, fase `3`)

Al terminar: un despacho que litiga cuenta sus plazos procesales con el binario y lee al Constitucional desde el BOE, y una asesoría tiene su vertical —fiscal primero— con sus normas, plazos, doctrina y escritos. Son quienes más profundidad piden (ADR 0040; `docs/USO.md`, 2026-10-19). Entran H9.1 y H27 y, en el orden que pida `docs/USO.md`, las verticales del backlog: cada una, un hito de paquete —normas con `vertical:`, referencias, cobertura y evals sobre el molde de H32— y uno por cada hueco de capacidad —una fuente propia con sus condiciones revisadas, plazos propios como datos del motor de H36, doctrina o escritos—. Con la segunda vertical entra el mecanismo de packs (backlog · distribución): una gestoría pide fiscal, laboral y mercantil juntas. Los candidatos de despacho del backlog —el TJUE, la doctrina sectorial, la cartera de asuntos de un despacho, la revisión de escritos ajenos en PDF, un conector de jurisprudencia de pago si un despacho tiene licencia y el ADR de H34 lo permite— entran por el mismo orden.

### Transparencia y fiscalización municipal (detrás de la fase 3; `backlog/transparencia-y-fiscalizacion/`)

Era la fase 3 hasta el 2026-10-21. Consulta lo que hacen las administraciones —qué contratan y qué subvencionan, qué sociedades y administradores aparecen en BORME, cómo van sus presupuestos— y no resuelve la situación de nadie, así que va detrás de la fase 3, con sus números y sus entradas, que esperan en el backlog; lo que el uso pida lo adelanta (ADR 0013). Las fuentes son nacionales y sirven para cualquier municipio; PLACSP, BDNS y BORME recogen a todas las administraciones, y los hitos se validan con el ayuntamiento porque es el órgano cuyo DIR3 da `territorio` (ADR 0027). Todos los applets nacen emitiendo al grafo. H17 (BORME) sirve también a la vertical mercantil y entra con ella si llega antes. El grupo «Fiscalizar y vigilar» del backlog necesita esta sección entera.

---

## 5. Lo que deliberadamente NO se hace (hasta que un hito lo pida)

- Framework de DI, ORM, generador de CLI: Kong + composición manual bastan.
- Extraer `internal/cli` a librería (la regla de los tres applets se cumple en H8; aun así, solo refactor interno, revisado al cerrar H9).
- Applets sin una skill que los use, o capacidades en Go que el modelo resuelve con una referencia generada.
- Casos especiales, YAML o código para un municipio concreto; configuración municipal curada a mano.
- Territorios distintos del de validación (BOCYL, BOP de otras provincias, régimen foral…) hasta que el uso los pida (backlog · territorios).
- Skills verticales (fiscal, laboral, mercantil…) que no sean especialización de una skill base ya existente; ni campo `vertical` en `data/normas.yaml` ni packs por vertical antes de la primera vertical.
- Una materia como fase del roadmap: las materias van al backlog y entran como paquetes sobre los motores de la base (2026-10-21).
- Aplicar los recursos, los plazos o los escritos del procedimiento común a lo que la ley regula aparte (DA 1ª.2 LPAC): se declara no cubierto hasta que llegue su vertical (ADR 0027).
- Anomalías, cruces, vigilancia y exportaciones del grafo antes de que existan las fuentes de la fase 3 (backlog · fiscalizar y vigilar). El grafo de las fases 1 y 2 es memoria y validador, no instrumento.
- `eli:cites` por regex, FTS5, `graph query` con SQL libre, promoción de props a columnas: sin consumidor hasta `boe analisis` o las anomalías.
- Crawler de sedes municipales mientras el boletín baste.
- Repositorio apt propio, winget, `homebrew-core`, junctions en Windows y un `--host` por agente que lea `.agents/skills/` sin más (ADR 0019): entran cuando un usuario los pida.
- MCP remoto con auth, cuotas o caché compartida. Es la única vía hacia ChatGPT y Claude en la web y en el móvil, y la persona lo ha aplazado: sustituiría la promesa del ADR 0027 y pide su propio ADR (ADR 0035).
- Triple store, SPARQL local, embeddings, sincronización entre máquinas.
- Telemetría o cualquier envío de preguntas, asuntos o datos de uso a un servidor de kitlegal, tampoco con consentimiento: lo promete la web y el grafo es local (ADR 0027).
- Paralelismo en descargas: primero correcto y respetuoso, luego rápido.
- Ingesta de CENDOJ, scraping de PETETE/DYCTEA/HJ sin TOS revisados.
- Cobertura como objetivo: el umbral existe para no retroceder, no para perseguirlo.
- Numerar el backlog: un candidato recibe número cuando entra en la fase en curso, no antes.

---

## 6. Ritual por hito (para que "lean" no degenere en "sin control")

1. La entrada del hito es su fichero en `docs/roadmap/todo/`, con objetivo, entrega, alcance, controles y aceptación; se lanza con `scripts/hito.sh H<n>`. Abrir rama `hNN-nombre-corto`; escribir primero las evals de la skill que el hito entrega o mejora (desde H5) y el test e2e (`testscript`) que describe la entrega.
2. Planificar de fuera adentro (qué debe resolver la skill → qué herramientas necesita) e implementar de dentro afuera: `core` → adaptador → applet → skill.
3. `make ci` local; PR con plantilla (objetivo, alcance, controles añadidos, decisiones, pendientes).
4. Revisión: `/code-review` y `/security-review` sobre la PR; si toca fuentes, `docs/SOURCES.md`.
5. Squash-merge; desde H19, la etiqueta y la release son acciones humanas que cualquier hito puede desencadenar.
6. Cambiar el orden es renombrar el prefijo de un fichero de `docs/roadmap/todo/`, y cambiar el alcance, editar su fichero; un hito terminado pasa a `done/` en su propia propuesta de cambio. Este documento solo cambia si cambian las fases o las reglas; el detalle vive en las PR y los ADR.
7. **Bitácora y repriorización.** Quien usa el kit apunta en `docs/USO.md` qué pidió, qué falló y qué faltó. Cada tres o cuatro hitos, antes de abrir el siguiente, se relee la bitácora: el siguiente hito sale de ahí (un fallo, una carencia, una fuente que el uso pidió) o del backlog; ante discrepancia gana la bitácora y este roadmap se actualiza (ADR 0013). Un candidato del backlog que entra pasa a `todo/` con el siguiente número libre, y `make roadmap-sync` lo añade a la lista que admite el workflow.
