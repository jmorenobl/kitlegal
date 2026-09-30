# Tasks: H7.4 · `boe-legislacion` sin el estado de la comprobación ni lo que no ha leído, y un job que juzga la respuesta a la pregunta

**Input**: `specs/014-h7-4-boe-legislacion-sin/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.8.0 aplicada (umbral de materialidad,
criterio de uso, proporcionalidad y convergencia en «Gates»; controles de umbral del ADR 0029; ADR 0030); H0-H7, H7.1,
H7.2, H7.3 y H19 en `main` (`internal/evals` con el formato común, la lista de tres familias, `umbrales` con el contrato
del ADR 0029, el job en paralelo con su definición comprobada en `make ci` y el sondeo; `boe-legislacion` v0.1.3; el
modelo que decide, `claude-sonnet-5-5`, ADR 0031). **Ninguna tarea crea el esqueleto ni los gates: `make ci` existe y
pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: el plan dice **«Aceptación e2e: no aplica»** (plan.md, «Aceptación e2e»; research D20): el hito no
cambia el binario —cambiarlo está fuera de alcance y `cmd/kitlegal` no enlaza el paquete de evals—, así que no hay
comportamiento de `kitlegal` que un guion `testscript` describa, y **no hay tarea `[aceptacion]`**. El hito cambia una
skill, así que sus evals van antes que `SKILL.md` (Definition of Done §1.10): la lista y el formato de eval (T001-T004)
y la eval 20 (T009) preceden a `boe-legislacion` v0.1.4 (T010). La aceptación del hito es la del job de evals que el
workflow lanza en el cierre tras la revisión final (SC-001, FR-102). En `make ci` la fijan, sin modelo, los tests que
traen las tareas (plan.md, «Tests nuevos», «Tests existentes que cambian» y «Controles de umbral»).

**Tests**: obligatorios (constitución §III; spec FR-091 a FR-100; plan.md, «Controles mecánicos»). Cada tarea de código
trae su test y la implementación mínima que lo hace pasar, en el mismo diff; los nombres de test son los de plan.md.

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —su test y su implementación en el mismo
diff, `make ci` en verde por sí solas— T003, T004, T006, T007, T008 y T011. Las demás, cada una por su razón y con su
verificación:

- **T001, T002, T005 y T009** son las `[datos]`: tocan `schemas/` o un `testdata/`, y llevan además solo lo que el dato
  obliga a cambiar a la vez (plan.md, *Complexity Tracking*, segunda fila; precedente, H7.2 T001 y H7.3 T001). T001 y
  T002: el esquema, el campo del tipo que lo lee (el lector estricto rechaza una clave que el tipo no tiene) y los
  ficheros y casos que lo cumplen; ninguna lógica. T005: el transcript sintético del caso de `TestLeerSesion` y la
  lectura del primer `result` —el caso nuevo con la lectura de hoy deja `make ci` en rojo, y la lectura nueva sin el caso
  no tiene test que la vea—. T009: la eval 20, sus dos derivadas escritas por código (`TestGrabacionesDerivadas
  -actualizar-derivadas`) y las dos comprobaciones de test que las fijan (`derivadasDelGrafoPrevio` y
  `compruebaLaSesion`).
- **T010** es la skill (`SKILL.md` no es código con test propio) con la lista entera: trae en el mismo diff sus
  controles mecánicos —el calibrado sobre tres informes, `expresiones-de-la-skill`, `ordenes-para-powershell`,
  `TestOrdenesParaPowerShell` y `TestJuzgarLasClasesDeLaRespuesta`—, porque con la lista entera la prosa de v0.1.3 da 9
  defectos (research V19; plan.md, paso 6).
- **T012** es documentación: su verificación es que cada fichero, test, clave y campo que nombran `CHANGELOG.md`,
  `CONTRIBUTING.md`, el comentario del paquete de evals y los comentarios de la definición del job existe, y `make ci`.
- **T013** es el cierre de la Definition of Done, sin código de producto: escribe la lista de lo tocado, mide la
  cobertura y ejecuta quickstart.md salvo lo que abre una sesión con modelo (la tercera orden de §8) y lo que llega tras
  el run (§9); solo añade tests si una cifra de cobertura queda bajo su umbral.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea **todas** las rutas que crea, modifica o retira, y solo toca esas (más `go.mod`,
`go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). El cierre en la
plataforma —publicar la rama, abrir la propuesta de cambio y medir CI y el job de evals de SC-001 y FR-102— lo hace el
workflow tras la revisión final; la tercera orden de quickstart.md §8 (el sondeo con modelo) queda para quien lea el
informe final.

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC; rutas`

- **[aceptacion]**: ninguna (plan.md, «Aceptación e2e: no aplica»).
- **[datos]**: la tarea crea o modifica ficheros bajo `schemas/` o bajo un `testdata/`. Son cuatro: T001 (el esquema de
  la lista), T002 (el esquema de eval), T005 (el caso de lectura de sesión) y T009 (las derivadas del grafo previo de la
  eval 20). H7.4 no graba nada de ninguna fuente (plan.md, «Datos externos»; research D21): las respuestas grabadas de
  `a1-30` y `da-3` de `BOE-A-2017-12902` están desde H4 en la unión de grabaciones de las evals; no hay manifiesto
  `grabaciones.json` nuevo, ni test de grabación, ni nada que grabar para el paso `grabar_datos`; las derivadas las
  escribe código a partir de lo grabado, y el transcript de T005 es sintético (no es un dato externo). Los tres informes
  versionados del calibrado solo se leen (FR-090).
- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia (ver «Oportunidades de paralelismo»).
- **[Story]**: US1 quien pregunta lee la norma, sin el estado de la comprobación ni una redacción que nadie leyó · US2
  la skill se activa ante toda pregunta de su ámbito · US3 dos preceptos cambiados, dos líneas que dicen cuál es cada uno
  · US4 los umbrales nuevos los hace cumplir el job · US5 el job juzga la respuesta a la pregunta · US6 una tanda de
  sesiones por commit y skill · US7 quien lanza el sondeo lee su error de uso, no una traza. La documentación y el cierre
  no llevan historia.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`, `gocyclo`…, también sobre los ficheros con las
  etiquetas `evals` e `integration`) · tests con `-race` · `test-integration` · `test-tiempos` · `vuln` · `schema-check`
  · `skills-check` · secretos · módulos. Ningún registro de la verificación en la raíz del repositorio. Las salidas de
  `go test` y de `make` se leen con `rtk proxy` o con sondas positivas, porque el proxy de la sesión las resume.
- **Sin red ni modelo**: ningún test abre una conexión ni una sesión con modelo; `claude`, `strace`, `go` y las
  respuestas de `gh` son sustitutos o JSON sintéticos escritos por los tests en `t.TempDir()`. Ninguna tarea graba,
  ejecuta `make evals`, `scripts/evals.sh`, el sondeo con modelo, `TestTandaDelCommit` (consulta `gh`) ni `claude`.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona.
