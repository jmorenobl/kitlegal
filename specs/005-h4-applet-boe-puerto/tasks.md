# Tasks: H4 · Applet `boe`: puerto de `boe.py`

**Input**: `specs/005-h4-applet-boe-puerto/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` aplicada; H0-H3 cerrados (módulo,
binario multicall, `Makefile` con `ci`, `test-integration` y `test-e2e`, `.golangci.yml` con `depguard`, `forbidigo`,
`gosec` y `misspell`, kernel con `schema.ConClase`, `cli.Clasificar` y `cli.Montador`, `internal/httpx` con `Replay`
estricto y grabación, `internal/cache` con `SoloLectura` y `ConReloj`, `internal/arch_test.go`). **Ninguna tarea de H4
crea el esqueleto ni los gates: `make ci` existe y pasa desde H0**, y cada tarea lo deja en verde al terminar.

**Tests**: obligatorios (constitución §III; spec, User Scenarios). Cada tarea escribe su test antes del código que lo
hace pasar; los nombres son los del inventario de plan.md.

**Modo**: desatendido. Las tareas se ejecutan y verifican **una a una** por el workflow `hito`, con guardián de diff
por tarea. Por eso cada tarea es una **rebanada vertical** que deja `make ci` en verde y **declara en su propia línea
todas las rutas** que va a crear o modificar, y **solo** esas (plan, obligación 3).

## Formato: `[ID] [P?] [etiquetas] [Story] Descripción con rutas`

- **[P]**: la tarea no comparte fichero con sus vecinas ni depende de ellas (en el bucle del workflow van igual en
  secuencia).
- **[Story]**: historia del spec — US1 leer un artículo citable · US2 la segunda consulta no sale a la red · US3
  encontrar la norma · US4 recorrer la estructura y leer varios bloques · US5 estado y relaciones de la norma · US6
  contratos que no divergen · US7 enterarse de que la fuente cambió. Las fases 1 y 2 (fundación) y las tareas de
  cierre no llevan historia.
- **[datos]**: la tarea solo toca material bajo `testdata/` o `schemas/` y provoca una pausa de revisión humana. Son
  **seis**: T008 (manifiesto; en su pausa una persona fija la fila de la fuente y sus constantes, graba y escribe las
  referencias), T010 (sintéticos), T025 (guion de argumentos tolerante), T028 (guiones e2e), T030 (golden) y T032
  (esquemas). Ninguna mezcla otro trabajo ni toca código ni `.golangci.yml`.
- **[plataforma]**: la tarea necesita la plataforma remota. Solo hay **una**: T037, la última, único punto del hito
  autorizado a empujar la rama y a abrir la propuesta de cambio. Fusionar nunca es del workflow.

## Batería de verificación por tarea

- **`make ci`** en cada tarea: formato · lint (`depguard`, `forbidigo`, `gosec`, `misspell`, `paralleltest`…, con
  `run.build-tags` que desde T007 incluye `grabacion` y desde T034 `fuentes`) · tests con `-race` (unitarios, fuzz con
  semillas y e2e) · `test-integration` · vulnerabilidades · `schema-check` (real desde T029) · secretos · integridad y
  `tidy -diff` de módulos.
- **Cero red en todo el bucle**: todo test es offline contra `httpx.Replay` sobre las grabaciones y los sintéticos, con
  la caché en `t.TempDir()` o en el `$WORK` del guion. **`KITLEGAL_RECORD`, `scripts/grabar-fixtures.sh` y
  `make verify-sources` no se ejecutan en ninguna tarea**: la grabación la hace una persona en la pausa de T008 y la
  verificación contra la red, el flujo nocturno.
- **Ninguna dependencia nueva** (FR-125): `go.mod` no gana módulos; `go.sum` solo si `go mod tidy` lo exige.
- **Sin atajos**: ni `//nolint`, ni linter desactivado, ni exclusión nueva, ni test saltado, ni error silenciado, ni
  umbral rebajado, ni `panic`. Si un test queda legítimamente en rojo porque su implementación pertenece a otra tarea,
  la tarea está mal delimitada: se anota en `gates/tarea-Tnnn.md`, se redelimita y se deja sin marcar.
- **Lo que fija la persona no lo toca el ejecutor**: desde T009, ninguna línea de tarea contiene la ruta completa de la
  fila de la fuente, de las constantes de la fuente, del manifiesto, de las grabaciones ni de las referencias, ni de un
  directorio que las contenga; se nombran sin directorio (`SOURCES.md`, `terminos.go`, `grabaciones.json`, «las
  grabaciones», «las referencias») o por el test que las lee, de modo que el guardián rechaza cualquier cambio en ellas
  (plan, obligación 3.2; comprobado reproduciendo la extracción de rutas del workflow sobre cada línea).
- **`rtk`**: toda orden cuya salida se filtre o compare (`go test -v | grep`, `git status --porcelain`, `git diff`) se
  ejecuta con `rtk proxy` (plan, obligación 12).

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —test antes que código, `make ci` en verde
por sí solas— T001-T007, T011-T022, T026, T029 y T034. Las demás, cada una por su razón, tampoco dejan nada a medias:
**T008, T010, T025, T028, T030 y T032** son `[datos]` y solo pueden tocar `testdata/` o `schemas/` (en T008 y T010 los
ficheros que añaden no los lee ningún test hasta la tarea siguiente; T025, T028, T030 y T032 los validan tests ya
existentes); **T009, T023, T024, T031 y T033** son tests sobre código o datos ya existentes, el único caso en que test e
implementación pueden ir separados (T009 cubre lo que confirmó la persona; T031 y T033, lo que generaron T030 y T032);
**T027** cambia la raíz de composición del binario de e2e, que los guiones existentes siguen ejerciendo en verde, y su
comprobación específica —`cronometra` y los seis verbos por el binario— son los guiones de T028, que no pueden ir en la
misma tarea por la regla `[datos]`; **T035** es documentación; **T036** es validación sobre el árbol terminado, cuyo único
fichero escrito es la evidencia; **T037** solo publica. En T001, `internal/core/source.go` no tiene sentencias (una
interfaz y un tipo) y lo vigilan `compruebaDominioPuro` y la lista `core` de `depguard`; el resto de T001 lleva su test.

---

## Phase 1: Dominio, kernel y red (fundación)

**Objetivo**: que el dominio declare el puerto de fuente y la fecha de consulta, que el kernel feche el sobre con ella y
que `httpx` sepa pedir un formato y decir cuándo emitió, antes de que exista ningún adaptador (constitución: `core` →
adaptador → applet).

- [X] T001 El dominio declara el puerto de fuente y la fecha de consulta, y el kernel la usa: `internal/core/schema/sobre.go` añade `Procedencia.FechaConsulta time.Time` (el valor cero significa «no la declara quien consulta»; `Validar` no cambia) con `TestProcedenciaFechaDeConsultaOpcional` en su fichero de test; `internal/cli/sobre.go` hace que `Montador` feche el sobre con `FechaConsulta` cuando la procedencia es válida y la fecha no es cero, en éxito y en fallo, y con su reloj en otro caso (procedencia inválida o cero → `kitlegal.cli` y `kitlegal:cli` con el reloj del montador), sin que la huella dependa de la fecha, con `TestMontadorFechaDeConsulta` (con fecha → esa, en éxito y en fallo; sin fecha → reloj; procedencia inválida → kernel y reloj) en su fichero de test; `internal/core/source.go` con `Source` (`Name`, `Fetch(ctx, ec, consulta)`, `TTL(consulta)`, `Terms`), `Consulta` (`Verbo()`) y `Terminos{URL, Revisados}`, importando solo `context`, `time` y el paquete `schema` y sin sentencias; `internal/core/doc.go` con «Source nace con H4»; y `docs/ADR/0015-puerto-source-y-fecha-de-consulta.md` con el contexto de FR-096, la firma del puerto, la ampliación retrocompatible de `Procedencia` sin tocar `Applet` ni `Resultado` y las alternativas descartadas de research D2 y D3 (FR-096, FR-121, contrato puerto-y-applet §1 y §2, punto 7 de la Definition of Done).

- [X] T002 [P] `httpx` pide el formato de cada recurso: `internal/httpx/peticion.go` añade `Peticion.Acepta` (vacío → la petición no lleva `Accept`); `internal/httpx/cliente.go` lo valida con `mime.ParseMediaType` y sin caracteres de control (inválido → «argumentos», código 2, sin abrir nada, también en ensayo) y pone `Accept` en la petición y en cada salto de redirección; `internal/httpx/robots.go` garantiza que la obtención del `robots.txt` no lo lleva; con `TestPedirConAcepta` en `internal/httpx/cliente_test.go` (cabecera enviada y grabada; vacía → sin cabecera; inválida → 2 sin petición; ausente en la petición del `robots.txt`; la reproducción sigue emparejando solo por método y dirección) (FR-003, contrato httpx-acepta-e-instante §1, research D4).

