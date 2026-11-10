# Feature Specification: H41.1 · `plazos-y-territorio` manda leer sus referencias sin condición

**Feature Branch**: `035-h41-1-plazos-y-territorio`

**Created**: 2026-11-10

**Status**: Draft

**Input**: el fichero del hito `docs/roadmap/in-progress/H41.1-plazos-y-territorio-manda-leer-sin-condicion.md`, la
única instrucción humana del run (ADR 0018).

## Contexto

H41 midió en el laboratorio tres formas de las dos órdenes de leer de `plazos-y-territorio` y no tocó la skill del
producto, que sigue siendo la de v0.9.0. El ADR 0053 eligió `plazos-y-territorio-sin-condicion` (punto 1) por lo medido en
`evidencias/sondeos-20261109/`: con `claude-sonnet-5-5`, Claude Code 2.1.319, las evals 02, 05, 10 y 11 y 40 sesiones por
variante y modo, abre `municipios.md` en 39 y 40 de 40, frente a 3 y 11 de `plazos-y-territorio-control`, con intervalos
que no se tocan, y no queda por debajo en «Activan antes» ni en `calculo_de_plazos.md`. La lista de control abrió
`municipios.md` en 17 de 40 en modo orden en el sondeo de la #245 y en 3 en el de la #248: una comparación con ella solo
vale dentro de un mismo sondeo (ADR 0053, «Consecuencias»).

Lo que gana la skill: en toda pregunta a la que llega, el agente abre `municipios.md` antes de la primera orden o
llamada —qué hacer si falta el municipio o no sale con 0, qué trae `data`, la plantilla del territorio y su cobertura—, y
en toda pregunta que llega al paso 8 abre `calculo_de_plazos.md` antes de situar el plazo.

Estado de partida, comprobado en esta sesión: `diff -rq skills/plazos-y-territorio
laboratorio/plazos-y-territorio/plazos-y-territorio-control/skills/plazos-y-territorio` no da ninguna diferencia. Los
cuerpos (10 930 caracteres el de v0.9.0, el tope sin paquetes; 10 991 el de `sin-condicion`; 10 882 el de `sin-condicion`
con M1) son los de `specs/034-h41-las-skills-del/contracts/variantes.md` §3, medidos en H41; esta sesión no los ha vuelto
a medir.

## Clarifications

### Session 2026-11-10

- Q: ¿Cómo trata la regla del verbo la viñeta de jerarquia_normativa.md (TestRemisionesDelRepositorio sin excluir plazos-y-territorio; FR-007)? → A: opción C. TestRemisionesDelRepositorio deja de saltarse plazos-y-territorio y le aplica las dos partes: la regla del verbo, de las expresiones y de la referencia sin orden sobre todo el cuerpo, y la lista del plan con sus tres frases enteras y en orden. Lo que da la regla con una skill del producto se compara por igualdad con una tabla, por skill, de los fallos que admite (mismo patrón que fallosDeControl, tabla aparte); una skill que no está en la tabla admite cero. Para plazos-y-territorio la tabla lleva exactamente «plazos-y-territorio: references/jerarquia_normativa.md: un enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)», con un comentario que dice por qué; el test sale en rojo si la regla da otro fallo o deja de dar ese (auto: criterio c; fuente: docs/roadmap/done/H41.1-plazos-y-territorio-manda-leer-sin-condicion.md «Entrega», «Fuera» y «Controles»; FR-002, FR-005 y FR-007; ADR 0053, «Decisión» 1 y «Lo que cuesta»; internal/skills/remisiones_test.go)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La skill manda leer `municipios.md` y `calculo_de_plazos.md` sin condición (Priority: P1)

Una persona pregunta a su agente por el territorio de un municipio o por hasta cuándo puede recurrir una resolución. El
agente activa `plazos-y-territorio` y, antes de su primera orden o llamada, lee `municipios.md`; si la pregunta llega a
un plazo, al final del paso 7 lee `calculo_de_plazos.md`, antes del 8. El resto de la skill —el protocolo, las reglas,
el frontmatter y la descripción— es el de v0.9.0.

**Why this priority**: es el objetivo del hito y lo único que cambia lo que hace el agente con la skill.

**Independent Test**: `make ci` compara `skills/plazos-y-territorio` con `plazos-y-territorio-sin-condicion` con M1
aplicado (FR-005) y sale en verde solo si coinciden.

**Acceptance Scenarios**:

1. **Dado** el `SKILL.md` de `skills/plazos-y-territorio` tras el hito, **Cuando** se busca la primera frase del
   «Protocolo», **Entonces** es «Antes de la primera orden o llamada, lee [municipios.md](references/municipios.md): qué
   hacer si falta el municipio o no sale con 0, qué trae `data`, la plantilla del territorio y qué dice su cobertura.»,
   en un párrafo propio delante de «Territorio: del 1 al 4…», con los saltos de línea de §2.1 del contrato de H41.
2. **Dado** el mismo `SKILL.md`, **Cuando** se lee el paso 7, **Entonces** termina en «Un acto presunto no tiene fecha:
   ningún plazo. Antes del 8, lee [calculo_de_plazos.md](references/calculo_de_plazos.md): los municipios que cuentan,
   el plazo sin calcular y la lectura en la fecha de inicio.».
3. **Dada** la lista de «Referencias» del mismo `SKILL.md`, **Cuando** se leen sus viñetas, **Entonces** no hay ninguna
   con un enlace a `references/municipios.md` ni a `references/calculo_de_plazos.md`, y la de `jerarquia_normativa.md`
   y la región de los paquetes son las de v0.9.0.
4. **Dado** el mismo `SKILL.md`, **Cuando** se lee el paso 1, **Entonces** conserva «Si no lo dice, **pregúntalo** y
   espera».

---

### User Story 2 - M1: la frase del 4 o 5 pasa a `calculo_de_plazos.md` (Priority: P1)

Para que el cuerpo no pase del tope, la frase del paso 10 sobre la lectura que sale con 4 o 5 sale del cuerpo y va al
final de la referencia que ahora se manda leer siempre antes del paso 8 (ADR 0052, punto 4; ADR 0053, punto 1).

**Why this priority**: sin M1, `sin-condicion` mide 10 991 caracteres y pasa de los 10 930 que caben con la línea de un
paquete.

**Independent Test**: `make ci` comprueba que la frase no está en el cuerpo y sí al final de la referencia (FR-003,
FR-005).

**Acceptance Scenarios**:

1. **Dado** el `SKILL.md` de `skills/plazos-y-territorio` tras el hito, **Cuando** se busca «Si sale con 4 o 5,
   **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin afirmar qué dice.», con los blancos juntados,
   **Entonces** no está, y el paso 10 termina en «…y citarlo con `⚠ REDACCIÓN A FECHA:`.».
2. **Dado** `skills/plazos-y-territorio/references/calculo_de_plazos.md` tras el hito, **Cuando** se leen sus últimas
   líneas, **Entonces** el fichero de v0.9.0 sigue entero y, detrás de una línea en blanco, termina con esa frase,
   palabra por palabra.
3. **Dadas** `administrativo.md`, `jerarquia_normativa.md` y `municipios.md` de la skill, **Cuando** se comparan con las
   de `plazos-y-territorio-control`, **Entonces** son byte a byte iguales.

---

### User Story 3 - `make ci` hace cumplir la forma, el laboratorio, las remisiones y el tope (Priority: P1)

Quien cambie después la skill o el laboratorio ve en rojo cualquier texto que se aparte de lo aplicado, cualquier byte
cambiado de una variante del laboratorio, una remisión que no sea la del plan o un cuerpo por encima del tope.

**Why this priority**: son los «Controles» del hito; sin ellos lo aplicado no queda protegido.

**Independent Test**: cada test de FR-005 a FR-008 sale en verde con la skill del hito y en rojo con una mutación de un
solo cambio (un byte del cuerpo, de una referencia o de una variante del laboratorio).

**Acceptance Scenarios**:

1. **Dado** el repositorio tras el hito, **Cuando** se ejecuta `make ci`, **Entonces** sale con 0, con
   `TestSkillsDelProductoDesdeControl`, `TestVariantesDeControl`, `TestVariantesDePlazosYTerritorio`,
   `TestRemisionesDelRepositorio`, `TestVerticalDeEnsayo` y el guion `h41-skills-instaladas.txtar` de FR-013 en verde.
2. **Dado** `skills/plazos-y-territorio` con un carácter del cuerpo cambiado fuera de lo que genera `make skills-sync`,
   **Cuando** se ejecuta `make ci`, **Entonces** `TestSkillsDelProductoDesdeControl` sale en rojo con la ruta del
   fichero que difiere.
