---
id: H36
fase: "1"
---

# H36 · El motor de cómputo descrito por datos (ADR 0061)

- **Objetivo**: que un plazo de otra materia entre como datos, con sus preceptos, y no como una rama nueva del código
  ni un esquema nuevo de la herramienta (principio X; ADR 0044, punto 2). Hoy hay dos cómputos fijos en Go, `lpac` y
  `contencioso`. El código decide por el código de un tipo qué cómputo aplica, que el silencio da el día siguiente, qué
  tipos fijan su duración, cuál lleva un precepto por unidad, cuántos municipios cuentan y el aviso del art. 183 LOPJ.
  `LeerTipos` no admite un tipo más que los siete de H9, y la ayuda y el esquema de `calcular` los nombran. El motor
  se valida con una segunda implementación que exige al diseño lo que la LPAC no tiene: el cómputo civil de la LEC
  (ADR 0059, punto 1).
- **Decisiones que esta entrada aplica**: el ADR 0061, que va en la misma propuesta con la nota en el `Estado` del ADR
  0042. El run no los cambia ni escribe otro. Los contratos que cambian están en su punto 7:
  - el esquema de `calcular`, que deja de depender de los datos;
  - las evals, que toman los avisos de los datos;
  - la forma de `data/plazos.yaml` y las reglas de `LeerTipos`;
  - la API de tipos y cómputos de `internal/core/plazos`;
  - del contrato de H9, FR-007, FR-008, FR-010, `contracts/applet-plazos.md` §1, §2, §5, §6 y §7,
    `contracts/datos-de-plazos.md` §2 y research D3, D6, D9 y D17;
  - el punto 3 del ADR 0042.

  No cambia lo que se calcula: con los datos de hoy, el `data` de cada plazo, su `hash`, los avisos y los errores son
  los mismos, byte a byte.
- **Lo leído antes de escribirla**, con `kitlegal boe articulo` el 2026-11-22. La ley es la Ley 1/2000, de
  Enjuiciamiento Civil, que `kitlegal boe buscar` da como `BOE-A-2000-323`.
  - **`a130`, versión 20221223**, de la LO 14/2022, `BOE-A-2022-21800`, vigente desde el 2022-12-23. El art. 130.2 hace
    inhábiles, además de los sábados, los domingos y los festivos de la comunidad o la localidad, «los días que median
    entre el 24 de diciembre y el 6 de enero del año siguiente, ambos inclusive» y «los días del mes de agosto». Con
    `--fecha 2022-12-22` sale la versión 20151006, que decía «los días 24 y 31 de diciembre».
  - **`a133`, versión 20091104**, vigente desde el 2010-05-04:
    - 133.1: el día siguiente y el vencimiento, «que expirará a las veinticuatro horas»;
    - 133.2: «en el cómputo de los plazos señalados por días se excluirán los inhábiles»;
    - 133.3: los meses de fecha a fecha, y el último del mes si no hay día equivalente;
    - 133.4: la prórroga al siguiente hábil.

    La nota del bloque dice que la Ley 13/2009 cambió los apartados 2 y 4. Con `--fecha 2010-05-03`, los apartados 1 y
    3 dicen lo mismo en la versión 20000108, vigente desde el 2001-01-08.
  - **Los arts. 131, 132, 134, 135 y 136 LEC**, para dejar fuera lo que no es cómputo (ADR 0061, punto 5).
  - **El art. 5 CC** (`BOE-A-1889-4763`, `art5`, versión 19740709): su apartado 2 dice que «en el cómputo civil de los
    plazos no se excluyen los días inhábiles». Es la opción que el ADR 0061 deja fuera.
  - **Los preceptos de los cómputos que vienen detrás**, para no reabrir el motor:
    - de lo social, el art. 43.4 LRJS (`BOE-A-2011-15936`, `a43`, versión 20221223);
    - de la revisión tributaria, los arts. 223.1 y 235.1 LGT;
    - los arts. 182.1, 183 y 185 LOPJ y 128.2 LJCA, del cómputo `contencioso`.