- [X] T003 `httpx` dice cuándo emitió cada petición: `internal/httpx/instante.go` (nuevo) con la marca de emisión en el contexto, de clave sin exportar, y `conMarcaDeEmision`, que escribe la hora justo antes de entregar la petición al transporte y que cada intento sobrescribe; `internal/httpx/peticion.go` añade `Respuesta.Instante` (cero en ensayo); `internal/httpx/errores.go` añade `Error.Instante`; `internal/httpx/cliente.go` añade `ConHora(ahora func() time.Time)` (nula → «argumentos» al construir; vale para `New` y para `Replay`), coloca la marca en las dos cadenas en el orden del contrato §4, la reinicia antes de cada salto y la copia al volver en `Respuesta.Instante` o `Error.Instante` (vacía → la hora al volver); `internal/httpx/robots.go` obtiene su fichero con un contexto que oculta la marca; con `TestConHoraRechazaNula` en `internal/httpx/cliente_test.go` y `TestInstanteDeEmision` en `internal/httpx/instante_test.go`, con los nueve subtests `acierto`, `acierto-tras-reintentos`, `fallo-tras-reintentos`, `denegada-por-robots`, `sin-turno`, `redireccion`, `reproduccion`, `ensayo` y `argumentos`, sobre servidores de prueba de la biblioteca estándar (solo dentro de este paquete, regla R2) y un reloj que da un instante distinto en cada llamada; y, porque el contrato §4 pone la marca entre la identificación y la reproducción y el instante sale de la hora del cliente, `internal/httpx/reproducir_test.go` ajusta el subtest «identifica la petición» de `TestReplayGarantiasVigentes` (bajo la identificación, la marca de emisión, y bajo ella la reproducción y nada más) y `TestReplayEsDeterminista` (los dos clientes con la misma `ConHora` fija), e `internal/httpx/grabar.go` corrige solo el comentario de `decoradorDeGrabacion`, que deja de ser el escalón de más abajo; y, porque su verificación (`make ci`) se puso en rojo dos veces por `TestContextoCanceladoDuranteLaMigracion` del paquete de la caché, un fallo intermitente de H3 ajeno a `httpx` (SQLITE_BUSY en la lectura directa que sigue al fallo de `New`, con la conexión de la migración todavía abierta porque el paquete `sql` de la biblioteca estándar deshace en su propia goroutine, `Tx.awaitDone`, la transacción cuyo contexto termina): `internal/cache/migraciones.go` hace que `empiezaLaTransaccion` abra la transacción de la migración con `context.WithoutCancel(ctx)`, de modo que deshacerla sea cosa de `aplica` y ocurra antes de volver, manteniendo el contexto de quien llama en cada sentencia y en la espera por tramos; e `internal/cache/migraciones_test.go` exige en `TestContextoCanceladoDuranteLaMigracion`, justo después del fallo de `New`, que no existan `-wal` ni `-shm` (rojo sin el arreglo, verde con él); ninguna otra línea de la caché (FR-095, FR-096, SC-008, contrato httpx-acepta-e-instante §2-§4, research D4; contrato de H3 esquema-y-apertura §5, «ninguno deja conexión abierta»; redelimitada en los intentos 1 y 2, ver gates/tarea-T003.md).

**Checkpoint**: el dominio sabe qué es una fuente y el sobre puede llevar la fecha real de la consulta; `httpx` pide XML
o JSON y fecha cada petición. El binario no cambia.

---

## Phase 2: El porte documentado y las piezas puras del adaptador (fundación)

**Objetivo**: leer `boe.py` línea a línea y anotarlo **antes** de portar (FR-120), y construir lo que no depende de
ninguna respuesta de la fuente: errores con clase, identificadores, direcciones, consulta de búsqueda y consultas.

- [X] T004 [P] El porte anotado antes de portar: `internal/source/boe/doc.go` crea el paquete `boe` con el comentario del paquete y su sección de comportamientos no obvios de la copia congelada `boe.py`, con las 32 entradas de FR-120 una por línea de lista, cada una con su número, sus líneas, el comportamiento, el destino (se porta · se adapta y por qué · no se porta y por qué) y el requisito o *Fuera de alcance*, en la forma del contrato esquemas-fixtures-y-controles §9; con `TestDocAnotaElPorte` en `internal/source/boe/doc_test.go` (analiza `doc.go` con el analizador sintáctico de la biblioteca estándar y exige las entradas 1 a 32, una vez cada una, con líneas, destino y requisito); y `.golangci.yml`, solo `misspell.ignore-rules`, con `materias` y `regulares`, cada una con su comentario (FR-120, SC-010, research D18).

- [X] T005 Errores con clase e identificadores de entrada: `internal/source/boe/errores.go` con `Error{URL, Instante, Causa}`, `Error()`, `Unwrap()` y `Clase()` (implementa `schema.ConClase`) y sus constructores sin exportar por clase («argumentos», «no encontrado», «fuente no disponible», «límite o TOS», «inesperado»), con `TestErrorDeBoeMensajes` en su fichero de test (cada forma de mensaje del contrato errores-y-codigos §2, y la clase y el código que `cli.Clasificar` y `cli.CodigoSalida` dan a cada constructor, también envuelto con `%w`); `internal/source/boe/ids.go` con `ValidarNorma` y `ValidarBloque` (gramáticas de data-model §5; fuera de ellas, «argumentos» nombrando el valor y la forma esperada) y `TipoDesdeID` con las diez reglas en su orden (data-model §4); con `TestValidarNorma`, `TestValidarBloque` (vacío, separadores de ruta, `?`, `#`, `%`, espacios, controles, 65 caracteres), `TestTipoDesdeID` (`a21`, `da3`, `dt1`, `ti`, `cv3`, `preambulo`, uno sin regla) y `FuzzIDDeBloque` (semillas `a21`, `da3` y `dt1`; nunca entra en pánico, todo id aceptado es seguro como segmento de la petición y el tipo inferido es determinista) en `internal/source/boe/ids_test.go`; y `.golangci.yml`, solo `misspell.ignore-rules`, con `capitulo` y `disposicion`, cada una con su comentario (FR-040, FR-080, FR-081, FR-082, FR-100, SC-007, research D9, D10 y D18).

- [X] T006 Direcciones, consulta de búsqueda y consultas: `internal/source/boe/fuente.go` nace solo con la constante `NombreDeLaFuente` = `boe.legislacion-consolidada`, que usan el arnés de grabación y la reproducción (el resto de la fuente llega en T016); `internal/source/boe/direcciones.go` con la base de la API y las direcciones del bloque, el índice, los metadatos, el análisis y la norma, y las públicas de `act.php` de la norma y del bloque (data-model §8), con `TestDirecciones` (todas `https` de `www.boe.es`, ninguna en el espacio reservado); `internal/source/boe/busqueda.go` con la consulta de `boe.py` (con ` AND `, ` OR `, ` NOT `, `titulo:`, `materia:` o una comilla doble → tal cual; si no, cada palabra como `titulo:<palabra>` unidas con ` AND `; cero palabras → «argumentos»; límite de 10 resultados) y la codificación byte a byte como `json.dumps` seguido de `quote` (research D8), con `TestConsultaDeBusqueda` (operadores, varias palabras con la dirección exacta de `procedimiento administrativo común` del contrato verbos-y-salidas §1, una palabra recortada, espacio en blanco de Python, vacía → 2) y `TestCodificacionComoPython` (ASCII, no ASCII, fuera del plano básico, comillas, barra invertida, controles, barra); `internal/source/boe/consultas.go` con `ConsultaBuscar`, `ConsultaIndice`, `ConsultaArticulo`, `ConsultaArticulos`, `ConsultaMetadatos` y `ConsultaAnalisis`, que implementan `core.Consulta`, con `TestVerbosDeLasConsultas`; y `.golangci.yml`, solo `misspell.ignore-rules`, con `administrativo` y su comentario (FR-002, FR-003, FR-030, FR-031, research D2, D8 y D18).

**Checkpoint**: el porte está anotado y las piezas que no dependen de la fuente existen y están probadas. Nada consulta
todavía la red ni la caché.

---

## Phase 3: La fuente y sus datos protegidos

**Objetivo**: que una persona fije la fila de la fuente y sus constantes, grabe con el ritmo revisado y escriba las
referencias del diff de aceptación **antes** de que exista ningún código de lectura, y que el ejecutor compruebe después
que todo está, sin poder cambiarlo.

- [X] T007 [US6] Arnés de grabación y fila de la fuente como propuesta (el ejecutor no graba ni escribe ninguna fecha de revisión): `internal/source/boe/grabacion_test.go` con `//go:build grabacion` y `TestGrabarFixtures`, que exige `KITLEGAL_RECORD=1` (si no, `t.Fatal`), lee el manifiesto `grabaciones.json` del paquete (recursos `busqueda`, `indice`, `metadatos`, `analisis` y `bloque`) y pide cada recurso con `httpx.New` (`ConFuente(NombreDeLaFuente)`, `ConRaizDeGrabacion` sobre la carpeta de datos de prueba del paquete, `ConIntervalo(IntervaloEntrePeticiones)`), con la dirección y el `Accept` que construyen `direcciones.go` y `busqueda.go` y sin ninguna lectura de respuestas; `scripts/grabar-fixtures.sh`, que ejecuta exactamente esa prueba con la etiqueta `grabacion` y `KITLEGAL_RECORD=1` como fija el contrato esquemas-fixtures-y-controles §3.2, y que ninguna tarea ejecuta; `docs/SOURCES.md` con la tabla del contrato §8 y la fila **propuesta** `boe.legislacion-consolidada` (applet `boe`, base de la API, licencia y `robots.txt` por revisar, URL de términos propuesta, ritmo `1s`, XML y JSON sin autenticación, «Revisado» `pendiente`); `internal/source/boe/terminos.go` con `IntervaloEntrePeticiones = time.Second` y `terminosDeUso` iguales a la propuesta, con `Revisados` cero; `internal/source/boe/terminos_test.go`, interno, con `TestFuenteCoincideConSources` (localiza la fila por su primera celda y exige ritmo igual a `IntervaloEntrePeticiones`, URL de términos igual a `terminosDeUso.URL` y «Revisado» igual a `terminosDeUso.Revisados` como fecha `AAAA-MM-DD` no cero, admitiendo **exactamente** la pareja `pendiente` con cero); y `.golangci.yml` con `grabacion` en `run.build-tags` (FR-113, FR-121, FR-122, FR-123, punto 8 de la Definition of Done, research D12, D13 y D15).

