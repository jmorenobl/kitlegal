# Research: H23.1 · `jurisprudencia` coteja el documento antes de decir nada de él, y un umbral sin modelo lo hace cumplir

**Modo**: desatendido. Cada decisión técnica se tomó con el «Criterio de decisión autónoma» de la constitución y queda
abajo (D1-D16) con su alternativa rechazada. Lo que se afirma del repositorio, con fichero y línea o con la orden que lo
dio, está en la tabla V. Lo medido en esta sesión está en la tabla M, sobre un prototipo fuera del repositorio: una copia
de la cabeza `ef35b40` sacada con `git archive HEAD` a `$TMPDIR/h231-proto`, con el cambio de producción y el de los
tests de este plan. Queda ahí, fuera del repositorio: esta sesión no tuvo permiso para borrarlo, y ninguna tarea parte
de él ni lo necesita. Lo que no se pudo comprobar son los supuestos S1-S4.

Ningún `NEEDS CLARIFICATION` queda abierto: el Technical Context del plan no tiene ninguno.

## V · Comprobado en el repositorio

| Id | Qué | Dónde |
|---|---|---|
| V1 | Go 1.27: `go version` da `go1.27.2 darwin/arm64`; `go.mod` fija `go 1.27.0` y `toolchain go1.27.2` | `go version`; `go.mod:3,5` |
| V2 | El cambio solo usa paquetes que los dos ficheros ya importan: `informe.go` usa `slices` y `strings` (p. ej. `informe.go:1327`, `:1365`), y `umbrales.go` importa `fmt`, `strconv` y `strings` (`umbrales.go:3-7`). Ninguna dependencia nueva | los dos ficheros; el prototipo compila sin tocar ningún `import` (M2) |
| V3 | El criterio con el que el juicio da por ejecutado un comando esperado de `cita`: `satisface` exige `consultoConExito` —`Consulta` y código 0— y el mismo applet (`juzgar.go:968-991`, `:999-1003`); en la forma cotejar llama a `pideLaCita`, que con `ROJ` vacío y `ConTexto` falso solo mira el verbo (`sentencias.go:489-493`). Así, `satisface(inv, ComandoEsperado{Applet: "cita", Verbo: "cotejar"})` es exactamente el cotejo de FR-002: `cita cotejar` con cualquier argumento, que consultó y terminó con 0 | `internal/evals/juzgar.go`, `internal/evals/sentencias.go` |
| V4 | `Consulta` es falso si la orden pide la ayuda o lleva `--describe` o `--dry-run` (`trazas.go:253-256`). Una llamada a una herramienta es siempre una consulta, y su código es 0 si su resultado no es un error, el de su clase si lo es y ninguno si no tiene resultado (`invocacionDeLaLlamada` y `codigoDeLaLlamada`, `juzgar.go:539-589`); su `Error` es `is_error` verdadero o un sobre con `ok` falso (`sesion.go:199-202`) | `internal/evals/trazas.go`, `juzgar.go`, `sesion.go` |
| V5 | `invocacionesDe` da las invocaciones con las que se juzga una sesión: las de su traza y, detrás, las de sus llamadas a las herramientas del servidor (`juzgar.go:493-516`). Devuelve falso solo si `verbosDeLasHerramientas` falla; es un `sync.OnceValues` (`juzgar.go:622`), y `LeerSesion` lo llama antes de leer nada y deja ilegible la sesión si falla (`sesion.go:271-274`): una sesión leída nunca da falso | `internal/evals/juzgar.go`, `sesion.go` |
| V6 | Las respuestas de cada modo, las del `total` de `sin_activar:<modelo>:<modo>`, son las de `respuestasQueSeJuzgan` (`informe.go:1605-1643`): de las series que pide el plan, del modelo que decide y de su modo, sin las ilegibles, sin las sin medir (`SinMedir`), sin las que el tope cortó (`SesionTerminada` falso) y sin las de una eval que no activa la skill; las de una eval sin binario ni servidor van a un grupo sin modo, que no entra en ningún umbral | `internal/evals/informe.go` |
| V7 | `medirLasRespuestas` (`informe.go:1324-1345`) solo existe para una skill con juez o con alguna eval que declara `sentencias`, y con sentencias mide cada respuesta de los grupos de los modos (`medirLasCitasSinDocumento`, `:1354-1368`): es el patrón que sigue la medida nueva | `internal/evals/informe.go` |
| V8 | Los umbrales de las respuestas de cada modo van en `respuestasMedidas.umbrales` (`umbrales.go:266-279`): `sin_activar`, el recuento de cada clase del juez y, con sentencias, `cita_sin_documento`. El motivo de uno que decide y no se cumple es `motivoDeUnUmbral`, «umbral %s: %s, y tiene que ser %s» (`:100`), seguido, si las nombra, de las respuestas que cuentan (`nombradasEn`, `:561-567`), separadas por «; » (`separadorDeLasMarcadas`, `:103`) | `internal/evals/umbrales.go` |
| V9 | Piden `cita cotejar` la 04, la 06, la 08 y la 09 de `jurisprudencia` (`applet: cita`, `verbo: cotejar`), las cuatro con `activa: true`; la 01 pide `cita preparar`. La 06 pide además `cita preparar` con el mismo ROJ (`06-documento-que-no-es-el-pedido.yaml:48-55`) | `grep -nE "cotejar\|activa:" evals/jurisprudencia/*.yaml` |
| V10 | El sondeo usa `respuestasQueSeJuzgan` (`sondeo.go:396`) y no `medirLasRespuestas` ni `umbralesDelInforme`: el cambio no le llega, y `sondeo.go` no se toca (FR-033) | `grep -nE "medirLasRespuestas\|umbralesDelInforme\|respuestasQueSeJuzgan" internal/evals/*.go` |
| V11 | `informe.json` se escribe con `json.Marshal` de `encoding/json/v2` y sangría (`informe.go:919`), que no escapa `<`: el informe versionado del cierre de H26 lleva `"comparacion":"<="` tal cual (`jq -r '.umbrales[0] \| tojson'`) | `internal/evals/informe.go`; `specs/021-h26-el-umbral-de/gates/evals/jurisprudencia.json` |
| V12 | El informe final da un requisito como «comprobado por su control» si todas las celdas «Dónde» de sus filas lo están: `ci:<ruta>:<Test>` exige que `<ruta>` exista en la cabeza, que tenga `func <Test>(` o `<Test>:` y `make ci` en verde; `evals:<skill>:<nombre>` busca en `umbrales` del informe del job el elemento con ese nombre exacto, y exige que lo publique, que sea coherente, que decida y que se cumpla (`estado_control`, `scripts/workflow/informe.sh:346-369`; uso, `:541-571`) | `scripts/workflow/informe.sh` |
| V13 | `precheck.sh plan` exige `plan.md`, `research.md`, ninguna marca de aclaración, `## Constitution Check`, la línea «Aceptación e2e» y `## Controles de umbral` con cada fila nombrando requisitos del spec y un control con forma (`scripts/workflow/precheck.sh:47-75`; `controles_de_umbral` y `controles_de_celda`, `scripts/workflow/comun.sh:148-166`) | `scripts/workflow/` |
| V14 | El binario no enlaza `internal/evals`: `go list -deps ./cmd/kitlegal \| grep -c 'kitlegal/internal/evals$'` da 0 | la orden |
| V15 | `make skills-check` ejecuta `TestSkillsDelRepositorio`, `TestOrdenesDeLasSkillsEmpotradas`, `TestEvalsDelRepositorio` y otros cuatro (`Makefile:126-127`); `make ci` lleva `fmt-check` (`golangci-lint fmt --diff`), `lint`, `test` (`go test -race -shuffle=on`), `schema-check` y `skills-check` (`Makefile:86-87`, `:102-103`, `:114-115`, `:122-123`, `:264-266`); `golangci-lint` va fijado por `tools/golangci-lint/go.mod` (`Makefile:41`) | `Makefile` |
| V16 | `skills/jurisprudencia/SKILL.md` tiene 197 líneas (`wc -l`), y `TestSkillsDelRepositorio` rechaza 300 (`internal/skills/sincronia_test.go:497-503`). `TestOrdenesDeLasSkillsEmpotradas` solo mira la región generada de la tabla de comandos (`internal/app/skills_test.go:306-321`) | `wc -l`; los tests |
| V17 | Los guiones de `skills install` comparan el `SKILL.md` instalado con el empotrado (`cmp`), sin huella escrita: `internal/app/testdata/script/h19-skills-install-local.txtar:71`, `h19-skills-doctor-copia.txtar:48`, `internal/skills/testdata/script/instalar.txtar:44`. La única copia versionada es `skills/jurisprudencia/SKILL.md` (`git ls-files`) | `git grep -nE "jurisprudencia/SKILL\.md"` |
| V18 | `TestInformeFinalDeLosUmbrales/como-en-el-cierre-de-h26` lee los umbrales de `jurisprudencia` de `umbralesDeJurisprudencia(nil)` y espera «jurisprudencia: 13 umbrales, los 9 que deciden se cumplen…» (`internal/evals/informe_test.go:1845`, `:1853`) | `internal/evals/informe_test.go` |
| V19 | El arnés de `TestUmbralesDeJurisprudencia` (`informe_test.go:5091-5147`): casos en `casosDeLosUmbralesDeJurisprudencia` (`:5150-5366`), `exigirLosUmbralesDeJurisprudencia` exige trece y nueve (`:5401-5404`), `umbralesDeJurisprudencia` (`:5527-5553`), 30 respuestas por modo (`:4972`), sesiones de `escribirSesionDeJurisprudencia` (`:5689-5751`) con las sesiones modelo de `sesionesModeloDeLasDiezEvals` (`sentencias_test.go:741`); la de la 06 coteja con el golden `cotejar-roj-cruzado` y prepara (`sentencias_test.go:822-838`), y su respuesta no lleva ningún ECLI ni ninguna cita (`:764-767`) | `internal/evals/informe_test.go`, `sentencias_test.go` |
| V20 | El sobre de `cita cotejar` del caso de la 06 está versionado: `internal/app/testdata/cita/cotejar-roj-cruzado.json`, 937 bytes en una línea, con `es_la_pedida: false` y un hallazgo `documento-distinto` con `cruce: numero-de-resolucion` y la explicación «…1088/2023 es el número de resolución del documento, no su ROJ.» | `wc -c`; `jq -c '.data \| {es_la_pedida, correspondencia, hallazgos}'` |
| V21 | Los comentarios de `internal/evals` que el hito deja incompletos: `doc.go:178-194` («cada una de las dos skills con juez lleva trece, de los que deciden nueve»; «En jurisprudencia… cita_sin_documento»); `umbrales.go:117-126` (nombres de `Umbral.Nombre`), `:163-226` (`umbralesDelInforme`, «trece… nueve»), `:255-265` (`umbrales`), `:517-535` y `:555-560` (motivos); `informe.go:762-772` (paso 4 de `EscribirInforme`) | los ficheros |
| V22 | Los pasajes de los documentos: `CONTRIBUTING.md:1144-1145` («en `jurisprudencia`, otros trece»), `:1172` («son también trece, y deciden nueve»), `:1176-1185` (su tabla), `:1187-1190` (orden; «cuatro de los trece»), `:1194-1201` (`cita_sin_documento`); `docs/JURISPRUDENCIA.md:227` («el job decide con `cita_sin_documento`»); `CLAUDE.md:67` («Lo que comprueba un guion —la cita, las líneas `⚠`, la activación, `cita_sin_documento`, la medida del juez—»); `docs/WORKFLOW.md:139` solo dice «en `boe-legislacion` deciden nueve», que sigue siendo cierto; `CHANGELOG.md`: *Unreleased* en `:13-63`, `## [0.6.0]` en `:65`, la v0.1 de `jurisprudencia` en `:110` y la v0.1.7 de `boe-legislacion`, en «Cambiado», en `:157` | `grep -n`, `sed -n` |
| V23 | `.github/workflows/evals.yml:157` dice «Cada una de las dos skills con juez lleva trece umbrales, nueve que deciden»: tras el hito es falso para `jurisprudencia`, y el fichero queda fuera (FR-052; spec, «Fuera de alcance») | `sed -n 153,159p .github/workflows/evals.yml` |

