# Registro de los mutantes: H20

Lo escribe la tarea T012 (FR-090, SC-008; plan.md, paso 9 y fila de FR-090 de «Controles de umbral»; quickstart.md §7;
contracts/evals-redaccion-fechada.md §5). Los cinco mutantes son los de la tarea, en su orden. Es un registro del run:
nada lo lee en `make ci`.

**Resultado: 5 de 5.** Cada mutante deja en rojo al menos uno de los tests de su fila, y ninguno ha necesitado que se
refuerce un test: T012 no cambia nada bajo `internal/`.

| Mutante | Tests en rojo (subtests en rojo) | Órdenes de §7 con código 1 |
|---|---|---|
| 1, la versión elegida por `fecha_publicacion` | `TestArticuloEnUnaFecha` (7), `TestObservadoDeBoe` (1) | 2 de 4 |
| 2, siempre la última | `TestArticuloEnUnaFecha` (4), `TestObservadoDeBoe` (1), `TestLecturaFechadaEnElGrafo` (2) | 3 de 4 |
| 3, con una fecha anterior al bloque, la primera versión | `TestArticuloEnUnaFecha` (2), `TestObservadoDeBoe` (1) | 2 de 4 |
| 4, la lectura fechada contada como lectura | `TestObservadoDeBoe` (1), `TestLecturaFechadaEnElGrafo` (2) | 2 de 4 |
| 5, el `decide` de `umbralDeRedaccionSinFechar` en falso | `TestUmbralesDeLasTresSkills` (1), `TestUmbralesDeBoeLegislacion` (13) | 1 de 4 |

En la columna de los subtests cuentan las hojas: `TestArticuloEnUnaFecha/un_bloque_que_no_se_puede_situar`, que agrupa
casos, no se cuenta aparte. Cada mutante deja en rojo los tests que le asigna contracts/evals-redaccion-fechada.md §5, y
los que no toca salen con `ok`: los mutantes 1 a 4 no tocan `internal/evals`, y el 5 no toca `internal/source/boe` ni
`internal/app`. Que los mutantes 1 y 3 no dejen en rojo `TestLecturaFechadaEnElGrafo` no contradice el contrato, que
les asigna como control `TestArticuloEnUnaFecha` (y los dos lo dejan en rojo); el 2 lo deja en rojo además de los suyos.

## Dónde y sobre qué

- Cabeza: `693e6a6c318afaa178b2211670163af44683e5f3` (`feat(H20): T011`). `go version go1.27.2 darwin/arm64`. El árbol
  de trabajo solo tenía cambios en `specs/024-h20-boe-articulo-fecha/gates/` (`git diff --quiet HEAD -- internal/ data/
  skills/ evals/` salió con 0).
- Árbol aparte: `git worktree add --detach "$TMPDIR/h20-mutantes" HEAD` (`/private/tmp/claude-501/h20-mutantes`, en
  `693e6a6`), fuera del repositorio. Ningún mutante se aplicó en el árbol de trabajo.
- Cada mutante se aplicó con la herramienta de edición sobre el fichero del árbol aparte, cambiando una sola línea, y su
  cambio se comprobó con `git -C "$TMPDIR/h20-mutantes" diff -U0`, que se copia abajo tal cual. Tras cada uno,
  `git -C "$TMPDIR/h20-mutantes" checkout -- <fichero>` restauró la cabeza antes del siguiente: el `diff` de cada
  mutante enseña solo su línea.