- [X] T008 [datos] [US6] Manifiesto de grabación, la lista cerrada que una persona grabará: `internal/source/boe/testdata/grabaciones.json` con `fuente` `boe.legislacion-consolidada` y los 22 recursos de la tabla del contrato esquemas-fixtures-y-controles §3.1, cada uno con su `para`, sin dirección ni `Accept` (los construye el código); ningún otro fichero. En la pausa humana que provoca, una persona sigue el procedimiento del contrato §3.2 en su orden —revisión de los términos de uso y del `robots.txt` (si prohíben el acceso automatizado, rechaza y el hito se detiene: FR-123), fila definitiva con la fecha real de la revisión, constantes de la fuente alineadas con ella y comprobadas con `TestFuenteCoincideConSources`, grabación con el ritmo revisado, comprobación de S1, S2, S4 y S7 con los suplentes si hacen falta, y las cinco referencias escritas a mano desde `boe.py` y revisadas— y lo confirma en la rama; el ejecutor no hace nada de ello (FR-113, FR-116, FR-123, SC-001, SC-005, research D12 y D15).

- [X] T009 [US1] Grabaciones, referencias y fila revisadas, antes de todo código de lectura: `internal/source/boe/casos_test.go` con el manifiesto y las seis referencias como casos cerrados, sus ayudas de lectura por `filepath.Clean`, `TestGrabacionesCompletas` (cada entrada de `grabaciones.json` la sirve `httpx.Replay` sobre las grabaciones, con la dirección que construyen `direcciones.go` y `busqueda.go`) y `TestReferenciasCompletas` (exactamente seis referencias de cuatro normas, una `BOE-A-2015-10565-a21`, de la lista del contrato esquemas-fixtures-y-controles §5 o con sus suplentes; cada una sin claves desconocidas, con los ocho campos y un `boe_py` no vacío en cada uno, y con `grabacion_bloque` y `grabacion_metadatos` que nombran grabaciones existentes; no compara contenido); `internal/source/boe/ids_test.go` añade `TestGramaticaCubreLosIndicesGrabados` (todo id de bloque de las tres grabaciones de índice, recorridas con el decodificador JSON genérico y sin el código de lectura, casa con la gramática de `ValidarBloque`); `internal/source/boe/terminos_test.go` retira de `TestFuenteCoincideConSources` la admisión de `pendiente`, de modo que desde aquí la fila de `SOURCES.md` y las constantes de `terminos.go` tienen que coincidir con una fecha real; e `internal/source/boe/grabacion_test.go`, el arnés de grabación, deja de declarar su propio lector del manifiesto, sus recursos y la construcción de sus peticiones y usa los de `internal/source/boe/casos_test.go`, de modo que se reproduce exactamente la petición que se graba y el código no queda duplicado (redelimitada en el intento 1, ver gates/tarea-T009.md). La tarea **no cambia** la fila, `terminos.go`, el manifiesto, las grabaciones ni las referencias: si falta una grabación o una referencia, si la fila sigue en `pendiente`, o si la grabación de un recurso inexistente (`a9999`, `BOE-A-2099-99999`) no responde 404 (S2), se detiene sin marcarse y lo anota en `gates/tarea-T009.md` para decisión humana (FR-080, FR-113, FR-116, FR-122, SC-001, SC-005, obligación 2 del plan).

- [X] T010 [datos] [US1] Sintéticos, antes del código de lectura: `internal/source/boe/testdata/sintetico/` con los siete escenarios del contrato esquemas-fixtures-y-controles §4 (`bloque-ilegible`, `bloque-sin-elemento`, `metadatos-caidos`, `metadatos-ilegibles`, `fuente-caida`, `limite` y `avisos`), cada uno en la subcarpeta con el nombre de la fuente, copiados de la grabación de origen que nombra la tabla (bloque `a21`, metadatos o índice de `BOE-A-2015-10565`) con el cambio mínimo que fija el contrato y sin cambiar la petición grabada (método y dirección), para que la reproducción la empareje; ningún otro fichero (FR-013, FR-014, FR-100, SC-008, SC-012, research D12).

**Checkpoint**: la fuente tiene fila revisada, constantes atadas a ella, grabaciones reales, sintéticos y las cinco
referencias escritas por una persona, y un test que falla si algo de ello falta.

---

## Phase 4: Lecturas, de dentro afuera

**Objetivo**: interpretar lo que responde la fuente con las reglas de `boe.py` y las adaptaciones del spec, con tablas
en `_test.go`, antes de que exista ningún verbo.

- [X] T011 [US5] Lectura de las respuestas JSON: `internal/source/boe/lectura.go` con las reglas J1-J9 de data-model §3.1 —raíz objeto; `data` ausente o nula como vacía; objeto suelto → lista de un elemento donde FR-070 lo pide, también dentro de los envoltorios; primer elemento de `data` en metadatos y análisis; índice anidado o plano; envoltorios `materia`, `nota`, `anterior` y `posterior`; elementos que no son objeto, saltados; texto: cadena, ausente o nulo → vacío, cualquier otro tipo → «fuente no disponible» nombrando el campo, nunca el marcador `?`—, con `TestLeerEnvoltorio`, `TestComoLista` y `TestTextoDe` de tabla, sin ficheros de datos (FR-016, FR-040, FR-070, research D7).

- [X] T012 [US1] Lectura del bloque XML con el recorrido de `ElementTree`: `internal/source/boe/bloque.go` con las reglas X1-X8 de data-model §3.2 —UTF-8 válido; declaración de codificación ignorada; raíz única; el primer descendiente `bloque` que no es la raíz; atributos `titulo` y `tipo` del bloque y `fecha_publicacion`, `fecha_vigencia` e `id_norma` de la última `version` de primer nivel (fecha `original` si falta o sin versiones; vacíos los demás); texto con `CharData`, `CDATA` y `tail`, sin que comentarios ni instrucciones de proceso lo corten; normalización de líneas— y cualquier cuerpo ilegible o sin bloque → «fuente no disponible», nunca el cuerpo ni un recorte, con `TestLeerBloque` de tabla (con versiones, sin versiones, última sin fecha, fecha vacía presente, `tail`, CRLF, CDATA, comentario, entidades, atributos con saltos, bloque anidado, raíz `bloque`, ilegible, no UTF-8, declaración ISO-8859-1 con UTF-8, dos raíces) (FR-010, FR-011, FR-014, FR-016, research D6).

- [X] T013 [US5] Tipos de `data` y avisos de vigencia: `internal/source/boe/datos.go` con `Articulo`, `Aviso`, `ResultadoDeBusqueda`, `Indice`, `EntradaDeIndice`, `Metadatos`, `EstadoDeConsolidacion`, `Analisis`, `Materia`, `Referencias`, `ReferenciaAnterior` y `ReferenciaPosterior` con las claves de data-model §2 (todas obligatorias; vacío y lista vacía en lugar de ausente), etiquetas `json` y `jsonschema` (`codigo` enumerado con exactamente `consolidacion-no-finalizada`, `derogada` y `vigencia-agotada`; `tipo` de `EntradaDeIndice` con los valores de `TiposDeBloque`; `hash_texto` con su patrón), `TiposDeBloque()` y la huella `sha256:` del texto, con `TestEnumeradosDeLosDatos` (los enumerados del esquema que genera la biblioteca de esquemas ya fijada son exactamente `CodigosDeAviso()` y `TiposDeBloque()`, y todo tipo que da `TipoDesdeID` está entre ellos) y `TestHashTexto` (cambia si y solo si cambia el texto); e `internal/source/boe/avisos.go` con `CodigosDeAviso()` y los avisos derivados de los metadatos con las tres condiciones de `_check_vigencia`, su orden y sus frases literales (data-model §2.2), con `TestAvisosDe` (las ocho combinaciones; código numérico → sin aviso) y `TestCodigosDeAviso` (FR-012, FR-015, FR-050, SC-013, research D11).

- [X] T014 [US2] La consulta guardada: `internal/source/boe/entradas.go` con la clave `boe.legislacion-consolidada|1|<verbo>|<dirección>` y el contenido `fecha_consulta`, `url` y `datos`, con la fecha en RFC 3339 con nanosegundos y desplazamiento, leído sin admitir claves desconocidas (ilegible → «inesperado» nombrando la clave), con `TestClaveDeEntrada` (dos consultas distintas nunca comparten clave; la de metadatos es la misma para `metadatos`, `articulo` y `articulos`), `TestEntradaIdaYVuelta` (fecha y url byte a byte) y `TestEntradaIlegibleEsInesperado` (FR-090, FR-096, data-model §6, research D5).

