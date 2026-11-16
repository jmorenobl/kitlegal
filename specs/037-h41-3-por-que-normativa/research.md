# Research: H41.3 · Por qué `normativa` no se activa con todas las skills instaladas

Decisiones (D), comprobaciones hechas en esta sesión (V) y supuestos que no se pueden comprobar sin red ni modelo (S).
Cada decisión dice la elegida, las alternativas y por qué (constitución, «Criterio de decisión autónoma», punto 3).
Skills aplicadas, con el reparto de `golang-how-to`: `golang-project-layout` y `golang-naming` (ningún paquete nuevo; el
fichero nuevo se llama por lo que declara), `golang-refactoring` (la clasificación de las líneas del `SKILL.md` se
comparte sin cambiar su comportamiento, con los tests de hoy en verde antes y después), `golang-testing` (tests de tabla
con `t.Parallel()`, fixtures en `testdata/`, árboles de prueba en `t.TempDir()`, testscript), `golang-error-handling`
(los errores de los argumentos, juntos con `errors.Join` como hoy; ninguno silenciado), `golang-cli` (dos argumentos de
una orden de `make` del repositorio, validados antes de hacer nada; ninguna bandera del binario) y `golang-lint`
(ningún `//nolint`).

## Decisiones

### D1. La variante: un fichero de texto por skill en `laboratorio/producto-descripciones-cortas/`

**Decisión**: la variante es el directorio `laboratorio/producto-descripciones-cortas/` con un fichero `<skill>.txt` por
skill de `skills/` y nada más; el texto de cada `description` es el fichero sin su salto de línea final, la regla con que
el arnés de H39 lee `instrucciones.txt` y `parametro.txt` (`textoDeLaVariante`, `internal/evals/laboratorio_sondeo.go:785-794`).
El nombre del directorio es el valor de `INSTALACION` que lo instala.

**Por qué**: FR-001 pide «solo cuatro textos de `description` … y nada más de ninguna skill». Un fichero por texto no
tiene sintaxis que citar ni claves que repetir: en un directorio no caben dos entradas con el mismo nombre, así que «una
repetida» (FR-002) no tiene forma de darse y la comprobación no necesita un caso para ella (D3). El texto se lee tal
cual, sin interpretar YAML, y la composición lo codifica una sola vez, al escribir el frontmatter (D5).

**Alternativas**: (a) un YAML con una clave por skill: obliga a tratar claves repetidas, valores que no son cadena y la
cita de `:` o `#` dos veces —al leer la variante y al componer—; (b) una copia de `skills/` con cada `SKILL.md`
cambiado: guarda cuerpos y referencias, que FR-001 excluye y que se quedarían atrás cuando otro hito los cambie;
(c) `laboratorio/instalaciones/producto-descripciones-cortas/`: un nivel más para una sola variante (la única que el
spec pide; las demás, fuera de alcance).

### D2. Los cuatro textos

**Decisión**: los de contracts/variante.md §2, contados en esta sesión (V2): `consulta-juridica` 417 runas,
`jurisprudencia` 420, `normativa` 426 y `plazos-y-territorio` 427; 1690 en total, frente a 3313 hoy (V1). Ninguno tiene
`<`, así que ninguno tiene una etiqueta XML de ADR 0046, punto 3. Cada uno empieza por lo que nombra su capacidad y
conserva, de su `description` de hoy, los disparadores que caben en el orden en que hoy aparecen; los que deja los
enumera contracts/variante.md §2. Criterio de lo que se deja: primero los ejemplos literales y las listas de materias o
de abreviaturas, después lo que repite otra frase del mismo texto y lo que el cuerpo de la skill ya dice (cómo se cita,
qué se dice de lo que no está delante); nunca la frase que ata la skill a su herramienta («Se activa antes de
`boe_articulo`…», «… de `cita_preparar` o `cita_cotejar`»), ni la de «aunque el agente crea saberlo», que es lo que mide la
eval 04, ni la de `consulta-juridica` «antes que las demás skills de kitlegal», cuyo cambio confundiría el remedio con
otra causa en «Activan otra skill».

