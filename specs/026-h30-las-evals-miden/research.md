# Research: H30 · Las evals miden solo lo que cambia

Fase 0 del plan. Cada decisión lleva la alternativa rechazada y por qué (constitución, «Criterio de decisión autónoma»,
3); cada afirmación sobre una herramienta externa dice dónde se comprobó en esta sesión (V*) o figura como supuesto no
verificado (S*). Ninguna duda queda abierta: las cinco del spec las cerró `clarify`
(spec.md, «Clarifications»).

## Verificaciones hechas en esta sesión

- **V1 · `go list -json`.** `go help list` (Go 1.27.3, el de `go.mod`, `toolchain go1.27.3`) documenta en el struct
  `Package` los campos que usa el cálculo: `ImportPath`, `Dir`, `Standard`, `Module` (con `Path` y `Version`),
  `GoFiles` («excluding CgoFiles, TestGoFiles, XTestGoFiles»), `EmbedFiles`, `Imports`, `TestImports`, `XTestImports`,
  `Error` y `DepsErrors`; y que con `-e` la orden no imprime los errores por la salida de error, sino en esos campos.
- **V2 · `go list` con la plataforma del job.** Ejecutado en esta sesión:
  `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go -C <repo> list -tags evals -deps -e -f '…' ./...` termina sin error y da,
  sin red, todos los paquetes del módulo y sus dependencias con su módulo y versión (p. ej.
  `modernc.org/sqlite@v1.60.1`, `github.com/alecthomas/kong@v1.16.1`); los del módulo, con sus ficheros: el paquete
  raíz (`skills.go`) empotra 8 patrones de `skills/`, `data` empotra 25 ficheros de `data/`, y `internal/app` tiene 17
  ficheros compilados. `./...` incluye `internal/workflow`, `internal/source/bocm` e `internal/source/boe/fiestas`, que
  el binario no enlaza. De los externos, `golang.org/x/sys/unix` y `github.com/segmentio/asm/cpu/x86` dependen de la
  plataforma: por eso se fija (D3).
- **V3 · `go mod edit -json`.** `go help mod edit`: «usage: go mod edit [editing flags] [-fmt|-print|-json] [go.mod]» y
  «a different target file can be specified after the editing flags»; `-json` imprime el `go.mod` en JSON. Sirve para
  leer el `go.mod` de la base desde un fichero temporal sin tocar el del árbol.
- **V4 · `encoding/json/v2` y los miembros desconocidos.** `go doc encoding/json/v2`: «by default both v1 and v2 ignore
  unknown names» y `RejectUnknownMembers` es la opción que los rechaza. Un informe de `main` escrito con campos que este
  hito no conoce se decodifica igual (D8).
- **V5 · La tasa de referencia.** Un programa de Go en `$TMPDIR` (fuera del repositorio, borrable, sin dependencias)
  calculó la regla de FR-024 con la cola de la binomial en coma flotante: `boe-legislacion` (`umbral: 0.028`, 120
  respuestas, límite de hoy 3) da *p*ref = 0,0085 con P(Bin(120; 0,0085) > 3) = 0,01970 y límite k(120) = 3;
  `jurisprudencia` (`umbral: 0.017`, 60 respuestas, límite 1) da *p*ref = 0,0036 con P = 0,01998 y k(60) = 1; el mutante
  de SC-010 (`umbral: 0.017` en `boe-legislacion`, límite 2 de 120) da *p*ref = 0,0047 y k(120) = 2. Con menos
  respuestas, la regla admite más que la proporción: con 10 respuestas, k = 1 en las dos skills, donde la proporción de
  hoy admite 0. Las cifras coinciden con las del spec (Q4: ≈0,85 % y ≈0,36 %).
- **V6 · git.** `git ls-files -h`: `-c, --cached`, `-o, --others`, `--exclude-standard` y `-z`; `git merge-base -h`:
  «git merge-base [-a | --all] <commit> <commit>...».
