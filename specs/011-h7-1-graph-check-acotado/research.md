# Research: H7.1 · `graph check` acotado, señales que se apagan, salida legible y H7 sin lo que no pasa el umbral

**Modo**: desatendido. Cada decisión (D1-D24) lleva la alternativa rechazada y el motivo, con el «Criterio de decisión
autónoma» de la constitución. Toda afirmación sobre una herramienta o dependencia externa remite a la tabla V, con
dónde se comprobó en local y sin red; lo que no se puede comprobar así va como supuesto (S1-S3) y no se afirma como
hecho en ningún artefacto.

## Verificaciones en local (V)

Sondas ejecutadas el 2026-09-28 con `go1.27.1 darwin/arm64` y los módulos de `go.sum` (sin red: `GOFLAGS` por
defecto, módulos en `$GOMODCACHE`). Las de SQLite se ejecutaron como un `_test.go` temporal en `internal/graph`
(`TestProtoH71`, `TestProtoH71Bloqueo`), usando `cadenaDeEscritura`, `cadenaDeLectura` y `versionRegistrada` del
propio paquete, y el fichero se borró en el mismo paso (`git status` limpio después).

| # | Afirmación | Dónde se comprobó | Resultado |
|---|---|---|---|
| V1 | `modernc.org/sqlite v1.59.0` lleva SQLite 3.53.4 | sonda: `select sqlite_version()` | `3.53.4` |
| V2 | La cadena de escritura de H7 (`aplicar.go:32-35`, sin `mode`) crea `world.db` en su sitio con la primera consulta | sonda sobre un directorio sin `world.db` | el fichero existe tras la primera consulta |
| V3 | Crear antes el fichero con `os.OpenFile(O_RDWR\|O_CREATE, 0o600)` y abrirlo después con SQLite conserva `0600` (el precedente es `internal/cache/abrir.go:96-112`) | sonda | `-rw-------` |
| V4 | `PRAGMA journal_mode=WAL` fuera de transacción devuelve `wal` sobre un fichero recién creado | sonda | `wal` |
| V5 | Con `_pragma=foreign_keys(1)` (la cadena de escritura), una arista con un extremo que no está en `nodes` falla con `SQLITE_CONSTRAINT_FOREIGNKEY` (787), también dentro de una transacción | sonda | `constraint failed: FOREIGN KEY constraint failed (787)` en los dos casos |
| V6 | `INSERT INTO lecturas … ON CONFLICT (bloque) DO UPDATE SET …` sustituye la fila del bloque con los valores dados (y el `SET` lee la fila anterior si se le pide) | sonda: A, B y B otra vez | `(B, A)` y después `(B, B)` |
| V7 | `json_extract(props, '$.identificador')` y `json_each(?)` con una lista JSON como parámetro están disponibles | sonda | encuentran la `Norma` y filtran el `Bloque` pedido |
| V8 | La cadena de lectura sin auxiliares (`mode=rw&…&query_only(1)`) rechaza una escritura de SQL y, al cerrar, no deja `-wal` ni `-shm` | sonda | `attempt to write a readonly database (8)`; en el directorio solo queda `world.db` |
| V9 | Un `world.db` que es un directorio da `SQLITE_CANTOPEN` (14) en la primera consulta de la cadena de escritura; un fichero que no es una base, `SQLITE_NOTADB` (26), sin escribir en él | sonda | `unable to open database file (14)`; `file is not a database (26)` |
| V10 | Un escritor sobre la base propia en WAL con `PRAGMA locking_mode=EXCLUSIVE` y una transacción de escritura abierta bloquea a un lector con las dos cadenas de lectura: `SQLITE_BUSY` tras el tramo de `busy_timeout(100)` | sonda | `database is locked (5)` a los ~110 ms, en los dos modos |
| V11 | Kong admite un argumento de posición opcional seguido de otro opcional acumulativo (`[]string`), siempre que ningún obligatorio vaya detrás de un opcional y el acumulativo sea el último | `github.com/alecthomas/kong@v1.16.1/build.go:232-245` (`validatePositionalArguments`), `context.go:1024-1034` (`checkMissingPositionals`), `tag.go:262-274` | admitido |
| V12 | La línea de uso de Kong es el camino del verbo, los argumentos —los opcionales de posición anidados— y `[flags]` al final: `graph check [<norma> [<bloques> ...]] [flags]` | `kong@v1.16.1/help.go:121,160` (`Usage: %s %s` con `Summary()`), `model.go:159-192` (`Node.Summary`: banderas obligatorias, argumentos y `[flags]` si hay alguna opcional) y `315-330` (`Value.Summary`); la de `show` hoy: `^Usage: graph show <id> \[flags\]$` (`h7-grafo-applet.txtar:44`) | como se dice |
| V13 | La ayuda de un verbo (`kitlegal graph check --help`) escribe su descripción partida en líneas de ~80 columnas | binario de `main` construido en `bin/kitlegal` (ignorado por git) | la descripción de H7 sale en dos líneas: un guion la busca con `\s+` entre palabras, como hizo H7 con `--no-graph` |
| V14 | `(?i:…)` de `regexp` pliega mayúsculas con `unicode.SimpleFold`, así que `(?i:REDACCIÓN)` casa con «redacción» | `$(go env GOROOT)/src/regexp/syntax/parse.go:386`; sonda | casa |
| V15 | `--describe` no distingue un argumento de posición de una bandera: la entrada es un objeto con una propiedad por campo y `required` para los de posición no opcionales | `internal/cli/describe.go:175-196`, `475-491` | sin distinción |
| V16 | El generador de la tabla de comandos escribe todo argumento no obligatorio como `[--nombre]` | `internal/skills/comandos.go:575-590`; tests `internal/skills/comandos_test.go:762` y `internal/app/skills_test.go:205` | `[--nombre]` |
| V17 | El kernel escribe `Resultado.Legible` en lugar de la tabla mínima sin `--json`, y le añade el salto final si falta | `internal/cli/sobre.go:125-131`, `internal/render/texto.go:21-33` | sí |
| V18 | La tabla mínima empieza por las líneas `fuente`, `url`, `fecha_consulta` y `hash` | `internal/render/tabla.go:73-78` | sí |
| V19 | El «rojo primero» del workflow da por inválido un guion cuya salida contenga `unknown command`, `cannot parse`, `unexpected command` o `usage: ` (minúsculas) | `scripts/workflow/aceptacion.sh:55` | la ayuda de Kong escribe `Usage:` con mayúscula |
| V20 | Solo `boe articulo` y `boe articulos` emiten `eli:has_version`, una vez por bloque distinto de la invocación | `internal/source/boe/grafo.go:22-41,74`, `articulo.go:112`; `grep` de `RelacionTieneVersion` fuera de tests | solo ahí |
| V21 | Solo `boe` declara vigencia en lo que observa (604 800 s); `territorio` no | `internal/source/boe/fuente.go:31`; `grep` de `Vigencia` en `internal/app/territorio.go` e `internal/core/territorio` | solo `boe` |
| V22 | La preparación de la caché de una sesión de eval no entrega al grafo (el registro no llama a `EntregarAlGrafo`); la del grafo previo sí, a `dirCache/world.db` | `internal/evals/preparar.go:147-168`, `336-380`, `385-397` | sí |
| V23 | El informe del job de evals da la tasa por serie (eval y modelo) y los avisos solo por sesión | `internal/evals/informe.go:124-208`, `829-849`, `892-893` | sí |
| V24 | Una tarea `[aceptacion]` se verifica con el guardián, el «rojo primero» **y** `make ci` | `scripts/workflow/verificar.sh:34-50` | `make ci` también |
| V25 | Una entrega cuya transacción inmediata se confirma sin cambiar ninguna fila no cambia ningún fichero del directorio | `TestApplyIdempotente` (`internal/graph/almacen_test.go:290-339`, `huellasDelArbol` tras cada entrega repetida), ejecutado en local | `ok` |
| V26 | Con `Norma *string arg:"" optional:""` y `Bloques []string arg:"" optional:""`, Kong deja `Norma` en `nil` sin argumentos y en `""` con un argumento vacío, y reparte `BOE-A-… a21 a22` y `BOE-A-… " "` en norma y bloques sin error | `kong@v1.16.1/mapper.go:293` (`ptrMapper`); sonda temporal en `internal/cli` (`kong.New` + `Parse`), borrada tras ejecutarla | como se dice |
| V27 | El único constructor de un lote con procedencia escribe la fecha de consulta con `time.RFC3339Nano` a partir del `time.Time` del sobre y copia la vigencia de lo observado | `internal/cli/entrega.go:28-35` (`LoteDe`); `grep` de `core.Lote{` fuera de tests | solo ahí |
| V28 | Los emisores solo ponen cadenas en los `Datos` de un nodo, y ninguno emite `Persona`; un argumento llega al applet en UTF-8 válido salvo el `cli.Literal` de `graph show`, que no emite | `internal/source/boe/grafo.go:63-71` y `internal/core/territorio/grafo.go:29-39` (los dos únicos `schema.Nodo{` fuera de tests con datos); `grep` de `TipoPersona` fuera de tests (solo `internal/core/grafo`); `kong@v1.16.1/scanner.go:188` y `mapper.go:956-965` (`jsonTranscode` pasa el argumento por `json.Marshal`); `internal/app/grafo.go:110` | como se dice |
| V29 | Los dos emisores solo ponen en `Operaciones` valores `schema.Nodo`, `schema.Arista` y `schema.Texto` —nunca `nil` ni un puntero—, cada arista detrás de sus dos extremos y ninguna dos veces (`boe` emite una vez cada bloque distinto), y un id repetido en un lote siempre con los mismos datos (la `Norma` de cada bloque de `boe articulos`); `LoteDe` los copia tal cual. Tras FR-075 ningún `Rechazo` lleva una arista | `internal/source/boe/grafo.go:22-40,64-75` e `internal/core/territorio/grafo.go:30-43`; `internal/cli/entrega.go:28-35`; `grep` de `Operaciones` y de `Rechazo{` fuera de tests: los únicos que llevan una arista son `ValidarContraGrafoVacio` (`internal/core/grafo/lote.go:92-94`), `validarArista` (`lote.go:276-280`) y `leerArista` (`internal/graph/aplicar.go:459`), que se retiran (D20) | como se dice |

