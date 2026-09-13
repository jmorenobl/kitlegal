# Quickstart: validación de H4

Guía **ejecutable** para comprobar que H4 entrega lo que dice. Cada escenario se ejecuta tal cual, desde la raíz del
repositorio, sobre la rama `h4-applet-boe-puerto` **con el hito ya implementado y confirmado**: no es la implementación,
es cómo se verifica. Los nombres de test son los del inventario de [plan.md](./plan.md) («Inventario de tests»).

**Sin efectos colaterales.** Ningún escenario crea ni modifica un fichero versionado, toca el índice o el historial de
git, ni escribe en la caché real de la cuenta (`~/.cache/kitlegal/`). Lo que hay que construir o romper a propósito se
hace en `/tmp/kitlegal-quickstart-h4/`: un binario, un enlace y un **clon desechable** del repositorio (`git clone`, que
solo lee el repositorio), que los prerrequisitos crean vacía y la limpieza borra al final. Los prerrequisitos empiezan
borrando esa carpeta por si una ejecución interrumpida la dejó con contenido (un reintento de la tarea de cierre o una
reanudación tras un límite de uso, antes de llegar a la limpieza): sin eso, `git clone` del escenario 7.b fallaría porque
su destino ya existe, y con él los escenarios 8, 9 y 13.b. Una ejecución retomada empieza siempre por los prerrequisitos. Lo único que puede quedar en el árbol son `coverage.out` y
`coverage-integration.out` de `make test`/`make ci`, ignorados desde H0 (`.gitignore`: `/coverage.*`). Nada se redirige a
un fichero: toda salida va a la pantalla.

**El directorio del feature queda fuera de las comprobaciones del árbol.** Cuando la guía se ejecuta dentro del workflow
`hito` (tarea de cierre), el propio workflow reescribe antes de cada tarea sus ficheros de estado versionados
(`specs/005-h4-applet-boe-puerto/gates/tarea-actual.json` y `gates/tareas-intentos.json`), que aparecen como modificados
(` M`), y la tarea de cierre escribe allí su evidencia. Por eso las órdenes de `git status` descartan
`specs/005-h4-applet-boe-puerto/` **en cualquier estado** (`^.. ` casa las dos columnas de estado de `--porcelain`), no solo
lo no seguido.

**Cada orden se ejecuta tal cual en el modo desatendido, sin sustituciones.** La tarea de cierre ejecuta esta guía en una
sesión `claude -p --permission-mode acceptEdits` (`scripts/hito.sh`, `scripts/paso.sh`). Esa sesión solo ejecuta sin pedir
aprobación lo que casa con la lista de permitidos de `.claude/settings.json`, y solo escribe dentro del repositorio. Una
orden que pide aprobación no se ejecuta, así que el escenario no mediría nada. Por eso todas las órdenes están escritas en
una forma que ese modo acepta. Cada forma se comprobó con una sonda en una sesión desatendida el 2026-09-13, y las
sondas del final de los prerrequisitos la vuelven a comprobar en cada ejecución:

| Necesidad | Forma de la guía | Forma que pide aprobación y no se usa |
|---|---|---|
| Crear, borrar o editar en `/tmp/kitlegal-quickstart-h4/` | `rtk proxy mkdir`, `rtk proxy rm -r`, `rtk proxy chmod`, `rtk proxy ln`, `rtk proxy perl -0pi` | `mkdir` y `rm` fuera del repositorio (aunque estén en la lista), `perl`, `printf … >> fichero` |
| Ejecutar el binario o el enlace de la carpeta temporal | `rtk proxy /tmp/kitlegal-quickstart-h4/kitlegal …` | `/tmp/kitlegal-quickstart-h4/kitlegal …` |
| Leer el código de salida | `rtk proxy sh -c '…; echo "código $?"'` | `…; echo "código $?"` fuera de `sh -c` |
| Pasar una variable de entorno | `rtk proxy env VARIABLE=valor …` | `VARIABLE=valor orden` y `env` sin `rtk proxy` |
| Comprobar que algo no existe | `rtk proxy test ! -e … && echo "…"` | `test` sin `rtk proxy` |
| Comparar dos estados | `… 2>&1 \| rtk proxy shasum`, dos huellas en pantalla | `shasum` sin `rtk proxy`, `> fichero` fuera del repositorio |

