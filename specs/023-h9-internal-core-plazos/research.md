# Research: H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1

Fase 0 del plan. Cuatro tablas: **V**, lo comprobado en esta sesión contra el repositorio, la toolchain o un
prototipo, con dónde; **M**, lo medido; **S**, lo que no se puede comprobar sin red o sin modelo, como supuesto, con lo
que pasa si no se cumple; y **D**, cada decisión con la alternativa rechazada y por qué. Ninguna sesión con modelo, ni
red, ni `make ci` entero: lo que no se pudo medir se dice.

Las sondas y el prototipo de esta sesión no dejaron nada en el árbol: dos tests temporales en
`internal/core/territorio/zz_sonda_test.go`, borrados tras ejecutarlos, y una copia de trabajo desacoplada en
`$TMPDIR/proto-h9` (`git worktree add --detach`, retirada con `git worktree remove --force`). `git status` al terminar:
solo `specs/023-h9-internal-core-plazos/`.

## V · Comprobado en esta sesión

| Id | Qué | Dónde y cómo |
|---|---|---|
| V1 | Las filas `boe.fiestas-laborales` y `bocm.fiestas-locales` de `docs/SOURCES.md` tienen «Revisado» 2026-10-17, y `boe.legislacion-consolidada`, 2026-09-13. La rama no tiene commits sobre `main` (`git status`: solo el directorio del feature), así que el fichero es el de `main` | `grep -n` en `docs/SOURCES.md` (líneas 42, 73 y 74) |
| V2 | `grabar-datos.sh` solo graba manifiestos cuya ruta casa con `(^|/)testdata/grabaciones\.json$`, con una sola `fuente` por manifiesto, ejecuta `go test -tags=grabacion -run '^TestGrabar' ./<paquete>/` y admite solo lo escrito bajo `<paquete>/testdata/` y `evidencias/<hito>/`; lo demás lo revierte | `scripts/workflow/grabar-datos.sh`, líneas 8-30, 75-77, 84, 103-119 |
| V3 | `testdata/evals/grabaciones.json`, el manifiesto de H5, no casa con ese patrón (`testdata/evals/grabaciones.json` no termina en `testdata/grabaciones.json`): `grabar_datos` no puede grabar en él. Su último cambio es de H6 | regex de V2; `git log -- testdata/evals/grabaciones.json` |
| V4 | `TestGrabarFixtures` (`internal/source/boe`) pide **todo** su manifiesto cada vez; `TestGrabarEvals` (`internal/evals`) siembra una caché desde las grabaciones y pide solo lo que falta | `internal/source/boe/grabacion_test.go:40-70`; `internal/evals/grabacion_test.go:36-115` |
| V5 | `httpx`: `ConFuente` rechaza un nombre vacío o con `/`, `\` o `..`; `ConRaizDeGrabacion` exige directorio; la grabación escribe `<raíz>/<fuente>/<método>_<url>.json` con el cuerpo en `cuerpo` si es UTF-8 y en `cuerpo_base64` si no; solo `KITLEGAL_RECORD=1` la enciende; `Replay(dir)` sirve lo grabado | `internal/httpx/cliente.go:109-174`, `grabar.go:16-81`, `reproducir.go:117-134` |
| V6 | El BOCM publica una línea «Municipio: fechas.» por municipio o entidad; las correcciones dicen «establecer para los municipios que se relacionan las siguientes fiestas laborales» y reescriben la línea entera; catorce «no comunicado» en la principal de 2026, que las dos correcciones de 2026 completan. Leganés 2026: «14 de agosto y 9 de octubre» | `evidencias/adr-0039/bocm/*.xml`, leídos con `sed` y `grep` |
| V7 | **Seis nombres del BOCM no son el del INE** con el pliegue de `territorio` (`formasDelNombre`, `Plegar`): «Berzosa de Lozoya», «Braojos de la Sierra», «Buitrago de Lozoya» (2025 y 2026), «Talamanca del Jarama» (corrección de 2025), «Moraleja de En medio» y «Valdetorres del Jarama» (2026); el INE escribe «Berzosa del Lozoya», «Braojos», «Buitrago del Lozoya», «Talamanca de Jarama», «Moraleja de Enmedio» y «Valdetorres de Jarama». No casan tampoco «Cerceda», «Mataelpino» y «El Espartal», que son entidades | sonda 1: test temporal en `internal/core/territorio` con el registro real y las siete disposiciones |
| V8 | **Tres líneas con erratas de forma** además de la de Valdeolmos-Alalpardo: sin dos puntos («Nuevo Baztán 11 de mayo…», «Olmeda de las Fuentes 6 de abril…», 2026), sin «de» ante el mes («12 diciembre», «15 septiembre»), con un punto suelto o un blanco antes del final («.18 de agosto», «14 de octubre .») y «no comunicado por el ayuntamiento» en minúscula | `grep -vE` con la gramática estricta sobre `evidencias/adr-0039/bocm/*.xml` |
| V9 | **Con la regla de cruce de D13 y la gramática tolerante de D13, la derivación da 179 de 179 municipios con entrada en 2025 y en 2026**, una sola línea sin leer por año (Valdeolmos-Alalpardo), tres líneas sin municipio por año (Cerceda, Mataelpino, El Espartal) y, contra las filas de `entidad_codigo` 0 de 2025 del fichero CKAN, **exactamente las diez diferencias del `LEEME.md`**: Belmonte de Tajo, Cabanillas de la Sierra, La Hiruela, Montejo de la Sierra, Patones y Talamanca de Jarama (faltan en el fichero), Daganzo de Arriba y Majadahonda (corrección sin aplicar), Fuente el Saz de Jarama (fecha) y Valdeolmos-Alalpardo (lectura). Leganés: 2025 `08-14`, `10-10`; 2026 `08-14`, `10-09`. Brunete 2026: `05-15`, `09-14`, de `BOCM-20251229-19`. Madrid 2026: `05-15`, `11-09`. El Vellón 2026: `10-02`, `10-05`. El Boalo 2026: `01-20`, `05-15` | sonda 2: test temporal con la derivación entera y el cotejo; salida en esta sesión |
| V10 | El fichero CKAN: 354 filas de 2025, 173 municipios con `entidad_codigo` 0 (179 menos los seis que añadieron las correcciones); las de entidad (Cerceda y Mataelpino en `023`, El Espartal en `168`, Alalpardo en `162`) no tienen código INE propio | `jq` sobre `evidencias/adr-0039/festivos_locales_historicos.json` |
| V11 | Un applet nuevo en los dos registros rompe exactamente: `TestPuntoDeEntrada` (`cmd/kitlegal/main_test.go`), `TestRegistroDeProduccion`, `TestRegistroDeE2E`, `TestHerramientasDelServidor` (12 y 15 herramientas → 13 y 16), `TestEsquemasCubrenTodosLosVerbos` (falta `schemas/plazo.json`) y los guiones `argumentos`, `h21-mcp-herramientas`, `h21-mcp-proceso` y `h21-mcp-protocolo`. **No** rompe `internal/empaquetado` (su lista pide que estén, no que sean las únicas: `ejecutar_test.go:33-41`; y `snapshot_piezas_test.go`, en la raíz, con la etiqueta `snapshot` y fuera de `make ci`, también usa `assert.Subset`, línea 607) ni los guiones de la instalación de desarrollo (`-tags integration`) | prototipo en `$TMPDIR/proto-h9`: `go test ./...` y `go test -tags integration ./internal/skills/... ./internal/empaquetado/... ./internal/app/...` |
| V12 | `TestTablaDeComandosCoincideConLaGramatica` ejecuta cada orden con «x» en cada argumento y `--describe`: un `enum` de Kong en un argumento de posición o una bandera entera la hacen salir con 2. Con cadenas, pasa | prototipo: los dos fallos y el verde, en esta sesión |
| V13 | `--describe` del prototipo: `entrada` con `tipo`, `inicio`, `municipios` (lista), `duracion` y `unidad`, y `x-banderas: ["duracion","unidad"]`; la fila generada de la tabla es ``kitlegal plazos calcular <tipo> <inicio> <municipios>... [--duracion <duracion>] [--unidad <unidad>]`` con `plazos_calcular` | prototipo: `go run ./cmd/kitlegal plazos calcular x x x --describe` y `-args -regenerar-skills` |
| V14 | Dos ficheros más en `skills/legal-core/references/` no rompen ningún guion `h19-skills-*` ni los de la instalación de desarrollo | prototipo: `TestEntregaDelHito` y `go test -tags integration ./internal/skills/` |
| V15 | Con nueve evals en `legal-core`, `TestDefinicionDelJob` falla en `peor-caso` y en `tope-con-las-evals-del-repositorio`: su literal `peorCasoDeLegalCore` pasa de `6469 s = 485 s + (⌈19 / 2⌉ + ⌈18 / 2⌉ + ⌈6 / 2⌉) × …` a `14629 s = 485 s + (⌈49 / 2⌉ + ⌈48 / 2⌉ + ⌈6 / 2⌉) × (22 s + 240 s + 10 s)`, el de FR-050 | prototipo con cinco evals copiadas de la 01 |
| V16 | `depguard` R1 se aplica a los `_test.go` del dominio (`files: "**/internal/core/**"`) y deniega `os`, `io`, `log/slog` y los adaptadores; solo `internal/app` importa el paquete `data` | `.golangci.yml:105-154`; `git grep 'kitlegal/data"'` |
| V17 | El fallo por un nombre de varios municipios de `territorio` es `{clase: argumentos, mensaje}` con los candidatos en el mensaje, «<código INE> <nombre> (<provincia>)» separados por «; » | `internal/core/territorio/errores.go:67-83`; `internal/app/testdata/script/territorio-matriz.txtar:125-132` |
| V18 | La forma fija de un aviso es `⚠ <ETIQUETA>:` con la tolerancia de `patronDeEtiqueta`; la cita es `[…, BOE-A-…, bloque <id>]` (`formaDeCita`) | `internal/evals/avisos.go:16-93`; `citas.go:8-44` |
| V19 | `medirLasRespuestas` da umbrales de las respuestas solo con juez o con sentencias; por modo, en este orden: `sin_activar`, las clases, `cita_sin_documento`, `documento_sin_cotejar`; el motivo de uno que decide es `umbral <nombre>: <medida>, y tiene que ser <condición>` seguido de las sesiones que cuentan separadas por «; » | `internal/evals/informe.go:1360-1401`; `umbrales.go:116-128, 284-316, 583-622` |
| V20 | Los ids de bloque del índice grabado de la LPAC: `a20`, `a21`, `a24`, `a30`, `a112`, `a121`, `a122`, `a123`, `a124` y `daprimera` (la disposición adicional primera); todos válidos para `boe.ValidarBloque` | `jq` y `grep -o` sobre `internal/source/boe/testdata/…/id_BOE-A-2015-10565_texto_indice.json`; `internal/source/boe/ids.go:68-80` |
| V21 | Grabado ya: índice y metadatos de la LJCA (`BOE-A-1998-16718`), la LOPJ (`BOE-A-1985-12666`), la LGT (`BOE-A-2003-23186`) y el Código Civil (`BOE-A-1889-4763`); el art. 20 de la LTAIBG (`BOE-A-2013-12887`, `a20`); y de la LPAC, `a21` y `a1`, `a22`, `a23`. Los ids de bloque que el plan pide de esas normas, en V29 | `ls testdata/evals/boe.legislacion-consolidada/` e `internal/source/boe/testdata/boe.legislacion-consolidada/` |
| V22 | `description` de una skill: 1 024 caracteres como mucho; la de `legal-core` v0 tiene unos 855 | `internal/skills/frontmatter.go:53, 438`; `wc -m` |
| V23 | `testing/quick` está congelado («frozen and is not accepting new features») y usa `math/rand` | `go doc testing/quick` (go1.27.2) |
| V24 | El esquema de eval tiene una forma por comando (`oneOf`), `prohibidos` admite cualquier `{applet, verbo}`, y su `fecha` es `AAAAMMDD` | `jq` sobre `schemas/eval.yaml.json` |
| V25 | Los días de la semana de las fechas de los casos | `cal -h 2026`, `cal -h 1 2027`, `cal -h 2 2027` |
| V26 | Los esquemas publicados se llaman como su entidad (`municipio.json`, `cita.json`), con la tabla `ficherosDeEsquemas` | `internal/app/esquemas_test.go:42-69` |
| V27 | El módulo de YAML del repositorio es `go.yaml.in/yaml/v3` v3.0.5, la ruta mantenida de `yaml.v3` que fija la constitución; lo usan ya `internal/core/territorio` (`KnownFields`) e `internal/skills` | `go.mod:15`; `git grep '"go.yaml.in/yaml/v3"'` |
| V28 | Los documentos que hoy dicen de `legal-core` lo que el hito deja falso: `CONTRIBUTING.md` 85 (lista de applets), 382 (esquemas), 471 (verbo con su herramienta), 600-630 (formato y reglas del conjunto), 944 (concurrencia `1`), 963 (cuatro evals, 42 sesiones), 1145 (`umbrales` vacío) y 1210; `README.md` 24-37; `docs/WORKFLOW.md` no nombra `legal-core` | `grep -n` |
| V29 | **Los ids de bloque de los índices grabados de las otras normas del manifiesto de preceptos**: en la LOPJ, los arts. 182, 183 y 185 son `acientoochentaydos`, `acientoochentaytres` y `acientoochentaycinco` («Artículo ciento ochenta y dos», «… tres», «… cinco»; actualizados el 2022-12-23, el 2022-12-23 y el 1985-07-02), y no hay ningún `a182`, `a183` ni `a185`; en el Código Civil, el art. 5 es `art5` («Art 5», 1974-07-09), y no hay ningún `a5`; en la LJCA, `a46` y `a128`; en la LTAIBG, `a20` y `a24`; en la LGT, `a97`. Todos con la forma de `boe.ValidarBloque` | `grep -o` y `grep -c` sobre `testdata/evals/boe.legislacion-consolidada/GET_…_id_<norma>_texto_indice.json` de `BOE-A-1985-12666`, `BOE-A-1889-4763`, `BOE-A-1998-16718`, `BOE-A-2013-12887` y `BOE-A-2003-23186`; `internal/source/boe/ids.go:80-82` |

## M · Medido

| Id | Qué | Valor |
|---|---|---|
| M1 | Sobre del caso (a), con `--json`, como lo fija contracts/applet-plazos.md §4, con el `hash` de 64 cifras y el salto de línea final | 772 bytes (`wc -c` sobre el ejemplo de §4 con el `hash` entero; el mismo método da los 854 del sobre con las tres claves que §4 ya no lleva) |
| M2 | Sobre del caso (b), con un aviso, igual | 858 bytes (940 con las tres claves) |
| M3 | `--describe` de `plazos calcular` en el prototipo, en una línea; su `entrada` sola | 4 252 y 531 bytes. El entero se midió con la salida del prototipo, anterior a quitar de §4 tres claves; con el binario del hito, medido en el barrido con `jq -c` y `wc -c`, son 4 132 y 531 bytes: la `entrada`, la única cifra que usa contracts/applet-plazos.md §8, no cambia |
| M4 | `SKILL.md` de `legal-core` en `main`; con la tabla de `plazos` regenerada | 190 y 196 líneas |
| M5 | Peor caso del trabajo de `legal-core` con nueve evals, a dos sesiones | 14 629 s (V15), dentro de los 352 minutos (21 120 s) |

No se ha medido: el tiempo de `plazos calcular` (no hay implementación), las sesiones del job (no hay sesiones con
modelo en un paso) ni `make ci` entero.

## S · Supuestos no verificados

| Id | Supuesto | Si no se cumple |
|---|---|---|
| S1 | **Formato de la resolución del BOE** (`/diario_boe/txt.php?id=…`): el anexo es una tabla con una fila por fecha y una columna por comunidad o ciudad autónoma, con la marca «***» (fiesta nacional no sustituible), «**» (nacional respecto de la que no se ejerció la sustitución) o «*» (de comunidad autónoma), y notas que limitan alguna fiesta a una parte de la comunidad. No se ha podido leer: no está grabada y la sesión no tiene red | La tarea del analizador lo escribe contra la grabación que deja `grabar_datos` antes que ella (D15). La semántica de D14 no cambia: ámbito por la marca, parte por la nota; si una marca o una nota no se pudiera leer sin interpretarla, ese festivo queda fuera, como en el BOCM (ADR 0039, punto 3). **No se cumplió** (T007, `gates/supuestos.md`): la leyenda de las tres resoluciones grabadas da «*» a la nacional no sustituible, «**» a la nacional respecto de la que no se ejerció la sustitución y «***» a la de comunidad autónoma, y manda lo grabado (D14) |
| S2 | Los nombres de las comunidades de la tabla del BOE casan, plegados y con la regla compacta de D13, con los de `data/territorio/comunidades/` | El test de derivación falla nombrando la columna; la regla de cruce es la misma de D13 y no admite una lista de alias. **No se cumplió** (T007, `gates/supuestos.md`): cinco nombres de las cabeceras no casan por la forma compacta («Asturias», «Com. Madrid», «C. Foral Navarra», «Ciudad de Ceuta», «Ciudad de Melilla»), y detrás de ella van dos reglas genéricas, sin alias: las mismas palabras con la abreviada con punto, y el nombre con algo delante unido por «de» o «del» (`TestCruceConLasComunidades`) |
| S3 | **Festivos decisivos de los casos que no salen del BOCM**: 12 de octubre de 2026, 25 de diciembre de 2026 y 1 de enero de 2027 (nacionales no sustituibles), 6 de enero de 2027 (Epifanía, nacional en Madrid), 1 de enero de 2025; y ningún festivo autonómico de Madrid ni de Castilla y León en los días que deciden cada caso (contracts/casos-de-plazos.md). Se eligieron los tramos para que solo decidan estos días, los fines de semana y los locales leídos del BOCM (V9) | El caso falla en `make ci`. La tarea de la tabla se escribe después de derivar `data/festivos/` y lee de ahí el festivo; si lo grabado contradice un razonamiento, corrige el razonamiento y la fecha con lo grabado y lo anota en `gates/supuestos.md`. Los guiones de aceptación, que se congelan antes, solo usan casos cuyo día decisivo es el 9 o el 12 de octubre de 2026, el 14 de enero de 2027, el 15 de septiembre o el 28 de diciembre de 2026 y el 1 al 3 de enero de 2025 (D27) |
| S4 | **Texto vigente de los preceptos** de LPAC (arts. 21.3, 24.1, 24.4, 30, 112.2, 112.4, 122.1, 124.1, DA 1ª.2), LJCA (46.1, 128.2), LOPJ (182.1 con «los días 24 y 31 de diciembre», 183 con los días del 24 de diciembre al 6 de enero, 185), LTAIBG (20.1, 24.2, 24.6), LGT (97), Código Civil (5.1) y texto refundido de Tráfico (1), tal como se recuerdan, sin leerlos en esta sesión. Los ids de sus bloques no se recuerdan: salen de los índices grabados (V20, V29), salvo el de Tráfico (S8) | Lo comprueba el test de D16 contra lo grabado, por apartado y con sus palabras clave. Si el art. 182.1 grabado no enumera el 24 y el 31 de diciembre, el cómputo no los trata como inhábiles y la tarea lo anota; si una duración no está en su apartado, el test falla y el tipo se corrige con lo grabado. **El art. 182.1 grabado no los enumera** (redacción de la LO 14/2022; T005, `gates/supuestos.md`): el cómputo del contencioso no trata el 24 ni el 31 de diciembre como inhábiles |
| S5 | El identificador del texto refundido de la Ley sobre Tráfico es `BOE-A-2015-11722` | No se escribe de memoria: lo resuelve la búsqueda del manifiesto (D15), como los de `data/normas.yaml`, y `data/cobertura.yaml` lleva el que da lo grabado |
| S6 | Lo que harán las sesiones del modelo que decide con `legal-core` v1: que se active con las preguntas de plazos, que escriba las formas fijas y que no pida `plazos calcular` en las de cobertura | Lo mide el job de evals del cierre (SC-001); un rojo de lo que entrega el hito se repara (FR-071) |
| S7 | Lint del código nuevo (`golangci-lint` con `gocyclo`, `dupl`, `gosec`, `misspell`…) | Lo ve `make ci` en cada tarea; ningún `//nolint` (constitución, criterio 1) |
| S8 | El art. 1 del texto refundido de Tráfico es el bloque `a1` de su índice consolidado, la forma de la LPAC, del mismo año (V20, V21), y no la de la LOPJ o el Código Civil (V29). Su índice no está grabado y la sesión no tiene red | `TestGrabarEvals`, en el paso `grabar_datos` tras la tarea `[datos]` del paso 2, pide un bloque que no existe y falla las dos veces; `grabar-datos.sh` deja el dossier «dependencia externa inaccesible» y detiene el run hasta que una persona corrija el id en el manifiesto (`data/cobertura.yaml` llega después, en el paso 4, con el que quede). Ningún otro bloque depende de él |

## D · Decisiones

### D1 · Dónde vive cada pieza

- **Dominio**: `internal/core/plazos`, puro (R1): el calendario de festivos con su cobertura, los tipos, el cómputo,
  los avisos y sus textos, los errores de argumentos, la lectura de `data/plazos.yaml` y de `data/festivos/` y la
  escritura de `data/festivos/` en su forma canónica. Recibe bytes; no abre ficheros.
- **Applet**: `internal/app/plazos.go`, en los dos registros, con `data/festivos/` y `data/plazos.yaml` empotrados por
  el paquete `data`, como `data/territorio/` (V16: solo `internal/app` lo importa).
- **Adaptadores sin consulta en ejecución**: `internal/source/boe/fiestas` (fuente `boe.fiestas-laborales`) e
  `internal/source/bocm` (fuente `bocm.fiestas-locales`). Cada uno lleva su manifiesto, su `TestGrabar*`, el analizador
  de su boletín y la derivación de su fichero de `data/festivos/`.
- **Verificación de los preceptos y grabaciones de las evals**: `internal/evals`, junto a `TestIdentificadoresDeLasNormas`.

Rechazadas: (a) el analizador de los boletines en el dominio: es el formato de una fuente, y el hexágono lo pone en
`internal/source/<fuente>` (constitución IV); (b) un solo paquete para las dos fuentes: `grabar_datos` admite una
`fuente` por manifiesto y un manifiesto por paquete (V2); (c) añadir los boletines a `internal/source/boe`: ya tiene
manifiesto, de `boe.legislacion-consolidada`, y `TestGrabarFixtures` lo pide entero cada vez (V4).

### D2 · La gramática de `plazos calcular`

`kitlegal plazos calcular <tipo> <inicio> <municipios>... [--duracion <n>] [--unidad <unidad>]`. Todo son cadenas, y
el dominio decide qué vale, con clase `argumentos`: un `enum` de Kong o una bandera entera rompen la tabla de comandos
generada (V12), y es el patrón de `territorio` y `cita`, donde el dominio decide. Los municipios van por posición, como
los términos de `boe buscar`: la herramienta los recibe como la lista `municipios` (V13). Rechazada: banderas
`--interesado`, `--sede` y `--juzgado`: dan nombres a los papeles, pero FR-012 cuenta municipios («un número de
municipios distinto») y la herramienta tendría tres argumentos opcionales cuya combinación válida depende del tipo.

### D3 · Los tipos

Siete códigos: `alzada`, `reposicion`, `contencioso`, `acceso-resolver`, `acceso-reclamar`, `silencio` y `generico`.
`silencio` y `generico` no fijan su duración: exigen `--duracion` y `--unidad` (`meses`, `dias-habiles`,
`dias-naturales`); los demás rechazan las dos. Un solo `generico` con unidad, y no tres tipos, porque la unidad es la
misma entrada que pide el silencio (FR-009, FR-012: «una duración del silencio sin su unidad») y el precepto lo da la
unidad (art. 30.2 o 30.4). Rechazada: tres tipos genéricos y la unidad del silencio dentro del código
(`silencio-meses`): FR-012 trata la unidad como una entrada que puede faltar.

### D4 · Municipios que cuentan

En un plazo de la LPAC, exactamente dos, primero el del interesado y después el de la sede del órgano, que pueden ser
el mismo (FR-007); en el contencioso, uno (FR-010). Cada uno se resuelve con `territorio.Registro.Resolver`, el mismo
que `territorio resolver`. `data.municipios` los lista una vez cada uno, en el orden dado: dos veces el mismo código es
un municipio que cuenta.

### D5 · Lo que devuelve

Contracts/applet-plazos.md §4: `fecha`, `tipo`, `preceptos`, `municipios`, `festivos` (los saltados), `avisos` y
`cobertura`, lo que enumera FR-001 y nada más. Rechazadas: devolver la fecha de inicio, la duración y el cómputo, que
quien llama ya tiene o lee en `references/recursos_y_plazos.md` y que ningún requisito pide, como `comunidad` (D29).
`cobertura` va siempre y es la
declaración de principio IX: por ámbito y por tramos de años, hasta qué disposición llegan los datos o qué falta; cada
tramo de años cargado es un año (cada año tiene su disposición) y cada hueco es un tramo, así que una llamada lleva como
mucho cinco entradas por ámbito sea cual sea la duración. Rechazada: una entrada por año del tramo, que crece con la
duración pedida (criterio de uso).

### D6 · Avisos y formas fijas

Cuatro códigos, cada uno con su etiqueta y su texto en forma fija (contracts/applet-plazos.md §5):
`faltan-festivos-nacionales-y-autonomicos`, `faltan-festivos-locales`, `dias-del-art-183-lopj` y
`festivo-de-una-parte`. El texto de un aviso de cobertura lleva ya hasta qué disposición llegan los datos de su ámbito:
así la skill lo dice solo junto al aviso (FR-041) copiando su línea. Dos códigos de cobertura, y no uno, porque las dos
ausencias son distintas (US2.3 las da a la vez) y la eval (b) espera la local. Las dos formas de la respuesta que no
emite el binario —`⚠ PLAZO:` y `⚠ NO CUBIERTO:`— viven en `internal/evals`, que las comprueba, y en `SKILL.md`.

### D7 · El cómputo de la LPAC

Arts. 30.2 a 30.6 LPAC. Un día es inhábil si es sábado o domingo, o festivo nacional o autonómico de la comunidad de
alguno de los municipios que cuentan, o local de alguno (30.2 y 30.6). Días hábiles: el enésimo día hábil desde el
siguiente a la fecha de inicio (30.3). Días naturales: el enésimo día natural, y si es inhábil, el primer hábil
siguiente (30.2, párrafo segundo, y 30.5). Meses: el mismo día del mes de vencimiento, o el último si no lo hay (30.4),
y si es inhábil, el primer hábil siguiente (30.5). Los festivos saltados son los del tramo que no se contaron (días
hábiles) o los de la prórroga (días naturales y meses); los fines de semana no son festivos y no se listan. Un festivo
autonómico limitado a una parte de la comunidad no hace inhábil el día y, si está en el tramo, da su aviso (FR-011).

### D8 · El silencio

FR-009 pide fijar desde qué día se cuenta y cómo se llega al día del silencio. Se cuenta como cualquier plazo del art.
30 desde el día siguiente a la fecha de inicio (FR-002; arts. 30.3 y 30.4), que es la del acuerdo de iniciación o la de
entrada de la solicitud en el registro del órgano competente (art. 21.3.a y b), con la duración y la unidad que da
quien llama y con la prórroga del 30.5 si el último día es inhábil: ese día es el **vencimiento** del plazo máximo. El
silencio se produce al vencer el plazo sin resolución notificada (art. 24.1) y sus efectos corren «desde el vencimiento
del plazo máximo» (art. 24.4): el día que se da es el **siguiente al vencimiento**, el primero en que la Administración
ya no ha resuelto en plazo. Rechazada: dar el propio vencimiento, en el que la Administración todavía puede notificar.

### D9 · El contencioso

- **Agosto** (art. 128.2 LJCA: «no correrá el plazo para interponer el recurso contencioso-administrativo»): el plazo
  de dos meses del art. 46.1 se cuenta de fecha a fecha (art. 185.1 LOPJ, que remite al Código Civil, art. 5.1). Si la
  fecha de inicio es de agosto o el día siguiente lo es, el plazo empieza el 1 de septiembre, como si la notificación
  fuera del 31 de agosto. Si entre el día siguiente al inicio y el vencimiento de fecha a fecha cae algún día de
  agosto, el vencimiento pasa al mismo día del mes siguiente (o al último de ese mes): agosto entero no corre y el plazo
  se alarga en él. Es la fecha más temprana de las lecturas posibles; la otra lectura de una notificación de agosto
  (empezar a contar el 1 de septiembre como día de notificación) da un día más tarde.
- **Último día** (art. 182.1 LOPJ): inhábil si es sábado, domingo, fiesta nacional o festivo a efectos laborales de la
  comunidad o de la localidad de la sede del juzgado; se prorroga al primer día hábil siguiente con esos mismos
  inhábiles (art. 185.2 LOPJ). El 24 y el 31 de diciembre no: el art. 182.1 grabado, en la redacción de la LO
  14/2022, ya no los enumera (S4; corregido con lo grabado en T005, `gates/supuestos.md`).
- **Días del 24 de diciembre al 6 de enero** (art. 183 LOPJ): se cuentan; si el tramo pasa por alguno, aviso
  `dias-del-art-183-lopj`. Lo mismo vale para un último día que cae en agosto tras una prórroga: solo los inhábiles del
  182.1 lo mueven, y es la fecha más temprana.

Rechazada: descontar agosto día a día (31 días más), que en un plazo de meses no tiene día equivalente y da otra fecha
según el mes de inicio.

### D10 · Errores y su orden

Clase `argumentos` (2): el tipo; la fecha (no `AAAA-MM-DD` o no un día del calendario); la duración (falta en
`silencio` o `generico`, no es un entero positivo, o se da a un tipo que la fija); la unidad (falta con duración, no es
del vocabulario, o se da a un tipo que la fija); el número de municipios. Se comprueban en ese orden y el mensaje nombra
el primero. Después, los municipios: si alguno daría el 2 en `territorio`, 2, con el mensaje de `territorio` de cada
uno detrás del municipio pedido, separados por « | » (V17); y si todo vale y alguno no está en la relación, 3
(`no-encontrado`). Unas fuentes que no cargan, `inesperado` (1), como en `territorio`. La unidad que se da a un tipo
que la fija no está en la lista de FR-012, pero es la misma regla que su duración, y aceptarla sin efecto escondería
un error de quien llama (lectura conservadora; `gates/supuestos.md`).

### D11 · El sobre

`fuente` `kitlegal.plazos` y `url` `kitlegal:applet/plazos`, el espacio de los applets calculados (ADR 0006); la fecha
de consulta la pone el reloj del kernel, como en `cita`: `data` y `hash` son iguales en cada ejecución (FR-013).
Ninguna operación de grafo: `Resultado.Grafo` va vacío. `--offline` no cambia nada y `--dry-run` no describe nada,
porque no hay ninguna capa con efectos.

### D12 · Los ficheros de `data/festivos/`

Por boletín y comunidad (principio IX): `data/festivos/nacionales-y-autonomicos.yaml`, de las resoluciones del BOE, y
`data/festivos/locales/13.yaml`, de las del BOCM, con el código de la comunidad como en
`data/territorio/comunidades/13.yaml`. Cada uno lleva su cabecera de fichero generado, sus años con sus disposiciones
y sus filas, cada fila con su `source`. Contracts/datos-de-plazos.md §1. Rechazada: un fichero por año, que repartiría
la misma resolución en tres sitios.

### D13 · La derivación del BOCM

- **Orden**: las disposiciones de un año, por fecha de publicación (la del identificador); una línea de una posterior
  sustituye entera la de ese municipio en las anteriores, con su `source` (V6).
- **Gramática de la línea** (V8): `.⇥<nombre>[:] <cuerpo>[.]`. El nombre va hasta los dos puntos o, sin ellos, hasta la
  primera cifra; el cuerpo es «no comunicado por el ayuntamiento» (sin distinguir mayúsculas) o una lista de fechas
  `<d>[ y <d>] [de ]<mes>` unidas por « y », con blancos y puntos sueltos tolerados. Cualquier otra cosa —paréntesis,
  un mes que no es un mes— es una línea que no se lee sin interpretarla (FR-023).
- **Cruce con `data/territorio/`** (V7): solo con los municipios de la comunidad del boletín (FR-022). Primero, la
  forma compacta: cada forma conocida del nombre (`territorio.Plegar` y las del artículo pospuesto) sin las palabras
  «de», «del», «la», «las», «los» y «el» y sin blancos; si casa con un solo municipio, es ese. Si no casa con ninguno,
  el nombre de la línea que empieza por todas las palabras de una forma de un único municipio seguidas de «de» o
  «del» («Braojos de la Sierra» → «Braojos»). Con cero o más de un candidato, la línea queda fuera con el motivo
  `sin-municipio`. Ninguna lista de alias, ningún municipio con trato propio (principio IX). Los municipios de la
  comunidad y sus formas los da `territorio.Registro.MunicipiosDe`, nuevo, para que `data/territorio/` tenga un solo
  lector y las formas, una sola definición (`formasDelNombre`).
- **Motivos de falta** (vocabulario cerrado, FR-022): `no-comunicado`, `linea-sin-leer`, `sin-linea` (el municipio no
  tiene línea en ninguna disposición del año) para un municipio; `sin-municipio` para una línea.

Rechazadas: el pliegue de `territorio` tal cual, que deja fuera seis nombres y da cuatro diferencias de más con el
fichero CKAN (V7, V9; FR-084 falla); una tabla de alias, que es un caso por municipio y no sale de ninguna fuente; el
fichero CKAN como tabla de nombres a códigos, que el ADR 0039 (punto 5) excluye como fuente. La regla de prefijo solo
se aplica si la compacta no da ninguno, y no casa con ninguna de las entidades de hoy (V9).

### D14 · La derivación del BOE

Por cada resolución, cada fila del anexo da una fecha del año de la resolución y, por cada columna marcada, la
comunidad con su ámbito: `nacional` con «*» o «**», `autonomico` con «***» (lo grabado contradijo a S1, que los daba al
revés; T007, `gates/supuestos.md`). Una nota que limita la fiesta a una parte
de una comunidad da la fila con `parte` y sin efecto en el cómputo (FR-011, FR-025); las notas se leen cita a cita, y
una sustitución en una parte, como la de Arán, da dos filas con `parte`: la sustituida, del resto de la comunidad, y
la que la sustituye (revisión final, `gates/supuestos.md`; contracts/datos-de-plazos.md §1.1). Las columnas se cruzan con las 19
comunidades de `data/territorio/comunidades/` con la regla compacta de D13 (S2) y, como S2 no se cumplió con lo grabado,
con las dos reglas genéricas que la siguen, sin lista de alias (T007, `gates/supuestos.md`). Una fila que no se lee sin
interpretarla queda fuera, como en el BOCM. El test comprueba que las 19 tienen festivos cada año (FR-020: todas las
que trae cada resolución), y nada más del recuento: ni un tope de festivos por comunidad ni por municipio, que ningún
requisito pide y que contaría también las filas con `parte` que FR-025 manda guardar.

### D15 · Qué se graba, dónde y con qué

| Manifiesto | `fuente` | `TestGrabar*` | Escribe en |
|---|---|---|---|
| `internal/source/boe/fiestas/testdata/grabaciones.json` | `boe.fiestas-laborales` | `TestGrabarFiestasLaborales`, a 1 s | `evidencias/h9/boe.fiestas-laborales/` |
| `internal/source/bocm/testdata/grabaciones.json` | `bocm.fiestas-locales` | `TestGrabarFiestasLocales`, a 10 s | `evidencias/h9/bocm.fiestas-locales/` |
| `internal/evals/testdata/grabaciones.json` | `boe.legislacion-consolidada` | `TestGrabarEvals`, que gana el segundo manifiesto | `internal/evals/testdata/boe.legislacion-consolidada/` |

Los dos de festivos piden con `httpx.New(ConFuente, ConRaizDeGrabacion($KITLEGAL_EVIDENCIAS), ConIntervalo)`, y la
derivación los lee con `httpx.Replay` sobre esa misma raíz: el formato de grabación tiene un solo lector (V5). El de
preceptos tiene la forma del manifiesto de H5 (`busqueda`, `titulo_empieza_por`, `bloques`), de modo que el
identificador de cada norma, el del texto refundido de Tráfico incluido, sale de la búsqueda grabada (S5) y lo
comprueba el mismo lector que `data/normas.yaml`. `TestGrabarEvals` recorre los dos manifiestos y graba solo lo que
falta (V4): lo de H5 ya está grabado, así que solo escribe en la raíz nueva. La raíz nueva entra en
`UnionDeGrabaciones` como tercer conjunto, y `TestGrabacionesSinSolape` la cubre. Rechazadas: ampliar
`testdata/evals/grabaciones.json`, que `grabar_datos` no puede grabar (V3); y ampliar el manifiesto de
`internal/source/boe`, que se volvería a grabar entero y cambiaría fixtures de otros tests (V4).

### D16 · La verificación de los preceptos (FR-030, FR-031, FR-010)

Un test de `internal/evals`, con la caché preparada desde las grabaciones, lee cada bloque con `boe articulo` en
proceso, como lo lee la skill, y comprueba: que cada norma y cada bloque de `data/plazos.yaml` y de
`data/cobertura.yaml` están grabados; que el párrafo del apartado de cada tipo que fija la duración (el que empieza por
«1.» en el 124.1) lleva la duración en palabras («un mes», «dos meses»); que cada identificador de
`data/cobertura.yaml` es el que da la búsqueda grabada; y que el texto de cada precepto del cómputo y de la cobertura
lleva sus palabras clave (contracts/datos-de-plazos.md §4). Así ningún fragmento de texto legal se copia a `data/`, y
lo que el plan recuerda de cada precepto (S4) lo confirma o lo rechaza lo grabado. Rechazada: un campo `texto` con el
fragmento en `data/plazos.yaml`, que sería texto legal copiado a mano en un fichero de `data/`.

### D17 · `data/plazos.yaml` y `data/cobertura.yaml`

`data/plazos.yaml` lo lee el binario: un solo lector, `plazos.LeerTipos`, que usa también el generador de
`references/recursos_y_plazos.md`. `data/cobertura.yaml` solo lo lee la skill: va con su esquema
`schemas/cobertura.yaml.json` y su lector en `internal/skills`, como `normas.yaml` y `jerarquia.yaml`. Nueve entradas
(contracts/datos-de-plazos.md §3): las cuatro letras de la DA 1ª.2, con la c repartida por materias, y los arts. 112.2 y
112.4; llevan norma la aplicación de los tributos (LGT, art. 97) y los sancionadores de tráfico (texto refundido, art.
1). Ninguna más lleva norma: el plan no cita de lo grabado el precepto que haría aplicable otra (FR-031; supuesto del
spec). Rechazado: un esquema JSON para `data/plazos.yaml` además de su lector, dos definiciones del mismo fichero.

### D18 · El formato de eval

Dos claves y una forma de comando (contracts/evals-legal-core.md §1): `plazo` (`fecha` `AAAA-MM-DD`, `avisos` con los
códigos de `plazos`, `razonamiento`), `no_cubierto` (`norma`) y el comando `{applet: plazos, verbo: calcular, tipo}`.
Los avisos van dentro de `plazo` y no en `avisos`, que es el vocabulario de vigencia del BOE (`boe.CodigosDeAviso`) y
su enumerado lo ata un test al binario. El enumerado de `plazo.avisos` lo ata otro test a `plazos.CodigosDeAviso`.

### D19 · Las formas fijas de la respuesta

- `⚠ PLAZO: <AAAA-MM-DD> — <tipo> [<…>, <identificador>, bloque <bloque>]`: la etiqueta, la fecha justo detrás de los
  dos puntos y al menos una cita en la misma línea.
- `⚠ NO CUBIERTO: <procedimiento> … (<identificador>) …`: la etiqueta y, en la misma línea, el identificador de la norma
  que lo rige o el de la LPAC.
- Los avisos: la línea `texto` de cada uno, tal cual.

Las comprueba `internal/evals` con la gramática de `patronDeEtiqueta` (V18).

### D20 · Los umbrales

`plazo_sin_fecha:<modelo>:<modo>` y `cobertura_sin_declarar:<modelo>:<modo>`, con `<=` 0 y `decide: true`, detrás de
`documento_sin_cotejar` en el orden de cada modo (V19). Una skill tiene umbrales de sus respuestas si tiene juez o
alguna eval declara `sentencias`, `plazo` o `no_cubierto`; ningún nombre de skill decide. Con el plan del job,
`legal-core` lleva seis: `sin_activar`, `plazo_sin_fecha` y `cobertura_sin_declarar` del modo orden y los mismos del
modo herramienta. El motivo de uno que no se cumple es el de todo umbral que decide, seguido de cada sesión que cuenta
por su nombre, en orden de sesión, separadas por «; », como `documento_sin_cotejar`. Contracts/evals-legal-core.md §3.

### D21 · `legal-core` v1

Contracts/skill-legal-core.md y el borrador de contracts/SKILL-legal-core.prototipo.md. `kitlegal-applets: territorio
plazos`; `kitlegal-referencias` gana `recursos_y_plazos` y `cobertura_por_materia`. El precepto lo lee la skill con
`boe articulo` por medio de `boe-legislacion`, que es la que sabe pedirlo y citarlo; `legal-core` no lleva la tabla de
`boe` (FR-046 le da solo la de `territorio` y `plazos`), y por eso las evals de plazo no dicen que `boe-legislacion` no
se active. Rechazada: añadir `boe` a `kitlegal-applets`, que traería su tabla entera a `SKILL.md` y la duplicaría.

### D22 · Los property tests

Sobre una rejilla exhaustiva y determinista —cada fecha de inicio de 2025 a 2027, cada N de 1 a 40 y cada pareja de
un juego de municipios sintéticos con festivos en días distintos— en lugar de `testing/quick` (congelado y con
`math/rand`, V23) o `pgregory.net/rapid` (fuera de la lista de dependencias). Cubre más casos que una muestra y no
depende de una semilla.

### D23 · Dónde vive la tabla de casos

En `internal/app/plazos_casos_test.go` (`TestCasosDePlazos`): llama al applet como lo llama la skill, con los
municipios por su nombre, y con los ficheros empotrados; el caso de Brunete antes de su corrección compone el applet
con los festivos que deriva `internal/source/bocm` de la resolución principal sola. Rechazado: el dominio, cuyos tests
no pueden importar `data` ni los adaptadores (V16).

### D24 · `TestDefinicionDelJob` y FR-058

FR-058 dice que el hito no cambia `.github/workflows/evals.yml` ni los tests que leen su concurrencia. El hito no
cambia ni la definición ni la concurrencia; pero `TestDefinicionDelJob` lleva escrito el peor caso de `legal-core`
con cuatro evals, y FR-050 hace que sea de nueve (V15). Se cambia solo ese literal y las líneas de los dos subtests que
lo incluyen; ningún valor de concurrencia, ningún tope ni ninguna entrada de la matriz. Es el control que FR-050
nombra. Anotado en `gates/supuestos.md` y en Complexity Tracking.

### D25 · Sin esquema para `data/festivos/`

Los dos ficheros los escribe el código y el test de derivación los compara byte a byte con lo que sale de lo grabado;
el dominio solo los lee, y lo que no se puede leer sigue la regla genérica (defecto `inesperado`, código 1). Ni un
esquema JSON ni reglas de carga propias: un fichero distinto de lo derivado solo lo crea una edición a mano, que el
test de derivación ya rechaza.

### D26 · Las preguntas de las evals de cobertura dan el municipio

(c) y (d) dicen dónde vive la persona y la fecha de la notificación: la pregunta trae todo lo que `plazos calcular`
necesitaría y tienta a pedirlo y a dar una fecha, que es lo que el umbral de la cobertura vigila; la skill declara el
asunto en el paso 7, antes de preguntar nada, sin usar ni el municipio ni la fecha (paso 1 de FR-040). Sin ellos, la
pregunta de SC-003, la skill no tiene nada más que preguntar y responde igual. (e) nombra la sede del juzgado por su
municipio.

### D27 · Los guiones de aceptación solo usan días decisivos ciertos

Se congelan antes de grabar: sus fechas solo dependen del 9 y el 12 de octubre de 2026 (local de Leganés leída del
BOCM y nacional no sustituible), del 15 de septiembre y el 28 de diciembre de 2026 y el 14 de enero de 2027 (días
laborables sin fiesta) y del 1 de enero de 2025 (nacional no sustituible). Los demás casos van a la tabla, que se
escribe con `data/festivos/` ya derivado (S3).

### D28 · Ninguna duración máxima

FR-012 no fija una; la salida está acotada sea cual sea la duración (festivos solo de los años cargados, `cobertura`
por tramos, un aviso por ámbito y otro por cada festivo de una parte en los años cargados), y el cómputo es lineal en
la duración. Una duración que no cabe en un entero es «no es un entero positivo».

### D29 · Sin `comunidad` en los festivos saltados

FR-001 pide de cada festivo su fecha, su ámbito, su municipio si es local y su `source`. La comunidad de un autonómico
la sabe la skill por `municipios` y `territorio`; añadirla es una clave que ningún requisito pide.

### D30 · El caso de la entidad menor y el de antes de la corrección

El de la entidad menor es un plazo de El Vellón cuyo último día es el 9 de octubre de 2026, fiesta de El Espartal y no
de El Vellón: cuenta como hábil. El de Brunete usa la derivación con solo `BOCM-20251212-34` para 2026: Brunete queda
`no-comunicado` y el 14 de septiembre, que es su fiesta según la corrección, cuenta como hábil y avisa.

### D31 · Reparación del cierre: la línea del plazo con varios preceptos, la materia que no se dice y los municipios

La medición del cierre sobre `3f14e3a` (job de evals `39168409746`, trabajo `evals (legal-core)`) salió en rojo por
los dos umbrales `plazo_sin_fecha:claude-sonnet-5-5`: 1 de 9 en el modo orden y 2 de 9 en el modo herramienta, las tres
respuestas de `09-plazo-contencioso-agosto`, que además pasó 1 de 3 en orden y 0 de 3 en herramienta con el umbral 2.
Lo que el job dejó (`gates/evals/legal-core.json`) enseña tres causas, todas de la skill, ninguna de la eval ni del
applet, que dio en las ocho sesiones que lo pidieron el 2026-09-15 con `a46` y `a128` en `data.preceptos`:

- **Los dos preceptos en unos mismos corchetes.** Las tres sesiones del modo herramienta pusieron la cita del
  contencioso como `[Ley 29/1998 (LJCA), art. 46.1 y 128.2, BOE-A-1998-16718, bloques a46 y a128]` (dos veces) o
  `[…, bloque a46; art. 128.2, BOE-A-1998-16718, bloque a128]`: `formaDeCita` solo lee los corchetes que terminan en
  `<identificador>, bloque <id>]`, así que las dos primeras no llevan ninguna cita —la línea `⚠ PLAZO:` no cuenta sin
  cita en su línea— y la tercera solo la de `a128`. La plantilla de «Cómo se presenta un plazo» enseñaba una sola cita
  y el paso 11 decía «con su tipo y su cita»; con dos preceptos, el modelo las fundió. La skill dice ahora que cada
  precepto de `data.preceptos` va en sus propios corchetes, en la misma línea, y enseña la línea del contencioso con
  las dos citas. **Rechazada**: ampliar `formaDeCita` para leer «bloques a46 y a128»: la forma es la del contrato de
  `boe-legislacion` §3 («una cita por bloque»), un umbral en rojo no se arregla en lo que mide (ADR 0029) y el binario
  no interpreta la redacción de una respuesta.
- **La materia que la conversación no dice.** La sesión 03 del modo orden no consultó nada y se detuvo a preguntar la
  materia, leyendo el «averigua» del paso 7 como permiso para preguntarla, y afirmó de memoria que el art. 46.1 LJCA
  «cuenta de forma especial en agosto». FR-040.1 ordena comprobar la cobertura antes de preguntar municipios o fecha,
  no preguntar la materia, y la historia (e) y la eval 09 no la dan y esperan el plazo. El paso 7 dice ahora que, sin
  materia en la conversación, el asunto no cae en ninguna entrada: no se pregunta ni se detiene la respuesta, se sigue
  con el 8 y se avisa al responder de que la fecha no vale para lo que la ley regula aparte, con la referencia.
  **Rechazada**: preguntarla siempre; dejaría sin fecha la pregunta de (e) tal como la hace quien la trae.
- **Madrid sin resolver.** La sesión 01 del modo orden pasó «Madrid» a `plazos calcular` sin `territorio resolver`,
  porque `plazos` acepta un nombre; el paso 8 ya mandaba resolverlos por los pasos 1 a 4, pero nada lo hacía
  condición de la llamada. El paso 9 dice ahora que a `plazos` solo llegan municipios que `territorio` resolvió en la
  conversación. **Rechazada**: que `plazos` rechace un nombre y exija el código INE; cambiaría la gramática de D2 y la
  aceptación congelada, y el paso 8 ya es el que da el territorio.

Contracts/skill-legal-core.md (filas 7, 9 y 11 y §3), quickstart §13 (ocho líneas `⚠`) y el CHANGELOG lo dicen igual.
Ni `09-plazo-contencioso-agosto.yaml`, ni `formaDeCita`, ni los umbrales cambian; el efecto lo mide el job en la
medición siguiente, y esta sesión no lo ha medido.