- **V7 · El repositorio** (grep y lecturas de esta sesión):
  - Tablas de comandos (entre `<!-- inicio de la tabla de comandos` y `<!-- fin de la tabla de comandos -->`, constantes
    de `internal/skills/comandos.go:18-21`): `boe-legislacion` usa `boe` y `graph`; `jurisprudencia`, `cita`;
    `legal-core`, `plazos` y `territorio`.
  - Cada applet de `internal/app` devuelve su nombre con un literal: `func (appletBoe) Nombre() string { return "boe" }`
    (`boe.go:87`), y así `cita.go:53`, `grafo.go:76` (`graph`), `instalacion.go:115` (`skills`), `mcp.go:58`,
    `plazos.go:104` y `territorio.go:51`.
  - Menciones entre skills: `legal-core` nombra `boe-legislacion` (p. ej. línea 12) y no `jurisprudencia`;
    `jurisprudencia` nombra `boe-legislacion` (línea 202) y no `legal-core`; `boe-legislacion` no nombra ninguna.
  - La cabecera de cada referencia es `<!-- generado desde data/<fichero>, no editar -->`
    (`internal/skills/referencias.go:146`).
  - Las grabaciones que el arnés sirve a toda sesión son las tres raíces de `internal/evals/grabaciones.go:18-35`
    (`GrabacionesDeH4`, `GrabacionesDeH5`, `GrabacionesDeH9`), más `GrafosPrevios/<nombre>` de la eval con
    `grafo_previo`; `prepararLaCache` (`preparar.go:344`) junta las consultas de todas las evals de la skill.
  - Todos los umbrales de las respuestas que deciden tienen umbral 0 (`umbrales.go:14-66`) salvo `<clase>:<modelo>`,
    que compara `medida/total ≤ clase.Umbral` (`umbralesConjuntos`, `umbrales.go:540`; `compararUmbral`, 1094).
  - `scripts/workflow/informe.sh:264-310` rehace esa comparación con `medida/total` y `umbral` y marca INCOHERENTE el
    umbral cuyo `cumple` no casa; `scripts/workflow/serie.sh fila` (93-190) lee `.evals[]`, `.juez.respuestas`,
    `.juez.sin_juzgar`, `.sesiones_sin_medir`, `.motivos` y los umbrales `<clase>:<modelo>` y
    `<clase>:<modelo>:<modo>` por su nombre exacto, y falla si su `medida`/`total` no casan con su recuento.
  - `scripts/workflow/comun.sh:205-213` (`informe_del_trabajo`) extrae el informe del registro de un trabajo con
    `gh run view <ejecución> --job <trabajo> --log`, quitando trabajo, paso y hora, y la marca de orden de bytes de cada
    trozo; lo usa el cierre desde H7.3.
  - `internal/evals/tanda.go` consulta la plataforma con órdenes constantes por `sh -c` y cada valor en una variable del
    entorno (gosec G204), y su test usa un `gh` sustituto (`tanda_test.go:19`).
  - `release_test.go` (raíz) decodifica cada trabajo de `release.yml` en estricto (`KnownFields`) con
    `trabajoDelFlujo` (`needs`, `runs-on`, `permissions`, `env`, `steps`); `publicar` publica la release, el cask y el
    manifiesto del bucket con goreleaser, y `catalogo` (detrás de `humo`) publica el catálogo.
  - Ningún flujo de `.github/workflows/` usa `actions/upload-artifact` ni `actions/cache`.
- **V8 · Los módulos de lo común.** Un programa de Go en `$TMPDIR` ejecutó el `go list` de V2 en JSON y recorrió el
  grafo desde los paquetes de lo común (D4), con las importaciones de los tests de `internal/evals`: sin entrar en los
  paquetes del módulo que no son de lo común alcanza 26 módulos externos, y sin cortar, los mismos 26.

## Supuestos no verificados (sin red en esta sesión)

`gh` no se puede ejecutar en el sandbox de esta sesión (no lee su configuración), y la API de GitHub no está en local.

- **S1.** `gh run list --json` admite `headSha`, `createdAt` y `conclusion` además de `databaseId` y `status` (los dos
  que ya usa `tanda.go`), y lista las ejecuciones de la más reciente a la más antigua. El código ordena igual por
  `createdAt` y `databaseId`, sin depender del orden de `gh`.
- **S2.** `GET repos/{owner}/{repo}/pulls?state=closed&base=main` da de cada propuesta `number`, `merged_at` (nulo si
  se cerró sin fusionar) y `head.sha`, que para una fusionada es la cabeza al fusionarla.
- **S3.** `GET repos/{owner}/{repo}/pulls/{n}/files` da `previous_filename` en un fichero renombrado (el campo
  `filename` lo usa hoy el trabajo `cambios`, línea 73 de `evals.yml`) y compara la cabeza con la base de fusión.
- **S4.** `actions/checkout@v7` con `fetch-depth: 0` trae la historia de todas las ramas, de modo que el commit
  `pull_request.base.sha` está en el clon (`release.yml` usa ya `fetch-depth: 0`, para otra cosa).
