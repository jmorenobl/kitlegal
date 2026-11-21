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

**Coste.** `internal/cache` exporta una lista cerrada, `superficieDelContrato` (`internal/cache/superficie_test.go:59-93` en
`main`; 60-99 en la rama), y `TestSuperficieExportada` la compara en los dos sentidos con lo que el paquete declara
exportado (líneas 112-140 en `main`; la comparación, en las 133-144 de la rama):
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
fecha de modificación es de hace menos de la caducidad (`ahora().Sub(mtime) < caducidad`, con `ahora` = `time.Now`,
lo mismo que `time.Since`; gates/supuestos.md, T003), no tomado y sin tocarlo; si
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
  reinstalación. Y dos del paso 8 que no se cumple: (3) el `TomarCandado` de B ejecuta antes, entera, una invocación de
  un binario mayor, que reinstala y suelta, y (4) antes de que B lo tome se edita un fichero del global. En los dos, B
  toma el candado, da la línea de hoy de ahora, no escribe nada y deja el disco como estaba.
- `TestReinstalacionFalloAMitad` (FR-011, FR-019, US4.5): para cada operación n de una reinstalación con las cuatro
  fases (una retirada intacta, un fichero que se reescribe, un enlace de host nuevo y el manifiesto), un `Escritor`
  que falla en la operación n: la línea de hoy del estado anterior, `liberar` llamado, y después `Planificar`+`Aplicar`
  con el pedido de `install -g` sin conflictos y con el estado final de la reinstalación sin fallos. Si la operación n
  es el enlace, `Aplicar` no se para y sigue con su copia (FR 024 de H19): la reinstalación sale bien y da la línea de
  FR-006, como `install`. Es el patrón de `TestFalloAMitadSeCompleta` (`aplicar_test.go:341-393`). Con una salvedad,
  la de `install`: si la operación n deja a medias la retirada de la skill de nombre retirado, `install -g` la conserva
  «sin retirar» con lo que queda (FR-027 de H32), y el test lo espera así (gates/supuestos.md, T006).

