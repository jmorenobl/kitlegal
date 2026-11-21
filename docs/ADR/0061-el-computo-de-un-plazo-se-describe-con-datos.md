# 0061 · El cómputo de un plazo se describe con datos, con los ejes que piden sus preceptos y ninguno más

- **Estado**: propuesta (2026-11-22); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-22
- **Hito**: H36. Se escribe con su entrada, antes del run y fuera de él, en una sesión interactiva. Es el «ADR del
  motor» que piden la entrada de H36 y el ADR 0059 (punto 1). Cambia el contrato de H9
  (`specs/023-h9-internal-core-plazos/`), la API de tipos y cómputos de `internal/core/plazos` y el esquema de
  `kitlegal plazos calcular` (punto 7). Precisa el ADR 0042 en lo que exige `LeerTipos` (su punto 3). No cambia lo que
  `kitlegal plazos` calcula ni el `data` que devuelve.

## Contexto y problema

`kitlegal plazos calcular` cuenta hoy con dos cómputos fijos en Go, `lpac` y `contencioso`
(`internal/core/plazos/computo.go`). El código decide por el código de un tipo o de un cómputo en seis sitios:

- qué cómputo se aplica (`computar`);
- que el silencio da el día siguiente al vencimiento (`TipoSilencio`);
- qué tipos fijan su duración (`fijaSuDuracion`);
- qué tipo lleva un precepto por unidad (`unidades`, solo en `generico`);
- cuántos municipios cuentan (`municipiosDelComputo`);
- el aviso de los días del art. 183 LOPJ, con su texto.

`LeerTipos` exige los siete códigos de FR-008 de H9 y ninguno más. La ayuda de los argumentos de `calcular` y los
enumerados del esquema de salida nombran los siete tipos y los cuatro avisos en etiquetas de Go. Así, un plazo de otra
materia no puede entrar como datos: pide una rama nueva del código y un esquema nuevo de la herramienta, contra el
principio X.

El ADR 0059 fija que H36 valide el motor con el cómputo civil de la LEC. Sus preceptos, leídos con `kitlegal boe
articulo` el 2026-11-22, son estos:

- **Art. 130.2 LEC** (`BOE-A-2000-323`, bloque `a130`, versión 20221223, de la LO 14/2022, vigente desde el
  2022-12-23). Son inhábiles «los sábados y domingos, y los días que median entre el 24 de diciembre y el 6 de enero
  del año siguiente, ambos inclusive, los días de fiesta nacional y los festivos a efectos laborales en la respectiva
  Comunidad Autónoma o localidad». Sigue: «También serán inhábiles los días del mes de agosto». La redacción vigente
  el 2022-12-22 (versión 20151006) decía «los días 24 y 31 de diciembre» en lugar del periodo.
- **Art. 133 LEC** (bloque `a133`, versión 20091104, vigente desde el 2010-05-04):
  - 133.1: los plazos empiezan a correr el día siguiente, «y se contará en ellos el día del vencimiento, que expirará a
    las veinticuatro horas»;
  - 133.2: «en el cómputo de los plazos señalados por días se excluirán los inhábiles»;
  - 133.3: los plazos por meses o por años, de fecha a fecha, y el último del mes si no hay día equivalente;
  - 133.4: los que concluyen en un día inhábil, prorrogados hasta el siguiente hábil.

  La Ley 13/2009 cambió los apartados 2 y 4. Los apartados 1 y 3 dicen lo mismo en la redacción de 2001
  (versión 20000108, vigente desde el 2001-01-08).

Hay dos cómputos más que vienen detrás y que el motor no debe obligar a reabrir:

- **El social de H9.1.** El art. 43.4 LRJS (`BOE-A-2011-15936`, `a43`, versión 20221223) hace inhábiles los mismos
  dos periodos, salvo en las modalidades que enumera.
- **La revisión tributaria de H42.** Los arts. 223.1 y 235.1 LGT (`BOE-A-2003-23186`) fijan un mes desde el día
  siguiente, sin regla de cómputo propia.

El riesgo que nombra la entrada es parametrizar de más: un eje que ningún precepto pide es una opción que nadie
comprueba.

## Opciones consideradas

1. **Añadir el cómputo civil como un tercer `if` del código.** Rechazada: es lo que el principio X prohíbe. H9.1 y H42
   pedirían un cuarto y un quinto.