## Supuestos no verificables en local (S)

- **S1**: que el job de evals, con Sonnet 5 y tres repeticiones, dé la eval 19 aprobada y con `red` vacío (SC-006). Se
  mide en la propuesta de cambio, tras la revisión final; el run no lo mide desde una tarea.
- **S2**: que la misma pregunta hecha dos veces en Claude Code lleve la forma en la primera respuesta y no en la
  segunda (SC-007). Lo comprueba la persona al leer el informe; su base en el binario son los pasos 2 y 4 de SC-003,
  que la suite congelada sí prueba.
- **S3**: que la cota de 3 s de H7 SC 008 y la de 150 ms de H7 SC 007 sigan dándose en el runner de CI con la tabla de
  lecturas (una consulta por clave primaria y un `INSERT … ON CONFLICT` por bloque leído). Se mide en `test-tiempos`;
  en local no hay runner.

## Decisiones

### D1 · Cómo se guardan las lecturas: una tabla con dos ranuras por bloque

**Decisión**: migración `0002_lecturas.sql` con `lecturas(bloque PK, ultima, anterior)`, las tres con clave ajena a
`nodes` y ninguna nula (data-model §1). Cada lectura desplaza `ultima` a `anterior`; la primera de un bloque guarda
`(v, v)`. La entrega calcula la fila nueva en Go y la escribe solo si cambia.
**Por qué**: FR-023 y FR-030 solo miran la última lectura y la anterior; guardar más no lo pide nada (FR-015, «Fuera de
alcance»: `graph history`). El tamaño es una fila por bloque distinto (2 400 con la medida) y no crece con las preguntas.
Con `(v, v)` en la primera lectura y la escritura solo si cambia, una entrega repetida idéntica no escribe nada, como
exigen H7 FR 022 y los tests que se quedan (`TestApplyIdempotente`, `TestIntegracionIdempotencia`: «la entrega
repetida no escribe nada»); con `anterior` nulo en la primera, la segunda entrega idéntica escribiría la fila.
**Alternativas**: (a) una fila por lectura —crece con cada pregunta sin consumidor—; (b) columnas nuevas en `nodes` —
mezcla con la historia de observaciones que leen `show` y `stats` (FR-022) y obliga a reescribir filas de H7—; (c)
derivarlo de `last_seen` —una lectura servida por la caché conserva la `fecha_consulta` original (H4, ADR 0015), así
que el paso 4 de FR-025 no se distingue del 3—. Rechazadas.

