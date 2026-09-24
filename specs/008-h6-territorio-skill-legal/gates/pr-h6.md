<!-- Propuesta de cambio de H6. La escribió la tarea de cierre (T025, intento 2) con la medida hecha sobre 2df74db (`feat(H6): T024`) el 2026-09-22; la tarea de plataforma (T026, intento 2) la puso al día con T029 (2b81164) y T026, intento 3, registró la primera ejecución de aceptación válida (35722605048, sobre 2b81164). La revisión final (2026-09-22 a 2026-09-24) la corrigió en dos rondas y registró en gates/evals-cierre.md la ejecución de aceptación vigente: la repetida sobre la cabeza tras las correcciones fuera de specs/ (ver «Revisión final»). -->

## Objetivo

«Que toda skill sepa, antes de razonar, en qué territorio está la pregunta y qué normas vertebrales aplican. Es la
pieza que hace genéricas a las demás» (`docs/ROADMAP.md` §4, H6, primer hito de la fase 1). H6 entrega dos cosas:

1. **El applet `territorio`**, con el verbo `resolver`: dado el nombre de un municipio o su código INE —de cinco cifras
   o de seis con el dígito de control—, devuelve municipio, código INE, provincia, comunidad autónoma, DIR3 del
   ayuntamiento, régimen (`comun` o `foral`), boletines aplicables y `cobertura`: qué del territorio está configurado y
   verificado y qué no. No pide nada por red: su dato vive congelado en `data/territorio/`, generado por una persona
   fuera de la ejecución a partir de la relación oficial del INE y del Registro de Entidades Locales (ADR 0017), y
   viaja dentro del binario. Un nombre que comparten varios municipios devuelve todos los candidatos y código `2`.
   Con él entran los dos primeros identificadores del proyecto, código INE y DIR3 (`internal/core/ids`), con su
   gramática y su dígito de control.
2. **La skill `legal-core` v0**, la skill madre: su protocolo **empieza por identificar el territorio** y solo después
   razona, con `references/leyes_vertebrales.md` y `references/jerarquia_normativa.md` generadas desde `data/`, y tres
   evals —municipio cubierto, municipio no cubierto y no activación— escritas antes que ella.

La regla que gobierna el hito es la genericidad territorial (constitución, principio IX): **ningún caso especial para
un municipio** en código, en `data/` ni en la skill; fuera del territorio configurado —hoy, solo la Comunidad de
Madrid— la salida **declara su cobertura** en lugar de fallar, de vaciarse o de inventar un boletín o un DIR3 que nadie
ha verificado.

## Alcance

Frente a `main` (`04a2aaa`, `docs(h6): registrar fuentes congeladas para territorio y festivos (ADR 0017) (#38)`), en
`6d08752` (`fix(H6): los motivos de la ronda 2 de la revisión final`), el último commit que cambia algo fuera de
`specs/`: 42 commits; fuera de `specs/`, **157 ficheros, 30 500 líneas añadidas y 670 retiradas**, de las que 16 445 son los ficheros congelados de `data/territorio/` y 1 121 las grabaciones del BOE. En
`specs/008-h6-territorio-skill-legal/`, los artefactos del hito, que siguen cambiando con lo que registran el cierre,
la plataforma y la revisión. Por árboles:

- **`internal/core/ids`** (nuevo, 7 ficheros Go y 26 de corpus): `ine.go` (`AnalizarCodigoINE`,
  `AnalizarCodigoINEConDigito`, `ComprobarDigito`), `dir3.go` (`AnalizarDIR3`, `DIR3DeAyuntamiento`), `errores.go` (errores
  tipados con la clase `argumentos`), `doc.go`, sus tests y los dos objetivos de fuzz `FuzzCodigoINE` y
  `FuzzCodigoDIR3`, con trece semillas versionadas cada uno en `testdata/fuzz/`.
- **`internal/core/territorio`** (nuevo, 16 ficheros Go): `fuentes.go` (los cuatro tipos de fichero congelado y
  `Cargar`, que valida la integridad entre ficheros), `filas.go` (T029: la relación y la correspondencia, cuando están
  en su forma de una fila por línea, se leen sin construir el árbol de nodos del lector de YAML; cualquier otro fichero
  lo sigue leyendo el lector entero, con los mismos defectos), `registro.go`, `nombres.go` (el pliegue `Plegar` y las
  formas alternativas derivadas del nombre oficial), `resolver.go`, `salida.go` (las ocho claves de `data`, cada dato
  con su `source`, y `cobertura`), `errores.go`, `doc.go` y sus tests, todos sintéticos —entre ellos `coste_test.go`,
  el control del coste de la carga—: el dominio recibe bytes y no lee ningún fichero.
- **`data/`**: `datos.go` (paquete `data`: los `//go:embed` y el recorrido del subárbol de comunidades, sin
  interpretar nada) y `datos_test.go`; `territorio/municipios.yaml` (8 132
  municipios de la relación del INE con provincia, comunidad y dígito de control; `fecha: 2026-02-04`),
  `territorio/dir3.yaml` (8 132 filas; la regla, verificada contra DIR3 real), `territorio/estado.yaml` (el BOE) y
  `territorio/comunidades/` (19 ficheros, uno por comunidad y ciudad autónoma, con régimen y provincias; solo `13.yaml`,
  la Comunidad de Madrid, trae `boletines`: el BOCM, autonómico y provincial con su motivo); `jerarquia.yaml` (los
  cinco niveles, su boletín, sus tipos de norma y las cuatro reglas de interpretación); `normas.yaml` gana siete
  normas (Código Civil, LJCA, LEC, LOPJ, LGS, LOPDGDD y LGP, con identificador, título y rango copiados de su
  búsqueda grabada; la LRBRL y la LCSP ya estaban desde H5) y la marca `vertebral: true` en las quince de la tabla de `refs/` §1.4.
- **`internal/app`** (10 ficheros): `territorio.go` (el applet, registrado junto a `boe`), `registro.go` (inyección
  de `FuentesEmbebidas`), `esquemas_test.go` (tabla de esquemas parametrizada por applet), `skills_test.go` (casos
  negativos parametrizados por skill y `skillsExigidas` con las dos skills), el guion e2e `territorio-matriz.txtar`
  (206 líneas), los dos guiones existentes (`ayuda`, `argumentos`) actualizados por el segundo applet,
  `ejemplo/kitlegal-e2e/main.go` y los tests de todo ello.