**Comprobado contra la eval 04**: su pregunta nombra «el artículo 66 de la Ley General Tributaria»
(`evals/normativa/04-lgt-prescripcion.yaml:3`), que el texto de `normativa` cubre con «un artículo, ley o real decreto»;
la lista de abreviaturas (`LGT` incluida) se deja.

**Alternativas**: recortar cada texto por igual hasta 425 (cortaría frases a medias); dejar más margen bajo 1700 (dejaría
fuera más disparadores sin que el spec lo pida). La tercera persona y lo que dice cada texto no se juzgan en este hito
(FR-001; ADR 0046, punto 4).

### D3. La comprobación de la variante (FR-002) es una función que usan `make ci` y el arnés

**Decisión**: `descripcionesCortas(raiz string) (map[string]string, []string, error)` en
`internal/evals/laboratorio_descripciones.go` lee `skills/` y la variante de la raíz y devuelve las descripciones por
skill, los defectos —uno por línea, con la forma de contracts/variante.md §3— y solo como error lo que no es un defecto
de la variante (un directorio que no se puede leer: defecto `inesperado`, ADR 0023). Las skills de `skills/` son sus
subdirectorios, la regla con que `INSTALACION=producto` las copia hoy (`copiarLasDemasSkillsDelArbol`,
`laboratorio_sondeo.go:899-919`). La usan `TestDescripcionesCortas` (`make ci`, FR-002, SC-001) y la comprobación de los
argumentos del arnés con `INSTALACION=producto-descripciones-cortas` (FR-016), que antepone `INSTALACION: ` a cada
defecto.

**Por qué**: una sola definición de lo que vale; el test de la variante del repositorio y el rechazo del arnés no pueden
divergir. Leer `skills/` del árbol en cada `make ci` hace que una quinta skill sin descripción salga en rojo con su nombre
(US1, escenario 4) sin tocar el test.

**Alternativas**: dos comprobaciones (una en el test, otra en el arnés), que pueden divergir; un objetivo nuevo de
`make`, que no hace falta: `make test` ya ejecuta `go test ./...` (`Makefile:96-97`), y `make ci` depende de `test`
(`Makefile:308`).

### D4. La etiqueta XML: la definición de `internal/skills`, exportada

**Decisión**: `internal/skills` exporta `PrimeraEtiquetaXML(texto string) string`, la primera coincidencia de
`etiquetaXML` (`internal/skills/frontmatter.go:67-71`), y `defectosDeLaEtiqueta` (`:478-488`) la usa en lugar de llamar a
la expresión directamente; la comprobación de la variante la usa para el defecto de FR-002.

**Por qué**: FR-002 pide «la definición de ADR 0046, punto 3», que es una sola expresión en el repositorio.

**Alternativas**: (a) copiar la expresión en `internal/evals`: dos definiciones que pueden divergir; (b) validar cada
texto con `skills.ValidarFrontmatter` sobre un frontmatter sintético: mezcla el límite de 900 con el de 450 y da el
mensaje de un `SKILL.md`, no el de la variante.

### D5. La composición: `yaml.Marshal` de la clave `description` en el lugar de sus líneas

**Decisión**: `conLaDescripcion(contenido []byte, texto string) ([]byte, error)` sustituye, en el `SKILL.md`, las líneas de
la clave `description` del frontmatter —la de la clave y las sangradas que la siguen— por las que da
`yaml.Marshal(map[string]string{"description": texto})` de `go.yaml.in/yaml/v3`; todas las demás líneas quedan tal cual
y en su sitio. Qué líneas son de la clave lo decide la misma clasificación que `descripcionYCuerpo`
(`internal/evals/laboratorio.go:303-346`), que pasa a una función iteradora del mismo fichero que usan las dos.

**Por qué**: FR-011 pide que cada otro byte sea el de `skills/` y que la `description` leída sea el texto exacto, también
con `:`, comillas o `#` (caso límite). El emisor de YAML elige la forma del escalar que lo representa —plano si puede,
entrecomillado si no— y escapa lo que no es imprimible; no parte líneas, porque su anchura por defecto es ilimitada (V3).
Compartir la clasificación con `descripcionYCuerpo` garantiza que «lo que no es la `description`» es lo mismo para la
composición y para el test que la compara.

