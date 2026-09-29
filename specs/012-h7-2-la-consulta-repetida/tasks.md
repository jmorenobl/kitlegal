# Tasks: H7.2 · La consulta repetida con evidencia coherente, y respuestas sin la maquinaria interna

**Input**: `specs/012-h7-2-la-consulta-repetida/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` 2.6.0 aplicada (umbral de materialidad,
criterio de uso y proporcionalidad en «Gates»); H0-H7, H7.1 y H19 en `main` (`internal/evals` con el formato común,
`grafo_previo`, `hallazgos` y `formas` en el informe; `boe-legislacion` v0.1.1 con la forma `⚠ REDACCIÓN MODIFICADA:`;
`TestGrabacionesDerivadas` con las derivadas del e2e y la del grafo previo de la eval 19 que se retira). **Ninguna tarea
crea el esqueleto ni los gates: `make ci` existe y pasa**, y cada tarea lo deja en verde al terminar.

**Aceptación**: el plan dice **«Aceptación e2e: no aplica»** (plan.md, «Aceptación e2e»; research D19): el hito no
cambia el binario —`cmd/kitlegal` no enlaza `internal/evals` (V17)—, así que no hay comportamiento suyo que un guion
`testscript` describa, y **no hay tarea `[aceptacion]`**. La aceptación del hito es la de la skill: el job de evals que el
workflow lanza en el cierre (SC-001) y el escenario §6 de quickstart.md, que queda fuera del run porque necesita modelo
(SC-002, FR-062). En `make ci` la fijan, sin modelo, los tests que traen las tareas (plan.md, «Tests nuevos»).

**Tests**: obligatorios (constitución §III; spec FR-080 a FR-085; plan.md, «Controles mecánicos»). Las tareas de código
traen su test y la implementación mínima que lo hace pasar, en el mismo diff; los nombres de test son los de las tablas
«Tests nuevos» y «Tests existentes que cambian» de plan.md.

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —su test y su implementación en el mismo
diff, `make ci` en verde por sí solas— T002, T003, T004, T005, T007 y T010. Las demás, cada una por su razón y con su
verificación:

- **T001**, **T006** y **T008** son `[datos]` indivisibles con código de test (plan.md, *Complexity Tracking*;
  precedente de H7.1 T010 y T018): T001 es el esquema de la lista con los casos de `formato_test.go` que lo fijan (el Go
  que la lee llega en T002); T006 es la derivada del grafo previo, que solo existe si la escribe la derivación de
  `grafo_test.go` y que `TestGrabacionesDerivadas` exige con su comprobación; T008 retira la derivada anterior junto con
  su entrada y la comparación carpeta a carpeta, que, introducida antes, dejaría `make ci` en rojo mientras existiera
  esa entrada.
- **T009** es la skill (`SKILL.md` no es código con test propio): su control mecánico son los subtests
  `avisos-de-la-skill`, `hallazgos-de-la-skill` y `expresiones-de-la-skill` (este último, de T003) y `make skills-check`;
  su verificación añade un mutante temporal retirado antes de `make ci`. Su efecto en las respuestas lo mide el job del
  cierre (SC-001).
- **T011** es documentación: su verificación es que cada fichero, test y campo que nombran `CHANGELOG.md` y
  `CONTRIBUTING.md` existe, y `make ci`.
- **T012** es el cierre de la Definition of Done, sin código de producto: escribe la lista de lo tocado, mide la
  cobertura y ejecuta quickstart.md salvo §6; solo añade tests si una cifra de cobertura queda bajo su umbral.

**Modo**: desatendido (ADR 0018). El workflow `hito` ejecuta y verifica las tareas **una a una**, con guardián de diff:
cada tarea declara en su propia línea **todas** las rutas que crea, modifica o retira, y solo toca esas (más `go.mod`,
`go.sum`, `CHANGELOG.md`, el directorio del feature y el `x_test.go` de cada `x.go` declarado). El cierre en la
plataforma (publicar la rama, CI y el job de evals de SC-001 y FR-087) lo hace el workflow tras la revisión final; el
escenario §6 de quickstart.md (SC-002) queda para después del run, al leer el informe final (FR-062).

## Formato: `- [ ] Tnnn [etiqueta?] [Story?] Descripción — FR/SC; rutas`

- **[aceptacion]**: ninguna (plan.md, «Aceptación e2e: no aplica»).
- **[datos]**: la tarea crea, modifica o retira ficheros bajo `schemas/` o bajo un `testdata/`. Son **tres**: T001 (el
  esquema de la lista), T006 (la derivada `lcsp-a1-30-redaccion-original`) y T008 (la retirada de
  `lpac-a21-version-anterior`). H7.2 no graba nada de ninguna fuente: no hay manifiesto `grabaciones.json` nuevo, ni test
  de grabación, ni nada que grabar para el paso `grabar_datos` (plan.md, «Datos externos»; FR-014): la eval nueva usa las
  grabaciones de H4 de `BOE-A-2017-12902` (bloque `a1-30`, índice y metadatos; research V16), y la derivada la produce
  código del repositorio a partir de ellas (FR-010), nunca una persona ni la memoria.
- **[P]**: ninguna tarea lo lleva; en el bucle del workflow todo va en secuencia (ver «Oportunidades de paralelismo»).
- **[Story]**: US1 quien pregunta lee la norma, no la herramienta · US2 la eval de la consulta repetida plantea una
  situación que puede darse · US3 una expresión prohibida hace que la sesión no pase · US4 nada de lo dicho en otra
  conversación (lo cubren la familia `otra_conversacion` de T003 y la viñeta C5 de T009, que llevan la etiqueta de su
  historia principal) · US5 cualquiera repite la aceptación de la consulta repetida. La documentación y el cierre no
  llevan historia.

## Batería de verificación por tarea

- **`make ci` en primer plano** y la tarea marcada `[X]` en el mismo turno en que termina en verde: formato · lint
  (`depguard`, `forbidigo`, `gosec`, `misspell`, `revive`, `dupl`…, también sobre los ficheros con la etiqueta `evals`) ·
  tests con `-race` · `test-integration` · `test-tiempos` · `vuln` · `schema-check` · `skills-check` · secretos ·
  módulos. Ningún registro de la verificación en la raíz del repositorio. Las salidas de `go test` y de `make` se leen
  con `rtk proxy` o con sondas positivas, porque el proxy de la sesión las resume.
- **Sin red ni modelo**: ningún test abre una conexión; `boe` responde desde las grabaciones de H4 y H5
  (`UnionDeGrabaciones()`). Ninguna tarea graba, ejecuta `make evals`, `make verify-sources`, `claude` ni el escenario
  §6 de quickstart.md.
