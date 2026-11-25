---
id: H46
fase: "1"
depende_de: [H45.1]
---

# H46 · Los cuerpos de las skills de capacidad, con sitio para las materias (ADR 0063)

- **Objetivo**: que una materia nueva quepa en el cuerpo de cada skill de capacidad sin tocar su protocolo. Hoy no
  caben dos: el paquete de H42 sin ningún `.go` (`git diff 5153857c^ abfd73ca -- . ':!*.go' ':!.github'`), sobre la
  cabeza de H45.1 (`4eb02776`, fusionada como `7aaad57f`), deja `make ci` en rojo en el guion `h45-paquete-de-mas`, con
  el cuerpo de `normativa` en 11 030 caracteres de 11 000 (ADR 0046). El
  hito adelgaza los cuerpos de `normativa`, `plazos-y-territorio` y `jurisprudencia` con progressive disclosure —en el
  cuerpo, solo lo que toda respuesta necesita— y reserva en cada uno 1400 caracteres para las líneas de los paquetes,
  unas veinte materias, comprobados por `make ci`. Las respuestas no cambian: lo deciden las evals.
- **Decisiones que esta entrada aplica**: el ADR 0063 entero, que sustituye el punto 3 del ADR 0062. Siguen como están,
  y el run no los cambia ni escribe otro: el tope de 11 000 del cuerpo y los 900 de la descripción (ADR 0046, punto 2);
  la línea por paquete, con su enlace y su condición, generada como hoy; que ninguna referencia enlaza otra
  (`docs/SKILLS.md` §4); las formas de las remisiones (ADR 0051 a 0053); los controles de contenido (`docs/SKILLS.md`
  §1); y las descripciones de las cuatro skills (ADR 0057).
- **Lo medido antes de escribirla**, el 2026-11-24, en runas, como el gate (`docs/SKILLS.md` §3), sobre `main`:

  | | `normativa` | `plazos-y-territorio` | `jurisprudencia` |
  |---|---|---|---|
  | Cuerpo | 10 750 | 10 686 | 10 842 |
  | Líneas de paquete (sin las marcas) | 0 | 94 | 0 |
  | Tabla de comandos generada | 2673 | 1498 | 1302 |
  | … de ella, la columna «Qué devuelve en `data`», la línea del sobre y la de las banderas comunes | 1237 | 569 | 506 |
  | Bloques `text` (las formas fijas) | 887 | 1142 | 152 |
  | Reglas | 1298 | 2168 | 1480 |

- **La fuente de los textos es el prototipo.** Lo que entrega este hito en `skills/` es, byte a byte, lo que tiene la
  rama `cuerpos-prototipo` en `f89093dd` (`git diff main f89093dd -- skills/`). Con él, `normativa` mide 9498,
  `plazos-y-territorio` 9577 (9483 sin la línea de lo administrativo) y `jurisprudencia` 9473; `make ci` sale en verde,
  también con el paquete fiscal de H42 encima y sin ningún `.go` suyo (rama `cuerpos-con-fiscal`, sobre `26514fae`,
  que tiene los mismos `skills/` que `f89093dd`: 9568 y 9647 en las dos skills que lo reciben). Los cambios de Go y de
  tests del prototipo son una guía, no la entrega: el plan los
  rehace con sus requisitos y sus mutantes.