- **Lo comprobado en el código**, leyéndolo y con `codegraph impact`:
  - las decisiones por código están en `computar`, `municipiosDelComputo`, `fijaSuDuracion`, `TipoSilencio`, la regla
    de `unidades` de `LeerTipos` y los avisos (`internal/core/plazos`);
  - la ayuda de `argumentosDeCalcular` (`internal/app/plazos.go`) y los enumerados de `salida.go` son etiquetas de
    struct que el kernel refleja (`internal/cli/describe.go`). `--help` y `--describe` no cargan los datos: con
    `data/plazos.yaml` sin cargar responden con 0, y solo ejecutar da 1 (`internal/app/plazos_test.go`, caso
    «sin-los-tipos»);
  - `internal/evals` usa `plazos.CodigosDeAviso` y `plazos.EtiquetaDeAviso` en Go que no es de test
    (`internal/evals/plazos.go`). `schemas/eval.yaml.json` enumera a mano los cuatro códigos de aviso;
  - `TestPreceptosDeLosPlazos` solo lee un cómputo cuando lo nombra un tipo (`internal/evals/grabaciones_test.go`), y
    `UnionDeGrabaciones` junta las grabaciones de H4, H5 y H9 (`internal/evals/grabaciones.go`);
  - `TestCasosDePlazos` compara la fecha y los códigos de los avisos, no su texto (`internal/app/plazos_casos_test.go`).
