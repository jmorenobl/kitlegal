# Mutantes y comprobación contra `main` de H45

Registro de T018 (FR-027; FR-008, FR-018, FR-025, FR-026; SC-001 a SC-009, SC-011; plan.md, «Mutantes» y «Controles de
umbral»; research D13). Una entrada por control de «Criterios del hito (literales)» del spec, con la orden, lo esperado
y lo visto. Todo lo que sigue se ejecutó en la sesión de T018 y se vio terminar; lo que no se pudo medir lo dice su
entrada.

## La copia

- `<copia>` es `/tmp/claude/t018`, en el directorio temporal del sandbox: `git clone --no-hardlinks` del repositorio en
  la cabeza de la rama (`7e9029cf`, T017), con `origin/main` en `c715e13a`, que es el `main` del repositorio y la base
  de la rama (`git merge-base main HEAD`). Ningún mutante se aplicó al árbol de trabajo, al índice ni al historial.
- Los dos guiones congelados de `aceptacion/`, copiados sin tocarlos en `<copia>/internal/app/testdata/script/` como
  `h45-paquete-de-mas.txtar` y `h45-skills-instaladas.txtar`, el nombre que les da `scripts/workflow/aceptacion.sh
  activar`.
- Línea base, sin mutante: `go -C <copia> test -count=1 ./internal/skills/ ./internal/evals/` → `ok` en los dos;
  `go -C <copia> test -count=1 -v -run '^TestEntregaDelHito$/^(h45-paquete-de-mas|h45-skills-instaladas)$'
  ./internal/app/` → los dos en `PASS`, con `copia: hecha`, `sincronía: 0`, `tests enteros: 0` y `subtests: 0`.
- Cada mutante: un `sed -i` sobre un fichero de la copia, o `git -C <copia> checkout origin/main -- <fichero>` (el de
  `main` en la copia), o un fichero añadido; su control, con `go -C <copia> test -count=1 -run '<patrón>' <paquete>`; y
  deshecho con `git -C <copia> checkout HEAD -- <carpeta>` o `rm` de lo añadido. Tras cada uno, `git -C <copia> status
  --short` solo daba los dos guiones activados.
- Las comparaciones con `main` se hicieron en el árbol de trabajo con `git diff main -- <rutas>`, que solo lee.

En las tablas, «Orden» es el `-run` y el paquete de `go -C <copia> test -count=1`; «En rojo», lo que se vio.

## Aceptación

### A1 · La vertical de ensayo entra con un tipo y una fila con norma, sin Go ni `SKILL.md` a mano (US2; SC-002)

- Orden: `go test -count=1 -v -run '^TestVerticalDeEnsayo$' ./internal/skills/`, en el árbol.
- Esperado: `PASS`, con `mutante` y `sin-la-region`.
- Visto: `--- PASS: TestVerticalDeEnsayo`, `--- PASS: TestVerticalDeEnsayo/mutante`, `--- PASS:
  TestVerticalDeEnsayo/sin-la-region`. Sus mutantes, en C2.

### A2 · Con los datos de hoy, las referencias generadas salen como en `main`; de las escritas a mano solo cambia `con_herramientas.md` de `normativa` (US1; SC-001)

- `git diff main -- skills/plazos-y-territorio/references/ skills/normativa/references/normas.md | wc -l` → `0`.
- `git diff --numstat main -- skills/ data/` → tres líneas y nada en `data/`: `0 4 skills/normativa/SKILL.md`, `5 0
  skills/normativa/references/con_herramientas.md`, `3 3 skills/plazos-y-territorio/SKILL.md`.
- `git diff main -- skills/normativa/ skills/plazos-y-territorio/SKILL.md`: en `normativa/SKILL.md`, fuera las cuatro
  líneas de §C.2 con su sangría (`@@ -35,10 +35,6 @@`); en `con_herramientas.md`, al final, una línea en blanco y esas
  cuatro sin sangría (`@@ -36,3 +36,8 @@`); en `plazos-y-territorio/SKILL.md`, un solo trozo, `@@ -1,9 +1,9 @@`, las tres
  líneas de la `description` del frontmatter.
- `make -C <copia> skills-sync` → `ok`, y `git -C <copia> status --short` después: solo los dos guiones activados.
- `make skills-check`, en el árbol → `ok` en `internal/app`, `internal/skills` e `internal/evals`.

### A3 · Con una copia de `data/` y `evals/` con un paquete más, ningún test de la lista se pone en rojo (US4; SC-004)

- Orden: `-v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/`, en la copia.
- Esperado: `PASS`, con los cuatro códigos 0.
- Visto: `--- PASS: TestEntregaDelHito/h45-paquete-de-mas (7.65s)`; en su salida, `copia: hecha`, `sincronía: 0`,
  `tests enteros: 0`, `subtests: 0`. Su mutante, L00 de C4.

### A4 · `plazos-y-territorio` se sigue activando ante sus preguntas de hoy con la descripción nueva (US9.3; SC-010)

No medido: lo mide el job de evals del cierre (los umbrales `sin_activar` de cada skill y el juez de las descripciones,
`habla_en_primera_o_segunda_persona`), con sesiones con modelo, que este paso no abre (ADR 0032). Lo que sí se comprobó:
la descripción instalada es la de FR-023 byte a byte (C7) y la variante mide 429 caracteres (`tr -d '\n' <
laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt | wc -m` → `429`).

## Controles

### C1 · Lo de hoy, igual (FR-008, FR-026; US1; SC-001)

Contra `main`: A2. Y en el árbol, `git diff --stat main -- internal/app/testdata/plazos/ internal/core/plazos/
internal/app/plazos.go schemas/ data/` y los cinco `internal/app/testdata/script/h9-plazos-*.txtar` → vacío (`wc -l`:
`0`). `git diff main -- internal/app/plazos_casos_test.go`: dos trozos, el de `19-eval-c-tributos` (`@@ -202,43`) y la
búsqueda de su fila (`@@ -273,19`: `entrada.Procedimiento == caso.procedimiento` → `caso.es(entrada)` y el mensaje del
`require.Len`); ninguna entrada ni fecha esperada de otro caso. `git diff main --
internal/app/testdata/script/h36-plazos-de-hoy.txtar`: el comentario de cabecera y `cmp stdout error-tipo.json` → un
`stdout` con su expresión, con el fichero `error-tipo.json` fuera; nada más.

