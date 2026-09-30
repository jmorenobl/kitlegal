# Quickstart: validar H7.4

Escenarios para comprobar la entrega en la rama del hito, con el hito implementado. Todos se ejecutan desde la raíz del
repositorio. Salvo el sondeo con modelo de §8 y §9, ninguno usa la red ni abre una sesión con modelo. **Efectos**: `go test` y `go build`
escriben solo en las cachés de Go de quien los ejecuta; `make ci` deja además `coverage.out` y
`coverage-integration.out` en la raíz, que `.gitignore` ignora; el sondeo (§8) escribe en su temporal de `TMPDIR` y lo
borra al salir. Ningún escenario cambia el árbol de trabajo versionado, el índice ni el historial de git: se comprueba
con `git status --porcelain` antes y después, que da lo mismo.

Requisitos: el toolchain de Go que fija `go.mod` y, para §8, Claude Code en el `PATH` y la credencial de la suscripción
en `CLAUDE_CODE_OAUTH_TOKEN` (`claude setup-token`).

## 1. Todo `make ci` (FR-091; SC-011)

```bash
make ci
```

Termina con `ci: todos los controles en verde`, con `skills-check` y `schema-check` sin drift.

## 2. La lista, su calibrado y las formas de la skill (FR-032, FR-033, FR-021, FR-093; SC-002 a SC-004)

```bash
go test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/
go test -count=1 -run '^TestEvalsDelRepositorio$/^(expresiones-calibradas|expresiones-en-los-bloques|expresiones-de-la-skill|prosa-de-la-skill)$' -v ./internal/evals/
```

Pasa; con `-v`, las cuatro subpruebas en `PASS`. El reparto del calibrado es el de
[contracts/lista-de-expresiones.md](./contracts/lista-de-expresiones.md) §4 (36, 11 y 9).

## 3. El juicio de las dos clases y de la eval 20 (FR-034, FR-052, FR-094; SC-007)

```bash
go test -count=1 -run '^(TestJuzgarLasClasesDeLaRespuesta|TestExtraerExpresionesProhibidas|TestJuzgar)' ./internal/evals/
```

Pasa: cada frase de la bitácora marcada en su clase, las formas fijas y la respuesta de FR-022 sin marcar, las mismas
palabras de las formas fijas fuera de ellas marcadas («la que se consultó antes», `se consultó antes`), la sesión de
`legal-core` que activa `boe-legislacion` sin pasar, y las dos líneas `⚠ REDACCIÓN MODIFICADA:` de la eval 20 juzgadas
por su bloque y sus fechas.

## 4. `boe-legislacion` v0.1.4 (FR-021, FR-024, FR-026, FR-095; SC-004, SC-005, SC-011)

```bash
make skills-check
wc -l < skills/boe-legislacion/SKILL.md
go test -count=1 -run '^(TestOrdenesParaPowerShell|TestEvalsDelRepositorio)$' ./internal/evals/
```

`skills-check` en verde; menos de 300 líneas; las dos órdenes de lectura y comprobación con su forma para PowerShell.

## 5. Los umbrales y los recuentos (FR-040 a FR-048, FR-097; SC-006)

```bash
go test -count=1 -run '^(TestUmbralesDelInforme|TestInformeMarkdownDeLosUmbrales|TestInformeConSesionesSinMedir)$' -v ./internal/evals/
```

Pasa; con `-v`, los casos de [contracts/informe-del-job.md](./contracts/informe-del-job.md) §6, entre ellos
`tres-de-54-y-seis-sin-terminar`.

## 6. La respuesta a la pregunta (FR-060, FR-098; SC-008)

```bash
go test -count=1 -run '^TestLeerSesion$/^respuesta-antes-de-una-tarea-en-segundo-plano$' -v ./internal/evals/
```

Pasa: la respuesta leída es la de la pregunta, no la réplica a la tarea en segundo plano.

## 7. La tanda única y la eval 20 en el job (FR-054, FR-070, FR-071, FR-100; SC-010, SC-012)

```bash
go test -count=1 -run '^TestDefinicionDelJob$' -v ./internal/evals/
go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/
```

`TestDefinicionDelJob` pasa con `del-repositorio`, `sinteticas`, `errores` (la de hoy), `segundo-disparo` y
`estado-de-la-tanda`;
`TestGrabacionesDerivadas`, con las dos derivadas de `lcsp-a1-30-y-da-3-redaccion-original`.

## 8. El sondeo: errores de uso sin traza y, con modelo, la activación (FR-080, FR-081; SC-009; S7)

Sin credencial (no abre sesiones ni usa la red; `make` sale con 2 porque el guion sale con 1):

```bash
env -u CLAUDE_CODE_OAUTH_TOKEN make evals-sondeo SKILL=boe-legislacion EVALS=04 MODELO=claude-sonnet-5-5 REPETICIONES=1; echo "código: $?"
```

La salida de error es solo `falta la credencial: CLAUDE_CODE_OAUTH_TOKEN, …` (y la línea de error de `make`), sin nada de
`go test` ni de testify; `código: 2`.

Con argumentos que no valen:

```bash
make evals-sondeo SKILL=boe-legislacion EVALS=04 MODELO= REPETICIONES=0; echo "código: $?"
```

En la salida de error, solo `MODELO: está vacío` y `REPETICIONES: «0» no es un entero mayor o igual que 1`, en ese
orden (y la línea de error de `make`), con la credencial o sin ella; la salida estándar, vacía; `código: 2`.

Con modelo (lo lanza una persona, al leer el informe final; consume la suscripción y no es un veredicto): la 04, la 19
y la 20 con el modelo que decide, tres repeticiones.

```bash
make evals-sondeo SKILL=boe-legislacion EVALS=04,19,20 MODELO=claude-sonnet-5-5 REPETICIONES=3
```

Se espera la 04 en 3 de 3 (la skill activada), la 20 con sus dos líneas y ninguna respuesta con expresiones.

## 9. El job de cierre (FR-102; SC-001)

Lo lanza el workflow sobre la propuesta de cambio; aquí solo se lee su informe, que el cierre deja en
`specs/014-h7-4-boe-legislacion-sin/gates/evals/boe-legislacion.json`:

```bash
jq -c '.veredicto, .red, [.umbrales[] | {nombre, medida, total, cumple, decide}]' specs/014-h7-4-boe-legislacion-sin/gates/evals/boe-legislacion.json
jq -c '[.tasas[] | select(.eval | test("^(04|20)-")) | {eval, modelo, pasan, sesiones, formas}]' specs/014-h7-4-boe-legislacion-sin/gates/evals/boe-legislacion.json
jq -c '.veredicto, .umbrales' specs/014-h7-4-boe-legislacion-sin/gates/evals/legal-core.json
```

Se espera `"aprobado"`, `[]` y los cuatro umbrales de «Controles de umbral» con `decide: true` y `cumple: true`
(`expresiones_prohibidas:claude-sonnet-5-5` como mucho 2 de 54; `sin_activar` y `redaccion_no_leida`, 0; la duración,
≤ 900); la 04 pasando con su tasa y la 20 con sus dos líneas en `formas`; `legal-core` aprobado y con `[]`. Y una sola
ejecución de `evals (boe-legislacion)` por commit medido: `gh pr checks` lista una `evals (boe-legislacion)` y, de la
otra ejecución, `evals (${{ matrix.skill }})` en `skipping`.
