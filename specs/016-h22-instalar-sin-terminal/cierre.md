# H22 · Cierre de la Definition of Done

Cierre de T007 (FR-053, FR-068, FR-069, FR-070, FR-080; SC-001, SC-011) sobre `94fce58`, que es T006, el último commit
antes de T007. La rama está al día con `main`: `git merge-base main HEAD` da la cabeza de `main`, `751b76e`, así que
`git diff --name-status main` da exactamente lo que cambia el hito. Las seis partes son las de la tarea: lo creado y lo
modificado, la cobertura, el quickstart, los controles de umbral, los supuestos que siguen sin medir y lo que queda para
la persona.

Todo lo que se cita aquí se ejecutó en la sesión de T007, en primer plano, y se vio terminar, o está en el repositorio.
Lo que no se pudo medir se dice tal cual (§3, escenario 10; §4, «Las cinco medidas»; §5). **Nada de (1) a (4) deja de
cuadrar**: las diferencias que hay son de cómo se ejecutó el quickstart en esta sesión (§3, «Cómo se ejecutó») y de dos
cifras del plan que no son umbrales (§3, «Dos cifras del plan»). Esta tarea no toca código: su único fichero es este.

## 1. Lo creado y lo modificado en el hito

`git diff --name-status main`, antes de escribir este fichero, da **55 ficheros: 42 `A` y 13 `M`**, ninguna `D` ni `R`
(7 510 líneas añadidas y 192 quitadas). 30 están en `specs/016-h22-instalar-sin-terminal/`, todos `A`; con este
`cierre.md` serán 31. Los otros 25 (4 383 líneas añadidas y 192 quitadas), con la tarea que los tocó
(`git log --format=%s main..HEAD -- <fichero>`) y sus líneas (`git diff --numstat main`):

| Estado | Fichero | Tarea | Líneas |
|---|---|---|---|
| M | `internal/app/herramientas.go` | T001 | +28 −0 |
| M | `internal/app/herramientas_test.go` | T001 | +125 −0 |
| A | `cmd/empaquetar/main.go` | T002 | +21 |
| A | `internal/empaquetado/doc.go` | T002 | +30 |
| A | `internal/empaquetado/textos.go` | T002 | +52 |
| A | `internal/empaquetado/textos_test.go` | T002 | +71 |
| A | `internal/empaquetado/piezas.go` | T002 | +458 |
| A | `internal/empaquetado/piezas_test.go` | T002 | +765 |
| A | `internal/empaquetado/catalogo.go` | T002 | +114 |
| A | `internal/empaquetado/catalogo_test.go` | T002 | +163 |
| A | `internal/empaquetado/ejecutar.go` | T002 | +179 |
| A | `internal/empaquetado/ejecutar_test.go` | T002 | +397 |
| A | `internal/empaquetado/export_test.go` | T002 | +16 |
| M | `internal/arch_test.go` | T002 | +57 −8 |
| M | `.golangci.yml` | T002 | +7 −1 |
| A | `snapshot_piezas_test.go` | T003, T004 | +873 |
| M | `snapshot_test.go` | T003 | +11 −0 |
| M | `.goreleaser.yaml` | T003 | +42 −6 |
| M | `release_test.go` | T003, T004, T005 | +530 −101 |
| M | `Makefile` | T004 | +15 −2 |
| M | `.github/workflows/ci.yml` | T004 | +18 −1 |
| M | `.github/workflows/release.yml` | T005 | +180 −4 |
| M | `README.md` | T006 | +105 −42 |
| M | `CONTRIBUTING.md` | T006 | +106 −27 |
| M | `CHANGELOG.md` | T006 | +20 −0 |

Los 30 del feature son `spec.md`, `plan.md`, `research.md`, `data-model.md`, `quickstart.md`, `tasks.md`, los tres de
`contracts/`, `checklists/requirements.md` y 20 de `gates/`. Cada tarea tocó solo las rutas que declara en `tasks.md`.

### Lo que no está en la lista (FR-053, FR-080)

Cada comprobación es una orden que cuenta líneas, ejecutada en esta sesión, y todas dan **0**:

