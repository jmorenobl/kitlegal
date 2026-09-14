# T024 · Ninguna ruta de fallo sin clase, sobre el error real

## Intento 1 (2026-09-13)

**Marcada `[X]`.** La tarea no añade producto. Fuera del directorio del feature solo cambia
`internal/source/boe/errores_test.go`.

### Lo que añade

- `TestClasesDeErrorDeBoe`: una subprueba por cada fila 2-4, 6-10 y 16-22 del contrato errores-y-codigos, con 56
  situaciones en total. Antes de las subpruebas, la prueba exige que la lista de filas sea exactamente esa, de modo
  que una fila no puede desaparecer en silencio.
- Cada situación se provoca con el código ya existente y pasa por un único comprobador, `casoDeFallo.comprueba`, que
  exige sobre el error real:
  - que provocarla no entra en pánico (`require.NotPanics`);
  - la clase de `cli.Clasificar` y el código de `cli.CodigoSalida`, y que el código no es 6;
  - un fragmento del mensaje que solo da esa situación, para que el caso no pase por otro fallo de la misma clase
    (por ejemplo, una grabación ausente);
  - el sobre de la fila. Sin dirección, el resultado va a cero y lo firma el kernel. Con ella, la procedencia lleva
    `boe.legislacion-consolidada`, la dirección del recurso y el instante de la petición que falló, o ninguno si no
    hubo petición, y entonces la fecha la pone el montaje.
- Situaciones por fila:
  - Filas 2-4: validadores sobre los cinco verbos con norma (2), el bloque de `articulo` y el último de `articulos` (3)
    y la búsqueda vacía y la de solo espacio en blanco (4). En todas, ni se abre la caché ni se construye el cliente.
  - Filas 6-8: grabaciones de las normas inexistentes (6, 7) y `data` vacío de un `Pedidor` de prueba (8), para
    índice, metadatos y análisis.
  - Filas 9-10: `Pedidor` de prueba con 404 en `buscar` y con 400, 401, 403, 405, 406 y 410 repartidos entre los
    verbos. El 410 va sobre el índice de una norma inexistente: solo el 404 es «no encontrado».
  - Fila 16: los sintéticos `bloque-ilegible` y `bloque-sin-elemento`, y cuerpos no UTF-8 y con dos raíces.
  - Fila 17: el sintético `metadatos-ilegibles`, una raíz que es lista (índice), un primer elemento que es cadena
    (análisis) y un número donde se espera texto (búsqueda).
  - Fila 18: `articulo` a21 con los metadatos fallando con cada clase posible. Con 3, `data` vacío; con 4, el
    sintético `metadatos-caidos`; con 5, el sintético `limite`.
  - Fila 19: `articulos`, con el 404 del segundo bloque (3), un 403 del segundo bloque (4) y el 429 de los metadatos
    del primero (5).
  - Fila 20: una entrada vigente escrita con la caché real bajo la clave de los metadatos, que no se puede leer. No se
    construye cliente ni se pide nada.
  - Fila 21: ver la decisión de más abajo; son 15 situaciones.
  - Fila 22: `Nueva` sin dependencias, una `Fuente` construida a cero y una consulta de otro tipo.
- Auxiliares en el mismo fichero: `rechazada`, `pedida`, `salvoEn` (reproduce las grabaciones salvo una dirección),
  `conLaEntradaIlegible`, `sinDependencias`, `sinComponer`, `casosDeLaFila21`, `casoDeLaCache`, `cacheQueFalla` y
  `falloDeLaCacheDePrueba`.

### Decisión: la fila 21

- **Invocación.** `metadatos BOE-A-2015-10565` sobre las grabaciones, sin `--offline` ni `--dry-run`, con una
  `AperturaDeCache` de prueba. Su caché, `cacheQueFalla`, guarda en memoria lo que se escribe y falla al leer, al
  escribir o al cerrar. Su error, `*falloDeLaCacheDePrueba`, implementa `schema.ConClase` y se usa por puntero, de
  modo que `errors.Is` lo alcanza por identidad.
