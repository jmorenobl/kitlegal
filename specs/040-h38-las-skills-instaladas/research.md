# Research: H38 · Las skills instaladas se actualizan con el binario

Decisiones técnicas del plan, cada una con lo elegido, por qué y lo descartado. Lo que dice cada una sobre una
herramienta o una dependencia externa se comprobó en local, y la decisión lo cita; lo que no se pudo comprobar sin red
está en «Supuestos no verificados». Lo que sí se ejecutó en esta sesión, y nada más, está en «Lo medido en esta sesión».

## Lo medido en esta sesión

- `go version`: `go1.27.4 linux/amd64`.
- `go test ./internal/core/instalacion -run 'TestMismaVersion|TestFormaSemVer' -count=1`: 75 pruebas en verde (la
  regla de hoy, `MismaVersion`, y la forma, `FormaSemVer`, de las que parte D6).
- `go tool -modfile=tools/golangci-lint/go.mod golangci-lint --version`: `2.13.2`.
- Las tres líneas, escritas en ficheros y contadas con `wc -c` (con su salto): la de FR-006 con `v0.2.0` y `v0.1.0`,
  119 bytes; la misma con `v0.10.1` y `v0.9.0`, 120; la de FR-009 con `v0.2.0` y `v0.1.0`, 122; la de hoy con ` -g`,
  123.
- `du -sb skills`: 81 433 bytes en 18 ficheros de 4 skills (`find skills -type f`).
- `go build -o /tmp/claude/h38/kitlegal ./cmd/kitlegal` terminó bien sobre un directorio que no existía, y lo creó
  (25,9 MB): es lo que `quickstart.md` §3 da por hecho con `-o "$t/v1/kitlegal"`.

No se ejecutó ningún binario construido: el sandbox de esta sesión pide aprobación para lanzarlos y la sesión no la
pide. Por eso los escenarios de `quickstart.md` están escritos contra lo que el hito entrega y no se han visto
terminar aquí (ver «Supuestos no verificados»).

## D1 · La decisión y la reinstalación viven en el dominio `instalacion`

**Decisión.** Un tipo nuevo del dominio, `instalacion.Reinstalacion`, con los puertos que ya existen (`Disco`,
`Enlazador`, `Escritor`) y uno nuevo, `TomarCandado`, decide y, si puede, reinstala con un solo método,
`Avisar(permiso)`, que devuelve la línea que la invocación escribe (contracts/aviso-y-reinstalacion.md §2). Reutiliza
`Aviso` (la línea de hoy), `Planificar` (el plan y los conflictos de `install`) y `Aplicar` (las escrituras), así que
el árbol que deja es el de `install -g` por construcción (FR-005).

**Por qué.** El dominio ya tiene las tres piezas y un disco en memoria para probarlas
(`internal/core/instalacion/discoEnMemoria_test.go`). Con un puerto del candado, la concurrencia de FR-010 se prueba
con intercalados fijados por el test, sin procesos ni esperas (D7). La regla R1 sigue: el dominio no hace entrada y
salida; las hacen `internal/disco` y `internal/cache`.

**Descartado.** Orquestarlo en `internal/app/aviso.go` con el disco real: los tests necesitarían un `HOME` temporal y
procesos de verdad para la concurrencia, que no se pueden intercalar de forma determinista. Un verbo o una orden nueva:
FR-016 deja la CLI igual.

## D2 · El permiso lo decide el kernel en el único punto del aviso

**Decisión.** `resolverApplet` (`internal/app/main.go`), en el mismo sitio en que hoy llama a `avisar` (tras
`cli.Analizar` y antes de atender la decisión), calcula el permiso de la invocación y se lo pasa al avisador:
`SoloAvisar` si el análisis falló, si la decisión es `DecisionDescribir` o si los argumentos del verbo son un
`servidor` (`mcp serve`, con o sin `--dry-run`); `Ensayar` si la decisión es ejecutar con `--dry-run`; `Reinstalar`
en cualquier otra ejecución. La ayuda y `skills` siguen sin buscar el aviso (`buscaElAviso`, sin cambios), y las
llamadas de herramienta siguen con el registro sin avisador (`Registro.sinAvisador`, `internal/app/registro.go`).