- [X] T015 [US2] Pedir y clasificar: `internal/source/boe/peticiones.go` con la interfaz `Pedidor` del contrato puerto-y-applet §3.1, la petición GET con `Acepta` XML para el bloque y JSON para el resto, el instante tomado de `Respuesta.Instante` o de `Error.Instante`, el ensayo sin pedir, y la clasificación de las filas 6-17 del contrato errores-y-codigos (404 de bloque, índice, metadatos o análisis → «no encontrado» con su dirección; 404 en búsqueda y cualquier otro estado no 2xx → «fuente no disponible»; los errores de `httpx`, con su clase, su dirección y su instante), con `TestPedirClasificaEstados` sobre un `Pedidor` de prueba (200, 404 por recurso, 404 en búsqueda, 400, 403, ensayo, error de `httpx`, instante) (FR-003, FR-095, FR-100, FR-101, research D10).

**Checkpoint**: todo lo que la fuente puede responder se interpreta o se clasifica con su código, y lo que se guarda
tiene forma fija. Ningún verbo existe todavía.

---

## Phase 5: La fuente y los seis verbos

**Objetivo**: la fuente completa implementando `core.Source`, un verbo por tarea en el orden del plan
(`metadatos` → `articulo` → `articulos` → `indice` → `buscar` → `analisis`), y los tests que cruzan los seis.

- [X] T016 [US2] La fuente que implementa el puerto, y el comparador de golden: `internal/source/boe/fuente.go` completa `Fuente` con `Nueva(opciones…)` y las opciones `ConCliente` (se llama solo si hay que pedir), `ConCache` (se llama tras validar la consulta; solo lectura con `--offline` o `--dry-run`) y `ConRegistrador` (por omisión descarta), `Name`, `TTL` por consulta (300 s para `buscar` y `metadatos`; 604 800 s para `indice`, `articulo`, `articulos` y `analisis`), `Terms()` que devuelve `terminosDeUso`, `Fetch` que cierra la caché que abrió uniendo el error del cierre, y `var _ core.Source = (*Fuente)(nil)`; hasta que cada verbo llega en su tarea, `Fetch` trata su consulta como de otro tipo («inesperado»); con `TestNuevaRechazaDependenciasAusentes`, `TestFuenteNombreVigenciasYTerminos` y `TestConsultaDeOtroTipo` en `internal/source/boe/fuente_test.go`, y en ese mismo fichero el comparador `TestGolden` con la bandera `-actualizar-golden`, que compara byte a byte cada golden que exista con su caso (forma canónica del contrato esquemas-fixtures-y-controles §2, caché vacía, reproducción de las grabaciones), falla ante un fichero sin caso, no compara nada si no existe ninguno y escribe con `0o600` desde un auxiliar distinto del que lee; los trece casos de golden en `internal/source/boe/casos_test.go`; y `.golangci.yml`, solo `misspell.ignore-rules`, con `dependencias` y su comentario (FR-090, FR-091, FR-112, FR-121, FR-127, research D2, D5 y D18).

- [X] T017 [US5] `metadatos`: `internal/source/boe/metadatos.go`, con su caso en `Fetch` de `internal/source/boe/fuente.go`, valida la norma, lee su entrada (vigente → servida con su `url` y su `fecha_consulta`; ausente o caducada con `--offline` → 4 sin pedir; con `--dry-run`, una línea de ensayo y nada escrito), pide con JSON, clasifica, lee (`data` vacío → 3 sin guardar), compone `Metadatos` con los avisos de `avisos.go` y escribe la entrada con 300 s y el instante de la petición; `url` del sobre = la de los metadatos; con `TestMetadatos` (`vigente`, `derogada`, `inexistente`, `limite`, `tres-avisos`) sobre las grabaciones y los sintéticos con `httpx.Replay` y `httpx.ConHora`, y la caché en `t.TempDir()` con `cache.ConReloj` del mismo reloj, comprobando `fuente` y `url` de `www.boe.es` en éxito y en fallo (FR-002, FR-050, FR-051, FR-090-FR-094, FR-096, FR-101, US5 escenarios 1 y 3).

- [X] T018 [US1] `articulo`: `internal/source/boe/articulo.go`, con su caso en `Fetch` de `internal/source/boe/fuente.go`, sigue el flujo de data-model §7.1: valida norma y bloque; entrada del artículo vigente → servida; si no, pide el bloque con XML (404 → 3 sin pedir los metadatos; ilegible o sin bloque → 4), toma los metadatos de su entrada vigente o los pide y escribe esa entrada, falla con la clase y la `url` de los metadatos sin emitir el texto si no se obtienen o no se interpretan, compone `Articulo` (avisos, `url_eli`, `hash_texto`, dirección pública) con `fecha_consulta` igual a la más antigua de bloque y metadatos y escribe la entrada del artículo; `url` del sobre = la del bloque; con `TestArticulo` (`bloque-vigente`, `derogada`, `bloque-inexistente`, `metadatos-en-cache`, `metadatos-caidos`, `metadatos-ilegibles`, `bloque-ilegible`, `bloque-sin-elemento`, `tres-avisos`, con las peticiones contadas), `TestFechaDeConsultaDeArticulo` (reloj controlado: metadatos guardados en t0 y bloque pedido en t1 → t0, también al servirlo de su entrada) y `TestArticuloCoincideConBoePy` (un subtest por referencia: proyecta `data` sobre los ocho campos de FR-116, con los avisos como la lista ordenada de sus `texto`, y exige igualdad campo a campo nombrando el campo distinto; ante una discrepancia se corrige el código, nunca la referencia) (FR-010-FR-016, FR-090, FR-093, FR-096, FR-101, FR-116, SC-001, SC-012, SC-013, US1 escenarios 1-3, 5 y 6).

- [X] T019 [US4] `articulos`: `internal/source/boe/articulo.go` añade el verbo, con su caso en `Fetch` de `internal/source/boe/fuente.go`, siguiendo el flujo de data-model §7.2: valida la norma y todos los bloques antes de nada; recorre los ids distintos en el orden de su primera aparición; sirve los guardados; pide en secuencia los que faltan y los metadatos como mucho una vez por invocación (también en ensayo, descritos una sola vez); se detiene en el primer fallo con su clase y la `url` de la petición que falló, dejando escritos los bloques anteriores; `data` en el orden pedido y con repeticiones, `url` del sobre = la de la norma y `fecha_consulta` = la más antigua de sus elementos; con `TestArticulos` (`tres-bloques` con cuatro peticiones, `id-repetido` con dos, `segundo-inexistente` con tres y el primero en caché, `mezcla-de-cache` con la fecha más antigua) (FR-020, FR-021, FR-096, FR-101, SC-003, SC-008, US4 escenarios 2 y 3).

- [X] T020 [US4] `indice`: `internal/source/boe/indice.go`, con su caso en `Fetch` de `internal/source/boe/fuente.go`, valida la norma, sirve la entrada vigente, pide con JSON, lee el índice anidado o plano con el tipo de `TipoDesdeID` para cada bloque en el orden de la fuente (`data` vacío → 3 sin guardar) y escribe la entrada con 604 800 s; `url` del sobre = la del índice; con `TestIndice` (`lpac`, `inexistente`, `fuente-caida`) (FR-040, FR-041, FR-070, FR-091, US4 escenarios 1 y 4).

- [X] T021 [US3] `buscar`: `internal/source/boe/buscar.go`, con su caso en `Fetch` de `internal/source/boe/fuente.go`, une el texto con un espacio, valida (cero palabras sin operadores → 2 sin abrir la caché), construye la dirección con `busqueda.go`, sirve la entrada vigente, pide con JSON, lee los resultados en el orden de la fuente (objeto suelto → lista) con la dirección pública de cada uno, y escribe la entrada con 300 s también cuando no hay resultados; `url` del sobre = la de la búsqueda; con `TestBuscar` (`con-resultados`, `sin-resultados-guardada`, `texto-vacio`) (FR-030, FR-031, FR-032, FR-070, US3).

- [X] T022 [US5] `analisis`: `internal/source/boe/analisis.go`, con su caso en `Fetch` de `internal/source/boe/fuente.go`, valida la norma, sirve la entrada vigente, pide con JSON, lee materias, notas, referencias anteriores (con el texto completo, sin recorte) y posteriores con sus envoltorios y la normalización a lista (`data` vacío → 3 sin guardar) y escribe la entrada con 604 800 s; `url` del sobre = la del análisis; con `TestAnalisis` (`lpac`, `inexistente`, `texto-completo`); y, porque `analisis`, `metadatos` e `indice` comparten el mismo esqueleto de verbo (validar la norma antes de abrir nada, `invocar`, resolver con la caché de la invocación y componer el resultado con la dirección del recurso) y `dupl` lo marca en cuanto existe el tercero, `internal/source/boe/fuente.go` añade `resolverRecursoDeLaNorma`, que lo recoge, y los tres verbos lo usan: `internal/source/boe/analisis.go` con `analisisDeLaNorma`, `internal/source/boe/metadatos.go` con el `metadatosDeLaNorma` que ya tiene e `internal/source/boe/indice.go` con un `indiceDeLaNorma` nuevo, sin cambiar ningún comportamiento ni ningún test de esos dos verbos (FR-060, FR-061, FR-070, US5 escenario 2; redelimitada en el intento 1, ver gates/tarea-T022.md).

