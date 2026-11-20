# Research: H33 · Cobertura en positivo por capacidad y materia

Decisiones del plan (D), comprobaciones hechas en esta sesión (V) y supuestos que no se han podido medir (S). Cada
decisión dice la alternativa rechazada y por qué (constitución, «Criterio de decisión autónoma», punto 3). Ningún
NEEDS CLARIFICATION queda abierto: el spec y el ADR 0058 fijan textos, datos, contrato del generador, evals y controles.

## Comprobaciones (V)

Todas ejecutadas en esta sesión, en la rama `039-h33-cobertura-en-positivo` (cabeza `45d24724`, contenido de `main`
`7a7526e8` más el fichero del hito movido a `done/`), salvo donde se dice otra cosa.

- **V1 · Los cuerpos.** `git show <rev>:skills/<skill>/SKILL.md | awk 'BEGIN{n=0} /^---$/ && n<2 {n++; next} n>=2' | wc -m`
  da, en `HEAD`, 3810 (`consulta-juridica`) y 10 882 (`plazos-y-territorio`), las cifras del spec; en `h33-prototipo`
  (`27413343`), 4481 y 10 686, las de FR-008 y FR-009. El `git diff HEAD h33-prototipo` de los dos `SKILL.md` y de
  `references/administrativo.md` son, línea a línea, los cambios de FR-005, FR-007 a FR-009 (contracts/skills.md los
  copia con sus cortes de línea): los bytes que entrega H33 en lo escrito a mano son esos.
- **V2 · El grep de testscript.** `github.com/rogpeppe/go-internal@v1.16.0/testscript/cmd.go:690` compila el patrón como
  `regexp.Compile("(?m)" + pattern)` sobre el fichero entero; `-count=N` exige `len(re.FindAllString(text, -1)) == N`
  (línea 716). Un patrón con `\n` casa a través de líneas, como en los guiones de H41.
- **V3 · El tope del cuerpo con una expresión regular.** `regexp/syntax` (Go 1.27.4, `$(go env GOROOT)/src/regexp/syntax/parse.go:436`
  y `:1003`) rechaza una repetición de más de 1000 y una repetición anidada cuyo producto pase de 1000
  (`(?:(?s:.){1000}){10}` da `invalid repeat count`, ejecutado). Diez `(?s:.){1000}` concatenados y un `(?s:.){N}` no
  anidan: un programa en `$TMPDIR` (`go run`) con `(?m)\A---\n(?s:.*?)\n---\n` + `(?s:.){1000}`×10 + `(?s:.){882}` casa
  con el `SKILL.md` de `plazos-y-territorio` de `HEAD` (10 882 caracteres) y con `{883}` no, en menos de 1 ms; el patrón
  mide 153 bytes. `.` con `(?s)` es una runa: cuenta caracteres como `wc -m`.
- **V4 · El esquema y sus mensajes.** `santhosh-tekuri/jsonschema/v6@v6.0.3`: `kind/kind.go:180` («value must be one of
  %s», cada valor entre comillas simples por `quote`, `:616`), `:318` («missing property %s»), `:364` («additional
  properties %s not allowed»), `:454` («minLength: got %d, want %d»); `validator.go:623-631` evalúa `if` y, según
  casa, `then` o `else` sobre la misma instancia, y añade su error sin cambiar la ubicación. `internal/skills/esquemas.go:406-411`
  presenta las hojas del error (las causas sin causas): un `{norma, precepto}` mal escrito da, con `else`, la misma hoja
  y la misma ruta que hoy con el `required` del objeto. `TestFormatoDePlazos` lo confirma en la implementación.