- **Sin personas a mitad**: ninguna tarea la hace, la revisa ni la desbloquea una persona.
- **Ningún test escribe en `~/.cache/kitlegal`**: toda caché y todo grafo de test van bajo `t.TempDir()`.
- **Sin atajos** (plan.md, «Comprobación contra la rúbrica», g): ningún `//nolint` nuevo, ningún `t.Skip`, TODO ni
  error silenciado; ninguna exclusión de lint nueva. Si `dupl`, `gocyclo` o `paralleltest` marcan algo, se reestructura.
- **Proporcionalidad** (constitución, «Gates»): ninguna tarea añade un caso, un mensaje ni un test para un estado sin vía
  real; una lista, una derivada o una respuesta que no se pueden leer siguen las reglas que ya hay (fichero mal formado,
  sesión ilegible, test que falla), sin caso propio (plan.md, «Trazabilidad»).
- **Frontera de FR-056**: `internal/evals` y el esquema común de eval no cambian más allá de la lista, la derivación real
  del grafo previo y el punto de entrada de la comprobación del quickstart; `Juzgar` no cambia de firma (research D5).
- **Vocabulario**: los comentarios llevan tilde; los caracteres que no son ASCII de los tests se escriben con su escape
  de Go (`\xc3\xb3`…), no con `\u`.
- **Sondas y mutantes momentáneos** (T003, T008, T009): se crean y se retiran dentro de la tarea, antes de `make ci` y
  del commit.

---

## Fase 1: US3 — Una expresión prohibida hace que la sesión no pase, sin juicio de ningún modelo (P1; va primero por el orden de dentro afuera)

**Objetivo**: el formato común de eval expresa una lista por skill; `boe-legislacion` tiene la suya, calibrada contra
las 93 respuestas de H7.1; `Juzgar` la aplica a las evals que activan la skill y el informe publica lo encontrado por
sesión y el recuento por modelo. Va antes que la eval nueva y que `SKILL.md` porque las dos se juzgan con ella (plan.md,
«Orden de implementación», pasos 1-4). **Prueba independiente**: el test del esquema de `formato_test.go`,
`TestLeerConjunto`, `TestExtraerExpresionesProhibidas`, `TestJuzgarLasExpresionesProhibidas`, los tests del informe y
los subtests `expresiones-calibradas`, `expresiones-en-los-bloques` y `expresiones-de-la-skill` de
`TestEvalsDelRepositorio`; en el job del cierre, las expresiones por sesión y el recuento por modelo (SC-001).

