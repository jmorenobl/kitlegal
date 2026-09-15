# T045 · intento 1 de 3 (2026-09-15): detenida sin marcar y redelimitada

Base `208547b` (`feat(H5): T044`). **Sin ningún cambio de código ni de datos**: el test que se empezó a escribir
(los cinco casos de T044 en la tabla de `TestLeerTrazas`) se deshizo con `git checkout` al detener la tarea, y el árbol
queda como en la base salvo esta nota y la línea de T045 en `tasks.md`. La línea de T045 manda: «si [la sonda V65]
muestra otra forma, la tarea se detiene sin marcarse y lo anota en `gates/tarea-T045.md`». La sonda mostró **cuatro
formas** de la llamada que el fin del proceso deja en curso, y no la única que la tarea preveía; una de ellas es la de
todo `connect` en curso, que el cambio tal como estaba escrito habría rechazado expresamente («`= ?` sin
` <unfinished ...>` delante del paréntesis de cierre sigue siendo ilegible»). La tarea queda `[ ]` con su línea
redelimitada para el intento siguiente (mismas rutas; abajo, «Redelimitación»).

## Las sondas (V65)

Contenedor desechable de la imagen `kitlegal-t036-a` (`ubuntu:24.04`, aarch64; `strace -V` → `strace -- version 6.8`),
como root, con `docker run --rm --network none --cap-add SYS_PTRACE` y un directorio temporal montado. Cada repetición
ejecuta `strace -ff -e trace=execve,connect,clone,clone3,fork,vfork -s 131072 -o <dir>/t -- <programa>`, las opciones de
la orden de la sesión (contrato del job §3.2), en un directorio vacío. Los programas son desechables, fuera del
repositorio, compilados con Go 1.27.1 (`GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath`):

- **hilos**: 64 goroutines que hacen `runtime.LockOSThread()` y `time.Sleep(time.Hour)`, y `os.Exit(0)` desde `main`
  sin esperarlas. Cada goroutine que se queda con su hilo y se bloquea obliga al planificador a crear otro hilo, desde
  el hilo que esté planificando; `os.Exit` llama a `exit_group` mientras tanto.

  ```go
  func main() {
      for range 64 {
          go func() {
              runtime.LockOSThread()
              time.Sleep(time.Hour)
          }()
      }
      os.Exit(0)
  }
  ```

- **connect**: un socket que escucha en `127.0.0.1` con cola 0 y no acepta (como la sonda de V54), cuatro goroutines
  con un `syscall.Connect` bloqueante a su dirección (la primera conexión cabe en la cola; las demás esperan dentro de
  `connect`) y `os.Exit(0)` desde `main` a los 500 ms.
- **binario**: el `kitlegal` del árbol en `208547b` detrás de un enlace `boe`, con
  `boe articulo BOE-A-2015-10565 a21 --offline --json` y `KITLEGAL_CACHE_DIR` vacío.

Tras cada repetición, un guion cuenta las líneas que no tienen ninguna forma de data-model §9 (con las direcciones
`0x…`, los puertos y los descriptores normalizados) y las trazas con más de un fichero sin línea de creación, y guarda
la primera traza de cada combinación.

| Sonda | Repeticiones | Código de `strace` | Trazas con líneas de otra forma | Trazas con un fichero huérfano |
|---|---|---|---|---|
| hilos, primera pasada (hasta la primera `unfinished`) | 22 | 0 | 1 | 0 |
| hilos | 1000 | 0 en las 1000 | 66 | 8 |
| connect | 5 | 0 en las 5 | 5 | 0 |
| binario | 500 | 4 en las 500 | 0 | 0 |

**Formas en las 1000 repeticiones de hilos** (trazas en las que aparece cada una): 53 `???( <unfinished ...>`; 11
`clone(child_stack=0x…, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM) = ?`; 1
`clone(child_stack=0x…, flags=CLONE_VM|CLONE_FS|CLONE_FILES|CLONE_SIGHAND|CLONE_THREAD|CLONE_SYSVSEM <unfinished ...>`,
sin paréntesis de cierre ni resultado; y 1 `clone(child_stack=0x…, flags=…) = ? <unavailable>`. **En las 5 de connect**,
cada `connect` que esperaba (tres por traza, en el hilo principal o en otro) es
`connect(<fd>, {sa_family=AF_INET, sin_port=htons(<puerto>), sin_addr=inet_addr("127.0.0.1")}, 16) = ?`, **sin**
` <unfinished ...>`; el que cupo en la cola, `= 0`. En ninguna traza hay `<... resumed>` ni `<detached ...>`, y a cada
línea de estas formas la sigue en su fichero, directamente, `+++ exited with 0 +++`.

