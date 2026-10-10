# Research: H25 · El juez de `jurisprudencia`

**Fecha**: 2026-10-10 · **Rama**: `020-h25-el-juez-de` · **Cabeza**: `ece5df8` · **Modo**: desatendido.

Este documento tiene cinco partes: lo verificado leyendo el repositorio (tabla V), lo medido en esta sesión (tabla M),
lo que no se ha podido verificar (supuestos S), la traza de las dos frases de `SKILL.md` a las respuestas que las
provocan (FR-080) y las decisiones (D), cada una con su alternativa rechazada.

**Cómo se midió.** Todo lo de la tabla M se ejecutó en primer plano en una copia de la cabeza hecha con
`git archive HEAD` en el directorio temporal, fuera del repositorio, y se vio terminar. En esa copia se añadieron la
carpeta del juez de `jurisprudencia`, las cuatro evals, las dos frases de `SKILL.md` y tres ficheros de test de un solo
uso. En el árbol de trabajo solo se ha escrito `specs/020-h25-el-juez-de/`. No se abrió ninguna sesión con modelo, no
se ejecutó ningún guion de Python ni se leyó ninguna credencial. Las órdenes no pidieron nada a la red.

## V. Verificado leyendo el repositorio

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | Go 1.27; el módulo fija `go 1.27.0` y `toolchain go1.27.2` | `go version` (ejecutado: `go1.27.2 darwin/arm64`); `go.mod:3`, `go.mod:5` |
| V2 | Las dependencias directas son las de hoy y el hito no añade ninguna: `kong`, `invopop/jsonschema`, `go-sdk`, `go-internal`, `santhosh-tekuri/jsonschema/v6`, `testify`, `robotstxt`, el lector de YAML (`go.yaml.in/yaml/v3`), `x/time` y `modernc.org/sqlite` | `go.mod:8-17` |
| V3 | El lector de los casos no rechaza una clave que su tipo no tiene: `ValidarDocumentoYAML` termina en `(*yaml.Node).Decode`, sin `KnownFields` | `internal/skills/esquemas.go:143-172`; ejecutado en M1 |
| V4 | El juez del job no nombra ninguna skill: `ejecutarElJob` lo lee de la carpeta `juez` del directorio de evals y comprueba su medida antes de abrir nada | `internal/evals/ejecucion.go:84-106`; `internal/evals/juez.go:119-146` |
| V5 | Los textos de una sesión son los de toda orden de Bash que nombra `kitlegal` y los de toda llamada a una herramienta del registro, en los dos modos, también cuando fallan | `internal/evals/sesion.go:162-179`, `619-699` |
| V6 | Los umbrales no nombran ninguna skill. Por modo: `sin_activar`, el de cada clase del juez y, con evals que declaran `sentencias`, `cita_sin_documento`; detrás, los dos de la medida por clase que decide; los de la duración de las sesiones solo con objetivo mayor que 0; y `duracion_del_juez` por modo. Con dos clases, sentencias, dos modos y sin objetivo son 4 + 4 + 2 + 2 = 12, y deciden todos menos el de la clase que solo se publica en cada modo: 10. **Leído, no ejecutado con esta skill**: lo fija el test de FR-111 | `internal/evals/umbrales.go:192-217`, `230-243`, `252-278` |
| V7 | El campo propio de una skill en el voto es `precepto`, en tres sitios: `VotoDeClase.Precepto`, `dichoDeClase.Precepto` y la columna «Precepto» de `informe.md` | `internal/evals/juez.go:680-683`, `956`, `982-987`; `internal/evals/informe.go:129-131`, `1926-1945` |
| V8 | El reconstructor de hoy registra `boe` y `graph`, parte la orden por blancos (`strings.Fields`), no da entrada estándar y no lee el código de la invocación | `internal/evals/medida.go:433-436`, `884-901`, `908-948` |
| V9 | `Registro.LeerDe` registra la entrada estándar que el kernel da al verbo que la lee; `AppletCita()` no tiene dependencias; `app.Main` devuelve el código sin terminar el proceso | `internal/app/registro.go:143-145`; `internal/app/cita.go:45-47`, `165-183`; `internal/app/main.go:149-163` |
| V10 | El mensaje del voto es el de la validación carácter a carácter, y la orden del voto, la de su guion | `internal/evals/juez.go:287-327` frente a `evidencias/adr-0037-jurisprudencia/guiones/juez.py:46-57`; `scripts/evals-voto.sh` (su `exec claude …`) frente a `juez.py:72-74` |
| V11 | Las reglas de la reconstrucción de la validación: `partir` (texto pegado tras la primera línea en blanco si la pregunta lleva `\n\nRoj:`), `recortes` (`documento`, `fallo`, `apartado-2`), `argumentos` (la orden se parte en cada blanco seguido de `--` y una minúscula), `repetir` (entrada estándar solo para `cotejar` sin `--documento`), `textos_del_job` (sin texto pegado no hay `cotejar`; código distinto del informe, error) y `textos_del_sondeo` (órdenes de Bash que empiezan por `kitlegal `) | `evidencias/adr-0037-jurisprudencia/guiones/casos.py:42-106`; `guiones/preguntas.py:22-54` |
| V12 | `preguntas.json` da, por id, la eval o la plantilla con `{fragmento}` y `{ficha}`, y la ruta del fragmento desde la raíz | `evidencias/adr-0037-jurisprudencia/preguntas.json` |
| V13 | La tabla de `TestCopiasDelJuez` tiene una fila; `TestMedidaVersionada` recorre la tabla en `del-repositorio`, pero sus mutaciones y las de `TestEjecucionSinMedir` copian solo la primera fila | `internal/evals/medida_test.go:259-292`, `524-541`, `610-612`; `internal/evals/ejecucion_test.go:102-123` |
| V14 | `TestDefinicionDelJob` exige a `jurisprudencia` concurrencia 1 y objetivo 0; del trabajo `medida` comprueba el modelo, la versión, la condición, que no tenga `needs` y su tope, pero no qué skills lleva su matriz | `internal/evals/definicion_test.go:143-147`, `551-585`, `615-622` |
| V15 | Las reglas del conjunto de `jurisprudencia` exigen exactamente seis evals y cuentan cinco clases por la forma de lo que espera cada eval, sin valores | `internal/evals/conjunto.go:308-350`, `744-835` |
| V16 | `TestPreguntasConElFragmento` fija que las evals 04 y 06 llevan el fragmento byte a byte, por su tamaño (2 353) y su huella | `internal/evals/conjunto_test.go:3317-3392` |
| V17 | `SKILL.md` de `jurisprudencia` tiene 195 líneas; el único pasaje que nombra el CAPTCHA está en sus líneas 18-19 y la viñeta del equivalente, en las 76-77 | `skills/jurisprudencia/SKILL.md` |
| V18 | El guardián de diff exige `[datos]` para `testdata/` y `schemas/`, rechaza todo cambio en `evidencias/` y, fuera de una tarea, todo cambio en `.github/`; `evals/` no está protegido | `scripts/workflow/guardian-diff.sh:74-102` |
| V19 | La definición del job: `jurisprudencia` con `concurrencia: 1`; topes de 352 y 269 minutos; la matriz de `medida` solo con `boe-legislacion` | `.github/workflows/evals.yml:176-184`, el `timeout-minutes` de cada trabajo, `372-381` |
| V20 | Las seis evals de H23 son hoy las del commit del informe de su cierre (`bf4befd`) | `git diff --stat bf4befd HEAD -- evals/jurisprudencia/` (ejecutado: sin salida) |
| V21 | `make ci` deja `coverage.out` y `coverage-integration.out`, que git ignora | `Makefile:86-91`, `264`; `.gitignore` (`/coverage.*`, `/bin/`) |
| V22 | La fase `plan` del precheck exige `plan.md`, `research.md`, `## Constitution Check`, «Aceptación e2e» y `## Controles de umbral` con filas que nombran requisitos del spec y un control con la forma `ci:<ruta>[:<Test>]` o `evals:<skill>:<nombre>` | `scripts/workflow/precheck.sh:47-75`; `scripts/workflow/comun.sh` (`controles_de_umbral`, `controles_de_celda`) |
| V23 | El punto de entrada de la medida y su guion reciben la skill y no nombran ninguna | `internal/evals/job_test.go` (`TestMedidaDelJuez`); `scripts/evals-medir-juez.sh`; `Makefile:147-155` |
| V24 | `cita` no figura como fuente: `cendoj.jurisprudencia` salió de `docs/SOURCES.md` el 2026-10-07 | `docs/SOURCES.md:44` |

