# Contrato: lo que cambia en dos skills, en el control de PowerShell y en el laboratorio de H41

## 1. El movimiento de `normativa` (FR-009, §C.2)

- `skills/normativa/SKILL.md` pierde sus líneas 38 a 41 de `main`, las cuatro con dos espacios de sangría:
  «En PowerShell 5.1:», la valla ` ```powershell`, la orden y la valla de cierre. Nada más cambia en el cuerpo (ni la
  orden de Bash de las líneas 35-37, ni «La comprobación, nunca sin argumentos…» de la 42).
- `skills/normativa/references/con_herramientas.md` acaba, tras una línea en blanco, con esas cuatro líneas sin la
  sangría (bytes de §C.2 del spec). El fichero pasa de 2 442 bytes a unos 2 616; el cuerpo, de 10 926/10 927 a unos 10 751
  caracteres.
- Quién lo lee: el agente que usa `normativa`, una vez por conversación y antes de la primera orden (el cuerpo lo
  manda en su línea 16); la forma movida la usa quien está en PowerShell 5.1 sin herramientas MCP, una vez por bloque
  leído sin fecha. No crece con el uso.

## 2. El control `ordenes-para-powershell` (FR-011, FR-012, FR-013, §C.3)

Regla nueva, solo para las órdenes del `SKILL.md`: una orden de Bash de lectura y comprobación que no tiene detrás su
forma para PowerShell queda cubierta si alguna referencia a mano de la misma skill lleva una orden PowerShell con el
mismo verbo, la misma norma y los mismos bloques, en cualquier línea. Las órdenes de cada referencia siguen exigiendo
la suya detrás, en su fichero. Ningún mensaje cambia (`sinSuForma`, `conjunto_test.go:4024`).

Casos de `TestControlesSobreLasReferenciasAMano` (`mutantesDeLasOrdenes`), sobre la copia de `normativa` ya movida:

| §C.3 | Caso | Mutación | Esperado |
|---|---|---|---|
| 1 | nuevo, `orden-del-skill-md-con-su-forma-en-una-referencia` | ninguna: el estado de la rama | sin defecto |
| 2 | nuevo, `forma-en-una-referencia-con-otros-bloques` | la orden de `con_herramientas.md` con otro bloque | `normativa: SKILL.md: línea 36, «…»` + `sinSuForma` |
| 3 | nuevo, `forma-en-ninguna-referencia` | quitar de `con_herramientas.md` lo movido | el mismo defecto del `SKILL.md` |
| 4 | nuevo, `orden-de-una-referencia-con-su-forma-en-el-skill-md` | una orden de Bash en `a_mano.md` y su forma para PowerShell añadida solo al `SKILL.md` | `normativa: references/a_mano.md: línea 4, «…»` + `sinSuForma` |
| 5 | `orden-con-su-forma-en-otra-referencia`, de hoy | — | su defecto, como hoy |
| 6 | `sin-las-ordenes-de-articulos`, de hoy | — | «ninguna orden de Bash lee con «kitlegal boe articulos»…», como hoy |

Los casos de hoy que no son de §C.3 se quedan; no entra ninguno más.

`docs/SKILLS.md`, fila de `ordenes-para-powershell` (línea 215), pasa a decir: «Cada orden de lectura o de comprobación
de una referencia lleva detrás, en el mismo fichero, su forma para PowerShell; la de una orden del `SKILL.md` va detrás
de ella o en cualquier referencia a mano de la skill, con el mismo verbo, la misma norma y los mismos bloques, se lea
esa referencia siempre o solo en algunos casos».

## 3. El laboratorio de H41 (FR-010, FR-014)

- `compararLosFicheros` no da «sobra» para `skills/<skill>/references/<v>.md` si `<v>` es un `vertical` de
  `data/normas.yaml` y la base (la variante `desde` o el control) no tiene ese fichero. Si la base lo tiene, se compara.
- La declaración de `normativa` gana el `cambio` y el `movimiento` de research D7.
- Para probarlo, la comparación devuelve sus diferencias (`diferenciasDeLosFicheros`, que `compararLosFicheros` pasa a
  `assert`) y un subtest nuevo, `referencias-de-un-paquete`, las pide sobre los ficheros leídos del producto con un
  fichero añadido o cambiado en memoria (SC-009): `skills/normativa/references/administrativo.md` (un vertical; el
  control de `normativa` no la tiene) → 0 diferencias; la misma como `materia-inventada.md` → 1, «sobra»; un byte
  cambiado fuera de las regiones de `skills/plazos-y-territorio/references/administrativo.md`, que su base tiene → 1. El texto movido de vuelta
  al cuerpo pone en rojo el subtest `movimientos` de hoy (US5.4).

## 4. La descripción de `plazos-y-territorio` (FR-023, FR-024)

- El texto de FR-023 (429 caracteres, `wc -m` sin el salto) en la `description` del frontmatter y en
  `laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt`, byte a byte; la suma de las cuatro de la variante,
  1 692 (tope 1 700); cada una, como mucho 450; cada `description` del producto, como mucho 900
  (`internal/skills/frontmatter.go:58`).
- La lee el agente en cada sesión con las skills instaladas para elegir skill: 429 caracteres por sesión, sin crecer con
  el uso.
- Cambian con ella (research D8): `descripcionCortaDePlazosYTerritorio`; las sumas de «una de 451», «una suma de 1701» y
  «una quinta skill sin texto» de `TestDescripcionesDelProducto` y la de «una que empieza por <ejemplo>» de
  `TestDescripcionesCortas`, derivadas; el guion `h41-4-skills-instaladas.txtar` y su copia congelada
  `specs/038-h41-4-las-skills-del/aceptacion/skills-instaladas.txtar`, en su línea 34 (la descripción con `( |\n  )` entre
  palabras), en una tarea `[datos]`.
- `CHANGELOG.md`, *Unreleased*, «Cambiado»: la descripción nueva de `plazos-y-territorio`, que deja de nombrar una
  materia, y su trabajo de evals a tres sesiones a la vez.
