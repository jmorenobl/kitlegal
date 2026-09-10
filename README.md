# kitlegal

`kitlegal` es un binario Go **multicall** y un conjunto de skills agénticas para consultar fuentes
legales públicas españolas —BOE, PLACSP, BDNS, BORME, EUR-Lex…— desde la línea de órdenes o desde un
agente. Un solo ejecutable: el applet lo elige `os.Args[0]` o el primer argumento, de modo que
`kitlegal boe articulo …` y un symlink `boe -> kitlegal` son la misma cosa. Toda respuesta va envuelta
en `{ok, fuente, url, fecha_consulta, hash, data}`: sin fuente, URL, fecha de consulta y hash no hay
cita, y sin cita no hay respuesta.

Solo se automatizan fuentes públicas. Cualquier acción que exija identidad —presentar un escrito,
recoger una notificación— termina en un fichero listo para firmar, nunca en un envío a una sede.

## Qué entrega este hito (H0)

H0 es el **esqueleto del repositorio y sus controles**: un repo todavía sin fuentes legales, pero
blindado, para que cualquier línea de Go que entre después pase por los mismos gates. Concretamente:

- El binario `kitlegal` con un único verbo, `version`, que imprime versión, commit y fecha de
  construcción.
- El `Makefile` como única superficie de invocación de los controles: la misma orden que ejecutas tú
  es la que ejecutan el gancho de pre-commit y la integración continua.
- Los controles activos: formato, análisis estático, tests con detector de carreras, vulnerabilidades
  conocidas, detección de secretos e integridad de los módulos.

Los applets de fuentes (`boe`, `placsp`, `bdns`…) llegan en hitos posteriores; el orden está en
[`docs/ROADMAP.md`](docs/ROADMAP.md).

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

Único verbo de este hito. Sin banderas y sin subverbos:

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

Cualquier otra invocación —sin verbo, con un verbo desconocido o con un argumento sobrante— escribe
una línea de uso en la salida de error y termina con código `2`, que es «args» en la tabla de códigos
de salida estables del proyecto:

```console
$ ./bin/kitlegal inventado
uso: kitlegal version
$ echo $?
2
```

## Ejecutar los controles

`make ci` es **el veredicto del repositorio**: encadena los ocho controles, falla nombrando el que
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
| `make vuln` | Vulnerabilidades conocidas (consulta la base de datos de Go: requiere red) | sí |
| `make schema-check` | Validación de salidas contra esquemas (los aportan H4 y H11) | sí |
| `make secrets` | Detección de secretos en todo el árbol | sí |
| `make mod-verify` | Integridad del módulo raíz y de cada módulo de herramienta | sí |
| `make mod-tidy-check` | Comprueba que `go.mod` y `go.sum` están saneados | sí |
| `make fmt` | **Corrige** el formato; por eso no forma parte de `ci` | no |
| `make lint-fast` | Análisis estático rápido, el del gancho de pre-commit | no |
| `make test-integration` | Tests con la etiqueta de compilación `integration` | no |

`make help` —el objetivo por defecto— enumera todas las órdenes, incluidas las que existen pero
reciben su contenido en un hito posterior (`test-e2e`, `skills-sync`, `release`).

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
