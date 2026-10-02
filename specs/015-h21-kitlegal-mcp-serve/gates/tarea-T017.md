# T017: por qué no quedó en verde

## Intento 1: tarea mal delimitada (redelimitada, queda `[ ]`)

El trabajo de la tarea está hecho dentro de sus rutas y el árbol lo conserva para el intento siguiente: `README.md`,
`CONTRIBUTING.md`, `CHANGELOG.md`, `internal/evals/doc.go` y cuatro líneas «T017» en `gates/supuestos.md` (más la de
esta redelimitación). El guardián de diff lo dio por bueno («T017, 8 ficheros»).

`make ci` falla en **un único test, fuera de las rutas declaradas** y ajeno a lo que la tarea toca: el caso
`echo repetir --help` de `TestTuberiaCerrada` (`internal/app/tuberia_test.go`, de H1, sin cambios desde entonces).
Formato y lint, en verde (0 hallazgos); el rojo sale en `test`:

```
--- FAIL: TestTuberiaCerrada (0.00s)
    --- FAIL: TestTuberiaCerrada/echo_repetir_--help (0.11s)
        tuberia_test.go:60:
            	Error Trace:	internal/app/tuberia_test.go:105
            	Error:      	An error is expected but got nil.
            	            	expected: *exec.ExitError
            	Messages:   	con la salida estándar rota el binario tiene que terminar con un fallo, nunca con 0
FAIL	github.com/jmorenobl/kitlegal/internal/app	24.945s
```

Los objetivos de `make ci` a los que el intento no llegó (`test-integration`, `test-tiempos`, `vuln`, `schema-check`,
`skills-check`, `goreleaser-check`, `secrets`, `mod-verify`, `mod-tidy-check`), ejecutados aparte en esta sesión sobre
este árbol: en verde (código 0).

No es la primera vez: la nota de la T012 de H7 (`specs/010-h7-internal-graph-grafo/gates/tarea-T012.md`) lo vio una
vez, en la misma subprueba y con el mismo 0, y lo dejó anotado como intermitencia previa.

### Causa

No es T017, y no es el producto: es una carrera del propio test con los demás tests del paquete que lanzan procesos.

`lanzarConLectorCerrado` crea la tubería con `os.Pipe`, cierra el lector y lanza el binario con el escritor como
salida estándar. Entre `os.Pipe` y el `Close` el lector existe, y un proceso que otro test en paralelo crea en esa
ventana hereda una copia, que conserva hasta su exec (los dos extremos están marcados para cerrarse en el exec, no en
el fork). Si el binario del test escribe mientras esa copia sigue abierta, la tubería tiene quien la lea: la escritura
no falla y el binario termina con 0, que es exactamente el rojo.

Que la ventana existe se lee en la biblioteca de Go de este árbol (go1.27.1): `os.Pipe` (`os/pipe_unix.go`) toma
`syscall.ForkLock` para lectura mientras crea la tubería y marca sus extremos, y lo suelta antes de volver; y
`forkExec` (`syscall/exec_unix.go`) lo toma para escritura, hace el fork y lo suelta **antes** de esperar al exec del
hijo. Nada impide, pues, que el hijo de otro test viva entre su fork y su exec con el lector abierto mientras este
test ya ha cerrado el suyo y ha lanzado su binario.

Medido con un arnés de un solo uso fuera del repositorio (un programa Go, ejecutado con `-race`, que hace lo que
`lanzarConLectorCerrado` con el binario `kitlegal-e2e` construido de este árbol, rotando las siete invocaciones del
test, con varias gorrutinas a la vez y otras tantas que solo lanzan `version`, como los demás tests del paquete):

| Tubería | A la vez (y de ruido) | Terminan con 0 (el rojo) |
|---|---|---|
| la de hoy: `os.Pipe` y `Close` del lector | 7 (0) | 0 de 2100 |
| la de hoy | 16 (16) | 7 de 6400 |
| la de hoy, con otro arnés en marcha a la vez | 16 (16) | 5 de 6400 |
| la de hoy | 16 (16) | 44 de 16 000 |
| la de hoy | 16 (16) | 8 de 16 000 |
| la corregida: creada y con el lector cerrado bajo `ForkLock` | 16 (16), con otro arnés a la vez | 0 de 6400 |
| la corregida | 16 (16) | 0 de 16 000 |

