# Workflow por hito con spec-kit

Cada hito de `ROADMAP.md` se implementa con una pasada del workflow `hito` de spec-kit (`.specify/workflows/hito/workflow.yml`). El workflow encadena los comandos `speckit.*` en modo headless (`claude -p`) e intercala **gates automáticos** que evalúa Claude con la sección «Criterio de decisión autónoma» de `.specify/memory/constitution.md`.

## Piezas

| Pieza | Dónde | Para qué |
|---|---|---|
| Constitución | `.specify/memory/constitution.md` | Principios, restricciones, DoD y el criterio con el que se decide sin humano |
| Workflow | `.specify/workflows/hito/workflow.yml` | Secuencia, gates, bucle de reparación de CI |
| Lanzador y supervisor | `scripts/hito.sh` | Comprueba `main` limpio, exporta flags de Claude y el wrapper de modelo, lanza el run y lo supervisa: reanuda ante límites de uso y fallos transitorios, se detiene ante gates y paradas deliberadas |
| Guardia de push | `scripts/lefthook/pre-push/guardia-push.sh` | Gancho `pre-push` de lefthook: rechaza `main`, push forzados, borrados y etiquetas, sea cual sea la orden |
| Wrapper de modelo | `scripts/claude-modelo.sh` | Ejecutable de Claude para spec-kit: traduce `--model <modelo>@<esfuerzo>` a `--model` + `--effort` |
| Coste por paso | `scripts/coste-run.sh` | Consumo y coste estimado de un run por paso y por rol, desde los transcripts |
| Extensión git | `.specify/extensions/git/` | Rama `NNN-hN-slug` por hito y auto-commit (Conventional Commits) tras cada fase |
| Skills | `.claude/skills/speckit-*` | Los comandos `/speckit-*` (generados por `specify init`, no editar a mano) |
| Artefactos | `specs/NNN-hN-slug/` | `spec.md`, `plan.md`, `research.md`, `tasks.md`, `checklists/`, `gates/*.json` |

## Secuencia

```
extraer_hito (shell: sección del hito en ROADMAP.md → JSON)
→ specify (modo desatendido: alcance = sección; ambigüedades como [NEEDS CLARIFICATION], sin resolver)
→ clarify_preguntas (command: solo formula preguntas, sin recomendación → gates/clarify-preguntas.json)
→ resolver_clarify (prompt en PROCESO NUEVO, contexto limpio → gates/clarify-respuestas.json)
→ check_clarify (shell: todas respondidas; ninguna escalada) → clarify_integrar (command: integra respuestas en spec.md)
→ precheck_spec (shell) → ronda_spec ×2 [juez_spec → leer_gate_spec → corrector_spec si rechazado y corregible] → check_gate_spec
→ [supervisado] gate humano
→ plan → precheck_plan → ronda_plan ×2 [juez → leer → corrector] → check_gate_plan → [supervisado] gate humano
→ tasks → analyze → precheck_tasks (formato, ids, rutas declaradas, [datos]) → ronda_tasks ×2 → check_gate_tasks
→ bucle por tarea (do-while):
     siguiente_tarea (shell: primera "- [ ] Tnnn", intentos, base git, rutas declaradas, [datos] → gates/tarea-actual.json)
     → si la tarea es [plataforma] y .claude/settings.json no permite push ni gh: gate humano antes de gastar intentos
     → implementar_tarea (command implement, solo esa tarea; intento 1 con modelo_implementacion, reintentos con modelo_escalada; solo una tarea [plataforma] puede empujar la rama y abrir la PR)
     → guardian_diff (shell: rutas declaradas; testdata/ y schemas/ solo con [datos])
     → verificar (shell: make ci | go build+vet+test; log en gates/ci.log)
     → si falla: reparar (prompt con las últimas 80 líneas del log) → guardian_diff_reparacion → verificar_reparacion
     → commit_tarea → clasificar_datos → gate humano si la tarea [datos] modificó material existente de testdata/ o schemas/ o añadió fixtures o esquemas (independiente de `modo`)
→ ci (make ci del hito) → si falla: do-while (reparar_hito + ci_reintento) ×3 → ci_tras_reparacion
→ converge → implement_restante
→ ronda_revision ×3 [preparar_jueces → fan-out en paralelo: juez A (DoD) ∥ juez B (adversarial) → leer_revision → corrector sobre la unión de motivos si algún rechazo es corregible]
→ ci_final (commitea los veredictos; make ci; ambos jueces aprobado; sin tareas pendientes; árbol limpio)
→ rutas_sensibles → gate humano forzado si el diff toca docs/SOURCES.md, data/anomalias/ o una fuente nueva
→ publicar_rama (git push -u origin <rama>; gh pr create hacia main si no existe propuesta; nunca merge)
→ [supervisado] gate humano final
```

