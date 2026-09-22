# Tasks: H6 · `territorio` + skill `legal-core` v0

**Input**: `specs/008-h6-territorio-skill-legal/` (spec.md, plan.md, research.md, data-model.md, contracts/, quickstart.md)

**Prerrequisitos**: spec.md y plan.md leídos; `.specify/memory/constitution.md` aplicada; H0-H5.1 cerrados (kernel de la
CLI con las ocho banderas globales, el sobre de seis claves y los exit codes estables; `internal/httpx`;
`internal/cache`; el applet `boe` con sus esquemas publicados; `internal/skills` con `ValidarDocumentoYAML`,
`CompilarEsquema`, `TestSkillsDelRepositorio` y `TestNormasDelRepositorio`; `internal/evals` con `LeerEval`, `Juzgar`,
`ComprobarConjuntoDeBoeLegislacion`, `EscribirInforme` y `TestEvalsDelRepositorio`; las 18 evals de
`evals/boe-legislacion/`; `Makefile` con `ci`, `test`, `schema-check`, `skills-check` y `skills-sync`; ADR 0016 y
ADR 0017). **Ninguna tarea de H6 crea el esqueleto ni los gates: `make ci` existe y pasa desde H0**, y cada tarea lo
deja en verde al terminar.

**Tests**: obligatorios (constitución §III; spec, FR-092 a FR-097; «Controles mecánicos» de plan.md). Cada tarea escribe
su test antes del código que lo hace pasar (rojo → verde); los nombres de test y de subtest son los del «Inventario de
tests» de plan.md y los de los contratos.

**Modo**: desatendido. Las tareas se ejecutan y verifican **una a una** por el workflow `hito`, con guardián de diff por
tarea. Cada tarea es una **rebanada vertical** que deja `make ci` en verde y **declara en su propia línea todas las
rutas** que va a crear o modificar, y **solo** esas (plan, «Obligaciones», punto 3).

## Formato: `[ID] [P?] [etiquetas] [Story] Descripción con rutas`

- **[P]**: la tarea no comparte fichero con sus vecinas ni depende de ellas (en el bucle del workflow van igual en
  secuencia).
- **[Story]**: historia del spec — US1 el municipio del territorio configurado con su terreno completo · US2 el mismo
  comando fuera del territorio configurado sin inventar nada · US3 el régimen foral marcado · US4 nombre ambiguo,
  inexistente y código mal formado, con sus identificadores · US5 la skill `legal-core` que empieza por el territorio y
  lo demuestra en sus evals · US6 los identificadores de las leyes vertebrales verificados, no recordados · US7 la
  correspondencia INE→DIR3 generada, verificada y congelada fuera de la ejecución. Las tareas de documentación, cierre y
  plataforma no llevan historia.
- **[datos]**: la tarea toca material bajo `testdata/` o `schemas/`, o el material congelado que la pausa humana revisa
  con él. Son **doce**: las diez del plan (obligación 1) —T001, T002, T003, T005, T010, T012, T013, T014, T015 y
  T018—; T027, que entró al redelimitar T007 en su primer intento: los patrones de identificador que T001 y T003
  fijaron en tres esquemas no son las gramáticas de los analizadores, que es lo que el subtest `gramaticas` de T007
  exige, y cambiarlos es material de esquema (nota de T007 en `gates/`); y T028, que entró al redelimitar T022 en su
  primer intento: el guion de reinstalación de `make install` exige que el directorio personal de skills tenga solo
  `boe-legislacion`, en cuanto `legal-core` existe la instalación la enlaza también, y cambiar el guion es material de
  test existente (nota de T022 en `gates/`). Ninguna mezcla trabajo ajeno a su material y **solo dos tocan código**,
  por la razón escrita en cada una y en *Complexity Tracking* de plan.md: T010 (registrar el applet, sin lo cual el
  esquema no se puede generar, D16) y T018 (la expectativa de test que la forma del esquema de eval impone, D28). T003
  lleva además la fila de `docs/SOURCES.md` y el registro de la muestra porque FR-049 y FR-047 los ponen dentro de esa
  misma pausa. **Diez provocan pausa humana** —T001, T002, T003, T027, T010, T013, T014, T015, T018 y T028: material
  existente modificado o esquema nuevo—, y **T005 y T012 no**, porque son ficheros nuevos bajo el `testdata/` de un
  paquete, fuera del territorio de fixtures; a esos dos los revisa la revisión final del hito (FR-086).
- **[plataforma]**: la tarea necesita la plataforma remota. Es **una**, la última: T026 (publicar la rama, abrir la
  propuesta de cambio y leer y registrar la ejecución de aceptación). Fusionar nunca es del workflow.

## Batería de verificación por tarea

- **`make ci`** en cada tarea, **en primer plano** y con la tarea marcada `[X]` en el mismo turno en que termina en
  verde: formato · lint (`depguard`, `forbidigo`, `gosec`, `misspell`, `dupl`, `paralleltest`, `revive`…) · tests con
  `-race`, e2e incluido · `test-integration` · vulnerabilidades · `schema-check` · `skills-check` · secretos ·
  integridad y `tidy -diff` de módulos. Ningún registro de la verificación se escribe en la raíz del repositorio.
- **Cero red en todo el bucle** (FR-043, FR-093, ADR 0017): todo test es offline —el applet no tiene red por
  construcción, el dominio se prueba con fuentes sintéticas y el repositorio con sus ficheros congelados—. **Ninguna
  tarea ejecuta `KITLEGAL_RECORD`, el guion de grabación, `make evals`, `make verify-sources`, `gh` ni ninguna
  descarga, salvo T026, que solo usa `gh` y `git push`**; la relación del INE, el volcado del REL y las grabaciones de
  las búsquedas las obtiene una persona en las pausas de T001, T003 y T015, fuera del repositorio. En la pausa de T002
  **no se descarga nada**: lo que aporta la persona es el nombre, el régimen y —solo en la Comunidad de Madrid— los
  boletines de cada comunidad, y lo demás se copia de `municipios.yaml`, ya congelado.
- **Nada escrito de memoria** (plan, obligación 2): ningún identificador `BOE-A-…`, título, rango, código INE, dígito de
  control ni DIR3 se escribe sin haberlo leído de una respuesta grabada o de un fichero congelado ya revisado. Si falta,
  la tarea se detiene sin marcarse y lo anota en `gates/tarea-Tnnn.md`.
- **Sin cambios fuera del alcance** (plan, obligación 6): ninguna tarea declara ficheros de `internal/httpx`,
  `internal/cache`, `internal/source/boe`, `internal/cli` ni `internal/core/schema`; `skills/boe-legislacion/SKILL.md`
  no se toca; `skills/boe-legislacion/references/normas.md` cambia **solo por regeneración** y solo en T017; ninguna
  eval de `boe-legislacion` se modifica.
- **Genericidad territorial** (FR-024, principio IX): ningún municipio ni comunidad aparece como caso especial en código
  ni en los datos; los municipios concretos solo viven en fixtures, en el guion e2e y en las evals.
- **`misspell`** (plan, obligación 4): en Go no se escriben sueltas y sin tilde `aspectos`, `configuracion` ni
  `autonomos`; se usan `aspecto` en singular, `configuración` y `autónomas`. No se añade ninguna entrada a
  `ignore-rules` salvo que la implementación demuestre que necesita una literal, y entonces con su comentario de motivo
  y declarando `.golangci.yml` en la tarea.
- **Sin atajos** (plan, obligación 5): ni `//nolint`, ni linter desactivado, ni exclusión nueva, ni test saltado, ni
  error silenciado, ni umbral rebajado, ni `panic` en rutas de usuario. Si un test queda legítimamente en rojo porque su
  implementación pertenece a otra tarea, la tarea está mal delimitada: se anota en `gates/tarea-Tnnn.md`, se redelimita
  y se deja sin marcar.
- **`rtk`** (plan, obligación 11): toda orden cuya salida se filtre, se cuente o se compare (`go test … | grep`,
  `git status --porcelain`, `git diff`, `make`) se ejecuta con `rtk proxy`, cada etapa de la tubería incluida.

**Rebanadas verticales y excepciones declaradas.** Son rebanadas completas —test antes que código, `make ci` en verde
por sí solas— T004, T006, T007, T008, T009, T011, T016, T017, T019, T020, T021 y T022. Las demás tampoco dejan nada a
medias, cada una por su razón: **T001, T002, T003, T005, T013, T014 y T027** son `[datos]` y solo traen material
congelado, esquemas o corpus que ningún lector mira todavía, de modo que `make ci` sigue en verde y el control que los
valida llega en la tarea que lo puede dejar en verde (T007 para los ficheros de territorio, los patrones de sus esquemas
y el corpus del pliegue, T004 y T005 para el fuzz, T016 para la jerarquía y T017 para las marcas `vertebral`); **T010, T012, T015 y T018** son `[datos]`
indivisibles, con la pieza de código o de test que un control existente ata a su material (D16, D28, D29) o con el guion
e2e que ejerce lo ya registrado; **T028** es `[datos]` porque cambia un guion de test existente, y deja `make ci` en
verde por sí sola, con una sola skill y después con dos; **T023** es el job, que ningún `make ci` ejecuta; **T024** es documentación; **T025** es
validación sobre el árbol terminado, cuyo único fichero escrito está en el directorio del hito; **T026** solo publica,
lee y registra.

