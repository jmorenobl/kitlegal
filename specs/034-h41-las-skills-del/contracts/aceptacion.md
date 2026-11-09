# Contrato: los guiones de aceptación congelados

Research D18; V7, V8. Los escribe la primera tarea en `specs/034-h41-las-skills-del/aceptacion/` y quedan congelados;
el workflow los activa como `internal/app/testdata/script/h41-<nombre>.txtar` y los ejecuta `TestEntregaDelHito` en
`make ci`. El arnés e2e no cambia: `kitlegal` va en el `PATH` y `KITLEGAL_LABORATORIO_BIN` es el `laboratorio` del árbol
(H40). Cada guion abre con un comentario que nombra lo que cubre (historias, FR y SC) y no contiene los textos `usage: `,
`unknown command`, `unexpected command` ni `cannot parse`. Cada uno falla hoy por una aserción, no por no poder leerse.

## 1. `aceptacion/laboratorio-tabla.txtar` — historia 5 (FR-032, FR-040, FR-041; SC-006)

Órdenes, en este orden; cada `sh -c … || true` va delante de su orden con código, para que antes de implementar el guion
falle en la comparación y no por el código (el patrón de H40):

```text
exec sh -c '"$KITLEGAL_LABORATORIO_BIN" tabla normativa > normativa.out || true'
cmp normativa.out normativa-tabla.md
exec $KITLEGAL_LABORATORIO_BIN tabla normativa
cmp stdout normativa-tabla.md
! stderr .
exec sh -c '"$KITLEGAL_LABORATORIO_BIN" tabla plazos-y-territorio > plazos.out || true'
cmp plazos.out plazos-y-territorio-tabla.md
exec $KITLEGAL_LABORATORIO_BIN tabla plazos-y-territorio
cmp stdout plazos-y-territorio-tabla.md
! stderr .
exec sh -c '"$KITLEGAL_LABORATORIO_BIN" tabla jurisprudencia > jurisprudencia.out || true'
cmp jurisprudencia.out jurisprudencia-tabla.md
exec $KITLEGAL_LABORATORIO_BIN tabla jurisprudencia
cmp stdout jurisprudencia-tabla.md
! stderr .
! exec $KITLEGAL_LABORATORIO_BIN tabla mezcla
! stdout .
stderr '^laboratorio: \S*mezcla/h41/normativa-control no es ninguna variante de jurisprudencia$'
```

Ficheros del archivo. Cada sesión lleva `codigo-de-la-sesion`, `sesion.err` y `sesion.jsonl` con la forma de las de H40
(`system/init` con el modelo y la versión; un `assistant` con su `message.id` por respuesta; los resultados en mensajes
`user`; y `result`). Las rutas de las skills van bajo `/s/claude/skills/`. Cada `sondeo.json`, `{"modelo":
"claude-sonnet-5-5","claude_code":"2.1.317"}` en una línea y, donde se dice, su `referencias_del_producto`. El número
entre paréntesis es el turno.

- **`normativa/h41/`**, con `referencias_del_producto` = las ocho de la familia y `nueva.md`, en orden de nombre:

  | Sesión | Lo que hace |
  |---|---|
  | `normativa-control/herramienta/04-lgt-prescripcion-01` | `Skill` `normativa` (1); `mcp__kitlegal__boe_articulo` (2); respuesta |
  | `normativa-producto/orden/22-lcsp-contrato-menor-en-2019-01` | `Skill` (1); `Bash` `cd /s/claude/skills/normativa/references` (2); `Bash` `cat normas.md && cd /s/claude/skills/normativa/references` (3); `Bash` `cd /s/claude/skills/normativa/references && cat "redaccion_en_una_fecha.md"` (4); `Bash` `kitlegal boe articulo BOE-A-2017-12902 a1-30 --fecha 2019-06-01 --json` (5); respuesta |
  | `normativa-producto/herramienta/04-lgt-prescripcion-01` | `Skill` (1); `Bash` `cd /s/claude/skills/normativa/references && cat con_herramientas.md && head -n 40 ./normas.md` (2); `mcp__kitlegal__boe_articulo` (3); respuesta |
  | `normativa-remisiones-en-orden/herramienta/05-trlrhl-impuestos-municipales-01` | `Skill` (1); `Read` de `/s/claude/skills/normativa/references/con_herramientas.md` y de `…/normas.md` (2); `mcp__kitlegal__boe_articulo` (3); respuesta |

