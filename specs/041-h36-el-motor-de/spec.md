# Feature Specification: H36 · El motor de cómputo descrito por datos (ADR 0061)

**Feature Branch**: `041-h36-el-motor-de`

**Created**: 2026-11-22

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H36-el-motor-de-computo-descrito.md`, que aplica el ADR 0061
(`docs/ADR/0061-el-computo-de-un-plazo-se-describe-con-datos.md`) y precisa el punto 3 del ADR 0042.

## Contexto

`kitlegal plazos calcular` cuenta hoy con dos cómputos fijos en Go, `lpac` y `contencioso`. El código decide por el
código de un tipo o de un cómputo qué cómputo se aplica, que el silencio da el día siguiente al vencimiento, qué tipos
fijan su duración, cuál lleva un precepto por unidad, cuántos municipios cuentan y el aviso de los días del art. 183
LOPJ (`internal/core/plazos`: `computar`, `TipoSilencio`, `fijaSuDuracion`, la regla de `unidades` de `LeerTipos`,
`municipiosDelComputo`, `municipiosQueNoCuentan` y `avisos`). `LeerTipos` exige los siete códigos de H9 y ninguno más,
y la ayuda de los argumentos de `calcular` (`internal/app/plazos.go`) y los enumerados de su salida
(`internal/core/plazos/salida.go`) los nombran en etiquetas de struct. Un plazo de otra materia no puede entrar como
datos: pediría una rama del código y un esquema nuevo de la herramienta, contra el principio X.

H36 describe los cómputos y los tipos en `data/plazos.yaml`, con los tres ejes de un cómputo y los cuatro de un tipo
que funda el ADR 0061 en preceptos leídos, y valida el motor con una segunda implementación, el cómputo civil de la
LEC (arts. 130.2 y 133 LEC, `BOE-A-2000-323`). Con los datos de hoy no cambia nada de lo que ve quien usa el binario:
el `data` de cada plazo, su `hash`, los avisos, los errores y sus mensajes salen iguales, byte a byte. Lo que cambia es
la forma de los datos, las reglas de `LeerTipos`, la API de tipos y cómputos de `internal/core/plazos` y el esquema de
`calcular`, que deja de depender de los datos. El cómputo civil no lo pide todavía ningún tipo: sus tipos procesales
son de H9.1.

**Comprobado en el repositorio al escribir este spec** (lecturas de esta sesión):

- `data/plazos.yaml` (rama `main`, `82d4ccdd`): siete tipos y dos cómputos; el art. 128.2 LJCA es un precepto del tipo
  `contencioso` (`fija: agosto`) y el art. 183 LOPJ, del cómputo `contencioso` (`fija: aviso`).
- El texto del aviso del art. 183 y los dos mensajes de municipios que no cuentan son literales de
  `internal/core/plazos/avisos.go` y `errores.go`; los reproduce §C.5 y §C.6.
- La ayuda de `tipo`, `inicio`, `municipios`, `duracion` y `unidad` (`internal/app/plazos.go`, líneas 133 a 137) nombra
  los tipos `alzada` a `generico` y el cómputo `contencioso`; el enumerado `tipo` de `Plazo` (`salida.go`, línea 24)
  repite los siete códigos. Ningún otro literal ni etiqueta de struct de `internal/core/plazos`, `internal/app`,
  `internal/skills` e `internal/evals` es, o lleva como palabra entera, un código de tipo o de cómputo (`grep -w`).
- La tabla de comandos de `skills/plazos-y-territorio/SKILL.md` (línea 143) lleva la forma de uso y la descripción del
  verbo, no la ayuda de los argumentos; cuántos municipios cuentan lo dicen
  `skills/plazos-y-territorio/references/calculo_de_plazos.md` (líneas 9 y 10) y `SKILL.md` (línea 58); la fecha de
  inicio del silencio, `references/administrativo.md` («El silencio»); los tipos, su duración y su cómputo, la región
  `recursos_y_plazos` de `references/administrativo.md`.
- La copia congelada de `internal/app/testdata/script/h39-describe.txtar` es
  `specs/031-h39-c-omo-se/aceptacion/describe.txtar`, y sus líneas 39, 42 y 43 comparan la ayuda de hoy.
- La rejilla de las propiedades (`internal/core/plazos/propiedades_test.go`): inicios de 2025 a 2027, duraciones de 1 a
  40, municipios `81001` «Lugar Uno» y `81002` «Lugar Dos» (comunidad `81`, con fichero local) y `82001` «Lugar Tres»
  (comunidad `82`, sin fichero local); nacionales `01-01`, `01-06`, `05-01`, `08-15`, `10-12`, `11-01`, `12-06`,
  `12-08`, `12-25`; autonómicos de la `81` `03-19` y `05-02`, de la `82` `04-23` y `09-08`; de una parte de la `82`
  («Isla de la Rejilla») `06-24`; locales de `81001` `05-15` y `11-09`, de `81002` `06-13` y `10-09`; cada año igual.
- Los días de la semana de los casos de §C.9 y §C.10 se han comprobado con `date` en esta sesión.

## Clarifications

Ninguna en esta sesión: el fichero del hito y el ADR 0061 fijan los ejes, las reglas de `LeerTipos`, el error nuevo,
el cómputo civil, sus casos, los controles y la aceptación. Los nombres de las claves, que el hito deja al spec, los
fija §C.1.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Con los datos de hoy, `plazos calcular` da lo mismo, byte a byte (Priority: P1)

La skill `plazos-y-territorio` (paso 9) y quien usa el binario piden un plazo con los mismos argumentos que en `main`
y reciben el mismo sobre, salvo `fecha_consulta`: la misma fecha, los mismos preceptos, festivos, avisos, cobertura y
`hash`, y en un error el mismo código, la misma clase y el mismo mensaje.

**Why this priority**: es la condición del hito («no cambia lo que se calcula») y la prueba de que describir los
cómputos con datos no ha cambiado ninguno.

**Independent Test**: los goldens capturados con el binario de `main` (`cd3ad148`) y los de H9, comparados con el
binario de la rama en `make ci` (FR-021).

**Acceptance Scenarios**:

1. **Dado** el binario de la rama con los datos de hoy, **Cuando** se calculan los casos de fecha de
   `TestCasosDePlazos` (del 1 al 18 y el 21), **Entonces** cada uno da la misma fecha y los mismos códigos de aviso que
   con el binario de `main`.
2. **Dado** el sobre de los casos 11 (`silencio 2026-07-09 Leganés Leganés --duracion 3 --unidad meses`), 15
   (`contencioso 2026-10-27 Madrid`) y 16 (`contencioso 2026-10-31 Madrid`) capturado con el binario de `main`, **Cuando**
   la rama ejecuta la misma orden, **Entonces** el sobre es el mismo, byte a byte, salvo `fecha_consulta`.
3. **Dado** el sobre de error de cada orden de §C.6 capturado con el binario de `main`, **Cuando** la rama ejecuta la
   misma orden, **Entonces** el código, el sobre y el mensaje son los mismos, salvo `fecha_consulta`.
4. **Dado** un cambio de una palabra en el texto del aviso del art. 183 en `data/plazos.yaml`, **Cuando** corre `make
   ci`, **Entonces** falla el golden del caso 15.
5. **Dado** el tipo `silencio` con `da: vencimiento` en lugar de `dia-siguiente`, **Cuando** corre `make ci`,
   **Entonces** falla el golden del caso 11.
6. **Dado** el texto del municipio del cómputo `contencioso` cambiado en `data/plazos.yaml`, **Cuando** corre `make ci`,
   **Entonces** falla el golden del error de municipios que no cuentan en `contencioso`.

---

### User Story 2 - Un cómputo de otra materia entra como datos: el civil de la LEC (Priority: P1)

Quien escribe el paquete de una materia (H9.1, H42) describe su cómputo en `data/plazos.yaml` con sus preceptos, y el
motor lo cuenta sin una rama nueva del código. El cómputo civil de la LEC lo prueba: un municipio, días hábiles y
meses, y agosto y del 24 de diciembre al 6 de enero inhábiles.

**Why this priority**: es el objetivo del hito; la segunda implementación exige al diseño lo que la LPAC no tiene
(periodos inhábiles, un municipio, una unidad que no se admite).

**Independent Test**: los goldens C1 a C10 de §C.9, calculados con el cálculo del dominio, los festivos de
`data/festivos/`, el cómputo `civil` del `data/plazos.yaml` real y dos tipos de prueba (FR-023).

**Acceptance Scenarios**:

1. **Dado** el cómputo `civil` de `data/plazos.yaml`, un tipo de prueba en días hábiles y la sede Madrid (`28079`),
   **Cuando** se calcula desde el 2026-07-24 con 10 días hábiles, **Entonces** la fecha es el 2026-09-07, sin festivos
   saltados ni avisos (C1).
2. **Dado** el mismo cómputo y un tipo de prueba en meses, **Cuando** se calcula desde el 2026-06-15 con 2 meses,
   **Entonces** la fecha es el 2026-09-01 (C3).
3. **Dado** el mismo cómputo, **Cuando** se calcula cada fila de §C.9, **Entonces** da la fecha, los festivos saltados,
   los avisos o el código de su fila.
4. **Dado** agosto como `aviso` en lugar de `inhabil`, **Cuando** se calcula C1, **Entonces** da el 2026-08-07 y el
   golden falla; **dado** agosto como `no-corre`, C3 da el 2026-09-15; **dado** el cómputo sin el periodo del 24 de
   diciembre al 6 de enero, C2 da el 2026-12-28; **dado** agosto desde el día 2, C9 da el 2025-08-01; **dado** el
   calendario sin los festivos locales de la sede, C10 da el 2026-11-09. En los cinco, el test falla.
5. **Dado** el `data/plazos.yaml` de la rama, **Cuando** lo lee `LeerTipos`, **Entonces** ningún tipo leído lleva el
   cómputo `civil`, así que ninguna orden `kitlegal plazos calcular` cuenta con él (FR-014).

---

### User Story 3 - Un cómputo de ensayo entra sin tocar Go (Priority: P1)

Un cómputo inventado, que no es de ninguna materia, con su periodo inhábil, su periodo de aviso y dos tipos, se lee y
se cuenta con el mismo motor, y ningún Go que no sea de test nombra sus códigos ni los de `data/plazos.yaml`.

**Why this priority**: prueba que cada eje es un dato y no una rama del código (principio X); sin él, los dos cómputos
de hoy y el civil podrían seguir decididos por su código.

**Independent Test**: el fichero de §C.10 en `internal/core/plazos/testdata/`, sus casos E1 a E13 sobre la rejilla, y
el test de `go/ast` de FR-010.

**Acceptance Scenarios**:

1. **Dado** el fichero de ensayo leído por `LeerTipos`, **Cuando** se calculan E1 a E13 sobre la rejilla, **Entonces**
   cada uno da lo de su fila de §C.10.
2. **Dado** un motor que solo reconoce agosto como periodo, **Cuando** se calcula E1, **Entonces** da el 2026-04-07 y el
   test falla.
3. **Dado** un motor que ignora `dia-siguiente` fuera del tipo del silencio, **Cuando** se calcula E5, **Entonces** da
   el 2026-06-08 y el test falla.
4. **Dado** un motor que no comprueba las unidades del cómputo, **Cuando** se calcula E12, **Entonces** da una fecha en
   lugar del código 2 y el test falla.
5. **Dado** un motor que da los avisos de periodo con un texto fijo, **Cuando** se calcula E7, **Entonces** el texto
   del aviso no es el de §C.10 y el test falla.
6. **Dado** el Go de la rama, **Cuando** corre el test de FR-010, **Entonces** no encuentra ningún literal que sea un
   código de tipo, de cómputo o de aviso de periodo de `data/plazos.yaml` o del fichero de ensayo, ni ninguna etiqueta
   de struct que lo lleve como palabra entera; **dado** `if tipo.Codigo == "silencio"` en `internal/core/plazos`, o la
   ayuda de `calcular` de hoy, el test falla.

---

### User Story 4 - El esquema de `calcular` no depende de los datos (Priority: P2)

El agente que conecta con el servidor MCP recibe el esquema de `plazos_calcular`, y quien pide `--help` o `--describe`
lo lee: no nombra tipos ni cómputos, remite a las referencias del paquete y sale igual con cualquier `data/plazos.yaml`,
también con uno que no carga.

**Why this priority**: el esquema de la herramienta es el protocolo con que el agente llama al binario; si cambiara
con cada paquete de materia, cada paquete cambiaría la herramienta (ADR 0061, punto 7).

**Independent Test**: `--describe` de `calcular` con tres juegos de datos, por el mecanismo con que `plazos_test.go` da
al applet unas fuentes que no cargan (caso «sin-los-tipos»).

**Acceptance Scenarios**:

1. **Dado** el binario de la rama, **Cuando** se pide `kitlegal plazos calcular x x x --describe` con los datos de hoy,
   con una copia de `data/plazos.yaml` con un tipo y un aviso de periodo más y con un `data/plazos.yaml` que no carga,
   **Entonces** las tres salidas son iguales, byte a byte, y son `internal/app/testdata/plazos/describe.json`; las dos
   últimas salen con 0.
2. **Dado** el mismo binario, **Cuando** se pide `kitlegal plazos calcular --help` con un `data/plazos.yaml` que no
   carga, **Entonces** sale con 0, y solo ejecutar el verbo sale con 1 (`inesperado`).
3. **Dado** el esquema de salida, **Cuando** se lee `tipo` y `avisos[].codigo`, **Entonces** llevan el `pattern` de
   §C.8 y ningún `enum`.
4. **Dado** la ayuda o los enumerados de hoy en las etiquetas, **Cuando** corre `make ci`, **Entonces** fallan el golden
   de `describe.json` y el test de FR-010.

---

### User Story 5 - `LeerTipos` lee y comprueba la forma nueva (Priority: P2)

Quien escribe `data/plazos.yaml` recibe, al cargar el binario o en `make ci`, todos los defectos del fichero a la vez,
cada uno con el tipo o el cómputo que lo tiene, con las reglas de §C.2.

**Why this priority**: un dato mal escrito daría una fecha equivocada sin aviso; cada regla de §C.2 cierra una vía.

**Independent Test**: subtests con copias rotas de `data/plazos.yaml`, una por regla (FR-013).

**Acceptance Scenarios**:

1. **Dado** `data/plazos.yaml` de la rama, **Cuando** lo lee `LeerTipos`, **Entonces** devuelve los siete tipos de hoy
   y ningún defecto, con el cómputo `civil` sin tipos.
2. **Dado** una copia con un tipo que nombra un cómputo que no existe, **Cuando** la lee `LeerTipos`, **Entonces** el
   error nombra ese tipo y ese cómputo; y lo mismo con cada regla de §C.2, cada una en su copia.
3. **Dado** una copia con dos defectos de reglas distintas, **Cuando** la lee, **Entonces** el error los da los dos.
4. **Dado** `LeerTipos` sin una de las reglas de §C.2, **Cuando** corre su subtest, **Entonces** falla.

---

### User Story 6 - Los preceptos del cómputo civil se comprueban contra lo grabado (Priority: P3)

`TestPreceptosDeLosPlazos` lee también los cómputos sin tipos, y los cinco preceptos del civil están grabados, llevan
sus palabras clave y su `vigente_desde` es el que dan sus redacciones grabadas (ADR 0042).

**Why this priority**: la regla del ADR 0042 solo protege si el `vigente_desde` es verdad; un cómputo sin tipos no lo
miraba nadie.

**Independent Test**: `TestPreceptosDeLosPlazos` en `make ci`, con las grabaciones de la tarea `[datos]` (FR-016).

**Acceptance Scenarios**:

1. **Dado** lo grabado de `a130` y `a133` de `BOE-A-2000-323`, **Cuando** corre `TestPreceptosDeLosPlazos`, **Entonces**
   pasa con los cinco preceptos del civil.
2. **Dado** el `vigente_desde` del 133.1 con el 2010-05-04, o el del 130.2 con el 2015-10-07, **Cuando** corre,
   **Entonces** falla con el defecto de ese precepto.

### Edge Cases

- **Un periodo que cruza el año** (del 24 de diciembre al 6 de enero): cuenta igual en los dos años que toca (C2, C4).
- **Un festivo dentro de un periodo `inhabil`** que no cae en sábado ni domingo se da como festivo saltado, con la regla
  de hoy (C2, C4, C9); en sábado o domingo, no (C3).
- **Un festivo de una parte dentro de un periodo `inhabil`**: el día es inhábil, así que no avisa, con la regla de hoy
  («uno en un día inhábil por otra causa no avisa»).
- **Un plazo que pasa varias veces por un periodo `aviso`**: un solo aviso de ese periodo por plazo, como el del art.
  183 hoy.
- **`dia-siguiente` que cae en un día inhábil**: no se prorroga, como el silencio hoy (research D8 de H9; E6).
- **El inicio en un periodo `inhabil`**: el cómputo empieza igual el día siguiente; los días del periodo no cuentan
  (E3).
- **Un `data/plazos.yaml` que no carga** o con cualquier defecto: no hay tipos; `--help` y `--describe` responden con 0
  y ejecutar sale con 1 (`inesperado`), como hoy. Lo que llega al binario de fuera manipulado a mano sigue la regla
  genérica (defecto `inesperado`, código 1; ADR 0023), sin casos propios.

## Requirements *(mandatory)*

### Functional Requirements

**Los datos**

- **FR-001**: `data/plazos.yaml` MUST describir cada cómputo con los tres ejes del ADR 0061 (punto 2) y ninguno más, con
  las claves de §C.1: los municipios que cuentan, uno o dos, con el texto que nombra a cada uno; las unidades que
  admite, de las tres del motor; y sus periodos del año, cada uno con su precepto y su efecto, `inhabil`, `no-corre` o
  `aviso`; el de `aviso`, con el código, la etiqueta y el texto del aviso. Cada cómputo lleva además sus preceptos,
  como hoy. Comprobable: US5.1 y la lectura de §C.1.
- **FR-002**: `data/plazos.yaml` MUST describir cada tipo con los ejes del ADR 0061 (punto 3), con las claves de §C.1:
  su cómputo; su duración fija o, si no la tiene, la que recibe con una unidad de su cómputo; si la lleva, un precepto
  por unidad de su cómputo; y la fecha que da, `vencimiento` o `dia-siguiente`. Comprobable: US5.1.
- **FR-003**: `lpac` y `contencioso` MUST pasar a ser datos con la forma de §C.1, y con ellos los siete tipos de hoy,
  sin cambiar lo que calculan (FR-011). `contencioso` admite solo `meses`. El art. 128.2 LJCA sigue entre los preceptos
  del tipo `contencioso` y en su `data.preceptos`. Comprobable: US1 y el contenido de §C.1.

**El motor**

- **FR-004**: `internal/core/plazos` MUST contar cualquier cómputo de los datos con sus ejes y con lo que el ADR 0061
  (punto 4) deja fijo en el motor: el día inicial es el siguiente al de inicio; los meses, de fecha a fecha, con el
  último del mes si no hay día equivalente; un último día inhábil se prorroga al primer hábil; y son inhábiles de base
  el sábado, el domingo y los festivos nacionales, autonómicos y locales de los municipios que cuentan. Los efectos de
  los periodos son los de §C.3. Comprobable: US1, US2, US3 y FR-024.
- **FR-005**: La fecha que se da MUST ser el vencimiento en un tipo con `da: vencimiento`, y el día siguiente al
  vencimiento, sin prórroga, en uno con `da: dia-siguiente`, cualquiera que sea su código o su cómputo. Comprobable:
  US1.5 (caso 11) y E5, E6.
- **FR-006**: El número de municipios que pide un tipo MUST ser el de su cómputo, y el mensaje de un número distinto
  MUST componerse, al ejecutar, con el texto que el cómputo da a cada municipio, con la forma de §C.6; con los datos de
  hoy, los dos mensajes salen iguales que en `main`. Comprobable: US1.3, US1.6, C8 y E13.
- **FR-007**: Una unidad dada a un tipo sin duración que su cómputo no admite MUST salir con 2 (`argumentos`), con el
  mensaje de §C.6, comprobada detrás de los errores de la unidad y delante del número de municipios, en el orden de
  research D10 de H9. Con los datos de hoy no se da: `lpac` admite las tres unidades y `contencioso` no recibe
  duración. Comprobable: C7 y E12.
- **FR-008**: Un periodo `aviso` MUST dar su aviso si el tramo del plazo —del día siguiente al de inicio a la fecha que
  se da— pasa por algún día del periodo, una vez por plazo, con el texto `⚠ <etiqueta>: <texto>` de sus datos. Los
  avisos de periodo MUST ir detrás de los dos de los festivos que faltan y delante de los de un festivo de una parte, en
  el orden de los periodos de su cómputo en el fichero. Comprobable: US1.4 (caso 15), E7 a E11.
- **FR-009**: `CodigosDeAviso` y `EtiquetaDeAviso` MUST recibir los tipos leídos y dar los tres avisos del motor
  (`faltan-festivos-nacionales-y-autonomicos`, `faltan-festivos-locales`, `festivo-de-una-parte`) y los de los periodos
  de los datos: los dos de los festivos que faltan, los de periodo en el orden de los cómputos y de sus periodos en el
  fichero, y el de un festivo de una parte. `internal/evals` MUST tomar de ahí los códigos y las etiquetas de los avisos, y
  `ComprobarCodigosDeAvisoDelPlazo` sigue exigiendo que `schemas/eval.yaml.json` enumere exactamente esos códigos, que
  con los datos de hoy son los cuatro de hoy en su orden. Comprobable: los tests de `internal/evals` de FR-017 y el
  que comprueba `schemas/eval.yaml.json` con `ComprobarCodigosDeAvisoDelPlazo` (`formato_test.go`), en `make ci`.
- **FR-010**: Ningún Go que no sea de test de `internal/core/plazos`, `internal/app`, `internal/skills` e
  `internal/evals` MUST nombrar un tipo, un cómputo ni el aviso de un periodo: un test que recorre ese Go con `go/ast`
  falla si un literal de cadena es, entero, el código de un tipo, de un cómputo o de un aviso de periodo de
  `data/plazos.yaml` o del fichero de ensayo (§C.10), o si una etiqueta de struct lo lleva como palabra entera. Quedan
  fuera los tres avisos del motor y la descripción del verbo `calcular`, que es de H9.1. Desaparecen las constantes
  `Tipo…`, `Computo…`, `AvisoDiasDelArt183`, `CodigosDeTipo` y `Computos`. Comprobable: US3.6.
- **FR-011**: Con los datos de hoy, el `data` de cada plazo, su `hash`, los avisos, su orden, los errores y sus mensajes
  MUST ser los de `main`, byte a byte, y ninguna clave ni valor nuevo entra en el `data`, en los avisos ni en los
  mensajes. Comprobable: US1 y FR-021.

**El esquema de `calcular`**

- **FR-012**: El esquema de `calcular` MUST no depender de los datos (ADR 0061, punto 7):
  - la ayuda de `tipo`, `inicio`, `municipios`, `duracion` y `unidad` es la de §C.7: no nombra tipos ni cómputos, dice
    que el tipo, cuántos municipios cuentan, qué es la fecha de inicio y qué tipos reciben su duración están en las
    referencias del paquete, y sigue nombrando las tres unidades del motor;
  - `tipo` y `avisos[].codigo` de la salida llevan el `pattern` de §C.8 y no un `enum`; los demás enumerados de la
    salida (`festivos[].ambito`, `cobertura[].ambito`, `cobertura[].falta`) no cambian;
  - `--help` y `--describe` no cargan los datos: con un `data/plazos.yaml` que no carga responden con 0;
  - la descripción del verbo no cambia, porque sale en la tabla de comandos de `SKILL.md`.

  Cambian `internal/app/testdata/plazos/describe.json`, la parte de `plazos calcular` de `schemas/plazo.json` y los e2e
  que los comparan (FR-017). Comprobable: US4.

**`LeerTipos`**

- **FR-013**: `LeerTipos` MUST dejar de exigir los siete códigos y los dos cómputos, admitir un cómputo sin tipos y
  exigir las reglas de §C.2. Sus defectos siguen la forma de hoy: todos a la vez, cada uno con el tipo o el cómputo, y,
  al cargar el binario, código 1 (`inesperado`). Comprobable: US5, un subtest por regla.

**El cómputo civil**

- **FR-014**: `data/plazos.yaml` MUST llevar el cómputo `civil` de §C.1: un municipio, la sede; las unidades
  `dias-habiles` y `meses`; los periodos `inhabil` de agosto y del 24 de diciembre al 6 de enero, con el art. 130.2 LEC;
  y los preceptos 130.2, 133.1 (`fija: expiracion`), 133.2, 133.3 y 133.4 LEC. Ningún tipo de `data/plazos.yaml` cuenta
  con él, y el civil no entra en lo que la skill lleva a leer en las evals, que sigue saliendo de los tipos.
  Comprobable: US2.5, US5.1 y §C.9.
- **FR-015**: Cada precepto del civil MUST llevar el `vigente_desde` que dan sus redacciones grabadas (ADR 0042): el
  2022-12-23 el 130.2, el 2010-05-04 el 133.2 y el 133.4, y el 2001-01-08 el 133.1 y el 133.3.
  `TestPreceptosDeLosPlazos` MUST leer también los cómputos sin tipos, y su tabla de palabras clave gana en `a130`
  «24 de diciembre», «6 de enero» y «agosto», y en `a133` «día siguiente», «veinticuatro horas», «se excluirán los
  inhábiles», «de fecha a fecha» y «prorrogados». Comprobable: US6.
- **FR-016**: Una tarea `[datos]` MUST grabar los bloques `a130` y `a133` de `BOE-A-2000-323`, con las redacciones que
  `TestPreceptosDeLosPlazos` necesita para FR-015, en un manifiesto nuevo de H36 bajo `evidencias/h36/`, y añadir su
  raíz a `UnionDeGrabaciones`. El BOE consolidado ya tiene su fila revisada en `docs/SOURCES.md`. Comprobable: US6.1.

**Los tests, los contratos y la documentación**

- **FR-017**: Los tests de H9 que cambian MUST ser solo estos, y en ninguno cambian las entradas ni lo esperado de un
  caso de cálculo; cambian las constantes que desaparecen, los fixtures con la forma vieja de `data/plazos.yaml`
  (`plazosDelContrato`, definida en `tipos_test.go` y leída en `calcular_test.go`) y las pruebas de las reglas que se quitan:
  - `internal/core/plazos`: `calcular_test.go`, `propiedades_test.go`, `tipos_test.go`, `salida_test.go` y
    `avisos_test.go`;
  - `internal/app/plazos_test.go`: los enumerados en las etiquetas;
  - `internal/skills/referencias_test.go`: las constantes;
  - `internal/evals`: los tests de `CodigosDeAviso` y de `ComprobarCodigosDeAvisoDelPlazo`, y en
    `grabaciones_test.go` solo lo que pide FR-015 de `TestPreceptosDeLosPlazos`;
  - por el esquema: `internal/app/testdata/plazos/describe.json`, la parte de `plazos calcular` de `schemas/plazo.json`
    e `internal/app/testdata/script/h39-describe.txtar`, con su copia `specs/031-h39-c-omo-se/aceptacion/describe.txtar`.

  Tocar `internal/app/testdata/` y las copias congeladas de `specs/` lo hace una tarea `[datos]` que las declara.
  Comprobable: el diff de la rama contra `main` no toca otro test de H9 ni otra entrada o fecha esperada de un caso.
- **FR-018**: Los contratos de H9 que cambian MUST quedar reescritos en este spec (§C) y citados desde él; los ficheros
  de `specs/023-h9-internal-core-plazos/` no se tocan. Comprobable: el diff de la rama no toca ese directorio.
- **FR-019**: `CHANGELOG.md` MUST llevar en «Cambiado», bajo *Unreleased*: que los tipos y los cómputos de plazos son
  datos; que el esquema de `plazos calcular` ya no enumera los tipos ni los avisos; y que el cómputo civil de la LEC
  está en el motor, pero no se pide todavía desde `kitlegal plazos`. Comprobable: lectura del diff.
- **FR-020**: Las referencias generadas desde `data/plazos.yaml` MUST salir iguales, y ninguna skill, su protocolo, sus
  referencias ni sus evals cambian. Comprobable: SC-007.

**Los controles**

- **FR-021**: La tarea `[aceptacion]` MUST capturar con el binario de `main` (`cd3ad148`) los goldens que faltan: el
  sobre entero, sin `fecha_consulta`, de los casos 11, 15 y 16, y el de un error de cada fila de §C.6 que ese binario
  puede dar, con las órdenes de §C.6, incluidos los dos de municipios que no cuentan. Se comparan también, sin cambiar
  ninguna entrada ni lo esperado: `caso-a.json`, `caso-b.json`, `caso-e.json`, los guiones
  `internal/app/testdata/script/h9-plazos-*.txtar` con sus copias de `specs/023-h9-internal-core-plazos/aceptacion/`,
  `TestCasosDePlazos`, `TestCalcular` y `TestComprobarPeticion`. Comprobable: SC-001.
- **FR-022**: Un fichero de `internal/core/plazos/testdata/` MUST describir el cómputo de ensayo de §C.10, con la forma de
  §C.1 y sin tocar Go, y sus casos E1 a E13 MUST dar lo de su tabla. Comprobable: SC-003.
- **FR-023**: Los goldens C1 a C10 de §C.9 MUST calcularse con el cálculo del dominio (`plazos.Calcular`), los festivos de
  `data/festivos/`, el cómputo `civil` del `data/plazos.yaml` real y dos tipos de prueba de
  `internal/core/plazos/testdata/`, sin duración, que cuentan con él: uno en `dias-habiles`, con el 133.2 como precepto,
  y otro en `meses`, con el 133.3. Comprobable: SC-002.
- **FR-024**: `TestPropiedadesDelComputo` MUST seguir pasando sin cambiar su rejilla, y MUST ganar las propiedades del
  cómputo civil sobre la misma rejilla con un municipio cada vez y los dos tipos de prueba de FR-023: la fecha nunca es
  inhábil para la sede ni cae en un periodo `inhabil`; en días hábiles, entre el día siguiente al inicio y la fecha hay
  exactamente N días que no son inhábiles de base ni de un periodo; en meses, la fecha es la de fecha a fecha o, si esa
  es inhábil, el primer hábil siguiente. Comprobable: SC-004.
- **FR-025**: El cierre MUST correr las evals como siempre, midiendo todas las skills sin reutilizar nada de `main`,
  porque el hito toca `internal/app` e `internal/evals`; si una clase cae, la reparación no toca para que pase ningún
  texto de las skills. Comprobable: SC-008.

### Key Entities

- **Cómputo**: la regla con que se cuentan los plazos de una o varias normas. Lleva su código, sus municipios (uno o
  dos textos), sus unidades, sus periodos y sus preceptos (§C.1).
- **Periodo**: un tramo del año, de un día y mes a otro, que puede cruzar el año; lleva su precepto y su efecto, y, si
  es `aviso`, su aviso (código, etiqueta y texto).
- **Tipo**: un plazo con nombre que se pide a `calcular`. Lleva su código, su nombre, su cómputo, lo que da (`da`), su
  duración o sus `unidades`, y sus preceptos.
- **Aviso del motor**: los tres de cobertura y de una parte, fijos en el motor; los de periodo son datos.

### Quién consume cada salida, cuánto y hasta cuándo

- **El `data` de `plazos calcular`** lo pide `plazos-y-territorio` en su paso 9, una vez por plazo que pide la
  pregunta (casi siempre una; dos con la reposición y el contencioso), y lo usa entero: la fecha en la línea `⚠ PLAZO:`,
  cada precepto de `data.preceptos` para leerlo en el paso 10, y cada aviso copiado tal cual. No cambia con los datos de
  hoy (FR-011): de 0,77 a 0,86 KB en un plazo de un mes, cada festivo unos 90 bytes, cada aviso unos 250, cada tramo de
  cobertura unos 100 (research M1 y M2 de H9); el peor caso que estima `contracts/applet-plazos.md` §8 de H9, menos de
  15 KB por llamada, es una estimación y no un umbral, y este hito no lo cambia porque el `data` es el mismo. `plazos` no lee la caché ni el grafo: con cientos de normas y miles de bloques consultados, la llamada
  siguiente da lo mismo que la primera. Un aviso de periodo añade una línea, unos 250 bytes con el texto del art. 183;
  sale una vez por plazo y solo si el tramo pasa por el periodo, y deja de darse en la llamada cuyo tramo no lo toca.
- **El error de FR-007** lo recibe la skill en `data.mensaje`, una vez por llamada con una unidad que el cómputo no
  admite, unos 120 bytes; con los datos de hoy no se da, y deja de darse en cuanto la llamada pide una unidad que el
  cómputo admite.
- **El esquema de `plazos_calcular`** lo recibe el agente una vez por conexión con el servidor MCP. Su tamaño lo fija
  el golden `describe.json` (FR-012): los cinco textos de ayuda de §C.7 suman 603 bytes, y los de hoy 567 (contados con
  `wc -c` sobre los textos de §C.7 y de `plazos.go`, líneas 133 a 137, en esta sesión), y los dos `enum` de la salida pasan a dos `pattern` más cortos. Ya
  no crece con cada tipo ni cada aviso que entre: es el mismo con cualquier `data/plazos.yaml`.
- **El cómputo civil** no lo consume nadie desde el binario hasta H9.1 (US2.5): solo los tests de FR-023 y FR-024.
- **`CodigosDeAviso` y `EtiquetaDeAviso`** los consume `internal/evals` al juzgar una respuesta de las evals, una vez
  por respuesta; dan hoy cuatro códigos y crecen con un código por aviso de periodo de los datos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Con el binario de la rama y los datos de hoy, el 100 % de los casos de fecha de `TestCasosDePlazos` (del 1
  al 18 y el 21) dan la misma fecha y los mismos códigos de aviso que con el de `main`, y el 100 % de los goldens
  capturados con el de `main` (FR-021) dan el mismo sobre salvo `fecha_consulta`. Lo hacen cumplir, en `make ci`, el
  test que compara cada golden con la salida de la rama y `TestCasosDePlazos`, que salen en rojo ante la primera
  diferencia; los mutantes de US1.4, US1.5 y US1.6 lo prueban.
- **SC-002**: Los 10 casos C1 a C10 dan lo de su tabla (fecha, festivos saltados, códigos de aviso o código de salida
  y mensaje). Lo hace cumplir el test de FR-023 en `make ci`, que sale en rojo con una fila distinta; los cinco mutantes
  de US2.4 lo prueban.
- **SC-003**: Los 13 casos E1 a E13 dan lo de su tabla, y el test de FR-010 encuentra 0 literales o etiquetas con un
  código de `data/plazos.yaml` o del fichero de ensayo. Lo hacen cumplir, en `make ci`, el test del cómputo de ensayo y
  el de FR-010, que salen en rojo con una fila distinta o con una aparición; los mutantes de US3.2 a US3.6 lo prueban.
- **SC-004**: `TestPropiedadesDelComputo` pasa en el 100 % de su rejilla, la de H9 sin cambios y la del civil de
  FR-024. Sale en rojo en `make ci` con el primer caso que incumple una propiedad; el mutante que cuenta como hábiles
  los días de un periodo `inhabil` lo prueba.
- **SC-005**: `--describe` de `calcular` sale igual, byte a byte, en los 3 juegos de datos de US4.1, y `LeerTipos`
  detecta cada una de las reglas de §C.2 en su copia rota (una por regla). Lo hacen cumplir, en `make ci`, el test de
  `describe.json` con los tres juegos (`internal/app`), `TestEsquemasPublicados` (`schema-check`) y los subtests de
  `LeerTipos`, que salen en rojo si una salida difiere o si una copia rota no da su defecto; quitar cada regla, y volver
  a las etiquetas de hoy, lo prueban.
- **SC-006**: `TestPreceptosDeLosPlazos` pasa con los 5 preceptos del civil: cada bloque grabado, con sus palabras
  clave y con el `vigente_desde` de sus redacciones grabadas. Sale en rojo en `make ci` con un defecto; los mutantes de
  US6.2 lo prueban.
- **SC-007**: `make skills-sync` deja 0 diferencias en `skills/`. Lo hace cumplir `make ci`: la comprobación de las
  regiones generadas de `internal/skills` (`regiones_test.go`), que sale en rojo si una región no es la que se genera
  desde `data/`.
- **SC-008**: En el cierre, ninguna clase de las evals de ninguna skill queda por debajo de su umbral (ADR 0038 y
  0045); no se espera ningún cambio frente a `main`, porque ningún texto de skill cambia. Lo hace cumplir el job de
  evals del cierre (`.github/workflows/evals.yml`), que sale en rojo si una clase incumple su umbral.
- **SC-009**: `make ci` en verde con los únicos tests de H9 cambiados que nombra FR-017.

## Criterios del hito (literales)

**Aceptación** (del fichero del hito; US1.1, US1.2, US2.3, US3.1, US3.6, US4.1; SC-001, SC-002, SC-003, SC-005):

- Con el binario de la rama y los datos de hoy, los casos de fecha de `TestCasosDePlazos` (del 1 al 18 y el 21) dan la
  misma fecha y los mismos avisos que con el de `main`. Los goldens capturados con el de `main` dan el mismo sobre,
  salvo `fecha_consulta`.
- Los casos C1 a C10 dan lo de su tabla.
- El cómputo de ensayo calcula sus casos, y el Go de la rama no nombra ninguno de sus códigos ni de los de
  `data/plazos.yaml`.
- `--describe` de `calcular` sale igual con los datos de hoy, con un tipo más y sin datos.

**Controles** (del fichero del hito). Cada uno, con la comprobación que lo pone en rojo; esa comprobación se hace con
un mutante o contra `main` y se registra en `mutantes.md`:

- **El `data` de hoy, byte a byte** (FR-011, FR-021; US1; SC-001). La tarea `[aceptacion]` captura con el binario de
  `main` (`cd3ad148`) los goldens que faltan: el sobre entero, sin `fecha_consulta`, de los casos 11, 15 y 16, y el de
  un error de cada fila de §C.6, incluidos los dos de municipios que no cuentan. Se comparan también los goldens
  `caso-a.json`, `caso-b.json` y `caso-e.json`, los guiones `internal/app/testdata/script/h9-plazos-*.txtar` con sus
  copias de `specs/023-h9-internal-core-plazos/aceptacion/`, `TestCasosDePlazos`, `TestCalcular` y
  `TestComprobarPeticion`. Ninguna entrada ni lo esperado de esos casos cambia. En rojo: una palabra cambiada en el
  texto del aviso del art. 183 rompe el golden del caso 15; el silencio con `vencimiento` en lugar de `dia-siguiente`
  rompe el caso 11; el texto de los municipios de `contencioso` cambiado rompe el golden de su error.
- **El esquema de `calcular`, sin datos** (FR-012; US4; SC-005). `describe.json` y `schemas/plazo.json` salen iguales
  con una copia de `data/plazos.yaml` con un tipo y un aviso de periodo más, y con un `data/plazos.yaml` que no carga,
  como el caso «sin-los-tipos». En rojo: volver a las etiquetas de hoy.
- **Ningún código de los datos en Go** (FR-010; US3.6; SC-003). Un test recorre con `go/ast` el Go que no es de test de
  `internal/core/plazos`, `internal/app`, `internal/skills` e `internal/evals`. Falla si un literal de cadena es el
  código de un tipo, de un cómputo o de un aviso de periodo, o si una etiqueta de struct lo lleva como palabra entera.
  Los códigos salen de `data/plazos.yaml` y del fichero del cómputo de ensayo. Los tres avisos de cobertura del motor
  quedan fuera, porque son del motor, y también la descripción del verbo, que es de H9.1. En rojo: volver a escribir
  `if tipo.Codigo == "silencio"`, y volver a la ayuda de hoy.
- **Las propiedades de H9** (FR-024; SC-004). `TestPropiedadesDelComputo` sigue pasando sin cambiar su rejilla. Se
  añaden las del cómputo civil, sobre la misma rejilla con un municipio: la fecha nunca es inhábil para la sede ni cae
  en un periodo `inhabil`; en días hábiles, entre el día siguiente al inicio y la fecha hay exactamente N días que no
  son inhábiles de base ni de un periodo; en meses, la fecha es la de fecha a fecha o, si esa es inhábil, el primer
  hábil siguiente. En rojo: un mutante que cuenta como hábiles los días de un periodo `inhabil`.
- **Los goldens del cómputo civil** (FR-023; US2; SC-002). Se calculan con `plazos.Calcular`, los festivos de
  `data/festivos/` y el cómputo `civil` del `data/plazos.yaml` real, junto con unos tipos de prueba de
  `internal/core/plazos/testdata/`, sin duración, que cuentan con él: uno con `dias-habiles`, que lleva como precepto el
  133.2; otro con `meses`, que lleva el 133.3. La tabla es la de §C.9. En rojo: con agosto como `aviso` en lugar de
  `inhabil`, C1 da el 7 de agosto; con agosto como `no-corre`, C3 da el 15 de septiembre; sin el periodo del 24 de
  diciembre, C2 da el lunes 28 de diciembre; con agosto desde el día 2, C9 da el 1 de agosto; sin los festivos locales
  de la sede, C10 da el 9 de noviembre.
- **Un cómputo de ensayo que entra sin tocar Go** (FR-022; US3; SC-003). Un fichero en `internal/core/plazos/testdata/`
  describe un cómputo inventado, que no es de ninguna materia, con un municipio, las unidades `dias-habiles` y
  `dias-naturales`, un periodo `inhabil` que no es agosto ni la Navidad y un periodo `aviso` con su propio código, su
  etiqueta y su texto. El fichero tiene dos tipos: uno con duración fija y `dia-siguiente`, y otro sin duración. Los
  casos, calculados a mano en §C.10 con la rejilla de las propiedades, prueban cada eje, las fronteras de los dos
  periodos, su primer día y su último día laborables, y que el aviso sale con su forma `⚠ <ETIQUETA>:`. En rojo, un
  mutante por eje: un motor que solo reconoce agosto como periodo; uno que ignora `dia-siguiente` fuera del silencio;
  uno que no comprueba las unidades del cómputo; y uno que da los avisos de periodo con un texto fijo.
- **`LeerTipos`** (FR-013; US5; SC-005). Subtests con copias rotas de `data/plazos.yaml`. Cada uno espera el defecto que
  nombra el tipo o el cómputo, con una copia por cada regla de la entrega (§C.2). En rojo: quitar cada regla.
- **Los preceptos del cómputo civil contra lo grabado** (FR-015, FR-016; US6; SC-006). `TestPreceptosDeLosPlazos`: cada
  bloque está grabado; lleva sus palabras clave; cada `vigente_desde` es el de las redacciones grabadas. En rojo: el
  `vigente_desde` del 133.1 con el 2010-05-04, que es el del bloque y no el del apartado; y el del 130.2 con el
  2015-10-07, un día en que el apartado decía otra cosa.
- **Las referencias** (FR-020; SC-007): `make skills-sync` no deja diferencias.
- **Las evals** (FR-025; SC-008): el cierre las corre como siempre. El hito toca `internal/app` e `internal/evals`, lo
  común de las evals, así que mide todas las skills sin reutilizar nada de `main`. Ningún texto de skill cambia, así que
  no se espera ningún cambio frente a `main` (ADR 0038). Si una clase cae, la reparación no toca para que pase ningún
  texto de las skills.

## C · Los contratos que cambian

Esta sección reescribe, para H36, lo que cambia del contrato de H9 (`specs/023-h9-internal-core-plazos/`): FR 007, FR
008 y FR 010 de su spec, `contracts/applet-plazos.md` §1, §2, §5, §6 y §7, `contracts/datos-de-plazos.md` §2 y research
D3, D6, D9 y D17, en lo que dicen de que el código decide; y el punto 3 del ADR 0042, en que `LeerTipos` ya no exige
«los dos cómputos». Lo demás de esos documentos sigue en vigor. Con los datos de hoy, los tipos, sus municipios, sus
duraciones, sus preceptos, sus avisos y sus errores son los que dicen FR 007, FR 008 y FR 010 de H9, ahora descritos en
`data/plazos.yaml` y no decididos por su código.

### C.1 `data/plazos.yaml` (sustituye a `contracts/datos-de-plazos.md` §2 y a research D3 y D17 de H9)

Dos listas, `tipos` y `computos`, en este orden. Las claves y sus valores:

**Un tipo** (`tipos[]`):

| Clave | Obligatoria | Valor |
|---|---|---|
| `codigo` | sí | un código (§C.8), único entre los tipos |
| `nombre` | sí | el que lo describe |
| `computo` | sí | el `codigo` de un cómputo de `computos` |
| `da` | sí | `vencimiento` o `dia-siguiente` (FR-005) |
| `duracion` | no | `{cantidad, unidad}`: un entero de 1 o más y una unidad de su cómputo; sin ella, el tipo recibe la suya con una unidad de su cómputo |
| `unidades` | no, y solo sin `duracion` | de cada unidad de su cómputo, exactamente esas, al precepto que da su cómputo |
| `preceptos` | sí, uno o más, salvo con `unidades` | `{norma, bloque, apartado, fija, vigente_desde}`, como hoy |

**Un cómputo** (`computos[]`):

| Clave | Obligatoria | Valor |
|---|---|---|
| `codigo` | sí | un código (§C.8), único entre los cómputos |
| `municipios` | sí | una lista de uno o dos textos, cada uno el que nombra a un municipio que cuenta, en el orden en que se dan |
| `unidades` | sí | una lista no vacía y sin repetidos de las tres unidades del motor: `meses`, `dias-habiles`, `dias-naturales` |
| `periodos` | no | una lista de periodos, en el orden en que dan sus avisos |
| `preceptos` | sí, uno o más | como hoy |

**Un periodo** (`periodos[]`): `desde` y `hasta`, cada uno `MM-DD` de un día que tienen todos los años (no el `02-29`);
si `hasta` es anterior a `desde`, el periodo cruza el año. `efecto`: `inhabil`, `no-corre` o `aviso`. `precepto`:
`{norma, apartado}`, sin `bloque`, que nombra el precepto con esa norma y ese apartado que ya está entre los
`preceptos` del cómputo o, si el cómputo tiene tipos, entre los `preceptos` de cada uno de ellos; su bloque, su `fija` y
su `vigente_desde` son los de ese precepto. `aviso`, solo y siempre con `efecto: aviso`: `{codigo, etiqueta, texto}`,
con un código (§C.8), una etiqueta y un texto no vacíos y de una línea, sin `:` en la etiqueta.

Lo que el fichero de la rama dice de los dos cómputos de hoy y del civil (los tipos de hoy no cambian, salvo que cada
uno lleva `da`: `dia-siguiente` el `silencio` y `vencimiento` los otros seis):

```yaml
computos:
  - codigo: lpac
    municipios: [el del interesado, el de la sede del órgano]
    unidades: [meses, dias-habiles, dias-naturales]
    preceptos: [los cinco del art. 30 LPAC de hoy]
  - codigo: contencioso
    municipios: [el de la sede del juzgado]
    unidades: [meses]
    periodos:
      - {desde: "08-01", hasta: "08-31", efecto: no-corre,
         precepto: {norma: BOE-A-1998-16718, apartado: "128.2"}}
      - {desde: "12-24", hasta: "01-06", efecto: aviso,
         precepto: {norma: BOE-A-1985-12666, apartado: "183"},
         aviso: {codigo: dias-del-art-183-lopj, etiqueta: DÍAS DEL ART. 183 LOPJ, texto: "<el de §C.5>"}}
    preceptos: [los cinco de la LOPJ y del Código Civil de hoy]
  - codigo: civil
    municipios: [el de la sede del tribunal]
    unidades: [meses, dias-habiles]
    periodos:
      - {desde: "08-01", hasta: "08-31", efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}
      - {desde: "12-24", hasta: "01-06", efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}
    preceptos:
      - {norma: BOE-A-2000-323, bloque: a130, apartado: "130.2", fija: inhabiles, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.1", fija: expiracion, vigente_desde: "2001-01-08"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.2", fija: habiles, vigente_desde: "2010-05-04"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.3", fija: meses, vigente_desde: "2001-01-08"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.4", fija: prorroga, vigente_desde: "2010-05-04"}
