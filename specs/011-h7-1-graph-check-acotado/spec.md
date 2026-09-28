# Feature Specification: H7.1 · `graph check` acotado a la pregunta, con señales que se apagan y salida legible; H7 sin lo que no pasa el umbral (ADR 0028)

**Feature Branch**: `011-h7-1-graph-check-acotado`

**Created**: 2026-09-28

**Status**: Draft

**Input**: Sección «#### H7.1 · `graph check` acotado a la pregunta, con señales que se apagan y salida legible; H7 sin lo que no pasa el umbral (ADR 0028)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

H7 dio a `boe-legislacion` una memoria de consultas que no escala ni se calla. Con un uso sostenido medido en el binario —300 normas y 2 400 bloques consultados, el 90 % hace más de una semana— `kitlegal graph check --json` devuelve 4 812 hallazgos y 3 045 524 bytes, la skill lo pide dos veces por pregunta (≈ 6,1 MB), y lo que atañe a la norma de la pregunta son unos 20 hallazgos (≈ 13 KB). Y ninguna señal se apaga: `version-obsoleta` se da para siempre sobre toda redacción superada, y `fuente-caducada` también sobre redacciones superadas que ninguna lectura puede renovar, así que la skill repetiría «la redacción ha cambiado» en cada pregunta sobre ese artículo (bitácora `docs/USO.md`, entrada del 2026-09-28).

H7.1 no entrega skill nueva: arregla lo que H7 dio a la que existe (constitución, principio VIII), para que `boe-legislacion` diga que la redacción de un artículo ha cambiado **en la respuesta en la que cambió, y solo en esa**, con una comprobación que cuesta lo que la pregunta. Son siete piezas:

1. **`graph check` acotado** a una norma y, si se quiere, a bloques de esa norma, con **50 hallazgos como mucho** en cualquier forma, también sin argumentos, y un recuento de lo que hay y de lo omitido.
2. **`version-obsoleta` que se apaga**: dice que la redacción de un bloque cambió respecto de su lectura anterior y deja de darse con la lectura siguiente.
3. **`fuente-caducada` solo sobre lo vigente**: nunca sobre una redacción superada.
4. **`boe-legislacion` v0.1.1**: una comprobación por norma citada, después de leer, que traslada cada cambio de redacción con la forma fija `⚠ REDACCIÓN MODIFICADA:` y no traslada `fuente-caducada`.
5. **La eval de la consulta repetida** (la 19) comprueba esa forma sin juicio de ningún modelo.
6. **Salida legible** de `graph stats`, `show` y `check` sin `--json` (ADR 0026).
7. **H7 sin lo que no pasa el umbral de materialidad** (constitución 2.6.0, «Gates»; ADR 0028): se retiran los mecanismos y los tests de estados a los que el binario no llega, cubiertos por la regla genérica (defecto `inesperado`, código 1, ADR 0023).

La línea de H7 sigue gobernando: **el grafo dice qué hay que volver a comprobar, nunca qué dice el artículo** (ADR 0014).

**Cómo se citan los requisitos de H7.** Los requisitos, historias y criterios del spec de H7 (`specs/010-h7-internal-graph-grafo/spec.md`) se citan como «H7 FR 060», «H7 SC 011» o «H7 US2.3», sin guion, para no confundirlos con los de este spec. Los ids con guion (FR-001, SC-001…) son siempre de este spec.

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H7.1, de lo que este spec tiene que cumplir tal cual. El Objetivo y el Alcance, que citan requisitos de H7 por su id, se reflejan en «Relación con H7» y en cada requisito.

