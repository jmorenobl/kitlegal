# Contrato: las variantes de `normativa`

FR-001 a FR-008; SC-001 a SC-003. Research D1 a D6, D12; data-model §1 a §3.

## 1. Dónde viven

```text
laboratorio/normativa/
├── normativa-control/skills/normativa/
│   ├── SKILL.md
│   └── references/{buscar_la_norma,como_se_cita,con_herramientas,lecturas_que_fallan,normas,
│                   redaccion_en_una_fecha,redaccion_modificada,varios_bloques}.md
├── normativa-remisiones-en-orden/skills/normativa/…      # los mismos nueve ficheros
├── normativa-cuerpo-corto/skills/normativa/…             # los mismos nueve ficheros
└── normativa-testigo-en-referencia/skills/normativa/…    # los mismos nueve ficheros
```

Cada variante es una skill completa con el nombre `normativa` (el `name` de su frontmatter) y nada más en su
directorio: es lo que el arnés copia como directorio de skills de sus sesiones (contracts/arnes.md §3), así que cada
sesión tiene solo la variante de su fila (FR-008). Fuera de `skills/` (no se empotra), de `evals/` (el job no la lee) y
de lo que mide el job (ninguna ruta de `laboratorio/` elige una skill); nada de esto lo regenera `make skills-sync`. Los
36 ficheros los escriben las tareas: copias de `skills/normativa` y, en los tres declarados, el resultado de aplicarles
su declaración. El tope de 11 000 caracteres de cuerpo no se aplica a ellas (FR-007): sigue en `skills/`.

En todo lo que sigue, **el cuerpo** de un `SKILL.md` son las líneas que siguen a la que cierra su frontmatter (la 17 hoy),
y **control** son los nueve ficheros de `normativa-control`.

## 2. `normativa-control` (FR-002)

Los nueve ficheros de `skills/normativa`, con las mismas rutas relativas, ni uno de más ni de menos, y los mismos bytes.
Hoy: cuerpo de 10 929 caracteres y ocho remisiones, una a cada referencia (research V1).

## 3. `normativa-remisiones-en-orden` (FR-003)

Control con cinco sustituciones en `SKILL.md`; las otras ocho rutas, las de control. Cada «antes» está una sola vez en
el `SKILL.md` de control; cada «después» son los bytes que lo sustituyen; y el «después», sin la marca `- ` de la viñeta
y con cada salto de línea y la sangría de la línea que sigue contados como un espacio, es carácter a carácter el texto
del hito. Ninguna sustitución toca un byte fuera de su «antes». Con las cinco, el cuerpo tiene 11 084 caracteres (+155)
y las mismas ocho remisiones (research V2).

**1. El párrafo del «Protocolo»**

```text
antes:   Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la tienes, con su
         herramienta, tras leer [con_herramientas](references/con_herramientas.md).⏎
después: Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md). Sigue los pasos en orden,
         cada operación con `kitlegal … --json` desde el `PATH` o, si la tienes, con su herramienta.⏎
hito:    Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md). Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la tienes, con su herramienta.
```

**2. El párrafo del paso 1**

```text
antes:   Toma el `BOE-A-…` de la tabla de [normas](references/normas.md); si no está, búscala con
         [buscar_la_norma](references/buscar_la_norma.md).⏎
después: Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee
         [buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.⏎
hito:    Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee [buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.
```

**3. El principio de la viñeta del paso 2 que empieza por «Sin fecha —con ella»** (lo que sigue, desde « una vez, con»,
es el de control)

```text
antes:   - Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, lee cada bloque
después: - Con una fecha en la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) antes de leer el
           bloque. Sin fecha, lee cada bloque
hito:    Con una fecha en la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) antes de leer el bloque. Sin fecha, lee cada bloque
```

**4. La viñeta del paso 2 de varios bloques**

```text
antes:   - Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no termina con `0`, antes de
           seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).⏎
después: - Para leer varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md). Si una orden no
           termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de seguir.⏎
hito:    Para leer varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md). Si una orden no termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de seguir.
```

**5. En «Cómo se cita», lo que sigue a «también del remitido» hasta el punto de la remisión** (lo que sigue, desde
« Cada línea `⚠`», es el de control)

```text
antes:   ;⏎
         repásalas, y las dudas, en [como_se_cita](references/como_se_cita.md).
después: .⏎
         Antes de escribir la respuesta, lee [como_se_cita](references/como_se_cita.md) y repasa con ella las⏎
         citas.
hito:    . Antes de escribir la respuesta, lee [como_se_cita](references/como_se_cita.md) y repasa con ella las citas.
```

En los bloques, `⏎` es el salto de línea del fichero; el texto empieza tras las nueve columnas del rótulo, y la sangría
de dos espacios de las líneas de continuación de una viñeta (3 y 4) es del fichero. La remisión de «Redacción
modificada» no cambia.

## 4. `normativa-cuerpo-corto` (FR-004)

Control con otro `SKILL.md`: el mismo frontmatter, byte a byte, y el cuerpo que da esta regla sobre el cuerpo de
control; las otras ocho rutas, las de control.

- Un **encabezado** es una línea del cuerpo que empieza por `## ` o por `### `; su **sección**, él y las líneas que lo
  siguen hasta el siguiente encabezado. Las líneas anteriores al primero son el preámbulo.
