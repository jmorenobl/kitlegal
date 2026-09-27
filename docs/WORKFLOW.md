# Workflow por hito con spec-kit

Cada hito de `ROADMAP.md` se implementa con una pasada del workflow `hito` de spec-kit (`.specify/workflows/hito/workflow.yml`). La persona está en los dos extremos y en ningún punto intermedio (ADR 0018): **escribe la entrada** —la sección del hito en `docs/ROADMAP.md`— y **lee el informe final**, que llega como cuerpo de la propuesta de cambio. Entre medias, los comandos `speckit.*` corren en modo headless (`claude -p`), los gates los deciden scripts y jueces con rúbrica, y el run solo se detiene por una causa mayor que detecta el código.

## Piezas

| Pieza | Dónde | Para qué |
|---|---|---|
| Constitución | `.specify/memory/constitution.md` | Principios, restricciones, DoD, las tres capas de gate y el criterio con el que se decide sin persona |
| Workflow | `.specify/workflows/hito/workflow.yml` | Secuencia, rondas de juez, bucle de tareas, revisión, cierre e informe |
| Pasos shell | `scripts/workflow/*.sh` | La lógica determinista de cada paso: prechecks, rondas, guardián, verificación, cuarentena, aceptación congelada, grabación, cierre e informe |
| Lanzador y supervisor | `scripts/hito.sh` | Comprueba `main` limpio y las credenciales, toma el candado de sesión única, impide el reposo del Mac y supervisa el run: espera ante límites de uso, reanuda fallos transitorios y solo se detiene ante una causa mayor o el rechazo de entrada |
| Sesión única | `scripts/workflow/sesion-unica.sh` + gancho `PreToolUse` en `.claude/settings.json` | Mientras vive un run, ninguna otra sesión de Claude Code puede editar ni cambiar el historial de ese árbol |
| Guardia de push | `scripts/lefthook/pre-push/guardia-push.sh` | Gancho `pre-push` de lefthook: rechaza `main`, push forzados, borrados y etiquetas, sea cual sea la orden |
| Wrapper de modelo | `scripts/claude-modelo.sh` | Ejecutable de Claude para spec-kit: traduce `--model <modelo>@<esfuerzo>` a `--model` + `--effort` |
| Coste por paso | `scripts/coste-run.sh` | Consumo y coste estimado de un run por paso y por rol, desde los transcripts |
| Extensión git | `.specify/extensions/git/` | Rama `NNN-hN-slug` por hito |
| Skills | `.claude/skills/speckit-*` | Los comandos `/speckit-*` (generados por `specify init`, no editar a mano) |
| Artefactos | `specs/NNN-hN-slug/` | `spec.md`, `plan.md`, `research.md`, `tasks.md`, `checklists/`, `aceptacion/`, `gates/` |
| Evidencia de datos | `evidencias/<hito>/` | Material de origen grabado por `grabar_datos`, con `manifiesto.json` de huellas |

## Secuencia

```
extraer_hito (la sección del hito: la entrada de la persona)
→ specify → clarify_preguntas → ronda_clarify ×3 [resolver_clarify (proceso nuevo) → check_clarify] → completar_clarify → clarify_integrar
→ ronda_spec ×4 [precheck_spec → juez_spec (juez de entrada) → leer_gate_spec → corrector_spec] → check_gate_spec   ← único rechazo posible antes del final
→ plan → ronda_plan ×4 [precheck → juez → leer → corrector] → check_gate_plan → commit
→ tasks → analyze → ronda_tasks ×4 [precheck → juez → leer → corrector] → check_gate_tasks → commit
→ bucle_tareas:
     siguiente_tarea (primera "- [ ] Tnnn" fuera de cuarentena; si no queda ninguna, converge una vez)
     → implementar_tarea (intento 1: modelo_implementacion; 2 y 3: modelo_escalada con el diagnóstico del anterior)
     → verificar (guardián de diff + «rojo primero» si [aceptacion] + manifiesto grabable si [datos] + make ci)
     → si rojo: redelimitada → reintento | reparar → verificar_reparacion → si sigue rojo, el diagnóstico va a la nota de la tarea
     → estado_cierre → commit_tarea → congelar_aceptacion ([aceptacion]) → grabar_datos ([datos] con manifiesto)
     (tras tres intentos sin verde: cuarentena y sigue)
→ activar_aceptacion (la suite congelada entra en make ci)
→ ci (make ci del hito) → si rojo: bucle_reparacion ×3 [reparar_hito → cerrar_reparacion]
→ barrido (artefactos y documentación contra el producto) → cerrar_barrido
→ ronda_revision ×4 [juez A ∥ juez B → leer_revision → corrector_revision → cerrar_correccion] → cerrar_revision
→ publicar (push + propuesta de cambio) → bucle_cierre ×3 [medir_cierre (CI + evals sobre la cabeza) → reparar_cierre]
→ informe_final (cuerpo de la propuesta de cambio)
```