- **`internal/skills`** (14 ficheros): `territorio.go` (validación de los cuatro ficheros congelados contra su
  esquema) y `jerarquia.go` (lector de `data/jerarquia.yaml`), la tabla de generadores de referencias en
  `referencias.go` (rompe el acoplamiento «nombre de la referencia = fichero de datos»), `frontmatter.go`
  (`kitlegal-referencias`), `normas.go` (`Vertebral`), `sincronia.go`, `export_test.go`, sus tests —entre ellos
  `TestTerritorioDelRepositorio`, con nueve subtests sobre el corpus real, 900 líneas— y el guion de reinstalación
  `instalar-de-nuevo.txtar`, que deja de nombrar las skills.
- **`internal/evals`** (12 ficheros): `territorio.go` (el esperado de territorio y su juicio por forma fija),
  `formato.go` (cuarta variante de comando y `Territorio`), `juzgar.go` (`TerritorioEncontrado`, `TerritorioAusente`
  y `formaDelComando`, la discriminación de la variante en un solo sitio), `conjunto.go` (`ComprobarConjunto`
  parametrizado por reglas: `ReglasDeBoeLegislacion()` y `ReglasDeLegalCore()`), `consultas.go`, `informe.go` (dos
  columnas nuevas) y sus tests.
- **`internal/arch_test.go`** y **`cmd/kitlegal/main_test.go`**: `go.yaml.in/yaml/v3` en `modulosDelBinario` con su
  motivo (abajo, *Dependencias*), y el segundo applet en las expectativas del multicall.
- **`schemas/`** (8 ficheros, todos en tareas `[datos]` con pausa): `municipio.json` (el contrato publicado del
  applet, generado desde `--describe`), `territorio-municipios.yaml.json`, `territorio-dir3.yaml.json`,
  `territorio-estado.yaml.json`, `territorio-comunidad.yaml.json` y `jerarquia.yaml.json` (nuevos);
  `normas.yaml.json` (propiedad `vertebral` y los rangos nuevos del `enum`) y `eval.yaml.json` (variante de territorio,
  esperado `territorio` y la regla del esperado verificable).
- **`skills/legal-core/`** (nuevo): `SKILL.md` (170 líneas), `references/leyes_vertebrales.md` y
  `references/jerarquia_normativa.md` (regeneradas, con su cabecera «generado desde…, no editar») y el enlace
  `scripts/territorio`. **`skills/boe-legislacion/`**: solo `references/normas.md`, regenerado con las siete normas
  (7 inserciones); `SKILL.md` y sus evals, intactos.
- **`evals/legal-core/`** (nuevo, 3 ficheros): `01-territorio-municipio-cubierto.yaml`,
  `02-territorio-municipio-no-cubierto.yaml` y `03-no-activa-receta-de-cocina.yaml`; ninguna con `citas`, porque la
  skill no afirma contenido de norma. Confirmadas en `5c11d1d` (T020), antes que la skill (`7800adf`, T022).
- **`testdata/evals/`** (22 ficheros, pausa de T015): las siete búsquedas del BOE, sus siete índices y sus siete
  metadatos, grabados por una persona, y `grabaciones.json` con las siete entradas nuevas.
- **`.github/workflows/evals.yml`**: matriz de skills (`boe-legislacion`, `legal-core`) con `fail-fast: false`, un
  informe por skill en la misma ejecución; la prueba de red solo en el trabajo de `boe-legislacion`; el filtro de rutas
  gana `internal/core/*`.
- **`Makefile`**: cambian dos recetas. `skills-check`, cuya expresión `-run` gana `TestTerritorioDelRepositorio` y
  `TestJerarquiaDelRepositorio`; y `release`, cuyo comentario y cuyo mensaje dicen ahora H19, el hito de la release
  (ADR 0013), en vez de H6 (revisión final, `c481451`).
- **Documentación** (4): `CHANGELOG.md` (bloque *De H6* en «Añadido» de *Unreleased*), `README.md` y
  `CONTRIBUTING.md` (el applet, la skill, `data/territorio/` y la fila de `make skills-check` de sus tablas de
  controles) y `docs/SOURCES.md` (la fila `mpt.rel` con la fecha del volcado, 2026-09-21, y la fila
  `ine.codigos-territoriales` de las tablas de códigos de comunidad y provincia, las dos dentro de la pausa de T003;
  y, en la revisión final, la nota al pie `[^rel-dir3]`, que pasa de anunciar la verificación del DIR3 a describir la
  que se hizo, en `9a77c6b`, y la «Fecha del fichero» de `ine.codigos-territoriales`, que pasa de 2026-09-21 a
  2026-09-20, el día de la descarga, en `6d08752`).

**Sin cambios**, como exige el spec: `internal/core/schema`, `internal/cli`, `internal/httpx`, `internal/cache`,
`internal/source/boe`, `internal/render`, `cmd/kitlegal/main.go`, `scripts/` (ningún caso nuevo en
`verify-sources.sh`: `grep -c territorio` da `0`), `.golangci.yml`, `codecov.yml`, `docs/ADR/`, `go.mod`, `go.sum`,
`skills/boe-legislacion/SKILL.md` y `evals/boe-legislacion/` (diff vacío frente a `main`).

**Fuera de alcance** (spec, *Fuera de alcance*): consultar en red el INE, el REL o el inventario DIR3; la procedencia
del sobre fuera del espacio `kitlegal.` / `kitlegal:` de ADR 0006; el régimen como configuración por comunidad;
otra fuente de municipios que la relación del INE; plazos, recursos y competencia (H9); operaciones de grafo (H7,
ADR 0014); otros territorios configurados (backlog, «territorios»); ADR nuevo. Puntos 7 y 12 de la Definition of Done:
no aplican; el 8, solo en su parte de documento (las fuentes no se piden en red, ADR 0017).

## Dependencias (constitución §V)