**Test de extremo a extremo de la entrega.** Es `internal/app/testdata/script/territorio-matriz.txtar`, la matriz
territorial completa (cubierto, no cubierto, foral, ambiguo), y entra en **T012**, la primera tarea que puede dejarlo en
verde: el guion invoca el binario de e2e, y hasta T010 el applet no está registrado en él ni su contrato publicado.
Antes de T012 la entrega está cubierta por los tests de `internal/app` de T009 y T011, que ejercen el mismo verbo desde
dentro del proceso.

---

## Phase 1: Los datos congelados del territorio (fundación; US7, US1)

**Objetivo**: que exista, versionado y validado por su esquema, el terreno del que sale toda respuesta del applet:
municipios del INE, configuración por comunidad y correspondencia INE→DIR3 verificada. Todo lo genera una persona en la
pausa, con las descargas hechas fuera del repositorio.

- [X] T001 [datos] [US7] Relación congelada de municipios y su esquema: `schemas/territorio-municipios.yaml.json` (nuevo, JSON Schema 2020-12 con `additionalProperties: false`, raíz con `fecha` en `AAAA-MM-DD` y `source`, y mapa `municipios` con `propertyNames.pattern` `^[0-9]{5}$` y cada fila con sus cinco columnas obligatorias: `dc` de una cifra, `nombre` no vacío, `provincia` y `comunidad` de dos cifras) y `data/territorio/municipios.yaml` (una línea por municipio, `fecha` la de referencia del fichero del INE y `source` `ine.municipios`), escritos por una persona en la pausa a partir de la relación oficial del INE descargada **fuera del repositorio**; el ejecutor desatendido no descarga nada y no escribe ningún código INE ni dígito de memoria —si el material no está disponible se detiene sin marcarse y lo anota en `gates/tarea-T001.md`—; no toca código y `make ci` sigue en verde porque todavía ningún lector mira el fichero (FR-040, FR-041, FR-044, FR-045, FR-086, US7 escenario 1, SC-007, data-model §3.1, contrato de datos §1 y §4).

- [X] T002 [datos] [US1] Configuración por comunidad y boletín estatal: `schemas/territorio-comunidad.yaml.json` y `schemas/territorio-estado.yaml.json` (nuevos, con la forma de data-model §3.3 y §3.4 y `additionalProperties: false`: `fecha` y `source` en la raíz; en la comunidad, `codigo` de dos cifras, `nombre`, `regimen` obligatorio en el enumerado `comun`\|`foral`, `provincias` con al menos una entrada y `boletines` **opcional** con `autonomico` y `provincial`, cada uno con `codigo`, `nombre`, `url` y `motivo`), `data/territorio/estado.yaml` con el boletín estatal y `data/territorio/comunidades/` con los 19 ficheros de comunidad y ciudad autónoma, cada uno con el código que da nombre a su fichero y con su régimen, y **solo el de la Comunidad de Madrid con `boletines`**, donde el BOCM figura a la vez como autonómico y como provincial con el motivo de la equivalencia escrito en el propio fichero; ninguna otra comunidad trae boletines, ninguna comunidad uniprovincial distinta deduce nada, y añadir un territorio será rellenar su fichero; **nada se escribe de memoria y esta tarea no descarga nada**: el código de cada comunidad y las provincias que declara se copian de las columnas `comunidad` y `provincia` de `municipios.yaml`, que T001 dejó congelado, y el nombre y el régimen de cada comunidad, y el código, el nombre, la dirección y el motivo del BOCM, los aporta y los revisa una persona en la pausa, que comprueba además que las provincias de los 19 ficheros cubren, sin repetirse, todas las que aparecen en la relación, porque es lo que `Cargar` exigirá en T006 (data-model §2.1); si ese material no está disponible, la tarea se detiene sin marcarse y lo anota en `gates/tarea-T002.md`; no toca código y `make ci` sigue en verde (FR-044, FR-045, FR-050, FR-051, FR-052, FR-053, FR-055, FR-086, US1 escenario 2, US3, SC-003, SC-005, data-model §3.3 y §3.4).

- [X] T003 [datos] [US7] Correspondencia verificada INE→DIR3, su fuente y el registro de la muestra: `schemas/territorio-dir3.yaml.json` (`fecha`, `source`, mapa `correspondencia` con `propertyNames.pattern` `^[0-9]{5}$` y valor `^L01[0-9]{6}$`, `additionalProperties: false`), `data/territorio/dir3.yaml` con **solo las filas verificadas** —entra todo municipio con fila en el REL cuyo número de inscripción es coherente con su código INE y su dígito de control oficial; lo que no, queda fuera— con `fecha` la del volcado del REL y `source` `mpt.rel`, la fila `mpt.rel` de `docs/SOURCES.md` con esa fecha —**sin añadir ningún caso a** `verify-sources.sh`, porque estas fuentes no se piden en red— y `specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md` con una fila por municipio de la muestra (municipio, código INE, DIR3 derivado, DIR3 real, de dónde salió el real y si coincide) que incluye al menos un municipio fusionado o renombrado, uno de régimen foral y uno con entidades locales menores, y con la conclusión de si la derivación queda como regla o pasa a tabla; todo lo escribe y lo decide una persona en la pausa, el ejecutor no descarga, no deriva y no inventa ningún código, y `make ci` sigue en verde (FR-042, FR-043, FR-044, FR-045, FR-046, FR-047, FR-048, FR-049, FR-086, US7 escenarios 1, 2 y 3, SC-007, SC-008, SC-014, contrato de datos §4 y §5).

**Checkpoint**: el terreno está congelado, versionado y descrito por cuatro esquemas; nada de ello se pide por red y la
evidencia de la muestra está registrada.

---

## Phase 2: Los identificadores del proyecto (US4)

**Objetivo**: que código INE y DIR3 sean tipos con gramática, dígito de control comparado y errores tipados, probados
con tabla y con fuzz, antes de que nadie resuelva un municipio.

- [X] T004 [US4] Código INE y DIR3 en el dominio, con sus tests primero: `internal/core/ids/doc.go` (comentario de paquete, que `revive` exige), `internal/core/ids/ine.go` (`CodigoINE` inmutable que solo se construye analizando, `AnalizarCodigoINE`, `AnalizarCodigoINEConDigito`, `String()` de cinco cifras con sus ceros, `Provincia()` y el **método** `ComprobarDigito(declarado, oficial byte) error`, que compara y nombra los dos dígitos; cifras comprobadas byte a byte sin `regexp`, provincia `01`-`52` y municipio `001`-`999`), `internal/core/ids/dir3.go` (`DIR3` con la forma `L01PPMMMD`, `AnalizarDIR3` que acepta la letra en mayúscula y en minúscula y la normaliza a mayúscula, `DIR3DeAyuntamiento`, `String()`, `CodigoINE()` y `Digito()`) e `internal/core/ids/errores.go` (tipo de error propio con `Clase()` igual a `schema.ClaseArgumentos`, que nombra la entrada con `%q` y dice qué tiene de malo, nunca `panic` ni valor por omisión), con `TestAnalizarCodigoINE`, `TestComprobarDigito`, `TestIdaYVuelta`, `TestAnalizarDIR3`, `TestDIR3DeAyuntamiento`, `TestClaseDeLosErrores`, `TestSuperficieDeIds` (lo exportado es exactamente la lista del contrato §1: ni ELI, ni ECLI, ni CELEX, ni NIF) y los objetivos `FuzzCodigoINE` y `FuzzCodigoDIR3`, cada uno en el fichero de test de su analizador y con las semillas `F.Add` del contrato §6, escritos antes que el código; dominio puro y sin entrada ni salida ni siquiera en los `_test.go`, y los dos analizadores en ficheros distintos compartiendo la comprobación de cifras para no disparar `dupl` (FR-030, FR-031, FR-032, FR-033, FR-034, FR-035, FR-096, FR-097, US4 escenario 3, SC-006, contrato de identificadores §1 a §5).

- [X] T005 [datos] [US4] Corpus semilla versionado del fuzz: `internal/core/ids/testdata/fuzz/FuzzCodigoINE/` e `internal/core/ids/testdata/fuzz/FuzzCodigoDIR3/` con las entradas mínimas del contrato de identificadores §6 (cinco y seis cifras, ceros por delante, provincia fuera de rango, municipio `000`, cadena vacía, caracteres que no son cifras ASCII, longitudes imposibles, y el DIR3 en mayúscula, en minúscula, corto y con letra ajena), que `make ci` ejecuta **sin** `-fuzz`, como corpus semilla; el ejecutor no lanza ningún fuzz largo ni abre la red, no toca código y `make ci` queda en verde (FR-035, FR-086, SC-006, contrato de identificadores §6).

**Checkpoint**: los dos identificadores aceptan y rechazan según su gramática y su dígito, tienen ida y vuelta estable y
no entran en `panic` con el corpus versionado.

---

## Phase 3: El dominio del territorio y sus datos dentro del binario (US1 a US4, US7)

**Objetivo**: que exista la resolución completa —carga con integridad, índices, pliegue de nombres, códigos de salida,
salida con cobertura— sobre fuentes sintéticas, y que los ficheros congelados viajen dentro del binario y se validen en
`make ci`.

