# Contrato: la lista de expresiones prohibidas, su juicio y su informe

FR-050 a FR-056, FR-082 a FR-085; SC-001, SC-003, SC-004, SC-006. Decisiones en research D1-D11.

## 1. El fichero y su esquema

`evals/<skill>/expresiones-prohibidas.yaml`, opcional, uno por skill. `LeerConjunto` lo reconoce por ese nombre exacto;
el resto de entradas siguen siendo ficheros de eval (`<nn>-<descripción>.yaml`).

`schemas/expresiones-prohibidas.yaml.json` (tarea `[datos]`):

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://kitlegal.es/schemas/expresiones-prohibidas.yaml.json",
  "title": "evals/<skill>/expresiones-prohibidas.yaml",
  "type": "object",
  "additionalProperties": false,
  "required": ["maquinaria", "otra_conversacion"],
  "properties": {
    "maquinaria": { "$ref": "#/$defs/expresiones" },
    "otra_conversacion": { "$ref": "#/$defs/expresiones" }
  },
  "$defs": {
    "expresiones": {
      "type": "array", "minItems": 1,
      "items": { "type": "string", "pattern": "^[^\\s*_]+( [^\\s*_]+)*$" }
    }
  }
}
```

Lectura: con el lector común de documentos YAML (`skills.ValidarDocumentoYAML`: una clave repetida es un defecto con sus
dos líneas) y ese esquema, compilado una vez como el de eval. Una entrada con ese nombre que no es un fichero regular,
que no se puede leer o que no valida es un `FicheroMalFormado` cuyo error empieza por `expresiones-prohibidas.yaml`
(FR-055). Leída, queda en `Conjunto.Prohibidas` y en `Eval.Prohibidas` de cada eval de la carpeta; si falta o está mal
formada, las dos quedan vacías y las evals se juzgan sin lista.

Consecuencias de un fichero mal formado, las de cualquier otro (research D8): `make ci` falla nombrándolo
(`TestEvalsDelRepositorio`, subtest `formato`); `PrepararSesion` y `TestPlanDeSesiones` no preparan ni planifican; el
informe lo lista en `ficheros_mal_formados` y su motivo da `fallo`.

## 2. La lista de `boe-legislacion`

`evals/boe-legislacion/expresiones-prohibidas.yaml`, con este contenido exacto (research D4):

```yaml
# Expresiones que no lleva la respuesta de boe-legislacion en una eval que activa la skill (H7.2, FR-050): la maquinaria
# interna y lo dicho en otra conversación. Se comparan por la forma, sin distinguir mayúsculas y con blancos y énfasis de
# Markdown entre palabras, delimitadas como palabras; lo dicho con otras palabras no se detecta.
maquinaria:
  - memoria de consultas
  - hallazgo
  - hallazgos
  - graph check
  - graph show
  - graph stats
  - kitlegal graph
  - version-obsoleta
  - fuente-caducada
  - código de salida
  - códigos de salida
  - código 0
  - código 1
  - código 2
  - código 3
  - código 4
  - código 5
  - código 6
  - código 7
  - exit code
  - json
  - sobre de salida
otra_conversacion:
  - te dije
  - te respondí
  - te confirmé
  - te indiqué
  - te comenté
  - te contesté
  - te expliqué
  - te habría dicho
  - te habría respondido
  - te habría confirmado
  - te habría indicado
  - te habría comentado
  - te habría contestado
  - te habría explicado
  - conversación anterior
  - conversaciones anteriores
```

Tamaño: 38 expresiones, ≈ 1 KB con la cabecera. Fija: no crece con el uso del kit.

## 3. La comparación

`ExtraerExpresionesProhibidas(texto string, lista ExpresionesProhibidas) []string`: las expresiones de la lista, en su
orden (maquinaria y después otra conversación) y sin repetir, cuya expresión regular casa en algún punto del texto; `nil`
si ninguna. La expresión regular de cada una, compilada una sola vez:

```text
(?:^|[^\p{L}\p{N}])(?i:<palabra₁>)[\t\p{Zs}*_]*[\t\p{Zs}][\t\p{Zs}*_]*(?i:<palabra₂>)…(?:$|[^\p{L}\p{N}])
```

con cada palabra de `strings.Fields(expresion)` pasada por `regexp.QuoteMeta` y el separador `entrePalabrasDeAviso` de
`internal/evals/avisos.go`, el de H5.1.

| Texto | Encontradas |
|---|---|
| `Sin hallazgos en la memoria de consultas. Ya tengo todo lo necesario para responder.` | `memoria de consultas`, `hallazgos` |
| `SIN HALLAZGOS` | `hallazgos` |
| `Sin  hallazgos en la **memoria de consultas**` | `memoria de consultas`, `hallazgos` |
| `la *memoria* de _consultas_` | `memoria de consultas` |
| `` Comprobado con `graph check` (código 0). `` | `graph check`, `código 0` |
| `Este texto coincide con lo que te habría confirmado antes` | `te habría confirmado` |
| `te confirmé.` | `te confirmé` |
| `⚠ REDACCIÓN MODIFICADA: la redacción con fecha de vigencia 20180309, la que se consultó antes, ha sido sustituida por la de 20200206, que es la que se cita.` | ninguna |
| `⚠ NORMA DEROGADA: esta norma ha sido derogada.` | ninguna |
| `No hay avisos de vigencia sobre este bloque.` | ninguna |
| `No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.` | ninguna |
| `hallazgoss` | ninguna |
| `Sin hallazgos en la memoria` + salto de línea + `de consultas` | `hallazgos` (el salto no es un blanco tolerado) |

## 4. El juicio

`Juzgar(eval, sesion, skill)` no cambia de firma. Si `eval.Activa` y `eval.Prohibidas` no está vacía:

1. `ResultadoDeEval.ExpresionesProhibidas = ExtraerExpresionesProhibidas(sesion.Respuesta, eval.Prohibidas)`;
2. un motivo `expresión prohibida: <expresión>` por cada una, detrás de los del territorio ausente y delante del del
   modelo que pone `EscribirInforme`;
3. `Pasa` exige además que no haya ninguna.

Con `activa: false`, o en una skill sin lista, la lista queda vacía y el juicio es el de antes del hito (FR-052). Nada
más cambia en `Juzgar`. La sesión de la prueba de red se juzga con la eval 01 y, por tanto, con la lista.

`ResultadoDeEval` gana, detrás de `territorio_ausente`:

```json
"expresiones_prohibidas": ["memoria de consultas", "hallazgos"],
```

(≈ 90 bytes con el sangrado de `informe.json`, un elemento por línea; `[]` si ninguna, research V4).

## 5. El informe

**`informe.json`.** Por sesión, el campo de §4. En la raíz, detrás de `tasas`:

```json
"expresiones_prohibidas_por_modelo": [
  {"modelo": "claude-sonnet-5", "con_alguna": 2, "respuestas": 51},
  {"modelo": "claude-haiku-4-5-20251001", "con_alguna": 0, "respuestas": 30}
],
```

Un elemento por modelo del job (el que decide y después los informativos). `respuestas`: sesiones juzgadas de las series
planificadas cuya eval activa la skill; `con_alguna`: de ellas, las que llevan alguna expresión. La prueba de red no
cuenta (serie no planificada; FR-053). Skill sin lista: `[]`.

**`informe.md`.** En la tabla «Sesiones», una columna `Expresiones prohibidas` entre `Territorio ausente` y `Resultado`,
con las encontradas separadas por `, ` o `ninguna`. Detrás de «## Tasas por eval», una sección:

```text
## Expresiones prohibidas por modelo

