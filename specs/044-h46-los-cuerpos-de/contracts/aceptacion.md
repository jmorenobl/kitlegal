# Contrato: la suite de aceptación (FR-021; SC-006)

Dos guiones que escribe T001 en `specs/044-h46-los-cuerpos-de/aceptacion/` y que el workflow activa como
`internal/app/testdata/script/h46-tabla-de-tres-columnas.txtar` y `h46-textos-en-su-referencia.txtar`
(`scripts/workflow/aceptacion.sh activar`; research V4). Los dos comprueban lo que instala `kitlegal skills install
--dir destino --json` desde el binario de la rama, sin red. El defecto de FR-005 no lo da ninguna orden de `kitlegal`, y
no es de la suite (FR-021): lo fijan `casosDePresupuestos` y sus mutantes.

Medidos en esta sesión, con el contenido de abajo, en un clon (research P1): los dos `PASS` sobre los `skills/` de
`f89093dd`; los dos `FAIL` por una aserción sobre los de `main` (`tabla-de-tres-columnas.txtar:20` y
`textos-en-su-referencia.txtar:25`, `no match for …`), que es el «rojo primero»; y, con cada aserción invertida salvo la
del bloque de la cobertura, que vale en los dos, `PASS` sobre los de `main`: cada una discrimina.

Cómo se acota cada búsqueda (research D8): la región de la tabla, entre sus dos marcas, con `(?s)` para que `.*?` cruce
líneas y `[^\n]` dentro de cada alternativa; un apartado, desde su `## ` sin otro `#` por medio (`[^#]*`: ninguno de
estos apartados lleva uno); «termina con», `\z`; el primer párrafo de una referencia, lo que sigue a su título `# ` hasta
la primera línea en blanco (`(?:[^\n]|\n[^\n])*`); y cada frase citada, con `\s+` entre sus palabras. El `grep` de
testscript compila `(?m)` delante de cada patrón y lo busca en el fichero entero (research V2).

## 1. `tabla-de-tres-columnas.txtar` (US1.1 · FR-001, FR-021 · SC-006)

```text
# tabla-de-tres-columnas.txtar
# Lo que instala `kitlegal skills install` de las tres skills de capacidad: la
# tabla de comandos de cada una, entre sus marcas, con la cabecera de tres
# columnas «Orden», «Herramienta» y «Qué hace» y su separador, y sin una cabecera
# de más columnas, sin un separador de cuatro, sin una fila con una cuarta celda,
# sin «Qué devuelve en», sin la línea del sobre y sin la de las banderas comunes
# (contracts/aceptacion.md §1 de H46). Cada negativa busca solo entre las dos
# marcas de la tabla, la región que nombra el spec: `(?s)` deja que `.*?` cruce
# líneas, y `[^\n]` no lo hace.
#
# Cubre: US1.1 · FR-001, FR-021 · SC-006
#
# Que los bytes sean los de RenderizarTabla y que los mismos errores sigan
# saliendo lo comprueban TestRenderizarTabla, TestRegenerarYComparar,
# TestCuerpoDeLaSkill y la deriva de TestSkillsDelRepositorio (FR-002 a FR-004).

exec kitlegal skills install --dir destino --json

# --- La cabecera de tres columnas, con su separador y una fila detrás (US1.1) ---
grep '^\| Orden \| Herramienta \| Qué hace \|\n\|---\|---\|---\|\n\| `kitlegal ' destino/normativa/SKILL.md
grep '^\| Orden \| Herramienta \| Qué hace \|\n\|---\|---\|---\|\n\| `kitlegal ' destino/plazos-y-territorio/SKILL.md
grep '^\| Orden \| Herramienta \| Qué hace \|\n\|---\|---\|---\|\n\| `kitlegal ' destino/jurisprudencia/SKILL.md

