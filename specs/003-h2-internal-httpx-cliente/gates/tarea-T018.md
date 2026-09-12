# T018 · intentos 1, 2 y 3: por qué no quedó en verde y cómo se cerró

**Estado**: T018 **cerrada y marcada en el intento 3** (sección final de este fichero). El **intento 1**
(primera parte) diagnosticó un test inestable que dejaba `make ci` en rojo en ≈ 3 de cada 1000
ejecuciones y que no podía tocar sin salirse de sus rutas; el **intento 2** redelimitó la tarea para
declarar el fichero y dejó el arreglo comprobado sobre una copia; el **intento 3** reprodujo el rojo
sobre el árbol, aplicó el arreglo, comprobó el verde con la misma orden y repitió la validación agregada
entera (ver [`evidencia-cierre.md`](./evidencia-cierre.md)). Las dos primeras secciones se conservan tal
como se escribieron, porque son el registro de por qué el hito no se cerró con el aprobado del azar.

---

# Intento 1

## El fallo

```
$ rtk proxy make ci
--- FAIL: TestReintentosAgotados (0.00s)
    --- FAIL: TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo (0.01s)
        assertion_order.go:40:
            	Error Trace:	internal/httpx/reintentos_test.go:331
            	Error:      	"15.107804241s" is not less than "15.049587612s"
            	Test:       	TestReintentosAgotados/ocho_intentos_recorren_el_retardo_hasta_su_techo
            	Messages:   	cada espera arranca donde termina la anterior (SC-005, D9)
FAIL	github.com/jmorenobl/kitlegal/internal/httpx	15.166s
make: *** [test] Error 1
```

**Reproducido**: 3 fallos en 1000 ejecuciones (≈ 0,3 %).

```
$ rtk proxy go test -count=1000 -run '^TestReintentosAgotados$/^ocho_intentos_recorren_el_retardo_hasta_su_techo$' ./internal/httpx/ | grep -c 'is not less than'
3
```

No lo causa T018: esta tarea solo toca markdown (`quickstart.md`, `docs/PENDIENTES.md` y los dos ficheros
de `gates/`). El primer `make ci` del día, antes de tocar nada, sí salió verde: es cuestión de suerte.

## La causa, de raíz

El producto está bien; **la aserción del test es la que no se sostiene en el último tramo**.

`baseDelIntento` (`internal/httpx/reintentos.go:190`) es `min(500 ms · 2^(n−1), 30 s)`, y
`esperaDelIntento` aplica *equal jitter*: la espera cae en `[base/2, base)`. Con `ConIntentos(8)` hay
siete esperas:

| Espera tras el intento | Base | Banda |
|---|---|---|
| 1 | 500 ms | `[250 ms, 500 ms)` |
| 2 | 1 s | `[500 ms, 1 s)` |
| 3 | 2 s | `[1 s, 2 s)` |
| 4 | 4 s | `[2 s, 4 s)` |
| 5 | 8 s | `[4 s, 8 s)` |
| 6 | 16 s | `[8 s, **16 s**)` |
| 7 | **30 s (techo)** | `[**15 s**, 30 s)` |

Mientras la base **dobla**, el suelo de cada banda es el techo de la anterior y las esperas crecen
siempre: es lo que dice el comentario de `reintentos.go:158-163` y lo que SC-005 pide. En cuanto la base
**topa** (el séptimo intento pide 32 s y se queda en 30 s), esa propiedad deja de valer: la banda 6
`[8, 16)` y la banda 7 `[15, 30)` **se solapan en `[15 s, 16 s)`**. Si las dos caen ahí y la séptima sale
menor, `assert.IsIncreasing` (`reintentos_test.go:331`) falla aunque las dos esperas sean correctas —y
`exigeEsperaDelIntento` las da por buenas, porque lo son—. Los dos fallos observados son exactamente eso:
`15,107 s → 15,049 s` y `15,818 s → 15,201 s`.

La ironía es que la subprueba que rompe es justo la que existe para **recorrer el retardo hasta su
techo**: la aserción contradice el propósito del caso que la ejecuta.

## El arreglo

En `internal/httpx/reintentos_test.go`, acotar `IsIncreasing` al tramo en que la base dobla, que es donde
la propiedad es cierta, y dejar el tramo del techo a `exigeEsperaDelIntento`, que ya lo cubre con su
banda. El corte no es un número mágico: es el último intento cuya base sigue por debajo del techo
(`500 ms · 2^(n−1) ≤ 30 s` ⇒ `n ≤ 6`), y conviene derivarlo con la misma ley literal que ya usa
`exigeEsperaDelIntento` para no volver a ser un eco de las constantes del paquete.

