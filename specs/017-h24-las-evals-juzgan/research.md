# Research: H24 · Las evals juzgan el significado con un modelo

Decisiones técnicas del plan, tomadas con el «Criterio de decisión autónoma» de la constitución. Cada una lleva su
alternativa rechazada.

**Cómo se midió.** Con tres herramientas de un solo uso escritas en el directorio temporal de la sesión, fuera del
repositorio y sin versionar: una lee los seis informes, `casos.yaml` y los votos versionados y cuenta (M1 a M3, M5,
M6); otra lista los blancos de `unicode.IsSpace` (V6); y otra aplica el prototipo de `SKILL.md` a una copia y le pasa
la regla de la prosa (V14). Ninguna usa la red ni abre una sesión con modelo. En esta sesión no se ha ejecutado
`claude`, `python3` ni `npm`: los deniega la política de las sesiones de un paso o piden una aprobación que no hay.
Lo que dependía de ellos está en «Supuestos».

## Verificado en local

| Id | Hecho | Dónde |
|---|---|---|
| V1 | `normal` quita `*`, `_` y el acento grave y colapsa `\s+`; `frase_esta` exige la frase no vacía y subcadena; `prompt_de` arma el mensaje; `voto_real` da la orden, el entorno de cuatro variables, el directorio y la lectura de la salida | `evidencias/adr-0037/guiones/juez.py` |
| V2 | Un informe versionado lleva de cada sesión `respuesta`, `eval`, `modo` e `invocaciones` (`orden`, `codigo`, `llamada`), y no la pregunta | `specs/01{1..6}-*/gates/evals/boe-legislacion.json`, con `jq` |
| V3 | `LeerSesion` solo lee las llamadas a herramientas del registro y solo guarda de su resultado si es un error; el resultado de `Bash` se descarta. `textosDelContenido` ya da los textos de un `tool_result` | `internal/evals/sesion.go:546-668` |
| V4 | La preparación en proceso existe: `Preparar`, `prepararGrafoPrevio`, `registroDeBoe`, `ejecutarConsultas`; y el applet `graph` se compone con su almacén (`AppletGrafo`, `DependenciasDeGrafo.Almacen`) | `internal/evals/preparar.go:147-445`; `internal/app/grafo.go:40-68` |
| V5 | En `regexp` de Go, `\s` es `[\t\n\f\r ]`: no incluye `\v` ni ningún blanco de Unicode | `go doc regexp/syntax` |
| V6 | `unicode.IsSpace` da 25 puntos de código: U+0009 a U+000D, U+0020, U+0085, U+00A0, U+1680, U+2000 a U+200A, U+2028, U+2029, U+202F, U+205F y U+3000 | `go doc unicode.IsSpace` y la lista de la herramienta |
| V7 | `Cmd.WaitDelay` acota la espera tras cancelarse el contexto y cierra las tuberías que el hijo deja abiertas | `go doc os/exec.Cmd.WaitDelay` |
| V8 | El paquete ejecuta procesos con el guion como único argumento variable, o con `sh -c` y una orden constante, y no lleva ningún `//nolint` por ello | `internal/evals/sesiones.go:709`, `tanda.go:303-311` |
| V9 | La política de las sesiones de un paso deniega `claude` en posición de orden y `scripts/evals*.sh`; el objetivo `make evals-medir-juez` no casa con su patrón de `make` | `scripts/workflow/politica-paso.sh:44-59`; en esta sesión denegó `claude --version` |
| V10 | El informe final lee `expresiones_prohibidas_por_modelo` con `// []`, busca un control `ci:<ruta>:<Test>` como `^func <Test>\(` en esa ruta y toma por nombre de un control `evals:` todo lo que sigue al segundo `:` | `scripts/workflow/informe.sh:207-243, 250-272` |
| V11 | El job fija `MODELO_DE_EVALS`, `VERSION_DE_CLAUDE_CODE` (2.1.284) y `timeout-minutes: 240`; `tanda` corre con cualquier despacho; la entrada del despacho se compara como `inputs.prueba_de_red == true`; ningún `checkout` de `evals.yml` pide historia | `.github/workflows/evals.yml` |
| V12 | El peor caso de hoy es 14 357 s, y `fallosDelTope` solo exige que el tope lo cubra | `internal/evals/definicion.go:29-44, 271-348`; `definicion_test.go:361-383` |
| V13 | `SKILL.md` falla con 300 líneas o más (`maximoDeLineas = 299`), y el mutante `dos-inicios` añade una línea a una copia del `SKILL.md` real: con 299 daría un defecto de más. El tope efectivo es 298 | `internal/skills/sincronia.go:25-27`; `internal/app/skills_test.go:983-992`; `specs/014-…/research.md:498-499` |
| V14 | El prototipo de v0.1.7 se aplica limpio (`git apply --check`; 19 líneas fuera, 19 dentro) y deja 298 líneas, con las mismas 14 de más de 120 caracteres que v0.1.6, todas de código o de la región generada. La regla de la prosa de D22 da cinco párrafos en v0.1.6 (líneas 111, 113, 139, 194 y 215) y ninguno en el prototipo; las 87 expresiones de la lista de hoy no dan ninguno en ninguna de las dos | `contracts/skill-boe-legislacion-v0.1.7.diff`; la herramienta replica `parrafosDeLaProsa` y `formaDeExpresion`, no las ejecuta |
| V15 | v0.1.3 ya decía «si el sobre trae avisos», `norma_modificadora` en el paso 4, «del sobre» dos veces y `fecha_vigencia` en «Redacción modificada» | `git show 196ee05:skills/boe-legislacion/SKILL.md`, líneas 106, 108, 133, 162 y 188 |
| V16 | Entre el informe de H7.1 (`a8aeb0d`) y hoy, en `evals/boe-legislacion/` solo hay ficheros añadidos y uno borrado, `19-lpac-articulo-21-redaccion-cambiada.yaml`: las preguntas de las evals 01 a 18 son las mismas. `c4819d1` borró esa eval y su grafo previo `lpac-a21-version-anterior`, una derivada de `versionDelArticulo21` con la fecha `20151002` | `git diff --name-status a8aeb0d HEAD -- evals/boe-legislacion/`; `git show c4819d1^:internal/app/grafo_test.go` |
| V17 | `LeerConjunto` lee como eval toda entrada que no sea la lista, y un directorio es un fichero mal formado | `internal/evals/conjunto.go:72-164` |
| V18 | El patrón de las expresiones de la lista no admite `_` | `schemas/expresiones-prohibidas.yaml.json` |
| V19 | `parrafosDeLaProsa` cambia cada tramo de código en línea por un espacio; `TestProsaDeLaSkill` fija esa extracción; `expresiones-calibradas`, `expresiones-en-los-bloques` y `expresiones-de-la-skill` comparan la lista con respuestas, bloques y formas | `internal/evals/conjunto_test.go:1306-1322, 1602-1839, 1956-2211` |
| V20 | La lista se aplica a respuestas en `Juzgar`, en el recuento del informe y del sondeo y en `comprobarConsultaRepetida` | `juzgar.go:908-913`, `informe.go:381-395, 1041-1083`, `sondeo.go:269`, `consulta_repetida.go:62` |
| V21 | `TestGrabacionesDerivadas` está en `internal/app`, que no puede importar `internal/evals` desde sus tests | `internal/app/grafo_test.go:2527-2533, 2724` |
| V22 | En bash 3.2, `$(…)` quita los saltos de línea finales, y con `$(…; printf x)` seguido de `${v%x}` se conservan: así lee el guion del voto la rúbrica entera | `/bin/bash` 3.2.57 del equipo, con `od -c` |

