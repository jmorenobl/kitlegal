# Cierre de la Definition of Done · H20 (T014)

Sesión del 2026-10-19 sobre la cabeza `0176fd0` (T013) de `024-h20-boe-articulo-fecha`, con `main` en `0b2a21d`, que es
también la base de la rama (`git merge-base HEAD main`): `git diff --name-status main` es lo que traen los 15 commits de
la rama, más los tres ficheros de `gates/` que el workflow tenía modificados al empezar la tarea (`modelos.jsonl`,
`tarea-actual.json`, `tareas-intentos.json`). Todo lo que se afirma aquí se ejecutó y se vio terminar en esta sesión, o
está en el repositorio; lo que no se pudo medir se dice tal cual. Ninguna tarea está en cuarentena
(`gates/cuarentena.json`, `"tareas": {}`), T001 a T013 están en `[X]` y la checklist del spec tiene 16 de 16.

Cómo se ejecutó (desviaciones de entorno respecto al quickstart, ninguna de producto):

- El sandbox de la sesión rechaza en la línea de órdenes las variables (`$TMPDIR`, `$?`, las de un `for`). Los registros
  de las órdenes fueron a `bin/t014/` —dentro del repositorio pero ignorado por git (`git check-ignore -v bin/t014` →
  `.gitignore:2:/bin/`)—, que se borró al terminar; y el árbol aparte de §7 se creó con una ruta literal fuera del
  repositorio, `/private/tmp/claude-501/h20-t014`, en el mismo directorio temporal que usó T012.
- §4 se ejercitó con copias `zz-h20-*.txtar` de los seis guiones de `aceptacion/` en `internal/app/testdata/script/`, que
  se borraron (quedan 0).
- Las sumas por paquete de §2 se hicieron con un `awk` en línea sobre los perfiles (el sandbox no deja `awk -f`); el total
  que da coincide con el de `go tool cover -func` en los dos perfiles.

## 1. Lo creado o modificado, y lo que no cambia

`git diff --name-status main`: 105 entradas, 55 `A`, 50 `M`, 0 `D` (la lista entera, en el anexo; este `cierre.md` es la
106.ª). Por áreas:

| Área | Entradas | Qué son |
|---|---|---|
| `internal/source/boe/` | 2 A + 14 M | la lectura fechada (`fecha.go`), `ValidarFecha`, `ConsultaArticulo.Fecha`, la entrada de las versiones, `resolverElBloque` genérica, lo que entrega al grafo, y sus tests |
| `internal/core/`, `internal/graph/`, `internal/cli/` | 5 M, 3 M, 2 M | `SoloLoNuevo` en `schema.Observado`, en el lote consolidado y en la entrega, y `aplicar` que no pisa lo guardado |
| `internal/app/` | 4 M | `--fecha` en la orden y `fecha` en `boe_articulo`, y sus tests (`boe_test.go`, `grafo_test.go`, `skills_test.go`) |
| `internal/evals/` | 2 A + 15 M | `fechadas.go` (las líneas `⚠ REDACCIÓN …` de la respuesta), el formato con `fecha` y `redaccion_fechada`, el juicio, la caché preparada con `--fecha`, el umbral `redaccion_sin_fechar` y sus tests |
| `schemas/` | 2 M | `bloque.json` (la entrada con `fecha` y la salida `boe.ArticuloFechado`) y `eval.yaml.json` |
| `skills/boe-legislacion/` | 1 M | `SKILL.md` v0.2 (298 líneas) |
| `evals/boe-legislacion/` | 2 A | las evals 22 y 23 |
| Documentos | 4 M | `CHANGELOG.md`, `CONTRIBUTING.md`, `README.md`, `docs/WORKFLOW.md` (T013) |
| `specs/024-h20-boe-articulo-fecha/` | 49 A | los artefactos del hito (30 en `gates/`, 6 guiones en `aceptacion/`) |

**Lo que no cambia** (FR-059, FR-063, FR-070; tasks.md, «Batería»). Una sola orden con todas las rutas de la lista:

