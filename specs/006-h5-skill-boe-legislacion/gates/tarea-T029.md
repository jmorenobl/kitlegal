# T029 · intento 1 · detenida sin marcar

## Por qué se detuvo

La segunda orden del escenario 4 de `quickstart.md` está bloqueada en la sesión desatendida, también sola y tal cual:

```text
wc -l /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/SKILL.md
```

```text
wc in '/private/tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/SKILL.md' was blocked. For security, Claude
Code may only count lines/words/bytes in files from the allowed working directories for this session:
'/Users/jorge/Projects/kitlegal'.
```

La tarea manda detenerse ante una orden que pide aprobación, sin sustituirla por otra. El defecto es de la guía: su
tabla de formas daba `wc` por permitido con rutas de la carpeta temporal, y no lo está.

**Cómo se aisló.** El bloque entero del escenario 4, en una sola invocación, se rechazó con «This Bash command contains
multiple operations. The following part requires approval:» y el texto de casi todo el bloque, sin nombrar la orden.
Por eso cada orden se repitió sola:

- la primera (`rtk proxy perl -0pi -e '$n = () = /\n/g; …'`) pasó y dejó el `SKILL.md` del clon en 300 líneas;
- la segunda, sola, dio el bloqueo de arriba.

Tres sondas validan el arreglo. No son resultado del escenario:

- `rtk proxy wc -l` sobre el mismo fichero da `300 /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/SKILL.md`.
- El mismo programa de `perl` con `-0pe`, en tubería con `rtk proxy wc -l` y agrupado con un `echo`, pasa y da `300`: el
  rechazo del bloque venía solo de `wc`.
- Ninguna otra orden del quickstart usa la carpeta temporal sin `rtk proxy`: la búsqueda de las líneas con
  `/tmp/kitlegal-quickstart-h5` que no empiezan por `rtk proxy `, `git -C `, `git clone ` o `make -C ` solo da prosa y la
  línea bloqueada.

## Arreglo, en `quickstart.md`

Solo cambia la forma de invocación; no cambian tests, patrones, filtros ni código.

- **Escenario 4**: `rtk proxy wc -l /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/SKILL.md`. La mutación, el
  objetivo y lo esperado (`300` líneas y `código 2`) no cambian.
- **Tabla de formas**: una fila nueva, «Contar las líneas de un fichero de la carpeta temporal», con `rtk proxy wc -l`.
  El párrafo siguiente ya no pone `wc` entre las órdenes permitidas tal cual con rutas temporales; sobre ficheros del
  repositorio sigue valiendo (escenario 1).
- **Prerrequisitos**: una sonda nueva, `rtk proxy wc -l /tmp/kitlegal-quickstart-h5/repo/.agents/.gitattributes`, que
  debe dar `1` seguido de la ruta (el fichero tiene la única línea de research D18). Así, un bloqueo de la forma se ve
  antes de romper el clon.

## Lo ejecutado antes de parar

Sobre `be901f7`, el 2026-09-15 de madrugada. Nada de esto vale como evidencia del cierre: el intento siguiente empieza
por los prerrequisitos y lo repite todo. Los escenarios 1, 2, 7, 9, 10 y 11 se lanzaron a la vez, porque no comparten
ficheros (el 2 trabaja en el clon y los demás solo leen el árbol); después, el 3 y el 4, en orden.

- **`make ci`** (03:36:13 a 03:36:58): `0 issues.`, los paquetes en `ok` en los dos perfiles,
  `ci: todos los controles en verde` y `código 0`.
- **Cobertura con `go tool cover -func`**: total 95,3 % en `coverage.out` y 95,8 % en `coverage-integration.out`.
  - Sumas de sentencias del perfil unitario: global 5579/5860; `internal/core` 73/81 = 90,1 % en los dos perfiles;
    `internal/cli` 348/353 = 98,6 %; `internal/skills` 1134/1184 = 95,8 %; `internal/evals` 1365/1440 = 94,8 %;
    `internal/app` 445/507 = 87,8 %.
  - Unión de los dos perfiles: 5608/5860 = 95,7 %.
