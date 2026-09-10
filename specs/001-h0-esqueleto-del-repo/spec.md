# Feature Specification: H0 · Esqueleto del repo y gates de CI

**Feature Branch**: `h0-esqueleto-del-repo`

**Created**: 2026-09-09

**Status**: Draft

**Input**: Sección `#### H0 · Esqueleto del repo y gates de CI` de `docs/ROADMAP.md` (modo desatendido).

## Resumen

Hoy el repositorio solo contiene documentos de diseño: no hay `go.mod`, ni binario, ni tests, ni CI. H0 convierte ese repositorio en **un repositorio vacío pero blindado**: la estructura mínima de un módulo Go publicable y, sobre todo, el conjunto de controles automáticos que toda línea de Go que entre después tendrá que atravesar, tanto en la máquina de quien contribuye como en la integración continua.

El valor de H0 no es funcionalidad para un usuario final: es que a partir de aquí **ningún cambio pueda entrar sin pasar por formato, lint, tests con detector de carreras, análisis de vulnerabilidades, SAST, detección de secretos y verificación de la cadena de suministro**, y que ese veredicto sea reproducible en local y en la PR.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Veredicto de calidad reproducible en local (Priority: P1)

Quien contribuye al proyecto (persona o agente) clona el repositorio, ejecuta un único comando y obtiene, sin configuración adicional y sin acceso a la red más allá de la descarga de dependencias, un veredicto de si su árbol de trabajo cumple todos los controles del proyecto. El mismo comando es el que ejecutará la integración continua, de modo que un verde en local predice un verde en la PR.

**Why this priority**: Es la entrega nuclear del hito (`make build test lint vuln` en verde) y la precondición de todos los hitos posteriores: el ritual por hito (`docs/ROADMAP.md` §6) exige `make ci` local antes de abrir la PR, y la *Definition of Done* (§1) se apoya entera en esos comandos. Sin esto, ningún otro control es exigible.

**Independent Test**: Se puede probar por completo clonando el repositorio en una máquina limpia con Go instalado y ejecutando los comandos de construcción, test, lint y análisis de vulnerabilidades: todos terminan con éxito y sin avisos. Entrega valor por sí sola aunque no exista aún ningún applet.

**Acceptance Scenarios**:

1. **Given** un clon limpio del repositorio en la revisión de H0, **When** se ejecuta la orden de construcción, **Then** se produce un binario ejecutable del proyecto y el comando termina con éxito.
2. **Given** un clon limpio, **When** se ejecuta la orden de tests, **Then** los tests se ejecutan con el detector de carreras activado y terminan con éxito.
3. **Given** un clon limpio, **When** se ejecuta la orden de lint, **Then** el análisis estático completo termina sin hallazgos y con éxito.
4. **Given** un clon limpio, **When** se ejecuta la orden de análisis de vulnerabilidades, **Then** no se reportan vulnerabilidades conocidas en el módulo ni en sus dependencias y el comando termina con éxito.
5. **Given** un clon limpio, **When** se ejecuta la orden agregada de integración continua, **Then** ejecuta comprobación de formato, lint, tests con detector de carreras, análisis de vulnerabilidades, comprobación de esquemas, detección de secretos, verificación de la integridad de los módulos y comprobación de dependencias saneadas, y termina con éxito.
6. **Given** un árbol de trabajo con un fichero Go sin formatear, **When** se ejecuta la orden agregada de integración continua, **Then** el comando falla, indica el fichero afectado y deja el árbol de trabajo sin modificar (la comprobación de formato dentro de esa orden verifica, no corrige).
7. **Given** un clon limpio, **When** se ejecuta la orden agregada de integración continua, **Then** termina con éxito sin haber modificado ningún fichero del árbol de trabajo.

---

### User Story 2 - La PR es la que bloquea, y bloquea rápido (Priority: P2)

Quien mantiene el proyecto abre una propuesta de cambio. La integración continua ejecuta el mismo conjunto de controles y decide: si el cambio viola una regla del proyecto, la propuesta queda bloqueada con el motivo señalado; si el cambio está limpio, el veredicto llega en un tiempo lo bastante corto como para no romper el ritmo de trabajo.

**Why this priority**: Es el criterio de aceptación literal del hito y lo que hace exigibles los controles frente a la buena voluntad. Depende de la P1 (los comandos deben existir antes de poder invocarse desde la integración continua), pero es lo que cierra el hito.

**Independent Test**: Se puede probar de forma aislada abriendo dos propuestas de cambio contra el repositorio: una que introduce deliberadamente una violación y otra que no, y comprobando el veredicto y la duración de cada una.

**Acceptance Scenarios**:

1. **Given** el repositorio con los controles de H0 activos, **When** se abre una propuesta de cambio que añade una escritura directa a la salida estándar (`fmt.Println`) en un fichero bajo `internal/`, **Then** el control de lint falla y la propuesta queda bloqueada.
2. **Given** el repositorio con los controles de H0 activos, **When** se abre una propuesta de cambio que no viola ninguna regla, **Then** todos los controles pasan en menos de 3 minutos.
3. **Given** una propuesta de cambio que introduce un secreto (credencial, token) en el diff, **When** se ejecutan los controles, **Then** el control de detección de secretos falla y la propuesta queda bloqueada.
4. **Given** una propuesta de cambio en la que el fichero de dependencias no está saneado (declara dependencias que no se usan o falta alguna que sí), **When** se ejecutan los controles, **Then** el control de dependencias mínimas falla y señala la diferencia.
5. **Given** una propuesta de cambio que añade una dependencia nueva sin justificación, **When** se revisa la propuesta, **Then** la guía de contribución del repositorio exige esa justificación de forma explícita.
6. **Given** una propuesta de cambio con una vulnerabilidad conocida en una dependencia, **When** se ejecutan los controles, **Then** el análisis de vulnerabilidades falla.

---

