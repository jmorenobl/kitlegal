# Research: H41 · Las skills del producto mandan leer sus referencias

Decisiones del plan (D), con la alternativa rechazada y por qué, y verificaciones hechas en esta sesión (V), con dónde se
comprobó. Lo que no se pudo comprobar sin red ni modelo va como supuesto no verificado (S).

## Verificaciones

- **V1 · `skills/` es la de `v0.9.0` y `normativa-control` es `skills/normativa`.** `git diff --quiet v0.9.0 HEAD --
  skills/` sale con 0, y `diff -r laboratorio/normativa/normativa-control/skills/normativa skills/normativa` no da
  nada (ejecutadas en esta sesión). Así `<skill>-control` de cada familia se copia de `skills/<skill>` del árbol de
  partida y es, a la vez, la de `main` al lanzar el hito y la de `v0.9.0` (FR-020).
- **V2 · Cómo cuenta `make ci` el cuerpo.** `internal/skills/sincronia.go:576-614`: el cuerpo es lo que sigue a la línea
  `---` que cierra el frontmatter, contado con `utf8.RuneCount`; el defecto es `<skill>: SKILL.md: cuerpo de N
  caracteres (máximo 11000)` (`maximoDeCaracteresDelCuerpo = 11000`, l. 31). Lo comprueba `TestSkillsDelRepositorio`
  (`internal/app/skills_test.go:88`), en `make skills-check` y en `make ci`.
- **V3 · Los cuerpos de hoy y los del hito, medidos.** Con `tail -n +<cierre+1> SKILL.md | LC_ALL=en_US.UTF-8 wc -m`
  sobre el árbol y sobre los borradores que dan los cambios de contracts/remisiones.md, aplicados a copias en el
  directorio temporal: `normativa` 10 929 → **10 981**; `plazos-y-territorio` 10 930 → **10 982**; `jurisprudencia`
  10 842 → **10 842**; `normativa-remisiones-en-orden`, 11 084 (la cifra del ADR 0051). `consulta-juridica` no cambia.
- **V4 · Las suites congeladas que leen el texto de las skills.** `internal/app/testdata/script/h32-normativa.txtar`
  (congelada, en `TestEntregaDelHito`) exige: en el paso 2 de `normativa`, la frase «Habla de otras conversaciones, que
  quien pregunta no conoce: un bloque sin entrada `version-obsoleta` no da nada a la respuesta, ni una línea, ni una
  frase junto a los avisos o a la fecha de vigencia.»; los pasos 3 y 6 y la primera viñeta del paso 4, línea a línea;
  **las cinco formas fijas —también `⚠ REDACCIÓN MODIFICADA:` y `⚠ SIN CONSULTA AL BOE:`— como líneas de un bloque
  `text` del `SKILL.md`** (`formas: 5 de 5`); la tabla de comandos; y ningún cuerpo de capacidad con una norma de la
  lista mínima. `h32-paquetes.txtar` exige el bloque generado de `plazos-y-territorio` entre sus marcas y ningún cuerpo
  con una norma del paquete. `grep` de las frases que el hito cambia («tras leer», «Antes de dar la consulta»,
  «repásalas», «Nada de otra conversación», «Cuándo leer cada una») en `internal/`, `cmd/`, `scripts/` y los guiones no
  da ninguna comprobación que las fije. Los cambios de contracts/remisiones.md respetan todo eso.
- **V5 · El sondeo de H40 en el repositorio.** `evidencias/sondeos-20261104/tabla.md`: `normativa-control` abre
  `como_se_cita.md` en 0 de 40 en los dos modos y `redaccion_en_una_fecha.md` en 10 de 40 en modo orden (las diez de la
  eval 22) y 3 en modo herramienta; `normativa-remisiones-en-orden`, 20 de 40 `como_se_cita.md`.
