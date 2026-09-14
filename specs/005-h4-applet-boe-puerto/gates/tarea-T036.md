# T036 · intento 1 de 3: parada por una orden de la guía que pide aprobación

**Estado**: T036 **sin marcar**. La medida de cierre no está hecha y `gates/pr-h4.md` no se ha tocado: la evidencia de
cierre la escribe el intento que ejecute la guía entera. Base del intento: `55cc107`.

## Qué pasó

`quickstart.md` se ejecutó desde los prerrequisitos, con cada orden tal cual:

| Paso | Resultado |
|---|---|
| Prerrequisitos | `go1.27.1 darwin/arm64`, rama `h4-applet-boe-puerto`, ninguna línea entre `git status` y «fin del estado», carpeta temporal vacía |
| Sondas de las formas | `código 3`; `valor`; dos huellas distintas (`8eb7bb88…` y `7e877020…`); la sonda de ausencia no imprimió nada (salida 1) y la carpeta volvió a quedar vacía |
| Escenario 1 | Huellas de `~/.cache/kitlegal` iguales antes y después (`519f40d4…`). `make test` sin `FAIL` ni `WARNING: DATA RACE`, con `ok` en `cmd/kitlegal`, `internal`, `internal/app`, `internal/cache`, `internal/cli`, `internal/core/schema`, `internal/httpx`, `internal/render` e `internal/source/boe`. Ninguna línea en el `git status` de los datos protegidos |
| Escenario 2 | **No se ejecutó.** La sesión contestó: «This Bash command contains multiple operations. The following part requires approval: rtk proxy go test -count=1 -v -run '^TestArticuloCoincideConBoePy$' ./internal/source/boe/ \» |
| Escenario 3 | Cinco `--- PASS`. Se lanzó a la vez que el 2, antes de ver la parada, así que no cuenta como medida: el intento siguiente empieza de nuevo por los prerrequisitos |

La regla de la tarea obliga a detenerse ahí sin sustituir la orden, y así se hizo. Los escenarios 4 a 16 no se
ejecutaron. La limpieza de la guía borró la carpeta temporal («carpeta temporal borrada»).

## Causa

