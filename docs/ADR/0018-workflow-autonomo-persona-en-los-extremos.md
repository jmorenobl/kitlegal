# 0018 · Workflow autónomo: la persona en los extremos

- **Estado**: aceptada
- **Fecha**: 2026-09-25
- **Hito**: transversal (workflow `hito` 2.0.0 y constitución 2.0.0, antes de H7). Sustituye en parte al ADR 0007:
  las pausas humanas, la parada por motivos «no corregibles» y las tareas `[plataforma]`. Complementa al ADR 0017:
  dice cómo se obtienen las descargas públicas de las que sale un fichero congelado de `data/`.

## Contexto y problema

El workflow `hito` se diseñó para ir solo, pero la constitución reservaba a una persona decisiones a mitad del run
(criterio 4, «escalar en lugar de adivinar»; capa 3, pausas por fixtures, fuentes y anomalías) y dejaba que un modelo
detuviera el run (`escalar: true` en clarify, `corregible: false` en los jueces). El run de H6 (`2478560f`,
2026-09-20 a 2026-09-24) se detuvo cuatro veces y ninguna de ellas era una causa que exigiera a una persona:

| Parada | Causa real |
|---|---|
| `check_gate_plan` | Tres rondas agotadas con motivos que el propio juez daba por corregibles |
| `gate_humano_datos` | La relación del INE la tenía que descargar y escribir una persona, con red |
| `siguiente_tarea` | T015 y T026 agotaron intentos: una esperaba a una persona con red; otra, el CI remoto |
| `ci_final` | Los jueces marcaron `corregible: false` motivos que después se corrigieron a mano, fuera del run |

Además, la revisión final dio **13 rondas**: cada una encontraba otra afirmación de un artefacto o de la
documentación que el producto final ya no sostenía. La ejecución de cierre de las evals, una tarea `[plataforma]`
dentro del bucle, se tuvo que repetir en casi todas, porque cada corrección de la revisión la dejaba sin cubrir la
cabeza (en H5 ya había pasado: `specs/006-h5-skill-boe-legislacion/gates/revision-pendiente.md`). Una copia duplicada de la conversación estuvo trabajando en
paralelo sobre el mismo árbol que el run. Y la evidencia de las pausas `[datos]` (fichas del PAG, volcados del REL,
tablas del INE) vivía en una carpeta temporal del sistema, de donde desaparecieron ficheros entre sesiones.

La persona que lleva el proyecto propuso el 2026-09-24 un diseño («SDD autónomo») con tres ideas: la persona solo
aparece al principio (la entrada) y al final (valida el resultado); parar es una decisión del código y nunca de un
modelo; y ante un fallo el sistema cambia de estrategia en lugar de detenerse.

## Opciones consideradas

1. **Mantener la capa 3 y el criterio 4 y afinar el supervisor.** Rechazada: las paradas de H6 no son fallos del
   supervisor sino reglas que piden a una persona en medio. Afinarlas reduce su número, no su causa.
2. **El diseño completo sobre un orquestador nuevo** (ADK, LangGraph o uno propio): generación de tests por un agente
   distinto del implementador con mutantes por criterio y retrotraducción, DAG de tareas en paralelo con worktrees y
   cambio de ejecutor en la escalera. No se descarta, pero no es lo que causa las paradas: spec-kit cumple los siete
   requisitos de runtime del diseño (orden fijo, condicionales y bucles con tope, reparto, pasos deterministas, modelo
   por paso, estado persistente y pausa excepcional). Se aplaza hasta tener los datos de H7 con esta versión.
3. **La versión sencilla del diseño sobre spec-kit.** La persona escribe la entrada y lee el informe; todo lo demás
   es automático; solo el código para el run, por una lista cerrada de causas. Elegida.

## Decisión

1. **La entrada es la sección del hito** en `docs/ROADMAP.md`: la única instrucción humana del run. `specify` la
   convierte en spec; `clarify` pregunta en un proceso y responde en otro, con el criterio de la constitución, y cierra
   toda ambigüedad que toque alcance, frontera humana, privacidad, términos de uso, anomalías o una decisión cerrada con
   la **lectura conservadora**: lo dudoso no se implementa, va a «Fuera de alcance» y queda marcado para el informe.