## Medido

| Id | Medida |
|---|---|
| M1 | Respuestas del modelo que decide en evals que activan la skill, sin la 21, en los seis informes: 51, 51, 51, 54, 108 y 108. Son 423, todas terminadas |
| M2 | `casos.yaml`: 259 casos; 212 `defecto` y 47 `correcto`; 5 `bitacora`, 140 `derivado` y 114 `lectura`. Los 259 tienen su sesión en su informe y ninguna respuesta está vacía. Los 140 derivados tienen una lectura con código 0 del bloque que quitan; en 2, con `boe articulos`. Dos casos, uno `bitacora` y uno `derivado`, son de la sesión `19-lpac-articulo-21-redaccion-cambiada-claude-sonnet-5-01` del informe de H7.1, cuya eval no está en `evals/boe-legislacion/`. Con los votos versionados y la regla por orden: 212 defectos marcados con tres votos, 0 correctos con un voto que marque y 683 votos pedidos |
| M3 | Votos versionados: 1 189 líneas (609 y 580), ninguno anulado ni con error; 636 síes en `afirma_lo_no_leido` y 425 en `cuenta_su_proceso`. El `valida` recalculado en Go con D8 coincide con el versionado en todos, y también con solo los blancos ASCII; ni las respuestas ni las frases llevan ningún blanco que no sea ASCII. 949 727 B en total: 799 B por voto |
| M4 | Bytes de los ejemplos de los contratos (`wc -c`): un elemento de `umbrales`, 325; una respuesta con tres votos, 1 415; la medida impresa, 656; las cinco líneas del sondeo, 432; los motivos, 394, 174 y 143; `clases.yaml` sin comentario, 132; el aviso de ejemplo, 80 |
| M5 | Respuestas que nombran «el sobre», `fecha_vigencia` o `norma_modificadora`: 0 de 51 en cada informe de H7.1, H7.2 y H7.3; 21 de 54 en el de H7.4 (9, 13 y 5 por expresión); 46 de 108 en el de H21 (31, 22 y 11); 45 de 108 en el de H22 (34, 18 y 6) |
| M6 | Lo que invocaron esas sesiones: `boe indice`, `articulo`, `articulos`, `buscar`, `metadatos` y `analisis`, `graph check` y `mcp serve`, por orden o por herramienta. Ningún otro applet |
| M7 | De las 67 respuestas con defecto de `lectura.md`, 20 son de los informes de H7.4, H21 y H22 (3, 11 y 6). Clasificadas por este plan: 7 dicen que no han leído un precepto y de qué trata; 10 nombran un precepto remitido o no cubierto con su materia; 3 exponen la regla de un precepto no leído |