## M · Medido en esta sesión, sobre el prototipo

Todas las órdenes, sobre `$TMPDIR/h231-proto` (`git archive HEAD` de `ef35b40`) con `go -C`, sin red y sin modelo.

| Id | Qué | Resultado |
|---|---|---|
| M1 | `go test -count=1 -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/` sobre la cabeza sin cambios | `ok`, 4,9 s |
| M2 | Solo el cambio de producción (`informe.go`, `umbrales.go`), con `go test -count=1 ./internal/evals/` | Falla solo `TestUmbralesDeJurisprudencia`: sus diez casos de la tabla, que esperan trece umbrales; `el-elemento-del-contrato` y el resto del paquete pasan. `go vet` sin defectos |
| M3 | Con el arnés y los casos de este plan | Los quince subtests de `TestUmbralesDeJurisprudencia` pasan: los diez de antes, `el-elemento-del-contrato` y los cuatro nuevos |
| M4 | `TestInformeFinalDeLosUmbrales/como-en-el-cierre-de-h26` | Falla si lee los quince de `umbralesDeJurisprudencia(nil)`; pasa sin cambiar lo que espera con `umbralesDeJurisprudenciaComoEnH26()`, los quince sin los dos `documento_sin_cotejar` |
| M5 | `TestEvalsDelRepositorio/piden-el-cotejo` | Pasa: con `pideElCotejo`, la 04, la 06, la 08 y la 09 de `jurisprudencia`, y ninguna de `boe-legislacion` ni de `legal-core` |
| M6 | `go test -race -shuffle=on -count=1 ./internal/evals/` con todo | `ok`, 56 a 61 s |
| M7 | `golangci-lint run ./internal/evals/` y `golangci-lint fmt --diff ./internal/evals/...`, los dos con `-modfile=tools/golangci-lint/go.mod` | `0 issues.` y ninguna diferencia |
| M8 | Los dos mutantes de FR-064, cada uno una línea con `sed -i.orig` (comprobado con `git diff --no-index -U0`), y `go test -count=1 -v -run '^TestUmbralesDeJurisprudencia$'` | Mutante 1 (cuenta solo si ninguna sesión del modo cotejó): fallan `una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta` y `una-sin-activar-y-sin-cotejar`. Mutante 2 (`decide` en falso): fallan los catorce casos de la tabla |
| M9 | Bytes de los elementos, con un test de usar y tirar que codificaba cada `Umbral` con `encoding/json` (v1, que escribe `<` como `<`: 5 bytes más por elemento) | `documento_sin_cotejar:claude-sonnet-5-5:orden`, 372 con 1 de 12; `…:herramienta`, 383 con 0 de 12; los quince, 4 918. En la forma de `informe.json`, sin ese escape (V11): 367, 378 y 4 843. El ejemplo de contracts/umbral-documento-sin-cotejar.md §2, el del modo orden con 0 de 12, mide 366 (`printf '%s' … \| wc -c`) |
| M10 | El motivo, con la sesión 01 de la 06 sin cotejo en el modo orden (el del prototipo, y `printf '%s' … \| wc -c`) | 151 bytes: `umbral documento_sin_cotejar:claude-sonnet-5-5:orden: 1 de 12 (8,3 %), y tiene que ser ≤ 0,0 %: 06-documento-que-no-es-el-pedido-claude-sonnet-5-5-01`. Su cabecera con «12 de 12 (100,0 %)» mide 101 bytes en el modo orden y 107 en el herramienta |
| M11 | Las filas de `informe.md` | 92 bytes la del modo orden con 1 de 12; 99 la del modo herramienta con 0 de 12 |
| M12 | Los nombres de las sesiones de las cuatro evals, con `awk` | Modo orden: 40, 53, 53 y 54 bytes (04, 06, 08 y 09), 600 las doce; modo herramienta: 52, 65, 65 y 66, 744 las doce. El motivo más largo, con las doce y sus once «; »: 101 + 600 + 22 = 723 bytes en el modo orden y 107 + 744 + 22 = 873 en el herramienta |
| M13 | Los trece elementos del cierre de H26: `jq '[.umbrales[] \| tojson \| utf8bytelength] \| add'` sobre `specs/021-h26-el-umbral-de/gates/evals/jurisprudencia.json` | 4 100 bytes. El spec daba 4 108 y su juez, 4 094 caracteres: ninguna de las tres cifras cambia el diseño |
| M14 | El tamaño del cambio de producción del prototipo, con `git diff --no-index --numstat` | `informe.go` +94, `umbrales.go` +71 −7, sin contar los comentarios de cabecera que la tarea corrige (V21) |

