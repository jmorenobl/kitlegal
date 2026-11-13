# T004: por qué no quedó en verde

## Intento 1: la tarea necesita una ruta que no declara (redelimitada)

**Causa.** `ResultadoDeEval` (internal/evals/juzgar.go) no es solo el resultado del informe del job: también es el del
juicio del sondeo de `make evals-sondeo`, que juzga con su propio `juzgarSesionDelSondeo` (internal/evals/sondeo.go:480).
`TestJuicioDelSondeo` (internal/evals/sondeo_test.go) escribe el informe del job y compara, sesión a sesión, cada
`ResultadoDeEval` del job pasado por `juicioSinLaTraza` (sondeo_test.go:577) con el del sondeo. Desde esta tarea el del
job lleva `Activaciones` (`[]` o la lista) y el del sondeo no (nil), así que las 19 subpruebas salen en rojo. Arreglarlo
pide tocar internal/evals/sondeo_test.go, que no está entre las rutas de la tarea (ni es el `_test.go` de un `.go`
declarado), y el guardián lo rechazaría.

**Lectura elegida** (gates/supuestos.md, `[alcance] T004`): el sondeo no da la lista. El spec, el plan, el data-model y los
contratos solo la piden en el informe del job y en la tabla del laboratorio; nada nombra `make evals-sondeo` ni
`juzgarSesionDelSondeo`. Así que internal/evals/sondeo.go no cambia, y el comparador del test quita la lista del juicio
del job como quita lo que sale de la traza.

**Redelimitación.** La línea de T004 en tasks.md declara ahora internal/evals/sondeo_test.go: `juicioSinLaTraza` deja
también `Activaciones` en nil, con su comentario, y nada más de `TestJuicioDelSondeo` cambia.

**Arreglo comprobado en este intento y revertido** (el árbol no lo lleva, para no salir de las rutas): con la línea

```go
	resultado.Activaciones = nil
```

al principio de `juicioSinLaTraza`, `go test ./internal/evals/ -run 'TestJuicioDelSondeo'` dio `31 passed`. Le falta
el comentario de la función (añadir que quita también la lista de activaciones, que el sondeo no lee: H41.2 solo la da
en el informe del job).

### Lo que ya está en el árbol (lo hereda el intento siguiente)

- internal/evals/juzgar.go: `ResultadoDeEval.Activaciones []Activacion` con `json:"activaciones,omitzero"`, detrás de
  `Activada`, con su comentario.
- internal/evals/activaciones.go: `activacionesDeLaSesion(directorio, reconocedor, modo)` (con `leerLosUsos` y
  `activacionesDeLosUsos`; el error de una entrada de `Skill` mal formada va detrás de `sesion.jsonl: `) y
  `textoDeLasActivaciones` (`ninguna` con la lista vacía).
- internal/evals/informe.go: `EscribirInforme` construye el reconocedor una vez, delante de `juzgarSesiones` (su error,
  `el informe no se puede escribir: …`); `juzgarSesiones` y `juzgarSesion` lo reciben; la lista se lee tras la traza de
  toda sesión que `LeerSesion` leyó, con el modo de la sesión, y su error es un motivo `sesión ilegible: …`; se fija en
  el resultado de la sesión legible (tras `juzgarEnModo`) y en el de la ilegible (junto a `Activada`); la sección de
  cada sesión lleva, detrás de «Modelo de la sesión:», `Skills activadas: <texto>` o `Skills activadas: sin leer` sin
  lista.
- Tests: `TestTextoDeLasActivaciones` (activaciones_test.go); en informe_test.go, `TestInformeConLasActivaciones` con la
  lista exacta de `informe.json` (compactada) y el párrafo de cada una de las diez sesiones; `comprobarSesionIlegible`
  (caso `sesion-ilegible` de `TestInforme`) sin la clave y con `Skills activadas: sin leer`; `resultadoCrudo` gana
  `Activaciones`; `TestInformeDeLaSuiteCompletaNoCambia` quita además `sinLasActivaciones` y `sinLasSkillsActivadas`,
  con sus doce casos en `casosDeLasActivacionesQueSeQuitan`, y compara con las dos constantes sin tocarlas.

### Verificación de este intento

- Rojo antes del código: sin `textoDeLasActivaciones`, el paquete no compilaba; con el campo y la función pero sin
  tocar informe.go, en rojo `TestInformeConLasActivaciones` (las diez sesiones sin la clave), `TestInforme/sesion-ilegible`
  (`expected: "Skills activadas: sin leer"`, `actual: "Pregunta:"`) y `TestInformeDeLaSuiteCompletaNoCambia`.
