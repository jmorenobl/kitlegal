# Contrato: el guion de aceptación de H41.1 y el bloque de H41 que cambia

FR-001 a FR-003, FR-013; SC-001. Research V5, V6, V9, V10, D6, D7.

## 1. `aceptacion/skills-instaladas.txtar` (historias 1 y 2)

Lo escribe la primera tarea, `[aceptacion]`, en `specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar`,
y queda congelado; el workflow lo activa como `internal/app/testdata/script/h41-1-skills-instaladas.txtar`, que corre
`TestEntregaDelHito` en `make ci`. Hoy falla por una aserción en su línea 21 (research V6); con la forma aplicada, pasa
(V5). Su contenido, byte a byte:

```text
# skills-instaladas.txtar
# Lo que instala `kitlegal skills install` de plazos-y-territorio con la forma
# sin condición de H41 y el movimiento M1 (contracts/aceptacion.md de H41.1): la
# orden de leer municipios.md, en un párrafo propio al principio del
# «Protocolo», delante de «Territorio: del 1 al 4…»; la de calculo_de_plazos.md,
# al final del paso 7, antes del 8; la lista de «Referencias» sin sus dos
# viñetas, terminada en la de jerarquia_normativa.md; «Si no lo dice,
# **pregúntalo** y espera» en el paso 1; y la frase del 4 o 5 fuera del paso 10
# y al final de calculo_de_plazos.md, detrás de una línea en blanco.
#
# Cubre: US1, US2 · FR-001, FR-002, FR-003 · SC-001
#
# Que los ficheros sean, byte a byte, la variante con M1, el tope de 11 000
# caracteres y las remisiones del plan los comprueban
# TestSkillsDelProductoDesdeControl, TestSkillsDelRepositorio,
# TestVerticalDeEnsayo y TestRemisionesDelRepositorio.

exec kitlegal skills install --dir destino --json

# --- Las dos órdenes de leer, cada una en su sitio y una vez (US1, 1 y 2) ---
grep '^## Protocolo\n\nAntes de la primera orden o llamada, lee \[municipios\.md\]\(references/municipios\.md\): qué hacer si falta el municipio o\nno sale con 0, qué trae `data`, la plantilla del territorio y qué dice su cobertura\.\n\nTerritorio: del 1 al 4, y el 5 y el 6 si va más allá\.' destino/plazos-y-territorio/SKILL.md
grep 'Un acto presunto no tiene fecha: ningún plazo\. Antes del 8, lee\n   \[calculo_de_plazos\.md\]\(references/calculo_de_plazos\.md\): los municipios que cuentan, el plazo sin calcular y la\n   lectura en la fecha de inicio\.\n8\. \*\*Situar el plazo\*\*' destino/plazos-y-territorio/SKILL.md
grep -count=1 '\]\(references/municipios\.md\)' destino/plazos-y-territorio/SKILL.md
grep -count=1 '\]\(references/calculo_de_plazos\.md\)' destino/plazos-y-territorio/SKILL.md

# --- «Referencias» sin las dos viñetas, terminada en la de
# jerarquia_normativa.md (US1, 3) ---
! grep '^- \[(municipios|calculo_de_plazos)\.md\]' destino/plazos-y-territorio/SKILL.md
grep '^<!-- fin de los paquetes -->\n- \[jerarquia_normativa\.md\]\(references/jerarquia_normativa\.md\): para el nivel que regula algo, con el autonómico de la\n  `comunidad` de `data`, y la norma que prevalece\.\n\n## Protocolo$' destino/plazos-y-territorio/SKILL.md

# --- La pregunta del paso 1 (US1, 4) ---
grep '^1\. \*\*El municipio\*\*, de la conversación\. Si no lo dice, \*\*pregúntalo\*\* y espera: no lo supongas' destino/plazos-y-territorio/SKILL.md

# --- La frase del 4 o 5: fuera del paso 10 y al final de
# calculo_de_plazos.md (US2, 1 y 2) ---
! grep 'Si\s+sale\s+con\s+4\s+o\s+5,\s+\*\*`⚠\s+PLAZO:`\s+se\s+mantiene\*\*' destino/plazos-y-territorio/SKILL.md
grep 'y citarlo con\n    `⚠ REDACCIÓN A FECHA:`\.\n11\. \*\*Responder\*\*' destino/plazos-y-territorio/SKILL.md
grep 'que no has leído\. Dilo así\.\n\nSi sale con 4 o 5, \*\*`⚠ PLAZO:` se mantiene\*\*: di aparte que no has podido leerlo, sin afirmar qué dice\.\n\z' destino/plazos-y-territorio/references/calculo_de_plazos.md
```

