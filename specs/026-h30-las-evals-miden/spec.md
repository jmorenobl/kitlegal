# Feature Specification: H30 · Las evals miden solo lo que cambia: dependencias de cada eval y resultados por huella

**Feature Branch**: `026-h30-las-evals-miden`

**Created**: 2026-10-22

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H30-las-evals-miden-solo.md`, la única entrada humana del run
(ADR 0018).

## Contexto

El job de evals (`.github/workflows/evals.yml`) decide si mide mirando si la propuesta de cambio toca alguna ruta de una
lista fija (`skills/`, `evals/`, `data/`, `cmd/`, `internal/` salvo `internal/workflow/`, `skills.go`, `scripts/evals.sh`,
`scripts/evals-sesion.sh`, el propio fichero del job, `schemas/eval.yaml.json` y `Makefile`), y si la toca, o si alguien
pone la etiqueta `evals`, mide las tres skills enteras —`boe-legislacion` con 23 evals, `jurisprudencia` con 10 y
`legal-core` con 12— en los dos modos, con cada eval repetida tres veces por modelo y modo. El tope del trabajo de
`boe-legislacion` es de 352 minutos para un peor caso calculado de 316 (comentarios de `evals.yml`). Una propuesta que
solo cambia una frase de `skills/jurisprudencia/SKILL.md` abre las mismas sesiones que una que cambia el kernel. Con las
materias como paquetes sobre las mismas skills de capacidad (ADR 0044), cada paquete nuevo multiplicaría ese coste en
cada propuesta.

Lo que gana cada skill es que un cambio suyo se mide sin esperar ni gastar la medida de las demás, y que lo que no cambia
conserva el resultado que ya tiene en `main`. Ninguna skill consume las salidas de este hito: las consumen quien revisa
una propuesta de cambio, el cierre del workflow (`scripts/workflow/cierre.sh` y `scripts/workflow/serie.sh`, que este hito
no toca) y quien etiqueta una release.

## Clarifications

### Session 2026-10-22

- Q: Cuando una eval que la propuesta no elige (FR-006) no tiene ningún resultado reutilizable de una medida de `main` con su huella (eval nueva, registro caducado, informe ilegible, plataforma que no responde), ¿qué hace con ella la ejecución con reutilización (apertura de la propuesta o etiqueta `evals`)? → A: Se mide en esa misma ejecución, abriendo sesiones de su skill aunque la propuesta no toque esa skill, solo las de esa eval y con la concurrencia de hoy (FR-020); el informe dice «sin resultado reutilizable de `main`» y cuenta para el veredicto como cualquier eval medida (auto: criterio a; fuente: `docs/roadmap/done/H30-las-evals-miden-solo.md`, entrega (4) y (5), Alcance y Aceptación; FR-013, FR-016, FR-017, FR-022)
- Q: ¿Debe H30 definir cómo se identifica la materia de una eval y de una referencia, para que la referencia de una materia solo sea dependencia de las evals de esa materia, o basta la granularidad por skill de hoy, en la que cada eval depende de todas las referencias de sus skills? → A: Basta la granularidad por skill: H30 no define marca de materia ni punto de extensión; el formato del paquete de materia y la asignación llegan con H32 (auto: conservadora; criterio d; fuente: `docs/roadmap/done/H30-las-evals-miden-solo.md` (Objetivo, Controles); `docs/roadmap/todo/040-H32-arquitectura-de-skills-una-entrada-generica.md`; constitución, Criterio de decisión autónoma 2 y 4; principio X)
- Q: En el umbral en tasa de una clase del juez (ADR 0038), ¿cómo cuentan las respuestas de las evals reutilizadas en una ejecución con reutilización? → A: No cuentan: el umbral `<clase>:<modelo>` se aplica solo a las respuestas medidas, con el límite de marcadas que da la regla del ADR 0038 (punto 3) para ese número de respuestas medidas y la tasa de referencia de Q4; con todas las respuestas de la skill, el umbral de hoy, que no cambia; el informe da aparte las marcadas y el total de las reutilizadas, sin decidir (auto: criterio a; fuente: entrega (4) y (6) del hito; `scripts/workflow/serie.sh`; `docs/ADR/0038`, puntos 3 y 4; constitución, «Umbrales que se cumplen»; FR-015, FR-023, FR-024, FR-026)
- Q: ¿Qué tasa de referencia fija el ADR para cada skill con juez, con la que el test de SC-010 comprueba el criterio de 1/50? → A: La mayor tasa con la que el umbral vigente de la ejecución completa de cada skill cumple aún 1/50, derivada de ese umbral y de las respuestas de la ejecución completa y redondeada hacia abajo; el ADR escribe la derivación y, como su resultado, las cifras de hoy: `boe-legislacion`, 3 marcadas de 120, ≈0,85 %; `jurisprudencia`, 1 de 60, ≈0,36 %; si una persona estrecha el umbral, la tasa de referencia que se deriva baja con él; el ADR deja escrito que la tasa acumulada de la serie de `boe-legislacion` (18 de 1.424, el 1,26 %) está por encima, y subir el umbral pediría otro ADR (auto: criterio c; fuente: FR-024, SC-010, SC-013, FR-026; `docs/ADR/0038`, puntos 3 a 5; `evals/*/juez/clases.yaml`; `specs/*/gates/serie.jsonl`; `evidencias/adr-0038/serie-previa.jsonl`)
- Q: ¿Qué ejecuciones de una propuesta de cambio fusionada en `main` cuentan como medida de `main` cuyos resultados se pueden reutilizar por su huella? → A: Solo las ejecuciones sobre el commit que era la cabeza de la propuesta al fusionarla (el `head.sha` que da la plataforma de la propuesta fusionada) que cumplen además FR-010; si esa cabeza no tiene ejecución, la propuesta no aporta medida de `main` y sus evals se miden cuando las necesite otra propuesta (auto: criterio a; fuente: `docs/roadmap/done/H30-las-evals-miden-solo.md`, entrega (4); FR-010, FR-013, «Fuera de alcance»; `docs/ROADMAP.md`, §1)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Una propuesta mide las evals que ha podido cambiar, y nada más (Priority: P1)

Quien abre una propuesta que toca una skill quiere que el job mida las evals cuyo resultado puede haber cambiado, y no las
de las demás skills. El job calcula, para cada eval, el conjunto de rutas del repositorio de las que depende, y elige las
evals cuyo conjunto toca la propuesta; lo común, la descripción de cualquier skill y lo que no se sabe asignar las eligen
todas.

**Why this priority**: es el objetivo del hito; sin la elección no hay nada que reutilizar ni que ahorrar.

**Independent Test**: con el árbol del repositorio y una lista de rutas cambiadas, el cálculo da las evals elegidas; se
prueba en `make ci` sin red ni sesiones.

**Acceptance Scenarios**:

1. **Dado** una propuesta que solo cambia el cuerpo de `skills/jurisprudencia/SKILL.md`, **Cuando** el job decide qué
   mide, **Entonces** elige las 10 evals de `jurisprudencia` y ninguna de `boe-legislacion` ni de `legal-core`, y no abre
   ninguna sesión de `boe-legislacion` ni de `legal-core` mientras tengan una medida de `main` reutilizable (FR-012,
   FR-014).
2. **Dado** una propuesta que solo cambia una referencia de `legal-core` (`skills/legal-core/references/recursos_y_plazos.md`),
   **Cuando** el job decide qué mide, **Entonces** elige las 12 evals de `legal-core` y ninguna de `boe-legislacion` ni
   de `jurisprudencia`.
3. **Dado** una propuesta que solo cambia el cuerpo de `skills/boe-legislacion/SKILL.md`, **Cuando** el job decide qué
   mide, **Entonces** elige las 23 de `boe-legislacion` y también las 12 de `legal-core`, que la nombra en su `SKILL.md`,
   y las 10 de `jurisprudencia`, que también la nombra.
4. **Dado** una propuesta que cambia un fichero de `internal/cli`, **Cuando** el job decide qué mide, **Entonces** elige
   las 45 evals.
5. **Dado** una propuesta que solo cambia la descripción de una skill cualquiera, **Cuando** el job decide qué mide,
   **Entonces** elige las 45 evals.
6. **Dado** una propuesta que cambia una ruta de las que miden las evals que el cálculo no asigna a ninguna eval ni a lo
   común, **Cuando** el job decide qué mide, **Entonces** elige las 45 evals y el informe nombra esa ruta.
7. **Dado** una propuesta que solo cambia rutas que las evals no miden (`docs/`, `web/`, `specs/`), **Cuando** se abre,
   **Entonces** no elige ninguna eval y el job no abre ninguna sesión.

---

### User Story 2 - Lo que no cambia reutiliza el resultado de `main` por su huella (Priority: P1)

El resultado de cada eval se guarda con su huella. Cuando el job tiene que dar el resultado de una eval que la propuesta
no elige y una medida de `main` tiene ya uno con la misma huella, lo reutiliza —pase o no— sin abrir sesión, y su informe lo dice. El cierre del
workflow, que pone la etiqueta `evals`, obtiene así un informe completo de cada skill sin volver a medir lo que no ha
cambiado.

**Why this priority**: sin reutilización, la etiqueta `evals` volvería a medirlo todo o dejaría sin resultado lo no
elegido.

**Independent Test**: con informes de medidas de `main` como datos de prueba y el árbol de una propuesta, el job decide
qué reutiliza y qué mide; se prueba en `make ci` sin red ni sesiones.

**Acceptance Scenarios**:

1. **Dado** una propuesta que no toca nada de lo que las evals miden y una medida de `main` recuperable con la huella de
   cada una de las 45 evals, **Cuando** se pone la etiqueta `evals`, **Entonces** el job no abre ninguna sesión, y el
   informe de cada skill da sus evals como reutilizadas, cada una con su huella y la medida de la que viene.
2. **Dado** una eval cuyo resultado en esa medida de `main` era un fallo, **Cuando** se reutiliza, **Entonces** el
   veredicto del informe es el mismo que si se hubiera medido con ese resultado.
3. **Dado** una eval no elegida cuya única medida de `main` con su huella ya no se puede leer (registro caducado o
   informe ilegible), **Cuando** el job da su resultado, **Entonces** se mide en esa ejecución, con el motivo «sin
   resultado reutilizable de `main`» (FR-014).
4. **Dado** un informe con evals medidas y reutilizadas, **Cuando** `scripts/workflow/serie.sh fila` (el de `main`, sin
   cambios) escribe la fila de la serie, **Entonces** sale con 0 y la fila cuenta solo las respuestas medidas.

---

### User Story 3 - La release exige la suite completa del commit etiquetado (Priority: P1)

Quien etiqueta una versión quiere que lo publicado tenga detrás una medida completa de las evals sobre ese mismo commit,
sin resultados reutilizados. Si no la tiene, la release falla antes de publicar nada.

**Why this priority**: la reutilización hace que ninguna propuesta mida ya la suite entera; la release es el punto en el
que se exige.

**Independent Test**: un test de `make ci` lee `release.yml` y sale en rojo si la release no exige esa ejecución.

**Acceptance Scenarios**:

1. **Dado** una etiqueta `v*` sobre un commit sin ninguna ejecución completa de la suite, **Cuando** corre la release,
   **Entonces** falla antes de publicar la release, el cask, el manifiesto del bucket o el catálogo, y su mensaje nombra
   el commit y cómo lanzar la suite completa.
2. **Dado** una etiqueta sobre un commit cuya única ejecución tiene alguna eval reutilizada, o algún trabajo de skill en
   rojo, o falta la de alguna skill, **Cuando** corre la release, **Entonces** falla igual.
3. **Dado** una etiqueta sobre un commit con una ejecución completa, en verde y sin nada reutilizado, **Cuando** corre la
   release, **Entonces** publica como hoy.

---

### User Story 4 - La suite completa a petición (Priority: P2)

Una persona puede pedir la suite entera, sin reutilizar nada, con la etiqueta `evals-completa` o con el despacho manual;
es la ejecución que exige la release.

**Why this priority**: es la vía para obtener lo que exige la release y para volver a medir un resultado reutilizado.

**Independent Test**: un test de `make ci` sobre la definición del job comprueba qué disparador mide todo y cuál
reutiliza.

**Acceptance Scenarios**:

1. **Dado** una propuesta, **Cuando** se pone la etiqueta `evals-completa`, **Entonces** el job mide las 45 evals y el
   informe no tiene ninguna reutilizada.
2. **Dado** cualquier rama o etiqueta, **Cuando** una persona lanza el despacho manual sin `medir_al_juez`, **Entonces**
   el job mide las 45 evals sin reutilizar ninguna.

---

### User Story 5 - Un ADR decide cómo cuenta lo reutilizado en el umbral en tasa y en la serie (Priority: P2)

El umbral en tasa de una clase del juez se fijó para las respuestas de una ejecución completa (ADR 0038: como mucho 3
marcadas de 108 en `boe-legislacion`). Con menos respuestas medidas, la misma proporción pone en rojo sin motivo más de
1 de cada 50 ejecuciones. Un ADR decide cómo cuenta lo reutilizado en ese umbral y en la serie de cierres, y el job lo
aplica.

**Why this priority**: sin esa decisión, la reutilización rompe el criterio del ADR 0038.

**Independent Test**: un test de `make ci` comprueba el criterio con la regla del ADR para cada número de respuestas
medidas posible.

**Acceptance Scenarios**:

1. **Dado** la regla del ADR y la tasa de referencia de cada skill con juez, **Cuando** se mide con cualquier número de
   respuestas del modelo que decide entre 1 y todas las de la skill, **Entonces** la probabilidad de que la ejecución
   salga en rojo sin motivo es como mucho 1/50.

---

### Edge Cases

- Una eval nueva no tiene ninguna medida de `main` con su huella: se mide (FR-013).
- Una eval retirada no aparece en ningún informe; su resultado anterior no se reutiliza para otra.
- Una propuesta que añade, retira o renombra un directorio de `skills/` cambia las descripciones que compiten por la
  activación: elige las 45 evals (FR-004).
- Un fichero renombrado cuenta como cambio de su ruta de antes y de la de después (FR-007).
- Dos medidas de `main` con la misma huella y resultados distintos: se reutiliza la más reciente (FR-012).
- Una medida de `main` que a su vez reutilizó una eval: el resultado reutilizable es el de la medida en que se abrieron
  sus sesiones (FR-011), mientras esa medida tenga menos de 90 días; si no, la eval se mide (FR-012).
- Una eval que en la medida de `main` quedó con alguna sesión sin medir o alguna respuesta sin juzgar no es reutilizable:
  se mide (FR-012).
- El despacho con `prueba_de_red` o la etiqueta `evals-prueba-de-red`: suite completa y la sesión de red, como hoy
  (FR-018).
- La etiqueta `evals` y la `evals-completa` sobre el mismo commit a la vez: la completa no se queda sin medir (FR-021).
- Cualquier informe, registro o respuesta de la plataforma manipulado a mano o malformado lo cubre la regla genérica:
  el job no lo reutiliza y mide la eval, o el paso falla como defecto `inesperado` (ADR 0023), sin más casos.

## Requirements *(mandatory)*

### Functional Requirements

#### El conjunto de dependencias de cada eval

- **FR-001**: Para cada eval (`evals/<skill>/NN-*.yaml`), el cálculo obtiene su conjunto de dependencias del propio
  repositorio, sin ninguna lista escrita a mano por eval ni por skill. El conjunto lo forman:
  (a) el fichero de la eval;
  (b) las skills de la eval: la skill que mide, las que nombra su `SKILL.md` —el nombre de otro directorio de `skills/`
  en cualquier parte del fichero— y, de cada una de ellas, las que a su vez nombra; de cada una, su `SKILL.md` y sus
  `references/`;
  (c) los `data/*.yaml` de los que se generan esas referencias, los que dice la cabecera «generado desde» de cada una;
  (d) de cada applet de la tabla de comandos de cada skill de la eval, sean o no de los `comandos:` que espera la eval,
  sus paquetes Go: los del repositorio que importa el código del applet en la composición (hoy su fichero de
  `internal/app`) fuera de lo común (FR-002), y los que `go list -deps` da desde ellos; de cada uno, los ficheros que
  compila y los que empotra, y la versión de cada módulo externo del que dependen, la que fijan `go.mod` y `go.sum`. Lo
  común no entra por (d): el paquete raíz, que empotra las skills, llega a un conjunto solo por (b). Los `_test.go` y
  `testdata/` de esos paquetes no entran en el binario: no son de ningún conjunto ni rutas que las evals miden (FR-004),
  salvo las grabaciones de (e);
  (e) las grabaciones que el arnés sirve a las sesiones de la eval;
  (f) si su skill tiene juez, la carpeta del juez de su skill —rúbrica, casos, medida, clases y esquema— y, en cualquier
  skill, los demás ficheros de `evals/<skill>/` que no son una eval.
  *Criterio*: un test de `make ci` sobre el árbol del repositorio comprueba, para una eval de cada skill, que el conjunto
  contiene cada una de las seis partes —para una de `legal-core`, `skills/boe-legislacion/SKILL.md`,
  `data/plazos.yaml` y un fichero de `internal/core/plazos`; para una de `jurisprudencia`,
  `skills/boe-legislacion/SKILL.md` y un fichero de `internal/core/cita`— y que no contiene el fichero de otra eval ni,
  en el de una de `jurisprudencia`, `skills/legal-core/SKILL.md`.

- **FR-002**: Lo común elige todas las evals de todas las skills: el kernel y la composición del binario (hoy `cmd/` e
  `internal/app`), `internal/cli`, la red (`internal/httpx`), la caché (`internal/cache`), el servidor MCP
  (`internal/mcp`), el arnés de las evals (`internal/evals`, los guiones `scripts/evals*.sh` que ejecuta el job y los
  esquemas de las evals y del juez), el job (`.github/workflows/evals.yml`, que fija el modelo que decide, los
  informativos, las repeticiones y las versiones de Claude Code), lo que instala el binario y las skills en el job
  (`skills.go` y `Makefile`), la versión de Go que fija `go.mod` (`go` y `toolchain`) y la de cada módulo externo del que dependen los
  paquetes de lo común. Que un fichero común aparezca también en el conjunto de FR-001 no cambia nada: lo elige
  todo. *Criterio*: mutantes de `make ci`: cambiar un fichero de `internal/cli` elige las 45 evals, y cambiar en
  `go.mod` la versión de `modernc.org/sqlite`, de la caché, también.

- **FR-003**: La descripción de cada skill —el campo `description` de la cabecera de su `SKILL.md`— compite con las demás
  por la activación que mide cada eval: cambiar la de cualquier skill elige todas las evals de todas las skills. Cambiar
  solo el cuerpo de un `SKILL.md` no es cambiar su descripción. *Criterio*: mutante de `make ci`: cambiar solo la
  descripción de `legal-core` elige las 45; cambiar solo el cuerpo de `skills/jurisprudencia/SKILL.md` elige las 10 de
  `jurisprudencia` y ninguna más.

- **FR-004**: Las rutas que las evals miden son las que hoy lanzan el job al abrir una propuesta (las de «Contexto»), más
  las grabaciones de FR-001(e), los guiones y esquemas del arnés de FR-002, y `go.mod` y `go.sum`. Un cambio de `go.mod`
  o `go.sum` elige las evals de cuyo conjunto cambia la versión de algún módulo, y todas si cambia la de un módulo de lo
  común o la versión de Go, o si no cambia la de ningún módulo de un conjunto ni de lo común. Una ruta cambiada de ese conjunto que no
  está en el conjunto de ninguna eval ni en lo común —también un directorio de `skills/` o de `evals/` que se añade, se
  retira o se renombra— es una ruta sin asignar y elige todas las evals. Una ruta cambiada fuera de ese conjunto no elige
  ninguna. *Criterio*: mutantes de `make ci`: un paquete nuevo bajo `internal/` del que no depende ningún applet ni lo
  común elige las 45 y el resultado nombra la ruta; un cambio solo en `docs/` no elige ninguna.

- **FR-005**: Si el cálculo no puede obtener el conjunto de una eval —una skill nombrada o un applet de la tabla de
  comandos que no existe, una referencia sin su cabecera «generado desde», un paquete que `go list` no resuelve—, `make
  ci` falla nombrando la eval y la causa. Si le pasa al job, elige todas las evals y el informe dice por qué.
  *Criterio*: test de `make ci` que recorre todas las evals del repositorio (hoy 45) y falla si alguna no tiene un
  conjunto calculable o tiene uno vacío; y un mutante con una skill que nombra un applet inexistente hace fallar el
  test.

#### La elección

- **FR-006**: Las evals elegidas por una propuesta de cambio son las de cuyo conjunto cambia alguna ruta, más todas si
  cambia algo de lo común (FR-002), una descripción (FR-003) o una ruta sin asignar (FR-004). La elección es por eval, no
  por skill: cambiar el fichero de una eval elige esa eval y ninguna otra.
  *Criterio*: mutante de `make ci`: cambiar `evals/jurisprudencia/03-por-materia.yaml` elige esa eval y ninguna más.

- **FR-007**: Las rutas cambiadas de una propuesta son las que la plataforma da como ficheros de la propuesta respecto de
  su base; un fichero renombrado cuenta con su ruta de antes y con la de después.
  *Criterio*: test de `make ci`: renombrar una referencia de `legal-core` elige las evals de `legal-core`.

- **FR-008**: El resultado de la elección, con el motivo de cada eval elegida —las rutas cambiadas de su conjunto que la
  eligen, o «común», «descripción» o «sin asignar» con la ruta—, es determinista: el mismo árbol y las mismas rutas
  cambiadas dan la misma elección, en el mismo orden. *Criterio*: golden de la matriz en `make ci` (FR-019).

#### La huella

- **FR-009**: La huella de una eval es un valor que cambia si y solo si cambia algo que la elige: el modelo que decide, la
  versión de Claude Code de las sesiones, la huella del juez si su skill lo tiene —modelo del juez, versión de Claude Code
  de sus votos y el contenido de su rúbrica, sus casos y su medida—, el contenido de cada ruta de su conjunto (FR-001),
  de lo común (FR-002), de la descripción de cada skill (FR-003) y de las rutas sin asignar (FR-004), y las versiones de
  Go y de los módulos externos de su conjunto y de lo común. No depende del
  runner, de la fecha, del orden en que se recorren los ficheros ni de la ruta del directorio de trabajo.
  *Criterio*: test de `make ci`: para cada clase de ruta de FR-001 a FR-004, cambiar su contenido cambia la huella de las
  evals que elige y no la de las demás; cambiar en `go.mod` la versión de un módulo de lo común cambia la de todas;
  cambiar un fichero de `docs/` no cambia ninguna; calcular dos veces la huella del
  mismo árbol desde dos directorios distintos da lo mismo.

#### La reutilización

- **FR-010**: Una medida de `main` es una ejecución del job de evals sobre la cabeza de una propuesta de cambio fusionada
  en `main` —el commit que era su cabeza al fusionarla, el `head.sha` que la plataforma da de la propuesta fusionada—
  en la que el trabajo de cada skill terminó, en verde o en rojo, con un informe legible. No lo son la de una propuesta
  sin fusionar, una cancelada, una a la que le falta el informe de alguna skill, ni una sobre otro commit de la propuesta
  fusionada. Si la cabeza no tiene ejecución, la propuesta no aporta ninguna medida de `main`.
  *Criterio*: test de `make ci` con respuestas grabadas de la plataforma: una ejecución cancelada y una de una propuesta
  cerrada sin fusionar no se usan.

- **FR-011**: El resultado de una eval en una medida es todo lo que su informe da de ella: sus series por modelo y modo,
  también las de los modelos informativos, su veredicto y lo que aportan sus respuestas a cada umbral del informe. Se
  guarda con su huella. Si esa medida lo reutilizó a su vez, el resultado y su procedencia son los de la medida en que se
  abrieron sus sesiones.

- **FR-012**: En una ejecución con reutilización (FR-016, FR-017), las evals elegidas (FR-006) se miden y no se buscan.
  Cada eval no elegida reutiliza el resultado de la medida de `main` más reciente con su misma huella, pase o no, sin
  abrir sesión, si ese resultado se puede leer, está completo —ninguna sesión de sus series quedó sin medir y, con
  juez, ninguna respuesta del modelo que decide quedó sin juzgar— y la ejecución en que se abrieron sus sesiones (su
  procedencia, FR-011) tiene menos de 90 días, la vida de su registro en la plataforma. La búsqueda recorre como mucho
  las 10 medidas de `main` más recientes y para en cuanto cada eval no elegida tiene su resultado; lo que queda sin él
  se mide (FR-013).
  *Criterio*: tests de `make ci`: con dos medidas de la misma huella se reutiliza la más reciente; una eval con una
  sesión sin medir en su medida se mide; una eval cuya huella solo está en la undécima medida más reciente se mide; una
  eval que la medida más reciente reutilizó de una ejecución de hace 90 días o más se mide.

- **FR-013**: Una eval sin ninguna medida de `main` con su huella, entre las que recorre la búsqueda (FR-012), cuyo
  resultado se pueda leer y esté completo y cuya procedencia tenga menos de 90 días —eval nueva, huella cambiada,
  registro caducado o borrado, procedencia de 90 días o más, informe ilegible, plataforma que no responde— se mide.

- **FR-014**: Una eval **no elegida** por la propuesta (FR-006) que no tiene un resultado reutilizable de `main` con su
  huella (FR-013) se mide en esa misma ejecución, con ese motivo, «sin resultado reutilizable de `main`» (FR-022): abre
  solo las sesiones de esa eval, dentro del trabajo `evals (<skill>)` de su skill, aunque la propuesta no toque esa
  skill, con la concurrencia de hoy (FR-020), y cuenta para el veredicto como cualquier eval medida. Una propuesta que
  no elige ninguna eval no corre ningún trabajo (FR-016); la etiqueta `evals` sí lo corre (FR-017). El resultado medido
  pasa al informe de esa ejecución y, en cuanto la propuesta se fusiona, vuelve a ser reutilizable (FR-011).
  *Criterio*: test de `make ci` con respuestas grabadas de la plataforma sin medida de `main` para una eval no elegida:
  la decisión del job la planifica con ese motivo y no planifica ninguna otra sesión.

- **FR-015**: El veredicto del informe de una skill con evals reutilizadas es el mismo que daría con esas evals medidas y
  con exactamente el resultado reutilizado: una serie que decide y no pasa lo pone en fallo, y un umbral con 0 se
  incumple con las respuestas reutilizadas que lo incumplían, y el informe da por separado lo medido y lo reutilizado de
  cada umbral. La duración de las sesiones y de los votos es la de lo medido. La única excepción es el umbral en tasa de
  una clase del juez, que solo cuenta las respuestas medidas (FR-024); de lo reutilizado, el informe da aparte las
  marcadas y el total, sin decidir.
  *Criterio*: test de `make ci`: un informe con una eval reutilizada con una serie que decide en fallo, o con una
  respuesta sin la skill activada, sale en fallo.

#### Los disparadores y la matriz

- **FR-016**: Al abrir o reabrir una propuesta de cambio, si elige alguna eval, el job da el resultado de todas las evals
  de cada skill, midiendo lo que no puede reutilizar (FR-012, FR-013, FR-014); si no elige ninguna, no abre ninguna
  sesión ni corre ningún trabajo de evals, como hoy con una propuesta que no toca lo que las evals miden.

- **FR-017**: La etiqueta `evals` —la que pone el cierre del workflow— hace lo mismo que la apertura con cualquier
  propuesta, también una que no elige ninguna eval: mide lo elegido y reutiliza lo demás.
  *Criterio*: test de `make ci` sobre la decisión del job (SC-007).

- **FR-018**: La etiqueta `evals-completa` y el despacho manual sin `medir_al_juez` miden la suite completa: todas las
  evals de cada skill, sin reutilizar ninguna. La etiqueta `evals-prueba-de-red` y el despacho con `prueba_de_red` miden
  también la suite completa, con la sesión de prueba de red de hoy. La etiqueta `evals-medir-juez` y la entrada
  `medir_al_juez` siguen lanzando la medida del juez y no las evals, como hoy.

- **FR-019**: El job construye su matriz con las evals que tiene que medir. Cada skill tiene en cada ejecución que mide o
  reutiliza su trabajo `evals (<skill>)`, con el nombre de hoy y su informe, aunque no abra ninguna sesión porque lo
  reutiliza todo; una skill sin nada que medir abre 0 sesiones. *Criterio*: golden de la matriz en `make ci` para un
  conjunto fijo de propuestas —cada uno de los mutantes de FR-002 a FR-007, una sin nada elegido y la suite completa—:
  qué evals mide cada skill y cuáles reutiliza.

- **FR-020**: Ninguna ejecución abre más sesiones a la vez que hoy: como mucho un trabajo de evals por skill, cada uno con
  su concurrencia de hoy (5 en `boe-legislacion`, 4 en `jurisprudencia`, 2 en `legal-core`), 11 sesiones del modelo
  que decide o informativas a la vez entre las tres. Las repeticiones, el umbral de 2 de 3 por serie, los modelos y las
  versiones de Claude Code no cambian. *Criterio*: el test de `make ci` sobre la definición del job sale en rojo si la
  matriz admite más de un trabajo por skill a la vez o una concurrencia mayor que esas.

- **FR-021**: La regla de una tanda por commit sigue: una ejecución no mide si otra anterior sin terminar ya mide el
  mismo commit con al menos lo mismo. Una de la suite completa no la sustituye una con reutilización en curso; una con
  reutilización sí la puede sustituir una de la suite completa en curso. *Criterio*: test de `make ci` de la decisión
  de la tanda con las dos combinaciones.

#### El informe

- **FR-022**: El informe de cada skill separa lo medido de lo reutilizado: de cada eval dice si se midió o se reutilizó y
  su huella; de cada medida, por qué —elegida, con su motivo de FR-008; sin resultado reutilizable de `main`; o suite
  completa—; y de cada reutilizada, la medida de la que viene (ejecución, trabajo, commit, propuesta y fecha). Da además
  cuántas evals midió, cuántas reutilizó y cuántas sesiones abrió.

- **FR-023**: Lo reutilizado no entra en los campos que lee `scripts/workflow/serie.sh` de la 2.11.0: las sesiones de
  `.evals[]`, las respuestas y las sin juzgar de `.juez`, `.sesiones_sin_medir` y la medida y el total de los umbrales
  `<clase>:<modelo>` y `<clase>:<modelo>:<modo>` cuentan solo lo medido. *Criterio*: test de `make ci` que ejecuta
  `scripts/workflow/serie.sh fila` tal como está en `main` sobre un informe con evals medidas y reutilizadas: sale con 0
  y su fila es la misma que con un informe de solo las medidas.

#### El umbral en tasa y la serie de cierres

- **FR-024**: Un ADR nuevo (`docs/ADR/`, el siguiente número libre) decide cómo cuenta lo reutilizado en el umbral en tasa
  de una clase del juez (ADR 0038) y en la serie de cierres, y fija, por skill con juez, la tasa de referencia con la que
  se comprueba su criterio. Mantiene el criterio del ADR 0038: una ejecución sale en rojo sin motivo como mucho 1 de cada
  50 veces, sea cual sea el número de respuestas medidas. No sube ningún umbral (ADR 0038, punto 4). El job aplica su
  regla. Si la decisión pide cambiar `scripts/workflow/`, el ADR lo dice y ese cambio no entra en el hito.
  La decisión es esta: lo reutilizado no cuenta en el umbral `<clase>:<modelo>` ni en el recuento
  `<clase>:<modelo>:<modo>`, que se aplican solo a las respuestas medidas del modelo que decide, con los dos modos
  juntos; el límite de marcadas para *n* respuestas medidas es el *k* más estricto con P(Bin(*n*, *p*ref) > *k*) ≤ 0,02
  (regla del ADR 0038, punto 3), que con todas las respuestas de la skill es el umbral de hoy (3 de 120 en
  `boe-legislacion`, 1 de 60 en `jurisprudencia`); la serie de cierres cuenta solo lo medido (FR-023), así que
  `scripts/workflow/serie.sh` no cambia. La tasa de referencia *p*ref de cada skill se deriva del umbral vigente: es la
  mayor, redondeada hacia abajo a centésimas de punto porcentual, con la que el límite de su ejecución completa —el
  `umbral` de su clase en `evals/<skill>/juez/clases.yaml` por las respuestas que juzga la ejecución completa— cumple
  aún 1/50. El ADR escribe esa derivación y, como su resultado, las cifras de hoy (≈0,85 % en `boe-legislacion` y
  ≈0,36 % en `jurisprudencia`); si una persona estrecha el `umbral` de `clases.yaml` (ADR 0038, punto 4), *p*ref baja
  con él y el límite se estrecha para todo *n*. El ADR deja dicho que la tasa acumulada de la serie de
  `boe-legislacion` (18 de 1.424, el 1,26 %) está por encima, y que subir un umbral pediría otro ADR.
  *Criterio*: test de `make ci` que, para cada skill con juez y cada número de respuestas medidas del modelo que decide
  entre 1 y todas las de la skill en sus dos modos, calcula con la regla del job la probabilidad de rojo sin motivo a la
  tasa de referencia derivada del `clases.yaml` vigente y sale en rojo si alguna pasa de 0,02; y un mutante que baja el
  `umbral` de `boe-legislacion` a 2 de 120 hace salir en rojo el test si el límite del job con 120 respuestas no es 2.

#### La release

- **FR-025**: La release (`release.yml`) comprueba, antes de publicar nada, que el commit etiquetado tiene una ejecución
  del job de evals de la suite completa (FR-018) en la que el trabajo de cada skill con evals en ese commit terminó en
  verde, con su informe, todas las evals de ese commit medidas y ninguna reutilizada. Si no la tiene, falla, no publica
  la release, ni el cask, ni el manifiesto del bucket, ni el catálogo, y su mensaje nombra el commit, lo que falta y cómo
  lanzar la suite completa sobre él.
  *Criterio*: test de `make ci` sobre `release.yml` que sale en rojo si falta la comprobación, si el trabajo que publica no
  depende de ella, o si la comprobación acepta otro commit, una ejecución en rojo, una con alguna eval reutilizada o una
  sin el informe de alguna skill.

#### Lo que no cambia

- **FR-026**: El hito no cambia qué mide ninguna eval ni sus umbrales: ni las evals de `evals/`, ni las clases y sus
  umbrales del juez, ni las repeticiones, ni la regla de 2 de 3, ni los modelos. Una ejecución de la suite completa da el
  mismo informe que antes del hito, con los campos nuevos de FR-022. *Criterio*: test de `make ci`: el informe de la suite
  completa de unos datos de prueba es el de antes del hito más esos campos.

- **FR-027**: El hito no cambia ningún fichero de `scripts/workflow/`, `.specify/` ni `internal/workflow/`; el job sigue
  dando a `scripts/workflow/cierre.sh` lo que lee hoy: un trabajo `evals (<skill>)` por skill con su informe entre sus
  marcas. *Criterio*: el diff de la propuesta del hito respecto de `main` no tiene ninguna ruta de esos directorios.

### Key Entities

- **Conjunto de dependencias de una eval**: las rutas de FR-001; se calcula, no se escribe.
- **Lo común**: las rutas de FR-002, que eligen todas las evals.
- **Ruta sin asignar**: una ruta de las que miden las evals que no está en ningún conjunto ni en lo común (FR-004).
- **Elección**: las evals elegidas por una propuesta, con su motivo (FR-006, FR-008).
- **Huella de una eval**: el valor de FR-009.
- **Medida de `main`**: la ejecución de FR-010.
- **Resultado de una eval**: lo de FR-011, con su huella y su procedencia.
- **Ejecución de la suite completa**: la de FR-018, la única que vale para la release (FR-025).

## Uso, de fuera adentro

Ninguna skill pide ninguna de estas salidas ni las consulta en una pregunta: las consumen personas y el cierre del
workflow, y crecen con el número de evals, no con las preguntas ni con los meses de uso de kitlegal (sus normas y
bloques consultados no entran en ninguna sesión de evals).

- **La elección y su motivo (FR-006, FR-008)**: la calcula el job una vez por ejecución que mide y la lee quien revisa la
  propuesta, en el registro y en el informe. Una línea por eval elegida, de unos 150 bytes, con como mucho las rutas de
  la propuesta que la eligen: hoy, una por cada una de las 45 evals en el peor caso. Con las materias como paquetes —cientos de evals— crece con las
  evals elegidas, no con la historia. Deja de darse en cuanto la propuesta se fusiona: sobre `main`, ninguna eval está
  elegida.
- **Medida o reutilizada, la huella y la procedencia (FR-022)**: las lee quien revisa el informe de cada skill, una vez
  por ejecución, y el cierre las copia a `gates/evals/<skill>.json`. Una entrada por eval —huella de 64 caracteres,
  ejecución, trabajo, commit de 40, propuesta y fecha—, unos 250 bytes: hoy 23, 10 y 12 entradas por skill, unos 11 KB
  entre las tres; con cientos de evals, decenas de KB por ejecución. La marca «reutilizada» de una eval deja de darse en
  la ejecución en que cambia su huella, en que la ejecución en que se abrieron sus sesiones cumple 90 días —la vida de
  su registro en la plataforma—, aunque una medida posterior la haya reutilizado, o en que su medida de `main` sale de
  las 10 más recientes (FR-012), y entonces se mide.
- **La búsqueda de medidas de `main` (FR-010, FR-012)**: la hace el job una vez por ejecución con reutilización que deja
  alguna eval sin elegir, solo para esas evals; recorre las medidas de `main` legibles de la más reciente hacia atrás,
  como mucho 10, y para cuando cada una tiene su resultado: lee una vez el informe de cada skill de cada medida que
  recorre. Con los informes de una medida de hoy (922, 349 y 534 KiB en `specs/025-h26-1-las-evals-sirven/gates/evals/`,
  unos 1,8 MiB), el peor caso son 10 medidas × 3 informes, unos 18 MiB, y no crece con lo acumulado en la ventana de 90
  días de la plataforma. Lo normal es una sola medida: la más reciente da resultado de cada eval (FR-011), con las
  huellas de `main` en su cabeza. Al llegar a la cota, lo que queda sin resultado se mide.
- **El fallo de la release (FR-025)**: lo lee quien empuja la etiqueta, una vez por etiqueta; un mensaje de unas líneas.
  Deja de darse cuando el commit tiene su ejecución de la suite completa en verde.
- **La serie de cierres (FR-023)**: el cierre escribe una fila por skill con juez y medición, como hoy; la fila solo
  cuenta lo medido, y su tamaño no cambia.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make ci` en verde con todos los tests de este hito. Lo hace cumplir el propio `make ci`, que el workflow
  mide antes de la revisión y en el cierre.
- **SC-002**: Las 45 evals de hoy, y cualquiera que se añada, tienen un conjunto de dependencias calculable y no vacío
  (FR-001, FR-005): el test de `make ci` sale en rojo con 1 eval sin él.
- **SC-003**: Los mutantes de la elección dan exactamente: la referencia de `legal-core`, sus 12 evals y 0 de las demás;
  el cuerpo de `skills/boe-legislacion/SKILL.md`, sus 23 más las 12 de `legal-core` (y las 10 de `jurisprudencia`);
  `internal/cli`, las 45; una descripción, las 45; una ruta sin asignar, las 45 (FR-002 a FR-004). Cada mutante es un
  test de `make ci` que sale en rojo si la elección difiere en una sola eval.
- **SC-004**: El golden de la matriz (FR-019) coincide byte a byte; `make ci` sale en rojo con cualquier diferencia.
- **SC-005**: El test de `release.yml` (FR-025) sale en rojo con cada uno de sus mutantes: sin la comprobación, sin la
  dependencia del trabajo que publica, aceptando otro commit, una ejecución en rojo, una eval reutilizada o un informe
  que falta.
- **SC-006**: Una propuesta que solo toca el cuerpo de `skills/jurisprudencia/SKILL.md` abre 0 sesiones de
  `boe-legislacion` y 0 de `legal-core`, con respuestas grabadas de la plataforma que traen una medida de `main`
  reutilizable para cada eval de `boe-legislacion` y de `legal-core` (si esa medida falta, FR-014 las mide): test de
  `make ci` sobre la decisión del job, que sale en rojo si planifica una sola.
- **SC-007**: Con la etiqueta `evals` sobre una propuesta que no toca nada de lo que las evals miden, y medidas de `main`
  recuperables con la huella de cada eval, el job abre 0 sesiones y las 45 evals salen reutilizadas con su huella: test de
  `make ci` con respuestas grabadas de la plataforma, que sale en rojo si planifica una sola sesión.
- **SC-008**: La huella cumple FR-009 para cada clase de ruta: el test de `make ci` sale en rojo si una ruta que elige una
  eval no cambia su huella, o si una que no la elige la cambia.
- **SC-009**: `scripts/workflow/serie.sh fila`, sin cambios, sale con 0 sobre un informe con reutilizadas y su fila es la
  de solo las medidas (FR-023); el test de `make ci` sale en rojo si no.
- **SC-010**: Con la regla del ADR (FR-024), la probabilidad de rojo sin motivo es ≤ 0,02 para cada número de respuestas
  medidas de 1 a todas, en cada skill con juez, a la tasa derivada del `clases.yaml` vigente; el test de `make ci` sale
  en rojo con un solo número por encima, o si estrechar el `umbral` de `clases.yaml` no estrecha el límite del job.
- **SC-011**: Como mucho 11 sesiones a la vez entre las tres skills, y 5, 4 y 2 por skill (FR-020); el test de la
  definición del job en `make ci` sale en rojo si la definición admite más.
- **SC-012**: El informe de la suite completa es el de antes del hito más los campos de FR-022 (FR-026): el test de
  `make ci` sale en rojo con cualquier otra diferencia.
- **SC-013**: El ADR de FR-024 está en `docs/ADR/` con su decisión sobre el umbral en tasa y la serie, y con la
  derivación de la tasa de referencia desde el `umbral` de `clases.yaml` y las cifras que da hoy para cada skill con
  juez; el test de SC-010 usa la tasa así derivada.
- **SC-014**: La medición de cierre del hito corre con el job nuevo: cada skill deja su trabajo `evals (<skill>)` con su
  informe, que dice de cada eval si se midió o se reutilizó y por qué, y el cierre escribe la fila de la serie de cada skill
  con juez; si la fila no casa con los umbrales de su informe, el cierre se detiene (comprobación de `cierre.sh`, que ya
  existe).
- **SC-015**: El diff del hito no toca `scripts/workflow/`, `.specify/` ni `internal/workflow/` (FR-027).
- **SC-016**: `CHANGELOG.md` (*Unreleased*) y la documentación del job (su contrato y `docs/WORKFLOW.md` si nombra los
  disparadores) dicen qué mide cada disparador y que la release exige la suite completa.

### Criterios del hito, tal como los escribe

**Controles** (SC-002 a SC-005): «un test en `make ci` que exige un conjunto de dependencias calculable para cada eval;
mutantes: tocar la referencia de una skill elige sus evals y no las de otra, tocar el cuerpo de
`skills/boe-legislacion/SKILL.md` elige también las de `legal-core`, tocar `internal/cli` las elige todas, tocar una
descripción las elige todas, y una ruta sin asignar las elige todas; golden de la matriz; un test de `release.yml` que
sale en rojo si la release no exige la suite completa del commit etiquetado.»

**Aceptación** (SC-006 y SC-007): «una propuesta que solo toca el cuerpo de `skills/jurisprudencia/SKILL.md` no abre
ninguna sesión de `boe-legislacion` ni de `legal-core`; y la etiqueta `evals` sobre una propuesta que no toca nada de lo
que las evals miden no abre ninguna sesión: reutiliza los resultados de `main` con su huella.»

**Alcance** (FR-020, FR-026): «no cambia qué mide ninguna eval ni sus umbrales, ni abre más sesiones a la vez contra la
suscripción.»

## Assumptions

- Las propuestas se fusionan con squash (ROADMAP §6): la cabeza de una propuesta fusionada no es un commit de `main`, y la
  huella, que depende del contenido y no del commit, es lo que une una medida de esa cabeza con una propuesta posterior.
- Hasta que se fusione el hito, ninguna medida de `main` tiene huella: la medición de cierre del propio hito lo mide todo
  (toca el arnés y el job, que son lo común), y la reutilización solo se puede ver en la plataforma desde la primera
  propuesta posterior. Por eso SC-006 y SC-007 se comprueban en `make ci` con respuestas grabadas de la plataforma.
- La plataforma da al job la lista de ficheros de una propuesta y los informes de ejecuciones anteriores mientras su
  registro no caduca (90 días), con el permiso de lectura de Actions que ya tiene el trabajo `tanda`.
- La etiqueta `evals-completa` la crea en el repositorio una persona; el job reacciona a su nombre.
- Hoy ninguna eval ni ninguna referencia es de una materia: cada eval depende de todas las referencias de sus skills.
  Por eso en FR-001(b) «la referencia de su materia» son hoy todas las `references/` de las skills de la eval; H30 no
  define ninguna marca de materia —ni en las evals, ni en las referencias, ni en `data/`— ni deja un punto de extensión:
  cómo se sabe la materia de una eval y de una referencia lo fija el molde de vertical de H32, y con él se ajustará el
  cálculo. Supuesto conservador: el ejemplo del objetivo del hito (tocar la referencia fiscal no mide las evals de la
  civil) no se puede comprobar hasta que haya dos paquetes de materia sobre la misma skill; los controles y la
  aceptación del hito, que son por skill, sí se cumplen.
- `jurisprudencia` nombra `boe-legislacion` en su `SKILL.md` («se consulta con `boe-legislacion`»): por FR-001(b), lo que
  elige las evals de `boe-legislacion` por su skill elige también las de `jurisprudencia`. Ninguna skill nombra
  `jurisprudencia`.
- Las skills que una skill nombra se siguen también a través de otra (FR-001(b)); hoy ninguna cadena pasa de un salto, y
  la regla no cambia ninguna elección de hoy.

## Fuera de alcance

- Cambiar `scripts/workflow/` —`serie.sh`, `cierre.sh`, `informe.sh`— o el workflow `hito`: si el ADR de FR-024 lo pide,
  va en su propia propuesta de cambio.
- Reutilizar resultados de ejecuciones que no son medidas de `main` (FR-010): de la propia propuesta, de otra sin
  fusionar o del despacho manual sobre `main`.
- Ampliar las rutas que las evals miden a `tools/` (sus `go.mod` no entran en el binario) u otras que hoy no lanzan el
  job, salvo las grabaciones, los guiones del arnés, `go.mod` y `go.sum` (FR-004).
- Que un cambio en la descripción de una herramienta MCP de un applet que no está en la tabla de comandos de las skills de
  una eval elija esa eval: el hito asigna los paquetes de los applets de esa tabla.
- Cambiar las evals, los umbrales, las clases del juez, las repeticiones, la regla de 2 de 3, los modelos, las versiones
  de Claude Code o la concurrencia de cada skill.
- Abrir sesiones de una skill en más de un trabajo a la vez, o un trabajo por eval.
- Crear la etiqueta `evals-completa` en GitHub, etiquetar una versión o publicar una release: son acciones de una persona.
- Medir la suite completa en `main` de forma periódica o automática.
- Definir una marca de materia en las evals, las referencias o `data/`, o dejar un punto de extensión para ella: llega
  con el molde de vertical de H32 (aclaración de la sesión 2026-10-22).
- Reutilizar la ejecución de un commit de la propuesta fusionada que no sea el que era su cabeza al fusionarla, aunque la
  huella coincida (FR-010).
- Contar lo reutilizado en el umbral en tasa de una clase del juez o en la serie de cierres, y subir ningún umbral
  (FR-024).
