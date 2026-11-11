# Research: H41.2 · Por qué la eval 04 no activa `normativa` en modo herramienta en el job

Decisiones (D), lo comprobado en esta sesión contra el repositorio o la toolchain local (V) y lo que no se ha podido
comprobar sin red ni modelo (S). Cada decisión, con la alternativa rechazada y su porqué (constitución, «Criterio de
decisión autónoma», 1 a 4). Las cifras de partida del spec (las cuatro caídas, los 0 de 60 de H39 y H40, las 70 sesiones
con 2.1.317) no se han medido en esta sesión: son las del hito y de la D21 de `specs/034-h41-las-skills-del/research.md`.

## Fuentes externas

Ninguna. El hito no graba ni lee ninguna fuente: no hay manifiesto `grabaciones.json`, test `//go:build grabacion`
`TestGrabar*` ni tarea que grabe; el run no escribe en `evidencias/` ni en `data/`. Las sesiones del sondeo de la persona
(FR-030) se sirven de las grabaciones que ya usa el job, como en H39, H40 y H41. `docs/SOURCES.md` no cambia.

## Decisiones

### D1. Una sola lectura de los usos con su turno, para el informe del job y para la tabla

**Decisión.** La lista de FR-001 la calculan el informe del job y la tabla con el mismo código: `leerLosUsos` —hoy en
`internal/evals/laboratorio_tabla.go:877-941`, que da cada bloque `tool_use` del transcript con el turno de H39— y una
función nueva, `activacionesDeLosUsos`, que con esos usos y la regla de la primera herramienta del modo da la lista
(data-model §1). Las dos pasan, con lo que necesitan de la tabla (`usoDelTranscript`, `usosDelMensaje`, `conResultado`,
`bloqueConTipo`, `activaAntes`, `skillDelUso` y la regla de la primera herramienta, D2), a `internal/evals/activaciones.go`.
`leerLosUsos` deja de exigir `message.id` en cada mensaje `assistant`: el turno se decide con la misma regla y el id tal
como se lee, también vacío.

**Por qué.** FR-001 manda que «turno», «activar», «primera herramienta» y «antes» sean los de la tabla de H39, y el spec
(«Assumptions») pide que el informe y la tabla «lean la activación igual»: con un solo código no pueden divergir. La
exigencia de `message.id` se quita porque el informe lee la lista de sesiones que `LeerSesion` ya ha leído, y `LeerSesion`
no la exige (`internal/evals/sesion.go:580-628`, `leerAssistant`): con ella, una sesión que hoy es legible pasaría a
ilegible, y FR-004 pide lo contrario. Los transcripts que escriben en Go los tests del informe no llevan `message.id`
(V4), así que con la exigencia cambiarían los motivos y el veredicto de esos tests. En todo transcript con `message.id`
—los de Claude Code: V5, S1—, la regla da los mismos turnos que hoy, y la tabla de todo sondeo del arnés sale igual
(FR-022; lo comprueban las suites congeladas de H39, H40 y H41 y `TestTablaDeLasFamilias`, `TestTablaDeNormativa` y
`TestTablaDelLaboratorio`, sin cambiar un fichero ni un valor esperado). Ningún test ni guion congelado prueba la
exigencia (`grep -rn 'message.id'` sobre `internal/evals/*_test.go`, las tres carpetas `aceptacion/` y
`internal/app/testdata/script/` no da nada en esta sesión). Un transcript sin `message.id` no lo escribe Claude Code: el
plan no le promete ni le prueba nada (umbral de materialidad).

**Alternativas rechazadas.**
- *Un campo nuevo en `Sesion`, calculado en la pasada de `LeerSesion`*: `Sesion` se compara entera en
  `TestLeerSesion` y en los tests de `internal/evals/sesion_test.go:234`, `:428` y `:1550` (V6), y un campo poblado
  cambiaría sus valores esperados; y `LeerSesion` no sabe el modo, que la regla necesita (D2).
- *Una segunda regla de turno dentro de `leerTranscript`*: dos implementaciones de la misma definición, que es lo que
  FR-001 quiere evitar.
- *Mantener la exigencia y dejar la lista sin dar si falla*: o silencia un error (constitución, «Criterio», 1) o, si
  la sesión pasa a ilegible, cambia el veredicto de los tests del informe (FR-004).