### D2 · Qué es una lectura en un lote

**Decisión**: cada arista `eli:has_version` del lote consolidado es una lectura de su `Bloque` que vio su `BloqueVersion`
(`Consolidado.Lecturas()`, en `internal/core/grafo`).
**Por qué**: es lo que ya emiten `boe articulo` y `boe articulos` y solo ellos (V20), una por bloque e invocación
aunque se nombre dos veces, y `Consolidar` deja una arista por terna: FR-020 sale del emisor que existe sin campo nuevo
en `schema.Observado` ni cambio en `boe` (spec, «Fuera de alcance»: no cambiar qué emite `boe`). No hay lectura sin
entrega confirmada, así que `--no-graph`, `--dry-run`, un código distinto de 0 y una entrega que falla no cuentan.
**Alternativa**: un campo `Lecturas` en `schema.Observado` que marque el applet. Rechazada: duplica lo que la arista ya
dice y cambia el contrato del `Resultado` (ADR 0005, ADR 0014) sin que nada lo pida.

### D3 · Un bloque sin fila: la redacción observada la última, calculada en Go

**Decisión**: un bloque sin fila en `lecturas` —un `world.db` de H7, o un bloque que H7 observó y nadie ha vuelto a
leer— cuenta con una lectura, la de la `BloqueVersion` suya con la última observación más reciente (instante de
`fecha_consulta`; a igualdad, id menor comparando bytes): una fila de partida `(R, R)` (FR-026 y su supuesto). Una
sola función de dominio, `grafo.RedaccionVistaSinLecturas`, la calcula en dos sitios: al comprobar (bloque sin fila)
y al entregar la primera lectura de ese bloque, **antes** de aplicar el lote, para poner `anterior`.
**Por qué**: FR-026 exige leer un `world.db` de H7 sin perder nada y sin dar `version-obsoleta` hasta que una lectura
nueva vea otra redacción; la regla vive una sola vez, en Go, y las migraciones siguen siendo SQL puro (H7 D12).
**Alternativas**: (a) sembrar las filas en la migración con SQL (`row_number() OVER (… ORDER BY last_seen)`) —duplica
la regla en SQL, compara instantes como texto, y aun así la lectura de un `world.db` v1 sin migrar necesita la regla
en Go—; (b) un paso de siembra en Go dentro de `migrar` —añade a las migraciones un mecanismo que hoy no tienen para
un estado que el cálculo perezoso ya cubre—. Rechazadas.

### D4 · Un `world.db` de versión 1 se lee sin migrar

**Decisión**: `Leer` acepta una versión entre 1 y la conocida (2); sin la tabla `lecturas`, todos los bloques son «sin
fila» (D3). La primera entrega migra a la 2 dentro de su transacción (H7 FR 013). Una versión posterior sigue siendo
«tiene el esquema en la versión N… no se modifica» (H7 FR 012).
**Por qué**: los verbos de `graph` no escriben (H7 FR 004, FR 005; FR-014); SC-012 pide `check` y `stats` sobre un
`world.db` de H7.
**Alternativa**: migrar al leer. Rechazada: sería escribir desde un verbo de lectura. Tampoco se copia la regla de la
caché (versión distinta de 0 y de la conocida = ajena, `internal/cache/abrir.go:144-148`): la caché nunca ha migrado y
el grafo sí tiene que leer lo que dejó H7.