La fusión a `main` (squash-merge) y el release son siempre acciones humanas. El workflow deja la rama del hito commiteada y en verde, la empuja a `origin` y abre la propuesta de cambio si no existe (`publicar_rama`; antes, la tarea `[plataforma]` del hito si la hay). Nunca hace `merge`, ni push a `main`, ni push forzado: además de los permisos de `.claude/settings.json`, `scripts/lefthook/pre-push/guardia-push.sh` (gancho `pre-push` de lefthook, instalado con `make hooks` o `lefthook install`) rechaza esas referencias aunque el YAML cambiara. Una persona lo salta con `KITLEGAL_PUSH_HUMANO=1` (etiquetas de release). Motivación y datos en `docs/ADR/0007-workflow-desatendido.md`.

## Tres capas de gate

Sigue la sección «Gates» de la constitución: cada comprobación vive en la capa más baja que pueda verificarla.

1. **Mecánica (shell, sin LLM).** `precheck_*` corre antes de cualquier juez: marcadores pendientes, checklists, sección "Fuera de alcance", formato e ids de tareas, rutas declaradas, etiqueta `[datos]`. `verificar`/`ci` ejecutan `make ci`, que debe encadenar lint, `-race`, schema-check, drift de `references/`, test de arquitectura y golden files de citas. `leer_gate_*` valida el JSON del juez y su coherencia (aprobado ⇔ todos los criterios cumplen).
2. **Juez LLM con rúbrica.** `juez_spec`, `juez_plan`, `juez_tasks`, `revision_juez_a` y `revision_juez_b` reciben criterios fijos (`a`…`i`) y escriben `{"veredicto","corregible","criterios":[{id,criterio,cumple,evidencia}],"motivos"}`. **No corrigen nada.** Revisan de forma exhaustiva (todos los incumplimientos en el primer veredicto, no uno por ronda) y, en rondas posteriores, comprueban primero los motivos anteriores. Si rechazan y el motivo es corregible, un `corrector_*` (proceso distinto) aplica los motivos, generalizando cada uno a la clase de defecto que lo causa, y se vuelve a juzgar, con tope de dos correcciones. Si el motivo requiere decisión humana (`corregible: false`), el `check_gate_*` para el run. Los autores (`specify`, `plan`, `tasks`) reciben la orden de comprobar su artefacto contra la rúbrica del juez antes de entregar: en H1 los tres gates se rechazaron exactamente una vez, y esa ronda es evitable sin bajar el listón porque el juez sigue juzgando igual.
3. **Humano.** Pausas que no dependen de `modo` ni admiten pre-aprobación por input: tras una tarea `[datos]` que modifique material existente de `testdata/` o `schemas/` o añada fixtures grabados o esquemas (`clasificar_datos`: ficheros nuevos bajo `testdata/` de raíz, `internal/source/**` o `schemas/`; el material de test nuevo fuera de ese territorio, como código Go o guiones `.txtar` bajo `internal/<pkg>/testdata/`, sigue exigiendo `[datos]` y guardián pero no pausa: lo revisan los jueces finales), tras una tarea `[plataforma]` cuando los permisos de `.claude/settings.json` no permiten `git push` ni `gh pr create`, y al final si el diff toca `docs/SOURCES.md`, `data/anomalias/` o crea un directorio nuevo bajo `internal/source/`. El supervisor se detiene y avisa; se reanudan con `scripts/hito.sh --resume <run_id>` desde un terminal, que pregunta approve/reject.