- Verde con el código: `TestTextoDeLasActivaciones`, `TestInformeConLasActivaciones`, `TestInforme`,
  `TestInformeDeLaSuiteCompletaNoCambia` y `TestActivacionesDeLosUsos`: `81 passed`.
- Mutantes, cada uno en rojo y revertido: «antes» y «después» cambiados (rojo `TestTextoDeLasActivaciones` y las ocho
  sesiones con alguna activación de `TestInformeConLasActivaciones`); sin `omitzero` (rojo `TestInforme/sesion-ilegible`,
  «el resultado de la sesión sin leer no lleva la clave activaciones»); el párrafo sin lista como `ninguna` (rojo
  `TestInforme/sesion-ilegible`); un `.` detrás de «Modelo de la sesión: …» (rojo
  `TestInformeDeLaSuiteCompletaNoCambia`, «la línea 119 es distinta»).
- `make ci`: `fmt-check` y `lint` en verde; `test` en rojo solo por `TestJuicioDelSondeo` (sus 19 subpruebas). Extracto:

```
--- FAIL: TestJuicioDelSondeo (0.00s)
    --- FAIL: TestJuicioDelSondeo/aprobado (0.00s)
    ...
            	Diff:
            	--- Expected
            	+++ Actual
            	@@ -9,4 +9,3 @@
            	  Activada: (bool) false,
            	- Activaciones: ([]evals.Activacion) {
            	- },
            	+ Activaciones: ([]evals.Activacion) <nil>,
            	  PorMateria: (bool) false,
            	Test:       	TestJuicioDelSondeo/aprobado
            	Messages:   	la sesión 11-no-activa-programacion es la del job sin lo que sale de la traza
FAIL	github.com/jmorenobl/kitlegal/internal/evals	255.189s
make: *** [test] Error 1
```

  Los objetivos de `make ci` detrás de `test` (`test-integration`, `test-tiempos`, `vuln`, …) no llegaron a correr.

**Para el intento siguiente**: aplicar el arreglo de arriba en `juicioSinLaTraza` con su comentario, volver a ver en
rojo los cuatro mutantes si se quiere confirmar sobre el árbol heredado, y correr `make ci` entero.

## Intento 2: en verde

**Lo que cambia sobre el árbol heredado.** Solo internal/evals/sondeo_test.go: `juicioSinLaTraza` deja `Activaciones` en
nil, con su comentario (el sondeo no lee la lista; `juzgarSesionDelSondeo` e internal/evals/sondeo.go no cambian). Lo
demás del intento 1 se revisó contra contracts/informe.md §1 a §3 y §5 y queda como estaba.

**Verificación de este intento**, todo ejecutado en esta sesión:

- Rojo antes del arreglo: `go test ./internal/evals/ -run 'TestJuicioDelSondeo'`, en rojo por la comparación de
  sondeo_test.go:523.
- Verde con el arreglo: `TestJuicioDelSondeo`, `TestTextoDeLasActivaciones`, `TestInformeConLasActivaciones`,
  `TestInforme`, `TestInformeDeLaSuiteCompletaNoCambia`, `TestActivacionesDeLosUsos`, `TestJuzgarSinLaLista` y
  `TestLaActivacionEnJSON`: `116 passed`.
- Mutantes, cada uno en rojo y revertido: «antes» y «después» cambiados (rojo tres casos de
  `TestTextoDeLasActivaciones` y las ocho sesiones con alguna activación de `TestInformeConLasActivaciones`); sin
  `omitzero` (rojo `TestInforme/sesion-ilegible`, «el resultado de la sesión sin leer no lleva la clave activaciones»);
  el párrafo sin lista como `ninguna` (rojo `TestInforme/sesion-ilegible`, `expected: "Skills activadas: sin leer"`,
  `actual: "Skills activadas: ninguna"`); un `.` detrás de «Modelo de la sesión: …» (rojo
  `TestInformeDeLaSuiteCompletaNoCambia`, «la línea 119 es distinta»).
- `make ci`: en una sola orden no cabe en los 10 minutos que la sesión da a una orden en primer plano (se cortó por
  tiempo, sin veredicto; solo `internal/evals` tarda 305 s en `make test`). Se corrieron sus doce objetivos, en su
  orden y cada uno en primer plano, todos en verde: `fmt-check` y `lint` (`0 issues.`), `test`, `test-integration`,
  `test-tiempos`, `vuln` (`No vulnerabilities found.`), `schema-check`, `skills-check`, `goreleaser-check`, `secrets`
  (`no leaks found`), `mod-verify` y `mod-tidy-check`. No se vio la línea final `ci: todos los controles en verde`,
  que solo escribe la orden entera.