**Por qué.** Es el punto donde el kernel ya sabe todo lo que FR-001 pide (el análisis, la decisión, `--dry-run` y el
verbo), y es el punto en que sale hoy la línea (FR-008). `mcp serve` se reconoce por la interfaz `servidor`, como ya lo
reconoce `ejecutarVerbo` (`main.go`, «El verbo que sirve no tiene plazo»), sin escribir su nombre en otro sitio.

**Descartado.** Que el avisador mire `os.Args`: repetiría el análisis y no sabría si falló. Comparar el applet y el
verbo con los literales `mcp` y `serve`: un segundo sitio con el nombre del servidor.

## D3 · Dos evaluaciones: una antes del candado y otra bajo él

**Decisión.** Con `Reinstalar`, el dominio evalúa FR-003 (a), (b) y (c) sobre el disco; solo si se cumplen toma el
candado; con el candado tomado vuelve a evaluar sobre el disco de ese momento, y es esa segunda evaluación la que
decide: sin línea de hoy, ninguna línea (otra invocación ya reinstaló); con algo que ya no se cumple, la línea de hoy
de ese estado; si se cumple, aplica su plan.

**Por qué.** FR-010 pide las dos cosas: el candado «se intenta tomar solo cuando se cumplen FR-003 (a), (b) y (c)», y
de dos invocaciones a la vez «como mucho una reinstala; las demás dan la línea de hoy o, si la miran cuando ya terminó,
ninguna». Sin la segunda evaluación, la invocación que evaluó antes de que la otra escribiera y toma el candado
cuando la otra ya lo soltó aplicaría un plan del estado viejo y daría también la línea de FR-006.

**Descartado.** Tomar el candado primero: escribiría el candado y crearía el directorio de la caché en invocaciones que
no reinstalan, contra FR-010 y FR-013. Una sola evaluación: el caso de arriba (SC-004).

## D4 · El candado, en `internal/cache`

**Decisión.** `cache.TomarCandado(nombre string, caducidad time.Duration) (liberar func(), tomado bool)`, en un fichero
nuevo `internal/cache/candado.go`: resuelve el directorio con `Directorio()`, lo crea con `os.MkdirAll` y el mismo
`permisosDelDirectorio` (0o700) con que la caché crea el suyo (`internal/cache/abrir.go:26` y `:55`), y crea
`<directorio>/<nombre>` con `os.OpenFile(…, O_WRONLY|O_CREATE|O_EXCL, 0o600)`, vacío. `liberar` lo borra. El nombre,
`skills-global.lock`, y la caducidad, 10 minutos, son constantes de la composición en `internal/app/aviso.go`, que
pasa al dominio la función que lo toma (contracts/aviso-y-reinstalacion.md §4).

**Por qué.** FR-010 fija el directorio (`cache.Directorio()`) y los permisos (los de la caché), y los dos viven en
`internal/cache`: ahí no se duplica ninguno. `O_EXCL` con `O_CREATE` es la creación exclusiva («used with O_CREATE,
file must not exist», `go doc os.O_EXCL`), también en Windows, donde `syscall.Open` la traduce a `CREATE_NEW`
(`$(go env GOROOT)/src/syscall/syscall_windows.go:436-437`). `os.MkdirAll` aplica los permisos antes de la umask a lo
que crea y no hace nada si ya existe (`go doc os.MkdirAll`).

**Coste.** `internal/cache` exporta una lista cerrada, `superficieDelContrato` (`internal/cache/superficie_test.go:59-93`),
y `TestSuperficieExportada` la compara en los dos sentidos con lo que el paquete declara exportado (líneas 112-140):
`cache.TomarCandado` la pondría en rojo. La lista gana la entrada `"func TomarCandado"` y su comentario nombra el
candado de H38 (contracts/aviso-y-reinstalacion.md §4), como ganó `"func Directorio"` en H7
(specs/010-h7-internal-graph-grafo/plan.md:494). La función no recibe ni devuelve nada de `database/sql` ni de SQLite
y no ejecuta ninguna sentencia, así que no abre lo que ese test cierra (FR-005 de H3).