```bash
git diff --name-only main -- skills/legal-core/ skills/jurisprudencia/ evals/legal-core/ evals/jurisprudencia/ \
  evals/boe-legislacion/juez/ evals/boe-legislacion/expresiones-prohibidas.yaml .github/ scripts/ .specify/ \
  evidencias/ docs/SOURCES.md docs/ROADMAP.md docs/USO.md docs/HISTORIAL.md docs/ADR/ web/ CLAUDE.md Makefile \
  internal/mcp/ internal/app/herramientas.go internal/cache/ internal/httpx/ internal/evals/juez.go \
  internal/evals/medida.go internal/evals/sondeo.go internal/evals/definicion.go internal/app/testdata/ \
  internal/source/boe/testdata/ internal/evals/testdata/ codecov.yml .golangci.yml go.mod go.sum | wc -l
```

dio **0**. Además:

- **Las veintiuna evals de hoy**: `git diff --name-status main -- evals/` da solo
  `A evals/boe-legislacion/22-lcsp-contrato-menor-en-2019.yaml` y `A evals/boe-legislacion/23-lpac-articulo-21-en-2016.yaml`;
  de la `01` a la `21`, ninguna.
- **Lo grabado y sus manifiestos**: 0 entradas del diff con `testdata` o `grabaciones` en la ruta.
- **Los guiones e2e de hoy**: 0 entradas bajo `internal/app/testdata/` (en la orden de arriba).
- **Los specs de otros hitos**: `git diff --name-only main -- specs/ | grep -vc '^specs/024-h20-boe-articulo-fecha/'` → 0.
- **La suite congelada**: `git log -- specs/024-h20-boe-articulo-fecha/aceptacion/` solo nombra `8bdd3b9 feat(H20): T001`,
  el commit de `gates/aceptacion-congelada.json`.

**Ninguna petición nueva en la lectura fechada** (FR-006):

- En el código, las dos lecturas piden por el mismo camino: `articuloDelBloque` (`internal/source/boe/articulo.go:228`)
  y `versionesDeLaLecturaFechada` (`internal/source/boe/fecha.go:79`) llaman a `resolverElBloque`, que pide el bloque
  con `pedirElBloque` y los metadatos de la norma con `resolverMetadatos` (`articulo.go:267` y `:272`), y nada más. La
  lectura fechada cambia la clave de la entrada (`claveDeLasVersiones`) y lo que se compone, no lo que se pide.
- Ejecutado en esta sesión, con `--dry-run` sobre un bloque que la caché local no tenía:

  ```text
  $ go run ./cmd/kitlegal boe articulo BOE-A-1985-5392 a20 --dry-run --json
  --dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1985-5392/texto/bloque/a20
  --dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1985-5392/metadatos
  $ go run ./cmd/kitlegal boe articulo BOE-A-1985-5392 a20 --fecha 2019-06-01 --dry-run --json
  --dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1985-5392/texto/bloque/a20
  --dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1985-5392/metadatos
  ```

  (Con `a22`, la lectura sin fecha no enseñó ninguna petición porque la caché local de la máquina ya tenía su entrada
  de hoy; la fechada, las mismas dos de arriba con `a22`.)
- En `make ci`, `fecha-codigos.txtar` exige esas dos peticiones con `--dry-run` y la caché vacía, y los guiones
  `fecha-version`, `fecha-herramienta`, `fecha-main` y `fecha-offline` leen sobre lo grabado de H4 en `reproduccion/`
  sin red, de modo que una petición no grabada los dejaría en rojo; los seis, en verde en §3.4. El diff no toca
  `docs/SOURCES.md` ni ninguna grabación.

## 2. Cobertura (Definition of Done §1.9)

Sobre los perfiles que dejó el `make ci` de esta sesión (19:15:29 a 19:20:40). El total, con
`go tool cover -func=<perfil> | tail -1`; por paquete, sentencias cubiertas sobre sentencias, sin contar dos veces un
bloque que aparece en más de un binario de test:

| Perfil | Global (≥ 70 %) | `internal/core/**` (≥ 85 %) | `internal/source/boe` | `internal/graph` | `internal/evals` |
|---|---|---|---|---|---|
| `coverage.out` | **96,7 %** (12 263 / 12 680) | **98,5 %** (2 657 / 2 697) | 99,1 % (806 / 813) | 88,0 % (456 / 518) | 97,4 % (4 059 / 4 168) |
| `coverage-integration.out` | **97,1 %** (12 309 / 12 680) | **98,5 %** (2 657 / 2 697) | 99,1 % (806 / 813) | 88,4 % (458 / 518) | 97,8 % (4 076 / 4 168) |

