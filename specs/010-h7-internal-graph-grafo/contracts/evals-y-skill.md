# Contrato: formato de eval ampliado, la eval de la consulta repetida y `boe-legislacion` v0.1

FR-080 a FR-087. Decisiones: [../research.md](../research.md) D22, D25-D27.

## 1. Formato común de eval (`schemas/eval.yaml.json`, `internal/evals/formato.go`)

Tres añadidos, opcionales y compatibles hacia atrás: toda eval existente de `boe-legislacion` y `legal-core` se lee y
se juzga igual (FR-086).

```json
"comandos": { "items": { "oneOf": [ …las cuatro de hoy…, { "$ref": "#/$defs/comando-comprobacion" } ] } },
"prohibidos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/comando-prohibido" } },
"grafo_previo": {
  "type": "object", "additionalProperties": false, "required": ["grabaciones", "comandos"],
  "properties": {
    "grabaciones": { "type": "string", "pattern": "^[a-z0-9]+(-[a-z0-9]+)*$" },
    "comandos": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/comando-bloque" } }
  }
},
"$defs": {
  "comando-comprobacion": { "type": "object", "additionalProperties": false, "required": ["applet", "verbo"],
    "properties": { "applet": { "$ref": "#/$defs/applet" }, "verbo": { "const": "check" } } },
  "comando-prohibido": { "type": "object", "additionalProperties": false, "required": ["applet", "verbo"],
    "properties": { "applet": { "$ref": "#/$defs/applet" }, "verbo": { "type": "string", "pattern": "^[a-z]+$" } } }
}
```

- `prohibidos` y `grafo_previo` solo con `activa: true`: el `else` del esquema los añade a su lista de claves
  prohibidas.
- `avisos` no cambia: su enumerado sigue siendo `boe.CodigosDeAviso()`; los hallazgos de `graph check` no son avisos.
- Go: `Eval` gana `Prohibidos []ComandoProhibido` (`yaml:"prohibidos"`) y `GrafoPrevio GrafoPrevio`
  (`yaml:"grafo_previo"`); `formaDelComando` gana `formaComprobacion` para `verbo: check`, cuyo texto es
  `<applet> check` y que no genera ninguna consulta que preparar.

## 2. Juicio (`internal/evals/juzgar.go`, `informe.go`)

- Un comando de la forma comprobación lo satisface una invocación del mismo applet con verbo `check` que consulta y
  termina con 0.
- Un comando prohibido lo ejecuta **toda invocación que consulta** (ni ayuda, ni `--describe`, ni `--dry-run`) del
  mismo applet y verbo, termine como termine. Cada uno ejecutado va a `comandos_prohibidos_ejecutados` (nueva clave de
  cada eval de `informe.json`, lista vacía cuando no hay) y deja un motivo `comando prohibido ejecutado: <applet>
  <verbo>`; la eval no pasa. El informe legible publica la columna. Sin `prohibidos`, el juicio es el de antes.
- Las citas se siguen comparando por identificador.

## 3. Preparación de la sesión (`internal/evals/preparar.go`)

Si la eval de la sesión lleva `grafo_previo`, **antes** de llenar la caché de la sesión:

1. copia las grabaciones de `UnionDeGrabaciones()` y, encima, `testdata/evals/grafo-previo/<grabaciones>/` en un
   temporal de reproducción;
2. monta un registro con `boe` sobre esa reproducción, una caché **temporal** que se descarta y
   `EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(<caché de la sesión>)))`;
3. ejecuta cada comando como `kitlegal boe articulo <norma> <bloque> --json` con `app.Main`; un código distinto de 0,
   o **cualquier** salida de error (la línea de una entrega fallida), es una `Falta`.

Después prepara la caché como hoy, con un registro **sin** entrega al grafo. Así el grafo de la sesión solo tiene la
versión anterior y la caché sirve la grabada; solo el grafo previo de la eval de la sesión se prepara, no el de las
demás. `TestEvalsDelRepositorio` gana la subprueba `grafo-previo`: cada conjunto nombrado existe y su preparación, en
temporales, no da ninguna falta y deja en el grafo un `BloqueVersion` por comando.

## 4. La eval (`evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml`)

