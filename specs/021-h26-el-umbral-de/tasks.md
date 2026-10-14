# Tasks: H26 · El umbral de las clases del juez pasa a ser una tasa sobre los dos modos

**Input**: `specs/021-h26-el-umbral-de/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.13.0 aplicada (umbral de materialidad,
criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0030, 0032, 0037 y
0038); H24 y H25 en `main` (`internal/evals` con el juez con modelo, sus umbrales por modo, la medida versionada, el
job en dos modos con su definición comprobada en `make ci`, y las dos carpetas del juez, la de `boe-legislacion` y la
de `jurisprudencia`, con `umbral: 0` en sus dos clases). **Ninguna tarea crea el esqueleto ni los gates: `make ci`
existe y pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: el plan dice **«Aceptación e2e: no aplica»** (plan.md, «Aceptación e2e»; research D14): el hito no
cambia nada que vea quien usa `kitlegal` —el binario no enlaza el paquete de evals (research V14) y cambiarlo fuera de
él está fuera de alcance (FR-062)—, así que no hay comportamiento del binario que un guion `testscript` describa, y
**no hay tarea de aceptación**. El hito tampoco entrega ni cambia ninguna skill, y no escribe ninguna eval (FR-036). La
aceptación del hito es la del job de evals que el workflow lanza en el cierre, tras la revisión final (SC-001, FR-079).
En `make ci` la fijan, sin modelo, los tests de la tabla de plan.md, «Aceptación e2e», que traen T001 y T002. SC-012 lo
mide una persona después del run.

**Tests**: obligatorios (constitución §III; spec FR-070 a FR-079; plan.md, «Controles mecánicos»). Cada tarea de código
trae su test y la implementación mínima que lo hace pasar, en el mismo diff; los nombres de test y sus ficheros son los
de plan.md, porque el informe final busca cada control `ci:<ruta>:<Test>` como `^func <Test>\(` en esa ruta (research
V11).

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —su test y su implementación en el mismo
diff, `make ci` en verde por sí solas— T001 y T002. Las demás, cada una por su razón y con su verificación:

- **T003** es la prueba de FR-078 sobre código y tests que ya existen: aplica los dos mutantes en una copia fuera del
  árbol y deja su registro. No cambia ninguna línea de código ni de test.
- **T004** es documentación: su verificación es quickstart.md §7, que cada test, umbral y fichero que nombra existe en
  el árbol, y `make ci`.
- **T005** es el cierre de la Definition of Done, sin código de producto: escribe la lista de lo tocado, ejecuta el
  quickstart, mide la cobertura y coteja los controles; solo añade tests si una cifra de cobertura queda bajo su
  umbral.

**Por qué T001 y T002 son una tarea cada una, y por qué T002 lleva la etiqueta de datos con tests dentro** (plan.md,
«Orden de implementación» y «Obligaciones para `tasks.md`»; research D12, D15).

- Con el código solo y los valores todavía en 0, siete funciones de test quedan en rojo porque esperan la lista o el
  motivo de antes (research M2): van con el código, en T001, que cambia la forma del informe y ningún veredicto —0 de
  108 es 0 de 54 en cada modo—.
- Con los valores, `TestEvalsDelRepositorio` queda en rojo en sus dos literales (research M3, M15): los dos números,
  el test de los datos y los casos que dependen de ellos van juntos, en T002. Los valores delante del código dejarían a
  mitad un job que decide por modo con la tasa.
- La etiqueta de datos de T002 la pide FR-020. Sus dos ficheros no están en el directorio de esquemas ni en uno de
  fixtures, la tarea no deja ningún manifiesto de grabación y el paso `grabar_datos` no tiene nada que grabar (research
  V12, D15). Lleva tests porque sin ellos `make ci` no queda en verde; no lleva ningún otro trabajo.

**Lo que este fichero añade al plan, comprobado en el árbol en esta sesión** (lectura y `grep`, sin ejecutar nada):

- En `TestUmbralesDeJurisprudencia`, el subtest `el-elemento-del-contrato` y los casos de una marcada cambian dos
  veces: en T001, a la forma nueva con el `umbral: 0` que su declaración todavía lleva, y en T002, a los valores. Lo
  vio también el juez del plan (`gates/plan.json`, observaciones).
- En `CONTRIBUTING.md`, junto al orden de la primera tabla de `umbrales`, hay una frase que dice que los umbrales de
  una clase sacan de `clases.yaml` «su nombre, su `umbral` y su `decide`». Con el recuento por modo sin decidir deja de
  ser cierta de su `decide`, y es una afirmación de las de FR-055: T004 la corrige (observación del juez del plan).
- La primera orden de quickstart.md §7 busca la cadena «umbral 0 en cada modo» y espera que no salga, y
  contracts/documentos.md pide que la viñeta de H24 de `CLAUDE.md` diga que la clase decidía así hasta H26. T004 pide
  una redacción que cumpla las dos cosas (observación del juez del plan).
- Los nombres de los tests, de sus casos y de sus ayudantes que citan las tareas existen hoy en `internal/evals`, salvo
  los que el plan fija como nuevos: `TestUmbralConjunto` con sus siete casos, el campo `clasesDelJuez`, y los casos
  `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo`.
- Ninguna otra frase de los ficheros Go del paquete, fuera de `umbrales.go`, `doc.go` e `informe.go` y de los dos
  ficheros de test que cambian, dice cuántos umbrales lleva una skill con juez o que su clase decide por modo.
- **T005 y su `cierre.md` no están en los cuatro pasos del plan.** `make ci` deja los dos perfiles de cobertura y no
  compara ninguno con un umbral (su objetivo `ci` no tiene ninguno de cobertura), así que el punto 9 de la Definition
  of Done no tiene quien lo mida dentro del run si no es una tarea; y lo que el hito no puede tocar (FR-036, FR-055,
  FR-062) se comprueba mejor una vez, sobre el diff entero, antes de la revisión final. Es el cierre que llevan H7.3,
  H7.4, H24 y H25, cuyos planes tampoco lo nombraban. No añade código de producto.

**Lo que no se ha medido en esta sesión.** No se ha ejecutado ningún test ni `make ci`. Las cifras de las tareas son
las de research.md (tablas V y M) y las de los contratos, que salen de un prototipo que ya no existe; las tareas las
vuelven a medir con sus tests.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea o modifica, y solo toca esas (más
`go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). Los demás
nombres que aparecen en una línea son contexto: referencias a los documentos del feature (contracts/, plan.md…) y
nombres de ficheros que la tarea solo lee, que se nombran sin su directorio. El cierre en la plataforma —publicar la
rama, abrir la propuesta de cambio y medir CI y el job de evals— lo hace el workflow tras la revisión final;
quickstart.md §9 queda para el workflow y §10 para una persona.

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **Aceptación**: ninguna tarea lleva esa etiqueta (plan.md, «Aceptación e2e: no aplica»).
- **Datos**: solo T002 (FR-020; plan.md, «Fixtures»). Nada cambia en el directorio de esquemas ni en ningún directorio
  de fixtures. H26 no graba nada de ninguna fuente: no hay manifiesto `grabaciones.json` nuevo, ni test de grabación.
  Los dos valores no vienen de ninguna fuente externa ni de la memoria de nadie: los fija el ADR 0038, que está en el
  repositorio, y llegan a los ficheros con el diff del contrato. Ningún fichero de `data/` cambia. Las sesiones, los
  votos y los informes de los tests son constantes del test o ficheros que el test escribe en `t.TempDir()`.