Verdes, en la copia y en el árbol: `-run '^(TestCasosDePlazos|TestAppletPlazos)$' ./internal/app/` → `ok` (39 subtests
en `PASS` con `-v` en la copia); `-run '^TestEntregaDelHito$/^(h9-plazos-.*|h36-plazos-de-hoy)$' ./internal/app/` →
`h9-plazos-cobertura`, `h9-plazos-calcular`, `h9-plazos-herramienta`, `h9-plazos-sin-efectos`, `h9-plazos-codigos` y
`h36-plazos-de-hoy` en `PASS`.

| # | Mutante | Orden | En rojo |
|---|---|---|---|
| M01 | la regla de hoy, que es también «un tipo asignado a todas las referencias» de US2.4: `filasDeLosTipos(tipos, porIdentificador)` en lugar de `delPaquete` en `internal/skills/regiones.go:584` | `'^TestVerticalDeEnsayo$' ./internal/skills/` | `vertical_de_ensayo_test.go:114`: «plazos-y-territorio: references/administrativo.md: 101 líneas sin un índice que nombre sus apartados ## antes del primero (más de 100)»: la línea 101 es la fila del tipo de ensayo, que la regla de hoy escribe también en `administrativo.md` |
| M01b | M01 con `maximoDeLineasSinIndice = 200` en `internal/skills/markdown.go:27`, para ver la comprobación del reparto, que el `require` de la línea 114 no deja llegar | la de M01 | `vertical_de_ensayo_test.go:240` (desde `:126`, `comprobarElReparto`): «la fila del tipo de ensayo sale en la referencia de la vertical y en ninguna otra», `Not equal` |
| S1 | una región de la referencia del repositorio distinta de la generada: `\| 1 mes \|` → `\| un mes \|` en la línea 50 de `skills/plazos-y-territorio/references/administrativo.md` | `'^(TestRegiones\|TestSkillsDelRepositorio)$' ./internal/skills/ ./internal/app/`; `make -C <copia> skills-check` | `TestSkillsDelRepositorio/skills` (`skills_test.go:106`: «plazos-y-territorio: references/administrativo.md: contenido-distinto») y `/sincronia-con-referencias-a-mano/consulta-juridica/desde-el-byte-cambiado` y `/desde-la-cabecera-retirada`; `make: *** [Makefile:137: skills-check] Error 1`. `TestRegiones` sigue en verde: desde T003 lee su propio árbol, no el repositorio (S1b) |
| S1b | la comparación ciega al contenido: `if true \|\| bytes.Equal(actual, contenido)` en `internal/skills/sincronia.go:916` | `'^TestRegiones$/^sincronia$' ./internal/skills/` | `TestRegiones/sincronia`, `regiones_test.go:276` (la deriva de la región distinta de la generada), `:282` y `:284` |
| S2 | un byte no declarado en una skill del producto: un espacio al final de la última línea de `skills/jurisprudencia/SKILL.md` | `'^TestSkillsDelProductoDesdeControl$' ./internal/evals/` | `TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`: «skills/jurisprudencia/SKILL.md difiere de lo que dan su control y su declaración» |
| F1 | otra fecha en un caso (`gates/supuestos.md`): el silencio con `da: vencimiento` en `data/plazos.yaml:54` | `'^TestCasosDePlazos$' ./internal/app/`; `'^TestEntregaDelHito$/^(h9-plazos-.*\|h36-plazos-de-hoy)$' ./internal/app/` | `TestCasosDePlazos/11-silencio`; `TestEntregaDelHito/h36-plazos-de-hoy`: `h36-plazos-de-hoy.txtar:32: stdout and caso-11.json differ` |
| F2 | el mensaje de un tipo desconocido sin « ni »: `texto += ", "` en `internal/core/plazos/tipos.go:820` | `'^TestEntregaDelHito$/^h36-plazos-de-hoy$' ./internal/app/` | `h36-plazos-de-hoy.txtar:42: no match for` la expresión del sobre de `error-tipo` |
| F3 | otra clase: `return schema.ClaseNoEncontrado` en `internal/core/plazos/errores.go:35` | la de F2 | `h36-plazos-de-hoy.txtar:41: unexpected command failure` (el `exec` que espera el código 2 recibe otro) |
| F4 | el mensaje sin el tipo pedido: `argumentos("el tipo no es %s", enumeracion(codigos))` en `internal/core/plazos/errores.go:75` | la de F2 | `h36-plazos-de-hoy.txtar:42: no match for` la expresión del sobre de `error-tipo` |
| F5 | los códigos del mensaje de un tipo desconocido en otro orden: `codigos[1], codigos[2] = codigos[2], codigos[1]` antes del `return` de `tipoDesconocido`, `internal/core/plazos/errores.go:75`. Corrida en la revisión final, en el árbol de trabajo y deshecha después (`git diff --stat -- internal/core/` vacío) | `'^TestCasosDePlazos$' ./internal/app/` | `TestCasosDePlazos/tipo-que-no-existe`, `plazos_casos_test.go:312`: lo esperado, derivado de `data/plazos.yaml`, es «el tipo "inventado" no es alzada, reposicion, contencioso, acceso-resolver, acceso-reclamar, silencio ni generico», el mensaje del `error-tipo.json` de `main`; lo dado, con `contencioso` y `reposicion` cambiados. La expresión de `h36-plazos-de-hoy.txtar:42` lo acepta |

### C2 · El molde, en `TestVerticalDeEnsayo` crecida (FR-002, FR-005, FR-006, FR-020; US2; SC-002)

Los cuatro mutantes de US2.4. Orden de todos: `'^TestVerticalDeEnsayo$' ./internal/skills/`.

