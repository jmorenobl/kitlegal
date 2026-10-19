# Feature Specification: H20 · `boe articulo --fecha`: la redacción vigente en una fecha + `boe-legislacion` v0.2

**Feature Branch**: `024-h20-boe-articulo-fecha`

**Created**: 2026-10-19

**Status**: Draft

**Input**: Sección «#### H20 · `boe articulo --fecha`: la redacción vigente en una fecha + `boe-legislacion` v0.2 (ADR 0027)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Quien lleva un asunto discute con la ley de la fecha que importa —la del acto, la de los hechos, la de un escrito—, no con la de hoy. Un despacho que recurre una resolución de 2019 necesita la redacción de 2019; si la norma cambió en medio, citarle la de hoy es citarle otra norma. Hoy `kitlegal boe articulo` da solo la última redacción de un bloque, aunque la respuesta del BOE las trae todas. El hito hace tres cosas:

1. **`kitlegal boe articulo <norma> <bloque> --fecha AAAA-MM-DD`**, y el parámetro `fecha` de su herramienta MCP `boe_articulo`: la redacción del bloque vigente en esa fecha, elegida por el binario (principio VIII), con la fecha pedida y, si hoy rige otra, su fecha de vigencia y su norma modificadora. Sin `--fecha`, nada cambia.
2. **`boe-legislacion` v0.2**, que, cuando la pregunta lleva una fecha, lee el bloque con `--fecha`, cita esa redacción y lo dice en líneas de forma fija: la fecha pedida y la de vigencia de lo citado, que hoy la redacción es otra si lo es, o que en esa fecha el precepto no tenía redacción vigente.
3. **Dos evals nuevas que deciden y un umbral sin modelo, con 0**, que hacen fallar el job de evals si una sesión no pide la redacción fechada o no da la línea con las fechas esperadas.

Es lo que H8 (validar un texto a su fecha) y H11 (fundamentar un escrito en la redacción vigente a la fecha del acto) dan por hecho.

**Cómo se citan los requisitos de otros hitos.** Los de H4 a H9 se citan sin guion: «H7.1 FR 020». Los ids con guion (FR-001, SC-001…) son siempre de este spec. `<modelo>` es el id del modelo que decide (ADR 0031), hoy `claude-sonnet-5-5`; `<modo>` es `orden` u `herramienta` (H21).

**Términos de este spec.**

- **Lectura fechada**: una invocación de `kitlegal boe articulo` con `--fecha`, o de la herramienta `boe_articulo` con `fecha`. **Lectura sin fecha**: la de hoy, sin ellos.
- **Fecha pedida**: la de `--fecha` o `fecha`.
- **Versión**: cada una de las redacciones de un bloque que trae la respuesta del BOE (`/texto/bloque/<id>`), con su `id_norma`, su `fecha_publicacion` y su `fecha_vigencia`.
- **Versión elegida**: la que devuelve una lectura fechada (FR-002).
- **Redacción de hoy**: la que devuelve hoy la lectura sin fecha, la última versión del bloque (`internal/source/boe/doc.go`, nota [5]). **Redacción posterior**: la redacción de hoy cuando no es la versión elegida.
- **Eval con fecha**: una eval con el esperado de una redacción fechada (FR-052).

## Clarifications

### Session 2026-10-19

- Q: En una lectura fechada de un bloque que tiene versiones con fecha_vigencia y alguna sin ella, ¿cuándo «depende» la elección de la versión sin fecha_vigencia, de modo que la lectura falla con la fila de FR-004 en lugar de elegir entre las versiones con fecha? (FR-004) → A: Siempre que el bloque tenga alguna versión sin `fecha_vigencia`, sea cual sea la fecha pedida (también anterior a la primera con fecha, que no da el 3 de FR-003): falla con la fila de FR-004; solo un bloque con todas las versiones con fecha se elige como FR-002 (auto: criterio a; fuente: docs/ROADMAP.md H20, «La versión de una fecha»; FR-004; internal/source/boe/doc.go, nota [5])
- Q: En la eval (a), ¿la línea de forma fija de la redacción posterior (que hoy la redacción es otra, con 20200206 y BOE-A-2020-1651) entra en la medida del umbral de la redacción fechada, o se comprueba solo en la regla de 2 de 3 de la serie de la eval? (FR-054) → A: Entra en la medida: la sesión que no la lleva, o la lleva con otros valores, cuenta (como mucho una vez); la regla de 2 de 3 no sustituye al umbral; FR-089 gana un caso sintético (auto: criterio a; fuente: docs/ROADMAP.md H20, «El umbral» y «Las evals»; ADR 0038; constitución, «Umbrales que se cumplen»)
- Q: Cuando la última versión de un bloque tiene una fecha_vigencia posterior al día de la consulta (publicada y aún sin entrar en vigor) y la lectura fechada elige otra versión, ¿qué da el binario como redacción posterior y qué dice de ella la línea de forma fija de FR-041? (FR-010, FR-041) → A: La última versión del bloque, la redacción de hoy de la nota [5], con su `fecha_vigencia` y su `norma_modificadora` aunque sea futura; el binario no mira el reloj y la línea no afirma que rija el día de la consulta (auto: criterio c; fuente: spec, «Términos»; internal/source/boe/doc.go, nota [5]; docs/ROADMAP.md H20, «La salida»; FR-022)
- Q: Ante una pregunta con un año o un mes sin día («en 2019», «en marzo de 2018»), ¿qué hace boe-legislacion: pedir el día o leer con una fecha que deriva del periodo, y con cuál? (FR-045) → A: Lee con el primer y con el último día del periodo; si las dos lecturas dan la misma versión (misma `fecha_vigencia` y `norma_modificadora`), la cita diciendo las fechas usadas, y si dan resultados distintos, pide el día (auto: criterio c; fuente: docs/ROADMAP.md H20, «La skill»; FR-002, FR-041, FR-043)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H20. El Objetivo, la Entrega y el Alcance se reflejan en los requisitos.