- **V5 · El paso `grabar_datos`.** `scripts/workflow/grabar-datos.sh:84` y `:103` buscan los manifiestos con
  `(^|/)testdata/grabaciones\.json$`, `:111` toma el paquete de su ruta, `:119` ejecuta
  `go test -tags=grabacion -run '^TestGrabar' ./<paquete>/` con `KITLEGAL_RECORD=1` y `KITLEGAL_EVIDENCIAS`, y `:132-133`
  revierte todo fichero que no esté en `<paquete>/testdata/` ni en `evidencias/<hito>/`. `testdata/evals/grabaciones.json`
  (el manifiesto de H5, donde el prototipo añadió `a59`) no termina en `testdata/grabaciones.json`: no lo detecta, y lo
  que se grabara en `testdata/evals/` se revertiría.
- **V6 · Qué graba `TestGrabarEvals`.** `internal/evals/grabacion_test.go:34-49`: graba los manifiestos de H5 y de H9 y,
  para cada consulta, sirve primero desde las grabaciones de H4 y de H5 y solo pide a la fuente lo que falta. La
  búsqueda, los metadatos y el índice del ET (`BOE-A-2015-11430`) están en `testdata/evals/boe.legislacion-consolidada/`
  (`ls`): una entrada del ET en el manifiesto de H9 solo graba `texto_bloque_a59`. `TestGrabacionesSinSolape`
  (`internal/evals/grabaciones_test.go:269`) sigue en verde: ese nombre no está en otro conjunto. La Ley 19/2013 está ya
  en los dos manifiestos con bloques distintos (precedente). `verificarLasNormasDeLaCobertura`
  (`grabaciones_test.go:1585`) resuelve cada entrada por su búsqueda grabada, que la unión ya tiene: la entrada nueva no
  pone en rojo `make ci` antes de grabarse, y ningún test exige que cada bloque del manifiesto esté grabado (`grep -n
  "\.Bloques" internal/evals/*.go`).
- **V7 · `awk` en los guiones.** `go test ./internal/app/ -run 'TestEntregaDelHito/h32-(paquetes|capacidades)$' -count=1 -v`
  termina con los dos en `PASS`: sus `exec awk -f …` corren bajo testscript en esta máquina (`/usr/bin/awk` es `gawk`).
- **V7b · El guion, probado.** Con `go run github.com/rogpeppe/go-internal/cmd/testscript` (v1.16.0, del `go.mod`) sobre
  un `.txtar` en `/tmp/claude/guion/` con las aserciones de contracts/aceptacion.md §2 aplicadas a los ficheros de
  `h33-prototipo` (los textos que entrega el hito, V1), sacados con `git show`: `PASS`. En el mismo guion, sobre los de
  `HEAD`, el `grep` de la plantilla no casa, `filas.awk` da `filas: 0 de 9, 0 con norma` y el tope del cuerpo casa (10 882
  > 10 860): los rojos que se esperan. `filas.awk` da `filas: 9 de 9, 2 con norma` con `gawk`, `mawk` y `busybox awk`, en
  el locale de la sesión y con `LC_ALL=C` (otro `.txtar`, `PASS`). `git status` no cambió en el repositorio.
- **V8 · La base en verde.** `go test ./internal/skills/ -run 'TestLeerCobertura|TestRegiones|TestRegionDeLaCobertura|TestVerticalDeEnsayo' -count=1`
  (28 pasan) y `go test ./internal/app/ -run 'TestEntregaDelHito/h32-(paquetes|capacidades)|TestSkillsDelRepositorio' -count=1`
  (231 pasan).
- **V9 · Tamaños.** Con `grep … | wc -m -c` sobre `git show`: las cuatro líneas de la región de lo cubierto miden 755
  caracteres y 763 bytes (la de los plazos, 306/311; la del territorio, 116/117); las 9 filas de la tabla de la
  cobertura, hoy 2980/3047, pasan a 3043/3128 (el paréntesis del spec, «la cola nueva es más corta», no se cumple: son
  81 bytes más, nueve por fila, por los dos puntos tras el asunto y la cola de 54 caracteres); la de los tributos,
  417/429; la plantilla, 89/93; la marca, 138/140.
