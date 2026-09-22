# Research: H6 · `territorio` + skill `legal-core` v0

**Modo**: desatendido. Cada decisión se tomó con la sección «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y queda aquí con su alternativa rechazada y su motivo (criterio 3). Ninguna afecta a
alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada, de modo que no procede escalar
(criterio 4).

Tres bloques: la **tabla de verificación** (V1-V47), lo comprobado contra el código o la documentación disponibles en
local, con la orden o el `fichero:línea` de la comprobación; las **decisiones** (D1-D29); y los **supuestos no
verificados** (S1-S9), que son lo que no se puede comprobar sin red, sin la plataforma o sin el fichero que la tarea
`[datos]` genera.

## Tabla de verificación

Todo lo de esta tabla se comprobó en este repositorio, en esta máquina y sin red. «Orden» es lo que se ejecutó;
`fichero:línea`, lo que se leyó.

| # | Afirmación | Comprobado en |
|---|---|---|
| V1 | Go es `1.27.0` con `toolchain go1.27.1`, y el `Makefile` exporta esa cadena de herramientas a todas las recetas | `go.mod:3,5`; `Makefile:14-20` |
| V2 | Un patrón `//go:embed` se interpreta **relativo al directorio del paquete**, no puede contener `..`, no sigue enlaces simbólicos, no casa ficheros fuera del módulo, y si no casa nada **falla la compilación**; solo vale en variables de ámbito de paquete de tipo `string`, `[]byte` o `embed.FS` | orden `go doc embed` |
| V3 | Un objetivo de fuzz ejecuta su **corpus semilla** —`F.Add` y el contenido de `testdata/fuzz/<NombreDelFuzz>`— en un `go test` normal, sin `-fuzz` | orden `go doc testing.F` |
| V4 | De 124 palabras del vocabulario de H6 pasadas por el `misspell` del repositorio, solo marca tres: `aspectos`→`aspects`, `configuracion`→`configuration` y `autonomos`→`autonomous`. **No** marca `municipio(s)`, `provincia(s)`, `comunidad(es)`, `territorio(s)`, `boletin(es)`, `cobertura`, `candidato(s)`, `foral(es)`, `regimen`, `jerarquia`, `ambiguo(s)`, `uniprovincial`, `autonomica`, `ayuntamiento`, `digito`, `relacion`, `inscripcion`, `derivacion`, `aspecto` (singular), `correspondencia`, `muestra`, `evidencia`, `declarado`, `nivel(es)`, `reglas`, `coincidencia(s)` | dos sondas de un fichero por sonda en `internal/core/`, una palabra por línea (la deduplicación por línea de `.golangci.yml` enmascararía dos en la misma), con la orden `go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./internal/core/`; ficheros borrados tras cada sonda |
| V5 | `compruebaDominioPuro` **solo expande paquetes del propio módulo** (`g.esDelModulo(importacion)`), de modo que una dependencia externa no propaga sus importaciones al dominio: `internal/core/x → go.yaml.in/yaml/v3` no arrastra el `io` que esa biblioteca usa por dentro | `internal/arch_test.go:460-498`, en particular `:491` |
| V6 | La comparación de paquete denegado es **por componente de ruta** (`primerPrefijo`/`cuelgaDe`), así que el `io` denegado al dominio deniega también `io/fs`; lo mismo hace `depguard` con `pkg: io` | `internal/arch_test.go:480`, `:693-695`; `.golangci.yml:117` |
| V7 | La lista `core` de `depguard` deniega al dominio `os`, `io`, `log`, `log/slog`, `net/http`, `database/sql` y los ocho paquetes internos, y con `run.tests: true` alcanza también a los `_test.go` del dominio: un test de `internal/core/**` **no puede importar `internal/cli`** | `.golangci.yml:10`, `:88-122` |
| V8 | `TestEsquemasCubrenTodosLosVerbos` exige que **cada fichero de la tabla exista** y que **cada verbo del registro de producción tenga su parte en exactamente uno**; el nombre de la parte se compone con el applet cableado (`publicadaEn["boe "+verbo]`) y la tabla se indexa por posición (`ficherosDeEsquemas[0], [1]`) | `internal/app/esquemas_test.go:289-368`, `:398-420`, `:302`, `:412` |
| V9 | `clasificar_datos` **solo pausa** si la tarea lleva `[datos]` **y** su diff toca `(^|/)testdata/` o `^schemas/`: un fichero existente ahí, o uno nuevo bajo `testdata/*`, `internal/source/*` o `schemas/*`. Un fichero nuevo bajo `data/` **no** provoca pausa | `.specify/workflows/hito/workflow.yml:928-946` |
| V10 | `precheck_tasks` **rechaza el hito** si una línea de tarea nombra `testdata/` o `schemas/` sin la etiqueta `[datos]`, y rechaza una tarea sin ninguna ruta declarable | `.specify/workflows/hito/workflow.yml:468-471` |
| V11 | `modulosDelBinario` es una lista escrita a mano que falla en cuanto el binario enlaza un módulo nuevo; hoy incluye `go.yaml.in/yaml/v4` (llega por `github.com/pb33f/ordered-map/v2`) y **no** incluye `go.yaml.in/yaml/v3` | `internal/arch_test.go:120-167`, `:214-243` |
| V12 | `go.yaml.in/yaml/v3` ofrece `Decoder.KnownFields(bool)`, que rechaza claves que no son campos del `struct` | orden `go doc go.yaml.in/yaml/v3.Decoder`, `go doc go.yaml.in/yaml/v3.Decoder.KnownFields` |
| V13 | Convención de los tipos de `data`: **ninguna etiqueta lleva `omitempty`**, todo campo se emite siempre, un valor que la fuente no da va como cadena vacía y una lista sin elementos va como `[]`, nunca `null` | `internal/source/boe/datos.go:10-15` |
| V14 | Las restricciones formales del esquema salen de etiquetas `jsonschema:"…"` sobre el tipo de `data`: `enum=`, `pattern=`, `minLength=`, `format=` | `internal/source/boe/datos.go:49,65,109`; `internal/core/schema/sobre.go:32,34,42`; `internal/core/schema/error.go:79` |
| V15 | Con `--dry-run` **el applet se ejecuta igual** (la bandera viaja en el contexto y la honra cada capa con efectos), el kernel **no emite sobre** y escribe la descripción en la salida de error; el código sigue siendo el que corresponda al resultado | `internal/app/main.go:357-376`, `:402-424`; `docs/ADR/0011-ensayo-por-capas.md` §Decisión |
| V16 | `Procedencia.FechaConsulta` con valor cero significa «la fecha la pone el reloj del montador»; con valor, el montador usa **la que declara quien consultó** | `docs/ADR/0015-puerto-source-y-fecha-de-consulta.md` §La fecha de consulta; `internal/cli/sobre.go` (`Montador`) |
| V17 | El mensaje del error del applet llega **literal** al `mensaje` del sobre de fallo (`schema.DatosError{Clase: Clasificar(err), Mensaje: err.Error()}`), y el sobre de fallo solo se escribe con `--json`; el mensaje va siempre a la salida de error | `internal/cli/sobre.go:100-122`, `:160-168`; `internal/core/schema/error.go:74-80` |
| V18 | `Clasificar` mira primero los cinco sentinelas del kernel con `errors.Is` y después `errors.As` a `schema.ConClase`; `CodigoSalida` es el único punto de traducción y `codigoDeClase` es un `switch` sin `default` que `exhaustive` vigila | `internal/cli/errors.go:28-59`, `:90-123`, `:136-152` |
| V19 | Los umbrales de cobertura son bloqueantes y sin exclusiones: global 70 %, `internal/core/**` 85 %, `internal/cli/**` 90 % | `codecov.yml:8-11,17-20,38-45,54-61` |
| V20 | El único `//go:embed` del repositorio embebe un subdirectorio **del propio paquete** (`internal/cache/migraciones/`); no hay ningún paquete Go en la raíz del módulo | `internal/cache/migraciones.go:20-25`; `ls *.go` vacío |
| V21 | El descubrimiento de skills es por sistema de ficheros (`os.ReadDir` de `skills/`), y lo que cada skill declara —applets y referencias generadas— vive en `metadata.kitlegal-applets` y `metadata.kitlegal-referencias` de su frontmatter | `internal/skills/skill.go:57-83`; `internal/skills/frontmatter.go:42-47`, `:104-126` |
| V22 | La generación de una referencia está cableada en un `switch` por nombre, y el nombre declarado **tiene que ser también** el del fichero de datos (`data/<nombre>.yaml`) y el del generado (`references/<nombre>.md`) | `internal/skills/sincronia.go:343-350`, `:316-330`; `internal/skills/frontmatter.go:509-537` |
| V23 | La cabecera «generado desde …, no editar», el título y las columnas de la referencia son constantes específicas de normas, no parámetros | `internal/skills/referencias.go:12-37`, `:57-75` |
| V24 | La tabla de comandos se genera desde `--describe` para los applets que la skill declara; **no hace falta tocar código** para una skill nueva con otro applet, y todos los verbos de una tabla deben compartir sobre y banderas globales | `internal/skills/comandos.go:18-32`, `:488-540`; `internal/app/skills_test.go:761-811` |
| V25 | `TestSkillsDelRepositorio` recorre **todas** las skills en los controles de deriva, de normas nombradas y de instrucciones de evals; sus casos negativos están cableados a `boe-legislacion` y no ejercitarían nada de una skill nueva | `internal/app/skills_test.go:67-128`, `:576-604`, `:648-738`, `:910` |
| V26 | `TestEvalsDelRepositorio/formato` descubre **toda** carpeta de `evals/` y le aplica el formato común; las reglas del conjunto, en cambio, solo se aplican a `evals/boe-legislacion` | `internal/evals/conjunto_test.go:627-641`, `:710-733`, `:643-684` |
| V27 | Las reglas del conjunto viven en una lista de reglas recorrida por una función con el nombre de la skill (`ComprobarConjuntoDeBoeLegislacion`), con sus constantes en el mismo fichero | `internal/evals/conjunto.go:134-190` |
| V28 | El formato de eval valida **en carga** contra `schemas/eval.yaml.json`; la exclusividad de las tres variantes de `comandos` y la regla «activa ⇒ `comandos` + `citas`» viven en el esquema, no en Go, y el `struct` es un plano con todos los campos | `internal/evals/formato.go:15,19-87,117-131`; `schemas/eval.yaml.json:13-34` |
| V29 | La variante de un comando esperado se discrimina **por su verbo**, con un `switch` de tres ramas repetido en tres sitios, y la rama `default` supone «consulta de una norma del BOE» | `internal/evals/juzgar.go:348-355`, `:445-454`; `internal/evals/consultas.go:88-102` |
| V30 | Que la respuesta pasó por el binario se comprueba con la **traza del sistema** (`strace`) y el registro de producción, no con el texto de la respuesta; las citas y los avisos se extraen de la respuesta por su forma fija | `internal/evals/trazas.go:213-246`, `:300-360`; `internal/evals/citas.go:13-14`; `internal/evals/avisos.go:48-83` |
| V31 | El umbral «2 de 3» es del flujo (`REPETICIONES_DE_EVALS: 3`, `UMBRAL_DE_EVALS: 2`), y el job evalúa **una sola skill por ejecución** (`SKILL_EVALUADA` escalar). El filtro de rutas es un `case` de shell con los patrones `skills/*`, `evals/*`, `data/*`, `internal/source/boe/*`, `internal/cli/*`, `internal/evals/*`, `scripts/evals.sh`, `.github/workflows/evals.yml`, `schemas/eval.yaml.json` y `Makefile`: **`data/*` ya cubre `data/territorio/…`** —en un patrón de `case`, `*` casa también las barras—, y **`internal/core/*` no está** | `.github/workflows/evals.yml:49-56`, `:76-86` |
| V32 | `TestIdentificadoresDeLasNormas` exige, para **cada** norma de `data/normas.yaml`, exactamente una entrada del manifiesto `testdata/evals/grabaciones.json` que la resuelva por prefijo de título sobre una búsqueda grabada | `internal/evals/grabaciones_test.go:301-358`, `:450-470`, `:508+` |
| V33 | `scripts/verify-sources.sh` ejecuta un único test etiquetado con un solo caso (`boe articulo`) y no entra en `make ci` | `scripts/verify-sources.sh:22`; `internal/app/fuentes_red_test.go`; `Makefile:107-109,162` |
| V34 | La lista de applets del binario está fijada en tres sitios que fallarán al registrar uno nuevo | `internal/app/registro_test.go:240,244`; `cmd/kitlegal/main_test.go:18`; `internal/app/testdata/script/argumentos.txtar:24,30` |
| V35 | El entorno del e2e da `$KITLEGAL_BIN` (ruta absoluta), el binario también en el `PATH`, `$KITLEGAL_CACHE_DIR`, copia de las grabaciones y una orden propia `cronometra <máximo> <programa> …`; los códigos de salida exactos se comprueban con `exec sh -c '…; test $? -eq N'` y el contenido con expresiones ancladas `\A…\z` | `internal/app/e2e_test.go:38-70`, `:195-232`, `:246-296`; `internal/app/testdata/script/boe-codigos.txtar:13-52` |
| V36 | Un applet que no consulta fuentes firma en el espacio reservado (`kitlegal.<nombre>` / `kitlegal:applet/<nombre>`), y ese literal **no puede** aparecer en código de producción bajo `internal/source/` | `docs/ADR/0006-sobre-de-salida-y-huella.md` §El espacio de nombres reservado; `internal/app/ejemplo/echo.go:15-18`; `internal/arch_test.go:263-304` |
| V37 | La validación de un YAML de `data/` contra su esquema se hace con `skills.ValidarDocumentoYAML[T]`: un solo documento, **rechazo de claves repetidas**, normalización a tipos JSON y validación con `santhosh-tekuri/jsonschema/v6` con `AssertFormat()`; el esquema se localiza por ruta relativa al paquete y se compila una vez con `sync.OnceValues` | `internal/skills/esquemas.go:30-49`, `:143-172`; `internal/skills/normas.go:19,87-126` |
| V38 | `TestGramaticasCoincidenConBoe` ata el `pattern` de un esquema de datos al validador de Go que analiza lo mismo, comprobando que aceptan y rechazan el mismo conjunto | `internal/evals/formato_test.go:416-470` |
| V39 | `misspell.ignore-rules` tiene hoy **20** entradas, cada una con su comentario de motivo | orden `awk '/ignore-rules:/{f=1;next} f&&/^        - /{c++; print $2} f&&/^      [a-z]/{exit}' .golangci.yml \| wc -l` → `20` |
| V40 | Un argumento de posición sin `optional:""` se describe como `required` **y** `cli.Analizar` lo exige **antes** de la decisión de describir: `--describe` sin el posicional termina en 2 y no emite esquema; con él, emite el esquema del verbo | `internal/cli/describe.go:475-491` (`exigidoPorLaGramatica`); `internal/app/main.go:301-320` (`Analizar` antes del `switch analisis.Decision`); órdenes `go run ./cmd/kitlegal boe indice --describe` → «argumentos inválidos: expected "<norma>"», exit 2, y `go run ./cmd/kitlegal boe indice BOE-A-2015-10565 --describe` → `boe indice` / `norma`; así lo invocan los guiones del e2e (`internal/app/testdata/script/boe-verbos.txtar:16-39`) y el quickstart de H4 (`specs/005-h4-applet-boe-puerto/quickstart.md:285`) |
| V41 | Con el gancho de `rtk` activo, las dos formas del quickstart que leen la ausencia de salida imprimen **solo su rótulo**: `rtk proxy git status --porcelain \| rtk proxy grep -vE '^.. specs/008-h6-territorio-skill-legal/' ; echo "fin del estado"` → `fin del estado`; `rtk proxy git diff --stat main -- evals/boe-legislacion ; echo "fin del diff"` → `fin del diff`. Son sondas **positivas**: la forma con `\| wc -l` y sin `rtk proxy` cuenta 1, porque la orden reescrita emite una línea vacía | las dos órdenes, ejecutadas en este repositorio |
| V42 | Los tests de repositorio de `internal/skills` leen el fichero real de `data/` **por ruta relativa al directorio del paquete** (`../../data/normas.yaml`), como el esquema (V37), y **pueden importar el dominio y los adaptadores**: la lista `core` de `depguard` solo acota `**/internal/core/**`, de modo que `internal/skills` no tiene denegado nada. Es el patrón de `TestNormasDelRepositorio`: leer con `os.ReadFile`, validar con el lector común y comprobar sobre lo leído | `internal/skills/normas_test.go:332-349` y sus importaciones `:17-21` (`internal/core/schema`, `internal/evals`, `internal/httpx`, `internal/source/boe`); `.golangci.yml:93` |
| V43 | `make test` mide cobertura **por paquete**: `go test -race -shuffle=on -coverprofile=coverage.out ./...`, **sin `-coverpkg`**, así que un test de `internal/skills` que ejerce código de `internal/core/territorio` **no cuenta** para el umbral de `internal/core/**` | `Makefile:69`; `codecov.yml:8-11` |
| V44 | Cambiar el `then` de `schemas/eval.yaml.json` a `required: [comandos]` + `anyOf` **cambia el error que ve `LeerEval`**: `incumplimientosDe` desciende hasta las hojas del árbol del validador —un nodo `anyOf` tiene causas, así que no es hoja—, `unir` las junta con `errors.Join` y `TestLeerEval` compara con `require.EqualError`. El caso `positiva-sin-citas` deja de ver «`missing property 'citas'`» a secas y pasa a ver las hojas de las dos ramas del `anyOf`; el texto de un `anyOf` fallido es «`'anyOf' failed`» | `internal/skills/esquemas.go:405-415`, `:521-533`; `internal/evals/formato_test.go:131-133`, `:274`; `schemas/eval.yaml.json:32`; `github.com/santhosh-tekuri/jsonschema/v6@v6.0.3/kind/kind.go:80-84` |
| V45 | `skills-check` no descubre tests: ejecuta por nombre cuatro tests de repositorio (`TestSkillsDelRepositorio`, `TestNormasDelRepositorio`, `TestEvalsDelRepositorio`, `TestIdentificadoresDeLasNormas`) sobre `./internal/app/`, `./internal/skills/` y `./internal/evals/`, de modo que **cada test de repositorio nuevo exige tocar su expresión `-run`** en la tarea que lo crea | `Makefile:103-105` |
| V46 | `ValidarDocumentoYAML[T]` decodifica con `raiz.Decode(&leido)`, **sin `KnownFields`**: lo que rechaza un campo no declarado es el `additionalProperties: false` del esquema, no el tipo Go. Añadir una propiedad opcional al esquema antes de que exista el campo del `struct` no rompe nada | `internal/skills/esquemas.go:143-172`; `schemas/normas.yaml.json`; `internal/skills/normas_test.go` (caso `campo no declarado: vertical`) |
| V47 | El `enum` de `rango` de `schemas/normas.yaml.json` está atado por **igualdad**, no por inclusión, al conjunto de `rango.texto` de todos los resultados de todas las búsquedas de todos los conjuntos grabados: `assert.Equal(rangosDeLasBusquedasGrabadas(t, evals.UnionDeGrabaciones()), rangosDelEsquema(t, contenido))`, ambos ordenados. Hoy ese `enum` son nueve valores —Constitución, Decreto Foral Legislativo, Ley, Ley Foral, Orden, Real Decreto, Real Decreto Legislativo, Real Decreto-ley, Resolución— **sin «Ley Orgánica»**, así que grabar una búsqueda que devuelva una ley orgánica y no ampliar el `enum` (o al revés) deja `make ci` en rojo | `internal/skills/normas_test.go:375-380` (subtest), `:415` (`rangosDelEsquema`), `:455-475` (`rangosDeLasBusquedasGrabadas`), `:513-556` (`rangosDeLaBusqueda`); `schemas/normas.yaml.json:23-34`; `internal/evals/grabaciones.go:31` |

