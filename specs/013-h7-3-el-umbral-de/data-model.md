# Data model: H7.3

Entidades del spec («Key Entities») y lo que el plan les da en el repositorio. Ninguna vive en el binario: todas son de
`internal/evals` (el formato común de eval, el job y el sondeo), de sus guiones y de los datos de las evals. Decisiones
en [research.md](./research.md) (D1-D22); formatos exactos, con ejemplos y bytes, en [contracts/](./contracts/).

## 1. Umbral del informe (FR-001 a FR-008)

Tipo Go `evals.Umbral`, elemento de `Informe.Umbrales` (`umbrales` en `informe.json`, siempre presente).

| Campo Go | Clave JSON | Tipo | Regla |
|---|---|---|---|
| `Nombre` | `nombre` | `string` | `expresiones_prohibidas:<modelo>` o `duracion_de_las_sesiones`; único en el informe |
| `Descripcion` | `descripcion` | `string` | una línea en español |
| `Medida` | `medida` | `float64` | respuestas con alguna expresión, o segundos |
| `Total` | `total` | `*int`, `omitzero` | solo en los de expresiones: respuestas medidas del modelo en evals que activan la skill |
| `Comparacion` | `comparacion` | `string` | siempre `"<="` en este hito |
| `Umbral` | `umbral` | `float64` | `0.05` en los de expresiones; el objetivo en segundos en el de duración |
| `Cumple` | `cumple` | `bool` | `valor <= umbral` en `float64`, con `valor = medida/total` (0 si `total` es 0) o `medida` |
| `Decide` | `decide` | `bool` | verdadero en el del modelo que decide y en el de duración; falso en los de los informativos |

Cuándo existen (sin ningún caso por skill): uno de expresiones por modelo del job si la skill tiene lista; el de
duración si el job le da objetivo (> 0). `legal-core`: `[]`. `boe-legislacion`: tres (research D5).

## 2. Sesión leída (`Sesion`, ampliada; FR-033, FR-040)

Lo nuevo que `LeerSesion` toma del transcript `stream-json` (research D6, V2, V3):

| Campo | Qué es |
|---|---|
| `Reintentos []ReintentoDeLaAPI` | cada mensaje `system` con `subtype` `api_retry`, en su orden: `Intento` (`attempt`), `Maximo` (`max_retries`), `Error` (`error`: `rate_limit`, `overloaded`…); uno sin esos tres campos hace la sesión ilegible |
| `TerminaEnReintento bool` | el último mensaje del transcript es un `api_retry` |
| `ErrorDelResultado string` | el `result` del último mensaje del transcript si es un `result` con `is_error` verdadero, sea cual sea el código de la sesión; vacío si no |
| `MotivoSinTerminar` | el de antes y, si `ErrorDelResultado` no está vacío, con su texto: `result con is_error` (código 0) pasa a `result con is_error: <ErrorDelResultado>`, y `código <n>` (un código que no es del tope) pasa a `código <n>: result con is_error: <ErrorDelResultado>`, p. ej. `código 1: result con is_error: Failed to authenticate. API Error: 401 OAuth access token is invalid.` (research V18); los del tope no cambian |

`ReintentosPorLimiteDeRitmo()`: los de `Reintentos` con `Error` `rate_limit`.

## 3. Sesión sin medir por límite de uso (FR-040 a FR-044)

Clasificación sin modelo, en este orden (research D6):

| Clase | Condición | Motivo en la sesión |
|---|---|---|
| (a) límite de uso | último mensaje `result` con `is_error` cuyo texto empieza por `You've hit your`, `You've reached your` o `You're out of` | `sin medir por límite de uso: mensaje del límite de uso: <texto>` |
| (b) reintentos agotados | la sesión no terminó y su último reintento es `rate_limit` con `Intento ≥ Maximo` | `sin medir por límite de uso: reintentos por rate_limit agotados` |
| (c) cortada durante reintentos | `Cortada` (124 o 137) y `TerminaEnReintento` con el último `rate_limit` | `sin medir por límite de uso: cortada por el tope durante reintentos por rate_limit` |
| sin abrir | el repartidor no la abrió tras una sesión (a) | `sin abrir tras el límite de uso` (solo en la raíz del informe) |

Efectos: la sesión no pasa ni falla; su resultado lleva `sin_medir` y ese único motivo; no entra en la medida ni en el
total de las expresiones; su serie queda sin medir; el informe la lista en `sesiones_sin_medir`; el veredicto es
`fallo` con el motivo propio. Solo (a) detiene el repartidor.

Tipo `SesionSinMedir` (`sesiones_sin_medir` en la raíz): `sesion`, `eval`, `modelo`, `motivo`.

## 4. Resultado, serie e informe (campos nuevos)

| Tipo | Campo nuevo (clave JSON) | Regla |
|---|---|---|
| `ResultadoDeEval` | `ReintentosPorLimiteDeRitmo` (`reintentos_por_limite_de_ritmo`) | de su sesión; 0 si no se pudo leer |
| `ResultadoDeEval` | `SinMedir` (`sin_medir`) | vacío si se midió; si no, la clase de §3 |
| `TasaDelInforme` | `SinMedir` (`sin_medir`) | sesiones sin medir de la serie, abiertas o no; con alguna, `pasa` falso y la serie no da el motivo «pasan N de M» |
| `Informe` | `Umbrales` (`umbrales`) | §1 |
| `Informe` | `DuracionDeLasSesiones` (`duracion_de_las_sesiones`) | segundos enteros, redondeados hacia arriba (research D14) |
| `Informe` | `ReintentosPorLimiteDeRitmo` (`reintentos_por_limite_de_ritmo`) | suma de los de cada sesión |
| `Informe` | `SesionesSinMedir` (`sesiones_sin_medir`) | §3, en orden de sesión; `[]` si ninguna |
| `RecuentoDeExpresiones` | — | `respuestas` deja fuera las sesiones sin medir |
| `InformeAEscribir` | `DuracionDeLasSesiones`, `ObjetivoDeDuracion`, `SinAbrir []SesionPlanificada` | lo que el repartidor mide y no abrió; objetivo 0 = sin objetivo |

