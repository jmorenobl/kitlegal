---

description: "Tareas de H41.4 · Las skills del producto, con las descripciones cortas que quitan la caída de la eval 04"
---

# Tasks: H41.4 · Las skills del producto, con las descripciones cortas que quitan la caída de la eval 04

**Input**: `specs/038-h41-4-las-skills-del/` — [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md),
[data-model.md](data-model.md), [quickstart.md](quickstart.md) y [contracts/](contracts/) (`skills.md`,
`controles.md`, `aceptacion.md`).

**Tests**: obligatorios (constitución III). La primera tarea escribe la suite de aceptación, que el workflow congela;
cada tarea posterior trae su test y el cambio que lo pone en verde, y deja `make ci` en verde por sí sola.

**Organización**: una fase por historia, en el orden en que se ejecutan una tras otra (plan.md «Orden de
implementación»). La línea de cada tarea declara todas las rutas que crea o modifica (el guardián rechaza cualquier
otra; `go.mod`, `go.sum`, `CHANGELOG.md` y el directorio del feature siempre se admiten, y declarar `x.go` admite
`x_test.go`).

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: puede ir en paralelo (ficheros distintos, sin dependencia de una tarea sin terminar). Ninguna lo lleva: T003 y
  T004 escriben el mismo fichero, y T002 deja las descripciones de las que parten las demás.
- **[Story]**: la historia de [spec.md](spec.md) a la que sirve (US1 a US3; US4 no tiene tareas: la hacen el cierre del
  workflow y el sondeo fuera del run, FR-010 y FR-011).
- **[aceptacion]**: la suite congelada (solo T001). **[datos]**: ninguna tarea toca ficheros bajo `testdata/` ni
  `schemas/` (plan.md «Gates», «Fixtures»: ninguno; «Datos externos»: ninguno).

## Convenciones de todas las tareas

- `make ci` en verde al terminar cada una, en primer plano y por objetivos si hace falta (`make fmt-check`, `make lint`,
  `make test`, `make test-integration`, `make vuln`, `make schema-check`, `make skills-check`…). Ninguna usa la red más
  que la de las herramientas de Go; ninguna graba (no hay datos externos: plan.md «Datos externos»; ni manifiesto
  `grabaciones.json` ni `TestGrabar*`); ninguna escribe en `data/` ni fuera del repositorio y de `$TMPDIR`; ninguna abre
  una sesión con modelo ni ejecuta `make evals`, `make evals-sondeo`, `make evals-laboratorio`, `TestSondeo*` ni los
  guiones `scripts/evals*.sh`.
- No se tocan `laboratorio/`, `evals/`, `data/`, `web/`, `mcp/`, `internal/empaquetado/`, `.goreleaser.yaml`,
  `.github/`, `Makefile`, `.golangci.yml`, `scripts/`, `cmd/`, el arnés (los `internal/evals/*.go` que no son
  `_test.go`), las referencias de las cuatro skills, ningún fichero de código de producción ni ningún test congelado
  salvo `TestSkillsDelProductoDesdeControl` (FR-003, FR-005 a FR-007; plan.md «Project Structure»). Las suites
  congeladas de H39 a H41.3 (`specs/031-h39-c-omo-se/aceptacion/`, `specs/032-h40-por-qu-e/aceptacion/`,
  `specs/034-h41-las-skills-del/aceptacion/`, `specs/035-h41-1-plazos-y-territorio/aceptacion/`,
  `specs/036-h41-2-por-qu-e/aceptacion/`, `specs/037-h41-3-por-que-normativa/aceptacion/` y sus copias activadas en la
  carpeta de guiones de `internal/app`) siguen en verde sin cambiar un fichero suyo (FR-009).
- Los textos son los de los contratos: los cuatro bloques y la viñeta del changelog, los de
  [contracts/skills.md](contracts/skills.md) §2 y §4; las comprobaciones, sus casos y sus mensajes, los de
  [contracts/controles.md](contracts/controles.md) §1 a §3; el guion, el de [contracts/aceptacion.md](contracts/aceptacion.md)
  §2. Nada se inventa fuera de ellos.