- **Prerrequisitos**: `go version go1.27.1 darwin/arm64`, rama `h5-skill-boe-legislacion`, estado solo
  `fin del estado`, clon sin mensajes, `código 0` y `sin HOME temporal todavía`.
- **1**: `make skills-check` con `ok` en los tres paquetes; 159 líneas; enlace `../../../bin/instalado/kitlegal`; solo
  `SKILL.md`, `references/normas.md` y `scripts/boe`; cabecera `1`.
- **2**: las dos regeneraciones en 0 y los dos estados solo `fin del estado`; en el `Makefile`, `skills-sync` sin anuncio
  de H5.
- **3**: los tres casos dan `código 2`.
  - (a) `boe-legislacion: references/normas.md: contenido-distinto`.
  - (b) Lo mismo, más `BOE-A-2015-10565: el título no coincide con la búsqueda grabada: …` de
    `TestIdentificadoresDeLasNormas`.
  - (c) `boe-legislacion: SKILL.md: contenido-distinto`.
  - Los subtests de `TestSkillsDelRepositorio` que copian el árbol fallan también, porque copian el defecto y esperan
    un árbol limpio. No contradice lo esperado.
- **4**: detenido en su segunda orden (arriba). No se ejecutaron los escenarios 5, 6 y 8.
- **7**: todo en `PASS`, sin ningún `FAIL`, con cada subtest que nombra la guía. Por test: `TestJuzgar` 18,
  `TestLeerSesion` 15, `TestLeerTrazas` 21, `TestInforme` 13, `TestInterpretarInvocacion` 12, `TestExtraerCitas` 5,
  `TestLeerTrazasSinFicheros` 2 y `TestEscribirInformeSinSusEntradas` 6 (este último incluye
  `informe-json-no-se-puede-escribir`).
- **9**: las cinco claves del frontmatter; los seis elementos del protocolo (líneas 30, 36, 54, 95, 99 y 101); la cita
  de ejemplo en la línea 112; y la búsqueda final, solo `fin de la búsqueda`. Revisión de SC-004 leyendo `## Reglas`:
  las reglas 1 a 4 son FR-006, FR-010, FR-011 y FR-015, 4 de 4.
- **10**: solo `fin de la búsqueda`.
- **11**: todo lo esperado.
  - `set` para los dos atributos en `.agents/skills/golang-how-to/SKILL.md` y `unspecified` en los otros dos.
  - El diff, solo `fin del diff`, y `0` entradas «En H5».
  - Los tres directorios en `README.md` (líneas 151-153), y las cuatro órdenes en los tres documentos.
  - Formato 2/1/1 y job 3/2/2.
  - Diez `1` y nueve `0`; las cuatro filas de `README.md` y el comentario de `make install`, `1`; la frase de
    `make help`, `0`.
  - Las cuatro filas de `CONTRIBUTING.md`, `1`; la tabla posterior, `0`; el párrafo, `1`; la entrada del
    `CHANGELOG.md`, `1`.
- **Limpieza**, tras detenerse: la carpeta temporal (con el `SKILL.md` del clon aún en 300 líneas) se borró, y el
  estado del árbol dio solo `fin del estado`.
- **Datos para `pr-h5.md`** (no se escribió: exige la guía entera):
  - Frente a `main`: 355 ficheros (305 fuera de `specs/`, con 20 254 líneas añadidas y 50 retiradas).
  - `go.mod` solo pasa `go.yaml.in/yaml/v3 v3.0.5` de indirecta a directa; ni `go.sum` ni `codecov.yml` aparecen en el
    diff.
  - `.golangci.yml`: solo la etiqueta `evals` y las cinco palabras de D20.
  - `0` líneas `//nolint` y `0` `t.Skip` añadidas en ficheros `.go`.

## Verificación al cerrar el intento

`make ci` tras corregir el quickstart y escribir esta nota (03:45:07 a 03:45:54): `0 issues.`, los paquetes en `ok` en
los dos perfiles, `gitleaks` «no leaks found», `ci: todos los controles en verde` y `código 0`. Cambios en
`git status --porcelain`, todos dentro del directorio del feature:

- `quickstart.md` y esta nota;
- `gates/tarea-actual.json` y `gates/tareas-intentos.json`, del workflow.

La tarea queda `[ ]`: la guía no se completó, y `gates/pr-h5.md` no existe todavía.

## Para el intento siguiente

- Empezar por los prerrequisitos con el quickstart corregido y ejecutar cada orden tal cual. Si un bloque de varias
  órdenes se rechaza entero, repetir cada orden sola antes de concluir cuál pide aprobación.
- `gates/pr-h5.md` lo escribe el intento que complete la guía, con sus propias medidas.

# T029 · intento 2 · en verde

Sobre `be901f7`, el 2026-09-15 entre las 03:48 y las 03:53 (hora de Madrid), con el `quickstart.md` corregido en el
intento 1 todavía sin confirmar en el árbol. La guía se ejecutó entera desde los prerrequisitos: `make ci`, los
prerrequisitos, los escenarios 1 a 11 y la limpieza, cada orden tal cual la escribe el quickstart. **Ninguna orden pidió
aprobación**: la sonda nueva de los prerrequisitos (`rtk proxy wc -l …/.agents/.gitattributes`) dio `1` seguido de la
ruta, y la segunda orden del escenario 4, con `rtk proxy wc -l`, dio `300`. Cada bloque de un escenario se lanzó en una
sola invocación; los escenarios 1, 2, 7, 9, 10 y 11 a la vez (no comparten ficheros) y después 3, 4, 5, 6 y 8 en orden
sobre el clon, con una sonda de `git status --porcelain` del clon antes de cada uno, que dio siempre vacío.

Todo lo esperado se cumplió; los resultados, los negativos, la salida literal del escenario 11 y la cobertura están en
`gates/pr-h5.md`, escrito por este intento. Cuando la salida de un escenario negativo superó el tamaño que la
herramienta muestra, el veredicto (`código`, mensajes de fallo y líneas `PASS`) se leyó con `rtk proxy grep` sobre la
copia que la herramienta guarda; nada se redirigió a ficheros.

## Verificación al cerrar el intento

`make ci` tras escribir `gates/pr-h5.md` y esta nota (terminado a las 03:59): `0 issues.`, los paquetes en `ok` en los
dos perfiles, `govulncheck` sin vulnerabilidades, `schema-check` y `skills-check` en `ok`, `gitleaks` «no leaks found»,
`go mod verify` en la raíz y los cuatro módulos de herramienta, `tidy -diff`, `ci: todos los controles en verde` y
`código 0`. Cambios en `git status --porcelain`, todos dentro del directorio del feature: `quickstart.md` (del intento
1), `gates/pr-h5.md` (nuevo), esta nota, `tasks.md` (T029 marcada) y `gates/tarea-actual.json` y
`gates/tareas-intentos.json`, del workflow. Ningún fichero fuera del directorio del feature cambió (FR-013, FR-044,
FR-083, FR-085 y la tarea).

La tarea queda `[X]`.

# T029 · intento 2 · verificación en rojo tras marcar: redelimitada

El paso `verificar` del workflow (`make ci` sobre el árbol, con T029 ya marcada y `pr-h5.md` escrito) cayó el 2026-09-15
hacia las 04:00 en `internal/httpx` › `TestReintentosDosErroresYUnAcierto`:

```text
reintentos_test.go:79: "85.3865ms" is not greater than or equal to "90ms"
Messages: cada reintento espera su turno en el sitio: el decorador va por encima del ritmo (FR-021)
```

El `make ci` del propio intento, minutos antes, había salido en verde. El fallo es intermitente y no lo introduce H5:
el fichero es de H2 (`012c235`) y H3/T018 ya midió ≈ 3 por mil en otra aserción del mismo fichero.

## Causa raíz

