# Tasks: H5.1 · Avisos de vigencia en las evals

**Input**: `specs/007-h5-1-avisos-de-vigencia/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` aplicada; H0-H5 cerrados (binario
multicall con el applet `boe` y sus avisos `consolidacion-no-finalizada`, `derogada` y `vigencia-agotada`; `Makefile` con
`ci`, `skills-check`, `skills-sync` y `schema-check`; `internal/evals` con `LeerEval`, `Juzgar`, `EscribirInforme`,
`TestEvalsDelRepositorio`, `TestIdentificadoresDeLasNormas` y el arnés `TestGrabarEvals`; `internal/skills` con
`CompilarEsquema`; las 17 evals de `evals/boe-legislacion/`; las grabaciones de H4 y de H5; ADR 0016). **Ninguna tarea de
H5.1 crea el esqueleto ni los gates: `make ci` existe y pasa desde H0**, y cada tarea lo deja en verde al terminar.

**Tests**: obligatorios (constitución §III; spec, FR-035, FR-043, controles del hito). Cada tarea escribe su test antes
del código que lo hace pasar (rojo → verde); los nombres de test y de subtest son los del «Inventario de tests» de
plan.md y los de los contratos.

**Modo**: desatendido. Las tareas se ejecutan y verifican **una a una** por el workflow `hito`, con guardián de diff por
tarea. Cada tarea es una **rebanada vertical** que deja `make ci` en verde y **declara en su propia línea todas las
rutas** que va a crear o modificar, y **solo** esas (plan, «Obligaciones», punto 1).

## Formato: `[ID] [P?] [etiquetas] [Story] Descripción con rutas`

- **[P]**: la tarea no comparte fichero con sus vecinas ni depende de ellas (en el bucle del workflow van igual en
  secuencia).
- **[Story]**: historia del spec — US1 quien consulta una norma derogada ve el aviso con su forma fija · US2 quien
  mantiene las evals exige avisos y el juicio los compara por su forma fija · US3 el informe publica qué avisos llevó cada
  sesión · US4 la forma de cada aviso tiene una sola fuente de verdad vigilada en `make ci` · US5 una eval sobre una norma
  derogada, medida sin red y sin decidir el veredicto. Las tareas de documentación, cierre y plataforma no llevan
  historia.
- **[datos]**: la tarea solo toca material bajo `testdata/` o `schemas/`. Son **dos**, las del plan (obligación 2): T003
  (el esquema de eval, fichero existente modificado → pausa) y T008 (el manifiesto de grabación, fichero existente
  modificado → pausa, en la que una persona graba el índice). Ninguna mezcla otro trabajo ni toca código.
- **[plataforma]**: la tarea necesita la plataforma remota. Es **una**, la última: T014 (publicar la rama, abrir la
  propuesta de cambio y leer y registrar la ejecución de aceptación). Fusionar nunca es del workflow.

## Batería de verificación por tarea

- **`make ci`** en cada tarea, **en primer plano** y con la tarea marcada `[X]` en el mismo turno en que termina en
  verde: formato · lint (`depguard`, `forbidigo`, `gosec`, `misspell`, `dupl`, `paralleltest`…, con `run.build-tags` que
  incluye `grabacion`) · tests con `-race` · `test-integration` · vulnerabilidades · `schema-check` · `skills-check`
  (con `TestEvalsDelRepositorio`, sus subtests nuevos incluidos) · secretos · integridad y `tidy -diff` de módulos.
  Ningún registro de la verificación se escribe en la raíz del repositorio.
- **Cero red en todo el bucle**: todo test es offline (entradas sintéticas en constantes o en `t.TempDir()`, esquemas
  sintéticos compilados, `httpx.Replay` sobre lo grabado). **`KITLEGAL_RECORD`, el guion de grabación, `make evals`,
  `make verify-sources` y `gh` no se ejecutan en ninguna tarea salvo T014** (que solo usa `gh` y `git push`); la grabación
  la hace una persona en la pausa de T008 (plan, obligación 7).
- **Sin cambios del kernel, del dominio ni de la salida del applet** (spec, *Fuera de alcance*; FR-012): ninguna tarea
  declara ficheros de `internal/app`, `internal/cli`, `internal/httpx`, `internal/cache`, `cmd/` ni del paquete del
  dominio; en la fuente `boe` solo `avisos.go` y su test (T001). Ningún guion `testscript` ni golden cambia.
- **Sin etiquetas copiadas** (FR-011, plan obligación 6): ningún fichero de código de `internal/evals` (lo que no termina
  en `_test.go`) escribe `NORMA DEROGADA`, `VIGENCIA AGOTADA` ni `TEXTO POSIBLEMENTE DESACTUALIZADO`, tampoco en un
  comentario; los comentarios hablan de «la marca, la etiqueta y los dos puntos». `TestEtiquetasSoloDesdeBoe` (T002) lo
  vigila desde que existe.
- **Texto literal** (plan, obligación 5): el esquema, la entrada del manifiesto, la norma, la eval y `SKILL.md` reciben
  exactamente los textos de sus contratos; lo que copia de lo grabado (título, rango) se copia solo de la búsqueda
  grabada, y si no coincide la tarea se detiene sin marcarse y lo anota en `gates/tarea-Tnnn.md`.
- **Sin atajos**: ni `//nolint`, ni linter desactivado, ni exclusión nueva, ni test saltado, ni error silenciado, ni
  umbral rebajado, ni `panic` en rutas de usuario (`regexp.MustCompile` solo con piezas fijas y palabras pasadas por
  `regexp.QuoteMeta`, research D2). Si un test queda legítimamente en rojo porque su implementación pertenece a otra
  tarea, la tarea está mal delimitada: se anota en `gates/tarea-Tnnn.md`, se redelimita y se deja sin marcar.