- **V10 · El peor caso del trabajo.** `internal/evals/definicion_test.go:3096` y `:3222-3228`: `consulta-juridica`, con
  dos evals, `SesionesPorGrupo: []int{13, 12}`, concurrencia 1, 8485 s, y cada tanda dura 272 s
  (`.github/workflows/evals.yml:274`). Si cada eval añade seis sesiones por grupo —lo que da 13 y 12 con dos, y lo que
  dan 3 repeticiones por cada uno de los 2 modelos del plan—, con cuatro son `{25, 24}` y 8485 + 24 × 272 = 15 013 s, la
  cifra de FR-019, bajo el tope de 352 minutos (21 120 s). Es una cuenta, no una medida: la medida es la de
  `TestDefinicionDelJob`, que recalcula el peor caso con las evals del repositorio y fija sus valores.
- **V11 · Quién lee el `SKILL.md`.** `internal/evals/dependencias.go:986`: el `SKILL.md` de una skill es parte de las
  dependencias de sus evals; lo que `make skills-sync` escribe en la región de lo cubierto cambia sus bytes, y la
  elección de las propuestas lo ve sin código nuevo.
- **V12 · Los umbrales nacen sin caso especial.** `internal/evals/umbrales.go:379-390`: una skill tiene umbrales de sus
  respuestas si alguna eval es de cobertura, y los de cada modo llevan `sin_activar` y `cobertura_sin_declarar`, que
  deciden con 0. Con `esDeCobertura` contando `fuera_de_lo_declarado`, `consulta-juridica` los gana sin tocar
  `umbralesDelInforme`. Sus totales: 12 respuestas por modo en `sin_activar` (4 evals × 3) y 6 en
  `cobertura_sin_declarar` (2 × 3), las «6 respuestas por modo» del hito; `plazos-y-territorio`, 9 (`informe_test.go:7193`).
- **V13 · Las copias de una skill.** `internal/app/skills_test.go:2562-2565`: `copiaConLaSkill` copia la skill, las de su
  lista de la base y `data/`; la copia de `consulta-juridica` lleva `plazos-y-territorio`, cuya región genera las filas
  de `regulados_aparte`, así que el defecto «no la genera ninguna región», que ahora también da `consulta-juridica`
  (FR-006), no aparece en sus subtests.
- **V14 · La fuente.** `docs/SOURCES.md:42`: `boe.legislacion-consolidada`, «Revisado» 2026-09-13 (en `main`: la rama
  no la cambia).
- **V15 · El YAML de lo cubierto.** El lector del módulo es `go.yaml.in/yaml/v3` v3.0.6 (`go.mod:15`;
  `internal/skills/esquemas.go:18`). Un programa en `/tmp/claude/yamlprueba` con ese módulo, sacado de la caché de módulos
  (`go -C … run .`), lee el `data/cobertura.yaml` de `h33-prototipo` y da el motivo y las cuatro filas con sus alcances
  tal cual, el de los plazos con su comilla invertida dentro de un escalar sin comillas, y `materia: administrativo` solo
  en esa fila.

## Decisiones (D)

### D1 · La región de lo cubierto, con los lectores de las regiones (FR-001, FR-003, FR-007)

**Decisión.** Un fichero nuevo, `internal/skills/cubierto.go`, con la región (`regionDeLoCubierto`, un `regionDeMarcas`
con las marcas de FR-007) y su generador como método de `datosDeLasRegiones`: lee `data/cobertura.yaml` y
`data/normas.yaml` con los lectores perezosos que ya usan las regiones de las referencias (`leerUnDato`, `usar`), una vez
por sincronía, y da los mismos defectos que da hoy una región sin uno de sus YAML o con defectos en él. `regenerarSkill`
(`sincronia.go`) lo llama justo detrás de `paquetes.regenerar`, sobre el `SKILL.md` ya con sus otras regiones.
**Rechazada.** La del prototipo: `paquetes` lee `data/cobertura.yaml` por su cuenta y, si falla, deja la región sin
tocar y sin defecto (`if … err == nil`): calla el error, que FR-007 pide dar, y lee el fichero dos veces.