## Decisiones

### D1 · Orden del hito: de fuera adentro al planificar, de dentro afuera al implementar

**Decisión.** Se planifica desde la skill (`legal-core` y sus evals) hacia la herramienta, y se implementa
`core/ids → core/territorio → data embebida → applet → esquema publicado → e2e → skill → evals → job`
(constitución, «Flujo de trabajo» 2; `docs/ROADMAP.md` §6.1). Las evals de `legal-core` se escriben **antes** que
`SKILL.md` (FR-083).

**Alternativas.** *Empezar por el applet porque es lo más grande*: deja la skill —el producto— al final y sin evals
que la midan, contra el principio VIII. *Empezar por la skill entera*: no hay applet que invocar, y su tabla de
comandos se genera desde `--describe`.

### D2 · Dónde vive cada pieza

**Decisión.**

| Pieza | Paquete | Por qué |
|---|---|---|
| Código INE y DIR3 | `internal/core/ids` | Lo nombra `docs/ROADMAP.md` §2 y el alcance de H6; es cálculo puro sobre cadenas |
| Municipio, provincia, comunidad, régimen, boletines, cobertura y la resolución | `internal/core/territorio` | Lo nombra `docs/ROADMAP.md` §2; es dominio sin entrada ni salida |
| Los ficheros congelados, empaquetados | paquete `data` (fichero `data/datos.go`) | `//go:embed` no sube de directorio (V2) y los ficheros tienen que estar en `data/territorio/` (FR-041, FR-042) |
| El applet y su composición | `internal/app/territorio.go` | Es donde vive el contrato `Applet` (ADR 0005) |

