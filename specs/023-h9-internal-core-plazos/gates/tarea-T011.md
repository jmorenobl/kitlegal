# T011 — intento 1: redelimitada

## Causa

La tarea estaba mal delimitada: con las nueve evals de `evals/legal-core/`, dos tests que cuentan las evals de
`legal-core` del repositorio dejan de pasar, y sus ficheros no eran de las rutas congeladas del intento
(`gates/tarea-actual.json`), ni `x_test.go` de un `x.go` declarado (`plan.go` e `informe.go` no lo son):

- `internal/evals/plan_test.go`, `TestPlanEnDosModos/del-repositorio/legal-core` (`probarElPlanDelRepositorio`): espera
  6, 6 y 2 series y 18, 18 y 6 sesiones por modo, las de las cuatro evals de antes; con nueve son 16, 16 y 2 y 48, 48
  y 6 (y 49 en el modo orden con la prueba de red), las de contracts/evals-legal-core.md §5.
- `internal/evals/informe_test.go`, `TestInformeEnDosModos/legal-core` (`exigirElInformeDeLaSkillSinUmbrales`): su lista
  `deModo` tiene las tres evals de modo de antes y espera 14 series; con las cinco nuevas son ocho y 34 series. Sus
  `umbrales` siguen vacíos en este intento —`umbrales.go` no cambia hasta T012—, así que solo cambia la lista.

`make ci` en el árbol, en primer plano, al terminar el intento (salida de `make -C <repo> ci`, exit 2):

```
0 issues.
--- FAIL: TestInformeEnDosModos (0.00s)
    --- FAIL: TestInformeEnDosModos/legal-core (0.05s)
--- FAIL: TestPlanEnDosModos (0.00s)
    --- FAIL: TestPlanEnDosModos/del-repositorio (0.01s)
        --- FAIL: TestPlanEnDosModos/del-repositorio/legal-core (0.01s)
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	110.386s
make: *** [test] Error 1
```

(`TestPlanEnDosModos`: `expected: map[evals.Modo]int{"":2, "herramienta":6, "orden":6}`, `actual: …{"":2,
"herramienta":16, "orden":16}`, y lo mismo con las sesiones; `TestInformeEnDosModos`: las series esperadas son las 14
de antes y el informe da 34.)

La línea de T011 en `tasks.md` gana las dos rutas, `internal/evals/plan_test.go` e `internal/evals/informe_test.go`, y
una frase que acota lo que cambia en cada una. Ningún fichero fuera de las rutas congeladas se ha tocado en el árbol.

## Lo que deja el intento 1 en el árbol (todo de las rutas declaradas)

Hecho entero, con el test primero (rojo de compilación: `unknown field Tipo in struct literal of type
ComandoEsperado`, y después los tres mensajes exactos del esquema que ganan `missing property 'plazo'`):

- `schemas/eval.yaml.json`: `plazo` (`fecha` con el patrón `AAAA-MM-DD` de `$defs/dia`, `avisos` obligatorio, sin
  repetir, con el enumerado `$defs/codigo-de-aviso-del-plazo`, `razonamiento` no vacío), `no_cubierto` (`norma`
  `BOE-A-…`), `not` de los dos juntos, los dos en las ramas que los prohíben (no activa y sin binario ni servidor),
  `no_cubierto` exime de `comandos` como `sentencias`, `plazo` en el `anyOf` del esperado verificable, y
  `comando-calcular` en el `oneOf` de los comandos.
- `internal/evals/formato.go`: `Eval.Plazo` (`PlazoEsperado`), `Eval.NoCubierto` (`NoCubiertoEsperado`),
  `ComandoEsperado.Tipo`, `verboCalcular`, `formaCalcular`, `esDePlazo` y `esDeCobertura`.
- `internal/evals/plazos.go`, nuevo: `ExtraerLineasDelPlazo`, `ExtraerNoCubierto`, `ExtraerAvisosDelPlazo`,
  `ComprobarCodigosDeAvisoDelPlazo`, `juzgarElPlazo`, `cumpleElPlazo`, `pideElPlazo` y `tipoDelCalculo`.
- `internal/evals/juzgar.go`: `plazo_encontrado`, `avisos_del_plazo_ausentes`, `no_cubierto_declarado` y
  `lineas_del_plazo` en `ResultadoDeEval`; `juzgarElPlazo` entre las redacciones y el territorio;
  `cumpleLoEsperado(eval)` con `cumpleElPlazo`; `satisface` y `textoDelComando` con `formaCalcular`.
