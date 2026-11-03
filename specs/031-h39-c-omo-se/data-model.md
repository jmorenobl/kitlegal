# Data model: H39

Entidades del hito, sus campos y sus reglas. Los formatos de fichero y de salida, con sus bytes, están en
`contracts/`; aquí, qué es cada cosa y de dónde sale. El binario `kitlegal` no gana ninguna entidad: solo cambia lo que
emite su esquema de entrada (§1).

## 1. Descripción de un parámetro (FR-001 a FR-003)

- **Qué es**: la clave `description` de la propiedad de un argumento o una bandera propia de un verbo, en la entrada del
  documento de `--describe` y en el `inputSchema` de su herramienta.
- **De dónde sale**: el valor de la etiqueta `help` del campo del struct de argumentos del verbo, tal cual (research V1).
- **Reglas**: un campo sin `help`, o con `help` vacía, no la lleva; las ocho banderas globales no la llevan; nada más
  del esquema cambia (ni el orden de las propiedades, ni `required`, ni `x-banderas`, ni la salida).
- **Cuántas**: 28 hoy, 22 de ellas en verbos con herramienta (las 6 de `skills` no dan herramienta).

## 2. Declaración del laboratorio (`laboratorio/variantes.yaml`)

| Campo | Tipo | Regla |
|---|---|---|
| `skill` | nombre de skill | el de la skill de laboratorio: `articulos-del-boe`; no es ninguna de `skills/` |
| `entrada` | nombre de skill | el de la skill de entrada: `consulta-legal`; no es ninguna de `skills/` |
| `herramienta` | `<applet>_<verbo>` | la de las cuatro evals y la del parámetro de `parametro`: `boe_articulo`; está en `NombresDeHerramientas` |
| `parametro` | nombre de propiedad | `bloque`; es una propiedad de la entrada de `herramienta` |
| `referencia` | nombre de fichero | `cierre.md`: el único fichero de `references/` en las variantes con referencia |
| `variantes` | lista | las ocho de FR-020 a FR-027, en este orden; cada una con `nombre`, `modos` y `cambia` |

Cada variante:

| Campo | Regla |
|---|---|
| `nombre` | uno de `control`, `instrucciones`, `entrada`, `referencia-secundaria`, `parametro`, `referencia-paso`, `referencia-en-el-punto`, `descripcion-sin-herramientas`; sin repetir |
| `modos` | subconjunto no vacío de `[orden, herramienta]`, en ese orden: `[herramienta]` en `instrucciones` y `parametro`, los dos en las demás |
| `cambia` | subconjunto de las partes de §3: vacío en `control`; `[instrucciones]`; `[entrada]`; `[cuerpo, referencia]` en las tres de referencia; `[cuerpo, parametro]`; `[descripcion]` |

El YAML se lee con el lector estricto del repositorio (`skills.ValidarDocumentoYAML`, claves desconocidas y repetidas
rechazadas); un fichero que no cumple es un defecto de `TestVariantesDelLaboratorio`, no un estado que el arnés trate.

## 3. Variante y sus partes

Una variante es el directorio `laboratorio/variantes/<nombre>/`, con `skills/<skill>/` y, según `cambia`, lo demás.
Sus **partes**, las que FR-028 compara con las de `control`:

| Parte | Qué bytes son | En `control` |
|---|---|---|
| `descripcion` | las líneas de la clave `description:` del frontmatter de `skills/<skill>/SKILL.md`: la de la clave y las sangradas que la siguen | las de su `SKILL.md` |
| `cuerpo` | el resto de ese `SKILL.md`: el frontmatter sin esas líneas y todo lo que sigue a la línea que lo cierra | el de su `SKILL.md` |
| `referencia` | los ficheros de `skills/<skill>/references/`, con su nombre y sus bytes | ninguno |
| `entrada` | el directorio `skills/<entrada>/`, con sus ficheros | no existe |
| `instrucciones` | `instrucciones.txt` | no existe: la sesión anuncia `mcp.Instrucciones` |
| `parametro` | `parametro.txt` | no existe: la sesión anuncia la ayuda de `bloque` |