### D2. La primera herramienta de kitlegal, por modo, con la regla de la tabla

**Decisión.** La regla de `esDeKitlegal` y `consultaKitlegal` (`laboratorio_tabla.go:723-779`) pasa, sin cambiar, a un
tipo pequeño con las herramientas del registro y el intérprete de las órdenes (`reconocedorDeKitlegal`, data-model §1),
que usan la tabla y el informe. En el modo herramienta, un uso cuyo nombre, o lo que sigue a su último `__`, es una
herramienta del registro (no lo es `ToolSearch`); en el modo orden, una orden de `Bash` con una orden simple que invoca
`kitlegal` con un verbo que da herramienta y consulta (no lo son `which kitlegal`, `--help` ni `skills list`). En el modo
de una eval sin binario ni servidor —el `""` de `modoDeLaSesion` (`sesion.go:355-368`)—, ningún uso lo es, y toda
activación de esa sesión es «antes».

**Por qué.** FR-001 define la primera herramienta en los dos modos y dice que «en una sesión sin ninguna, toda activación
es antes»; el caso límite del spec pone como ejemplo «una eval sin binario ni servidor». El job tiene esas sesiones:
`evals/normativa/21-sin-binario-ni-servidor.yaml` y `evals/plazos-y-territorio/04-sin-binario-ni-servidor.yaml` (V15).
La tabla solo cuenta los dos modos (`contarLaVariante` rechaza otro, `laboratorio_tabla.go:551-556`), así que su
resultado no cambia.

**Alternativa rechazada.** *Contar en esas sesiones una orden `kitlegal` de `Bash`, como en el modo orden*: la sesión no
tiene binario, y el spec la pone como ejemplo de sesión sin herramienta de kitlegal.

### D3. La forma de la lista en `informe.json`

**Decisión.** Cada elemento de `evals` lleva, detrás de `activada`, `activaciones`: una lista de objetos
`{"skill": <input.skill tal cual>, "antes": <bool>}` en el orden del transcript; `[]` si la sesión no activó ninguna;
y la clave no está si la lista no se leyó (D5). En Go, `Activaciones []Activacion` con `json:"activaciones,omitzero"`.

**Por qué.** `encoding/json/v2`, el de `informe.go` (`informe.go:1142`), escribe una lista nil como `[]` y `omitzero`
omite solo la nil (V1): así «no leída» (sin clave) no se confunde con «ninguna» (`[]`), como pide FR-005, y una sesión
sin la clave se lee igual que una de un informe anterior a este hito. Es el precedente de `descripcion,omitzero`
(`informe.go:329`). El booleano es el hecho que pide FR-001 («si va antes … o no»); «antes»/«después» es su texto en
`informe.md` (D4).

**Alternativas rechazadas.** *Un puntero, `null` si no se leyó*: otra forma para lo mismo que ya da `omitzero`, y un
lector tendría que distinguir `null` de ausente en informes de antes y de después. *Un texto `"antes"`/`"después"` en
JSON*: dos valores de texto para un booleano.

### D4. La lista en `informe.md`

**Decisión.** En la sección de cada sesión, detrás de «Modelo de la sesión:», un párrafo
`Skills activadas: <skill> (antes|después), …` en el orden de la lista; `Skills activadas: ninguna` con la lista vacía;
`Skills activadas: sin leer` sin lista, con el `sinLeer` con el que la sección dice hoy lo no leído
(`informe.go:85`, `:2719`). El mismo texto lo da la tabla en su detalle (contracts/tabla.md §3), con una sola función
(`textoDeLasActivaciones`, data-model §1).

**Por qué.** FR-002 pide la lista en la sección de la sesión y, vacía, que diga que no activó ninguna; FR-004 pide que
la tabla de sesiones, con su columna «Activada», no cambie, así que no se le añade columna.

### D5. Qué hace el informe si no puede leer la lista

**Decisión.** La lista se da de toda sesión que `LeerSesion` leyó, también de la que es ilegible por otra causa (su eval,
su traza o su modo), igual que `activada` (`informe.go:1465-1473`). Si leerla falla, el error es un motivo de sesión
ilegible más, `sesión ilegible: <error>`, y la lista no se da.