### D5 · `version-obsoleta` por lecturas; `fuente-caducada` solo sobre lo vigente

**Decisión**: data-model §5. `version-obsoleta` sobre la redacción de `anterior` cuando la de `ultima` tiene fecha de
vigencia válida y estrictamente posterior; `fuente-caducada` con la condición de H7 FR 066 sobre `Norma`, `Bloque` y
la redacción vista de cada bloque; las plantillas de las explicaciones no cambian. Se retiran `versionesObsoletas` y
`compararRecencia`.
**Por qué**: FR-023, FR-024, FR-030. Con `territorio` sin vigencia declarada (V21), la restricción a esos tres tipos no
cambia nada de lo que ve nadie hoy y es la lectura literal de FR-030.
**Alternativa**: dejar `fuente-caducada` sobre todo nodo salvo las redacciones superadas. Rechazada: FR-030 dice «solo
sobre la `Norma`, sobre el `Bloque` y sobre la redacción que vio la última lectura».

### D6 · Argumentos de `check`: de posición y opcionales, como los de `boe`

**Decisión**: `kitlegal graph check [<norma> [<bloques>...]]`: `Norma *string arg:"" optional:""` y
`Bloques []string arg:"" optional:""` (V11, V26). Sin norma (`nil`), todo lo consultado; con una norma dada —también
vacía, como la de una variable sin valor en `kitlegal graph check "$NORMA"`—, antes de abrir nada se valida con
`boe.ValidarNorma` (la misma gramática de `boe articulo`, FR-004), y cada bloque, contra vacío o solo
`unicode.IsSpace`; los dos errores envuelven `cli.ErrArgumentos` («argumentos inválidos: …», clase `argumentos`, 2), y
el mensaje nombra el valor. Con un `string`, `""` no se distinguiría de no dar la norma y comprobaría todo con 0.
**Por qué**: el mismo orden que `boe articulo <norma> <bloque>` y `boe articulos <norma> <bloques>...`, así que la skill
pasa a `check` lo que acaba de pasar a `boe`; `Juzgar` reconoce la norma con el `esDeLaNorma` que ya usa
(`internal/evals/juzgar.go:466-468`); y `graph check ine:28074` sigue saliendo con 2 y `argumentos inválidos` en la
salida de error, como afirma `h7-grafo-codigos.txtar:85-92`, que FR-081 congela en esa sección.
**Alternativa**: banderas `--norma` y `--bloque`. Rechazada: órdenes más largas para la skill, distintas de las de
`boe`, y `Juzgar` tendría que analizar `--norma X` y `--norma=X`.

### D7 · La tabla de comandos escribe los opcionales de posición como Kong

**Decisión**: `sintaxisDeLaOrden` (`internal/skills/comandos.go:575-590`) escribe un argumento no obligatorio como
`[<nombre>` (o `[<nombre>...`), cerrando todos los corchetes al final, como la ayuda de Kong (V12):
`kitlegal graph check [<norma> [<bloques>...]]`. Cambian con ello `comandos_test.go:762`, `skills_test.go:205` y
`invocacionDeLaSintaxis` (`skills_test.go:1311-1331`).
**Por qué**: la tabla de `SKILL.md` es lo que lee el agente (FR-007, FR-040); con la regla de hoy diría
`[--norma] [--bloques]`, una orden que no existe. `--describe` no distingue posición de bandera (V15), pero en
`kitlegal` todos los argumentos propios de un verbo que presenta una tabla son de posición: las únicas banderas
propias son las de `skills install|list|doctor` (`internal/app/instalacion.go:164-167,185-186`), y ninguna skill
declara el applet `skills`.
**Alternativas**: (a) dejar `[--nombre]` —la tabla enseñaría una orden inválida—; (b) marcar en `--describe` qué
argumento es de posición —cambiaría el esquema publicado de todos los verbos con argumentos (`schemas/bloque.json`,
`norma.json`, `municipio.json`…) sin que el hito lo pida—. Rechazadas.

### D8 · El ámbito se lee acotado en SQL

**Decisión**: `Lectura.Instantanea(ctx, ambito)`. Con norma: la `Norma` por `type = 'Norma'` y
`json_extract(props, '$.identificador')` (V7); sus `Bloque` por `eli:has_part` (clave primaria de `edges`), filtrados
por `json_extract(props, '$.bloque')` en `json_each(?)` si se nombran; sus `BloqueVersion` por `eli:has_version`; y las
filas de `lecturas` de esos bloques (si la versión es 2). Sin norma: todo, como en H7, más todas las filas.
**Por qué**: la comprobación de una pregunta cuesta lo que su norma —una pasada por las `Norma` (cientos) y lecturas
por clave del resto—, no lo acumulado (criterio de uso). La instantánea acotada está cerrada bajo lo que usan las
citas (`[BOE-A-…, bloque …]`), así que `Comprobar` aplica las mismas reglas a todo lo que recibe.
**Alternativa**: leer todo y filtrar en Go. Rechazada: cada pregunta leería el grafo entero (≈ 5 000 nodos con la
medida) para devolver unos pocos hallazgos.