`go`, `git`, `make` y `echo` con texto literal están permitidos tal cual, también con rutas de la carpeta temporal
(`go build -o`, `git clone`, `make -C`, `go -C`).

## Prerrequisitos

```bash
go version                          # go1.27.1 (directiva toolchain de go.mod)
git --version
make --version
git rev-parse --abbrev-ref HEAD     # h4-applet-boe-puerto
rtk proxy git status --porcelain | rtk proxy grep -vE '^.. specs/005-h4-applet-boe-puerto/' ; echo "fin del estado"
rtk proxy mkdir -p /tmp/kitlegal-quickstart-h4
rtk proxy chmod -R u+w /tmp/kitlegal-quickstart-h4
rtk proxy rm -r /tmp/kitlegal-quickstart-h4
rtk proxy mkdir /tmp/kitlegal-quickstart-h4
rtk proxy ls -A /tmp/kitlegal-quickstart-h4 ; echo "fin de la carpeta"
```

**Esperado**: entre la orden de `git status` y «fin del estado» no aparece ninguna línea: nada sin confirmar fuera del
directorio del feature (los ficheros de estado del workflow y la evidencia de cierre, que viven en él, quedan fuera).
Entre `ls -A` y «fin de la carpeta» tampoco aparece ninguna: la carpeta temporal existe y está vacía, haya o no quedado
algo de una ejecución anterior. `mkdir -p` garantiza que exista antes de borrarla, `chmod -R u+w` evita que `rm -r` pida
confirmación en un terminal por los objetos de git de solo lectura del clon, y `mkdir` la vuelve a crear. No se usa
`rm -rf`, que la configuración del agente deniega.

### Sondas de las formas

```bash
rtk proxy sh -c 'sh -c "exit 3"; echo "código $?"'
rtk proxy env KITLEGAL_SONDA=valor printenv KITLEGAL_SONDA
rtk proxy ls -laR /tmp/kitlegal-quickstart-h4 2>&1 | rtk proxy shasum
rtk proxy mkdir /tmp/kitlegal-quickstart-h4/sonda
rtk proxy ls -laR /tmp/kitlegal-quickstart-h4 2>&1 | rtk proxy shasum
rtk proxy test ! -e /tmp/kitlegal-quickstart-h4/sonda && echo "sonda de ausencia: esta línea no debe aparecer"
rtk proxy rm -r /tmp/kitlegal-quickstart-h4/sonda
rtk proxy ls -A /tmp/kitlegal-quickstart-h4 ; echo "fin de las sondas"
```

**Esperado**, en orden:

1. `código 3`: el código de salida llega entero, no solo cero o distinto de cero.
2. `valor`: `env` entrega la variable.
3. Dos huellas **distintas**: la comparación de estados delata un cambio. Con el mismo estado, las huellas son iguales
   (escenario 1).
4. Ninguna línea entre la segunda huella y «fin de las sondas»: la comprobación de ausencia no da verde cuando el
   directorio existe, y la carpeta vuelve a quedar vacía.

Si una orden de la guía pide aprobación o no da lo esperado en una sonda, la guía no se puede ejecutar tal cual. La
tarea de cierre se detiene y lo anota: no sustituye la orden por otra a su criterio.

Los bloques se ejecutan con `bash` o `zsh`. Fuera de `go`, `git`, `make` y `rtk`, solo herramientas POSIX y `perl`
(`chmod`, `env`, `grep`, `ln`, `ls`, `mkdir`, `printenv`, `rm`, `sh`, `shasum`, `tail`, `test`). **Ninguna comprobación
necesita red**, salvo la primera ejecución con la caché de módulos fría, `make vuln` (dentro de `make ci`) y el escenario
15.b, marcado.