- **`rtk`**: toda orden cuya salida se filtre o compare (`go test -v | grep`, `git status --porcelain`, `git diff`,
  `make`) se ejecuta con `rtk proxy`, cada etapa de la tubería incluida (quickstart, tabla de formas).

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —test antes que código, `make ci` en verde por
sí solas— T001, T002, T004, T005, T006 y T011. Las demás, cada una por su razón, tampoco dejan nada a medias: **T003 y
T008** son `[datos]` y solo pueden tocar `schemas/` o `testdata/`; el esquema de T003 lo lee `LeerEval` sin romper nada
porque ninguna eval lleva todavía `avisos` y `Decode` ignora la clave (research V5), y lo cubren los casos de T004; la
entrada de T008 la leen `TestManifiestoDeGrabaciones` y `TestIdentificadoresDeLasNormas`, que siguen en verde con la
premisa corregida en T007 (research V14); **T007** cambia la premisa de un test y un arnés con etiqueta `grabacion` que
ninguna tarea ejecuta (toca la red), comprobado porque el test sigue en verde y el arnés compila y pasa el lint con esa
etiqueta dentro de `make ci`; **T009 y T010** son datos del repositorio (la norma con su referencia regenerada y la eval
18) que leen tests ya existentes (`TestNormasDelRepositorio`, `TestSkillsDelRepositorio`,
`TestIdentificadoresDeLasNormas`, `TestEvalsDelRepositorio`), sin código nuevo; **T012** es documentación; **T013** es
validación sobre el árbol terminado, cuyo único fichero escrito es el cuerpo de la PR en el directorio del hito;
**T014** solo publica, lee y registra.

**Test de extremo a extremo de la entrega.** H5.1 no cambia el comportamiento visible del binario, así que no añade ni
cambia ningún guion `testscript` (spec, *Fuera de alcance*; plan, *Complexity Tracking*): los existentes siguen en verde
sin cambios como prueba de FR-012 en cada `make ci`. Lo que describe la entrega de extremo a extremo es la **eval
`18-lrjpac-norma-derogada.yaml`**, que entra en **T010**, la primera tarea que puede dejarla en verde (necesita el campo
`avisos` de T003-T004, la grabación de la pausa de T008 y la norma de T009): `make ci` ejecuta sin red cada una de sus
consultas (`TestEvalsDelRepositorio/grabado`), y el job de evals la ejecuta con la skill instalada en T014.

---

## Phase 1: Etiquetas y forma fija (fundación; US4, US2)

**Objetivo**: que la etiqueta de cada aviso tenga una sola fuente de verdad en el binario y que exista, sobre entradas
sintéticas, la función que reconoce la forma fija y las dos comprobaciones mecánicas, antes de tocar el formato.

- [ ] T001 [US4] Etiquetas de aviso exportadas por el binario, sin cambiar un byte de su salida: `internal/source/boe/avisos.go` con las tres constantes no exportadas de etiqueta (`TEXTO POSIBLEMENTE DESACTUALIZADO`, `NORMA DEROGADA`, `VIGENCIA AGOTADA`, en el orden de las condiciones), las tres frases compuestas con ellas (`"⚠ " + etiqueta + ": …"`, idénticas byte a byte a las actuales) y `EtiquetasDeAviso() map[string]string`, un mapa nuevo en cada llamada con la etiqueta de cada código de `CodigosDeAviso()` y el comentario del contrato de la forma fija §2; `internal/source/boe/avisos_test.go` con `TestEtiquetasDeAviso` escrito primero (`exactamente-tres`: igual al literal de los tres códigos con sus etiquetas y claves iguales a `CodigosDeAviso()`; `frases-con-su-forma`: con las tres condiciones, cada `Aviso.Texto` de `avisosDe` empieza por `"⚠ " + etiqueta + ":"` de su código; `cada-llamada-su-mapa`); `TestAvisosDe`, `TestCodigosDeAviso`, los golden de la fuente, los esquemas publicados de norma y bloque y los guiones e2e sin cambios y en verde; nada más cambia en el paquete (FR-010, FR-011, FR-012, US4 escenarios 1 y 4, SC-005, research D1 y V9).

- [ ] T002 [US2] Forma fija y comprobaciones mecánicas sobre entradas sintéticas: `internal/evals/avisos.go` (nuevo) con `ExtraerAvisos(texto string) []string` —una expresión por código compilada una sola vez con `sync.OnceValue` desde `boe.EtiquetasDeAviso()` con la gramática y las piezas exactas del contrato de la forma fija §1 (marca U+26A0 con selector U+FE0E o U+FE0F opcional, separador de blancos horizontales de categoría Zs, tabulador, `*` y `_`, al menos un blanco entre palabras, cada palabra de `strings.Fields` de la etiqueta con `regexp.QuoteMeta` sin distinguir mayúsculas, y `:` final, todo en una línea); devuelve los códigos en el orden de `CodigosDeAviso()`, sin repetir, o nil—, `ComprobarFormasDeAviso(texto string) error` y `ComprobarCodigosDeAviso(esquema *jsonschema.Schema) error` (códigos del `Enum` al que llegan la propiedad `avisos`, sus `items` y cada `$ref`), las dos con `errors.Join` de un error por código y los mensajes y el orden exactos de data-model §6, sin escribir ninguna etiqueta en el fichero; `internal/evals/avisos_test.go` escrito primero con `TestExtraerAvisos` (cada subtest del contrato de la forma fija §4, con su texto literal y sus códigos esperados), `TestComprobarFormasDeAviso` (`completo`, `falta-la-etiqueta`, `falta-la-marca`, `faltan-los-dos-puntos`, `ninguna`, con el error exacto `falta la forma fija del aviso derogada: ⚠ NORMA DEROGADA:`), `TestComprobarCodigosDeAviso` (`exacto` con el enumerado detrás de un `$ref`, `falta-un-codigo`, `sobra-un-codigo`, `sin-enumerado`, `sin-avisos`, sobre esquemas sintéticos en constantes compilados con `skills.CompilarEsquema`, mensajes exactos del contrato de formato, juicio e informe §3) y `TestEtiquetasSoloDesdeBoe` (ningún `*.go` del paquete que no termine en `_test.go` contiene una etiqueta de `boe.EtiquetasDeAviso()`, y al menos uno leído); nada del paquete usa todavía las funciones nuevas (FR-011, FR-013, FR-014, FR-032, FR-033, US2 escenarios 3 a 6, US4 escenarios 2, 3 y 5, SC-003, SC-005, research D2, D3, D5, D7 y D8).