- **Entrega**:
  - **`data/plazos.yaml` describe los tipos y los cómputos**, con los ejes del ADR 0061 (puntos 2 y 3) y ninguno más:
    - **en cada cómputo**: los municipios que cuentan, uno o dos, con el texto que nombra a cada uno; las unidades que
      admite; y sus periodos del año. Cada periodo lleva su precepto, que ya está entre los del cómputo o los de sus
      tipos, y un efecto `inhabil`, `no-corre` o `aviso`. El de `aviso` lleva el código, la etiqueta y el texto del
      aviso;
    - **en cada tipo**: su cómputo; su duración fija o, si no la tiene, la que recibe con una unidad de su cómputo; si
      la lleva, un precepto por unidad de su cómputo; y la fecha que da, `vencimiento` o `dia-siguiente`.

    Los nombres de las claves los fija el spec. `lpac` y `contencioso` pasan a ser datos, y con ellos los siete tipos
    de hoy, sin cambiar lo que calculan. `contencioso` admite solo `meses`. El art. 128.2 LJCA sigue entre los
    preceptos del tipo `contencioso` y en su `data.preceptos`.
  - **El motor**: `internal/core/plazos` cuenta cualquier cómputo de `data/plazos.yaml` con esos ejes y con lo que el
    ADR 0061 (punto 4) deja fijo: el día siguiente, los meses de fecha a fecha, la prórroga al primer hábil y los
    inhábiles de base.
    - Ningún Go que no sea de test nombra un tipo, un cómputo ni el aviso de un periodo, ni en un literal ni en una
      etiqueta de struct.
    - El mensaje de los municipios que no cuentan se compone, al ejecutar, con el texto de los municipios del cómputo,
      y con los datos de hoy sale igual.
    - `CodigosDeAviso` y `EtiquetaDeAviso` reciben los tipos leídos y dan los tres avisos del motor y los de los
      periodos, en el orden del ADR 0061 (punto 2).
  - **El esquema de `calcular`** (ADR 0061, punto 7):
    - la ayuda de los cinco argumentos deja de nombrar tipos y cómputos. Dice que el tipo, cuántos municipios cuentan,
      qué es la fecha de inicio y qué tipos reciben su duración están en la referencia del paquete, y sigue nombrando
      las tres unidades del motor;
    - los enumerados `tipo` y `avisos[].codigo` de la salida pasan a ser un `pattern` de la gramática del código;
    - `--help` y `--describe` siguen sin cargar los datos;
    - la descripción del verbo no cambia, porque sale en la tabla de comandos de `SKILL.md`.
  - **`LeerTipos`**:
    - deja de exigir los siete códigos;
    - admite un cómputo sin tipos;
    - exige:
      - que cada tipo nombre un cómputo que existe;
      - que la unidad de un tipo con duración sea una que admite su cómputo;
      - que las `unidades` de un tipo, si las lleva, sean las de su cómputo;
      - que el precepto de cada periodo esté entre los del cómputo o los de cada uno de sus tipos;
      - que un `no-corre` sea un mes entero de 30 o 31 días, en un cómputo que admite solo `meses`;
      - que dos periodos de un cómputo no se solapen;
      - que el código y la etiqueta de cada aviso de periodo sean únicos y distintos de los tres del motor.

    Sus defectos siguen la forma de hoy: todos a la vez, con el tipo o el cómputo, y código 1.
  - **Un error nuevo de argumentos** (código 2), detrás de los de la unidad en el orden de research D10 de H9: una
    unidad que el cómputo del tipo no admite. Con los datos de hoy no se da, porque `lpac` admite las tres unidades y
    `contencioso` no recibe duración.
  - **El cómputo `civil`** en `data/plazos.yaml`:
    - un municipio, la sede;
    - las unidades `dias-habiles` y `meses`;
    - los periodos `inhabil` de agosto y del 24 de diciembre al 6 de enero, con el 130.2;
    - los preceptos 130.2, 133.1 (`fija: expiracion`), 133.2, 133.3 y 133.4 LEC.

    Cada precepto lleva su `vigente_desde`, el que den sus redacciones grabadas (ADR 0042). Lo leído da el 2022-12-23
    para el 130.2, el 2010-05-04 para el 133.2 y el 133.4 y el 2001-01-08 para el 133.1 y el 133.3, y lo confirma
    `TestPreceptosDeLosPlazos`. Ningún tipo de `data/plazos.yaml` lo usa: los tipos procesales son de H9.1. Lo que la
    skill lleva a leer en las evals sigue saliendo de los tipos, así que el civil no entra ahí.
  - **`TestPreceptosDeLosPlazos`** lee también los cómputos sin tipos. Su tabla de palabras clave gana las de los dos
    bloques:
    - en el 130: «24 de diciembre», «6 de enero» y «agosto»;
    - en el 133: «día siguiente», «veinticuatro horas», «se excluirán los inhábiles», «de fecha a fecha» y
      «prorrogados».
  - **Una tarea `[datos]`** que graba los bloques `a130` y `a133` de `BOE-A-2000-323`, con las redacciones que
    `TestPreceptosDeLosPlazos` necesita. Lo graba en un manifiesto nuevo de H36, `evidencias/h36/`, y añade su raíz a
    `UnionDeGrabaciones`. El BOE consolidado ya tiene su fila revisada en `docs/SOURCES.md`.
  - **Los tests de H9 que cambian.** Los declara esta entrada, como H38 declaró los de H19. Cambian las constantes que
    desaparecen, los fixtures con la forma vieja de `data/plazos.yaml` (`plazosDelContrato` en `calcular_test.go`) y
    las pruebas de las reglas que se quitan. En ningún caso cambian las entradas ni lo esperado de un caso de cálculo:
    - `internal/core/plazos`: `calcular_test.go`, `propiedades_test.go`, `tipos_test.go`, `salida_test.go` y
      `avisos_test.go`;
    - `internal/app/plazos_test.go`: los enumerados en las etiquetas;
    - `internal/skills/referencias_test.go`: las constantes;
    - `internal/evals`: los tests de `CodigosDeAviso` y de `ComprobarCodigosDeAvisoDelPlazo`;
    - por el esquema: `internal/app/testdata/plazos/describe.json`, la parte de `plazos calcular` de
      `schemas/plazo.json` y `internal/app/testdata/script/h39-describe.txtar`, con su copia en `specs/`.

    Tocar `internal/app/testdata/` y las copias congeladas lo hace una tarea `[datos]` que las declara.
  - **Los contratos de H9 que cambian**, reescritos en el spec de H36 y citados desde él. Los ficheros de
    `specs/023-h9-internal-core-plazos/` no se tocan.
  - **`CHANGELOG.md`**, en «Cambiado»:
    - los tipos y los cómputos de plazos son datos;
    - el esquema de `plazos calcular` ya no enumera los tipos ni los avisos;
    - el cómputo civil de la LEC está en el motor, pero no se pide todavía desde `kitlegal plazos`.
