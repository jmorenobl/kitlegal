# Contrato: la forma fija de los avisos

FR-001 a FR-005, FR-010 a FR-012, FR-014, FR-032, FR-033; US1, US4; SC-001, SC-003, SC-005. Decisiones: research.md
D1, D2, D7, D8, D15. Entidades: data-model §1, §2 y §6.

## 1. Gramática

Un aviso está en un texto si el texto contiene su forma, en una misma línea:

```text
forma     = marca [variante] sep palabra₁ { entre palabraᵢ } sep ":"
marca     = U+26A0 (⚠)
variante  = U+FE0F | U+FE0E
blanco    = U+0009 | un carácter de la categoría Zs de Unicode
sep       = { blanco | "*" | "_" }
entre     = sep blanco sep
palabraᵢ  = la i-ésima palabra de la etiqueta (strings.Fields), comparada sin distinguir mayúsculas
```

En Go, una expresión por código con etiqueta, compilada una sola vez (`sync.OnceValue`) desde `boe.EtiquetasDeAviso()`:

| Pieza | Expresión |
|---|---|
| marca y variante | `\x{26A0}[\x{FE0E}\x{FE0F}]?` |
| `sep` | `[\t\p{Zs}*_]*` |
| `entre` | `[\t\p{Zs}*_]*[\t\p{Zs}][\t\p{Zs}*_]*` |
| palabra | `(?i:` + `regexp.QuoteMeta(palabra)` + `)` |
| final | `:` |

Un código cuya etiqueta no tiene palabras no tiene expresión y nunca se encuentra. Lo que sigue a los dos puntos no se lee.
Las tolerancias y lo que no se tolera, con ejemplos comprobados, están en research D2 (V6, V7).

## 2. API

### `internal/source/boe` (`avisos.go`)

```go
// Constantes no exportadas, una por etiqueta, en el orden de las condiciones:
// etiquetaDeConsolidacionNoFinalizada = "TEXTO POSIBLEMENTE DESACTUALIZADO"
// etiquetaDeNormaDerogada             = "NORMA DEROGADA"
// etiquetaDeVigenciaAgotada           = "VIGENCIA AGOTADA"
// Las frases se componen con ellas: fraseDeNormaDerogada = "⚠ " + etiquetaDeNormaDerogada + ": esta norma ha sido derogada."

// EtiquetasDeAviso es la etiqueta de cada código de CodigosDeAviso: lo que su frase lleva entre la marca y los dos
// puntos, y la única fuente de verdad de la forma fija con la que se traslada un aviso. Cada llamada devuelve un mapa
// nuevo, que quien lo recibe puede cambiar sin cambiar el de nadie más.
func EtiquetasDeAviso() map[string]string
```

Nada más cambia en el paquete: ni `CodigosDeAviso`, ni `avisosDe`, ni las condiciones, ni el texto de ninguna frase, ni
`Aviso`, ni lo que emiten `articulo`, `articulos` y `metadatos` (FR-012).

### `internal/evals` (`avisos.go`)

```go
// ExtraerAvisos devuelve los códigos de boe.CodigosDeAviso cuya forma fija —la marca, la etiqueta de
// boe.EtiquetasDeAviso y los dos puntos, con las tolerancias de la gramática— lleva el texto, en el orden de
// CodigosDeAviso y sin repetir, o nil si no lleva ninguna. No lee lo que sigue a la forma ni ninguna otra redacción.
func ExtraerAvisos(texto string) []string

// ComprobarFormasDeAviso comprueba que el texto lleva la forma fija de cada código de boe.CodigosDeAviso, con
// ExtraerAvisos: devuelve un error por código que falta, en el orden de CodigosDeAviso y unidos con errors.Join, o
// nil si están todos.
func ComprobarFormasDeAviso(texto string) error
```

Mensaje de cada defecto de `ComprobarFormasDeAviso`: `falta la forma fija del aviso <código>: ⚠ <etiqueta>:`, con la
etiqueta tomada de `boe.EtiquetasDeAviso()`.