2. **El único rechazo antes del final es el de entrada.** El juez del spec puede devolverlo con una lista cerrada de
   motivos —`contradiccion`, `sin_criterio_comprobable` y `objetivo_vacio`—, cada uno con el fragmento, por qué impide
   derivar un test y una pregunta cerrada. El workflow escribe `gates/informe-rechazo.md` y el run termina ahí: la
   persona corrige la sección del hito y lo relanza. Cualquier otro motivo de cualquier juez va a un corrector.
3. **Ningún modelo detiene el run.** Desaparecen `escalar` y `corregible: false`. Los prechecks mecánicos no paran: sus
   defectos entran en la ronda como motivos. Si las rondas de un juez se agotan (cuatro veredictos, tres correcciones),
   el run sigue con la última versión y los motivos pendientes van a `gates/<fase>-pendiente.md` y al informe.
4. **Solo el código detiene el run, por una causa mayor**, y cada una deja un dossier (`gates/dossier.md`) con la
   causa, la evidencia y qué hace falta para seguir: DAG bloqueado (tres tareas seguidas en cuarentena sin ninguna en
   verde entre medias), dependencia externa inaccesible (una fuente que no responde tras un reintento, o `git push`
   que falla), credencial ausente (`claude`, `gh` u `origin`, que `hito.sh` comprueba al arrancar) y presupuesto de
   tiempo agotado (`KITLEGAL_TIEMPO_MAXIMO`, 48 h de reloj por omisión). Un límite de uso no es una causa mayor: es una
   espera, con cambio de familia de modelo si hay repuesto.
5. **Cuarentena en lugar de parada.** Una tarea que no queda en verde tras tres intentos (el primero con
   `modelo_implementacion`, los siguientes con `modelo_escalada` y el diagnóstico del anterior) aparta su trabajo como
   parche en `gates/cuarentena/Tnnn.patch`, devuelve el árbol al último commit en verde y el bucle sigue con la
   siguiente. Lo mismo hace el guardián global con un corrector o un reparador que toque lo que no debe.
6. **La suite de aceptación se escribe primero y se congela.** La primera tarea, `[aceptacion]`, escribe desde el spec
   los guiones e2e de la entrega en el directorio del feature y, si toca una skill, sus evals. Cada guion tiene que
   fallar contra el código de ese momento por una aserción («rojo primero»); después su huella queda en
   `gates/aceptacion-congelada.json` y el guardián de diff rechaza cualquier cambio. Al acabar el bucle se activan en
   `internal/app/testdata/script/` y `make ci` tiene que pasarlos. Quien implementa no escribe el test que le juzga.
7. **Los datos externos los graba un paso sin modelo** (`grabar_datos`), después del commit de la tarea `[datos]` que
   deja el manifiesto `<paquete>/testdata/grabaciones.json` y su test de grabación (`//go:build grabacion`,
   `TestGrabar*`). Solo graba fuentes con fila en `docs/SOURCES.md` de `main` con «Revisado» fechado: los términos
   de uso, la licencia y el `robots.txt` los sigue revisando una persona, pero **antes** del run. Pide con
   `internal/httpx` (ritmo, `robots.txt`, User-Agent) y deja las respuestas que reproducen los tests en `testdata/` del
   paquete y el material de origen del que se derivan ficheros de `data/` en `evidencias/<hito>/`. Los ficheros de
   `data/` derivados los produce código del repositorio; nadie los escribe a mano ni de memoria. El ejecutor sigue sin
   red.
8. **La evidencia de datos vive en el repositorio**, en `evidencias/<hito>/`, con `manifiesto.json` (fichero, huella
   SHA-256, bytes, fuente, fecha). Las otras dos opciones pierden algo que el proyecto necesita: un repositorio aparte
   separa la evidencia del commit que la usa y hace que la derivación de `data/` no se pueda reproducir desde un solo
   `git clone`; adjuntarla a una release no sirve hasta H19 y ata la evidencia de un hito a un artefacto de otro. Los
   fixtures grabados ya se versionan (`internal/source/boe/testdata/`, 1,2 MB); la evidencia es el mismo tipo de
   material. Un fichero de más de 20 MiB no se versiona: queda su huella en el manifiesto y una copia local en
   `~/.local/share/kitlegal/evidencias/<hito>/`. Que la licencia de la fuente permita redistribuir el material la
   comprueba la persona al revisar su fila de `docs/SOURCES.md`; las cinco fuentes actuales lo permiten con atribución
   (CC BY 4.0 o Ley 37/2007).