- [X] T001 [datos] [US3] El esquema de la lista de expresiones prohibidas con los casos que lo fijan, y **nada más** (contracts/lista-y-juicio.md §1; data-model §1; research D1, D2; plan.md paso 1): schemas/expresiones-prohibidas.yaml.json tal cual el contrato —`$schema` 2020-12, `$id` `https://kitlegal.es/schemas/expresiones-prohibidas.yaml.json`, `title`, objeto con `additionalProperties: false` y `maquinaria` y `otra_conversacion` obligatorias, cada una `$ref` a `$defs.expresiones`: lista con `minItems: 1` de cadenas con el patrón `^[^\s*_]+( [^\s*_]+)*$`—; en internal/evals/formato_test.go, un test del esquema publicado que lo lee de su ruta, lo compila con `skills.CompilarEsquema` y valida contra él con `skills.ValidarDocumentoYAML` documentos escritos en el test: una lista con las dos familias valida; sin `maquinaria`, sin `otra_conversacion`, con una familia vacía, con una clave de más y con una expresión con un blanco en un extremo o con `*` no validan (una lista mal formada es un fichero mal formado, FR-055); ninguna carpeta de evals lleva lista todavía y no hay código de producto (el Go que la lee llega en T002); en .golangci.yml, `conversacion` en las `ignore-rules` de `misspell`, entre `controles` y `defectos` y con su comentario, como las demás: la clave `otra_conversacion` que fija el contrato la lee misspell como «conversation», también entre guiones bajos y guiones, así que los documentos del test no pasan `make lint` sin ella, y T002 la escribe en la etiqueta YAML de `OtraConversacion`, que no se puede partir (redelimitada tras el intento 1, que lo encontró; su nota da la entrada comprobada); verificación: `make ci` en verde (`schema-check` no la compara: no sale de `--describe`) — FR-050, FR-055, FR-082; SC-006; rutas: schemas/expresiones-prohibidas.yaml.json, internal/evals/formato_test.go, .golangci.yml
- [X] T002 [US3] La lista en el formato común de eval: su lectura, su comparación y su lugar en el conjunto (contracts/lista-y-juicio.md §1 y §3; data-model §1 y §2; research D3, D5, D8; plan.md paso 2): internal/evals/prohibidas.go con `ExpresionesProhibidas{Maquinaria, OtraConversacion []string}` (`yaml:"maquinaria"`, `yaml:"otra_conversacion"`), el esquema publicado de T001 compilado una sola vez como el de eval, la lectura con el lector común de documentos YAML (`skills.ValidarDocumentoYAML`: una clave repetida es un defecto con sus dos líneas) y `ExtraerExpresionesProhibidas(texto string, lista ExpresionesProhibidas) []string` —las expresiones de la lista, en su orden (maquinaria y después otra conversación) y sin repetir, cuya expresión regular casa en algún punto del texto; `nil` si ninguna—, con cada expresión compilada una sola vez como `(?:^|[^\p{L}\p{N}])`, sus palabras de `strings.Fields` como `(?i:<regexp.QuoteMeta(palabra)>)` unidas por `entrePalabrasDeAviso` de internal/evals/avisos.go (la tolerancia de H5.1), y `(?:$|[^\p{L}\p{N}])`; internal/evals/conjunto.go: `LeerConjunto` reconoce `expresiones-prohibidas.yaml` por su nombre exacto —no es un fichero de eval—, la lee y la valida, y la deja en `Conjunto.Prohibidas` y copiada en `Eval.Prohibidas` de cada eval de la carpeta; si la entrada no es un fichero regular, no se puede leer o no valida, es un `FicheroMalFormado` cuyo error empieza por `expresiones-prohibidas.yaml`, y las evals de la carpeta se leen sin lista; internal/evals/formato.go: `Eval.Prohibidas ExpresionesProhibidas` con `yaml:"-"`, como `Fichero`; internal/evals/doc.go dice que el paquete lee también la lista; `Juzgar` y el informe no cambian todavía (T004, T005); tests: `TestExtraerExpresionesProhibidas` en internal/evals/prohibidas_test.go, con una lista escrita en el test, sobre cada fila de la tabla de contracts/lista-y-juicio.md §3 —mayúsculas, dos espacios, énfasis de Markdown alrededor y dentro de las palabras, `graph check` entre comillas invertidas, `(código 0)`, «te habría confirmado», «te confirmé.», la línea `⚠ REDACCIÓN MODIFICADA:` con sus dos fechas, `⚠ NORMA DEROGADA:`, «No hay avisos de vigencia sobre este bloque.» y «No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.» sin ninguna, `hallazgoss` sin ninguna y el salto de línea que parte una expresión—, el orden de la lista y sin repetir; `TestLeerConjunto` en internal/evals/conjunto_test.go con una carpeta con la lista bien formada (en `Conjunto.Prohibidas` y en cada `Eval.Prohibidas`), con una mal formada (un `FicheroMalFormado` que la nombra y evals sin lista) y sin lista (como antes del hito); verificación: `make ci` en verde con `TestEvalsDelRepositorio` sin cambios (ninguna carpeta del repositorio tiene lista todavía) — FR-050, FR-051, FR-055, FR-056, FR-082; SC-006; US3.4, US3.5, US3.9; rutas: internal/evals/prohibidas.go, internal/evals/conjunto.go, internal/evals/formato.go, internal/evals/doc.go, internal/evals/conjunto_test.go
- [X] T003 [US3] La lista de `boe-legislacion`, calibrada (contracts/lista-y-juicio.md §2 y §6; research D4, D9, D10, D11, V1; plan.md paso 4): evals/boe-legislacion/expresiones-prohibidas.yaml con el contenido exacto de contracts/lista-y-juicio.md §2 (su cabecera y las 38 expresiones: 22 de `maquinaria` y 16 de `otra_conversacion`); en internal/evals/conjunto_test.go, tres subtests de `TestEvalsDelRepositorio` que leen la lista con su `LeerConjunto` y exigen antes su premisa, para no pasar en vacío (la lista no está vacía y hay qué mirar): `expresiones-calibradas` lee ../../specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json en un tipo local (`evals[].eval`, `evals[].respuesta`; premisa, 93 respuestas), aplica a cada respuesta `ExtraerExpresionesProhibidas` y exige, por las dos cifras del fichero de la eval —nunca su nombre: el de la eval 19 retirada no puede aparecer en un test (FR-020)— y por familia, exactamente las 35 del reparto de FR-084 —`maquinaria` en 02 (1), 03 (3), 04 (3), 05 (3), 06 (3), 07 (3), 08 (2), 09 (2), 13 (3), 14 (3), 15 (3), 16 (2) y 17 (3); `otra_conversacion` en 19 (1)— y ninguna en las otras 58, con un mensaje que nombra cada eval cuyo reparto no es el esperado; `expresiones-en-los-bloques` lee con `boe articulo --json`, en proceso y sobre `UnionDeGrabaciones()` con una caché temporal, cada bloque de `ConsultasNecesarias` de las evals y, sobre las grabaciones de cada grafo previo, cada bloque de sus comandos, y exige que ninguna expresión case con el `texto` de ninguno (premisa: algún bloque leído, con el recuento en el mensaje); `expresiones-de-la-skill` exige que ninguna case con la forma escrita (`formaEscrita`) de cada etiqueta de `boe.EtiquetasDeAviso()` y de `grafo.EtiquetasDeHallazgo()`, ni con el contenido de ningún bloque de código `text` de skills/boe-legislacion/SKILL.md, que son los textos que la skill enseña a escribir (premisa: hay bloques `text`, entre ellos el de la línea `⚠ REDACCIÓN MODIFICADA:`); `Juzgar` todavía no la aplica (T004); verificación: `make ci` con `skills-check` en verde, y con `hallazgos` quitada un momento de la lista, `expresiones-calibradas` falla nombrando las evals cuyo reparto cambia (se restaura antes de `make ci`) — FR-043, FR-050, FR-051, FR-084, FR-085; SC-003, SC-004; US4.2; rutas: evals/boe-legislacion/expresiones-prohibidas.yaml, internal/evals/conjunto_test.go
- [ ] T004 [US3] `Juzgar` aplica la lista a las evals que activan la skill (contracts/lista-y-juicio.md §4; data-model §2; research D5, D6; plan.md paso 3): internal/evals/juzgar.go: `ResultadoDeEval.ExpresionesProhibidas []string` (`json:"expresiones_prohibidas"`, detrás de `territorio_ausente`; `[]` en el JSON si no hay ninguna, research V4); si `eval.Activa` y `eval.Prohibidas` no está vacía, las de `ExtraerExpresionesProhibidas(sesion.Respuesta, eval.Prohibidas)`, un motivo `expresión prohibida: <expresión>` por cada una, detrás de los del territorio ausente y delante del del modelo que pone `EscribirInforme`, y `Pasa` exige además que no haya ninguna; con `activa: false` o en una skill sin lista, la lista queda vacía y el juicio es el de antes del hito; la firma de `Juzgar` no cambia y nada más cambia en él; tests en internal/evals/juzgar_test.go: `TestJuzgarLasExpresionesProhibidas`, con la lista del repositorio leída con `LeerConjunto` y una eval positiva del repositorio, sobre una sesión que cumple todo lo demás: sin expresiones (pasa, ninguna); «Sin hallazgos en la memoria de consultas» (`memoria de consultas` y `hallazgos`, sus dos motivos en orden, no pasa); «te confirmé» (la familia `otra_conversacion`, no pasa); la misma expresión con otras mayúsculas, con espacios de más y dentro de un énfasis de Markdown (encontrada igual); la línea `⚠ REDACCIÓN MODIFICADA:` con sus dos fechas y nada más de la lista (ninguna, pasa); una eval de no activación del repositorio y una de `legal-core`, sin lista, con una respuesta con expresiones (se juzgan como antes del hito); y el campo en el JSON como lo codifica `EscribirInforme`; los tests que comparan un `ResultadoDeEval` entero ganan el campo, en internal/evals/juzgar_test.go y, si alguno lo compara literal, en internal/evals/informe_test.go; verificación: `make ci` en verde — FR-051, FR-052, FR-054, FR-083; SC-006; US3.1-US3.6; rutas: internal/evals/juzgar.go, internal/evals/juzgar_test.go, internal/evals/informe_test.go
- [ ] T005 [US3] El informe publica las expresiones por sesión y el recuento por modelo (contracts/lista-y-juicio.md §5; data-model §3; research D7, D8; plan.md paso 3): internal/evals/informe.go: `RecuentoDeExpresiones{Modelo string; ConAlguna int; Respuestas int}` (`modelo`, `con_alguna`, `respuestas`) e `Informe.ExpresionesProhibidasPorModelo` (`expresiones_prohibidas_por_modelo`, detrás de `tasas`): un elemento por modelo del job, el que decide y después los informativos en su orden, con `Respuestas` = sesiones juzgadas (no ilegibles) de las series planificadas cuya eval activa la skill y `ConAlguna` = las de esas con alguna expresión; la sesión de la prueba de red y las demás series no planificadas se juzgan con la lista y publican sus expresiones, pero no cuentan; una skill sin lista (`Conjunto.Prohibidas` vacía) da `[]`; en informe.md, la columna «Expresiones prohibidas» de la tabla «Sesiones», entre «Territorio ausente» y «Resultado» (las encontradas separadas por `, ` o `ninguna`), y detrás de «## Tasas por eval» la sección «## Expresiones prohibidas por modelo» con la tabla `Modelo | Respuestas con alguna expresión | Respuestas en evals que activan la skill` o, sin lista, el párrafo `la skill no tiene lista de expresiones prohibidas`; la regla del veredicto no cambia (FR-054); tests en internal/evals/informe_test.go, con los casos nuevos armados en `t.TempDir()` sobre copias de los casos versionados del paquete, que no se tocan (como `TestInformeConProhibidos`): una carpeta de evals con lista y sesiones con y sin expresiones da `expresiones_prohibidas` por sesión en el JSON y en la columna, y el recuento por modelo con el modelo que decide y uno informativo; el mismo caso con una sesión de la prueba de red con expresiones da el mismo recuento y publica sus expresiones en su sesión; una skill sin lista da `[]` y el párrafo; una lista mal formada queda en `ficheros_mal_formados` y el veredicto es `fallo`; con tres repeticiones y umbral 2, una serie de una eval que decide con dos sesiones con alguna expresión da `fallo`, y la misma serie de una eval informativa publica su tasa sin decidir; las tablas «Sesiones» de los casos existentes ganan la columna (`ninguna`) y su informe, el párrafo; verificación: `make ci` en verde — FR-053, FR-054, FR-055, FR-056; SC-001 (lo que el informe publica); US3.7, US3.8, US3.9; rutas: internal/evals/informe.go, internal/evals/informe_test.go