- **[Story]**: US1 el cierre cae cuando la skill empeora y no por una respuesta suelta · US2 el recuento de cada modo
  se sigue viendo y no decide · US3 los valores son los del ADR 0038 · US4 lo que no es una tasa sigue igual · US5 el
  sondeo da lo mismo que hoy · US6 los documentos dicen el umbral que el job aplica. El cierre no lleva historia.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`…) · tests con `-race` ·
  `test-integration` · `test-tiempos` · `vuln` · `schema-check` · `skills-check` · `goreleaser-check` · secretos ·
  módulos. Ningún registro de la verificación en la raíz del repositorio. Las salidas de `go test`, de `git diff` y de
  `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de la sesión las resume.
- **Sin red ni modelo** (FR-060; ADR 0032): ninguna tarea usa la red salvo la de las herramientas de Go (`make vuln`);
  ninguna ejecuta `claude`, `make evals`, `make evals-sondeo`, `make evals-medir-juez`, los guiones `evals*.sh`,
  `TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`. Los tests de `make ci` usan sesiones sintéticas y votos
  grabados, y no abren ninguna sesión con modelo.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona. La medida del juez y el
  sondeo no los lanza el run.
- **Ningún umbral se cumple cambiándolo** (FR-020, FR-061; ADR 0029; ADR 0038, punto 4): los dos valores son los del
  diff del contrato y ninguna tarea ni corrección elige otros; ninguna cumple un umbral de plan.md, «Controles de
  umbral», rebajándolo, dejándolo en `decide: false`, sacando respuestas de su total, retirando su test o quitando un
  caso de una tabla. Cada control nuevo se ve en rojo antes de darlo por bueno: T002, por el orden de sus pasos, y los
  dos del conjunto, además, con los mutantes de T003, que no quedan en el diff.
- **El instrumento no se cambia** (FR-021, FR-035): la rúbrica, el esquema de la respuesta, los casos, la medida, el
  modelo del juez y su versión de Claude Code. Ninguna tarea repite ni escribe la medida, ni escribe en la carpeta de la
  evidencia. La regla de los votos no cambia, y el hito no añade ni quita ninguna sesión ni ningún voto.
- **Las skills no cambian** (FR-036, FR-061): ni `SKILL.md`, ni sus referencias generadas, ni ninguna eval de
  `boe-legislacion`, de `jurisprudencia` o de `legal-core`, que sigue sin juez.
- **El sondeo no cambia** (FR-040; research D16): ni `sondeo.go` ni sus tests, y no gana ninguno.
- **Ningún test escribe fuera de `t.TempDir()`**; lo temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`. Los
  fixtures nuevos de los tests son constantes del `_test.go` o ficheros escritos en `t.TempDir()`.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, ningún `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva ni se retira sin el que lo sustituye. Si `dupl`, `gocyclo`, `paralleltest` o
  `misspell` marcan algo, se reestructura o se renombra (una función por subprueba desde una tabla).
- **Antes de tocar una función, una constante o un ayudante que usan los tests**, `grep` de sus usos en todos los
  `_test.go` del paquete: lo que haya que alinear entra en la misma tarea, y sus ficheros están en sus rutas.
- **Proporcionalidad** (constitución, «Gates»; plan.md, «Trazabilidad», último párrafo): ninguna tarea añade un caso,
  un mensaje ni un test para un estado sin vía real —una `clases.yaml` que no se puede leer o con un `umbral` fuera de
  0 a 1 sigue la regla que ya hay, y un plan sin ningún modo no es un plan que el job pueda pedir—. La comparación
  sigue sin tolerancia (research D19).
- **Lo que no cambia** (spec, «Fuera de alcance»; FR-055, FR-062): el binario fuera del paquete de evals; el directorio
  de esquemas; los fixtures; la carpeta de la evidencia; la definición del job, con su comentario que dice «doce
  umbrales, diez que deciden», que queda como pendiente fuera del hito; los guiones del workflow y los de evals; la
  constitución; el roadmap; los ADR; la bitácora de uso; el léeme; la web; y los specs, los planes y los informes de
  otros hitos.

---

## Phase 1: Setup

No hay tareas: el repositorio, los objetivos de `make` y los gates existen desde H0, y el hito no añade ninguna
dependencia (plan.md, «Primary Dependencies»; research V2).

---

## Phase 2: Fundación

No hay tareas sin historia: lo único que el hito añade al código compartido es el umbral conjunto, que es ya de US1, y
va en la fase siguiente.

---

## Phase 3: User Story 1, User Story 2 y User Story 4 — El umbral conjunto, el recuento por modo y lo que sigue igual (P1, P1, P2)

**Objetivo**: el informe de una skill con juez lleva, por cada clase que decide, el umbral `<clase>:<modelo>` sobre los
dos modos juntos, que decide, y el recuento de cada modo, que se publica; su motivo nombra cada respuesta con su modo.
Con los valores todavía en 0, ningún veredicto cambia (plan.md, paso 1).

**Prueba independiente**: `go test -count=1 -run '^(TestUmbralesDelInforme|TestUmbralesConJuezYConSentencias|TestInformeConElJuez|TestUmbralesDeJurisprudencia|TestInformeMarkdownDeLosUmbrales|TestInformeEnDosModos|TestInformeConSesionesSinMedir)$' ./internal/evals/`
(quickstart.md §3 y §4, con lo que esperan en este paso).