- **S5.** La plataforma conserva el registro de una ejecución 90 días desde su `createdAt`, la vida que da el spec
  («Assumptions»): el de una medida de menos de 90 días se puede leer. Si el repositorio la acortara, la orden que lee
  el registro fallaría y la búsqueda pararía con su error escrito (D17); nada se reutilizaría de más.
- **S6.** `gh workflow run evals.yml --ref <etiqueta>` despacha el flujo sobre el commit de una etiqueta.
- **S7.** En un evento `pull_request` (también `labeled`), `github.event.pull_request.base.sha` es el commit de la rama
  base; y en el `push` de una etiqueta, `github.ref_name` es su nombre (solo va al mensaje de la release).
- **S8.** El `--jq` de `gh` (gojq) admite `//` y `@tsv`; hoy el repositorio solo usa con él filtros simples
  (`.[].filename`, `.url`, `.started_at`).

Lo que el repositorio ya usa en producción no es supuesto: `gh api --paginate` de los ficheros de una propuesta
(`evals.yml:73`), `gh run view --json jobs` y `gh run list --commit` (la tanda), `gh run view --job --log` (el cierre),
`${{ !cancelled() }}`, `$RUNNER_TEMP` y los permisos `actions: read` y `pull-requests: read` (`evals.yml`).

Si alguno no se cumple, el código no reutiliza (la eval se mide, FR-013, y la orden que falló queda escrita con su
error en el registro y en el informe, D17) o la release falla con su mensaje: ninguno puede hacer que se reutilice un
resultado que no es de su huella.

## Decisiones

### D1 · Dónde vive: el arnés de las evals (`internal/evals`)

**Decisión.** Todo el cálculo —conjuntos, elección, huella, búsqueda de medidas de `main`, decisión de cada trabajo,
límite de la tasa y comprobación de la release— vive en `internal/evals`, el arnés, que el binario no enlaza y que ya es
lo común (FR-002). Los puntos de entrada del job son tests con la etiqueta `evals`, como `TestTandaDelCommit` y
`TestEjecucionDelJob`: `TestEleccionDeLaPropuesta`, `TestDecisionDelTrabajo` y `TestSuiteCompletaDelCommit`. Dos
lecturas de formatos que ya posee `internal/skills` se exportan allí (D11).

**Alternativas.** Un ejecutable en `cmd/`: añadiría un paquete `main` a la composición del binario (lo común) y otra
superficie que nadie más usa. Guiones de shell: el cálculo de `go list`, el texto canónico de la huella y la binomial no
caben con fiabilidad en `jq`, y `scripts/workflow/` está vedado (FR-027).

### D2 · El conjunto de una eval son entradas, no una lista de ficheros

**Decisión.** El conjunto de FR-001 se calcula sobre la cabeza y se guarda como entradas de tres clases: *fichero* (el
de la eval, cada `SKILL.md`, cada `data/*.yaml` de una cabecera, cada fichero empotrado), *directorio* (las
`references/` de cada skill, las raíces de grabaciones, la carpeta `juez/`) y *paquete* (el directorio de un paquete Go:
sus `.go` que no son `_test.go`). Una ruta cambiada pertenece a un conjunto si casa con alguna entrada, exista o no en la
cabeza: así una referencia renombrada o borrada sigue siendo de su skill (FR-007), y un `.go` borrado, de su paquete.
Para la huella, cada entrada se expande a los ficheros que existen en la cabeza.

**Alternativa.** Calcular los conjuntos en la base y en la cabeza y unirlos: pide un segundo árbol de trabajo y otro
`go list`, para el mismo resultado en todo lo que una propuesta real cambia.

### D3 · Los paquetes de un applet: `go/parser` sobre `internal/app` y un solo `go list`

**Decisión.** El fichero de cada applet es el de `internal/app` cuyo método `Nombre()` devuelve su nombre como literal
(V7); sus importaciones del módulo (con `go/parser`, solo importaciones) son el punto de partida de FR-001(d). Una sola
orden, `go list -tags evals -deps -json -e ./...` desde la raíz, con `GOOS=linux`, `GOARCH=amd64` y `CGO_ENABLED=0` en
su entorno, da el grafo entero; el cierre de cada applet se recorre en Go con `Imports`. Del cierre entran los
`GoFiles` y `EmbedFiles` de cada paquete del módulo que no es de lo común ni el paquete raíz (FR-001(d): «Lo común no
entra por (d)»; el raíz «llega a un conjunto solo por (b)»), y el módulo y la versión de cada paquete externo. Un
`Error` o `DepsErrors` en un paquete del cierre es la causa de FR-005 de las evals que lo alcanzan. La plataforma
fijada es la del runner del job y hace la huella independiente de dónde se calcula (FR-009, V2).