### D2 · La plantilla y los motivos (FR-004 a FR-006)

**Decisión.** `formasDeNoCubierto` pasa a ser `lineaDeNoCubierto{plantilla, motivoConNorma, motivoSinNorma}`:
`plantillaDelCuerpo` da la única línea de los bloques `text` del `SKILL.md` que empieza por `⚠ NO CUBIERTO: ` y lleva
`<motivo>` (con otra cantidad, ninguna), y `lineaDeNoCubierto.conLosMotivosDe`, las dos líneas de los bloques `text` de la
referencia que llevan `<precepto>` y no empiezan por `⚠`, una con `<norma>` y otra sin ella. Cada fila: primero
`<motivo>` por el motivo de su caso, y después un solo `strings.NewReplacer` con todos los marcadores conocidos
(`<asunto>` y `<procedimiento>` por el procedimiento): lo sustituido no se vuelve a mirar. Ningún texto de la línea vive
en Go (FR-006): solo las etiquetas y los marcadores, como hoy. **Rechazada.** Conservar las dos formas del cuerpo (lo que
el ADR 0058, punto 1, retira) o escribir los motivos en Go (FR-006 lo prohíbe).

### D3 · El motivo fijo en la prosa (FR-003)

**Decisión.** En toda skill con plantilla, la sincronía exige que el `SKILL.md`, con cada tramo de blancos (saltos de
línea incluidos) reducido a un espacio (`strings.Fields` y `strings.Join`), contenga `«` + motivo + `»`, con el motivo de
`data/cobertura.yaml` tratado igual. Se comprueba donde ya se sabe si la skill tiene plantilla (`datos.regenerar`), con
el `SKILL.md` regenerado. **Rechazada.** Buscarlo solo en el párrafo de §4 o en una sección por su título: la prosa de
`plazos-y-territorio` lo dice en «Cómo se presenta un plazo» y la de `consulta-juridica` en §4, partido en dos líneas;
mirar el fichero entero es lo más simple que cumple las dos.

### D4 · La materia de una fila de lo cubierto (FR-003)

**Decisión.** El generador de la región de lo cubierto comprueba cada `materia` contra los `vertical` de
`data/normas.yaml` (el mismo lector de D1) y, si alguna no es ninguno, da el defecto y deja la región como está.
**Rechazada.** Comprobarlo en el esquema (un `enum` de verticales duplicaría `data/normas.yaml`) o en `LeerCobertura`
(no conoce las normas).

### D5 · El esquema de la cobertura (FR-001, FR-002)

**Decisión.** `schemas/cobertura.yaml.json` gana `motivo_de_lo_no_declarado` (obligatorio, cadena de al menos un
carácter) y `cubierto` (lista de objetos sin claves de más, con `capacidad` del `enum` de FR-001 y `alcance`, de al menos
un carácter, obligatorios, y `materia` opcional, `{"type": "string"}` y nada más: lo que valga lo decide solo la
sincronía, D4). `regulados_aparte` no cambia un byte. **Rechazadas.** Hacer obligatoria `cubierto` o darle
`minItems`, como el prototipo: ningún requisito lo pide (FR-002 enumera los rechazos), y `TestLeerCobertura` ya fija las
cuatro filas del repositorio. Dar a `materia` el patrón de `vertical` de `schemas/normas.yaml.json`, como el prototipo:
FR-002 no enumera ese rechazo, todo `vertical` ya lo cumple, y con él `materia: Laboral` daría el defecto del esquema
en lugar del de FR-003 (contracts/datos-y-regiones.md §3). Reformatear el esquema entero (el prototipo cambia 108
líneas): solo se añade lo nuevo.

### D6 · El formato de eval (FR-012, FR-015)