## Clarificación sin sesgo

`clarify` se divide en tres pasos para que quien formula las preguntas no sea quien las responde:

1. **`clarify_preguntas`** ejecuta `/speckit-clarify` con la orden de solo escribir las preguntas y sus opciones en `gates/clarify-preguntas.json`, sin recomendación ni opción preferida, y sin tocar el spec.
2. **`resolver_clarify`** es un `prompt` que spec-kit lanza como un proceso `claude -p` nuevo: no comparte conversación con el paso anterior y se le dice que decida solo con documentos (hito, `CLAUDE.md`, `refs/`, constitución). Aplica el criterio en este orden: (a) lo determinan las fuentes → esa es la respuesta; (b) no está especificado → "fuera de alcance: no se implementa"; (c) varias opciones válidas → la de mayor calidad y mejores prácticas, con la alternativa rechazada; (d) alcance, frontera humana, privacidad, TOS, anomalías o decisión cerrada → `escalar: true`.
3. **`check_clarify`** detiene el run si alguna respuesta está escalada. Se edita `gates/clarify-respuestas.json` a mano y se reanuda. **`clarify_integrar`** vuelve a llamar a `/speckit-clarify` solo para integrar los pares Q/A en `spec.md` con marca `(auto: criterio, fuente)`.

## Implementación tarea a tarea y guardián de diff

Con `granularidad=tarea` (valor por defecto) el bucle toma la primera línea `- [ ] Tnnn` de `tasks.md`, anota el commit base y las rutas que la tarea declara, lanza `/speckit-implement` restringido a esa tarea y después:

- **`guardian_diff`** compara `git diff --name-only <base>` más los ficheros nuevos con las rutas declaradas. Falla, y el run se para, si un fichero queda fuera (scope creep) o si toca `testdata/` o `schemas/` sin etiqueta `[datos]` (arreglar el test en vez del código, grabar fixtures en el bucle). Siempre permitidos: `go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature, y `x_test.go` cuando se declara `x.go`. Declarar un directorio (`internal/cli/`) permite todo lo que cuelga de él.
- **`verificar`** ejecuta la batería determinista. Si falla, `reparar` recibe **las últimas 80 líneas de la salida** en el prompt y la ruta del log completo, con la orden de arreglar la causa y no el control; después vuelven a pasar el guardián y la verificación. Si sigue en rojo, el run se detiene.
- Una tarea que no queda marcada `[X]` tras 3 intentos detiene el run (`gates/tareas-intentos.json`, `gates/tarea-Tnnn.md`).
- Una tarea `[datos]` pasa por `clasificar_datos` tras el commit: gate humano solo si modificó material existente de `testdata/` o `schemas/` o añadió fixtures grabados o esquemas.
- Una tarea `[plataforma]` (evidencia en la propuesta de cambio, estados de CI o Codecov) es la única en la que el ejecutor puede hacer `git push -u origin <rama>` y usar `gh pr create|view|checks` y `gh run list`; va la última de `tasks.md` y `precheck_tasks` exige la etiqueta a toda tarea que nombre `gh pr`, `gh run`, «pull request» o «propuesta de cambio». Si `.claude/settings.json` conserva `Bash(git push:*)` en `deny` o no permite `gh pr create`, `siguiente_tarea` lo detecta y el run pausa en un gate antes de gastar intentos (en H1, T021 quemó sus tres intentos por esto).

Para que la verificación por tarea tenga sentido, `tasks` recibe la regla de que cada tarea es una rebanada vertical (test + implementación) que deja `make ci` en verde por sí sola, con todas sus rutas declaradas, y `precheck_tasks` + `juez_tasks` lo comprueban. Mientras no exista `Makefile` con objetivo `ci`, la batería es `go build ./... && go vet ./... && go test -race ./...`; sin `go.mod` no hay nada que verificar (primeras tareas de H0).

`granularidad=hito` conserva la pasada única de `implement` sin guardián por tarea (más barata, menos control).

## Revisión final con dos jueces

`preparar_jueces` genera los dos papeles y `revision_jueces` (paso `fan-out` del motor, `max_concurrency: 2`) lanza la plantilla `revision_juez` una vez por juez, **en paralelo** y sin que ninguno vea el veredicto del otro: el item `a` (`modelo_revisor`) evalúa la Definition of Done con rúbrica; el item `b` (`modelo_juez`, otra familia de modelo para que los dos votos no compartan puntos ciegos) parte de la hipótesis contraria y busca evidencia de atajos, fixtures retocados, tests vacíos, promesas de los contratos que el binario no cumple, alcance excedido y violaciones que el linter no ve, mutando el código en una copia desechable para comprobar que los tests detectan lo que dicen. Ninguno modifica ficheros ni ejecuta `make ci` en el árbol (corren a la vez). `leer_revision` combina: ambos aprueban → sigue; cualquier rechazo con `corregible: true` → `corrector_revision` aplica la **unión** de los motivos y nueva ronda (hasta tres veredictos y dos correcciones); algún motivo con `corregible: false` o rondas agotadas → `ci_final` para el run para un humano.

La regla anterior («desacuerdo → humano») se retiró con datos de H1: en la ronda 1 el juez A (`sonnet@max`, 3 min) no encontró nada y el B (`opus@max`, 26 min) encontró 13 defectos reales y corregibles; en la ronda 2 se invirtió (A cinco motivos nuevos, B aprobado). El desacuerdo medía la profundidad de cada juez, no la ambigüedad del código; por eso los dos jueces deben tener la misma capacidad y la diversidad se pone en el prompt y en la familia de modelo. Detalle en `docs/ADR/0007-workflow-desatendido.md`.

## Modelo por paso

Cada paso `command` y `prompt` lleva `model: "{{ inputs.modelo_<rol> }}"`. El valor es `<modelo>` o `<modelo>@<esfuerzo>`: alias (`fable`, `opus`, `sonnet`, `haiku`) o nombre completo, y esfuerzo `low`, `medium`, `high`, `xhigh` o `max` (Haiku 4.5 no admite esfuerzo). spec-kit solo sabe pasar `--model`, así que `scripts/hito.sh` exporta `SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE=scripts/claude-modelo.sh`: el wrapper valida el valor con una expresión estricta y lo traduce a `claude … --model <modelo> --effort <esfuerzo>`. **Sin el wrapper, `claude` rechaza los valores con `@`**: lanza siempre con `scripts/hito.sh` o exporta esa variable.

Los roles agrupan pasos por el tipo de trabajo, no por fase:

| Input | Pasos | Por defecto | Razón |
|---|---|---|---|
| `modelo_decision` | `resolver_clarify`, `plan`, `corrector_plan` | `fable@high` | Decisiones cuyos errores se arrastran a todas las tareas. El plan fija herramientas, versiones y CI; su corrector necesita ver la premisa equivocada, no solo la línea citada |
| `modelo_juez` | `juez_spec`, `juez_plan`, `juez_tasks`, `revision_juez` (item `b`) | `fable@xhigh` | El juez debe ser al menos tan capaz como el autor; `xhigh` para que el primer veredicto sea exhaustivo y ahorre rondas |
| `modelo_revisor` | `revision_juez` (item `a`) | `opus@xhigh` | Rúbrica DoD casi toda comprobable; un modelo distinto de `modelo_juez` hace que los dos votos no fallen a la vez |
| `modelo_redaccion` | `specify`, `clarify_preguntas`, `tasks`, `corrector_spec`, `corrector_tasks` | `opus@high` | Artefactos largos con muchas reglas; los correctores aplican motivos concretos del juez |
| `modelo_implementacion` | `implementar_tarea` (intento 1), `implement`, `implement_restante` | `opus@xhigh` | Código y depuración; `xhigh` es el nivel recomendado para trabajo agéntico de código |
| `modelo_escalada` | `implementar_tarea_escalada` (intentos 2 y 3), `reparar`, `reparar_hito`, `corrector_revision` | `fable@xhigh` | Solo actúa cuando `modelo_implementacion` ya falló: no repetir con el mismo modelo lo que acaba de salir mal. No cuesta nada si todo va bien |
| `modelo_analisis` | `clarify_integrar`, `analyze`, `converge` | `sonnet@high` | Lectura, contraste y transformación de artefactos; los fallos de `converge` los cubren los dos jueces finales |

Criterio de coste: estos pasos son sobre todo lectura de contexto en caché, y Fable 5.1 lee de caché a mitad de precio que Opus 5. Con el mismo perfil de tokens, pasar un paso de Opus a Fable lo encarece un 5-30 %, no el doble del precio nominal; en los jueces, que escriben más, en torno al 60 %. Medido en el run de H0 con `scripts/coste-run.sh`.

Sobrescritura por run:

```bash
KITLEGAL_MODELO_IMPLEMENTACION=opus@max KITLEGAL_MODELO_ANALISIS=haiku scripts/hito.sh H0
SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE=$PWD/scripts/claude-modelo.sh \
  specify workflow run hito -i hito=H0 -i modelo_juez=fable@max
