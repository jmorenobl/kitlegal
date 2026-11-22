# Contrato: la API de `internal/core/plazos` y lo que toma `internal/evals`

Reescribe, en lo que cambia, `specs/023-h9-internal-core-plazos/contracts/applet-plazos.md` §2 y §5 (spec §C.3 a §C.6).
Lo que este contrato no nombra no cambia: `Peticion`, `MunicipioResuelto`, `Calcular`, `ComprobarPeticion`, `Plazo` y
sus tipos de salida, `Calendario`, `CargarCalendario`, los vocabularios de festivos y de cobertura, `Unidad`,
`Unidades()`, sus tres constantes, `FijaDuracion`, `Duracion` y `Precepto`.

## 1. Lo exportado que cambia

| Antes (H9) | Después (H36) | Por qué |
|---|---|---|
| `TipoAlzada` … `TipoGenerico`, `CodigosDeTipo()` | — | FR-010 |
| `ComputoLPAC`, `ComputoContencioso`, `Computos()` | — (el tipo `Computo string` se queda: es el del código) | FR-010 |
| `AvisoDiasDelArt183` | — (el aviso del art. 183 es un dato, spec §C.5) | FR-010 |
| `PreceptosDeUnComputo` | — (lo sustituye el cómputo descrito, sin exportar) | research D1 |
| `Tipo` | gana `Da string` (`yaml:"da"`) y un campo sin exportar con su cómputo | FR-002, FR-005 |
| `CodigosDeAviso() []string` | `CodigosDeAviso(tipos []Tipo) []string` | FR-009 |
| `EtiquetaDeAviso(codigo string) (string, bool)` | `EtiquetaDeAviso(tipos []Tipo, codigo string) (string, bool)` | FR-009 |
| `LeerTipos(contenido []byte) ([]Tipo, error)` | la misma firma, con las reglas de §C.2 | FR-013 |

Siguen las tres constantes de los avisos del motor (`AvisoFaltanNacionalesYAutonomicos`, `AvisoFaltanLocales`,
`AvisoFestivoDeUnaParte`). `CodigosDeAviso(tipos)` devuelve una lista nueva en cada llamada: los dos de los festivos
que faltan, los códigos de los avisos de periodo de los cómputos que nombran `tipos` —en el orden de esos cómputos en
el fichero y, en cada uno, de sus periodos— y `festivo-de-una-parte`. Con `nil`, los tres del motor. Con los tipos de
`data/plazos.yaml`, los cuatro de hoy en su orden.

## 2. El cómputo (spec §C.3)

Sin ningún código de tipo ni de cómputo en la decisión (FR-004, FR-010):

1. **Inhábil**: el de hoy (sábado, domingo, festivo nacional o autonómico sin parte de la comunidad de algún municipio
   que cuenta, festivo local de alguno) o un día de un periodo `inhabil` del cómputo del tipo.
2. **Con la unidad de la duración**, desde el día siguiente a `inicio`:
   - `dias-habiles`: el enésimo día no inhábil; los festivos saltados, los de cada día inhábil del tramo que no es sábado
     ni domingo (también los de un periodo `inhabil`).
   - `dias-naturales`: el enésimo día y la prórroga.
   - `meses`: la partida es `inicio`; para cada periodo `no-corre`, si `inicio` o el día siguiente cae en él, la partida
     pasa a su último día de ese año; el vencimiento, de fecha a fecha desde la partida; si del día siguiente a la
     partida al vencimiento cae algún día de un periodo `no-corre`, el vencimiento pasa al mismo día del mes siguiente
     (o al último de ese mes); después, la prórroga. Con agosto, es research D9 de H9 tal cual.
   - La prórroga lleva un último día inhábil al primer día que no lo es, con los festivos que salta.
3. **`da`**: con `vencimiento`, la fecha es el vencimiento; con `dia-siguiente`, el día siguiente al vencimiento, sin
   prórroga (FR-005). En los dos, un día después del 9999-12-31 es el error de hoy.
4. **El tramo** del plazo, para los avisos y la cobertura, va del día siguiente a `inicio` a la fecha que se da, como hoy.

`pasaPor(desde, hasta, periodo)` dice si algún día del tramo es del periodo, que puede cruzar el año (del `12-24` al
`01-06`), mirando cada año que toca el tramo; es la función de hoy con el periodo como dato.

## 3. Los avisos (spec §C.5; FR-008)

