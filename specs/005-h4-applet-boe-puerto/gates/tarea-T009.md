# T009 · Grabaciones, referencias y fila revisadas

## Intento 1 (2026-09-13)

**Sin marcar: parada para decisión humana** (obligación 2 del plan; research D9). Los tres tests de la tarea están
escritos y dos quedan en rojo **por los datos que fija una persona, no por el código**: faltan las cinco referencias, dos
de los cinco artículos del diff no existen en la fuente y la gramática del id de bloque no cubre los ids que usa la
fuente. Este intento no ha tocado la fila de `SOURCES.md`, `terminos.go`, el manifiesto, las grabaciones ni las
referencias, ni ha usado la red o `KITLEGAL_RECORD`.

### Lo que deja escrito (sin commitear: el workflow no commitea una tarea redelimitada)

- `internal/source/boe/casos_test.go` (nuevo): la carpeta de las grabaciones, el manifiesto y las referencias; los
  recursos del manifiesto y la petición de cada uno; la lista cerrada de los artículos del diff (tres fijos, dos
  candidatos y dos suplentes, contrato §3.1 y §5); los tipos de la referencia con sus ocho campos; y las ayudas de lectura
  por `filepath.Clean`, sin claves desconocidas ni nada detrás del valor.
  - `TestGrabacionesCompletas`: **22 de 22 en verde**. `httpx.Replay` sirve cada entrada del manifiesto con la dirección de
    `direcciones.go` y `busqueda.go`.
  - `TestReferenciasCompletas`: el subtest de las referencias reales **en rojo**
    (`open testdata/referencias: no such file or directory`). Los 18 subtests que demuestran sobre carpetas temporales
    que la comprobación no pasa en vacío están en verde: las del contrato, con los dos suplentes, sin carpeta, cuatro,
    seis, sin la de `a21`, de dos normas, fuera de la lista, fichero ajeno, clave desconocida arriba y en un campo, de otro
    artículo, grabación inexistente o con carpeta, sin campos, sin un campo, sin valor y `boe_py` vacío.
- `internal/source/boe/ids_test.go`: `TestGramaticaCubreLosIndicesGrabados` **en rojo** en los tres índices (detalle en
  el bloqueo 2). Decodifica con el decodificador JSON genérico el cuerpo que sirve `httpx.Replay`, toma cada clave `id`
  bajo `data` y exige al menos un id por índice.
- `internal/source/boe/terminos_test.go`: `TestFuenteCoincideConSources` ya no admite `pendiente`, y `pendiente` con
  cero o con fecha es ahora un error de tabla. **16 de 16 en verde** contra la fila real, revisada el 2026-09-13 e igual a
  `terminos.go`.

### Verificación

- `make ci`: exit 2. `fmt-check` en verde; `lint` en rojo solo por `dupl`, entre `casos_test.go` 361-411 y
  `grabacion_test.go` 147-193 (ver «Redelimitación»).
- `make test`: todos los paquetes en verde salvo `internal/source/boe`, que falla solo en `TestReferenciasCompletas` y
  `TestGramaticaCubreLosIndicesGrabados`.

### Lo confirmado en la pausa de T008 (commit 7dff6de), comprobado sobre las grabaciones

| Comprobación | Resultado |
|---|---|
| Fila de la fuente y `terminos.go` | ✓ revisada el 2026-09-13; ritmo `1s`; términos en `aviso_legal/index.php` |
| Una grabación por entrada del manifiesto | ✓ 22 de 22 (`TestGrabacionesCompletas`) |
| S2: lo inexistente responde 404 | ✓ bloque `a9999`, y metadatos, índice y análisis de `BOE-A-2099-99999` |
| S1: formato por `Accept` | ✓ bloques en `application/xml`; búsquedas, índices, metadatos y análisis en 200 con `application/json` (los 404 de recursos JSON llegan con `application/xml`: da igual, porque el 404 se clasifica antes de leer el cuerpo) |
| S4: XML | ✓ raíz `response`, sin espacio de nombres, UTF-8 (bloque `a21`) |
| S7: una versión y varias | ✓ entre lo grabado: LPAC `a1` con 1 `version`, LRBRL `a22` con 9 |
| Referencias | ✗ no existe `internal/source/boe/testdata/referencias/` |
| Artículos LCSP del diff | ✗ `BOE-A-2017-12902` `a118` y `da3` responden 404 («La información solicitada no existe») |
| Gramática frente a los índices | ✗ ver bloqueo 2 |