- **Alcance**. El riesgo es parametrizar de más: solo entran los ejes que el ADR 0061 funda en un precepto leído.
  Lo de fuera es una lista cerrada:
  - **cómputos y unidades que no entran**: el cómputo civil sustantivo del art. 5.2 CC, la unidad `anios` y los
    plazos de fecha fija, como el calendario del contribuyente (ADR 0061, punto 5);
  - **lo que no es el cómputo del plazo**:
    - las actuaciones urgentes y la habilitación (arts. 131 y 133.2, párrafo segundo, LEC);
    - el plazo que corre desde el fin de otro (art. 133.1, párrafo segundo);
    - la interrupción por fuerza mayor (art. 134);
    - la presentación hasta las quince horas del día hábil siguiente (art. 135.5);
    - la hora del vencimiento como dato de la salida;
  - **lo que es de H9.1**:
    - un tipo en `data/plazos.yaml` que cuente con `civil`;
    - el cómputo social y la modalidad procesal de la LRJS como dato de entrada;
    - cualquier bandera o argumento nuevo de `kitlegal plazos calcular`;
    - la descripción del verbo `calcular`;
  - **lo que es de H42**: los tipos de la revisión tributaria;
  - **lo que no cambia**:
    - una clave nueva o un valor distinto en el `data` de la salida, en los avisos o en los mensajes, con los datos de
      hoy;
    - el cómputo `contencioso`: lo que H9 decidió de los días del art. 183 LOPJ (research S4) se mantiene, aunque el
      art. 183 grabado, en la redacción de la LO 14/2022, haga también inhábil agosto para las actuaciones judiciales;
    - `data/festivos/`, `data/cobertura.yaml` y `data/normas.yaml`;
    - cualquier skill, su protocolo, sus referencias o sus evals. Las referencias generadas desde `data/plazos.yaml`
      salen iguales;
    - la forma de `vigente_desde` y la regla del ADR 0042.
