# Quickstart: validación de H2

Guía **ejecutable** para comprobar que H2 entrega lo que dice. Cada escenario se ejecuta tal cual, desde
la raíz del repositorio, sobre la rama `h2-internal-httpx-cliente` y **con el hito ya implementado**: no
es la implementación, es cómo se verifica. Los nombres de test son los del inventario de
[plan.md](./plan.md) («Inventario de tests»).

**Sin efectos colaterales.** Ningún escenario crea ni modifica un fichero versionado. Los que necesitan
introducir violaciones deliberadas (escenario 10) lo hacen sobre una **copia desechable del árbol fuera
del repositorio** (`$(mktemp -d)`, sin `.git`) que borran al terminar: el árbol de trabajo queda limpio
por construcción. Lo único que queda en el árbol tras los escenarios es `bin/` y `coverage.out`, ambos en
`.gitignore` desde H0. **Ningún escenario toca el índice ni el historial de git, ni graba ningún fixture:
todas las grabaciones se escriben en directorios temporales de los propios tests** (`t.TempDir()`), que
Go borra al terminar. La medida de SC-008 es diferencial y acotada, como fija el spec: se compara
`git status --porcelain` restringido a las rutas donde podría aparecer una grabación, antes y después.

## Prerrequisitos

```bash
go version                          # go1.27.1 (la directiva `toolchain` de go.mod)
git --version
make --version
git rev-parse --abbrev-ref HEAD     # h2-internal-httpx-cliente
git status --porcelain | grep -v '^?? specs/' ; true   # vacío: nada versionado modificado
```

Los bloques se ejecutan con **`bash` o `zsh`**. Fuera de `go`, `git` y `make`, solo herramientas POSIX
(`grep`, `diff`, `mktemp`, `tar`, `wc`, `test`). **Ninguna comprobación necesita red**, salvo la primera
ejecución de `make ci` con la caché fría (descarga de módulos y compilación de herramientas) y
`make vuln`, que consulta la base de vulnerabilidades.

Varios escenarios filtran con `grep` la salida **en crudo** de `go test`. Si el terminal lleva un
envoltorio que la resume (por ejemplo `rtk`, que sustituye las líneas `--- PASS` por un recuento),
ejecutar la orden sin envoltorio (`rtk proxy go test …`) o leer el registro íntegro que ese envoltorio
guarda; el resultado esperado no cambia.

Los tests del hito **no conocen ninguna dirección fuera de la máquina**. Se comprueba mecánicamente:

```bash
grep -rhoE 'https?://[^"'"'"'`) ]+' internal/httpx/*_test.go internal/httpx/testdata \
  | grep -viE '^https?://(127\.0\.0\.1|localhost|fuente\.prueba|otra\.prueba)(:[0-9]+)?(/|$)' \
  | grep -vx 'https://ventanillalegal.es/bot' ; test $? -eq 1 && echo "solo direcciones locales"