## M. Medido en esta sesión

| # | Qué | Resultado |
|---|---|---|
| M1 | `leerCasosEtiquetados`, tal como está hoy, con los casos de `jurisprudencia` | Lee los 249, de la clase `afirma_lo_no_leido`. En los 127 derivados deja `Quitado` no nulo y sin norma ni bloque: el `texto` se pierde |
| M2 | Prototipo de la reconstrucción con las reglas de V11, en Go y en proceso (`app.Main` con `AppletCita`, la entrada estándar por `Registro.LeerDe`) | **249 de 249 resueltos**: 138 del informe de H23, 39 y 72 de los sondeos; 122 sin quitar nada, 45 sin el documento, 41 sin el fallo y 41 sin su apartado 2.º. 136 órdenes `cita` repetidas: **0 con un código distinto del del informe**, 0 sin salida estándar y 0 con salida de error. 0 órdenes `cotejar` en los 45 sin el documento. Textos por caso: 120 con ninguno, 111 con uno y 18 con dos. Mensaje del voto: 2 958 bytes de media y 7 024 el mayor. 39 ms |
| M3 | El reconstructor de hoy, sin cambiar, con los 71 casos leídos del informe de H23 | Los «resuelve» sin error y con textos que son el sobre `ok: false` de `kitlegal.cli`: el applet `cita` no está en su registro. 52 ms; nada queda en el temporal. Las evals de `jurisprudencia` no necesitan ninguna consulta preparada (0) |
| M4 | `go test ./...` con la carpeta del juez y `SKILL.md` v0.1 añadidos, y nada más | 23 paquetes en verde y `internal/evals` en rojo por cuatro tests: `TestCopiasDelJuez/del-repositorio`, `TestEvalsDelRepositorio/conjunto-jurisprudencia` (exige que la skill no tenga juez), `TestUmbralesDeJurisprudencia` (lo mismo, en sus cuatro casos) y `TestDefinicionDelJob` (`del-repositorio`, `peor-caso` y `tope-con-las-evals-del-repositorio`). Ningún test de otro paquete depende del texto de `SKILL.md` |
| M5 | Las diez evals, leídas con `LeerConjunto` | 10 evals y ningún fichero mal formado. Las cuatro preguntas nuevas son, byte a byte, las de `preguntas.json` compuestas: 39, 2 496, 423 y 123 bytes. Con las reglas de hoy dan cuatro defectos: tamaño (10 y no 6), «número y fecha» (3 y no 2), «materia» (2 y no 1) y «documento» (3 y no 1). Por clase: 3, 2, 3, 1 («no cubierta») y 1 («documento distinto») |
| M6 | Los peores casos, con `peorCaso` del repositorio, las diez evals, los 249 casos y la definición cambiada a cuatro | `evals (jurisprudencia)`: **12 577 s** = 485 + (⌈61 / 4⌉ + ⌈60 / 4⌉) × 272 + 60 + (⌈30 × 6 / 4⌉ + ⌈30 × 6 / 4⌉) × 40. `medida`: **15 505 s** = 485 + 60 + ⌈249 × 6 / 4⌉ × 40. De una en una, 47 857 s y 60 305 s. `boe-legislacion`, 21 097 s y 16 105 s, y `legal-core`, 12 181 s, no cambian. Caben en 352 minutos (21 120 s) y en 269 (16 140 s). 120 sesiones por job, y 30 respuestas juzgadas por modo |
| M7 | El binario de la cabeza con las órdenes de las evals (g) y (j) | `cita preparar --resolucion 241/2013 --fecha 2013-05-09 --json`: código 0, la dirección `https://www.poderjudicial.es/search/indexAN.jsp` y las casillas «Nº Resolución» con `241/2013` y «Fecha resolución», «Desde» y «Hasta», con `09/05/2013`. `cita preparar --texto "cláusula suelo transparencia" --json`: código 0 y la dirección de la búsqueda. Con `KITLEGAL_CACHE_DIR` en un directorio que no existía, sigue sin existir |
| M8 | Las 12 respuestas del sondeo con la skill a las cuatro preguntas que pasan a ser evals, contra lo que cada eval espera sin modelo | Las 3 de la (g) llevan la línea al principio de una línea, la dirección, las dos casillas con su valor y ninguna cita. Las 6 de la (h) y la (i) llevan `STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]`. Las 3 de la (j) llevan la dirección de búsqueda que devolvió su orden y ninguna cita ni ECLI. Solo en el modo orden |
| M9 | Las respuestas que nombran el CAPTCHA y las que dan el equivalente | 8 de las 35 del modelo que decide en el informe de H23 y 6 de las 18 del sondeo con la skill nombran el CAPTCHA; 0 de las de Haiku. 6 respuestas del informe dan el ECLI equivalente, y 1 dice que es derivado. Sus sesiones y sus frases, en «Traza de las dos frases» |
| M10 | Tamaños | Los doce umbrales de un informe, 3 772 bytes, y una entrada de `juez.respuestas` con tres votos en las dos clases, 1 510 bytes, en el prototipo de esta sesión; con el código de la cabeza, medidos en el barrido, 3 767 —escritos sin blancos y con todas sus medidas en 0— y 1 619 —la del caso `con-sentencia` de `TestInformeConElJuez`—, que son los que dan plan.md y los contratos. Las 18 respuestas de la skill en el sondeo: de 768 a 2 695 bytes, 1 604 de media. `medida.json`, 715 bytes. El fragmento, 2 353 bytes, y su ficha, 316 |
| M11 | El diff de `SKILL.md` v0.1 (contracts/skill-jurisprudencia-v0.1.diff) | `git apply --check` lo admite sobre la cabeza. Deja 197 líneas |
| M12 | Componer las evals (h) e (i) con una orden | Con `cp` de la cabecera, `sed` del fragmento y `printf` del resto, la pregunta sale igual a la de `preguntas.json` (M5). Con `cat` de la cabecera, en esta máquina salió una línea en blanco de más: un proxy de órdenes reescribe su salida |

