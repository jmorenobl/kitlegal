# Implementation Plan: H19 · Instalar sin clonar: release `v0.1.0` y el binario por gestor de paquetes, con las skills dentro

**Branch**: `009-h19-instalar-sin-clonar` | **Date**: 2026-09-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/009-h19-instalar-sin-clonar/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están en [research.md](./research.md) (D1-D34) con su alternativa rechazada y su
motivo. Toda afirmación sobre una herramienta o dependencia externa remite a la tabla de verificación de
[research.md](./research.md) (V1-V48), con el `fichero:línea` del módulo en la caché local, la salida de `go doc` o la
orden ejecutada; lo que no se puede comprobar sin red, sin la plataforma o sin una etiqueta está declarado como
supuesto (S1-S15) y **no se afirma como hecho** en ningún punto de este plan.

## Summary

H19 no entrega skill nueva: **protege las dos que existen haciéndolas instalables sin clonar** (principio VIII). Cinco
piezas: (1) el applet `skills` (`install`, `list`, `doctor`) con las skills empotradas en el binario, un manifiesto
por ámbito y un aviso sin red en los demás applets; (2) la release —`.goreleaser.yaml`, `release.yml` solo en `v*` y
`scripts/install.sh`—; (3) `goreleaser check` en `make ci` y `make release` como snapshot; (4) los dos `SKILL.md`
invocan `kitlegal <applet> …` desde el `PATH`, sin `scripts/`; y (5) `make install` como bucle de desarrollo y del job
de evals.

Decisiones que sostienen el diseño:

1. **Dominio puro con el disco detrás de un puerto** (D1, D2): `internal/core/instalacion` decide qué es de quién
   (FR-041), qué se escribe y en qué orden, el manifiesto, los hallazgos de `doctor` con su orden, la igualdad de
   versiones y el aviso, sin hacer E/S; `internal/disco` implementa los puertos sobre el sistema de ficheros real. Es lo
   que permite probar de forma exhaustiva las tablas de FR-041/FR-065 y las propiedades de FR-044 y FR-066 sobre un
   disco en memoria.
2. **Nada se lee ni se escribe a través de un enlace** (D6): `Lstat`, `Readlink` y, para leer, `Open` + `Stat` +
   `SameFile` sobre ficheros regulares; `os.Root` se descarta porque sigue enlaces.
3. **Un orden de aplicación que hace cierto FR-044** (D7): retirar → enlazar → manifiesto final → escribir. Tras un
   fallo en cualquier operación, una nueva ejecución completa sin conflictos; lo prueba un test que falla en cada
   operación del plan.
4. **La versión del binario llega al registro al construirlo** (D4) y **el aviso lo emite la composición tras analizar
   el verbo** (D5), con exactamente las invocaciones de FR-070.
5. **El manifiesto es JSON canónico y estricto** (D10): determinista byte a byte entre máquinas, ilegible ante
   cualquier desviación de su forma.
6. **goreleaser v2.18.1 como módulo de herramienta y configuración sin nada obsoleto** (D20, D21): `homebrew_casks`
   (no `brews`, que haría salir `check` con 2), `nfpms.maintainer`, `release.github` fijado; la huella de `install.sh`
   en `checksums.txt` (D23); el cask retira la cuarentena de macOS (D25).
7. **`install.sh` por las URL de descarga de GitHub y con un origen sustituible** (D22, D27): archivos sin versión en el
   nombre, `KITLEGAL_INSTALL_URL` para los tests, y sus pruebas son guiones de aceptación que corren en `make ci`
   contra un origen local y en CI contra el snapshot (D28).
8. **El arnés e2e gana lo que la aceptación necesita** (D24): binarios con versión y con un enlazador que falla, la
   orden `arbol` para comparar el disco byte a byte sin seguir enlaces, y un origen de release local; todo fijado de
   antemano en [contracts/arnes-e2e.md](./contracts/arnes-e2e.md), contra el que T001 escribe la suite congelada; los
   formatos de salida que la suite afirma salen de los contratos del producto que enumera su §6.

Artefactos de diseño: [data-model.md](./data-model.md) y [contracts/](./contracts/)
([applet-skills](./contracts/applet-skills.md), [manifiesto](./contracts/manifiesto.md),
[aviso](./contracts/aviso.md), [arnes-e2e](./contracts/arnes-e2e.md), [release](./contracts/release.md),
[skills-e-invocacion](./contracts/skills-e-invocacion.md)); validación en [quickstart.md](./quickstart.md).

## Technical Context

**Language/Version**: Go 1.27 sin cambios (`go 1.27.0` + `toolchain go1.27.1`, que el `Makefile` lee y exporta).
POSIX `sh` para `scripts/install.sh`; YAML para `.goreleaser.yaml` y los flujos de GitHub Actions; Markdown para las
skills.

**Primary Dependencies**: las de `go.mod`, **sin ninguna nueva en el módulo**. Biblioteca estándar que usa el código
nuevo: `embed` e `io/fs` (paquete raíz y `internal/app`), `encoding/json/v2` y `encoding/json/jsontext` (dominio; ya
los usa `internal/skills`, V21), `crypto/sha256`, `regexp`, `path`. `go.yaml.in/yaml/v3` (§V) solo en `_test.go` de la raíz, fuera del
binario. **Herramientas**: goreleaser v2.18.1 como módulo nuevo de `tools/` (D20); syft y cosign solo en `release.yml`
(FR-096); acciones nuevas en `release.yml` (D30). Todo en *Complexity Tracking*.

**Storage**: ficheros del proyecto o de `HOME`: `.agents/skills/<skill>/`, `.agents/skills/kitlegal.json` y
`.claude/skills/<skill>` (enlace o copia). Sin caché ni base de datos.

**Testing**: `make test` (unitarios del dominio con disco en memoria, adaptador sobre `t.TempDir()`, applet, kernel,
configuración de la release, y los e2e con `-race`), `make test-integration` (`TestInstalacion` reescrito),
`make schema-check` (gana `schemas/instalacion.json`), `make skills-check` (gana el defecto de `scripts/` y
`TestOrdenesDeLasSkillsEmpotradas`), `make goreleaser-check` (nuevo, en `ci`), `make snapshot-check` (nuevo, fuera de
`ci`, en el trabajo de snapshot). **Sin red**: ni el applet ni el aviso la usan (D32), `install.sh` se prueba contra
`file://` con los proxies cerrados (D24).

**Target Platform**: darwin, linux y windows en amd64 y arm64, `CGO_ENABLED=0`, `-trimpath`. Desarrollo en
darwin/arm64; `make ci` y el trabajo de snapshot en `ubuntu-latest` (linux/amd64 según el supuesto S15); job de evals en
`ubuntu-24.04`. Sin CI
en Windows (el recurso de copia se prueba con el enlazador que falla).

**Project Type**: CLI multicall + skills del estándar Agent Skills empotradas + release multiplataforma.

**Performance Goals**: el aviso cuesta como mucho seis `Lstat` y la lectura de un fichero pequeño por invocación; el
guion `boe-cache-rapida.txtar` (200 ms) sigue en verde con él.

**Constraints**: nunca leer, escribir ni hashear a través de un enlace por debajo de la raíz del ámbito (FR-028);
atómico e idempotente (FR-040 a FR-045); la predicción de enlace o copia se sondea en el sistema de ficheros del propio
ámbito, nunca en `TMPDIR` ni por encima de la raíz, y la sonda deja el directorio byte a byte igual (D9); ningún cambio en el kernel de `internal/cli` (spec, *Fuera de alcance*);
sobre de fallo de ADR 0006 intacto (FR-052); ninguna tarea publica, etiqueta ni necesita `PUBLISHER_TOKEN`
(FR-097, FR-115).