- **Entrega**:
  1. **La tabla de comandos, sin lo que da la llamada** (ADR 0063, punto 2). `RenderizarTabla`
     (`internal/skills/comandos.go`) escribe las columnas «Orden», «Herramienta» y «Qué hace», y ni la columna «Qué
     devuelve en `data`», ni la línea del sobre, ni la de las banderas comunes. `LeerDescripcionDeVerbo` sigue leyendo
     lo mismo, con los mismos defectos, y `RenderizarTabla` sigue exigiendo el mismo sobre y las mismas banderas en
     todas las filas de una tabla. Los tests de la tabla (`TestRenderizarTabla` en `comandos_test.go`;
     `TestRegenerarYComparar` y `TestCuerpoDeLaSkill`, con la región de `alfa` de `sincronia_test.go`) fijan los bytes
     nuevos, y el fixture del subtest «control» de `internal/app/skills_test.go`, que copia la cabecera de cuatro
     columnas, pasa a la de tres. `make skills-sync` rehace las tablas de las tres skills.
  2. **El sitio de las materias, comprobado** (ADR 0063, punto 1). En `defectosDelCuerpo`
     (`internal/skills/sincronia.go`), un defecto más para el `SKILL.md` que lleva las marcas de los paquetes: el cuerpo
     sin lo que hay entre ellas, de más de 9600 caracteres, da
     `<skill>: SKILL.md: cuerpo sin las líneas de los paquetes de <n> caracteres (máximo 9600)`. El que no lleva las
     marcas no lo tiene. `casosDePresupuestos` (`internal/app/skills_test.go`, subtest `presupuestos` de
     `TestSkillsDelRepositorio`), que hoy ensancha cada skill a 11 000 fuera de las marcas y espera cero defectos,
     cambia para las skills con marcas: ensanchado fuera de ellas, 9600 sin las líneas de los paquetes en verde, 9601
     con el defecto nuevo, y 11 001 con los dos; sin marcas, como hoy. Sin ese cambio, el subtest y el guion
     `h45-paquete-de-mas`, que lo corre en su paso de sincronía, salen en rojo. `TestVerticalDeEnsayo` deja de
     comprobar que cabe «otra línea de paquete de su misma longitud» (`comprobarElSitioDeOtroPaquete`), que añadía
     esa línea fuera de las marcas y bajaría el tope real a unos 9530; lo demás de ese test no cambia.
  3. **Los textos de `skills/`**, los del prototipo (arriba), que hacen esto (ADR 0063, punto 3):
     - `plazos-y-territorio`: el párrafo de «Comandos» sobre la herramienta o la orden y los códigos va a un apartado
       «Herramienta u orden» al final de `references/municipios.md`, y en «Comandos» queda «`kitlegal` se invoca desde
       el `PATH`, y cada herramienta, por el nombre de su fila.». De «Cómo se presenta el territorio», cómo se escribe
       cada dato y que la cobertura se explica detrás pasan a un apartado «Cómo se presenta cada dato» de la misma
       referencia, delante del anterior; en el cuerpo quedan el apartado, una frase y las tres líneas de la cobertura,
       que son una forma fija. El primer párrafo de la referencia deja de remitir al cuerpo y dice lo que trae; la orden
       de leerla, que trae «cómo se pide cada operación».
     - `jurisprudencia`: la tabla «Cómo se dio / Qué se pide» y la viñeta del número de resolución van a un apartado «Lo
       que se pide» al final de `references/consulta_preparada.md`, cuyo primer párrafo lo dice; la orden de leerla, que
       trae «qué se pide según cómo se dio la referencia», y detrás, en el cuerpo, «Un número sin fecha no se prepara de
       ninguna forma, tampoco como ROJ: lleva la línea y se pide la fecha.». La viñeta de la referencia en la forma
       pedida y la de las sentencias que menciona el documento van a un apartado «La referencia pedida y las sentencias
       que menciona» al final de `references/documento_aportado.md`, y el inciso de «Comandos» sobre la salida 0 de un
       documento que no es el pedido, a su apartado «Lo que devuelve».
     - `normativa`: «desde el `PATH` o, si la tienes, con su herramienta» pasa a «desde el `PATH` o con su
       herramienta». Nada más.
  4. **Lo que fija el texto de hoy en Go**, al día con lo anterior y sin cambiar lo que comprueba: las declaraciones del
     producto de `TestSkillsDelProductoDesdeControl` (`internal/evals/laboratorio_familias_test.go`, un cambio por
     texto, en el cuerpo o en una referencia, con los bytes del prototipo); y las remisiones del plan
     (`remisionesDelPlan`, `internal/skills/remisiones_test.go`), con las dos frases nuevas.
  5. **Los guiones que fijan el texto de hoy**: `internal/app/testdata/script/h41-skills-instaladas.txtar`, que busca
     la orden de `consulta_preparada.md` con su frase nueva en el cuerpo y la tabla en la referencia; y
     `h41-1-skills-instaladas.txtar`, con la orden nueva de `municipios.md`, igual en su copia congelada
     (`specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar`).
  6. **La guía y el registro**: `docs/SKILLS.md` §1 (qué lleva la tabla de comandos, con sus tres columnas, como texto
     nuevo; y, en «La vertical de ensayo», sin el sitio para otra línea de su misma longitud), §3 (el sitio de las
     materias, con su porqué y las cifras al cerrar) y §5 (el defecto nuevo en su tabla y
     los 9600 en «En los límites, en verde»); `CHANGELOG.md`, con la versión de cada skill que cambia.