**Alternativas.** *Mover los datos junto al paquete que los embebe* (`internal/core/territorio/data/`): es lo que hace
`internal/cache/migraciones` (V20) y no exigiría paquete nuevo, pero contradice FR-041 y FR-042, que fijan
`data/territorio/` como ubicación versionada, y `docs/ROADMAP.md` §2, que hace de `data/` la fuente de verdad de las
dos capas. *Un paquete en la raíz del módulo* (`package kitlegal` en `embed.go`): funciona, pero ocupa el nombre de
importación público del módulo, que la constitución reserva a `pkg/legalkit`, y mete código Go en la raíz, hoy sin
ninguno (V20). *Un enlace simbólico de `data/territorio` dentro del paquete*: `go:embed` no sigue enlaces (V2).
*Generar un fichero Go con los 8.132 municipios como literales*: elimina el análisis en ejecución y convierte la
ausencia en error de compilación, pero mete 8.000 líneas generadas en el árbol, exige su propio control de deriva y
contradice la clarificación del spec (sesión 2026-09-20, Q4), que fija `//go:embed` y la validación contra el esquema
en test.

### D3 · El dominio recibe bytes, no un sistema de ficheros

**Decisión.** `internal/core/territorio` expone `Cargar(Fuentes) (*Registro, error)`, donde `Fuentes` es un valor puro
con los contenidos ya leídos (`Municipios []byte`, `DIR3 []byte`, `Estado []byte`, `Comunidades map[string][]byte`).
El paquete `data` embebe y expone esos bytes; `internal/app` los pasa.

**Motivo.** `io/fs` está denegado al dominio: la lista `core` de `depguard` deniega `io` y la comparación es por
componente, así que `io/fs` cuelga de `io` (V6, V7); `compruebaDominioPuro` aplica la misma regla sobre el grafo real
(V5). Además, recibir bytes deja el dominio comprobable con datos sintéticos, sin fixtures ni ficheros.

**Alternativas.** *`fs.FS` como puerto*: es lo natural en Go y lo que usa `internal/cache`, pero rompe R1 (V6, V7) y
obligaría a sacar `territorio` de `internal/core`, contra `docs/ROADMAP.md` §2. *Que el dominio importe el paquete
`data`*: pasa los linters —`embed` no está denegado—, pero ata el dominio a un único juego de datos y deja sus tests
dependiendo del contenido real del repositorio.

### D4 · Formato de los ficheros congelados: YAML indexado por código, una línea por fila

**Decisión.** YAML, como el resto de `data/` (`docs/ROADMAP.md` §2, tabla «Configuración por datos»; `CLAUDE.md`,
«Skills sin código»), con un mapa indexado por el código y el valor en forma de flujo, de modo que cada municipio
ocupe **una línea** y un diff señale exactamente lo que cambió (FR-041):

```yaml
fecha: 2026-02-04
source: ine.municipios
municipios:
  "28074": {dc: "5", nombre: "Leganés", provincia: "28", comunidad: "13"}
```

La fila lleva las cinco columnas que FR-040 exige —código, dígito de control, nombre, provincia y comunidad—, y es la
**única forma** del fichero: la misma en [data-model.md](./data-model.md) §2.2, §3.1 y en el contrato de datos. La
`comunidad` de la fila es redundante con la que declara el fichero de la provincia, y esa redundancia se ata en la
integridad de `Cargar` (data-model §2.1, punto 6), donde además se declara cuál de los dos caminos es el autoritativo.

**Motivo.** El mapa por código da unicidad gratis: el lector común rechaza la clave repetida (V37) y el esquema la fija
con `propertyNames.pattern`, igual que `data/normas.yaml`. La forma de flujo mantiene una línea por fila. Desde T029,
además, la carga lee en esta forma las filas de la relación y de la correspondencia sin pasar por el lector de YAML, que
lee entero cualquier fichero escrito de otra manera (S3).

**Alternativas.** *CSV o TSV*: más compacto y el formato natural del origen, pero no se puede validar con JSON Schema,
que es lo que FR-044 exige. *JSON*: se valida igual y se analiza más rápido con la biblioteca estándar, pero
`docs/ROADMAP.md` §2 dice «un territorio o un boletín nuevo es un YAML» y rompería la homogeneidad de `data/`.
*Una lista de objetos en bloque*: legible, pero multiplica por cinco las líneas del fichero y hace ilegible el diff.