### Bloqueo 1 (humano): faltan las cinco referencias

El commit de la pausa trae la fila, `terminos.go` y las 22 grabaciones, pero no las referencias que el contrato §3.2
(paso 7) y §5 piden escribir a mano desde `refs/boe.py` antes de confirmar. El ejecutor no las crea, no las completa y no
las ajusta (FR-116, research D12). Qué artículos llevan depende de los bloqueos 2 y 3.

### Bloqueo 2 (humano: conflicto con el spec): la gramática del id de bloque no cubre los ids reales

| Índice grabado | Ids | Fuera de `^[A-Za-z0-9]{1,64}$` | Caracteres ajenos | Ejemplos |
|---|---|---|---|---|
| `BOE-A-2015-10565` (LPAC) | 195 | 23 | `-` | `ci-2`, `s1-2`, `da-2`, `da-3` |
| `BOE-A-1985-5392` (LRBRL) | 231 | 38 | `-` `.` | `a7-2` = «Artículo 70 quater», `a85bis.` = «Artículo 85 bis.», `a85ter.`, `primera-2` |
| `BOE-A-2017-12902` (LCSP) | 557 | 523 | `-` | `a1-29` = «Artículo 117», `a1-30` = «Artículo 118», `da-3` = «Disposición adicional tercera» |

- **Consecuencia.** Con la gramática fijada en T005, `kitlegal boe articulo BOE-A-2017-12902 a1-30` daría código 2: casi
  toda la LCSP y parte de la LPAC y de la LRBRL quedarían fuera de `articulo` y `articulos`. Eso choca con el criterio del
  propio FR-080: «se aceptan los ids formados como los del índice de la fuente».
- **Por qué no lo decide el ejecutor.** FR-080 concreta esa forma como «letras y dígitos». data-model §5 fija
  `^[A-Za-z0-9]{1,64}$`. `TestValidarBloque` exige código 2 para `a-21` (caso «guion») y para `.` (caso «punto»). research
  D9 lo prevé: «si hiciera falta un carácter que FR-080 no admite, es un conflicto con el spec y se escala». Además,
  `ids.go` no está entre las rutas de esta tarea.
- **Recomendación.** Enmendar FR-080 para admitir `-` y `.`, que son caracteres no reservados del RFC 3986 §2.3: nunca
  se escapan y no pueden alterar la ruta, la consulta ni el fragmento. Por ejemplo, `^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$`:
  exigir que empiece por letra o dígito excluye los segmentos `.` y `..`.
  - Las propiedades de `FuzzIDDeBloque` siguen valiendo: `url.PathEscape(id) == id`, ninguno de `/ ? # %`, como mucho 64
    bytes, y la ruta termina en `/<id>`.
  - `TipoDesdeID` ya clasifica estos ids sin cambios: `a1-30` es artículo, `da-3` disposición adicional, `ci-2` capítulo
    y `s1-2` sección.
  - Arrastra: FR-080 y su entrada de *Assumptions*, data-model §5, research D9, la «forma esperada» del contrato
    errores-y-codigos (fila 3), la entrada 28 de `doc.go` si nombra la gramática, e `ids.go` e `ids_test.go` (casos
    «guion» y «punto») en la tarea que la decisión declare.
  - **Alternativa rechazada:** mantener la gramática, que deja fuera de la herramienta la mayor parte de la LCSP.

### Bloqueo 3 (humano: artículos del diff de aceptación): `a118` y `da3` de la LCSP no existen