**Alternativas**: (a) escribir el texto con `json.Marshal` como cadena de comillas dobles: no escapa caracteres que YAML
no admite sin escapar (U+007F, controles C1); (b) un bloque plegado `>-` como el de hoy: no representa cualquier texto
(espacios dobles, espacios al final); (c) reescribir todo el frontmatter con un `yaml.Node`: vuelve a emitir las demás
claves y cambia sus bytes, contra FR-011.

### D6. `INSTALACION=producto-descripciones-cortas`: sus reglas y sus mensajes

**Decisión**: `errorDeLaInstalacion` (`laboratorio_sondeo.go:453-465`) gana el valor nuevo y recibe las variantes que
valen. De las reglas 1 a 4 se aplica la primera que se da, con una línea como mucho, como hoy; la 5 se suma:

1. vacío: ningún error;
2. ni `producto` ni `producto-descripciones-cortas`: `INSTALACION: «<v>» no es producto`, el de hoy byte a byte (FR-016;
   `gates/supuestos.md`, corrector_spec);
3. cualquiera de los dos con la familia de H39: el de hoy, `INSTALACION: las variantes del laboratorio de H39 no son de
   ninguna familia del producto`;
4. `producto-descripciones-cortas` con alguna variante que vale y no es `normativa-producto`:
   `INSTALACION: producto-descripciones-cortas solo vale con VARIANTES=normativa-producto`;
5. y, solo con `producto-descripciones-cortas`, un `INSTALACION: <defecto>` por cada defecto de la variante (D3).

Sin familia —ninguna variante vale, o las que valen se mezclan—, `producto` sigue sin dar error, como hoy; el 4 se da
igual, porque depende de las variantes y no de la familia. La variante de la regla es una constante declarada junto a la
familia de `normativa` en `laboratorio_familias.go` (`varianteDelProductoDeNormativa`, que la declaración de la familia
usa en lugar del literal de hoy, `:82` y `:87`), y el valor nuevo de `INSTALACION`, otra a su lado (clarificación de
2026-11-16). La variante solo se lee con el valor nuevo: sin él, una variante rota no cambia nada (FR-015).

**Por qué**: FR-010 y FR-016: cualquier variante que no sea solo `normativa-producto` es error de `INSTALACION` (las
versionadas, las `-producto` y versionadas de las otras dos familias), y con las de H39 se conserva el mensaje de hoy.
El 4 sin mirar la familia da también el mensaje de `INSTALACION` en una mezcla con `normativa-producto`
(`normativa-producto,jurisprudencia-producto`), que FR-010 cuenta entre las que no valen.

**Alternativas**: (a) dar el 4 solo con familia, como `producto`: la mezcla con `normativa-producto` saldría solo con el
error de la mezcla, sin nombrar `INSTALACION`; (b) aceptar cualquier variante de `normativa`: una fila etiquetada
`normativa-control` mediría la cabeza compuesta (spec, «Assumptions»).

### D7. `PRESUPUESTO`: la forma, el rango, los mensajes y el camino hasta el arnés

**Decisión**: décimo argumento de `make evals-laboratorio`, detrás de `EVAL`: la receta pasa `"$(PRESUPUESTO)"` detrás de
`"$(EVAL)"`; `scripts/evals-laboratorio.sh` exige diez argumentos y lo da a `go test` como `-presupuesto` delante de
`-temporal`; `scripts/evals-sondeo-llavero.sh` no cambia de código, porque pasa sus argumentos con `"$@"` (`:38`); y
`TestSondeoDelLaboratorio` (`internal/evals/job_test.go:96-106`) gana la bandera. La comprobación:

- vacío: ningún error, sin presupuesto;
- con valor que no es `^[1-9][0-9]*$` (`formaDeUnEnteroPositivo`, `internal/evals/sondeo.go:34`) o no cabe en un `int`
  (`strconv.Atoi`): `PRESUPUESTO: «<v>» no es un entero mayor que 0`, el texto de `SESIONES` con `EVAL`
  (`laboratorio_sondeo.go:611-626`);
- con valor que vale y la familia de H39: `PRESUPUESTO: las variantes del laboratorio de H39 no son de ninguna familia
  del producto`.

Una sola línea por argumento: la forma antes que la familia. El error va el último, detrás del de `EVAL`, en el orden de
los argumentos.

**Por qué**: FR-013 fija la forma y el rango de `SESIONES` (`0`, `0100`, `+1000000` y `99999999999999999999` no valen) y
FR-016 el error con H39. Un argumento posicional más es la forma en que llegan hoy los otros nueve.

**Alternativas**: tomar `SLASH_COMMAND_TOOL_CHAR_BUDGET` del entorno de quien lanza: lo que FR-014 prohíbe; exportar la
variable de `make` al entorno de `go test`: otra vía distinta de la de los demás argumentos, sin su comprobación en
`TestRecetasDelSondeo`.

### D8. El entorno de cada sesión

**Decisión**: con `PRESUPUESTO=N`, el entorno base de los repartos es el de hoy, `entornoDelSondeo(s.Entorno, temporal)`
(`laboratorio_sondeo.go:282`), con `SLASH_COMMAND_TOOL_CHAR_BUDGET=N` detrás; `entornoDeLaSesion` pone encima lo de cada
sesión con `sobreLaBase`, que conserva la base salvo las variables que sustituye (`internal/evals/sesiones.go:534-549` y
`:673-688`). Sin `PRESUPUESTO`, ningún cambio de código: el entorno de las sesiones es una lista cerrada —las de
`variablesQueVenLasSesionesDelSondeo`, más `PATH` y `HOME` (`internal/evals/sondeo.go:74-76` y `:1043-1061`)— y la
variable de quien lanza no llega (V7). `TestArnesConElPresupuesto` fija las dos cosas con la variable puesta en el
entorno de quien lanza, y, sin `PRESUPUESTO`, que el entorno de cada sesión es, valor a valor, el que componen esas
funciones de hoy (contracts/arnes.md §6).

**Por qué**: FR-013 («nada más de su entorno cambia») y FR-014; el spec dice que hoy ya no llega y pide el control, no
otro mecanismo. Quitarla además explícitamente sería un segundo mecanismo para lo que ya hace la lista cerrada.

**Alternativas**: pasar la variable a `entornoDeLaSesion`: tocaría la función que comparten el job y el sondeo de una
skill, que este hito no cambia.

### D9. La constancia de `sondeo.json`

**Decisión**: `sondeoLeido` (`internal/evals/laboratorio_tabla.go:71-76`) gana, detrás de `skills_instaladas`,
`instalacion` (cadena, con `omitempty`) y `presupuesto` (número entero, con `omitempty`). El arnés escribe `instalacion`
solo con `INSTALACION=producto-descripciones-cortas` y `presupuesto` solo con `PRESUPUESTO`; sin las dos opciones el
fichero es el de hoy byte a byte (FR-012, FR-015). `presupuesto` es un número porque cuenta caracteres, y su forma
canónica (sin ceros delante ni signo) hace que el número escrito sea «el valor tal como se pidió».

**Por qué**: FR-012 y su clarificación: la constancia, solo con alguna de las dos opciones. `omitempty` omite la cadena
vacía y el 0 (V5), que no son valores válidos de las opciones.

**Alternativas**: escribir siempre las dos claves: cambiaría el `sondeo.json` de hoy (FR-012, FR-015); el presupuesto como
cadena: no aporta nada, porque la forma es única.

### D10. La tabla: dos columnas, solo con la constancia, y el orden de las filas