- Se quedan, en el orden de control: el preámbulo; enteras, las secciones de `## Protocolo`, `### 1. Identificar la
  norma`, `### 2. Leer`, `## Cómo se cita`, `## Redacción en una fecha` y `## Redacción modificada`; y de la de
  `### 4. Responder`, su encabezado escrito `### 3. Responder`, sus líneas anteriores a la primera viñeta (las que
  empiezan por `- `), la viñeta que empieza por `- Cada afirmación sobre el contenido de una norma lleva su cita` —su
  línea y las que la siguen con dos espacios delante— y las líneas vacías que siguen a la primera viñeta.
- No se queda nada más: ni las secciones de `### 3. Completar`, `### 5. …`, `### 6. …`, `## Comandos` (con su tabla) y
  `## Reglas`, ni las otras cuatro viñetas del paso 4.
- El cuerpo termina en su última línea no vacía, con su salto.
- Cada encabezado que se queda o se renombra está una sola vez en el cuerpo de control, y la viñeta, una sola vez en su
  sección; si no, la diferencia ya no se aplica.

Hoy da un cuerpo de 3298 caracteres, con las ocho remisiones (research V3), que empieza y termina así:

```text
⏎
## Protocolo⏎
⏎
Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la tienes, con su⏎
…
- Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no termina con `0`, antes de⏎
  seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).⏎
⏎
### 3. Responder⏎
⏎
- Cada afirmación sobre el contenido de una norma lleva su cita, de lo que `kitlegal boe` devolvió en esta⏎
  conversación.⏎
⏎
## Cómo se cita⏎
…
⚠ REDACCIÓN MODIFICADA: <forma legible> [<identificador>, bloque <id>]: la redacción con fecha de vigencia AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.⏎
```⏎
```

## 5. `normativa-testigo-en-referencia` (FR-005)

Control con dos cambios; las otras siete rutas, las de control.

- **`SKILL.md`**: una sustitución, como las de §3, con
  `antes` = `[con_herramientas](references/con_herramientas.md).` (una sola vez en control) y
  `después` = `[con_herramientas](references/con_herramientas.md), que dice también cómo termina la respuesta.`: el
  cuerpo gana 44 caracteres (10 973) y ninguna línea pasa de 120 (research V3).
- **`references/con_herramientas.md`**: el de control seguido de un salto de línea, `## Cómo termina la respuesta`, otro
  salto de línea y el párrafo de la regla de `laboratorio/variantes/referencia-paso/skills/articulos-del-boe/references/cierre.md`
  —lo que sigue a sus dos primeras líneas, byte a byte: una línea que termina en `Sello de lectura: KL-7Q4`, con su
  salto—. Son 197 caracteres más (research V3, V4):

  ```text
  …resuelve con `boe_indice`, sea cual sea su `data.clase`: lo primero que puede fallar es el id.⏎
  ⏎
  ## Cómo termina la respuesta⏎
  ⏎
  La respuesta termina con una última línea, sola y sin ningún formato —ni negrita, ni cursiva, ni comillas, ni código—, que dice exactamente: Sello de lectura: KL-7Q4⏎
  ```

## 6. Uso, de fuera adentro

Las lee el agente de una sesión del sondeo: el `SKILL.md` de su variante, una vez, al activar `normativa`, y cada
referencia cuando la abre, que es lo que se mide; en la aceptación, 200 sesiones, una variante por sesión. Su tamaño es
fijo —el de §2 a §5: de 3298 a 11 084 caracteres de cuerpo y ocho referencias de 909 a 3510 bytes (la de
`testigo-en-referencia`, `con_herramientas.md` con 197 más)—, no crece con el uso del producto —cientos de normas y miles
de bloques consultados no las tocan— y solo cambia cuando cambia `skills/normativa` (FR-002). No son una señal. Ninguna
skill del producto ni el binario que se distribuye las leen.

## 7. Controles

- **`TestVariantesDeNormativa`** (`internal/evals/laboratorio_normativa_test.go`, en `make ci`; FR-002 a FR-006,
  SC-001, SC-002):
  - `normativa-control` frente a `skills/normativa`: el mismo conjunto de rutas y los mismos bytes; falla con la ruta del
    primer fichero que difiere, falta o sobra.
  - cada una de las otras tres frente a la que se calcula desde control con su declaración (§3 a §5): falla con la ruta
    del primer fichero que difiere; con la sustitución, el encabezado o la viñeta que ya no se aplica sobre control y
    cuántas veces está; o con el «después» cuyo texto normalizado no es el del hito.
  - en las tres, el cuerpo tiene una remisión `](references/<f>.md)` a cada una de las ocho referencias de control y
    ninguna más; y ningún nombre de variante de `normativa` es el de una variante de H39.
- **`TestLaboratorioFueraDelProducto`** (`internal/evals/laboratorio_test.go`, el de H39, generalizado; FR-001, SC-003;
  research D12): recorre todo `laboratorio/`; deja fuera de la comprobación de bytes los ficheros que son copia de uno
  de `skills/` —`…/skills/<s>/<r>` con los bytes de `skills/<s>/<r>`—; y suma los nombres de los directorios de
  `laboratorio/normativa/` a los que no pueden ser un tramo de una ruta de `kitlegal.Skills()` ni de lo que instala
  `kitlegal skills install --dir`, ni una skill que ejecute el job, ni una carpeta de `evals/`. Ninguna ruta de
  `laboratorio/` mide (`rutaQueMide`). Con los ficheros de H39, hace lo mismo que hoy.
- **`TestSkillsDelRepositorio`** (`internal/app/skills_test.go`, sin cambios): el tope de 11 000 caracteres sigue en
  `skills/` (FR-007).
