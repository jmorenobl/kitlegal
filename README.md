# kitlegal

`kitlegal` es un conjunto de **skills agénticas** para consultar fuentes legales públicas españolas
—BOE, PLACSP, BDNS, BORME, EUR-Lex…— y actuar en tu municipio, sea cual sea: qué contrata y subvenciona
tu ayuntamiento, qué dicen sus ordenanzas, cuándo vence un plazo, qué escrito presentar. Las skills
razonan; las herramientas deterministas que usan las da un binario Go **multicall**, que también se
puede usar desde la línea de órdenes. Un solo ejecutable: el applet lo elige `os.Args[0]` o el primer
argumento, de modo que `kitlegal boe articulo …` y un symlink `boe -> kitlegal` son la misma cosa.
Toda respuesta va envuelta en `{ok, fuente, url, fecha_consulta, hash, data}`: sin fuente, URL, fecha
de consulta y hash no hay cita, y sin cita no hay respuesta.

Solo se automatizan fuentes públicas. Cualquier acción que exija identidad —presentar un escrito,
recoger una notificación— termina en un fichero listo para firmar, nunca en un envío a una sede.

## Qué entrega este hito (H4)

H4 trae el **primer applet con fuente**: `boe`, que consulta la API de Legislación Consolidada del BOE
—cualquier norma consolidada, estatal o autonómica, por su identificador `BOE-A-…`— y devuelve cada
respuesta lista para citar, con `fuente` `boe.legislacion-consolidada`, la `url` de la API consultada y
la `fecha_consulta` en que se obtuvo lo que `data` contiene.

```console
$ ./bin/kitlegal boe --help
uso: boe <verbo> [banderas]

Consulta la legislación consolidada del BOE y la devuelve lista para citar.

verbos:
  buscar     Busca normas consolidadas por las palabras de su título o con una consulta de la fuente.
  indice     Devuelve los bloques de una norma consolidada, en el orden de la fuente.
  articulo   Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia.
  articulos  Devuelve el texto vigente de varios bloques de una norma, en el orden pedido.
  metadatos  Devuelve los datos de una norma y los avisos de su vigencia.
  analisis   Devuelve las materias, las notas y las referencias de una norma.

Las banderas de cada verbo, en «boe <verbo> --help».
```

- **`kitlegal boe articulo BOE-A-2015-10565 a21`** devuelve el texto vigente del artículo 21 de la Ley
  39/2015, su huella, su dirección pública y la ELI de la norma, con los avisos de su vigencia. Es el
  porte de `refs/boe.py`, la skill `boe-fiscal` en Python: nunca emite un texto cuya vigencia no ha
  podido comprobar.
- **Caché local**: la segunda consulta idéntica no sale a la red. Vive en `~/.cache/kitlegal/cache.db`
  (en otra carpeta con `KITLEGAL_CACHE_DIR`), con una vigencia por verbo —5 minutos para `buscar` y
  `metadatos`, 7 días para los demás—, y nunca guarda un fallo. Lo servido desde la caché conserva la
  `fecha_consulta` de cuando se pidió, no la de la invocación. `--offline` responde solo con lo guardado
  y vigente —si falta, código `4` sin pedir nada— y `--dry-run` describe en la salida de error cada
  petición que habría emitido, sin emitir ninguna.
- **Una fuente pública, pedida con cuidado**: cada petición se identifica, respeta el `robots.txt`, deja
  al menos un segundo con la anterior al mismo sitio y se reintenta ante fallos transitorios. Licencia,
  términos de uso y ritmo están en la fila de la fuente de [`docs/SOURCES.md`](docs/SOURCES.md). Los
  fallos salen con los códigos estables: `2` una norma o un bloque mal escritos, `3` lo que la fuente no
  tiene, `4` la fuente caída o una respuesta que ya no se sabe interpretar, `5` límite de peticiones o
  `robots.txt`; nunca `6`.
- **Contratos publicados**: `schemas/norma.json` (`buscar`, `indice`, `metadatos`, `analisis`) y
  `schemas/bloque.json` (`articulo`, `articulos`) describen la entrada y la salida de cada verbo. Se
  generan desde `--describe`, toda salida se valida contra ellos en los tests y `make schema-check`
  falla si divergen.