---

## Fase 2: US2 — La eval de la consulta repetida plantea una situación que puede darse (P1)

**Objetivo**: la eval 19 pasa al art. 118 LCSP, con un grafo previo derivado por código de la grabación de H4 (la
redacción original) y la caché con la vigente; el control de derivaciones exige a toda derivada del grafo previo una
redacción que la grabada trae; se retira la eval del art. 21 LPAC con su grafo previo escrito a mano. **Prueba
independiente**: `TestGrabacionesDerivadas` y `TestGrabacionesDerivadasInventadas`, la subprueba `grafo-previo` de
`TestEvalsDelRepositorio`, `TestEstadoPrevioDeLaRedaccionCambiada` y quickstart.md §2, §4 y §5; en el job del cierre, la
eval 19 con `⚠ REDACCIÓN MODIFICADA:` y su tasa (SC-001).

- [ ] T006 [datos] [US2] La derivada del grafo previo de la eval nueva, producida por código y con su control, y **nada más** (contracts/eval-y-derivada.md §2 y §3; data-model §5 y §6; research D13, V15, V18; plan.md paso 5): en internal/app/grafo_test.go, (a) la derivación: de la grabación de H4 del mismo nombre del paquete `boe` (la del bloque `a1-30` de `BOE-A-2017-12902`) y una fecha de vigencia, la lee en un tipo local con los campos del formato de grabación de `httpx` en su orden (`formato`, `grabado_en`, `peticion{metodo, url, cabeceras}`, `respuesta{estado, cabeceras, cuerpo}`), localiza con `encoding/xml` las `<version>` hijas del `<bloque>` en `respuesta.cuerpo`, quita los bytes desde el final del `</version>` de la de esa fecha hasta el final del de la última (las redacciones posteriores y el blanco que las precede) y la escribe con el codificador de las grabaciones (sangrado de dos espacios, sin escapar HTML, salto de línea final), con su premisa: con la fecha de la última redacción devuelve la grabación byte a byte; (b) la bandera `-actualizar-derivadas`, que antes de comprobar escribe cada derivada del grafo previo desde su grabación, como `-actualizar-esquemas`; (c) la lista de derivadas del grafo previo, de un tipo propio y aparte de `grabacionesDerivadas()` —subcarpeta de `grafosPreviosDeLasEvals`, fichero, argumentos de `boe` y fecha de vigencia de la redacción que da, sin comprobación propia—, con la entrada `lcsp-a1-30-redaccion-original` (`articulo BOE-A-2017-12902 a1-30`, `20180309`); (d) `TestGrabacionesDerivadas` aplica a toda entrada de esa lista la comprobación 1 (el fichero es, byte a byte, la derivación sobre su grabación) y la 2 (servida en lugar de la grabación, `boe articulo … --json` da un `Articulo` igual, campo a campo, a exactamente uno de los que da `boe` sobre la grabación reducida a cada una de sus redacciones, y es el de la fecha declarada: `fecha_vigencia` 20180309, `fecha_version` 20171109, `norma_modificadora` `BOE-A-2017-12902`, el texto original y su `hash_texto`, tal como los trae la grabada), escrita como una función que devuelve un error que nombra el fichero; sus ficheros entran en la comparación de hoy, las dos carpetas juntas (la comparación carpeta a carpeta llega en T008), y las cuatro derivadas del e2e y la entrada que retira T008 no cambian; (e) `TestGrabacionesDerivadasInventadas`: tres derivadas armadas desde la grabación en `t.TempDir()` —la fecha 20151002 con el texto original, un párrafo de más al final de la redacción original, y las dos cosas, como la retirada— no pasan la comprobación 2, cada una con un error que nombra su fichero y dice que la redacción que da no es ninguna de las de la grabada; después, la derivada testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_a1-30.json, escrita con `go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/ -args -actualizar-derivadas` y nunca a mano; ninguna grabación nueva ni red (FR-014: el bloque, el índice y los metadatos de `BOE-A-2017-12902` ya están grabados en H4); verificación: una segunda ejecución de esa orden deja el fichero con la misma huella, `TestGrabacionesDerivadas` (con el subtest `lcsp-a1-30-redaccion-original`) y `TestGrabacionesDerivadasInventadas` en verde, y `make ci` en verde — FR-010, FR-011, FR-012, FR-013, FR-014, FR-081; SC-005; US2.3, US2.4; rutas: internal/app/grafo_test.go, testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/
- [ ] T007 [US2] La eval nueva de la consulta repetida en lugar de la 19 anterior, con su estado previo comprobado en `make ci` (contracts/eval-y-derivada.md §1, §4 y §5; data-model §4; research D12, D14, D15, V16, V21; plan.md paso 6): evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml con el contenido exacto de contracts/eval-y-derivada.md §1 (su comentario, la pregunta literal de FR-001, `activa: true`, `informativa: true`, `grafo_previo` `lcsp-a1-30-redaccion-original` con el comando `boe BOE-A-2017-12902 a1-30`, los comandos del bloque y de `graph check` con `norma: BOE-A-2017-12902`, `graph show` prohibido, la cita `BOE-A-2017-12902 a1-30` y `hallazgos: [version-obsoleta]`), y se retira evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml; en internal/evals/conjunto_test.go, `compruebaElGrafoPrevio` (subtest `grafo-previo`) sigue exigiendo lo de H7 y gana, para cada eval con `grafo_previo`, lo que hace la sesión: anota la `fecha_vigencia` de la `BloqueVersion` que deja el grafo previo; lee en proceso cada bloque de sus comandos con `boe articulo <norma> <bloque> --json` sobre `UnionDeGrabaciones()`, con una caché temporal y entregando al grafo de la sesión; ejecuta en proceso `graph check <norma> <bloques> --json` por cada norma de sus comandos; y exige código 0 en las dos, que las clases de `data.hallazgos` sean exactamente las de `hallazgos` de la eval y que cada `version-obsoleta` lleve en `fecha_vigencia` la anotada y en `fecha_vigencia_reciente` la leída (con la 19, uno, de 20180309 a 20200206); los tests que nombran lo retirado siguen comprobando lo mismo con lo nuevo (FR-021), sin desactivar, saltar ni retirar ninguno: internal/evals/formato_test.go (los documentos sintéticos con `grabaciones: lpac-a21-version-anterior` pasan a `lcsp-a1-30-redaccion-original`), internal/evals/consultas_test.go (el grafo previo sintético, igual; el nombre sintético `06-lpac-articulo-21-redaccion-cambiada.yaml` se queda), internal/evals/juzgar_test.go (`ficheroDeLaConsultaRepetida` pasa a `19-lcsp-contrato-menor-redaccion-cambiada.yaml`, y `evalDeLaConsultaRepetida` y los juicios que la usan, a la forma de la nueva: `BOE-A-2017-12902`, `a1-30`, la comprobación con la norma) e internal/evals/preparar_test.go (`TestEstadoPrevioDeLaRedaccionCambiada`, que lee la eval 19 del repositorio por `ficheroDeLaConsultaRepetida`, pasa a exigir el ámbito `BOE-A-2017-12902`, a leer en la sesión el bloque `a1-30` y a encontrar exactamente un `version-obsoleta`, sobre la redacción de 20180309 con 20200206 como la reciente; `TestPrepararGrafoPrevio`, con su derivada sintética escrita por el propio test, no cambia); verificación: `make ci` con `skills-check` en verde —`formato`, `conjunto` (19 evals, 10 positivas que deciden, informativas de la 13 a la 19, reglas sin cambios), `grabado`, `grafo-previo` y `expresiones-en-los-bloques`, que cuenta ya la redacción original del bloque `a1-30`—, y quickstart.md §5 da `19`, la pregunta literal, `grabaciones: lcsp-a1-30-redaccion-original` y `- version-obsoleta` — FR-001, FR-002, FR-003, FR-004, FR-020, FR-021, FR-030, FR-085; SC-004, SC-007; US2.1, US2.2; rutas: evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml, evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml, internal/evals/conjunto_test.go, internal/evals/formato_test.go, internal/evals/consultas_test.go, internal/evals/juzgar_test.go, internal/evals/preparar_test.go
- [ ] T008 [datos] [US2] Retirada del grafo previo de la eval 19 anterior y comparación carpeta a carpeta, y **nada más** (contracts/eval-y-derivada.md §3 y §5; data-model §8; research D13, D15; plan.md paso 7): se retira testdata/evals/grafo-previo/lpac-a21-version-anterior/ entera; en internal/app/grafo_test.go salen su entrada de `grabacionesDerivadas()` y `parrafoDeLaVersionAnterior`, y los comentarios de `grabacionesDerivadas()` y de `TestGrabacionesDerivadas` dejan de nombrarlas (`versionDelArticulo21` se queda: la usan `version-posterior` y `version-ulterior`); con ellas, `TestGrabacionesDerivadas` compara carpeta a carpeta y en los dos sentidos —los ficheros de `derivadasDelE2E` con las entradas de `grabacionesDerivadas()`, y los de `grafosPreviosDeLasEvals` con las de la lista de derivadas del grafo previo de T006—, de modo que la carpeta, y no la clase de la entrada, decide la comprobación: una entrada del e2e con carpeta de grafo previo falla nombrada (es una entrada del e2e sin fichero en su carpeta, y su fichero, uno del grafo previo sin entrada); las cuatro derivadas del e2e siguen con su clase y su comprobación sin cambios (FR-013); verificación: con una entrada del e2e apuntada un momento a una carpeta de grafo previo, `TestGrabacionesDerivadas` falla nombrándola (se retira antes de `make ci`); quickstart.md §4 da `retirada la eval`, `retirado su grafo previo` y `coincidencias: 1`; y `make ci` en verde — FR-011, FR-012, FR-013, FR-020, FR-081; SC-005, SC-007; US2.5; rutas: testdata/evals/grafo-previo/lpac-a21-version-anterior/, internal/app/grafo_test.go