**Decisión.** En `schemas/eval.yaml.json`, `no_cubierto` admite `fuera_de_lo_declarado` (`const: true`) con
`if: {required: [fuera_de_lo_declarado]}`, `then: {not: {anyOf: [{required: [norma]}, {required: [precepto]}]}}` y
`else: {required: [norma, precepto]}`; y la eval gana `tambien_la_resuelven`, lista no vacía y sin repetir de nombres
con el patrón de `se_activan`. En `formato.go`, `NoCubiertoEsperado.FueraDeLoDeclarado` y `Eval.TambienLaResuelven`;
`esDeCobertura` cuenta las dos variantes. **Rechazada.** `oneOf` (FR-012 lo descarta: cambia los mensajes de una eval
`{norma, precepto}` mal escrita, que pasarían a ser los de las dos ramas).

### D7 · El juicio de la declaración (FR-013)

**Decisión.** `NoCubiertoEsperado.declarada`: con `FueraDeLoDeclarado`, basta una declaración (`ExtraerNoCubierto`, que
ya tolera énfasis y `⚠️`); si no, la de hoy (norma y cita en la misma línea). El motivo de la eval sin declarar es
`declaración de no cubierto ausente: ` + la norma o la constante `fuera de lo declarado` (`motivoFueraDeLoDeclarado`,
en `plazos.go`). La línea `⚠ PLAZO:` sigue dando su motivo, `línea del plazo en un asunto no cubierto: <fecha>`, y
cuenta en `cobertura_sin_declarar` como hoy. **Rechazada.** Comprobar el motivo o la cola de la plantilla: FR-013 lo
prohíbe, y el despacho 44568601898 muestra respuestas correctas con otras palabras.

### D8 · `tambien_la_resuelven` (FR-015)

**Decisión.** `Eval.activadaPor(sesion, skill)`: la skill de la carpeta o alguna de `TambienLaResuelven`. La usan los dos
sitios que fijan `ResultadoDeEval.Activada` (`juzgarEnModo`, `juzgar.go:549`, y `juzgarSesion`, `informe.go:1485`), de
los que salen el veredicto de la eval y `sin_activar`. `ComprobarLasSkillsNombradas` mira también la clave nueva.
`se_activan`, `no_se_activan` (`juzgar.go:847`, `:870`) y el sondeo del laboratorio (`sondeo.go:546`) no cambian:
**rechazado** extenderlo al sondeo, que FR-015 no nombra y que mide la persona, fuera del run (FR-026; lectura
conservadora).

### D9 · Los umbrales (FR-014)

**Decisión.** Ningún código de `umbralesDelInforme` cambia (V12): solo la constante
`descripcionDelUmbralDeCoberturaSinDeclarar`, con el texto de FR-014, que no copia ningún texto de la línea.
**Rechazada.** Un caso para `consulta-juridica` (el ADR 0058, punto 6, lo excluye).

### D10 · La regla «frontera en dos materias» (FR-017)

**Decisión.** `incumplimientoDeLaFronteraEnDosMaterias`, la cuarta regla de `ReglasDeConsultaJuridica`, detrás de «una
sola capacidad»: cuenta las normas de `data/normas.yaml` sin `vertical` con una eval activa que la cita
(`normasCitadas`) y no es de cobertura y otra activa que la cita, es de cobertura con `fuera_de_lo_declarado` y prohíbe
`plazos calcular`; con menos de dos, el mensaje de contracts/evals.md §4. Las tres reglas de hoy no cambian.
**Rechazada.** Fijar las normas (Ley 50/1980, ET) en la regla: un paquete que tome una de ellas obliga a mover sus evals
a otra norma sin paquete, no a cambiar el código (principio X).

### D11 · Lo grabado: el art. 59 del ET en el manifiesto de H9 (FR-018)

