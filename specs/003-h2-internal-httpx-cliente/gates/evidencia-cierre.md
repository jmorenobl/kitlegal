# Evidencia de cierre de H2 (T018)

Ejecución de los escenarios **1, 7, 11 y 12** de [`quickstart.md`](../quickstart.md) y de la orden
**entera** de los prerrequisitos que comprueba que los tests no conocen ninguna dirección fuera de la
máquina. Cubre SC-008, SC-010, SC-011, SC-014, SC-015, los puntos 1, 2 y 9 de la Definition of Done y
las obligaciones 7, 8, 10 y 11 del plan. Incluye además el arreglo del test inestable que el intento 1
de T018 encontró y que el intento 2 redelimitó (SC-005, D9), con el rojo reproducido antes y el verde
comprobado después.

- **Rama**: `h2-internal-httpx-cliente`
- **Base**: `6d98e9c` (`feat(H2): T018`, el commit del intento 2)
- **Fecha**: 2026-09-12
- **Entorno**: `go version go1.27.1 darwin/arm64`, `git 2.50.1`, `GNU Make 3.81`
- **Intento**: 3 de 3. Esta evidencia sustituye a la del intento 1; lo que aquí se mide se ha vuelto a
  ejecutar entero sobre el árbol con el arreglo aplicado.

---

## Aviso de método: el envoltorio del terminal reescribe salidas

El quickstart ya advierte de que un envoltorio que resuma la salida de `go test` invalida los filtros por
`grep` y manda ejecutar la orden con `rtk proxy`. Al cerrar el hito se comprobó que **el mismo envoltorio
reescribe también `git status` y `git diff`**, y eso afecta a dos medidas de esta tarea:

- `git status --porcelain -- …` devuelve la cadena literal `ok`. Una medida diferencial hecha sobre esa
  salida compara `"ok"` con `"ok"`: **sale verde siempre**, haya o no grabaciones nuevas.
- `git diff main -- .golangci.yml` devuelve un resumen con las líneas añadidas sangradas dos espacios,
  de modo que `grep -E '^\+'` no casa ninguna: el primer filtro del escenario 12 imprimiría «ninguna
  regla ni exclusión nueva» **sin haber mirado el diff**.

Todo lo que sigue está medido con `rtk proxy`. Dos comprobaciones de que las medidas no son vacuas,
repetidas en este intento:

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
$ rtk proxy git status --porcelain -- specs
 M specs/003-h2-internal-httpx-cliente/gates/tarea-actual.json
 M specs/003-h2-internal-httpx-cliente/gates/tareas-intentos.json
```

Un fichero nuevo sin versionar aparece como `??`: una grabación que apareciera bajo `testdata/` se vería.

Segundo aviso, menor: en el intérprete del agente la orden de los prerrequisitos emite una línea
`grep:  : No such file or directory` —un operando vacío que inyecta el envoltorio, no la orden—. No
cambia ni los ficheros recorridos ni el resultado; más abajo se demuestra con una sonda positiva.

---

## Prerrequisitos · «solo direcciones locales» (SC-011, obligación 11)

La orden tal como la escribió el plan no imprimía la confirmación: sacaba nueve tokens que **no nombran
ninguna máquina** (dos puntos finales de los mensajes de `errores_test.go`, el bucle local IPv6
`http://[::1` del caso «dirección inanalizable» y la dirección sin sitio `http:///norma` del caso «un
esquema http sin sitio»). Lo que fallaba era el **troceado**, no el árbol: el intento 1 enmendó la
orden en `quickstart.md` (un `sed -E 's/:+$//'`, `\[::1\]?` en la lista y `?` en el grupo del host) sin
tocar ningún test y sin relajar el control sobre el host, que sigue obligado a acabar en `(/|$)`. La
explicación completa está en el propio quickstart, bajo «Prerrequisitos».

### Que no se ha aflojado: once sondas

Orden **entera** sobre el árbol real más un fichero de infracciones deliberadas, escrito **fuera del
repositorio** (en el directorio temporal de la sesión) y pasado como operando adicional al primer `grep`.
Salen las once, incluidas las seis que intentan colar un host ajeno usando un nombre admitido como
prefijo o como userinfo, y **no se imprime la confirmación**:

```
$ grep -rhoE 'https?://[^"'"'"'`) ]+' internal/httpx/*_test.go internal/httpx/testdata <sondas_test.go> \
    | sed -E 's/:+$//' \
    | grep -viE '^https?://(127\.0\.0\.1|localhost|\[::1\]?|fuente\.prueba|otra\.prueba)?(:[0-9]+)?(/|$)' \
    | grep -vx 'https://ventanillalegal.es/bot' ; test $? -eq 1 && echo "solo direcciones locales"
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

**SC-011 / obligación 11: verde.** Ninguna línea antes de la confirmación. El arreglo de
`reintentos_test.go` no introduce ninguna dirección: el fichero se recorre en esta orden.

---

## El test inestable: rojo reproducido, arreglo acotado, verde comprobado (SC-005, D9)

El intento 1 encontró que `TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo`
falla en ≈ 3 de cada 1000 ejecuciones con el producto correcto: `assert.IsIncreasing` exigía orden
estricto sobre las siete esperas, y la ley `b = min(500 ms · 2^(n−1), 30 s)` solo lo garantiza
**mientras la base dobla**. En la séptima espera la base pide 32 s y el techo la deja en 30 s, de modo que
su banda `[15 s, 30 s)` se solapa con la sexta, `[8 s, 16 s)`, en `[15 s, 16 s)`. Diagnóstico completo en
[`tarea-T018.md`](./tarea-T018.md); el intento 2 redelimitó la tarea para declarar el fichero.

### Rojo antes, sobre el árbol sin tocar (base `6d98e9c`)

```
$ rtk proxy go test -count=1000 -run '^TestReintentosAgotados$/^ocho' ./internal/httpx/
--- FAIL: TestReintentosAgotados (0.00s)
    --- FAIL: TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo (0.01s)
        assertion_order.go:40:
            	Error Trace:	internal/httpx/reintentos_test.go:331
            	Error:      	"15.724237191s" is not less than "15.575673471s"
            	Messages:   	cada espera arranca donde termina la anterior (SC-005, D9)
--- FAIL: TestReintentosAgotados (0.00s)
    --- FAIL: TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo (0.01s)
            	Error:      	"15.66550447s" is not less than "15.420038058s"
--- FAIL: TestReintentosAgotados (0.00s)
    --- FAIL: TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo (0.01s)
            	Error:      	"15.08506848s" is not less than "15.007670607s"
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/httpx	9.296s
```

**3 fallos en 1000**, los tres con las dos esperas en `[15 s, 16 s)`: exactamente el solape previsto.
Tercera medida consistente (intento 1: 3/1000; intento 2: 3/1000; intento 3: 3/1000).

### El arreglo, en `internal/httpx/reintentos_test.go` y en nada más

Tres ediciones, las mismas que el intento 2 dejó comprobadas sobre una copia:

- **(a)** En `TestReintentosAgotados`, la aserción de orden pasa a
  `assert.IsIncreasing(t, pedidas[:doblan], …)` con `doblan := esperasMientrasLaBaseDobla(len(pedidas))`,
  precedida de un comentario que explica el solape; el mensaje dice ahora «mientras la base dobla, cada
  espera arranca donde termina la anterior (SC-005, D9)».
- **(b)** En `exigeEsperaDelIntento`, el mensaje del `assert.Less` dice la ley completa: «… que es su
  techo y, mientras la base dobla, el suelo de la siguiente (SC-005, D9)».
- **(c)** Función nueva `esperasMientrasLaBaseDobla(esperas int) int`, derivada de la **misma ley
  literal** que `exigeEsperaDelIntento` —500 ms, factor 2, techo 30 s— y no de las constantes del paquete,
  para ser contraste y no eco. Con siete esperas devuelve 6; con dos, 2; con cero, 0.

`rtk proxy git diff -- internal/httpx/reintentos_test.go`: 3 tramos, +27 −2 líneas. **No cambia** el
decorador de reintentos (`reintentos.go`), ninguna otra subprueba ni ningún otro fichero del paquete; la
banda `[b/2, b)` sigue comprobándose en las siete esperas por `exigeEsperaDelIntento`, y las subpruebas
de tres y de un intento no ven ningún cambio de comportamiento (con dos esperas, `doblan` es 2).

