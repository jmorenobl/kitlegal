# Feature Specification: H45.1 · Los tests del molde, sin el estado de hoy del repositorio (ADR 0062)

**Feature Branch**: `043-h45-1-los-tests-del`

**Created**: 2026-11-24

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H45.1-los-tests-del-molde-sin-el-estado-de-hoy.md` al
lanzarlo (hoy en `docs/roadmap/done/`), que aplica
los puntos 5 (ningún test fija con una lista lo que hay en `data/` o `evals/`) y 9 (H42 entra solo con datos) del ADR
0062 (`docs/ADR/0062-un-paquete-trae-sus-plazos-y-lo-suyo-de-lo-regulado-aparte.md`). Los demás puntos no cambian; el
run no cambia ningún ADR ni escribe otro.

## Contexto

El ADR 0062 promete que el paquete tributario de H42 entra solo con datos: sin un fichero `.go`, de `.github/` ni un
`SKILL.md` escrito a mano. H45 quitó las listas literales de los tests, pero el prototipo de H42 del 2026-11-24, montado
sobre `main` (`95442cf0`) en una copia fuera del repositorio con los datos, las referencias y las evals de la entrada de
H42, enseñó que cuatro tests siguen atados a lo que hay hoy en `data/`, `evals/` y los `SKILL.md`: con esos datos,
`make ci` cae en ellos y en nada más de Go, sin que nada del producto esté mal. Este hito los hace genéricos; con los
datos de `main`, todo sale igual. No cambia el binario, ni una skill, ni una referencia, ni una eval: lo que gana es que
quien escribe un paquete (H42 y los que sigan) no se encuentra con un rojo que no es suyo.

**Comprobado en el repositorio al escribir este spec** (lecturas y órdenes de esta sesión, rama
`043-h45-1-los-tests-del` desde `main`, `e1555a7d`):

- `TestCoberturaDelRepositorio` (`internal/skills/cobertura_test.go:334-400`) escribe las nueve tuplas con el texto de
  cada procedimiento (`:382-397`) y filtra las dos listas, la escrita y la leída, con `deLaBase` (`:349-364`), que mira
  la norma que rige de cada tupla y, sin ella o sin `vertical`, la de su precepto. La lista escrita se filtra con la
  norma que rige que ella misma lleva, no con la de la fila leída.
- `mutantesDeLasOrdenes` (`internal/evals/conjunto_test.go:5307-5337`) escribe «`SKILL.md: línea 36`» en el defecto
  esperado de `forma-en-una-referencia-con-otros-bloques` y `forma-en-ninguna-referencia`. En
  `skills/normativa/SKILL.md`, la orden de Bash del paso 2 está en la línea 36, y la región de los paquetes
  (`<!-- paquetes: lo genera make skills-sync -->`, líneas 25-26) va antes, en el paso 1: hoy vacía. La sincronía
  escribe en ella una línea por paquete con referencia en la skill (`internal/skills/paquetes.go:291-301`).
- `probarCuatroEvalsMas` (`internal/evals/definicion_test.go:3413-3482`) copia todas las evals del repositorio, añade
  cuatro copias de la primera eval de modo que decide de `plazos-y-territorio` y escribe el tope, 352 minutos
  (`topeEnMinutos`, `:3420`).
- `llevaNoCubierto` (`internal/evals/conjunto_test.go:2958-2960`) da por buena toda eval con `no_cubierto` distinto del
  vacío; lo usan `probarLosPreceptosDeLaCobertura` (`:2919-2952`) y `preceptosFueraDeLaCobertura` (`:3199-3219`). Hoy
  ninguna de las 12 evals de `evals/plazos-y-territorio/` es de lo no declarado; sí las hay en `consulta-juridica`.
- El guion `internal/app/testdata/script/h45-paquete-de-mas.txtar` y su copia congelada
  `specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar` son iguales byte a byte (407 líneas, `cmp`). El
  paquete de prueba solo añade: una norma con `vertical: prueba`, un tipo, una fila de lo cubierto, una de lo regulado
  aparte, `prueba.md` en `plazos-y-territorio` y una eval, `99-` más la primera de `evals/plazos-y-territorio/`.
  `TestDefinicionDelJob` ya está entre sus tests enteros; `TestControlesSobreLasReferenciasAMano` no está en ninguna de
  sus dos listas.
- Medido con un test de un solo uso, escrito y borrado en esta sesión, con la definición del repositorio y `peorCaso`:
  una carpeta de `plazos-y-territorio` con 16 copias de `01-territorio-municipio-cubierto.yaml` (la primera eval de modo
  que decide) da un peor caso de 5 h 22 min 45 s a tres sesiones a la vez y de 7 h 47 min 49 s a dos, sobre un tope de
  352 min; con 17, 5 h 40 min 53 s a tres; con 18, 5 h 59 min 1 s a tres, por encima del tope. Las 12 del repositorio,
  a tres, 4 h 1 min 9 s.

## Clarifications

### Session 2026-11-24

Sin preguntas en la redacción: lo que el hito deja a elegir va a «Assumptions», salvo esta decisión:

- Q: ¿Cómo corre `TestDefinicionDelJob` en la copia con un paquete más? → A: por la lista de subtests, al menos
  `cuatro-evals-mas`, que con 16 copias fijas no depende de cuántas evals tenga el módulo; sale de los tests enteros, y
  sobre la copia no corren `del-repositorio`, `peor-caso` ni `tope-con-las-evals-del-repositorio`, que miden contra el
  tope las evals del módulo copiado más las cinco de la copia. Descartado: dejarlo entero, que con las 16 de H42 da 21 evals en la copia, no caben, y el
  guion sale en rojo en el `make ci` de H42, que no puede tocarlo. Es la lectura literal del hito («la lista de
  subtests de la copia gana … `TestDefinicionDelJob`»). (auto)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Con los datos de `main`, todo sale igual (Priority: P1)

Quien lleva el proyecto quiere que arreglar los tests no cambie nada de lo que ve una persona o una skill: con los datos
de `main`, los cuatro tests comprueban lo mismo que hoy y el binario, las skills, las referencias y las evals salen como
en `main`.

**Why this priority**: es la condición para que el hito sea solo de tests; sin ella, H42 no sabría si un rojo es del
molde o del paquete.

**Independent Test**: en el árbol de la rama, `make ci` en verde y `make skills-check` sin diferencias; ningún fichero
del diff contra `main` está en `data/`, `evals/`, `skills/` ni `.github/`, ni es un `.go` que no termina en `_test.go`.

**Acceptance Scenarios**:

1. **Dado** el árbol de la rama, **Cuando** se corre `make ci`, **Entonces** sale en verde.
2. **Dado** el árbol de la rama, **Cuando** se corre `make skills-check`, **Entonces** sale en verde: ninguna referencia
   generada ni ninguna salida del binario distinta de la de `main`.
3. **Dado** el diff de la rama contra `main`, **Cuando** se listan sus ficheros, **Entonces** ninguno está en `data/`,
   `evals/`, `skills/` ni `.github/`, ni es un `.go` que no termina en `_test.go` (el programa `paquete-de-mas.go` no
   es un fichero del diff: vive dentro de los dos guiones `.txtar`).
4. **Dado** el árbol de la rama, **Cuando** se corren `TestControlesSobreLasReferenciasAMano` con `-v`, **Entonces**
   `forma-en-una-referencia-con-otros-bloques` y `forma-en-ninguna-referencia` pasan con el defecto en la línea 36 del
   `SKILL.md` de `normativa`, el mismo mensaje que hoy.

---

### User Story 2 - Un paquete que estrecha una fila de la base o le da una norma no pone en rojo la cobertura (Priority: P1)

Quien escribe un paquete (H42) estrecha el texto de una fila de lo regulado aparte de la base, o le da una norma que la
rige y la fila pasa a la referencia de su paquete; `TestCoberturaDelRepositorio` sigue en verde porque ya no escribe el
texto de cada procedimiento ni filtra lo esperado con una norma escrita a mano. Lo que veía de lo administrativo lo sigue
viendo.

**Why this priority**: es el primero de los cuatro rojos del prototipo de H42 (el sancionador «en materia tributaria y
aduanera», que pasa a ser solo aduanera, y la fila del art. 112.4, que gana la LGT y pasa a `fiscal.md`).

**Independent Test**: en la copia con un paquete más (US6), `TestCoberturaDelRepositorio` en verde; en el árbol, los
mutantes L06 y L07 de H45 la ponen en rojo, y L04 y L05 los ve `h32-paquetes`.

**Acceptance Scenarios**:

1. **Dado** una copia del repositorio en la que una fila de la base sin norma cambia el texto de su procedimiento,
   **Cuando** se corre `TestCoberturaDelRepositorio`, **Entonces** sale en verde.
2. **Dado** una copia del repositorio en la que una fila de la base sin norma gana por norma la de otro paquete y pasa a
   su referencia, **Cuando** se corre `TestCoberturaDelRepositorio`, **Entonces** sale en verde.
3. **Dado** el árbol con el mutante L06 de H45 (fuera la norma de la fila de tráfico, `data/cobertura.yaml:39`),
   **Cuando** se corre `TestCoberturaDelRepositorio`, **Entonces** sale en rojo.
4. **Dado** el árbol con el mutante L07 de H45 (`LeerCobertura` que devuelve cada fila con `Norma = nil`), **Cuando** se
   corre `TestCoberturaDelRepositorio`, **Entonces** sale en rojo.
5. **Dado** el árbol con el mutante L04 (una fila quitada) o L05 (el texto de una fila cambiado) de H45, **Cuando** se
   corre `TestEntregaDelHito/h32-paquetes`, **Entonces** sale en rojo con `no match for ^filas: 33 de 33$`.

---

### User Story 3 - La línea del paso 2 de `normativa` sale de la skill que se comprueba (Priority: P1)

Quien escribe un paquete con referencia en `normativa` añade una línea a la región de los paquetes del cuerpo, antes del
paso 2; los dos casos de `TestControlesSobreLasReferenciasAMano` que esperan el defecto en la orden del paso 2 lo esperan
en la línea en que esa orden está en la skill que comprueban, no en la 36.

**Why this priority**: segundo rojo del prototipo de H42; cae con el primer paquete con referencia en `normativa`.

**Independent Test**: en la copia con un paquete más, con la referencia de su paquete en `normativa`, los dos casos en
verde con el defecto una línea más abajo que en el `SKILL.md` del módulo copiado; en el árbol, en la 36.

**Acceptance Scenarios**:

1. **Dado** una copia del repositorio con un paquete con referencia en `normativa` y la sincronía hecha, **Cuando** se
   corre `TestControlesSobreLasReferenciasAMano`, **Entonces** sale en verde y sus dos casos esperan el defecto en la
   línea en que está la orden de Bash del paso 2 en el `SKILL.md` de esa copia.
2. **Dado** el árbol con el mutante P6 de H45 (`defectosDeLasOrdenes(delSkillMd, nil)`, sin las órdenes de las
   referencias), **Cuando** se corre `TestControlesSobreLasReferenciasAMano`, **Entonces**
   `forma-en-una-referencia-con-otros-bloques` y `forma-en-ninguna-referencia` salen en rojo.
3. **Dado** el árbol con un control que da el defecto de la orden del paso 2 en otra línea que la suya, **Cuando** se
   corre `TestControlesSobreLasReferenciasAMano`, **Entonces** los dos casos salen en rojo.

---

### User Story 4 - El hueco del trabajo de evals se mide con un número fijo de evals (Priority: P1)

`TestDefinicionDelJob/cuatro-evals-mas` deja de probar el hueco que quedaba el día de H45 sobre las evals del
repositorio: mide una copia con 16 evals de `plazos-y-territorio`, copias de una eval de modo, que con la definición del
repositorio cabe en el tope a tres sesiones a la vez y no cabe a dos. Que las evals del repositorio quepan lo siguen
exigiendo `probarElPeorCasoDelTrabajo` (`peor-caso`) y `probarElTopeConLasEvalsDelRepositorio`
(`tope-con-las-evals-del-repositorio`).

**Why this priority**: tercer rojo del prototipo de H42: con sus 16 evals, sumar cuatro da 20, que no caben, aunque las
16 sí.

**Independent Test**: en el árbol y en la copia con un paquete más, `cuatro-evals-mas` en verde; con `plazos-y-territorio`
a dos (T1 de H45) o `peorCaso` con otra cuenta (L29 de H45), en rojo.

**Acceptance Scenarios**:

1. **Dado** el árbol, o una copia con más evals en `plazos-y-territorio`, **Cuando** se corre
   `TestDefinicionDelJob/cuatro-evals-mas`, **Entonces** mide una carpeta de `plazos-y-territorio` con exactamente 16
   evals, todas copias de una eval de modo de las del repositorio, y sale en verde.
2. **Dado** el árbol con el mutante T1 de H45 (`plazos-y-territorio` a dos en `.github/workflows/evals.yml` y en
   `ajustesDelTrabajo`), **Cuando** se corre `TestDefinicionDelJob`, **Entonces** `cuatro-evals-mas` sale en rojo.
3. **Dado** el árbol con el mutante L29 de H45 (`Modos: []Modo{ModoOrden}` en `peorCaso`), **Cuando** se corre
   `TestDefinicionDelJob`, **Entonces** `cuatro-evals-mas` sale en rojo.
4. **Dado** una copia con evals de más en `plazos-y-territorio` que a tres sesiones no caben en el tope, **Cuando** se
   corre `TestDefinicionDelJob`, **Entonces** `tope-con-las-evals-del-repositorio` sale en rojo, como hoy.

---

### User Story 5 - La eval de lo no declarado no se mide como una de lo regulado aparte (Priority: P1)

Una eval de `plazos-y-territorio` con `no_cubierto: {fuera_de_lo_declarado: true}` no lleva precepto, y el esquema no
se lo deja llevar; `precepto-de-la-cobertura` la deja fuera y sigue exigiendo el precepto de una fila de
`data/cobertura.yaml` a las demás con `no_cubierto`, al menos una.

**Why this priority**: cuarto rojo del prototipo de H42: su eval 16 es la primera de lo no declarado en esa skill y da
«no_cubierto.precepto , bloque  no es el precepto de ninguna fila».

**Independent Test**: en la copia con un paquete más, con una eval de lo no declarado en `plazos-y-territorio`, el
subtest en verde; en el árbol, con L33 de H45 (bloque `a999` en la 12), en rojo con su mensaje de hoy.

**Acceptance Scenarios**:

1. **Dado** una copia con una eval de lo no declarado en `plazos-y-territorio`, **Cuando** se corre
   `TestEvalsDelRepositorio/precepto-de-la-cobertura`, **Entonces** sale en verde.
2. **Dado** el árbol con el mutante L33 de H45, **Cuando** se corre el subtest, **Entonces** sale en rojo con
   «12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: no_cubierto.precepto BOE-A-2015-10565, bloque a999 no es
   el precepto de ninguna fila de data/cobertura.yaml».
3. **Dado** un conjunto de `plazos-y-territorio` cuyas evals con `no_cubierto` son todas de lo no declarado, **Cuando**
   se corre el subtest, **Entonces** sale en rojo: no queda ninguna con precepto.

---

### User Story 6 - La copia con un paquete más ejerce los cuatro casos (Priority: P1)

El guion `h45-paquete-de-mas` y su copia congelada pasan de un paquete que solo añade a uno que hace lo que hará H42:
estrecha una fila de la base sin partirla, da norma a otra, lleva una referencia en `normativa` y añade evals de modo y
una de lo no declarado en `plazos-y-territorio`. Corre en la copia, además de su lista de hoy,
`TestControlesSobreLasReferenciasAMano` y, de `TestDefinicionDelJob`, al menos `cuatro-evals-mas` y nunca los que
miden contra el tope las evals del módulo copiado (FR-011), y todo sigue en verde.

**Why this priority**: es la aceptación del hito; sin ella, que los cuatro tests sean genéricos no lo comprueba nada
dentro de `make ci`.

**Independent Test**: `go test -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/` en verde con
`copia: hecha` y sus tres códigos en 0; con cualquiera de los cuatro tests devuelto a su versión de `main`, en rojo.

**Acceptance Scenarios**:

1. **Dado** el árbol de la rama, **Cuando** `make ci` corre `TestEntregaDelHito/h45-paquete-de-mas`, **Entonces** pasa
   y su salida da `copia: hecha`, `sincronía: 0`, `tests enteros: 0` y `subtests: 0`.
2. **Dado** el árbol de la rama con uno de los cuatro tests (`TestCoberturaDelRepositorio`, los dos casos de
   `mutantesDeLasOrdenes`, `cuatro-evals-mas` o `llevaNoCubierto`) devuelto a su versión de `main`, **Cuando** se corre
   `TestEntregaDelHito/h45-paquete-de-mas`, **Entonces** sale en rojo, con `tests enteros: 1` o `subtests: 1` y el test
   devuelto nombrado en la salida de errores.
3. **Dado** la copia que monta el guion, sincronizada, **Cuando** se mira su `skills/plazos-y-territorio/references/`,
   **Entonces** la fila estrechada está en alguna referencia de un paquete con su texto de `c991808` como subcadena, y la
   fila a la que dio norma está en `prueba.md` y no en `administrativo.md`.
4. **Dado** la copia que monta el guion, sincronizada, **Cuando** se mira `skills/normativa/SKILL.md`, **Entonces** la
   región de los paquetes lleva la línea del paquete de prueba y la orden de Bash del paso 2 está una línea más abajo
   que en el `SKILL.md` del módulo copiado.
5. **Dado** los dos ficheros del guion, el de `internal/app/testdata/script/` y el de
   `specs/042-h45-el-molde-admite/aceptacion/`, **Cuando** se comparan, **Entonces** son iguales byte a byte.

---

### Edge Cases

- Un paquete que parte una fila de la base (la estrecha y añade otra suya, como el sancionador tributario de H42): la
  estrechada queda como en US2.1 y la nueva es de otro paquete, que `TestCoberturaDelRepositorio` no mira; la copia lo
  ejerce con la fila estrechada y la fila de lo regulado aparte que el paquete de prueba ya añade.
- La fila de la aplicación de los tributos (DA 1ª.2.a con la LGT): la fijan `TestCasosDePlazos/19-eval-c-tributos` y los
  guiones de H32 y H33, que este hito no toca; la copia no la estrecha ni le cambia la norma, como tampoco a las otras
  dos que cambia H42 (FR-010).
- Con `main` no hay ninguna eval de lo no declarado en `plazos-y-territorio`: el filtro nuevo de `llevaNoCubierto` da
  las mismas evals que hoy.
- Con `main` la región de los paquetes de `normativa` está vacía: la línea derivada es la 36.

## Requirements *(mandatory)*

### Functional Requirements

**`TestCoberturaDelRepositorio`** (`internal/skills/cobertura_test.go`)

- **FR-001**: El test no escribe el texto de ningún procedimiento. Lo esperado son, por cada fila de lo regulado aparte,
  la norma de su precepto, su apartado y la norma que la rige (vacía si no la lleva), obtenidas sin `LeerCobertura`:
  leídas de `data/cobertura.yaml` en crudo o escritas a mano sin el texto del procedimiento. Criterio: ninguna de las
  nueve cadenas de procedimiento de hoy (`cobertura_test.go:384-397`) queda en el cuerpo del test, y la copia de US6 con
  la fila estrechada lo deja en verde.
- **FR-002**: El filtro que decide qué filas son de lo administrativo sigue escrito en el test, sin `paqueteDeLaFila` ni
  `filasDeLaCobertura`, y una fila que deja la base —porque gana por norma la de otro paquete— no pone el test en rojo.
  Criterio: la copia de US6, con la fila a la que da norma ya en `prueba.md`, lo deja en verde.
- **FR-003**: Con los datos de `main`, lo comparado son las nueve tuplas de hoy (norma del precepto, apartado y norma que
  rige), todas, en su orden. Criterio: en el árbol, verde; con L06 de H45, rojo; con L07 de H45, rojo.
- **FR-004**: Los mutantes L04 (una fila quitada) y L05 (el texto de una fila cambiado) de H45 los sigue poniendo en rojo
  `TestEntregaDelHito/h32-paquetes` («filas: 33 de 33» contra `c991808`), que este hito no cambia. Criterio: con cada
  uno, `h32-paquetes.txtar` falla con `no match for ^filas: 33 de 33$`.

**`mutantesDeLasOrdenes`** (`internal/evals/conjunto_test.go`)

- **FR-005**: El número de línea del defecto esperado en `forma-en-una-referencia-con-otros-bloques` y
  `forma-en-ninguna-referencia` sale del `SKILL.md` de `normativa` que el caso comprueba: la línea en que está la orden
  de Bash del paso 2. Ningún número de línea del `SKILL.md` queda escrito para esos dos casos. Criterio: en el árbol, el
  defecto esperado es el de hoy, línea 36, byte a byte; en la copia de US6, verde con el defecto una línea más abajo
  que en el `SKILL.md` del módulo copiado; con P6 de H45, los dos
  casos en rojo; con un control que da la orden del paso 2 en otra línea, los dos en rojo.

**`TestDefinicionDelJob/cuatro-evals-mas`** (`internal/evals/definicion_test.go`)

- **FR-006**: El subtest mide una copia de las evals en la que la carpeta de `plazos-y-territorio` tiene exactamente 16
  evals, todas copias de una eval de modo que decide de las del repositorio; no suma evals a las del repositorio. Con
  la definición del repositorio, su peor caso cabe en el tope a la concurrencia que la definición da a
  `plazos-y-territorio`, y con esa definición y `plazos-y-territorio` a dos no cabe. El subtest sale en rojo si el peor
  caso a la concurrencia del repositorio pasa del tope, o si a dos no pasa. Criterio: verde en el árbol y en la copia de
  US6; rojo con T1 y con L29 de H45.
- **FR-007**: El subtest no escribe minutos ni segundos: el tope sale de la definición leída y la duración, de
  `peorCaso`, comparada con `peorCasoEsperado` como hoy. Criterio: sin la constante `topeEnMinutos` ni otra duración
  literal en el subtest; L29 lo pone en rojo.
- **FR-008**: Que las evals del repositorio quepan en el tope lo siguen exigiendo `probarElPeorCasoDelTrabajo`
  (`peor-caso`) y `probarElTopeConLasEvalsDelRepositorio` (`tope-con-las-evals-del-repositorio`), que este hito no
  cambia. Criterio: en una copia con evals de más que a tres no caben, `tope-con-las-evals-del-repositorio` en rojo.

**`llevaNoCubierto`** (`internal/evals/conjunto_test.go`)

- **FR-009**: `llevaNoCubierto` deja fuera las evals con `no_cubierto.fuera_de_lo_declarado`; las demás con
  `no_cubierto` siguen exigiendo por precepto el de una fila de `data/cobertura.yaml`, y el subtest sigue exigiendo al
  menos una. El mutante `a999` (L33 de H45) no cambia. Criterio: verde en el árbol y en la copia de US6, con su eval de
  lo no declarado; rojo con L33 y su mensaje de hoy; rojo si todas las evals con `no_cubierto` son de lo no declarado.

**La copia con un paquete más** (`internal/app/testdata/script/h45-paquete-de-mas.txtar` y
`specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`)

- **FR-010**: El paquete de prueba de la copia, además de lo que añade hoy: (a) estrecha el texto de una fila de lo
  regulado aparte de la base, sin norma, sin partirla y de forma que su texto de `c991808` sigue como subcadena;
  (b) da por norma que la rige la del paquete de prueba a otra fila de la base sin norma, que pasa a `prueba.md`;
  (c) lleva una referencia en `normativa`, con lo que su línea entra en la región de los paquetes del cuerpo; y (d)
  añade en `plazos-y-territorio` cuatro evals de modo y una de lo no declarado. (a) y (b) van sobre filas de la base
  que el hito no dice que cambie H42: ni la de la aplicación de los tributos, ni la del sancionador «en materia
  tributaria y aduanera», ni la del art. 112.4. Criterio: los escenarios 3 y 4 de US6, comprobados por el guion sobre
  la copia.
- **FR-011**: La lista de subtests que el guion corre en la copia gana `TestControlesSobreLasReferenciasAMano` y, de
  `TestDefinicionDelJob`, al menos `cuatro-evals-mas`; `TestDefinicionDelJob` sale de los tests enteros, y sobre la
  copia no corren `del-repositorio`, `peor-caso` ni `tope-con-las-evals-del-repositorio`, que miden contra el tope las
  evals del módulo copiado. La
  sincronía, los tests enteros y los subtests salen con 0. Criterio: el escenario 1 de US6.
- **FR-012**: Con cualquiera de los cuatro tests devuelto a su versión de `main`, el guion sale en rojo. Criterio: el
  escenario 2 de US6, uno por test, en el registro de mutantes del hito.
- **FR-013**: Los dos ficheros del guion quedan iguales byte a byte, y el programa `paquete-de-mas.go` y los
  `paquete/*.yaml` viven dentro de ellos. Criterio: `cmp` de los dos ficheros, sin diferencias.

**Lo de `main`**

- **FR-014**: Con los datos de `main`, todos los tests en verde y ninguna salida del binario ni referencia distinta.
  Criterio: `make ci` y `make skills-check` en verde en el árbol; el diff contra `main` cumple el escenario 3 de US1.

### Key Entities

- **Tupla de lo regulado aparte**: norma del precepto, apartado y norma que rige de una fila de `regulados_aparte` de
  `data/cobertura.yaml`, sin el texto del procedimiento.
- **Paquete de prueba de la copia**: el de `h45-paquete-de-mas`, `vertical: prueba`, con la norma de tráfico
  (`BOE-A-2015-11722`), sus entradas en `data/`, sus referencias en `plazos-y-territorio` y en `normativa` y sus evals.
- **Orden de Bash del paso 2**: la línea de `skills/normativa/SKILL.md` con
  `kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check BOE-A-1978-31229 a140 --json`.

### Quién consume cada salida, cuánto y hasta cuándo

El hito no cambia ninguna salida del binario, ni una skill, ni una referencia, ni una eval (FR-014): ninguna skill pide
nada nuevo, y por pregunta de un usuario se invoca cero veces. Lo que cambia lo leen dos consumidores, ninguno de ellos
en tiempo de consulta:

- **Los mensajes de los cuatro tests**: los lee quien escribe un paquete, persona o run (H42 y los siguientes), una vez
  por `make ci`; uno por defecto, con el fichero, la eval o la línea que falla, unos 200 bytes. Leen `data/`, `evals/` y
  `skills/`, no la caché ni el grafo: con meses de uso, cientos de normas y miles de bloques consultados, son los
  mismos. Cada uno deja de darse cuando el dato que nombra se corrige; con un paquete correcto no se da.
- **La salida del guion de la copia**: la lee `TestEntregaDelHito` en cada `make ci`; cinco líneas, más el principio de
  la salida de cada orden que no sale con 0, hasta 1 200 bytes cada una (como hoy). No depende del uso. Deja de dar un
  código distinto de 0 en cuanto el test que cae vuelve a ser genérico.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el árbol de la rama, `make ci` en verde. Control: `make ci`.
- **SC-002**: En el árbol de la rama, `make skills-check` en verde: cero diferencias en las referencias generadas y en
  las salidas del binario respecto de `main`. Control: `make skills-check`, que sale en rojo con una sola diferencia.
- **SC-003**: `TestEntregaDelHito/h45-paquete-de-mas` en verde dentro de `make ci`, con `copia:
  hecha`, `sincronía: 0`, `tests enteros: 0` y `subtests: 0` entre sus cinco líneas (la quinta, `orden del paso 2 de
  normativa, líneas más abajo: 1`). Control: el guion, que falla en `make ci` si cualquiera de
  los tres códigos no es 0.
- **SC-004**: Cada uno de los cuatro tests, en rojo con su mutante de hoy —L06 y L07 para `TestCoberturaDelRepositorio`;
  P6 para los dos casos de `mutantesDeLasOrdenes`; T1 y L29 para `cuatro-evals-mas`; L33 para
  `precepto-de-la-cobertura`— y en verde en la copia con un paquete más. Control: el registro de mutantes del hito, con
  la orden, lo esperado y lo visto de cada uno.
- **SC-005**: Con cada uno de los cuatro tests devuelto a su versión de `main`, uno a uno, `h45-paquete-de-mas` en rojo:
  la copia ejerce los cuatro casos. Control: el mismo registro (lo ejecuta una tarea del run, no `make ci`).
- **SC-006**: `cuatro-evals-mas` mide exactamente 16 evals de `plazos-y-territorio`; a la concurrencia del repositorio
  su peor caso no pasa del tope de la definición, y a dos sí. Control: el propio subtest, que sale en rojo en `make ci`
  si la primera medida pasa del tope o la segunda no.
- **SC-007**: L04 y L05 de H45, en rojo en `h32-paquetes` con `no match for ^filas: 33 de 33$`. Control: el registro de
  mutantes.

## Criterios del hito (literales)

- **Controles** (del hito): «cada test, en rojo con su mutante de hoy y en verde en la copia con un paquete más; con
  `main`, todos los tests en verde y ninguna salida del binario ni referencia distinta (`make skills-check`)». En este
  spec: SC-004, SC-002 y SC-001; FR-003, FR-005, FR-006, FR-009 y FR-014.
- **Aceptación** (del hito): «la copia `h45-paquete-de-mas` con los cuatro casos, en verde dentro de `make ci`; `make ci`
  en verde con `main`». En este spec: SC-003, SC-001; US6; FR-010 a FR-013.
- «Que el prototipo de H42 pase sin ningún `.go` lo comprueba la persona antes de lanzar H42»: fuera de este run (ver
  «Fuera de alcance»).

## Assumptions

- **Las cuatro evals de modo de la copia incluyen la `99-` de hoy**: la copia añade a las del módulo copiado (12 con
  `main`) cinco evals, cuatro copias de una eval de modo (la `99-` de hoy entre ellas) y una de lo no declarado. Es la
  lectura que menos añade del «añade cuatro evals de modo y una de lo no declarado». Con `cuatro-evals-mas` de `main`,
  la copia de `main` mide 17 + 4 = 21 evals, que no caben a tres, y el guion sale en rojo (FR-012).
- **La norma que se da a la fila de la base es la del paquete de prueba** (`BOE-A-2015-11722`, `vertical: prueba`), para
  que la fila pase a `prueba.md` como la del art. 112.4 pasa a `fiscal.md` con H42: es el caso que hoy pone en rojo
  `TestCoberturaDelRepositorio`. Su precepto aplicable es uno ya grabado de esa norma.
- **«La lista de subtests gana `TestDefinicionDelJob`»**: lo decide la Clarification (auto) de la sesión 2026-11-24
  (FR-011): sobre la copia no corre ningún subtest que mida contra el tope las evals del módulo copiado
  (`del-repositorio`, `peor-caso`, `tope-con-las-evals-del-repositorio`). `TestControlesSobreLasReferenciasAMano` se añade a la lista, entero o con sus dos casos.
- **Del esperado de `TestCoberturaDelRepositorio`**: el hito admite dos formas («leídas del YAML en crudo, o escritas a
  mano sin el texto del procedimiento»); el plan elige la que cumple a la vez FR-001 a FR-003, y en particular que L06
  siga en rojo en este test.
- **El guion de la copia vive en `testdata/`**: su cambio, y el de su copia congelada de H45, van en una tarea `[datos]`
  (constitución, guardián de diff), y la persona los revisa al leer el informe (capa 3: modificación de ficheros de
  `testdata/`).
- Los mutantes L04 a L07, L29, L33, T1 y P6 son los de `specs/042-h45-el-molde-admite/mutantes.md`, con su orden.

## Dependencias

- H45 (fusionado en `main`): el molde, la copia `h45-paquete-de-mas` y los mutantes que este hito reutiliza.
- H42 depende de este hito.

## Fuera de alcance

- Cualquier cambio de producción en Go (todo `.go` que no termina en `_test.go`; `paquete-de-mas.go` vive dentro de los
  guiones y es de la copia), en `data/`, en `evals/`, en las skills (`skills/`) o en `.github/`.
- Los guiones de H32 y H33 que fijan la fila de los tributos (`h32-paquetes`, `h33-cobertura-instalada` y sus copias
  congeladas): los cambia H42 con sus datos.
- El tope y la concurrencia del trabajo de evals.
- Los tests de la lista de H45 que no son estos cuatro, y `TestNormasDelRepositorio/paquete-de-lo-administrativo`.
- Comprobar en las evals de lo no declarado algo más que dejarlas fuera de `precepto-de-la-cobertura`.
- Los datos, las referencias y las evals de H42, y comprobar que su prototipo pasa sin ningún `.go`: lo hace la persona
  antes de lanzar H42.
- Cambiar los puntos del ADR 0062 o escribir otro ADR.