## S. Supuestos no verificados

- **S1 · El juez vota los casos reconstruidos en Go como votó los de la validación.** No se ha abierto ningún voto
  (ADR 0032). El mensaje se compone con la misma función y los textos, con las mismas reglas; no hay ningún fichero
  versionado con los 249 mensajes de la validación con el que comparar bytes, y `guiones/casos.py` no se ha ejecutado
  (FR-040). Si no se cumple, lo dice la medida que lanza una persona antes de fusionar (SC-014).
- **S2 · Nueve sesiones a la vez no chocan con el límite de ritmo.** No medido: lo mide el cierre. Si choca, las
  sesiones cortadas quedan sin medir y el veredicto es `fallo` por la ejecución; la concurrencia no se cambia (FR-073).
- **S3 · 7,6 s por voto.** Es la cifra de la validación (`LEEME.md`), no una medida de esta sesión. Con ella, las 30
  respuestas de un modo son unos 228 s, bajo los 900 de `duracion_del_juez`.
- **S4 · Las series de las cuatro evals nuevas pasan en el modo herramienta.** El sondeo solo midió el modo orden (M8).
- **S5 · Las dos frases de `SKILL.md` v0.1 quitan el CAPTCHA de las respuestas y hacen que el equivalente se dé como
  deducido.** No medido: el run no lanza ningún sondeo. Lo lee una persona tras el cierre (SC-014).
