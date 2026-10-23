# T005: por qué no quedó en verde

## Intento 1: tarea mal delimitada (redelimitada, queda `[ ]`)

El trabajo de la tarea está hecho dentro de sus rutas y el árbol lo conserva para el intento siguiente:
`internal/evals/eleccion.go` y `internal/evals/eleccion_test.go` (sin commitear), más las líneas «T005» de
`gates/supuestos.md`.

`make ci` falla en **un único test, fuera de las rutas declaradas** y ajeno a lo que la tarea toca:
`TestEjecutarSesionesConUnError` (`internal/evals/sesiones_test.go`, de H7.3). Formato, lint y `test` (el paquete de
evals incluido, con el mismo código), en verde; el rojo sale en la segunda pasada, `test-integration`:

```
--- FAIL: TestEjecutarSesionesConUnError (15.57s)
    sesiones_test.go:1112:
        	Error Trace:	/Users/jorge/Projects/kitlegal/internal/evals/sesiones_test.go:1011
        	            				/Users/jorge/Projects/kitlegal/internal/evals/sesiones_test.go:1112
        	Error:      	Should be true
        	Messages:   	la sesion 01-lpac-articulo-21-claude-haiku-4-5-20251001-01 se abre
FAIL	github.com/jmorenobl/kitlegal/internal/evals	201.080s
```

Los objetivos de `make ci` a los que el intento no llegó (`test-tiempos`, `vuln`, `schema-check`, `skills-check`,
`goreleaser-check`, `secrets`, `mod-verify`, `mod-tidy-check`) no se han ejecutado en esta sesión: los correrá el
`make ci` del reintento.

### Causa

No es T005. `TestEleccionDeLosMutantes` tarda 0,27 s (`-v -race -tags=integration`) y no añade carga apreciable, y el
test que falla, aislado en este mismo árbol (`-race -tags=integration -count=10`), pasa 10 de 10.

Es una carrera del propio test con la carga de `make ci`. Con `Concurrencia` 2, la primera y la segunda sesión se abren
a la vez, y cada una crea su directorio, se prepara (`PrepararSesion`) y arranca su guion (`scripts/evals-sesion.sh`)
antes de que su `claude` sustituto anote su llegada (`claude/<sesión>` en el directorio común). La segunda duerme 1 s
y, al terminar, deja su hueco; la tercera lo ocupa, falla al crear `trabajo/` y cierra el reparto
(`repartidor.fallar` → `cerrar`). Si la primera llega a `ejecutarElGuion` con el reparto ya cerrado, vuelve con
`errSesionInterrumpida` sin arrancar su guion (y si ya lo arrancó, el TERM al grupo alcanza al guion antes de que
ejecute `claude`): la sesión no «llega», y el test la exige abierta y cerrada por TERM (`exigirCerrada`, línea 1011).
Basta con que la primera tarde algo más de 1 s que la segunda en llegar a su `claude`, y eso solo ocurre bajo la carga
de la pasada `test-integration`, con todos los paquetes con `-race` a la vez: el test tardó 15,57 s en el rojo, cuando
con `-count=10` aislado el paquete entero tarda 37 s.

Reproducido con un mutante sobre una copia desechable del árbol de HEAD (`git worktree` en `$TMPDIR`, fuera del
repositorio): en `abrirSesion` (`internal/evals/sesiones.go`), 2 s de `time.Sleep` antes de `ejecutarElGuion` para la
sesión cuyo nombre termina en `-01`. Con el test de hoy: `FAIL`, en la misma línea y con el mismo mensaje («la sesion
01-lpac-articulo-21-claude-haiku-4-5-20251001-01 se abre»), en 3,52 s.

No se repite `make ci` hasta un verde: sería cerrar con un verde probabilístico un rojo que puede volver en cualquier
`make ci` de lo que queda de hito (constitución, «Criterio de decisión autónoma», punto 1), como en T010 de H21
(`specs/015-h21-kitlegal-mcp-serve/gates/tarea-T010.md`). El arreglo está en dos ficheros que T005 no declara, así que
la tarea se redelimita.

### El arreglo

El test deja de depender de cuándo llega la primera sesión: la segunda no termina por su cuenta, sino cuando la
primera ha empezado su espera, que es cuando ya atiende TERM (la marca `espera-empezada` la anota quien duerme, tras su
`exec`). Así la tercera solo falla con la primera abierta y esperando, llegue cuando llegue; y la primera duerme 5 s,
como las dos de `TestEjecutarSesionesConElContextoCancelado`, que se cancelan nada más verse esa misma marca.