### D5 · Cuatro ficheros de datos y cuatro esquemas

**Decisión.**

| Fichero | Qué fija | `source` de sus datos |
|---|---|---|
| `data/territorio/municipios.yaml` | Los 8.132 municipios: código INE, dígito de control, nombre oficial, provincia y comunidad | `ine.municipios` |
| `data/territorio/dir3.yaml` | La correspondencia INE→DIR3, **solo las filas verificadas** | `mpt.rel` |
| `data/territorio/estado.yaml` | Lo nacional que no es por comunidad: el boletín estatal | el propio fichero |
| `data/territorio/comunidades/<código>.yaml` | Una por comunidad y ciudad autónoma (19): nombre, régimen, provincias y, si está configurado, sus boletines | `ine.municipios` (nombres) y el propio fichero (régimen y boletines) |

Esquemas: `schemas/territorio-municipios.yaml.json`, `schemas/territorio-dir3.yaml.json`,
`schemas/territorio-comunidad.yaml.json` —uno solo para los 19 ficheros de comunidad— y
`schemas/territorio-estado.yaml.json`, porque `estado.yaml` tiene forma propia y no cabe en el de comunidad. Son
**cuatro** esquemas nuevos para `data/territorio/`, más `schemas/jerarquia.yaml.json` (FR-067).

**Motivo.** Separar municipios de DIR3 es lo que pide FR-042 y lo que permite que una fila sin DIR3 verificado
signifique exactamente eso (FR-048). Un fichero por comunidad hace que añadir un territorio sea añadir o rellenar un
fichero, sin tocar código ni skills (FR-053), y que el régimen —dato nacional— exista para las 19 aunque solo Madrid
tenga boletines (FR-051, FR-055).

**Alternativas.** *Un solo fichero con todo*: un cambio en la configuración de una comunidad tocaría el fichero de
8.132 municipios. *El DIR3 como columna de `municipios.yaml`*: cumple la letra de FR-042, pero mezcla dos orígenes y
dos fechas de fichero en cada fila y obliga a repetir el `source` fila a fila. *Las provincias en su propio fichero*:
un quinto esquema para 52 filas que solo se leen a través de su comunidad.

### D6 · `fecha_consulta`: la fecha del fichero congelado más antiguo que sostiene la respuesta

**Decisión.** `territorio` declara `Procedencia.FechaConsulta` con la **más antigua** de las fechas de los ficheros de
`data/territorio/` que sostienen ese `data` (el de municipios, el de su comunidad, el del estado y, si la trae, la fila
de DIR3), a medianoche UTC. Cada fichero lleva su `fecha` obligatoria en la raíz.

**Motivo.** Es literalmente la regla de FR-096 de H4, recogida en el ADR 0015: la fecha del sobre es la de la consulta
que produjo el contenido —aquí, la del volcado congelado— y, cuando `data` se sostiene sobre varias, la más antigua,
«para que la cita nunca aparente más frescura que su parte más vieja». Y tiene una consecuencia que el spec exige por
otro lado: la salida deja de depender del reloj, así que **es byte a byte la misma en dos ejecuciones y con
`--offline`** (US1 escenario 4, SC-001), que con el reloj del montador sería imposible.

**Alternativas.** *Dejar la fecha a cero y que la fije el reloj del montador* (lo que hacen los applets de ejemplo):
es lo más simple, pero hace pasar por recién consultado un volcado de hace meses —lo que el ADR 0015 rechaza
explícitamente en su opción c— y rompe la igualdad byte a byte que pide SC-001. *La fecha de compilación del binario*:
no es la fecha de ningún dato y volvería a mentir sobre la antigüedad.

### D7 · `--dry-run` no cambia lo que hace el applet, y no emite sobre

**Decisión.** `territorio` no rellena `Resultado.Ensayo`: no tiene ninguna capa con efectos que describir. Con
`--dry-run`, el applet se ejecuta igual, el kernel escribe su línea en la salida de error y **no emite sobre**, que es
el comportamiento que H1 fijó y el ADR 0011 documenta (V15). Con `--offline`, el applet devuelve exactamente lo mismo
que sin la bandera, byte a byte (D6).

**Motivo.** FR-009 pide que las dos banderas «se comporten sin cambiar la respuesta» y concreta el caso de `--offline`.
Para `--dry-run`, la forma de presentación es una decisión cerrada del kernel (H1, ADR 0011) que un applet no puede
cambiar y que el spec no reabre; lo que H6 garantiza es lo que está en su mano: que el applet haga lo mismo con la
bandera y sin ella. Queda escrito aquí para que nadie lo lea como un incumplimiento.

### D8 · `internal/core/ids`: análisis y normalización con tipo, no solo validación

**Decisión.** API del paquete:

```go
type CodigoINE struct{ /* provincia, municipio */ }          // 5 cifras, inmutable
func AnalizarCodigoINE(entrada string) (CodigoINE, error)     // 5 cifras
func AnalizarCodigoINEConDigito(entrada string) (CodigoINE, byte, error) // 6 cifras: código + dígito declarado
func (c CodigoINE) String() string                            // "28074"
func (c CodigoINE) Provincia() string                         // "28"
func (c CodigoINE) ComprobarDigito(declarado, oficial byte) error

type DIR3 struct{ /* … */ }
func AnalizarDIR3(entrada string) (DIR3, error)               // L01PPMMMDC
func DIR3DeAyuntamiento(c CodigoINE, digito byte) DIR3
func (d DIR3) String() string
func (d DIR3) CodigoINE() CodigoINE
func (d DIR3) Digito() byte
```

Los errores son de un tipo propio del paquete que implementa `schema.ConClase` devolviendo `schema.ClaseArgumentos`,
de modo que el kernel los traduce a código 2 sin que el dominio importe `internal/cli` (V7, V18).

**`ComprobarDigito` es un método de `CodigoINE`**, no una función de paquete: el comportamiento vive en el tipo que
solo se construye analizando, y así lo escriben esta decisión, [data-model.md](./data-model.md) §1.1 y el contrato
[identificadores-ine-y-dir3](./contracts/identificadores-ine-y-dir3.md) §1 y §3. Es la forma que exige
`TestSuperficieDeIds`, que compara lo exportado con esa lista «exactamente».

**Un fichero de test por fichero de código, también para el fuzz**: `ine.go` → `ine_test.go` (con `FuzzCodigoINE`),
`dir3.go` → `dir3_test.go` (con `FuzzCodigoDIR3`), `errores.go` → `errores_test.go`; no hay ningún `ids.go` ni, por
tanto, ningún `ids_test.go`. El corpus versionado se indexa por el **nombre del objetivo**
(`testdata/fuzz/FuzzCodigoINE/`, `testdata/fuzz/FuzzCodigoDIR3/`), no por el del fichero que lo declara, así que la
regla no le cuesta nada (V3).

**Motivo.** FR-032 pide normalización idempotente e ida y vuelta estable, que con funciones `Validar*` sueltas —lo que
hace hoy `internal/source/boe/ids.go`— no se puede expresar: hace falta un tipo que solo se construya analizando.

**Alternativas.** *Copiar el patrón de `boe`* (`ValidarCodigoINE(string) error`): más corto y coherente con lo
existente, pero deja la normalización en quien llame y no da la propiedad de ida y vuelta que FR-032 exige.
*Un tipo `string` con métodos*: cualquier cadena sería un código válido por conversión.

### D9 · El dígito de control es un **dato oficial**, no un algoritmo

**Decisión.** `internal/core/ids` analiza la forma (cinco cifras, o seis con el dígito detrás) y **compara** el dígito
declarado con el oficial, que viaja en `data/territorio/municipios.yaml` (columna DC de la relación del INE). No se
implementa ningún algoritmo de cálculo del dígito.

**Motivo.** El dígito de control del INE es un dato publicado, no una función que este proyecto pueda derivar y
verificar sin red; escribir un algoritmo de memoria sería exactamente lo que ADR 0017 prohíbe con el DIR3. Con el
dígito como dato, la comprobación es correcta por construcción para los 8.132 municipios y no hace falta ninguna
hipótesis. El orden de los fallos queda: código **bien formado** —provincia `01`-`52` y municipio `001`-`999`— pero de
un municipio que no está en la relación → 3 (FR-011); entrada de solo cifras que no cumple esa gramática (provincia `00`
o mayor que `52`, municipio `000`, longitud imposible) → 2, porque no llega a ser un código; código de un municipio que
sí está, con dígito distinto del oficial → 2 (FR-012).