**Por qué.** FR-005: «de una sesión ilegible no se da, como lo demás leído de ella». Tras `LeerSesion`, `leerLosUsos` sin
la exigencia de D1 solo falla con un transcript que Claude Code no escribe (un `message.id` que no es texto, un mensaje
cuyo `message` no es un objeto): es la regla genérica de lo que llega manipulado, y no se diseña ni se prueba aparte.

### D6. Cómo se comprueba que lo demás del informe no cambia

**Decisión.** Dos controles. `TestInformeConLasActivaciones` fija, de las sesiones de FR-006, `activada`, la medida y el
total de `sin_activar` de cada modo, las tasas y el veredicto con los valores que dan hoy sus reglas (contracts/informe.md
§5). Y `TestInformeDeLaSuiteCompletaNoCambia` (`informe_test.go:8376-8400`), que fija byte a byte el informe de la suite
completa, quita de lo escrito, además del bloque `origen` y la sección «Medidas y reutilizadas» de H30, las claves
`activaciones` de `informe.json` y los párrafos «Skills activadas:» de `informe.md`, y lo compara con
`informeJSONDeLaSuiteCompleta` e `informeMDDeLaSuiteCompleta` sin tocarlas; sus casos de «lo que se quita» ganan los
que fijan que se quita eso y nada más.

**Por qué.** SC-002 pide 0 diferencias. Quitar lo nuevo y comparar con las constantes de hoy deja todo lo demás bajo
control byte a byte, como hizo H30 con su bloque (el propio comentario del test, `:8361-8375`).

**Alternativa rechazada.** *Reescribir las dos constantes con la lista*: el test dejaría de fijar que lo de antes no
cambia.

### D7. La opción de instalación de FR-010: `INSTALACION=producto`

**Decisión.** `make evals-laboratorio … INSTALACION=producto`. Vacía o ausente es «sin la opción»; su único valor es
`producto`. Cualquier otro valor, o `producto` con variantes del laboratorio de H39, es un error de argumentos
(contracts/arnes.md §2).

**Por qué.** Nombra lo que cambia, la instalación de cada sesión, con el nombre que el spec da a las del producto. Un solo
valor porque el spec pide una sola instalación nueva («Fuera de alcance»: ninguna otra combinación).

**Alternativas rechazadas.** `SKILLS=…`: se lee como el `SKILL` de `make evals-sondeo`, que es una skill. `TODAS=1` o
`PRODUCTO=si`: un booleano disfrazado, sin decir qué es lo que pasa a ser «todo».

### D8. El argumento de FR-013: `EVAL=<nn>`

**Decisión.** `make evals-laboratorio … EVAL=04`: el número de dos cifras de una de las cuatro evals del plan de la
familia (`evalsDelSondeo`, `laboratorio_familias.go:100-150`), la que empieza por `<nn>-`. Vacío o ausente, el plan de
hoy. Con él, `SESIONES` es un entero mayor que 0. Un número que no es de ninguna eval del plan de la familia pedida, o
`EVAL` con variantes de H39, es un error de argumentos.

**Por qué.** El precedente del proyecto es `EVALS` de `make evals-sondeo`, que pide evals por su número de dos cifras
(`sondeo.go:888-901`), y FR-030 dice «la eval 04». En singular, porque pedir más de una eval queda fuera de alcance.

**Alternativas rechazadas.** *El nombre del fichero (`04-lgt-prescripcion.yaml`)*: más largo y fuera del precedente.
*Reutilizar `EVALS` con una lista*: prometería varias evals, que el spec deja fuera.

### D9. Dónde van los dos argumentos en la cadena del sondeo

**Decisión.** Detrás de los siete de hoy, en el orden `INSTALACION` y `EVAL`, en la receta (`Makefile:156`), en
`scripts/evals-laboratorio.sh` (que pasa a exigir nueve y los da a `go test` como `-instalacion` y `-eval`) y en
`ArgumentosDelLaboratorio`. `scripts/evals-sondeo-llavero.sh` ya pasa `"$@"` tal cual (su línea 38) y solo cambia su
comentario. Sus errores van detrás del de `DIRECTORIO`, cada uno con el nombre del argumento de `make`.