**Checkpoint**: el binario exporta sus tres etiquetas con la salida intacta; la forma fija se reconoce con su tolerancia
exacta y las dos comprobaciones fallan nombrando el código, todo probado sin tocar el formato, el juicio ni la skill.

---

## Phase 2: `avisos` en el formato común de eval (US2)

**Objetivo**: que una eval que activa la skill pueda declarar los avisos esperados, validados contra los códigos del
binario, y que `make ci` vigile que el esquema publicado admite exactamente esos códigos.

- [ ] T003 [datos] [US2] Esquema del formato común de eval: `schemas/eval.yaml.json` con exactamente los tres cambios del contrato de formato, juicio e informe §1 y ninguno más —propiedad `avisos` detrás de `citas`, `type: array`, `minItems: 1` e `items` con `$ref` a la definición `codigo-de-aviso`; `{ "required": ["avisos"] }` como tercera alternativa del `anyOf` de la rama `else`; y la definición `codigo-de-aviso` como primera entrada de `$defs`, con el enumerado `consolidacion-no-finalizada`, `derogada`, `vigencia-agotada` en una sola línea y en el orden de `CodigosDeAviso()`—; ningún otro fichero; `make ci` sigue en verde porque ninguna eval lleva `avisos` (research V5). Provoca la pausa humana de revisión (esquema existente modificado) (FR-020, FR-021, FR-022, FR-055, SC-002, SC-009, research D4).

- [ ] T004 [US2] Campo `Avisos` de `Eval` y comprobación del esquema publicado: `internal/evals/formato.go` con `Avisos []string` y etiqueta `yaml:"avisos"` detrás de `Citas`, con el comentario literal del contrato de formato, juicio e informe §2 (`LeerEval` no cambia: valida el esquema); `internal/evals/formato_test.go` con los casos nuevos de `TestLeerEval`, escritos primero sobre documentos en constantes (`avisos` con `Avisos` en el orden del fichero, `aviso-desconocido`, `no-activa-con-avisos`, `avisos-vacio` y `citas-vacio` con el mismo `minItems: got 0, want 1` en su ruta, `aviso-repetido` y `cita-repetida` leídas con el repetido en su lista, `informativa-con-avisos`), cada error con el texto exacto del contrato §6 y la línea que fija su documento; `internal/evals/conjunto_test.go` con el subtest `avisos-del-esquema` de `TestEvalsDelRepositorio` (`ComprobarCodigosDeAviso(esquemaDeEval())` sobre el esquema publicado → nil); las 17 evals existentes siguen válidas sin cambios (FR-013, FR-020, FR-021, FR-022, FR-023, US2 escenarios 1, 2, 9 y 10, US4 escenario 2, SC-002, SC-005, research D4, D5 y D6).

**Checkpoint**: el formato admite `avisos` solo con `activa: true` y solo con los códigos del binario; el esquema y
`CodigosDeAviso()` no pueden divergir sin que `make skills-check` falle nombrando el código.

---

## Phase 3: Juicio de los avisos (US2)

**Objetivo**: que `Juzgar` reparta los avisos esperados por su forma fija y que un aviso ausente impida pasar.

- [ ] T005 [US2] Reparto de avisos en `Juzgar`: `internal/evals/juzgar.go` con la constante `motivoDeAvisoAusente = "aviso ausente: "` junto a las de los motivos, `AvisosEncontrados` y `AvisosAusentes` (`json:"avisos_encontrados"` y `json:"avisos_ausentes"`) detrás de `CitasAusentes` en `ResultadoDeEval` con el comentario del contrato de formato, juicio e informe §4, `repartirAvisos(eval.Avisos, ExtraerAvisos(sesion.Respuesta))` justo después de `repartirCitas` y con su misma forma, `Pasa` que exige además `len(AvisosAusentes) == 0`, el orden de motivos del contrato §4 (cada aviso ausente detrás de cada cita ausente y antes del del modelo) y los comentarios de `Juzgar`, `Motivos` y `Pasa` nombrando los avisos sin escribir ninguna etiqueta; `internal/evals/juzgar_test.go` con los casos nuevos de `TestJuzgar`, escritos primero sobre la eval 01 con `avisos` y la sesión que pasa cambiando solo la respuesta (`aviso-con-su-forma-fija`, `aviso-con-variantes-toleradas` con un juicio por variante, `aviso-ausente`, `aviso-con-otra-redaccion`, `aviso-negado` con el comando y la cita presentes, `forma-fija-y-lo-contrario`, `aviso-no-esperado`, `avisos-en-el-orden-de-la-eval`, `aviso-repetido`, `cita-y-aviso-ausentes`) y los casos existentes, sin `avisos`, con las dos listas nulas y el mismo resultado; y `.golangci.yml`, solo para añadir a `misspell.ignore-rules` la palabra `variantes` con el comentario de su motivo (el nombre del subtest `aviso-con-variantes-toleradas`, fijado por el inventario de tests de plan.md y quickstart.md §5, que `misspell` v0.8.0 lee como «variants»), sin ningún otro cambio en ese fichero; `TestInforme` sigue en verde porque las listas nulas se escriben como `[]` (research V8) (FR-030, FR-031, FR-032, FR-033, FR-034, FR-035, US2 escenarios 3 a 8, SC-003, research D3).

