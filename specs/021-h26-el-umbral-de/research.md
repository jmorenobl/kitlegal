# Research: H26 · El umbral de las clases del juez pasa a ser una tasa sobre los dos modos

**Fecha**: 2026-10-13 · **Cabeza**: `869eaf8` (rama `021-h26-el-umbral-de`, igual que `main` salvo el directorio del
feature) · **Modo**: desatendido.

Cuatro tablas. **V**, lo comprobado leyendo el repositorio o la herramienta en esta sesión, con dónde. **M**, lo
medido en esta sesión ejecutándolo. **S**, lo que no se ha podido comprobar, dicho como supuesto. **D**, las
decisiones, cada una con su alternativa rechazada.

**El prototipo.** Las medidas M2 a M11 salen de un prototipo fuera del repositorio: un worktree de `869eaf8` en el
directorio temporal (`git worktree add --detach "$TMPDIR/h26-proto" HEAD`), con el cambio de código de D1 a D5, las
dos `clases.yaml` con sus valores y un fichero de test de prueba que arma las ejecuciones sintéticas y escribe lo que
da el informe. M15 y M16 salen de un segundo worktree, sin ese código. Los dos se retiraron al terminar
(`git worktree remove --force`) y `git status` quedó como al empezar (M13). Nada de ellos está en el árbol: lo que el
plan fija son sus resultados, no su código.

## V · Comprobado en el repositorio o en la herramienta

