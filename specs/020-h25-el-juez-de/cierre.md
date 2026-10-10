# H25 · Cierre de la Definition of Done

Cierre de T008 (FR-001, FR-005, FR-011, FR-026, FR-027, FR-030, FR-033, FR-073, FR-083, FR-095, FR-096, FR-100,
FR-101; SC-012) sobre `7f0d96c`, que es T007, el último commit antes de T008. La rama está al día con `main`:
`git merge-base main HEAD` da la cabeza de `main`, `ece5df8`, así que `git diff --name-status main` da exactamente lo
que cambia el hito. Las siete partes son las de la tarea: lo creado, lo modificado y lo que no cambia (§1), las copias
y `SKILL.md` (§2), la cobertura (§3), el quickstart (§4), los controles de umbral (§5), los umbrales, el instrumento y
los atajos (§6) y los apartados de las reparaciones del cierre (§7).

Todo lo que se cita aquí se ejecutó en la sesión de T008, en primer plano, y se vio terminar, o está en el
repositorio. Lo que no se pudo medir se dice tal cual (§3, «El diff» y «Los recuentos»; §4, escenarios 10 y 11; §5,
las diez filas `evals:`; §6, «Sin modelo»; §8).

**T008 no toca código ni añade ningún test.** Las cifras de cobertura quedan sobre sus umbrales (§3), y la tarea solo
pide tests si alguna queda debajo. Sus únicos cambios son este fichero, dos líneas en `gates/supuestos.md` y su marca
en tasks.md. Las siete partes cuadran con lo que la tarea pide comprobar, y las veinticinco órdenes de quickstart.md
§0 a §9 dan lo esperado (§4). T008 dejó anotada una cosa que el hito dejaba falsa y que no podía tocar: un comentario
de `internal/evals/umbrales.go`.

**T009, después, corrige ese comentario y otros dos**, sobre `a8c553f`, que es T008: el de `umbralesDelInforme` en
`internal/evals/umbrales.go`, el de `TestUmbralesConJuezYConSentencias` en `internal/evals/umbrales_test.go` y el de
`TestConsultasNecesarias` en `internal/evals/consultas_test.go`. Solo cambian líneas de comentario: ni código, ni
tests, ni lo que ningún test espera (§1 y §6). De este fichero, T009 cambia lo que eso cambia, medido en su sesión
con las mismas órdenes y con los tres ficheros ya editados: esta cabecera; la lista y los recuentos de §1; lo que §5 y
§6 decían de que `umbrales.go` no cambiaba, con sus números de línea vueltos a leer; el recuento de «Atajos» (§6); y
§8, sin su primer punto. **La cobertura (§3), el quickstart (§4) y los controles de §5 son los que midió T008 sobre
`7f0d96c`, y T009 no los ha vuelto a medir.**

**La revisión final, después, retira dos mecanismos de `internal/evals/medida.go`**, sobre `062c399`, que es el
barrido (ronda 1, proporcionalidad). El primero, la lectura de una invocación de un sondeo cuya entrada lleva en
`command` algo que no es un texto: la clave vuelve a ser un texto, y el informe que lleve otra cosa no se lee, con la
regla que ya hay. El segundo, la rama de `argumentosDeLaInvocacion` que repartía las órdenes del applet `cita`, que
`repetirLasInvocaciones` ya reparte: la función vuelve a ser, byte a byte, la de `main`. En `medida_test.go` se va la
invocación que sostenía el primero, y `TestArgumentosDeLaInvocacion` da cada orden a su regla —las de `cita`, a
`partirLaOrdenDeCita`— sin cambiar lo que ningún caso espera. Los tests que fijan los 249 casos y sus textos siguen
en verde sin cambiar lo que esperan. De este fichero cambia lo que eso cambia, medido en la sesión del corrector con
las mismas órdenes: esta cabecera; los recuentos de §1; en §3, los cuatro números de línea y el párrafo «Después de
la revisión final»; los números de línea de `medida_test.go` en §4, §5 y §6; y, en §6, el recuento de «Atajos» y lo
que «Nada se corrigió después de crearlo» dice de `medida.go`. El quickstart (§4) no se ha vuelto a ejecutar entero:
de sus órdenes, esa sesión repite la de su §4 con `-v`, que da las mismas 92, y `make ci`, que termina en verde.

**El cierre, después, repara, y la revisión final juzga la reparación.** Tras la medición 1 (`3f99446`), la
reparación no cambió nada fuera de `gates/`. Tras la medición 2 (`e380f0a`), la reparación `0dc549c` cambia una viñeta de
`skills/boe-legislacion/SKILL.md`, que pasa a v0.1.8 (R1), y hace que el juez pida otra vez, una sola, el voto que su
tope corta, con el corte publicado en `juez.votos_cortados` (R2): `juez.go`, `informe.go` y `doc.go`; cinco ficheros
de test de `internal/evals`, uno de ellos, `sondeo_test.go`, nuevo en el diff; `CHANGELOG.md` y `CONTRIBUTING.md`
(research.md, «Reparaciones del cierre»). La ronda 3 de la revisión final devuelve a la de `main` la viñeta de
«Redacción modificada» que R1 había reescrito para ganar una línea, deja ese `SKILL.md` en 299 líneas y, para que
`make ci` las admita, cambia el caso `dos-inicios` de `TestSkillsDelRepositorio` (`internal/app/skills_test.go`). De
este fichero cambia lo que todo eso cambia, medido en la sesión del corrector de esa ronda con las mismas órdenes,
con sus cambios en el árbol y sin commit: esta cabecera; los recuentos, las tablas y «Lo que no cambia» de §1; en §3
y en §4, el párrafo «Después de las reparaciones del cierre»; los números de línea de `informe_test.go`, de
`juez_test.go` y de `medida_test.go` y los recuentos de tres tests en §5 y §6; en §5, lo que ya se sabe de las diez
filas `evals:`; en §6, «Nada se corrigió después de crearlo», lo que dice de `gates/evals/` y el recuento de
«Atajos»; §7; y §8. El quickstart (§4) no se ha vuelto a ejecutar entero: de sus órdenes, esa sesión repite con `-v`
los tests de §1 a §7, en una sola orden, y `make ci`, que termina en verde. Lo que cada sección dice haber medido
sobre `7f0d96c`, en T009 o en la ronda 1 es lo que se midió entonces.

## 1. Lo creado y lo modificado en el hito, y lo que no cambia

`git diff --name-status main`, con los tres comentarios de T009 ya corregidos, da **65 ficheros: 43 `A` y 22 `M`**,
ninguna `D` ni `R`. 34 están en `specs/020-h25-el-juez-de/`, todos `A`, con este `cierre.md`. Los otros 31 son 9 `A` y
22 `M` (`git diff --shortstat main -- . ':!specs'`: 7 698 líneas añadidas y 855 quitadas): los 28 que midió T008
antes de escribir este fichero (7 676 y 842) y los tres de T009 (22 y 13). El total con `specs/` no se da: cambia con
este mismo fichero; T008 midió 10 999 líneas añadidas y 842 quitadas en 61 ficheros, antes de escribirlo.
`git ls-files --others --exclude-standard` da un fichero sin seguimiento, `gates/converge-hecho`, que deja el
workflow, y `git status --porcelain`, antes de editar este fichero en T009, da además los tres ficheros de Go de la
tarea y los tres que el workflow lleva modificados (`tasks.md`, `gates/tarea-actual.json` y
`gates/tareas-intentos.json`). Son las cifras de cuando T009 escribió esto. Con su commit (`adf90b7`), que versiona
`gates/converge-hecho`, la misma orden da 66 ficheros, 44 `A` y 22 `M`, 35 de ellos en `specs/020-h25-el-juez-de/`
(medido en el barrido); los 31 de fuera de `specs/`, con sus 7 698 líneas añadidas y 855 quitadas, son los mismos, y
el barrido no toca ninguno. Los de `specs/` siguen creciendo con lo que el workflow deja en `gates/`. La revisión
final cambia dos de los 31, `medida.go` y `medida_test.go`: con ellos, `git diff --shortstat main -- . ':!specs'` da
7 694 líneas añadidas y 843 quitadas en los mismos 31 ficheros (medido en la sesión del corrector). Las reparaciones
del cierre y la ronda 3 de la revisión final cambian nueve de esos 31 y añaden tres:
`skills/boe-legislacion/SKILL.md`, `internal/evals/sondeo_test.go` e `internal/app/skills_test.go`. Con ellos,
`git diff --name-status main` da 89 ficheros, 64 `A` y 25 `M`, 55 de ellos en `specs/020-h25-el-juez-de/`, y
`git diff --shortstat main -- . ':!specs'`, 8 082 líneas añadidas y 969 quitadas en 34 ficheros, 9 `A` y 25 `M`
(medido en la sesión del corrector de la ronda 3, con sus cambios en el árbol).

