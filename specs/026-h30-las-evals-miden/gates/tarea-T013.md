# T013 · Intento 1: redelimitada (falta declarar `internal/evals/cierre_test.go`)

## Causa

contracts/job-de-evals.md §2 fija el `run` del paso «Ejecutar las evals»:

```
DECISION_DE_EVALS="$RUNNER_TEMP/decision-de-evals.json" make evals SKILL="$SKILL_EVALUADA" 2>&1
```

`TestSalidasDeLosPasosQueImprimen` (`internal/evals/cierre_test.go`, de H25) fija que los dos pasos del flujo que
imprimen un fichero entre marcas —el informe del job y la medida del juez— llevan `2>&1`, y los encuentra recogiendo
las líneas de `evals.yml` que, sin sangría, empiezan por `run: ` y siguen con `make evals`. Con el `run` del contrato,
el de «Ejecutar las evals» empieza por `DECISION_DE_EVALS=` y el test ya no lo recoge, aunque lleve su `2>&1`. Salida de
`make fmt-check lint test` en este intento, en primer plano (la misma en `make test-integration`):

```
--- FAIL: TestSalidasDeLosPasosQueImprimen (0.00s)
    cierre_test.go:213:
        	Error:      	Not equal:
        	            	expected: "make evals SKILL=\"$SKILL_EVALUADA\" 2>&1\nmake evals-medir-juez SKILL=\"$SKILL_EVALUADA\" 2>&1"
        	            	actual  : "make evals-medir-juez SKILL=\"$SKILL_EVALUADA\" 2>&1"
        	Test:       	TestSalidasDeLosPasosQueImprimen
        	Messages:   	los pasos que ejecutan las evals y la medida del juez, en su orden, con las dos salidas por una tubería
FAIL	github.com/jmorenobl/kitlegal/internal/evals	118.863s
make: *** [test] Error 1
```

`cierre_test.go` no es el `_test.go` de ningún fichero declarado (no hay `cierre.go`), y plan.md («Tests existentes que
cambian lo que esperan») no lo nombra. La única salida dentro de las rutas declaradas sería apartarse del `run` del
contrato —pasar `DECISION_DE_EVALS` por el `env` del paso para que el `run` siga empezando por `make evals`—, y eso es
cambiar el contrato para esquivar la heurística de otro test: un atajo. La tarea estaba mal delimitada: su línea en
`tasks.md` declara ahora `internal/evals/cierre_test.go`, con lo que cambia en él.

## Lo que falta (intento 2)

Solo en `internal/evals/cierre_test.go`, `TestSalidasDeLosPasosQueImprimen`: reconocer la orden de un `run:` tras las
asignaciones `NOMBRE=valor` con que empiece (la de `DECISION_DE_EVALS`), y esperar los dos `run` enteros, en su orden:

```
DECISION_DE_EVALS="$RUNNER_TEMP/decision-de-evals.json" make evals SKILL="$SKILL_EVALUADA" 2>&1
make evals-medir-juez SKILL="$SKILL_EVALUADA" 2>&1
```

Sin perder lo que comprueba: que son esos dos pasos, en ese orden, y que los dos llevan `2>&1`. Después, `make ci` y
marcar `[X]`.

## Estado del árbol que hereda el intento 2

Implementado en este intento, todo dentro de las rutas declaradas:

- `.github/workflows/evals.yml`: la cabecera dice qué mide cada disparador y que la release exige la suite completa
  (SC-016); `cambios` con `timeout-minutes: 15`, `contents: read` y `pull-requests: read`, checkout de la cabeza con
  `fetch-depth: 0`, `setup-go` y el paso «Elegir las evals que la propuesta ha podido cambiar» (id `elegir`, con
  `GH_TOKEN`, `GH_REPO`, `PROPUESTA` y `BASE_DE_LA_PROPUESTA`) con `TestEleccionDeLaPropuesta`, y la salida `elige`; la
  `tanda` con la condición de §2 (`evals-completa` y `needs.cambios.outputs.elige == 'si'`), `MODO_DE_EVALS` en su `env`
  (§1), las salidas `medir` y `modo`, `-modo "$MODO_DE_EVALS"` en `TestTandaDelCommit` y la marca «Esta ejecución mide
  la suite completa» detrás de la de hoy; en `evals`, `contents`, `actions` y `pull-requests` de lectura,
  `MODO_DE_EVALS`, `PROPUESTA` y `BASE_DE_LA_PROPUESTA` en el `env`, `fetch-depth: 0`, el paso «Decidir qué mide y qué
  reutiliza» tras instalar Go con `GH_TOKEN` y `GH_REPO` solo en él, «Ejecutar las evals» con `DECISION_DE_EVALS` y,
  el último, «El informe del trabajo está escrito» con `if: ${{ !cancelled() }}`. Matriz, concurrencia (5, 4 y 2),
  modelos, repeticiones y tope (352 min), sin cambios; el comentario del tope da los peores casos con la decisión
  (19 877 s, 13 477 s y 20 425 s, el mayor el de `legal-core`).