### Líneas literales de las trazas guardadas

Con los números reales de hilo (el intento siguiente las copia en los tests; las trazas se quedaron en un directorio
temporal que puede desaparecer):

| Traza | Fichero y línea | Texto | Qué la rodea |
|---|---|---|---|
| hilos, primera pasada, repetición 22 | `t.258`, 1 | `???( <unfinished ...>` | línea 2 `+++ exited with 0 +++`; `t.258` lo crea `t.256`, línea 1, `clone(child_stack=0xc413c518000, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM) = 258`; el raíz `t.254` empieza por `execve("/sonda/sonda", ["/sonda/sonda"], 0xffffcbba8c90 /* 4 vars */) = 0` |
| hilos `49d6c222` | `t.336`, 1 | `???( <unfinished ...>` | línea 2 `+++ exited with 0 +++`; lo crea `t.334`, línea 1, `= 336`; ningún huérfano |
| hilos `32269785` | `t.4030`, 1 | `clone(child_stack=0x2bb2e2418000, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM) = ?` | línea 2 `+++ exited with 0 +++`; `t.4030` lo crea el raíz `t.4028`, línea 3; **`t.4032`, huérfano**: solo `+++ exited with 0 +++` y ninguna línea lo crea (el raíz crea 4029, 4030 y 4031) |
| hilos `b6609767` | `t.5665`, 2 | `clone(child_stack=0x6b63b5b94000, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM) = ?` | línea 1 `clone(child_stack=0x6b63b5b98000, …) = 5667`, línea 3 `+++ exited with 0 +++`; ningún huérfano |
| hilos `3f18ff32` | `t.20425`, 2 | `clone(child_stack=0x203cb5a94000, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM) = ? <unavailable>` | línea 1 `clone(child_stack=0x203cb5a98000, …) = 20427`, línea 3 `+++ exited with 0 +++`; **`t.20428`, huérfano**: solo `+++ exited with 0 +++` (el raíz `t.20423` crea 20424, 20425 y 20426; `t.20425` crea 20427 y `t.20427` crea 20429) |
| hilos `adb8222a` | `t.15079`, 2 | `clone(child_stack=0x666394d64000, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM <unfinished ...>` | línea 1 `clone(child_stack=0x666394d98000, …) = 15081`, línea 3 `+++ exited with 0 +++`; `t.15082`, que crea el raíz (`= 15082`), tiene como línea 1 el mismo texto, con la misma pila, y como línea 2 `+++ exited with 0 +++`; ningún huérfano |
| connect `4385e6e0` | `t.14`, 5; `t.17`, 2; `t.18`, 1 | `connect(9, {sa_family=AF_INET, sin_port=htons(60929), sin_addr=inet_addr("127.0.0.1")}, 16) = ?` (y `connect(10, …) = ?`, `connect(8, …) = ?`) | la siguiente de cada una, `+++ exited with 0 +++`; `t.14` es el raíz, con su `execve` y tres `clone` antes; `t.16`, línea 2, `connect(7, …) = 0` |
| connect `800061f5` | `t.138`, 7; `t.142`, 1; `t.143`, 1 | `connect(10, {sa_family=AF_INET, sin_port=htons(42445), sin_addr=inet_addr("127.0.0.1")}, 16) = ?` (y `connect(7, …) = ?`, `connect(9, …) = ?`) | `t.138`, el raíz, con `connect(8, …) = 0` en la línea 6; la siguiente de cada una, `+++ exited with 0 +++` |

Las ocho trazas con un fichero huérfano tienen todas una `clone` sin resultado: sus combinaciones guardadas (formas
normalizadas y número de ficheros sin línea de creación) son solo dos, `clone(…) = ?` con dos ficheros sin creación
(siete trazas) y `clone(…) = ? <unavailable>` con dos (la única traza con esa forma). En todas, el huérfano tiene solo
`+++ exited with 0 +++`. Ninguna traza con `???(` o con la forma sin cerrar tiene huérfano.

