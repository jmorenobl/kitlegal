# Tasks: H0 · Esqueleto del repo y gates de CI

**Input**: `specs/001-h0-esqueleto-del-repo/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: plan.md y spec.md leídos; `.specify/memory/constitution.md` aplicada.

**Modo**: desatendido. Las tareas se ejecutan y verifican **una a una** por el workflow `hito`, con guardián
de diff por tarea. Por eso cada tarea es una **rebanada vertical** que incluye su comprobación y deja la
batería determinista en verde al terminar, y **declara en su propia línea todas las rutas** que va a crear o
modificar.

## Formato: `[ID] [P?] [Story] Descripción con rutas`

- **[P]**: la tarea no depende de la anterior más allá del orden (ficheros disjuntos).
- **[Story]**: historia de usuario del spec (US1 veredicto local · US2 la PR bloquea · US3 `version` ·
  US4 documentación).
- Sin tareas **[datos]**: H0 no crea ni modifica fixtures ni esquemas (plan.md, «Fixtures»; «Fuera de
  alcance» del spec). Ninguna tarea toca esos directorios ni graba nada contra la red.

## Batería de verificación por tarea

- Desde T001 en adelante existe `Makefile` con objetivo `ci`, así que la verificación es **`make ci`**:
  comprobación de formato · lint · tests con detector de carreras · vulnerabilidades · comprobación de
  esquemas · secretos · integridad de módulos · dependencias saneadas.
- `make ci` no modifica ningún fichero versionado (M1 del contrato de órdenes), así que la verificación
  nunca ensucia el diff que juzga el guardián.

---

## Phase 1: Fundación — el veredicto existe y está en verde

**Objetivo**: que a partir de la primera tarea toda línea que entre pase por `make ci`.

- [X] T001 [US1] Módulo, binario y veredicto en verde: `go.mod` con la ruta de módulo que fija FR-001 —el repositorio jmorenobl kitlegal en GitHub— y las directivas `go 1.26.0` y `toolchain go1.26.6` y sin dependencias de producto; `cmd/kitlegal/main.go` con `main` reducido a `os.Exit` sobre `run args stdout stderr` que devuelve int, las variables version, commit y fecha inyectables por ldflags con los valores por defecto dev, none y unknown, tres líneas de salida y código 0 para el verbo version y línea de uso en stderr con código 2 para verbo ausente, desconocido o argumento sobrante, sin panic y sin red, más su tabla de casos con `t.Parallel` y escritores en memoria en el fichero de test del mismo paquete; los módulos de herramienta `tools/golangci-lint/`, `tools/govulncheck/` y `tools/gitleaks/`, cada uno creado con `go get -tool` seguido de `go mod tidy` sobre su propio modfile para que su fichero de sumas quede completo; `.golangci.yml` en formato v2 con los 22 linters de FR-013, los formateadores gofumpt y goimports, misspell con diccionario inglés y la regla forbidigo que prohíbe `fmt.Print` y sus variantes acotada por ruta al árbol interno que H0 no crea; `.gitleaksignore` en la raíz con las dos huellas justificadas de la documentación vendorizada del paquete de skills de agente en formato fichero:regla:línea, una línea de comentario por huella y sin desactivar ninguna regla ni excluir ninguna otra ruta; y `Makefile` con las órdenes build test test-integration test-e2e lint fmt vuln schema-check skills-sync release ci install fmt-check lint-fast secrets mod-verify mod-tidy-check check-tools hooks help, `export GOTOOLCHAIN` derivado por awk de la directiva toolchain con `$(error …)` si falta, LDFLAGS idénticos en build e install, CGO_ENABLED=0 solo en build e install, marcadores verdes que nombran el objeto ausente y el hito que lo aporta, `release` fallando con código distinto de 0 hasta H6, `mod-verify` ejecutando la verificación de integridad en el módulo raíz y en cada modfile de herramienta **que exista en ese momento**, resuelto con un glob sobre el directorio de herramientas y nunca con una lista fija de cuatro rutas —así la orden queda en verde con los tres módulos que crea esta tarea y cubre el cuarto por sí sola en cuanto T003 añada `tools/lefthook/`—, y `ci` encadenando fmt-check lint test vuln schema-check secrets mod-verify mod-tidy-check sin mutar el árbol de trabajo.

**Checkpoint**: `make build test lint vuln` y `make ci` en verde sobre un clon limpio (SC-001, SC-002,
SC-003, SC-011, SC-012).

---

## Phase 2: Controles que completan el veredicto

- [X] T002 [P] [US1] `codecov.yml` con el umbral global ≥ 70 % y un componente para el dominio interno ≥ 85 %, ambos declarados como estado que falla y no informativo, sin exclusiones ni excepciones y sin configuración condicional que haya que editar cuando ese componente empiece a tener ficheros; comprobado ejecutando `make test` y leyendo el total del perfil de cobertura con `go tool cover -func`, que debe superar el 70 % con los tests del propio hito.
- [X] T003 [US1] Ganchos de pre-commit: `lefthook.yml` con un bloque pre-commit que invoca únicamente órdenes del `Makefile` —`make fmt` con `stage_fixed` y glob de ficheros Go, `make lint-fast`, `make secrets` y `make mod-tidy-check`— y el módulo de herramienta `tools/lefthook/` creado igual que los demás; comprobado con `make hooks` y el commit de prueba del escenario 8 de la guía de validación del hito —una mutación de espaciado sobre una línea que ya existe, más una línea de comentario nueva para que el índice no quede vacío tras formatear, nunca una declaración añadida al final del fichero—, tras el cual el fichero debe quedar formateado, re-preparado y confirmado; si `stage_fixed` no re-preparase los ficheros corregidos en esa versión de la herramienta, sustituirlo por un `git add` explícito de los ficheros preparados. El commit de prueba **se deshace siempre**, de modo que al terminar la tarea `cmd/kitlegal/main.go` y el historial quedan exactamente como al empezar —esta tarea no declara ese fichero, así que un commit de prueba sin deshacer haría que el guardián de diff la rechazara—, y el deshacer va **anclado a la revisión de partida y acotado al fichero mutado**: `antes=$(git rev-parse HEAD)` antes del commit de prueba y, después, `git reset --soft "$antes"` seguido de `git restore --source="$antes" --staged --worktree -- cmd/kitlegal/main.go`. Ni un `HEAD~1` relativo ni un `git reset --hard`. El `HEAD~1` a ciegas borraría el commit de la tarea anterior si el gancho hubiera impedido el de prueba, mientras que el anclado es idempotente tanto si el commit se creó como si no; y el `--hard` descartaría **toda** modificación sin confirmar de los ficheros versionados, entre ellas las que el propio workflow escribe en `gates/tarea-actual.json` y `gates/tareas-intentos.json` antes de cada tarea y confirma después, con lo que el guardián de diff leería el id y las rutas de T002 y rechazaría `lefthook.yml` y `tools/lefthook/`. Comprobación final: `git rev-parse HEAD` igual a `$antes` y `git diff --stat "$antes" -- cmd/kitlegal/main.go` vacío —no «el árbol limpio», que bajo el workflow nunca lo está mientras el directorio del feature tenga cambios sin confirmar, y perseguirlo con un `git checkout .` o un `git clean` destruiría el trabajo de la tarea—.

---

## Phase 3: La PR es la que bloquea (US2)

**Objetivo**: que los mismos controles se ejecuten en la plataforma, invocados por las mismas órdenes.

- [X] T004 [US2] `.github/workflows/ci.yml`: un solo job disparado por pull_request y por push a la rama principal cuyos pasos propios se limitan a preparar el entorno —obtener el código, instalar Go con go-version-file apuntando al go.mod y sin declarar go-version ni check-latest, y conservar entre ejecuciones la caché de dependencias, de construcción y de herramientas— más ejecutar `make ci` y publicar el perfil de cobertura con fail_ci_if_error activado; permisos mínimos de solo lectura del contenido; ningún control aplicado por una vía distinta de las órdenes del `Makefile` y ninguna variable CGO_ENABLED ni GOTOOLCHAIN exportada en el entorno del job.
- [X] T005 [P] [US2] `.github/workflows/nightly.yml`: flujo programado a diario sobre la rama principal que prepara el entorno igual que el flujo de PR y ejecuta `make ci`, sin ninguna verificación contra fuentes reales, sin invocar `release` y sin ningún control que no esté ya especificado en el hito.
- [X] T006 [P] [US2] `.github/workflows/codeql.yml`: análisis estático de seguridad semanal con la acción oficial de CodeQL sobre el lenguaje go y la suite security-extended, con el permiso de escritura de eventos de seguridad acotado a ese job.
- [X] T007 [P] [US2] `.github/dependabot.yml` con revisión semanal y seis entradas: una gomod para el módulo raíz, una gomod por cada uno de los cuatro módulos de herramienta —de modo que las versiones fijadas de los controles se actualicen solas— y una github-actions para las acciones que usan los flujos.

---

## Phase 4: Documentación fundacional (US4)

- [X] T008 [P] [US4] `LICENSE` con el texto canónico y sin editar de Apache-2.0, licencia del núcleo conforme a la decisión open-core.
- [ ] T009 [P] [US4] `README.md`: qué es kitlegal y qué entrega este hito, los dos únicos prerrequisitos —una cadena Go 1.21 o superior y git, sin instalar ninguna herramienta de control ni ningún parche concreto de Go—, cómo construir e instalar el binario, cómo invocar el verbo version y qué imprime, y cómo ejecutar los controles con las órdenes del `Makefile`, incluida la instalación de los ganchos.
- [ ] T010 [US4] `CONTRIBUTING.md`: el ritual por hito y la estructura de propuesta de cambio —objetivo, alcance, controles añadidos, decisiones, pendientes—, Conventional Commits, versionado semántico, el catálogo de controles con la orden que ejecuta cada uno y la advertencia de que el gancho no es la autoridad final, la exigencia explícita de justificar toda dependencia nueva, las cinco órdenes que existen pero reciben su contenido en un hito posterior, el comportamiento de `make vuln` sin red y por qué no se captura su error, el procedimiento trazable para excluir un falso positivo de la detección de secretos por huella y con comentario justificativo sin desactivar el control, y el procedimiento manual de subida de la directiva toolchain de go.mod cuyo disparador es un hallazgo del análisis de vulnerabilidades.
- [ ] T011 [P] [US4] Los cuatro ADR fundacionales en formato MADR corto —contexto y problema, opciones consideradas, decisión y consecuencias, ninguno vacío ni marcador de posición—: `docs/ADR/0001-multicall.md`, `docs/ADR/0002-sqlite-sin-cgo.md`, `docs/ADR/0003-no-cendoj-masivo.md` y `docs/ADR/0004-frontera-humana.md`, que registran decisiones ya cerradas sin reabrirlas ni matizarlas.
- [ ] T012 [US4] `CHANGELOG.md` con la sección Unreleased y la entrada de este hito: el binario con su verbo version y el conjunto de controles que quedan activos, siguiendo la convención de changelog que el propio `CONTRIBUTING.md` declara.

---

## Phase 5: Validación de los criterios de éxito

- [ ] T013 [US1] Ejecutar en local los escenarios 1 a 10 y 12 de `specs/001-h0-esqueleto-del-repo/quickstart.md` —veredicto reproducible, árbol intacto antes y después de `make ci`, el formato como gate que falla sin corregir, salida y códigos de salida del verbo version contra la revisión construida, cobertura por encima del umbral, órdenes con el objeto ausente, prerrequisitos ausentes y pin de toolchain indiferente al entorno, ganchos de pre-commit, detección de secretos y dependencias saneadas— y registrar el resultado de cada escenario, con el comando y su salida relevante, en `specs/001-h0-esqueleto-del-repo/gates/validacion-local.md`, dejando al terminar cada fichero que un escenario haya mutado exactamente como estaba —comprobado ruta a ruta (`git diff --stat -- <ruta>` vacío), no exigiendo un árbol globalmente limpio, que bajo el workflow nunca lo está porque el propio registro y el estado de la tarea son cambios sin confirmar—, de modo que el único cambio que esta tarea aporta al árbol es ese registro. **Ningún secreto detectable entra en ese registro**: el token que el escenario 9 genera al vuelo no se transcribe —ni en el comando, ni en la salida, ni en un extracto del fichero mutado—, se anota redactado o se anota solo el veredicto de `make secrets` y de `make ci`; como `make secrets` es `gitleaks dir .` sobre todo el árbol, `specs/` incluido, y H0 no crea fichero de exclusiones, un token literal en el registro dejaría en rojo el propio `make ci` con el que se verifica esta tarea y todos los posteriores, y por un fichero de gates, no por un control. La misma regla vale para cualquier otro valor que las reglas de detección puedan reconocer.
- [ ] T014 [US2] Ejecutar el escenario 11 de `specs/001-h0-esqueleto-del-repo/quickstart.md` con dos ramas desechables **creadas cada una desde la rama del hito, nunca una de la otra** —volver a la rama del hito antes de crear la segunda, y comprobar con el estado del árbol que no hereda el fichero de la primera— y **abrir la propuesta de cambio de cada una contra la rama del hito**, porque el flujo de integración continua solo se dispara por propuesta de cambio y por push a la rama principal, de modo que un push a una rama lateral no ejecutaría ningún control en la plataforma: una rama añade una escritura directa a stdout bajo el árbol interno prohibido y su propuesta debe quedar bloqueada por lint, y la otra está limpia y su propuesta debe pasar todos los controles en menos de 3 minutos; medir la duración en caliente y en frío, comprobar que el secreto de subida de cobertura está dado de alta y que el componente del dominio interno no aparece como fallo ni como pendiente bloqueante, y verificar si la actualización automática de dependencias propone también subir la directiva toolchain. **Cada commit de prueba prepara únicamente su propio fichero** —`git add internal/prueba/p.go` en la rama sucia, `git add CHANGELOG.md` en la limpia—, nunca `git add -A`: bajo el workflow el directorio del feature tiene cambios versionados sin confirmar (`gates/tarea-actual.json`, `gates/tareas-intentos.json` y cualquier avance del propio registro), y un `git add -A` los confirmaría en la rama desechable, con lo que al volver a la rama del hito revertirían a HEAD, el `git branch -D` se llevaría el único ejemplar del registro y el commit de esta tarea saldría etiquetado con el id de T013; como alternativa equivalente, crear cada rama en un `git worktree add` temporal para que el árbol de la rama del hito no cambie en ningún momento. El commit de la rama sucia lleva además `--no-verify`, porque el gancho de pre-commit instalado en T003 ejecuta el mismo lint que ese fichero infringe a propósito y sin ello no habría propuesta de cambio que evaluar, que es la autoridad que este escenario mide; el de la rama limpia no lo lleva. Registrar los dos números y los cuatro resultados en `specs/001-h0-esqueleto-del-repo/gates/verificacion-pr.md` **en la rama del hito y solo después** de cerrar las dos propuestas sin integrarlas y borrar ambas ramas en local y en el remoto. Si el repositorio todavía no tiene remoto en la plataforma o falta el secreto de cobertura —prerrequisitos humanos, no artefactos del repositorio—, dejar escrito en ese mismo fichero qué queda pendiente y por qué, sin simular ningún resultado y sin rebajar ninguna bandera para esquivarlo.

---

## Dependencias y orden de ejecución

El orden es **estrictamente secuencial y ejecutable**: ninguna tarea depende de una posterior.

- **T001** no depende de nada y es la única que puede fallar por ausencia de infraestructura: entrega a la
  vez el módulo, el binario testeado y el `Makefile`, porque `make ci` no puede estar en verde sin lint,
  vulnerabilidades y secretos, y esos tres necesitan sus módulos de herramienta. Es el «esqueleto mínimo».
- **T002** y **T003** completan controles que no forman parte del veredicto agregado (cobertura publicada,
  pre-commit) y por eso pueden ir después sin romper nada.
- **T004** depende de que `make ci` exista (T001). **T005**, **T006** y **T007** son ficheros disjuntos
  entre sí y con T004.
- **T008** a **T012** son documentación: no alteran ningún control, así que `make ci` sigue verde. T012
  (changelog) va después de T010 porque la convención de changelog se declara allí.
- **T013** valida en local todo lo anterior; **T014** valida lo que solo se puede comprobar en la
  plataforma y cierra el criterio de aceptación literal del hito.

### Oportunidades de paralelismo

Marcadas con **[P]**: T002, T005, T006, T007, T008, T009, T011. En el bucle del workflow se ejecutan igual
en secuencia (una tarea, un guardián, una verificación); la marca indica que no comparten fichero ni
dependencia y podrían repartirse si alguien las hiciera a mano.

---

## Trazabilidad

### Historias de usuario → tareas

| Historia | Prioridad | Tareas | Prueba independiente |
|---|---|---|---|
| US1 · Veredicto de calidad reproducible en local | P1 | T001, T002, T003, T013 | Clon limpio con Go y git: `make build test lint vuln` y `make ci` terminan con éxito y el árbol queda intacto |
| US2 · La PR es la que bloquea, y bloquea rápido | P2 | T004, T005, T006, T007, T014 | Dos propuestas de cambio desechables, una sucia y una limpia: veredicto y duración |
| US3 · Saber qué binario se está ejecutando | P3 | T001 (binario y tests), T013 (escenario 4) | Construir el binario, invocar el verbo version y comparar el commit con el de la revisión |
| US4 · Entender el proyecto y sus reglas sin preguntar | P4 | T008, T009, T010, T011, T012 | Leer el repositorio: licencia, guía, changelog y los cuatro ADR con contexto y consecuencias |

### Requisitos → tareas

| Requisitos | Tarea |
|---|---|
| FR-001 a FR-011, FR-012 a FR-014, FR-017, FR-037 a FR-042 | T001 |
| FR-029 | T002 |
| FR-022 | T003 |
| FR-025 | T004 |
| FR-027 | T005 |
| FR-026 | T006 |
| FR-028 | T007 |
| FR-030 | T008 |
| FR-031 | T009 |
| FR-021, FR-032 | T010 |
| FR-024, FR-033 a FR-036 | T011 |
| FR-023 | T012 |
| SC-001 a SC-003, SC-007, SC-010 a SC-012 | T013 |
| SC-004, SC-008 | T013 (vía los escenarios 3, 9, 10 y 12 de `quickstart.md`) |
| SC-005, SC-006, SC-009 | T014 |

FR-015, FR-016, FR-018, FR-019 y FR-020 quedan cubiertos por T001 (las órdenes que los ejecutan dentro de
`ci`) y ejercidos por T004 (la integración continua invoca esa misma orden).

SC-004 (equivalencia local ↔ CI) y SC-008 (inventario completo de artefactos) no tienen tarea propia
porque no son controles nuevos sino comprobaciones sobre lo ya construido: los ejecuta T013 dentro de los
escenarios que ya recorre —el 3 y el 9 y el 10 muestran que un control que falla en la propuesta de cambio
falla igual en local, porque es literalmente la misma orden; el 12 recorre el inventario de artefactos—.
La fila anterior lo deja explícito para que esta tabla se pueda auditar sin abrir `quickstart.md`.

### Definition of Done (`docs/ROADMAP.md` §1) → tareas

| # | Punto | Cómo se cumple |
|---|---|---|
| 1 | `make ci` en verde | T001 lo crea y lo deja verde; toda tarea posterior lo mantiene |
| 2 | Tests unitarios offline; fixtures si toca red | T001 (tabla de casos con escritores en memoria). **Sin fixtures**: H0 no toca ninguna red ni crea directorio de fixtures |
| 3 | Sin escrituras a stdout ni `os.Exit` fuera de los paquetes autorizados | T001 (regla forbidigo acotada por ruta; `os.Exit` una sola vez, en el punto de entrada) |
| 4 | Salidas validadas contra esquemas | **No aplica**: el verbo version emite texto plano, no el sobre; la validación entra con los esquemas en H4/H11 |
| 5 | Errores tipados → exit code estable; sin panic | T001 (código 0 y código 2, ambos de la tabla estable; `run` devuelve int y nunca entra en pánico) |
| 6 | Comportamiento visible → e2e y `CHANGELOG.md` | T012. **Sin e2e**: el runner de tests de guion es alcance de H1, y el único comportamiento visible ya está cubierto por los tests del punto de entrada |
| 7 | Decisión de arquitectura → ADR | T011 |
| 8 | Fuente externa → fila en el inventario de fuentes y caso de verificación | **No aplica**: H0 no toca ninguna fuente; ese inventario y su script nacen en H4 |
| 9 | Cobertura global ≥ 70 % y del dominio ≥ 85 % | T002 (umbrales) + T001 (los tests que superan el global) + T013 (comprobación) |

---

## Estrategia de implementación

1. **T001 es el MVP y el gate**: al terminarla, el repositorio ya está blindado y cualquier trabajo
   posterior atraviesa los controles. Si algo va a costar, es aquí.
2. **T002 a T007** añaden los controles que no operan sobre el árbol de trabajo o que no forman parte del
   veredicto agregado.
3. **T008 a T012** hacen el repositorio legible para quien llega por primera vez.
4. **T013 y T014** convierten los criterios de éxito en evidencia registrada.

---

## Notas

- **Todas las rutas van en la línea de la tarea.** El guardián de diff rechaza cualquier fichero fuera de
  ellas. Siempre permitidos sin declarar: `go.mod`, `go.sum`, `CHANGELOG.md` y el propio directorio del
  feature; declarar un fichero `.go` permite además su fichero de test; declarar un directorio permite todo
  lo que cuelga de él.
- **Sin fixtures ni esquemas en H0**, así que ninguna tarea lleva la etiqueta `[datos]` y ninguna pausa
  para revisión humana por ese motivo. `KITLEGAL_RECORD` no se usa en ningún punto del hito: no existe
  todavía nada que grabar.
- **Vía de exclusión de falsos positivos de la detección de secretos**: se documenta en `CONTRIBUTING.md`
  (T010) como procedimiento por huella, con comentario justificativo y sin desactivar el control. **T001
  crea `.gitleaksignore` con dos huellas reales**, no un fichero vacío: el árbol heredado trae dos líneas
  de ejemplo marcadas `// DON'T` en la documentación vendorizada de `samber/cc-skills-golang` que la regla
  `generic-api-key` detecta y que no son credenciales. Como nace con contenido justificado, SC-008 se
  cumple. Antes de esta decisión el árbol traía además una copia duplicada del mismo paquete de skills en
  `agent/`, sin referencias y eliminada en el commit anterior al hito; por eso las huellas son dos y no
  cuatro. Decisión humana del 2026-09-10 en
  [gates/decision-humana-secretos.md](./gates/decision-humana-secretos.md).
- **`.gitignore` ya existe** y cubre los artefactos que generan las órdenes (`bin`, perfiles de cobertura,
  el directorio de asunto local). Ninguna tarea lo modifica.
- **Nada de atajos** al reparar una verificación en rojo: ni desactivar linters, ni excepciones sin
  justificar en el código, ni tests saltados, ni errores silenciados. Si un test sigue legítimamente en
  rojo porque su implementación pertenece a una tarea posterior, la tarea está mal delimitada y hay que
  anotarlo, no arreglarlo por la vía rápida.