- **Ningún fichero de `testdata`, `schemas`, `data`, `skills`, `evals`, `web`, `docs` ni `evidencias`**, a cualquier
  profundidad: `git diff --name-status main -- testdata '*/testdata/*' schemas data skills evals web docs evidencias`
  no imprime nada, y `git diff --name-only main` filtrado por
  `(^|/)(testdata|schemas|data|skills|evals|web|docs|evidencias)/` tampoco. Las dos formas se vieron funcionar antes de
  fiarse del cero: la expresión casa las cinco rutas de prueba que debe casar, y el mismo `git diff` con `-- internal`
  da 13.
- **Ningún spec anterior**: de `git diff --name-only main -- specs`, ninguna línea queda fuera de
  `specs/016-h22-instalar-sin-terminal/`.
- **Nada del workflow `hito` ni de sus guiones**: `git diff --name-status main -- .specify scripts .claude .agents
  skills-lock.json` no imprime nada.
- **Ni ADR, ni `SOURCES.md`, ni dependencias, ni lo empotrado**: `git diff --name-status main -- docs/ADR
  docs/SOURCES.md go.mod go.sum tools scripts/verify-sources.sh mcp plugin skills.go` no imprime nada. `mcp/icon.png`
  es el de `main`.

Es decir: ni fuente nueva, ni fila nueva en `docs/SOURCES.md`, ni esquema, ni fixture, ni grabación, ni eval, ni skill,
ni ADR, ni dependencia. El hito no cambia ninguna decisión de arquitectura que no cambiara ya el ADR 0035, y no hay
ningún ADR nuevo que lo diga de otra manera. `git ls-files --others --exclude-standard` no da ningún fichero sin
seguimiento. En las 4 383 líneas añadidas fuera de `specs/` no hay ningún `nolint`, `t.Skip`, `TODO`, `FIXME` ni `XXX`
(0 coincidencias).

### La Definition of Done, punto a punto (`ROADMAP.md` §1)

| Punto | Aplica | Qué se ve en la cabeza |
|---|---|---|
| 1 · `make ci` en verde | sí | §3, escenario 8: `ci: todos los controles en verde`. |
| 2 · tests offline; fixtures si toca red | tests sí; fixtures no | Los siete tests del paso, `TestHerramientasAnunciadas` y `TestElBinarioNoEnlazaElPaso` (§4). El hito no toca la red: ningún fixture ni grabación (arriba). |
| 3 · sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | sí | En `cmd/empaquetar` e `internal/empaquetado`, la única coincidencia de `net/http`, `os.Exit`, `fmt.Print`, `panic(`, `database/sql` o `modernc` es el `os.Exit` de `cmd/empaquetar/main.go:20`, que R4 admite en `cmd/`. Los imports del paso son biblioteca estándar, la raíz del módulo e `internal/app` (`go list`). `golangci-lint`: `0 issues.` |
| 4 · esquemas y `schema-check` | sin esquema nuevo | `make schema-check` en verde y `schemas/` sin cambios (§3, escenario 7). |
| 5 · errores con código estable, sin `panic` | el paso sale con 0 o con 1 | `TestEjecutar`, 21 subpruebas; §3, escenario 3: código 1 y su línea `empaquetar: …`. |
| 6 · e2e y `CHANGELOG.md` | e2e no aplica; `CHANGELOG.md` sí | `CHANGELOG.md` +20 (T006). Ningún guion `testscript` cambia. |
| 7 · ADR | no | `docs/ADR` sin cambios. |
| 8 · `SOURCES.md` | no | `docs/SOURCES.md` y `scripts/verify-sources.sh` sin cambios. |
| 9 · cobertura | sí | §2. |
| 10 · evals y skills | no | `skills/`, `evals/` y `data/` sin cambios; `make skills-check` en verde (§3, escenario 7). |
| 11, 12, 13 · territorio, grafo, plazos | no | — |

## 2. Cobertura (Definition of Done §1.9)