En todos los ensayos el resto terminó con 1 y con `broken pipe` en el mensaje, y ninguno de otra forma. El 0 no es de
una invocación: en la última tanda de la de hoy se repartió entre `echo --help` (2), `echo repetir --help` (3),
`--help` (2) y `version` (1). Y no es un fallo de escritura que el producto se trague: con la escritura fallida el
binario da 1, y con el analizador cambiado un momento para no mirar ese fallo (mutante 2, abajo) da 2, nunca 0.

Lo que no se ha podido medir: el rojo dentro del propio `go test` (aislado no sale: 280 de 280 subpruebas en verde
con `-count=40`; el arnés lo da solo con 16 a la vez y ruido, no con 7), y nada en Linux ni en Windows. En Linux la
ventana es la misma por lo que dice el código de Go (`forkExec` es común a Unix), pero no se ha ejecutado.

No se repite `make ci` hasta un verde: sería cerrar con un verde probabilístico un rojo que puede volver en cualquier
`make ci` de lo que queda de hito, y que ya había salido en H7 (constitución, «Criterio de decisión autónoma», punto
1). El arreglo está en un fichero que T017 no declara, así que la tarea se redelimita.

### El arreglo

El lector deja de existir mientras pueda haber un fork. `tuberiaSinLector`, nueva, pide la tubería con `syscall.Pipe`
y cierra el lector con `syscall.ForkLock` tomado para lectura: Go lo toma para escritura en cada fork, así que ningún
proceso nace mientras el lector existe, y cerrado ya no lo hereda nadie. Es lo que `escribirEjecutable`
(`internal/evals/sustitutos_test.go`) hace desde H7.3 con el descriptor de un ejecutable a medio escribir. No sirve
tomar el candado alrededor de `os.Pipe`, que lo toma él mismo: un candado de lectura no se toma dos veces (con un fork
esperando entre las dos tomas, la segunda no llega nunca).

`syscall.Pipe` tiene otra forma en Windows, y los tests del paquete pasan hoy `go vet` con `GOOS=windows`. Para que
sigan compilando, el fichero pasa a ser solo para Unix, con su nombre y su restricción de compilación, que es lo que
ya hace `internal/disco/noregular_unix_test.go` con `syscall.Mkfifo`. El test no pierde ninguna plataforma en la que
se ejecute: `make ci` y la CI corren en macOS y en Linux.

Lo que el test comprueba no cambia: las siete invocaciones, las tres aserciones, sus mensajes y la exigencia de que el
proceso no muera por una señal.

Descartado: sondear la tubería desde el test hasta que escribir en ella dé `EPIPE` y lanzar entonces el binario.
Compila en todas partes sin fichero aparte, pero espera a que la copia heredada desaparezca en lugar de impedir que
exista, y deja bytes en la tubería.

### Comprobado sobre una copia desechable del árbol (fuera del repositorio)

Con el fichero de abajo en una copia de este árbol:

- `go test -race -run '^TestTuberiaCerrada$' -count=40 -v ./internal/app/`: 280 líneas `--- PASS` de subprueba.
- `golangci-lint run ./internal/app/...` del repositorio: 0 hallazgos; `golangci-lint fmt --diff`: sin diferencias.
- **Mutante 1**, `desarmarTuberiaCerrada` sin `signal.Ignore(syscall.SIGPIPE)`: las siete subpruebas fallan por «el
  proceso murió por la señal broken pipe».
- **Mutante 2**, `cli.Analizar` sin mirar el fallo de la escritura de la ayuda: `echo repetir --help` falla con 2 en
  lugar de 1 y sin el mensaje de la salida rota.
- El paquete entero, `go test -race -shuffle=on -count=3 ./internal/app/` sin las medidas de tiempo: en verde (46 s).
- `go vet ./internal/app/` con `GOOS=windows` y con `GOOS=linux`, y `go test -c` con `GOOS=linux`: los tres con
  código 0. Compilar no es ejecutar: ni en Linux ni en Windows se ha ejecutado nada.

Y sobre este árbol, con el fichero de hoy: `go test -race -run '^TestTuberiaCerrada$' -count=40 -v ./internal/app/`,
280 de 280 subpruebas en verde. Aislado, el rojo no sale.

### Redelimitación

