# Contrato: el job de evals, en paralelo y con las mismas garantías

FR-030 a FR-037, FR-044, FR-050, FR-052, FR-094; SC-001, SC-008. Decisiones en [research.md](../research.md) D7-D15 y
D19; tipos en [data-model.md](../data-model.md) §6-§7. Sustituye, del contrato job-de-evals de H5 (§3.2), que las
sesiones se abran una tras otra desde bash; todo lo que garantiza cada sesión se queda (FR-031).

## 1. `scripts/evals.sh`

- Variables obligatorias: las de hoy y `CONCURRENCIA_DE_EVALS`, con la forma `^[1-9][0-9]*$`; si falta o no la tiene,
  `evals: la concurrencia <valor> tiene que ser un entero mayor o igual que 1` y código 1, antes de la primera sesión,
  como las repeticiones (FR-030). `OBJETIVO_DE_DURACION_DE_EVALS`, opcional, `0` si no está.
- Comprobaciones previas: las de hoy (Linux con `strace` y `claude`, sin Python como root, el proxy cerrado, los
  ficheros de eval y lo grabado, la skill instalada y `kitlegal` en el `PATH`), sin exigir ya `timeout` (D9).
- En lugar del plan en `plan.tsv`, del bucle de sesiones y de `TestInformeDelJob`, una orden:

  ```bash
  go test -tags evals -count=1 -timeout 0 -run '^TestEjecucionDelJob$' ./internal/evals/ -args \
    -skill "$skill" -modelo-que-decide "$MODELO_DE_EVALS" -modelos-informativos "$MODELOS_INFORMATIVOS_DE_EVALS" \
    -repeticiones "$REPETICIONES_DE_EVALS" -umbral "$UMBRAL_DE_EVALS" -concurrencia "$CONCURRENCIA_DE_EVALS" \
    -prueba-de-red="${PRUEBA_DE_RED:-false}" -objetivo-de-duracion "${OBJETIVO_DE_DURACION_DE_EVALS:-0}" \
    -skills "$HOME/.claude/skills" -commit "$COMMIT_EVALUADO" -sin-python "$salida/sin-python.txt" \
    -sesiones "$salida/sesiones" -informe "$salida"
  ```

  y, como hoy, la comprobación de que `informe.md` e `informe.json` se escribieron, su impresión entre marcas, el
  resumen de la ejecución y la salida con el código de la orden.

## 2. `scripts/evals-sesion.sh` (D8)

Guion de bash, sin argumentos, que el repartidor ejecuta en `trabajo/` de la sesión. Lee `../pregunta.txt` y
`../modelo.txt` y ejecuta, con `exec`, la orden de hoy carácter a carácter:

```text
claude -p "<pregunta>" --model <modelo> --output-format stream-json --verbose --max-turns 30 --no-session-persistence --setting-sources user --settings '{"sandbox":{"enabled":false}}' --permission-mode bypassPermissions --disallowedTools WebFetch WebSearch
```

con `strace -ff -e trace=execve,connect,clone,clone3,fork,vfork -s 131072 -o ../traza/t --` delante si
`KITLEGAL_EVALS_TRAZA` vale `si`. No usa nada posterior a bash 3.2 (S5). Su salida estándar y su salida de error son las
de la sesión (`sesion.jsonl`, `sesion.err`), que abre el repartidor.

## 3. El repartidor (D7, D9, D14)

`ejecutarSesiones(interrupcion <-chan struct{}, SesionesAEjecutar) (EjecucionDeSesiones, error)`, con `interrupcion` el
`Done()` del contexto que las entradas cancelan con `SIGINT` y `SIGTERM`. No recibe un `context.Context`: prepara en
proceso con `PrepararSesion`, que llega a `app.Main`, y `contextcheck` marca toda función con contexto que llegue a él
(research V59 y D14 de H5); el contexto de `exec.CommandContext` es el del tope de cada sesión. Lo que abre una sesión
es `abrirSesion(interrupcion, SesionesAEjecutar, SesionPlanificada)` (T006; `gates/supuestos.md`):

1. recorre el plan en su orden y abre una sesión cuando hay menos de `Concurrencia` abiertas;
2. para cada una: crea su directorio con `trabajo/`, `cache/`, `traza/`, `tmp/` (0700) y `claude/skills/<entrada>`
   —un enlace simbólico absoluto a cada entrada de `Skills`—; la prepara con `PrepararSesion` (las evals de la skill,
   `UnionDeGrabaciones()`, la eval, el modelo, la prueba de red); ejecuta el guion en `trabajo/`, en su propio grupo de
   procesos, con el entorno de §4; y escribe `codigo-de-la-sesion`;
3. el tope: a los `Tope` (240 s) envía `TERM` al grupo y, pasados `MargenDelTope` (10 s), `KILL`; el código es 124 si
   bastó `TERM`, 137 si hizo falta `KILL`, y el del proceso si terminó antes;