- **Controles**. Cada uno, con la comprobación que lo pone en rojo. Esa comprobación la hace el spec con un mutante o
  contra `main`, y se registra en `mutantes.md`.
  - **El `data` de hoy, byte a byte.**
    - La tarea `[aceptacion]` captura con el binario de `main` (`cd3ad148`) los goldens que faltan: el sobre entero,
      sin `fecha_consulta`, de los casos 11, 15 y 16, y el de un error de cada fila de `contracts/applet-plazos.md`
      §6, incluidos los dos de municipios que no cuentan.
    - Se comparan también los goldens `caso-a.json`, `caso-b.json` y `caso-e.json`, los guiones
      `internal/app/testdata/script/h9-plazos-*.txtar` con sus copias de `specs/023-h9-internal-core-plazos/aceptacion/`,
      `TestCasosDePlazos`, `TestCalcular` y `TestComprobarPeticion`. Ninguna entrada ni lo esperado de esos casos
      cambia.
    - En rojo: una palabra cambiada en el texto del aviso del art. 183 rompe el golden del caso 15; el silencio con
      `vencimiento` en lugar de `dia-siguiente` rompe el caso 11; el texto de los municipios de `contencioso` cambiado
      rompe el golden de su error.
  - **El esquema de `calcular`, sin datos.**
    - `describe.json` y `schemas/plazo.json` salen iguales con una copia de `data/plazos.yaml` con un tipo y un aviso de
      periodo más, y con un `data/plazos.yaml` que no carga, como el caso «sin-los-tipos».
    - En rojo: volver a las etiquetas de hoy.
  - **Ningún código de los datos en Go.**
    - Un test recorre con `go/ast` el Go que no es de test de `internal/core/plazos`, `internal/app`,
      `internal/skills` e `internal/evals`. Falla si un literal de cadena es el código de un tipo, de un cómputo o de un
      aviso de periodo, o si una etiqueta de struct lo lleva como palabra entera. Los códigos salen de
      `data/plazos.yaml` y del fichero del cómputo de ensayo. Los tres avisos de cobertura del motor quedan fuera,
      porque son del motor, y también la descripción del verbo, que es de H9.1.
    - En rojo: volver a escribir `if tipo.Codigo == "silencio"`, y volver a la ayuda de hoy.
  - **Las propiedades de H9** (`TestPropiedadesDelComputo`) siguen pasando sin cambiar su rejilla.
    - Se añaden las del cómputo civil, sobre la misma rejilla con un municipio:
      - la fecha nunca es inhábil para la sede ni cae en un periodo `inhabil`;
      - en días hábiles, entre el día siguiente al inicio y la fecha hay exactamente N días que no son inhábiles de
        base ni de un periodo;
      - en meses, la fecha es la de fecha a fecha o, si esa es inhábil, el primer hábil siguiente.
    - En rojo: un mutante que cuenta como hábiles los días de un periodo `inhabil`.
  - **Los goldens del cómputo civil.**
    - Se calculan con `plazos.Calcular`, los festivos de `data/festivos/` y el cómputo `civil` del `data/plazos.yaml` real,
      junto con unos tipos de prueba de `internal/core/plazos/testdata/`, sin duración, que cuentan con él:
      - uno con `dias-habiles`, que lleva como precepto el 133.2;
      - otro con `meses`, que lleva el 133.3.

      Esos preceptos rigen desde antes del 2022-12-22, así que el caso C6 solo nombra el 130.2. La sede es Madrid
      (`28079`). Los festivos son los nacionales de `BOE-A-2024-21316`, `BOE-A-2025-21667` y `BOE-A-2026-21254`, y los
      locales de Madrid de 2025 y 2026, sin locales de 2027.

      | # | Inicio | Duración | Fecha | Festivos saltados | Avisos | Razonamiento |
      |---|---|---|---|---|---|---|
      | C1 | 2026-07-24 | 10 días hábiles | 2026-09-07 | — | — | Desde el sábado 25: del lunes 27 al viernes 31, del 1 al 5; agosto, inhábil (130.2); del martes 1 al viernes 4 de septiembre, del 6 al 9; 5 y 6, fin de semana; lunes 7, el 10. Por la LPAC, con los mismos municipios, sería el viernes 7 de agosto |
      | C2 | 2026-12-18 | 5 días hábiles | 2027-01-08 | 2026-12-25, 2027-01-01 y 2027-01-06, nacionales | `faltan-festivos-locales` | Lunes 21, martes 22 y miércoles 23, del 1 al 3; del 24 al 6, inhábiles (130.2); jueves 7 y viernes 8, el 4 y el 5. Los festivos del periodo que no son sábado ni domingo se dan como saltados, con la regla de hoy (`contracts/applet-plazos.md` §2 de H9). No hay locales de Madrid de 2027 |
      | C3 | 2026-06-15 | 2 meses | 2026-09-01 | — | — | De fecha a fecha, el sábado 15 de agosto (133.3); agosto es inhábil y se prorroga al primer hábil (133.4), el martes 1 de septiembre. El 15 es sábado y no se da como festivo. El contencioso daría el 15 de septiembre (caso 3 de H9) |
      | C4 | 2026-10-27 | 2 meses | 2027-01-07 | 2027-01-01 y 2027-01-06, nacionales | `faltan-festivos-locales` | De fecha a fecha, el domingo 27 de diciembre; del 28 al 6, inhábiles; jueves 7. El contencioso daría el lunes 28 de diciembre con el aviso del art. 183 (caso 15 de H9) |
      | C5 | 2026-07-10 | 2 meses | 2026-09-10 | — | — | Agosto no para los meses en la LEC: de fecha a fecha, el jueves 10 de septiembre, hábil |
      | C6 | 2022-12-22 | 2 meses | — | — | — | Código 3: el art. 130.2 rige así desde el 2022-12-23 (ADR 0042). El mensaje nombra `BOE-A-2000-323`, bloque `a130`, apartado 130.2, y ningún otro precepto |
      | C7 | 2026-07-24 | 10 días naturales | — | — | — | Código 2: el cómputo civil no admite días naturales (133.2) |
      | C8 | 2026-07-24, con dos municipios | 10 días hábiles | — | — | — | Código 2: el cómputo civil cuenta con un municipio, la sede, con el texto de sus datos |
      | C9 | 2025-07-30 | 2 días hábiles | 2025-09-01 | 2025-08-15, nacional | — | Jueves 31 de julio, el 1; agosto, inhábil, y el viernes 1 de agosto es el primer día del periodo; 30 y 31 de agosto, fin de semana; lunes 1 de septiembre, el 2 |
      | C10 | 2026-11-02 | 5 días hábiles | 2026-11-10 | 2026-11-09, local de Madrid (`BOCM-20251212-34`) | — | Del martes 3 al viernes 6, del 1 al 4; 7 y 8, fin de semana; lunes 9, fiesta local de la sede; martes 10, el 5 |

      Los días de la semana son del calendario, y los festivos, de `data/festivos/` en `main` (`cd3ad148`).
    - En rojo:
      - con agosto como `aviso` en lugar de `inhabil`, C1 da el 7 de agosto;
      - con agosto como `no-corre`, C3 da el 15 de septiembre;
      - sin el periodo del 24 de diciembre, C2 da el lunes 28 de diciembre;
      - con agosto desde el día 2, C9 da el 1 de agosto;
      - sin los festivos locales de la sede, C10 da el 9 de noviembre.
  - **Un cómputo de ensayo que entra sin tocar Go.**
    - Un fichero en `internal/core/plazos/testdata/` describe un cómputo inventado, que no es de ninguna materia, con:
      - un municipio;
      - las unidades `dias-habiles` y `dias-naturales`;
      - un periodo `inhabil` que no es agosto ni la Navidad;
      - un periodo `aviso` con su propio código, su etiqueta y su texto.
    - El fichero tiene dos tipos: uno con duración fija y `dia-siguiente`, y otro sin duración.
    - Los casos, calculados a mano en el spec con la rejilla de las propiedades, prueban cada eje. Prueban también las
      fronteras de los dos periodos, su primer día y su último día laborables, y que el aviso sale con su forma
      `⚠ <ETIQUETA>:`.
    - En rojo, un mutante por eje: un motor que solo reconoce agosto como periodo; uno que ignora `dia-siguiente` fuera
      del silencio; uno que no comprueba las unidades del cómputo; y uno que da los avisos de periodo con un texto fijo.
  - **`LeerTipos`.**
    - Subtests con copias rotas de `data/plazos.yaml`. Cada uno espera el defecto que nombra el tipo o el cómputo, con
      una copia por cada regla de la entrega.
    - En rojo: quitar cada regla.
  - **Los preceptos del cómputo civil contra lo grabado** (`TestPreceptosDeLosPlazos`):
    - cada bloque está grabado;
    - lleva sus palabras clave;
    - cada `vigente_desde` es el de las redacciones grabadas.
    - En rojo: el `vigente_desde` del 133.1 con el 2010-05-04, que es el del bloque y no el del apartado; y el del
      130.2 con el 2015-10-07, un día en que el apartado decía otra cosa.
  - **Las referencias**: `make skills-sync` no deja diferencias.
  - **Las evals**: el cierre las corre como siempre. El hito toca `internal/app` e `internal/evals`, lo común de las
    evals, así que mide todas las skills sin reutilizar nada de `main`. Ningún texto de skill cambia, así que no se
    espera ningún cambio frente a `main` (ADR 0038). Si una clase cae, la reparación no toca para que pase ningún texto
    de las skills.