- [X] T006 [US1] Dominio del territorio con tests sintéticos, uno por fichero: `internal/core/territorio/doc.go`; `internal/core/territorio/fuentes.go` con `Fuentes` (cuatro campos de bytes), `Ficheros` y los tipos `FicheroDeMunicipios`, `FicheroDeDIR3`, `FicheroDeEstado` y `FicheroDeComunidad` **exportados**, y `Cargar(Fuentes) (*Registro, error)`, que decodifica, comprueba los seis puntos de integridad de data-model §2.1 —toda provincia declarada por una sola comunidad, todo municipio de la correspondencia existente y con DIR3 coherente con su código y su dígito, ningún código repetido, régimen `comun` o `foral`, nombre de fichero igual al código que declara dentro y **la `comunidad` de cada municipio igual a la de la comunidad que declara su provincia**— y construye los índices; `internal/core/territorio/nombres.go` con `Plegar` exportado (minúsculas, diacríticos a su letra base, separadores y espacios colapsados) y las formas conocidas derivadas **del nombre oficial**, sin caso especial por municipio: la del INE, la del artículo pospuesto antepuesto y cada lado de un nombre bilingüe; `internal/core/territorio/registro.go` con los índices por código y por forma plegada y los candidatos ordenados por código INE; `internal/core/territorio/resolver.go` con `Resolver` (solo cifras → código de cinco o seis cifras, si no, nombre; primero «¿está en la relación?» y después el dígito); `internal/core/territorio/salida.go` con el territorio resuelto de data-model §2.5 (ocho claves, `source` en cada dato, `boletines` con siempre el estatal y solo los niveles configurados, `cobertura` de tres claves con vocabulario cerrado y sin ningún valor que signifique «no existe», la invariante del DIR3 no verificado y la fecha más antigua de los ficheros que sostienen la respuesta); e `internal/core/territorio/errores.go` (solo `schema.ClaseArgumentos` y `schema.ClaseNoEncontrado`, cada mensaje nombrando la entrada); con `TestCargar`, `TestRegistro`, `TestPlegar`, `TestFormasDelNombre`, `TestPlegarEsIdempotente`, `TestResolver`, `TestTerritorioResuelto`, `TestFechaMasAntigua` y `TestClaseDeLosErroresDeTerritorio` escritos antes y **enteros sobre fuentes sintéticas**: ningún test del dominio abre un fichero (FR-006, FR-008, FR-010 a FR-015, FR-020 a FR-024, FR-054, FR-055, FR-096, FR-097, US1 a US4, SC-002, SC-003, SC-008, data-model §2.1 a §2.8).

- [X] T027 [datos] [US7] Los patrones de identificador de los esquemas de territorio, iguales a las gramáticas de los analizadores, antes del control que los ata: en `schemas/territorio-municipios.yaml.json` (la clave de `municipios` y la `provincia` de cada fila), `schemas/territorio-dir3.yaml.json` (la clave y el valor de `correspondencia`) y `schemas/territorio-comunidad.yaml.json` (la clave de `provincias`), cada `pattern` que describe un código INE, su provincia o un DIR3 pasa a aceptar exactamente lo que aceptan `AnalizarCodigoINE` y `AnalizarDIR3` —el código INE, `^(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})$`, con la provincia de `01` a `52` y el municipio de `001` a `999`; la provincia sola, `^(0[1-9]|[1-4][0-9]|5[0-2])$`; y el DIR3, `^[Ll]01(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})[0-9]$`, con la letra en mayúscula o en minúscula, como el analizador—, sin tocar ningún otro patrón ni ninguna otra propiedad, y data-model §3.1 y §3.2, que describen esos patrones, pasan a decir lo mismo; es tarea de datos aparte porque el subtest `gramaticas` de T007 exige esa igualdad (contrato de identificadores §2) y los patrones que fijaron T001 y T003 —`^[0-9]{5}$`, `^[0-9]{2}$` y `^L01[0-9]{6}$`— aceptan la provincia `00` o mayor que `52` y el municipio `000` y rechazan la `l` minúscula, mientras T007 no lleva la etiqueta que deja tocar esquemas; la pausa revisa los tres esquemas juntos; no toca código ni datos: los ficheros congelados siguen validando —todos sus códigos están en rango, lo que T007 comprueba— y `make ci` sigue en verde porque hasta T007 ningún lector mira estos esquemas (FR-030, FR-031, FR-044, FR-086, contrato de identificadores §2, contrato de datos §2).

- [X] T007 [US1] Los datos congelados dentro del binario y su validación en `make ci`: `data/datos.go` (paquete `data` con su comentario de paquete y **solo** directivas `//go:embed` para los tres ficheros sueltos y el subárbol de comunidades, más `Comunidades() (map[string][]byte, error)` que devuelve un mapa por código; sin más lógica, y un patrón que no case es error de compilación, no modo de fallo en ejecución) e `internal/skills/territorio.go` con `LeerTerritorio`, que valida los cuatro documentos contra sus esquemas con el lector común —clave repetida incluida— usando los tipos que exporta el dominio, sin repetirlos, e `internal/skills/export_test.go`, que expone a los tests del paquete externo el compilador de esos esquemas, como ya hace con el de normas, para fijar sus errores de lectura y de compilación; su test lleva `TestLeerTerritorio` sobre documentos sintéticos (válido, clave desconocida, patrón, clave repetida), `TestCompilarEsquemaDelTerritorioDesdeUnaRuta` y `TestTerritorioDelRepositorio` con los subtests `esquema`, `integridad` (delegando en `Cargar`), `fuentes` (todo `source` emitido es un identificador de fila de `SOURCES.md` o la ruta de un fichero congelado que existe), `solo-madrid-configurada`, `regimen-de-todas`, `gramaticas` (los `pattern` de los tres esquemas, que T027 deja iguales a sus gramáticas, aceptan y rechazan exactamente lo mismo que `AnalizarCodigoINE` y `AnalizarDIR3`), `pliegue-cubre-el-corpus`, `nombres-alcanzables` y `ningun-nombre-es-solo-cifras`, los tres últimos sobre el corpus congelado real, y **todos ellos exigen que existan los cuatro ficheros y las 19 comunidades**, para no pasar en vacío; y `Makefile`, cuya expresión `-run` de `make skills-check` gana `TestTerritorioDelRepositorio` (FR-024, FR-044, FR-056, FR-097, US7 escenario 4, SC-005, SC-007, SC-012, contrato de datos §2 y §3).

**Checkpoint**: el territorio se resuelve entero en memoria desde bytes, y los ficheros reales viajan en el binario y se
validan, con su integridad y su pliegue, dentro de `make ci`.

---

## Phase 4: El applet, su contrato publicado y su activación (US1, US2)

**Objetivo**: que `territorio resolver` exista, publique su contrato sin deriva y responda con los códigos de salida
estables, sin inventar ningún boletín.

- [X] T008 [US1] Tabla de esquemas publicados parametrizada por applet, **sin añadir ninguna fila**: `internal/app/esquemas_test.go`, donde `ficherosDeEsquemas` gana la columna del applet y los nueve puntos que hoy escriben `boe` literal pasan a leerla, de modo que el título `<applet> · <entidad>` y el nombre de la parte publicada se compongan de la tabla en vez de estar cableados; `TestEsquemasPublicados` y `TestEsquemasCubrenTodosLosVerbos` siguen en verde sobre `norma.json` y `bloque.json` sin regenerar nada y `make schema-check` no cambia de resultado (FR-007, FR-092, SC-012, contrato del applet §6, plan *Complexity Tracking*).

- [X] T009 [US1] El applet `territorio` con su verbo `resolver`, probado sobre un registro local del test y **sin registrarlo todavía** en producción: `internal/app/territorio.go` con `AppletTerritorio(fuentes territorio.Fuentes) Applet` —único argumento posicional obligatorio `Consulta`, ninguna bandera propia (hereda las ocho globales), análisis del registro una sola vez con `sync.OnceValues` capturado en el valor del applet, de modo que `--help` y `--describe` no lo analicen, procedencia del espacio reservado (`kitlegal.territorio` y `kitlegal:applet/territorio`) con la fecha más antigua de los ficheros que sostienen la respuesta, y unas fuentes que no analizan tratadas como defecto de composición, nunca como código de usuario—, con `TestResolverDevuelveElTerritorio` (las ocho claves de `data`, ni una más ni una menos, sin `omitempty`, el `source` de cada dato, los boletines del municipio cubierto y los del no cubierto, `cobertura` con sus tres claves y la invariante del DIR3 no verificado) y `TestCodigosDeTerritorio` (`ambiguo`, cuyo mensaje lleva **todos** los candidatos en la forma `<código INE> <nombre> (<provincia>)` separados por `; ` y ordenados por código INE, `no-encontrado` —por nombre y por un código **bien formado**: provincia dentro de `01`-`52` y municipio ausente del registro del test—, `codigo-mal-formado` —cifras que no forman código, provincia `00` o mayor que `52` y municipio `000`, todos con clase `argumentos` y **nunca** con 3—, `digito-incorrecto`, `sin-argumento`, y que **ninguna** invocación devuelve 4, 5 ni 6) escritos antes que el código; y en la misma tarea `internal/arch_test.go`, donde la lista `modulosDelBinario` gana `go.yaml.in/yaml/v3` **con su motivo**, porque el cierre de dependencias del binario recorre el paquete de composición entero y el módulo queda enlazado por la mera existencia de este fichero, esté o no registrado el applet (FR-001 a FR-006, FR-008 a FR-016, FR-021, FR-022, FR-023, FR-096, US1 a US4, SC-001 a SC-004, contrato del applet §1 a §5 y §7).