Se mide sobre los dos perfiles que deja el `make ci` de §3, escenario 8, en verde y con la caché de tests vacía
(`go clean -testcache` justo antes): `coverage.out` (`make test`) y `coverage-integration.out` (`make
test-integration`). El global sale de `go tool cover -func` sobre cada perfil; el de cada árbol, de `go tool cover
-func` sobre el perfil filtrado a sus líneas, con su cabecera `mode:`, escrito en el directorio temporal. Los recuentos
de sentencias salen de los bloques del perfil, contando cada bloque una vez, y coinciden con el porcentaje de `go tool
cover`. «Unión» son los dos perfiles fundidos bloque a bloque, como hace Codecov con los dos que publica la CI.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | **97,0 %** (12 687 de 13 084 sentencias) | **97,4 %** (12 749 de 13 084) | 97,4 % | ≥ 70 % |
| `internal/core/**` (dominio) | **98,6 %** (2 545 de 2 581) | **98,6 %** (2 545 de 2 581) | 98,6 % | ≥ 85 % |
| `internal/cli/**` | 98,5 % (528 de 536) | 98,5 % | 98,5 % | ≥ 90 % |
| `internal/empaquetado` (el paquete del paso) | **98,3 %** (173 de 176) | **98,3 %** (173 de 176) | 98,3 % | — |
| `cmd/empaquetar` (su `main`) | 0,0 % (0 de 1) | 0,0 % | 0,0 % | — |
| `internal/app/herramientas.go` | 100 % (117 de 117) | 100 % | 100 % | — |

Los tres umbrales de proyecto se cumplen, y no hace falta ningún test. El hito no toca el dominio: `internal/core` no
está en la lista de §1 y da las mismas 2 545 de 2 581 sentencias que el cierre de H21.

**El paquete del paso, función a función.** De sus 20 funciones, 17 están al 100 % y tres no: `ejecutarPiezas` (94,4 %),
`pluginDe` (94,4 %) y `comprimir` (92,9 %). Las tres sentencias sin cubrir son ramas de error sin vía con las entradas
de producción:

- `internal/empaquetado/ejecutar.go:119`: el registro de applets de producción que no se construye. Lo impide
  `TestRegistroDeProduccion`, y el comentario del fichero lo dice.
- `internal/empaquetado/piezas.go:336`: `fs.ReadFile` que falla sobre un fichero que `fs.WalkDir` acaba de dar, en el
  árbol de skills empotrado.
- `internal/empaquetado/piezas.go:407`: cerrar un `zip.Writer` que escribe en memoria.

La cuarta sentencia sin cubrir del hito es la de `cmd/empaquetar/main.go:20`, el `os.Exit(empaquetado.Ejecutar(…))`:
ningún test ejecuta un `main`, igual que el de `cmd/kitlegal` (0,0 % en los dos perfiles). Lo ejecutan de verdad §3,
escenarios 2 a 5.

**El diff, como estimación.** `codecov/patch` es bloqueante con `target: auto`, la cobertura de la base. `main` no se
midió en esta sesión. Lo que sí se midió: todo lo nuevo de `internal/app/herramientas.go` está cubierto (el fichero
entero, 117 de 117) y las sentencias nuevas sin cubrir son las cuatro de arriba, de las 177 de `internal/empaquetado` y
`cmd/empaquetar`. El cierre de H21 da 12 901 sentencias en su cabeza y 111 en `herramientas.go`; con esas dos cifras,
que están en el repositorio y no son una medida de hoy, el hito añade 183 sentencias y cubre 179 (97,8 %), por encima
del 97,4 % de la unión. Lo que diga `codecov/patch` solo se sabe en la propuesta de cambio.

## 3. Quickstart (§1 a §9)

Los nueve escenarios se ejecutaron en orden y en primer plano, con sus órdenes tal cual, sobre `94fce58` y con el árbol
limpio salvo los dos ficheros de `gates/` que el workflow lleva modificados (`tarea-actual.json` y
`tareas-intentos.json`).

