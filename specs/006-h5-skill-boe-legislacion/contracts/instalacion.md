# Contrato: `make install`

FR-050 a FR-055, US3, SC-006. Decisiones: research.md D5 (destino del enlace) y D15 (pruebas sobre un árbol mínimo).

## 1. Receta

```make
## install: instala kitlegal en el directorio de binarios de Go y enlaza las skills en ~/.claude/skills
install: check-tools
	scripts/instalar-skills.sh --comprobar
	CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/kitlegal
	scripts/instalar-skills.sh "$$(go list -f '{{.Target}}' ./cmd/kitlegal)"
```

- La primera línea busca los conflictos (§2, pasos 1 a 5) antes de instalar nada: con uno, `make install` falla sin
  crear ni cambiar nada, tampoco el binario del directorio de binarios de Go (data-model §11.1). Sin ella, el
  `go install` habría instalado o sobrescrito el binario antes de que el guion fallara.
- La segunda es la de H0 sin cambios: el binario va a `$GOBIN` o, si no está definido, a `$GOPATH/bin` o
  `$HOME/go/bin`, con los mismos datos de versión que `make build` (FR-050; `go help install`).
- `go list -f '{{.Target}}'` da la ruta de instalación que acaba de usar `go install` («Target string // install path»,
  `go help list`), así que el guion nunca calcula por su cuenta dónde quedó el binario.

## 2. `scripts/instalar-skills.sh --comprobar | <binario>`

Bash con `set -euo pipefail`, como el resto de `scripts/`. Sin Python ni ningún otro intérprete.

| Paso | Qué hace | Fallo |
|---|---|---|
| 1 | raíz = directorio padre del del guion, físico (`cd "$(dirname "$0")/.." && pwd -P`) | — |
| 2 | exige un argumento: `--comprobar` o un fichero ejecutable | `instalar-skills: uso: scripts/instalar-skills.sh --comprobar \| <binario instalado>` o `instalar-skills: <ruta> no existe o no es ejecutable`, código 1 |
| 3 | exige `HOME` no vacío | `instalar-skills: HOME no está definido`, código 1 |
| 4 | por cada directorio `skills/<nombre>/`, clasifica `$HOME/.claude/skills/<nombre>` (data-model §11.1) | — |
| 5 | si hay algún conflicto, escribe una línea por cada uno en la salida de error y termina **sin crear ni cambiar nada**; si no, con `--comprobar` termina aquí con 0 | `instalar-skills: conflicto: <ruta> ya existe y no es un enlace a <destino>; no se modifica`, código 1 |
| 6 | `mkdir -p <raíz>/bin/instalado` y `ln -sfn <binario> <raíz>/bin/instalado/kitlegal` | error de la orden, código distinto de 0 |
| 7 | `mkdir -p "$HOME/.claude/skills"`; crea con `ln -s` los enlaces que faltan | error de la orden |
| 8 | una línea por skill en la salida estándar: `instalar-skills: <nombre> → <raíz>/skills/<nombre>`, y otra para el binario | — |

- El destino de un enlace existente se lee con `readlink` sin `-f`, que es portable entre macOS y Linux y compara el
  destino literal.
- Repetir la instalación deja el mismo estado y termina con 0 (FR-053): los enlaces correctos se dejan y
  `bin/instalado/kitlegal` se rehace igual.

## 3. Resolución de `scripts/boe` tras instalar

`~/.claude/skills/boe-legislacion` → `<raíz>/skills/boe-legislacion`; dentro, `scripts/boe` →
`../../../bin/instalado/kitlegal`, que el sistema resuelve desde el directorio físico del enlace, es decir
`<raíz>/bin/instalado/kitlegal` → binario instalado. El nombre de invocación sigue siendo `…/scripts/boe`, así que el
despacho multicall elige `boe` (`nombreDeInvocacion`, `internal/app/despacho.go`). Comprobado en local con enlaces
equivalentes (research.md D5).

## 4. Pruebas (FR-055)

`TestInstalacion` (`internal/skills/instalacion_test.go`, `//go:build integration`, así que lo ejecuta
`make test-integration`, que está en `make ci`), con `testscript` sobre los cuatro guiones de
`internal/skills/testdata/script/` y sobre `instalar-con-enlace-roto`, que el propio test escribe desde la constante
`guionDelEnlaceRoto` en un `t.TempDir()` y ejecuta igual. Lleva la etiqueta porque ejecuta `make` y compila el binario:
la skill `golang-testing` pide separar con etiqueta lo que no es un test unitario rápido (research.md D21).

**Nunca sobre el repositorio real.** `Setup` copia en `$WORK/repo` un árbol mínimo que basta para `make install`:
`Makefile`, `go.mod`, `go.sum`, `scripts/instalar-skills.sh`, el directorio `skills/` entero con sus enlaces tal cual, y
los ficheros `.go` y embebidos de cada paquete del módulo del que depende `./cmd/kitlegal` (`go list -deps` con
`Dir`, `GoFiles` y `EmbedFiles`). Así ni `bin/instalado/` ni ningún otro fichero se escribe en el árbol de quien ejecuta
los tests.

