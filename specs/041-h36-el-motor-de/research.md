# Research: H36 · El motor de cómputo descrito por datos

Decisiones del plan (D), con la alternativa descartada y por qué, verificaciones hechas en esta sesión contra el
repositorio y la caché de módulos (V) y supuestos que no se han podido comprobar sin red o sin ejecutar el binario (S).
Ninguna decisión reabre el ADR 0061, el ADR 0042 ni las decisiones cerradas de `CLAUDE.md`; el contrato es el §C del
spec, y aquí solo se decide cómo se construye.

## Decisiones

### D1 · La API: `LeerTipos` sigue dando `[]Tipo`, y cada tipo lleva su cómputo leído

- **Decisión**: `LeerTipos(contenido []byte) ([]Tipo, error)` no cambia de firma. `Tipo` conserva `Codigo`, `Nombre`,
  `Computo` (el código, del tipo `Computo string`, que se queda), `Duracion`, `Preceptos`, `Unidades` y
  `PreceptosDelComputo`, y gana `Da` (exportado, porque el lector YAML solo rellena campos exportados) y un campo sin
  exportar con el cómputo descrito que `LeerTipos` le copia, igual que hoy copia `PreceptosDelComputo`. El cómputo
  descrito (municipios, unidades, periodos, preceptos y su posición en el fichero) es un tipo sin exportar
  (data-model §1). Desaparecen las constantes de FR-010 y el tipo exportado `PreceptosDeUnComputo`, que solo usa
  `tipos.go` (V9) y que la forma nueva sustituye.
- **Alternativas**: (a) que `LeerTipos` devuelva una estructura con tipos y cómputos: cambia todas sus llamadas y tests
  que FR-017 no deja tocar, como `internal/evals/preparar_test.go:1062` (V9); (b) exportar el cómputo descrito: no lo
  lee nadie fuera del paquete (las referencias no cambian, FR-020, y las evals solo piden códigos y etiquetas, FR-009).
- **Por qué**: es la superficie más pequeña que cumple FR-001 a FR-009 y deja fuera del diff los tests que FR-017 no
  nombra.

### D2 · Un solo algoritmo, con los ejes como datos (FR-004, FR-005; §C.3)

- **Decisión**: `computar` deja de preguntar por el código del cómputo o del tipo. Con la unidad de la duración:
  días hábiles, el enésimo día no inhábil; días naturales, el enésimo día y la prórroga; meses, la partida y el
  vencimiento de cada periodo `no-corre` con la regla de research D9 de H9 escrita para un mes cualquiera (el día de
  `inicio` o el siguiente que cae en él lleva la partida a su último día; un tramo que pasa por él lleva el
  vencimiento al mismo día del mes siguiente) y la prórroga. `inhabil` es el de hoy más los días de los periodos
  `inhabil` del cómputo; `pasaPor` recibe un periodo `MM-DD` a `MM-DD`, que puede cruzar el año. Con `da:
  dia-siguiente`, el día siguiente al vencimiento, sin prórroga, y la comprobación del 9999-12-31 de hoy.
- **Alternativas**: conservar `lpac` y `contencioso` como funciones elegidas por un dato: el código seguiría decidiendo
  qué se cuenta, contra el principio X y FR-010 (lo rechaza el ADR 0061, punto 4).
- **Por qué**: con los datos de §C.1, `lpac` (sin periodos) y `contencioso` (agosto `no-corre`, del 24 de diciembre al 6
  de enero `aviso`) dan lo de hoy (D9 de H9 sin cambios), y el civil y el ensayo salen del mismo código (§C.9, §C.10).

### D3 · El orden de los avisos y de `CodigosDeAviso` (FR-008, FR-009)

- **Decisión**: los avisos de periodo de un plazo van detrás de los de los festivos que faltan y delante de los de una
  parte, en el orden de los periodos de su cómputo. `CodigosDeAviso(tipos []Tipo)` da los dos del motor de los
  festivos que faltan, los códigos de aviso de los cómputos que nombran los tipos —ordenados por la posición del
  cómputo en el fichero, que `LeerTipos` anota, y en cada uno por la de sus periodos— y `festivo-de-una-parte`.
  `EtiquetaDeAviso(tipos, codigo)` da la de los tres del motor o la del periodo que la lleva.
- **Alternativas**: ordenar los cómputos por el primer tipo que los nombra: no es el orden del fichero que fija FR-009.
- **Por qué**: FR-009 lo dice literal y cuesta un entero por cómputo. Con los datos de hoy, los cuatro de hoy en su
  orden (V13).

### D4 · `internal/evals` lee los tipos del binario una vez (FR-009)