**Por qué.** Los siete de hoy no cambian de sitio, y la salida de error de cualquier petición de hoy no cambia: con los
dos vacíos no dan error (FR-011, FR-014). `TestRecetasDelSondeo` (`laboratorio_sondeo_test.go:1612-1688`) es el control
de la receta: su caso de `evals-laboratorio` gana los dos argumentos. `TestGuionDelLaboratorio`
(`laboratorio_sondeo_test.go:1434-1600`) es el del guion: hoy le pasa exactamente siete argumentos y exige que `go`
reciba `ordenDelLaboratorio`, con `-temporal` el último; sus casos ganan los dos argumentos, `ordenDelLaboratorio`
espera `-instalacion` y `-eval` delante de `-temporal`, y su caso de uso pasa un argumento menos de los nueve.

### D10. La instalación de las skills del producto en cada sesión

**Decisión.** Con `INSTALACION=producto`, el arnés deja en `skills/<variante>/` del temporal la copia de la variante como
hoy (`copiarLasSkillsDeLaVariante`, `laboratorio_sondeo.go:727-745`) y, a su lado, la de cada directorio de `skills/` del
árbol que no es la skill de la familia, con `os.CopyFS`, como hoy la de `<skill>-producto` (`copiarLasSkills`,
`:749-755`). Cada sesión enlaza cada entrada de ese directorio (`enlazarLasSkills`, `sesiones.go:506-527`), así que ve las
cuatro: la variante en el lugar de la de su familia y las otras tres de `skills/`, byte a byte. En `<skill>-producto` la
copia es `skills/` entera.

**Por qué.** FR-010: «como las deja `make install`», que el spec («Assumptions») lee como las cuatro de `skills/`, cada una
como está en el árbol; `make install` es `skills install -g --host claude` con el binario del árbol (`Makefile:72-75`), que
empotra `SKILL.md` y `references/*` de cada skill byte a byte (`skills.go:1-25`), y el job instala así
(`.github/workflows/evals.yml:392-398`). Los directorios de las variantes versionadas solo tienen la skill de su familia,
que es lo que fijan sus huellas (`TestVariantesDeControl`), así que ningún nombre se repite. La copia se hace una vez, en
`preparar` (`laboratorio_sondeo.go:686-718`), antes de abrir ninguna sesión, así que todas las sesiones de una invocación
ven la misma instalación, la que anota `skills_instaladas` (D11).

**Alternativas rechazadas.** *Ejecutar `skills install` con el `kitlegal` que `ConstruirLosProgramas` deja en el
temporal, y retirar después la skill de la familia para poner la variante en su lugar*: `skills install` escribe en su
ámbito —`./.agents/skills/` o, con `-g`, el del `HOME`, que en el sondeo es `home/`, uno para todas sus sesiones
(`preparar`, `laboratorio_sondeo.go:680-697`)— las skills, su manifiesto `kitlegal.json` y una entrada por host, enlace o
copia (`internal/core/instalacion/manifiesto.go:60-75`), y no en `skills/<variante>/`, el directorio que
`enlazarLasSkills` enlaza en cada sesión. Las variantes de una misma invocación, como `normativa-control,normativa-producto`
en FR-033 de H41, compartirían un ámbito con una sola skill de la familia; el manifiesto declararía la skill que la sustitución ya ha quitado; y `skills_instaladas`
tendría que leerse de otro sitio. Lo que instalaría de cada skill son los mismos bytes que la copia —el binario empotra
`SKILL.md` y `references/*` de `skills/` tal cual (`skills.go:1-25`)—, así que FR-010 no gana nada con ello. *Enlazar
`skills/<otra>` del árbol en `skills/<variante>/` sin copiarlo*: cada sesión leería el árbol mientras dura la invocación,
y un cambio en `skills/` durante ella cambiaría lo que ven las sesiones siguientes, contra la instalación única para todas
las sesiones de un sondeo que anota `skills_instaladas` (D11) y contra la copia que H41 ya hace de `<skill>-producto` (su
research D14).

### D11. La constancia de FR-012: `skills_instaladas` en `sondeo.json`

**Decisión.** Todo sondeo de una familia del producto escribe en su `sondeo.json`, detrás de `referencias_del_producto`
si la lleva, `skills_instaladas`: los nombres de las entradas de `skills/<variante>/` del temporal, en orden de nombre,
que son las de cada sesión del sondeo (data-model §3). Sin la opción, `["<skill>"]`; con ella, las cuatro. Los sondeos de
H39 no la escriben, y su `sondeo.json` sale como hoy.