- `internal/evals/consultas.go`: `calcular` no pide ninguna consulta.
- `internal/evals/conjunto.go`: «esperado verificable» admite `plazo` y `no_cubierto`; «cobertura por materia», detrás
  de él.
- Las cinco evals de `evals/legal-core/` (05 a 09), con la fecha y los avisos de los casos 1, 2 y 3 y la norma de
  Tráfico de `data/cobertura.yaml` (`BOE-A-2015-11722`).
- Tests: `TestFormatoDePlazos` y `TestFormaDelComando/calcular` (`formato_test.go`), los tres mensajes exactos de
  `TestLeerEval` y `TestFormatoDeSentencias`, `TestJuzgarPlazos` (`plazos_test.go`),
  `TestConsultasNecesarias/calcular-sin-consultas`, `TestConjuntoDeEvals/legal-core` con la eval de plazo y la de
  cobertura en el conjunto sintético (`sin-la-de-cobertura`, `sin-la-de-plazo`, `tamaño` y `orden-de-la-tabla`),
  `TestEvalsDelRepositorio/conjunto-legal-core` con las nueve evals bien formadas y sus consultas servidas sin red por
  `UnionDeGrabaciones`, y en `definicion_test.go` el literal `peorCasoDeLegalCore` (14 629 s), el caso de
  `probarElPeorCasoDelTrabajo` (49, 48 y 6) y la línea de `legal-core` en `por-debajo-del-de-jurisprudencia`, cuyo tope
  de 209 minutos (12 540 s) ya no cubre los 14 629 s; con sus comentarios.
- `gates/supuestos.md`: cinco líneas de T011.

## Lo que hace el intento 2

1. Aplicar este diff, el único que falta (verificado en una copia, abajo):

```diff
--- a/internal/evals/informe_test.go
+++ b/internal/evals/informe_test.go
@@ -2106,10 +2106,11 @@ func leerElInformeFinal(t *testing.T, informe string) informeFinalLeido {
 // duración que le da la definición del job, sin ninguna sesión, y exige que la
 // skill siga sin umbrales (FR-045 de H21; FR-037 de H24): no tiene juez ni
 // objetivo, y umbrales es [],
-// también con las duraciones de sus tandas; y que sus series sean las de sus tres
+// también con las duraciones de sus tandas; y que sus series sean las de sus
 // evals de modo en cada modo, con los dos modelos, y las de su eval sin binario
-// ni servidor, una sola vez, con las del modelo que decide decidiendo: 14, en el
-// orden del contrato.
+// ni servidor, una sola vez, con las del modelo que decide decidiendo, en el
+// orden del contrato: desde H9, las de sus ocho evals de modo, 34
+// (contracts/evals-legal-core.md §5 de H9; FR-050).
 func exigirElInformeDeLaSkillSinUmbrales(t *testing.T) {
 	t.Helper()
 
@@ -2161,7 +2162,9 @@ func exigirElInformeDeLaSkillSinUmbrales(t *testing.T) {
 
 	deModo := []string{
 		"01-territorio-municipio-cubierto.yaml", "02-territorio-municipio-no-cubierto.yaml",
-		"03-no-activa-receta-de-cocina.yaml",
+		"03-no-activa-receta-de-cocina.yaml", "05-plazo-reposicion-leganes.yaml",
+		"06-plazo-reposicion-tordesillas.yaml", "07-cobertura-liquidacion-tributaria.yaml",
+		"08-cobertura-multa-de-trafico.yaml", "09-plazo-contencioso-agosto.yaml",
 	}
 
 	var esperadas []serie
--- a/internal/evals/plan_test.go
+++ b/internal/evals/plan_test.go
@@ -285,8 +285,9 @@ func probarLaPruebaDeRedEnDosModos(t *testing.T) {
 // probarElPlanDelRepositorio fija, con las evals del repositorio y los modelos y
 // las repeticiones de la definición del job, las series y las sesiones de cada
 // tanda del plan de cada skill, las de la tabla de contracts/evals-en-dos-modos.md
-// §2.1: 96, 96 y 6 sesiones en boe-legislacion y 18, 18 y 6 en legal-core, y
-// una más en el modo orden con la prueba de red.
+// §2.1: 96, 96 y 6 sesiones en boe-legislacion y, desde H9, con las nueve evals
+// de legal-core, 48, 48 y 6 en legal-core (contracts/evals-legal-core.md §5 de
+// H9; FR-050), y una más en el modo orden con la prueba de red.
 func probarElPlanDelRepositorio(t *testing.T) {
 	t.Parallel()
 
@@ -305,8 +306,8 @@ func probarElPlanDelRepositorio(t *testing.T) {
 		},
 		{
 			skill:    "legal-core",
-			series:   map[Modo]int{ModoOrden: 6, ModoHerramienta: 6, "": 2},
-			sesiones: map[Modo]int{ModoOrden: 18, ModoHerramienta: 18, "": 6},
+			series:   map[Modo]int{ModoOrden: 16, ModoHerramienta: 16, "": 2},
+			sesiones: map[Modo]int{ModoOrden: 48, ModoHerramienta: 48, "": 6},
 		},
 	}
```

