---

description: "Tareas de H45.1 · Los tests del molde, sin el estado de hoy del repositorio (ADR 0062)"
---

# Tasks: H45.1 · Los tests del molde, sin el estado de hoy del repositorio (ADR 0062)

**Input**: `specs/043-h45-1-los-tests-del/` — spec.md, plan.md, research.md (P1, D1-D9, V1-V13), data-model.md,
contracts/ (tests-genericos.md, copia-con-un-paquete-mas.md), quickstart.md (Q1-Q8).

**Tests**: cada tarea lleva su test y el cambio mínimo que lo pone en verde, y deja `make ci` en verde al terminar. Aquí
el cambio *es* de tests: los cuatro tests de T001-T004 son a la vez lo que se cambia y su comprobación, y cada tarea ve
su test en rojo con el mutante del plan antes de darla por hecha. **Aceptación e2e: no aplica** (plan, Complexity
Tracking; research D6): no hay tarea `[aceptacion]`; la aceptación del hito es la copia `h45-paquete-de-mas` con los
cuatro casos, que cambia la tarea `[datos]` T005, y `make ci` en verde con los datos de `main`.

**Reglas que valen para todas las tareas**: sin red (salvo la de las herramientas de Go), sin `KITLEGAL_RECORD`, sin
sesiones con modelo; nada se escribe en el directorio `evidencias` ni en `data/`, `evals/`, `skills/` o `.github/`
(FR-014, US1.3); ningún `.go` que no termine en `_test.go` cambia; ningún `//nolint` ni `t.Skip` nuevo; ninguna
importación nueva en los ficheros de test (plan, Constitution Check). Cada mutante se aplica en una copia del módulo en
`$TMPDIR` (con el cambio de la tarea, aún sin commit, copiado encima de un `git clone --no-hardlinks` del repositorio),
nunca en el árbol de trabajo (research D7). Ninguna tarea toca `internal/app/testdata/script/h32-paquetes.txtar` ni
`h33-cobertura-instalada.txtar` (Fuera de alcance).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: puede ir en paralelo con las demás [P] (ficheros distintos, sin dependencias entre ellas). El bucle del
  workflow las ejecuta igualmente una a una, en orden.
- **[USn]**: la historia del spec a la que sirve.
- **[datos]**: toca ficheros bajo `testdata/` y no mezcla otro trabajo (solo T005).

---

## Phase 1: Setup

Sin tareas: el hito no crea paquetes, ficheros de Go, dependencias ni objetivos de `make` (plan, Structure Decision), y
no hay suite de aceptación que escribir («Aceptación e2e: no aplica»).

## Phase 2: Foundational

Sin tareas: los cuatro tests son independientes entre sí y, con los datos de `main`, cada uno comprueba lo mismo que hoy,
así que cada cambio deja en verde el guion `h45-paquete-de-mas` de hoy (research P1, D6). Lo único que bloquea es T005,
que solo pasa con T001-T004 hechos.

---

## Phase 3: User Story 2 — la cobertura no pone en rojo un paquete que estrecha una fila o le da norma (Priority: P1) 🎯 MVP

**Goal**: `TestCoberturaDelRepositorio` deja de escribir el texto de cada procedimiento y de filtrar lo esperado con una
norma escrita a mano.

**Independent Test**: en el árbol, `go test -count=1 -v -run '^TestCoberturaDelRepositorio$' ./internal/skills/` en verde
(Q7); con L06 y L07 de H45, en rojo (Q8).

- [X] T001 [P] [US2] `TestCoberturaDelRepositorio` genérico en `internal/skills/cobertura_test.go` (FR-001, FR-002, FR-003, FR-004; US2.1-US2.5; SC-004, SC-007; contracts/tests-genericos.md §1; data-model §1; research D1): `type tupla struct{ delPrecepto, apartado, queLaRige string }` local al test, sin campo de procedimiento; `deLaBase` con una tupla por fila de `LeerCobertura(...).ReguladosAparte`, en su orden, solo si es de `paqueteDeLaBase` por el filtro escrito en el test (el `vertical` de la norma que la rige en `normasDelRepositorioPorIdentificador`; sin norma o sin `vertical`, el de la norma de su precepto), sin `paqueteDeLaFila` ni `filasDeLaCobertura`; `nueve := []tupla{…}` las nueve tuplas de data-model §1 a mano y en su orden con `const lpac = "BOE-A-2015-10565"`; la comprobación por subsecuencia del contrato (`casadas`, `assert.Equal(t, len(deLaBase), casadas, …)`) con el mensaje de §1 («…: de sus filas de administrativo, `<leídas>`, solo las `<n>` primeras son, en su orden, de las nueve tuplas de contracts/datos-de-plazos.md §3 de H9, `<nueve>`»), y el comentario del test con la regla y con que lo que falta o cambia de texto lo ve `h32-paquetes` (FR-004). Ninguna de las nueve cadenas de procedimiento de hoy (`cobertura_test.go:384-397`) queda en el cuerpo del test (comprobar con `grep -n`). Ver en rojo, en una copia del módulo en `$TMPDIR` con el cambio, L06 (`sed -i 39d data/cobertura.yaml` → «solo las 5 primeras») y L07 (`LeerCobertura` con `Norma = nil`, la orden de Q8 → `expected: 9`, `actual: 0`); `go test -count=1 -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/` sigue en verde con el guion de hoy; `make ci` en verde.