- **Decisión**: `internal/evals/plazos.go` lee `plazos.LeerTipos(data.Plazos)` una sola vez (`sync.OnceValues`) y de
  ahí saca los códigos y las etiquetas de `formasDeAvisoDelPlazo`, `ExtraerAvisosDelPlazo` y
  `ComprobarCodigosDeAvisoDelPlazo`, que mantienen sus firmas. `ComprobarCodigosDeAvisoDelPlazo` devuelve el error de
  lectura si lo hay; `ExtraerAvisosDelPlazo` reconoce los avisos de los tipos que ha leído. Un `data.Plazos` que no
  carga es la regla genérica: no pasa `make ci` (`TestLeerTipos`, `TestAppletPlazos`, la comprobación del esquema de
  eval de `formato_test.go`) y el binario sale con 1 en cada `calcular`, así que ninguna eval de plazo pasa. No se
  diseña nada más para ese estado.
- **Alternativas**: (a) pasar los tipos por el juicio de cada eval (`juzgarElPlazo`, que no devuelve errores): cambia
  la fontanería del juicio por un estado que no pasa `make ci` (criterio n); (b) cambiar la firma de
  `ExtraerAvisosDelPlazo`: su test (`plazos_test.go`) no está entre los que deja tocar FR-017; (c) leer
  `data/plazos.yaml` del directorio de datos como `consultas.go`: los avisos que se juzgan son los del binario de la
  sesión, que empotra `data.Plazos`.
- **Por qué**: `internal/evals` ya depende de `internal/app`, que importa `data` (`internal/app/registro.go:13`), así que
  la importación directa no añade nada al conjunto de lo que mide (V10); ninguna regla de `depguard` la prohíbe (V11).

### D5 · El esquema de `calcular`: etiquetas fijas (FR-012; §C.7, §C.8)

- **Decisión**: las cinco etiquetas `help` de `argumentosDeCalcular` pasan a los textos de §C.7, byte a byte; `Tipo` de
  `Plazo` y `Codigo` de `Aviso` llevan `jsonschema:"pattern=^[a-z0-9]+(-[a-z0-9]+)*$"` y ningún `enum`; los demás
  enumerados y la descripción del verbo no cambian. `--help` y `--describe` ya no cargan los datos hoy: el análisis es
  perezoso (`sync.OnceValues` en `AppletPlazos`, `internal/app/plazos.go:59-65`) y el caso «sin-los-tipos» lo
  comprueba (`plazos_test.go:607-640`). El test de US4.1 compara las tres salidas con `describe.json`.
- **Alternativas**: componer la ayuda o los enumerados desde los datos: lo rechaza el ADR 0061, punto 7, y el spec
  («Fuera de alcance»).
- **Por qué**: el patrón no lleva comas, así que invopop lo toma entero (V3).

### D6 · El test de FR-010: en `internal/app`, sobre el Go de los cuatro paquetes

- **Decisión**: `internal/app/plazos_codigos_test.go`, `TestNingunCodigoDePlazosEnGo`. Lee los códigos de tipo, de
  cómputo y de aviso de periodo de `data/plazos.yaml` y de `internal/core/plazos/testdata/computo-de-ensayo.yaml`
  con una estructura YAML propia del test (sin `KnownFields`: solo saca códigos; `LeerTipos` ya valida los dos
  ficheros en sus tests). Recorre con `go/parser.ParseFile` cada `.go` que no es `_test.go` de
  `internal/core/plazos`, `internal/app`, `internal/skills` e `internal/evals` (sin subdirectorios), y da un defecto
  por cada `*ast.BasicLit` de cadena que, sin comillas (`strconv.Unquote`), es entero un código, y por cada etiqueta
  de campo (`ast.Field.Tag`) que lo lleva como palabra entera, con el límite de palabra de `grep -w`
  (`\b<código>\b` de `regexp`, ASCII), que es con el que el spec comprobó el estado de hoy. Los tres avisos del motor
  no son códigos de esos ficheros y la descripción del verbo no es una etiqueta ni ningún literal suyo es entero un
  código (V14): la exclusión del spec no pide código propio.
- **Alternativas**: (a) en `internal/core/plazos`: sus `_test.go` no pueden importar `os` (depguard R1 vigila los tests
  del dominio, `.golangci.yml:100-117`) ni empotrar `../../../data/plazos.yaml`; (b) en `internal/arch_test.go`: sus
  reglas son de importación, no de los datos de un applet; (c) un límite de palabra que cuente el guion como letra:
  dejaría pasar una ayuda que diga «contencioso-administrativo» en una etiqueta, y no es el criterio con que el spec
  midió el estado de hoy.