## S · Supuestos no verificados

- **S1. Lo que harán las sesiones del job de cierre.** Que con la skill del hito las 24 sesiones de `claude-sonnet-5-5`
  de las evals 04, 06, 08 y 09 coteje cada una su documento no se puede medir sin abrir sesiones con modelo, que este
  paso no abre (ADR 0032). Lo mide el job de cierre (SC-001); con 6 sesiones de la 06 por cierre, el arreglo frente a la
  suerte solo lo distingue el sondeo de la persona (SC-010).
- **S2. Que la traza de una sesión real registre un `cita cotejar` del modo orden como invocación.** Es el
  comportamiento del juicio desde H23 —el comando esperado `cita cotejar` de la 06 ya sale en `comandos_ejecutados`—,
  probado por los tests del juicio (`TestJuzgarSentencias`, `TestJuzgarLasLlamadas`) con trazas sintéticas; no se ha
  vuelto a ver en una sesión real en esta sesión.
- **S3. La línea base de 6 de 48.** Sale de `docs/USO.md` (2026-10-15, línea 22) y del hito; no se ha recalculado aquí.
- **S4. `sed -i.orig` con GNU sed.** Las dos sustituciones de quickstart §4 se han ejecutado con el `sed` de macOS
  (BSD), que es el de la máquina del run (M8). Que GNU sed acepte también el sufijo pegado a `-i` es su sintaxis
  documentada, pero no se ha comprobado en esta sesión: no hay GNU sed en ella.

