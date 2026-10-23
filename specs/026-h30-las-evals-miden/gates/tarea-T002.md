# T002 · Intento 1: redelimitada (falta declarar `internal/evals/informe.go`)

## Causa

La regla de la tasa necesita, en el umbral conjunto de la clase que decide, *N*: las respuestas que la ejecución
completa del plan da a juzgar en sus modos. El umbral conjunto lo compone `juicioDelJuez.umbralesConjuntos`
(`umbrales.go`), que solo tiene las respuestas medidas; las evals del plan no llegan a `umbrales.go`:
`umbralesDelInforme(e, respuestas, lecturas)` las recibe sin ellas, y quien tiene `conjunto.Evals` es `informe.go`
(`EscribirInforme` → `medirLasRespuestas` → `juezDelInforme.juzgar`). Las alternativas dentro de las rutas declaradas
son atajos: volver a leer la carpeta de evals con `LeerConjunto(e.Evals)` desde `umbralesDelInforme` (lectura doble y un
error que no tendría dónde ir) o colgar *N* de `lecturasMedidas` (que mide otra cosa). La tarea estaba mal delimitada:
`informe.go` no estaba entre sus rutas, y el guardián lo rechaza (salida de `scripts/workflow/guardian-diff.sh tarea`
en este intento, antes de redelimitar):

```
guardián: T002 tocó ficheros fuera de sus rutas declaradas: internal/evals/informe.go
```

La línea de T002 en `tasks.md` declara ahora `internal/evals/informe.go`, con lo que cambia en él.

## Estado del árbol que hereda el intento 2

La tarea está implementada y `make ci` está en verde en este intento (`ci: todos los controles en verde`, salida 0,
en primer plano); solo falta que el guardián la vea con la línea nueva y marcarla `[X]`:

- `internal/evals/tasa.go`: `limiteDeLaTasa` (*N*, *K*, *p*ref en diezmilésimas), `nuevoLimiteDeLaTasa` (*K* con la
  comparación del job, `compararUmbral`; *p*ref por bisección), `para(n)` = k(*n*), con enteros exactos (`math/big`), y
  `PlanDeEvals.respuestasQueSeJuzganEnSusModos` (*N*: las sesiones del plan que `seJuzga` da al juez, en los modos).
- `internal/evals/umbrales.go`: `umbralesConjuntos` compara con `compararConLaTasa` (cumple si medida ≤ k(*n*); publica
  el umbral de la clase si con él la proporción da el mismo límite y, si no, k(*n*)/*n*).
- `internal/evals/informe.go` (la ruta que faltaba): `juicioDelJuez.delPlan` (*N*) y `juezDelInforme.juzgar` recibe
  las evals y lo pone con `e.plan(evals).respuestasQueSeJuzganEnSusModos()`.
- Tests: `tasa_test.go` (`TestLimiteDeLaTasa`, con el mutante de `umbral: 0.017` en memoria); `umbrales_test.go`
  (`TestUmbralConjunto`: los casos de 104 respuestas esperan k(104) = 3 y publican 3/104, con un caso nuevo de 4 que
  falla; `TestUmbralesConJuezYConSentencias` con `delPlan: 4`; `TestUmbralConjuntoConMenosRespuestas`, el ejemplo del
  contrato §4: 60 de 120 en `boe-legislacion`, k(60) = 2, umbral 1/30); `informe_test.go`, el comentario del conjunto
  de `jurisprudencia` con 59 respuestas (k(59) = 1, sin cambio de valores). `TestInformeDeLaSuiteCompletaNoCambia` sigue
  en verde sin tocar su golden.

Cifras de `TestLimiteDeLaTasa` (`t.Log`, ejecutado en este intento), para T016: `boe-legislacion`, umbral 0.028:
*p*ref = 0,85 %, *N* = 120, k(*N*) = 3; `jurisprudencia`, umbral 0.017: *p*ref = 0,36 %, *N* = 60, k(*N*) = 1;
mutante `boe-legislacion` con 0.017: *p*ref = 0,47 %, k(120) = 2.

Mutantes comprobados en una copia desechable (borrada), cada uno en rojo: k(*n*) que devuelve *K* para todo *n*
(`TestLimiteDeLaTasa`, `TestUmbralConjuntoConMenosRespuestas`); el criterio 1/40 en lugar de 1/50
(`TestLimiteDeLaTasa`); *N* sin pasar desde `informe.go` (`TestUmbralConjunto`, `TestUmbralesDeJurisprudencia`,
`TestInformeDeLaSuiteCompletaNoCambia`); y el umbral publicado siempre el de la clase (`TestUmbralConjunto`,
`TestUmbralConjuntoConMenosRespuestas`).

## Intento 2: verde y marcada

Sin tocar código: el árbol heredado se revisó contra data-model §9, contracts/informe-del-job.md §4 y research D12
(*N* del plan con `Modo != ""`, porque las series sin binario ni servidor van sin modo en `PlanDeEvals.Series`; *K* con
`compararUmbral`; *p*ref por bisección sobre un criterio monótono en *p*; k(*n*) ≤ *n* y k(0) = 0 sin dividir por 0 al
publicar). Ejecutado en primer plano en este intento:

- `go test ./internal/evals/ -run 'TestLimiteDeLaTasa|TestUmbralConjunto|TestUmbralConjuntoConMenosRespuestas|…'`: en
  verde, con el `t.Log` de las cifras de arriba (0,85 % / 120 / 3; 0,36 % / 60 / 1; mutante 0,47 % / 120 / 2).
- Mutante en el árbol, deshecho byte a byte después (`para` devolviendo *K* para todo *n*): `TestLimiteDeLaTasa` (los
  tres subtests) y `TestUmbralConjuntoConMenosRespuestas` (los tres) en rojo; `TestUmbralConjunto` en verde, porque
  k(104) = *K* = 3.
- `make ci`: `ci: todos los controles en verde`, salida 0.
- `scripts/workflow/guardian-diff.sh tarea` no se pudo ejecutar en la sesión (la orden pide aprobación en su sandbox);
  lo pasa `verificar.sh` antes de `make ci`. Lo tocado, según `git status`: `internal/evals/tasa.go`, `umbrales.go` e
  `informe.go` (declarados), `tasa_test.go`, `umbrales_test.go` e `informe_test.go` (`x_test.go` de los declarados) y
  ficheros del directorio del feature; nada en `testdata/`, `schemas/` ni `evidencias/`, ningún `t.Skip`.

T002 queda `[X]`.
