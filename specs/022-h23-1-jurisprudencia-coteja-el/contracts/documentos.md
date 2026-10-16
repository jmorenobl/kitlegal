# Contrato: lo que dicen los documentos (FR-040 a FR-045)

Cada pasaje que el hito deja falso o incompleto, qué tiene que decir después y qué no se toca. La redacción la
escribe el run; cada cifra, fichero, test y umbral que nombre sale del árbol (FR-045). Ningún test ata este texto al
código, y no se añade ninguno. Las líneas son las de `main` en `ef35b40` (research V22).

## `CHANGELOG.md`, *Unreleased* (FR-040, FR-065)

Dos entradas nuevas, dentro de `## [Unreleased]` (líneas 13 a 63) y en ninguna versión publicada:

- En **«Añadido»**: el umbral `documento_sin_cotejar:<modelo>:<modo>`. Qué cuenta: las sesiones de una eval que pide
  `cita cotejar` sin ningún `cita cotejar` que consultara y terminara bien, ni como orden ni como llamada a
  `cita_cotejar`, cada una una vez como mucho. Sobre qué: las respuestas medidas del modelo que decide en ese modo, las
  del `total` de `sin_activar`, de las evals que piden el cotejo; hoy, 12 por modo. Con 0, y decide; su motivo nombra
  cada sesión que cuenta. `umbrales` de `jurisprudencia` pasa de trece elementos, nueve que deciden, a quince, once que
  deciden; `boe-legislacion` sigue con trece y `legal-core` con `[]`.
- En **«Corregido»** (research D15): `jurisprudencia` v0.1.1, que sustituye a la v0.1. Con un documento en la
  conversación, coteja su ficha antes de decir nada de él, y si es la sentencia pedida lo dice `es_la_pedida`, nunca
  una comparación a ojo. Con las cifras del defecto: en la eval 06, 6 de las 48 respuestas de `claude-sonnet-5-5` de las
  ocho tandas medibles desde el 2026-10-10 no ejecutaron `cita cotejar` y dieron el documento por la sentencia pedida,
  4 en el modo orden y 2 en el modo herramienta (`docs/USO.md`, 2026-10-15). Qué no cambia: la `description`, la tabla
  de comandos, los pasos 2 y 4 a 7 y las reglas; y las líneas de `SKILL.md` que tenga al acabar.

## `CONTRIBUTING.md`, «Job de evals» (FR-041)

| Línea en `main` | Hoy | Después |
|---|---|---|
| 1144-1145 | «en `jurisprudencia`, otros trece, más abajo» | quince, más abajo. Lo de `boe-legislacion` («son trece, nueve de ellos decidiendo») y `legal-core` (`[]`) no cambia |
| 1172 | «**En `jurisprudencia` son también trece, y deciden nueve**» | quince, y deciden once |
| 1176-1185, la tabla | Ocho filas | Una fila más, detrás de la de `cita_sin_documento:<modelo>:<modo>`: `documento_sin_cotejar:<modelo>:<modo>`; `medida`, las de una eval que pide `cita cotejar` cuya sesión no ejecutó ningún `cita cotejar` que consultara y terminara bien, ni como orden ni como llamada a `cita_cotejar`; `total`, sus respuestas medidas en ese modo, en las evals que piden el cotejo (12); `umbral`, `0`; `decide`, sí |
| 1187-1190 | «Primero van los cuatro del modo orden y los cuatro del modo herramienta»; «cuatro de los trece» | los cinco y los cinco; cuatro de los quince. El resto del orden no cambia |
| Detrás de 1194-1201, el párrafo de `cita_sin_documento` | — | Un párrafo nuevo: `documento_sin_cotejar` es un hecho de la sesión y se mide sin modelo, con lo que el juicio ya extrae de cada sesión —las invocaciones de su traza y sus llamadas a las herramientas, con el criterio con el que da por ejecutado un comando esperado de `cita`, sin mirar los argumentos—; lo tiene, con los demás umbrales de sus respuestas, toda skill con alguna eval que pide el cotejo, sin que el código nombre ninguna (`TestUmbralesDeJurisprudencia`, `TestEvalsDelRepositorio`); y su motivo: `umbral documento_sin_cotejar:claude-sonnet-5-5:orden: 1 de 12 (8,3 %), y tiene que ser ≤ 0,0 %: <sesión>`, con `; <sesión>` por cada otra |

No cambian: la frase de 1141-1142 («existen solo si la skill tiene juez o si alguna de sus evals declara
`sentencias`»), que sigue siendo cierta (research D4); la tabla y el orden de `boe-legislacion`; el párrafo de
`cita_sin_documento`; y lo que «Formato común de eval» dice de él (`:824`).

## `docs/JURISPRUDENCIA.md` (FR-042)

| Línea en `main` | Hoy | Después |
|---|---|---|
| 227 | «y el job decide con `cita_sin_documento`, un hecho de la sesión» | decide con dos hechos de la sesión: `cita_sin_documento` y `documento_sin_cotejar`, que con un documento delante la sesión lo cotejó |

## `CLAUDE.md` (FR-043)

| Línea en `main` | Hoy | Después |
|---|---|---|
| 67 | «Lo que comprueba un guion —la cita, las líneas `⚠`, la activación, `cita_sin_documento`, la medida del juez— sigue con umbral 0» | la enumeración nombra también el cotejo del documento (`documento_sin_cotejar`) |

No se tocan la viñeta de H23.1 de «Estado del repositorio» ni la frase del siguiente hito: las escribe la persona al
fusionar (spec, «Fuera de alcance»).

## `docs/WORKFLOW.md` (FR-044)

No cambia. Lo único que dice de los umbrales que deciden está en «`medir_cierre`» (`:139`): «en `boe-legislacion`
deciden nueve», con su lista, que sigue siendo cierta. Ninguna frase dice cuántos umbrales lleva `jurisprudencia`
(research V22).

## Lo que no se toca (FR-045, FR-052)

`docs/ROADMAP.md`, ningún ADR, `README.md`, `docs/USO.md`, `docs/HISTORIAL.md`, la web, `.github/` —tampoco el
comentario de `.github/workflows/evals.yml:157`, que queda para la persona (research V23)—, y el spec, el plan o los
informes de un hito anterior.
