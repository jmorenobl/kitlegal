# Feature Specification: H41.2 · Por qué la eval 04 no activa `normativa` en modo herramienta en el job

**Feature Branch**: `036-h41-2-por-qu-e`

**Created**: 2026-11-11

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H41.2-por-que-la-eval-04-no-activa-en-modo-herramienta.md`
(entrada de la persona, ADR 0018).

## Contexto

El job de evals ha visto cuatro sesiones de la eval 04 de `normativa` (`04-lgt-prescripcion`) que, en **modo
herramienta** (el agente llama a `<applet>_<verbo>` del servidor `kitlegal mcp serve`), piden `boe_articulo` sin ningún
`tool_use` de la herramienta `Skill`, y por eso cuentan en `sin_activar`, cuyo umbral es cero (ADR 0038 y 0045). Los
sondeos de la persona no han visto ninguna. Antes de tocar el umbral hay que saber si la caída es de la skill, del
cliente o del entorno del job; el hito elige medirlo, que es la serie que la D27 de H32 y la D21 de H41 dejaron a una
persona.

Evidencia de partida, tal como la dan el hito y la D21 de `specs/034-h41-las-skills-del/research.md` (no medida en esta
sesión):

- las cuatro caídas son de la eval 04, en modo herramienta, con `claude-sonnet-5-5` y la misma forma: la sesión `-02`
  del cierre de `7f8ef89` y la `-01` del de `75bf303` en H32, una de la suite de la release v0.9.0 (`docs/USO.md`,
  2026-11-02) y la `-01` del primer cierre de H41 (ejecución 42288846648). En los cierres de H39 y H40, 0 de 60. La
  serie no tiene denominador cerrado;
- en los sondeos de la persona, todas las sesiones de la eval 04 en modo herramienta activaron la skill antes de la
  primera herramienta: 70 con Claude Code 2.1.317 (`evidencias/sondeos-20261104/`, `evidencias/sondeos-20261106/`) y
  las de `evidencias/sondeos-20261109/` con 2.1.319;
- el job y los sondeos difieren al menos en la versión de Claude Code (el job fija `VERSION_DE_CLAUDE_CODE: 2.1.284` en
  `.github/workflows/evals.yml`) y en las skills instaladas (el job instala las cuatro del producto con `make install`;
  el arnés del laboratorio, solo la variante de la sesión);
- el informe del job no dice qué skill activó cada sesión, así que no se sabe si la caída activó otra skill
  (`consulta-juridica`) o ninguna.

Comprobado en esta sesión, en el repositorio: el informe del job lee de cada sesión los nombres de sus `tool_use` de
`Skill` en el orden del transcript (`internal/evals/sesion.go`, `SkillsActivadas`), pero solo publica si la sesión
activó la skill evaluada (`activada` en `informe.json` y la columna «Activada» de `informe.md`); el arnés del
laboratorio (`make evals-laboratorio`, siete argumentos) abre en cada variante y modo `SESIONES/4` sesiones de cada una
de las cuatro evals de la familia y deja en cada sesión solo la skill de su variante
(`specs/034-h41-las-skills-del/contracts/arnes.md` §3); y la tabla de una familia del producto da una fila por variante,
modo, modelo y versión de Claude Code (`specs/034-h41-las-skills-del/contracts/tabla.md` §3).

Para leer la tabla del sondeo (calculado en esta sesión con `awk` y la fórmula de Wilson, z = 1,96): 0 de 40 da
0,0–8,8 %, y 0 de 160, 0,0–2,3 %; contra una fila con 0 de 40, otra fila de 40 no toca su intervalo desde 8 de 40
(10,5–34,8 %; 7 de 40 da 8,7–32,0 %).

## Clarifications

### Session 2026-11-11

- Q: ¿Cómo se consigue que el sondeo de H41.2 tenga 40 sesiones de la eval 04 de normativa por fila, si hoy el arnés abre SESIONES/4 sesiones de cada una de las cuatro evals de la familia, no deja elegir la eval y la tabla agrega todas las evals de una fila (el desglose por eval solo está en el nombre de cada sesión del detalle)? (FR-013) → A: A. `make evals-laboratorio` gana un argumento para pedir una sola eval del plan de la familia; con él, las SESIONES sesiones de cada par variante × modo son de esa eval (SESIONES entero > 0, sin exigir múltiplo de 4); sin él, el arnés abre las mismas sesiones que hoy, byte a byte; un valor que no es eval del plan de la familia, o el argumento con variantes de H39, es un error de argumentos de FR-011; la tabla no cambia (auto: criterio a; fuente: hito H41.2, «Entrega» y «Aceptación» (160 sesiones); FR-013 y FR-030; arnés de hoy en `internal/evals/laboratorio_sondeo.go` e `internal/evals/laboratorio_familias.go`; constitución, «Criterio de decisión autónoma», 1 y 2)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - El informe del job dice qué skills activó cada sesión (Priority: P1)

Quien lee el informe del job (`informe.json` e `informe.md`) —la persona antes de fusionar, también cuando el cierre
sale en rojo— ve, por cada sesión, qué skills activó, en qué orden y si fue antes o después de la primera herramienta de kitlegal.
Una sesión que cuenta en `sin_activar` dice si activó otra skill o ninguna. Nada de lo que el informe ya cuenta cambia.

**Why this priority**: es lo que el informe del job no deja saber hoy de la caída, y lo que la persona necesita para
leer la próxima.

**Independent Test**: en `make ci`, sin modelo, el informe de unas sesiones grabadas da por sesión la lista esperada
de skills activadas y las mismas medidas, umbrales y veredicto que hoy.

**Acceptance Scenarios**:

1. **Dado** una sesión grabada de una eval de `normativa` que activa `normativa` en un turno anterior al de la primera
   herramienta de kitlegal, **Cuando** el job escribe su informe, **Entonces** la sesión lleva en `informe.json` y en
   su sección de `informe.md` una sola activación, `normativa`, antes de la primera herramienta (FR-001, FR-002).
2. **Dado** una sesión grabada que llama a `boe_articulo` sin ningún `tool_use` de `Skill`, **Cuando** el job escribe
   su informe, **Entonces** su lista de activaciones está vacía, su sección de `informe.md` dice que no activó ninguna
   skill, y cuenta en `sin_activar` como hoy (FR-001 a FR-004).
3. **Dado** una sesión grabada que activa `consulta-juridica` antes de la primera herramienta y ninguna otra skill,
   **Cuando** el job escribe su informe, **Entonces** su lista es `consulta-juridica`, antes, y la sesión cuenta en
   `sin_activar` como hoy (FR-001, FR-003, FR-004).
4. **Dado** una sesión grabada que activa `consulta-juridica` antes de la primera herramienta y `normativa` después,
   **Cuando** el job escribe su informe, **Entonces** su lista es `consulta-juridica` (antes) y `normativa` (después),
   en ese orden, y la sesión no cuenta en `sin_activar`, como hoy (FR-001, FR-004).
5. **Dado** el juego de sesiones grabadas de FR-006, **Cuando** el job escribe su informe, **Entonces** `sin_activar`,
   cada otro umbral, cada tasa y el veredicto salen igual que con el informe de hoy (FR-004, SC-002).

---

### User Story 2 - El arnés instala, si se le pide, las skills del producto como el job (Priority: P1)

La persona lanza `make evals-laboratorio` con una opción nueva y cada sesión tiene instaladas todas las skills de
`skills/` como las deja `make install`, con la variante en el lugar de la skill de su familia. Sin la opción, el arnés
instala lo que instala hoy.

**Why this priority**: es una de las dos diferencias entre el job y los sondeos; sin ella el sondeo no puede separarla
de la versión de Claude Code.

**Independent Test**: en `make ci`, con un sustituto de `claude` sin modelo ni red, se comprueban las skills que deja
el arnés en cada sesión, con la opción y sin ella.

**Acceptance Scenarios**:

1. **Dado** una petición de `normativa-producto` con la opción, **Cuando** el arnés prepara cada sesión, **Entonces**
   la sesión tiene instaladas `consulta-juridica`, `jurisprudencia`, `normativa` y `plazos-y-territorio`, cada una
   idéntica byte a byte a la de `skills/`, y ninguna más ni repetida (FR-010).
2. **Dado** una petición de una variante versionada de una familia del producto (por ejemplo `normativa-control`) con
   la opción, **Cuando** el arnés prepara cada sesión, **Entonces** la skill de la familia es la variante, byte a byte,
   y las otras tres son las de `skills/`, byte a byte, sin ninguna repetida (FR-010).
3. **Dado** cualquier petición válida sin la opción, **Cuando** el arnés prepara cada sesión, **Entonces** la sesión
   tiene exactamente lo que instala hoy, byte a byte: solo la skill de su variante (FR-010, SC-003).
4. **Dado** una petición con la opción y variantes del laboratorio de H39, o con un valor de la opción que no vale,
   **Cuando** la persona lanza el arnés, **Entonces** sale con un código distinto de 0, dice qué no vale con el nombre
   del argumento de `make` y no abre ninguna sesión, no construye nada y no crea `DIRECTORIO` (FR-011, SC-004).

---

### User Story 3 - La tabla dice qué skills tenía instaladas cada fila y qué skill activó (Priority: P1)

Con las transcripciones de un sondeo de una familia del producto, `make laboratorio-tabla` da por fila las skills que
tenían instaladas sus sesiones, las sesiones que activan la skill de la familia antes de la primera herramienta, las
que activan otra skill y las que no activan ninguna, cada cifra con su intervalo de Wilson al 95 %, y el detalle de
cada sesión con sus activaciones en orden.

**Why this priority**: es la medida que lee el ADR.

**Independent Test**: en `make ci`, sin red ni modelo, el guion da las cifras exactas de unas transcripciones fijas, y
la tabla de los sondeos fijos de H40 y H41 sale igual byte a byte.

**Acceptance Scenarios**:

1. **Dado** un sondeo fijo de `normativa-producto` en modo herramienta, con la misma versión de Claude Code, una vez
   con solo `normativa` instalada y otra con las cuatro skills del producto, **Cuando** el guion lo cuenta,
   **Entonces** da dos filas, una por instalación, cada una con sus skills instaladas (FR-020).
2. **Dado** sesiones fijas que activan `normativa` antes de la primera herramienta, después, `consulta-juridica` sola,
   las dos y ninguna, **Cuando** el guion las cuenta, **Entonces** cada recuento, cada intervalo y cada lista del
   detalle es el esperado: 0 de 40 → 0,0–8,8 %; 1 de 40 → 0,4–12,9 %; 40 de 40 → 91,2–100,0 % (FR-020, FR-024,
   SC-005).
3. **Dado** los sondeos fijos de las suites congeladas de H40 y H41, que no dejan constancia de las skills instaladas,
   **Cuando** el guion los cuenta, **Entonces** la tabla y el detalle salen igual que hoy, byte a byte (FR-022, SC-006).

---

### User Story 4 - La persona mide y lleva la tabla a un ADR (Priority: P2)

Con la propuesta de cambio abierta, la persona lanza el sondeo de H41.2 —`normativa-producto`, eval 04, modo
herramienta, `claude-sonnet-5-5`, cuatro filas de 40 sesiones—, deja su tabla en `evidencias/`, añade la tabla y su
lectura a la propuesta de cambio y lleva la tabla a un ADR que dice qué se hace con la caída.

**Why this priority**: es la aceptación del hito, pero la hace una persona fuera del run (ADR 0032): el run entrega lo
que la hace posible, y la fusión la deciden los controles.

**Independent Test**: el informe final del run dice que la propuesta queda pendiente del sondeo y de su lectura; antes
de fusionar, `evidencias/` tiene la tabla y la propuesta su lectura.

**Acceptance Scenarios**:

1. **Dado** el arnés y el guion en la rama, **Cuando** la persona lanza el sondeo con las cuatro filas de FR-030,
   **Entonces** la tabla, con el modelo, la versión de Claude Code y las skills instaladas de cada fila, queda en
   `evidencias/` con, por fila, las sesiones que activan `normativa` antes de la primera herramienta, las que activan
   otra skill y las que no activan ninguna, con sus intervalos de Wilson al 95 % (FR-030).
2. **Dado** la tabla, **Cuando** una fila reproduce la caída y otra no, con intervalos que no se tocan, **Entonces** el
   ADR dice qué diferencia la causa y qué hito o propuesta la arregla; **Cuando** ninguna la reproduce, el ADR dice la
   tasa que queda acotada con las 160 sesiones y si la serie se sigue llevando a una persona o se decide sobre el
   umbral (FR-031).

---

### Edge Cases

- **Sesión que activa la skill de su eval solo después de la primera herramienta**: no cuenta en `sin_activar`, como
  hoy (`sin_activar` no mira el turno); su lista lo dice, «después» (FR-001, FR-004).
- **Skill activada varias veces en una sesión**: cada `tool_use` de `Skill` es una entrada de la lista, en su orden
  (FR-001); en la tabla, la sesión cuenta una vez en cada recuento (FR-020).
- **Sesión sin ninguna herramienta de kitlegal** (una eval sin binario ni servidor, o una sesión que responde sin
  llamar a ninguna): toda activación cuenta como antes de la primera herramienta, como en H40 (FR-001, FR-020).
- **`Skill` en el mismo turno que la primera herramienta**: no es antes, como en H39 (FR-001, FR-020).
- **Sesión ilegible en el job** (un fichero de la sesión que el guion del job no llegó a escribir): su lista no se da,
  como el resto de lo leído de la sesión, y no se toma por «ninguna» (FR-005).
- **Informe del job anterior a este hito**, versionado en el repositorio: se lee como hoy (FR-005).
- **Sesión del sondeo sin terminar**: va en su columna y no entra en `n`, como en H40; fila con `n` = 0: recuentos e
  intervalos vacíos (FR-021).
- **Sondeo anterior a este hito junto a uno de este hito en la misma tabla**: sus filas tienen como skills instaladas
  la de la familia sola, que es lo que instalaba el arnés de H40 y H41 (FR-022).
- **Transcripción, informe o sondeo que el guion no sabe leer** (manipulado a mano, truncado por algo que no es el
  job ni el arnés): defecto `inesperado`, código 1 (ADR 0023), sin más casos.

## Requirements *(mandatory)*

### Functional Requirements

**El informe del job**

- **FR-001**: Por cada sesión leída, `informe.json` da la lista de sus activaciones de skills: una entrada por cada
  bloque `tool_use` de la herramienta `Skill` del transcript, en su orden, con el nombre de su `input.skill` tal cual y
  si va antes de la primera herramienta de kitlegal o no. «Turno», «activar una skill», «primera herramienta de
  kitlegal» en cada modo y «antes» son los de la tabla de H39 (`specs/031-h39-c-omo-se/spec.md`, el requisito de lo que cuenta su tabla): antes es en
  un turno anterior al de la primera llamada a una herramienta del registro en modo herramienta, o al de la primera
  orden `kitlegal` en modo orden; en el mismo turno no es antes; en una sesión sin ninguna, toda activación es antes.
  Sin ningún `tool_use` de `Skill`, la lista está vacía.
- **FR-002**: La sección de cada sesión de `informe.md` da la misma lista, en el mismo orden, cada nombre con «antes» o
  «después» de la primera herramienta; con la lista vacía, dice que la sesión no activó ninguna skill.
- **FR-003**: Una sesión que cuenta en `sin_activar` dice con su lista si activó otra skill, y cuál, o ninguna: la
  lista no tiene ninguna entrada con el nombre de la skill evaluada.
- **FR-004**: Cómo se cuenta `sin_activar`, su umbral y todo lo demás que el informe mide, juzga o decide no cambian:
  el campo `activada`, la columna «Activada» de la tabla de sesiones, cada tasa, cada umbral, cada motivo y el
  veredicto salen igual que hoy para las mismas sesiones.
- **FR-005**: La lista sale de lo que el informe ya lee de la sesión: de una sesión ilegible no se da, como lo demás
  leído de ella, y no se toma por vacía. Un informe del job escrito antes de este hito, de los que el repositorio
  versiona y lee, se lee como hoy.
- **FR-006**: `make ci` comprueba el informe con sesiones grabadas de los dos modos que activan la skill de su eval,
  otra skill, las dos y ninguna, antes y después de la primera herramienta, incluida una con `Skill` en el mismo turno
  que la primera herramienta: el test falla si una lista, un orden o un «antes»/«después» de `informe.json` o de
  `informe.md` difiere del esperado, o si `sin_activar` u otra medida de FR-004 difiere de la de hoy en cualquiera de
  ellas.

**El arnés del laboratorio**

- **FR-010**: `make evals-laboratorio` acepta una opción que, en las tres familias del producto (`normativa`,
  `plazos-y-territorio` y `jurisprudencia`), deja en cada sesión instaladas todas las skills de `skills/` como las deja
  `make install` —cada una idéntica byte a byte, fichero a fichero, a la de `skills/` del árbol que lanza el sondeo—,
  con la variante de la sesión en el lugar de la skill de su familia, y ningún nombre repetido. En `<skill>-producto`,
  la variante es la copia de `skills/<skill>`, y la sesión tiene `skills/` tal cual. Sin la opción, el arnés instala
  en cada sesión exactamente lo que instala hoy (`specs/034-h41-las-skills-del/contracts/arnes.md` §3), byte a byte.
  Las sesiones que abre no dependen de la opción: las de hoy sin el argumento de FR-013, y las de FR-013 con él. Cómo
  se llama la opción y qué valores toma lo decide el plan; ausente o vacía es «sin la opción».
- **FR-011**: La opción de FR-010 con variantes del laboratorio de H39, que no son de ninguna familia del producto, o
  con un valor que no vale; el argumento de FR-013 con un valor que no es ninguna eval del plan de la familia pedida
  o con variantes del laboratorio de H39, que no tienen un plan de una familia del producto, es un error de
  argumentos como los de hoy: junto a los demás errores, uno por línea y con el
  nombre del argumento de `make`, sale con un código distinto de 0, no abre ninguna sesión, no construye nada y no crea
  `DIRECTORIO`.
- **FR-012**: Cada sondeo de una familia del producto con el arnés de este hito, con la opción y sin ella, deja constancia de las skills que tenía
  instaladas cada una de sus sesiones (sus nombres), donde la lee el guion de la tabla; ni la copia de las skills ni
  su huella se guardan en `DIRECTORIO`, como hoy.
- **FR-013**: `make evals-laboratorio` acepta un argumento que pide una sola eval del plan de la familia de las
  variantes pedidas: su valor es una de las cuatro evals de ese plan (`specs/034-h41-las-skills-del/data-model.md` §1;
  en `normativa`, `04-lgt-prescripcion.yaml`, `05-trlrhl-impuestos-municipales.yaml`,
  `15-irpf-rendimientos-por-materia.yaml` o `22-lcsp-contrato-menor-en-2019.yaml`). Con él, el plan de cada par
  variante × modo son `SESIONES` sesiones de esa eval, de nombre `<eval sin .yaml>-<nn>`; `SESIONES` es entonces un
  entero mayor que 0, y no se le pide que sea múltiplo de 4. Sin el argumento, o vacío, el arnés abre las mismas
  sesiones que hoy, byte a byte: `SESIONES` múltiplo de 4 y `SESIONES/4` de cada una de las cuatro evals
  (`specs/034-h41-las-skills-del/contracts/arnes.md` §2 y §3). Un valor que no es ninguna eval del plan de la familia
  pedida, o el argumento con variantes del laboratorio de H39, es un error de argumentos de FR-011. Cómo se llama el
  argumento lo decide el plan, como la opción de FR-010; el precedente del proyecto es `EVALS` de
  `make evals-sondeo`. La tabla no cambia por esto: como todas las sesiones de una fila son de la eval 04, sus cifras
  son las de la eval 04, sin filtro.
- **FR-014**: Lo demás del arnés no cambia: comprueba antes de abrir ninguna sesión que la versión de Claude Code
  instalada es la de `CLAUDE_CODE` y que el modelo es un id completo, y los anota; lo lanza una persona con la
  credencial del sondeo (ADR 0032), nunca `make ci`, un gancho, el job de evals ni un paso del workflow. El run no lo
  lanza, no graba nada en `testdata/` y no escribe en `evidencias/`.
- **FR-015**: `make ci` comprueba el arnés sin red ni modelo, con un sustituto de `claude` que anota sus invocaciones:
  con la opción, en una variante versionada y en la variante `-producto` de cada una de las tres familias, las skills
  que deja instaladas en cada sesión —las cuatro del producto, la variante en el lugar de la de su familia, cada una
  byte a byte, ninguna repetida— y la constancia de FR-012; sin la opción, lo que instala hoy, byte a byte, con los
  valores esperados de hoy sin cambiar; con el argumento de FR-013, una variante de cada familia del producto abre solo
  sesiones de la eval pedida, tantas como `SESIONES`, y sin él, los valores esperados de hoy no cambian; y con cada
  petición inválida de FR-011, 0 sesiones abiertas. El test falla con la primera sesión cuyas skills o eval difieren
  de las esperadas o con una petición inválida que abre alguna sesión.

**La tabla del sondeo**

- **FR-020**: En la tabla de una familia del producto, de un sondeo que deja la constancia de FR-012, hay una fila por
  variante, modo, modelo, versión de Claude Code y skills instaladas, y cada fila da, además de lo de hoy:
  - las skills instaladas de sus sesiones, por nombre;
  - las que activan otra skill que la de la familia, en cualquier turno, y los nombres de esas otras skills, cada uno
    con las sesiones que lo activan;
  - las que no activan ninguna skill (ningún `tool_use` de `Skill`).

  «Activan antes» sigue siendo las que activan la skill de la familia antes de la primera herramienta de kitlegal, como
  hoy. Cada recuento va como «k de n» con su porcentaje y su intervalo de Wilson al 95 % (z = 1,96), en porcentaje con
  un decimal, como hoy. Una sesión que activa la skill de la familia y otra cuenta en las dos columnas que le tocan; una
  skill activada varias veces cuenta una vez. El detalle de cada sesión da su lista de activaciones en orden, cada una
  con «antes» o «después» de la primera herramienta, con las definiciones de FR-001.
- **FR-021**: `n` son las sesiones terminadas de la fila, como hoy; con `n` = 0, los recuentos e intervalos nuevos
  salen vacíos como los de hoy.
- **FR-022**: Una tabla de sondeos que no dejan la constancia de FR-012 —todos los anteriores a este hito— sale igual
  que hoy, byte a byte, tabla y detalle. En una tabla que junta sondeos con constancia y sin ella, las filas de los que
  no la tienen dan como skills instaladas la de la familia sola, y sus activaciones salen de sus transcripciones como
  las demás.
- **FR-023**: La tabla de la familia del laboratorio de H39 no cambia.
- **FR-024**: `make ci` comprueba el guion con transcripciones fijas de la familia `normativa` en modo herramienta, con
  dos instalaciones de la misma variante, modelo y versión, y sesiones que activan la skill de la familia antes y
  después de la primera herramienta, solo otra skill, las dos y ninguna, una sesión sin terminar y una fila sin ninguna
  terminada, y con los intervalos de la historia 3: el test falla si una fila, un recuento, un intervalo, un nombre o
  una lista del detalle difiere del esperado.

**La aceptación de la persona y lo que no cambia**

- **FR-030**: Con la propuesta de cambio abierta, la persona lanza el sondeo de H41.2 con el arnés de este hito:
  `normativa-producto`, solo la eval 04 (el argumento de FR-013), solo el modo herramienta, `claude-sonnet-5-5`, en
  cuatro filas —Claude Code 2.1.284 y la versión del día, cada una con solo `normativa` instalada y con todas las del
  producto (la opción de FR-010)— y 40 sesiones por fila: cuatro invocaciones con `VARIANTES=normativa-producto
  MODOS=herramienta SESIONES=40 MODELO=claude-sonnet-5-5`, el argumento de FR-013 con la eval 04 y `CLAUDE_CODE`, cada
  una con la opción de FR-010 o sin ella, una fila de la tabla por invocación: 4 × 40 = 160 sesiones, todas de la
  eval 04. Las versiones las elige con `CLAUDE_CODE`, sin tocar el
  job. Su tabla queda en `evidencias/` con, por fila, las sesiones que activan `normativa` antes de la primera
  herramienta, las que activan otra skill y las que no activan ninguna, y sus intervalos de Wilson al 95 %.
- **FR-031**: La tabla y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la propuesta no se fusiona
  sin ellas; no deciden la fusión, que la deciden los controles. La tabla va a un ADR que dice qué se hace con la
  caída: si una fila la reproduce y otra no, con intervalos que no se tocan, qué diferencia la causa y qué hito o
  propuesta la arregla; si ninguna la reproduce, la tasa que queda acotada con las 160 sesiones, y si la serie se sigue
  llevando a una persona o se decide sobre el umbral. El run no lanza el sondeo ni escribe el ADR, y su informe final
  dice que la propuesta queda pendiente de la tabla y su lectura.
- **FR-032**: Las suites de aceptación congeladas de H39, H40 y H41
  (`specs/031-h39-c-omo-se/aceptacion/`, `specs/032-h40-por-qu-e/aceptacion/` y
  `specs/034-h41-las-skills-del/aceptacion/`) siguen en verde en `make ci` sin cambiar ningún fichero suyo.
- **FR-033**: Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos. Si el
  cierre da en rojo la caída que este hito mide (`sin_activar` en la eval 04 de `normativa` en modo herramienta), la
  reparación no la arregla tocando una skill, una eval, un umbral ni las versiones del job; el informe final da el
  umbral sin cumplir con sus cifras, como hoy, y la lista de activaciones de esas sesiones está en el `informe.json` y
  el `informe.md` del job del cierre (FR-001, FR-002), para que decida una persona.

### Key Entities

- **Activación de una skill**: un `tool_use` de `Skill` de una sesión, con el nombre de la skill y si va antes de la
  primera herramienta de kitlegal.
- **Lista de activaciones de una sesión**: sus activaciones en el orden del transcript; vacía si no activó ninguna.
- **Instalación de una sesión del sondeo**: las skills que tiene instaladas: la de su variante sola (sin la opción) o
  las cuatro del producto con la variante en el lugar de la de su familia (con ella).
- **Fila de la tabla**: variante, modo, modelo, versión de Claude Code e instalación, con sus recuentos e intervalos.
- **Sondeo de H41.2**: las 160 sesiones de FR-030 y su tabla en `evidencias/`.

### Uso, de fuera adentro

Ninguna skill del producto ni el binario que se distribuye consume nada de lo que entrega el hito: lo consumen la
persona que lee el informe del job o la tabla y escribe el ADR. Lo que gana la skill es indirecto: saber de dónde viene la caída de la eval 04 antes de cambiar `normativa`, su
umbral o el job.

- **Lista de activaciones en el informe del job** (FR-001, FR-002). La escribe el job una vez por sesión y la lee la
  persona al leer el informe de una medición, también la del cierre cuando sale en rojo: sobre todo la de las sesiones que cuentan en
  `sin_activar`, una o ninguna por medición en la serie de partida. Tamaño: una entrada por `tool_use` de `Skill`,
  acotada por los turnos de la sesión; lo esperado, sin medir en esta sesión, es una o dos entradas de unos 45 bytes en
  `informe.json` (el nombre más «antes») y una línea en `informe.md`. Con unos centenares de sesiones por medición son
  unas decenas de kilobytes como mucho. No crece con el uso del producto —cientos de normas y miles de bloques
  consultados no la cambian—, solo con las sesiones que planifica el job. No es una señal que persista: cada medición
  escribe su informe, y la de la siguiente cabeza da otra lista.
- **Instalación del arnés y su constancia** (FR-010, FR-012). La lee el agente de cada sesión del sondeo al elegir qué
  skill activar —es justo lo que se mide— y el guion de la tabla una vez por sondeo. Tamaño fijo: las cuatro skills de
  `skills/` o la variante sola, y cuatro nombres como mucho en la constancia. No crece con el uso del producto.
- **Columnas nuevas de la tabla** (FR-020). Las lee la persona una vez, para la lectura y el ADR: cuatro filas en el
  sondeo de H41.2, cada una con su instalación (cuatro nombres como mucho), tres recuentos con su intervalo y los
  nombres de las otras skills activadas (tres como mucho con la opción), más el detalle de 160 sesiones con una o dos
  activaciones cada una. Su tamaño lo fijan las filas y las sesiones pedidas, no el uso del producto. Deja de leerse
  cuando el ADR está escrito; un sondeo nuevo da otra tabla.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las sesiones grabadas de FR-006 da en `informe.json` y en `informe.md` su lista de
  activaciones exacta —nombres, orden y «antes»/«después»—. Control: el test de FR-006 en `make ci` falla con la
  primera que difiere.
- **SC-002**: 0 diferencias en `sin_activar`, en las demás medidas de FR-004 y en el veredicto del informe de las
  sesiones grabadas de FR-006 respecto de hoy. Control: el mismo test falla si una difiere.
- **SC-003**: Sin la opción, 0 bytes de diferencia entre lo que el arnés instala en cada sesión y lo que instala hoy;
  con ella, 4 skills por sesión, 0 repetidas y 0 bytes de diferencia con la variante y con `skills/`. Control: el test
  de FR-015 en `make ci` falla con la primera sesión que difiere.
- **SC-004**: 0 sesiones abiertas con cada petición inválida de FR-011. Control: el test de FR-015 falla si una abre
  alguna.
- **SC-009**: Con el argumento de FR-013, el 100 % de las sesiones que abre el arnés son de la eval pedida y son
  tantas como `SESIONES`; sin él, 0 diferencias con las de hoy. Control: el test de FR-015 falla con la primera
  sesión que difiere.
- **SC-005**: El guion da los recuentos, los extremos de Wilson con un decimal y las listas del detalle exactos de las
  transcripciones fijas de FR-024. Control: su test en `make ci` falla con el primer valor que difiere.
- **SC-006**: 0 bytes de diferencia en la tabla y el detalle de los sondeos fijos de las suites congeladas de H40 y H41,
  y las suites congeladas de los tres hitos de FR-032 en verde. Control: esas suites en `make ci` fallan con la primera diferencia.
- **SC-007**: Las evals de las skills del producto en verde en la medición del cierre, en los dos modos, con
  `sin_activar` ≤ 0 como hoy. Control: el job de evals del cierre sale en rojo si una no pasa su umbral, y el informe
  final da sus cifras (FR-033).
- **SC-008**: Antes de fusionar, la tabla de las cuatro filas de 40 sesiones del sondeo de H41.2 está en `evidencias/`
  y la propuesta de cambio lleva su lectura. Lo comprueba la persona (capa 3): ningún control automático lo mide, y el
  run no lo hace (FR-031).

## Criterios del hito (literales)

**Controles**:

- el informe con las skills activadas, con sesiones grabadas que activan la skill de su eval, otra skill, las dos o
  ninguna, antes y después de la primera herramienta; `sin_activar` sale igual que hoy en todas (FR-006, SC-001,
  SC-002);
- el arnés con la opción nueva, sin red ni modelo: qué skills deja instaladas en cada sesión, con la variante en el
  lugar de la de su familia y ninguna repetida; sin la opción, lo que instala hoy, byte a byte (FR-015, SC-003);
- las suites congeladas de H39, H40 y H41 siguen en verde (FR-032, SC-006).

Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos (FR-033, SC-007).

**Aceptación**: el sondeo de la persona deja su tabla en `evidencias/`, con las sesiones de cada fila que activan
`normativa` antes de la primera herramienta, las que activan otra skill y las que no activan ninguna, y sus intervalos
de Wilson al 95 % (FR-030). No decide la fusión, que valen los controles: la tabla y su lectura se añaden a la propuesta
de cambio antes de fusionarla, y la propuesta no se fusiona sin ellas. La tabla va a un ADR que dice qué se hace con la
caída:

- si una fila la reproduce y otra no, con intervalos que no se tocan, qué diferencia la causa y qué hito o propuesta la
  arregla;
- si ninguna la reproduce, la tasa que queda acotada con las 160 sesiones, y si la serie se sigue llevando a una
  persona o se decide sobre el umbral (FR-031, SC-008).

El run no lanza el sondeo (FR-014, FR-031).

## Assumptions

- Las cifras de partida (las cuatro caídas, los 0 de 60 de H39 y H40, las 70 sesiones con 2.1.317) son las que dan el
  hito y la D21 de H41; esta sesión no las ha medido.
- «Las skills de su `tool_use` de `Skill`» es el `input.skill` de cada bloque, tal cual, como lo lee hoy el informe
  (`SkillsActivadas`); «antes o después de la primera herramienta» se mide con las definiciones de la tabla de H39, para
  que el informe del job y la tabla del sondeo lean la activación igual.
- «Como las deja `make install`» es lo que deja `kitlegal skills install -g --host claude` con el binario del árbol:
  las cuatro skills de `skills/`, cada una como está en el árbol.
- «Sin la opción, el arnés hace lo que hace hoy» se lee con su control en el hito: lo que instala no cambia en un
  byte, y las sesiones que abre no dependen de la opción (las de hoy sin el argumento de FR-013, las de FR-013 con él).
  Que el sondeo deje constancia de su instalación también sin la opción (FR-012) no cambia nada de eso, y es lo que deja a la tabla decir las skills instaladas de las filas sin la opción, que el sondeo de
  H41.2 necesita (dos de sus cuatro filas).
- La opción solo tiene sentido en las familias del producto: la del laboratorio de H39 no es ninguna skill de
  `skills/` cuyo lugar ocupar, y con ella la opción es un error de argumentos (FR-011), la lectura que menos
  comportamiento añade.
- El sondeo de H41.2 son 160 sesiones y no 640 porque el arnés puede pedir una sola eval (FR-013): con el arnés de
  hoy, `SESIONES=160` por fila abriría las cuatro evals y su tabla agregaría las cuatro.
- El modelo se pasa como id completo porque un alias lo resuelve el cliente (ADR 0031).

## Fuera de alcance

- Cambiar ninguna skill, su cuerpo, sus referencias o su descripción, también la de `consulta-juridica`.
- Cambiar el umbral de `sin_activar`, cómo se cuenta o ningún otro umbral, límite ni tope.
- Cambiar `VERSION_DE_CLAUDE_CODE` o `MODELO_DE_EVALS` del job, que van juntos y piden su propia propuesta (ADR 0031).
- Evals nuevas, y cambiar las que ya existen, también la eval 04.
- Arreglar la causa, si sale: la elige el ADR que lea el sondeo.
- Medir con el juez en el sondeo, que cuenta activación, sin modelo.
- Guardar las sesiones del job de evals como artefacto, que sigue en `backlog/medida-de-las-skills.md`.
- Que el informe del job dé otras cosas de la sesión que sus activaciones (qué herramientas carga o llama,
  `ToolSearch`, la lectura del `SKILL.md`), o que la lista de activaciones entre en un umbral, una tasa o el veredicto.
- Columnas nuevas en la tabla de la familia del laboratorio de H39, la constancia de FR-012 en sus sondeos y la opción
  de FR-010 con sus variantes (FR-011).
- Instalaciones distintas de las dos que pide el hito (la variante sola, o todas las del producto con la variante en
  su lugar): otras combinaciones de skills, skills de fuera de `skills/` o más de una variante por sesión.
- Un filtro por eval en el guion de la tabla o columnas por eval: la tabla agrega las sesiones de la fila como hoy
  (FR-013); pedir más de una eval a la vez al arnés; pedir una eval en una familia del laboratorio de H39.
- Que el job de evals mida las variantes o lance el arnés, o que `make ci` abra sesiones con modelo.
- Lanzar el sondeo y escribir el ADR dentro del run, grabar lecturas nuevas o escribir en `evidencias/` desde el run.
- Cambiar el informe final del workflow (`scripts/workflow/informe.sh`) para que dé las activaciones de cada sesión.