La fusión a `main` (squash-merge) y el release los decide siempre una persona; en una sesión interactiva, a petición suya, los ejecuta el agente con confirmación de cada orden (ADR 0020). El workflow deja la rama del hito commiteada, la empuja a `origin`, abre la propuesta de cambio y pone el informe como cuerpo. Nunca hace `merge`, ni push a `main`, ni push forzado: además de los permisos de `.claude/settings.json`, `scripts/lefthook/pre-push/guardia-push.sh` (instalado con `make hooks` o `lefthook install`) rechaza esas referencias aunque el YAML cambiara. Una persona lo salta con `KITLEGAL_PUSH_HUMANO=1` (etiquetas de release); esa orden es `ask` en `.claude/settings.json`, así que en headless no se ejecuta. Motivación en `docs/ADR/0007-workflow-desatendido.md` y `docs/ADR/0018-workflow-autonomo-persona-en-los-extremos.md`.

## Qué detiene el run

Nada que decida un modelo. Solo esto:

| Parada | Quién la detecta | Qué deja |
|---|---|---|
| Rechazo de entrada | `check_gate_spec`: el juez de entrada lista motivos en `entrada` con tipo `contradiccion`, `sin_criterio_comprobable` u `objetivo_vacio` | `gates/informe-rechazo.md`: fragmento, por qué impide derivar un test y una pregunta cerrada por motivo. Se corrige la sección del hito y se relanza desde `main` |
| DAG bloqueado | `siguiente_tarea`: tres tareas seguidas en cuarentena sin ninguna en verde entre medias | `gates/dossier.md` |
| Dependencia externa inaccesible | `grabar_datos` (la fuente no responde tras un reintento), `publicar` o `medir_cierre` (push, `gh` o comprobaciones que no terminan en 3 h) | `gates/dossier.md` |
| Credencial ausente | `hito.sh` al arrancar (`claude`, `gh auth status`, `origin`) y el cierre | mensaje del supervisor o `gates/dossier.md` |
| Presupuesto de tiempo | `hito.sh`: el run supera `KITLEGAL_TIEMPO_MAXIMO` segundos (48 h) desde su primer paso | mensaje del supervisor |

Todo lo demás sigue: un gate que agota sus rondas pasa con sus motivos a `gates/<fase>-pendiente.md`; una tarea que no sale entra en cuarentena; un corrector que toca lo que no debe se aparta; un cierre que sigue en rojo tras tres mediciones queda así en el informe.

## Tres capas de gate

Sigue la sección «Gates» de la constitución: cada comprobación vive en la capa más baja que pueda verificarla.