- Las órdenes, con cada mutante y sin ninguno, son las cuatro de quickstart.md §7, en primer plano, en un guion de un
  solo uso fuera del repositorio y del árbol aparte (`$TMPDIR/h20-ordenes/correr.sh`), que las ejecuta en este orden,
  deja la salida entera de `go test` en `$TMPDIR/h20-ordenes/mN.log`, precedida del `git diff -U0` del árbol aparte, y
  escribe tras cada una `== exit=<código>`; el guion se lanzó con `rtk proxy bash`:

  ```bash
  go -C "$TMPDIR/h20-mutantes" test -count=1 -run '^TestArticuloEnUnaFecha$' ./internal/source/boe/
  go -C "$TMPDIR/h20-mutantes" test -count=1 -run '^TestObservadoDeBoe$' ./internal/source/boe/
  go -C "$TMPDIR/h20-mutantes" test -count=1 -run '^TestLecturaFechadaEnElGrafo$' ./internal/app/
  go -C "$TMPDIR/h20-mutantes" test -count=1 -run '^(TestUmbralesDeBoeLegislacion|TestUmbralesDeLasTresSkills)$' ./internal/evals/
  ```

  Los mensajes de abajo son líneas de esos registros, leídas con `grep` y con la lectura de sus líneas; cada mensaje va
  con su subtest por el campo `Test:` de su aserción y con el sitio de su `Error Trace`. Un valor muy largo se recorta
  con `[…]`.
- El árbol aparte se retiró al terminar con `git worktree remove --force "$TMPDIR/h20-mutantes"`, y `git worktree list`
  ya no lo nombra: solo enseña el árbol de trabajo y `wt-marcas` (rama `workflow-marcas-de-escritura`), que no es de
  esta tarea y sigue como estaba.

## Sin mutar

Las cuatro órdenes sobre el árbol aparte recién creado salieron con 0:

```text
ok  	github.com/jmorenobl/kitlegal/internal/source/boe	0.476s
ok  	github.com/jmorenobl/kitlegal/internal/source/boe	0.288s
ok  	github.com/jmorenobl/kitlegal/internal/app	3.193s
ok  	github.com/jmorenobl/kitlegal/internal/evals	3.147s
```

## Mutante 1: la versión elegida por `fecha_publicacion`

Las versiones se sitúan por su `fecha_version`, que es la `fecha_publicacion` del BOE, y no por su `fecha_vigencia`.
`internal/source/boe/fecha.go`, en `vigenciasDeLasVersiones`:

```diff
@@ -190 +190 @@ func vigenciasDeLasVersiones(versiones []versionGuardada) ([]time.Time, bool) {
-		vigencia, err := time.Parse(formaDeLaVigencia, version.FechaVigencia)
+		vigencia, err := time.Parse(formaDeLaVigencia, version.FechaVersion)
```

Códigos de las órdenes: 1, 1, 0, 0.

`TestArticuloEnUnaFecha`, en rojo:

- `TestArticuloEnUnaFecha/entre_publicación_y_vigencia` (`fecha_test.go:150`, desde `metadatos_test.go:467`):
  `An error is expected but got nil.` `expected: *boe.Error`. El art. 21 de la Ley 39/2015 el 2016-06-01 sale con 0.
- `TestArticuloEnUnaFecha/anterior_a_la_primera_vigencia` (`fecha_test.go:150`, desde `metadatos_test.go:467`): el mismo
  mensaje. El art. 22 de la Ley 7/1985 el 1985-04-22 sale con 0.
- `TestArticuloEnUnaFecha/la_víspera,_ya_publicada_la_de_ese_día` (`fecha_test.go:164`): `Not equal:`, con este diff:

  ```text
  -  FechaVersion: (string) (len=8) "20070529",
  -  FechaVigencia: (string) (len=8) "20070701",
  -  NormaModificadora: (string) (len=16) "BOE-A-2007-10701",
  +  FechaVersion: (string) (len=8) "20071031",
  +  FechaVigencia: (string) (len=8) "20080430",
  +  NormaModificadora: (string) (len=16) "BOE-A-2007-18874",
  [...]
  -  HashTexto: (string) (len=71) "sha256:50a1cc5d30109aeb4747a003df8b6855ab23631c7f3fcf81b6e068f7a571343e",
  +  HashTexto: (string) (len=71) "sha256:4c8982b1fee3d8414a0b3fdf7d71e4b269b5a1ec43197edeae71a171e1d151fe",
  ```