---

## Fase 3: US1 y US4 — Quien pregunta lee la norma, no la herramienta, y nada de lo dicho en otra conversación (P1, P2)

**Objetivo**: `boe-legislacion` v0.1.2, con los cambios C1-C9 trazados a la causa de raíz del ruido (research, «Causa de
raíz del ruido»; plan.md, «Cambios de `SKILL.md` trazados a la causa de raíz»). Va detrás de sus evals —la lista (T003),
su juicio (T004) y la eval nueva (T007)— (Definition of Done §1.10). **Prueba independiente**: `make skills-check` y los
subtests `avisos-de-la-skill`, `hallazgos-de-la-skill` y `expresiones-de-la-skill`; su efecto, en el job del cierre (la
lista en todas las respuestas de las evals que activan la skill, ≤ 2 de 51 con Sonnet 5 y ≤ 1 de 30 con Haiku 4.5;
SC-001).

- [ ] T009 [US1] `boe-legislacion` v0.1.2: la respuesta no cuenta la comprobación ni lo dicho en otra conversación, con cada cambio trazado a la causa de raíz (contracts/skill-boe-legislacion.md §1 y §2; research «Causa de raíz del ruido» y D16; plan.md, «Cambios de `SKILL.md` trazados a la causa de raíz»): skills/boe-legislacion/SKILL.md cambia solo en C1-C9, con los textos del contrato (la redacción final puede ajustar el orden de las palabras, no lo que piden): en el paso 5, la primera viñeta sin «traslada lo que encuentre» —si da `version-obsoleta`, la forma fija de «Memoria de consultas»; si no, nada de ella— (C4), y detrás dos viñetas nuevas: «La respuesta empieza por lo que se pregunta», con su porqué (quien pregunta no ve las órdenes ni lo que devuelven) y todo lo que la respuesta no nombra salvo la forma `⚠ REDACCIÓN MODIFICADA:` —la memoria de consultas, `kitlegal graph` ni sus verbos, los códigos de salida, los hallazgos, las clases `version-obsoleta` y `fuente-caducada`, el JSON ni el sobre— (C1, C2, C3), y «Nada de otra conversación», con por qué, sin `version-obsoleta`, no se dice nada de lo consultado antes (C5); en «Memoria de consultas», la viñeta de `version-obsoleta` con las dos fechas de vigencia tal como las da el hallazgo (`AAAAMMDD`) en la misma línea que la forma, la superada y la leída (C6), y la última viñeta, que remite al paso 5 y, con otro código, a la regla 7 (C9); la regla 2 dice lo que no se pudo consultar por lo que significa para quien pregunta, sin el código (C8); la regla 7, reescrita, pide responder con el texto de `kitlegal boe` y decir, sin nombrar la memoria de consultas, `kitlegal graph` ni el código, «No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.», en un bloque de código `text` (C7); se quedan H7 FR 081, el código 0 como resultado, H7.1 FR 040, 041, 043 y 045, la forma de la cita y la de los avisos, la tabla de comandos generada, el frontmatter, los pasos 1 a 4 y las reglas 1, 3, 4, 5 y 6; menos de 300 líneas, sin nombrar evals, el job ni modelos (H5 FR 077); verificación: `make ci` con `skills-check` en verde —`avisos-de-la-skill`, `hallazgos-de-la-skill` y `expresiones-de-la-skill`, que mira ya también el bloque de la regla 7—; con `memoria de consultas` puesta un momento en ese bloque, `expresiones-de-la-skill` falla (se restaura antes de `make ci`); y `wc -l` de la skill por debajo de 300 — FR-040, FR-041, FR-042, FR-043, FR-044, FR-045, FR-046, FR-047; SC-001 (su efecto, en el job del cierre), SC-008; US1.1-US1.4, US4.1; rutas: skills/boe-legislacion/SKILL.md

