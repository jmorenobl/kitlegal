# Evidencia de T020 — los doce escenarios de `quickstart.md`

Ejecución local del 2026-09-11, rama `h1-kernel-cli-multicall`, sobre `fdd17e3` (árbol de trabajo con
los dos ficheros de gates del propio bucle modificados, nada más).

| Prerrequisito | Valor observado |
|---|---|
| `go version` | `go1.27.1 darwin/arm64` — el parche exacto de la directiva `toolchain` |
| `git --version` | `2.50.1 (Apple Git-155)` |
| `make --version` | `GNU Make 3.81` |
| `python3 --version` | `3.11.9` (disponible, pero no utilizable: ver §*Desviaciones de ejecución*) |
| rama | `h1-kernel-cli-multicall` |
| `git status --porcelain` al empezar | `M gates/tarea-actual.json`, `M gates/tareas-intentos.json` — los dos ficheros que el propio bucle del workflow escribe. Ninguna otra ruta. |

**Veredicto global: los doce escenarios pasan.** Los cinco controles de arquitectura se han visto fallar
uno por uno y volver a pasar; `make ci` termina en verde y el árbol de trabajo queda exactamente como
empezó.

---

## Desviaciones de ejecución (leer antes que el resto)

La sesión que ejecutó T020 corre con una lista de órdenes permitidas (`.claude/settings.json`) que es
anterior a `quickstart.md` y **no incluye** `python3 -c`, `mktemp`, `tar`, `diff`, `sleep` ni la ejecución
de un binario en una ruta arbitraria. Ampliar esa lista habría exigido crear `.claude/settings.local.json`
—que **no** está en `.gitignore` y, por tanto, habría ensuciado el árbol y lo habría rechazado el guardián
de diff— o tocar la configuración global del usuario. Ninguna de las dos cosas cabe en T020, así que cada
herramienta se ha sustituido por otra **equivalente y permitida**, sin debilitar ninguna comprobación:

| `quickstart.md` usa | Aquí se usó | Por qué es equivalente |
|---|---|---|
| `python3 -c 'json.load(...)'` | `jq -e .` / `jq -S 'del(.fecha_consulta)'` | Analiza el documento entero: si sobrara un prefijo o un sufijo, `jq -e` fallaría igual que `json.load` |
| `diff <(…) <(…)` | `git diff --no-index --exit-code` | Comparación byte a byte con código de salida |
| `mktemp -d` + `tar -cf - --exclude=.git` | `mkdir -p /tmp/t020-copia/kitlegal` + `cp -R` de todas las entradas salvo `.git`, `bin/` y `coverage.out` | Misma copia, fuera del repositorio y sin `.git` |
| `cd "$COPIA" && make lint` | `make -C /tmp/t020-copia/kitlegal lint`, `go -C … test` | Idéntico; evita el `cd` |
| binario del e2e en `$(mktemp -d)` | `go build -o ./bin/kitlegal ./internal/app/testdata/kitlegal-e2e` | **Mismo programa**, mismo código fuente; `bin/` está en `.gitignore` desde H0. Es también la única ruta ejecutable permitida. Al llegar al escenario 12, `make build` lo sustituye por el binario distribuido |
| `$E2E/salida.json`, `$E2E/err.txt`… | `bin/t020/…` | La redirección fuera del directorio de trabajo está bloqueada; `bin/` está ignorado, así que el árbol sigue limpio. `bin/t020/` se borró al terminar |

**Tres pasos manuales no se pudieron ejecutar** y se registran como tales, con la comprobación
automatizada que cubre exactamente lo mismo —ejecutada, no afirmada—:

| Paso no ejecutado | Motivo | Qué lo cubre, ejecutado aquí |
|---|---|---|
| `ln -s kitlegal echo && ./echo hola` (escenarios 3 y 12) | La lista de órdenes solo permite ejecutar `./kitlegal` y `./bin/kitlegal`; un enlace llamado `echo` no es ninguno de los dos | `TestEntregaDelHito/multicall` (guion `multicall.txtar`), que crea el enlace simbólico y compara salida, salida de error y código — **PASS** |
| `KITLEGAL_LOG=error … --dry-run` y `KITLEGAL_LOG=debug` sin banderas (escenario 7) | El prefijo de variable de entorno no pasó el análisis de la orden | `TestDryRun/la_descripción_se_ve_con_el_registro_de_eventos_apagado` y `TestRegistro/la_variable_manda_sobre_la_bandera,_subiéndolo_sin_ella` — **PASS** |
| `rm -rf "$(dirname "$COPIA")"` (final del escenario 9) | El borrado fuera del directorio de trabajo está bloqueado | Las siete violaciones se **neutralizaron sobrescribiéndolas** y la copia se verificó en verde otra vez (lint y test de arquitectura). El directorio `/tmp/t020-copia` sigue en disco: **hay que borrarlo a mano** con `rm -rf /tmp/t020-copia` |

Ninguna de las tres afecta al árbol del repositorio.

---

## Escenario 1 — Entrega literal: el sobre citable (US1 · SC-001) ✅

```
$ ./bin/kitlegal echo hola --json           # binario del e2e
{"ok":true,"fuente":"kitlegal.echo","url":"kitlegal:applet/echo",
 "fecha_consulta":"2026-09-11T20:11:10.676857+02:00",
 "hash":"sha256:f2a2800485840f313f12e331f6627b274d7cc6ae006f2d79e821832990a53fac",
 "data":{"mensaje":"hola"}}
código de salida: 0
```

- Las seis claves y **ninguna séptima**: `["data","fecha_consulta","fuente","hash","ok","url"]`.
- Procedencia y marca temporal: `kitlegal.echo  kitlegal:applet/echo  2026-09-11T20:11:18.677703+02:00  sha256:f2a2800`.
- **Verbo por omisión**: `echo hola` y `echo repetir hola`, normalizados sin `fecha_consulta`, son idénticos
  (`git diff --no-index` sin diferencias) → «el verbo por omisión no cambia el resultado».
- Sin `--json`, la tabla mínima conserva los cuatro datos de procedencia (FR-043):

  ```
  fuente          kitlegal.echo
  url             kitlegal:applet/echo
  fecha_consulta  2026-09-11T20:11:33.087178+02:00
  hash            sha256:f2a2800485840f313f12e331f6627b274d7cc6ae006f2d79e821832990a53fac
  mensaje         hola
  ```

## Escenario 2 — Con `--json`, en stdout solo hay JSON (SC-002, criterio literal del hito) ✅

```
$ KITLEGAL_LOG=debug ./bin/kitlegal echo hola --json --verbose > salida.json 2> errores.txt
código: 0
$ jq -e . salida.json       → stdout es JSON completo y nada más
$ wc -c < errores.txt       → 153
```

El registro al máximo detalle no ensucia la salida estándar; la primera línea de `errores.txt` es
`time=… level=INFO msg=invocación applet=echo verbo=repetir duracion=171.792µs argumentos="[repetir hola --json --verbose]"`.

## Escenario 3 — Un binario, muchos nombres (US2 · SC-003) ✅

El enlace simbólico lo ejerce el guion `multicall.txtar` (ver *Desviaciones*). Los casos de contorno del
despacho, ejecutados a mano:

| Invocación | Código | Salida de error |
|---|---|---|
| `./bin/kitlegal` | **2** | `argumentos inválidos: no se ha indicado ningún applet; applets disponibles: contar, echo` |
| `./bin/kitlegal noexiste` | **2** | `argumentos inválidos: "noexiste" no es ningún applet de kitlegal; applets disponibles: contar, echo` — **nombra** el applet pedido |
| `./bin/kitlegal contar hola` | **2** | `… el applet "contar" no declara ningún verbo por omisión, así que hay que nombrar uno; verbos de contar: letras, palabras` |

Y los siete guiones del e2e, ejecutados con `testscript`:

```
--- PASS: TestEntregaDelHito (0.00s)
    --- PASS: TestEntregaDelHito/describe        --- PASS: TestEntregaDelHito/ayuda
    --- PASS: TestEntregaDelHito/solo-json       --- PASS: TestEntregaDelHito/entrega
    --- PASS: TestEntregaDelHito/verbo-obligatorio
    --- PASS: TestEntregaDelHito/multicall       ← el enlace simbólico
    --- PASS: TestEntregaDelHito/argumentos
ok  github.com/jmorenobl/kitlegal/internal/app  0.821s
```

## Escenario 4 — La huella depende solo del contenido (SC-005) ✅

| Invocación | `hash` | `fecha_consulta` |
|---|---|---|
| `echo hola` | `sha256:f2a2800485…90a53fac` | `…20:12:30.850464` |
| `echo hola` | `sha256:f2a2800485…90a53fac` | `…20:12:30.854054` |
| `echo holb` | `sha256:e18043fc36…a4414da8` | `…20:12:30.857825` |

Mismo contenido e instante distinto → misma huella; un byte de diferencia → huella distinta.
`go test ./internal/core/schema/ -run TestHuella -v`: **24 subtests PASS**, incluidos
«el mismo contenido con las claves en otro orden produce la misma huella», «un entero mayor que la
precisión de la coma flotante conserva su literal» y «la huella no depende de fecha_consulta».

## Escenario 5 — Los fallos se distinguen sin leer el mensaje (US3 · SC-006 · SC-014 · SC-015) ✅

```
--- PASS: TestCodigoSalida          (correcto, argumentos, no-encontrado, fuente-no-disponible,
                                     limite-o-tos, identidad-humana, inesperado → los siete códigos)
--- PASS: TestSobreDeFallo          (los mismos, más «applet no registrado» y «bandera desconocida»)
```

Los dos casos **anteriores a la ejecución del applet**, a mano:

| Invocación | Código | `ok` / `fuente` / `url` / `data.clase` |
|---|---|---|
| `echo hola --jsno --json` | **2** | `false` · `kitlegal.cli` · `kitlegal:cli` · `argumentos` |
| `noexiste --json` | **2** | `false` · … · `argumentos` |

y el mensaje para la persona, en `stderr` (187 y 102 bytes), nunca en `stdout`:
`argumentos inválidos: unknown flag --jsno, did you mean "--json"?`.

`--timeout` inválido **sin** `--json` — la otra mitad de la regla (contrato del sobre §5): salida estándar
**vacía** y mensaje solo en la de error.

| `--timeout` | Código | `stdout` | `stderr` |
|---|---|---|---|
| `0s` | **2** | 0 bytes | `--timeout espera una duración positiva y recibió 0s` |
| `-1s` | **2** | 0 bytes | 235 bytes |
| `abc` | **2** | 0 bytes | `--timeout: expected duration but got "abc": time: invalid duration "abc"` |

Tubería cerrada y plazo agotado, que un guion no puede simular:

```
--- PASS: TestEscrituraFallida   (5 subtests, incluido «los dos descriptores rotos dejan solo el código»)
--- PASS: TestCodigoSalida/inesperado
--- PASS: TestPlazoAgotado
```

## Escenario 6 — El applet se describe a sí mismo (US5 · SC-007) ✅

```
$ ./bin/kitlegal echo hola --describe
https://json-schema.org/draft/2020-12/schema
["entrada","salida"]
la salida no contiene ningún sobre: el applet no se ejecutó
```

`go test ./internal/cli/ -run 'TestDescribe|TestContrato' -v`: **PASS**, incluidos «un esquema válido del
borrador 2020-12 con las dos partes», «describirse excluye ejecutar», «el sobre de éxito de ese verbo
valida» y «el sobre de fallo de ese mismo verbo valida» en sus **siete** clases (SC-015).

## Escenario 7 — Las mismas banderas en todos los applets (US4 · SC-010) ✅

