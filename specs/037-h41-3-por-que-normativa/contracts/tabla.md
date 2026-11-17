# Contrato: la tabla con la instalación y el presupuesto de cada fila

FR-020 a FR-024; SC-006, SC-007. Research D10 y D11; data-model §6. Lo que no se nombra aquí es la tabla de H41.2 tal cual
(`specs/036-h41-2-por-qu-e/contracts/tabla.md`).

## 1. Cómo se lanza

Como hoy, sin red, modelo ni credencial: `make laboratorio-tabla DIRECTORIO=<ruta>` (`go run ./cmd/laboratorio tabla
<ruta>`), con la ruta de un sondeo o de un directorio de sondeos. Un `sondeo.json`, una sesión o un transcript que no sabe
leer es el defecto `inesperado` de hoy, con la ruta y código 1.

## 2. Cuándo lleva las columnas nuevas

| Sondeos de la ruta | Tabla |
|---|---|
| de H39 | la de H39, byte a byte (FR-023) |
| de una familia del producto, ninguno con `instalacion` ni `presupuesto` en su `sondeo.json` (todos los de antes de este hito y los de este hito sin las dos opciones) | la de hoy, byte a byte, tabla y detalle (FR-022) |
| de una familia del producto, alguno con `instalacion` o `presupuesto` | la de §3 |

## 3. Qué da y en qué orden

Detrás de «Skills instaladas», dos columnas:

| Columna | Qué | Forma |
|---|---|---|
| Instalación | `instalacion` del `sondeo.json`; sin ella, `producto` si `skills_instaladas` tiene alguna skill además de la de la familia, y si no —solo ella, o sin la clave—, la skill de la familia | `normativa` · `producto` · `producto-descripciones-cortas` |
| Presupuesto | `presupuesto` del `sondeo.json`, o vacía | `1000000` o nada |

Una fila por variante, modo, modelo, versión de Claude Code, instalación, presupuesto y skills instaladas, en ese orden
—el presupuesto vacío antes que cualquier valor, y los valores de menor a mayor—: las cuatro filas del sondeo de H41.3
salen como las enumera FR-021, `sola`, `control`, `diagnóstico` y `remedio`. Lo demás de la fila —«Sin terminar», `n`,
«Activan antes», «Activan otra skill», «Otras skills», «No activan ninguna», las referencias y la regla— y todo el detalle
se cuentan y se escriben como en H41.2 (FR-020). Una fila sin sesiones terminadas da su instalación y su presupuesto y
los recuentos vacíos, como hoy.

## 4. Ejemplo

La tabla de los sondeos fijos de §6, la que dan `laboratorio tabla` sobre ellos y el guion congelado, byte a byte, con su
salto de línea final:

```text
# Sondeo de las variantes de normativa

| Variante | Modo | Modelo | Claude Code | Skills instaladas | Instalación | Presupuesto | Sin terminar | n | Activan antes | Activan otra skill | Otras skills | No activan ninguna | buscar_la_norma.md | como_se_cita.md | con_herramientas.md | lecturas_que_fallan.md | normas.md | redaccion_en_una_fecha.md | redaccion_modificada.md | varios_bloques.md | Cumplen la regla |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| normativa-producto | herramienta | claude-sonnet-5-5 | 2.1.284 | normativa | normativa |  | 0 | 2 | 1 de 2 · 50,0 % [9,5–90,5 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | ninguna | 1 de 2 · 50,0 % [9,5–90,5 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] |  |
| normativa-producto | herramienta | claude-sonnet-5-5 | 2.1.284 | consulta-juridica, jurisprudencia, normativa, plazos-y-territorio | producto |  | 0 | 3 | 1 de 3 · 33,3 % [6,1–79,2 %] | 1 de 3 · 33,3 % [6,1–79,2 %] | consulta-juridica: 1 de 3 · 33,3 % [6,1–79,2 %] | 1 de 3 · 33,3 % [6,1–79,2 %] | 0 de 3 · 0,0 % [0,0–56,2 %] | 0 de 3 · 0,0 % [0,0–56,2 %] | 0 de 3 · 0,0 % [0,0–56,2 %] | 0 de 3 · 0,0 % [0,0–56,2 %] | 0 de 3 · 0,0 % [0,0–56,2 %] | 0 de 3 · 0,0 % [0,0–56,2 %] | 0 de 3 · 0,0 % [0,0–56,2 %] | 0 de 3 · 0,0 % [0,0–56,2 %] |  |
| normativa-producto | herramienta | claude-sonnet-5-5 | 2.1.284 | consulta-juridica, jurisprudencia, normativa, plazos-y-territorio | producto | 1000000 | 0 | 2 | 2 de 2 · 100,0 % [34,2–100,0 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | ninguna | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] |  |
| normativa-producto | herramienta | claude-sonnet-5-5 | 2.1.284 | consulta-juridica, jurisprudencia, normativa, plazos-y-territorio | producto-descripciones-cortas |  | 1 | 2 | 1 de 2 · 50,0 % [9,5–90,5 %] | 1 de 2 · 50,0 % [9,5–90,5 %] | consulta-juridica: 1 de 2 · 50,0 % [9,5–90,5 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] |  |
| normativa-producto | herramienta | claude-sonnet-5-5 | 2.1.322 | consulta-juridica, jurisprudencia, normativa, plazos-y-territorio | producto |  | 1 | 0 |  |  |  |  |  |  |  |  |  |  |  |  |  |

## Detalle

| Sesión | Terminada | Activa antes | Activaciones | buscar_la_norma.md | como_se_cita.md | con_herramientas.md | lecturas_que_fallan.md | normas.md | redaccion_en_una_fecha.md | redaccion_modificada.md | varios_bloques.md | Cumple |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| h41-3-sola/normativa-producto/herramienta/04-lgt-prescripcion-01 | sí | sí | normativa (antes) | no | no | no | no | no | no | no | no |  |
| h41-3-sola/normativa-producto/herramienta/04-lgt-prescripcion-02 | sí | no | ninguna | no | no | no | no | no | no | no | no |  |
| h41-3-control/normativa-producto/herramienta/04-lgt-prescripcion-01 | sí | sí | normativa (antes) | no | no | no | no | no | no | no | no |  |
| h41-3-control/normativa-producto/herramienta/04-lgt-prescripcion-02 | sí | no | consulta-juridica (antes) | no | no | no | no | no | no | no | no |  |
| h41-3-control/normativa-producto/herramienta/04-lgt-prescripcion-03 | sí | no | ninguna | no | no | no | no | no | no | no | no |  |
| h41-3-diagnostico/normativa-producto/herramienta/04-lgt-prescripcion-01 | sí | sí | normativa (antes) | no | no | no | no | no | no | no | no |  |
| h41-3-diagnostico/normativa-producto/herramienta/04-lgt-prescripcion-02 | sí | sí | normativa (antes) | no | no | no | no | no | no | no | no |  |
| h41-3-remedio/normativa-producto/herramienta/04-lgt-prescripcion-01 | sí | sí | normativa (antes) | no | no | no | no | no | no | no | no |  |
| h41-3-remedio/normativa-producto/herramienta/04-lgt-prescripcion-02 | sí | no | consulta-juridica (antes) | no | no | no | no | no | no | no | no |  |
| h41-3-remedio/normativa-producto/herramienta/04-lgt-prescripcion-03 | no: tope de 240 s agotado (código 124) |  |  |  |  |  |  |  |  |  |  |  |
| h41-2/normativa-producto/herramienta/04-lgt-prescripcion-01 | no: tope de 240 s agotado (código 124) |  |  |  |  |  |  |  |  |  |  |  |
```

Los intervalos, comprobados con la fórmula de `recuento` (research V6). La última fila es la de un sondeo de H41.2 junto a
los de H41.3: sin la constancia, su instalación sale de sus `skills_instaladas`, `producto`, y su presupuesto, vacío
(FR-022); y, con su única sesión sin terminar, da su instalación y su presupuesto con los recuentos vacíos, como hoy una
fila con `n` = 0 (caso límite de FR-020).

## 5. Uso, de fuera adentro