Conviene además que el mensaje diga la versión completa de la ley: «cada espera arranca donde termina la
anterior **mientras la base dobla**; al llegar al techo las dos últimas comparten banda».

## Por qué no lo arregla T018

`internal/httpx/reintentos_test.go` **no está en las rutas declaradas** de T018
(`docs/PENDIENTES.md`, `internal/cli`, `internal/core/**`, `quickstart.md`,
`gates/evidencia-cierre.md`), y la tarea tiene prohibido tocar o rehacer el trabajo de otra (es de T008).
El guardián de diff rechazaría la tarea entera. Se anota aquí para que se decida dónde va el arreglo.

## Aviso para el intento siguiente

**Un reintento de T018 saldrá verde el 99,7 % de las veces sin que nadie haya arreglado nada.** Si el
intento 2 pasa, el hito cierra con el test inestable dentro y el punto 1 de la Definition of Done
(«`make ci` en verde») deja de ser cierto de forma duradera: volverá a caer en la integración continua,
en la propuesta de cambio o en cualquier hito posterior, y entonces sin el diagnóstico delante.

Lo que se pide: **no cerrar el hito por el aprobado del reintento**. O se amplían las rutas de T018 con
`internal/httpx/reintentos_test.go`, o se abre una tarea para el arreglo antes de T019.

## Lo que sí está hecho y verificado

Todo lo demás de T018, con su evidencia en `evidencia-cierre.md`:

- escenarios **1, 7, 11 y 12** ejecutados y verdes (SC-008, SC-010, SC-014, SC-015);
- umbrales de cobertura: global **93,3 %** (≥ 70), `internal/core/**` **90,1 %** (≥ 85), `internal/cli`
  **98,6 %** (≥ 90), ninguno rebajado;
- diff de `.golangci.yml` frente a `main`: solo comentarios y las cuatro palabras de `misspell` que
  autoriza el supuesto S2; ninguna regla, ningún linter, ninguna exclusión, ningún `nolint`;
- medida diferencial de `git status --porcelain` idéntica antes y después (escenarios 1 y 7);
- verbos de los dos binarios idénticos a los de H1;
- la orden **entera** de los prerrequisitos imprime «solo direcciones locales» (hizo falta corregir el
  troceado de la propia orden, no ningún test: ver `evidencia-cierre.md`);
- `docs/PENDIENTES.md` actualizado (fixtures → primera tarea `[datos]` de H4; entrada nueva para H4).

El resto de `make ci` —formato, `golangci-lint` con 0 *issues*, `govulncheck`, `gitleaks`, `go mod
verify`, `go mod tidy -diff`— pasa en todas las ejecuciones; lo único rojo es la subprueba de arriba.

---

# Intento 2 · redelimitar la tarea y dejar el arreglo comprobado

**Estado al terminar**: T018 sigue **sin marcar**, a propósito. Este intento **no** ha tocado ningún fichero
de producción ni de test: solo `tasks.md` (la línea de T018 y una nota en §Notas), `research.md` (D9,
«Resultado (implementación)») y este fichero, todos dentro del directorio del feature. `make ci` se ha
ejecutado igualmente al terminar (ver abajo).

## Por qué no se cierra en este intento

El diagnóstico del intento 1 se ha vuelto a comprobar sobre el código (`reintentos.go:190-202` y
`reintentos_test.go:324-331`) y contra el spec: **SC-005 mide tres intentos**, es decir dos esperas, y
**D9** enuncia «el mínimo del intento *n+1* es el máximo del intento *n*» a partir de la ley
`b = min(500 ms · 2^(n−1), 30 s)`, que solo lo cumple mientras la base dobla. El producto aplica esa ley
tal cual; la aserción `IsIncreasing` sobre las siete esperas del caso de ocho intentos pide lo que D9 no
promete en el techo. **Es el test.**

Las rutas de un intento las congela `siguiente_tarea` en `gates/tarea-actual.json` **antes** de lanzarlo,
a partir de la línea de la tarea en `tasks.md` tal como estaba; `guardian_diff` lee ese JSON, no vuelve a
leer `tasks.md`. Para este intento las rutas eran `docs/PENDIENTES.md`, `internal/cli`, `internal/core/**`,
`quickstart.md` y `gates/evidencia-cierre.md`: `internal/httpx/reintentos_test.go` no está, y tocarlo
pararía el run en el guardián. Las salidas, por orden:

- **Reintentar y marcar si sale verde** (99,7 % de probabilidad): es cerrar el hito con un test inestable
  dentro y dar por cierto el punto 1 de la Definition of Done sabiendo que no lo es. Descartado.
- **Aplicar el arreglo y dejar que el guardián pare el run**: arregla la raíz pero convierte un hito
  desatendido en una parada deliberada que alguien tiene que resolver a mano. Descartado mientras haya
  una salida que no pare.
- **Ampliar a mano las rutas de `gates/tarea-actual.json`**: el agente se ampliaría las rutas en el mismo
  paso en que las usa; es relajar el gate (`gates/tarea-T010.md` ya lo descartó). Descartado.
- **Redelimitar T018 en `tasks.md` y dejarla `[ ]`** para que el intento 3 la cierre con las rutas
  ampliadas: es el procedimiento que T010 estableció para el falso positivo de `misspell` y que §Notas de
  `tasks.md` recoge («se anota, se redelimita acotado y se deja sin marcar para que el siguiente intento
  la cierre»). El guardián de este intento pasa (todo queda en el directorio del feature), `verificar`
  pasa y el run sigue sin pausa. **Es la salida aplicada.**

## Lo que ha cambiado en este intento

1. **`tasks.md`, línea de T018**: declara ahora `internal/httpx/reintentos_test.go`, «acotado a **la
   aserción de orden de `TestReintentosAgotados`** —acotarla al tramo en que la base del retardo dobla,
   derivado con la misma ley literal que `exigeEsperaDelIntento`, y decir la ley completa en el mensaje de
   la banda—, sin tocar el decorador de reintentos, ninguna otra subprueba ni ningún otro fichero del
   paquete, con el rojo reproducido antes (`-count=1000` sobre la subprueba de ocho intentos) y el verde
   comprobado después con la misma orden». El extractor de rutas del workflow, ejecutado sobre la línea
   nueva con su propio filtro, devuelve exactamente: `docs/PENDIENTES.md`, `internal/cli`,
   `internal/core/**`, `internal/httpx/reintentos_test.go`, `quickstart.md`,
   `specs/003-h2-internal-httpx-cliente/gates/evidencia-cierre.md`. La línea no nombra el fichero de
   producto del decorador a propósito: el extractor declararía como ruta cualquier token con forma de
   fichero, aunque estuviera precedido de «sin tocar».
2. **`tasks.md` §Notas**: nota nueva sobre la redelimitación de T018, con la causa y por qué no es un
   atajo.
3. **`research.md`, D9**: párrafo «Resultado (implementación, T018)»: la propiedad «crecen» es exacta
   mientras la base dobla (seis primeras esperas); la séptima topa el techo y su banda se solapa con la
   sexta en `[15 s, 16 s)`; SC-005 nunca llega ahí; la decisión no cambia.

## El arreglo, comprobado sobre una copia desechable fuera del repositorio

Copia con `rsync -a --exclude=.git --exclude=bin --exclude=coverage.out` en el directorio temporal de la
sesión; el árbol de trabajo no se ha tocado. Sobre la copia se aplicaron **exactamente** estas tres
ediciones a `internal/httpx/reintentos_test.go`, y ninguna otra:

**(a)** En `TestReintentosAgotados`, la aserción de orden (hoy `assert.IsIncreasing(t, pedidas, …)`, justo
después del bucle que llama a `exigeEsperaDelIntento`):

```go
			for intento, pedida := range pedidas {
				exigeEsperaDelIntento(t, intento+1, pedida)
			}

			// El orden estricto solo vale mientras la base dobla, que es cuando
			// el suelo de cada banda es el techo de la anterior. Al topar el
			// techo, la banda [15 s, 30 s) se solapa con la anterior,
			// [8 s, 16 s), y esas dos esperas pueden salir en cualquier orden
			// sin dejar de ser correctas: ese tramo lo cubre ya la banda que
			// exigeEsperaDelIntento acaba de comprobar.
			doblan := esperasMientrasLaBaseDobla(len(pedidas))
			assert.IsIncreasing(t, pedidas[:doblan],
				"mientras la base dobla, cada espera arranca donde termina la anterior (SC-005, D9)")
```

**(b)** En `exigeEsperaDelIntento`, el mensaje del `assert.Less` pasa a decir la ley completa:

```go
		"y no alcanza la base del intento %d, que es su techo y, mientras la base dobla, el suelo de la siguiente (SC-005, D9)",
```