3. **Dada** una variante de `laboratorio/plazos-y-territorio/` con un byte cambiado, **Cuando** se ejecuta `make ci`,
   **Entonces** `TestVariantesDeControl` (si es `plazos-y-territorio-control`) o `TestVariantesDePlazosYTerritorio` (si
   es una de las tres formas) sale en rojo con su ruta.
4. **Dada** una remisión de `plazos-y-territorio` cuya frase no es la de la lista del plan del hito, **Cuando** se
   ejecuta `make ci`, **Entonces** `TestRemisionesDelRepositorio` sale en rojo con «plazos-y-territorio: la remisión
   <n> no es la del plan».
5. **Dado** el cuerpo de `plazos-y-territorio` con la línea de la vertical de ensayo de H32, **Cuando** se ejecuta
   `make ci`, **Entonces** mide 10 952 caracteres como mucho, y si pasara de 11 000, `TestVerticalDeEnsayo` saldría en
   rojo.

---

### User Story 4 - La persona mide la forma aplicada con el sondeo de H41.1 (Priority: P2)

La persona lanza, fuera del run y con su credencial, el sondeo de H41.1 en el arnés de H41 sin cambiarlo, deja su tabla
en `evidencias/` y la añade, con su lectura, a la propuesta de cambio. Con ella decide si se fusiona.

**Why this priority**: es la aceptación del hito, pero la mide una persona; el run solo deja la skill lista para medirse.

**Independent Test**: con la tabla de la persona, cada condición de SC-003 se lee celda a celda.

**Acceptance Scenarios**:

1. **Dado** el repositorio tras el hito, **Cuando** la persona ejecuta `make evals-laboratorio
   VARIANTES=plazos-y-territorio-control,plazos-y-territorio-producto MODOS=orden,herramienta SESIONES=40
   MODELO=claude-sonnet-5-5 CLAUDE_CODE=<versión> CONCURRENCIA=<n> DIRECTORIO=<dir>` y después `make laboratorio-tabla
   DIRECTORIO=<dir>`, **Entonces** la tabla tiene cuatro filas —dos variantes por dos modos— y 160 sesiones, diez por
   eval (02, 05, 10 y 11) en cada fila, y `plazos-y-territorio-producto` es `skills/plazos-y-territorio` con la forma
   aplicada.
2. **Dada** esa tabla, **Cuando** la persona la lee con intervalos de Wilson al 95 %, **Entonces** el hito vale si, en
   cada modo, `producto` no queda por debajo de `control` en «Activan antes» ni en `administrativo.md`,
   `calculo_de_plazos.md` y `municipios.md`, y en `municipios.md` queda por encima en los dos modos, con intervalos que no
   se tocan (SC-003).
3. **Dada** una tabla que no cumple SC-003, **Cuando** la persona la lee, **Entonces** la propuesta no se fusiona y lo
   medido va a un ADR que dice qué medir después.

---

### User Story 5 - La guía y el registro de cambios dicen lo medido (Priority: P3)

Quien escriba o revise una skill lee en `docs/SKILLS.md` §1 («Cómo se remite») la regla que dejó lo medido en
`normativa` y `plazos-y-territorio`, y quien instale una versión lee en `CHANGELOG.md` que `plazos-y-territorio` cambió.

**Why this priority**: documenta; no cambia lo que hace el agente.

**Independent Test**: lectura de los dos ficheros contra FR-009 y FR-010.

**Acceptance Scenarios**:

1. **Dado** `docs/SKILLS.md` tras el hito, **Cuando** se lee el apartado «Cómo se remite» de §1, **Entonces** tiene una
   línea que dice que lo que la skill necesita en las preguntas a las que llega se manda leer sin condición, antes de la
   primera orden o llamada o del paso que lo usa, con lo medido en `normativa` y `plazos-y-territorio`, y no dice que el
   punto 2 del ADR 0052 sea una regla.
2. **Dado** `CHANGELOG.md` tras el hito, **Cuando** se lee *Unreleased*, **Entonces** tiene una entrada de H41.1 por el
   cambio de `plazos-y-territorio`.

### Edge Cases

- **La región de los paquetes.** La línea del paquete la escribe `make skills-sync` desde `data/`; el hito no la toca, y
  con la vertical de ensayo de H32 la región lleva una línea más (69 caracteres y su salto). Con ella el cuerpo mide
  10 882 + 70 = 10 952 caracteres, por debajo de 11 000 (FR-004).