Varios escenarios filtran con `grep` la salida **en crudo** de `go test -v` o de `git`. Por eso, en un terminal con un
envoltorio que la resume (por ejemplo `rtk`), cada etapa de la tubería va precedida de `rtk proxy`, como aquí. `rtk proxy`
también esquiva las funciones de la shell que sustituyen `grep` por otro programa. En un terminal sin envoltorio se omite el
prefijo `rtk proxy` y el resultado esperado no cambia: `rtk proxy sh -c '…'` pasa a ser `sh -c '…'`.

---

## Escenario 1 — La suite entera sin red, con detector de carreras, sin tocar la caché real ni los datos protegidos (SC-005, SC-014)

```bash
rtk proxy ls -laR ~/.cache/kitlegal 2>&1 | rtk proxy shasum
make test
rtk proxy ls -laR ~/.cache/kitlegal 2>&1 | rtk proxy shasum
rtk proxy git status --porcelain -- internal/source/boe/testdata schemas internal/app/testdata docs/SOURCES.md ; echo "fin del estado"
```

**Esperado**:

- Las dos huellas de `shasum` son iguales: la caché de la cuenta no se ha tocado. La sonda 3 de los prerrequisitos
  demuestra que la comparación delata un cambio. Si `~/.cache/kitlegal` no existe, el listado es el error de `ls` y la
  medida detecta igualmente que aparezca.
- `make test` termina sin `FAIL` ni `WARNING: DATA RACE`, con `ok` en `internal/source/boe`, `internal/app`,
  `internal/httpx`, `internal/cli` e `internal`.
- Entre `git status` y «fin del estado» no aparece ninguna línea: ningún test escribe golden, esquemas ni grabaciones.

---

## Escenario 2 — Aceptación: el `data` de `articulo` coincide con `boe.py` en 6 artículos de 4 leyes (SC-001, FR-116)

```bash
rtk proxy go test -count=1 -v -run '^TestArticuloCoincideConBoePy$' ./internal/source/boe/ \
  | rtk proxy grep -E '^\s*--- (PASS|FAIL): TestArticuloCoincideConBoePy/'
```

**Esperado**: seis líneas `--- PASS`, una por referencia (`BOE-A-2015-10565-a21`, `BOE-A-2015-10565-a1`,
`BOE-A-1985-5392-a22`, `BOE-A-2017-12902-a1-30`, `BOE-A-2017-12902-da-3`, `BOE-A-1992-26318-a42`), y ninguna `--- FAIL`.
Seis artículos de cuatro leyes: el criterio de SC-001 (cinco de tres) se cumple con holgura (6 ≥ 5, 4 ≥ 3), y el último
es el único que compara avisos con contenido real. Cada subtest compara, campo a
campo, título, tipo, fecha de la versión, fecha de vigencia, norma modificadora, texto, avisos (lista de textos) y dirección
pública con la referencia que salió de ejecutar `refs/boe.py` —sin modificarlo y sin red, contra la misma grabación— y
que una persona revisó campo a campo, en la misma pausa en que grabó. Ejecutar este escenario no necesita Python: las
referencias están versionadas como dato.

---

## Escenario 3 — Caché: segunda consulta sin red, vigencias, `--offline`, ensayo y fallos no guardados (SC-003, SC-004, SC-012, FR-093, FR-094)

```bash
rtk proxy go test -count=1 -v \
  -run '^(TestCacheDeLosSeisVerbos|TestOfflineDeLosSeisVerbos|TestEnsayoDeLosSeisVerbos|TestFallosNoSeGuardan|TestArticulos)$' \
  ./internal/source/boe/ | rtk proxy grep -E '^--- (PASS|FAIL): '
```

**Esperado**: cinco líneas `--- PASS` (una por test) y ninguna `--- FAIL`. Lo que comprueban:

- La segunda consulta de cada verbo se resuelve con una reproducción estricta sobre un directorio vacío, con el mismo
  `data`, `url` y `fecha_consulta`.
- Pasada su vigencia (300 s o 7 días, con reloj controlado), se vuelve a pedir.
- `articulos` de tres bloques sin caché cuenta cuatro peticiones. Si el segundo no existe, cuenta tres y deja el primero
  en caché.
- Con `--offline` y con `--dry-run`, la caché queda intacta.

---

## Escenario 4 — La fecha de consulta es la de la consulta (FR-096, SC-008)