## D · Decisiones

### D1. Dónde se mide: con las respuestas de cada modo, como `cita_sin_documento`

**Decisión.** La medida va en `medirLasRespuestas` (`informe.go`), sobre las respuestas de los grupos de los modos de
`respuestasQueSeJuzgan`, las mismas del `total` de `sin_activar` (V6). Cada `respuestaDelJuez` gana dos campos,
`pideElCotejo` y `cotejo`, que llena `medirLosCotejos`, como `medirLasCitasSinDocumento` llena `sinDocumento` (V7). El
umbral se compone en `respuestasMedidas.umbrales`, detrás de `cita_sin_documento` de cada modo (V8).

**Por qué.** FR-001 define el `total` como las sesiones del `total` de `sin_activar` de ese modo que son de una eval
que pide el cotejo. Con los mismos grupos, una sesión sin medir, una cortada, una de Haiku o una de la eval sin binario
ni servidor quedan fuera por la misma regla que en los demás umbrales (FR-005, FR-001), sin ningún caso aparte.

**Alternativas rechazadas.**
- *Anotarlo en el juicio de cada sesión* (`juzgarEnModo`, en `ResultadoDeEval`): cambia el tipo que se publica en
  `evals[]` de `informe.json` o, con un campo sin exportar, lo que comparan los tests del juicio, que FR-031 no deja
  cambiar; y el sondeo juzga sus sesiones con `Juzgar`, que llama a `juzgarEnModo` (`sondeo.go:530`), y FR-033 no deja
  cambiar lo que da.