**Checkpoint**: una eval con `avisos` solo pasa si la respuesta lleva la forma fija de cada uno; la otra redacción y la
negación dejan el aviso ausente; las evals sin `avisos` se juzgan igual que en H5.

---

## Phase 4: El informe publica los avisos (US3)

**Objetivo**: que el informe publique, por sesión, los avisos encontrados y ausentes y su motivo, junto a las citas.

- [ ] T006 [US3] Avisos en el informe: `internal/evals/informe.go` con `encabezadosDeSesiones` ganando «Avisos encontrados» y «Avisos ausentes» detrás de «Citas ausentes» y `filasDeSesiones` con `unidosOVacio(resultado.AvisosEncontrados, ningunoEnElInforme)` y `unidosOVacio(resultado.AvisosAusentes, ningunoEnElInforme)` detrás de la celda de citas ausentes, con su comentario, y nada más (contrato de formato, juicio e informe §5); `internal/evals/informe_test.go` con `TestInformeConAvisos` escrito primero —copia con `os.CopyFS` a un `t.TempDir()` el caso `aprobado` de las ejecuciones sintéticas del informe, añade en la copia `avisos` `derogada` y `vigencia-agotada` a la eval 01 y antepone `⚠ NORMA DEROGADA: esta norma ha sido derogada. ` a la respuesta de la sesión 01 exigiendo exactamente dos sustituciones, y llama a `EscribirInforme` sobre la copia; subtests `uno-encontrado-y-otro-ausente` y `aviso-detras-de-la-cita` con todo lo que fija el contrato §6 (claves en `informe.json` en crudo, motivos de la sesión y de la raíz, la sesión 11 con `[]` y `[]`, cabecera de la tabla de sesiones con las once celdas, filas de la 01 y de la 11, respuesta publicada)— y `TestInforme/aprobado` ampliado con `"avisos_encontrados": []` y `"avisos_ausentes": []` en cada sesión (`informeCrudo` con los dos campos como `jsontext.Value`); ningún material nuevo ni modificado en ninguna carpeta de datos de prueba (FR-040, FR-041, FR-042, FR-043, US3 escenarios 1, 2 y 3, SC-004, research D9 y D10, V19 y V20).

**Checkpoint**: un rojo por aviso ausente se distingue en el informe de uno por cita ausente, y la respuesta de cada
sesión sigue publicada.

---

## Phase 5: La eval de la norma derogada, su norma y su grabación (US5)

**Objetivo**: que la Ley 30/1992 tenga su índice grabado por una persona, su entrada en `data/normas.yaml` y una eval
informativa que exige sus dos avisos, confirmada antes de cualquier cambio de `SKILL.md`.

- [ ] T007 [P] [US5] Arnés con la unión de grabaciones y premisa de la verificación de identificadores, antes del manifiesto: `internal/evals/grabacion_test.go` (etiqueta `grabacion`), donde el método `pedir` de `grabacionDeEvals` siembra la consulta desde `UnionDeGrabaciones()` en lugar de solo desde las grabaciones de H4, su mensaje de error dice «desde las grabaciones de H4 y de H5» y el comentario de `TestGrabarEvals`, punto 1, dice que solo se pide a la fuente lo que no está grabado en ninguno de los dos conjuntos, sin ningún otro cambio en el arnés ni en el guion de grabación; e `internal/evals/grabaciones_test.go`, donde `otraNormaDeLaBusqueda` recibe el manifiesto y la posición de la entrada y elige el primer resultado de su búsqueda grabada que no está en la tabla de normas y cuyo título no empieza por el `titulo_empieza_por` de ninguna entrada del manifiesto, con su comentario, y `TestIdentificadoresDeLasNormas` la llama con esos argumentos (contrato de la eval y la grabación §1); verde porque la premisa sigue eligiendo `BOE-A-1992-26318` mientras el manifiesto no tenga la entrada nueva (research V14) y el arnés, que ninguna tarea ejecuta, compila y pasa el lint con su etiqueta; son cambios planificados de la premisa y del arnés, no correcciones de un rojo (FR-054, FR-055, FR-056, research D11 y D12).

- [ ] T008 [datos] [US5] Entrada de la Ley 30/1992 en el manifiesto de grabación: `testdata/evals/grabaciones.json` gana, al final de `normas` y en una línea con la forma de las demás, exactamente la entrada del contrato de la eval y la grabación §2 (búsqueda `procedimiento administrativo común`, `titulo_empieza_por` `Ley 30/1992,`, bloques `a42` y su `para` literal); ningún otro fichero; el ejecutor no graba, no usa la red ni escribe ninguna grabación. Provoca la pausa humana (fichero existente modificado), en la que una persona sigue el procedimiento del contrato §4 —ejecuta `grabar-evals.sh`, restaura la grabación reescrita de `robots.txt`, comprueba que el único fichero nuevo es el índice de `BOE-A-1992-26318` y que lleva el bloque `a42`, ejecuta los cuatro tests del procedimiento y confirma solo ese índice en su propio commit— y aprueba; si la búsqueda no resuelve la norma o el índice no lleva `a42`, restaura, rechaza y lo anota (FR-054, FR-055, FR-057, US5 escenario 3, SC-006, SC-009, research D11, V11 a V18 y S5).

- [ ] T009 [US5] Norma y referencia: `data/normas.yaml` gana, al final de `normas`, exactamente la entrada del contrato de la eval y la grabación §5 (`BOE-A-1992-26318` con el título y el rango copiados de la búsqueda grabada `procedimiento administrativo común`, `abreviatura: LRJPAC` y sus dos materias, sin ninguna marca de derogación), y `skills/boe-legislacion/references/normas.md` se regenera solo con `make skills-sync`, nunca a mano; verde sin cambiar ningún test: `TestNormasDelRepositorio` (la tabla contra su esquema), `TestSkillsDelRepositorio` (referencia sin deriva) y `TestIdentificadoresDeLasNormas` (la entrada nueva de T008 resuelve la norma con su título y los cuatro subtests negativos siguen fallando donde deben); si el título o el rango grabados no coinciden con el contrato, se detiene sin marcarse y lo anota en `gates/tarea-T009.md` (FR-056, US5 escenario 4, SC-006, research D13, V14 a V16).