### User Story 3 - Saber exactamente qué binario se está ejecutando (Priority: P3)

Cualquier persona o agente que tenga el binario instalado puede preguntarle qué es: qué versión declara, de qué revisión exacta del repositorio procede y cuándo se construyó. Con eso, un informe de error o una consulta legal reproducida más tarde son atribuibles a un artefacto concreto.

**Why this priority**: Forma parte de la entrega literal del hito y es el único comportamiento observable del binario en H0, pero no bloquea ninguno de los controles anteriores.

**Independent Test**: Se puede probar de forma aislada construyendo el binario e invocando su subcomando de versión, comprobando que los tres datos aparecen y que la revisión coincide con la del árbol construido.

**Acceptance Scenarios**:

1. **Given** el binario construido mediante la orden de construcción del proyecto, **When** se invoca `kitlegal version`, **Then** la salida contiene la versión, el identificador de la revisión (commit) y la fecha de construcción.
2. **Given** el binario construido, **When** se invoca `kitlegal version`, **Then** el comando termina con código de salida 0.
3. **Given** un binario construido desde una revisión concreta del repositorio, **When** se compara el commit impreso con el de esa revisión, **Then** coinciden.

---

### User Story 4 - Entender el proyecto y sus reglas sin preguntar (Priority: P4)

Alguien que llega al repositorio por primera vez —una persona externa, o un agente en una sesión nueva— encuentra en el propio repositorio qué es el proyecto, bajo qué licencia puede usarlo, cómo contribuir y qué controles se le van a aplicar, y por qué se tomaron las decisiones de arquitectura que condicionan todo lo demás.

**Why this priority**: Es parte del alcance literal del hito (licencia, README, guía de contribución, changelog y cuatro ADR) y condición de la *Definition of Done* (§1.6 y §1.7), pero no bloquea la ejecución de los controles.

**Independent Test**: Se puede probar de forma aislada revisando el repositorio: existen los documentos, la licencia es la decidida y cada ADR registra su decisión con contexto y consecuencias.

**Acceptance Scenarios**:

1. **Given** el repositorio en la revisión de H0, **When** se consulta el fichero de licencia, **Then** contiene Apache-2.0 como licencia del núcleo.
2. **Given** el repositorio, **When** se consulta la guía de contribución, **Then** documenta el ritual por hito, la convención de commits, el versionado semántico, los controles que se ejecutarán y la exigencia de justificar toda dependencia nueva.
3. **Given** el repositorio, **When** se consulta el changelog, **Then** existe una sección *Unreleased* lista para recibir entradas.
4. **Given** el repositorio, **When** se consulta el directorio de decisiones de arquitectura, **Then** existen las cuatro decisiones fundacionales (multicall, SQLite sin cgo, no ingesta masiva de CENDOJ, frontera humana), cada una con contexto, decisión y consecuencias.

---

### Edge Cases

- **El repositorio no tiene código de producto todavía.** Los controles que analizan código deben terminar con éxito sobre un módulo con un único punto de entrada mínimo, sin reportar «no hay paquetes que analizar» como error.
- **No existen aún `schemas/`, `data/`, `skills/`, `testdata/` ni configuración de release.** Las órdenes del `Makefile` que dependen de ellos se invocan igualmente desde la integración continua o desde quien contribuye: `test-e2e`, `schema-check` y `skills-sync` anuncian el objeto ausente y el hito que lo aporta y terminan con éxito (`schema-check` pasa vacíamente dentro de `ci`); `test-integration` ejecuta ya el comando real sobre un conjunto vacío de tests; `release` falla con código ≠ 0 porque una acción con efectos externos que no puede ejecutarse no debe simular éxito.
- **Umbrales de cobertura sobre un repositorio prácticamente sin código.** El umbral global del 70 % es exigible desde H0 y se cumple con los tests que el propio hito aporta sobre el comportamiento de `version`; el umbral de `internal/core/**` no se evalúa mientras no existan ficheros en esa ruta.
- **El formato como gate y no como corrección silenciosa.** Corregir el formato es útil en la máquina de quien contribuye, pero un veredicto que modifica el árbol para poder aprobarlo no es un gate. La orden de formato mutante y su comprobación no mutante son dos modos del mismo control; solo el segundo participa del veredicto (FR-012).
- **`fmt.Println` está prohibido en `internal/` pero el punto de entrada tiene que escribir la versión en algún sitio.** El control debe distinguir el paquete autorizado del prohibido, no prohibir la escritura en toda la base de código: la regla se acota por ruta a `internal/**`, que en H0 no contiene ningún fichero, y `cmd/` queda autorizado.
- **El detector de secretos ante un falso positivo.** Debe existir una vía documentada y trazable para excluir una coincidencia legítima, sin desactivar el control.
- **Contribución desde una máquina sin las herramientas instaladas.** El proyecto aprovisiona por sí mismo las herramientas de los controles, de modo que el único prerrequisito es una cadena Go y `git`; ante su ausencia el proyecto debe indicar de forma inequívoca qué falta y cómo obtenerlo, en lugar de fallar con un error opaco del intérprete de órdenes.
- **Los ganchos de pre-commit no están instalados.** Un cambio puede llegar a la propuesta sin haber pasado el pre-commit; la integración continua debe cubrir los mismos controles y ser la autoridad final.
- **Ejecución de los controles sin conexión a la red.** El análisis de vulnerabilidades consulta una base de datos remota; su comportamiento sin red debe ser explícito y no confundirse con un verde.

## Requirements *(mandatory)*

### Requisitos funcionales

#### Estructura del módulo y binario