**Dependencias: ninguna nueva.** `go.mod` y `go.sum` no aparecen en el diff frente a `main`. Lo que cambia es **qué
enlaza el binario distribuido**: un único módulo, `go.yaml.in/yaml/v3`, pasa de dependencia directa que solo usaban
`internal/skills` e `internal/evals` a módulo enlazado, y entra en `modulosDelBinario` (`internal/arch_test.go`, T009)
con este motivo: `internal/core/territorio` analiza con él los ficheros congelados de `data/territorio/`, que son YAML
como todo `data/` (`docs/ROADMAP.md` §2); entra por `internal/app`, cuyas importaciones sigue `go list -deps` todas, en
cuanto existe el applet; es la biblioteca de YAML que el repositorio fijó en H5; y `go.yaml.in/yaml/v4`, que el binario
ya enlazaba por `github.com/pb33f/ordered-map/v2`, no tiene ninguna versión estable, motivo por el que H5 la rechazó
(research.md D15; plan, *Complexity Tracking*). Las alternativas rechazadas: JSON en lugar de YAML (rompe la
homogeneidad de `data/`) y analizar en `internal/app` (deja al dominio sin validar lo que recibe). El pliegue de
nombres es propio y no promueve `golang.org/x/text` a directa (D10). Sin cambio de versión en ninguna herramienta de
control ni en el job de evals.

## Controles añadidos

Los 28 controles de la tabla del plan («Controles mecánicos que este hito añade o toca») están en el árbol: 27 dentro
de `make ci` y el último, la aceptación con modelo, fuera de él por diseño; T029 añadió uno más dentro de `make ci`,
el coste de la carga. Lo que pasa a ser mecánico, con el escenario del quickstart que lo demuestra:

- **Ficheros congelados contra su esquema, con integridad y procedencia** (`TestTerritorioDelRepositorio/esquema`,
  `/integridad`, `/fuentes`, `/madrid-configurada`, `/regimen-de-todas`, dentro de `make skills-check`): una fila
  cuya `comunidad` no es la de su provincia falla con «el municipio 01001 declara la comunidad 99
  (…/99.yaml) y su provincia 01 es de la comunidad 16 (…/16.yaml)» (escenario 9).
- **El pliegue atado al corpus real** (`/pliegue-cubre-el-corpus`, `/nombres-alcanzables`,
  `/ningun-nombre-es-solo-cifras`): una runa nueva en la relación del INE hace fallar `make ci` con «el municipio 01001,
  "ØAlegría-Dulantzi", se pliega a "øalegria dulantzi", con "ø" fuera del alfabeto del pliegue» en lugar de pasar en
  silencio; ningún municipio queda inalcanzable por su nombre oficial (escenario 9).
- **Gramáticas atadas a los esquemas de datos** (`/gramaticas`, 76 casos): los `pattern` de los tres esquemas con
  código INE, provincia o DIR3 aceptan y rechazan exactamente lo mismo que los analizadores (T027).
- **Identificadores** (`TestAnalizarCodigoINE`, `TestAnalizarDIR3`, `TestComprobarDigito`, `TestIdaYVuelta`; fuzz
  `FuzzCodigoINE` y `FuzzCodigoDIR3` con corpus semilla en `make test`): `2807`, `28074 `, provincia `00` o `99`,
  municipio `000` y `X01280748` se rechazan nombrando qué tienen de malo, sin `panic`; en la campaña de 30 s, 3 791 151
  y 3 929 765 ejecuciones sin hallazgo (escenarios 6 y 10).
- **Forma del sobre y códigos del applet** (`TestResolverDevuelveElTerritorio`, `TestCodigosDeTerritorio`,
  `TestSalidaDeTerritorioContraSchemas`): ocho claves siempre, `cobertura` con tres, la salida real validada contra el
  esquema publicado; `0` resuelto, `2` ambiguo o mal formado, `3` no encontrado, nunca `4`, `5` ni `6` (escenarios 2 y
  6).
- **`--describe` sin deriva y cobertura de esquemas** (`TestEsquemasPublicados`, `TestEsquemasCubrenTodosLosVerbos`,
  en `make schema-check`): `schemas/municipio.json` alterado falla con «el fichero no es la serialización canónica de
  sus partes» (escenario 8).
- **Matriz territorial en e2e** (`territorio-matriz.txtar`, en `TestEntregaDelHito`): Leganés (cubierto), Tordesillas
  (no cubierto, sin ningún boletín no configurado en la salida), Abáigar y Amurrio (forales) y un nombre ambiguo; nombre,
  código, código con dígito y `--offline` byte a byte iguales (`cmp`); la invocación desde otro directorio; y el
  tiempo con `cronometra` (escenarios 3 y 7).
- **Coste de la carga del territorio** (`TestCosteDeLaCarga`, T029, con `TestFormaDeLasFilas`, `TestCargarEnFilas`,
  `TestNombresDificiles` y `FuzzFilasComoElLector`): cargar una relación y una correspondencia sintéticas del tamaño
  real (8 132 filas cada una) no puede pasar de 259 688 asignaciones ni de 12 485 302 bytes —un tercio de lo que
  asignaba la carga de `47f3090`, medido antes de cambiarla—, medido con asignaciones y no con tiempo para que no
  dependa de la máquina; y la lectura en su forma da exactamente lo que da el lector de YAML, defecto a defecto, sobre
  50 ficheros escritos a mano y bajo fuzz (`gates/tarea-T029.md`).
- **Skill nueva bajo los mismos controles negativos que `boe-legislacion`** (`TestSkillsDelRepositorio`, casos
  parametrizados por skill): una referencia editada a mano falla con «legal-core: references/jerarquia_normativa.md:
  contenido-distinto»; `make skills-sync` es idempotente (escenario 11).
- **Referencias generadas desde `data/`** (`TestRegenerarYComparar`, `TestNormasDelRepositorio/vertebrales`,
  `TestJerarquiaDelRepositorio`): exactamente quince normas `vertebral`; los cinco niveles y las cuatro reglas en su
  orden.
- **Identificadores de las normas nuevas contra su búsqueda grabada** (`TestIdentificadoresDeLasNormas`): un
  `BOE-A-…` escrito de memoria falla nombrando la norma (escenario 11).
