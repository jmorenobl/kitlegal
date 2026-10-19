# Contrato: `kitlegal boe articulo … --fecha` y `boe_articulo` con `fecha`

Lo que el binario promete con la lectura fechada. La lectura sin fecha no cambia en nada (FR-012): ni su código, ni su
caché, ni lo que entrega al grafo. Las cifras de bytes son las medidas en esta sesión con el prototipo (research M1 a
M3), sobre lo grabado de H4.

## 1. Invocación

```text
kitlegal boe articulo <norma> <bloque> [--fecha AAAA-MM-DD] [banderas globales]
```

- **Orden**: `--fecha` va en cualquier sitio tras el verbo, con su valor detrás o con `=`.
- **Herramienta**: `boe_articulo` con `norma`, `bloque` y el parámetro opcional `fecha` (string). Lo genera el
  registro desde el verbo, sin código del servidor (research V14); la llamada es la orden
  `boe articulo --fecha=<fecha> -- <norma> <bloque> --json …` y devuelve su mismo sobre (FR-001, FR-084).
- **`--describe`**: la entrada gana la propiedad `fecha` (string, con la ayuda de la bandera) y `x-banderas: ["fecha"]`;
  la salida, en la rama de `ok: true`, apunta a `boe.ArticuloFechado`, que lleva las doce propiedades de hoy, obligatorias
  como hoy, y `fecha_pedida` y `redaccion_posterior`, opcionales, más la definición `boe.RedaccionPosterior`. El
  `--describe` de `articulos` y los de los demás verbos no cambian (FR-012, FR-070, FR-085). El golden es
  `schemas/bloque.json`, regenerado.
- **Tabla de comandos de la skill** (regenerada con `make skills-sync`): la fila de `boe articulo` pasa a
  `kitlegal boe articulo <norma> <bloque> [--fecha <fecha>]` y a «objeto con … `url_eli`, `fecha_pedida`,
  `redaccion_posterior`»; la de `boe articulos`, igual.

Orden de validación, antes de abrir la caché o pedir nada (nota [28]): norma, bloque, fecha. Con las tres mal, el 2 de
la norma.

## 2. La versión de una fecha

Con un bloque cuyas versiones tienen todas `fecha_vigencia` (situable), la elegida es la última, en el orden del
bloque, cuya fecha de vigencia no es posterior a la pedida; la redacción posterior, la última del bloque si no es la
elegida (FR-002, FR-010, FR-013). Comparación de días del calendario; `fecha_publicacion` no interviene.

**Golden de FR-081** (`TestArticuloEnUnaFecha`, valores escritos a mano en el test; huellas medidas en research M1):

| Caso | Bloque | `--fecha` | Código | `fecha_version` | `fecha_vigencia` | `norma_modificadora` | `hash_texto` | `redaccion_posterior` |
|---|---|---|---|---|---|---|---|---|
| anterior a la primera vigencia | `BOE-A-1985-5392` `a22` | `1985-04-22` | 3 (mensaje con `19850423`) | — | — | — | — | — |
| intermedia | ídem | `2005-06-15` | 0 | `20031217` | `20040101` | `BOE-A-2003-23103` | `sha256:ce4210549efa0b8a103e99ccfd4137f6b66a5f300f661a94e2d2545b08847183` | `20151031`, `BOE-A-2015-11723` |
| el día que entra en vigor | ídem | `2008-04-30` | 0 | `20071031` | `20080430` | `BOE-A-2007-18874` | `sha256:4c8982b1fee3d8414a0b3fdf7d71e4b269b5a1ec43197edeae71a171e1d151fe` | `20151031`, `BOE-A-2015-11723` |
| la víspera, ya publicada la de ese día | ídem | `2008-04-29` | 0 | `20070529` | `20070701` | `BOE-A-2007-10701` | `sha256:50a1cc5d30109aeb4747a003df8b6855ab23631c7f3fcf81b6e068f7a571343e` | `20151031`, `BOE-A-2015-11723` |
| posterior a la última | ídem | `2019-06-01` | 0 | `20151031` | `20151031` | `BOE-A-2015-11723` | el de `testdata/golden/articulo-BOE-A-1985-5392-a22.json`, con su mismo `texto` | ninguna |
| entre publicación y vigencia | `BOE-A-2015-10565` `a21` | `2016-06-01` | 3 (mensaje con `20161002`) | — | — | — | — | — |
| US1 | `BOE-A-2017-12902` `a1-30` | `2019-06-01` | 0 | `20171109` | `20180309` | `BOE-A-2017-12902` | `sha256:a461184434a16f37601a3e6b53e89042ce31682e45a9dd35fecd97733d55b8e5` | `20200206`, `BOE-A-2020-1651` |