- [X] T023 [US2] La caché de los seis verbos, sobre el código ya existente: `internal/source/boe/fuente_test.go` añade `TestCacheDeLosSeisVerbos` (la segunda consulta idéntica contra `httpx.Replay` sobre una carpeta vacía: cero peticiones y el mismo `data`, `url` y `fecha_consulta` con el reloj adelantado dentro de la vigencia; pasada la vigencia de su verbo, se vuelve a pedir; `articulo` con los metadatos en caché emite una petición y `metadatos` tras un `articulo`, ninguna), `TestOfflineDeLosSeisVerbos` (entrada vigente → 0 con la fecha guardada; ausente o caducada → 4, cero peticiones y la caché intacta), `TestEnsayoDeLosSeisVerbos` (cero peticiones, nada escrito y una línea por cada petición que se habría emitido, sin repetir ninguna) y `TestFallosNoSeGuardan` (bloque inexistente, fuente caída, bloque ilegible y metadatos caídos: repetir la consulta vuelve a pedir); la tarea no añade producto (FR-090-FR-096, SC-003, SC-004, SC-012, US2 escenarios 1 y 3-8).

- [X] T024 [US6] Ninguna ruta de fallo sin clase, sobre el error real: `internal/source/boe/errores_test.go` añade `TestClasesDeErrorDeBoe`, con una subprueba por cada fila 2-4, 6-10 y 16-22 del contrato errores-y-codigos, provocando cada situación con el código ya existente (validadores, `Pedidor` de prueba, grabaciones y sintéticos, entrada de caché ilegible, `AperturaDeCache` de prueba, `Nueva` sin dependencias) y comprobando la clase de `cli.Clasificar` y el código de `cli.CodigoSalida`, que ninguna da 6 y que ninguna entra en pánico; la subprueba de la fila 21 usa `metadatos` de `BOE-A-2015-10565` sobre las grabaciones con una `AperturaDeCache` de prueba cuyo error implementa `schema.ConClase` y declara una de las tres clases que produce `cache.Error` («argumentos», «fuente no disponible» e «inesperado», cada una en al menos un caso), y cubre cuatro casos: la apertura falla (ninguna petición); `Get` falla fuera de solo lectura (ninguna petición); `Put` falla después de obtener la respuesta (una petición y ninguna entrada escrita); y `Close` falla, tras una invocación correcta (clase del cierre) y tras una fallida (clase del fallo de la invocación, con los dos errores unidos); en cada caso exige además que el error de la caché siga alcanzable con `errors.Is`, `fuente` `boe.legislacion-consolidada`, la `url` de los metadatos y la fecha de consulta sin declarar, para que la ponga el montaje; la tarea no añade producto (FR-090, FR-100, SC-008, punto 5 de la Definition of Done).

**Checkpoint**: la fuente responde los seis verbos contra las grabaciones con caché, `--offline`, ensayo y códigos
estables, y `articulo` coincide con las cinco referencias escritas a mano desde `boe.py`.

---

## Phase 6: El applet en los binarios y el e2e de la entrega

**Objetivo**: `kitlegal boe …` y `boe …` en el binario distribuido sin `panic`, el binario de e2e con `boe` sobre la
reproducción y los guiones que describen la entrega en la primera tarea que puede dejarlos en verde.

- [X] T025 [datos] [US6] Guion de argumentos tolerante, antes de registrar `boe` en el binario de e2e: `internal/app/testdata/script/argumentos.txtar`, líneas 24 y 30, pasan de `stderr 'applets disponibles: contar, echo'` a `stderr 'applets disponibles: (boe, )?contar, echo'`, que vale antes y después del registro; ningún otro cambio ni fichero (FR-001, FR-114, research D13).

- [X] T026 [US1] El applet `boe` en el binario distribuido: `internal/app/boe.go` con `DependenciasDeBoe`, `DependenciasDeRed()` (cliente con `httpx.New`, `ConFuente(boe.NombreDeLaFuente)`, `ConIntervalo(boe.IntervaloEntrePeticiones)` y el registrador) y `AppletBoe` con los seis verbos y los argumentos del contrato puerto-y-applet §4, ninguno por omisión, que construye la fuente en cada invocación (caché con `cache.New`, en solo lectura con `--offline` o `--dry-run`) y devuelve `Fetch` tal cual; con `TestAppletBoe`, `TestDependenciasDeRed` (sin red), `TestCodigosDeSalidaDeBoe` (filas 1-6, 11, 14, 18, 19 y 21 del contrato errores-y-codigos por `app.Main` en proceso sobre las grabaciones y los sintéticos de la fuente: código, clase, `fuente`, `url` y `fecha_consulta` con reloj controlado; la fila 21 con `DependenciasDeBoe` cuya `Cache` lleva `cache.ConDirectorio` sobre un fichero regular creado en `t.TempDir()`, de modo que `cache.New` falla al abrir con «argumentos» y el sobre de `articulo BOE-A-2015-10565 a21` sale con código 2, `fuente` `boe.legislacion-consolidada`, la `url` del bloque consultado, `fecha_consulta` del montaje y ninguna petición) y `TestSinGrafoNiAsuntoNoCambianLaSalida` en `internal/app/boe_test.go`; `internal/app/registro.go` con `RegistroDeProduccion() (*Registro, error)` que registra `AppletBoe(DependenciasDeRed())`, y `TestRegistroDeProduccion` (exactamente `boe`) en `internal/app/registro_test.go`; `internal/app/main.go` con `Arrancar` (error del registro → sobre `kitlegal.cli` de clase «inesperado», mensaje en la salida de error y código 1; si no, `Main`), y `TestArrancar` (`registro-valido`, `registro-invalido`) en `internal/app/main_test.go`; `internal/app/ayuda_test.go` e `internal/app/despacho_test.go` pasan a `&Registro{}` solo en `TestDespachoConRegistroVacio` y en la subprueba del registro vacío, con sus comentarios reescritos sobre ese sujeto; `cmd/kitlegal/main.go` termina con `os.Exit(app.Arrancar(os.Args, app.RegistroDeProduccion, …))` y `cmd/kitlegal/main_test.go` ejerce esa misma composición (`applets disponibles: boe` en «sin applet», «inventado» y «echo no es del binario distribuido»; nuevo caso «`boe` sin verbo → 2»; el comentario de `TestVersionDeLaIdentificacion` deja de afirmar que el binario no enlaza la red); e `internal/arch_test.go` retira `TestElBinarioNoEnlazaHTTPX` y `TestElBinarioNoEnlazaCache`, amplía `modulosDelBinario` con los módulos que `go list -deps` mide al enlazar `boe`, cada uno justificado en su comentario (y anotados con su justificación en `specs/005-h4-applet-boe-puerto/gates/pr-h4.md`), y añade `TestLasFuentesNoFirmanComoKitlegal` (ningún literal que empiece por `kitlegal.` o `kitlegal:` en los ficheros de producción de los paquetes de fuentes; exige al menos un paquete y nombra fichero y línea) (FR-001, FR-002, FR-100, FR-101, FR-122, FR-124, FR-127, FR-128, SC-008, SC-011, SC-013, research D14 y D16, US1 escenarios 4-6).

- [X] T027 [US6] El binario de e2e responde con `boe` sin red: `internal/app/ejemplo/kitlegal-e2e/main.go` deja el `panic` y termina con `os.Exit(app.Arrancar(…))` sobre un registro de `echo`, `contar` y `boe`, cuyo cliente es `httpx.Replay` sobre la carpeta relativa de reproducción con el nombre de la fuente (`ConFuente(boe.NombreDeLaFuente)`, registrador), sin leer ninguna variable de entorno; `internal/app/e2e_test.go` gana en `TestEntregaDelHito` un `Setup` que copia con `os.CopyFS` las grabaciones de la fuente a la carpeta de reproducción del `$WORK` del guion y fija `KITLEGAL_CACHE_DIR` dentro de `$WORK`, y la orden `cronometra <máximo> <programa> <argumentos…>` (con `TestScript.Exec`; falla si el código no es 0 o si la duración alcanza el máximo), con todos los guiones existentes en verde, `argumentos` incluido gracias a T025 (FR-114, FR-117, research D13 y D16).

- [X] T028 [datos] [US6] Los guiones e2e de la entrega, en la primera tarea que puede dejarlos en verde: `internal/app/testdata/script/boe-verbos.txtar`, `internal/app/testdata/script/boe-multicall.txtar`, `internal/app/testdata/script/boe-offline.txtar`, `internal/app/testdata/script/boe-codigos.txtar` e `internal/app/testdata/script/boe-cache-rapida.txtar` con lo que fija la tabla del contrato esquemas-fixtures-y-controles §6 (los seis verbos con `--json` sobre la reproducción, `--describe` de los seis y la ayuda; el enlace `boe` con la misma salida; `--offline` con y sin entrada; los códigos 2, 3 y 4 con su clase y su `url`; diez invocaciones de `cronometra 200ms` con la reproducción vaciada), e `internal/app/testdata/script/argumentos.txtar`, líneas 24 y 30, ajustadas a `stderr 'applets disponibles: boe, contar, echo'`; ningún otro fichero (FR-001, FR-092, FR-100, FR-101, FR-114, FR-117, SC-002, SC-004, SC-005, US1 escenario 4, US2 escenarios 2 y 3, US6 escenario 4).

