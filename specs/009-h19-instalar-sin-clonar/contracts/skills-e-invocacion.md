# Contrato: las skills invocan `kitlegal` desde el `PATH`

Requisitos: FR-080 a FR-086, FR-125 a FR-128, FR-130 a FR-133. Decisiones en [../research.md](../research.md) D3,
D17, D18 y D19.

## 1. Lo empotrado

`skills.go` en la raíz del módulo (paquete `kitlegal`):

```go
//go:embed skills/*/SKILL.md skills/*/references/*
var skills embed.FS

// Skills devuelve lo empotrado tal cual, con rutas skills/<skill>/….
func Skills() fs.FS
```

`internal/app` lo lee a `[]instalacion.SkillEmpotrada`: una skill por directorio `skills/<n>/` **con** `SKILL.md`; sus
ficheros son `SKILL.md` y todo fichero bajo `references/`, con ruta relativa al directorio de la skill y huella
SHA-256. `TestSkillsEmpotradas` (raíz) exige que el conjunto y los bytes empotrados sean exactamente los del árbol
(FR-001, FR-003, FR-004).

## 2. `SKILL.md`

- La **región generada** (entre las dos marcas) titula cada applet `### \`kitlegal <applet>\`` y escribe cada orden
  `\`kitlegal <applet> <verbo> <argumentos>\``; la regenera `make skills-sync` desde `--describe` (FR-080, FR-133).
- En el **texto libre**, toda invocación `scripts/boe` pasa a `kitlegal boe` y `scripts/territorio` a `kitlegal
  territorio` (FR-081). Las frases que dicen de dónde sale el binario pasan a decir que `kitlegal` se invoca desde el
  `PATH`, y la de `legal-core` que hoy dice «si `scripts/territorio` no resuelve a un binario» pasa a «si `kitlegal` no
  está en el `PATH`». Nada más cambia: protocolo, reglas, forma de la cita y de los avisos intactos. Se comprueba así:
  sustituyendo de vuelta la forma de invocar, `git diff main -- skills/*/SKILL.md` solo deja esas frases
  (quickstart §6).
- `metadata.kitlegal-applets` y `metadata.kitlegal-referencias` no cambian.

## 3. `skills-sync` y `skills-check`

- Sin enlaces: desaparecen `internal/skills/enlaces.go`, `EnlacesEsperados`, las derivas `enlace-ausente`,
  `enlace-sobrante` y `enlace-con-otro-destino`, y los arreglos que creaban o retiraban enlaces en `scripts/` (FR-080,
  FR-083).
- Defecto nuevo, en la sincronización y en la comprobación: `<skill>: una skill no lleva scripts/ (ADR 0019)` si existe
  cualquier entrada `skills/<skill>/scripts` (vista con `Lstat`). Hace fallar `make skills-sync` y `make skills-check`
  (FR-082). `skills-sync` no la retira.
- `make skills-check` gana `TestOrdenesDeLasSkillsEmpotradas` (en `internal/app`): cada orden de la región generada de
  cada `SKILL.md` **empotrado** empieza por `kitlegal <applet> <verbo>` con un applet y un verbo del registro de
  producción (FR-084).

## 4. Lo que se retira

`skills/boe-legislacion/scripts/`, `skills/legal-core/scripts/`, `scripts/instalar-skills.sh`,
`internal/skills/enlaces.go` y `enlaces_test.go`, los casos de enlaces de `internal/app/skills_test.go`, y todo paso del
`Makefile` o de CI que cree o lea `bin/instalado/` (FR-083). `TestSinInstalacionPorEnlaces` (raíz) lo vigila (SC-014);
nace con el cambio del job de evals (§6), porque hasta entonces `.github/workflows/evals.yml:147` nombra
`bin/instalado`.

## 5. `make install` y su prueba