`internal/core/**` por paquete, igual en los dos perfiles: `cita` 100 %, `grafo` 92,5 % (308 / 333), `ids` 100 %,
`instalacion` 99,7 %, `plazos` 99,3 %, `schema` 89,6 % (69 / 77), `territorio` 100 %.

**Ninguna cifra queda por debajo de su umbral**, así que no hay hueco que pase a la revisión final por esta vía. Para su
lectura, las líneas del código que el hito escribe o toca y que ningún test de su paquete recorre (bloques con
`count` 0 en los dos perfiles):

- `internal/source/boe/articulo.go:53` — el `return` del error de `ValidarFecha` dentro de `articulo`. La validación la
  prueba `TestValidarFecha` directamente y, por el binario, `fecha-codigos.txtar` (FR-005), pero ningún test de
  `internal/source/boe` llama a la fuente con una fecha mal escrita.
- `internal/source/boe/fecha.go:131-133` — el error `inesperado` de una fecha que `ValidarFecha` ya aceptó y
  `time.Parse` no lee: un estado sin vía real, que cubre la regla genérica (tasks.md, «Proporcionalidad»).
- `internal/evals/informe.go:1723` — en `llevaLasLineasFechadas`, la rama de una eval que espera la redacción a una
  fecha **sin** redacción posterior (`esperada.Posterior == nil`). El formato la admite, pero ninguna eval del
  repositorio (la 22 lleva `posterior:` y la 23 `vigente_desde:`) ni ninguna sesión sintética la ejercita.
- `internal/source/boe/articulo.go:278` y `:289-290`, e `internal/graph/aplicar.go:278-279` y `:290-291` — ramas de
  error que ya existían (fallo de la caché bajo `--dry-run`, fallo al guardar, fallo de escritura de SQLite); el hito
  solo cambió la línea vecina (el tipo genérico `[T]` en `resolverElBloque`, el `SoloLoNuevo` en
  `aplicarConsolidado`).

## 3. Quickstart §0 a §10, orden a orden

### 3.0 Precondiciones (§0)

`go version` → `go version go1.27.2 darwin/arm64` (`go.mod`: `go 1.27.0`, `toolchain go1.27.2`).
`git rev-parse --abbrev-ref HEAD` → `024-h20-boe-articulo-fecha`. Lo esperado.

### 3.1 Lo que ya vale hoy (§1)

- `go test -count=1 -run '^TestEntregaDelHito$/^boe-codigos$' ./internal/app/` → `--- PASS: TestEntregaDelHito/boe-codigos`,
  `ok … internal/app 2.583s`. Esperado: `ok`.
- `make skills-check` → `ok` en `internal/app`, `internal/skills` e `internal/evals`. Esperado: `ok`.

### 3.2 La lectura fechada en el adaptador (§2)

`go test -count=1 -v -run '^(TestValidarFecha|TestLeerBloque|TestArticuloEnUnaFecha|TestArticuloEnUnaFechaYLaCache|TestFallosNoSeGuardan|TestObservadoDeBoe|TestGolden|TestArticulo)$' ./internal/source/boe/`
→ los ocho `--- PASS`, 153 subtests en `PASS`, 0 `FAIL`, `ok … internal/source/boe 0.477s`. Esperado: `ok`, con los siete
casos de contracts/articulo-fecha.md §2 —`US1`, `intermedia`, `el_día_que_entra_en_vigor`,
`la_víspera,_ya_publicada_la_de_ese_día`, `posterior_a_la_última`, `anterior_a_la_primera_vigencia` y
`entre_publicación_y_vigencia`; los seis de FR-081 y SC-004 son los seis últimos—, los de FR-004
(`un_bloque_que_no_se_puede_situar`), el de los avisos (`los_avisos_son_los_de_la_norma_hoy`) y los golden de hoy: los
13 ficheros de `internal/source/boe/testdata/golden/`, que el diff no toca (FR-012).

### 3.3 El grafo (§3)

- `go test -count=1 -run '^(TestEntregaDelMontador)$' ./internal/cli/` → `--- PASS: TestEntregaDelMontador`, `ok`.
- `go test -count=1 -run '^(TestApplySoloLoNuevo|TestIntegracionLecturaFechada)$' ./internal/graph/` → `ok`, pero
  **solo corre `TestApplySoloLoNuevo`**: `TestIntegracionLecturaFechada` está en `internal/graph/integracion_test.go`,
  con `//go:build integration` en su línea 1, y la orden del quickstart no lleva `-tags=integration`. Con la etiqueta,
  `go test -count=1 -tags=integration -v -run '^(TestApplySoloLoNuevo|TestIntegracionLecturaFechada)$' ./internal/graph/`
  → los dos `--- PASS`, `ok`; y `make ci` lo corre en `test-integration`. Es una imprecisión de la orden del quickstart,
  no del producto.