### Verde después, con la misma orden

```
$ rtk proxy go test -count=1000 -run '^TestReintentosAgotados$/^ocho' ./internal/httpx/
ok  	github.com/jmorenobl/kitlegal/internal/httpx	9.306s
$ rtk proxy go test -count=1000 -run '^TestReintentosAgotados$' ./internal/httpx/
ok  	github.com/jmorenobl/kitlegal/internal/httpx	9.223s
$ rtk proxy go test -race -count=1 -v -run '^TestReintentos' ./internal/httpx/ | grep -E '^(--- |    --- |ok|FAIL)'
--- PASS: TestReintentosCierraCuerposDescartados (0.00s)
--- PASS: TestReintentosNoRepite4xx (0.00s)
    --- PASS: TestReintentosNoRepite4xx/una_petición_mal_formada_no_mejora_por_repetirla (0.00s)
    --- PASS: TestReintentosNoRepite4xx/el_«no_encontrado»_se_entrega_tal_cual_y_no_se_repite (0.00s)
    --- PASS: TestReintentosNoRepite4xx/el_límite_de_peticiones_tiene_su_clase_y_tampoco_se_repite (0.00s)
--- PASS: TestReintentosCancelacionGana (0.00s)
    --- PASS: TestReintentosCancelacionGana/la_cancelación_con_la_petición_en_vuelo_no_se_reintenta (0.00s)
    --- PASS: TestReintentosCancelacionGana/la_cancelación_durante_la_espera_no_deja_emitir_el_intento_siguiente (0.00s)
--- PASS: TestReintentosAgotados (0.00s)
    --- PASS: TestReintentosAgotados/un_solo_intento_es_la_forma_de_no_reintentar (0.00s)
    --- PASS: TestReintentosAgotados/sin_declarar_nada_son_los_tres_intentos_por_omisión (0.00s)
    --- PASS: TestReintentosAgotados/un_fallo_de_transporte_se_repite_igual_que_un_error_del_servidor (0.00s)
    --- PASS: TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo (0.01s)
--- PASS: TestReintentosDosErroresYUnAcierto (0.61s)
ok  	github.com/jmorenobl/kitlegal/internal/httpx	2.000s
$ rtk proxy make lint
0 issues.
```

**0 fallos en 1000** sobre la subprueba que fallaba, **0 en 1000** sobre toda la tabla, los cinco tests
de reintentos con sus doce subpruebas en `PASS` con detector de carreras, y el lint limpio (sin ningún
`nolint`; el comentario evita «comparte», que `misspell` marcaba, y dice «se solapa»).

### Que acotar no ha aflojado: mutación sobre una copia desechable

Copia del árbol con `rsync -a --exclude=.git --exclude=bin --exclude=coverage.out` en el directorio
temporal de la sesión, con **una sola** mutación en `reintentos.go`: `esperaDelIntento` pasa de *equal
jitter* (`mitad + aleatoria%mitad`, banda `[b/2, b)`) a *full jitter* (`aleatoria % (mitad*2)`, banda
`[0, b)`). El árbol de trabajo no se ha tocado.

```
$ rtk proxy go test -C <copia-mutada> -count=20 -run '^(TestReintentosAgotados|TestReintentosDosErroresYUnAcierto)$' ./internal/httpx/
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/httpx	12.478s
```

| Subprueba | Cae con *full jitter* |
|---|---|
| `TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo` | **20/20** |
| `TestReintentosAgotados/sin_declarar_nada_son_los_tres_intentos_por_omisión` | 16/20 |
| `TestReintentosDosErroresYUnAcierto` | 17/20 |

La fila 10 del inventario de tests del plan («con *full jitter*, no crecen») sigue siendo cierta con la
aserción acotada: la banda `[b/2, b)` y el orden en el tramo que dobla la tumban.

---

## Escenario 1 · La suite entera, sin red, con `-race` y sin grabar nada (SC-007, SC-008, SC-011)