Debajo están los hitos de fundación, que `boe` hereda sin escribir nada de ellos. H1 es el **kernel de
la línea de órdenes**, que fija la forma de invocarse, de fallar y de citar de todos los applets:

- **Despacho multicall**: el applet lo elige el nombre de invocación —un enlace `boe -> kitlegal`
  ejecuta el applet `boe`— y, si ese nombre no está registrado, el primer argumento.
- **Ocho banderas globales** que ningún applet escribe: `--json`, `--timeout`, `--offline`,
  `--dry-run`, `--describe`, `--no-graph`, `--asunto` y `--verbose`.
- **Sobre de salida** `{ok, fuente, url, fecha_consulta, hash, data}` y **códigos de salida estables**
  (`0` ok · `2` args · `3` no encontrado · `4` fuente no disponible · `5` límite o TOS · `6` requiere
  identidad humana; `1` queda para el fallo inesperado), traducidos en un único punto.
- **`--describe`**: esquema JSON 2020-12 de la entrada y la salida de cada verbo, derivado por
  reflexión y sin nada escrito a mano, para que un agente descubra el contrato sin leer código.

H2 dejó `internal/httpx`, la única puerta a la red, y H3 la caché local en SQLite. H0 dejó el
**esqueleto del repositorio y sus controles**: el `Makefile` como única superficie de invocación, y los
controles de formato, análisis estático, tests con detector de carreras, vulnerabilidades conocidas,
secretos e integridad de los módulos.

El binario distribuido registra **un solo applet, `boe`**: los de ejemplo (`echo`, `contar`) viven solo
en los tests. Los applets de las demás fuentes (`placsp`, `bdns`…) llegan en hitos posteriores; el orden
está en [`docs/ROADMAP.md`](docs/ROADMAP.md).

## Prerrequisitos

Exactamente dos, y nada más:

| Prerrequisito | Comprobación |
|---|---|
| Go 1.21 o superior | `go version` |
| `git` | `git --version` |

**No hay que instalar ninguna herramienta de control.** `golangci-lint`, `govulncheck`, `gitleaks` y
`lefthook` se construyen solos, con la versión fijada en `tools/<herramienta>/go.mod`, la primera vez
que se invoca la orden que los usa.

**Tampoco hay que instalar un parche concreto de Go, ni importa cuál tengas.** `go.mod` declara la
directiva `toolchain` y el `Makefile` exporta `GOTOOLCHAIN` con ese valor, así que todas las órdenes
se ejecutan con ese parche exacto —el mismo que ejecuta la integración continua— y el go command lo
descarga y lo verifica solo si falta. Cualquier `go` ≥ 1.21 sirve.

```bash
git clone https://github.com/jmorenobl/kitlegal.git
cd kitlegal
make check-tools     # comprueba go, git y que el toolchain fijado es obtenible
```

> La **primera** ejecución compila las herramientas desde fuente y, si el parche fijado no está en la
> caché, lo descarga: requiere red y tarda varios minutos. Las siguientes las sirve la caché de
> construcción de Go en segundos.

## Construir e instalar

```bash
make build       # deja el ejecutable en bin/kitlegal
make install     # lo instala en el directorio de binarios de Go ($GOBIN, o $HOME/go/bin)
```

Las dos inyectan los mismos datos de construcción —versión, commit y fecha— y compilan sin cgo y con
`-trimpath`.

## `kitlegal version`

Verbo reservado del kernel, que se reconoce antes que el registro de applets. Sin banderas y sin
subverbos:

```console
$ ./bin/kitlegal version
kitlegal a3dee64-dirty
commit: a3dee64269ee98da45b2f0a96202899fdaac9354
fecha:  2026-09-10T19:27:47Z
```

Tres líneas en la salida estándar, salida de error vacía y código de salida `0`. La versión sale de
`git describe --tags --always --dirty`: mientras no haya ninguna etiqueta es el commit abreviado, con
el sufijo `-dirty` si el árbol tiene cambios sin confirmar; en cuanto exista `v0.1.0` pasará a leerse
`v0.1.0-3-ga3dee64`. El commit coincide carácter a carácter con `git rev-parse HEAD` de la revisión
construida y la fecha es el instante de construcción en UTC. Un binario hecho con `make build` o
`make install` nunca imprime los valores por defecto del código (`dev`, `none`, `unknown`); si los
ves, lo estás ejecutando con `go run`.