- **Por qué no existen.** En la fuente, esos artículos se llaman `a1-30` (artículo 118) y `da-3` (disposición adicional
  tercera). Por eso las grabaciones 12 y 13 del manifiesto son dos 404.
- **Por qué los suplentes no bastan.** Los suplentes de §3.1 (LPAC `a5`, LRBRL `a1`) solo sustituyen a los recursos 7 u
  11. Sin la LCSP, los candidatos y los suplentes son de dos leyes, y FR-116 y SC-001 piden cinco artículos de tres.
- **Opción A (recomendada, requiere el bloqueo 2 resuelto como se recomienda).**
  - Sustituir `a118` → `a1-30` y `da3` → `da-3` en el manifiesto (recursos 12 y 13), en el contrato
    esquemas-fixtures-y-controles (§2, los golden `articulo-BOE-A-2017-12902-a118` y `-da3`; §3.1 y §5), en research
    D12 y en la lista cerrada `articulosFijosDelDiff` de `casos_test.go`.
  - Volver a grabar esos dos recursos y escribir las cinco referencias.
- **Opción B (sin ampliar la gramática).** Una tercera ley cuyos ids casen con la gramática actual (por ejemplo
  `BOE-A-1992-26318` `a42`, ya grabado, con dos versiones y derogada) y una disposición con id alfanumérico. Deja sin
  resolver el defecto del bloqueo 2.

### Redelimitación hecha en este intento (decidible por el ejecutor)

- **El problema.** `TestGrabacionesCompletas` necesita construir las mismas peticiones que `TestGrabarFixtures`, pero
  `grabacion_test.go` (etiqueta `grabacion`) no estaba declarado. Copiar su lector y su `switch` hace saltar `dupl`, que
  se ejecuta con esa etiqueta, y además dejaría dos construcciones que pueden divergir.
- **La redelimitación.** La línea de T009 declara ahora `internal/source/boe/grabacion_test.go`. El arnés conserva
  `raizDeGrabacion`, `valorQueGraba`, `TestGrabarFixtures` y un `peticionesDeGrabacion(ruta)`, que lee con
  `leerManifiesto` y construye todas las peticiones con `recursoDelManifiesto.peticion()` antes de pedir ninguna.
  Retira `manifiestoDeGrabacion`, `recursoDeGrabacion`, `peticionesDelManifiesto`, `peticionDeGrabacion`,
  `rutaDelManifiesto` (se usa `ficheroDelManifiesto`), `aceptaXML`, `aceptaJSON` y las constantes `recurso*`.
- **Cómo se comprobó**, sobre una copia desechable del árbol:
  - `golangci-lint run ./internal/source/boe/...` sin hallazgos (primero saltó `comparte` en un comentario, reescrito);
  - `go vet -tags grabacion` y los tests con esa etiqueta, en verde.
- **Rutas extraídas.** Con la expresión del workflow (`workflow.yml` 615), la línea declara solo
  `internal/source/boe/casos_test.go`, `internal/source/boe/grabacion_test.go`, `internal/source/boe/ids_test.go`,
  `internal/source/boe/terminos_test.go`, `busqueda.go`, `direcciones.go`, `grabaciones.json`, `SOURCES.md`,
  `terminos.go` y `gates/tarea-T009.md`. Ninguna ruta completa de lo que fija la persona.

### Para quien decide y para el intento siguiente

- **Decidir antes de reanudar.** Mientras los bloqueos 1-3 sigan abiertos, los intentos 2 y 3 llegarán a esta misma
  parada.
- **Commit de la pausa.** Si se confirman datos a mano, no arrastrar con `git add -A` los tres ficheros de test sin
  commitear a un commit de `[datos]`. Quedan en el árbol para el intento siguiente.
- **Qué hará el intento siguiente**, con las decisiones tomadas y la línea ajustada a ellas:
  - aplicar el cambio de `grabacion_test.go` descrito arriba;
  - si cambian los artículos del diff, actualizar la lista cerrada de `casos_test.go`;
  - dejar `TestGramaticaCubreLosIndicesGrabados` y `TestReferenciasCompletas` en verde sin tocar los datos;
  - `make ci` en primer plano y marcar `[X]` en el mismo turno.