- **Controles**:
  - «`make ci` en verde, con `schema-check`, el drift de `references/` y las reglas del conjunto;»
  - «golden de un bloque con varias versiones, el art. 22 de la Ley 7/1985, consultado en una fecha anterior a su primera vigencia, en una intermedia, en el día en que entra en vigor una versión y en el anterior, y en una posterior a la última, que da el mismo texto y la misma huella que sin `--fecha`; y del art. 21 de la Ley 39/2015 entre su publicación y su entrada en vigor, que da «no encontrado»;»
  - «e2e con replay y con `--offline`; e2e de que la salida sin `--fecha` es la de `main` byte a byte; e2e de la herramienta MCP con `fecha`; golden de `--describe`;»
  - «tests del grafo: una lectura con `--fecha` emite su `BloqueVersion`, no cambia la fila de lecturas del bloque, ni da ni apaga `version-obsoleta` y no recibe `fuente-caducada`;»
  - «tests de la caché: una lectura con `--fecha` no se sirve de la entrada de hoy, y la entrada de hoy no cambia;»
  - «la sección «Controles de umbral» del plan con una fila por cada umbral nuevo y las de los que ya decidían; tests del umbral con sesiones sintéticas;»
  - «un mutante por control: con la versión elegida por `fecha_publicacion`, con la última siempre, con una fecha anterior al bloque que devuelve la primera versión, con la lectura fechada contada como lectura del bloque o con el `decide` del umbral en falso, algún test falla;»
  - «`CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio.»
- **Aceptación**: «en el informe del job de cierre de `boe-legislacion`, el umbral nuevo lleva `decide: true` y se cumple en los dos modos; `afirma_lo_no_leido:<modelo>`, `sin_activar`, las citas, los avisos, la medida del juez y las duraciones se cumplen; ninguna respuesta queda «sin juzgar» y el veredicto es aprobado; `red` está vacío. `jurisprudencia` y `legal-core`, si salen en rojo, no son motivo de este hito y quedan en el informe. El informe final da el umbral nuevo como «comprobado por su control». Después, y fuera del run porque es humano: con la release publicada, en la app de escritorio de Claude con el plugin, «¿qué exigía el artículo 118 de la LCSP para un contrato menor el 1 de junio de 2019?» se responde con la redacción vigente desde el 9 de marzo de 2018, con su cita y su fecha, avisando de que hoy rige la del Real Decreto-ley 3/2020, desde el 6 de febrero de 2020.»
- **Al cerrar**: «la fase 1 sigue con H9.1 (ADR 0040). H8 y H11 pueden dar por hecha la redacción de un bloque a una fecha.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega 1, `boe articulo --fecha` y `boe_articulo` con `fecha` | FR-001 a FR-007, FR-010 a FR-014, FR-020 a FR-022, FR-030 a FR-032, US1, US2, US4 |
| Entrega 2, el formato de las evals | FR-051, FR-052 |
| Entrega 3, `boe-legislacion` v0.2 | FR-040 a FR-049, US1, US2, US3 |
| Entrega 4, dos evals y un umbral sin modelo | FR-050, FR-053 a FR-059, US5 |
| Entrega 5, documentos | FR-060 a FR-063 |
| Lo que no cambia y el cierre | FR-070 a FR-072 |
| Controles | FR-080 a FR-091, SC-004 a SC-012 |
| Aceptación | SC-001 a SC-003 |
| Al cerrar | FR-092 |

## Relación con hitos anteriores

Ningún spec, plan ni informe de un hito anterior se edita: son el registro de sus runs. No hay ADR nuevo: la redacción a una fecha la sitúa el ADR 0027; las lecturas del grafo, el ADR 0014 con la enmienda de H7.1 que transcribe el hito; los umbrales con cero de lo que comprueba un guion, el ADR 0038.

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- **«`kitlegal boe articulo` da solo la redacción vigente»** (H4; `SKILL.md` de `boe-legislacion`, «Redacción modificada» y paso 5): vale para la lectura sin fecha; la lectura fechada da la de la fecha pedida (FR-001, FR-049).
- **Una lectura es cada arista `eli:has_version` de una entrega** (H7.1; `Consolidado.Lecturas`, `internal/core/grafo/lecturas.go`): vale para las lecturas sin fecha; una lectura fechada no es lectura del bloque (FR-031).
- **El formato de las evals** (H5 y siguientes): un comando de la forma bloque puede llevar una fecha, y una eval, el esperado de una redacción fechada (FR-051, FR-052).

**Se queda**, sin que este spec lo vuelva a definir:

- **La lectura sin fecha** (H4, nota [5]): qué versión da, sus campos, su caché, sus códigos de salida y lo que emite al grafo (H7, H7.1).
- **Las banderas globales** (`--json`, `--timeout`, `--offline`, `--dry-run`, `--no-graph`, `--asunto`, `--verbose`): valen con `--fecha` como sin ella.
- **Qué sesiones se miden y qué es una sesión sin medir** (H7.3, H21): las del modelo que decide, por modo, en las evals que activan la skill; una sesión sin medir da su motivo «de la ejecución, no de la skill».
- **El contrato de `umbrales`** (ADR 0029): `nombre`, `descripcion`, `medida`, `total`, `comparacion`, `umbral`, `cumple` y `decide`.
- **El modelo que decide, las repeticiones (3) y la regla de 2 de 3 por eval** (ADR 0016 y 0031), el juez (ADR 0037), su umbral en tasa (ADR 0038) y la definición del job: su tope de 352 minutos, la concurrencia de `boe-legislacion` a cinco sesiones y su objetivo de duración de 900 s (#168, en `main`).
- **El sobre** (`ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`) y **el contrato de resultados** (ADR 0023).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La redacción que regía en la fecha del acto (Priority: P1)

Un despacho prepara un recurso contra la adjudicación de un contrato menor de junio de 2019 y pregunta qué exigía entonces el art. 118 de la LCSP para su expediente. La respuesta cita la redacción original, vigente desde el 9 de marzo de 2018, con su cita del bloque, la fecha pedida y la de vigencia, y avisa de que hoy rige otra, la de `BOE-A-2020-1651` (el Real Decreto-ley 3/2020), vigente desde el 6 de febrero de 2020. No cuenta qué dice la de hoy si no la ha leído.

**Why this priority**: es el objetivo del hito. Citar la redacción de hoy a un asunto de 2019 es citar otra norma.

**Independent Test**: la eval (a) de FR-050 en el job de evals y su lectura con `--offline` contra lo grabado, sin ninguna otra pieza de la skill.

**Acceptance Scenarios**:

1. **Dado** lo grabado del bloque `a1-30` de `BOE-A-2017-12902`, **Cuando** se ejecuta `kitlegal boe articulo BOE-A-2017-12902 a1-30 --fecha 2019-06-01 --json`, **Entonces** sale con 0 y `data` lleva la versión con `fecha_vigencia` `20180309` y `norma_modificadora` `BOE-A-2017-12902`, su texto y su huella, la fecha pedida, y la redacción posterior con su fecha de vigencia `20200206` y su norma modificadora `BOE-A-2020-1651`.
2. **Dada** la pregunta de la eval (a), **Cuando** responde `boe-legislacion` en el modo orden o en el modo herramienta, **Entonces** la sesión lee el bloque `a1-30` con la fecha 2019-06-01, y la respuesta cita el bloque como en «Cómo se cita» y lleva la línea de forma fija con la fecha pedida y la fecha de vigencia `20180309`, y la de la redacción posterior con `20200206` y `BOE-A-2020-1651`, tal como los da el binario.

---

### User Story 2 - En esa fecha el precepto no tenía redacción vigente (Priority: P1)

Alguien pregunta qué decía el art. 21 de la Ley 39/2015 el 1 de junio de 2016. La ley se publicó el 2 de octubre de 2015 y entró en vigor el 2 de octubre de 2016: en la fecha pedida el precepto no tenía redacción vigente. La respuesta lo dice en una línea de forma fija, con desde cuándo la tiene, no cita el texto como vigente entonces y no dice qué norma regía sin haberla leído.

**Why this priority**: elegir por la fecha de publicación daría por vigente un texto que no regía; es el error que el hito evita.

**Independent Test**: la eval (b) de FR-050 y el golden del art. 21 de FR-081.

**Acceptance Scenarios**:

1. **Dado** lo grabado del bloque `a21` de `BOE-A-2015-10565`, **Cuando** se ejecuta `kitlegal boe articulo BOE-A-2015-10565 a21 --fecha 2016-06-01`, **Entonces** sale con 3 (`no-encontrado`), sin texto, y el mensaje nombra la fecha pedida y la de vigencia de la primera redacción, `20161002`.
2. **Dada** la pregunta de la eval (b), **Cuando** responde `boe-legislacion` en cualquiera de los dos modos, **Entonces** la sesión lee el bloque `a21` con la fecha 2016-06-01, y la respuesta lleva la línea de forma fija de que en esa fecha el precepto no tenía redacción vigente y la tiene desde el `20161002`, y no cita el texto del art. 21 como vigente en esa fecha.

---

### User Story 3 - La pregunta sin fecha sigue igual (Priority: P1)

Las preguntas sin fecha —qué dice hoy un artículo, dónde se regula una materia— se responden como hasta ahora, con la redacción de hoy y su cita, y la salida del binario sin `--fecha` no cambia ni un byte.

**Why this priority**: las veintiuna evals de hoy, las otras skills y la web leen la salida sin fecha; cambiarla las rompería.

**Independent Test**: el e2e de FR-083 contra la salida de `main` y las veintiuna evals, que no cambian (FR-050).

**Acceptance Scenarios**:

1. **Dado** lo grabado, **Cuando** se ejecuta `kitlegal boe articulo BOE-A-1985-5392 a22` sin `--fecha`, con `--json` y sin él, **Entonces** la salida y el código son byte a byte los de `main`.
2. **Dada** una pregunta sin fecha, **Cuando** responde `boe-legislacion`, **Entonces** lee el bloque sin `--fecha` y no lleva ninguna de las líneas de forma fija de la redacción fechada (FR-046).

---

### User Story 4 - Una lectura fechada no engaña al grafo (Priority: P2)

Quien leyó el art. 22 de la LRBRL hoy y pide luego su redacción de 2005 no debe ver en su siguiente consulta que «la redacción ha cambiado»: la de 2005 la pidió a propósito y no es la vigente.

**Why this priority**: `graph check` es la memoria con la que `boe-legislacion` avisa de cambios de redacción (H7.2); una lectura fechada contada como lectura daría avisos falsos.

**Independent Test**: los tests del grafo de FR-086.

**Acceptance Scenarios**:

1. **Dado** un `world.db` en el que el bloque `a22` tiene una lectura sin fecha, **Cuando** se lee con `--fecha 2005-06-15` y después sin fecha, **Entonces** la fila de lecturas del bloque y lo que da `graph check` tras la segunda lectura son los mismos que sin la lectura fechada, y el grafo tiene el `BloqueVersion` de la versión con `fecha_vigencia` `20040101`.

---

### User Story 5 - El job sale en rojo si la respuesta no da la redacción fechada (Priority: P1)

El job de evals de `boe-legislacion` falla si una sesión de una eval con fecha no pide la redacción fechada o no da la línea con las fechas esperadas.

**Why this priority**: sin un umbral que decide, la skill podría citar la redacción de hoy a una pregunta con fecha y el cierre saldría en verde.

**Independent Test**: los tests del umbral con sesiones sintéticas (FR-089).

**Acceptance Scenarios**:

1. **Dada** una sesión sintética del modelo que decide en la eval (a) que lee `a1-30` sin `--fecha`, **Cuando** se calcula el informe, **Entonces** el umbral nuevo de su modo cuenta 1, no se cumple y el veredicto es `fallo`.
2. **Dada** una sesión que lee con la fecha de la eval y lleva la línea con las fechas esperadas, **Cuando** se calcula el informe, **Entonces** no cuenta.

---

### Edge Cases

- **La fecha pedida es el día en que entra en vigor una versión**: se elige esa (FR-002).
- **La fecha pedida cae entre la publicación y la entrada en vigor de una versión**: rige la anterior, o ninguna si es la primera (FR-002, FR-003).
- **La fecha pedida es anterior a la primera vigencia del bloque**: código 3 con la fecha de vigencia de la primera redacción en el mensaje (FR-003).
- **La fecha pedida es posterior a la última vigencia**: la redacción de hoy, con el mismo texto y la misma huella que sin `--fecha`, y sin redacción posterior (FR-013).
- **Una versión sin `fecha_vigencia`, o un bloque sin versiones**: no se sitúan con otra fecha; con cualquier versión sin `fecha_vigencia` en el bloque, toda lectura fechada falla (FR-004).
- **Una fecha mal escrita o que no existe en el calendario**: código 2 (FR-005).
- **Una norma o un bloque que no existen, o la fuente que no responde**: como en la lectura sin fecha (FR-007).
- **`--offline` sin nada guardado que sirva la lectura fechada**: falla como la caché sin entrada (FR-020).
- **Una norma hoy derogada leída en una fecha en que estaba vigente**: lleva los avisos de hoy (FR-011).
- **Una pregunta por un año o un mes sin día**: dos lecturas, con el primer y el último día del periodo; la misma versión se cita, versiones distintas piden el día (FR-045).
- **Una pregunta que sitúa el momento por un hecho sin fecha** («cuando se adjudicó mi contrato»): la skill pide el día y no cita ninguna redacción como la de ese momento (FR-040).
- **La última versión del bloque con una `fecha_vigencia` posterior al día de la consulta**: es la redacción posterior si no es la elegida, con su fecha, sin que el binario mire el reloj (FR-041).
- **Ficheros del kit manipulados a mano**: lo cubre la regla genérica, defecto `inesperado` y código 1 (ADR 0023), sin casos.

## Requirements *(mandatory)*

### Functional Requirements

#### La lectura fechada: `boe articulo --fecha`

- **FR-001**: `kitlegal boe articulo <norma> <bloque>` MUST admitir `--fecha AAAA-MM-DD` y devolver la versión del bloque vigente en esa fecha (FR-002). La herramienta MCP `boe_articulo` MUST admitir el parámetro opcional `fecha`, generado desde `--describe` como los demás, y dar el mismo resultado que la orden con la misma entrada. Sin `--fecha` ni `fecha`, la lectura es la de hoy (FR-012). Comprobable: el golden de FR-081, el e2e de la herramienta de FR-084 y el golden de `--describe` de FR-085.
- **FR-002**: La versión elegida MUST ser la de mayor `fecha_vigencia` que no es posterior a la fecha pedida, comparadas como fechas del calendario; el día en que una versión entra en vigor, es esa. MUST elegirse por `fecha_vigencia` y nunca por `fecha_publicacion`: entre la publicación y la entrada en vigor de una versión rige la anterior, o ninguna. Comprobable: el golden de FR-081 y los mutantes de FR-090.
- **FR-003**: Una fecha pedida anterior a la `fecha_vigencia` de la primera versión del bloque MUST salir con 3 (`no-encontrado`, ADR 0023), sin texto, y su mensaje MUST nombrar la fecha pedida y la fecha de vigencia de la primera redacción, en una forma que fija el plan y que la distingue del 3 de un bloque que no está en la norma. Comprobable: el golden de FR-081 (el art. 22 en una fecha anterior a 19850423 y el art. 21 de la Ley 39/2015 en 2016-06-01) y el mutante de FR-090.
- **FR-004**: Una versión sin `fecha_vigencia` y un bloque sin versiones (nota [5]) MUST NOT situarse con otra fecha, ni con su `fecha_publicacion` ni con la de otra versión. Una lectura fechada de un bloque sin versiones MUST fallar con la fila del ADR 0023 que fije el plan, sin devolver texto. Una lectura fechada de un bloque con alguna versión sin `fecha_vigencia` MUST fallar con esa misma fila siempre que el bloque tenga alguna versión sin `fecha_vigencia`, sea cual sea la fecha pedida y sin texto; también con una fecha anterior a la primera versión con fecha, que no da el 3 de FR-003, porque la versión sin fecha pudo entrar en vigor antes. Sin su `fecha_vigencia`, una versión pudo entrar en vigor cualquier día: para cualquier fecha pedida pudo ser la vigente, y situarla entre las de otras versiones es lo que este requisito prohíbe. Solo en un bloque cuyas versiones tienen todas `fecha_vigencia` MUST elegirse como FR-002. La lectura sin fecha no cambia (nota [5]). Comprobable: un test con una respuesta del BOE sin versiones y otro con una versión sin `fecha_vigencia`, construidos en el test, que comprueba el fallo con fechas pedidas anteriores, intermedias y posteriores a las versiones con fecha.
- **FR-005**: `--fecha` MUST ser una fecha `AAAA-MM-DD` que existe en el calendario. Si no lo es —otra forma, un día que no existe como 2019-02-30, o vacía—, la lectura MUST salir con 2 (`argumentos`) antes de abrir la caché o pedir nada, como hoy una norma o un bloque mal formados. Comprobable: el e2e de FR-082.
- **FR-006**: La lectura fechada MUST pedir los mismos recursos que la lectura sin fecha del mismo bloque (el bloque y los metadatos de su norma), con la misma fila de `docs/SOURCES.md` (`boe.legislacion-consolidada`), sin grabaciones nuevas en `testdata/`. Comprobable: el e2e con replay de FR-082, que sale en rojo si pide algo que no está grabado, y el diff del hito, que no toca `docs/SOURCES.md` ni añade grabaciones del BOE.
- **FR-007**: Con `--fecha`, una norma o un bloque mal formados, un bloque que no está en la norma, la fuente que no responde o limita el ritmo, y las banderas globales MUST dar lo mismo que en la lectura sin fecha. Comprobable: el e2e de FR-082.

#### La salida

- **FR-010**: Con el código 0, `data` de una lectura fechada MUST llevar el mismo objeto que la lectura sin fecha —`norma`, `bloque`, `titulo`, `tipo`, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, `url_eli`—, con los datos de la versión elegida tal como los da la nota [5] para la última, y además:
  - la fecha pedida;
  - si hay redacción posterior, su fecha de vigencia y su norma modificadora, de la misma respuesta del BOE; si no la hay, nada que diga que la hay.

  Los nombres y la forma de los campos nuevos los fija el plan. La salida legible, sin `--json`, MUST llevar también la fecha pedida y, si la hay, la redacción posterior. Comprobable: el golden de FR-081.
- **FR-011**: Los avisos de vigencia de una lectura fechada MUST ser los de la norma tal como está hoy, los mismos que da la lectura sin fecha: una redacción antigua de una norma hoy derogada sigue avisando. Comprobable: un test con lo grabado sintético de avisos (`internal/source/boe/testdata/sintetico/avisos/`) leído con `--fecha`.
- **FR-012**: Sin `--fecha`, la salida de `boe articulo` —`data`, `hash`, el sobre, la salida legible y el código, en éxito y en fallo— MUST ser byte a byte la de `main`. `--describe` MUST cambiar solo por la bandera nueva y los campos nuevos de FR-010. Comprobable: el e2e de FR-083 y el golden de FR-085.
- **FR-013**: Con una fecha pedida posterior a la última `fecha_vigencia` del bloque, la lectura fechada MUST dar el mismo `texto` y el mismo `hash_texto` que la lectura sin fecha, y ninguna redacción posterior. Comprobable: el golden de FR-081.
- **FR-014**: Toda salida de una lectura fechada MUST validar contra `schemas/`, y `make schema-check` MUST fallar si `schemas/` no coincide con lo que emite `--describe` (Definition of Done §1.4). Comprobable: `make ci`.

#### La caché

- **FR-020**: Una lectura fechada MUST NOT servirse de la entrada de hoy del artículo, la que guarda el artículo compuesto con la redacción de hoy bajo la clave de su norma y su bloque (`articuloDelBloque`, `internal/source/boe/articulo.go`). Con `--offline` y solo esa entrada en la caché, MUST fallar como hoy falla la caché sin entrada con `--offline`: 4 (`fuente-no-disponible`). Comprobable: el test de la caché de FR-087.
- **FR-021**: La entrada de hoy MUST NOT cambiar de forma ni de clave, y una lectura fechada MUST NOT escribir en ella la versión elegida: tras una lectura fechada de una versión que no es la de hoy, la lectura sin fecha con `--offline` da byte a byte lo mismo que antes. Comprobable: el test de la caché de FR-087.
- **FR-022**: Cómo se guarda una lectura fechada lo fija el plan. Con `--offline` y algo guardado y vigente que la sirva, MUST dar el mismo `data` y el mismo `hash` que la misma lectura fechada sin `--offline`, con su fecha de consulta. Comprobable: el e2e con `--offline` de FR-082.

#### El grafo (ADR 0014)

- **FR-030**: Una lectura fechada con código 0, sin `--no-graph` ni `--dry-run`, MUST emitir en su `Resultado` el `BloqueVersion` de la versión elegida, con el id y los datos con que lo emite hoy la lectura sin fecha para su versión (fecha de vigencia, fecha de versión, norma modificadora y huella), con la `Procedencia` del propio resultado como `source` y sin cambiar el contrato `Applet` (ADR 0005). Ninguna operación MUST entrar sin `source`, y aplicar dos veces el mismo resultado MUST NOT duplicar nada (Definition of Done §1.12). Qué más emite con él lo fija el plan. Comprobable: el test del grafo de FR-086.
- **FR-031**: Una lectura fechada MUST NOT contar como lectura del bloque para `graph check`, sea cual sea la versión que devuelve:
  - la fila de lecturas del bloque MUST quedar como estaba;
  - lo que da `graph check` sobre el bloque y sus versiones MUST ser lo mismo antes y después: ni da ni apaga `version-obsoleta`, y la versión elegida no pasa a ser la vigente de su bloque, así que no recibe `fuente-caducada`;
  - la siguiente lectura sin fecha del bloque MUST dar en `graph check` lo mismo que daría sin la lectura fechada, también en un `world.db` en el que el bloque no tiene fila de lecturas (H7.1).

  Cómo se emite sin contar lo fija el plan. Comprobable: los tests del grafo de FR-086 y el mutante de FR-090.
- **FR-032**: Con `--no-graph` o `--dry-run`, una lectura fechada MUST NOT entregar nada al grafo, como la lectura sin fecha. Comprobable: el test del grafo de FR-086.

#### La skill: `boe-legislacion` v0.2

- **FR-040**: Cuando la pregunta lleva una fecha —«el día D», «en 2019», o un hecho («cuando se dictó la resolución») cuya fecha da la pregunta o la conversación—, `skills/boe-legislacion/SKILL.md` MUST leer el bloque con `kitlegal boe articulo … --fecha` o con `boe_articulo` y `fecha`, y citar esa redacción. La cita del bloque MUST seguir la forma de «Cómo se cita». Si la pregunta sitúa el momento por un hecho y ni ella ni la conversación dan su fecha, la skill MUST pedir el día y MUST NOT leer ni citar ninguna redacción como la de ese momento. Comprobable: el umbral de FR-054 en las evals con fecha; lo del hecho sin fecha, los jueces de la revisión final.
- **FR-041**: Tras una lectura fechada con código 0, la respuesta MUST llevar una línea de forma fija con la fecha pedida y la fecha de vigencia de la redacción citada, y, si el binario da una redacción posterior, otra línea de forma fija que dice que hoy la redacción es otra, con su fecha de vigencia y su norma modificadora, tal como las da el binario. La línea de la redacción posterior MUST presentarla como la redacción de hoy del bloque, con la fecha de vigencia y la norma modificadora tal como las da el binario, sin afirmar que rija o esté vigente el día de la consulta ni comparar su fecha con ese día; con una vigencia futura, la propia fecha dice que aún no ha entrado en vigor. La respuesta MUST NOT decir hasta cuándo rigió la redacción citada ni presentar la fecha de vigencia de la posterior como su fin: entre las dos puede haber otras versiones. Comprobable: el umbral de FR-054 y la regla de su serie (FR-057); lo de hasta cuándo, los jueces de la revisión final.
- **FR-042**: Para decir qué dice hoy el bloque, la skill MUST leerlo sin fecha y citarlo. MUST NOT describir, resumir ni comparar una redacción que no ha leído en la conversación (ADR 0037, frontera del precepto identificado): la línea de la redacción posterior da su fecha y su norma, nada de su contenido. Comprobable: `afirma_lo_no_leido:<modelo>` (FR-058).
- **FR-043**: Si la lectura fechada sale con 3 con el mensaje de FR-003, la respuesta MUST llevar una línea de forma fija que dice que en esa fecha el precepto no tenía redacción vigente y desde cuándo la tiene, con la fecha de vigencia de la primera redacción que da el mensaje. MUST NOT citar el texto como vigente en esa fecha ni decir qué norma regía entonces sin haberla leído. Un 3 de un bloque que no está en la norma sigue con la regla 2. Comprobable: el umbral de FR-054 en la eval (b) y `afirma_lo_no_leido:<modelo>`.
- **FR-044**: Los avisos de vigencia de una lectura fechada MUST trasladarse como hoy (regla 3, «Cómo se cita»). Comprobable: la regla de la serie de cada eval con su clave `avisos`.
- **FR-045**: Ante una pregunta por un año o un mes sin día, la skill MUST hacer dos lecturas fechadas del bloque, una con el primer día y otra con el último del periodo (del año: 01-01 y 12-31; del mes: el día 1 y su último día). Dos lecturas dan la misma versión si sus salidas dan la misma `fecha_vigencia` y la misma `norma_modificadora`. Si es la misma versión, la skill MUST citarla como la del periodo, con la línea de FR-041 de cada lectura, que dice las fechas usadas, y con la de la redacción posterior si la hay. Si las dos dan el 3 de FR-003, la respuesta MUST llevar la línea de FR-043 de cada lectura. Si dan resultados distintos (dos versiones, o un 3 y una versión), la redacción cambió dentro del periodo: la skill MUST NOT citar ninguna como la del periodo y MUST pedir el día. El texto lo fija el plan. Comprobable: lo leen los jueces de la revisión final.
- **FR-046**: Cuando la pregunta no lleva una fecha, ni un año o un mes, ni sitúa el momento por un hecho (FR-040, FR-045), la skill MUST leer sin `--fecha` y su respuesta MUST NOT llevar ninguna de las líneas de FR-041 y FR-043. La sección «Redacción modificada» y la regla 7 MUST referirse solo a lecturas sin fecha: si la skill pide `graph check` tras una lectura fechada lo fija el plan, y la línea `⚠ REDACCIÓN MODIFICADA:` y la frase de la regla 7 MUST NOT ir en la respuesta por una lectura fechada. Comprobable: lo esperado de los comandos de las veintiuna evals de hoy, que no cambian y no casa con una lectura fechada (FR-051), en la regla de su serie; las líneas y «Redacción modificada», los jueces de la revisión final.
- **FR-047**: La `description` de `boe-legislacion` MUST ganar las preguntas con fecha, para que la skill se active con ellas, y seguir ganando las de hoy. **Control**: `sin_activar:<modelo>:<modo>`, con 0 y que decide, cuenta las sesiones de las evals con fecha (FR-050): con una que no activa la skill, el job de `boe-legislacion` sale en rojo.
- **FR-048**: `SKILL.md` MUST seguir por debajo de 300 líneas (hoy 299) sin perder ninguna regla, con la tabla de comandos regenerada desde `--describe`, y `references/` MUST seguir siendo solo lo generado desde `data/`. Las formas fijas de FR-041 y FR-043 MUST ir tal cual en bloques `text`, con el texto que fija el plan, de modo que un guion las compruebe sin modelo. **Control**: `make skills-check`, en `make ci`, falla con 300 líneas o más, o con una tabla o unas `references/` distintas de las regeneradas. Que no se pierda ninguna regla lo leen los jueces de la revisión final.
- **FR-049**: Lo que `SKILL.md` dice y deja de ser cierto con la lectura fechada —hoy, que `kitlegal boe articulo` da solo la redacción vigente— MUST corregirse para que valga para la lectura sin fecha, y solo eso. Comprobable: lo leen los jueces de la revisión final.

#### Las evals y los umbrales

- **FR-050**: Las veintiuna evals de `boe-legislacion` MUST seguir como están, y MUST entrar dos, que activan la skill y deciden —ni `informativa` ni `por_materia`, porque nombran el artículo, como la 01—, y ninguna más:
  - (a) «¿qué exigía el artículo 118 de la LCSP para el expediente de un contrato menor el 1 de junio de 2019?», que se responde con la redacción vigente desde el 20180309, la original, y avisa de que hoy es la de `BOE-A-2020-1651` (el Real Decreto-ley 3/2020, que el binario no nombra), vigente desde el 20200206;
  - (b) «¿qué decía el artículo 21 de la Ley 39/2015 el 1 de junio de 2016?», que se responde diciendo que en esa fecha no tenía redacción vigente y que la tiene desde el 20161002.

  Las dos piden `boe articulo` con `--fecha` sobre su bloque (`a1-30` de `BOE-A-2017-12902` con 2019-06-01; `a21` de `BOE-A-2015-10565` con 2016-06-01). **Control**: `TestDefinicionDelJob`, en `make ci`, recalcula el peor caso del trabajo de `boe-legislacion` con las evals del repositorio y su concurrencia de cinco, y falla si el tope de 352 minutos no lo cubre.
- **FR-051**: El formato de las evals MUST admitir la fecha de un comando de la forma bloque. Un comando esperado con fecha casa solo con una lectura de ese bloque con esa fecha, como orden con `--fecha` o como herramienta con `fecha`; uno sin fecha casa solo con una lectura sin fecha. Su clave la fija el plan. **Control**: la validación de las evals contra `schemas/eval.yaml.json`, en `make ci`, y los tests del umbral de FR-089.
- **FR-052**: El formato de las evals MUST admitir el esperado de una redacción fechada, que se comprueba sin modelo por las líneas de forma fija de FR-041 y FR-043: o la fecha pedida y la fecha de vigencia de la redacción citada, con la fecha de vigencia y la norma modificadora de la posterior si la hay; o que en la fecha pedida no había redacción vigente, con la fecha de vigencia de la primera. Sus claves las fija el plan; `schemas/eval.yaml.json` y las reglas del conjunto de `boe-legislacion` lo admiten en una eval positiva. **Control**: la validación de las evals contra el esquema y las reglas del conjunto, en `make ci`.
- **FR-053**: La caché preparada de las evals MUST servir la lectura fechada de cada bloque que sirve sin fecha, entre ellos los dos de FR-050, que están en lo grabado de H4 (`internal/source/boe/testdata/boe.legislacion-consolidada/`) con todas sus versiones; ninguna sesión llega a la red. **Control**: el job de evals sale en rojo si `red` no está vacío.
- **FR-054**: Por cada modo, `umbrales` de `boe-legislacion` MUST llevar un umbral de la redacción fechada, con el contrato del ADR 0029:
  - `total`: las sesiones de ese modo que cuentan en el `total` de `sin_activar:<modelo>:<modo>` y son de una eval con fecha (6 por modo con 3 repeticiones);
  - `medida`: las de esas sesiones que no ejecutan `boe articulo` con la fecha de la eval sobre su bloque, ni como orden ni como herramienta (FR-051), o cuya respuesta no lleva la línea de forma fija con las fechas esperadas (FR-052), que en la eval (a) son la fecha pedida y la vigencia `20180309` y también la línea de la redacción posterior (que hoy la redacción es otra) con `20200206` y `BOE-A-2020-1651` tal como las da el binario; una respuesta que no lleva esa línea, o la lleva con otros valores, cuenta. El guion que calcula la `medida` compara sin modelo ambas líneas; la regla de 2 de 3 de la serie (FR-057) sigue con sus comandos, sus citas y sus avisos y no sustituye a este umbral. Cada sesión cuenta una vez como mucho;
  - `comparacion` `"<="`, `umbral` 0 y `decide: true`.

  Su nombre lo fija el plan. **Control**: con una sesión que cuenta, el veredicto es `fallo` y el job de evals de `boe-legislacion` sale en rojo.
- **FR-055**: El umbral de FR-054 MUST:
  - poner el veredicto en `fallo` cuando no se cumple, con un motivo que nombra el umbral con su medida y su condición, como el de todo umbral que decide, y detrás cada sesión que cuenta, por su nombre y en orden de sesión; el texto lo fija el plan; uno que se cumple no da motivo;
  - dejar fuera de su `total` y de su `medida` las sesiones sin medir, que siguen dando su motivo «de la ejecución, no de la skill», las de un modelo informativo (hoy `claude-haiku-4-5-20251001`) y las de la eval sin binario ni servidor;
  - salir de las claves de cada eval, no de una lista en el código, y sin ningún caso por skill.

  Comprobable: los tests de FR-089.
- **FR-056**: `umbrales` de `boe-legislacion` MUST llevar exactamente sus trece elementos de hoy y el umbral de FR-054 de cada modo: quince, once que deciden. El orden lo fija el plan. `umbrales` de `legal-core` MUST seguir con sus seis, los seis que deciden, y el de `jurisprudencia`, con sus quince, once que deciden. **Control**: un test de `make ci` falla con otros elementos, otro orden o otro `decide` en alguno (FR-089).
- **FR-057**: Las dos evals con fecha MUST seguir además la regla de 2 de 3 de su serie, con sus comandos esperados, sus citas y sus avisos, que fija el plan, como las de hoy. Comprobable: el job de evals.
- **FR-058**: Las respuestas de las dos evals nuevas MUST entrar en `afirma_lo_no_leido` (ADR 0037): de 108 por ejecución, con los dos modos, a 120. El umbral MUST seguir en la proporción de `evals/boe-legislacion/juez/clases.yaml`, 0,028, que con 120 admite 3; la rúbrica, los casos etiquetados y la medida del juez MUST NOT cambiar. **Control**: `afirma_lo_no_leido:<modelo>`, que decide, hace fallar el job con más de 3 marcadas de 120; una respuesta sin juzgar pone el veredicto en `fallo`; y `medida_del_juez:…`, que deciden, lo hacen fallar si la medida no corresponde a la rúbrica, a los casos, al modelo del juez y a su versión de Claude Code.
- **FR-059**: El hito MUST NOT cambiar `.github/workflows/evals.yml` ni los tests que leen su concurrencia: la de `boe-legislacion` a cinco y su objetivo de duración de 900 s están en `main` (#168). Comprobable: el diff del hito. **Control** del objetivo: `duracion_de_las_sesiones:<modo>` y `duracion_del_juez:<modo>`, que deciden, hacen fallar el job por encima de 900 s.

#### Documentos

- **FR-060**: `CHANGELOG.md` (*Unreleased*) MUST registrar `--fecha` en `boe articulo` y `fecha` en `boe_articulo`, con la regla de elección, el 3 de una fecha anterior al bloque y la redacción posterior; `boe-legislacion` v0.2; las dos evals y el umbral nuevo (qué cuenta, sobre qué sesiones, con 0 y que decide); y que una lectura fechada no cuenta como lectura del bloque en `graph check`. Las secciones de las versiones publicadas MUST NOT cambiar. Comprobable: SC-012.
- **FR-061**: `README.md`, `CONTRIBUTING.md` y `docs/WORKFLOW.md` MUST corregirse donde lo que dicen de `boe articulo`, de `boe-legislacion`, de sus evals, de su formato o de sus umbrales quede falso o incompleto con el hito, y solo ahí. Comprobable: SC-012.
- **FR-062**: Esa corrección MUST NOT reescribir lo que sigue siendo cierto, ni prometer ninguna release, ni añadir un test que ate ese texto al código. Cada cifra sale del árbol y no de memoria. Comprobable: SC-012.
- **FR-063**: MUST NOT tocarse `docs/ROADMAP.md`, ningún ADR, `docs/USO.md`, `docs/HISTORIAL.md`, `docs/SOURCES.md`, la web, ni el spec, el plan o los informes de un hito anterior. Comprobable: el diff del hito.

#### Lo que no cambia y el cierre

- **FR-070**: MUST NOT cambiar `boe articulos`, `boe analisis` ni los demás verbos de `boe`; `legal-core`, `jurisprudencia` y sus evals; ni `graph check` fuera de FR-031. Comprobable: el diff del hito no toca `skills/legal-core/`, `skills/jurisprudencia/` ni sus evals, y los e2e de los demás verbos siguen en verde.
- **FR-071**: Ninguna corrección ni reparación del run MUST cumplir un umbral cambiando su valor, dejándolo en `decide: false`, sacando sesiones de su `total` o cambiando las evals (ADR 0029). Un rojo de un umbral de `boe-legislacion` es de este hito, y su reparación puede tocar la skill, pero nunca su umbral, la rúbrica, los casos ni la medida del juez (ADR 0038). Si el cierre sale en rojo por un umbral de `jurisprudencia` o de `legal-core`, ninguna reparación las toca ni toca sus umbrales: queda en el informe final para la persona (ADR 0038, punto 8). Comprobable: lo leen los jueces de la revisión final contra el diff de cada reparación.
- **FR-072**: Dentro del run, ninguna sesión MUST abrir una sesión de evals ni un sondeo (ADR 0032): los tests usan sesiones sintéticas. Comprobable: lo leen los jueces de la revisión final contra los tests de FR-089.

#### Controles y Definition of Done

- **FR-080**: `make ci` MUST quedar en verde, con `schema-check`, el drift de `references/` y las reglas del conjunto.
- **FR-081**: Un golden MUST fijar, contra lo grabado:
  - el art. 22 de la Ley 7/1985 (`BOE-A-1985-5392`, `a22`), con nueve versiones, consultado el 1985-04-22, anterior a su primera vigencia (3, con `19850423` en el mensaje); el 2005-06-15, una intermedia (la versión con vigencia `20040101`, de `BOE-A-2003-23103`, con redacción posterior `20151031`, `BOE-A-2015-11723`); el 2008-04-30, el día en que entra en vigor la versión de `BOE-A-2007-18874` (esa versión), y el 2008-04-29, el anterior (la de `20070701`, `BOE-A-2007-10701`, aunque la de `BOE-A-2007-18874` se publicó el 20071031); y el 2019-06-01, posterior a la última (la de `20151031`, con el mismo `texto` y el mismo `hash_texto` que sin `--fecha` y sin redacción posterior);
  - el art. 21 de la Ley 39/2015 (`BOE-A-2015-10565`, `a21`) el 2016-06-01, entre su publicación (20151002) y su entrada en vigor (20161002), que da 3, con `20161002` en el mensaje.
- **FR-082**: MUST haber e2e de la lectura fechada con replay y con `--offline` (la misma lectura fechada, tras una con replay, da el mismo `data` y el mismo `hash`; sin nada guardado que la sirva, 4), y de sus códigos de FR-003, FR-005 y FR-007.
- **FR-083**: Un e2e MUST comprobar que la salida sin `--fecha` de `boe articulo` es la de `main` byte a byte, con `--json` y sin él, en éxito y en fallo.
- **FR-084**: Un e2e MUST comprobar la herramienta MCP `boe_articulo` con `fecha`: anuncia el parámetro, devuelve lo mismo que la orden con la misma entrada, y con una fecha anterior al bloque da el error de FR-003.
- **FR-085**: Un golden MUST fijar el `--describe` de `boe articulo` con la bandera `--fecha` y los campos nuevos.
- **FR-086**: Los tests del grafo MUST comprobar que una lectura fechada emite el `BloqueVersion` de la versión elegida con su `source` (FR-030), no cambia la fila de lecturas del bloque, ni da ni apaga `version-obsoleta`, no recibe `fuente-caducada`, y deja la siguiente lectura sin fecha con el `graph check` que tendría sin ella, también en un bloque sin fila de lecturas (FR-031); que aplicada dos veces no duplica nada; y que con `--no-graph` o `--dry-run` no entrega nada (FR-032).
- **FR-087**: Los tests de la caché MUST comprobar que una lectura fechada no se sirve de la entrada de hoy (FR-020) y que la entrada de hoy no cambia de forma, de clave ni de contenido tras una lectura fechada (FR-021).
- **FR-088**: La sección «Controles de umbral» del plan MUST tener una fila, en `evals:boe-legislacion:<nombre>`, por cada umbral de FR-054, y las de los que ya decidían en `boe-legislacion`, `jurisprudencia` y `legal-core`, que no cambian. MUST NOT tener fila un umbral que no decide. Comprobable: SC-002, que el informe final decide sin modelo con esas filas.
- **FR-089**: Los tests del umbral con sesiones sintéticas MUST cubrir: una sesión de una eval con fecha que no lee su bloque con la fecha de la eval (sin fecha, o con otra fecha) cuenta y da `fallo`; una que lo lee como orden, o como herramienta, y lleva la línea con las fechas esperadas no cuenta; una que lo lee con la fecha y no lleva la línea, o la lleva con otras fechas, cuenta; en la eval (a), una que lee con la fecha y lleva la línea de la fecha pedida y la vigencia pero no la de la redacción posterior, o la lleva con otra fecha o con otra norma, cuenta; en la eval que espera que no había redacción vigente, la línea de FR-043 con su fecha no cuenta y la de FR-041 cuenta; una eval sin fecha no entra en el `total`; una sesión sin medir no cuenta y da su motivo de la ejecución; una sesión de Haiku no entra; un comando esperado sin fecha no casa con una lectura fechada (FR-051). Y un test MUST comprobar los elementos de `umbrales` de las tres skills (FR-056).
- **FR-090**: Un mutante por control. Con cada uno, por separado, algún test MUST fallar, y ningún mutante queda en el árbol: con la versión elegida por `fecha_publicacion`; con la última siempre; con una fecha anterior al bloque que devuelve la primera versión; con la lectura fechada contada como lectura del bloque; y con el `decide` del umbral de FR-054 en falso. La tarea que lo comprueba dice qué test falla con cada uno.
- **FR-091**: `CHANGELOG.md` (*Unreleased*) MUST llevar lo de FR-060, y el job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final).
- **FR-092**: El informe final MUST llevar la nota del cierre: la fase 1 sigue con H9.1 (ADR 0040), y H8 y H11 pueden dar por hecha la redacción de un bloque a una fecha. Es una nota, no trabajo del hito.

### Key Entities

- **Versión de un bloque**: su `id_norma`, su `fecha_publicacion`, su `fecha_vigencia` y su texto, tal como los trae la respuesta del BOE al bloque.
- **Lectura fechada**: norma, bloque y fecha pedida; su resultado es el artículo de la versión elegida con la fecha pedida y, si la hay, la redacción posterior (FR-010), o el 3 de FR-003.
- **Redacción posterior**: fecha de vigencia y norma modificadora de la redacción de hoy, cuando no es la elegida.
- **Esperado de una redacción fechada** de una eval: las fechas de la línea de FR-041, o la de FR-043 (FR-052).
- **Umbral de la redacción fechada** de `boe-legislacion`, por modo (FR-054).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre de `boe-legislacion`, el umbral nuevo (FR-054) lleva `decide: true` y se cumple en los dos modos, 0 de 6 en cada uno; `afirma_lo_no_leido:<modelo>` se cumple, como mucho 3 de 120; `sin_activar` se cumple en los dos modos, 0 de 60; las series de cada eval que decide, con sus citas y sus avisos, pasan; la medida del juez y las duraciones (900 s) se cumplen; ninguna respuesta queda «sin juzgar»; el veredicto es aprobado y `red` está vacío. `jurisprudencia` y `legal-core`, si salen en rojo, no son motivo de este hito y quedan en el informe (FR-071). **Control**: el job de evals de `boe-legislacion` sale en rojo, con el veredicto `fallo`, si una sesión cuenta en el umbral nuevo, si un umbral que decide no se cumple, si una respuesta queda sin juzgar, si una sesión queda sin medir, si una serie que decide no pasa o si llega una petición a la red; el cierre del workflow cuenta cada rojo.
- **SC-002**: El informe final da el umbral nuevo como «comprobado por su control». **Control**: lo decide sin modelo el informe final del workflow, solo para un umbral con fila en «Controles de umbral» del plan (FR-088) que el job publica, cumple y hace fallar el job.
- **SC-003**: Después del run, y fuera de él porque es humano, con la release publicada y en la app de escritorio de Claude con el plugin: «¿qué exigía el artículo 118 de la LCSP para un contrato menor el 1 de junio de 2019?» se responde con la redacción vigente desde el 9 de marzo de 2018, con su cita y su fecha, avisando de que hoy rige la del Real Decreto-ley 3/2020, desde el 6 de febrero de 2020. Lo mide una persona: no tiene control en el run.
- **SC-004**: Los seis casos del golden de FR-081 dan su versión, su código y sus fechas esperados: 0 diferencias. **Control**: el golden falla en `make ci`.
- **SC-005**: La salida sin `--fecha` de `boe articulo` difiere de la de `main` en 0 bytes. **Control**: el e2e de FR-083 falla en `make ci`.
- **SC-006**: Tras una lectura fechada, la fila de lecturas del bloque y lo que da `graph check` sobre él cambian en 0 elementos. **Control**: los tests de FR-086 fallan en `make ci`.
- **SC-007**: Una lectura fechada se sirve de la entrada de hoy 0 veces, y la entrada de hoy cambia en 0 bytes. **Control**: los tests de FR-087 fallan en `make ci`.
- **SC-008**: Con cada uno de los cinco mutantes de FR-090 falla al menos un test: 5 de 5. **Control**: los tests de FR-081 a FR-089, en `make ci`; lo comprueban la tarea de FR-090 y la revisión final.
- **SC-009**: `umbrales` lleva en `boe-legislacion` exactamente los quince de FR-056, once con `decide: true`; seis con seis que deciden en `legal-core`, y quince con once en `jurisprudencia`. **Control**: el test de FR-089 falla en `make ci`.
- **SC-010**: Los tests de FR-089 cubren sus nueve casos de sesiones sintéticas y el de los elementos de `umbrales`: 10 de 10. **Control**: fallan en `make ci`.
- **SC-011**: `skills/boe-legislacion/SKILL.md` tiene menos de 300 líneas, y su tabla de comandos y sus `references/` son las regeneradas. **Control**: `make skills-check`, en `make ci` (FR-048).
- **SC-012**: `README.md`, `CONTRIBUTING.md`, `docs/WORKFLOW.md` y *Unreleased* de `CHANGELOG.md` llevan 0 afirmaciones sobre `boe articulo`, `boe-legislacion`, sus evals, su formato o sus umbrales que el hito deje falsas, y cada cifra, fichero, test y umbral que nombran existe en el árbol. **Control**: la revisión final, por lectura contra el árbol; ningún test de `make ci` ata ese texto al código (FR-062).

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). Lo que gana la skill es citar la redacción de la fecha que importa, y decir, si la hay, que hoy rige otra. Ninguna salida que lee una skill crece con el uso del kit (cientos de normas y miles de bloques consultados en meses): la lectura fechada da un bloque, acotado por la pregunta, y no lee nada de lo acumulado. Las cifras de tamaño son estimaciones sobre lo grabado (el bloque más grande que leen las evals, el `a17` de `BOE-A-2006-20764`, es una respuesta de 37,5 KB con todas sus versiones): en esta sesión no hay binario con `--fecha` que medir.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| El resultado de la lectura fechada (FR-010) | `boe-legislacion`, cuando la pregunta lleva una fecha (FR-040): una vez por bloque que necesita la respuesta, casi siempre uno, como mucho los que lee hoy sin fecha, o dos con un año o un mes sin día (FR-045: el primer y el último día del periodo); y una lectura sin fecha más de un bloque solo si la respuesta dice qué dice hoy (FR-042). Por bloque, como mucho tres salidas, cada una del tamaño de una versión. Cita el texto, da la línea de FR-041 con la fecha pedida y la de vigencia, y la de la redacción posterior si viene. | El de la lectura sin fecha de hoy —el texto de una sola versión, de unos cientos de bytes a unas decenas de KB, y unos 0,8 KB del resto del objeto y del sobre— más unos 100 bytes: la fecha pedida, y la fecha de vigencia y la norma modificadora de la posterior. | Se compone en cada lectura. La redacción posterior solo se da cuando la versión elegida no es la de hoy: deja de darse con una fecha pedida desde la última vigencia, y cambia cuando el BOE añade una versión al bloque. |
| El 3 de una fecha anterior al bloque, con su mensaje (FR-003) | `boe-legislacion`, una vez por lectura fechada que lo da, para la línea de FR-043. | Una línea de mensaje. | Solo con una fecha pedida anterior a la primera vigencia del bloque. |
| La línea de la redacción fechada, la de la posterior y la de «sin redacción vigente» en la respuesta (FR-041, FR-043) | La persona, una vez por respuesta y bloque leído con fecha, o una por cada una de las dos lecturas de FR-045 con un año o un mes sin día: como mucho dos de FR-041 o de FR-043 y una de la posterior por bloque. El job, que las comprueba sin modelo. | Una línea cada una. | Solo en respuestas a preguntas con fecha (FR-046); no se acumulan. |
| El `BloqueVersion` de la lectura fechada en el grafo (FR-030) | Ninguna skill lo pide por la lectura fechada: no cuenta como lectura (FR-031), así que `graph check` no da nada por él. | Un nodo por versión distinta leída con fecha: como mucho las versiones de los bloques leídos así (nueve en el art. 22 de la LRBRL). | No da señales: ni `version-obsoleta` ni `fuente-caducada` (FR-031). |
| Lo guardado en la caché por una lectura fechada (FR-022) | Solo el binario, para servir la misma lectura fechada con `--offline`. Ninguna skill lo lee. | Lo fija el plan. | No da señales. |
| El umbral de FR-054 en `umbrales` y su motivo de `fallo` (FR-055) | El job, una vez, que decide con él el veredicto; el informe final del workflow, que lo lee sin modelo (SC-002); la persona y la reparación del cierre. | Dos elementos más en `boe-legislacion`, de unos 300 bytes cada uno. Cada motivo, unos 90 bytes y unos 55 por sesión que nombra, acotado por las 6 sesiones del modo: menos de 0,5 KB. | Cada job lo mide de nuevo; el `cumple: false` y su motivo dejan de darse en el primer job en que ninguna sesión cuenta. |
| `SKILL.md` de `boe-legislacion` (FR-040 a FR-049) | El agente que activa la skill, una vez por conversación. | Menos de 300 líneas (299 en `main`). | No da señales. |

## Fuera de alcance

Del hito, literal:

- «el diff entre versiones y `graph history` (backlog · profundidad);»
- «`--fecha` en `articulos` y en `analisis`;»
- «decir si la norma estaba derogada en la fecha pedida: los metadatos traen `fecha_derogacion` (la de la Ley 30/1992 es 20210402, y no la de entrada en vigor de las leyes que la derogaron), pero los avisos son los de hoy, como dice «La salida»;»
- «la forma de citar una redacción que no es la vigente, más allá de lo que fije el plan para `boe-legislacion`;»
- «una clase nueva del juez, que pide rúbrica, casos etiquetados y medida (ADR 0037);»
- «cambiar el modelo que decide, las repeticiones o la regla de 2 de 3 por eval (ADR 0016 y 0031);»
- «extraer `internal/cli`.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Un verbo o una salida que liste todas las versiones de un bloque, o que dé su texto sin elegir una por fecha.
- Usar la lectura fechada para decir qué cambió tras un `version-obsoleta` («Redacción modificada» sigue como está, FR-046).
- `--fecha` en `legal-core`, en `plazos` o en `jurisprudencia`, y cualquier cambio de sus evals.
- Más evals que las dos de FR-050, y cualquier cambio de `.github/workflows/evals.yml` (FR-059).
- Una fila nueva de `docs/SOURCES.md`, otra fuente o grabaciones nuevas del BOE (FR-006).
- Cambiar las `instructions` del servidor MCP.
- La viñeta de H20 de «Estado del repositorio» de `CLAUDE.md` y su frase del siguiente hito: las escribe la persona al fusionar.
- Editar el spec, el plan o los informes de un hito anterior; un ADR nuevo; y la release, que decide una persona.

## Assumptions

- **Lo que H4, H7, H7.1, H7.3 y H21 ya definen vale sin volver a escribirse** («Relación con hitos anteriores», «Se queda»).
- **Lo grabado de las evals.** El hito dice que `testdata/evals/` ya tiene los dos bloques; en esta sesión, `testdata/evals/boe.legislacion-consolidada/` no tiene `a1-30` de `BOE-A-2017-12902` ni `a21` de `BOE-A-2015-10565`, que están en lo grabado de H4, `internal/source/boe/testdata/boe.legislacion-consolidada/`. La caché de las evals sale de la unión de lo grabado de H4, H5 y H9 (`GrabacionesDeH4`, `GrabacionesDeH5` y `GrabacionesDeH9`, `internal/evals/grabaciones.go`), así que las dos evals se sirven de lo grabado sin grabar nada (FR-053).
- **Las versiones de lo grabado**, leídas con `grep` en esta sesión: el `a22` de `BOE-A-1985-5392` trae nueve, con `fecha_vigencia` 19850423, 19960609, 19970416, 19990512, 20040101, 20070701, 20080430, 20080627 y 20151031, y la de 20080430 (`BOE-A-2007-18874`) se publicó el 20071031; el `a1-30` de `BOE-A-2017-12902`, dos, de `BOE-A-2017-12902` (publicada el 20171109, vigente desde el 20180309) y de `BOE-A-2020-1651` (20200205 y 20200206); el `a21` de `BOE-A-2015-10565`, una, publicada el 20151002 y vigente desde el 20161002. De ahí salen las fechas del golden de FR-081.
- **Las cifras de hoy de `umbrales`**, del último cierre versionado (`specs/023-h9-internal-core-plazos/gates/evals/`, leído con `jq` en esta sesión): `boe-legislacion`, trece elementos, nueve que deciden, con `sin_activar:<modelo>:<modo>` de 54 por modo y `afirma_lo_no_leido:<modelo>` de 108 con umbral 0,028; `legal-core`, seis, seis que deciden; `jurisprudencia`, quince, once que deciden; `repeticiones` 3 y `red` vacío. Con las dos evals nuevas, 60 sesiones por modo en `sin_activar` y 120 respuestas en `afirma_lo_no_leido` (FR-058, SC-001).
- **La concurrencia del job** está a cinco para `boe-legislacion` en `.github/workflows/evals.yml` de `main`, con objetivo de duración de 900 s (leído en esta sesión). El hito da el peor caso a cinco con veintiuna evals en 18 977 s y con cuatro más en 20 977 s; el comentario del fichero da 17 329 s. Ninguna cifra se ha medido en esta sesión: el control es `TestDefinicionDelJob` (FR-050).
- **«Las citas, los avisos»** de la Aceptación se leen como la regla de 2 de 3 de la serie de cada eval que decide, con sus claves `citas` y `avisos` (FR-057): no son elementos de `umbrales` hoy ni lo son con el hito.
- **La forma fija en la respuesta** se comprueba como la `⚠ REDACCIÓN MODIFICADA:` de H7.2: por su texto fijo y sus fechas, sin el juicio de ningún modelo (FR-048, FR-052).
- **La lectura fechada pide lo mismo que la de hoy** (FR-006): el bloque, con todas sus versiones, y los metadatos de la norma, de los que salen los avisos y el `url_eli`.
- **Los nombres técnicos que aparecen** (`--fecha`, `fecha`, `boe_articulo`, `umbrales`, `sin_activar`, `afirma_lo_no_leido`, `TestDefinicionDelJob`, `make skills-check`, `schemas/eval.yaml.json`, `articuloDelBloque`, `Consolidado.Lecturas`) existen en el repositorio o los fija el hito. Quedan para el plan los nombres de los campos nuevos de `data`, la forma del mensaje de FR-003, la fila de FR-004, cómo se guarda y cómo se emite la lectura fechada, las claves de las evals, el nombre y el orden del umbral nuevo, el texto de las formas fijas y los comandos, citas y avisos de las dos evals.