- `internal/evals/sustitutos_test.go`: el sustituto de `claude` lee `<sesión>.tras` del directorio de transcripts, una
  ruta, y espera, cada 50 ms, a que exista antes de empezar su espera (`sleep` con fracción lo admiten el de macOS y el
  de GNU; `sh` se lo pasa sin interpretarlo); el método `empezarLaEsperaCuandoExista(t, sesión, ruta)` lo escribe.
- `internal/evals/sesiones_test.go`: `TestEjecutarSesionesConUnError` con la espera de la base a 0, la primera sesión a
  5 s y la segunda tras `s.esperaEmpezada(primera)`; su comentario cuenta la carrera.

```diff
--- a/internal/evals/sesiones_test.go
+++ b/internal/evals/sesiones_test.go
@@ -1088,19 +1088,29 @@
 // cierra las abiertas con la secuencia del tope, como la interrupción, y vuelve
 // con ese error y no con el de las sesiones que cierra
 // (contracts/ejecucion-del-job.md §3 de H7.3; research D7 de H7.3; FR-037): con
-// Concurrencia 2, la primera sesión duerme 5 s, la segunda 1 s y la tercera no
-// se puede crear porque su directorio ya tiene trabajo/. La segunda termina, la
-// tercera no llega a abrirse y la primera recibe TERM sin cumplir su espera;
-// ninguna posterior a la tercera se abre ni tiene directorio, no queda ningún
-// proceso en el grupo de ninguna, y el error nombra la tercera sesión.
+// Concurrencia 2, la primera sesión duerme 5 s; la segunda no duerme, pero no
+// termina hasta que la primera ha empezado su espera; y la tercera no se puede
+// crear porque su directorio ya tiene trabajo/. La segunda termina, la tercera
+// no llega a abrirse y la primera recibe TERM sin cumplir su espera; ninguna
+// posterior a la tercera se abre ni tiene directorio, no queda ningún proceso en
+// el grupo de ninguna, y el error nombra la tercera sesión.
+//
+// La tercera solo se abre cuando la segunda deja su hueco, y su fallo cierra el
+// reparto: si la segunda terminara por su cuenta, bajo la carga de make ci la
+// primera, que se abre a la vez, puede no haber llegado aún a su claude —cada
+// sesión crea su directorio, se prepara y arranca su guion antes— y el cierre la
+// encontraría sin abrir, y el test la exigiría cerrada por TERM. La segunda
+// espera por eso a la marca con la que la primera empieza su espera, con la que
+// ya atiende TERM.
 func TestEjecutarSesionesConUnError(t *testing.T) {
 	t.Parallel()
 
 	s := escribirSustitutos(t)
-	ejecucion := sesionesDelRepartidor(t, s, 2, 1, false)
+	ejecucion := sesionesDelRepartidor(t, s, 2, 0, false)
 
-	primera, tercera := ejecucion.Plan[0].Nombre, ejecucion.Plan[2].Nombre
+	primera, segunda, tercera := ejecucion.Plan[0].Nombre, ejecucion.Plan[1].Nombre, ejecucion.Plan[2].Nombre
 	s.escribirEspera(t, primera, 5)
+	s.empezarLaEsperaCuandoExista(t, segunda, s.esperaEmpezada(primera))
 	require.NoError(t, os.MkdirAll(filepath.Join(ejecucion.Sesiones, tercera, "trabajo"), 0o700))
 
 	_, err := ejecutarSesiones(t.Context().Done(), ejecucion)
--- a/internal/evals/sustitutos_test.go
+++ b/internal/evals/sustitutos_test.go
@@ -34,7 +34,10 @@
 // anotada al empezarla y al cumplirla, duerme, se envía una señal y sale con el
 // código que le digan las variables KITLEGAL_SUSTITUTO_*. La espera de una
 // sesión con <sesión>.espera en el directorio de transcripts es la de ese
-// fichero. Con variableDeCalentar, termina con 0 sin hacer nada.
+// fichero, y la de una sesión con <sesión>.tras, que lleva una ruta, no empieza
+// hasta que esa ruta existe: así una sesión no termina antes de que otra haya
+// empezado su espera, llegue cuando llegue. Con variableDeCalentar, termina con
+// 0 sin hacer nada.
 //
 // Si no ignora TERM, quien duerme es un sh aparte, hijo de una subshell que
 // atiende TERM, lo anota y sale, y la shell, al recibirlo, sale cuando ha
@@ -90,6 +93,10 @@
 if [ -f "$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.espera" ]; then
 	espera=$(cat "$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.espera")
 fi
+if [ -f "$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.tras" ]; then
+	tras=$(cat "$KITLEGAL_SUSTITUTO_TRANSCRIPTS/$sesion.tras")
+	while [ ! -e "$tras" ]; do sleep 0.05; done
+fi
 if [ "${KITLEGAL_SUSTITUTO_IGNORA_TERM:-no}" = si ]; then
 	trap '' TERM
 	: > "$anotaciones/espera-empezada"
@@ -332,6 +339,14 @@
 	require.NoError(t, os.WriteFile(filepath.Join(s.transcripts, sesion+".espera"), []byte(strconv.Itoa(segundos)), 0o600))
 }
 
+// empezarLaEsperaCuandoExista deja dicho al sustituto de claude de la sesión
+// del nombre dado que no empiece su espera hasta que exista la ruta dada.
+func (s sustitutos) empezarLaEsperaCuandoExista(t *testing.T, sesion, ruta string) {
+	t.Helper()
+
+	require.NoError(t, os.WriteFile(filepath.Join(s.transcripts, sesion+".tras"), []byte(ruta), 0o600))
+}
+
 // abiertasAlLlegar es cuántas sesiones abiertas contó el sustituto de claude al
 // llegar a la sesión, la suya incluida.
 func (s sustitutos) abiertasAlLlegar(t *testing.T, sesion string) int {
```