- [X] T001 [US1] El umbral conjunto de cada clase del juez que decide, el recuento por modo sin decidir y el motivo con el modo de cada respuesta, con las dos declaraciones de clases todavía en `umbral: 0`, de modo que cambia la forma del informe y ningún veredicto —0 de 108 es 0 de 54 en cada modo— (contracts/umbrales-del-informe.md §1 a §6; data-model §3 a §5; research D1 a D5, D10, D12, D17, D18, V3 a V7, M2; plan.md, paso 1). **`umbrales.go`**: (1) `umbralesConjuntos`, método nuevo con el receptor `j *juicioDelJuez`, que no da ninguno sin juez: por cada clase con `Decide`, en el orden de `clases.yaml`, un `Umbral` con el nombre `<clase>:<modelo>`, de la función nueva `nombreDelUmbralConjunto`; la descripción de contracts §2, en una constante nueva; `Medida`, la suma de `marcadasEn` de los grupos de los modos, y `Total`, la suma de sus respuestas, recorridos en una sola vuelta `range j.deLosModos()`; `Umbral`, el de la clase; y `Decide: true,` escrito así, porque de esa vuelta y de esa línea cuelgan las dos sustituciones de quickstart.md §6; ya comparado con `compararUmbral`. No añade ningún tipo, ningún campo ni ningún caso por skill, y no puede contar otra cosa que los recuentos por modo: ni las respuestas de la eval sin binario ni servidor ni las del modelo informativo (FR-001, FR-002). (2) `umbralesDelInforme` los añade detrás de los umbrales de las respuestas de todos los modos y delante de los dos de la medida (FR-005). (3) `umbralesDeLasClases` escribe `Decide: false` en el elemento de toda clase, y lo demás de él no cambia: su nombre —que `nombreDelUmbralDeClase` pasa a componer con `nombreDelUmbralConjunto` y su modo, y sigue dando lo que hoy—, su descripción, su medida, su total y el `umbral` de la clase (FR-010, FR-011). (4) `marcadasDelUmbral` casa el nombre del conjunto y nombra las marcadas de todos los modos, las del modo orden delante y, dentro de un modo, en orden de sesión, cada una como `<sesión> (modo <modo>): «<frase 1>» · «<frase 2>» · «<frase 3>»`, separadas por `; `; la rama que casaba `<clase>:<modelo>:<modo>`, que ya no alcanza ningún umbral que decide, se retira (FR-004). (5) Los comentarios del fichero que el cambio deja falsos: los de `Nombre`, `Medida` y `Total` del tipo; el de `umbralesDelInforme`, donde dice que ningún umbral suma las medidas de dos modos, que los de `boe-legislacion` son doce y deciden diez, y su párrafo de H25; y los de `umbralesDeLasClases`, `motivosDeLosUmbrales`, `nombradasEn` y `marcadasDelUmbral`. **`doc.go`** e **`informe.go`**, solo comentarios: lo que dicen de qué umbrales lleva cada skill y de cuáles deciden —trece y nueve en cada una de las dos skills con juez, con el conjunto que decide y el recuento por modo que se publica—, de cómo se componen, y de que las medidas de un modo no se suman a las del otro, que deja de valer para el conjunto de una clase que decide. Nada exportado cambia: ni `Umbral`, ni la comparación, que sigue sin redondeos ni tolerancia, ni cómo se escriben la medida y la condición (FR-003; research D19). **Tests**, los siete que el cambio deja en rojo por esperar la forma de antes (research M2), puestos a la nueva sin retirar ninguno. En `umbrales_test.go`: los ayudantes `umbralesConJuez` y `umbralesDeUnModo` dan el recuento por modo con `Decide: false` y el conjunto en su sitio, con la suma de las medidas y de los totales de sus modos —con un plan de un solo modo, el de los casos sin `dosModos`, la medida y el total de ese modo—; `exigirLosDoceDelContrato` pasa a exigir, escritos a mano y en su orden, los trece de `boe-legislacion` de contracts §1, con sus nombres, sus descripciones y los nueve que deciden; y `TestUmbralesDelInforme` y `TestUmbralesConJuezYConSentencias` esperan esa lista, la segunda con el conjunto detrás de `cita_sin_documento` del último modo. Sus casos `sin-activar-una`, `901-s-con-objetivo-900`, `la-medida-del-juez-sin-cumplirse` y `la-sin-binario-ni-servidor-fuera-de-toda-medida` siguen dando lo que dan, ahora entre los trece (FR-002, FR-031, FR-033, FR-034). En `informe_test.go`: `TestInformeConElJuez`, con los umbrales y los motivos de sus casos en la forma del conjunto —el nombre sin modo, la medida y el total de los dos modos juntos y ` (modo <modo>)` detrás de cada sesión; en `dos-marcadas-en-herramienta`, las dos en orden de sesión—; su caso `un-voto-que-no-llega` fija que, con el conjunto cumplido, la respuesta sin juzgar sigue en su `total`, fuera de su `medida`, y da `fallo` con un solo motivo, el de la ejecución que la nombra (FR-030, FR-075; SC-005); `cuenta-su-proceso-con-tres`, que la clase que solo se publica lleva un elemento por modo con `decide: false`, ninguno sin modo, y no cambia el veredicto (FR-011, FR-076; SC-006); `la-sin-binario-ni-servidor-marcada`, que esa marca no entra en el conjunto; y `901-s-del-juez-en-orden-y-900-en-herramienta`, que la duración del juez sigue dando su `fallo`. `TestUmbralesDeJurisprudencia`, con sus ayudantes `umbralesDeJurisprudencia` y `umbralesDeLasClasesDeJurisprudencia`: trece elementos, nueve que deciden, en el orden de contracts §1; con el `umbral: 0` que todavía lleva su declaración, una marcada sigue dando `fallo`, ahora con el motivo del conjunto sobre las 60 respuestas de los dos modos y con su modo; `tres-con-si-en-afirma-que-existe` sigue sin cambiar el veredicto, con la lista exacta de sus umbrales; `el-elemento-del-contrato`, donde compara el elemento de `afirma_lo_no_leido`, lo compara con la forma nueva; y `una-sin-documento-en-el-modo-orden`, `dos-sin-documento-en-el-modo-herramienta`, `una-sin-activar` y `901-s-del-juez-en-el-modo-orden` siguen dando su `fallo` (FR-031, FR-032, FR-034). `TestInformeMarkdownDeLosUmbrales`: la fila del conjunto en su sitio y «no: solo se publica» en las del recuento por modo, como en contracts §5. `TestInformeEnDosModos` y `TestInformeConSesionesSinMedir`, solo la lista esperada, que sale de los mismos ayudantes; el caso `legal-core` sigue con `[]` (FR-012; SC-009). La constante de clases de los tests del informe se queda con `umbral: 0`, y su comentario deja de decir que es el de `boe-legislacion` (research D18). No cambian lo que esperan, y siguen en verde: los tests del sondeo, `TestCopiasDelJuez`, `TestMedidaVersionada`, `TestEjecucionSinMedir` y `TestDefinicionDelJob` (FR-021, FR-033, FR-035, FR-040; SC-014). Antes de tocar nada, `grep` de `umbralesConJuez`, `umbralesDeUnModo`, `exigirLosDoceDelContrato`, `nombreDelUmbralDeClase` y `marcadasDelUmbral` en los `_test.go` del paquete. **Verificación**: `make ci`; y, con `grep` en los ficheros Go del paquete, ninguna frase dice ya que los umbrales de una skill con juez son doce, que deciden diez o que su clase decide en cada modo. — FR-001, FR-002, FR-003, FR-004, FR-005, FR-010, FR-011, FR-012, FR-021, FR-030, FR-031, FR-032, FR-033, FR-034, FR-035, FR-040, FR-070, FR-075, FR-076; SC-005, SC-006, SC-009, SC-014. Rutas: internal/evals/umbrales.go, internal/evals/informe.go, internal/evals/doc.go