### Los 34 ficheros

Cada uno con la tarea que lo tocó (`git log --format=%s main..HEAD -- <ruta>`) y sus líneas
(`git diff --numstat main`). Los tres de T009 están en el árbol y sin commit cuando se escribe esto: su tarea es la
de la sesión que los edita. Las líneas de `medida.go` y de `medida_test.go` son las de después de la revisión final,
que los edita (antes, +803 −105 y +2 493 −242). Las de los ficheros que tocan las reparaciones del cierre y la ronda 3
de la revisión final son las de después de ellas, medidas en la sesión del corrector de esa ronda, con las de antes
entre paréntesis.

**La carpeta del juez** (5, todos `A`, de T004), en `evals/jurisprudencia/juez/`:

| Fichero | Líneas | Qué es |
|---|---|---|
| `rubrica.md` | +62 | Copia de la rúbrica validada (§2). |
| `esquema.json` | +1 | Copia de la forma de la respuesta del juez (§2). |
| `casos.yaml` | +1 547 | Copia de los 249 casos etiquetados (§2). |
| `medida.json` | +27 | Copia de la medida versionada, con 0 de 125 y 0 de 124 (§2). |
| `clases.yaml` | +10 | El único que se escribe en la carpeta: `afirma_lo_no_leido`, que decide con 0, y `afirma_que_existe`, que solo se publica. |

**Las evals** (4, todas `A`, de T003), en `evals/jurisprudencia/`: `07-resumen-de-una-conocida.yaml` (+20),
`08-doctrina-con-el-fallo-delante.yaml` (+56), `09-de-que-trata-con-la-ficha-sola.yaml` (+28) y
`10-doctrina-dada-por-hecha.yaml` (+14).

**La skill** (1, `M`, de T006): `skills/jurisprudencia/SKILL.md`, +9 −7, que queda en 197 líneas (§2).

**La otra skill** (1, `M`, de la reparación R1 del cierre y de la ronda 3 de la revisión final):
`skills/boe-legislacion/SKILL.md`, +5 −4, que queda en 299 líneas. Su diff con `main` es un solo trozo,
`@@ -131,10 +131,11 @@`, la viñeta «De un precepto que no has leído, nada» (research.md, R1).

**El flujo** (1, `M`, de T003, T004 y T007): `.github/workflows/evals.yml`, +37 −21. Fuera de sus comentarios, el
diff son seis líneas: `concurrencia: 1` pasa a `concurrencia: 4` en la entrada de `jurisprudencia` del trabajo
`evals`; la matriz del trabajo `medida` pasa de `[boe-legislacion]` a `[boe-legislacion, jurisprudencia]` y gana su
entrada de `include`, con `concurrencia: 4`; y la descripción de la entrada `medir_al_juez` dice «de las skills que lo
tienen» donde decía «de boe-legislacion».

**Los ficheros de Go** (18, todos `M`): 17 en `internal/evals/` y, el último de la tabla, uno en `internal/app/`.

| Fichero | Tarea | Líneas |
|---|---|---|
| `juez.go` | T001 y la reparación R2 | +83 −28 (antes, +15 −2) |
| `informe.go` | T001 y la reparación R2 | +98 −17 (antes, +56 −16) |
| `medida.go` | T002 y la revisión final | +795 −105 |
| `conjunto.go` | T003 | +30 −21 |
| `doc.go` | T007 y la reparación R2 | +33 −13 (antes, +24 −7) |
| `umbrales.go` | T009 | +12 −5 |
| `juez_test.go` | T001, T004 y la reparación R2 | +440 −35 (antes, +322 −3) |
| `informe_test.go` | T001, T003, T004 y la reparación R2 | +805 −145 (antes, +738 −120) |
| `medida_test.go` | T002, T004, T005, la revisión final y la reparación R2 | +2 523 −241 (antes, +2 497 −230) |
| `conjunto_test.go` | T003, T004 | +404 −55 |
| `definicion_test.go` | T003, T004 | +415 −122 |
| `ejecucion_test.go` | T004 y la reparación R2 | +91 −19 (antes, +85 −17) |
| `sentencias_test.go` | T003 | +168 −43 |
| `sesion_test.go` | T004 | +114 |
| `sondeo_test.go` | la reparación R2 | +9 −8 (antes no estaba en el diff) |
| `umbrales_test.go` | T009 | +8 −6 |
| `consultas_test.go` | T009 | +2 −2 |
| `internal/app/skills_test.go` | la ronda 3 de la revisión final | +6 −4 (antes no estaba en el diff) |

Son 1 051 líneas añadidas a los seis que no son tests (eran 932), 4 979 a los once de test de `internal/evals` (eran
4 753 en diez) y 6 al de `internal/app`. Ningún fichero ni paquete de Go es nuevo. `sentencias_test.go` no estaba en
la lista de plan.md, «Source Code»: lo añade tasks.md, que dice por qué en su cabecera. Desde el barrido, esa lista lo
nombra, con `umbrales.go` y los dos ficheros de test de T009; y desde la ronda 3 de la revisión final, lo que tocan
las reparaciones del cierre.

**Los tres de T009 solo cambian comentarios.** En `git diff main` de los tres, cada línea añadida o quitada empieza
por `//`: un trozo en cada fichero, el final del comentario de `umbralesDelInforme` (`@@ -184,11 +184,18 @@`), el de
`TestUmbralesConJuezYConSentencias` (`@@ -533,12 +533,14 @@`) y el de `TestConsultasNecesarias`
(`@@ -24,8 +24,8 @@`). Ni una constante, ni una firma, ni un caso, ni un nombre de test. `gofmt -l` de los tres no
nombra ninguno. `grep` de «no tiene juez ni objetivo», de «tiene hoy las dos» y de «las seis evals» en ellos no da
nada, y antes de editarlos daba una línea en cada uno (`umbrales.go:188`, `umbrales_test.go:537` y
`consultas_test.go:27`). En los ficheros Go del paquete, ninguna otra frase dice que `jurisprudencia` no tiene juez,
que sus evals son seis o que sus umbrales son cuatro: «las seis» de `conjunto_test.go:1297` son las evals de H23,
nombradas junto a las cuatro de H25, y el total de 18 de `informe_test.go:4734` (era la línea 4685 antes de la
reparación R2) es el del contrato de H23, a propósito.

**La documentación** (4, todos `M`, de T007): `CHANGELOG.md` (+94 −15; antes, +74 −15), `CONTRIBUTING.md` (+118 −51;
antes, +106 −44), `docs/JURISPRUDENCIA.md` (+17 −1) y `docs/WORKFLOW.md` (+1 −1). Los dos primeros los cambian
también las reparaciones del cierre, y `CHANGELOG.md`, la ronda 3 de la revisión final.

### Lo que no cambia (FR-005, FR-011, FR-033, FR-073, FR-095, FR-096)

Cada comprobación es un `git diff --name-status main -- <rutas>`, ejecutado en esta sesión, que **no imprime nada**.
Entre paréntesis, los ficheros versionados que hay en esas rutas (`git ls-files`), para que el cero no sea el de una
ruta vacía:

- **La carpeta de la evidencia**: `evidencias` (49), con `evidencias/adr-0037-jurisprudencia` (18),
  `evidencias/adr-0037` (13) y `evidencias/adr-0036` (3). `git rev-parse main:evidencias` y `HEAD:evidencias` dan el
  mismo árbol, `7db1e2cd…`, y `git log main..HEAD -- evidencias` no nombra ningún commit.
- **El directorio de esquemas**: `schemas` (16), con el mismo árbol en `main` y en la cabeza, `3d689100…`.
- **Ningún fixture ni grabación**: el `testdata/` de la raíz (62); todo lo que hay bajo un `testdata/` en cualquier
  directorio (`'*testdata/*'`, 622); los dos manifiestos `grabaciones.json` con sus dos `grabacion_test.go` (4); e
  `internal/source`, `internal/httpx` e `internal/cache` (158). Tampoco `docs/SOURCES.md` ni
  `scripts/verify-sources.sh` (2). El paso `grabar_datos` no tenía nada que grabar.