**Por qué.** `sondeo.json` es el único fichero con el que el arnés describe un sondeo y la tabla lo lee
(`laboratorio_tabla.go:59-70`); una invocación es una instalación para todas sus sesiones, así que basta una lista por
sondeo. Solo los nombres: ni la copia ni una huella (FR-012). FR-012 la pide «con la opción y sin ella»: así la tabla da
la instalación de las dos filas sin la opción del sondeo de FR-030. Cambia un valor esperado de hoy: el `sondeo.json`
byte a byte que exige `exigirElSondeoDelArnes` (`laboratorio_sondeo_test.go:603-611`) gana la clave en las familias del
producto; lo que exige de cada sesión (`exigirLaSesionDeLaFamilia` y los de H39 y H40) y las sesiones que abre no
cambian (FR-015).

**Alternativas rechazadas.** *Un fichero aparte en `DIRECTORIO`*: un segundo fichero de descripción del sondeo que la tabla
tendría que buscar. *Escribirla solo con la opción*: FR-012 la pide sin ella, y la tabla no distinguiría una fila de un
sondeo de antes de una de este hito.

### D12. La tabla con la constancia y sin ella

**Decisión.** Si alguno de los sondeos de la tabla de una familia del producto tiene `skills_instaladas`, la tabla lleva
las columnas nuevas (contracts/tabla.md §3) y la fila se define también por su instalación; la de un sondeo sin la
constancia da la skill de la familia sola. Si ninguno la tiene, la tabla y el detalle son los de hoy, byte a byte. La de
H39 no cambia.

**Por qué.** FR-020 (la fila por instalación y sus columnas), FR-022 (sin constancia, igual que hoy; mezclados, la de la
familia sola) y FR-023.

### D13. Las columnas nuevas y su forma

**Decisión.** Tras «Claude Code», «Skills instaladas» (los nombres, unidos por «, »); tras «Activan antes», «Activan otra
skill», «Otras skills» y «No activan ninguna». «Otras skills» da cada otra skill, en orden de nombre, con
`<skill>: <k> de <n> · <p> % [<lo>–<hi> %]`, unidas por «; », y «ninguna» si no hay; los otros tres recuentos, como
«Activan antes». En el detalle, tras «Activa antes», «Activaciones», con el texto de D4. Con `n` = 0, vacías.

**Por qué.** FR-020 pide los nombres de las otras skills «cada uno con las sesiones que lo activan» y «cada recuento como
k de n con su porcentaje y su intervalo»: la cifra de cada nombre es un recuento, y va con su intervalo. La instalación
va con las demás celdas que definen la fila, para que se lea también en una fila con `n` = 0 (FR-021). Una sola celda
deja la misma cabecera en todas las tablas de una familia, con cualquier skill que activen las sesiones.

**Alternativas rechazadas.** *Una columna por cada otra skill, como las de las referencias*: las de las referencias son
las de la familia, fijas para cada familia (contracts/tabla.md §2 y data-model §8 de H41), y no dependen de lo que hagan
las sesiones; una columna por otra skill dependería de qué skills activan las sesiones de la tabla, así que la cabecera
cambiaría de una tabla a otra de la misma familia, y al añadir un sondeo al directorio, y dos tablas ya no se leerían
columna a columna. Las filas sin la opción, que solo tienen instalada la skill de su familia, llevarían además una celda
por cada skill que no tenían.

### D14. Fixtures y tamaño de las pruebas

**Decisión.** Las sesiones de FR-006 y las transcripciones de FR-024 son ficheros fijos en `testdata/`, escritos por
tareas `[datos]` sin red (plan, «Gates»): las del informe con la forma de los casos de `testdata/sesiones/informe/` y
`testdata/sesiones/leer-llamadas/`; las de la tabla, con la de `testdata/laboratorio/familias/`. Los intervalos de 40 de
FR-024 salen de un sondeo que el test monta con 40 copias de dos sesiones fijas, como `probarCuarentaSesiones` de H39
(`laboratorio_tabla_test.go:100-119`). El arnés se prueba con `SESIONES` de 1 a 4, como H41 (su contracts/arnes.md §5).

**Por qué.** Ninguna sesión fija de 40 copias en el repositorio, y cada sesión del arnés se prepara como en el job.