| § | Órdenes | Esperado | Obtenido | ¿Cuadra? |
|---|---|---|---|---|
| 1 | cuatro `go test -count=1` | las cuatro, `ok` | `ok … internal/empaquetado 0.438s`, `ok … internal/app 2.108s`, `ok … internal 0.637s`, `ok … kitlegal 0.357s` | sí |
| 2 | el paso dos veces sobre los mismos binarios de prueba; `cmp`; `unzip -Z1`; `jq` | `iguales byte a byte`; cuatro entradas de la extensión; seis del plugin; `0.0.0-prueba`, `10`, `71` | las dos ejecuciones, código 0 y sin salida; `iguales byte a byte`; `manifest.json`, `icon.png`, `server/kitlegal`, `server/kitlegal.exe`; `.claude-plugin/plugin.json` y los cinco ficheros de `skills/` (dos de `boe-legislacion`, tres de `legal-core`); `0.0.0-prueba`, `10`, `71` | sí |
| 3 | el paso sin el binario de Windows | `empaquetar: falta el binario de Windows: open …/no-existe.exe: no such file or directory`, `exit status 1`, `código: 1` | esas tres líneas, con la ruta del temporal | sí |
| 4 | la huella del plugin; `empaquetar catalogo`; `jq`; `rm -rf "$trabajo"` | `1`, `archive`, `https://github.com/jmorenobl/kitlegal/releases/download/v0.4.0/kitlegal-plugin.zip`, `0.4.0` | esas cuatro líneas; el temporal, retirado (`ls -d` ya no lo encuentra) | sí |
| 5 | `make release`; `ls`; `grep` de `checksums.txt`; `ls dist/*.tar.gz dist/*.zip` | los dos ficheros; dos líneas de `checksums.txt`; los seis archivos más `kitlegal-plugin.zip`, sin `kitlegal_darwin_all.tar.gz` | `release succeeded after 9s`, con el gancho `go run ./cmd/empaquetar piezas -version 0.3.2-SNAPSHOT-94fce58 …` tras `universal binaries`; `dist/kitlegal-plugin.zip` y `dist/kitlegal.mcpb`; una línea de cada uno en `checksums.txt`, con la huella que `shasum -a 256` da al fichero; siete nombres en el último listado, los seis archivos y `dist/kitlegal-plugin.zip`, y ningún `darwin_all` | sí |
| 6 | `make snapshot-check`; la lista de subpruebas | `ok` en sus dos órdenes; diez subpruebas, las cuatro de hoy y las seis nuevas | `ok … kitlegal 1.165s` y `ok … internal/app 10.906s`; diez `--- PASS`: `seis-archivos`, `sin-sbom-ni-firmas`, `version-del-binario`, `checksums-de-los-archivos`, `dos-piezas`, `manifiesto-de-la-extension`, `binarios-de-la-extension`, `icono-de-la-extension`, `servidor-de-la-extension`, `skills-del-plugin` | sí |
| 7 | `make schema-check`; `make skills-check`; `TestHerramientasDelServidor` | las tres en verde, sin haber tocado lo que comprueban | `ok` en `TestEsquemasPublicados`; `ok` en los tres paquetes de `skills-check`; `ok … internal/app 2.243s`. `schemas/`, `skills/`, `data/` y `evals/` no están en la lista de §1, y `herramientas_test.go` solo gana líneas (+125 −0) | sí |
| 8 | `make ci` | `ci: todos los controles en verde`, con `goreleaser check` | código 0 y esa línea al final de un registro de 100 líneas, sin `FAIL`, sin `(cached)` y sin truncar; `0 issues.` del lint; `1 configuration file(s) validated` de `goreleaser check`; `no leaks found`; `all modules verified` | sí |
| 9 | `rm -rf dist`; `git status --short` | nada que no se viera antes de §1 | `dist` ya no existe; la salida de `git status --short` es byte a byte la guardada antes de §1 (los dos ficheros de `gates/`) | sí |
| 10 | — | — | **No ejecutado**: es de la persona o del workflow (abajo) | — |

**Los tests de §1, por su nombre.** Una ejecución aparte con `-v` de `./internal/empaquetado/` da 54 líneas
`--- PASS` y ninguna `FAIL` ni `SKIP`: `TestPiezas` (2 subpruebas), `TestPiezasReproducibles`, `TestPiezasSinEntrada`
(7), `TestIcono` (5), `TestDescripcionCorta` (3), `TestCatalogo` (9) y `TestEjecutar` (21). `TestHerramientasAnunciadas`
pasa con sus cuatro subpruebas, los dos clientes por los dos registros, y `TestConfiguracionDeLaRelease`, con 23, entre
ellas las dos nuevas, `universal` y `catalogo`.

**`make vuln`, dentro de §8.** `govulncheck` llegó a su red y dice `No vulnerabilities found.` y `Your code is affected
by 0 vulnerabilities.` Añade que hay 1 vulnerabilidad en un módulo requerido a la que el código no llama; el control
sale en verde y el hito no cambia `go.mod`.

**Cómo se ejecutó.** Tres cosas difieren de teclear el quickstart en una terminal, y ninguna cambia una orden:

- §2 a §4 comparten la variable `trabajo`, y cada orden de la sesión abre una shell nueva, así que las tres fueron un
  solo guion escrito en el directorio temporal, con las órdenes del quickstart copiadas tal cual y una línea de marca
  entre ellas. `mktemp -d` sin plantilla creó el temporal en el directorio temporal del usuario que da macOS
  (`/var/folders/…/T/tmp.…`), no bajo `$TMPDIR`; está dentro de lo que el sandbox deja escribir, y §4 lo retiró. Antes
  de ese guion hubo un primer intento de §2 con una carpeta fija bajo `$TMPDIR` —las dos ejecuciones del paso, con
  código 0—, que se retiró sin terminar para repetir §2 a §4 enteros.
- Las órdenes llevan delante `rtk proxy`, que da la salida sin resumir; la de §8 se guardó en un fichero del directorio
  temporal, fuera del repositorio, para leerla entera.
- Antes de §8, `go clean -testcache`, para que los dos perfiles de §2 salgan de tests ejecutados y no de la caché.

**Dos cifras del plan que no salen igual, y no son umbrales.** `make release` tardó 9 s y no los 18 s del equipo del
plan. `dist/kitlegal.mcpb` mide 42 744 909 bytes y `dist/kitlegal-plugin.zip`, 17 738; plan.md, «Uso», da 42 711 814 y
17 739, medidos con el prototipo y otra versión en el manifiesto. Ningún requisito fija un tamaño ni un tiempo.

**§10, no ejecutado.** Ninguna sesión del run ejecuta `claude`, publica ni mide en la plataforma:

- `make plugin-check` (`claude plugin validate`, SC-008): lo ejecuta el trabajo `snapshot` de la propuesta de cambio.
- `release.yml`: lo dispara la etiqueta que empuja una persona. Ahí se ejecutan por primera vez la atestación de los dos
  ficheros, los seis pasos nuevos de `humo` y el trabajo `catalogo`.
- La prueba de SC-002: de una persona, tras publicar una release.
- El cierre —publicar la rama, abrir la propuesta de cambio y medir la CI y las evals—: del workflow, tras la revisión
  final.

## 4. Controles de umbral (plan.md)

Por cada una de las catorce filas de plan.md, «Controles de umbral», el test que la controla. Todos existen en la cabeza
—la línea de su `func` está abajo, y el árbol de trabajo es `94fce58` fuera de `gates/`— y `make ci` los ejecuta: `go
test -list` sin etiquetas de construcción los da en sus cuatro paquetes (ninguno de sus ficheros lleva `//go:build`),
el `-skip` de `make test` y de `make test-integration` solo aparta `TestMedidasDeTiempo` y `TestCosteDelGrafo`, y el
registro del `make ci` de §3 trae `ok` para los cuatro paquetes en las dos órdenes.

