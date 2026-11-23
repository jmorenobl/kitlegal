# Feature Specification: H45 · El molde admite un paquete con sus plazos y su norma, sin código (ADR 0062)

**Feature Branch**: `042-h45-el-molde-admite`

**Created**: 2026-11-23

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H45-el-molde-admite-un-paquete-sin-codigo.md`, que aplica los
puntos 1 a 5, 7 y 8 del ADR 0062 (`docs/ADR/0062-un-paquete-trae-sus-plazos-y-lo-suyo-de-lo-regulado-aparte.md`). El
punto 6 (el cómputo de la revisión tributaria) es de H42.

## Contexto

El principio X promete que una materia entra como paquete —sus normas, sus tipos de plazo, lo suyo de lo regulado
aparte, sus referencias y sus evals— sin tocar Go ni el protocolo de ninguna skill. Un prototipo del paquete tributario
de H42, montado el 2026-11-22 sobre `main` (`c5cf9a1b`) fuera del repositorio, enseñó que hoy no es verdad: la región
de los plazos escribe todos los tipos en la referencia de cada paquete; las filas de lo regulado aparte van al paquete
de su precepto (en todas, la LPAC) y no al de la norma que las rige; el cuerpo de `normativa` no tiene sitio para la
línea de un paquete más y la de la vertical de ensayo; las grabaciones de un hito nuevo piden una constante en Go; una
veintena de tests fijan con una lista literal lo que hay en `data/` y en `evals/`; el trabajo de evals de
`plazos-y-territorio` no admite ninguna eval más; y la descripción de `plazos-y-territorio` nombra una materia.

H45 arregla el molde, en Go y genérico, y lo valida con sus dos implementaciones: lo administrativo, que sale igual, y
la vertical de ensayo, que crece con un tipo de plazo y una fila con norma. H42 entra después solo con datos.

**Comprobado en el repositorio al escribir este spec** (lecturas de esta sesión, rama `main`, `c715e13a`):

- `filasDeLosRecursosYPlazos` (`internal/skills/regiones.go`, línea 529) no usa el contexto de la región: escribe todos
  los tipos de `data/plazos.yaml`. `filasDeLaCobertura` (línea 495) y `filasSinRegion` (línea 818) asignan cada fila de
  `regulados_aparte` por `entrada.Precepto.Norma`; el defecto de `filasSinRegion` es «`data/cobertura.yaml: la fila de
  «<procedimiento>» no la genera ninguna región cobertura_por_materia`».
- `data/cobertura.yaml`: dos filas llevan `norma` (la LGT, `BOE-A-2003-23186`, y el texto refundido de tráfico,
  `BOE-A-2015-11722`), y ninguna de las dos normas lleva `vertical` en `data/normas.yaml`; el precepto de todas las filas
  es de la LPAC (`BOE-A-2015-10565`). `data/normas.yaml` lleva diez normas con `vertical: administrativo`, entre ellas
  la LPAC, la LJCA (`BOE-A-1998-16718`) y la Ley 19/2013 (`BOE-A-2013-12887`), que son las de los preceptos y las
  unidades de los siete tipos de `data/plazos.yaml`.
- `skills/plazos-y-territorio/references/administrativo.md` mide 100 líneas y 10 635 bytes; su región
  `recursos_y_plazos` (líneas 46 a 58), 1 361 bytes; su región `cobertura_por_materia` (líneas 28 a 40), 3 185 bytes.
- El tope del cuerpo es `maximoDeCaracteresDelCuerpo = 11000` (`internal/skills/sincronia.go`, línea 31), contado en
  runas. Con `awk`, en esta sesión, los cuerpos miden 10 926 (`normativa`), 10 842 (`jurisprudencia`), 10 686
  (`plazos-y-territorio`) y 4 481 (`consulta-juridica`); el hito da 10 927, 10 843 y 10 687 con la cuenta de la
  sincronía. Las marcas de los paquetes de `normativa` y de `jurisprudencia` están vacías; la de `plazos-y-territorio`
  lleva la línea de lo administrativo. La línea de la vertical de ensayo mide 69 caracteres y su salto
  (`lineaDelEnsayo`, `internal/skills/vertical_de_ensayo_test.go`, línea 38).
- Lo que se mueve de `skills/normativa/SKILL.md` son sus líneas 38 a 41 («En PowerShell 5.1:» y su bloque
  `powershell`), 176 caracteres. `skills/normativa/references/con_herramientas.md` mide 38 líneas y 2 442 bytes, y el
  cuerpo manda leerla «Antes de la primera orden o llamada» (línea 16). Ningún guion de `internal/app/testdata/script/`
  ni copia congelada de `specs/*/aceptacion/` lleva «En PowerShell 5.1» ni `LASTEXITCODE` de `normativa`.
- `UnionDeGrabaciones` (`internal/evals/grabaciones.go`, línea 74) devuelve `GrabacionesDeH4`, `GrabacionesDeH5`,
  `GrabacionesDeH9` y `GrabacionesDeH36` (`../../evidencias/h36/boe.legislacion-consolidada`). En `evidencias/`, solo
  `h36/` tiene una carpeta `boe.legislacion-consolidada/`; `h9/` tiene `boe.fiestas-laborales/` y
  `bocm.fiestas-locales/`; `h6/`, `h23/`, `adr-*` y `sondeos-*` tienen `manifiesto.json` y ninguna. `grabar-datos.sh`
  (línea 100) nombra la carpeta de un hito con su id en minúsculas y el punto como guion (`H9.1` → `h9-1`).
- `.github/workflows/evals.yml`: concurrencia 1 (`consulta-juridica`), 3 (`jurisprudencia`), 5 (`normativa`) y 2
  (`plazos-y-territorio`), once en total; `sesionesALaVezComoMucho = 11` (`internal/evals/definicion_test.go`, línea
  162); el plan `[]int{67, 66, 6}` con `Concurrencia: 2` (línea 3242); el tope del repositorio, 352 min (línea 1622).
  `evals/plazos-y-territorio/` tiene 12 evals; `evals/normativa/`, 25 ficheros.
- La descripción de `plazos-y-territorio` mide hoy 427 caracteres, y la de «Entrega», 429 (contados con `wc -m` en
  esta sesión). Las cuatro de `laboratorio/producto-descripciones-cortas/` suman hoy 1 690 sin sus saltos (418, 421,
  427 y 428 con el salto, `wc -m`); con la nueva, 1 692.
- La fila del control en `docs/SKILLS.md` (línea 215): «`ordenes-para-powershell` | `normativa` | Cada orden de
  lectura o de comprobación lleva detrás, en el mismo fichero, su forma para PowerShell».

## Clarifications

### Session 2026-11-23

- Q: ¿En qué referencias escritas a mano de la skill puede estar la forma para PowerShell de una orden del SKILL.md para
  que el control ordenes-para-powershell la dé por buena (FR-011, §C.3)? → A: En cualquiera de la misma skill (las que da
  `skills.ReferenciasAMano`), se lea siempre o solo en algunos casos: el control no interpreta lo que el cuerpo dice sobre
  cuándo se lee cada referencia. Las únicas condiciones son el mismo verbo, la misma norma y los mismos bloques; una orden
  de una referencia sigue pidiendo su forma detrás, en su mismo fichero. Los mutantes de
  `TestControlesSobreLasReferenciasAMano` son los cinco de §C.3, sin ninguno más (auto: criterio a; fuente: el hito H45,
  «Entrega» y «Controles»; FR-011, §C.3 y SC-008; constitución, «Criterio de decisión autónoma», punto 2).

Fuera de esa pregunta, el fichero del hito y el ADR 0062 fijan las reglas, los controles y la aceptación. Las dos
lecturas que el hito no escribe con todas sus palabras —qué sitio exige `TestVerticalDeEnsayo` en cada cuerpo y qué
pone en rojo la descripción de hoy— están en «Assumptions», con su porqué.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Con los datos de hoy, todo sale como en `main` (Priority: P1)

Quien usa las skills recibe las mismas referencias generadas, el mismo `kitlegal plazos` y los mismos cuerpos de
`SKILL.md`, salvo el movimiento de `normativa` y la descripción de `plazos-y-territorio`.

**Why this priority**: lo administrativo es la implementación que protege lo que ya funciona (principio X); el molde
nuevo no puede cambiar lo que lee hoy una skill.

**Independent Test**: `make skills-sync` sobre la rama no deja diferencias, y con los datos de `main` la salida del
binario y lo que comprueban `TestSkillsDelProductoDesdeControl`, `TestCasosDePlazos` y los goldens de H9 y H36 son los
mismos que en `main`, aunque lo esperado se escriba de otra forma, como piden FR-010, FR-014, FR-016, FR-017 y FR-023
(FR-008, FR-026).

**Acceptance Scenarios**:

1. **Dado** `data/` de `main`, **Cuando** se ejecuta `make skills-sync` en la rama, **Entonces** ninguna referencia
   generada de `skills/` cambia, byte a byte, y `references/administrativo.md` de `plazos-y-territorio` lleva los
   siete tipos (nueve filas de su tabla) y las nueve filas de lo regulado aparte de hoy.
2. **Dado** el árbol de la rama, **Cuando** se compara con `main`, **Entonces** los cuerpos de `SKILL.md` solo cambian
   en las líneas movidas de `normativa`, y los frontmatter solo en la `description` de `plazos-y-territorio`; de las
   referencias escritas a mano solo cambia `skills/normativa/references/con_herramientas.md`, que gana lo movido al
   final.
3. **Dado** el binario de la rama, **Cuando** corren `TestCasosDePlazos` y los goldens de H9 y H36, **Entonces** dan lo
   mismo que en `main`.
4. **Dado** la sincronía con la regla de hoy (todos los tipos en cada referencia con la región) y un tipo de ensayo en
   una copia, **Cuando** corre `TestVerticalDeEnsayo`, **Entonces** sale en rojo, porque el tipo aparece también en
   `administrativo.md`.

---

### User Story 2 - La vertical de ensayo entra con un tipo y una fila con norma, sin Go ni `SKILL.md` a mano (Priority: P1)

Quien escribe un paquete añade sus normas con `vertical`, sus tipos de plazo, sus filas de lo regulado aparte con su
norma y sus referencias, ejecuta `make skills-sync`, y cada tipo y cada fila sale solo en la referencia de su paquete.

**Why this priority**: es el objetivo del hito y lo que H42 necesita para entrar solo con datos.

**Independent Test**: `TestVerticalDeEnsayo` crecida (FR-019, FR-020).

**Acceptance Scenarios**:

1. **Dado** la vertical de ensayo con un tipo cuyo precepto es de una norma de ensayo y con el cómputo `lpac`, **Cuando**
   se instala en una copia y se sincroniza, **Entonces** el tipo sale en la región `recursos_y_plazos` de
   `skills/plazos-y-territorio/references/ensayo.md` y en ninguna otra referencia.
2. **Dado** una fila de lo regulado aparte con su precepto en la LPAC y su norma de ensayo, **Cuando** se sincroniza la
   copia, **Entonces** la fila sale en `ensayo.md` y no en `administrativo.md`, con la misma forma de línea que hoy.
3. **Dado** la copia sincronizada, **Cuando** se mide, **Entonces** no hay defectos ni derivas, el diff de la copia no
   tiene ningún `.go`, ningún `SKILL.md` cambia fuera de las marcas que genera la sincronía, y cada cuerpo de skill de
   capacidad cabe en 11 000 caracteres con la línea de ensayo y sitio para otra de su misma longitud.
4. **Dado** cada mutante —un tipo asignado a todas las referencias; una fila asignada por su precepto; `filasSinRegion`
   con la regla vieja y una referencia de ensayo sin la región; el cuerpo de `normativa` sin el movimiento—, **Cuando**
   corre `TestVerticalDeEnsayo`, **Entonces** sale en rojo.

---

### User Story 3 - Un tipo que no es de un paquete, o que es de dos, lo dice la sincronía (Priority: P1)

Quien escribe un paquete y se equivoca en las normas de un tipo recibe de `make skills-sync` un defecto con el código
del tipo, en lugar de un tipo que sale en la referencia equivocada o en ninguna.

**Why this priority**: sin el defecto, el tipo de un paquete mal declarado entra en silencio en otra materia, y la skill
daría un plazo de una materia con la referencia de otra (principio II).

**Independent Test**: subtests de la sincronía con copias de `data/` (FR-003, FR-004).

**Acceptance Scenarios**:

1. **Dado** un tipo cuyos preceptos (de `preceptos` y de `unidades`) son de normas sin `vertical`, **Cuando** se
   sincroniza, **Entonces** sale el defecto de §C.1 con su código, y la skill con la región no lleva nada regenerado.
2. **Dado** un tipo con un precepto de una norma de lo administrativo y otro de una norma de ensayo, **Cuando** se
   sincroniza, **Entonces** sale el defecto de §C.1 con su código y los dos paquetes.
3. **Dado** un tipo cuyos preceptos propios son de un paquete y cuyo cómputo cita normas de otro, **Cuando** se
   sincroniza, **Entonces** no hay defecto: los preceptos del cómputo no cuentan.
4. **Dado** la sincronía sin la regla, **Cuando** corren esos subtests, **Entonces** salen en rojo.

---

### User Story 4 - Un paquete más no rompe ningún test ajeno (Priority: P1)

Quien añade un tipo, una fila, una norma con `vertical` o una eval no paga con la edición de tests que no son de ese
dato.

**Why this priority**: es lo que hace que H42 entre sin Go (ADR 0062, punto 5).

**Independent Test**: la comprobación de SC-004 en `make ci`, que ejecuta los tests de la lista de FR-016 sobre una
copia de `data/` y de `evals/` con un paquete más (FR-016, FR-027).

**Acceptance Scenarios**:

1. **Dado** una copia de `data/` con un tipo, una fila y una norma con `vertical` de más, y una copia de
   `evals/plazos-y-territorio/` con una eval de más, **Cuando** corren los tests de la lista, **Entonces** siguen en
   verde.
2. **Dado** uno de esos tests con su lista literal de hoy devuelta, **Cuando** corre sobre esa copia, **Entonces** sale en
   rojo.
3. **Dado** cada test de la lista, **Cuando** se le aplica el mutante que lo ponía en rojo en `main`, **Entonces** sigue
   en rojo.

---

### User Story 5 - `normativa` hace sitio y el control de PowerShell lo admite (Priority: P1)

El modelo que usa `normativa` lee la forma para PowerShell 5.1 de la orden de lectura del paso 2 en
`references/con_herramientas.md`, que el cuerpo le manda leer antes de la primera orden, y el cuerpo gana el sitio de
un paquete real más.

**Why this priority**: sin el sitio, la línea del paquete de H42 y la de la vertical de ensayo no caben (ADR 0062,
punto 3).

**Independent Test**: `TestEvalsDelRepositorio/ordenes-para-powershell`, `TestControlesSobreLasReferenciasAMano` y
`TestSkillsDelProductoDesdeControl` (FR-009 a FR-013).

**Acceptance Scenarios**:

1. **Dado** la rama, **Cuando** se lee `skills/normativa/SKILL.md`, **Entonces** no lleva «En PowerShell 5.1:» ni su
   bloque, y `references/con_herramientas.md` los lleva al final, con el mismo texto sin la sangría de dos espacios.
2. **Dado** una orden del `SKILL.md` con su forma para PowerShell en una referencia a mano de la skill, con el mismo
   verbo, la misma norma y los mismos bloques, **Cuando** corre el control, **Entonces** no da defecto.
3. **Dado** cada mutante de §C.3, **Cuando** corre el control, **Entonces** da su defecto.
4. **Dado** la declaración de `normativa` del laboratorio de H41, **Cuando** corre `TestSkillsDelProductoDesdeControl`,
   **Entonces** pasa con lo movido declarado como un movimiento, y sale en rojo si el texto movido sigue en el cuerpo.

---

### User Story 6 - Un paquete entra sin tocar el laboratorio de H41 (Priority: P2)

Quien añade un paquete con su `references/<vertical>.md` en una skill del producto no tiene que declararla en el
laboratorio de H41.

**Why this priority**: sin esto, H42 tendría que tocar Go (`laboratorio_familias_test.go`).

**Independent Test**: `TestSkillsDelProductoDesdeControl` con una copia (FR-014).

**Acceptance Scenarios**:

1. **Dado** una skill del producto con una `references/<vertical>.md` que su base no tiene, con `<vertical>` un valor de
   `vertical:` de `data/normas.yaml`, **Cuando** corre el test, **Entonces** sigue en verde sin declararla.
2. **Dado** la misma referencia con un nombre que no es una vertical, **Cuando** corre el test, **Entonces** sale en rojo.
3. **Dado** un cambio en una referencia que la base ya tiene, **Cuando** corre el test, **Entonces** sale en rojo.

---

### User Story 7 - Las grabaciones de un hito entran sin Go (Priority: P2)

El job de evals y los tests reproducen con las grabaciones de cada hito que graba el BOE sin que ese hito añada una
constante.

**Why this priority**: H42 grabará en `evidencias/h42/` y no puede tocar Go (ADR 0062, punto 4).

**Independent Test**: `TestGrabacionesSinSolape` y `TestManifiestoDeGrabaciones` (FR-015).

**Acceptance Scenarios**:

1. **Dado** el repositorio, **Cuando** se pide la unión, **Entonces** es H4, H5, H9 y `evidencias/h36/…`, en ese orden, y
   `evidencias/h9/` no entra.
2. **Dado** una copia con `evidencias/h99/manifiesto.json` y `evidencias/h99/boe.legislacion-consolidada/`, **Cuando** se
   pide la unión, **Entonces** entra la carpeta de h99, detrás de la de h36, sin cambiar Go.
3. **Dado** un orden que pone h36 detrás de una raíz de número mayor, o un fichero repetido entre dos raíces, **Cuando**
   corren los tests, **Entonces** salen en rojo.

---

### User Story 8 - El trabajo de evals de `plazos-y-territorio` admite las evals de H42 (Priority: P2)

La persona que lee el cierre recibe la medida de `plazos-y-territorio` dentro del tope de tiempo del trabajo, también
con las cuatro evals que añadirá H42.

**Why this priority**: con concurrencia 2, una eval más pasa del tope de 352 min (ADR 0062, punto 7).

**Independent Test**: `TestDefinicionDelJob` (FR-021, FR-022).

**Acceptance Scenarios**:

1. **Dado** `.github/workflows/evals.yml` de la rama, **Cuando** corre `TestDefinicionDelJob`, **Entonces** las cuatro
   skills caben en su tope con doce sesiones a la vez como mucho.
2. **Dado** una copia de `evals/plazos-y-territorio/` con cuatro evals de más, **Cuando** corre el test, **Entonces** el
   peor caso de `plazos-y-territorio` cabe en 352 min.
3. **Dado** `plazos-y-territorio` a dos, o trece sesiones a la vez, **Cuando** corre el test, **Entonces** sale en rojo.

---

### User Story 9 - La descripción de `plazos-y-territorio` nombra su capacidad (Priority: P2)

El agente que elige skill lee una descripción que cuenta los plazos de las materias que la skill cubre y declara las
demás, sin nombrar la LPAC ni el contencioso.

**Why this priority**: con un segundo paquete, la de hoy deja de ser verdad y la skill podría no activarse ante un plazo
de otra materia (ADR 0062, punto 8).

**Independent Test**: `TestDescripcionesDelProducto`, `TestDescripcionesCortas`, `TestSkillsDelProductoDesdeControl`, el
guion `h41-4-skills-instaladas.txtar` y el juez de las descripciones del cierre (FR-023).

**Acceptance Scenarios**:

1. **Dado** la rama, **Cuando** se lee la `description` de `skills/plazos-y-territorio/SKILL.md` y la de la variante
   `laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt`, **Entonces** las dos son el texto de FR-023,
   byte a byte.
2. **Dado** la descripción de hoy, que nombra la LPAC, en el `SKILL.md` o en la variante, **Cuando** corre `make ci`,
   **Entonces** sale en rojo.
3. **Dado** el cierre, **Cuando** el job de evals mide `plazos-y-territorio`, **Entonces** la skill se sigue activando
   ante sus preguntas de hoy y el juez de las descripciones no la marca.

---

### Edge Cases

- Un tipo con preceptos de una norma con `vertical` y de otra sin él: es del paquete de la que lo lleva (FR-001).
- Un tipo sin `preceptos` y con `unidades` (como `generico` hoy): su paquete sale de los de `unidades` (FR-001).
- Una fila con `norma` de una norma que no está en `data/normas.yaml`, o que está sin `vertical`: va al paquete de su
  precepto, como hoy (FR-005); con los datos de hoy, las filas de la LGT y del tráfico.
- Un paquete con tipos cuya referencia de `plazos-y-territorio` no lleva la región `recursos_y_plazos`: sus tipos no
  salen en ninguna referencia; el hito no lo declara defecto (Fuera de alcance).
- Una carpeta de `evidencias/` con `boe.legislacion-consolidada/` y sin `manifiesto.json`, o con `manifiesto.json` y sin
  esa carpeta, no entra en la unión (FR-015).
- Lo que llega manipulado a mano (un `data/*.yaml` que no se puede leer, una carpeta de `evidencias/` truncada) lo cubre
  la regla genérica: la sincronía da el defecto de su lector o el error de lo que no se puede leer, y el binario, defecto
  `inesperado` con código 1 (ADR 0023), sin casos propios.

## Requirements *(mandatory)*

### Functional Requirements

**El paquete de un tipo de plazo (ADR 0062, punto 1)**

- **FR-001**: El paquete de un tipo de `data/plazos.yaml` es el `vertical` que comparten, en `data/normas.yaml`, las
  normas de sus preceptos propios: los de `preceptos` y los de `unidades`. Las normas sin `vertical` no cuentan; los
  preceptos del cómputo tampoco. Comprobable: US2.1, US3.3; con los datos de hoy, los siete tipos son de lo
  administrativo (US1.1).
- **FR-002**: La región `recursos_y_plazos` de la referencia de un paquete lista solo los tipos de ese paquete, en el
  orden de `data/plazos.yaml` y con la forma de fila de hoy. Comprobable: US1.1 (sin diferencias), US2.1 y el mutante
  «un tipo asignado a todas las referencias» (US2.4).
- **FR-003**: Un tipo cuyas normas no llevan ningún `vertical` es un defecto de la sincronía, con el texto de §C.1 y el
  código del tipo, como los demás defectos de `data/plazos.yaml`: la skill con la región no lleva nada regenerado y
  `make skills-sync` falla. Comprobable: US3.1, US3.4.
- **FR-004**: Un tipo cuyas normas llevan dos o más `vertical` distintos es un defecto de la sincronía, con el texto de
  §C.1, el código del tipo y sus paquetes. Comprobable: US3.2, US3.4.

**El paquete de una fila de lo regulado aparte (ADR 0062, punto 2)**

- **FR-005**: Una fila de `regulados_aparte` de `data/cobertura.yaml` es del paquete de la norma que la rige si la lleva
  (`norma`) y esa norma tiene `vertical` en `data/normas.yaml`; si no, del de la norma de su precepto, como hoy. La
  región `cobertura_por_materia` de la referencia de un paquete genera solo sus filas. Comprobable: US2.2 y el mutante
  «una fila asignada por su precepto» (US2.4); con los datos de hoy, ninguna fila cambia de referencia (US1.1).
- **FR-006**: La comprobación de que cada fila la genera una región (`filasSinRegion`) usa la misma regla que FR-005:
  una fila de un paquete cuya referencia no lleva la región es un defecto, con el texto de hoy («`data/cobertura.yaml:
  la fila de «<procedimiento>» no la genera ninguna región cobertura_por_materia`»). Comprobable: el mutante
  «`filasSinRegion` con la regla vieja y una referencia de ensayo sin la región» deja de dar el defecto y
  `TestVerticalDeEnsayo` sale en rojo (US2.4).
- **FR-007**: La línea de cada fila (la que sigue la plantilla `⚠ NO CUBIERTO:` de la skill) no cambia: nombra la norma
  que la rige y el precepto que dice que va aparte (ADR 0058, puntos 1 y 2). Comprobable: US1.1 y US2.2.

**Lo de hoy, igual**

- **FR-008**: Con `data/` de `main`, `make skills-sync` no cambia ninguna referencia generada, y los `SKILL.md` solo
  cambian en el movimiento de FR-009 y en la `description` de FR-023. Comprobable: US1.1 y US1.2, con la comprobación
  de las regiones de `make ci` y `TestSkillsDelProductoDesdeControl`.

**El sitio del cuerpo de `normativa` (ADR 0062, punto 3)**

- **FR-009**: Las líneas «En PowerShell 5.1:» y su bloque `powershell` de la orden de lectura del paso 2 de
  `skills/normativa/SKILL.md` salen del cuerpo y van, tal cual salvo la sangría de dos espacios, al final de
  `skills/normativa/references/con_herramientas.md`, en §C.2. No se mueve nada más: ni una orden de Bash, ni una línea
  `⚠`, ni una regla, ni «Por qué:» ni «Ojo:» del paso 3 (que congela `h32-normativa.txtar`). Comprobable: US5.1 y
  `TestSkillsDelProductoDesdeControl`, que sale en rojo con cualquier otro cambio.
- **FR-010**: La declaración de `normativa` en el laboratorio de H41 (`laboratorio_familias_test.go`) declara lo de
  FR-009 como un movimiento a `references/con_herramientas.md`. Comprobable: US5.4.
- **FR-011**: El control `TestEvalsDelRepositorio/ordenes-para-powershell` admite que la forma para PowerShell de una
  orden del `SKILL.md` esté en una referencia escrita a mano de la misma skill (cualquiera de las que da
  `skills.ReferenciasAMano`, se lea siempre o solo en algunos casos: el control no interpreta lo que el cuerpo dice sobre
  cuándo se lee), con el mismo verbo, la misma norma y los mismos bloques; no hay más condiciones. Una orden de una
  referencia sigue pidiendo la suya detrás, en su mismo fichero. Comprobable: US5.2 y los mutantes de §C.3 (US5.3).
- **FR-012**: Los mutantes de `TestControlesSobreLasReferenciasAMano` crecen con los casos de §C.3, los cinco y ninguno
  más. Comprobable: cada uno da su defecto.
- **FR-013**: La fila de `ordenes-para-powershell` en `docs/SKILLS.md` dice la regla de FR-011, incluida la ausencia de
  condición sobre cuándo se lee la referencia. Comprobable: lectura del fichero en la revisión.

**El laboratorio de H41**

- **FR-014**: `TestSkillsDelProductoDesdeControl` deja de exigir en la declaración de una skill las
  `references/<vertical>.md` que su base (la variante `desde`, si la declara, o su skill de control) no tiene, con
  `<vertical>` cada valor de `vertical:` de `data/normas.yaml`. La de un paquete que la base ya tiene se sigue
  comparando. Comprobable: US6.1 a US6.3.

**Las grabaciones (ADR 0062, punto 4)**

- **FR-015**: `UnionDeGrabaciones` devuelve las raíces de H4, H5 y H9 y, detrás, la carpeta
  `boe.legislacion-consolidada/` de cada `evidencias/<hito>/` que tiene `manifiesto.json` y esa carpeta, con la ruta
  relativa al paquete (`../../evidencias/…`) y en el orden del número del hito, no el alfabético (§C.4). La constante de
  H36 desaparece. `TestGrabacionesSinSolape` sigue exigiendo que dos raíces solo compartan el `robots.txt`.
  Comprobable: US7.1 a US7.3.

**Los tests sin listas (ADR 0062, punto 5)**

- **FR-016**: Los tests de §C.5 dejan de fijar con una lista literal lo que hay en `data/`, `skills/` o `evals/`: los que
  comprueban el repositorio derivan lo esperado de `data/` y de `evals/`, y los que comprueban una regla usan sus datos de
  `testdata/`. La lista es la del hito más los que el plan encuentre con un prototipo entero (incluido `make
  test-integration`). Cada caso sigue en rojo con el mutante que lo ponía en rojo. Queda fuera
  `TestNormasDelRepositorio/paquete-de-lo-administrativo`, que declara a propósito las diez normas del paquete.
  Comprobable: US4.1 a US4.3, con la comprobación de SC-004 en `make ci`.
- **FR-017**: Los guiones de `internal/app` de la lista (`h32-paquetes.txtar`, `h33-cobertura-instalada.txtar`,
  `h36-plazos-de-hoy.txtar` y `h41-4-skills-instaladas.txtar`) y sus copias congeladas de `specs/` los cambia una tarea
  `[datos]` que los declara; cada guion sigue en rojo con el mutante que lo ponía en rojo. Comprobable: el guardián de
  diff del workflow y `mutantes.md`.
- **FR-018**: Las reglas del conjunto de `normativa` (como mucho 23 evals, 10 positivas de normas distintas;
  `internal/evals/conjunto.go`) no cambian. Comprobable: diff de `conjunto.go` en esas reglas vacío frente a `main`.

**La vertical de ensayo**

- **FR-019**: `internal/skills/testdata/vertical-de-ensayo/` crece con un tipo de plazo cuyo precepto es de una norma de
  ensayo, con el cómputo `lpac`; una fila de `regulados_aparte` con su precepto en la LPAC y su `norma` de ensayo; y la
  región `recursos_y_plazos` en su `skills/plazos-y-territorio/references/ensayo.md`. Comprobable: US2.1, US2.2.
- **FR-020**: `TestVerticalDeEnsayo` exige que el tipo de ensayo salga solo en `ensayo.md`; que la fila con norma de
  ensayo salga en `ensayo.md` y no en `administrativo.md`; que todo entre sin `.go` ni `SKILL.md` escrito a mano; y que el
  cuerpo de cada skill de capacidad, con la línea de la vertical de ensayo y sitio para otra línea de paquete de su misma
  longitud (70 caracteres con el salto), quede en 11 000 caracteres como mucho. Comprobable: US2.3, US2.4.

**El trabajo de evals (ADR 0062, punto 7)**

- **FR-021**: En el trabajo `evals` de `.github/workflows/evals.yml`, `plazos-y-territorio` pasa a concurrencia 3, con su
  comentario; `ajustesDelTrabajo` lo dice, y `sesionesALaVezComoMucho` pasa de 11 a 12, que sustituye el tope de FR 020
  y SC 011 de H30. Comprobable: US8.1, US8.3.
- **FR-022**: `TestDefinicionDelJob` y `TestPlanEnDosModos` derivan el plan y el peor caso de `evals/` (sin `[]int{67, 66,
  6}` ni «5h45m25s» literales) y comprueban que cada trabajo cabe en su tope, también con una copia de
  `evals/plazos-y-territorio/` con cuatro evals de más. Comprobable: US8.2, US8.3.

**La descripción (ADR 0062, punto 8)**

- **FR-023**: La `description` de `skills/plazos-y-territorio/SKILL.md` y la variante
  `laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt` pasan a ser este texto, de 429 caracteres:

  > Calcula con kitlegal hasta qué día se puede recurrir una resolución, cuándo vence un plazo y cuándo hay silencio,
  > con los festivos que cuentan, e identifica el territorio de un municipio: INE, provincia, comunidad, DIR3 y
  > boletines. Cuenta los plazos de las materias que cubre y de las demás dice que no los cubre, sin darlos de memoria.
  > Se usa al preguntar cuándo vence un plazo o un recurso, o qué boletines tiene un municipio.

  Con ella cambian `descripcionCortaDePlazosYTerritorio` (`laboratorio_descripciones_test.go`), el guion
  `h41-4-skills-instaladas.txtar` y su copia congelada `specs/038-h41-4-las-skills-del/aceptacion/skills-instaladas.txtar`
  (tarea `[datos]`), y los subtests con sumas literales de `TestDescripcionesDelProducto` («una de 451», «una suma de
  1701», «una quinta skill sin texto») y el subtest «una que empieza por <ejemplo>» de `TestDescripcionesCortas`, que
  derivan su suma de las descripciones de hoy. El run no cambia el texto. Comprobable: US9.1, US9.2.

**El resto**

- **FR-024**: `CHANGELOG.md`, en *Unreleased* «Cambiado», dice la descripción nueva de `plazos-y-territorio` y la
  concurrencia de su trabajo de evals. Comprobable: lectura del fichero.
- **FR-025**: Ninguna eval cambia. El cierre mide todas las skills (el hito toca `internal/skills`, `internal/evals` y
  dos `SKILL.md`), sin reutilizar nada de `main`. Si una clase cae, se compara con `main` y con la serie (ADR 0038), y la
  reparación no toca otro texto de skill que el movido; la descripción no se cambia, y su rojo va al informe.
  Comprobable: diff de `evals/` vacío frente a `main`; SC-010.
- **FR-026**: `kitlegal plazos` no cambia: ni lo que calcula, ni su `data`, ni su esquema, ni sus errores.
  Comprobable: US1.3.
- **FR-027**: Cada control de «Criterios del hito (literales)» se ejecuta con su mutante o contra `main` y queda en
  `specs/042-h45-el-molde-admite/mutantes.md`, con la orden, el resultado esperado y el visto. Comprobable: el fichero,
  una entrada por control.

### Key Entities

- **Paquete**: una materia empaquetada; su nombre es un valor de `vertical:` de `data/normas.yaml`, y su referencia en
  cada skill de capacidad es `references/<vertical>.md`.
- **Tipo de plazo**: una entrada de `data/plazos.yaml` con sus preceptos propios (`preceptos`, `unidades`) y su cómputo;
  su paquete sale de FR-001.
- **Fila de lo regulado aparte**: una entrada de `regulados_aparte` de `data/cobertura.yaml` con su precepto y, a veces,
  la norma que la rige; su paquete sale de FR-005.
- **Raíz de grabaciones**: un directorio de respuestas grabadas del BOE que la reproducción copia en orden (FR-015).
- **Declaración del laboratorio**: lo que cambia y lo que se mueve una skill del producto respecto a su base (FR-010,
  FR-014).

### Quién consume cada salida, cuánto y hasta cuándo

- **`references/<paquete>.md` de `plazos-y-territorio`** (regiones `recursos_y_plazos` y `cobertura_por_materia`): la
  lee `plazos-y-territorio` cuando la pregunta es de ese paquete, por la línea de su marca de paquetes («Si la pregunta
  es de lo …, lee …»), una vez por pregunta, y usa cada fila de tipo para el código que pasa a `plazos calcular` y cada
  fila de lo regulado aparte para su línea `⚠ NO CUBIERTO:`. Con los datos de hoy no cambia: 100 líneas, 10 635 bytes;
  la región de los tipos, 1 361 bytes (nueve filas, unas 145 cada una), y la de lo regulado aparte, 3 185 (nueve
  filas, unas 340 cada una). Con
  FR-002 y FR-005 cada referencia crece con los tipos y las filas de su paquete, no con los de todos: con H42,
  `administrativo.md` sigue en sus 100 líneas y `fiscal.md` lleva solo los dos tipos y las filas tributarias. Se
  genera desde `data/`, no del uso: con meses de uso diario, cientos de normas y miles de bloques consultados, es la
  misma. Un tipo o una fila deja de salir cuando sale de `data/` o cambia de paquete, en el siguiente `make skills-sync`.
- **`references/con_herramientas.md` de `normativa`**: la lee `normativa` una vez por conversación, antes de la primera
  orden o llamada (cuerpo, línea 16); gana los 176 caracteres de FR-009 (de 2 442 a unos 2 616 bytes, sin la sangría).
  El modelo que no tiene herramientas MCP y está en PowerShell 5.1 usa la forma movida una vez por bloque leído sin
  fecha. No depende del uso acumulado.
- **El cuerpo de `normativa`**: lo carga el agente al activar la skill, una vez por conversación; baja de unos 10 927 a
  unos 10 751 caracteres.
- **La `description` de `plazos-y-territorio`**: la lee el agente en cada sesión con las skills instaladas, para elegir
  skill; pasa de 427 a 429 caracteres, y la suma de las cuatro de la variante, de 1 690 a 1 692, bajo el tope de 1 700.
- **Los defectos de FR-003 y FR-004**: los lee quien escribe un paquete, persona o run, una vez por `make skills-sync`;
  una línea por tipo mal declarado, unos 120 bytes; dejan de darse en cuanto las normas del tipo comparten un
  `vertical`.
- **`UnionDeGrabaciones`**: la usan el job de evals (`medida.go`, `sesiones.go`, `dependencias.go`) y los tests, una vez
  por preparación de una sesión; crece con una raíz por hito que graba el BOE, no con el uso.
- **El informe del trabajo de evals**: lo lee la persona al cierre; su forma no cambia. Doce sesiones a la vez pueden
  topar con el límite de la suscripción; el cierre lo enseña como fallo de la fuente (ADR 0062, «En contra»).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Con `data/` de `main`, `make skills-sync` deja 0 diferencias en las referencias generadas de `skills/`, y
  fuera de lo movido de `normativa` y de la `description` de `plazos-y-territorio` 0 bytes cambian en los ficheros de las
  skills del producto. Lo hacen cumplir, en `make ci`, la comprobación de las regiones de `internal/skills` (sale en
  rojo con una región distinta de la generada) y `TestSkillsDelProductoDesdeControl` (sale en rojo con cualquier byte no
  declarado).
- **SC-002**: En `TestVerticalDeEnsayo` crecida, el tipo de ensayo sale en 1 referencia (`ensayo.md`) y en 0 más; la fila
  con norma de ensayo, en `ensayo.md` y en 0 líneas de `administrativo.md`; el diff de la copia tiene 0 ficheros `.go` y
  0 `SKILL.md` cambiados fuera de las marcas; y cada cuerpo de skill de capacidad, con la línea de ensayo y 70
  caracteres más, mide 11 000 como mucho. Lo hace cumplir `TestVerticalDeEnsayo` en `make ci`, que sale en rojo con
  cualquiera de los cuatro mutantes de US2.4.
- **SC-003**: Los 2 defectos nuevos (FR-003, FR-004) salen con su tipo en su copia, y el caso de US3.3 da 0 defectos.
  Los subtests de la sincronía en `make ci` salen en rojo sin la regla.
- **SC-004**: Con la copia de US4.1, 0 tests de la lista de §C.5 salen en rojo; con la lista literal devuelta a uno, ese
  sale en rojo. Lo hace cumplir una comprobación de la suite de aceptación, en `make ci`, que ejecuta los tests de §C.5
  sobre la copia de `data/` y de `evals/` de US4.1 y sale en rojo si alguno falla (cómo corren sobre la copia los
  guiones de `internal/app`, con los datos empotrados en el binario, es del plan). `mutantes.md` registra su mutante, la
  lista literal devuelta a uno de ellos, con la orden y lo visto (FR-027). En `make ci`, cada test de la lista sigue en
  rojo con su mutante de `main` (US4.3).
- **SC-005**: La unión del repositorio tiene 4 raíces (H4, H5, H9 y h36), 0 de `evidencias/h9/`; con h99 en una copia, 5,
  h99 la última. `TestGrabacionesSinSolape` y `TestManifiestoDeGrabaciones` salen en rojo, en `make ci`, con h36 detrás
  de una raíz de número mayor o con un fichero repetido entre dos raíces.
- **SC-006**: La suma de las concurrencias del trabajo `evals` es como mucho 12, y el peor caso de cada trabajo cabe en
  su tope (352 min el de `plazos-y-territorio`), con las evals del repositorio y con cuatro más en una copia de
  `plazos-y-territorio`. `TestDefinicionDelJob` sale en rojo, en `make ci`, con `plazos-y-territorio` a 2 o con 13
  sesiones a la vez.
- **SC-007**: La `description` de `plazos-y-territorio` mide 429 caracteres (como mucho 450), las cuatro de la variante
  suman como mucho 1 700 y cada `description` del producto, como mucho 900. Lo hacen cumplir, en `make ci`,
  `TestDescripcionesCortas` y `TestDescripcionesDelProducto` (rojo por encima de un tope), y `TestSkillsDelProductoDesdeControl`
  y el guion `h41-4-skills-instaladas.txtar` (rojo con la descripción de hoy). En el cierre, el juez de las
  descripciones marca 0 (umbral 0): el job de evals sale en rojo si la marca.
- **SC-008**: El control de órdenes para PowerShell da 0 defectos con la forma de una orden del `SKILL.md` en una
  referencia a mano, y 1 defecto en cada uno de los 5 mutantes de §C.3. Lo hacen cumplir, en `make ci`,
  `TestEvalsDelRepositorio/ordenes-para-powershell` y `TestControlesSobreLasReferenciasAMano`.
- **SC-009**: `TestSkillsDelProductoDesdeControl` pasa con una `references/<vertical>.md` de una vertical de
  `data/normas.yaml` que su base no tiene, y sale en rojo, en `make ci`, en los 2 mutantes de US6.2 y US6.3.
- **SC-010**: En el cierre, ninguna clase que decide de ninguna skill queda por debajo de su umbral (ADR 0038 y 0045), y
  `plazos-y-territorio` se sigue activando ante sus preguntas de hoy (regla por serie del ADR 0016 en cada eval). Lo hace
  cumplir el job de evals del cierre (`.github/workflows/evals.yml`), que sale en rojo si una clase incumple su umbral o
  una eval su serie.
- **SC-011**: `make ci` en verde.

## Criterios del hito (literales)

**Aceptación** (del fichero del hito):

- la vertical de ensayo entra con un tipo y una fila con norma sin Go ni `SKILL.md` escrito a mano (US2; SC-002);
- con los datos de hoy, todas las referencias generadas salen como en `main`, y de las escritas a mano solo cambia
  `references/con_herramientas.md` de `normativa`, que gana lo movido (US1; SC-001);
- con una copia de `data/` y `evals/` con un paquete más, ningún test de la lista se pone en rojo (US4; SC-004);
- `plazos-y-territorio` se sigue activando ante sus preguntas de hoy con la descripción nueva (US9.3; SC-010).

**Controles** (del fichero del hito). Cada uno con la comprobación que lo pone en rojo, que se hace con un mutante o
contra `main` y se registra en `mutantes.md` (FR-027):

- **Lo de hoy, igual**: con los datos de `main`, `make skills-sync` no cambia ninguna referencia, y los cuerpos de
  `SKILL.md` solo cambian en el movimiento de `normativa` y en la descripción de `plazos-y-territorio`. Los casos de
  `TestCasosDePlazos` y los goldens de H9 y H36 dan lo mismo. En rojo: la sincronía con la regla de hoy y un tipo de
  ensayo, que lo escribe también en `administrativo.md` (FR-008, FR-026; US1; SC-001).
- **El molde**, en `TestVerticalDeEnsayo` crecida. En rojo: un tipo asignado a todas las referencias; una fila asignada
  por su precepto; `filasSinRegion` con la regla vieja y una referencia de ensayo sin la región, que deja de dar su
  defecto; la línea de `normativa` sin el movimiento (FR-002, FR-005, FR-006, FR-020; US2; SC-002).
- **Los defectos nuevos de la sincronía**: un tipo cuyos preceptos son de dos paquetes, y uno sin paquete, salen con su
  tipo. En rojo: quitar la regla (FR-003, FR-004; US3; SC-003).
- **Los tests sin listas**: con una copia de `data/` con un tipo, una fila y una norma con `vertical` de más, y una copia
  de `evals/plazos-y-territorio/` con una eval de más, los tests de la lista siguen en verde. En rojo: devolver a uno de
  ellos su lista literal (FR-016; US4; SC-004).
- **Las grabaciones**: `TestGrabacionesSinSolape` y `TestManifiestoDeGrabaciones` con las raíces sumadas solas,
  `evidencias/h36/` incluida y `evidencias/h9/` fuera; una carpeta `evidencias/h99/` de prueba en una copia entra sin
  Go. En rojo: h36 ordenado tras una raíz de número mayor; un fichero repetido entre dos raíces (FR-015; US7; SC-005).
- **El tope del trabajo de evals** (`TestDefinicionDelJob`): las cuatro skills caben, con doce sesiones a la vez como
  mucho, y con cuatro evals de más en una copia de `plazos-y-territorio`. En rojo: `plazos-y-territorio` a dos; trece
  sesiones a la vez (FR-021, FR-022; US8; SC-006).
- **La descripción**, con su juez de descripciones y bajo los topes de la variante corta (450 cada una, 1700 la suma).
  En rojo: la de hoy, que nombra la LPAC (FR-023; US9; SC-007).
- **Las órdenes para PowerShell**: una orden del `SKILL.md` con su forma en una referencia a mano no da defecto. En
  rojo: esa forma con otros bloques; esa forma en ninguna referencia; una orden de una referencia con su forma solo en
  el `SKILL.md`; una orden de una referencia con su forma en otra referencia; y la orden de articulos quitada, que sigue
  dando su defecto de verbo (FR-011, FR-012; US5; SC-008).
- **El laboratorio sin las referencias de un paquete**: con una `references/<vertical>.md` de una vertical de
  `data/normas.yaml` que su base no tiene, `TestSkillsDelProductoDesdeControl` sigue en verde. En rojo: la misma
  referencia con un nombre que no es una vertical; cambiar una referencia que la base ya tiene (FR-014; US6; SC-009).
- **Las evals**: el cierre mide todas las skills, porque el hito toca `internal/skills`, `internal/evals` y dos
  `SKILL.md`. Ninguna eval cambia. Si una clase cae, se compara con `main` y con la serie (ADR 0038), y la reparación no
  toca otro texto de skill que el movido; la descripción no se cambia (va al informe) (FR-025; SC-010).

## C · Los contratos que cambian

### C.1 Los defectos nuevos de la sincronía

Van, como los demás defectos de `data/plazos.yaml`, precedidos de `data/plazos.yaml: `, en cada skill con una
referencia que lleva la región `recursos_y_plazos`:

- sin paquete (FR-003): `el tipo <código> no es de ningún paquete: ninguna norma de sus preceptos lleva vertical en
  data/normas.yaml`
- de más de uno (FR-004): `el tipo <código> es de más de un paquete: <vertical>, <vertical>` con los paquetes en orden
  alfabético, separados por coma y espacio.

### C.2 Lo que gana `references/con_herramientas.md` de `normativa`

Al final del fichero, tras una línea en blanco, las líneas 38 a 41 de `skills/normativa/SKILL.md` de `main` sin sus
dos espacios iniciales:

````text
En PowerShell 5.1:
```powershell
kitlegal boe articulo BOE-A-1978-31229 a140 --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check BOE-A-1978-31229 a140 --json }
```
````

### C.3 Los casos del control de las órdenes para PowerShell

| Caso | Resultado |
|---|---|
| orden del `SKILL.md` con su forma en una referencia a mano de la skill, mismo verbo, norma y bloques | sin defecto |
| esa forma con otros bloques | defecto de la orden |
| esa forma en ninguna referencia | defecto de la orden |
| orden de una referencia con su forma solo en el `SKILL.md` | defecto de la orden |
| orden de una referencia con su forma en otra referencia | defecto de la orden |
| la orden de `boe articulos` quitada | su defecto de verbo, como hoy |

### C.4 Qué es `evidencias/<hito>/` y en qué orden va

Una carpeta de `evidencias/` es de un hito si su nombre es el que le da `scripts/workflow/grabar-datos.sh`: `h` y el
número del hito, y para un subhito, un guion y su número (`h36`, `h9-1`). Las raíces de esas carpetas van detrás de las
de H4, H5 y H9, por el número del hito y después por el del subhito, en orden numérico (h9-1 antes que h36, h36 antes
que h42). Las carpetas con otro nombre (`adr-*`, `sondeos-*`) no entran.

### C.5 Los tests de la lista

Del fichero del hito: `TestLeerCobertura/repositorio`, `TestCoberturaDelRepositorio`, `TestRegionDeLoCubierto/repositorio`
y `/sin-data-cobertura`, `TestRenderizarRecursosYPlazos/desde-data`, `TestRegionDeLaCobertura` (`internal/skills`);
`TestCasosDePlazos/19-eval-c-tributos` y los guiones `h32-paquetes.txtar`, `h33-cobertura-instalada.txtar` y
`h36-plazos-de-hoy.txtar` con sus copias congeladas (`internal/app`); `TestDefinicionDelJob/peor-caso` y
`/del-repositorio`, `TestPlanEnDosModos/del-repositorio`, `TestEvalsDelRepositorio/precepto-de-la-cobertura`,
`TestUmbralesDePlazosYTerritorio`, `TestUmbralesDeNormativa`, `TestUmbralesDeLasSkills`,
`TestInformeEnDosModos/plazos-y-territorio`, `TestMatrizDeLasPropuestas` y `TestSkillsDelProductoDesdeControl`
(`internal/evals`); y, del segundo prototipo, `TestEvalsDelRepositorio/ordenes-para-powershell`, la declaración de
`normativa` y la variante de descripciones del laboratorio de H41, los subtests con sumas literales de
`TestDescripcionesDelProducto` y «una que empieza por <ejemplo>» de `TestDescripcionesCortas`, y el guion
`h41-4-skills-instaladas.txtar` con su copia congelada. Más los que encuentre el plan con un prototipo entero (FR-016).

## Assumptions

- **El sitio de `TestVerticalDeEnsayo`** (FR-020): el hito pide que el mutante «la línea de `normativa` sin el
  movimiento» la ponga en rojo. Con el tope solo, no lo haría: sin el movimiento, `normativa` con la línea de ensayo mide
  10 997. Exigir el sitio de otra línea de paquete de la misma longitud (70 caracteres) es la única lectura que lo pone en
  rojo (11 067) y deja en verde a `jurisprudencia` (10 983), que el hito no toca; un margen de 100 la pondría en rojo
  (10 913 con ensayo). Es lo que dice el ADR 0062, punto 3: «`normativa` admite un paquete real más».
- **El rojo de la descripción de hoy** (FR-023, SC-007): lo da la comparación con el texto fijado (la variante, el
  `SKILL.md` y el guion `h41-4-skills-instaladas.txtar`). El hito no pide un control nuevo que rechace que una
  descripción nombre una norma o una materia; no se implementa (Fuera de alcance).
- **Qué es `<hito>` en `evidencias/`** (§C.4): el nombre que da `grabar-datos.sh`, la única que escribe esas carpetas;
  como dos raíces no comparten ficheros (FR-015), el orden no cambia lo que se reproduce, y lo fija el hito.
- **La copia de US4.1** es un paquete coherente: la norma con su `vertical`, el tipo con un precepto de esa norma, la fila
  con esa norma, y las referencias del paquete con sus regiones, para que la sincronía no dé defectos.
- **Antes de lanzar** (lo hace la persona, no el run, y no es requisito de este spec): fusionar la propuesta con esta
  entrada, la de H42 y el ADR 0062; la medida previa de las cuatro skills sobre `h45-prototipo` con su criterio fijado;
  `df -h /tmp` en delorean; y el sondeo del alias (ADR 0031).
- **Dependencias**: H36, hecho. H42 depende de este hito.

## Fuera de alcance

- Cualquier norma, tipo, fila, referencia o eval de una materia nueva: son de H42.
- El punto 6 del ADR 0062 (el cómputo de la revisión tributaria): de H42.
- Cambiar la forma de la región de los paquetes (una referencia con la lista, o una línea con todos): el ADR 0062 lo
  deja para cuando el siguiente paquete no quepa.
- Un campo `paquete:` en `data/plazos.yaml` o en `data/cobertura.yaml` (ADR 0062, opción 2).
- Las reglas del conjunto de `normativa`.
- Cualquier cambio en `kitlegal plazos`, en lo que calcula o en su esquema.
- Cualquier cambio de protocolo fuera del movimiento de `normativa` y de la descripción de `plazos-y-territorio`.
- Cambiar el ADR 0062 o escribir otro.
- Un defecto para un tipo cuyo paquete no tiene ninguna referencia con la región `recursos_y_plazos`: el hito solo
  declara defecto el tipo sin paquete y el de dos.
- Un control nuevo que rechace que la descripción de una skill de capacidad nombre una norma o una materia.
- Que el control de las órdenes para PowerShell interprete cuándo manda leer el cuerpo cada referencia, o que exija que
  esté enlazada desde el paso de la orden o leída sin condición: no se implementa; ningún mutante más allá de los cinco de
  §C.3 (Clarifications).
- Cambiar el texto de la descripción de FR-023 en el run, aunque una medida la señale: va al informe.
- La medida previa de las skills y lo demás de «Antes de lanzar»: los hace la persona.