- `go test -count=1 -run '^(TestLecturaFechadaEnElGrafo|TestLaSalidaDeBoeNoCambiaConElGrafo|TestLaEntregaLlevaLaProcedenciaDelSobre)$' ./internal/app/`
  → los tres `--- PASS`, `ok`.

Esperado: `ok` en las tres (FR-030 a FR-032, FR-086).

### 3.4 La suite de aceptación (§4)

- `go test -count=1 -run '^TestEntregaDelHito$/^h20-' ./internal/app/` → `ok … [no tests to run]`: los guiones aún no
  están en `internal/app/testdata/script/` (0 ficheros `h20-*`); el workflow los activa tras el bucle. Lo esperado antes
  de la activación.
- Con las seis copias `zz-h20-fecha-{codigos,grafo,herramienta,main,offline,version}.txtar` de `aceptacion/`,
  `go test -count=1 -v -run '^TestEntregaDelHito$/^zz-' ./internal/app/` → los seis `--- PASS`
  (`zz-h20-fecha-herramienta` 0,88 s, `-offline` 0,89 s, `-version` 0,91 s, `-main` 0,91 s, `-codigos` 0,97 s,
  `-grafo` 1,52 s), `ok … internal/app 3.422s`. Las copias se borraron y `git status` volvió a lo de antes. Esperado
  tras el hito: los seis en verde.

### 3.5 Esquemas, `--describe` y la skill (§5)

- `make schema-check` → `ok` (`TestEsquemasPublicados`).
- `make skills-check` → `ok` en los tres paquetes; `wc -l skills/boe-legislacion/SKILL.md` → 298 (< 300, FR-048).
- `go test -count=1 -run '^(TestHerramientasDelServidor|TestSalidaDeBoeContraSchemas)$' ./internal/app/` → los dos
  `--- PASS`, `ok`.
- `go run ./cmd/kitlegal boe articulo BOE-A-2015-10565 a21 --describe` → 4 933 bytes en stdout, 0 en stderr: la entrada
  con la propiedad `"fecha": {"type": "string"}` y `"x-banderas": ["fecha"]`, y en `$defs` `boe.ArticuloFechado`,
  `boe.RedaccionPosterior` y `boe.Aviso`, con `fecha_pedida` y `redaccion_posterior`. `git status`, igual que antes.
  Esperado tras el hito: eso, y `--describe` sin abrir la caché ni pedir nada.

### 3.6 Las evals y el umbral (§6)

- `go test -count=1 -run '^(TestExtraerRedaccionesFechadas|TestFormatoDeLasRedaccionesFechadas|TestJuzgarLecturasFechadas)$' ./internal/evals/`
  → los tres `--- PASS`, `ok`.
- `go test -count=1 -run '^(TestUmbralesDeBoeLegislacion|TestUmbralesDeLasTresSkills|TestDefinicionDelJob|TestPlanEnDosModos|TestConsultasNecesarias|TestConjuntoDeEvals)$' ./internal/evals/`
  → los seis `--- PASS`, `ok`.
- `go test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/` → `--- PASS`, `ok`, con las 23 evals de
  `evals/boe-legislacion/` (`01` a `23`).

Lo esperado, en el árbol: los quince umbrales de `boe-legislacion` con once que deciden
(`internal/evals/informe_test.go:6638`, `deciden: 11`, contra `umbralesDeBoeLegislacion()` de `:7334`), sobre 60
respuestas por modo y 6 con fecha (`:6964-6965`); el peor caso del job, `18977 s` (`definicion_test.go:1877` y `:1960`),
bajo el tope de 352 minutos (`definicion_test.go:1057`; `.github/workflows/evals.yml:226`, `timeout-minutes: 352`).
Ninguno abrió una sesión con modelo.

### 3.7 Los mutantes (§7)

Repetido en esta sesión, sobre la cabeza `0176fd0` (desde la de T012, `693e6a6`, `git diff --stat 693e6a6 HEAD --
internal/ data/ skills/ evals/ schemas/ cmd/` está vacío: solo cambiaron documentos y `specs/`):

