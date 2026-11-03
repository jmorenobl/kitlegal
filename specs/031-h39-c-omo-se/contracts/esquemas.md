# Contrato: la ayuda de cada argumento en `--describe` y en las herramientas MCP

FR-001 a FR-005; SC-001, SC-002. Research V1 a V8, D1 a D3.

## 1. Forma

En la entrada de cada documento de `kitlegal <applet> <verbo> … --describe` y en el `inputSchema` de cada herramienta
`<applet>_<verbo>` de `kitlegal mcp serve`, la propiedad de cada argumento o bandera **propia del verbo** cuya etiqueta
`help` no está vacía lleva `"description"` con ese texto, carácter a carácter una vez decodificado el JSON; en el
documento de `--describe`, además, sin escapar `<`, `>` ni `&` (research V4); en el cable de MCP el SDK puede
escribirlos como `<`… al serializar el `inputSchema` (`internal/mcp/servir.go:107`), que se decodifica igual. La
clave va detrás de las que ya lleva la propiedad (`type`, `items`, `$ref`…), en el orden de campos del esquema
de invopop (`internal/cli`, research V3). Nada más cambia: ni las ocho banderas globales, que no la llevan; ni el orden
de las propiedades, `required`, `additionalProperties`, `x-banderas` ni la salida. `description` de la raíz del
documento sigue siendo la ayuda del verbo.

Para describir un verbo hay que darle sus argumentos obligatorios, porque el análisis va antes (research V5).

## 2. Ejemplo con sus bytes

`kitlegal boe articulo BOE-A-2015-10565 a21 --describe`, la entrada (hoy, en las mismas líneas, solo `"type"`):

```json
    "entrada": {
      "properties": {
        "norma": {
          "type": "string",
          "description": "Identificador de la norma en el BOE: BOE-A-<año>-<número>."
        },
        "bloque": {
          "type": "string",
          "description": "Id del bloque, del campo id de kitlegal boe indice (a21, da-3, a1-30…): no se compone con el número del artículo, que no siempre es su id."
        },
        "fecha": {
          "type": "string",
          "description": "Día cuya redacción del bloque se quiere, la que regía ese día, en lugar de la vigente."
        },
        "json": {
          "type": "boolean"
        },
```

En `boe_articulo`, el `inputSchema` (JSON compacto) lleva las mismas tres `description` en `norma`, `bloque` y `fecha`.
En `skills install --describe`, `global` lleva «Actúa en el ámbito global: $HOME/.agents/skills y los hosts de HOME.»,
con `$HOME` tal cual (research V1).

## 3. Uso, de fuera adentro

- **Quién la lee**: el agente en modo herramienta, al cargar el esquema de una herramienta (`ToolSearch`), una vez por
  herramienta y sesión; con `normativa`, dos por pregunta (`boe_indice` y `boe_articulo`). En modo orden, quien pide
  `--describe`, que ninguna skill pide.
- **Tamaño**: texto fijo del binario. `--describe` de `boe articulo` pasa de 4989 bytes a unos 5,4 KB (tres líneas de
  29 bytes más la ayuda). Las 22 ayudas de verbos con herramienta caben en las 2373 de las 28 (research V2): lo que
  añade la lista entera de herramientas está acotado por ellas más 17 bytes de clave por parámetro, y lo que añaden las
  dos herramientas que carga una pregunta de `normativa`, por unos 400 bytes. No crece con cientos de normas ni miles de
  bloques consultados: cambia solo cuando cambia la ayuda de un argumento.
- **Señal**: ninguna; no avisa de nada que caduque.

## 4. Lo que se regenera y lo que no cambia

- `schemas/{norma,bloque,municipio,instalacion,grafo,cita,plazo}.json` se regeneran con
  `go test ./internal/app/ -run '^TestEsquemasPublicados$' -args -actualizar-esquemas`; `schemas/servidor.json` no
  cambia (`mcp serve` no tiene argumentos). `make schema-check` en verde (FR-004).
- `internal/app/testdata/plazos/describe.json` (con `-actualizar-golden-de-plazos`) y la sección
  `main-articulos.describe` de `internal/app/testdata/script/h20-fecha-main.txtar` se rehacen con la salida del binario.
- `skills/` no cambia: la tabla de comandos no lee la `description` de las propiedades (research V7), y la deriva de
  `TestSkillsDelRepositorio` falla si el generador la cambiara (FR-004, SC-002).
- `CHANGELOG.md`, *Unreleased*, sección *Changed*: «Los parámetros de `--describe` y de las herramientas del servidor
  MCP llevan como `description` la ayuda de su argumento» (FR-005).

## 5. Controles

- `TestAyudaComoDescripcion` (`internal/cli`): con structs de prueba, una propiedad con ayuda, una sin ella y las
  globales, en los dos esquemas (FR-001, FR-003).
- `TestAyudasDeLosArgumentos` (`internal/app`): todos los verbos de `RegistroDeProduccion`, contra la ayuda del modelo
  de Kong de cada verbo (research D2), en `--describe` y en `tools/list` (FR-001, FR-002, FR-003, SC-001).
- `TestEsquemasPublicados` (`make schema-check`), `TestHerramientasDelServidor` y `TestSkillsDelRepositorio`, como hoy.
