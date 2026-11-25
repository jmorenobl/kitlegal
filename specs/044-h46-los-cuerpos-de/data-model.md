# Data model: H46 · Los cuerpos de las skills de capacidad, con sitio para las materias

Ninguna entidad nueva del binario: ningún tipo exportado, campo de `data`, esquema ni fichero de `data/` (FR-019). Lo
que cambia son tres textos que lee el agente, una regla de `skills-check` y los tests que fijan unos y otra.

## 1. El cuerpo de un `SKILL.md` y su región de los paquetes

- **Cuerpo**: los bytes que siguen a la línea `---` que cierra el frontmatter, hasta el final, como los deja
  `make skills-sync` (con la tabla de comandos generada y las líneas de los paquetes); se mide en runas
  (`utf8.RuneCount`), no en bytes (`docs/SKILLS.md` §3; `cuerpoDeLaSkill`, `internal/skills/sincronia.go:635`).
- **Región de los paquetes**: lo que hay entre la línea `<!-- paquetes: lo genera make skills-sync -->` y la línea
  `<!-- fin de los paquetes -->`, sin ellas: `entreMarcas{desde, hasta}` de `regionDeLosPaquetes.buscar`
  (`internal/skills/paquetes.go:58-91`). Una línea por paquete con referencia en la skill, `- Si la pregunta es de lo
  <paquete>, lee [<paquete>](references/<paquete>.md).\n`, de 52 + 3n caracteres con n los del nombre. Hoy: una en
  `plazos-y-territorio` (lo administrativo, 94), ninguna en `normativa` ni en `jurisprudencia`.
- **Cuerpo sin las líneas de los paquetes** (FR-005): runas del cuerpo menos runas de la región. Las dos marcas cuentan.
  Sin las dos marcas en pareja no se define, y la skill no tiene el defecto de FR-005 (FR-006); las que no van en pareja
  ya dan el suyo.

| Medida | Tope | Defecto si se pasa | Skills a las que se aplica |
|---|---|---|---|
| Líneas del cuerpo | 499 | `<skill>: SKILL.md: cuerpo de <n> líneas (máximo 499)` | todas (sin cambios) |
| Runas del cuerpo | 11 000 | `<skill>: SKILL.md: cuerpo de <n> caracteres (máximo 11000)` | todas (sin cambios) |
| Runas del cuerpo sin las líneas de los paquetes | 9600 (`maximoSinLosPaquetes`) | `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de <n> caracteres (máximo 9600)` | las que llevan el par de marcas |

Orden de los defectos del cuerpo: líneas, caracteres y, detrás, el de FR-005. Con los textos del hito (research P1):
`normativa` 9498 / 9498, `plazos-y-territorio` 9577 / 9483, `jurisprudencia` 9473 / 9473 (cuerpo / sin las líneas).

## 2. La tabla de comandos

La región generada desde `--describe` entre `<!-- inicio de la tabla de comandos: generada desde --describe con make
skills-sync, no editar -->` y `<!-- fin de la tabla de comandos -->`: por cada applet de `kitlegal-applets`, un título
`### \`kitlegal <applet>\``, la cabecera `| Orden | Herramienta | Qué hace |`, el separador `|---|---|---|` y una fila por
verbo con tres celdas: la sintaxis de la orden, la herramienta `<applet>_<verbo>` y lo que hace. Termina con una línea en
blanco. Ya no lleva la columna «Qué devuelve en `data`», la línea del sobre ni la de las banderas comunes (FR-001).

Lo que `LeerDescripcionDeVerbo` lee de cada documento no cambia (`DescripcionDeVerbo`: `Applet`, `Verbo`, `Hace`,
`Argumentos`, `BanderasPropias`, `Devuelve`, `Banderas`, `Sobre`), y `RenderizarTabla` sigue exigiendo el mismo sobre y
las mismas banderas en todas las filas (FR-002). Bytes y ejemplo en contracts/tabla-de-comandos.md.

## 3. Los casos de `casosDePresupuestos` (FR-008)

Por skill del árbol, cada uno sobre su copia; `p` son las runas entre las marcas en el `SKILL.md` del árbol.

| Caso | Skill | Alteración | Defectos esperados, en su orden |
|---|---|---|---|
| `cuatrocientas-noventa-y-nueve-lineas` | todas | `partirElCuerpo(…, 499)` | ninguno |
| `quinientas-lineas` | todas | `partirElCuerpo(…, 500)` | `<skill>: SKILL.md: cuerpo de 500 líneas (máximo 499)` |
| `once-mil-caracteres` | sin marcas | `ensancharElCuerpo(…, 11000)` | ninguno |
| `once-mil-y-un-caracteres` | sin marcas | `ensancharElCuerpo(…, 11001)` | `<skill>: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)` |
| `nueve-mil-seiscientos-sin-los-paquetes` | con marcas | `ensancharElCuerpo(…, 9600+p)` | ninguno |
| `nueve-mil-seiscientos-uno-sin-los-paquetes` | con marcas | `ensancharElCuerpo(…, 9601+p)` | `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de 9601 caracteres (máximo 9600)` |
| `once-mil-y-un-caracteres` | con marcas | `ensancharElCuerpo(…, 11001)` | `<skill>: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)` y `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de <11001−p> caracteres (máximo 9600)` |

Hoy: `consulta-juridica` sin marcas; `normativa` y `jurisprudencia` con marcas y `p = 0`; `plazos-y-territorio` con
`p = 94`. En la copia de `h45-paquete-de-mas`, `normativa` lleva `p = 70` (el paquete `prueba`).

## 4. Las declaraciones del producto (FR-014)

`declaracionesDelProducto[skill]` (`internal/evals/laboratorio_familias_test.go:436`): `cambios []cambioDelCuerpo`
(`antes`, `despues`: líneas enteras del cuerpo, fuera de lo generado) y `enLasReferencias []cambioEnUnaReferencia`
(`referencia`, `antes`, `despues`). El hito añade 15 entradas, una por texto cambiado o movido, con los bytes de
`f89093dd` (contracts/textos-y-controles.md §2). Ni `desde`, ni `movimientos`, ni el frontmatter cambian.

## 5. Lo que fija el texto en los tests y en los guiones

| Control | Qué fija | Cambio |
|---|---|---|
| `remisionesDelPlan` (`internal/skills/remisiones_test.go:71`) | frases de las órdenes de leer, por skill | las de `municipios.md` y `consulta_preparada.md`, nuevas |
| `h41-skills-instaladas.txtar` | la orden de `consulta_preparada.md` y la tabla «Cómo se dio / Qué se pide» | la orden con su frase nueva en el `SKILL.md`; la tabla, en la referencia |
| `h41-1-skills-instaladas.txtar` y su copia congelada | la orden de `municipios.md` al principio del «Protocolo» | la orden nueva; los dos ficheros iguales |
| `regionDeAlfa` (`internal/skills/sincronia_test.go:118`) | la tabla generada de `alfa` | tres columnas, sin las dos líneas |
| `columnasDeLaTabla`, `filaDe*`, `pieDeLaTabla` (`internal/skills/comandos_test.go:799-823`) | los bytes de `RenderizarTabla` | tres columnas; el pie, una línea en blanco |
| fixture de `TestOrdenesDeLasSkillsEmpotradas/control` (`internal/app/skills_test.go:397-410`) | filas de la región | cabecera de tres columnas, filas sin su cuarta celda |