Las ocho banderas sobre un applet que no declara ninguna, código **0**:

```
$ ./bin/kitlegal echo hola --json --timeout 5s --offline --no-graph --asunto demo --verbose
{"ok":true,"fuente":"kitlegal.echo",…,"data":{"mensaje":"hola"}}
```

`--dry-run`: `stdout` **0 bytes** también con `--json`, código **0**, y la descripción en `stderr`
(154 bytes), escrita por el presentador:

```
--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "echo", el verbo "repetir",
con los argumentos ["repetir" "hola" "--dry-run" "--json"]
```

Las tres reglas en tabla, y la privacidad del registro:

```
--- PASS: TestDryRun/la_salida_estándar_queda_vacía_también_con_--json
--- PASS: TestDryRun/la_descripción_se_ve_con_el_registro_de_eventos_apagado
--- PASS: TestRegistroPrivacidad  (4 subtests: solo con --verbose/debug viajan los argumentos)
--- PASS: TestRegistro            (16 subtests: precedencia de KITLEGAL_LOG sobre la bandera)
```

`--verbose` no toca `stdout` (documentos idénticos salvo `fecha_consulta`) y sí añade detalle en `stderr`
(0 → 152 bytes). La ayuda es texto y `--json` no la altera, en cualquier orden:
`echo --help`, `echo --json --help` y `echo --help --json` dan **el mismo fichero**, y enumera los verbos
del applet (`repetir  Devuelve el mensaje tal y como se escribió. (por omisión)`).

Herencia sin código duplicado, sobre los **dos** applets de ejemplo:

```
--- PASS: TestAppletHereda/echo    (6 subtests)
--- PASS: TestAppletHereda/contar  (6 subtests)
```

## Escenario 8 — Las reglas de arquitectura fallan si se violan (US6 · SC-008) ✅

Copia desechable en `/tmp/t020-copia/kitlegal`, **sin `.git`, sin `bin/`, sin `coverage.out`**.
Línea base de la copia: `make lint` → `0 issues`; `TestArquitectura` → **PASS** (R1, R2, R3).

Las violaciones se introdujeron **una por una** —cada una sola en el árbol, comprobada, y retirada antes
de la siguiente—, que es la lectura de T020 y da una evidencia más fuerte que introducirlas todas juntas:
cada fallo nombra **una** regla y no puede confundirse con el de otra.

| # | Regla | Fichero de la violación | `make lint` | `TestArquitectura` |
|---|---|---|---|---|
| 1 | **R1** paquete interno | `internal/core/prueba/violacion_r1.go` → `internal/render` | ❌ **falla** | ❌ **falla** |
| 2 | **R1 bis** I/O de la biblioteca estándar | mismo fichero → `log/slog` | ❌ **falla** | ❌ **falla** |
| 3 | **R2** `net/http` | `internal/cli/violacion_r2.go` | ❌ **falla** | ❌ **falla** |
| 4 | **R3** `database/sql` | `internal/cli/violacion_r2.go` | ❌ **falla** | ❌ **falla** |
| 5 | **R4** `os.Exit` | `internal/app/violacion_r4.go` | ❌ **falla** | ✅ pasa — *correcto* |
| 6 | **R5** `fmt.Println` | `internal/render/violacion_r5.go` | ❌ **falla** | ✅ pasa — *correcto* |
| 7 | **R5-descriptores** `os.Stdout` | mismo fichero | ❌ **falla** | ✅ pasa — *correcto* |

`make lint` terminó con código distinto de 0 en los siete casos (`make: *** [lint] Error 1`). Los mensajes,
literales, **nombran la regla violada** (FR-053):