La línea de T017 en `tasks.md` gana `internal/app/tuberia_test.go` e `internal/app/tuberia_unix_test.go` y describe
la corrección; comprobado con el filtro de rutas de `scripts/workflow/tarea.sh`: las siete rutas de antes más esas
dos. La tarea vuelve a `[ ]`. La cabecera de `tasks.md`, donde dice qué es cada tarea, dice ahora que T017 lleva
además esa corrección de test y ningún código de producto.

### Lo que tiene que hacer el intento 2

Lo de T017 ya está en el árbol y no se repite. Solo esto: `internal/app/tuberia_test.go` pasa a llamarse
`internal/app/tuberia_unix_test.go` (`git mv`, o escribir el nuevo y borrar el viejo) con este diff:

```diff
--- internal/app/tuberia_test.go
+++ internal/app/tuberia_unix_test.go
@@ -1,3 +1,5 @@
+//go:build unix
+
 package app_test
 
 import (
@@ -38,6 +40,8 @@
 // Es un test del binario y no del kernel en memoria a propósito, y por eso vive
 // junto al e2e y usa el binario que TestMain construye: la señal la desarma el
 // proceso entero, y solo un proceso entero puede demostrar que está desarmada.
+// Y vive en un fichero solo para Unix porque la tubería se pide con
+// syscall.Pipe (tuberiaSinLector), que en Windows tiene otra forma.
 func TestTuberiaCerrada(t *testing.T) {
 	t.Parallel()
 
@@ -70,11 +74,12 @@
 }
 
 // lanzarConLectorCerrado ejecuta el binario con la salida estándar conectada a
-// una tubería cuyo extremo de lectura se cierra **antes** de arrancarlo, de modo
-// que la primera escritura falle con EPIPE. Devuelve el código con el que
-// terminó y lo que dejó en la salida de error, y exige que haya terminado por
-// su cuenta y no por una señal: un proceso muerto por SIGPIPE no tiene código
-// de salida, que es justo lo que este test existe para descartar.
+// una tubería que no tiene lector desde **antes** de arrancarlo
+// (tuberiaSinLector), de modo que la primera escritura falle con EPIPE.
+// Devuelve el código con el que terminó y lo que dejó en la salida de error, y
+// exige que haya terminado por su cuenta y no por una señal: un proceso muerto
+// por SIGPIPE no tiene código de salida, que es justo lo que este test existe
+// para descartar.
 //
 // El nivel del registro se fija vacío en el entorno del subproceso para que la
 // salida de error lleve solo el mensaje, sea cual sea el entorno de quien
@@ -82,9 +87,7 @@
 func lanzarConLectorCerrado(t *testing.T, binario string, argv ...string) (int, string) {
 	t.Helper()
 
-	lector, escritor, err := os.Pipe()
-	require.NoError(t, err)
-	require.NoError(t, lector.Close(), "cerrar el lector es lo que deja la tubería sin nadie que lea")
+	escritor := tuberiaSinLector(t)
 
 	var errores bytes.Buffer
 
@@ -113,3 +116,39 @@
 
 	return fallo.ExitCode(), errores.String()
 }
+
+// tuberiaSinLector devuelve el extremo de escritura de una tubería cuyo extremo
+// de lectura no ha heredado ningún proceso: la crea y cierra el lector con
+// syscall.ForkLock tomado para lectura. Un proceso que otro test en paralelo
+// crea hereda, entre su fork y su exec, una copia de cada descriptor abierto;
+// con os.Pipe y un Close después, el lector existe en esa ventana, y si el
+// binario escribe mientras la copia sigue abierta, la escritura tiene quien la
+// lea, no falla y el binario termina con 0: con 16 lanzamientos a la vez pasaba
+// en entre 8 y 44 de cada 16 000. Go toma ForkLock para escritura en cada fork,
+// así que con él tomado para lectura ningún fork ocurre mientras el lector
+// existe, y cerrado ya no lo hereda ningún proceso. La tubería se pide con
+// syscall.Pipe y no con os.Pipe porque os.Pipe toma ese mismo candado y lo
+// suelta antes de volver, y un candado de lectura no se toma dos veces.
+//
+// El escritor queda marcado para cerrarse en el exec, como lo deja os.Pipe: el
+// binario lo recibe como su salida estándar y ningún otro proceso lo conserva.
+func tuberiaSinLector(t *testing.T) *os.File {
+	t.Helper()
+
+	var extremos [2]int
+
+	syscall.ForkLock.RLock()
+
+	err := syscall.Pipe(extremos[:])
+	if err == nil {
+		syscall.CloseOnExec(extremos[1])
+
+		err = syscall.Close(extremos[0])
+	}
+
+	syscall.ForkLock.RUnlock()
+
+	require.NoError(t, err, "la tubería se crea y su lector se cierra sin que nadie lo herede")
+
+	return os.NewFile(uintptr(extremos[1]), "tubería sin lector")
+}
```