```bash
rtk proxy go test -count=1 -v -run '^TestFechaDeConsultaDeArticulo$' ./internal/source/boe/ | rtk proxy grep -E '^--- (PASS|FAIL): '
rtk proxy go test -count=1 -v -run '^TestMontadorFechaDeConsulta$' ./internal/cli/ | rtk proxy grep -E '^--- (PASS|FAIL): '
rtk proxy go test -count=1 -v -run '^TestInstanteDeEmision$' ./internal/httpx/ | rtk proxy grep -E '^\s*--- (PASS|FAIL): TestInstanteDeEmision/'
```

**Esperado**, sin ningún `FAIL`:

- `--- PASS: TestFechaDeConsultaDeArticulo`.
- `--- PASS: TestMontadorFechaDeConsulta`.
- Nueve líneas `--- PASS` de `TestInstanteDeEmision/…`: `acierto`, `acierto-tras-reintentos`, `fallo-tras-reintentos`,
  `denegada-por-robots`, `sin-turno`, `redireccion`, `reproduccion`, `ensayo` y `argumentos`.

---

## Escenario 5 — e2e: los seis verbos sobre el binario compilado, el enlace `boe`, `--offline` y los códigos (SC-005, FR-114)

```bash
rtk proxy go test -count=1 -v -run '^TestEntregaDelHito$/^boe-' ./internal/app/ \
  | rtk proxy grep -E '^\s*--- (PASS|FAIL): TestEntregaDelHito/boe-'
```

**Esperado**: cinco líneas `--- PASS`: `TestEntregaDelHito/boe-verbos`, `boe-multicall`, `boe-offline`, `boe-codigos` y
`boe-cache-rapida`. El binario lo construye el propio test en un directorio temporal y responde desde las grabaciones copiadas al
`$WORK` de cada guion: ninguna conexión.

---

## Escenario 6 — Menos de 200 ms con la entrada en caché (SC-002, FR-117)

```bash
rtk proxy go test -count=1 -v -run '^TestEntregaDelHito$/^boe-cache-rapida$' ./internal/app/ \
  | rtk proxy grep -E 'cronometra|--- (PASS|FAIL): TestEntregaDelHito/boe-cache-rapida'
```

**Esperado**: `--- PASS: TestEntregaDelHito/boe-cache-rapida`. Con `-v`, el registro del guion muestra las diez
invocaciones de `cronometra 200ms …`. El guion falla si alguna tarda 200 ms o más, o si emite una petición: la
reproducción está vacía, así que terminaría con código 1.

---

## Escenario 7 — Contratos: `schemas/` coincide con `--describe`, la salida valida y la edición a mano se detecta (SC-006, FR-110, FR-111)

### 7.a En el árbol

```bash
make schema-check
rtk proxy go test -count=1 -v -run '^(TestSalidaDeBoeContraSchemas|TestEsquemasCubrenTodosLosVerbos)$' ./internal/app/ | rtk proxy grep -E '^--- (PASS|FAIL): '
```

**Esperado**: `make schema-check` termina con `ok  	github.com/jmorenobl/kitlegal/internal/app`; las dos líneas son `--- PASS`.

### 7.b En un clon desechable

```bash
git clone --quiet . /tmp/kitlegal-quickstart-h4/copia
rtk proxy perl -0pi -e 's/"minLength": 1/"minLength": 2/' /tmp/kitlegal-quickstart-h4/copia/schemas/bloque.json
rtk proxy git -C /tmp/kitlegal-quickstart-h4/copia diff --stat
make -C /tmp/kitlegal-quickstart-h4/copia schema-check 2>&1 | rtk proxy grep -E 'schemas/bloque\.json|FAIL' ; echo "fin de la comprobación"
```

**Esperado**:

- `git diff --stat` del clon da `schemas/bloque.json | 2 +-` y `1 file changed, 1 insertion(+), 1 deletion(-)`. Es la
  sonda de que la edición se ha hecho: sin `/g`, `perl -0` cambia solo la primera aparición.
- Antes de «fin de la comprobación» aparecen una línea que nombra `schemas/bloque.json` y la parte de «articulo», y
  líneas `FAIL`. La primera aparición de `"minLength": 1` en ese fichero está en la parte de `articulo`, por el orden de
  claves.