### D9 · La forma de `data` de `check`: plana, con los totales por clase

**Decisión**: `{"norma":…,"bloques":[…],"version-obsoleta":n,"fuente-caducada":m,"omitidos":o,"hallazgos":[…]}`
(data-model §6; contracts/applet-graph.md §3). `norma` vacía y `bloques` `[]` sin argumentos; los totales llevan como
clave el nombre de la clase, el mismo que la salida legible.
**Por qué**: FR-012 pide la lista, el total de cada clase, los omitidos y el ámbito; y SC-005 acota a ≤ 3 800 bytes la
comprobación con cinco bloques cambiados. Medido con bytes reales (plan, «Uso»): un `version-obsoleta` del art. 21 de la
LPAC pesa 691 bytes, el sobre sin `data` 197 y esta `data` sin hallazgos con cinco bloques, 139: 197 + 139 + 5 × 691 +
4 = **3 795** bytes. Con `data` anidada (`"ambito":{…}` y `"totales":{…}`) serían 3 818, por encima de SC-005.
**Alternativa**: la anidada. Rechazada por lo anterior. Y `"ambito": null` sin argumentos, rechazada: un nulo en el
esquema reflejado por `invopop/jsonschema` no se describe como tal, y una forma fija es más fácil de consumir.

### D10 · Cota y orden

**Decisión**: `grafo.MaximoDeHallazgos = 50`; orden `version-obsoleta` → `fuente-caducada` → id por bytes; `Comprobar`
calcula todos, los ordena, cuenta y corta. Sin bandera (FR-015).
**Por qué**: FR-010, FR-011, FR-013. Calcularlos todos cuesta lo que H7 ya medía (H7 SC 008, 3 s sobre 10 000 nodos,
S3), y los totales los necesitan.

### D11 · La salida legible la compone el applet a partir de `data`

**Decisión**: `internal/app/grafo_legible.go` con `legibleDeStats`, `legibleDeShow` y `legibleDeCheck`, puestas en
`Resultado.Legible` (ADR 0026, V17), como `instalacion_legible.go` de `skills`: texto determinista, sin tabuladores ni
secuencias de escape, columnas alineadas con espacios, sin las líneas del sobre (V18) ni pares ruta/valor; las
plantillas exactas en contracts/applet-graph.md §5.
**Por qué**: FR-060 a FR-064; es el patrón del repositorio (el kernel no cambia).
**Alternativa**: una estrategia de `internal/render` para `graph`. Rechazada: `render` no conoce los tipos del dominio
y ADR 0026 decidió que lo compone el applet.

### D12 · La etiqueta de `version-obsoleta` vive en el dominio del grafo

**Decisión**: `grafo.EtiquetasDeHallazgo()` = `{version-obsoleta: "REDACCIÓN MODIFICADA"}` en
`internal/core/grafo/vocabulario.go`. `internal/evals` compone la forma fija con el mismo constructor que los avisos
(se extrae de `formasDeAviso`, `avisos.go:48-64`, una función `formaFija(etiqueta)` que usan los dos), y añade
`ExtraerHallazgos`, `ComprobarFormasDeHallazgo` y `ComprobarClasesDeHallazgo` (el enumerado del esquema = las clases
etiquetadas). Un test fija que la etiqueta no coincide con ninguna de `boe.EtiquetasDeAviso()`.
**Por qué**: FR-045: el binario es la única fuente de verdad, como `boe.EtiquetasDeAviso` para los avisos; la clase es
del dominio del grafo, no de `boe`. Meterla en `boe.EtiquetasDeAviso` rompería «avisos-del-esquema»
(`conjunto_test.go:940-948`) y añadiría un valor a `avisos` (FR-054).
**Alternativa**: escribir la etiqueta en `internal/evals`. Rechazada: no sería del binario.

### D13 · Formato común de eval: `hallazgos` y la norma de la comprobación

**Decisión**: data-model §8 y contracts/evals-y-skill.md §1-§2. `hallazgos` (lista del enumerado
`clase-de-hallazgo`) y `norma` opcional en `comando-comprobacion`; `Juzgar` reparte `hallazgos_encontrados` y
`hallazgos_ausentes` con `ExtraerHallazgos` y exige la norma con `esDeLaNorma`. El texto del comando con norma es
`graph check <norma>`.
**Por qué**: FR-050 a FR-054; compatible hacia atrás porque las dos piezas son opcionales y ninguna otra eval las usa.
**Alternativa**: reutilizar `avisos` con un valor nuevo. Rechazada por FR-054 («MUST NOT añadir valores a `avisos`»).

### D14 · El informe declara las formas que exige cada eval

**Decisión**: `TasaDelInforme` gana `formas` (la forma fija literal de cada hallazgo que exige la eval,
`["⚠ REDACCIÓN MODIFICADA:"]` en la 19, `[]` en las demás) y la tabla de tasas de `informe.md` una columna «Formas
exigidas»; cada sesión publica `hallazgos_encontrados` y `hallazgos_ausentes`, y su tabla, dos columnas como las de
avisos.
**Por qué**: FR-055 y SC-006: la tasa ya sale por serie (V23); faltaba declarar la forma junto a ella. `informe.sh` del
workflow no cambia: lee `tasas` y no le afecta un campo más (el workflow no se toca desde un hito).