---

## Fase 4: US5 — Cualquiera repite la aceptación de la consulta repetida (P2)

**Objetivo**: la comprobación mecánica de las dos respuestas del escenario §6 de quickstart.md, con el mismo código que
el juicio. El escenario ya está escrito en quickstart.md; ninguna tarea lo ejecuta (FR-062). **Prueba independiente**:
`TestCondicionesDeLaConsultaRepetida` en `make ci`; el escenario §6, fuera del run, al leer el informe final (SC-002).

- [ ] T010 [US5] La comprobación mecánica de la consulta repetida del quickstart (contracts/comprobacion-del-quickstart.md; data-model §7; research D17, D18, V13; plan.md paso 9): internal/evals/consulta_repetida.go con `comprobarConsultaRepetida(primera, segunda Sesion, prohibidas ExpresionesProhibidas, superada, leida string) []string`, que devuelve una línea por lo que falla, en el orden y con los textos de contracts/comprobacion-del-quickstart.md §2 —cada conversación terminada (`Terminada`, con `MotivoSinTerminar`); 1, alguna línea de la primera respuesta casa con la forma de `version-obsoleta` de `formasDeHallazgo`, la del juicio, y contiene las dos fechas como palabras (`contieneComoPalabra`); 2, la segunda respuesta no la lleva (`ExtraerHallazgos`); 3, ninguna de las dos lleva expresiones de la lista (`ExtraerExpresionesProhibidas`)—, y una conversación que no terminó no se mira más; sin réplica de la lista ni de la comparación, y sin criterios nuevos en `Juzgar`, campos en el formato de eval ni en el informe (FR-056); en internal/evals/job_test.go (etiqueta `evals`, fuera de `make ci` y del run), `TestComprobarConsultaRepetida` con las banderas nuevas `-primera`, `-segunda`, `-fecha-superada` y `-fecha-leida` y la `-skill` de siempre, todas obligatorias (`exigirBanderas`): lee la lista con `LeerConjunto` de la carpeta de evals de la skill —con ficheros mal formados, falla nombrándolos, como `TestPlanDeSesiones`—, cada conversación con `LeerSesion` (un error, `la <primera|segunda> conversación no se ha podido leer: <error>`), falla con todas las líneas de `comprobarConsultaRepetida`, una por renglón, y, sin ninguna, registra `se cumplen las tres condiciones: la forma con <superada> y <leida> en la primera respuesta, sin ella en la segunda, y ninguna expresión prohibida en las dos`; tests: `TestCondicionesDeLaConsultaRepetida` en internal/evals/consulta_repetida_test.go, con sesiones construidas en el test y la lista del repositorio: las dos respuestas buenas, ninguna línea; la primera sin la forma; la forma en una línea y las fechas en otra; la segunda con la forma; una expresión en la primera; una en la segunda; la primera sin terminar; y tres fallos a la vez, sus tres líneas en orden; verificación: `make ci` en verde (el lint carga la etiqueta `evals`, research V19), `go vet -tags evals ./internal/evals/` sin nada, `go test -tags evals -list '^TestComprobarConsultaRepetida$' ./internal/evals/` lo lista, y la orden de la comprobación de quickstart.md §6 sin sus banderas falla nombrando las que faltan; el escenario §6 no se ejecuta (necesita modelo, FR-062) — FR-056, FR-060, FR-061, FR-062; SC-002 (su comprobación); US5.3; rutas: internal/evals/consulta_repetida.go, internal/evals/consulta_repetida_test.go, internal/evals/job_test.go

---

## Fase 5: Documentación y Definition of Done

**Propósito**: `CHANGELOG.md` y `CONTRIBUTING.md` sin nada que el hito deje falso, y el cierre de la Definition of Done
que aplica al hito.