1. **Mecánica (shell, sin LLM).** `scripts/workflow/precheck.sh` corre antes de cada juez de artefactos y escribe sus defectos en `gates/<fase>-precheck.txt`: marcadores pendientes, checklists, «Fuera de alcance», ids de requisito únicos y citados, escenarios Dado/Cuando/Entonces, «Aceptación e2e» en el plan, formato e ids de tareas, rutas declaradas, `[datos]`, una sola `[aceptacion]` y la primera, y ninguna tarea de plataforma, de persona ni con `KITLEGAL_RECORD`. **No paran el run**: `gate.sh leer` fuerza el rechazo mientras quede alguno y el corrector los arregla. `verificar.sh` ejecuta el guardián de diff y `make ci`, que encadena lint, `-race`, schema-check, drift de `references/`, test de arquitectura y golden files de citas.
2. **Juez LLM con rúbrica.** `juez_spec`, `juez_plan`, `juez_tasks` y los dos jueces finales reciben criterios fijos y escriben `{"veredicto","criterios":[{id,criterio,cumple,evidencia}],"motivos"}` (el de entrada, además, `"entrada":[…]`). **No corrigen nada.** Revisan de forma exhaustiva y, en rondas posteriores, comprueban primero los motivos anteriores. Si rechazan, un corrector (proceso distinto) aplica los motivos generalizando cada uno a su clase de defecto, y se vuelve a juzgar: hasta cuatro veredictos y tres correcciones. Un veredicto ausente o mal formado repite la ronda con un juez nuevo. Un motivo de alcance, frontera humana, privacidad, TOS, anomalías o decisión cerrada lo aplica el corrector con la lectura conservadora y deja una línea en `gates/supuestos.md`. Los autores (`specify`, `plan`, `tasks`) comprueban su artefacto contra la rúbrica y el precheck antes de entregar.
3. **Humano, en los extremos.** Antes del run, la sección del hito y la fila revisada de `docs/SOURCES.md` en `main` para toda fuente que se vaya a grabar. Después, el informe final: fuentes, anomalías, adaptadores nuevos, fixtures y esquemas modificados, grabaciones, evidencia y supuestos, antes de fusionar.

## Clarificación sin sesgo y sin persona

`clarify` se divide para que quien formula las preguntas no sea quien las responde:

1. **`clarify_preguntas`** ejecuta `/speckit-clarify` con la orden de solo escribir las preguntas y sus opciones en `gates/clarify-preguntas.json`, sin recomendación ni opción preferida, y sin tocar el spec.
2. **`resolver_clarify`** es un `prompt` en un proceso `claude -p` nuevo que decide solo con documentos (hito, `CLAUDE.md`, `refs/`, constitución): (a) lo determinan las fuentes → esa es la respuesta; (b) no está especificado → «fuera de alcance: no se implementa»; (c) varias opciones válidas → la de mayor calidad, con la alternativa rechazada; (d) alcance, frontera humana, privacidad, TOS, anomalías o decisión cerrada → **lectura conservadora** con `conservadora: true`. Nunca escala.
3. **`check_clarify`** comprueba que cada pregunta tiene respuesta; si falta alguna, `resolver_clarify` lo intenta de nuevo (tres veces). **`completar_clarify`** cierra lo que quede sin respuesta como «fuera de alcance». **`clarify_integrar`** lleva los pares Q/A a `spec.md` con marca `(auto: criterio …)` o `(auto: conservadora …)`.

## Implementación tarea a tarea

El bucle toma la primera línea `- [ ] Tnnn` que no está en cuarentena, anota el commit base, las rutas que declara y la huella de su nota, y lanza `/speckit-implement` restringido a esa tarea.