| # | Mutante | En rojo |
|---|---|---|
| M01 | un tipo asignado a todas las referencias (C1) | `:114`, la fila 101 de `administrativo.md`; con M01b, `:240` |
| M02 | una fila asignada por su precepto: `if false && entrada.Norma != nil` en `paqueteDeLaFila`, `internal/skills/regiones.go:531` | `vertical_de_ensayo_test.go:114`: el mismo defecto de 101 líneas, que es la fila con norma de la vertical en `administrativo.md` |
| M02b | M02 con `maximoDeLineasSinIndice = 200` | `:245` «la región de la cobertura de la referencia de la vertical genera la fila con norma»; `:250` «la fila con norma sale en la referencia de la vertical y en ninguna otra, tampoco en administrativo.md», con `skills/plazos-y-territorio/references/administrativo.md` en lugar de `ensayo.md`; y `/sin-la-region`, `:172` |
| M03 | `filasSinRegion` con la regla vieja: `porIdentificador[entrada.Precepto.Norma].Vertical` en lugar de `paqueteDeLaFila(…)`, `internal/skills/regiones.go:903` | `TestVerticalDeEnsayo/sin-la-region`, `:172`: los defectos dados no tienen «plazos-y-territorio: data/cobertura.yaml: la fila de «procedimientos de ensayo que la ley común deja aparte» no la genera ninguna región cobertura_por_materia» |
| M04 | la línea de `normativa` sin el movimiento: `git -C <copia> checkout origin/main -- skills/normativa/SKILL.md skills/normativa/references/con_herramientas.md` | `:142` → `:273` (`comprobarElSitioDeOtroPaquete`): «normativa: SKILL.md: cuerpo de 11066 caracteres (máximo 11000)». El spec («Assumptions») dice 11 067; lo medido es 11 066, que también pasa del tope |

### C3 · Los defectos nuevos de la sincronía (FR-003, FR-004; US3; SC-003)

Verdes, en el árbol: `go test -count=1 -v -run
'^TestRegiones$/^defectos$/^(tipo-sin-paquete|tipo-de-dos-paquetes|computo-de-otro-paquete)$' ./internal/skills/` → los
tres en `PASS`.

| # | Mutante | Orden | En rojo |
|---|---|---|---|
| D1 | quitar la regla: `defectos = nil` antes de `if len(defectos) > 0` en `filasDeLosRecursosYPlazos`, `internal/skills/regiones.go:580` | `'^TestRegiones$' ./internal/skills/` | `TestRegiones/defectos/tipo-de-dos-paquetes` y `/tipo-sin-paquete`, cada uno en `regiones_test.go:434` (el defecto), `:444` («con defectos no hay ninguna referencia regenerada») y `:445` |
| D2 | contar los preceptos del cómputo: `tipo.PreceptosDelComputo` dentro del `slices.Concat` de `paquetesDelTipo`, `internal/skills/regiones.go:600` | la de D1 | `TestRegiones/defectos/computo-de-otro-paquete`, `regiones_test.go:509`: «plazos-y-territorio: data/plazos.yaml: el tipo prueba-de-otro-computo es de más de un paquete: administrativo, ensayo»; y `/tipo-sin-paquete` (`:434`, `:444`: «[plazos-y-territorio: administrativo.md]», `:445`) |
| D3 | los paquetes sin ordenar: `slices.Sort(paquetes)` quitado de `paquetesDelTipo`, `internal/skills/regiones.go`. Corrida en la revisión final, en el árbol de trabajo, como D4 y D5, y deshecha después | la de D1 | `TestRegiones/defectos/tipo-de-dos-paquetes`, `regiones_test.go:444`: «… es de más de un paquete: ensayo, administrativo»; su tipo lleva el precepto de ensayo delante de dos de la LPAC |
| D4 | los paquetes sin quitar los repetidos: `return paquetes` en lugar de `return slices.Compact(paquetes)` en `paquetesDelTipo` | la de D1 | cada tipo de lo administrativo con más de un precepto da un defecto, como «el tipo contencioso es de más de un paquete: administrativo, administrativo»: `TestRegiones/sincronia`, `/paquete-de-la-fila`, `/defectos/tipo-sin-paquete` y los demás en rojo |
| D5 | la fila con su precepto en una norma que no está, de un paquete: `if false && !esta` en `paqueteDeLaFila`, `internal/skills/regiones.go` | la de D1 | `TestRegiones/defectos/fila-con-su-precepto-en-una-norma-que-no-esta`, `regiones_test.go:444` (no da el defecto de la fila sin región, el de `main`), `:454` («[plazos-y-territorio: administrativo.md plazos-y-territorio: ensayo.md]») y `:455` |

### C4 · Los tests sin listas (FR-016; US4; SC-004)

| # | Mutante | Orden | En rojo |
|---|---|---|---|
| L00 | la lista literal de hoy devuelta a `TestCoberturaDelRepositorio`: `}), leidas, ` en lugar de `}), deLaBase(leidas), ` en `internal/skills/cobertura_test.go:398`, que compara todas las filas leídas con las nueve, como en `main` | `'^TestCoberturaDelRepositorio$' ./internal/skills/`; `-v '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/` | en la copia sin el paquete, `ok`; en `h45-paquete-de-mas`: `sincronía: 0`, `tests enteros: 1`, `subtests: 0`, en su salida de errores la orden de los tests enteros con `exit status 1` y `--- FAIL: TestCoberturaDelRepositorio`, y `h45-paquete-de-mas.txtar:41: no match for ^tests enteros: 0$` |

El mutante de cada test de la lista, con la columna «Rojo con» de contracts/tests-sin-listas.md. Con todos deshechos, en
la copia con un paquete más todos en verde (A3).