- En `TestArticuloEnUnaFecha/un_bloque_que_no_se_puede_situar`, los cuatro bloques con alguna versión sin una
  `fecha_vigencia` que sea un día, cada uno con dos aserciones en rojo desde `fecha_test.go:247`:
  - `.../una_versión_con_una_fecha_vigencia_que_no_es_un_día`, `.../una_versión_con_fecha_vigencia_vacía` y
    `.../una_versión_sin_fecha_vigencia`: en `metadatos_test.go:473`, `Not equal:` con
    `expected: "el BOE no da la fecha de vigencia de cada redacción del bloque a21 de BOE-A-2015-10565: no se puede saber
    cuál tenía el 2000-01-01 (https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21)"`
    y `actual  : "el bloque a21 de BOE-A-2015-10565 no tenía redacción vigente el 2000-01-01: su primera redacción tiene
    fecha de vigencia 20100101 (https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21)"`;
    y en `metadatos_test.go:467`, `An error is expected but got nil.` `expected: *boe.Error`.
  - `.../la_primera_versión_sin_fecha_vigencia`: lo mismo, con `actual  : "el bloque a21 de BOE-A-2015-10565 no tenía
    redacción vigente el 2000-01-01: su primera redacción tiene fecha de vigencia  (https://[…]/bloque/a21)"` —la fecha
    de vigencia, vacía—.

`TestObservadoDeBoe`, en rojo:

- `TestObservadoDeBoe/articulo-con-fecha-sin-redaccion-vigente` (`grafo_test.go:72`, desde `grafo_test.go:262`):
  `An error is expected but got nil.` `Messages: premisa: la lectura fechada falla`.

`TestLecturaFechadaEnElGrafo`, `TestUmbralesDeBoeLegislacion` y `TestUmbralesDeLasTresSkills`: `ok`.

## Mutante 2: siempre la última

La versión con la que se compone el artículo es siempre la última del bloque; la elección del día pedido sigue
decidiendo el 3 y la redacción posterior. `internal/source/boe/fecha.go`, en `elegirLaVersion`:

```diff
@@ -154 +154 @@ func elegirLaVersion(norma, bloque, fecha string, guardadas consultaResuelta[ver
-	version := versiones[elegida]
+	version := versiones[len(versiones)-1]
```

Códigos de las órdenes: 1, 1, 1, 0.

`TestArticuloEnUnaFecha`, en rojo, las cuatro con `Not equal:` en `fecha_test.go:164` y la última versión del bloque en
lugar de la del día pedido:

- `TestArticuloEnUnaFecha/US1`: `expected: […] FechaVersion:"20171109", FechaVigencia:"20180309",
  NormaModificadora:"BOE-A-2017-12902" […]`; `actual  : […] FechaVersion:"20200205", FechaVigencia:"20200206",
  NormaModificadora:"BOE-A-2020-1651" […]`.
- `TestArticuloEnUnaFecha/el_día_que_entra_en_vigor`: `expected: […] FechaVersion:"20071031", FechaVigencia:"20080430",
  NormaModificadora:"BOE-A-2007-18874" […]`; `actual  : […] FechaVersion:"20151031", FechaVigencia:"20151031",
  NormaModificadora:"BOE-A-2015-11723" […]`.
- `TestArticuloEnUnaFecha/la_víspera,_ya_publicada_la_de_ese_día`: `expected: […] FechaVersion:"20070529",
  FechaVigencia:"20070701", NormaModificadora:"BOE-A-2007-10701" […]`; `actual  : […] FechaVersion:"20151031",
  FechaVigencia:"20151031", NormaModificadora:"BOE-A-2015-11723" […]`.
- `TestArticuloEnUnaFecha/intermedia`: `expected: […] FechaVersion:"20031217", FechaVigencia:"20040101",
  NormaModificadora:"BOE-A-2003-23103" […]`; `actual  : […] FechaVersion:"20151031", FechaVigencia:"20151031",
  NormaModificadora:"BOE-A-2015-11723" […]`.

