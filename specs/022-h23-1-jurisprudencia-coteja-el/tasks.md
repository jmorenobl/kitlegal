# Tasks: H23.1 · `jurisprudencia` coteja el documento antes de decir nada de él, y un umbral sin modelo lo hace cumplir

**Input**: `specs/022-h23-1-jurisprudencia-coteja-el/` (spec.md, plan.md, research.md, data-model.md, contracts/,
quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.13.0 aplicada (umbral de
materialidad, criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0032,
0036, 0037 y 0038); H26 en `main` (`internal/evals` con los trece umbrales de `jurisprudencia`, nueve que deciden, el
juicio de cada sesión con sus invocaciones de la traza y de las llamadas, y `cita_sin_documento` por modo). **Ninguna
tarea crea el esqueleto ni los gates: `make ci` existe y pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: el plan dice **«Aceptación e2e: no aplica»** (plan.md, «Aceptación e2e»; research D14): el hito no
cambia el código del binario —`cmd/kitlegal` no enlaza el paquete de evals (research V14) y cambiarlo fuera de él está
fuera de alcance (FR-052)—, y el `SKILL.md` empotrado se instala igual que hoy (research V17). No hay comportamiento de
`kitlegal` que un guion `testscript` describa, así que **no hay tarea de aceptación**. Las evals que piden el cotejo
(la 04, la 06, la 08 y la 09 de `jurisprudencia`) ya existen y FR-032 no deja cambiarlas: la aceptación de la skill y
del instrumento es el job de evals que el workflow lanza en el cierre, tras la revisión final (SC-001, FR-066). En
`make ci` la fijan, sin modelo, los tests de la tabla de plan.md, «Aceptación e2e», que trae T001, y `make
skills-check`, que T003 deja en verde. SC-010 lo mide una persona después del run, fuera de él.

**Tests**: obligatorios (constitución §III; spec FR-060 a FR-064; plan.md, «Controles mecánicos»). T001 trae sus tests
y la implementación que los hace pasar en el mismo diff; los nombres de test y sus ficheros son los de plan.md, porque
el informe final busca cada control `ci:<ruta>:<Test>` como `^func <Test>\(` en esa ruta (research V12).

**Rebanadas verticales y excepciones declaradas.** Es rebanada completa —su test y su implementación en el mismo diff,
`make ci` en verde por sí sola— T001. Las demás, cada una por su razón y con su verificación:

- **T002** es la prueba de FR-064 sobre el código y los tests que deja T001: aplica los dos mutantes en una copia fuera
  del árbol y deja su registro. No cambia ninguna línea de código ni de test.
- **T003** es la skill, que no lleva código: su control de forma es `make skills-check` (`TestSkillsDelRepositorio` y
  `TestOrdenesDeLasSkillsEmpotradas`), que ya existe y no cambia (FR-023); su efecto lo mide el job de cierre con el
  umbral que construye T001 (SC-001).
- **T004** es documentación: su verificación es quickstart.md §6, que cada test, umbral y fichero que nombra existe en
  el árbol, y `make ci`.
- **T005** es el cierre de la Definition of Done, sin código de producto: la lista de lo tocado, el quickstart, la
  cobertura y el cotejo de los controles; solo añade tests si una cifra de cobertura queda bajo su umbral.

**Por qué el umbral y todos sus tests son una sola tarea** (plan.md, «Orden de implementación», paso 1; research M2):
con el código de producción solo, `TestUmbralesDeJurisprudencia` queda en rojo porque espera trece elementos donde el
informe da quince, y `TestInformeFinalDeLosUmbrales/como-en-el-cierre-de-h26` porque lee los umbrales de H26 de
`umbralesDeJurisprudencia(nil)`, que pasa a dar quince (research M4). Los tests que esperan la forma nueva y el código
que la da van juntos, o `make ci` no queda en verde.

**Lo que este fichero añade al plan, comprobado en el árbol en esta sesión** (lectura y `grep`, sin ejecutar ningún
test):

- Existen hoy, con el nombre que les da el plan: `TestUmbralesDeJurisprudencia` (`informe_test.go:5091`), su tabla
  `casosDeLosUmbralesDeJurisprudencia` con el caso `ninguna-cuenta` (`:5221`), `medidasDeJurisprudencia` (`:4980`),
  `mensajesDeLosVotosDeJurisprudencia` (`:5465`), `umbralesDeJurisprudencia` (`:5527`), `escribirSesionDeJurisprudencia`
  (`:5689`), `cambiarElFinalDeLaSesion` (`:2759`), `TestInformeFinalDeLosUmbrales` con `como-en-el-cierre-de-h26`
  (`:1830`, `:1840`) y `umbralesDeBoeLegislacionComoEnH26` (`:1897`); `TestEvalsDelRepositorio`
  (`conjunto_test.go:1899`); y, en producción, `grupoJuzgado` (`informe.go:520`), `medirLasRespuestas` (`:1324`),
  `medirLasCitasSinDocumento` (`:1354`), `respuestasQueSeJuzgan` (`:1605`), `respuestasMedidas.umbrales`
  (`umbrales.go:266`), `nombradasEn` (`:561`), `sinDocumentoDelUmbral` (`:577`), `satisface` (`juzgar.go:968`) e
  `invocacionesDe` (`juzgar.go:493`). Los nombres que el plan fija como nuevos no existen todavía.
- Los ayudantes del arnés de `jurisprudencia` solo los usa `informe_test.go`: ningún otro `_test.go` del paquete los
  nombra.
- Las frases de los ficheros Go del paquete que dicen cuántos umbrales lleva una skill con juez son
  `doc.go:183` y `umbrales.go:218` (research V21, que da además los comentarios de `informe.go` y de `EscribirInforme`).
- Lo que el punto 10 de la Definition of Done pide de una skill que cambia —evals escritas antes y en verde, `SKILL.md`
  por debajo de 300 líneas y sin drift— lo cumplen las cuatro evals que ya piden el cotejo, que FR-032 no deja tocar y
  que mide el job de cierre, y `make skills-check` en T003.
- **T005 y su `cierre.md` no están en los cuatro pasos del plan.** `make ci` deja los perfiles de cobertura y no los
  compara con ningún umbral, así que el punto 9 de la Definition of Done no tiene quien lo mida dentro del run si no es
  una tarea; y lo que el hito no puede tocar (FR-032 a FR-034, FR-045, FR-052) se comprueba mejor una vez, sobre el diff
  entero, antes de la revisión final. Es el cierre que llevan H7.3, H7.4, H24, H25 y H26. No añade código de producto.

**Lo que no se ha medido en esta sesión.** No se ha ejecutado ningún test ni `make ci`. Las cifras de las tareas son
las de research.md (tablas V y M) y las de los contratos, que salen de un prototipo que ya no existe; las tareas las
vuelven a medir con sus tests.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea, tras «Rutas:», **todas** las rutas que crea o modifica, y solo toca esas (más
`go.mod`, `go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). Los demás
nombres que aparecen en una línea son contexto: referencias a los documentos del feature (contracts/, plan.md…) y
nombres de ficheros que la tarea solo lee, que se nombran sin su directorio. El cierre en la plataforma —publicar la
rama, abrir la propuesta de cambio y medir CI y el job de evals— lo hace el workflow tras la revisión final;
quickstart.md §8 queda para el workflow y §9 para una persona, después del run.

## Formato: `- [ ] Tnnn [Story?] Descripción — FR/SC. Rutas: …`

- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia.
- **Aceptación**: ninguna tarea lleva esa etiqueta (plan.md, «Aceptación e2e: no aplica»).
- **Datos**: ninguna tarea lleva esa etiqueta (plan.md, «Fixtures»; research D16). Nada cambia en el directorio de
  esquemas, en ningún directorio de fixtures ni en la carpeta de la evidencia. H23.1 no graba nada de ninguna fuente:
  ni manifiesto `grabaciones.json` nuevo, ni test de grabación, y el paso `grabar_datos` no tiene nada que grabar.
  Ningún fichero de `data/` cambia. Las sesiones, los votos y los informes de los tests son constantes del test o
  ficheros que el test escribe en `t.TempDir()`.
- **[Story]**: US1 el job sale en rojo si una sesión da por bueno un documento sin cotejarlo · US2 la skill coteja el
  documento antes de decir nada de él · US3 lo que ya decidía sigue decidiendo igual · US4 los documentos dicen los
  umbrales que el job aplica. El cierre no lleva historia.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`, `paralleltest`…) · tests con `-race` ·
  `test-integration` · `test-tiempos` · `vuln` · `schema-check` · `skills-check` · `goreleaser-check` · secretos ·
  módulos. Ningún registro de la verificación en la raíz del repositorio. Las salidas de `go test`, de `git diff` y de
  `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de la sesión las resume.
- **Sin red ni modelo** (FR-051; ADR 0032): ninguna tarea usa la red salvo la de las herramientas de Go (`make vuln`);
  ninguna ejecuta `claude`, `make evals`, `make evals-sondeo`, `make evals-medir-juez`, los guiones `evals*.sh`,
  `TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`. Los tests de `make ci` usan sesiones sintéticas y votos
  grabados, y no abren ninguna sesión con modelo.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona. La medida del juez y el
  sondeo no los lanza el run.
- **Ningún umbral se cumple cambiándolo** (FR-050; ADR 0029): ninguna tarea ni corrección cumple un umbral de plan.md,
  «Controles de umbral», cambiando su valor, dejándolo en `decide: false`, sacando sesiones de su `total`, retirando su
  test, quitando un caso de una tabla o cambiando las evals, la rúbrica, los casos o la medida. Cada control nuevo se ve
  en rojo antes de darlo por bueno: los casos de T001, por el orden de sus pasos, y con los mutantes de T002, que no
  quedan en el diff.
- **El instrumento no se cambia** (FR-032): las evals de `jurisprudencia` y sus repeticiones, la rúbrica, `clases.yaml`,
  los casos, la medida versionada, el modelo del juez y su versión de Claude Code. Ninguna tarea repite ni escribe la
  medida, ni escribe en la carpeta de la evidencia. El hito no añade ni quita ninguna sesión ni ningún voto.
- **Lo que sigue igual** (FR-030, FR-031, FR-033, FR-034): los trece umbrales de antes, el juicio de cada sesión
  (`juzgar.go`, `trazas.go`, sin tocar), la tasa de cada serie, el sondeo (`sondeo.go` y sus tests, sin tocar),
  `boe-legislacion` y `legal-core`.
- **Ningún test escribe fuera de `t.TempDir()`**; lo temporal de una tarea va bajo `$TMPDIR`, nunca bajo `/tmp`.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint`, ningún `t.Skip`, TODO ni error
  silenciado; ningún test se desactiva ni se retira sin el que lo sustituye. Si `dupl`, `gocyclo`, `paralleltest` o
  `misspell` marcan algo, se reestructura o se renombra. El segundo valor de `invocacionesDe` no se mira, y no es un
  error silenciado: research D3 da por qué.
- **Antes de tocar una función, una constante o un ayudante que usan los tests**, `grep` de sus usos en todos los
  `_test.go` del paquete: lo que haya que alinear entra en la misma tarea, y sus ficheros están en sus rutas.
- **Proporcionalidad** (constitución, «Gates»; plan.md, «Trazabilidad», último párrafo): ninguna tarea añade un caso,
  un mensaje ni un test para un estado sin vía real —una skill con una eval que pide el cotejo y sin juez ni
  `sentencias` (research D4), el falso de `invocacionesDe` (research D3), o los casos límite del criterio de cotejo
  (código 2, `--dry-run`, cotejo sin `--roj`, varios cotejos), que ya prueban los tests del juicio (research D2, D8)—.
- **Lo que no cambia** (spec, «Fuera de alcance»; FR-032 a FR-034, FR-045, FR-052): el binario fuera del paquete de
  evals; `boe-legislacion`, `legal-core` y las evals de las tres skills; el directorio de esquemas; los fixtures; la
  carpeta de la evidencia; la definición del job y todo `.github/`; los guiones de `scripts/` y `.specify/`; la
  constitución; el roadmap; los ADR; `WORKFLOW.md`; la bitácora de uso; `HISTORIAL.md`; el léeme; la web; y los specs,
  los planes y los informes de otros hitos.

---

## Phase 1: Setup

No hay tareas: el repositorio, los objetivos de `make` y los gates existen desde H0, y el hito no añade ninguna
dependencia (plan.md, «Primary Dependencies»; research V2).

---

## Phase 2: Fundación

No hay tareas sin historia: lo único que el hito añade al código compartido es el umbral, que es ya de US1, y va en la
fase siguiente.

---

## Phase 3: User Story 1 y User Story 3 — El umbral `documento_sin_cotejar` y lo que ya decidía (P1, P2)

**Objetivo**: el informe del job de una skill con evals que piden el cotejo lleva, por cada modo, el umbral
`documento_sin_cotejar:<modelo>:<modo>` detrás de `cita_sin_documento`, que decide con 0 y cuyo motivo nombra cada
sesión; los trece de antes siguen como están (plan.md, paso 1).

**Prueba independiente**: quickstart.md §1, §2 (la primera orden) y §3.

- [X] T001 [US1] El umbral `documento_sin_cotejar:<modelo>:<modo>` en el informe del job, con los tests que fijan los quince umbrales de `jurisprudencia` y sus cuatro casos nuevos, en una sola tarea porque el código cambia la forma de `umbrales` y los tests que la esperan tienen que cambiar con él (contracts/umbral-documento-sin-cotejar.md §1 a §5; data-model §1 a §5; research D1 a D11, V3 a V9, V18, V19, V21, M2 a M5, M7; plan.md, paso 1 y «Trazabilidad»). **Orden**: primero los casos de `TestUmbralesDeJurisprudencia` —los cuatro nuevos y la lista de quince— y, con el código de producción de hoy, `go test -count=1 -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/` en rojo por una aserción (trece elementos y no quince, ningún `documento_sin_cotejar`, `aprobado` donde se espera `fallo`), anotado en la nota de la tarea; después el código. El subtest `piden-el-cotejo` llama a `pideElCotejo`, que no compila sin el código: se escribe con él, y se ve en rojo por una aserción quitando un momento la 09 de su lista esperada, también anotado en la nota. **`informe.go`**: (1) `pideElCotejo(eval)`, función sin exportar: `eval.Activa` y algún comando de `Comandos` con `Applet` `cita` y `Verbo` `cotejar`, sin ninguna lista de evals ni de skills (FR-001; research D5). (2) `comandoDelCotejo`, la variable del comando esperado `cita cotejar` sin ROJ, y `cotejaUnDocumento(sesion)`: alguna invocación de `invocacionesDe` —sobre un `ResultadoDeEval` de usar y tirar, sin mirar su segundo valor— satisface `comandoDelCotejo` con `satisface`, el criterio con el que el juicio da por ejecutado un comando esperado: la orden de la traza o la llamada a `cita_cotejar`, que consultó y terminó bien (FR-002; research D2, D3). (3) `respuestaDelJuez` gana `pideElCotejo` y `cotejo`, y `respuestasMedidas` gana `conCotejo`; `medirLosCotejos`, llamado desde `medirLasRespuestas` junto a `medirLasCitasSinDocumento`, los llena sobre las respuestas de los grupos de los modos de `respuestasQueSeJuzgan`, las del `total` de `sin_activar`, y pone `conCotejo` si alguna eval pide el cotejo; `medirLasRespuestas` sigue sin medir nada en una skill sin juez y sin `sentencias` (FR-001, FR-005; research D1, D4). (4) `grupoJuzgado.queLePidenElCotejo()` y `grupoJuzgado.sinCotejar()`, esta con su `if !respuesta.cotejo {` escrito así, porque de esa línea cuelga el mutante 1 de quickstart.md §4: cada sesión una vez como mucho, en orden de sesión. (5) Los comentarios de `EscribirInforme` y de lo que el cambio deja incompleto (research V21). **`umbrales.go`**: (6) `umbralDeDocumentoSinCotejar`, con su `Decide: true,` escrito así, porque de esa línea cuelga el mutante 2; `nombreDelUmbralDeDocumentoSinCotejar`; y sus tres constantes: el prefijo `documento_sin_cotejar`, la descripción de contracts §2 y el umbral 0; `medida` las de `sinCotejar()`, `total` las de `queLePidenElCotejo()`, comparado como los demás, 0 si `total` es 0 (FR-001, FR-004; research D6). (7) `respuestasMedidas.umbrales` lo añade, si `conCotejo`, detrás de `cita_sin_documento` de cada modo, de modo que `jurisprudencia` da los quince de contracts §1 en su orden, once que deciden, y los trece de antes no cambian ni de nombre, ni de valores, ni de orden entre sí (FR-006, FR-030). (8) `sinCotejarDelUmbral`, llamada desde `nombradasEn` como `sinDocumentoDelUmbral`: el motivo de contracts §3, cada sesión por su nombre, en orden de sesión, separadas por `; `, sin el modo detrás (FR-003; research D7). (9) Los comentarios que dicen cuántos umbrales lleva una skill y cuáles deciden (`umbrales.go:218` y los de research V21). **`doc.go`**, solo comentario: lo que dice de los umbrales de cada skill (`doc.go:183`): `jurisprudencia`, quince y once que deciden, con los dos `documento_sin_cotejar`; `boe-legislacion`, trece y nueve. Nada exportado cambia: ningún tipo, función ni campo exportado nuevo, ningún import nuevo, ningún `fmt.Print*` (plan.md, R5). **Tests**, en `informe_test.go`: `TestUmbralesDeJurisprudencia` y su comentario de cabecera dicen y exigen los quince, once que deciden, con sus valores escritos a mano; su arnés gana `escribirSinElCotejo` —borra la sesión y la escribe de nuevo con `escribirSesionDeJurisprudencia` y la sesión modelo de su eval sin la operación `cotejar`, ni la orden en la traza ni la llamada en el transcript, y se aplica antes que los demás cambios del caso (research D9)—, `umbralDocumentoSinCotejar`, los campos `sinCotejo` y `sinMedir` del caso, `sinCotejar` y `sinMedir` de `medidasDeJurisprudencia`, la constante de las 12 respuestas y la sesión sin medir en `mensajesDeLosVotosDeJurisprudencia`, que no se vota (research D10); `ninguna-cuenta` exige los dos `documento_sin_cotejar` a 0 de 12 con las 24 sesiones que cotejan por la traza y por la llamada (SC-003); los otros nueve casos de antes y `el-elemento-del-contrato` siguen dando lo que dan, ahora con los quince y los dos nuevos a 0 de 12; y los cuatro casos nuevos de contracts §5, con sus umbrales, motivos y veredicto escritos a mano: `una-sin-cotejar-en-el-modo-orden` (la 06, modo orden, 01, sin cotejo, y la misma de `claude-haiku-4-5-20251001`: 1 de 12, `cumple: false`, el motivo de contracts §3 y `fallo`; Haiku fuera del `total` y de la `medida`), `una-sin-cotejar-en-el-modo-herramienta` (la 06, modo herramienta, 01, sin su llamada a `cita_cotejar`: 1 de 12 y su motivo), `una-sin-activar-y-sin-cotejar` (la 06, modo orden, 02, sin cotejo y sin activación: 1 de 12 y `sin_activar` 1 de 30, los dos motivos) y `una-sin-medir-y-sin-cotejar` (la 06, modo herramienta, 01, sin cotejo y sin medir por el límite de uso, con `cambiarElFinalDeLaSesion` y el mensaje de `TestInformeConSesionesSinMedir`: 0 de 11, `cumple: true`, los demás de ese modo sobre 29, el conjunto sobre 59 y un solo motivo, el de la ejecución, que la nombra). En los cuatro, las 18 sesiones por modo de las evals que no piden el cotejo quedan fuera del `total` (FR-001, FR-002, FR-003, FR-004, FR-005, FR-062; SC-002, SC-003, SC-004; US3, escenario 3). `TestInformeFinalDeLosUmbrales/como-en-el-cierre-de-h26` lee los trece de H26 con el ayudante nuevo `umbralesDeJurisprudenciaComoEnH26`, que quita de los quince los dos `documento_sin_cotejar`, y sigue esperando lo mismo (FR-030; research D11). En `conjunto_test.go`: `TestEvalsDelRepositorio` gana el subtest `piden-el-cotejo`, con las evals de las tres skills del repositorio: piden el cotejo la 04, la 06, la 08 y la 09 de `jurisprudencia`, y ninguna de `boe-legislacion` ni de `legal-core` (FR-001, FR-007, FR-063; SC-005). No cambian lo que esperan, y siguen en verde: `TestUmbralesDelInforme` (los trece de `boe-legislacion`, nueve que deciden), `TestInformeEnDosModos` (`legal-core` con `[]`), `TestUmbralesConJuezYConSentencias`, `TestUmbralConjunto`, `TestInforme`, `TestInformeConElJuez`, `TestInformeMarkdownDeLosUmbrales`, `TestInformeConSesionesSinMedir`, los del juicio, `TestCopiasDelJuez`, `TestMedidaVersionada`, `TestDefinicionDelJob` y los del sondeo (FR-007, FR-030 a FR-033). Antes de tocar nada, `grep` de `umbralesDeJurisprudencia`, `medidasDeJurisprudencia`, `mensajesDeLosVotosDeJurisprudencia`, `escribirSesionDeJurisprudencia` y `nombradasEn` en los `_test.go` del paquete. **Verificación**: quickstart.md §1, §2 (primera orden) y §3, leídos con `rtk proxy`; `make ci`; y, con `grep` en los ficheros Go del paquete, ninguna frase dice ya que `jurisprudencia` lleva trece umbrales o que deciden nueve. — FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-030, FR-031, FR-032, FR-033, FR-060, FR-062, FR-063; SC-002, SC-003, SC-004, SC-005. Rutas: internal/evals/informe.go, internal/evals/umbrales.go, internal/evals/doc.go, internal/evals/informe_test.go, internal/evals/conjunto_test.go

---

## Phase 4: User Story 1 (control) — Los dos mutantes del umbral (P1)

**Objetivo**: la prueba de que el control de FR-001 falla por encima de su umbral: si cuenta solo cuando ninguna sesión
del modo cotejó, o si deja de decidir, falla algún caso de `TestUmbralesDeJurisprudencia` (plan.md, paso 2).

**Prueba independiente**: quickstart.md §4.

- [X] T002 [US1] Los dos mutantes del umbral, aplicados uno cada vez en una copia de la cabeza fuera del árbol, y su registro (quickstart.md §4; contracts/umbral-documento-sin-cotejar.md §6; research D12, M8; plan.md, paso 2). Es la prueba de que el control de FR-001 falla por encima de su umbral (ADR 0029), sobre el código y los tests que deja T001: no cambia ninguna línea de código ni de test, ni nada bajo `internal/`. Con un worktree de la cabeza en el directorio temporal, que es `$TMPDIR`, y las órdenes de quickstart.md §4 en primer plano y en su orden: **mutante 1**, en `sinCotejar`, `if !respuesta.cotejo {` pasa a contar solo si ninguna sesión del modo cotejó; **mutante 2**, en `umbralDeDocumentoSinCotejar`, `Decide: true,` pasa a `Decide: false,`. Tras cada sustitución, la guarda de §4 confirma que el fichero de la copia ha cambiado; si dice que el mutante no se ha aplicado, se ajusta la sustitución al código —el mutante es el de contracts §6— y el registro lleva la orden usada. Con cada uno se ejecuta en la copia `TestUmbralesDeJurisprudencia`, que tiene que terminar en rojo, y se restaura el fichero antes del siguiente. specs/022-h23-1-jurisprudencia-coteja-el/mutantes.md recoge, por mutante: la línea que cambia, antes y después; la orden ejecutada; y los subtests que fallan, copiados de la salida leída con `rtk proxy`. Con el mutante 1 tienen que estar entre ellos `una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta` y `una-sin-activar-y-sin-cotejar`, que esperan 1 de 12 y reciben 0; y con el mutante 2, los catorce casos de la tabla: 2 de 2 mutantes con algún caso en rojo (SC-006). Termina con el worktree retirado (`git worktree remove --force`) y `git status --short` igual que antes de empezar, salvo el directorio del feature: ningún mutante ha estado en el árbol. Si con un mutante no falla ningún caso, el control no basta y la tarea no se marca: lo deja escrito en su nota y amplía esta línea con el `_test.go` del caso que falta, que el intento siguiente relee. **Verificación**: lo anterior y `make ci`. — FR-001, FR-062, FR-064; SC-006. Rutas: specs/022-h23-1-jurisprudencia-coteja-el/mutantes.md

---

## Phase 5: User Story 2 — La skill coteja el documento antes de decir nada de él (P1)

**Objetivo**: `jurisprudencia` v0.1.1: con un documento en la conversación, lo primero es cotejar su ficha, nada se
dice de él antes, y si es la pedida lo dice `es_la_pedida`, nunca una comparación a ojo (plan.md, paso 3).

**Prueba independiente**: quickstart.md §5 y `make skills-check`.

- [X] T003 [US2] Los pasos 1 y 3 del «Protocolo» de `jurisprudencia`, y nada más de la skill (contracts/skill-jurisprudencia.md §1 a §3; research D13, V15, V16; plan.md, paso 3). **Paso 1** («Qué sentencias hacen falta, y cuáles están delante»): su última viñeta (hoy `SKILL.md:48`) dice que con un documento en la conversación lo primero es el paso 3, antes de decir nada de él; las del Tribunal Constitucional siguen yendo por el paso 7. **Paso 3** («Con el documento delante: cotejar su ficha»), al principio, antes de cómo se pasa la ficha (hoy `:83-89`), tres instrucciones: (a) el cotejo es lo primero que se hace con un documento que está en la conversación —texto pegado o PDF leído—: su ficha va a `cita_cotejar`, o a su orden, antes de escribir ninguna frase sobre él (FR-020); (b) nada del documento antes del cotejo: ni que es la sentencia pedida, ni que coincide con ella, ni de qué trata (FR-021); (c) si es o no la pedida lo dice `es_la_pedida`, lo que devuelve la operación, y nunca una comparación hecha a ojo entre el documento y la referencia, con el caso en que falla: 1088/2023 es el número de resolución del documento del ROJ STS 3144/2023, y el ROJ STS 1088/2023 es el de otra sentencia (FR-022). Lo que el paso 3 ya dice sigue igual y en su sitio —la ficha y nada más, de `Roj:` a `Tipo de Resolución:`; la referencia en la forma en que se pidió; la orden con la ficha por la entrada estándar o en `--documento`; el PDF; lo que devuelve; qué hacer si no es la pedida; `correspondencia`; el código 2—; si para decir (c) hace falta mover la frase que dice que devuelve `es_la_pedida`, se mueve sin cambiar lo que dice. No cambian: el frontmatter, con la misma `description`, sin campo de versión; el título, la entrada y la introducción del «Protocolo» (líneas 1 a 38); el paso 2; los pasos 4 a 7; «Comandos», con su tabla generada, y «Reglas» (FR-023, FR-024). Ningún bloque de código nuevo con una orden que no esté ya en la skill, y ninguna herramienta fuera de `cita_preparar` y `cita_cotejar`. Ningún otro fichero de la skill ni de las otras dos cambia, y ninguna eval (FR-032, FR-034). **Verificación**: quickstart.md §5 —menos de 300 líneas; las tres órdenes `cmp` sin salida; el diff, solo en los pasos 1 y 3—, `make skills-check` y `make ci`. — FR-020, FR-021, FR-022, FR-023, FR-024, FR-032, FR-034; SC-007. Rutas: skills/jurisprudencia/SKILL.md

---

## Phase 6: User Story 4 — Los documentos dicen los umbrales que el job aplica (P3)

**Objetivo**: `CHANGELOG.md` (*Unreleased*), `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md` y `CLAUDE.md` dicen el
umbral nuevo y la skill v0.1.1; `WORKFLOW.md` no cambia (plan.md, paso 4).

**Prueba independiente**: quickstart.md §6.

- [X] T004 [US4] Lo que los cuatro documentos dicen del umbral y de la skill, con contracts/documentos.md, cada cifra, fichero, test y umbral sacados del árbol (research D15, V22, V23; plan.md, paso 4). **`CHANGELOG.md`**, dentro de `## [Unreleased]` y en ninguna versión publicada (FR-040, FR-065): en «Añadido», el umbral `documento_sin_cotejar:<modelo>:<modo>` —qué cuenta, sobre qué (12 por modo hoy), con 0 y decide, su motivo con cada sesión; `jurisprudencia` de trece elementos, nueve que deciden, a quince, once que deciden; `boe-legislacion` sigue con trece y `legal-core` con `[]`—; y en «Corregido», `jurisprudencia` v0.1.1, que sustituye a la v0.1, con las cifras del defecto (6 de 48 respuestas de `claude-sonnet-5-5` en la eval 06, 4 en el modo orden y 2 en el herramienta, según la bitácora de uso del 2026-10-15) y lo que no cambia, con las líneas de `SKILL.md` que deja T003. **`CONTRIBUTING.md`**, «Job de evals» (FR-041): «otros trece» pasa a quince; «son también trece, y deciden nueve» pasa a quince y once; la tabla de `jurisprudencia` gana la fila de `documento_sin_cotejar:<modelo>:<modo>` detrás de la de `cita_sin_documento:<modelo>:<modo>`; «los cuatro del modo orden y los cuatro del modo herramienta» y «cuatro de los trece» pasan a los cinco y los cinco y a cuatro de los quince; y un párrafo nuevo, detrás del de `cita_sin_documento`, con lo de contracts/documentos.md: un hecho de la sesión, medido sin modelo con las invocaciones de su traza y sus llamadas, con el criterio del comando esperado; toda skill con una eval que pide el cotejo, sin que el código nombre ninguna (`TestUmbralesDeJurisprudencia`, `TestEvalsDelRepositorio`); y su motivo. No cambian la frase de que existen solo con juez o con `sentencias`, lo de `boe-legislacion` y `legal-core`, ni el párrafo de `cita_sin_documento`. **`docs/JURISPRUDENCIA.md`** (FR-042): donde dice que el job decide con `cita_sin_documento`, dice que decide con dos hechos de la sesión, `cita_sin_documento` y `documento_sin_cotejar`. **`CLAUDE.md`** (FR-043): la enumeración de lo que comprueba un guion con umbral 0 nombra también el cotejo del documento (`documento_sin_cotejar`); la viñeta de H23.1 y la frase del siguiente hito no se tocan. **`WORKFLOW.md`** se lee y no cambia: ninguna frase dice cuántos umbrales lleva `jurisprudencia` (FR-044). Ninguna otra frase se reescribe (FR-045). **Verificación**: quickstart.md §6, orden a orden y con lo esperado de cada una; y `make ci`. — FR-040, FR-041, FR-042, FR-043, FR-044, FR-045, FR-065; SC-008. Rutas: CHANGELOG.md, CONTRIBUTING.md, docs/JURISPRUDENCIA.md, CLAUDE.md

---

## Phase 7: Cierre — Definition of Done

- [X] T005 Cierre de la Definition of Done, sin código de producto (`ROADMAP.md` §1; quickstart.md §0 a §3, §5 a §7): specs/022-h23-1-jurisprudencia-coteja-el/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de lo creado o modificado —los cinco ficheros Go del paquete de evals, `SKILL.md` de `jurisprudencia`, los cuatro documentos y los artefactos del feature— y la comprobación, con las órdenes de quickstart.md §3, de que no cambia nada de lo que el hito deja como está: `sondeo.go`, los guiones, `boe-legislacion` y `legal-core`, las evals, el directorio de esquemas, la carpeta de la evidencia, los fixtures, `.github/`, `.specify/`, el binario fuera del paquete de evals, el roadmap, los ADR, la bitácora de uso, `HISTORIAL.md`, `WORKFLOW.md`, el léeme, la web y los artefactos de otros hitos (FR-032, FR-033, FR-034, FR-045, FR-052); (2) el resultado de quickstart.md §0 a §3, §5, §6 y §7, orden a orden, con lo esperado de cada una y leído con `rtk proxy`, y la constancia de que §4 lo ejecutó T002, de que §8 lo hace el workflow y de que §9 queda para después del run; (3) la cobertura sobre los perfiles que deja `make ci`, con `go tool cover -func` —global ≥ 70 % y el dominio ≥ 85 % (Definition of Done §1.9)— y la de las funciones nuevas del paquete de evals; (4) la tabla de plan.md, «Controles de umbral», con el test de cada fila `ci:` encontrado como `^func <Test>\(` en su ruta, que es como lo busca el informe final, los objetivos de `make` de sus filas, y el nombre de cada una de las once filas `evals:` cotejado, carácter a carácter, con los quince de contracts/umbral-documento-sin-cotejar.md §1, con `claude-sonnet-5-5` como modelo (FR-061; SC-009); y (5) la constancia, cada una con su orden, de que ningún umbral se cumplió cambiando su valor, dejándolo en `decide: false` o sacando sesiones de su total; de que las evals, la rúbrica, los casos y la medida de la carpeta del juez son los de `main`; y de que el diff del hito no añade ningún `//nolint`, `t.Skip` ni TODO, no abrió ninguna sesión con modelo y no lanzó la medida ni un sondeo (FR-050, FR-051, FR-060). Solo si una cifra de cobertura queda bajo su umbral se añaden los tests que faltan, en el `_test.go` de un fichero declarado. **Verificación**: `make ci`. — FR-032, FR-033, FR-034, FR-045, FR-050, FR-051, FR-052, FR-060, FR-061; SC-009. Rutas: specs/022-h23-1-jurisprudencia-coteja-el/cierre.md, internal/evals/informe_test.go, internal/evals/umbrales_test.go

---

## Dependencias y orden de ejecución

El orden es el de plan.md, «Orden de implementación (de dentro afuera)», más el cierre, y es estrictamente secuencial:
T001 → T005. Ninguna tarea depende de una posterior.

| Tarea | Paso del plan | Necesita |
|---|---|---|
| T001 | 1 | — |
| T002 | 2 | T001 (`sinCotejar` y `umbralDeDocumentoSinCotejar`, que es lo que se muta, y los casos que ven cada mutante) |
| T003 | 3 | — (su efecto lo mide el umbral de T001 en el job de cierre) |
| T004 | 4 | T001 (el umbral y los tests que nombra existen en el árbol) y T003 (las líneas de `SKILL.md`) |
| T005 | — | T001 a T004 |

### Historias

- **US1** (T001, T002) es el control: el umbral que hace salir el job en rojo si una sesión no coteja, con la prueba de
  que falla. **US2** (T003) es la skill que el umbral protege; su efecto lo mide el job de cierre (SC-001). **US3**
  (T001) es lo que el cambio no arrastra: los trece de antes, el juicio de cada sesión, `boe-legislacion` y
  `legal-core`. **US4** (T004) son los documentos.
- Las historias son independientes en su prueba: cada fase nombra el escenario del quickstart que la comprueba sola.
  US1 y US3 comparten `informe.go`, `umbrales.go` y los mismos tests, y por eso van en la misma tarea.

### Oportunidades de paralelismo

Ninguna dentro del run: el workflow ejecuta las tareas una a una. Ninguna tarea lleva `[P]`. Fuera de un run, T003 no
depende de T001 ni de T002 y podría ir en paralelo con ellas.

---

## Trazabilidad

### Requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001 | T001 (`pideElCotejo`, el `total`, el elemento por modo), T002 (los mutantes) |
| FR-002 | T001 (`cotejaUnDocumento` con `satisface`; `ninguna-cuenta` por la traza y por la llamada) |
| FR-003 | T001 (el motivo y el `fallo`) |
| FR-004 | T001 (la comparación, 0 si `total` es 0) |
| FR-005 | T001 (`una-sin-medir-y-sin-cotejar`) |
| FR-006 | T001 (los quince, once que deciden, en su orden) |
| FR-007 | T001 (`TestUmbralesDelInforme`, `TestInformeEnDosModos`, sin cambiar; `piden-el-cotejo`) |
| FR-020, FR-021, FR-022 | T003 |
| FR-023, FR-024 | T003 (`make skills-check`, quickstart.md §5) |
| FR-030 | T001 (los trece de antes, sin cambiar; `umbralesDeJurisprudenciaComoEnH26`) |
| FR-031 | T001 (los tests del juicio y de las series, sin cambiar); restricción de todas (batería) |
| FR-032 | Restricción de todas (batería); T001 (`TestCopiasDelJuez`, `TestDefinicionDelJob`, sin cambiar); T005 lo deja constatado |
| FR-033 | T001 (los tests del sondeo, sin cambiar); T005 |
| FR-034 | T003 (solo `SKILL.md` de `jurisprudencia`); T005 |
| FR-040, FR-065 | T004 (`CHANGELOG.md`) |
| FR-041, FR-042, FR-043, FR-044, FR-045 | T004; T005 constata FR-045 |
| FR-050 | Restricción de todas (batería); T005. Rige además las reparaciones del cierre, que hace el workflow y no una tarea |
| FR-051 | Restricción de todas (batería); T005 |
| FR-052 | Restricción de todas (rutas declaradas); T005 |
| FR-060 | Todas (`make ci` por tarea); T005 |
| FR-061 | plan.md, «Controles de umbral»; tabla siguiente; T005 |
| FR-062 | T001 (los cuatro casos nuevos y `ninguna-cuenta`) |
| FR-063 | T001 (los quince en `TestUmbralesDeJurisprudencia`; `piden-el-cotejo`) |
| FR-064 | T002 |
| FR-066 | Lo hace el workflow tras la revisión final, no una tarea; T001 construye el umbral que mide |
| SC-001 | Lo mide el cierre del workflow, que no es una tarea; T001 construye los dos umbrales que lo hacen fallar y T003 la skill que tiene que cumplirlos |
| SC-002, SC-003, SC-004, SC-005 | T001 |
| SC-006 | T002 |
| SC-007 | T003 |
| SC-008 | T004; lo comprueba la revisión final por lectura |
| SC-009 | Lo decide sin modelo el informe final del workflow; T001 construye los dos umbrales y T005 coteja el nombre de cada fila |
| SC-010 | Después del run, una persona; no es una tarea |

### Controles de umbral (plan.md) → tarea que construye el control

| Fila de plan.md | Control | Tarea | Test que lo ve fallar |
|---|---|---|---|
| FR-001 a FR-004, SC-001, SC-009 (modo orden) | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden` | T001 (el umbral), T002 (los mutantes) | `TestUmbralesDeJurisprudencia`: `una-sin-cotejar-en-el-modo-orden` y `una-sin-activar-y-sin-cotejar` dan `fallo` |
| FR-001 a FR-004, SC-001, SC-009 (modo herramienta) | `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta` | T001, T002 | `TestUmbralesDeJurisprudencia`: `una-sin-cotejar-en-el-modo-herramienta` da `fallo` |
| FR-030, SC-001 (`sin_activar`, los dos modos) | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:<modo>` | El de hoy; T001 lo deja entre los quince | `TestUmbralesDeJurisprudencia`: `una-sin-activar` y `una-sin-activar-y-sin-cotejar` |
| FR-030, SC-001 (`cita_sin_documento`, los dos modos) | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:<modo>` | El de hoy; T001 | `TestUmbralesDeJurisprudencia`: `una-sin-documento-en-el-modo-orden` y `dos-sin-documento-en-el-modo-herramienta` |
| FR-030, SC-001 (la clase del juez) | `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` | El de hoy; T001 | `TestUmbralesDeJurisprudencia`: `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo` |
| FR-030, FR-032, SC-001 (defectos y correctos) | `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:…` | El de hoy; ninguna tarea lo cambia | `TestMedidaVersionada`, sin cambiar: cada recuento distinto de 0 |
| FR-030, SC-001 (juez, los dos modos) | `evals:jurisprudencia:duracion_del_juez:<modo>` | El de hoy; T001 | `TestUmbralesDeJurisprudencia`: `901-s-del-juez-en-el-modo-orden` |
| FR-005, FR-062, SC-002, SC-003, SC-004 | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` | T001 | Sus cuatro casos nuevos, en rojo con el código de hoy (T001, primer paso) y con cada mutante (T002) |
| FR-006, FR-063, SC-005 (`jurisprudencia`) | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia` | T001 | La lista exacta de quince en cada caso: en rojo con el código de hoy (trece) y con el mutante 2 (nueve que deciden) |
| FR-007, FR-063, SC-005 (`boe-legislacion` y `legal-core`) | `TestUmbralesDelInforme`, `TestInformeEnDosModos`, `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` | Los dos primeros, los de hoy; T001 añade `piden-el-cotejo` | Los trece escritos a mano y `[]`: un elemento de más los rompe; `piden-el-cotejo`, con una eval de más o de menos |
| FR-064, SC-006 | `TestUmbralesDeJurisprudencia` | T002 (los tests, T001) | Los dos mutantes, en la copia: 2 de 2 |
| FR-023, SC-007 | `ci:Makefile:skills-check` | El de hoy; T003 lo deja en verde | `TestSkillsDelRepositorio` con 300 líneas o una tabla distinta de la regenerada |
| FR-030, FR-031 | `TestUmbralesConJuezYConSentencias`, `TestInformeConElJuez`, `TestJuzgar`, `TestInforme` | Los de hoy; T001 los deja sin cambiar | Sus casos de H24 a H26 |
| FR-032 | `TestCopiasDelJuez`, `TestDefinicionDelJob`, `ci:Makefile:schema-check` | Los de hoy; ninguna tarea los cambia | Sus casos de H24 y de H25 |
| FR-033 | `ci:internal/evals/sondeo_test.go:TestSalidaDelSondeo` | El de hoy; ninguna tarea lo cambia | Sus líneas esperadas |
| FR-060 | `ci:Makefile:ci` | Todas | El objetivo entero |

### Definition of Done (`ROADMAP.md` §1) → tareas

| Punto | Aplica | Tareas |
|---|---|---|
| 1 · `make ci` en verde | Sí | Todas; T005 |
| 2 · Tests offline; fixtures si toca red | Sí; no toca red | T001; ningún fixture grabado |
| 3 · Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | T001 (ningún import nuevo; el elemento y el motivo van al informe) |
| 4 · Esquemas y `schema-check` | Ningún esquema cambia; `schema-check` sigue en verde | T001 lo ejecuta (quickstart.md §2); T005 lo constata |
| 5 · Errores tipados y códigos estables | Ningún código de `kitlegal` cambia | — |
| 6 · e2e y `CHANGELOG.md` | «Aceptación e2e: no aplica»; `CHANGELOG.md`, sí | T004 |
| 7 · ADR | No: la decisión cabe en los ADR 0029, 0036 y 0038, ya aceptados (plan.md, «Summary») | — |
| 8 · `SOURCES.md` | No: ninguna fuente nueva ni cambiada (research D16) | — |
| 9 · Cobertura | Sí | T005 |
| 10 · Skill: evals antes, `SKILL.md` < 300, sin drift | Sí: las evals que piden el cotejo ya existen y FR-032 no deja cambiarlas; las mide el job de cierre | T003 (`make skills-check`); T005 lo constata |
| 11 · Dimensión territorial | No | — |
| 12 · Operaciones de grafo | No: ningún applet cambia | — |
| 13 · Recursos, plazos o escritos | No | — |

---

## Estrategia de implementación

- **De dentro afuera** (plan.md): el umbral con los tests puestos a su forma nueva; los mutantes; la skill; los
  documentos; y el cierre.
- **Lo mínimo que ya da valor** es T001: un job que sale en rojo cuando una sesión da por bueno un documento sin
  cotejarlo. No es una entrega parcial: el run ejecuta las cinco, y SC-001 pide el cierre entero, con la skill de T003
  cumpliendo el umbral.
- **Lo que el run no puede medir** (research S1 a S4): qué dará el job de cierre con la skill v0.1.1. Si el cierre sale
  en rojo por `documento_sin_cotejar`, ninguna reparación cambia el umbral, lo deja en `decide: false`, saca sesiones de
  su `total` ni toca las evals (FR-050): queda en el informe final, con las sesiones que nombra su motivo.
- **Después del run** (SC-010; quickstart.md §9): una persona lanza el sondeo de la eval 06 con la skill del hito y
  decide la fusión; escribe la viñeta de H23.1 de `CLAUDE.md`; y corrige, fuera del hito, el comentario de la
  definición del job (research V23).

---

## Comprobación contra la rúbrica de `juez_tasks` y `precheck.sh tasks`

- `precheck.sh tasks`: el guion no se pudo ejecutar en esta sesión (el sandbox pide aprobación para lanzarlo); sus
  comprobaciones de la fase tasks se reprodujeron con `grep` y sus mismos patrones sobre las líneas de tarea, sin
  ningún hallazgo: cinco líneas con el formato `- [ ] Tnnn`, ids únicos y correlativos; todas con rutas; ninguna
  nombra el directorio de esquemas, uno de fixtures ni la carpeta de la evidencia por su ruta, ni la variable de
  grabación, ni una acción de plataforma, ni pide nada a una persona a mitad; ninguna tarea de aceptación, como pide
  «Aceptación e2e: no aplica». El único defecto que puede dar antes del análisis es que falta `gates/analyze.md`, que
  escribe el paso siguiente.
- a · `analisis_critico`: lo escribe el paso de análisis después de este fichero.
- b · `trazabilidad`: cada tarea cita sus FR y SC; «Requisitos → tareas» cubre FR-001 a FR-066 y SC-001 a SC-010; y
  «Controles de umbral» da la tarea y el test de cada fila de plan.md.
- c · `rebanadas_verdes`: T001 trae sus tests y su implementación; las excepciones están declaradas arriba; el orden es
  el del plan y la tabla de dependencias no tiene ninguna hacia delante.
- d · `rutas_declaradas`: tras «Rutas:», ficheros concretos; ninguna ruta genérica ni ningún directorio. T001 declara los
  cinco ficheros que nombra el plan y ningún otro de `internal/`; T002, solo `mutantes.md`; T003, solo el `SKILL.md`; T004,
  los cuatro documentos.
- e · `datos_separados`: ninguna tarea toca el directorio de esquemas ni uno de fixtures; ningún dato externo ni
  grabación.
- f · `dod`: tabla «Definition of Done → tareas»; `CHANGELOG.md` en T004; sin ADR ni `SOURCES.md`, con su motivo; los
  esquemas y los fixtures no cambian.
- g · `checklist_veraz`: `checklists/requirements.md` no cambia, y no tiene ningún ítem sin marcar.
- h · `aceptacion_primero`: el plan justifica «no aplica»; las evals de la skill ya existen y FR-032 no deja cambiarlas.
- i · `autonomia`: ninguna tarea exige a una persona, publica, ni mide en la plataforma.
