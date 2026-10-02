# Research · H22 · Instalar sin terminal

Fase 0 del plan. Modo desatendido: cada decisión se tomó con el «Criterio de decisión autónoma» de la constitución y
lleva su alternativa rechazada. Toda afirmación sobre una herramienta o dependencia externa remite a la tabla V
(comprobada en local, en esta sesión) o a la tabla S (supuesto que no se pudo comprobar sin red, sin publicar una release
o sin ejecutar `claude`).

**Material de un solo uso, fuera de lo versionado** (en el directorio temporal de la sesión; nada entra en el
repositorio):

- Un worktree de `751b76e` (la cabeza de la rama, igual a `main`) con el **prototipo del paso**: `cmd/empaquetar`,
  `internal/empaquetado` (un fichero), `app.HerramientasAnunciadas` y `.goreleaser.yaml` con `universal_binaries`, el
  gancho, `archives.ids` y los `extra_files`. Sobre él se ejecutaron `goreleaser check`, cuatro snapshots (el del plan,
  uno con el gancho roto, uno sin `ids` y otra vez el del plan), un test de etiqueta `snapshot` con las comprobaciones
  que el plan pide a `TestSnapshot` (`zz_h22_test.go`), `golangci-lint` y los guiones `instalador-`. Retirado al
  terminar (`git worktree list` solo da el árbol del repositorio).
- `h22-humo.sh` y `h22-pasos-humo.sh`: los cuerpos de los pasos nuevos de `humo`, ejecutados contra el `dist/` del
  prototipo con `cp` en lugar de `gh`.
- `h22-catalogo-paso/`: el cuerpo del paso que publica el catálogo, con un `gh` de mentira; y `h22-quickstart.sh`, los
  escenarios §2 a §5 del quickstart sobre el worktree.

Nada de ello abrió una sesión con modelo, ejecutó `claude` ni leyó credenciales, y no usó la red. Tres
lecturas se hicieron fuera del repositorio y de la caché de módulos, de solo lectura, porque son el código de la
herramienta de la que se afirma algo y están en el equipo: el paquete `@anthropic-ai/mcpb` 2.1.2 de la caché de `npx`
(V11), el binario de Claude Code 2.1.284 (V12; leído con `grep -a` y `tail -c`, sin ejecutarlo) y la ayuda de `gh`
2.101.0 con un directorio de configuración vacío (V22; el suyo, que guarda la credencial, está vetado a la sesión y no
se leyó).

## V · Verificado en local, en esta sesión