```
$ rtk proxy git status --porcelain -- testdata internal/httpx/testdata internal/source   # ANTES: 0 bytes
$ rtk proxy go test -race -count=1 ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.384s
ok  	github.com/jmorenobl/kitlegal/internal	1.862s
ok  	github.com/jmorenobl/kitlegal/internal/app	3.268s
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo	[no test files]
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.007s
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.477s
ok  	github.com/jmorenobl/kitlegal/internal/httpx	15.768s
ok  	github.com/jmorenobl/kitlegal/internal/render	2.735s
$ rtk proxy git status --porcelain -- testdata internal/httpx/testdata internal/source   # DESPUES: 0 bytes
$ diff antes despues && echo "SC-008: ninguna grabación nueva ni cambiada bajo testdata/"
SC-008: ninguna grabación nueva ni cambiada bajo testdata/
$ test ! -d testdata && echo "sin testdata/ de raíz"
sin testdata/ de raíz
```

`ok` en todos los paquetes, `internal/httpx` incluido, sin salida a Internet y con detector de carreras.
La medida **diferencial** de SC-008 es idéntica antes y después (0 bytes en las dos capturas, `wc -c`),
y no hay `testdata/` en la raíz.

**Sobre «ningún fichero versionado modificado»**: esa línea no se imprime, y es lo esperado en mitad de
la tarea. El árbol lleva modificados exactamente tres ficheros:

```
$ rtk proxy git status --porcelain
 M internal/httpx/reintentos_test.go
 M specs/003-h2-internal-httpx-cliente/gates/tarea-actual.json
 M specs/003-h2-internal-httpx-cliente/gates/tareas-intentos.json
```

El primero es el arreglo de esta tarea (ruta declarada). Los otros dos son el estado del propio workflow
`hito`, que el supervisor escribe **antes** de lanzar la tarea: ya estaban así al empezar T018 y ningún
escenario los toca. `bin/` y `coverage.out` no aparecen ni como `??`, porque están en `.gitignore`
desde H0.

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
ok  	github.com/jmorenobl/kitlegal/internal/httpx	1.345s
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
CGO_ENABLED=0 go build -trimpath -ldflags "-X main.version=6d98e9c-dirty … -X github.com/jmorenobl/kitlegal/internal/httpx.version=6d98e9c-dirty" -o bin/kitlegal ./cmd/kitlegal
$ bin/kitlegal echo hola ; echo "código: $?"
argumentos inválidos: "echo" no es ningún applet de kitlegal; este binario no registra ningún applet
código: 2