| Id | Qué | Dónde |
|---|---|---|
| V1 | Go 1.27: `go 1.27.0`, `toolchain go1.27.2`; el equipo tiene `go1.27.2 darwin/arm64` | `go.mod`; `go version` |
| V2 | El cambio no pide ninguna dependencia: `umbrales.go` importa `fmt`, `strconv` y `strings`, y le bastan | `internal/evals/umbrales.go:3-7` |
| V3 | `umbralesDelInforme` compone, en este orden: los de las respuestas por modo (`respuestasMedidas.umbrales`: `sin_activar`, el de cada clase y `cita_sin_documento`), los dos de la medida, los de la duración de las sesiones y los de la del juez. Ningún nombre de skill | `umbrales.go:199-250` |
| V4 | El umbral de una clase en un modo sale de `umbralesDeLasClases`, con `Umbral: clase.Umbral` y `Decide: clase.Decide`; su medida es `len(grupo.marcadasEn(posicion))` y su total, `len(grupo.respuestas)` | `umbrales.go:259-285` |
| V5 | `motivosDeLosUmbrales` salta el umbral que no decide, el que se cumple y los de la duración; el motivo de una clase lo completa `marcadasDelUmbral`, que hoy casa el nombre `<clase>:<modelo>:<modo>` | `umbrales.go:443-543` |
| V6 | Los grupos de los modos excluyen el de la eval sin binario ni servidor (`gruposDeLosModos`); una respuesta sin juzgar no está en `marcadasEn` y sigue en `respuestas`; se juzgan solo las sesiones del modelo que decide, planificadas, legibles, medidas, terminadas y de evals que activan la skill | `informe.go:559-569`, `673-683`, `1585-1623` |
| V7 | `cumple` es `valor() <= Umbral` en `float64`, con `valor` 0 si `total` es 0; la medida y la condición se escriben con un decimal y coma (`medidaEscrita`, `condicionEscrita`, `porcentajeConComa`) | `umbrales.go:408-426`, `567-591` |
| V8 | El sondeo junta las respuestas de todos los grupos en uno y cuenta `marcadasEn` por clase; no llama a `umbralesDelInforme` ni escribe umbrales | `sondeo.go:395-419`, `137` |
| V9 | El esquema de `clases.yaml` admite en `umbral` un número de 0 a 1; `LeerConjunto` valida con él la carpeta del juez | `schemas/juez-clases.yaml.json`; `internal/evals/juez.go:45` |
| V10 | `evidencias/adr-0037/` y `evidencias/adr-0037-jurisprudencia/` no tienen `clases.yaml`: se copian la rúbrica, el esquema, los casos y la medida | `ls` de las dos carpetas; `medida_test.go:650` |
| V11 | El informe final busca cada control `evals:<skill>:<nombre>` por nombre exacto (`select(.nombre == $n)`, con la skill en el campo 2 y el nombre del 3 en adelante), y solo lo da por «ok» si `decide` y `cumple`; uno con `decide: false` es «se publica sin hacer fallar el job». De `ci:<ruta>:<Test>` exige que la ruta exista en la cabeza y que tenga `func <Test>(` o el objetivo `<Test>:` | `scripts/workflow/informe.sh:250-273`; `comun.sh:148-166` |
| V12 | El guardián de diff protege `testdata/` y `schemas/` (solo con `[datos]`) y `evidencias/` (nunca); `evals/` no es una ruta protegida. Una tarea `[datos]` que no cambia ningún manifiesto pasa su comprobación (`grabar-datos.sh comprobar` sale con 0) | `scripts/workflow/guardian-diff.sh:70-106`; `grabar-datos.sh:80-93` |
| V13 | Nada bajo `internal/evals/testdata/` nombra `afirma_lo_no_leido` ni un umbral: ningún fixture cambia | `grep -rl afirma_lo_no_leido` |
| V14 | El binario no enlaza `internal/evals`: fuera del paquete solo lo importa `internal/skills/normas_test.go` | `grep -rln 'kitlegal/internal/evals"'` |
| V15 | Los informes del último cierre versionado llevan 12 elementos en `umbrales`, 10 con `decide: true`, en `boe-legislacion` y en `jurisprudencia` (totales de 54 y de 30 por modo), y `[]` en `legal-core`. Suman 3 482 y 3 757 bytes (3 495 y 3 770 con sus comas y corchetes) | `jq` sobre `specs/020-h25-el-juez-de/gates/evals/*.json` |
| V16 | El motivo real de una clase del juez, en ese cierre: 953 bytes con 2 respuestas; su cabecera mide 101 bytes y cada respuesta, con su sesión y sus tres frases, 425 | `jq '.motivos[]'` del mismo informe; `wc -c` |
| V17 | El job decide con `claude-sonnet-5-5`; `boe-legislacion` tiene objetivo de 900 s y `jurisprudencia`, ninguno | `.github/workflows/evals.yml:186-189`, `216` |
| V18 | `make ci` ejecuta `test` (`go test -race -shuffle=on`), `schema-check` (`TestEsquemasPublicados`) y `skills-check`, que lleva `TestEvalsDelRepositorio` | `Makefile:86-87`, `122-127`, `264` |
| V19 | Los tests ya tienen con qué armar lo que pide el hito: `ejecucionConUmbrales`, cuya ejecución «como el cierre» lleva 18 evals que activan la skill (54 respuestas del modelo que decide por modo) y que ya recibe el contenido de un fichero de la carpeta del juez (`medidaDelJuez`); `cortarTrasLaRespuesta`; el votante de votos grabados; y `armarSesionesDeJurisprudencia`, con las evals y la carpeta del juez del repositorio (30 por modo) | `umbrales_test.go:61-128`, `242-245`, `1364`; `informe_test.go:3657-3766`, `5162` |
| V20 | La declaración de clases de los tests del informe es una constante con `umbral: 0` | `informe_test.go:3578-3580` |
| V21 | Dónde dicen hoy los documentos el umbral de antes: `CONTRIBUTING.md` 724, 727-728, 1105-1108, 1134-1137, 1142, 1149-1151, 1160, 1167 y 1174-1177; `docs/JURISPRUDENCIA.md` 240-243; `docs/WORKFLOW.md` 125; `CLAUDE.md` 21 y 74; y en *Unreleased* de `CHANGELOG.md`, bajo «Cambiado», 129, 140, 146, 214-215, 224-225 y 231-233. `CONTRIBUTING.md` 1262-1280 (el sondeo) no cambia | lectura de cada fichero |
| V22 | El comentario de `.github/workflows/evals.yml` que dice «doce umbrales, diez que deciden» | sus líneas 152-154 |
| V23 | `go -C <dir>` cambia de directorio antes de la orden; `-run` recibe una expresión que se parte por cada `/` fuera de corchetes, un trozo por nivel de subtest; `git worktree add` admite `--detach` | `go help build` («-C dir»); `go help testflag` («-run regexp»); `git worktree add -h`; usados en M1 a M13 |

