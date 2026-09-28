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
- Go: `Eval` gana `GrafoPrevio GrafoPrevio` (`yaml:"grafo_previo"`, delante de `Comandos`) y
  `Prohibidos []ComandoProhibido` (`yaml:"prohibidos"`, detrás de `Comandos`); `ComandoProhibido{Applet, Verbo}` y
  `GrafoPrevio{Grabaciones string; Comandos []ComandoEsperado}`, los dos con sus etiquetas `yaml`; `formaDelComando`
  gana `formaComprobacion` para `verbo: check`, cuyo texto es `<applet> check` y que no genera ninguna consulta que
  preparar.

## 2. Juicio (`internal/evals/juzgar.go`, `informe.go`)

- Un comando de la forma comprobación lo satisface una invocación del mismo applet con verbo `check` que consulta y
  termina con 0.
- Un comando prohibido lo ejecuta **toda invocación que consulta** (ni ayuda, ni `--describe`, ni `--dry-run`) del
  mismo applet y verbo, termine como termine, también sin código (la que dejó el corte de la sesión). Cada prohibido
  de la eval que ejecuta alguna invocación va, una vez, en el orden de la eval y con sus repeticiones, a
  `comandos_prohibidos_ejecutados` (nueva clave de cada eval de `informe.json`, detrás de `comandos_ausentes`; lista
  vacía cuando no hay) y deja un motivo `comando prohibido ejecutado: <applet> <verbo>`, detrás de los de los
  comandos ausentes y delante de los de las citas; la eval no pasa. El informe legible publica la columna «Comandos
  prohibidos ejecutados», detrás de «Comandos ausentes» («ninguno» si no hay). Sin `prohibidos`, el juicio es el de
  antes.
- Las citas se siguen comparando por identificador.

## 3. Preparación de la sesión (`internal/evals/preparar.go`)

Si la eval de la sesión lleva `grafo_previo`, **antes** de llenar la caché de la sesión (`prepararGrafoPrevio`, sin
exportar):

1. copia en un temporal de reproducción los conjuntos de `SesionAPreparar.Grabaciones` (en el job,
   `UnionDeGrabaciones()`), en su orden, y, encima, `<SesionAPreparar.GrafosPrevios>/<grabaciones>/`, cuyo valor cero
   es la constante `GrafosPrevios` (`../../testdata/evals/grafo-previo`, relativa a `internal/evals`);
2. monta un registro con `boe` sobre `httpx.Replay` de esa reproducción, una caché **temporal** que se descarta y
   `EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(<caché de la sesión>)))`;
3. ejecuta cada comando, en su orden, como `kitlegal <applet> articulo <norma> <bloque> --json` con `app.Main`; un
   código distinto de 0, o **cualquier** salida de error (la línea de una entrega fallida), es una `Falta` con el
   punto `PuntoGrafoPrevio` («grafo previo»), que `Falta.String` nombra «el comando del grafo previo <applet>
   articulo <norma> <bloque>».

Con alguna falta o un error, `PrepararSesion` los devuelve tal cual, sin preparar la caché ni escribir `eval.txt`,
`modelo.txt` ni `pregunta.txt`. Si no, prepara la caché como hoy, con un registro **sin** entrega al grafo. Así el grafo de la sesión solo tiene la
versión anterior y la caché sirve la grabada; solo el grafo previo de la eval de la sesión se prepara, no el de las
demás. `TestEvalsDelRepositorio` gana la subprueba `grafo-previo`: en toda eval de cada carpeta de `evals/` que lleva
`grafo_previo` (y exige que alguna lo lleve), cada conjunto nombrado existe y su preparación, en temporales, no da
ninguna falta y deja en el grafo, para cada comando, la `Norma`, el `Bloque` y un `BloqueVersion` unidos por
`eli:has_part` y `eli:has_version`.

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
- Protocolo (FR-080), con una sección nueva «Memoria de consultas»: resuelto el `BOE-A-…` de la norma de la
  pregunta (paso 2) —no el de cada norma remitida— y antes de leer ninguno de sus bloques,
  `kitlegal graph check --json`; cuando ya no queda nada por leer (el paso 4 puede volver al 3), al empezar el paso 5
  y antes de redactar la respuesta, otra vez. En la respuesta, los hallazgos de las dos comprobaciones cuya
  `explicacion` nombra el `BOE-A-…` de la norma de la pregunta, solo (el de la `Norma`) o en la cita de un bloque de
  esa norma leído para responder (los de su `Bloque` y sus `BloqueVersion`), agrupados por `clase` y dichos **una
  vez** cada clase: con `version-obsoleta`, que la redacción ha cambiado respecto de la consultada antes, con las
  fechas de vigencia (`fecha_vigencia` y `fecha_vigencia_reciente`); con `fuente-caducada`, que la consulta anterior
  había caducado y la respuesta se apoya en la lectura nueva. Los demás hallazgos —de otras normas, también las
  leídas por una remisión, y de bloques no leídos para responder— no se trasladan; un hallazgo no lleva la forma fija
  de los avisos.
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
| Forma comprobación (sin consultas que preparar), prohibidos (con y sin código 0; ayuda y `--describe` no cuentan), informe con la clave nueva | `TestFormaDelComando` (`formato_test.go`), `TestConsultasNecesarias`, `TestJuzgar/…`, `TestInforme` |
| Preparación del grafo previo (grafo con la anterior, caché con la grabada, falta ante stderr) | `TestPrepararGrafoPrevio` (`internal/evals/preparar_test.go`) |
| Conjunto de `boe-legislacion` con la eval nueva informativa; conjuntos existentes intactos | `TestEvalsDelRepositorio` (`conjunto`, `conjunto-legal-core`, `grafo-previo`) |
| `SKILL.md`: frontmatter, < 300 líneas, tabla sin deriva con `kitlegal graph check`, cada orden empotrada registrada | `TestSkillsDelRepositorio`, `TestOrdenesDeLasSkillsEmpotradas`, `make skills-check` |
| Protocolo (FR-080 a FR-082) | jueces de la revisión final (capa 2) y la eval en el job (SC-012) |