# --- Nada de cuatro columnas, ni el sobre, ni las banderas, en la región (US1.1) ---
! grep '(?s)<!-- inicio de la tabla de comandos: [^\n]*\n.*?(?:^\| Orden \| Herramienta \| Qué hace \|[^\n]|^\|---\|---\|---\|-|^\| `[^\n]*` \| `[a-z_]+` \| (?:[^|\n]|\\\|)* \| |Qué devuelve en|^La orden y la herramienta de cada fila devuelven el mismo sobre|^Banderas comunes:).*?<!-- fin de la tabla de comandos -->' destino/normativa/SKILL.md
! grep '(?s)<!-- inicio de la tabla de comandos: [^\n]*\n.*?(?:^\| Orden \| Herramienta \| Qué hace \|[^\n]|^\|---\|---\|---\|-|^\| `[^\n]*` \| `[a-z_]+` \| (?:[^|\n]|\\\|)* \| |Qué devuelve en|^La orden y la herramienta de cada fila devuelven el mismo sobre|^Banderas comunes:).*?<!-- fin de la tabla de comandos -->' destino/plazos-y-territorio/SKILL.md
! grep '(?s)<!-- inicio de la tabla de comandos: [^\n]*\n.*?(?:^\| Orden \| Herramienta \| Qué hace \|[^\n]|^\|---\|---\|---\|-|^\| `[^\n]*` \| `[a-z_]+` \| (?:[^|\n]|\\\|)* \| |Qué devuelve en|^La orden y la herramienta de cada fila devuelven el mismo sobre|^Banderas comunes:).*?<!-- fin de la tabla de comandos -->' destino/jurisprudencia/SKILL.md
```

Las seis alternativas de la negativa: una cabecera con algo detrás de «Qué hace |» (de más columnas), un separador de
cuatro, una fila con una cuarta celda (tres celdas cerradas por ` | ` y algo detrás; una `\|` escapada no cierra
celda), «Qué devuelve en», la línea del sobre y la de las banderas, estas dos al principio de una línea (US1.1).

## 2. `textos-en-su-referencia.txtar` (US3.1 a US3.6 · FR-010, FR-011, FR-012, FR-021 · SC-006)

```text
# textos-en-su-referencia.txtar
# Lo que instala `kitlegal skills install` de lo que el hito mueve del cuerpo a una
# referencia que su clase de pregunta lee sin condición (contracts/aceptacion.md
# §2 de H46): en plazos-y-territorio, la herramienta o la orden y los códigos, a
# `## Herramienta u orden` de municipios.md, y cómo se escribe cada dato, a
# `## Cómo se presenta cada dato`, los dos al final y en ese orden, con las tres
# líneas de la cobertura en el cuerpo; en jurisprudencia, la tabla «Cómo se dio /
# Qué se pide» y la viñeta del número de resolución, a `## Lo que se pide` de
# consulta_preparada.md, y las dos viñetas de la referencia pedida y de las
# sentencias que menciona, a su apartado de documento_aportado.md, con el inciso
# de la salida 0 en «Lo que devuelve»; y en normativa, «o con su herramienta».
# Cada frase citada lleva `\s+` entre sus palabras: el texto se parte en líneas de
# 120 columnas, y dónde no lo fija el spec. Un apartado va de su `## ` al
# siguiente `#` (`[^#]*`; ninguno de estos lleva otro), y «al final», hasta `\z`.
#
# Cubre: US3.1-US3.6 · FR-010, FR-011, FR-012, FR-021 · SC-006
#
# Que los ficheros sean, byte a byte, los del prototipo lo comprueban
# TestSkillsDelProductoDesdeControl y TestRemisionesDelRepositorio (FR-014,
# FR-015).

exec kitlegal skills install --dir destino --json

# --- plazos-y-territorio: la herramienta o la orden (US3.1) ---
grep '`kitlegal`\s+se\s+invoca\s+desde\s+el\s+`PATH`,\s+y\s+cada\s+herramienta,\s+por\s+el\s+nombre\s+de\s+su\s+fila\.' destino/plazos-y-territorio/SKILL.md
! grep 'Con\s+la\s+herramienta\s+de\s+su\s+fila\s+si\s+entre\s+las\s+tuyas\s+hay\s+una\s+con\s+ese\s+nombre' destino/plazos-y-territorio/SKILL.md
grep '\n## Cómo se presenta cada dato\n[^#]*\n## Herramienta u orden\n[^#]*\z' destino/plazos-y-territorio/references/municipios.md
grep '\n## Herramienta u orden\n[^#]*Con\s+la\s+herramienta\s+de\s+su\s+fila' destino/plazos-y-territorio/references/municipios.md
grep '\n## Herramienta u orden\n[^#]*2\s+\(`argumentos`\)' destino/plazos-y-territorio/references/municipios.md
grep '\n## Herramienta u orden\n[^#]*3\s+\(`no-encontrado`\)' destino/plazos-y-territorio/references/municipios.md