**Alternativas.** El registro de applets en tiempo de ejecución con reflexión: da el tipo, no el fichero, y obliga a
enlazar `internal/app` en el arnés. Un `go list` por applet: el mismo grafo varias veces.

### D4 · Lo común y las rutas que miden las evals

**Decisión.** Lo común (FR-002) es una lista fija de rutas en el código, la del spec: `cmd/`, `internal/app/`,
`internal/cli/`, `internal/httpx/`, `internal/cache/`, `internal/mcp/`, `internal/evals/` (entero: sus tests son el
arnés), `skills.go`, `Makefile`, `.github/workflows/evals.yml`, los guiones que ejecuta el job (`scripts/evals.sh`,
`scripts/evals-sesion.sh`, `scripts/evals-voto.sh` y `scripts/evals-medir-juez.sh`) y los esquemas que lee el arnés
(`schemas/eval.yaml.json`, `schemas/expresiones-prohibidas.yaml.json`, `schemas/juez-clases.yaml.json`); más `go` y
`toolchain` de `go.mod` y los módulos de los que dependen los paquetes de lo común, con las importaciones de los tests
de `internal/evals`, que son el arnés. Ese cierre no entra en un paquete del módulo que no es de lo común: lo que usa un
applet es de su conjunto (FR-001(d)), y lo que usa lo común, de lo común. Hoy da los 26 módulos del binario y del arnés
(V8), así que cualquier cambio de versión elige las 45; la rama de FR-004 que elige solo las evals de cuyo conjunto
cambia un módulo se aplicará a un módulo que solo use un applet.

**Alternativa rechazada (módulos).** El cierre de `go list -deps` sin cortar: como `internal/app` compone todos los
applets, todo módulo de un conjunto sería de lo común por construcción, y esa rama de FR-004 no podría darse nunca.

Las rutas que miden las evals (FR-004) son las de hoy más las que añade el spec. Solo quedan fuera, por FR-001(d), los
`_test.go` y `testdata/` de los paquetes que entran en algún conjunto por (d), salvo las raíces de grabaciones de
FR-001(e). Los de un paquete de lo común son rutas de lo común (FR-002 nombra `cmd/`, `internal/app`, `internal/cli`…
enteros), y los de un paquete del módulo que no es de ningún conjunto ni de lo común (p. ej. `internal/render`), rutas
sin asignar (FR-004), con su línea `ruta` en la huella de todas
([contracts/eleccion-y-huella.md](contracts/eleccion-y-huella.md) §4): una propuesta que solo cambia un test de
`internal/cli` o un guion de `internal/app/testdata/script/` elige las 45.

### D5 · La base de la propuesta, en local; las rutas cambiadas, de la plataforma

**Decisión.** Las rutas cambiadas son las de `GET pulls/{n}/files` (FR-007), con `filename` y `previous_filename`. Lo
que la elección necesita de la base —la descripción de cada `SKILL.md` que cambia (FR-003), los directorios de
`skills/` y `evals/` (FR-004) y el `go.mod` (FR-004)— se lee en local de la base de fusión, `git merge-base
<pull_request.base.sha> HEAD`, con el checkout entero (S4): `git show <base>:<ruta>` y `git ls-tree --name-only`. El
`go.mod` de la base se lee con `go mod edit -json` sobre un fichero temporal (V3).

**Alternativa.** Pedir a la API el contenido de cada fichero en la base: una petición por `SKILL.md` y por `go.mod`, y
otro cliente de la plataforma que probar.

### D6 · La huella: sha256 de un texto canónico versionado

**Decisión.** La huella de una eval es el sha256 en hexadecimal (64 caracteres) de un texto con una línea por
ingrediente de FR-009, ordenado por bytes, que empieza por `huella-de-eval v1`: modelo que decide, Claude Code de las
sesiones, juez (modelo, Claude Code de sus votos y sha256 de `rubrica.md`, `casos.yaml` y `medida.json`) si la skill lo
tiene, `go` y `toolchain`, cada módulo de su conjunto y de lo común con su versión, la descripción de cada skill (su
sha256) y cada ruta de su conjunto, de lo común y sin asignar con el sha256 de su contenido. Los ficheros son los de
`git ls-files -z --cached --others --exclude-standard` (los del árbol sin los ignorados) y las rutas, relativas a la
raíz: no depende del runner, de la fecha, del orden de recorrido ni del directorio (FR-009). Los modelos y las versiones
llegan por banderas desde el `env` del job.

