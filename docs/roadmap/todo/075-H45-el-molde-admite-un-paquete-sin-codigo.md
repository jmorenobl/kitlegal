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
      de una skill del producto esté en la declaración del laboratorio de H41.

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
  - **El cuerpo de `normativa` hace sitio** (punto 3): con una línea de paquete más y la de la vertical de ensayo, queda
    bajo las 11 000, con al menos 100 de margen. El spec elige qué texto pasa a una referencia que el cuerpo ya manda
    leer. No se mueve ninguna orden de leer, ninguna línea `⚠` ni ninguna regla, y la declaración del laboratorio de
    H41 lo dice como un movimiento.
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
  - **La descripción de `plazos-y-territorio`** deja de nombrar una materia (punto 8): cuenta los plazos que cubren sus
    paquetes y declara lo demás. Se escribe con el criterio del ADR 0055, bajo 900 caracteres, y se actualiza la
    variante de descripciones de la declaración del laboratorio de H41.
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
  - **La descripción**, con su juez de descripciones y bajo 900. En rojo: la de hoy, que nombra la LPAC.
  - **Las evals**: el cierre mide todas las skills, porque el hito toca `internal/skills`, `internal/evals` y dos
    `SKILL.md`. Ninguna eval cambia. Si una clase cae, se compara con `main` y con la serie (ADR 0038), y la reparación
    no toca otro texto de skill que el movido y la descripción.
- **Aceptación**:
  - la vertical de ensayo entra con un tipo y una fila con norma sin Go ni `SKILL.md` escrito a mano;
  - con los datos de hoy, todas las referencias salen como en `main`;
  - con una copia de `data/` y `evals/` con un paquete más, ningún test de la lista se pone en rojo;
  - `plazos-y-territorio` se sigue activando ante sus preguntas de hoy con la descripción nueva.
- **Antes de lanzar** (lo hace la persona, no el run):
  - fusionar la propuesta de cambio que trae esta entrada, la de H42 y el ADR 0062;
  - **una medida previa** de `normativa` y `plazos-y-territorio` sobre un prototipo con el movimiento y la descripción,
    unas 360 sesiones según el plan de `peorCaso` (223 y 139), sin los votos del juez. Un rojo se compara con `main`, y
    como mucho un segundo sondeo por pregunta; se lee siempre una muestra de respuestas;
  - en delorean, `df -h /tmp`;
  - el sondeo del alias: `claude -p --model sonnet --output-format json 'Responde solo: ok'` da en `modelUsage` el
    `MODELO_DE_EVALS` de `.github/workflows/evals.yml` (ADR 0031);
  - mover este fichero a `in-progress/`, sin su prefijo.
- **Dependencias**: H36, hecho. H42 depende de este hito.
