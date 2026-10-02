# Contrato · las dos skills con la orden y la herramienta

Qué cambia en `skills/boe-legislacion/SKILL.md` y en `skills/legal-core/SKILL.md`, en la tabla que genera
`make skills-sync` y en las comprobaciones de `make ci`. Requisitos: FR-030 a FR-036, FR-061 y FR-077. Las referencias V
y D son de [research.md](../research.md).

## 1. Qué pide cada skill, y cuántas veces por pregunta

No cambia respecto de hoy: cambia la forma de pedirlo.

| Skill | Operación | Orden | Herramienta | Veces por pregunta |
|---|---|---|---|---|
| `boe-legislacion` | buscar la norma | `kitlegal boe buscar <texto>... --json` | `boe_buscar` (`texto`) | 0 si la norma está en `references/normas.md`; 1 o más si no |
| | índice | `kitlegal boe indice <norma> --json` | `boe_indice` (`norma`) | 0 o 1 por norma |
| | leer un bloque y comprobar su redacción | `kitlegal boe articulo <norma> <bloque> --json && kitlegal graph check <norma> <bloque> --json` | `boe_articulo` (`norma`, `bloque`) y, recibido su resultado y si no es un error, `graph_check` (`norma`, `bloques`) | 1 por bloque: dos llamadas |
| | leer varios bloques | `kitlegal boe articulos … && kitlegal graph check …` | `boe_articulos` y después `graph_check` con los mismos `bloques` | 1 por grupo de bloques |
| | vigencia y modificaciones | `kitlegal boe metadatos`, `kitlegal boe analisis` | `boe_metadatos`, `boe_analisis` | 0 o 1 por norma |
| `legal-core` | territorio | `kitlegal territorio resolver <consulta> --json` | `territorio_resolver` (`consulta`) | 1 por municipio |

`graph_show` y `graph_stats` siguen en la tabla de `boe-legislacion` (su frontmatter declara el applet `graph`) y el
protocolo sigue sin pedirlas.

## 2. Cómo dicen las dos formas

El texto exacto de los dos ficheros es el del prototipo: [skills-prototipo.diff](./skills-prototipo.diff), aplicado
sobre `22b5bda`. Con él, `boe-legislacion` tiene **297 líneas** y `legal-core`, 190, y pasan las comprobaciones de hoy
(V35). Una línea del prototipo pasa de las 120 columnas (la de «por su orden o por su herramienta»): la tarea la parte
sin cambiar el recuento. El diff no toca la región generada, que cambia §4.

| # | Dónde (`boe-legislacion`) | Qué pasa a decir | Requisito |
|---|---|---|---|
| C1 | Principio de «Protocolo» | Las dos formas: la herramienta, si entre las del agente hay una con el nombre `<applet>_<verbo>`, sola o detrás de un prefijo (`mcp__kitlegal__boe_articulo`), con sus argumentos por nombre, y «si la tienes, úsala siempre»; la orden, si no. Que donde un paso encadena dos órdenes con `&&` son dos llamadas seguidas, la segunda solo si la primera no falló; y que donde dice que una orden termina con un código, con herramientas es `data.clase` | FR-032, FR-033, FR-034 |
| C2 | Paso 2 | `boe_buscar` junto a `kitlegal boe buscar`; el ejemplo pasa de bloque a código en línea | FR-032 |
| C3 | Paso 3, índice | `boe_indice` junto a su orden; el ejemplo, en línea | FR-032 |
| C4 | Paso 3, lectura y comprobación | Detrás de las dos órdenes (Bash y PowerShell, que se quedan): «Con herramientas, llama a `boe_articulo` con `norma` y `bloque` y, recibido su resultado y solo si no es un error, a `graph_check` con la misma `norma` y ese bloque en `bloques`» | FR-033 |
| C5 | Paso 3, varios bloques | `boe_articulos` y `graph_check` junto a sus órdenes; «una lectura de varios bloques falla entera…», con `boe_articulo` para pedirlos por separado | FR-032, FR-033 |
| C6 | Paso 3, metadatos y análisis | `boe_metadatos` y `boe_analisis` junto a sus órdenes, en línea | FR-032 |
| C7 | Pasos 4 y 5 | «ni con sus herramientas»; «después de la última orden o llamada»; «por su orden o por su herramienta» | FR-032 |
| C8 | «Redacción modificada» | `graph_check` junto a `kitlegal graph check`; «en su misma orden o en la llamada siguiente» | FR-033 |
| C9 | «Comandos», antes de la tabla | La correspondencia entera: `2` o `argumentos`, `3` o `no-encontrado`, `4` o `fuente-no-disponible`, `5` o `limite-o-tos`, `6` o `identidad-humana`, `1` o `inesperado` | FR-034 |
| C10 | Regla 2 | Su última frase pasa a «Si no tienes ni la herramienta ni el binario, vale la regla 8» | FR-035 |
| C11 | Regla 7 | «termina con otro código —o `graph_check` devuelve un error—» | FR-034 |
| C12 | Regla 8, nueva | §3 | FR-035 |