### D15. La aceptación e2e: un guion, el de la tabla

**Decisión.** Un guion congelado, `aceptacion/laboratorio-tabla.txtar` (historia 3), sobre `laboratorio tabla`, el
programa que lanza la persona, como H40 y H41. Las historias 1 y 2 no tienen una orden del binario que un guion pueda
lanzar: el informe lo escribe `EscribirInforme` dentro de `go test` (`scripts/evals.sh`) y el arnés abre sesiones con
credencial; su aceptación en `make ci` son sus tests (plan, «Aceptación e2e»). La historia 4 la hace la persona.

### D16. Lo que cambia en la documentación

**Decisión.** `CONTRIBUTING.md` (la lista de lo que lleva `informe.json` en «Job de evals»; la tabla de argumentos, con
`INSTALACION` y `EVAL` y lo que cambia de `SESIONES`, y las columnas de la tabla en «El sondeo del laboratorio»), el
comentario de `evals-laboratorio` en el `Makefile` («siete argumentos»), las cabeceras de `scripts/evals-laboratorio.sh` y
`scripts/evals-sondeo-llavero.sh`, y `CHANGELOG.md` (*Unreleased*).

**Por qué.** Lo que hoy documentan («siete argumentos», «`SESIONES` múltiplo de 4», las columnas de la tabla) deja de ser
cierto con FR-010, FR-013 y FR-020. Ningún ADR: el que lee el sondeo lo escribe la persona (FR-031).

### D17. El sondeo y el ADR quedan para la persona

**Decisión.** El run no lanza el sondeo de FR-030 ni escribe el ADR de FR-031; el plan lo deja en `gates/supuestos.md`
para que el informe final lo dé en «Supuestos y pendientes», como H41 (su `gates/informe-final.md:32`), y quickstart.md
§8 da las órdenes.

**Por qué.** FR-014, FR-031 y ADR 0032: abre sesiones con modelo y credencial.

## Comprobado en esta sesión

- **V1** `encoding/json/v2`: «By default, a nil slice is encoded as an empty JSON array, unless the FormatNilSliceAsNull
  option is specified», y «only a nil slice or map is omitted under "omitzero", while an empty slice or map is omitted
  under "omitempty"» (`go -C <repo> doc encoding/json/v2` y `…/v2.Marshal`, con la toolchain del módulo, go1.27.3).
- **V2** Bytes que añade la lista a un elemento de `evals` de `informe.json`, con `jsontext.WithIndent("  ")` como
  `informe.go:1142`: 0 sin lista, 26 con `[]`, 109 con una activación y 195 con dos. Medido con un programa en
  `$TMPDIR` con `/usr/local/go` y `GOEXPERIMENT=jsonv2` (la toolchain del módulo solo se resuelve dentro del repositorio);
  la API es la misma.
- **V3** Los extremos de Wilson de los ejemplos, con una copia de `recuento` e `intervaloDeWilson`
  (`laboratorio_tabla.go:1240-1259`) en `$TMPDIR`: 0 de 40 → [0,0–8,8 %]; 1 de 40 → [0,4–12,9 %]; 40 de 40 →
  [91,2–100,0 %]; 1 de 2 → [9,5–90,5 %]; 0 de 2 → [0,0–65,8 %]; 1 de 4 → [4,6–69,9 %]; 3 de 4 → [30,1–95,4 %]; 0 de 4 →
  [0,0–49,0 %]; 0 de 160 → [0,0–2,3 %]. Coinciden con los del spec.
- **V4** Transcripts de los tests sin `message.id`: `mensajeDeLlamadas` (`sesion_test.go:677-687`), el mensaje final de
  la ejecución sintética (`informe_test.go:6442`) y `mensajeAssistantDeTexto` (`limites_test.go:20`).
- **V5** Los 318 mensajes `assistant` de los `sesion.jsonl` de `internal/evals/testdata/` llevan `"id":"msg_…"`; ninguno
  sin él (`grep -rh --include=sesion.jsonl '"type":"assistant"' internal/evals/testdata`, con y sin `-v`, en esta sesión).
- **V6** `TestLeerSesion` y los de `sesion_test.go:234`, `:428` y `:1550` comparan la `Sesion` entera con
  `assert.Equal`.