- **Entrega**: «(1) `kitlegal graph check` acotado a una norma y, si se quiere, a bloques de esa norma, con una salida de 50 hallazgos como mucho en cualquier forma, también sin argumentos; (2) `version-obsoleta` que dice que la redacción cambió respecto de la lectura anterior de ese bloque y se apaga con la siguiente; (3) `fuente-caducada` solo sobre lo vigente; (4) `boe-legislacion` v0.1.1: una comprobación por norma citada, después de leer, que traslada el cambio de redacción con una forma fija y no traslada `fuente-caducada`; (5) la eval de la consulta repetida comprueba esa forma sin modelo; (6) salida legible de `graph stats`, `show` y `check` sin `--json`; (7) H7 sin los mecanismos y los tests de estados a los que el binario no llega, cubiertos por la regla genérica.»
- **Controles**: «`make ci` en verde, con `schema-check` y `skills-check` sin drift; e2e (la suite congelada de H7.1) con relojes fijados y redacciones derivadas de la grabación de H4, como las de H7: la secuencia de cinco lecturas de `version-obsoleta` paso a paso, `fuente-caducada` tras el paso 5 sobre la `Norma`, el `Bloque` y C y no sobre A ni B, el acotado por norma y por norma y bloque en un grafo con dos normas, una norma desconocida (0 y ningún hallazgo) y un identificador sin forma BOE (2), la salida legible de `stats`, `show` y `check` —con hallazgos y sin ellos— y la de `--json` sin cambios, y el fichero que no es una base SQLite como vehículo de la regla genérica en los verbos de `graph` y en la entrega; un test de integración que siembra con el propio almacén el grafo de la medida (300 normas, 2 400 bloques, 240 con dos redacciones, el 90 % consultado hace más de una semana) y comprueba que `graph check --json` sin argumentos da 50 hallazgos, los `version-obsoleta` primero, con el recuento por clase y lo omitido correctos, en ≤ 40 000 bytes, y que con la norma y un bloque da solo los de ese bloque y su norma; una comprobación mecánica en `make ci` de que `SKILL.md` lleva la etiqueta que da el binario para `version-obsoleta`, como la de los avisos de H5.1; tests de `Juzgar` con la forma encontrada, ausente y dicha con otras palabras (ausente); los ficheros retirados ya no existen y ningún test ni rama de código trata enlaces simbólicos, diarios de rollback, `-shm` sueltos, permisos cambiados ni bases de fuera; la cobertura, dentro de los umbrales de `codecov.yml`; `CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio.»
- **Aceptación**: «sobre el grafo sembrado de la medida, `graph check --json` sin argumentos pasa de 3 045 524 bytes a ≤ 40 000, y con la norma y el bloque de una pregunta lleva solo los hallazgos de esa norma y ese bloque; la secuencia de cinco lecturas da exactamente lo que dice el Alcance; en el informe del job de evals, la eval 19 declara la forma `⚠ REDACCIÓN MODIFICADA:` y su tasa, con el veredicto aprobado (el hito cambia `SKILL.md`, Definition of Done §1.10) y `red` vacío; y en Claude Code, la misma pregunta sobre un artículo cuya redacción acaba de cambiar, hecha dos veces seguidas, lleva `⚠ REDACCIÓN MODIFICADA:` en la primera respuesta y no en la segunda.»
- **Fuera de alcance**: «que `boe-legislacion` traslade `fuente-caducada`, y usarla en H8 y H10, que es de esos hitos; promover la eval 19 a decisoria; una eval que exija que la respuesta a la pregunta siguiente no lleve la forma, porque el apagado se prueba en el binario y la skill solo traslada lo que `graph check` da; que `version-obsoleta` espere a que una respuesta la diga: la apaga cualquier lectura del bloque, también la de una persona con `kitlegal boe articulo` o la de otra skill, y entonces la respuesta siguiente de `boe-legislacion` sobre ese artículo no avisa del cambio (es la consecuencia aceptada de que la señal dependa solo de las lecturas y no de un estado de «ya dicho»); `graph history`, el diff entre redacciones y cualquier clase o regla nueva de `check` (backlog · profundidad, H10); paginar `check` o elegir la cota con una bandera; `boe articulo --fecha` (H20); salida legible de applets distintos de `graph`; que `legal-core` use `graph`; retirar nada de la lista «Se queda»; y reescribir `internal/graph`, o editar el spec, el plan o los guiones de H7, fuera de lo que dice el Alcance.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega (1), la cota y su cálculo | FR-001 a FR-015, US3, SC-001, SC-002 |
| Entrega (2), la secuencia de cinco lecturas y el `world.db` de H7 | FR-020 a FR-027, US2, SC-003, SC-012 |
| Entrega (3) | FR-030 a FR-032, US4, SC-004 |
| Entrega (4) | FR-040 a FR-048, US1, SC-005 |
| Entrega (5) | FR-050 a FR-055, US1, SC-006, SC-008 |
| Entrega (6) | FR-060 a FR-064, US5, SC-010 |
| Entrega (7), retirada y «Se queda» | FR-070 a FR-078, US6, SC-011, SC-013 |
| Guiones de H7 y esquemas | FR-080 a FR-083 |
| Controles | FR-090 a FR-098, SC-001 a SC-014 |
| Aceptación | SC-001, SC-002, SC-003, SC-006, SC-007 |

## Relación con H7

`specs/010-h7-internal-graph-grafo/` no se edita: es el registro de su run, como el de H5 tras H5.1. No cambia ninguna decisión de arquitectura —`graph check` sigue con las dos clases del ADR 0014, y los ADR 0023 y 0026 se aplican tal cual—, así que no hay ADR nuevo.

**Sustituye** (lo que dice H7 deja de valer y vale lo de este spec):

- H7 FR 055 (sin `--json`, tabla mínima o texto legible): FR-060 a FR-064.
- H7 FR 060 (`check` sin argumentos sobre el grafo entero): FR-001 a FR-015.
- H7 FR 062, en el orden entre clases (por clase comparando bytes, que ponía `fuente-caducada` delante): FR-011; el orden por id dentro de cada clase se mantiene.
- H7 FR 063 (`version-obsoleta` sobre toda versión con otra posterior): FR-020 a FR-027.
- H7 FR 066 (`fuente-caducada` sobre todo nodo caducado): FR-030 a FR-032.
- H7 FR 080 (dos comprobaciones por pregunta y trasladar las dos clases): FR-040 a FR-048.
- H7 FR 085, en lo que exige de la respuesta (no comprobar el traslado): FR-050 a FR-055.
- Las clarificaciones de H7 Q2 (la eval no comprueba el traslado) y Q5 (`fuente-caducada` también sobre versiones superadas); y, con H7 FR 080, la Q4 (dos comprobaciones por pregunta), que es su consecuencia.

**Retira o reduce**: H7 FR 010 (enumeración de inutilizables), H7 FR 023 (niveles del desempate), H7 FR 024 (rechazos que ningún emisor produce), H7 FR 088 (la matriz de sus tests), H7 US2.3 (orden inverso), H7 US3.5 y H7 US6.4 (inutilizables en la entrega y en los verbos) y H7 SC 011 (su matriz), en FR-070 a FR-077. Y, porque el hito retira los casos no ASCII del patrón de `Persona`, los ejemplos no ASCII de H7 FR 025, de H7 SC 010 y de su clarificación Q6 dejan de exigirse (FR-074).

Todo lo demás de H7 sigue en vigor.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - `boe-legislacion` dice que la redacción cambió en la respuesta en que cambió, y solo en esa (Priority: P1)

Una persona pregunta por un artículo que ya se leyó cuando su redacción era otra. La skill lee el bloque una vez con `kitlegal boe articulo`, después pide `kitlegal graph check` con esa norma y el bloque leído, recibe un `version-obsoleta` y responde citando el texto que acaba de leer, con una línea `⚠ REDACCIÓN MODIFICADA:` y las dos fechas de vigencia. Si vuelve a preguntar lo mismo, la skill vuelve a leer, `graph check` ya no da nada y la respuesta no lleva la forma.

**Why this priority**: es lo que el hito dice que gana la skill y la condición de aceptación.

**Independent Test**: eval 19 de `evals/boe-legislacion/` en el job de evals; y, fuera del run, la misma pregunta dos veces en Claude Code (SC-007).

**Acceptance Scenarios**:

1. **Dado** una sesión cuyo grafo ya registró una lectura del art. 21 LPAC con la redacción de fecha de vigencia 20151002 y cuya caché sirve la de 20161002, **Cuando** se pregunta qué dice ahora ese artículo, **Entonces** la sesión ejecuta `kitlegal boe articulo` de `BOE-A-2015-10565` `a21`, ejecuta `kitlegal graph check` con la norma `BOE-A-2015-10565`, no ejecuta `kitlegal graph show`, y la respuesta lleva la cita `[BOE-A-2015-10565, bloque a21]` y la forma `⚠ REDACCIÓN MODIFICADA:` (FR-040, FR-042, FR-050).
2. **Dado** una respuesta que dice «la redacción ha cambiado» sin la forma fija, **Cuando** se juzga la eval 19, **Entonces** la forma queda ausente y la sesión no pasa (FR-052).
3. **Dado** que la skill ha leído k bloques de una norma y ninguno cambió de redacción, **Cuando** pide `graph check` con esa norma y esos bloques, **Entonces** sale con 0 y ningún hallazgo, y la respuesta no dice nada de la memoria de consultas (FR-043, FR-044).
4. **Dado** una pregunta cuya respuesta cita bloques de dos normas (una por remisión), **Cuando** la skill termina de leer, **Entonces** pide `graph check` dos veces, una con cada norma y sus bloques leídos, y ninguna sin argumentos (FR-040).
5. **Dado** un `graph check` que sale con un código distinto de 0, **Cuando** la skill responde, **Entonces** responde con el texto de `kitlegal boe` y dice que no se ha podido comprobar la memoria de consultas (H7 FR 082, que se queda; FR-046).

---

### User Story 2 - `version-obsoleta` se da una vez, con la lectura que ve el cambio, y se apaga con la siguiente (Priority: P1)

Cada vez que `boe articulo` o `boe articulos` devuelve un bloque y entrega su resultado al grafo, el grafo cuenta una lectura de ese bloque y qué redacción vio. `graph check` dice que la redacción cambió cuando la última lectura vio una redacción posterior a la que vio la anterior, y deja de decirlo en cuanto otra lectura ve la nueva.

**Why this priority**: es la señal que la skill traslada; sin apagado, la skill repite el aviso en cada pregunta.

**Independent Test**: e2e con relojes fijados y tres redacciones A, B y C del mismo bloque derivadas de la grabación de H4, con fechas de vigencia crecientes; `graph check --json` tras cada paso.

**Acceptance Scenarios** (la secuencia del hito, paso a paso; FR-025):

1. **Dado** un directorio de caché vacío, **Cuando** se lee el bloque y la fuente sirve A, **Entonces** `graph check` sale con 0 y no da ningún hallazgo.
2. **Dado** el estado del escenario 1, **Cuando**, caducada la caché, se lee de nuevo y la fuente sirve B, **Entonces** `graph check` da exactamente un `version-obsoleta`, sobre A, con la fecha de vigencia de A y la de B; y repetir `graph check` sin leer da el mismo hallazgo.
3. **Dado** el estado del escenario 2, **Cuando** se lee otra vez con `--no-graph`, **Entonces** `graph check` sigue dando el mismo hallazgo.
4. **Dado** el estado del escenario 3, **Cuando** se lee otra vez y la caché sirve B, **Entonces** `graph check` ya no da ningún `version-obsoleta`, y el `graph check` siguiente tampoco.
5. **Dado** el estado del escenario 4, **Cuando**, caducada la caché, la fuente sirve C, **Entonces** `graph check` da exactamente un `version-obsoleta`, sobre B, con la fecha de vigencia de B y la de C.
6. **Dado** un `world.db` escrito por H7 en el que un bloque tiene registradas dos redacciones, **Cuando** se ejecuta `graph check`, **Entonces** no da ningún `version-obsoleta` para ese bloque, y `graph show` y `graph stats` dan lo mismo que daban con H7 (FR-026).

---

### User Story 3 - `graph check` cuesta lo que la pregunta, y nunca más de 50 hallazgos (Priority: P1)

La skill, o una persona, pide `graph check` con una norma y, si quiere, algunos de sus bloques, y recibe solo lo que atañe a esa norma y esos bloques. Sin argumentos, una persona repasa todo lo consultado y recibe como mucho 50 hallazgos, los `version-obsoleta` primero, con cuántos hay de cada clase y cuántos se han omitido.

**Why this priority**: es el defecto medido (≈ 6,1 MB por pregunta) y el primer criterio de aceptación.

**Independent Test**: e2e en un grafo con dos normas; y test de integración sobre el grafo sembrado de la medida.

**Acceptance Scenarios**:

1. **Dado** un grafo con hallazgos en dos normas, **Cuando** se ejecuta `graph check --json` con la primera norma, **Entonces** sale con 0 y todos los hallazgos son de la `Norma` de esa norma, de sus `Bloque` o de sus `BloqueVersion`, y están todos los de esa norma (FR-002).
2. **Dado** el mismo grafo, **Cuando** se ejecuta con la primera norma y uno de sus bloques, **Entonces** los hallazgos son solo los de esa `Norma`, ese `Bloque` y sus `BloqueVersion` (FR-002).
3. **Dado** el mismo grafo, **Cuando** se ejecuta con una norma con forma BOE que el grafo no conoce, o con una norma conocida y un bloque que el grafo no conoce, **Entonces** sale con 0, `ok` es `true` y no da ningún hallazgo de ese bloque ni de otra norma (FR-003).
4. **Dado** cualquier estado, **Cuando** se ejecuta con un identificador sin la forma `BOE-A-<año>-<número>` o con un bloque vacío o de solo espacio en blanco, **Entonces** sale con 2 y la clase `argumentos` (FR-004).
5. **Dado** el grafo sembrado de la medida (300 normas, 2 400 bloques, 240 con dos redacciones, el 90 % consultado hace más de una semana), **Cuando** se ejecuta `graph check --json` sin argumentos, **Entonces** da 50 hallazgos, todos `version-obsoleta`, `data` dice el total de cada clase y cuántos se han omitido, y la salida ocupa ≤ 40 000 bytes (FR-010 a FR-013, SC-001).
6. **Dado** el mismo grafo, **Cuando** se ejecuta con la norma y un bloque, **Entonces** da solo los hallazgos de ese bloque, de sus redacciones y de su norma (SC-002).

---

### User Story 4 - `fuente-caducada` solo sobre lo vigente (Priority: P2)

Quien repasa la memoria entera ve qué consultas han caducado, pero solo de lo que una lectura nueva puede renovar: la `Norma`, el `Bloque` y la redacción que vio la última lectura del bloque. Nunca sobre una redacción superada.

**Why this priority**: deja de dar una señal que ninguna lectura apaga; la skill no la traslada, así que no cambia ninguna respuesta.

**Independent Test**: e2e tras el paso 5 de la secuencia, sin leer y con el reloj pasada la vigencia.

**Acceptance Scenarios**:

1. **Dado** el estado del paso 5 de US2, **Cuando** se ejecuta `graph check --json` sin leer y con el reloj pasada la vigencia de la lectura de C, **Entonces** da `fuente-caducada` sobre la `Norma`, el `Bloque` y C, ninguno sobre A ni sobre B, y además el `version-obsoleta` sobre B (FR-032).
2. **Dado** ese estado, **Cuando** se lee de nuevo el bloque y la fuente sirve C, **Entonces** `graph check` ya no da `fuente-caducada` sobre ninguno de los tres ni `version-obsoleta` sobre B (FR-031).

---

### User Story 5 - Una persona lee `graph` sin `--json` (Priority: P2)

Una persona que depura o repasa su memoria ejecuta `kitlegal graph stats`, `kitlegal graph show <id>` o `kitlegal graph check` sin `--json` y lee un texto pensado para ella, no la tabla mínima del kernel.

**Why this priority**: `graph` es el único applet que una persona consulta a mano para saber qué recuerda el kit; la tabla mínima no le dice qué hacer.

**Independent Test**: e2e de los tres verbos sin `--json`, con hallazgos y sin ellos, comparados con su salida `--json`.

**Acceptance Scenarios**:

1. **Dado** un grafo poblado, **Cuando** se ejecuta `graph stats` sin `--json`, **Entonces** la salida no es la tabla mínima y dice cuántos nodos, aristas y textos hay, y cada recuento por tipo y fuente y por relación y fuente de su `data` (FR-061).
2. **Dado** el mismo grafo, **Cuando** se ejecuta `graph show` de un `Bloque`, **Entonces** la salida dice su tipo y sus datos, su primera y su última observación, la fuente, la url y la fecha de consulta de la última, y cada arista con su relación, el otro extremo y su procedencia, sin ninguna línea del texto del bloque (FR-062).
3. **Dado** un grafo con hallazgos, **Cuando** se ejecuta `graph check` sin `--json` y sin argumentos, **Entonces** la salida da cada hallazgo listado con su explicación, agrupados por clase, con el total de cada clase y lo omitido, y termina diciendo cómo acotarla a una norma (FR-063).
4. **Dado** un grafo sin hallazgos para la norma y el bloque pedidos, **Cuando** se ejecuta `graph check` sin `--json` con ellos, **Entonces** la salida es una frase que dice que no hay nada que volver a comprobar y nombra esa norma y ese bloque (FR-063).
5. **Dado** cualquiera de esos estados, **Cuando** se ejecutan los mismos verbos con `--json`, **Entonces** la salida es la que fijan FR-012 y H7 FR 053 y FR 054: la salida legible no cambia nada de la de `--json` (FR-060).

---

### User Story 6 - H7 sin lo que no pasa el umbral, con la regla genérica (Priority: P3)

Lo que llega de fuera manipulado a mano deja de tener comportamiento propio: un `world.db` que el binario no puede usar es un defecto `inesperado`, código 1, en los verbos de `graph`, y en la entrega un aviso de una línea sin cambiar la salida del applet. Se retiran los mecanismos y tests que solo servían a esos estados, y se quedan los que tienen vía real.

**Why this priority**: no cambia nada que vea una persona o una skill en un uso real; reduce el código y los tests que hay que mantener.

**Independent Test**: e2e con un `world.db` que no es una base SQLite; comprobación de que los ficheros retirados no existen.

**Acceptance Scenarios**:

1. **Dado** un `world.db` que no es una base SQLite, **Cuando** se ejecuta `graph stats`, `graph show <id>` o `graph check`, **Entonces** sale con 1, clase `inesperado`, y el mensaje nombra la ruta de `world.db` (FR-070).
2. **Dado** ese mismo `world.db`, **Cuando** se ejecuta `boe articulo` sin `--no-graph`, **Entonces** sale con 0, la salida estándar es la misma que con `--no-graph` y la salida de error lleva una línea que nombra `world.db` (FR-070).
3. **Dado** el árbol del repositorio tras el hito, **Cuando** se buscan los ficheros retirados y el tratamiento de enlaces simbólicos, diarios de rollback, `-shm` sueltos, permisos cambiados y bases de fuera, **Entonces** los ficheros no existen y ningún test ni rama de código trata esos estados (FR-071 a FR-073, FR-095).

---

### Edge Cases

- **Norma sin ELI** (la respuesta de `boe articulo` no trae un `url_eli` con el segmento `eli/`): no emite (H7 FR 040), así que no hay lectura ni hallazgos; `graph check` con esa norma sale con 0 y ninguno (FR-003, FR-020).
- **El mismo bloque dos veces en una invocación de `boe articulos`**: cuenta como una sola lectura (FR-020).
- **Una lectura que no llega al grafo** (código distinto de 0, `--no-graph`, `--dry-run` o una entrega que falla, H7 FR 033): no es lectura y no apaga nada (FR-020).
- **Dos lecturas que ven redacciones con la misma fecha de vigencia y distinta huella**, o una redacción sin fecha de vigencia válida (H7 FR 064): no dan `version-obsoleta` (FR-023).
- **Dos invocaciones concurrentes que leen el mismo bloque**: sus lecturas se ordenan por el orden en que se aplican sus entregas (FR-021).
- **Más de 50 hallazgos en el ámbito pedido**: se listan los 50 primeros del orden de FR-011 y `data` dice cuántos se omiten (FR-010 a FR-012).
- **Un bloque nombrado que el grafo no conoce junto a otro que sí**: solo salen los del conocido y los de su norma (FR-003).
- **`world.db` ausente o sin esquema** (H7 FR 004): `graph check`, con argumentos o sin ellos, sale con 0 y ningún hallazgo.
- **`world.db` escrito por H7**: cada bloque cuenta con una lectura; ningún `version-obsoleta` hasta que una lectura nueva vea otra redacción (FR-026).
- **Cualquier otro `world.db` que el binario no puede usar**: regla genérica, 1 e `inesperado` en los verbos de `graph`, aviso de una línea en la entrega (FR-070).

## Requirements *(mandatory)*

### Functional Requirements

#### `graph check`: ámbito y argumentos

- **FR-001**: `graph check` MUST admitir dos formas: sin argumentos, o con un identificador de norma y, opcionalmente, uno o varios ids de bloque de esa norma. Cómo se pasan los argumentos lo fija el plan. Comprobable: con la norma `BOE-A-2015-10565`, y con esa norma y el bloque `a21`, sale con 0.
- **FR-002**: Con una norma, `graph check` MUST devolver solo los hallazgos de la `Norma` cuyo identificador BOE es esa norma, de los `Bloque` nombrados de esa norma —o de todos los suyos, si no se nombra ninguno— y de las `BloqueVersion` de esos bloques, y MUST devolver todos los de ese ámbito que dan las reglas de FR-023 y FR-030 (con la cota de FR-010). Comprobable: en un grafo con hallazgos en dos normas, ningún hallazgo de la otra norma sale con la primera, con bloques o sin ellos.
- **FR-003**: Una norma que el grafo no conoce, o un bloque de esa norma que el grafo no conoce, MUST NOT ser un error: sale con 0, `ok` es `true` y no aporta ningún hallazgo; los bloques conocidos que se nombren junto a él dan los suyos. Un `world.db` ausente o sin esquema se lee como un grafo vacío (H7 FR 004), con argumentos o sin ellos.
- **FR-004**: Un identificador de norma sin la forma `BOE-A-<año>-<número>` —la misma que acepta `kitlegal boe articulo`: `BOE-A-`, cuatro cifras ASCII, un guion y de una a nueve cifras ASCII—, o un id de bloque vacío o formado solo por caracteres de espacio en blanco (la propiedad `White_Space` de Unicode, como en H7 FR 052), MUST hacer salir a `graph check` con el código 2 y la clase `argumentos`, con un mensaje que nombra el valor. Comprobable: con `a21` como norma, con `BOE-B-2015-10565`, o con la norma `BOE-A-2015-10565` y un bloque ` `, sale con 2.
- **FR-005**: Sin argumentos, `graph check` MUST comprobar todo lo consultado, con la misma cota (FR-010) y la misma forma de `data` (FR-012).
- **FR-006**: `graph check` MUST salir con 0 con hallazgos o sin ellos, con `ok: true` (ADR 0023); solo sale con otro código si no ha podido comprobar: 2 por argumentos (FR-004 y H7 FR 011), 1 por un `world.db` que no puede usar (FR-070) o por un esquema posterior (H7 FR 012), y 4 o 1 por la espera de H7 FR 014.
- **FR-007**: La descripción de `graph check` —en `--describe`, en la ayuda y en la tabla de comandos generada de `SKILL.md`— MUST decir que se acota a una norma y a bloques de esa norma y que lista 50 hallazgos como mucho.

#### `graph check`: la cota y la forma de `data`

- **FR-010**: Ninguna invocación de `graph check` MUST listar más de 50 hallazgos, en ninguna de sus formas.
- **FR-011**: Los hallazgos del ámbito MUST ordenarse con todos los `version-obsoleta` antes que todos los `fuente-caducada` y, dentro de cada clase, por id del nodo comparando bytes (como H7 FR 062); si hay más de 50, se listan los 50 primeros de ese orden. Dos ejecuciones sobre el mismo grafo, con los mismos argumentos y el mismo instante de comprobación, dan la misma lista.
- **FR-012**: `data` de `graph check` MUST llevar: la lista de hallazgos listados (vacía, nunca nula), cada uno con lo que exige H7 FR 061; el total de hallazgos de cada una de las dos clases en el ámbito, contando los omitidos; el número de hallazgos omitidos, igual a la suma de los dos totales menos los listados; y el ámbito pedido (la norma y los bloques nombrados, o que es todo lo consultado). Su forma cambia respecto de H7 —que daba solo la lista—, porque H7 no ha salido en ninguna release y nada fuera del repositorio depende de ella; `schemas/grafo.json` la describe y `make schema-check` la vigila.
- **FR-013**: Con `--json`, la salida estándar de `graph check` sin argumentos o solo con la norma MUST ocupar ≤ 40 000 bytes sobre el grafo sembrado de la medida (FR-092), frente a los 3 045 524 bytes de H7. Cálculo (criterio de uso, ADR 0028): en la medida, un hallazgo pesa 633 bytes de media (3 045 524 entre 4 812) y un `version-obsoleta`, que lleva dos fechas, unos 700 como mucho; el sobre con el recuento y el ámbito, unos 300; 50 × 700 + 300 ≈ 35 KB, con cualquier número de normas y bloques consultados.
- **FR-014**: `graph check` MUST NOT modificar el grafo ni las lecturas: repetir `graph check` sin leer da el mismo resultado (H7 FR 005).
- **FR-015**: Paginar la lista, elegir la cota o acotar por otra cosa que la norma y sus bloques MUST NOT ofrecerse (Fuera de alcance).

#### Lecturas y `version-obsoleta`

- **FR-020**: Una **lectura** de un bloque MUST ser cada invocación de `kitlegal boe articulo` o `kitlegal boe articulos` que devuelve ese bloque y entrega su resultado al grafo —la sirva la fuente o la caché, con `--offline` o sin él—, una por bloque e invocación aunque la invocación nombre el bloque más de una vez. Con `--no-graph` o `--dry-run`, con un código distinto de 0, con una respuesta que no emite (H7 FR 040) o con una entrega que falla (H7 FR 033), no hay lectura.
- **FR-021**: Cada lectura MUST registrar qué redacción (`BloqueVersion`) vio. Las lecturas de un bloque se ordenan por el orden en que se aplican sus entregas al grafo, que se aplican una tras otra también cuando dos invocaciones concurren (H7 FR 014); «la última lectura» y «la lectura anterior» de un bloque se entienden en ese orden.
- **FR-022**: Las lecturas MUST NOT cambiar lo que devuelven `graph show` ni `graph stats`: repetir una consulta servida por la caché deja sus salidas byte a byte como estaban (H7 US1.3 y H7 SC 001). Comprobable: `internal/app/testdata/script/h7-grafo-memoria.txtar` pasa sin cambios.
- **FR-023**: `version-obsoleta` MUST darse sobre un bloque si y solo si su última lectura vio una redacción cuya fecha de vigencia es estrictamente posterior a la de la redacción que vio la lectura anterior, con las dos fechas válidas según H7 FR 064. Se da una vez, sobre la redacción que vio la lectura anterior (su id es el de esa `BloqueVersion`), con la fecha de vigencia de esa redacción y la de la que vio la última lectura como valores propios, y con la procedencia y la explicación que fija H7 FR 063 tomando como versión más reciente la que vio la última lectura. En cualquier otro caso —una sola lectura, la misma redacción, la misma fecha de vigencia con otra huella o una fecha no válida— no se da.
- **FR-024**: `version-obsoleta` MUST dejar de darse en cuanto otra lectura del bloque ve una redacción que no es estrictamente posterior a la que vio la lectura anterior a ella; en particular, con la lectura siguiente que ve la redacción nueva. Cualquier lectura apaga la señal, la haga la skill, una persona u otra skill.
- **FR-025**: Con A, B y C tres redacciones del mismo bloque con fechas de vigencia crecientes, la secuencia MUST dar exactamente esto: (1) se lee el bloque y la fuente sirve A: `graph check` no da nada; (2) caducada la caché, se lee de nuevo y la fuente sirve B: da un `version-obsoleta` sobre A, con las fechas de A y de B, y repetir `graph check` sin leer da el mismo; (3) se lee otra vez con `--no-graph`: sigue dando el mismo; (4) se lee otra vez y la caché sirve B: ya no da nada, y el `graph check` siguiente tampoco; (5) caducada la caché, la fuente sirve C: un `version-obsoleta` sobre B, con las fechas de B y de C.
- **FR-026**: En un `world.db` escrito por H7, cada bloque MUST contar con una lectura, la de la redacción suya que se observó por última vez: no se pierde nada de lo registrado —`graph show` y `graph stats` dan lo mismo que antes— y ningún bloque da `version-obsoleta` hasta que una lectura nueva vea otra redacción. Comprobable: sobre un `world.db` de H7 con un bloque con dos redacciones, `graph check` no da `version-obsoleta`; tras una lectura que ve una redacción posterior a la última observada, da uno sobre esta.
- **FR-027**: El orden inverso de H7 US2.3 se retira, sin comportamiento ni test propios: H7 US2.3 y su sección de `h7-grafo-version-obsoleta.txtar` salen con FR-080, y pedir una redacción anterior a propósito es `--fecha`, de H20.

#### `fuente-caducada`, solo sobre lo vigente

- **FR-030**: `fuente-caducada` MUST darse, como en H7 FR 066 (condición de caducidad, valores propios y explicación), solo sobre la `Norma`, sobre el `Bloque` y sobre la redacción que vio la última lectura de ese bloque, y MUST NOT darse nunca sobre una `BloqueVersion` que no es la que vio la última lectura de su bloque. Los nodos sin vigencia declarada siguen sin darla (H7 FR 066).
- **FR-031**: `fuente-caducada` sobre un nodo MUST dejar de darse con la siguiente lectura que renueva su observación: pasada la vigencia, la caché no sirve la consulta (H4) y con `--offline` sin entrada vigente la invocación sale con 4 sin lectura (H3), así que toda lectura posterior a la caducidad pregunta a la fuente. La `Norma` la renueva la lectura de cualquiera de sus bloques.
- **FR-032**: Tras el paso 5 de FR-025, sin leer y pasada la vigencia de la lectura de C, `graph check` MUST dar `fuente-caducada` sobre la `Norma`, el `Bloque` y C, y ninguno sobre A ni sobre B, además del `version-obsoleta` sobre B.

#### Skill `boe-legislacion` v0.1.1

- **FR-040**: `skills/boe-legislacion/SKILL.md` MUST pedir `kitlegal graph check --json` una vez por cada norma cuyos bloques cita la respuesta —una sola en la mayoría de las preguntas; una más por cada norma a la que lleve una remisión—, después de leer y antes de responder, con esa norma y los bloques de ella leídos. MUST NOT pedirlo antes de leer ni sin argumentos.
- **FR-041**: `SKILL.md` MUST pedir que cada bloque se lea una sola vez por pregunta, porque una segunda lectura apagaría la señal antes de comprobarla (FR-024).
- **FR-042**: `SKILL.md` MUST pedir que cada `version-obsoleta` se traslade con una forma fija: `⚠`, seguido de la etiqueta `REDACCIÓN MODIFICADA` y dos puntos, en la misma línea, con las dos fechas de vigencia detrás (la de la redacción superada y la de la leída), y MUST escribir esa forma literal (`⚠ REDACCIÓN MODIFICADA:`). Decirlo con otras palabras no traslada el hallazgo.
- **FR-043**: `SKILL.md` MUST pedir que `fuente-caducada` no se traslade: la respuesta cita siempre el texto que la skill acaba de leer, que la caché no sirve pasada su vigencia, y tras leer la señal no puede darse sobre lo leído (FR-030, FR-031).
- **FR-044**: Con `graph check` terminado con 0 y sin `version-obsoleta`, la respuesta MUST NOT decir nada de la memoria de consultas; con otro código se aplica FR-046 (H7 FR 082).
- **FR-045**: La etiqueta `REDACCIÓN MODIFICADA` MUST darla el binario como única fuente de verdad, igual que las de los avisos de H5.1, y MUST NOT coincidir con ninguna etiqueta de aviso (`NORMA DEROGADA`, `VIGENCIA AGOTADA`, `TEXTO POSIBLEMENTE DESACTUALIZADO`). Una comprobación mecánica en `make ci` MUST fallar, nombrando `version-obsoleta`, si `SKILL.md` no lleva la forma completa —`⚠`, esa etiqueta y dos puntos—, reconocida con la misma función y las mismas tolerancias que `Juzgar` (FR-051).
- **FR-046**: Se mantienen en `SKILL.md` H7 FR 081 (el texto citado sale de `kitlegal boe`, nunca de `graph`) y H7 FR 082 (código 0 es un resultado; con otro código, responder con el texto de `kitlegal boe` y decir que no se pudo comprobar la memoria), la forma de la cita y la de los avisos de vigencia; `SKILL.md` MUST seguir por debajo de 300 líneas, con frontmatter válido, sin nombrar evals, el job ni modelos, y con la tabla de comandos generada sin drift (`make skills-check`).
- **FR-047**: `SKILL.md` MUST NOT cambiar más allá de lo que piden FR-040 a FR-046.
- **FR-048**: `CHANGELOG.md` (*Unreleased*) MUST registrar `boe-legislacion` v0.1.1 con lo que cambia para quien la usa.

#### Eval 19 y formato común de eval

- **FR-050**: La eval 19 de `evals/boe-legislacion/` —el grafo ya registró una redacción anterior del art. 21 LPAC y la caché sirve la vigente— MUST exigir, además de lo que ya exige (leer `BOE-A-2015-10565` `a21` con `kitlegal boe articulo`, no ejecutar `kitlegal graph show` y citar el bloque): que la sesión ejecute `kitlegal graph check` con la norma `BOE-A-2015-10565` (con bloques o sin ellos), y que la respuesta lleve la forma `⚠ REDACCIÓN MODIFICADA:`.
- **FR-051**: La forma MUST compararse por la forma y sin juicio de ningún modelo, con las tolerancias de H5.1 —la variante de presentación del emoji, el énfasis de Markdown, los espacios y las mayúsculas— y exacta en la etiqueta.
- **FR-052**: `Juzgar` MUST dar la forma por encontrada si la respuesta la lleva, por ausente si no la lleva, y por ausente si la respuesta dice lo mismo con otras palabras (p. ej., «la redacción ha cambiado» sin la forma); con la forma ausente, la sesión no pasa.
- **FR-053**: El estado previo de la sesión de la eval 19 MUST incluir una lectura del bloque con la redacción anterior, de modo que la única lectura que hace la sesión, servida por la caché, dé en `graph check` con la norma un `version-obsoleta` sobre la redacción anterior; sin que nada toque la red.
- **FR-054**: El formato común de eval y `schemas/eval.yaml.json` MUST poder expresar lo que exige FR-050 —la forma fija de una clase de hallazgo de `graph check` que la respuesta tiene que llevar, con los valores de las clases cuya etiqueta da el binario (en este hito, solo `version-obsoleta`), y la norma de un `graph check` esperado—, compatibles hacia atrás: las demás evals de `boe-legislacion` y `legal-core` se leen y se juzgan igual. MUST NOT añadir valores a `avisos` (H7 FR 086). Un valor de clase que el binario no etiqueta es un fichero de eval mal formado, como un aviso desconocido en H5.1.
- **FR-055**: La eval 19 MUST seguir informativa (ADR 0016): las reglas del conjunto (exactamente 10 positivas que deciden, materias distintas) no cambian. El informe del job de evals MUST declarar para ella la forma `⚠ REDACCIÓN MODIFICADA:` y su tasa (ADR 0028), que es el dato para decidir si se promueve.

#### Salida legible de `graph`

- **FR-060**: Sin `--json`, `graph stats`, `graph show` y `graph check` MUST presentar un texto para personas (ADR 0026) en lugar de la tabla mínima del kernel: sin las líneas de procedencia del sobre (`fuente`, `url`, `fecha_consulta`, `hash`) ni pares ruta/valor aplanados. Todo dato que da el texto sale de los valores de `data`, y aparte de ellos solo lleva las frases que pide FR-063; no lleva texto legal (H7 FR 070), es determinista (mismo grafo, mismos argumentos y mismo instante dan los mismos bytes) y no cambia nada de la salida con `--json`. Los fallos siguen sin forma legible propia: su mensaje va a la salida de error.
- **FR-061**: `graph stats` sin `--json` MUST decir cuántos nodos, aristas y textos hay, y el recuento de cada par (tipo, fuente) de nodos y (relación, fuente) de aristas que da su `data`.
- **FR-062**: `graph show` sin `--json` MUST decir el id del nodo, su tipo y sus datos; su primera y su última observación, y la fuente, la url y la fecha de consulta de la última; y cada arista saliente y entrante con su relación, el otro extremo y su procedencia.
- **FR-063**: `graph check` sin `--json` MUST dar cada hallazgo listado con su explicación, agrupados por clase (`version-obsoleta` primero), con el total de cada clase y cuántos se han omitido. Sin hallazgos, MUST ser una frase que dice que no hay nada que volver a comprobar y en qué ámbito: la norma y los bloques pedidos, o todo lo consultado. Sin argumentos, con hallazgos o sin ellos, MUST terminar diciendo cómo acotar la comprobación a una norma.
- **FR-064**: Ningún otro applet cambia su salida sin `--json` (Fuera de alcance).

#### Retirada de H7 y regla genérica

- **FR-070**: Un `world.db` que el binario no puede usar, por cualquier causa distinta de las que trata FR-077, MUST hacer salir a los verbos de `graph` con el código 1, la clase `inesperado` y la ruta de `world.db` en el mensaje; en la entrega de un applet se trata como en H7 FR 033: la salida estándar y el código del applet no cambian y la salida de error lleva una línea que nombra `world.db`. Nada se promete sobre los bytes de `world.db`, sus ficheros auxiliares ni su recuperación (ADR 0023; constitución, «Gates»): para ese caso no vale la promesa de H7 FR 033 de que el grafo quede como estaba. Sustituye la enumeración de H7 FR 010 (dañado, directorio, sin permisos) y su promesa de no modificar el fichero, y reduce H7 US3.5, H7 US6.4, H7 SC 011 y H7 FR 088 a un solo caso, un fichero que no es una base SQLite, como vehículo de la regla.
- **FR-071**: `world.db` MUST crearse en su sitio, sin un fichero temporal con otro nombre ni enlace duro: se retiran la publicación con el temporal `world.db-nuevo-*` y `os.Link` (`internal/graph/publicar.go`, `publicar_test.go`), la costura del enlazador y la limpieza de temporales. Una creación interrumpida deja como mucho el `world.db` sin esquema que ya tratan H7 FR 004 y H7 FR 013. Con `KITLEGAL_CACHE_DIR` en un sistema de ficheros sin enlaces duros, la primera entrega crea `world.db` y no añade ninguna línea de aviso.
- **FR-072**: MUST retirarse, con sus tests y con lo que solo les sirve: la comprobación del permiso de escritura de `world.db` al leer y al entregar y su reapertura en solo lectura inmutable; el tratamiento de un `world.db` que es un enlace simbólico (`internal/graph/integracion_enlace_test.go`), de un diario de rollback (`world.db-journal` caliente, frío o vacío) y de un `-shm` suelto; y el paso a WAL de una base de fuera, con sus cotas y listas de bytes (cabeceras, `auto_vacuum`, diarios `PERSIST` y `TRUNCATE`, otras versiones de SQLite), en `internal/graph/aplicar.go` y `aplicar_test.go`. Esos estados quedan bajo FR-070.
- **FR-073**: Ningún test ni rama de código MUST tratar enlaces simbólicos, diarios de rollback, `-shm` sueltos, permisos cambiados de `world.db` ni bases de fuera.
- **FR-074**: La regla de `Persona` con NIF, NIE o DNI MUST quedarse (constitución VII), con su expresión en clases ASCII y los ejemplos ASCII de H7 FR 025; los casos no ASCII del patrón (`ſ`, U+212A, U+00A0, cifras de anchura completa…) MUST retirarse de `internal/core/grafo/persona_test.go`, y los ejemplos no ASCII de H7 FR 025, H7 SC 010 y su clarificación Q6 dejan de exigirse.
- **FR-075**: `Apply` MUST seguir rechazando el lote entero, sin escribir nada, por una operación sin fuente, dirección o fecha de consulta; por un id que llega con otro tipo que el guardado o que el que le da otra operación del lote; y por un texto cuya huella no es la de su cuerpo o cuya huella ya guarda otro cuerpo; cada uno con un test y sin enumerar más casos. MUST retirarse, con su código y sus tests, los rechazos propios de H7 FR 024 que ningún emisor produce: dirección que no es un URI absoluto, nodo sin id o sin tipo, arista sin relación y arista con un extremo ausente (que la clave ajena de las aristas ya rechaza sin código propio).
- **FR-076**: El desempate de H7 FR 023 MUST quedarse en la `url` —con el mismo instante de consulta, gana la observación de `url` menor comparando bytes— y en que una observación idéntica no cambia nada; MUST retirarse, con su código y sus tests, los niveles siguientes (fuente, `fecha_consulta` tal como se escribe, vigencia y datos en JSON canónico RFC 8785), que solo deciden entre observaciones que nadie produce.
- **FR-077**: MUST quedarse, con sus tests, lo que tiene vía real: H7 FR 004 y H7 FR 013 (`world.db` de 0 bytes o sin esquema, y el `-wal` que deja una escritura propia interrumpida); H7 FR 014 (concurrencia); H7 FR 012 (esquema de una versión posterior, sin migrarlo hacia atrás); y H7 FR 011 y H7 FR 033 (`KITLEGAL_CACHE_DIR`, `HOME` y un directorio de caché que no se puede escribir).
- **FR-078**: `internal/graph` MUST NOT reescribirse más allá de lo que retiran FR-070 a FR-076 y de lo que piden las lecturas (FR-020 a FR-026) y el ámbito de `check` (FR-001 a FR-012).

#### Guiones de aceptación de H7 y esquemas

- **FR-080**: Este hito MUST modificar, en tareas `[datos]` y solo en lo que se dice, estos guiones de H7; cada cambio quita o adapta aserciones de lo que el hito retira o cambia, nunca para tapar un fallo de lo que se queda:
  - `internal/app/testdata/script/h7-grafo-codigos.txtar`: en «world.db que no es una base de datos», las aserciones de que el fichero no cambia y de que no quedan auxiliares (se quedan el 1, la clase `inesperado` y la ruta en el mensaje); la sección «world.db que es un directorio», entera; y la forma de `data` de `graph check`;
  - `internal/app/testdata/script/h7-grafo-entrega-fallida.txtar`: lo mismo en la entrega (se quedan el 0, la misma salida estándar y la línea que nombra `world.db`); la sección «world.db que es un directorio», entera;
  - `internal/app/testdata/script/h7-grafo-concurrencia.txtar`: la aserción de que no queda ningún temporal `world.db-nuevo-*`;
  - `internal/app/testdata/script/h7-grafo-version-obsoleta.txtar`: la sección del orden inverso (H7 US2.3), y lo que cambian la regla nueva y la forma de `data`;
  - `internal/app/testdata/script/h7-grafo-fuente-caducada.txtar`, `h7-grafo-no-emiten.txtar` y `h7-grafo-applet.txtar`: la forma de `data` y la descripción de `graph check`;
  - `internal/app/testdata/script/h7-grafo-show.txtar`: la sección «Sin --json, la tabla mínima del kernel», por la salida legible.
- **FR-081**: Ningún otro guion de H7 ni de hitos anteriores MUST cambiar.
- **FR-082**: `schemas/grafo.json` MUST regenerarse con la forma nueva de `graph check` (FR-012) y `schemas/eval.yaml.json` MUST ganar lo que pide FR-054, sin añadir valores a `avisos`; los dos, en tareas `[datos]`.
- **FR-083**: `specs/010-h7-internal-graph-grafo/` MUST NOT editarse, y no hay ADR nuevo.

#### Controles y Definition of Done

- **FR-090**: `make ci` MUST quedar en verde, con `schema-check` y `skills-check` sin drift.
- **FR-091**: La suite de aceptación congelada de H7.1 MUST ser e2e con relojes fijados y redacciones derivadas de la grabación de H4, como las de H7, sin grabación nueva, y cubrir: la secuencia de FR-025 paso a paso; FR-032; el acotado por norma y por norma y bloque en un grafo con dos normas; una norma desconocida (0 y ningún hallazgo) y un identificador sin forma BOE (2); la salida legible de `stats`, `show` y `check`, con hallazgos y sin ellos, y la de `--json` sin cambios; y el fichero que no es una base SQLite en los verbos de `graph` y en la entrega (FR-070).
- **FR-092**: Un test de integración MUST sembrar con el propio almacén el grafo de la medida —300 normas, 2 400 bloques, 240 de ellos con dos lecturas que ven dos redacciones de fechas de vigencia crecientes, y el 90 % consultado hace más de una semana respecto del instante de la comprobación— y comprobar que `graph check --json` sin argumentos da 50 hallazgos, los `version-obsoleta` primero, con el total de cada clase y lo omitido correctos, en ≤ 40 000 bytes; y que con la norma y un bloque da solo los de ese bloque, sus redacciones y su norma.
- **FR-093**: La comprobación mecánica de FR-045 MUST ejecutarse en `make ci`.
- **FR-094**: Los tests de `Juzgar` MUST cubrir la forma encontrada, ausente y dicha con otras palabras (ausente) (FR-052).
- **FR-095**: Los ficheros retirados (`internal/graph/publicar.go`, `internal/graph/publicar_test.go`, `internal/graph/integracion_enlace_test.go`) MUST no existir, y FR-073 MUST cumplirse.
- **FR-096**: La cobertura MUST quedar dentro de los umbrales de `codecov.yml` (global ≥ 70 %, `internal/core/**` ≥ 85 %, y el diff sin retroceder respecto de la base).
- **FR-097**: `CHANGELOG.md` (*Unreleased*) MUST registrar los cambios visibles: el ámbito y la cota de `graph check` y la forma nueva de su `data`, el apagado de `version-obsoleta`, `fuente-caducada` solo sobre lo vigente, la salida legible de `graph`, la regla genérica y lo retirado, y `boe-legislacion` v0.1.1 (FR-048).
- **FR-098**: El job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final), y su informe da lo que pide SC-006.

### Key Entities

- **Lectura**: la llegada al grafo de un bloque devuelto por una invocación de `boe articulo` o `boe articulos` que entrega; registra qué redacción vio y su orden entre las lecturas del bloque. No se ve en `graph show` ni en `graph stats`.
- **Redacción vigente de un bloque**: la `BloqueVersion` que vio su última lectura; la única de sus redacciones que puede dar `fuente-caducada`.
- **Ámbito de la comprobación**: la norma y los bloques nombrados en `graph check`, o todo lo consultado.
- **Hallazgo**: el de H7 FR 061 (clase, id del nodo, explicación, procedencia y valores propios), ahora dentro de un `data` con recuento y ámbito.
- **Forma fija de `version-obsoleta`**: `⚠`, la etiqueta `REDACCIÓN MODIFICADA` que da el binario, y dos puntos, en la misma línea.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Sobre el grafo sembrado de la medida, `graph check --json` sin argumentos pasa de 3 045 524 bytes a ≤ 40 000, con exactamente 50 hallazgos, los 50 `version-obsoleta` (hay 240), y el total de cada clase y los omitidos iguales a los que da contar el grafo sembrado.
- **SC-002**: Sobre el mismo grafo, con la norma y el bloque de una pregunta, el 100 % de los hallazgos son de esa norma, ese bloque o sus redacciones, y están todos los de ese ámbito.
- **SC-003**: La secuencia de cinco lecturas da exactamente lo que dice FR-025: 0, 1 (sobre A, con las fechas de A y de B; el mismo al repetir), 1 (el mismo tras `--no-graph`), 0 (también en el `graph check` siguiente) y 1 (sobre B, con las fechas de B y de C) hallazgos `version-obsoleta`.
- **SC-004**: Tras el paso 5, sin leer y pasada la vigencia, `graph check` da exactamente 4 hallazgos: `fuente-caducada` sobre la `Norma`, el `Bloque` y C, y `version-obsoleta` sobre B; 0 hallazgos `fuente-caducada` sobre A o B.
- **SC-005**: Lo que lee la skill por pregunta: con k bloques leídos de una norma, `graph check` con esa norma y esos bloques da como mucho k hallazgos, todos `version-obsoleta` y ninguno `fuente-caducada`; con 5 bloques cuya última lectura vio una redacción nueva, ≤ 3 800 bytes con `--json`; sin cambios, 0 hallazgos (≈ 300 bytes); frente a ≈ 6,1 MB por pregunta en H7.
- **SC-006**: En el informe del job de evals, la eval 19 declara la forma `⚠ REDACCIÓN MODIFICADA:` y su tasa, con el veredicto aprobado (el hito cambia `SKILL.md`, Definition of Done §1.10) y `red` vacío.
- **SC-007**: En Claude Code, la misma pregunta sobre un artículo cuya redacción acaba de cambiar, hecha dos veces seguidas, lleva `⚠ REDACCIÓN MODIFICADA:` en la primera respuesta y no en la segunda. Lo comprueba la persona al leer el informe final, fuera del run; su base en el binario son los pasos 2 y 4 de SC-003.
- **SC-008**: Los tests de `Juzgar` distinguen los tres casos de FR-052 (encontrada, ausente, con otras palabras) y la comprobación mecánica de FR-045 falla nombrando `version-obsoleta` sobre un `SKILL.md` sin la forma y pasa sobre el real.
- **SC-009**: Con la norma desconocida `BOE-A-2099-99999`, `graph check` sale con 0 y 0 hallazgos; con `a21` como norma, o con un bloque de solo espacios, sale con 2.
- **SC-010**: Sin `--json`, la salida de `stats`, `show` y `check` (con hallazgos, sin ellos y sin argumentos) no lleva ninguna de las cuatro líneas de procedencia del sobre, lleva cada valor que FR-061 a FR-063 piden de su `data`, y 0 líneas del texto de un bloque de 20 caracteres o más; con `--json`, `stats` y `show` dan byte a byte lo mismo que antes del hito.
- **SC-011**: Con un `world.db` que no es una base SQLite, los tres verbos de `graph` salen con 1 y `inesperado` con la ruta en el mensaje en el 100 % de los casos, y `boe articulo` sale con 0, con la misma salida estándar que con `--no-graph` y una línea en la salida de error que nombra `world.db`.
- **SC-012**: Sobre un `world.db` escrito por H7 con un bloque con dos redacciones, `graph check` da 0 `version-obsoleta` y `graph stats` da los mismos recuentos que antes.
- **SC-013**: Los tres ficheros retirados no existen y ningún test ni rama de código trata enlaces simbólicos, diarios de rollback, `-shm` sueltos, permisos cambiados de `world.db` ni bases de fuera.
- **SC-014**: `make ci` en verde, con `schema-check` y `skills-check` sin drift, la cobertura dentro de los umbrales de `codecov.yml`, `SKILL.md` de `boe-legislacion` por debajo de 300 líneas y la entrada de `CHANGELOG.md` (*Unreleased*).

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). Volumen de referencia: la medida de la bitácora, 300 normas y 2 400 bloques consultados en meses de uso, el 90 % hace más de una semana.

| Salida | Quién la pide, cuántas veces por pregunta y qué hace con cada elemento | Tamaño con meses de uso | Cuándo deja de darse cada señal |
|---|---|---|---|
| `graph check --json` con la norma y los bloques leídos | `boe-legislacion`, paso de responder: una vez por norma citada (una en la mayoría de las preguntas, una más por remisión; FR-040). Traslada cada `version-obsoleta` con la forma fija (FR-042); no hay `fuente-caducada` que ignorar, porque lo recién leído no la da (FR-030, FR-031). | ≤ k hallazgos con k bloques leídos: ≤ 3 800 bytes con cinco que cambiaron, ≈ 300 sin cambios, lo habitual (SC-005). No depende de lo acumulado: con la medida, igual. En H7, ≈ 6,1 MB por pregunta. | `version-obsoleta`: la da la lectura que ve la redacción nueva; la apaga la lectura siguiente del bloque (FR-024): la misma pregunta al día siguiente lee otra vez, la caché sirve la misma redacción y no da nada; al cabo de un mes, solo si la fuente publica otra redacción. |
| `graph check` sin argumentos (con `--json` o sin él) | Una persona que repasa su memoria, a mano y de vez en cuando; ninguna skill (FR-040). Lee los hallazgos, los recuentos y cómo acotar por norma (FR-063). | ≤ 50 hallazgos, ≤ 40 000 bytes con cualquier volumen (SC-001): de 4 812 hallazgos medidos, 50 listados y el resto contado. | `version-obsoleta`: como arriba. `fuente-caducada`: sobre la `Norma`, el `Bloque` y la redacción vigente cuya consulta pasó su vigencia (7 días); la apaga la lectura siguiente de ese bloque (FR-031), y nunca se da sobre una redacción superada. Es un estado de la memoria, no un aviso: se cuenta y se acota, y quien la pide pide precisamente ese estado. |
| `graph check` solo con la norma | Una persona que acota; los hitos que razonan sobre citas sin volver a leerlas (H8, H10), fuera de este hito. La skill siempre nombra los bloques leídos. | ≤ 50 hallazgos, ≤ 40 000 bytes; con la media de la medida (8 bloques por norma), ≈ 17 hallazgos. | Como arriba. |
| `graph stats` (con `--json` o sin él) | Una persona; ninguna skill. | Una línea por par (tipo, fuente) y (relación, fuente): hoy 5 tipos y 3 relaciones de 2 fuentes, ≈ 1 KB con cualquier volumen. | No da señales. |
| `graph show <id>` (con `--json` o sin él) | Una persona que depura; ninguna skill: el protocolo de `boe-legislacion` solo usa `check` (FR-040) y la eval 19 prohíbe `show`. | Un nodo y sus aristas: una `Norma` tiene una arista por bloque consultado de esa norma (8 de media en la medida; como mucho los de la norma, cientos en la más larga), acotado por la norma y no por la memoria entera. Acotarlo más no lo pide el hito (Fuera de alcance). | No da señales. |

## Fuera de alcance

Del hito (literal arriba): que `boe-legislacion` traslade `fuente-caducada`, y usarla en H8 y H10; promover la eval 19 a decisoria; una eval que exija que la respuesta a la pregunta siguiente no lleve la forma; que `version-obsoleta` espere a que una respuesta la diga (la apaga cualquier lectura del bloque, también la de una persona o la de otra skill, y la respuesta siguiente de `boe-legislacion` sobre ese artículo no avisa del cambio); `graph history`, el diff entre redacciones y cualquier clase o regla nueva de `check`; paginar `check` o elegir la cota con una bandera; `boe articulo --fecha` (H20); salida legible de applets distintos de `graph`; que `legal-core` use `graph`; retirar nada de la lista «Se queda» (FR-077); y reescribir `internal/graph`, o editar el spec, el plan o los guiones de H7, fuera de lo que dice el Alcance.

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Que la salida de `graph check` —el hallazgo, su explicación o su salida legible— lleve la forma `⚠ REDACCIÓN MODIFICADA:`: el binario da la etiqueta como única fuente de verdad (FR-045), como las de los avisos, y la forma la escribe la skill.
- Enseñar las lecturas en `graph show` o `graph stats`, o cualquier verbo que las liste.
- Acotar, paginar o resumir `graph show`, y cambiar la `data` de `graph show` o de `graph stats`.
- Guardar la procedencia de la primera observación de un nodo o arista para enseñarla en la salida legible: H7 guarda solo su fecha (H7 FR 002).
- Acotar `graph check` por fuente, tipo de nodo o fecha, o con varias normas en una invocación.
- Cambiar qué emite `boe` o `territorio`, o las reglas de H7 FR 064 sobre fechas de vigencia no válidas o iguales.
- Distinguir el orden de dos lecturas por otra cosa que el orden en que se aplican sus entregas.
- Especificar qué hace el binario, más allá de la regla genérica (FR-070), con un `world.db` enlazado, con permisos cambiados, con diarios o `-shm` ajenos, o escrito por otra aplicación.

## Assumptions

- H7 no ha salido en ninguna release (la última es v0.3.1, anterior a H7): la forma nueva de `data` de `graph check` no rompe a nadie fuera del repositorio.
- La vigencia de una lectura de `boe articulo` es la de la caché de H4 (7 días); la caché no sirve una consulta caducada y, con `--offline`, una consulta sin entrada vigente sale con 4 (H3), así que toda lectura posterior a la caducidad pregunta a la fuente y renueva la observación (FR-031). Una lectura servida por la caché conserva la `fecha_consulta` de la consulta original (H4, ADR 0015).
- Las cifras de la medida (4 812 hallazgos, 3 045 524 bytes, 633 bytes de media) son las de la bitácora del 2026-09-28 y del ADR 0028; la cota de ≤ 40 000 bytes la fija el hito con su cálculo.
- La primera observación de un nodo o arista solo guarda su fecha (H7 FR 002); la salida legible de `show` da la procedencia de la última, que es lo que `data` dice.
- «La redacción que se observó por última vez» de un bloque en un `world.db` de H7 (FR-026) es la `BloqueVersion` de ese bloque con la última observación más reciente. Con H7 la fuente no vuelve a servir una redacción anterior (H7 US2.3 no pasa el umbral), así que coincide con la de fecha de vigencia más alta.
- Las redacciones A, B y C de la suite y el estado previo de la eval 19 se derivan de la grabación de H4 cambiando fecha de vigencia y texto, como las de H7; no hace falta ninguna grabación nueva ni fuente sin fila revisada en `docs/SOURCES.md`.
- SC-007 lo comprueba una persona después del run, al leer el informe final (constitución, capa 3); el run no tiene pausas humanas (ADR 0018).
- Los nombres técnicos que aparecen (`world.db`, `Apply`, `Juzgar`, `schemas/*.json`, los ficheros que se retiran) los fija el hito o una decisión cerrada; los nombres de campos de `data` y del formato de eval, y cómo se guardan las lecturas, son del plan.