**(c)** Inmediatamente antes del comentario de `esperasAnotadas`, el tramo derivado de la misma ley
literal:

```go
// esperasMientrasLaBaseDobla cuenta, de una serie de esperas consecutivas,
// las primeras cuya base no la ha cortado todavía el techo: en ese tramo la
// base dobla de una espera a la siguiente, cada banda arranca donde termina
// la anterior y el orden estricto es cierto. Se deriva de la misma ley
// literal que exigeEsperaDelIntento —500 ms, factor 2, techo 30 s— y no de
// las constantes del paquete, por la misma razón: ser contraste y no eco.
func esperasMientrasLaBaseDobla(esperas int) int {
	base := 500 * time.Millisecond
	tramo := 0

	for tramo < esperas && base <= 30*time.Second {
		tramo++
		base *= 2
	}

	return tramo
}
```

Traza: con siete esperas devuelve **6** (bases 500 ms, 1, 2, 4, 8 y 16 s; la siguiente pediría 32 s y el
techo la corta), con dos devuelve 2 y con cero, 0; `pedidas[:doblan]` nunca se sale del corte.

**Aviso S2, ya resuelto en el texto de arriba**: la primera redacción del comentario de (a) decía «comparte
suelo» y `misspell` lo marcó (`comparte` → `compare`). Se reescribió en español («se solapa»), que es lo
que §Notas manda; el texto de arriba es el que pasa el lint. No usar «comparte» en ese comentario.

### Medidas sobre la copia

| Comprobación | Orden | Resultado |
|---|---|---|
| Lint del paquete con el arreglo | `make -C <copia> lint` | `0 issues.` |
| Todos los tests de reintentos, con detector de carreras | `go test -race -count=1 -v -run '^TestReintentos' ./internal/httpx/` | 5 tests y 12 subpruebas en `PASS`, `ok` |
| La subprueba que fallaba, mil veces | `go test -count=1000 -run '^TestReintentosAgotados$' ./internal/httpx/` | `ok`, 0 fallos |
| **Rojo antes**, sobre el árbol **sin** arreglo | `go test -count=1000 -run '^TestReintentosAgotados$/^ocho' ./internal/httpx/` | **3** `is not less than` en 1000 (el intento 1 midió 3/1000 también) |
| El test acotado **sigue teniendo dientes** (mutación en una segunda copia: `esperaDelIntento` devuelve *full jitter* `aleatoria % (mitad*2)`, es decir `[0, b)`) | la orden de mutación de abajo, 20 veces | `FAIL`: la subprueba de ocho intentos cae **19/20**, la de tres por omisión **16/20** y `TestReintentosDosErroresYUnAcierto` **18/20** (por la banda `[b/2, b)` y por el orden en el tramo que dobla) |

Orden de la mutación (sobre la segunda copia, con `go test -C`):

```bash
go test -C <copia-mutada> -count=20 -run '^(TestReintentosAgotados|TestReintentosDosErroresYUnAcierto)$' ./internal/httpx/
```

La última fila es la que dice que acotar la aserción no ha aflojado nada: la fila 10 del inventario de
tests del plan («con *full jitter*, no crecen») sigue siendo cierta con el test acotado.

## Qué hace el intento 3, en orden

1. Leer este fichero. Las rutas de la tarea incluyen ya `internal/httpx/reintentos_test.go`; **solo** ese
   fichero cambia fuera del directorio del feature.
2. **Rojo primero**: sobre el árbol sin tocar, `rtk proxy go test -count=1000 -run
   '^TestReintentosAgotados$/^ocho' ./internal/httpx/` y contar las líneas `is not less than` (≈ 3; tarda
   alrededor de un minuto). Si en una tirada salieran 0, la reproducción vale igual: está medida dos veces
   (intento 1 e intento 2, 3/1000 cada una) y la causa es aritmética, no una carrera.
3. Aplicar **(a)**, **(b)** y **(c)** tal cual, con la herramienta de edición (en esta sesión `python3`,
   `tee` y `cd` combinado con escritura estaban denegados por permisos; `go test -C <dir>` sí funciona).
4. Verde: la misma orden de 2 con `-count=1000` → `ok`; `rtk proxy make ci` → «ci: todos los controles en
   verde».
