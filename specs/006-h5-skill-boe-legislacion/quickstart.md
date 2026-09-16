# Quickstart: validación de H5

Guía **ejecutable** para comprobar que H5 entrega lo que dice. Cada escenario se ejecuta tal cual, desde la raíz del
repositorio, sobre la rama `h5-skill-boe-legislacion` **con el hito ya implementado y confirmado**. Los nombres de test
son los del inventario de [plan.md](./plan.md) («Inventario de tests»); los formatos, los de [contracts/](./contracts/).

**Sin efectos colaterales.** Ningún escenario crea ni modifica un fichero del árbol de trabajo, toca el índice o el
historial de git, escribe en `~/.claude/skills` o en el directorio de binarios de la cuenta, ni en la caché de
`kitlegal` de la cuenta. Todo lo que se construye, instala o rompe a propósito se hace en
`/tmp/kitlegal-quickstart-h5/`: un **clon desechable** del repositorio (`git clone` solo lee el repositorio), un `HOME`
y un `GOBIN` temporales. Los prerrequisitos empiezan borrando esa carpeta por si una ejecución interrumpida la dejó con
contenido, y la limpieza la borra al final; una ejecución retomada empieza siempre por los prerrequisitos. Lo único que
puede aparecer en el árbol son `coverage.out` y `coverage-integration.out` de las órdenes de `make` que ejecutan
`go test` con perfil, ignorados desde H0 (`.gitignore`: `/coverage.*`). Los escenarios de §12 son de plataforma: los
ejecuta la tarea `[plataforma]` y solo escriben en la propuesta de cambio (etiquetas) y en
`specs/006-h5-skill-boe-legislacion/gates/`, como declara cada uno.

**El directorio del hito queda fuera de las comprobaciones del árbol**: el workflow `hito` reescribe ahí sus ficheros de
estado antes de cada tarea, así que `git status` descarta `specs/006-h5-skill-boe-legislacion/` en cualquier estado.

**Cada orden se ejecuta tal cual en el modo desatendido**, sin sustituciones, con las formas que la sesión
`claude -p --permission-mode acceptEdits` acepta sin pedir aprobación (lista de permitidos de `.claude/settings.json`):