```yaml
# Eval informativa (FR-085, FR-087): el grafo de la sesión ya registró una versión anterior del art. 21 LPAC
# (derivada de la grabación, con fecha de vigencia 20151002) y la caché sirve la grabada (20161002). La sesión tiene
# que leer el bloque con kitlegal boe articulo, comprobar la memoria con kitlegal graph check, no pedir graph show y
# citar el bloque. Que la respuesta diga que la redacción cambió no se comprueba mecánicamente (clarificación Q2).
pregunta: "Ya te pregunté hace tiempo por el artículo 21 de la Ley 39/2015. ¿Qué dice ahora?"
activa: true
informativa: true
grafo_previo:
  grabaciones: lpac-a21-version-anterior
  comandos:
    - applet: boe
      norma: BOE-A-2015-10565
      bloque: a21
comandos:
  - applet: boe
    norma: BOE-A-2015-10565
    bloque: a21
  - applet: graph
    verbo: check
prohibidos:
  - applet: graph
    verbo: show
citas:
  - norma: BOE-A-2015-10565
    bloque: a21
```

Informativa (ADR 0016): se ejecuta con las repeticiones del job y publica su tasa sin decidir; las reglas del conjunto
(10 positivas que deciden, materias distintas) no cambian (FR-087). La derivada
`testdata/evals/grafo-previo/lpac-a21-version-anterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json`
entra por una tarea `[datos]` (research D22).

## 5. `skills/boe-legislacion/SKILL.md` v0.1

- Frontmatter: `kitlegal-applets: boe graph`; la región generada gana la tabla de `kitlegal graph` (`show`, `stats`,
  `check`) con `make skills-sync` (FR-083).
- Protocolo (FR-080): tras resolver el `BOE-A-…` (paso 2) y antes de leer, `kitlegal graph check --json`; tras leer
  (paso 3) y antes de responder, otra vez. En la respuesta, los hallazgos cuya explicación nombra el `BOE-A-…` de la
  norma de la pregunta, agrupados por clase y dichos **una vez** cada clase: con `version-obsoleta`, que la redacción
  ha cambiado respecto de la consultada antes, con las fechas de vigencia; con `fuente-caducada`, que la consulta
  anterior había caducado y la respuesta se apoya en la lectura nueva. Los demás hallazgos no se trasladan.
- Reglas (FR-081, FR-082): el texto citado sale siempre de `kitlegal boe articulo` o `articulos`; nunca se cita,
  parafrasea ni reconstruye texto a partir de la salida de un verbo de `graph`; `graph check` con 0 es un resultado,
  con hallazgos o sin ellos, nunca un fallo de la herramienta; con otro código, se responde igual con el texto de
  `kitlegal boe` y se dice que no se ha podido comprobar la memoria de consultas.
- Nada más cambia: ni la forma de la cita, ni la de los avisos, ni el resto del protocolo; menos de 300 líneas; sin
  nombrar evals, el job ni modelos (FR-083).

## 6. Qué lo vigila

| Control | Test |
|---|---|
| El esquema admite las formas nuevas y rechaza sus variantes mal formadas y en evals de no activación; las existentes validan igual | `TestLeerEval`, `TestEsquemaDeEval` (`internal/evals/formato_test.go`) |
| Forma comprobación, prohibidos (con y sin código 0; ayuda y `--describe` no cuentan), informe con la clave nueva | `TestFormaDelComando`, `TestJuzgar/…`, `TestInforme` |
| Preparación del grafo previo (grafo con la anterior, caché con la grabada, falta ante stderr) | `TestPrepararGrafoPrevio` (`internal/evals/preparar_test.go`) |
| Conjunto de `boe-legislacion` con la eval nueva informativa; conjuntos existentes intactos | `TestEvalsDelRepositorio` (`conjunto-…`, `grafo-previo`) |
| `SKILL.md`: frontmatter, < 300 líneas, tabla sin deriva con `kitlegal graph check`, cada orden empotrada registrada | `TestSkillsDelRepositorio`, `TestOrdenesDeLasSkillsEmpotradas`, `make skills-check` |
| Protocolo (FR-080 a FR-082) | jueces de la revisión final (capa 2) y la eval en el job (SC-012) |