La lista de permitidos comprueba por separado cada etapa de una orden compuesta. Si una etapa termina en `\`, es decir, si
la línea acaba en continuación y la siguiente empieza por `|`, no casa con `Bash(rtk:*)` y pide aprobación. La
continuación dentro de una etapa sí pasa: el escenario 3, partido en tres líneas con el `|` en la última, se ejecutó.

Sondas inocuas en la misma sesión:

- `rtk proxy printenv HOME \` y, en la línea siguiente, `  | rtk proxy grep -c /` → pide aprobación, con el mismo mensaje.
- `rtk proxy printenv \` y, en la línea siguiente, `  HOME | rtk proxy grep -c /` → `1`.
- `rtk proxy env KITLEGAL_SONDA=valor \` y, en la línea siguiente,
  `  printenv KITLEGAL_SONDA 2>&1 | rtk proxy grep -c valor ; echo "fin de la tubería partida"` → `1` y «fin de la tubería
  partida». Es la forma de los escenarios 8 y 13.b, con redirección y `;`.

Siete órdenes de la guía tenían la forma rechazada: las de los escenarios 2, 5, 6, 8, 13.a, 13.b y 14 (salen con
`grep -n -B1 -E '^\s+\|'` sobre `quickstart.md`). Las formas de la guía se sondaron al escribirla, pero la tabla no decía
cómo partir una tubería en varias líneas y ninguna sonda lo comprobaba.

## Arreglo, dentro del directorio del feature

En `quickstart.md`:

- Las siete órdenes llevan el `|` en la misma línea que el final de la etapa anterior, como el escenario 3. Los tests,
  paquetes, patrones y filtros no cambian: solo se mueve el salto de línea. Tras el cambio, la misma búsqueda no
  encuentra ninguna.
- La tabla de formas gana la fila «Partir una orden larga en varias líneas», seguida de un párrafo que remite a esta nota.
- Las sondas de los prerrequisitos ganan la tercera, la de la tubería partida: la misma orden sondada arriba, con su
  resultado esperado. Las dos siguientes pasan a ser la 4 y la 5, y el escenario 1 cita la sonda 4. Ningún otro
  artefacto del feature cita las sondas por número.

No cambia nada del código, de `testdata/`, de `schemas/` ni de fuera del directorio del feature. Tampoco la línea de la
tarea en `tasks.md`: sus rutas siguen siendo las mismas.

## Verificación de este intento

- `make ci` en primer plano, después del arreglo: `ci: todos los controles en verde` (formato, lint con 0 hallazgos,
  `test` y `test-integration` con `-race`, `govulncheck` sin vulnerabilidades, `schema-check`, `gitleaks` sin
  filtraciones, `go mod verify` y `tidy -diff`).
- `git status --porcelain` con el filtro de la guía: ninguna línea fuera del directorio del feature.

## Qué hace el intento 2

Empezar otra vez por los prerrequisitos, que vacían la carpeta temporal, y ejecutar la guía corregida entera: escenarios 1
a 16 salvo el 15.b, con la sonda nueva incluida. Después, escribir en `gates/pr-h4.md` la evidencia que pide la tarea y
marcarla si `make ci` está en verde. Si otra orden pidiera aprobación, se vuelve a parar y a anotar aquí.

## Intento 2 (2026-09-13, 22:12–22:35): la guía entera en verde y la evidencia escrita

**Estado**: T036 **marcada `[X]`** tras `make ci` en primer plano en verde. Base del intento: `55cc107`, la misma que el
intento 1; el árbol llevaba además el `quickstart.md` corregido del intento 1, sin confirmar.

### Qué pasó

La guía se ejecutó desde los prerrequisitos, con cada orden tal cual, sin que ninguna pidiera aprobación:

| Paso | Resultado |
|---|---|
| Prerrequisitos | `go1.27.1 darwin/arm64`, rama `h4-applet-boe-puerto`, ninguna línea entre `git status` y «fin del estado», carpeta temporal vacía |
| Sondas 1-5 | `código 3`; `valor`; `1` y «fin de la tubería partida» (la sonda nueva, con la forma que el intento 1 fijó); dos huellas distintas (`2292a1c0…`, `3477ca7a…`); nada entre la segunda huella y «fin de las sondas» |
| 1 | Huellas de `~/.cache/kitlegal` iguales (`519f40d4…`); `make test` con `ok` en los diez paquetes, sin `FAIL` ni carreras; nada en el `git status` de los datos protegidos |
| 2 | Seis `--- PASS` de `TestArticuloCoincideConBoePy/…`, ningún `FAIL`. La orden que en el intento 1 pidió aprobación se ejecutó sin pedirla |
| 3, 4, 5, 11, 12, 13.a, 14, 15.a | Todos los `--- PASS` que la guía espera (5; 1+1+9; 5; 5; 3; 4 y `make lint` con `0 issues.`; 5, la fila de `SOURCES.md` y `1`; 1), ningún `FAIL` |
| 6 | Diez `cronometra 200ms …` y `--- PASS: TestEntregaDelHito/boe-cache-rapida (0.41s)` |
| 7.a | `make schema-check` → `ok`; dos `--- PASS` |
| 7.b | Clon creado; `schemas/bloque.json \| 2 +-`, `1 file changed`; `schema-check` del clon falla nombrando `schemas/bloque.json` y «articulo» |
| 8 | `2 files changed`; `TestGolden` falla nombrando `articulo-BOE-A-2015-10565-a21` y la línea 10 |
| 9 | Fuzz `PASS` sin `Failing input` (387 058 ejecuciones, 11 s); tres `--- PASS` |
| 10 | Los ocho puntos esperados: ayuda, `1`, `código 2`, `código 4` por el enlace, `código 0`, las tres líneas de `--dry-run`, la caché no creada, la sonda de ausencia muda |
| 13.b | `tail -1` da la línea; el test falla nombrando `direcciones.go:72` |
| 16 | `ci: todos los controles en verde`; ninguna línea de `git status` fuera del directorio del feature |
| Limpieza | «carpeta temporal borrada» |

Los escenarios 2, 3 y 4 se lanzaron a la vez, igual que 5 y 7.a y que 7.b, 10, 11, 12, 13.a, 14 y 15.a (independientes
entre sí); el 6 fue solo, para que la medida de 200 ms no compartiera la máquina; 8, 9 y 13.b, en ese orden, sobre el
clon; el 16, al final.

### Evidencia escrita

`gates/pr-h4.md` completo con la plantilla del ritual: objetivo, alcance, dependencias (la sección de T026, intacta, más
la constatación de hoy: diecinueve líneas de `go list -deps`, `go.mod`, `go.sum` y `codecov.yml` sin diff frente a
`main`), controles añadidos, evidencia (la tabla de escenarios, la salida de `boe-cache-rapida`, los negativos y la
cobertura: global 95,2 % en el perfil unitario y 96,1 % en la unión, `internal/core` 90,1 %, `internal/cli` 98,6 %;
ningún umbral rebajado; `0` `//nolint` y `0` `t.Skip` en `.go`), decisiones y pendientes (S8 y S9 como pendientes de la
primera ejecución, T037, y un hallazgo: `refs/__pycache__/boe.cpython-311.pyc` versionado por el commit `[datos]`
`17fca5b`, fuera de las rutas de esta tarea y de cualquier otra, para la revisión humana antes de fusionar).

### Verificación de este intento

- `make ci` en primer plano tras escribir la evidencia: en verde (la misma cadena que el escenario 16, con `gitleaks`
  sobre el árbol que ya incluye `pr-h4.md`).
- `git status --porcelain` con el filtro de la guía: ninguna línea fuera del directorio del feature. Dentro de él, los
  dos ficheros de estado del workflow, `quickstart.md`, `pr-h4.md`, `tarea-T036.md` y `tasks.md` (la marca).