**Alternativas.** Los id de blob de `git ls-tree`: atan la huella al formato de objetos de git y a que el árbol esté
commiteado, y `make ci` corre sobre un árbol con cambios. El hash del repositorio entero: cambiaría con cualquier
propuesta, y ninguna eval se reutilizaría.

### D7 · `go.mod` y `go.sum`: las versiones de cada conjunto y el resto, de todas

**Decisión.** FR-009 pone en la huella «las versiones de Go y de los módulos externos de su conjunto y de lo común»: la
de cada eval lleva `go`, `toolchain` y la versión de cada módulo de su conjunto y de lo común. Lo demás de `go.mod` y
`go.sum` —que son rutas que miden y no son de ningún conjunto ni de lo común (FR-004)— entra en la de todas como ruta sin
asignar, sin lo que ya entra por las versiones: el sha256 del `go.mod` de `go mod edit -json` sin `Go`, `Toolchain` ni
las entradas de `Require` de esos módulos (con `Replace`, `Exclude`, `Godebug`, `Tool`…), y el de las líneas de
`go.sum` cuyo módulo no es de ningún conjunto ni de lo común. Así, cambiar la versión de un módulo cambia la huella de
las evals que lo tienen y de ninguna más (la elección de FR-004), y un `replace`, un `exclude`, un `godebug` o una línea
de `go.sum` de otro módulo cambian la de todas (FR-004 las elige todas).

**Alternativas.** Solo las versiones: un `replace` o un `godebug` cambiarían el binario sin cambiar ninguna huella, y una
propuesta posterior podría reutilizar un resultado medido antes de ese cambio. Los dos ficheros enteros en todas: subir
un módulo que solo usa un applet cambiaría todas las huellas, y la elección por conjunto de FR-004 no ahorraría nada.

**Lo que queda.** Un cambio de `go.sum` que solo añade líneas de hash de un módulo de un conjunto o de lo común en la
misma versión elige las 45 (última rama de FR-004) sin cambiar ninguna huella: esas evals se miden porque están
elegidas, y nada de lo que las mide ha cambiado (el código de un módulo en una versión es fijo).

### D8 · Las medidas de `main`: ejecuciones sobre cabezas fusionadas, validadas por los pasos

**Decisión.** Cada trabajo `evals (<skill>)` que reutiliza busca él mismo, solo para las evals no elegidas de su
skill: lista las ejecuciones del flujo (`gh run list --workflow evals.yml --limit 200`, S1) y las propuestas cerradas
hacia `main` (`gh api` de `pulls?state=closed&base=main&sort=updated&direction=desc&per_page=100`, S2); una medida de
`main` es una ejecución terminada, no cancelada, cuyo `headSha` es la cabeza de una fusionada (FR-010) y en la que cada
trabajo `evals (<skill>)` terminó en verde o en rojo con su paso nuevo «El informe del trabajo está escrito» en
`success` (`gh run view --json jobs`, la orden de la tanda), y de menos de 90 días. Se recorren de la más reciente a la
más antigua por `createdAt`, como mucho 10 (FR-012), y de cada una se lee solo el informe de la propia skill (la
extracción de `comun.sh`, V7); un informe que no se puede sacar del registro deja la medida sin aportar nada, y una
orden de `gh` que falla para la búsqueda con su error escrito (D17): en los dos casos, lo que queda se mide (FR-013).

**Vida de un resultado.** Un resultado de `main` solo se reutiliza si la ejecución en que se abrieron sus sesiones —su
procedencia, que una reutilizada copia tal cual (FR-011)— tiene menos de 90 días, la vida de su registro (spec, «Uso» y
«Assumptions»; FR-012). Sin esa cota, un resultado de huella estable pasaría de cada medida a la siguiente sin límite y
el informe citaría una procedencia cuyo registro ya no se puede abrir. La comparación es con el instante de la decisión,
que la recibe como dato.

**Alternativa rechazada (vida).** Volver a abrir el registro de la procedencia para comprobar que sigue: una lectura
más por eval reutilizada, de hasta 922 KiB, para saber lo que la fecha ya dice.

**Alternativas.** Un trabajo central que lee los tres informes y los pasa a los de cada skill: los resultados no caben
con seguridad en las salidas de un trabajo y pasarlos pide `actions/upload-artifact`, que el repositorio no usa (V7).
Que cada trabajo lea los tres informes para validar la medida: tres veces la lectura. El paso que marca el informe
escrito es el patrón de la tanda (la marca «Esta ejecución mide el commit»), y con él cada informe se lee una sola vez
por medida recorrida.