### D15 · La eval 19 no cambia su estado previo

**Decisión**: `evals/boe-legislacion/19-…yaml` gana `norma: BOE-A-2015-10565` en su comando `graph check` y
`hallazgos: [version-obsoleta]`, y su comentario deja de decir que el traslado no se comprueba. Su `grafo_previo` no
cambia.
**Por qué**: la preparación del grafo previo es una entrega de la lectura de 20151002 (`dirCache/world.db`) y la de la
caché de la sesión no entrega (V22): la única lectura de la sesión, servida por la caché (20161002), deja la fila
`(20161002, 20151002)` y `check` con la norma da el `version-obsoleta` sobre 20151002 (FR-053), sin red.
**Cuándo**: no en la tarea `[aceptacion]`: con las claves nuevas, el fichero no valida contra el esquema de hoy y
`make ci` (que la verifica, V24) quedaría en rojo. Cambia en la tarea que implementa FR-050-FR-054, detrás de la tarea
`[datos]` del esquema.

### D16 · `boe-legislacion` v0.1.1

**Decisión**: contracts/evals-y-skill.md §4: una comprobación por norma citada, después de leer y con los bloques
leídos; cada bloque leído una vez; `version-obsoleta` con la forma fija; `fuente-caducada` no se traslada; sin
`version-obsoleta` y con 0, nada; con otro código, la regla 7 de hoy. Se retira la comprobación antes de leer y la
frase «Un hallazgo no es un aviso de vigencia: no lleva la forma fija de los avisos».
**Por qué**: FR-040 a FR-047. Con la tabla regenerada, `SKILL.md` queda en unas 250 líneas (hoy 246).

### D17 · Lectura del almacén tras la retirada

**Decisión**: contracts/almacen-world-db.md §3: ausente o de 0 bytes → grafo vacío sin abrir SQLite (H7 FR 004); si hay
`world.db-wal` junto a `world.db`, `mode=ro` (lee lo confirmado por una escritura propia interrumpida o por otra
invocación abierta); si no, `mode=rw` con `query_only` (V8). Cualquier fallo que no sea la espera o el plazo es
`errorInutilizable` (1, la ruta en el mensaje). Se retiran la comprobación de permisos, el modo inmutable y su
reapertura, `EvalSymlinks`, la búsqueda de `-journal` y `-shm`, y la clasificación 776/1544/14.
**Por qué**: FR-070, FR-072, FR-077. El `-wal` propio y la concurrencia tienen vía real; lo demás, no.

### D18 · Escritura en su sitio

**Decisión**: contracts/almacen-world-db.md §4: `os.MkdirAll(dir, 0o700)` (fallo → «no se puede escribir world.db en
…», H7 FR 011/033), `os.OpenFile(world.db, O_RDWR|O_CREATE, 0o600)` y cerrar (V3), y después lo de H7: cadena de
escritura, WAL si la versión es 0, transacción inmediata, migraciones, lecturas previas (D3), nodos, textos, aristas y
filas de `lecturas`. Se retiran `publicar.go`, la costura `enlazar`, la limpieza de temporales y de directorios, la
comprobación de permisos y `ValidarContraGrafoVacio`.
**Por qué**: FR-071: con `KITLEGAL_CACHE_DIR` en un sistema sin enlaces duros la primera entrega crea `world.db`; una
creación interrumpida deja como mucho el fichero sin esquema (0 bytes, o en WAL sin tablas) que H7 FR 004 y FR 013 ya
leen y completan. Los permisos 0700/0600 son los de H7 (research D8 de H7) y los de la caché.
**Alternativa**: que SQLite cree el fichero (V2). Rechazada: nacería con los permisos del `umask` (0644).

### D19 · Errores y mensajes

**Decisión**: se retiran `errorEsDirectorio`, `errorDeTransaccionInterrumpida`, `errorDeFicheroNoEscribible` y
`errorDePublicacion`; `errorInutilizable` pasa a ser `grafo: "<ruta>" no es una base de datos utilizable: <causa>`, sin
«; no se modifica» ni el detalle de «acceso denegado». El error de `check` sobre lo que ninguna entrega guarda pierde
también «; no se modifica» (`internal/app/grafo.go:176-177`). El de esquema posterior lo conserva (H7 FR 012 se queda).
**Por qué**: FR-070 («nada se promete sobre los bytes de `world.db`»): un mensaje que dice «no se modifica» lo promete.
Los guiones de H7 que afirman ese texto cambian en esa aserción (FR-080, contracts/arnes-e2e.md §3).

### D20 · Validación y desempate en `internal/core/grafo`