- **Por qué**: `internal/app` lee ficheros, ya compone los datos del applet y es donde vive la ayuda que el test vigila.
  `parser.ParseDir` se evita porque devuelve `ast.Package` (V5).

### D7 · Los tipos de prueba se añaden al fichero base por texto

- **Decisión**: `internal/core/plazos/testdata/tipos-de-prueba-civil.yaml` es un documento con solo `tipos:` y los dos
  tipos de FR-023. Los tests lo combinan con su fichero base sustituyendo su única línea `tipos:` por el contenido del
  fichero de prueba (que empieza por `tipos:` y sigue con sus elementos, con la misma sangría), tras exigir que el base
  tiene exactamente una; después, `LeerTipos`. El base es `data/plazos.yaml` en `internal/app` y la copia del contrato
  (`plazosDelContrato`) en `internal/core/plazos`. La copia con un tipo y un aviso de periodo más de US4.1 se hace igual,
  por sustitución de texto, como hace `sustituida` en `grabaciones_test.go`.
- **Alternativas**: (a) combinar con `yaml.Node`: más código para lo mismo, sobre dos ficheros del repositorio que
  fijan sus propios tests; (b) que `LeerTipos` admita varios contenidos: API de producción para un test.
- **Por qué**: es lo más simple que cumple FR-023 y FR-024; un cambio de forma del base lo detecta el `require`.

### D8 · Dónde va cada test de cálculo (FR-022, FR-023, FR-024)

- **Decisión**: C1 a C10, y que ningún tipo de `data/plazos.yaml` lleva el civil (US2.5, US5.1), en
  `internal/app/plazos_civil_test.go`, `TestComputoCivil`: los festivos reales de `data/festivos/` y el
  `data/plazos.yaml` real (`fuentesDePlazos(t)`), Madrid resuelto con el registro de territorio, como hace
  `comoLoCalculaElDominio`, y `plazos.Calcular`. E1 a E13 en `internal/core/plazos/ensayo_test.go`,
  `TestComputoDeEnsayo`, con la rejilla de las propiedades y el fichero de ensayo empotrado (`//go:embed`, como
  `internal/core/ids/errores_test.go:156`). Las propiedades del civil, en `TestPropiedadesDelComputo`, sobre la
  copia del contrato, que con FR-017 pasa a la forma nueva de §C.1 con el cómputo civil, más los tipos de prueba
  empotrados.
- **Alternativas**: importar `data` desde un test del dominio para leer el fichero real: `depguard` no lo prohíbe, pero
  la regla R1 dice que el dominio no importa lo empotrado y recibe los bytes de `internal/app`; la copia del contrato es
  como H9 ya alimenta las propiedades (`tiposDelContrato`).
- **Por qué**: FR-023 pide el fichero y los festivos reales, que solo lee `internal/app`; FR-024 nombra
  `TestPropiedadesDelComputo`, que vive en el dominio y usa la rejilla.

### D9 · La grabación de H36 (FR-016)

- **Decisión**: la tarea `[datos]` añade al manifiesto de grabación del paquete, `internal/evals/testdata/grabaciones.json`
  (`"fuente": "boe.legislacion-consolidada"`), una entrada de la Ley 1/2000 con los bloques `a130` y `a133`, y hace que
  `TestGrabarEvals` (`//go:build grabacion`) grabe lo que falta bajo `$KITLEGAL_EVIDENCIAS` cuando el paso
  `grabar_datos` la da, y bajo la raíz de su conjunto si no (como hoy, para `scripts/grabar-evals.sh`). `Preparar` siembra
  la caché desde `UnionDeGrabaciones`, así que solo faltan los dos bloques: la búsqueda, los metadatos y el índice de la
  LEC ya están en el conjunto de H5 (V16). Quedan en `evidencias/h36/boe.legislacion-consolidada/`, y `grabar_datos`
  escribe `evidencias/h36/manifiesto.json` con la huella de cada uno: es el manifiesto nuevo de H36 que pide el hito.
  Tras la grabación, otra tarea añade `GrabacionesDeH36 = "../../evidencias/h36/" + boe.NombreDeLaFuente` a
  `UnionDeGrabaciones`, detrás de los de H4, H5 y H9 (contracts/grabacion.md).
- **Alternativas**: (a) un manifiesto de peticiones en `evidencias/h36/grabaciones.json`: ninguna tarea escribe en
  `evidencias/` (precheck de `tasks`; guardián) y `grabar_datos` solo graba cuando cambia un
  `<paquete>/testdata/grabaciones.json` (V7); (b) grabar en la raíz de H9, como H26.1 a H33: FR-016 pide
  `evidencias/h36/` y una raíz nueva en la unión; (c) añadir la raíz a la unión en la misma tarea que el manifiesto: hasta
  que se graba, el directorio no existe y `copiarGrabaciones` falla en cada test que prepara la unión (V8).