- [ ] T010 [US5] Eval de la norma derogada, antes que la skill: `evals/boe-legislacion/18-lrjpac-norma-derogada.yaml` con el contenido literal del contrato de la eval y la grabación §6 (comentario, pregunta «¿Qué dice el artículo 42 de la Ley 30/1992 sobre la obligación de resolver?», `activa: true`, `informativa: true`, comando esperado bloque `boe` `BOE-A-1992-26318` `a42`, cita esperada `BOE-A-1992-26318` `a42` y `avisos` `derogada` y `vigencia-agotada`, en ese orden); es el test de extremo a extremo de la entrega y queda en verde sin tocar código ni tests con `TestEvalsDelRepositorio` (`formato` con 18 evals, `conjunto` con exactamente 10 positivas que deciden y la 18 informativa, «materias distintas», «informativas» y «tamaño» sin defectos, `normas-conocidas`, `grabado` preparando y sirviendo con `--offline` su bloque, su índice y sus metadatos con código 0); se confirma en su propio commit, anterior al primero que cambia la skill (T011); si `grabado` falla porque falta el índice grabado en la pausa de T008, se detiene sin marcarse, no graba y lo anota en `gates/tarea-T010.md` (FR-050, FR-051, FR-052, FR-053, FR-057, US5 escenarios 1, 2, 3 y 5, SC-006, SC-009, research D14 y D16, V15 y V21).

**Checkpoint**: 18 evals en `evals/boe-legislacion/`, la 18 informativa con sus dos avisos y todo lo que necesita grabado
y servido sin red; la tabla y su referencia incluyen la Ley 30/1992 sin marcar su derogación.

---

## Phase 6: La forma fija en la skill (US1, US4)

**Objetivo**: que `SKILL.md` fije la forma con la que la respuesta traslada cada aviso y que `make ci` exija la forma
completa de cada código.

- [ ] T011 [US1] Forma fija de los avisos en la skill y su comprobación: `skills/boe-legislacion/SKILL.md` con exactamente los tres cambios literales del contrato de la forma fija §3 y ninguno más —el último punto de «### 5. Responder citando»; detrás del último punto de «## Cómo se cita», el párrafo de la forma fija, la lista con `⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:` y `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` con su código, el bloque de ejemplo `⚠ NORMA DEROGADA: esta norma ha sido derogada.` y la regla de la etiqueta entera en la misma línea; y la regla 3 de «## Reglas»—, con la región generada intacta, 199 líneas y ninguna añadida de más de 120 caracteres; `internal/evals/conjunto_test.go` con el subtest `avisos-de-la-skill` de `TestEvalsDelRepositorio` (`ComprobarFormasDeAviso` sobre el `SKILL.md` de `boe-legislacion` leído del repositorio → nil); verde con `TestSkillsDelRepositorio` (frontmatter, menos de 300 líneas, tabla de comandos sin deriva, sin nombrar evals, job, modelos ni normas); el commit de esta tarea es el primero de la rama que cambia `SKILL.md` (FR-001, FR-002, FR-003, FR-004, FR-005, FR-014, US1 escenarios 1 a 4, US4 escenario 3, SC-001, SC-005, SC-009, research D6, D7 y D15, V29 y V33).

**Checkpoint**: la skill enseña la forma fija de los tres avisos y `make skills-check` falla nombrando el código si
alguna se pierde.

---

## Phase 7: Documentación y cierre

- [ ] T012 [P] Documentación del hito: `CHANGELOG.md` con un bloque *De H5.1* en «Añadido» de *Unreleased* (la forma fija de los avisos en la skill, el campo `avisos` del formato común de eval, el reparto de avisos en el informe y la eval informativa de la norma derogada); `README.md` y `CONTRIBUTING.md` con la fila `avisos` en su tabla del formato común de eval (opcional, solo con `activa: true`, códigos de aviso del binario) y, en `CONTRIBUTING.md`, «y lleva la forma fija de cada aviso de `avisos`» en la frase que dice cuándo pasa una eval; ninguna otra cifra ni sección, ni ADR, ni tabla de fuentes, ni bitácora de uso (FR-061, SC-008, Definition of Done §1.6, research D18).

- [ ] T013 Cierre sin tocar el árbol fuera del directorio del hito: ejecuta en primer plano `make ci` y, tal cual y en orden, los prerrequisitos, los escenarios 1 a 10 y la limpieza de `specs/007-h5-1-avisos-de-vigencia/quickstart.md`, con las formas `rtk proxy` de su tabla, y se detiene sin marcarse y lo anota en `gates/tarea-T013.md` ante cualquier resultado distinto del esperado; mide la cobertura global (≥ 70 %) con `go tool cover -func` sobre el perfil que deja `make ci` y confirma que el paquete del dominio no cambia respecto de `main`; y escribe `specs/007-h5-1-avisos-de-vigencia/gates/pr-h5.1.md`, el cuerpo de la PR del hito, con la plantilla del ritual (objetivo, alcance, controles añadidos, decisiones, pendientes), «Dependencias: ninguna nueva» (constitución §V), las medidas fechadas por el commit sobre el que se tomaron (`make ci`, cobertura, 199 líneas de `SKILL.md`, 18 evals con 10 que deciden) y, en pendientes, la regla de repetición de la aceptación del contrato de la ejecución de aceptación §6; ningún otro fichero (FR-060, FR-070, SC-001 a SC-006, SC-008, SC-009, Definition of Done §1.1 y §1.9, ritual §6.3, plan obligación 9).