```

**Esperado**: «solo direcciones locales» y ninguna línea antes. (El primer `grep` corta cada dirección
en comillas dobles o simples, acento grave, paréntesis de cierre o espacio, de modo que una dirección
dentro de una cadena en crudo de Go o del sufijo `(+https://…/bot)` sale limpia; la orden se ha
reproducido con `/usr/bin/grep` de macOS sobre muestras de las tablas del plan, con y sin
infracciones.) Los cuatro hosts admitidos son los
únicos con los que el plan escribe **todas** sus tablas (control 18 y obligación 11 de plan.md):
`127.0.0.1` y `localhost` son los servidores de bucle local que levantan los tests; `fuente.prueba` y
`otra.prueba` son los hosts ficticios de los fixtures de reproducción, de la tabla de la clave de sitio
(`TestClaveDeSitio`, escenario 4) y de la tabla de nombres del contrato de grabación §2
(`TestNombreDeGrabacion`), y nunca se conectan; el segundo filtro no distingue mayúsculas porque la
tabla de la clave de sitio incluye `http://Fuente.Prueba/ruta?q=1` para comprobar que el host se pasa a
minúsculas. `https://ventanillalegal.es/bot` es la dirección de la identificación, que se envía, no se
consulta; el tercer filtro la descarta **exacta** (`-x`), y por eso `TestAgenteDeUsuario` construye su
expresión regular con `regexp.QuoteMeta` sobre el sufijo literal ` (+https://ventanillalegal.es/bot)`
(research D5): el fuente del test contiene la dirección sin barras invertidas y el primer `grep` extrae
de él exactamente ese token. Si alguna línea aparece —`http://x` en una tabla, un `https://www.boe.es`
en un ejemplo, o `https://ventanillalegal\.es/bot\` por una expresión regular escapada a mano—, la
confirmación no se imprime y hay que corregir el test, no el filtro.

---

## Escenario 1 — La suite entera, sin red y con detector de carreras, sin grabar nada (SC-007, SC-008, SC-011)

```bash
ANTES=$(git status --porcelain -- testdata internal/httpx/testdata internal/source 2>/dev/null)
go test -race -count=1 ./...
DESPUES=$(git status --porcelain -- testdata internal/httpx/testdata internal/source 2>/dev/null)
test "$ANTES" = "$DESPUES" && echo "SC-008: ninguna grabación nueva ni cambiada bajo testdata/"
git status --porcelain | grep -v '^??' ; test $? -eq 1 && echo "ningún fichero versionado modificado"
```

**Esperado**: `ok` en todos los paquetes (incluido `internal/httpx`), las dos líneas de confirmación, y
ningún directorio `testdata/<fuente>/` nuevo en la raíz (`test ! -d testdata && echo "sin testdata/ de raíz"`).

---

## Escenario 2 — Identificación en toda petición y versión del binario (SC-001, FR-007)

```bash
go test -race -count=1 -run '^(TestAgenteDeUsuario|TestIdentificacionEnTodaPeticion)$' ./internal/httpx/
```

**Esperado**: `ok`. `TestIdentificacionEnTodaPeticion` cuenta en el servidor local las peticiones sin
la cabecera `User-Agent: kitlegal/<versión> (+https://ventanillalegal.es/bot)` —recurso, `robots.txt`,
reintentos y saltos de redirección— y exige cero.

La versión llega por `-ldflags`, con la misma `VERSION` del `Makefile` que alimenta a `kitlegal version`:

```bash
grep -n 'internal/httpx.version=\$(VERSION)' Makefile          # la inyección, en LDFLAGS
go test -count=1 -v -ldflags '-X github.com/jmorenobl/kitlegal/internal/httpx.version=9.9.9-prueba' \
  -run '^TestAgenteDeUsuario$' ./internal/httpx/ | grep -F 'kitlegal/9.9.9-prueba (+https://ventanillalegal.es/bot)'
go test -count=1 -run '^TestVersionDeLaIdentificacion$' ./cmd/kitlegal/   # sin -ldflags: "dev" en los dos
make build && bin/kitlegal version | grep -F "kitlegal $(git describe --tags --always --dirty)"
```

**Esperado**: la línea de `grep` del `Makefile`; una línea del **registro** del test (la que
`TestAgenteDeUsuario` escribe con `t.Log`, con el prefijo `agente_test.go:<línea>:`) que contiene
`kitlegal/9.9.9-prueba (+https://ventanillalegal.es/bot)`, y el test en `PASS`, no en fallo: la
expresión regular del test se construye sobre la variable `version`, no sobre el literal `dev`, y con
`regexp.QuoteMeta` sobre el sufijo ` (+https://ventanillalegal.es/bot)`, así que acepta la versión
inyectada y no deja en el fuente ninguna dirección escapada a mano (research D5; prerrequisitos); `ok`
en `cmd/kitlegal`; y la primera línea de `version` con el mismo `git describe`.

---

## Escenario 3 — `robots.txt`: se pide una vez por sitio y una ruta desautorizada no se emite (SC-003, US1-4, US1-8, US2-4)

```bash
go test -race -count=1 -v -run '^TestRobots' ./internal/httpx/ 2>&1 | grep -E '^(=== RUN|--- (PASS|FAIL)|ok|FAIL)'
```

**Esperado**: `PASS` en `TestRobotsDeniegaLaRuta` (cero accesos a la ruta en el servidor, clase 5),
`TestRobotsSePideUnaVezPorSitio` (diez rutas → un `GET /robots.txt`), `TestRobotsCasosDeObtencion` (la
tabla de FR-015: 404 → permite, vacío → permite, 5xx → 5, 429 → 5 sin reintento y cacheado, ilegible →
5, plazo → 4), `TestRobotsRedirigido` (US1-8: cadena seguida y evaluada; cadena excedida → 5),
`TestRobotsNoSeEvaluaASiMismo` y `TestRobotsPorSitio` (`http` y `https`, o dos puertos, tienen
`robots.txt` distintos).

---

## Escenario 4 — Clave de sitio y ritmo por sitio (SC-004, FR-013, FR-019, FR-022)

```bash
go test -race -count=1 -v -run '^(TestClaveDeSitio|TestRitmo)' ./internal/httpx/ 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'
```

**Esperado**: `PASS` en `TestClaveDeSitio` (la tabla de la clave, sobre los hosts ficticios
`fuente.prueba` y `otra.prueba`: `http://fuente.prueba` → `http://fuente.prueba:80`,
`https://fuente.prueba` → `https://fuente.prueba:443`, puerto explícito conservado,
`http://Fuente.Prueba/ruta?q=1` → host en minúsculas con ruta y consulta fuera,
`http://fuente.prueba` ≠ `https://fuente.prueba`, `http://fuente.prueba` ≠ `http://otra.prueba`,
`http://127.0.0.1:53211` ≠ `http://127.0.0.1:53212`),
`TestRitmoSeparaPeticionesDelMismoSitio` (dos peticiones al mismo servidor llegan separadas al menos el
intervalo configurado), `TestRitmoNoRetrasaOtroSitio` (un segundo servidor en `127.0.0.1` con otro
puerto no espera por el primero) y `TestRitmoRespetaElContexto` (contexto cancelado durante la espera
→ no se emite, clase 4).

---

## Escenario 5 — Reintentos con retardo creciente y aleatorio (SC-005, FR-025 a FR-027)

```bash
go test -race -count=1 -v -run '^TestReintentos' ./internal/httpx/ 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL|\s+reintentos_test)'
```

**Esperado**: `PASS` en `TestReintentosDosErroresYUnAcierto` (tres peticiones exactas en el servidor;
las dos esperas registradas por el reloj inyectado crecen y difieren entre dos ejecuciones),
`TestReintentosNoRepite4xx` (404 y 429: una sola petición), `TestReintentosCancelacionGana` (contexto
cancelado entre intentos → sin intento posterior, clase 4), `TestReintentosCierraCuerposDescartados` y
`TestReintentosAgotados` (5xx tras tres intentos → clase 4).

---

## Escenario 6 — Las dieciséis situaciones de fallo, con su código de salida (SC-006, SC-016)

```bash
go test -race -count=1 -v -run '^TestClasesDeError$' ./internal/httpx/ 2>&1 | grep -cE '^\s+--- PASS: TestClasesDeError/'
go test -race -count=1 -run '^(TestErrorEnvueltoConservaLaClase|TestErrorRetryAfter|TestErrorMensajesEnEspanol)$' ./internal/httpx/
go test -race -count=1 -run '^TestClasificar$' ./internal/cli/
```

**Esperado**: `16` (una subprueba por fila de `contracts/errores-y-ensayo.md` §3, cada una comprobando
`cli.Clasificar` y `cli.CodigoSalida` sobre el error real: 4, 4, 4, 5, 5, 5, 5, 5, 2, 2, 2, 2, 2, 1, 1,
1), y `ok` en las otras dos órdenes (`TestClasificar` de H1 gana los casos de `schema.ConClase`).

---

## Escenario 7 — Grabación en un directorio temporal y solo con la variable (SC-008, US4)

```bash
ANTES=$(git status --porcelain -- testdata internal/httpx/testdata internal/source 2>/dev/null)
go test -race -count=1 -v -run '^(TestGrabar|TestNombreDeGrabacion)' ./internal/httpx/ 2>&1 | grep -E '^(--- (PASS|FAIL)|ok|FAIL)'
DESPUES=$(git status --porcelain -- testdata internal/httpx/testdata internal/source 2>/dev/null)
test "$ANTES" = "$DESPUES" && echo "sin cambios bajo testdata/"
```

(`git status` sale con 0 haya o no cambios, así que comprobar su código de salida no comprueba nada:
la medida es la diferencial de SC-008, la misma del escenario 1, y solo imprime la confirmación si la
salida restringida a esas rutas es idéntica antes y después.)

**Esperado**: `PASS` en `TestGrabarEscribeElFichero` (con `t.Setenv("KITLEGAL_RECORD", "1")` y una raíz
en `t.TempDir()`: aparece `<raíz>/prueba/<nombre>.json` con petición identificada y respuesta),
`TestGrabarNoAlteraLaRespuesta`, `TestGrabarFormatoEstable` (dos grabaciones idénticas byte a byte
salvo `grabado_en` y `Date`), `TestGrabarSinVariableNoEscribe`, `TestGrabarConfiguracionIncompleta`
(sin fuente o sin raíz → clase 2 nombrando la opción, nada escrito), `TestGrabarRaizInvalida` (→ 2),
`TestGrabarColision` (→ 1 nombrando las dos peticiones), `TestGrabarValorDeVariableInvalido` (→ 2),
`TestGrabarCuerpoBinario` (`cuerpo_base64`) y `TestNombreDeGrabacion` (la tabla del contrato §2 y los
diez pares petición → fichero de los fixtures). Y «sin cambios bajo testdata/»: la salida restringida
de `git status` es la misma antes y después, porque las grabaciones solo existieron en el temporal del
test.

---

## Escenario 8 — Reproducción determinista desde un directorio preparado a mano (SC-007, US3)

```bash
ls internal/httpx/testdata/reproduccion/prueba/ | wc -l                       # 10
ls internal/httpx/testdata/reproduccion/prueba/
UNO=$(mktemp) ; DOS=$(mktemp)
go test -race -count=1 -v -run '^TestReplay' ./internal/httpx/ 2>&1 | grep -E '^--- (PASS|FAIL)' | sed -E 's/ \([0-9.]+s\)$//' | sort > "$UNO"
go test -race -count=1 -v -run '^TestReplay' -shuffle=on ./internal/httpx/ 2>&1 | grep -E '^--- (PASS|FAIL)' | sed -E 's/ \([0-9.]+s\)$//' | sort > "$DOS"
diff "$UNO" "$DOS" && echo "mismo resultado con otro orden"
rm -f "$UNO" "$DOS"
```

**Esperado**: `10`, y el listado con exactamente los diez ficheros de la lista cerrada de
[plan.md, «Fixtures»](./plan.md#fixtures): `GET_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json`,
`HEAD_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json`, `GET_http_fuente.prueba_documento.pdf.json`,
`GET_http_fuente.prueba_antigua.json`, `GET_http_fuente.prueba_movida.json`,
`GET_http_fuente.prueba_bucle.json`, `GET_http_fuente.prueba_caida.json`,
`GET_http_fuente.prueba_limitada.json`, `GET_http_fuente.prueba_a_b.json` y
`GET_http_fuente.prueba_robots.txt.json` (este último es una grabación de `robots.txt` que deniega todo
y que la reproducción deja sin usar, FR-049). `PASS` en `TestReplaySirveLasGrabaciones` (estado,
cabeceras y cuerpo grabados: texto, HEAD con cuerpo vacío y binario desde `cuerpo_base64`),
`TestReplayEsDeterminista`, `TestReplayPeticionSinGrabacion` (clase 1 nombrando la petición),
`TestReplayGrabacionDeOtraPeticion` (se pide `/a_b`, el fichero guarda `/a,b`: clase 1 nombrando las
dos), `TestReplayGarantiasVigentes` (sin contexto no compila; identificación presente; `POST` → 2),
`TestReplaySinRobots` (`con-grabacion`: el `robots.txt` grabado queda sin usar y todo se sirve;
`sin-grabacion`: copia del directorio sin ese fichero, todo se sirve; en los dos casos sin esperas),
`TestReplayRedirecciones` (302 relativo grabado → recurso final; 302 absoluto con destino ausente → 1;
302 en bucle → 4), `TestReplayEstadoDeErrorGrabado` (503 → 4, 429 → 5 con `Espera` de 120 s, sin
reintentos) y `TestReplayRechazaOpciones` (`KITLEGAL_RECORD=1` o `ConRaizDeGrabacion` → 2); y «mismo
resultado con otro orden».

---

## Escenario 9 — `--dry-run` con el kernel en proceso: cero accesos, descripción visible, código 0 (SC-013, US5, US6)

```bash
go test -race -count=1 -v -run '^TestAdaptadorDePruebaConElKernel$' ./internal/httpx/ 2>&1 \
  | grep -E '(--- (PASS|FAIL)|se habría pedido GET http://127\.0\.0\.1:[0-9]+/|ok|FAIL)'
```

**Esperado**: `PASS` en las tres subpruebas —`ensayo` (el servidor no registra ningún acceso, ni al
recurso ni a `robots.txt`; `app.Main` devuelve 0; la salida estándar está vacía; la salida de error
contiene la línea de H1 y a continuación `--dry-run: se habría pedido GET http://127.0.0.1:<puerto>/…`,
que el test imprime), `consulta` (la petición llega identificada y el sobre cita la dirección final) y
`plazo` (`--timeout` corto contra un servidor lento → código 4)— y la línea `se habría pedido` en la
salida. El adaptador de prueba —el tipo `adaptadorDePrueba`, con el verbo `consultar` (research D16)—
está en `internal/httpx/adaptador_test.go` (`package httpx_test`) y no en ningún fichero de producción
ni en ningún registro de binario:

```bash
grep -rln --include='*.go' --exclude='*_test.go' 'adaptadorDePrueba\|"consultar"' internal/app cmd internal/httpx ; test $? -eq 1 && echo "no registrado en ningún binario"
```

**Esperado**: «no registrado en ningún binario». El filtro excluye los `_test.go` y busca dos
identificadores que no existen en H1 fuera de tests (en H1 `internal/app/codigos_test.go` contiene la
palabra «consultar» en un comentario y `internal/app/registro_test.go` declara `appletDePrueba`; por
eso el filtro no busca esas formas). `grep -l` sin coincidencias sale con 1; si algún fichero de
producción nombrara el adaptador o registrara el verbo, saldría con 0 y la confirmación no se imprime.

---

## Escenario 10 — Los controles mecánicos fallan ante cada intento de saltarse la garantía (SC-012, FR-053, FR-054)

Sobre una **copia desechable** fuera del repositorio; el árbol de trabajo no se toca.

```bash
COPIA=$(mktemp -d)
tar --exclude=.git --exclude=bin --exclude=coverage.out -cf - . | tar -xf - -C "$COPIA"

# 10.a — importar net/http fuera de internal/httpx: lo detectan depguard (R2), noctx, bodyclose y el test de arquitectura
cat > "$COPIA/internal/app/violacion.go" <<'EOF'
package app

import "net/http"

func violacion() (int, error) {
	respuesta, err := http.Get("http://127.0.0.1:1/")
	if err != nil {
		return 0, err
	}

	return respuesta.StatusCode, nil
}
EOF
make -C "$COPIA" lint 2>&1 | grep -E 'R2:|must not be called|response body must be closed' ; echo "lint: $?  (0 = ha nombrado las reglas)"
( cd "$COPIA" && go test -count=1 ./internal/ -run 'TestArquitectura/R2' 2>&1 | grep -E 'R2 · importación reservada|FAIL' )
rm "$COPIA/internal/app/violacion.go"

# 10.b — crear una petición sin contexto dentro de internal/httpx: lo detecta noctx
cat > "$COPIA/internal/httpx/violacion.go" <<'EOF'
package httpx

import "net/http"

func sinContexto() (*http.Request, error) { return http.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil) }
EOF
make -C "$COPIA" lint 2>&1 | grep -F 'net/http.NewRequest must not be called' ; echo "noctx: $?"
rm "$COPIA/internal/httpx/violacion.go"

# 10.c — llamar a Pedir sin contexto: no compila
cat > "$COPIA/internal/httpx/violacion_test.go" <<'EOF'
package httpx_test

import (
	"testing"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

func TestSinContexto(t *testing.T) {
	c, _ := httpx.New()
	_, _ = c.Pedir(schema.Contexto{}, httpx.Peticion{Metodo: "GET", URL: "http://127.0.0.1:1/"})
}
EOF
( cd "$COPIA" && go vet ./internal/httpx/ 2>&1 | grep -E 'not enough arguments|cannot use' ) ; echo "compilador: $?"
rm "$COPIA/internal/httpx/violacion_test.go"

# 10.d — R2 ya no pasa en vacío: sin el dueño, el test de arquitectura falla.
# Se retira también el único fichero de test fuera del paquete que lo importa (cmd/kitlegal/main_test.go),
# para que `go list ./...` siga resolviendo el módulo y el fallo sea el del test y no el de la carga.
rm -r "$COPIA/internal/httpx" "$COPIA/cmd/kitlegal/main_test.go"
( cd "$COPIA" && go test -count=1 ./internal/ -run 'TestArquitectura/R2' 2>&1 | grep -E 'internal/httpx|FAIL' )

rm -rf "$COPIA"
git status --porcelain | grep -v '^??' ; test $? -eq 1 && echo "árbol de trabajo intacto"
```

**Esperado**: en 10.a el lint imprime al menos una línea con `R2:` (depguard), una con `must not be
called` (`noctx`) y una con `response body must be closed` (`bodyclose`), y el test de arquitectura falla
nombrando «R2 · importación reservada»; en 10.b el lint nombra `net/http.NewRequest must not be called`;
en 10.c `go vet` rechaza la llamada con `not enough arguments`; en 10.d el test de arquitectura falla
porque el grafo no contiene `internal/httpx`; y al final «árbol de trabajo intacto».

---

## Escenario 11 — La superficie del binario no cambia (SC-014, FR-062)

```bash
make build
bin/kitlegal echo hola ; echo "código: $?"
#   argumentos inválidos: "echo" no es ningún applet de kitlegal; este binario no registra ningún applet
#   código: 2
E2E=$(mktemp -d)
go build -o "$E2E/kitlegal" ./internal/app/ejemplo/kitlegal-e2e
"$E2E/kitlegal" ; echo "código: $?"
#   argumentos inválidos: no se ha indicado ningún applet; applets disponibles: contar, echo
#   código: 2
"$E2E/kitlegal" consultar http://127.0.0.1:1/ ; echo "código: $?"
#   argumentos inválidos: "consultar" no es ningún applet de kitlegal; applets disponibles: contar, echo
#   código: 2
rm -rf "$E2E"
go test -count=1 -run '^(TestElBinarioNoEnlazaHTTPX|TestDependenciasDelBinario|TestElBinarioNoEnlazaLosEjemplos)$' ./internal/
```

**Esperado**: exactamente los mensajes y códigos anotados —los mismos que al cerrar H1: el registro de
producción sigue vacío y el de e2e sigue con `contar` y `echo`, sin `consultar`— y `ok` en los tres
tests (el binario distribuido no enlaza `internal/httpx`, ni `x/time`, ni `robotstxt`, ni los ejemplos).

---

## Escenario 12 — `make ci` en verde, sin supresiones nuevas y con los umbrales de cobertura (SC-010, SC-015)

```bash
make ci
go tool cover -func=coverage.out | tail -1                                  # total ≥ 70 %
go test -count=1 -cover ./internal/core/... ./internal/httpx/               # internal/core ≥ 85 % por paquete
git diff main -- .golangci.yml | grep -E '^\+' | grep -v '^\+\s*#' | grep -v '^+++' ; test $? -eq 1 && echo "ninguna regla ni exclusión nueva"
grep -rn 'nolint' internal/httpx/ ; test $? -eq 1 && echo "ningún nolint en internal/httpx"
```

**Esperado**: «ci: todos los controles en verde»; `total: (statements) ≥ 70.0%`; una línea `ok …
internal/core/schema … coverage: ≥ 85 %` (en `main` está en 90,1 %) y la de `internal/httpx`; el diff
de `.golangci.yml` frente a `main` solo añade líneas de comentario; ningún `nolint` en el paquete nuevo.