- **Las seis evals de H23**: de `01-existe-con-numero-y-fecha.yaml` a `06-documento-que-no-es-el-pedido.yaml` (6).
- **Las evals de `boe-legislacion`, su juez y `legal-core`, byte a byte**: `evals/boe-legislacion` (27), con su
  carpeta del juez, `evals/boe-legislacion/juez` (5); `evals/legal-core` (4) y `skills/legal-core` (3). `git rev-parse`
  da el mismo árbol en `main` y en la cabeza a cada una de las cuatro carpetas: `799416c8…`, `c9a22f89…`, `20d74b03…`
  y `01a63a0a…`. En el flujo, las entradas de las dos skills siguen con `concurrencia: 4` y
  `objetivo_de_duracion: 900`, y con `concurrencia: 1` y `objetivo_de_duracion: 0`. **`skills/boe-legislacion` (2) sí
  cambia desde la reparación R1 del cierre**: hasta ella era también la de `main` (`60633173…`), y hoy su `SKILL.md`
  lleva una viñeta cambiada, +5 −4 (arriba, «La otra skill»); su `references/normas.md` no cambia.
- **El applet `cita` y su dominio**: `internal/app/cita.go` y `cita_test.go` (2); `internal/core/cita` (20) e
  `internal/core/ids` (60), e `internal/core` entero (164), con el árbol de `main`, `0ab0f68b…`. De `internal/app`
  (120), que hasta la ronda 3 de la revisión final era entero el de `main`, cambia un solo fichero, de test:
  `skills_test.go`, en el caso `dos-inicios`; ningún fichero que no sea un test. Tampoco cambian `cmd` (3), ni
  `internal/cli`, `internal/mcp` e `internal/skills` (54).
- **El guion del voto**: `scripts/evals-voto.sh` (1); `git rev-parse` de `main` y de la cabeza y `git hash-object`
  del árbol dan el mismo objeto, `6931a6e0…`. Tampoco nada de `scripts` entero (31).
- **`mensajeDelVoto`**: `juez.go` sí está en el diff, así que se compara la función. Con su comentario son 35 líneas
  en `main` y 35 en el árbol, y `cmp` no da ninguna diferencia. El diff de `juez.go` eran cuatro trozos, todos del
  campo `sentencia` de T001: el comentario de `VotoDeClase`, su campo `Sentencia`, la línea que lo copia en
  `pedirElVoto` y el campo de `dichoDeClase`. Desde la reparación R2 del cierre son diez: esos cuatro y seis del voto
  que el tope corta —las causas del tope, `Cortados` en `JuicioDeRespuesta`, el comentario y el cuerpo de `juzgar`,
  `votosDelNumero` con `motivoDelVoto`, y `pedirElVoto`, que devuelve la causa sin el número del voto—. El primero
  empieza en la línea 638, y `mensajeDelVoto` ocupa de la 293 a la 327: sigue sin estar en ninguno. Tampoco la
  constante de la petición, que dice «las dos preguntas» (`juez.go:290`).
- **Los dos topes, y el modelo y la versión del juez** (FR-033): en `main` y en el árbol, `timeout-minutes: 352` en el
  trabajo `evals` y `timeout-minutes: 269` en el de `medida`; `MODELO_DEL_JUEZ: claude-opus-5-5` y
  `VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.289`, en los dos trabajos. Siguen también `MODELO_DE_EVALS: claude-sonnet-5-5`,
  `VERSION_DE_CLAUDE_CODE: 2.1.284`, `MODELOS_INFORMATIVOS_DE_EVALS: claude-haiku-4-5-20251001`,
  `REPETICIONES_DE_EVALS: 3` y `UMBRAL_DE_EVALS: 2`. `git log -G` de esas ocho claves sobre `evals.yml`, en
  `main..HEAD`, no nombra ningún commit: solo cambian de número de línea.
- **El límite de ritmo** (FR-073): `reintentos_por_limite_de_ritmo` está en `juzgar.go`, que no está en el diff, y en
  `informe.go`, que sí. En lo añadido y lo quitado fuera de `specs/`, ninguna línea nombra el límite de ritmo.
  `ejecucion.go`, `sesion.go`, `plan.go`, `definicion.go`, `sentencias.go`, `formato.go` y `juzgar.go` (7) no
  cambian. `umbrales.go` sí está en el diff desde T009, y solo en un comentario (arriba, y §6). Las nueve sesiones a
  la vez no se han medido: las mide el cierre.
- **Ningún ADR**: `docs/ADR` (37), con el mismo árbol, `6ec9525b…`. De `docs/`, el diff solo nombra
  `docs/JURISPRUDENCIA.md` y `docs/WORKFLOW.md`.
- **La constitución**: `.specify/memory/constitution.md` (1), y `.specify` entero (45), con el mismo árbol,
  `6578722f…`.
- **Los guiones del workflow**: `scripts/workflow` (16), con el mismo árbol, `55a68d6c…`.
- **Ningún spec, plan ni informe de otro hito**: `-- specs ':!specs/020-h25-el-juez-de'` (999).
- **Lo demás**: `CLAUDE.md`, `docs/ROADMAP.md`, `docs/USO.md` y `web` (69); `data` (26); `go.mod`, `go.sum`, `tools` y
  `Makefile` (13); `.github` sin `evals.yml` (6); `.golangci.yml`, `codecov.yml` y `.goreleaser.yaml` (3); y
  `README.md` (1).

Las mismas órdenes sí imprimen donde hay cambios: con `-- internal/evals` salen 16 líneas; con
`-- evals/jurisprudencia`, 9; con `-- docs`, 2; y con `-- skills` y `-- .github`, una cada una. Tras las
reparaciones del cierre y la ronda 3 de la revisión final, con `-- internal/evals` salen 17; con `-- skills`, 2; y con
`-- internal/app`, una. Todas las demás comprobaciones de esta lista siguen sin imprimir nada, repetidas en la sesión
del corrector de esa ronda.

### La Definition of Done, punto a punto (`ROADMAP.md` §1)

| Punto | Aplica | Qué se ve en la cabeza |
|---|---|---|
| 1 · `make ci` en verde | sí | §4, escenario 9: código 0 y `ci: todos los controles en verde`, con la caché de tests vacía. |
| 2 · tests offline; fixtures si toca red | tests sí; fixtures no | Los tests de §4 y §5. El hito no toca la red: ninguna grabación ni fuente (arriba). |
| 3 · sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | sí | En las 932 líneas añadidas a los seis `.go` que no son tests, 0 coincidencias de `"net/http"`, `"net"`, `os.Exit`, `fmt.Print`, `os.Stdout`, `os.Stderr`, `panic(`, `"database/sql"`, `modernc.org` e `internal/httpx`. `golangci-lint`, en el `make ci` de §4: `0 issues.` Con la reparación R2 son 1 051 líneas, y la misma búsqueda sigue sin dar ninguna; `0 issues.` en el `make ci` de la ronda 3. |
| 4 · esquemas y `schema-check` | ningún esquema cambia | `schemas` sin cambios; `TestEsquemasPublicados`, `ok` dentro de `make ci`. |
| 5 · errores con código estable | ningún código de `kitlegal` cambia | `internal/cli` e `internal/core` sin cambios; de `internal/app`, solo un caso de un test (ronda 3 de la revisión final). |
| 6 · e2e y `CHANGELOG.md` | «Aceptación e2e: no aplica»; `CHANGELOG.md`, sí | `CHANGELOG.md` +74 −15, en *Unreleased*, con «La skill `jurisprudencia` v0.1» en su línea 58 (T007; SC-012). Con las dos entradas de las reparaciones del cierre, +94 −15. |
| 7 · ADR | no | `docs/ADR` sin cambios: la decisión es el ADR 0037. |
| 8 · `SOURCES.md` | no | `docs/SOURCES.md` sin cambios: ninguna fuente. |
| 9 · cobertura | sí | §3. |
| 10 · skill: evals antes, `SKILL.md` < 300, sin drift | sí | Las evals son de T003 (`363d682`) y la skill, de T006 (`028ca7c`); 197 líneas; `make skills-check` en verde (§4, escenario 8). El `SKILL.md` de `boe-legislacion`, que cambia la reparación R1, tiene 299, y `skills-check` pasa en el `make ci` de la ronda 3. |
| 11 · dimensión territorial | no | — |
| 12 · operaciones de grafo | no | Ningún applet cambia. |
| 13 · recursos, plazos o escritos | no | — |

## 2. Las cuatro copias y `SKILL.md` (FR-001, FR-030, FR-083)

### Las copias

`cmp` de cada copia de `evals/jurisprudencia/juez/` con su original de `evidencias/adr-0037-jurisprudencia/`: **las
cuatro sin salida y con código 0**.