La aserción compara **cada par de llegadas consecutivas al servidor** con «intervalo menos holgura» (100 ms − 10 ms).
Pero lo que el limitador (`x/time/rate`, cubo de un token) fija con exactitud son los **turnos**, anclados al primero:
el turno k no llega antes de k intervalos tras el turno cero (el del permiso del sitio). La llegada al servidor añade a
cada turno lo que tarde en despacharse —el despertar tras el temporizador, que bajo `-race` y con todos los paquetes en
paralelo puede retrasarse decenas de milisegundos; el transporte; la conexión, que la primera abre y las demás
reutilizan—, y ese añadido no es igual en todas. Si la llegada k−1 se despacha 15 ms tarde y la k puntual, el par mide
85 ms sin que el ritmo haya fallado: la aserción exige más que el contrato. La del ritmo (`ritmo_test.go`, 25 ms de
holgura sobre 150 ms) tiene el mismo defecto, solo con más margen.

**Reproducción determinista**, sobre un clon desechable de `be901f7` en `/tmp/kitlegal-reparar-t029/clon` (borrado al
terminar), con un mutante del decorador de ritmo que duerme 15 ms tras el turno del **primer** despacho del recurso de
cada cliente (una implementación correcta: nunca emite antes de su turno):

- aserción antigua de los reintentos: cae **10 de 10** (`81.8ms` … `85.6ms`, la misma cifra que la verificación);
- la del ritmo, con 25 ms de holgura, pasa.

La reproducción del rojo sin mutante no sale: 150 vueltas de `-race` con carga de otros paquetes en paralelo, todas en
verde. Es un fallo raro; el mutante es la evidencia de por qué ocurre.

## Por qué no lo arregla T029

Las rutas congeladas de T029 (`gates/tarea-actual.json`) son `.in/yaml/v3`, `go.yaml`, `quickstart.md` y
`gates/pr-h5.md` —las dos primeras, restos del extractor sobre `go.yaml.in/yaml/v3`—, y su propia línea manda que no
modifique ningún fichero fuera del directorio del feature (FR-013, FR-044, FR-083, FR-085). El paso `reparar` no puede
ampliar rutas: el guardián de la reparación las lee del mismo JSON. Reintentar `make ci` hasta el verde cerraría el hito
con el test inestable dentro (lo que H3/T018 pidió no hacer). Se aplica el procedimiento de H3/T018 y H4/T037: **tarea
nueva antes de T029** con las rutas del arreglo, y T029 vuelve a `[ ]`.

## El arreglo, comprobado sobre el clon

Solo dos ficheros, `internal/httpx/reintentos_test.go` e `internal/httpx/ritmo_test.go`, y solo sus aserciones sobre las
llegadas; ni el decorador, ni el cubo, ni la cadena, ni otro test. En los dos, cada llegada se mide contra un
`comienzo := time.Now()` tomado **antes** de la primera petición de la operación, sin holgura: la llegada i no se
adelanta a su turno, i intervalos tras el comienzo (el permiso del sitio es el turno cero; en los reintentos, el intento
n va al menos n intervalos por detrás). La cota es exacta —la fija el limitador, no el reloj de la máquina— y es la
misma que ya usaba el test del ritmo para la duración total.