- **Guardián de diff** (`scripts/workflow/guardian-diff.sh tarea`): el diff desde la base se limita a las rutas declaradas, más `go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y `x_test.go` de un `x.go` declarado; declarar un directorio permite todo lo que cuelga de él. `testdata/` y `schemas/` solo con `[datos]`; `evidencias/` nunca (la escribe `grabar_datos`); la suite de aceptación congelada nunca; ningún `t.Skip` nuevo. `.golangci.yml` se admite sin declararlo solo cuando el diff se limita a añadir entradas a `misspell.ignore-rules` (palabras españolas que el diccionario inglés toma por erratas). Una violación es un rojo como otro: el reparador la ve en `gates/ci.log`.
- **Verificación** (`scripts/workflow/verificar.sh tarea`): guardián; «rojo primero» si la tarea es `[aceptacion]`; manifiesto grabable si es `[datos]`; y `make ci` (sin `Makefile`, `go build ./... && go vet ./... && go test -race ./...`). La última línea de `gates/ci.log` es `kitlegal-verificacion exit=N`, que `estado_cierre` lee en lugar de fiarse de salidas de pasos que quizá no se ejecutaron.
- **Rojo**: si la tarea quedó **redelimitada** (sigue `[ ]` y su nota cambió en este intento, normalmente con la línea de `tasks.md` ampliada), no se repara ni se commitea: el intento siguiente relee la línea con sus rutas nuevas y hereda el árbol. En cualquier otro rojo, `reparar` recibe las últimas 80 líneas del log; si sigue en rojo, `registrar_fallo` añade el diagnóstico a `gates/tarea-Tnnn.md` y el intento siguiente (con `modelo_escalada`) parte de ahí.
- **Cuarentena**: una tarea que no queda `[X]` en verde tras tres intentos —o marcada `[X]` sin que dos iteraciones de cierre la dejen commiteada— aparta su trabajo como `gates/cuarentena/Tnnn.patch`, el árbol vuelve al último commit en verde (el directorio del feature conserva su contenido, salvo la suite congelada) y el bucle sigue. Tres cuarentenas seguidas sin ninguna tarea en verde entre medias son una causa mayor: DAG bloqueado.
- **Aceptación congelada**: la primera tarea, `[aceptacion]`, escribe desde el spec los guiones testscript de la entrega en `<feature>/aceptacion/*.txtar` (y las evals de la skill si el hito la toca). `scripts/workflow/aceptacion.sh rojo-primero` copia cada guion un momento a `internal/app/testdata/script/` y exige que falle por una aserción: uno que pasa sin implementación, o que falla por no poder leerse, no prueba nada. Tras el commit, `congelar` guarda las huellas en `gates/aceptacion-congelada.json`; al acabar el bucle, `activar` los copia a `internal/app/testdata/script/<hito>-*.txtar` y desde ahí `make ci` los ejecuta. Si el plan dice «Aceptación e2e: no aplica», no hay tarea `[aceptacion]`.
- **Datos externos**: una tarea `[datos]` deja el manifiesto `<paquete>/testdata/grabaciones.json` (con `fuente`) y su test `//go:build grabacion` `TestGrabar*`. Tras su commit, `grabar_datos` (`scripts/workflow/grabar-datos.sh grabar`) comprueba que la fuente tiene fila en `docs/SOURCES.md` de `main` con «Revisado» fechado, ejecuta el test con `KITLEGAL_RECORD=1` y `KITLEGAL_EVIDENCIAS=evidencias/<hito>`, admite solo lo escrito bajo `testdata/` del paquete y `evidencias/<hito>/`, registra huellas en `gates/grabaciones.md` y `evidencias/<hito>/manifiesto.json` (los ficheros de más de 20 MiB solo con su huella; copia en `~/.local/share/kitlegal/evidencias/<hito>/`) y commitea `chore(<hito>): grabaciones de …`. Los ficheros de `data/` derivados los produce código a partir de lo grabado.
- **Converge**: cuando no queda ninguna tarea pendiente, `converge` compara el código con los artefactos una sola vez y añade a `tasks.md` lo que falte; esas tareas pasan por el mismo bucle.

Para que la verificación por tarea tenga sentido, `tasks` recibe la regla de que cada tarea es una rebanada vertical (test + implementación) que deja `make ci` en verde por sí sola, con todas sus rutas declaradas.

## Del make ci del hito a la revisión

- **`ci`** ejecuta `verificar.sh global` con la suite congelada ya activada. En rojo, `bucle_reparacion` (hasta tres intentos) llama a `reparar_hito` y `cerrar_reparacion`.
- **Pasos globales** (`scripts/workflow/global.sh`): `base` fija el commit de partida y `cerrar` verifica desde él con el guardián global —nadie toca `testdata/`, `schemas/`, `evidencias/`, la configuración de verificación (`.golangci.yml`, `codecov.yml`, `lefthook.yml`, `.github/`, `tools/`, `scripts/workflow/`, `.specify/`, `.claude/`) ni la suite congelada— y `make ci`. En verde, commitea. Si el guardián lo rechaza, **aparta** todo lo hecho desde la base como `gates/aparcado-N.patch` y vuelve a ella. Con `make ci` en rojo, commitea marcándolo en el mensaje para que el paso siguiente lo vea, salvo en el barrido, que se aparta.
- **Barrido** (`barrido`, `modelo_redaccion`): tras H6, donde la revisión final dio 13 rondas por afirmaciones que el producto ya no sostenía, un paso contrasta de una vez los artefactos del hito y la documentación del repositorio (README, CONTRIBUTING, CHANGELOG, `docs/`, `SKILL.md`, comentarios de paquete, cifras y ejemplos) con el código y el binario, y corrige el texto, nunca el código.

## Revisión final con dos jueces

`preparar_jueces` genera los dos papeles y `revision_jueces` (paso `fan-out`, `max_concurrency: 2`) lanza la plantilla `revision_juez` una vez por juez, **en paralelo** y sin que ninguno vea el veredicto del otro: el item `a` (`modelo_revisor`) evalúa la Definition of Done con rúbrica; el item `b` (`modelo_adversario`, otra familia de modelo) parte de la hipótesis contraria y busca atajos, fixtures o guiones congelados retocados, datos de `data/` que no salen de lo grabado, tests vacíos, promesas que el binario no cumple y alcance excedido, mutando el código en una copia desechable. Ninguno modifica ficheros ni ejecuta `make ci` en el árbol. `leer_revision` archiva los dos veredictos de la ronda (`gates/revision-{a,b}-r<n>.json`): ambos aprueban → sigue; cualquier rechazo → `corrector_revision` aplica la **unión** de los motivos, `cerrar_correccion` verifica y commitea, y nueva ronda (hasta cuatro veredictos). `cerrar_revision` commitea los veredictos; si las rondas se agotan sin aprobar, los motivos van a `gates/revision-pendiente.md` y al informe, y el run sigue.

La regla «desacuerdo → humano» se retiró con datos de H1 (ADR 0007): el desacuerdo medía la profundidad de cada juez, no la ambigüedad del código; por eso los dos jueces tienen la misma capacidad y la diversidad está en el prompt y en la familia de modelo.

## Cierre en la plataforma e informe final

El cierre va **después** de la revisión, sobre la cabeza que se va a fusionar: en H5 y H6 la ejecución de cierre de las evals era una tarea `[plataforma]` anterior a la revisión, y cada corrección la dejaba sin cubrir la cabeza. Ya no existen las tareas `[plataforma]`.

- **`publicar`** (`scripts/workflow/cierre.sh publicar`) commitea los registros del directorio del feature, empuja la rama y abre la propuesta de cambio hacia `main` si no existe.
- **`medir_cierre`** (`cierre.sh medir`) quita y vuelve a poner la etiqueta `evals` (el botón de «vuelve a medir» de `.github/workflows/evals.yml`), espera a que terminen todas las comprobaciones de GitHub Actions sobre la cabeza (hasta 3 h) y escribe `gates/cierre.json` y, de cada ejecución en rojo, el final de `gh run view --log-failed` en `gates/cierre.log`. Los estados de Codecov son informativos. En rojo, `reparar_cierre` (`modelo_escalada`) arregla, `cerrar_cierre` verifica y commitea, y se vuelve a medir (hasta tres mediciones).
- **`informe_final`** (`scripts/workflow/informe.sh`) escribe sin modelo `gates/informe-final.md`, lo commitea, lo empuja y lo pone de cuerpo de la propuesta: estado local y remoto; trazabilidad de cada FR/SC del spec a sus tareas (con su estado) y a los guiones de aceptación; supuestos (clarify, correctores, rondas agotadas); cuarentena; lo que la capa 3 reserva a la persona; commits posteriores a la revisión; duración del run. `scripts/workflow/informe.sh <hito> --solo-ver` lo muestra sin tocar nada.

## Modelo por paso

Cada paso `command` y `prompt` lleva `model: "{{ inputs.modelo_<rol> }}"`. El valor es `<modelo>` o `<modelo>@<esfuerzo>`: alias (`fable`, `opus`, `sonnet`, `haiku`) o nombre completo, y esfuerzo `low`, `medium`, `high`, `xhigh` o `max` (Haiku 4.5 no admite esfuerzo). spec-kit solo sabe pasar `--model`, así que `scripts/hito.sh` exporta `SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE=scripts/claude-modelo.sh`: el wrapper valida el valor con una expresión estricta y lo traduce a `claude … --model <modelo> --effort <esfuerzo>`. **Sin el wrapper, `claude` rechaza los valores con `@`**: lanza siempre con `scripts/hito.sh` o exporta esa variable.

Los roles agrupan pasos por el tipo de trabajo, no por fase:

| Input | Pasos | Por defecto | Razón |
|---|---|---|---|
| `modelo_decision` | `resolver_clarify`, `plan`, `corrector_plan` | `opus@xhigh` | Decisiones cuyos errores se arrastran a todas las tareas; sin persona en medio, la lectura conservadora tiene que ser la buena |
| `modelo_juez` | `juez_spec`, `juez_plan`, `juez_tasks` | `opus@xhigh` | El juez debe ser al menos tan capaz como el autor; el de entrada es el único que puede devolver el hito |
| `modelo_revisor` | `revision_juez` (item `a`) | `opus@xhigh` | Rúbrica DoD casi toda comprobable |
| `modelo_adversario` | `revision_juez` (item `b`) | `fable@xhigh` | Último gate antes del cierre: otra familia que `modelo_revisor` para que los dos votos no fallen a la vez (ADR 0007) |
| `modelo_redaccion` | `specify`, `clarify_preguntas`, `tasks`, `corrector_spec`, `corrector_tasks`, `barrido` | `opus@high` | Artefactos largos con muchas reglas; los correctores y el barrido aplican motivos concretos o contrastan texto con código |
| `modelo_implementacion` | `implementar_tarea` (intento 1), `corrector_revision` | `opus@xhigh` | Código y depuración; `xhigh` es el nivel recomendado para trabajo agéntico de código |
| `modelo_escalada` | `implementar_tarea_escalada` (intentos 2 y 3), `reparar`, `reparar_hito`, `reparar_cierre` | `fable@xhigh` | Solo actúa cuando algo ya falló: no repetir con el mismo modelo lo que acaba de salir mal |
| `modelo_analisis` | `clarify_integrar`, `analyze`, `converge` | `sonnet@high` | Lectura, contraste y transformación de artefactos |

Criterio de coste, medido con `scripts/coste-run.sh` en los runs de H2 (`bfa8c3ac`) y H3 (`94612c4d`), a precio de lista: estos pasos son sobre todo lectura de contexto en caché, y Fable 5.1 lee de caché a mitad de precio que Opus 5, así que con el mismo perfil de tokens un paso en Fable cuesta entre 1,4 y 1,7 veces lo que en Opus, no el doble del precio nominal. Hasta el workflow 1.7.0 Fable (jueces, decisión y escalada) era el 44 % del coste de H3; los jueces en Fable no ahorraron rondas frente a los de Opus de H1, y las escaladas de H2 y H3 no fueron límites de capacidad. Desde 1.8.0 Fable queda donde la estructura lo pide (juez adversarial y escalada). Para revisarlo con datos nuevos: `scripts/coste-run.sh` da el coste por paso y por rol de cada run.

Sobrescritura por run:

```bash
KITLEGAL_MODELO_IMPLEMENTACION=opus@max KITLEGAL_MODELO_ANALISIS=haiku scripts/hito.sh H7
```

Los pasos `shell` no usan modelo. Para fijar un modelo distinto en un solo paso sin tocar los inputs, edita su `model:` en el YAML o usa un overlay (`specify workflow overlay add …`).

## Uso

```bash
scripts/hito.sh H7                     # lanza el hito y lo lleva hasta el informe final
scripts/hito.sh --resume <run_id>      # tras resolver una causa mayor (el dossier dice qué hace falta)
scripts/hito.sh --clasificar <run_id>  # qué haría el supervisor con ese run
KITLEGAL_MODELO_FALLBACK=fable=opus,opus=sonnet KITLEGAL_TIEMPO_MAXIMO=259200 scripts/hito.sh H7
scripts/workflow/informe.sh H7 --solo-ver   # el informe tal como está ahora, sin tocar nada
scripts/paso.sh juez_plan H7           # relanzar a mano un paso prompt (con el run parado)
scripts/paso.sh revision_juez_b H7     # un juez del fan-out de la revisión final (items a y b)
scripts/coste-run.sh [run_id]          # coste por paso y por rol (por defecto, el run más reciente)
specify workflow status [<run_id>]     # runs y estado
```

Estado de cada run en `.specify/workflows/runs/<run_id>/` (`state.json`, `inputs.json`, `log.jsonl`).

## Supervisor

`specify workflow run` termina en cuanto un paso falla, y el motor no reintenta nada. `scripts/hito.sh` envuelve `run`/`resume` en un bucle que clasifica cada parada leyendo `state.json`, `log.jsonl` y el transcript de la sesión headless (`~/.claude/projects/<ruta>/`), y actúa con una lista cerrada de acciones:

| Clase | Cómo se reconoce | Acción |
|---|---|---|
| `completado` | `status: completed` | Termina y avisa: el informe es el cuerpo de la propuesta |
| `rechazo` | Falla `check_gate_spec` | Se detiene: informe de rechazo en `gates/informe-rechazo.md` |
| `causa_mayor` | Un paso shell sale con 3 | Se detiene: dossier en `gates/dossier.md` |
| `limite` | Paso `prompt`/`command` fallido y la sesión headless posterior a su inicio terminó con `rate_limit`, «spend limit», «usage limit», `overloaded`, `api_error` o «API Error: NNN» | Cambia de familia de modelo (`KITLEGAL_MODELO_FALLBACK`, por defecto `fable=opus,opus=sonnet`, conservando el esfuerzo) en todos los roles que usaban la familia que falló, espera `KITLEGAL_ESPERA_LIMITE` s (60) y reanuda; sin repuesto, espera `KITLEGAL_ESPERA_SIN_REPUESTO` s (1800) y reanuda con los mismos modelos. No consume reanudaciones |
| `transitorio` | Cualquier otro fallo de un paso | Reanuda; el mismo paso no se reanuda más de dos veces, y hay un tope de `KITLEGAL_MAX_REANUDACIONES` (8) por invocación |

Antes de cada reanudación comprueba el presupuesto de tiempo (`KITLEGAL_TIEMPO_MAXIMO`, 48 h desde el primer paso del run). Al arrancar comprueba `claude`, la sesión de `gh` y el acceso a `origin`, y toma el candado de sesión única. Mientras vive, mantiene un `caffeinate -i` ligado a su PID: en H3 el Mac se durmió a mitad de `plan` y el hito tardó casi dos horas más. El supervisor nunca edita artefactos, veredictos, `tasks.md` ni permisos: solo relanza `specify workflow resume`. En macOS avisa con una notificación de escritorio al terminar o detenerse.

## Una sola sesión por run

En H6 una copia duplicada de la conversación estuvo trabajando en paralelo sobre el mismo árbol que el run. `scripts/hito.sh` toma un candado en el git-dir del árbol de trabajo (`kitlegal-run.lock`, uno por worktree) con su PID y un testigo aleatorio, y exporta el testigo en `KITLEGAL_RUN_TESTIGO`: lo heredan `specify`, los pasos shell y las sesiones `claude -p` del run. El gancho `PreToolUse` de `.claude/settings.json` (`scripts/workflow/sesion-unica.sh gancho`) rechaza, en cualquier sesión sin ese testigo, `Edit`, `Write`, `MultiEdit`, `NotebookEdit` y las órdenes de `Bash` que cambian el árbol o el historial (`git add|commit|checkout|reset|push…`, `rm`, `mv`, `sed -i`, `scripts/hito.sh`, `scripts/paso.sh`…); leer y `scripts/hito.sh --clasificar` siguen permitidos. `scripts/paso.sh` se niega a correr con un run vivo. Un candado cuyo PID ya no existe se ignora y se recupera.

## Permisos de Claude en modo headless

`claude -p` respeta `.claude/settings.json`. Para que `implement` pueda compilar y testear sin prompts el proyecto necesita una lista de permisos como la del fichero (`deny` gana a `allow`; como `Bash(git:*)` ya permite `git push`, la política se expresa en `deny`: `main`, push forzados, borrados, etiquetas y el `git push` sin argumentos quedan denegados; la garantía real la da `scripts/lefthook/pre-push/guardia-push.sh`, que ve las referencias que git va a enviar). Desde la 2.0.0 publicar y medir en la plataforma lo hacen los pasos shell `publicar`, `medir_cierre` e `informe_final`, que no pasan por los permisos de Claude, y los prompts de los pasos con modelo prohíben `git push` y `gh`. Los permisos siguen admitiendo `git push` de ramas y `gh pr create|edit|view|checks` porque los usan las sesiones interactivas; lo que nunca admite ninguna sesión (fusionar, empujar a `main`, forzar, borrar, etiquetar) lo sigue garantizando el gancho `pre-push`.

`scripts/hito.sh` exporta `SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS="--permission-mode acceptEdits"`. En un entorno aislado (contenedor o VM) puede sustituirse por `--dangerously-skip-permissions`.

## Limitaciones conocidas

- Los pasos `command` transmiten la salida al terminal y no la capturan; por eso los jueces escriben ficheros en `gates/` y `analyze` recibe la instrucción de guardar su informe.
- Los pasos `shell` tienen `timeout` explícito (1800 s los que ejecutan `make ci`, 3600 s `grabar_datos`, 11 400 s `medir_cierre`); los pasos `prompt`, 3600 s. Un corte dentro de una ronda lo trata el supervisor como transitorio, y `resume` repite la ronda entera.
- **`resume` reejecuta el paso de nivel superior**, no el paso anidado que falló (`engine.py`: «resume will re-run the parent step and its nested body»). Dentro de `bucle_tareas` significa una **iteración nueva**: `siguiente_tarea` detecta una tarea marcada `[X]` con cambios sin commitear y emite una iteración de **cierre** (sin implementar: verificación y commit con la base original). Las rondas cuentan en `gates/<fase>-rondas`, que solo pone a cero el paso `iniciar_*` anterior al bucle; tras un `resume` con las rondas agotadas, el bucle ejecuta un solo juez y `cerrar` decide.
- **El motor conserva la última salida de cada id de paso entre iteraciones del bucle**, y una condición que lea la de un paso que no se ejecutó en esta iteración lee la de la anterior (H4, T003: 145 vueltas). Regla para el YAML: una condición solo puede leer salidas de pasos que se ejecutan incondicionalmente en la misma iteración; `estado_cierre` decide el commit con la última línea de `gates/ci.log` y con `tasks.md`.
- **Cada run congela el workflow**: `specify workflow run` copia la definición a `.specify/workflows/runs/<run_id>/workflow.yml` y `resume` la lee de ahí. Los scripts de `scripts/workflow/` no se congelan: un cambio en ellos afecta a los runs en curso.
- El guardián extrae rutas de la línea de la tarea: tokens con `/`, con extensión conocida, ficheros de raíz sin extensión que terminan en `ignore`, `.editorconfig`, `Makefile` y `LICENSE`. Una tarea que toque muchos ficheros debe declarar directorios.
- El supervisor reconoce el límite de uso por el texto del transcript; un error nuevo con otra redacción se clasifica como `transitorio`. Las sesiones headless se reconocen por el entrypoint `sdk-cli`, que fija `scripts/claude-modelo.sh` (para runs anteriores a esa corrección: `KITLEGAL_ENTRYPOINTS_HEADLESS=sdk-cli,claude-vscode`).
- La batería por tarea ejecuta `make ci` completo tras cada tarea; en hitos grandes es lento pero determinista.
- `inputs.hito` se interpola en pasos `shell`; está restringido por `enum`. No añadir inputs libres a pasos `shell`.
- Si `speckit init` se actualiza (`specify integration upgrade`), regenera `.claude/skills/speckit-*`; el workflow, `scripts/workflow/` y la constitución no se tocan.