- [ ] T011 Documentación que el hito deja falsa (contracts/skill-boe-legislacion.md §5; research D20, V21; plan.md paso 10): CHANGELOG.md en *Unreleased* —H7 y H7.1 no han salido en ninguna release, así que describe el comportamiento final—: la entrada de `boe-legislacion` v0.1.1 pasa a v0.1.2 y dice, para quien la usa, que la respuesta empieza por lo que se pregunta y no cuenta la comprobación (ni la memoria de consultas, `graph`, códigos, hallazgos, clases, JSON ni sobre), salvo la línea `⚠ REDACCIÓN MODIFICADA:` con sus dos fechas tal como las da el hallazgo; que no habla de lo dicho en otra conversación ni, sin cambio, de lo consultado antes; que un fallo de `graph check` se dice como «no se ha podido comprobar si la redacción ha cambiado desde una consulta anterior» y uno de `kitlegal boe`, por lo que significa y sin el código; y que sustituye a la v0.1.1, que no llegó a publicarse (FR-048); la entrada de la eval de la consulta repetida pasa a la eval nueva sobre el art. 118 de la LCSP, con la redacción original derivada de la grabación en el grafo y la vigente en la caché, y dice que retira la del art. 21 de la LPAC, cuyo grafo previo sembraba una redacción escrita a mano; y una entrada nueva, la lista de expresiones prohibidas del formato común de eval, que el informe comprueba sin modelo en las evals de `boe-legislacion` que activan la skill, por sesión y con el recuento por modelo, y que decide en las que deciden (FR-086); CONTRIBUTING.md, «Formato común de eval»: el fichero `expresiones-prohibidas.yaml`, opcional y uno por skill, su esquema, sus dos familias, la comparación por la forma con la tolerancia de H5.1 y la limitación declarada (lo dicho con otras palabras no se detecta), y la condición nueva de «Una sesión de una eval pasa si…» (FR-050); «Job de evals»: `expresiones_prohibidas` por sesión y `expresiones_prohibidas_por_modelo`, con la sesión de la prueba de red fuera del recuento (FR-053); README.md y CLAUDE.md no cambian (research D20); verificación: cada fichero, test y campo que nombran existe (`ls`, `go test -list`, búsqueda en internal/evals/informe.go) y `make ci` en verde — FR-048, FR-050, FR-053, FR-086; SC-008; rutas: CHANGELOG.md, CONTRIBUTING.md
- [ ] T012 Cierre de la Definition of Done, sin código de producto: specs/012-h7-2-la-consulta-repetida/cierre.md con (1) la lista, sacada de `git diff --name-status main`, de los esquemas y datos de prueba creados, modificados o retirados en el hito (el esquema de la lista, la derivada `lcsp-a1-30-redaccion-original` y la carpeta `lpac-a21-version-anterior` retirada) y de las evals (la lista, la eval 19 nueva y la retirada), para la capa 3 del informe final; (2) la cobertura sobre `coverage.out` y `coverage-integration.out` de `make ci` con `go tool cover -func` —global ≥ 70 % e `internal/core/**` ≥ 85 %—; (3) el resultado de quickstart.md §1 a §5 y §7 —§1 con `--- PASS` en cada subtest nombrado, contados con `rtk proxy` para que no pase en vacío; §2 dejando la derivada sin cambios; §4 sin coincidencias—, y que §6 no se ejecuta porque necesita modelo y queda fuera del run (FR-062), con la comprobación de que sus órdenes nombran lo que existe: la eval 19 nueva, las banderas de `TestPrepararSesion` y de `TestComprobarConsultaRepetida` (`go test -tags evals -list`) y `kitlegal skills install` con `--host` (`go run ./cmd/kitlegal skills install --help`); (4) que no hay ADR nuevo, que `git diff --quiet main -- specs/010-h7-internal-graph-grafo specs/011-h7-1-graph-check-acotado` sale con 0 (FR-070), que el binario no cambia —`git diff --name-only main -- cmd internal ':!internal/evals' ':!*_test.go'` no imprime nada y `go list -deps ./cmd/kitlegal` no nombra `internal/evals`—, que el esquema común de eval no cambia (`git diff --quiet main -- '*eval.yaml.json'`), que no hay fuente ni grabación nuevas para docs/SOURCES.md, y que la skill tiene menos de 300 líneas; si una cifra de cobertura queda bajo su umbral, se añaden los tests que faltan en los ficheros de test declarados, sin tocar código de producto; verificación: `make ci` en verde — FR-056, FR-062, FR-070, FR-080; SC-007, SC-008; rutas: specs/012-h7-2-la-consulta-repetida/cierre.md, internal/evals/prohibidas_test.go, internal/evals/consulta_repetida_test.go, internal/evals/informe_test.go, internal/evals/juzgar_test.go

---

## Dependencias y orden de ejecución

El orden es el de plan.md, «Orden de implementación» (de dentro afuera), no el de prioridad de las historias, con una
precisión: la lista del repositorio (plan.md paso 4) va antes que `Juzgar` y el informe (paso 3), porque
`TestJuzgarLasExpresionesProhibidas` juzga «con la lista del repositorio» (plan.md, «Tests nuevos») y la lista no
depende de `Juzgar`. La lista y su juicio van antes que la eval nueva y que `SKILL.md`, porque las dos se juzgan con ella
y las evals van antes que la skill (Definition of Done §1.10).

| Tarea | Depende de | Por qué |
|---|---|---|
| T001 | — | el esquema es el material que lee T002; su test lo compila desde su ruta |
| T002 | T001 | `LeerConjunto` valida la lista contra el esquema de T001 |
| T003 | T002 | `LeerConjunto` ya reconoce la lista (antes, sería un fichero de eval mal nombrado) y los subtests usan `ExtraerExpresionesProhibidas` |
| T004 | T003 | `TestJuzgarLasExpresionesProhibidas` juzga con la lista del repositorio |
| T005 | T004 | el informe publica el campo de `ResultadoDeEval` y cuenta las sesiones con alguna expresión |
| T006 | — (orden del bucle) | la derivada y su control no dependen de la lista; van antes de la eval que la usa |
| T007 | T003, T006 | la eval 19 prepara su grafo previo con la derivada de T006; `expresiones-en-los-bloques` (T003) cuenta ya la redacción original |
| T008 | T006, T007 | ninguna eval usa ya la carpeta retirada; la comparación carpeta a carpeta usa la lista de derivadas del grafo previo de T006 |
| T009 | T003, T004, T007 | las evals y su juicio van antes que `SKILL.md` (DoD §1.10); `expresiones-de-la-skill` vigila el bloque de la regla 7 |
| T010 | T002, T003 | la comprobación usa `ExtraerExpresionesProhibidas` y la lista del repositorio |
| T011 | T005, T009, T010 | la documentación describe el comportamiento final: el informe, la skill y lo retirado |
| T012 | todas | enumera lo tocado, mide la cobertura final y ejecuta el quickstart salvo §6 |

Tras la última tarea, el workflow ejecuta `make ci`, hace la revisión final y el cierre en la plataforma, con el job de
evals (SC-001, FR-087); nada de eso es una tarea. No hay suite congelada que activar («Aceptación e2e: no aplica»).