- `internal/evals/definicion.go`: lee el trabajo `cambios` (`trabajoDelFlujo`), el `env` de la tanda, y los permisos,
  los pasos (`pasoDelFlujo`, con `uses`, `with` y `env`) y las demás claves de la matriz del trabajo evals;
  `plazoDeLaDecision` (el `-timeout` del go test del paso de la decisión, con espacio o con igual; 0 sin el paso; error
  sin `-timeout`, sin valor o con uno que no es una duración mayor que 0); el peor caso suma ese plazo (`Decision`) y lo
  presenta tras lo de fuera de las sesiones. Nada exportado nuevo.
- `scripts/evals.sh`: tras el argumento y antes de cualquier otra comprobación, sin `DECISION_DE_EVALS`, vacía o sin
  ser un fichero legible, «evals: falta DECISION_DE_EVALS» y salida 1; con ella, `-decision` con su ruta absoluta (supuesto
  en gates/supuestos.md).
- Tests: `definicion_test.go`, `TestDefinicionDelJob` con lo nuevo (trabajo `cambios`, modo y salidas de la tanda, sus
  dos marcas como penúltimo y último paso, permisos, `fetch-depth`, el paso de la decisión entre `setup-go` y la
  ejecución, `GH_TOKEN` y `GH_REPO` solo en él, `DECISION_DE_EVALS`, la marca final, una combinación por skill y 11
  sesiones a la vez como mucho), con 36 definiciones sintéticas nuevas, el subtest `plazo-de-la-decision` y los peores
  casos con los 900 s; `estado-de-la-tanda` comprueba que sus pasos sintéticos son los tres últimos de la tanda de la
  definición; `sustitutos_test.go`, `TestGuionDeLasEvalsSinLaDecision` (8 casos, con el `claude` sustituto delante en el
  `PATH`).

Verificación en este intento, en primer plano:

- `go test -run '^(TestDefinicionDelJob|TestJuezDeLaDefinicionDelJob|TestGuionDeLasEvalsSinLaDecision|TestTandaConLaSuiteCompleta)$' ./internal/evals/`:
  en verde. Antes de tocar `evals.yml` y el guion, los dos tests nuevos salían en rojo por aserción (la definición de
  hoy: «la definición del job no es la de …»; el guion: «evals: falta MODELO_DE_EVALS», «evals: necesita Linux con
  strace y claude»).
- `go vet ./internal/evals/` y `go vet -tags evals ./internal/evals/`: en verde.
- `make fmt-check lint`: `0 issues.`; `make test` y `make test-integration`: solo el rojo de arriba;
  `make test-tiempos vuln schema-check skills-check goreleaser-check secrets mod-verify mod-tidy-check`: en verde.
- Mutante en un árbol aparte (copia en el directorio temporal, `/tmp/claude-501/kitlegal-mutante-t013`, que la sesión
  no pudo borrar: la orden pedía aprobación): el guion sin la comprobación deja `TestGuionDeLasEvalsSinLaDecision` en
  rojo en 7 de sus 8 casos («DECISION_DE_EVALS: unbound variable», «evals: falta MODELO_DE_EVALS», «evals: necesita
  Linux con strace y claude»); el del fichero legible sigue en verde, como debe.
- Los mutantes de la definición son las definiciones sintéticas de `TestDefinicionDelJob`, en memoria: cada una da
  exactamente su línea.

T013 queda `[ ]`.

## Intento 2: en verde

Solo cambió `internal/evals/cierre_test.go`, la ruta que el intento 1 añadió a la línea de la tarea:

- `TestSalidasDeLosPasosQueImprimen` reconoce la orden de un `run:` tras las asignaciones `NOMBRE=valor` con que empieza
  (`ordenTrasLasAsignaciones`, con `esAsignacionDeVariable`: un nombre de variable del shell seguido de un igual; una
  palabra que no lo es empieza la orden) y espera los dos `run` enteros, en su orden: el de «Ejecutar las evals» con
  `DECISION_DE_EVALS="$RUNNER_TEMP/decision-de-evals.json"` delante y el de la medida del juez, los dos con `2>&1`.
- `TestOrdenTrasLasAsignaciones`, nuevo, fija qué quita la función (9 casos: sin asignaciones, una, dos, la asignación
  de make tras la orden, `2>&1`, un nombre que empieza por cifra, una sola palabra, solo una asignación y vacía).

Verificación en este intento, en primer plano: `go test -run '^(TestSalidasDeLosPasosQueImprimen|TestOrdenTrasLasAsignaciones)$' ./internal/evals/`
en verde (antes del cambio, el primero salía en rojo con la salida del intento 1); `make ci` entero, en cuatro órdenes
(`make fmt-check lint`; `make test`; `make test-integration test-tiempos`; `make vuln schema-check skills-check
goreleaser-check secrets mod-verify mod-tidy-check`), todo en verde (`0 issues.`; `internal/evals` 134,8 s, 96,3 %). El
mutante del guion y los de la definición no se repitieron: el árbol que los probó en el intento 1 no cambió.

T013 queda `[X]`.