En cada caso con 0, `hash_texto` es la huella de su `texto`, `fecha_pedida` la fecha pedida, y `titulo`, `tipo`,
`avisos`, `url` y `url_eli` los de la lectura sin fecha del mismo bloque.

## 3. Errores y mensajes

| Caso | Código y clase | `data.mensaje` |
|---|---|---|
| `--fecha` que no es un día AAAA-MM-DD (`2019-02-30`, `2019-6-1`, vacía…) | 2 `argumentos` | `la fecha "<valor>" no es un día AAAA-MM-DD del calendario, como 2019-06-01` |
| Fecha anterior a la primera vigencia (FR-003) | 3 `no-encontrado` | `el bloque <bloque> de <norma> no tenía redacción vigente el <AAAA-MM-DD>: su primera redacción tiene fecha de vigencia <AAAAMMDD> (<dirección del bloque>)` |
| Bloque sin versiones, o con alguna sin `fecha_vigencia` (FR-004) | 3 `no-encontrado` | `el BOE no da la fecha de vigencia de cada redacción del bloque <bloque> de <norma>: no se puede saber cuál tenía el <AAAA-MM-DD> (<dirección del bloque>)` |
| Bloque que no está en la norma | 3, el de hoy | `la norma <norma> no tiene el bloque <bloque> (<dirección>)`, como hoy |
| Norma o bloque mal formados, fuente caída o con límite, bloque ilegible, metadatos caídos | los de hoy | los de hoy, sin escribir la entrada de las versiones (FR-007) |
| `--offline` sin la entrada de las versiones | 4 `fuente-no-disponible` | el de hoy de una entrada ausente con `--offline`, con el verbo `versiones` y su clave (FR-020) |

Ejemplo, FR-003 (US2):

```text
el bloque a21 de BOE-A-2015-10565 no tenía redacción vigente el 2016-06-01: su primera redacción tiene fecha de vigencia 20161002 (https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21)
```

El 3 de FR-003 y el de FR-004 llevan la dirección del bloque y la fecha de consulta de las versiones; ningún fallo de
la elección devuelve texto.

## 4. La salida

`data` de una lectura fechada con 0 es `boe.ArticuloFechado` (data-model §4): los doce campos de hoy de la versión
elegida, `fecha_pedida` y, si la hay, `redaccion_posterior`. Ejemplo, US1, con el texto abreviado:

```json
{"norma":"BOE-A-2017-12902","bloque":"a1-30","titulo":"Artículo 118","tipo":"precepto","fecha_version":"20171109","fecha_vigencia":"20180309","norma_modificadora":"BOE-A-2017-12902","texto":"…","hash_texto":"sha256:a461184434a16f37601a3e6b53e89042ce31682e45a9dd35fecd97733d55b8e5","avisos":[],"url":"https://www.boe.es/buscar/act.php?id=BOE-A-2017-12902#a1-30","url_eli":"https://www.boe.es/eli/es/l/2017/11/08/9","fecha_pedida":"2019-06-01","redaccion_posterior":{"fecha_vigencia":"20200206","norma_modificadora":"BOE-A-2020-1651"}}
```

- **Bytes**: con el texto entero, 2 226 bytes de `data` (1 688 del texto); los dos campos nuevos, 119 bytes (research
  M2); sin redacción posterior, `fecha_pedida` sola, 28 bytes. El sobre añade unos 290 bytes (clave, URL del bloque,
  fecha de consulta y huella): unos 2,5 KB en total. El mayor de los bloques que leen las evals, el `a17` de
  `BOE-A-2006-20764`, sigue acotado por el texto de una sola versión.
- **Salida legible** (sin `--json`): la tabla mínima de hoy con dos filas más, `fecha_pedida` y, si la hay,
  `redaccion_posterior.fecha_vigencia` y `redaccion_posterior.norma_modificadora` (research V39).
- **Avisos**: los de la norma hoy, los mismos de la lectura sin fecha; una redacción antigua de una norma hoy derogada
  sigue avisando (FR-011).
- **Validación**: toda salida de la lectura fechada valida contra `schemas/bloque.json` (FR-014).

## 5. La caché

