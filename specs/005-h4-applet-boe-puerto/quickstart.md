# Quickstart: validación de H4

Guía **ejecutable** para comprobar que H4 entrega lo que dice. Cada escenario se ejecuta tal cual, desde la raíz del
repositorio, sobre la rama `h4-applet-boe-puerto` **con el hito ya implementado y confirmado**: no es la implementación,
es cómo se verifica. Los nombres de test son los del inventario de [plan.md](./plan.md) («Inventario de tests»).

**Sin efectos colaterales.** Ningún escenario crea ni modifica un fichero versionado, toca el índice o el historial de
git, ni escribe en la caché real de la cuenta (`~/.cache/kitlegal/`). Lo que hay que construir o romper a propósito se
hace en `/tmp/kitlegal-quickstart-h4/`: un binario, un enlace y un **clon desechable** del repositorio (`git clone`, que
solo lee el repositorio), que el último escenario borra. Lo único que puede quedar en el árbol son `coverage.out` y
`coverage-integration.out` de `make test`/`make ci`, ignorados desde H0 (`.gitignore`: `/coverage.*`).

## Prerrequisitos

```bash
go version                          # go1.27.1 (directiva toolchain de go.mod)
git --version
make --version
git rev-parse --abbrev-ref HEAD     # h4-applet-boe-puerto
rtk proxy git status --porcelain | grep -v '^?? specs/' ; echo "fin del estado"
mkdir -p /tmp/kitlegal-quickstart-h4
```

**Esperado**: entre la orden de `git status` y «fin del estado» no aparece ninguna línea (nada versionado sin confirmar).

Los bloques se ejecutan con `bash` o `zsh`. Fuera de `go`, `git` y `make`, solo herramientas POSIX y `perl` (`grep`,
`shasum`, `ls`, `ln`, `printf`, `rm`, `test`). **Ninguna comprobación necesita red**, salvo la primera ejecución con la caché
de módulos fría, `make vuln` (dentro de `make ci`) y el escenario 15.b, marcado.

Varios escenarios filtran con `grep` la salida **en crudo** de `go test -v` o de `git`. En un terminal con un envoltorio que
la resume (por ejemplo `rtk`), las órdenes van precedidas de `rtk proxy`, como aquí; sin envoltorio, se omite el prefijo y
el resultado esperado no cambia.

---

## Escenario 1 — La suite entera sin red, con detector de carreras, sin tocar la caché real ni los datos protegidos (SC-005, SC-014)

```bash
rtk proxy ls -laR ~/.cache/kitlegal 2>&1 | shasum
make test
rtk proxy ls -laR ~/.cache/kitlegal 2>&1 | shasum
rtk proxy git status --porcelain -- internal/source/boe/testdata schemas internal/app/testdata docs/SOURCES.md ; echo "fin del estado"
```

**Esperado**: las dos huellas de `shasum` son iguales (la caché de la cuenta no se ha tocado); `make test` termina sin `FAIL`
ni `WARNING: DATA RACE`, con `ok` en `internal/source/boe`, `internal/app`, `internal/httpx`, `internal/cli` e `internal`;
entre `git status` y «fin del estado» no aparece ninguna línea (ningún test escribe golden, esquemas ni grabaciones).

---

## Escenario 2 — Aceptación: el `data` de `articulo` coincide con `boe.py` en 5 artículos de 3 leyes (SC-001, FR-116)

```bash
rtk proxy go test -count=1 -v -run '^TestArticuloCoincideConBoePy$' ./internal/source/boe/ \
  | grep -E '^\s*--- (PASS|FAIL): TestArticuloCoincideConBoePy/'
```

**Esperado**: cinco líneas `--- PASS`, una por referencia (`BOE-A-2015-10565-a21`, `BOE-A-2015-10565-a1`,
`BOE-A-1985-5392-a22`, `BOE-A-2017-12902-a118`, `BOE-A-2017-12902-da3`), y ninguna `--- FAIL`. Cada subtest compara, campo a
campo, título, tipo, fecha de la versión, fecha de vigencia, norma modificadora, texto, avisos (lista de textos) y dirección
pública con la referencia que una persona escribió a mano desde `refs/boe.py` sobre la grabación, en la misma pausa en
que grabó.

---

## Escenario 3 — Caché: segunda consulta sin red, vigencias, `--offline`, ensayo y fallos no guardados (SC-003, SC-004, SC-012, FR-093, FR-094)