```diff
--- a/internal/httpx/reintentos_test.go
+++ b/internal/httpx/reintentos_test.go
@@ -47,20 +47,25 @@ func TestReintentosDosErroresYUnAcierto(t *testing.T) {
 func dosErroresYUnAcierto(t *testing.T) []time.Duration {
 	t.Helper()

-	// El intervalo es el que separa las llegadas. La holgura es la del test del
-	// ritmo, dos órdenes de magnitud por debajo: lo que el limitador espacia con
-	// exactitud es el turno, y la llegada le añade lo que tarde en despacharse
-	// el manejador.
+	// El intervalo es el que separa los turnos del sitio. Lo que el limitador
+	// fija con exactitud es el turno; la llegada al servidor le añade lo que
+	// tarde en despacharse, y ese añadido no es igual en todas —un despertar
+	// tardío de la máquina, la conexión que la primera abre y las demás
+	// reutilizan—, así que dos llegadas consecutivas pueden acercarse por
+	// debajo del intervalo sin que el ritmo haya fallado. Por eso cada llegada
+	// se mide contra el comienzo de la operación, anterior al primer turno, y
+	// no contra la llegada anterior: esa cota es exacta, la fija el limitador y
+	// no el reloj de la máquina, y no lleva holgura.
 	const intervalo = 100 * time.Millisecond

-	const holgura = 10 * time.Millisecond
-
 	anotador := &llegadas{}
 	esperas := &esperasAnotadas{}
 	servidor, contador := servidorIdentificado(t,
 		anotador.anota(servidorQueFalla(2, http.StatusServiceUnavailable)))
 	cliente := clienteDePrueba(t, ConIntervalo(intervalo), conReloj(esperas.reloj))

+	comienzo := time.Now()
+
 	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
 		Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
@@ -75,9 +80,13 @@ func dosErroresYUnAcierto(t *testing.T) []time.Duration {
 	anotadas := anotador.instantes()
 	require.Len(t, anotadas, 3)

-	for intento := 1; intento < len(anotadas); intento++ {
-		assert.GreaterOrEqual(t, anotadas[intento].Sub(anotadas[intento-1]), intervalo-holgura,
-			"cada reintento espera su turno en el sitio: el decorador va por encima del ritmo (FR-021)")
+	// El robots.txt del sitio ocupa el turno cero, así que el intento n no
+	// puede llegar antes de n veces el intervalo tras el comienzo. Si los reintentos no
+	// pasaran por el limitador, el segundo y el tercero llegarían con el
+	// primero, a un solo intervalo del comienzo.
+	for intento, llegada := range anotadas {
+		assert.GreaterOrEqual(t, llegada.Sub(comienzo), time.Duration(intento+1)*intervalo,
+			"el intento %d espera su turno en el sitio: el decorador va por encima del ritmo (FR-021)", intento+1)
 	}
--- a/internal/httpx/ritmo_test.go
+++ b/internal/httpx/ritmo_test.go
@@ -23,11 +23,11 @@ import (
 // La medida empieza en la primera petición que el cliente hace al sitio, la de
 // su robots.txt, que ocupa el primer turno del cubo. Si el cubo admitiera más de
-// un token sería justo el primer par —robots.txt y recurso— el que saldría sin
-// separación, y medir solo a partir del segundo lo dejaría pasar: por eso se
-// anota también esa llegada y se exige la separación de **todos** los pares
-// consecutivos, y por eso la operación entera tiene que ocupar el intervalo
-// tantas veces como peticiones se hacen (D8).
+// un token sería justo la primera petición tras el robots.txt la que saldría
+// sin esperar, y medir solo a partir de la segunda lo dejaría pasar: por eso se
+// anota también esa llegada y se exige que **ninguna** se adelante a su turno,
+// contado desde el comienzo, y por eso la operación entera tiene que ocupar el
+// intervalo tantas veces como peticiones se hacen (D8).
 func TestRitmoSeparaPeticionesDelMismoSitio(t *testing.T) {
@@ -68,17 +68,18 @@ func TestRitmoSeparaPeticionesDelMismoSitio(t *testing.T) {
 	require.Len(t, anotadas, peticiones+1, "el robots.txt y las tres peticiones")

 	// Lo que el limitador espacia con exactitud es el turno; la llegada añade a
-	// cada turno lo que tarde en despacharse el manejador, y la primera —la del
-	// robots.txt— paga además la apertura de la conexión, que las siguientes
-	// reutilizan y que se resta de la separación del primer par sin ser ritmo.
-	// De ahí la holgura, muy por debajo del intervalo: sin limitador, con un
-	// cupo compartido entre sitios o con ráfaga 2, la separación de algún par no
-	// sería «un intervalo menos una holgura» sino prácticamente cero.
-	const holgura = 25 * time.Millisecond
-
-	for i := 1; i < len(anotadas); i++ {
-		assert.GreaterOrEqual(t, anotadas[i].Sub(anotadas[i-1]), intervalo-holgura,
-			"las llegadas %d y %d al mismo sitio van separadas al menos el intervalo (SC-004)", i-1, i)
+	// cada turno lo que tarde en despacharse, y ese añadido no es igual en
+	// todas: la primera —la del robots.txt— paga además la apertura de la
+	// conexión, que las siguientes reutilizan, y cualquiera puede pagar un
+	// despertar tardío de la máquina. Dos llegadas consecutivas pueden por eso
+	// acercarse por debajo del intervalo sin que el ritmo haya fallado. Lo que
+	// sí es exacto es que la llegada i no se adelanta a su turno, que va i
+	// veces el intervalo por detrás del comienzo: la cota la fija el limitador,
+	// no el reloj de la máquina, y no lleva holgura. Sin limitador o con ráfaga
+	// 2, alguna llegada se adelantaría a su turno.
+	for i, llegada := range anotadas {
+		assert.GreaterOrEqual(t, llegada.Sub(comienzo), time.Duration(i)*intervalo,
+			"la llegada %d al sitio no se adelanta a su turno, a %d × intervalo del comienzo (SC-004)", i, i)
 	}
```