**El código de la copia, por ruta física.** Cada fichero entra en la copia por su ruta dentro del repositorio, la del
`Dir` de su paquete relativa a la raíz, y un paquete del módulo fuera de la raíz hace fallar el test. La raíz y cada
`Dir` se comparan tras resolver sus enlaces simbólicos (`filepath.EvalSymlinks`), porque un mismo árbol se alcanza por
más de una ruta —en macOS, un clon bajo `/tmp` o bajo el temporal de `mktemp` cuelga de `/private`— y cada lado la
escribe a su manera: la raíz, por la del directorio de trabajo del test, que `go test` escribe por la ruta desde la que
se lanzó; `go list`, por la de su `PWD` si nombra su directorio de trabajo y por la física si no. Sin resolverlos,
`make ci` desde una ruta con un enlace fallaba con «el paquete … del módulo está fuera del repositorio» (revisión final,
ronda 3). `TestFicherosDelBinario`, en el mismo fichero y con la misma etiqueta, lo fija con un enlace de `t.TempDir()`
al repositorio: el enlace como raíz, con `go list` sin `PWD`, que da cada `Dir` por la ruta física
(`raiz-por-un-enlace`); y la raíz física con `PWD` en el enlace, que da cada `Dir` por el enlace (`dir-por-un-enlace`).
En los dos casos exige la misma lista que desde la ruta física con `go list` sin `PWD`: sin resolver los enlaces de la
raíz cae el primero, y sin resolver los de cada `Dir`, el segundo.

Entorno de cada guion:

| Variable | Valor | Por qué |
|---|---|---|
| `HOME` | `$WORK/home` | FR-055: nunca el directorio personal real |
| `GOBIN` | `$WORK/gobin` (vacía en el guion sin `GOBIN`) | FR-055: nunca el directorio de binarios real |
| `GOPATH` | `$WORK/gopath` | que el caso sin `GOBIN` no caiga en `$HOME/go/bin` real |
| `GOMODCACHE`, `GOCACHE` | los de `go env` del proceso de test | leer módulos y toolchain ya descargados y reutilizar la caché de construcción que el propio `go test` usa (misma práctica que el e2e de `internal/app`) |
| `GOPROXY` | `off` | ninguna descarga: si algo faltara, el guion falla en lugar de tocar la red |
| `GOENV` | `off` | no leer la configuración de Go de ninguna cuenta (`go help environment`) |
| `GOFLAGS` | `-mod=readonly` | no reescribir `go.mod` ni `go.sum` de la copia |

| Guion | Qué comprueba |
|---|---|
| `instalar.txtar` | `~/.claude/skills` no existe; `make -C $WORK/repo install` termina en 0; existe `$WORK/gobin/kitlegal`; `readlink` del enlace de la skill termina en `/repo/skills/boe-legislacion` (el guion
escribe la raíz **física**, y en macOS el temporal de `$WORK` cuelga de `/private`); `readlink $WORK/repo/bin/instalado/kitlegal`
da `$WORK/gobin/kitlegal`; `$HOME/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a21 --describe` termina en 0 con `"title": "boe articulo"` (US3-1, SC-006) |
| `instalar-de-nuevo.txtar` | dos instalaciones seguidas terminan en 0; mismo `readlink`; `ls $HOME/.claude/skills` da solo `boe-legislacion` (US3-2, FR-053) |
| `instalar-con-conflicto.txtar` | con un directorio con un fichero dentro, y en otra ronda con un enlace a otro sitio, en `$HOME/.claude/skills/boe-legislacion`: `make install` termina distinto de 0, la salida de error casa `conflicto: .*boe-legislacion`, y la entrada sigue igual (el fichero existe; el enlace ajeno conserva su destino) (US3-3, FR-054) |
| `instalar-sin-gobin.txtar` | sin `GOBIN`: el binario queda en `$WORK/gopath/bin/kitlegal`, `readlink $WORK/repo/bin/instalado/kitlegal` lo nombra y `scripts/boe … --describe` responde (FR-050, caso límite del spec) |
| `instalar-con-enlace-roto` (`guionDelEnlaceRoto`) | con un enlace a una ruta inexistente en `$HOME/.claude/skills/boe-legislacion`: `make install` termina distinto de 0, la salida de error casa `conflicto: .*boe-legislacion`, el enlace conserva su destino, y no existen ni `$WORK/gobin/kitlegal` ni `$WORK/repo/bin/instalado`, porque la receta busca los conflictos antes de `go install` (US3-3, FR-054; data-model §11.1) |

Los cuatro primeros guiones son ficheros nuevos bajo `testdata/`: van en una tarea `[datos]` (no provocan pausa por no
estar bajo `testdata/` de raíz, `internal/source/*` ni `schemas/`; `workflow.yml`, `clasificar_datos`). El del enlace
roto llegó con la revisión final, que no añade ficheros bajo `testdata/`, y por eso lo escribe el test, como las trazas
de `TestLeerTrazasSinTerminar`.