**Checkpoint**: el hito está implementado, validado con la guía y con su cuerpo de PR escrito; solo queda la plataforma.

---

## Phase 8: Plataforma (al final; fusionar es humano)

- [ ] T014 [plataforma] Publicación y ejecución de aceptación: ejecuta tal cual `specs/007-h5-1-avisos-de-vigencia/quickstart.md` §11 —11.1 prerrequisitos (sesión de `gh`, secreto `CLAUDE_CODE_OAUTH_TOKEN`, etiqueta `evals`, cuerpo presente); 11.2 `git push -u origin 007-h5-1-avisos-de-vigencia` y, si no existe, `gh pr create` hacia `main` con el título `feat(H5.1): Avisos de vigencia en las evals` y el cuerpo `specs/007-h5-1-avisos-de-vigencia/gates/pr-h5.1.md`; 11.3 identificar como ejecución de aceptación la primera ejecución `evals` de la rama con evento `pull_request` y esperarla; 11.4 leer `informe.json` e `informe.md` entre las marcas del registro; 11.5 el programa de `jq` (veredicto `aprobado`, `red` vacío, la tasa de la eval 18 con el modelo que decide sobre 3 sesiones y el reparto de sus dos avisos en cada una) y la comprobación de que el commit del informe es el `headSha` y de que hasta la cabeza solo cambian ficheros del directorio del hito— y §11.6 solo si un commit posterior a la ejecución cambió algo fuera de ese directorio; registra en `specs/007-h5-1-avisos-de-vigencia/gates/evals-aceptacion.md` los datos de data-model §10 (ruta de identificación, enlace, `databaseId`, `headSha`, conclusión, duración del paso «Ejecutar las evals» y las dos salidas enteras de §11.5); no modifica ningún otro fichero; si falta un prerrequisito, falta una marca del registro, el programa da `false` o la comprobación del commit falla, registra la salida, se detiene sin marcarse y lo anota en `gates/tarea-T014.md` (un arreglo del repositorio va en una tarea nueva colocada antes de esta; no se relanza la ejecución para buscar otro resultado); nunca fusiona, ni empuja a `main`, ni fuerza, ni borra ramas, etiquetas o releases (FR-070, FR-071, FR-072, FR-073, SC-007, research D17 y S1 a S4, contrato de la ejecución de aceptación).

---

## Definition of Done: qué punto cubre qué tarea

| Punto (`docs/ROADMAP.md` §1) | Aplica | Tarea |
|---|---|---|
| 1. `make ci` en verde | Sí | Cada tarea (batería); T013 lo repite sobre el árbol terminado |
| 2. Tests offline; fixtures grabados si toca red | Sí | T001, T002, T004, T005, T006, T007 (tests); T008 (la única grabación nueva, el índice de `BOE-A-1992-26318`, la hace una persona en su pausa); T010 (`grabado` sin red) |
| 3. Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | Batería (`depguard`, `forbidigo`, `TestArquitectura`); T001, T002, T005 y T006 no importan nada de eso |
| 4. Salida del applet contra sus esquemas; `schema-check` | Sí, sin cambios | T001 (salida byte a byte igual, esquemas de norma y bloque sin cambios); el esquema de eval, no de salida, en T003 |
| 5. Errores tipados y exit codes; sin `panic` | Sí, sin cambios | Ningún código de salida cambia; T002 devuelve errores que nombran el código |
| 6. Comportamiento visible: e2e y `CHANGELOG.md` | `CHANGELOG.md` sí; `testscript` no | T012 (`CHANGELOG.md`, `README.md`, `CONTRIBUTING.md`); el binario no cambia su comportamiento visible, así que ningún guion cambia (spec, *Fuera de alcance*) y los existentes siguen en verde |
| 7. ADR si cambia una decisión de arquitectura | No | El mecanismo lo decide el hito en el roadmap; ninguna decisión de arquitectura cambia (spec, *Fuera de alcance*) |
| 8. `docs/SOURCES.md` y `verify-sources.sh` si toca una fuente | No | Ninguna fuente nueva ni cambio en cómo se consulta (spec, *Fuera de alcance*) |
| 9. Cobertura `core` ≥ 85 % y global ≥ 70 % | Sí | T013 (global medida; el dominio no cambia) |
| 10. Skill: evals antes que el código, `SKILL.md` < 300, `references/` sin deriva | Sí | T010 antes que T011 (FR-052); T011 (199 líneas, región intacta); T009 (`references/normas.md` regenerado) |
| 11. Dimensión territorial | No | La Ley 30/1992 es estatal; nada territorial |
| 12. Grafo (desde H7) | No | Anterior a H7 |

## Trazabilidad: requisito → tarea

| Requisito | Tareas |
|---|---|
| FR-001, FR-002, FR-003, FR-004, FR-005 | T011 (T013 lo valida con quickstart §2) |
| FR-010 | T001 |
| FR-011 | T001, T002 (`TestEtiquetasSoloDesdeBoe`), T005 |
| FR-012 | T001 (T013, quickstart §3) |
| FR-013 | T002 (`ComprobarCodigosDeAviso`), T004 (`avisos-del-esquema`) |
| FR-014 | T002 (`ComprobarFormasDeAviso`), T011 (`avisos-de-la-skill`) |
| FR-020, FR-021, FR-022 | T003, T004 |
| FR-023 | T004 |
| FR-030, FR-031, FR-034, FR-035 | T005 |
| FR-032, FR-033 | T002, T005 |
| FR-040, FR-041, FR-042, FR-043 | T006 (el motivo lo genera T005) |
| FR-050, FR-051, FR-052, FR-053 | T010 (FR-052: T010 antes que T011) |
| FR-054 | T007, T008 |
| FR-055 | T003, T007, T008 |
| FR-056 | T007 (premisa), T009 |
| FR-057 | T008, T010 |
| FR-060 | Cada tarea; T013 |
| FR-061 | T012 |
| FR-070 | T013 (cuerpo), T014 |
| FR-071, FR-072, FR-073 | T014 |
| SC-001 | T011, T013 |
| SC-002 | T003, T004 |
| SC-003 | T002, T005 |
| SC-004 | T006 |
| SC-005 | T001, T002, T004, T011 |
| SC-006 | T008, T009, T010, T013 |
| SC-007 | T014 |
| SC-008 | T012, T013 |
| SC-009 | T003, T008, T010, T011, T013 (quickstart §8) |