- **V6 · Lo que trae `kitlegal` en los bloques de las evals 04, 05, 15 y 22 de `normativa`.** Ids `a66`, `a59`, `a17` y
  `a1-30` (los YAML de las evals): ninguno termina en punto. Avisos: `internal/source/boe/avisos.go:39-41` avisa de una
  consolidación no finalizada con el código `4`; las grabaciones de metadatos de la LGT, el TRLRHL y la LIRPF
  (`testdata/evals/boe.legislacion-consolidada/…_metadatos.json`) y la de la LCSP
  (`internal/source/boe/testdata/boe.legislacion-consolidada/…BOE-A-2017-12902_metadatos.json`) traen el código `3`,
  `estatus_derogacion` `N` y `vigencia_agotada` `N`. No se ha localizado en esta sesión la copia exacta de los metadatos
  de la LCSP que sirve la sesión de la eval 22; si la sirviera otra, no cambia la aceptación (D12).
- **V7 · Las sesiones del arnés se preparan como las del job.** `ejecutarSesiones` llama a `PrepararSesion`
  (`internal/evals/sesiones.go:439`), que prepara la caché con las consultas de todas las evals y `data/`, el grafo
  previo de la eval y quita lo que su `sin_servir` no sirve (`internal/evals/preparar.go:344-360`): la eval 11 de
  `plazos-y-territorio`, con `sin_servir`, se prepara en el arnés como en el job.
- **V8 · Biblioteca estándar.** `go doc os.ReadDir`: «returning all its directory entries sorted by filename».
  `go doc os.CopyFS`: copia el árbol, no sobrescribe ficheros que existen, «Symbolic links in dir are followed».
  `go doc encoding/json.Unmarshal`: «object keys which don't have a corresponding struct field are ignored».
  `go doc encoding/json.Marshal`: `omitempty` omite un slice de longitud cero. Go 1.27.3 (`go version`; `go.mod`: `go
  1.27.0`, `toolchain go1.27.3`).
- **V9 · testscript y `go test -run`.** `go doc github.com/rogpeppe/go-internal/testscript` (v1.16.0, del `go.mod`):
  `grep` y `cmp` son órdenes integradas; `grep`, `stdout` y `stderr` compilan su patrón con `(?m)` delante
  (`testscript/cmd.go:690` en la caché de módulos), así que `^` y `$` son de línea y `\n` y `\s` cruzan líneas. El nombre
  del subtest es el del fichero sin `.txtar` (`testscript.go:323-341`), y `go help testflag` dice que `-run` se parte por
  las barras que no van entre corchetes: `go test -run '^TestEntregaDelHito$/^h40-' ./internal/app/` ejecuta solo el
  guion de H40 (ejecutado en esta sesión, pasa). `.goreleaser.yaml` construye solo `main: ./cmd/kitlegal` (l. 30).
- **V10 · El arnés e2e ya tiene `cmd/laboratorio`.** `internal/app/e2e_test.go:171` (`KITLEGAL_LABORATORIO_BIN`, de H40):
  los guiones del hito no cambian el arnés.
- **V11 · `internal/skills` y sus enlaces.** Lo importan solo paquetes de `internal/evals` y tests (`grep -rl` de su ruta
  de importación fuera de `_test.go`): no va en `kitlegal`. Sus líneas fuera de los bloques de código y sus enlaces los
  dan `lineasFueraDeBloques` y `enlacesDeMarkdown` (`internal/skills/markdown.go:46-108`), y
  `defectosDeEnlacesEIndices` (`internal/skills/referencias.go:267`) ya exige que el cuerpo enlace cada referencia. Sus
  tests del repositorio leen `../..` (`instalacion_test.go:43`, `territorio_test.go:263`). Los fixtures sintéticos de
  `sincronia_test.go` (l. 74-76, 1028), de `paquetes_test.go` (l. 198) y la copia de `internal/app/skills_test.go`
  (l. 1275-1311) enlazan referencias sin «lee».