## M · Medido en esta sesión

| Id | Qué se ejecutó | Resultado |
|---|---|---|
| M1 | `go test -count=1 ./internal/evals/` sobre `869eaf8`, sin `-race` | `ok`, 29,4 s |
| M2 | Lo mismo con el cambio de código solo, y `umbral: 0` en las dos declaraciones | 7 funciones en rojo, con 51 subtests, todas porque esperan la lista o el motivo de antes: `TestUmbralesDelInforme`, `TestUmbralesConJuezYConSentencias`, `TestInformeMarkdownDeLosUmbrales`, `TestInformeEnDosModos`, `TestInformeConSesionesSinMedir`, `TestInformeConElJuez` y `TestUmbralesDeJurisprudencia`. En verde, sin tocarlas, las del sondeo, `TestCopiasDelJuez`, `TestMedidaVersionada`, `TestEjecucionSinMedir` y `TestDefinicionDelJob` |
| M3 | Lo mismo con el código y los valores 0.028 y 0.017 | Las 7 de M2 y `TestEvalsDelRepositorio` (`formato`, `conjunto-jurisprudencia`), que compara las clases leídas con `Umbral: 0`: 8 funciones y 53 subtests. `TestEsquemasPublicados` (`schema-check`), `ok` |
| M4 | La ejecución sintética «como el cierre» en dos modos (108 respuestas juzgadas), con la `clases.yaml` de `boe-legislacion` copiada a su carpeta del juez y N respuestas con tres votos afirmativos | 3 y 0, 0 y 3, 2 y 1: `aprobado`, sin motivos. 4 y 0, 2 y 2: `fallo`, con un motivo. 13 elementos, 9 con `decide: true`, en todos |
| M5 | La misma, con cuatro sesiones del modelo que decide cortadas tras su respuesta —la tercera de las evals informativas 11 a 14, dos de cada modo—: 104 respuestas juzgadas | Con 2 marcadas, `aprobado` («2 de 104», cumple). Con 3, `fallo`: «3 de 104 (2,9 %), y tiene que ser ≤ 2,8 %». Con una sola cortada y 3 marcadas, `fallo` con «3 de 107 (2,8 %), y tiene que ser ≤ 2,8 %», como fijó la clarificación del spec |
| M6 | Las sesiones sintéticas de `jurisprudencia` con las evals y la carpeta del juez del repositorio (60 respuestas juzgadas) | 1 marcada, en cualquiera de los dos modos: `aprobado`, y el recuento de su modo, «1 de 30», con `cumple: false` y `decide: false`. 2 en un modo, o 1 en cada uno: `fallo`, «2 de 60 (3,3 %), y tiene que ser ≤ 1,7 %». 13 elementos, 9 que deciden |
| M7 | Los bytes de cada elemento escrito, con todo en 0 | `boe-legislacion`: 13 elementos, 3 822 bytes (3 836 con comas y corchetes); eran 3 482. `jurisprudencia`: 4 099 (4 113); eran 3 757. El umbral conjunto mide 335 y 334 bytes; un recuento por modo, de 329 a 342 |
| M8 | Los bytes del motivo | Cabecera, 90 bytes (`boe-legislacion`) y 89 (`jurisprudencia`). Cada respuesta añade a lo de hoy ` (modo orden)`, 13 bytes, o ` (modo herramienta)`, 19. Con las frases del cierre de H25 (V16) son unos 440 bytes por respuesta: 1,9 KB con 4 y 1 KB con 2 |
| M9 | Los dos mutantes sobre el prototipo, uno cada vez | **Un solo modo** (el conjunto suma solo el primer modo): `boe-legislacion` 3 y 0 y 2 y 1 pasan a `fallo`, y `jurisprudencia` 1 y 0 pasa a `fallo`; en todos los casos cambian además `medida` y `total` (54, o 30). **Sin decidir** (`Decide: false` en el conjunto): todos dan `aprobado`, también 4 y 0, 2 y 2, 104 con 3 y los dos de 2 en `jurisprudencia`, y deciden 8 |
| M10 | Las dos sustituciones de `sed` del quickstart, sobre el `umbrales.go` del prototipo | Cada una cambia una sola línea, dentro de `umbralesConjuntos`: la 320 (`range j.deLosModos()` → `…[:1]`) y la 331 (`Decide: true,` → `false,`). Las otras tres vueltas `range …deLosModos()` del fichero no cambian |
| M11 | El test de los datos con los literales 0.028 y 0.017 | `TestEvalsDelRepositorio`, `ok`. Con `umbral: 0.03` en la `clases.yaml` de `boe-legislacion`, `FAIL` en `TestEvalsDelRepositorio/formato`, que enseña `0.028` frente a `0.03` |
| M12 | `jq -n '[(3/108) <= 0.028, (4/108) <= 0.028, (2/104) <= 0.028, (3/104) <= 0.028, (3/107) <= 0.028, (1/60) <= 0.017, (2/60) <= 0.017, (1/58) <= 0.017, (3/54) <= 0.028, (1/30) <= 0.017]'` | `[true,false,true,false,false,true,false,false,false,false]`: lo mismo que da Go en M4 a M6. El informe final, que rehace la comparación con `jq`, no verá un `cumple` incoherente |
| M13 | `git worktree add --detach` y `git worktree remove --force` con un directorio de `$TMPDIR`; `git status --short` después | Solo `?? specs/021-h26-el-umbral-de/`, como al empezar |
| M14 | `go test -count=1 -run '^(TestCopiasDelJuez\|TestMedidaVersionada\|TestEjecucionSinMedir\|TestDefinicionDelJob\|TestJuicioDelSondeo\|TestSalidaDelSondeo\|TestSondear)$' ./internal/evals/`, en el repositorio | En verde (218 tests): es la orden del escenario 5 del quickstart, que vale ya hoy |
| M15 | `contracts/clases-yaml.diff`: `git apply --check` sobre el árbol; y, aplicado en un segundo worktree de usar y tirar, `git apply --reverse --check`, `grep -n '^    umbral:'`, `TestEsquemasPublicados` y `TestEvalsDelRepositorio` | Casa con el árbol de hoy (6 líneas añadidas y 3 quitadas, ninguna de más de 120 columnas). Aplicado: `0.028`, `0`, `0.017` y `0`; la comprobación inversa pasa; el esquema, `ok`; y `TestEvalsDelRepositorio` falla solo en los dos literales (`0` frente a `0.028` y a `0.017`): las declaraciones se leen, y la tarea que las cambia tiene que cambiar el test. Sobre el árbol de hoy, la comprobación inversa dice que el parche no casa |
| M16 | Las órdenes del escenario 6 que no dependen del código nuevo, en ese segundo worktree: la sustitución sobre el `umbrales.go` de hoy, la guarda `git diff --no-index --quiet`, `git -C … checkout --` y `go -C … test -run '^TestUmbralConjunto$'` | Sin el método que fija el plan, la sustitución no cambia nada y la guarda lo dice; con las dos copias mutadas de M10, la guarda calla. El test, `ok … [no tests to run]`. En esta sesión `cmp` sobre rutas de fuera del repositorio lo rechazó el arnés: por eso la guarda es de `git` |