```bash
rtk proxy go test -count=1 -v \
  -run '^(TestCacheDeLosSeisVerbos|TestOfflineDeLosSeisVerbos|TestEnsayoDeLosSeisVerbos|TestFallosNoSeGuardan|TestArticulos)$' \
  ./internal/source/boe/ | grep -E '^--- (PASS|FAIL): '
```

**Esperado**: cinco líneas `--- PASS` (una por test) y ninguna `--- FAIL`. Entre ellos: la segunda consulta de cada verbo se
resuelve con una reproducción estricta sobre un directorio vacío, con el mismo `data`, `url` y `fecha_consulta`; pasada su
vigencia (300 s o 7 días, reloj controlado) se vuelve a pedir; `articulos` de tres bloques sin caché cuenta cuatro peticiones y
con el segundo inexistente, tres, dejando el primero en caché; con `--offline` y con `--dry-run` la caché queda intacta.

---

## Escenario 4 — La fecha de consulta es la de la consulta (FR-096, SC-008)

```bash
rtk proxy go test -count=1 -v -run '^TestFechaDeConsultaDeArticulo$' ./internal/source/boe/ | grep -E '^--- (PASS|FAIL): '
rtk proxy go test -count=1 -v -run '^TestMontadorFechaDeConsulta$' ./internal/cli/ | grep -E '^--- (PASS|FAIL): '
rtk proxy go test -count=1 -v -run '^TestInstanteDeEmision$' ./internal/httpx/ | grep -E '^\s*--- (PASS|FAIL): TestInstanteDeEmision/'
```

**Esperado**: `--- PASS: TestFechaDeConsultaDeArticulo`, `--- PASS: TestMontadorFechaDeConsulta` y nueve líneas `--- PASS` de
`TestInstanteDeEmision/…` (`acierto`, `acierto-tras-reintentos`, `fallo-tras-reintentos`, `denegada-por-robots`, `sin-turno`,
`redireccion`, `reproduccion`, `ensayo`, `argumentos`), sin ningún `FAIL`.

---

## Escenario 5 — e2e: los seis verbos sobre el binario compilado, el enlace `boe`, `--offline` y los códigos (SC-005, FR-114)

```bash
rtk proxy go test -count=1 -v -run '^TestEntregaDelHito$/^boe-' ./internal/app/ \
  | grep -E '^\s*--- (PASS|FAIL): TestEntregaDelHito/boe-'
```

**Esperado**: cinco líneas `--- PASS`: `TestEntregaDelHito/boe-verbos`, `boe-multicall`, `boe-offline`, `boe-codigos` y
`boe-cache-rapida`. El binario lo construye el propio test en un directorio temporal y responde desde las grabaciones copiadas al
`$WORK` de cada guion: ninguna conexión.

---

## Escenario 6 — Menos de 200 ms con la entrada en caché (SC-002, FR-117)

```bash
rtk proxy go test -count=1 -v -run '^TestEntregaDelHito$/^boe-cache-rapida$' ./internal/app/ \
  | grep -E 'cronometra|--- (PASS|FAIL): TestEntregaDelHito/boe-cache-rapida'
```

**Esperado**: `--- PASS: TestEntregaDelHito/boe-cache-rapida`. Con `-v`, el registro del guion muestra las diez invocaciones de
`cronometra 200ms …`; cualquiera que tardara 200 ms o más, o que emitiera una petición (la reproducción está vacía: terminaría con
código 1), haría fallar el guion.

---

## Escenario 7 — Contratos: `schemas/` coincide con `--describe`, la salida valida y la edición a mano se detecta (SC-006, FR-110, FR-111)

### 7.a En el árbol

```bash
make schema-check
rtk proxy go test -count=1 -v -run '^(TestSalidaDeBoeContraSchemas|TestEsquemasCubrenTodosLosVerbos)$' ./internal/app/ | grep -E '^--- (PASS|FAIL): '
```

**Esperado**: `make schema-check` termina con `ok  	github.com/jmorenobl/kitlegal/internal/app`; las dos líneas son `--- PASS`.

### 7.b En un clon desechable

```bash
git clone --quiet . /tmp/kitlegal-quickstart-h4/copia
perl -0pi -e 's/"minLength": 1/"minLength": 2/' /tmp/kitlegal-quickstart-h4/copia/schemas/bloque.json
make -C /tmp/kitlegal-quickstart-h4/copia schema-check 2>&1 | grep -E 'schemas/bloque\.json|FAIL' ; echo "fin de la comprobación"
```