| # | Test | Mutante | Orden | En rojo |
|---|---|---|---|---|
| L01 | `TestLeerCobertura/repositorio` | un texto de una fila de la base: «la provincia, la comunidad» → «la provincia y la comunidad» en `data/cobertura.yaml:26` | `'^TestLeerCobertura$/^repositorio$' ./internal/skills/` | `TestLeerCobertura/repositorio`: «../../data/cobertura.yaml trae las cuatro filas de lo cubierto de la base de FR-001, en su orden» |
| L02 | `TestLeerCobertura/repositorio` | el orden: la fila de `normativa` detrás de la de `jurisprudencia` (`sed '18{N;h;d};21G'`) | la de L01 | el mismo mensaje |
| L03 | `TestLeerCobertura/repositorio` | la materia: `materia: materia-inventada` en `data/cobertura.yaml:24` | la de L01 | «la fila de lo cubierto de la capacidad "plazos" … lleva por materia el vertical de una norma de ../../data/normas.yaml» y el de L01 |
| L04 | `TestCoberturaDelRepositorio` | una fila del paquete quitada: fuera la del orden social (`data/cobertura.yaml:35-36`) | `'^TestCoberturaDelRepositorio$' ./internal/skills/` | `--- FAIL: TestCoberturaDelRepositorio` |
| L05 | `TestCoberturaDelRepositorio` | cambiada: «en el orden social» → «en el orden laboral» | la de L04 | `--- FAIL: TestCoberturaDelRepositorio` |
| L06 | `TestCoberturaDelRepositorio` | sin su norma: fuera `data/cobertura.yaml:39`, la norma de la fila de tráfico | la de L04 | `--- FAIL: TestCoberturaDelRepositorio` |
| L07 | `TestCoberturaDelRepositorio` | `LeerCobertura` sin `Norma`: el resultado de `ValidarDocumentoYAML` con `Norma = nil` en cada fila, `internal/skills/cobertura.go:127` | la de L04 | «../../data/cobertura.yaml trae, de las nueve entradas de contracts/datos-de-plazos.md §3, las de administrativo, en su orden» |
| L08 | `TestRegionDeLoCubierto/repositorio` | `lineasDeLoCubierto` sin una fila: `range cobertura.Cubierto[1:]`, `internal/skills/cubierto.go:78` | `'^TestRegionDeLoCubierto$' ./internal/skills/` | `TestRegionDeLoCubierto/repositorio` |
| L09 | `TestRegionDeLoCubierto/repositorio` | en otro orden: `range slices.Concat(cobertura.Cubierto[1:], cobertura.Cubierto[:1])` | la de L08 | `TestRegionDeLoCubierto/repositorio` |
| L10 | `TestRegionDeLoCubierto/repositorio` | con la materia dentro: `fila.Alcance + " (" + fila.Materia + ")\n"`, `cubierto.go:87` | la de L08 | `TestRegionDeLoCubierto/repositorio` |
| L11 | `TestRegionDeLoCubierto/sin-data-cobertura` | el defecto de los datos que faltan callado: `case false && errors.As(err, &queFalla) && queFalla.causa == nil` en `defectosDeLosDatosQueFaltan`, `internal/skills/regiones.go:862` | la de L08 | `TestRegionDeLoCubierto/sin-data-cobertura` |
| L12 | `TestRenderizarRecursosYPlazos/desde-data` | `filasDeLosTipos` con una fila menos: `range tipos[1:]`, `internal/skills/referencias.go:480` | `'^TestRenderizarRecursosYPlazos$/^desde-data$' ./internal/skills/` | `TestRenderizarRecursosYPlazos/desde-data` (sin el filtro, también `/cada-unidad-con-su-nombre`, `/unidad-sin-nombre-tal-cual` y `/bytes-exactos`) |
| L13 | `TestRenderizarRecursosYPlazos/desde-data` | una región sin regenerar: la de S1 | `'^TestRenderizarRecursosYPlazos$' ./internal/skills/` | `TestRenderizarRecursosYPlazos/desde-data` |
| L14 | `TestRegionDeLaCobertura` | el filtro de `filasDeLaCobertura` quitado: `if false && paqueteDeLaFila(…) != region.paquete`, `internal/skills/regiones.go:512` | `'^(TestRegiones\|TestVerticalDeEnsayo\|TestRegionDeLaCobertura)$' ./internal/skills/`; `-v '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/` | con los datos de hoy, `TestRegionDeLaCobertura` en verde, como dice el contrato: todas las filas son del único paquete y `n` derivado no cambia; lo ven `TestRegiones/paquete-de-la-fila` y `TestVerticalDeEnsayo`; y en la copia con un paquete más, `sincronía: 1`, `tests enteros: 1` (`--- FAIL: TestRegionDeLaCobertura`, `TestRegiones/paquete-de-la-fila`…) y `subtests: 1` |
| L15 | `TestRegionDeLaCobertura` | el motivo con norma sin sustituir: `marcadorDeLaNorma, marcadorDeLaNorma+nombre[:0]` en `internal/skills/regiones.go:389` (con `marcadorDeLaNorma, marcadorDeLaNorma` solo, el paquete no compila: `nombre` sin usar) | `'^TestRegionDeLaCobertura$' ./internal/skills/` | «la fila de "aplicación de los tributos…" es la de la plantilla con su motivo», la de «tráfico y seguridad vial» y «la referencia del repositorio es la que genera la sincronía» |
| L16 | `TestRegionDeLaCobertura` | una región sin regenerar: «económico-administrativas» → «económico administrativas» en la fila de la línea 39 de `skills/plazos-y-territorio/references/administrativo.md` | la de L15 | `--- FAIL: TestRegionDeLaCobertura` |
| L17 | `TestRegiones` | los de C3 y S1b | | C3, S1b |
| L18 | `TestVerticalDeEnsayo` | la fila de la vertical en otro sitio: la rotación de L09 | `'^TestVerticalDeEnsayo$' ./internal/skills/` | «la fila de lo cubierto de la vertical es la última línea de la región» y «la fila de la vertical va detrás de la que era la última de la región» |
| L19 | `TestCasosDePlazos/19-eval-c-tributos` | la fila sin `norma`: fuera `data/cobertura.yaml:30` | `'^TestCasosDePlazos$/^19-eval-c-tributos$' ./internal/app/` | «data/cobertura.yaml lleva una entrada, la del precepto DA 1ª.2.a de BOE-A-2015-10565 regida por BOE-A-2003-23186» |
| L20 | `TestCasosDePlazos/19-eval-c-tributos` | otro `aplicable`: `{bloque: a98, apartado: "98"}` | la de L19 | «el precepto que la hace aplicable» |
| L21 | `TestCasosDePlazos/19-eval-c-tributos` | otro precepto: `apartado: "DA 1ª.2.b"` en `data/cobertura.yaml:29` | la de L19 | el mensaje de L19 |
| L22 | `TestCasosDePlazos/19-eval-c-tributos` | repetida: las líneas 28 a 30 otra vez detrás de la 30 | la de L19 | el mensaje de L19 |
| L23 | `h32-paquetes` | una fila del generador cambiada: la de S1 (la skill empotrada en el binario de la copia) | `'^TestEntregaDelHito$/^(h32-paquetes\|h33-cobertura-instalada)$' ./internal/app/` | `h32-paquetes.txtar:37: no match for ^filas: 33 de 33$` |
| L24 | `h32-paquetes` | perdida: fuera la línea 50 de `administrativo.md` | la de L23 | la de L23 |
| L25 | `h33-cobertura-instalada` | una fila perdida: fuera la línea 39 de `administrativo.md` | la de L23 | `h33-cobertura-instalada.txtar:43: no match for ^cuadran: sí$`; también `h32-paquetes.txtar:37` |
| L26 | `h33-cobertura-instalada` | lo cubierto perdido: fuera la línea 44 de `skills/consulta-juridica/SKILL.md` (`- **Territorio**: …`) | `'^TestEntregaDelHito$/^h33-cobertura-instalada$' ./internal/app/` | `h33-cobertura-instalada.txtar:39: no match for` las cuatro líneas de la base tras la marca de lo cubierto |
| L27 | `h36-plazos-de-hoy` | F2, F3 y F4 de C1 | | C1 |
| L28 | `h41-4-skills-instaladas` | la descripción de hoy: DE1 de C7 | | C7 |
| L29 | `TestDefinicionDelJob/peor-caso` | `peorCaso` con otra cuenta: `Modos: []Modo{ModoOrden}` en `internal/evals/definicion.go:773` | `'^TestDefinicionDelJob$/^(peor-caso\|del-repositorio\|tope-con-las-evals-del-repositorio\|cuatro-evals-mas)$' ./internal/evals/` | `TestDefinicionDelJob/peor-caso` (`consulta-juridica`, `jurisprudencia`, `normativa`, `plazos-y-territorio`) y `/cuatro-evals-mas` |
| L30 | `TestDefinicionDelJob` | la concurrencia de hoy con cuatro evals más, y trece sesiones: T1 y T2 de C6 | | C6 |
| L31 | `TestPlanEnDosModos/del-repositorio` | el plan sin la sesión de red: `if false && p.PruebaDeRed && delModoOrden > 0`, `internal/evals/plan.go:240` | `'^(TestDefinicionDelJob\|TestPlanEnDosModos)$' ./internal/evals/` | `TestPlanEnDosModos/del-repositorio/plazos-y-territorio` y `/normativa`, y `/sondeo`, `/prueba-de-red`, `/sesiones`; `TestDefinicionDelJob/cuatro-evals-mas` |
| L32 | `TestPlanEnDosModos/del-repositorio` | sin un modo: `range p.modos()[:1]`, `internal/evals/plan.go:187` | `'^TestPlanEnDosModos$/^del-repositorio$' ./internal/evals/` | `/del-repositorio/plazos-y-territorio` y `/normativa`: «series por modo», «sesiones por modo», «sesiones por modo con la prueba de red», «una tanda por grupo, en su orden». Con `'^(TestDefinicionDelJob\|TestPlanEnDosModos)$'` entero, `/series` y `/sesiones` caen y otro subtest da `panic: runtime error: index out of range [23] with length 15` antes de llegar a `del-repositorio`; por eso la orden de la fila |
| L33 | `TestEvalsDelRepositorio/precepto-de-la-cobertura` | la 12 con el bloque `a999` en su `no_cubierto.precepto` (`evals/plazos-y-territorio/12-…yaml:15`, en la copia) | `'^TestEvalsDelRepositorio$/^precepto-de-la-cobertura$' ./internal/evals/` | «12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: no_cubierto.precepto BOE-A-2015-10565, bloque a999 no es el precepto de ninguna fila de data/cobertura.yaml» |
| L34 | `TestEvalsDelRepositorio/conjunto-plazos-y-territorio` | `LeerConjunto` que pierde un fichero: `range entradas[:len(entradas)-1]`, `internal/evals/conjunto.go:95` | `'^TestEvalsDelRepositorio$/^conjunto-plazos-y-territorio$' ./internal/evals/` | «../../evals/plazos-y-territorio tiene una eval por cada uno de sus *.yaml» |
| L35 | `TestEvalsDelRepositorio/ordenes-para-powershell` | los de C8 | | C8 |
| L36 | `TestUmbralesDePlazosYTerritorio` | quitar `if r.conVigencia`: `if false && r.conVigencia`, `internal/evals/umbrales.go:625` | `'^TestUmbralesDePlazosYTerritorio$' ./internal/evals/` | `/una-de-vigencia-sin-declarar`, `/una-de-cobertura-sin-declarar`, `/dos-de-cobertura-con-la-linea-del-plazo`… |
| L37 | `TestUmbralesDePlazosYTerritorio` | `lectura_sin_intentar` que decide para todo modelo: `Decide: true`, `umbrales.go:1164` | la de L36 | `/las-de-haiku`, `/ninguna-cuenta`, `/una-de-plazo-sin-la-fecha`… |
| L38 | `TestUmbralesDeNormativa` | `esperaUnaRedaccionFechada` cambiada: `return false && eval.RedaccionFechada != nil`, `internal/evals/informe.go:1766` | `'^TestUmbralesDeNormativa$' ./internal/evals/` | `/la-de-hoy-con-otra-fecha`, `/la-de-hoy-con-otra-norma`, `/lee-con-la-fecha`… |
| L39 | `TestUmbralesDeNormativa` | el umbral de duración fuera: `if false && e.ObjetivoDeDuracion > 0`, `umbrales.go:438` | la de L38 | `/lee-con-la-fecha`, `/redaccion-a-fecha-en-la-b`, `/sin-la-de-hoy`… |
| L40 | `TestUmbralesDeLasSkills` | `umbralesDelInforme` en otro orden: los conjuntos del juez delante de los de las respuestas, `umbrales.go:431-432` | `'^TestUmbralesDeLasSkills$' ./internal/evals/` | `/jurisprudencia` y `/normativa` |
| L41 | `TestUmbralesDeLasSkills` | `fuera_de_lo_grabado` que decide: `Decide: true` en su umbral, `umbrales.go:1149` | la de L40 | `/jurisprudencia`, `/consulta-juridica/con-las-sesiones-del-plan`… |
| L42 | `TestInformeEnDosModos/plazos-y-territorio` | una serie en otro orden: `range slices.Concat(p.Evals[1:], p.Evals[:1])`, `internal/evals/plan.go:188` | `'^TestInformeEnDosModos$/^plazos-y-territorio$' ./internal/evals/` | `TestInformeEnDosModos/plazos-y-territorio` |
| L43 | `TestInformeEnDosModos/plazos-y-territorio` | una serie que no decide: `Decide: !eval.Informativa && !strings.HasPrefix(eval.Fichero, "01-")`, `plan.go:210` | la de L42 | `TestInformeEnDosModos/plazos-y-territorio` |
| L44 | `TestMatrizDeLasPropuestas` | el formato de `matrizDeLaEjecucion`: `"; reutilizadas "` en `internal/evals/eleccion.go:387` | `'^TestMatrizDeLasPropuestas$' ./internal/evals/` | `--- FAIL: TestMatrizDeLasPropuestas` |
| L45 | `TestMatrizDeLasPropuestas` | la regla de elección: `votos[indice].anotar(motivo, "")`, sin la ruta que elige, `eleccion.go:175` | la de L44 | `--- FAIL: TestMatrizDeLasPropuestas` |
| L46 | `TestSkillsDelProductoDesdeControl` | los de C9 y S2 | | C9, C1 |
| L47 | `TestDescripcionesDelProducto`, `TestDescripcionesCortas` | DE3 de C7; y sus casos en memoria de T016 (`una_de_451`, `una_suma_de_1701`, `una_quinta_skill_sin_texto`, `una_de_451_runas`, `cuatro_de_450,_450,_400_y_401`, `una_que_empieza_por_<ejemplo>`), en `PASS` con `-v` | | C7 |