- **Una pregunta solo de territorio** (evals 02): llega a la orden de `municipios.md` y no a la de
  `calculo_de_plazos.md`, que está al final del paso 7.
- **Una pregunta fuera de lo cubierto o con acto presunto**: el paso 7 la para antes del 8; la orden de
  `calculo_de_plazos.md` va al final del paso 7, de modo que el agente puede leerla también en ese caso, igual que en la
  variante medida (ADR 0053, «Lo que cuesta»).
- **`jerarquia_normativa.md`**: ninguna de las cuatro evals llega a su condición (research D15 de H41); su celda sale en
  la tabla y no entra en la comparación de SC-003. Su viñeta no lleva «lee»: es el único fallo de la regla que
  `TestRemisionesDelRepositorio` admite para la skill (FR-007).
- **Ficheros de la skill o del laboratorio cambiados a mano fuera del binario**: no son una vía del producto; los
  controles de `make ci` los dan en rojo y el binario los trata con la regla genérica (defecto `inesperado`, código 1,
  ADR 0023).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Los tramos del `SKILL.md` de `skills/plazos-y-territorio` que cambian son, palabra por palabra y con sus
  saltos de línea, los de `specs/034-h41-las-skills-del/contracts/variantes.md` §2.1: salen de «Referencias» las viñetas
  de `municipios.md` y de `calculo_de_plazos.md`; el «Protocolo» empieza con el párrafo «Antes de la primera orden o
  llamada, lee [municipios.md](references/municipios.md): …» seguido de una línea en blanco; y el paso 7 termina con
  «Antes del 8, lee [calculo_de_plazos.md](references/calculo_de_plazos.md): …». Se comprueba con FR-005.
- **FR-002**: La línea de `jerarquia_normativa.md`, la del paquete (región entre `<!-- paquetes: …` y `<!-- fin de los
  paquetes -->`), «Si no lo dice, **pregúntalo** y espera» en el paso 1, el frontmatter —la descripción incluida— y el
  resto del cuerpo quedan como en v0.9.0, salvo el movimiento de FR-003. Se comprueba con FR-005.
- **FR-003**: El movimiento M1, y ningún otro: la frase del paso 10 «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di
  aparte que no has podido leerlo, sin afirmar qué dice.» (104 caracteres; los 105 del ADR 0053 cuentan también el
  espacio que la precede en el paso 10) sale del cuerpo y va entera, palabra por
  palabra, al final de `references/calculo_de_plazos.md`, detrás de una línea en blanco. `administrativo.md`,
  `jerarquia_normativa.md` y `municipios.md` no cambian. Se comprueba con FR-005.
- **FR-004**: El cuerpo del `SKILL.md` de `plazos-y-territorio` mide 10 882 caracteres, y ningún cuerpo de una skill del
  producto pasa de 11 000, tampoco el de `plazos-y-territorio` con la línea de la vertical de ensayo de H32 (10 952).
  Controles: el tope de 11 000 lo hace cumplir `TestSkillsDelRepositorio` (`make ci` en rojo si un cuerpo pasa de
  11 000) y, con la línea de la vertical, `TestVerticalDeEnsayo` (en rojo si la sincronía de la copia da un defecto, y un
  cuerpo por encima del tope lo es); la cifra exacta de 10 882 la hace cumplir FR-005, porque cualquier carácter de más o
  de menos en el cuerpo lo aparta de la variante con M1.
- **FR-005**: `TestSkillsDelProductoDesdeControl` comprueba que `skills/plazos-y-territorio` es
  `plazos-y-territorio-sin-condicion` con M1 y nada más: fuera de lo que genera `make skills-sync`, sus ficheros son, ruta
  a ruta y byte a byte, los de la variante con la frase de FR-003 quitada del cuerpo y añadida al final de
  `references/calculo_de_plazos.md`; la frase movida no está en el cuerpo y está en la referencia; y el frontmatter es el
  de control. El subtest de H41 que comparaba `skills/plazos-y-territorio` con `plazos-y-territorio-control` «hasta
  H41.1» desaparece, sustituido por esta comprobación. Sale en rojo con la ruta del primer fichero que difiere, falta o
  sobra. Que lo generado sea lo que da `make skills-sync` lo sigue comprobando `TestSkillsDelRepositorio`.