| Fichero | Bytes | `sha256` (principio) | `cmp` con su original |
|---|---|---|---|
| `rubrica.md` | 6 007 | `e87abe125391c7e8` | sin salida, código 0 |
| `esquema.json` | 611 | `5e509a544b41ae46` | sin salida, código 0 |
| `casos.yaml` | 63 870 | `aa05c7792471b10d` | sin salida, código 0 |
| `medida.json` | 715 | `ee34bf308aa375a3` | sin salida, código 0 |

La misma orden sí habla cuando hay diferencia: `cmp` de `clases.yaml` con `rubrica.md` dice
`differ: char 1, line 1` y sale con 1. `ls -A` de la carpeta da los cinco ficheros y nada más, y
`git log main..HEAD` sobre ella nombra un solo commit, el de T004 (`180e79e`): las copias se escribieron una vez y
ninguna tarea posterior las tocó.

`medida.json` lleva `skill: jurisprudencia`, `clase: afirma_lo_no_leido`, `fecha: 2026-10-09`,
`modelo_del_juez: claude-opus-5-5`, `version_de_claude_code: 2.1.289`, `defectos` con 125 casos y 0 sin marcar, y
`correctos` con 124 y 0 marcados. Las dos huellas que guarda son las de las copias: `shasum -a 256` de `rubrica.md`
da `e87abe12…021136aa` y el de `casos.yaml`, `aa05c779…00fa537e`, enteras e iguales a las del fichero. El modelo y la
versión son los que fija el flujo (§1).

### `SKILL.md` contra el de `main`

`git diff --numstat main -- skills/jurisprudencia/SKILL.md` da `9  7`: 195 líneas en `main` y 197 en el árbol. El diff
tiene **dos trozos y nada más**:

- `@@ -16,12 +16,13 @@`, el párrafo inicial (C1, el CAPTCHA): «el buscador pide un CAPTCHA a los programas» pasa a
  «el buscador cierra el paso a los programas con un CAPTCHA», y se añade que el obstáculo es de los programas, que a
  la persona no le sale y que la respuesta no se lo anuncia ni le pide que resuelva ninguno. El resto del párrafo es
  el mismo texto, con otros cortes de línea.
- `@@ -74,7 +75,8 @@`, la viñeta del equivalente (C2): tras «puedes nombrarlo junto a la referencia» se añade «y
  entonces dilo como lo que es: deducido de esa referencia, sin que nadie lo haya comprobado».

Que son los dos pasajes del contrato y solo ellos: el `SKILL.md` de `main`, con
`contracts/skill-jurisprudencia-v0.1.diff` aplicado en un directorio temporal (`patch -p1`, código 0), es **idéntico
byte a byte** al del árbol (`cmp`, sin salida). Solo un commit toca el fichero, el de T006 (`028ca7c`).
`grep -c CAPTCHA` da 2, las dos menciones del párrafo inicial.

## 3. Cobertura (Definition of Done §1.9)

Se mide sobre los dos perfiles que deja el `make ci` de §4, escenario 9, en verde y con la caché de tests vacía
(`go clean -testcache` justo antes; el registro no trae ningún `(cached)`): `coverage.out` (`make test`) y
`coverage-integration.out` (`make test-integration`). El porcentaje sale de `go tool cover -func`: el global, sobre
cada perfil; el de cada árbol o fichero, sobre el perfil filtrado a sus líneas, con su cabecera `mode:`, escrito en el
directorio temporal. Los recuentos de sentencias salen de los bloques del perfil —su segundo campo—, contando cada
bloque una vez. «Unión» son los dos perfiles fundidos bloque a bloque, como hace Codecov con los dos que publica la
CI.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | **96,5 %** (10 701 de 11 095 sentencias) | **96,9 %** (10 749 de 11 095) | 96,9 % | ≥ 70 % |
| `internal/core/**` (dominio) | **98,3 %** (2 100 de 2 136) | **98,3 %** (2 100 de 2 136) | 98,3 % | ≥ 85 % |
| `internal/core/cita` | 100 % (188 de 188) | 100 % (188 de 188) | 100 % | — |
| `internal/core/ids` | 100 % (151 de 151) | 100 % (151 de 151) | 100 % | — |
| `internal/cli/**` | 98,0 % (398 de 406) | 98,0 % (398 de 406) | 98,0 % | ≥ 90 % |
| `internal/evals` (el paquete de evals) | **97,2 %** (3 772 de 3 880) | **97,7 %** (3 789 de 3 880) | 97,7 % | — |
| `internal/evals/juez.go` | 99,2 % (254 de 256) | 99,2 % (254 de 256) | 99,2 % | — |
| `internal/evals/informe.go` | 98,8 % (556 de 563) | 98,8 % (556 de 563) | 98,8 % | — |
| `internal/evals/medida.go` | 97,6 % (494 de 506) | 97,6 % (494 de 506) | 97,6 % | — |
| `internal/evals/conjunto.go` | 98,8 % (254 de 257) | 98,8 % (254 de 257) | 98,8 % | — |

**Los umbrales se cumplen y no hace falta ningún test**: T008 no toca ningún `_test.go`. El porcentaje de
`go tool cover -func` coincide con el de los bloques en todas las filas, y el del paquete de evals, con el que
imprime `go test` en el registro de `make ci` (97,2 % y 97,7 %).

**Lo que no llega al 100 % en los cuatro ficheros que la tarea declara.** Son 24 bloques de una sentencia, los mismos
en los dos perfiles, en 18 de sus 209 funciones. Veinte están en líneas que ya eran de `main` (`git blame`), y
cuatro, en líneas de T002, todas de `medida.go`:

- `:883` (`leerInforme`, 93,3 %): el informe versionado no es JSON.
- `:957` (`preguntaDelSondeo`, 92,9 %): las evals de hoy de la skill no se pueden leer.
- `:1258` (`registroDeLaSesion`, 80,0 %): un applet no se puede registrar.
- `:1472` (`verbo`, 66,7 %): una orden sin verbo.

Las dos primeras son ficheros versionados que no se pueden leer, que plan.md, «Trazabilidad», último párrafo, deja a
la regla que ya hay. A las otras dos no llega ningún test, tampoco los que reconstruyen los casos del repositorio.
Ninguna tiene un test, y la tarea no pide añadirlo: ningún umbral depende de ellas.

**Después de la revisión final.** La corrección quita dos sentencias de `medida.go`, las de la rama retirada de
`argumentosDeLaInvocacion`, que tenían test. Sobre los dos perfiles que deja el `make ci` de la sesión del corrector,
en verde, y con los recuentos sacados de los bloques como arriba: `medida.go`, 492 de 504 en los dos (97,6 %); el
paquete de evals, 3 770 y 3 787 de 3 878 (97,2 % y 97,7 %); y el global, 10 699 y 10 745 de 11 093, con 96,5 % y
96,9 % en `go tool cover -func`. En las tres, la unión da lo que el perfil de integración. Los bloques sin test de
`medida.go` siguen siendo doce, y los cuatro de T002 son los de arriba, con el número de línea que tienen hoy y el
mismo porcentaje por función. En el global del perfil de integración hay dos sentencias cubiertas menos de las que
da restar esas dos a la cifra de T008 (10 747): no son del paquete de evals, cuyo recuento cuadra, y no se ha buscado
de qué paquete son. Lo demás de esta sección es lo que midió T008 sobre `7f0d96c`.

**Después de las reparaciones del cierre.** R2 añade 17 sentencias al paquete de evals, siete a `juez.go` y diez a
`informe.go`, todas con test. Sobre los dos perfiles que deja el `make ci` de la sesión del corrector de la ronda 3,
en verde, y con los recuentos sacados de los bloques como arriba: `juez.go`, 261 de 263 en los dos (99,2 %);
`informe.go`, 566 de 573 en los dos (98,8 %); `medida.go` y `conjunto.go`, como estaban, 492 de 504 y 254 de 257; el
paquete de evals, 3 787 y 3 804 de 3 895 (97,2 % y 97,7 %, los que imprime `go test` en el registro); el dominio e
`internal/cli`, como estaban, 2 100 de 2 136 y 398 de 406; y el global, 10 716 y 10 762 de 11 110, con 96,5 % y
96,9 % en `go tool cover -func`. Las sentencias sin test de `juez.go` y de `informe.go` siguen siendo dos y siete:
ninguna es de la reparación. La ronda 3 de la revisión final cambia un caso de un test y una viñeta de una skill, y
no toca ninguna sentencia. Las filas de `juez.go`, de `informe.go`, del paquete y del global de la tabla son las de
T008.