| Necesidad | Forma de la guía | Forma que pide aprobación y no se usa |
|---|---|---|
| Crear, borrar o editar en `/tmp/kitlegal-quickstart-h5/` | `rtk proxy mkdir`, `rtk proxy rm`, `rtk proxy chmod`, `rtk proxy ln`, `rtk proxy perl -0pi` | `mkdir` y `rm` fuera del repositorio, `perl`, `printf … >> fichero` |
| Ejecutar un enlace o un binario de la carpeta temporal | `rtk proxy /tmp/kitlegal-quickstart-h5/…` | `/tmp/kitlegal-quickstart-h5/…` |
| Leer el código de salida o fijar variables para una orden | `rtk proxy sh -c '…; echo "código $?"'` | `…; echo "código $?"` fuera de `sh -c`, `VARIABLE=valor orden` |
| Comprobar que algo existe o no | `rtk proxy test …  && echo "…"` | `test` sin `rtk proxy` |
| Leer un enlace | `rtk proxy readlink …` | `readlink` sin `rtk proxy` |
| Filtrar la salida de `go test -v`, `git status` o `git diff` | `rtk proxy go test …`, `rtk proxy git …`, con cada etapa de la tubería con `rtk proxy` | sin `rtk proxy`, que reescribe esas salidas |
| Contar las líneas de un fichero de la carpeta temporal | `rtk proxy wc -l /tmp/kitlegal-quickstart-h5/…` | `wc` sin `rtk proxy`, que Claude Code bloquea fuera del directorio de trabajo, también como orden sola |
| Una orden larga | en una sola línea | una línea que termina en `\` seguida de otra que empieza por `\|` |

`go`, `git`, `make` y `echo` con texto literal están permitidos tal cual, también con rutas de la carpeta temporal
(`git clone`, `make -C`, `git -C`); `wc`, solo con ficheros del repositorio.

## Prerrequisitos

```bash
go version
git --version
make --version
git rev-parse --abbrev-ref HEAD
rtk proxy git status --porcelain | rtk proxy grep -vE '^.. specs/006-h5-skill-boe-legislacion/' ; echo "fin del estado"
rtk proxy mkdir -p /tmp/kitlegal-quickstart-h5
rtk proxy chmod -R u+w /tmp/kitlegal-quickstart-h5
rtk proxy rm -r /tmp/kitlegal-quickstart-h5
rtk proxy mkdir /tmp/kitlegal-quickstart-h5
git clone --quiet --branch h5-skill-boe-legislacion . /tmp/kitlegal-quickstart-h5/repo
rtk proxy sh -c 'true; echo "código $?"'
rtk proxy test ! -e /tmp/kitlegal-quickstart-h5/home && echo "sin HOME temporal todavía"
rtk proxy wc -l /tmp/kitlegal-quickstart-h5/repo/.agents/.gitattributes
```

Esperado: `go version go1.27.1 …`; la rama `h5-skill-boe-legislacion`; el estado muestra solo `fin del estado`; el
clon se crea sin mensajes; las tres sondas imprimen `código 0`, `sin HOME temporal todavía` y `1` seguido de la ruta
del `.gitattributes` del clon, que tiene una sola línea (research.md D18): es la forma con la que el escenario 4 cuenta
las líneas de un fichero del clon.

## 1. La skill y sus controles en verde (FR-040 a FR-043, FR-075, SC-005 en limpio, SC-008, SC-010)

```bash
make skills-check
wc -l skills/boe-legislacion/SKILL.md
rtk proxy readlink skills/boe-legislacion/scripts/boe
rtk proxy ls skills/boe-legislacion skills/boe-legislacion/references skills/boe-legislacion/scripts
rtk proxy grep -c '<!-- generado desde data/normas.yaml, no editar -->' skills/boe-legislacion/references/normas.md
```

Esperado: `make skills-check` termina en 0 con `ok` para `internal/app`, `internal/skills` e `internal/evals`; menos de
300 líneas; el enlace da `../../../bin/instalado/kitlegal`; `SKILL.md`, `references/normas.md` y `scripts/boe` y nada
más; la cabecera aparece `1` vez.

## 2. `make skills-sync` no cambia nada sobre el árbol limpio (US5-1, FR-033, FR-034)

```bash
make -C /tmp/kitlegal-quickstart-h5/repo skills-sync
rtk proxy git -C /tmp/kitlegal-quickstart-h5/repo status --porcelain ; echo "fin del estado"
make -C /tmp/kitlegal-quickstart-h5/repo skills-sync
rtk proxy git -C /tmp/kitlegal-quickstart-h5/repo status --porcelain ; echo "fin del estado"
rtk proxy grep -n 'skills-sync' /tmp/kitlegal-quickstart-h5/repo/Makefile
```

Esperado: las dos regeneraciones terminan en 0 y los dos estados muestran solo `fin del estado`; el `Makefile` ya no
dice que las skills lleguen en H5.

## 3. Deriva de lo generado (US5-2, US5-3, FR-042, SC-005)

Cada caso rompe el clon, comprueba y lo restaura con `git checkout`.

```bash
rtk proxy perl -0pi -e 's/\z/| a mano | | | | |\n/' /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/references/normas.md
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- skills/boe-legislacion/references/normas.md
rtk proxy perl -0pi -e 's/Ley 39\/2015, de 1 de octubre,/Ley 39\/2015, de 2 de octubre,/' /tmp/kitlegal-quickstart-h5/repo/data/normas.yaml
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- data/normas.yaml
rtk proxy perl -0pi -e 's/Devuelve el texto vigente de un bloque de una norma/Devuelve el texto de un bloque de una norma/' /tmp/kitlegal-quickstart-h5/repo/internal/app/boe.go
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- internal/app/boe.go
```

Esperado, en este orden: (a) `código 2` con un fallo que nombra `boe-legislacion` y `references/normas.md`
(`contenido-distinto`); (b) `código 2` con un fallo que nombra `boe-legislacion` y `references/normas.md`, y otro de
`TestIdentificadoresDeLasNormas` que nombra `BOE-A-2015-10565` y el título que no coincide (SC-008); (c) `código 2` con
un fallo que nombra `boe-legislacion` y `SKILL.md`.

## 4. Defectos de la skill (US5-4, US5-5, FR-036, FR-040, FR-041, SC-005)

```bash
rtk proxy perl -0pi -e '$n = () = /\n/g; $_ .= "\n" x (300 - $n)' /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/SKILL.md
rtk proxy wc -l /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/SKILL.md
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- skills/boe-legislacion/SKILL.md
rtk proxy perl -0pi -e 's/^name: boe-legislacion$/name: Boe-Legislacion/m' /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/SKILL.md
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- skills/boe-legislacion/SKILL.md
rtk proxy rm /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/scripts/boe
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- skills/boe-legislacion/scripts/boe
rtk proxy ln -sfn ../../../bin/kitlegal /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/scripts/boe
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- skills/boe-legislacion/scripts/boe
rtk proxy ln -s ../../../bin/instalado/kitlegal /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/scripts/cita
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
rtk proxy rm /tmp/kitlegal-quickstart-h5/repo/skills/boe-legislacion/scripts/cita
rtk proxy go test -count=1 -run '^TestSkillsDelRepositorio$' -v ./internal/app/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: `300` líneas y `código 2` con un fallo que nombra `boe-legislacion` y las 300 líneas; `código 2` nombrando
`boe-legislacion` y el `name`; `código 2` con `enlace-ausente` de `scripts/boe`; `código 2` con
`enlace-con-otro-destino` de `scripts/boe`; `código 2` con `enlace-sobrante` de `scripts/cita`; y en el árbol real, todos
los subtests de `TestSkillsDelRepositorio` en `PASS` (frontmatter, enlaces, región, trescientas líneas, regenerar dos
veces, normas nombradas, sin instrucciones de evals).

## 5. `data/normas.yaml` y su esquema (US6, FR-021, FR-025, SC-005)