**Sin copias de las etiquetas.** Ningún fichero `*.go` de `internal/evals` que no termine en `_test.go` contiene ninguna
etiqueta de `boe.EtiquetasDeAviso()`, tampoco en un comentario (US4-5, SC-005). Los tests sí escriben las formas
literales.

## 3. `skills/boe-legislacion/SKILL.md`

Tres cambios, y ninguno más (FR-004). Nada fuera de la región generada nombra evals, el job ni modelos; la región
generada no cambia; el fichero pasa de 181 a 199 líneas, por debajo de 300 (FR-005): 2 más en el paso 5, 15 en «Cómo se
cita» y 1 en la regla 3 (research V33).

**Paso 5** («### 5. Responder citando»): el punto

~~~markdown
- Traslada los avisos de vigencia del sobre y recuerda que los textos consolidados del BOE tienen carácter informativo.
~~~

pasa a ser

~~~markdown
- Traslada cada aviso de vigencia del sobre con su forma fija: `⚠`, la etiqueta del aviso tal como la da el binario y
  dos puntos, seguidos de la frase del binario o de una explicación (más en «Cómo se cita»). Recuerda que los textos
  consolidados del BOE tienen carácter informativo.
~~~

**«## Cómo se cita»**: detrás de su último punto («- Una cita por bloque. …») se añade

~~~markdown

Los avisos de vigencia también tienen forma fija. Cada aviso del sobre va en la respuesta con `⚠`, la etiqueta del
aviso tal como la da el binario —lo que su `texto` lleva entre `⚠` y los dos puntos— y dos puntos, seguidos de la frase
del binario o de una explicación:

- `⚠ NORMA DEROGADA:` para el aviso `derogada`.
- `⚠ VIGENCIA AGOTADA:` para el aviso `vigencia-agotada`.
- `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` para el aviso `consolidacion-no-finalizada`.

```text
⚠ NORMA DEROGADA: esta norma ha sido derogada.
```

- La etiqueta va entera y sin cambiar ninguna palabra, con `⚠` delante y los dos puntos detrás, todo en la misma línea.
  Decir con otras palabras que la norma está derogada no traslada el aviso.
~~~

**Regla 3** («## Reglas»): la regla

~~~markdown
3. **Trasladar la vigencia.** Traslada los avisos de vigencia que devuelve el binario (derogada, vigencia agotada,
   consolidación no finalizada) y no presentes como vigente el texto de una norma derogada. Recuerda que los textos
   consolidados del BOE tienen carácter informativo y no son asesoramiento.
~~~

pasa a ser

~~~markdown
3. **Trasladar la vigencia.** Traslada cada aviso de vigencia que devuelve el binario (derogada, vigencia agotada,
   consolidación no finalizada) con su forma fija —`⚠`, la etiqueta del aviso tal como la da el binario y dos puntos,
   con la frase del binario o una explicación detrás— y no presentes como vigente el texto de una norma derogada.
   Recuerda que los textos consolidados del BOE tienen carácter informativo y no son asesoramiento.
~~~

El ejemplo es, byte a byte, la frase que el binario emite para `derogada` (data-model §1), sin ninguna explicación
añadida: todo lo que enseña sale del sobre. No nombra ninguna norma: no reproduce la pregunta de ninguna eval y no
activa la comprobación de normas nombradas (research D15, V29). Ninguna línea añadida pasa de 120 caracteres (la más
larga tiene 119), como todas las del fichero fuera de la región generada (research V33).

## 4. Tests

| Test (fichero) | Qué fija |
|---|---|
| `TestEtiquetasDeAviso` (`internal/source/boe/avisos_test.go`) | `exactamente-tres`: el mapa es igual al literal de los tres códigos con sus tres etiquetas y sus claves son las de `CodigosDeAviso()`; `frases-con-su-forma`: con las tres condiciones, cada `Aviso.Texto` de `avisosDe` empieza por `"⚠ " + etiqueta + ":"` de su código; `cada-llamada-su-mapa`: cambiar el mapa devuelto no cambia el de la llamada siguiente. `TestAvisosDe` y `TestCodigosDeAviso` no cambian y siguen comparando las frases literales |
| `TestExtraerAvisos` (`internal/evals/avisos_test.go`) | Un subtest por fila, con el texto literal y los códigos esperados: `forma-fija` (`⚠ NORMA DEROGADA: esta norma ha sido derogada.` → `derogada`); `las-tres-etiquetas` (para cada código de `boe.CodigosDeAviso()`, `"⚠ " + etiqueta + ":"` → ese código); `selector-emoji` (U+FE0F), `selector-texto` (U+FE0E), `enfasis-envolvente` (`**⚠ NORMA DEROGADA:**`), `enfasis-en-la-etiqueta` (`⚠ **NORMA DEROGADA**:`), `enfasis-con-guiones-bajos` (`_⚠ NORMA DEROGADA:_`, `⚠ __NORMA DEROGADA__:`), `espacios` (`⚠  NORMA \t DEROGADA :`), `blancos-de-unicode` (U+00A0 y U+202F), `sin-blanco-tras-la-marca` (`⚠NORMA DEROGADA:`), `minusculas` (`⚠ norma derogada:`), `mayusculas-mezcladas` (`⚠ Norma Derogada:`), `dentro-de-una-cita-en-bloque` (`> ⚠ NORMA DEROGADA:`), `en-otra-linea` (texto, salto y la forma) → `derogada`; `orden-de-los-codigos` (`⚠ VIGENCIA AGOTADA:` antes que `⚠ NORMA DEROGADA:` → `derogada`, `vigencia-agotada`); `repetida` (la forma dos veces → una vez); `forma-fija-y-lo-contrario` (`⚠ NORMA DEROGADA: pero sigue en vigor.` → `derogada`, la limitación declarada); y → ninguno: `otra-redaccion` («Esta ley fue derogada.»), `negacion` («La Ley 30/1992 sigue en vigor.»), `palabra-de-mas` (`⚠ NORMA PARCIALMENTE DEROGADA:`), `palabras-pegadas` (`⚠ NORMADEROGADA:`), `palabra-distinta` (`⚠ NORMA DEROGADAS:`), `sin-marca` (`NORMA DEROGADA:`), `sin-dos-puntos` (`⚠ NORMA DEROGADA`), `dos-puntos-de-ancho-completo` (`⚠ NORMA DEROGADA：`), `dos-selectores` (U+FE0F dos veces), `salto-dentro` (`⚠ NORMA` salto `DEROGADA:`), `tachado` (`⚠ NORMA ~~DEROGADA~~:`), `otro-emoji` (`🚨 NORMA DEROGADA:`), `vacio` |
| `TestComprobarFormasDeAviso` (`internal/evals/avisos_test.go`) | `completo`: las tres formas → `nil`; `falta-la-etiqueta` (`⚠ :` en lugar de la forma de `derogada`), `falta-la-marca` (`NORMA DEROGADA:`), `faltan-los-dos-puntos` (`⚠ NORMA DEROGADA`): el error es exactamente `falta la forma fija del aviso derogada: ⚠ NORMA DEROGADA:`; `ninguna`: los tres errores, en el orden de `CodigosDeAviso()` |
| `TestEtiquetasSoloDesdeBoe` (`internal/evals/avisos_test.go`) | Ningún `*.go` del directorio del paquete que no termine en `_test.go` contiene una etiqueta de `boe.EtiquetasDeAviso()`; y hay al menos uno leído (`avisos.go`), para no pasar en vacío |
| `TestEvalsDelRepositorio/avisos-de-la-skill` (`internal/evals/conjunto_test.go`) | `ComprobarFormasDeAviso` sobre `../../skills/boe-legislacion/SKILL.md` → `nil`. Llega con el cambio de `SKILL.md` (plan, orden de implementación, paso 11) |

Demostración sobre un clon (quickstart §4): quitar de `SKILL.md` la marca, la etiqueta o los dos puntos de la forma de
`derogada` hace fallar `avisos-de-la-skill` nombrando `derogada`.