Mutantes que salieron en verde, que no cuentan como el rojo de su control y que se sustituyeron:

- `if true || r.conVigencia` (`umbrales.go:625`): `TestUmbralesDePlazosYTerritorio` en verde. Es equivalente para esa
  skill, que tiene la eval de vigencia; en las demás sí cae (`'^(TestUmbralesDeLasSkills|TestUmbralesDeNormativa)$'`:
  `TestUmbralesDeLasSkills/jurisprudencia`, `/consulta-juridica`, `/normativa`). El de la fila L36 es quitar el umbral.
- `return eval.RedaccionFechada != nil && !strings.HasPrefix(eval.Fichero, "0")` (`informe.go:1766`):
  `TestUmbralesDeNormativa` en verde. Es equivalente: las evals fechadas de `normativa` son la 22 y la 23. Se sustituyó
  por L38.
- `range cambios[:min(1, len(cambios))]` en `elegir` (`eleccion.go:137`): `TestMatrizDeLasPropuestas` en verde. No se
  averiguó por qué ningún caso de la matriz lo ve; se sustituyó por L45, que cambia la regla dentro de `eleccion`.

### C5 · Las grabaciones (FR-015; US7; SC-005)

Verdes, en el árbol: `go test -count=1 -run '^(TestGrabacionesSinSolape|TestManifiestoDeGrabaciones|TestDefinicionDelJob|TestPlanEnDosModos)$' ./internal/evals/` → `ok`.