- **FR-001**: El repositorio DEBE declarar un módulo Go con la ruta `github.com/jmorenobl/kitlegal` y fijar la versión estable actual de Go.
- **FR-002**: El repositorio DEBE contener un punto de entrada mínimo en `cmd/kitlegal/main.go`, sin funcionalidad de producto más allá de lo exigido por FR-003. Todo el código de H0 DEBE vivir en `cmd/kitlegal/`; H0 NO DEBE crear ningún paquete bajo `internal/`.
- **FR-003**: El binario DEBE ofrecer un subcomando `version` que imprima la versión declarada, el identificador de la revisión (commit) y la fecha de construcción, y termine con código de salida 0.
- **FR-004**: Los tres datos de FR-003 DEBEN residir en el paquete del punto de entrada e inyectarse en el momento de la construcción por la orden de construcción del proyecto, de modo que el binario producido por esa orden nunca los muestre como desconocidos.
- **FR-040**: El comportamiento de FR-003 —los tres datos y el código de salida— DEBE quedar cubierto por tests automáticos del módulo que ejecuta la orden `test`, y la cobertura resultante en la revisión de H0 DEBE superar el umbral global declarado en FR-029 sin exclusiones ni excepciones. Cómo se estructura el punto de entrada para hacerlo testable es decisión del plan (ver «Clarifications»).
- **FR-005**: El fichero de dependencias DEBE estar saneado y acompañado de su fichero de sumas de verificación, y la integridad de los módulos descargados DEBE ser verificable.

#### Órdenes del proyecto

- **FR-006**: El repositorio DEBE exponer un `Makefile` con, al menos, las órdenes `build`, `test`, `test-integration`, `test-e2e`, `lint`, `fmt`, `vuln`, `schema-check`, `skills-sync`, `release`, `ci` e `install`.
- **FR-007**: La orden `ci` DEBE ejecutar, como mínimo, la comprobación de formato, `lint`, `test` (con detector de carreras), `vuln`, `schema-check`, la detección de secretos (FR-018), la verificación de la integridad de los módulos (FR-019) y la comprobación de que el fichero de dependencias está saneado (FR-020), y DEBE fallar si cualquiera de ellas falla. Con eso, `ci` cubre por sí sola todos los controles de FR-012 a FR-020 exigibles sobre un árbol de trabajo, que es lo que hace cierta la equivalencia local ↔ integración continua de SC-004. La comprobación de formato dentro de `ci` es la no mutante de FR-012: `ci` NO DEBE modificar ningún fichero del árbol de trabajo, de modo que su veredicto no dependa de haber corregido antes lo que juzga. La orden `release` NO forma parte de `ci` (FR-011).
- **FR-008**: La orden `test` DEBE ejecutar los tests con el detector de carreras activado y DEBE generar el perfil de cobertura que el flujo `ci` publica (FR-029).
- **FR-009**: Las órdenes `build`, `test`, `lint` y `vuln` DEBEN terminar con éxito sobre el repositorio en la revisión de H0.
- **FR-010**: Cuando un prerrequisito externo requerido por una orden no esté disponible, la orden DEBE fallar con un mensaje que lo identifique y explique cómo obtenerlo. Dado FR-042, los únicos prerrequisitos externos en H0 son la cadena de herramientas Go y `git`, cuya presencia y versión DEBE comprobar el `Makefile`.
- **FR-011**: Las órdenes del `Makefile` cuyo objeto (esquemas, skills, tests e2e, configuración de release) no existe todavía en H0 DEBEN comportarse de forma explícita y determinista, sin no-ops silenciosos:
  - `test-integration` DEBE ejecutar ya el comando real de tests de integración con detector de carreras y etiqueta de compilación `integration`, que en H0 opera sobre un conjunto vacío de tests y termina con éxito.
  - `test-e2e`, `schema-check` y `skills-sync` DEBEN imprimir una línea que nombre el objeto ausente y el hito que lo aporta, y terminar con éxito; el hito correspondiente sustituye su cuerpo por el comando real.
  - `release` DEBE terminar con código de salida distinto de 0 y un mensaje que indique que no estará configurada hasta el hito del release firmado, por ser una acción con efectos externos que no puede simular éxito; NO DEBE formar parte de `ci` ni del flujo `nightly`.
  - Cada una de estas órdenes DEBE documentarse en `CONTRIBUTING.md` indicando que existe pero recibe su contenido en un hito posterior.
- **FR-041**: La orden `install` DEBE instalar el binario `kitlegal` en el directorio de binarios del usuario aplicando los mismos datos de versión inyectados que la orden `build` (FR-004). El enlace de las skills no forma parte de `install` en H0.
- **FR-042**: Las herramientas Go de los controles (`golangci-lint`, `govulncheck`, `gofumpt`, `gitleaks`, `lefthook`) DEBEN cumplir, sea cual sea el mecanismo que elija el plan: (a) versión fijada y explícita, idéntica en la máquina de quien contribuye y en la integración continua; (b) integridad verificable con el fichero de sumas del módulo; (c) actualización automática por Dependabot (FR-028); (d) obtención y ejecución sin más prerrequisitos que la cadena Go y `git` (FR-010), sin instalación manual previa; (e) sin añadir sus dependencias al fichero de dependencias del producto (FR-005). La integración continua DEBE ejecutar los controles a través de las mismas órdenes del `Makefile` que se ejecutan en local, y no mediante descargas de binarios ajenas al módulo. El mecanismo concreto queda para el plan (ver «Clarifications»).

#### Controles de calidad (§3 de `docs/ROADMAP.md`, marcados H0)