- **Por qué**: es la única forma de cumplir FR-016 con el paso `grabar_datos` tal como está, sin que ninguna tarea
  escriba en `evidencias/` ni toque la red.

### D10 · Los tests que FR-016 obliga a tocar fuera de FR-017, y lo que cambia por FR-015

- **Decisión**: dos desviaciones de la lista cerrada de FR-017, las dos por FR-016 y acotadas a lo que exige:
  1. `TestGrabacionesSinSolape` (`grabaciones_test.go:269-298`) exige hoy que la unión sean exactamente H4, H5 y H9;
     su lista esperada gana `GrabacionesDeH36` y nada más: su comprobación de que dos conjuntos solo comparten la
     grabación de `robots.txt` cubre el de H36 sin cambios, porque el cliente que graba pide `robots.txt` y lo graba en
     su raíz (V22).
  2. `TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`) construye hoy el cliente de cada
     raíz con `httpx.ConRaizDeGrabacion(raiz)`, la de su conjunto (línea 94); cambia solo esa raíz, que es
     `$KITLEGAL_EVIDENCIAS` cuando la variable no está vacía, como la da el paso `grabar_datos` (D9), y la de su
     conjunto si lo está, como hoy, para `scripts/grabar-evals.sh`. Nada más del test cambia.

  Por FR-015, en `grabaciones_test.go` cambian solo `verificarPreceptos`, que gana las lecturas de los cómputos sin
  tipos (D11), y la tabla `palabrasClaveDeLosPreceptos`. `lecturasDeLosTipos` y `loQueLlevaALeerElRepositorio` no
  cambian (V23).
- **Alternativas**: no añadir la raíz a la unión, o grabar en la raíz del conjunto de H9: incumplen FR-016 (D9 (b)).
- **Por qué**: FR-017 limita lo que cambia en `grabaciones_test.go` a lo de `TestPreceptosDeLosPlazos` y no nombra
  `grabacion_test.go`; la lista de la unión y la raíz del cliente son consecuencia directa de FR-016: sin la primera,
  `make ci` sale en rojo; sin la segunda, lo grabado no queda en `evidencias/h36/`. Se registran aquí y en
  «Decisiones» del plan.

### D11 · `TestPreceptosDeLosPlazos` lee los cómputos sin tipos (FR-015)

- **Decisión**: `verificarPreceptos` añade, detrás de `lecturasDeLosTipos(tipos)`, las lecturas de los preceptos de
  cada cómputo que no nombra ningún tipo, en el orden del fichero y nombradas «el cómputo `<código>`», como las de hoy.
  Las saca de los mismos bytes que ya lee, con una estructura YAML propia del test (`computos[].codigo` y
  `computos[].preceptos`, en `plazos.Precepto`; `go.yaml.in/yaml/v3`, que los tests de `internal/evals` ya importan en
  `formato_test.go` y `medida_test.go`). Con los datos de H36 son los cinco preceptos del civil. `lecturasDeLosTipos` y
  `loQueLlevaALeerElRepositorio` no cambian. La tabla `palabrasClaveDeLosPreceptos` gana las palabras de FR-015 en
  `a130` y `a133`, que solo lee `verificarPreceptos` (por `lecturasDeLaTabla`), no el oráculo.
- **Alternativas**: (a) que `lecturasDeLosTipos` lea todos los cómputos: es también la fuente de
  `loQueLlevaALeerElRepositorio` (`grabaciones_test.go:1214`) y de lo nombrado por los tipos en `TestLecturasDeLosDatos`
  (`consultas_test.go:586`), y ese oráculo lo comparan `TestLecturasDeLosDatos` (`consultas_test.go:569-573`, con
  `ElementsMatch` contra lo que prepara `consultas.go` desde `bloquesDeLosTipos`, que solo lee los cómputos de los
  tipos: `consultas.go:287`, `:340-359`) y `TestEvalsDelRepositorio/lo-que-lleva-a-leer` (`conjunto_test.go:3299-3304`).
  El civil entraría en el oráculo con cuatro órdenes de más (`boe articulo BOE-A-2000-323 a130` y `a133`, con y sin
  `--fecha 9999-12-31`; el índice y los metadatos de la LEC ya están, porque la LEC está en `data/normas.yaml`) y los
  dos tests saldrían en rojo. Ponerlos en verde exigiría cambiar `consultas.go`, con lo que el civil entraría en lo que
  la skill lleva a leer, contra FR-014, o tocar `consultas_test.go` y `conjunto_test.go`, que FR-017 no nombra.
  (b) Leer en `verificarPreceptos` todos los cómputos sin repetir los de los tipos: el mismo resultado con una
  comparación más. (c) Exportar del dominio un lector de cómputos: API de producción que solo usa un test.