- `internal/evals` no nombra `os.Stdin`, `os.Stdout`, `os.Stderr` ni llama a `os.Exit` (R4, R5); ningún `panic`; lo que
  llega manipulado (un `SKILL.md` sin frontmatter, una variante ilegible) es el error de la comprobación con su ruta,
  defecto `inesperado` (ADR 0023), sin casos propios. Ningún `//nolint` y `.golangci.yml` no cambia: se lee con
  `fs.ReadDir` y `fs.ReadFile` sobre `os.DirFS`, se escribe con `leerFichero` y `escribirFichero`
  (`internal/evals/preparar.go`) y cada mutante abre con `t.Helper()` (research V6, D10). Tests de tabla con
  `t.Parallel()` y `testify` (`require`/`assert`), como los de `internal/evals`.
- Cada control se ve fallar con un mutante (la medida pasada del umbral) antes de darlo por bueno (ADR 0029): los de las
  comprobaciones nuevas son subtests permanentes sobre copias en `t.TempDir()`; los que se ven a mano se aplican sobre
  una copia bajo `$TMPDIR`, nunca sobre el árbol versionado, y no se versionan.

---

## Phase 1: Setup — la suite de aceptación

**Propósito**: fijar, antes de cualquier cambio, lo que `kitlegal skills install` tiene que instalar de las cuatro
skills (ADR 0018; plan.md «Aceptación e2e»; contracts/aceptacion.md).

- [X] T001 [aceptacion] [US1] Escribir, a partir del spec y de [contracts/aceptacion.md](contracts/aceptacion.md) §2, sin código de producto y sin evals nuevas (las evals de las cuatro skills no cambian: plan.md «Project Structure», FR-010 las mide tal cual), el guion testscript specs/038-h41-4-las-skills-del/aceptacion/skills-instaladas.txtar, byte a byte el texto del bloque de §2 con su salto de línea final (el workflow lo activa como `h41-4-skills-instaladas.txtar` en la carpeta de guiones de internal/app al terminar el bucle y lo ejecuta `TestEntregaDelHito`; el arnés e2e no cambia). **Cabecera**: el comentario de §2, con la línea `# Cubre: US1 · FR-001, FR-002 · SC-001`. **Órdenes**: `exec kitlegal skills install --dir destino --json` y, por skill (`consulta-juridica`, `jurisprudencia`, `normativa`, `plazos-y-territorio`), sus dos `grep` sobre `destino/<skill>/SKILL.md`: el del frontmatter con `name:`, `description: >-` y el texto de la variante con cada espacio como `( |\n  )`, seguido de `metadata:` o, en `consulta-juridica`, del cierre `---`; y el de al menos dos líneas sangradas (8 `grep` en total). El guion no afirma dónde se corta cada línea ni nada del cuerpo, las referencias o los límites (research D5). Falla hoy por una aserción, `skills-instaladas.txtar:22: no match` en el primer `grep` (research V9), lo que el workflow comprueba antes de congelarlo; ninguna tarea posterior lo toca.

**Checkpoint**: la suite está escrita y en rojo por una aserción; desde aquí queda congelada.

---

## Phase 2: Foundational

Ninguna: no hay paquete, tipo ni auxiliar de producción nuevo; las comprobaciones reutilizan `descripcionesCortas`,
`descripcionYCuerpo` y `leerLosFicheros` de `internal/evals` (plan.md, principio V).

---

## Phase 3: User Story 1 — Las cuatro skills del producto llevan las descripciones medidas (P1) 🎯 MVP

**Goal**: la `description` de cada skill de `skills/` es, byte a byte, la de
`laboratorio/producto-descripciones-cortas/<skill>.txt`, en un bloque `>-` de más de una línea, con 450 por skill y 1700
entre las cuatro como límites (FR-001, FR-002, FR-004).

