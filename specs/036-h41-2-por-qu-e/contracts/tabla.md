# Contrato: la tabla con las skills instaladas y las activaciones

FR-020 a FR-024; SC-005, SC-006. Research D12 a D14; data-model §1 y §4. Lo que no se nombra aquí es el guion de H41 tal
cual (`specs/034-h41-las-skills-del/contracts/tabla.md`), y el de H39 para su familia, byte a byte.

## 1. Cómo se lanza

```sh
make laboratorio-tabla DIRECTORIO=<ruta>     # go run ./cmd/laboratorio tabla <ruta>
```

Sin red, modelo ni credencial; la misma orden, la misma salida de error y los mismos códigos. La familia, la de la primera
variante, como en H41. Lo que el guion no sabe leer sigue la regla genérica del spec: defecto `inesperado`, código 1, sin
más casos.

## 2. Cuándo lleva las columnas nuevas

| Sondeos de la ruta | Tabla |
|---|---|
| de H39 | la de H39, byte a byte (FR-023) |
| de una familia del producto, ninguno con `skills_instaladas` en su `sondeo.json` (todos los anteriores a este hito) | la de hoy, byte a byte, tabla y detalle (FR-022) |
| de una familia del producto, alguno con `skills_instaladas` | la de §3; la fila de un sondeo sin la clave da como instalación la skill de la familia sola (FR-022) |

## 3. Qué cuenta y cómo se escribe

Una fila por variante, modo, modelo, versión de Claude Code y skills instaladas, en ese orden (la instalación, por orden de
texto). Columnas, con las de hoy en su sitio:

```text
| Variante | Modo | Modelo | Claude Code | Skills instaladas | Sin terminar | n | Activan antes | Activan otra skill | Otras skills | No activan ninguna | <una por referencia> | Cumplen la regla |
```

(«Cumplen la regla», solo en la de `normativa`, como hoy.) Sobre las `n` sesiones terminadas de la fila (FR-021), con la
lista de cada una (contracts/informe.md §1):

| Columna | Qué | Forma |
|---|---|---|
| Skills instaladas | `skills_instaladas` del sondeo, unidos por «, »; o la skill de la familia | `normativa` · `consulta-juridica, jurisprudencia, normativa, plazos-y-territorio` |
| Activan antes | las que activan la skill de la familia antes de la primera herramienta, como hoy | `k de n · p % [lo–hi %]` |
| Activan otra skill | las que activan, en cualquier turno, alguna skill que no es la de la familia | `k de n · p % [lo–hi %]` |
| Otras skills | cada otra skill, en orden de nombre, con las sesiones que la activan, unidas por «; »; `ninguna` si no hay | `consulta-juridica: k de n · p % [lo–hi %]` |
| No activan ninguna | las que no tienen ningún `tool_use` de `Skill` | `k de n · p % [lo–hi %]` |

Wilson al 95 % (z = 1,96), con un decimal y coma decimal, como hoy. Una sesión cuenta una vez en cada recuento: la que
activa la de la familia y otra, en las dos columnas; la que activa una skill dos veces, una vez. Con `n` = 0, las celdas de
los recuentos, vacías como hoy; las cinco primeras y la instalación, llenas.

El detalle, con «Activaciones» detrás de «Activa antes»: la lista de la sesión con «antes» o «después», como el párrafo
de `informe.md` (contracts/informe.md §3), o `ninguna`; vacía en una sesión sin terminar.

## 4. Ejemplo

La de las sesiones fijas de §5, tres sondeos de `normativa-producto` en modo herramienta con `claude-sonnet-5-5`: `h41-2-a`
(2.1.284, `normativa` sola), `h41-2-b` (2.1.284, las cuatro) y `h41` (2.1.317, de H41, sin `skills_instaladas`). Las
referencias son las ocho de `normativa`, que ninguna sesión abre:

```text
# Sondeo de las variantes de normativa

| Variante | Modo | Modelo | Claude Code | Skills instaladas | Sin terminar | n | Activan antes | Activan otra skill | Otras skills | No activan ninguna | buscar_la_norma.md | como_se_cita.md | con_herramientas.md | lecturas_que_fallan.md | normas.md | redaccion_en_una_fecha.md | redaccion_modificada.md | varios_bloques.md | Cumplen la regla |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| normativa-producto | herramienta | claude-sonnet-5-5 | 2.1.284 | consulta-juridica, jurisprudencia, normativa, plazos-y-territorio | 1 | 4 | 1 de 4 · 25,0 % [4,6–69,9 %] | 3 de 4 · 75,0 % [30,1–95,4 %] | consulta-juridica: 3 de 4 · 75,0 % [30,1–95,4 %] | 1 de 4 · 25,0 % [4,6–69,9 %] | 0 de 4 · 0,0 % [0,0–49,0 %] | 0 de 4 · 0,0 % [0,0–49,0 %] | 0 de 4 · 0,0 % [0,0–49,0 %] | 0 de 4 · 0,0 % [0,0–49,0 %] | 0 de 4 · 0,0 % [0,0–49,0 %] | 0 de 4 · 0,0 % [0,0–49,0 %] | 0 de 4 · 0,0 % [0,0–49,0 %] | 0 de 4 · 0,0 % [0,0–49,0 %] |  |
| normativa-producto | herramienta | claude-sonnet-5-5 | 2.1.284 | normativa | 0 | 2 | 1 de 2 · 50,0 % [9,5–90,5 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | ninguna | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] |  |
| normativa-producto | herramienta | claude-sonnet-5-5 | 2.1.317 | normativa | 1 | 0 |  |  |  |  |  |  |  |  |  |  |  |  |  |

## Detalle

| Sesión | Terminada | Activa antes | Activaciones | buscar_la_norma.md | como_se_cita.md | con_herramientas.md | lecturas_que_fallan.md | normas.md | redaccion_en_una_fecha.md | redaccion_modificada.md | varios_bloques.md | Cumple |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-01 | sí | no | consulta-juridica (antes) | no | no | no | no | no | no | no | no |  |
| h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-02 | sí | no | consulta-juridica (antes), normativa (después) | no | no | no | no | no | no | no | no |  |
| h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-03 | sí | sí | normativa (antes), consulta-juridica (después), consulta-juridica (después) | no | no | no | no | no | no | no | no |  |
| h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-04 | sí | no | ninguna | no | no | no | no | no | no | no | no |  |
| h41-2-b/normativa-producto/herramienta/04-lgt-prescripcion-05 | no: tope de 240 s agotado (código 124) |  |  |  |  |  |  |  |  |  |  |  |
| h41-2-a/normativa-producto/herramienta/04-lgt-prescripcion-01 | sí | sí | normativa (antes) | no | no | no | no | no | no | no | no |  |
| h41-2-a/normativa-producto/herramienta/04-lgt-prescripcion-02 | sí | no | normativa (después) | no | no | no | no | no | no | no | no |  |
| h41/normativa-producto/herramienta/04-lgt-prescripcion-01 | no: tope de 240 s agotado (código 124) |  |  |  |  |  |  |  |  |  |  |  |
```

La primera fila es la de la instalación de las cuatro porque «consulta-juridica, …» va antes que «normativa» por orden de
texto. Una fila de 40 con las cuatro skills, 39 sesiones como `h41-2-a/…-01` y una como `h41-2-b/…-03`, da en sus cuatro
recuentos `40 de 40 · 100,0 % [91,2–100,0 %]`, `1 de 40 · 2,5 % [0,4–12,9 %]`,
`consulta-juridica: 1 de 40 · 2,5 % [0,4–12,9 %]` y `0 de 40 · 0,0 % [0,0–8,8 %]` (research V3).

## 5. Uso, de fuera adentro

La lee la persona una vez, para la lectura de la propuesta de cambio y el ADR de FR-031. En el sondeo de FR-030, cuatro
filas de unos 570 bytes con las cuatro skills y algo menos con una (research V16), y el detalle de 160 sesiones de 150 a
190 bytes: unos 27 KB con las cabeceras. Su tamaño lo fijan las filas y las sesiones pedidas, no el uso del producto ni
nada que se acumule. No es una señal: no avisa de nada, deja de leerse cuando el ADR está escrito, y un sondeo nuevo da
otra tabla. Ninguna skill del producto la pide.

## 6. Controles

- **`TestTablaConLasActivaciones`** (`internal/evals/laboratorio_tabla_test.go`, en `make ci`; FR-020, FR-021, FR-022,
  FR-024; SC-005): sobre `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/` la salida es, byte a byte,
  su `tabla.md`, la de §4; y sobre un sondeo que el test monta con 39 copias de `h41-2-a/…-01` y una de `h41-2-b/…-03`,
  con las cuatro skills en `skills_instaladas`, la fila lleva las cuatro celdas de 40 de §4. Falla con la primera fila,
  recuento, intervalo, nombre o lista que difiere.

  | Sondeo | Sesión (`normativa-producto/herramienta/04-lgt-prescripcion-…`) | Lo que hace (turno) |
  |---|---|---|
  | `h41-2-a` | `-01` | `Skill` `normativa` (1); `mcp__kitlegal__boe_articulo` (2); respuesta |
  | `h41-2-a` | `-02` | `Skill` `normativa` y `mcp__kitlegal__boe_articulo` en la misma respuesta (1); respuesta |
  | `h41-2-b` | `-01` | `Skill` `consulta-juridica` (1); `boe_articulo` (2); respuesta |
  | `h41-2-b` | `-02` | `Skill` `consulta-juridica` (1); `boe_articulo` (2); `Skill` `normativa` (3); respuesta |
  | `h41-2-b` | `-03` | `Skill` `normativa` (1); `boe_articulo` (2); `Skill` `consulta-juridica` (3); `Skill` `consulta-juridica` (4); respuesta |
  | `h41-2-b` | `-04` | `boe_articulo` (1); respuesta |
  | `h41-2-b` | `-05` | sin terminar, código 124 |
  | `h41` | `-01` | sin terminar, código 124 |

  Cada `sondeo.json`, con `referencias_del_producto` = las ocho de `normativa`; los dos de `h41-2-*`, con su
  `skills_instaladas`.
- **El guion congelado** `h41-2-laboratorio-tabla.txtar` (contracts/aceptacion.md), en `TestEntregaDelHito`.
- **Sin cambiar ningún valor esperado** (FR-022, FR-023; SC-006): `TestTablaDeLasFamilias`, `TestTablaDeNormativa`,
  `TestTablaDelLaboratorio`, `TestLecturaConCd`, `TestReglaTestigo` y `TestOrdenesDeBash`, con sus fixtures, y las suites
  congeladas de H39, H40 y H41.
