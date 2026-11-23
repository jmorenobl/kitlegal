---
id: H45
fase: "1"
depende_de: [H36]
---

# H45 · El molde admite un paquete con sus plazos y su norma, sin código (ADR 0062)

- **Objetivo**: que una materia entre como paquete con sus normas, sus tipos de plazo, lo suyo de lo regulado aparte y
  sus evals sin tocar Go ni el protocolo de ninguna skill, como promete el principio X. Hoy no es verdad: lo enseñó un
  prototipo del paquete tributario de H42, montado el 2026-11-22 sobre `main` (`c5cf9a1b`) en una copia fuera del
  repositorio (ADR 0062, «Contexto»). Este hito arregla el molde, en Go y genérico, y lo valida con lo administrativo y
  la vertical de ensayo. H42 entra después solo con datos.
- **Decisiones que esta entrada aplica**: el ADR 0062, puntos 1 a 5, 7 y 8, que va en la misma propuesta. El run no lo
  cambia ni escribe otro. El punto 6, el cómputo de la revisión tributaria, es de H42.
- **Lo comprobado en el código y en el prototipo** (con la LGT y la Ley del IRPF con `vertical: fiscal`, dos tipos de
  plazo de la LGT, filas estrechadas y dos `fiscal.md`):
  - **La región de los plazos no sabe de paquetes.** `filasDeLosRecursosYPlazos` (`internal/skills/regiones.go`)
    escribe todos los tipos de `data/plazos.yaml` en cada referencia con la región: los dos tipos tributarios entraban
    también en `administrativo.md`, que pasaba de 100 líneas sin índice, y `make skills-sync` no terminaba.
  - **Las filas de lo regulado aparte van por su precepto.** `filasDeLaCobertura` y `filasSinRegion` asignan cada fila
    al paquete de la norma de su precepto, que en todas es la LPAC: lo tributario que sigue aparte se quedaba en
    `administrativo.md`, y una pregunta tributaria lee la referencia de lo fiscal.
  - **El cuerpo de `normativa` no tiene sitio.** Mide 10 927 caracteres, con el tope en 11 000 (ADR 0046). La línea de un
    paquete suma unos 70: con uno, 10 997; con la vertical de ensayo además, 11 066, y `TestVerticalDeEnsayo` sale en
    rojo. `jurisprudencia` mide 10 843; `plazos-y-territorio`, 10 687.
  - **Las grabaciones de un hito piden Go.** `UnionDeGrabaciones` (`internal/evals/grabaciones.go`) enumera H4, H5, H9 y
    H36 en una constante por hito. `grabar_datos` deja lo de cada hito en `evidencias/<hito>/`, con `manifiesto.json` en
    esa carpeta (`scripts/workflow/grabar-datos.sh`); `evidencias/h9/` lleva carpetas de festivos. La función la usa
    también el job de evals (`medida.go`, `sesiones.go` y `dependencias.go`).
  - **Tests que fijan con una lista literal lo que hay en `data/`, `skills/` o `evals/`**, rotos por el prototipo o por
    evals de más:
    - `internal/skills`: `TestLeerCobertura/repositorio` y `TestCoberturaDelRepositorio` (`cobertura_test.go`);
      `TestRegionDeLoCubierto/repositorio` y `/sin-data-cobertura` (`cubierto_test.go`, que espera solo
      `administrativo.md` con la región); `TestRenderizarRecursosYPlazos/desde-data` y `TestRegionDeLaCobertura`
      (`referencias_test.go`, con `filaDeLosTributos`);
    - `internal/app`: `TestCasosDePlazos/19-eval-c-tributos` (`plazos_casos_test.go`, con el texto de la fila de la DA
      1ª.2.a), y los guiones `h32-paquetes.txtar`, `h33-cobertura-instalada.txtar` y `h36-plazos-de-hoy.txtar`
      (`error-tipo.json`, el mensaje de un tipo desconocido, que nombra todos los tipos), con sus copias congeladas de
      `specs/`;
    - `internal/evals`: `TestDefinicionDelJob/peor-caso` y `/del-repositorio` (`definicion_test.go`, con el plan
      `[]int{67, 66, 6}` y «5h45m25s»), `TestPlanEnDosModos/del-repositorio`,
      `TestEvalsDelRepositorio/precepto-de-la-cobertura` (`conjunto_test.go`, un mapa de las evals de cobertura),
      `TestUmbralesDePlazosYTerritorio`, `TestUmbralesDeNormativa`, `TestUmbralesDeLasSkills`,
      `TestInformeEnDosModos/plazos-y-territorio` (`informe_test.go`) y `TestMatrizDeLasPropuestas`;
    - `internal/evals`: `TestSkillsDelProductoDesdeControl` (`laboratorio_familias_test.go`), que exige que cada fichero
      de una skill del producto esté en la declaración del laboratorio de H41: con un paquete nuevo, sus dos
      `references/<vertical>.md`.

    Un segundo prototipo, del 2026-11-22 sobre el mismo `main`, con el movimiento de `normativa` y la descripción nueva
    de «Entrega», rompe además:
    - `TestEvalsDelRepositorio/ordenes-para-powershell` (`conjunto_test.go`), que exige la forma para PowerShell de cada
      orden en su mismo fichero;
    - la declaración de `normativa` y la variante de descripciones del laboratorio de H41
      (`laboratorio_familias_test.go`, `laboratorio_descripciones_test.go` con `descripcionCortaDePlazosYTerritorio`, y
      `laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt`);
    - los subtests con sumas literales de caracteres de `TestDescripcionesDelProducto` («una de 451», «una suma de
      1701» y «una quinta skill sin texto», con 1715, 1701 y 1710), y el subtest «una que empieza por <ejemplo>» de
      `TestDescripcionesCortas`, que con la suma nueva pasa también el máximo de 1700;
    - el guion `h41-4-skills-instaladas.txtar` de `internal/app`, que lleva la descripción palabra a palabra, y su copia
      congelada `specs/038-h41-4-las-skills-del/aceptacion/skills-instaladas.txtar`.

    No es de esta lista `TestNormasDelRepositorio/paquete-de-lo-administrativo`, que declara a propósito las diez normas
    del paquete administrativo (ADR 0062, punto 5). El prototipo no corrió `make test-integration`: el plan completa la
    lista con un prototipo entero.
  - **El trabajo de evals de `plazos-y-territorio` no admite ninguna eval más.** Con la cuenta de
    `TestDefinicionDelJob`: 12 evals a concurrencia 2, 5 h 45 min, sobre un tope de 352 min; cada eval suma unos 27 min.
    Con 16, 7 h 34 min a dos y 5 h 14 min a tres. `sesionesALaVezComoMucho = 11` (`internal/evals/definicion_test.go`,
    FR-020 y SC-011 de H30) no admite la concurrencia 3. `normativa`, 335,3 min con sus 23 evals.
  - **La descripción de `plazos-y-territorio` nombra una materia**: cuenta los plazos «del procedimiento común (LPAC) o
    el del contencioso» y «no cuenta plazos de otras materias».
