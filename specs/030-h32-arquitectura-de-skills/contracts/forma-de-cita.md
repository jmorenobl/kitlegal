# Contrato · Una sola forma de cita y las formas compartidas (FR-010, FR-011, FR-013, FR-014)

## 1. El control de la forma única (FR-010)

Subtest `forma-de-cita` de `TestSkillsDelRepositorio` (`internal/app/skills_test.go`), dentro de `make skills-check`.

- **Dónde mira**: cada `skills/*/SKILL.md` y cada `skills/*/references/*.md`, escritos a mano o generados, línea a
  línea, también dentro de los bloques de código y de las tablas.
- **Qué es una cita de un bloque**: toda aparición de `BOE-A-<año>-<número>` o de `<identificador>` seguida de
  `, bloque ` (expresión `(BOE-A-[0-9]{4}-[0-9]{1,9}|<identificador>), bloque `).
- **Cuándo está en la forma única**: si el carácter anterior es `[` y desde él casa
  `\[(BOE-A-[0-9]{4}-[0-9]{1,9}|<identificador>), bloque (<id>|[A-Za-z0-9][A-Za-z0-9.-]{0,63})\]`: dentro de los
  corchetes, el identificador, una coma, un espacio, la palabra `bloque`, un espacio y el id —el real o el marcador
  `<id>`—, y nada más.
- **El defecto**: `<skill>: <fichero>:<línea>: cita de un bloque fuera de la forma única: «<línea>»`, uno por línea.
- **Sus mutantes**, en el mismo subtest, sobre copias en `t.TempDir()`: una referencia a mano con
  `(BOE-A-2015-10565, bloque daprimera)` y otra con `[art. 46.1 de la LJCA, BOE-A-1998-16718, bloque a46]` dan cada una
  su defecto, con su fichero y su línea; `[<identificador>, bloque <bloque>]` también (el marcador del id es `<id>`).

Lo que no es una cita de un bloque no lo mira: `(BOE-A-2003-23186)` —la norma que rige, sin `, bloque `—, un
identificador en su columna de `normas.md`, «el bloque daprimera de la LPAC».

En `c991808` lo ponen en rojo `legal-core/SKILL.md` (las dos líneas `⚠ PLAZO:` y la de `⚠ PLAZO SIN CALCULAR:`), las
nueve filas de `cobertura_por_materia.md` y las de `recursos_y_plazos.md` y las líneas 8 y 16 de
`boe-legislacion/references/como_se_cita.md`; la entrega las cambia ([textos.md](textos.md)).

## 2. El control de las formas compartidas (FR-013)

Subtest `formas-compartidas` de `TestSkillsDelRepositorio`.

- **Las líneas que mira**: las de un bloque `text` (entre ```` ```text ```` y ```` ``` ````) de cada
  `skills/*/SKILL.md` que, sin los blancos del principio, empiezan por `⚠ <ETIQUETA>:`, con la etiqueta en mayúsculas
  (letras, espacios y cifras) hasta los dos puntos.
- **La regla**: para cada etiqueta, el conjunto de sus líneas es el mismo en cada `SKILL.md` que la lleva.
- **El defecto**: `<skill A> y <skill B>: la forma ⚠ <ETIQUETA>: difiere: «<línea de A>» frente a «<línea de B>»`.
- **Su mutante**: una palabra cambiada en la línea `⚠ SIN CONSULTA AL BOE:` de `normativa` en una copia da el defecto
  con las dos skills y la etiqueta.

En la entrega comparten etiqueta `normativa` y `plazos-y-territorio`: `SIN CONSULTA AL BOE`, `REDACCIÓN A FECHA` y
`REDACCIÓN DE HOY`, una línea cada una, iguales (comprobado sobre los borradores en esta sesión).

## 3. La línea `⚠ NO CUBIERTO:` (FR-011)

Las dos formas del cuerpo de `plazos-y-territorio`, con sus marcadores (la primera con la norma que lo rige, como hoy,
por su nombre y su identificador entre paréntesis, que no es una cita de un bloque):

```text
⚠ NO CUBIERTO: <procedimiento> se rige por <norma> (<identificador de la norma>), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (<precepto> [<identificador>, bloque <id>]).
⚠ NO CUBIERTO: <procedimiento> lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (<precepto> [<identificador>, bloque <id>]).
```

La línea de cada fila, entera, la da la referencia del paquete ([paquetes.md](paquetes.md) §3).

## 4. La medida de `cobertura_sin_declarar` (FR-014)

`juzgarElPlazo` (`internal/evals/plazos.go`) cuenta como declarada la respuesta de una eval de cobertura solo si **una
misma línea** `⚠ NO CUBIERTO:` (la forma que ya reconoce `formaDeNoCubierto`) lleva:

1. el identificador `no_cubierto.norma` de la eval, como hoy; y
2. la cita, en la forma única, del precepto de su fila: la pareja `no_cubierto.precepto.norma` y
   `no_cubierto.precepto.bloque`, extraída de esa línea con `ExtraerCitas` (`internal/evals/citas.go`).

`ExtraerNoCubierto` pasa a devolver, por cada línea, sus identificadores y sus citas, y `NoCubiertoDeclarado` exige las
dos cosas en la misma línea. El umbral no cambia: `cobertura_sin_declarar`, 0 en cada modo, decide
(`umbralDeCoberturaSinDeclarar`, `internal/evals/umbrales.go`).

`no_cubierto.precepto` lo añade el hito a las evals de cobertura (07, 08 y 12) y al esquema (`schemas/eval.yaml.json`,
`no_cubierto: {norma, precepto: {norma, bloque}}`, `precepto` obligatorio), y `TestEvalsDelRepositorio` comprueba que
cada `no_cubierto.precepto` es el `precepto` de una fila de `data/cobertura.yaml` (la de la eval 07, la de los tributos;
la 08, la del tráfico; la 12, la de la Seguridad Social: las tres, `BOE-A-2015-10565` y `daprimera`).

Tests, en `TestJuzgarPlazos` (`internal/evals/plazos_test.go`): una respuesta con la norma esperada y sin la cita del
precepto no cuenta como declarada; con la norma en una línea y la cita en otra, tampoco; con las dos en la misma
línea, sí. Las respuestas de ejemplo de hoy (`declaracionDeLaLGT`, `declaracionSinNorma`, `declaracionDeLaLeyDeTrafico`,
`respuestaDeLa12`) pasan a llevar la cita.