# --- plazos-y-territorio: las tres líneas de la cobertura siguen en el cuerpo, y
# cómo se escribe cada dato va a la referencia (US3.2) ---
grep '\n## Cómo se presenta el territorio\n[^#]*```text\nboletin_autonomico: <valor>\nboletin_provincial: <valor>\ndir3: <valor>\n```\n' destino/plazos-y-territorio/SKILL.md
! grep 'un\s+boletín,\s+por\s+su\s+`codigo`\s+y\s+su\s+`nombre`\s+exactos' destino/plazos-y-territorio/SKILL.md
grep '\n## Cómo se presenta cada dato\n[^#]*un\s+boletín,\s+por\s+su\s+`codigo`\s+y\s+su\s+`nombre`\s+exactos' destino/plazos-y-territorio/references/municipios.md

# --- plazos-y-territorio: la orden de leer municipios.md y su primer párrafo (US3.3) ---
grep 'lee\s+\[municipios\.md\]\(references/municipios\.md\):\s+cómo\s+se\s+pide\s+cada\s+operación' destino/plazos-y-territorio/SKILL.md
! grep '\A# [^\n]*\n\n(?:[^\n]|\n[^\n])*Cómo\s+se\s+presenta\s+el\s+territorio' destino/plazos-y-territorio/references/municipios.md

# --- jurisprudencia: lo que se pide, en consulta_preparada.md (US3.4) ---
! grep '\| Cómo se dio \| Qué se pide \|' destino/jurisprudencia/SKILL.md
! grep 'lleva\s+el\s+\*\*número\s+de\s+resolución\*\*,\s+no\s+el\s+ROJ' destino/jurisprudencia/SKILL.md
grep 'lee\s+\[references/consulta_preparada\.md\]\(references/consulta_preparada\.md\):\s+qué\s+se\s+pide\s+según\s+cómo\s+se\s+dio\s+la\s+referencia[^.]*\.\s+Un\s+número\s+sin\s+fecha\s+no\s+se\s+prepara\s+de\s+ninguna\s+forma,\s+tampoco\s+como\s+ROJ:\s+lleva\s+la\s+línea\s+y\s+se\s+pide\s+la\s+fecha\.' destino/jurisprudencia/SKILL.md
grep '\n## Lo que se pide\n[^#]*\z' destino/jurisprudencia/references/consulta_preparada.md
grep '\n## Lo que se pide\n[^#]*\n\| Cómo se dio \| Qué se pide \|\n' destino/jurisprudencia/references/consulta_preparada.md
grep '\n## Lo que se pide\n[^#]*\n- [^#]*lleva\s+el\s+\*\*número\s+de\s+resolución\*\*,\s+no\s+el\s+ROJ' destino/jurisprudencia/references/consulta_preparada.md

# --- jurisprudencia: la referencia pedida y las sentencias que menciona, en
# documento_aportado.md (US3.5) ---
! grep 'da\s+además\s+la\s+referencia\s+en\s+la\s+forma\s+en\s+que\s+se\s+pidió' destino/jurisprudencia/SKILL.md
! grep 'Las\s+sentencias\s+que\s+el\s+documento\s+menciona\s+no\s+se\s+citan' destino/jurisprudencia/SKILL.md
! grep 'también\s+cuando\s+el\s+documento\s+no\s+es\s+el\s+pedido' destino/jurisprudencia/SKILL.md
grep '\n## La referencia pedida y las sentencias que menciona\n[^#]*\z' destino/jurisprudencia/references/documento_aportado.md
grep '\n## La referencia pedida y las sentencias que menciona\n[^#]*\n- Si\s+se\s+había\s+pedido\s+una\s+sentencia\s+concreta,\s+da\s+además\s+la\s+referencia\s+en\s+la\s+forma\s+en\s+que\s+se\s+pidió' destino/jurisprudencia/references/documento_aportado.md
grep '\n## La referencia pedida y las sentencias que menciona\n[^#]*\n- Las\s+sentencias\s+que\s+el\s+documento\s+menciona\s+no\s+se\s+citan' destino/jurisprudencia/references/documento_aportado.md
grep '\n## Lo que devuelve\n[^#]*La\s+orden\s+sale\s+con\s+0\s+también\s+cuando\s+el\s+documento\s+no\s+es\s+el\s+pedido' destino/jurisprudencia/references/documento_aportado.md

# --- normativa: «o con su herramienta» (US3.6) ---
grep 'desde\s+el\s+`PATH`\s+o\s+con\s+su\s+herramienta' destino/normativa/SKILL.md
! grep 'si\s+la\s+tienes,\s+con\s+su\s+herramienta' destino/normativa/SKILL.md
```

US3.7 (`skills/` igual a `f89093dd`) no es de la suite: es una comparación con Git, que comprueban
`TestSkillsDelProductoDesdeControl` y la deriva de `skills-check` en `make ci`, y quickstart Q4 (plan, «Controles de
umbral»).