- **Tareas y orden.** Los puntos 3, 4 y 5 van en una sola tarea `[datos]`, como T002 de H41.1
  (`specs/035-h41-1-plazos-y-territorio/tasks.md`): con los textos nuevos, cada parte que falte deja `make ci` en rojo,
  y la verificación de cada tarea corre `make ci` entero (`scripts/workflow/verificar.sh`). El orden es el punto 1, esa
  tarea y, después, el punto 2: con la regla antes que los textos, `plazos-y-territorio` y `jurisprudencia` quedan por
  encima de 9600 sin las líneas de los paquetes y `make ci` sale en rojo. El punto 6 va con lo que describe o al final.
- **La suite de aceptación** (la primera tarea) comprueba, sobre las skills instaladas con `kitlegal skills install`,
  que la tabla de cada skill de capacidad tiene tres columnas y ni la línea del sobre ni la de las banderas, que las
  tres líneas de la cobertura siguen en el cuerpo de `plazos-y-territorio`, y que lo movido está en su referencia y no
  en el cuerpo. El defecto nuevo no lo da ninguna orden de `kitlegal`: lo fijan `casosDePresupuestos` y sus mutantes,
  no la suite.
- **Alcance**: solo lo de «Entrega». Fuera, en una lista cerrada: la forma de la línea de un paquete y de las marcas;
  una referencia índice o cualquier referencia nueva; mover una línea `⚠`, una forma fija, una regla invariante o una
  orden de leer; cambiar una descripción; el tope de 11 000, el de 900 y el de 500 líneas; `--describe`, el esquema de
  salida de las herramientas y lo que devuelve cualquier applet; las evals, sus umbrales y el job; los datos de
  `data/`; `consulta-juridica`, que no tiene tabla de comandos y cuyas regiones de la base crecen con cada paquete sin
  reserva propia (ADR 0063, «En contra»); y el directorio `laboratorio/`.
- **Controles**: `make skills-check` sin defectos y sin deriva, con los cuerpos bajo los 9600 sin las líneas de los
  paquetes; el defecto nuevo, en rojo con su mutante; `TestSkillsDelProductoDesdeControl` con cada cambio declarado;
  `TestRemisionesDelRepositorio` con las frases nuevas; ningún `kitlegal` que devuelva otra cosa (los tests de cada
  applet sin tocar).
- **Aceptación**:
  - `make ci` en verde, con el guion `h45-paquete-de-mas` y su paquete de más;
  - el cierre de evals (todas las skills, sin reutilizar nada de `main`): ninguna clase que decide por debajo de su
    umbral y ninguna serie de `claude-sonnet-5-5` por debajo de 2 de 3 en una eval que decide. Del criterio de la medida
    previa solo rigen esas dos condiciones: la muestra leída y el segundo despacho son de la medida, no del cierre;
  - la persona comprueba antes de lanzar H42 que su prototipo, rebasado sobre `main` con este hito, da `make ci` en
    verde sin ningún `.go`.
- **Antes de fusionar esta entrada**: la medida previa del ADR 0063 (punto 4), sobre `cuerpos-prototipo`, leída con su
  criterio y anotada en la propuesta de cambio.
- **Antes de lanzar** (lo hace la persona, no el run):
  - `df -h /tmp` y sitio en la cuota de `/tmp` de la persona (el `TMPDIR` de las sesiones es `/tmp/claude-<uid>`);
  - el sondeo del alias: `claude -p --model sonnet --output-format json 'Responde solo: ok'` da en `modelUsage` el
    `MODELO_DE_EVALS` de `.github/workflows/evals.yml` (ADR 0031);
  - mover este fichero a `in-progress/`, sin su prefijo.
- **Dependencias**: H45.1. H42 depende de este hito.