### Lectura de esas trazas

- **Con el lector de `208547b`** (arnés temporal en una copia desechable del árbol): las ocho trazas son ilegibles, cada
  una en su primera línea de estas formas y con el motivo del runner, «no es ninguna de las formas de línea de la
  traza: execve, clone, clone3, fork, vfork o connect con su resultado, una señal o la línea final»: `t.258` línea 1,
  `t.336` línea 1, `t.4030` línea 1, `t.5665` línea 2, `t.20425` línea 2, `t.15079` línea 2, `t.14` línea 5 y
  `t.138` línea 7.
- **Con un prototipo del diseño redelimitado** en la misma copia (no es la implementación: sin tests nuevos ni
  documentación): las ocho legibles, sin invocaciones (los programas no son applets); los tres casos legibles de T044
  con lo que fija el contrato del job §9 (`clone-sin-terminar` y `hilo-de-clone-sin-terminar`, la invocación 2000 con
  `boe articulo BOE-A-1978-31229 a140 --json`, código 0 y sin conexiones; `connect-sin-terminar`, código 0 y la
  conexión `203.0.113.7:443` con resultado `?`, sin resultado y de clase `red`); los dos negativos ilegibles con su
  motivo (`clone-sin-terminar-seguido-de-otra-llamada`, `t.2001` línea 1 y «a la llamada sin resultado solo pueden
  seguirla líneas de señal y la línea final»; `huerfano-sin-clone-sin-terminar`, `t.2000` y `t.2002` sin la línea que
  los crea); y `TestLeerTrazas`, `TestLeerTrazasConDefectos`, `TestLeerTrazasSinFicheros`,
  `TestLeerTrazasSinInvocaciones`, `TestLeerLlamadaConRelleno` y `TestInforme` en verde.

## Qué dicen las formas

Las cuatro formas son coherentes con dos caminos de `strace` para un hilo que muere dentro de una llamada, y todas las
observaciones casan con ellos:

- **(A) `<llamada>(<argumentos>[ <unfinished ...>]) = ?`**: el hilo llega a su parada de fin (el proceso sale) dentro
  de la llamada, y `strace` cierra la línea con `) = ?`. La marca ` <unfinished ...>` delante del paréntesis solo sale
  si al decodificador de la llamada le quedaba algo por escribir a la salida: la `clone` de amd64 lleva `CLONE_SETTLS` y
  escribe `tls=0x…` a la salida (la línea normal del runner termina en `, tls=0x…) = N` y la sin terminar, en las
  banderas y la marca); la de arm64 no lleva ese argumento y sale sin la marca (11 trazas); `connect` escribe toda su
  dirección a la entrada y sale siempre sin ella (5 de 5). La marca, por tanto, no dice nada que la lectura necesite.
- **(B) `<llamada>(<argumentos>) = ? <unavailable>`**: la llamada llegó a su parada de salida, pero `strace` no pudo leer
  el resultado del hilo, que ya estaba muriendo; la `clone` de `3f18ff32` sí creó su hilo (el huérfano `t.20428`).
- **(C) `<llamada>(<argumentos> <unfinished ...>`** y **(D) `???( <unfinished ...>`**: `strace` escribió la entrada de la
  llamada, y el siguiente suceso de ese hilo fue su línea final, así que termina la línea con ` <unfinished ...>` sin
  paréntesis ni resultado. `???` es el nombre que escribe cuando no pudo leer siquiera qué llamada era: en las dos
  trazas guardadas es la única línea de un hilo recién creado, que murió en la parada de entrada de su primera llamada;
  como murió en esa parada, la llamada no llegó a ejecutarse y no hay conexión que atribuir.
- **El huérfano**: el hilo que una `clone` sin terminar llegó a crear y que el núcleo mató con el proceso antes de
  ninguna llamada trazada. `strace` le abre su fichero y escribe su línea final, pero ninguna línea tiene su número
  como resultado.

## Consecuencias para T044 y para T045 tal como estaba escrita

1. **El `connect` en curso de una invocación que sale sería ilegible.** Es exactamente el caso que FR-076 no puede
   perder (una petición que salió mientras el proceso terminaba), y la línea de T045 lo declaraba ilegible. Con la
   `clone` de arm64 pasa lo mismo.