La tabla la lee la persona una vez, para la lectura de la propuesta y el ADR; ninguna `SKILL.md` la pide: 0 invocaciones
por pregunta. En el sondeo de H41.3 son cuatro filas, que el guion lee de cuatro `sondeo.json` —una lectura por sondeo,
como hoy—, y 640 líneas de detalle. Las dos columnas nuevas añaden 29 bytes a la cabecera, 8 al separador y, por fila,
el nombre de la instalación, el presupuesto y sus dos separadores: 15 bytes en `sola`, 14 en `control`, 21 en
`diagnóstico` y 35 en `remedio`, 85 en las cuatro, sobre unos 2,3 KB de filas (las de §4 miden de 479 a 598 bytes) y unos
93 KB de detalle, que no cambia (640 líneas de unos 145 bytes; las de §4 miden de 132 a 153). Medido sobre la tabla de §4
en esta sesión. Su tamaño lo fijan las filas y las sesiones que se piden, no el uso del producto: cientos de normas y
miles de bloques consultados no lo cambian. Las columnas solo se dan si algún
sondeo de la ruta lleva la constancia; una tabla sin ella sale como hoy. Deja de leerse cuando el ADR está escrito; otro
sondeo da otra tabla.

## 6. Controles y fixtures

- **Fixtures** (tarea `[datos]`, sin red): `internal/evals/testdata/laboratorio/familias/instalacion/sondeo/`, cinco
  sondeos de `normativa-producto` en modo herramienta con `claude-sonnet-5-5`, y `tabla.md`, la de §4. Cada `sondeo.json`,
  el de contracts/arnes.md §5 de su fila, con `claude_code` `2.1.284`, y el de `h41-2`, el de control con `2.1.322`:

  | Sondeo | `sondeo.json` | Sesiones, en `normativa-producto/herramienta/` |
  |---|---|---|
  | `h41-3-sola` | sola | `-01` activa `normativa` antes; `-02` ninguna |
  | `h41-3-control` | control | `-01` activa `normativa` antes; `-02` solo `consulta-juridica`, antes; `-03` ninguna |
  | `h41-3-diagnostico` | diagnóstico | `-01` y `-02` activan `normativa` antes |
  | `h41-3-remedio` | remedio | `-01` activa `normativa` antes; `-02` solo `consulta-juridica`, antes; `-03` sin terminar |
  | `h41-2` | control, con `2.1.322` | `-01` sin terminar |

  Cada sesión, con la forma de las de la suite congelada de H41.2
  (`specs/036-h41-2-por-qu-e/aceptacion/laboratorio-tabla.txtar`): `codigo-de-la-sesion`, `sesion.err` vacío y
  `sesion.jsonl` con `system/init` con el modelo y la versión de su sondeo, un `assistant` con su `message.id` por
  respuesta, los resultados en mensajes `user` y `result` con la respuesta. «Activa `normativa` antes»: un `tool_use`
  de `Skill` con `{"skill":"normativa"}` y, en el turno siguiente, `mcp__kitlegal__boe_articulo`; «solo `consulta-juridica`,
  antes»: lo mismo con `{"skill":"consulta-juridica"}`; «ninguna»: `mcp__kitlegal__boe_articulo` sin ningún `Skill`; «sin
  terminar»: código `124`, `Skill` de `consulta-juridica` y `mcp__kitlegal__boe_articulo`, sin `result`. Ninguna abre una
  referencia.
- **`TestTablaConLaInstalacion`** (`internal/evals/laboratorio_tabla_test.go`): sobre esos sondeos, la salida es `tabla.md`
  byte a byte, con código 0 y nada en la salida de error (FR-024, SC-006); y la instalación de una fila sin la
  constancia en una tabla con ella, en sus cuatro casos —`instalacion` en el `sondeo.json`; las cuatro skills; solo la de
  la familia; sin `skills_instaladas`— (FR-020, FR-022).
- **Sin la constancia** (FR-022, FR-023, SC-007): `TestTablaConLasActivaciones`, `TestTablaDeLasFamilias`,
  `TestTablaDeNormativa` y `TestTablaDelLaboratorio`, con sus tablas fijas sin cambiar; y las suites congeladas de H39 a
  H41.2 en `TestEntregaDelHito`.
- **El guion congelado**: contracts/aceptacion.md.