- **S6 · Un trabajo de un runner hospedado de GitHub no pasa de 360 minutos.** Lo dice el hito; no es verificable en
  local.
- **S7 · La versión 2.1.289 de Claude Code se instala en el prefijo del juez como hasta ahora.** El paso no cambia; no
  se ha ejecutado `npm`.
- **S8 · Con `fail-fast: false`, el trabajo de la medida de una skill que falla no cancela el de la otra.** Es cómo se
  comporta la matriz de GitHub Actions, que no se puede comprobar en local. La clave ya está en el trabajo `medida` y
  en el trabajo `evals`, cuyos informes por skill llegan cada uno con su veredicto desde H6. Si no se cumpliera, la
  persona vería una medida cancelada, no una medida falsa.
- **S9 · Las dos órdenes de `sed` que componen las evals (h) e (i) dan lo mismo fuera de macOS.** Se han ejecutado con
  el `sed` de macOS (M12). Son expresiones de POSIX, y el run las ejecuta en esta misma máquina; los bytes los fija el
  test de FR-110, no la orden.

## Traza de las dos frases de `SKILL.md` (FR-080)

**El CAPTCHA.** `SKILL.md` v0 lo nombra una sola vez, en su párrafo inicial: «el buscador pide un CAPTCHA a los
programas, y no se sortea». La frase da el motivo como un hecho del buscador, sin decir que a la persona no le sale ni
qué hace la respuesta con él. Las respuestas lo repiten como motivo o, peor, se lo anuncian a la persona.

| Informe | Sesión | Frase |
|---|---|---|
| Cierre de H23 | `01-existe-con-numero-y-fecha-claude-sonnet-5-5-01` | «kitlegal no lo hace porque pide un CAPTCHA» |
| Cierre de H23 | `01-existe-con-numero-y-fecha-claude-sonnet-5-5-03` | «Pulsa buscar. Es posible que te pida un CAPTCHA.» |
| Cierre de H23 | `02-resumen-sin-documento-herramienta-claude-sonnet-5-5-01` | «El buscador te pedirá un CAPTCHA, que tienes que resolver tú.» |
| Cierre de H23 | `03-por-materia-claude-sonnet-5-5-01` | «kitlegal no consulta el buscador del CENDOJ, porque pide un CAPTCHA» |
| Cierre de H23 | `03-por-materia-claude-sonnet-5-5-02` | «Resuelve el CAPTCHA si te lo pide.» |
| Cierre de H23 | `03-por-materia-claude-sonnet-5-5-03` | «Es posible que el buscador te pida un CAPTCHA. Tendrás que resolverlo tú» |
| Cierre de H23 | `03-por-materia-herramienta-claude-sonnet-5-5-01` | «porque pide un CAPTCHA que un programa no puede resolver» |
| Cierre de H23 | `03-por-materia-herramienta-claude-sonnet-5-5-03` | «kitlegal no consulta el buscador, que pide un CAPTCHA» |
| Sondeo con la skill | `07-resumen-de-una-conocida-con-skill-01` | «que pide un CAPTCHA y por eso lo tiene que usar usted» |
| Sondeo con la skill | `07-resumen-de-una-conocida-con-skill-02` | «kitlegal no consulta el buscador del CENDOJ, que pide un CAPTCHA» |
| Sondeo con la skill | `08-doctrina-por-materia-con-skill-01` | «porque pide un CAPTCHA, así que la búsqueda tiene que hacerla usted» |
| Sondeo con la skill | `08-doctrina-por-materia-con-skill-02` | «kitlegal no consulta el buscador del CENDOJ, porque pide un CAPTCHA» |
| Sondeo con la skill | `08-doctrina-por-materia-con-skill-03` | «El buscador pide un CAPTCHA, así que esa parte tiene que hacerla usted» |
| Sondeo con la skill | `12-doctrina-dada-por-hecha-con-skill-02` | «Si el buscador le pide resolver un CAPTCHA, resuélvalo.» |

Son 8 de 35 y 6 de 18 (M9). Cinco le dicen a la persona que le saldrá o que lo resuelva.

**El equivalente.** La viñeta del paso 2 de v0 dice que, si la operación da un equivalente, «puedes nombrarlo junto a
la referencia», y no dice cómo. Las seis respuestas del cierre de H23 que lo dan son las de la eval (f), cuyo
`cita preparar --roj` deduce `ECLI:ES:TS:2023:1088`:

| Sesión | Frase | ¿Dice que es deducido? |
|---|---|---|
| `06-documento-que-no-es-el-pedido-claude-sonnet-5-5-01` | «La herramienta da como ECLI equivalente ECLI:ES:TS:2023:1088. Al ser un dato derivado, compruébalo contra el documento descargado.» | sí |
| `06-documento-que-no-es-el-pedido-claude-sonnet-5-5-02` | «El equivalente en ECLI de ese ROJ sería ECLI:ES:TS:2023:1088.» | no |
| `06-documento-que-no-es-el-pedido-claude-sonnet-5-5-03` | «(el ECLI equivalente sería ECLI:ES:TS:2023:1088)» | no |
| `06-documento-que-no-es-el-pedido-herramienta-claude-sonnet-5-5-01` | «La consulta también da el ECLI equivalente, ECLI:ES:TS:2023:1088.» | no |
| `06-documento-que-no-es-el-pedido-herramienta-claude-sonnet-5-5-02` | «El equivalente ECLI de ese ROJ sería ECLI:ES:TS:2023:1088.» | no |
| `06-documento-que-no-es-el-pedido-herramienta-claude-sonnet-5-5-03` | «Su ECLI equivalente sería ECLI:ES:TS:2023:1088.» | no |

Para el juez ninguna de las seis es un defecto (FR-003): lo que cambia es cómo lo lee la persona.

## D. Decisiones

**Skills de Go aplicadas** (`golang-how-to`, que para este hito lleva a `golang-testing`): cada test vive en el
`_test.go` del fichero que prueba y en su orden; tablas con subtests con nombre y `t.Parallel` donde no hay `t.Setenv`;
los tests afirman lo observable —el texto reconstruido, el umbral, la línea del defecto— y no cómo se calcula; ningún
test depende del orden ni de la red; y no se añade ningún paquete, dependencia ni ejecutable (D1).

**D1 · Todo en `internal/evals`, sin paquete nuevo.** El juez, la medida y la reconstrucción ya viven ahí (H24).
*Rechazada*: un paquete aparte para la reconstrucción. Usa los tipos internos del juez y de la sesión, y ningún
requisito lo pide.

**D2 · El campo `sentencia` es un segundo campo opcional del voto, junto a `precepto`.** `VotoDeClase` y lo que se lee
de cada clase llevan `Sentencia *string`, con la clave `sentencia` y sin clave cuando el esquema de la clase no la
tiene. En `informe.md`, la columna de los votos se llama «Sentencia» si los votos de la tabla la llevan y «Precepto» en
otro caso. *Rechazadas*: un campo genérico con la clave leída del esquema, que pide una codificación a medida para dos
skills; dos columnas fijas, que cambian el `informe.md` de `boe-legislacion` (FR-005); y un nombre neutro, que cambia
el esquema validado y el informe de `boe-legislacion`.

**D3 · `quitado` gana la clave `texto`, junto a `norma` y `bloque`.** El tipo pasa a llamarse `Quitado`. Con `texto`
se quita una parte del texto pegado en la pregunta; con `norma` y `bloque`, el texto de un bloque, como hasta ahora. El
lector de hoy pierde `texto` sin decir nada (M1). *Rechazada*: otra clave o un tipo por skill. Los casos llegan
validados y no cambian. La clave `regla` del fichero dice con qué cuenta se etiquetó el derivado, no se usa y no se
lee.

**D4 · Las órdenes `cita` de un caso del job se repiten en proceso, con `AppletCita` en el registro de la sesión.**
Es el camino de H24 con un applet más: la base y los directorios de la sesión se preparan igual, y tardan 52 ms en 71
sesiones (M3). *Rechazadas*: ejecutar el binario compilado, como hizo la validación, que pide construirlo en el trabajo
de la medida y en `make ci`; y un segundo camino con un registro solo de `cita`, que serían dos caminos para lo mismo.

**D5 · La orden `cita` se convierte en argumentos con la regla de la validación, no con la de H24.** Se parte en cada
blanco seguido de `--` y una minúscula: lo de antes es el applet, el verbo y la referencia; cada trozo es una bandera
con su valor, que puede llevar espacios y, en `--documento`, saltos de línea. Va siempre con `--json`, una vez. La
regla de H24, partir por blancos, rompería `--roj STS 1088/2023` y `--texto cláusula suelo`. Las órdenes de `boe` y de
`graph` siguen con la de H24 (FR-005). *Rechazada*: una sola regla nueva para todas, que cambiaría los textos con los
que se midió el juez de `boe-legislacion`.

**D6 · El texto de una orden repetida es su salida estándar.** La validación tomaba la salida de error si la estándar
estaba vacía. Con `--json` ninguna de las 136 la deja vacía (M2). *Rechazada*: copiar esa alternativa, que no cambia
ningún texto.

**D7 · El código de la orden repetida se compara con el del informe solo en las órdenes `cita`.** Lo pide FR-041 para
ese informe. Sin esa comprobación, el reconstructor de hoy da por buenos textos de error (M3). *Rechazada*: comparar
también las de `boe`. FR-005 deja su reconstrucción como está, y no se ha medido qué daría.

**D8 · Un informe es de un sondeo si lleva la clave `sondeo`, y sus preguntas están en `preguntas.json`, junto a él.**
El caso solo nombra el informe. *Rechazadas*: decidir por la ruta, que ata el código a un directorio; y una clave nueva
en el caso, que no puede cambiar.

**D9 · La pregunta de una eval que nombra `preguntas.json` se lee de las evals de hoy.** La validación la leía de git
en el commit del informe. Son las mismas (V20), no cambian (FR-065), y ni el trabajo de la medida ni `make ci` tienen
por qué tener esa historia. *Rechazada*: leerla de git, como H24 rechazó leer de git la eval retirada.

