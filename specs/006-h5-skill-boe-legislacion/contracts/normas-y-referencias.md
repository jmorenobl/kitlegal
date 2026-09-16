# Contrato: `data/normas.yaml`, `schemas/normas.yaml.json` y `references/normas.md`

FR-020 a FR-025, FR-030, FR-031, FR-034, FR-043, US6, SC-008. Decisiones: research.md D8 (forma indexada por
identificador) y D9 (verificación offline de identificadores).

## 1. `data/normas.yaml`

```yaml
# Normas que referencian las skills de kitlegal. Única fuente de verdad de references/normas.md.
# Cada identificador está verificado contra la búsqueda grabada de la fuente (TestIdentificadoresDeLasNormas).
normas:
  BOE-A-2015-10565:
    titulo: "Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común de las Administraciones Públicas."
    rango: Ley
    abreviatura: LPAC
    materias:
      - procedimiento administrativo
```

- Una entrada por norma de data-model §4.1. En H5 son las diez normas de las evals positivas (§4), que incluyen las
  cinco del hito y la ley del IRPF de la eval que reproduce `boe-fiscal`.
- El identificador y el título se copian **de la búsqueda grabada** (§2 del contrato de evals y grabaciones), nunca de
  memoria (FR-023). De las diez, tres están ya verificadas por grabaciones de H4 (títulos leídos de sus metadatos
  grabados): `BOE-A-2015-10565` («Ley 39/2015, …»), `BOE-A-2017-12902` («Ley 9/2017, …») y `BOE-A-1985-5392`
  («Ley 7/1985, …»). Las otras siete las fija la pausa de grabación.
- Sin campo `vertical` ni ningún otro no declarado (FR-021).

## 2. `schemas/normas.yaml.json`

JSON Schema 2020-12 con `$id` `https://ventanillalegal.es/schemas/normas.yaml.json`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://ventanillalegal.es/schemas/normas.yaml.json",
  "title": "data/normas.yaml",
  "type": "object",
  "additionalProperties": false,
  "required": ["normas"],
  "properties": {
    "normas": {
      "type": "object",
      "minProperties": 1,
      "propertyNames": { "pattern": "^BOE-A-[0-9]{4}-[0-9]{1,9}$" },
      "additionalProperties": { "$ref": "#/$defs/norma" }
    }
  },
  "$defs": {
    "norma": {
      "type": "object",
      "additionalProperties": false,
      "required": ["titulo", "rango", "materias"],
      "properties": {
        "titulo": { "type": "string", "minLength": 1 },
        "rango": { "enum": ["…vocabulario grabado…"] },
        "abreviatura": { "type": "string", "minLength": 1 },
        "materias": {
          "type": "array", "minItems": 1, "uniqueItems": true,
          "items": { "type": "string", "minLength": 1 }
        }
      }
    }
  }
}
```

- **Identificadores repetidos.** El esquema declara el identificador como nombre de propiedad de `normas`, así que el
  documento que llega al validador no puede tenerlo dos veces. En el YAML, la clave repetida la rechaza antes el lector
  común (data-model, «Lectura de documentos YAML»; research.md D8): recorre el `yaml.Node` y presenta el defecto
  nombrando la norma y sus dos líneas. Analizar el YAML a un `yaml.Node` no la detecta —la comprobación de repetidos de
  `go.yaml.in/yaml/v3` solo corre al decodificar a un mapa, a `any` o a una estructura (research.md V43)—, y sin ese
  recorrido la segunda entrada sustituiría a la primera en silencio. La alternativa de lista con un comprobador aparte
  está descartada en research.md D8.
- **`rango`**: `enum` con los valores de `rango` que devuelve la fuente para las normas grabadas (H4 + H5), escritos
  tal cual. Verificados ya en grabaciones de H4: `Ley`, `Real Decreto-ley`, `Real Decreto`. Los demás que aparezcan
  (p. ej. el de un real decreto legislativo o el de la Constitución) se leen de las búsquedas grabadas en la pausa y se
  añaden en la tarea `[datos]` del esquema; `TestEsquemaDeNormas/rangos-grabados` comprueba que el `enum` es
  exactamente el conjunto de rangos de las búsquedas grabadas.
- Es un fichero nuevo bajo `schemas/`: tarea `[datos]` con pausa (constitución, capa 3).

## 3. Validación (FR-025, FR-043)

`internal/skills.LeerNormas(contenido []byte) ([]Norma, error)`:

Con el lector común de data-model («Lectura de documentos YAML»):

1. analiza el YAML a un `yaml.Node` (conserva las líneas) y lo recorre rechazando toda clave repetida; un identificador
   repetido bajo `normas` es el defecto `repetido en las líneas N y M` de esa norma. Es este recorrido el que detecta la
   repetición: analizar a `yaml.Node` no la comprueba (research.md V43);
2. convierte el nodo a `any` con `(*yaml.Node).Decode`, lo normaliza a tipos JSON (lo codifica en JSON y lo lee con
   `jsonschema.UnmarshalJSON`, research.md V44) y lo valida contra `schemas/normas.yaml.json` con
   `santhosh-tekuri/jsonschema/v6` (`AssertFormat`);
3. cada error nombra la **norma** (el identificador de la entrada, o la clave mal formada) y el **defecto**
   (`campo no declarado: vertical`, `falta titulo`, `materias vacía`, `identificador con otra forma`, `repetido en las
   líneas N y M`).

| Test | Qué fija |
|---|---|
| `TestLeerNormas` | un caso por defecto de US6-2 y US6-3 y SC-005 (`vertical`, campo no declarado, sin título, sin rango, sin materias, identificador mal formado, e **identificador repetido**: la misma clave `BOE-A-…` dos veces bajo `normas`, con títulos distintos, falla nombrando la norma y sus dos líneas en lugar de quedarse con la segunda), cada uno nombrando norma y defecto; un documento válido devuelve las normas en el orden de §4 |
| `TestEsquemaDeNormas` | el esquema compila; `rangos-grabados` (§2) |
| `TestNormasDelRepositorio` | `data/normas.yaml` real válido (US6-1) |
| `TestGramaticasCoincidenConBoe` | el patrón de `propertyNames` acepta y rechaza lo mismo que `boe.ValidarNorma` (data-model, cabecera) |

## 4. Normas de H5

| Abreviatura | Norma (por número y año) | Identificador | Materias |
|---|---|---|---|
| LPAC | Ley 39/2015 | `BOE-A-2015-10565` (grabación de H4) | procedimiento administrativo |
| LCSP | Ley 9/2017 | `BOE-A-2017-12902` (grabación de H4) | contratación pública |
| LRBRL | Ley 7/1985 | `BOE-A-1985-5392` (grabación de H4) | régimen local |
| LGT | Ley 58/2003 | lo fija la pausa | tributos |
| TRLRHL | Real Decreto Legislativo 2/2004 | lo fija la pausa | tributos, haciendas locales |
| LIRPF | Ley 35/2006 | lo fija la pausa (`refs/boe.py` usa `BOE-A-2006-20764`, que se verifica igual) | tributos, impuesto sobre la renta de las personas físicas |
| LRJSP | Ley 40/2015 | lo fija la pausa | régimen jurídico del sector público, potestad sancionadora |
| LTAIBG | Ley 19/2013 | lo fija la pausa | transparencia, acceso a la información pública |
| CE | Constitución Española | lo fija la pausa | organización territorial, autonomía local |
| ET | Real Decreto Legislativo 2/2015 | lo fija la pausa | relaciones laborales |

## 5. `references/normas.md` (FR-030, FR-031, FR-034)

`internal/skills.RenderizarNormas([]Norma) []byte`. Salida exacta, con `\n` como fin de línea y un salto final:

```markdown
<!-- generado desde data/normas.yaml, no editar -->