- `git worktree add --detach /private/tmp/claude-501/h20-t014 HEAD` → `ok`, en `0176fd0`.
- Sin mutantes, las cuatro órdenes de §7 con `go -C`: `ok` en las cuatro.
- Cada mutante, con la herramienta de edición en el árbol aparte, comprobado con `git -C … diff -U0` (una línea) y
  revertido con `git -C … checkout -- <fichero>` (`git diff --stat` vacío) antes del siguiente:

| Mutante (línea) | Tests en rojo (hojas en rojo) | Órdenes con `FAIL` |
|---|---|---|
| 1, `fecha.go:190`, la vigencia leída de `FechaVersion` | `TestArticuloEnUnaFecha` (7: `anterior_a_la_primera_vigencia`, `entre_publicación_y_vigencia`, `la_víspera,_ya_publicada_la_de_ese_día` y cuatro de `un_bloque_que_no_se_puede_situar`), `TestObservadoDeBoe` (`articulo-con-fecha-sin-redaccion-vigente`) | 2 de 4 |
| 2, `fecha.go:154`, siempre la última | `TestArticuloEnUnaFecha` (`US1`, `el_día_que_entra_en_vigor`, `intermedia`, `la_víspera,_ya_publicada_la_de_ese_día`), `TestObservadoDeBoe` (`articulo-con-fecha`), `TestLecturaFechadaEnElGrafo` (las dos de «una versión nueva») | 3 de 4 |
| 3, `fecha.go:141`, `elegida := 0` | `TestArticuloEnUnaFecha` (`anterior_a_la_primera_vigencia`, `entre_publicación_y_vigencia`), `TestObservadoDeBoe` (`articulo-con-fecha-sin-redaccion-vigente`) | 2 de 4 |
| 4, `grafo.go:87`, las operaciones de la lectura sin fecha | `TestObservadoDeBoe` (`articulo-con-fecha`), `TestLecturaFechadaEnElGrafo` (las dos de «una versión nueva») | 2 de 4 |
| 5, `umbrales.go:668`, `Decide: false` | `TestUmbralesDeBoeLegislacion` (13 subtests), `TestUmbralesDeLasTresSkills` (`boe-legislacion`) | la de `internal/evals`, la única que se corrió con él: `go list -test -deps ./internal/app/ ./internal/source/boe/` no nombra `internal/evals`, así que las otras tres no lo compilan |

- `git worktree remove --force /private/tmp/claude-501/h20-t014` → `git worktree list` solo nombra el árbol de trabajo
  y `wt-marcas` (de otra rama, que no es de esta tarea), y el directorio ya no existe.

Esperado: con cada mutante, el test de su fila falla; sin mutantes, todos pasan. **5 de 5**, y los mismos tests y
recuentos de hojas que registra `mutantes.md` (T012).

### 3.8 Todo (§8)

`make ci` en primer plano (con `rtk proxy`, el registro en `bin/t014/`) → última línea `ci: todos los controles en
verde`, 61 líneas `ok` de `go test`, `No vulnerabilities found.`, `goreleaser check`, `gitleaks`, `go mod verify`
(`all modules verified`) y `go mod tidy -diff` limpios; 5 min 11 s. Esperado: verde, con `schema-check`, el drift de
`references/` y las reglas del conjunto (FR-080). La ejecución final, tras escribir este fichero, en «Verificación».

### 3.9 y 3.10

- §9 (el cierre: empujar la rama, abrir la propuesta de cambio y medir el job de evals sobre la cabeza, con los dos
  `redaccion_sin_fechar:claude-sonnet-5-5:<modo>` en 0 de 6 y `decide: true`) **lo hace el workflow** tras la revisión
  final; no es de esta tarea y no se ha ejecutado aquí (SC-001, SC-002).
- §10 (la pregunta del art. 118 LCSP el 1 de junio de 2019 en la app de escritorio con el plugin) **queda para una
  persona, después del run**, con la release publicada (SC-003).

## 4. Controles de umbral (plan.md)

Filas `ci:` — cada test encontrado como `^func <Test>\(` en su ruta (línea entre paréntesis) y en verde en esta sesión
(§3.2 a §3.6 y `make ci`); los guiones, como copias `zz-` en §3.4:

| Requisito | Control | Encontrado | Verde |
|---|---|---|---|
| FR-002, FR-003, FR-013, FR-081, SC-004 | `internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha` (54), `h20-fecha-version.txtar` | sí | sí |
| FR-004, FR-011 | `internal/source/boe/fecha_test.go:TestArticuloEnUnaFecha` (54) | sí | sí |
| FR-005, FR-007, FR-082 | `internal/source/boe/ids_test.go:TestValidarFecha` (167), `fuente_test.go:TestFallosNoSeGuardan` (833), `h20-fecha-codigos.txtar` | sí | sí |
| FR-012, FR-083, SC-005 | `h20-fecha-main.txtar`, `internal/source/boe/fuente_test.go:TestGolden` (1214) | sí | sí |
| FR-020 a FR-022, FR-082, FR-087, SC-007 | `internal/source/boe/fecha_test.go:TestArticuloEnUnaFechaYLaCache` (280), `h20-fecha-offline.txtar` | sí | sí |
| FR-030 a FR-032, FR-086, SC-006 | `internal/graph/almacen_test.go:TestApplySoloLoNuevo` (559), `internal/graph/integracion_test.go:TestIntegracionLecturaFechada` (1550, etiqueta `integration`), `internal/app/grafo_test.go:TestLecturaFechadaEnElGrafo` (2301), `h20-fecha-grafo.txtar` | sí | sí |
| FR-001, FR-084 | `h20-fecha-herramienta.txtar`, `internal/app/herramientas_test.go:TestHerramientasDelServidor` (138) | sí | sí |
| FR-014, FR-085 | `Makefile:schema-check` (122), `internal/app/boe_test.go:TestSalidaDeBoeContraSchemas` (1229) | sí | sí |
| FR-048, SC-011 | `Makefile:skills-check` (126), `internal/evals/conjunto_test.go:TestEvalsDelRepositorio` (1993) | sí | sí |
| FR-050 | `internal/evals/definicion_test.go:TestDefinicionDelJob` (196) | sí | sí |
| FR-051, FR-052 | `internal/evals/formato_test.go:TestFormatoDeLasRedaccionesFechadas` (1630), `juzgar_test.go:TestJuzgarLecturasFechadas` (3654), `conjunto_test.go:TestEvalsDelRepositorio` (1993) | sí | sí |
| FR-053 | `internal/evals/conjunto_test.go:TestEvalsDelRepositorio` (1993) | sí | sí |
| FR-055, FR-089, SC-010 | `internal/evals/informe_test.go:TestUmbralesDeBoeLegislacion` (7056), `juzgar_test.go:TestJuzgarLecturasFechadas` (3654), `informe_test.go:TestUmbralesDeLasTresSkills` (6591) | sí | sí |
| FR-056, SC-009 | `internal/evals/informe_test.go:TestUmbralesDeLasTresSkills` (6591) | sí | sí |
| FR-090, SC-008 | `TestArticuloEnUnaFecha` (54), `internal/source/boe/grafo_test.go:TestObservadoDeBoe` (56), `TestLecturaFechadaEnElGrafo` (2301), `TestUmbralesDeBoeLegislacion` (7056), `TestUmbralesDeLasTresSkills` (6591) | sí | sí, y 5 de 5 mutantes en rojo (§3.7) |
| FR-080 | `Makefile:ci` (264) | sí | sí |

Filas `evals:` — no se miden aquí (necesitan el job con modelo, §9); lo que hay en el árbol:

| Umbral | Origen | En el árbol |
|---|---|---|
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` y `:herramienta` | **nuevos, de T010** | `umbralDeRedaccionSinFechar` (`internal/evals/umbrales.go:659`): `Umbral: umbralDeRedaccionesSinFechar` (`= 0`, `umbrales.go:42`), `Decide: true` (`:668`), sobre las respuestas de las evals con fecha de su modo (6 en el plan del job, `informe_test.go:6646-6648`); cuarto y octavo de los quince, detrás de `cuenta_su_proceso` de su modo (contracts/evals-redaccion-fechada.md §3) |
| `sin_activar:claude-sonnet-5-5:orden` y `:herramienta` | el de hoy, que sigue | 0 de 60 por modo (`informe_test.go:6645` y `:6647`) |
| `afirma_lo_no_leido:claude-sonnet-5-5` | el de hoy, que sigue | tasa `0.028` (`evals/boe-legislacion/juez/clases.yaml:9`, sin cambios), sobre 120 respuestas (`informe_test.go:6649`) |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` y `:correctos_marcados` | el de hoy, con la medida versionada | `evals/boe-legislacion/juez/` sin cambios |
| `duracion_de_las_sesiones:<modo>`, `duracion_del_juez:<modo>` | el de hoy | sin cambios en `internal/evals/definicion.go` ni en `.github/` |
| Los once de `jurisprudencia` y los seis de `legal-core` | los de hoy, que no cambian | `TestUmbralesDeLasTresSkills`: `jurisprudencia` con 11 que deciden (`:6653`) y `legal-core` con sus seis nombres, todos `decide` (`:6615-6622`); `evals/` y `skills/` de las dos, sin cambios (§1) |