```bash
rtk proxy perl -0pi -e 's/(    abreviatura: LPAC\n)/$1    vertical: fiscal\n/' /tmp/kitlegal-quickstart-h5/repo/data/normas.yaml
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- data/normas.yaml
rtk proxy perl -0pi -e 's/^  BOE-A-2015-10565:$/  BOE-A-15-10565:/m' /tmp/kitlegal-quickstart-h5/repo/data/normas.yaml
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- data/normas.yaml
rtk proxy go test -count=1 -run '^(TestLeerNormas|TestEsquemaDeNormas|TestRenderizarNormas)$' -v ./internal/skills/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: `código 2` nombrando `BOE-A-2015-10565` y `vertical`; `código 2` nombrando `BOE-A-15-10565` y la forma del
identificador; y los subtests de los tres tests en `PASS`, entre ellos los de norma sin título, sin rango, sin materias e
identificador repetido.

## 6. Evals: formato, conjunto y lo grabado (US4-4, US4-6, FR-060 a FR-064, FR-075, SC-010)

```bash
rtk proxy ls evals/boe-legislacion
rtk proxy grep -l 'reproduce: boe-fiscal' evals/boe-legislacion/06-irpf-rendimientos-del-trabajo.yaml
rtk proxy perl -0pi -e 's/^pregunta: .*\n//m' /tmp/kitlegal-quickstart-h5/repo/evals/boe-legislacion/01-lpac-articulo-21.yaml
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- evals/boe-legislacion/01-lpac-articulo-21.yaml
rtk proxy rm /tmp/kitlegal-quickstart-h5/repo/internal/source/boe/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_metadatos.json
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- internal/source/boe/testdata/boe.legislacion-consolidada
rtk proxy perl -0pi -e 's/("titulo_empieza_por"\s*:\s*("(?:[^"\\]|\\.)*"))(.*?"titulo_empieza_por"\s*:\s*)"(?:[^"\\]|\\.)*"/$1$3$2/s' /tmp/kitlegal-quickstart-h5/repo/testdata/evals/grabaciones.json
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h5/repo skills-check; echo "código $?"'
git -C /tmp/kitlegal-quickstart-h5/repo checkout -- testdata/evals/grabaciones.json
rtk proxy go test -count=1 -run '^(TestEvalsDelRepositorio|TestConjuntoDeEvals|TestLeerConjunto|TestLeerEval|TestPrepararYComprobar|TestPrepararDirectorioDeSesion|TestManifiestoDeGrabaciones|TestGrabacionesSinSolape)$' -v ./internal/evals/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: los doce ficheros de `contracts/evals-y-grabaciones.md` §2; el fichero 06 listado; `código 2` nombrando
`01-lpac-articulo-21.yaml` y `pregunta`; `código 2` nombrando `01-lpac-articulo-21.yaml` y
`boe articulo BOE-A-2015-10565 a21`; `código 2` con un fallo de `TestIdentificadoresDeLasNormas` que nombra las
entradas 1 y 2 del manifiesto y su `titulo_empieza_por` repetido (la sustitución copia en la segunda entrada el prefijo
de la primera, sea cual sea el texto que dejó la pausa de grabación; contrato de evals §3.1, regla 5); y todos los
subtests en `PASS`, entre ellos `TestManifiestoDeGrabaciones/clave-repetida`, `TestManifiestoDeGrabaciones/prefijo-repetido`,
`TestManifiestoDeGrabaciones/prefijo-de-otro-prefijo`, `TestManifiestoDeGrabaciones/repositorio`, `sin-metadatos-de-un-bloque-esperado`,
`sin-indice-de-una-norma`, `TestLeerConjunto/con-mal-formadas`, `TestLeerConjunto/entradas-que-no-son-evals`,
`TestEvalsDelRepositorio/conjunto`, `TestEvalsDelRepositorio/normas-conocidas`,
`TestPrepararDirectorioDeSesion/eval-normal`, `TestPrepararDirectorioDeSesion/prueba-de-red`,
`TestPrepararDirectorioDeSesion/eval-inexistente`, `TestPrepararDirectorioDeSesion/eval-mal-formada` y
`TestPrepararDirectorioDeSesion/con-faltas` (contrato de evals §1 y §2, contrato del job §9).

## 7. Comparación mecánica (US4-2, US4-3, US4-5, US4-7, US4-8, FR-072, FR-076, SC-009)

```bash
rtk proxy go test -count=1 -run '^(TestInterpretarInvocacion|TestExtraerCitas|TestJuzgar|TestLeerSesion|TestLeerTrazas|TestLeerTrazasSinFicheros|TestInforme|TestEscribirInformeSinSusEntradas)$' -v ./internal/evals/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: todos en `PASS`, entre ellos `TestJuzgar/sc-009-otro-bloque`, `TestJuzgar/sc-009-otra-norma`,
`TestJuzgar/articulos-satisface-un-bloque`, `TestJuzgar/metadatos-no-satisface-indice`,
`TestJuzgar/fuera-de-lo-grabado-con-y-sin-offline`, `TestJuzgar/positiva-no-activada`,
`TestJuzgar/no-activa-pero-activada`, `TestJuzgar/sesion-sin-terminar-no-activa`,
`TestJuzgar/sesion-sin-terminar-positiva`, `TestJuzgar/bloque-leido-sin-codigo`,
`TestJuzgar/describe-y-dry-run-no-satisfacen`, `TestJuzgar/otra-fallida`, `TestLeerSesion/codigo-distinto-de-cero`, `TestLeerSesion/tope-agotado`,
`TestLeerSesion/senal-tras-el-tope`, `TestLeerSesion/sin-result`, `TestLeerSesion/result-con-is-error`,
`TestLeerSesion/sin-fichero-de-codigo`, `TestLeerSesion/codigo-no-entero`, `TestLeerSesion/sin-salida-de-error`,
`TestLeerSesion/sin-mensajes`,
`TestLeerTrazas/hilo-por-clone3`, `TestLeerTrazas/hilo-por-clone`, `TestLeerTrazas/hilo-de-un-hilo`,
`TestLeerTrazas/fichero-sin-origen`, `TestLeerTrazas/connect-fuera-de-la-invocacion`, `TestLeerTrazas/lineas-de-senal`,
`TestLeerTrazas/cortada-por-el-tope`, `TestLeerTrazas/sin-linea-final-sin-corte`,
`TestLeerTrazas/cortada-con-llamada-interrumpida`, `TestLeerTrazas/llamada-interrumpida-sin-corte`,
`TestLeerTrazas/llamada-interrumpida-antes-del-final`,
`TestLeerTrazas/connect-publico-en-curso`,
`TestLeerTrazas/connect-ipv6-publico-en-curso`, `TestInforme/aprobado`,
`TestInforme/fuera-de-lo-grabado-no-cambia-el-veredicto`, `TestInforme/sesion-sin-terminar`,
`TestInforme/sesion-ilegible`, `TestInforme/llegada-a-la-red`, `TestInforme/sesion-cortada-con-invocaciones`,
`TestInforme/eval-sin-sesion`, `TestInforme/sin-eval-txt`, `TestInforme/eval-desconocida`,
`TestInforme/sin-pregunta-txt`, `TestInforme/traza-ilegible`, `TestLeerTrazasSinFicheros/vacio`,
`TestLeerTrazasSinFicheros/inexistente`, `TestEscribirInformeSinSusEntradas/sin-python-inexistente`,
`TestEscribirInformeSinSusEntradas/sesiones-inexistente`, `TestEscribirInformeSinSusEntradas/evals-inexistente`,
`TestEscribirInformeSinSusEntradas/destino-es-un-fichero` y `TestEscribirInformeSinSusEntradas/evals-y-sesiones-vacios`:
los nombres del inventario de tests de plan.md, del contrato de evals §6 y del contrato del job §9, los de `TestInforme`
sobre el árbol de sesiones sintéticas del contrato del job §9.1 y los de `TestEscribirInformeSinSusEntradas` y
`TestLeerTrazasSinFicheros` sobre directorios temporales.
`TestInforme/aprobado` fija además la cabecera del informe (data-model §10.3; contrato del job §5 y §9): `modelo` y
`commit` iguales a las dos constantes que el test pasa a `EscribirInforme`; `modelos_de_sesion` igual a
`["claude-haiku-4-5"]` y `versiones_de_claude_code` igual a `["2.1.270", "2.1.269"]`, las de los transcripts sintéticos;
`sin_python` igual byte a byte a `informe/sin-python.txt`; y esos mismos valores en las líneas `Modelo del job:`,
`Modelos de las sesiones:`, `Versiones de Claude Code:` y `Commit:` y en la sección `## Comprobación sin Python` de
`informe.md`; y la `respuesta` de `01-lpac-articulo-21` igual al `result` de su transcript, con `fin_de_la_sesion`
`result success`, y su pregunta y su respuesta en la sección de la sesión de `informe.md` (data-model §10.2).
`TestInforme/sesion-sin-terminar`, `respuesta` vacía y `fin_de_la_sesion` `system`, el `type` del último mensaje de su
transcript. `TestInforme/fichero-mal-formado` y `/llegada-a-la-red`, que los `motivos` de la raíz son exactamente
`02-sin-pregunta.yaml: mal formado: …` y
`01-lpac-articulo-21: petición llegada a la red: boe articulo BOE-A-2015-10565 a9998 --json → 203.0.113.7:443`, y esa
línea en los motivos de `informe.md`.
`TestInforme/sesion-ilegible`, que las dos listas quedan vacías cuando ninguna sesión se pudo leer.
`TestJuzgar/describe-y-dry-run-no-satisfacen`, que una invocación con `--describe` o `--dry-run` no satisface ningún
comando esperado; `/otra-fallida`, que las invocaciones con código 2 o 3 van a `otras_fallidas` y no a
`fuera_de_lo_grabado`.
`TestInforme/eval-sin-sesion`, que una eval bien formada sin ninguna sesión que la juzgue da veredicto `fallo` con el
motivo `11-no-activa-programacion.yaml: sin ninguna sesión`, aunque la única sesión pase; `/sin-eval-txt`,
`/eval-desconocida` y `/sin-pregunta-txt`, que un `eval.txt` que falta o no nombra ninguna eval, o un `pregunta.txt`
que falta, dejan la sesión sin pasar con `sesión ilegible: <fichero>: …`. `TestEscribirInformeSinSusEntradas`, que
`EscribirInforme` devuelve un error que nombra la ruta, sin escribir `informe.md` ni `informe.json`, si no puede leer
`SinPython`, `Sesiones` o `Evals` o no puede escribir en `Destino`, y que sin ninguna eval ni sesión escribe el informe
con veredicto `fallo` y el motivo `ninguna eval bien formada que juzgar`.