**Alternativas.** *Implementar el algoritmo del dígito y verificarlo contra la relación*: daría validación sin
consultar la tabla, pero el algoritmo sería una hipótesis escrita de memoria, y la tabla está embebida de todos modos.
*Ignorar el dígito y aceptar solo cinco cifras*: incumple FR-012 y el caso límite «código con y sin dígito de control».

### D10 · Coincidencia por nombre: pliegue propio verificado contra el corpus

**Decisión.** La normalización de nombres vive en `internal/core/territorio` y es un **pliegue explícito**: minúsculas
ASCII y españolas, sustitución de cada letra con diacrítico por su letra base (`á é í ó ú à è ì ò ù ä ë ï ö ü â ê î ô û
ñ ç ·`…), colapso de espacios y de signos de puntuación separadores. La tabla del pliegue es una constante del paquete,
la función que la aplica se exporta como `territorio.Plegar` —la comparten el registro, el juicio de evals y el control
de corpus (D27)— y **tres tests mecánicos la atan al corpus congelado**: (1) toda runa que aparece en algún nombre
oficial está cubierta por el pliegue o es ASCII —una runa nueva hace fallar `make ci`, no pasa en silencio—; (2) ningún
municipio queda inalcanzable: por su nombre oficial se resuelve él, o se declara una ambigüedad que lo nombra entre sus
candidatos; (3) ningún nombre plegado es solo cifras, porque esa forma la lee la resolución como un código (data-model
§2.6). **Los tres viven en `internal/skills`, no en el dominio** (D27): son los ficheros congelados reales, y el
dominio no puede leerlos.

Las formas alternativas se derivan del nombre oficial, sin caso especial por municipio (FR-024): `Coruña, A` da también
`A Coruña`; `Donostia/San Sebastián` da `Donostia` y `San Sebastián`, y cada una su forma con artículo antepuesto.
Un nombre que corresponde a más de un municipio devuelve candidatos y código 2 (FR-013), también cuando la coincidencia
llega por una forma alternativa: eso es ambigüedad declarada, no colapso.

**Motivo.** `golang.org/x/text/unicode/norm` resolvería la descomposición en una línea, pero **no está en la lista de
dependencias de la constitución §V** y hoy es indirecta (`go.mod:30`); promoverla exigiría justificación en
*Complexity Tracking* para algo que un corpus **cerrado y versionado** permite resolver mejor: con el pliegue propio,
lo que cubre y lo que no es explícito y está probado sobre las 8.132 filas reales, mientras que con NFD la cobertura
sería general pero nadie comprobaría que el resultado no colapsa dos municipios.

**Alternativas.** *`golang.org/x/text/unicode/norm` + `runes.Remove(runes.In(unicode.Mn))`*: dependencia directa nueva
fuera de §V, y seguiría necesitando los mismos tres tests sobre el corpus. *Guardar la forma normalizada en el propio fichero
de datos*: duplica en los datos lo que el código tiene que saber hacer igualmente con la entrada de quien pregunta.
*Comparar sin plegar*: `leganes` no encontraría `Leganés`, contra FR-015.

### D11 · Forma de `data` del verbo `resolver`

**Decisión.** Ocho claves de primer nivel, las ocho de la entrega, cada una con el `source` del dato que la sostiene
(FR-005), y sin `omitempty` (V13): lo que no hay va como cadena vacía y las listas vacías van como `[]`.

```json
{
  "municipio":  {"nombre": "Leganés", "source": "ine.municipios"},
  "codigo_ine": {"codigo": "28074", "digito_de_control": "5", "source": "ine.municipios"},
  "provincia":  {"codigo": "28", "nombre": "Madrid", "source": "ine.codigos-territoriales"},
  "comunidad":  {"codigo": "13", "nombre": "Comunidad de Madrid", "source": "ine.codigos-territoriales"},
  "dir3":       {"codigo": "L01280745", "source": "mpt.rel"},
  "regimen":    {"valor": "comun", "source": "data/territorio/comunidades/13.yaml"},
  "boletines":  [{"nivel": "estatal", "codigo": "BOE", "nombre": "…", "url": "…", "motivo": "", "source": "data/territorio/estado.yaml"}],
  "cobertura":  {"boletin_autonomico": "configurado", "boletin_provincial": "configurado", "dir3": "verificado"}
}
```

`cobertura` tiene **exactamente tres claves**, siempre presentes (FR-020), con vocabularios cerrados:
`configurado|no-configurado` para los dos boletines y `verificado|no-verificado` para el DIR3. **Ningún valor significa
«no existe»**, y eso es lo que hace mecánicamente cierto FR-022: un test exige que el enumerado sea exactamente ese.
Invariante atada en test: `dir3.codigo == ""` ⟺ `dir3.source == ""` ⟺ `cobertura.dir3 == "no-verificado"`.

**Motivo.** El `source` dentro de cada dato sobrevive a que una skill extraiga un trozo, y es la forma en que H7 podrá
emitir `Municipio` y `Organo` con procedencia sin volver a decidir nada (ADR 0014).

**Alternativas.** *Un mapa `fuentes` al final*: menos ruidoso, pero separa el dato de su procedencia justo cuando
alguien lo copia. *Omitir `dir3` cuando no está verificado*: rompe la convención de que todo campo se emite siempre
(V13) y obliga a cada consumidor a distinguir «no está» de «no lo sé».

### D12 · Los boletines no configurados no aparecen, y la cobertura lo dice

**Decisión.** `boletines` lleva **solo** los niveles configurados, y siempre el estatal (FR-008). Fuera del territorio
configurado no hay ninguna entrada autonómica ni provincial, ni nombre, ni código, ni URL (FR-021), y `cobertura` los
declara `no-configurado`. En la Comunidad de Madrid el BOCM figura dos veces —autonómico y provincial— y la entrada
provincial lleva el `motivo` que la propia configuración fija (FR-052).

**Alternativas.** *Emitir la entrada con los campos vacíos*: un consumidor podría leer «boletín provincial: (vacío)»
como «no tiene», que es justo lo que FR-022 prohíbe. *Deducir que en toda comunidad uniprovincial el autonómico hace
de provincial*: lo prohíbe el caso límite del spec; la equivalencia es una decisión por territorio y va en su fichero.

### D13 · Procedencia del sobre y del fallo

**Decisión.** `fuente: "kitlegal.territorio"`, `url: "kitlegal:applet/territorio"` (ADR 0006, fila «applet calculado»;
V36). Los fallos que **decide el applet** —nombre ambiguo, municipio inexistente, código mal formado— viajan con esa
misma procedencia y su fecha (D6); los que decide el kernel antes de llegar al applet —falta el argumento, bandera
desconocida— salen con `kitlegal.cli` / `kitlegal:cli`, como en cualquier otro applet (V17).

**Motivo.** Un sobre de fallo con la procedencia del applet es una «cita negativa útil» (ADR 0006) y hace verificable
en el e2e de qué capa vino cada código.

### D14 · Ambigüedad: los candidatos viajan en el mensaje

**Decisión.** El error de nombre ambiguo lleva en su mensaje todos los candidatos con su código INE y su provincia,
ordenados por código INE, en una forma fija: `<código> <nombre> (<provincia>)`, separados por `; `. El mensaje llega
literal al `mensaje` del sobre de fallo (V17) y a la salida de error.

**Motivo.** El `data` de un sobre de fallo es exactamente `{clase, mensaje}` (ADR 0006): no hay dónde poner una lista
estructurada sin romper el contrato, y FR-014 lo asume explícitamente.

**Alternativas.** *Un sobre de éxito con la lista de candidatos y código 0*: la skill no podría distinguir «este es tu
municipio» de «elige uno», y el spec pide código 2. *Añadir una clave al `data` del fallo*: cambia ADR 0006.

### D15 · El paquete `data` y lo que el binario pasa a enlazar

**Decisión.** `data/datos.go` (paquete `data`) declara los embebidos y nada más: sin lógica, sin lectura de disco.
`internal/app` los inyecta. Como el dominio analiza YAML, el binario pasa a enlazar `go.yaml.in/yaml/v3`, que hay que
añadir a `modulosDelBinario` con su motivo (V11).

**Motivo.** `go.yaml.in/yaml/v3` es la dependencia de YAML fijada por el repositorio desde H5 y ya es directa.

**Alternativas.** *Usar `go.yaml.in/yaml/v4`, que el binario ya enlaza* (V11): no añadiría ningún módulo, pero solo
tiene versiones candidatas y H5 la rechazó por eso mismo (`specs/006-…/plan.md`, *Complexity Tracking*); además
convivirían dos API de YAML en el mismo árbol. *Analizar el YAML en `internal/app` y pasar tipos al dominio*: mueve el
análisis a la raíz de composición, que no es su sitio, y deja al dominio sin poder validar lo que recibe.