2. **Un lenguaje de reglas general**, con cualquier unidad, cualquier día inicial, prórroga opcional y periodos
   arbitrarios. Rechazada: son ejes que no pide ningún precepto de los cómputos que existen, de la LEC, de la LRJS ni
   de la LGT, y cada uno multiplica los casos sin comprobar.
3. **Describir con datos solo lo que distingue a los cómputos leídos**, y dejar en el motor, con su precepto, lo que
   todos dicen igual. Elegida.

Para el esquema de la herramienta había dos caminos:

- **Componer la ayuda y los enumerados desde los datos.** Rechazado. El kernel refleja etiquetas fijas
  (`internal/cli/describe.go`), y el contrato de H9 es que `--help` y `--describe` no cargan los datos, así que
  responden aunque `data/plazos.yaml` no cargue (`contracts/applet-plazos.md` §6; `internal/app/plazos.go`). Además,
  el esquema de la herramienta MCP cambiaría con cada paquete de materia, y es el protocolo con que el agente llama al
  binario.
- **Un esquema que no depende de los datos.** Elegido: lo que varía por materia vive en `data/plazos.yaml` y en la
  referencia que se genera de él, que es donde la skill toma el tipo (`plazos-y-territorio`, paso 9).

## Decisión

1. **Lo que pasa a datos.** `data/plazos.yaml` describe los tipos y los cómputos. Ningún Go que no sea de test nombra
   un tipo, un cómputo ni el aviso de un periodo, ni en un literal ni en una etiqueta de struct. Desaparecen las seis
   decisiones de «Contexto», las constantes `Tipo…`, `Computo…` y `AvisoDiasDelArt183`, `CodigosDeTipo`, `Computos` y
   la regla de los siete códigos de `LeerTipos`. `CodigosDeAviso` y `EtiquetaDeAviso` pasan a recibir los tipos
   leídos: dan los tres avisos del motor y los de los periodos de los datos. El mensaje de los municipios que no
   cuentan se compone, al ejecutar, con el texto que cada cómputo da a sus municipios.

2. **Los ejes de un cómputo** son tres. Cada uno sale de un precepto leído que lo distingue de otro cómputo:
   - **Los municipios que cuentan**: su número, uno o dos, y el texto que nombra a cada uno. El de la LPAC tiene dos,
     el del interesado y el de la sede del órgano (art. 30.6). El contencioso y el civil tienen uno, la sede
     («la respectiva Comunidad Autónoma o localidad»: art. 182.1 LOPJ y art. 130.2 LEC).
   - **Las unidades que admite**, de las tres que conoce el motor (`dias-habiles`, `dias-naturales` y `meses`):
     - la LPAC, las tres (arts. 30.2 y 30.4);
     - el contencioso, `meses`, la única que usa hoy;
     - el civil, `dias-habiles` y `meses`, porque sus días son siempre hábiles (art. 133.2 LEC).

     Una unidad que el cómputo del tipo no admite es un error de argumentos.
   - **Los periodos del año**. Cada uno va de un día y mes a otro, puede cruzar el año y lleva su precepto: un
     precepto que ya está entre los del cómputo o entre los de cada tipo que cuenta con él. Así entra en la
     comprobación del ADR 0042 donde ya está. Dos periodos de un cómputo no se solapan. Cada periodo tiene uno de tres
     efectos:
     - `inhabil`. Ningún día del periodo es hábil: no cuenta en días hábiles y lleva la prórroga. Son el agosto y el 24
       de diciembre al 6 de enero del civil (art. 130.2 LEC, con el 133.2 y el 133.4).
     - `no-corre`. El plazo por meses no corre durante el periodo, con la regla de H9 (research D9): si el inicio o el
       día siguiente cae en él, se cuenta desde su último día; si el tramo pasa por él, el vencimiento va al mismo día
       del mes siguiente. Es el agosto del contencioso (art. 128.2 LJCA). Como la regla mueve un mes, solo vale para un
       periodo que es un mes entero de 30 o 31 días, del día 1 al último, y solo en un cómputo que admite únicamente
       `meses`. Un plazo por días que no corre en agosto (el art. 128.2 lo dice de «ningún otro plazo» de la LJCA)
       pediría otra regla y su ADR.
     - `aviso`. Los días se cuentan, y si el tramo pasa por el periodo sale el aviso que el periodo lleva en los datos:
       su código, su etiqueta y su texto. Es el 24 de diciembre al 6 de enero del contencioso, con el aviso
       `dias-del-art-183-lopj` y su texto de hoy. El código y la etiqueta de un aviso de periodo son únicos en el
       fichero y distintos de los del motor. Los avisos de periodo van detrás de los dos de los festivos que faltan y
       delante del de un festivo de una parte, en el orden del fichero, que es el orden de hoy.

   El social de H9.1 cabe en estos ejes: sus dos periodos `inhabil`, y una modalidad del art. 43.4 que los exceptúa es
   otro cómputo sin ellos. Cómo elige la entrada entre los dos lo decide H9.1. Lo mismo vale para las actuaciones
   urgentes del art. 133.2, párrafo segundo, LEC, si un tipo las pide.