- **Entrega**:
  - **Los tipos de una referencia son los de su paquete** (punto 1): el `vertical` de la norma de los preceptos propios
    del tipo, los de `preceptos` y los de `unidades`, que comparten uno. Un tipo sin paquete, o con dos, es un defecto
    de la sincronía, con el tipo. Los preceptos del cómputo no cuentan.
  - **Una fila de lo regulado aparte es del paquete de la norma que la rige**, si la lleva y esa norma tiene
    `vertical`; si no, del de su precepto (punto 2). La regla es la misma en `filasDeLaCobertura` y en `filasSinRegion`:
    una fila con la norma de un paquete cuya referencia no lleva la región es un defecto. La línea no cambia.
  - **El cuerpo de `normativa` hace sitio** (punto 3): la forma para PowerShell 5.1 de la orden de lectura del paso 2
    (la frase «En PowerShell 5.1:» y su bloque `powershell`) sale del cuerpo y va, tal cual salvo la sangría, al final de
    `references/con_herramientas.md`, que el cuerpo manda leer antes de la primera orden. El cuerpo baja de 10 927 a
    unos 10 750 caracteres; con una línea de paquete más y la de la vertical de ensayo, unos 10 890, bajo las 11 000
    con más de 100 de margen. Es lo único que se mueve: «Por qué:» y «Ojo:» del paso 3 son el arreglo medido de
    `afirma_lo_no_leido` (`evidencias/sondeos-20261026/LEEME.md`) y los congela `h32-normativa.txtar`. No se mueve
    ninguna orden de Bash, ninguna línea `⚠` ni ninguna regla, y la declaración de `normativa` en el laboratorio de
    H41 lo dice como un movimiento.
  - **El control de las órdenes para PowerShell** (`TestEvalsDelRepositorio/ordenes-para-powershell`) admite que la
    forma para PowerShell de una orden del `SKILL.md` esté en una referencia a mano de la skill, con el mismo verbo, la
    misma norma y los mismos bloques. Solo para el `SKILL.md`: una orden de una referencia sigue pidiendo la suya
    detrás, en su mismo fichero. Los mutantes de `TestControlesSobreLasReferenciasAMano` crecen con el caso, y la fila
    del control en `docs/SKILLS.md` lo dice.
  - **El laboratorio de H41** (`TestSkillsDelProductoDesdeControl`) deja de exigir en la declaración las
    `references/<vertical>.md` de un paquete que su base (la variante `desde`, si la declara, o su skill de control) no tiene, con `<vertical>` cada valor de
    `vertical:` de `data/normas.yaml`. La de un paquete que la base ya tiene se sigue comparando. Así un paquete entra
    sin tocar el laboratorio, y H42 no toca Go.
  - **`UnionDeGrabaciones` suma sola las raíces** (punto 4): H4, H5 y H9, y la carpeta `boe.legislacion-consolidada/`
    de cada `evidencias/<hito>/` con `manifiesto.json`, relativa al paquete (`../../evidencias/`), en el orden del
    número del hito y no el alfabético. `evidencias/h9/` no entra. La constante de H36 desaparece.
  - **Los tests de la lista** dejan de fijar lo que hay en `data/` y en `evals/` (punto 5): los del repositorio derivan
    lo esperado de `data/` y de `evals/`, y los de una regla usan sus datos de `testdata/`. Los guiones de
    `internal/app` y sus copias congeladas los cambia una tarea `[datos]` que los declara. Ningún caso pierde lo que
    comprueba: cada uno sigue en rojo con el mutante que lo ponía en rojo. Las reglas del conjunto de `normativa`
    (como mucho 23 evals, 10 positivas de normas distintas) no cambian.
  - **La vertical de ensayo crece** (`internal/skills/testdata/vertical-de-ensayo/`):
    - un tipo de plazo cuyo precepto es de una norma de ensayo, con el cómputo `lpac`;
    - una fila de lo regulado aparte con su precepto en la LPAC y su norma de ensayo;
    - la región `recursos_y_plazos` en su `ensayo.md` de `plazos-y-territorio`.

    `TestVerticalDeEnsayo` exige que el tipo salga solo en `ensayo.md`, que la fila salga en `ensayo.md` y no en
    `administrativo.md`, que los cuerpos queden bajo el tope y que todo entre sin Go ni `SKILL.md` escrito a mano.
  - **`plazos-y-territorio` pasa a tres sesiones a la vez** en el trabajo `evals` de `.github/workflows/evals.yml`,
    con su comentario, `ajustesDelTrabajo` y `sesionesALaVezComoMucho` de 11 a 12 (punto 7, que sustituye el tope de
    H30). Con sus 12 evals de hoy, el peor caso baja; el sitio es para las de H42.
  - **La descripción de `plazos-y-territorio`** deja de nombrar una materia (punto 8), con el criterio del ADR 0055.
    Es esta, de 429 caracteres, la que mide la medida previa:

    > Calcula con kitlegal hasta qué día se puede recurrir una resolución, cuándo vence un plazo y cuándo hay
    > silencio, con los festivos que cuentan, e identifica el territorio de un municipio: INE, provincia, comunidad,
    > DIR3 y boletines. Cuenta los plazos de las materias que cubre y de las demás dice que no los cubre, sin darlos de
    > memoria. Se usa al preguntar cuándo vence un plazo o un recurso, o qué boletines tiene un municipio.

    Se actualizan con ella la variante de descripciones del laboratorio de H41, el guion `h41-4-skills-instaladas.txtar`
    (una tarea `[datos]`, con su copia congelada) y los subtests con sumas literales, que pasan a derivar de la suma de las
    descripciones de hoy. El run no cambia su texto: es el que mide la medida previa, y un rojo que la señale va al
    informe.
  - **`CHANGELOG.md`**, en «Cambiado»: la descripción de `plazos-y-territorio` y la concurrencia de su trabajo de evals.
