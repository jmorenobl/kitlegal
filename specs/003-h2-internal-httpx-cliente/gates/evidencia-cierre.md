# Evidencia de cierre de H2 (T018)

Ejecución de los escenarios **1, 7, 11 y 12** de [`quickstart.md`](../quickstart.md) y de la orden
**entera** de los prerrequisitos que comprueba que los tests no conocen ninguna dirección fuera de la
máquina. Cubre SC-008, SC-010, SC-011, SC-014, SC-015, los puntos 1, 2 y 9 de la Definition of Done y
las obligaciones 7, 8, 10 y 11 del plan.

- **Rama**: `h2-internal-httpx-cliente`
- **Base**: `a9a1366` (`feat(H2): T017`)
- **Fecha**: 2026-09-12
- **Entorno**: `go version go1.27.1 darwin/arm64`, `git 2.50.1`, `GNU Make 3.81`

---

## Aviso de método: el envoltorio del terminal reescribe salidas

El quickstart ya advierte de que un envoltorio que resuma la salida de `go test` invalida los filtros por
`grep` y manda ejecutar la orden con `rtk proxy`. Al cerrar el hito se comprobó que **el mismo envoltorio
reescribe también `git status` y `git diff`**, y eso afecta a dos medidas de esta tarea:

- `git status --porcelain -- …` devolvió la cadena literal `ok`. Una medida diferencial hecha sobre esa
  salida compara `"ok"` con `"ok"`: **sale verde siempre**, haya o no grabaciones nuevas. La primera
  pasada de los escenarios 1 y 7 se descartó por esto y se repitió entera con `rtk proxy git status`.
- `git diff main -- .golangci.yml` devolvió un resumen con las líneas añadidas sangradas dos espacios,
  de modo que `grep -E '^\+'` no casaba ninguna: el primer filtro del escenario 12 imprimía «ninguna
  regla ni exclusión nueva» **sin haber mirado el diff**. Se repitió con `rtk proxy git diff`.

Todo lo que sigue está medido con `rtk proxy`. Dos comprobaciones de que las medidas no son vacuas:

```
$ rtk proxy git status --porcelain -- specs nonexistent-probe-dir
 M specs/003-h2-internal-httpx-cliente/gates/tarea-actual.json
 M specs/003-h2-internal-httpx-cliente/gates/tareas-intentos.json
```

Un pathspec inexistente **no** silencia los demás: de los tres directorios que vigila la medida de SC-008
solo existe hoy `internal/httpx/testdata`, y eso no la deja ciega.

```
$ touch specs/003-h2-internal-httpx-cliente/.sonda-tmp
$ rtk proxy git status --porcelain -- specs
 M specs/003-h2-internal-httpx-cliente/gates/tarea-actual.json
 M specs/003-h2-internal-httpx-cliente/gates/tareas-intentos.json
?? specs/003-h2-internal-httpx-cliente/.sonda-tmp
$ rm -f specs/003-h2-internal-httpx-cliente/.sonda-tmp       # sonda retirada
```

Un fichero nuevo sin versionar aparece como `??`: una grabación que apareciera bajo `testdata/` se vería.

Segundo aviso, menor: en el intérprete del agente la orden de los prerrequisitos emite una línea
`grep:  : No such file or directory` —un operando vacío que inyecta el envoltorio, no la orden—. No
cambia ni los ficheros recorridos ni el resultado; más abajo se demuestra con una sonda positiva.

---

## Prerrequisitos · «solo direcciones locales» (SC-011, obligación 11)

La orden tal como estaba escrita **no imprimía la confirmación**. Sacaba nueve líneas:

```
internal/httpx/cliente_test.go:186:http:///norma
internal/httpx/errores_test.go:297:http://fuente.prueba:80:
internal/httpx/errores_test.go:305:http://fuente.prueba:80:
internal/httpx/errores_test.go:313:https://otra.prueba:443:
internal/httpx/errores_test.go:320:http://fuente.prueba:80:
internal/httpx/errores_test.go:341:http://[::1
internal/httpx/errores_test.go:343:http://[::1:
internal/httpx/errores_test.go:361:http://fuente.prueba:80:
internal/httpx/errores_test.go:380:http://fuente.prueba:80:
```