- **`plazos-y-territorio/h41/`**, con la lista de las cuatro referencias: las cinco sesiones del ejemplo de
  contracts/tabla.md §3, una por variante:

  | Sesión | Lo que hace |
  |---|---|
  | `plazos-y-territorio-control/orden/02-territorio-municipio-no-cubierto-01` | `Skill` (1); `Read` de `/s/claude/skills/plazos-y-territorio/references/municipios.md` (2); `Bash` `kitlegal territorio resolver Tordesillas --json` (3); respuesta |
  | `plazos-y-territorio-producto/herramienta/05-plazo-reposicion-leganes-01` | `Skill` (1); `Bash` `cd /s/claude/skills/plazos-y-territorio/references && cat administrativo.md` (2); `Bash` `cd /s/claude/skills/plazos-y-territorio/references && cat ./calculo_de_plazos.md` (3); `mcp__kitlegal__territorio_resolver` (4); respuesta |
  | `plazos-y-territorio-sin-condicion/orden/05-plazo-reposicion-leganes-01` | `Skill` (1); `Bash` `cd /s/claude/skills/plazos-y-territorio/references/ && cat municipios.md calculo_de_plazos.md` (2); `Bash` `kitlegal territorio resolver Leganés --json` (3); respuesta |
  | `plazos-y-territorio-condicion-por-orden/herramienta/11-plazo-reposicion-sin-leer-el-precepto-01` | `Skill` (1); `mcp__kitlegal__territorio_resolver` (2); `Bash` `cd /s/claude/skills/plazos-y-territorio/references` (3); `Bash` `cat calculo_de_plazos.md` (4); respuesta |
  | `plazos-y-territorio-lista-con-verbo/orden/10-plazo-reposicion-anterior-a-la-lpac-01` | `Skill` (1); `Bash` `cd /s/claude/skills/plazos-y-territorio/references && cat municipios.md.bak calculo_de_plazos.md` (2); `Bash` `kitlegal territorio resolver Leganés --json` (3); respuesta |

- **`jurisprudencia/h41/`**, con la lista de las dos: `jurisprudencia-control/herramienta/01-existe-con-numero-y-fecha-01`:
  `Skill` `jurisprudencia` (1), `mcp__kitlegal__cita_preparar` (2), respuesta; y
  `jurisprudencia-producto/orden/04-documento-pegado-01`: `Skill` (1), `Bash` `cd
  /s/claude/skills/jurisprudencia/references && cat 'documento_aportado.md'` (2), `Bash` `kitlegal cita cotejar
  --documento "Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144" --json` (3), respuesta.
- **`mezcla/h41/`**, sin lista: `jurisprudencia-control/orden/01-existe-con-numero-y-fecha-01` (una sesión terminada
  cualquiera), que es la primera variante en orden de nombre y hace de `jurisprudencia` la familia de la tabla, y
  `normativa-control/orden/04-lgt-prescripcion-01`, que es de otra familia.
- **`normativa-tabla.md`**, **`plazos-y-territorio-tabla.md`** y **`jurisprudencia-tabla.md`**: la tabla y el detalle de
  sus sesiones según contracts/tabla.md §2 y §3. En la de `normativa`, la columna `nueva.md` va vacía en las filas de
  `normativa-control` y `normativa-remisiones-en-orden` y a «0 de 1» en las de `normativa-producto`; la sesión de modo
  orden abre `redaccion_en_una_fecha.md` en el turno 4 y no `normas.md`; la de herramienta, `con_herramientas.md` y
  `normas.md` en el turno 2. La de `plazos-y-territorio` es la del ejemplo de contracts/tabla.md §3.

Cubre: la lectura con `cd` que cuenta (con comillas dobles y simples, con `./`, con `head`, con barra final, dos
referencias en una orden) y la que no (el `cd` en otra orden, el nombre antes del `cd`, un nombre con sufijo); la fila de
cada una de las cinco variantes de `plazos-y-territorio`; la columna de una referencia de la lista del producto y su
celda vacía en las variantes que no la tienen; «Cumplen la regla» solo en `normativa`; la familia de la primera
variante; y el defecto de un directorio de otra familia.

**Antes de la implementación falla por una aserción**: `laboratorio tabla normativa` sale con 1 (`…/normativa-producto
no es ninguna variante de normativa`) y sin tabla, y `cmp normativa.out normativa-tabla.md` falla.

## 2. `aceptacion/skills-instaladas.txtar` — historias 1, 2 y 3 (FR-001 a FR-006, FR-010, FR-025; SC-001, SC-004, SC-005, SC-009)

Lo que el binario instala, con `exec kitlegal skills install --dir destino --json`, y después, sobre lo instalado:

```text
grep '## Protocolo\s+Antes de la primera orden o llamada, lee \[con_herramientas\]\(references/con_herramientas\.md\):' destino/normativa/SKILL.md
grep '### 1\. Identificar la norma\s+Lee \[normas\]\(references/normas\.md\) y toma el `BOE-A-…` de su tabla; si no está, lee\s+\[buscar_la_norma\]\(references/buscar_la_norma\.md\)' destino/normativa/SKILL.md
grep 'Con una fecha en la pregunta, lee \[redaccion_en_una_fecha\]\(references/redaccion_en_una_fecha\.md\) antes de leer el\s+bloque' destino/normativa/SKILL.md
grep 'Por cada entrada `version-obsoleta`, o si se pregunta qué cambió, lee\s+\[redaccion_modificada\]\(references/redaccion_modificada\.md\)' destino/normativa/SKILL.md
grep 'Con avisos o un id que termina en punto, lee \[como_se_cita\]\(references/como_se_cita\.md\)' destino/normativa/SKILL.md
grep 'lee antes \[varios_bloques\]\(references/varios_bloques\.md\)' destino/normativa/SKILL.md
grep 'lee \[lecturas_que_fallan\]\(references/lecturas_que_fallan\.md\) antes de seguir' destino/normativa/SKILL.md
! grep 'tras leer \[con_herramientas\]' destino/normativa/SKILL.md
! grep 'las dudas, en \[como_se_cita\]' destino/normativa/SKILL.md
! grep 'Nada de otra conversación' destino/normativa/SKILL.md
! grep 'Si una lectura falla, di qué bloques faltan y por qué\.' destino/normativa/SKILL.md
grep '\n- \*\*Nada de otra conversación\*\*, ni para afirmar, confirmar o desmentir: de ellas solo va la\s+línea `⚠ REDACCIÓN MODIFICADA:` de cada bloque que la trae\.' destino/normativa/references/redaccion_modificada.md
grep '^Si una lectura falla, di qué bloques faltan y por qué\.$' destino/normativa/references/lecturas_que_fallan.md
grep '(?s)Antes de pedir una u otra, lee\s+\[references/consulta_preparada\.md\]\(references/consulta_preparada\.md\).*\| Cómo se dio \| Qué se pide \|' destino/jurisprudencia/SKILL.md
grep '### 3\. Con el documento delante: cotejar su ficha\s+Con cada documento, antes del cotejo, lee \[references/documento_aportado\.md\]' destino/jurisprudencia/SKILL.md
! grep 'Antes de dar la consulta' destino/jurisprudencia/SKILL.md
grep '^- \[municipios\.md\]\(references/municipios\.md\): al presentar un territorio, si falta el municipio o no sale con 0, y$' destino/plazos-y-territorio/SKILL.md
grep '^- \[calculo_de_plazos\.md\]\(references/calculo_de_plazos\.md\): en los municipios que cuentan, un plazo sin calcular y una$' destino/plazos-y-territorio/SKILL.md
grep '^1\. \*\*El municipio\*\*, de la conversación\. Si no lo dice, \*\*pregúntalo\*\* y espera: no lo supongas' destino/plazos-y-territorio/SKILL.md
! grep '[Ll]ee \[(municipios|calculo_de_plazos)\.md\]' destino/plazos-y-territorio/SKILL.md
! exists destino/normativa-control destino/jurisprudencia-control destino/plazos-y-territorio-control
! exists destino/plazos-y-territorio-sin-condicion destino/plazos-y-territorio-condicion-por-orden destino/plazos-y-territorio-lista-con-verbo
```

Cubre, en lo que instala `kitlegal`: cada orden de leer de contracts/remisiones.md §2 y §4, con su sitio donde lo fija
el texto que la rodea —la de `con_herramientas.md`, lo primero del «Protocolo»; la de `normas.md`, lo primero del paso
1; la de `consulta_preparada.md`, antes de la tabla de lo que se pide; la de `documento_aportado.md`, lo primero del paso
3—; que no queda ninguna de las remisiones que el ADR 0051 no cuenta como órdenes; que lo movido está en su referencia y
no en el cuerpo; que `plazos-y-territorio` conserva la lista de control y «Si no lo dice, **pregúntalo** y espera» en el
paso 1, sin ninguna de las órdenes de las variantes; y que ninguna variante del laboratorio se instala. El tope de
11 000 caracteres, la regla y la lista de las remisiones, el cálculo desde control y la igualdad byte a byte de
`plazos-y-territorio` los comprueban `TestSkillsDelRepositorio`, `TestRemisionesDelRepositorio` y
`TestSkillsDelProductoDesdeControl`.

**Antes de la implementación falla por una aserción**: el primer `grep` no encuentra la orden en el `SKILL.md` de hoy.

## 3. Lo que no cubren

La historia 4 (el arnés) abre sesiones con modelo y con la credencial de quien lo lanza: su aceptación es
`TestArnesDeLasFamilias` con el sustituto de `claude`, en `make ci` (contracts/arnes.md §5). La historia 6 la hace la
persona antes de fusionar (FR-052, FR-053). Las suites congeladas de H39 y H40 siguen como están (FR-042).