Nota de M4: una expresión de `-run` con una barra dentro de una alternancia no selecciona nada (`go test` la parte por
`/` antes de compilarla). Las órdenes del quickstart nombran solo funciones de test.

## S · Supuestos, sin verificar

- **S1.** `make ci` entero con el cambio no se ha ejecutado: el prototipo corrió `go test` sin `-race`, sin lint y sin
  cobertura. Lo ejecuta cada tarea, y es su verificación.
- **S2.** Lo que tardan los casos nuevos con `-race` no se ha medido. Sin `-race`, cada ejecución de 108 respuestas
  tarda 0,3 s (M4): los siete casos de `boe-legislacion`, unos 2 s.
- **S3.** El escenario 6 del quickstart usa `sed -i.orig` con un rango de direcciones. Se ha probado en macOS (M10). Que
  el `sed` de GNU lo acepte igual no se ha comprobado aquí; el run corre en macOS.
- **S4.** Qué dará el job de cierre. El ADR 0038 calcula que una ejecución de `boe-legislacion` a la tasa de hoy pasa
  el 98 % de las veces con «como mucho 3»; no es una medida de esta sesión. Si sale en rojo por el umbral conjunto, vale
  FR-061.
- **S5.** `scripts/workflow/informe.sh` no se ha ejecutado contra un informe con el umbral conjunto: lo ejecuta el
  workflow. Se ha leído su código (V11) y rehecho sus comparaciones (M12).