El árbol de trabajo no cambia: la edición es en el clon. `git clone` no encuentra un destino existente, porque los
prerrequisitos han dejado vacía la carpeta temporal.

---

## Escenario 8 — Golden: un byte distinto en `data` se detecta (FR-112, US6-3)

```bash
rtk proxy perl -0pi -e 's/"texto": "/"texto": "X/' /tmp/kitlegal-quickstart-h4/copia/internal/source/boe/testdata/golden/articulo-BOE-A-2015-10565-a21.json
rtk proxy git -C /tmp/kitlegal-quickstart-h4/copia diff --stat
rtk proxy go -C /tmp/kitlegal-quickstart-h4/copia test -count=1 -run '^TestGolden$' ./internal/source/boe/ 2>&1 \
  | rtk proxy grep -E 'articulo-BOE-A-2015-10565-a21|FAIL' ; echo "fin de la comprobación"
```

**Esperado** (requiere el clon del escenario 7.b):

- `git diff --stat` del clon nombra `internal/source/boe/testdata/golden/articulo-BOE-A-2015-10565-a21.json` con `2 +-`,
  además de `schemas/bloque.json` del escenario 7.b, y termina con `2 files changed`.
- Antes de «fin de la comprobación» aparecen líneas que nombran `articulo-BOE-A-2015-10565-a21` y líneas `FAIL`.

---

## Escenario 9 — Fuzz ligero del analizador de ids de bloque (SC-007, FR-082)

```bash
rtk proxy go -C /tmp/kitlegal-quickstart-h4/copia test -run '^$' -fuzz '^FuzzIDDeBloque$' -fuzztime 10s ./internal/source/boe/
rtk proxy go test -count=1 -v -run '^(TestTipoDesdeID|TestValidarBloque|TestGramaticaCubreLosIndicesGrabados)$' ./internal/source/boe/ | rtk proxy grep -E '^--- (PASS|FAIL): '
```

**Esperado** (requiere el clon del escenario 7.b):

- El fuzz termina con `PASS` y sin `Failing input`. Se ejecuta sobre el clon, así que cualquier corpus que escribiera se
  queda allí y no en el árbol.
- Después, tres líneas `--- PASS`. Entre lo que comprueban: `a21`, `da3` y `dt1` se clasifican como artículo,
  disposición adicional y disposición transitoria; `a1-30` y `a85bis.` (las dos semillas con guion y con punto), como
  artículo; y se aceptan todos los ids de los índices grabados.

---

## Escenario 10 — El binario distribuido a mano, sin red: autodescripción, argumentos, `--offline`, `--dry-run` y el enlace (FR-001, FR-092, FR-094, FR-100, FR-101)

```bash
go build -o /tmp/kitlegal-quickstart-h4/kitlegal ./cmd/kitlegal
rtk proxy ln -sf kitlegal /tmp/kitlegal-quickstart-h4/boe
rtk proxy /tmp/kitlegal-quickstart-h4/kitlegal --help | rtk proxy grep -E '^  boe '
rtk proxy /tmp/kitlegal-quickstart-h4/kitlegal boe articulo BOE-A-2015-10565 a21 --describe | rtk proxy grep -c 'consolidacion-no-finalizada'
rtk proxy env KITLEGAL_CACHE_DIR=/tmp/kitlegal-quickstart-h4/cache-sin-crear sh -c '/tmp/kitlegal-quickstart-h4/kitlegal boe articulo BOE-A-2015 a21 --json 2>/dev/null; echo "código $?"'
rtk proxy env KITLEGAL_CACHE_DIR=/tmp/kitlegal-quickstart-h4/cache-sin-crear sh -c '/tmp/kitlegal-quickstart-h4/boe articulo BOE-A-2015-10565 a21 --offline --json 2>/dev/null; echo "código $?"'
rtk proxy env KITLEGAL_CACHE_DIR=/tmp/kitlegal-quickstart-h4/cache-sin-crear sh -c '/tmp/kitlegal-quickstart-h4/kitlegal boe articulo BOE-A-2015-10565 a21 --dry-run 2>/dev/null; echo "código $?"'
rtk proxy env KITLEGAL_CACHE_DIR=/tmp/kitlegal-quickstart-h4/cache-sin-crear sh -c '/tmp/kitlegal-quickstart-h4/kitlegal boe articulo BOE-A-2015-10565 a21 --dry-run 2>&1 >/dev/null'
rtk proxy test ! -e /tmp/kitlegal-quickstart-h4/cache-sin-crear && echo "ni los argumentos inválidos, ni --offline ni --dry-run han creado la caché"
rtk proxy test ! -e /tmp/kitlegal-quickstart-h4/boe && echo "sonda de ausencia: esta línea no debe aparecer"
```