- **Aceptación**:
  - Con el binario de la rama y los datos de hoy, los casos de fecha de `TestCasosDePlazos` (del 1 al 18 y el 21) dan la
    misma fecha y los mismos avisos que con el de `main`. Los goldens capturados con el de `main` dan el mismo sobre,
    salvo `fecha_consulta`.
  - Los casos C1 a C10 dan lo de su tabla.
  - El cómputo de ensayo calcula sus casos, y el Go de la rama no nombra ninguno de sus códigos ni de los de
    `data/plazos.yaml`.
  - `--describe` de `calcular` sale igual con los datos de hoy, con un tipo más y sin datos.
- **Antes de lanzar** (lo hace la persona, no el run):
  - fusionar la propuesta de cambio que trae esta entrada y el ADR 0061;
  - en delorean, `df -h /tmp`, por la cuota que tumbó `make ci` local en H38;
  - el sondeo del alias: `claude -p --model sonnet --output-format json 'Responde solo: ok'` tiene que dar en
    `modelUsage` el `MODELO_DE_EVALS` de `.github/workflows/evals.yml` (ADR 0031);
  - contar con que el cierre mide todas las skills, porque el hito toca lo común de las evals;
  - mover este fichero a `in-progress/`, sin su prefijo.

  No hace falta ninguna medida previa con modelo: el hito no cambia ningún texto que lea una sesión.
- **Dependencias**: ninguna pendiente. H38 está hecho. H42 y H9.1 dependen de este hito y van detrás (ADR 0059).