## D · Decisiones

### D1 · Dónde nace el umbral conjunto

- **Decisión**: un método de `juicioDelJuez`, `umbralesConjuntos`, que da un umbral por cada clase con `decide: true`,
  en el orden de `clases.yaml`: su medida es la suma de `marcadasEn` de los grupos de los modos y su total, la suma de
  sus respuestas. `umbralesDelInforme` lo añade detrás de los umbrales de las respuestas de los modos y delante de los
  dos de la medida (FR-005).
- **Por qué**: es la suma de lo que ya miden los recuentos por modo (FR-002), sobre los mismos grupos, así que no
  puede contar otra cosa que ellos: ni la eval sin binario ni servidor ni el modelo informativo (V6). No añade ningún
  tipo, ningún campo ni ningún caso por skill (FR-001).
- **Rechazadas**: componerlo dentro de `respuestasMedidas.umbrales`, que recorre los modos y mezcla lo de un modo con
  lo que no es de ninguno; un campo nuevo en `Umbral` que diga a qué modos se refiere, que cambia el contrato del ADR
  0029 (fuera de alcance); y calcularlo en `motivosDelInforme`, fuera de `umbrales`, que lo dejaría sin publicar.

### D2 · El recuento por modo deja de decidir

- **Decisión**: `umbralesDeLasClases` escribe `Decide: false` en el elemento de toda clase. Lo demás del elemento no
  cambia: su nombre, su descripción, su medida, su total y el `umbral` de la clase (FR-010, FR-011).
- **Por qué**: con ello `clase.Decide` dice tres cosas y ninguna más: si la clase tiene umbral conjunto, si tiene los
  dos de la medida y con cuántos votos se marca. El `cumple` del recuento sigue siendo el del contrato.
- **Rechazadas**: quitar el recuento de la clase que decide (el hito pide publicarlo); darle otro umbral, por ejemplo
  el doble, para que su `cumple` no salga en falso con 3 de 54 (el hito pide «el mismo `umbral`»).

### D3 · El nombre

- **Decisión**: `<clase>:<modelo>`, de una función `nombreDelUmbralConjunto(clase, modelo)`; el del recuento por modo
  pasa a componerse con ella y su modo, y sigue dando lo que hoy.
- **Rechazadas**: un sufijo (`:todos`, `:conjunto`), que no es el nombre que fijan el hito y FR-001.

### D4 · El motivo nombra cada respuesta con su modo