La serie que el plan pide cuenta sus sesiones sin abrir como sin medir: el motivo «hay N sesiones y el plan pide M»
queda para las que faltan por otra causa.

## 5. Lista de expresiones prohibidas (FR-020 a FR-024)

| Aspecto | Valor |
|---|---|
| Claves | `maquinaria`, `otra_conversacion` y **`anuncio`**, las tres obligatorias; ninguna otra |
| Tipo Go | `ExpresionesProhibidas{Maquinaria, OtraConversacion, Anuncio []string}` (`yaml:"anuncio"`) |
| Orden de búsqueda | maquinaria, otra conversación, anuncio |
| Contenido de `anuncio` en `boe-legislacion` | las 14 expresiones de research D2 |
| Esquema | `schemas/expresiones-prohibidas.yaml.json`, `anuncio` en `required` y `properties` |

## 6. Concurrencia de evals y definición del job (FR-030, FR-034, FR-035, FR-051)

`.github/workflows/evals.yml`, trabajo `evals`:

| Clave | Valor |
|---|---|
| `name` | `evals (${{ matrix.skill }})` |
| `strategy.matrix.skill` | `[boe-legislacion, legal-core]` |
| `strategy.matrix.include` | `{skill: boe-legislacion, concurrencia: 4, objetivo_de_duracion: 900}`, `{skill: legal-core, concurrencia: 1, objetivo_de_duracion: 0}` |
| `concurrency` | `group: evals-${{ github.event.pull_request.head.sha \|\| github.sha }}-${{ matrix.skill }}`, `cancel-in-progress: false` |
| `timeout-minutes` | `120` (≥ el peor caso de research D13 de cada skill) |
| `env.CONCURRENCIA_DE_EVALS` | `${{ matrix.concurrencia }}` |
| `env.OBJETIVO_DE_DURACION_DE_EVALS` | `${{ matrix.objetivo_de_duracion }}` |

Lector Go: `leerDefinicionDelJob(ruta) (DefinicionDelJob, error)`, con `Nombre`, `Grupo`, `CancelaLaEnCurso *bool`,
`ConcurrenciaDeFlujo bool`, `TopeEnMinutos int`, `Env map[string]string`, `PorSkill map[string]AjustesDeSkill{Concurrencia,
ObjetivoDeDuracion int}`, y los modelos, repeticiones del `env`. Lo usan `TestDefinicionDelJob` y el sondeo.

## 7. Ejecución de las sesiones (repartidor; FR-030 a FR-032, FR-037, FR-044, FR-050)

| Tipo | Campos |
|---|---|
| `SesionesAEjecutar` | `Plan []SesionPlanificada`, `Concurrencia int`, `Evals` (directorio de la skill), `Sesiones` (directorio de salida), `Skills` (directorio de skills instaladas que se enlaza), `Guion` (ruta de `scripts/evals-sesion.sh`), `Entorno []string` (base), `Traza bool`, `Tope` (240 s), `MargenDelTope` (10 s) |
| `EjecucionDeSesiones` | `Abiertas []string`, `SinAbrir []SesionPlanificada`, `Duracion time.Duration` |

Directorio de cada sesión: `trabajo/`, `cache/`, `traza/`, `claude/skills/<entrada>` (enlaces), `tmp/` (0700),
`eval.txt`, `modelo.txt`, `pregunta.txt` (de `PrepararSesion`), `sesion.jsonl`, `sesion.err`, `codigo-de-la-sesion`
(del repartidor). Variables que el repartidor pone sobre la base: `KITLEGAL_CACHE_DIR`, los cuatro de proxy a
`http://127.0.0.1:9`, `NO_PROXY`/`no_proxy` `api.anthropic.com`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`,
`CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0`, `CLAUDE_CONFIG_DIR`, `TMPDIR`, `CLAUDE_CODE_TMPDIR` y, con `Traza`,
`KITLEGAL_EVALS_TRAZA=si`.

## 8. Sondeo (FR-060 a FR-068)

| Argumento de `make` | Bandera de `TestSondeo` | Regla | Defecto |
|---|---|---|---|
| `SKILL` | `-skill` | forma de `name`, con carpeta de evals y en la matriz del job | ninguno |
| `EVALS` | `-evals` | `nn` de dos cifras separados por comas; cada uno, el prefijo de una eval bien formada de la skill | ninguno |
| `MODELO` | `-modelo` | no vacío, forma de id de modelo | ninguno |
| `REPETICIONES` | `-repeticiones` | entero ≥ 1 | ninguno |
| `CONCURRENCIA` | `-concurrencia` | vacío o entero ≥ 1 | la del job para la skill |
| — | `-temporal` | el directorio que crea `scripts/evals-sondeo.sh` | — |

Contenido del temporal: `bin/kitlegal`, `home/` (con las skills instaladas), `sesiones/<sesión>/` (§7), `go-test.log`,
`salida.txt`. Lo borra el guion al terminar.

Salida: texto, [contracts/sondeo.md](./contracts/sondeo.md) §4.