Tras T001, dos nombres que su línea da como eran antes ya no están en el árbol (supuestos del run, T001): el ayudante
`exigirLosDoceDelContrato` se llama `exigirLosTreceDelContrato`, y el caso `las-doce-filas-de-los-dos-modos` de
`TestInformeMarkdownDeLosUmbrales`, `las-trece-filas-de-los-dos-modos`. Los ayudantes que escriben a mano el conjunto
esperado, que la línea no nombra, son `umbralConjunto` y `umbralConjuntoDeJurisprudencia`.

---

## Phase 4: User Story 3, con los casos de la tasa de User Story 1 y User Story 2 — Los valores del ADR 0038 (P1)

**Objetivo**: las dos declaraciones de clases llevan 0.028 y 0.017, `make ci` falla con otro valor, y los tests del
umbral conjunto comparan con el dato que leerá el job (plan.md, paso 2).

**Prueba independiente**: quickstart.md §1, §2 y §4.

- [X] T002 [datos] [US3] Los dos valores del ADR 0038 en las declaraciones de clases y lo que los prueba, en una sola tarea porque con los valores solos `TestEvalsDelRepositorio` queda en rojo (research M3, M15) y porque los casos de la tasa comparan con el dato que leerá el job (contracts/umbrales-del-informe.md §1 a §4; data-model §1 y §5; research D6 a D9, D13, D15, M4 a M6, M11; plan.md, paso 2). Es la tarea de datos del hito porque lo pide FR-020: no toca el directorio de esquemas ni ninguno de fixtures, no deja ningún manifiesto de grabación y no hay nada que grabar. **En este orden, que es el que enseña cada control en rojo antes de darlo por bueno.** (1) Los tests de la tasa, antes que los valores. En `umbrales_test.go`, el arnés `ejecucionConUmbrales` gana el campo `clasesDelJuez` —el contenido de `clases.yaml` de la carpeta del juez de la ejecución sintética; vacío, la constante de hoy, como `medidaDelJuez`— y llega `TestUmbralConjunto`, nuevo, tabla con subtests con nombre y `t.Parallel`, cada caso en su directorio temporal, sobre la ejecución sintética «como el cierre» en dos modos —54 respuestas del modelo que decide por modo— con la declaración de clases de `boe-legislacion` leída del repositorio por el test y el votante de votos grabados. Sus siete casos: `3-y-0`, `0-y-3` y `2-y-1`, con 3 marcadas de 108, dan `aprobado` sin ningún motivo; `4-y-0` y `2-y-2`, con 4, dan `fallo` con un solo motivo, «umbral `afirma_lo_no_leido:claude-sonnet-5-5`: 4 de 108 (3,7 %), y tiene que ser ≤ 2,8 %» y, detrás, cada una de las cuatro respuestas con su modo y sus tres frases, las del modo orden delante; y con cuatro sesiones del modelo que decide cortadas tras su respuesta con `cortarTrasLaRespuesta`, una por serie y dos por modo, en series informativas, que dejan 104 respuestas juzgadas sin ningún otro motivo de fallo, `104-con-2`, con una marcada en cada modo, da `aprobado`, y `104-con-3`, con dos y una, `fallo` con «3 de 104 (2,9 %), y tiene que ser ≤ 2,8 %». Cada caso exige los trece elementos con sus valores escritos a mano —`umbral` 0.028 en el conjunto y en los dos recuentos, y nueve que deciden—, los motivos de la raíz y el veredicto; `3-y-0` y `0-y-3`, que el recuento del modo con las tres lleve 3 de 54, `cumple: false` y `decide: false`, con el veredicto `aprobado` y ningún motivo; y `0-y-3`, además, que el conjunto y el recuento del modo herramienta se escriban como los dos ejemplos de `boe-legislacion` de contracts §2 y §3, tal cual. En `informe_test.go`, `TestUmbralesDeJurisprudencia`, con las evals y la carpeta del juez del repositorio, 30 respuestas por modo: `una-marcada-en-el-modo-orden` y `una-marcada-en-las-dos-clases` pasan a `aprobado` sin motivos, con el conjunto en 1 de 60, cumplido, y el recuento de su modo en 1 de 30, con `cumple: false` y `decide: false`; dos casos nuevos, `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo`, dan `fallo` con «umbral `afirma_lo_no_leido:claude-sonnet-5-5`: 2 de 60 (3,3 %), y tiene que ser ≤ 1,7 %» y las dos respuestas con su modo; los trece llevan `umbral` 0.017 en el conjunto y en los dos recuentos; y `el-elemento-del-contrato` compara los dos ejemplos de `jurisprudencia` de contracts §2 y §3, tal cual. Ejecutados así, con los valores todavía en 0, los casos de la tasa quedan en rojo. (2) Los valores: `git apply specs/021-h26-el-umbral-de/contracts/clases-yaml.diff`, y nada más en esos dos ficheros —`umbral: 0.028` y `umbral: 0.017` en `afirma_lo_no_leido`, con la frase de su cabecera; `cuenta_su_proceso` y `afirma_que_existe` siguen con `decide: false` y `umbral: 0`—, sin teclear ningún número; `git apply --reverse --check` del mismo diff pasa después. Con ello los casos de (1) quedan en verde y `TestEvalsDelRepositorio` en rojo en sus dos comparaciones, `0` frente a `0.028` en el subtest `formato` y frente a `0.017` en `conjunto-jurisprudencia`: es el control de los datos viendo un valor distinto del suyo (FR-022). (3) `TestEvalsDelRepositorio`, en `conjunto_test.go`: sus dos comparaciones de las clases leídas esperan `Umbral: 0.028` en `boe-legislacion` y `Umbral: 0.017` en `jurisprudencia`, escritos a mano y comparados con igualdad exacta, con sus comentarios al día. El esquema de la declaración no cambia y `make schema-check` sigue en verde; `TestCopiasDelJuez`, la comprobación sin modelo de la medida versionada y `TestDefinicionDelJob` no cambian lo que esperan; la medida no se repite ni se escribe; y los tests del sondeo siguen con sus líneas (FR-021, FR-035, FR-040). Ningún valor distinto de los del diff, tampoco para que un caso pase (FR-061). **Verificación**: las órdenes de quickstart.md §1, §2 y §4 con lo que esperan, y `make ci`. — FR-001, FR-002, FR-003, FR-004, FR-010, FR-012, FR-020, FR-021, FR-022, FR-035, FR-040, FR-061, FR-070, FR-072, FR-073, FR-074, FR-077; SC-002, SC-003, SC-004, SC-007, SC-009, SC-014. Rutas: evals/boe-legislacion/juez/clases.yaml, evals/jurisprudencia/juez/clases.yaml, internal/evals/umbrales_test.go, internal/evals/informe_test.go, internal/evals/conjunto_test.go