**Independent Test**: `go test -run '^TestDescripcionesDelProducto$' ./internal/evals/` en verde con la rama y con las
líneas exactas en cada mutante de contracts/controles.md §1.3; el guion de T001, activado en una copia bajo `$TMPDIR`, en
verde.

- [ ] T002 [US1] Las cuatro descripciones de la variante en `skills/` y el único cambio de una suite congelada, en una sola tarea porque con las descripciones y sin el cambio `TestSkillsDelProductoDesdeControl` sale en rojo (FR-001, FR-002, FR-003, FR-004, FR-005, FR-009; SC-001, SC-002, SC-003, SC-004; contracts/skills.md §1 a §3; contracts/controles.md §3; research V1, V2, V4, V7, D3, D4, D8). **Skills**: skills/consulta-juridica/SKILL.md, skills/jurisprudencia/SKILL.md, skills/normativa/SKILL.md y skills/plazos-y-territorio/SKILL.md: solo las líneas de la clave `description` (la que empieza por `description:` y las sangradas que la siguen, la regla de `lineasDelSkill`) dejan su sitio, detrás de `name:`, al bloque de su skill de contracts/skills.md §2, byte a byte: `description: >-` y cuatro líneas sangradas con dos espacios, de como mucho 120 caracteres; nada más del fichero cambia. **Test que cambia**: internal/evals/laboratorio_familias_test.go, solo lo de contracts/controles.md §3.1: `TestSkillsDelProductoDesdeControl` lee una vez `descripcionesCortas(raizDelRepositorio)` con `require.NoError` y pasa el texto de la skill a `comprobarElProductoDesdeControl` en un parámetro más; en «desde control», lo esperado y lo obtenido se comparan con `compararLosFicheros` con el `SKILL.md` de los dos sin las líneas de la clave (`descripcionYCuerpo`); en «frontmatter», el del control y el de `skills/<skill>`, los dos sin las líneas de la clave, con `assert.Equal`; en los dos, además, la `description` que lee `skills.LeerFrontmatter` de `skills/<skill>/SKILL.md` es el texto de la variante, con el mensaje `la description de skills/<skill>/SKILL.md es la de laboratorio/producto-descripciones-cortas/<skill>.txt`; los auxiliares nuevos del fichero `sinLasLineasDeLaDescripcion` y `exigirLaDescripcionDeLaVariante`; y los comentarios de `TestSkillsDelProductoDesdeControl`, `comprobarElProductoDesdeControl` y `declaracionDelProducto` dicen que el frontmatter es el de su control con la `description` de la variante (FR-005 de H41.4). No cambian (§3.2) `declaracionesDelProducto`, `sinLoGenerado`, `conLaDeclaracion`, `compararLosFicheros` ni sus mensajes, el subtest «movimientos», `huellasDeLosControles` ni `TestVariantesDeControl`. **Verificación**: `go test -count=1 ./...` en verde, con `TestDescripcionesCortas`, `TestDescripcionesCortasPorLaEstructura`, `TestConLaDescripcion`, `TestVariantesDeControl`, `TestEleccionDeLosMutantes`, `TestSkillsDelRepositorio` y `TestArnesDeLasFamilias` sin cambiar ningún valor esperado (research V4); `make skills-sync` seguido de `git status --porcelain` no da ningún fichero fuera de los cinco de esta tarea (FR-007; research V5); `ValidarFrontmatter` en `TestSkillsDelRepositorio` no da ningún defecto de 900 ni de etiqueta XML (FR-004), y su rojo con la medida pasada del umbral es el de los casos de internal/skills/frontmatter_test.go («description de 901 caracteres (máximo 900)», «description con la etiqueta XML …»), que no cambian; y, sobre una copia del repositorio bajo `$TMPDIR`, los rojos de contracts/controles.md §3.3 (research V7): `## Protocolo` → `## Protocolos` en `normativa`, un carácter en `name` de `jurisprudencia`, un carácter al final de `jurisprudencia/references/consulta_preparada.md`, `normativa/references/otra.md` de más y la `description` de `normativa` sin su punto final, cada uno con su ruta o su mensaje; con el guion de T001 copiado como `h41-4-skills-instaladas.txtar` en la carpeta de guiones de internal/app de esa copia, `go test -run '^TestEntregaDelHito$' ./internal/app/` en verde (8 de 8 `grep`).
- [ ] T003 [US1] `TestDescripcionesDelProducto`, el control de FR-001, FR-002 y FR-004 en `make ci`, con sus nueve casos (FR-001, FR-002, FR-004; SC-001, SC-003; contracts/controles.md §1; research D1, D6, D7, V6, V8). **Fichero nuevo**: internal/evals/descripciones_del_producto_test.go, en el paquete `evals`, con su comentario de cabecera que dice qué protege (H41.4, FR-001, FR-002, FR-004). **Comprobación** (§1.1): `defectosDeLasDescripcionesDelProducto(raiz string) ([]string, error)` lee los textos de la variante con `descripcionesCortas(raiz)` (sus defectos se descartan: los da `TestDescripcionesCortas`) y cada subdirectorio de `skills/` en orden de nombre (`fs.ReadDir` sobre `os.DirFS(raiz)`), su `SKILL.md` con `fs.ReadFile` y su frontmatter con `skills.LeerFrontmatter`; por skill, en este orden y como mucho una de cada clase, las líneas del texto (`skills/<skill>: no tiene descripción en laboratorio/producto-descripciones-cortas` o `skills/<skill>/SKILL.md: la description no es la de laboratorio/producto-descripciones-cortas/<skill>.txt`), de la forma (con las líneas de la clave que da `descripcionYCuerpo`, si no empiezan por `description: >-` y su salto o no son al menos tres: `skills/<skill>/SKILL.md: la description no está en un bloque >- de más de una línea`) y de la longitud (más de `maximoDeUnaDescripcionCorta` runas: `skills/<skill>/SKILL.md: la description tiene <n> caracteres (máximo 450)`); al final, si la suma de todas las skills de `skills/` pasa de `maximoDeLasDescripcionesCortas`, `skills: las descripciones suman <n> caracteres (máximo 1700)`; la forma y la longitud se miden en toda skill, tenga o no texto; el error, sin defectos, el de un directorio o fichero ilegible o un frontmatter que no lee `skills.LeerFrontmatter`, con su ruta delante. **Auxiliares de los mutantes** (plan.md «Trazabilidad»): `copiaDeLasSkillsYLaVariante(t)`, que copia `skills/` y `laboratorio/producto-descripciones-cortas/` del repositorio a `t.TempDir()` con `os.CopyFS` y da la raíz de la copia, y `cambiarElSkill(t, raiz, skill, de, a)`, que sustituye la primera aparición con `leerFichero` y `escribirFichero` y exige que cambie algo, los dos con `t.Helper()`. **Test** (§1.2 y §1.3): `TestDescripcionesDelProducto`, un subtest «la rama» sobre la raíz del repositorio que exige cero líneas y uno por cada caso de la tabla de §1.3 sobre su copia, con sus líneas exactas y en su orden (`compararLasLineasDeLosDefectos`): un carácter añadido en `normativa`, uno quitado en `consulta-juridica` y uno cambiado en `plazos-y-territorio` (US1.2); `jurisprudencia` en un escalar de una línea con el `SKILL.md` que da `skillConLaDescripcion` y su texto de la variante, y `plazos-y-territorio` en un bloque `>-` de una línea (FR-002); `normativa` de 451 runas (con la suma de 1715) y `plazos-y-territorio` de 438 con la suma de 1701 (FR-004); y `skills/quinta/SKILL.md` sin texto en la variante, con la suma de 1710 (caso límite del spec). Con las descripciones de T002 la rama sale en verde; cada mutante es la medida pasada de su umbral y su subtest exige el rojo (ADR 0029).