**Scale/Scope**: 1 applet con 3 verbos · 2 paquetes nuevos (`internal/core/instalacion`, `internal/disco`) y el paquete
raíz · 1 esquema publicado nuevo · 1 módulo de herramienta · 1 flujo nuevo y 3 tocados · 1 guion POSIX nuevo y 1
retirado · 2 `SKILL.md` · 18 guiones de aceptación y 1 guion e2e existente tocado · 4 guiones de integración
reescritos · documentación.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H19 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | H19 **no toca ninguna fuente** ni añade ninguna petición HTTP: el applet y el aviso no alcanzan `net` (subprueba de arquitectura nueva, D32) e `install.sh` solo descarga la release pública del propio proyecto por HTTPS. **Frontera humana intacta**: etiquetar, publicar, hacer público el repositorio y crear el tap y el bucket son humanos; `release.yml` solo corre en una etiqueta `v*` que empuja una persona (FR-110, FR-115) y ninguna tarea la dispara ni usa `PUBLISHER_TOKEN` (FR-097). Nada con identidad | ✅ Cumple |
| **II** | Nada sin cita ni fuente | Los tres verbos emiten el sobre de seis claves con la procedencia de applet calculado (`kitlegal.skills`, `kitlegal:applet/skills`; ADR 0006, D15). No hay contenido legal ni grafo. Las skills instaladas son **byte a byte** las empotradas (FR-003) y conservan su protocolo de cita intacto (FR-081) | ✅ Cumple |
| **III** | Tests primero y offline | T001 escribe los 18 guiones de aceptación desde el spec y **quedan congelados** antes de cualquier código (ADR 0018); cada tarea de código trae su test, y las que no lo traen (la suite, `install.sh`, la documentación y el cierre de la DoD) se declaran una a una en tasks.md con su verificación. Todo test es offline: el dominio con disco en memoria, el adaptador en `t.TempDir()`, `install.sh` contra `file://` con los proxies cerrados; ningún fixture de fuente nuevo. Toda salida correcta del applet se valida contra `schemas/instalacion.json` en test (FR-053). Umbrales intactos (`internal/core/**` ≥ 85 %, global ≥ 70 %), y el dominio nuevo cae en el de `internal/core/**` | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | Dominio puro en `internal/core/instalacion`, que **define sus puertos** (`Disco`, `Escritor`, `Enlazador`) y no importa `os`, `io` ni `io/fs` (R1, V26); adaptador en `internal/disco`, que R1 pasa a denegar al dominio junto con el paquete raíz, también en sus tests (D32); composición en `internal/app`. Errores tipados con `schema.ConClase` que el kernel traduce a 2 y 1 (data-model §9); ningún `panic` en rutas de usuario (`FuzzLeerManifiesto`). Reglas de dependencia, abajo | ✅ Cumple con justificación (paquete adaptador nuevo) |
| **V** | Simplicidad y dependencias fijadas | **Ninguna dependencia nueva del módulo.** Herramientas y acciones de la release —goreleaser como módulo de `tools/`, syft, cosign, `attest-build-provenance`, `cosign-installer`, `sbom-action/download-syft`— son las que el hito nombra (ROADMAP §3, «Cadena de suministro» y «Commits / versiones») y van a *Complexity Tracking*. Sin DI, ORM ni generador de CLI; `internal/cli` ni se extrae ni se toca | ✅ Cumple con justificación |
| **VI** | Un binario, convenciones de agente | `skills` es un applet más del mismo ejecutable: hereda las ocho banderas globales, `--describe` emite su esquema, `--dry-run` describe por el canal del kernel (D13). Las skills dejan de depender del despacho por `os.Args[0]` y lo invocan como `kitlegal <applet>` desde el `PATH` (ADR 0019), y el despacho multicall sigue en el kernel | ✅ Cumple |
| **VII** | Grafo y privacidad | Sin grafo (FR-054, H7). El manifiesto no lleva rutas absolutas, fechas, usuarios ni máquinas (FR-032). El único dato personal que entra en un artefacto es el `maintainer` de los paquetes, que es la identidad de autor ya pública en el historial (D26), anotado como supuesto para la persona | ✅ Cumple |
| **VIII** | Skills primero; el binario es la herramienta | Es el principio que el hito aplica: las skills viajan dentro del binario y se instalan con `kitlegal skills install`, **sin `scripts/`** y con la tabla de comandos en `kitlegal <applet> …` (constitución 2.1.0, ADR 0019). A Go va solo lo determinista y verificable (huellas, manifiesto, conflictos, versiones); el protocolo de las skills no cambia (FR-081) y lo miden sus evals sin tocarlas (FR-086) | ✅ Cumple |
| **IX** | Genericidad territorial, validación local | Sin dimensión territorial: el hito no toca `territorio`, `data/` ni ningún municipio. `legal-core` sigue siendo genérica; su única edición es la forma de invocar | ✅ Cumple (sin objeto territorial) |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H19 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni entrada y salida | `internal/core/instalacion` recibe las skills como bytes y el disco por puertos que define; no importa `os`, `io`, `io/fs`, `log`, ni el paquete raíz ni `internal/disco`. Hoy las reglas ejecutables no nombran ni `internal/disco` ni el paquete raíz, y los `_test.go` solo los ve `depguard` (el grafo de `TestArquitectura` sale de `go list -deps ./...`, sin tests: V47); por eso **R1 se amplía en el paso 3** (D32), y desde ahí alcanza también a los tests del dominio, que usan un disco en memoria | `depguard`, lista `core` (también `_test.go`), gana `github.com/jmorenobl/kitlegal/internal/disco` y el paquete raíz como entrada **exacta** `github.com/jmorenobl/kitlegal$` (V45, V46); `TestArquitectura` R1, que recorre todo `internal/core/**`, gana `disco` en `paquetesInternos` y el paquete raíz como denegación exacta (`==`, no `cuelgaDe`, que capturaría el módulo entero). Una sonda temporal (`package instalacion_test` que importa cada uno) pone `make lint` en rojo: `internal/disco` en el paso 3 y el paquete raíz en el paso 4, cuando existe | ✅ Cumple con la ampliación de R1 (paso 3) |
| Solo `internal/httpx` importa `net/http` | Ni el dominio, ni `internal/disco`, ni el paquete raíz, ni el applet lo importan; además, subprueba nueva: el dominio, el adaptador y el paquete raíz **no alcanzan `net`** (D32) | `depguard` lista `red` + `TestArquitectura` R2 + subprueba D32 | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Ningún paquete nuevo abre una base de datos | `depguard` lista `sql` + `TestArquitectura` R3 | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Ningún `package main` nuevo; el binario de e2e ya tiene su excepción acotada y gana solo constantes de composición | `forbidigo` `^os\.Exit$`, sin excepciones nuevas | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs con `slog` a stderr | El applet devuelve un `Resultado`; el aviso sale por `Presentador.Aviso` (stderr, D5); el dominio y el adaptador no escriben en ningún descriptor. `install.sh` no es Go | `forbidigo` `^fmt\.Print…$`, `^os\.Stdout$`, `^os\.Stderr$`, sin excepciones nuevas | ✅ Cumple |
| `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no existe (H7) | — (sin objeto) | ✅ Cumple |
| Los applets de ejemplo no se enlazan en el binario distribuido (ADR 0010) | El binario de e2e registra `skills` igual que el distribuido; el enlazador que falla vive en el `package main` de e2e, no en `internal/disco` | `depguard` lista `ejemplo` + `TestElBinarioNoEnlazaLosEjemplos` | ✅ Cumple |
| Ningún adaptador firma en el espacio reservado `kitlegal.` / `kitlegal:` (ADR 0006) | `skills` firma ahí y puede: es un applet calculado y su código vive en `internal/app`, no bajo `internal/source/` (V42) | `TestLasFuentesNoFirmanComoKitlegal` | ✅ Cumple |
| El binario no enlaza módulos no justificados (`modulosDelBinario`) | Ninguno nuevo: solo biblioteca estándar | `TestDependenciasDelBinario` en las seis plataformas, sin cambios en su lista | ✅ Cumple |

### Gates (constitución, «Gates»)

- **Capa 1 (mecánica)**, lo que H19 añade o toca (detalle en «Controles mecánicos»): la suite de aceptación congelada
  con su rojo-primero; los tests del dominio (tablas de conflictos y hallazgos, propiedades de FR-044 y FR-066, fuzz del
  manifiesto); la salida del applet contra `schemas/instalacion.json` y `make schema-check` sin deriva; `make
  skills-check` con el defecto de `scripts/` y las órdenes de las tablas empotradas; `goreleaser check` en `make ci`;
  la configuración de la release y los flujos comprobados por `TestConfiguracionDeLaRelease`; la ausencia de la
  instalación por enlaces (`TestSinInstalacionPorEnlaces`); el snapshot y el instalador contra él en CI; R1 ampliada a
  `internal/disco` y al paquete raíz, y la subprueba de arquitectura sin red (D32). En el workflow: guardián de diff con
  `[datos]` para `schemas/` y `testdata/`.
- **Capa 2 (jueces)**: que la edición de los `SKILL.md` se limita a la forma de invocar (FR-081; quickstart §6a) y que
  `README`, `CONTRIBUTING` y `CHANGELOG` están alineados con el `Makefile` (SC-020) lo juzgan los dos jueces de la
  revisión final; que las skills siguen activando y respondiendo igual, las evals del job con la instalación nueva
  (FR-128, SC-019), que mide el cierre del workflow.
- **Capa 3 (humano)**, siempre fuera del run: el esquema nuevo `schemas/instalacion.json`, los guiones **existentes**
  que cambian (`internal/app/testdata/script/argumentos.txtar` y los cuatro de `internal/skills/testdata/script/`) y
  los guiones activados, que el informe final enumera (FR-145); el `maintainer` de los paquetes (D26); y, tras
  fusionar, la etiqueta `v0.1.0`, `release.yml` con su humo y las dos órdenes en un Mac limpio (FR-150). **Ninguna
  pausa a mitad del run** (ADR 0018).

**Reglas del modo desatendido**: el ejecutor arregla el código, nunca un test ni un fixture congelado; no usa
`KITLEGAL_RECORD`, `make evals`, `make verify-sources`, ni ejecuta `release.yml`, `git tag` o `git push`; la única red
que usa es la del toolchain y el proxy de módulos de Go (crear `tools/goreleaser/go.sum`, como H0; `make vuln`), que no
es una fuente de datos (S10). Las tareas `[datos]` que tocan código además de `schemas/` o `testdata/` son dos,
razonadas en *Complexity Tracking*.

**Veredicto del gate: PASA.** Las piezas marcadas «con justificación» están en *Complexity Tracking*; ninguna afecta a
alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada.

### Re-evaluación tras la fase 1 (diseño)

- **`internal/core/instalacion` e `internal/disco`** (§IV, estructura): el primero es dominio en su sitio
  (`internal/core/`); el segundo es un adaptador fuera de la lista de §2, justificado en *Complexity Tracking*.
- **Paquete raíz `kitlegal`** (`skills.go` y sus tests): lo fija el hito (FR-002); no importa nada del módulo y solo
  `embed`/`io/fs` en producción.
- **Firma de composición `Arrancar(construir func(version string) …)` y `RegistroDeProduccion(version)`** (§IV): cambia
  la API de composición de `internal/app` (contrato puerto-y-applet §5 de H4), no el kernel de `internal/cli` ni el
  contrato `Applet` (ADR 0005). Tres llamadores fuera del paquete: `cmd/kitlegal`, el binario de e2e e
  `internal/evals/trazas.go` (V35); y, dentro de `internal/app`, los tests que la tabla «Tests existentes que cambian»
  lista en el paso 5 (fuera, `cmd/kitlegal/main_test.go:128` compila sin cambios).
- **El aviso no propaga el error de su escritura** (constitución, «sin atajos»): lo exige FR-072, está razonado (D5)
  y lo fija un test; va a *Complexity Tracking* para que no se lea como un error silenciado.
- **`.golangci.yml`** e **`internal/arch_test.go`**: `run.build-tags` gana la etiqueta `snapshot` (D33), y R1 gana
  denegaciones para el adaptador nuevo `internal/disco` y para el paquete raíz (entrada exacta), en `depguard` y en
  `TestArquitectura` (D32); ninguna regla se relaja: las únicas ediciones de reglas añaden denegaciones.

**Veredicto tras el diseño: PASA**, sin ninguna violación no justificada.

## Project Structure

### Documentation (this feature)

```text
specs/009-h19-instalar-sin-clonar/
├── spec.md
├── plan.md                  # este fichero
├── research.md              # D1-D34, V1-V48, S1-S15
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── applet-skills.md
│   ├── manifiesto.md
│   ├── aviso.md
│   ├── arnes-e2e.md
│   ├── release.md
│   └── skills-e-invocacion.md
├── aceptacion/              # T001: 18 guiones .txtar, congelados (ADR 0018)
├── checklists/
├── gates/
└── tasks.md                 # /speckit-tasks
```

### Source Code (repository root)

```text
skills.go                          NUEVO   paquete kitlegal: //go:embed de skills/*/SKILL.md y skills/*/references/*; Skills() fs.FS
skills_test.go                     NUEVO   TestSkillsEmpotradas (FR-001, FR-003, FR-004)
release_test.go                    NUEVO   TestSinInstalacionPorEnlaces (paso 12), TestConfiguracionDeLaRelease (pasos 14 y 15)
snapshot_test.go                   NUEVO   //go:build snapshot · TestSnapshot (FR-120, SC-016)
.goreleaser.yaml                   NUEVO   contracts/release.md §2
Makefile                           CAMBIA  install, release, goreleaser-check, snapshot-check, ci, help
.golangci.yml                      CAMBIA  run.build-tags + snapshot; depguard `core`: + internal/disco y github.com/jmorenobl/kitlegal$ (D32, paso 3)
.github/
├── workflows/ci.yml               CAMBIA  trabajo snapshot
├── workflows/release.yml          NUEVO   publicar + humo, solo en v*
├── workflows/evals.yml            CAMBIA  PATH explícito; kitlegal por command -v
└── dependabot.yml                 CAMBIA  /tools/goreleaser
tools/goreleaser/{go.mod,go.sum}   NUEVO   goreleaser v2.18.1
scripts/
├── install.sh                     NUEVO   contracts/release.md §7
├── instalar-skills.sh             SE RETIRA
├── skills-sync.sh                 CAMBIA  comentario: sin enlaces
└── evals.sh                       CAMBIA  kitlegal en el PATH antes de la primera sesión
skills/
├── boe-legislacion/SKILL.md       CAMBIA  kitlegal boe … (FR-081) y tabla regenerada
├── boe-legislacion/scripts/       SE RETIRA
├── legal-core/SKILL.md            CAMBIA  kitlegal territorio … y tabla regenerada
└── legal-core/scripts/            SE RETIRA
schemas/instalacion.json           NUEVO   [datos] desde --describe
cmd/kitlegal/main.go               SIN CAMBIO DE CÓDIGO: sigue pasando app.RegistroDeProduccion, que ahora recibe la versión de Arrancar
cmd/kitlegal/main_test.go          CAMBIA  paso 6b: appletsDelBinario = "applets disponibles: boe, skills, territorio" (l. 18; usada en 84, 90, 100) y caso «skills sin verbo»
internal/
├── arch_test.go                   CAMBIA  paso 3: R1 con `disco` en paquetesInternos (nueve) y el paquete raíz como denegación exacta; paso 9: subprueba sin red (D32)
├── core/instalacion/              NUEVO   dominio puro
│   ├── doc.go                     el paquete, sus puertos y sus invariantes
│   ├── puertos.go                 Disco, Escritor, Enlazador, Entrada
│   ├── skill.go                   SkillEmpotrada, FicheroEmpotrado
│   ├── ambito.go                  Ámbito local/global/dir, rutas presentadas
│   ├── version.go                 FormaSemVer, MismaVersion
│   ├── manifiesto.go              Manifiesto, LeerManifiesto, Bytes canónicos
│   ├── invocacion.go              validación de §2 del contrato del applet
│   ├── plan.go                    Planificar: conflictos (§4) y plan en cuatro fases (§5)
│   ├── aplicar.go                 Aplicar(plan, escritor)
│   ├── doctor.go                  Diagnosticar: hallazgos (§6)
│   ├── ordenes.go                 orden de shell POSIX de cada hallazgo, comillas
│   ├── aviso.go                   Aviso (contracts/aviso.md)
│   ├── salida.go                  SkillInstalada, Listado, Diagnostico…
│   ├── errores.go                 Invocación, Conflictos, Hallazgos, ÁmbitoIlegible (ConClase)
│   └── *_test.go                  disco en memoria (discoEnMemoria_test.go), tablas, propiedades, fuzz
├── disco/                         NUEVO   adaptador sobre el sistema de ficheros
│   ├── doc.go
│   ├── examinar.go                Examinar, Nombres (Lstat, Readlink, Stat del enlace)
│   ├── leer.go                    Huella, Leer (Lstat → Open → Stat → SameFile)
│   ├── escribir.go                CrearDirectorio, EscribirFichero (temporal + rename), Retirar
│   ├── enlazador.go               Enlazador del sistema: Disponible(directorio) (sonda en ese directorio, retirada), Enlazar
│   └── *_test.go                  t.TempDir(): enlaces, ciclos, tubería con nombre, permisos, atomicidad
├── app/
│   ├── instalacion.go             NUEVO   applet skills: AppletSkills(DependenciasDeSkills{versión, empotradas, Enlazador}), DependenciasDeSkillsDelSistema(version); verbos, argumentos Kong, procedencia
│   ├── empotradas.go              NUEVO   kitlegal.Skills() → []instalacion.SkillEmpotrada
│   ├── aviso.go                   NUEVO   composición del aviso (disco, HOME, versión, empotradas)
│   ├── main.go                    CAMBIA  Arrancar(construir func(version string)); aviso tras Analizar
│   ├── registro.go                CAMBIA  RegistroDeProduccion(version); Registro.Avisar; registra skills
│   ├── main_test.go               CAMBIA  paso 5: ayudante arrancar (l. 196-202) y cierres de TestArrancar con func(version string); caso nuevo: construir recibe la versión
│   ├── registro_test.go           CAMBIA  paso 5: RegistroDeProduccion("") (l. 239); paso 6b: lista boe, skills, territorio (l. 243) y los tres verbos de skills
│   ├── esquemas_test.go           CAMBIA  paso 5: RegistroDeProduccion("") (l. 428, 578); paso 6b: fila instalacion.json
│   ├── skills_test.go             CAMBIA  paso 5: RegistroDeProduccion("") (l. 168, 1066); paso 11: sin enlaces, kitlegal <applet>, TestOrdenesDeLasSkillsEmpotradas (detalle en «Tests existentes que cambian»)
│   ├── instalacion_test.go        NUEVO   verbos, códigos, salida contra el esquema
│   ├── aviso_test.go              NUEVO   invocaciones que avisan y que no; stdout y código intactos
│   ├── e2e_test.go                CAMBIA  arnés: binarios, variables, arbol, origen, proxies (contracts/arnes-e2e.md)
│   ├── ejemplo/kitlegal-e2e/main.go CAMBIA registra skills y el aviso; enlazador que falla (tipo del package main) sustituido en DependenciasDeSkills si lo elige una variable -X
│   └── testdata/script/
│       ├── argumentos.txtar       CAMBIA  [datos] lista de applets con skills
│       └── h19-*.txtar            NUEVOS  activación del workflow (copias congeladas de aceptacion/)
├── skills/
│   ├── enlaces.go, enlaces_test.go SE RETIRAN
│   ├── comandos.go, sincronia.go  CAMBIAN  kitlegal <applet>; sin enlaces; defecto de scripts/
│   ├── comandos_test.go           CAMBIA  paso 11: TestRenderizarTabla y TestSustituirRegion con kitlegal <applet> (detalle en «Tests existentes que cambian»)
│   ├── sincronia_test.go          CAMBIA  paso 11: sin enlaces ni destinoDeLosEnlacesDePrueba; defecto de scripts/ (detalle en «Tests existentes que cambian»)
│   ├── instalacion_test.go        CAMBIA  make install nuevo (integración)
│   └── testdata/script/instalar*.txtar CAMBIAN [datos]
└── evals/
    ├── preparar.go                CAMBIA  prueba de red con kitlegal boe
    ├── preparar_test.go           CAMBIA  paso 12: preguntaDeLaPruebaDeRed (l. 63-68) con kitlegal boe
    └── trazas.go                  CAMBIA  RegistroDeProduccion("")