## Supuestos no verificados

| Id | Supuesto | Qué pasa si no se cumple |
|---|---|---|
| S1 | `npm install --prefix <dir> <paquete>` deja el ejecutable en `<dir>/node_modules/.bin/` | El paso de instalación falla al pedir su versión, antes de la primera sesión |
| S2 | En Python, `\s` sobre texto casa los blancos de `str.isspace`: los de V6 más U+001C a U+001F | Ninguna diferencia hoy (M3). Con otro conjunto, una frase con uno de esos cuatro caracteres se comprobaría distinto que en la validación |
| S3 | Instalar el segundo Claude Code cuesta menos de 60 s, y reconstruir los 259 casos en proceso cabe en lo que el peor caso reserva fuera de los votos | El tope del trabajo no cubriría su peor caso real: se ve en la primera ejecución, y se corrige la cota |
| S4 | Un trabajo de un runner de GitHub puede durar 6 horas | Con menos, 352 minutos no serían válidos |
| S5 | Las banderas de `voto_real` y la forma de la salida de `claude -p --output-format json` (`is_error`, `structured_output`, `result`) son las de Claude Code 2.1.289. No se ha ejecutado aquí (V9); la evidencia es `juez.py` y los 4 238 votos sin error de la validación (ADR 0037) | Cada voto no llegaría a darse y el job saldría en rojo con ese motivo |
| S6 | Un voto tarda unos 8 s (ADR 0037). No hay distribución versionada: los votos de `evidencias/` no llevan su duración | Con votos de más de 35 s, respuestas sin juzgar y el job en rojo por la ejecución; el remedio es de una persona (D6) |
| S7 | En el transcript, el `tool_result` de `Bash` trae la salida de la orden en `content`, como texto o como bloques de texto, igual que el de una herramienta del servidor (H21, research S5) | Los textos del modo orden saldrían vacíos y el juez marcaría las respuestas: lo vería el cierre, con 54 marcadas en un modo |
| S8 | v0.1.7 deja `afirma_lo_no_leido` en 0 y no rompe las evals 19 y 20. No se puede medir sin sesiones con modelo | Lo mide el cierre; lo repara `reparar_cierre` sobre `SKILL.md` |