```

Los pasos `shell` no usan modelo. Para fijar un modelo distinto en un solo paso sin tocar los inputs, edita su `model:` en el YAML o usa un overlay (`specify workflow overlay add …`).

Para ajustar la asignación con datos, `scripts/coste-run.sh [run_id]` reconstruye desde los transcripts de Claude Code (spec-kit no conserva la salida de `claude -p`) las sesiones, turnos, tokens de salida y coste estimado de cada paso y de cada rol del run.

## Uso

```bash
scripts/hito.sh H0                     # desatendido, tarea a tarea
scripts/hito.sh H0 supervisado         # con pausas humanas
SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE=$PWD/scripts/claude-modelo.sh \
  specify workflow run hito -i hito=H0 -i granularidad=hito   # implement en una pasada
specify workflow status                # runs y estado
specify workflow status <run_id>
scripts/hito.sh --resume <run_id>      # tras corregir a mano un artefacto
scripts/hito.sh --resume <run_id> veredicto_plan=approve
scripts/hito.sh --clasificar <run_id>  # qué haría el supervisor con ese run (completado, gate, deliberado, transitorio, limite)
KITLEGAL_MODELO_FALLBACK=fable=opus,opus=sonnet KITLEGAL_MAX_REANUDACIONES=8 scripts/hito.sh H2
scripts/paso.sh juez_plan H0           # relanzar a mano un paso prompt (juez, corrector) con su modelo
scripts/paso.sh corrector_plan H0 opus@xhigh
scripts/paso.sh revision_juez_b H2     # un juez del fan-out de la revisión final (items a y b)
scripts/coste-run.sh [run_id]          # coste por paso y por rol (por defecto, el run más reciente)
specify workflow resolve hito          # ver el workflow compuesto con overlays
```

Estado de cada run en `.specify/workflows/runs/<run_id>/` (`state.json`, `inputs.json`, `log.jsonl`).

## Supervisor

`specify workflow run` termina en cuanto un paso falla o un gate pausa, y el motor no reintenta nada. En el run de H1 eso costó 3,4 h de huecos en cuatro paradas, de las que solo una era una decisión humana. `scripts/hito.sh` envuelve `run`/`resume` en un bucle que clasifica cada parada leyendo `state.json`, `log.jsonl` y el transcript de la sesión headless (`~/.claude/projects/<ruta>/`), y actúa con una lista cerrada de acciones:

| Clase | Cómo se reconoce | Acción |
|---|---|---|
| `completado` | `status: completed` | Termina y avisa |
| `gate` | `status: paused` | Se detiene: la pausa es humana por construcción |
| `limite` | Paso `prompt`/`command` fallido y la sesión headless posterior a su inicio terminó con `rate_limit`, «spend limit», «usage limit», `overloaded`, `api_error` o «API Error: NNN» (un `500` suelto no basta: en H3 un juez escribió «1 500 líneas») | Cambia de familia de modelo (`KITLEGAL_MODELO_FALLBACK`, por defecto `fable=opus,opus=sonnet`, conservando el esfuerzo) en **todos** los roles que usaban la familia que falló, espera `KITLEGAL_ESPERA_LIMITE` s (60) y reanuda con `--input modelo_<rol>=…` |
| `transitorio` | Paso `prompt`/`command` fallido sin límite de uso, o paso shell `commit_*` fallido | Reanuda una vez; si el mismo paso vuelve a fallar, se detiene |
| `deliberado` | Cualquier otro shell fallido (`precheck_*`, `check_*`, `leer_*`, `guardian_*`, `siguiente_tarea` con intentos agotados, `verificar_reparacion`, `ci_tras_reparacion`, `ci_final`, `publicar_rama`) | Se detiene y muestra la salida del paso: son paradas a propósito |

Tope de `KITLEGAL_MAX_REANUDACIONES` (8) reanudaciones por invocación. El supervisor nunca edita artefactos, veredictos, `tasks.md` ni permisos: solo relanza `specify workflow resume`, con la misma lista de acciones para el humano que lo lee. En macOS avisa con una notificación de escritorio al detenerse. `scripts/hito.sh --clasificar <run_id>` muestra la clasificación sin actuar.

## Permisos de Claude en modo headless

`claude -p` respeta `.claude/settings.json`. Para que `implement` pueda compilar y testear sin prompts, y para que una tarea `[plataforma]` pueda empujar la rama del hito y abrir o leer la propuesta de cambio, el proyecto necesita una lista de permisos como esta (`deny` gana a `allow`; como `Bash(git:*)` ya permite `git push`, la política se expresa en `deny`: `main`, push forzados, borrados, etiquetas y el `git push` sin argumentos quedan denegados; la garantía real la da `scripts/lefthook/pre-push/guardia-push.sh`, que ve las referencias que git va a enviar):

```json
{
  "permissions": {
    "allow": [
      "Read", "Edit", "Write", "Glob", "Grep",
      "Bash(rtk:*)", "Bash(go:*)", "Bash(gofmt:*)", "Bash(gofumpt:*)", "Bash(goimports:*)",
      "Bash(golangci-lint:*)", "Bash(govulncheck:*)", "Bash(goreleaser:*)", "Bash(lefthook:*)",
      "Bash(gitleaks:*)", "Bash(make:*)", "Bash(git:*)", "Bash(jq:*)",
      "Bash(ls:*)", "Bash(cat:*)", "Bash(head:*)", "Bash(tail:*)", "Bash(wc:*)", "Bash(find:*)",
      "Bash(grep:*)", "Bash(rg:*)", "Bash(sed:*)", "Bash(awk:*)", "Bash(mkdir:*)", "Bash(ln:*)",
      "Bash(chmod:*)", "Bash(cp:*)", "Bash(mv:*)", "Bash(touch:*)", "Bash(date:*)",
      "Bash(specify:*)", "Bash(.specify/scripts/bash/*)", "Bash(.specify/extensions/git/scripts/bash/*)",
      "Bash(./kitlegal:*)", "Bash(./bin/kitlegal:*)", "Bash(rm:*)",
      "Bash(gh auth status:*)", "Bash(gh pr create:*)", "Bash(gh pr view:*)", "Bash(gh pr list:*)",
      "Bash(gh pr checks:*)", "Bash(gh pr diff:*)", "Bash(gh pr edit:*)",
      "Bash(gh run list:*)", "Bash(gh run view:*)", "Bash(gh run watch:*)"
    ],
    "deny": [
      "Bash(git push)", "Bash(git push origin main:*)", "Bash(git push -u origin main:*)", "Bash(git push origin HEAD:main:*)",
      "Bash(git push --force:*)", "Bash(git push -f:*)", "Bash(git push --force-with-lease:*)",
      "Bash(git push --delete:*)", "Bash(git push -d:*)", "Bash(git push origin --delete:*)", "Bash(git push origin :*)",
      "Bash(git push --all:*)", "Bash(git push --mirror:*)", "Bash(git push --tags:*)",
      "Bash(git merge:*)", "Bash(git reset --hard:*)",
      "Bash(gh pr merge:*)", "Bash(gh pr close:*)", "Bash(gh release:*)", "Bash(gh repo delete:*)",
      "Bash(rm -rf:*)", "Bash(curl:*)", "Bash(wget:*)", "Bash(sudo:*)"
    ]
  }
}
```

`scripts/hito.sh` exporta `SPECKIT_INTEGRATION_CLAUDE_EXTRA_ARGS="--permission-mode acceptEdits"`. En un entorno aislado (contenedor o VM) puede sustituirse por `--dangerously-skip-permissions`.

## Limitaciones conocidas

- Los pasos `command` transmiten la salida al terminal y no la capturan; por eso los jueces escriben ficheros en `gates/` y `analyze` recibe la instrucción de guardar su informe.
- Los pasos `prompt` y `shell` tienen `timeout` explícito (1800 s); `make ci` debe caber en ese margen.
- Un `shell` que falla detiene el run salvo `continue_on_error: true`; solo lo llevan los pasos cuyo fallo se enruta a una reparación (`verificar`, `ci`, `ci_reintento`). Los `precheck_*`, `check_*`, `guardian_diff*` y `leer_*` fallan a propósito para parar.
- Rondas juez → corrector: hasta tres veredictos y dos correcciones (`gates/<fase>-rondas` cuenta; el corrector no actúa en la tercera ronda), de modo que toda corrección se vuelve a juzgar. Si el tercer veredicto sigue rechazado, `check_gate_*` (o `ci_final`) para. Para una ronda extra a mano: `scripts/paso.sh juez_plan H0` y después `scripts/hito.sh --resume <run_id>`; los jueces finales se relanzan con `scripts/paso.sh revision_juez_a|b <hito>`, que resuelve el item del fan-out ejecutando `preparar_jueces`.
- Los dos jueces finales corren a la vez sobre el mismo árbol; por eso su prompt les prohíbe `make ci` y cualquier escritura fuera de un temporal. `scripts/coste-run.sh` atribuye las sesiones que arrancan en el solape por familia de modelo, así que conviene que `modelo_revisor` y `modelo_juez` sean de familias distintas.
- `precheck_tasks` exige `[plataforma]` a toda tarea que nombre `gh pr`, `gh run`, «pull request» o «propuesta de cambio»; una tarea que necesite la plataforma con otras palabras la detecta el criterio e del `juez_tasks`.
- El supervisor reconoce el límite de uso por el texto del transcript; un error nuevo con otra redacción se clasifica como `transitorio` y se reanuda una sola vez.
- Las sesiones headless se reconocen en los transcripts por el entrypoint `sdk-cli`, que `scripts/claude-modelo.sh` fija en el entorno: sin eso, un hito lanzado desde la extensión de VS Code (que exporta `CLAUDE_CODE_ENTRYPOINT=claude-vscode` a sus procesos hijos) dejaba sesiones que ni `limite_api` ni `scripts/coste-run.sh` reconocían (run 94612c4d de H3, 2026-09-12). Para runs grabados antes de esa corrección: `KITLEGAL_ENTRYPOINTS_HEADLESS=sdk-cli,claude-vscode`; las sesiones interactivas de la extensión comparten ese entrypoint y se descartan porque su primer mensaje lleva `origin.kind = human`, que las headless no traen. `coste-run.sh` agrupa en la fila «manual (paso.sh)» las sesiones headless que empiezan fuera de todo paso (rondas a mano tras una parada).
- Los hooks de auto-commit de la extensión git son opcionales y en headless no se ejecutan; los commits los hacen pasos `shell` deterministas: `commit_artefactos_plan`, `commit_artefactos_tasks`, `commit_tarea` (uno por tarea, mensaje `feat(Hn): Tnnn`), `commit_restante` y, al principio de `ci_final`, el de los veredictos de la revisión final (`gates/revision-{a,b}.json`, `gates/revision-rondas`). Este último vive dentro de `ci_final` porque un `--resume` tras re-juzgar a mano con `scripts/paso.sh` vuelve a ejecutar ese paso; sin él, `ci_final` fallaría por árbol sucio aunque los dos jueces aprobaran.
- El guardián de diff extrae rutas de la línea de la tarea: tokens con `/`, con extensión conocida, ficheros de raíz sin extensión que terminan en `ignore` (`.gitleaksignore`, `.gitignore`), `.editorconfig`, `Makefile` y `LICENSE`. Una tarea que toque muchos ficheros debe declarar directorios; un fichero de raíz con otro nombre extensionless no es declarable y hay que ampliar la expresión en el YAML.
- **`resume` reejecuta el paso de nivel superior**, no el paso anidado que falló (`engine.py`: «resume will re-run the parent step and its nested body»). Fuera de los bucles no se nota: `check_gate_spec` o `ci_final` se repiten tal cual. Dentro de `bucle_tareas` significa una **iteración nueva**: `siguiente_tarea` vuelve a elegir tarea y el trabajo sin commitear de la anterior contaría como diff de la siguiente. Desde 1.6.1 `siguiente_tarea` lo detecta —tarea de `gates/tarea-actual.json` marcada `[X]` y árbol sucio fuera del directorio del feature— y emite una iteración de **cierre** (`cierre: true`): sin implementar nada pasa por `guardian_diff` con la base original, `verificar` y `commit_tarea`, y solo después elige la siguiente. Procedimiento tras una parada dentro del bucle: arreglar la causa, dejar la tarea `[X]` si ya está en verde (o `[ ]` para que el bucle la reintente) y reanudar. `scripts/paso.sh` también lanza pasos `shell` (`guardian_diff`, `verificar`, `commit_tarea`) para cerrar una tarea a mano si hiciera falta.
- **Cada run congela el workflow**: `specify workflow run` copia la definición a `.specify/workflows/runs/<run_id>/workflow.yml` y `resume` la lee de ahí. Editar `.specify/workflows/hito/workflow.yml` a mitad de un run no afecta a ese run. Para aplicar un cambio a un run parado, copia la definición nueva sobre el snapshot y comprueba que `current_step_index` sigue apuntando al paso correcto (`state.json`), porque insertar pasos desplaza los índices.
- La batería por tarea ejecuta `make ci` completo tras cada tarea; en hitos grandes es lento pero determinista. Si hace falta, añadir un objetivo `make check` más rápido y usarlo en `verificar`.
- `inputs.hito` se interpola en un `shell`; está restringido por `enum`. No añadir inputs libres a pasos `shell`.
- Si `speckit init` se actualiza (`specify integration upgrade`), regenera `.claude/skills/speckit-*`; el workflow y la constitución no se tocan.