- **FR-012**: El repositorio DEBE aplicar `gofumpt` y `goimports` como formateadores, integrados en la configuración de lint, y DEBE ofrecer sobre ellos —misma herramienta, misma configuración— dos modos: uno **mutante**, que es la orden `fmt` y corrige el árbol de trabajo, y uno de **verificación**, que no modifica ningún fichero y falla nombrando los ficheros que no están formateados. El modo de verificación es el que ejecutan la orden `ci` (FR-007) y la integración continua (FR-025), de modo que el formato es un gate demostrable: un fichero sin formatear bloquea la propuesta de cambio, y retirar el control lo dejaría pasar.
- **FR-013**: El repositorio DEBE incluir una configuración de `golangci-lint` estricta que active al menos: `errcheck`, `govet`, `staticcheck`, `revive`, `gocritic`, `errorlint`, `exhaustive`, `nilerr`, `bodyclose`, `noctx`, `contextcheck`, `sqlclosecheck`, `rowserrcheck`, `nolintlint`, `misspell` (español e inglés), `gocyclo`, `dupl`, `testifylint`, `thelper`, `paralleltest`, `tparallel` y `gosec`.
- **FR-014**: La configuración de lint DEBE activar `forbidigo` con una única regla que prohíba las escrituras directas a la salida estándar (`fmt.Print*`), acotada por ruta a `internal/**`, de modo que un `fmt.Println` bajo `internal/` falle en lint y `cmd/` quede autorizado en H0. H0 NO DEBE activar `depguard` ni las demás reglas de arquitectura, que corresponden a H1.
- **FR-015**: La integración continua DEBE ejecutar los tests con el detector de carreras.
- **FR-016**: La integración continua DEBE ejecutar `govulncheck` en cada propuesta de cambio y fallar si aparece una vulnerabilidad conocida.
- **FR-017**: El repositorio DEBE incorporar análisis estático de seguridad: `gosec` como parte del lint y un análisis CodeQL con periodicidad semanal.
- **FR-018**: El repositorio DEBE ejecutar detección de secretos (`gitleaks`) tanto en el gancho de pre-commit como dentro de la orden `ci` (FR-007) y, por tanto, en la integración continua, que ejecuta esa misma orden (FR-025, FR-042).
- **FR-019**: El repositorio DEBE controlar la cadena de suministro con verificación de la integridad de los módulos (`go mod verify`) dentro de la orden `ci` (FR-007), fichero de sumas obligatorio y actualizaciones automáticas de dependencias con periodicidad semanal (Dependabot, FR-028).
- **FR-020**: La orden `ci` (FR-007) y, por tanto, la integración continua DEBEN comprobar que el fichero de dependencias está saneado (`go mod tidy -diff`) y fallar señalando la diferencia en caso contrario.
- **FR-021**: La guía de contribución (`CONTRIBUTING.md`, FR-032) DEBE exigir justificación explícita para toda dependencia nueva, y DEBE recoger la estructura de propuesta de cambio del ritual por hito (`docs/ROADMAP.md` §6.3). H0 no añade fichero de plantilla de propuesta de cambio, que no figura en el alcance del hito (ver «Fuera de alcance»).
- **FR-022**: El repositorio DEBE configurar ganchos de pre-commit (`lefthook`) que ejecuten formato —aquí sí el modo mutante de FR-012, porque el gancho corrige antes de confirmar, no emite el veredicto—, lint en modo rápido, detección de secretos y comprobación de dependencias saneadas. El gancho no sustituye a `ci`: la autoridad final es la integración continua (ver «Edge Cases»).
- **FR-023**: El repositorio DEBE documentar y adoptar Conventional Commits y versionado semántico, y mantener un `CHANGELOG.md` con sección *Unreleased*.
- **FR-024**: El repositorio DEBE registrar las decisiones de arquitectura en `docs/ADR/` con formato MADR corto.

#### Automatización de la integración continua

- **FR-025**: El repositorio DEBE definir un flujo `ci` que se dispare en cada propuesta de cambio y ejecute los controles de FR-012 a FR-020 invocando las órdenes del `Makefile` (FR-042). Ese flujo NO DEBE aplicar ningún control de FR-012 a FR-020 por una vía distinta de esas órdenes; sus pasos propios se limitan a preparar el entorno (obtener el código, instalar Go, restaurar la caché) y a publicar el perfil de cobertura que produce la orden `test` (FR-029). Ese flujo DEBE conservar entre ejecuciones la caché de dependencias y de herramientas, para que el veredicto de una propuesta limpia se obtenga en el tiempo que exige SC-006.
- **FR-026**: El repositorio DEBE definir un flujo `codeql` de análisis de seguridad con periodicidad semanal.
- **FR-027**: El repositorio DEBE definir un flujo `nightly` programado con periodicidad diaria que ejecute la orden agregada de integración continua sobre la rama principal. Ese flujo NO DEBE ejecutar ninguna verificación contra fuentes reales ni ningún otro control que no esté ya especificado en este hito.
- **FR-028**: El repositorio DEBE definir la configuración de Dependabot con revisión semanal, que DEBE cubrir tanto las dependencias del producto como las versiones fijadas de las herramientas de los controles (FR-042) y las acciones que usen los flujos de integración continua.
- **FR-029**: El repositorio DEBE definir la configuración de cobertura en `codecov.yml` declarando desde H0 los dos umbrales de la *Definition of Done* —global ≥ 70 % y un componente para `internal/core/**` ≥ 85 %— como estado que falla, no informativo. El umbral global DEBE cumplirse ya en la revisión de H0 con los tests del propio hito (FR-040); el componente `internal/core/**` no se evalúa mientras no existan ficheros en esa ruta y empieza a aplicarse por sí solo cuando aparezcan, sin configuración condicional que haya que editar después. El flujo `ci` DEBE publicar el perfil de cobertura generado por la orden `test`.

#### Documentación fundacional

- **FR-030**: El repositorio DEBE incluir `LICENSE` con Apache-2.0 como licencia del núcleo, conforme a la decisión open-core.
- **FR-031**: El repositorio DEBE incluir `README.md` que describa qué es `kitlegal`, cómo construirlo y cómo ejecutar los controles.
- **FR-032**: El repositorio DEBE incluir `CONTRIBUTING.md` que documente el ritual por hito, los controles exigidos, la convención de commits y la exigencia de FR-021.
- **FR-033**: El repositorio DEBE incluir `docs/ADR/0001-multicall.md`, que registre la decisión de un único ejecutable multicall despachado por `os.Args[0]` o el primer argumento.
- **FR-034**: El repositorio DEBE incluir `docs/ADR/0002-sqlite-sin-cgo.md`, que registre la decisión de usar SQLite sin cgo.
- **FR-035**: El repositorio DEBE incluir `docs/ADR/0003-no-cendoj-masivo.md`, que registre que de CENDOJ solo se resuelven ECLI y metadatos, nunca ingesta masiva ni scraping del buscador.
- **FR-036**: El repositorio DEBE incluir `docs/ADR/0004-frontera-humana.md`, que registre que solo se automatizan fuentes públicas y que toda acción con identidad termina en «fichero listo para firmar» y código de salida 6.