Las invocaciones de los puntos 3 y 4 tiran la salida de error (`2>/dev/null`) para que se vea solo el sobre de la salida
estándar. La quinta tira la salida estándar para que se vea solo la de error. El `sh -c` de cada una ejecuta el binario
de la carpeta temporal por su ruta, de modo que `boe` llega como `os.Args[0]` al invocar por el enlace. Todas fijan
`KITLEGAL_CACHE_DIR` en una carpeta que no existe, y así la última comprobación cubre las cuatro.

**Esperado**, en orden:

1. Una línea de la ayuda que empieza por `  boe ` con la descripción del applet.
2. `1`: el esquema de `articulo` declara el enumerado de `codigo` (una línea con los tres valores o la primera de ellos).
3. Un sobre `{"ok":false,"fuente":"kitlegal.cli","url":"kitlegal:cli",…,"data":{"clase":"argumentos",…}}` y `código 2`, sin
   ninguna petición ni caché (id de norma mal formado).
4. Un sobre `{"ok":false,"fuente":"boe.legislacion-consolidada","url":"https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21",…,"data":{"clase":"fuente-no-disponible",…}}`
   y `código 4`: invocado por el enlace `boe`, sin red y sin entrada.
5. Solo `código 0`: la salida estándar de `--dry-run` está vacía.
6. En la salida de error, tres líneas:
   - `--dry-run: no se ha ejecutado nada; …`
   - `--dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21`
   - `--dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/metadatos`
7. «ni los argumentos inválidos, ni --offline ni --dry-run han creado la caché».
8. Ninguna línea: la sonda de ausencia sobre el enlace, que sí existe, no da verde.

---

## Escenario 11 — Códigos 2, 3, 4 y 5 por el kernel, con su sobre y su fecha (SC-008, SC-011, SC-013, FR-101, FR-128)

```bash
rtk proxy go test -count=1 -v \
  -run '^(TestCodigosDeSalidaDeBoe|TestClasesDeErrorDeBoe|TestPedirClasificaEstados|TestSinGrafoNiAsuntoNoCambianLaSalida|TestAppletBoe)$' \
  ./internal/source/boe/ ./internal/app/ | rtk proxy grep -E '^--- (PASS|FAIL): '
```

**Esperado**: cinco líneas `--- PASS` y ninguna `--- FAIL`. Ningún caso produce el código 6.

---

## Escenario 12 — `httpx` pide el formato de cada recurso (FR-003)

```bash
rtk proxy go test -count=1 -v -run '^(TestPedirConAcepta|TestConHoraRechazaNula)$' ./internal/httpx/ | rtk proxy grep -E '^--- (PASS|FAIL): '
rtk proxy go test -count=1 -v -run '^TestDirecciones$' ./internal/source/boe/ | rtk proxy grep -E '^--- (PASS|FAIL): '
```

**Esperado**: tres líneas `--- PASS`.

---

## Escenario 13 — Arquitectura, superficie del binario y espacio reservado (FR-002, FR-124, FR-127, SC-011)

### 13.a En el árbol

```bash
rtk proxy go test -count=1 -v -run '^(TestArquitectura|TestDependenciasDelBinario|TestElBinarioNoEnlazaLosEjemplos|TestLasFuentesNoFirmanComoKitlegal)$' ./internal/ \
  | rtk proxy grep -E '^--- (PASS|FAIL): '
make lint
```