- [X] T010 [datos] [US1] Contrato publicado del applet y su activación, indivisibles: `schemas/municipio.json` generado con la receta existente (`TestEsquemasPublicados` con `-actualizar-esquemas`; **no se escribe a mano**) con su `$id` y su parte `$defs.resolver` en la forma canónica, la fila nueva de la tabla de `internal/app/esquemas_test.go` (entidad `municipio`, verbo `resolver`), el registro del applet en `internal/app/registro.go` y en `internal/app/ejemplo/kitlegal-e2e/main.go` —los dos con las fuentes embebidas, que no dependen del entorno— y los cuatro sitios que fijan la lista de applets: `internal/app/registro_test.go`, `cmd/kitlegal/main_test.go`, `internal/app/testdata/script/argumentos.txtar` e `internal/app/testdata/script/ayuda.txtar`; nada más, y el código que toca es exactamente el que un control existente hace inseparable del fichero protegido, porque en cuanto el verbo está en el registro `TestEsquemasCubrenTodosLosVerbos` exige su parte publicada y el esquema se genera **desde** el applet ya registrado: separarlos dejaría `make ci` en rojo entre dos tareas; la pausa revisa juntos el contrato publicado y los dos guiones existentes (FR-007, FR-086, FR-092, SC-012, contrato del applet §6, plan *Complexity Tracking* D16).

- [X] T011 [US2] La salida real contra el contrato ya publicado: `internal/app/territorio_test.go` gana `TestSalidaDeTerritorioContraSchemas` (la salida del verbo, para un municipio cubierto y para uno no cubierto, valida contra el contrato publicado `municipio.json` compilado del fichero real) y `TestSalidaSinBoletinesNoConfigurados` (en la salida del municipio no cubierto no aparece el código, ni el nombre, ni la dirección de ningún boletín no configurado, `boletines` trae solo el estatal y `cobertura` declara `no-configurado` el autonómico y el provincial, sin ningún valor que permita concluir que no existan); no toca ningún otro fichero (FR-020, FR-021, FR-022, FR-092, US2 escenarios 2, 3 y 4, SC-002, Definition of Done §1.4).

**Checkpoint**: `territorio resolver` responde con su sobre, su `data` de ocho claves y sus códigos 0, 2 y 3, publica su
contrato sin deriva y no nombra nunca un boletín que no esté configurado.

---

## Phase 5: La matriz territorial de extremo a extremo (US1 a US4)

**Objetivo**: que la entrega del hito esté descrita por un guion `testscript` con los cuatro casos del control, el
determinismo y el tiempo de respuesta.

- [X] T012 [datos] [US1] Matriz territorial en el e2e, el guion que describe la entrega: `internal/app/testdata/script/territorio-matriz.txtar` con Leganés (cubierto: exit 0, las ocho claves y la cobertura configurada), Tordesillas (datos nacionales completos, los dos boletines declarados no cubiertos y **ninguna** aparición del código ni del nombre de un boletín no configurado), un municipio de Navarra o del País Vasco (régimen foral marcado aunque su comunidad no esté configurada), un nombre ambiguo (exit 2 con todos sus candidatos), un código **bien formado** que no está en la relación (exit 3) —provincia de las que trae el fichero congelado y municipio que ese fichero no tiene, tomado de él y comprobado contra él, nunca escrito de memoria— y un código con la provincia fuera de `01`-`52`, que **no** llega a ser un código y da exit 2 con clase `argumentos`, nunca 3; la igualdad byte a byte entre la consulta por nombre, por código y con `--offline` (`cmp`), la misma salida ejecutando el binario **desde otro directorio de trabajo**, `--describe` **con su argumento posicional** —sin él la invocación termina en 2— y la orden `cronometra` por debajo de 200 ms; no toca código y `make ci` lo ejecuta dentro de `test` (FR-009, FR-056, FR-086, FR-090, FR-091, US1 a US4, SC-001 a SC-004, contrato del applet §8).

**Checkpoint**: la entrega del hito está comprobada de extremo a extremo, con sus cuatro casos territoriales, su
determinismo y su tiempo.

---

## Phase 6: Leyes vertebrales y jerarquía normativa (US6, US5)

**Objetivo**: que las quince leyes vertebrales estén en los datos con su identificador **verificado contra una búsqueda
grabada**, que la jerarquía exista como dato y tenga lector y control propios, y que las referencias se generen desde
una tabla de generadores.

- [X] T013 [datos] [US6] Campo `vertebral` en el esquema de normas, **solo la propiedad nueva**: `schemas/normas.yaml.json` gana la propiedad booleana **opcional** `vertebral` en cada norma, compatible hacia atrás (ninguna norma ya válida deja de serlo) y sin tocar el `enum` de `rango`, que `TestEsquemaDeNormas/rangos-grabados` compara por igualdad con el conjunto exacto de rangos de las búsquedas grabadas y que por tanto solo puede crecer en T015, con ellas; ninguna norma trae todavía la marca y el lector decodifica sin `KnownFields`, así que no toca código y `make ci` queda en verde (FR-067, FR-070, FR-085, US6, SC-010, data-model §4.1).

- [X] T014 [datos] [US5] La jerarquía normativa como dato: `schemas/jerarquia.yaml.json` (nuevo: `niveles` con `nivel` en un enumerado cerrado de cinco valores en orden —`ue`, `estado`, `comunidad-autonoma`, `provincia`, `municipio`—, `nombre`, `boletin` y `normas`, y `reglas` con `regla` en un enumerado de las cuatro reglas de interpretación y su `enunciado`; `additionalProperties: false`) y `data/jerarquia.yaml` con los cinco niveles, el boletín y los tipos de norma de cada uno, y las cuatro reglas (competencia antes que jerarquía, ley posterior, ley especial y reglamento nunca contra ley); hasta T016 no hay lector ni control que lo mire, así que no toca código y `make ci` sigue en verde (FR-044, FR-066, FR-067, FR-086, US5, SC-009, data-model §4.2).

- [X] T015 [datos] [US6] Manifiesto y grabaciones de las siete búsquedas nuevas, con el `enum` de `rango` que fijan: `testdata/evals/` gana las siete entradas del manifiesto de grabación y las respuestas grabadas de `boe buscar` que resuelven las siete leyes vertebrales que faltan, **grabadas por una persona en la pausa** (el ejecutor no usa la red, ni la variable de grabación, ni el guion de grabar), y en la misma tarea `schemas/normas.yaml.json` gana en su `enum` de `rango` los valores que esas grabaciones introducen —al menos «Ley Orgánica»—, **copiados de la respuesta grabada y nunca de memoria**; es indivisible porque `TestEsquemaDeNormas/rangos-grabados` compara ese enumerado por igualdad con el conjunto exacto de rangos grabados, de modo que cualquiera de los dos órdenes por separado deja `make ci` en rojo; la pausa revisa las dos piezas juntas y comprueba que cada búsqueda resuelve su norma, y si alguna no lo hace se rechaza y se anota (FR-071, FR-072, FR-073, FR-085, FR-086, US6 escenario 2, SC-010, plan *Complexity Tracking* D29).

- [X] T016 [US5] El lector de la jerarquía normativa y su control, antes que el generador que la usa: `internal/skills/jerarquia.go`, lector nuevo del fichero `jerarquia.yaml` que lo valida contra su esquema con el lector común —clave repetida incluida— y devuelve los niveles en el orden del enumerado y las cuatro reglas, con `TestLeerJerarquia` sobre documentos sintéticos (válido, nivel fuera del enumerado, regla desconocida, clave desconocida, clave repetida), `TestEsquemaDeJerarquia` y `TestJerarquiaDelRepositorio` (el fichero real: los cinco niveles en su orden y las cuatro reglas, **exigiendo que existan** para no pasar en vacío) escritos antes que el código; y `Makefile`, cuya expresión `-run` de `make skills-check` gana `TestJerarquiaDelRepositorio`; no toca datos, normas ni generadores, y `make ci` queda en verde por sí sola (FR-044, FR-066, FR-067, US5, SC-009, data-model §4.2, contrato de la skill §4).

- [X] T017 [US6] Las quince leyes vertebrales y la tabla de generadores de referencias: `data/normas.yaml` con las siete normas nuevas —identificador `BOE-A-…`, título y rango **copiados de su búsqueda grabada**, con sus materias y su abreviatura— y la marca `vertebral: true` en las quince de la tabla de leyes vertebrales del mapa del sistema legal (`mapa-sistema-legal-skills.md` §1.4) y en ninguna más; `internal/skills/normas.go` con el campo `Vertebral bool`; `internal/skills/referencias.go` e `internal/skills/sincronia.go` con la **tabla de generadores** que sustituye al `switch` por nombre (`normas` → todas las normas, `leyes_vertebrales` → solo las marcadas, `jerarquia_normativa` → niveles y reglas leídos con el lector de T016) y con la cabecera, el título y las columnas derivadas de ella en vez de constantes de normas; `internal/skills/frontmatter.go`, donde cada referencia declara su fichero de datos y una referencia declarada sin generador sigue siendo un defecto con su mensaje; y `skills/boe-legislacion/references/normas.md` regenerado **solo** con `make skills-sync`, nunca a mano, que es lo único que cambia de esa skill; y `internal/app/skills_test.go`, solo en `probarNormasNombradas`, cuyo caso negativo daba como ausentes de la tabla la LEC y la LOPDGDD, dos de las siete normas que entran aquí, y pasa a nombrar normas con el número 0, que ninguna norma lleva, para no depender de lo que le falte a la tabla (redelimitada en el intento 1: lo razona `gates/tarea-T017.md`); es indivisible porque los mismos controles atan las tres piezas: `TestRegenerarYComparar` compara la salida de la tabla de generadores con las referencias del repositorio —que cambian al añadirse las normas—, `TestNormasDelRepositorio/vertebrales` exige que las marcadas sean exactamente las quince y `TestSkillsDelRepositorio`, que `make skills-sync` ejecuta, queda en rojo en cuanto la LEC y la LOPDGDD están en la tabla, de modo que separar datos, generadores y ese caso deja `make ci` en rojo entre dos tareas; con `TestRenderizarNormas`, `TestRenderizarLeyesVertebrales`, `TestRenderizarJerarquia`, `TestRegenerarYComparar`, `TestLeerFrontmatter`, `TestValidarFrontmatter`, `TestNormasDelRepositorio/vertebrales` (las quince y ninguna más), `TestIdentificadoresDeLasNormas` y `TestSkillsDelRepositorio` en verde; si un título o un rango grabados no coinciden con lo esperado, se detiene sin marcarse y lo anota en `gates/tarea-T017.md` (FR-065, FR-066, FR-067, FR-070, FR-071, FR-072, FR-074, US6 escenarios 1, 2 y 3, SC-009, SC-010, contrato de la skill §4).