## 5. Nota para el informe final (FR-092)

No es trabajo de ninguna tarea; la lleva el informe final: **la fase 1 sigue con H9.1 (los plazos procesales civiles y
sociales; ADR 0040), y H8 y H11 pueden dar por hecha la redacción de un bloque a una fecha** (`boe articulo --fecha` y el
parámetro `fecha` de `boe_articulo`).

## Verificación de la tarea

- `make ci` en primer plano, en verde las dos veces: la de 19:15:29 a 19:20:40 (§3.8) y la final, con este fichero ya
  escrito, de 19:30:45 a 19:35:46 (`ci: todos los controles en verde`, 61 líneas `ok`, ninguna `FAIL`). Sus perfiles
  dan las mismas cifras de §2 (96,7 % y 97,1 % de total; 2 657 / 2 697 en `internal/core/**`; 806 / 813, 458 / 518 y
  4 076 / 4 168 en el de integración).
- `git status --short`: solo `specs/024-h20-boe-articulo-fecha/cierre.md` nuevo, la marca de T014 en `tasks.md` y los
  tres ficheros de `gates/` que el workflow tenía modificados al empezar. `bin/t014/` (ignorado) se borró; ninguna copia
  `zz-` queda en `internal/app/testdata/script/`; ningún árbol aparte queda en `git worktree list`.
- Esta tarea no escribe tests ni código de producto, no usa la red (salvo la de las herramientas de Go en `make vuln`),
  no graba nada, no escribe en `evidencias/` y no abre ninguna sesión con modelo.

## Anexo: `git diff --name-status main`