## Skills de Go aplicadas

`golang-how-to` orquesta. `golang-testing`: un fichero de test por fichero fuente (`juez_test.go`, `medida_test.go`,
`ejecucion_test.go`), tablas con subpruebas con nombre y `t.Parallel`, comportamiento observable y no detalles
internos, y la etiqueta `evals` para lo que abre sesiones. `golang-context`: el contexto de cada voto deriva del de
las señales del punto de entrada, con su plazo y su `cancel` en todos los caminos; nunca `context.Background()` a
mitad de la cadena.

## Decisiones

**D1. Todo en `internal/evals`, sin paquete nuevo.** Tres ficheros nuevos, cada uno con su test: `juez.go` (clases,
textos, mensaje, frase, voto, regla), `medida.go` (medida, copias, casos, reconstrucción, ejecución) y `ejecucion.go`
(el recorrido del job, que hoy está dentro de `TestEjecucionDelJob`). Rechazado: un paquete `internal/evals/juez`,
porque el juez necesita la sesión, el plan, el informe y la preparación, que no están exportados.

**D2. Los textos se leen del transcript con una regla por herramienta, no por modo.** Un texto es el resultado de un
`Bash` cuya orden nombra `kitlegal`, o de una herramienta del registro. Rechazado: reconstruirlos como la validación,
porque el ADR 0037 pide los del transcript; y una regla distinta por modo, porque una sesión del modo herramienta que
intenta la orden deja un error que también es un texto (FR-001).

**D3. El mensaje es el de `prompt_de`, literal.** También su última frase, «las dos preguntas»: la medida se hizo con
ella (FR-004). Una skill con otro número de clases tendrá que repetir su medida con su mensaje; no es de este hito.

**D4. El voto lo abre un guion, `scripts/evals-voto.sh`.** Es como el paquete abre una sesión (V8): el guion es el
único argumento, y los valores van en ficheros. La rúbrica se lee entera, con su salto de línea final, como en
`juez.py`. El entorno queda en las cuatro variables de la validación. Rechazado: `exec.Command` con los argumentos en
Go, que `gosec` (G204) marca y pediría un `//nolint`; y `sh -c` con los valores en variables de entorno, que cambiaría
el entorno del voto.

**D5. El votante es una función, `func(mensaje string) ([]byte, error)`.** Los tests dan una que devuelve salidas
grabadas; los puntos de entrada, una que ejecuta el guion con el contexto de sus señales. La salida se lee como
`voto_real` y, además, se valida contra `esquema.json` (FR-007). Rechazado: una interfaz con más métodos, que nadie
necesita; y pasar el contexto por `EscribirInforme`, que obligaría a cambiar su firma en todos sus tests.

**D6. Tope de un voto: 35 s, más 5 s de margen. Respuestas a la vez: la `concurrencia` de la skill (4).** El peor
caso del trabajo tiene que caber en 6 horas (S4) con el de las sesiones, que ya ocupa 14 357 s. Con seis votos por
respuesta como mucho y la cota de la instalación (S3), `(81 + 81 + 5) × (tope + margen)` tiene que ser menor que
7 183 s: 43 s como mucho por voto. 35 s son más de cuatro veces la media de la validación (S6). Rechazados: 900 s, el del guion de la validación, que da un peor
caso de días; más concurrencia, sin ninguna medida que la respalde (la validación votó con 3); y un plazo por modo que
deje sin juzgar lo que no quepa, que es una causa de «sin juzgar» que el spec no define. Si el cierre da respuestas sin
juzgar por el tope, lo dice su motivo y lo decide una persona: subir la concurrencia de la skill baja el número de
tandas y deja sitio a un tope mayor.