- **Por qué**: FR-015 pide leer «también los cómputos sin tipos», y FR-014, que el civil no entre en lo que la skill
  lleva a leer en las evals, que sigue saliendo de los tipos. Las lecturas y los defectos de hoy salen iguales, y las
  del civil, detrás de las de los tipos.

### D12 · Las copias congeladas de H39 y su huella

- **Decisión**: la tarea `[datos]` de FR-017 cambia las cinco líneas de la ayuda de `calcular` (39 a 43) en
  `internal/app/testdata/script/h39-describe.txtar` y en su copia `specs/031-h39-c-omo-se/aceptacion/describe.txtar`,
  idénticas (V15). `specs/031-h39-c-omo-se/gates/aceptacion-congelada.json` no se toca.
- **Alternativas**: actualizar también esa huella, como hizo H33 con la suya.
- **Por qué**: el guardián solo vigila la suite congelada del hito en curso (`guardian-diff.sh:108-117`), ningún test la
  lee (V15) y FR-017 nombra solo el guion y su copia: lo no especificado no se implementa.

### D13 · El error nuevo y el orden de las comprobaciones (FR-006, FR-007; §C.6)

- **Decisión**: en `leerDuracion`, tras «la unidad no es del vocabulario» y solo en un tipo sin duración, «el cómputo no
  la admite», con el mensaje de §C.6 y las unidades del cómputo en el orden de `Unidades()` unidas con `alternativas`;
  después, el número de municipios, con el mensaje compuesto con los textos del cómputo. Los dos son
  `errorDeArgumentos` (clase `argumentos`, código 2).
- **Alternativas**: comprobar la unidad del cómputo después de los municipios: contradice el orden de FR-007.
- **Por qué**: es el orden de research D10 de H9 con un paso más donde lo pone el spec.

### D14 · `data/plazos.yaml` se fija desde el spec y lo comprueba lo grabado

- **Decisión**: el contenido de la rama es el de contracts/datos-de-plazos.md §1, que escribe §C.1 completo: los
  preceptos de hoy sin cambiar, `da` en cada tipo detrás de `duracion` y delante de `preceptos` o `unidades`, los tres
  ejes de cada cómputo y el civil. No sale de un derivador: el fichero no copia texto de ninguna norma, nombra cada
  precepto por su identificador, su bloque y su apartado, y `TestPreceptosDeLosPlazos` comprueba contra lo grabado que
  cada bloque existe, que lleva sus palabras clave y que cada `vigente_desde` es el de sus redacciones grabadas.
- **Alternativas**: derivar el fichero de lo grabado: lo que dice es una lectura jurídica (qué precepto fija qué), no un
  dato que se extraiga; H9 lo fijó igual («los fija este plan desde el spec, como `data/normas.yaml`, y los comprueba lo
  grabado», `specs/023-h9-internal-core-plazos/plan.md:399`).
- **Por qué**: es el precedente del repositorio para este fichero, y la verdad terreno (lo grabado) lo comprueba en
  `make ci`.

### D15 · Las skills de Go

`golang-how-to` se ha cargado en esta sesión; su tabla de carga reparte este hito entre `golang-refactoring` (el
cambio por etapas de un paquete existente), `golang-naming`, `golang-code-style`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-testing` y `golang-stretchr-testify`. Las secundarias no se han cargado; se aplican
las reglas que el repositorio ya sigue con ellas: ningún paquete nuevo (`golang-project-layout`); tipos sin exportar
para lo que no sale del paquete y vocabularios que devuelven una lista nueva en cada llamada (`golang-safety`);
defectos acumulados con `errors.Join` y errores tipados con su clase (`golang-error-handling`); tests de tabla con
`t.Parallel`, `require` para las premisas y `assert` para lo comparado, goldens con su bandera `-actualizar-…`
(`golang-testing`, `golang-stretchr-testify`); ningún `//nolint` ni `t.Skip`. Sin CLI nueva: la ayuda es de Kong y no
cambia la gramática.

## Verificaciones (en esta sesión)