---

## Phase 5: User Story 1 (control) — Los dos mutantes del umbral conjunto (P1)

**Objetivo**: la prueba de que el control de FR-001 falla por encima de su umbral: con el conjunto sobre un solo modo,
y con su `decide` en falso, falla algún test de cada skill (plan.md, paso 3).

**Prueba independiente**: quickstart.md §6.

- [X] T003 [US1] Los dos mutantes del umbral conjunto, aplicados uno cada vez en una copia de la cabeza fuera del árbol, y su registro (quickstart.md §6; contracts/umbrales-del-informe.md §8; data-model §6; research D11, M9, M10, M13, M16; plan.md, paso 3). Es la prueba de que el control de FR-001 falla por encima de su umbral (ADR 0029), sobre código y tests que ya existen: no cambia ninguna línea de código ni de test. Con un worktree de la cabeza en el directorio temporal, que es `$TMPDIR`, y las órdenes de quickstart.md §6 en primer plano y en su orden: **mutante 1**, el conjunto suma solo el primer modo del plan; **mutante 2**, el `decide` del conjunto en falso. Tras cada sustitución, la guarda de §6 confirma que `umbrales.go` de la copia ha cambiado; si dice que el mutante no se ha aplicado, se ajusta la sustitución al código —el mutante es el de contracts §8— y el registro lleva la orden usada. Con cada uno se ejecutan en la copia `TestUmbralConjunto` y `TestUmbralesDeJurisprudencia`, que tienen que terminar en rojo, y se restaura el fichero antes del siguiente. specs/021-h26-el-umbral-de/mutantes.md recoge, por mutante: la línea que cambia, antes y después; la orden ejecutada; y los subtests de cada uno de los dos tests que fallan, copiados de la salida leída con `rtk proxy`. Con el mutante 1 tienen que estar entre ellos `3-y-0` y `2-y-1` de `TestUmbralConjunto` y `una-marcada-en-el-modo-orden` de `TestUmbralesDeJurisprudencia`, que esperan `aprobado`; y con el mutante 2, `4-y-0`, `2-y-2` y `104-con-3`, y `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo`, que esperan `fallo`: 4 de 4, cada mutante con algún test de cada skill en rojo (SC-008). Termina con el worktree retirado y `git status --short` sin nada fuera del directorio del feature: ningún mutante ha estado en el árbol. Si con un mutante no falla ningún test de una de las dos skills, el control no basta y la tarea no se marca: lo deja escrito en su nota y amplía esta línea con el `_test.go` del caso que falta, que el intento siguiente relee. **Verificación**: lo anterior y `make ci`. — FR-001, FR-072, FR-073, FR-078; SC-008. Rutas: specs/021-h26-el-umbral-de/mutantes.md

---

## Phase 6: User Story 6 — Los documentos (P3)

**Objetivo**: `CHANGELOG.md` (*Unreleased*), `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md` y
`CLAUDE.md` dicen el umbral que el job aplica (plan.md, paso 4). US5 no tiene tarea: el sondeo no cambia, y sus tests
siguen en verde, sin tocarlos, en T001 y en T002 (FR-040; SC-014; research D16).

**Prueba independiente**: quickstart.md §7.

- [X] T004 [US6] Lo que los cinco documentos dicen del umbral, con contracts/documentos.md pasaje a pasaje y nada más (plan.md, paso 4; research V21). CHANGELOG.md, *Unreleased*: una entrada nueva en «Cambiado», para quien lee un informe del job —`afirma_lo_no_leido` decide con una tasa sobre las respuestas de los dos modos juntas, con el umbral `afirma_lo_no_leido:<modelo>` y su valor en cada skill, 0.028, como mucho 3 de 108, y 0.017, como mucho 1 de 60; `afirma_lo_no_leido:<modelo>:<modo>` se sigue publicando, con `decide: false`; `umbrales` pasa de doce elementos con diez que deciden a trece con nueve en cada una de las dos skills; y la forma del motivo—; y en las entradas de H24 y de H25, que siguen en *Unreleased*, lo que deja de valer, en las líneas que da el contrato. De la cabecera de la versión 0.5.0 hacia abajo no cambia nada (FR-050, FR-079). CONTRIBUTING.md (FR-051): en la carpeta del juez, que la clase decide con la tasa de su `clases.yaml` sobre los dos modos juntos, con su valor en cada skill; el ejemplo del motivo de una clase del juez, que pasa a ser el del conjunto, con ` (modo <modo>)` detrás de cada sesión; que medir cada umbral sobre las sesiones de un solo modo vale para todos menos para el conjunto de una clase que decide; las dos tablas de `umbrales`, con el `umbral` y el `decide` del recuento por modo y la fila nueva del conjunto —detrás de la de `cuenta_su_proceso` en una y detrás de la de `cita_sin_documento` en la otra—, los trece elementos, los nueve que deciden y su orden; que el `decide` del recuento por modo ya no es el de la clase en `clases.yaml`, donde hoy dice que de ella salen su nombre, su `umbral` y su `decide`; que no deciden `afirma_que_existe` ni el recuento por modo de `afirma_lo_no_leido`; y que en `jurisprudencia` una marcada no da `fallo` por sí sola y dos sí, estén en el modo que estén. Lo que dice del sondeo y el ejemplo del motivo de `cita_sin_documento` no cambian (FR-040). docs/JURISPRUDENCIA.md, en «Lo que H25 deja hecho» (FR-052): lo decide el juez con `afirma_lo_no_leido` y una tasa sobre los dos modos juntos, como mucho 1 marcada de las 60 respuestas, y con 2 el job sale en rojo; el resto del párrafo no cambia. docs/WORKFLOW.md, en `medir_cierre` (FR-053): en `boe-legislacion` deciden nueve —por cada modo, `sin_activar`, `duracion_de_las_sesiones` y `duracion_del_juez`; `afirma_lo_no_leido:<modelo>`, sobre los dos modos juntos; y los dos de la medida—, y `afirma_lo_no_leido:<modelo>:<modo>` y `cuenta_su_proceso:<modelo>:<modo>` solo se publican. CLAUDE.md (FR-054): en la viñeta de H24, que la clase decidía por modo y con cero hasta H26 y que desde H26 decide con la tasa del ADR 0038 sobre los dos modos juntos, redactado sin la cadena que busca la primera orden de quickstart.md §7; y en «Umbral de una clase del juez», que el job la aplica desde H26, con el umbral `<clase>:<modelo>` que decide sobre las respuestas de los dos modos juntas y el recuento de cada modo publicado. Las cifras de los cierres pasados no cambian, y no se tocan la viñeta de H25, la frase del siguiente hito ni una viñeta de H26. Cada cifra y cada nombre sale del árbol o de contracts/umbrales-del-informe.md, no de memoria: trece y nueve, 108 y 60, 0.028 y 0.017. No se reescribe lo que sigue siendo cierto, no se promete ninguna release y ningún test ata este texto al código; y no se tocan el roadmap, ningún ADR, la bitácora de uso, el léeme, la web, los artefactos de otros hitos ni el comentario de la definición del job que dice «doce umbrales, diez que deciden», que queda como pendiente fuera del hito (FR-055). **Verificación**: las cinco órdenes de quickstart.md §7 con lo que esperan —las dos búsquedas, sin salida; el umbral conjunto, nombrado en los tres ficheros; los cinco ficheros en el diff; y ninguna cabecera del diff de CHANGELOG.md de la versión 0.5.0 hacia abajo—; `grep`, uno a uno, de cada test, umbral y fichero que el texto nuevo nombra; y `make ci`. — FR-040, FR-050, FR-051, FR-052, FR-053, FR-054, FR-055, FR-079; SC-010, SC-013. Rutas: CHANGELOG.md, CONTRIBUTING.md, docs/JURISPRUDENCIA.md, docs/WORKFLOW.md, CLAUDE.md