- *Leerlo de `comandos_ejecutados`*: depende del ROJ del comando esperado de cada eval, y un cotejo sin `--roj` en la
  06 tiene que no contar (spec, «Edge Cases»).

### D2. Qué es cotejar: el criterio del juicio, con el comando esperado sin ROJ

**Decisión.** `cotejaUnDocumento(sesion)` dice si alguna invocación de `invocacionesDe(sesion)` satisface
`comandoDelCotejo`, el comando esperado `cita cotejar` sin ROJ: `satisface(invocacion, comandoDelCotejo)` (V3, V5).

**Por qué.** Es literalmente lo que pide FR-002: «el mismo criterio con el que el juicio de la sesión da hoy por
ejecutado un comando esperado de `cita`, sin mirar sus argumentos». Con la orden, código 0 y sin `--dry-run`,
`--describe` ni la ayuda; con la herramienta, un resultado que no es un error (V4). Los casos límite del spec —código 2,
`--dry-run`, un cotejo sin `--roj`, varios cotejos— salen de ese criterio, que ya prueban los tests del juicio; ninguno
lleva un test propio en este hito.

**Alternativas rechazadas.**
- *Los ECLI que leyó un `cita cotejar` en sus sobres* (`ExtraerECLICotejados`): depende de `--json` y de leer la
  salida, y no es el criterio de FR-002; una orden sin `--json` que coteja no contaría como cotejo.
- *Comparar el nombre de la herramienta con `cita_cotejar`*: una lista paralela al registro, que `invocacionesDe` ya
  traduce con `verbosDeLasHerramientas`.

### D3. El falso de `invocacionesDe` no se trata aparte

**Decisión.** `cotejaUnDocumento` llama a `invocacionesDe` sobre un `ResultadoDeEval` de usar y tirar y no mira su
segundo valor.

