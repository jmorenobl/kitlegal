# Quickstart: validación de H1

Guía **ejecutable** para comprobar que H1 entrega lo que dice. Cada escenario se ejecuta tal cual, desde
la raíz del repositorio, sobre la rama `h1-kernel-cli-multicall` y **con el hito ya implementado**: esto
no es la implementación, es cómo se verifica.

**Sin efectos colaterales.** Ningún escenario crea ni modifica un fichero versionado del árbol de trabajo.
Los escenarios **8 y 9**, que necesitan introducir violaciones deliberadas, lo hacen sobre una **copia
desechable del árbol fuera del repositorio** (`$(mktemp -d)`, sin `.git`), que se borra al terminar: el
árbol de trabajo queda limpio **por construcción**, no por una limpieza cuidadosa, y en ningún instante
hay un fichero de más bajo `internal/` ni bajo `internal/app/testdata/` —que el guardián de diff protege
como material de test (plan.md, obligaciones 2, 3 y 8)—. El escenario 12 solo produce `bin/`, que está en
`.gitignore` desde H0, igual que `coverage.out`. Todos cierran con `git status --porcelain` vacío.
**Ningún escenario toca el índice ni el historial de git, ni escribe fuera del repositorio salvo en
directorios temporales que él mismo borra.**

## Prerrequisitos

```bash
go version            # go1.27.1 (el parche que fija la directiva `toolchain` de go.mod)
git --version
make --version
python3 --version     # solo para verificar: analiza el JSON de los escenarios 1-7
git rev-parse --abbrev-ref HEAD   # h1-kernel-cli-multicall
git status --porcelain            # vacío antes de empezar
```

`python3` **no es dependencia del producto**: lo usan los escenarios 1, 2, 3, 4, 5, 6 y 7 únicamente para
analizar el JSON emitido y demostrar que no queda texto fuera del documento —comprobación que no puede
hacerse con `grep` sin debilitarla—. Viene de serie en `ubuntu-latest` y en macOS con las herramientas de
línea de órdenes de Xcode. Lo que cubre en CI no es este guion sino los tests, que no lo necesitan.

Los bloques se ejecutan con **`bash` o `zsh`**: varios usan sustitución de procesos (`diff <(…) <(…)`),
que no existe en un `sh` estricto. Fuera de eso, solo herramientas POSIX (`ln`, `diff`, `mktemp`, `wc`,
`grep`, `test`).

Ninguna de las comprobaciones de abajo necesita red, salvo la primera ejecución de `make ci`, que descarga
las dependencias y compila las herramientas si la caché está fría (`make vuln` sí consulta la base de
vulnerabilidades).

**El binario del e2e.** La entrega literal del hito se demuestra contra un binario que **registra los
applets de ejemplo**; el binario distribuido no los registra (FR-009). Varios escenarios lo construyen
así, y lo borran al terminar:

```bash
E2E=$(mktemp -d)
go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e
# … escenario …
rm -rf "$E2E"
```

---

## Escenario 1 — Entrega literal: el sobre citable (US1 · SC-001)

```bash
E2E=$(mktemp -d)
go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e

"$E2E/kitlegal" echo hola --json
echo "código de salida: $?"
```

**Esperado**: un único documento JSON en la salida estándar, con exactamente las seis claves del sobre,
`ok` verdadero y `data` con lo pedido; código de salida `0`.