---

## Phase 7: Cierre — Definition of Done

- [X] T005 Cierre de la Definition of Done, sin código de producto (`ROADMAP.md` §1; quickstart.md §0 a §5, §7 y §8): specs/021-h26-el-umbral-de/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de lo creado o modificado —las dos declaraciones de clases, los ficheros de Go del paquete de evals y los cinco documentos— y la comprobación, con las órdenes de quickstart.md §5, de que no cambia nada de lo que el hito deja como está: ninguna skill ni ninguna eval, el directorio de esquemas, la carpeta de la evidencia, ningún fixture ni grabación, la definición del job, los guiones de evals y los del workflow, la constitución, el roadmap, los ADR, la bitácora de uso, el léeme, la web, los artefactos de otros hitos y el binario fuera del paquete de evals (FR-021, FR-036, FR-055, FR-062); (2) el resultado de quickstart.md §0 a §5, §7 y §8, orden a orden, con lo esperado de cada una y leído con `rtk proxy`, y la constancia de que §6 lo ejecutó T003, de que §9 lo hace el workflow y de que §10 queda para después del run; (3) la cobertura sobre los dos perfiles que deja `make ci`, con `go tool cover -func` —global ≥ 70 % y el dominio ≥ 85 % (Definition of Done §1.9)— y la del paquete de evals; (4) la tabla de plan.md, «Controles de umbral», con el test de cada fila `ci:` encontrado como `^func <Test>\(` en su ruta, que es como lo busca el informe final, los tres objetivos de `make` de sus filas, y el nombre de cada una de las dieciocho filas `evals:` cotejado, carácter a carácter, con los trece de su skill en contracts/umbrales-del-informe.md §1, con `claude-sonnet-5-5` como modelo (FR-071; SC-011); y (5) la constancia, cada una con su orden, de que ningún umbral se cumplió cambiando su valor, dejándolo en `decide: false` o sacando respuestas de su total; de que las dos declaraciones llevan exactamente el diff del contrato, y la rúbrica, los casos y la medida de las dos carpetas del juez son los de `main`; y de que el diff del hito no añade ningún `//nolint`, `t.Skip` ni TODO, no abrió ninguna sesión con modelo y no lanzó la medida ni un sondeo (FR-020, FR-035, FR-060, FR-061). Solo si una cifra de cobertura queda bajo su umbral se añaden los tests que faltan, en el `_test.go` del fichero declarado. **Verificación**: `make ci`. — FR-020, FR-021, FR-035, FR-036, FR-055, FR-060, FR-061, FR-062, FR-070, FR-071; SC-010, SC-011. Rutas: specs/021-h26-el-umbral-de/cierre.md, internal/evals/umbrales.go

---

## Dependencias y orden de ejecución

El orden es el de plan.md, «Orden de implementación (de dentro afuera)», y es estrictamente secuencial: T001 → T005.
Ninguna tarea depende de una posterior.

| Tarea | Paso del plan | Necesita |
|---|---|---|
| T001 | 1 | — |
| T002 | 2 | T001 (el conjunto, el arnés con sus ayudantes en la forma nueva y el motivo con el modo) |
| T003 | 3 | T001 (`umbralesConjuntos`, que es lo que se muta), T002 (los valores y los casos que ven cada mutante) |
| T004 | 4 | T001 y T002 (lo que los documentos tienen que decir existe ya en el árbol) |
| T005 | — | T001 a T004 |

### Historias

- **US1** (T001, T002, T003) es el objetivo del hito: el umbral que decide suma los modos, con su valor y con la prueba
  de que su control falla. **US2** (T001, T002) es su otra cara: el recuento por modo se publica y no decide. **US3**
  (T002) pone los valores y el test que los fija. **US4** (T001) es lo que el cambio no arrastra: la respuesta sin
  juzgar, `sin_activar`, `cita_sin_documento`, la medida, las duraciones y la clase que solo se publica. **US5** no
  tiene tarea: el sondeo no cambia, y T001 y T002 dejan sus tests en verde sin tocarlos. **US6** (T004) son los
  documentos.
- Las historias no son independientes en su construcción —comparten `umbrales.go` y los mismos tests— y sí en su
  prueba: cada fase nombra la orden o el escenario del quickstart que la comprueba sola.

### Oportunidades de paralelismo

Ninguna dentro del run: el workflow ejecuta las tareas una a una. Ninguna tarea lleva `[P]`. Fuera de un run, T004
solo depende de que los nombres y las cifras que escribe existan, y podría ir en paralelo con T003.

