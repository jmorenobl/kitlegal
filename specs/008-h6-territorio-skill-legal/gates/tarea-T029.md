# T029: carga del territorio dentro de su presupuesto de tiempo

## Intento 1 (2026-09-22)

### 1. Medida antes de cambiar nada (sobre `47f3090`)

El control nuevo, `TestCosteDeLaCarga` (`internal/core/territorio/coste_test.go`), carga unas fuentes sintéticas del
tamaño real: 8 132 filas en la relación y 8 132 en la correspondencia, escritas como las congeladas (research.md D4, una
fila por línea), con las 52 provincias repartidas entre 19 comunidades y nombres inventados que mezclan bilingües,
artículos pospuestos, diacríticos y apóstrofos. Mide asignaciones y bytes, no tiempo: la media de cuatro cargas tras una
que no cuenta, con un solo procesador lógico, como `testing.AllocsPerRun`, y en un test que no corre en paralelo con
ningún otro.

Con la carga de `47f3090` y el control recién escrito, tres medidas de cada modo:

| Modo | Asignaciones por carga | Bytes por carga |
|---|---|---|
| `go test` | 779 065 · 780 793 · 779 155 | 37 455 908 · 37 465 116 · 37 456 372 |
| `go test -race -shuffle=on` | 779 624 · 778 160 · 781 271 | 40 567 528 · 40 544 104 · 40 593 888 |

Es del orden de lo medido con los datos reales en la nota de T026 (756 111 asignaciones y 36 444 020 bytes). Los
máximos del control son un tercio de la menor medida sin `-race`, la más estricta: **259 688 asignaciones** (779 065 / 3)
y **12 485 302 bytes** (37 455 908 / 3). Las medidas varían en torno a un 0,2 % de una ejecución a otra, por la semilla
de dispersión de los mapas; con el margen de un tercio, el veredicto no depende de eso.

Rojo, antes de cambiar el código: `"780940" is not less than or equal to "259688"` y
`"37465888" is not less than or equal to "12485302"`.

### 2. Arreglo, en `internal/core/territorio/`

- **Las filas en su forma se leen sin el lector de YAML** (`filas.go`). Un fichero de filas está en su forma si su
  cabecera llega hasta una línea que es exactamente `municipios:` (o `correspondencia:`), el lector de YAML la decodifica
  sin defectos y ve esa clave sin valor en esa misma línea y columna, y cada línea que sigue hasta el final es una fila
  `  "<clave>": {dc: "…", nombre: "…", provincia: "…", comunidad: "…"}` (o `  "<clave>": "…"`). Dentro de cada texto
  entre comillas dobles: los únicos escapes son `\"` y `\\`, y cada carácter es de los que el lector admite y no cuenta
  como salto de línea (ni el retorno de carro, ni NEL, ni U+2028 ni U+2029); la clave no pasa de 1 022 bytes, el límite
  del lector para una clave implícita. **Cualquier otro fichero lo lee entero el lector de YAML**, con el mismo código
  que antes (`decodificarConElLector`), así que sus defectos son los mismos. Los dos caminos comparten la búsqueda de la
  clave repetida (`filasLeidas`), que da el mismo texto con las mismas líneas.
- **Ordenar el registro ya no escribe cada código en cada comparación** (`registro.go`). Con solo el primer cambio, el
  control seguía en rojo, con 304 701 asignaciones: la ordenación de `indexar` llamaba a `CodigoINE.String()`, que
  concatena y asigna, dos veces por comparación (unas 77 000 asignaciones por carga). Ahora `indexar` indexa por código,
  escribiendo cada código una sola vez, y recorre las claves de ese índice en orden. El contrato no cambia: los
  candidatos de cada forma salen en orden de código venga como venga la lista (`TestRegistro`, «candidatos en orden de
  código», sigue indexando la lista al revés). `compararPorCodigo` queda sin uso en producción y la prueba compara ella
  misma.
- El lector de una fila recibe el texto de la fila y no un puntero a un lector, que escapaba al montón en cada línea
  (unas 16 000 asignaciones por carga).

### 3. Después