**Esperado**: cuatro líneas `--- PASS` (y no aparecen `TestElBinarioNoEnlazaHTTPX` ni `TestElBinarioNoEnlazaCache`, retirados);
`make lint` sin hallazgos.

### 13.b En el clon desechable

```bash
rtk proxy perl -0pi -e 's/\z/\nvar _ = "kitlegal:applet\/boe"\n/' /tmp/kitlegal-quickstart-h4/copia/internal/source/boe/direcciones.go
rtk proxy tail -1 /tmp/kitlegal-quickstart-h4/copia/internal/source/boe/direcciones.go
rtk proxy go -C /tmp/kitlegal-quickstart-h4/copia test -count=1 -run '^TestLasFuentesNoFirmanComoKitlegal$' ./internal/ 2>&1 \
  | rtk proxy grep -E 'internal/source/boe/direcciones\.go|FAIL' ; echo "fin de la comprobación"
```

**Esperado** (requiere el clon del escenario 7.b):

- `tail -1` da `var _ = "kitlegal:applet/boe"`. Es la sonda de que la línea se ha añadido al final del fichero, detrás de
  una línea en blanco.
- Antes de «fin de la comprobación» aparecen una línea que nombra `internal/source/boe/direcciones.go` con su número de
  línea y líneas `FAIL`.

---

## Escenario 14 — El porte documentado y la fuente atada a `docs/SOURCES.md` (SC-010, FR-120-FR-123)

```bash
rtk proxy go test -count=1 -v -run '^(TestDocAnotaElPorte|TestFuenteCoincideConSources|TestFuenteNombreVigenciasYTerminos|TestGrabacionesCompletas|TestReferenciasCompletas)$' ./internal/source/boe/ \
  | rtk proxy grep -E '^--- (PASS|FAIL): '
rtk proxy grep -F '| `boe.legislacion-consolidada` |' docs/SOURCES.md
rtk proxy grep -F '| `boe.legislacion-consolidada` |' docs/SOURCES.md | rtk proxy grep -c -E '\| [0-9]{4}-[0-9]{2}-[0-9]{2} \|$'
```

**Esperado**:

- Cinco líneas `--- PASS`.
- La fila de la fuente en `docs/SOURCES.md`, con ritmo, términos y fecha de revisión.
- `1`: la última celda, «Revisado», es una fecha real y no `pendiente`. La fijó una persona en la pausa del manifiesto.

---

## Escenario 15 — Verificación contra la fuente real (SC-009, FR-115)

### 15.a Sin red: la verificación detecta una respuesta que ya no se interpreta

```bash
rtk proxy go test -count=1 -v -run '^TestVerificacionDeFuentesDetectaCambios$' ./internal/app/ | rtk proxy grep -E '^--- (PASS|FAIL): '
```

**Esperado**: `--- PASS: TestVerificacionDeFuentesDetectaCambios`. Con la grabación real, la verificación pasa. Con el
sintético `bloque-ilegible`, falla nombrando «boe articulo».

### 15.b Con red (la ejecuta el flujo nocturno)

```bash
make verify-sources
```

**Esperado**: `ok  	github.com/jmorenobl/kitlegal/internal/app`. Usa una caché temporal del propio test y no escribe en el
árbol ni en `~/.cache/kitlegal`.

---

## Escenario 16 — El veredicto del repositorio

```bash
make ci
rtk proxy git status --porcelain | rtk proxy grep -vE '^.. specs/005-h4-applet-boe-puerto/' ; echo "fin del estado"
```

**Esperado**: `ci: todos los controles en verde` y ninguna línea entre `git status` y «fin del estado». `make ci` no deja
nada sin confirmar fuera del directorio del feature. Los ficheros de estado del workflow y la evidencia de cierre quedan
fuera, como en los prerrequisitos.

---

## Limpieza

```bash
rtk proxy chmod -R u+w /tmp/kitlegal-quickstart-h4
rtk proxy rm -r /tmp/kitlegal-quickstart-h4
rtk proxy test ! -e /tmp/kitlegal-quickstart-h4 && echo "carpeta temporal borrada"
```

**Esperado**: «carpeta temporal borrada».