### D16 · El esquema publicado del applet y el orden que impone

**Decisión.** `schemas/municipio.json` (el nombre lo fija D25), generado desde `--describe` con la receta existente
(`TestEsquemasPublicados` y su bandera `-actualizar-esquemas`). Antes hay que **parametrizar por applet** la tabla de
`internal/app/esquemas_test.go`, hoy cableada a `boe` en nueve puntos (V8). Y, por la forma de los controles, **el
registro del applet en producción y la publicación de su esquema tienen que ir en la misma tarea**: en cuanto
`territorio resolver` está en el registro, `TestEsquemasCubrenTodosLosVerbos` exige su parte publicada, y el esquema no
se puede generar antes de que el applet exista (V8). Esa tarea lleva `[datos]` y su pausa, en la que una persona revisa
el contrato publicado.

**Alternativas.** *Dos tareas (registrar, luego publicar)*: dejaría `make ci` en rojo entre ellas, contra la regla de
rebanadas verdes. *Publicar el esquema escrito a mano antes del applet*: el contenido tiene que coincidir byte a byte
con lo que `--describe` emite, y corregirlo desde una tarea de código tocaría `schemas/`, que el guardián rechaza
(V10). *Relajar el control de cobertura de esquemas para que tolere un verbo sin publicar*: es arreglar el control en
lugar del código, prohibido por el criterio 1 de la constitución.

### D17 · La tarea `[datos]` de los ficheros congelados también trae su esquema

**Decisión.** Cada tarea `[datos]` que genera un fichero de `data/territorio/` incluye **su esquema** bajo `schemas/`.

**Motivo.** FR-045 exige pausa humana para la generación de esos ficheros, y la pausa **solo se dispara** si el diff de
la tarea toca `testdata/` o `schemas/` (V9): un fichero nuevo bajo `data/` no pausa. Emparejar el dato con su contrato
es además lo natural —se revisan juntos— y no mezcla otro trabajo: ninguna de esas tareas toca código.

**Alternativas.** *Dejar la generación en una tarea sin pausa*: incumple FR-045. *Cambiar `clasificar_datos` para que
vigile `data/`*: es tocar el proceso desde dentro del hito, que el guardián rechaza por alcance.

### D18 · El registro de la verificación del DIR3

**Decisión.** La evidencia de FR-046 y FR-047 se escribe en
`specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md`: una fila por municipio de la muestra con el código
derivado, el código real y de dónde salió ese código real, más la conclusión (regla confirmada o tabla). La fila del
REL de `docs/SOURCES.md` se actualiza con la fecha del volcado dentro de la misma tarea `[datos]` (FR-049).

**Alternativas.** *Un documento en `docs/`*: es evidencia de un hito, no documentación permanente del producto.
*Dentro del propio fichero de datos*: mezclaría la evidencia con el dato y la haría viajar en el binario.

### D19 · La correspondencia entra por regla verificada, no municipio a municipio

**Decisión.** Entra en `dir3.yaml` toda fila del REL cuyo número de inscripción es coherente con el código INE y su
dígito de control oficial; la muestra de FR-046 verifica **la regla**. Un municipio sin fila o con número incoherente
queda fuera y se declara en `cobertura`. Si la muestra revela discrepancias más allá de casos aislados, la derivación
pasa a ser tabla y **eso lo decide una persona en la pausa** (spec, *Clarifications* Q1; ADR 0017, consecuencias).

### D20 · Las dos referencias generadas de `legal-core`

**Decisión.** Se rompe el acoplamiento «nombre de la referencia = nombre del fichero de datos» (V22) sustituyendo el
`switch` por una tabla que declara, por referencia, su fichero de datos y su generador:

| Referencia | Datos | Generador |
|---|---|---|
| `normas` | `data/normas.yaml` | todas las normas (sin cambio) |
| `leyes_vertebrales` | `data/normas.yaml` | solo las marcadas `vertebral: true` |
| `jerarquia_normativa` | `data/jerarquia.yaml` | niveles, boletín de cada nivel y reglas de interpretación |

La cabecera «generado desde `<fichero>`, no editar», el título y las columnas dejan de ser constantes de normas y pasan
a derivarse de esa tabla (V23). `defectosDeLasReferencias` consulta la tabla en lugar de componer `data/<nombre>.yaml`.

**Alternativas.** *Llamar `normas_vertebrales` a la referencia y crear `data/normas_vertebrales.yaml`*: conserva el
acoplamiento sin tocar código, pero duplica las normas en dos ficheros y rompe la fuente única de verdad, contra la
clarificación Q2 del spec y `CLAUDE.md`. *Generar las tres referencias desde un fichero nuevo que las agrupe*: mismo
problema.

### D21 · El formato común de eval crece por donde ya crecía

**Decisión.** Cuatro cambios, compatibles hacia atrás (FR-084):

1. **Cuarta variante de comando** en `schemas/eval.yaml.json`: `{applet, verbo: resolver, municipio}`, como un `$ref`
   más del `oneOf`; en Go, un campo `Municipio` más del `struct` plano (V28).
2. **Discriminación explícita de la variante** en un único sitio (`formaDelComando`), que consumen el juicio, el texto
   del comando y las consultas necesarias, en lugar de tres `switch` por verbo con `default` «consulta de norma»
   (V29). Sin esto, `territorio resolver` caería en el `default` y se juzgaría como una consulta del BOE.
3. **Esperado de territorio**: campo `territorio` con las claves que la respuesta debe declarar (comunidad, provincia,
   boletines y aspectos de cobertura), extraído de la respuesta por su forma fija, como las citas y los avisos (V30),
   y repartido en encontrados y ausentes. Un esperado ausente impide que la eval pase.
4. **La regla del esperado verificable**: el `then` del esquema pasa de `required: [comandos, citas]` a
   `required: [comandos]` más `anyOf: [required citas, required territorio]`; `citas` sigue siendo obligatoria en toda
   eval que afirme contenido de norma, porque quien lo afirma es quien declara `citas`. Este cambio **arrastra
   consigo** la expectativa de un caso de `TestLeerEval`, que compara el mensaje literal del defecto (V44): las dos
   cosas van en la misma tarea `[datos]` (D28).

**Motivo.** El formato es común a todas las skills desde H5, y el spec fija que se extienda en lugar de crear uno
propio (*Fuera de alcance*).

**Alternativas.** *Un formato propio de `legal-core`*: excluido por el spec. *Juzgar el territorio leyendo la
redacción libre de la respuesta*: lo prohíbe la constitución (capa 1) y lo descartó H5.1 para los avisos; por eso el
esperado se compara por forma fija.

### D22 · Reglas del conjunto: se parametrizan, no se copian

**Decisión.** `ComprobarConjuntoDeBoeLegislacion` pasa a ser `ComprobarConjunto(evals, normas, reglas)` con dos juegos
de reglas: `ReglasDeBoeLegislacion()` (las diez de hoy, sin cambiar ninguna) y `ReglasDeLegalCore()` (al menos una eval
de municipio cubierto, otra de municipio no cubierto, otra de no activación, y ninguna eval activa sin esperado
verificable). `TestEvalsDelRepositorio` gana el subtest que aplica las reglas nuevas a `evals/legal-core`.

**Motivo.** Hoy las reglas están cableadas al nombre de una skill (V27) y `evals/legal-core/` solo quedaría vigilada
por el formato (V26). Copiar la función sería la tercera copia en H8.

### D23 · El job de evals mide las dos skills en una ejecución

**Decisión.** `.github/workflows/evals.yml` pasa a una matriz de skills (`boe-legislacion`, `legal-core`) con
`fail-fast: false`, un informe por skill; y el filtro de rutas que dispara el job gana **`internal/core/*`**, porque
desde H6 lo que las evals miden depende del dominio del territorio. `data/territorio/…` **ya lo cubre** el patrón
`data/*` que el filtro tiene (V31), así que no se añade nada por ahí.

**Motivo.** SC-015 exige que **en la misma ejecución** pasen las evals de `legal-core` y sigan pasando las de
`boe-legislacion`. Una ejecución del flujo con dos trabajos cumple eso y mantiene los informes separados.

**Alternativas.** *Pasar las dos skills en una variable y recorrerlas en `scripts/evals.sh`*: un solo informe con dos
skills mezcladas y el doble de tiempo en serie. *Dos ejecuciones*: no son «la misma ejecución».

### D24 · Lint y vocabulario