README.md, CONTRIBUTING.md, CHANGELOG.md CAMBIAN
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`: dominio en `internal/core/instalacion`, composición en
`internal/app`, un adaptador por capacidad de E/S (`internal/disco`, nuevo), herramientas en `tools/`, contratos en
`schemas/`, fixtures en `testdata/` y la raíz del módulo solo con lo que el hito pone allí (`skills.go`,
`.goreleaser.yaml`).

## Aceptación e2e

**Aceptación e2e:** 18 guiones testscript, escritos por T001 desde el spec y el contrato
[contracts/arnes-e2e.md](./contracts/arnes-e2e.md) —con los formatos de salida de los contratos del producto que
enumera su §6: mensajes de validación de applet-skills §2, claves de `data` de §4, líneas de conflictos y hallazgos de
§5-§6, códigos de §8, la línea del aviso de aviso.md §4 y los mensajes de `install.sh` de release.md §7— en
`specs/009-h19-instalar-sin-clonar/aceptacion/`, congelados, y
activados al final como `internal/app/testdata/script/h19-*.txtar` (los ejecuta `TestEntregaDelHito` en `make ci`); los
dos `instalador-*` se ejecutan además contra el snapshot en `make snapshot-check` (antes de la activación, con copias momentáneas de
los congelados: contra el origen local en la tarea del paso 13 y contra el snapshot en las de los pasos 14 y 15;
obligación 6). Por historia de usuario:

| Guion | Historia | FR y SC que cubre |
|---|---|---|
| `skills-install-local` | US1 | US1.1, US1.4, US1.6 · FR-001, FR-003, FR-004, FR-010, FR-011, FR-014, FR-015, FR-016, FR-030, FR-031, FR-032, FR-050, FR-051, FR-122 (con el `kitlegal` sin inyecciones) · SC-001, SC-006, SC-021 |
| `skills-install-hosts` | US1 | US1.2, US1.3, US1.5 · FR-020, FR-021, FR-022, FR-023, FR-025 · SC-002, SC-003 |
| `skills-conflictos-entradas` | US1 | FR-040, FR-041 (a)-(d) y adopción del enlace esperado, FR-042, FR-043, FR-052 · SC-008, SC-009 |
| `skills-conflictos-rutas` | US1 | FR-022, FR-023, FR-026, FR-027, FR-035, FR-041 (g)-(h), FR-061, FR-067 · SC-009 |
| `skills-conflictos-dentro` | US1 | FR-028, FR-041 (e)-(g), FR-047 · SC-009 |
| `skills-dry-run` | US1 | FR-048 · SC-010 |
| `skills-invocacion` | US2 | US2.1, US2.2 · FR-081, FR-082, FR-084 · SC-014 |
| `skills-idempotencia` | US3 | US3.4 · FR-033, FR-034, FR-045, FR-046 · SC-007 |
| `skills-aviso` | US3 | US3.1, US3.2, US3.5, US3.6, US3.7 (metadatos de construcción), US3.9 (manifiesto local sin `HOME`) · FR-070, FR-071, FR-072, FR-075, FR-077 · SC-013, SC-022 |
| `skills-aviso-sin-aviso` | US3 | US3.3, US3.7 (`v` inicial), US3.8, US3.9 · FR-070, FR-073, FR-074, FR-076, FR-072 (`./.agents` como fichero y `./.agents/skills` como enlace colgando, con un global de otra versión: sin aviso) · SC-013 |
| `skills-list-doctor` | US4 | US4.1, US4.5 · FR-060, FR-061, FR-062, FR-067, FR-068 · SC-011 |
| `skills-doctor-hallazgos` | US4 | US4.2, US4.3, US4.4, US4.6, US4.7, US4.10, US4.11 · FR-065, FR-066, FR-077 · SC-011 |
| `skills-doctor-copia` | US4 | US4.3 (copia), US4.7 (sin `.claude/`), US4.8 · FR-024, FR-046, FR-069 · SC-012 |
| `skills-no-empotrada` | US4 | US4.9, US3.8 · FR-010, FR-036 · SC-011, SC-013 |
| `instalador-correcto` | US5 | US5.4, US5.6 (versión con y sin `v`, `KITLEGAL_INSTALL_DIR`, `HOME` vacío con directorio) · FR-100 a FR-107 · SC-017 |
| `instalador-rechazos` | US5 | US5.5, US5.6 (rechazos) · FR-100 a FR-104, FR-108 · SC-017 |
| `skills-ambito-global` | US6 | US6.1, US6.4 · FR-012 · SC-004 |
| `skills-ambito-dir` | US6 | US6.2, US6.3, US6.5 · FR-013, FR-052 (precedencia del exit 2 sobre el 1: `-g --dir` sin `HOME`, `install desconocida -g` con `HOME` vacío, skill desconocida con una carpeta ajena, `--dir --host claude` sobre un manifiesto con entradas de host) · SC-005 |

Lo que ningún guion del binario puede ejercer tiene su propia aceptación, también automática: US2 con
`TestInstalacion` (`make install` con `HOME` temporal: FR-125, FR-126, SC-018) y con el job de evals que mide el cierre
del workflow (FR-127, FR-128, SC-019); US5 con `TestConfiguracionDeLaRelease` (FR-090 a FR-098, FR-110 a FR-114,
FR-120), `make goreleaser-check` (FR-094) y `make snapshot-check` en el trabajo de snapshot (FR-095, FR-120, SC-016).
El trabajo de humo de `release.yml` y la instalación en un Mac limpio (FR-150, SC-023) son humanos, tras fusionar.

## Controles mecánicos que este hito añade o toca

### Objetivos del `Makefile`

| Objetivo | Cambio | En `ci` |
|---|---|---|
| `goreleaser-check` | **nuevo**: `$(GORELEASER) check` (FR-094) | **sí** |
| `release` | deja de fallar: construye el snapshot con `--snapshot --clean --skip=publish,sign,sbom` (FR-095) | no |
| `snapshot-check` | **nuevo**: `TestSnapshot` (etiqueta `snapshot`) y los guiones del instalador con `KITLEGAL_DIST` (FR-120); sin ningún guion `instalador-` en `internal/app/testdata/script/` falla, y dentro del run se verifica como fija la obligación 6 | no (trabajo de snapshot) |
| `install` | `go install` + `kitlegal skills install -g --host claude` con el binario recién instalado (FR-125) | no |
| `skills-sync` | sin enlaces; texto de `help` nuevo (FR-080, FR-121) | no |
| `skills-check` | gana `TestOrdenesDeLasSkillsEmpotradas` en su `-run` (FR-084) y, por `TestSkillsDelRepositorio`, el defecto de `scripts/` (FR-082) | sí |
| `ci` | añade `goreleaser-check` | — |
| `mod-verify` | cubre `tools/goreleaser` sin cambios (descubre `tools/*/go.mod`) | sí |

### Tests

| Test | Dónde | Qué fija |
|---|---|---|
| `TestSkillsEmpotradas` | `skills_test.go` (raíz) | lo empotrado = `skills/*/SKILL.md` + `references/**` del árbol, byte a byte; una skill sin `SKILL.md` no entra (FR-001, FR-003, FR-004) |
| `TestConfiguracionDeLaRelease` | `release_test.go` (raíz) | contracts/release.md §2, §3, §5 y §6 (§2 y §3 desde el paso 14; §5 y §6 en el 15, cuando existen los flujos): plataformas, las cuatro `-X` iguales al `Makefile`, secciones, tokens, disparadores, permisos, humo, trabajo de snapshot sin secretos (FR-090 a FR-097, FR-110 a FR-114, FR-120) |
| `TestSinInstalacionPorEnlaces` | `release_test.go` (raíz) | ni `SKILL.md`, ni `Makefile`, ni `.github/` nombran `scripts/boe`, `scripts/territorio` ni `bin/instalado`; no existen `scripts/instalar-skills.sh`, `internal/skills/enlaces.go` ni `skills/*/scripts` (FR-083, SC-014) |
| `TestSnapshot` | `snapshot_test.go` (raíz, `snapshot`) | contracts/release.md §4 (SC-016) |
| `TestFormaSemVer`, `TestMismaVersion` | `internal/core/instalacion` | D31, FR-073, FR-077 |
| `TestLeerManifiesto`, `TestManifiestoCanonico`, `FuzzLeerManifiesto` | `internal/core/instalacion` | contracts/manifiesto.md: cada regla de forma → ilegible; mismos bytes en dos «máquinas»; ningún pánico (FR-032, FR-035) |
| `TestValidarInvocacion` | `internal/core/instalacion` | contracts/applet-skills.md §2, en orden y sin examinar el disco (FR-010, FR-012, FR-013, FR-020) |
| `TestConflictos` | `internal/core/instalacion` | una fila por caso de data-model §4, cada uno con exactamente una clase (FR-041, SC-009) |
| `TestPlan` | `internal/core/instalacion` | estados, adopción, subconjunto, skill no empotrada intacta, repuesta, retirada de lo no empotrado, copia ↔ enlace, quitar host del manifiesto (FR-034, FR-036, FR-046, FR-047); `Disponible` solo se pregunta por el directorio de la sonda del ámbito (data-model §3), y con un `Enlazador` sintético que admite enlaces fuera del ámbito y no dentro, la segunda ejecución da `sin cambios` con el plan vacío y `--dry-run` da la misma salida que la orden real (FR-045, FR-048; D9) |
| `TestFalloAMitadSeCompleta` | `internal/core/instalacion` | FR-044 para cada operación del plan de varios escenarios (D7) |
| `TestHallazgos`, `TestOrdenesDeDoctor` | `internal/core/instalacion` | data-model §6: clase, ruta, orden, orden de la lista, `--host claude`, comillas (FR-065, FR-066); con el mismo `Enlazador` sintético de `TestPlan`, una copia declarada no da el hallazgo «copia», cuya orden no la arreglaría (FR-066 (i), FR-069; D9) |
| `TestOrdenesDeDoctorArreglan` | `internal/core/instalacion` | FR-066 (i) y (ii) sobre el disco en memoria, simulando `rm` e `install` |
| `TestAviso` | `internal/core/instalacion` | contracts/aviso.md §2-§4 (FR-070 a FR-073, FR-077) |
| `TestExaminar`, `TestLeerNoSigueEnlaces`, `TestNoAbreLoQueNoEsRegular`, `TestEscrituraAtomica`, `TestRetirar` | `internal/disco` | D6, D8 sobre un árbol real (tubería con `syscall.Mkfifo` solo en Unix con su etiqueta de compilación) |
| `TestEnlazadorDelSistema` | `internal/disco` | D9: la sonda se hace **en el directorio que se pasa** y lo deja con las mismas entradas y bytes; en un directorio sin permiso de escritura, `Disponible` es falso aunque `TMPDIR` admita enlaces; con `TMPDIR` apuntando a un fichero, un directorio escribible sigue dando verdadero (no usa `TMPDIR`); un nombre de sonda que ya existe se reintenta con otro sin tocar lo que había; retirar la sonda que falla es un error (con la retirada sustituida en el test interno del paquete) |
| `TestAppletSkills` | `internal/app/instalacion_test.go` | verbos, banderas, códigos 0/1/2, sobre de fallo con clase `inesperado`, `--describe`, `--dry-run` por `Ensayo` (FR-050 a FR-054), sobre un registro local del test y sin registrar el applet (paso 6a); con un `Enlazador` sintético que siempre falla en `DependenciasDeSkills`, la entrada de host queda como copia y `doctor` no la señala (FR-024, FR-069); `DependenciasDeSkillsDelSistema` lleva la versión, lo empotrado y un `Enlazador` no nulo |
| `TestSalidaDeSkillsContraSchemas` | `internal/app/instalacion_test.go` | toda salida correcta, también sin manifiesto, valida contra `schemas/instalacion.json` (FR-053); nace en el paso 6c, con el esquema ya publicado |
| `TestAvisoDelKernel` | `internal/app/aviso_test.go` | con un avisador sintético: lo llaman exactamente las invocaciones de FR-070; stdout byte a byte y código iguales; un `Aviso` que falla no cambia nada (D5) |
| `TestOrdenesDeLasSkillsEmpotradas` | `internal/app/skills_test.go` | FR-084 sobre los `SKILL.md` **empotrados** |
| `TestArbol` | `internal/app/e2e_test.go` | la orden `arbol` del arnés (formato, sin seguir enlaces, errores), como `TestCronometra` |
| `TestEntregaDelHito` | `internal/app/e2e_test.go` | los 18 guiones `h19-*` y los existentes |
| `TestInstalacion` | `internal/skills/instalacion_test.go` (`integration`) | contracts/skills-e-invocacion.md §5 (FR-125, FR-126, SC-018) |
| `TestSkillsDelRepositorio` | `internal/app/skills_test.go` | sin casos de enlaces; caso nuevo: `scripts/` es un defecto (FR-082) |
| `TestArquitectura` | `internal/arch_test.go` | R1 con `internal/disco` y el paquete raíz (exacto) desde el paso 3, y la subprueba sin red desde el paso 9 (D32, SC-022) |
| `TestArrancar` | `internal/app/main_test.go` | gana un caso: `construir` recibe exactamente la versión que recibe `Arrancar` (D4) |
| `TestRegistroDeProduccion` | `internal/app/registro_test.go` | el registro de producción es `boe`, `skills`, `territorio`, y `skills` con sus tres verbos (FR-010) |
| `TestPuntoDeEntrada` | `cmd/kitlegal/main_test.go` | la lista de applets del binario distribuido con `skills`; `kitlegal skills` sin verbo sale con 2 nombrando `skills` y enumerando sus tres verbos (contracts/applet-skills.md §1) |

### Tests existentes que cambian, por paso

Cada paso del orden de implementación rompe tests que ya existen: la tarea que lo implementa **declara estos ficheros
en sus rutas y los cambia en el mismo diff**, porque sin ellos el guardián de diff rechaza la tarea o `make ci` queda en
rojo. Barrido hecho con `git grep` sobre `RegistroDeProduccion`, `Arrancar(`, las listas literales de applets,
`scripts/`, `bin/instalado`, `EnlacesEsperados` y los destinos de enlace, en `*_test.go` y `*.txtar`.

| Paso | Fichero y líneas (en `main` hoy) | Qué lo rompe | Cambio |
|---|---|---|---|
| 5 | `internal/app/main_test.go:196-202` (ayudante `arrancar(t, construir func() (*Registro, error), …)`) y los cierres `construir` de `TestArrancar` (l. 220 y 256) | la firma nueva `Arrancar(…, construir func(version string) (*Registro, error), …)`: no compilan | pasan a `func(string) (*Registro, error)`, con el parámetro en blanco (`_ string`) en los cierres que no miran la versión (`revive`, `unused-parameter`); `TestArrancar` gana el caso de la versión, cuyo cierre sí la nombra (D4) |
| 5 | `internal/app/registro_test.go:239`, `internal/app/esquemas_test.go:428` y `:578`, `internal/app/skills_test.go:168` y `:1066` | `RegistroDeProduccion()` pasa a `RegistroDeProduccion(version string)`: no compilan | `RegistroDeProduccion("")` (un binario sin versión SemVer, que no avisa ni compara: FR-073) |
| 5 | `cmd/kitlegal/main_test.go:128` (`app.Arrancar(caso.argv, app.RegistroDeProduccion, …)`) | — | **no cambia**: el valor de función ya tiene la firma nueva, como en `cmd/kitlegal/main.go` |
| 6b | `internal/app/registro_test.go:243` (`[]string{"boe", "territorio"}`) y el comentario de `TestRegistroDeProduccion` (l. 228-235) | registrar `skills` en producción | `boe`, `skills`, `territorio`, y `skills` con sus tres verbos |
| 6b | `cmd/kitlegal/main_test.go:18` (`appletsDelBinario = "applets disponibles: boe, territorio"`, usada en l. 84, 90 y 100) y su comentario (l. 14-17) | registrar `skills` en producción: el despacho enumera `boe, skills, territorio` | la constante con `skills`; caso nuevo `skills` sin verbo (exit 2 y sus tres verbos) |
| 6b | `internal/app/testdata/script/argumentos.txtar:24` y `:30` | registrar `skills` en el binario de e2e | `boe, contar, echo, skills, territorio` |
| 6b | `internal/app/esquemas_test.go:58`, tabla `ficherosDeEsquemas` | `TestEsquemasCubrenTodosLosVerbos` exige la parte publicada de cada verbo del registro (V27) | fila `instalacion.json` |
| 6b | `internal/app/skills_test.go:165-210` (`TestTablaDeComandosCoincideConLaGramatica`) | — | **no se edita, pero se ejerce**: recorre `registro.Nombres()`, que ya incluye `skills`; `RenderizarTabla` exige a sus tres verbos las mismas banderas globales y el mismo sobre que a los de `boe` (`internal/skills/comandos.go:512` y `:540-553`) y cada fila, con `--describe`, sale con 0 describiendo su verbo |
| 11 | `internal/skills/comandos_test.go:667-674` (`filaDeBuscar`, `filaDeArticulo`), las `esperada` de `TestRenderizarTabla` (l. 733-765) y `nueva` de `TestSustituirRegion` (l. 848) | la tabla se titula y se escribe con `kitlegal <applet>` | `` ### `kitlegal boe` `` y `kitlegal boe buscar …` (y lo mismo con `ejemplo`) |
| 11 | `internal/skills/sincronia_test.go`: `regionDeAlfa` (l. 91-105); en `TestRegenerarYComparar`, los `Enlaces` esperados y las comprobaciones de `scripts/` de `escribir-sincroniza` (l. 200-238) y los enlaces y el fichero que `segundo-escribir-no-cambia-nada` crea en `scripts/` (l. 246-248); en `probarDerivas`, los casos `enlace-ausente`, `enlace-sobrante`, `enlace-en-una-skill-que-no-declara-applets`, `enlace-con-otro-destino` y `fichero-regular-en-lugar-de-enlace` (l. 409-470); en `probarDefectos`, la deriva de control de `beta` (l. 702-718, `scripts/uno` `enlace-sobrante`); en `TestEscribirSinAplicarUnArreglo`, el caso `enlazar-lo-que-otra-skill-ya-enlazo` (l. 898-914) | desaparecen las derivas y los arreglos de enlaces, `skills.Enlace` y `Regenerado.Enlaces`, y `scripts/` pasa a ser un defecto; `destinoDeLosEnlacesDePrueba` vive en `enlaces_test.go`, que se retira | `kitlegal dos`/`kitlegal uno` en la región; fuera los casos y comprobaciones de enlaces; la deriva de control de `probarDefectos` pasa a una que sigue existiendo (una referencia sobrante en `beta`); caso nuevo en `probarDefectos`: una entrada `scripts` en una skill es el defecto de FR-082 |
| 11 | `internal/app/skills_test.go`: la descripción de `-regenerar-skills` (l. 26-33), `destinoDeLosEnlacesDeScripts` y `enlaceSobrante` (l. 40-46), `casosDeEnlaces` (l. 672-728) y su uso en `TestSkillsDelRepositorio` (l. 122), los enlaces de `probarRegenerarDosVeces` (l. 818-826), `enlazarEnLaCopia` (l. 1359-1364) si queda sin uso; el literal `"scripts/boe articulo <norma> [--bloque]"` (l. 206) e `invocacionDeLaSintaxis` (l. 1187-1206) | lo mismo, y la sintaxis de la tabla pasa de `scripts/<applet> <verbo> …` a `kitlegal <applet> <verbo> …`: `invocacionDeLaSintaxis` toma el primer campo como programa y el segundo como verbo, y con la forma nueva el verbo es el tercero | fuera lo de enlaces (`unused` rechaza lo que quede sin uso); `"kitlegal boe articulo <norma> [--bloque]"`; `invocacionDeLaSintaxis` toma `kitlegal`, el applet y el verbo; caso nuevo en `TestSkillsDelRepositorio`: `scripts/` es un defecto (FR-082) |
| 11 | `internal/skills/enlaces_test.go` | se retira con `enlaces.go` | — |
| 12 | `internal/evals/preparar_test.go:63-68` (`preguntaDeLaPruebaDeRed`, con `~/.claude/skills/boe-legislacion/scripts/boe …`) | `preparar.go` escribe `kitlegal boe articulo …` | el texto literal nuevo de contracts/skills-e-invocacion.md §6 |

**Lo que no cambia**, y por qué, para que ninguna tarea lo «limpie»: las trazas y constantes de `internal/evals` que
nombran `scripts/boe` (`trazas_test.go:24` y sus casos, `juzgar_test.go:41`, `internal/evals/testdata/sesiones/**`)
prueban el lector de trazas, que reconoce el applet por el nombre de invocación multicall y acepta `scripts/boe` y
`kitlegal boe` (`internal/evals/trazas.go:221`); `TestSinInstalacionPorEnlaces` no las vigila (solo `SKILL.md`,
`Makefile` y `.github/`). `internal/app/testdata/script/ayuda.txtar` compara por expresión regular y `territorio` sigue
siendo el nombre más largo, así que `skills` no mueve la columna.

**Orden entre pasos.** `TestSinInstalacionPorEnlaces` exige que `.github/` no nombre `bin/instalado`, y
`.github/workflows/evals.yml:147` lo nombra hasta el paso 12: por eso ese test nace en el paso 12, no en el 11.

### Fixtures, `testdata/` y `schemas/` (tareas `[datos]`, FR-145)

| Fichero | Cambio | Por qué |
|---|---|---|
| `schemas/instalacion.json` | nuevo, generado con `TestEsquemasPublicados -actualizar-esquemas` | FR-053, D16 |
| `internal/app/testdata/script/argumentos.txtar` | la lista literal de applets gana `skills` (líneas 24 y 30) | V29: registrar el applet la cambia |
| `internal/skills/testdata/script/instalar.txtar`, `instalar-de-nuevo.txtar`, `instalar-con-conflicto.txtar`, `instalar-sin-gobin.txtar` | reescritos (contracts/skills-e-invocacion.md §5) | FR-125, FR-126 |
| `internal/app/testdata/script/h19-*.txtar` | los copia el workflow al activar la suite | ADR 0018 |

Ningún fixture de fuente ni grabación nueva. El corpus de `FuzzLeerManifiesto` va en `f.Add` dentro del test, sin
ficheros en `testdata/fuzz/`.

### CI

`ci.yml` gana el trabajo `snapshot` (contracts/release.md §5); `release.yml` es nuevo (§6); `evals.yml` cambia
(contracts/skills-e-invocacion.md §6); `dependabot.yml` gana `tools/goreleaser`. El trabajo `ci` sigue siendo `make ci`.

## Datos externos

**Ninguno.** H19 no graba ni deriva datos de ninguna fuente: no hay manifiesto `grabaciones.json`, ni test `//go:build
grabacion`, ni paso `grabar_datos`, ni material en `evidencias/`, ni fila nueva en `docs/SOURCES.md`. Lo empotrado sale
de `skills/`, que ya está en el repositorio y lo genera `make skills-sync` desde `data/`. La única red que usa una tarea
es la del proxy de módulos de Go para `tools/goreleaser/go.sum` (S10), que no es una fuente de datos.

## Orden de implementación (de dentro afuera)

1. **T001 `[aceptacion]`**: los 18 guiones contra contracts/arnes-e2e.md y los formatos de los contratos que enumera su
   §6, con la precondición en rojo; ningún código.
2. **Dominio** `internal/core/instalacion`, por capas y con sus tests: versiones → manifiesto → ámbito e invocación →
   puertos y disco en memoria → conflictos y plan → aplicar (FR-044) → doctor y órdenes (FR-066) → aviso.
3. **Adaptador** `internal/disco` con sus tests sobre un árbol real (el `Enlazador` con la sonda en el directorio que
   se le pasa, D9), y **R1 ampliada** en la misma tarea (D32): `internal/disco` y `github.com/jmorenobl/kitlegal$` en la
   lista `core` de `depguard` (`.golangci.yml`), `disco` en `paquetesInternos` y el paquete raíz como denegación exacta
   en `internal/arch_test.go`. Verificación: una sonda temporal —`internal/core/instalacion/sonda_test.go`, `package
   instalacion_test`, con `import _ "github.com/jmorenobl/kitlegal/internal/disco"` sola en su línea— pone `make lint`
   en rojo nombrando R1; se retira antes de `make ci`. (La entrada del paquete raíz ya está, pero ese paquete aún no
   existe: importarlo no compilaría y el lint no llegaría a `depguard`.)
4. **Empotrado** `skills.go` + `TestSkillsEmpotradas`, y `internal/app/empotradas.go`. Verificación: la misma sonda
   temporal, ahora con `import _ "github.com/jmorenobl/kitlegal"`, pone `make lint` en rojo nombrando R1, y la sonda
   se retira antes de `make ci`.
5. **Composición**: `Arrancar(construir func(version string))`, `RegistroDeProduccion(_ string)` y sus llamadores
   (`cmd/kitlegal`, e2e, `internal/evals/trazas.go`), sin registrar todavía el applet; con los tests existentes que la
   firma rompe («Tests existentes que cambian», paso 5): `internal/app/main_test.go`, `registro_test.go`,
   `esquemas_test.go` y `skills_test.go`. Como en este paso nadie usa todavía la versión dentro de
   `RegistroDeProduccion` ni de `registroDeE2E`, sus parámetros van **en blanco** (`_ string`, como
   `internal/app/main_test.go:407` con `func(_ context.Context)`), igual que en los cierres `construir` de los tests que
   no la miran; si no, `revive` (reglas por omisión, `unused-parameter`) deja `make lint` en rojo. El paso 6b los
   nombra al pasar la versión al applet.
6. **Applet `skills`, en tres tareas, con el patrón D16 de H6** (H6 T009-T011):
   - **6a, sin `[datos]`**: `internal/app/instalacion.go` con los tres verbos (argumentos Kong, la validación del
     dominio, la composición con el disco y lo empotrado, la procedencia, el exit 1, `--dry-run` y `--describe`) y
     `TestAppletSkills` sobre un registro local del test, **sin registrar el applet** ni en producción ni en e2e. El
     constructor exportado sigue el patrón de `AppletBoe(DependenciasDeBoe)`: `AppletSkills(DependenciasDeSkills)`, con
     la versión del binario, lo empotrado y el puerto `Enlazador`, que el applet **recibe y no compone dentro** para
     que el creador de enlaces pueda sustituirse en test y en el binario de e2e (FR-024); y
     `DependenciasDeSkillsDelSistema(version)`, las del binario distribuido con el `Enlazador` del sistema de
     `internal/disco`, como `DependenciasDeRed()` para `boe`. `TestAppletSkills` pasa un `Enlazador` sintético que
     siempre falla y comprueba el recurso de copia (FR-024, FR-069).
   - **6b, `[datos]` indivisible y nada más**: el registro en `internal/app/registro.go` y en
     `internal/app/ejemplo/kitlegal-e2e/main.go` (que nombran aquí el parámetro de versión y registran
     `AppletSkills(DependenciasDeSkillsDelSistema(version))`, con el `Enlazador` del sistema en los dos) +
     `schemas/instalacion.json` generado + fila de `esquemas_test.go` + las tres listas literales (`argumentos.txtar`,
     `internal/app/registro_test.go`, `cmd/kitlegal/main_test.go`) (*Complexity Tracking*).
   - **6c, sin `[datos]`**: `TestSalidaDeSkillsContraSchemas`, la salida real contra el esquema ya publicado.
7. **Aviso** en la composición y en el kernel de `internal/app`, con `aviso_test.go`.
8. **Arnés e2e** (contracts/arnes-e2e.md): binarios con versión y enlazador que falla (un tipo del `package main` de
   e2e que `registroDeE2E` pone en el campo `Enlazador` de `DependenciasDeSkillsDelSistema(version)` cuando una
   variable de cadena `-X` lo elige, sin tocar `internal/app/instalacion.go`), variables, `arbol` +
   `TestArbol`, origen de release local, proxies. Desde aquí, la suite de aceptación se puede ejecutar copiándola un
   momento a `testdata/script/` para medir el avance.
9. **Arquitectura**: subprueba D32. Solo añade un test, sin implementación: la propiedad la dan los pasos 3 y 4; su
   verificación es un mutante temporal (una importación de `net` en el dominio) que la pone en rojo y se retira antes
   de `make ci`, como la tarea 6c solo añade `TestSalidaDeSkillsContraSchemas`.
10. **`[datos]` indivisible**: `make install` nuevo + `TestInstalacion` y sus cuatro guiones (*Complexity Tracking*).
11. **Skills**: `internal/skills` sin enlaces y con el defecto de `scripts/`; `SKILL.md` en `kitlegal <applet>`; `make
    skills-sync`; retirar `skills/*/scripts/`, `scripts/instalar-skills.sh`, `enlaces.go` y `enlaces_test.go`;
    `TestOrdenesDeLasSkillsEmpotradas`; con los tests existentes del paso 11 (`internal/skills/comandos_test.go`,
    `internal/skills/sincronia_test.go`, `internal/app/skills_test.go`).
12. **Job de evals**: `evals.yml`, `evals.sh`, `preparar.go` y `internal/evals/preparar_test.go`; y
    `TestSinInstalacionPorEnlaces` (`release_test.go`), que nace aquí porque `.github/workflows/evals.yml:147` nombra
    `bin/instalado` hasta este paso.
13. **`scripts/install.sh`**, antes que la release porque `.goreleaser.yaml` lo nombra en `checksum.extra_files` y
    goreleaser falla en el snapshot con una ruta literal que no existe (V48). Verificación: los guiones
    `instalador-*` congelados, copiados **un momento** a `internal/app/testdata/script/zz-<nombre>.txtar`, pasan con el
    origen local del arnés (`go test -count=1 -run '^TestEntregaDelHito$/instalador-' ./internal/app/`); las copias se
    retiran con `rm -f` antes de `make ci` y del commit, y `git status --porcelain internal/app/testdata/script` vacío
    lo confirma (research D28, «Cómo se verifica dentro del run»).
14. **Release**: `tools/goreleaser`, `.goreleaser.yaml`, objetivos `goreleaser-check` (en `ci`), `release` y
    `snapshot-check`, `TestConfiguracionDeLaRelease` en lo que fija de `.goreleaser.yaml` y del `Makefile`
    (contracts/release.md §2 y §3), `TestSnapshot`, etiqueta `snapshot` en `.golangci.yml`, Dependabot. Verificación:
    `make release` (S9); `make snapshot-check` **sin** guiones `instalador-` en `testdata/script/` sale con error y el
    mensaje `KITLEGAL_DIST: ningún guion instalador- que ejecutar` (contracts/arnes-e2e.md §5: la garantía de no pasar
    en vacío, ejercida); y con las copias momentáneas del paso 13, `make snapshot-check` completo en verde; las copias
    se retiran antes de `make ci`.
15. **CI**: trabajo `snapshot` en `ci.yml`, `release.yml`, y lo que `TestConfiguracionDeLaRelease` fija de los dos
    flujos (contracts/release.md §5 y §6), que no puede comprobarse antes de que existan; la misma verificación que el
    paso 14 (`make release` y `make snapshot-check` con las copias momentáneas, retiradas antes de `make ci`).
16. **Documentación**: README, CONTRIBUTING, CHANGELOG (*Unreleased*).

## Complexity Tracking

| Violación o pieza nueva | Por qué es necesaria | Alternativa más simple rechazada porque |
|---|---|---|
| Paquete adaptador `internal/disco`, fuera de la lista de adaptadores de ROADMAP §2 | El disco local es una capacidad de E/S que ningún adaptador existente cubre; la lectura sin seguir enlaces, la escritura atómica y el recurso de copia necesitan tests propios contra un árbol real (D2). Entra en R1 en la misma tarea que lo crea, en `depguard` y en `TestArquitectura` (D32), para que ningún test del dominio pueda importarlo | Meterlo en `internal/app` pondría E/S en la raíz de composición; `internal/store` es el repositorio SQLite del grafo del asunto (H10) |
| Paquete de dominio `internal/core/instalacion`, no listado en `CLAUDE.md` (`internal/core/{territorio,cita,…}`) | Las tablas de FR-041/FR-065 y las propiedades de FR-044/FR-066 solo se prueban de forma exhaustiva sin E/S (D1) | Lógica en el applet o en un adaptador: sin disco en memoria y fuera del umbral de `internal/core/**` |
| goreleaser v2.18.1 como módulo de `tools/` | Lo exige el hito (FR-096) y ROADMAP §3 lo sitúa en H19; versión fijada por `go.sum` como las demás herramientas (D20) | `goreleaser-action`: una segunda forma de fijar la versión, y no daría `goreleaser check` en `make ci` |
| syft y cosign, instalados solo en `release.yml` | SBOM y firma keyless de la release (FR-093; ROADMAP §3, «Cadena de suministro»); fuera de `tools/` por la clarificación del spec (FR-096) | En `tools/`: dos módulos grandes que ningún control de `make ci` usa |
| Acciones nuevas en `release.yml`: `actions/attest-build-provenance`, `sigstore/cosign-installer`, `anchore/sbom-action/download-syft` | Atestación SLSA (FR-112) y las dos herramientas anteriores en el runner; fijadas por su etiqueta mayor como el resto de flujos (V36; versiones: S2) | Instalar con `go install` en el runner: descarga y compila dos grafos grandes en cada release sin ganar verificación |
| `[datos]` que mezcla código: registrar `skills` (en `registro.go` y en el binario de e2e) + `schemas/instalacion.json` + fila de `esquemas_test.go` + `argumentos.txtar` + `internal/app/registro_test.go` + `cmd/kitlegal/main_test.go` (paso 6b), **y nada más** | En cuanto el verbo está registrado, `TestEsquemasCubrenTodosLosVerbos` exige su parte publicada (V27), el esquema se genera desde el applet registrado, y las tres listas literales de applets —`argumentos.txtar:24,30` (V29), `registro_test.go:243` y `appletsDelBinario` de `cmd/kitlegal/main_test.go:18`— cambian a la vez: cualquier otro orden deja `make ci` en rojo. El cuerpo del applet no está atado a ningún fichero protegido: va antes, en 6a, probado sobre un registro local del test sin registrarlo, y la salida contra el esquema publicado después, en 6c. Mismo patrón que D16 de H6 (H6 T009-T011) | Registrar en una tarea y publicar el esquema en otra deja `make ci` en rojo; meter el applet entero en la tarea `[datos]` mezcla código que ningún control hace inseparable del esquema |
| `[datos]` que mezcla código: receta `install` del `Makefile` + `instalacion_test.go` + sus cuatro guiones (paso 10) | Los guiones de `internal/skills/testdata/script/` comprueban exactamente lo que hace la receta: cambiar una sin los otros deja `TestInstalacion` en rojo | Dos tareas: cualquiera de los dos órdenes deja `make test-integration` en rojo |
| El aviso no propaga el error de escribir su línea (D5) | FR-072: el aviso nunca cambia el código de salida ni la salida estándar; un test lo fija | Propagarlo violaría FR-072; registrarlo iría al mismo descriptor roto |
| Cambio de firma de `app.Arrancar` y `app.RegistroDeProduccion` (API de composición, contrato puerto-y-applet §5 de H4) | La versión del binario tiene que llegar al applet y al aviso sin una quinta inyección (FR-091) ni estado global (D4) | Quinta `-X`: la prohíbe FR-091; llevarla en `schema.Contexto`: cambiaría el kernel para todos los applets |

## Obligaciones que este plan traslada a `tasks.md`

1. **T001 es la única tarea `[aceptacion]`** y la primera: escribe los 18 guiones de «Aceptación e2e» en
   `specs/009-h19-instalar-sin-clonar/aceptacion/`, cada uno con la precondición de contracts/arnes-e2e.md §1 como
   **primeras órdenes** y un comentario de cabecera con sus FR/SC; del arnés usa solo lo de contracts/arnes-e2e.md, y
   cada formato de salida que afirma lo copia del spec, si lo fija de forma literal, o del contrato del producto que
   nombra contracts/arnes-e2e.md §6 (mensajes de validación, applet-skills §2; claves de `data`, §4; líneas
   `<clase>: <ruta>` y `<clase>: <ruta>: <orden>` con su cabecera y su orden, §5-§6; códigos de salida, §8; la línea
   del aviso, aviso.md §4; el prefijo `install.sh: `, la línea `export PATH='<dir>':"$PATH"` y la última línea,
   release.md §7), sin inventar otro, de modo que las tareas que implementan esos contratos puedan satisfacerla; sin
   código de producto y sin evals nuevas (FR-086: las evals no se tocan).
2. Las rutas del arnés, los nombres de los binarios, las variables y el formato de `arbol` son los de
   contracts/arnes-e2e.md, tal cual: la suite congelada depende de ellos.
3. Registrar el applet (producción y e2e), publicar `schemas/instalacion.json`, añadir su fila y actualizar las tres
   listas literales de applets —`argumentos.txtar`, `internal/app/registro_test.go:243` y `appletsDelBinario` de
   `cmd/kitlegal/main_test.go:18`— van en **una** tarea `[datos]` (paso 6b), y nada más: el cuerpo del applet y
   `TestAppletSkills`, sin registrarlo, en la tarea anterior sin `[datos]` (6a), y `TestSalidaDeSkillsContraSchemas`
   en la siguiente, también sin `[datos]` (6c); `make install` y `TestInstalacion` con sus guiones, en **otra** tarea
   `[datos]` (paso 10).
4. Cada tarea deja `make ci` en verde; la que añade `goreleaser-check` a `ci` crea antes (o a la vez) `tools/goreleaser`
   y `.goreleaser.yaml`, porque sin ellos `make ci` fallaría; y `scripts/install.sh` existe antes que `.goreleaser.yaml`
   (paso 13), porque `make release` falla sin él (V48). Ninguna tarea comprueba lo que un paso posterior crea: un test
   que lee un fichero nace en la tarea que lo crea o después (`TestSinInstalacionPorEnlaces` en el paso 12, lo de
   `release.yml` y `ci.yml` de `TestConfiguracionDeLaRelease` en el 15), y el arnés del paso 8 exporta
   `KITLEGAL_INSTALADOR` sin exigir que `install.sh` exista.
5. Ninguna tarea ejecuta `release.yml`, empuja, etiqueta, usa `PUBLISHER_TOKEN`, `KITLEGAL_RECORD`, `make evals` ni
   `make verify-sources`; el cierre (CI y evals remotas) lo hace el workflow. Ninguna tarea declara rutas de `evals/`
   (FR-086), así que el guardián de diff rechaza cualquier cambio en ellas.
6. `make snapshot-check` completo necesita guiones `instalador-` en `internal/app/testdata/script/`, que solo llegan
   con la activación de la suite, después del bucle de tareas (contracts/arnes-e2e.md §5; research D28, «Cómo se
   verifica dentro del run»). Por eso las tareas de `install.sh` (paso 13), de la release (paso 14) y de CI (paso 15)
   usan copias momentáneas de los guiones congelados `aceptacion/instalador-*.txtar` en
   `internal/app/testdata/script/zz-<nombre>.txtar`, que no se editan, no se declaran y se retiran con `rm -f` antes de
   `make ci` y del commit: la del paso 13 los ejecuta contra el origen local; la del paso 14 ejecuta `make release`
   (S9: que nfpm acepta la versión de snapshot), comprueba que `make snapshot-check` **sin** las copias sale con error
   por falta de guiones `instalador-` y lo ejecuta completo **con** ellas; la del paso 15, `make release` y `make
   snapshot-check` completo con ellas. `install.sh` va antes que la release porque el snapshot lo necesita (V48), y
   `TestConfiguracionDeLaRelease` gana lo de `release.yml` y `ci.yml` en el paso 15, que es cuando existen. La línea
   de esas tareas en `tasks.md` remite a este procedimiento («obligación 6 del plan; research D28») **sin nombrar la
   ruta de las copias**: no es una ruta de su diff, las tareas no llevan `[datos]` y el precheck de `tasks` leería
   `testdata/` sin la etiqueta. La salida en `dist/` está ignorada por git.
7. Tareas explícitas de Definition of Done: `CHANGELOG.md` (*Unreleased*), `README.md` y `CONTRIBUTING.md` alineados con
   el `Makefile`, `schemas/instalacion.json`, y la lista de ficheros de `testdata/` y `schemas/` tocados para el informe
   final (FR-145). Sin ADR nuevo (ADR 0019 ya decide) ni fila de `docs/SOURCES.md` (ninguna fuente).
8. Las menciones desfasadas fuera de alcance (p. ej. el diagrama de ROADMAP §2 con `scripts/<applet>`, la fila de
   «Commits / versiones» de ROADMAP §3) van a *Pendientes* de la propuesta de cambio, no a una tarea.
9. El `maintainer` de los paquetes (D26) se anota en `gates/supuestos.md` para el informe final.
10. Cada tarea declara en sus rutas, y cambia en su diff, los tests existentes que su paso rompe según «Tests
    existentes que cambian, por paso» (pasos 5, 6b, 11 y 12), y no toca los que esa tabla da por no cambiados.
    `TestSinInstalacionPorEnlaces` va en la tarea del paso 12, no en la del 11. La ampliación de R1 (`.golangci.yml` e
    `internal/arch_test.go`) va en la tarea del paso 3, que declara los dos ficheros; la sonda temporal
    `internal/core/instalacion/sonda_test.go` se crea y se retira dentro de la tarea del paso 3 y de la del 4, sin
    quedar en ningún diff.

## Comprobación contra la rúbrica del juez (`juez_plan`, a-l) y `precheck.sh plan`

| Criterio | Dónde se cumple |
|---|---|
| a. constitution_check | «Constitution Check»: un ítem por principio I-IX, uno por regla de dependencia (seis de ROADMAP §2 y tres de control del binario), gates por capa y re-evaluación |
| b. dependencias | Ninguna dependencia nueva del módulo; herramientas y acciones en *Complexity Tracking* |
| c. reglas_dependencia | Tabla de reglas; dominio sin E/S, adaptador sin red, stdout solo por el presentador, `os.Exit` sin excepciones nuevas; R1 ampliada a `internal/disco` y al paquete raíz (exacto) en `depguard` y en `TestArquitectura` desde el paso 3, con sonda en rojo (D32, V45-V47) |
| d. errores_exit_codes | data-model §9; contracts/applet-skills.md §2, §5 y §8; `schema.ConClase` |
| e. tests_primero | «Aceptación e2e» (18 guiones congelados con su FR/SC), «Tests», «Tests existentes que cambian, por paso» (obligación 10), «Fixtures», «Objetivos del Makefile»; `make snapshot-check` dentro del run, obligación 6 |
| f. alcance | research D34 y *Fuera de alcance* del spec; nada que no pida un FR |
| g. sin_atajos | Ningún `//nolint`, `t.Skip` ni TODO previsto; el único error no propagado está razonado y probado (D5, *Complexity Tracking*) |
| h. mejor_alternativa | research D1-D34, cada una con su alternativa rechazada; D9 rechaza la sonda en un temporal propio y la retirada diferida de la copia |
| i. afirmaciones_verificadas | research V1-V48 con `fichero:línea` u orden local; S1-S15 como supuestos |
| j. quickstart_ejecutable | quickstart.md: órdenes y rutas reales, efectos declarados (`bin/`, `dist/`, cobertura, `$T`) y comprobación final de `git status` |
| k. datos_externos | «Datos externos»: ninguno; ninguna fuente, grabación ni evidencia |
| l. autonomia | Obligaciones 5 y 6; capa 3 solo antes y después del run |
| precheck | `## Constitution Check` presente; línea «Aceptación e2e:» presente; ninguna marca de aclaración pendiente; `research.md` presente |
