# Contrato: el guion de aceptación

FR-025 (historia 5); cubre también lo instalado de FR-004, FR-007 y FR-010 y SC-001, SC-002, SC-003 y SC-005. Lo
escribe la primera tarea en `specs/039-h33-cobertura-en-positivo/aceptacion/cobertura-instalada.txtar`, sin código de
producto; el workflow comprueba que hoy falla por una aserción (la primera: hoy ningún `SKILL.md` lleva la plantilla),
lo congela y, al terminar el bucle, lo activa como `internal/app/testdata/script/h33-cobertura-instalada.txtar`, que
ejecuta `TestEntregaDelHito` en `make ci`.

## 1. Qué afirma y qué no

Afirma, sobre lo que instala `kitlegal skills install`, solo lo que el spec fija byte a byte: la plantilla y la marca
(FR-004), la región de lo cubierto con los cuatro alcances de FR-001 (FR-007), la fila de los tributos (historia 3,
escenario 1), que cada fila de la tabla es la plantilla con su procedimiento y uno de los dos motivos de FR-005, y el
tope de 10 860 caracteres del cuerpo de `plazos-y-territorio` (FR-010). No afirma qué filas llevan norma más allá de
la cuenta (2, las entradas de `regulados_aparte` que traen `norma`, que no cambian), ni cortes de línea de la prosa, ni
nada de lo que comprueban los tests de contracts/controles.md.

## 2. El guion

```text
# cobertura-instalada.txtar
# Lo que instala `kitlegal skills install` de la cobertura en positivo (contracts/aceptacion.md de H33):
# consulta-juridica y plazos-y-territorio llevan la plantilla única ⚠ NO CUBIERTO:, la misma línea en las dos, y
# solo consulta-juridica la marca ⚠ NO VERIFICADO:, en la línea siguiente del mismo bloque text; ninguna línea de
# la plantilla dice «procedimiento común»; la región de lo cubierto de consulta-juridica tiene una línea por cada
# una de las cuatro filas de cubierto de data/cobertura.yaml; cada fila de la tabla cobertura_por_materia de
# administrativo.md es la plantilla con su procedimiento y uno de los dos motivos, y la de los tributos es la del
# spec byte a byte; y el cuerpo de plazos-y-territorio, lo que sigue al --- que cierra el frontmatter, tiene como
# mucho 10 860 caracteres. El tope lo mide una expresión regular de Go, que cuenta runas: diez (?s:.){1000} y un
# (?s:.){861} seguidos son 10 861 caracteres, y el guion exige que no los haya. filas.awk compara bytes y sale
# siempre con 0: el veredicto es el de stdout.
#
# Cubre: US5 · FR-004, FR-007, FR-010, FR-025 · SC-001, SC-002, SC-003, SC-005
#
# Que lo regenerado sea lo que da make skills-sync, el texto escrito a mano de cada skill y los defectos de la
# sincronía lo comprueban TestSkillsDelRepositorio, TestRegionDeLoCubierto, TestRegionDeLaCobertura,
# TestSkillsDelProductoDesdeControl y TestConsultaJuridicaDesdeLaBase.

exec kitlegal skills install --dir destino --json

# --- La plantilla, sola en su bloque en plazos-y-territorio (US5, 1 y 4) ---
grep '\n```text\n⚠ NO CUBIERTO: <asunto>: <motivo>\. kitlegal no da sus plazos, órganos, vías ni escritos\.\n```\n' destino/plazos-y-territorio/SKILL.md
grep -count=1 '^⚠ NO CUBIERTO: ' destino/plazos-y-territorio/SKILL.md
! grep '⚠ NO VERIFICADO:' destino/plazos-y-territorio/SKILL.md

# --- La misma plantilla y, detrás, la marca, en consulta-juridica (US5, 1 y 4) ---
grep '\n```text\n⚠ NO CUBIERTO: <asunto>: <motivo>\. kitlegal no da sus plazos, órganos, vías ni escritos\.\n⚠ NO VERIFICADO: <lo encontrado>: <forma legible> \[<identificador>, bloque <id>\]; kitlegal no ha comprobado que se aplique a este asunto\.\n```\n' destino/consulta-juridica/SKILL.md
grep -count=1 '^⚠ NO CUBIERTO: ' destino/consulta-juridica/SKILL.md

# --- Ninguna línea de la plantilla dice «procedimiento común» (US5, 4) ---
! grep '^⚠ NO CUBIERTO: .*procedimiento común' destino/consulta-juridica/SKILL.md
! grep '^⚠ NO CUBIERTO: .*procedimiento común' destino/plazos-y-territorio/SKILL.md

# --- La región de lo cubierto, una línea por fila de cubierto (US5, 2) ---
grep '\n<!-- lo que cubre la base: lo genera make skills-sync -->\n- \*\*Normativa\*\*: leer y citar cualquier norma consolidada del BOE, estatal o autonómica, de cualquier materia y en cualquier fecha\.\n- \*\*Jurisprudencia\*\*: preparar la consulta exacta de una sentencia, que la persona busca con su navegador, y cotejar el documento que trae\. No consulta ningún buscador ni dice si una sentencia existe\.\n- \*\*Plazos\*\*: solo los del procedimiento administrativo común, los del acceso a la información pública y la interposición del recurso contencioso-administrativo, con los festivos de los municipios que cuentan\. Lo que la ley regula aparte de ese procedimiento lo declara `plazos-y-territorio` con su línea\.\n- \*\*Territorio\*\*: la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un municipio\.\n<!-- fin de lo que cubre la base -->\n' destino/consulta-juridica/SKILL.md

# --- Cada fila de la tabla, la plantilla con su motivo; la de los tributos, byte a byte (US5, 3) ---
exec awk -f filas.awk destino/plazos-y-territorio/references/administrativo.md
stdout '^filas: 9 de 9, 2 con norma$'
grep '^\| aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa \| ⚠ NO CUBIERTO: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa: se rige por Ley 58/2003, General Tributaria \(BOE-A-2003-23186\), no por el procedimiento común \(DA 1ª\.2\.a LPAC \[BOE-A-2015-10565, bloque daprimera\]\)\. kitlegal no da sus plazos, órganos, vías ni escritos\. \|$' destino/plazos-y-territorio/references/administrativo.md

# --- El cuerpo de plazos-y-territorio, como mucho 10 860 caracteres (FR-010) ---
! grep '\A---\n(?s:.*?)\n---\n(?s:.){1000}(?s:.){1000}(?s:.){1000}(?s:.){1000}(?s:.){1000}(?s:.){1000}(?s:.){1000}(?s:.){1000}(?s:.){1000}(?s:.){1000}(?s:.){861}' destino/plazos-y-territorio/SKILL.md

-- filas.awk --
# filas.awk <administrativo.md>: de las filas de la región cobertura_por_materia,
# sin su encabezado, cuántas son la plantilla con el procedimiento de su primera
# celda y uno de los dos motivos, de cuántas, y cuántas llevan el de la fila con
# norma. Solo compara cadenas: el mismo resultado con bytes o con caracteres.
BEGIN {
	cola = ". kitlegal no da sus plazos, órganos, vías ni escritos."
	cita = "\\(.+ \\[BOE-A-[0-9]+-[0-9]+, bloque [A-Za-z0-9.-]+\\]\\)$"
	conNorma = "^se rige por .+ \\(BOE-A-[0-9]+-[0-9]+\\), no por el procedimiento común " cita
	sinNorma = "^lo regula aparte la ley, no el procedimiento común " cita
}
index($0, "<!-- inicio de cobertura_por_materia:") == 1 { dentro = 1; next }
$0 == "<!-- fin de cobertura_por_materia -->" { dentro = 0; next }
dentro && /^\| / && $0 != "| Procedimiento | Línea que da la respuesta |" {
	filas++
	fila = substr($0, 3, length($0) - 4)
	corte = index(fila, " | ")
	if (corte == 0) next
	principio = "⚠ NO CUBIERTO: " substr(fila, 1, corte - 1) ": "
	linea = substr(fila, corte + 3)
	if (index(linea, principio) != 1 || length(linea) <= length(principio) + length(cola)) next
	if (substr(linea, length(linea) - length(cola) + 1) != cola) next
	motivo = substr(linea, length(principio) + 1, length(linea) - length(principio) - length(cola))
	if (motivo ~ conNorma) { bien++; deNorma++ } else if (motivo ~ sinNorma) bien++
}
END { printf "filas: %d de %d, %d con norma\n", bien, filas, deNorma }
```