4. al terminar cada sesión la lee con `LeerSesion`; si es de la clase (a) (data-model §3), no abre ninguna más: espera
   a las abiertas y devuelve las no abiertas en `SinAbrir` (FR-044). Tras (b) o (c), sigue; tras una sesión que
   `LeerSesion` no puede leer, también: el informe la juzga como ilegible, como hoy (T007; `gates/supuestos.md`);
5. `Duracion`: desde antes de preparar la primera hasta que termina la última;
6. un error de preparación o de E/S, o `interrupcion` cerrado (`SIGINT`, `SIGTERM`), cierra las abiertas con la
   secuencia del tope y devuelve el error —el de la interrupción es `errSesionInterrumpida`—: ninguna sesión se
   reintenta ni se duplica (FR-037). Con uno o varios errores, devuelve todos unidos con `errors.Join`, sin el de las
   sesiones que se cierran por ellos; con la interrupción, el de cada sesión interrumpida, que la nombra y que
   `errors.Is` reconoce como `errSesionInterrumpida`, o `errSesionInterrumpida` a secas si no había ninguna abierta
   (T007; `gates/supuestos.md`).

## 4. El entorno de cada sesión (D10, D16)

Sobre la base (`os.Environ()` en el job; la lista de [sondeo.md](./sondeo.md) §5 en el sondeo), sustituyendo lo que ya
hubiera con el mismo nombre:

| Variable | Valor |
|---|---|
| `KITLEGAL_CACHE_DIR` | `<sesión>/cache` |
| `HTTP_PROXY`, `HTTPS_PROXY`, `http_proxy`, `https_proxy` | `http://127.0.0.1:9` |
| `NO_PROXY`, `no_proxy` | `api.anthropic.com` |
| `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` | `1` |
| `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` | `0` |
| `CLAUDE_CONFIG_DIR` | `<sesión>/claude` |
| `TMPDIR`, `CLAUDE_CODE_TMPDIR` | `<sesión>/tmp` |
| `KITLEGAL_EVALS_TRAZA` | `si` en el job; ausente en el sondeo |

## 5. `TestEjecucionDelJob` (etiqueta `evals`)

Banderas: `-skill`, `-modelo-que-decide`, `-modelos-informativos`, `-repeticiones`, `-umbral`, `-concurrencia`,
`-prueba-de-red`, `-objetivo-de-duracion`, `-skills`, `-commit`, `-sin-python`, `-sesiones`, `-informe`; las mismas
comprobaciones de banderas vacías que hoy (`exigirBanderas`). Lee el conjunto (sin ficheros mal formados, como
`TestPlanDeSesiones`), compone el `PlanDeEvals` y lo comprueba, reparte las sesiones con `Traza: true`, la ruta absoluta
de `scripts/evals-sesion.sh` y `os.Environ()`, y escribe el informe con `EscribirInforme`, la duración y `SinAbrir`.
Falla con un error o con el veredicto `fallo`, nombrando sus motivos. Cancela su contexto con `SIGINT` y `SIGTERM`.
Se retiran `TestPlanDeSesiones`, `TestPrepararSesion` y `TestInformeDelJob` (research D19).

## 6. `.github/workflows/evals.yml`, trabajo `evals` (D11, D12, D13)

```yaml
  evals:
    name: evals (${{ matrix.skill }})
    needs: [cambios]
    if: …                                   # sin cambios
    concurrency:
      group: evals-${{ github.event.pull_request.head.sha || github.sha }}-${{ matrix.skill }}
      cancel-in-progress: false
    strategy:
      fail-fast: false
      matrix:
        skill: [boe-legislacion, legal-core]
        include:
          - skill: boe-legislacion
            concurrencia: 4
            objetivo_de_duracion: 900
          - skill: legal-core
            concurrencia: 1
            objetivo_de_duracion: 0
    runs-on: ubuntu-24.04
    timeout-minutes: 120
    env:
      …                                     # las de hoy, y además:
      CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}
      OBJETIVO_DE_DURACION_DE_EVALS: ${{ matrix.objetivo_de_duracion }}
```

Sin `concurrency` de nivel de flujo. El filtro del trabajo `cambios` gana `scripts/evals-sesion.sh`. Nada más cambia:
eventos, permisos, pasos y la retirada de Python se quedan.

## 7. `TestDefinicionDelJob` (`definicion_test.go`, en `make ci`; FR-094)

Lee `.github/workflows/evals.yml` con `leerDefinicionDelJob` y falla, nombrando la clave, si:

1. el trabajo `evals` no tiene `concurrency.group` igual a
   `evals-${{ github.event.pull_request.head.sha || github.sha }}-${{ matrix.skill }}`, o `cancel-in-progress` no está
   o no vale `false`, o el flujo tiene `concurrency` de nivel de flujo (FR-034);
2. `name` no es `evals (${{ matrix.skill }})` (el nombre por el que el cierre lee cada informe; FR-034), o
   `strategy.matrix.skill` no es `[boe-legislacion, legal-core]`, con las que se recorre el peor caso del punto 4
   (T008; `gates/supuestos.md`);