- **Decisión**: detrás de la medida y la condición, cada respuesta marcada como `<sesión> (modo <modo>): «<frase 1>» ·
  «<frase 2>» · «<frase 3>»`, separadas por `; `, las del modo orden delante y, en un modo, en orden de sesión. Es la
  forma de hoy con ` (modo <modo>)` detrás de la sesión. `marcadasDelUmbral` pasa a casar el nombre del umbral conjunto
  y recorre los grupos de los modos.
- **Por qué**: el spec pide que el modo se lea sin deducirlo del nombre de la sesión (FR-004), y el nombre solo lo deja
  ver en el modo herramienta. El paréntesis detrás de la sesión es como el informe ya escribe lo que acompaña a una
  sesión: `<sesión> (<motivo>)` en las respuestas sin juzgar y `<ECLI> (<condición>)` en `cita_sin_documento`.
- **Rechazadas**: agrupar por modo con un encabezado («en el modo orden: …»), que deja el modo de una respuesta varias
  respuestas atrás y obliga a decidir qué se escribe de un modo sin marcas; y conservar además la rama que casa
  `<clase>:<modelo>:<modo>`, que desde D2 no la alcanza ningún umbral que decida (V5): se retira.

### D5 · La descripción

- **Decisión**: «Respuestas de `<modelo>`, con todos los modos juntos, que el juez marca en `<clase>` con sus tres
  votos, sobre sus respuestas juzgadas en las evals que activan la skill».
- **Por qué**: es la del recuento por modo con «con todos los modos juntos» donde aquella dice «en el modo `<modo>`»
  (FR-001).

### D6 · Los tests de `boe-legislacion`, con la ejecución sintética y la declaración del repositorio

- **Decisión**: un test nuevo, `TestUmbralConjunto`, con la ejecución sintética «como el cierre» en dos modos —la de
  `TestUmbralesDelInforme`: 18 evals que activan la skill, 54 respuestas del modelo que decide por modo— y el votante
  de votos grabados de `TestInformeConElJuez`. La carpeta del juez de esa ejecución lleva la `clases.yaml` de
  `evals/boe-legislacion/juez/`, leída del repositorio por el test: el arnés gana para ello un campo, `clasesDelJuez`,
  como el que ya tiene para la medida. Lo que el test espera va escrito a mano: 0.028, «≤ 2,8 %», 3 de 108.
- **Por qué**: son sesiones sintéticas escritas con `EscribirInforme`, como pide el hito, y dan el veredicto. Y el
  umbral que comparan es el dato que leerá el job, no una copia: si alguien lo cambia, el caso de 3 de 108 o el de 4 lo
  ve. Verificado en M4 y M5.
- **Rechazadas**: armar sesiones que pasen las 21 evals reales de `boe-legislacion`, como se hizo para las 10 de
  `jurisprudencia`: un arnés nuevo para un código que no tiene ningún caso por skill; probar `umbralesDelInforme` en
  memoria, sin sesiones, que no ve el veredicto ni el motivo de la raíz; y escribir `umbral: 0.028` en una constante del
  test, que seguiría en verde con otro valor en el repositorio.

### D7 · Las 104 respuestas

- **Decisión**: la misma ejecución con cuatro sesiones del modelo que decide cortadas tras su respuesta
  (`cortarTrasLaRespuesta`), una por serie y dos por modo, en series informativas: no cuentan como respuesta medida y
  su serie no da ningún motivo. Quedan 52 por modo.
- **Por qué**: es la vía real por la que el job llega a menos de 108 sin otro motivo de fallo (spec, «Edge Cases»).
  Verificado en M5.
- **Rechazadas**: una eval menos, que da 102 y no 104; sesiones sin medir por un límite de uso, que ponen el veredicto
  en `fallo` por su motivo y no dejan ver el del umbral.

### D8 · Los tests de `jurisprudencia`, en su tabla de hoy

- **Decisión**: tres casos en `TestUmbralesDeJurisprudencia`, con las evals y la carpeta del juez del repositorio: el
  de una marcada en el modo orden, que ya existe y pasa de `fallo` a `aprobado`, y dos nuevos, dos marcadas en el modo
  orden y una en cada modo. El de una marcada en las dos clases, en el modo herramienta, pasa también a `aprobado`.