**Checkpoint**: las cuatro descripciones son las de la variante y `make ci` lo comprueba con sus mutantes.

---

## Phase 4: User Story 2 — Nada más de las skills cambia (P1)

**Goal**: fuera de las líneas de la clave, ninguna de las cuatro skills cambia; `normativa`, `jurisprudencia` y
`plazos-y-territorio` ya lo comprueba T002 con `TestSkillsDelProductoDesdeControl` (FR-005); falta `consulta-juridica`,
contra la base de la rama (FR-003).

**Independent Test**: `go test -run '^TestConsultaJuridicaDesdeLaBase$' ./internal/evals/` en verde con la rama, con el
defecto exacto en los cuatro mutantes de contracts/controles.md §2.3 y sin defecto cuando solo cambia la clave.

- [ ] T004 [US2] `TestConsultaJuridicaDesdeLaBase`, el control de FR-003 para `consulta-juridica`, con sus seis casos (FR-003; SC-002; contracts/controles.md §2; research D2, V3, V8). **Fichero**: internal/evals/descripciones_del_producto_test.go, el de T003, al que solo se añade. **Lo fijado** (§2.1): `huellasDeConsultaJuridicaEnLaBase`, la SHA-256 en hexadecimal de cada fichero de `skills/consulta-juridica/` en `ab951c22` por su ruta relativa a la skill, la del `SKILL.md` de su contenido sin las líneas de la clave: un solo fichero, `SKILL.md` → `bd089909060d0fd82c1c6d2ee4ae148f6f0a1b2b7d7eefd476d164dc610a205a`, con su comentario que dice de dónde sale (la base de la rama, `ab951c227e1fc8acac8bd6b23534e275e4a36a76`, y cómo se calculó). **Comprobación** (§2.2): `defectoDeConsultaJuridica(raiz string) (string, error)` lee `skills/consulta-juridica/` con `leerLosFicheros`, quita al `SKILL.md` las líneas de la clave con `descripcionYCuerpo` y, en orden de ruta sobre la unión de las dos listas, da el primero de `skills/consulta-juridica/<ruta> falta: está en la base de la rama (ab951c22)`, `skills/consulta-juridica/<ruta> sobra: no está en la base de la rama (ab951c22)` o `skills/consulta-juridica/<ruta> difiere de la base de la rama (ab951c22)`, o la cadena vacía; el error, sin defecto, el de `leerLosFicheros`. **Test** (§2.3): un subtest «la rama» sobre la raíz del repositorio sin defecto, y uno por caso sobre la copia de `copiaDeLasSkillsYLaVariante`, con `cambiarElSkill` donde el caso cambia el `SKILL.md`: el cuerpo (`## Método` → `## Métodos`) y `name` (`name: consulta-juridica` → `name: consulta-juridico`), los dos con `skills/consulta-juridica/SKILL.md difiere de la base de la rama (ab951c22)`; un fichero de más (`references/otra.md` con `# Otra\n`) y uno de menos (sin `SKILL.md`), con su ruta y su clase; y solo la clave (`Punto de entrada de toda` → `Punto de entrada de cualquier`), sin defecto (US2.3). Cada mutante es un byte o un fichero de diferencia, la medida pasada del umbral de 0, y su subtest exige el rojo (ADR 0029).