| # | Mutante | Orden | En rojo |
|---|---|---|---|
| G1 | h36 tras una raíz de número mayor: el orden por la ruta, `return cmp.Compare(una.ruta, otra.ruta)`, en `internal/evals/grabaciones.go:142` | `'^(TestManifiestoDeGrabaciones\|TestGrabacionesSinSolape)$' ./internal/evals/` | `TestManifiestoDeGrabaciones/raices-de-los-hitos/orden-y-carpetas-que-no-entran` |
| G2 | `evidencias/h9/` dentro: en la copia, `evidencias/h9/boe.legislacion-consolidada/` con el `robots.txt` de h36 (h9 ya tiene `manifiesto.json`) | `'^TestGrabacionesSinSolape$' ./internal/evals/` | «../../evidencias/h9/boe.legislacion-consolidada: las carpetas de h9 no son grabaciones del BOE» |
| G3 | un fichero repetido entre dos raíces: en la copia, el `GET_…BOE-A-1889-4763_metadatos.json` de H5 (`testdata/evals/boe.legislacion-consolidada/`) copiado en `evidencias/h36/boe.legislacion-consolidada/` | la de G2 | «los conjuntos de grabaciones ../../testdata/evals/boe.legislacion-consolidada y ../../evidencias/h36/boe.legislacion-consolidada solo coinciden en GET_https_www.boe.es_robots.txt.json» |

La carpeta `evidencias/h99/` entra sin Go: en la copia, `evidencias/h99/manifiesto.json` (el de h36) y
`evidencias/h99/boe.legislacion-consolidada/` con el `robots.txt` de h36, y un test de un solo uso,
`internal/evals/zz_union_t018_test.go`, que solo vivió en la copia y escribe `UnionDeGrabaciones()` con `t.Logf`. Sin
h99: las raíces de H4, H5 y H9 y `../../evidencias/h36/boe.legislacion-consolidada`, 4. Con h99: las mismas y
`../../evidencias/h99/boe.legislacion-consolidada`, la quinta y última; `TestGrabacionesSinSolape` y
`TestManifiestoDeGrabaciones` en `PASS`. Los dos ficheros se borraron después.

### C6 · El tope del trabajo de evals (FR-021, FR-022; US8; SC-006)

`grep -n -A1 'skill: plazos-y-territorio' .github/workflows/evals.yml` → `259: - skill: plazos-y-territorio`,
`260: concurrencia: 3`. Con cuatro evals más en una copia, `TestDefinicionDelJob/cuatro-evals-mas` en verde (dentro de
`ok` en el árbol y en la copia). Cada mutante cambia a la vez `evals.yml:260` y `ajustesDelTrabajo` de
`internal/evals/definicion_test.go:264`, para que la comparación de la definición no lo tape.

| # | Mutante | Orden | En rojo |
|---|---|---|---|
| T1 | `plazos-y-territorio` a dos | `'^(TestDefinicionDelJob\|TestPlanEnDosModos)$' ./internal/evals/` | `TestDefinicionDelJob/cuatro-evals-mas`: «"5h52m0s" is not greater than or equal to "7h34m13s"», «con cuatro evals más y la definición del repositorio, el tope cubre el peor caso de plazos-y-territorio, 27253 s = 485 s + 900 s + (⌈91 / 2⌉ + ⌈90 / 2⌉ + ⌈6 / 2⌉) × (22 s + 240 s + 10 s) + 60 s + 6 × (35 s + 5 s)»; también `/tope-con-las-evals-del-repositorio` («premisa: la concurrencia de plazos-y-territorio») y `/sinteticas/decidir-sin-la-ejecucion` |
| T2 | trece sesiones a la vez: `plazos-y-territorio` a cuatro | `'^TestDefinicionDelJob$/^(sinteticas/la-del-contrato\|tope-con-las-evals-del-repositorio\|del-repositorio)$' ./internal/evals/` | `TestDefinicionDelJob/del-repositorio`: «jobs.evals.strategy.matrix: admite 13 sesiones a la vez entre sus 4 combinaciones, y lo esperado es como mucho 12»; con el test entero caen además `/tope-con-las-evals-del-repositorio` y los casos de `/sinteticas`, que parten de la definición del repositorio |
| T3 | `peorCaso` con otra cuenta | L29 de C4 | |
| T4 | el plan sin la sesión de red; sin un modo | L31 y L32 de C4 | |