**Checkpoint**: US2 en verde en el árbol y en rojo con L06 y L07.

---

## Phase 4: User Story 3 — la línea del paso 2 de `normativa` sale de la skill que se comprueba (Priority: P1)

**Goal**: los dos casos de `mutantesDeLasOrdenes` que esperan el defecto en la orden del paso 2 la esperan en la línea en
que está en el `SKILL.md`, no en la 36 escrita.

**Independent Test**: `go test -count=1 -v -run '^TestControlesSobreLasReferenciasAMano$/^(forma-en-una-referencia-con-otros-bloques|forma-en-ninguna-referencia)$' ./internal/evals/` en verde con el defecto en la línea 36 (Q6); con P6 y con «otra línea», en rojo (Q8).

- [X] T002 [P] [US3] La línea del paso 2, contada en `internal/evals/conjunto_test.go` (FR-005; US1.4, US3.1-US3.3; SC-004; contracts/tests-genericos.md §2; data-model §2; research D2): `func mutantesDeLasOrdenes(t *testing.T) []mutanteDeUnControl` con `t.Helper()`, y su llamada en `TestControlesSobreLasReferenciasAMano` pasa `t`; el defecto esperado de `forma-en-una-referencia-con-otros-bloques` y `forma-en-ninguna-referencia` pasa a `fmt.Sprintf("normativa: SKILL.md: línea %d, «%s»%s", lineaDeLaOrdenDelPasoDos(t), bashDelPasoDos, sinSuForma)`; nueva `func lineaDeLaOrdenDelPasoDos(t *testing.T) int` que lee `filepath.Join(raizDeLasSkills, directorioDeLasSkills, skillNormativa, ficheroDeLaSkill)` con `contenidoDelFichero`, recorre `strings.Lines` con un contador propio desde 1 y guarda cada línea cuyo `strings.TrimSpace` es `bashDelPasoDos`, con `require.Len(t, lineas, 1, "premisa: la orden de Bash del paso 2 está en una sola línea del SKILL.md de normativa")`, sin `tramosDeCodigo` ni `ordenesDeLectura`. Ningún otro mutante cambia y ningún número de línea del `SKILL.md` queda escrito para esos dos casos. Comprobar que con `main` el defecto esperado es el de hoy byte a byte, con «línea 36» (Q6, con `-v`); ver en rojo los dos casos, en una copia del módulo en `$TMPDIR` con el cambio, con P6 (`defectosDeLasOrdenes(delSkillMd, nil)`) y con «otra línea» (`tramoDeCodigo{linea: numero + 1, …}` en `tramosDeCodigo`), con las órdenes de Q8; `h45-paquete-de-mas` de hoy sigue en verde; `make ci` en verde.

**Checkpoint**: US3 en verde en el árbol, con el mensaje de hoy, y en rojo con P6 y con otra línea.

---

## Phase 5: User Story 4 — el hueco del trabajo de evals se mide con 16 evals fijas (Priority: P1)

**Goal**: `cuatro-evals-mas` mide una carpeta de `plazos-y-territorio` con exactamente 16 evals de modo, con el tope de
la definición leída.

**Independent Test**: `go test -count=1 -v -run '^TestDefinicionDelJob$/^cuatro-evals-mas$' ./internal/evals/` en verde
(Q7); con T1 y con L29, en rojo; con ocho evals de más, `tope-con-las-evals-del-repositorio` en rojo y este en verde (Q8).