**Checkpoint**: lo demás de las cuatro skills está fijado en `make ci`: tres familias contra su control y
`consulta-juridica` contra la base.

---

## Phase 5: User Story 3 — Las suites congeladas, el laboratorio y lo generado siguen como estaban (P1)

**Goal**: el mantenedor ve `make ci` en verde con un único cambio declarado en una suite congelada (T002), el laboratorio
intacto, nada generado que regenerar y el cambio dicho en el changelog (FR-006 a FR-009).

**Independent Test**: `make ci` en verde en la cabeza de la rama; `git diff --name-only ab951c22` solo nombra los
ficheros de T001 a T005 y los del directorio del feature (quickstart, escenario 4); `make skills-sync` no cambia ningún
fichero; la viñeta de `CHANGELOG.md` está en «[Unreleased]» → «Cambiado».

- [ ] T005 [US3] La entrada del changelog y la verificación final de la rama (FR-006, FR-007, FR-008, FR-009; SC-004; DoD §1; contracts/skills.md §4; research D9). **Fichero**: CHANGELOG.md: al final de «## [Unreleased]» → «### Cambiado», la viñeta de contracts/skills.md §4 byte a byte, que nombra las cuatro skills, dice que su `description` pasa a ser la de la variante `producto-descripciones-cortas` medida en H41.3, que nada más de las skills cambia y que es para quitar la caída de la eval 04 de `normativa` en modo herramienta (ADR 0055); nada más del fichero cambia. **Verificación**, sin cambiar otro fichero: `make ci` en verde; `make skills-sync` seguido de `git status --porcelain` sin ningún fichero (FR-007); `git diff --name-only ab951c22` (el árbol de trabajo contra la base) sin ninguna ruta bajo `laboratorio/`, `evals/`, `web/`, `mcp/`, `internal/empaquetado/`, `.github/` ni de un test congelado distinto de internal/evals/laboratorio_familias_test.go (FR-006, FR-009; quickstart, escenario 4). Sin ADR nuevo (el ADR 0055 ya decide el cambio), sin `docs/SOURCES.md` (ninguna fuente), sin esquemas de salida (ninguna salida de applet cambia) y sin fixtures (ninguna grabación): plan.md «Gates» y «Datos externos».