3. **Los ejes de un tipo** son los de hoy, ahora descritos y no decididos por su código:
   - su cómputo;
   - su duración, fija o recibida. Un tipo sin `duracion` la recibe con una unidad de su cómputo;
   - si un tipo sin duración lleva `unidades`, un precepto por cada unidad de su cómputo, ni una más ni una menos
     (hoy, `generico`);
   - **la fecha que da**: `vencimiento`, o `dia-siguiente` al vencimiento, como el silencio (art. 24 LPAC; research D8
     de H9).

4. **Lo que queda en el motor y no es un eje.** Todos los preceptos leídos de los cómputos de hoy, del civil, del
   social y de la revisión tributaria lo dicen igual, y cada cómputo lo cita entre sus preceptos:
   - **El día inicial** es el siguiente al de inicio (arts. 30.3 y 30.4 LPAC, art. 133.1 LEC y art. 5.1 CC, al que
     remite el art. 185.1 LOPJ).
   - **Los meses**, de fecha a fecha, con el último del mes si no hay día equivalente (art. 30.4 LPAC, art. 133.3 LEC y
     art. 5.1 CC).
   - **La prórroga** de un último día inhábil al primer hábil (art. 30.5 LPAC, art. 133.4 LEC y art. 185.2 LOPJ).
   - **Los inhábiles de base**: sábados, domingos y festivos nacionales, autonómicos y locales de los municipios que
     cuentan (arts. 30.2 y 30.6 LPAC, art. 130.2 LEC y art. 182.1 LOPJ). `docs/ROADMAP.md` lista «el sábado y el
     domingo siempre inhábiles» entre lo administrativo por construcción. Tras leer los preceptos, no lo es: todos los
     cómputos de la cola los hacen inhábiles. La única excepción leída es el art. 5.2 CC (punto 5).
   - **El vencimiento a las veinticuatro horas** (art. 133.1 LEC) no cambia el día que se da. Entra como precepto del
     cómputo civil (`fija: expiracion`), sin eje ni clave nueva en la salida.

5. **Lo que no entra**, cada cosa con su motivo:
   - **El cómputo civil sustantivo del art. 5.2 CC**, en el que «no se excluyen los días inhábiles». Pediría dos ejes,
     sin inhábiles y sin prórroga, que no usa ningún tipo de la cola. Si una materia lo pide, entra con un hito de la
     base, con sus dos implementaciones.
   - **Los plazos de fecha fija**, que no se cuentan desde un día inicial, como el calendario del contribuyente. No
     son un cómputo: son una fecha de un calendario publicado. Si entran, lo hacen con su ADR y su fuente.
   - **La unidad `anios`** (art. 30.4 LPAC y art. 133.3 LEC). Ningún tipo de la cola la pide.
   - **Lo que no es el cómputo del plazo**:
     - las actuaciones urgentes y la habilitación (arts. 131 y 133.2, párrafo segundo, LEC);
     - el plazo que corre desde el fin de otro (art. 133.1, párrafo segundo);
     - la interrupción por fuerza mayor (art. 134);
     - la presentación hasta las quince horas del día hábil siguiente (art. 135.5).

     Lo que de esto pidan los tipos procesales lo decide H9.1, como ya dice su entrada sobre el art. 135.
   - **Las modalidades de la LRJS** como dato de entrada, los tipos procesales, la skill y sus evals: siguen en H9.1
     (ADR 0059).

6. **La segunda implementación.** El cómputo civil (`codigo: civil`, el nombre que le da la entrada de H9.1) entra en
   `data/plazos.yaml` con los preceptos de los arts. 130.2 y 133.1 a 133.4 LEC.
   - Cada precepto lleva su `vigente_desde`, el que den sus redacciones grabadas (ADR 0042).
     `TestPreceptosDeLosPlazos` los comprueba, también los de un cómputo sin tipos.
   - Ningún tipo de `data/plazos.yaml` lo usa todavía, porque los tipos procesales son de H9.1. `LeerTipos` admite un
     cómputo sin tipos. Lo que la skill lleva a leer en las evals sigue saliendo de los tipos, así que el civil no entra
     ahí hasta H9.1.
   - Sus goldens se calculan con `plazos.Calcular` y unos tipos de prueba que cuentan con él, fuera del binario.

