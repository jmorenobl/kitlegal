# Contrato: `avisos` en el formato, el juicio y el informe

FR-013, FR-020 a FR-023, FR-030 a FR-035, FR-040 a FR-043; US2, US3, US4-2; SC-002 a SC-005. Decisiones: research.md
D3 a D6, D9, D10. Entidades: data-model §3 a §6. Amplía, sin cambiar nada más, los contratos de H5
`specs/006-h5-skill-boe-legislacion/contracts/evals-y-grabaciones.md` §1 y §6 y `job-de-evals.md` §5.

## 1. `schemas/eval.yaml.json` (tarea `[datos]`, con pausa)

Tres cambios, y ninguno más:

```json
    "citas": {
      "type": "array", "minItems": 1,
      "items": { "$ref": "#/$defs/cita" }
    },
    "avisos": {
      "type": "array", "minItems": 1,
      "items": { "$ref": "#/$defs/codigo-de-aviso" }
    }
```

```json
  "else": { "not": { "anyOf": [{ "required": ["comandos"] }, { "required": ["citas"] }, { "required": ["avisos"] }] } },
```

```json
  "$defs": {
    "codigo-de-aviso": { "enum": ["consolidacion-no-finalizada", "derogada", "vigencia-agotada"] },
```

El enumerado va en una sola línea, en el orden de `boe.CodigosDeAviso()` y como primera entrada de `$defs`. Con el
esquema cambiado y el código todavía sin el campo, `make ci` sigue en verde (research V5).

## 2. `internal/evals/formato.go`

```go
	// Avisos son los códigos de aviso de vigencia cuya forma fija tiene que llevar la respuesta, en el orden del
	// fichero; vacío si la eval no los espera. Solo los admite una eval que activa la skill, y sus valores son los de
	// boe.CodigosDeAviso (FR-020 a FR-022 de H5.1).
	Avisos []string `yaml:"avisos"`
```

detrás de `Citas` en `Eval`. `LeerEval` no cambia: el esquema hace la validación.

## 3. Comprobación de FR-013 (`internal/evals/avisos.go`)

```go
// ComprobarCodigosDeAviso comprueba que los códigos que el esquema de eval compilado admite en avisos —los del
// enumerado al que llegan su items y cada $ref— son exactamente los de boe.CodigosDeAviso: devuelve un error por código
// del binario que el esquema no enumera, en el orden de CodigosDeAviso, y otro por código que enumera sin ser del
// binario, en el orden del esquema, unidos con errors.Join; o nil si coinciden.
func ComprobarCodigosDeAviso(esquema *jsonschema.Schema) error
```

Mensajes: `el esquema de eval no enumera el código de aviso <código> en avisos` y `el esquema de eval enumera en avisos
el código <código>, que no es un código de aviso del binario`. Un esquema sin la propiedad `avisos`, sin `items` o sin
enumerado da el primer mensaje para cada uno de los tres códigos.

## 4. `internal/evals/juzgar.go`

- Constante `motivoDeAvisoAusente = "aviso ausente: "`, junto a las de los motivos.
- `ResultadoDeEval`, detrás de `CitasAusentes`:

  ```go
	// AvisosEncontrados y AvisosAusentes reparten los avisos esperados, en el orden de la eval y con sus
	// repeticiones, entre los que la respuesta lleva con su forma fija y los que no (ExtraerAvisos), cada uno con su
	// código.
	AvisosEncontrados []string `json:"avisos_encontrados"`
	AvisosAusentes    []string `json:"avisos_ausentes"`
  ```

- `Juzgar`, justo después de `repartirCitas`: `resultado.repartirAvisos(eval.Avisos, ExtraerAvisos(sesion.Respuesta))`,
  con la misma forma que `repartirCitas`: encontrado si está entre los extraídos; si no, ausente y
  `motivoDeAvisoAusente + código`.
- `Pasa = sesion.Terminada && Activa == Activada && len(ComandosAusentes) == 0 && len(CitasAusentes) == 0 &&
  len(AvisosAusentes) == 0`.
- Se actualizan los comentarios de `Juzgar`, `Motivos` y `Pasa` para nombrar los avisos.

Orden de los motivos de una sesión: sin terminar o ilegible; activación; cada comando ausente; cada cita ausente; **cada
aviso ausente**; el del modelo distinto (lo añade `exigirElModeloPedido`, que no cambia).

## 5. `internal/evals/informe.go`

- `encabezadosDeSesiones`: `"Sesión", "Eval", "Modelo", "Activa", "Activada", "Sesión terminada", "Comandos ausentes",
  "Citas ausentes", "Avisos encontrados", "Avisos ausentes", "Resultado"`.
- `filasDeSesiones`: detrás de la celda de citas ausentes, `unidosOVacio(resultado.AvisosEncontrados,
  ningunoEnElInforme)` y `unidosOVacio(resultado.AvisosAusentes, ningunoEnElInforme)`; su comentario nombra las dos
  columnas nuevas.
- Nada más cambia: las secciones de sesión siguen publicando la respuesta, y los motivos de la raíz ya llevan los de cada
  sesión.

Fragmento de `informe.json` de una sesión de la eval 18 que trasladó uno de los dos avisos:

```json
      "citas_encontradas": ["BOE-A-1992-26318 a42"],
      "citas_ausentes": [],
      "avisos_encontrados": ["derogada"],
      "avisos_ausentes": ["vigencia-agotada"],
      …
      "motivos": ["aviso ausente: vigencia-agotada"],
      "pasa": false
```

y, en una sesión de una eval sin `avisos`, `"avisos_encontrados": []` y `"avisos_ausentes": []`.