| # | Qué | Dónde se comprobó |
|---|---|---|
| V1 | Go 1.27 (`go 1.27.0`, `toolchain go1.27.4`); kong v1.17.0, invopop/jsonschema v0.14.0, go-internal v1.16.0, santhosh-tekuri/jsonschema v6.0.3, testify v1.12.1, `go.yaml.in/yaml/v3` v3.0.6: ninguna nueva | `go.mod` (`grep -n`) |
| V2 | Entre `cd3ad148` y la cabeza solo cambian `docs/` y `specs/`: el binario de la rama antes de la primera tarea de producto es el de `main` | `git diff --name-only cd3ad148 HEAD` filtrado, vacío |
| V3 | invopop parte las etiquetas `jsonschema` por las comas sin escapar y toma `pattern=` entero tras el primer `=` | `~/go/pkg/mod/github.com/invopop/jsonschema@v0.14.0/reflect.go:615`, `:774-783`, `:1153-1174` |
| V4 | `go.yaml.in/yaml/v3` ignora los campos sin exportar; tiene `Node` con `Kind` y `Content`, y `Decoder.KnownFields` | `~/go/pkg/mod/go.yaml.in/yaml/v3@v3.0.6/yaml.go:547`; `go doc go.yaml.in/yaml/v3 Node` y `Decoder.KnownFields` |
| V5 | `go/ast.BasicLit` (`Kind`, `Value`), `go/ast.StructType`, `go/parser.ParseFile`; `ParseDir` devuelve `map[string]*ast.Package` | `go doc go/ast BasicLit`, `go doc go/ast StructType`, `go doc go/parser ParseFile`, `go doc go/parser ParseDir` |
| V6 | testscript admite `stdout -count=N` y `cmp stdout <fichero>`; los guiones de hoy usan `exec sh -c` y `cmp stdout` | `go doc github.com/rogpeppe/go-internal/testscript` (líneas 171-208); `internal/app/testdata/script/boe-cache-rapida.txtar:20`, `h9-plazos-codigos.txtar` |
| V7 | `grabar_datos` graba solo si la tarea cambia un `<paquete>/testdata/grabaciones.json`, ejecuta `TestGrabar*` del paquete con `KITLEGAL_RECORD=1` y `KITLEGAL_EVIDENCIAS=$PWD/evidencias/<hito>`, admite solo ficheros nuevos bajo `<paquete>/testdata/` y `evidencias/<hito>/`, y anota cada evidencia en `evidencias/<hito>/manifiesto.json` con su huella; el hito H36 da `evidencias/h36` | `scripts/workflow/grabar-datos.sh:100-150`; precheck de `tasks`, `scripts/workflow/precheck.sh` (tareas que escriben en `evidencias/`) |
| V8 | `copiarGrabaciones` falla con un conjunto que no existe; `Preparar` y `copiaDeLaUnionDeGrabaciones` copian la unión entera | `internal/evals/preparar.go:495-500`; `internal/evals/grabaciones_test.go:676-686` |
| V9 | Fuera de `internal/core/plazos`, nadie usa `PreceptosDeUnComputo`, las constantes `Tipo…`, `Computo…` ni `AvisoDiasDelArt183` en código que no es de test; en tests, `CodigosDeTipo` y los enumerados en `internal/app/plazos_test.go:260,410-412`, las constantes en `internal/skills/referencias_test.go:300-336`, `plazos.Computo` en `internal/evals/grabaciones_test.go:1118`, `PreceptosDelComputo` en `preparar_test.go:1062` y `grabaciones_test.go:1142` | `grep -rn` sobre `internal/` |
| V10 | `internal/evals` importa `internal/app`, que importa `data` | `grep` de importaciones en `internal/evals/*.go` e `internal/app/registro.go:13` |
| V11 | `depguard` no tiene ninguna regla sobre `github.com/jmorenobl/kitlegal/data`; R1 vigila también los `_test.go` del dominio y les deniega `os` e `io` | `.golangci.yml:78-200` (`grep`); comentario de `.golangci.yml:100-111` |
| V12 | `ComprobarCodigosDeAvisoDelPlazo` no tiene ninguna llamada fuera de los tests | `grep -rn` en `internal/` y `cmd/` (sin resultados) |
| V13 | Los cuatro códigos de hoy y su orden en el esquema de eval publicado y en `describe.json` | `internal/core/plazos/avisos.go:12-23`; `jq` sobre `internal/app/testdata/plazos/describe.json` |
| V14 | Con `grep -w`, ningún literal ni etiqueta de Go que no es de test de los cuatro paquetes lleva `civil`, `ensayo`, `ensayo-fijo`, `ensayo-libre`, `periodo-de-ensayo` ni `lpac` salvo identificadores y un literal más largo (`"10-plazo-reposicion-anterior-a-la-lpac.yaml"`, que no es entero un código) | `grep -rnw` sobre `internal/core/plazos`, `internal/app`, `internal/skills`, `internal/evals` |
| V15 | `h39-describe.txtar` y su copia congelada son idénticos y comparan la ayuda de los cinco argumentos en las líneas 39 a 43; ningún test lee `specs/*/gates/aceptacion-congelada.json` | `diff`; `grep -n`; `grep -rln aceptacion-congelada --include=*_test.go internal` (vacío) |
| V16 | La LEC está en el manifiesto de H5 sin bloques (`"busqueda": "enjuiciamiento civil"`, `"titulo_empieza_por": "Ley 1/2000,"`), y sus metadatos y su índice están grabados en `testdata/evals/boe.legislacion-consolidada/`; el índice grabado lleva los ids `a130` y `a133`; ningún bloque de `BOE-A-2000-323` está grabado; la LEC no está en `data/cobertura.yaml`, y el manifiesto de H9 ya repite una norma del de H5 (el Código Civil) | `grep -n 'Ley 1/2000' testdata/evals/grabaciones.json`; `ls` de los tres conjuntos; `grep -o` sobre el índice grabado; `grep -c` en `data/cobertura.yaml` (0); `jq` del manifiesto de H9 |
| V17 | La fila de `boe.legislacion-consolidada` en `docs/SOURCES.md` de `main` tiene «Revisado» 2026-09-13 | `git show main:docs/SOURCES.md` (línea 42) |
| V18 | Las órdenes de los escenarios: `TestAppletPlazos` (19 en verde), `TestPropiedadesDelComputo` y `TestLeerTipos` (48), `TestRegiones` (21), `TestEntregaDelHito/^h9-plazos` (6), `TestCasosDePlazos` (22), `TestPreceptosDeLosPlazos`, `TestGrabacionesSinSolape` y `TestManifiestoDeGrabaciones` (25), `TestEsquemasPublicados` (17), sin cambios en el árbol | `go test -count=1 -run …`, esta sesión; `git status --porcelain` después |
| V19 | Tamaños: `caso-a.json` 767 B, `caso-b.json` 853 B (con un aviso de 196 B de texto), `caso-e.json` 670 B, `describe.json` 8317 B (353 líneas), `schemas/plazo.json` 10 093 B; en `describe.json`, la propiedad `tipo` mide 121 B y `codigo` del aviso 143 B, compactas; el texto del aviso del art. 183 con su prefijo, 335 B, y su objeto JSON entero, 380 B; el mensaje de C7, 115 B; el aviso del ensayo, 113 B; con el binario de `main` (V24), el sobre del caso 11, 840 B, y los de los casos 15 y 16, 1065 B cada uno (805 y 1030 B con `fecha_consulta` vaciado) | `wc -c`, `jq -c … \| wc -c` y `printf … \| wc -c`; `./bin/kitlegal plazos calcular … --json \| wc -c`, esta sesión |
| V20 | El paso 9 de `plazos-y-territorio` pide `plazos_calcular` o `kitlegal plazos calcular`; la tabla de comandos (línea 143) lleva la descripción del verbo | `grep -n` en `skills/plazos-y-territorio/SKILL.md` |
| V22 | El cliente que graba graba también las peticiones que origina él mismo, como la de `robots.txt`, y la raíz de H9 lleva la suya | `internal/httpx/nombre.go:34-39`, `internal/httpx/robots.go:180`; `ls internal/evals/testdata/boe.legislacion-consolidada/` (1 fichero de `robots`) |
| V21 | La activación copia cada guion de `aceptacion/` a `internal/app/testdata/script/h36-<nombre>`; el rojo primero exige que cada guion falle hoy por una aserción | `scripts/workflow/aceptacion.sh:1-40`, `:74-84` |
| V23 | Llaman a `lecturasDeLosTipos` `verificarPreceptos` (`grabaciones_test.go:1082`), `loQueLlevaALeerElRepositorio` (`:1214`) y `TestLecturasDeLosDatos` (`consultas_test.go:586`); a `loQueLlevaALeerElRepositorio`, `TestLecturasDeLosDatos` (`consultas_test.go:569`) y `probarLoQueLlevaALeer`, el subtest `lo-que-lleva-a-leer` de `TestEvalsDelRepositorio` (`conjunto_test.go:3301-3304`); `bloquesDeLosTipos` (`consultas.go:340-359`) solo lee los cómputos de los tipos; la LEC está en `data/normas.yaml` y no en `data/plazos.yaml` ni en `data/cobertura.yaml` | `grep -rn lecturasDeLosTipos\|loQueLlevaALeerElRepositorio\|bloquesDeLosTipos internal`, lectura de esas líneas y `grep -c BOE-A-2000-323` sobre los tres ficheros, esta sesión |
| V24 | Con el árbol de hoy (`git diff --name-only cd3ad148 HEAD` solo nombra `docs/`), `make build` deja `bin/kitlegal` (`Makefile:63-65`; `bin/` en `.gitignore:2`) y `git status` no cambia; `./bin/kitlegal plazos calcular <orden> --json \| sed -E …` corre y las doce órdenes de contracts/aceptacion.md §2 y §3 salen con el código de sus tablas (0 los casos; 2 los errores, salvo `error-inexistente`, 3), sin `KITLEGAL_CACHE_DIR` y con `~/.cache/kitlegal` fuera de lo que el sandbox deja escribir; la cabecera de txtar con `sed -e '1i -- <golden> --'` y `>>` a una ruta fija del repositorio escribe el sobre tal cual; la salida de error del binario lleva líneas que no son del sobre (el aviso de las skills instaladas y la traza de `slog`), que la herramienta muestra mezcladas con la salida estándar; `go run ./cmd/kitlegal` sale con 1 donde el binario sale con 2; una redirección a `"$TMPDIR/…"` se deniega («path is runtime-determined»). La sesión del run se abre con `--permission-mode acceptEdits` (`scripts/paso.sh:139`) y `autoAllowBashIfSandboxed: false` (`scripts/claude-modelo.sh:239`), y `.claude/settings.json` permite `./bin/kitlegal`, `make`, `sed`, `git` y `go`, no `sh`, `echo` ni `env`; el juez vio denegada la ejecución de `"$TMPDIR/kitlegal-main"` («Command name is runtime-determined», `gates/plan.json`) | las órdenes, esta sesión; `grep -n` en esos ficheros |