2. `make ci` en primer plano y, en verde, marcar T011 `[X]`.

## Verificación hecha en este intento, sobre una copia

Copia del árbol con `.git` en `$TMPDIR` (`rsync -a`), con el árbol del intento y los dos cambios del diff de arriba, y
nada más (`git diff --stat` de la copia = el del árbol más esos dos ficheros):

- `make -C <copia> ci`: exit 0 dos veces, la última sobre el árbol final del intento («ci: todos los controles en
  verde»; `ok github.com/jmorenobl/kitlegal/internal/evals` en `test`, `test-integration` y el subconjunto de los tests
  del repositorio; `0 issues.` en `lint`).
- Mutantes de la verificación de la tarea, cada uno aplicado en la copia, con el test que cae, y restaurado (`cmp` con
  el árbol):
  - La forma `⚠ PLAZO:` con la fecha en cualquier punto de la línea (`[^\n]*` en lugar del separador detrás de los dos
    puntos): `TestJuzgarPlazos/lineas-del-plazo/fecha-detras-de-la-cita` y
    `TestJuzgarPlazos/eval-de-plazo/sin-la-fecha-delante-de-la-cita`, que espera el motivo
    `plazo ausente: 2026-10-13`.
  - La eval 07 sin `no_cubierto`: `TestEvalsDelRepositorio/formato` y `/conjunto-legal-core` caen por la eval mal
    formada y por las ocho evals, pero la regla «cobertura por materia» no, porque la 08 sigue siendo de cobertura
    (anotado en `gates/supuestos.md`). Con la 07 y la 08 sin `no_cubierto`, `/conjunto-legal-core` da además el
    defecto de la regla: `cobertura por materia: el conjunto lleva al menos una eval de plazo (plazo) y una de
    cobertura (no_cubierto); de plazo, 05-plazo-reposicion-leganes.yaml, 06-plazo-reposicion-tordesillas.yaml,
    09-plazo-contencioso-agosto.yaml; de cobertura, ningún fichero`.
  - Una definición sintética con el tope del trabajo evals en 243 minutos (14 580 s), por debajo de los 14 629 s, que
    esperara solo el fallo de `boe-legislacion`: `TestDefinicionDelJob/tope-con-las-evals-del-repositorio` cae con la
    línea `jobs.evals.timeout-minutes: vale 243 (14580 s), y lo esperado es al menos el peor caso de legal-core, 14629 s
    = 485 s + (⌈49 / 2⌉ + ⌈48 / 2⌉ + ⌈6 / 2⌉) × (22 s + 240 s + 10 s)`.

# T011 — intento 2: en verde

Aplicado el diff de arriba a `internal/evals/plan_test.go` e `internal/evals/informe_test.go`, y nada más fuera de lo
que dejó el intento 1. Sobre el árbol, en primer plano:

- Los once tests de la tarea (`TestPlanEnDosModos`, `TestInformeEnDosModos`, `TestJuzgarPlazos`,
  `TestEvalsDelRepositorio`, `TestDefinicionDelJob`, `TestFormatoDePlazos`, `TestConjuntoDeEvals`,
  `TestConsultasNecesarias`, `TestFormaDelComando`, `TestLeerEval` y `TestFormatoDeSentencias`): `ok`.
- Los tres mutantes de la verificación, repetidos en el árbol y restaurados con el editor (el cambio inverso exacto;
  `gofmt -l` limpio, las evals 07 y 08 con sus 832 y 822 bytes, y `git diff --stat` de `definicion_test.go` el del
  intento 1, 28 y 20 líneas): los mismos tres fallos del intento 1, con los mismos motivos y mensajes.
- `make ci`: «ci: todos los controles en verde» (`0 issues.` en `lint`; `ok …/internal/evals` en `test`,
  `test-integration` y el subconjunto de los tests del repositorio; `govulncheck`, `goreleaser check`, `gitleaks`,
  `go mod verify` y `go mod tidy -diff` sin reparos).

T011 marcada `[X]`.
