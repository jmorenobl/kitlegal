# Quickstart: validar H19 · Instalar sin clonar

Guía de validación de la entrega, escenario a escenario. Cada escenario remite a su contrato y a sus FR/SC; no repite
el detalle.

## Antes de empezar

- Desde la raíz del repositorio, en **una sola sesión** de `sh`/`bash` (los escenarios comparten `REPO` y `T`), con
  Go ≥ 1.21, `git`, `make` y `curl`.
- Red: solo la de las herramientas de Go —el toolchain que fija `go.mod`, el proxy de módulos la primera vez que se
  construye `tools/goreleaser` (research S10) y la base de datos de `make vuln`—. Ningún escenario pide nada a una
  fuente legal ni a GitHub.
- Efectos en el árbol: `bin/` (el arranque y los escenarios 7 y 11), `dist/` (escenarios 9 y 10) y los perfiles de cobertura de `make ci`,
  **todos ignorados por git** (`.gitignore`). Ningún escenario toca el índice, el historial ni un fichero versionado;
  todo lo demás vive en `$T`, que el último escenario borra. `git status --porcelain` al final es el mismo que al
  principio.

```sh
REPO=$(git rev-parse --show-toplevel)
cd "$REPO"
INICIAL=$(git status --porcelain)
T=$(mktemp -d)
make build
```

## 1. `make ci` en verde, con `goreleaser check` (FR-094, FR-140, SC-015)

```sh
make ci
```

Esperado: termina con `ci: todos los controles en verde`; en su salida aparece la ejecución de `goreleaser check`
(objetivo `goreleaser-check`) con código 0.

## 2. Instalación local sin `.claude/` (US1, SC-001, SC-021; contracts/applet-skills.md §4.1)

```sh
mkdir "$T/p1" && cd "$T/p1"
HOME="$T/vacio" "$REPO/bin/kitlegal" skills install --json
ls -A                                   # .agents
ls -A .agents/skills                    # boe-legislacion  kitlegal.json  legal-core
test ! -e "$T/vacio" && echo "nada en HOME"
cmp .agents/skills/boe-legislacion/SKILL.md "$REPO/skills/boe-legislacion/SKILL.md" && echo "SKILL.md empotrado"
cd "$REPO"
```

Esperado: exit 0; el sobre con `"fuente":"kitlegal.skills"` y `data` con las dos skills en `"estado":"instalada"` y
`"enlaces":[]`; solo `.agents/`; nada en `HOME`; los ficheros son los del árbol.

## 3. Con `.claude/`: enlaces relativos, y segunda ejecución sin cambios (SC-002, SC-007)

```sh
mkdir -p "$T/p2/.claude" && cd "$T/p2"
"$REPO/bin/kitlegal" skills install
readlink .claude/skills/boe-legislacion  # ../../.agents/skills/boe-legislacion
"$REPO/bin/kitlegal" skills install --json | grep -o '"estado":"[a-z ]*"'
cd "$REPO"
```

Esperado: `readlink` imprime el destino literal relativo; la segunda ejecución da `"estado":"sin cambios"` dos veces.

## 4. Global con `HOME` temporal y `--dir` (US6, SC-004, SC-005)

```sh
mkdir -p "$T/h1/.claude" "$T/p3" && cd "$T/p3"
HOME="$T/h1" "$REPO/bin/kitlegal" skills install -g
ls -A "$T/h1/.agents/skills"; readlink "$T/h1/.claude/skills/legal-core"; ls -A
"$REPO/bin/kitlegal" skills install --dir "$T/otro" && ls -A "$T/otro"
HOME="$T/h1" "$REPO/bin/kitlegal" skills install -g --dir "$T/otro"; echo "código $?"
cd "$REPO"
```

Esperado: skills y manifiesto en `$T/h1/.agents/skills`, enlace `../../.agents/skills/legal-core`, `p3` vacío; con
`--dir`, las skills y `kitlegal.json` en `$T/otro`; `-g` con `--dir` → `código 2` y el mensaje `-g y --dir se excluyen`.