## Supuestos no verificados

S1, los tamaños de los sobres de los casos 11, 15 y 16, ya no es un supuesto: están medidos con el binario de `main`
(V19, V24).

- **S2 · La captura de la tarea `[aceptacion]`** (contracts/aceptacion.md §1) construye con `make build` y ejecuta
  `./bin/kitlegal … --json | sed … >> <guion>` desde el árbol antes de cualquier cambio de producto, que es el código de
  `cd3ad148` (V2). En esta sesión corre de principio a fin y da los códigos de las tablas (V24); las órdenes que usa
  (`make`, `./bin/kitlegal`, `sed`, `git`) son las que `.claude/settings.json` permite y la sesión del run lee. No se ha
  ejecutado dentro de `scripts/claude-modelo.sh`. Alternativas rechazadas: (a) construir y ejecutar
  `"$TMPDIR/kitlegal-main"`: el nombre de la orden se decide al ejecutarla, y la sesión del juez del plan, que se abre
  como las del run, lo denegó (V24);
  (b) `go run ./cmd/kitlegal`: corre, pero sale con 1 donde el binario sale con 2, así que no deja ver el código del
  binario de `main` (V24); (c) copiar cada sobre a mano desde lo que muestra la herramienta: la salida de error del
  binario sale mezclada con la estándar (V24), y un byte de más en un sobre congelado deja `cmp stdout` en rojo para
  siempre en la rama.