| Fila de plan.md | Test de `make ci` | Dónde está | Lo que el árbol da hoy |
|---|---|---|---|
| Descripción, 120 caracteres | `TestDescripcionCorta` | `internal/empaquetado/textos_test.go:25` | 71 caracteres (§3, escenario 2); la de 121 falla y la de 120 pasa |
| Icono, PNG de 512 × 512 px | `TestIcono`, `TestPiezas` | `internal/empaquetado/piezas_test.go:715` y `:391` | `mcp/icon.png`: PNG de 512 × 512 (`file`); los de 256 × 256, 256 × 512, 512 × 256 y lo que no es un PNG fallan |
| 0 bytes entre dos ejecuciones | `TestPiezasReproducibles` | `internal/empaquetado/piezas_test.go:549` | además, §3, escenario 2: `iguales byte a byte` |
| 4 entradas en el `.mcpb` | `TestPiezas` | `internal/empaquetado/piezas_test.go:391` | cuatro (§3, escenario 2) |
| `tools`, el conjunto exacto | `TestEjecutar`, `TestHerramientasAnunciadas` | `internal/empaquetado/ejecutar_test.go:52`; `internal/app/herramientas_test.go:499` | diez (§3, escenario 2); los diez nombres están escritos en `ejecutar_test.go` |
| 0 ficheros distintos en `skills` | `TestEjecutar`, `TestPiezas` | los mismos | `plugin.json` y cinco ficheros de skills (§3, escenario 2) |
| Manifiesto `0.3`, campos exactos | `TestPiezas` | el mismo | — |
| Binarios, byte a byte | `TestPiezas`, `TestConfiguracionDeLaRelease` (`universal`, `plataformas`) | `release_test.go:570` | `universal_binaries`: `id: kitlegal-universal`, `ids: [kitlegal]`, `replace: false` |
| 2 de 2 ficheros y huellas | `TestPiezas`, `TestConfiguracionDeLaRelease` (`checksums`, `publicacion`) | los mismos | los dos, detrás de `install.sh`, en `checksum.extra_files` y en `release.extra_files` |
| 6 archivos, ninguno universal | `TestConfiguracionDeLaRelease` (`archivos`, `universal`) | el mismo | `archives` con `ids: [kitlegal]` |
| El servidor desde el `.mcpb` | `TestPiezas`, `TestEntregaDelHito` (guion `h21-mcp-proceso`) | `internal/app/e2e_test.go:514`, sin tocar; el guion, en `internal/app/testdata/script/` | `TestEntregaDelHito/h21-mcp-proceso`, `PASS` |
| `claude plugin validate`, 2 de 2 | `TestConfiguracionDeLaRelease` (`trabajo-de-snapshot`, `objetivos-del-makefile`, `ayuda`) | el mismo | el trabajo `snapshot` de `ci.yml` tiene seis pasos; `VERSION_DE_CLAUDE_CODE` es `2.1.284` en `ci.yml` y en `evals.yml` |
| 0 paquetes del paso en el binario | `TestElBinarioNoEnlazaElPaso` | `internal/arch_test.go:281` | `go list -deps ./cmd/kitlegal` no nombra `internal/empaquetado` (0 líneas) |
| 3 trabajos, 2 pasos con el token, 9 sujetos, 12 pasos de `humo` | `TestConfiguracionDeLaRelease` (`flujo-de-la-release`, `tokens-de-la-publicacion`, `atestacion`, `humo`, `catalogo`) | el mismo | `publicar`, `humo` y `catalogo`; `secrets.PUBLISHER_TOKEN` en las líneas 66 y 317 de `release.yml` y en ningún otro flujo ni en el `Makefile`; nueve rutas en `subject-path`; 12 pasos en `humo` y 4 en `catalogo` |

Que cada control se pone en rojo lo vio su tarea, con sus casos negativos o con mutantes momentáneos que no quedan en
el diff (`tasks.md`, «Controles de umbral»): esta tarea no los repite ni añade ninguno.

### Las cinco medidas sobre el snapshot real

Su control es el trabajo `snapshot` de la propuesta de cambio, que el cierre del workflow cuenta (`medir_cierre`). **Las
cinco quedan para ese trabajo**: es el que decide, en `ubuntu-latest`. Lo que esta sesión vio de cada una, en macOS
`arm64` y sin valor de veredicto:

| Medida | Control del cierre | En esta sesión (§3, escenarios 5 y 6) |
|---|---|---|
| Los bytes de las arquitecturas (FR-062) | `TestSnapshot/binarios-de-la-extension` | `PASS` |
| Los dos ficheros en `dist/` y en `checksums.txt` (FR-060) | `TestSnapshot/dos-piezas` | `PASS` |
| Los seis archivos, ninguno universal (FR-006) | `TestSnapshot/seis-archivos` | `PASS` |
| El servidor arrancado desde el `.mcpb` (FR-064) | `TestSnapshot/servidor-de-la-extension` | `PASS`, con el binario de `darwin/arm64` |
| `claude plugin validate`, 2 de 2 (FR-065) | `TestPluginValido`, con `make plugin-check` | **Sin medir**: ninguna sesión del run ejecuta `claude`. Se vio que el test compila con `-tags=snapshot` (`go test -list` lo da) y nada más |

Las otras tres subpruebas nuevas (`manifiesto-de-la-extension`, `icono-de-la-extension`, `skills-del-plugin`) también
dieron `PASS` aquí, y las mide el mismo trabajo. En Linux no se midió nada: el arranque del servidor y `kitlegal skills
install` con el entorno de dos variables (`gates/supuestos.md`, T003) los verá ese trabajo por primera vez.

## 5. Los supuestos S1 a S11 de research

**Los once siguen sin medir.** Ninguna tarea del run pudo comprobarlos y esta tampoco: lo que hay en `make ci` fija la
definición de los flujos, no su ejecución.