- **Formato de eval ampliado y compatible** (`TestLeerEval`, `TestEsquemaDeEval`, `TestFormaDelComando`,
  `TestJuzgar/territorio-*`, `TestEvalsDelRepositorio/formato`, `/conjunto-legal-core`, `/cobertura-del-esquema`):
  las 18 evals de `boe-legislacion` validan sin cambiar un byte; una eval activa sin `citas` ni `territorio` es un
  fichero mal formado; el enumerado de cobertura del esquema y el vocabulario del applet dicen lo mismo (escenario 12).
- **Reglas de arquitectura y perfil del binario** (`TestArquitectura` R1-R3, `TestDependenciasDelBinario`,
  `depguard`, `forbidigo`): el dominio no importa `os`, `io` ni `io/fs`, tampoco en sus tests; un módulo enlazado sin
  justificar falla.
- **Que nada de esto toca la red**: `TestArquitectura` R2 y ningún caso nuevo en `scripts/verify-sources.sh`
  (escenario 13).
- Los de H0-H5.1 (`-race`, `govulncheck`, `gosec`, `gitleaks`, `go mod verify`, `tidy -diff`) siguen sin exclusiones
  nuevas.
- **Aceptación con modelo** (fuera de `make ci`): la primera ejecución `evals` de la rama disparada por la propuesta de
  cambio, con sus dos trabajos de la matriz; la registra T026 en `gates/evals-cierre.md` (SC-013, SC-015).

## Evidencia

Medida por T025 (intento 2) el **2026-09-22, entre las 11:55 y las 12:03 (hora de Madrid), sobre `2df74db`**
(`feat(H6): T024`, la cabeza de la rama), ejecutando en una sesión desatendida y en primer plano `make ci` y después
`quickstart.md` desde los prerrequisitos hasta la limpieza (escenarios 1 a 13), con cada orden tal cual y en orden y
las formas `rtk proxy` de su tabla. **Ningún resultado distinto del esperado.** El intento 1 (mismo commit, 11:40 a
11:49) dio uno, en el escenario 11, y era de la guía y no del producto: lo corrigió en `quickstart.md` (abajo,
*Pendientes*).

| Escenario | Resultado |
|---|---|
| Prerrequisitos | `go version go1.27.1 darwin/arm64`; rama `008-h6-territorio-skill-legal`; `make check-tools` en verde; solo `fin del estado`; carpeta temporal recreada y vacía; `bin/kitlegal` compilado; `código: 0` |
| 1 · `make ci` | **En verde de 11:55:25 a 11:56:24**: `0 issues.`; los trece paquetes con tests en `ok` en los dos perfiles (`internal/core/ids` y `internal/core/territorio` al 100 %); `govulncheck` «No vulnerabilities found.» y «Your code is affected by 0 vulnerabilities.» (ver *Pendientes*); `schema-check` y `skills-check` (`internal/app`, `internal/skills`, `internal/evals`) en `ok`; `gitleaks` «no leaks found»; `all modules verified` en la raíz y en los cuatro módulos de herramientas; `go mod tidy -diff` sin salida; `ci: todos los controles en verde` |
| 2 · cubierto | Sobre con `ok` verdadero, `kitlegal.territorio`, `kitlegal:applet/territorio`, `fecha_consulta` `2026-02-04T00:00:00Z` y `hash` `sha256:`; las ocho claves; `configurado`, `configurado`, `verificado`; `estatal BOE`, `autonomico BOCM`, `provincial BOCM` con su motivo; `L01280745` y `comun`; `código: 0` |
| 3 · determinismo | `las cuatro salidas son idénticas byte a byte`, sin ningún mensaje de `cmp` |
| 4 · no cubierto | `no-configurado` en los dos boletines y `verificado`; solo `estatal`; Valladolid, Castilla y León, `comun`; recuento `0` (código de `grep` 1); `código: 0` |
| 5 · foral | `comunidad foral: 15 · municipio: Abáigar`; `foral`; `no-configurado` |
| 6 · códigos | `nombre de más de un municipio: Arroyomolinos` → **2**, clase `argumentos`, «el nombre "Arroyomolinos" es el de 2 municipios de la relación; consulta uno por su código INE: 10023 Arroyomolinos (Cáceres); 28015 Arroyomolinos (Madrid)»; `Municipio Que No Existe` → **3**; `bien formado y ausente de la relación: 01999` → **3**, `no-encontrado`; `99999` → **2** («la provincia "99" no está entre 01 y 52»); `2807` → **2** («tiene 4 cifras y la forma PPMMM tiene 5»); `280746` → **2** («el dígito de control recibido es "6" y el oficial es "5"»); ningún 4, 5 ni 6 |
| 7 · e2e | `ok` (`TestEntregaDelHito/territorio-matriz`) |
| 8 · contrato | `schema-check` en `ok`; `territorio resolver` y `consulta`; `territorio · municipio` y `resolver`; sin posicional, `argumentos inválidos: expected "<consulta>"` y `código: 2`; en el clon alterado, `schemas/municipio.json: el fichero no es la serialización canónica de sus partes` y `código: 2` |
| 9 · datos | Los nueve subtests en `PASS`; cabecera `fecha: "2026-02-04"` y `source: ine.municipios`; **8132** municipios; **19** comunidades; **18** sin `boletines`; el registro de verificación tal como estaba entonces (la revisión final lo reescribió: *Decisiones*); `todos los municipios tienen DIR3 verificado`; integridad rota en el clon → `integridad` falla nombrando `01001`, `99` y `16`, `código: 2`; runa fuera del pliegue → `pliegue-cubre-el-corpus` falla nombrando `01001` y `Ø`, `código: 2` |
| 10 · ids y fuzz | `ok`; trece entradas de corpus en cada objetivo; `FuzzCodigoINE` 3 791 151 ejecuciones y `FuzzCodigoDIR3` 3 929 765 en 30 s, los dos `PASS` y sin hallazgo; el árbol, limpio después |
| 11 · skill | Frontmatter con `kitlegal-applets: territorio` y `kitlegal-referencias: leyes_vertebrales jerarquia_normativa`; **170** líneas; las dos cabeceras «generado desde…, no editar»; `../../../bin/instalado/kitlegal`; **15** marcas `vertebral: true`; solo `references/normas.md` (7 inserciones) en `skills/boe-legislacion`; los tres tests en `ok`; en el clon, `skills-sync` lo deja limpio (solo `fin del estado del clon`) y la edición a mano falla con «legal-core: references/jerarquia_normativa.md: contenido-distinto», `código: 2` |
| 12 · evals | Las tres evals; `formato`, `conjunto`, `conjunto-legal-core`, `normas-conocidas`, `grabado` y `cobertura-del-esquema` en `PASS` (y `avisos-del-esquema` y `avisos-de-la-skill`, de H5.1); `5c11d1d feat(H6): T020` antes que `7800adf feat(H6): T022`; solo `fin del diff` |
| 13 · documentación | `CHANGELOG.md:26` y `:281` (*De H6*); `legal-core` en `README.md` y en `CONTRIBUTING.md` (3 veces), y `data/territorio/` en los dos; la fila `mpt.rel` con 2026-09-21; `0` y `código de grep: 1` |
| Limpieza | La carpeta temporal desaparece con sus cuatro clones; el estado del árbol da solo `fin del estado` |