9. **Barrido global antes de los jueces.** Tras el `make ci` del hito, un paso contrasta de una vez todos los
   artefactos del hito y la documentación del repositorio con el código y el binario, y corrige el texto (nunca el
   código). Si deja `make ci` en rojo, su cambio se aparta.
10. **El cierre en la plataforma va después de la revisión final**, sobre la cabeza que se fusionará, y lo hace el
    workflow, no una tarea: empuja la rama, abre la propuesta si no existe, vuelve a poner la etiqueta `evals` y espera
    a todas las comprobaciones de GitHub Actions. Si algo sale en rojo, un reparador arregla, el workflow commitea y se
    vuelve a medir (hasta tres mediciones). Desaparece la etiqueta `[plataforma]`.
11. **Una sola sesión por run.** `hito.sh` toma un candado por árbol de trabajo y exporta un testigo que heredan
    `specify`, los pasos y las sesiones `claude -p` del run. Un gancho `PreToolUse` de Claude Code rechaza en cualquier
    otra sesión las ediciones y las órdenes que cambian el árbol o el historial mientras el run vive; leer sigue
    permitido, y `paso.sh` se niega a correr.
12. **El informe final es lo que lee la persona**, y va como cuerpo de la propuesta de cambio: estado local y remoto,
    trazabilidad de cada FR y SC a sus tareas y guiones, supuestos, tareas en cuarentena, lo que la capa 3 reserva a la
    persona (fuentes, anomalías, fuentes nuevas, fixtures, grabaciones y evidencias) y los commits posteriores a la
    revisión. Con él decide: fusiona, o corrige la entrada y relanza. Nunca corrige código ni tests a mano: toda
    corrección entra por la sección del hito. Fusionar sigue siendo humano (ADR 0007, gancho `pre-push`).

La constitución pasa a la 2.0.0: el criterio 4 cambia de «escalar» a «cerrar de forma conservadora», la capa 3 se
reparte entre el antes (entrada y filas de `docs/SOURCES.md`) y el después (informe y fusión), y las reglas del modo
desatendido recogen la grabación sin modelo, la suite congelada y la evidencia. Detalle operativo en
`docs/WORKFLOW.md`; la lógica de los pasos shell, en `scripts/workflow/`.

## Consecuencias

- Un run sin causa mayor termina con la rama empujada, la propuesta abierta y el informe como cuerpo. Las paradas que
  quedan son las de la lista del punto 4, cada una con su dossier, y el rechazo de entrada.
- La persona ya no revisa fixtures, grabaciones ni cambios de `docs/SOURCES.md` a mitad del run: los revisa en el
  informe, antes de fusionar. El riesgo es que una grabación equivocada alimente varias tareas; lo acotan que la fuente
  ya estaba revisada, que el paso no tiene modelo y que nada llega a `main` sin la fusión humana.
- Un guion de aceptación mal escrito no se puede arreglar dentro del run: las tareas que lo necesitan acaban en
  cuarentena y el informe lo enseña. Se corrige en la entrada y se relanza. Es el precio de que quien implementa no
  pueda tocar el test que le juzga.
- Una decisión conservadora que no era la esperada aparece en el informe, en «Supuestos»; se corrige detallando la
  sección del hito, no poniendo pausas.
- La evidencia de H6 sigue en `~/.local/share/kitlegal/evidencias/h6/`; llevarla a `evidencias/h6/` es un cambio
  aparte, con la licencia de cada fichero comprobada.
- Quedan para después, con los datos de H7: mutantes por criterio y retrotraducción de los tests, DAG en paralelo y
  cambio de ejecutor. Se medirá H7 frente a H6: paradas a mitad del run (4), rondas de la revisión
  final (13), tiempo de la persona y calidad (lo que la persona encuentre en la revisión final).
- Operación: `scripts/hito.sh H<n>` ya no admite modo; `scripts/hito.sh --resume <run_id>` sigue sirviendo tras una
  causa mayor. `granularidad=hito` desaparece: toda tarea pasa por el guardián y la cuarentena.