En `legal-core`: el mismo bloque al principio de «Protocolo», para `territorio_resolver`; el paso 2 nombra la
herramienta y su `consulta`; el paso 4 pone la clase junto a cada código (`argumentos`, `no-encontrado`, «o cualquier
otra `clase`»); «Comandos» da la correspondencia; y la regla 7, nueva, sustituye a la viñeta del `PATH`.

**Lo que no cambia** (FR-036): los cinco pasos y sus reglas, la `description` y el frontmatter, la forma de la cita, de
los avisos, de `⚠ REDACCIÓN MODIFICADA:` y de la frase de la regla 7, las dos órdenes para PowerShell, y que `SKILL.md` no
nombra evals, job ni modelos. Las tres órdenes sueltas que pasan de bloque a código en línea y las líneas en blanco que
se quitan junto a los bloques son lo que hace sitio: no cambian lo que se pide.

## 3. La línea `⚠ SIN CONSULTA AL BOE:`

Forma, igual en las dos skills, en un bloque `text`:

```text
⚠ SIN CONSULTA AL BOE: <causa>. Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/
```

- **Cuándo**: el agente no tiene la herramienta y la orden falla porque `kitlegal` no está —el shell no lo encuentra— o
  no puede ejecutar ninguna orden.
- **Qué lleva la respuesta**: esa línea, con la causa en lugar del marcador y la dirección en la misma línea. Nada del
  contenido de la norma —en `legal-core`, tampoco ningún dato de territorio—, ni de memoria ni con salvedades, y ninguna
  cita.
- **Cuándo no**: si tiene la herramienta y la llamada falla, lee la `clase` y aplica la regla de ese código (regla 2 en
  `boe-legislacion`; paso 4 en `legal-core`).
- **Tamaño**: una línea de unos 150 B más la causa; una por respuesta, no una por norma.
- **Cuándo deja de darse**: en la primera pregunta tras instalar kitlegal o declarar el servidor.
- **La lista de expresiones**: la línea no lleva ninguna (V35), así que no se declara forma fija; la lista no cambia.

## 4. La tabla generada

`RenderizarTabla` (`internal/skills/comandos.go`) escribe, por applet:

```markdown
### `kitlegal boe`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal boe articulo <norma> <bloque>` | `boe_articulo` | Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia. | objeto con `norma`, `bloque`, … |
```

y, tras la última tabla, en lugar de «Todas devuelven el sobre…»:

```markdown
La orden y la herramienta de cada fila devuelven el mismo sobre: `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.
```

La línea de las banderas comunes no cambia. El nombre de la herramienta sale del applet y del verbo de la descripción de
`--describe`, como la orden. Ninguna fila ni línea de más: 9 filas en `boe-legislacion` y 1 en `legal-core`, como hoy.

## 5. Comprobaciones de `make ci`

| Qué | Dónde | Requisito |
|---|---|---|
| La región de cada `SKILL.md` es la que se generaría (sin drift), frontmatter, sin `scripts/`, sin instrucciones de evals, `SKILL.md` < 300 líneas | `TestSkillsDelRepositorio`, en `make skills-check` | FR-030, FR-036 |
| La tabla coincide con la gramática de cada verbo, con su columna nueva | `TestTablaDeComandosCoincideConLaGramatica`; `TestRenderizarTabla` (`internal/skills/comandos_test.go`) | FR-030 |
| Toda orden de la tabla existe en el registro **y toda herramienta, entre las del servidor** (las de `herramientasDe` con el registro de producción); con un control que ve fallar una herramienta inventada y una orden inventada | `TestOrdenesDeLasSkillsEmpotradas`, en `make skills-check` | FR-031 |
| La prosa de `boe-legislacion` sin expresiones de la lista ni fechas con cifras; la respuesta hecha de sus bloques `text`, sin expresiones; el calibrado en 36, 11 y 9; cada orden con `&&`, con su forma para PowerShell | `TestEvalsDelRepositorio`, en `make skills-check`, sin cambios | FR-036 |
| Cada `SKILL.md` dice la línea de §3 tal cual en un bloque `text`, y la línea casa con lo que el juicio reconoce (`ExtraerSinConsulta`) | subprueba `linea-sin-consulta` de `TestEvalsDelRepositorio` | FR-035 |

## 6. Uso, de fuera adentro

`SKILL.md` lo lee el modelo una vez por conversación en que la skill se activa: 297 líneas en `boe-legislacion` (hoy,
298) y 190 en `legal-core` (hoy, 170). La tabla tiene las mismas filas, cada una con unos 20 B más. No da señales.

## 7. `CHANGELOG.md` y versiones

*Unreleased* registra `boe-legislacion` v0.1.5 y `legal-core` v0.1, con lo que cambia para quien las usa: cada operación
de dos formas, la herramienta primero si el agente la tiene, y la línea `⚠ SIN CONSULTA AL BOE:` (FR-061).