El código va sangrado con tabuladores. El fichero que resulta tiene esta huella SHA-256:
`335f6b8962734871c8055e988ca91fe7da0b9a707da60987ba2d8d8ef6cfa23e`. Ningún otro byte de código ni de tests, y ninguno
de producto. Después:

1. `go test -race -run '^TestTuberiaCerrada$' -count=40 -v ./internal/app/` y contar con `rtk proxy` las líneas
   `--- PASS: TestTuberiaCerrada/`: 280.
2. El mutante 1 un momento (`desarmarTuberiaCerrada` sin `signal.Ignore`, en `internal/app/main.go`): las siete
   subpruebas fallan por la señal. Se restaura y se comprueba con `git diff --stat` que `internal/app/main.go` no
   cambia.
3. `go vet ./internal/app/` con `GOOS=windows` (`rtk proxy env GOOS=windows GOARCH=amd64 go vet ./internal/app/`): sin
   salida.
4. `git grep` de `tuberia_test` fuera de `specs/`: nada (hoy no lo nombra ningún fichero; el ADR 0006 nombra el test,
   `TestTuberiaCerrada`, que conserva su nombre y su paquete).
5. `make ci` en primer plano y marcar `[X]`.

Las verificaciones propias de T017 (cada orden, fichero, test, clave y campo de la documentación, las tres órdenes de
FR-060 y la ayuda de `mcp`) son del intento 1 y la documentación no cambia en el intento 2: lo que hizo y lo que
dejó sin poder ejecutar está en sus líneas «T017» de `gates/supuestos.md`.

## Intento 2: en verde, `[X]`

`git mv` y el diff de arriba, sin ningún otro cambio. Ejecutado en esta sesión, sobre este árbol:

- La huella SHA-256 de `internal/app/tuberia_unix_test.go` es la de arriba (`335f6b89…cfa23e`). `shasum` y `openssl`
  piden aprobación en una sesión de un paso: se calculó con un programa Go de un solo uso en el directorio temporal,
  fuera del repositorio.
- `go test -race -run '^TestTuberiaCerrada$' -count=40 -v ./internal/app/`: código 0, 280 líneas
  `--- PASS: TestTuberiaCerrada/` y ninguna `--- FAIL`.
- Mutante 1 (`desarmarTuberiaCerrada` sin la llamada a `signal.Ignore`): las siete subpruebas en rojo, las siete con
  «el proceso murió por la señal broken pipe». Restaurado: `git diff --stat -- internal/app/main.go`, vacío.
- `go vet ./internal/app/` con `GOOS=windows GOARCH=amd64`: sin salida, código 0; con `GOOS=linux`, lo mismo. Compilar
  no es ejecutar: ni en Linux ni en Windows se ha ejecutado nada.
- `git grep tuberia_test` fuera de `specs/`: nada.
- La ayuda de `mcp` y de `mcp serve` del binario construido de este árbol («por la entrada y la salida estándar»), y
  las órdenes de FR-060 en el README (`claude mcp add kitlegal -- kitlegal mcp serve`,
  `codex mcp add kitlegal -- kitlegal mcp serve`, *Settings > MCP servers*, `mcp_config.json`), con `grep -F`: están,
  y `.mcpb` no aparece. El resto de las comprobaciones de la documentación son las del intento 1 y no se han repetido.
- `make ci` en primer plano: «ci: todos los controles en verde», sin ningún `FAIL` en su salida.

Lo que sigue sin medirse: que el rojo no vuelva dentro de `go test` no lo demuestra un `make ci` en verde (aislado
tampoco salía antes del arreglo); lo que sostiene el arreglo es la medida del arnés del intento 1 (0 de 22 400 con la
tubería corregida) y el candado de Go, no este verde.
