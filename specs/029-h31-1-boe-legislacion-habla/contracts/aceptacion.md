# Contrato: la suite de aceptación (ADR 0018; FR-001 a FR-004, FR-006, FR-007, FR-009, FR-010; SC-003)

La escribe T001 en `specs/029-h31-1-boe-legislacion-habla/aceptacion/`, queda congelada y, al terminar el bucle, el
workflow la activa en `internal/app/testdata/script/`, donde la ejecuta `TestEntregaDelHito` en `make ci`. Tres
guiones, uno por historia de usuario con comportamiento en el `SKILL.md` (US1 a US3), que comparan la skill del
repositorio con la medida: `$KITLEGAL_SKILLS/boe-legislacion/` frente a
`$KITLEGAL_SKILLS/../evidencias/sondeos-20261026/brazo-W/skill/`. `KITLEGAL_SKILLS` es la ruta absoluta del `skills/`
del repositorio que el arnés deja en cada guion (`internal/app/e2e_test.go:356-367`), y `h23-cita-herramientas.txtar`
ya lee así `evidencias/` (línea 44).

## 1. Reglas de forma

- **Cada comparación sale con 0 e imprime su veredicto**; la aserción es la de `stdout`, `cmp` o `! grep`. Es lo que
  pide `scripts/workflow/aceptacion.sh rojo-primero`: un guion cuya salida dice `unexpected command`, `cannot parse`,
  `unknown command` o `usage: ` no cuenta como rojo por una aserción (línea 55), y un `exec` que sale con 1 imprime
  `unexpected command failure`. Visto en esta sesión con un prototipo (research V5).
- **Solo `exec awk` y `exec ls`**, además de las órdenes de testscript (`env`, `cp`, `cmp`, `stdout`, `grep`). Los
  programas awk largos van en ficheros del propio guion (`-- pasaje.awk --`), y lo que va entre comillas simples no se
  expande (`testscript/doc.go:82-84`, go-internal v1.16.0). Ningún guion escribe fuera de su directorio de trabajo ni
  usa la red.
- **Las líneas de la skill medida son las del spec** (48 a 59, 63 a 68, 78 a 86 y 88 a 100); en la rama, el tramo se
  busca por su encabezado o su primera línea, porque las líneas cambian de número. La primera viñeta del paso 4 acaba
  antes de la viñeta siguiente (`^- `), que la sigue en la skill medida y en la entrega.
- **awk de POSIX**: `-v`, `-f`, `match`/`RSTART`/`RLENGTH`, `sub`, `in` y funciones, sin extensiones de gawk ni
  anclas dentro de una alternancia. Probados con el awk de macOS (`awk version 20200816`); con el del runner, no
  (research S5).

## 2. Los tres guiones

| Guion | Historia | Comprueba | Cubre | Hoy, con la skill de `main` |
|---|---|---|---|---|
| `h31-1-responder.txtar` | US1, escenario 1 | la primera viñeta del paso 4 y el paso 6, idénticos línea a línea | FR-002, SC-003 | rojo en `stdout '^idéntico 6$'`: «distinto: 6 líneas frente a 4» |
| `h31-1-lee-lo-que-nombra.txtar` | US2, escenario 1 | el paso 3 y el paso 5, idénticos línea a línea: la regla de leer lo que se nombra va en el cuerpo | FR-002, SC-003 | rojo en `stdout '^idéntico 12$'`: «distinto: 12 líneas frente a 4» |
| `h31-1-sitio-en-el-cuerpo.txtar` | US3, escenarios 2 y 3 | frontmatter; cada línea de un bloque `text`; la tabla; los encabezados `##` y `###`; número y encabezado en negrita de cada regla; las ocho referencias, cada una igual y ninguna más; y que el cuerpo no nombra las normas de FR-009 | FR-001, FR-003, FR-004, FR-006, FR-007, FR-009, FR-010, SC-003 | rojo en `stdout '^encabezados: idéntico 14$'`: «distinto: 14 frente a 13» (en `main` no hay paso 6) |

Con `skill-md-v0.3.1.md` como `SKILL.md` y las referencias de la skill medida, los tres pasan; con la skill de `main`,
los tres fallan por esas aserciones, y un mutante que cambia el encabezado de la regla 5 o que nombra «TRLRHL» en el
cuerpo pone en rojo el tercero (research V5). Lo que no comprueban: el tamaño del cuerpo (FR-005), que es del gate de
`make ci`; y el texto reescrito (FR-006), que coteja la revisión final con la tabla de
[skill-boe-legislacion-v0-3-1.md](skill-boe-legislacion-v0-3-1.md) §2.

US4 no tiene guion: su prueba independiente son los tests de `lecturas` de H26.1 (spec, US4), y preparar la caché de
una sesión de evals es código de `internal/evals`, no del binario (research D5). US5 a US7 son el cierre, la guía y
el sondeo de una persona.