**Cobertura**, con `go tool cover -func` sobre el `coverage.out` y el `coverage-integration.out` que dejó ese `make ci`
sobre `2df74db` (y, por árbol, la suma de sentencias del perfil contando cada bloque una vez):

| Umbral | Exigido | Perfil unitario (`make test`) | Perfil de integración |
|---|---|---|---|
| Global (`codecov/project`) | ≥ 70 % | **96,8 %** (6786/7007) | **97,3 %** (6815/7007; `go tool cover -func`: 97,2 %) |
| `internal/core/**` | ≥ 85 % | **98,6 %** (583/591) | 98,6 % |
| `internal/core/ids` | — | 100,0 % (64/64) | 100,0 % |
| `internal/core/territorio` | — | 100,0 % (446/446) | 100,0 % |
| `internal/core/schema` | — | 90,1 % (73/81), sin cambios desde H5.1 | 90,1 % |
| `internal/cli/**` | ≥ 90 % | 98,6 % (348/353), sin cambios | 98,6 % |
| `internal/app` | — | 93,6 % (466/498) | 93,6 % |
| `internal/skills` | — | 98,2 % (1269/1292) | 98,2 % |
| `internal/evals` | — | 99,1 % (1905/1922) | 99,1 % |

**Todo el dominio que añade H6 está cubierto al 100 % por sus tests sintéticos**, que es lo que la obligación 7 del
plan exige: `make test` mide por paquete y sin `-coverpkg`, así que lo que `TestTerritorioDelRepositorio` ejerce del
dominio desde `internal/skills` no cuenta para el umbral. El paquete `data` estaba al 0 % cuando se midió esta
tabla, porque sus tests eran los de `internal/app`; la revisión final le dio los suyos (`datos_test.go`, las tres
ramas de error del recorrido de comunidades sobre `fstest.MapFS`) y queda al 100 %. **Ningún umbral se rebaja**:
`codecov.yml` no aparece en el diff.

**Una sola supresión nueva, razonada**: `//nolint:paralleltest` en `internal/core/territorio/coste_test.go:36`, porque la prueba mide lo que asigna todo el proceso y otra en paralelo contaría lo suyo; `0` `t.Skip` y `0` `TODO` añadidos en ficheros `.go` frente
a `main`; ninguna entrada nueva en `misspell.ignore-rules` (`.golangci.yml` no cambia).

**Sin red en las tareas**: ninguna ejecutó `KITLEGAL_RECORD`, `scripts/grabar-evals.sh`, `make evals`,
`make verify-sources` ni ninguna descarga. La relación del INE (T001), los 19 ficheros de comunidad (T002), el
volcado del REL con la correspondencia y su registro de verificación (T003), los patrones de los esquemas (T027), el
esquema publicado (T010), los guiones e2e (T012), la marca `vertebral` (T013), la jerarquía (T014), las siete búsquedas
grabadas con sus índices (T015), el esquema de eval (T018) y el guion de reinstalación (T028) los aportó o revisó una
persona en la pausa de su tarea `[datos]`; el corpus de fuzz (T005) y los guiones e2e (T012) no tienen pausa y los
revisa la revisión final del hito (FR-086).

## Decisiones

- **La derivación del DIR3 es una regla, verificada contra DIR3 real** (T003 y revisión final;
  `gates/verificacion-dir3.md`): `L` + el número de inscripción del REL. La muestra de FR-046 —siete municipios:
  Leganés; los fusionados Oza-Cesuras y Cerdedo-Cotobade; los forales Pamplona/Iruña y Vitoria-Gasteiz; Riello y
  Vitoria-Gasteiz, con 37 y 61 entidades locales menores según el volcado `eatimes` del REL; y Soba, duplicada en el
  volcado de municipios— se comparó el 2026-09-24 con las fichas de unidad orgánica del directorio del Punto de Acceso
  General, consultadas dentro de su `robots.txt`: cada ficha da el código derivado como el de «Ayuntamiento de» ese
  municipio, siete coincidencias y ninguna discrepancia. Aparte, la regla se aplica sin discrepancia a los 8 132
  municipios —el número de inscripción del REL coincide con el código INE y su dígito de control en todos—, que es el
  criterio de entrada de FR-048; así que `dir3.yaml` entra en extenso con las 8 132 filas y ningún municipio queda
  `no-verificado`. La primera versión del registro (pausa del 2026-09-21) daba esa coherencia REL↔INE por
  verificación, y era circular: la rechazaron los dos jueces en la revisión final y se rehízo con la consulta al
  directorio. La alternativa de FR-048 —la derivación como tabla de excepciones— no hizo falta porque no hubo
  discrepancias; la de recalcular la regla en ejecución se rechaza porque el fichero es el registro de lo verificado
  en una fecha, no un cálculo.
- **El dígito de control es un dato oficial, no un algoritmo** (D9): `ids` compara el dígito declarado con el que
  viaja en la relación; escribir el algoritmo de memoria sería lo que ADR 0017 prohíbe con el DIR3. Consecuencia en
  los códigos: provincia fuera de `01`-`52` o municipio `000` no llega a ser un código (`2`), un código bien formado
  ausente es `3`, un dígito distinto del oficial es `2`.