## 8. `make install` (US3, FR-050 a FR-054, SC-006)

```bash
rtk proxy sh -c 'GOMODCACHE="$(go env GOMODCACHE)" GOCACHE="$(go env GOCACHE)" HOME=/tmp/kitlegal-quickstart-h5/home GOBIN=/tmp/kitlegal-quickstart-h5/gobin GOENV=off GOPROXY=off make -C /tmp/kitlegal-quickstart-h5/repo install; echo "código $?"'
rtk proxy readlink /tmp/kitlegal-quickstart-h5/home/.claude/skills/boe-legislacion
rtk proxy readlink /tmp/kitlegal-quickstart-h5/repo/bin/instalado/kitlegal
rtk proxy /tmp/kitlegal-quickstart-h5/home/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a21 --describe | rtk proxy grep '"title"'
rtk proxy sh -c 'GOMODCACHE="$(go env GOMODCACHE)" GOCACHE="$(go env GOCACHE)" HOME=/tmp/kitlegal-quickstart-h5/home GOBIN=/tmp/kitlegal-quickstart-h5/gobin GOENV=off GOPROXY=off make -C /tmp/kitlegal-quickstart-h5/repo install; echo "código $?"'
rtk proxy ls /tmp/kitlegal-quickstart-h5/home/.claude/skills
rtk proxy mkdir -p /tmp/kitlegal-quickstart-h5/otro-home/.claude/skills/boe-legislacion
rtk proxy sh -c 'GOMODCACHE="$(go env GOMODCACHE)" GOCACHE="$(go env GOCACHE)" HOME=/tmp/kitlegal-quickstart-h5/otro-home GOBIN=/tmp/kitlegal-quickstart-h5/gobin GOENV=off GOPROXY=off make -C /tmp/kitlegal-quickstart-h5/repo install; echo "código $?"'
rtk proxy test -d /tmp/kitlegal-quickstart-h5/otro-home/.claude/skills/boe-legislacion && echo "la entrada en conflicto sigue siendo un directorio"
rtk proxy go test -tags=integration -count=1 -run '^(TestInstalacion|TestFicherosDelBinario)$' -v ./internal/skills/ | rtk proxy grep -E -- '--- (PASS|FAIL)'
```

Esperado: `código 0`; el enlace de la skill da la ruta física del clon, que termina en
`kitlegal-quickstart-h5/repo/skills/boe-legislacion` (en macOS empieza por `/private/tmp/`); el del binario,
`/tmp/kitlegal-quickstart-h5/gobin/kitlegal`; `"title": "boe articulo"` desde el binario instalado a través de
`scripts/boe`; la segunda instalación `código 0` y en `~/.claude/skills` temporal solo `boe-legislacion`; con la entrada
ajena, `código 2` y una línea `instalar-skills: conflicto: …/boe-legislacion …; no se modifica`, y la entrada sigue siendo
un directorio; los cinco guiones de `TestInstalacion` en `PASS` (`instalar`, `instalar-de-nuevo`,
`instalar-con-conflicto`, `instalar-sin-gobin`, `instalar-con-enlace-roto`) y los dos casos de `TestFicherosDelBinario`
en `PASS` (`raiz-por-un-enlace`, `dir-por-un-enlace`).