**Decisión**: con algún sondeo con `instalacion` o `presupuesto` en su `sondeo.json`, la tabla de una familia del
producto da, detrás de «Skills instaladas», «Instalación» y «Presupuesto»; la clave de una fila gana la instalación y el
presupuesto, que solo se llenan en esa tabla, y las filas se ordenan por variante, modo, modelo, versión, instalación,
presupuesto (vacío primero, después de menor a mayor) y skills instaladas. El detalle no cambia (FR-020). Sin ningún
sondeo con la constancia, la instalación y el presupuesto de todas las claves van vacíos y la tabla es la de hoy, byte a
byte (FR-022).

**Por qué**: con la instalación delante de las skills instaladas, las cuatro filas de FR-021 salen en el orden del sondeo
—`normativa`, `producto`, `producto` con `1000000`, `producto-descripciones-cortas`— (US3, escenario 1), porque así se
ordenan los nombres; con ella detrás, la de `sola` saldría la última, porque `consulta-juridica, …` va antes que
`normativa`. Llenarla solo con la constancia deja las tablas de hoy con su orden de hoy: en ellas la instalación
deducida pondría `normativa` antes que las cuatro y cambiaría la tabla congelada de H41.2.

**Alternativas**: (a) columnas siempre: cambian las tablas fijas y las suites congeladas (FR-022, FR-032); (b) las dos
claves detrás de las skills instaladas: más simple, pero da las filas en un orden que no es el del spec.

### D11. La instalación de una fila sin la constancia

**Decisión**: en una tabla con la constancia, la instalación de un sondeo sin la clave `instalacion` es `producto` si su
`skills_instaladas` tiene alguna skill además de la de la familia, y la skill de la familia si solo la tiene a ella o no
la anota (FR-020, FR-022); el presupuesto, el suyo o vacío.

**Por qué**: el arnés solo escribe dos listas: la skill de la familia sola, o todas las de `skills/` del árbol con
`INSTALACION=producto` (`skillsInstaladasDelSondeo`, `laboratorio_sondeo.go:961-982`). Con «alguna además de la de la
familia», un sondeo de un árbol con una quinta skill sigue siendo `producto`.

**Alternativas**: comparar con una lista fija de cuatro nombres: un sondeo de un árbol con otra skill más saldría como la
skill de la familia, que no es lo que tenía instalado.

### D12. Los tests y los fixtures

**Decisión**:

- `TestDescripcionesCortas` (`laboratorio_descripciones_test.go`): la variante del repositorio, sin defectos, con una
  descripción por skill de `skills/`; y, sobre árboles que el test monta en `t.TempDir()` —un `skills/` con
  subdirectorios vacíos y la variante con sus ficheros—, los defectos exactos de contracts/variante.md §3: falta una,
  sobra una, 451 runas, 1701 en total, `<ejemplo>` y una quinta skill sin descripción. Sin fixtures en `testdata/`: el
  test escribe lo que necesita, que es de lo que trata.
- `TestConLaDescripcion` (mismo fichero): con cada `SKILL.md` de `skills/` y textos con `: `, comillas, `#` y `«»`, la
  `description` que lee `skills.LeerFrontmatter` es el texto exacto, `skills.ValidarFrontmatter` no da ningún defecto y
  lo que no es la `description` es el original, con la clasificación compartida (D5). `ValidarFrontmatter` se llama con la
  raíz del repositorio, para las referencias declaradas, y con los applets que declara el propio frontmatter como
  registrados: la composición no toca `metadata` —lo prueba la igualdad de lo demás— y que los declarados estén
  registrados ya lo exige `make skills-check` sobre `skills/` (`Makefile:308`). Lo mismo en el test del arnés.
- `TestArnesConLasDescripcionesCortas` y `TestArnesConElPresupuesto` (`laboratorio_sondeo_test.go`), con el sustituto de
  `claude` de hoy, que además vuelca su entorno en una anotación por sesión (`entorno.XXXXXX`, D13): los casos de
  contracts/arnes.md §6.
- `TestTablaConLaInstalacion` (`laboratorio_tabla_test.go`): la tabla de los sondeos fijos de
  `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/` es, byte a byte, su `tabla.md` (contracts/tabla.md
  §4 y §6), y la instalación de una fila sin la constancia, en sus cuatro casos de D11.