**D7. El juez vota dentro de `EscribirInforme`, entre el juicio sin modelo y la composición del informe.** Es donde
ya se sabe qué respuestas son las del `total` de `sin_activar`. El reloj se inyecta (`Ahora`), como en
`esperarLaDecision`, para que un test fije 900 y 901 s. Rechazado: votar antes, desde el punto de entrada, que
obligaría a leer y juzgar dos veces cada sesión; y `testing/synctest`, que no hace falta con el reloj inyectado.

**D8. Los blancos de la comprobación de la frase son los de `unicode.IsSpace` más U+001C a U+001F.** Es el conjunto
que iguala al de Python según S2. `\s` de Go no sirve (V5). Hoy no cambia ningún voto (M3). Rechazado: solo los ASCII,
que dejaría fuera el espacio de no separación que el spec nombra.

**D9. La regla es genérica sobre las clases de `clases.yaml`.** Se sigue votando mientras alguna clase que decide
tenga todos sus votos en sí; una clase que solo se publica cuenta con el primero. El voto nulo es del voto entero
(spec, «Assumptions»).

**D10. `juez` es una clave nueva de la raíz de `informe.json`, `null` si la skill no tiene juez.** Lleva las
respuestas con algún voto afirmativo y las que quedan sin juzgar. Rechazado: los votos dentro de cada elemento de
`evals`, que haría crecer las 198 entradas aunque no tengan votos.

**D11. Con el instrumento sin medir, el informe lleva el veredicto, el motivo y nada medido.** `umbrales` va vacío:
los de `sin_activar` y la duración, sin ninguna sesión, saldrían cumplidos sin haber medido nada. El recorrido está en
`ejecutarElJob`, que recibe quién abre las sesiones, para que un test de `make ci` vea que no se llama (FR-104).

**D12. La versión de Claude Code del equipo, en el sondeo, es la que declaran los transcripts de sus sesiones.** Se
abren con el mismo `claude` que los votos. Rechazado: ejecutar `claude --version`, un proceso más cuya salida no está
verificada.

**D13. Los umbrales de las respuestas existen si la skill tiene juez.** Hoy `sin_activar` existe «si la skill tiene
lista»; con la lista fuera del informe, esa condición la dejaría decidiendo qué se publica. Rechazado: siempre, que
daría dos umbrales a `legal-core` (FR-037).

**D14. La lista deja de aplicarse a toda respuesta.** Sale de `Juzgar`, del resultado de cada sesión, del recuento,
de `umbrales` y del sondeo (FR-070), y también de `comprobarConsultaRepetida`, la comprobación a mano del quickstart
de H7.2, que es otra respuesta de un modelo: FR-070 dice que la lista deja de juzgar respuestas, y el ADR 0037, que
una lista no sirve para el texto de un modelo. Se retiran `expresiones-calibradas`, `expresiones-en-los-bloques` y
`expresiones-de-la-skill` (FR-071). Se quedan el fichero, su esquema, su lectura, `ExtraerExpresionesProhibidas` y la
comprobación de la prosa. Rechazado: dejar `comprobarConsultaRepetida` como está, por no estar en la enumeración de
FR-070: la lista seguiría juzgando dos respuestas.

**D15. Los dos casos de la eval retirada se resuelven restaurando su eval y su grafo previo.** El spec supone que los
259 casos se reconstruyen con lo que hay en `main`; dos no (M2, V16). Sin la pregunta no hay mensaje, y sin el grafo
previo la salida de `graph check` no es la que el juez vio. Se restauran los dos ficheros, byte a byte con `git show
c4819d1^:…`, bajo `testdata/evals/retiradas/`, en una tarea `[datos]`, y la derivada recupera su control. No es una
grabación nueva del BOE. Rechazados: leerlos de la historia de git al ejecutar, que no existe en el `checkout` del job
ni en el de CI (V11); escribir la pregunta en el código, que es escribirla a mano; y reconstruir sin grafo previo.

**D16. El control de derivaciones de los casos es `TestGrabacionesDerivadas` de `internal/evals`.** El de
`internal/app` no puede leer casos ni reconstruir (V21). Lleva el nombre que da el hito, en el paquete que tiene los
casos; el de `internal/app` gana la entrada de la derivada restaurada.