**Checkpoint**: la rama está completa para el cierre del workflow.

---

## Phase 6: User Story 4 — La decisión con el sondeo y con las evals del cierre (P2)

Sin tareas. La medición de las evals de las cuatro skills en los dos modos y el juicio de la tercera persona de sus
descripciones los hace el cierre del workflow tras la revisión final, con el job de evals que el hito no toca (FR-010,
SC-005; plan.md «Controles de umbral», filas `evals:`); el sondeo de las dos filas de 320 sesiones de la eval 04 se lanza
fuera del run con las órdenes del quickstart, escenario 8, que ya existen (FR-011, SC-006; ADR 0055, punto 4). Ninguna
tarea lo ejecuta, lo prepara ni lo espera.

---

## Controles de umbral → tareas

Cada fila de plan.md «## Controles de umbral», con la tarea que construye su control o que pone la medida bajo un
control que ya existe, y el test que lo ve fallar con la medida pasada del umbral (ADR 0029).

| Fila (requisito · umbral) | Control | Tarea | Su rojo |
|---|---|---|---|
| FR-001, SC-001 · 0 bytes de diferencia con la variante; 0 skills sin texto | `TestDescripcionesDelProducto` (nuevo) | T003 | los mutantes de un carácter añadido, quitado y cambiado y el de la quinta skill (§1.3) |
| FR-002 · 0 skills fuera de un bloque `>-` de dos o más líneas sangradas | `TestDescripcionesDelProducto` (nuevo) | T003 | los mutantes del escalar y del bloque de una línea (§1.3) |
| FR-004, SC-003 · 450 por descripción y 1700 la suma | `TestDescripcionesDelProducto` (nuevo) | T003 | los mutantes de 451 y de 1701 (§1.3) |
| FR-004, SC-003 · 900 y 0 etiquetas XML | `TestSkillsDelRepositorio` (`ValidarFrontmatter`), sin cambiar | T002 pone las descripciones bajo él | los casos de 901 caracteres y de etiqueta XML de internal/skills/frontmatter_test.go, que no cambian |
| FR-003, SC-002 · `consulta-juridica`: 0 bytes fuera de la clave, 0 ficheros de más o de menos | `TestConsultaJuridicaDesdeLaBase` (nuevo) | T004 | los mutantes del cuerpo, `name`, un fichero de más y uno de menos (§2.3) |
| FR-003, FR-005, SC-002 · las tres familias: 0 bytes fuera de lo generado y de la clave; frontmatter del control; `description` de la variante | `TestSkillsDelProductoDesdeControl` (cambiado) | T002 | los mutantes de contracts/controles.md §3.3 sobre la copia de `$TMPDIR` (research V7) |
| FR-006 · 0 bytes en la variante | `TestDescripcionesCortas`, sin cambiar | T002 y T005 lo corren sin tocar `laboratorio/` | sus propios casos de H41.3, que no cambian |
| FR-006 · 0 bytes en las variantes de control | `TestVariantesDeControl`, sin cambiar | T002 y T005 lo corren sin tocar `laboratorio/` | sus huellas por ruta, que no cambian |
| FR-007 · 0 ficheros que cambie `make skills-sync` | `TestSkillsDelRepositorio`, sin cambiar | T002 y T005 (`make skills-sync` + `git status --porcelain`) | su comparación con lo regenerado, que no cambia |
| FR-009, SC-004 · 0 tests en rojo | `make test` en `make ci` | todas, T005 al final | cualquier test en rojo |
| FR-009 · `TestConLaDescripcion` y el mutante de `plazos-y-territorio` de `TestEleccionDeLosMutantes` en verde sin cambios | los dos tests, tal cual | T002 (la clave sigue en un bloque `>-`) | sus propios casos, que no cambian |
| FR-001, FR-002, SC-001 · 8 de 8 `grep` sobre lo instalado | el guion congelado `h41-4-skills-instaladas.txtar` | T001 lo escribe; T002 lo pone en verde | hoy, `skills-instaladas.txtar:22: no match` (research V9), que el workflow comprueba al congelarlo |
| FR-010, SC-005 · las evals de las cuatro skills y el juicio de la tercera persona en verde | el job de evals del cierre, sin cambiar | ninguna: lo mide el cierre del workflow (Phase 6) | el veredicto del job con la eval o la descripción que cae |

