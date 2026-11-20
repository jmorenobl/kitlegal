# Data model: H33 · Cobertura en positivo por capacidad y materia

Las entidades de «Key Entities» del spec, con dónde viven, sus campos y sus reglas. Los textos literales son los de
FR-001, FR-004, FR-005, FR-008 y FR-009; aquí no se repiten. Las formas exactas de cada salida están en contracts/.

## 1. `data/cobertura.yaml` (`skills.Cobertura`)

Leído por `skills.LeerCobertura` contra `schemas/cobertura.yaml.json` (research D5). Solo lo lee la sincronía de las
skills; el binario no lo lleva.

| Campo YAML | Go (`internal/skills/cobertura.go`) | Regla |
|---|---|---|
| `motivo_de_lo_no_declarado` | `Cobertura.MotivoDeLoNoDeclarado string` | obligatorio, al menos un carácter; hoy `fuera de lo que kitlegal cubre` |
| `cubierto` | `Cobertura.Cubierto []FilaCubierta` | opcional; en el repositorio, las 4 filas de FR-001 en su orden |
| `regulados_aparte` | `Cobertura.ReguladosAparte []ReguladoAparte` | sin cambios: obligatorio, al menos una fila, sus 9 filas de hoy |

### 1.1 Fila de lo cubierto (`skills.FilaCubierta`)

| Campo YAML | Go | Regla |
|---|---|---|
| `capacidad` | `Capacidad string` | obligatorio; `normativa`, `jurisprudencia`, `plazos` o `territorio` (`enum`) |
| `alcance` | `Alcance string` | obligatorio, al menos un carácter |
| `materia` | `Materia string` | opcional, cadena (el esquema no le pone más regla); si va, el `vertical` de alguna norma de `data/normas.yaml`, que lo comprueba solo la sincronía (research D4) |

Sin claves de más (`additionalProperties: false`): `normas` es un defecto. Sin `materia`, la fila sirve a todas las
materias. Dos filas de una capacidad son dos filas.

## 2. La línea de lo no cubierto (`internal/skills/regiones.go`)

`lineaDeNoCubierto` sustituye a `formasDeNoCubierto` (research D2):

| Campo | De dónde sale | Regla |
|---|---|---|
| `plantilla` | la única línea de los bloques `text` del `SKILL.md`, sin sus blancos del principio, que empieza por `⚠ NO CUBIERTO: ` y lleva `<motivo>` | con 0 o con 2 o más, la skill no tiene plantilla |
| `motivoConNorma` | la línea de los bloques `text` de la referencia con la región que lleva `<precepto>` y `<norma>` y no empieza por `⚠` | exactamente una |
| `motivoSinNorma` | la que lleva `<precepto>`, no `<norma>` y no empieza por `⚠` | exactamente una |

Marcadores conocidos (`marcadoresDeLaCobertura`): `<procedimiento>`, `<asunto>`, `<motivo>`, `<norma>`,
`<identificador de la norma>`, `<precepto>`, `<identificador>`, `<id>`. La fila de una entrada de `regulados_aparte`:
la plantilla con `<motivo>` por el motivo de su caso (con norma si la entrada la trae) y, en una sola pasada,
`<asunto>` y `<procedimiento>` por el procedimiento, `<norma>` por su nombre, `<identificador de la norma>` por su
identificador, `<precepto>` por la forma legible del apartado, `<identificador>` por la norma del precepto e `<id>` por
su bloque.

`regionesDeLaSkill.conFormas` pasa a `conPlantilla`: las skills con plantilla son las que dan los defectos de las filas
sin región (`anotarFilasSinRegion`) y las que la sincronía exige que digan el motivo fijo (research D3). En el
repositorio: `consulta-juridica` y `plazos-y-territorio`.

## 3. La región de lo cubierto (`internal/skills/cubierto.go`)

| Elemento | Valor |
|---|---|
| marca de inicio | `<!-- lo que cubre la base: lo genera make skills-sync -->` |
| marca de fin | `<!-- fin de lo que cubre la base -->` |
| línea por fila | `- **` + capacidad con la primera letra en mayúscula + `**: ` + alcance + `\n`, en el orden de `cubierto` |
| skills que la llevan | las que tienen sus marcas; hoy solo `consulta-juridica` |

Estados: con `data/cobertura.yaml` y `data/normas.yaml` legibles y cada `materia` un `vertical`, lo de entre las
marcas es la lista; si no, queda como está y la skill da el defecto (contracts/datos-y-regiones.md §3). Las marcas que
no van en pareja dan el defecto de hoy de `regionDeMarcas.buscar`.

## 4. La eval de cobertura (`internal/evals/formato.go`)

| Campo YAML | Go | Regla |
|---|---|---|
| `no_cubierto.norma`, `no_cubierto.precepto` | `NoCubiertoEsperado.Norma`, `.Precepto` | sin cambios; juntos (lo regulado aparte) |
| `no_cubierto.fuera_de_lo_declarado` | `NoCubiertoEsperado.FueraDeLoDeclarado bool` | solo `true`; nunca con `norma` ni con `precepto` |
| `tambien_la_resuelven` | `Eval.TambienLaResuelven []string` | opcional; al menos un nombre, sin repetir, cada uno una skill de `skills/` (regla «skills nombradas») |

`Eval.esDeCobertura()`: `Norma != ""` o `FueraDeLoDeclarado`. `Eval.activadaPor(sesion, skill)`: la sesión activó la
skill o alguna de `TambienLaResuelven`.

## 5. Las evals del hito

| Eval | Lado de la frontera | Norma sin `vertical` | Cobertura |
|---|---|---|---|
| `consulta-juridica/01` (la póliza) | fuera | `BOE-A-1980-22501` | `fuera_de_lo_declarado` |
| `consulta-juridica/02` (vacaciones) | dentro | `BOE-A-2015-11430` | — |
| `consulta-juridica/03` (la prórroga), nueva | dentro | `BOE-A-1980-22501` | — |
| `consulta-juridica/04` (el despido), nueva | fuera | `BOE-A-2015-11430` | `fuera_de_lo_declarado` |
| `normativa/17` | — | — | gana `tambien_la_resuelven: [plazos-y-territorio]` |
| `plazos-y-territorio/07`, `08`, `12` | lo regulado aparte | — | `{norma, precepto}`, sin cambiar de fichero |

Regla «frontera en dos materias»: las normas sin `vertical` con una eval activa que la cita y no es de cobertura y otra
activa que la cita, es de cobertura con `fuera_de_lo_declarado` y prohíbe `plazos calcular`; al menos dos. Hoy, las dos
de la tabla.

## 6. Lo grabado

Una entrada nueva del manifiesto de H9 (`internal/evals/testdata/grabaciones.json`): el ET con el bloque `a59`
(research D11). Lo grabado: `internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-11430_texto_bloque_a59.json`.