## 9. Skill: protocolo, cita y reglas (FR-001 a FR-015, FR-077, SC-004, SC-012 en su parte estática)

```bash
rtk proxy grep -nE '^(name|description|metadata|  kitlegal-applets|  kitlegal-referencias):' skills/boe-legislacion/SKILL.md
rtk proxy grep -niE 'Identificar la norma|Resolver|Leer índice y bloques|citando|ley y reglamento|variación autonómica' skills/boe-legislacion/SKILL.md
rtk proxy grep -nF '[BOE-A-2015-10565, bloque a21]' skills/boe-legislacion/SKILL.md
rtk proxy grep -niwE 'KITLEGAL_CACHE_DIR|evals?|job|GitHub Actions|claude-[a-z]+-[0-9]+' skills/boe-legislacion/SKILL.md skills/boe-legislacion/references/normas.md ; echo "fin de la búsqueda"
```

Esperado: las cinco claves del frontmatter; los seis elementos del protocolo, cada uno en su paso o regla; el ejemplo de
cita en la forma fija; y la última búsqueda solo `fin de la búsqueda`. La revisión de las cuatro reglas de SC-004
(FR-006, FR-010, FR-011, FR-015) se hace leyendo la sección `## Reglas`.

## 10. Sin Python en el producto, en los controles ni en el job (FR-013, FR-044, SC-007)

```bash
rtk proxy grep -rnE '^[[:space:]]*(exec[[:space:]]+)?(/usr/bin/env[[:space:]]+)?(python|pypy)[0-9.]*([[:space:]]|$)|^#!.*(python|pypy)' skills data evals schemas/normas.yaml.json schemas/eval.yaml.json scripts/skills-sync.sh scripts/instalar-skills.sh scripts/grabar-evals.sh scripts/evals.sh Makefile .github/workflows/evals.yml ; echo "fin de la búsqueda"
```

Esperado: solo `fin de la búsqueda`.

## 11. Directorios `skills`, atributos y pendientes (FR-083 a FR-085, SC-011)

```bash
git check-attr linguist-vendored linguist-generated -- .agents/skills/golang-how-to/SKILL.md skills/boe-legislacion/SKILL.md .agents/.gitattributes
rtk proxy git diff --stat main...HEAD -- .claude/skills skills-lock.json .agents/skills ; echo "fin del diff"
rtk proxy grep -c '^## En H5' docs/PENDIENTES.md ; echo "fin del recuento"
rtk proxy grep -nE '`skills/`|`\.agents/skills/`|`\.claude/skills/`' README.md
rtk proxy grep -nE 'make install|make skills-sync|make skills-check|make evals' README.md CONTRIBUTING.md CHANGELOG.md
rtk proxy grep -c 'formato común de eval' README.md CONTRIBUTING.md CHANGELOG.md ; echo "fin del formato"
rtk proxy grep -c 'job de evals' README.md CONTRIBUTING.md CHANGELOG.md ; echo "fin del job"
rtk proxy grep -c 'encadena los diez controles' README.md ; echo "fin de diez"
rtk proxy grep -c 'nueve controles' README.md ; echo "fin de nueve"
rtk proxy grep -c '^| `make skills-check` |.*| sí |' README.md ; echo "fin de skills-check en README"
rtk proxy grep -c '^| `make skills-sync` |.*| no |' README.md ; echo "fin de skills-sync en README"
rtk proxy grep -c '^| `make evals` |.*| no |' README.md ; echo "fin de evals en README"
rtk proxy grep -c '^| `make test-integration` |.*instalación' README.md ; echo "fin de test-integration en README"
rtk proxy grep -c '^make install .*enlaza las skills' README.md ; echo "fin de install en README"
rtk proxy grep -c '(`skills-sync`, `release`)' README.md ; echo "fin de make help en README"
rtk proxy grep -c '| `make skills-check` | sí |' CONTRIBUTING.md ; echo "fin de skills-check en CONTRIBUTING"
rtk proxy grep -c '| `make skills-sync` | no — ' CONTRIBUTING.md ; echo "fin de skills-sync en CONTRIBUTING"
rtk proxy grep -c '| `make evals` | no — ' CONTRIBUTING.md ; echo "fin de evals en CONTRIBUTING"
rtk proxy grep -c '^| Tests con la etiqueta `integration`.*instalación' CONTRIBUTING.md ; echo "fin de integration en CONTRIBUTING"
rtk proxy grep -c '| `make skills-sync` | Anuncia' CONTRIBUTING.md ; echo "fin de la tabla posterior"
rtk proxy grep -c 'desde H5, `make skills-sync`' CONTRIBUTING.md ; echo "fin del párrafo de CONTRIBUTING"
rtk proxy grep -c '\*\*H5 — skill `boe-legislacion`\*\*' CHANGELOG.md ; echo "fin de la introducción del CHANGELOG"
```

Esperado: `set` para los dos atributos en `.agents/skills/golang-how-to/SKILL.md` y `unspecified` en los otros dos; el
diff solo `fin del diff`; el recuento `0`; el `README.md` explica los tres directorios; las cuatro órdenes aparecen en los
tres documentos; en las búsquedas del formato y del job, `README.md`, `CONTRIBUTING.md` y `CHANGELOG.md` con un recuento
de 1 o más cada uno (un `:0` en cualquiera de los tres es un fallo de FR-083 y SC-011), seguido de su línea de fin; y en
las quince búsquedas siguientes, que miden «las comprobaciones de skills en `make ci`» con la documentación alineada con
el `Makefile`, `1` en la de «diez controles», `0` en la de «nueve controles», `1` en cada fila de `README.md` (`skills-check`
con «sí», `skills-sync` y `evals` con «no», `test-integration` con la instalación), `1` en el comentario de
`make install`, `0` en la frase de `make help` con `skills-sync`, `1` en cada fila de la tabla «Los controles» de
`CONTRIBUTING.md` (`skills-check` con «sí», `skills-sync` y `evals` con «no —» y su motivo, la de los tests `integration`
con la instalación), `0` en la fila de `make skills-sync` de la tabla de órdenes con contenido posterior, `1` en su
párrafo y `1` en la entrada de H5 de la introducción de *Unreleased*, cada recuento seguido de su línea de fin. Que GitHub saque `.agents/skills/` de las estadísticas y lo pliegue en los diffs es el supuesto S8 de
research.md; la tarea `[plataforma]` registra lo comprobable (esta salida) en `gates/pr-h5.md`.