5. Repetir los escenarios **1, 7, 11 y 12** del quickstart y la orden **entera** de los prerrequisitos
   —siempre con `rtk proxy` delante de `go test`, `git status` y `git diff`, por lo que explica el aviso
   de método de `evidencia-cierre.md`— y **refrescar `evidencia-cierre.md`**: base nueva, el aviso del
   punto 1 de la Definition of Done retirado, una sección con el arreglo y sus medidas (las de arriba,
   repetidas sobre el árbol), el veredicto con todo en ✅ y `internal/httpx/reintentos_test.go` en la
   lista de ficheros que toca T018.
6. Marcar T018 `[X]`. Es el último intento: si algo no cuadra, anotarlo aquí y dejarla `[ ]`.

## Observación sobre el workflow, sin efecto en esta tarea

`siguiente_tarea` marca `datos: true` para T018 porque su texto contiene el literal `[datos]` (al hablar
de «la primera tarea `[datos]` de H4»), no porque la tarea lleve la etiqueta. Aquí es inocuo —ningún
intento toca `testdata/` ni `schemas/`—, pero significa que el guardián no habría rechazado un cambio bajo
`testdata/` en T018. Si se quiere cerrar, la etiqueta debería reconocerse solo al principio de la línea
(`- [ ] Tnnn [datos]`), igual que hace `precheck_tasks` con el formato. Queda anotado para el workflow,
no para este hito.

## `make ci` de este intento

Ejecutado sobre el árbol de trabajo al terminar, con las ediciones de arriba (todas markdown):
«ci: todos los controles en verde», `golangci-lint` con `0 issues.`, suite con `-race -shuffle=on` en
`ok` en todos los paquetes (`internal/httpx` 94,8 %, `internal/cli` 98,6 %, `internal/core/schema`
90,1 %, mismas cifras que el intento 1), `govulncheck` sin vulnerabilidades, `gitleaks` sin fugas, `go mod
verify` y `go mod tidy -diff` limpios. Es el aprobado del 99,7 % que el intento 1 advirtió que no debía
tomarse por el cierre: T018 queda `[ ]` por lo dicho en «Por qué no se cierra en este intento», y el
intento 3 la cierra con el arreglo aplicado. `git status --porcelain` al terminar: solo `tasks.md`,
`research.md`, este fichero y los dos JSON de estado del supervisor, todos en el directorio del feature;
nada bajo `internal/`, `testdata/` ni `schemas/`.

---

# Intento 3 · el arreglo aplicado, el rojo reproducido y el verde comprobado

**Estado al terminar**: T018 **marcada `[X]`**. Se ha seguido, en orden, la lista «Qué hace el intento 3»
del intento 2. Rutas congeladas por `siguiente_tarea` para este intento: `docs/PENDIENTES.md`,
`internal/cli`, `internal/core/**`, `internal/httpx/reintentos_test.go`, `quickstart.md` y
`gates/evidencia-cierre.md`; el único fichero que cambia fuera del directorio del feature es
`internal/httpx/reintentos_test.go`.

## 1. Rojo primero, sobre el árbol sin tocar (base `6d98e9c`)

```
$ rtk proxy go test -count=1000 -run '^TestReintentosAgotados$/^ocho' ./internal/httpx/ | grep -c 'is not less than'
3
```

Las tres parejas: `15,724 s → 15,576 s`, `15,666 s → 15,420 s`, `15,085 s → 15,008 s`. Todas en
`[15 s, 16 s)`, el solape de las bandas sexta y séptima. Tercera medida consistente: 3/1000 en cada uno
de los tres intentos. La causa es aritmética, no una carrera.

## 2. El arreglo: (a), (b) y (c) tal cual

Aplicadas con la herramienta de edición las tres ediciones que el intento 2 dejó escritas y comprobadas
sobre la copia, sin ningún cambio de texto. `rtk proxy git diff -- internal/httpx/reintentos_test.go`:
3 tramos, +27 −2. `reintentos.go` no cambia; ninguna otra subprueba ni ningún otro fichero del paquete
cambia.

Traza de `esperasMientrasLaBaseDobla` sobre el árbol: con siete esperas (caso de ocho intentos) devuelve
6 —bases 500 ms, 1, 2, 4, 8 y 16 s; la siguiente pediría 32 s y `base <= 30 s` la corta—; con dos (tres
intentos) devuelve 2, así que esa subprueba y la de un intento no ven ningún cambio.

## 3. Verde después, con la misma orden