- **FR-006**: `plazos-y-territorio-control` y las tres variantes de H41 (`plazos-y-territorio-sin-condicion`,
  `plazos-y-territorio-condicion-por-orden`, `plazos-y-territorio-lista-con-verbo`) no cambian un byte. Controles:
  `TestVariantesDeControl` (huellas SHA-256 de los ficheros de control) y `TestVariantesDePlazosYTerritorio` (cada forma,
  su declaración aplicada sobre control), sin cambiar ningún valor esperado; los dos salen en rojo con la ruta del
  fichero que difiere, falta o sobra.
- **FR-007**: `TestRemisionesDelRepositorio` se aplica a `plazos-y-territorio`, que deja de estar entre las skills que
  declara fuera «hasta H41.1». Sus remisiones del plan son, en orden, las frases del cuerpo con un enlace a una
  referencia fuera de la región de los paquetes: la de `jerarquia_normativa.md` de v0.9.0, la de `municipios.md` de
  FR-001 y la de `calculo_de_plazos.md` de FR-001. El test sale en rojo con «plazos-y-territorio: la remisión <n> no es
  la del plan» si una frase difiere, y con el defecto de la regla si una frase con un enlace a una referencia incumple
  la regla del verbo o de las expresiones. Aplica las dos partes del test: la regla del verbo, de las expresiones y de la
  referencia sin orden sobre todo el cuerpo, región de los paquetes incluida, y la lista del plan. La regla se compara por
  igualdad con una tabla de fallos admitidos por skill del producto (mismo patrón que `fallosDeControl`, en una tabla
  aparte porque esta es del producto y aquella del laboratorio); una skill que no está en la tabla admite cero fallos.
  Para `plazos-y-territorio` la tabla lleva exactamente «plazos-y-territorio: references/jerarquia_normativa.md: un
  enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)», con un comentario que dice por qué: la
  línea queda como en v0.9.0 (FR-002; ADR 0053, punto 1 y «Lo que cuesta») y ninguna eval llega a su condición. El test
  sale en rojo si la regla da cualquier otro fallo y también si deja de dar ese, de modo que el hito que cambie esa
  viñeta tiene que retirar la declaración.
- **FR-008**: El sondeo de H41.1 corre en el arnés de H41 sin cambiarlo: la orden de US4, escenario 1, con
  `plazos-y-territorio-control` y `plazos-y-territorio-producto`, los dos modos, las cuatro evals de H41 (02, 05, 10 y
  11) y diez sesiones por eval, cuatro filas y 160 sesiones, y su tabla con `make laboratorio-tabla`. Ni las recetas, ni
  los argumentos, ni las reglas de lectura del arnés, ni la tabla, ni el generador de las referencias cambian; sus tests
  (`TestRecetasDelSondeo` y los de la tabla y el laboratorio) siguen en verde sin cambiar ningún valor esperado. Las tres
  variantes de H41 no son filas de este sondeo. El run no lo lanza.
- **FR-009**: `docs/SKILLS.md` §1 («Cómo se remite») lleva una línea, sobre lo que dejó la #248 y con lo medido en
  `normativa` y `plazos-y-territorio` (ADR 0053, punto 3): lo que la skill necesita en las preguntas a las que llega se
  manda leer sin condición, antes de la primera orden o llamada o del paso que lo usa. No recoge el punto 2 del ADR 0052,
  que sigue siendo una hipótesis. Las dos frases del mismo apartado que describen a `plazos-y-territorio` como fuera de
  la regla «hasta H41.1» («sigue como estaba hasta H41.1…» y «—salvo `plazos-y-territorio`, que el test declara fuera
  hasta H41.1—») dejan de decirlo, porque describen el control que cambia FR-007: pasan a describir el control tal como queda, con la
  viñeta de `jerarquia_normativa.md` de `plazos-y-territorio` como el fallo que el test admite (FR-007), porque quitarlas
  sin más haría decir a la guía que toda frase con un enlace lleva «lee».
- **FR-010**: `CHANGELOG.md`, en *Unreleased*, lleva una entrada de H41.1 por el cambio de `plazos-y-territorio`: sus dos
  órdenes de leer sin condición y la frase del 4 o 5 en `calculo_de_plazos.md`.
- **FR-011**: Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos. Control: el
  job de evals de la medición del cierre (`medir_cierre`), que sale en rojo si una eval de una skill del producto no pasa
  en un modo; las evals no cambian.