- [X] T003 [P] [US4] `probarCuatroEvalsMas` con una carpeta de 16 en `internal/evals/definicion_test.go` (FR-006, FR-007, FR-008; US4.1-US4.4; SC-004, SC-006; contracts/tests-genericos.md §3; data-model §3; research D3): el subtest conserva su nombre `cuatro-evals-mas`; constantes `skill = "plazos-y-territorio"`, `evalsDeLaCarpeta = 16` y `deDosEnDos = 2`, sin `topeEnMinutos`, `evalsDeMas`, `primeraCopia` ni la premisa del tope de 352; la definición con `leerDefinicionDelJob(rutaDeLaDefinicionDelJob)` y `tope := time.Duration(leida.TopeEnMinutos) * time.Minute`; la copia de las evals con `os.CopyFS` en `t.TempDir()` y la eval copiada, la primera de modo que decide de `LeerConjunto` (`!SinBinarioNiServidor && !Informativa`), como hoy; en la carpeta de la skill de la copia, abierta con `os.OpenRoot`, `carpeta.Remove(eval.Fichero)` de cada eval de `LeerConjunto` y `carpeta.WriteFile(fmt.Sprintf("%02d-copia-%s", numero+1, descripcion), contenido, 0o600)` para `numero` de 0 a 15; `require.Len(t, conLasCopias.Evals, evalsDeLaCarpeta, …)`; `peorCaso` comparado con `peorCasoEsperado` como hoy; `assert.GreaterOrEqual(t, tope, peor.Duracion(), …)` a la concurrencia del repositorio y `assert.Less(t, tope, peorADos.Duracion(), …)` con la skill a dos, con los dos mensajes de §3; el comentario de la función dice lo que mide y que las del repositorio las exige `probarElTopeConLasEvalsDelRepositorio`, que no cambia. Ninguna duración literal queda en el subtest (comprobar con `grep -n`). Ver en rojo, en una copia del módulo en `$TMPDIR` con el cambio, T1 («"5h52m0s" is not greater than or equal to …») y L29, y con ocho copias de más en `evals/plazos-y-territorio/` de esa copia, `tope-con-las-evals-del-repositorio` en rojo y `cuatro-evals-mas` en verde, con las órdenes de Q8; `h45-paquete-de-mas` de hoy sigue en verde; `make ci` en verde.

**Checkpoint**: US4 en verde en el árbol, en rojo con T1 y L29, y FR-008 sin cambiar.

---

## Phase 6: User Story 5 — la eval de lo no declarado no se mide como una de lo regulado aparte (Priority: P1)

**Goal**: `llevaNoCubierto` deja fuera las evals con `no_cubierto.fuera_de_lo_declarado`.

**Independent Test**: `go test -count=1 -v -run '^TestEvalsDelRepositorio$/^precepto-de-la-cobertura$' ./internal/evals/`
en verde (Q7); con L33, en rojo con su mensaje de hoy; con las tres de lo regulado aparte pasadas a lo no declarado, en
rojo (Q8).

- [X] T004 [US5] `llevaNoCubierto` sin lo no declarado en `internal/evals/conjunto_test.go` (FR-009; US5.1-US5.3; SC-004; contracts/tests-genericos.md §4; data-model §4; research D4): `return eval.NoCubierto != NoCubiertoEsperado{} && !eval.NoCubierto.FueraDeLoDeclarado`, con un comentario que dice que deja fuera las de lo no declarado, que no llevan precepto; en `probarLosPreceptosDeLaCobertura`, el mensaje del `require.NotEmpty` pasa a «%s tiene al menos una eval con no_cubierto de lo regulado aparte»; `preceptosFueraDeLaCobertura` y el mutante `a999` no cambian. Ver en rojo, en una copia del módulo en `$TMPDIR` con el cambio, L33 (bloque `a999` en la 12, con el mensaje de hoy «12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml: no_cubierto.precepto BOE-A-2015-10565, bloque a999 no es el precepto de ninguna fila de data/cobertura.yaml») y las 07, 08 y 12 pasadas a `fuera_de_lo_declarado: true` (`Should NOT be empty`), con las órdenes de Q8; `h45-paquete-de-mas` de hoy sigue en verde; `make ci` en verde. Va detrás de T002 porque toca el mismo fichero.

**Checkpoint**: los cuatro tests, genéricos; con los datos de `main`, todo sale igual (US1).

---

## Phase 7: User Story 6 — la copia con un paquete más ejerce los cuatro casos (Priority: P1)

**Goal**: el guion `h45-paquete-de-mas` y su copia congelada de H45 montan un paquete que hace lo que hará H42 y corren
los dos tests que faltaban en su lista.

