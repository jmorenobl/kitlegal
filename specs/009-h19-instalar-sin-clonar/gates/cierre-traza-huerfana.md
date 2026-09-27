# Cierre · La traza con el hilo huérfano que muere en la parada de entrada

## Qué pasó

La primera medición de cierre (`gates/cierre.json`, 2026-09-27T04:00:51Z) sobre `aa9b628` dejó en rojo el trabajo
`evals (boe-legislacion)` de la ejecución 36291141634 (trabajo 108541351437); `ci`, `snapshot`, `evals (legal-core)`
y los cuatro estados de Codecov, en verde.

- Único motivo de la raíz del informe:

  ```text
  18-lrjpac-norma-derogada-claude-sonnet-5-01: sesión ilegible: traza: traza ilegible: …/traza: …/traza/t.39746, …/traza/t.40041 no tienen la línea clone, clone3, fork o vfork que los crea, y solo puede faltarle a uno con llamadas, el del proceso que arrancó strace
  ```

- La serie 18 con Sonnet llega al umbral (2 de 3): las otras dos sesiones pasan con los avisos `derogada` y
  `vigencia-agotada`, y `red` está vacío. Como en T015 de H5.1, la sesión ilegible hace `fallo` el veredicto entero,
  porque sin traza legible no se puede afirmar que la sesión no llegó a la red (FR-076 de H5).
- La ejecución no sube las trazas: solo se tiene el error. `t.39746` es el raíz (el proceso `claude`); `t.40041`, un
  fichero sin línea de creación al que el lector cuenta alguna llamada.

## Hipótesis

La regla 1 de data-model §9 de H5 admite el **huérfano** —un fichero sin la línea que lo crea— solo si no tiene
ninguna llamada y alguna `clone`, `clone3`, `fork` o `vfork` de la traza quedó sin terminar (research V65: 8 de 1000
trazas). Un fichero sin creador con la llamada desconocida (forma D: `???( <unfinished ...>` o `???() = ?`) tiene una
llamada para el lector, así que pasaba por segundo raíz. `t.40041` sería el hilo que una `clone` sin terminar de
`kitlegal` llegó a crear y que el núcleo mató en la parada de entrada de una llamada que strace no llegó a identificar,
cuya creación strace no vio.

## Sonda

Contenedor `kitlegal-t036-a` (ubuntu:24.04, aarch64, strace 6.8), `docker run --rm --network none --cap-add
SYS_PTRACE`, el programa de V65 (64 goroutines con `runtime.LockOSThread()` y `time.Sleep(time.Hour)`, `os.Exit(0)`
desde `main`) compilado con `GOOS=linux GOARCH=arm64 CGO_ENABLED=0`, con una espera de N µs antes de salir, bajo la
orden de strace de la sesión (`-ff -e trace=execve,connect,clone,clone3,fork,vfork -s 131072 -o t`). Guiones y binario
fuera del repositorio, sin versionar. Primer guion: por traza, los ficheros sin línea de creación (ninguna línea de otro
fichero termina en `= N`) que no son el raíz y tienen algo más que su línea final.

| espera | repeticiones | trazas con alguna llamada sin terminar | huérfanos con solo la línea final | ficheros sin creador con líneas |
|---|---|---|---|---|
| 0 µs | 4000 | 48 | 22 | 0 |
| 20 µs | 3000 | 168 | 107 | 5, en 3 trazas |
| 100 µs | 3000 | 987 | 705 | 17, en 17 trazas |

De los 22 ficheros sin creador con líneas, 20 tienen exactamente `???()`, 35 espacios y `= ?` —la línea del runner de
T015, carácter a carácter— seguida de `+++ exited with 0 +++`, siempre en una traza con una `clone` de forma A
(`= ?`) o B (`= ? <unavailable>`); una traza tiene dos (repetición 2037, con una A y una B). Los otros 2 están en la
repetición 1783, donde además **dos líneas crean el mismo hilo** (`t.194419`, línea 3, y `t.194420`, línea 1, las dos
`= 194423`) y `t.194423` y `t.194424` tienen `syscall_0x2f777(0x7b9a036dfe00, 0, 0x6a7341b073188, 0, 0x3e8, 0x4e20
<unfinished ...>` (`0x2f777` = 194423): ilegible igualmente por «dos líneas crean», que es lo correcto.