`GR` es `$(go env GOMODCACHE)/github.com/goreleaser/goreleaser/v2@v2.18.1`, la versión que fija
`tools/goreleaser/go.mod`.

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | goreleaser ejecuta sus fases en este orden: ganchos `before` → `build` → `universal binaries` → … → `archive` → `nfpm` → `sbom` → `checksums` → `sign` → … → `publish` → `artifacts.json`. La configuración solo admite ganchos en cuatro sitios: `before.hooks` (antes de compilar), `builds[].hooks` (uno por plataforma), `universal_binaries[].hooks` y `dockers_v2[].hooks`; ninguno en `archives` ni en `checksum` | `GR/internal/pipeline/pipeline.go:82-175`; `GR/pkg/config/config.go:526, 627, 1262, 1306-1308` |
| V2 | `universal_binaries[]` admite `id`, `ids`, `name_template`, `replace`, `hooks.pre`, `hooks.post` y `mod_timestamp`. Sin `ids`, vale `[id]`; sin `id`, el nombre del proyecto. El binario queda en `dist/<id>_darwin_all/<nombre>`, con modo `0755`. El gancho `post` se ejecuta después de escribirlo, con `.Path` como su ruta. Con `replace: false`, los dos binarios de macOS siguen entre los artefactos | `GR/pkg/config/config.go:622-629`; `GR/internal/pipe/universalbinary/universalbinary.go:36-48, 54-78, 151` |
| V3 | El universal es una cabecera de enteros de 32 bits en big-endian —la magia `0xcafebabe`, el número de arquitecturas y, por cada una, CPU, subtipo, desplazamiento, tamaño y alineación (14)— seguida de cada binario copiado tal cual en su desplazamiento, alineado a 16 384 bytes. El orden de las arquitecturas es el de la lista de artefactos, que no está fijado | `universalbinary.go:136-137, 162-222`. Medido: amd64 (CPU `0x01000007`) en el desplazamiento 16 384, 26 548 416 bytes; arm64 (`0x0100000c`) en 26 574 848, 25 219 954 bytes |
| V4 | Un gancho es una orden con plantilla (`cmd`, `dir`, `env`, `output`): se parte en palabras sin pasar por un shell, se ejecuta con el entorno del proceso de goreleaser y, si sale con un código distinto de 0, goreleaser termina con 1 y enseña lo que la orden escribió | `universalbinary.go:83-121`; `GR/internal/shell/shell.go:20-56`. Medido con el gancho apuntando a un binario que no existe: `release failed … post hook failed: exit status 1`, con la línea `empaquetar: falta el binario de Windows: …`, código 1, y `dist/` sin archivos ni `checksums.txt` |
| V5 | La plantilla del gancho tiene `.Version` —la etiqueta sin `v`; en un snapshot, la del snapshot— y `.Path`. En el snapshot del prototipo, `-version 0.3.2-SNAPSHOT-751b76e -macos dist/kitlegal-universal_darwin_all/kitlegal` | `GR/internal/tmpl/tmpl.go:49-99`; registro del snapshot |
| V6 | `archives[]` empaqueta los binarios y los universales cuyo `id` está en su `ids`; sin `ids`, todos. Con `universal_binaries` sin `id` propio y `archives` sin `ids`, sale un séptimo archivo, `kitlegal_darwin_all.tar.gz`; con `id: kitlegal-universal` y `archives[0].ids: [kitlegal]`, salen los seis de hoy | `GR/internal/pipe/archive/archive.go:117-127`. Medido: los dos snapshots, con siete y con seis archivos |
| V7 | `builds[].id` vale el nombre del proyecto si no se escribe, y cada binario queda en `dist/<id>_<objetivo>/`: `dist/kitlegal_windows_amd64_v1/kitlegal.exe`, `dist/kitlegal_darwin_arm64_v8.0/kitlegal`… El sufijo (`_v1`, `_v8.0`) es de goreleaser | `GR/internal/pipe/build/build.go:122-125, 237-245`; `dist/artifacts.json` del prototipo |
| V8 | `checksum.extra_files` se resuelve al calcular las huellas, después del gancho: `./dist/kitlegal.mcpb` y `./dist/kitlegal-plugin.zip` ya existen. `checksums.txt` queda ordenado por nombre. El universal no entra en `checksums.txt` ni se sube: su tipo no está entre los que se publican | `GR/internal/pipe/checksums/checksums.go:171-200`; `GR/internal/artifact/artifact.go:136-154`. Medido: 13 líneas, con `kitlegal-plugin.zip` y `kitlegal.mcpb` y sin ninguna del universal |
| V9 | `release.extra_files` se resuelve al publicar, con la misma función, y cada fichero se sube con su nombre | `GR/internal/pipe/release/release.go:162-175`. No ejecutado: publicar es de la release (S6) |
| V10 | `goreleaser check` sale con 0 con las secciones nuevas (`builds[0].id`, `universal_binaries`, `archives[0].ids`, los dos `extra_files` en `checksum` y en `release`): ninguna propiedad obsoleta | ejecutado en el worktree: `1 configuration file(s) validated`, código 0 |
| V11 | El snapshot entero —seis plataformas, el universal, el gancho con `go run`, los seis archivos, los cuatro paquetes y las huellas— tarda 18 s en este equipo (M3, con la caché de compilación caliente). `dist/artifacts.json` da el universal como tipo `Binary`, `goos` `darwin`, `goarch` `all`, y los seis archivos de hoy como `Archive`; el cask y el manifiesto de Scoop no cambian (siguen apuntando a los archivos por arquitectura). `TestSnapshot`, tal como está en `main`, pasa contra ese `dist/` | registro del snapshot; `jq` sobre `artifacts.json`; `go test -tags=snapshot -run '^TestSnapshot$' .` |
| V12 | Claude Code 2.1.284 (la versión que fija `evals.yml`): `plugin.json` es `{name, displayName?, version?, description?, author? {name, email?, url?}, homepage?, repository?, license?, keywords?, …}`; `marketplace.json` es `{name, owner {name, email?, url?}, plugins[], version?, description?, metadata?, …}`, con `name` sin espacios y sin las formas reservadas a Anthropic (`kitlegal-plugins` no lo es); cada entrada es un `plugin.json` parcial más `name`, `source` y, entre otros, `strict` (por omisión `true`: el plugin trae su propio `plugin.json`); la fuente `archive` es `{source: "archive", url, sha256?}`, con `url` HTTPS de un zip cuya raíz, o un único directorio que la envuelve, lleva `.claude-plugin/`, y `sha256` de 64 dígitos hexadecimales, que se comprueba en cada descarga; la señal de actualización es la versión (`plugin.json`, si no la de la entrada, si no la huella). `claude plugin validate <ruta>`, con un directorio, valida su `.claude-plugin/marketplace.json` o su `.claude-plugin/plugin.json` (el catálogo, si están los dos) | los textos de los esquemas y de la ayuda en el binario `~/.local/share/claude/versions/2.1.284`, alrededor de los bytes 73 695 200, 177 766 716, 177 768 300, 177 801 600, 177 810 600 y 177 814 600. La orden no se ejecutó (S3) |
| V13 | El manifiesto de MCP Bundle `0.3` exige `name`, `version`, `description`, `author` (con `name`) y `server` (`type` ∈ `python`, `node`, `binary`; `entry_point`; `mcp_config` con `command` y, opcionales, `args`, `env` y `platform_overrides`, un objeto por sistema con `command`, `args` y `env`), no admite campos fuera de su lista, y entre los que admite están `manifest_version` (`"0.3"`), `display_name`, `long_description`, `homepage`, `license`, `icon`, `tools` (`[{name, description?}]`), `compatibility.platforms` (`darwin`, `win32`, `linux`) y `user_config`. El manifiesto del prototipo lo cumple, y su icono también | `schemas/mcpb-manifest-v0.3.schema.json` del paquete `@anthropic-ai/mcpb` 2.1.2 (caché de `npx` del equipo, `~/.npm/_npx/a7f3b3c2f2763cc1/node_modules/@anthropic-ai/mcpb`); `node …/dist/cli/cli.js validate <manifest.json>` sobre el `.mcpb` extraído: `Manifest schema validation passes!`, `Icon validation passed`, código 0 |
| V14 | El prototipo del paso, sobre el snapshot: `kitlegal.mcpb` de 42 711 814 bytes con cuatro entradas (`manifest.json`, 2 939 bytes; `icon.png`, 20 539; `server/kitlegal`, 51 794 802, modo `0755`; `server/kitlegal.exe`, 26 474 496); `kitlegal-plugin.zip` de 17 739 bytes con seis (`.claude-plugin/plugin.json`, 260 bytes, y los cinco ficheros de las dos skills, 44 237). `tools` del manifiesto, compacto, 1 384 bytes; la descripción corta, 71 caracteres. El catálogo de una versión, 619 bytes. Las dos líneas nuevas de `checksums.txt`, 166 bytes | `unzip -Z`, `unzip -p … \| wc -c`, `jq` |
| V15 | Dos ejecuciones del paso sobre los mismos binarios dan los mismos dos ficheros byte a byte, e iguales a los que dejó el gancho en el snapshot: orden fijo de entradas, fecha fija, modos fijos y `compress/flate` de la biblioteca estándar. Dos snapshots distintos no dan el mismo `.mcpb`, porque cada uno compila con otra `main.fecha`; el plugin sí es el mismo | `cmp` de las tres parejas (0); `sha256sum` de los dos snapshots. Cada ejecución, con `go run` y la caché caliente, 1,2 s y 0,7 s |
| V16 | Sin una entrada, el prototipo sale con 1 y la nombra: `empaquetar: falta el binario de Windows: open …: no such file or directory`; `empaquetar: falta el icono: …`; con una carpeta de salida que no existe, `empaquetar: no se puede escribir …/kitlegal.mcpb: …` | ejecutado, tres casos |
| V17 | `debug/macho.NewFatFile`, de la biblioteca estándar, lee el universal de goreleaser: dos arquitecturas con `Cpu` (`macho.CpuAmd64`, `macho.CpuArm64`), `Offset` y `Size`; los bytes `[Offset, Offset+Size)` de cada una son los del binario `kitlegal` de `kitlegal_darwin_<arquitectura>.tar.gz`, y `server/kitlegal.exe`, el de `kitlegal_windows_amd64.zip` | `zz_h22_test.go` del prototipo, en verde |
| V18 | Con el `.mcpb` extraído en una ruta con espacios y un carácter no ASCII, el binario de la plataforma en `server/kitlegal`, la orden del `mcp_config` con `${__dirname}` resuelto, los argumentos `mcp serve` y `/` como directorio de trabajo, el cliente del SDK de `internal/mcp/mcptest` hace el saludo (protocolo 2026-07-28) y lista 10 herramientas, con los nombres y las descripciones de `tools`. **El orden de `tools/list` no es el del registro** (lo pone el SDK): hay que comparar conjuntos, no posiciones | el mismo test: en su primera versión, que comparaba por posición, falló en `graph_check` frente a `graph_stats` |
| V19 | `kitlegal skills install`, en un directorio vacío y con un `HOME` vacío, deja `.agents/skills/<skill>/…` y el manifiesto `.agents/skills/kitlegal.json`; cada carpeta de skill es idéntica a la de `skills/` del plugin | `diff -r` (0) y el mismo test |
| V20 | Los pasos nuevos de `humo`, con órdenes de shell y sin el código del repositorio, pasan contra el `dist/` del prototipo en 3,2 s: las huellas de los cinco ficheros, el manifiesto con `jq`, las dos arquitecturas con `od`, `head`, `tail`, `tar` y `sha256sum`, y un cliente MCP hecho con una tubería con nombre que arranca el servidor, saluda (protocolo 2025-06-18) y lista las 10 herramientas. Con otra etiqueta falla en el manifiesto. Dos trampas medidas: el `od` de macOS añade una línea en blanco (`awk 'NF == 4'`), y `tail -c +N … \| head -c M` hace fallar el paso con 141 bajo `set -o pipefail`, porque `head` cierra la tubería (`head -c … \| tail -c …` no) | `h22-pasos-humo.sh`, con el bash 3.2 y las órdenes de macOS (S1) |
| V21 | `golangci-lint`, con la configuración del repositorio, sobre el prototipo: ningún hallazgo de `depguard` (el test de la raíz importa `internal/mcp/mcptest` y el paso importa `internal/app` y el paquete raíz), de `forbidigo` (`os.Exit` y `os.Stderr` en `cmd/empaquetar`, que cubre la excepción `^cmd/`) ni de G204 (la orden es un parámetro y los argumentos, constantes). Sí marcó: `Autor`/`autor` (`misspell` lo lee como `Author`; `Autoria`, `Creador`, `Editor`, `Titular`, `Manifiesto`, `Compatibilidad` y `Sobrescritura` no); `os.WriteFile` con `0644` y `os.MkdirAll` con `0755` (G306, G301; con `os.OpenRoot` y `Root.WriteFile` no hay hallazgo); y, del prototipo y sin interés para el diseño, una función de complejidad 16, un `fmt.Fprintf` sin comprobar y el formato | `go tool -modfile=tools/golangci-lint/go.mod golangci-lint run . ./internal/empaquetado/... ./cmd/empaquetar/... ./internal/app/`, dos veces |
| V22 | `gh` 2.101.0: `gh api` tiene `--method`, `--input` (con `-` lee el cuerpo de la entrada estándar), `--jq` y `-f`; `gh release download`, `--pattern` (repetible) y `--dir`; `gh attestation verify`, `--repo`; y toda orden que falla sale con 1 | `gh api --help`, `gh release download --help`, `gh attestation verify --help`, `gh help exit-codes`, con `GH_CONFIG_DIR` en un directorio vacío |
| V23 | `release_test.go`, del paquete de tests de la raíz, ya declara un tipo `empaquetado`: importar ahí `internal/empaquetado` sin otro nombre no compila | `go test` del prototipo: `empaquetado already declared through import of package` |
| V24 | Los guiones `instalador-` pasan contra el `dist/` nuevo: las dos líneas de más en `checksums.txt` no les afectan. `TestConfiguracionDeLaRelease` falla con el `.goreleaser.yaml` nuevo hasta que se le enseñan las secciones nuevas, porque lo lee de forma estricta | `KITLEGAL_DIST=… go test -run '^TestEntregaDelHito$/instalador-' ./internal/app/` (verde, 10 s); `go test -run '^TestConfiguracionDeLaRelease$' .` (rojo) |
| V25 | El informe final da un control `ci:<ruta>:<Test>` por comprobado si la ruta está en la cabeza, lleva `func <Test>(` o el objetivo `<Test>:` y `make ci` está en verde. No ejecuta nada | `scripts/workflow/informe.sh:250-258` |
| V26 | `evals.yml` fija la versión de Claude Code en `jobs.evals.env.VERSION_DE_CLAUDE_CODE` (`2.1.284`) y la instala con `npm install -g "@anthropic-ai/claude-code@${VERSION_DE_CLAUDE_CODE}"` | `.github/workflows/evals.yml:179, 210` |
| V27 | El servidor anuncia de cada herramienta el nombre `<applet>_<verbo>` y `Verbo.Descripcion`, los dos tomados de `verbosAnunciados`; `app.NombresDeHerramientas` ya da los nombres desde ahí, y `TestHerramientasDelServidor` los compara con lo anunciado. `app.RegistroDeProduccion` no abre ni pide nada al construirse | `internal/app/herramientas.go:46-90, 131-150`; `internal/app/registro.go` (comentario de `RegistroDeProduccion`) |
| V28 | El binario distribuido es el cierre de `./cmd/kitlegal`: `TestDependenciasDelBinario` y `TestElBinarioNoEnlazaLosEjemplos` lo miden con `go list -deps`. `paquetesInternos` (R1) son hoy diez | `internal/arch_test.go:260-320, 869-871` |
| V29 | `mcp/icon.png` es un PNG de 512 × 512 px, de 20 539 bytes | `file mcp/icon.png`; `wc -c` |
| V30 | El cuerpo del paso que publica el catálogo ([contracts/release.md §6.3](./contracts/release.md)) es sintaxis válida de bash y su lógica hace lo que dice: sin fichero anterior envía `{message, content}`, y con él, además, su `sha`; `content` es el catálogo en base64 sin saltos de línea (828 caracteres para 619 bytes) y se decodifica al mismo documento | `bash -n` y dos ejecuciones con un `gh` de mentira que da o no da un `sha` y enseña el cuerpo del `PUT`. El `gh` de verdad no se ejecutó (S2) |
| V31 | Los escenarios §2 a §5 de [quickstart.md](./quickstart.md) dan lo que dicen, y no dejan nada en el árbol fuera de `dist/` | ejecutados en el worktree del prototipo (`h22-quickstart.sh`); `git status --short` solo enseñó los ficheros del propio prototipo |
| V32 | El servidor arranca igual cuando `server/kitlegal` es un enlace simbólico al binario: desde una ruta con espacios y un carácter no ASCII y con `/` de directorio de trabajo, lista 10 herramientas. El repositorio ya documenta por qué no conviene que un test escriba un ejecutable y lo lance: en Linux, si otro test lanza un proceso mientras el fichero está abierto para escribir, la ejecución falla con «text file busy» (golang/go#22315) | un `kitlegal` compilado de la cabeza de la rama, enlazado y arrancado con el cliente de shell de V20; `internal/evals/sustitutos_test.go:225-236`, `internal/app/tuberia_unix_test.go:122-140` |

## S · Supuestos que no se pudieron comprobar

| # | Supuesto | Por qué no se comprobó | Qué lo cubre |
|---|---|---|---|
| S1 | Los pasos nuevos de `humo` funcionan igual en `ubuntu-latest`: `od -An -tu1 -j -N`, `head -c`, `tail -c`, `mkfifo`, `tar -xzOf`, `unzip -p`, `jq -e` y `sha256sum --check --strict` son de POSIX o ya los usa `humo` | Medidos en macOS (bash 3.2, órdenes de BSD); no hay un Linux en la sesión | `TestConfiguracionDeLaRelease` fija sus líneas y `bash -n` los lee. Su primera ejecución real es la primera release (SC-002) |
| S2 | `gh api --method PUT repos/…/contents/<ruta>` crea el fichero sin `sha` y lo sustituye con el `sha` del que hay; el `GET` de una ruta que no existe sale con 1 | Sin red y sin la credencial de `gh` | El paso falla y su trabajo sale en rojo si no escribe (FR-031); se ve en la primera release |
| S3 | `claude plugin validate` se ejecuta en el runner sin credencial ni sesión con modelo, no descarga la dirección del catálogo y da por válidos el plugin y el catálogo del plan | Ninguna sesión del run ejecuta `claude` (ADR 0032). Los esquemas sí se leyeron (V12) | El trabajo `snapshot` de la propuesta de cambio (SC-008): si falla, el cierre lo cuenta |
| S4 | `npm install -g @anthropic-ai/claude-code@2.1.284` funciona en `ubuntu-latest`, como en el `ubuntu-24.04` del job de evals | Sin red | El mismo trabajo |
| S5 | `actions/attest-build-provenance@v4` atesta dos ficheros más en `subject-path`, y `gh attestation verify` los verifica como al archivo | Solo se ejecuta al etiquetar | `humo` (FR-040) |
| S6 | Al publicar, goreleaser sube los dos ficheros de `release.extra_files` (V9, leído y no ejecutado) | Publicar es de la release | `humo`, que los descarga |
| S7 | La app de escritorio de Claude acepta el `.mcpb` del paso: conserva el bit de ejecución de `server/kitlegal` al extraerlo y aplica `platform_overrides.win32` en Windows | Es la prueba humana | SC-002; `mcpb validate` da el manifiesto por válido (V13) |
| S8 | La app admite un catálogo con una entrada de fuente `archive` | Pendiente 5 del ADR 0035 | SC-002, que lo anota. Subir el zip no depende de ello |
| S9 | `PUBLISHER_TOKEN` puede escribir en `jmorenobl/kitlegal-plugins` | Es de la persona, antes de la primera release (FR-032) | El trabajo `catalogo` sale en rojo si no puede |
| S10 | En `publicar`, sin caché de compilación, el `go run` del gancho añade a la release el tiempo de compilar el paso para el runner | No se puede ejecutar el flujo | Sin umbral: el spec no pide tiempo |
| S11 | En el trabajo `catalogo`, `actions/checkout` sin `ref` obtiene el commit de la etiqueta que dispara el flujo, como hoy en `publicar`, y `ubuntu-latest` trae `gh`, `jq`, `unzip`, `base64` y `npm` | Los flujos no se ejecutan en la sesión | Los mismos flujos de hoy ya dependen de ello (`publicar`, `humo`, el job de evals) |

## D · Decisiones

### D1 · El paso vive en `cmd/empaquetar` y en `internal/empaquetado`

**Decisión**: un `main` mínimo, `cmd/empaquetar/main.go`, que inyecta los descriptores y termina con el código que
devuelve `empaquetado.Ejecutar`, como `cmd/kitlegal` con `app.Arrancar`; la lógica, en `internal/empaquetado`.
goreleaser solo construye `./cmd/kitlegal`, así que el paso no viaja en ningún archivo.

**Por qué**: es la forma del repositorio (un `main` que no decide nada) y la de la skill `golang-project-layout`
(«All `main` packages must reside in `cmd/` with minimal logic»). Los tests de la raíz necesitan importar los textos
(D4), y de un paquete `main` no se puede.

**Alternativas**: *un módulo en `tools/`*: no puede importar `internal/` del módulo raíz, y el paso necesita el
registro y lo empotrado. *Un guion de shell con `zip` y `jq`*: el hito pide Go, y `tools` saldría de una segunda lista.
*Un applet o una bandera del binario*: lo prohíbe FR-001. *Todo en `package main`*: los tests del snapshot no podrían
comparar con los textos sin escribirlos otra vez.

### D2 · `tools` sale del registro, en el proceso del paso

**Decisión**: el paso llama a `app.RegistroDeProduccion(version)` y a una función nueva y exportada de `internal/app`,
`HerramientasAnunciadas(registro)`, que da el nombre y la descripción de cada verbo de `verbosAnunciados`: la misma
fuente de la que `herramientasDe` saca lo que el servidor anuncia (V27). El paso se compila del mismo árbol que los
binarios que empaqueta.

**Por qué**: FR-014 pide que salga del registro «como las herramientas del servidor». Que lo que da la función es lo que
el servidor anuncia por MCP lo comprueba `TestHerramientasAnunciadas` en `make ci`, con el servidor en proceso; y que el
manifiesto del snapshot coincide con lo que anuncia su binario, `make snapshot-check`.

**Alternativa**: *arrancar el binario del runner y preguntarle por MCP*: el paso necesitaría un cliente MCP (el SDK
solo se importa desde `internal/mcp`, R7), dependería de que el snapshot construya la plataforma del runner y tendría
más formas de fallar, para llegar al mismo conjunto.

### D3 · Las skills del plugin son lo empotrado

**Decisión**: el paso recorre `kitlegal.Skills()` (`skills.go`) y escribe cada fichero con su ruta, `skills/<skill>/…`.

**Por qué**: FR-020 pide «cada skill empotrada en el binario de esa construcción», y lo empotrado es exactamente eso:
la misma directiva `//go:embed`, compilada del mismo árbol. Una skill nueva llega sin tocar el paso (US5). Un directorio
de `skills/` con `references/` y sin `SKILL.md`, que `skills install` ignora, no existe en el árbol:
`TestSkillsEmpotradas` falla con él en `make ci`.

**Alternativa**: *leer `skills/` del disco*: habría que repetir qué entra y qué no, que es lo que la directiva ya dice.

### D4 · Los textos son constantes de `internal/empaquetado/textos.go`

**Decisión**: `NombreVisible`, `Descripcion`, `DescripcionLarga` y `Autoria`, exportadas, en un solo fichero. De ahí
los leen el manifiesto, `plugin.json` y el catálogo. Sus valores:

- `NombreVisible`: `kitlegal`.
- `Descripcion`: `Tu asistente de IA responde con la ley vigente del BOE y la cita exacta` (71 caracteres): la que
  goreleaser ya da al cask, al bucket y a los paquetes.
- `DescripcionLarga`: un párrafo con qué hace, que todo corre en el equipo y que hace falta también el plugin; su texto,
  en [data-model.md §5](./data-model.md).
- `Autoria`: `kitlegal`, sin correo ni nombre de persona: el buzón del proyecto ya está en la web y en el README.

**Por qué**: FR-015 pide un solo sitio. Constantes compiladas en el paso no añaden una entrada que pueda faltar, y un
test de `make ci` mide su longitud. El nombre `Autoria` y no `Autor` es por `misspell` (V21).

**Alternativas**: *un YAML en `data/`*: `data/` es la fuente de verdad de las skills y parte de ella viaja en el
binario; los textos de la ficha no son de ninguna skill. *Ficheros JSON de plantilla en `mcp/` y `plugin/`*: los textos
quedarían escritos en más de uno, o las plantillas a medio rellenar por código.

Los valores son un supuesto del plan, de impacto `comportamiento` (es lo que la persona ve en la ficha).

### D5 · La plantilla del catálogo es código, con tres parámetros

**Decisión**: `internal/empaquetado/catalogo.go` es el fichero del que sale el catálogo de cada etiqueta (spec, «Key
Entities»): compone `marketplace.json` con los textos de D4 y tres valores, la versión, la dirección —que deriva de la
versión— y la huella. Lo escribe la orden `empaquetar catalogo -version <v> -sha256 <huella> -salida <fichero>`. La
usan el trabajo `catalogo` de `release.yml`, con la huella de `checksums.txt` de la release, y `make plugin-check`, con la
versión y la huella del snapshot (los «valores de prueba» del spec).

**Por qué**: los textos tienen que salir del sitio único de FR-015, y un fichero JSON con huecos los escribiría otra vez
o pediría un segundo mecanismo que los rellene. Así hay una sola forma de componer el catálogo, probada en `make ci`
(`TestCatalogo`) y validada con `claude plugin validate` con la misma función.

**Alternativa**: *un `marketplace.json` versionado con marcadores y `jq` o `sed` en `release.yml`*: los textos por
duplicado, o un test más para que no diverjan; y la lógica de rellenar, en shell, repetida en dos flujos y sin test.

### D6 · El paso lo ejecuta el gancho `post` de `universal_binaries`

**Decisión**:

```yaml
universal_binaries:
  - id: kitlegal-universal
    ids: [kitlegal]
    replace: false
    hooks:
      post:
        - cmd: go run ./cmd/empaquetar piezas -version {{ .Version }} -macos {{ .Path }} -windows dist/kitlegal_windows_amd64_v1/kitlegal.exe -icono mcp/icon.png -salida dist
          output: true
```

**Por qué**: es el único sitio en el que goreleaser tiene ya todos los binarios y el universal, y todavía no ha
calculado las huellas (V1, V2, V8): lo que el paso deja en `dist/` entra en `checksums.txt`, bajo su firma (FR-002).
`.Version` es la versión de FR-013 en la release y en el snapshot (V5). Si el paso falla, falla goreleaser (V4; FR-005).

La ruta del binario de Windows es literal, con el sufijo `_v1` que pone goreleaser (V7). Si una versión de goreleaser
lo cambiara, el paso fallaría nombrando el fichero y el trabajo `snapshot` saldría en rojo en la propuesta que la
sube: falla a la vista y antes de ninguna release.

**Alternativas**: *`before.hooks`*: se ejecuta antes de compilar. *El gancho `post` de `builds`*: corre una vez por
plataforma y antes de que exista el universal. *Una orden de `make release` detrás de goreleaser*: los ficheros no
entrarían en `checksums.txt` ni irían firmados, y `release.yml`, que llama a goreleaser, no la ejecutaría.

### D7 · El universal lo hace goreleaser, con un `id` propio, y `archives` se acota

**Decisión** (Clarifications, pregunta 1): `builds[0].id: kitlegal`, escrito; `universal_binaries[0]` con
`id: kitlegal-universal`, `ids: [kitlegal]` y `replace: false`; `archives[0].ids: [kitlegal]`.

**Por qué**: con eso salen los seis archivos de hoy y ninguno más (V6), el universal no se publica (V8) y
`TestSnapshot` sigue en verde (V11). `builds[0].id` se escribe, aunque es el valor por omisión (V7), porque dos `ids`
lo nombran.

### D8 · Los dos zips son reproducibles por construcción

**Decisión**: entradas en un orden fijo; ninguna entrada de directorio; método Deflate de `archive/zip`; fecha de
modificación fija, 1980-01-01T00:00:00Z; modo `0755` en los dos binarios y `0644` en lo demás; ningún comentario. Los
tres documentos JSON salen de tipos con sus campos en un orden fijo, con dos espacios de sangría, sin escapar `<`, `>`
ni `&`, y con salto de línea final.

**Por qué**: FR-004. Medido: V15.

**Alternativas**: *sin comprimir*: 78 MB en lugar de 43. *La fecha del commit*: una entrada más que no hace falta.

### D9 · Las comprobaciones del snapshot son subpruebas de `TestSnapshot`

**Decisión**: seis subpruebas nuevas en la tabla de `TestSnapshot`, en un fichero nuevo de la raíz con la etiqueta
`snapshot`: la receta de `make snapshot-check` no cambia. Leen los zips con `archive/zip`, el universal con
`debug/macho` (V17), hablan con el servidor con `internal/mcp/mcptest` (V18) y comparan las skills con lo que instala el
binario del snapshot (V19).

**Alternativa**: *guiones `testscript`, como los `instalador-`*: leer un zip, un universal y hablar MCP pediría órdenes
nuevas del arnés o `unzip` y `jq` del sistema; en Go son la biblioteca estándar y un cliente que ya existe.

### D10 · `claude plugin validate` va en un objetivo aparte, `make plugin-check`

**Decisión**: un test con la etiqueta `snapshot`, `TestPluginValido`, extrae el plugin del snapshot en un directorio
temporal, compone el catálogo de la versión del snapshot con la huella de `checksums.txt` (D5) y ejecuta
`claude plugin validate .` en cada uno; falla si la orden no está o sale con un código distinto de 0. Lo ejecuta
`make plugin-check`, fuera de `make ci` y de `make snapshot-check`. El trabajo `snapshot` de la CI instala Claude Code en
la versión de `evals.yml` y lo llama.

**Por qué**: FR-065. Aparte de `snapshot-check`, porque las tareas del run ejecutan `make snapshot-check` y ninguna
sesión del run ejecuta `claude`. Como objetivo del `Makefile` y no como pasos de shell del flujo, porque lo que la CI
ejecuta es lo que se ejecuta en local (`ci.yml`, cabecera).

### D11 · Dos clientes MCP: el del SDK en el snapshot y uno de shell en `humo`

**Decisión**: `make snapshot-check` usa `mcptest.Abrir` (el cliente del SDK) sobre un proceso lanzado con la orden del
manifiesto. `humo`, que no tiene el código (FR-043), usa un cliente escrito en su paso: una tubería con nombre de la
que lee las respuestas, tres mensajes (`initialize`, `notifications/initialized`, `tools/list`) y `jq` (V20).

**Por qué**: el cliente de shell espera cada respuesta antes de enviar lo siguiente y antes de cerrar la entrada, que es
lo que el servidor necesita para escribirla (H21, research V12), sin esperas de tiempo ni procesos en segundo plano.

**Alternativa**: *descargar un cliente en `humo`*: más superficie y más red para listar unos nombres.

### D12 · El trabajo `catalogo` compone con el paso y publica con la API de contenidos

**Decisión**: un tercer trabajo, `catalogo`, con `needs: [humo]` y `permissions: contents: read`: obtiene el código de
la etiqueta sin dejar la credencial en el clon, instala Go, descarga `checksums.txt` de la release, compone el catálogo
con `go run ./cmd/empaquetar catalogo` y lo escribe en `jmorenobl/kitlegal-plugins` con `gh api --method PUT` sobre
`.claude-plugin/marketplace.json`, con `PUBLISHER_TOKEN` solo en el entorno de ese paso.

**Por qué**: Clarifications, pregunta 3. Una llamada a la API escribe el fichero y deja un commit, sin clonar ni
configurar una identidad ni una credencial de git. Volver a ejecutar el trabajo escribe lo mismo.

**Alternativas**: *`git clone`, commit y `git push`*: la credencial en la dirección o en la configuración del clon.
*Rellenar con `jq`, sin Go*: es la alternativa de D5. *Pasar el catálogo de `publicar` a `catalogo` como artefacto del
flujo*: dos acciones más, y la huella dejaría de salir de `checksums.txt` publicado, que es lo que pide FR-031.

### D13 · Aceptación e2e: no aplica

**Decisión**: el plan dice «Aceptación e2e: no aplica».

**Por qué**: el hito no añade ni cambia ningún comportamiento del binario: FR-001 y FR-080 lo prohíben (los applets,
los verbos, `--describe`, las herramientas y las skills son los de hoy). Lo que entrega son dos ficheros de la release,
dos flujos y documentación. Su aceptación en el run es el trabajo `snapshot` de la propuesta de cambio (SC-001), y las
evals no cambian. Que el binario sigue igual lo fijan los guiones y los tests que ya hay.

**Alternativa**: *guiones `testscript` que ejecuten el paso*: el arnés e2e construye y ejerce `kitlegal`, no un
programa de construcción; el paso se prueba en proceso, en su paquete.

### D14 · El paso sale con 0 o con 1

**Decisión**: 0 si escribe lo pedido; 1 con cualquier fallo, también con unos argumentos que no valen, y una línea
`empaquetar: …` en la salida de error.

**Por qué**: FR-005 pide «un código distinto de 0», y quien lo consume es goreleaser o un paso de un flujo, que solo
distinguen 0 de lo demás. La tabla del ADR 0023 es del contrato de resultados del binario, que no cambia.

### D15 · El orden de `tools` es el del registro, y los controles comparan conjuntos

**Decisión**: `tools` va en el orden de `HerramientasAnunciadas` (applets por nombre, verbos en el orden de su
catálogo). `make snapshot-check` y `humo` comparan por nombre, no por posición.

**Por qué**: FR-014 habla del «conjunto», y el orden de `tools/list` lo pone el SDK (V18).

### D16 · Los ficheros de salida se escriben con `os.Root`

**Decisión**: `kitlegal.mcpb` y `kitlegal-plugin.zip` se escriben con `os.OpenRoot(<salida>)` y `Root.WriteFile`, con
modo `0644`, como el resto de `dist/`; el catálogo, igual, en la carpeta de su fichero. La carpeta tiene que existir. El
paso no escribe nada en la salida estándar: el catálogo va a un fichero, `-salida`, y no a una tubería.

**Por qué**: `os.WriteFile` con `0644` es un hallazgo de `gosec` (V21), y no se prevé ningún `//nolint`. Sin salida
estándar, la regla R5 no tiene nada que decir del paso: `cmd/empaquetar/main.go` solo nombra `os.Stderr`.

**Alternativa**: *el catálogo por la salida estándar*, con una bandera menos: obligaría a justificar que un paquete que
no es `internal/render` escriba en ella.

### D17 · Un control de que el binario distribuido no enlaza el paso

**Decisión**: `TestElBinarioNoEnlazaElPaso`, en `internal/arch_test.go`, sobre el cierre de `./cmd/kitlegal`, como
`TestElBinarioNoEnlazaLosEjemplos`. `internal/empaquetado` entra en `paquetesInternos` y en la lista `core` de
`depguard` (R1), como `disco` en H19 y `mcp` en H21.

**Por qué**: FR-001 («MUST NOT … entrar en el binario distribuido») y FR-069.

### D18 · Datos externos: ninguno

**Decisión**: ninguna fuente, ninguna grabación y ningún fichero nuevo en `testdata/`, `schemas/` ni `data/`. El esquema
oficial del manifiesto no se versiona (Clarifications, pregunta 2): se leyó en local para verificar la forma (V13) y no
se copia. Los tests usan binarios de prueba y un icono que crean en `t.TempDir()`, y `mcp/icon.png`, que ya está
versionado.

### D19 · En los tests de la raíz, el paquete se importa como `paso`

**Decisión**: `import paso "github.com/jmorenobl/kitlegal/internal/empaquetado"`.

**Por qué**: V23. Renombrar el tipo de `release_test.go` es tocar más de lo que el hito pide.

### D20 · Los umbrales de la ficha los comprueba el propio paso, y un test lo ve fallar

**Decisión**: el paso rechaza una descripción corta de más de 120 caracteres y un icono que no sea un PNG de
512 × 512 px (`image/png`, `DecodeConfig`), con su línea de error y código 1. `TestDescripcionCorta` lo ve con 120 y con
121 y mide la del repositorio; `TestIcono` lo ve con un PNG de otro tamaño y con algo que no es un PNG, y mide
`mcp/icon.png`.

**Por qué**: SC-009 pide un test que falle «con 121 caracteres o con un icono de otro tamaño». El paso lee el icono de
todos modos: mirarle la cabecera no añade ninguna entrada.