En `internal/cache/candado_test.go`, `TestTomarCandado` (FR-010) con `t.Setenv(KITLEGAL_CACHE_DIR)` y sin
`t.Parallel()`, como `TestAdaptadorConVariableDeEntorno` (`internal/cache/adaptador_test.go:488-496`): libre →
tomado y el fichero existe; `liberar` → no existe; reciente (`os.Chtimes` a hace 9 min 30 s) → no tomado, intacto;
huérfano (a hace 10 min y a hace un año) → tomado; la frontera exacta, con un reloj parado en el `tomarCandado` sin
exportar (a un nanosegundo de los 10 minutos, no tomado; a los 10 justos, tomado; gates/supuestos.md, T003); sin
directorio → creado con 0o700 y tomado; `KITLEGAL_CACHE_DIR` vacía o que es un fichero, un directorio que no se puede
crear, una creación que falla por otra causa → no tomado y nada creado; un huérfano que no se puede retirar → no
tomado y queda. `paralleltest` no pide `t.Parallel()` a una prueba que llama a
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
que el kernel tomó en H19 para la escritura del aviso (`avisar`, `internal/app/main.go:430-446` en `main`, 469-485 en la rama; research.md D5 de
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

## Reparaciones del cierre

La medición 1 del cierre (cabeza `3473dd4d`, `gates/cierre.json`, 2026-11-21; ejecución 44928879479) sale en rojo por
una sola comprobación, `evals (normativa)` (trabajo 136921530208): `ci`, `snapshot`, la web y los otros tres jobs de
evals pasan. La reparación la decide `reparar_cierre` con el informe del job (`gates/evals/normativa.json`); no toca
ninguna eval, la rúbrica, los casos, la medida, ningún umbral ni lo grabado (ADR 0029). **Se aparta de «Fuera de
alcance» del spec («Cambiar ninguna `SKILL.md`, `data/`, evals o la web») y de lo que decían plan.md («Constitution
Check», principio VIII) y tasks.md: cambian dos referencias de dos skills**, y queda registrada como lo que es (ADR
0030). Su efecto no se ha medido: este paso no abre sesiones con modelo (ADR 0032) y lo mide el job de evals de la
medición siguiente.

**R1. Dos referencias dejan de dar pie a decir lo que no se ha leído.**

**Evidencia.** El veredicto `fallo` tiene un solo motivo: `afirma_lo_no_leido:claude-sonnet-5-5`, 4 de 120 (3,3 %; el
tope es 2,8 %, como mucho 3 de 120), 1 de 60 en el modo orden y 3 de 60 en el modo herramienta, las 4 en evals por
materia. Los demás umbrales que deciden se cumplen, ninguna respuesta queda sin juzgar ni ninguna sesión sin medir, y
el juez cumple su medida (0 de 212 y 0 de 47). El hito no cambia `skills/` (`git diff main...HEAD -- skills/`, vacío):
el árbol de `normativa` es el del cierre de H41.4 (0 de 120) y el de H33 (2 de 120). Las 4 respuestas, con los tres
votos del juez, son de dos formas:

- **La frase que la referencia daba hecha.** `17-ltaibg-plazo-por-materia-claude-sonnet-5-5-01` (orden) y
  `17-ltaibg-plazo-por-materia-herramienta-claude-sonnet-5-5-02` leen solo el art. 20 de la Ley 19/2013 y dicen ante
  quién se presenta la reclamación del art. 24, con «(art. 24.6)». Las dos activan `plazos-y-territorio` —la eval lo
  admite desde H33 (`tambien_la_resuelven`)—, y la frase es la de su referencia `administrativo.md`, «La reclamación
  del art. 24.2 LTAIBG», que mandaba señalar la variación autonómica dando ella misma lo que prevé el art. 24.6, sin
  pedir su lectura: el modelo la traslada, y el texto del art. 24 no está en la sesión. No es un caso suelto: la misma
  frase marcó `17-ltaibg-plazo-por-materia-claude-sonnet-5-5-02` en el cierre de H33
  (`specs/039-h33-cobertura-en-positivo/gates/evals/normativa.json`). Son 3 marcas en dos cierres seguidos, y en este,
  2 de las 6 respuestas de la eval 17 con el modelo que decide.
- **La lectura fallida, contada dentro de una materia o retirando lo dicho.**
  `14-trlrhl-impuestos-por-materia-herramienta-claude-sonnet-5-5-02` pide el bloque `a78` del TRLRHL, que sale con `5`
  (fuera de lo grabado) en la lectura de varios bloques y en dos lecturas sueltas, y lo dice dentro de la viñeta de
  uno de los impuestos del art. 59 («la lectura de su artículo 78 falló porque la fuente limitó la consulta»): el juez
  lee ahí de qué trata el art. 78.
  `13-lrbrl-atribuciones-por-materia-herramienta-claude-sonnet-5-5-01` pide `a123` de la Ley 7/1985, que sale con `5`,
  y retira lo que había escrito antes repitiéndolo («Antes mencioné que el artículo 123 también atribuye…»).
  `lecturas_que_fallan.md` ya pedía decir qué bloques faltan, por su número, y por qué, sin decir de qué tratan, pero
  no decía dónde va esa frase ni qué se hace con lo escrito antes del fallo.

**Decisión.** (1) `skills/plazos-y-territorio/references/administrativo.md`, «La reclamación del art. 24.2 LTAIBG»:
la referencia deja de dar ante quién se reclama. Pide señalarlo como variación autonómica solo con el art. 24 LTAIBG
`[BOE-A-2013-12887, bloque a24]` leído en la conversación (paso 10 o paso 6) y con su cita; sin su texto —también en
el plazo para resolver, cuyo precepto solo remite a la reclamación—, ni ante quién se reclama ni qué prevé el art. 24:
o se lee antes de escribir, o la reclamación va con las palabras del precepto leído. El bloque `a24` está entre las
`lecturas` de la eval 17, que entran en la caché preparada de sus sesiones (`internal/evals/consultas.go`, punto 4).
Cuatro líneas donde había dos: la referencia queda en 100, el tope sin índice. (2) `skills/normativa/references/lecturas_que_fallan.md`, «Qué dice la respuesta»,
gana un párrafo: lo no leído se dice en una frase aparte, al final, con el número y la causa y nada más («No he podido
leer el artículo N de <norma>: <causa>.»), nunca en la viñeta ni en la frase de una materia ni colgada de ella, y lo
escrito de ese bloque antes de que la lectura fallara no se retira con sus palabras, porque retirarlo así lo repite.
Los ejemplos van con marcadores, sin contenido legal. (3) Los dos cambios se declaran donde el repositorio controla
lo que cada skill del producto cambia respecto a su base del laboratorio: `declaracionesDelProducto`
(`internal/evals/laboratorio_familias_test.go`) gana un `cambioEnUnaReferencia` en `normativa` y otro en
`plazos-y-territorio`, la vía que abrió H33 (su contracts/controles.md §3); sin ellos, `TestSkillsDelProductoDesdeControl`
da las dos referencias por distintas de su base. Las variantes de `laboratorio/` no cambian. Ningún `SKILL.md` cambia,
ni ninguna forma fija `⚠`, eval, rúbrica, caso, medida, umbral ni grabación, y ningún otro test.

**Por qué se hace aun así.** El umbral que no se cumple es de las skills, no del hito ni de la ejecución, y un umbral
del job que no se cumple se arregla en la skill, no rebajándolo ni dejándolo solo publicado (ADR 0029). Las dos formas
tienen su causa en un texto de las skills: una referencia que da contenido legal para trasladarlo sin lectura va
contra la regla 2 de la propia skill («el texto de un artículo, solo de `kitlegal boe articulo`, con su cita») y
contra «nunca inventar contenido legal»; y la referencia de las lecturas que fallan no decía dónde va su aviso.

**Alternativas descartadas.**

- *No tocar nada y volver a medir*: la primera forma viene del texto de la referencia y lleva dos cierres seguidos;
  repetir la medida sin cambiarlo es esperar a que la tasa baje sola.
- *Grabar `a78` y `a123` en las `lecturas` de las evals 13 y 14*, como H31.1 hizo con otros bloques (ADR 0043, clase
  b): cambia las evals y lo grabado, que este paso no toca, y la skill tiene que contar bien una lectura que falla,
  llegue o no una eval a ella. Queda a la vista de la persona: `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`
  da 10 de 66 en este informe, y solo se publica.
- *Dejar en la referencia lo que prevé el art. 24.6 y pedir además su lectura*: con el contenido delante, el modelo
  lo traslada aunque no lea; con el texto leído no lo necesita.
- *Llevar la regla al cuerpo de `normativa`* (paso 6): el caso lo trae una lectura fallida, y lo que solo ese caso
  necesita va en la referencia que se manda leer entonces (docs/SKILLS.md; ADR 0052, punto 4); el cuerpo, además,
  es lo que el spec nombra en «Fuera de alcance».
- *Un índice en `administrativo.md`* para pasar de 100 líneas: no hace falta, el texto cabe.

**Lo comprobado en este paso, sin modelo.** `make skills-check` y `make ci`, en verde con el cambio. Que las dos
formas dejen de darse no se ha medido aquí: lo mide el job de evals de la medición siguiente, en los dos modos.

## Datos externos

Ninguno. El hito no consulta ni graba ninguna fuente, no escribe en `data/` ni en `evidencias/` y no toca
`docs/SOURCES.md`; el paso `grabar_datos` no tiene nada que grabar.

## Supuestos no verificados

- **El `touch` del runner de CI.** El de esta máquina admite `-t <STAMP>` con `[[CC]YY]MMDDhhmm[.ss]` (`touch
  --help`); el de `ubuntu-latest` no se ha visto aquí. `-t` con ese formato es de POSIX.
- **Windows.** Más allá de `O_EXCL` → `CREATE_NEW` (D4), nada de la reinstalación se ha comprobado en Windows; FR-017
  lo declara sin probar en `CHANGELOG.md`.
- **Los escenarios de `quickstart.md`** describen el binario entregado: no se ejecutaron en la sesión del plan (ver
  «Lo medido en esta sesión») y con el código de `main` fallan, porque aquel binario solo da la línea de H19. Los de su
  §3 los ejecutó el barrido global con los binarios de la rama, sin un host que enlazar, y dieron lo esperado
  (quickstart.md, al principio).