**Por qué.** `invocacionesDe` no devuelve ningún error: devuelve falso, y anota su motivo en el resultado, solo si el
registro de herramientas no se puede construir, y una sesión leída nunca llega ahí (V5). Si llegara, la sesión ya no
pasa por el motivo que su propio juicio anota (`juzgar.go:444`), y aquí sus llamadas no contarían: la sesión contaría
como sin cotejo, que es la lectura conservadora. No hay nada que prometer ni probar para ese estado (constitución,
«Gates», umbral de materialidad).

**Alternativa rechazada.** Partir `invocacionesDe` en una función que devuelva un error: cambia el código del juicio,
que FR-031 quiere quieto, para un estado al que no se llega.

### D4. El umbral vive dentro de los umbrales de las respuestas: no cambia cuándo existen

**Decisión.** `medirLasRespuestas` sigue devolviendo nil para una skill sin juez y sin ninguna eval que declare
`sentencias`. Dentro, `conCotejo` dice si alguna eval pide el cotejo (`pideElCotejo`), y con él se mide y se publica
`documento_sin_cotejar`.

**Por qué.** Con eso, FR-001 se cumple en toda skill que existe y en toda la que entrega el hito: la única con evals que
piden el cotejo, `jurisprudencia`, tiene juez y sentencias, y `boe-legislacion` y `legal-core` no tienen ninguna (M5).
Una skill con una eval que pide el cotejo y sin juez ni `sentencias` no la entrega ningún emisor del repositorio: el
juez del spec lo dejó anotado como por debajo del umbral de materialidad, y no se diseña ni se prueba para ella. Así,
además, `CONTRIBUTING.md:1141-1142` («existen solo si la skill tiene juez o si alguna de sus evals declara
`sentencias`») sigue siendo cierto y no cambia.

**Alternativa rechazada.** Añadir `!conCotejo` a la condición de `medirLasRespuestas`: solo sirve a esa skill que no
existe, y traería con ella `sin_activar` a una skill que hoy no lo tiene.

### D5. Qué es una eval que pide el cotejo

**Decisión.** `pideElCotejo(eval)`: `eval.Activa` y algún comando con `Applet` `cita` y `Verbo` `cotejar`. La eval de
cada respuesta es la de su sesión (`resultado.Eval`, el `Fichero` de la eval, `informe.go:1242-1257`).

**Por qué.** Es la definición del spec («Key Entities») y la de FR-001: «Qué evals entran lo dicen sus `comandos`, no
una lista en el código». Las respuestas de los grupos son ya de evals que activan la skill (V6); `Activa` está porque la
definición lo dice, no porque cambie nada hoy.

### D6. Nombre, descripción y orden

**Decisión.** El nombre es `documento_sin_cotejar:<modelo>:<modo>`, con su prefijo como los demás de las respuestas.
La descripción: «Respuestas de `<modelo>` en el modo `<modo>` cuya sesión no ejecutó ningún cita cotejar que consultara
y terminara bien, sobre sus respuestas medidas en las evals que activan la skill y piden cita cotejar». Va detrás de
`cita_sin_documento` de su modo: los quince de `jurisprudencia` quedan como en contracts §1.

**Por qué.** Es el nombre del hito. La descripción dice qué sesiones cuenta y sobre cuáles (FR-001). El orden deja
juntos los dos hechos de la sesión sobre sentencias de cada modo y no mueve ninguno de los trece entre sí: el
conjunto, la medida y las duraciones siguen detrás de los dos modos (FR-006, FR-030).

**Alternativa rechazada.** Ponerlo detrás de los de los dos modos, como un bloque aparte: separaría el umbral del modo
que nombra y cambiaría la regla de orden que `CONTRIBUTING.md` ya enuncia («primero los del modo orden y los del modo
herramienta»).

### D7. El motivo nombra cada sesión, sin nada más

**Decisión.** «umbral documento_sin_cotejar:`<modelo>`:`<modo>`: `<m>` de `<t>` (`<p>` %), y tiene que ser ≤ 0,0 %:
`<sesión>`; `<sesión>`…», con las sesiones que cuentan en orden de sesión y separadas por «; ». Lo compone
`nombradasEn` con una función nueva, `sinCotejarDelUmbral`, como `sinDocumentoDelUmbral` (V8).

**Por qué.** FR-003 pide el umbral, su medida y su condición, como todo umbral que decide, y detrás cada sesión por su
nombre y en orden. El modo ya va en el nombre del umbral, y el de la sesión del modo herramienta lo lleva además en su
nombre: no hace falta escribirlo detrás de cada una, como hace el conjunto de una clase, que junta los dos modos.