## Obligaciones que el plan trasladó a este fichero

| Obligación del plan | Dónde se cumple |
|---|---|
| 1. Orden y rebanadas | Una tarea por paso de «Orden de implementación», en su orden: paso 1 → T001, 2 → T002, 3 → T003, 4 → T004, 5 → T005, 6 → T006, 7 → T007, 8 → T008, 9 → T009, 10 → T010, 11 → T011, 12 → T012, 13 → T013, 14 → T014; cada una con su test y las rutas de su fila |
| 2. Tareas `[datos]` | Solo T003 (el esquema de eval) y T008 (el manifiesto); ninguna otra línea contiene esas dos carpetas como ruta, ni el texto de las etiquetas de datos o de plataforma fuera de las tareas que las llevan |
| 3. La pausa del paso 8 | T008 escribe solo la entrada; la persona sigue el contrato de la eval y la grabación §4; T009 parte de ese commit |
| 4. Eval antes que la skill | T010 y T011, tareas y commits distintos, en ese orden |
| 5. Texto literal | T003, T008, T009, T010 y T011, con el apartado de su contrato |
| 6. Sin etiquetas copiadas | Batería; T002 (`TestEtiquetasSoloDesdeBoe`), T005 |
| 7. Sin red ni grabación en tareas | Batería; T008 (el ejecutor no graba), T010 (se detiene si falta el índice) |
| 8. Definition of Done | T012 (§1.6); T003 y T008 (esquema y grabación); tabla «Definition of Done» |
| 9. Cierre | T013 |
| 10. Plataforma | T014, la última, sola; nunca fusiona |

## Dependencias y orden

El orden es estrictamente secuencial: **ninguna tarea depende de una posterior**.

- **T001 → T002**: `ExtraerAvisos` y las comprobaciones construyen sus expresiones y sus mensajes desde
  `boe.EtiquetasDeAviso()`.
- **T002 → T003 → T004**: el esquema (`[datos]`) antes que el campo y sus casos de formato (research V5); el subtest
  `avisos-del-esquema` (T004) usa `ComprobarCodigosDeAviso` (T002) y el esquema cambiado (T003).
- **T004 → T005 → T006**: `Juzgar` reparte `Eval.Avisos` (T004) con `ExtraerAvisos` (T002); el informe publica los
  campos y los motivos de `ResultadoDeEval` (T005).
- **T007 → T008 → T009 → T010**: la premisa corregida (T007) antes que la entrada del manifiesto (T008), que sin ella
  rompería `TestIdentificadoresDeLasNormas` (research V14), y el arnés con la unión antes de la pausa en que la persona lo
  ejecuta; la entrada del manifiesto antes que la norma (T009), que `TestIdentificadoresDeLasNormas` exige que resuelva
  alguna entrada; la norma antes que la eval (T010), por la regla «normas conocidas»; la grabación de la pausa de T008
  antes que la eval, por `grabado`; y la eval necesita además el campo `avisos` (T003, T004). T007 no depende de T003-T006.
- **T010 → T011**: la eval se confirma antes que el primer cambio de `SKILL.md` (FR-052, SC-009); `avisos-de-la-skill`
  usa `ComprobarFormasDeAviso` (T002).
- **T011 → T013 → T014**: T012 no depende de T011 y puede ir antes; el cierre valida el árbol terminado (quickstart §2 y
  §10 leen la skill y la documentación) y escribe el cuerpo que T014 publica.

## Oportunidades de paralelismo

En el workflow todas las tareas van en secuencia. Fuera de él podrían adelantarse **T007** (solo toca dos ficheros de
test que ninguna tarea de las fases 1 a 4 comparte) y **T012** (solo documentación). El resto comparte `formato_test.go`,
`conjunto_test.go` o depende de la tarea anterior.

```text
Tarea: "T007 arnés con la unión de grabaciones y premisa de otraNormaDeLaBusqueda"
Tarea: "T012 CHANGELOG, README y CONTRIBUTING con avisos"
```

## Estrategia de implementación

### MVP (US2 + US4 sobre US1)

1. Fases 1 a 3 (T001-T005): etiquetas, forma fija, comprobaciones, formato y juicio. **Validar**: `make ci` en verde con
   `TestExtraerAvisos`, `TestComprobarCodigosDeAviso`, `TestLeerEval` y `TestJuzgar`; ya se puede exigir un aviso en una
   eval y juzgarlo sin modelo.
2. T011 cierra el MVP con la regla escrita de la skill, pero va detrás de la eval (FR-052), así que el MVP real incluye
   la fase 5.

### Entrega incremental

- US3 (T006) hace visible el reparto en el informe.
- US5 (T007-T010) añade la grabación, la norma y la eval informativa.
- US1 (T011) fija la forma en la skill y la vigila en `make ci`.
- T012-T013 documentan y validan; T014 publica y registra la aceptación.

## Notas

- **Lo que este hito no crea, y no por olvido**: guion `testscript` nuevo, ADR, fila de la tabla de fuentes, cambios en
  el kernel, en el dominio o en la salida del applet, una eval para `consolidacion-no-finalizada`, un recuento adicional
  por aviso en el informe, la promoción de la eval 18 a decisoria, cambios en el job de evals, en `docs/USO.md` o en el
  `Makefile` (spec, *Fuera de alcance*).