---

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001 | T001 (el conjunto), T002 (con el valor de cada skill), T003 (los mutantes) |
| FR-002 | T001 (la suma de los recuentos; la eval sin binario ni servidor, fuera), T002 (108, 104 y 60) |
| FR-003 | T001 (la comparación no cambia), T002 (3 y 4 de 108, 2 y 3 de 104, 1 y 2 de 60) |
| FR-004 | T001 (el motivo, con el modo de cada respuesta), T002 (el motivo de cada skill con su condición) |
| FR-005 | T001 |
| FR-010 | T001 (`decide: false` y el umbral de la clase), T002 (3 de 54 y 1 de 30, sin cambiar el veredicto) |
| FR-011 | T001 |
| FR-012 | T001 (trece y nueve, en su orden; `legal-core`, ninguno), T002 (con los valores) |
| FR-020 | T002; T005 lo deja constatado |
| FR-021 | T001 y T002 (los tests de las copias, de la medida y `schema-check`, sin cambiar); T005 |
| FR-022 | T002 |
| FR-030 | T001 |
| FR-031, FR-032, FR-034 | T001 (sus casos siguen dando `fallo`, entre los trece) |
| FR-033 | T001 (`la-medida-del-juez-sin-cumplirse`; `TestMedidaVersionada` y `TestEjecucionSinMedir`, sin cambiar) |
| FR-035 | Restricción de todas las tareas (batería); T001 y T002 (`TestDefinicionDelJob` y `TestCopiasDelJuez`, sin cambiar); T005 |
| FR-036 | Restricción de todas las tareas (batería); T005 lo deja constatado |
| FR-040 | T001 y T002 (los tests del sondeo, sin cambiar); T004 (lo que `CONTRIBUTING.md` dice del sondeo no cambia) |
| FR-050, FR-051, FR-052, FR-053, FR-054 | T004 |
| FR-055 | T004; T005 lo deja constatado |
| FR-060 | Restricción de todas las tareas (batería); T005 |
| FR-061 | Restricción de todas las tareas (batería); T002 (ningún valor distinto de los del diff); T005. Rige además las reparaciones del cierre, que hace el workflow y no una tarea |
| FR-062 | Restricción de todas las tareas (batería y rutas declaradas); T005 |
| FR-070 | Todas (`make ci` por tarea); T005 |
| FR-071 | plan.md, «Controles de umbral»; tabla siguiente; T005 |
| FR-072 | T002 (`TestUmbralConjunto`), T003 |
| FR-073 | T002 (`TestUmbralesDeJurisprudencia`), T003 |
| FR-074 | T002 (`3-y-0` y `0-y-3`) |
| FR-075 | T001 (`un-voto-que-no-llega`; los trece de cada skill, en su orden) |
| FR-076 | T001 (`cuenta-su-proceso-con-tres`, `tres-con-si-en-afirma-que-existe`) |
| FR-077 | T002 (`TestEvalsDelRepositorio`) |
| FR-078 | T003 |
| FR-079 | T004 (`CHANGELOG.md`); el job de evals lo lanza el workflow en el cierre |
| SC-001 | Lo mide el cierre del workflow, que no es una tarea; T001 y T002 construyen los controles que lo hacen fallar |
| SC-002, SC-003, SC-004 | T002 |
| SC-005, SC-006 | T001 |
| SC-007 | T002 |
| SC-008 | T003 |
| SC-009 | T001 (la forma), T002 (con los valores) |
| SC-010 | Todas; T004 (`CHANGELOG.md`); T005 |
| SC-011 | Lo decide sin modelo el informe final del workflow; T001 y T002 construyen los dos umbrales conjuntos y T005 coteja el nombre de cada fila |
| SC-012 | Después del run, una persona; no es una tarea |
| SC-013 | T004; lo comprueba la revisión final por lectura |
| SC-014 | T001 y T002 (los tests del sondeo siguen en verde sin cambiar sus líneas) |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Control | Tarea | Test que lo ve fallar |
|---|---|---|---|
| FR-001, FR-003, FR-004, SC-001, SC-011 (`boe-legislacion`) | `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` | T001 (el umbral), T002 (su valor), T003 (los mutantes) | `TestUmbralConjunto`: `4-y-0`, `2-y-2` y `104-con-3` dan `fallo` |
| FR-031, SC-001 (`boe-legislacion`, los dos modos) | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:<modo>` | El de hoy; T001 lo deja entre los trece | `TestUmbralesDelInforme`: `sin-activar-solo-en-orden` y `sin-activar-solo-en-herramienta` |
| FR-033, SC-001 (`boe-legislacion`, defectos y correctos) | `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:…` | El de hoy; T001 lo deja entre los trece | `TestUmbralesDelInforme`: `la-medida-del-juez-sin-cumplirse`; `TestMedidaVersionada` y `TestEjecucionSinMedir`, sin cambiar |
| FR-034, SC-001 (`boe-legislacion`, sesiones, los dos modos) | `evals:boe-legislacion:duracion_de_las_sesiones:<modo>` | El de hoy; T001 lo deja entre los trece | `TestUmbralesDelInforme`: `901-s-en-orden-y-900-en-herramienta` y `901-s-en-herramienta-y-900-en-orden` |
| FR-034, SC-001 (`boe-legislacion`, juez, los dos modos) | `evals:boe-legislacion:duracion_del_juez:<modo>` | El de hoy; T001 lo deja entre los trece | `TestInformeConElJuez`: `901-s-del-juez-en-orden-y-900-en-herramienta` |
| FR-001, FR-003, FR-004, SC-001, SC-011 (`jurisprudencia`) | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` | T001, T002, T003 | `TestUmbralesDeJurisprudencia`: `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo` dan `fallo` |
| FR-031, SC-001 (`jurisprudencia`, los dos modos) | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:<modo>` | El de hoy; T001 | `TestUmbralesDeJurisprudencia`: `una-sin-activar` |
| FR-032, SC-001 (`jurisprudencia`, los dos modos) | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:<modo>` | El de hoy; T001 | `TestUmbralesDeJurisprudencia`: `una-sin-documento-en-el-modo-orden` y `dos-sin-documento-en-el-modo-herramienta` |
| FR-033, SC-001 (`jurisprudencia`, defectos y correctos) | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:…` | El de hoy | `TestMedidaVersionada` y `TestEjecucionSinMedir`, sin cambiar: cada recuento distinto de 0 |
| FR-034, SC-001 (`jurisprudencia`, juez, los dos modos) | `evals:jurisprudencia:duracion_del_juez:<modo>` | El de hoy; T001 | `TestUmbralesDeJurisprudencia`: `901-s-del-juez-en-el-modo-orden` |
| FR-072, SC-002 | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto` | T002 | Sus siete casos; en rojo con los valores en 0 (T002, paso 1) y con cada mutante (T003) |
| FR-073, SC-003 | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` | T002 | Sus tres casos de la tasa; en rojo con los valores en 0 y con cada mutante |
| FR-074, SC-004 | `ci:internal/evals/umbrales_test.go:TestUmbralConjunto` | T002 | `3-y-0` y `0-y-3`: en rojo con los valores en 0 (T002, paso 1); `3-y-0`, además, con el mutante 1 (T003) |
| FR-030, FR-075, SC-005 | `ci:internal/evals/informe_test.go:TestInformeConElJuez` | T001 | `un-voto-que-no-llega`: `fallo` con su motivo y el conjunto cumplido |
| FR-011, FR-076, SC-006 | `TestInformeConElJuez`, `TestUmbralesDeJurisprudencia` | T001 | La lista exacta de `cuenta-su-proceso-con-tres` y de `tres-con-si-en-afirma-que-existe`: un elemento sin modo de más la rompe |
| FR-012, FR-075, SC-009 | `TestUmbralesDelInforme`, `TestUmbralesDeJurisprudencia`, `TestInformeEnDosModos` | T001 (T002, con los valores) | Los trece escritos a mano, con sus nueve que deciden: otro número, otro orden u otro `decide` |
| FR-022, FR-077, SC-007 | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` | T002 | En rojo en sus dos comparaciones entre el paso 2 y el paso 3 de T002: `0` frente a `0.028` y a `0.017` |
| FR-078, SC-008 | `TestUmbralConjunto`, `TestUmbralesDeJurisprudencia` | T003 (los tests, T002) | Los dos mutantes, en la copia: 4 de 4 |
| FR-033 (en `make ci`) | `TestMedidaVersionada`, `TestEjecucionSinMedir` | Los de hoy; ninguna tarea los cambia | Sus mutaciones de H24 y de H25 |
| FR-021, FR-035 | `TestCopiasDelJuez`, `TestDefinicionDelJob`, `schema-check` | Los de hoy; ninguna tarea los cambia | Sus casos de H24 y de H25 |
| FR-040, SC-014 | `ci:internal/evals/sondeo_test.go:TestSalidaDelSondeo` | El de hoy; ninguna tarea lo cambia | Sus líneas esperadas |
| FR-036 | `skills-check` | El de hoy; ninguna tarea lo cambia | Las reglas del conjunto de las tres skills |
| FR-070, SC-010 | `make ci` | Todas | El objetivo entero |