## 5. `doctor` encuentra y su orden arregla (US4, SC-011; contracts/applet-skills.md §5-§6)

```sh
cd "$T/p2"
echo "editado" >> .agents/skills/legal-core/SKILL.md
PATH="$REPO/bin:$PATH" kitlegal skills doctor; echo "código $?"
PATH="$REPO/bin:$PATH" sh -c "$(PATH="$REPO/bin:$PATH" kitlegal skills doctor 2>&1 | sed -n 's/^fichero editado: [^:]*: //p')"
PATH="$REPO/bin:$PATH" kitlegal skills doctor; echo "código $?"
cd "$REPO"
```

Esperado: el primer `doctor` sale con 1 y la línea `fichero editado: .agents/skills/legal-core/SKILL.md: rm --
'.agents/skills/legal-core/SKILL.md' && kitlegal skills install legal-core --host claude`; tras ejecutar esa orden, el
segundo sale con 0.

## 6. Las skills sin instalación por enlaces (FR-081, FR-082, FR-083, SC-014)

Dos escenarios, porque cada uno vale desde una tarea distinta: el 6a desde que los `SKILL.md` pasan a `kitlegal
<applet>` y se retiran sus `scripts/` (plan, paso 11); el 6b desde que el job de evals deja de nombrar `bin/instalado`
(plan, paso 12), porque hasta entonces `.github/workflows/evals.yml` lo nombra. Con la entrega completa, los dos.

### 6a. Los `SKILL.md` solo cambian en la forma de invocar, y ninguna skill tiene `scripts/` (FR-081, FR-082)

```sh
for s in boe-legislacion legal-core; do
  git show "main:skills/$s/SKILL.md" \
    | sed -e 's#scripts/boe#kitlegal boe#g' -e 's#scripts/territorio#kitlegal territorio#g' > "$T/$s.md"
  diff "$T/$s.md" "skills/$s/SKILL.md"
done
ls skills/*/scripts 2>/dev/null; echo "sin scripts/: $?"
```

Esperado: `diff` solo muestra las frases que dicen de dónde sale el binario (que `kitlegal` se invoca desde el `PATH`);
`ls` no encuentra nada (código distinto de 0).

### 6b. Ningún resto de la instalación por enlaces en `skills`, `Makefile` y `.github` (FR-083, SC-014)

```sh
grep -rn -e 'scripts/boe' -e 'scripts/territorio' -e 'bin/instalado' skills Makefile .github; echo "sin restos: $?"
```

Esperado: `grep` no encuentra nada (código 1). Es lo mismo que fija `TestSinInstalacionPorEnlaces`.

## 7. El aviso de versión, sin red (US3, SC-013; contracts/aviso.md)

```sh
make build VERSION=v0.1.0 && cp bin/kitlegal "$T/k1"
make build VERSION=v0.2.0 && cp bin/kitlegal "$T/k2"
mkdir "$T/p4" && cd "$T/p4"
"$T/k1" skills install > /dev/null
"$T/k2" territorio resolver Leganés --json > "$T/con.json" 2> "$T/con.err"; echo "código $?"
"$T/k1" territorio resolver Leganés --json > "$T/sin.json" 2> "$T/sin.err"
cat "$T/con.err"; wc -l < "$T/sin.err"
"$T/k2" skills install --json | grep -o '"estado":"[a-z ]*"'
"$T/k2" territorio resolver Leganés > /dev/null
cd "$REPO"
```

Esperado: `código 0` y, en `con.err`, exactamente la línea `aviso: las skills instaladas son de kitlegal v0.1.0 y este
binario es kitlegal v0.2.0; ejecuta: kitlegal skills install`; `sin.err` vacío (`0`); la reinstalación con `k2` da
`actualizada` dos veces, y después ya no hay aviso. (`bin/kitlegal` queda con la versión `v0.2.0`; el escenario 11 lo
reconstruye.)

## 8. `make install` con `HOME` y `GOBIN` temporales (US2, FR-125, FR-126, SC-018)