#### Invariantes del proyecto que H0 no puede violar

- **FR-037**: El código introducido en H0 NO DEBE realizar ninguna petición de red ni escribir en disco fuera del directorio de construcción.
- **FR-038**: El código introducido en H0 NO DEBE contener `panic` en rutas de usuario.
- **FR-039**: Los códigos de salida estables del proyecto (0 ok · 2 args · 3 no encontrado · 4 fuente no disponible · 5 rate-limited/TOS · 6 requiere identidad humana) NO DEBEN ser contradichos por el punto de entrada de H0.

### Criterios del hito, literales

Se transcriben aquí, sin alterar, los apartados «Controles» y «Aceptación» de `#### H0 · Esqueleto del repo y gates de CI` en `docs/ROADMAP.md`:

- **Controles**: todos los de §3 marcados H0.
- **Aceptación**: PR de prueba con un `fmt.Println` en `internal/` falla en lint; PR limpia pasa en < 3 min.

La transcripción de «Controles» se desarrolla en FR-012 a FR-024; la de «Aceptación», en SC-005 y SC-006.

## Success Criteria *(mandatory)*

### Resultados medibles

- **SC-001**: Sobre un clon limpio del repositorio, las cuatro órdenes de construcción, tests, lint y análisis de vulnerabilidades terminan con éxito, sin hallazgos y sin avisos.
- **SC-002**: La orden agregada de integración continua ejecuta comprobación de formato, lint, tests con detector de carreras, análisis de vulnerabilidades, comprobación de esquemas, detección de secretos, verificación de módulos y comprobación de dependencias saneadas; termina con éxito y deja el árbol de trabajo sin modificar (no hay diferencias con `git status` antes y después).
- **SC-003**: `kitlegal version` imprime los tres datos —versión, commit y fecha— y termina con código de salida 0; el commit impreso coincide con el de la revisión construida.
- **SC-004**: El veredicto local y el de la integración continua coinciden en los dos sentidos: un árbol que pasa la orden agregada `ci` en local pasa los controles de la propuesta de cambio, y un árbol que falla un control en la propuesta de cambio falla también en local. La coincidencia es verificable porque la orden `ci` incluye todos los controles de FR-012 a FR-020 —formato, lint, race, vulnerabilidades, SAST vía lint, secretos, verificación de módulos y dependencias saneadas— y la integración continua no ejecuta ningún control fuera de esas órdenes (FR-007, FR-025, FR-042), con la misma versión fijada de cada herramienta. Quedan fuera de esta equivalencia, por no operar sobre un árbol de trabajo, el análisis CodeQL semanal (FR-026) y Dependabot (FR-028).
- **SC-005**: Una propuesta de cambio de prueba que añade un `fmt.Println` en un fichero bajo `internal/` falla en el control de lint. *(criterio de aceptación literal del hito)*
- **SC-006**: Una propuesta de cambio limpia pasa todos los controles en menos de 3 minutos. *(criterio de aceptación literal del hito; condición de medida en «Assumptions», garantizada por la caché que exige FR-025)*
- **SC-007**: Cada uno de los once controles de `docs/ROADMAP.md` §3 marcados H0 —formato, lint, race detector, vulnerabilidades, SAST, secretos, cadena de suministro, dependencias mínimas, pre-commit, commits/versiones y ADR— está activo y es demostrable con una comprobación concreta que falla si se retira el control.
- **SC-008**: Todos los ficheros y directorios enumerados en el «Alcance» de H0 (`docs/ROADMAP.md`) existen en el repositorio y cumplen su requisito funcional correspondiente (FR-001 a FR-042), sin cifra ni recorte: `go.mod`, `cmd/kitlegal/main.go`, `Makefile`, `.golangci.yml`, `lefthook.yml`, los tres flujos `.github/workflows/{ci,codeql,nightly}.yml`, `.github/dependabot.yml`, `codecov.yml`, `LICENSE`, `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md` y los cuatro ADR `docs/ADR/{0001-multicall,0002-sqlite-sin-cgo,0003-no-cendoj-masivo,0004-frontera-humana}.md`. A ellos se añaden los artefactos que exigen otros requisitos de este spec: el fichero de sumas de verificación (FR-005), los tests que cubren el punto de entrada (FR-040) y los artefactos donde quedan fijadas las versiones de las herramientas (FR-042). Ninguno queda vacío ni como marcador de posición sin contenido.
- **SC-009**: Una propuesta de cambio que introduce un secreto, una dependencia con vulnerabilidad conocida o un fichero de dependencias sin sanear queda bloqueada por el control correspondiente, que forma parte de la orden agregada `ci` (FR-007) y por tanto produce el mismo veredicto ejecutado en local.
- **SC-010**: Alguien que llega por primera vez puede construir el proyecto y ejecutar todos sus controles siguiendo únicamente `README.md` y `CONTRIBUTING.md`, sin consultar a nadie y sin más prerrequisitos que Go y `git`.
- **SC-011**: La revisión de H0 supera el umbral global de cobertura declarado (≥ 70 %) con sus propios tests, de modo que el control de cobertura está activo y en verde sin excepciones ni exclusiones.
- **SC-012**: Las órdenes del `Makefile` cuyo objeto aún no existe se comportan como declara FR-011: `test-integration`, `test-e2e`, `schema-check` y `skills-sync` terminan con éxito anunciando lo que falta, y `release` falla con un mensaje que nombra el hito que la dotará de contenido.