### Oportunidades de paralelismo

En el bucle del workflow todo va en secuencia, y ninguna tarea lleva `[P]`: casi todas comparten fichero con la anterior
o la siguiente (`internal/evals/conjunto_test.go`, `internal/evals/juzgar_test.go`, `internal/app/grafo_test.go`).
Independencias que el orden no aprovecha:

- T006 (la derivada y su control, en `internal/app`) no depende de nada de la fase 1.
- T010 (la comprobación del quickstart) solo depende de T002 y T003; no comparte fichero con T004-T009.

### Ejemplo de ejecución en paralelo (fuera del workflow)

```text
# Tras T003, dos agentes podrían avanzar a la vez:
T004, T005            juicio e informe               (internal/evals/juzgar.go, internal/evals/informe.go)
T006                  derivada y su control          (internal/app/grafo_test.go y la derivada)
# Y, en cualquier momento tras T003:
T010                  comprobación del quickstart    (internal/evals/consulta_repetida.go, internal/evals/job_test.go)
```

## Criterios de prueba independiente por historia

| Historia | Prueba independiente | Tareas |
|---|---|---|
| US1 (P1) quien pregunta lee la norma | `make skills-check`, `avisos-de-la-skill`, `hallazgos-de-la-skill`, `expresiones-de-la-skill`; la lista en todas las respuestas de las evals que activan la skill, en el job del cierre (SC-001) | T009 (y la lista y su juicio, T003-T005) |
| US2 (P1) la eval de la consulta repetida | `TestGrabacionesDerivadas`, `TestGrabacionesDerivadasInventadas`, `grafo-previo`, `TestEstadoPrevioDeLaRedaccionCambiada`, quickstart.md §2, §4, §5; la eval 19 en el job del cierre | T006, T007, T008 |
| US3 (P1) una expresión prohibida hace que la sesión no pase | el test del esquema, `TestLeerConjunto`, `TestExtraerExpresionesProhibidas`, `TestJuzgarLasExpresionesProhibidas`, los tests del informe, `expresiones-calibradas`, `expresiones-en-los-bloques`, `expresiones-de-la-skill` | T001-T005 |
| US4 (P2) nada de lo dicho en otra conversación | la familia `otra_conversacion` y `expresiones-calibradas` (la de «te habría confirmado antes», SC-003); la viñeta C5 de `SKILL.md` | T003, T009 |
| US5 (P2) cualquiera repite la aceptación | `TestCondicionesDeLaConsultaRepetida`; el escenario §6 de quickstart.md, fuera del run (SC-002) | T010 |

## Estrategia de implementación

- **MVP** (lo que hace visible y decide el defecto medido): T001-T005, la lista en el formato común, calibrada contra
  H7.1, aplicada por `Juzgar` y publicada en el informe (US3).
- **Incremento 2**: T006-T008, la eval de la consulta repetida con evidencia coherente y el control de derivaciones que
  lo exige (US2).
- **Incremento 3**: T009, `boe-legislacion` v0.1.2 (US1, US4), medida por la lista en el job del cierre.
- **Incremento 4**: T010, la comprobación mecánica del escenario del quickstart (US5).
- **Cierre**: T011-T012, documentación y Definition of Done.

## Trazabilidad: requisitos → tareas

| Requisito | Tareas |
|---|---|
| FR-001 a FR-004 | T007 (FR-004: la forma y la tasa las declara el informe desde `hallazgos`, mecanismo de H7.1) |
| FR-010 a FR-014 | T006 (FR-013 también T008; FR-014: sin grabación nueva) |
| FR-020, FR-021 | T007, T008 |
| FR-030 | T007 |
| FR-040 a FR-047 | T009 (FR-043 también T003, `expresiones-de-la-skill`) |
| FR-048 | T011 |
| FR-050 | T001, T002, T003, T011 |
| FR-051 | T002, T003, T004 |
| FR-052 | T004 |
| FR-053 | T005, T011 |
| FR-054 | T004, T005 |
| FR-055 | T001, T002, T005 |
| FR-056 | T002, T010, T012 (y la batería de verificación) |
| FR-060 a FR-062 | T010 (el escenario, en quickstart.md; T012 comprueba que sus órdenes nombran lo que existe) |
| FR-070 | T012 |
| FR-080 | todas; T012 |
| FR-081 | T006, T008 |
| FR-082 | T001, T002 |
| FR-083 | T004 |
| FR-084 | T003 |
| FR-085 | T003, T007 |
| FR-086 | T011 |
| FR-087 | el cierre del workflow (no es una tarea) |
| SC-001 | T004, T005, T009 y el job de evals del cierre del workflow |
| SC-002 | T010 (su comprobación); el escenario §6, fuera del run (FR-062) |
| SC-003 | T003 |
| SC-004 | T003, T007 |
| SC-005 | T006, T008 |
| SC-006 | T001, T002, T004 |
| SC-007 | T007, T008, T012 |
| SC-008 | T009, T011, T012 |

## Definition of Done (docs/ROADMAP.md §1) → tareas

| Punto | Aplica | Tarea |
|---|---|---|
| 1. `make ci` en verde | sí | todas; T012 lo comprueba al final |
| 2. Tests offline | sí; ninguna grabación nueva, una derivada producida por código | cada tarea; T006, T008 |
| 3. Reglas de dependencia | sí, sin cambios: `internal/evals` lee el grafo con `graph.Leer` y ejecuta `graph check` en proceso (R3) | T007 |
| 4. Salida contra `schemas/` y `schema-check` | el binario no cambia; el esquema nuevo es el de la lista, que no sale de `--describe` | T001 |
| 5. Errores tipados, sin `panic` | sí: una lista mal formada es un `FicheroMalFormado` | T002, T005 |
| 6. e2e y `CHANGELOG.md` | sin e2e: el binario no cambia («Aceptación e2e: no aplica»); `CHANGELOG.md` sí | T011 |
| 7. ADR | no: ninguna decisión de arquitectura cambia (FR-070) | T012 lo deja escrito |
| 8. `docs/SOURCES.md` | no: ninguna fuente nueva (FR-014) | T012 lo deja escrito |
| 9. Cobertura | sí | T002-T005, T010; T012 la mide |
| 10. Skills: evals antes de `SKILL.md`, < 300 líneas, tabla sin drift | sí | T003, T007 (evals), T004 (su juicio), T009 (`SKILL.md`) |
| 11. Dimensión territorial | no: ni la lista, ni la eval (norma estatal), ni la skill se particularizan | — |
| 12. Grafo | sin cambios en el grafo; el texto citado sigue saliendo de `kitlegal boe` (FR-046) | T009 |
| 13. Recursos, plazos o escritos (desde H9) | no aplica | — |