## 12. Plataforma: prueba de red y ejecución de cierre (FR-070 a FR-082, SC-001 a SC-003, SC-012)

Los ejecuta la tarea `[plataforma]` con la propuesta de cambio abierta. Escriben: las etiquetas de la propuesta de cambio y
tres ficheros del directorio del hito (`gates/prueba-de-red.md`, `gates/evals-cierre.md`, `gates/aceptacion.md`).

### 12.1 Prerrequisitos de plataforma

```bash
rtk proxy gh secret list
rtk proxy gh label list --search evals
gh pr view --json number,headRefOid,labels
```

Esperado: `CLAUDE_CODE_OAUTH_TOKEN` (token de suscripción de `claude setup-token`) en los secretos; las etiquetas `evals` y `evals-prueba-de-red`; la propuesta de cambio de la
rama. Si falta algo, la tarea se detiene y lo anota (lo da de alta una persona).

**Qué ejecución se lee.** Cada ejecución se identifica por la etiqueta que la dispara, nunca como «la última» de la
lista, que puede ser la de un intento anterior (misma rama, mismo evento, ya terminada) o la de otra etiqueta. Por eso:
(1) antes de poner la etiqueta, si está puesta, se quita (solo si `gh pr view` la lista, así que la orden no depende de
lo que haga `gh` al quitar una etiqueta que no está), de modo que siempre hay un evento `labeled` nuevo y una ejecución
nueva (contrato del job §7); (2) cada orden que lee una ejecución toma el instante del **último** evento `labeled` de
esa etiqueta en la propuesta de cambio (API de eventos de la incidencia), ordenando los instantes en lugar de fiarse del
orden en que la API los devuelve, y elige, de las ejecuciones de la rama con evento `pull_request` cuyo `workflowName`
es `evals` (el `name:` del job, contrato del job §1), la **primera creada en ese instante o después**, por su
`databaseId`. Ninguna orden usa `gh run list --workflow evals.yml`: que `gh` resuelva por el nombre de su fichero un
flujo que hasta la fusión solo está en la rama de la propuesta de cambio no está verificado (research.md V14), y el
campo `workflowName` sí lo expone `gh run list --json` (V45). Cada orden repite esa búsqueda entera, de modo que se
ejecuta tal cual y vale igual si la sesión se retoma. Comportamiento de la plataforma que esto supone (formato de los
instantes, instante de la ejecución, ejecución nueva al volver a poner la etiqueta, valor de `workflowName`):
research.md D22, S12, que la prueba de red registra en `gates/prueba-de-red.md` antes de leer ningún informe.

### 12.2 Prueba de red (SC-012)

```bash
rtk proxy sh -c 'set -e; puestas=$(gh pr view --json labels --jq ".labels[].name"); if printf "%s\n" "$puestas" | grep -qxF evals-prueba-de-red; then gh pr edit --remove-label evals-prueba-de-red; fi; echo "la etiqueta evals-prueba-de-red no está puesta"'
gh pr edit --add-label evals-prueba-de-red
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals-prueba-de-red\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; echo "etiqueta puesta: $t"; gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,headSha,status,conclusion,createdAt,url --jq "map(select(.createdAt >= \"$t\")) | sort_by(.createdAt) | {posteriores_a_la_etiqueta: map({databaseId, workflowName, createdAt}), evals: (map(select(.workflowName == \"evals\")) | .[0] // \"todavía no hay ninguna ejecución de evals posterior a la etiqueta\")}"'
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals-prueba-de-red\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; id=$(gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; set +e; gh run watch "$id" --exit-status; echo "código $?"'
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals-prueba-de-red\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; id=$(gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; test "$(gh run view "$id" --json status --jq .status)" = completed; registro=$(gh run view "$id" --log); for f in informe.md informe.json; do printf "%s\n" "$registro" | grep -qF -- "--- inicio de $f ---"; printf "%s\n" "$registro" | grep -qF -- "--- fin de $f ---"; done; for f in informe.md informe.json; do printf "%s\n" "$registro" | sed -n "/--- inicio de $f ---/,/--- fin de $f ---/p"; done'
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals-prueba-de-red\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; id=$(gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; test "$(gh run view "$id" --json status --jq .status)" = completed; registro=$(gh run view "$id" --log); m="la retirada de Python"; printf "%s\n" "$registro" | grep -qF -- "--- inicio de $m ---"; printf "%s\n" "$registro" | grep -qF -- "--- fin de $m ---"; printf "%s\n" "$registro" | sed -n "/--- inicio de $m ---/,/--- fin de $m ---/p"'
rtk proxy sh -c 'set -e; puestas=$(gh pr view --json labels --jq ".labels[].name"); if printf "%s\n" "$puestas" | grep -qxF evals-prueba-de-red; then gh pr edit --remove-label evals-prueba-de-red; fi; echo "la etiqueta evals-prueba-de-red no está puesta"'
```

Solo si la quinta o la sexta fallan con la ejecución ya terminada, porque falta alguna marca (el paso «Retirar Python
del runner» o una comprobación previa del guion detuvieron el job antes del informe):

```bash
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals-prueba-de-red\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; id=$(gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; test "$(gh run view "$id" --json status --jq .status)" = completed; gh run view "$id" --log-failed'
```

