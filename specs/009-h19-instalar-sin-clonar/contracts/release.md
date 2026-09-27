# Contrato: release, snapshot, CI e `install.sh`

Requisitos: FR-090 a FR-122. Decisiones en [../research.md](../research.md) D20 a D30; lo que no se puede comprobar
sin red o sin etiqueta, en S1-S5.

## 1. `tools/goreleaser/`

`tools/goreleaser/go.mod` (módulo `github.com/jmorenobl/kitlegal/tools/goreleaser`, `go 1.27.1`, `tool
github.com/goreleaser/goreleaser/v2`, v2.18.1) y su `go.sum`. Se invoca como las demás herramientas:

```make
GORELEASER := go tool -modfile=tools/goreleaser/go.mod goreleaser
```

`make mod-verify` lo cubre sin cambios (descubre `tools/*/go.mod`); `.github/dependabot.yml` gana su entrada `gomod`
para `/tools/goreleaser`, agrupada como las otras. syft y cosign **no** entran en `tools/` (FR-096).

## 2. `.goreleaser.yaml`

Sin ninguna propiedad obsoleta en v2.18.1 (V3): `goreleaser check` sale con 0 (FR-094). Lo que fija, y lo que
`TestConfiguracionDeLaRelease` comprueba en `make test`:

| Sección | Contenido | FR |
|---|---|---|
| `version`, `project_name` | `2`, `kitlegal` | — |
| `builds[0]` | `main: ./cmd/kitlegal`, `binary: kitlegal`, `env: [CGO_ENABLED=0]`, `flags: [-trimpath]`, `goos: [darwin, linux, windows]`, `goarch: [amd64, arm64]` | FR-090 |
| `builds[0].ldflags` | exactamente cuatro `-X`: `main.version={{ if .IsSnapshot }}{{ .Version }}{{ else }}{{ .Tag }}{{ end }}`, `main.commit={{ .FullCommit }}`, `main.fecha={{ .Date }}`, `github.com/jmorenobl/kitlegal/internal/httpx.version=` con la misma plantilla que `main.version`; los mismos cuatro símbolos que el `LDFLAGS` del `Makefile` | FR-091, FR-092 |
| `archives[0]` | `name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"`, `formats: [tar.gz]`, `format_overrides: [{goos: windows, formats: [zip]}]`; el binario en la raíz del archivo | FR-093 |
| `checksum` | `name_template: checksums.txt` (SHA-256 por omisión), `extra_files: [{glob: ./scripts/install.sh}]` | FR-093 |
| `sboms` | una entrada, `artifacts: archive`: el valor por omisión escrito explícito (syft, un SBOM por archivo) | FR-093 |
| `signs` | `cmd: cosign`, `artifacts: checksum`, `signature: "${artifact}.sigstore.json"`, `args: [sign-blob, "--bundle=${signature}", "${artifact}", --yes]` | FR-093 |
| `homebrew_casks[0]` | `name: kitlegal`, `repository: {owner: jmorenobl, name: homebrew-tap, token: "{{ .Env.PUBLISHER_TOKEN }}"}`, `homepage`, `description`, `hooks.post.install` que retira `com.apple.quarantine` del binario en macOS | FR-093, FR-097 |
| `scoops[0]` | `name: kitlegal`, `repository: {owner: jmorenobl, name: scoop-bucket, token: "{{ .Env.PUBLISHER_TOKEN }}"}`, `homepage`, `description`, `license: Apache-2.0` | FR-093, FR-097 |
| `nfpms[0]` | `package_name: kitlegal`, `formats: [deb, rpm]`, `maintainer: "Jorge <jmorenobl@gmail.com>"`, `description`, `homepage`, `license: Apache-2.0` | FR-093 |
| `release` | `github: {owner: jmorenobl, name: kitlegal}`, `extra_files: [{glob: ./scripts/install.sh}]` | FR-093, FR-109 |
| `changelog` | `use: git`, `sort: asc`, `groups` por expresión regular de Conventional Commits (`feat`, `fix`, resto) | FR-093 |

`PUBLISHER_TOKEN` aparece **solo** en los dos `token` (la plantilla se evalúa solo al publicar: V11, V12). `homepage`:
`https://github.com/jmorenobl/kitlegal`.

## 3. Objetivos del `Makefile`