- **Casos.** Los cuatro de la tarea se prueban con cada una de las tres clases de `cache.Error` (argumentos, fuente no
  disponible e inesperado), y el del cierre en sus dos variantes: son cinco situaciones por clase, quince en total.
- **Qué exige cada situación.** Además del comprobador común:
  - la caché se abre una sola vez y fuera de solo lectura, y se cierra si se abrió;
  - el número de peticiones: ninguna si falla la apertura o la lectura, y una si falla la escritura o el cierre tras
    una invocación correcta;
  - las entradas escritas: ninguna, salvo en el cierre tras una invocación correcta, cuyo `Put` sí se hizo;
  - que cada error de la caché siga alcanzable con `errors.Is`.
- **Cierre tras una invocación fallida.** El fallo de la invocación es el de la lectura, y no uno de HTTP: la tarea
  exige la fecha sin declarar en todos los casos, y un fallo de HTTP lleva el instante de su petición. El cierre falla
  con otra clase, para que se vea que prevalece la de la invocación, y los dos errores tienen que ser alcanzables con
  `errors.Is`.
- **La lectura con «fuente no disponible».** Fuera de solo lectura es un fallo y no una ausencia. Es la única
  situación en la que un error mal clasificado haría pedir a la fuente (sonda M4).

### Rojo → verde

Son tests sobre código ya existente, la excepción que declara `tasks.md` para T024. Para comprobar que no pasan en
vacío, un guion de un solo uso fuera del repositorio (`/tmp/sondas-t024.pl`, sin versionar) aplicó doce mutaciones
del producto. Las aplicó de una en una: tras cada una ejecutó `TestClasesDeErrorDeBoe` y restauró el fichero con
`git checkout`. Al terminar no queda ninguna marca `SONDA-T024` en `internal/`, y `git status` solo muestra
`errores_test.go` y los ficheros del workflow.

| Sonda | Mutación | Casos en rojo |
|---|---|---|
| M1 | `claseDelFalloDeLaCache` da identidad humana en vez de inesperado | los 5 de la fila 21 con «inesperado» |
| M2 | `invocar` ignora el fallo del cierre | los 6 de cierre de la fila 21 |
| M3 | `guardar` ignora el fallo de `Put` | los 3 de escritura de la fila 21 |
| M4 | el fallo de `Get` fuera de solo lectura se toma por ausencia | los 2 de la fila 21 con lectura «fuente no disponible» |
| M5 | el 404 de la búsqueda es «no encontrado» | fila 9 |
| M6 | la entrada ilegible se toma por ausencia | fila 20 |
| M7 | el sobre de fallo pierde el instante de la petición | los 28 de las filas 6-10 y 16-19 |
| M8 | el defecto al componer es «argumentos» | fila 22: `nueva-sin-dependencias` y `fuente-sin-componer` |
| M9 | `buscar` no rechaza la búsqueda sin palabras | los 2 de la fila 4 |
| M10 | `articulo` no valida el bloque | fila 3, `articulo` |
| M11 | la raíz JSON que no es objeto se lee como vacía | fila 17, `raiz-que-no-es-objeto` |
| M12 | `invocar` no comprueba que la fuente se compuso con `Nueva` | fila 22, `fuente-sin-componer`, por pánico |

### Verificación de este intento

- `TestClasesDeErrorDeBoe` con `-race`: 72 `--- PASS` (1 test, 15 filas y 56 situaciones), ningún `FAIL` ni `SKIP`.
- `golangci-lint run ./internal/source/boe/...` con el binario fijado por el repo: 0 hallazgos. El primer intento del
  lint pidió formato (gofumpt): `clase` y `codigo` iban en la misma línea y partían la alineación. Se separaron en dos
  líneas en toda la tabla.
- `make ci` en primer plano, con salida 0 y «ci: todos los controles en verde». Incluye lint, tests con `-race` en los
  dos perfiles, `govulncheck`, `gitleaks`, módulos y `tidy -diff`. Cobertura de `internal/source/boe`: 99,4 %.