**Decisión.** Una entrada nueva al final de `internal/evals/testdata/grabaciones.json` (`"fuente":
"boe.legislacion-consolidada"`): `{"busqueda": "estatuto de los trabajadores", "titulo_empieza_por": "Real Decreto
Legislativo 2/2015,", "bloques": ["a59"], "para": …}`, la búsqueda y el prefijo de la entrada del ET del manifiesto de
H5. `grabar_datos` la graba tras el commit de su tarea `[datos]` con `TestGrabarEvals` (`//go:build grabacion`,
`internal/evals/grabacion_test.go`), que pide por `internal/httpx` solo el bloque (V6) y lo deja en
`internal/evals/testdata/boe.legislacion-consolidada/`. El bloque trae todas sus versiones: `boe articulo … a59 --fecha
2026-11-02` se sirve de él sin red. Nada de `data/` sale de lo grabado y no hay material de origen: nada va a
`evidencias/H33/`. **Rechazada.** La del prototipo, `a59` en el manifiesto de H5 (`testdata/evals/grabaciones.json`):
`grabar_datos` no lo detecta y revertiría lo grabado allí (V5).

### D12 · `data/cobertura.yaml` (FR-001)

**Decisión.** Las dos claves nuevas las escribe una tarea con los textos literales de FR-001, delante de
`regulados_aparte`, y la cabecera del fichero deja de decir que solo lleva lo regulado aparte. No son datos de una
fuente externa: son la declaración de lo que kitlegal cubre, que fija el hito palabra por palabra, como H9 escribió los
procedimientos de `regulados_aparte` (research D17 de H9); ningún texto legal se copia. **Rechazada.** Derivarlas de algo
grabado: no hay fuente de la que salgan.

### D13 · La suite congelada de `plazos-y-territorio` (FR-020)

**Decisión.** En `internal/evals/laboratorio_familias_test.go`: `declaracionDelProducto` gana `enLasReferencias
[]cambioEnUnaReferencia{referencia, antes, despues}`, que `conLaDeclaracion` aplica como los del cuerpo (el `antes`, una
sola vez, sobre la referencia de la base sin lo generado); `sinLoGenerado` reduce a sus marcas también las regiones de
las referencias escritas a mano (una línea que empieza por `<!-- inicio de ` y termina en ` con make skills-sync, no
editar -->`, hasta `<!-- fin de <generador> -->`); y la declaración de `plazos-y-territorio` gana los tres cambios de
FR-009 y el de `administrativo.md` (contracts/controles.md §3). **Rechazada.** Comparar las regiones de las referencias
byte a byte con la base, que obligaría a declarar como cambio cada fila regenerada.

### D14 · La huella de `consulta-juridica` (FR-008, FR-020)

**Decisión.** `huellasDeConsultaJuridicaEnLaBase` pasa a guardar la SHA-256 del `SKILL.md` de la entrega, sin las
líneas de la clave `description` y con sus dos regiones generadas (la de lo cubierto y la de los paquetes de la base)
reducidas a sus marcas, con `conLasRegionesGeneradasReducidas`, cuya lista `regionesGeneradas` gana esas dos regiones;
`baseDeLaRama` pasa a ser `laEntrega`, «la entrega de H33» en los mensajes. La huella la calcula la tarea que escribe el
`SKILL.md`, dos veces (con el código del test y con `awk` y `sha256sum`), como H41.4. **Rechazada.** Leer la base con
`git` en el test: el `make ci` del cierre corre sobre un clon superficial (research D2 de H41.4).

### D15 · El guion de aceptación (FR-025)

**Decisión.** Un guion, `cobertura-instalada.txtar`, sobre `kitlegal skills install --dir destino --json`: `grep` con
los bytes literales de la plantilla, la marca, la región de lo cubierto y la fila de los tributos (V2); un `awk` que,
solo con `index`, `substr`, `length` y expresiones sobre cadenas, recorre la región de la tabla y cuenta las filas que
son la plantilla con su procedimiento y uno de los dos motivos (lo que un `grep` sin referencias hacia atrás no puede);
y un `! grep` con la expresión de V3 para el tope de 10 860 caracteres. Probado en verde sobre lo que entrega el hito y
en rojo sobre `HEAD` (V7b). **Rechazadas.** Contar caracteres con `awk`: `length` cuenta bytes o caracteres según la
implementación y el locale, y el tope tiene que ser de caracteres en cualquier `awk`; y una orden propia de
testscript: el guion se escribe y tiene que fallar por una aserción antes de que exista ningún código nuevo.