$ go build -o <tmp>/kitlegal-e2e ./internal/app/ejemplo/kitlegal-e2e
$ <tmp>/kitlegal-e2e ; echo "código: $?"
argumentos inválidos: no se ha indicado ningún applet; applets disponibles: contar, echo
código: 2
```

Literalmente los mensajes y códigos que anota el quickstart, es decir **los mismos que al cerrar H1**: el
registro de producción sigue vacío y el de e2e sigue teniendo exactamente `contar` y `echo`.

```
$ rtk proxy go test -count=1 -run '^(TestElBinarioNoEnlazaHTTPX|TestDependenciasDelBinario|TestElBinarioNoEnlazaLosEjemplos)$' ./internal/
ok  	github.com/jmorenobl/kitlegal/internal	0.500s
```

El binario distribuido no enlaza `internal/httpx`, ni `x/time`, ni `robotstxt`, ni los ejemplos.

**Una orden del escenario no pudo ejecutarse en esta sesión**, igual que en el intento 1:
`<tmp>/kitlegal-e2e consultar http://127.0.0.1:1/` la rechaza el envoltorio del agente (no la rechaza el
binario), y también la variante sin dirección. No es una limitación del producto y lo que demuestra está
cubierto por otras tres medidas que sí se ejecutaron:

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
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.332s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	1.829s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	3.328s	coverage: 92.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.249s	coverage: 98.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.470s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	14.390s	coverage: 94.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	2.199s	coverage: 95.8% of statements
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H10 (contrato)
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
no leaks found
go mod verify · all modules verified (raíz y los cuatro módulos de herramienta)
go mod tidy -diff
ci: todos los controles en verde
```

`make ci` ejecuta la suite con `-race` y `-shuffle=on`. **Punto 1 de la Definition of Done: verde**, y
esta vez de forma duradera: la única subprueba que lo dejaba en rojo en ≈ 3 de cada 1000 ejecuciones
está arreglada de raíz y medida ×1000 (sección anterior). Una segunda ejecución de `make ci` al terminar
la tarea, con todos los ficheros de esta evidencia ya escritos, se recoge al final.

### Umbrales (SC-015, obligación 10, punto 9 de la DoD)

```
$ go tool cover -func=coverage.out | tail -1
total:										(statements)			93.3%
$ rtk proxy go test -count=1 -cover ./internal/core/... ./internal/httpx/ ./internal/cli/
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	0.303s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	14.126s	coverage: 94.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	0.653s	coverage: 98.6% of statements
```

| Umbral | Exigido | Medido | |
|---|---|---|---|
| Global | ≥ 70 % | **93,3 %** | ✅ |
| `internal/core/**` | ≥ 85 % | **90,1 %** | ✅ |
| `internal/cli` | ≥ 90 % | **98,6 %** | ✅ |

Las mismas cifras que en los intentos 1 y 2: el arreglo solo toca un fichero de test y no cambia la
cobertura del paquete (94,8 %). **Ninguno se ha rebajado**: `codecov.yml` no aparece en el diff frente a
`main` y conserva sus tres objetivos (`70%`, `85%`, `90%`). El único cambio del `Makefile` frente a
`main` es la cuarta inyección de `-ldflags` que lleva `VERSION` a la identificación (FR-007), con su
comentario; nada de cobertura.

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
  autoriza, y la única supresión que el hito admite. T018 no añade ninguna: el arreglo del test se
  redactó en español para no necesitarla.

Ninguna regla, ningún linter y ninguna exclusión nuevos: las tres claves (`ignore-rules`, `exclusions`,
`rules`) aparecen en el diff **como contexto**, ninguna con `+`. Ningún `//nolint` en el paquete nuevo.

---

## `docs/PENDIENTES.md` (obligación 8 del plan)

Actualizado en el intento 1 y ya en la base de este intento; verificado de nuevo:

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
| SC-005 / D9 · la aserción de orden se sostiene: 0/1000 tras 3/1000, y el test sigue tumbando un *full jitter* | ✅ |
| SC-008 · medida diferencial de `git status` idéntica (escenarios 1 y 7) | ✅ |
| SC-010 · sin supresiones nuevas en `.golangci.yml` ni `nolint` | ✅ |
| SC-011 · «solo direcciones locales» con la orden entera, y once sondas delatadas | ✅ |
| SC-014 · verbos de los dos binarios idénticos a H1 | ✅ |
| SC-015 · umbrales respetados sin rebajar ninguno | ✅ |
| **DoD 1 · `make ci` en verde** | ✅ (y el único test inestable, arreglado de raíz) |
| DoD 2 · tests offline, ningún fixture grabado contra una fuente real | ✅ |
| DoD 9 · cobertura `internal/core/**` ≥ 85 % y global ≥ 70 % | ✅ |
| Obligaciones 7, 8, 10 y 11 del plan | ✅ |

**T018 queda marcada.** Todo su contenido está verificado sobre el árbol con el arreglo aplicado.

**Ficheros que toca T018** (en sus tres intentos): `internal/httpx/reintentos_test.go` (intento 3: la
aserción de orden de `TestReintentosAgotados`, el mensaje de la banda y la función que deriva el tramo),
`specs/003-h2-internal-httpx-cliente/quickstart.md` (intento 1: la orden de los prerrequisitos y su
explicación), `docs/PENDIENTES.md` (intento 1), `research.md` D9 (intento 2: «Resultado
(implementación, T018)»), `tasks.md` (intento 2: redelimitación; intento 3: la marca), este fichero y
`tarea-T018.md`. Ningún fichero de producción y nada bajo `testdata/` ni `schemas/`.

## `make ci` final del intento 3

Se ejecuta al terminar, con todos los ficheros de arriba ya escritos; su resultado se anota en
[`tarea-T018.md`](./tarea-T018.md), sección «Intento 3».