- **V12 · Los umbrales del job que deciden.** Los informes del cierre de H40 en el repositorio
  (`specs/032-h40-por-qu-e/gates/evals/*.json`, `jq` de `umbrales[] | select(.decide == true) | .nombre`) dan los
  nombres de la sección «Controles de umbral» del plan; `consulta-juridica` no publica ninguno con `decide: true`.
- **V13 · Lo que ya funciona hoy en los escenarios de quickstart.md.** En esta sesión: `make skills-check` (verde),
  `go test -count=1 -run '^TestEntregaDelHito$/^h40-' ./internal/app/` (verde) y `go run ./cmd/laboratorio tabla
  internal/evals/testdata/laboratorio/normativa/sondeo` con salida igual, byte a byte (`cmp`), a su `tabla.md`; el árbol
  quedó sin cambios (`git status --porcelain` solo da el directorio del feature).

## Supuestos no verificados

- **S1.** Que con las órdenes del hito `normativa-producto` abra `con_herramientas.md` y `normas.md` como
  `remisiones-en-orden` en modo herramienta, y que ninguna familia quede por debajo de su `control`: es la aceptación de
  la persona (FR-052), con sesiones con modelo que esta sesión no abre.
- **S2.** Que las evals de las tres skills sigan en verde en el cierre (FR-051): lo mide el job del cierre.

## Decisiones

### D1 · Las remisiones de `normativa` (FR-001 a FR-004)

**Decisión.** El punto de partida son las de `normativa-remisiones-en-orden` (contracts/variantes.md §3 de H40), con lo
que trae cada referencia y el sitio que pide FR-003; el texto exacto, en contracts/remisiones.md §2:

| Referencia | Sitio | Condición | Lo que trae |
|---|---|---|---|
| `con_herramientas.md` | párrafo del «Protocolo», antes de la primera orden o llamada | ninguna | cómo se pide cada una con herramienta |
| `normas.md` | paso 1, antes de toda orden | ninguna | su tabla, con el `BOE-A-…` |
| `buscar_la_norma.md` | paso 1 | la norma no está en la tabla | cómo buscarla («búscala como dice») |
| `redaccion_en_una_fecha.md` | paso 2, viñeta de la lectura | una fecha en la pregunta; antes de leer el bloque | cómo se lee |
| `redaccion_modificada.md` | paso 2, viñeta de la comprobación | cada entrada `version-obsoleta`, o se pregunta qué cambió | qué se dice |
| `como_se_cita.md` | paso 2, viñeta propia | un bloque con avisos o un id que termina en punto | cómo se citan |
| `varios_bloques.md` | paso 2, viñeta de varios bloques | antes de leer varios bloques a la vez | su orden |
| `lecturas_que_fallan.md` | paso 2, viñeta de varios bloques | una orden que no termina con `0`; antes de seguir | qué hacer |

**`como_se_cita.md` (FR-004).** Su orden va en el paso 2, con la condición que pone el caso —un bloque con avisos o con
un id que termina en punto—, porque lo que la referencia añade a la forma que el cuerpo ya da son esos casos («Ids que
terminan en punto», «Los avisos»); los demás apartados repiten el cuerpo, que conserva la forma que toda respuesta usa
(ADR 0046, punto 1; ADR 0050, punto 1). Rechazadas: (a) «antes de escribir la respuesta», la de
`remisiones-en-orden`, que el ADR 0051 (punto 3) descarta: 20 de 40; (b) sin condición antes de la primera orden o
llamada: la leería toda pregunta (1961 bytes más) y el spec («Uso, de fuera adentro») solo manda leer así
`con_herramientas.md` y `normas.md`; (c) en «Cómo se cita»: no es un paso ni va antes de una llamada (FR-003).

**`redaccion_modificada.md`.** Su orden sale de la sección «Redacción modificada», que no es un paso y va detrás del
protocolo, al paso 2, donde `kitlegal graph check` trae la entrada (FR-003); la sección conserva su encabezado y la forma
fija en su bloque `text` (V4). Rechazada: dejarla en la sección con lo que trae, que no cumple FR-003.