- **Ningún test escribe fuera de `t.TempDir()`**, salvo los temporales que `PrepararSesion`, sin cambios desde H5, crea y
  retira ella misma.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint` nuevo, ningún `t.Skip`, TODO ni error
  silenciado; ninguna exclusión de lint nueva. G204 se resuelve con las órdenes constantes de `gh` y los valores en el
  entorno (research D16). Si `dupl`, `gocyclo`, `paralleltest` o `contextcheck` marcan algo, se reestructura.
- **Los umbrales no se rebajan** (FR-047, ADR 0029): ninguna tarea ni corrección cumple un umbral de FR-040 a FR-043
  rebajándolo, retirándolo, dejándolo en `decide: false`, sacando evals del total o recortando la lista para que case
  menos; la regla por serie del ADR 0016 y H7.2 FR 054 no cambia (FR-062).
- **Proporcionalidad** (constitución, «Gates»): ninguna tarea añade un caso, un mensaje ni un test para un estado sin vía
  real; un transcript, una respuesta de `gh`, una definición o una lista que no se pueden leer siguen las reglas que ya
  hay (sesión ilegible, fichero mal formado, test que falla, error con código 1), sin caso propio (plan.md,
  «Trazabilidad»).
- **El binario no cambia**: ningún fichero de `cmd/` ni de `internal/` fuera del paquete de evals, salvo
  `internal/app/grafo_test.go` (T009); ningún `os.Exit`, `fmt.Print*`, `os.Stdout`, `os.Stderr` ni `net/http` nuevo en Go
  (R2, R4, R5). Ni los artefactos de H7.1, H7.2 y H7.3 (FR-090) ni los guiones del workflow (FR-072) cambian.
- **Guiones de bash**: sin nada posterior a bash 3.2; comprobados con `bash -n`.
- **Vocabulario**: los comentarios llevan tilde; los caracteres que no son ASCII de los tests se escriben con su escape
  de Go (`\xc3\xb3`…), no con `\u` (las frases de la bitácora y las formas nuevas de la lista incluidas).
- **Sondas y mutantes momentáneos** (T003-T011): se crean y se retiran dentro de la tarea, antes de `make ci` y del
  commit; cualquier copia del repositorio va a un directorio temporal fuera del árbol, nunca dentro.

---

## Fase 1: US1 y US2 — La lista con sus dos clases y el formato de eval (P1; va primero por el orden de dentro afuera)

**Objetivo**: la lista gana la familia de la clase B y las formas fijas que se quitan antes de buscar; el formato de
eval gana `no_se_activan` y `redacciones_modificadas`; las tres evals de `legal-core` declaran que `boe-legislacion` no
se activa. Van antes que el juicio, los umbrales y la skill (Definition of Done §1.10). **Prueba independiente**: los
casos de los dos esquemas en `formato_test.go`, `TestLeerConjunto`, `TestFormatoDeLasEvalsDeCadaSkill`,
`TestExtraerExpresionesProhibidas` y `expresiones-calibradas` sin cambios en su tabla (quickstart.md §2).

- [X] T001 [datos] [US1] La familia de la clase B y las formas fijas en el esquema de la lista, en la lista y en el tipo que la lee, con los casos que los fijan, y **nada más** (contracts/lista-de-expresiones.md §1 y §2; data-model §1; research D6, D7; plan.md, «Orden de implementación», paso 1, y *Complexity Tracking*): schemas/expresiones-prohibidas.yaml.json: `required` pasa a `["maquinaria", "otra_conversacion", "anuncio", "redaccion_no_leida", "formas_fijas"]`; `redaccion_no_leida` con el mismo `$ref` a la definición común `expresiones` que las otras tres; `formas_fijas` con el tipo de contracts §2 carácter a carácter (lista de al menos una cadena de una línea con algún carácter fuera de los marcadores); `additionalProperties: false` se queda; evals/boe-legislacion/expresiones-prohibidas.yaml: el comentario de cabecera de contracts §1 (cuatro familias, las dos clases y las formas fijas que se quitan antes de buscar) y, detrás de `anuncio` y sin tocar sus 14 expresiones ni las 22 de `maquinaria` ni las 16 de `otra_conversacion`, `redaccion_no_leida` con solo su primera expresión, `ya no exige` (las otras nueve y las 25 nuevas de `anuncio` entran con la skill, en T010: con ellas, la prosa de v0.1.3 da 9 defectos, research V19), y `formas_fijas` con las dos formas definitivas de contracts §1 carácter a carácter (la línea `⚠ REDACCIÓN MODIFICADA: <cita>: …` con sus dos `<fecha>`, y la frase de la regla 7 sin su punto final); internal/evals/prohibidas.go: `ExpresionesProhibidas.RedaccionNoLeida []string` (`yaml:"redaccion_no_leida"`) y `ExpresionesProhibidas.FormasFijas []string` (`yaml:"formas_fijas"`), con su comentario —sin ellos, el lector estricto rechaza la lista que el esquema exige—; `ExtraerExpresionesProhibidas` y el recuento no cambian todavía (T003); tests: en internal/evals/formato_test.go, los casos de contracts §2 —la lista del repositorio valida; sin `redaccion_no_leida`, sin `formas_fijas`, con `formas_fijas: []` y con una forma de dos líneas no validan (una lista mal formada es un fichero mal formado, H7.2 FR 055)—; las listas sintéticas y los `ExpresionesProhibidas` literales que el esquema o el lector estricto exigen completos, buscados con `git grep -n anuncio` en los tests del paquete de evals —hoy en internal/evals/formato_test.go, internal/evals/conjunto_test.go, internal/evals/informe_test.go, internal/evals/juzgar_test.go e internal/evals/prohibidas_test.go—, ganan las dos claves, sin desactivar ningún caso; `TestLeerConjunto` (internal/evals/conjunto_test.go) con una lista de las cinco claves: las dos nuevas quedan en `Conjunto.Prohibidas` y en cada `Eval.Prohibidas`; verificación: `make ci` en verde con `TestEvalsDelRepositorio` sin cambios —la extracción aún no lee las claves nuevas, así que el calibrado sigue en 35 de 93 y 10 de 93— — FR-030, FR-031, FR-035; SC-002; US1-6; rutas: schemas/expresiones-prohibidas.yaml.json, evals/boe-legislacion/expresiones-prohibidas.yaml, internal/evals/prohibidas.go, internal/evals/formato_test.go, internal/evals/conjunto_test.go, internal/evals/informe_test.go, internal/evals/juzgar_test.go
- [X] T002 [datos] [US2] `no_se_activan` y `redacciones_modificadas` en el esquema de eval y en el tipo que la lee, y las tres evals de `legal-core` que declaran que `boe-legislacion` no se activa, con los casos que los fijan, y **nada más** (contracts/evals-y-juicio.md §1 y §5; data-model §2; research D9; plan.md paso 1 y *Complexity Tracking*): schemas/eval.yaml.json gana las propiedades `no_se_activan` y `redacciones_modificadas` y las definiciones `redaccion-modificada` y `fecha` de contracts §1 carácter a carácter; `no_se_activan` vale en toda eval y `redacciones_modificadas` entra en la lista del `else`, como `hallazgos` (una eval de no activación no la admite); internal/evals/formato.go: `Eval.NoSeActivan []string` (`yaml:"no_se_activan"`), `Eval.RedaccionesModificadas []RedaccionEsperada` (`yaml:"redacciones_modificadas"`) y el tipo `RedaccionEsperada` con `Norma`, `Bloque`, `FechaVigencia` y `FechaVigenciaReciente` (`norma`, `bloque`, `fecha_vigencia`, `fecha_vigencia_reciente`), con sus comentarios; `Juzgar` no los mira todavía (T004); evals/legal-core/01-territorio-municipio-cubierto.yaml, evals/legal-core/02-territorio-municipio-no-cubierto.yaml y evals/legal-core/03-no-activa-receta-de-cocina.yaml ganan, detrás de `activa`, `no_se_activan: [boe-legislacion]` y nada más (FR-004; sus preguntas, comandos y juicio no cambian, fuera de alcance); tests: en internal/evals/formato_test.go, los casos de contracts §1 —una eval con cada clave bien formada valida y se lee con sus campos; `no_se_activan: []`, un nombre que no es de skill, uno repetido, `redacciones_modificadas` en una eval de no activación, una sin `fecha_vigencia_reciente`, una fecha de siete cifras y una con el mes 13 no validan—; `TestLeerConjunto` (internal/evals/conjunto_test.go) lee `NoSeActivan` de una eval sintética de `legal-core` y `RedaccionesModificadas` de una positiva; los resultados y evals esperados enteros que lean las evals de `legal-core` del repositorio, en internal/evals/conjunto_test.go e internal/evals/juzgar_test.go, ganan el campo, sin desactivar ninguno; verificación: `TestFormatoDeLasEvalsDeCadaSkill` pasa con las tres de `legal-core` y su clave nueva; ningún juicio cambia; `make ci` en verde — FR-003, FR-004, FR-053; US2-3, US3-2; rutas: schemas/eval.yaml.json, internal/evals/formato.go, internal/evals/formato_test.go, internal/evals/conjunto_test.go, internal/evals/juzgar_test.go, evals/legal-core/01-territorio-municipio-cubierto.yaml, evals/legal-core/02-territorio-municipio-no-cubierto.yaml, evals/legal-core/03-no-activa-receta-de-cocina.yaml
- [X] T003 [US1] Las formas fijas se quitan antes de buscar y la familia de la clase B se busca como las otras tres (contracts/lista-de-expresiones.md §3 y §7; data-model §1; research D6, D7, V18; plan.md paso 1): internal/evals/prohibidas.go: `ExtraerExpresionesProhibidas(texto, lista)` (1) quita cada forma de `FormasFijas`, en su orden y en todas sus apariciones, compilada a una expresión regular como fija contracts §3, punto 1 —su texto con `regexp.QuoteMeta`, cada blanco como `[ \t]+`, `<fecha>` como `[0-9]{8}`, `<cita>` como el grupo opcional con la cita y lo que la sigue en la forma hasta el blanco siguiente (en la línea, `(?:<cita>:[ \t]+)?`), y un `⚠` inicial que admite los blancos y el énfasis de Markdown de la forma fija de los avisos (H5.1)—; (2) quita la marca, la etiqueta y los dos puntos de cada aviso de vigencia con las expresiones de `ExtraerAvisos` (`boe.EtiquetasDeAviso`) y deja lo que sigue en la línea; (3) cambia cada tramo quitado por un salto de línea; y (4) busca como hoy (H7.2 FR 051, sin cambios) en el orden maquinaria, otra conversación, anuncio y redacción no leída, sin repetir; las formas se compilan una vez por lista y no por respuesta; una lista sin familias ni formas no quita ni encuentra nada; `esDeLaClaseB(expresion string) bool`, método sin exportar, dice si una expresión es de `RedaccionNoLeida`; internal/evals/avisos.go solo si hace falta compartir sus expresiones de etiqueta, sin cambiar lo que da `ExtraerAvisos`; internal/evals/informe.go: el recuento trata la lista como vacía solo si lo están las cuatro familias; `Juzgar`, la regla por serie, las expresiones por sesión y el recuento por modelo toman la extracción sin más cambios (FR-036); tests: `TestExtraerExpresionesProhibidas` (internal/evals/prohibidas_test.go), con una lista sintética que lleva las dos formas fijas de contracts §1 y, en `anuncio`, `se consultó antes` y `consulta anterior`, gana los casos de contracts §6, último párrafo —la línea `⚠ REDACCIÓN MODIFICADA:` con su cita, con énfasis, con otra cita y sin cita (la de v0.1.3) se quita en los cuatro; la frase de la regla 7 con punto, sin él y seguida de un paréntesis, también; una etiqueta de aviso seguida de una expresión en la misma línea da la expresión; «⚠ REDACCIÓN MODIFICADA: la redacción que se consultó antes ha sido sustituida.» da `se consultó antes`; «la que se consultó antes» sola, fuera de la forma, la da; una lista sin formas fijas no quita nada; el orden de las cuatro familias; y `esDeLaClaseB` verdadero para `ya no exige` y falso para `ya puedo responder`—; el caso `sin-comprobar-la-redaccion` pasa a una lista con formas fijas; «No hay avisos de vigencia sobre este bloque.» sigue sin dar ninguna (FR-012); en internal/evals/conjunto_test.go, `listaDelRepositorio` exige además `redaccion_no_leida` y `formas_fijas` no vacías, para que ninguna subprueba pase en vacío; los resultados esperados de internal/evals/juzgar_test.go e internal/evals/informe_test.go que cambien por lo quitado se actualizan, sin desactivar ninguno; verificación: `expresiones-calibradas` sigue en 35 y 10 sin tocar su tabla (`ya no exige` no marca ninguna respuesta de los informes de H7.1 y H7.2, research V20, y lo quitado no lleva ninguna expresión de la lista de hoy); con la forma fija de la línea quitada un momento de la lista sintética, el caso de la línea con cita falla nombrando `se consultó antes` (el mutante se retira antes de `make ci`); `make ci` en verde — FR-012, FR-030, FR-031, FR-036, FR-093; SC-002, SC-003; US1-4, US1-5; rutas: internal/evals/prohibidas.go, internal/evals/avisos.go, internal/evals/informe.go, internal/evals/conjunto_test.go, internal/evals/juzgar_test.go

---

## Fase 2: US3 y US2 — El juicio de las líneas esperadas y de las skills que no se activan (P1)

**Objetivo**: `Juzgar` exige cada línea `⚠ REDACCIÓN MODIFICADA:` que la eval declara, por su bloque y sus dos fechas, y
que no se active ninguna skill de `no_se_activan`; el informe publica las líneas encontradas y las ausentes. **Prueba
independiente**: los casos de la eval 20 y de `legal-core` en `TestJuzgar` y `TestInformeMarkdownDeLosUmbrales`
(quickstart.md §3; SC-007).

- [X] T004 [US3] El juicio de las líneas `⚠ REDACCIÓN MODIFICADA:` esperadas y de las skills que no se activan, y su informe (contracts/evals-y-juicio.md §2; contracts/informe-del-job.md §4; data-model §2 y §4; research D9, D10; plan.md paso 2): internal/evals/redacciones.go, nuevo: `ExtraerRedaccionesModificadas(texto string) []RedaccionEsperada` —por cada línea con la forma de la etiqueta de `version-obsoleta` (las expresiones de `formasDeHallazgo`, las de `ExtraerHallazgos`), la primera cita de la línea (`ExtraerCitas`) y sus dos primeras fechas de ocho cifras con un mes y un día posibles y sin otra cifra a los lados, en su orden; nada de una línea sin cita o con menos de dos fechas—, y el texto `<norma> <bloque> <fecha_vigencia> <fecha_vigencia_reciente>` de `RedaccionEsperada`; internal/evals/juzgar.go: los motivos en el orden de contracts §2 —detrás de la activación, «se activó la skill <nombre>, que la eval dice que no se activa» por cada skill de `NoSeActivan` que la sesión activó; detrás de los hallazgos ausentes, «redacción modificada ausente: <texto>» por cada esperada que no da ninguna línea con su norma, su bloque y sus dos fechas en su orden—; `Pasa` exige además las dos cosas; `ResultadoDeEval` gana `RedaccionesEncontradas` y `RedaccionesAusentes` (`redacciones_modificadas_encontradas`, `redacciones_modificadas_ausentes`, con el texto de cada una en el orden de la eval; `[]`, nunca `null`, si la eval no las espera); una eval sin las claves nuevas se juzga como hoy (FR-053, FR-062); internal/evals/informe.go: `formas` de la serie gana, detrás de las de sus hallazgos, `⚠ REDACCIÓN MODIFICADA: <texto>` por cada esperada; en informe.md, la columna «Formas exigidas» de «Tasas por eval» las lleva, y «Sesiones» gana, detrás de «Hallazgos ausentes», «Redacciones modificadas encontradas» y «Redacciones modificadas ausentes» (`ninguna` si no hay); tests: en `TestJuzgar` (internal/evals/juzgar_test.go), con una eval sintética con las dos `redacciones_modificadas` de la eval 20 (contracts/evals-y-juicio.md §3), los casos de contracts §2 —con las dos líneas, una por bloque con su cita y sus fechas, pasa; con una sola, con las dos sin cita, con la cita de un bloque y las fechas del otro, con las fechas en otro orden o con las dos en una sola línea, no pasa y nombra cada ausente; una eval sin la clave, el juicio de hoy—, y la sesión de una eval de `legal-core` (`activa: true`, `no_se_activan: [boe-legislacion]`) que activa `legal-core` y `boe-legislacion` no pasa, con el motivo que nombra `boe-legislacion`, mientras que la misma sin activar `boe-legislacion` pasa (FR-094); `TestInformeMarkdownDeLosUmbrales` (internal/evals/informe_test.go) gana las columnas y las formas exigidas de contracts/informe-del-job.md §4; los resultados e informes esperados enteros de internal/evals/juzgar_test.go, internal/evals/informe_test.go, internal/evals/umbrales_test.go e internal/evals/sondeo_test.go ganan las dos claves vacías y las columnas, sin desactivar ninguno; verificación: con la comparación de las líneas reducida un momento a la norma y el bloque, el caso de la cita de un bloque con las fechas del otro falla (se restaura antes de `make ci`); `make ci` en verde — FR-003, FR-004, FR-023, FR-052, FR-053, FR-062, FR-094; SC-007; US2-3, US3-1, US3-2; rutas: internal/evals/redacciones.go, internal/evals/juzgar.go, internal/evals/informe.go, internal/evals/umbrales_test.go, internal/evals/sondeo_test.go

---

## Fase 3: US5 y US4 — La respuesta a la pregunta, los recuentos y los umbrales que deciden (P1 y P2)

**Objetivo**: lo que se juzga de una sesión es la respuesta del turno de la pregunta (el primer `result`); los recuentos
cuentan solo las sesiones terminadas; `sin_activar:<modelo>` y `redaccion_no_leida:<modelo>` deciden con 0, junto a las
expresiones (≤ 5 %) y la duración (≤ 900 s). La lectura (US5) va antes que los umbrales (US4) porque las respuestas que
cuentan son las terminadas con su respuesta a la pregunta (FR-045). **Prueba independiente**: `TestLeerSesion`,
`TestUmbralesDelInforme`, `TestInformeMarkdownDeLosUmbrales` y `TestInformeConSesionesSinMedir` (quickstart.md §5 y §6;
SC-006, SC-008); en el job del cierre, los cuatro umbrales (SC-001).

- [X] T005 [datos] [US5] La respuesta que se juzga es la de la pregunta, el primer `result`, con el caso que lo fija (contracts/evals-y-juicio.md §6; data-model §3; research D12, V3, V4, V5, V17; plan.md paso 3 y *Complexity Tracking*: el caso nuevo con la lectura de hoy deja `make ci` en rojo, y la lectura nueva sin el caso no tiene test que la vea, así que van juntos): internal/evals/testdata/sesiones/leer-sesion/respuesta-antes-de-una-tarea-en-segundo-plano/, nuevo, un transcript sintético: `sesion.jsonl` con, en este orden, `system/init`; un `assistant` con el `tool_use` de `Skill` (`boe-legislacion`); un `assistant` con un `tool_use` de `Bash` en segundo plano; su `user` con el `tool_result`; un `assistant` con el texto de la respuesta; un `result` `success` con la respuesta con cita (el texto de `respuestaConCita` de los tests del paquete); un `system` con `subtype` `task_notification` (`task_id`, `tool_use_id`, research V4); un `assistant` con la réplica; y un `result` `success` con la réplica («Esa tarea en segundo plano era solo una búsqueda auxiliar…»); `codigo-de-la-sesion` con `0` y `sesion.err` vacío; internal/evals/sesion.go: `LeerSesion` toma `Sesion.Respuesta` del **primer** mensaje `result` del transcript si tiene `subtype` `success` e `is_error` falso (vacía si no), y deja `Fin`, `Terminada`, `MotivoSinTerminar` y `ErrorDelResultado` del último mensaje y las activaciones de todo el transcript, como hoy (FR-062); el comentario que decía que cuenta la del último dice lo nuevo; el job y el sondeo la comparten sin más cambios; tests: `TestLeerSesion` (internal/evals/sesion_test.go) gana el caso `respuesta-antes-de-una-tarea-en-segundo-plano` —activada, terminada, `Respuesta` la de la pregunta y `Fin` `result success`—; los casos de hoy no cambian; verificación: con la lectura del último `result` restaurada un momento, el caso nuevo falla con la réplica como respuesta (se retira antes de `make ci`); quickstart.md §6; `make ci` en verde — FR-060, FR-062, FR-098; SC-008; US5-1; rutas: internal/evals/testdata/sesiones/leer-sesion/respuesta-antes-de-una-tarea-en-segundo-plano/, internal/evals/sesion.go
- [X] T006 [US4] Los recuentos cuentan solo las sesiones terminadas, y `sin_activar` y `redaccion_no_leida` del modelo que decide deciden con 0 (contracts/informe-del-job.md §1, §2, §3, §5 y §6; data-model §5 y §6; research D13, D14, V6, V7; plan.md paso 4; controles de umbral de FR-040, FR-041, FR-042, FR-043, FR-045 y FR-097, y el de «ninguna sin medir»): internal/evals/informe.go: el recuento por modelo (`recontarExpresiones` y su tipo interno) cuenta como respuesta medida solo la sesión juzgada, legible, no sin medir, **terminada** (`sesion_terminada`), de una serie que pide el plan (no la de la prueba de red) y de una eval que activa la skill, y da, además de `respuestas` y `conAlguna`, `conRedaccionNoLeida` (con alguna expresión de `redaccion_no_leida`, por `esDeLaClaseB`) y `sinActivar` (sin la skill activada); `expresiones_prohibidas_por_modelo` publica lo de hoy (`modelo`, `con_alguna`, `respuestas`) con ese recuento; la sesión sin terminar sigue publicada con su motivo y sin pasar en su serie (H5); internal/evals/umbrales.go: en una skill con lista, en este orden, `expresiones_prohibidas:<que decide>` (0,05), `sin_activar:<que decide>` (0) y `redaccion_no_leida:<que decide>` (0), los tres con `total` igual a `respuestas`, `comparacion` `"<="`, `decide: true` y la `descripcion` de contracts §2 carácter a carácter; después, `expresiones_prohibidas:` de cada informativo con `decide: false`; detrás, `duracion_de_las_sesiones` si hay objetivo, como hoy; en una skill sin lista, solo el de la duración (`legal-core`: `[]`, FR-048); ni `sin_activar` ni `redaccion_no_leida` para el informativo (FR-044); `cumple` en `float64` sin redondeos (0 si el total es 0); el motivo de la raíz de hoy, `umbral <nombre>: <medida> de <total> (<p> %), y tiene que ser ≤ <u> %`, por cada uno que decide y no se cumple; uno que se cumple no cambia el veredicto (FR-046); internal/evals/sondeo.go: la línea del recuento de la salida del sondeo usa el mismo recuento, solo de las terminadas (FR-082), y nada más de su salida cambia; tests: `TestUmbralesDelInforme` (internal/evals/umbrales_test.go) gana, con sesiones sintéticas del modelo que decide en 18 evals que activan la skill (10 que deciden y 8 informativas, 54 respuestas) escritas con `EscribirInforme` en `t.TempDir()` y la lista del repositorio, los casos de contracts/informe-del-job.md §6 con su veredicto y sus motivos esperados —`sin-activar-una` (`fallo`, con «umbral sin_activar:claude-sonnet-5-5: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %»), `sin-activar-ninguna`, `redaccion-no-leida-una` (una respuesta con `ya no exige`: 1 en `redaccion_no_leida` y en `expresiones_prohibidas`, `fallo` con su motivo), `redaccion-no-leida-ninguna`, `tres-de-54` (`fallo`), `dos-de-54` (se cumple), `tres-de-54-y-seis-sin-terminar` (seis sesiones más cortadas por el tope, código 124: `total` 54 y no 60, no se cumple, y las seis publicadas con su motivo y fuera de `expresiones_prohibidas_por_modelo`) y `una-sin-activar-con-expresion` (cuenta una vez en cada umbral)—, y en todos rehacer la comparación de cada elemento con su `medida`, su `total`, su `comparacion` y su `umbral` da su `cumple`; los casos de hoy (901 s y 900 s de la duración, el informativo con `decide: false`, `legal-core` con `[]`) siguen; `TestInformeMarkdownDeLosUmbrales` (internal/evals/informe_test.go) con las cinco filas; `TestSalidaDelSondeo` y `TestJuicioDelSondeo` (internal/evals/sondeo_test.go), con una sesión sin terminar fuera de la línea del recuento; los informes esperados que arman la lista del repositorio o una con familias, en internal/evals/informe_test.go, internal/evals/umbrales_test.go e internal/evals/sondeo_test.go, ganan los dos umbrales, y los que contaban sesiones sin terminar dejan de contarlas, sin desactivar ninguno; verificación: con el filtro de las terminadas quitado un momento, `tres-de-54-y-seis-sin-terminar` falla por el total; con `sin_activar` a umbral 1 un momento, `sin-activar-una` falla; con `redaccion_no_leida` en `decide: false` un momento, `redaccion-no-leida-una` falla; y `TestInformeConSesionesSinMedir` sigue viendo fallar el mutante de H7.3 (la clasificación desconectada del juicio del informe); los mutantes se retiran antes de `make ci`; quickstart.md §5; `make ci` en verde — FR-036, FR-037, FR-040, FR-041, FR-042, FR-043, FR-044, FR-045, FR-046, FR-047, FR-048, FR-061, FR-062, FR-082, FR-097; SC-001 (el control), SC-006; US2-4, US4-1, US4-2, US4-3, US5-2; rutas: internal/evals/informe.go, internal/evals/umbrales.go, internal/evals/sondeo.go

---

## Fase 4: US6 — Una tanda de sesiones por commit y skill (P2)

**Objetivo**: un trabajo `tanda`, antes de `evals`, decide si la ejecución mide: no mide si una anterior sobre el mismo
commit, sin terminar, ya mide; así el segundo disparo salta `evals` entero, sin abrir sesiones ni dejar una comprobación
roja o de más, y la etiqueta sobre un commit ya medido vuelve a medir. La definición lleva además el tope de 122 min que
exige la eval 20 (T009). **Prueba independiente**: `TestDefinicionDelJob` con `del-repositorio`, `sinteticas`,
`segundo-disparo` y `estado-de-la-tanda` (quickstart.md §7; SC-010).

- [X] T007 [US6] La decisión de la tanda, pura y probada sin red (contracts/tanda-del-job.md §2 y §4, subpruebas `segundo-disparo` y `estado-de-la-tanda`; data-model §7; research D15, S2, S4, V11; plan.md paso 7; control de umbral de FR-070 y FR-071): internal/evals/tanda.go, nuevo y todo sin exportar: `ejecucionDelCommit` (`id int64`, `terminada bool`, `tanda estadoDeLaTanda`), `estadoDeLaTanda` (`tandaSinDecidir`, `tandaQueMide`, `tandaQueNoMide`), `decisionDeLaTanda` (`mide bool`, `pendientes []int64`) y `decidirLaTanda(propia int64, ejecuciones []ejecucionDelCommit) decisionDeLaTanda` con las reglas de contracts §2 —solo cuentan las de `id` menor que la propia y sin terminar; con alguna `tandaQueMide`, no mide y sin pendientes; si no, con alguna `tandaSinDecidir`, mide con ellas como pendientes; si no, mide—; la lectura del JSON de la lista de ejecuciones del flujo (`databaseId`, `status`; `terminada` es `status == "completed"`) y la del JSON de los trabajos de una ejecución (`jobs` con `name`, `status`, `conclusion` y `steps` con `name` y `conclusion`), que da `tandaQueMide` si el trabajo `tanda` está `completed` y su paso «Esta ejecución mide el commit» tiene `conclusion` `success`, `tandaSinDecidir` si no hay trabajo `tanda` o no está `completed`, y `tandaQueNoMide` en cualquier otro caso (saltado, fallido, cancelado o la marca saltada); el nombre del trabajo y el del paso, constantes del paquete; `esperarLaDecision`, que consulta con una función recibida, decide y, con pendientes, espera `esperaEntreConsultas` (10 s) con una espera recibida y vuelve a consultar; pasada `esperaMaximaDeLaDecision` (10 min) con alguna pendiente, mide y lo dice en lo que devuelve para el registro; un error de la consulta es un error que la nombra; nada ejecuta `gh` todavía (T008); tests: en `TestDefinicionDelJob` (internal/evals/definicion_test.go), la subprueba `segundo-disparo` con la tabla de contracts §4 (`anterior-que-mide-corre`, `anterior-que-mide-espera`, `anterior-sin-decidir-que-mide` con dos consultas, `anterior-sin-decidir-que-no-mide`, `anterior-que-no-mide`, `anterior-terminada` —mide: la etiqueta vuelve a medir, FR-071—, `posterior-que-mide`, `sola` y `espera-agotada`), con consultas sintéticas y un reloj y una espera falsos, sin dormir de verdad; y la subprueba `estado-de-la-tanda`, con el JSON sintético de los trabajos de una ejecución —la marca en `success`, la marca `skipped`, `tanda` `in_progress`, sin `tanda` y `tanda` `skipped`—; verificación: con la regla de solo las anteriores quitada un momento, `posterior-que-mide` falla, y con la de sin terminar quitada, `anterior-terminada` falla (se retiran antes de `make ci`); `make ci` en verde — FR-070, FR-071, FR-100; SC-010; US6-1, US6-2; rutas: internal/evals/tanda.go, internal/evals/definicion_test.go
- [X] T008 [US6] La definición del job con el trabajo `tanda`, el punto de entrada que decide y el tope de 122 min, comprobados en `make ci` (contracts/tanda-del-job.md §1, §2 «Las órdenes», §3 y §4; research D15, D16, D17, S3, S9, V12, V13; plan.md paso 7; controles de umbral de FR-054 —el tope—, FR-070, FR-071 y FR-100): .github/workflows/evals.yml: el trabajo `tanda` de contracts §1 carácter a carácter —su comentario, `needs: [cambios]`, el `if` que hoy lleva `evals`, `runs-on`, `timeout-minutes: 15`, `permissions` con `contents: read` y `actions: read`, sin `concurrency`, `outputs.medir`, y los pasos de checkout del commit evaluado, Go con la caché, `decidir` (con el `run` de contracts §1, que ejecuta `-run '^TestTandaDelCommit$'` con `-commit`, `-ejecucion` y `-salida "$GITHUB_OUTPUT"`, y `GH_TOKEN`, `GH_REPO`, `COMMIT_EVALUADO` y `EJECUCION` en su `env`) y la marca «Esta ejecución mide el commit» con `if: steps.decidir.outputs.medir == 'si'`—; el trabajo `evals` pasa a `needs: [tanda]` e `if: ${{ !cancelled() && needs.tanda.outputs.medir == 'si' }}`, con el comentario de contracts §1, y a `timeout-minutes: 122` (7 285 s del peor caso con 20 evals y 97 sesiones, research V13); su `concurrency` por commit y skill con `cancel-in-progress: false`, la matriz, `env` y los pasos no cambian; internal/evals/tanda.go: las dos órdenes constantes de `gh` de contracts §2, «Las órdenes» (la lista de ejecuciones de `evals.yml` sobre `"$COMMIT_EVALUADO"` con `--limit 100 --json databaseId,status`, y los trabajos de `"$EJECUCION_ANTERIOR"` con `--json jobs`), ejecutadas con `exec.CommandContext(ctx, "sh", "-c", <constante>)` y los valores en el entorno de la orden (`os.Environ()` más su variable), sin ningún dato en el texto de la orden (G204, research V12); una orden que termina con otro código que `0` es un error que la nombra con lo que escribió; internal/evals/job_test.go (etiqueta `evals`): `TestTandaDelCommit`, con `-commit` (la de hoy), `-ejecucion` y `-salida`, consulta con esas órdenes, decide con `esperarLaDecision` (T007) y la espera real, registra con `t.Logf` qué ejecuciones anteriores miró y por qué mide o no, y añade a `-salida` una línea `medir=si` o `medir=no`; falla solo con un error; no abre sesiones ni escribe fuera de `-salida`; internal/evals/definicion.go: `leerDefinicionDelJob` lee además de `jobs.tanda` su `if`, `permissions`, `concurrency`, `outputs`, `timeout-minutes` y sus pasos (`id`, `name`, `if`, `run`), y `needs` e `if` de `jobs.evals`; tests: en `TestDefinicionDelJob` (internal/evals/definicion_test.go), `del-repositorio` gana una línea por clave que no es la de contracts §4 —el `if` de `tanda`; `tanda` sin `concurrency`; `actions: read`; `outputs.medir`; el paso `decidir` con `-run '^TestTandaDelCommit$'`, `-commit`, `-ejecucion` y `-salida "$GITHUB_OUTPUT"`; el último paso con el nombre de la marca de T007 y su `if`; `evals` con `needs: [tanda]` y su `if`—, además de lo de hoy (sin `concurrency` de flujo, `cancel-in-progress: false` y el tope que cubre el peor caso); `sinteticas` gana una definición por cada una de esas claves que se aparta sola y ve su línea —sin `tanda`, con `concurrency` en `tanda`, sin `actions: read`, `evals` sin `needs: [tanda]` o con otro `if`, la marca con otro nombre o sin su `if`—; verificación: en una copia de `git archive` en un temporal fuera del repositorio, con el `if` de `evals` sin `needs.tanda.outputs.medir`, `TestDefinicionDelJob` da `FAIL` nombrando la clave, y con `timeout-minutes: 100`, falla por el tope (la copia se borra); `go vet` con la etiqueta `evals` sobre el paquete sin salida, y la lista de sus tests con esa etiqueta nombra `TestTandaDelCommit`; ninguna tarea ejecuta `TestTandaDelCommit` (consulta `gh`); release_test.go, que también lee la definición del job, sigue en verde sin tocarlo; quickstart.md §7, su primera orden; `make ci` en verde — FR-043, FR-054, FR-070, FR-071, FR-072, FR-100; SC-010; US6-1, US6-2; rutas: .github/workflows/evals.yml, internal/evals/tanda.go, internal/evals/job_test.go, internal/evals/definicion.go

---

## Fase 5: US3 — La eval de dos bloques de la LCSP con la redacción cambiada (P1)

**Objetivo**: la eval 20, informativa, lee `a1-30` y `da-3` de la LCSP con un grafo previo que vio sus redacciones
originales, y exige dos líneas `⚠ REDACCIÓN MODIFICADA:`, cada una con su cita y sus fechas. Va antes que la skill
(Definition of Done §1.10) y después del tope de T008 (con 20 evals, el peor caso pasa de 7 200 s). **Prueba
independiente**: `TestGrabacionesDerivadas` y `TestEvalsDelRepositorio/grafo-previo` (quickstart.md §7; SC-012).

- [ ] T009 [datos] [US3] La eval 20 y su grafo previo, con las dos derivadas escritas por código y comprobadas (contracts/evals-y-juicio.md §3 y §4; data-model §9; research D11, D21, V14, V15, V16, V21; plan.md paso 5 y *Complexity Tracking*): evals/boe-legislacion/20-lcsp-dos-bloques-redaccion-cambiada.yaml, nueva, carácter a carácter la de contracts §3 (su comentario; la pregunta de FR-050; `activa: true`; `informativa: true`; `grafo_previo` con `lcsp-a1-30-y-da-3-redaccion-original` y los dos comandos de bloque; los comandos de `a1-30`, de `da-3` y de `graph check` con la norma; `graph show` prohibido; las dos citas; `hallazgos: [version-obsoleta]`; y las dos `redacciones_modificadas`, `a1-30` de 20180309 a 20200206 y `da-3` de 20180309 a 20230101); internal/app/grafo_test.go: `derivadasDelGrafoPrevio()` gana dos entradas con la subcarpeta `lcsp-a1-30-y-da-3-redaccion-original` —argumentos `articulo BOE-A-2017-12902 a1-30` y `articulo BOE-A-2017-12902 da-3`, con la fecha 20180309 y el fichero de la grabación de H4 de cada bloque—; testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/, nuevo, con las dos derivadas de contracts §4 escritas por `TestGrabacionesDerivadas` con la bandera `-actualizar-derivadas` (la de `a1-30`, byte a byte la de la 19; la de `da-3`, su grabación sin la redacción de 20230101); internal/evals/conjunto_test.go: `compruebaLaSesion` exige además que cada redacción de `RedaccionesModificadas` sea un `version-obsoleta` de su norma y su bloque, con su `FechaVigencia` como la superada y su `FechaVigenciaReciente` como la leída (la eval no puede esperar unas fechas que las grabaciones no dan), y `probarGrafosPrevios` prepara el de la eval 20 como el de la 19; ninguna grabación nueva: las de `a1-30` y `da-3` de `BOE-A-2017-12902` están desde H4 en la unión de grabaciones de las evals, y el paso `grabar_datos` no tiene nada que grabar (FR-055; research D21); verificación: `TestGrabacionesDerivadas` sin la bandera pasa con las dos (byte a byte su derivación, la redacción de su fecha que trae la grabada y ningún fichero sin entrada en la carpeta); con la fecha reciente de `da-3` en la eval puesta un momento a 20200206, `TestEvalsDelRepositorio/grafo-previo` falla nombrando la eval y el bloque (se restaura antes de `make ci`); `TestEvalsDelRepositorio` pasa con 20 evals (10 positivas que deciden, 2 de no activación y 8 informativas, FR-054) y `expresiones-en-los-bloques` sin ninguna expresión en `da-3` ni en las derivadas; `TestDefinicionDelJob` pasa con el tope de T008 (7 285 s ≤ 7 320 s); quickstart.md §7, su segunda orden; `make ci` en verde — FR-050, FR-051, FR-052, FR-054, FR-055, FR-096; SC-003, SC-012; US3-1, US3-4; rutas: evals/boe-legislacion/20-lcsp-dos-bloques-redaccion-cambiada.yaml, testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/, internal/app/grafo_test.go, internal/evals/conjunto_test.go

---

## Fase 6: US1, US2 y US3 — `boe-legislacion` v0.1.4 con la lista entera (P1)

**Objetivo**: la skill se activa ante toda pregunta de su ámbito, no cuenta la comprobación ni anuncia la respuesta, no
describe una redacción que no ha leído, da una línea `⚠ REDACCIÓN MODIFICADA:` con la cita de su bloque por precepto
cambiado y enseña la orden de lectura y comprobación también para PowerShell; la lista gana las 25 formas de la clase A y
las nueve que faltan de la B, calibrada sobre tres informes. **Prueba independiente**: `expresiones-calibradas`,
`expresiones-en-los-bloques`, `expresiones-de-la-skill`, `prosa-de-la-skill`, `ordenes-para-powershell`,
`TestOrdenesParaPowerShell`, `TestJuzgarLasClasesDeLaRespuesta` y `make skills-check` (quickstart.md §2, §3 y §4;
SC-002 a SC-005, SC-007, SC-011); en el job del cierre, los umbrales y la 04 (SC-001).

- [ ] T010 [US1] `boe-legislacion` v0.1.4 y la lista entera, calibrada sobre tres informes y comprobada contra la skill (contracts/skill-boe-legislacion.md §1 a §4; contracts/lista-de-expresiones.md §1, §4, §5 y §6; research «Causa de raíz», «Calibrado», D1 a D8, V18 a V21 y V27; plan.md paso 6 y «Cambios de SKILL.md trazados a la causa de raíz»; controles de umbral de FR-032, FR-033, FR-021, FR-024, FR-034 y FR-026; una sola tarea porque con la lista entera la prosa de v0.1.3 da 9 defectos, research V19): skills/boe-legislacion/SKILL.md con los cambios C1 a C10 de contracts §1 y su texto, cada uno trazado a su causa —C1, la `description` que se usa siempre que la respuesta dependa de lo que dice una norma, también cuando se cree conocer la respuesta (FR-001 a FR-003); C2 y C4, cada orden de lectura y comprobación, la de `articulo` y la de `articulos`, con su forma para PowerShell detrás, `…; if ($LASTEXITCODE -eq 0) { … }`, con la misma norma y los mismos bloques (FR-024); C3, C5, C8 y C10, la prosa sin «la comprobación» ni «lectura anterior» (FR-021); C6, el paso 4 sin «si lo leído basta» y con «Lo que decidas en este paso no va en la respuesta» (FR-010); C7, lo que lleva la respuesta; C9, la sección «Redacción modificada», con la línea `⚠ REDACCIÓN MODIFICADA:` con la cita de su bloque y sus dos fechas `AAAAMMDD`, una por bloque, por qué no se describe la redacción superada, qué sí se dice y qué responder si se pregunta qué cambió (FR-011, FR-022, FR-023)—; lo de contracts §2 se queda (la frase fija de la regla 7, la cita, los avisos, las reglas 1 a 6, la lectura de uno en uno con la comprobación detrás una vez por orden, la región generada intacta, ninguna fecha escrita con cifras y ningún nombre de eval, del job ni de un modelo, FR-025); menos de 300 líneas y `description` de hasta 1024 caracteres (FR-026); evals/boe-legislacion/expresiones-prohibidas.yaml: `anuncio` gana, detrás de sus 14, las 25 formas de contracts §1 en su orden, y `redaccion_no_leida`, detrás de `ya no exige`, las otras nueve: 87 expresiones, ninguna que case con «No hay avisos de vigencia sobre este bloque» ni con la respuesta de FR-022 (FR-012); internal/evals/conjunto_test.go: `expresiones-calibradas` lee los tres informes versionados (se añade el de H7.3, con su premisa de 93 respuestas; los tres solo se leen, FR-090) y compara cinco columnas —`maquinaria`, `otra_conversacion`, `anuncio`, `redaccion_no_leida` y `alguna`— por las dos cifras de cada eval con la tabla de contracts/lista-de-expresiones.md §4 (36, 11 y 9 de 93; 0 en las demás), con el mensaje que nombra el informe, la eval, la columna, lo contado y lo calibrado; `expresiones-de-la-skill` compone la respuesta de contracts/lista-de-expresiones.md §5 (la forma escrita de cada aviso y de cada clase de hallazgo, cada una con su frase de ejemplo, y cada bloque `text` de `SKILL.md` con sus marcadores sustituidos) y exige que no se marque, que cada forma fija de la lista quite algo de ella y, como premisa, que sin quitarlas lleve `se consultó antes` y `consulta anterior`; la subprueba nueva `ordenes-para-powershell` aplica a `SKILL.md` `defectosDeLasOrdenesParaPowerShell(markdown string) []string`, nueva y solo de tests, con las reglas de contracts/skill-boe-legislacion.md §4; `prosa-de-la-skill` no cambia de código y aplica la lista entera; `TestOrdenesParaPowerShell` (internal/evals/conjunto_test.go) la fija con Markdown sintético —las dos órdenes con sus dos formas, sin error; sin la de PowerShell; con `&&` en ella; sin el `if`; con otra norma u otros bloques en la de PowerShell; con otros bloques en la comprobación; sin la de `articulos`—; `TestJuzgarLasClasesDeLaRespuesta` (internal/evals/juzgar_test.go), nuevo, con la lista del repositorio y una eval que activa la skill: cada frase de contracts/lista-de-expresiones.md §6, entera y en una respuesta con una cita esperada, marcada en su clase («A»: alguna expresión y ninguna de `redaccion_no_leida`; «B»: alguna de `redaccion_no_leida`); sin marcar, la respuesta compuesta de §5, «No hay avisos de vigencia sobre este bloque.», «No puedo decirte qué contenía la redacción anterior.» y «Cito la redacción vigente; la que había antes no la he leído, así que no puedo decir qué ha cambiado.»; y marcadas en la clase A las mismas palabras de las formas fijas fuera de ellas; los resultados esperados de los tests cuyo texto lleva una forma nueva y que aplican la lista del repositorio o una con esas formas, buscados antes con la orden de research V27 en los tests del paquete de evals —hoy el caso de internal/evals/juzgar_test.go que gana `consulta anterior`, el de la frase de la regla 7 de internal/evals/prohibidas_test.go, que la forma fija quita, y los dos de `TestCondicionesDeLaConsultaRepetida` en internal/evals/consulta_repetida_test.go, que ganan `se consultó antes`—, y los informes armados con la lista del repositorio en internal/evals/informe_test.go, internal/evals/umbrales_test.go e internal/evals/sondeo_test.go, se actualizan sin desactivar ninguno; verificación: `defectosDeLaProsa` con la lista entera sobre la SKILL.md de `main`, en un `_test.go` temporal que se borra antes de `make ci`, da los 9 defectos de research V19; con `vigente hasta` quitada un momento de la lista, `expresiones-calibradas` falla nombrando la columna `redaccion_no_leida` de la 19 del informe de H7.3; con la forma de PowerShell de `articulos` quitada un momento de `SKILL.md`, `ordenes-para-powershell` falla nombrando su orden de Bash (los mutantes se retiran antes de `make ci`); quickstart.md §2, §3 y §4; `make skills-check` y `make ci` en verde — FR-001, FR-002, FR-003, FR-010, FR-011, FR-012, FR-020, FR-021, FR-022, FR-023, FR-024, FR-025, FR-026, FR-030, FR-032, FR-033, FR-034, FR-037, FR-093, FR-094, FR-095; SC-002, SC-003, SC-004, SC-005, SC-007, SC-011; US1-1 a US1-7, US2-1, US2-2, US3-3; rutas: skills/boe-legislacion/SKILL.md, evals/boe-legislacion/expresiones-prohibidas.yaml, internal/evals/conjunto_test.go, internal/evals/juzgar_test.go, internal/evals/prohibidas_test.go, internal/evals/consulta_repetida_test.go, internal/evals/informe_test.go, internal/evals/umbrales_test.go, internal/evals/sondeo_test.go

---

## Fase 7: US7 — Quien lanza el sondeo lee su error de uso, no una traza (P3)

**Objetivo**: con un argumento que no vale o sin la credencial, la salida de error del sondeo es solo su mensaje, con el
código de hoy y sin abrir ninguna sesión; un fallo de construcción sigue dando el registro. **Prueba independiente**:
`TestGuionDelSondeo`, `TestComprobarElSondeo` y `TestSondear` (quickstart.md §8, sus dos primeras órdenes; SC-009).

- [ ] T011 [US7] Los errores de uso del sondeo, sin la traza de `go test` (contracts/sondeo.md §1 a §4; data-model §8; research D18, V23, V26; plan.md paso 8; control de umbral de FR-080): internal/evals/sondeo.go: `errorDeUso`, tipo sin exportar que envuelve, con el mismo texto de hoy, lo que `comprobarElSondeo` devuelve por los argumentos (los de `SKILL`, `EVALS`, `MODELO`, `REPETICIONES` y `CONCURRENCIA`, unidos, uno por línea, en ese orden) o por la credencial (`errSinSuscripcion`); no lo son el error de leer la definición del job, construir el binario, instalar las skills, repartir las sesiones ni crear el directorio de sesiones (FR-081); con un `errorDeUso`, `sondear` no prepara el árbol ni llama al repartidor; internal/evals/job_test.go (etiqueta `evals`): `TestSondeo`, con un `errorDeUso` (`errors.As`), escribe su mensaje con un salto de línea final en `uso.txt` del temporal y termina sin fallar, sin escribir `salida.txt`; con otro error falla como hoy (`require.NoError`); sin error escribe `salida.txt` como hoy; scripts/evals-sondeo.sh, tras `go test` y en este orden: si falló, imprime `go-test.log` en la salida de error y sale con 1 (hoy); si `uso.txt` existe y no está vacío, lo imprime en la salida de error, sin nada más, y sale con 1; si no, imprime `salida.txt` (hoy); el temporal se borra al salir, con el código que sea (hoy); bash sin nada posterior a 3.2; su comentario de cabecera lo dice; tests: `TestGuionDelSondeo` (internal/evals/sondeo_test.go, con el `go` sustituto de hoy) gana dos casos —el sustituto deja en `uso.txt` las dos líneas de contracts §3 (`MODELO: está vacío` y `REPETICIONES: «0» no es un entero mayor o igual que 1`) y sale con 0: el guion sale con 1, su salida de error es exactamente esas dos líneas y la estándar está vacía; lo mismo con la línea de la credencial—, y el de hoy en que el sustituto falla sigue dando el registro entero con 1 (FR-081); `TestComprobarElSondeo` comprueba además que sus errores de argumentos y de credencial son `errorDeUso` y que el de una definición del job que no se puede leer no lo es; `TestSondear`, que con un `errorDeUso` no se prepara el árbol ni se abre ninguna sesión; verificación: `bash -n scripts/evals-sondeo.sh`; con la rama de `uso.txt` quitada un momento del guion, los dos casos nuevos de `TestGuionDelSondeo` fallan (se restaura antes de `make ci`); `go vet` con la etiqueta `evals` sobre el paquete sin salida; `make ci` en verde — FR-080, FR-081, FR-082, FR-099; SC-009; US7-1, US7-2, US7-3; rutas: internal/evals/sondeo.go, internal/evals/job_test.go, scripts/evals-sondeo.sh

---

## Fase 8: Documentación y cierre de la Definition of Done

**Objetivo**: la documentación que el hito deja falsa dice lo nuevo, y el cierre de la Definition of Done queda
comprobado y escrito para el informe final. **Prueba independiente**: `make ci` y `cierre.md`.

- [ ] T012 Documentación que el hito deja falsa (research D19, V25; contracts/skill-boe-legislacion.md §6; plan.md paso 9): CHANGELOG.md, *Unreleased* (H7 a H7.3 no han salido en ninguna release, así que describe el comportamiento final): en «Cambiado», la entrada de `boe-legislacion` pasa a v0.1.4 con el texto de contracts/skill-boe-legislacion.md §6 —se activa ante toda pregunta cuya respuesta dependa de lo que dice una norma, también si el modelo cree saberla; la respuesta no cuenta la comprobación ni lo que el agente va a hacer, y no describe una redacción que no ha leído (dice que cita la vigente y que la anterior no la ha leído); cada línea `⚠ REDACCIÓN MODIFICADA:` dice de qué precepto es, con su cita; y cada orden de lectura y comprobación tiene su forma para PowerShell—, y dice que sustituye a la v0.1.3, que no llegó a publicarse; y las entradas del job de evals: la lista por clases (`redaccion_no_leida`, y `formas_fijas`, que se quitan antes de buscar), los umbrales `sin_activar:<modelo>` y `redaccion_no_leida:<modelo>` que deciden con 0, la eval de dos bloques de la LCSP, el juicio sobre la respuesta a la pregunta (el primer `result`) con los recuentos solo de las sesiones terminadas, una tanda por commit y skill con el trabajo `tanda`, y los errores de uso del sondeo sin la traza (FR-027, FR-101); CONTRIBUTING.md, «Job de evals»: la tanda única en lugar de la espera del segundo disparo (research V25), la respuesta juzgada, los tres umbrales de la respuesta, la lista con `redaccion_no_leida` y `formas_fijas`, y las claves `no_se_activan` y `redacciones_modificadas` del formato de eval; internal/evals/doc.go: lo mismo en el comentario del paquete (la respuesta del primer `result`, la decisión de la tanda y el error de uso del sondeo); .github/workflows/evals.yml: los comentarios que aún describan la espera del segundo disparo o el último `result` dicen lo nuevo, sin cambiar ninguna clave; verificación: cada fichero, test, variable, clave y campo que nombran existe (`git grep` de cada uno); `TestDefinicionDelJob` sigue en verde; `make ci` en verde — FR-027, FR-101; SC-011; rutas: CHANGELOG.md, CONTRIBUTING.md, internal/evals/doc.go, .github/workflows/evals.yml
- [ ] T013 Cierre de la Definition of Done, sin código de producto: specs/014-h7-4-boe-legislacion-sin/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de los esquemas, las evals, la lista y los datos de prueba creados o modificados en el hito (los dos esquemas, la lista, la eval 20, las tres evals de `legal-core`, la carpeta de las dos derivadas y el caso de lectura de sesión), y que ninguna grabación ni ninguna fuente cambia, para la capa 3 del informe final; (2) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % y el dominio ≥ 85 % (Definition of Done §1.9; el hito no toca el dominio)— y la de los ficheros nuevos del paquete de evals, tanda.go y redacciones.go; (3) el resultado de quickstart.md §1 a §7 —§2 a §7 con `--- PASS` en cada test y subprueba nombrados, contados con `rtk proxy` para que no pase en vacío— y el de las dos primeras órdenes de §8, sin la credencial y con argumentos que no valen, que no abren ninguna sesión (la salida de error es solo el mensaje, sin nada de `go test` ni de testify, y `make` sale con 2); la tercera orden de §8, con modelo, y §9, que lee el informe del job del cierre, quedan fuera del run; (4) con `git diff --name-status main`: que no hay ADR nuevo; que los artefactos de H7.1, H7.2 y H7.3 no cambian (FR-090); que los guiones del workflow no cambian (FR-072); que el binario no cambia (nada fuera del paquete de evals salvo el test de derivadas de la aplicación); que la tabla de fuentes no cambia (ninguna fuente nueva); que la SKILL.md de `boe-legislacion` tiene menos de 300 líneas; que ningún umbral de FR-040 a FR-043 se ha rebajado, retirado ni dejado sin decidir y que la lista no ha perdido ninguna expresión (FR-047); y que plan.md tiene las cuatro filas del cierre de FR-092; si una cifra de cobertura queda bajo su umbral, se añaden los tests que faltan en los ficheros de test declarados, sin tocar código de producto; verificación: `make ci` en verde — FR-047, FR-072, FR-090, FR-091, FR-092; SC-011; rutas: specs/014-h7-4-boe-legislacion-sin/cierre.md, internal/evals/definicion_test.go, internal/evals/juzgar_test.go, internal/evals/umbrales_test.go, internal/evals/sondeo_test.go, internal/evals/prohibidas_test.go

---

## Dependencias y orden de ejecución

El orden es el de plan.md, «Orden de implementación» (de dentro afuera), con tres precisiones: el paso 1 se parte en
T001 y T002 (los dos esquemas, cada uno con su tipo y sus ficheros, sin lógica) y T003 (la supresión de las formas y la
cuarta familia en la extracción), para que las tareas `[datos]` no lleven lógica; el paso 7 (la tanda) se parte en T007
(la decisión, pura) y T008 (la definición y el punto de entrada) y va **antes** que el paso 5 (la eval 20, T009), porque
la definición es la que lleva `timeout-minutes: 122` y, sin él, `TestDefinicionDelJob` falla con 20 evals: así T009 no
mezcla la definición del job con sus datos; y la documentación (paso 9) va antes del cierre (T013).

| Tarea | Depende de | Por qué |
|---|---|---|
| T001 | — | el esquema de la lista, su campo y la lista cambian juntos o la lista es un fichero mal formado |
| T002 | — (orden del bucle) | el esquema de eval, sus campos y las evals de `legal-core` cambian juntos |
| T003 | T001 | la extracción lee `FormasFijas` y `RedaccionNoLeida`; `listaDelRepositorio` las exige |
| T004 | T002, T003 | el juicio lee `NoSeActivan` y `RedaccionesModificadas`; las líneas esperadas se extraen con las formas de hallazgo y las citas |
| T005 | — (orden del bucle) | la lectura del transcript no depende de la lista ni del formato |
| T006 | T003, T004, T005 | el recuento usa `esDeLaClaseB`, los resultados de T004 y las sesiones terminadas con la respuesta de T005 |
| T007 | — (orden del bucle) | la decisión es pura y no depende del juicio |
| T008 | T007 | el punto de entrada decide con `esperarLaDecision`; la definición nombra la marca de T007 |
| T009 | T002, T004, T008 | la eval declara `redacciones_modificadas` (T002), que `compruebaLaSesion` comprueba; con 20 evals, el tope de T008 cubre el peor caso |
| T010 | T003, T004, T006, T009 | la lista entera se aplica con las formas fijas; `TestJuzgarLasClasesDeLaRespuesta` usa la clase B; las evals (también la 20) van antes que la skill (DoD §1.10) |
| T011 | T006 | el sondeo comparte el recuento de T006; sus errores de uso no dependen de lo demás |
| T012 | T006, T008, T010, T011 | la documentación describe el comportamiento final: el juicio, los umbrales, la tanda, la skill y el sondeo |
| T013 | todas | enumera lo tocado, mide la cobertura final y ejecuta el quickstart salvo lo que abre una sesión con modelo y lo posterior al run |

Tras la última tarea, el workflow ejecuta `make ci`, hace la revisión final y el cierre en la plataforma, con el job de
evals (SC-001, FR-102); nada de eso es una tarea. No hay suite congelada que activar («Aceptación e2e: no aplica»).

### Oportunidades de paralelismo

En el bucle del workflow todo va en secuencia, y ninguna tarea lleva `[P]`: casi todas comparten fichero con la anterior
o la siguiente (`internal/evals/conjunto_test.go`, `juzgar_test.go`, `informe_test.go`, `umbrales_test.go`,
`sondeo_test.go`, `definicion_test.go`, `job_test.go`, `tanda.go`). Independencias que el orden no aprovecha:

- T002 (el formato de eval) no depende de T001.
- T005 (la respuesta a la pregunta) no depende de T001-T004.
- T007 (la decisión de la tanda) no depende de nada del juicio; T008 solo de T007.
- T011 (el sondeo) solo depende del recuento de T006.

### Ejemplo de ejecución en paralelo (fuera del workflow)

```text
# Tras T001, tres agentes podrían avanzar a la vez:
T002 → T003 → T004 → T005 → T006    formato, juicio y umbrales   (internal/evals/formato.go, prohibidas.go, juzgar.go, sesion.go, informe.go, umbrales.go)
T007 → T008                          la tanda por commit          (internal/evals/tanda.go, definicion.go, .github/workflows/evals.yml)
# Y, tras T006 y T008, T009 → T010 (la eval 20 y la skill) en paralelo con T011 (el sondeo).
```

## Criterios de prueba independiente por historia

| Historia | Prueba independiente | Tareas |
|---|---|---|
| US1 (P1) la norma, sin la comprobación ni una redacción no leída | los casos del esquema de la lista, `TestExtraerExpresionesProhibidas`, `expresiones-calibradas` (36, 11 y 9), `expresiones-en-los-bloques`, `expresiones-de-la-skill`, `prosa-de-la-skill`, `TestJuzgarLasClasesDeLaRespuesta` (quickstart.md §2 y §3); `expresiones_prohibidas` y `redaccion_no_leida` en el job del cierre (SC-001) | T001, T003, T010 |
| US2 (P1) la skill se activa ante toda pregunta de su ámbito | los casos del esquema de eval, el de `legal-core` que activa `boe-legislacion` en `TestJuzgar`, `sin-activar-*` de `TestUmbralesDelInforme` (quickstart.md §3 y §5); `sin_activar` y la 04 en el job del cierre (SC-001) | T002, T004, T006, T010 |
| US3 (P1) dos preceptos, dos líneas | los casos de la eval 20 en `TestJuzgar`, `TestGrabacionesDerivadas`, `grafo-previo`, `ordenes-para-powershell`, `TestOrdenesParaPowerShell` (quickstart.md §3, §4 y §7); la eval 20 con sus dos líneas y su tasa en el job del cierre (SC-001) | T004, T009, T010 |
| US4 (P1) los umbrales los hace cumplir el job | `TestUmbralesDelInforme`, `TestInformeMarkdownDeLosUmbrales`, `TestInformeConSesionesSinMedir` (quickstart.md §5); los cuatro umbrales en el job del cierre (SC-001) | T006 |
| US5 (P2) el job juzga la respuesta a la pregunta | `TestLeerSesion/respuesta-antes-de-una-tarea-en-segundo-plano`, `tres-de-54-y-seis-sin-terminar` (quickstart.md §5 y §6) | T005, T006 |
| US6 (P2) una tanda por commit y skill | `TestDefinicionDelJob` con `del-repositorio`, `sinteticas`, `segundo-disparo` y `estado-de-la-tanda` (quickstart.md §7) | T007, T008 |
| US7 (P3) el error de uso del sondeo, sin traza | `TestGuionDelSondeo`, `TestComprobarElSondeo`, `TestSondear`; las dos primeras órdenes de quickstart.md §8 (T013) | T011 |

## Estrategia de implementación

- **MVP** (lo que hace que el job mida lo que el hito pide y deje de salir en verde con una respuesta sin fuente):
  T001-T006, la lista con sus dos clases y sus formas fijas, el juicio de las líneas esperadas y de `no_se_activan`, la
  respuesta a la pregunta y los umbrales `sin_activar` y `redaccion_no_leida` que deciden (US1 en su parte de la lista,
  US2 y US3 en su parte del juicio, US4, US5).
- **Incremento 2**: T007-T008, una tanda por commit y skill con su definición comprobada en `make ci` (US6).
- **Incremento 3**: T009-T010, la eval 20 y `boe-legislacion` v0.1.4 con la lista entera, lo que el job del cierre mide
  con los umbrales de T006 (US1, US2, US3; SC-001).
- **Incremento 4**: T011, los errores de uso del sondeo (US7).
- **Cierre**: T012-T013, la documentación y la Definition of Done. El job de evals del cierre y la tercera orden de
  quickstart.md §8 quedan fuera del bucle de tareas.

## Controles de umbral → tareas (ADR 0029)

Cada fila de «## Controles de umbral» de plan.md, con la tarea que construye su control y el test que lo ve fallar
cuando la medida pasa del umbral.

| Fila de plan.md | Dónde | Tarea | Test que lo ve fallar |
|---|---|---|---|
| FR-040, SC-001 (expresiones, ≤ 2 de 54) | `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5` | T006 (el umbral con el total de las terminadas y su motivo), T001, T003 y T010 (la lista con las dos clases) | `TestUmbralesDelInforme`, `tres-de-54` → `fallo` |
| FR-041, SC-001 (sin activar, 0) | `evals:boe-legislacion:sin_activar:claude-sonnet-5-5` | T006 | `TestUmbralesDelInforme`, `sin-activar-una` → `fallo` (mutante de T006 con umbral 1) |
| FR-042, SC-001 (clase B, 0) | `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5` | T006 (umbral), T003 (`esDeLaClaseB`), T010 (la familia entera) | `TestUmbralesDelInforme`, `redaccion-no-leida-una` → `fallo` (mutante de T006 con `decide: false`) |
| FR-043, SC-001 (duración ≤ 900 s) | `evals:boe-legislacion:duracion_de_las_sesiones` | T006 (el umbral se queda detrás de los nuevos), T008 (el objetivo 900 de la definición sigue comprobado) | `TestUmbralesDelInforme`, caso de 901 s → `fallo` |
| SC-001 (ninguna sin medir) | `ci:internal/evals/informe_test.go:TestInformeConSesionesSinMedir` | T006 (el recuento de terminadas lo conserva) | `TestInformeConSesionesSinMedir` (mutante de H7.3, comprobado en T006) |
| FR-032, FR-093, SC-002 (36, 11 y 9) | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` | T010 (tres informes y cinco columnas; T003 lo conserva con dos) | `expresiones-calibradas` (mutante de T010 sin `vigente hasta`) |
| FR-033, FR-093, SC-003 (0 en bloques y formas) | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` | T010 (`expresiones-de-la-skill`), T009 (los bloques de `da-3` y las derivadas), T003 (la supresión) | `expresiones-de-la-skill` (premisa sin quitar las formas; cada forma quita algo) y `expresiones-en-los-bloques` |
| FR-021, FR-093, SC-004 (0 en la prosa y 0 fechas) | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` | T010 | `prosa-de-la-skill` con la SKILL.md de `main` (9 defectos) |
| FR-024, FR-095, SC-005 (2 de 2 órdenes con PowerShell) | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, `ci:internal/evals/conjunto_test.go:TestOrdenesParaPowerShell` | T010 | `ordenes-para-powershell` (mutante sin la de `articulos`) y `TestOrdenesParaPowerShell` |
| FR-041, FR-042, FR-045, FR-097, SC-006 (1 → `fallo`, 0 → cumple; 3 y 2 de 54; total 54 con seis sin terminar) | `ci:internal/evals/umbrales_test.go:TestUmbralesDelInforme` | T006 | los casos de contracts/informe-del-job.md §6 (mutante sin el filtro de las terminadas) |
| FR-034, FR-094, SC-007 (cada frase en su clase; 0 marcas en las formas y en la respuesta de FR-022) | `ci:internal/evals/juzgar_test.go:TestJuzgarLasClasesDeLaRespuesta` | T010 (y T004, la sesión de `legal-core`) | `TestJuzgarLasClasesDeLaRespuesta` y el caso de `legal-core` de `TestJuzgar` |
| FR-080, FR-099, SC-009 (0 líneas de `go test` en un error de uso; 0 sesiones) | `ci:internal/evals/sondeo_test.go:TestGuionDelSondeo`, `ci:internal/evals/sondeo_test.go:TestSondear` | T011 | `TestGuionDelSondeo` (mutante sin la rama de `uso.txt`) y `TestSondear` |
| FR-070, FR-071, FR-100, SC-010 (≤ 1 tanda; 0 comprobaciones rojas o de más; la etiqueta vuelve a medir) | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` | T007 (la decisión), T008 (la definición) | `segundo-disparo` (mutantes de T007) y `del-repositorio` (copia de T008 sin `needs.tanda.outputs.medir`) |
| FR-054 (tope del trabajo ≥ 7 285 s) | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` | T008 | `del-repositorio` (copia de T008 con `timeout-minutes: 100`) |
| FR-026, SC-011 (`SKILL.md` < 300 líneas; `description` ≤ 1024) | `ci:Makefile:skills-check` | T010 (lo respeta; el control ya existe en `TestSkillsDelRepositorio`) | `TestSkillsDelRepositorio` |