3. `include` no da `concurrencia` 4 a `boe-legislacion` y 1 a `legal-core`, u `objetivo_de_duracion` 900 y 0, o `env` no
   pasa `CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}` y `OBJETIVO_DE_DURACION_DE_EVALS: ${{
   matrix.objetivo_de_duracion }}` (FR-030, FR-051);
4. para alguna skill de la matriz, `timeout-minutes × 60` es menor que su peor caso (FR-035):
   `485 s + ⌈N / C⌉ × (22 s + 240 s + 10 s)`, con `N` las sesiones de `PlanDeEvals` sobre `evals/<skill>/`, los modelos y
   las repeticiones del `env` y la prueba de red, y `C` su concurrencia (research D13). Hoy: 7 013 s y 5 653 s frente a
   7 200 s.

El error nombra el valor encontrado y el esperado; con el peor caso, sus cuatro términos.

## 8. Tests del repartidor, con sustitutos (en `make ci`, también en macOS; FR-036, FR-094)

Los sustitutos de `claude` y `strace` son guiones POSIX que el test escribe en `t.TempDir()` y pone delante en el `PATH`
de la base (research D18). Nadie abre una sesión con modelo.

| Test | Qué comprueba | Requisito |
|---|---|---|
| `TestEjecutarSesionesEnParalelo` | con un plan de 8 sesiones sintéticas (evals de un directorio temporal) y `Traza: true`: con `Concurrencia` 4, el máximo de sesiones abiertas a la vez que cuentan los sustitutos es ≤ 4 y ≥ 2; con 1, es 1; directorio de trabajo, `CLAUDE_CONFIG_DIR`, `TMPDIR`, `CLAUDE_CODE_TMPDIR` y `KITLEGAL_CACHE_DIR` distintos en cada sesión y dentro de su directorio; `claude/skills/` con un enlace por skill; y `informe.json` e `informe.md` iguales byte a byte con 4 y con 1, escritos con la misma duración | FR-030, FR-032, FR-094; SC-008; US3-1 |
| `TestEjecutarSesionesTrasElLimiteDeUso` | la sesión k da el transcript (a): con `Concurrencia` 2 no se abre ninguna posterior a las ya abiertas, las abiertas terminan y `SinAbrir` son exactamente las que faltan del plan; con (b) en la sesión k, se abren todas | FR-044, FR-093; SC-007; US3-2, US3-3 |
| `TestTopeDeLaSesion` | con un tope de 1 s y un margen de 1 s: un sustituto que duerme 5 s deja 124; uno que ignora `TERM`, 137; uno que termina antes, su código | FR-031 |
| `TestEjecutarSesionesConElContextoCancelado` | con el contexto cuyo `Done()` es `interrupcion` cancelado mientras hay sesiones abiertas (lo que hace `SIGINT` al sondeo): vuelve con `errSesionInterrumpida`, no abre ninguna más y ningún sustituto sigue vivo; es el mismo camino que un error de preparación | FR-037, FR-064 |
| `TestEjecutarSesionesConUnError` | con `Concurrencia` 2, una sesión cuyo directorio no se puede crear: cierra las abiertas con la secuencia del tope, no abre ninguna posterior y vuelve con el error que nombra esa sesión | FR-037 |
| `TestEjecutarSesionesSinConcurrencia` | una `Concurrencia` menor que 1 (`0`, `-1`) es un error antes de abrir ninguna sesión | FR-030 |
| `TestInterrumpirLaSesion`, `TestSesionQueNoSePuedePreparar`, `TestAbrirUnaSesion` | una sola sesión: la interrupción antes y después de abrirla (`errSesionInterrumpida`, sin `codigo-de-la-sesion`); la preparación con faltas, que no la abre; y su directorio, su orden, su entorno de §4 y su traza | FR-031, FR-032, FR-037 |
| `TestDefinicionDelJob` | §7 | FR-030, FR-034, FR-035, FR-094; SC-008; US3-6 |

## 9. Uso, de fuera adentro

- **Quién**: el job, en cada propuesta de cambio que toca lo que las evals miden y en cada etiqueta `evals`; quien ajusta
  la concurrencia, con los reintentos publicados.
- **Consumo**: el mismo número de sesiones que hoy por ejecución (93 o 94 en `boe-legislacion`, 18 en `legal-core`), en
  unos 9 min de sesiones en lugar de unos 34. En el cierre, la segunda ejecución sobre el mismo commit espera a la
  primera: el pico es de 5 sesiones a la vez (4 + 1), dentro de lo medido con esta cuenta (dos ejecuciones simultáneas,
  hasta 4 sesiones a la vez, las 111 de cada una con `result success`).
- **Señales**: los reintentos, la duración y las sesiones sin medir se miden en cada ejecución; ninguna sobrevive a la
  siguiente.