**Esperado**: antes de «fin de la comprobación» aparecen una línea que nombra `schemas/bloque.json` y la parte de «articulo»
(la primera aparición de `"minLength": 1` en ese fichero está en la parte de `articulo`, por el orden de claves) y líneas `FAIL`.
El árbol de trabajo no cambia: la edición es en el clon.

---

## Escenario 8 — Golden: un byte distinto en `data` se detecta (FR-112, US6-3)

```bash
perl -0pi -e 's/"texto": "/"texto": "X/' /tmp/kitlegal-quickstart-h4/copia/internal/source/boe/testdata/golden/articulo-BOE-A-2015-10565-a21.json
rtk proxy go -C /tmp/kitlegal-quickstart-h4/copia test -count=1 -run '^TestGolden$' ./internal/source/boe/ 2>&1 \
  | grep -E 'articulo-BOE-A-2015-10565-a21|FAIL' ; echo "fin de la comprobación"
```

**Esperado**: líneas que nombran `articulo-BOE-A-2015-10565-a21` y `FAIL` antes de «fin de la comprobación». (Requiere el clon del
escenario 7.b.)

---

## Escenario 9 — Fuzz ligero del analizador de ids de bloque (SC-007, FR-082)

```bash
rtk proxy go -C /tmp/kitlegal-quickstart-h4/copia test -run '^$' -fuzz '^FuzzIDDeBloque$' -fuzztime 10s ./internal/source/boe/
rtk proxy go test -count=1 -v -run '^(TestTipoDesdeID|TestValidarBloque|TestGramaticaCubreLosIndicesGrabados)$' ./internal/source/boe/ | grep -E '^--- (PASS|FAIL): '
```

**Esperado** (requiere el clon del escenario 7.b): el fuzz termina con `PASS` y sin `Failing input`; se ejecuta sobre el clon, de modo que cualquier corpus que
escribiera quedaría allí y no en el árbol. Después, tres líneas `--- PASS` (entre ellas, `a21`, `da3` y `dt1` clasificados como
artículo, disposición adicional y disposición transitoria, y todos los ids de los índices grabados aceptados).

---

## Escenario 10 — El binario distribuido a mano, sin red: autodescripción, argumentos, `--offline`, `--dry-run` y el enlace (FR-001, FR-092, FR-094, FR-100, FR-101)

```bash
go build -o /tmp/kitlegal-quickstart-h4/kitlegal ./cmd/kitlegal
ln -sf kitlegal /tmp/kitlegal-quickstart-h4/boe
/tmp/kitlegal-quickstart-h4/kitlegal --help | grep -E '^  boe '
/tmp/kitlegal-quickstart-h4/kitlegal boe articulo BOE-A-2015-10565 a21 --describe | grep -c 'consolidacion-no-finalizada'
/tmp/kitlegal-quickstart-h4/kitlegal boe articulo BOE-A-2015 a21 --json ; echo "código $?"
env KITLEGAL_CACHE_DIR=/tmp/kitlegal-quickstart-h4/cache-sin-crear /tmp/kitlegal-quickstart-h4/boe articulo BOE-A-2015-10565 a21 --offline --json ; echo "código $?"
env KITLEGAL_CACHE_DIR=/tmp/kitlegal-quickstart-h4/cache-sin-crear /tmp/kitlegal-quickstart-h4/kitlegal boe articulo BOE-A-2015-10565 a21 --dry-run ; echo "código $?"
test ! -e /tmp/kitlegal-quickstart-h4/cache-sin-crear && echo "ni --offline ni --dry-run han creado la caché"
```

**Esperado**, en orden:

1. Una línea de la ayuda que empieza por `  boe ` con la descripción del applet.
2. `1`: el esquema de `articulo` declara el enumerado de `codigo` (una línea con los tres valores o la primera de ellos).
3. Un sobre `{"ok":false,"fuente":"kitlegal.cli","url":"kitlegal:cli",…,"data":{"clase":"argumentos",…}}` y `código 2`, sin
   ninguna petición ni caché (id de norma mal formado).
4. Un sobre `{"ok":false,"fuente":"boe.legislacion-consolidada","url":"https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21",…,"data":{"clase":"fuente-no-disponible",…}}`
   y `código 4`: invocado por el enlace `boe`, sin red y sin entrada.
5. En la salida de error, tres líneas: `--dry-run: no se ha ejecutado nada; …`, `--dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21`
   y `--dry-run: se habría pedido GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/metadatos`;
   salida estándar vacía y `código 0`.
6. «ni --offline ni --dry-run han creado la caché».