- **Rechazadas**: un test aparte con su propio arnés, que repetiría el de la tabla.

### D9 · El test de los datos es `TestEvalsDelRepositorio`

- **Decisión**: sus dos comparaciones de las clases leídas pasan a esperar `Umbral: 0.028` y `Umbral: 0.017`, escritos
  a mano. Corre en `test` y en `skills-check` (V18). Verificado en M11.
- **Por qué**: es donde `make ci` ya compara las clases del repositorio con un literal. La igualdad es exacta a
  propósito: los dos lados son el mismo decimal convertido a `float64`, y lo que se quiere es el valor declarado, no uno
  cercano.
- **Rechazadas**: un test nuevo que haga lo mismo; leer el número del texto del ADR, que ataría un test a la prosa de
  un documento.

### D10 · La respuesta sin juzgar y la clase que solo se publica, en los casos que ya hay

- **Decisión**: FR-075 y FR-076 los cubren casos de hoy cuya lista esperada de umbrales pasa a llevar el conjunto:
  `un-voto-que-no-llega` (el conjunto con `cumple: true`, la respuesta en su `total` y fuera de su `medida`, y `fallo`
  por el motivo de la ejecución), `cuenta-su-proceso-con-tres` y, en `jurisprudencia`,
  `tres-con-si-en-afirma-que-existe` (la lista es exacta: ningún elemento sin modo de la clase que solo se publica).
  Los trece elementos de cada skill, escritos a mano y en su orden, los fijan `TestUmbralesDelInforme` y
  `TestUmbralesDeJurisprudencia`.
- **Rechazadas**: tests nuevos para lo que esos casos ya arman.

### D11 · Los mutantes, en una copia fuera del árbol

- **Decisión**: la tarea de FR-078 los aplica en un worktree de la cabeza en `$TMPDIR`, uno cada vez, con las dos
  sustituciones del escenario 6 del quickstart, ancladas a `umbralesConjuntos`; ejecuta allí `TestUmbralConjunto` y
  `TestUmbralesDeJurisprudencia`; escribe en `specs/021-h26-el-umbral-de/mutantes.md` qué subtests fallan con cada uno;
  y retira el worktree.
- **Por qué**: ningún mutante puede quedar en el árbol porque nunca está en él. El registro es donde «la tarea dice qué
  test falla con cada mutante» (FR-078), y la revisión final puede repetir el escenario. Verificado en M9, M10, M13 y
  M16.
- **Rechazadas**: editar `umbrales.go` en el árbol y deshacerlo, que deja el mutante a un fallo de la sesión de quedar
  en el diff; un parche para `git apply`, que no casa si el código final difiere en una línea de contexto; y una
  herramienta de mutación, que sería una dependencia nueva.

### D12 · El orden

- **Decisión**: (1) el mecanismo, con los tests de hoy puestos a su forma nueva y los valores todavía en 0; (2) los
  valores, en una tarea `[datos]`, con el test de los datos y los casos que dependen de ellos; (3) los mutantes; (4) los
  documentos.
- **Por qué**: con `umbral: 0`, el conjunto da el mismo veredicto que los dos recuentos de hoy (0 de 108 es 0 de 54 en
  cada modo), así que el paso 1 cambia la forma y ningún veredicto, y sus siete tests en rojo (M2) se corrigen en él. El
  paso 2 cambia dos números y lo que los prueba. Cada paso deja `make ci` en verde.
- **Rechazadas**: una sola tarea con todo, que pone en un mismo intento el código, los datos y ocho funciones de test;
  y los valores primero, que deja a mitad un job que decidiría por modo con la tasa y obliga a escribir dos veces lo
  que esperan los tests de `jurisprudencia`.

### D13 · Las dos `clases.yaml`