**D10 · Lo reconstruido de una sesión se recuerda por informe, sesión y texto quitado.** Con el documento recortado
cambia lo que `cotejar` recibe por la entrada estándar. Los derivados por bloque siguen quitando el texto después, como
hoy.

**D11 · Los tests no leen los votos de la evidencia.** Vota un votante que responde según la etiqueta de cada caso,
como en H24 (su D18). *Rechazada*: reproducir los 499 de `votos.jsonl`, que una persona puede sustituir al versionar
otra medida.

**D12 · La matriz de `medida` que el test espera se deriva.** Son las skills de la matriz de `evals` con carpeta de
juez, en su orden, cada una con la concurrencia de su trabajo `evals`. *Rechazada*: una lista escrita en el test, que
habría que acordarse de cambiar con el próximo juez. Derivada, una skill con juez que falte en la medida pone
`make ci` en rojo.

**D13 · Las reglas del conjunto cuentan las diez por sus clases.** Tamaño, 10; «número y fecha», 3; «materia», 2;
«documento», 3; «no cubierta», 1; «documento distinto», 1 (M5). La clase «número y fecha» pasa a exigir además una
dirección, que FR-061 nombra y que sus tres evals llevan. Los valores de cada eval no los fija ninguna regla, como en
H23: son los que devuelve el binario (M7), y un valor equivocado hace fallar su serie en el job. *Rechazada*: una regla
con los valores de las cuatro, que copiaría las evals en el código.

**D14 · Las cuatro evals se numeran seguidas, 07 a 10.** Es la convención de las tres skills y el orden de sus letras.
*Rechazada*: conservar los números del sondeo (07, 09, 11 y 12), que deja dos huecos. El test de las preguntas lleva la
tabla que une cada eval con su pregunta del sondeo.

**D15 · Las evals (h) e (i) se componen con `cp`, `sed` y `printf`, sin `cat`.** Es lo que salió bien en M12. Lo que
garantiza los bytes es el test de FR-110, no la orden.

**D16 · El texto de las dos frases de `SKILL.md` v0.1 es el del diff.** El párrafo inicial dice de quién es el
obstáculo y que la respuesta no lo anuncia, y la viñeta del equivalente, que se da como deducido y sin comprobar. Dice
la razón en lugar de prohibir: el modelo repite lo que entiende. *Rechazadas*: quitar la palabra CAPTCHA, porque FR-081
pide decir por qué kitlegal no consulta; y una regla más en «Reglas», que tocaría lo que FR-083 deja igual.

**D17 · Ninguna tarea `[datos]`.** Nada entra en `schemas/` ni en `testdata/` (V18): los tests leen la evidencia y los
informes versionados, y lo sintético lo escriben en su directorio temporal.

**D18 · «Aceptación e2e: no aplica».** El hito no cambia el binario ni el applet `cita` (FR-096); su aceptación es el
job de cierre (SC-001). Precedente: H24.

**D19 · Datos externos: ninguno.** Ninguna fuente se consulta ni se graba: el CENDOJ no se consulta (ADR 0036), y el
fragmento y los informes ya están versionados.

**D20 · La entrada del despacho cambia solo su descripción.** Deja de nombrar una skill. La condición del trabajo y la
etiqueta son las de hoy (FR-050).

## Reparaciones del cierre (FR-027)

La medición 1 del cierre (cabeza `3f99446`) salió en rojo solo por `evals (boe-legislacion)` y sin informe recogido;
la reparación no tocó nada fuera de `gates/` (`gates/supuestos.md`). La medición 2 (cabeza `e380f0a`,
`gates/cierre.json`, 2026-10-10) salió en rojo por `evals (boe-legislacion)` y por `evals (jurisprudencia)`, con los
dos informes en `gates/evals/`. Las dos reparaciones que siguen las decide `reparar_cierre` con esos informes; ninguna
toca una eval, la rúbrica, los casos, la medida ni un umbral (ADR 0029). **Las dos se apartan de lo que piden la
sección del hito y el spec, y quedan registradas como lo que son** (ADR 0030): la primera cambia `boe-legislacion`,
que la sección del hito pone en «Fuera de alcance» y cuyo `SKILL.md` FR-005 manda dejar como está —FR-027, que
encabeza este apartado, solo da permiso para el `SKILL.md` de `jurisprudencia`—; la segunda decide lo que H24 dejó
fuera de alcance y reservó a una persona (H24 research D6). Cada una dice abajo por qué se hace aun así. Su efecto lo
mide el job de evals de la medición siguiente, que es la tercera y última (FR-027); ninguna sesión del run abre
sesiones con modelo (ADR 0032).