`TestObservadoDeBoe`, en rojo:

- `TestObservadoDeBoe/articulo-con-fecha` (`grafo_test.go:80`, desde `grafo_test.go:196`): `Not equal:` con
  `expected: "sha256:ce4210549efa0b8a103e99ccfd4137f6b66a5f300f661a94e2d2545b08847183"` y
  `actual  : "sha256:2a29d264c6a3749916785fdd84c817b6b0d83925e78048380f3903d469aab4ea"`.
  `Messages: premisa: la versión elegida es la de 20040101`.

`TestLecturaFechadaEnElGrafo`, en rojo:

- `TestLecturaFechadaEnElGrafo/sin_fila_de_lecturas_del_bloque,_una_versión_nueva` (`grafo_test.go:2326`): `Not equal:`
  con `expected: "20040101"` y `actual  : "20151031"`. `Messages: premisa: la versión elegida es la del caso`.
- `TestLecturaFechadaEnElGrafo/con_la_fila_de_lecturas_del_bloque,_una_versión_nueva` (`grafo_test.go:2326`): el mismo
  mensaje.

`TestUmbralesDeBoeLegislacion` y `TestUmbralesDeLasTresSkills`: `ok`.

## Mutante 3: con una fecha anterior al bloque, la primera versión

La elección parte de la primera versión en lugar de ninguna: un día anterior a la primera vigencia la elige en lugar de
salir con 3. `internal/source/boe/fecha.go`, en `elegirLaVersion`:

```diff
@@ -141 +141 @@ func elegirLaVersion(norma, bloque, fecha string, guardadas consultaResuelta[ver
-	elegida := -1
+	elegida := 0
```

Códigos de las órdenes: 1, 1, 0, 0.

`TestArticuloEnUnaFecha`, en rojo:

- `TestArticuloEnUnaFecha/entre_publicación_y_vigencia` (`fecha_test.go:150`, desde `metadatos_test.go:467`):
  `An error is expected but got nil.` `expected: *boe.Error`. Es el `a21` de la Ley 39/2015 el 2016-06-01.
- `TestArticuloEnUnaFecha/anterior_a_la_primera_vigencia` (`fecha_test.go:150`, desde `metadatos_test.go:467`): el mismo
  mensaje. Es el art. 22 de la Ley 7/1985 el 1985-04-22.

`TestObservadoDeBoe`, en rojo:

- `TestObservadoDeBoe/articulo-con-fecha-sin-redaccion-vigente` (`grafo_test.go:72`, desde `grafo_test.go:262`):
  `An error is expected but got nil.` `Messages: premisa: la lectura fechada falla`.

`TestLecturaFechadaEnElGrafo`, `TestUmbralesDeBoeLegislacion` y `TestUmbralesDeLasTresSkills`: `ok`.

## Mutante 4: la lectura fechada contada como lectura

La lectura fechada emite las operaciones de la lectura sin fecha, `operacionesDelArticulo` —la Norma, el Bloque, la
BloqueVersion, la arista `eli:has_part`, la `eli:has_version` y el texto—, con la misma vigencia y el mismo
`SoloLoNuevo`. `internal/source/boe/grafo.go`, en `observadoDeLaLecturaFechada`: la línea en blanco tras la rama sin ELI
pasa a devolver eso, y el `return` original queda detrás, sin alcanzarse:

```diff
@@ -87 +87 @@ func observadoDeLaLecturaFechada(articulo Articulo) schema.Observado {
-
+	return schema.Observado{Vigencia: vigenciaLarga, SoloLoNuevo: true, Operaciones: operacionesDelArticulo(articulo)}
```

Códigos de las órdenes: 0, 1, 1, 0.

`TestObservadoDeBoe`, en rojo:

- `TestObservadoDeBoe/articulo-con-fecha` (`grafo_test.go:80`, desde `grafo_test.go:212`), dos veces, `Not equal:` con
  `expected: schema.Observado{Vigencia:604800000000000, SoloLoNuevo:true, Operaciones:[]schema.Operacion{schema.Nodo{ID:"eli/es/l/1985/04/02/7#a22@20040101:sha256:ce4210549efa0b8a103e99ccfd4137f6b66a5f300f661a94e2d2545b08847183", […]`
  y `actual  : schema.Observado{Vigencia:604800000000000, SoloLoNuevo:true, Operaciones:[]schema.Operacion{schema.Nodo{ID:"eli/es/l/1985/04/02/7", Tipo:"Norma", Datos:map[string]interface {}{"identificador":"BOE-A-1985-5392"}}, […]`;
  una con `Messages: con {JSON:false Timeout:0s Offline:false DryRun:false SinGrafo:false Asunto:}` y otra con
  `Messages: con {JSON:false Timeout:0s Offline:true DryRun:false SinGrafo:false Asunto:}`.

`TestLecturaFechadaEnElGrafo`, en rojo, los dos subtests con las mismas siete aserciones:

- `TestLecturaFechadaEnElGrafo/con_la_fila_de_lecturas_del_bloque,_una_versión_nueva` y
  `TestLecturaFechadaEnElGrafo/sin_fila_de_lecturas_del_bloque,_una_versión_nueva`:
  - `grafo_test.go:2331`: `elements differ`, `extra elements in list B:` `"arista eli:has_version"`, con
    `listA: "nodo BloqueVersion", "texto"` y `listB: "texto", "nodo BloqueVersion", "arista eli:has_version"`.
    `Messages: lo que crea la lectura fechada`.
  - `grafo_test.go:2335`, tres `Not equal:`, con `Messages: tras la lectura fechada: la misma fila de lecturas`,
    `Messages: tras la lectura fechada: el mismo graph check del bloque` y
    `Messages: tras la lectura fechada: el mismo graph check de todo`. En el subtest con la fila, la de lecturas del
    bloque pasa de `Ultima:"eli/es/l/1985/04/02/7#a22@20151031:sha256:2a29d264c6a3749916785fdd84c817b6b0d83925e78048380f3903d469aab4ea"`
    a `Ultima:"eli/es/l/1985/04/02/7#a22@20040101:sha256:ce4210549efa0b8a103e99ccfd4137f6b66a5f300f661a94e2d2545b08847183"`;
    en el subtest sin ella, el esperado tiene la fila del `a21` de la Ley 39/2015
    (`Bloque:"eli/es/l/2015/10/01/39#a21"`) y la primera del real es una fila nueva del art. 22 de la Ley 7/1985 con
    `Ultima:"eli/es/l/1985/04/02/7#a22@20040101:[…]"`. Las dos de graph check dan otra salida de `kitlegal.graph`, con
    otro `hash`.
  - `grafo_test.go:2345`, tres `Not equal:`, con
    `Messages: la siguiente lectura sin fecha, con la lectura fechada y sin ella: la misma fila de lecturas`,
    `[…]: el mismo graph check del bloque` y `[…]: el mismo graph check de todo`.

`TestArticuloEnUnaFecha`, `TestUmbralesDeBoeLegislacion` y `TestUmbralesDeLasTresSkills`: `ok`.

## Mutante 5: el `decide` de `umbralDeRedaccionSinFechar` en falso

Los dos umbrales `redaccion_sin_fechar:<modelo>:<modo>` se publican sin decidir. `internal/evals/umbrales.go`, en
`umbralDeRedaccionSinFechar`:

```diff
@@ -668 +668 @@ func umbralDeRedaccionSinFechar(modelo string, delModo grupoJuzgado) Umbral {
-		Decide:      true,
+		Decide:      false,
```

Códigos de las órdenes: 0, 0, 0, 1.