**Descartado.** En `internal/disco`: tendría que llamar igual a `cache.Directorio()` y copiar 0o700. En `internal/app`,
para no ampliar esa superficie: lo mismo, porque `permisosDelDirectorio` no se exporta (`internal/cache/abrir.go:26`),
y exportarlo sería otra entrada de la misma lista. Un cierre del
sistema (`flock`): `syscall.Flock` existe en Linux (`go doc syscall.Flock`) y no en Windows (`GOOS=windows go doc
syscall.Flock` y `… syscall.LockFileEx`: «no symbol»), y el spec fija un candado por fecha de modificación. Un
candado en `~/.agents/skills`: lo prohíben FR-010 y FR-016.

## D5 · Cuándo el candado no se toma, y la sustitución del huérfano

**Decisión.** `TomarCandado` devuelve `tomado == false`, sin haber creado nada que quede, si `Directorio()` falla, si
`MkdirAll` falla, o si la creación exclusiva falla por otra causa que «ya existe». Si ya existe: `os.Lstat`; si su
fecha de modificación es de hace menos de la caducidad (`time.Since(mtime) < caducidad`), no tomado y sin tocarlo; si
es de hace la caducidad o más, `os.Remove` y una sola creación exclusiva más; si la retirada o esa creación fallan, no
tomado. Nunca espera ni reintenta más (fuera de alcance del spec).

**Por qué.** Es FR-010 frase a frase, con la operación más simple para cada una.

**Lo que no se diseña, y por qué.** Comprobar la fecha y retirar el fichero no son una sola operación del sistema de
ficheros: dos invocaciones que encuentran el mismo huérfano en el mismo instante pueden retirarlo y crearlo las dos
(la segunda retira el que la primera acaba de crear). Con la exclusión que da el sistema —un cierre de fichero— no
haría falta caducidad, y es lo que D4 descarta. El caso pide un huérfano, que solo deja una invocación interrumpida a
mitad de su reinstalación, y dos invocaciones que lo examinan en la misma ventana de microsegundos; y entonces las dos
evalúan bajo el candado el mismo estado y aplican el mismo plan, con escrituras atómicas de los mismos bytes
(`Escritor.EscribirFichero`, `internal/core/instalacion/puertos.go`), así que el árbol queda como con una. No se
añade una ficha de sustitución ni un segundo candado: cada uno pide a su vez una caducidad y repite el problema
(constitución V). SC-004 se comprueba con dos invocaciones a la vez sin huérfano, que es lo que dice.

## D6 · La precedencia SemVer, propia y sin dependencia

**Decisión.** Una función no exportada de `internal/core/instalacion/version.go`, junto a `FormaSemVer` y
`MismaVersion`, que compara dos versiones con forma SemVer en precedencia SemVer 2.0.0 §11: sin la `v` inicial y sin
los metadatos de construcción; el núcleo, número a número; una versión sin prerelease va después de la misma con
prerelease; dos prereleases, identificador a identificador: los numéricos por su valor, los alfanuméricos por sus
bytes ASCII, un numérico antes que un alfanumérico y, si todos los anteriores son iguales, la que tiene más
identificadores va después. Un número se compara por su longitud y, con la misma, por sus bytes: la gramática no
admite ceros a la izquierda en el núcleo ni en un identificador numérico (`version.go`, `numeroSemVer`), así que es
exacto y no desborda con números de cualquier longitud. Quien la llama ha comprobado antes la forma con `FormaSemVer`
(FR-004: sin forma, no reinstala).

**Por qué.** FR-004 pide exactamente eso, con la `v` opcional, como `FormaSemVer` (FR 073 de H19).

**Descartado.** `golang.org/x/mod/semver`: no está entre las dependencias fijadas (constitución V; `docs/ROADMAP.md`
§3; `go.mod` no la nombra) y además exige la `v` inicial y admite `vMAJOR` y `vMAJOR.MINOR` como abreviaturas
(`~/go/pkg/mod/golang.org/x/mod@v0.42.0/semver/semver.go:5-22`), que no son la gramática de `FormaSemVer`. `go/version`
de la biblioteca estándar ordena versiones de la cadena de herramientas de Go («go1.21.0», `go doc go/version`), no
SemVer.

## D7 · Los unitarios: por la API pública, con puertos falsos y sin esperas