- **FR-012**: Fuera de lo que piden FR-001 a FR-011 y FR-013, nada más cambia: ni `normativa`, `jurisprudencia` o `consulta-juridica`, sus referencias o sus
  descripciones; ni la descripción de `plazos-y-territorio`; ni ningún umbral, límite ni tope; ni el arnés, sus reglas de
  lectura o el generador de las referencias desde `data/`; ni las evals. `TestSkillsDelProductoDesdeControl` sigue
  comprobando `normativa` y `jurisprudencia` contra su declaración de H41 sin cambiar ningún valor esperado.
- **FR-013**: El guion de aceptación de H41 `internal/app/testdata/script/h41-skills-instaladas.txtar`, que `make ci`
  corre sobre lo que instala `kitlegal skills install` desde las skills empotradas, deja de afirmar en su bloque de
  `plazos-y-territorio` (líneas 48-53) el estado de v0.9.0 y pasa a afirmar la forma aplicada: las dos órdenes de
  FR-001, cada una en su sitio; la lista de «Referencias» sin las viñetas de `municipios.md` ni de
  `calculo_de_plazos.md`; «Si no lo dice, **pregúntalo** y espera» en el paso 1; y la frase de FR-003 fuera del cuerpo
  y al final de `references/calculo_de_plazos.md`. Las frases de su cabecera que describen ese bloque (líneas 11-13 y
  17-20: «conserva su lista de control», «la igualdad byte a byte de plazos-y-territorio con su control») pasan a
  describir lo que afirma. Es la misma sustitución que FR-005 hace con el subtest «hasta H41.1»; el resto del guion
  —`normativa`, `jurisprudencia` y que ninguna variante del laboratorio se instala— no cambia. Sale en rojo con la
  línea del guion que no se cumple.

### Uso, de fuera adentro

- **El cuerpo del `SKILL.md`.** Lo lee el agente una vez por pregunta, al activar la skill: 10 882 caracteres, 10 952 con
  la línea de un paquete más, nunca más de 11 000 (FR-004). No crece con el uso: cientos de normas y miles de bloques
  consultados no lo tocan; la región de los paquetes crece con los paquetes declarados en `data/`, no con las consultas.
- **`municipios.md`** (3 944 bytes, escrita a mano, no cambia): la pide la orden de FR-001 una vez por pregunta, en toda
  pregunta a la que llega la skill, antes de la primera orden o llamada; el agente toma de ella qué hacer si falta el
  municipio o no sale con 0, qué trae `data`, la plantilla del territorio y la cobertura.
- **`calculo_de_plazos.md`** (2 760 bytes de v0.9.0 más la frase de M1, unos 110 bytes): la pide la orden de FR-001 una
  vez por pregunta, al final del paso 7, en toda pregunta que llega a él; el agente toma de ella los municipios que
  cuentan, el plazo sin calcular, la lectura en la fecha de inicio y, con M1, lo que dice si la lectura sale con 4 o 5.
  Lo movido se lee solo donde se lee la referencia, en vez de en toda pregunta con el cuerpo.
- **Por pregunta**: una de territorio lee el `SKILL.md` + `municipios.md`, unos 16 KB (16 057 bytes); una de plazos, además
  `calculo_de_plazos.md` y la referencia del paquete, como en v0.9.0. Ninguna referencia crece con las normas ni los
  bloques consultados; solo cambian cuando un hito las cambia.
- **Las invocaciones de `kitlegal` por pregunta** no cambian: el hito no añade ni quita ninguna orden ni llamada, solo
  cambia dónde y sin qué condición se manda leer dos referencias.
- **Señales.** Ninguna en el producto: el contenido no depende de la consulta ni caduca. Los defectos de FR-004 a FR-008
  los lee quien cambia una skill o el laboratorio, una vez por `make ci` o `make skills-check`, una línea por fichero o
  remisión que difiere; no dependen del uso y dejan de darse cuando lo cambiado vuelve a ser lo declarado.
- **La tabla del sondeo** la lee la persona una vez, antes de fusionar: cuatro filas, una por variante y modo, y el
  detalle de 160 sesiones. No se repite en el producto ni con el uso.

### Key Entities

- **Skill del producto `plazos-y-territorio`**: `skills/plazos-y-territorio/`, el `SKILL.md` y cuatro referencias; la
  que empotra el binario e instala `kitlegal skills install`, y la que el sondeo mide como `plazos-y-territorio-producto`.