| Lectura | Lee | Escribe |
|---|---|---|
| Sin `--fecha` | la entrada de hoy (`articulo`) | la entrada de hoy, como hoy |
| Con `--fecha` | la entrada de las versiones (`versiones`) | la entrada de las versiones, antes de elegir y solo con el bloque y sus metadatos obtenidos e interpretados; con un fallo de la fuente o al interpretar, nada, como hoy (FR-007) |

- Clave de ejemplo:
  `boe.legislacion-consolidada|1|versiones|https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2017-12902/texto/bloque/a1-30`.
- Datos (data-model §5): 3 843 bytes en el `a1-30`, 37 561 en el `a22` y 3 856 en el `a21` (research M3).
- Con `--offline` y la entrada vigente, la lectura de cualquier fecha del bloque da el mismo `data`, el mismo `hash` y
  la misma fecha de consulta que la lectura que la guardó (FR-022); sin ella, 4 (FR-020).
- Tras una lectura fechada, la entrada de hoy del bloque no cambia en ningún byte, y la lectura sin fecha con
  `--offline` da lo mismo que antes (FR-021).
- `--dry-run`: las líneas de las peticiones que haría, como la lectura sin fecha; sin escribir nada.

## 6. El grafo

Una lectura fechada con 0, sin `--no-graph` ni `--dry-run` y con ELI, emite en su `Resultado`:

```go
schema.Observado{
	Vigencia:    vigenciaLarga,
	SoloLoNuevo: true,
	Operaciones: []schema.Operacion{
		schema.Nodo{ID: "<eli>#<bloque>@<fecha_vigencia>:<hash_texto>", Tipo: grafo.TipoBloqueVersion, Datos: /* fecha_vigencia, fecha_version, norma_modificadora, hash_texto */},
		schema.Texto{Huella: "<hash_texto>", Cuerpo: "<texto>"},
	},
}
```

Ejemplo de id, la intermedia del `a22`:
`eli/es/l/1985/04/02/7#a22@20040101:sha256:ce4210549efa0b8a103e99ccfd4137f6b66a5f300f661a94e2d2545b08847183`.

- La entrega le pone la procedencia del sobre como `source` (FR-030). Con `SoloLoNuevo`, un nodo o un texto que
  `world.db` ya guarda queda como estaba; uno nuevo entra con esa procedencia. Entregar dos veces el mismo resultado
  no cambia nada la segunda (FR-030).
- Sin `eli:has_version`, no es una lectura: la fila de lecturas del bloque no cambia; `graph check` sobre el bloque
  da lo mismo antes y después, y la siguiente lectura sin fecha da el `graph check` que daría sin ella, también en un
  bloque sin fila de lecturas (FR-031).
- Con `--no-graph` o `--dry-run`, nada (FR-032).

## 7. Uso, de fuera adentro

| Salida | Quién la pide y cuántas veces | Tamaño con meses de uso | Cuándo deja de darse cada señal |
|---|---|---|---|
| El sobre de la lectura fechada | `boe-legislacion`, cuando la pregunta lleva una fecha: una vez por bloque que necesita la respuesta, dos con un año o un mes sin día (FR-045); y, solo si la respuesta dice qué dice hoy, una lectura sin fecha más (FR-042) | el de una versión: unos 2,5 KB en el `a1-30`; no lee nada de lo acumulado, así que no crece con cientos de normas ni miles de bloques consultados | `redaccion_posterior` solo se da si la elegida no es la última: deja de darse con una fecha desde la última vigencia, y cambia cuando el BOE añade una versión al bloque |
| El 3 de FR-003 | `boe-legislacion`, una vez por lectura que lo da, para su línea `⚠ SIN REDACCIÓN VIGENTE:` | un mensaje de unos 200 bytes | solo con una fecha anterior a la primera vigencia del bloque |
| La entrada de las versiones | solo el binario, por su clave, para la lectura fechada y su `--offline`; ninguna skill la lee | una por bloque leído con fecha, como mucho la respuesta del BOE a ese bloque (37,6 KB el `a22`); sin purga, como las de hoy (research V8): con mil bloques leídos con fecha, del orden de decenas de MB | no da señales |
| El `BloqueVersion` y su texto | ninguna skill: no es una lectura (FR-031), así que `graph check` no lo nombra | un nodo y un texto por versión distinta leída con fecha, como mucho las versiones de los bloques leídos así (nueve en el `a22`) | no da señales: no es redacción vista de ningún bloque, así que ni `version-obsoleta` ni `fuente-caducada` |