**Checkpoint**: la jerarquía tiene lector y control propios, las quince leyes vertebrales están marcadas con su
identificador verificado, y las referencias se generan desde una tabla que admite tres generadores sin cablear nombres.

---

## Phase 7: El formato común de eval, extendido y compatible (US5)

**Objetivo**: que una eval pueda declarar y juzgar mecánicamente que la respuesta pasó por `territorio resolver` y trae
el territorio identificado, sin que ninguna eval de `boe-legislacion` deje de ser válida.

- [X] T018 [datos] [US5] Esquema del formato común de eval, con la expectativa de test que su forma impone: `schemas/eval.yaml.json` con exactamente los tres cambios del contrato de evals §1 y ninguno más —la cuarta variante `comando-territorio` (`applet`, `verbo` constante `resolver` y `municipio` no vacío, `additionalProperties: false`) como un `$ref` más del `oneOf` de `comandos`; el esperado `territorio` con `minProperties: 1` y las claves `comunidad`, `provincia`, `boletines` y `cobertura`, esta última sobre el enumerado cerrado `aspecto-de-cobertura` con las seis combinaciones del vocabulario del applet; y el `then` que pasa a `required: ["comandos"]` con `anyOf` de `citas` o `territorio`, con `territorio` añadido a lo que el `else` prohíbe— y, en la misma tarea, `internal/evals/formato.go` y `internal/evals/formato_test.go`, donde el caso `positiva-sin-citas` pasa a `positiva-sin-esperado-verificable` con el mensaje **que el esquema nuevo produce** (el nodo `anyOf` tiene causas y el lector desciende a las hojas de sus dos ramas), sin relajar la comparación; las evals de `boe-legislacion` siguen validando sin tocar un byte y la pausa revisa esquema y expectativa juntos (FR-084, FR-085, US5, SC-011, SC-015, contrato de evals §1, plan *Complexity Tracking* D28).

- [X] T019 [US5] Formato, juicio, consultas e informe con la variante de territorio: `internal/evals/formato.go` (campo `Municipio` en el comando esperado, campo `Territorio` en la eval y `formaDelComando` como **único** sitio que decide la variante, en lugar de los tres `switch` por verbo), `internal/evals/territorio.go` (nuevo, hermano de `citas.go`: `ExtraerTerritorio`, que saca de la respuesta lo declarado por su **forma fija** —comunidad y provincia con el pliegue del dominio, el código del boletín exacto y como palabra, y el aspecto de cobertura en la forma `<aspecto>: <valor>` con tolerancia a los espacios y a las mayúsculas y exacta en las dos palabras—), `internal/evals/juzgar.go` (`satisface` y `textoDelComando` para el comando de territorio, `TerritorioEncontrado` y `TerritorioAusente` en el resultado con su motivo, y la condición de `Pasa` que gana que no quede ninguno ausente), `internal/evals/consultas.go` (un comando de territorio **no genera ninguna consulta que grabar**, y la regla de lo grabado sigue aplicándose a los comandos del BOE), `internal/evals/conjunto.go` (`ComprobarConjunto(evals, normas, reglas)` parametrizado, con las diez reglas de `boe-legislacion` intactas y `ReglasDeLegalCore()` con `tamaño`, `cubierto`, `no cubierto`, `no activación` y `esperado verificable`) e `internal/evals/informe.go` (las dos columnas «Territorio encontrado» y «Territorio ausente»), con `TestLeerEval`, `TestFormaDelComando`, `TestExtraerTerritorio`, `TestJuzgar/territorio-satisface`, `/territorio-otro-municipio-no-satisface`, `/territorio-ausente`, `/territorio-y-citas`, `TestConsultasNecesarias`, `TestConjuntoDeEvals` sobre evals sintéticas, `TestInforme` y el subtest `TestEvalsDelRepositorio/cobertura-del-esquema` (el enumerado del esquema y el vocabulario del applet dicen lo mismo) escritos antes; el subtest `conjunto-legal-core` **no entra aquí**, porque exige tres evals que todavía no existen (FR-043, FR-084, US5, SC-011, SC-015, contrato de evals §2 y §3).

**Checkpoint**: el formato común admite y juzga la variante de territorio sin juicio de modelo, y todo lo escrito en H5
sigue siendo válido y en verde.

---

## Phase 8: Las evals y la skill `legal-core` (US5)

**Objetivo**: que las evals existan **antes** que el código de la skill y que la skill madre entregue el protocolo que
empieza por el territorio, con sus dos referencias generadas.

- [X] T020 [US5] Las tres evals de la skill, antes que su `SKILL.md`, y el control que las exige: primero `internal/evals/conjunto_test.go` con el subtest `TestEvalsDelRepositorio/conjunto-legal-core`, que aplica `ReglasDeLegalCore()` a la carpeta de evals de la skill y queda en rojo, y a continuación los tres ficheros que lo devuelven a verde —`evals/legal-core/01-territorio-municipio-cubierto.yaml` (activa, la pregunta de la aceptación sobre el municipio de referencia, comando de territorio y esperado con comunidad, provincia y su boletín), `evals/legal-core/02-territorio-municipio-no-cubierto.yaml` (la misma pregunta sobre un municipio de una comunidad sin configuración, con comunidad, provincia y los dos aspectos `no-configurado` en el esperado) y `evals/legal-core/03-no-activa-receta-de-cocina.yaml` (una pregunta que no es de territorio ni de derecho, con `activa: false`)—, **ninguna con `citas`**, porque `legal-core` no afirma contenido de norma; los municipios concretos viven aquí y en los fixtures, nunca en el código ni en los datos; se confirman en un commit anterior al primero que crea la skill (FR-024, FR-069, FR-080, FR-081, FR-082, FR-083, US5 escenarios 1, 2 y 3, SC-011, SC-015, contrato de evals §3 y §4).

- [X] T021 [US5] Casos negativos de skills parametrizados por skill, **sin añadir ninguna skill todavía**: `internal/app/skills_test.go`, donde la constante `skillDelHito` pasa a ser la lista `skillsExigidas` —con un solo elemento por ahora—, la exigencia de `TestSkillsDelRepositorio/skills` pasa de `require.Contains` a `require.Subset` sobre ella, y los casos negativos `/skills`, `/normas-nombradas` y `/sin-instrucciones-de-evals` dejan de estar cableados a `boe-legislacion` y se ejercen sobre cada skill del recorrido; con una sola skill instalada el recorrido pasa por una sola, `make ci` sigue en verde y ningún control cambia de resultado (FR-060, FR-064, SC-009, contrato de la skill §5, plan *Complexity Tracking* D26).

- [X] T028 [datos] [US5] La reinstalación comprobada sin nombrar las skills, antes de que llegue la segunda: `internal/skills/testdata/script/instalar-de-nuevo.txtar`, cuya última comprobación deja de exigir que el directorio personal de skills tenga exactamente `boe-legislacion` y pasa a exigir una entrada por cada skill del árbol copiado y ninguna más —la lista sale de recorrer los directorios de skills del árbol como los recorre la instalación, tiene que traer `boe-legislacion` para no pasar en vacío y se compara byte a byte con el listado del directorio personal—, con el comentario de cabecera alineado y nada más; es tarea de datos aparte porque ese guion es material de test existente y T022 no lleva la etiqueta que deja tocarlo, mientras que en cuanto `legal-core` existe la instalación la enlaza también y la expectativa literal queda en rojo, y es divisible porque la comprobación nueva está en verde con una skill y con dos (redelimitación de T022 en su intento 1, razonada en su nota); la pausa revisa el guion; no toca código ni ningún otro guion, y `make ci` queda en verde por sí sola (FR-060, FR-085; FR-053 y SC-006 de H5).