```

En el fichero, cada tipo lleva sus claves de hoy en el orden de hoy, y `da` delante de `preceptos` o de `unidades`:
así los fragmentos de texto del fichero que usan los tests de H9 y de H26.1 siguen casando sin cambiar, y una sola
vez, porque el precepto de un periodo no lleva `bloque` (`bloque: a128,` solo está en el tipo `contencioso`).

El `fija` de los preceptos del civil sigue el vocabulario del fichero de hoy (`inhabiles`, `habiles`, `meses`,
`prorroga`), con `expiracion` del hito para el 133.1. El de `fija: duracion` sigue exigido en un precepto de cada tipo
con `duracion`, como hoy.

### C.2 Las reglas de `LeerTipos` (sustituye a la última frase de `contracts/datos-de-plazos.md` §2 de H9 y al punto 3 del ADR 0042)

`LeerTipos` deja de exigir los siete códigos y los dos cómputos, y admite un cómputo sin tipos. Exige, cada regla con
su subtest (FR-013):

1. que cada tipo nombre un cómputo que existe;
2. que la unidad de un tipo con duración sea una que admite su cómputo;
3. que las `unidades` de un tipo, si las lleva, sean las de su cómputo, ni una más ni una menos, cada una con su
   precepto, y que no lleve también `duracion`;
4. que la norma y el apartado del precepto de cada periodo sean los de un precepto del cómputo o, si el cómputo tiene
   tipos, los de un precepto de cada uno de ellos;
5. que un periodo `no-corre` sea un mes entero de 30 o 31 días, del día 1 al último, en un cómputo que admite solo
   `meses`;
6. que dos periodos de un cómputo no compartan ningún día;
7. que el código y la etiqueta de cada aviso de periodo sean únicos en el fichero y distintos de los de los tres avisos
   del motor;
8. que cada código de tipo, de cómputo y de aviso de periodo siga la gramática de §C.8, porque la salida los lleva y el
   esquema los describe con ella;
9. que cada cómputo lleve uno o dos municipios, una o más unidades del motor sin repetir, y periodos con `desde`,
   `hasta`, `efecto`, `precepto` y, en los de `aviso` y solo en ellos, su `aviso` completo (§C.1);
10. que cada tipo lleve `da` con uno de sus dos valores.

Siguen las reglas de hoy que no dependen de un código: cada tipo una vez, con nombre, con algún precepto y, con
`duracion`, una de 1 o más y un precepto que la fija; cada cómputo una vez y con algún precepto; y cada precepto con
sus cinco datos y un `vigente_desde` `AAAA-MM-DD`. Con cualquier defecto no hay tipos: el error los dice todos, cada uno
con el tipo o el cómputo que lo tiene, y el binario sale con 1 (`inesperado`) al ejecutar `calcular`.

### C.3 El cómputo (sustituye a `contracts/applet-plazos.md` §2 y a research D9 de H9)

Inhábiles de base: sábado, domingo, festivo nacional o autonómico (sin `parte`) de la comunidad de alguno de los
municipios que cuentan, o festivo local de alguno. Inhábil, además, un día de un periodo `inhabil` de su cómputo. Se
cuenta desde el día siguiente a `inicio`:

- días hábiles: la fecha es el enésimo día que no es inhábil;
- días naturales: el enésimo día y, si es inhábil, el primer día que no lo es;
- meses: el mismo día del mes de vencimiento, o el último de ese mes si no lo hay, y, si es inhábil, el primer día que
  no lo es.

Un periodo `no-corre` (solo en un cómputo de solo `meses`) sigue la regla de research D9 de H9 para agosto: si `inicio`
o el día siguiente cae en él, el cómputo parte de su último día; si entre el día siguiente al punto de partida y el
vencimiento de fecha a fecha cae algún día del periodo, el vencimiento pasa al mismo día del mes siguiente (o al último
de ese mes). Un periodo `aviso` no cambia el cómputo: sus días se cuentan y su aviso sale con FR-008. Con `da:
dia-siguiente`, la fecha es el día siguiente al vencimiento así calculado, sin prórroga.

En `contencioso`, con los datos de §C.1: el agosto `no-corre` y el 24 de diciembre al 6 de enero `aviso` dan lo que
dan hoy; lo que H9 decidió de los días del art. 183 LOPJ (research S4 de H9) se mantiene. **Festivos saltados**: como
hoy, en días hábiles los festivos del tramo que no se contaron, también los de un periodo `inhabil`; en días naturales y
en meses, los de la prórroga; nunca un sábado ni un domingo.

### C.4 Los municipios (sustituye a FR 007 y al primer punto de FR 010 de H9, y a research D4 en lo que decide por código)

Un tipo pide tantos municipios como textos lleva `municipios` de su cómputo: con los datos de hoy, dos en `lpac` (el del
interesado y el de la sede del órgano, que pueden ser el mismo) y uno en `contencioso` (el de la sede del juzgado).
Un día inhábil en cualquiera de ellos es inhábil para el cómputo.

### C.5 Avisos (sustituye a `contracts/applet-plazos.md` §5 y a research D6 de H9)

Los avisos del motor no cambian: `faltan-festivos-nacionales-y-autonomicos`, `faltan-festivos-locales` y
`festivo-de-una-parte`, con sus etiquetas y sus textos de H9. Los de periodo son datos: su código, su etiqueta y su
texto los da su periodo, y el texto de `data.avisos[]` es `⚠ <etiqueta>: <texto>`. Con los datos de hoy, el único es
el del cómputo `contencioso`, con la etiqueta `DÍAS DEL ART. 183 LOPJ` y este texto, byte a byte el de hoy:

```text
el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ declara inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, porque el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más temprana posible.
```

Orden en `data.avisos`: el de los nacionales y autonómicos que faltan; los de los locales que faltan; los de periodo,
en el orden de FR-008; los de un festivo de una parte. Con los datos de hoy, el de H9.

### C.6 Errores (sustituye a `contracts/applet-plazos.md` §6 y a research D10 de H9 en lo que cambia)

En este orden; el mensaje nombra la entrada. La columna «Orden capturada» es la de FR-021, con el binario de `main`:

| Defecto | Código | Clase | Orden capturada |
|---|---|---|---|
| el tipo no es uno de los de `data/plazos.yaml` | 2 | `argumentos` | `inventado 2026-09-09 Leganés Leganés` |
| `inicio` no es `AAAA-MM-DD` o no es un día del calendario | 2 | `argumentos` | `reposicion 2026-02-30 Leganés Leganés` |
| `--duracion` falta en un tipo sin `duracion`, no es un entero positivo, o se da a un tipo que la fija | 2 | `argumentos` | `generico 2026-09-09 Leganés Leganés --unidad dias-habiles` |
| `--unidad` falta con `--duracion`, no es del vocabulario del motor, o se da a un tipo que fija su duración | 2 | `argumentos` | `generico 2026-09-09 Leganés Leganés --duracion 3 --unidad semanas` |
| **nuevo**: `--unidad` es del vocabulario y el cómputo del tipo no la admite | 2 | `argumentos` | — (con los datos de hoy no se da; C7, E12) |
| no son tantos municipios como los del cómputo del tipo | 2 | `argumentos` | `reposicion 2026-09-09 Leganés` y `contencioso 2026-06-15 Madrid Leganés` |
| un municipio que `territorio resolver` daría con 2 | 2 | `argumentos` | `reposicion 2026-09-09 Arroyomolinos Leganés` |
| todo lo demás vale y algún municipio no está en la relación | 3 | `no-encontrado` | `reposicion 2026-09-09 Inexistente Leganés` |
| el plazo acabaría después del 9999-12-31 | 2 | `argumentos` | `generico 9999-12-31 Leganés Leganés --duracion 1 --unidad dias-naturales` |
| `data/festivos/`, `data/plazos.yaml` o `data/territorio/` del binario no cargan | 1 | `inesperado` | — (el binario de `main` carga sus datos empotrados; lo cubre el caso «sin-los-tipos») |

Cada orden lleva `kitlegal plazos calcular` delante y `--json` detrás. Los mensajes nuevos o compuestos:

- **Municipios que no cuentan** (FR-006), con un municipio: `el tipo "<tipo>" se cuenta con un municipio, <texto>, y se
  han dado <n>`; con dos: `el tipo "<tipo>" se cuenta con dos municipios, <texto 1> y <texto 2>, que pueden ser el
  mismo, y se han dado <n>`. Con los datos de hoy, los dos de `main`.
- **Unidad que el cómputo no admite** (FR-007): `el tipo "<tipo>" no admite la unidad "<unidad>": su cómputo,
  "<cómputo>", admite <unidades>`, con las unidades del cómputo en el orden del motor (`meses`, `dias-habiles`,
  `dias-naturales`) unidas como hoy en las alternativas («a o b», «a, b o c»).

Nunca 4, 5 ni 6.

### C.7 La ayuda de los argumentos (sustituye a `contracts/applet-plazos.md` §1 en la columna «Qué es»)

La forma de la orden, la herramienta `plazos_calcular` y sus argumentos no cambian. La ayuda, byte a byte:

| Argumento | Ayuda |
|---|---|
| `tipo` | `Tipo de plazo, por su código: los tipos y el cómputo con que se cuenta cada uno están en las referencias del paquete.` |
| `inicio` | `Fecha de inicio, AAAA-MM-DD: qué fecha es en cada tipo lo dicen las referencias del paquete.` |
| `municipios` | `Municipios que cuentan, por su nombre o su código INE: cuántos y cuáles pide el cómputo de cada tipo lo dicen las referencias del paquete.` |
| `duracion` | `Duración, un entero positivo; solo, y siempre, en los tipos que la reciben, que dicen las referencias del paquete.` |
| `unidad` | `Unidad de la duración: meses, dias-habiles o dias-naturales, una que admita el cómputo del tipo; solo, y siempre, con la duración.` |

La tabla de tipos de `contracts/applet-plazos.md` §1 de H9 deja de ser un contrato del applet: los tipos son los de
`data/plazos.yaml`, y la skill los lee en la región `recursos_y_plazos` de su referencia, que no cambia.

### C.8 La gramática de un código y el esquema (sustituye a `contracts/applet-plazos.md` §7 de H9 en `tipo` y `avisos[].codigo`)

Un código de tipo, de cómputo o de aviso de periodo es una o más palabras de minúsculas ASCII y cifras unidas por
guiones: `^[a-z0-9]+(-[a-z0-9]+)*$`, la gramática que el repositorio usa ya para otros códigos (`schemas/eval.yaml.json`).
`tipo` y `avisos[].codigo` de la salida de `--describe` llevan ese `pattern` y ningún `enum`. Los enumerados de
`festivos[].ambito`, `cobertura[].ambito` y `cobertura[].falta` siguen en las etiquetas y un test sigue exigiendo que
digan lo mismo que sus vocabularios del dominio. `internal/app/testdata/plazos/describe.json` y la parte de `plazos
calcular` de `schemas/plazo.json` son la forma canónica, y salen iguales con cualquier `data/plazos.yaml` (US4.1).

### C.9 Los goldens del cómputo civil

Con el cálculo del dominio, el cómputo `civil` del `data/plazos.yaml` de la rama, los festivos de `data/festivos/` en
`main` (`cd3ad148`) —los nacionales de `BOE-A-2024-21316`, `BOE-A-2025-21667` y `BOE-A-2026-21254`, y los locales de
Madrid de 2025 y 2026, sin locales de 2027— y dos tipos de prueba de `internal/core/plazos/testdata/` que cuentan con
él, sin duración y con `da: vencimiento`: `prueba-civil-dias`, con el 133.2 como precepto, para los casos en días, y
`prueba-civil-meses`, con el 133.3, para los de meses. La sede es Madrid (`28079`). Esos preceptos rigen desde antes
del 2022-12-22, así que C6 solo nombra el 130.2.

| # | Inicio | Duración | Fecha | Festivos saltados | Avisos | Razonamiento |
|---|---|---|---|---|---|---|
| C1 | 2026-07-24 | 10 días hábiles | 2026-09-07 | — | — | Desde el sábado 25: del lunes 27 al viernes 31, del 1 al 5; agosto, inhábil (130.2); del martes 1 al viernes 4 de septiembre, del 6 al 9; 5 y 6, fin de semana; lunes 7, el 10. Por la LPAC, con los mismos municipios, sería el viernes 7 de agosto |
| C2 | 2026-12-18 | 5 días hábiles | 2027-01-08 | 2026-12-25, 2027-01-01 y 2027-01-06, nacionales | `faltan-festivos-locales` | Lunes 21, martes 22 y miércoles 23, del 1 al 3; del 24 al 6, inhábiles (130.2); jueves 7 y viernes 8, el 4 y el 5. Los festivos del periodo que no son sábado ni domingo se dan como saltados, con la regla de hoy (§C.3). No hay locales de Madrid de 2027 |
| C3 | 2026-06-15 | 2 meses | 2026-09-01 | — | — | De fecha a fecha, el sábado 15 de agosto (133.3); agosto es inhábil y se prorroga al primer hábil (133.4), el martes 1 de septiembre. El 15 es sábado y no se da como festivo. El contencioso daría el 15 de septiembre (caso 3 de H9) |
| C4 | 2026-10-27 | 2 meses | 2027-01-07 | 2027-01-01 y 2027-01-06, nacionales | `faltan-festivos-locales` | De fecha a fecha, el domingo 27 de diciembre; del 28 al 6, inhábiles; jueves 7. El contencioso daría el lunes 28 de diciembre con el aviso del art. 183 (caso 15 de H9) |
| C5 | 2026-07-10 | 2 meses | 2026-09-10 | — | — | Agosto no para los meses en la LEC: de fecha a fecha, el jueves 10 de septiembre, hábil |
| C6 | 2022-12-22 | 2 meses | — | — | — | Código 3: el art. 130.2 rige así desde el 2022-12-23 (ADR 0042). El mensaje nombra `BOE-A-2000-323`, bloque `a130`, apartado 130.2, y ningún otro precepto |
| C7 | 2026-07-24 | 10 días naturales | — | — | — | Código 2: el cómputo civil no admite días naturales (133.2); mensaje de FR-007: `el tipo "prueba-civil-dias" no admite la unidad "dias-naturales": su cómputo, "civil", admite meses o dias-habiles` |
| C8 | 2026-07-24, con dos municipios (Madrid y Leganés) | 10 días hábiles | — | — | — | Código 2: el cómputo civil cuenta con un municipio, la sede, con el texto de sus datos: `el tipo "prueba-civil-dias" se cuenta con un municipio, el de la sede del tribunal, y se han dado 2` |
| C9 | 2025-07-30 | 2 días hábiles | 2025-09-01 | 2025-08-15, nacional | — | Jueves 31 de julio, el 1; agosto, inhábil, y el viernes 1 de agosto es el primer día del periodo; 30 y 31 de agosto, fin de semana; lunes 1 de septiembre, el 2 |
| C10 | 2026-11-02 | 5 días hábiles | 2026-11-10 | 2026-11-09, local de Madrid (`BOCM-20251212-34`) | — | Del martes 3 al viernes 6, del 1 al 4; 7 y 8, fin de semana; lunes 9, fiesta local de la sede; martes 10, el 5 |

Los días de la semana son del calendario, y los festivos, de `data/festivos/` en `main` (`cd3ad148`).

### C.10 El cómputo de ensayo

Un fichero de `internal/core/plazos/testdata/` con la forma de §C.1 y estos códigos y periodos; sus preceptos son
inventados, de la norma `BOE-A-0000-1`, con `vigente_desde` `2000-01-01`:

- cómputo `ensayo`: `municipios: [el del lugar del ensayo]`; `unidades: [dias-habiles, dias-naturales]`; periodos, del
  `04-06` al `04-10` `inhabil`, y del `10-19` al `10-23` `aviso`, con el aviso `{codigo: periodo-de-ensayo, etiqueta:
  PERIODO DE ENSAYO, texto: "el plazo pasa por el periodo de ensayo, del 19 al 23 de octubre; sus días se han
  contado."}`. En 2026, el 6 y el 10 de abril son lunes y viernes, y el 19 y el 23 de octubre, lunes y viernes: el
  primer y el último día de cada periodo son laborables en la rejilla;
- tipo `ensayo-fijo`: duración fija de 5 días hábiles, `da: dia-siguiente`;
- tipo `ensayo-libre`: sin duración, `da: vencimiento`.

Los casos van sobre la rejilla de las propiedades (Contexto), con el municipio `81001` «Lugar Uno» salvo donde se dice.
En la rejilla, abril y octubre de 2026 no tienen ningún festivo de `81001`. «Aviso del periodo» es
`⚠ PERIODO DE ENSAYO: el plazo pasa por el periodo de ensayo, del 19 al 23 de octubre; sus días se han contado.`

| # | Eje | Petición | Fecha | Avisos | Razonamiento |
|---|---|---|---|---|---|
| E1 | `inhabil`, días hábiles | `ensayo-libre`, 2026-04-02, 3 días hábiles | 2026-04-14 | — | Viernes 3, el 1; 4 y 5, fin de semana; del 6 al 10, inhábiles; 11 y 12, fin de semana; lunes 13 y martes 14, el 2 y el 3. Sin el periodo, el martes 7 |
| E2 | primer día del periodo | `ensayo-libre`, 2026-04-03, 1 día hábil | 2026-04-13 | — | 4 y 5, fin de semana; el lunes 6, primer día, inhábil; hasta el 10; lunes 13. Con el periodo desde el 7, el 6 |
| E3 | último día del periodo | `ensayo-libre`, 2026-04-09, 1 día hábil | 2026-04-13 | — | El viernes 10, último día, inhábil; 11 y 12, fin de semana; lunes 13. Con el periodo hasta el 9, el 10 |
| E4 | `inhabil`, días naturales | `ensayo-libre`, 2026-03-27, 10 días naturales | 2026-04-13 | — | Diez días desde el sábado 28: el lunes 6 de abril, inhábil del periodo; prórroga hasta el lunes 13. Sin el periodo, el 6 |
| E5 | `dia-siguiente` | `ensayo-fijo`, 2026-06-01 | 2026-06-09 | — | Cinco días hábiles: del martes 2 al viernes 5, del 1 al 4; lunes 8, el 5, el vencimiento; da el día siguiente, el martes 9. Ignorando `dia-siguiente`, el 8 |
| E6 | `dia-siguiente` sin prórroga | `ensayo-fijo`, 2026-06-05 | 2026-06-13 | — | Del lunes 8 al viernes 12, el vencimiento; el día siguiente es el sábado 13, que no se prorroga (§C.3) |
| E7 | `aviso`, primer día | `ensayo-libre`, 2026-10-16, 1 día hábil | 2026-10-19 | aviso del periodo | 17 y 18, fin de semana; el lunes 19, primer día del periodo, se cuenta: el tramo del 17 al 19 pasa por él |
| E8 | `aviso`, último día | `ensayo-libre`, 2026-10-22, 1 día hábil | 2026-10-23 | aviso del periodo | El viernes 23, último día del periodo, se cuenta y el tramo pasa por él. Con el periodo hasta el 22, sin aviso |
| E9 | `aviso`, después | `ensayo-libre`, 2026-10-23, 1 día hábil | 2026-10-26 | — | El tramo, del 24 al 26, empieza tras el último día del periodo |
| E10 | `aviso`, antes | `ensayo-libre`, 2026-10-15, 1 día hábil | 2026-10-16 | — | El tramo, el viernes 16, acaba antes del primer día del periodo |
| E11 | orden de los avisos | `ensayo-libre`, 2026-06-23, 120 días naturales, municipio `82001` «Lugar Tres» | 2026-10-21 | `faltan-festivos-locales`, `periodo-de-ensayo`, `festivo-de-una-parte` | Ciento veinte días desde el miércoles 24 de junio: el miércoles 21 de octubre, hábil en la `82`. La `82` no tiene fichero local; el tramo pasa del 19 al 21 de octubre por el periodo; el 24 de junio, miércoles, es fiesta solo en la Isla de la Rejilla y se contó como hábil. Orden de FR-008 |
| E12 | unidades del cómputo | `ensayo-libre`, 2026-06-01, 1 mes | — | — | Código 2: `el tipo "ensayo-libre" no admite la unidad "meses": su cómputo, "ensayo", admite dias-habiles o dias-naturales` |
| E13 | municipios del cómputo | `ensayo-libre`, 2026-06-01, 1 día hábil, municipios `81001` y `81002` | — | — | Código 2: `el tipo "ensayo-libre" se cuenta con un municipio, el del lugar del ensayo, y se han dado 2` |

Ningún caso de ensayo salta festivos: en E1 a E10, E12 y E13 no hay festivo de `81001` en el tramo, y en E11 los días
naturales solo dan los de la prórroga, que no hay.

## Assumptions

- La gramática de §C.8 es la de `schemas/eval.yaml.json`; que `LeerTipos` la exija (§C.2, regla 8) mantiene lo que hoy
  garantiza la regla de los siete códigos: que el `tipo` y los códigos de aviso de la salida cumplen el esquema que la
  describe (FR 085 de H9). Todos los códigos de hoy, del civil y del ensayo la cumplen.
- Los textos del municipio del civil (`el de la sede del tribunal`) y del ensayo los fija este spec: el hito dice «un
  municipio, la sede», y el art. 130.2 LEC habla de «la respectiva Comunidad Autónoma o localidad» del tribunal.
- `da` es obligatoria en cada tipo (§C.2, regla 10): es un eje de todo tipo (ADR 0061, punto 3), y sin valor por omisión
  ningún tipo da el día siguiente por descuido.
- La fila «no cargan» de §C.6 no se captura con el binario de `main`, que lleva sus datos empotrados y los carga; su
  forma (código 1, `inesperado`) la sigue comprobando el caso «sin-los-tipos», cuyo mensaje es el defecto de
  `LeerTipos`.
- Las referencias del paquete a las que remite la ayuda de §C.7 son las de `skills/plazos-y-territorio/references/`, que
  ya dicen el tipo, su duración y su cómputo (`administrativo.md`), cuántos municipios cuentan
  (`calculo_de_plazos.md`) y la fecha de inicio del silencio (`administrativo.md`); el hito no las cambia.
- La tarea `[datos]` de FR-016 graba solo del BOE consolidado, que tiene su fila revisada en `docs/SOURCES.md` de
  `main` (ADR 0018), y sin modelo.

## Fuera de alcance

Lista cerrada del hito, y lo que no está en él:

- **Cómputos y unidades que no entran**: el cómputo civil sustantivo del art. 5.2 CC; la unidad `anios`; los plazos de
  fecha fija, como el calendario del contribuyente (ADR 0061, punto 5).
- **Lo que no es el cómputo del plazo**: las actuaciones urgentes y la habilitación (arts. 131 y 133.2, párrafo
  segundo, LEC); el plazo que corre desde el fin de otro (art. 133.1, párrafo segundo); la interrupción por fuerza
  mayor (art. 134); la presentación hasta las quince horas del día hábil siguiente (art. 135.5); la hora del
  vencimiento como dato de la salida.
- **Lo que es de H9.1**: un tipo en `data/plazos.yaml` que cuente con `civil`; el cómputo social y la modalidad procesal
  de la LRJS como dato de entrada; cualquier bandera o argumento nuevo de `kitlegal plazos calcular`; la descripción del
  verbo `calcular`.
- **Lo que es de H42**: los tipos de la revisión tributaria.
- **Lo que no cambia**: una clave nueva o un valor distinto en el `data` de la salida, en los avisos o en los mensajes,
  con los datos de hoy; el cómputo `contencioso`, en el que lo que H9 decidió de los días del art. 183 LOPJ (research S4
  de H9) se mantiene, aunque el art. 183 grabado, en la redacción de la LO 14/2022, haga también inhábil agosto para
  las actuaciones judiciales; `data/festivos/`, `data/cobertura.yaml` y `data/normas.yaml`; cualquier skill, su
  protocolo, sus referencias o sus evals; la forma de `vigente_desde` y la regla del ADR 0042.
- **Lo que el ADR 0061 rechaza**: un lenguaje de reglas general (cualquier unidad, cualquier día inicial, prórroga
  opcional), componer la ayuda o los enumerados del esquema desde los datos, y un `no-corre` que no sea un mes entero en
  un cómputo de solo `meses`, o en un plazo por días.
- **Ningún ADR nuevo** ni cambio de los ADR 0061 y 0042, que van en la propuesta que trae la entrada; ningún fichero de
  `specs/023-h9-internal-core-plazos/`.
- **Un esquema JSON para `data/plazos.yaml`** además de su lector (rechazado en research D17 de H9).
- **Un aviso de periodo nuevo en `schemas/eval.yaml.json`**: con los datos de hoy no hay ninguno; lo añade a mano el
  paquete que lo traiga (ADR 0061, punto 7).