**Los recuentos no son comparables con los del cierre de H23.** Aquel fichero dio 15 015 sentencias en el global y
2 910 en el dominio; hoy son 11 095 y 2 136, con las mismas 36 sin cubrir en el dominio y con un dominio que entre
aquel commit (`19f332f`) y `main` solo cambia en un fichero (+23 −7). Los bloques sí cuadran: 9 433 entonces y 9 598
hoy. Entre los dos cierres `go.mod` pasó de `toolchain go1.27.1` a `go1.27.2`. Aquí el recuento se ha contrastado por
dos caminos —la suma por bloque en `awk` y `sort -u` con suma—, que dan lo mismo. No se ha buscado por qué difieren
los de entonces. No cambia ninguna cifra de la tabla: los porcentajes son los de `go tool cover -func`.

**El perfil de integración repite bloques.** `coverage.out` trae 9 598 líneas de bloque, una por bloque;
`coverage-integration.out`, 15 540 para los mismos 9 598 bloques: 2 971 salen tres veces, los de catorce paquetes
—los seis de `internal/core`, `internal/cli`, `internal/graph`, `internal/mcp` y `data`, entre ellos—. Es lo que ya
anotó el cierre de H23, y tampoco aquí se ha buscado por qué. No cambia ninguna cifra: los recuentos cuentan cada
bloque una vez.

**El diff.** `codecov/patch` es bloqueante con `target: auto`, la cobertura de la base, y `main` no se midió en esta
sesión, así que no se estima. Lo que diga solo se sabe en la propuesta de cambio.

## 4. Quickstart (§0 a §9)

Los diez escenarios se ejecutaron en su orden, en primer plano, sobre `7f0d96c` y con el árbol limpio salvo los dos
ficheros de `gates/`. Son veinticinco órdenes, y **las veinticinco dan lo esperado**. Las veinticuatro de §0 a §8 se
ejecutaron dos veces, la segunda con la caché en un directorio bajo `$TMPDIR` (abajo, «Cómo se ejecutó»), y dan lo
mismo en las dos.

| § | Orden | Esperado | Obtenido | ¿Cuadra? |
|---|---|---|---|---|
| 0 | `go version` | Go 1.27 | `go version go1.27.2 darwin/arm64` | sí |
| 0 | `make build` | `bin/kitlegal` | código 0; `bin/kitlegal`, de 25 361 410 bytes | sí |
| 0 | `export KITLEGAL_CACHE_DIR="$(mktemp -d)"` | — | código 0; un directorio nuevo y vacío | sí |
| 1 | `ls evals/jurisprudencia/juez` | cinco ficheros: `casos.yaml`, `clases.yaml`, `esquema.json`, `medida.json`, `rubrica.md` | esos cinco | sí |
| 1 | `cmp evidencias/adr-0037-jurisprudencia/rubrica.md evals/jurisprudencia/juez/rubrica.md` | sin salida | sin salida, código 0 | sí |
| 1 | `cmp …/esquema.json …/juez/esquema.json` | sin salida | sin salida, código 0 | sí |
| 1 | `cmp …/casos.yaml …/juez/casos.yaml` | sin salida | sin salida, código 0 | sí |
| 1 | `cmp …/medida.json …/juez/medida.json` | sin salida | sin salida, código 0 | sí |
| 1 | `go test -count=1 -run '^TestCopiasDelJuez$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	0.432s` | sí |
| 2 | `go test -count=1 -run '^(TestMedidaVersionada\|TestEjecucionSinMedir)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	0.393s` | sí |
| 3 | `ls evals/jurisprudencia` | diez evals, de la `01-…` a la `10-doctrina-dada-por-hecha.yaml`, y la carpeta `juez` | las diez, con esos nombres, y `juez` | sí |
| 3 | `go test -count=1 -run '^(TestPreguntasDelSondeo\|TestPreguntasConElFragmento\|TestConjuntoDeEvals\|TestEvalsDelRepositorio)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	0.488s` | sí |
| 3 | `bin/kitlegal cita preparar --resolucion 241/2013 --fecha 2013-05-09 --json` | código 0, `"direccion":"https://www.poderjudicial.es/search/indexAN.jsp"` y tres casillas: «Nº Resolución» con `241/2013` y «Fecha resolución», «Desde» y «Hasta», con `09/05/2013` | código 0; esa `direccion`, literal, y tres casillas: `{"nombre":"Nº Resolución","valor":"241/2013"}` y las dos de «Fecha resolución», con `"campo":"Desde"` y `"campo":"Hasta"` y `"valor":"09/05/2013"` | sí |
| 3 | `bin/kitlegal cita preparar --texto "cláusula suelo transparencia" --json` | código 0 y `"direccion":"https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo%20transparencia/1/AN"` | código 0; esa `direccion`, literal, y `"casillas":[]` | sí |
| 3 | `bin/kitlegal cita cotejar --json < evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` | código 0, `"roj":"STS 3144/2023"`, `"ecli":"ECLI:ES:TS:2023:3144"` y `"url":"kitlegal:documento/sha256:4886e0c8…ee27"`, la huella del fragmento | código 0; los tres, literales, con la huella entera, que es la que da `shasum -a 256` del fragmento; `"correspondencia":"se-corresponden"` y `"hallazgos":[]` | sí |
| 4 | `go test -count=1 -run '^(TestArgumentosDeLaInvocacion\|TestTextoQuitado\|TestResolverCasos\|TestReconstruccionDeJurisprudencia\|TestGrabacionesDerivadas)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	2.256s` | sí |
| 5 | `go test -count=1 -run '^TestEjecucionDeLaMedida$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	2.393s` | sí |
| 6 | `go test -count=1 -run '^(TestVotoDelJuez\|TestMensajeDelVoto\|TestTextosDeLaSesion\|TestUmbralesDeJurisprudencia\|TestInformeConElJuez)$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	1.329s` | sí |
| 7 | `grep -n -A 2 'skill: jurisprudencia' .github/workflows/evals.yml` | dos entradas de `include` con `concurrencia: 4`, la del trabajo `evals` —con `objetivo_de_duracion: 0`— y la del trabajo `medida` | las dos: en la línea 187, con `concurrencia: 4` y `objetivo_de_duracion: 0`, y en la 394, con `concurrencia: 4` | sí |
| 7 | `go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/` | `ok` | `ok  	…/internal/evals	0.426s` | sí |
| 8 | `wc -l skills/jurisprudencia/SKILL.md` | 197 líneas | `197` | sí |
| 8 | `grep -c CAPTCHA skills/jurisprudencia/SKILL.md` | `2` | `2` | sí |
| 8 | `git diff --stat main -- skills/jurisprudencia/SKILL.md` | un fichero con 9 líneas añadidas y 7 quitadas | `1 file changed, 9 insertions(+), 7 deletions(-)` | sí |
| 8 | `make skills-check` | en verde | código 0; `ok` en `internal/app`, `internal/skills` e `internal/evals` | sí |
| 9 | `make ci` | en verde, con `schema-check`, `skills-check` y las reglas del conjunto | código 0 y `ci: todos los controles en verde`, en 301 s y con la caché de tests vacía. Registro de 102 líneas: 53 `ok`, ningún `FAIL` ni `(cached)`; `0 issues.` del lint; `No vulnerabilities found.` y `Your code is affected by 0 vulnerabilities.`; `ok` en `TestEsquemasPublicados` y en los tres paquetes de `skills-check`, que lleva `TestEvalsDelRepositorio` con las reglas del conjunto; `1 configuration file(s) validated`; `no leaks found`; `all modules verified`, seis veces; `go mod tidy -diff` sin salida | sí |

`govulncheck` dice además que hay dos vulnerabilidades en módulos requeridos que el código no llama. Es la misma línea
que trae el registro de la verificación de T007 (`gates/ci.log`), y no hace fallar el control.

**Los tests, por su nombre.** Cada orden de `go test` se ejecutó además con `-v`, para contar sus tests y subpruebas y
que ninguna pasara en vacío. Ninguna trae un `--- FAIL` ni un `--- SKIP`. Las cifras son líneas `--- PASS` en todos
los niveles, contando el propio test:

- §1, 21: `TestCopiasDelJuez`, con `del-repositorio`, `boe-legislacion`, `jurisprudencia` y
  `skill-fuera-de-la-tabla`. La de `jurisprudencia` lleva ocho: cada una de las cuatro copias, cambiada y sin ella.
- §2, 50: `TestMedidaVersionada` (25) y `TestEjecucionSinMedir` (25), cada una con `boe-legislacion` y
  `jurisprudencia`. La de esta skill lleva las seis mutaciones —`rubrica-cambiada`, `casos-cambiados`,
  `modelo-fijado-cambiado`, `version-fijada-cambiada`, `defectos-sin-marcar` y `correctos-marcados`— y cuatro casos
  más, y en `TestEjecucionSinMedir`, además, `medida-que-corresponde`.