- **Variantes del laboratorio de `plazos-y-territorio`**: `plazos-y-territorio-control` (v0.9.0) y las tres formas de
  H41, en `laboratorio/plazos-y-territorio/`, fuera del producto; no cambian.
- **Movimiento M1**: la frase del paso 10 sobre el 4 o 5, que pasa del cuerpo al final de `calculo_de_plazos.md`.
- **Remisiones del plan**: las frases del cuerpo con un enlace a una referencia, fuera de la región de los paquetes, en
  su orden, que `TestRemisionesDelRepositorio` exige enteras.
- **Tabla del sondeo de H41.1**: la que imprime `make laboratorio-tabla` sobre las 160 sesiones de la persona, en
  `evidencias/`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Con la skill del hito, `make ci` sale con 0, con el guion `h41-skills-instaladas.txtar` de FR-013 en
  verde, y cada uno de `TestSkillsDelProductoDesdeControl`,
  `TestVariantesDeControl`, `TestVariantesDePlazosYTerritorio`, `TestRemisionesDelRepositorio` y
  `TestVerticalDeEnsayo` sale en rojo con una mutación de un solo cambio en lo que protege (un carácter del cuerpo o de
  `calculo_de_plazos.md` de la skill, un byte de una variante del laboratorio, una palabra de una remisión, un cuerpo
  por encima de 11 000 caracteres con la línea de la vertical).
- **SC-002**: El cuerpo de `plazos-y-territorio` mide 10 882 caracteres y, con la línea de la vertical de ensayo,
  10 952; ningún cuerpo de una skill del producto pasa de 11 000. Si uno pasa de 11 000, `make ci` sale en rojo
  (`TestSkillsDelRepositorio`; con la línea de la vertical, `TestVerticalDeEnsayo`); si el de `plazos-y-territorio` no
  mide 10 882, sale en rojo `TestSkillsDelProductoDesdeControl`.
- **SC-003**: Aceptación del hito, medida por la persona y no por el run: el sondeo de la persona, con
  `claude-sonnet-5-5` y 40 sesiones por variante y modo, deja su tabla en `evidencias/`, y el hito vale si cumple todo
  esto (ADR 0053, punto 2), con intervalos de Wilson al 95 %:
  - en cada modo, `plazos-y-territorio-producto` no queda por debajo de `plazos-y-territorio-control` en las sesiones
    que activan la skill antes de la primera herramienta ni en ninguna referencia a cuya condición llegan las evals
    (`administrativo.md`, `calculo_de_plazos.md` y `municipios.md`). «Por debajo» quiere decir con un intervalo que no se
    toca con el de `control`;
  - en `municipios.md`, `producto` queda por encima de `control` en los dos modos, con intervalos que no se tocan.

  El run no lanza el sondeo. La tabla y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la
  propuesta no se fusiona sin ellas. Si no cumple, no se fusiona, y lo medido va a un ADR que dice qué medir después. Lo
  hace cumplir la decisión de fusionar, que es siempre de una persona (ADR 0018 y 0021).
- **SC-004**: Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos; si una no
  pasa, el job de evals del cierre sale en rojo.
- **SC-005**: Cero bytes cambiados en `laboratorio/plazos-y-territorio/`, en las skills `normativa`, `jurisprudencia` y
  `consulta-juridica`, en `evals/` y en el arnés del laboratorio; en `laboratorio/`, lo hacen cumplir
  `TestVariantesDeControl` y `TestVariantesDePlazosYTerritorio`, y en `normativa` y `jurisprudencia`,
  `TestSkillsDelProductoDesdeControl`.

### Criterios del hito, literales

**Controles** (del fichero del hito):

- `make ci` comprueba que `skills/plazos-y-territorio` es `plazos-y-territorio-sin-condicion` con M1 y nada más
  (`TestSkillsDelProductoDesdeControl`): sustituye el subtest de H41 que la compara con `plazos-y-territorio-control`
  «hasta H41.1»;
- que `plazos-y-territorio-control` y las tres variantes de H41 no cambian (`TestVariantesDeControl`,
  `TestVariantesDePlazosYTerritorio`);
- la frase de cada remisión de `plazos-y-territorio` (`TestRemisionesDelRepositorio`), que deja de declararla fuera
  hasta H41.1;
- el cuerpo de cada skill del producto, con 11 000 caracteres como mucho, también el de `plazos-y-territorio` con la
  línea de la vertical de ensayo de H32 (`TestVerticalDeEnsayo`).