`TestUmbralesDeLasTresSkills`, en rojo:

- `TestUmbralesDeLasTresSkills/boe-legislacion`:
  - `informe_test.go:6672`: `Not equal:` entre los umbrales esperados y los del informe, con dos hunks en el diff, uno
    por modo: `-  Decide: (bool) true` / `+  Decide: (bool) false`.
  - `informe_test.go:6677`: `should have 11 item(s), but has 9` (los que deciden).

`TestUmbralesDeBoeLegislacion`, en rojo, trece subtests; todas sus aserciones en rojo salen con la línea
`informe_test.go:7063` (la primera de `sin-medir`, con `Error Trace` en `umbrales_test.go:1405`, `informe_test.go:7270` e
`informe_test.go:7063`):

- En los trece —`sin-medir`, `lee-sin-la-fecha`, `la-de-hoy-con-otra-fecha`, `redaccion-a-fecha-en-la-b`, `de-haiku`,
  `la-de-hoy-con-otra-norma`, `sin-redaccion-vigente`, `una-eval-sin-fecha`, `sin-la-de-hoy`, `con-otra-fecha`,
  `lee-con-la-fecha`, `lee-otra-fecha` y `sin-la-linea`—: `Not equal:` entre los umbrales esperados y los del informe
  (`expected: []evals.Umbral{evals.Umbral{Nombre:"sin_activar:claude-sonnet-5-5:orden", […]`), `Not equal:` entre sus
  JSON (`expected: "[{\"nombre\":\"sin_activar:claude-sonnet-5-5:orden\", […]`) y `should have 11 item(s), but has 9`
  con `Messages: deciden todos menos los recuentos de las dos clases del juez de cada modo`.
- En los ocho en que el umbral no se cumple, además, el motivo que falta, la línea que falta y el veredicto:
  `Not equal:` con `expected: []string{"<motivo>"}` y `actual  : []string{}`; `Should be true` con
  `Messages: falta la línea "- <motivo>" en: […]`; y `Not equal:` con `expected: "fallo"` y `actual  : "aprobado"`.
  El motivo de cada uno:
  - `lee-sin-la-fecha`: `umbral redaccion_sin_fechar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 22-lcsp-contrato-menor-en-2019-claude-sonnet-5-5-01`.
  - `la-de-hoy-con-otra-fecha`: `umbral redaccion_sin_fechar:claude-sonnet-5-5:herramienta: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 22-lcsp-contrato-menor-en-2019-herramienta-claude-sonnet-5-5-02`.
  - `redaccion-a-fecha-en-la-b`: `umbral redaccion_sin_fechar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 23-lpac-articulo-21-en-2016-claude-sonnet-5-5-01`.
  - `la-de-hoy-con-otra-norma`: `umbral redaccion_sin_fechar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 22-lcsp-contrato-menor-en-2019-claude-sonnet-5-5-01`.
  - `sin-la-de-hoy`: `umbral redaccion_sin_fechar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 22-lcsp-contrato-menor-en-2019-claude-sonnet-5-5-03`.
  - `con-otra-fecha`: `umbral redaccion_sin_fechar:claude-sonnet-5-5:herramienta: 2 de 6 (33,3 %), y tiene que ser ≤ 0,0 %: 22-lcsp-contrato-menor-en-2019-herramienta-claude-sonnet-5-5-01; 22-lcsp-contrato-menor-en-2019-herramienta-claude-sonnet-5-5-03`.
  - `lee-otra-fecha`: `umbral redaccion_sin_fechar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 22-lcsp-contrato-menor-en-2019-claude-sonnet-5-5-01`.
  - `sin-la-linea`: `umbral redaccion_sin_fechar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: 22-lcsp-contrato-menor-en-2019-claude-sonnet-5-5-02`.

`TestArticuloEnUnaFecha`, `TestObservadoDeBoe` y `TestLecturaFechadaEnElGrafo`: `ok`.
