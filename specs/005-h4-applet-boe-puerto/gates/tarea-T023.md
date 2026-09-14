# T023 · La caché de los seis verbos

## Intento 1 (2026-09-13)

**Marcada `[X]`.** La tarea no añade producto: el único fichero tocado fuera del directorio del feature es
`internal/source/boe/fuente_test.go`.

### Lo que añade

- Una tabla de consultas, `consultasDeLosSeisVerbos`, con los seis verbos sobre la Ley 39/2015 y la búsqueda de
  «procedimiento administrativo común», más la búsqueda sin resultados, que se guarda y se sirve como cualquier otro
  resultado (FR-032, SC-003). La de `articulos` pide `a21 a22 a21 a23`: un id repetido y tres bloques distintos que
  comparten una sola petición de metadatos. Cada consulta lleva su url, su vigencia, las peticiones que emite sin caché
  y su fallo con `--offline`.
- `TestCacheDeLosSeisVerbos` (9 subtests). Primera invocación sobre las grabaciones, con sus peticiones exactas. Después,
  con el reloj en el último nanosegundo de la vigencia, otra invocación sobre la misma caché con `httpx.Replay` sobre
  una carpeta vacía: cero clientes construidos, cero peticiones y el mismo `Resultado` (data, url y fecha). Al cumplirse
  la vigencia, una tercera invocación vuelve a pedir las mismas peticiones y lleva la fecha nueva. Además,
  `articulo-con-los-metadatos-en-cache` (solo el bloque, con la fecha de los metadatos) y `metadatos-tras-un-articulo`
  (ninguna petición).
- `TestOfflineDeLosSeisVerbos` (7 consultas × `--offline` y `--offline --dry-run` × vigente, caducada y ausente). Vigente
  da lo guardado con su fecha; caducada y ausente dan el fallo de la fila 5. En los tres casos no se pide nada y la
  caché queda intacta.
- `TestEnsayoDeLosSeisVerbos` (7 consultas × `sin-cache` y `con-la-entrada-caducada`). Sale el resultado exacto del
  ensayo, una línea por petición y ninguna repetida, cada petición entregada en ensayo y nada escrito. Con la entrada
  caducada, además, `--offline` sigue sin servirla.
- `TestFallosNoSeGuardan` (bloque inexistente, fuente caída, bloque ilegible y metadatos caídos). Dos consultas
  seguidas fallan con su clase, sin datos, con la url de la petición que falla y su propio instante, y la segunda
  vuelve a entregar las mismas peticiones. Con `--offline` no hay entrada.
- Auxiliares en el mismo fichero: `resuelvePidiendo`, `trasGuardar`, `compruebaEnsayo`, `otroSobreLaMismaCache`,
  `compruebaSinPedirNada` y `baseDeLaCache`.

### Decisión: qué es «la caché intacta»

La primera versión comparaba todos los ficheros de la carpeta de la caché antes y después, y fallaba en todos los casos
con caché. Una lectura en solo lectura crea `cache.db-wal` (vacío) y `cache.db-shm`, mientras `cache.db` queda idéntico
byte a byte. Es lo que fija H3: `specs/004-h3-internal-cache-sqlite/contracts/esquema-y-apertura.md` §6 (garantía de
SC-003: «`cache.db` es idéntico byte a byte […]; `-wal` y `-shm` pueden aparecer y no cuentan») y §7. Por eso
`baseDeLaCache` compara solo `cache.db`, y exige que exista y no esté vacío.

### Rojo → verde

Son tests sobre código ya existente, la excepción que declara `tasks.md`. Para comprobar que no pasan en vacío, se
aplicaron mutaciones del producto por tandas en el árbol de trabajo. Cada tanda se ejecutó con los cuatro tests y se
retiró antes de la siguiente. Al terminar, `grep SONDA-T023` no encuentra nada y `git diff` solo muestra `fuente_test.go`
y los ficheros del workflow.

- **Tanda 1**, en `articulo.go`:
  - A escribe el artículo sin avisos tras fallar los metadatos, la mutación de la fila 8 del inventario de `plan.md`. La
    detecta `TestFallosNoSeGuardan/metadatos-caidos`.
  - D quita la memoria de los metadatos de la invocación. La detecta `TestEnsayoDeLosSeisVerbos/articulos` en sus dos
    subtests: la línea de los metadatos sale repetida.
- **Tanda 2**, en `fuente.go`: B no sirve nunca la entrada vigente. Caen los 9 subtests de `TestCacheDeLosSeisVerbos` y
  los 14 `…-vigente` de `TestOfflineDeLosSeisVerbos`.
- **Tanda 3**, en `fuente.go`:
  - C hace que `--offline` sin entrada pida. Caen los 28 `…-caducada` y `…-ausente` de `TestOfflineDeLosSeisVerbos` y
    los 4 de `TestFallosNoSeGuardan`, cuya comprobación con `--offline` pasa a pedir.
  - E abre la caché en modo normal con `--dry-run`. Caen los 7 `sin-cache` de `TestEnsayoDeLosSeisVerbos`; esos
    subtests no llevan `--offline`, así que C no los alcanza.
  - Los 7 `con-la-entrada-caducada` de `TestEnsayoDeLosSeisVerbos` también caen, al menos por C.

### Verificación de este intento

- Los cuatro tests: 87 `--- PASS` (4 tests y 83 subtests), ningún `FAIL` ni `SKIP`. También en verde con `-race`.
- `golangci-lint run ./internal/source/boe/...`, con el binario fijado por el repo: 0 hallazgos.
- `make ci` en primer plano: «ci: todos los controles en verde». Incluye lint, tests con `-race` en los dos perfiles,
  `govulncheck`, `gitleaks`, módulos y `tidy -diff`.