Las evals de las skills del producto salen en verde en la medición del cierre, en los dos modos.

**Aceptación** (del fichero del hito): el sondeo de la persona, con `claude-sonnet-5-5` y 40 sesiones por variante y
modo, deja su tabla en `evidencias/`, y el hito vale si cumple todo esto (ADR 0053, punto 2), con intervalos de Wilson
al 95 %:

- en cada modo, `plazos-y-territorio-producto` no queda por debajo de `plazos-y-territorio-control` en las sesiones que
  activan la skill antes de la primera herramienta ni en ninguna referencia a cuya condición llegan las evals. «Por
  debajo» quiere decir con un intervalo que no se toca con el de `control`;
- en `municipios.md`, `producto` queda por encima de `control` en los dos modos, con intervalos que no se tocan.

El run no lanza el sondeo. La tabla y su lectura se añaden a la propuesta de cambio antes de fusionarla, y la propuesta
no se fusiona sin ellas. Si no cumple, no se fusiona, y lo medido va a un ADR que dice qué medir después.

## Assumptions

- Las referencias «a cuya condición llegan las evals» en el sondeo son `administrativo.md`, `calculo_de_plazos.md` y
  `municipios.md`; a la de `jerarquia_normativa.md` no llega ninguna de las cuatro (research D15 de H41; ADR 0053, «Lo
  medido»).
- Los recuentos de caracteres del cuerpo son los de `specs/034-h41-las-skills-del/contracts/variantes.md` §3, con la
  misma medida que `TestSkillsDelRepositorio`; esta sesión no los ha vuelto a medir.
- `plazos-y-territorio-producto` no tiene directorio en el laboratorio: es `skills/plazos-y-territorio` en cada sondeo
  (`specs/034-h41-las-skills-del/contracts/variantes.md` §1), así que el arnés mide la forma aplicada sin cambiar.
- La tabla del sondeo va a un directorio `evidencias/sondeos-<fecha>/`, como los de H40 y H41.
- Las dos frases de §1 de `docs/SKILLS.md` que dicen que `plazos-y-territorio` queda fuera «hasta H41.1» se corrigen
  (FR-009) porque describen el control que el hito cambia; no es contenido nuevo de la guía.
- FR-013 modifica un fichero existente de `testdata/` (`internal/app/testdata/script/h41-skills-instaladas.txtar`): lo
  revisa la persona con el informe final, antes de fusionar (constitución, «Gates», capa 3).

## Fuera de alcance

- Los movimientos M2 y M3, y cualquier otro texto que salga del cuerpo o cambie en él.
- Las formas `condicion-por-orden` y `lista-con-verbo`, y cambiar o retirar ninguna variante del laboratorio; su
  retirada, si llega, la decide quien lea la medida (ADR 0053, «Consecuencias»).
- Cambiar `normativa`, `jurisprudencia` o `consulta-juridica`, sus referencias o sus descripciones, y la descripción de
  `plazos-y-territorio`.
- Cambiar ningún umbral, límite ni tope, tampoco los 11 000 caracteres del cuerpo (ADR 0046, punto 2).
- Cambiar el arnés del laboratorio, sus reglas de lectura, la tabla o el generador de las referencias desde `data/`.
- Evals nuevas, y cambiar las que ya existen.
- La serie de `sin_activar` en la eval 04 de `normativa`, que mide H41.2.
- Medir con el juez en el sondeo, que cuenta lecturas y activación, sin modelo.
- Lanzar el sondeo dentro del run, y cualquier sesión con modelo en él.
- El punto 2 del ADR 0052 en `docs/SKILLS.md`, que sigue siendo una hipótesis.
- El ejemplo de remisión de `docs/SKILLS.md` §1 («Con un apartado «Referencias», una línea por referencia, como
  `plazos-y-territorio`», con la viñeta de `municipios.md`), que tras FR-001 ya no está en la skill y que, con FR-007,
  `TestRemisionesDelRepositorio` da en rojo en una skill del producto: el hito pide en §1 solo la línea de FR-009, y la
  persona lo ve en el informe final. (El barrido global lo cambió después por la viñeta de `jerarquia_normativa.md`,
  que la skill conserva, para que la guía no cite una línea que ya no está; `gates/supuestos.md`.)
- El ADR que haría falta si el sondeo no cumple SC-003: lo escribe quien lea la medida, fuera del run.
- Fusionar, etiquetar o publicar (ADR 0018, 0020 y 0021).