**Checkpoint**: `kitlegal boe articulo BOE-A-2015-10565 a21` funciona en el binario distribuido y, sobre el binario de
e2e sin red, los seis verbos, el enlace, `--offline`, los códigos y los 200 ms en caché están cubiertos.

---

## Phase 7: Contratos publicados

**Objetivo**: golden y esquemas en tres pasos (comparador → `[datos]` → cobertura) para que ningún test quede
desactivado y `make ci` siga en verde en cada tarea.

- [X] T029 [US6] El comparador de esquemas y `make schema-check` real: `internal/app/esquemas_test.go` con `TestEsquemasPublicados`, que regenera en memoria `norma.json` (`analisis`, `buscar`, `indice`, `metadatos`) y `bloque.json` (`articulo`, `articulos`) desde `--describe` de los verbos registrados en producción, en la forma canónica del contrato esquemas-fixtures-y-controles §1 con el `$id` de cada verbo, compara por verbo y entero cada fichero publicado que exista, con los dos mensajes del contrato que nombran fichero y verbo, no compara nada si no existe ninguno, y con la bandera `-actualizar-esquemas` los escribe con `0o600` desde un auxiliar distinto del que lee; y `Makefile`, cuyo objetivo `schema-check` deja de ser un aviso y ejecuta solo ese test en el paquete del applet, con `check-tools` como prerrequisito, y sigue en `ci` (FR-110, SC-006, research D11).

- [X] T030 [datos] [US6] Golden de `data`: `internal/source/boe/testdata/golden/` con los trece ficheros de los casos del contrato esquemas-fixtures-y-controles §2, generados con `TestGolden` y su bandera `-actualizar-golden` mediante la orden de esa sección (no se copia aquí) y revisados en la pausa; ningún otro fichero (FR-112, SC-005, US6 escenario 3).

- [ ] T031 [US6] Cobertura de golden: `internal/source/boe/fuente_test.go` añade `TestGoldenCubreTodosLosCasos` (existen los trece golden de la lista cerrada y cubren los seis verbos; uno de más o de menos falla nombrándolo) (FR-112, SC-005).

- [ ] T032 [datos] [US6] Esquemas publicados: `schemas/norma.json` y `schemas/bloque.json` generados con `TestEsquemasPublicados` y su bandera `-actualizar-esquemas` mediante la orden de la sección §1 del contrato esquemas-fixtures-y-controles (no se copia aquí) y revisados en la pausa, incluido el enumerado de `codigo` con exactamente sus tres valores en `articulo`, `articulos` y `metadatos`; ningún otro fichero (FR-012, FR-110, SC-006, US6 escenario 1).

- [ ] T033 [US6] Contratos que no divergen de lo emitido: `internal/app/esquemas_test.go` añade `TestEsquemasCubrenTodosLosVerbos` (los dos ficheros publicados existen y cada verbo registrado en producción está en exactamente uno), e `internal/app/boe_test.go` añade `TestSalidaDeBoeContraSchemas` (el sobre real de éxito y los de fallo con códigos 2, 3 y 4 de los seis verbos, validados con el validador de esquemas ya fijado y `AssertFormat` contra la parte de su verbo leída de `norma.json` o `bloque.json` y compilada por su `$id`; un `data` con una clave de más o de menos no valida; `Aviso.codigo` con exactamente tres valores); la tarea no añade producto (FR-012, FR-111, SC-006, punto 4 de la Definition of Done, US6 escenarios 1 y 2).

**Checkpoint**: lo que `--describe` emite coincide con lo publicado, toda salida valida contra su esquema leído del
fichero y un byte distinto en `data` hace fallar su golden.

---

## Phase 8: Verificación nocturna, documentación y cierre

- [ ] T034 [US7] Enterarse de que la fuente cambió: `internal/app/fuentes_test.go` con `verificarArticulo` (invoca `boe articulo BOE-A-2015-10565 a21 --json` por `app.Main`; exige código 0, sobre válido contra la parte `articulo` de `bloque.json` y `data.texto` no vacío; su error empieza por «caso «boe articulo»:») y `TestVerificacionDeFuentesDetectaCambios` (con la reproducción de las grabaciones → sin error; con el sintético `bloque-ilegible` → error que nombra el caso); `internal/app/fuentes_red_test.go` con `//go:build fuentes` y `TestVerificarFuentes` (`AppletBoe(DependenciasDeRed())` y la caché en `t.TempDir()`); `scripts/verify-sources.sh`, que ejecuta solo ese test con la etiqueta `fuentes`; `Makefile` con el objetivo `verify-sources` (prerrequisito `check-tools`) que llama al script y queda fuera de `ci`; `.github/workflows/nightly.yml` con el trabajo `fuentes` del contrato esquemas-fixtures-y-controles §7 (`contents: read`, `issues: write`, `make verify-sources` y, si falla, comentar la incidencia abierta «verify-sources: boe articulo» o crearla); y `.golangci.yml` con `fuentes` en `run.build-tags`; ni el script ni `make verify-sources` se ejecutan en ninguna tarea, porque tocan la red (FR-115, SC-009, punto 8 de la Definition of Done, research D13 y D15, US7).

- [ ] T035 Documentación alineada con los objetivos de `make` y registro de cambios: `CHANGELOG.md` › *Unreleased* › *Añadido* (applet `boe` con sus seis verbos, esquemas publicados `norma.json` y `bloque.json`, `make verify-sources` y la verificación nocturna) y › *Cambiado* (`make schema-check` activo; `fecha_consulta` declarada por la fuente, también en lo servido desde la caché; `app.Arrancar`); `README.md` y `CONTRIBUTING.md` con las filas de `make schema-check` (activo) y `make verify-sources` (con red, solo en el flujo nocturno); y `docs/PENDIENTES.md` sin las dos entradas de H4 que el hito resuelve (dónde viven los fixtures grabados; lo que el primer adaptador retira y amplía) (FR-126, punto 6 de la Definition of Done, research D17, obligación 8 del plan).

- [ ] T036 Cierre del hito con la medida hecha, sin tocar el árbol: ejecutar los escenarios 1 a 16 de `quickstart.md` salvo el 15.b (el único con red), con cada orden tal cual la escribe el quickstart y sin sustituirla por otra a criterio del ejecutor (sus formas son las que la sesión desatendida ejecuta sin pedir aprobación: `rtk proxy` delante de lo que la lista de permitidos no cubre y de lo que ejecuta, crea, borra o edita en la carpeta temporal, el código de salida impreso dentro de `rtk proxy sh -c`, las variables con `rtk proxy env` y nada redirigido a ficheros; las sondas de sus prerrequisitos las vuelven a comprobar, y si una orden pide aprobación o una sonda no da lo esperado la tarea se detiene y lo anota en la evidencia en vez de improvisar), y los negativos solo sobre el clon desechable que el quickstart crea en su carpeta temporal, que sus prerrequisitos dejan vacía (borran lo que dejara un intento anterior) y su limpieza borra al final; si la tarea se reintenta o se reanuda, la guía vuelve a empezar por los prerrequisitos y no por el escenario en que se quedó, y las comprobaciones de `git status` con el filtro del quickstart, que descarta el directorio del feature en cualquier estado porque el workflow reescribe allí sus ficheros de estado antes de esta tarea (un ` M` de esos ficheros no es un fallo; cualquier línea fuera del directorio del feature sí); y escribir en `specs/005-h4-applet-boe-puerto/gates/pr-h4.md`, fechado por commit y con la plantilla del ritual (objetivo, alcance, controles añadidos, decisiones, pendientes), la salida de `boe-cache-rapida`, la cobertura global (≥ 70 %) y la del dominio (≥ 85 %) sin rebajar ningún umbral, la lista de módulos del binario con su justificación y la constancia de que no entra ninguna dependencia nueva, el resultado de los escenarios negativos, y los supuestos S8 (incidencia nocturna) y S9 (duración en integración continua) como pendientes de la primera ejecución; la tarea no modifica ningún fichero fuera del directorio del feature (FR-124, FR-125, SC-002, SC-014, puntos 1 y 9 de la Definition of Done, obligaciones 9-11 del plan).

- [ ] T037 [plataforma] Publicar la rama del hito y abrir la propuesta de cambio: `git push -u origin h4-applet-boe-puerto`; consultar primero `gh pr view h4-applet-boe-puerto` y, solo si no existe propuesta de cambio (para que un reintento de la tarea no falle), `gh pr create --base main --head h4-applet-boe-puerto --title "feat(H4): applet boe, puerto de boe.py" --body-file specs/005-h4-applet-boe-puerto/gates/pr-h4.md` (sin terminal, `gh pr create` exige `--title` además del cuerpo); esperar y leer `gh pr checks h4-applet-boe-puerto --watch` y `gh run list --branch h4-applet-boe-puerto`, y dejar el estado de la integración continua y de Codecov en `specs/005-h4-applet-boe-puerto/gates/evidencia-plataforma.md`; no fusiona, no empuja a `main`, no fuerza y no crea etiquetas: la fusión es siempre humana (punto 1 de la Definition of Done, §6 del roadmap, ADR 0007).

---

## Definition of Done: qué punto cubre qué tarea