Cómo se lee cada `grep` (research V9): el patrón se compila con `(?m)` delante y se busca en el fichero entero; `\n` es
un salto de línea, `\s` cualquier blanco y `\z` el final del fichero; `-count=1` exige exactamente una aparición. Que el
escenario 3 de la historia 2 (las otras tres referencias byte a byte) y la región de los paquetes (la genera `make
skills-sync`) no estén en el guion es deliberado: los comparan `TestSkillsDelProductoDesdeControl` y
`TestSkillsDelRepositorio` (research D6).

## 2. El bloque de `plazos-y-territorio` de `h41-skills-instaladas.txtar` (FR-013)

`internal/app/testdata/script/h41-skills-instaladas.txtar` es un fichero existente de `testdata/`: lo cambia una tarea
`[datos]`, la misma que cambia la skill (research D7). Cambian sus líneas 11-13 y 17-20 de cabecera y su bloque de las
líneas 48-53; lo demás —`normativa`, `jurisprudencia`, que ninguna variante se instala y la línea «Cubre:»— queda como
está.

Las líneas 11-13 de cabecera, desde «de la consulta que se da.», pasan a:

```text
# de la consulta que se da. En plazos-y-territorio, desde H41.1, sus dos órdenes
# de leer sin condición —la de municipios.md, lo primero del «Protocolo»; la de
# calculo_de_plazos.md, al final del paso 7— y ninguna de sus dos viñetas de
# «Referencias», «Si no lo dice, **pregúntalo** y espera» en el paso 1 y la
# frase del 4 o 5 al final de calculo_de_plazos.md y no en el cuerpo. Y ninguna
# variante del laboratorio se instala.
```

Las líneas 17-20, a:

```text
# El tope de 11 000 caracteres, la regla y la lista de las remisiones y el cálculo
# de cada skill desde su control, o en plazos-y-territorio desde su forma sin
# condición, los comprueban TestSkillsDelRepositorio,
# TestRemisionesDelRepositorio y TestSkillsDelProductoDesdeControl.
```

Y el bloque de las líneas 48-53, a:

```text
# --- plazos-y-territorio, desde H41.1: sus dos órdenes de leer, cada una en su
# sitio, sin sus viñetas de «Referencias», la pregunta del paso 1 y la frase del
# 4 o 5 al final de calculo_de_plazos.md (US3) ---
grep '## Protocolo\s+Antes de la primera orden o llamada, lee \[municipios\.md\]\(references/municipios\.md\):' destino/plazos-y-territorio/SKILL.md
grep 'Un acto presunto no tiene fecha: ningún plazo\. Antes del 8, lee\s+\[calculo_de_plazos\.md\]\(references/calculo_de_plazos\.md\):' destino/plazos-y-territorio/SKILL.md
! grep '^- \[(municipios|calculo_de_plazos)\.md\]' destino/plazos-y-territorio/SKILL.md
grep '^1\. \*\*El municipio\*\*, de la conversación\. Si no lo dice, \*\*pregúntalo\*\* y espera: no lo supongas' destino/plazos-y-territorio/SKILL.md
! grep 'Si\s+sale\s+con\s+4\s+o\s+5' destino/plazos-y-territorio/SKILL.md
grep '\n\nSi sale con 4 o 5, \*\*`⚠ PLAZO:` se mantiene\*\*: di aparte que no has podido leerlo, sin afirmar qué dice\.\n\z' destino/plazos-y-territorio/references/calculo_de_plazos.md
```

(«US3» es la historia 3 de H41, como en el resto del guion.) Con la skill de H41.1, el bloque de hoy sale en rojo en su
línea 50 (research V5), y el bloque nuevo pasa; con la skill de hoy, el bloque nuevo no se cumple, porque su primer
`grep` busca una orden que hoy no está: por eso van en la misma tarea.