2. **B, C y D también serían ilegibles**, y D es la más frecuente de todas (53 de 1000 en esta sonda). En la prueba de
   red, con 22 invocaciones por ejecución, el cierre seguiría dependiendo de una carrera.
3. **`connect-sin-terminar` de T044 no lleva la forma que `strace` escribe para `connect`** (`16 <unfinished ...>) = ?`,
   frente a `16) = ?` en 5 de 5). Con la marca opcional delante de `) = ?` en cualquier llamada, el caso sigue siendo
   legible y fija justo eso: que la marca también es opcional en `connect`. La forma real del `connect` entra en los
   tests sobre literales. No hace falta ninguna tarea de datos: rehacer ese fichero pararía el run en la revisión
   humana de `clasificar_datos` (modifica material existente de `testdata/`), y el caso sigue siendo válido.
4. **El huérfano existe** (8 de 1000), siempre con una `clone` de forma A o B en la traza: la regla 1 de T045 era
   correcta, pero tiene que contar como «creación sin terminar» también la de las formas B y C.

## Redelimitación

La línea de T045 en `tasks.md` se reescribe con **las mismas rutas** (`internal/evals/trazas.go`,
`internal/evals/trazas_test.go` y el directorio del feature) y sin repetir las sondas, cuyas líneas reales están en
esta nota:

- Las cuatro formas A, B, C y D, en cualquier sesión, como llamada sin resultado a la que solo siguen líneas de señal y
  la línea final; en A la marca es opcional; B y C solo como se escribieron; D sin argumentos. No crean hilos; una
  `execve` así no es una `execve` con 0; un `connect` así se clasifica por su dirección; `???` no tiene conexión.
- `? ERRNO (…)` sigue siendo solo del corte; la marca con un resultado (`= N`, `= ? ERRNO (…)`, `= ? <unavailable>`),
  `???(` con argumentos y la forma sin cerrar de una llamada fuera del filtro siguen siendo ilegibles.
- Huérfanos admitidos si alguna `clone`, `clone3`, `fork` o `vfork` de la traza es de forma A, B o C.
- Tests: los cinco casos de T044 en `TestLeerTrazas`; un test sobre literales con la línea del runner y las líneas
  reales de esta nota; y un test sobre trazas escritas en `t.TempDir()` con esas líneas (las trazas legibles de V65,
  un `connect` de forma A público y local en una invocación, y los defectos).
- Documentación: data-model §9, contrato del job §4, §9 y §9.1, research V53, V54, V65 y S4, y `gates/pr-h5.md`.

## Alternativas rechazadas

- **Implementar en este intento el diseño ampliado**: la línea de T045 mandaba detenerse ante otra forma, y el cambio
  que hace falta no es el que estaba escrito (regla 5 con cuatro formas, otro test, otras alternativas). Redelimitar es
  el procedimiento.
- **Admitir solo la forma del runner** (la de T045 tal como estaba): deja fuera el `connect` en curso (FR-076) y las
  otras tres formas, que salen del mismo `strace` y del mismo mecanismo.
- **Filtrar esas llamadas con la opción de estado de `strace`** en la orden de la sesión (`strace -h` de 6.8:
  `-e status=SET, --status=SET`, con los estados `successful`, `failed`, `unfinished`, `unavailable` y `detached`; A,
  C y D son las `unfinished`, y B, la `unavailable`): ocultaría el `connect` en curso de una invocación, que es lo que
  FR-076 obliga a ver, y cambia el contrato del job §3.2.
- **Ignorar las líneas que no casan**: dejaría hilos y conexiones sin atribuir en silencio.
- **Tratar `???(` como una conexión de destino desconocido**: no hay destino que informar y la llamada no se ejecutó;
  cada sesión con esa línea fallaría sin haber llegado a la red.
- **Una tarea de datos que rehaga `connect-sin-terminar` con la forma real**: pausa el run y no hace falta (punto 3).

## Verificación

`make ci` en primer plano sobre este árbol (sin cambios de código; `tasks.md` y esta nota son los únicos ficheros
tocados, además de los de estado del workflow): código 0 y `ci: todos los controles en verde` (log en `gates/ci.log`,
ignorado). La tarea no se marca: el verde es el de la base, no el de T045, y su línea mandaba detenerse ante otra
forma. La cobertura de `internal/evals` en la base, para el intento siguiente: 99,0 % de las sentencias
(`go test -count=1 -cover ./internal/evals/`).