## 3. Por qué así

- Los `grep` llevan los bytes literales del spec: la plantilla y la marca (FR-004), los alcances (FR-001) y la fila de
  los tributos. Con `(?m)` delante, como los compila testscript (research V2), `\n` cruza líneas y `^` y `$` son de
  línea: la región, entre sus dos marcas, tiene esas cuatro líneas y ninguna más. La misma plantilla literal en los dos
  `SKILL.md` es «la misma línea en los dos»; `-count=1` exige que sea la única que empieza por la etiqueta.
- El tope del cuerpo: `(?s:.)` es una runa, y una repetición de Go no pasa de 1000 ni anida por encima de 1000 (research
  V3); diez seguidas y una de 861 no anidan. Probado sobre el `SKILL.md` de hoy: casa con 10 882 y no con 10 883.
- `filas.awk`: lo que un `grep` no puede comprobar sin referencias hacia atrás, que la línea repite el procedimiento de
  su primera celda. Usa `index`, `substr`, `length` y una expresión sobre cadenas, que dan lo mismo contando bytes o
  caracteres porque todas las posiciones salen de las mismas funciones: con `gawk`, `mawk` y `busybox awk`, en UTF-8 y
  en C, da `filas: 9 de 9, 2 con norma` sobre la tabla que entrega el hito y `filas: 0 de 9, 0 con norma` sobre la de
  hoy (research V7b).
- Hoy falla la primera aserción, el `grep` de la plantilla de `plazos-y-territorio`; sobre los textos de la entrega
  (los de `h33-prototipo`), todas pasan (research V7b).