| Modelo | Respuestas con alguna expresión | Respuestas en evals que activan la skill |
| --- | --- | --- |
| claude-sonnet-5 | 2 | 51 |
| claude-haiku-4-5-20251001 | 0 | 30 |
```

o, sin lista, el párrafo `la skill no tiene lista de expresiones prohibidas`.

El veredicto no cambia de regla (FR-054): una sesión con una expresión es una sesión que no pasa, y su serie decide con
el umbral de siempre si es de las que deciden.

## 6. Controles en `make ci`

| Control | Dónde | Cubre |
|---|---|---|
| El esquema compila y fija la forma: una lista válida, y una mal formada (sin una familia) no valida | `internal/evals/formato_test.go` (tarea `[datos]` con el esquema) | FR-050, FR-082 |
| `LeerConjunto` con una lista bien formada la deja en `Conjunto.Prohibidas` y en cada `Eval.Prohibidas`; con una mal formada, un `FicheroMalFormado` que la nombra y evals sin lista | `internal/evals/conjunto_test.go` (`TestLeerConjunto`) | FR-050, FR-055, FR-082 |
| `ExtraerExpresionesProhibidas`: orden de la lista, sin repetir, extremos de palabra | `internal/evals/prohibidas_test.go` | FR-051 |
| `Juzgar` con la lista del repositorio: sin expresiones; una de cada familia; mayúsculas, blancos y énfasis; la línea `⚠ REDACCIÓN MODIFICADA:` con sus dos fechas; una eval de no activación y una sin lista, como antes | `internal/evals/juzgar_test.go` | FR-051, FR-052, FR-083, SC-006 |
| Informe: expresiones por sesión (JSON y columna), recuento por modelo igual con y sin prueba de red, skill sin lista, lista mal formada → `fallo`, una serie que decide con dos sesiones con expresiones → `fallo` | `internal/evals/informe_test.go` | FR-053, FR-054, FR-055 |
| Calibrado: las 93 respuestas de H7.1, 35 marcadas con el reparto de FR-084 por eval y familia, 0 de las otras 58 | subtest `expresiones-calibradas` de `TestEvalsDelRepositorio` | FR-084, SC-003, US4.2 |
| Bloques: ninguna expresión en el `texto` que da `boe articulo` de los bloques de las evals y de sus grafos previos | subtest `expresiones-en-los-bloques` | FR-085, SC-004 |
| Skill: ninguna expresión en las formas fijas de avisos y hallazgos ni en los bloques ```` ```text ```` de `SKILL.md` | subtest `expresiones-de-la-skill` | FR-043, FR-051 |

Los tres subtests leen la lista de `evals/boe-legislacion/` con el `LeerConjunto` de `TestEvalsDelRepositorio` y exigen
antes su premisa, para no pasar en vacío: la lista no está vacía, y hay respuestas, bloques o formas y bloques `text` que
mirar.

## 7. Uso, de fuera adentro

| Salida | Quién y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| `expresiones_prohibidas` de una sesión | El job, que juzga con ella, y la persona del informe final; una vez por sesión y job | `[]` lo habitual; ≈ 90 B con dos (§4); como mucho las 38, ≈ 1,2 KB; por job, 93 sesiones × ≤ 1,2 KB ≤ 112 KB en el peor caso | Cada job la mide otra vez sobre su commit; no se acumula |
| `expresiones_prohibidas_por_modelo` | La persona del informe final, que la compara con el umbral de SC-001; una vez por job | Dos elementos de ≈ 100 B con el sangrado | Como arriba |
| La lista | El job, al juzgar; las subpruebas de §6 en `make ci` | ≈ 1 KB, fija | No da señales |

Nada de esto depende del uso del kit (cientos de normas, miles de bloques): la lista es fija y el informe crece con las
sesiones del job, no con lo consultado.