```
R1  internal/core/prueba/violacion_r1.go:5:8: import 'github.com/jmorenobl/kitlegal/internal/render'
    is not allowed from list 'core': R1: el dominio no presenta nada; la presentación se inyecta
    desde la raíz de composición (depguard)
R1b internal/core/prueba/violacion_r1.go:5:8: import 'log/slog' is not allowed from list 'core':
    R1: el dominio no registra eventos; el *slog.Logger llega como argumento al método Ejecutar,
    nunca dentro de schema.Contexto (depguard)
R2  internal/cli/violacion_r2.go:3:8: import 'net/http' is not allowed from list 'red': R2: la red
    se usa a través de internal/httpx, que concentra reintentos, límite por sitio, robots.txt y
    User-Agent identificable (depguard)
R3  internal/cli/violacion_r2.go:3:8: import 'database/sql' is not allowed from list 'sql': R3: el
    acceso a base de datos vive en internal/{cache,store,graph}; el resto del árbol los usa por su
    interfaz (depguard)
R4  internal/app/violacion_r4.go:6:22: use of `os.Exit` forbidden because "R4: solo la raíz de
    composición de cada binario termina el proceso; todo lo demás devuelve un error o un código"
    (forbidigo)
R5  internal/render/violacion_r5.go:6:22: use of `fmt.Println` forbidden because "R5: la salida se
    emite por el escritor que recibe la función; solo internal/render escribe en la salida estándar"
    (forbidigo)
R5d internal/render/violacion_r5.go:10:51: use of `os.Stdout` forbidden because "R5-descriptores:
    os.Stdout solo se nombra en la raíz de composición, que lo inyecta; todo lo demás recibe un
    io.Writer" (forbidigo)
```

Capa 2, **independiente de `.golangci.yml`** — falla por R1, R2 y R3 y **solo** por ellas:

```
R1  arch_test.go:61: R1 · el dominio es puro: …/internal/core/prueba importa
    "github.com/jmorenobl/kitlegal/internal/render" … (contracts/reglas-de-arquitectura.md R1).
R1b arch_test.go:61: R1 · el dominio es puro: …/internal/core/prueba importa "log/slog"
    (lo prohíbe "log") …
R2  arch_test.go:67: R2 · importación reservada: …/internal/cli importa "net/http" … Está reservada
    a …/internal/httpx …
R3  arch_test.go:80: R3 · importación reservada: …/internal/cli importa "database/sql" … Está
    reservada a …/internal/{cache,store,graph} …
```

Con **R4**, **R5** y **R5-descriptores** activas, `TestArquitectura` sigue en `ok … 0.234s`: son reglas de
**símbolo** y el grafo de `go list -deps` no las ve. Es exactamente lo que T016 decidió y lo que el lint
cubre; fingir que el test las vigila daría una garantía falsa.

**Vuelta al verde sobre la copia sana**: `make lint` → `0 issues`; `TestArquitectura` → `PASS` en R1, R2 y R3.

## Escenario 9 — Los applets de ejemplo pasan el mismo listón, y la excepción del e2e no se desborda (D19) ✅

Sobre la misma copia, ya limpia. Los comodines de Go no descienden a `testdata`: si `TESTDATA_PKGS` no los
enumerase, este material quedaría sin lintar en silencio.

| Defecto introducido en la copia | Ruta | `make lint` |
|---|---|---|
| `fmt.Println` | `internal/app/testdata/ejemplo/violacion.go` | ❌ **falla**, `R5:` — el directorio **sí** se linta |
| `os.Exit` | `internal/app/testdata/ejemplo/violacion.go` | ❌ **falla**, `R4:` — la excepción de la raíz de composición **no** alcanza al paquete vecino |
| `os.Stdout` | `internal/app/testdata/ejemplo/violacion.go` | ❌ **falla**, `R5-descriptores:` — tampoco la de descriptores |
| `fmt.Println` | `internal/app/testdata/kitlegal-e2e/violacion_fmt.go` | ❌ **falla**, `R5:` — **dentro de la propia raíz de composición excepcionada** |

Las cuatro con código distinto de 0. El último caso es el que cierra la obligación 8: la excepción de
`^internal/app/testdata/kitlegal-e2e/` está escrita por **marca** (`R4:` y `R5-descriptores:`), nunca sobre
el linter entero, así que `fmt.Print*` —que lleva la marca `R5:` y no tiene excepción en todo el árbol—
falla **en las dos rutas**.