---

## Dependencies & Execution Order

### Orden de las fases

- **Phase 1 (T001)**: primero, sin dependencias; congela la aceptación.
- **Phase 3 (T002, T003)**: T002 cambia las descripciones y el test que con ellas saldría en rojo; T003 necesita las
  descripciones de T002 (con las de hoy, «la rama» de `TestDescripcionesDelProducto` sale en rojo).
- **Phase 4 (T004)**: después de T003, porque añade al mismo fichero y reutiliza `copiaDeLasSkillsYLaVariante` y
  `cambiarElSkill`.
- **Phase 5 (T005)**: la última; verifica la rama entera.
- **Phase 6**: sin tareas.

### Dentro de cada tarea

El cambio y su test van juntos y la tarea termina con `make ci` en verde; los mutantes se ven fallar antes de darla por
buena.

### Oportunidades de paralelismo

Ninguna: T003 y T004 escriben el mismo fichero, T003 parte de T002 y T005 verifica lo anterior.

---

## Parallel Example

No aplica: las cinco tareas van en secuencia (T001 → T002 → T003 → T004 → T005).

---

## Implementation Strategy

### MVP

T001 y T002: las cuatro skills con las descripciones de la variante, el guion en verde y `TestSkillsDelProductoDesdeControl`
cambiado. Es lo que entrega el hito (US1); T003 y T004 lo protegen en `make ci` y T005 lo deja dicho.

### Entrega incremental

T001 (rojo congelado) → T002 (verde: entrega) → T003 (control de la descripción, su forma y su longitud) → T004 (control
de `consulta-juridica`) → T005 (changelog y verificación final). Cada paso deja `make ci` en verde.

---

## Notes

- Lo que el plan deja para la persona (S1 y que las dos tablas del sondeo se sacan por separado: plan.md «Cierre y
  pendientes») va en el informe final del workflow, no en una tarea.
- Si la medición del cierre da un rojo, la reparación no cambia ninguna descripción para que pase (FR-010).