La invocación no nombra verbo: `echo` declara `repetir` como **verbo por omisión** y el kernel lo inserta
antes de analizar la gramática ([research.md D26](./research.md#d26--verbo-por-omisión-cómo-funciona-kitlegal-echo-hola-sin-nombrar-verbo)).
Nombrarlo explícitamente da exactamente lo mismo:

```bash
diff <("$E2E/kitlegal" echo hola --json         | python3 -c 'import json,sys;d=json.load(sys.stdin);d.pop("fecha_consulta");print(sorted(d.items()))') \
     <("$E2E/kitlegal" echo repetir hola --json | python3 -c 'import json,sys;d=json.load(sys.stdin);d.pop("fecha_consulta");print(sorted(d.items()))') \
  && echo "el verbo por omisión no cambia el resultado"
```

Comprobación mecánica de las seis claves y de que no hay una séptima:

```bash
"$E2E/kitlegal" echo hola --json | python3 -c 'import json,sys; d=json.load(sys.stdin); print(sorted(d))'
# ['data', 'fecha_consulta', 'fuente', 'hash', 'ok', 'url']
```

Y la procedencia y la marca temporal:

```bash
"$E2E/kitlegal" echo hola --json \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); print(d["fuente"], d["url"], d["fecha_consulta"], d["hash"][:14])'
# kitlegal.echo kitlegal:applet/echo 2026-…+02:00 sha256:…
```

La misma invocación **sin** `--json` presenta la tabla mínima, que conserva los cuatro datos de
procedencia (FR-043):

```bash
"$E2E/kitlegal" echo hola
rm -rf "$E2E"
```

**Esperado**: líneas `fuente`, `url`, `fecha_consulta` y `hash`, y después el contenido.

---

## Escenario 2 — Con `--json`, en stdout solo hay JSON (SC-002, criterio literal del hito)

El caso duro: el registro de eventos al máximo detalle **no** puede ensuciar la salida estándar.

```bash
E2E=$(mktemp -d)
go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e

KITLEGAL_LOG=debug "$E2E/kitlegal" echo hola --json --verbose > "$E2E/salida.json" 2> "$E2E/errores.txt"
echo "código: $?"

# el documento se analiza por completo y no queda texto fuera de él
python3 -c 'import json;json.load(open("'"$E2E"'/salida.json"));print("stdout es JSON completo y nada más")'

# y sí hubo registro, en la salida de error
test -s "$E2E/errores.txt" && echo "stderr contiene el registro de eventos"
rm -rf "$E2E"
```

**Esperado**: el análisis del JSON no falla (no hay prefijo ni sufijo), y `stderr` no está vacío.

---

## Escenario 3 — Un binario, muchos nombres (US2 · SC-003)

```bash
E2E=$(mktemp -d)
go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e
cd "$E2E"

ln -s kitlegal echo
./echo hola --json           > enlace.json  2> enlace.err;  echo "enlace: $?"
./kitlegal echo hola --json  > argumento.json 2> argumento.err; echo "argumento: $?"

# idénticos salvo lo que depende del instante de la consulta
diff <(python3 -c 'import json;d=json.load(open("enlace.json"));d.pop("fecha_consulta");print(sorted(d.items()))') \
     <(python3 -c 'import json;d=json.load(open("argumento.json"));d.pop("fecha_consulta");print(sorted(d.items()))') \
  && echo "salida estándar idéntica"
diff enlace.err argumento.err && echo "salida de error idéntica"

cd - >/dev/null
rm -rf "$E2E"
```

**Esperado**: mismo código de salida, misma salida de error y misma salida estándar salvo
`fecha_consulta`.

Los casos de contorno del despacho, en la misma sesión:

| Invocación | Esperado |
|---|---|
| `./kitlegal` (sin argumentos) | código `2`, lista de applets disponibles en **stderr** |
| `./kitlegal noexiste` | código `2`, mensaje que **nombra** `noexiste`, lista en stderr |
| `ln -s kitlegal kitlegal-dev && ./kitlegal-dev echo hola` | funciona: el nombre no registrado se trata como invocación por nombre propio |
| `./echo boe` | ejecuta el applet `echo` con el argumento `boe`: el nombre de invocación manda (FR-004) |
| `./kitlegal <segundo applet> hola` | código `2`: ese applet declara dos verbos y **ninguno por omisión**, así que nombrar el verbo es obligatorio |

---

## Escenario 4 — La huella depende solo del contenido (SC-005)

```bash
E2E=$(mktemp -d)
go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e

h1=$("$E2E/kitlegal" echo hola --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["hash"])')
sleep 1
h2=$("$E2E/kitlegal" echo hola --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["hash"])')
h3=$("$E2E/kitlegal" echo holb --json | python3 -c 'import json,sys;print(json.load(sys.stdin)["hash"])')

[ "$h1" = "$h2" ] && echo "mismo contenido, misma huella (y distinto instante)"
[ "$h1" != "$h3" ] && echo "un byte de diferencia, huella distinta"
rm -rf "$E2E"
```

Las propiedades que un script no puede ejercer —independencia del orden de las claves, números grandes sin
pérdida de precisión— están cubiertas por los tests unitarios:

```bash
go test ./internal/core/schema/ -run TestHuella -v
```

---

## Escenario 5 — Los fallos se distinguen sin leer el mensaje (US3 · SC-006 · SC-014 · SC-015)

Los seis códigos y sus sobres de fallo se fuerzan desde el applet de prueba y se comprueban en tabla. La
tabla vive en la raíz de composición del kernel —que es quien traduce el error a código de salida—, no en
`internal/cli`:

```bash
go test ./internal/app/ -run 'TestCodigoSalida|TestSobreDeFallo' -v
```

**Esperado**: un caso por cada código —`0`, `1`, `2`, `3`, `4`, `5`, `6`— y, para los seis de fallo, la
comprobación de que la salida estándar lleva el sobre de seis claves con `ok` falso, la clase y el mensaje
dentro de `data`, y `fuente`/`url` no vacías.

Los dos casos **anteriores a la ejecución del applet** se pueden ejercer a mano. Son los que obligan a que
el kernel sepa que se pidió `--json` **antes** de haber podido analizar la línea de órdenes: lo aporta el
pre-escaneo acotado de
[research.md D25](./research.md#d25--pre-escaneo-acotado-de-argv-antes-de-la-gramática).

```bash
E2E=$(mktemp -d)
go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e

# bandera desconocida
"$E2E/kitlegal" echo hola --jsno --json > "$E2E/a.json" 2> "$E2E/a.err"; echo "código: $?"   # 2
python3 -c 'import json;d=json.load(open("'"$E2E"'/a.json"));print(d["ok"], d["fuente"], d["url"], d["data"]["clase"])'
# False kitlegal.cli kitlegal:cli argumentos

# applet no registrado
"$E2E/kitlegal" noexiste --json > "$E2E/b.json" 2> "$E2E/b.err"; echo "código: $?"           # 2
python3 -c 'import json;d=json.load(open("'"$E2E"'/b.json"));print(d["ok"], d["data"]["clase"])'
# False argumentos

# el mensaje para la persona va a stderr, nunca a stdout
test -s "$E2E/b.err" && echo "mensaje en stderr"
rm -rf "$E2E"
```

Y `--timeout` inválido, que también es error de argumentos. Aquí **sin** `--json`, lo que además comprueba
la otra mitad de la regla: un fallo sin salida legible por máquina deja la **salida estándar vacía** y solo
escribe el mensaje en la de error (contrato del sobre §5):

```bash
E2E=$(mktemp -d); go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e
for t in 0s -1s abc; do
  "$E2E/kitlegal" echo hola --timeout "$t" > "$E2E/out.txt" 2> "$E2E/err.txt"
  echo "--timeout $t → código: $?"                                   # 2
  test ! -s "$E2E/out.txt" && test -s "$E2E/err.txt" \
    && echo "  stdout vacío, mensaje en stderr"
done
rm -rf "$E2E"
```

Y el caso de la tubería cerrada —escribir en stdout puede fallar, y eso es un código de salida, nunca un
pánico— con un escritor que siempre falla, que un guion no puede simular de forma portable:

```bash
go test ./internal/render/ -run TestEscrituraFallida -v
go test ./internal/app/  -run 'TestCodigoSalida/inesperado' -v
```

**Esperado**: el primero, que el fallo de escritura se **propaga**, que no se intenta un segundo sobre por
el descriptor roto y que no hay pánico; el segundo, que ese error propagado sale con código `1` —la mitad
del contrato que vive en la traducción única, no en el presentador—.

Y el plazo agotado, la otra mitad de la regla de `--timeout`: no solo un applet que devuelve el error
tipado produce el código `4`; **vencer el plazo también**. Lo fuerza un applet de prueba que tarda más que
el plazo que se le da, lo que un guion no puede hacer con `echo`, que responde al instante:

```bash
go test ./internal/app/ -run TestPlazoAgotado -v
```

**Esperado**: código `4` y sobre de fallo con la clase de fuente no disponible, igual que si el applet
hubiera devuelto el error tipado.

---

## Escenario 6 — El applet se describe a sí mismo (US5 · SC-007)

```bash
E2E=$(mktemp -d)
go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e

"$E2E/kitlegal" echo hola --describe > "$E2E/esquema.json"; echo "código: $?"   # 0
python3 -c 'import json;d=json.load(open("'"$E2E"'/esquema.json"));print(d["$schema"]);print(sorted(d["properties"]))'
# https://json-schema.org/draft/2020-12/schema
# ['entrada', 'salida']
rm -rf "$E2E"
```

**Esperado**: esquema del borrador 2020-12, con las dos partes, y código `0` **sin haber ejecutado** el
applet (la salida no contiene ningún sobre).

La invocación lleva `hola` y no nombra verbo, así que `--describe` describe el **verbo por omisión** de
`echo`: la autodescripción siempre se refiere a un verbo concreto, el nombrado o el de omisión.

Que el esquema sea válido para un validador y que el sobre de fallo valide contra él (SC-015) lo
comprueban los tests de contrato:

```bash
go test ./internal/cli/ -run 'TestDescribe|TestContrato' -v
```

---

## Escenario 7 — Las mismas banderas en todos los applets (US4 · SC-010)

Las ocho banderas, aceptadas por un applet que no declara ninguna:

```bash
E2E=$(mktemp -d); go build -o "$E2E/kitlegal" ./internal/app/testdata/kitlegal-e2e
"$E2E/kitlegal" echo hola --json --timeout 5s --offline --no-graph --asunto demo --verbose
echo "código: $?"   # 0
```

`--dry-run`: salida estándar **vacía**, descripción en la salida de error, código `0` —también con
`--json` y sin `--verbose`. La descripción **no** es un registro de eventos: la escribe el presentador, de
modo que ningún nivel puede ocultarla ([research.md D10](./research.md#d10---dry-run-no-corta-viaja-en-el-contexto-y-la-descripción-va-a-stderr)):

```bash
"$E2E/kitlegal" echo hola --dry-run --json > "$E2E/out.txt" 2> "$E2E/err.txt"
echo "código: $?"                       # 0
test ! -s "$E2E/out.txt" && echo "stdout vacío"
test -s "$E2E/err.txt"  && echo "descripción en stderr, visible sin --verbose"

# el caso duro: con el registro de eventos apagado hasta el máximo, la descripción SIGUE estando
KITLEGAL_LOG=error "$E2E/kitlegal" echo hola --dry-run > "$E2E/out2.txt" 2> "$E2E/err2.txt"
echo "código: $?"                       # 0
test ! -s "$E2E/out2.txt" && echo "stdout vacío"
test -s "$E2E/err2.txt"  && echo "descripción en stderr, que ningún nivel de registro puede ocultar"
```

Las tres reglas de `--dry-run` —salida estándar vacía **también con `--json`**, descripción en la de error
**visible con el registro apagado hasta el máximo** y código `0`— no se quedan en esta comprobación manual:
las cubre en tabla el test de la raíz de composición, que es lo que hace que una regresión se vea en
`make ci` y no solo aquí:

```bash
go test ./internal/app/ -run TestDryRun -v
```

Y que el registro de eventos no lleve los argumentos salvo en el nivel de depuración (FR-039):

```bash
go test ./internal/cli/ -run TestRegistroPrivacidad -v
```

`--verbose` y `KITLEGAL_LOG` cambian el detalle del registro **sin alterar la salida estándar**:

```bash
"$E2E/kitlegal" echo hola --json                  > "$E2E/n.json" 2> "$E2E/n.err"
"$E2E/kitlegal" echo hola --json --verbose        > "$E2E/v.json" 2> "$E2E/v.err"
KITLEGAL_LOG=debug "$E2E/kitlegal" echo hola --json > "$E2E/e.json" 2> "$E2E/e.err"
diff <(python3 -c 'import json;d=json.load(open("'"$E2E"'/n.json"));d.pop("fecha_consulta");print(sorted(d.items()))') \
     <(python3 -c 'import json;d=json.load(open("'"$E2E"'/v.json"));d.pop("fecha_consulta");print(sorted(d.items()))') \
  && echo "--verbose no toca stdout"
[ "$(wc -c < "$E2E/v.err")" -gt "$(wc -c < "$E2E/n.err")" ] && echo "--verbose sí añade detalle en stderr"
[ "$(wc -c < "$E2E/e.err")" -gt "$(wc -c < "$E2E/n.err")" ] && echo "KITLEGAL_LOG también, sin ninguna bandera"
```

`--json --help` se comporta igual que `--help` solo, en cualquier orden (respuesta Q5 del `clarify`):

```bash
"$E2E/kitlegal" echo --help          > "$E2E/h1.txt"; echo "código: $?"   # 0
"$E2E/kitlegal" echo --json --help   > "$E2E/h2.txt"; echo "código: $?"   # 0
"$E2E/kitlegal" echo --help --json   > "$E2E/h3.txt"; echo "código: $?"   # 0
diff "$E2E/h1.txt" "$E2E/h2.txt" && diff "$E2E/h1.txt" "$E2E/h3.txt" && echo "la ayuda es texto, y --json no la altera"
grep -q repetir "$E2E/h1.txt" && echo "la ayuda del applet enumera sus verbos"
rm -rf "$E2E"
```

**Esperado**: las tres salidas son idénticas y enumeran los **verbos del applet**. Pedir la ayuda de un
applet no resuelve ningún verbo: es el único caso en que la normalización del verbo por omisión se
suprime, para que `kitlegal echo --help` no acabe describiendo un solo verbo
([`contracts/registro-y-describe.md`](./contracts/registro-y-describe.md) §2 bis).

Que el segundo applet de ejemplo hereda exactamente lo mismo **sin código duplicado** lo comprueba el
test:

```bash
go test ./internal/app/ -run TestAppletHereda -v
```

---

## Escenario 8 — Las reglas de arquitectura fallan si se violan (US6 · SC-008)

**Nada de esto ocurre dentro del repositorio.** El escenario trabaja sobre una **copia desechable del
árbol**, fuera de él y sin `.git`: es lo que permite introducir seis ficheros y un directorio de violación
sin que el árbol de trabajo se ensucie en ningún momento —incluido `internal/app/testdata/`, que el
guardián de diff protege— y sin depender de acordarse de borrarlos (plan.md, obligación 3).

```bash
COPIA="$(mktemp -d)/kitlegal"
mkdir -p "$COPIA"
tar -cf - --exclude=./.git --exclude=./bin --exclude=./coverage.out . | tar -xf - -C "$COPIA"
cd "$COPIA"
```

La copia lleva `go.mod`, `Makefile`, `.golangci.yml` y los módulos de `tools/`, así que `make lint` y
`go test` funcionan exactamente igual que en el repositorio; `check-tools` no exige estar dentro de un
repositorio de git, y la caché de módulos es la misma, de modo que no hace falta red.

Las violaciones ejercen las cinco reglas —R1 en sus dos mitades, la de paquetes internos y la de I/O de la
biblioteca estándar— y **las dos capas en R1, R2 y R3**; R4 y R5 son reglas de **símbolo** y solo las ve el
lint, porque una llamada a `os.Exit` o una referencia a `os.Stdout` no aparecen en el grafo de
`go list -deps` ([`contracts/reglas-de-arquitectura.md`](./contracts/reglas-de-arquitectura.md) §2).

```bash
# R1 · el dominio no importa adaptadores (en un paquete nuevo, para que no haya ciclo de importación)
mkdir -p internal/core/prueba
cat > internal/core/prueba/violacion_r1.go <<'EOF'
package prueba

import "github.com/jmorenobl/kitlegal/internal/render"

// ViolacionR1 existe solo para que la regla R1 tenga algo que rechazar.
var ViolacionR1 = render.Nuevo
EOF

# R1 bis · el dominio tampoco importa I/O de la biblioteca estándar: es lo que impide
# que un *slog.Logger vuelva a instalarse dentro de schema.Contexto
cat > internal/core/prueba/violacion_r1_io.go <<'EOF'
package prueba

import "log/slog"

// ViolacionR1IO existe solo para que la parte de I/O de la regla R1 tenga algo que rechazar.
var ViolacionR1IO *slog.Logger
EOF

# R2 · solo internal/httpx importa net/http
cat > internal/cli/violacion_r2.go <<'EOF'
package cli

import "net/http"

// ViolacionR2 existe solo para que la regla R2 tenga algo que rechazar.
var ViolacionR2 = http.MethodGet
EOF

# R3 · solo cache, store y graph importan database/sql
cat > internal/cli/violacion_r3.go <<'EOF'
package cli

import "database/sql"

// ViolacionR3 existe solo para que la regla R3 tenga algo que rechazar.
var ViolacionR3 = sql.ErrNoRows
EOF

# R4 · solo internal/cli y cmd/ terminan el proceso
cat > internal/app/violacion_r4.go <<'EOF'
package app

import "os"

// ViolacionR4 existe solo para que la regla R4 tenga algo que rechazar.
func ViolacionR4() { os.Exit(0) }
EOF

# R5 · solo internal/render escribe en la salida estándar
cat > internal/render/violacion_r5.go <<'EOF'
package render

import "fmt"

// ViolacionR5 existe solo para que la regla R5 tenga algo que rechazar.
func ViolacionR5() { fmt.Println("esto no puede pasar el lint") }
EOF
```

Capa 1, el lint —debe fallar por **las cinco** y **nombrar la regla** en cada caso (`R1`/lista `core`,
lista `red`, lista `sql`, marca `R4:`, marcas `R5:` y `R5-descriptores:`):

```bash
make lint; echo "código de make lint: $?"    # distinto de 0
```

Capa 2, el test de arquitectura, **independiente de la configuración del lint** —debe fallar por R1, R2 y
R3, y **solo** por ellas, aunque no se haya tocado `.golangci.yml`:

```bash
go test ./internal/ -run TestArquitectura -v; echo "código: $?"    # distinto de 0
```

**Esperado**: el test nombra R1 (dos veces: paquete interno e I/O de la biblioteca estándar), R2 y R3, y
**no** menciona R4 ni R5, que no son observables en el grafo de dependencias. Que esas dos están vivas lo
demuestra el fallo del lint de la capa 1 y, en su alcance, el escenario 9.

Comprobación de que el control vuelve a pasar sobre el árbol sano, retirando una a una las violaciones
**dentro de la copia**:

```bash
rm -rf internal/core/prueba
rm -f internal/cli/violacion_r2.go internal/cli/violacion_r3.go \
      internal/app/violacion_r4.go internal/render/violacion_r5.go
make lint                                        # vuelve a pasar
go test ./internal/ -run TestArquitectura -v     # vuelve a pasar
```

La copia se conserva para el escenario 9, que sigue justo aquí y parte del mismo árbol sano.

---

## Escenario 9 — Los applets de ejemplo pasan el mismo listón, y la excepción del e2e no se desborda (D19)

Los comodines de Go **no descienden a `testdata`**: si los paquetes de ejemplo no estuvieran enumerados
explícitamente en el `Makefile`, quedarían sin lintar sin que nadie se enterase. Este escenario comprueba
que sí lo están **y** que la única excepción de lint que H1 concede —la del `package main` del binario de
e2e, que necesita `os.Exit` y los descriptores por ser una raíz de composición— está acotada a esa ruta.

**Se ejecuta sobre la misma copia desechable del escenario 8**, ya limpia de violaciones. Es la única
forma admisible: los ficheros de este escenario cuelgan de `internal/app/testdata/`, que la capa 3 de la
constitución protege, y crearlos en el árbol de trabajo convertiría la validación en una tarea `[datos]`
con pausa de revisión humana (plan.md, obligaciones 2 y 8). Si se llega aquí sin haber hecho el escenario
8, la copia se prepara igual:

```bash
# solo si no se viene del escenario 8, y desde la raíz del repositorio
COPIA="$(mktemp -d)/kitlegal"; mkdir -p "$COPIA"
tar -cf - --exclude=./.git --exclude=./bin --exclude=./coverage.out . | tar -xf - -C "$COPIA"
cd "$COPIA"
```

```bash
cat > internal/app/testdata/ejemplo/violacion.go <<'EOF'
package ejemplo

import "fmt"

// Violacion existe solo para comprobar que este directorio sí se linta.
func Violacion() { fmt.Println("bajo testdata también se linta") }
EOF

make lint; echo "código: $?"      # distinto de 0, nombrando `R5:` sobre este fichero

rm -f internal/app/testdata/ejemplo/violacion.go
```

La excepción es del paquete vecino, no de `testdata/`. Lo mismo que **pasa** en `kitlegal-e2e` debe
**fallar** en `ejemplo`:

```bash
cat > internal/app/testdata/ejemplo/violacion_exit.go <<'EOF'
package ejemplo

import "os"

// ViolacionExit existe solo para comprobar que la excepción de la raíz de composición
// no alcanza a este paquete.
func ViolacionExit() { os.Exit(0) }
EOF

make lint; echo "código: $?"      # distinto de 0, nombrando `R4:` sobre este fichero

rm -f internal/app/testdata/ejemplo/violacion_exit.go
```

Y al revés: el `main` del e2e usa `os.Exit` y `os.Stdout`/`os.Stderr` **sin ninguna directiva de
supresión**, y el lint pasa. Que no haya ninguna es comprobable:

```bash
grep -rn "nolint" internal/ cmd/ ; echo "sin //nolint: $?"   # 1 = ninguna coincidencia
make lint; echo "código: $?"      # 0
```

Y se destruye la copia, que es todo lo que hubo que limpiar:

```bash
cd - >/dev/null
rm -rf "$(dirname "$COPIA")"
git status --porcelain            # vacío: el árbol de trabajo nunca llegó a tocarse
```

**Si el primer `make lint` pasara**, la causa es que `golangci-lint` está excluyendo `testdata/`: se
declara la ruta en `.golangci.yml` para que deje de hacerlo. **No** se renuncia al control. **Si el
segundo pasara**, la excepción está escrita por ruta demasiado ancha o sin la marca de la regla: se
estrecha hasta que falle.

---

## Escenario 10 — Cobertura de `internal/cli` ≥ 90 % (SC-004, criterio literal del hito)

```bash
go test -cover ./internal/cli/
# ok  github.com/jmorenobl/kitlegal/internal/cli  0.0Xs  coverage: 9X.X% of statements
```

Y los dos umbrales que H0 dejó vigentes:

```bash
go test -cover ./internal/core/...   # ≥ 85 %
make test && go tool cover -func=coverage.out | tail -1   # total global ≥ 70 %
```

Los tres umbrales están declarados como estados **bloqueantes** en `codecov.yml`, sin exclusiones:
`internal/cli` ≥ 90 %, `internal/core` ≥ 85 %, global ≥ 70 %.

---

## Escenario 11 — El veredicto agregado, con el árbol intacto (SC-013)

```bash
make ci
echo "código: $?"          # 0
git status --porcelain     # vacío: `make ci` no modifica ningún fichero versionado
```

**Esperado**: `ci: todos los controles en verde`, incluidos los que H1 añade —test de arquitectura,
contratos del sobre, fixtures de códigos de salida y e2e— y **todos** los de H0 sin exclusiones nuevas
(FR-057).

El e2e por separado, que en H0 solo anunciaba el hito que lo aportaría:

```bash
make test-e2e
```

**Esperado**: ejecuta los guiones de `internal/app/testdata/script/` —ayuda, enlace simbólico, código 2 con
argumentos inválidos y salida JSON analizable— contra un binario construido por el propio test, sin red y
sin escribir fuera de su directorio temporal (SC-009).

---

## Escenario 12 — El binario distribuido no registra `echo`, y `version` no cambió (FR-009 · D16)

```bash
make build      # produce bin/kitlegal, que está en .gitignore

./bin/kitlegal version
echo "código: $?"          # 0, y las mismas tres líneas que en H0

./bin/kitlegal echo hola
echo "código: $?"          # 2: `echo` no está registrado en el binario distribuido

cd bin && ln -s kitlegal echo && ./echo hola; echo "código: $?"   # 2, por la misma razón
rm -f echo && cd - >/dev/null
```

**Esperado**: `version` sigue imprimiendo versión, commit y fecha con el formato de
[`contracts/cli-version.md` de H0](../001-h0-esqueleto-del-repo/contracts/cli-version.md) —lo único que
cambió es que el texto sale por `internal/render`—, y `echo` termina con código 2 como cualquier otro
nombre no registrado.

```bash
git status --porcelain    # vacío
```

---

## Resumen de trazabilidad

| Escenario | Historia / criterio | Requisitos principales |
|---|---|---|
| 1 · Sobre citable | US1 · SC-001 | FR-010 … FR-017, FR-041, FR-043 |
| 2 · Solo JSON en stdout | SC-002 *(literal del hito)* | FR-036, FR-040, FR-042 |
| 3 · Enlace simbólico | US2 · SC-003 | FR-002 … FR-007 |
| 4 · Huella | SC-005 | FR-011, FR-012 |
| 5 · Códigos de salida y sobre de fallo | US3 · SC-006, SC-014, SC-015 | FR-020, FR-027, FR-029 … FR-035, FR-045 |
| 6 · `--describe` | US5 · SC-007 | FR-046 … FR-049 |
| 7 · Las ocho banderas | US4 · SC-010, SC-011 | FR-018 … FR-028, FR-037 |
| 8 · Reglas de arquitectura | US6 · SC-008 | FR-050 … FR-053 |
| 9 · Lint de los ejemplos y alcance de la excepción del e2e | — (criterio §1 de la constitución) | FR-009, FR-035, FR-040, FR-057 |
| 10 · Cobertura | SC-004 *(literal del hito)* | FR-056 |
| 11 · `make ci` y e2e | SC-009, SC-013 | FR-054, FR-055, FR-057 |
| 12 · Binario distribuido | FR-009 · SC-003 (segunda mitad) | FR-005, FR-006, FR-040 |

**Lo que ningún escenario ejercita, y por qué**: no hay red (FR-058) ni escritura en disco fuera de
temporales (FR-059) porque H1 no tiene fuentes ni almacenamiento; la prohibición de que un adaptador use
el espacio de nombres reservado `kitlegal.` / `kitlegal:` (FR-016) no tiene objeto hasta H4.