> Detalle observado: con el defecto declarado pero sin usar, `unused` lo señala y `forbidigo` no llega a
> verlo. Añadiendo `var _ = violacionFmt` el fallo es el esperado (`R5:`). No es un hueco del control —el
> código muerto ya lo rechaza otro linter—, pero conviene saberlo si se repite el escenario.

Y al revés: el `main` del e2e usa `os.Exit`, `os.Stdout` y `os.Stderr` **sin ninguna directiva de
supresión** y el lint pasa (`0 issues`). Comprobado en el árbol real:

```
$ grep -rn nolint internal/app/testdata/kitlegal-e2e cmd            → sin coincidencias (código 1)
$ grep -rn "nolint:forbidigo\|nolint:depguard\|nolint:all" internal cmd → sin coincidencias (código 1)
```

> ⚠️ **Discrepancia con `quickstart.md`.** El escenario 9 escribe
> `grep -rn "nolint" internal/ cmd/ ; echo "sin //nolint: $?"   # 1 = ninguna coincidencia`.
> Esa orden **no** da 1 en este árbol: hay 18 coincidencias, todas `//nolint:misspell` («Descripcion» es
> español y el diccionario de misspell es solo inglés, research.md D9) más un `//nolint:gosec` justificado
> en `internal/arch_test.go:315`. Son específicas y explicadas, y `nolintlint` con `require-explanation` y
> `require-specific` obliga a que lo sean. Lo que el escenario quiere demostrar —que **las raíces de
> composición no suprimen nada** y que **nadie silencia `forbidigo` ni `depguard`**— sí se cumple, y es lo
> que verifican las dos órdenes acotadas de arriba. La orden del quickstart está sencillamente escrita más
> ancha de lo que la afirmación necesita; corregirla cae fuera de las rutas que T020 puede tocar.

**Limpieza**: las siete violaciones se neutralizaron sobrescribiéndolas y la copia volvió a `0 issues` y a
`TestArquitectura` en verde. `/tmp/t020-copia` **queda en disco** (ver *Desviaciones*): bórrese con
`rm -rf /tmp/t020-copia`. El árbol del repositorio no se tocó en ningún instante.

## Escenario 10 — Cobertura (SC-004, criterio literal del hito) ✅

| Árbol | Umbral | Observado |
|---|---|---|
| `internal/cli` | ≥ 90 % | **98.3 %** |
| `internal/core/...` | ≥ 85 % | **90.0 %** |
| global | ≥ 70 % | **93.8 %** |

Los tres están declarados en `codecov.yml` como estados **bloqueantes** (`informational: false`), sin
exclusiones: `target: 70%` global, componente `internal_core` `target: 85%`, componente `internal_cli`
`target: 90%`.

## Escenario 11 — El veredicto agregado, con el árbol intacto (SC-013) ✅

```
$ make ci
…
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  github.com/jmorenobl/kitlegal/cmd/kitlegal         1.247s  coverage:  0.0%
ok  github.com/jmorenobl/kitlegal/internal             1.542s  coverage: [no statements]
ok  github.com/jmorenobl/kitlegal/internal/app         2.412s  coverage: 90.8%   ← el e2e va aquí dentro
ok  github.com/jmorenobl/kitlegal/internal/cli         1.793s  coverage: 98.3%
ok  github.com/jmorenobl/kitlegal/internal/core/schema 1.532s  coverage: 90.0%
ok  github.com/jmorenobl/kitlegal/internal/render      2.062s  coverage: 95.8%
No vulnerabilities found.
all modules verified  (raíz + gitleaks + golangci-lint + govulncheck + lefthook)
ci: todos los controles en verde
código: 0
```

`fmt-check`, `lint` (`0 issues`, con los dos paquetes de `testdata` enumerados), `test`, `vuln`,
`schema-check`, `secrets` (`no leaks found`), `mod-verify` y `mod-tidy-check`: los ocho de H0 más lo que
H1 añade, **sin ninguna exclusión nueva** (FR-057).