- El guion congelado de contracts/aceptacion.md.

**Por qué**: cada uno se traza a un FR/SC (plan.md, «Trazabilidad»). El test unitario de la tabla y el guion congelado
leen los mismos sondeos: el guion es la aceptación e2e, que corre el `laboratorio` del árbol como proceso aparte; el
unitario da la cobertura del código nuevo de `internal/evals` (principio III), como en H41.2.

### D13. El sustituto de `claude` vuelca su entorno

**Decisión**: el guion del sustituto (`guionDelClaudeDelLaboratorio`, `laboratorio_sondeo_test.go:49-62`) escribe, en cada
sesión, una anotación `entorno.XXXXXX` con la ruta de la sesión en la primera línea y la salida de `env` detrás. Los tests
nuevos leen de ella las líneas `SLASH_COMMAND_TOOL_CHAR_BUDGET=` de cada sesión y los nombres de todas sus variables.

**Por qué**: FR-018 pide el sustituto «que anota sus invocaciones y su entorno», y hoy no anota el entorno (V8). Las
anotaciones de hoy se leen por su prefijo (`sesion.`, `version.`; `anotadas`, `:105-120`), así que una más no cambia lo
que leen los tests de hoy ni sus valores esperados (FR-015). Ninguno de esos tests lee el entorno (V8), así que el de
hoy lo fija el caso sin `PRESUPUESTO`: el que vuelca cada sesión es, nombre a nombre y valor a valor, el que componen
`entornoDelSondeo`, `entornoDeLaSesion` y, en modo herramienta, `sinElDirectorioEnElPATH` con el entorno, el temporal y
el directorio de la sesión del test, salvo `PWD`, `SHLVL` y `_`, que añaden `bash` y `sh` al arrancar (V17), y como
conjunto, sin exigir el orden, que los intérpretes no tienen por qué conservar (FR-015, SC-004). «Nada más de su entorno cambia» con `PRESUPUESTO` (FR-013) se
comprueba por nombres: los de cada sesión con él son los de la misma sesión sin él, más `SLASH_COMMAND_TOOL_CHAR_BUDGET`;
los valores con rutas cambian de un `DIRECTORIO` a otro.

**Alternativas**: anotar solo la variable: no deja comprobar que no cambia nada más; un test unitario de la función que
compone el entorno: no ve lo que llega de verdad a la sesión a través de `scripts/evals-sesion.sh`.

### D14. Documentación que deja de ser cierta

**Decisión**: el comentario de `evals-laboratorio` del `Makefile` (`:157-167`), las cabeceras de
`scripts/evals-laboratorio.sh` y de `scripts/evals-sondeo-llavero.sh` («nueve argumentos»), la sección «El sondeo del
laboratorio» de `CONTRIBUTING.md` (sinopsis, tabla de argumentos y columnas) y `CHANGELOG.md` (*Unreleased*).

**Por qué**: con el hito dicen algo falso (nueve argumentos, `INSTALACION` vacío o `producto`, las columnas de la tabla).

### D15. El sondeo y el ADR son de la persona

**Decisión**: el run no lanza el sondeo de FR-030 ni escribe el ADR de FR-031; `gates/supuestos.md` lo deja dicho y el
informe final lo da en «Supuestos y pendientes». Las órdenes, en contracts/arnes.md §1 y quickstart.md §7.

**Por qué**: abre 640 sesiones con modelo y la credencial del sondeo, que ninguna sesión del run tiene (ADR 0032; FR-017).

### D16. Las evals del cierre (FR-034)

**Decisión**: el hito no toca ninguna skill, eval, umbral ni el job. Si la medición del cierre sale en rojo solo por
`sin_activar` de la eval 04 de `normativa` en modo herramienta, la reparación no la arregla ni la esquiva: el informe
final da el umbral sin cumplir con sus cifras y la deja dicha para la persona (ADR 0054, «Consecuencias»).

## Fuentes externas