## 6. Tests

| Test (fichero) | Qué fija |
|---|---|
| `TestLeerEval` (`internal/evals/formato_test.go`), casos nuevos | `avisos`: una positiva con `avisos` `derogada` y `vigencia-agotada` se lee con `Avisos` en ese orden (US2-10); `aviso-desconocido`: `otro` → error `01-lpac-articulo-21.yaml: avisos/0, línea <n>: value must be one of 'consolidacion-no-finalizada', 'derogada', 'vigencia-agotada'` (US2-1); `no-activa-con-avisos` → `01-lpac-articulo-21.yaml: línea 1: 'not' failed` (US2-2); `avisos-vacio` y `citas-vacio` → los dos con `minItems: got 0, want 1` en su ruta; `aviso-repetido` y `cita-repetida` → las dos se leen, con el repetido en su lista (US2-9); `informativa-con-avisos` se lee (US5-1). Las líneas `<n>` las fija el documento de cada caso |
| `TestComprobarCodigosDeAviso` (`internal/evals/avisos_test.go`) | Sobre esquemas sintéticos compilados con `skills.CompilarEsquema`: `exacto` (el enumerado detrás de un `$ref`, como el publicado) → `nil`; `falta-un-codigo` (sin `derogada`) → exactamente `el esquema de eval no enumera el código de aviso derogada en avisos`; `sobra-un-codigo` (con `otro-aviso`) → exactamente `el esquema de eval enumera en avisos el código otro-aviso, que no es un código de aviso del binario`; `sin-enumerado` (`items` de tipo cadena) → los tres «no enumera»; `sin-avisos` → los tres «no enumera» (US4-2, SC-005) |
| `TestEvalsDelRepositorio/avisos-del-esquema` (`internal/evals/conjunto_test.go`) | `ComprobarCodigosDeAviso(esquemaDeEval())` → `nil` sobre el esquema publicado (FR-013) |
| `TestJuzgar` (`internal/evals/juzgar_test.go`), casos nuevos sobre la eval 01 con `avisos` y la sesión que pasa, cambiando solo la respuesta | `aviso-con-su-forma-fija` (US2-3): `derogada` encontrado, ausentes vacío y pasa; `aviso-con-variantes-toleradas` (US2-4): un juicio por variante —selector U+FE0F, énfasis envolvente, énfasis en la etiqueta, espacios y U+00A0, minúsculas— y en todos encontrado y pasa; `aviso-ausente` (US2-7): eval con `derogada` y `vigencia-agotada` y respuesta sin forma → los dos ausentes en ese orden, dos motivos y no pasa; `aviso-con-otra-redaccion` (US2-5): «Esta ley fue derogada.» → ausente y no pasa; `aviso-negado` (US2-6): «La Ley 30/1992 sigue en vigor.» con el comando y la cita presentes → ausente y no pasa; `forma-fija-y-lo-contrario`: `⚠ NORMA DEROGADA: pero sigue en vigor.` → encontrado y pasa (limitación declarada); `aviso-no-esperado`: la forma de `vigencia-agotada` en la respuesta de una eval que solo espera `derogada` y la lleva → resultado igual que sin ella; `avisos-en-el-orden-de-la-eval`: eval `vigencia-agotada`, `derogada` y respuesta con las dos formas en el otro orden → encontrados en el orden de la eval; `aviso-repetido`: `derogada` dos veces en la eval → dos veces encontrado; `cita-y-aviso-ausentes`: sin cita ni forma → motivos `cita ausente: …` y después `aviso ausente: derogada`. Los casos existentes, sin `avisos`, siguen con las dos listas nulas y el mismo resultado (US2-8, FR-034) |
| `TestInformeConAvisos` (`internal/evals/informe_test.go`) | Copia `testdata/sesiones/informe/aprobado` con `os.CopyFS` a un `t.TempDir()`, añade a `evals/01-lpac-articulo-21.yaml` `avisos` `derogada` y `vigencia-agotada`, antepone `⚠ NORMA DEROGADA: esta norma ha sido derogada. ` a la respuesta en `sesiones/01-lpac-articulo-21/sesion.jsonl` exigiendo dos sustituciones, y llama a `EscribirInforme` con las entradas del caso sobre la copia. `uno-encontrado-y-otro-ausente` (US3-1, US3-2): la sesión 01 con `avisos_encontrados` `["derogada"]` y `avisos_ausentes` `["vigencia-agotada"]` en `informe.json` en crudo, motivos `["aviso ausente: vigencia-agotada"]`, no pasa, la raíz con `01-lpac-articulo-21: aviso ausente: vigencia-agotada` detrás de la tasa, la sesión 11 con `[]` y `[]`, la cabecera de la tabla de `## Sesiones` con las once columnas, la fila de la 01 con `derogada` y `vigencia-agotada` en sus columnas y la de la 11 con `ninguno` y `ninguno`, y la sección de la sesión 01 con la respuesta que lleva la forma; `aviso-detras-de-la-cita`: además quita de la respuesta `[BOE-A-2015-10565, bloque a21]` (dos sustituciones) → motivos de la sesión `cita ausente: BOE-A-2015-10565 a21` y después `aviso ausente: vigencia-agotada` |
| `TestInforme/aprobado` (`internal/evals/informe_test.go`) | Además de lo de H5, cada sesión lleva `"avisos_encontrados": []` y `"avisos_ausentes": []` escritos así en `informe.json` (US3-3, FR-040): `informeCrudo` gana los dos campos como `jsontext.Value` |

Ningún test de este contrato añade ni modifica material bajo ningún directorio `testdata/`.