## 3. El texto validado

Los tres guiones, tal como pasaron y fallaron en el prototipo de esta sesión (copia del repositorio en `$TMPDIR`, con
`go test -count=1 -v -run '^TestEntregaDelHito$/^zz-h31-1-' ./internal/app/`). T001 los escribe con estos nombres, sin
el prefijo `zz-`.

~~~~text
# h31-1-responder.txtar
# La respuesta empieza por lo que se pregunta y se repasa antes de enviarla: la
# primera viñeta del paso 4 y el paso 6 del SKILL.md de boe-legislacion son, línea
# a línea, las líneas 63 a 68 y 88 a 100 del SKILL.md de la skill medida,
# evidencias/sondeos-20261026/brazo-W/skill/.
#
# Cubre: US1, escenario 1 · FR-002 · SC-003.
#
# pasaje.awk compara y sale siempre con 0; la aserción es la de stdout.

env MEDIDA=$KITLEGAL_SKILLS/../evidencias/sondeos-20261026/brazo-W/skill
env RAMA=$KITLEGAL_SKILLS/boe-legislacion

# --- La primera viñeta del paso 4, «Responder» (líneas 63 a 68) ---
exec awk -v desde=63 -v hasta=68 -v 'ini=^- [*][*]La respuesta empieza por lo que se pregunta' -v 'fin=^- ' -f pasaje.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^idéntico 6$'

# --- El paso 6, «Repasar antes de enviar» (líneas 88 a 100) ---
exec awk -v desde=88 -v hasta=100 -v 'ini=^### 6[.] Repasar antes de enviar$' -v 'fin=^## ' -f pasaje.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^idéntico 13$'

-- pasaje.awk --
# Las líneas desde..hasta del primer fichero, la skill medida, frente al tramo del
# segundo, la rama, que empieza en la primera línea que casa con ini y acaba antes
# de la siguiente que casa con fin, sin las líneas en blanco del final.
FNR == NR { if (FNR >= desde && FNR <= hasta) m[++nm] = $0; next }
p && $0 ~ fin { p = 0 }
visto == 0 && $0 ~ ini { p = 1; visto = 1 }
p { r[++nr] = $0 }
END {
	while (nr > 0 && r[nr] == "") nr--
	v = "idéntico " nm
	if (nm != nr) v = "distinto: " nm " líneas frente a " nr
	else for (i = 1; i <= nm; i++) if (m[i] != r[i]) { v = "distinto en la línea " i; break }
	print v
}
~~~~

`h31-1-lee-lo-que-nombra.txtar` es el mismo guion con su comentario de cabecera («el paso 3 y el paso 5 … son, línea a
línea, las líneas 48 a 59 y 78 a 86 …; la regla de leer lo que se nombra va en el cuerpo», «Cubre: US2, escenario 1 ·
FR-002 · SC-003»), el mismo `pasaje.awk` y estas dos comparaciones:

~~~~text
# --- El paso 3, «Completar» (líneas 48 a 59) ---
exec awk -v desde=48 -v hasta=59 -v 'ini=^### 3[.] Completar$' -v 'fin=^### 4[.] ' -f pasaje.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^idéntico 12$'

# --- El paso 5, «De un precepto no leído, nada» (líneas 78 a 86) ---
exec awk -v desde=78 -v hasta=86 -v 'ini=^### 5[.] De un precepto no leído, nada$' -v 'fin=^### 6[.] ' -f pasaje.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^idéntico 9$'
~~~~

~~~~text
# h31-1-sitio-en-el-cuerpo.txtar
# El cuerpo cabe sin perder lo que toda respuesta necesita. Frente a la skill
# medida, evidencias/sondeos-20261026/brazo-W/skill/, el SKILL.md de
# boe-legislacion conserva el frontmatter, cada línea de sus bloques text, la
# tabla de comandos, los encabezados ## y ### en su orden y el número y el
# encabezado en negrita de cada regla; sus referencias son las de la skill medida,
# cada una igual y ninguna más, porque ningún texto sale del cuerpo; y su cuerpo
# no nombra las normas de las evals por materia.
#
# Cubre: US3, escenarios 2 y 3 · FR-001, FR-003, FR-004, FR-006, FR-007, FR-009,
# FR-010 · SC-003.
#
# extracto.awk y formas.awk comparan y salen siempre con 0; la aserción es la de
# stdout.

env MEDIDA=$KITLEGAL_SKILLS/../evidencias/sondeos-20261026/brazo-W/skill
env RAMA=$KITLEGAL_SKILLS/boe-legislacion

# --- El frontmatter, con la descripción (FR-010) ---
exec awk -v que=frontmatter -f extracto.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^frontmatter: idéntico 15$'

# --- Cada línea de un bloque text de la skill medida es una línea del SKILL.md (FR-003) ---
exec awk -f formas.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^formas: 5 de 5$'