### Comprobado sobre la copia desechable (fuera del repositorio)

Todo con `go -C <copia> test -race -tags=integration ./internal/evals/`:

- Test corregido **con el mutante** (la primera sesión 2 s tarde): `-count=3`, en verde (12,7 s).
- **Sin el mutante** (`sesiones.go` restaurado; `git diff --stat` de la copia nombra solo los dos ficheros de test):
  `-run '^TestEjecutarSesionesConUnError$' -count=10`, en verde (37 s); los tests del repartidor y de la sesión
  (`TestTopeDeLaSesion`, `TestInterrumpirLaSesion`, `TestEjecutarSesiones*`, `TestAbrirUnaSesion`) con `-count=3`, en
  verde (189 s); el paquete entero, `-count=1`, en verde (116 s).
- `golangci-lint run ./internal/evals/...` del repositorio (`go tool -modfile=tools/golangci-lint/go.mod`): 0
  hallazgos; `golangci-lint fmt --diff`: sin diferencias; `gofmt -l internal/evals/`: nada.

No se ha medido bajo la carga de un `make ci` completo ni en un ejecutor de Linux: con la espera por la marca, el test
ya no depende de cuándo llega la primera sesión; lo dirán el `make ci` del reintento y la CI de la propuesta de cambio.

### Redelimitación

La línea de T005 en `tasks.md` gana `internal/evals/sesiones_test.go` e `internal/evals/sustitutos_test.go` y describe
la corrección y el mutante que la comprueba; la tarea queda `[ ]` y el bucle la reintenta con la línea nueva. El
reintento aplica el diff de arriba, comprueba el mutante (y lo restaura antes de `make ci`) y repite `make ci` en primer
plano.

## Intento 2: en verde (queda `[X]`)

Sobre el árbol del intento 1 (`eleccion.go` y `eleccion_test.go` sin tocar: `TestEleccionDeLosMutantes` y
`TestReglasDeLaEleccion`, 43 subtests, en verde), el diff de arriba aplicado tal cual a los dos ficheros de test, y el
mutante comprobado en el propio árbol, en `abrirSesion` de `internal/evals/sesiones.go` (2 s de `time.Sleep` antes de
`ejecutarElGuion` para la sesión cuyo nombre termina en `-01`), todo con `-race -tags=integration`:

- Test de hoy con el mutante: `FAIL`, en `sesiones_test.go:1011` desde la 1112 (`exigirCerrada` de la primera sesión),
  como en el intento 1.
- Test corregido con el mutante: `-count=3`, en verde.
- `sesiones.go` restaurado con `git checkout` (`git status` nombra solo los cuatro ficheros de la tarea): el test
  corregido, `-count=5`, en verde; `TestTopeDeLaSesion`, `TestInterrumpirLaSesion`, `TestEjecutarSesiones*` y
  `TestAbrirUnaSesion`, `-count=2`, 42 en verde; `golangci-lint run` y `fmt --diff` sobre `internal/evals/`, sin
  hallazgos.

`make ci`: sus doce objetivos, en su orden y cada uno en primer plano, en tres órdenes consecutivas (`fmt-check lint
test`; `test-integration`; `test-tiempos vuln schema-check skills-check goreleaser-check secrets mod-verify
mod-tidy-check`), porque una sola orden de esta sesión tiene un tope de 10 minutos y `make ci` entero lo sobrepasa; el
objetivo `ci` no hace nada más que encadenarlos. Todos en verde; `internal/evals`, 148,6 s en `test`, y
`test-integration` entero sin ningún `FAIL`, con `TestEjecutarSesionesConUnError` dentro. No se ha medido en un
ejecutor de Linux: lo dirá la CI de la propuesta de cambio.
