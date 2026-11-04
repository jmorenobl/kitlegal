# Data model: H40 · Por qué `normativa` no lee sus referencias en modo herramienta

Lo que el hito añade o cambia, desde lo que lo consume. Ninguna entidad es del producto: ni el binario que se distribuye
ni ninguna skill las lee (spec, «Uso, de fuera adentro»).

## 1. Familia del laboratorio

Lo que el arnés y la tabla necesitan saber de un conjunto de variantes. Hay dos, y una invocación es de una sola (FR-010):

| Campo | H39 (sin cambios) | `normativa` (nueva) |
|---|---|---|
| Declaración | `laboratorio/variantes.yaml`, leída con su esquema | en Go, `internal/evals/laboratorio_normativa.go` (research D3) |
| Skill de cada variante | `articulos-del-boe` (y `consulta-legal` en `entrada`) | `normativa` |
| Directorio de las variantes | `laboratorio/variantes/<variante>/` | `laboratorio/normativa/<variante>/` |
| Variantes y modos | las ocho de `variantes.yaml` | `normativa-control` [orden, herramienta]; `normativa-remisiones-en-orden`, `normativa-cuerpo-corto`, `normativa-testigo-en-referencia` [herramienta] |
| Directorio de evals del reparto | `laboratorio/evals` | `evals/normativa` |
| Evals del sondeo | sus cuatro ficheros | `04-lgt-prescripcion.yaml`, `05-trlrhl-impuestos-municipales.yaml`, `15-irpf-rendimientos-por-materia.yaml`, `22-lcsp-contrato-menor-en-2019.yaml`, en ese orden |
| Referencias que cuenta la tabla | la de `referencia` en las variantes que cambian la referencia | las ocho de `skills/normativa/references/`, en orden de nombre, en todas |
| Regla testigo en la tabla | en todas las variantes | solo en `normativa-testigo-en-referencia` |
| Servidor del modo herramienta | `laboratorio servidor` en `instrucciones` y `parametro` | `kitlegal mcp serve` en todas (FR-008) |

Reglas: ningún nombre de variante es de las dos familias (lo comprueba `TestVariantesDeNormativa`); las dos tienen cuatro
evals, así que `SESIONES` es múltiplo de 4 en las dos; el nombre de una variante es el de su directorio.

## 2. Variante de `normativa`

Un directorio `laboratorio/normativa/<nombre>/` con una sola skill, `skills/normativa/`, de nueve ficheros: `SKILL.md`
(frontmatter con `name: normativa`) y las ocho referencias (contracts/variantes.md §1). Se copia tal cual como directorio
de skills de las sesiones de su fila. Cambia solo cuando cambia `skills/normativa` (FR-002) o el hito siguiente la retira;
no crece con el uso.

## 3. Declaración de una variante (en el test)

Lo que `TestVariantesDeNormativa` aplica a `normativa-control` para calcular la variante esperada (research D5):

- **Sustitución**: `{fichero, antes, después, hito}` —`antes`, una sola vez en el fichero de control; `después`, sus
  bytes en la variante; `hito`, el texto que fija el hito, igual a `después` sin la marca `- ` y con cada salto de línea y
  la sangría que lo sigue contados como un espacio—. `remisiones-en-orden`: cinco en `SKILL.md`;
  `testigo-en-referencia`: una en `SKILL.md` (contracts/variantes.md §3 y §5).
- **Regla del cuerpo corto**: las secciones que se quedan enteras, el encabezado que se renombra y la viñeta que se
  queda de su sección (contracts/variantes.md §4).
- **Añadido**: lo que sigue al `con_herramientas.md` de control en `testigo-en-referencia`, con el párrafo de
  `references/cierre.md` de H39 leído de ese fichero (contracts/variantes.md §5).

Una declaración que ya no se aplica sobre control —un `antes`, un encabezado o una viñeta que no está una sola vez— es un
fallo del test, con su motivo (FR-006).

## 4. Eval del sondeo

Una de las cuatro de §1, de `evals/normativa/`, con sus `comandos`, `citas` y, en la 15, sus `lecturas`, y en la 22, su
`redaccion_fechada`: las del job, sin copia. Cada sesión se prepara con las consultas de las 23 evals de la carpeta, como
en el job (research V8, D7).

## 5. Sondeo y sesión

Como en H39: `DIRECTORIO/sondeo.json` (`{"modelo","claude_code"}`) y `DIRECTORIO/<variante>/<modo>/<eval sin
.yaml>-<nn>/`, con la transcripción y lo demás que deja la sesión. Una sesión contada da:

| Campo | Qué es |
|---|---|
| `ruta` | la de su directorio, relativa a la ruta dada a la tabla |
| `motivo` | por qué no entra en `n` (sin terminar o sin respuesta); vacío si entra |
| `activaAntes` | activa la skill de la familia antes de la primera herramienta de kitlegal de su modo |
| aperturas | por cada referencia de la familia, el turno de su primera apertura; 0 si no la abre (H39: una referencia) |
| `cumple` | la respuesta cumple la regla testigo |

## 6. Tabla

Una fila por clave `{variante (posición en su familia), modo, modelo, versión}`, con sus sesiones; las celdas de la
familia (contracts/tabla.md §3). La de H39 no cambia (FR-023).