7. **Los contratos.**

   **Lo que no cambia, con los datos de hoy**:
   - el `data` de cada plazo que da `kitlegal plazos calcular`, y su `hash`;
   - los textos de los avisos y de los errores;
   - el orden de los avisos;
   - las referencias generadas desde `data/plazos.yaml`;
   - la descripción del verbo `calcular`, que sale en la tabla de comandos de `SKILL.md`. La pone al día H9.1, que
     añade los tipos procesales y toca la skill.

   **Lo que cambia, y se declara**:
   - **El esquema de `calcular` deja de depender de los datos.**
     - La ayuda de `tipo`, `inicio`, `municipios`, `duracion` y `unidad` deja de nombrar tipos y cómputos, y remite a
       la referencia del paquete. Sigue nombrando las tres unidades del motor.
     - Los enumerados `tipo` y `avisos[].codigo` de la salida pasan a ser un `pattern` de la gramática del código.
     - Cambian `internal/app/testdata/plazos/describe.json`, la parte de `plazos calcular` de `schemas/plazo.json` y
       los e2e que los comparan. `--help` y `--describe` siguen sin cargar los datos.
   - **Las evals.** `internal/evals` toma los códigos y las etiquetas de los avisos de los tipos leídos, no de
     constantes de Go. `ComprobarCodigosDeAvisoDelPlazo` sigue exigiendo que `schemas/eval.yaml.json` enumere
     exactamente esos códigos, ahora los de los datos. Por eso un aviso de periodo nuevo obliga a añadirlo a mano a
     `schemas/eval.yaml.json`, y lo hace el paquete que lo trae.
   - **`data/plazos.yaml`** tiene otra forma, con claves nuevas en los cómputos y en los tipos, y `LeerTipos` tiene
     otras reglas.
   - **La API de tipos y cómputos de `internal/core/plazos`**, de la que dependen `internal/app`, `internal/skills` e
     `internal/evals`. La de los festivos, que usan también `internal/source`, no cambia.
   - **Del contrato de H9**:
     - FR-007, FR-008 y FR-010: los tipos y sus municipios son los de `data/plazos.yaml`;
     - `contracts/applet-plazos.md` §1 (la ayuda), §2 (los cómputos, descritos), §5 (los avisos de periodo, de los
       datos), §6 (el error nuevo de unidad, que va tras los de la unidad en el orden de research D10) y §7 (los
       enumerados);
     - `contracts/datos-de-plazos.md` §2;
     - research D3, D6, D9 y D17, en lo que dicen de que el código decide.
   - **El punto 3 del ADR 0042**: `LeerTipos` deja de exigir «los dos cómputos». Cada precepto sigue llevando su
     `vigente_desde`.

   Ninguna skill cambia, ni su protocolo, ni sus referencias.

## Consecuencias

**A favor**

- Un cómputo y un tipo entran como datos con sus preceptos, y el esquema de la herramienta no cambia. H9.1 añade el
  social y los tipos procesales, y H42 los de la revisión tributaria, sin una rama nueva del motor.
- Cada eje tiene un precepto que lo distingue y un golden que lo prueba. Lo que todos los preceptos dicen igual sigue
  siendo una sola regla del motor.
- Que el `data` no cambie con los datos de hoy prueba que describir los cómputos no ha cambiado ninguno.

**En contra, y asumido**

- El cómputo civil llega sin tipos ni skill: hasta H9.1 nadie lo usa desde el binario. Así lo decidió el ADR 0059.
  Lo que H36 entrega es el motor y su segunda implementación probada, no un plazo procesal.
- El esquema de salida pierde sus dos enumerados. Quien quiere la lista de tipos la lee en la referencia del paquete,
  que es lo que ya hace la skill.
- `no-corre` solo admite un mes entero en un cómputo por meses. Un periodo de otra forma con ese efecto pediría otra
  regla y su ADR.
- El hito toca `internal/app` e `internal/evals`, lo común de las evals, así que el cierre mide todas las skills sin
  reutilizar nada de `main`.