**Independent Test**: `go test -count=1 -v -run '^TestEntregaDelHito$/^h45-paquete-de-mas$' ./internal/app/` en verde con
`copia: hecha`, `sincronía: 0`, `orden del paso 2 de normativa, líneas más abajo: 1`, `tests enteros: 0` y
`subtests: 0` (Q4); `cmp` de los dos guiones sin diferencias (Q5).

- [X] T005 [datos] [US6] La copia con un paquete más, en `internal/app/testdata/script/h45-paquete-de-mas.txtar` y `specs/042-h45-el-molde-admite/aceptacion/paquete-de-mas.txtar`, el mismo fichero byte a byte (FR-010, FR-011, FR-012, FR-013; US6.1-US6.5; SC-003; contracts/copia-con-un-paquete-mas.md §1-§5; data-model §5; research D5): el comentario de cabecera de §1 (lo que el paquete añade y las dos filas de la base que cambia; la lista de FR-016 de H45 con FR-011 de H45.1; «Cubre:» con «US6 · FR-010 a FR-013 · SC-003 de H45.1»); en el programa `paquete-de-mas.go` del archivo, los filtros `filtroDeLosTestsEnteros` (gana `TestControlesSobreLasReferenciasAMano`, pierde `TestDefinicionDelJob`) y `filtroDeLosSubtests` (gana `TestDefinicionDelJob` en el primer nivel y `cuatro-evals-mas` en el segundo, de modo que no corren `del-repositorio`, `peor-caso` ni `tope-con-las-evals-del-repositorio`) literales de §2; `cambiosDeLaBase` (estrecha «actuaciones y procedimientos en materia de extranjería y asilo» a «…, salvo los del paquete de prueba» y da a «procedimientos sancionadores en el orden social» la norma `BOE-A-2015-11722` con `aplicable: {bloque: a1, apartado: "1"}`) y `cambiarUnaVez`, que exige cada texto viejo una sola vez (`strings.Count`) y si no devuelve `fmt.Errorf("%s no lleva una sola vez %q", fichero, viejo)`; en `instalar`, tras `anadirEntradas`, `cambiarUnaVez` sobre `data/cobertura.yaml` y `paquete/prueba-de-normativa.md` en `skills/normativa/references/prueba.md`, cada llamada en una sola línea con el texto literal de §2, la primera eval de `evals/plazos-y-territorio/` con los prefijos `96-` a `99-` y `paquete/fuera-de-lo-declarado.yaml` como `95-fuera-de-lo-declarado.yaml`; `const ordenDelPasoDos` y `lineaDeLaOrden(fichero string) (int, error)`, que tras la sincronía imprime `orden del paso 2 de normativa, líneas más abajo: <copia − módulo>` (error con 0 o más de 1 línea → código 1); los dos ficheros nuevos del archivo de §3, `paquete/prueba-de-normativa.md` y `paquete/fuera-de-lo-declarado.yaml`, con su texto literal; y las comprobaciones de §4 tras las de hoy (los `grep` de la fila estrechada en `administrativo.md`, de la del orden social en `prueba.md` y su `! grep` en `administrativo.md`, el `grep` de la línea del paquete en `repo/skills/normativa/SKILL.md`, el `stdout` de `líneas más abajo: 1` y el `cmp` de los dos guiones con `$KITLEGAL_SKILLS/..`). Lo demás del guion (la copia del módulo, `anadirEntradas`, la sincronía, `ejecutar`, `principio`, `maxSalida` y los cuatro `paquete/*.yaml`/`prueba.md` de hoy) no cambia. Escribir el fichero una vez y copiarlo al otro (`cp`), y comprobar con `cmp` que son iguales (Q5); ver el guion en verde con las cinco líneas (Q4); ver en rojo, en una copia del módulo en `$TMPDIR` con el cambio, cada uno de los cuatro tests devuelto a `main` (`tests enteros: 1` o `subtests: 1`, con el test nombrado; §5), la copia congelada con un byte distinto («… differ») y los tres rojos de FR-010 de §5, con las órdenes de Q8 («Los casos de FR-010, sin su paso en el programa») y cada mutante del programa igual en los dos guiones: (1) `instalar` sin la llamada a `cambiarUnaVez`: cae el `grep` de la fila estrechada en `administrativo.md`; quitado ese `grep`, el de la fila del orden social en `prueba.md`; quitados los dos, su `! grep` en `administrativo.md` (`unexpected match for`); (2) `instalar` sin copiar `paquete/prueba-de-normativa.md`: la salida da `líneas más abajo: 0`, cae el `grep` de la línea del paquete en `repo/skills/normativa/SKILL.md` y, quitado ese `grep`, `no match for ^orden del paso 2 de normativa, líneas más abajo: 1$`; (3) la fila de extranjería de `data/cobertura.yaml` de la copia cambiada, de modo que su viejo de `cambiosDeLaBase` no está una vez: el `exec go run` sale con 1 y la salida de errores dice «no lleva una sola vez»; por cada uno, la orden, lo esperado y lo visto; `make ci` en verde. Es la única tarea `[datos]` y no toca nada fuera de esos dos ficheros.