Receta en [release.md](./release.md) §3. `TestInstalacion` (`internal/skills`, etiqueta `integration`) sigue
instalando sobre la copia mínima del árbol con `HOME`, `GOBIN` y `GOPATH` temporales y `GOPROXY=off`, y sus guiones de
`internal/skills/testdata/script/` pasan a comprobar:

| Guion | Qué comprueba |
|---|---|
| `instalar.txtar` | tras `make install`: el binario en el `GOBIN` temporal; `~/.agents/skills/boe-legislacion/`, `~/.agents/skills/legal-core/` y `~/.agents/skills/kitlegal.json`; `~/.claude/skills/<skill>` enlaces con destino literal `../../.agents/skills/<skill>` (`readlink`); y que el `SKILL.md` instalado es el del árbol (FR-125, FR-126, SC-018) |
| `instalar-de-nuevo.txtar` | una segunda `make install` dice `sin cambios` para las dos skills, deja `~/.agents/skills/kitlegal.json` byte a byte igual (`cp` antes y `cmp` después) y los dos enlaces con su destino literal (`readlink`); el detalle byte a byte de todo el árbol lo cubre el guion de aceptación `skills-idempotencia` |
| `instalar-con-conflicto.txtar` | con un `~/.claude/skills/boe-legislacion` de la instalación anterior (enlace absoluto a `repo/skills/boe-legislacion`), `make install` falla nombrándolo como «enlace a otro sitio» y no toca `~/.agents` ni `~/.claude` |
| `instalar-sin-gobin.txtar` | sin `GOBIN`, el binario queda en `$GOPATH/bin` y las skills se instalan igual |

El guion del enlace roto escrito en `instalacion_test.go` pasa a esperar «enlace roto». `rutasDeLaInstalacion` deja
de copiar `scripts/instalar-skills.sh` y copia `skills.go` por ser un `GoFile` del cierre de `./cmd/kitlegal` (lo
recoge `ficherosDelBinario`, con sus `EmbedFiles`).

## 6. Job de evals

- `.github/workflows/evals.yml`, paso «Instalar kitlegal y las skills como las deja make install»: `make install` y
  `dirname "$(go list -f '{{.Target}}' ./cmd/kitlegal)" >> "$GITHUB_PATH"`.
- Paso «Retirar Python del runner»: `kitlegal=$(readlink -e "$(command -v kitlegal)")`, que falla si no está; y
  `$HOME/.agents` entre lo que no se retira.
- `scripts/evals.sh`, comprobación 6: además de `~/.claude/skills/<skill>/SKILL.md`, que `kitlegal` está en el `PATH`.
- `internal/evals/preparar.go`: la sesión de prueba de red pide `kitlegal boe articulo BOE-A-2015-10565 a9998 --json` y
  `… --offline --json` (FR-127).
- Ni `evals/boe-legislacion/` ni `evals/legal-core/` cambian (FR-086): sus `comandos` se declaran por applet.

## 7. Documentación

- `README.md`: instalación para quien usa (las dos órdenes: `curl … | sh` o `brew install jmorenobl/tap/kitlegal`; en
  Windows `scoop bucket add` del bucket y `scoop install kitlegal`; después `kitlegal skills install`, con `-g` y
  `--host claude`), actualización (gestor de paquetes o repetir `install.sh`, y repetir `kitlegal skills install`), el
  aviso, y `make install` como bucle de desarrollo (FR-130).
- `CONTRIBUTING.md`: alineado con el `Makefile` (`install`, `release`, `snapshot-check`, `goreleaser-check`, `ci`,
  `skills-sync`) y el paso único para retirar lo que dejó el `make install` anterior: los enlaces absolutos de
  `~/.claude/skills/boe-legislacion` y `~/.claude/skills/legal-core` y `bin/instalado/` (FR-131).
- `CHANGELOG.md`, *Unreleased*: el applet `skills`, la release e `install.sh`, el cambio de invocación de las skills y lo
  retirado; la sección no se cierra (FR-132).
