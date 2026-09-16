# T032 · intento 1

## El cambio

Solo `internal/httpx/reintentos_test.go` e `internal/httpx/ritmo_test.go`, y en ellos solo la medida de las llegadas;
el decorador de ritmo, el de reintentos, `rafagaDelSitio` y la cadena del cliente no cambian.

- **Reintentos** (`dosErroresYUnAcierto`): `comienzo := time.Now()` antes de `cliente.Pedir`; la constante `holgura`
  (10 ms) desaparece y la aserción de pares consecutivos pasa a `llegada.Sub(comienzo) >= (n)·intervalo` para el
  intento n (1, 2 y 3), porque el robots.txt del sitio ocupa el turno cero.
- **Ritmo** (`TestRitmoSeparaPeticionesDelMismoSitio`): la constante `holgura` (25 ms) desaparece y la aserción de
  pares pasa a `llegada.Sub(comienzo) >= i·intervalo` para la llegada i (0 el robots.txt, 1 a 3 las peticiones),
  contra el `comienzo` que el test ya tomaba antes de la primera petición.
- Los comentarios de los dos tests dicen la ley entera: el limitador fija los turnos anclados al primero; la llegada
  añade un despacho que no es igual en todas; dos llegadas consecutivas pueden acercarse por debajo del intervalo sin
  que el ritmo falle; la cota contra el comienzo es la que el limitador garantiza y no lleva holgura. Sin
  `intervalos` en ningún fichero Go: «n veces el intervalo» y «%d × intervalo».

## Arnés, sobre copias desechables en `/tmp/kitlegal-t032/`

Una copia por variante: `antigua`, con `git clone` de `be901f7` (tests sin cambiar); `tardia`, `cadena`, `rafaga` y
`sinespera`, con `rsync -a --exclude .git --exclude bin` del árbol ya con los dos tests nuevos. Cada mutante se aplicó
solo, con la herramienta Edit, y se midió con `go -C <copia> test -race -v -run
'TestReintentosDosErroresYUnAcierto|TestRitmoSeparaPeticionesDelMismoSitio' ./internal/httpx/`.

Mutante del despacho tardío, una implementación correcta (nunca emite antes de su turno): en
`decoradorDeRitmo.RoundTrip`, tras `Wait`, un `atomic.Bool` del decorador duerme 15 ms el primer despacho que no es
`/robots.txt`.

| Variante | Resultado |
|---|---|
| `antigua` + despacho tardío, `-count=10` | `TestReintentosDosErroresYUnAcierto` **cae 10 de 10** (las dos ejecuciones de cada vuelta, 81,49 ms a 84,79 ms frente a 90 ms, la cifra de la verificación de T029); `TestRitmoSeparaPeticionesDelMismoSitio` pasa 10 de 10 |
| `tardia`: tests nuevos + despacho tardío, `-count=10` | **los dos pasan 10 de 10** |
| `cadena`: tests nuevos + `conReintentos` por debajo de `conRitmo` | cae el de reintentos: intento 2 a 103,80 y 104,47 ms frente a 200 ms, intento 3 a 104,51 y 105,14 ms frente a 300 ms |
| `rafaga`: tests nuevos + `rafagaDelSitio = 2` | caen los dos: primera llegada del ritmo a 1,64 ms frente a 150 ms (total 304 ms frente a 450 ms); intento 1 a 1,65 y 3,83 ms frente a 100 ms |
| `sinespera`: tests nuevos + `RoundTrip` sin `Wait` | caen los dos: todas las llegadas por debajo de 2,1 ms frente a 100–450 ms; total 2,07 ms frente a 450 ms |

En `cadena` el test del ritmo pasa, como debe: no tiene reintentos. En `rafaga` y `sinespera` el test del ritmo cae en
la aserción nueva de las llegadas, no solo en la de la duración total.

Las copias se borraron al terminar (`test ! -e /tmp/kitlegal-t032`).

## Sobre el árbol

- `golangci-lint fmt --diff` y `run` del linter fijado (`tools/golangci-lint/go.mod`) sobre `./internal/httpx/...`:
  sin diff de formato y `0 issues.`.
- `go test -race -count=20` de los dos tests: `ok`.
- `go test -race -shuffle=on -count=1 ./internal/httpx/`: `ok`.

## Verificación

`go clean -testcache` y `make ci` con esta nota ya escrita, en primer plano, con el log en `gates/ci.log` (ignorado):
`código 0`, `0 issues.`, `internal/httpx` en `ok` en los dos perfiles (97,2 %), «no leaks found» y
`ci: todos los controles en verde`, sin `FAIL` ni `truncated` en el log. En `git status --porcelain`, fuera del
directorio del feature solo cambian los dos ficheros de test declarados.

La tarea queda `[X]`.