- **Pliegue de nombres propio, atado al corpus** (D10, D27): en lugar de `golang.org/x/text/unicode/norm` (fuera de
  §V y sin garantía de no colapsar dos municipios), un pliegue explícito cuya cobertura se prueba sobre las 8 132
  filas reales; las formas alternativas (`Coruña, A` → `A Coruña`; `Donostia/San Sebastián` → las dos) se derivan del
  nombre oficial sin caso especial. Los tres controles del corpus viven en `internal/skills`, porque el dominio tiene
  denegada la entrada y salida también en sus tests (`depguard` con `run.tests`), y ponerlos en `internal/app` los
  dejaría fuera de `make skills-check`.
- **`fecha_consulta` es la del fichero congelado más antiguo** (D6), no el reloj: es la regla de ADR 0015 y lo que
  hace posible la igualdad byte a byte con `--offline` (SC-001). `--dry-run` no cambia lo que hace el applet ni emite
  sobre, como fijó H1 (D7).
- **Ocho claves sin `omitempty`, cada dato con su `source`, y `cobertura` con vocabulario cerrado** (D11, D12): un
  boletín no configurado no aparece —ni vacío—, porque una entrada vacía se leería como «no tiene», que es lo que FR-022
  prohíbe; la equivalencia autonómico = provincial en Madrid es una decisión por territorio escrita en su fichero, no
  deducida de que sea uniprovincial.
- **Procedencia `kitlegal.territorio` / `kitlegal:applet/territorio`** (D13; ADR 0006, fila «applet calculado»): los
  identificadores de fuente (`ine.municipios`, `ine.codigos-territoriales`, `mpt.rel`) viajan dentro de `data`, en
  cada dato: el de los nombres de provincia y comunidad es `ine.codigos-territoriales`, la fila de las tablas de
  códigos de donde salieron. Los candidatos del
  nombre ambiguo van en el mensaje, en forma fija y ordenados por código (D14), porque `data` de un fallo es
  `{clase, mensaje}` y añadirle una clave cambiaría ADR 0006.
- **Datos embebidos por un paquete `data` sin lógica** (D15; Clarifications Q4): `//go:embed` no sube de directorio,
  y FR-041 exige que los ficheros vivan en `data/territorio/`. Se rechazan moverlos a `internal/core/territorio/data/`
  (contradice el roadmap), un paquete en la raíz del módulo (ocupa el nombre reservado a `pkg/legalkit`) y un fichero Go
  generado con los municipios como literales (Q4).
- **El registro del applet y su esquema publicado van en la misma tarea `[datos]`** (D16): en cuanto
  `territorio resolver` está en el registro, `TestEsquemasCubrenTodosLosVerbos` exige su parte, y el esquema se genera
  desde el applet; separarlos deja `make ci` en rojo entre dos tareas. Misma razón para que el esquema de eval traiga la
  expectativa de test que su `anyOf` cambia (D28) y para que el `enum` de `rango` crezca con las grabaciones y no con la
  marca `vertebral` (D29): `TestEsquemaDeNormas/rangos-grabados` exige igualdad con el conjunto grabado en los dos
  sentidos.
- **Las gramáticas de los analizadores son los `pattern` de los esquemas** (T007 → T027): T001 y T003 habían fijado
  `^[0-9]{5}$` y `^L01[0-9]{6}$`, que aceptan `00074`, `53001` o `L01280000`; el subtest `gramaticas` exigía igualdad
  con `ids`, y cambiar los esquemas es material `[datos]`, así que se redelimitó en una tarea con pausa en lugar de
  relajar el control.
- **Referencias por tabla de generadores** (D20): `leyes_vertebrales` sale de `data/normas.yaml` con la marca
  `vertebral: true` (Q2), en lugar de un `data/leyes_vertebrales.yaml` que duplicaría las normas contra la fuente única
  de verdad.
- **El formato común de eval crece por donde ya crecía** (D21; Q3): cuarta variante de comando, esperado `territorio`
  comparado por forma fija (comunidad y provincia plegadas, código de boletín como palabra exacta, aspecto de cobertura
  en su forma `<aspecto>: <valor>`) y «toda eval activa exige al menos un esperado verificable»; ni formato propio ni
  juicio de la redacción libre (constitución, capa 1; H5.1 lo descartó para los avisos). Las reglas del conjunto y los
  casos negativos de skills se parametrizan en lugar de copiarse (D22, D26), y `TestSkillsDelRepositorio` pasa de
  `require.Contains` a `require.Subset` sobre `skillsExigidas`, primero con una skill (T021) y después con la exigencia
  de `legal-core` en rojo antes que sus ficheros (T022).
- **Matriz de skills en el job** (D23): dos trabajos en la misma ejecución con `fail-fast: false`, porque SC-015 exige
  «la misma ejecución»; recorrer las dos en `scripts/evals.sh` mezclaría los informes y doblaría el tiempo en serie.
- **`legal-core` delega el texto de norma en `boe-legislacion`, en un solo sentido** (Q5): su tabla de comandos lleva
  solo `territorio`, sus evals miden territorio, cobertura y no activación y ninguna lleva `citas`.
- **Las tareas `[datos]` se cierran con el esquema y la persona añade el material en la pausa** (T001, intento 2;
  T015): la pausa humana solo se dispara sobre un diff commiteado, así que dejar la tarea sin marcar hasta que el
  material exista la deja sin pausa jamás. T015 necesitó cuatro intentos porque sus dos piezas —el manifiesto y el
  `enum`— tenían ya un lector que las ponía en rojo sin las grabaciones, y solo una persona con red podía grabarlas.
- **El guion de reinstalación deja de nombrar las skills** (T022 → T028): con `legal-core` en el árbol, `make install`
  la enlaza también y `instalar-de-nuevo.txtar` exigía exactamente `boe-legislacion`; el guion es material de test y se
  cambió en su propia tarea `[datos]`, pasando a exigir una entrada por cada skill del árbol.
- **Los ejemplos de los casos negativos llevan el número 0** (T017, intento 2): la tabla de normas crece con cada hito
  —la LEC 1/2000 y la LOPDGDD 3/2018 eran dos de los ejemplos que «no están en la tabla»— y un ejemplo que pudiera entrar
  dejaría el caso sin fallo.