# --- La tabla de comandos, entre sus marcas (FR-004) ---
exec awk -v que=tabla -f extracto.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^tabla: idéntico 26$'

# --- Los encabezados ## y ###, con su texto y en su orden (FR-006) ---
exec awk -v que=encabezados -f extracto.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^encabezados: idéntico 14$'

# --- El número y el encabezado en negrita de cada regla (FR-006) ---
exec awk -v que=reglas -f extracto.awk $MEDIDA/SKILL.md $RAMA/SKILL.md
stdout '^reglas: idéntico 7$'

# --- Las referencias: las de la skill medida, cada una igual (FR-001, FR-007) ---
exec ls $RAMA/references
cp stdout referencias-de-la-rama.txt
exec ls $MEDIDA/references
cmp stdout referencias-de-la-rama.txt
cmp $RAMA/references/buscar_la_norma.md $MEDIDA/references/buscar_la_norma.md
cmp $RAMA/references/como_se_cita.md $MEDIDA/references/como_se_cita.md
cmp $RAMA/references/con_herramientas.md $MEDIDA/references/con_herramientas.md
cmp $RAMA/references/lecturas_que_fallan.md $MEDIDA/references/lecturas_que_fallan.md
cmp $RAMA/references/normas.md $MEDIDA/references/normas.md
cmp $RAMA/references/redaccion_en_una_fecha.md $MEDIDA/references/redaccion_en_una_fecha.md
cmp $RAMA/references/redaccion_modificada.md $MEDIDA/references/redaccion_modificada.md
cmp $RAMA/references/varios_bloques.md $MEDIDA/references/varios_bloques.md

# --- El cuerpo no nombra las normas de las evals por materia (FR-009) ---
exec awk 'd >= 2 { print; next } $0 == "---" { d++ }' $RAMA/SKILL.md
cp stdout cuerpo.txt
! grep 'BOE-A-2004-4214|BOE-A-2006-20764|BOE-A-2013-12887|Ley 35/2006|Real Decreto Legislativo 2/2004|Ley 19/2013|IRPF|TRLRHL|LTAIBG' cuerpo.txt

-- extracto.awk --
# Lo que pide que —frontmatter, tabla, encabezados o reglas— de cada uno de los dos
# ficheros, la skill medida y la rama, y si es lo mismo en los dos.
function guarda(linea) { if (f == 1) a[++na] = linea; else b[++nb] = linea }
FNR == 1 { f++; d = 0; t = 0; s = 0 }
que == "frontmatter" && d < 2 { guarda($0); if ($0 == "---") d++; next }
que == "tabla" && /^<!-- inicio de la tabla de comandos/ { t = 1 }
que == "tabla" && t { guarda($0) }
que == "tabla" && /^<!-- fin de la tabla de comandos -->$/ { t = 0 }
que == "encabezados" && /^###? / { guarda($0) }
que == "reglas" && /^## Reglas$/ { s = 1 }
que == "reglas" && s && match($0, /^[0-9]+\. [*][*][^*]+[*][*]/) { guarda(substr($0, RSTART, RLENGTH)) }
END {
	v = "idéntico " na
	if (na != nb) v = "distinto: " na " frente a " nb
	else for (i = 1; i <= na; i++) if (a[i] != b[i]) { v = "distinto en " i ": " b[i]; break }
	print que ": " v
}
-- formas.awk --
# Cada línea de un bloque text del primer fichero, la skill medida, sin los blancos
# de los extremos, es una línea del segundo, la rama, también sin ellos.
FNR == 1 { f++; t = 0 }
{ l = $0; sub(/^[ \t]+/, "", l); sub(/[ \t]+$/, "", l) }
f == 1 && l == "```text" { t = 1; next }
f == 1 && t && l == "```" { t = 0; next }
f == 1 && t { formas[++nf] = l; next }
f == 2 { lineas[l] = 1 }
END {
	estan = 0
	for (i = 1; i <= nf; i++) if (formas[i] in lineas) estan++; else print "falta: " formas[i]
	print "formas: " estan " de " nf
}
~~~~

La comprobación de FR-009 mira el cuerpo entero, que en `main` no nombra ninguna de esas normas (`grep -c`, 0, research
V7): es más estricta que el diff con `main` del criterio y equivalente a él mientras eso siga así. Las referencias no
se miran aquí porque son las de la skill medida (lo comprueba el mismo guion), y las dos que difieren de `main` no
nombran ninguna (`grep -l`, research V7).

## 4. Lo que fija la suite y deja de valer

Fija los pasajes medidos, las formas, la tabla, los encabezados y las reglas a los de la skill medida, como pide SC-003.
Un hito posterior que cambie alguno de ellos —H32 reparte las skills por capacidad— cambia lo que la suite describe y
la retira o la sustituye en su propio diff, como cualquier guion activado cuyo comportamiento cambia (research D4).