- **Alcance**. Fuera, en una lista cerrada:
  - cualquier norma, tipo, fila, referencia o eval de una materia nueva: son de H42;
  - cambiar la forma de la región de los paquetes (una referencia con la lista, o una línea con todos): el ADR 0062 lo
    deja para cuando el siguiente paquete no quepa;
  - un campo `paquete:` en `data/plazos.yaml` o en `data/cobertura.yaml` (ADR 0062, opción 2);
  - las reglas del conjunto de `normativa`;
  - cualquier cambio en `kitlegal plazos`, en lo que calcula o en su esquema;
  - cualquier cambio de protocolo fuera del movimiento de `normativa` y de la descripción de `plazos-y-territorio`.
- **Controles**. Cada uno con la comprobación que lo pone en rojo, que el spec hace con un mutante o contra `main` y
  registra en `mutantes.md`.
  - **Lo de hoy, igual**: con los datos de `main`, `make skills-sync` no cambia ninguna referencia, y los cuerpos de
    `SKILL.md` solo cambian en el movimiento de `normativa` y en la descripción de `plazos-y-territorio`. Los casos de
    `TestCasosDePlazos` y los goldens de H9 y H36 dan lo mismo. En rojo: la sincronía con la regla de hoy y un tipo de
    ensayo, que lo escribe también en `administrativo.md`.
  - **El molde**, en `TestVerticalDeEnsayo` crecida. En rojo: un tipo asignado a todas las referencias; una fila
    asignada por su precepto; `filasSinRegion` con la regla vieja y una referencia de ensayo sin la región, que deja de
    dar su defecto; la línea de `normativa` sin el movimiento.
  - **Los defectos nuevos de la sincronía**: un tipo cuyos preceptos son de dos paquetes, y uno sin paquete, salen con
    su tipo. En rojo: quitar la regla.
  - **Los tests sin listas**: con una copia de `data/` con un tipo, una fila y una norma con `vertical` de más, y una
    copia de `evals/plazos-y-territorio/` con una eval de más, los tests de la lista siguen en verde. En rojo: devolver
    a uno de ellos su lista literal.
  - **Las grabaciones**: `TestGrabacionesSinSolape` y `TestManifiestoDeGrabaciones` con las raíces sumadas solas,
    `evidencias/h36/` incluida y `evidencias/h9/` fuera; una carpeta `evidencias/h99/` de prueba en una copia entra sin
    Go. En rojo: h36 ordenado tras una raíz de número mayor; un fichero repetido entre dos raíces.
  - **El tope del trabajo de evals** (`TestDefinicionDelJob`): las cuatro skills caben, con doce sesiones a la vez como
    mucho, y con cuatro evals de más en una copia de `plazos-y-territorio`. En rojo: `plazos-y-territorio` a dos; trece
    sesiones a la vez.
  - **La descripción**, con su juez de descripciones y bajo los topes de la variante corta (450 cada una, 1700 la
    suma). En rojo: la de hoy, que nombra la LPAC.
  - **Las órdenes para PowerShell**: una orden del `SKILL.md` con su forma en una referencia a mano no da defecto. En
    rojo: esa forma con otros bloques; esa forma en ninguna referencia; una orden de una referencia con su forma solo
    en el `SKILL.md`; una orden de una referencia con su forma en otra referencia; y la orden de articulos quitada, que
    sigue dando su defecto de verbo.
  - **El laboratorio sin las referencias de un paquete**: con una `references/<vertical>.md` de una vertical de
    `data/normas.yaml` que su base no tiene, `TestSkillsDelProductoDesdeControl` sigue en verde. En rojo:
    la misma referencia con un nombre que no es una vertical; cambiar una referencia que la base ya tiene.
  - **Las evals**: el cierre mide todas las skills, porque el hito toca `internal/skills`, `internal/evals` y dos
    `SKILL.md`. Ninguna eval cambia. Si una clase cae, se compara con `main` y con la serie (ADR 0038), y la reparación
    no toca otro texto de skill que el movido; la descripción no se cambia (va al informe).