| Supuesto | Qué se da por cierto sin haberlo medido | Dónde se verá |
|---|---|---|
| S1 | Los seis pasos nuevos de `humo` funcionan en `ubuntu-latest` como en macOS, donde se midieron sus órdenes | La primera release. Hoy solo se fijan sus líneas y `bash -n` los lee |
| S2 | `gh api --method PUT …/contents/…` crea el `marketplace.json` y lo sustituye con el `sha` del que hay | La primera release: el trabajo `catalogo` |
| S3 | `claude plugin validate` corre en el runner sin credencial ni sesión con modelo, no descarga la dirección del catálogo y da por válidos el plugin y el catálogo | El trabajo `snapshot` de la propuesta de cambio, en el cierre (SC-008) |
| S4 | `npm install -g @anthropic-ai/claude-code@2.1.284` funciona en `ubuntu-latest` | El mismo trabajo |
| S5 | `actions/attest-build-provenance@v4` atesta los dos ficheros más y `gh attestation verify` los verifica | La primera release: `publicar` y `humo` |
| S6 | goreleaser, al publicar, sube los dos ficheros de `release.extra_files` | La primera release: `humo` los descarga |
| S7 | La app de escritorio de Claude acepta el `.mcpb`: conserva el bit de ejecución de `server/kitlegal` y aplica `platform_overrides.win32` en Windows | La prueba de SC-002 en macOS; en Windows, nadie lo ha probado |
| S8 | La app admite un catálogo con una entrada de fuente `archive` | La prueba de SC-002, que lo anota |
| S9 | `PUBLISHER_TOKEN` puede escribir en `jmorenobl/kitlegal-plugins` | La primera release: si no puede, `catalogo` sale en rojo. Darle el permiso es de la persona (§6) |
| S10 | En `publicar`, sin caché de compilación, el `go run` del gancho añade a la release lo que tarde en compilar el paso | La primera release. Sin umbral: el spec no pide tiempo. Los 9 s de §3 son con la caché de este equipo y no lo miden |
| S11 | En `catalogo`, `actions/checkout` sin `ref` obtiene el commit de la etiqueta, y `ubuntu-latest` trae `gh`, `jq`, `unzip`, `base64` y `npm` | `npm`, en el trabajo `snapshot` del cierre; lo demás, en la primera release |

De los once, el cierre de este run solo puede medir S3, S4 y la parte de `npm` de S11. Los demás no se miden hasta que
una persona etiqueta (S1, S2, S5, S6, S9, S10, S11) o hace la prueba a mano (S7, S8).

Los supuestos de las decisiones tomadas en el run —los textos de la ficha y la autoría, el nombre del catálogo, la
plantilla del catálogo como código, los tres apartados del README— están en `gates/supuestos.md`, para la capa 3.

## 6. Lo que queda para la persona tras el run

1. **Que `PUBLISHER_TOKEN` pueda escribir en `jmorenobl/kitlegal-plugins`**, antes de la primera release (FR-032; S9).
   Sin ese permiso, el trabajo `catalogo` sale en rojo, sin afectar a `publicar` ni a `humo`, y se puede volver a
   ejecutar solo.
2. **La etiqueta** (ADR 0020), después de leer el informe final y de fusionar.
3. **La release**: es la primera ejecución real de la atestación de los nueve sujetos, de los seis pasos nuevos de
   `humo` y de `catalogo` (S1, S2, S5, S6, S10, S11). Lo que falle ahí no lo ha visto ningún control del run.
4. **La prueba de SC-002**, tras publicar: en un Mac distinto del que compiló, sin kitlegal y sin abrir una terminal,
   descargar `kitlegal.mcpb` y `kitlegal-plugin.zip` con un navegador, instalar los dos y, en una conversación nueva y
   sin carpeta, preguntar «¿qué dice el art. 21 de la Ley 39/2015?». Esperado: `art. 21 de la Ley 39/2015
   [BOE-A-2015-10565, bloque a21]`. Con sus cuatro anotaciones:
   - si macOS bloquea el binario descargado;
   - cómo llega una versión nueva de la extensión y del plugin;
   - si la app admite el marketplace (S8);
   - qué responde Claude en la web con el plugin y sin la extensión.
5. **Versionar el esquema oficial de la versión `0.3` del manifiesto de MCP Bundle** y validar el manifiesto contra él
   (FR-017; spec, «Fuera de alcance»): en una propuesta de cambio propia o corrigiendo la entrada del hito y
   relanzando. Su fuente no tiene hoy fila en `docs/SOURCES.md`. Hasta entonces la forma exacta del manifiesto solo la
   prueba la instalación a mano de SC-002.