**R1. `boe-legislacion` v0.1.8: el número y la materia de lo no leído, tampoco juntos en el aviso final.** **Evidencia.**
`gates/evals/boe-legislacion.json`: `afirma_lo_no_leido:claude-sonnet-5-5:orden` 1 de 54 y
`afirma_lo_no_leido:claude-sonnet-5-5:herramienta` 1 de 54, frente al umbral 0 (H24 FR 020), las dos en la eval 15,
`15-irpf-rendimientos-por-materia.yaml`, con tres votos unánimes y la frase en la respuesta:
`15-irpf-rendimientos-por-materia-claude-sonnet-5-5-02`, «Algunos supuestos de los apartados 1 y 2 remiten a otros
preceptos, como el art. 7 sobre rentas exentas», y `15-irpf-rendimientos-por-materia-herramienta-claude-sonnet-5-5-03`,
«el artículo remite a otros preceptos, como el art. 7 sobre rentas exentas y la disposición adicional decimoctava».
Las dos sesiones leyeron solo el bloque `a17` de `BOE-A-2006-20764` tras su índice, cuya entrada del art. 7 es
`{"id":"a7","titulo":"Artículo 7"}` sin rúbrica (`testdata/evals/boe.legislacion-consolidada/…_texto_indice.json`), y
el texto del art. 17 remite «al artículo 7 de esta Ley» sin decir de qué trata: la materia no sale de ninguna lectura,
y es la forma b de contracts/skill-boe-legislacion.md §1 de H24, la misma frase que el informe de H22 dio al caso
etiquetado como defecto de `evals/boe-legislacion/juez/casos.yaml`. Las otras cuatro respuestas de la eval 15 pasan;
en el cierre de H24 las seis pasaron, con «sin perjuicio del art. 7» y «No he leído ese artículo». **Causa.** Las dos
frases están en el párrafo final que repasa las remisiones no seguidas, y ponen la materia con «sobre», una
construcción que la viñeta «De un precepto que no has leído, nada» de v0.1.7 no nombra —nombra «que es…», «los
artículos que regulan <materia> (N y M)» y «no puedo decir qué <regla> fija» (H24 D27 y D28)—, en un sitio, el aviso
final, que la viñeta solo trata con «puedes avisar de que una materia se regula en otra parte, sin nombrar precepto ni
regla»; y la viñeta solo prohíbe, sin dar el camino para decir de qué trata lo remitido, que es leerlo. **De qué se
aparta, y por qué se hace.** `boe-legislacion` está en «Fuera de alcance» de la sección del hito (`docs/ROADMAP.md`,
H25), el spec lo repite («Todo lo de `boe-legislacion` y de `legal-core`») y FR-005 dice que su `SKILL.md` MUST seguir
como está; FR-027 no la cubre. Se cambia aun así porque la aceptación del hito pide que el veredicto de
`boe-legislacion` siga aprobado (`docs/ROADMAP.md`, H25, «Aceptación»; SC-001), y el umbral que no se cumple es de la
skill y no de la ejecución: la misma construcción en dos sesiones independientes, una en cada modo. La alternativa
descartada es no tocar la skill y volver a medir la misma cabeza: deja al azar un umbral de 0 sobre 54 respuestas por
modo con una construcción que ya ha salido dos veces en una medición, y la medición siguiente es la última. **Qué
cambia.** La viñeta gana «el art. N sobre <materia>» entre las formas, «ni al avisar al final de lo que no has leído»
entre los sitios, y «para decir de qué trata lo remitido, léelo y cítalo; si no, la remisión va con las palabras del
texto leído y nada detrás»: una línea más, y `SKILL.md` pasa de 298 a 299 líneas, su máximo (`maximoDeLineas`,
`internal/skills/sincronia.go`). El diff de `SKILL.md` con `main` es ese solo trozo, +5 −4. `CHANGELOG.md`
(*Unreleased*) lleva v0.1.8. **La línea, y el caso `dos-inicios`.** La reparación pagó esa línea reescribiendo en una
la última viñeta de «Redacción modificada», un pasaje que la medición no toca, y dejó 298; la revisión final (ronda 3)
lo retira por alcance, y esa viñeta vuelve a ser, byte a byte, la de `main`. Con 299 líneas `make ci` salía en rojo
—medido en la sesión del corrector— en `TestSkillsDelRepositorio/region/boe-legislacion/dos-inicios`
(`internal/app/skills_test.go`): el caso añadía a una copia de la skill una segunda marca de inicio, en una línea más,
y la copia, con 300, daba además el defecto de las líneas. Es el «tope efectivo de 298» que anotaron H7.4
(`specs/014-…/contracts/skill-boe-legislacion.md`) y H24 (research V13). El caso pone ahora la marca repetida en el
lugar de la línea en blanco que precede a la de fin: la copia tiene las líneas de la skill real y su único defecto es
el de la marca. Lo que comprueba no cambia —con la detección de la marca repetida quitada, los tres `dos-inicios`
salen en rojo (mutante momentáneo, que no queda en el diff)—, y el límite lo siguen fijando
`doscientas-noventa-y-nueve` y `trescientas` del mismo test. Es un cambio en un test de `internal/app`, que el hito
no tocaba; lo pide aplicar ese motivo con `make ci` en verde.
**Rechazado:** las frases marcadas como ejemplo, porque son contenido legal del art. 7 y el modelo repite el vocabulario
de la prosa (H7.3; H24 D27); leer siempre el precepto remitido, que añade una lectura a preguntas que no la necesitan
(H24 D27); prohibir nombrar por su número lo no leído, que la rúbrica admite y la regla 2 necesita (H24 D28); dar
rúbricas en `boe indice`, que la fuente no trae (H24 D28); tocar la eval, la rúbrica, los casos, la medida o el
umbral (ADR 0029); y, para la línea, pagarla con otra viñeta de `SKILL.md`, que es prosa que ninguna medición pide
cambiar, o recortar la propia viñeta, cuyo texto fijó H24 con sus mediciones (D27, D28). **Sin medir:** el efecto de
v0.1.8 lo mide el job de la medición siguiente, en los dos modos.