### D8. Los tests van en las tablas que ya hay; ninguna función de test nueva

**Decisión.** Los casos de FR-062 son cuatro casos nuevos de `TestUmbralesDeJurisprudencia`, con las evals y la
carpeta del juez del repositorio: `una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta`,
`una-sin-activar-y-sin-cotejar` y `una-sin-medir-y-sin-cotejar`; y `ninguna-cuenta`, que ya existe, pasa a exigir 0 de
12 en los dos modos. La premisa de FR-007 con las evals reales es un subtest nuevo de `TestEvalsDelRepositorio`,
`piden-el-cotejo`, que ejecuta `make skills-check`.

**Por qué.** El arnés de `TestUmbralesDeJurisprudencia` ya escribe el informe del plan del job con las sesiones del
modo orden por la traza y las del modo herramienta por sus llamadas (V19): un cotejo como orden y una llamada a
`cita_cotejar` son ya las 24 sesiones de `ninguna-cuenta` (SC-003), y las 18 por modo de las evals que no lo piden están
fuera del `total` de 12 en todos los casos (SC-004). Con 11 sesiones del modo que cotejan y 1 que no, los casos sin
cotejo distinguen «cada sesión» de «ninguna sesión con cotejo» (FR-064; M8).

**Alternativas rechazadas.** *Una función de test nueva con respuestas escritas a mano*, como
`TestUmbralesConJuezYConSentencias`: no pasaría por `respuestasQueSeJuzgan`, y no probaría que la sesión sin medir y
la de Haiku quedan fuera (FR-005, FR-062). *Un caso para cada caso límite del criterio de cotejo* (código 2,
`--dry-run`, sin `--roj`): ese criterio es el del juicio, que ya lo prueban sus tests (D2).

### D9. Cómo se escribe una sesión sin cotejo

**Decisión.** Un ayudante del test, `escribirSinElCotejo`, borra el directorio de la sesión y la escribe de nuevo con
`escribirSesionDeJurisprudencia` y la sesión modelo de su eval sin la operación `cotejar`: ni la orden en la traza ni la
llamada en el transcript. Se aplica antes que los demás cambios del caso, porque escribir de nuevo la sesión los
desharía. Se elige la 06: su respuesta no lleva ningún ECLI ni ninguna cita (V19), y sin el cotejo no cuenta en
`cita_sin_documento`; la serie de la 06 pasa con 2 de 3, y el único motivo de la raíz es el del umbral nuevo (M3).

**Alternativa rechazada.** Quitar las líneas del cotejo del transcript y de la traza ya escritos: depende de la forma
exacta de cada fichero, que ya fija `escribirSesionDeJurisprudencia`.

### D10. La sesión sin medir del test

**Decisión.** El caso `una-sin-medir-y-sin-cotejar` deja sin medir la sesión 01 de la 06 del modo herramienta con
`cambiarElFinalDeLaSesion` y el mensaje del límite de uso, el mismo de `TestInformeConSesionesSinMedir`, sin su cotejo.
Espera `documento_sin_cotejar:…:herramienta` con 0 de 11, los demás umbrales de ese modo sobre 29 respuestas, el
conjunto sobre 59, 59 votos y un solo motivo, el de la ejecución, que la nombra (M3). `medidasDeJurisprudencia` gana
`sinMedir` y `sinCotejar`, y `mensajesDeLosVotosDeJurisprudencia`, la sesión sin medir, que no se vota.

**Por qué.** FR-005 y FR-062 piden que una sesión sin medir no cuente y dé su motivo; sin cotejo, la prueba ve que
tampoco está en la `medida`. La serie de esa eval no da motivo de tasa con una sesión sin medir (`informe.go:1571`,
`:1775`).

### D11. `TestInformeFinalDeLosUmbrales` no cambia lo que espera

**Decisión.** `umbralesDeJurisprudencia` pasa a dar los quince (es lo que da el informe con las evals del repositorio),
y el caso `como-en-el-cierre-de-h26` lee los trece de H26 con un ayudante nuevo, `umbralesDeJurisprudenciaComoEnH26`, que
quita de los quince los dos `documento_sin_cotejar`.