**Decisión.** En `internal/core/instalacion/reinstalacion_test.go` (paquete `instalacion_test`, como los demás del
dominio), sobre `discoEnMemoria`:

- `TestReglaDeVersiones` (FR-004, FR-019): tabla de pares (versiones instaladas, versión del binario) → la línea de
  FR-009 o la de hoy, con `Avisar(Ensayar)`, que no escribe ni toma el candado. Filas: menor (`v0.1.0` → `v0.2.0`),
  prerelease anterior a su final (`0.2.0-rc.1` → `0.2.0`) y al revés, `1.0.0-alpha` < `1.0.0-alpha.1`, numérico
  contra numérico (`1.0.0-2` < `1.0.0-11`), numérico antes que alfanumérico (`1.0.0-1` < `1.0.0-alpha`), núcleo por
  valor y no por bytes (`0.9.0` < `0.10.0`), solo metadatos de construcción (`0.1.0+abc` y `0.1.0`: la de hoy), un
  binario más viejo, una skill declarada más nueva que el binario con el manifiesto más viejo, y una versión
  instalada sin forma.
- `TestReinstalacion` (FR-003, FR-005, FR-006, FR-007, FR-014; principio III): los estados del spec sobre un ámbito
  global en memoria —sin tocar, un fichero que falta, uno editado, uno ajeno, una entrada de host ajena, una retirada
  intacta y una «sin retirar», un manifiesto local viejo, uno al día con el global viejo y un local inservible
  (`./.agents/skills/kitlegal.json` ilegible) con el global viejo— con `Avisar(Reinstalar)`: la línea, el estado del
  disco (`instantanea`) igual al de `Planificar`+`Aplicar` del mismo pedido o sin cambios, y `liberar` llamado una vez
  por cada candado tomado. Con el local inservible, ninguna línea, el disco sin cambios y el candado sin pedir: el
  local tapa al global (FR-003 a, «Edge Cases»).
- `TestReinstalacionConcurrente` (FR-010, SC-004): dos `Reinstalacion` sobre el mismo disco en memoria con un candado
  falso compartido, y los dos intercalados que importan forzados desde los propios puertos, sin gorrutinas: (1) la
  primera escritura de A ejecuta B entera, que encuentra el candado tomado y da la línea de hoy, y A da la de FR-006;
  (2) el `TomarCandado` de B, ya evaluado el estado viejo, ejecuta antes A entera, que reinstala y suelta; B toma el
  candado, vuelve a evaluar y no da ninguna línea. En los dos, una sola línea de FR-006 y el estado final el de una
  reinstalación.
- `TestReinstalacionFalloAMitad` (FR-011, FR-019, US4.5): para cada operación n de una reinstalación con las cuatro
  fases (una retirada intacta, un fichero que se reescribe, un enlace de host nuevo y el manifiesto), un `Escritor`
  que falla en la operación n: la línea de hoy del estado anterior, `liberar` llamado, y después `Planificar`+`Aplicar`
  con el pedido de `install -g` sin conflictos y con el estado final de la reinstalación sin fallos. Si la operación n
  es el enlace, `Aplicar` no se para y sigue con su copia (FR 024 de H19): la reinstalación sale bien y da la línea de
  FR-006, como `install`. Es el patrón de `TestFalloAMitadSeCompleta` (`aplicar_test.go:341-393`).

En `internal/cache/candado_test.go`, `TestTomarCandado` (FR-010) con `t.Setenv(KITLEGAL_CACHE_DIR)` y sin
`t.Parallel()`, como `TestAdaptadorConVariableDeEntorno` (`internal/cache/adaptador_test.go:488-496`): libre →
tomado y el fichero existe; `liberar` → no existe; reciente (`os.Chtimes` a hace 9 min 30 s) → no tomado, intacto;
huérfano (a hace 10 min y a hace un año) → tomado; sin directorio → creado con 0o700 y tomado; `KITLEGAL_CACHE_DIR`
vacía o que es un fichero → no tomado y nada creado. `paralleltest` no pide `t.Parallel()` a una prueba que llama a
`Setenv` (`~/go/pkg/mod/github.com/kunwardeep/paralleltest@v1.0.15/pkg/paralleltest/paralleltest.go:87,150,178,337`;
la versión es la de `tools/golangci-lint/go.sum:397`). `os.Chtimes` cambia la fecha de modificación (`go doc
os.Chtimes`).