**Decisión**: fuera `ValidarContraGrafoVacio`, la comprobación de extremos en `leerArista` (la clave ajena rechaza,
V5, y el lote entero se deshace: «no se pudo escribir…», 1), el URI no absoluto, el nodo sin id o sin tipo y la arista
sin origen, relación o destino (`validarArista` entera: un extremo vacío es un extremo ausente, que la clave ajena ya
rechaza), con sus tests; con ellos sale el caso de la arista de `nombrar` (`errores.go`), que solo nombraba esos
rechazos: tras FR-075 ningún `Rechazo` lleva una arista (V29). `desempatar` solo por `url`; `compararUltimas` por
instante y `url`; `Consolidar` guarda una vez un id repetido en el lote, el de su primera aparición, sin comparar sus
datos (los emisores lo repiten con los mismos datos, V29; el nivel de los datos canónicos sale con FR-076). Se quedan,
con un test cada uno: sin fuente, url o fecha; id con otro tipo (guardado o del lote); huella que no es la de su cuerpo
o que ya guarda otro cuerpo (los diez casos de `probarHuellas…` se reducen a uno por rechazo); la regla de `Persona`
con sus casos ASCII. Ninguna otra entrada del lote tiene caso ni test propio: una operación nula o que no es un valor
`schema.Nodo`, `schema.Arista` o `schema.Texto`, una arista delante de sus extremos y un id repetido con otros datos
(V29); una fecha de consulta que no es RFC 3339, una vigencia negativa o con fracción de segundo y unos datos sin
forma JSON canónica —un número que no es finito, una cadena o una clave que no es UTF-8, un valor que no es de JSON—,
en cualquier nodo o en una `Persona` (V27, V21, V28): no los produce ningún emisor. El código que hoy los rechaza
(`validarOperacion` con `nil` y con `default` en `lote.go`, `nombrar` con `default` en `errores.go`,
`validarProcedencia` en `lote.go`, `consolidar` en `internal/graph/aplicar.go:48`, `DatosCanonicos` en `canonico.go` y
sus usos al consolidar y en `validarPersona`) se queda: es la regla genérica en el código —lo que el binario no produce
no entra, y la entrega lo dice como `inesperado`—, y FR-075 cierra la lista de rechazos cuyo código se retira (FR-078,
en `internal/graph`). Lo que sale son todos sus casos de los tests:
- `lote_test.go`: en `probarLotesAceptados`, «una arista antes que sus extremos» (112-115), con la aserción de
  `ValidarContraGrafoVacio` (134) a la que servía; en `probarRechazosDeUnaOperacion`, «una operacion nula» (250) y
  «una operacion que no es un valor» (251-255), con la frase de su comentario que los anuncia (212-213); los cuatro de
  fecha y el de vigencia negativa de `probarRechazosDeLaProcedencia` (169-193); el nodo con datos sin forma JSON de
  `probarLotesQueNoSeConsolidan` (595-607), que se queda con el lote sin fuente; y los cuatro «un nodo repetido con
  datos menores/mayores, antes/después del primero» de `probarLotesConsolidados` (533-536), con `conMasDatos`,
  `conOtrosDatos` y `losDeMasDatos` (493-498, 512-516) y la frase de su comentario (487-488);
- `errores_test.go`, en `TestRechazo`: «una operacion que no es un valor» (63-67), y «un nodo sin id» (37), «una
  arista» (48-56) y «una arista vacia» (57), que solo nombran operaciones de los rechazos que FR-075 retira; los demás
  nombran las de los rechazos que se quedan (el lote, un nodo, la `Persona`, un texto) y se quedan;
- `almacen_test.go`: de `TestApplyRechazaElLote` (523), la vigencia con fracción de segundo y los datos sin forma JSON
  canónica;
- `persona_test.go`: «un numero que no es finito», «una cadena que no es UTF-8» y «un valor que no es de JSON»
  (215-217), con la constante `motivoDatosSinJSON` (19), que queda sin uso;
- `canonico_test.go`: `TestDatosCanonicosImposibles` entero (202-232).

**Por qué**: FR-074, FR-075 («cada uno con un test y sin enumerar más casos»), FR-076; constitución, «Gates» (umbral
de materialidad: los datos de un emisor que no existe los cubre la regla genérica).

### D21 · Filas dañadas y tests de inutilizables: un solo vehículo

**Decisión**: los tests que enumeran estados que solo deja una mano ajena —`TestApplySobreFilasDanadas`,
`TestLecturaDeFilasDanadas`, `compruebaGrafoIncomprobable` (`internal/app/grafo_test.go:929-948`), el directorio, la
base dañada, los permisos— se retiran; queda «no es una base» como vehículo de la regla genérica en los verbos de
`graph`, en la entrega y en `internal/graph` (`TestIntegracionInutilizables`). También sale el código que solo sirve a
esos estados: `errFilaDanada` y su clasificación aparte (un fallo al fusionar con lo guardado es
`errorDeEntradaSalida`, 1) y las comprobaciones de rango de la vigencia guardada en `vigenciaGuardada` e `indexar`. Las
lecturas de JSON y de fechas guardadas siguen devolviendo su error, que es la regla genérica (1), sin caso en los
contratos ni test propio. Salen, por eso, los tests que las ejercen con lo que ninguna entrega escribe —una fecha de
consulta que no es RFC 3339, guardada o llegando (la que llega ya la rechaza `validarProcedencia`, D20), y unos datos
sin forma JSON, que `world.db` no puede dar porque los guarda como texto JSON—:
- `observacion_test.go`: `historiaImposible` e `historiasImposibles` (296-318) y sus casos en
  `probarNodosQueNoSeFusionan` (407-412), `probarAristasQueNoSeFusionan` (456-461) y `probarTextosQueNoSeFusionan`
  (518-527), con `fechaImposible` y `mensajeFechaImposible` (41-45), que quedan sin uso;