- [X] T022 [US5] La skill `legal-core` v0, primero la exigencia y después sus ficheros: `internal/app/skills_test.go` añade `legal-core` a `skillsExigidas` —el control de que la skill existe queda en rojo—, y a continuación `skills/legal-core/SKILL.md` (frontmatter válido con `name: legal-core`, `description` no vacía, `metadata.kitlegal-applets` con **solo** `territorio` y `metadata.kitlegal-referencias` con `leyes_vertebrales jerarquia_normativa`; menos de 300 líneas contadas sobre el fichero regenerado; protocolo que **empieza por identificar el territorio** —y pregunta el municipio cuando la conversación no lo dice, sin suponerlo—, lo resuelve con el binario con `--json` sin dar por sabido ningún dato de territorio, lee `cobertura` y la traslada a la respuesta sin nombrar ningún boletín que el applet no haya devuelto, trata la ambigüedad ofreciendo los candidatos y la ausencia sin concluir «no existe», razona con las dos referencias y **delega en `boe-legislacion`**, en un solo sentido, el texto de cualquier artículo; región generada de la tabla de comandos con **solo** los verbos de `territorio`; las seis reglas invariantes del contrato §1.4; nada de plazos, recursos ni competencia, que son de H9; y sin mencionar evals, el job, modelos ni la caché), `skills/legal-core/references/leyes_vertebrales.md` y `skills/legal-core/references/jerarquia_normativa.md` **regeneradas con `make skills-sync`** desde los datos, con su cabecera literal «generado …, no editar» y sin el texto de ninguna norma, y el enlace `skills/legal-core/scripts/territorio` al binario instalado; `TestSkillsDelRepositorio`, `TestTablaDeComandosCoincideConLaGramatica` y `TestRegenerarYComparar` vuelven a verde con diff vacío (FR-060 a FR-069, FR-102, US5 escenario 4, SC-009, contrato de la skill §1 a §3).

**Checkpoint**: la skill madre está entregada, medida por tres evals escritas antes que ella y vigilada por los mismos
controles negativos que `boe-legislacion`.

---

## Phase 9: Job de evals y documentación

- [ ] T023 [US5] Matriz de skills en el job de evals: `.github/workflows/evals.yml` pasa de una skill a una matriz (`boe-legislacion` y `legal-core`) con `fail-fast: false` y un informe por skill dentro de la **misma** ejecución, la prueba de red se mantiene como bandera solo para `boe-legislacion` —`territorio` no tiene red que probar, porque no puede pedir nada por construcción— y el filtro de rutas gana el patrón de los paquetes de dominio (el de `core`, junto al de `data`, que ya cubre los ficheros congelados); el umbral no cambia (una eval pasa cuando al menos 2 de sus 3 sesiones pasan, y el informe lleva el commit evaluado y el identificador del modelo); ninguna tarea ejecuta `make evals` (FR-083, SC-015, contrato de evals §5).

- [ ] T024 [P] Documentación del hito alineada en la misma rama: `CHANGELOG.md` (sección *Unreleased*, «Añadido») con el applet `territorio` y su verbo `resolver`, la skill `legal-core` v0 con sus dos referencias generadas, los datos congelados de territorio y de jerarquía, los identificadores de código INE y DIR3 y la variante de territorio del formato común de eval; `README.md` y `CONTRIBUTING.md` donde enumeran applets, skills, carpetas de datos y pasos de trabajo, incluida la fila de `make skills-check` de sus tablas de controles, que pasa a cubrir también la validación de los ficheros congelados y de la jerarquía; ninguna otra cifra ni sección, **ningún ADR** —ADR 0017 ya decidió el origen de los datos y ADR 0006 la procedencia del sobre de un applet calculado, y este hito no cambia ninguna decisión de arquitectura— y ninguna fila nueva en la tabla de fuentes, que T003 dejó cierta (FR-098, FR-099, SC-014, Definition of Done §1.6 y §1.7, plan obligación 8).

**Checkpoint**: el job mide las dos skills en la misma ejecución y la documentación del repositorio dice lo que el hito
entrega.

---

## Phase 10: Cierre local

- [ ] T025 Cierre sin tocar el árbol fuera del directorio del hito: ejecuta en primer plano `make ci` y, tal cual y en orden, los prerrequisitos y los escenarios 1 a 13 de `specs/008-h6-territorio-skill-legal/quickstart.md`, con las formas `rtk proxy` de su tabla, y se detiene sin marcarse y lo anota en `gates/tarea-T025.md` ante cualquier resultado distinto del esperado; mide sobre el perfil que deja `make ci` la cobertura global (≥ 70 %) y la de los paquetes de dominio (≥ 85 %) con `go tool cover -func`; y escribe `specs/008-h6-territorio-skill-legal/gates/pr-h6.md`, el cuerpo de la publicación del hito, con la plantilla del ritual (objetivo, alcance, controles añadidos, decisiones, pendientes), «Dependencias: ninguna nueva» y el motivo del único módulo que pasa a enlazarse, las medidas fechadas por el commit sobre el que se tomaron y, en pendientes, la regla de SC-015 sobre qué puede cambiar después de la ejecución de aceptación; no escribe ningún otro fichero (FR-093, FR-094, FR-095, FR-096, FR-097, SC-012, Definition of Done §1.1 y §1.9, ritual §6.3, plan obligación 9).

**Checkpoint**: el hito está implementado y validado con su guía, con la cobertura medida y el cuerpo de la publicación
escrito; solo queda la plataforma.

---

## Phase 11: Plataforma (al final; fusionar es humano)

- [ ] T026 [plataforma] Publicación y ejecución de aceptación: ejecuta tal cual el §14 de `specs/008-h6-territorio-skill-legal/quickstart.md` —comprueba los prerrequisitos (sesión de `gh`, secreto del token, etiqueta del job y cuerpo presente), empuja la rama con `git push -u origin 008-h6-territorio-skill-legal` y, si no existe, la abre con `gh pr create` hacia `main` con el título del hito y el cuerpo `specs/008-h6-territorio-skill-legal/gates/pr-h6.md`; identifica y espera la primera ejecución de evals de la rama disparada por la propuesta de cambio; lee los dos informes de la matriz y comprueba que **en esa misma ejecución** pasan las tres evals de `legal-core` —las dos de activación positiva y la de no activación, cada una con al menos 2 de sus 3 sesiones—, que el conjunto de `boe-legislacion` sigue pasando sin que ninguno de sus ficheros de eval se haya modificado, que ninguna petición llegó a la red de una fuente, y que el commit del informe es de la rama y la cabeza solo difiere de él en ficheros bajo `specs/008-h6-territorio-skill-legal/`; comprueba además que la pregunta de la aceptación se responde para el municipio cubierto y para el no cubierto, con la cobertura explícita en el segundo— y registra los datos y las salidas enteras en `specs/008-h6-territorio-skill-legal/gates/evals-cierre.md`; si falta un prerrequisito o una comprobación falla, registra la salida, se detiene sin marcarse y lo anota en `gates/tarea-T026.md` (el arreglo va en una tarea nueva colocada **antes** de esta, y no se relanza la ejecución para buscar otro resultado); nunca fusiona, ni empuja a `main`, ni fuerza, ni borra ramas, etiquetas o releases (FR-100, FR-101, SC-013, SC-015, plan obligación 9).

---

## Definition of Done: qué punto cubre qué tarea

| Punto (`docs/ROADMAP.md` §1) | Aplica | Tarea |
|---|---|---|
| 1. `make ci` en verde | Sí | Cada tarea (batería); T025 lo repite sobre el árbol terminado |
| 2. Tests offline; fixtures grabados si toca red | Sí | T004, T006, T007, T009, T011, T016, T017, T019 a T022 (tests offline); T005 (corpus de fuzz), T012 (guion e2e), T015 (las únicas grabaciones nuevas, que hace una persona en su pausa); ninguna fuente de H6 se pide en red (FR-043) |
| 3. Sin `net/http`, `os.Exit`, `fmt.Print*` fuera de lo autorizado | Sí | T004, T006, T007 y T009 no importan nada de eso; lo vigilan `depguard`, `forbidigo` y `TestArquitectura` en cada `make ci`, y T009 actualiza la lista de módulos del binario con su motivo |
| 4. Salida del applet contra sus esquemas; `schema-check` | Sí | T008 (tabla parametrizada), T010 (`municipio.json` publicado sin deriva), T011 (la salida real validada) |
| 5. Errores tipados y exit codes; sin `panic` | Sí | T004 y T006 (errores con su clase), T009 (`TestCodigosDeTerritorio`: 0, 2 y 3, y nunca 4, 5 ni 6), T005 (fuzz sin `panic`) |
| 6. Comportamiento visible: e2e y `CHANGELOG.md` | Sí | T012 (`territorio-matriz.txtar`, más los dos guiones existentes que T010 actualiza), T024 (`CHANGELOG.md`) |
| 7. ADR si cambia una decisión de arquitectura | No | Ninguna decisión de arquitectura cambia: ADR 0017 fijó el origen de los datos y ADR 0006 la procedencia del sobre de un applet calculado (spec, *Fuera de alcance*); T024 lo deja dicho |
| 8. `docs/SOURCES.md` y `verify-sources.sh` si toca una fuente | Sí, en su parte de documento | T003 (fila `mpt.rel` con la fecha del volcado); **ningún caso nuevo** en el guion de verificación, porque estas fuentes no se piden en red (FR-049, ADR 0017) |
| 9. Cobertura `core` ≥ 85 % y global ≥ 70 % | Sí | T004 y T006 la dan con sus propios tests sintéticos; T025 la mide sobre el árbol terminado |
| 10. Skill: evals antes que el código, `SKILL.md` < 300, `references/` sin deriva | Sí | T020 antes que T022 (FR-083); T022 (frontmatter, menos de 300 líneas, dos referencias regeneradas); T016 y T017 (lector de la jerarquía, tabla de generadores y `references/normas.md` regenerado); T023 y T026 (el verde en el job) |
| 11. Dimensión territorial | Sí | T012 (matriz con municipio cubierto y no cubierto y la cobertura declarada); T006, T007 y T020 (ningún caso especial: los municipios solo en fixtures, e2e y evals) |
| 12. Grafo (desde H7) | No | H6 es anterior a H7: `territorio` no emite operaciones de grafo (ADR 0014; spec, *Fuera de alcance*) |