- **Ficheros existentes que se tocan**, exactamente: `internal/source/boe/avisos.go` y su test (T001);
  `schemas/eval.yaml.json` (T003); `internal/evals/formato.go` y `formato_test.go` (T004); `internal/evals/conjunto_test.go`
  (T004 y T011); `internal/evals/juzgar.go`, `juzgar_test.go` y `.golangci.yml` (T005); `internal/evals/informe.go` e
  `informe_test.go` (T006); `internal/evals/grabacion_test.go` y `grabaciones_test.go` (T007);
  `testdata/evals/grabaciones.json` (T008); `data/normas.yaml` y `skills/boe-legislacion/references/normas.md` (T009);
  `skills/boe-legislacion/SKILL.md` (T011); `CHANGELOG.md`, `README.md` y `CONTRIBUTING.md` (T012). Nuevos:
  `internal/evals/avisos.go` y su test (T002), `evals/boe-legislacion/18-lrjpac-norma-derogada.yaml` (T010), el índice
  grabado (la persona, en la pausa de T008) y `gates/pr-h5.1.md` y `gates/evals-aceptacion.md` en el directorio del hito
  (T013, T014).
- **`misspell`** (research D19; `misspell` v0.8.0 comprobado al generar estas tareas con la réplica del reemplazador de
  golangci-lint): de todos los nombres de test y subtest, mensajes y encabezados que fijan los contratos, **solo**
  `variantes` salta (en `aviso-con-variantes-toleradas`, T005), y por eso T005 declara `.golangci.yml` para añadirla a
  `misspell.ignore-rules` con su motivo (plan, *Structure Decision*). En el resto de fichero Go (comentarios, variables,
  mensajes de `require`) no se escriben sueltas, sin tilde, `variantes`, `columnas` (→ «celdas» o «la cabecera»),
  `presentacion` (→ «presentación», con tilde), `limitacion` (→ «limitación»), `directorios`, `recorre`, `comparte`,
  `presentes`, `activacion` ni `materias` (salvo como clave YAML dentro de una constante); toda palabra nueva sin tilde
  se comprueba antes contra el diccionario. Una palabra marcada que fija un contrato y no esté prevista aquí detiene la
  tarea, se anota en `gates/tarea-Tnnn.md` y se redelimita declarando `.golangci.yml`.
- **`gosec`**: lecturas de ficheros del repositorio por `filepath.Clean`; escrituras en `t.TempDir()` con `0o600`.
- **Rutas protegidas desde T003 y T008**: las líneas de las tareas de código nombran el esquema de eval, el manifiesto,
  las grabaciones y las ejecuciones sintéticas del informe sin su carpeta (`esquemaDeEval()`, «el esquema publicado», «la
  búsqueda grabada», «el caso `aprobado` de las ejecuciones sintéticas del informe») o por el test que los lee, de modo
  que el guardián rechaza cualquier cambio en ellos fuera de su tarea `[datos]`.
- **Premisa de H5.1 y cambios de test**: T007 cambia la premisa de `otraNormaDeLaBusqueda` y la siembra del arnés porque
  la entrada nueva del manifiesto las invalida (research D11, D12); no son correcciones de un test en rojo, y ninguna otra
  tarea toca un test existente fuera de sus casos nuevos.

## Comprobación contra la rúbrica del juez (`juez_tasks`, criterios a-g)

| Criterio | Dónde se cumple |
|---|---|
| a. analisis_critico | `gates/analyze.md` lo escribe el paso `analyze`, que sigue a este; las obligaciones 1-10 del plan están asignadas a una tarea concreta (tabla «Obligaciones»), y la única discrepancia encontrada al generar las tareas (`variantes` y `misspell`) está resuelta en plan.md, research.md D19 y T005 |
| b. trazabilidad | Tabla «Trazabilidad» con todos los FR (001-073) y SC (001-009) del spec; cada tarea cita en su línea los requisitos que cumple; ninguna añade nada fuera del spec ni del plan |
| c. rebanadas_verdes | Cada tarea de código lleva su test escrito primero y la implementación mínima; las excepciones (`[datos]`, premisa y arnés, datos del repositorio con sus tests, documentación, cierre, plataforma) están declaradas con su razón al principio y ninguna deja un test en rojo; orden secuencial justificado en «Dependencias y orden» con las comprobaciones del plan (V5, V14, V15); el e2e de la entrega (la eval 18) va en T010, la primera tarea que puede dejarlo en verde, y los guiones `testscript` siguen sin cambios |
| d. rutas_declaradas | Cada línea nombra por su ruta completa los ficheros que crea o cambia, cada fichero de test incluido (`avisos_test.go` en T001 y T002, `formato_test.go` y `conjunto_test.go` en T004, `juzgar_test.go` en T005, `informe_test.go` en T006, `grabacion_test.go` y `grabaciones_test.go` en T007, `conjunto_test.go` en T011), sin llaves, comodines ni rutas genéricas; `.golangci.yml` en T005; lo que una tarea solo lee va sin carpeta o por su test |
| e. datos_separados | Las dos `[datos]` (T003, T008) solo tocan `schemas/eval.yaml.json` y `testdata/evals/grabaciones.json`; ninguna línea de código contiene esas carpetas; T014 es la única `[plataforma]`, va la última y solo publica, lee y registra evidencia en el directorio del hito; ninguna otra línea nombra la plataforma remota |
| f. dod | Tabla «Definition of Done» con los doce puntos: los aplicables con su tarea (`CHANGELOG.md` en T012, esquema en T003, grabación en T008, evals en T010 antes que la skill en T011, cobertura en T013) y los no aplicables con su razón (ADR, `docs/SOURCES.md`, `testscript`, territorio, grafo) |
| g. checklist_veraz | `checklists/requirements.md` no lo modifica ninguna tarea; sus marcas describen el spec, que las tareas no cambian |