**La matriz (FR-019).** La de GitHub sigue siendo una combinación por skill, con su nombre `evals (<skill>)`, su
informe y su concurrencia de hoy, aunque no abra ninguna sesión; lo que mide y lo que reutiliza cada combinación lo da
su decisión, y eso es lo que fija el golden. Una matriz de GitHub por eval abriría más de un trabajo por skill
(fuera de alcance, FR-020).

### D9 · Lo que se guarda de una eval: sus series y lo que aporta a cada umbral

**Decisión.** El resultado de FR-011 que el informe guarda de cada eval, con su huella, son sus series (las entradas de
`tasas`, también las de los modelos informativos) y sus aportes a cada umbral de las respuestas: por umbral, la medida y
el total que dan sus respuestas, calculados con las mismas funciones del informe sobre las respuestas ya juzgadas de esa
eval, sin volver a votar. Una eval reutilizada lleva los mismos datos copiados, con la procedencia de la medida en que se
abrieron sus sesiones. Los aportes van dentro de la entrada de cada eval del bloque de FR-022 (`origen.evals[]`), que
es lo único que la suite completa añade al informe de hoy (SC-012).

**Alternativa.** Guardar los datos de cada sesión y volver a calcular los umbrales: el informe se compone desde las
sesiones juzgadas con sus trazas (`juzgarSesiones`, `informe.go:1296`), que no están en `informe.json`, así que no se
podrían recalcular.

### D10 · El informe: lo de siempre es lo medido; lo reutilizado, aparte

**Decisión.** Los campos de hoy dicen solo lo medido (FR-023). Un bloque nuevo, `origen`, dice el modo y, de cada eval,
si se midió o se reutilizó, su huella, el motivo o la procedencia, sus aportes y, si se reutilizó, sus series. Cada
umbral de las respuestas con aportes reutilizados lleva al lado un umbral más, `<nombre>:reutilizadas`, con la suma de
esos aportes: decide si el original decide —con umbral 0, la unión pasa si y solo si pasan las dos partes, de modo que
el veredicto es el mismo (FR-015)—, salvo el de la tasa de una clase, que solo se publica (FR-015, FR-024).

**Alternativa.** Un campo con el desglose dentro de cada umbral y `medida`/`total` con todo: `informe.sh` rehace
`medida/total ≤ umbral` y marcaría INCOHERENTE el de la tasa (V7), y `serie.sh` exige que su `medida`/`total` cuenten
solo lo medido; los dos están en `scripts/workflow/`, que el hito no toca (FR-027).

### D11 · Dos lecturas exportadas en `internal/skills`

**Decisión.** `internal/skills` posee las marcas de la tabla de comandos y la cabecera «generado desde». Se exportan dos
funciones de lectura junto a ellas: `AppletsDeLaTabla(contenido []byte) ([]string, error)` y
`DatosDeLaReferencia(contenido []byte) (string, error)`; la descripción ya se lee con `LeerFrontmatter`.

**Alternativa.** Repetir las marcas y el formato en `internal/evals`: dos fuentes del mismo formato que pueden
divergir.

### D12 · El límite de la tasa y cómo se publica

**Decisión.** `tasa.go` calcula, para la clase que decide, *N* (las respuestas del modelo que decide que juzga la
ejecución completa, las del plan entero con los dos modos), el límite de hoy *K* (el mayor *m* con `m/N ≤ umbral`, la
misma comparación del job), *p*ref (el mayor múltiplo de 0,0001 con P(Bin(*N*, *p*) > *K*) ≤ 0,02) y, para las *n*
respuestas medidas, k(*n*), el menor *k* con P(Bin(*n*, *p*ref) > *k*) ≤ 0,02 (FR-024). El umbral `<clase>:<modelo>`
compara `medida ≤ k(n)` y publica como `umbral` el de `clases.yaml` si con él sale el mismo límite y, si no, `k(n)/n`:
la comparación de `informe.sh` da siempre lo mismo que el job, y con todas las respuestas medidas el umbral publicado y
su resultado son los de hoy (FR-026).

**Alternativas.** Publicar siempre `k(n)/n`: con 120 respuestas daría 0,025 en lugar de 0,028 y cambiaría el informe de
la suite completa (SC-012). Aplicar la regla nueva solo con algo reutilizado: dejaría dos reglas para el mismo umbral.

### D13 · La tanda por commit distingue la suite completa

**Decisión.** El trabajo `tanda` recibe el modo (`completa` o `reutilizacion`, del evento) y deja dos marcas: la de hoy,
«Esta ejecución mide el commit», y una nueva, «Esta ejecución mide la suite completa», solo cuando mide la suite
completa. Una ejecución no mide si una anterior sin terminar ya mide con un modo igual o mayor (completa > reutilización)
(FR-021).