# Normas de referencia

| Norma | Abreviatura | Identificador | Rango | Materias |
|---|---|---|---|---|
| Ley 7/1985, de 2 de abril, reguladora de las Bases del Régimen Local. | LRBRL | `BOE-A-1985-5392` | Ley | régimen local |
```

- Orden: año del identificador y después número, los dos numéricamente (`BOE-A-1985-5392` antes que
  `BOE-A-2015-10565`, y `BOE-A-2015-10565` antes que `BOE-A-2015-10566`).
- Materias separadas por `, `; abreviatura ausente → celda vacía.
- En cualquier celda, `|` se escribe `\|` y un salto de línea se sustituye por un espacio.
- `TestRenderizarNormas`: bytes exactos de un caso con las tres variantes de escape y orden numérico; dos llamadas dan
  lo mismo.

## 6. Verificación offline de identificadores (FR-024, SC-008)

`TestIdentificadoresDeLasNormas` (`internal/evals`):

1. lee el manifiesto de H5 (`testdata/evals/grabaciones.json`) con `LeerManifiesto` (contrato de evals §3.1); si
   devuelve un error, falla con ese error antes de reproducir ninguna búsqueda. Después, para cada entrada, reproduce
   `boe buscar <busqueda> --json` en proceso con `app.Main`, sobre el registro de `boe` compuesto con `httpx.Replay` de
   la unión de grabaciones H4 + H5 (contrato de evals §5.1), y **resuelve la entrada** como el único resultado de su
   búsqueda cuyo `titulo` empieza por su `titulo_empieza_por`, con `strings.HasPrefix`, igual que `TestGrabarEvals`
   (contrato de evals §3.2): con 0 resultados así o con más de uno, falla nombrando la entrada y los títulos. Una
   búsqueda que termina con código distinto de 0 falla nombrando la entrada y la orden, y nunca cuenta como una búsqueda
   sin resultados;
2. para cada norma de `data/normas.yaml`, exige exactamente una entrada que la resuelva, con el mismo identificador
   **y** el mismo título: si ninguna entrada resuelve ese identificador, falla con «<id>: ninguna entrada del manifiesto
   lo resuelve»; si la entrada que lo resuelve da otro título, «<id>: el título no coincide con la búsqueda grabada:
   <grabado>». Que la norma aparezca entre los resultados de la búsqueda de otra entrada, sin ser el que esa entrada
   resuelve, no cuenta: así se cumple la regla de data-model §7.2 y el resultado no depende de lo que devuelvan las
   demás búsquedas grabadas.

Subtests negativos, cada uno sobre copias temporales de `data/normas.yaml` y del manifiesto:
`identificador-cambiado` (el identificador de una norma cambiado en la copia de `data/normas.yaml`: «ninguna entrada
del manifiesto lo resuelve»); `titulo-cambiado` (su título cambiado: «el título no coincide con la búsqueda grabada»);
`norma-sin-entrada` (la entrada de una norma retirada de la copia del manifiesto: «ninguna entrada del manifiesto lo
resuelve»); y `norma-en-la-busqueda-de-otra-entrada` (se añade a la copia de `data/normas.yaml` una norma tomada de los
resultados de la búsqueda grabada por H4 `procedimiento administrativo común` que no es la Ley 39/2015, con su
identificador y su título: aunque esa búsqueda la devuelve, ninguna entrada la resuelve, y falla con «ninguna entrada
del manifiesto lo resuelve»).