- **Aceptación**:
  - la vertical de ensayo entra con un tipo y una fila con norma sin Go ni `SKILL.md` escrito a mano;
  - con los datos de hoy, todas las referencias generadas salen como en `main`, y de las escritas a mano solo cambia
    `references/con_herramientas.md` de `normativa`, que gana lo movido;
  - con una copia de `data/` y `evals/` con un paquete más, ningún test de la lista se pone en rojo;
  - `plazos-y-territorio` se sigue activando ante sus preguntas de hoy con la descripción nueva.
- **Antes de lanzar** (lo hace la persona, no el run):
  - fusionar la propuesta de cambio que trae esta entrada, la de H42 y el ADR 0062;
  - **una medida previa** de las cuatro skills (el despacho mide todas: también `jurisprudencia` y
    `consulta-juridica`), unas 530 sesiones según el plan de `peorCaso` (223, 121, 49 y 139) sin los votos del juez, sobre la rama `h45-prototipo`: el movimiento de
    `normativa`, la descripción nueva y los tests que tocan, sin el molde en Go. El criterio, fijado antes de medir:
    - comparadas con las del cierre de H36 (`specs/041-h36-el-motor-de/gates/informe-final.md`, sección 3), ninguna
      clase que decide cae por debajo de su umbral, y ninguna eval de `normativa` ni de `plazos-y-territorio` del
      modelo que decide baja de 3/3 a menos de 2/3;
    - la activación de `plazos-y-territorio` en sus evals no baja frente a esa medida;
    - ninguna sesión de las evals 02 y 03 de `consulta-juridica` activa `plazos-y-territorio` (`no_se_activan`).

    Un rojo se compara con `main` y con la serie (ADR 0038), y como mucho un segundo despacho por pregunta; se lee
    siempre una muestra de respuestas;
  - en delorean, `df -h /tmp`;
  - el sondeo del alias: `claude -p --model sonnet --output-format json 'Responde solo: ok'` da en `modelUsage` el
    `MODELO_DE_EVALS` de `.github/workflows/evals.yml` (ADR 0031);
  - mover este fichero a `in-progress/`, sin su prefijo.
- **Dependencias**: H36, hecho. H42 depende de este hito.