| # | Punto | Dónde se cumple |
|---|---|---|
| 1 | `make ci` en verde (`fmt`, `lint`, `test` con `-race`, `test-integration`, `vuln`, `schema-check`) | Todas las tareas; T029 hace real `schema-check`; T036 lo mide sobre el árbol terminado; T037 lo lee en la integración continua |
| 2 | Tests unitarios offline; fixtures grabados en `testdata/<fuente>/` | Cada tarea de código con su test; manifiesto y grabación en T008 (la graba una persona en su pausa), comprobación en T009, sintéticos en T010; fixtures en `internal/source/boe/testdata/boe.legislacion-consolidada/` (research D12) |
| 3 | Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | `depguard` y `forbidigo` sin cambios en cada `make ci`; T026 (`Arrancar` devuelve el código, `os.Exit` solo en las dos raíces; arquitectura y superficie del binario) y T027 |
| 4 | Salida de applet validada contra `schemas/*.json`; `make schema-check` vigila | T029 (comparador y receta), T032 (`[datos]` esquemas), T033 (validación y cobertura) |
| 5 | Errores tipados con exit code estable; ningún `panic` | T005 (`boe.Error`), T015 (clasificación), T024 (tabla cerrada sobre el error real), T026 (`TestCodigosDeSalidaDeBoe`, `Arrancar`), T027 (el binario de e2e deja el `panic`), T005 (fuzz sin pánico) |
| 6 | Comportamiento visible: e2e (`testscript`) y `CHANGELOG.md` | T028 (cinco guiones y `argumentos.txtar`), T035 (`CHANGELOG.md`) |
| 7 | Decisión de arquitectura: ADR | T001 (`docs/ADR/0015-puerto-source-y-fecha-de-consulta.md`) |
| 8 | Fuente externa: fila de `docs/SOURCES.md` y caso en `scripts/verify-sources.sh` | T007 (fila propuesta y test que la ata), T008 (fila definitiva en la pausa humana), T009 (retira la admisión de `pendiente`), T034 (verificación nocturna) |
| 9 | Cobertura `internal/core/**` ≥ 85 % y global ≥ 70 % | T036 (medida, sin rebajar ningún umbral); T001 no añade sentencias al dominio más allá de las que prueba |
| 10 | Evals de skill | **No aplica**: H4 no entrega ni cambia ninguna skill (spec, *Fuera de alcance*) |
| 11 | Dimensión territorial | **No aplica**: la API es nacional y el applet no filtra por territorio (spec, *Fuera de alcance*) |
| 12 | Operaciones de grafo | **No aplica** hasta H7; los ids naturales quedan en `data` (T013, T018) |

## Trazabilidad: requisito → tarea

| Requisito | Tareas |
|---|---|
| FR-001 (applet `boe`, seis verbos, multicall, sin verbo → 2) | T026, T027, T028 (T025) |
| FR-002 (`fuente`, `url` de la API, espacio reservado vigilado) | T006, T017-T022, T026 |
| FR-003 (solo GET por `httpx`, XML o JSON como `boe.py`) | T002, T006, T015 |
| FR-010, FR-011 (datos de `articulo`, versión vigente) | T012, T013, T018 |
| FR-012 (avisos con `codigo` enumerado y `texto`) | T013, T018, T032, T033 |
| FR-013 (orden bloque → metadatos; nunca sin vigencia) | T018 (T010) |
| FR-014 (ilegible o sin bloque → 4) | T012, T018 (T010) |
| FR-015 (ids naturales en `data`) | T013, T018 |
| FR-016 (sin marcador `?`) | T011, T012, T013 |
| FR-020, FR-021 (`articulos`) | T019 |
| FR-030, FR-031, FR-032 (`buscar`) | T006, T021, T023 |
| FR-040, FR-041 (`indice`) | T005, T011, T020 |
| FR-050, FR-051 (`metadatos`) | T013, T017 |
| FR-060, FR-061 (`analisis`) | T022 |
| FR-070 (objeto suelto → lista) | T011, T020, T021, T022 |
| FR-080, FR-081, FR-082 (ids de entrada y fuzz) | T005, T009 |
| FR-090, FR-091 (caché por verbo, metadatos compartidos, vigencias) | T014, T016, T017, T018, T023, T024 |
| FR-092, FR-093, FR-094 (`--offline`, fallos no guardados, `--dry-run`) | T015, T017-T022, T023, T028 |
| FR-095 (`--timeout` con la clase de `httpx`) | T003, T015 |
| FR-096 (fecha de consulta de la consulta) | T001, T003, T014, T017, T018, T019, T023, T026 |
| FR-100 (códigos estables, nunca 6) | T005, T015, T024, T026, T028 |
| FR-101 (`url` del sobre de fallo) | T015, T017, T018, T019, T026, T028 |
| FR-110 (`--describe`, `schemas/`, `make schema-check`) | T029, T032 |
| FR-111 (validación contra el fichero) | T033 |
| FR-112 (golden) | T016, T030, T031 |
| FR-113 (fixtures grabados en `[datos]`, sin red) | T007, T008, T009 |
| FR-114 (e2e de los seis verbos en reproducción) | T025, T027, T028 |
| FR-115 (`scripts/verify-sources.sh`, nocturno, incidencia) | T034 |
| FR-116 (diff de aceptación contra referencias escritas a mano) | T008, T009, T018 |
| FR-117 (< 200 ms en caché) | T027, T028 |
| FR-120 (`doc.go`, 32 entradas) | T004 |
| FR-121 (`Source` con `Terms`) | T001, T007, T016 |
| FR-122 (ritmo desde `docs/SOURCES.md`; `robust_request` no se porta) | T007, T009, T026 |
| FR-123 (fila de la fuente; parada si los términos lo prohíben) | T007, T008 |
| FR-124 (superficie del binario ampliada y justificada) | T026, T036 |
| FR-125 (sin dependencias nuevas) | Todas (batería); T036 lo deja constatado |
| FR-126 (`CHANGELOG.md`) | T035 |
| FR-127 (reglas de dependencia del adaptador) | T016 y siguientes con `depguard` y `forbidigo`; T026 |
| FR-128 (`--no-graph`, `--asunto` sin efecto) | T026 |
| SC-001 | T008, T009, T018 |
| SC-002 | T027, T028, T036 |
| SC-003 | T019, T023 |
| SC-004 | T023, T028 |
| SC-005 | T008, T009, T028, T030, T031 |
| SC-006 | T029, T032, T033 |
| SC-007 | T005 |
| SC-008 | T003, T005, T015, T019, T024, T026 |
| SC-009 | T034 |
| SC-010 | T004 |
| SC-011 | T026 |
| SC-012 | T010, T018, T023 |
| SC-013 | T013, T018, T026 |
| SC-014 | T035, T036 |

## Obligaciones que el plan trasladó a este fichero

| Obligación del plan | Dónde se cumple |
|---|---|
| 1. Orden y pausas | Manifiesto (T008) con la persona fijando fila, constantes, grabación y referencias; sintéticos (T010) antes de las lecturas (T011-T015); `argumentos.txtar` tolerante (T025) antes del registro en el binario de e2e (T027); guiones e2e (T028) inmediatamente después; golden (T016 → T030 → T031) y esquemas (T029 → T032 → T033) en tres pasos |
| 2. El ejecutor no graba, no consulta la red ni escribe lo que fija la persona | Batería; T009 empieza por los tres tests y se detiene si falta algo o S2 no se cumple |
| 3. Rutas declaradas con precisión | Cada línea; sin llaves ni comodines; desde T009 ninguna ruta completa de lo fijado ni de un directorio que lo contenga (comprobado con la tubería del workflow) |
| 4. `misspell` | T004 (`materias`, `regulares`), T005 (`capitulo`, `disposicion`), T006 (`administrativo`), T016 (`dependencias`); ver Notas |
| 5. Sin `//nolint`, tests saltados ni errores silenciados | Batería; T016 y T029 escriben con `0o600` desde un auxiliar distinto del que lee |
| 6. FR-124 medido al enlazar | T026 |
| 7. ADR 0015 | T001 |
| 8. Documentación alineada | T035 |
| 9. Evidencias en `gates/pr-h4.md` | T026 (módulos), T036 (cierre) |
| 10. Umbrales | T036 |
| 11. `[plataforma]` la última; S8 y S9 pendientes | T036 (anotados), T037 |
| 12. `rtk proxy` | Batería; T036 |

## Dependencias y orden

El orden es estrictamente secuencial: **ninguna tarea depende de una posterior**.

- **T001 → T003**: el puerto y la fecha antes que nada que los use; `Acepta` (T002) e `Instante` (T003) antes de
  pedir (T015). T002 no depende de T001.
- **T004 → T006**: el porte se anota antes de portar (FR-120). `errores.go` va en T005, antes que los validadores que
  devuelven su clase y antes que la búsqueda vacía de T006. `NombreDeLaFuente` nace en T006 porque el arnés de
  grabación (T007) la usa antes de que exista el resto de la fuente (T016).
- **T007 → T008 → T009 → T010**: el arnés y la fila propuesta (T007) tienen que existir para que la persona grabe con
  el ritmo revisado en la pausa de T008; T009 comprueba lo que la persona confirmó; los sintéticos (T010) se copian de
  las grabaciones.
- **T011 → T015**: lecturas antes que los verbos. `datos.go` y `avisos.go` van juntos (T013) porque `Aviso` es un tipo
  de `data`; `peticiones.go` (T015) usa `Acepta`, `Instante` y los errores.