**Las de `lecturas_que_fallan.md`, `redaccion_modificada.md` y `como_se_cita.md`** dependen de lo que devuelve una
llamada, así que van en el paso de esa llamada, para leerlas cuando la llamada trae su caso; FR-003 lo admite («en el
paso que usa la referencia»). La de `lecturas_que_fallan.md` lo dice con «antes de seguir», como en
`remisiones-en-orden`; en las otras dos lo dicen el sitio y la condición, porque el cuerpo no tiene margen para repetirlo
(V3: 19 caracteres).

### D2 · El sitio en el cuerpo de `normativa` (FR-008, FR-009)

**Decisión.** Un movimiento: la viñeta **«Nada de otra conversación»** del paso 4 pasa, entera, al final de
`references/redaccion_modificada.md` (contracts/remisiones.md §3). Con él y las remisiones, el cuerpo queda en 10 981
caracteres (V3). Por qué ella: lo único que trae de otras conversaciones a una respuesta es una entrada
`version-obsoleta` de `kitlegal graph check`, y ese es el caso en que el paso 2 manda leer esa referencia; sin entrada,
lo cubre la frase del paso 2 que fija `h32-normativa` («un bloque sin entrada `version-obsoleta` no da nada a la
respuesta»), que no se mueve. No la necesita toda respuesta (ADR 0050 y 0051, consecuencias), y la clase del juez que
mira lo que se cuenta del proceso no decide en `normativa` (V12).

**Alternativas rechazadas.**
- Mover una forma fija (las líneas de «Redacción en una fecha» o de «Redacción modificada»): `h32-normativa`, congelada,
  exige las cinco formas en un bloque `text` del `SKILL.md` (V4), y `redaccion_en_una_fecha.md` y
  `redaccion_modificada.md` dicen que la forma está en el `SKILL.md`, frase que FR-007 no deja cambiar; además, la de la
  fecha la mide un umbral que decide (`redaccion_sin_fechar`).
- Los pasos 3 y 6, la primera viñeta del paso 4 y las dos frases del paso 5: fijados por `h32-normativa` (V4).
- La regla 7 (sin binario): su forma va en un bloque `text` del cuerpo (V4), y separar la regla de su forma la debilita
  justo donde no hay nada consultado.
- La variante de PowerShell 5.1 a una referencia nueva: ahorra unos 80 caracteres netos, no basta sola y añade una
  referencia y una columna a la tabla; con el movimiento elegido no hace falta.
- Recortar las remisiones hasta caber sin mover nada: con «lo que trae» de cada una, el cuerpo pasa de 11 000 (V3).

### D3 · Las remisiones de `plazos-y-territorio` (FR-005)