En `internal/app/aviso_test.go`, la tabla de `TestAvisoDelKernel` (`casosDelAvisoDelKernel`) gana el permiso esperado
de cada caso que avisa, que el avisador sintético registra (FR-001, FR-002): ejecutar → `Reinstalar`, `--dry-run` →
`Ensayar`, `--describe` y los errores de argumentos → `SoloAvisar`.

**Por qué.** Lo observable de FR-004 es si reinstala, y la API pública lo dice sin exportar el comparador. Los
intercalados con puertos falsos son deterministas y fallan con un mutante sin candado o sin la segunda evaluación; un
test con procesos y esperas no garantiza que se solapen (golang-concurrency: «Should this be synchronous instead?»;
todo lo nuevo es síncrono, sin gorrutinas).

**Descartado.** Exportar el comparador para probarlo: un símbolo público solo para un test. Un test de concurrencia
con gorrutinas y `time.Sleep`: frágil y no prueba el solape.

## D8 · La concurrencia en e2e, como aceptación y no como control

**Decisión.** El escenario 6 de US4 se escribe con dos órdenes en segundo plano con nombre, `exec … &a&` y `exec …
&b&`, y `wait a` y `wait b`, que dejan en `stdout` y `stderr` las de cada una (go-internal v1.16.0:
`testscript/doc.go:160-170` y `:240-249`; `testscript/cmd.go:571-578`). El guion exige exactamente una línea de FR-006
entre las dos, que la otra sea la de hoy o ninguna y el árbol de una reinstalación.

**Por qué.** Es la aceptación literal de US4.6. Que se solapen no está garantizado, así que el control de SC-004 que
falla con un mutante es `TestReinstalacionConcurrente` (D7), y el guion comprueba que en procesos reales no sale nada
más que eso.

## D9 · Cómo se construyen en e2e los estados del spec

**Decisión.** Con los binarios del arnés de H19 (`$KITLEGAL_V1_BIN`, `v0.1.0`; `$KITLEGAL_V2_BIN`, `v0.2.0`; mismas
skills empotradas, `internal/app/e2e_test.go:125-126`) y lo que ya usan los guiones de H19: `sed` dentro de `exec sh -c`
para los metadatos de construcción (`h19-skills-aviso.txtar:157`), un `olvida.sh` en el propio archivo del guion para
el fichero ajeno (`h19-skills-conflictos-dentro.txtar:171`), `symlink` a otro sitio para la entrada de host ajena,
`exec chmod -R a-w` y `exec chmod -R u+w` para el permiso (`h19-instalador-rechazos.txtar:117`), `exec touch -t
202001010000` para el huérfano, `cp` para el candado reciente y `arbol` para comparar árboles (ruta, tipo, permisos y
huella, `e2e_test.go:928-1060`). La skill de nombre retirado sale de una instalación v1: el directorio `normativa`
pasa a `boe-legislacion` y, en `kitlegal.json`, cada `"normativa` a `"boe-legislacion` (la comilla deja fuera
`jerarquia_normativa.md`); los bytes no cambian, así que sus huellas siguen siendo las declaradas e «intacta»; con una
línea añadida a su `SKILL.md`, «sin retirar». El detalle por guion está en contracts/aceptacion.md.

`touch -t [[CC]YY]MMDDhhmm[.ss]` y `chmod -R` están comprobados con `touch --help` y `chmod --help` en esta máquina.

## D10 · La orden de las consultas en e2e

**Decisión.** `territorio resolver Leganés --json --no-graph`, salvo el caso del exit 4, que es `boe articulo
BOE-A-2015-10565 a99 --offline` (como `h19-skills-aviso.txtar`). Las referencias «sin skills instaladas» se toman al
principio de cada guion, con la misma orden y un `HOME` vacío.

**Por qué.** La salida estándar de `territorio` no depende del reloj (su `fecha_consulta` es la de sus datos, como
dicen los guiones de H19), y con `--no-graph` la orden no escribe en la caché: `territorio` no la usa y `--no-graph`
elige el almacén nulo (`internal/app/main.go:202-207`). Así «el directorio de la caché queda como lo deja la misma
orden sin skills instaladas» (FR-013) se comprueba con `! exists $KITLEGAL_CACHE_DIR` donde no reinstala, y el
directorio que crea el candado se ve solo (FR-010).