```text
M	CHANGELOG.md
M	CONTRIBUTING.md
M	README.md
M	docs/WORKFLOW.md
A	evals/boe-legislacion/22-lcsp-contrato-menor-en-2019.yaml
A	evals/boe-legislacion/23-lpac-articulo-21-en-2016.yaml
M	internal/app/boe.go
M	internal/app/boe_test.go
M	internal/app/grafo_test.go
M	internal/app/skills_test.go
M	internal/cli/entrega.go
M	internal/cli/entrega_test.go
M	internal/core/grafo/lote.go
M	internal/core/grafo/lote_test.go
M	internal/core/graphstore.go
M	internal/core/schema/grafo.go
M	internal/core/schema/grafo_test.go
M	internal/evals/conjunto.go
M	internal/evals/conjunto_test.go
M	internal/evals/consultas.go
M	internal/evals/consultas_test.go
M	internal/evals/definicion_test.go
M	internal/evals/doc.go
A	internal/evals/fechadas.go
A	internal/evals/fechadas_test.go
M	internal/evals/formato.go
M	internal/evals/formato_test.go
M	internal/evals/informe.go
M	internal/evals/informe_test.go
M	internal/evals/juzgar.go
M	internal/evals/juzgar_test.go
M	internal/evals/plan_test.go
M	internal/evals/preparar_test.go
M	internal/evals/umbrales.go
M	internal/graph/almacen_test.go
M	internal/graph/aplicar.go
M	internal/graph/integracion_test.go
M	internal/source/boe/articulo.go
M	internal/source/boe/bloque.go
M	internal/source/boe/bloque_test.go
M	internal/source/boe/consultas.go
M	internal/source/boe/datos.go
M	internal/source/boe/datos_test.go
M	internal/source/boe/doc.go
M	internal/source/boe/entradas.go
M	internal/source/boe/entradas_test.go
A	internal/source/boe/fecha.go
A	internal/source/boe/fecha_test.go
M	internal/source/boe/fuente_test.go
M	internal/source/boe/grafo.go
M	internal/source/boe/grafo_test.go
M	internal/source/boe/ids.go
M	internal/source/boe/ids_test.go
M	schemas/bloque.json
M	schemas/eval.yaml.json
M	skills/boe-legislacion/SKILL.md
A	specs/024-h20-boe-articulo-fecha/aceptacion/fecha-codigos.txtar
A	specs/024-h20-boe-articulo-fecha/aceptacion/fecha-grafo.txtar
A	specs/024-h20-boe-articulo-fecha/aceptacion/fecha-herramienta.txtar
A	specs/024-h20-boe-articulo-fecha/aceptacion/fecha-main.txtar
A	specs/024-h20-boe-articulo-fecha/aceptacion/fecha-offline.txtar
A	specs/024-h20-boe-articulo-fecha/aceptacion/fecha-version.txtar
A	specs/024-h20-boe-articulo-fecha/checklists/requirements.md
A	specs/024-h20-boe-articulo-fecha/contracts/SKILL-boe-legislacion.prototipo.md
A	specs/024-h20-boe-articulo-fecha/contracts/articulo-fecha.md
A	specs/024-h20-boe-articulo-fecha/contracts/documentos.md
A	specs/024-h20-boe-articulo-fecha/contracts/evals-redaccion-fechada.md
A	specs/024-h20-boe-articulo-fecha/contracts/skill-boe-legislacion.md
A	specs/024-h20-boe-articulo-fecha/data-model.md
A	specs/024-h20-boe-articulo-fecha/gates/aceptacion-congelada.json
A	specs/024-h20-boe-articulo-fecha/gates/analyze.md
A	specs/024-h20-boe-articulo-fecha/gates/clarify-preguntas.json
A	specs/024-h20-boe-articulo-fecha/gates/clarify-respuestas.json
A	specs/024-h20-boe-articulo-fecha/gates/clarify-rondas
A	specs/024-h20-boe-articulo-fecha/gates/cuarentena.json
A	specs/024-h20-boe-articulo-fecha/gates/modelos.jsonl
A	specs/024-h20-boe-articulo-fecha/gates/plan-correccion-r1.diff
A	specs/024-h20-boe-articulo-fecha/gates/plan-precheck.txt
A	specs/024-h20-boe-articulo-fecha/gates/plan-r1.json
A	specs/024-h20-boe-articulo-fecha/gates/plan-r2.json
A	specs/024-h20-boe-articulo-fecha/gates/plan-rondas
A	specs/024-h20-boe-articulo-fecha/gates/plan.json
A	specs/024-h20-boe-articulo-fecha/gates/spec-correccion-r1.diff
A	specs/024-h20-boe-articulo-fecha/gates/spec-correccion-r2.diff
A	specs/024-h20-boe-articulo-fecha/gates/spec-precheck.txt
A	specs/024-h20-boe-articulo-fecha/gates/spec-r1.json
A	specs/024-h20-boe-articulo-fecha/gates/spec-r2.json
A	specs/024-h20-boe-articulo-fecha/gates/spec-r3.json
A	specs/024-h20-boe-articulo-fecha/gates/spec-rondas
A	specs/024-h20-boe-articulo-fecha/gates/spec.json
A	specs/024-h20-boe-articulo-fecha/gates/supuestos.md
A	specs/024-h20-boe-articulo-fecha/gates/tarea-actual.json
A	specs/024-h20-boe-articulo-fecha/gates/tareas-intentos.json
A	specs/024-h20-boe-articulo-fecha/gates/tasks-correccion-r1.diff
A	specs/024-h20-boe-articulo-fecha/gates/tasks-precheck.txt
A	specs/024-h20-boe-articulo-fecha/gates/tasks-r1.json
A	specs/024-h20-boe-articulo-fecha/gates/tasks-r2.json
A	specs/024-h20-boe-articulo-fecha/gates/tasks-rondas
A	specs/024-h20-boe-articulo-fecha/gates/tasks.json
A	specs/024-h20-boe-articulo-fecha/mutantes.md
A	specs/024-h20-boe-articulo-fecha/plan.md
A	specs/024-h20-boe-articulo-fecha/quickstart.md
A	specs/024-h20-boe-articulo-fecha/research.md
A	specs/024-h20-boe-articulo-fecha/spec.md
A	specs/024-h20-boe-articulo-fecha/tasks.md
```