`misspell` marca `intervalos` (→ *intervals*) en comentarios y mensajes de ficheros Go: por eso el diff dice «n veces el
intervalo» y «a %d × intervalo». Ninguna palabra nueva del diff entra en `.golangci.yml`.

**Resultados sobre el clon**, con `go -C` y el linter fijado por el repositorio:

| Sonda | Resultado |
|---|---|
| `golangci-lint fmt --diff` y `run` sobre `./internal/httpx/` | `0 issues.`, sin diff de formato |
| `go test -race -shuffle=on ./internal/httpx/` | `ok` |
| 20 vueltas `-race` de los dos tests arreglados | `ok` |
| Mutante «despacho tardío» (+15 ms tras el turno del primer despacho del recurso) | aserción antigua: cae 10/10; **nueva: pasa 10/10** |
| Mutante «reintentos por debajo del ritmo» (cadena con el ritmo por encima de los reintentos) | cae: intentos 2 y 3 a ~103 ms del comienzo, frente a 200 y 300 ms |
| Mutante «ráfaga de dos tokens» | caen los dos tests: primera llegada a ~1,4 ms frente a 100 y 150 ms; el total, 303 ms frente a 450 |
| Mutante «sin espera de turno» | caen los dos tests: todas las llegadas por debajo de 3 ms |
| `make -C <clon> ci` con el arreglo y sin mutantes | `código 0`, `0 issues.`, los paquetes en `ok` en los dos perfiles, «no leaks found», `ci: todos los controles en verde` |

Cada mutante se aplicó solo, con los demás revertidos; el clon quedó con solo los dos ficheros de test modificados y se
borró al terminar.

## Lo que cambia en este intento, todo dentro del directorio del feature

- `tasks.md`: línea nueva **T032** antes de T029 (Phase 11), con las dos rutas del arreglo y el arnés de esta nota como
  criterio de comprobación; T029 vuelve a `[ ]`; en *Notas*, la redelimitación y los dos ficheros existentes que T032
  toca. El extractor del workflow sobre la línea de T032 devuelve solo `gates/tarea-T029.md`,
  `internal/httpx/reintentos_test.go` e `internal/httpx/ritmo_test.go`.
- `gates/pr-h5.md`: nota de cabecera que lo declara superado; el intento 3 lo reescribe.
- Esta nota.

Nada fuera del directorio del feature cambia: el árbol sigue en `be901f7` con las mismas diferencias que dejó el
intento 2 (`quickstart.md`, `gates/`, `tasks.md`).

## Para el intento 3 de T029

- Corre después del commit de T032 (`feat(H5): T032`). Empieza por los prerrequisitos y repite la guía entera; lo del
  intento 2 no vale como evidencia porque el árbol cambia.