- **V7** `exigirElSondeoDelArnes` compara `sondeo.json` byte a byte (`laboratorio_sondeo_test.go:603-611`) y lo usan
  `TestArnesDeLasFamilias`, `TestArnesDeNormativa` y `TestArnesDelLaboratorio`.
- **V8** `TestInformeDeLaSuiteCompletaNoCambia` fija `informe.json` sin `origen` e `informe.md` sin «Medidas y
  reutilizadas» contra dos constantes (`informe_test.go:8376-8392`, `:9202`, `:10164`), con doce secciones de sesión.
- **V9** `make install` (`Makefile:72-75`), lo que empotra el binario (`skills.go:24`) y la instalación del job
  (`.github/workflows/evals.yml:392-398`); cada sesión enlaza las entradas de su directorio de skills
  (`sesiones.go:483-527`).
- **V10** Los umbrales de las respuestas, con `sin_activar` de cada modo, van si la skill tiene juez o alguna eval
  declara sentencias, es de plazo o de cobertura o espera una redacción fechada (`umbrales.go:340-400`, comentario de
  `umbralesDelInforme`).
- **V11** `encoding/json` (v1), el de `sondeo.json` (`laboratorio_tabla.go:4`): «object keys which don't have a
  corresponding struct field are ignored» (`go doc encoding/json.Unmarshal`) y `omitempty` omite una lista vacía
  (`go doc encoding/json.Marshal`). Un `sondeo.json` de este hito lo lee sin error el guion de antes, y uno de antes no
  lleva la clave.
- **V12** Los lectores de `informe.json` que el repositorio versiona: `scripts/workflow/serie.sh:95-139` (con `jq`, campos
  nombrados de `.evals[]`) y `entradasDelInforme` (`conjunto_test.go:3191-3224`, con su propio tipo, sobre informes de
  H7.1 a H7.3). Ninguno cambia, y ninguno lee una clave que no nombra (FR-005).
- **V13** El flag `-evals` ya existe en `job_test.go:83` (el de `make evals-sondeo`); los nuevos del arnés, `-instalacion`
  y `-eval`, no chocan con ninguno de `job_test.go:37-83`.
- **V14** Los umbrales con `decide: true` que el job publica de cada skill del producto en los dos modos: los que
  `specs/035-h41-1-plazos-y-territorio/gates/informe-final.md` da como comprobados en el cierre de H41.1 (plan,
  «Controles de umbral»). `consulta-juridica` no publica ninguno con `decide: true` (research V11 de H41).
- **V15** Evals sin binario ni servidor: `evals/normativa/21-sin-binario-ni-servidor.yaml` y
  `evals/plazos-y-territorio/04-sin-binario-ni-servidor.yaml` (`grep -l 'sin_binario_ni_servidor: true'`).
- **V17** `os.CopyFS` «copies the file system fsys into the directory dir, creating dir if necessary» y «will not
  overwrite existing files»: devuelve un error con `fs.ErrExist` si un fichero ya existe (`go -C <repo> doc os.CopyFS`).
  La copia de cada otra skill va a un directorio que no existe, `skills/<variante>/<otra>/` (research D10); un nombre
  repetido no llega a sobrescribir nada.
- **V16** Tamaños de los ejemplos (`wc -c` en esta sesión): `sondeo.json` de `normativa-producto` con la skill sola, 290
  bytes, y con las cuatro, 349; sin variante del producto, 89; el párrafo «Skills activadas» de una y de dos
  activaciones, 35 y 65; una fila de la tabla de 40 con las cuatro skills, 570; una línea del detalle con dos
  activaciones, 183.

## Sin verificar

- **S1** Que Claude Code escriba `message.id` en todo mensaje `assistant` de su stream-json. No hay código de Claude Code
  en local; lo sostienen los transcripts grabados del repositorio (V5) y el contrato de H39 que lo exigía. Si no lo
  escribiera, la tabla contaría los turnos de ese transcript con el id vacío en lugar de dar el defecto de hoy (D1).
- **S2** Que GNU Make pase `EVAL` e `INSTALACION` como variables normales a la receta (`make --version`: GNU Make 3.81).
  Lo comprueba `TestRecetasDelSondeo` en `make ci` con los dos argumentos.
- **S3** La causa de la caída: es lo que mide el sondeo de la persona (FR-030), no este run.