**Checkpoint**: la aceptación del hito (SC-003) en verde dentro de `make ci`.

---

## Phase 8: User Story 1 — con los datos de `main`, todo sale igual; el registro de los controles (Priority: P1)

**Goal**: comprobar en el árbol que el hito es solo de tests y dejar el registro de mutantes que piden SC-004, SC-005 y
SC-007.

**Independent Test**: `make ci` y `make skills-check` en verde (Q1, Q2); el diff contra `main` solo con los ficheros de
Q3; `mutantes.md` con una entrada por fila de la tabla «Mutantes» del plan.

- [ ] T006 [US1] Escribir `specs/043-h45-1-los-tests-del/mutantes.md` (SC-001, SC-002, SC-004, SC-005, SC-007; FR-003 a FR-010, FR-012, FR-013, FR-014; US1.1-US1.4; plan, tablas «Mutantes» y «Controles de umbral»; quickstart Q1-Q8; research D7): en el árbol, `make ci` (Q1), `make skills-check` (Q2), las tres órdenes de diff contra `main` de Q3 (las dos primeras sin salida; la tercera solo con los ficheros que lista Q3), `cmp` de Q5 y las órdenes de Q6 y Q7; y en un `git clone --no-hardlinks -q . "$TMPDIR/h451"` (con `origin/main`; V13), nunca en el árbol de trabajo, cada mutante de Q8 aplicado, su control corrido y deshecho hasta que `git -C "$TMPDIR/h451" status --short` no dé nada: L06 y L07 (`TestCoberturaDelRepositorio`), L04 y L05 con la sincronía (`h32-paquetes`, `no match for ^filas: 33 de 33$`, y `TestCoberturaDelRepositorio` en verde), P6 y otra línea (los dos casos), T1 y L29 (`cuatro-evals-mas`), las ocho evals de más (FR-008), L33 y las tres pasadas a lo no declarado (`precepto-de-la-cobertura`), cada uno de los cuatro tests devuelto a `main` (`h45-paquete-de-mas`, SC-005), la copia congelada distinta (FR-013), los tres de FR-010 (sin `cambiarUnaVez`: el `grep` de la fila estrechada, el de la fila del orden social y su `! grep`, uno tras otro; sin la referencia de `normativa`: el `grep` de su línea y el `stdout` de `líneas más abajo: 1`; con el viejo que no está una vez: el `exec go run` con 1 y «no lleva una sola vez») y una línea cambiada en una referencia generada de `skills/plazos-y-territorio/references/administrativo.md` del clon (`make skills-check` en rojo con una diferencia, el umbral de SC-002). Por cada uno, la orden exacta, lo esperado y lo visto (el principio de la salida real), sin dar por visto nada que no se haya ejecutado en la sesión. Comprobar al final que el árbol de trabajo solo cambia en `specs/043-h45-1-los-tests-del/mutantes.md` y que `make ci` sigue en verde.

**Checkpoint**: el hito, listo para la revisión final del workflow.

---

## Controles de umbral → tarea que construye el control