### C7 · La descripción (FR-023; US9; SC-007)

`tr -d '\n' < laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt | wc -m` → `429`. En el árbol,
`go test -count=1 -run '^(TestSkillsDelProductoDesdeControl|TestDescripcionesCortas|TestDescripcionesDelProducto)$'
./internal/evals/` → `ok`; en la copia, `-v -run '^TestEntregaDelHito$/^(h41-4-skills-instaladas|h45-skills-instaladas)$'
./internal/app/` → los dos en `PASS`.

| # | Mutante | Orden | En rojo |
|---|---|---|---|
| DE1 | la de hoy, que nombra la LPAC, en el `SKILL.md`: `git -C <copia> checkout origin/main -- skills/plazos-y-territorio/SKILL.md` | `'^(TestSkillsDelProductoDesdeControl\|TestDescripcionesDelProducto\|TestDescripcionesCortas)$' ./internal/evals/`; `'^TestEntregaDelHito$/^(h41-4-skills-instaladas\|h45-skills-instaladas)$' ./internal/app/` | `TestSkillsDelProductoDesdeControl/plazos-y-territorio/frontmatter` y `/desde_control` («la description de skills/plazos-y-territorio/SKILL.md es la de laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt»); `TestDescripcionesDelProducto/la_rama`, `/un_carácter_añadido`, `/una_quinta_skill_sin_texto`…; `h45-skills-instaladas.txtar:30` y `h41-4-skills-instaladas.txtar:34`: `no match for` la `description` de FR-023 en `destino/plazos-y-territorio/SKILL.md` |
| DE2 | la de hoy en la variante: `git -C <copia> checkout origin/main -- laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt` | `'^(TestDescripcionesCortas\|TestSkillsDelProductoDesdeControl)$' ./internal/evals/` | `TestDescripcionesCortas/la_variante_del_repositorio`; `TestSkillsDelProductoDesdeControl/plazos-y-territorio/frontmatter` y `/desde_control` |
| DE3 | por encima de los topes: 22 caracteres más en la variante (451) | `-v '^TestDescripcionesCortas$/^la_variante_del_repositorio$' ./internal/evals/` | «la descripción de plazos-y-territorio tiene 451 caracteres (máximo 450)» y «las descripciones suman 1714 caracteres (máximo 1700)» |

El juez de las descripciones (SC-007, SC-010) no se midió: lo mide el job de evals del cierre, con sesiones con modelo
(A4).

### C8 · Las órdenes para PowerShell (FR-011, FR-012; US5; SC-008)

En el árbol: `tail -n 4 skills/normativa/references/con_herramientas.md` → las cuatro líneas de §C.2; `grep -c 'En
PowerShell 5.1:' skills/normativa/SKILL.md` → `0`; `go test -count=1 -run
'^(TestEvalsDelRepositorio|TestControlesSobreLasReferenciasAMano)$' ./internal/evals/` → `ok`. Con `-v`, en la copia y
en el árbol, en `PASS`
`TestEvalsDelRepositorio/ordenes-para-powershell` (el caso 1 de §C.3 sobre el repositorio: sin defecto) y los casos en
memoria `orden-del-skill-md-con-su-forma-en-una-referencia`, `forma-en-una-referencia-con-otros-bloques`,
`forma-en-ninguna-referencia`, `orden-de-una-referencia-con-su-forma-en-el-skill-md`, `orden-con-su-forma-en-otra-referencia`
y `sin-las-ordenes-de-articulos`.

Los cinco de §C.3, sobre los ficheros de la copia. Orden de todos: `'^TestEvalsDelRepositorio$/^ordenes-para-powershell$'
./internal/evals/`.

| # | Mutante | En rojo |
|---|---|---|
| P1 | esa forma con otros bloques: `a140` → `a141` en la línea 42 de `skills/normativa/references/con_herramientas.md` | «normativa: SKILL.md: línea 36, «kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check BOE-A-1978-31229 a140 --json»: no la sigue su forma para PowerShell…» |
| P2 | esa forma en ninguna referencia: fuera las líneas 39 a 43 de `con_herramientas.md` | el mismo defecto |
| P3 | una orden de una referencia con su forma solo en el `SKILL.md`: la forma de `boe articulos` fuera de `references/varios_bloques.md` (su párrafo y su valla) y al final de `skills/normativa/SKILL.md` | «normativa: references/varios_bloques.md: línea 14, «kitlegal boe articulos <norma> <bloques>... --json && kitlegal graph check <norma> <bloques>... --json»: no la sigue su forma…» |
| P4 | una orden de una referencia con su forma en otra referencia: la misma forma, al final de `con_herramientas.md` | el mismo defecto de `varios_bloques.md`, línea 14 |
| P5 | la orden de `boe articulos` quitada: fuera de `varios_bloques.md` las dos líneas `kitlegal boe articulos <norma> …` | «normativa: ninguna orden de Bash lee con «kitlegal boe articulos» y comprueba detrás con kitlegal graph check», su defecto de verbo, como hoy |

Y la regla nueva del control, en `internal/evals/conjunto_test.go` de la copia. Orden:
`'^(TestEvalsDelRepositorio|TestControlesSobreLasReferenciasAMano)$' ./internal/evals/`.

| # | Mutante | En rojo |
|---|---|---|
| P6 | la regla de hoy: `defectosDeLasOrdenes(delSkillMd, nil)` en `:4792`, sin las órdenes de las referencias | `TestEvalsDelRepositorio/ordenes-para-powershell` (el repositorio, con la forma movida); `TestControlesSobreLasReferenciasAMano/orden-del-skill-md-con-su-forma-en-una-referencia`, y con ella `/forma-en-ninguna-referencia`, `/orden-de-una-referencia-con-su-forma-en-el-skill-md`, `/ordenes-de-articulos-en-skill-md` y `/forma-en-una-referencia-con-otros-bloques`, cuyos defectos esperados son exactos |
| P7 | las órdenes de una referencia, cubiertas por una forma del `SKILL.md`: `defectosDeLasOrdenes(ordenes, ordenesDeLectura(skill.skillMd.texto))` en `:4786` | `TestControlesSobreLasReferenciasAMano/orden-de-una-referencia-con-su-forma-en-el-skill-md` |

### C9 · El laboratorio sin las referencias de un paquete (FR-014; US6; SC-009)