Cualquier otra invocación —sin applet, con un applet desconocido o con un argumento sobrante tras
`version`— escribe el fallo en la salida de error, nombrando lo que no ha reconocido y enumerando los
applets que existen, y termina con código `2`, que es «args» en la tabla de códigos de salida
estables del proyecto:

```console
$ ./bin/kitlegal inventado
argumentos inválidos: "inventado" no es ningún applet de kitlegal; applets disponibles: boe
$ echo $?
2
```

`kitlegal --help` describe el uso y enumera los applets registrados, con código `0`.

## Ejecutar los controles

`make ci` es **el veredicto del repositorio**: encadena los nueve controles, falla nombrando el que
falla y no modifica ningún fichero versionado del árbol de trabajo.

```bash
make ci
```

La integración continua ejecuta esta misma orden, así que el veredicto local predice el de la
propuesta de cambio. Cada control se puede invocar por separado mientras se depura:

| Orden | Qué hace | ¿En `ci`? |
|---|---|---|
| `make fmt-check` | Comprueba el formato sin tocar ningún fichero | sí |
| `make lint` | Análisis estático completo, `gosec` incluido | sí |
| `make test` | Tests unitarios con detector de carreras; deja `coverage.out` | sí |
| `make test-integration` | Tests con la etiqueta de compilación `integration`, con detector de carreras: los que dependen del entorno (permisos del sistema de ficheros, dos procesos), siempre dentro de directorios temporales | sí |
| `make vuln` | Vulnerabilidades conocidas (consulta la base de datos de Go: requiere red) | sí |
| `make schema-check` | Comprueba que los esquemas publicados en `schemas/` (`norma.json` y `bloque.json`) coinciden con lo que emite `--describe` de cada verbo, sin escribir nada | sí |
| `make secrets` | Detección de secretos en todo el árbol | sí |
| `make mod-verify` | Integridad del módulo raíz y de cada módulo de herramienta | sí |
| `make mod-tidy-check` | Comprueba que `go.mod` y `go.sum` están saneados | sí |
| `make fmt` | **Corrige** el formato; por eso no forma parte de `ci` | no |
| `make lint-fast` | Análisis estático rápido, el del gancho de pre-commit | no |
| `make test-e2e` | Tests de extremo a extremo con `testscript`, contra un binario que registra los applets de ejemplo y `boe`, que responde desde sus grabaciones sin red | no |
| `make verify-sources` | Comprueba contra la fuente real que sus respuestas se siguen interpretando —hoy, `boe articulo` contra la API del BOE—: **requiere red** y lo ejecuta el flujo nocturno | no |

`make verify-sources` es el único control que pide algo a una fuente real, y por eso no está en `ci`:
todos los tests de `make ci` corren sin red, contra respuestas grabadas. Cada noche lo ejecuta el
trabajo `fuentes` del flujo `nightly`, que abre o comenta una incidencia si falla; el detalle está en
[`CONTRIBUTING.md`](CONTRIBUTING.md).

`make help` —el objetivo por defecto— enumera todas las órdenes, incluidas las que existen pero
reciben su contenido en un hito posterior (`skills-sync`, `release`).

### Ganchos de pre-commit

```bash
make hooks     # instala los ganchos con lefthook
```

Una vez instalados, cada commit corrige el formato —y vuelve a preparar lo que haya corregido, de
modo que se confirma ya formateado— y ejecuta `lint-fast`, `secrets` y `mod-tidy-check`. El gancho
**no es la autoridad final**: la integración continua vuelve a ejecutar
los mismos controles, y `make ci` es lo que hay que ver en verde antes de abrir una propuesta de
cambio.

## Cómo contribuir

El ritual por hito, el catálogo completo de controles, la convención de commits y las reglas sobre
dependencias nuevas están en [`CONTRIBUTING.md`](CONTRIBUTING.md). Las decisiones de arquitectura ya
cerradas están registradas en [`docs/ADR/`](docs/ADR/).

## Licencia

Apache-2.0. Ver [`LICENSE`](LICENSE).