**Por qué.** Ese caso reproduce un cierre que ya pasó y prueba el informe final, que el hito no cambia (FR-052): con el
ayudante sigue esperando lo mismo, «13 umbrales, los 9 que deciden» (M4), como `umbralesDeBoeLegislacionComoEnH26`.

**Alternativa rechazada.** Cambiar su texto esperado a «15 umbrales, los 11»: el caso dejaría de ser el del cierre de
H26 que dice su nombre, sin probar nada nuevo del informe final.

### D12. Los mutantes, en una copia fuera del árbol, con una línea cada uno

**Decisión.** El escenario 4 del quickstart, en un worktree de `$TMPDIR` que se retira al acabar: el mutante 1
cambia la condición de `sinCotejar` para que cuente solo si ninguna sesión del modo cotejó, y el mutante 2 pone
`Decide: false` en `umbralDeDocumentoSinCotejar`; cada uno con `sed -i.orig` sobre una sola línea. La tarea deja lo que
ve en `specs/022-h23-1-jurisprudencia-coteja-el/mutantes.md`.

**Por qué.** Es lo que pide FR-064, sin que ningún mutante pase por el árbol; las dos sustituciones se han probado tal
cual (M8). Es el método de H26.

### D13. La skill: el cotejo pasa a ser lo primero, y lo demás no se mueve

**Decisión.** En `skills/jurisprudencia/SKILL.md` cambian el paso 1, que manda el documento delante al paso 3 antes
de nada, y el paso 3, que abre con la regla: el cotejo de la ficha es lo primero que se hace con un documento de la
conversación, antes de cualquier frase sobre él; si es o no la sentencia pedida lo dice `es_la_pedida`, nunca una
comparación a ojo, que falla cuando el número de resolución del documento es el ROJ pedido. Lo demás del paso 3 —la
ficha y nada más, la referencia en la forma en que se pidió, la orden con la entrada estándar, los hallazgos, la
correspondencia y el código 2— sigue. No cambian el frontmatter, la entrada, los pasos 2 y 4 a 7, «Comandos» ni las
reglas (contracts/skill-jurisprudencia.md).

**Por qué.** FR-020 a FR-024. La redacción la escribe el run; el contrato fija dónde y qué tiene que decir.

**Alternativa rechazada.** Una regla nueva en «Reglas»: las reglas no se tocan (FR-024), y la instrucción tiene que estar
donde la skill decide qué hacer con el documento.

### D14. Sin aceptación e2e del binario

**Decisión.** «Aceptación e2e: no aplica».

**Por qué.** El hito no cambia el código del binario: `cmd/kitlegal` no enlaza `internal/evals` (V14), y cambiar el
binario fuera de ese paquete está fuera de alcance (FR-052). El `SKILL.md` viaja empotrado, y `kitlegal skills install`
lo copia igual que hoy: sus guiones comparan lo instalado con lo empotrado, sea cual sea el texto (V17). Lo que cambia la
skill lo juzga el job (SC-001). Precedentes: H24, H25 y H26.

### D15. Los documentos dicen lo nuevo y nada más

**Decisión.** Cambian `CHANGELOG.md` (*Unreleased*: la v0.1.1 de `jurisprudencia` en «Corregido», el umbral en
«Añadido»), `CONTRIBUTING.md` («Job de evals»), `docs/JURISPRUDENCIA.md:227` y `CLAUDE.md:67`, con el detalle de
contracts/documentos.md. `docs/WORKFLOW.md` no cambia (V22).

**Por qué.** FR-040 a FR-045. La v0.1.1 corrige un defecto de la v0.6.0 publicada (spec, «Resumen»): va en
«Corregido»; el umbral es un elemento nuevo de `umbrales`: va en «Añadido».

**Alternativa rechazada.** La skill en «Cambiado», como la v0.1.7 de `boe-legislacion` (V22): aquella cambiaba lo que
la skill decía, y esta corrige un paso que la skill se saltaba.

### D16. Ningún dato externo

**Decisión.** Ninguna fuente, grabación ni manifiesto `grabaciones.json`; el paso `grabar_datos` no tiene nada que
grabar, y ninguna tarea lleva `[datos]`.

**Por qué.** El hito no consulta ninguna fuente: mide lo que ya hay en las sesiones, cambia texto de una skill y de
cinco documentos. Ningún fichero de `testdata/`, `schemas/`, `evidencias/` ni `data/` cambia (FR-032, FR-052), y los
tests escriben sus sesiones sintéticas en `t.TempDir()`.