## Trazabilidad: requisito → tarea

| Requisito | Tareas |
|---|---|
| FR-001 a FR-008 | T006, T009, T010 (T025 lo valida con quickstart §2) |
| FR-009 | T009, T012 |
| FR-010 a FR-016 | T006, T009, T012 |
| FR-020, FR-021, FR-022 | T006, T009, T011, T012 |
| FR-023 | T003, T006, T009 |
| FR-024 | T006, T007, T020 |
| FR-030 a FR-034 | T004; T027 (los patrones de los esquemas de territorio, iguales a sus gramáticas) |
| FR-035 | T004 (objetivos y semillas), T005 (corpus versionado) |
| FR-040, FR-041 | T001 |
| FR-042 | T003 |
| FR-043 | T003, T019 (un comando de territorio no genera consulta), batería (`TestArquitectura` R2) |
| FR-044 | T001, T002, T003, T014 y T027 (esquemas), T007 y T016 (validación en `make ci`) |
| FR-045 | T001, T002, T003 |
| FR-046, FR-047, FR-048 | T003 |
| FR-049 | T003, T024 |
| FR-050 a FR-053 | T002 |
| FR-054, FR-055 | T002, T006 |
| FR-056 | T007, T012 |
| FR-060 | T021, T022; T028 (la reinstalación comprobada sin nombrar las skills) |
| FR-061 a FR-064 | T022 |
| FR-065, FR-066 | T016 (lector de la jerarquía), T017 (generadores), T022 (ficheros generados) |
| FR-067 | T013, T014, T016, T017, T022 |
| FR-068, FR-069 | T020, T022 |
| FR-070 | T013, T017 |
| FR-071, FR-072 | T015, T017 |
| FR-073 | T015 |
| FR-074 | T017 |
| FR-080, FR-081, FR-082 | T020 |
| FR-083 | T020 (antes que T022), T023, T026 |
| FR-084 | T018, T019 |
| FR-085 | T013, T015, T018, T028 |
| FR-086 | T001, T002, T003, T005, T010, T012, T014, T015, T027 |
| FR-090, FR-091 | T012 |
| FR-092 | T008, T010, T011 |
| FR-093, FR-094 | Cada tarea (batería); T025 |
| FR-095 | T004, T006 (cobertura del dominio), T025 (medida) |
| FR-096 | T004, T006, T009 |
| FR-097 | T004, T006, T007, T009 |
| FR-098, FR-099 | T024 |
| FR-100, FR-101 | T022 (la skill que responde), T026 (la ejecución que lo mide) |
| FR-102 | T022 |
| SC-001 | T009, T012 (T025, quickstart §2 y §3) |
| SC-002 | T011, T012 (T025, quickstart §4) |
| SC-003 | T002, T006, T007 (subtest `regimen-de-todas`), T012 (T025, quickstart §5) |
| SC-004 | T009, T012 (T025, quickstart §6) |
| SC-005 | T002, T007 |
| SC-006 | T004, T005 (T025, quickstart §10) |
| SC-007 | T001, T002, T003, T007 |
| SC-008 | T003, T006, T011 |
| SC-009 | T016, T017, T021, T022 (T025, quickstart §11) |
| SC-010 | T013, T015, T017 |
| SC-011 | T018, T019, T020 (T025, quickstart §12) |
| SC-012 | T007, T008, T010, T012, T025 |
| SC-013 | T026 |
| SC-014 | T003, T024 (T025, quickstart §13) |
| SC-015 | T018, T019, T020, T023, T026 |

## Obligaciones que el plan trasladó a este fichero

| Obligación del plan | Dónde se cumple |
|---|---|
| 1. Orden y pausas | Una tarea por paso de «Orden de implementación», en su orden: 1→T001, 2→T002, 3→T003, 4→T004, 5→T005, 6→T006, 7→T007, 8→T008, 9→T009, 10→T010, 11→T011, 12→T012, 13→T013, 14→T014, 15→T015, 16→T016 y T017 (el paso 16 se parte en dos rebanadas verdes: el lector de la jerarquía y, después, las normas vertebrales con la tabla de generadores), 17→T018, 18→T019, 19→T020, 20→T021, 21→T022, 22→T023, 23→T024; el paso 24 se parte en T025 (cierre local, que escribe el cuerpo) y T026 (`[plataforma]`). Las diez `[datos]` del plan son T001, T002, T003, T005, T010, T012, T013, T014, T015 y T018, y solo T010 y T018 declaran una ruta de código, con su razón; T027, la undécima, entró al redelimitar T007 y va justo antes que ella, sin código; T028, la duodécima, entró al redelimitar T022 y va justo antes que ella, sin código |
| 2. Sin descargar, grabar ni escribir de memoria | Batería; T001, T002, T003 y T015 dicen expresamente que el ejecutor no descarga ni graba; T017 se detiene si un título o un rango grabados no coinciden |
| 3. Rutas declaradas | Cada línea nombra por su ruta completa los ficheros que crea o cambia, ficheros de test incluidos cuando no los cubre su fichero de código; lo que solo se lee, se compara o se ejecuta va sin carpeta (`municipios.yaml`, `municipio.json`, `jerarquia.yaml`, `SOURCES.md`, `verify-sources.sh`, `mapa-sistema-legal-skills.md`) o por su test o su objetivo de `make`; desde T001 ninguna línea posterior deja extraer las carpetas de datos, de esquemas ni de fixtures |
| 4. `misspell` | Batería; ninguna tarea declara `.golangci.yml`, y si la implementación demostrara que necesita una entrada literal, se redelimita la tarea declarándolo con su motivo |
| 5. Sin atajos | Batería; ningún control se relaja (T018 lo dice expresamente para el mensaje del lector de evals) |
| 6. Sin cambios fuera del alcance | Batería; `skills/boe-legislacion/references/normas.md` solo en T017 y solo por regeneración |
| 7. `doc.go` y cobertura del dominio | T004 y T006 nacen con su `doc.go` y con sus propios tests sintéticos, que son los que cuentan para el umbral |
| 8. Documentación alineada | T024, incluida la fila de `make skills-check` que T007 y T016 cambian |
| 9. Evidencias en `gates/` | T003 (`verificacion-dir3.md`), T025 (`pr-h6.md`), T026 (`evals-cierre.md`); después de la ejecución de aceptación ninguna tarea toca código, datos, esquemas ni documentación |
| 10. Umbrales | T025 los mide; ninguna tarea los rebaja ni excluye un fichero |
| 11. `rtk` y las formas de la sesión | Batería; T025 y T026 ejecutan el quickstart con las formas de su tabla |
| 12. Sin pasar en vacío | T007 (los cuatro ficheros y las 19 comunidades), T016 (los cinco niveles y las cuatro reglas), T017 (`vertebrales`: las quince), T020 (el subtest antes que las tres evals), T022 (`legal-core` en `skillsExigidas` antes que sus ficheros) |

## Dependencias y orden

El orden es estrictamente secuencial: **ninguna tarea depende de una posterior**.

- **T001 → T002 → T003**: la integridad que comprueba `Cargar` exige que toda provincia esté declarada por una comunidad
  y que todo municipio de la correspondencia exista en la relación, así que el orden de los tres ficheros es el de sus
  referencias. Ninguno tiene lector todavía: los tres dejan `make ci` en verde.
- **T004 → T005**: el corpus se indexa por el nombre del objetivo de fuzz, que nace en T004.
- **T004 → T006**: el dominio del territorio analiza códigos INE y DIR3 con los tipos de T004.
- **T001, T002, T003, T006 → T007**: `TestTerritorioDelRepositorio` exige los cuatro ficheros reales y delega la
  integridad en `Cargar`; el paquete embebido no puede apuntar a ficheros que no existen.
- **T004 → T027 → T007**: los patrones de los esquemas se igualan a las gramáticas de T004 antes de que el subtest
  `gramaticas` de T007 los compare; T027 va en el fichero justo antes que T007, porque el bucle toma la primera tarea
  sin marcar en el orden del fichero y no por su número.
- **T007 → T009**: la raíz de producción y el binario de e2e pasan las fuentes embebidas.
- **T008 → T010**: la tabla de esquemas tiene que estar parametrizada antes de que exista una fila que no sea `boe`.
- **T009 → T010 → T011 → T012**: el esquema se genera desde el applet ya registrado; el test contra el contrato
  publicado necesita el fichero; el guion e2e necesita el applet registrado también en el binario de e2e.
- **T013 → T015 → T017**: la propiedad `vertebral` antes que las marcas; las grabaciones y el `enum` antes que las
  normas que los usan, porque el identificador, el título y el rango se copian de ellas.
- **T014 → T016**: el lector de la jerarquía y su control necesitan el fichero y su esquema.
- **T016 → T017**: el generador `jerarquia_normativa` de la tabla lee la jerarquía con el lector de T016, así que el
  lector y su control van antes que la tabla que los usa.
- **T018 → T019 → T020**: el esquema antes que el campo y su juicio; las reglas del conjunto antes que el subtest que
  las aplica a las evals.