La tercera orden se repite hasta que su campo `evals` muestre una ejecución (la plataforma tarda unos segundos en
crearla): mientras no la haya, escribe «todavía no hay ninguna ejecución de evals posterior a la etiqueta»; su `url` es
el enlace que se registra. Su campo `posteriores_a_la_etiqueta` lista cada ejecución de la rama creada desde la
etiqueta con su `workflowName`, y se registra para el supuesto S12 (6): si ahí aparece una ejecución posterior a la
etiqueta con un `workflowName` vacío o distinto de `evals` mientras `evals` sigue sin ninguna, el supuesto difiere y la
tarea se detiene y lo anota. La cuarta espera a que esa ejecución termine; si la herramienta la corta por su tope de
tiempo antes de que termine, se repite tal cual (sobre una ejecución ya terminada vuelve enseguida con su resultado).
La quinta solo lee el registro de una ejecución terminada: comprueba que están las cuatro marcas del contrato del job
§3.3 y después imprime `informe.md` e `informe.json` enteros, cada uno de su marca de inicio a su marca de fin, sin
tope de líneas; si la ejecución aún no ha terminado o falta cualquiera de las marcas, falla sin imprimir nada
(research.md V48). Las líneas salen tal como las da `gh run view --log`, con lo que esa orden ponga delante de cada una.
La sexta lee del mismo registro, con la misma forma y de marca a marca, la salida del paso «Retirar Python del runner»
(contrato del job §1), que no consulta ni purga paquetes: la línea `búsqueda: find / ( -path /proc -o -path /sys ) …`,
una línea `retirado: <ruta>` por cada fichero o instalación retirados y `búsqueda tras retirar: ninguno`. Las marcas las
compone el paso al ejecutarse, así que la orden no casa con el texto del paso aunque el registro lo reproduzca
(research.md V56 y V60). Si la quinta o la sexta fallan
sobre una ejecución terminada porque falta una marca, el job se detuvo antes del informe: la tarea registra lo que
imprime la orden de `--log-failed` que va tras el bloque, se detiene y lo anota con la ruta y el mensaje (`queda Python:
<ruta>`, `evals: hay Python accesible: <ruta>`, el error de `rm` o de `find`), y el arreglo va en una tarea nueva antes
de la ejecución de cierre (plan.md, obligación 1). La salida de la sexta y el `sin_python` del informe —`búsqueda: find /
( -path /proc -o -path /sys ) …`, `usuario: root` y `resultado: ninguno`— se registran en `gates/prueba-de-red.md` como
evidencia de los supuestos S2 (que `sudo` y la búsqueda de GNU findutils terminaron en el runner: la línea `búsqueda:`
seguida de las `retirado:` y de `búsqueda tras retirar: ninguno`) y S7 (research.md D22: qué intérpretes, bibliotecas e
instalaciones traía la imagen, cada uno por su línea `retirado:`), junto a lo que las sesiones ejecutaron con `claude` y
con el binario.
Esperado en el informe: en «fuera de lo grabado», dos filas con sesión `01-lpac-articulo-21-prueba-de-red` y eval
`01-lpac-articulo-21.yaml`, las dos invocaciones de `a9998`, con código 5 la que va sin `--offline` y 4 la que lo lleva
(research.md V41); ninguna entre los comandos ejecutados de esa sesión; en la sección de esa sesión, la invocación de
`a9998` sin `--offline` con una sola conexión, `127.0.0.1:9` de clase `local` (la del proxy que rechaza: la traza real de
research.md V53 tiene tres `connect` a esa dirección, y el informe presenta la pareja una vez), y la que lleva
`--offline` sin ninguna conexión; ninguna sesión con el motivo `sesión ilegible`, salvo, si la hubiera, una que el tope
cortó (código de la sesión 124 o 137); «ninguna petición llegó a la red de una fuente»; y el resultado de esa sesión igual
que el de `01-lpac-articulo-21`. Se registra en `gates/prueba-de-red.md` con el enlace a la ejecución, y con esa conexión
`local` y la ausencia de sesiones no cortadas (código distinto de 124 y 137, como el 0 de las terminadas) con traza
ilegible como evidencia del supuesto S4 (research.md D22): las líneas `connect`, las de creación de hilos y procesos, las
de señal y la atribución por hilos funcionan con las trazas reales del runner. Cada sesión cortada por el tope, legible o
no, se registra como evidencia contra el supuesto S9, con su motivo, y no cuenta para el S4: su traza puede quedar sin
las líneas finales de los procesos vivos o con una llamada interrumpida (research.md V54). Si falta la conexión o alguna
sesión no cortada es ilegible por su traza, la tarea se detiene y lo anota, con el fichero, la línea y el texto que el
informe da en el motivo.

### 12.3 Ejecución de cierre y aceptación (FR-080 a FR-082, SC-001 a SC-003)

> **Enmienda del 2026-09-16 (ADR 0016).** Las órdenes no cambian: la ejecución se sigue identificando por el último
> evento `labeled` de `evals` y se sigue leyendo el informe entre sus marcas. Lo que cambia es lo esperado. El modelo del
> informe es `claude-sonnet-5`, el que decide, y `claude-haiku-4-5-20251001` aparece en `modelos_informativos`. «Las diez
> positivas y las dos de no activación en verde» se lee ahora en la sección `## Tasas por eval`: las doce series del
> modelo que decide sobre evals que no son informativas tienen que llegar al umbral (2 de 3). Las cinco series
> informativas (evals 13 a 17) y las doce del modelo informativo se publican con su tasa y **no** hacen fallar la
> ejecución: que las informativas estén en rojo es el estado esperado hasta que llegue del backlog la herramienta que
> busca dentro de una norma el artículo que trata una materia. Sigue exigiéndose «ninguna petición llegó a la red de una
> fuente». Se anota además el tiempo del paso «Ejecutar las evals», que es la medida del presupuesto del contrato del
> job §1.

La ejecución se identifica igual que en §12.2, con la etiqueta `evals`. Si está puesta (p. ej. porque un intento
anterior la dejó), se quita antes de ponerla: sin evento `labeled` nuevo no arranca ninguna ejecución, y las órdenes que
siguen no encontrarían ninguna posterior a la etiqueta en lugar de leer la vieja.