**D17. Los textos de un caso se reconstruyen en proceso, con `app.Main`.** No hace falta el binario instalado, y
`make ci` puede ejecutarlo. Es el arnés de la validación con las mismas órdenes (M6). Rechazado: el binario del
`PATH`, que `make ci` no tiene.

**D18. Los tests de la ejecución de la medida usan un votante que responde según la etiqueta del caso.** No leen los
votos de `evidencias/`: cuando una persona versione otra medida, esos votos pueden no estar. M3 comprueba una vez, sin
dejar un test, que el código reproduce los 1 189.

**D19. Una medida por skill, de su clase que decide; sin esquema publicado.** Los dos umbrales se llaman como en el
spec, `medida_del_juez:<clase>:defectos_sin_marcar` y `…:correctos_marcados`.

**D20. `clases.yaml` con su esquema, y `LeerConjunto` aprende la carpeta `juez`** (V17). Un juez mal formado es un
fichero mal formado del conjunto, como una eval (FR-020).

**D21. Dos Claude Code en el trabajo `evals`, y un trabajo `medida` aparte.** El del juez, en un prefijo propio. La
medida no puede ser un paso de `evals`: el cierre lee `evals (<skill>)` por su nombre, y su tope es otro. Los dos
trabajos repiten el modelo y la versión del juez, y `TestDefinicionDelJob` exige que coincidan. Rechazado: `env` de
nivel de flujo, que separaría `MODELO_DEL_JUEZ` de `MODELO_DE_EVALS` (FR-090, «junto a»).

**D22. El vocabulario nuevo va a la lista, en una clave propia, y se busca en toda la prosa con su código en línea.**
Es la frontera que da el hito, «fuera de las órdenes y de la región generada». Rechazadas: buscarlo solo en los
párrafos o en las frases que nombran «la respuesta», porque no separa los usos (la frase «Sí dice lo que da la
lectura…» no la nombra, y el párrafo de la línea 194 sí) y deja sin vigilar «si el sobre trae avisos… trasládalo»; y
tres constantes en el test, porque el hito dice que entran en el vocabulario, que es el fichero.

**D23. `SKILL.md` v0.1.7 se entrega como un diff aplicable**, con cada cambio trazado a su causa
(`contracts/skill-boe-legislacion.md`). La regla nueva va en el paso 5, con su razón y no como una prohibición suelta,
como hizo H7.4 con la redacción superada. Sus cuatro líneas salen de frases que el fichero dice en otro sitio (V13,
V14). C7 quita los dos nombres de campo de la línea de la redacción modificada y dice las fechas por su orden: es el
riesgo de S8.

**D24. Aceptación e2e: no aplica.** El hito no cambia el binario; `cmd/kitlegal` no enlaza `internal/evals`.

**D25. Datos externos: ninguno.** Ninguna fuente, ninguna grabación nueva, ningún manifiesto que cambie. Lo que entra
en `testdata/` se restaura de la historia del repositorio (D15) y en `schemas/` se escribe en tareas `[datos]`.

**D26. No se añade una guarda por variable de entorno contra abrir el juez en un run.** `juez.py` la tenía. La
garantía es el sandbox de las sesiones de un paso (ADR 0032), y el guion del voto y el de la medida caen en el patrón
que la política ya deniega (V9). `scripts/workflow/` no se toca (FR-096).

## Datos externos

Ninguno (D25). `docs/SOURCES.md` no cambia y el paso `grabar_datos` no tiene nada que grabar.

## Observaciones para la persona

- El spec da por hecho que los 259 casos se reconstruyen con las grabaciones de `main`. Dos necesitan una eval y una
  derivada que H7.2 retiró (D15).
- `make evals-medir-juez` no casa con el patrón de `make` de `politica-paso.sh` (V9). No se toca: lo impide el
  sandbox.
- La versión de Claude Code de la medida, 2.1.289, se instala en el job junto a la de las sesiones, 2.1.284 (D21).