**R2. El voto que el tope corta se pide otra vez, una sola, dentro del presupuesto del voto nulo.** **Evidencia.**
`gates/evals/jurisprudencia.json`: veredicto `fallo` con el único motivo «de la ejecución, no de la skill: el juez dejó
1 respuestas sin juzgar: 03-por-materia-herramienta-claude-sonnet-5-5-03 (voto 1: tope de 35 s agotado)», con los doce
umbrales cumplidos (0 de 30 en `afirma_lo_no_leido` y en `cita_sin_documento` en cada modo, 40 s y 47 s de juez). La
sesión hace la misma llamada que sus dos hermanas, juzgadas dentro del tope, con una respuesta de 1 232 caracteres, de
tamaño parecido: no la explica el mensaje del voto. En la misma medición el juez tenía que juzgar 171 respuestas (54,
54, 3, 30 y 30: las de cada modo y las tres de la eval sin binario ni servidor de `boe-legislacion`, y las de cada
modo de `jurisprudencia`) y solo ese voto pasó de 35 s; en la medición 2 del cierre de H24 pasó lo mismo con una
respuesta (H24 research D28, V24). **Causa.** H24 D6 fijó el tope en 35 s «con la cola de los votos sin medir» (S6 de H24), y su contrato
(contracts/juez-y-voto.md §5 de H24) dejó el reintento fuera de alcance: un voto cortado deja la respuesta sin juzgar
y el trabajo en rojo, aunque el corte sea un hecho de la ejecución que no dice nada de la skill, y la cola de los
votos lo produce una vez cada una o dos mediciones. Subir el tope sin más lo rechaza el ADR 0029, y subirlo con su
cuenta cambia el peor caso y pide tocar `timeout-minutes` (CI), que esta reparación no toca; subir la concurrencia,
igual (H24 D6). **De qué se aparta, y por qué se hace.** H24 D6 dejó dicho que, si el cierre da respuestas sin juzgar
por el tope, «lo dice su motivo y lo decide una persona», y su contrato, que el voto que no llega no se reintenta
(fuera de alcance de su spec). Se decide aquí porque la aceptación de H25 pide que ninguna respuesta quede sin juzgar
(SC-001), el corte ha salido en la medición 2 de dos cierres seguidos con los umbrales de la skill cumplidos, y lo que
cambia cabe en las dos peticiones por voto que H24 ya contaba; lo que D6 apuntaba para la persona, subir la
concurrencia o el tope, sigue sin tocarse. Si lo quiere así lo decide ella, con el informe final, donde esta
reparación va con impacto de alcance. **Qué cambia.** `votosDelNumero` (`internal/evals/juez.go`) pide otra vez, una sola, el voto que el
tope corta, con su mismo número, como ya hacía con el nulo; dos peticiones por voto como mucho, sea cual sea la causa
de la primera, así que `votosPorRespuestaComoMucho` (6) y el peor caso de `TestDefinicionDelJob` no cambian, y los dos
topes tampoco. La respuesta queda sin juzgar solo si el tope corta también la repetición, con el motivo `voto <n>:
tope de 35 s agotado dos veces` (`causaDelTopeRepetido`); las otras tres causas siguen sin reintento. El corte se
publica: `JuicioDeRespuesta.Cortados`, `juez.votos_cortados` de informe.json —una entrada por voto cortado de una
respuesta juzgada, con `sesion` y `motivo`, en orden de sesión; `[]` sin ninguno; una respuesta sin juzgar sigue sin
llevar nada más que su motivo— y la tabla «Votos cortados por el tope y pedidos otra vez» de informe.md. Lo fijan
`TestVotoDelJuez` (el cortado que se pide otra vez, con su número; el cortado en el segundo voto; la repetición nula
que no se repite; el tope agotado dos veces; el cortado cuya repetición no es JSON), `TestOrdenDelVoto/tope` (las dos
peticiones cortadas con el sustituto de claude, los dos procesos terminados), `TestInformeConElJuez`
(`votos-cortados-y-pedidos-otra-vez`, y que una respuesta sin juzgar no va en `votos_cortados`),
`TestInformeMarkdownDeLosUmbrales`, `TestJuicioDelSondeo` y `TestEjecucionDeLaMedida`
(`un-correcto-marcado-tras-un-voto-cortado`). Documentado en `CONTRIBUTING.md`, `internal/evals/doc.go` y
`CHANGELOG.md`; contracts/juez-y-voto.md §5, §7 y §8 e informe-del-job.md §3 de H24 quedan como registro de H24 y esta
sección dice lo que cambia de ellos. **Rechazado:** subir el tope o la concurrencia (arriba); reintentar también la
sesión con error, la salida que no es JSON o la forma, porque llevan un resultado del proceso, un límite de uso entre
ellos, que repetir en el acto no cambia (H24 §5); repetir sin publicar el corte, porque lo medido se ve (ADR 0029 y
0030); y medir la cola de los votos, que no se puede medir en una sesión de un paso (ADR 0032) y que el job no
registra por voto. **Sin medir:** cuántos cortes da la medición siguiente lo dice `juez.votos_cortados` de su informe.