- **T016 → T022**: la fuente antes que los verbos; `metadatos` (T017) antes que `articulo` (T018), que lee su entrada;
  `articulos` (T019) sobre `articulo`. T023 y T024 cubren los seis verbos ya existentes.
- **T025 → T028**: el guion tolerante antes de que el binario de e2e registre `boe`; el registro de producción (T026)
  crea `AppletBoe` y `Arrancar`, que el binario de e2e (T027) usa; los guiones (T028) van justo detrás de T027, la
  primera tarea tras la cual pueden estar en verde, y antes de golden, esquemas y contratos, que no les hacen falta.
- **T029 → T033**: comparador de esquemas (necesita los verbos registrados en producción) → golden `[datos]` (el
  comparador de golden existe desde T016) → cobertura de golden → esquemas `[datos]` → contratos.
- **T034** necesita `bloque.json` (T032) y el sintético ilegible (T010). **T035 → T036 → T037**: documentación, medida
  y, al final, la única tarea que habla con la plataforma.

## Oportunidades de paralelismo

En el workflow todas las tareas van en secuencia. Fuera de él, solo **T002** (junto a T001) y **T004** (junto a T001-T003)
podrían repartirse: no comparten fichero con sus vecinas ni dependen de ellas. El resto comparte `fuente.go`,
`fuente_test.go`, `casos_test.go`, `.golangci.yml` o depende de la tarea anterior.

```text
Tarea: "T001 dominio y kernel: sobre.go del dominio y de la línea de órdenes, source.go, ADR 0015"
Tarea: "T002 Peticion.Acepta en httpx"
Tarea: "T004 doc.go con las 32 entradas de FR-120"
```

## Estrategia de implementación

### MVP (US1 + US2)

1. Fases 1-3: fundación, porte anotado, fila revisada, grabaciones, referencias y sintéticos.
2. Fase 4 y T016-T018: lecturas, fuente, `metadatos` y `articulo`.
3. **Validar**: `TestArticuloCoincideConBoePy` en verde para las cinco referencias (SC-001) y la caché de `articulo`
   con reloj controlado (US2).
4. T023, T026-T028: caché de los seis verbos, binario distribuido, binario de e2e y guiones; `boe-cache-rapida` mide los
   200 ms (SC-002).

### Entrega incremental

- US4 (T019, T020), US3 (T021) y US5 (T017, T022) se añaden verbo a verbo sin romper los anteriores.
- US6 (T029-T033) fija los contratos una vez existen los seis verbos; US7 (T034) vigila la fuente real cada noche.

## Notas

- **Lo que este hito no crea, y no por olvido**: `sumario`, `buscar-materia`, `materias`, `vigilar`, `eli`, grafo,
  `--asunto`, versiones históricas, paginación, revalidación, verbos de mantenimiento de caché, `internal/store`,
  `internal/graph`, `pkg/`, `skills/`, `data/`, evals ni `data/normas.yaml` (spec, *Fuera de alcance*).
- **Ficheros de H0-H3 que se tocan**, exactamente: `internal/core/doc.go`, `internal/core/schema/sobre.go`,
  `internal/cli/sobre.go` (T001); `internal/httpx/peticion.go`, `internal/httpx/cliente.go`, `internal/httpx/robots.go`
  (T002, T003), `internal/httpx/errores.go` (T003); `internal/cache/migraciones.go`, `internal/cache/migraciones_test.go`
  (T003, solo el arreglo del fallo intermitente de la verificación, ver gates/tarea-T003.md); `internal/app/registro.go`, `internal/app/main.go`,
  `internal/app/registro_test.go`, `internal/app/main_test.go`, `internal/app/ayuda_test.go`,
  `internal/app/despacho_test.go`, `cmd/kitlegal/main.go`, `cmd/kitlegal/main_test.go`, `internal/arch_test.go` (T026);
  `internal/app/ejemplo/kitlegal-e2e/main.go`, `internal/app/e2e_test.go` (T027);
  `internal/app/testdata/script/argumentos.txtar` (T025 y T028, solo `[datos]`); `Makefile` (T029, T034);
  `.golangci.yml` (T004, T005, T006, T007, T016, T034); `.github/workflows/nightly.yml` (T034); `docs/PENDIENTES.md`,
  `README.md`, `CONTRIBUTING.md` (T035).
- **`misspell`** (plan, obligación 4; research D18): las seis palabras se añaden a `misspell.ignore-rules`, cada una con
  su comentario, en la tarea que las introduce. Nunca se escriben sueltas sin tilde `adaptacion`, `compilacion`,
  `composicion`, `configuracion`, `conjuncion`, `constitucion`, `construccion`, `convencion`, `correccion`,
  `corrupcion`, `declaracion`, `distribucion`, `documentacion`, `emision`, `evaluacion`, `historicas`,
  `identificacion`, `implementacion`, `informacion`, `justificacion`, `omision`, `orientacion`, `parametros`,
  `presentacion`, `produccion`, `prohibicion`, `proposito`, `proteccion`, `regeneracion`, `transcripcion` ni
  `verificacion`: con tilde en comentarios, cadenas y mensajes, y con el verbo donde no caben tildes (subtests, casos,
  claves). Las palabras correctas que el diccionario marca (`decisiones`, `directos`, `directorios`, `recorre`,
  `clientes`, `comando`, `comandos`, `defectos`, `posicional`, `producto`, `momento`, `contradice`, `distribuye`,
  `resolverse`, `calcular`, `inaccesibles`) se sustituyen como indica D18. Una palabra marcada fuera de la lista se
  reescribe; si la fija un contrato, la tarea se detiene, se anota en `gates/tarea-Tnnn.md` y se redelimita declarando
  `.golangci.yml`.
- **`gosec`**: lecturas de datos de prueba por `filepath.Clean`; escrituras de golden y esquemas con `0o600` desde un
  auxiliar distinto del que lee (G703); ningún `exec` nuevo fuera de `testscript` y del `ejecutaGo` existente.
- **Por qué T026-T027 parten el paso 11 del plan**: el registro de producción obliga en la misma tarea a retirar las dos
  pruebas de superficie y a ampliar `modulosDelBinario` (si no, `make ci` falla), y el binario de e2e necesita
  `AppletBoe` y `Arrancar`; separar el binario de e2e deja los guiones (T028) justo detrás de la primera tarea tras la
  cual pueden estar en verde. El comparador de esquemas, que el plan agrupa en el mismo paso, no lo necesitan los
  guiones y va en T029, al abrir la fase de contratos.
- **Rutas protegidas desde T009**: la fila de la fuente, sus constantes, el manifiesto, las grabaciones y las
  referencias se nombran en las líneas sin su carpeta; los ficheros que cada tarea cambia van por su ruta completa y los
  subdirectorios de datos nuevos también (`internal/source/boe/testdata/sintetico/`, `internal/source/boe/testdata/golden/`).
  Las órdenes con ruta de paquete (regenerar golden y esquemas, quickstart) se citan por su sección del contrato o por su
  escenario, sin copiarlas.

## Comprobación contra la rúbrica del juez (`juez_tasks`, criterios a-g)

| Criterio | Dónde se cumple |
|---|---|
| a. analisis_critico | `gates/analyze.md` lo escribe el paso `analyze`, que sigue a este; las obligaciones 1-12 del plan están asignadas a una tarea concreta (tabla «Obligaciones») |
| b. trazabilidad | Tabla «Trazabilidad» con todos los FR (001-128) y SC (001-014) del spec; cada tarea cita en su línea los requisitos que cumple; ninguna añade nada fuera del spec |
| c. rebanadas_verdes | Cada tarea de código lleva su test y la implementación mínima; las excepciones están declaradas con su razón al principio (`[datos]`, tests sobre lo existente, T027, documentación, cierre, plataforma) y ninguna deja un test en rojo; orden secuencial justificado en «Dependencias y orden»; el e2e de la entrega (T028) va en la primera tarea que puede dejarlo en verde |
| d. rutas_declaradas | Cada línea nombra por su ruta completa los ficheros que crea o cambia, sin llaves, comodines ni rutas genéricas; las tareas que amplían un fichero de una tarea anterior lo declaran (`fuente.go` en T016-T022, `fuente_test.go` en T023 y T031, `casos_test.go` en T016, `ids_test.go` y `terminos_test.go` en T009, `articulo.go` en T019, `metadatos.go` e `indice.go` en T022, `errores_test.go` en T024, `esquemas_test.go` y `boe_test.go` en T033); lo que una tarea solo lee va sin carpeta o por su test; desde T009 ninguna ruta extraída coincide con lo que fija la persona ni con un directorio que lo contenga |
| e. datos_separados | Las seis `[datos]` (T008, T010, T025, T028, T030, T032) solo tocan `testdata/` o `schemas/` y ninguna línea de código contiene esas carpetas; T037 es la única `[plataforma]`, va la última y solo publica y lee estados |
| f. dod | Tabla «Definition of Done» con los doce puntos: nueve aplicables con su tarea (ADR en T001, `SOURCES.md` en T007-T009, `CHANGELOG.md` en T035, esquemas en T029-T033, fixtures en T008-T010) y tres no aplicables con su razón |
| g. checklist_veraz | `checklists/requirements.md` no lo modifica ninguna tarea |
