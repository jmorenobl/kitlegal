---
id: H46
fase: "1"
depende_de: [H45.1]
---

# H46 · Los cuerpos de las skills de capacidad, con sitio para las materias (ADR 0063)

- **Objetivo**: que una materia nueva quepa en el cuerpo de cada skill de capacidad sin tocar su protocolo. Hoy no
  caben dos: el prototipo de H42 sobre `main` (`7aaad57f`) deja `make ci` en rojo en el guion `h45-paquete-de-mas`, con
  el cuerpo de `normativa` en 11 030 caracteres de 11 000 (ADR 0046). El hito adelgaza los cuerpos de `normativa`,
  `plazos-y-territorio` y `jurisprudencia` con progressive disclosure —en el cuerpo, solo lo que toda respuesta
  necesita— y reserva en cada uno 1500 caracteres para las líneas de los paquetes, unas veinte materias, comprobados
  por `make ci`. Las respuestas no cambian: lo deciden las evals.
- **Decisiones que esta entrada aplica**: el ADR 0063 entero, que sustituye el punto 3 del ADR 0062. Siguen como están,
  y el run no los cambia ni escribe otro: el tope de 11 000 del cuerpo y los 900 de la descripción (ADR 0046, punto 2);
  la línea por paquete, con su enlace y su condición, generada como hoy; que ninguna referencia enlaza otra
  (`docs/SKILLS.md` §4); las formas de las remisiones (ADR 0051 a 0053) y sus listas del plan; los controles de
  contenido (`docs/SKILLS.md` §1); y las descripciones de las cuatro skills (ADR 0057).
- **Lo medido antes de escribirla**, el 2026-11-24, en runas, como el gate (`docs/SKILLS.md` §3), sobre `main`:

  | | `normativa` | `plazos-y-territorio` | `jurisprudencia` |
  |---|---|---|---|
  | Cuerpo | 10 750 | 10 686 | 10 842 |
  | Líneas de paquete (sin las marcas) | 0 | 94 | 0 |
  | Tabla de comandos generada | 2673 | 1498 | 1302 |
  | … de ella, la columna «Qué devuelve en `data`» | 883 | 205 | 171 |
  | … de ella, la línea del sobre y la de las banderas comunes | 318 | 318 | 318 |
  | Bloques `text` (las formas fijas) | 887 | 1142 | 152 |
  | Reglas | 1298 | 2168 | 1480 |

  Y sobre un prototipo con lo que pide «Entrega» (rama `cuerpos-prototipo`, sin propuesta de cambio): `normativa`
  9498, `plazos-y-territorio` 9388 (9294 sin la línea de lo administrativo) y `jurisprudencia` 9473; con el paquete
  fiscal de H42 encima, `make ci` en verde sin ningún `.go` de H42.
