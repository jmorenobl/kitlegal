# Research: H40 · Por qué `normativa` no lee sus referencias en modo herramienta

Decisiones del plan (D), lo comprobado en local en esta sesión (V), con dónde, y lo que no se ha podido comprobar sin
red ni sesión con modelo, como supuesto no verificado (S). Nada de lo que sigue se ha medido con un modelo: las cifras de
partida (#214, H39, `brazo-D`) son las que citan el hito y el ADR 0050.

## Verificaciones (V)

| Id | Qué | Dónde se comprobó |
|---|---|---|
| V1 | El cuerpo de `skills/normativa/SKILL.md` —desde la línea que sigue a la que cierra el frontmatter, la 17— tiene 10 929 caracteres; la skill tiene ocho referencias y su cuerpo ocho remisiones `](references/<f>.md)`, una a cada una | `tail -n +17 skills/normativa/SKILL.md \| wc -m`; programa desechable de medida en `$TMPDIR/h40-medida/` (fuera del repositorio) |
| V2 | Los cinco «antes» de `remisiones-en-orden` (contracts/variantes.md §3) están una sola vez en el `SKILL.md` de hoy; cada «después», con la regla del salto de línea de FR-003, da carácter a carácter el texto del hito; aplicados los cinco, el cuerpo tiene 11 084 caracteres (+155), las ocho remisiones, y ninguna línea de más de 120 caracteres que no tuviera ya `control` | el mismo programa de medida |
| V3 | La regla de `cuerpo-corto` de contracts/variantes.md §4 da un cuerpo de 3298 caracteres, con las ocho remisiones; la inserción de `testigo-en-referencia` tiene su ancla `[con_herramientas](references/con_herramientas.md).` una sola vez y deja el cuerpo en 10 973 (+44); la sección que añade a `con_herramientas.md` tiene 197 caracteres | el mismo programa de medida |
| V4 | `references/cierre.md` de la skill de laboratorio de H39 es `# Cómo termina la respuesta`, una línea vacía y el párrafo de la regla en una sola línea que termina en `Sello de lectura: KL-7Q4`, con su salto; es igual en las tres variantes de referencia. `con_herramientas.md` termina con un salto de línea, y su última sección es `## Encadenar y fallar` | `cat -et laboratorio/variantes/referencia-paso/skills/articulos-del-boe/references/cierre.md`; `cmp` con la de `referencia-secundaria`; `grep -n '^#'` y `od -c` de `skills/normativa/references/con_herramientas.md` |
| V5 | `TestLaboratorioFueraDelProducto` recorre **todo** `laboratorio/` y exige que ningún fichero empotrado o instalado tenga los bytes de uno suyo: una copia byte a byte de `skills/normativa` dentro de `laboratorio/` lo pondría en rojo. Los nombres de skill que vigila salen solo del patrón `laboratorio/variantes/*/skills/*` | `internal/evals/laboratorio_test.go:893-1039` |
| V6 | El binario empotra `skills/*/SKILL.md` y `skills/*/references/*` y nada más | `skills.go:24` (`//go:embed skills/*/SKILL.md skills/*/references/*`) |
| V7 | `laboratorio/` no es de lo común ni de lo que mide: ninguna ruta suya elige una skill. `internal/evals/`, `cmd/` y `Makefile` sí son de lo común: un cambio suyo elige las cuatro skills del producto en el cierre | `internal/evals/dependencias.go:30-56`, `:234-250` |
| V8 | Cada sesión se prepara con las consultas de **todas** las evals del directorio `Evals` del reparto, y su eval se elige por `Fichero`; es lo que hacen el job y `make evals-sondeo` con `evals/<skill>` | `internal/evals/sesiones.go:439-446`; `internal/evals/preparar.go:210-290` |
| V9 | El arnés de H39 copia `<variante>/skills/` con `os.CopyFS` al directorio de skills del reparto, y `servidorDelPar` devuelve la orden de hoy (vacía) para toda variante sin `instrucciones` ni `parametro` en `cambia` | `internal/evals/laboratorio_sondeo.go:519-614` |
| V10 | `os.CopyFS` copia un enlace simbólico como enlace (`case ModeSymlink: … Symlink(target, newPath)`) | `src/os/dir.go:160-165` del toolchain del módulo, go1.27.3 (`go env GOROOT`) |
| V11 | El arnés de H39 valida `VARIANTES` contra una sola declaración (`varianteDeclarada`), con el mensaje `«<v>» no es ninguna variante del laboratorio`, y `SESIONES` contra `evalsDelLaboratorio = 4` | `internal/evals/laboratorio_sondeo.go:36`, `:283-409` |
| V12 | La tabla de H39 decide la variante por el nombre del directorio, falla con `… no es ninguna variante de la declaración del laboratorio` si no es de la declaración, cuenta una sola referencia (`declarada.Referencia`) y escribe la regla testigo en toda fila | `internal/evals/laboratorio_tabla.go:287-353`, `:548-576`, `:758-795` |
| V13 | `make laboratorio-tabla DIRECTORIO=internal/evals/testdata/laboratorio/sondeo` da hoy, byte a byte, `internal/evals/testdata/laboratorio/tabla.md`, y esos ficheros son los de `main` | `make -s … \| cmp - internal/evals/testdata/laboratorio/tabla.md` (sin salida, código 0); `git diff --quiet main -- internal/evals/testdata/laboratorio laboratorio` |
| V14 | Las sesiones fijas de H39 son 21 (63 ficheros, unos 45 KB de contenido) en dos sondeos, `h39-a` y `h39-b`, con `sondeo.json` `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}` | `git archive HEAD internal/evals/testdata/laboratorio/sondeo \| wc -c` (112 640 con las cabeceras de tar); `cat` de `sondeo.json` |
| V15 | La suite congelada son los ficheros de `<feature_dir>/aceptacion/`; `activar` copia cada `*.txtar` a `internal/app/testdata/script/<hito en minúsculas>-<nombre>.txtar`, que ejecuta `TestEntregaDelHito`; «rojo primero» da por ilegible un guion cuya salida tiene `unknown command`, `cannot parse`, `unexpected command` o `usage: ` | `scripts/workflow/aceptacion.sh` |
| V16 | El arnés e2e construye solo el binario de e2e de `kitlegal` (siete construcciones, todas con el nombre `kitlegal`), deja en el `PATH` la de desarrollo y pasa a cada guion variables `KITLEGAL_*_BIN`; los guiones exigen `exec` explícito. `TestBinariosDelArnes` recorre una lista explícita de las siete, y `TestVariablesDeLosGuiones` fija que las rutas que nombra son absolutas | `internal/app/e2e_test.go:45-52`, `:144-158`, `:195-241`, `:389-411`, `:552-581`, `:1744-1764`, `:1893-1914` |
| V17 | `env NOMBRE` de testscript escribe `NOMBRE=valor` en el registro (`ts.Logf`), no en la salida estándar: no sirve de precondición con `stdout` | `~/go/pkg/mod/github.com/rogpeppe/go-internal@v1.16.0/testscript/cmd.go:205-229` (versión de `go.mod`) |
| V18 | El informe versionado de H32 tiene la sesión `05-trlrhl-impuestos-municipales-claude-sonnet-5-5-01` (modo orden), de la eval 05, activada y con una respuesta de 1222 caracteres que termina en «El texto consolidado es informativo y no sustituye el asesoramiento de un profesional.», sin la línea del sello | `jq` sobre `specs/030-h32-arquitectura-de-skills/gates/evals/normativa.json` |
| V19 | Las evals 04, 05, 15 y 22 de `normativa` son `04-lgt-prescripcion.yaml`, `05-trlrhl-impuestos-municipales.yaml`, `15-irpf-rendimientos-por-materia.yaml` (informativa, con cinco `lecturas`) y `22-lcsp-contrato-menor-en-2019.yaml` (con `redaccion_fechada`) | `ls evals/normativa/`; `cat` de las cuatro |
| V20 | En los informes del cierre de H39, los umbrales con `decide: true` son, en `normativa`, `sin_activar` y `redaccion_sin_fechar` de `claude-sonnet-5-5` en los dos modos y `afirma_lo_no_leido:claude-sonnet-5-5` (más la medida del juez y las duraciones); en `jurisprudencia`, `sin_activar`, `cita_sin_documento` y `documento_sin_cotejar` en los dos modos y `afirma_lo_no_leido`; en `plazos-y-territorio`, `sin_activar`, `plazo_sin_fecha`, `cobertura_sin_declarar`, `vigencia_sin_declarar`, `redaccion_sin_fechar` y `lectura_sin_intentar` en los dos modos; `consulta-juridica` no tiene ninguno (solo `fuera_de_lo_grabado`, con `decide: false`), y su veredicto sale de sus series (ADR 0016) | `jq '[.umbrales[] \| select(.decide==true)]'` sobre `specs/031-h39-c-omo-se/gates/evals/*.json` |
| V21 | Wilson al 95 % (z = 1,96), con la fórmula de `intervaloDeWilson`: 0/40 → 0,0–8,8; 20/40 → 35,2–64,8; 40/40 → 91,2–100,0; 9/40 → 12,3–37,5; 3/40 → 2,6–19,9; 1/1 → 20,7–100,0; 0/1 → 0,0–79,3; 1/2 → 9,5–90,5; 0/2 → 0,0–65,8; 2/2 → 34,2–100,0 | programa desechable en `$TMPDIR/h40-wilson/` con la fórmula de `internal/evals/laboratorio_tabla.go:833-842` |
| V22 | `TestArnesDelLaboratorio` pasa en este sandbox, sin red ni modelo, en 4,8 s; su caso válido de más sesiones, en 3,1 s | `go test -count=1 -run '^TestArnesDelLaboratorio$' -v ./internal/evals/` |
| V23 | El sondeo de H39 de la persona juntó dos lanzamientos con `make laboratorio-tabla`, porque el arnés combina todas las variantes pedidas con todos los modos pedidos | `evidencias/sondeos-20261103/LEEME.md` |
| V24 | `.goreleaser.yaml` construye solo `./cmd/kitlegal`; `cmd/` está exento de R4 y R5 en `.golangci.yml` | `.goreleaser.yaml:30-31`; `.golangci.yml:466-473` |
| V25 | El tope del cuerpo, `maximoDeCaracteresDelCuerpo = 11000`, lo aplica `internal/skills` a `skills/`, y lo comprueba `TestSkillsDelRepositorio` | `internal/skills/sincronia.go:31`; `internal/app/skills_test.go` |
| V26 | `make ci` escribe `coverage.out` y `coverage-integration.out` en la raíz, que `.gitignore` excluye (`/coverage.*`); `make vuln` es `govulncheck ./...`, cuya base por defecto es `https://vuln.go.dev` | `Makefile:42`, `:87`, `:91`, `:118-119`, `:294`; `.gitignore:6`; `go tool -modfile=tools/govulncheck/go.mod govulncheck -h` |

## Decisiones (D)

### D1. Dónde viven las variantes: `laboratorio/normativa/<variante>/skills/normativa/`

**Decisión.** Cada variante es un directorio de `laboratorio/normativa/` con una sola skill, `skills/normativa/`, completa:
`SKILL.md` y sus ocho referencias (contracts/variantes.md §1). Es la forma que el arnés de H39 ya copia como directorio
de skills del reparto (V9), así que cada sesión tiene solo la variante de su fila, con el nombre `normativa` (FR-008).

**Por qué.** Fuera de `skills/` no se empotra ni se instala (V6); `laboratorio/` no elige ninguna skill (V7); y el
patrón con el que `TestLaboratorioFueraDelProducto` saca los nombres de skill de H39 no la alcanza (V5), de modo que
`normativa` no pasa a ser un nombre prohibido.

**Alternativas.** (a) En `laboratorio/variantes/`, junto a las de H39: entrarían en `variantes.yaml`, cuyas claves
(`skill`, `entrada`, `referencia`…) son de la skill de H39, y cambiarían la declaración que fijan los tests de H39
(FR-023; «Fuera de alcance»: ampliar el laboratorio de H39). (b) Un directorio nuevo de primer nivel: otro sitio fuera
del producto que vigilar, sin ganar nada. (c) `internal/evals/testdata/`: es de lo común (V7), y cada cambio de una
variante mediría las cuatro skills del producto.

### D2. Nombres: `normativa-control`, `normativa-remisiones-en-orden`, `normativa-cuerpo-corto`, `normativa-testigo-en-referencia`

**Decisión.** El nombre de la variante lleva delante `normativa-` y es el de su directorio; es lo que la persona escribe
en `VARIANTES` y lo que sale en la tabla (FR-010: «cómo se nombran las variantes para no confundir las dos `control` lo
decide el plan»).

**Por qué.** Ningún nombre coincide con uno de H39 (V11), así que el arnés sabe la familia de cada nombre sin otro
argumento, y la tabla, la de cada fila, sin leer nada más que el directorio.

**Alternativas.** (a) Un argumento nuevo de `make` (`FAMILIA=`): cambia la receta, el guion del llavero y
`TestRecetasDelSondeo`, sin que el spec lo pida. (b) Renombrar solo la de control (`normativa`): asimétrico, y la tabla
seguiría teniendo filas `cuerpo-corto` sin decir de qué skill son. (c) `normativa/control`, con barra: el sondeo
tendría un nivel más de directorios y la tabla tendría que leerlo aparte.

### D3. La familia de `normativa`, declarada en Go y no en un YAML

**Decisión.** `internal/evals/laboratorio_normativa.go` declara la familia: la skill (`normativa`), el directorio de sus
variantes, sus cuatro variantes con sus modos, las cuatro evals del sondeo en `evals/normativa`, las ocho referencias
que cuenta la tabla y la variante con regla testigo (data-model §1). La de H39 sigue siendo la de
`laboratorio/variantes.yaml`, sin cambios.

**Por qué.** Lo fija entero el hito: no hay nada que una persona cambie sin cambiar el hito, y una declaración en Go no
necesita esquema, lector estricto ni tests del lector. La tabla no tiene que leer ningún fichero para la familia, así
que el guion de aceptación no lleva una copia de una declaración que no es de `main` (D13).

**Alternativas.** `laboratorio/normativa/variantes.yaml`, con su esquema y su lector, como la de H39: el mismo dato con
tres mecanismos más (esquema, lector, test del lector) y otro fichero que el guion congelado tendría que copiar.

### D4. Las variantes son copias completas, escritas por las tareas

**Decisión.** Las cuatro variantes son ficheros normales del repositorio (36: nueve por variante), no enlaces ni algo que
se genere al lanzar el sondeo. `make ci` las compara con `skills/normativa` y con su declaración (D5).

**Por qué.** FR-002 pide una copia que `make ci` compara con `skills/normativa` y que falla, nombrando el fichero, cuando
la skill cambia y la copia no; una variante generada al vuelo no podría desviarse y el control no mediría nada. Un enlace
tampoco: `os.CopyFS` lo copiaría como enlace (V10) y, si es relativo, dejaría de apuntar a su destino en el temporal.

### D5. Cómo comprueba `make ci` cada declaración: la variante esperada se calcula desde `control`

**Decisión.** `TestVariantesDeNormativa` calcula, desde los ficheros de `normativa-control`, cada variante esperada y la
compara byte a byte con la del repositorio, fichero a fichero (contracts/variantes.md §2 a §5):

- **`remisiones-en-orden`**: cinco sustituciones `{antes, después, hito}`; cada `antes` tiene que estar una sola vez en
  el `SKILL.md` de control (si no, la diferencia ya no se aplica), `después` normalizado con la regla de FR-003 —cada
  salto de línea con la sangría de la línea que sigue cuenta como un espacio, y sin la marca `- ` de la viñeta— tiene
  que ser el `hito` carácter a carácter, y la esperada es control con las cinco sustituciones.
- **`cuerpo-corto`**: la regla por secciones de contracts/variantes.md §4, con el encabezado renombrado; cada encabezado
  y la viñeta que se quedan tienen que estar una vez en control.
- **`testigo-en-referencia`**: la inserción en el párrafo del «Protocolo» y, al final de `con_herramientas.md`, la
  sección con el párrafo de `references/cierre.md` de H39, leído de ese fichero.
- En las tres, el cuerpo conserva las ocho remisiones, una a cada referencia, y lo demás es el de control.

**Por qué.** Es la forma más directa de FR-006 —«falla si algo fuera de lo declarado difiere en un byte, si un texto
declarado no es exactamente el que fija el hito, o si la diferencia declarada ya no se puede aplicar»—: una sola
comparación byte a byte por fichero, y la declaración que no se aplica falla antes, con su motivo.

**Alternativas.** (a) Comprobar las reglas sin calcular la esperada —líneas de control en su orden, presentes las que se
quedan, ausentes las que no—: deja libres las líneas vacías, que no declara ninguna variante. (b) Normalizar los saltos
de línea de todo el cuerpo: dejaría pasar un párrafo no declarado partido de otra forma, que es una diferencia de bytes
fuera de lo declarado.

### D6. Dónde se parten las líneas de lo sustituido

**Decisión.** El «después» de cada sustitución se escribe con las líneas partidas por espacios, sin ninguna línea nueva
de más de 120 caracteres y sin tocar ningún byte que no sea de lo sustituido; los bytes exactos están en
contracts/variantes.md §3 (V2: ninguna línea de más de 120 que `control` no tuviera ya).

**Por qué.** FR-003 admite partir lo sustituido («un salto de línea con la sangría de la línea que sigue cuenta como un
espacio»); fijar los bytes en el contrato deja una única variante esperada.

### D7. Las evals del sondeo: `evals/normativa`, con su preparación, y el plan con las cuatro

**Decisión.** En la familia de `normativa`, el directorio `Evals` del reparto es `evals/normativa`, el mismo que usan el
job y `make evals-sondeo`, y el plan abre `SESIONES/4` sesiones de cada una de las evals 04, 05, 15 y 22, en ese orden
(contracts/arnes.md §3). Cada sesión se prepara, como en el job, con las consultas de todas las evals de esa carpeta
(V8), servidas por lo grabado y sin red (ADR 0043).

**Por qué.** Son «las evals reales de `normativa` con sus lecturas grabadas» (FR-010): con la misma preparación que el
job, `control` se mide en las condiciones del producto salvo por las skills instaladas (FR-008).

**Alternativas.** Copiar las cuatro a `laboratorio/normativa/evals/`, como las de H39: duplica evals del producto, con
sus `lecturas` y su `redaccion_fechada`, que se desviarían cuando cambien.

### D8. El arnés y las dos familias

**Decisión.** `comprobar` busca cada nombre de `VARIANTES` en las dos familias: uno que no es de ninguna da el mensaje
de hoy (`«<v>» no es ninguna variante del laboratorio`); con nombres de las dos, una línea
`VARIANTES: se mezclan variantes del laboratorio de H39 y variantes de normativa`, detrás de las de los nombres. Los
pares, las comprobaciones de modelo, versión, sesiones, concurrencia y directorio, la preparación, la salida y el
orden de los errores son los de H39 (contracts/arnes.md §2); lo que cambia con la familia es de dónde se copian las
skills, qué evals se planifican y desde qué carpeta se preparan.

**Por qué.** FR-010 y la aclaración de la sesión de clarify: una invocación es de una sola familia, y mezclarlas es un
error de argumentos que no abre nada. El mensaje de una variante que no existe no cambia: sigue siendo cierto (las de
`normativa` también viven en `laboratorio/`), y `TestArnesDelLaboratorio` lo fija.

**Las cinco filas de la aceptación son dos lanzamientos**, como en el sondeo de H39 (V23): `normativa-control` en los
dos modos, y las otras tres en modo herramienta; `make laboratorio-tabla` los junta leyendo el directorio padre
(contracts/arnes.md §1). Pedir las cuatro con `MODOS=orden,herramienta` es pedir tres en un modo que no les corresponde,
y FR-010 lo hace error.

**Alternativas.** Que `MODOS` se aplique solo a las variantes que tienen ese modo: contradice FR-010 («una variante en un
modo que no le corresponde … es un error de argumentos»).

### D9. La tabla y las dos familias

**Decisión.** La familia de una tabla es la de la primera variante que lee —sondeos y variantes en orden de nombre—: si
su nombre es de `normativa`, la de `normativa`; si no, la de H39, con su comportamiento de hoy. En una tabla de
`normativa`, un directorio que no es una variante de `normativa` es un defecto: `<ruta> no es ninguna variante de
normativa`, código 1 y sin tabla (contracts/tabla.md §1).

**Por qué.** FR-010: «su tabla tiene las columnas de esa familia»; y FR-023: con las sesiones de H39, nada cambia.
Juntar en una ruta sondeos de las dos familias es una lectura que el guion no sabe hacer, y la regla genérica la cubre
(defecto `inesperado`, 1).

### D10. La tabla de `normativa`: una celda por referencia, con su turno

**Decisión.** Una fila por variante, modo, modelo y versión, con 16 columnas: las seis de H39 (variante, modo, modelo,
Claude Code, sin terminar, n), «Activan antes», una por cada referencia —en orden de nombre de fichero— con
`<k> de <n> · <p> % [<lo>–<hi> %]` y, si alguna sesión la abre, ` · turno <mediana> (<mín>–<máx>)`, y «Cumplen la regla»,
solo en `normativa-testigo-en-referencia` (contracts/tabla.md §3).

**Por qué.** FR-020 pide, por referencia, el recuento y el turno de la primera apertura; en la misma celda, cada turno va
con su referencia. El formato de cada cifra es el de H39 (recuento, Wilson y mediana con «,5»), con el mismo código.

**Alternativas.** Dos columnas por referencia, como las dos de la referencia de H39: 24 columnas, y el turno lejos de su
recuento. Las referencias en el orden de las remisiones del cuerpo: depende del cuerpo, que la reorganización cambiará.

### D11. La regla testigo de `normativa`: la de H39, con respuestas fijas a las cuatro evals

**Decisión.** El guion es `reglaTestigo`, sin cambios (FR-030). `TestReglaTestigoDeNormativa` lo comprueba con seis
respuestas fijas a las evals del sondeo en `internal/evals/testdata/laboratorio/normativa/respuestas/`, con el veredicto
en el nombre como en H39 —la que cumple; sin la línea; la línea en negrita, entre comillas y en código; la línea antes
del final—, y con una respuesta real de la skill del producto a la eval 05, leída del informe versionado de H32 (V18),
que no cumple (contracts/tabla.md §6).

**Por qué.** Son las formas que enumeran FR-030 y la historia 3, escenario 3. La respuesta de la skill del producto es
la de `control` en el sentido del spec: lo que responde `normativa` sin la regla.

**Alternativas.** Añadirlas a `testdata/laboratorio/respuestas/`: `TestReglaTestigo` exige que ese directorio tenga
exactamente las seis de H39, y habría que cambiar lo que fija.

### D12. FR-001 con el test de H39, generalizado

**Decisión.** `TestLaboratorioFueraDelProducto` sigue siendo el control de «fuera del producto» para todo `laboratorio/`
(contracts/variantes.md §6): (a) los ficheros de una variante que son copia de uno de `skills/` —la misma ruta dentro
de su skill y los mismos bytes— no entran en la comprobación de bytes, porque nada los distingue de su original; (b) los
nombres de las variantes de `normativa` (los directorios de `laboratorio/normativa/`) se suman a los nombres que no
pueden ser un tramo de ninguna ruta empotrada o instalada ni una skill del job ni una carpeta de `evals/`; (c) lo
demás, como hoy. Con los ficheros de H39, que no son copia de ninguno de `skills/`, el test hace lo mismo que hoy.

**Por qué.** Sin (a) el test se pone en rojo con `normativa-control` (V5), que es idéntica a `skills/normativa` por
FR-002. Con (a) y (b), lo que puede entrar en el binario o la instalación sin ser del producto —una variante bajo
`skills/` con su nombre, o el `SKILL.md` o el `con_herramientas.md` de una variante que difiere— sale en rojo.

**Alternativas.** Un test aparte para las variantes de `normativa` y dejar el de H39 recorriendo todo `laboratorio/`:
el de H39 seguiría en rojo. Uno que recorra solo lo de H39 y otro solo lo de `normativa`: dos recorridos del mismo árbol.

### D13. La aceptación e2e: un guion de la tabla, con el programa `laboratorio` en el arnés e2e

**Decisión.** La suite congelada es `aceptacion/laboratorio-tabla.txtar` (contracts/arnes-e2e.md §2). Lleva copias de
`main` de las sesiones fijas de H39, de su `tabla.md` y de `laboratorio/variantes.yaml`, y unas sesiones fijas de las
variantes de `normativa` con su tabla esperada; ejecuta `laboratorio tabla` sobre las dos y compara su salida byte a
byte. Para que pueda ejecutarlo, el arnés e2e de `internal/app` construye también `./cmd/laboratorio` y da su ruta en
`KITLEGAL_LABORATORIO_BIN` (contracts/arnes-e2e.md §1); lo hace la tarea `[aceptacion]` con el guion, porque sin el
programa el guion no puede fallar hoy por una aserción.

**Por qué.** FR-023 y SC-007 piden una suite de aceptación congelada que pase al guion las sesiones de H39 de `main` y
falle si cambia una columna o una cifra, también si el hito cambia esas sesiones o esa tabla; la suite congelada solo se
ejecuta como guion de `TestEntregaDelHito` (V15). Hoy el guion falla por una aserción: con el programa construido, la
parte de H39 pasa y la de `normativa` sale con 1 (`… no es ninguna variante de la declaración del laboratorio`), y
`exec` lo da por fallo.

**Alternativas.** (a) «Aceptación e2e: no aplica»: contradice FR-023. (b) Una precondición en el guion, como en H21,
con el cambio del arnés en una tarea posterior: no hay comportamiento de `kitlegal` que la sostenga, y `env NOMBRE` no
escribe en la salida estándar (V17); un `exec` de un programa que no existe falla por no poder ejecutarse, no por una
aserción. (c) Una orden del arnés que llame a la tabla en proceso: no ejecuta el programa que lanza la persona y ata los
tests de `internal/app` a `internal/evals`.

### D14. El control del arnés: un sustituto de `claude`, como en H39

**Decisión.** `TestArnesDeNormativa` usa la preparación sin construir y el sustituto de `claude` de
`TestArnesDelLaboratorio`: con cada forma inválida de SC-006 (variante inexistente; las tres variantes que solo se
miden en herramienta, pedidas juntas en modo orden, con una línea por variante; `SESIONES=6`; una mezcla; la versión
distinta y `MODELO=sonnet`), 0 sesiones y su mensaje. Sus casos válidos son los dos lanzamientos de la aceptación con
menos sesiones: `VARIANTES=normativa-control MODOS=orden,herramienta SESIONES=4`, una sesión de cada eval en cada modo,
y las otras tres variantes con `MODOS=herramienta SESIONES=8`, dos de cada eval en cada variante. En cada sesión,
`claude/skills/` tiene solo `normativa`, cuyo `SKILL.md` es el de su variante; sin `servidor.json` en orden y con la
orden de hoy en herramienta (contracts/arnes.md §5).

**Por qué.** SC-006 y FR-008 prometen 0 sesiones con cada par variante × modo que no se mide y n/4 por eval en cada uno
de los que sí: entre los tres casos se piden los ocho pares de la familia, cinco válidos y tres no, así que el test
falla si la declaración de la familia da a una variante un modo que no es suyo o le quita el suyo, o si el camino de un
modo instala otra skill que la de su variante. Los casos válidos usan la carpeta real `evals/normativa`, así que también
fallan si una de las cuatro evals deja de estar.

**Alternativas.** Un caso por par inválido: tres casos con la misma comprobación que uno, porque los errores de los
pares se dan todos juntos (§2). Un solo caso válido con las cuatro variantes en los dos modos: el arnés lo rechaza, y
no es ninguna orden que lance la persona.

### D15. Documentación

La ayuda de `make evals-laboratorio` y `make laboratorio-tabla` en el `Makefile`, «El sondeo del laboratorio» de
`CONTRIBUTING.md` y `CHANGELOG.md` (*Unreleased*) dicen que el arnés lanza también las variantes de `normativa`, con sus
nombres, sus modos, sus evals y los dos lanzamientos. El comentario de paquete de `cmd/laboratorio` dice que su tabla
cuenta las dos familias. Ninguna receta cambia.

### D16. Cierre

El run no lanza el sondeo ni escribe el ADR (FR-042): el informe final lo dice, por `gates/supuestos.md`. Si la medición
del cierre da un rojo de partida en modo herramienta (FR-043), la reparación no toca `skills/` ni reorganiza una skill, y
el informe da sus cifras para que decida una persona.

## Supuestos no verificados (S)

- **S1.** `claude --version` sigue escribiendo `<versión> (Claude Code)` (S1 de H39). Si no, el arnés dice que la versión
  no coincide y no abre nada: falla del lado seguro.
- **S2.** En una sesión real, `normativa` se activa con un `tool_use` de `Skill` con `input.skill` `normativa`, y una
  referencia se abre con un `Read` cuyo `file_path` termina en `normativa/references/<fichero>` o con un `Bash` que la
  nombra. En el repositorio no hay transcripciones reales de `normativa` (las de los tests son sintéticas); la regla es
  la de H39, con la que el sondeo real de H39 contó aperturas de 39 y 40 de 40 (`evidencias/sondeos-20261103/tabla.md`).
- **S3.** No se ha medido cuánto tarda `TestArnesDeNormativa`, cuyos dos casos válidos abren 8 y 24 sesiones con el
  sustituto: cada sesión se prepara con las consultas de las 23 evals
  de `evals/normativa`, más que las cuatro del laboratorio de H39, cuyo caso válido tarda 3,1 s (V22).
- **S4.** Que el sondeo de la persona, con la credencial, abra sus 200 sesiones lo comprueba la persona (FR-040); el run
  no puede.

## Fuentes externas

Ninguna. El hito no graba nada (FR-013): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe; las
sesiones del sondeo se sirven de las grabaciones de `testdata/evals/` que ya usa el job, y nada se escribe en
`evidencias/` ni en `data/`.