```sh
export GOMODCACHE="$(go env GOMODCACHE)" GOCACHE="$(go env GOCACHE)"
HOME="$T/h2" GOBIN="$T/gobin" make install
ls "$T/gobin/kitlegal" "$T/h2/.agents/skills"
readlink "$T/h2/.claude/skills/boe-legislacion" "$T/h2/.claude/skills/legal-core"
test ! -e bin/instalado && echo "sin bin/instalado"
```

Esperado: el binario en `$T/gobin`; `boe-legislacion`, `kitlegal.json` y `legal-core` en `$T/h2/.agents/skills`; los
dos enlaces con destino `../../.agents/skills/<skill>`; ningún `bin/instalado`. Las cachés se fijan antes de cambiar
`HOME` para no descargar nada; `GOBIN` en `$T` evita tocar el `kitlegal` instalado de la cuenta. Lo mismo, automático:
`go test -count=1 -tags=integration -run '^TestInstalacion$' ./internal/skills/`.

## 9. Snapshot de la release (US5, FR-095, FR-120, SC-016)

```sh
make release
ls dist/*.tar.gz dist/*.zip
make snapshot-check
```

Esperado: `make release` no publica ni firma ni genera SBOM; `ls` lista exactamente los seis archivos
`kitlegal_{darwin,linux}_{amd64,arm64}.tar.gz` y `kitlegal_windows_{amd64,arm64}.zip`; `make snapshot-check` pasa
`TestSnapshot` (checksums, sin SBOM ni firmas, `version` con la versión y el commit de `dist/metadata.json`) y los
guiones `h19-instalador-*` contra `dist/`. Este escenario valida la entrega, con la suite ya activada; antes de activarla
no hay guiones `instalador-` en `internal/app/testdata/script/` y `make snapshot-check` falla con `KITLEGAL_DIST: ningún
guion instalador- que ejecutar` (contracts/arnes-e2e.md §5), que es la garantía de que no pasa en vacío.

## 10. `install.sh` contra el snapshot, sin red (FR-100 a FR-107, SC-017; contracts/release.md §7)

```sh
so=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in x86_64|amd64) ar=amd64 ;; *) ar=arm64 ;; esac
mkdir -p "$T/origen/latest/download"
cp "dist/kitlegal_${so}_${ar}.tar.gz" dist/checksums.txt "$T/origen/latest/download/"
HOME="$T/h3" KITLEGAL_INSTALL_URL="file://$T/origen" https_proxy=http://127.0.0.1:9 HTTPS_PROXY=http://127.0.0.1:9 \
  sh scripts/install.sh
"$T/h3/.local/bin/kitlegal" version
```

Esperado: `install.sh` verifica la huella, instala en `$T/h3/.local/bin/kitlegal`, imprime la línea `export
PATH='…/.local/bin':"$PATH"` (el directorio no está en el `PATH`) y termina con `kitlegal skills install`; el binario
instalado imprime la versión del snapshot. Con el proxy cerrado, cualquier petición HTTP(S) habría fallado.

## 11. Aceptación e2e congelada y limpieza

```sh
go test -count=1 -run '^TestEntregaDelHito$/h19-' ./internal/app/
make build
rm -rf "$T" dist
test "$(git status --porcelain)" = "$INICIAL" && echo "árbol intacto"
```

Esperado: los 18 guiones `h19-*` en verde (una vez activados por el workflow); `bin/kitlegal` reconstruido con la
versión de `git describe`; `árbol intacto`.

## 12. Tras fusionar (humano, fuera del run; FR-150, SC-023)

Etiqueta `v0.1.0` empujada por una persona; `release.yml` verde con su trabajo de humo; en un Mac limpio `curl -fsSL
https://raw.githubusercontent.com/jmorenobl/kitlegal/main/scripts/install.sh | sh` (y, en otro,
`brew install jmorenobl/tap/kitlegal`) y `kitlegal skills install` en un proyecto vacío; Claude Code responde «¿qué dice
el art. 21 de la Ley 39/2015?» con `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`. Nada de esto lo ejecuta
el run.
