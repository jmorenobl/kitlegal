# Contrato: los textos de `skills/` y lo que los fija (FR-010 a FR-016)

Todo esto va en una sola tarea `[datos]` (FR-020; research D9), detrás de la de la tabla.

## 1. Los textos (FR-010 a FR-013)

`skills/` queda, byte a byte, como en `f89093dd`: `git checkout f89093dd -- skills/` (research D5) y, después, `git diff
--quiet f89093dd -- skills/`. Cambian seis ficheros, y solo ellos: `skills/normativa/SKILL.md`,
`skills/plazos-y-territorio/SKILL.md`, `skills/plazos-y-territorio/references/municipios.md`,
`skills/jurisprudencia/SKILL.md`, `skills/jurisprudencia/references/consulta_preparada.md` y
`skills/jurisprudencia/references/documento_aportado.md` (`git diff --stat main f89093dd -- skills/`). Lo que hace cada
cambio es el de FR-010 a FR-012; ninguna línea `⚠`, forma fija, regla invariante, orden de leer, descripción ni
frontmatter sale ni cambia, y ninguna referencia nueva ni enlace entre referencias (FR-013).

**Lo que lee el agente, por pregunta** (research P1; las órdenes de leer de hoy, ADR 0051 y 0053; ninguna invocación
más):

| Skill | Lo que carga | Con `main` | Con el hito |
|---|---|---|---|
| `normativa` | el cuerpo, una vez | 10 750 | 9498 |
| `plazos-y-territorio` | el cuerpo y `municipios.md`, antes de la primera orden, una vez | 10 686 + 3873 = 14 559 | 9577 + 4666 = 14 243 |
| `jurisprudencia` | el cuerpo; `consulta_preparada.md` antes de `cita_preparar`; `documento_aportado.md` antes de cada cotejo (k documentos) | 10 842 + 1997 + 2756·k | 9473 + 2698 + 3208·k |

En `jurisprudencia` la diferencia es 452·k − 668: con un documento (k = 1), 216 caracteres menos; con dos, 236 más, y
452 más por cada documento siguiente. Lo acotan los documentos que la persona trae a esa pregunta, no lo acumulado. Nada de esto
lee la caché ni el grafo: con meses de uso diario, cientos de normas y miles de bloques consultados, es igual. El cuerpo
solo crece con los paquetes, una línea de 52 + 3n cada uno, hasta los 1400 que deja FR-005.

## 2. Las declaraciones del producto (FR-014)

En `declaracionesDelProducto` (`internal/evals/laboratorio_familias_test.go`), 15 entradas nuevas, en el orden del
fichero y con los bytes de `f89093dd` (las del prototipo, `git diff main f89093dd --
internal/evals/laboratorio_familias_test.go`, son la referencia de sus bytes: el texto de cada `antes` y `despues` es
el de `skills/`):