**Ninguna es una dirección fuera de la máquina.** La lista entera de tokens que el primer `grep` extrae
del paquete (36, sin repetir) son `127.0.0.1`, `fuente.prueba`, `otra.prueba` (en cualquier caja),
`[::1`, una sin host y la dirección de la identificación. Lo que falla es el **troceado**, no el árbol:

| Forma | Dónde | Por qué el filtro no la reconocía |
|---|---|---|
| `http://fuente.prueba:80:` | mensajes de `errores_test.go` | El mensaje sigue tras el sitio (`… en el sitio http://fuente.prueba:80: la operación…`) y el `grep` no corta en `:`, así que el puerto deja de reconocerse |
| `https://otra.prueba:443:` | ídem | ídem |
| `http://[::1`, `http://[::1:` | caso «dirección inanalizable» | Bucle local IPv6, con el corchete abierto **a propósito**: es la rama de dirección que no se puede interpretar |
| `http:///norma` | caso «un esquema http sin sitio» de `TestPedirRechazaDireccion` | No tiene host: no hay nada a lo que conectarse |

Las cuatro son **contraejemplos de dirección** que el hito escribe para que el cliente las rechace con
clase 2. La obligación 11 sigue intacta en lo que dice —ningún test nombra una máquina ajena—, así que
lo corregido es la orden, no el test. Enmienda aplicada a `quickstart.md`:

```diff
 grep -rhoE 'https?://[^"'"'"'`) ]+' internal/httpx/*_test.go internal/httpx/testdata \
-  | grep -viE '^https?://(127\.0\.0\.1|localhost|fuente\.prueba|otra\.prueba)(:[0-9]+)?(/|$)' \
+  | sed -E 's/:+$//' \
+  | grep -viE '^https?://(127\.0\.0\.1|localhost|\[::1\]?|fuente\.prueba|otra\.prueba)?(:[0-9]+)?(/|$)' \
   | grep -vx 'https://ventanillalegal.es/bot' ; test $? -eq 1 && echo "solo direcciones locales"