Verde, en el árbol: `TestSkillsDelProductoDesdeControl` → `ok` (C7), y con `-v` en `PASS`
`/referencias-de-un-paquete/de-un-vertical-que-la-base-no-tiene`, `/de-una-materia-que-no-es-un-vertical` y
`/un-byte-en-una-de-un-vertical-que-la-base-tiene`. Orden de todos:
`'^TestSkillsDelProductoDesdeControl$' ./internal/evals/`.

| # | Cambio en la copia | Esperado | Visto |
|---|---|---|---|
| LB1 | `skills/normativa/references/administrativo.md`, copia de la de `plazos-y-territorio`: la referencia de un vertical de `data/normas.yaml` que la base de `normativa` no tiene | verde | `ok` |
| LB2 | la misma, como `materia-inventada.md`, un nombre que no es un vertical | rojo | `TestSkillsDelProductoDesdeControl/normativa/desde_control`: «skills/normativa/references/materia-inventada.md sobra: no está en lo que dan su control y su declaración» (y el subtest en memoria `/de-un-vertical-que-la-base-no-tiene`, que lee el mismo árbol) |
| LB3 | una referencia que la base ya tiene, cambiada: un espacio al final de la línea 1 de `skills/plazos-y-territorio/references/administrativo.md` | rojo | `TestSkillsDelProductoDesdeControl/plazos-y-territorio/desde_control`: «skills/plazos-y-territorio/references/administrativo.md difiere de lo que dan plazos-y-territorio-sin-condicion y su declaración» |
| LB4 | la regla, con cualquier nombre: `vertical != ""` en lugar de `slices.Contains(verticales, vertical)`, `internal/evals/laboratorio_familias_test.go:1239` | rojo | `/referencias-de-un-paquete/de-una-materia-que-no-es-un-vertical`: «"[]" should have 1 item(s), but has 0», «una referencia que no es de un vertical se declara» |
| LB5 | la regla, aunque la base tenga el fichero: `case esDeUnPaqueteQueLaBaseNoTiene(…)` sin `!loEspera &&` (`:1212`) y `false && laTiene` (`:1232`) | rojo | `/referencias-de-un-paquete/un-byte-en-una-de-un-vertical-que-la-base-tiene`: «"[]" should have 1 item(s), but has 0», «una referencia de un paquete que la base tiene se compara» |

### C10 · Las evals (FR-018, FR-025; SC-010)

- FR-025, en el árbol: `git diff --stat main -- internal/evals/conjunto.go evals/ | wc -l` → `0`: ninguna eval cambia.
- FR-018: el mismo diff, vacío también para `internal/evals/conjunto.go`, el código de producción del conjunto de evals,
  donde viven las reglas de `normativa` (`minimoDeEvals = 10`, `maximoDeEvals = 23`, `positivasDelConjunto = 10`,
  `conjunto.go:232-234`). Mutante E1: en la copia, `evals/normativa/99-copia-de-la-01.yaml`, copia de la 01, con
  `'^TestEvalsDelRepositorio$' ./internal/evals/` → `TestEvalsDelRepositorio/conjunto` (`conjunto_test.go:2644`): «tamaño:
  hay 24 evals y el conjunto lleva entre 10 y 23…» y «positivas: hay 11 evals positivas (activa: true) que deciden y el
  conjunto lleva exactamente 10…»; y `/conjunto-de-cada-carpeta`. Borrado después.
- SC-010 (ninguna clase que decide por debajo de su umbral, en las cuatro skills): no medido. Lo mide el job de evals del
  cierre con sesiones con modelo, que este paso no abre.

## El quickstart

| Escenario | Dónde | Medido | Resultado |
|---|---|---|---|
| 1 · Lo de hoy, igual | árbol | `make skills-check`; `git diff --stat main -- skills/`; `git diff main -- skills/plazos-y-territorio/references/ skills/normativa/references/normas.md`; los dos `go test` | `ok` ×3; el diff, solo en los tres ficheros esperados; vacío; `ok` y `ok` (A2, C1) |
| 2 · El molde | árbol | `go test -count=1 -v -run '^TestVerticalDeEnsayo$' ./internal/skills/` | `PASS`, con `mutante` y `sin-la-region` (A1) |
| 3 · Los defectos nuevos | árbol | la orden del quickstart | los tres subtests en `PASS` (C3) |
| 4 · Un paquete más | copia | `go test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/` | `PASS` (7.65 s). En el árbol el guion aún no está activado: lo activa el workflow al final |
| 5 · `normativa` hace sitio | árbol | `tail -n 4`, `grep -c`, `go test` | las cuatro líneas de §C.2; `0`; `ok` (C8) |
| 6 · El laboratorio y la descripción | árbol y copia | `wc -m`; `go test` de los tres tests; los dos guiones, en la copia | `429`; `ok`; `h41-4-skills-instaladas` y `h45-skills-instaladas` en `PASS` (C7) |
| 7 · Las grabaciones y el trabajo de evals | árbol | `go test` de los cuatro tests; `grep -n -A1` | `ok`; `concurrencia: 3` (C5, C6) |
| 8 · Todo | copia y árbol | `make ci` | en la copia, con los dos guiones de H45 activados: `0 issues.` y `ci: todos los controles en verde` (4 min 22 s); en el árbol, al cerrar T018 (abajo) |

No medido en esta sesión: A4 y SC-010 (las evals del cierre) y el juez de las descripciones (SC-007).

## `make ci` (SC-011)

- En la copia, con la suite de H45 activada y sin ningún mutante: `make -C /tmp/claude/t018 ci` → `0 issues.` y `ci:
  todos los controles en verde`.
- En el árbol de trabajo, al cerrar T018, con este registro escrito: `make ci` → `0 issues.` y `ci: todos los controles
  en verde` (4 min 8 s).

Cada fila `ci:` de «Controles de umbral» de plan.md tiene así su mutante: SC-001 y FR-008 (S1, S1b, S2), SC-002 (M01 a
M04), SC-003 (D1, D2), SC-004 (L00 y L01 a L47), SC-005 (G1 a G3), SC-006 (T1, T2, L29, L31, L32), SC-007 (DE1 a DE3),
SC-008 (P1 a P7), SC-009 (LB2 a LB5), FR-018 (E1), FR-026 (F1 a F4, L19 a L22) y SC-011 (`make ci`). Ningún control quedó
en verde con su mutante: los tres mutantes en verde de C4 son equivalentes o no los ejercita el test, y se sustituyeron.