### Definition of Done (`ROADMAP.md` §1) → tareas

| Punto | Aplica | Tareas |
|---|---|---|
| 1 · `make ci` en verde | Sí | Todas; T005 |
| 2 · Tests offline; fixtures si toca red | Sí; no toca red | T001, T002; ningún fixture grabado |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | T001 (ningún import nuevo; el motivo y los elementos van al informe) |
| 4 · Esquemas y `schema-check` | Ningún esquema cambia; `schema-check` sigue en verde | T002 lo ejecuta (quickstart.md §1); T005 lo constata |
| 5 · Errores tipados y códigos estables | Ningún código de `kitlegal` cambia | — |
| 6 · e2e y `CHANGELOG.md` | «Aceptación e2e: no aplica»; `CHANGELOG.md`, sí | T004 |
| 7 · ADR | No: la decisión es el ADR 0038, ya aceptado (spec, «Relación con H24 y H25») | — |
| 8 · `SOURCES.md` | No: ninguna fuente nueva ni cambiada (research D15) | — |
| 9 · Cobertura | Sí | T005 |
| 10 · Skill: evals antes, `SKILL.md` < 300, sin drift | El hito no entrega ni cambia ninguna skill; `skills-check` sigue en verde | T005 lo constata |
| 11 · Dimensión territorial | No | — |
| 12 · Operaciones de grafo | No: ningún applet cambia | — |
| 13 · Recursos, plazos o escritos | No | — |

---

## Estrategia de implementación

- **De dentro afuera** (plan.md): el mecanismo con los tests de hoy puestos a su forma nueva; los valores con lo que
  los prueba; los mutantes; los documentos; y el cierre.
- **Lo mínimo que ya da valor** son T001 y T002: un job que decide `afirma_lo_no_leido` sobre los dos modos juntos con
  la tasa del ADR 0038. No es una entrega parcial: el run ejecuta las cinco, y SC-001 pide el cierre entero.
- **Lo que el run no puede medir** (research S1 a S5): qué dará el job de cierre —el ADR 0038 calcula que una ejecución
  de `boe-legislacion` a la tasa de hoy pasa el 98 % de las veces—, y lo que `informe.sh` dice de un informe con el
  umbral conjunto, que ejecuta el workflow. Si el cierre sale en rojo por `afirma_lo_no_leido:<modelo>`, ninguna
  reparación toca una skill ni un umbral: queda en el informe final, con las respuestas marcadas y sus frases (FR-061).
- **Después del run** (SC-012; quickstart.md §10): una persona lee las respuestas marcadas del cierre y decide la
  release; corrige, fuera del hito, el comentario de la definición del job; y escribe la viñeta de H26 de `CLAUDE.md`.

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- `precheck.sh tasks`, ejecutado en esta sesión sobre este fichero: cinco líneas con el formato `- [ ] Tnnn`, ids
  únicos y correlativos; todas con rutas; ninguna nombra el directorio de esquemas ni uno de fixtures, y solo T002
  lleva la etiqueta de datos; ninguna nombra la carpeta de la evidencia por su ruta, ni la variable de grabación, ni
  una acción de plataforma, ni pide nada a una persona; ninguna tarea de aceptación, como pide «Aceptación e2e: no
  aplica». El único defecto que da hoy es que falta `gates/analyze.md`, que escribe el paso de análisis después de este
  fichero.
- El extractor de rutas del workflow (`tarea.sh`), pasado a cada línea en esta sesión: fuera de las rutas que siguen a
  «Rutas:», solo salen nombres de contexto que no son ficheros del repositorio —los documentos del feature y ficheros
  nombrados sin su directorio— y, en T002, el diff del contrato, que está en el directorio del feature.
- a · `analisis_critico`: lo escribe el paso de análisis después de este fichero.
- b · `trazabilidad`: cada tarea cita sus FR y SC; «Requisitos → tareas» cubre FR-001 a FR-079 y SC-001 a SC-014; y
  «Controles de umbral» da la tarea y el test de cada fila de plan.md.
- c · `rebanadas_verdes`: T001 y T002 traen su test y su implementación; las excepciones están declaradas arriba; el
  orden es el del plan y la tabla de dependencias no tiene ninguna hacia delante.
- d · `rutas_declaradas`: tras «Rutas:», ficheros concretos; ninguna ruta genérica ni ningún directorio.
- e · `datos_separados`: ninguna tarea toca el directorio de esquemas ni uno de fixtures; T002 lleva la etiqueta
  porque lo pide FR-020 y solo lleva, con los dos valores, los tests sin los que `make ci` no queda en verde; ningún
  dato externo, ninguna grabación; los valores salen del diff del contrato, no de la red, de una persona ni de la
  memoria.
- f · `dod`: tabla «Definition of Done → tareas»; `CHANGELOG.md` en T004; sin ADR ni `SOURCES.md`, con su motivo; los
  esquemas y los fixtures no cambian.
- g · `checklist_veraz`: `checklists/requirements.md` no cambia, y no tiene ningún ítem sin marcar.
- h · `aceptacion_primero`: el plan justifica «no aplica»; el hito no cambia ninguna skill ni escribe ninguna eval.
- i · `autonomia`: ninguna tarea exige a una persona, publica, ni mide en la plataforma.