## Assumptions

- **Los «usuarios» de este hito son quienes contribuyen** (personas y agentes) y la propia integración continua. H0 no entrega valor a un usuario final del producto legal; ese valor llega a partir de H4.
- **Plataforma de integración continua**: GitHub Actions, implícito en el alcance del hito (`.github/workflows/`, `.github/dependabot.yml`) y en la ruta del módulo `github.com/jmorenobl/kitlegal`.
- **`cmd/` es paquete autorizado para escribir en la salida estándar en H0.** La prohibición de la *Definition of Done* §1.3 y el criterio de aceptación del hito se refieren a `internal/`; `internal/render` como único escritor de stdout es una regla que entra con `internal/render`, en H1.
- **`internal/` no existe en el repositorio en H0**, así que la regla de FR-014 se configura por ruta y se demuestra sobre la propuesta de cambio de prueba desechable, no sobre código propio del hito. Mover los datos de versión al kernel CLI cuando exista (H1) es un cambio trivial que decide ese hito.
- **Los tests de H0 cubren el punto de entrada desde su propio paquete**, sin extraer paquetes que el hito no necesita y sin crear `internal/` (FR-002). La forma concreta de hacerlo testable queda para el plan.
- **La salida de `kitlegal version` en H0 es texto plano, sin el sobre `{ok, fuente, url, fecha_consulta, hash, data}`.** El sobre y las banderas globales (`--json`, `--describe`…) son alcance de H1 y no se anticipan aquí (constitución, *Criterio de decisión autónoma* §2).
- **«Go estable actual»** se interpreta como la última versión estable publicada de Go en el momento de implementar el hito, tal y como fija la constitución en *Restricciones técnicas*.
- **Construcción con `CGO_ENABLED=0` y `-trimpath`**, según *Restricciones técnicas* de la constitución.
- **El umbral de 3 minutos** se mide sobre la ejecución completa de los controles de la propuesta de cambio en un ejecutor estándar de la plataforma, con la caché de dependencias y de herramientas ya poblada, que es el estado normal del flujo porque FR-025 obliga a conservarla entre ejecuciones. La primera ejecución tras invalidarse esa caché (cambio de versión de Go o de una herramienta) puede excederlo; el plan DEBE medir ambos escenarios y dejar constancia del número en frío, y si el escenario en caliente no bastara para cumplir el criterio literal del hito, es el plan el que debe corregir el diseño del flujo, no el criterio.
- **La comprobación del criterio de aceptación se realiza con propuestas de cambio de prueba desechables**, no con cambios que vayan a integrarse en `main`.
- **La fusión a `main` y el release son acciones humanas** (constitución, *Flujo de trabajo y gates* §4); H0 configura los controles, no los automatiza hasta la fusión.
- **Los ficheros generados por herramientas de terceros** (licencia Apache-2.0) se incorporan en su forma canónica y no se editan.

## Dependencias

- Ninguna sobre otros hitos: H0 es el primero de la fase 0 y no depende de ningún artefacto previo del repositorio.
- Requiere una cadena de herramientas Go instalada y acceso a los registros públicos de módulos Go y de vulnerabilidades para que los controles se ejecuten.
- Requiere que el repositorio esté alojado en la plataforma de integración continua asumida más arriba para que los flujos automáticos se ejecuten.

## Fuera de alcance

Todo lo siguiente está definido en `docs/ROADMAP.md`, `CLAUDE.md`, `refs/` o la constitución, pero **no** pertenece al alcance de H0 y **no se implementa** en este hito:

- **Kernel CLI (H1)**: `internal/cli`, `internal/app`, `internal/core/schema`, `internal/render`, el registro y despacho de applets, las banderas globales (`--json`, `--timeout`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto`, `--verbose`), el sobre de salida, los errores tipados con su mapeo a códigos de salida y el applet de ejemplo `echo`.
- **Reglas de arquitectura completas (H1)**: `depguard` con las reglas de dependencia de §2 (`net/http` solo en `httpx`, SQLite solo en `cache|store|graph`, `os.Exit` solo en `cli` y `cmd`, `core` sin adaptadores), el test `internal/arch_test.go` y la ampliación de `forbidigo` que restringe la escritura en stdout a `internal/render`; en H0 solo se aborda lo mínimo que exige el criterio de aceptación (FR-014).
- **Cualquier paquete bajo `internal/` (H1 en adelante)**: H0 no crea el directorio `internal/` ni ningún paquete dentro de él, tampoco para alojar los datos de versión ni material de verificación del propio control de lint.
- **Tests unitarios, e2e y de integración de producto (H1, H3)**: `testify`, `testscript`, fixtures en `testdata/`, golden files. H0 solo garantiza que las órdenes que los ejecutarán existen y funcionan.
- **Cliente HTTP (H2)**: `internal/httpx`, User-Agent, `robots.txt`, rate limit, retries, grabación y reproducción de fixtures.
- **Caché y almacenamiento (H3, H12)**: `internal/cache`, `internal/store`, `internal/graph`, SQLite, migraciones.
- **Adaptadores de fuentes (H4 en adelante)**: `internal/source/boe` y cualquier otra fuente; `docs/SOURCES.md` y `scripts/verify-sources.sh`.
- **Skills, packs y datos (H5)**: `skills/`, `packs/`, `data/*.yaml`, `scripts/skills-sync.sh`, `evals/`, comprobación de deriva de `references/`. H0 solo declara la orden `skills-sync`; el enlace de `skills/*` desde la orden `install` también corresponde a H5.
- **Esquemas y contratos (H4 borrador, H11)**: `schemas/*.json`, generación desde `--describe`, tests de contrato. H0 solo declara la orden `schema-check`.
- **Release firmado (H6)**: `.goreleaser.yaml`, artefactos multiplataforma, checksums, SBOM (syft), firma cosign, provenance, changelog generado automáticamente, `install.sh`, tap de Homebrew, flujo de release por etiqueta y smoke post-release. H0 solo declara la orden `release`, que falla de forma explícita hasta ese hito (FR-011), y mantiene el `CHANGELOG.md` a mano.
- **Fuzzing, property tests, benchmarks y `apidiff`** (H7, H8, H9, H16, H24).
- **Evals de skills con modelo (H5) y jueces LLM**: pertenecen al workflow de spec-kit y a hitos posteriores, no a los gates de CI de H0.
- **Smoke contra red (H4)**: el flujo `nightly` de H0 no ejecuta verificación de fuentes reales.
- **Observabilidad (H1)**: `log/slog`, `--verbose`, `KITLEGAL_LOG`.
- **Cualquier applet de producto** (`boe`, `cita`, `plazos`, `graph`, `placsp`, `bdns`, `borme`, `mcp`, `escrito`, `expediente`…) y el symlink `boe -> kitlegal`.
- **`pkg/legalkit`** (H24).
- **Protección de rama, revisores obligatorios, plantillas de issue y cualquier otro ajuste de la plataforma de alojamiento** no enumerado en el alcance del hito. En particular, H0 deja los umbrales de cobertura declarados y verificables (FR-029), pero no configura que su estado impida la fusión.
- **Fichero de plantilla de propuesta de cambio** (`.github/PULL_REQUEST_TEMPLATE.md`): no figura en el alcance de H0 y, por el criterio «lo no especificado no se implementa», no se crea. La estructura de PR del ritual por hito (`docs/ROADMAP.md` §6.3: objetivo, alcance, controles añadidos, decisiones, pendientes) y la exigencia de justificar toda dependencia nueva se documentan en `CONTRIBUTING.md` (FR-021, FR-032); el hito que decida automatizarla añadirá el fichero.
- **Gestores de versiones de herramientas ajenos al ecosistema Go** (tipo `mise` o `asdf`) y cualquier registro de versiones paralelo al fichero de sumas del módulo: no se introducen, conforme a FR-042 (b) y (d).
- **Reglas de un modelo open-core más allá de elegir la licencia del núcleo**: no se define qué queda fuera del núcleo ni bajo qué licencia.

## Clarifications

### Session 2026-09-09

- Q: ¿Cómo deben comportarse en H0 las órdenes del `Makefile` cuyo objeto todavía no existe (`test-integration`, `test-e2e`, `schema-check`, `skills-sync`, `release`) y el flujo `nightly`? → A: Las cinco órdenes se declaran en H0 y ninguna es silenciosa. `test-integration` ejecuta ya el comando real (`go test -race -tags=integration ./...`), que en H0 pasa sobre un conjunto vacío de tests. `test-e2e`, `schema-check` y `skills-sync` imprimen una línea que nombra el objeto ausente y el hito que lo aporta y terminan con éxito; `ci` invoca `schema-check` y sigue en verde. `release` es la excepción: al ser una acción con efectos externos, termina con código ≠ 0 y el mensaje `release: sin configurar hasta H6 (.goreleaser.yaml)`, y no forma parte de `ci` ni de `nightly`. El flujo `nightly` ejecuta `make ci` sobre `main` a diario, sin ninguna verificación contra red. Cada marcador se documenta en `CONTRIBUTING.md`. (auto: criterio c; fuente: `docs/ROADMAP.md` §1.1 y §3 y alcance de H0; spec FR-006, FR-007, FR-011, FR-027 y «Fuera de alcance»; constitución «Criterio de decisión autónoma» §1 y §2)
- Q: ¿Qué mecanismo de lint activa H0 para que un `fmt.Println` en un fichero bajo `internal/` falle, dado que `docs/ROADMAP.md` §3 sitúa `depguard` y `forbidigo` en H1? → A: H0 activa `forbidigo` en `.golangci.yml` con una única regla —prohibir `fmt.Print*` (patrón `^fmt\.Print(|f|ln)$`)— acotada por ruta a `internal/**`, de modo que `cmd/` queda autorizado en H0. `depguard` con las reglas de dependencia de §2, el test `internal/arch_test.go` y la ampliación de `forbidigo` que restringe stdout a `internal/render` entran íntegros en H1, cuando existen los paquetes a los que se refieren. SC-005 se comprueba con la propuesta de cambio desechable que prescribe el criterio de aceptación. (auto: criterio a; fuente: `docs/ROADMAP.md` §1.3, §3 y «Aceptación» de H0; spec FR-014, «Fuera de alcance» y *Assumptions*; constitución §IV)
- Q: ¿Cómo se aplican en H0 los umbrales de cobertura de la *Definition of Done* (`internal/core/**` ≥ 85 %, global ≥ 70 %) sobre un repositorio que aún no tiene prácticamente código? → A: `codecov.yml` declara los dos umbrales tal cual desde H0 y como estado que falla, no informativo. H0 los cumple con el código que introduce: `main()` delega en una función `run(args, stdout, stderr) int` cubierta por `main_test.go` en `package main`, con lo que la cobertura del módulo supera el 70 %. El componente `internal/core/**` no tiene ficheros en H0 y Codecov no lo evalúa hasta que existan, sin necesidad de configuración condicional. La orden `test` genera el perfil que el flujo `ci` publica. Que el estado bloquee la fusión depende de la protección de rama, que sigue fuera de alcance. (auto: criterio a; fuente: `docs/ROADMAP.md` §1, §1.2, §1.9, §3 y §5; constitución §III y «Criterio de decisión autónoma» §1; spec FR-029 y «Fuera de alcance»)
- Q: ¿Debe H0 crear algún paquete bajo `internal/` en el repositorio, o el directorio `internal/` no existe hasta H1? → A: H0 no crea `internal/`. Todo el código del hito vive en `cmd/kitlegal/` (`main.go`, con las variables de versión, commit y fecha inyectadas en construcción, y `main_test.go`). La regla de `forbidigo` se configura por ruta (`internal/**`) y se demuestra con la propuesta de cambio de prueba desechable. Dónde vivirán los datos de versión cuando exista el kernel CLI (H1) o el release (H6) lo decide cada uno de esos hitos. (auto: criterio a; fuente: `docs/ROADMAP.md` «Alcance» y «Aceptación» de H0, §2 y H1; spec FR-002, *Assumptions* y «Fuera de alcance»; constitución «Criterio de decisión autónoma» §2)
- Q: ¿De dónde obtiene el repositorio las herramientas externas de los controles (`golangci-lint`, `govulncheck`, `gofumpt`, `gitleaks`, `lefthook`) y qué significa exactamente la orden `install` del `Makefile`? → A: Se fijan con directivas `tool` de Go y se ejecutan con `go tool <herramienta>` desde el `Makefile`, en un fichero de módulo dedicado a herramientas para que el `go.mod` del producto no arrastre sus dependencias. Así la versión es idéntica en local y en CI, queda verificada por su `go.sum` y Dependabot la actualiza semanalmente; los flujos de CI ejecutan las mismas órdenes del `Makefile` en lugar de acciones de terceros que descargan binarios. FR-010 queda para los prerrequisitos que no pasan por `go tool`: `go` y `git`. `install` es `go install ./cmd/kitlegal` con los mismos ldflags que `build`; el enlace de `skills/*` lo añade H5. (auto: criterio c; fuente: `docs/ROADMAP.md` H5 para la semántica de `install`; spec SC-004, SC-010, FR-010 y FR-019; constitución §V, «Restricciones técnicas» y «Criterio de decisión autónoma» §1)

### Correcciones tras el gate del spec (2026-09-09)

Decisiones tomadas al corregir el spec sobre los motivos del gate, registradas aquí conforme a «Criterio de decisión autónoma» §3. Ninguna afecta a alcance, frontera humana, privacidad, TOS ni a una decisión cerrada.

- **El formato es un gate, no una corrección previa al gate** (motivo [d]). La redacción anterior hacía que `ci` ejecutase la orden `fmt` mutante antes de `lint`, con lo que un fichero sin formatear se corregía y el veredicto salía verde: el control de §3 «Formato» dejaba de ser demostrable. Se separan los dos modos del mismo control (FR-012) y `ci` ejecuta solo el de verificación, sin modificar el árbol (FR-007, US1 escenarios 6 y 7, SC-002). Alternativa rechazada: quitar el formato de `ci` y dejarlo solo en `lint`; se rechaza porque la *Definition of Done* §1.1 enumera `fmt` dentro de `make ci` y porque el pre-commit seguiría necesitando el modo mutante. (auto: criterio a; fuente: `docs/ROADMAP.md` §1.1 y §3; constitución «Gates» capa 1)
- **`ci` incluye secretos, verificación de módulos y dependencias saneadas** (motivo [d]). Sin ello, tres controles de §3 marcados H0 no pertenecían a ninguna orden del `Makefile` y SC-004 («el veredicto local predice el de la PR») era falsa. Se añaden a FR-007 y se reformulan FR-018, FR-019 y FR-020 en términos de esa orden. Alternativa rechazada: reducir SC-004 a los controles que `ci` ya ejecutaba y declarar en FR-025 los que corre solo la integración continua; se rechaza porque FR-042 ya exige que la integración continua pase por las órdenes del `Makefile`, y porque un control que solo existe en la PR rompe el ciclo local de quien contribuye. (auto: criterio a; fuente: `docs/ROADMAP.md` §3 y §6.3; constitución «Gates» capa 1)
- **El recuento de ficheros del alcance se sustituye por la enumeración** (motivo [d]). SC-008 decía «los doce ficheros»; el alcance de H0 enumera dieciséis entradas y dieciocho ficheros. Se transcribe la lista y se añaden los artefactos que exigen FR-005, FR-040 y FR-042, sin cifra que pueda desajustarse. (auto: criterio a; fuente: `docs/ROADMAP.md` «Alcance» de H0)
- **FR-040 y FR-042 pasan de prescribir diseño a exigir resultado** (motivo [g]). El spec fija ahora *qué* debe cumplirse —comportamiento de `version` cubierto por tests y umbral global superado; herramientas con versión fijada, idéntica en local y en CI, verificable, actualizada por Dependabot, sin más prerrequisitos que Go y `git` y sin contaminar las dependencias del producto—; el *cómo* (`run(args, stdout, stderr) int`, directivas `tool` y fichero de módulo dedicado, `go tool`) permanece en las respuestas Q3 y Q5 de esta sección como rastro de la decisión automática y se materializa en el plan. Se añade a FR-028 que Dependabot cubre también esas herramientas, que hasta ahora solo constaba en la respuesta a Q5. (auto: criterio a; fuente: constitución «Criterio de decisión autónoma» §3)
- **Caché del flujo y medida de los 3 minutos** (observación no bloqueante del gate). El criterio literal del hito se mantiene sin condición en SC-006; FR-025 obliga a conservar la caché de dependencias y herramientas entre ejecuciones, y *Assumptions* traslada al plan la obligación de medir también la ejecución en frío y de corregir el diseño del flujo —no el criterio— si no bastara. (auto: criterio a; fuente: `docs/ROADMAP.md` «Aceptación» de H0)
- **Plantilla de propuesta de cambio** (observación no bloqueante del gate). FR-021 y US2 escenario 5 hablaban de «plantilla o guía»; se unifican en `CONTRIBUTING.md` y el fichero de plantilla se declara fuera de alcance por no figurar en el alcance del hito, con la estructura de PR del ritual documentada en la guía. (auto: criterio a; fuente: `docs/ROADMAP.md` «Alcance» de H0 y §6.3; constitución «Criterio de decisión autónoma» §2)