- §3, 109: `TestPreguntasDelSondeo` (5: un byte cambiado en la 07, la 08, la 09 y la 10),
  `TestPreguntasConElFragmento` (6: un byte cambiado en la 04, la 06, la 08 y la ficha, y
  `con-mas-del-fragmento-que-la-ficha`), `TestConjuntoDeEvals` (82, con `boe-legislacion`, `jurisprudencia` y
  `legal-core`) y `TestEvalsDelRepositorio` (16, entre ellas `conjunto-jurisprudencia`).
- §4, 92: `TestArgumentosDeLaInvocacion` (19), `TestTextoQuitado` (23), `TestResolverCasos` (40),
  `TestGrabacionesDerivadas` (9, con `boe-legislacion` y `jurisprudencia`) y `TestReconstruccionDeJurisprudencia`
  (1, sin subpruebas), que fija los 249 con 138, 39 y 72 por informe (`medida_test.go:3264` y `:3265`).
- §5, 23: `TestEjecucionDeLaMedida`, con `jurisprudencia-los-249-bien`, `jurisprudencia-un-defecto-sin-marcar`,
  `jurisprudencia-un-correcto-marcado` y `jurisprudencia-un-caso-que-no-se-resuelve`; los 499 votos son la constante
  `votosDeJurisprudencia` (`medida_test.go:4234`).
- §6, 66: `TestVotoDelJuez` (23), `TestMensajeDelVoto` (7, con una sesión de `jurisprudencia` de cada modo),
  `TestTextosDeLaSesion` (11, con `cita-cotejar-por-orden` y `cita-cotejar-por-llamada`), `TestInformeConElJuez`
  (15, con `con-sentencia`, `un-voto-que-no-llega` y `dos-sin-juzgar`) y `TestUmbralesDeJurisprudencia` (10:
  `el-elemento-del-contrato`, `ninguna-cuenta`, `una-marcada-en-el-modo-orden`, `una-marcada-en-las-dos-clases`,
  `tres-con-si-en-afirma-que-existe`, `901-s-del-juez-en-el-modo-orden`, `una-sin-documento-en-el-modo-orden`,
  `dos-sin-documento-en-el-modo-herramienta` y `una-sin-activar`).
- §7, 97: `TestDefinicionDelJob`, con siete de primer nivel, entre ellas `del-repositorio`, `sinteticas`, `peor-caso`
  y `tope-con-las-evals-del-repositorio`; esta lleva `jurisprudencia-de-una-en-una` y
  `jurisprudencia-de-una-en-una-en-la-medida`. Los 12 577 s y los 15 505 s están en `definicion_test.go:1858` y
  `:1874`, y los topes de 352 y 269 minutos, en `:1047` y `:1048`.

**Cómo se ejecutó.** Cuatro cosas difieren de teclear el quickstart en una terminal:

- Las órdenes de cada sección se sacaron de sus bloques de código de quickstart.md con `awk` y las ejecutó, una a una
  y diciendo el código de cada una, un guion de un solo uso escrito fuera del repositorio (`rtk proxy bash <guion>`,
  que da la salida sin resumir). La de §9 se lanzó aparte, también en primer plano, con `go clean -testcache` delante.
- **El directorio de la caché.** La tarea pide la caché en un directorio bajo `$TMPDIR`. La orden de §0, `mktemp -d`
  sin plantilla, crea en macOS el directorio en el temporal de usuario (`/var/folders/…/T/`), que en esta sesión no
  es `$TMPDIR` (`/tmp/claude-501`). Por eso §0 a §8 se ejecutaron dos veces: con la orden tal cual, y otra vez con
  esa sola orden sustituida por `mktemp -d "${TMPDIR%/}/kitlegal-h25.XXXXXX"`, que deja la caché bajo `$TMPDIR`. Las
  veinticuatro órdenes dan el mismo código y la misma salida estándar en las dos —comparadas con `cmp`, sin
  `fecha_consulta`, la fecha del enlazado ni la duración de los tests—. En las dos el directorio tiene 0 entradas al
  final (`ls -A`): `cita` no escribe en la caché. La tabla lleva las salidas de la primera. Queda en
  `gates/supuestos.md`.
- La salida de error se guardó aparte. Las tres órdenes del binario de §3 traen en ella la línea
  `aviso: las skills instaladas son de kitlegal v0.5.0 y este binario es kitlegal v0.5.0-26-g7f0d96c-dirty; ejecuta:
  kitlegal skills install -g`: es el aviso de H19, que depende de las skills instaladas en la cuenta de esta máquina y
  no de este hito. Ninguna otra orden escribe en ella.
- El registro de `make ci`, los de las ejecuciones con `-v` y los perfiles filtrados de §3 se guardaron en un
  directorio bajo `$TMPDIR`; los guiones y las salidas de la primera ejecución, en uno del temporal de usuario.
  Ninguno está en el repositorio.

`git status --porcelain` da lo mismo antes de §0 y después de cada sección, de §0 a §9: los dos ficheros de `gates/`.
`bin/kitlegal`, `coverage.out` y `coverage-integration.out` los ignora git (`git check-ignore`).

**Después de las reparaciones del cierre.** La sesión del corrector de la ronda 3 repite con `-v`, en una sola orden,
los tests de las órdenes `go test` de §1 a §7: ningún `--- FAIL` ni `--- SKIP`. Cambian tres recuentos, los de los
tests que toca R2: `TestVotoDelJuez`, 27 (eran 23); `TestInformeConElJuez`, 16 (eran 15), con
`votos-cortados-y-pedidos-otra-vez`; y `TestEjecucionDeLaMedida`, 24 (eran 23), con
`un-correcto-marcado-tras-un-voto-cortado`. Con ellos §5 da 24 y §6, 71; los demás tests dan lo de arriba. De los
que la reparación toca y el quickstart no nombra: `TestOrdenDelVoto`, 16; `TestInformeMarkdownDeLosUmbrales`, 8;
`TestJuicioDelSondeo`, 31; y `TestSalidaDelSondeo`, 9. `wc -l skills/boe-legislacion/SKILL.md` da 299, y
`TestSkillsDelRepositorio`, con `-v`, 113 líneas `--- PASS`, entre ellas los tres `dos-inicios`, el de cada skill. La
tabla, «Los tests, por su nombre» y «Cómo se ejecutó» son lo que se ejecutó sobre `7f0d96c`.

**§10 y §11, no ejecutados.** Ninguna sesión del run abre una sesión con modelo ni mide en la plataforma:

- §10, el cierre: lo hace el workflow tras la revisión final. Empuja la rama, abre la propuesta de cambio, pone la
  etiqueta `evals` y guarda el informe de cada skill en `gates/evals/`, que hoy no existe (SC-001).
- §11, después del run: una persona pone la etiqueta `evals-medir-juez` antes de fusionar, y lee las respuestas con
  algún voto afirmativo y las de las cuatro evals nuevas (SC-014).

## 5. Controles de umbral (plan.md)

Las 22 filas de plan.md, «Controles de umbral»: 10 con un control `evals:` y 12 con uno o más `ci:`.

### Las doce filas `ci:`

Cada control se busca como lo busca el informe final (`estado_control`, en el guion del informe): la ruta, con
`git cat-file -e HEAD:<ruta>`, y el test, con `git grep -nE '^(func <Test>\(|<Test>:)' HEAD -- <ruta>`. La tabla
nombra 18, 17 distintos —`TestUmbralesDeJurisprudencia` está en dos filas—: **los 17 están en su ruta**. La misma
búsqueda con un nombre que no existe no da nada y sale con 1. Ninguno de los siete ficheros de test lleva
`//go:build`, el `-skip` de `make test` y de `make test-integration` solo aparta `TestMedidasDeTiempo` y
`TestCosteDelGrafo`, y el registro de `make ci` trae `ok` para `internal/evals` en esas dos órdenes.