- Reescribe `gates/pr-h5.md` fechado por el commit nuevo, con la cobertura medida otra vez, y nombra T032 en «controles
  añadidos» o «decisiones»: la medida del ritmo pasa a la cota exacta del limitador, sin holgura.
- Es el tercer y último intento (`n > 3` detiene el run).

# T029 · intento 3 · en verde

Sobre `536359c` (`feat(H5): T032`), el 2026-09-15 entre las 04:28 y las 04:35 (hora de Madrid), con el árbol limpio
fuera del directorio del feature (dentro, solo los ficheros de estado del workflow). La guía se ejecutó entera desde
los prerrequisitos: `make ci`, los prerrequisitos, los escenarios 1 a 11 y la limpieza, cada orden tal cual la escribe
el `quickstart.md` ya confirmado (el que corrigió el intento 1). **Ninguna orden pidió aprobación** y ninguna sonda dio
algo distinto de lo esperado: la sonda de `rtk proxy wc -l` de los prerrequisitos dio `1` seguido de la ruta, y la
segunda orden del escenario 4 dio `300`. Como en el intento 2, cada bloque de un escenario se lanzó en una sola
invocación; los escenarios 1, 2, 7, 9, 10 y 11 a la vez y después 3, 4, 5, 6 y 8 en orden sobre el clon, con una sonda
de `git status --porcelain` del clon antes de cada uno, que dio siempre vacío. Una única orden auxiliar ajena a la guía,
de solo lectura: `rtk proxy test ! -e /tmp/kitlegal-quickstart-h5` tras la limpieza, para dejar constancia de que la
carpeta desapareció.

Todo lo esperado se cumplió; los resultados, los negativos, la salida literal del escenario 11 y la cobertura están en
`gates/pr-h5.md`, reescrito entero por este intento y fechado por `536359c`, con T032 en «Controles añadidos» y en
«Decisiones» y con la cabecera «SUPERADO» retirada. Las cifras de cobertura coinciden con las del intento 2 (T032 no
cambia ninguna sentencia de producto): global 95,3 % (`-func`, perfil unitario) y 95,8 % (perfil de integración),
`internal/core` 90,1 % (73/81), `internal/cli` 98,6 % (348/353), `internal/httpx` 97,2 % (792/815). Frente a `main`, el
diff pasa a 360 ficheros (307 fuera de `specs/`, 20 295 líneas añadidas y 77 retiradas) y 34 commits; los dos ficheros
de test de `internal/httpx` son los únicos de H0-H4 tocados que el intento 2 no listaba. Cuando la salida de un
escenario negativo superó el tamaño que la herramienta muestra, el veredicto (`código`, mensajes de fallo y líneas
`PASS`) se leyó con `rtk proxy grep` sobre la copia que la herramienta guarda; nada se redirigió a ficheros.

Lo que este intento no puede comprobar, y por qué: la pausa `[datos]` y la plataforma no intervienen (la tarea no lleva
ninguna de las dos etiquetas); `gh` y `git push` no se usaron. Los supuestos S8 y los de plataforma de research D22
quedan en `pr-h5.md` como pendientes de T030 y T031.

## Verificación al cerrar el intento

`make ci` en primer plano tras escribir `gates/pr-h5.md` y esta nota (terminado a las 04:39): `0 issues.`, los doce
paquetes en `ok` en los dos perfiles (`internal/httpx` 97,2 %, con los tests de T032), `govulncheck` sin
vulnerabilidades, `schema-check` y `skills-check` en `ok`, `gitleaks` «no leaks found», `go mod verify` en la raíz y
los cuatro módulos de herramienta, `tidy -diff`, `ci: todos los controles en verde` y `código 0`. Cambios en
`git status --porcelain`, todos dentro del directorio del feature: `gates/pr-h5.md` (reescrito), esta nota, `tasks.md`
(T029 marcada) y `gates/tarea-actual.json` y `gates/tareas-intentos.json`, del workflow. Ningún fichero fuera del
directorio del feature cambió (FR-013, FR-044, FR-083, FR-085 y la tarea); la carpeta temporal no existe.

La tarea queda `[X]`.