Ninguna. El hito no graba ninguna respuesta ni deriva nada a `data/`: no hay manifiesto `grabaciones.json`, test
`TestGrabar*` ni tarea que grabe. Las sesiones del sondeo de la persona se sirven de las grabaciones que ya usa el job,
con los proxies que rechazan toda petición (`variablesDeLasConsultas`, `sesiones.go:558-568`). Ninguna fuente sin fila
revisada en `docs/SOURCES.md` queda fuera, porque no se usa ninguna.

## Comprobaciones hechas en esta sesión

- **V1**. Las `description` de hoy, leídas del frontmatter (bloque `>-`, líneas unidas por un espacio) y contadas con
  `gawk` 5.3.2 en `LC_ALL=C.UTF-8`, que cuenta caracteres: `consulta-juridica` 705, `jurisprudencia` 842, `normativa`
  887, `plazos-y-territorio` 879; 3313 en total, la cifra del spec.
- **V2**. Los cuatro textos de contracts/variante.md §2, escritos en ficheros de una línea fuera del repositorio y
  contados igual: 417, 420, 426 y 427 runas, 1690 en total; 1732 bytes con los cuatro saltos de línea (`wc -c`); ninguna
  línea contiene `<`.
- **V3**. `go.yaml.in/yaml/v3` v3.0.6 (la de `go.mod:15`) inicializa el emisor con `best_width: -1`
  (`apic.go:105-113` en la caché de módulos), que el emisor lleva a `1<<31 - 1` (`emitterc.go:340-341`); ni `encode.go`
  ni `yaml.go` llaman a `yaml_emitter_set_width` (`grep` sin coincidencias). `yaml.Marshal` no parte un escalar largo.
- **V4**. `testscript` de `go-internal` v1.16.0 (`go.mod:11`) nombra cada guion por su fichero sin `.txtar`
  (`testscript/testscript.go:322-341`): el subtest de este hito es `TestEntregaDelHito/h41-3-laboratorio-tabla`.
- **V5**. `go doc encoding/json Marshal` (Go 1.27.4): `omitempty` omite «false, 0, a nil pointer, a nil interface value,
  and any empty array, slice, map, or string».
- **V6**. Los intervalos de las tablas de contracts/tabla.md §4, con un programa fuera del repositorio que repite
  `recuento`, `intervaloDeWilson` y `conUnDecimal` (`laboratorio_tabla.go:1209-1233`): 0 de 1 [0,0–79,3 %], 1 de 1
  [20,7–100,0 %], 0 de 2 [0,0–65,8 %], 1 de 2 [9,5–90,5 %], 2 de 2 [34,2–100,0 %], 0 de 3 [0,0–56,2 %], 1 de 3 [6,1–79,2 %];
  y, con 160, los del spec: 0 [0,0–2,3 %], 7 [2,1–8,8 %], 8 [2,6–9,6 %] y 12 [4,3–12,7 %]. Los de 2 coinciden con la tabla
  congelada de H41.2 (`specs/036-h41-2-por-qu-e/aceptacion/laboratorio-tabla.txtar:131`).
- **V7**. El entorno de las sesiones del arnés: `Entorno: entornoDelSondeo(s.Entorno, temporal)`
  (`laboratorio_sondeo.go:282`), que solo copia de la base las variables de `variablesQueVenLasSesionesDelSondeo`
  (`sondeo.go:74-76`, `:1049-1058`) y añade `PATH` y `HOME`; encima, `entornoDeLaSesion` (`sesiones.go:534-549`).
  `SLASH_COMMAND_TOOL_CHAR_BUDGET` no está en la lista.
- **V8**. El sustituto de `claude` de hoy anota la versión y la ruta de cada sesión, no su entorno
  (`laboratorio_sondeo_test.go:49-62`).
- **V9**. La tabla ordena hoy las filas por variante, modo, modelo, versión y skills instaladas (`compararClaves`,
  `laboratorio_tabla.go:910-928`), y la tabla congelada de H41.2 da la fila de las cuatro antes que la de `normativa`
  (`laboratorio-tabla.txtar:130-131`).