| Fila de «## Controles de umbral» de plan.md | Control | Tarea que lo construye y lo ve en rojo |
|---|---|---|
| SC-001 (`make ci` con 0 fallos) | `make ci` (existe; cada tarea lo deja en verde) | T001-T005 lo dejan en verde; T006 lo corre y lo registra |
| SC-002, FR-014 (0 diferencias en `skills/` y en las salidas del binario) | `make skills-check` (existe, sin cambios) | T006: verde en el árbol y rojo en el clon con una línea cambiada en una referencia generada |
| SC-003, FR-011 (los tres códigos en 0) | `h45-paquete-de-mas` | T005: lo construye; rojo con cada test devuelto a `main` (también en T006) |
| FR-010 (fila estrechada, fila del orden social, línea en `normativa`, orden una más abajo) | los `grep` y el `stdout` del guion | T005: los construye; rojo sin `cambiarUnaVez` (los dos `grep` y el `! grep` de las filas), sin la referencia de `normativa` (el `grep` de su línea y el `stdout`) y con un viejo de `cambiosDeLaBase` que no está una vez (`exec go run` con 1) (también en T006) |
| FR-013 (0 bytes de diferencia) | el `cmp` del guion | T005: lo construye; rojo con la copia congelada distinta (también en T006) |
| FR-003 (las 9 tuplas; 0 filas fuera de ellas o de su orden) | `TestCoberturaDelRepositorio` | T001: rojo con L06 y L07 (también en T006) |
| FR-004, SC-007 (33 de 33 filas) | `h32-paquetes` (existe, sin cambios) | T006: rojo con L04 y L05 |
| FR-005 (1 defecto exacto en la línea de la orden) | `TestControlesSobreLasReferenciasAMano` | T002: rojo con P6 y con otra línea (también en T006) |
| SC-006, FR-006, FR-007 (16 evals; cabe a su concurrencia y no a dos) | `TestDefinicionDelJob/cuatro-evals-mas` | T003: rojo con T1 y L29 (también en T006) |
| FR-008 (peor caso de las del repositorio ≤ tope) | `/tope-con-las-evals-del-repositorio` (existe, sin cambios) | T003 y T006: rojo con ocho evals de más |
| FR-009 (≥ 1 de lo regulado aparte; 0 con precepto ajeno) | `TestEvalsDelRepositorio/precepto-de-la-cobertura` | T004: rojo con L33 y con las tres de lo no declarado (también en T006) |

## Definition of Done (docs/ROADMAP.md §1) que aplica

| Punto | Aplica | Tarea |
|---|---|---|
| 1 · `make ci` en verde | sí | T001-T006 (cada una lo deja en verde; T006 lo registra) |
| 2 · tests offline; fixtures grabados si toca red | sí, sin red | T001-T004 (tests offline); T005 (el guion, en `testdata/`, sin red: `GOPROXY=off`). Ninguna grabación: plan, «Datos externos» |
| 3 · sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | sí | T001-T005: ninguna importación nueva; el `os.Exit` y la salida del programa viven en el `.txtar`, fuera del módulo; `make lint` dentro de `make ci` |
| 4 · esquemas | no: ningún sobre ni esquema cambia | — |
| 5 · errores tipados, sin `panic` | no: ningún código de producción | — |
| 6 · comportamiento visible: e2e y `CHANGELOG.md` | no cambia ninguno (FR-014; research D8): sin `CHANGELOG.md`; el guion e2e que sí cambia es la copia de H45 | T005 |
| 7 · ADR | no: el spec deja fuera cambiar el ADR 0062 o escribir otro | — |
| 8 · `docs/SOURCES.md` | no: ninguna fuente externa | — |
| 9 · cobertura | no cambia: solo código de test | — |
| 10 · skills y evals | no: ninguna skill, referencia ni eval cambia (FR-014) | — |
| 11-13 · territorio, grafo, plazos | no: ningún applet cambia | — |

## Dependencies & Execution Order

- **T001, T002, T003**: independientes entre sí (ficheros distintos) y de lo demás; cada una deja verde el guion de hoy.
- **T004**: detrás de T002 (mismo fichero, `internal/evals/conjunto_test.go`).
- **T005**: detrás de T001-T004: la copia nueva solo pasa con los cuatro tests genéricos (research P1, D6).
- **T006**: la última: registra los controles sobre el árbol con T001-T005 hechos y commiteados (el clon parte de `HEAD`).

Orden de ejecución del bucle: T001 → T002 → T003 → T004 → T005 → T006.

## Parallel Example

```text
# Sin dependencias entre ellas (ficheros distintos):
T001 internal/skills/cobertura_test.go
T002 internal/evals/conjunto_test.go
T003 internal/evals/definicion_test.go
# Después: T004 (conjunto_test.go, tras T002); T005 (los dos .txtar); T006 (mutantes.md).
```

## Implementation Strategy

- **MVP**: T001 (US2), el primero de los cuatro rojos del prototipo de H42.
- **Incremental**: T002-T004 quitan los otros tres rojos, uno por tarea, cada uno con `make ci` en verde y su mutante en
  rojo; T005 hace que la copia con un paquete más lo compruebe dentro de `make ci`; T006 deja el registro.
- Nada de esto lo publica ninguna tarea: el cierre lo hace el workflow tras la revisión final.
