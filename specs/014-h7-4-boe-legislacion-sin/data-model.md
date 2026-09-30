# Data model: H7.4

Lo que el hito añade o cambia en `internal/evals`, en los ficheros de `evals/` y en la definición del job. Nada del
binario cambia. Los nombres de Go son los que fija el plan; los contratos dan las formas exactas.

## 1. La lista de expresiones prohibidas (`ExpresionesProhibidas`)

`evals/<skill>/expresiones-prohibidas.yaml`, validada contra `schemas/expresiones-prohibidas.yaml.json`
([contracts/lista-de-expresiones.md](./contracts/lista-de-expresiones.md)).

| Campo (clave YAML) | Tipo | Clase | Cambio |
|---|---|---|---|
| `Maquinaria` (`maquinaria`) | `[]string` | A | sin cambios |
| `OtraConversacion` (`otra_conversacion`) | `[]string` | A | sin cambios |
| `Anuncio` (`anuncio`) | `[]string` | A | gana las 25 formas de la clase A (research D6) |
| `RedaccionNoLeida` (`redaccion_no_leida`) | `[]string` | **B** | nuevo, obligatorio en el esquema |
| `FormasFijas` (`formas_fijas`) | `[]string` | — | nuevo, obligatorio en el esquema: textos con los marcadores `<cita>` (opcional) y `<fecha>` que se quitan antes de buscar |

Reglas: cada expresión, una o más palabras sin blancos de más ni `*`/`_` (el patrón de hoy); cada forma fija, texto
de una línea con al menos un carácter fuera de los marcadores. La clase B es exactamente `RedaccionNoLeida`: la
alimenta el umbral `redaccion_no_leida:<modelo>`. Una lista que no cumple el esquema es un fichero mal formado (H7.2
FR 055). El valor cero sigue siendo el de una skill sin lista (`legal-core`).

`ExtraerExpresionesProhibidas(texto, lista)`: quita del texto cada forma fija de la lista (la línea
`⚠ REDACCIÓN MODIFICADA:`, con su cita o sin ella) y después la marca, la etiqueta y los dos puntos de cada aviso de
vigencia (`boe.EtiquetasDeAviso`), cambiando cada tramo por un salto de línea; y devuelve las expresiones que lleva, en el orden maquinaria, otra conversación, anuncio y redacción no leída,
sin repetir. `esDeLaClaseB(expresion)` (sin exportar) dice si una expresión de la lista es de `RedaccionNoLeida`.

## 2. La eval (`Eval`)

`evals/<skill>/<nn>-<descripción>.yaml`, validada contra `schemas/eval.yaml.json`
([contracts/evals-y-juicio.md](./contracts/evals-y-juicio.md)).

| Campo (clave YAML) | Tipo | Dónde | Juicio |
|---|---|---|---|
| `NoSeActivan` (`no_se_activan`) | `[]string`, nombres de skill, ≥ 1 | cualquier eval | cada una que la sesión activa es un motivo y la sesión no pasa |
| `RedaccionesModificadas` (`redacciones_modificadas`) | `[]RedaccionEsperada`, ≥ 1 | solo con `activa: true` | cada una que la respuesta no lleva en una línea `⚠ REDACCIÓN MODIFICADA:` es un motivo y la sesión no pasa |

`RedaccionEsperada`: `Norma` (`norma`, `BOE-A-…`), `Bloque` (`bloque`), `FechaVigencia` (`fecha_vigencia`, `AAAAMMDD`,
la superada) y `FechaVigenciaReciente` (`fecha_vigencia_reciente`, `AAAAMMDD`, la leída). Su texto, en el informe y en
los motivos: `<norma> <bloque> <fecha_vigencia> <fecha_vigencia_reciente>`.

`ExtraerRedaccionesModificadas(texto) []RedaccionEsperada`: por cada línea del texto con la forma de la etiqueta de
`version-obsoleta`, la primera cita de la línea y sus dos primeras fechas de ocho cifras, en su orden; una línea sin
cita o con menos de dos fechas no da ninguna.

## 3. La sesión leída (`Sesion`)

| Campo | Antes | Ahora |
|---|---|---|
| `Respuesta` | `result` del **último** mensaje `result`, si es `success` sin `is_error` | `result` del **primer** mensaje `result`, si es `success` sin `is_error` (la respuesta a la pregunta, research D12) |
| `Fin`, `Terminada`, `MotivoSinTerminar`, `ErrorDelResultado`, `SkillsActivadas`, `Reintentos` | — | sin cambios (el último mensaje; todas las activaciones del transcript) |

## 4. El juicio de una sesión (`ResultadoDeEval`)

Claves nuevas de cada elemento de `evals` en `informe.json`:

| Campo | JSON | Qué |
|---|---|---|
| `RedaccionesEncontradas`, `RedaccionesAusentes` | `redacciones_modificadas_encontradas`, `redacciones_modificadas_ausentes` | las esperadas, en su orden, repartidas; vacías si la eval no las espera |

