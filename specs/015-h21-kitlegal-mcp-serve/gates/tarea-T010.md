# T010: por qué no quedó en verde

## Intento 1: tarea mal delimitada (redelimitada, queda `[ ]`)

El trabajo de la tarea está hecho dentro de sus rutas y el árbol lo conserva para el intento siguiente: el esquema de
eval, `Eval.SinBinarioNiServidor`, las reglas del conjunto, las dos evals nuevas, el tope de 240 minutos, sus tests y
cuatro líneas «T010» en `gates/supuestos.md` (más la de esta redelimitación). El guardián de diff lo dio por bueno
(«T010, 12 ficheros»).

`make ci` falla en **un único test, fuera de las rutas declaradas** y ajeno a lo que la tarea toca: el caso `abierta`
de `TestInterrumpirLaSesion` (`internal/evals/sesiones_test.go`, de H7.3). Formato, lint y `test` (el paquete de evals
incluido, con el mismo código), en verde; el rojo sale en la segunda pasada, `test-integration`:

```
--- FAIL: TestInterrumpirLaSesion (0.00s)
    --- FAIL: TestInterrumpirLaSesion/abierta (4.14s)
        sesiones_test.go:200:
            	Error:      	Not equal:
            	            	expected: true
            	            	actual  : false
            	Messages:   	el sustituto recibe TERM
FAIL	github.com/jmorenobl/kitlegal/internal/evals	60.373s
```

Los objetivos de `make ci` a los que el intento no llegó (`test-tiempos`, `vuln`, `schema-check`, `skills-check`,
`goreleaser-check`, `secrets`, `mod-verify`, `mod-tidy-check`), ejecutados aparte sobre este árbol: en verde (código 0).

### Causa

No es T010. Es una carrera del `claude` sustituto (`sustitutoDeClaude`, `internal/evals/sustitutos_test.go`) con el
`sh` de macOS, que es bash 3.2. En la rama que atiende TERM, la subshell anota `espera-empezada`, lanza `sleep` en
segundo plano y lo espera con `wait`. El test interrumpe la sesión en cuanto ve la marca, y TERM llega a todo el grupo
cuando la subshell puede estar aún entre la marca y el `wait`. bash 3.2 solo interrumpe `wait` si ya ha empezado a
esperar: un TERM que llega cuando la orden `wait` ya ha comenzado pero aún no espera se queda pendiente hasta que
termina el proceso esperado. Medido aparte, con un guion de cinco líneas (`trap` de TERM, `sleep 3 &` y
`wait $(kill -TERM $$; echo $!)`, que hace llegar la señal justo ahí): bash 3.2 atiende el trap a los 3 s, y dash, a
los 0 s.

Con el TERM pendiente en la subshell caben dos finales, y los dos dejan la sesión sin `term-recibido`:

- `sleep` acaba de nacer y aún no ha hecho su exec: la señal la recibe con el manejador heredado y se pierde, nadie
  termina y hace falta el KILL del margen. Es lo que se ve casi siempre.
- `sleep` ya la recibe y muere: `wait` vuelve con 143 y, con `set -e`, la subshell sale sin ejecutar su trap.

Aislado no se reproduce: `go test -race -tags=integration -run '^TestInterrumpirLaSesion$' -count=300`, 300 de 300 en
verde. Hace falta carga. Con un arnés de un solo uso fuera del repositorio (un programa Go que ejecuta el sustituto en
su grupo de procesos con el entorno de una sesión, espera la marca, envía TERM al grupo, da 1 s de margen antes del
KILL y mira `term-recibido`), 16 ensayos a la vez:

| Sustituto | `sh` | Cómo se mira la marca | Sin `term-recibido` | Con KILL |
|---|---|---|---|---|
| el de hoy | bash 3.2 | cada 10 ms, como el test | 15 de 3000 | 15 |
| el de hoy | bash 3.2 | cada 100 µs y hasta 2 ms al azar antes de TERM | 46 de 3000 y 32 de 3000 | 43 y 32 |
| el de hoy, de uno en uno | bash 3.2 | cada 100 µs y hasta 2 ms al azar | 1 de 1000 | 0 |
| el de hoy | dash de macOS | cada 100 µs y hasta 2 ms al azar | 0 de 3000 | 0 |
| el corregido | bash 3.2 | cada 10 ms | 0 de 3000 | 0 |
| el corregido | bash 3.2 | cada 100 µs y hasta 2 ms al azar | 0 de 3000 | 0 |
| el corregido | dash de macOS | cada 100 µs y hasta 2 ms al azar | 0 de 3000 | 0 |

Con dash el sustituto de hoy no pierde la marca: la intermitencia es de los `make ci` locales en macOS. En un
ejecutor de Linux no se ha medido nada en esta sesión (no hay con qué); lo dirá la CI de la propuesta de cambio.

No se repite `make ci` hasta un verde: sería cerrar con un verde probabilístico un rojo que puede volver en cualquier
`make ci` de lo que queda de hito (constitución, «Criterio de decisión autónoma», punto 1). El arreglo está en un
fichero que T010 no declara, así que la tarea se redelimita.

### El arreglo

El sustituto deja de depender de cuándo llega la señal:

- **Todo en primer plano, sin `&` ni `wait`.** Una shell atiende un trap pendiente en cuanto termina la orden que
  tiene en marcha, llegue la señal cuando llegue; no hay ventana.
- **Quien duerme anota la marca.** Es un `sh -c` aparte, hijo de la subshell, que escribe `espera-empezada` y hace
  exec de `sleep`: cuando la marca existe, ese proceso ya ha hecho su exec y no tiene manejador de TERM, así que TERM
  lo termina siempre. No queda ningún `sleep` huérfano en el grupo.
- **Los errores no se tragan.** `|| exit $?` en la subshell y en la shell: con un TERM pendiente, el trap se ejecuta
  antes que ese `exit`; sin él, un fallo de quien duerme sale con su código, como hoy con `set -e`.

Lo que las marcas dicen no cambia: `term-recibido` lo escribe solo la subshell, otro proceso del grupo; con TERM solo
al líder, la subshell sigue esperando y hace falta el KILL (el arnés, con TERM solo al líder: 64 de 64 sin
`term-recibido` y con KILL, en bash 3.2 y en dash). La rama que ignora TERM no se toca.

Un efecto que antes era ocasional y ahora es fijo: la subshell deja en la salida de error de la sesión cortada la línea
con la que la shell dice que su hijo terminó por una señal (`Terminated: 15` en macOS). El sustituto de hoy ya la
dejaba con dash en 509 de 3000 ensayos. Ningún test mira la salida de error de una sesión cortada por TERM.

### Comprobado sobre una copia desechable del árbol (fuera del repositorio)

Con el diff de abajo aplicado a una copia de este árbol:

- `go test -race -tags=integration -run '^(TestTopeDeLaSesion|TestInterrumpirLaSesion|TestEjecutarSesionesConElContextoCancelado|TestEjecutarSesionesConUnError)$' -count=40 ./internal/evals/`:
  en verde (145 s).
- El paquete entero, `go test -race -shuffle=on -tags=integration -count=3 ./internal/evals/`: en verde (139 s).
- `golangci-lint run ./internal/evals/...` del repositorio: 0 hallazgos; `golangci-lint fmt --diff`: sin diferencias.
- **Mutante 1**, la interrupción envía TERM solo al líder (`syscall.Kill(pid, …)` en `esperarConElMargen`):
  `TestInterrumpirLaSesion/abierta` falla por «el sustituto recibe TERM» y `TestEjecutarSesionesConElContextoCancelado`
  por «la sesion … recibe TERM», en sus dos sesiones.