**Decisión.** Ninguna exclusión nueva de lint y ningún `//nolint`. En Go no se escribe suelta ninguna de las tres
palabras que `misspell` marca (V4): se escribe `aspecto` en singular —`cada aspecto de la cobertura`—, `configuración`
y `autónomas` con tilde. Si la implementación necesitara `aspectos` como clave o como identificador, entra en
`misspell.ignore-rules` con su comentario, como las 20 que ya están (V39); no se prevé, porque `cobertura` tiene tres claves
con nombre propio y ninguna se llama así (D11). `internal/core/ids` e `internal/core/territorio` llevan su `doc.go`
con comentario de paquete, que `revive` exige. Los dos analizadores de `ids` viven en ficheros distintos y comparten la
comprobación de cifras para no disparar `dupl`, cuyo umbral es 100 fichas.

### D25 · El fichero publicado se llama como su entidad: `schemas/municipio.json`

**Decisión.** El esquema publicado del applet es `schemas/municipio.json`, con `$id`
`https://ventanillalegal.es/schemas/municipio.json`, título `territorio · municipio` y la parte `$defs.resolver`. El
nombre del fichero es el de la **entidad**, no el del applet, porque es la regla que H4 dejó establecida: los dos
ficheros publicados son `norma.json` y `bloque.json`, con entidad `norma` y `bloque` y título `boe · <entidad>`
(`internal/app/esquemas_test.go:52-54`, `:133`, `:542`). La tabla `ficherosDeEsquemas`, al parametrizarse por applet (D16),
conserva sus tres columnas —`nombre`, `entidad`, `verbos`— y gana la del applet, que es lo que compone el título; la
regla «un fichero por entidad» sigue siendo una sola en los tres ficheros publicados.

**Alternativas.** *`schemas/territorio.json`*, el nombre del applet: es lo que se escribió en la primera redacción de
este plan y lo que sugiere la costumbre de agrupar por herramienta, pero deja dos reglas de nombrado conviviendo en
`schemas/` —dos ficheros por entidad y uno por applet— y obliga a que el lector sepa de cuál de las dos es cada
fichero. *Prefijar por applet* (`territorio-municipio.json`): resuelve de antemano una colisión de entidades entre
applets que hoy no existe, y que tampoco está resuelta para `norma.json`; renombrar los tres a la vez sería tocar
contrato publicado de H4 desde H6, fuera del alcance del hito. Si algún día dos applets publican la misma entidad, el
renombrado se decide entonces y alcanza a todos por igual.

**Consecuencia.** `schemas/territorio-municipios.yaml.json` (el esquema del **fichero de datos**) y
`schemas/municipio.json` (el del **contrato del applet**) son dos cosas distintas y sus nombres lo dicen: el sufijo
`.yaml.json` marca, como en `normas.yaml.json` y `eval.yaml.json`, el esquema de un fichero de `data/`.

### D26 · Los casos negativos de `TestSkillsDelRepositorio` se parametrizan por skill

**Decisión.** `internal/app/skills_test.go` deja de cablear `boe-legislacion` en sus casos negativos y los recorre
**por skill**. Hoy los controles de deriva, de normas nombradas y de instrucciones de evals recorren todas las skills,
pero los casos negativos —los que demuestran que cada control falla cuando debe— están atados a la constante
`skillDelHito = "boe-legislacion"` y a sus 32 usos (V25; `internal/app/skills_test.go:40`), de modo que con
`legal-core` en el árbol seguirían sin ejercitar nada suyo.

**Motivo.** El control 18 del plan y el contrato [skill-legal-core](./contracts/skill-legal-core.md) §5 se
comprometen a que los casos negativos se ejerzan también sobre `legal-core`; sin parametrizar, ese compromiso sería
falso y la skill nueva quedaría con controles que **pasan en vacío**, contra la obligación 12 del plan. Es el mismo
movimiento que D16 hace con la tabla de esquemas y D22 con las reglas del conjunto: quitar el cableado a un único
sujeto, sin añadir capacidad.

**Cuándo.** En su propia tarea (paso 20), **antes** de crear `skills/legal-core/`: la parametrización deja `make ci`
en verde recorriendo una sola skill, y cuando la segunda entra, los casos negativos la alcanzan sin tocar más código.

Lo que **no** queda en el paso 20 es la exigencia de que la skill exista. Hoy es
`require.Contains(t, nombres, skillDelHito)` con `skillDelHito = "boe-legislacion"`
(`internal/app/skills_test.go:38-40, :83`), la constante que existe para que las comprobaciones sobre `skills/` no
pasen en vacío (obligación 12 del plan). El paso 20 la convierte en la lista `skillsExigidas`, con un solo elemento y
`require.Subset`, y **el paso 21 vuelve al fichero** para añadirle `legal-core`: primero la exigencia, que queda en
rojo, y después los ficheros de la skill, que la devuelven a verde. Escribirla en el paso 20 dejaría un rojo entre dos
tareas; no escribirla en ninguno dejaría el control pasando en vacío justo para la skill que el hito entrega, que es
lo que la constante existe para impedir. Que `make skills-sync` **ejecute** ese fichero (`scripts/skills-sync.sh:7`)
no lo modifica, así que nada impide tocarlo en el paso 21.

**Alternativas.** *Dejarlo cableado y retirar el control del plan y del contrato*: es prometer menos de lo que el
hito necesita —la skill nueva se quedaría sin la demostración de que sus controles fallan cuando deben— y dejaría
`legal-core` en peores condiciones que `boe-legislacion`. *Copiar los casos negativos para `legal-core`*: segunda
copia que diverge en H8, el mismo defecto que D22 rechaza en `internal/evals`.

### D27 · Los controles sobre el corpus real viven en `internal/skills`, no en el dominio

**Decisión.** `TestPliegueCubreElCorpus`, `TestNombresAlcanzables` y `TestNingunNombreEsSoloCifras` **no son tests de
`internal/core/territorio`**: son tres subtests de `TestTerritorioDelRepositorio`, en
`internal/skills/territorio_test.go` —`pliegue-cubre-el-corpus`, `nombres-alcanzables` y
`ningun-nombre-es-solo-cifras`—. No hay ningún `corpus_test.go` en el dominio: **todos los tests del dominio son
sintéticos**, y los que miran los ficheros congelados reales están donde ya se leen esos ficheros.

**Motivo.** Un test del dominio **no puede leer un fichero**: la lista `core` de `depguard` deniega `os` e `io` a
`**/internal/core/**` y `run.tests: true` la extiende a los `_test.go` (V7), `io/fs` cuelga de `io` (V6) y
`//go:embed` no sube de directorio (V2). Leer `data/territorio/` desde `internal/core/territorio` solo sería posible
con un `//nolint`, que la obligación 5 del plan prohíbe. `internal/skills` es el sitio donde el repositorio ya hace
exactamente esto: lee el fichero real por ruta relativa al paquete, lo valida con el lector común y comprueba sobre lo
leído, y sus tests pueden importar el dominio (V42, patrón de `TestNormasDelRepositorio`). Ahí viven ya la validación
contra esquema, la integridad delegada en `Cargar` y el subtest `gramaticas` (contrato de identificadores §2), de modo
que **todo lo que se afirma del corpus congelado se comprueba en un solo test**, dentro de `make skills-check` y de
`make ci`.

**Lo que esto exige de la superficie del dominio** (data-model §2.1):

| Pieza | Por qué se exporta | Quién más la usa |
|---|---|---|
| `Plegar(nombre string) string` | El subtest de cobertura del pliegue comprueba que toda runa del corpus pliega al alfabeto cerrado | `internal/evals`, que compara el argumento del comando y los nombres de comunidad y provincia **plegados** (contrato de evals §2): sin exportarlo habría **dos pliegues** en el árbol, el defecto que D22 y D26 rechazan |
| `Ficheros` y los cuatro tipos de fichero congelado | `internal/skills` los valida contra su esquema con `ValidarDocumentoYAML[T]`, que necesita el tipo, y FR-044 pide hacerlo **sin repetir los tipos** | El propio `Cargar`, que decodifica `Fuentes` en ellos |

La alcanzabilidad se comprueba **por el comportamiento**, con `Cargar` + `Resolver` sobre los ficheros reales: no hace
falta exportar el índice ni las formas alternativas.

**Consecuencia sobre la cobertura.** `make test` mide por paquete y sin `-coverpkg` (V43): lo que estos subtests
ejercen **no cuenta** para el umbral de `internal/core/**`. La cobertura del dominio la sostienen sus tests sintéticos
(obligación 7 del plan), que es lo que D3 buscaba al hacerlo comprobable sin ficheros.