En `data.avisos`, en este orden: el de los nacionales y autonómicos que faltan; los de los locales que faltan; uno por
cada periodo `aviso` del cómputo por cuyo día pase el tramo, en el orden de sus periodos, con
`{codigo: <aviso.codigo>, texto: "⚠ <aviso.etiqueta>: <aviso.texto>"}`, una vez por plazo; los de un festivo de una
parte, que no avisan en un día inhábil (tampoco en uno de un periodo `inhabil`). Los textos de los avisos del motor no
cambian.

Ejemplo, el del caso 15 (`contencioso 2026-10-27 Madrid`), el mismo objeto de hoy en `data.avisos` (380 bytes: 335 del
texto, 45 del resto; medido con `wc -c`):

```json
{"codigo":"dias-del-art-183-lopj","texto":"⚠ DÍAS DEL ART. 183 LOPJ: el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ declara inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, porque el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más temprana posible."}
```

## 4. Los errores (spec §C.6)

`ComprobarPeticion` y `Calcular` comprueban, en este orden y con el primero que no vale: el tipo; la fecha de inicio; la
duración; la unidad (que falta con la duración, que no es del vocabulario, que se da a un tipo que la fija); **que el
cómputo del tipo admite la unidad dada a un tipo sin duración** (nuevo, FR-007); el número de municipios, que es el de
los textos de `municipios` de su cómputo (FR-006). Después, el ADR 0042 sobre los preceptos del tipo, el de su unidad y
los de su cómputo, como hoy. Los mensajes nuevos o compuestos, con la forma de §C.6:

| Defecto | Mensaje | Clase y código |
|---|---|---|
| unidad que el cómputo no admite | `el tipo "<tipo>" no admite la unidad "<unidad>": su cómputo, "<cómputo>", admite <unidades>`, con las unidades del cómputo en el orden de `Unidades()` unidas como `alternativas` («a o b», «a, b o c») | `argumentos`, 2 |
| municipios, con un texto | `el tipo "<tipo>" se cuenta con un municipio, <texto>, y se han dado <n>` | `argumentos`, 2 |
| municipios, con dos textos | `el tipo "<tipo>" se cuenta con dos municipios, <texto 1> y <texto 2>, que pueden ser el mismo, y se han dado <n>` | `argumentos`, 2 |

Con los datos de hoy, los dos mensajes de municipios son los de `main` byte a byte, y el de la unidad no se da (`lpac`
admite las tres y `contencioso` no recibe duración). Medidas: el de C7 mide 115 bytes (research V19).

## 5. `internal/evals` (FR-009)

`internal/evals/plazos.go` lee `plazos.LeerTipos(data.Plazos)` una vez (research D4) y saca de ahí, con
`plazos.CodigosDeAviso(tipos)` y `plazos.EtiquetaDeAviso(tipos, …)`, las formas fijas de `formasDeAvisoDelPlazo`. Las
firmas de `ExtraerAvisosDelPlazo(texto string) []string` y `ComprobarCodigosDeAvisoDelPlazo(esquema
*jsonschema.Schema) error` no cambian; la segunda exige que `schemas/eval.yaml.json` enumere exactamente esos códigos
(los cuatro de hoy en su orden) y devuelve el error de lectura de los tipos si lo hay. `schemas/eval.yaml.json` no
cambia (spec, «Fuera de alcance»).

## 6. Quién consume qué

| Salida | Quién la pide | Cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|---|
| `data` de `plazos calcular` | `plazos-y-territorio`, paso 9 (`SKILL.md:59`), y quien use el binario | una por plazo de la pregunta, una o dos | la de `main`, byte a byte (FR-011): de 670 a 853 B en los goldens de hoy (V19) | no es una señal: responde a la pregunta |
| aviso de un periodo | la misma skill, que lo copia tal cual | como mucho una vez por plazo y periodo | 335 B el texto del art. 183, unos 380 B con su objeto | en el plazo cuyo tramo no toca el periodo |
| error de FR-007 | la misma skill, en `data.mensaje` | una por llamada con una unidad que el cómputo no admite | 115 B en C7 | cuando la llamada pide una unidad que el cómputo admite; con los datos de hoy, nunca se da |
| `CodigosDeAviso`, `EtiquetaDeAviso` | `internal/evals`, al juzgar | una lectura por proceso | 4 códigos hoy, uno más por aviso de periodo de los datos | — |

`plazos` no lee la caché ni el grafo: con cientos de normas y miles de bloques consultados, cada llamada mide lo mismo
que la primera.