| | Asignaciones por carga | Bytes por carga |
|---|---|---|
| Control sintético, `go test` | 66 571 | 10 156 876 |
| Control sintético, `go test -race -shuffle=on -coverprofile` | 66 576 | 10 248 096 |
| Datos congelados reales (prueba temporal, borrada) | 63 333 (antes 756 111) | 9 805 428 (antes 36 444 020) |

Con los datos reales, un benchmark temporal da 9,0 ms por carga (antes 42 ms, nota de T026), y el binario de `make build`
tarda 0,017-0,018 s en `territorio resolver 28074 --json` en caliente (antes, unos 0,05 s), con la misma salida byte a
byte: la huella `sha256:b730a9076336bd84fa84c91f3efe59c462ee589c68432c06971b59dcd7a8d570` es la de la ejecución `ci`
35714659927. La prueba temporal (`zz_real_tmp_test.go`, fuera de `make ci` y borrada antes de él) comprobó además que los
dos ficheros congelados están en su forma y que la carga da con ellos exactamente lo que da el lector de YAML.

### 4. Mismo resultado y mismos defectos

- Los tests existentes siguen como estaban, con los mismos nombres de subtest que fija el inventario del plan; la tabla
  de `TestCargar` se extrajo a `casosDeCarga()` para que `TestCargarEnFilas` repita cada caso con la relación y la
  correspondencia en su forma y exija los mismos defectos. `TestCargar/clave-repetida` ya estaba escrito en la forma de D4
  y pasa ahora por el camino nuevo con el mismo texto y las mismas líneas.
- `TestFormaDeLasFilas`: 50 ficheros escritos a mano, 43 de relación y 7 de correspondencia, en su forma y fuera de ella (escapes, clave de 1 022 y de 1 023
  bytes, retorno de carro en la cabecera y en las filas, dos documentos, cabecera ilegible, NEL, U+2028, U+2029,
  caracteres de control, bytes que no son UTF-8…). Para cada uno exige si está en su forma y que la carga dé exactamente
  lo que daba la anterior —el mismo fichero y el mismo texto de defecto—, y lo que la carga lee sin defectos, igual que
  `yaml.Unmarshal` por su cuenta.
- `TestNombresDificiles` y `FuzzFilasComoElLector`: nombres con comillas dobles y simples, tildes, diéresis, eñe, barras
  y barras invertidas, apóstrofos, punto volado, espacios, indicadores de YAML, textos que sin comillas no lo serían,
  tabuladores, la marca de orden de bytes y runas de todos los tramos, como clave y como texto; se leen en su forma y dan
  lo mismo que el lector.
- Mutantes a mano, uno a uno y deshechos: aceptar la clave sin comprobar su línea, admitir los separadores de Unicode,
  subir la cota de la clave a 1 023 bytes, admitir cualquier escape, correr el número de línea y no deshacer los escapes.
  Todos caen. Excluir el tabulador y la marca de orden de bytes resultó innecesario —el lector los toma tal cual, y la
  regla quedó en «admitido y no salto de línea»—.
- Cobertura del paquete: 100,0 %.

### 5. Qué no se tocó

Ni el máximo de 200 ms, ni la orden `cronometra`, ni el guion e2e, ni los ficheros congelados, ni sus esquemas, ni el
kernel, ni el applet; ninguna dependencia nueva. research.md S3 recoge la medida y D4, que la carga se apoya en esa forma.
El e2e de la integración continua no se puede ejecutar desde aquí: lo mide el intento 2 de T026 al empujar la rama.

### 6. Verificación

`make ci` en primer plano, en verde: lint con `0 issues.`, `go test -race -shuffle=on` y la integración
(`-tags=integration`, con el guion `territorio-matriz.txtar` y su `cronometra` de 200 ms) en `ok` en todos los paquetes,
`internal/core/territorio` con `coverage: 100.0% of statements` en los dos, govulncheck, controles de esquemas y datos,
gitleaks, `go mod verify` y `go mod tidy -diff`; última línea: `ci: todos los controles en verde`. **Estado: marcada
[X].**
