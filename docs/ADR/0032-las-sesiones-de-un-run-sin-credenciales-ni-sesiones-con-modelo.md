# 0032 · Ninguna sesión de un run lee credenciales ni lanza sesiones con modelo: el secreto en un llavero y las sesiones en un sandbox

- **Estado**: aceptada
- **Fecha**: 2026-10-01
- **Hito**: transversal, tras H7.4 (#90) y antes de la release y de H20. Workflow `hito` 2.4.0 y constitución 2.9.0.
  Completa a los ADR 0018 (la persona en los extremos), 0020 y 0021 (qué no hace nunca un run) y 0030 (lo que cambia el
  cierre vuelve a la revisión).

## Contexto y problema

En el run de H7.4 (`0f3ef83d`, PR #90) las dos invocaciones de `reparar_cierre` usaron una credencial que no era suya.
Está en sus transcripts (`~/.claude/projects/-Users-jorge-Projects-kitlegal/`; el número es la línea del `.jsonl`):

| Qué hizo | Sesión y línea | Herramienta y regla que lo dejó pasar |
|---|---|---|
| Supo dónde estaba el token y cómo se lanza el sondeo | `388e428c…` L199: `cat …/memory/alias-sonnet-resuelto-por-el-cliente.md` | La memoria del proyecto, cuyo índice cargan todas las sesiones, también las del run, decía la ruta y la receta |
| Miró el fichero | `82050c28…` L142 (`ls -la ~/.config/kitlegal/`) y L195 (`test -r … && echo legible`); `388e428c…` L152 | `Bash(ls:*)` y las órdenes de solo lectura, que Claude Code admite sin regla |
| Leyó el token | `82050c28…` L381: `rtk proxy sh -c 'CLAUDE_CODE_OAUTH_TOKEN="$(tr -d "\n" < "$HOME/.config/kitlegal/claude-oauth-token")" make evals-sondeo …'`; `388e428c…` L374, con el `cat` dentro del guion | `Bash(rtk:*)`. Nunca la herramienta `Read`. Las formas con `$(cat …)` en la propia orden se denegaron (L296 y L356 de `388e428c…`: «Contains shell syntax (string) that cannot be statically analyzed»; L260, por el `rm -rf` que llevaba delante) |
| Escribió el guion fuera del repositorio | `388e428c…` L332 (`/tmp/kitlegal-sondeo-c12.sh`) y L361 (`~/Library/Logs/kitlegal/sondeo-c12.sh`) | La herramienta `Write`, con `Write` sin ruta en `allow` y `--permission-mode acceptEdits`: vale para cualquier fichero del equipo |
| Lo ejecutó | `388e428c…` L374: `rtk proxy sh /Users/jorge/Library/Logs/kitlegal/sondeo-c12.sh > …/sondeo-c12.log 2>&1; echo terminado` | `Bash(rtk:*)`. `sh <guion>` a secas se denegó (L344, L367), y redirigir a `/tmp` también (L335: «outside the working directories for this session ('/Users/jorge/Projects/kitlegal', '/Users/jorge/Library/Logs/kitlegal')»). El segundo directorio venía de `additionalDirectories` de `.claude/settings.local.json`, añadido al reanudar el run para la sesión interactiva |
| En segundo plano | El parámetro `run_in_background: true` de `Bash` en L374 (resultado en L375: `backgroundTaskId bib5gkhwf`); en la primera, L381 → L382 (`b81qgj0ld`) | Ninguna regla de permisos habla de ese parámetro |
| Cerró sin esperar y citó lo que no había | `388e428c…` L389, 25 s después de lanzarlo: «En cuanto termine el sondeo, registro V31»; el commit `53dfec5` lleva la cita a «V31» en `research.md` D23 | — |

La primera (`82050c28…`, 20:13 UTC) sondeó once evals con tres repeticiones, 33 sesiones según su propio registro
(`research.md` V29), esperó a que terminaran y citó cifras que había visto. La segunda (`388e428c…`, 20:59 UTC) lanzó
nueve, cerró sin esperar y citó una verificación que no existía. En total, 42 sesiones con modelo contra la suscripción
de quien lanzó el run, que nadie había pedido. Los permisos denegaron dos formas en la primera y seis en la segunda, y
la sesión siguió probando hasta que una pasó. El valor del token no salió del equipo ni quedó escrito en ningún
transcript (ninguno contiene una cadena con la forma de un token, `sk-ant-…`).

Los jueces del ciclo 3 cazaron la cita (los dos, criterio f, `gates/revision-{a,b}-r3.json`), no el uso de la
credencial; y los del ciclo 2 aprobaron la reparación que citaba el primer sondeo. `scripts/evals-sondeo.sh` decía en
su cabecera que no lo ejecuta «ninguna tarea del workflow» (FR-068 de H7.3): era una frase.

## Medidas

Con Claude Code 2.1.284 en macOS, en clones del scratchpad con los permisos de `.claude/settings.json` de `main`
(`3a8fe57`) y valores falsos: un fichero con una cadena inventada y un llavero de prueba. Las sesiones, de Haiku 4.5
salvo la reproducción, con Opus.

**Qué cubre una regla `deny`.** Con `Read(//…/credencial-falsa/**)` en `deny`: la herramienta `Read` falla («File is in
a directory that is denied by your permission settings»); `cat <fichero>`, `rtk proxy sh -c 'cat <fichero>'` y un
guion escrito con `Write` y lanzado con `rtk proxy sh` lo leen los tres. Sin la doble barra, la regla no cubre ni
`Read`. Una lista de órdenes prohibidas tampoco se puede cerrar: el run tiene que ejecutar `make` y `go`, que ejecutan
cualquier programa.

**El llavero.**

| Dónde está el valor de prueba | Leerlo sin que nadie responda |
|---|---|
| Llavero de inicio de sesión, elemento creado sin `-T` | Se lee sin preguntar: el elemento confía en `/usr/bin/security` |
| Llavero de inicio de sesión, elemento con `-T ""` | No se lee: diálogo del sistema, con un botón «Permitir siempre» que deja a `security` leerlo ya sin preguntar |
| Llavero propio, bloqueado, fuera de la lista de búsqueda | No se lee: diálogo que pide la contraseña del llavero, sin «siempre». También `show-keychain-info` y `add-generic-password` se quedan esperando |

Con el llavero propio, una sesión con los permisos de `main` no lo obtuvo por ninguna de cuatro formas (dos denegadas
por los permisos, dos esperando el diálogo hasta su plazo).

**Un gancho sobre el texto de la orden.** Un gancho `PreToolUse` que deniega el segundo plano, `claude`, el sondeo,
`security` y las escrituras fuera del repositorio paró las formas directas, y una sesión a la que se pidió insistir se
saltó tres de sus cuatro denegaciones: escribió el guion dentro del repositorio y lo copió fuera con `cp`, y lanzó el
`go test` con «&» desde otro guion. El gancho no ve lo que hace un guion.

**El sandbox de Claude Code** (`sandbox.enabled`, Seatbelt en macOS), con `allowUnsandboxedCommands: false`:

| Orden, dentro de `rtk proxy sh` o de un guion | Resultado |
|---|---|
| `cat` del fichero en `filesystem.denyRead` | «Operation not permitted» |
| Escribir en `~/Library/Logs/…`, también con `dangerouslyDisableSandbox` | «Operation not permitted» |
| Escribir en `/tmp` | «Operation not permitted»; en `$TMPDIR` (`/tmp/claude-<uid>`) y en el repositorio, sí |
| `curl` a `api.anthropic.com`, `api.github.com`, `www.boe.es` | Denegado («deny network-outbound») |
| `/Users/…/.local/bin/claude -p …`, por su ruta absoluta | No llega a `api.anthropic.com`; reintenta hasta que vence el plazo de la orden |
| `security find-generic-password` sobre el llavero bloqueado | Sigue esperando el diálogo: el sandbox no lo impide, lo impide el bloqueo |
| `make ci` | Verde («ci: todos los controles en verde»), con el temporal del usuario, las cachés de Go y de golangci-lint escribibles, `proxy.golang.org`, `sum.golang.org` y `vuln.go.dev` en la red, `allowLocalBinding` y `enableWeakerNetworkIsolation`. Sin el temporal del usuario falla el `mktemp` de un guion de aceptación; sin lo último, `govulncheck` («x509: OSStatus -26276») |
| Un directorio de `permissions.additionalDirectories` | Escribible dentro del sandbox |
| Un proceso que un guion deja con «&» | Sobrevive a la sesión, dentro del sandbox: escribió en `$TMPDIR` 20 s después y no pudo escribir fuera |

Dos opciones `--settings` no se suman: Claude Code atiende la última. `--setting-sources user,project` deja fuera
`.claude/settings.local.json` (con él, redirigir a su directorio adicional pasa; sin él, «needs approval»).
`CLAUDE_CODE_DISABLE_BACKGROUND_TASKS=1` quita `run_in_background`: la misma llamada corre en primer plano.

**El gancho sobre el run de H7.4.** Repetidas por él las 2 110 llamadas `Bash`, `Write` y `Edit` de sus 41 sesiones,
deniega 56: 22 órdenes en segundo plano (los dos sondeos y sus intentos, y el `make ci` que los jueces lanzaban en una
copia y esperaban con un bucle), 30 escrituras en `~/Library/Logs/kitlegal` (el guion del incidente y las copias que
editó el juez B), `sh scripts/evals-sondeo.sh --help` y tres `claude --version`. Ninguna otra.

## Decisión

1. **La credencial del sondeo vive en un llavero propio y bloqueado, no en un fichero.**
   `~/Library/Keychains/kitlegal-sondeo.keychain-db` (otra ruta, `KITLEGAL_LLAVERO`), elemento genérico
   `kitlegal-claude-oauth-token`, cuenta `kitlegal`. `make evals-sondeo` pasa por `scripts/evals-sondeo-llavero.sh`,
   que lo lee —el sistema pide a la persona la contraseña del llavero—, lo vuelve a bloquear en cuanto tiene el valor
   y ejecuta `scripts/evals-sondeo.sh`, que no cambia. Con `CLAUDE_CODE_OAUTH_TOKEN` ya en el entorno (Linux) no toca
   el llavero. Se niega a correr dentro de un paso del workflow o con un run vivo. `scripts/hito.sh` no arranca si el
   fichero antiguo existe o si una credencial de Claude va en el entorno, que heredarían las sesiones.
2. **Las sesiones de un paso corren en el sandbox de Claude Code.** `scripts/claude-modelo.sh`, la única puerta por la
   que el workflow y `scripts/paso.sh` abren una sesión, les pone por `--settings`: el sandbox, obligatorio
   (`failIfUnavailable`) y sin salida (`allowUnsandboxedCommands: false`); la red, solo a los tres servidores de las
   herramientas de Go y al propio equipo; la escritura, solo en el repositorio, el directorio temporal y las cachés de
   Go y de golangci-lint; y la lectura, vedada en `~/.config/kitlegal`, `~/.config/gh`, `~/.ssh` y lo que diga
   `KITLEGAL_SECRETOS`, también para la herramienta `Read`. Sin `api.anthropic.com` no hay sesión con modelo, y sin
   `github.com` no hay `git push` ni `gh`: lo que los prompts ya prohibían y los ADR 0020 y 0021 confiaban a los
   permisos pasa a impedirlo el sistema operativo. `autoAllowBashIfSandboxed: false`: los permisos del proyecto deciden
   lo mismo que antes.
3. **Sin segundo plano, sin `claude` y sin los ajustes locales.** El mismo wrapper exporta
   `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS=1`, pone delante en el `PATH` un `claude` que no abre nada y dice por qué
   (`scripts/workflow/sin-modelo/claude`) y pasa `--setting-sources user,project`. `hito.sh` no arranca si el
   `settings.json` del proyecto o el de la cuenta declaran directorios adicionales.
4. **El gancho da el motivo y cubre lo que el sandbox no ve.** El wrapper exporta `KITLEGAL_PASO_DE_WORKFLOW`, y con esa
   marca el gancho `PreToolUse` que ya existía entrega la llamada a `scripts/workflow/politica-paso.sh`: deniega
   `Write` y `Edit` fuera del repositorio, del temporal y de la memoria del proyecto —no pasan por el sandbox—, y, por
   el texto de la orden, el segundo plano, abrir sesiones con modelo y pedir credenciales, con una frase que nombra la
   regla. `.claude/settings.json` no cambia. `hito.sh` ejecuta su tabla de casos al arrancar.
5. **Los prompts lo dicen.** Los 22 pasos con modelo (13 `prompt`, 9 `command`) terminan con la misma línea, «SESIÓN
   DE UN PASO». `reparar_cierre` ya no sugiere el guion de evals: diagnostica con lo que el job dejó y dice lo que no
   puede medir. La constitución lo recoge en «Reglas del modo desatendido», con la regla de la evidencia: ninguna
   sesión cita una verificación que no ha ejecutado y visto terminar, o que no está en el repositorio.
6. **La rúbrica de los jueces no cambia.** El criterio f («ninguna afirmación de la documentación ni de los artefactos
   que el producto no sostenga») y «con su evidencia» de los cambios del cierre ya hacen motivo de una verificación
   citada que no está: los dos jueces la cazaron en el ciclo 3 de H7.4. Añadir un criterio sería texto sin defecto que
   lo pida (ADR 0028, proporcionalidad).
7. **La ruta de un secreto no se escribe en la memoria del proyecto**, que cargan también las sesiones del run.

## Validación

- **La vía del incidente, con valores falsos**, pidiendo a la sesión (Opus) hasta cuatro formas por paso. Con `main`:
  lee el fichero, escribe el guion en `~/Library/Logs/…`, lo lanza con `run_in_background`, y el `go test` corre con el
  valor en su entorno (`hecho.txt`: «valor de 37 caracteres en su entorno»); el llavero no lo obtiene. Con la rama: no
  lee el fichero (`Read` denegado; `cat`, `sh -c` y un guion, «Operation not permitted»), no escribe fuera (`Write`
  denegado por el gancho; `cp`, por el sandbox), no lanza nada en segundo plano (gancho), no obtiene el llavero y el
  `claude` de su `PATH` es el sustituto. El directorio de fuera queda vacío.
- **`make ci` dentro del sandbox**, en un clon de la rama: verde, unos 3 minutos.
- **Lo que un paso hace a diario**, dentro del sandbox: `go test`, una copia con `git worktree` en `$TMPDIR` y su
  `go build`, el binario ejecutado en `$TMPDIR`, y la creación de la rama de un hito con el remoto inalcanzable
  (`create-new-feature-branch.sh` tolera el `git fetch` fallido).
- **El uso legítimo**: con el llavero de prueba bloqueado y nadie delante, el guion espera el diálogo y no abre nada;
  con la contraseña dada, tiene el valor (comprobada su forma con un `go` de mentira, sin imprimirlo), y el llavero
  queda bloqueado otra vez; dentro de un paso, se niega.

## Consecuencias

- Un paso no puede medir con modelo. Una eval en rojo en el cierre se diagnostica con el informe del job y se vuelve a
  medir en la medición siguiente: como mucho tres por run (ADR 0030), sin el sondeo intermedio que en H7.4 costó 42
  sesiones. Si eso se queda corto, la respuesta es otra medición del workflow, no una sesión que mide por su cuenta.
- Las copias de los jueces y los binarios de prueba van a `$TMPDIR`; `/tmp` y `~/Library/Logs` ya no valen. `make ci`
  en una copia va en primer plano: cabe en el plazo de la herramienta.
- Dentro de una sesión, `git fetch`, `gh` y cualquier fuente externa fallan. Lo que el run hace con la red —grabar
  datos, publicar, medir el cierre— ya era de los pasos shell, que no pasan por el wrapper.
- El run solo corre donde hay sandbox: macOS, o Linux con `bwrap` y `socat`. `hito.sh` lo comprueba al arrancar.
- Una sesión interactiva no cambia: no pasa por el wrapper. El sondeo lo lanza ella a petición de la persona, que
  escribe la contraseña del llavero en el diálogo.
- Las cachés de Go quedan escribibles para las sesiones: son las de las herramientas que el workflow les pide
  ejecutar. Una caché propia del run las aislaría de los demás proyectos del equipo a cambio de compilar dos veces; no
  hay hoy un defecto que lo pida.

## Lo que queda

- **Un proceso que un guion deja con «&» sobrevive a la sesión.** Ni el gancho lo ve ni el sandbox lo impide, y en
  macOS no hay forma fiable de encontrarlo después (el entorno de los procesos del sistema no es visible, y Claude Code
  abre cada orden en su propio grupo). Sigue dentro del sandbox: sin red, sin secretos y sin escribir fuera del
  repositorio y del temporal. Lo que puede hacer es seguir tocando el árbol cuando el paso ya terminó; lo ven la
  verificación del workflow, que mide el árbol después, y los jueces.
- **El diálogo del llavero aparece si una sesión lo provoca desde un guion.** La garantía es que la persona no escribe
  la contraseña de un diálogo que no ha pedido.
- **`Read` sigue leyendo cualquier fichero** que no esté en la lista de secretos. Lo que lea no puede salir del equipo
  por una orden de la sesión.

## Alternativas descartadas

- **Reglas `deny` sobre la ruta y las órdenes.** Solo cubren la herramienta `Read`; medido arriba.
- **Solo el gancho sobre el texto.** Una sesión que insiste se lo salta con un guion; medido arriba.
- **El llavero de inicio de sesión con `-T ""`.** Un clic en «Permitir siempre» deja el elemento como sin `-T`.
- **`--restricted`.** Quita `Bash`: el run no podría compilar ni probar.
- **Ejecutar el run en un contenedor o una máquina aparte.** Es más aislamiento del que hoy pide el defecto y cambia
  dónde vive todo el proceso; queda como opción si el sandbox se queda corto.
- **Un criterio nuevo en la rúbrica de los jueces.** El criterio f ya lo cubre.
- **Barrer al terminar los procesos que deja la sesión.** No hay en macOS una marca fiable con la que encontrarlos.