- **Mutante 2**, el tope envía TERM solo al líder (en `orden.Cancel`): `TestTopeDeLaSesion/basta-term` falla por «el
  sustituto recibe TERM» y por el código, 137 en vez de 124.

### Redelimitación

La línea de T010 en `tasks.md` gana `internal/evals/sustitutos_test.go` y describe la corrección; comprobado con el
filtro de rutas de `scripts/workflow/tarea.sh`: las 16 rutas de antes más esa. La tarea vuelve a `[ ]`. La cabecera de
`tasks.md`, donde dice qué lleva cada tarea `[datos]`, dice ahora que T010 lleva además esa corrección de test y ningún
código de producto.

### Lo que tiene que hacer el intento 2

Lo de T010 ya está en el árbol y no se repite. Solo esto, en `internal/evals/sustitutos_test.go`:

```diff
@@ -36,14 +36,21 @@
 // sesión con <sesión>.espera en el directorio de transcripts es la de ese
 // fichero. Con variableDeCalentar, termina con 0 sin hacer nada.
 //
-// Si no ignora TERM, duerme en una subshell que lo atiende, lo anota y sale, y
-// la shell, al recibirlo, espera a la subshell antes de salir: la anotación
-// dice que TERM llegó a otro proceso del grupo, y si llega solo a la shell, la
-// subshell sigue durmiendo y la sesión no sale hasta el KILL. Las dos esperan
-// con wait, que TERM interrumpe, y la subshell anota que empieza la espera
-// cuando ya lo atiende: TERM no puede caer entre el fork y el exec de sleep,
-// donde la shell hija aún tiene su manejador y la señal se pierde. La marca de
-// abierta se retira antes de salir, también con TERM; con KILL queda.
+// Si no ignora TERM, quien duerme es un sh aparte, hijo de una subshell que
+// atiende TERM, lo anota y sale, y la shell, al recibirlo, sale cuando ha
+// terminado la subshell: la anotación dice que TERM llegó a otro proceso del
+// grupo, y si llega solo a la shell, la subshell sigue esperando y la sesión no
+// sale hasta el KILL. Todo va en primer plano, porque una shell atiende la señal
+// en cuanto termina la orden que tiene en marcha, llegue cuando llegue, y wait
+// solo se interrumpe si ya ha empezado a esperar: con sleep en segundo plano y
+// wait, un TERM que caía justo antes se quedaba sin atender hasta que terminaba
+// sleep, y con 16 sustitutos a la vez 15 de 3000 no anotaban el TERM. Y la marca
+// de que empieza la espera la anota quien duerme, tras su exec y sin manejador
+// de TERM: desde que existe, TERM lo termina siempre, y no puede caer entre el
+// fork y el exec, donde la shell hija aún tiene el manejador de su madre y la
+// señal se pierde. La subshell deja dicho en la salida de error que su hijo
+// terminó por una señal. La marca de abierta se retira antes de salir, también
+// con TERM; con KILL queda.
 const sustitutoDeClaude = `#!/bin/sh
 if [ "${KITLEGAL_SUSTITUTO_CALENTAR:-}" = si ]; then exit 0; fi
 set -eu
@@ -88,14 +95,11 @@
 	: > "$anotaciones/espera-empezada"
 	sleep "$espera"
 else
-	trap 'wait; rmdir "$abierta"; exit 143' TERM
+	trap 'rmdir "$abierta"; exit 143' TERM
 	(
 		trap ': > "$anotaciones/term-recibido"; exit 143' TERM
-		: > "$anotaciones/espera-empezada"
-		sleep "$espera" &
-		wait $!
-	) &
-	wait $!
+		/bin/sh -c ': > "$1"; exec sleep "$2"' sh "$anotaciones/espera-empezada" "$espera" || exit $?
+	) || exit $?
 fi
 : > "$anotaciones/espera-cumplida"
 rmdir "$abierta"