| Objetivo | Receta | En `ci` | FR |
|---|---|---|---|
| `goreleaser-check` | `$(GORELEASER) check` | sí | FR-094 |
| `release` | `$(GORELEASER) release --snapshot --clean --skip=publish,sign,sbom` | no | FR-095 |
| `snapshot-check` | `go test -count=1 -tags=snapshot -run '^TestSnapshot$$' .` y `KITLEGAL_DIST=$(CURDIR)/dist go test -count=1 -run '^TestEntregaDelHito$$/instalador-' ./internal/app/` | no | FR-120, FR-108 |
| `install` | `CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/kitlegal` y `"$$(go list -f '{{.Target}}' ./cmd/kitlegal)" skills install -g --host claude` | no | FR-125 |

`snapshot-check` necesita guiones `instalador-` en `internal/app/testdata/script/` (el arnés falla sin ninguno,
[arnes-e2e.md](./arnes-e2e.md) §5), que llegan con la activación de la suite tras el bucle de tareas; dentro del bucle, la
tarea de la release y la de CI lo ejecutan completo con copias momentáneas de los guiones congelados, y la de la release
comprueba además que sin ellas falla (research D28, «Cómo se verifica dentro del run»). `scripts/install.sh` existe
antes que `.goreleaser.yaml`: `checksum.extra_files` lo nombra con una ruta literal, y goreleaser falla en el snapshot si
no existe (research V48).

`ci` pasa a ser `fmt-check lint test test-integration vuln schema-check skills-check goreleaser-check secrets mod-verify
mod-tidy-check`. `make help` describe `install`, `release`, `snapshot-check`, `goreleaser-check` y `skills-sync` con su
comportamiento nuevo (FR-121).

`dist/` está en `.gitignore` (V37).

## 4. `TestSnapshot` (etiqueta `snapshot`, raíz del módulo)

Sobre `dist/` tras `make release`, falla nombrando lo que falta o sobra:

1. exactamente seis archivos de distribución: `kitlegal_{darwin,linux}_{amd64,arm64}.tar.gz` y
   `kitlegal_windows_{amd64,arm64}.zip` (los tipos `Archive` de `dist/artifacts.json`, V15);
2. `dist/checksums.txt` tiene, para cada uno, una línea `<sha256>  <nombre>` cuya huella es la calculada sobre el
   fichero (sin fijar el total de líneas: puede listar los `.deb`/`.rpm` y `install.sh`);
3. ningún `*.sbom.json`, `*.sig`, `*.sigstore.json` ni `*.pem` en `dist/` (el snapshot omite SBOM y firma);
4. el binario de la plataforma que ejecuta el test imprime en `version` la `version` y el `commit` de
   `dist/metadata.json`; sin archivo para esa plataforma, el test falla. En el trabajo de CI (`ubuntu-latest`), ese
   binario es el linux/amd64 (S15), y el test registra con `t.Log` qué plataforma ejecutó.

## 5. `.github/workflows/ci.yml`: trabajo `snapshot`

En `pull_request` y en `push` a `main`, junto al trabajo `ci` y sin tocarlo: `runs-on: ubuntu-latest`,
`permissions: contents: read`, **sin** `id-token` y **sin** ningún `secrets.*`; `actions/checkout` con `fetch-depth: 0`;
`actions/setup-go` con `go-version-file: go.mod` y la misma caché que `ci` (`go.sum` y `tools/*/go.sum`); `make
release`; `make snapshot-check` (FR-120).

## 6. `.github/workflows/release.yml`

`on: push: tags: ['v*']` y ningún otro evento (FR-110). Permisos por trabajo, ninguno a nivel de flujo.

**`publicar`** (`ubuntu-latest`; `contents: write`, `id-token: write`, `attestations: write`, FR-111):
`actions/checkout` con `fetch-depth: 0` y `persist-credentials: false`; `actions/setup-go` con `go-version-file:
go.mod` y `cache: false`; instalar syft
(`anchore/sbom-action/download-syft`) y cosign (`sigstore/cosign-installer`); `go tool -modfile=tools/goreleaser/go.mod
goreleaser release --clean` con `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}` y `PUBLISHER_TOKEN: ${{
secrets.PUBLISHER_TOKEN }}` **solo en el entorno de ese paso** (FR-114); `actions/attest-build-provenance` con
`subject-path` de los seis archivos y `dist/checksums.txt` (FR-112).

**`humo`** (`needs: publicar`; `ubuntu-latest`; `contents: read`, `attestations: read`; `GH_TOKEN: ${{ github.token
}}`), en un directorio temporal y sin el código del repositorio (FR-113):