Reglas (`TestVariantesDelLaboratorio`): una parte no declarada es idéntica byte a byte a la de control; una declarada
difiere; en la variante no hay ningún fichero que no sea de una parte. Y la regla testigo (§5) está donde la variante
la pone: en `cuerpo` en `control`, `instrucciones`, `entrada` y `descripcion-sin-herramientas`; en
`references/cierre.md`, y no en `cuerpo`, en las tres de referencia; en `parametro.txt`, y no en `cuerpo`, en
`parametro`. La `descripcion` de `control` nombra alguna herramienta de `NombresDeHerramientas` y la de
`descripcion-sin-herramientas`, ninguna (ni `kitlegal <applet>`); `instrucciones.txt` nombra la skill; el cuerpo de la
skill de entrada nombra la skill de laboratorio.

## 4. Eval de laboratorio (`laboratorio/evals/*.yaml`)

Cuatro, con el formato de `evals/` (`schemas/eval.yaml.json`): `pregunta`, `activa: true`, `comandos` (un comando
`applet: boe` con `norma` y `bloque`, que es una lectura de `boe_articulo`) y `citas`. Son las preguntas y las lecturas
de las evals 09, 04, 07 y 10 de `normativa` (research V19):

| Fichero | Norma | Bloque |
|---|---|---|
| `01-constitucion-articulo-140.yaml` | BOE-A-1978-31229 | a140 |
| `02-lgt-prescripcion.yaml` | BOE-A-2003-23186 | a66 |
| `03-lrjsp-principio-de-legalidad.yaml` | BOE-A-2015-10566 | a25 |
| `04-et-vacaciones.yaml` | BOE-A-2015-11430 | a38 |

Regla (`TestEvalsDelLaboratorio`): bien formadas (`LeerConjunto`), cuatro, cada comando una lectura de la herramienta
declarada, y todas sus consultas servidas sin red por `UnionDeGrabaciones()` (`Preparar` sin faltas y
`ComprobarSinRed` sin fallos).

## 5. Regla testigo

- **Línea**: `Sello de lectura: KL-7Q4`.
- **Cumple** una respuesta si su última línea no vacía, sin espacios a los lados, es exactamente esa línea.
- **Formas de no cumplir** que la regla distingue: la línea no aparece; aparece y no es la última no vacía; la última
  lleva el sello con formato u otro texto.
- **Veredicto**: `cumple` o `no cumple`; una palabra por sesión.

## 6. Sondeo y sesión del sondeo

- **Sondeo**: un directorio `DIRECTORIO` que crea el arnés, con `sondeo.json` (`modelo`, `claude_code`) y una sesión
  por cada `<variante>/<modo>/<eval sin .yaml>-<nn>/` (nn desde 01, con dos cifras).
- **Sesión**: el directorio que deja el reparto de hoy (`sesion.jsonl`, `sesion.err`, `codigo-de-la-sesion`,
  `servidor.json` en el modo herramienta, `claude/`, `trabajo/`…). Su variante, su modo y su eval son los de su ruta
  (FR-033); su modelo y su versión, los de `sondeo.json`.
- **Lo que el guion saca de una sesión**: si terminó con respuesta (`LeerSesion`: `Terminada` y `Respuesta` no vacía);
  y del transcript, el turno de cada `tool_use` (research D10): de la activación de la skill y de la de entrada, de la
  primera herramienta de kitlegal y de la primera apertura de la referencia; y el veredicto de la regla testigo sobre la
  respuesta.
- **Estados**: terminada con respuesta (entra en `n`); sin terminar o sin respuesta (columna propia, fuera de `n`); un
  directorio que el guion no sabe leer (defecto `inesperado`, sale con 1).

## 7. Fila de la tabla

Clave: variante, modo, modelo y versión de Claude Code. Campos: sesiones sin terminar; `n`; activan antes;
y, en `entrada`, activan la de entrada antes y activan las dos antes; y, en las tres de referencia, la abren y el turno
de su primera apertura (mediana, mínimo y máximo); y cumplen la regla. Las columnas que no son de la variante, vacías;
con `n` = 0, recuentos, porcentajes e intervalos vacíos.