**Obligación 4 — duración del flujo agregado con el e2e dentro:**

| Medida | Valor |
|---|---|
| `make ci` completo, caché tibia | **12 s** (18:19:19 → 18:19:31 UTC) |
| umbral que fijó H0 | 180 s (3 min) |
| margen | **~15×** por debajo |

`make test-e2e` por separado: `ok github.com/jmorenobl/kitlegal/internal/app 1.899s` — deja de anunciar el
hito ausente y ejecuta los siete guiones `.txtar` sin red y sin escribir fuera de su temporal (SC-009).

> La medida es **local y con caché tibia**; la del flujo de integración continua, que es la que H0
> acotó a 3 minutos, la registra T021 sobre la propuesta de cambio. Aun multiplicando por el coste de
> una caché fría, el margen es amplio.

```
$ git status --porcelain
 M specs/002-h1-kernel-cli-multicall/gates/tarea-actual.json
 M specs/002-h1-kernel-cli-multicall/gates/tareas-intentos.json
```

Exactamente las dos rutas que ya estaban modificadas antes de empezar, las dos del propio bucle del
workflow. `make ci` no modificó ningún fichero versionado.

## Escenario 12 — El binario distribuido no registra `echo`, y `version` no cambió (FR-009 · D16) ✅

```
$ make build                       # produce bin/kitlegal, que está en .gitignore
$ ./bin/kitlegal version
kitlegal fdd17e3-dirty
commit: fdd17e36a89445812292727edd97af9d60734191
fecha:  2026-09-11T18:19:49Z
código: 0
```

Las mismas tres líneas y el mismo código 0 que en H0 (`001-h0-esqueleto-del-repo/contracts/cli-version.md`);
lo único que cambió es que el texto sale por `internal/render`.

```
$ ./bin/kitlegal echo hola
argumentos inválidos: "echo" no es ningún applet de kitlegal; este binario no registra ningún applet
código: 2
```

El mensaje distingue el caso: no es que `echo` no exista, es que **este** binario no registra applet
alguno (FR-009). La mitad del enlace simbólico la cubren `TestRegistroDeProduccion` (el registro de
producción está vacío) y `TestDespachoIndistinguible` (el despacho por nombre de invocación y por primer
argumento recorren el mismo camino), ambos **PASS**.

---

## Trazabilidad de criterios de éxito

| Criterio | Escenario | Veredicto |
|---|---|---|
| SC-001 sobre citable | 1 | ✅ |
| SC-002 solo JSON en stdout | 2 | ✅ |
| SC-003 multicall | 3, 12 | ✅ (enlace simbólico vía `multicall.txtar`) |
| SC-004 cobertura `internal/cli` ≥ 90 % | 10 | ✅ 98.3 % |
| SC-005 huella reproducible | 4 | ✅ |
| SC-006 códigos de salida | 5 | ✅ |
| SC-007 `--describe` | 6 | ✅ |
| SC-008 reglas de arquitectura vivas | 8, 9 | ✅ las cinco, una por una |
| SC-009 e2e sin red ni escritura fuera del temporal | 11 | ✅ |
| SC-010 / SC-011 las ocho banderas | 7 | ✅ |
| SC-013 `make ci` en verde, árbol intacto | 11 | ✅ 12 s |
| SC-014 / SC-015 sobre de fallo y su esquema | 5, 6 | ✅ |

## Estado del árbol al terminar T020

- `bin/` contiene solo `kitlegal` (binario distribuido de `make build`), ignorado desde H0.
- `bin/t020/`, usado como espacio de trabajo de los escenarios 1–7, borrado.
- `coverage.out` regenerado por `make test`; ignorado (`coverage.*`).
- `git status --porcelain`: las dos rutas de gates del bucle y nada más.
- **Pendiente fuera del repositorio**: `rm -rf /tmp/t020-copia`.