```bash
rtk proxy sh -c 'set -e; puestas=$(gh pr view --json labels --jq ".labels[].name"); if printf "%s\n" "$puestas" | grep -qxF evals; then gh pr edit --remove-label evals; fi; echo "la etiqueta evals no está puesta"'
gh pr edit --add-label evals
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; echo "etiqueta puesta: $t"; gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,headSha,status,conclusion,createdAt,url --jq "map(select(.createdAt >= \"$t\")) | sort_by(.createdAt) | {posteriores_a_la_etiqueta: map({databaseId, workflowName, createdAt}), evals: (map(select(.workflowName == \"evals\")) | .[0] // \"todavía no hay ninguna ejecución de evals posterior a la etiqueta\")}"'
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; id=$(gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; set +e; gh run watch "$id" --exit-status; echo "código $?"'
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; id=$(gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; test "$(gh run view "$id" --json status --jq .status)" = completed; registro=$(gh run view "$id" --log); for f in informe.md informe.json; do printf "%s\n" "$registro" | grep -qF -- "--- inicio de $f ---"; printf "%s\n" "$registro" | grep -qF -- "--- fin de $f ---"; done; for f in informe.md informe.json; do printf "%s\n" "$registro" | sed -n "/--- inicio de $f ---/,/--- fin de $f ---/p"; done'
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; id=$(gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; test "$(gh run view "$id" --json status --jq .status)" = completed; c=$(gh run view "$id" --json headSha --jq .headSha); test -n "$c"; echo "commit evaluado: $c"; cambios=$(git diff --name-only "$c" HEAD); echo "ficheros cambiados entre el commit evaluado y la cabeza:"; printf "%s\n" "$cambios"; printf "%s\n" "$cambios" | while IFS= read -r f; do case "$f" in "") ;; specs/006-h5-skill-boe-legislacion/*) ;; *) echo "fuera del directorio del hito: $f"; exit 1 ;; esac; done; echo "todos bajo specs/006-h5-skill-boe-legislacion/"'
```

Solo si la quinta falla con la ejecución ya terminada, porque falta alguna marca (el paso «Retirar Python del runner» o
una comprobación previa del guion detuvieron el job antes del informe):

```bash
rtk proxy sh -c 'set -e; n=$(gh pr view --json number --jq .number); marcas=$(gh api --paginate "repos/{owner}/{repo}/issues/$n/events?per_page=100" --jq ".[] | select(.event == \"labeled\" and .label.name == \"evals\") | .created_at"); t=$(printf "%s\n" "$marcas" | sort | tail -n 1); test -n "$t"; id=$(gh run list --branch h5-skill-boe-legislacion --event pull_request --limit 100 --json databaseId,workflowName,createdAt --jq "map(select(.workflowName == \"evals\" and .createdAt >= \"$t\")) | sort_by(.createdAt) | .[0].databaseId // empty"); test -n "$id"; test "$(gh run view "$id" --json status --jq .status)" = completed; gh run view "$id" --log-failed'
```

Con ella, la ejecución de cierre no vale: la tarea registra su salida en `gates/evals-cierre.md`, se detiene y lo anota
con la ruta y el mensaje, como en §12.2.

La tercera orden se repite hasta que muestre la ejecución y la cuarta hasta que termine, como en §12.2; la quinta
imprime el informe entre sus marcas, igual que allí. La última es la evidencia de SC-003 y lo que se registra de ella es
exactamente lo que imprime (contrato del job §7): el commit evaluado, la lista **completa** de
`git diff --name-only <commit evaluado> HEAD` y, solo si cada línea de esa lista empieza por
`specs/006-h5-skill-boe-legislacion/`, la línea `todos bajo specs/006-h5-skill-boe-legislacion/`; en cuanto una no
empieza así, escribe `fuera del directorio del hito: <fichero>` y termina con 1. No descarta ningún error: un fallo de
`git diff` (p. ej. un commit que no está en el clon) detiene la orden antes de escribir la lista (research.md V49).

Esperado: `código 0`; veredicto `aprobado` con las diez positivas y las dos de no activación en verde, todas con la
sesión terminada, modelo `claude-haiku-4-5-20251001`, el commit evaluado, «ninguna petición llegó a la red de una
fuente» y, en la comprobación sin Python (`sin_python` de `informe.json`), exactamente las tres líneas del contrato del
job §3.1: `búsqueda: find / ( -path /proc -o -path /sys ) -prune -o …` (la búsqueda como root en todo el sistema de
ficheros), `usuario: root` y `resultado: ninguno`; y la última orden `commit evaluado: <sha>`, `ficheros cambiados entre el commit evaluado y la cabeza:`, la
lista (vacía o solo con ficheros bajo `specs/006-h5-skill-boe-legislacion/`) y
`todos bajo specs/006-h5-skill-boe-legislacion/`. Se registran `gates/evals-cierre.md` (el informe que imprime la quinta
orden y la salida entera de la última, tal cual) y `gates/aceptacion.md` (respuestas e invocaciones de las sesiones
`01-lpac-articulo-21` y `06-irpf-rendimientos-del-trabajo`, y el `sin_python` del informe como constancia de cómo se
comprobó que no había Python, FR-081) (contrato del job §7). Si
alguna sesión no terminó porque la cortó el tope (código 124 o 137), la ejecución de cierre no vale: la tarea lo anota en
`gates/evals-cierre.md` como evidencia contra el supuesto S9 (research.md D22) y se detiene.

## Limpieza

```bash
rtk proxy chmod -R u+w /tmp/kitlegal-quickstart-h5
rtk proxy rm -r /tmp/kitlegal-quickstart-h5
rtk proxy git status --porcelain | rtk proxy grep -vE '^.. specs/006-h5-skill-boe-legislacion/' ; echo "fin del estado"
```

Esperado: la carpeta desaparece y el estado del árbol muestra solo `fin del estado`.