**Alternativas rechazadas.** Leer el modo del informe de la ejecución anterior: la tanda decide mientras la otra
ejecución aún no ha terminado, y el informe solo existe al final de cada trabajo `evals (<skill>)`. Leerlo del evento
de la ejecución: las etiquetas `evals` y `evals-completa` dan el mismo evento (`pull_request`), así que el evento no lo
distingue. Poner el modo en el nombre del paso (`Esta ejecución mide el commit (${{ … }})`): obligaría a la tanda a
analizar el texto del nombre, cuando hoy lo compara entero con su conclusión (`marcaDeLaTanda` y `conclusionDeExito`
de `tanda.go`); dos marcas de nombre fijo, cada una con su `if:`, se leen igual que la de hoy.

### D14 · La release, un trabajo antes de publicar

**Decisión.** `release.yml` gana el trabajo `suite-completa`, del que depende `publicar` (y por él `humo` y
`catalogo`): ejecuta `TestSuiteCompletaDelCommit` con el commit etiquetado y la etiqueta, que busca una ejecución del
flujo `evals` sobre ese commit con cada trabajo de skill en verde y un informe de modo `completa`, con todas las evals de
la skill en ese commit medidas y ninguna reutilizada (FR-025). Si no la hay, falla con el commit, lo que falta de cada
ejecución y la orden `gh workflow run evals.yml --ref <etiqueta>` (S6).

**Alternativas rechazadas.** Un paso dentro de `publicar`, antes de goreleaser: ese trabajo tiene `contents: write`,
`id-token: write` y `attestations: write`, y su comentario dice que solo él los tiene (`release.yml:30-35`, leído en
esta sesión); la comprobación, que ejecuta Go y `gh` sobre la plataforma, solo necesita leer. En un trabajo aparte
corre con `contents: read` y `actions: read`, y que `publicar` dependa de ella es un `needs` que
`TestConfiguracionDeLaRelease` lee (criterio de FR-025). Un disparo `workflow_run` (la release al terminar el flujo
`evals`): la release dejaría de lanzarse con el `push` de la etiqueta, tendría que averiguar qué etiqueta corresponde a
cada ejecución de evals, y una etiqueta empujada sin la suite completa no daría el fallo con su mensaje que pide FR-025:
no pasaría nada.

### D15 · La plataforma en los tests: una interfaz y un `gh` sustituto

**Decisión.** La lógica habla con una interfaz pequeña de la plataforma (ficheros de una propuesta, propuestas
fusionadas, ejecuciones del flujo, trabajos de una ejecución, informe de un trabajo); los tests la implementan en
memoria con respuestas sintéticas que el propio test construye con la forma de las de `gh` (las «respuestas grabadas» del
spec). La implementación con `gh` se prueba aparte con el `gh` sustituto de la tanda, comprobando sus órdenes y la
extracción del informe de un registro con prefijos y marcas de orden de bytes. No hay `testdata/` nuevo ni paso
`grabar_datos`: la plataforma no es una fuente de `docs/SOURCES.md`.

**Alternativas rechazadas.** Respuestas reales grabadas de la API: pedirían un paso `grabar_datos` contra una fuente
que no está en `docs/SOURCES.md` y que en la sesión del paso no es accesible (sin red), y los casos que importan
(cancelada, sin fusionar, sin marca, 90 días, orden que falla) habría que fabricarlos igual. Un servidor HTTP local al
que apunte el `gh` de verdad: pediría ejecutar `gh` en `make ci`, que no se puede en el sandbox de las sesiones del run
(no lee su configuración, «Supuestos no verificados»), y lo que probaría de más es el propio `gh`, no la lógica del
arnés; la interfaz en memoria y el `gh` sustituto, como en `tanda_test.go`, prueban las dos mitades por separado.

### D16 · El golden de la matriz, en el test

**Decisión.** El golden de FR-019 es una constante del propio test (`TestMatrizDeLasPropuestas`), comparada byte a byte.

**Alternativa.** Un fichero en `testdata/`: pide una tarea `[datos]` aparte que no puede generarlo sin el código.

### D17 · La decisión de cada trabajo, acotada y sin fallar por la plataforma