## Trazabilidad: requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001, FR-002 | T010 (C1, trazado a la causa de research «Causa de raíz», activación) |
| FR-003 | T002, T004, T010 |
| FR-004 | T002, T004 |
| FR-010, FR-011 | T003 (la familia B se busca), T010 |
| FR-012 | T003, T010 |
| FR-020, FR-021, FR-022, FR-025, FR-026 | T010 |
| FR-023 | T004 (la línea con su cita se juzga), T009 (la eval la espera), T010 (la skill la enseña) |
| FR-024 | T010 |
| FR-027 | T012 |
| FR-030 | T001, T003, T010 |
| FR-031 | T001, T003 |
| FR-032 | T003 (35 y 10 se conservan), T010 (36, 11 y 9) |
| FR-033 | T003, T009, T010 |
| FR-034 | T010 |
| FR-035 | T001 |
| FR-036 | T003, T006 |
| FR-037 | T006 (el informe sigue publicando cada respuesta), T010 |
| FR-040, FR-041, FR-042, FR-044, FR-045, FR-046, FR-048 | T006 |
| FR-043 | T006, T008 |
| FR-047 | T006, T013 (y la batería: ningún umbral se rebaja) |
| FR-050, FR-051, FR-055 | T009 |
| FR-052 | T004, T009 |
| FR-053 | T002, T004 |
| FR-054 | T008 (el tope), T009 (20 evals) |
| FR-060 | T005 |
| FR-061 | T006 |
| FR-062 | T004, T005, T006 |
| FR-070, FR-071 | T007, T008 |
| FR-072 | T008, T013 (y la batería: los guiones del workflow no cambian) |
| FR-080, FR-081 | T011 |
| FR-082 | T006, T011 |
| FR-090 | T010 (los informes solo se leen), T013 (y la batería) |
| FR-091 | todas (cada una deja `make ci` en verde); T013 |
| FR-092 | plan.md, «Controles de umbral» (ya escrito); T013 lo comprueba |
| FR-093 | T003, T010 |
| FR-094 | T004, T010 |
| FR-095 | T010 |
| FR-096 | T009 |
| FR-097 | T006 |
| FR-098 | T005 |
| FR-099 | T011 |
| FR-100 | T007, T008 |
| FR-101 | T012 |
| FR-102 | el cierre del workflow, tras la revisión final (no es una tarea: rúbrica i); T006, T008 y T010 dejan el job listo |
| SC-001 | los controles de T006, T008 y T010; se mide en el job del cierre (FR-102) |
| SC-002 | T010 (T003 conserva el de hoy) |
| SC-003 | T003, T009, T010 |
| SC-004, SC-005 | T010 |
| SC-006 | T006 |
| SC-007 | T004, T010 |
| SC-008 | T005 |
| SC-009 | T011 |
| SC-010 | T007, T008 |
| SC-011 | T010, T012, T013 |
| SC-012 | T009 |