- **Ningún ADR** (T024): ADR 0017 ya decidió el origen de los datos y ADR 0006 la procedencia del sobre de un applet
  calculado; H6 no cambia ninguna decisión de arquitectura.
- **Sin exclusiones de lint** (D24): `aspecto` en singular, `configuración` y `autónomas` con tilde; los dos
  analizadores de `ids` en ficheros distintos con la comprobación de cifras compartida, para no disparar `dupl`.
- **La carga del territorio lee las filas en su forma sin el lector de YAML** (T029, tras el rojo de `ci` del intento
  1 de T026; `gates/tarea-T026.md` § 1 y `gates/tarea-T029.md`): en el runner, `territorio resolver Leganés` tardó
  235 ms con un máximo de 200 ms, porque cargar 830 KB de datos asignaba 36 MB en unas 756 000 asignaciones, el 72 %
  en el árbol de nodos que el lector construye para las filas. Se rechazó subir la cota, medir el mínimo de varias
  tiradas o sacar el cronometraje de `make ci`: el arreglo va a la causa y deja la carga con los datos reales en 9 ms y
  9,8 MB, con la misma salida byte a byte —la huella de Leganés es la de la ejecución `ci` del intento 1— y los
  mismos defectos, y sin tocar la cota, el guion, los datos, sus esquemas, el kernel, el applet ni `go.mod`.

## Revisión final

Ronda 1 (`gates/revision-a-r1.json`, `gates/revision-b-r1.json`): los dos jueces rechazaron, con seis motivos entre
los dos. Los seis están cerrados:

1. **La verificación del DIR3 era circular** (A y B, criterios `f` y `h`): comparaba el número de inscripción del REL
   con el código INE, la entrada de la derivación consigo misma, y le faltaba el caso con entidades locales menores.
   Rehecha contra el directorio del Punto de Acceso General, con los tres casos de FR-046 y siete coincidencias
   (*Decisiones*; `gates/verificacion-dir3.md`, nota final de `gates/tarea-T003.md`). `docs/SOURCES.md` describe ahora
   esa verificación y `quickstart.md` §9 vuelve a esperar las columnas de FR-047. `data/` no cambia.
2. **El `source` de los nombres de provincia y comunidad** (A y B): los 19 ficheros de comunidad declaran
   `ine.codigos-territoriales`, la fila de donde salieron esos nombres, en vez de `ine.municipios` (`c481451`, con el e2e,
   `data-model.md` y el CHANGELOG).
3. **La release es H19, no H6** (A y B): corregidas las afirmaciones de `Makefile`, `CONTRIBUTING.md` y `CHANGELOG.md`
   (`c481451`).
4. **El control que impedía añadir un territorio** (A): `solo-madrid-configurada` pasa a `madrid-configurada` y exige
   solo lo de FR-052; con boletines añadidos a otra comunidad, `make ci` sigue en verde (`c481451`).
5. **Ramas sin test** (A y B): `compilarEsquemaDeJerarquia` gana su test de las dos ramas de error y el paquete `data`
   pasa de 0 % a 100 % con `datos_test.go`; los tres tests de compilación de esquemas comparten ya una sola función
   (`c481451`).
6. **Inexactitudes de este cuerpo y de los artefactos** (A y B): la lista de las siete normas y la supresión `//nolint`
   declarada con su motivo (`e4c0b65`); el dígito de control de Leganés en `research.md` y `data-model.md` (`c481451`).
   `ComponerDIR3` por `DIR3DeAyuntamiento` quedó sin corregir y lo cerró la ronda 2.

Ronda 2 (`gates/revision-a-r2.json`, `gates/revision-b-r2.json`): los dos jueces rechazaron otra vez, con nueve
motivos entre los dos, todos corregibles y todos cerrados en `6d08752` y en el commit de este cuerpo:

1. **Una rama de error sin test** (A): la de `cuerpoDeLasLeyesVertebrales`. El caso nuevo
   `vertebrales-sin-normas-con-vertical` declara las referencias como `legal-core`, sin `normas` delante; el mutante
   que descarta el error de `LeerNormas` lo pone en rojo.
2. **La verificación del DIR3 atribuida al REL en `README.md` y `CHANGELOG.md`** (A): ahora dicen que el DIR3 sale
   del número de inscripción y que la regla se verificó contra el directorio DIR3 oficial. `CONTRIBUTING.md` dice que
   un refresco usa también las tablas de códigos del INE y vuelve a verificar la derivación contra DIR3 real.
3. **La fecha de `ine.codigos-territoriales` en `docs/SOURCES.md`** (A): 2026-09-20, el día de la descarga de T002
   —las copias que quedaron de esa descarga llevan esa fecha y los SHA-256 que registra `gates/tarea-T002.md`;
   comprobación anotada al final de `gates/tarea-T003.md`—, no el 2026-09-21, que es el de la revisión de la fila.
4. **`research.md`** (A): el `source` de los nombres de comunidad es `ine.codigos-territoriales`.
5. **`plan.md`, control 4** (B): la demostración es lo que `madrid-configurada` detecta, no añadir boletines a otra
   comunidad.
6. **Este cuerpo** (A y B): `DIR3DeAyuntamiento`, el comentario de cabecera, las cifras del *Alcance* (medidas en
   `6d08752`) y las 900 líneas de `TestTerritorioDelRepositorio`; y el cuerpo publicado en #39, que era anterior a
   `e4c0b65`, se volvió a sincronizar con `gh pr edit 39 --body-file` (`1d39df9`).

Ronda 3 (`gates/revision-a-r3.json`, `gates/revision-b-r3.json`): el juez B aprobó y el juez A rechazó con tres
motivos, todos de este cuerpo y bajo `specs/`, sin efecto sobre SC-015. Agotadas las dos rondas de corrección, Jorge
autorizó una tercera solo para ellos:

1. **El `Makefile`** (A): la viñeta del *Alcance* nombra las dos recetas que cambian, `skills-check` y `release`.
2. **«Sus SHA-256 coinciden»** (A): la comprobación que lo sostiene —las copias de la descarga de T002, con su fecha
   y sus hashes— queda anotada al final de `gates/tarea-T003.md`, y el punto 3 de la ronda 2 remite a ella.
3. **El resumen de `docs/SOURCES.md`** (A): la viñeta de *Documentación* incluye el cambio de fecha de la ronda 2,
   con su commit.