**Decisión.** `TestDecisionDelTrabajo` corre con `-timeout 15m` y da a lo que lee de la plataforma —los ficheros de la
propuesta y la búsqueda de medidas de `main`— un plazo de 10 minutos. Un error de la plataforma o el plazo agotado no lo
hacen fallar, y nunca quedan callados: la orden y su salida de error (o «plazo de 10 minutos agotado», con la orden en
curso) se escriben en el registro del paso, antes de las líneas de las evals, y en `origen.error_de_la_plataforma` del
fichero de la decisión, que el informe copia y `informe.md` enseña (data-model §8; contracts/informe-del-job.md §2 y
§6). Lo que pasa después depende de la orden:

- si no se pueden leer los ficheros de la propuesta, se eligen todas las evals de la skill con el motivo `plataforma`,
  sin rutas, y no se busca: es la lectura conservadora, la misma que FR-005 da a un conjunto incalculable; el trabajo
  `cambios`, con la misma orden, da `elige=si` y escribe el error en su registro;
- en la búsqueda, la primera orden que falla, o el plazo, la para: lo que queda sin resultado se mide con
  `sin_resultado_de_main` (FR-013, FR-014), y con el error al lado se sabe que se mide por él.

Un error del cálculo del conjunto elige todas las evals y el informe lo dice (FR-005). El peor caso del tope del trabajo
suma esos 15 minutos.

**Alternativas rechazadas.** Sin plazo propio, solo con el tope del trabajo: un `gh` colgado gastaría los 352 minutos
del tope y no quedaría tiempo para ninguna sesión. Hacer fallar el trabajo ante un error de la plataforma: dejaría en
rojo una propuesta cuyas evals pasan por algo ajeno a ellas, contra FR-013 («plataforma que no responde» → se mide).
Seguir con la medida anterior cuando una orden de la búsqueda falla, en lugar de parar: con la plataforma caída fallarían
las diez y el registro tendría diez errores para decir lo mismo; parar en el primero deja uno y mide lo que falta, que
es lo que pasaría igualmente. Un motivo nuevo para las evals que quedan al parar la búsqueda: FR-014 les da
`sin_resultado_de_main`, y el error escrito al lado ya dice por qué no hubo resultado.

### D18 · `scripts/evals.sh` exige la decisión

**Decisión.** `DECISION_DE_EVALS` pasa a ser obligatoria en `scripts/evals.sh` (la ruta del fichero que escribe la
decisión), y llega a `TestEjecucionDelJob` como `-decision`.

**Alternativa rechazada.** `DECISION_DE_EVALS` opcional, midiéndolo todo sin ella: solo el job ejecuta `make evals`, y
un camino sin decisión sería un segundo modo que nadie usa.

### D19 · El motivo de una eval elegida

**Decisión.** De cada eval elegida, `rutas` son todas las rutas cambiadas que la eligen, ordenadas, y `motivo` el de
más prioridad entre ellas: `comun`, `sin_conjunto` (FR-005 en el job), `plataforma` (sin los ficheros de la propuesta,
D17), `sin_asignar`, `descripcion` y `conjunto`. Una eval no elegida que se mide lleva `sin_resultado_de_main`; en la
suite completa, `suite_completa`.

**Alternativa rechazada.** Una entrada por ruta y eval, cada una con su motivo: repetiría la eval tantas veces como
rutas la eligen; con el motivo de más prioridad y todas sus rutas en una entrada, el registro, el informe y el golden
tienen una línea por eval (FR-008).

### D20 · Aceptación e2e: no aplica

El hito no cambia nada que vea quien usa `kitlegal`: el binario no enlaza `internal/evals`, ninguna skill ni eval
cambia (FR-026), y las dos funciones nuevas de `internal/skills` solo leen. Como en H24, H26 y H26.1, no hay
comportamiento del binario que un guion `testscript` describa; la aceptación del spec (SC-006, SC-007) son tests de
`make ci` sobre la decisión del job, y su medición de cierre es la del workflow (SC-014).

**Alternativa rechazada.** Un guion `testscript` sobre `kitlegal`: no describiría nada de la entrega (plan, «Complexity
Tracking»).

## Datos externos

Ninguno. El hito no lee ninguna fuente de `docs/SOURCES.md` ni graba nada: las respuestas de la plataforma de los tests
son sintéticas (D15), y `data/` no cambia.

## Fuera de alcance que el plan respeta

Lo de «Fuera de alcance» del spec: ningún cambio en `scripts/workflow/`, `.specify/` ni `internal/workflow/`; ninguna
marca de materia; ningún cambio en evals, umbrales, clases, repeticiones, modelos, versiones ni concurrencia; ninguna
medida periódica de `main`; crear la etiqueta `evals-completa`, etiquetar y publicar son de una persona.