| Skill | Dónde | Qué texto |
|---|---|---|
| `normativa` | cuerpo | «desde el `PATH` o, si la tienes, con su herramienta» → «desde el `PATH` o con su herramienta» |
| `jurisprudencia` | cuerpo | la orden de leer `consulta_preparada.md`, con «qué se pide según cómo se dio la referencia» y, detrás, «Un número sin fecha no se prepara de ninguna forma, tampoco como ROJ: lleva la línea y se pide la fecha.» |
| `jurisprudencia` | cuerpo | fuera la tabla «Cómo se dio / Qué se pide» y la viñeta del número de resolución |
| `jurisprudencia` | cuerpo | fuera la viñeta «Si se había pedido una sentencia concreta, da además la referencia…» |
| `jurisprudencia` | cuerpo | fuera la viñeta «Las sentencias que el documento menciona no se citan…» |
| `jurisprudencia` | cuerpo | en «Comandos», fuera el inciso «—también cuando el documento no es el pedido, que es un hallazgo y no un fallo—» |
| `jurisprudencia` | `consulta_preparada.md` | el primer párrafo dice que trae lo que se pide según cómo se dio la referencia |
| `jurisprudencia` | `consulta_preparada.md` | al final, el apartado `## Lo que se pide` con la tabla y la viñeta |
| `jurisprudencia` | `documento_aportado.md` | en «Lo que devuelve», «La orden sale con 0 también cuando el documento no es el pedido, que es un hallazgo y no un fallo.» |
| `jurisprudencia` | `documento_aportado.md` | al final, el apartado `## La referencia pedida y las sentencias que menciona` con las dos viñetas |
| `plazos-y-territorio` | cuerpo | la orden de leer `municipios.md`, con «cómo se pide cada operación» |
| `plazos-y-territorio` | cuerpo | «Cada dato como lo da `data`; un boletín…; la cobertura, …, y detrás puedes explicarla:» → «La cobertura, con la clave y el valor exactos, una línea por aspecto:» |
| `plazos-y-territorio` | cuerpo | el párrafo «Con la herramienta de su fila si…» → «`kitlegal` se invoca desde el `PATH`, y cada herramienta, por el nombre de su fila.» |
| `plazos-y-territorio` | `municipios.md` | el primer párrafo, sin remitir al cuerpo, con lo que trae |
| `plazos-y-territorio` | `municipios.md` | al final, `## Cómo se presenta cada dato` y `## Herramienta u orden` |

Con cualquiera de ellos deshecho en `skills/`, `TestSkillsDelProductoDesdeControl` sale en rojo (SC-004; el registro de
mutantes, uno por entrada).

## 3. Las remisiones del plan (FR-015)

En `remisionesDelPlan` (`internal/skills/remisiones_test.go`), las dos frases de `main` pasan a:

- `jurisprudencia`: «Antes de pedir una u otra, lee [references/consulta_preparada.md](references/consulta_preparada.md):
  qué se pide según cómo se dio la referencia, la orden de cada forma, cómo se dan las casillas y qué hacer con el
  equivalente que devuelva la operación.»
- `plazos-y-territorio`: «Antes de la primera orden o llamada, lee [municipios.md](references/municipios.md): cómo se pide
  cada operación, qué hacer si falta el municipio o no sale con 0, qué trae `data`, la plantilla del territorio y qué
  dice su cobertura.»

Con la frase de `main` en cualquiera de las dos, `TestRemisionesDelRepositorio` sale en rojo (SC-005).

## 4. Los guiones de H41 (FR-016)

- `internal/app/testdata/script/h41-skills-instaladas.txtar`: la línea

  ```text
  grep '(?s)Antes de pedir una u otra, lee\s+\[references/consulta_preparada\.md\]\(references/consulta_preparada\.md\).*\| Cómo se dio \| Qué se pide \|' destino/jurisprudencia/SKILL.md
  ```

  pasa a estas dos:

  ```text
  grep 'Antes de pedir una u otra, lee\s+\[references/consulta_preparada\.md\]\(references/consulta_preparada\.md\): qué se pide\s+según cómo se dio la referencia' destino/jurisprudencia/SKILL.md
  grep '^\| Cómo se dio \| Qué se pide \|$' destino/jurisprudencia/references/consulta_preparada.md
  ```

- `internal/app/testdata/script/h41-1-skills-instaladas.txtar` y
  `specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar`, la misma línea en los dos:

  ```text
  grep '^## Protocolo\n\nAntes de la primera orden o llamada, lee \[municipios\.md\]\(references/municipios\.md\): cómo se pide cada\noperación, qué hacer si falta el municipio o no sale con 0, qué trae `data`, la plantilla del territorio y qué dice su\ncobertura\.\n\nTerritorio: del 1 al 4, y el 5 y el 6 si va más allá\.' destino/plazos-y-territorio/SKILL.md
  ```

  en lugar de la de la orden de `main`; después, `cmp` de los dos ficheros, sin diferencias.

La copia congelada de H41 (`specs/034-h41-las-skills-del/aceptacion/skills-instaladas.txtar`) no se toca (spec,
«Assumptions»).