| Comprobación | Orden | Resultado |
|---|---|---|
| La subprueba que fallaba, mil veces | `rtk proxy go test -count=1000 -run '^TestReintentosAgotados$/^ocho' ./internal/httpx/` | `ok` (9,3 s), **0** `is not less than` |
| Toda la tabla, mil veces | `rtk proxy go test -count=1000 -run '^TestReintentosAgotados$' ./internal/httpx/` | `ok` (9,2 s) |
| Todos los tests de reintentos, con detector de carreras | `rtk proxy go test -race -count=1 -v -run '^TestReintentos' ./internal/httpx/` | 5 tests y 12 subpruebas en `PASS`, `ok` |
| Lint del árbol | `rtk proxy make lint` | `0 issues.` |
| **Mutación** en una copia desechable (`rsync` al temporal de la sesión; `esperaDelIntento` devuelve *full jitter* `aleatoria % (mitad*2)`) | `rtk proxy go test -C <copia> -count=20 -run '^(TestReintentosAgotados\|TestReintentosDosErroresYUnAcierto)$' ./internal/httpx/` | `FAIL`: ocho intentos **20/20**, tres por omisión 16/20, `TestReintentosDosErroresYUnAcierto` 17/20 |

La última fila repite sobre el árbol arreglado la prueba del intento 2 de que acotar la aserción no la
ha dejado sin dientes. El árbol de trabajo no se ha tocado para la mutación: solo la copia.

## 4. Validación agregada repetida entera

Escenarios **1, 7, 11 y 12** y la orden **entera** de los prerrequisitos, todos con `rtk proxy` delante
de `go test`, `git status` y `git diff`, más las sondas de método (pathspec inexistente, fichero `??`,
once direcciones prohibidas contra la orden de prerrequisitos desde un fichero fuera del árbol). Todo en
verde; salida literal y veredicto en [`evidencia-cierre.md`](./evidencia-cierre.md), reescrito sobre la
base nueva y con el aviso del punto 1 de la Definition of Done retirado. Cifras de cobertura idénticas
a los intentos 1 y 2 (global 93,3 %, `internal/core/schema` 90,1 %, `internal/cli` 98,6 %,
`internal/httpx` 94,8 %). La orden `kitlegal-e2e consultar …` del escenario 11 vuelve a rechazarla el
envoltorio del agente, no el binario; se cubre con las mismas tres medidas que en el intento 1.

`docs/PENDIENTES.md` ya estaba actualizado desde el intento 1 (en la base de este intento); verificado
de nuevo, sin cambios.

## 5. Lo que cambia en este intento

- `internal/httpx/reintentos_test.go`: el arreglo (a), (b), (c).
- `gates/evidencia-cierre.md`: reescrito entero.
- `tasks.md`: T018 pasa a `[X]`. Ninguna otra tarea se toca.
- Este fichero: cabecera y esta sección.

Nada bajo `testdata/` ni `schemas/`; `.golangci.yml` no se toca (el comentario nuevo del test evita
«comparte», que `misspell` marcaría, y dice «se solapa»).

## 6. `make ci` final

Ejecutado sobre el árbol de trabajo al terminar, con el arreglo y todos los ficheros de arriba ya
escritos: «ci: todos los controles en verde». `golangci-lint` con `0 issues.`; suite con `-race
-shuffle=on` en `ok` en todos los paquetes (`internal/httpx` 94,8 %, `internal/cli` 98,6 %,
`internal/core/schema` 90,1 %, `internal/app` 92,8 %, `internal/render` 95,8 %); `govulncheck` sin
vulnerabilidades; `gitleaks` sin fugas; `go mod verify` y `go mod tidy -diff` limpios. Es la segunda
ejecución completa de `make ci` en este intento (la primera es la del escenario 12), y va acompañada de
las dos tiradas de mil sobre la subprueba y la tabla que antes fallaban: esta vez el verde no es el del
azar.

`rtk proxy git status --porcelain` al terminar:

```
 M internal/httpx/reintentos_test.go
 M specs/003-h2-internal-httpx-cliente/gates/evidencia-cierre.md
 M specs/003-h2-internal-httpx-cliente/gates/tarea-T018.md
 M specs/003-h2-internal-httpx-cliente/gates/tarea-actual.json
 M specs/003-h2-internal-httpx-cliente/gates/tareas-intentos.json
 M specs/003-h2-internal-httpx-cliente/tasks.md
```

Un fichero de test en una ruta declarada, cuatro ficheros del directorio del feature y los dos JSON de
estado del supervisor, que ya estaban modificados al empezar. Nada bajo `testdata/` ni `schemas/`
(`git status --porcelain` restringido a esas rutas: vacío).