- **Entrega**:
  1. **La tabla de comandos, sin lo que da la llamada** (ADR 0063, punto 2). `RenderizarTabla`
     (`internal/skills/comandos.go`) escribe las columnas «Orden», «Herramienta» y «Qué hace», y ni la columna «Qué
     devuelve en `data`», ni la línea del sobre, ni la de las banderas comunes. `LeerDescripcionDeVerbo` sigue leyendo
     lo mismo, con los mismos defectos, y `RenderizarTabla` sigue exigiendo el mismo sobre y las mismas banderas en
     todas las filas de una tabla. Los tests de la tabla (`TestRenderizarTabla`, `TestRegenerarYComparar`,
     `TestCuerpoDeLaSkill`) fijan los bytes nuevos; `make skills-sync` rehace las tablas de las tres skills.
  2. **El sitio de las materias, comprobado** (ADR 0063, punto 1). Una regla de `skills-check`: el cuerpo de un
     `SKILL.md` con las marcas de los paquetes, sin las líneas entre ellas, tiene como mucho 9500 caracteres, con el
     defecto `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de <n> caracteres (máximo 9500)`, su mutante en
     rojo (una línea de más en el protocolo de `normativa`) y su límite en verde (9500). `TestVerticalDeEnsayo` deja de
     exigir sitio para «otra línea de paquete de su misma longitud», y sigue exigiendo el tope de 11 000 con la vertical
     de ensayo.
  3. **`plazos-y-territorio`** (ADR 0063, punto 3). El párrafo de «Comandos» que dice cómo se pide cada operación —la
     herramienta de su fila o la orden, y los códigos 0, 2 y 3— y el apartado «Cómo se presenta el territorio» entero
     salen del cuerpo y van, en ese orden, al final de `references/municipios.md`, como dos apartados `## ` («Cómo se
     presenta el territorio» y «Herramienta u orden»); la frase que remitía del primer párrafo de la referencia al
     cuerpo se quita. En «Comandos» queda «`kitlegal` se invoca desde el `PATH`, y cada herramienta, por el nombre de su
     fila.». La orden de leer `municipios.md` dice también que trae «cómo se pide cada operación».
  4. **`jurisprudencia`** (ADR 0063, punto 3). En el paso 2, la tabla «Cómo se dio / Qué se pide» y la viñeta del
     número de resolución van al final de `references/consulta_preparada.md`, en un apartado «Lo que se pide»; la orden
     de leerla dice que trae «qué se pide según cómo se dio la referencia», y detrás queda en el cuerpo «Un número sin
     fecha no se prepara de ninguna forma, tampoco como ROJ: lleva la línea y se pide la fecha.». La viñeta de la
     referencia en la forma pedida (paso 3) y la de las sentencias que menciona el documento (paso 4) van al final de
     `references/documento_aportado.md`, en un apartado «La referencia pedida y las sentencias que menciona», y el
     inciso de «Comandos» sobre la salida 0 de un documento que no es el pedido, a su apartado «Lo que devuelve».
  5. **`normativa`**: «desde el `PATH` o, si la tienes, con su herramienta» pasa a «desde el `PATH` o con su
     herramienta». Nada más.
  6. **Lo que fija el texto de hoy**, al día con lo anterior y sin cambiar lo que comprueba: las declaraciones del
     producto de `TestSkillsDelProductoDesdeControl` (`internal/evals/laboratorio_familias_test.go`, un cambio por
     texto, con los bytes de los puntos 3 a 5); las remisiones del plan (`remisionesDelPlan`,
     `internal/skills/remisiones_test.go`), con las dos frases nuevas; y los guiones `h41-skills-instaladas` y
     `h41-1-skills-instaladas` de `internal/app/testdata/script/`, con la copia congelada del segundo
     (`specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar`). El de `jurisprudencia` busca la orden
     con su frase nueva en el cuerpo y la tabla en `consulta_preparada.md`.
  7. **La guía y el registro**: `docs/SKILLS.md` §1 (la tabla de comandos, con sus tres columnas) y §3 (el sitio de
     las materias, con su porqué y las cifras al cerrar); `CHANGELOG.md`, con la versión de cada skill que cambia.
- **Alcance**: solo lo de «Entrega». Fuera, en una lista cerrada: la forma de la línea de un paquete y de las marcas;
  una referencia índice o cualquier referencia nueva; mover una línea `⚠`, una forma fija, una regla invariante o una
  orden de leer; cambiar una descripción; el tope de 11 000, el de 900 y el de 500 líneas; `--describe`, el esquema de
  salida de las herramientas y lo que devuelve cualquier applet; las evals, sus umbrales y el job; los datos de
  `data/`; `consulta-juridica`, que no tiene tabla ni paquetes en su protocolo y está en 4481; y el laboratorio.
- **Controles**: `make skills-check` sin defectos y sin deriva, con los cuerpos bajo los 9500 sin las líneas de los
  paquetes; la regla nueva, en rojo con su mutante; `TestSkillsDelProductoDesdeControl` con cada movimiento declarado;
  `TestRemisionesDelRepositorio` con las frases nuevas; ningún `kitlegal` que devuelva otra cosa (los tests de cada
  applet sin tocar).
- **Aceptación**:
  - `make ci` en verde;
  - el guion `h45-paquete-de-mas`, con su paquete de más, en verde;
  - el cierre de evals (todas las skills, sin reutilizar nada de `main`): ninguna clase que decide por debajo de su
    umbral y ninguna serie de `claude-sonnet-5-5` por debajo de 2 de 3 en una eval que decide. Del criterio de la medida
    previa solo rigen esas dos condiciones: la muestra leída y el segundo despacho son de la medida, no del cierre;
  - la persona comprueba antes de lanzar H42 que su prototipo, rebasado sobre `main` con este hito, da `make ci` en
    verde sin ningún `.go`.
- **Antes de lanzar** (lo hace la persona, no el run):
  - **la medida previa** del ADR 0063 (punto 4), sobre el prototipo `cuerpos-prototipo`, leída y anotada en la propuesta
    de cambio de esta entrada;
  - `df -h /tmp` y sitio en la cuota de `/tmp` de la persona (el `TMPDIR` de las sesiones es `/tmp/claude-<uid>`);
  - el sondeo del alias: `claude -p --model sonnet --output-format json 'Responde solo: ok'` da en `modelUsage` el
    `MODELO_DE_EVALS` de `.github/workflows/evals.yml` (ADR 0031);
  - mover este fichero a `in-progress/`, sin su prefijo.
- **Dependencias**: H45.1. H42 depende de este hito.