| Fila de plan.md | Control `ci:` | Dónde está | Lo que el árbol da hoy (§4) |
|---|---|---|---|
| SC-001 (sin juzgar) | `TestInformeConElJuez` | `internal/evals/informe_test.go:4018` | `PASS`, 16, con `un-voto-que-no-llega` |
| FR-001, FR-102, SC-002 | `TestCopiasDelJuez` | `internal/evals/medida_test.go:673` | `PASS`, 21, con las ocho de `jurisprudencia` |
| FR-072, FR-103, SC-003 | `TestDefinicionDelJob` | `internal/evals/definicion_test.go:194` | `PASS`, 97, con las dos mutaciones «de una en una» |
| FR-031, FR-104, SC-004 | `TestMedidaVersionada` y `TestEjecucionSinMedir` | `internal/evals/medida_test.go:275` e `internal/evals/ejecucion_test.go:110` | `PASS`, 25 y 25, con las seis mutaciones de `jurisprudencia` en cada una |
| FR-044, FR-105, SC-005 | `TestReconstruccionDeJurisprudencia` | `internal/evals/medida_test.go:3351` | `PASS` |
| FR-045, FR-106, SC-006 | `TestGrabacionesDerivadas` | `internal/evals/medida_test.go:3792` | `PASS`, 9, con `jurisprudencia` |
| FR-051, FR-107, SC-007 | `TestEjecucionDeLaMedida` | `internal/evals/medida_test.go:4404` | `PASS`, 24, con los cuatro casos de `jurisprudencia` |
| FR-013, FR-108, SC-008 | `TestVotoDelJuez` y `TestUmbralesDeJurisprudencia` | `internal/evals/juez_test.go:1296` e `internal/evals/informe_test.go:4711` | `PASS`, 27 y 10 |
| FR-010, FR-109, SC-009 | `TestTextosDeLaSesion` y `TestMensajeDelVoto` | `internal/evals/sesion_test.go:800` e `internal/evals/juez_test.go:424` | `PASS`, 11 y 7 |
| FR-065, FR-066, FR-110, SC-010 | `TestPreguntasDelSondeo`, `TestPreguntasConElFragmento`, `TestConjuntoDeEvals` y `TestEvalsDelRepositorio` | `internal/evals/conjunto_test.go:3707`, `:3540`, `:752` y `:1893` | `PASS`, 5, 6, 82 y 16 |
| FR-083, SC-011 | `Makefile:skills-check` | `Makefile:126` (`skills-check: check-tools`), que `ci` lleva entre sus dependencias (`Makefile:264`); `TestSkillsDelRepositorio`, en `internal/app/skills_test.go:83` | `make skills-check`, código 0; `SKILL.md`, 197 líneas |
| FR-111, SC-013 | `TestUmbralesDeJurisprudencia` | `internal/evals/informe_test.go:4711` | `PASS`, 10, con `901-s-del-juez-en-el-modo-orden` y `tres-con-si-en-afirma-que-existe` |

Que cada control se pone en rojo lo vio su tarea, con sus casos negativos o con mutantes momentáneos que no quedan en
el diff (tasks.md, «Controles de umbral»): esta tarea no los repite ni añade ninguno.

Los números de línea de `informe_test.go` y de `juez_test.go` y los recuentos de `TestInformeConElJuez`,
`TestEjecucionDeLaMedida` y `TestVotoDelJuez` son los de después de la reparación R2 del cierre, vueltos a leer en la
sesión del corrector de la ronda 3 (antes, `:4006`, `:4670`, `:1271` y `:422`, y 15, 23 y 23). Las demás líneas y
recuentos no cambian con ella. El control de la fila «sin juzgar» sigue fallando con un voto que no llega: desde R2,
el que el tope corta también al pedirse otra vez (plan.md, «Controles de umbral»).

### Las diez filas `evals:`

Su control es el informe del job de evals, que el workflow lanza en el cierre: **ninguna se mide en el run**, y aquí
no hay ninguna cifra suya.

| Filas | Control | Lo que se ve en el árbol |
|---|---|---|
| FR-020, SC-001 (dos modos) | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5:<modo>` | La clase, con `decide: true` y `umbral: 0`, en `clases.yaml`; el umbral lo construye `umbralesDeLasClases` (`internal/evals/umbrales.go`), cuyo código no cambia (§6) |
| FR-022, SC-001 (defectos y correctos) | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:<recuento>` | `umbralesDeLaMedida`, con `umbralDeLaMedidaDelJuez = 0` y `Decide: true`, sobre la medida versionada: 0 de 125 y 0 de 124 (§2) |
| FR-023, SC-001 (dos modos) | `evals:jurisprudencia:duracion_del_juez:<modo>` | `umbralesDeLaDuracion`, con `segundosDelJuezPorModo = 900` y `Decide: true` |
| FR-024, SC-001 (cita sin documento, dos modos) | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:<modo>` | El de H23, que sigue: `umbralDeCitaSinDocumento`, sin cambios |
| FR-024, SC-001 (sin activar, dos modos) | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:<modo>` | El de H23, que sigue: `umbralDeSinActivar`, sin cambios |

Los diez nombres están escritos, literales, en `internal/evals/informe_test.go` (`grep -c -F` de cada uno da al menos
1).

**Después, el cierre las ha medido.** En la medición 2, sobre `e380f0a` (`gates/evals/jurisprudencia.json`), las diez
se cumplen: 0 de 30 en `afirma_lo_no_leido`, en `cita_sin_documento` y en `sin_activar` en cada modo; 0 de 125 y 0 de
124 en la medida del juez; y 40 s y 47 s de juez. El veredicto de ese trabajo es `fallo` por otra cosa, una respuesta
sin juzgar, y el de `boe-legislacion`, por `afirma_lo_no_leido`, 1 de 54 en cada modo: son las dos reparaciones de
research.md, «Reparaciones del cierre». La cabeza con las reparaciones no se ha medido todavía: lo hace el workflow
tras la revisión final.

## 6. Umbrales, instrumento y atajos (FR-026, FR-095)

### Ningún umbral de FR-020, FR-022 o FR-023 se rebajó, pasó a `decide: false` ni perdió respuestas o casos

- **Los valores y `decide`.** `internal/evals/umbrales.go` está en el diff desde T009, y solo en un comentario: su
  diff con `main` es un trozo, `@@ -184,11 +184,18 @@`, de doce líneas añadidas y cinco quitadas que empiezan todas
  por `//`, el final del comentario de `umbralesDelInforme` (§1). Los valores, `Decide` y los totales son los de
  `main`, siete líneas más abajo los que van detrás de ese comentario: `umbralDeLaMedidaDelJuez = 0` y
  `segundosDelJuezPorModo = 900` (líneas 26 y 27), con `Decide: true` escrito en los dos umbrales de la medida y en el
  de la duración (líneas 361, 369 y 390). El de cada clase toma `Umbral` y `Decide` de `clases.yaml` (líneas 279 y
  280), que lleva `afirma_lo_no_leido` con `decide: true` y `umbral: 0`. `afirma_que_existe` lleva `decide: false`
  porque FR-021 la quiere solo publicada, desde T004: no es un umbral que haya dejado de decidir.
- **Los totales.** El de la clase es `len(grupo.respuestas)`, las respuestas juzgadas del modo (`umbrales.go:267`);
  los de la medida, los casos de la medida versionada. `TestUmbralesDeJurisprudencia` fija `"total":30` en el
  elemento de la clase que decide (`informe_test.go:4743`) y «1 de 30 (3,3 %)» en los motivos de una respuesta
  marcada en cada modo (`:4808` y `:4849`; las tres líneas, vueltas a leer tras la reparación R2). Las diez evals de `evals/jurisprudencia/` llevan `activa: true`, y con
  `REPETICIONES_DE_EVALS: 3`, que no cambia (§1), son las 30 por modo; el total de las sesiones reales lo da el job.
- **Los casos.** `casos.yaml` tiene 249 entradas, 125 con `etiqueta: defecto` y 124 con `etiqueta: correcto`, y es
  idéntico a su original (§2). `medida_test.go` lleva `defectosDeJurisprudencia = 125` y
  `correctosDeJurisprudencia = 124` (líneas 86 y 87).
- **La rúbrica, los casos y la medida son los de sus originales** (FR-026, FR-030): las cuatro copias, `cmp` sin
  salida (§2); la carpeta de la evidencia, con el mismo árbol que en `main` (§1); y un solo commit sobre la carpeta
  del juez, el de T004.
