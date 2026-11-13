# T007: la verificación que se cortó por tiempo

## Intento 1: `internal/evals` no cupo en los 10 minutos de `go test` por omisión

**Qué falló.** `make test` salió con `panic: test timed out after 10m0s` en `internal/evals` (601,6 s), con los demás
paquetes en `ok`. En `gates/ci.log` no hay ninguna línea `--- FAIL`: no falló ninguna aserción, se acabó el plazo con
tests todavía corriendo (`TestArnesDeLasFamilias`, `TestArnesDeNormativa`, `TestHuellaDeLasEvals`) y otros esperando
turno en `t.Parallel`.

**Causa.** El paquete tarda, con `-race`, casi lo que da el plazo por omisión, y cuánto tarda depende de la carga de la
máquina y del orden de `-shuffle`. Medido en esta sesión, el paquete solo, con `-race -count=1 -shuffle=on`:

| Qué | Tiempo |
|---|---|
| Sin `TestArnesConUnaEval` (`-skip`) | 288,6 s |
| Sin `TestArnesConUnaEval` (`-skip`), otra vez | 7 min 8 s |
| Con todo | cortado a los 9 min (`-timeout 9m`) |
| Con todo, otra vez | 456,0 s |
| Solo `TestArnesConUnaEval` | 14,8 s |

El test nuevo de la tarea cuesta 15 s a solas y no explica la diferencia: las dos medidas sin él se separan 140 s entre
sí. Lo que pesa son los tests del arnés del laboratorio, que entre todos abren unas 180 sesiones (contadas en sus
casos: 29 en `TestArnesDelLaboratorio`, 32 en `TestArnesDeNormativa`, 84 en `TestArnesDeLasFamilias`, 24 en
`TestArnesConLaInstalacionDelProducto` y 10 en `TestArnesConUnaEval`), cada una con su preparación de verdad: en el
perfil de CPU de dos casos de `TestArnesConUnaEval`, `prepararLaCache` es el 43 % de las muestras. No hay ningún
bloqueo: todos los tests terminan.

**Arreglo.** El Makefile da a `test` y a `test-integration` un plazo explícito, `PLAZO_DE_LOS_TESTS := 20m`, con su
porqué en un comentario. El control no se quita: un test colgado se sigue cortando, con su volcado, a los 20 minutos.
No se toca ningún test ni ningún valor esperado. El Makefile es una ruta de la tarea.

**Lo descartado.** Quitar o abaratar sesiones de los tests del arnés: las de `TestArnesConUnaEval` las fija la línea de
la tarea, y las demás son de H39, H40, H41 y T006, que la tarea manda dejar como están. Abaratar la preparación de cada
sesión (internal/evals/preparar.go, internal/evals/sesiones.go): es código de producción que la tarea no declara y que
el hito no pide.

**Lo que queda sin resolver.** El plazo da margen, no abarata el paquete: `make ci` lo corre dos veces (`test` y
`test-integration`), y el paso del workflow que ejecuta `make ci` tiene 1800 s (docs/WORKFLOW.md). Si el arnés sigue
ganando sesiones, lo que hay que decidir es su coste, y eso es de una persona y de otro hito.

**Verificación.** `make ci` no cabe en una sola orden en los 10 minutos que la sesión da a una orden en primer plano,
así que se corrieron sus doce objetivos por partes, todos en verde, con el Makefile ya cambiado:

- `make fmt-check lint`: `0 issues.`
- La receta de `test` (`go test -race -shuffle=on -timeout 20m -coverprofile=… -skip …`), en dos órdenes:
  `./internal/evals/`, `ok` en 330,8 s; y los otros 31 paquetes, 28 `ok` y los tres sin tests.
- La receta de `test-integration` (`go test -race -tags=integration -timeout 20m …`), igual: `./internal/evals/`, `ok`
  en 305,8 s; y los otros, todos `ok`.
- `make test-tiempos vuln schema-check skills-check goreleaser-check secrets mod-verify mod-tidy-check`: en verde
  («No vulnerabilities found», «no leaks found», «all modules verified»).

Sin medir: `make ci` en una sola orden, y `make test` con todos los paquetes a la vez tras el cambio, que es donde se
cortó. Los perfiles de cobertura de las órdenes partidas fueron al directorio temporal, no a `coverage.out`.