**Decisión.** Las tres de la lista de «Referencias» pasan a ser órdenes en el mismo sitio, con la condición que hoy da la
lista, el verbo «lee» antes del enlace y lo que trae (contracts/remisiones.md §4). La sección va antes del «Protocolo»,
antes de la primera orden o llamada (FR-003), donde está la orden de `administrativo.md`, generada, que no cambia (FR-005)
y se abre en 58 de 60 sesiones en modo herramienta (#214, citado por el hito). Lo que trae va corto —«rangos y reglas»,
«qué hacer», «su detalle»— porque el cuerpo tiene 70 caracteres de margen (V3: 10 982).

**Alternativa rechazada.** Llevar cada orden a su paso (5, 4 y 8 a 10): son seis sitios, no caben sin mover texto del
cuerpo, y FR-003 ya se cumple con la orden antes de la primera llamada.

### D4 · Las remisiones de `jurisprudencia` (FR-006)

**Decisión.** La de `consulta_preparada.md` pasa del final del paso 2 a la frase que nombra `cita_preparar`, delante de la
tabla de lo que se pide: «Antes de pedir una u otra, lee …» —la herramienta o su orden—, con lo mismo que trae hoy. La de `documento_aportado.md` pasa, sin
cambiar una letra, del final del paso 3 a su principio, delante de la viñeta que pide el cotejo. La del paso 6 no cambia:
ya es una orden, con su condición, antes de la llamada y en la misma frase. Cuerpo: 10 842 (V3).

**Por qué mover también la de `documento_aportado.md`.** FR-003 pide la orden «en el paso que usa la referencia y antes de
la orden o llamada de ese paso»; al final del paso 3 está detrás de la viñeta que manda cotejar lo primero. Moverla no
cuesta un carácter y la deja donde la de `consulta_preparada.md`. Rechazada: dejarla, que cumple la lectura temporal
(«antes del cotejo») pero no la de sitio.

### D5 · El control de FR-001 a FR-003 es un test del repositorio en `internal/skills` (SC-001)

**Decisión.** `TestRemisionesDelRepositorio` (`internal/skills/remisiones_test.go`, en el paquete, junto a
`TestNormasDelRepositorio` y los demás que leen `../..`): por cada skill de `skills/` con referencias —hoy, las tres del
hito— y cada enlace del cuerpo a `references/<f>` fuera de los bloques de código, su frase tiene la palabra «lee» o
«Lee» antes del enlace y no tiene «si la tienes», «antes de escribir la respuesta» ni «antes de dar la consulta», en
mayúscula o minúscula (D6); y cada
fichero de `references/` tiene al menos un enlace (contracts/remisiones.md §6). Usa el lector de líneas fuera de bloques
y la expresión del enlace de `internal/skills/markdown.go` (V11), sin cambiar código del paquete. Falla con
`<skill>: references/<f>: <regla>`. Un subtest aplica la misma regla al `SKILL.md` de cada `<skill>-control` y exige
exactamente sus fallos, así que el control se ve fallar por encima del umbral con las remisiones de antes. Corre en
`make ci` y en `make skills-check`, cuya expresión de tests gana su nombre.

**Alternativa rechazada.** Un defecto nuevo de `skills-check` en `sincronia.go`: los fixtures sintéticos de
`internal/skills/sincronia_test.go` (l. 74-76, 1028) y la copia con enlaces desnudos de `internal/app/skills_test.go`
(l. 1275-1311) enlazan referencias sin «lee», y sus listas exactas de defectos cambiarían en tests que el hito no pide
tocar; la regla es del contenido de las skills del repositorio, no de la forma que cualquier skill tiene que tener.

### D6 · Qué es la frase de un enlace (FR-001)

**Decisión.** La de FR-001, con dos precisiones que el lector de `internal/skills` necesita: una línea de un bloque de
código cuenta como línea en blanco (un bloque corta la frase; ningún enlace a una referencia va dentro), y «la marca de
un paso» es un encabezado (`#`) o un elemento de lista numerada (`1. `), como la de una viñeta (`- `, `* `). «Lee» es
la palabra entera (no vale «leer» ni «léelo»), y las tres expresiones se buscan con los blancos y saltos de línea
juntados en un espacio, porque el cuerpo va cortado a 120 columnas, y sin distinguir mayúsculas de minúsculas, como
«lee» o «Lee» en el verbo: la que el hito quita de `jurisprudencia` abre su frase («Antes de dar la consulta, lee
[references/consulta_preparada.md]…», `skills/jurisprudencia/SKILL.md:78`), y el subtest «control» la tiene que ver
fallar. Rechazadas: dejar que una frase cruce un bloque de código, que daría por buena una orden de leer escrita antes
del bloque para un enlace escrito después; y buscar las expresiones tal como se escriben en minúscula, que daría por
buena esa frase en `skills/` y haría fallar el subtest «control» contra su propia lista.

### D7 · El control de FR-007 y FR-009 calcula el producto desde el control (SC-004)

**Decisión.** Un test de `internal/evals` aplica a cada `<skill>-control` su declaración —las remisiones, como
sustituciones de bytes que están una sola vez en el `SKILL.md` de control, y los movimientos, como texto que sale del
`SKILL.md` y se añade al final de su referencia— y exige el resultado, byte a byte, en `skills/<skill>`; y por cada
movimiento, que su texto, con los blancos juntados y los enlaces relativos a la referencia, sea el que llega a la
referencia y que ya no esté en el cuerpo (contracts/remisiones.md §6). Es el mecanismo de `TestVariantesDeNormativa` de
H40, sobre otro par. **Queda fuera lo que genera `make skills-sync`** —las referencias con la cabecera «generado desde» y
las regiones entre las marcas de los paquetes y de la tabla de comandos—, que gobierna FR-010 con
`TestSkillsDelRepositorio`.

**Por qué fuera lo generado.** Una materia entra como datos, sin tocar el código (principio X): con lo generado dentro,
un paquete nuevo en `data/normas.yaml` cambiaría `normas.md` y el bloque de los paquetes y pondría este test en rojo.
Rechazada: comparar también lo generado.

### D8 · El control de `<skill>-control` es una lista de huellas (FR-020, SC-008)

**Decisión.** Un test con la huella SHA-256 de cada fichero de las tres `<skill>-control`, por ruta, escrita al copiarlas
(V1): falla con la ruta del que difiere, falta o sobra. Rechazadas: comparar con `skills/<skill>`, que deja de valer al
cambiar el producto (FR-020: «deja de compararse con `skills/normativa`»); y con `git show v0.9.0:…`, porque el `make ci`
de la entrega corre sobre un clon superficial (lo dice `h32-normativa.txtar`, l. 21) y no puede contar con la etiqueta.

### D9 · Las familias del producto, declaradas en Go (FR-020, FR-022, FR-030)

**Decisión.** Como la de `normativa` en H40 (research D3 de H40): `familiaDelLaboratorio` gana un campo, `producto` —el
nombre de la variante que se instala desde `skills/`, vacío en la de H39—, y hay tres familias del producto,
`normativa`, `plazos-y-territorio` y `jurisprudencia` (data-model §1). Las dos variantes nuevas de cada familia se miden
en los dos modos. El fichero `internal/evals/laboratorio_normativa.go` pasa a llamarse `laboratorio_familias.go`, porque
declara las tres. Rechazada: una declaración en YAML como la de H39, que añadiría un lector y un esquema para datos que
fija entero el hito.

### D10 · El arnés instala `<skill>-producto` desde `skills/` y deja la lista de sus referencias en `sondeo.json`

**Decisión.** Al preparar un sondeo con `<skill>-producto`, el arnés copia `skills/<skill>` del árbol en
`<temporal>/skills/<skill>-producto/<skill>` con `os.CopyFS` (V8) y escribe en `sondeo.json` la clave
`referencias_del_producto`: los nombres de los ficheros de `references/` de la copia, en orden de nombre (`os.ReadDir`,
V8). La clave es un campo más de `sondeoLeido` (`internal/evals/laboratorio_tabla.go`), el tipo con el que el arnés
escribe hoy el fichero y la tabla lo lee: un solo tipo para `sondeo.json`. Sin `<skill>-producto` en la petición la
clave no va (`omitempty`), y `sondeo.json` es el de hoy byte a byte. Los
guiones de H39 y H40 leen de `sondeo.json` el modelo y la versión y no ven la clave (V8: claves sin campo, ignoradas),
así que sus suites siguen igual (FR-033, FR-042). Rechazadas: un fichero aparte (otro lector y otro defecto), un segundo tipo en `laboratorio_sondeo.go` para escribirlo
(dos contratos del mismo fichero) y una copia
o una huella de la skill (fuera de alcance).

### D11 · El error de una mezcla de familias

**Decisión.** `VARIANTES: se mezclan variantes de <a>, variantes de <b> y variantes de <c>`, una por familia pedida, en el
orden de las familias —la de H39 se nombra «del laboratorio de H39»—, de modo que la mezcla de H39 y `normativa` da
exactamente el texto de H40 y `TestArnesDeNormativa` no cambia. Rechazada: un texto nuevo para todas las mezclas.

### D12 · Las evals del sondeo y las referencias a cuya condición llegan (FR-030, FR-052)

Criterio del hito: evals que ya tienen sus lecturas grabadas y que, entre las cuatro, llegan a la condición del mayor
número de remisiones; a igualdad, la que llega a más condiciones de una misma remisión. Ninguna es de no activación
(`activa: false`) ni de sesión sin binario (`sin_binario_ni_servidor`), que no llegan a ninguna remisión. Todas están en
`evals/<skill>/`, y `make skills-check` comprueba «lo grabado que necesitan» las evals (CONTRIBUTING.md, l. 196;
`TestEvalsDelRepositorio`, verde en esta sesión, V13).

| Familia | Eval | Referencias a cuya condición llega |
|---|---|---|
| `normativa` (las de H40) | 04, 05, 15 | `con_herramientas.md`, `normas.md` |
| | 22 (con fecha) | `con_herramientas.md`, `normas.md`, `redaccion_en_una_fecha.md` |
| `plazos-y-territorio` | 02-territorio-municipio-no-cubierto | `municipios.md` (presenta un territorio y explica una cobertura no configurada) |
| | 05-plazo-reposicion-leganes | `administrativo.md`, `calculo_de_plazos.md` (los municipios que cuentan) |
| | 10-plazo-reposicion-anterior-a-la-lpac | `administrativo.md`, `calculo_de_plazos.md` (plazo sin calcular, el `3`) |
| | 11-plazo-reposicion-sin-leer-el-precepto | `administrativo.md`, `calculo_de_plazos.md` (lectura fallida, por `sin_servir`; V7) |
| `jurisprudencia` | 01-existe-con-numero-y-fecha | `consulta_preparada.md` (paso 2) |
| | 03-por-materia | `consulta_preparada.md` (paso 6) |
| | 04-documento-pegado | `documento_aportado.md` |
| | 06-documento-que-no-es-el-pedido | `documento_aportado.md` y `consulta_preparada.md` (paso 3 y, tras él, paso 2) |

**No llegan:** en `normativa`, `buscar_la_norma.md`, `lecturas_que_fallan.md`, `redaccion_modificada.md` y
`varios_bloques.md` (ADR 0051) y `como_se_cita.md` (V6: ningún id con punto ni ningún aviso); en `plazos-y-territorio`,
`jerarquia_normativa.md` (ninguna de sus doce evals pregunta por el nivel que regula algo ni por la norma que prevalece).
Salen «0 de n» en la tabla y no entran en la comparación de FR-052. En `jurisprudencia` llegan las dos. Las de
`plazos-y-territorio` 05, 10 y 11 sitúan los municipios con los pasos 1 a 4, pero si eso es «presentar un territorio»
lo decide el agente: `municipios.md` cuenta como alcanzada por la 02. `jurisprudencia` no pide nada a la red (ADR
0036): sus evals no necesitan grabaciones.

**Rechazadas.** En `plazos-y-territorio`, la 01 en lugar de la 02 (las dos presentan un territorio; solo la 02 explica
una cobertura no configurada) y la 06 o la 09 en lugar de la 10 o la 11 (llegan a «los municipios que cuentan», como la
05, y no al plazo sin calcular ni a la lectura fallida). En `jurisprudencia`, la 02 o la 07 en lugar de la 01 (el mismo
paso 2) y la 08 o la 09 en lugar de la 04 (el mismo paso 3).

### D13 · La regla de lectura con `cd` (FR-041)

**Decisión.** En las familias del producto, una orden de `Bash` abre `<f>` si, partida en órdenes simples como hoy
(`ordenesSimples`, `internal/evals/laboratorio_tabla.go:1197`), tiene una cuya primera palabra que no es una asignación
es `cd` y cuya palabra siguiente, sin una barra final, es `<skill>/references` o termina en `/<skill>/references`, y
detrás, en otra orden simple de la misma orden, una palabra que es exactamente `<f>` o `./<f>`. «Termina en» se lee por
tramos de ruta, con la misma forma que la regla de `Read` de H40 (`<skill>/references/<f>` o terminada en
`/<skill>/references/<f>`), para que las dos digan lo mismo del directorio de la skill. Un `cd` posterior a otro sitio
no se modela: ninguna sesión de H40 lo hace y no cambia ninguna celda que una persona lea. La regla de H40 —la orden
contiene `references/<f>`— sigue aparte. La familia de H39 no la usa (FR-041; clarificación Q1). Rechazada: comparar
el texto del final sin tramos, que da otra regla que la de `Read`.

### D14 · Las referencias de la tabla (FR-040)

**Decisión.** Antes de contar, el guion lee cada sondeo: por cada directorio de variante versionada suma las referencias
de la familia (su declaración en Go) y por cada directorio de `<skill>-producto`, la lista de su `sondeo.json`; las
columnas son la unión, en orden de nombre. Una variante «tiene» una referencia si está en su declaración o, la de
producto, en la lista de alguno de sus sondeos; donde no la tiene, la celda de la fila y la del detalle van vacías. La
lista la escribe el arnés siempre que instala `<skill>-producto` (contracts/arnes.md §3.4): sin ella, o con una que no
es una lista de nombres, el `sondeo.json` se ha editado a mano (FR-040) y es lo que el guion no sabe leer, con la regla
genérica de H39 —defecto `inesperado`, código 1 y sin tabla (ADR 0023)—, sin mensaje ni test propios. Rechazadas:
columnas fijas de la declaración, que no darían la de una referencia nueva del producto (spec, «Edge Cases»); y un
mensaje y un caso de test para la lista que falta, que solo se alcanza editando el sondeo a mano.

### D15 · Aceptación e2e

**Decisión.** Dos guiones congelados (contracts/aceptacion.md): `laboratorio-tabla.txtar`, que ejecuta `laboratorio
tabla` sobre sesiones fijas de las tres familias, con lecturas con `cd` que cuentan y que no, y un directorio de otra
familia; y `remisiones.txtar`, que instala las skills con `kitlegal skills install --dir` y comprueba en lo instalado
cada orden de leer, que no queda ninguna remisión de las de antes y dónde está lo movido. El arnés (historia 3) abre
sesiones con modelo y credencial: su aceptación son tests de `make ci` con el sustituto de `claude`, como en H40.

### D16 · Lo que el hito documenta (FR-050)

`CHANGELOG.md` (*Unreleased*) da el cambio de las tres skills. `docs/SKILLS.md` recoge lo que la guía no tiene: en §1,
«Cómo se remite», que una orden cuya condición la da lo que devuelve una llamada va en el paso de esa llamada, detrás de
ella («antes de seguir»), que el sitio para las órdenes se hace llevando a la referencia que se manda leer en ese caso
lo que solo necesita ese caso, y que `make ci` y `make skills-check` comprueban la frase de cada remisión (D5). Además,
lo que hoy dice algo que
el hito deja de hacer cierto —`normativa-control` es `skills/normativa` (CONTRIBUTING.md, «El sondeo del laboratorio»),
«las dos familias» (la ayuda de `make evals-laboratorio` y `make laboratorio-tabla`, la cabecera de
`scripts/evals-laboratorio.sh` y el comentario de paquete de `cmd/laboratorio`)— se corrige en el mismo cambio.

## Fuentes externas

Ninguna. El hito no graba nada ni toca ninguna fuente (FR-031): no hay manifiesto `grabaciones.json`, test
`TestGrabar*` ni tarea de grabación; las sesiones del sondeo se sirven de las grabaciones que ya usa el job, y nada se
escribe en `evidencias/` ni en `data/` desde el run.

## Dependencias

Ninguna nueva: biblioteca estándar (`crypto/sha256`, `encoding/json`, `os`, `regexp`, `unicode`) y lo que ya usan
`internal/skills` e `internal/evals`; `stretchr/testify` y `rogpeppe/go-internal` en los tests (constitución V).
