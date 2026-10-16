# Data model: H23.1 · el umbral `documento_sin_cotejar` y la skill que coteja antes de nada

Ningún tipo exportado, ningún fichero de datos y ningún esquema nuevos. Todo lo que este hito añade a `internal/evals`
es sin exportar y vive en dos ficheros que ya existen, `informe.go` y `umbrales.go`; el contrato del ADR 0029 (`Umbral`)
no cambia. Los nombres son los que fija este plan (research D1-D7).

## 1. Eval que pide el cotejo

Una eval con `activa: true` y algún comando esperado con `applet: cita` y `verbo: cotejar`, lleve o no `roj` (spec,
«Key Entities»; FR-001).

- **Dónde**: `pideElCotejo(eval Eval) bool`, en `informe.go`.
- **Hoy**: la 04, la 06, la 08 y la 09 de `jurisprudencia`; ninguna de `boe-legislacion` ni de `legal-core`
  (research M5). Lo fija `TestEvalsDelRepositorio/piden-el-cotejo`.
- **Ninguna lista en el código**: lo dicen los `comandos` de cada eval.

## 2. Cotejo

Una ejecución de `cita cotejar` en la sesión, con cualquier argumento, que consultó y terminó bien (FR-002):

- como orden, en su traza: código 0 y sin `--dry-run`, `--describe` ni la ayuda;
- como llamada a la herramienta `cita_cotejar`: un resultado que no viene marcado como error.

- **Dónde**: `cotejaUnDocumento(sesion Sesion) bool`, en `informe.go`: alguna de las invocaciones de
  `invocacionesDe(sesion)` satisface `comandoDelCotejo`, el comando esperado `ComandoEsperado{Applet: "cita",
  Verbo: "cotejar"}`, sin ROJ (research D2, D3).
- **Lo que no mira**: los argumentos, el documento que pasó, lo que devolvió y lo que la respuesta dijo después.

## 3. Lo que se mide de cada respuesta

`respuestaDelJuez` (`informe.go`), la respuesta del modelo que decide en un grupo, gana dos campos:

| Campo | Qué es | Quién lo llena |
|---|---|---|
| `pideElCotejo bool` | Su eval pide el cotejo (§1) | `medirLosCotejos`, con la eval de su sesión (`resultado.Eval`) |
| `cotejo bool` | Su sesión cotejó (§2) | `medirLosCotejos`, con la sesión leída (`sesionJuzgada.sesion`) |

Los dos quedan falsos si ninguna eval de la skill pide el cotejo, o si la respuesta es del grupo sin modo (las evals
sin binario ni servidor), que no entra en ningún umbral.

`respuestasMedidas` (`informe.go`) gana `conCotejo bool`: alguna eval de la skill pide el cotejo. Lo calcula
`medirLasRespuestas`, que con él llama a `medirLosCotejos(evals, sesiones)`; la condición con la que
`medirLasRespuestas` devuelve nil —sin juez y sin ninguna eval que declare `sentencias`— no cambia (research D4).

`grupoJuzgado` (`informe.go`) gana dos métodos:

| Método | Qué da |
|---|---|
| `queLePidenElCotejo() []respuestaDelJuez` | Las respuestas del grupo con `pideElCotejo`, en orden de sesión: el `total` |
| `sinCotejar() []respuestaDelJuez` | De ellas, las que no tienen `cotejo`, en orden de sesión: la `medida` y las que nombra el motivo |

## 4. El umbral `documento_sin_cotejar:<modelo>:<modo>`

Un elemento de `umbrales` con el contrato del ADR 0029, uno por cada modo del plan, si `conCotejo` (FR-001):

| Campo | Valor |
|---|---|
| `nombre` | `documento_sin_cotejar:` + el modelo que decide + `:` + el modo |
| `descripcion` | «Respuestas de `<modelo>` en el modo `<modo>` cuya sesión no ejecutó ningún cita cotejar que consultara y terminara bien, sobre sus respuestas medidas en las evals que activan la skill y piden cita cotejar» |
| `medida` | `len(sinCotejar())` del grupo del modo |
| `total` | `len(queLePidenElCotejo())` del grupo del modo; hoy, 12 por modo |
| `comparacion` | `"<="` |
| `umbral` | `0` (`umbralDeRespuestasSinCotejar`) |
| `cumple` | `medida/total <= 0`, 0 si `total` es 0 (`compararUmbral`) |
| `decide` | `true` |

- **Dónde**: `umbralDeDocumentoSinCotejar(modelo string, delModo grupoJuzgado) Umbral` y
  `nombreDelUmbralDeDocumentoSinCotejar(modelo string, modo Modo) string`, en `umbrales.go`; lo añade
  `respuestasMedidas.umbrales`, detrás de `cita_sin_documento` de su modo.
- **Invariantes**: `0 <= medida <= total`; cada sesión cuenta una vez como mucho; ninguna sesión sin medir, cortada,
  de un modelo informativo o de una eval sin binario ni servidor está en el `total` ni en la `medida` (son las
  respuestas de `sin_activar`, research V6); `total <= ` el `total` de `sin_activar` del mismo modo.
- **Orden de los quince** de `jurisprudencia`: contracts/umbral-documento-sin-cotejar.md §1.

## 5. El motivo

Si el umbral no se cumple, `motivosDeLosUmbrales` da «umbral `<nombre>`: `<medida>` de `<total>` (`<p>` %), y tiene que
ser ≤ 0,0 %», y detrás, tras «: », las sesiones de `sinCotejar()` separadas por «; », que nombra la función nueva
`sinCotejarDelUmbral(nombre string) string` de `respuestasMedidas` desde `nombradasEn` (FR-003; research D7). Uno que
se cumple no da motivo. Ejemplo y tamaño: contracts §3.

## 6. Lo que no cambia

- `ResultadoDeEval` y todo el juicio de una sesión: `comandos_ejecutados`, `comandos_ausentes`, sus motivos y la tasa y
  el veredicto de cada serie (FR-031).
- Los trece umbrales de antes, con su nombre, su descripción, su medida, su total, su umbral y su `decide` (FR-030).
- `boe-legislacion` y `legal-core`: sus evals no piden el cotejo, y sus `umbrales` siguen con trece, nueve que deciden,
  y con `[]` (FR-007).
- El sondeo: usa `respuestasQueSeJuzgan` y no `medirLasRespuestas` (research V10).
- `informe.json` no tiene esquema en `schemas/`; su forma la fijan los tests.

## 7. El estado de la skill

`skills/jurisprudencia/SKILL.md` no tiene estados. Su versión, v0.1.1, la registra `CHANGELOG.md` (FR-024, FR-040).
Lo que cambia y lo que no: contracts/skill-jurisprudencia.md.