Segundo guion: en **todos** los ficheros de cada traza, las líneas que no son de ninguna forma conocida. 3000
repeticiones con 100 µs → 87 trazas; 2000 con 300 µs → 51. Dos formas:

- `syscall_0x<hex>(<seis argumentos, 0 o hexadecimal> <unfinished ...>` en 147 ficheros, **todos con línea de
  creación**, siempre como última llamada antes de la línea final, a veces tras varias `clone` completas (p. ej.
  `t.3345` de la repetición 123: cuatro `clone` y `syscall_0xd18(0, 0, 0, 0x1, 0x54eb09722000, 0x54eb096ac218
  <unfinished ...>`). Es la parada de entrada del hilo que muere cuando strace sí lee los registros pero el número no
  es de ninguna llamada que conozca, y strace no puede filtrarla porque no tiene nombre.
- `clone(child_stack=0x…, flags=…) = -1 (errno 18446744073709551481)` y `… (errno 18446744073709551438)` en 2 trazas
  (repeticiones 1478 y 2196): la `clone` que crea el último hilo de la traza, que existe con solo su línea final. Un
  errno que no es ninguno de Linux: el valor del registro de una llamada que terminaba con el proceso.

Ninguna `<... resumed>` ni `<detached ...>`; ninguna señal salvo `--- SIGURG {si_signo=SIGURG, si_code=SI_TKILL, …} ---`
(una vez, en un hilo con línea de creación).

## Decisión

1. **Un fichero sin línea de creación cuyas llamadas son todas desconocidas es un huérfano** (regla 1): `???`, abierta
   o cerrada, o `syscall_0x…` sin terminar, con la misma condición de siempre, que alguna `clone`, `clone3`, `fork` o
   `vfork` de la traza quedara sin terminar. Con cualquier llamada del filtro —también una que el fin del proceso dejó
   sin terminar— sigue siendo ilegible, porque no se puede atribuir.
2. **`syscall_0x<hex>(<seis argumentos crudos> <unfinished ...>` y su cierre `) = ?` se leen como la llamada
   desconocida**: no es del filtro, no se ejecutó y no tiene nada que atribuir. Con otros argumentos, otro resultado o
   la marca dentro, ilegible. La forma cerrada no la dio la sonda: se admite porque strace cierra así toda llamada sin
   terminar y T015 pagó una ejecución en rojo por admitir solo la abierta de `???`.
3. **Rechazado admitir `clone(…) = -1 (errno N)`**: es el resultado corrupto de una llamada del filtro, no la ausencia
   de una llamada; solo se ha visto en arm64, nunca en el runner; y admitirlo obligaría a decidir qué significa en una
   `execve` o un `connect`. Queda ilegible, con un test que lo fija (`clone-con-errno-sin-nombre`) y la evidencia aquí.
4. **Rechazado relanzar sin arreglar** (la carrera volvería), **tocar la eval** (el fallo es del lector, no de la skill)
   y **quitar la sesión ilegible del veredicto** (FR-076).
5. **El data-model de H5 (`specs/006`) no se edita**, como en T015 de H5.1: la regla queda en el código, en sus tests
   con las líneas literales de la sonda y en esta nota; el CHANGELOG lleva la entrada bajo *Corregido · De H19*.

## Cambios

- `internal/evals/trazas.go`: `hilo.conLlamadaDelFiltro`, `llamada.desconocida`, `prefijoDeNumeroDesconocido` y
  `formaDeArgumentosCrudos`; `raiz` clasifica por llamadas del filtro y sus tres errores lo dicen; `leerLlamada` y
  `leerLlamadaSinCerrar` leen `syscall_0x…`.
- `internal/evals/trazas_test.go`: `TestLeerTrazasHuerfanoConLlamadaDesconocida` (once casos, con las trazas reales de
  las repeticiones 2313, 2037, 123 y 1783) y los casos nuevos de `TestLeerLlamadaSinTerminar`.
- `CHANGELOG.md`: entrada bajo *Corregido*.

## Después

`make ci` en verde en local; el workflow commitea, empuja y vuelve a medir sobre la cabeza nueva.