- **T020 → T022**: las evals se confirman en un commit anterior al primero que crea la skill (FR-083).
- **T021 → T022**: los casos negativos se parametrizan antes de que exista una segunda skill a la que alcanzar.
- **T028 → T022**: el guion de reinstalación deja de nombrar las skills antes de que exista la segunda, que la
  instalación enlaza en cuanto está en el árbol; T028 va en el fichero justo antes que T022, porque el bucle toma la
  primera tarea sin marcar en el orden del fichero y no por su número.
- **T014, T016, T017 → T022**: las dos referencias se generan desde `data/`, así que la jerarquía, su lector, las marcas
  `vertebral` y la tabla de generadores tienen que existir antes.
- **T022 → T023 → T024 → T025 → T026**: el job mide la skill; la documentación describe lo entregado; el cierre valida el
  árbol terminado y escribe el cuerpo que la plataforma publica.

## Oportunidades de paralelismo

En el workflow todas las tareas van en secuencia. Fuera de él podrían adelantarse **T004 y T005** (el paquete de
identificadores no comparte fichero con las tres primeras tareas), **T008** (solo toca la tabla de esquemas), **T013 y
T014** (material de datos que ningún lector mira hasta T016 y T017) y **T024** (solo documentación). El resto comparte
fichero con su vecina o depende de ella.

```text
Tarea: "T004 internal/core/ids con sus tests y sus dos objetivos de fuzz"
Tarea: "T008 tabla de esquemas parametrizada por applet, sin filas nuevas"
Tarea: "T024 CHANGELOG, README y CONTRIBUTING del hito"
```

## Estrategia de implementación

### MVP (US1 y US2: la entrega y la aceptación)

1. Fases 1 a 4 (T001-T011): datos congelados, identificadores, dominio, applet y contrato publicado. **Validar**:
   `make ci` en verde con `TestTerritorioDelRepositorio`, `TestResolverDevuelveElTerritorio`, `TestCodigosDeTerritorio`
   y `TestSalidaSinBoletinesNoConfigurados`; ya se responde por municipio cubierto y no cubierto con la cobertura
   declarada.
2. Fase 5 (T012): la matriz territorial cierra el MVP de extremo a extremo.

### Entrega incremental

- US4 (T004, T005, T009) da los identificadores y la distinción entre ambiguo, inexistente y mal formado.
- US3 (T002, T006, T012) marca el régimen foral allí donde lo hay.
- US6 (T013, T015, T017) cierra la tabla de leyes vertebrales con identificadores verificados.
- US5 (T014, T016, T018 a T021, T028, T022 y T023) extiende el formato de eval, escribe las evals y entrega la skill
  madre.
- US7 (T001, T003, T027, T007) deja el terreno congelado, verificado y validado en `make ci`.
- T024-T026 documentan, validan y publican.

## Notas

- **Lo que este hito no crea, y no por olvido** (spec, *Fuera de alcance*): festivos y plazos, operaciones de grafo,
  `.kitlegal/config.yaml` y el asunto, otros verbos de `territorio`, otros territorios configurados, la carpeta de
  boletines y su motor, órganos distintos del ayuntamiento, consultas de red a INE, REL o DIR3, identificadores
  distintos de INE y DIR3, `internal/core/competencia`, `legal-core` v1, cambios en el protocolo, las referencias o las
  evals de `boe-legislacion` más allá de regenerar su tabla de normas, packs por vertical, geometría y callejero, un
  tercer estado en `cobertura`, la lectura de los datos desde el sistema de ficheros, un formato de evals propio de
  `legal-core`, la refactorización de `internal/cli` y un ADR nuevo.
- **Nombre de la tercera eval**: el contrato deja abierto el sufijo de la eval de no activación; aquí queda fijado como
  `03-no-activa-receta-de-cocina.yaml`, en la forma de las dos de no activación que ya tiene `boe-legislacion`.
- **Ficheros existentes que se tocan**, exactamente: `docs/SOURCES.md` (T003); los esquemas de municipios, DIR3 y
  comunidad que fijaron T001, T002 y T003 (T027, solo sus patrones de identificador); `internal/skills/export_test.go`
  (T007); `internal/app/esquemas_test.go` (T008 y
  T010); `internal/arch_test.go` (T009); `internal/app/registro.go`, `internal/app/registro_test.go`,
  `cmd/kitlegal/main_test.go`, `internal/app/ejemplo/kitlegal-e2e/main.go`,
  `internal/app/testdata/script/argumentos.txtar` e `internal/app/testdata/script/ayuda.txtar` (T010);
  `internal/app/territorio_test.go` (T009 y T011); `schemas/normas.yaml.json` (T013 el campo, T015 el `enum`);
  `testdata/evals/` (T015); `data/normas.yaml`, `internal/skills/normas.go`, `internal/skills/referencias.go`,
  `internal/skills/sincronia.go`, `internal/skills/frontmatter.go`,
  `skills/boe-legislacion/references/normas.md` (T017); `Makefile` (T007 y T016); `schemas/eval.yaml.json`,
  `internal/evals/formato.go` y `internal/evals/formato_test.go` (T018; `formato.go` también en T019);
  `internal/evals/juzgar.go`, `consultas.go`, `conjunto.go` e `informe.go` (T019); `internal/evals/conjunto_test.go`
  (T019 y T020); `internal/app/skills_test.go` (T017, T021 y T022);
  `internal/skills/testdata/script/instalar-de-nuevo.txtar` (T028); `.github/workflows/evals.yml` (T023); `CHANGELOG.md`,
  `README.md` y `CONTRIBUTING.md` (T024).
- **Rutas protegidas desde T001**: ninguna tarea posterior a una `[datos]` nombra con su carpeta el material que aquella
  fijó. Lo que solo se lee o se compara aparece sin carpeta (`municipios.yaml`, `municipio.json`, `jerarquia.yaml`,
  `SOURCES.md`, `verify-sources.sh`, `mapa-sistema-legal-skills.md`, `norma.json`, `bloque.json`), de modo que el
  guardián rechaza cualquier cambio en ellos fuera de la tarea que los declara.
- **Dos tareas tocan el mismo fichero y lo dicen las dos**: `internal/app/esquemas_test.go` (T008 y T010),
  `internal/app/territorio_test.go` (T009 y T011), `internal/evals/formato.go` y `formato_test.go` (T018 y T019),
  `internal/evals/conjunto_test.go` (T019 y T020), `internal/app/skills_test.go` (T017, T021 y T022),
  `schemas/normas.yaml.json` (T013 y T015) y `Makefile` (T007 y T016, cada una con el test que su `-run` gana).
- **`gosec`**: las lecturas de ficheros del repositorio pasan por `filepath.Clean`; las escrituras de test, a
  `t.TempDir()` con `0o600`.

## Comprobación contra la rúbrica del juez (`juez_tasks`, criterios a-g)

| Criterio | Dónde se cumple |
|---|---|
| a. analisis_critico | `gates/analyze.md` lo escribe el paso `analyze`, que sigue a este; las doce obligaciones que plan.md traslada están asignadas a tareas concretas en la tabla «Obligaciones», y las tres piezas indivisibles que el plan razona (D16, D28, D29) tienen su tarea única con el motivo escrito en la propia línea |
| b. trazabilidad | Tabla «Trazabilidad» con **todos** los FR del spec (001-016, 020-024, 030-035, 040-049, 050-056, 060-069, 070-074, 080-086, 090-102) y todos los SC (001-015); cada tarea cita en su línea los requisitos que cumple; ninguna tarea añade nada que el spec o el plan no pidan, y lo que el spec deja fuera está enumerado en «Notas» |
| c. rebanadas_verdes | Cada tarea de código lleva su test escrito primero y la implementación mínima que lo hace pasar; las excepciones están declaradas una a una con su razón en «Rebanadas verticales y excepciones declaradas» y ninguna deja un test en rojo al terminar; el orden es secuencial y está justificado en «Dependencias y orden»; el e2e que describe la entrega va en T012, la primera tarea que puede dejarlo en verde (necesita el applet registrado en el binario de e2e, T010) |
| d. rutas_declaradas | Cada línea nombra por su ruta completa los ficheros que crea o cambia, sin llaves, comodines ni rutas de paquete o de carpeta genérica; los ficheros de test van por su ruta cuando no los cubre la regla «declarar `x.go` permite `x_test.go`»; lo que solo se lee o se ejecuta va sin carpeta o por su test o su objetivo de `make` |
| e. datos_separados | Las doce `[datos]` (T001, T002, T003, T005, T010, T012, T013, T014, T015, T018, T027 y T028, estas dos últimas entradas al redelimitar T007 y T022) son las únicas líneas que nombran material bajo esquemas o fixtures; solo T010 y T018 declaran código, por la indivisibilidad que un control existente impone y que plan.md razona en *Complexity Tracking*, y las dos lo revisan en su pausa; T003 añade la fila de fuentes y el registro de la muestra porque FR-049 y FR-047 los ponen dentro de esa misma pausa; T026 es la única `[plataforma]`, va la última, no mezcla otro trabajo y nunca fusiona |
| f. dod | Tabla «Definition of Done» con los doce puntos: los aplicables con su tarea (`CHANGELOG.md` en T024, esquemas en T001, T002, T003, T010, T013, T014, T018 y T027, fixtures en T005, T012 y T015, `docs/SOURCES.md` en T003, evals antes que la skill en T020 y T022, cobertura en T025, e2e en T012) y los no aplicables con su razón (ADR y grafo) |
| g. checklist_veraz | `checklists/requirements.md` está enteramente marcado y ninguna tarea lo modifica; sus marcas describen el spec, que las tareas no cambian |