## D11 · Un fallo del candado o de la reinstalación no se propaga

**Decisión.** Ni `TomarCandado` ni `Avisar` devuelven error: lo que falla se queda en la línea de hoy, que es lo que
FR-010, FR-011 y FR-012 piden («un fallo de la reinstalación nunca es un error de la invocación»), y el borrado del
candado que no se puede borrar se deja como está (lo sustituye la siguiente tras la caducidad). Es la misma decisión
que el kernel tomó en H19 para la escritura del aviso (`avisar`, `internal/app/main.go:430-446`; research.md D5 de
H19), y no un error silenciado: el resultado observable es la línea.

**Descartado.** Devolver el error y registrarlo con `slog`: el avisador no tiene registrador y la línea ya dice lo
que pasó; cambiar el código de salida: lo prohíbe FR-012.

## D12 · La firma del avisador

**Decisión.** `type Avisador func(permiso instalacion.PermisoDelAviso) (linea string, hay bool)`, en
`internal/app/aviso.go`. `AvisoDeVersion` compone `instalacion.Reinstalacion` con `disco.Lector{}`,
`dependencias.Enlazador`, `disco.NuevoEscritor(dependencias.Enlazador)`, el candado de D4, `HOME` del entorno en cada
llamada, la versión, lo empotrado y `skillsRetiradas` (`internal/app/retiradas.go`): lo mismo con lo que instala
`install` (`instalacion.go:270-293`), así que reinstala lo que instalaría. Sin `Enlazador` en las dependencias —un
defecto de composición que el applet ya da como inesperado (`errSinEnlazador`)— el avisador solo da la línea de hoy,
para no entrar en pánico en una ruta de usuario (constitución IV).

**Descartado.** Un segundo registro en el `Registro` («reinstalador») al lado del avisador: dos funciones que deciden
la misma línea.

## D13 · Skills de Go aplicadas

`golang-how-to` orquesta; se aplicaron `golang-concurrency` (todo síncrono; candado entre procesos con doble
comprobación; intercalados deterministas sin esperas; `-race` en `make test`), y por su tabla de enrutado
`golang-testing` y `golang-stretchr-testify` (tablas, `t.Setenv` sin `t.Parallel`), `golang-error-handling` (D11),
`golang-safety` (el `Enlazador` nulo), `golang-naming` (`PermisoDelAviso`, `TomarCandado`, sin colisión con el `Modo`
de `manifiesto.go:62`), `golang-project-layout` (ningún paquete nuevo) y `golang-lint` (el switch del permiso
nombra sus tres valores: `exhaustive` v0.12.0 comprueba los switch sobre tipos de constantes «enum-like»,
`~/go/pkg/mod/github.com/nishanths/exhaustive@v0.12.0/doc.go:3-10`, versión de `tools/golangci-lint/go.sum:458`, con
`default-signifies-exhaustive: true` en `.golangci.yml:257`; permisos 0o600 y 0o700, y `filepath.Clean` en la ruta
del candado, como `internal/disco/escribir.go`).

## Datos externos

Ninguno. El hito no consulta ni graba ninguna fuente, no escribe en `data/` ni en `evidencias/` y no toca
`docs/SOURCES.md`; el paso `grabar_datos` no tiene nada que grabar.

## Supuestos no verificados

- **El `touch` del runner de CI.** El de esta máquina admite `-t <STAMP>` con `[[CC]YY]MMDDhhmm[.ss]` (`touch
  --help`); el de `ubuntu-latest` no se ha visto aquí. `-t` con ese formato es de POSIX.
- **Windows.** Más allá de `O_EXCL` → `CREATE_NEW` (D4), nada de la reinstalación se ha comprobado en Windows; FR-017
  lo declara sin probar en `CHANGELOG.md`.
- **Los escenarios de `quickstart.md`** describen el binario entregado: no se han ejecutado en esta sesión (ver «Lo
  medido en esta sesión») y con el código de hoy fallan, porque el binario de hoy solo da la línea de H19.