# T045 · intento 2 de 3 (2026-09-15): en verde, con la línea redelimitada

Base `208547b`, sin ninguna sonda nueva: las líneas reales son las de la tabla de arriba. Ficheros tocados:
`internal/evals/trazas.go`, `internal/evals/trazas_test.go` y, del directorio del feature, `data-model.md` (§9),
`contracts/job-de-evals.md` (§4, §9 y §9.1), `research.md` (V53, V54, V65 nueva, D12 y S4), `gates/pr-h5.md`
(*Decisiones* y *Pendientes*), esta nota y `tasks.md`.

## El test primero

Con el lector de la base (solo el campo `sinTerminar` añadido al tipo `llamada`, sin lógica, para que el paquete
compilara), `go test -run '^(TestLeerTrazas|TestLeerTrazasSinTerminar|TestLeerLlamadaSinTerminar)$'` dio rojo en lo
que la línea de la tarea fija: los tres casos legibles de T044 (`clone-sin-terminar`, `hilo-de-clone-sin-terminar` y
`connect-sin-terminar`) y las cuatro trazas enteras de V65 y los dos `connect` en curso de una invocación, todos con
el motivo exacto del runner («no es ninguna de las formas de línea de la traza: execve, clone, clone3, fork, vfork o
connect con su resultado, una señal o la línea final»); los negativos de T044 y de V65, con otro motivo o sin error; y
las seis líneas literales, con ese mismo motivo.

## El cambio

- `formaDeLlamada` admite, además de `N`, `-1 ERRNO (…)` y `? ERRNO (…)`, los resultados `?` y `? <unavailable>`
  (formas A y B); la marca ` <unfinished ...>` se recorta del final de los argumentos y solo cabe con el resultado `?`,
  y no cabe dentro de los argumentos. `formaDeLlamadaSinCerrar` lee la línea de entrada sin cerrar (forma C) y
  `???( <unfinished ...>` (forma D, sin argumentos). Las cuatro son una `llamada` con `sinResultado` y `sinTerminar`,
  sin valor numérico (`creaHilo` es falso) y con el texto tras `= ` como `resultado` (vacío en C y D); los argumentos de
  execve, clone, clone3 y connect se leen igual que con resultado (`leerArgumentos`, compartido).
- `lectorDeHilo`: la llamada sin resultado que exige `cortada` es solo la interrumpida (`? ERRNO`); a cualquiera de
  ellas solo pueden seguirla líneas de señal y la línea final, como antes.
- Regla 1 (`raiz`): el raíz es el único fichero sin línea de creación con alguna llamada; los ficheros sin esa línea
  y sin llamadas son huérfanos si alguna creación de la traza es una llamada sin terminar (`conCreacionSinTerminar`);
  quedan anotados en `traza.huerfanos`, fuera de `procesos` (no son invocaciones ni tienen conexiones) y fuera de los
  «sueltos» de la regla 2. Sin una creación así, el error los nombra («… ni ninguna llamada, y ninguna clone, clone3,
  fork o vfork de la traza quedó sin terminar»); con dos ficheros con llamadas sin esa línea, el error los nombra a
  ellos y no al huérfano.

## Verificación

- `go test -count=1 -cover ./internal/evals/`: verde; cobertura de sentencias 99,0 % (exacta: 98,982 %, 1459/1474,
  frente a 98,960 %, 1428/1443, en la base con los mismos dos ficheros en `git stash`); los quince bloques sin cubrir
  son los mismos de la base (los errores de `nuevoInterprete` y de lectura de fichero).
- Linter fijado por el repositorio (`go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./internal/evals/...`):
  0 hallazgos, `dupl` y `misspell` incluidos (`unfinished` y `unavailable` están en el diccionario como correcciones,
  no como erratas); `golangci-lint fmt` aplicado al test.
- `make ci` en primer plano sobre este árbol: código 0 y `ci: todos los controles en verde` (log en `gates/ci.log`,
  ignorado). La tarea se marca `[X]`.

Ninguna línea real de la nota quedó sin leerse con estas reglas, así que la tarea no se detiene. Lo que sigue sin
evidencia en el runner —las formas B, C y D y el fichero huérfano— queda anotado en research S4 y en `gates/pr-h5.md`,
y lo vigila la lectura de las trazas del intento siguiente de T030 y de la ejecución de cierre.