---

## Escenario 11 — Códigos 2, 3, 4 y 5 por el kernel, con su sobre y su fecha (SC-008, SC-011, SC-013, FR-101, FR-128)

```bash
rtk proxy go test -count=1 -v \
  -run '^(TestCodigosDeSalidaDeBoe|TestClasesDeErrorDeBoe|TestPedirClasificaEstados|TestSinGrafoNiAsuntoNoCambianLaSalida|TestAppletBoe)$' \
  ./internal/source/boe/ ./internal/app/ | grep -E '^--- (PASS|FAIL): '
```

**Esperado**: cinco líneas `--- PASS` y ninguna `--- FAIL`. Ningún caso produce el código 6.

---

## Escenario 12 — `httpx` pide el formato de cada recurso (FR-003)

```bash
rtk proxy go test -count=1 -v -run '^(TestPedirConAcepta|TestConHoraRechazaNula)$' ./internal/httpx/ | grep -E '^--- (PASS|FAIL): '
rtk proxy go test -count=1 -v -run '^TestDirecciones$' ./internal/source/boe/ | grep -E '^--- (PASS|FAIL): '
```

**Esperado**: tres líneas `--- PASS`.

---

## Escenario 13 — Arquitectura, superficie del binario y espacio reservado (FR-002, FR-124, FR-127, SC-011)

### 13.a En el árbol

```bash
rtk proxy go test -count=1 -v -run '^(TestArquitectura|TestDependenciasDelBinario|TestElBinarioNoEnlazaLosEjemplos|TestLasFuentesNoFirmanComoKitlegal)$' ./internal/ \
  | grep -E '^--- (PASS|FAIL): '
make lint
```

**Esperado**: cuatro líneas `--- PASS` (y no aparecen `TestElBinarioNoEnlazaHTTPX` ni `TestElBinarioNoEnlazaCache`, retirados);
`make lint` sin hallazgos.

### 13.b En el clon desechable

```bash
printf '\nvar _ = "kitlegal:applet/boe"\n' >> /tmp/kitlegal-quickstart-h4/copia/internal/source/boe/direcciones.go
rtk proxy go -C /tmp/kitlegal-quickstart-h4/copia test -count=1 -run '^TestLasFuentesNoFirmanComoKitlegal$' ./internal/ 2>&1 \
  | grep -E 'internal/source/boe/direcciones\.go|FAIL' ; echo "fin de la comprobación"
```

**Esperado**: una línea que nombra `internal/source/boe/direcciones.go` con su número de línea y líneas `FAIL`, antes de «fin de la
comprobación».

---

## Escenario 14 — El porte documentado y la fuente atada a `docs/SOURCES.md` (SC-010, FR-120-FR-123)

```bash
rtk proxy go test -count=1 -v -run '^(TestDocAnotaElPorte|TestFuenteCoincideConSources|TestFuenteNombreVigenciasYTerminos|TestGrabacionesCompletas|TestReferenciasCompletas)$' ./internal/source/boe/ \
  | grep -E '^--- (PASS|FAIL): '
grep -F '| `boe.legislacion-consolidada` |' docs/SOURCES.md
grep -F '| `boe.legislacion-consolidada` |' docs/SOURCES.md | grep -c -E '\| [0-9]{4}-[0-9]{2}-[0-9]{2} \|$'
```

**Esperado**: cinco líneas `--- PASS`; la fila de la fuente en `docs/SOURCES.md` con ritmo, términos y fecha de revisión; y
`1`: la última celda, «Revisado», es una fecha real y no `pendiente` (la fijó una persona en la pausa del manifiesto).

---

## Escenario 15 — Verificación contra la fuente real (SC-009, FR-115)

### 15.a Sin red: la verificación detecta una respuesta que ya no se interpreta

```bash
rtk proxy go test -count=1 -v -run '^TestVerificacionDeFuentesDetectaCambios$' ./internal/app/ | grep -E '^--- (PASS|FAIL): '
```

**Esperado**: `--- PASS: TestVerificacionDeFuentesDetectaCambios`: con la grabación real la verificación pasa y con el sintético
`bloque-ilegible` falla nombrando «boe articulo».

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
rtk proxy git status --porcelain | grep -v '^?? specs/' ; echo "fin del estado"
```

**Esperado**: `ci: todos los controles en verde` y, entre `git status` y «fin del estado», ninguna línea.

---

## Limpieza

```bash
rm -r /tmp/kitlegal-quickstart-h4
```