1. `gh release download "$GITHUB_REF_NAME" --repo jmorenobl/kitlegal --pattern kitlegal_linux_amd64.tar.gz --pattern
   checksums.txt`;
2. la línea de `kitlegal_linux_amd64.tar.gz` de `checksums.txt` verifica con `sha256sum -c`;
3. `gh attestation verify kitlegal_linux_amd64.tar.gz --repo jmorenobl/kitlegal` sale con 0;
4. extraído, `./kitlegal version` imprime en su primera línea `kitlegal $GITHUB_REF_NAME`;
5. con `KITLEGAL_CACHE_DIR` en un directorio vacío, `./kitlegal boe articulo BOE-A-2015-10565 a21 --offline` sale con 4;
6. en un directorio vacío, `kitlegal skills install` deja `.agents/skills/boe-legislacion/SKILL.md`.

Las acciones, por su etiqueta mayor como el resto de flujos (V36; versiones: S2). Ejecutarlo es humano: solo lo
dispara una etiqueta `v*` (FR-115); ninguna tarea lo ejecuta.

## 7. `scripts/install.sh`

```text
curl -fsSL https://raw.githubusercontent.com/jmorenobl/kitlegal/main/scripts/install.sh | sh
curl -fsSL …/install.sh | sh -s -- <versión>
sh install.sh [<versión>]
```

| Entrada | Regla |
|---|---|
| argumentos | 0 o 1; `<versión>` con forma SemVer 2.0.0, con o sin `v` (`0.1.0` y `v0.1.0` piden la etiqueta `v0.1.0`) |
| `uname -s` | `Darwin` → `darwin`, `Linux` → `linux`; otro valor: error que lo nombra |
| `uname -m` | `x86_64`/`amd64` → `amd64`, `arm64`/`aarch64` → `arm64`; otro valor: error que lo nombra |
| `KITLEGAL_INSTALL_DIR` | directorio de instalación si está definido y no vacío |
| `HOME` | si no, `$HOME/.local/bin`; sin ninguno de los dos, o con un `HOME` que no es una ruta absoluta o que solo tiene barras, error **antes de descargar** |
| `KITLEGAL_INSTALL_URL` | base de las URL, por omisión `https://github.com/jmorenobl/kitlegal/releases`; los tests la apuntan a `file://…` (FR-107) |

Descargas (`curl -fsSL`): `<base>/latest/download/<archivo>` sin versión, `<base>/download/v<versión>/<archivo>` con
ella, y `checksums.txt` del mismo sitio; `<archivo>` = `kitlegal_<os>_<arch>.tar.gz`.

Verificación: la línea de `checksums.txt` cuyo **segundo campo es exactamente** `<archivo>` (nunca por posición ni
subcadena); huella con `sha256sum` o, si no existe, `shasum -a 256`; sin línea, con más de una o con huella distinta,
error.

Instalación: solo el miembro `kitlegal` del archivo, copiado a un temporal del directorio de instalación (creado si
falta) y renombrado encima de `kitlegal`; un `kitlegal` anterior sigue intacto ante cualquier error. Ningún fichero de
arranque del shell ni fuera del directorio de instalación cambia; los temporales (`mktemp -d`) se retiran siempre.

Salida correcta (salida estándar): si el directorio no está en el `PATH`, una línea de explicación y la línea exacta
`export PATH='<directorio>':"$PATH"` (comilla simple escrita `'\''`), que evaluada en `sh` deja el directorio en el
`PATH`; la **última línea** es `kitlegal skills install`. Error (salida de error): una línea que empieza por
`install.sh: ` y nombra la versión («la última versión» sin argumento) y lo que falló; código distinto de 0; nada
instalado. Lo mismo en un locale UTF-8 con el `/bin/sh` de macOS (bash 3.2): ninguna variable va sin llaves delante
de un carácter que no es ASCII.

Guion cortado (`curl … | sh` que recibe solo una parte): todo lo que hace va en funciones y la llamada a `main`, en
la última línea, dentro de un grupo `{ main "$@"; }`, así que ningún corte ejecuta nada a medias; una guarda en la
segunda línea, que `main` retira al empezar, hace terminar con código 1 y una línea `install.sh: ` el corte que deja
las funciones enteras sin la llamada. Solo un corte antes de que la guarda llegue entera sale con 0, como un guion
vacío (revisión final).