**Alternativas.** *Pasar al test del dominio los bytes del paquete `data`*: pasaría el lint —`data` no está en ninguna
lista denegada— pero haría que el dominio importara `data` en sus tests, contra D3 y contra la fila R1 del
*Constitution Check*, y ataría los tests del dominio al contenido real del repositorio, que es justo lo que D3 evita.
*Ponerlos en `internal/app`*, que ya tiene los bytes embebidos: separa estos tres controles del resto de lo que se
comprueba del corpus, los deja fuera de `make skills-check` y mezcla dominio con composición. *Un `//nolint` en el
test del dominio*: prohibido por la obligación 5.

### D28 · La tarea `[datos]` del esquema de eval trae consigo el único cambio de código que su forma impone

**Decisión.** El paso 17 —`schemas/eval.yaml.json` con la variante de territorio y la regla del esperado verificable—
es la **segunda excepción razonada** al «ninguna tarea `[datos]` toca código», del mismo tipo que D16: incluye, en la
misma tarea y bajo la misma pausa, la expectativa de `internal/evals/formato_test.go` que el esquema nuevo cambia, y
`internal/evals/formato.go` si la composición del mensaje lo exige.

**Motivo.** El `then` pasa de `required: [comandos, citas]` a `required: [comandos]` + `anyOf` (D21, punto 4), y el
esquema se compila **del fichero real** (V28). Con el `anyOf`, `incumplimientosDe` desciende a las hojas de sus dos
ramas y el caso `positiva-sin-citas` de `TestLeerEval` deja de ver el mensaje literal que compara con `EqualError`
(V44). Es decir: el esquema y esa expectativa son **un solo cambio**; separarlos deja `make ci` en rojo entre dos
tareas, contra la regla de rebanadas verdes. La expectativa no se «arregla» para que pase: el mensaje forma parte del
contrato que la tarea cambia, y en la misma tarea el caso pasa a llamarse `positiva-sin-esperado-verificable` y a
declarar el mensaje que el esquema nuevo produce.

**Alternativas.** *Reordenar los pasos 17 y 18*: no hay orden que lo evite, porque el rojo está **dentro** del paso 17
y los tests nuevos del paso 18 necesitan el esquema ya cambiado. *Conservar el mensaje dejando `citas` obligatoria*:
es lo contrario de lo que FR-084 pide. *Tolerar el rojo entre las dos tareas*: rompe la regla de rebanadas verdes.
*Relajar la comparación del caso a `ErrorContains`*: arreglar el control en lugar del contrato, prohibido por el
criterio 1 de la constitución.

### D29 · El `enum` de `rango` del esquema de normas crece en la tarea de las grabaciones, no en la del campo `vertebral`

**Decisión.** `schemas/normas.yaml.json` lo tocan **dos** tareas `[datos]`: el paso 13 añade la propiedad `vertebral`
y el paso 15 amplía el `enum` de `rango`, en la misma tarea y la misma pausa que las siete búsquedas grabadas. Los
valores nuevos se copian de las respuestas grabadas, no se escriben de memoria (S7).

**Motivo.** `TestEsquemaDeNormas/rangos-grabados` no comprueba que el `enum` contenga los rangos grabados: comprueba
con `assert.Equal` que sea **exactamente** ese conjunto, ordenado (`internal/skills/normas_test.go:375-380`, con
`rangosDelEsquema` en :415 y `rangosDeLasBusquedasGrabadas` en :455-475 sobre `evals.UnionDeGrabaciones()`; V47). La
igualdad va en los dos sentidos, así que los dos órdenes posibles dejan `make ci` en rojo: ampliar el `enum` antes
mete un valor que ninguna grabación produce, y grabar antes mete un rango que el `enum` no admite. Hoy el `enum` no
tiene «Ley Orgánica» (`schemas/normas.yaml.json:23-34`) y dos de las siete leyes que faltan lo son —la LOPJ 6/1985 y
la LOPDGDD 3/2018—, de modo que el paso 15, tal como estaba escrito, dejaba el repositorio en rojo y las filas de esas
dos normas sin validar en el paso 16.

Es el mismo movimiento que D16 y D28: una tarea `[datos]` indivisible, cuya pausa revisa juntas las dos piezas que el
control ata. A diferencia de las dos, **no es una excepción al «ninguna tarea `[datos]` toca código»**: las dos
piezas —las grabaciones y el `enum`— son material `[datos]` y caen dentro de la misma pausa. Se queda en el paso 15 y no al revés —las grabaciones movidas al 13— porque el paso 13 no puede grabar:
la grabación la hace una persona en la pausa (FR-073) y el manifiesto es material del paso 15.

**Alternativas.** *Dejar el `enum` en el paso 13*: rojo desde el paso 15 hasta el 16. *Partirlo en tres tareas con un
rojo tolerado en medio*: rompe la regla de rebanadas verdes de la capa 3. *Relajar el control a «contiene»*: arregla
el control en lugar del contrato, prohibido por el criterio 1 de la constitución, y perdería lo que la igualdad da
hoy: detectar un valor del `enum` que ninguna búsqueda grabada respalda.

## Supuestos no verificados

Lo que no se puede comprobar sin red, sin la plataforma o sin el fichero que la tarea `[datos]` genera. Ninguno es un
hecho en este plan.

| # | Supuesto | Qué lo resolverá |
|---|---|---|
| **S1** | La relación del INE (`diccionario26.xlsx`) trae, por municipio, código de comunidad, código de provincia, código de municipio, dígito de control y nombre oficial | La tarea `[datos]` de `municipios.yaml`; si trae menos, el fichero declara lo que hay y `cobertura` lo dice |
| **S2** | El DIR3 del ayuntamiento es `L` + el número de inscripción del REL (`L01PPMMMDC`) | La muestra de FR-046, en la pausa de la tarea `[datos]` (ADR 0017, decisión 3) |
| **S3** | Son 8.132 municipios y el fichero de municipios ronda los 650 KB, de modo que su análisis en cada invocación está por debajo de 200 ms | `docs/SOURCES.md` da la cifra de municipios; el tiempo lo mide el e2e con `cronometra`. **Medido (T026 y T029, 2026-09-22)**: la cifra se confirma —8 132 filas en la relación y en la correspondencia, 632 KB y 183 KB— y el tiempo no: la ejecución `ci` de la propuesta de cambio midió 235 ms al resolver Leganés (nota de T026). Cargar asignaba 36,4 MB en 756 111 asignaciones, el 72 % en el árbol de nodos del lector de YAML y la decodificación fila a fila. Desde T029 las filas escritas en la forma de D4 se leen sin el lector, con el mismo resultado y los mismos defectos que él, y cualquier otra forma la lee él entera; ordenar el registro tampoco vuelve a escribir cada código en cada comparación. Con los datos reales, una carga asigna 9,8 MB en 63 333 asignaciones (9 ms en local, y 17 ms la invocación entera). `TestCosteDeLaCarga` lo fija sobre fuentes sintéticas del tamaño real: como mucho un tercio de las 779 065 asignaciones y los 37 455 908 bytes que la carga anterior necesitaba para ellas |
| **S4** | El pliegue de nombres no deja ningún municipio inalcanzable ni descubre runas fuera de la tabla | Los dos tests de D10, en cuanto exista `municipios.yaml` |
| **S5** | Los códigos INE de comunidad que se citan como ejemplo en estos artefactos (13 Madrid, 15 Navarra, 16 País Vasco, 18 y 19 para Ceuta y Melilla) | La tarea `[datos]`: los códigos salen del fichero del INE, no de estos documentos |
| **S6** | GitHub ejecuta la matriz de dos skills como una sola ejecución del flujo, con un identificador de ejecución común | La ejecución de cierre del hito (tarea `[plataforma]`) |
| **S7** | Los identificadores `BOE-A-…`, los títulos y los **rangos** de las siete leyes vertebrales que faltan en `data/normas.yaml` (Código Civil, LJCA, LEC, LOPJ, LGS, LOPDGDD y Ley General Presupuestaria) | La tarea `[datos]` que graba sus búsquedas: los tres se copian **de la respuesta grabada**, nunca de `refs/`, que advierte de que algunos están escritos de memoria (FR-071). De los rangos se sabe que al menos aparece «Ley Orgánica» (LOPJ y LOPDGDD), que hoy falta en el `enum`; cuál es el del Código Civil, de 1889, lo dice la grabación y no estos documentos (D29) |
| **S8** | Que el volcado del REL siga descargándose como dice su fila de `docs/SOURCES.md` | La persona que ejecuta la tarea `[datos]`; `scripts/verify-sources.sh` no puede vigilarlo (ADR 0017) |
| **S9** | El tiempo de la sesión de eval de `legal-core` cabe en el tope de 240 s del job | La ejecución de cierre |
