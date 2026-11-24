# T008: por qué el intento 1 no quedó en verde

## Lo hecho (queda en el árbol)

- `internal/evals/definicion_test.go` (el test primero, visto en rojo por aserción con `evals.yml` a dos):
  `sesionesALaVezComoMucho = 12` con su comentario; `ajustesDelTrabajo["plazos-y-territorio"]` con `Concurrencia: 3`;
  la definición sintética del contrato a tres, con los casos sintéticos al día (`peorCasoDeLegal` de 3045 s, el de 59
  minutos ya sin la línea de `plazos-y-territorio`, «admite 13 sesiones … como mucho 12» y el caso nuevo
  `plazos-y-territorio-de-dos-en-dos`); `peorCasoEsperado`, que deriva el peor caso de `LeerConjunto` sin el plan;
  `probarElPeorCasoDelTrabajo` sin `[]int{67, 66, 6}`, `20725 * time.Second` ni los textos de los cuatro trabajos
  (fuera las constantes `peorCasoDe…` que solo usaban ellos y el tope); el subtest `cuatro-evals-mas`
  (`probarCuatroEvalsMas`, con `os.CopyFS` y `os.Root`); y `tope-con-las-evals-del-repositorio`, con un caso por skill
  (el mayor tope que no cubre su peor caso), `por-debajo-de-todos` y las líneas de las mutaciones de H25 y H31
  calculadas con el peor caso de la definición que montan.
- `internal/evals/plan_test.go`: `probarElPlanDelRepositorio` con lo esperado de `seriesPorModoDe`, `modosDeLaEval` y
  `gruposDe`, que usa también `peorCasoEsperado`.
- `.github/workflows/evals.yml`: `concurrencia: 3` para `plazos-y-territorio`, el comentario del `include` con la cuenta
  a tres y las doce sesiones entre los cuatro, y una frase en el del tope; el tope, los modelos y los demás ajustes, sin
  tocar.
- `CHANGELOG.md`, *Unreleased* «Cambiado»: la concurrencia nueva del trabajo de evals de `plazos-y-territorio`.

Visto en esta sesión: `go test -count=1 -run '^(TestDefinicionDelJob|TestPlanEnDosModos)$' ./internal/evals/` en verde
(165 tests) y `golangci-lint run ./internal/evals/` con 0 issues. En rojo, antes del cambio de `evals.yml`:
`cuatro-evals-mas` (27 253 s a dos, la cuenta de research D9), `del-repositorio` (la línea del `include`) y la premisa
de `tope-con-las-evals-del-repositorio`. En la copia con un paquete más (el programa de `paquete-de-mas.txtar` corrido
en `$TMPDIR`, con `99-01-territorio-municipio-cubierto.yaml`), los dos tests en verde; y en esa copia, en rojo por
aserción: `normativa` a seis («admite 13 sesiones a la vez entre sus 4 combinaciones, y lo esperado es como mucho
12»); `plazos-y-territorio` a dos (`del-repositorio`, `peor-caso`, `cuatro-evals-mas` y la premisa del tope);
`peorCaso` sin la prueba de red (`peor-caso` y `cuatro-evals-mas`); y el plan sin la sesión de red o sin el modo
herramienta (`TestPlanEnDosModos/del-repositorio`). Cada mutante, deshecho en la copia; ninguno en el árbol.

## Por qué cae

`make ci`, visto en esta sesión, cae solo en `internal/evals`, en un test que la tarea no declara:

```text
--- FAIL: TestComprobarElSondeo (0.01s)
    --- FAIL: TestComprobarElSondeo/concurrencia-del-job-en-plazos-y-territorio (0.07s)
            	Error Trace:	/home/jorge/code/kitlegal/internal/evals/sondeo_test.go:1233
            	            				/home/jorge/code/kitlegal/internal/evals/sondeo_test.go:1174
            	            	@@ -139,3 +139,3 @@
            	            	  repeticiones: (int) 3,
            	            	- concurrencia: (int) 2,
            	            	+ concurrencia: (int) 3,
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	92.296s
make: *** [Makefile:97: test] Error 1
```

El caso (`sondeo_test.go:1251-1257`) comprueba que el sondeo toma la concurrencia del trabajo de evals de su skill
leyendo la definición del repositorio, y la espera escrita: `concurrencia: 2`. Con `plazos-y-territorio` a tres
(FR-021) es 3. Ni contracts/evals.md §2 ni research D9 lo contaban, y `internal/evals/sondeo_test.go` no está entre las
rutas de la tarea (no se declara `sondeo.go`), así que este intento no lo toca. Los demás paquetes, en verde.

## La tarea, mal delimitada: lo que se amplía

La línea de T008 en `tasks.md` declara ahora también `internal/evals/sondeo_test.go`: los casos
`concurrencia-del-job-en-normativa` y `concurrencia-del-job-en-plazos-y-territorio` de `TestComprobarElSondeo` esperan
`ajustesDelTrabajo[<skill>].Concurrencia` (de `definicion_test.go`, el mismo paquete) en lugar de `5` y `2` escritos:
así lo esperado sale del sitio que fija la concurrencia del contrato, sin leer el YAML que el caso comprueba. Lo demás
de la tarea está hecho: al intento siguiente le queda ese cambio, ver en rojo el caso con el `2` de hoy (ya visto
aquí) y `make ci` en verde. `TestComprobarElSondeo` no abre sesiones; el gancho de las sesiones de un paso solo
rechaza el filtro `TestSondeo`.

## Intento 2: en verde

- Test primero: `go test -count=1 -run '^TestComprobarElSondeo$/^concurrencia-'`, con el `2` escrito, en rojo por
  aserción en `concurrencia-del-job-en-plazos-y-territorio` (8 en verde, 2 en rojo: el caso y su test).
- `internal/evals/sondeo_test.go`: los casos `concurrencia-del-job-en-normativa` y
  `concurrencia-del-job-en-plazos-y-territorio` esperan `ajustesDelTrabajo[<skill>].Concurrencia`; el comentario de las
  dos skills que sondean y el de `TestComprobarElSondeo` ya no llevan cifras escritas (decían «5 y 2» y «4 y 1»). El
  caso `concurrencia-pedida` sigue con su `2`, que es la que pide el caso. Los mismos casos, después: 10 en verde.
- En una copia nueva con un paquete más (el programa de `paquete-de-mas.txtar`, en el directorio temporal: «copia:
  hecha», «sincronía: 0»; sus dos listas salen con 1, y el principio de su salida nombra
  `TestCoberturaDelRepositorio`, `TestEvalsDelRepositorio/conjunto-plazos-y-territorio` y
  `TestEntregaDelHito/h32-paquetes`, tests con listas de las tareas de US4 que siguen abiertas, no de esta),
  `TestDefinicionDelJob`, `TestPlanEnDosModos` y `TestComprobarElSondeo` en verde. En esa copia, en rojo por aserción y
  deshecho después:
  - `plazos-y-territorio` a dos: `del-repositorio`, `peor-caso` (22 357 s con sus trece evals), `cuatro-evals-mas`
    (28 885 s), la premisa de `tope-con-las-evals-del-repositorio` y
    `TestComprobarElSondeo/concurrencia-del-job-en-plazos-y-territorio`;
  - `normativa` a seis: «admite 13 sesiones a la vez entre sus 4 combinaciones, y lo esperado es como mucho 12»;
  - el tope a 300 minutos, para ver la cifra del comentario de `evals.yml` que ningún test fija: «el peor caso de
    normativa, 20117 s».
- `make ci`, en primer plano: «ci: todos los controles en verde» (`internal/evals`, 99 s con `-race`; lint, 0 issues).