### D16 · Lo que deja intacto el recorte de los presupuestos (FR-007)

**Decisión.** `partirElCuerpo` (`internal/app/skills_test.go`) trata las marcas de lo cubierto como las de la tabla de
comandos: entre ellas no parte nada. **Rechazada.** Excluir solo las líneas con comilla invertida: las de normativa,
jurisprudencia y territorio no la llevan y se partirían.

### D17 · La vertical de ensayo (FR-011)

**Decisión.** `anadirEntradas` inserta las entradas de cada clave de la vertical al final del bloque de esa misma clave
en la copia (`cubierto` está en medio del fichero; `regulados_aparte`, al final); `partirPorLasMarcas` conoce las dos
marcas de lo cubierto; `comprobarLoSincronizado` exige en la región de `consulta-juridica` la línea `- **Plazos**: los de
ensayo del molde de los paquetes.`. Los ficheros de la vertical los cambia una tarea `[datos]`. **Rechazada.** Añadir la
fila de la vertical al final del fichero bajo una segunda clave `cubierto`: el lector común rechaza la clave repetida
(`TestLeerCobertura`, caso `clave-repetida`).

### D18 · Las evals (FR-016, FR-019)

**Decisión.** `evals/consulta-juridica/03-seguro-prorroga.yaml` y `04-despido-plazo.yaml` con los campos de FR-016 y un
comentario que dice qué lado de la frontera mide cada una; la 01 gana `no_cubierto` y su comentario deja de hablar de la
prosa; la 17 de `normativa` gana `tambien_la_resuelven: [plazos-y-territorio]` con el comentario de contracts/evals.md
§5. Las escribe la tarea del formato (D6), no la de aceptación: con el esquema de hoy, `fuera_de_lo_declarado` y
`tambien_la_resuelven` son claves rechazadas y `make ci` quedaría en rojo. La 04 va en la tarea siguiente a la grabación
(D11), con el subtest de `TestCacheDeLasSesiones`. `.github/workflows/evals.yml` solo cambia su comentario.

### D19 · Las suites de H32 (FR-020)

**Decisión.** `h32-paquetes.txtar` (su `tributos.txt`) y `h32-capacidades.txtar` (`|2` → `|1` y el comentario de
cabecera) cambian en la misma tarea `[datos]` que cambia el `SKILL.md` de `plazos-y-territorio` y `administrativo.md`, para
que `make ci` siga en verde; las copias de `specs/030-h32-*/aceptacion/` no se tocan.

### D20 · Lo que no se construye

Ningún paquete, applet, orden, bandera ni objetivo de `make`; ninguna dependencia. `internal/core` no cambia. La web, la
documentación y el `CHANGELOG.md` van en su tarea, sin control en `make ci` (FR-022).

## Supuestos (S)

- **S1 · Lo instalado, no lo extraído.** El guion se ha probado sobre los ficheros de `h33-prototipo` sacados con
  `git show` (V7b), no sobre lo que instala `kitlegal skills install` de la entrega, que aún no existe; que lo instalado
  sea, byte a byte, lo empotrado de `skills/` lo prueban ya los guiones de H41 a H41.4. La tarea de aceptación lo ve
  fallar por la primera aserción (rojo primero).
- **S2 · El `awk` del runner.** `ubuntu-latest` (`.github/workflows/ci.yml:31`) ejecuta hoy los guiones de H32 con su
  `awk`: no se ha podido comprobar sin red qué implementación es. `filas.awk` da lo mismo con `gawk`, `mawk` y
  `busybox awk`, en UTF-8 y en C (V7b).
- **S3 · La medida del cierre.** Que las evals salgan en verde en el cierre (FR-024, SC-004) no se puede medir en el run
  antes del cierre; la medida previa de la persona (el despacho 44595560885, según el hito) no se ha repetido aquí.