- `comprobar_test.go`: `probarInstantaneasImposibles` (429-460) y su subtest «instantaneas que ninguna lectura da»
  (207), también el caso de la fecha, no solo los de vigencia que pierden su comprobación de rango;
- `salida_test.go`: `probarFichaSinFormaJSON` (132, 136-145);
- `internal/app/grafo_test.go`: con `compruebaGrafoIncomprobable`, su base `baseConUnaFechaIlegible` (2521-2547).

El código que devuelve esos errores (`FusionarNodo`, `FusionarArista` y `FusionarTexto` al leer una fecha, `indexar`
con la fecha guardada, `Ficha.MarshalJSON`) se queda como la regla genérica.
Sale también el subtest «una versión negativa no es una base utilizable» de `TestMigrar` (`migraciones_test.go:214`):
`migrar` solo registra las versiones 1 y 2, así que ese valor solo lo deja quien edita la base a mano. Las guardas que
evitan el `panic` de `lista[registrada:]` (`migraciones.go:134`, `aplicar.go:134`) se quedan como parte de la regla
genérica (constitución IV), sin caso en el contrato ni test propio.
**Por qué**: FR-070 sustituye la enumeración de H7 FR 010 («dañado, directorio, sin permisos») y reduce la matriz de H7
FR 088 y SC 011 a un caso, un fichero que no es una base SQLite.

### D22 · Los tests de espera al leer cambian de vehículo

**Decisión**: `TestLeerConElPlazoAgotado`, `TestLeerBloqueada` y `TestIntegracionPlazoYBloqueo` («al leer») retienen
`world.db` con una conexión sobre la base propia en WAL con `PRAGMA locking_mode=EXCLUSIVE` y una transacción de
escritura abierta (V10), en lugar de una base en modo rollback con `_txlock=exclusive`.
**Por qué**: H7 FR 014 se queda con sus tests (FR-077) y FR-073 prohíbe tratar diarios de rollback y bases de fuera.

### D23 · La redacción C y los relojes

**Decisión**: una derivada nueva, `internal/app/testdata/derivadas/version-ulterior/` (el bloque a21 de la LPAC con
`fecha_vigencia` `20260101` y el párrafo «[Redacción sintética de prueba: versión ulterior derivada de la grabación de
H4.]» al final del texto, con su huella), producida por un programa a partir de la grabación de H4 y fijada por
`TestGrabacionesDerivadas` (una entrada más en `grabacionesDerivadas()`, `internal/app/grafo_test.go:2298-2327`). La
secuencia lee con `$KITLEGAL_T0_BIN` y comprueba con `$KITLEGAL_T1_BIN`; FR-032 comprueba con `$KITLEGAL_T8_BIN`.
**Por qué**: FR-091 (redacciones derivadas de H4, sin grabación nueva). T8 es exactamente T1 + 7 días y en el límite
exacto no hay caducidad (H7, `grafo_test.go:639`): leyendo C con T1, FR-032 no se podría ver con T8, y la `Norma` y el
`Bloque` conservan como última la observación más reciente; leyendo todo con T0, a T8 han caducado los tres. Ningún
reloj nuevo en el arnés.

### D24 · Tests de medida y del `world.db` de H7

**Decisión**: (1) `internal/app/medida_test.go` (`//go:build integration`, `TestMedidaDelGrafo`) siembra con
`graph.Nuevo(graph.ConDirectorio(dir)).Apply` el grafo de la medida (contracts/arnes-e2e.md §4) y ejecuta el binario de
e2e con reloj T8 (`graph check --json` sin argumentos, y con la norma y un bloque). (2)
`internal/graph/integracion_test.go` gana `TestIntegracionGrafoDeH7`: entrega A (T0) y B (T1), devuelve la base a la
versión 1 (sin `lecturas` y sin la fila 2 de `schema_version`, exactamente lo que dejaba H7) y comprueba FR-026 y SC-012.
(3) `TestCosteDelGrafo` (`coste_test.go`) siembra cada versión en su propia entrega, como lecturas sucesivas, y compara
los totales por clase de la `data` nueva.
**Por qué**: FR-092, FR-026, SC-001, SC-002, SC-012; el lote con cuatro versiones de un bloque no lo emite nadie (D2).

### Skills de Go aplicadas

`golang-how-to` orquesta: `golang-database` (transacción inmediata con el upsert de V6, `json_extract` y claves
ajenas, errores con clase), `golang-testing` (tests de tabla, `t.Parallel`, integración con `//go:build integration` y
`t.TempDir()`), `golang-cli` (argumentos de posición con Kong, códigos 0/1/2), `golang-refactoring` (retirada por pasos
con la red de tests, sin reescribir `internal/graph`), `golang-lint` (sin `//nolint` nuevo: el único previsto es el
`paralleltest` que ya lleva `TestCosteDelGrafo` y que `TestMedidaDelGrafo` no necesita, porque no cronometra).

### Datos externos

Ninguno. No hay grabación nueva ni fuente nueva: la redacción C y el estado previo de la eval 19 derivan de la
grabación de H4 (fila de BOE revisada en `docs/SOURCES.md`), sin manifiesto `grabaciones.json` ni test `TestGrabar*`
nuevos, y `grabar_datos` no tiene nada que grabar.