@@ -387,8 +391,9 @@
 	return existe(t, filepath.Join(s.comun, sustitutoClaude, sesion, "espera-cumplida"))
 }
 
-// recibioTERM dice si TERM llegó en la sesión a la subshell en la que duerme el
-// sustituto de claude, otro proceso de su grupo, y la subshell lo atendió.
+// recibioTERM dice si TERM llegó en la sesión a la subshell que espera a quien
+// duerme en el sustituto de claude, otro proceso de su grupo, y la subshell lo
+// atendió.
 func (s sustitutos) recibioTERM(t *testing.T, sesion string) bool {
 	t.Helper()
```

Las líneas del guion van sangradas con tabuladores, como las de su alrededor. Ningún otro byte de código ni de tests:
ni `sesiones_test.go`, ni las cotas, ni la rama que ignora TERM. Después:

1. Los cuatro tests de arriba con `-race -tags=integration -count=40`, en verde.
2. El mutante 1 un momento (TERM solo al líder en `esperarConElMargen`): el caso `abierta` falla por «el sustituto
   recibe TERM». Se restaura y se comprueba con `git diff --stat` que `internal/evals/sesiones.go` no cambia.
3. `make ci` en primer plano y marcar `[X]`.

Las dos verificaciones propias de T010 se repitieron en esta sesión sobre la copia, con el árbol de la tarea, y no
hace falta volver a hacerlas (el código que ejercen no cambia en el intento 2):

- Con la regla `sin binario ni servidor` quitada de los dos juegos, `TestConjuntoDeEvals` falla en los casos que la
  ejercen: `sin-binario-ni-servidor-ninguna` y `sin-binario-ni-servidor-dos` en `boe-legislacion`,
  `sin-la-eval-sin-binario-ni-servidor` y `dos-evals-sin-binario-ni-servidor` en `legal-core` («la copia incumple solo
  esas reglas: []»), y `orden-de-la-tabla` en los dos.
- Con `timeout-minutes: 122`, `TestDefinicionDelJob/del-repositorio` falla por el tope: «vale 122 (7320 s), y lo
  esperado es al menos el peor caso de boe-legislacion, 7557 s».

## Intento 2: en verde (`[X]`)

Aplicado en `internal/evals/sustitutos_test.go` el diff de arriba, sin ningún otro byte de código ni de tests. Ejecutado
en esta sesión, sobre el árbol del repositorio:

- Los cuatro tests (`TestTopeDeLaSesion`, `TestInterrumpirLaSesion`, `TestEjecutarSesionesConElContextoCancelado` y
  `TestEjecutarSesionesConUnError`) con `-race -tags=integration -count=40`: en verde (140 s).
- Mutante 1 (TERM solo al líder en `esperarConElMargen`): `TestInterrumpirLaSesion/abierta` falla por «el sustituto
  recibe TERM». Restaurado: `git diff --stat -- internal/evals/sesiones.go`, vacío.
- Con la regla `sin binario ni servidor` quitada de los dos juegos, `TestConjuntoDeEvals` falla en
  `sin-binario-ni-servidor-ninguna`, `sin-binario-ni-servidor-dos` y `orden-de-la-tabla` (`boe-legislacion`) y en
  `sin-la-eval-sin-binario-ni-servidor`, `dos-evals-sin-binario-ni-servidor`, `orden-de-la-tabla` y `tamaño`
  (`legal-core`). Con `timeout-minutes: 122`, `TestDefinicionDelJob/del-repositorio` falla por el tope («vale 122
  (7320 s), y lo esperado es al menos el peor caso de boe-legislacion, 7557 s»). Restaurados los dos.
- `make skills-check`: en verde. `make ci`: en verde («ci: todos los controles en verde», código 0, lint con 0
  hallazgos).

Sigue sin medirse en un ejecutor de Linux: lo dirá la CI de la propuesta de cambio.