- **S3 · El BOE responde a `a130` y `a133`** en el paso `grabar_datos`; si no, es una dependencia externa inaccesible,
  causa mayor con su dossier.
- **S4 · Las redacciones grabadas dan los `vigente_desde` de FR-015** (2022-12-23, 2010-05-04 y 2001-01-08): lo dice el
  spec («Lo leído da…» del hito); lo comprueba `TestPreceptosDeLosPlazos` cuando la grabación existe.
- **S5 · Cómo parte Kong la ayuda de `--help` en líneas** no se ha comprobado: los guiones comparan la ayuda en la
  salida de `--describe`, que es JSON, y de `--help` solo el código de salida.
- **S6 · Los nombres de los umbrales de las evals** de «Controles de umbral» son los que publicó el cierre de H33
  (`specs/039-h33-cobertura-en-positivo/plan.md`); este hito no cambia ningún umbral ni ninguna clase.

## Datos externos

| Fuente | Fila en `docs/SOURCES.md` de `main` | Manifiesto | Test de grabación | Dónde queda | Qué fichero de `data/` deriva |
|---|---|---|---|---|---|
| `boe.legislacion-consolidada` | «Revisado» 2026-09-13 (V17) | `internal/evals/testdata/grabaciones.json`, entrada nueva de la Ley 1/2000 con `a130` y `a133` | `TestGrabarEvals`, `internal/evals/grabacion_test.go` (`//go:build grabacion`), con `internal/httpx` | `evidencias/h36/boe.legislacion-consolidada/` y su `evidencias/h36/manifiesto.json` (D9) | ninguno: verifican `data/plazos.yaml` (D14) |

Ninguna otra fuente: el hito no consulta PLACSP, BDNS, BORME ni ningún boletín, y no cambia `data/festivos/`.