Como las dos rondas tocan ficheros fuera de `specs/`, la ejecución de aceptación se repitió: primero sobre `9a77c6b`
(36011479943) y, tras la ronda 2, sobre `6d08752` (*Pendientes*, SC-015; `gates/evals-cierre.md`).

## Pendientes

- **Regla de SC-015 sobre lo que puede cambiar después de la ejecución de aceptación** (plan, obligación 9): el commit
  del informe de evals tiene que ser de la rama del hito, y **la cabeza que se fusiona solo puede diferir de él en
  ficheros bajo `specs/008-h6-territorio-skill-legal/`**. La ejecución 35722605048, sobre `2b81164` (`feat(H6): T029`),
  dejó de cubrirla cuando la revisión final tocó código, datos y documentación fuera de `specs/` (`c481451`: `Makefile`,
  `CHANGELOG.md`, `CONTRIBUTING.md`, `data/datos.go` y su test, los 19 ficheros de comunidad, el e2e y cinco tests;
  y `9a77c6b`: la nota al pie del DIR3 en `docs/SOURCES.md`); se repitió por etiqueta, como H5.1, sobre `9a77c6b`
  (36011479943), y la ronda 2 volvió a tocarlos (`6d08752`: `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`,
  `docs/SOURCES.md` y un test). **La vigente es la 36019842457, sobre `6d08752`**, y desde ese commit la cabeza solo
  cambia bajo `specs/008-h6-territorio-skill-legal/`. Las tres anteriores —35714659803 sobre `47f3090`, 35722605048
  sobre `2b81164` y 36011479943 sobre `9a77c6b`— quedan registradas en `gates/evals-cierre.md` y no cuentan.
- **Aceptación** (SC-013, SC-015): registrada en `gates/evals-cierre.md` («Revisión final, ronda 2»). En la ejecución
  36019842457, sobre `6d08752`, las dos skills salen `aprobado` con `claude-sonnet-5` decidiendo, 3 repeticiones y
  umbral 2: las tres evals de `legal-core` dan 3 de 3 con los dos modelos, las 18 de `boe-legislacion` dan 3 de 3 con
  `claude-sonnet-5` sin que sus ficheros hayan cambiado frente a `main`, ninguna petición llega a la red de una fuente,
  y las seis respuestas de Tordesillas declaran los dos aspectos `no-configurado` sin nombrar ningún boletín que el
  applet no devolvió. `ci` en verde sobre `6d08752` (36019821690), con los cuatro estados de Codecov en verde y con
  medida.
- **La segunda cláusula de SC-013 la lee una persona, no el juez** (`gates/tarea-T026.md` § 2 del intento 1 y
  «Intento 3»): el esperado de territorio de FR-084 declara lo que tiene que aparecer en la respuesta, no lo que no
  puede aparecer, así que el juez no ve si una respuesta nombra un boletín que el applet no devolvió. En el intento 1
  una sesión de seis lo hizo («normalmente el BOCyL», «BOP de Valladolid»), con veredicto `aprobado`; en el intento 3,
  ninguna; en la 36011479943, una del modelo informativo (`haiku-4-5-20251001-02`: «el Boletín Oficial de Castilla y
  León», con mayúsculas, el nombre propio del boletín autonómico), con veredicto `pasa`, y ninguna de las tres del
  modelo que decide; en la vigente, 36019842457, ninguna de las seis. El spec no fijaba si esa cláusula se cuenta
  respuesta a respuesta o por serie y umbral como SC-015, y en la 36011479943 las dos lecturas daban distinto (5 de 6
  frente a 3 de 3 y 2 de 3). **Decisión de Jorge (2026-09-24): vale la lectura por serie y umbral**, la misma que el
  resto de las evals —el modelo que decide no falló ninguna vez y el informativo no bloquea—. En la vigente las dos
  coinciden. Para que la cláusula la
  juzgue la máquina con el mismo umbral que el resto haría falta un esperado que liste los boletines que la respuesta
  no puede nombrar, o que el juez compare los nombrados con los devueltos: cambia FR-084, `schemas/eval.yaml.json` y
  `internal/evals`, así que es una pieza aparte, después de este hito.
- **La guía del hito se corrigió en este cierre** (T025, intento 1; `gates/tarea-T025.md`), en tres textos de
  `quickstart.md` que no casaban con los datos congelados, ninguno un defecto del producto: la sonda de §11 contaba el
  comentario de cabecera de `data/normas.yaml` que nombra la marca `vertebral: true` (16 en lugar de 15) y pasó a
  anclarse al campo sangrado; el nombre ambiguo de §6, `Villanueva`, no existe como nombre exacto en la relación (da
  3, no 2) y ahora se toma del propio registro, como el código ausente; y la expectativa de §9 sobre el registro de
  verificación, que T025 ajustó al registro de la pausa de T003 y la revisión final devolvió a lo que piden FR-046 y
  FR-047 al rehacer la verificación del DIR3 (*Decisiones*).
  `data/normas.yaml` no se tocó: quitarle la frase al comentario para que una sonda floja acertara sería arreglar el
  síntoma.
- **`govulncheck` y un módulo requerido**: el mismo `make ci` informa, sin fallar, de una vulnerabilidad en un módulo
  que el código no llama, `GO-2026-5970` («Infinite loop on invalid input in golang.org/x/text»; encontrada en
  `golang.org/x/text@v0.14.0`, corregida en v0.39.0), según `govulncheck -show verbose` sobre `2df74db`. H6 no toca
  `go.mod` ni `go.sum`, así que viene de `main`, como en H5.1; subir la versión queda fuera de este hito.
- **Integración continua y Codecov**: leer los estados de la propuesta de cambio antes de fusionar; la rama principal
  no impide fusionar en rojo. Fusionar es humano (ADR 0007).
- **Backlog, fuera de este hito**: configurar otras comunidades en `data/territorio/comunidades/` (grupo
  «territorios»: BOCYL, BOP multiprovinciales, forales), que es rellenar `boletines` sin tocar código; `territorio`
  emitiendo `Municipio` y `Organo` al grafo desde H7 (ADR 0014); `legal-core` v1 con plazos y recursos en H9.