- **V10**. `make test` ejecuta `go test -race -shuffle=on … ./...` (`Makefile:96-97`) y `ci` depende de `test`
  (`Makefile:308`): los tests de `internal/evals` sin etiqueta corren en `make ci`. `TestSondeoDelLaboratorio` solo
  compila con `-tags evals` (`internal/evals/job_test.go:1`).
- **V11**. `go.yaml.in/yaml/v3` es un módulo permitido en el cierre del binario (`internal/arch_test.go:221-227`) y
  ninguna regla de `depguard` restringe quién lo importa (`.golangci.yml`, reglas `core`, `red`, `sql`, `grafo`, `mcp` y
  `ejemplo`); `internal/evals` no está en el cierre de `cmd/kitlegal`.
- **V12**. Los umbrales del job con `decide: true` que el cierre de H41.2 dio como publicados: 11 de `normativa`, 11 de
  `jurisprudencia`, 12 de `plazos-y-territorio` y ninguno de `consulta-juridica`
  (`specs/036-h41-2-por-qu-e/gates/informe-final.md:12` y `:326`).
- **V13**. La toolchain: `go version` da `go1.27.4 linux/amd64`, la de `go.mod:5`.
- **V14**. `leerElContenidoDelLaboratorio` recorre todos los ficheros de `laboratorio/` (`laboratorio_test.go:1005-1031`):
  `TestLaboratorioFueraDelProducto` cubre los de la variante sin cambiar.
- **V15**. `yaml.Marshal(map[string]string{"description": texto})` de `go.yaml.in/yaml/v3` v3.0.6, en un módulo de prueba
  fuera del repositorio con el `go.sum` del repositorio y la caché de módulos, sin red: con cada uno de los cuatro textos
  de contracts/variante.md §2 y con `a: b`, `con "comillas"`, `con # almohadilla`, `- guion` y `«sin nada»`, da una sola
  línea; entrecomilla con comillas simples los que llevan `: `, ` #` o empiezan por `- `, deja planos los demás, no escapa
  los caracteres no ASCII, y `yaml.Unmarshal` de lo que da devuelve el texto exacto en los nueve casos.
- **V16**. `go list -deps ./cmd/kitlegal` no tiene `internal/evals` y sí `go.yaml.in/yaml/v3`, que ya llega por
  `internal/skills`: importarlo desde `internal/evals` no cambia el cierre del binario.
- **V17**. Lo que añaden los intérpretes al entorno que vuelca el sustituto, con un programa Go fuera del repositorio que
  lanza, con `exec.Cmd.Env` dado y `Dir` en `trabajo/`, un guion `#!/usr/bin/env bash` que hace `exec` de un sustituto
  `#!/bin/sh` (dash, `/bin/sh -> dash`) que ejecuta `env`, como `scripts/evals-sesion.sh:39` y el sustituto de hoy: cada
  variable dada llega con su valor, y se añaden solo `PWD` y `SHLVL`. `_` no sale con dash; lo pone `bash` cuando es él
  el `/bin/sh` (macOS), que no se ha medido aquí, y por eso va en la excepción. La sesión está en el directorio de
  `exec.Cmd.Dir` y `evals-sesion.sh` no cambia de directorio, así que no hay `OLDPWD`.

## Supuestos no verificados

- **S1**. Que Claude Code lee `SLASH_COMMAND_TOOL_CHAR_BUDGET` del entorno de la sesión como presupuesto del listado, y
  que `1000000` no recorta un listado de 21 skills y 49 órdenes: lo dicen el hito y la documentación que cita
  (https://code.claude.com/docs/en/skills), que esta sesión no puede leer sin red. Si no fuera así, la fila de diagnóstico
  lo mostraría al no separarse del control, y el ADR de FR-031 lo diría («no se demuestra»).
- **S2**. Que Claude Code lee la `description` de un frontmatter escrito con un escalar entrecomillado igual que la de
  uno escrito con `>-`: es YAML estándar, y `skills.LeerFrontmatter`, con la misma biblioteca que valida las skills, lee
  el texto exacto (`TestConLaDescripcion`); el lector de Claude Code no está en el repositorio.