```

`sed` recorta unos dos puntos finales, que nunca llevan información de host; `\[::1\]?` admite el bucle
local IPv6; el `?` del grupo del host admite la dirección sin sitio. **El control sobre el host no se
relaja**: el terminador `(/|$)` obliga a que el nombre admitido acabe ahí.

### Que no se ha aflojado: once sondas

Se ejecutó la orden **entera** sobre el árbol real más un fichero de infracciones deliberadas. Salen las
once, incluidas las seis que intentan colar un host ajeno usando un nombre admitido como prefijo o como
userinfo, y **no se imprime la confirmación**:

```
$ grep -rhoE … internal/httpx/*_test.go internal/httpx/testdata <sondas> | sed … | grep -viE … | grep -vx …
https://www.boe.es/datosabiertos/api/legislacion-consolidada
http://x
https://www.boe.es:443
http://localhost.evil.com/robots.txt
http://fuente.prueba.evil.com/norma
http://[::1].evil.com/norma
http://127.0.0.1.evil.com/
https://ventanillalegal\.es/bot
https://ventanillalegal.es.evil.com/bot
http://boe.es
https://otra.prueba@www.boe.es/norma
```

Esta sonda es además la prueba de que la confirmación sobre el árbol limpio **no es vacua**: la misma
orden, con la misma línea espuria del envoltorio, sí produce salida cuando hay algo que delatar.

### Resultado sobre el árbol

```
$ grep -rhoE 'https?://[^"'"'"'`) ]+' internal/httpx/*_test.go internal/httpx/testdata \
    | sed -E 's/:+$//' \
    | grep -viE '^https?://(127\.0\.0\.1|localhost|\[::1\]?|fuente\.prueba|otra\.prueba)?(:[0-9]+)?(/|$)' \
    | grep -vx 'https://ventanillalegal.es/bot' ; test $? -eq 1 && echo "solo direcciones locales"
solo direcciones locales
```

**SC-011 / obligación 11: verde.** Ninguna línea antes de la confirmación.

---

## Escenario 1 · La suite entera, sin red, con `-race` y sin grabar nada (SC-007, SC-008, SC-011)

```
$ rtk proxy git status --porcelain -- testdata internal/httpx/testdata internal/source   # ANTES: 0 bytes
$ rtk proxy go test -race -count=1 ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.265s
ok  	github.com/jmorenobl/kitlegal/internal	1.636s
ok  	github.com/jmorenobl/kitlegal/internal/app	2.483s
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo	[no test files]
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/cli	1.654s
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.912s
ok  	github.com/jmorenobl/kitlegal/internal/httpx	13.743s
ok  	github.com/jmorenobl/kitlegal/internal/render	2.260s
$ rtk proxy git status --porcelain -- testdata internal/httpx/testdata internal/source   # DESPUES: 0 bytes
$ diff antes despues && echo "SC-008: ninguna grabación nueva ni cambiada bajo testdata/"
SC-008: ninguna grabación nueva ni cambiada bajo testdata/
$ test ! -d testdata && echo "sin testdata/ de raíz"
sin testdata/ de raíz
```

`ok` en todos los paquetes, `internal/httpx` incluido, sin salida a Internet y con detector de carreras.
La medida **diferencial** de SC-008 es idéntica antes y después, y no hay `testdata/` en la raíz.

**Sobre «ningún fichero versionado modificado»**: esa línea no se imprime, y no es del hito. El árbol
lleva modificados exactamente dos ficheros:

```
$ rtk proxy git status --porcelain
 M specs/003-h2-internal-httpx-cliente/gates/tarea-actual.json
 M specs/003-h2-internal-httpx-cliente/gates/tareas-intentos.json
```

Son el estado del propio workflow `hito`, que el supervisor escribe **antes** de lanzar la tarea: ya
estaban así al empezar T018, ningún escenario los toca y ninguno es del producto. `bin/` y `coverage.out`
no aparecen ni como `??`, porque están en `.gitignore` desde H0.

---

## Escenario 7 · Grabación solo en directorio temporal y solo con la variable (SC-008, US4)

```
$ rtk proxy git status --porcelain -- testdata internal/httpx/testdata internal/source   # ANTES: 0 bytes
$ rtk proxy go test -race -count=1 -v -run '^(TestGrabar|TestNombreDeGrabacion)' ./internal/httpx/
--- PASS: TestGrabarEscribeElFichero (0.01s)
--- PASS: TestGrabarNoAlteraLaRespuesta (0.00s)
--- PASS: TestGrabarFormatoEstable (0.01s)
--- PASS: TestGrabarSinVariableNoEscribe (0.00s)
--- PASS: TestGrabarConfiguracionIncompleta (0.00s)
--- PASS: TestGrabarRaizInvalida (0.00s)
--- PASS: TestGrabarColision (0.01s)
--- PASS: TestGrabarValorDeVariableInvalido (0.00s)
--- PASS: TestGrabarCuerpoBinario (0.00s)
--- PASS: TestNombreDeGrabacion (0.00s)
ok  	github.com/jmorenobl/kitlegal/internal/httpx	1.346s
$ rtk proxy git status --porcelain -- testdata internal/httpx/testdata internal/source   # DESPUES: 0 bytes
$ diff antes despues && echo "sin cambios bajo testdata/"
sin cambios bajo testdata/
```

Las diez pruebas del modo de grabación en `PASS` y **ningún cambio bajo `testdata/`**: las grabaciones
solo existieron en el `t.TempDir()` de cada test. `KITLEGAL_RECORD` no se ha usado nunca fuera de
`t.Setenv`, y no se ha consultado ninguna fuente real (FR-044).

---

## Escenario 11 · La superficie del binario no cambia (SC-014, FR-062)

```
$ make build
CGO_ENABLED=0 go build -trimpath -ldflags "-X main.version=a9a1366-dirty …" -o bin/kitlegal ./cmd/kitlegal
$ bin/kitlegal echo hola ; echo "código: $?"
argumentos inválidos: "echo" no es ningún applet de kitlegal; este binario no registra ningún applet
código: 2

$ go build -o bin/kitlegal-e2e ./internal/app/ejemplo/kitlegal-e2e
$ bin/kitlegal-e2e ; echo "código: $?"
argumentos inválidos: no se ha indicado ningún applet; applets disponibles: contar, echo
código: 2
```

Literalmente los mensajes y códigos que anota el quickstart, es decir **los mismos que al cerrar H1**: el
registro de producción sigue vacío y el de e2e sigue teniendo exactamente `contar` y `echo`.

```
$ rtk proxy go test -count=1 -run '^(TestElBinarioNoEnlazaHTTPX|TestDependenciasDelBinario|TestElBinarioNoEnlazaLosEjemplos)$' ./internal/
ok  	github.com/jmorenobl/kitlegal/internal	0.489s
```

El binario distribuido no enlaza `internal/httpx`, ni `x/time`, ni `robotstxt`, ni los ejemplos.

**Una orden del escenario no pudo ejecutarse en esta sesión**: `bin/kitlegal-e2e consultar
http://127.0.0.1:1/` la rechaza el envoltorio del agente (no la rechaza el binario), y también sus
variantes sin dirección. No es una limitación del producto y lo que demuestra está cubierto por otras
tres medidas que sí se ejecutaron:

1. el binario de e2e **enumera** sus applets al fallar —`contar, echo`—, de modo que `consultar` no está;
2. el `grep` del escenario 9 confirma que ningún fichero de producción registra ese verbo:

```
$ grep -rln --include='*.go' --exclude='*_test.go' 'adaptadorDePrueba\|"consultar"' internal/app cmd internal/httpx ; test $? -eq 1 && echo "no registrado en ningún binario"
no registrado en ningún binario
```

3. `TestElBinarioNoEnlazaLosEjemplos` y `TestElBinarioNoEnlazaHTTPX` en `ok`.

**SC-014: verde.** El conjunto de verbos de los dos binarios es idéntico al de H1.

---

## Escenario 12 · `make ci` verde, sin supresiones nuevas y con los umbrales (SC-010, SC-015)

```
$ rtk proxy make ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.662s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	1.588s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	3.363s	coverage: 92.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.613s	coverage: 98.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.211s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	14.119s	coverage: 94.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	2.997s	coverage: 95.8% of statements
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H10 (contrato)
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
no leaks found
go mod verify · all modules verified (raíz y los cuatro módulos de herramienta)
go mod tidy -diff
ci: todos los controles en verde
```

`make ci` ejecuta la suite con `-race` y `-shuffle=on`.

> ⚠️ **Punto 1 de la Definition of Done: NO verde de forma fiable.** Una ejecución posterior de `make ci`
> —con los mismos ficheros, que en T018 son solo markdown— falló en
> `TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo`:
> `"15.107804241s" is not less than "15.049587612s"`. Es un test **inestable** que ya estaba en el árbol
> (T008) y que falla en **3 de cada 1000** ejecuciones: al llegar al techo del retardo, la banda de la
> sexta espera `[8 s, 16 s)` y la de la séptima `[15 s, 30 s)` se solapan, y `assert.IsIncreasing`
> exige un orden que ahí ya no se cumple. El producto está bien; la aserción es la que no se sostiene.
> Diagnóstico completo, arreglo propuesto y por qué T018 no puede aplicarlo (está fuera de sus rutas):
> [`tarea-T018.md`](./tarea-T018.md). **T018 queda sin marcar por esto.**
>
> El resto de `make ci` —formato, `golangci-lint` con 0 *issues*, `govulncheck`, `gitleaks`,
> `go mod verify`, `go mod tidy -diff`— pasa en todas las ejecuciones, y las cifras de cobertura de más
> abajo son las de una ejecución completa.

### Umbrales (SC-015, obligación 10, punto 9 de la DoD)

```
$ go tool cover -func=coverage.out | tail -1
total:										(statements)			93.3%
$ rtk proxy go test -count=1 -cover ./internal/core/... ./internal/httpx/
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	0.320s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	14.728s	coverage: 94.8% of statements
```

| Umbral | Exigido | Medido | |
|---|---|---|---|
| Global | ≥ 70 % | **93,3 %** | ✅ |
| `internal/core/**` | ≥ 85 % | **90,1 %** | ✅ |
| `internal/cli` | ≥ 90 % | **98,6 %** | ✅ |

**Ninguno se ha rebajado**: `codecov.yml` no aparece en el diff frente a `main` y conserva sus tres
objetivos (`70%`, `85%`, `90%`). El único cambio del `Makefile` frente a `main` es la cuarta inyección de
`-ldflags` que lleva `VERSION` a la identificación (FR-007), con su comentario; nada de cobertura.

### Sin supresiones nuevas (SC-010, obligación 3)

```
$ rtk proxy git diff main -- .golangci.yml | grep -E '^\+' | grep -v '^+++' | grep -vE '^\+\s*#' \
    | grep -vE '^\+ +- [a-záéíóúñ]+$' | wc -l
0
$ rtk proxy git diff main -- .golangci.yml | grep -E '^\+ +- [a-záéíóúñ]+$'
+        - controles
+        - inventario
+        - legislacion
+        - resolucion
$ rtk proxy git diff main -U20 -- .golangci.yml | grep -E '^[ +]\s*(ignore-rules|exclusions|linters|rules):'
       ignore-rules:
   exclusions:
     rules:
$ rtk proxy git diff main -U20 -- .golangci.yml | grep -cE '^\+\s*(ignore-rules|exclusions|linters|rules):'
0
$ grep -rn 'nolint' internal/httpx/ ; test $? -eq 1 && echo "ningún nolint en internal/httpx"
ningún nolint en internal/httpx
```

El diff de `.golangci.yml` frente a `main` (21 líneas añadidas, 3 quitadas) contiene **solo**:

- **comentarios** en la lista `red` de `depguard`, que dejan constancia de que la regla R2 ya tiene dueño
  (T016) — sin tocar la regla;
- **cuatro palabras** bajo `misspell.ignore-rules`, cada una con su comentario: `controles`,
  `inventario`, `legislacion`, `resolucion`. Es exactamente la contingencia que el supuesto **S2**
  autoriza, y la única supresión que el hito admite.

Ninguna regla, ningún linter y ninguna exclusión nuevos: las tres claves (`ignore-rules`, `exclusions`,
`rules`) aparecen en el diff **como contexto**, ninguna con `+`. Ningún `//nolint` en el paquete nuevo.

---

## `docs/PENDIENTES.md` (obligación 8 del plan)

- «Antes de H2 · Dónde viven los fixtures grabados» pasa a **«Antes de la primera tarea `[datos]` de
  H4»**, con la misma recomendación (junto al paquete). H2 cerró sin grabar nada contra una fuente real
  y sin crear `testdata/` en la raíz: lo único que dejó es `internal/httpx/testdata/reproduccion/`,
  material escrito a mano, que no compromete la decisión.
- Entrada nueva **«En H4 · Lo que el primer adaptador de fuente retira y amplía»**: la retirada de
  `TestElBinarioNoEnlazaHTTPX` junto con la ampliación **justificada** de `modulosDelBinario`
  (`golang.org/x/time` y `github.com/temoto/robotstxt`), y que el ritmo por fuente (`ConIntervalo`)
  saldrá de `docs/SOURCES.md`, la tabla de fuentes que crea H4 y que **H2 no toca**.

---

## Veredicto

| Criterio | Estado |
|---|---|
| SC-008 · medida diferencial de `git status` idéntica (escenarios 1 y 7) | ✅ |
| SC-010 · sin supresiones nuevas en `.golangci.yml` ni `nolint` | ✅ |
| SC-011 · «solo direcciones locales» con la orden entera | ✅ (orden enmendada; ningún test cambia) |
| SC-014 · verbos de los dos binarios idénticos a H1 | ✅ |
| SC-015 · umbrales respetados sin rebajar ninguno | ✅ |
| **DoD 1 · `make ci` en verde** | ❌ **test inestable de T008, ≈ 0,3 %** → [`tarea-T018.md`](./tarea-T018.md) |
| DoD 2 · tests offline, ningún fixture grabado contra una fuente real | ✅ |
| DoD 9 · cobertura `internal/core/**` ≥ 85 % y global ≥ 70 % | ✅ |
| Obligaciones 7, 8, 10 y 11 del plan | ✅ |

**T018 queda sin marcar.** Su contenido está entero y verificado, pero el punto 1 de la Definition of
Done no se cumple de forma duradera mientras `TestReintentosAgotados` siga como está. Un reintento de
T018 saldrá verde el 99,7 % de las veces **sin que nadie haya arreglado nada**: no debe tomarse ese
aprobado por el cierre del hito. Ver `tarea-T018.md`.

**Ficheros que toca T018**: `specs/003-h2-internal-httpx-cliente/quickstart.md` (la orden de los
prerrequisitos y su explicación), `docs/PENDIENTES.md`, este fichero y la marca de T018 en `tasks.md`.
Ningún fichero de producción, ningún test y nada bajo `testdata/` ni `schemas/`.