- **Nada se corrigió después de crearlo.** Cada pieza del instrumento la toca una sola tarea: la carpeta del juez,
  T004; `juez.go` e `informe.go`, T001; `medida.go`, T002; `conjunto.go`, T003; y `SKILL.md`, T006. `evals.yml` lo
  tocan T003, T004 y T007, esta última solo en comentarios (§1), y `umbrales.go`, T009, solo en un comentario. Las
  nueve tareas van por su primer intento (`gates/tareas-intentos.json`), la cuarentena está vacía y no hay ninguna
  nota `gates/tarea-T*.md`. La revisión final, después, retira de `medida.go` dos mecanismos sin requisito
  (cabecera): no toca la rúbrica, los casos, la medida ni ningún umbral, y `TestReconstruccionDeJurisprudencia`,
  `TestGrabacionesDerivadas` y `TestEjecucionDeLaMedida`, que fijan los 249 casos, sus textos y los 499 votos, siguen
  en verde sin cambiar lo que esperan. La reparación R2 del cierre, después, cambia `juez.go` e `informe.go`: el voto
  que el tope corta se pide otra vez, una sola. Tampoco toca la rúbrica, los casos, la medida ni ningún umbral:
  `umbrales.go`, `definicion.go` y `evals/` no están en su diff (`git diff --stat 4a127c5 0dc549c`, sin salida para
  esas rutas), y `votosPorRespuestaComoMucho` sigue en `2 * votosParaMarcar` (`definicion.go:53`), el presupuesto con
  el que `TestDefinicionDelJob` calcula el peor caso. Lo que cambia de lo que esperan los tests es el protocolo:
  donde bastaba un corte para dejar una respuesta sin juzgar hacen falta dos.

### Sin modelo, sin medida y sin sondeo (FR-095)

**Esta sesión** no abrió ninguna sesión con modelo ni usó la red, salvo la de `govulncheck` dentro de `make ci`: no
ha ejecutado `claude`, `make evals`, `make evals-sondeo`, `make evals-medir-juez`, ningún guion `scripts/evals*.sh`,
`TestEjecucionDelJob`, `TestMedidaDelJuez`, `TestSondeo` ni los guiones de Python de la evidencia. Esos tres tests
están en `internal/evals/job_test.go`, con `//go:build evals`, que no está en el diff y que `make ci` no compila.

**De las sesiones de T001 a T007**, esta no vio nada: lo que el repositorio dice de ellas es que la medida no se
repitió ni se escribió —`medida.json` es el de la evidencia, con fecha 2026-10-09, anterior al run—, que no hay
ningún informe de job ni de sondeo en el diff —`gates/evals/` no existe, y `evidencias/` no cambia— y que cada una
dejó `make ci` en verde en su verificación (`gates/ci.log` guarda la de T007).

**Después, `gates/evals/` existe**: lleva los tres informes que el cierre recogió del job de su medición 2
(`boe-legislacion.json`, `jurisprudencia.json` y `legal-core.json`). Los deja el workflow, que lanza el job en la
plataforma; ninguna sesión de un paso lo abre. La del corrector de la ronda 3 tampoco abrió ninguna sesión con
modelo, ni la medida ni un sondeo: de la red, solo `govulncheck` dentro de `make ci`.

### Atajos: ningún `//nolint`, `t.Skip` ni TODO nuevos

En las 7 694 líneas añadidas fuera de `specs/` (`git diff main -- . ':!specs'`), con las 22 de T009 y lo que cambia
la revisión final (eran 7 698 antes de ella), la búsqueda de
`nolint`, `t.Skip`, `.Skip(`, `SkipNow`, `TODO`, `FIXME` y `XXX` no da ninguna; la misma búsqueda con `fmt.Errorf` da
21. Ninguna línea añadida empieza por una asignación a `_`, que es como se tira un error. Con las reparaciones del
cierre y la ronda 3 de la revisión final son 8 082 líneas añadidas: las mismas búsquedas, repetidas en la sesión del
corrector de esa ronda, siguen sin dar ninguna, y la de `fmt.Errorf` da 24. Y en el árbol entero, contra `main`:

- **`//nolint:`**: `git grep -c '//nolint:' main -- '*.go'` y la misma orden sobre el árbol dan lo mismo, **7
  directivas en `main` y 7 en el árbol**, en los mismos seis ficheros de test (`coste_test.go`, `e2e_test.go`, con
  dos, y `tuberia_unix_test.go`, de `internal/app`; `internal/arch_test.go`;
  `internal/core/territorio/coste_test.go`; e `internal/evals/cierre_test.go`), ninguno en el diff.
- **`t.Skip`**: los dos de la cabeza son los dos de `main`, en las mismas líneas
  (`internal/cache/integracion_test.go:451` e `internal/evals/cierre_test.go:70`).
- **TODO, FIXME y XXX**: 12 líneas en `main` y 12 en el árbol, en `*.go`, `Makefile`, `scripts`, `.github`, `skills`,
  `evals` y `schemas`. De los ficheros que las llevan, el diff solo nombra `internal/evals/medida_test.go`, con dos:
  son la plantilla de `mktemp`, `kitlegal-medida-del-juez.XXXXXX`, en dos comentarios que ya estaban en `main` y que
  hoy están más abajo (líneas 5482 y 5516, tras la reparación R2). Ninguna es una tarea pendiente.

Los tres recuentos del árbol entero —7 directivas, dos `t.Skip` y 12 líneas— son también los de después de las
reparaciones del cierre y de la ronda 3 de la revisión final, que no añade ninguno en `internal/app/skills_test.go`.

## 7. Los apartados de las reparaciones del cierre (FR-027)

**Hoy dos de los tres llevan las reparaciones, y el tercero sigue como estaba.** plan.md, «Decisiones», lleva las dos
reparaciones del cierre, cada una con lo que se aparta de la sección del hito y del spec; research.md,
`## Reparaciones del cierre (FR-027)`, las desarrolla como R1 y R2, con la medición 2 como evidencia, lo rechazado y
lo que queda sin medir; y `gates/supuestos.md` las lleva en sus líneas `reparar_cierre`, las dos con impacto de
alcance. contracts/skill-jurisprudencia.md §6 sigue con «Ninguna todavía», y es cierto: ninguna reparación toca el
`SKILL.md` de `jurisprudencia`. `grep -ci 'ninguna todavía'` da 0, 0 y 1 (sesión del corrector de la ronda 3). Lo que
sigue es lo que había antes de la primera medición.

Los tres seguían entonces en su sitio, con «ninguna todavía»:

- plan.md, «Decisiones», línea 296 (era la 290 hasta el barrido, que añade seis líneas más arriba, en «Source Code»,
  en «Structure Decision» y en «Tests existentes que cambian»): «**Reparaciones del cierre** (FR-027): ninguna todavía. Si las hay, cada una se
  anota aquí y en research.md, con la medición como evidencia.»
- research.md, `## Reparaciones del cierre (FR-027)`, línea 221: «Ninguna todavía. Si una medición del cierre da pie a
  un cambio de `SKILL.md` fuera de los dos pasajes, la decisión se anota aquí […]».
- contracts/skill-jurisprudencia.md, `## 6. Reparaciones del cierre (FR-027)`, línea 87: «Ninguna todavía. Una
  reparación que toque `SKILL.md` añade aquí su cambio —C3 en adelante— […]».

`grep -ci 'ninguna todavía'` daba 1 en cada uno, y cuando se escribió esto solo un commit tocaba los tres, el del plan
(`2698bf9`): ninguna tarea los ha cambiado. El barrido, después, edita plan.md y research.md en otros apartados —las
cifras de «Uso», «Source Code», «Structure Decision» y «Tests existentes que cambian», y la fila M10—, y no toca ninguno de estos tres. Es donde una
reparación del cierre deja su traza, y hasta la primera medición no hay ninguna.

## 8. Lo que queda fuera del run

El comentario de `internal/evals/umbrales.go` que T008 anotó aquí, y en `gates/supuestos.md`, como algo que veía y no
podía tocar lo corrige T009 (§1). Queda esto:

- **Las diez filas `evals:`** (§5) y SC-001: las mide el job de cierre que lanza el workflow. Con ellas, lo que el
  run no puede medir (tasks.md, «Estrategia de implementación»): cuánto tarda un voto, si nueve sesiones a la vez
  chocan con el límite de ritmo, si las series de las cuatro evals nuevas pasan y si v0.1 quita el CAPTCHA de las
  respuestas.
  Las dos primeras mediciones del cierre ya están en `gates/` (§5): la tercera, sobre la cabeza con las reparaciones,
  es la última (FR-027).
- **El efecto de las dos reparaciones del cierre**: si `boe-legislacion` v0.1.8 quita de la eval 15 la materia de lo
  no leído, y cuántos votos corta el tope y llegan al repetirse, lo mide esa tercera medición (research.md, R1 y R2).
- **Lo que decide la reparación R2**, que H24 había dejado a una persona: a ella le queda decidir si quiere que el
  voto cortado se repita, con el informe final, donde va con impacto de alcance.
- **`codecov/patch`** (§3): solo se sabe en la propuesta de cambio.
- **quickstart.md §10 y §11**, SC-014 y SC-015: del workflow, de una persona y de la revisión final.
- **La verificación de cada tarea**, la de T008 y la de T009, es otro `make ci`, posterior a su última edición de este
  fichero: su resultado no está aquí.