`Motivos` gana, en este orden detrás de la activación: «se activó la skill <nombre>, que la eval dice que no se activa»
por cada una de `no_se_activan` activada; y detrás de los hallazgos, «redacción modificada ausente: <texto>» por cada
ausente. `Pasa` exige además que no se active ninguna de `no_se_activan` y que no falte ninguna redacción esperada. Las
expresiones prohibidas se juzgan como hoy (solo en evals que activan la skill), ahora con las formas fijas quitadas.

## 5. El recuento de las respuestas (interno) y lo publicado

`recuentoDeRespuestas`, por modelo del job (el que decide y después los informativos), sobre las sesiones juzgadas, no
ilegibles, no sin medir, **terminadas**, de series que pide el plan y de evals que activan la skill:

| Campo | Qué |
|---|---|
| `respuestas` | esas sesiones |
| `conAlguna` | las que llevan alguna expresión de la lista |
| `conRedaccionNoLeida` | las que llevan alguna de `redaccion_no_leida` |
| `sinActivar` | las que no activaron la skill |

`expresiones_prohibidas_por_modelo` publica, como hoy, `modelo`, `con_alguna` y `respuestas` (ahora sin las sesiones sin
terminar). Nil si la skill no tiene lista. El sondeo usa el mismo recuento para su línea de expresiones.

## 6. Los umbrales (`Umbral`, contrato del ADR 0029)

En `boe-legislacion` (skill con lista), cinco, en este orden ([contracts/informe-del-job.md](./contracts/informe-del-job.md)):

| `nombre` | `medida` | `total` | `comparacion` | `umbral` | `decide` |
|---|---|---|---|---|---|
| `expresiones_prohibidas:claude-sonnet-5-5` | `conAlguna` | `respuestas` | `<=` | 0.05 | `true` |
| `sin_activar:claude-sonnet-5-5` | `sinActivar` | `respuestas` | `<=` | 0 | `true` |
| `redaccion_no_leida:claude-sonnet-5-5` | `conRedaccionNoLeida` | `respuestas` | `<=` | 0 | `true` |
| `expresiones_prohibidas:claude-haiku-4-5-20251001` | `conAlguna` | `respuestas` | `<=` | 0.05 | `false` |
| `duracion_de_las_sesiones` | segundos | — | `<=` | 900 | `true` |

En `legal-core`, `[]`. Un umbral con `decide: true` que no se cumple es un motivo de la raíz y pone el veredicto en
`fallo` (el de la duración, con el prefijo de la ejecución, como hoy).

## 7. La decisión de la tanda (`decidirLaTanda`)

Entradas: el `databaseId` de la ejecución propia y, de cada ejecución del flujo sobre el commit, `ejecucionDelCommit`
(sin exportar, como todo lo de la tanda):

| Campo | Qué |
|---|---|
| `id` | `databaseId` |
| `terminada` | `status == "completed"` |
| `tanda` | `tandaSinDecidir` (su trabajo `tanda` no está o no está `completed`), `tandaQueMide` (`completed` con el paso «Esta ejecución mide el commit» en `success`) o `tandaQueNoMide` (cualquier otro caso) |

Salida, `decisionDeLaTanda`: `mide` (ninguna anterior sin terminar tiene `tandaQueMide`) y `pendientes` (las anteriores
sin terminar `tandaSinDecidir`, si ninguna mide). Solo cuentan las ejecuciones con `id` menor que la propia y sin
terminar. El bucle de `TestTandaDelCommit` consulta, decide y, con pendientes, espera 10 s y repite, como mucho 10 min;
agotado, mide. Escribe `medir=si` o `medir=no` en `-salida`.

## 8. El error de uso del sondeo (`errorDeUso`, sin exportar)

Envuelve lo que devuelve `comprobarElSondeo` por los argumentos o la credencial (el texto de hoy, un error por línea).
`TestSondeo` lo escribe en `uso.txt` del temporal y termina sin fallar; el guion lo imprime en la salida de error y sale
con 1. Los errores de leer la definición del job, construir el binario, instalar las skills o repartir las sesiones no
son de uso: `TestSondeo` falla y el guion imprime el registro, como hoy.

## 9. Ficheros de datos del hito

| Ruta | Qué | Tarea |
|---|---|---|
| `schemas/expresiones-prohibidas.yaml.json` | `redaccion_no_leida` y `formas_fijas`, obligatorias | `[datos]` |
| `schemas/eval.yaml.json` | `no_se_activan` y `redacciones_modificadas` | `[datos]` |
| `evals/boe-legislacion/expresiones-prohibidas.yaml` | las formas nuevas, `redaccion_no_leida` y `formas_fijas` | con el esquema |
| `evals/boe-legislacion/20-lcsp-dos-bloques-redaccion-cambiada.yaml` | la eval nueva | `[datos]` (con su grafo previo) |
| `evals/legal-core/0{1,2,3}-*.yaml` | `no_se_activan: [boe-legislacion]` | con el esquema |
| `testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/` | las dos derivadas, escritas por `TestGrabacionesDerivadas -actualizar-derivadas` | `[datos]` |
| `internal/evals/testdata/sesiones/leer-sesion/respuesta-antes-de-una-tarea-en-segundo-plano/` | transcript sintético, código y salida de error del caso de FR-098 | `[datos]` |