## Intento 2 (2026-09-13)

**En verde y marcada `[X]`.** Los tres bloqueos del intento 1 los resolvió la persona en el commit `17fca5b`
(`[datos]`): la gramática del id de bloque admite `-` y `.` (FR-080, data-model §5, `ids.go` e `ids_test.go`), el diff
pasa a seis artículos de cuatro leyes (`a1-30` y `da-3` en vez de `a118` y `da3`; `BOE-A-1992-26318 a42` como cuarto
fijo, el único con avisos reales), las seis referencias están escritas, y la fila de `SOURCES.md` y `terminos.go` dicen
`2026-09-13`. Ese commit arrastró además los tres ficheros de test del intento 1 y la redelimitación de
`grabacion_test.go`, así que este intento encontró la tarea casi hecha y no ha tocado la fila, `terminos.go`, el
manifiesto, las grabaciones ni las referencias, ni ha usado la red o `KITLEGAL_RECORD`.

### Comprobado sobre el árbol, sin cambiar datos

| Comprobación | Resultado |
|---|---|
| Fila de la fuente y `terminos.go` | ✓ `2026-09-13` en los dos; `pendiente` ya no se admite (`pendiente-con-cero` y `pendiente-con-fecha` son errores de tabla) |
| Una grabación por entrada del manifiesto | ✓ 22 de 22 servidas por `httpx.Replay` con la petición de `recursoDelManifiesto.peticion()` |
| S2: lo inexistente responde 404 | ✓ estado `404` grabado en el bloque `a9999` y en metadatos, índice y análisis de `BOE-A-2099-99999` |
| Referencias | ✓ seis ficheros de cuatro normas, entre ellos `BOE-A-2015-10565-a21.json`; ocho campos y `boe_py` no vacío en cada uno; sus grabaciones existen |
| Gramática frente a los índices | ✓ los tres índices grabados (LPAC, LRBRL, LCSP) casan enteros con `ValidarBloque` |
| Arnés de grabación | ✓ `go vet -tags grabacion` en verde; `grabacion_test.go` usa `leerManifiesto` y `peticion()` de `casos_test.go` |

### Lo único que cambia este intento

- `internal/source/boe/casos_test.go`: tres comentarios que seguían diciendo «cinco referencias», «tres fijos» y «tres
  normas» cuando las constantes ya decían 6, 4 y 4 (los de `TestReferenciasCompletas`, `comprobarReferencias` y
  `comprobarArticulosDelDiff`). Ningún cambio de código.

### Verificación

- `go test -count=1 ./internal/source/boe/`: ok. Sondas positivas con `-v`: `TestGrabacionesCompletas` 22 subtests,
  `TestReferenciasCompletas` 19 (la real y 18 sobre carpetas temporales), `TestGramaticaCubreLosIndicesGrabados` 3 y
  `TestFuenteCoincideConSources` 16 (la real y 15 en memoria), todos `PASS`.
- `golangci-lint` fijado por el repo sobre el paquete: 0 hallazgos (incluido `dupl`, el rojo del intento 1).
- `make ci` en primer plano: exit 0, «todos los controles en verde» (`gates/ci.log`; el paquete `boe` corre sin caché en
  la primera pasada, con cobertura 100 %).

### Observaciones para la persona (fuera de las rutas de esta tarea; no se tocan)

- El commit `17fca5b` versionó `refs/__pycache__/boe.cpython-311.pyc` (binario que dejó el guion de un solo uso de las
  referencias al importar `refs/boe.py`), y `.gitignore` no excluye `__pycache__/`. Conviene retirarlo del índice y añadir
  el patrón en un commit propio.
- El checkpoint de la fase 3 de `tasks.md` (tras T010) sigue diciendo «las cinco referencias»; el texto de T009 ya dice
  seis. No es una tarea, así que este intento no lo toca.