- **Decisión**: el valor y una frase en el comentario de cabecera que dice de dónde sale y quién lo estrecha. El diff
  exacto está en `contracts/clases-yaml.diff`, aplicable con `git apply`.
- **Por qué**: los valores son los del ADR y no los elige el run (FR-020): con el diff, la tarea no escribe ningún
  número de memoria.
- **Rechazadas**: una clave nueva para la procedencia del valor, que cambia el esquema (fuera de alcance).

### D14 · Sin aceptación e2e del binario

- **Decisión**: «Aceptación e2e: no aplica». El binario no enlaza `internal/evals` (V14) y el hito no cambia nada de
  lo que ve quien usa `kitlegal`. La aceptación es el job de cierre (SC-001), y en `make ci`, los tests de «Controles
  de umbral».
- **Rechazada**: un guion `testscript` sobre el binario sin cambios, que no describiría la entrega. Precedentes: H24 y
  H25.

### D15 · Sin datos externos

- **Decisión**: ninguna fuente, ningún manifiesto `grabaciones.json`, ningún `TestGrabar*`. Las tareas `[datos]` del
  hito solo cambian las dos `clases.yaml`, que no son `testdata/` ni `schemas/`: llevan la etiqueta porque lo pide
  FR-020, y su comprobación de grabación no tiene nada que comprobar (V12).

### D16 · El sondeo

- **Decisión**: `sondeo.go` y sus tests no cambian, y no se añade ningún test. FR-040 fija su comprobación: los tests
  que ya hay siguen pasando con las mismas líneas (M2, M14).
- **Rechazada**: un caso nuevo con «1 de 9 (11,1 %)», el ejemplo del spec: probaría un código que el hito no toca.

### D17 · Los comentarios de `internal/evals` que el cambio deja falsos

- **Decisión**: se corrigen con el código, en la misma tarea: en `umbrales.go` (que cada umbral es de un solo modo, que
  ninguno suma dos modos, «doce, y diez deciden»), en `doc.go` (lo que dice de qué umbrales lleva cada skill y cuáles
  deciden), en `informe.go` (cómo se componen) y en los tests que cambian.
- **Rechazada**: tocar el comentario de `.github/workflows/evals.yml` (V22): la clarificación del spec lo deja fuera, y
  queda como pendiente para la persona.

### D18 · La constante de clases de los tests del informe

- **Decisión**: se queda con `umbral: 0`, para los casos que fijan el mecanismo con pocas respuestas; su comentario deja
  de decir que es el umbral de `boe-legislacion`. Los casos que prueban la tasa usan la declaración del repositorio (D6).
- **Rechazada**: pasarla a 0.028, que son dos sitios con el mismo número.

### D19 · La comparación en coma flotante

- **Decisión**: se queda como está, `medida/total <= umbral` sin tolerancia: lo fija el contrato del ADR 0029.
- **Por qué**: no hay empate alcanzable en el que el redondeo decida. 0.028 es 7/250 y 0.017 es 17/1000: la proporción
  solo iguala el umbral con un total múltiplo de 250 o de 1000. Con los totales de hoy, lo más cercano es 3 de 107, a
  0,00004 del umbral. `jq` da lo mismo (M12).
- **Rechazada**: una tolerancia, que es lo que la skill `golang-safety` aconseja para comparar con `==` el resultado
  de una operación en coma flotante. Aquí no se compara con `==`, y una tolerancia dejaría el `cumple` del job sin
  poder rehacerse con los otros campos del elemento, que es lo que hace el informe final.

### D20 · Dónde vive el test nuevo

- **Decisión**: `TestUmbralConjunto`, en `internal/evals/umbrales_test.go`, tabla con subtests con nombre y
  `t.Parallel`, cada caso con su directorio temporal.
- **Por qué**: lo que prueba está en `umbrales.go`, y el repositorio pone cada test en el `_test.go` de su fichero
  (skill `golang-testing`: el fichero de test se nombra por el fichero que prueba). Usa los ayudantes de
  `informe_test.go`, que son del mismo paquete.
