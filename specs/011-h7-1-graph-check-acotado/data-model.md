# Data model: H7.1 · `graph check` acotado, lecturas que apagan `version-obsoleta` y H7 sin lo que no pasa el umbral

Lo que el hito añade o cambia en el modelo de H7 (`specs/010-h7-internal-graph-grafo/data-model.md`, que no se edita y
sigue valiendo en todo lo demás). La forma exacta de lo que se ve desde fuera está en [contracts/](./contracts/); las
decisiones, con su alternativa rechazada, en [research.md](./research.md).

## 1. Lecturas de un bloque (`world.db`, migración `0002_lecturas.sql`)

Una **lectura** es la llegada al grafo de un bloque devuelto por `boe articulo` o `boe articulos` que entrega (FR-020).
El grafo no guarda la historia de las lecturas: guarda, por bloque, las dos únicas que deciden algo —qué redacción vio
la última y cuál la anterior— (research D1).

```sql
-- internal/graph/migraciones/0002_lecturas.sql (versión 2 del esquema)
CREATE TABLE lecturas (
    bloque   TEXT PRIMARY KEY NOT NULL REFERENCES nodes(id), -- el Bloque leído
    ultima   TEXT NOT NULL REFERENCES nodes(id),             -- la BloqueVersion que vio su última lectura
    anterior TEXT NOT NULL REFERENCES nodes(id)              -- la que vio la lectura anterior; en la primera, la misma
) STRICT;
```

En la primera lectura de un bloque, `anterior` = `ultima`: no hay cambio que decir, y una entrega repetida idéntica deja
la fila como estaba y no escribe nada (H7 FR 022; `TestApplyIdempotente` y `TestIntegracionIdempotencia` se quedan).

- Una fila por bloque leído desde que existe la tabla: con la medida de la bitácora, 2 400 filas de unos 250 bytes
  (tres ids ELI, dos de ellos con la huella de su texto), unos 600 KB, que no crecen con las preguntas sino con los
  bloques distintos.
- Ningún verbo de `graph` la escribe (FR-014); `graph show` y `graph stats` no la leen (FR-022).
- `world.db` de versión 1 (escrito por H7): la tabla no existe. Se lee sin migrar (§3) y la primera entrega la crea
  dentro de su transacción, con el mecanismo de migraciones de H7 FR 013.

## 2. Qué es una lectura en un lote (`internal/core/grafo`)

```go
// Lectura es la de un bloque en una entrega: el Bloque y la BloqueVersion que vio.
type Lectura struct {
	Bloque  string // Origen de la arista eli:has_version
	Version string // Destino de esa arista
}

// Lecturas son las del lote consolidado: una por arista eli:has_version, en el orden de sus claves (FR-020).
func (c Consolidado) Lecturas() []Lectura
```

- Solo `boe articulo` y `boe articulos` emiten `eli:has_version` (`internal/source/boe/grafo.go:74`, llamado desde
  `articulo.go:112`), una por bloque distinto de la invocación (`observadoDeLosArticulos`, `grafo.go:22-41`, con su
  test de H7 «articulos-con-un-bloque-repetido», `internal/source/boe/grafo_test.go:58`): una lectura por bloque e
  invocación aunque se nombre dos veces (FR-020). Un lote con la misma arista dos veces no lo emite nadie (research
  V29) y no es un caso de `Lecturas`.
- `--no-graph`, `--dry-run`, un código distinto de 0, una respuesta sin ELI o una entrega que falla no llegan a
  confirmar ninguna transacción: no hay lectura (FR-020; H7 FR 031-034, FR 040).

## 3. Lo que lee `graph check` (`grafo.Instantanea`)

```go
type Instantanea struct {
	Nodos    []NodoDeInstantanea // como en H7
	Aristas  []schema.Arista     // como en H7
	Lecturas []LecturasDeBloque  // nuevo: las filas de lecturas de los bloques de la instantánea
}

// LecturasDeBloque es la fila de lecturas de un bloque.
type LecturasDeBloque struct {
	Bloque   string
	Ultima   string // BloqueVersion que vio la última lectura
	Anterior string // la que vio la anterior; en la primera lectura, la misma que Ultima
}

// Leida es la fila tras una lectura que ve version: (version, Ultima).
func (l LecturasDeBloque) Leida(version string) LecturasDeBloque
```

**Redacción vista** de un bloque (entidad «Redacción vigente» del spec):

| Estado del bloque | Redacción vista (`Ultima`) | Vista antes (`Anterior`) |
|---|---|---|
| Con fila en `lecturas` | la de la fila | la de la fila |
| Sin fila (`world.db` de H7, o bloque observado por H7 y no leído desde entonces) | `R`: la `BloqueVersion` suya con la última observación más reciente —instante de `fecha_consulta`; a igualdad, id menor comparando bytes— (`grafo.RedaccionVistaSinLecturas`; FR-026 y supuesto del spec) | `R` (una sola lectura: nada que decir) |

## 4. Transiciones (escritura, `Apply`)

Dentro de la transacción de la entrega, para cada lectura `(b, v)` del lote, **antes** de aplicar sus nodos y aristas,
se calcula la fila nueva con `Leida(v)` sobre la de partida:

- `b` con fila `(U, A)` → `(v, U)`;
- `b` sin fila → `(v, R)`, con `R` la redacción vista sin lecturas de §3 calculada sobre lo guardado **antes** de este
  lote, o `R` = `v` si el grafo no guarda ninguna versión de `b` (un bloque nuevo: `(v, v)`).

Después se aplican nodos, textos y aristas como en H7 y se escribe cada fila **solo si cambia** (el patrón de H7 para
nodos y aristas; la clave ajena exige que existan los nodos). Las entregas se aplican una tras otra (transacción
inmediata, H7 FR 014): «última» y «anterior» son el orden en que se confirman (FR-021).

La secuencia de FR-025, con A = 20161002 (grabación de H4), B = 20250101 (derivada `version-posterior`) y C = 20260101
(derivada `version-ulterior`, nueva):

| Paso | Lectura | Fila de `a21` | `version-obsoleta` |
|---|---|---|---|
| 1 | la fuente sirve A | `(A, A)` | ninguno |
| 2 | caducada la caché, la fuente sirve B | `(B, A)` | sobre A: 20161002 → 20250101 (y el mismo al repetir `check`) |
| 3 | con `--no-graph` | `(B, A)` (no hay lectura) | sobre A, el mismo |
| 4 | la caché sirve B | `(B, B)` | ninguno (tampoco al repetir) |
| 5 | caducada la caché, la fuente sirve C | `(C, B)` | sobre B: 20250101 → 20260101 |

## 5. Reglas de `graph check` (`grafo.Comprobar`)

- **`version-obsoleta`** (FR-023, FR-024): por cada bloque con fila `(U, A)` y las fechas de vigencia de `U` y de `A`
  válidas (H7 FR 064: ocho cifras que nombran un día) con la de `U` estrictamente posterior: un hallazgo
  con `id` = `A`, `fecha_vigencia` = la de `A`, `fecha_vigencia_reciente` = la de `U`, `procedencia` = la última
  observación de `U` y la explicación de H7 (plantilla sin cambios, contracts/applet-graph.md §4). En cualquier otro
  caso, ninguno. Sustituye la regla de `versionesObsoletas` de H7, que conserva el nombre y recibe las filas de
  `lecturas`, y retira `compararRecencia`.
- **`fuente-caducada`** (FR-030): la condición de H7 FR 066 (vigencia declarada y `fecha_consulta + vigencia`
  estrictamente anterior al instante), solo sobre nodos `Norma`, `Bloque` y las `BloqueVersion` que son la redacción
  vista de un bloque de la instantánea (§3). Nunca sobre otra `BloqueVersion` ni sobre otro tipo.
- **Orden** (FR-011): todos los `version-obsoleta` antes que todos los `fuente-caducada`; dentro de cada clase, por
  `id` comparando bytes.
- **Cota** (FR-010): `grafo.MaximoDeHallazgos = 50`. Se listan los primeros 50 del orden; los totales cuentan todos.

## 6. Ámbito (`grafo.Ambito`) y datos de `check` (`grafo.Comprobacion`)

```go
// Ambito es lo que se comprueba: una norma y, si se nombran, bloques suyos; Norma vacía, todo lo consultado.
type Ambito struct {
	Norma   string   // identificador BOE, BOE-A-<año>-<número>; "" = todo lo consultado
	Bloques []string // ids de bloque tal como se pidieron; vacío = todos los de la norma
}

// Comprobacion es el data de graph check (FR-012).
type Comprobacion struct {
	Norma           string     `json:"norma"`            // la pedida; "" sin argumentos
	Bloques         []string   `json:"bloques"`          // los pedidos, en su orden; [] si ninguno (nunca null)
	VersionObsoleta int        `json:"version-obsoleta"` // total de la clase en el ámbito, contando los omitidos
	FuenteCaducada  int        `json:"fuente-caducada"`  // ídem
	Omitidos        int        `json:"omitidos"`         // VersionObsoleta + FuenteCaducada - len(Hallazgos)
	Hallazgos       []Hallazgo `json:"hallazgos"`        // los listados, ≤ 50; [] si ninguno (nunca null)
}

func Comprobar(instantanea Instantanea, ambito Ambito, ahora time.Time) (Comprobacion, error)
```

- `Hallazgo` no cambia (H7 FR 061).
- Qué entra en el ámbito lo decide la lectura acotada de `internal/graph` (contracts/almacen-world-db.md §3): la
  instantánea de una norma solo trae su `Norma` (por `datos.identificador`), sus `Bloque` por `eli:has_part` —los
  nombrados por `datos.bloque`, o todos—, sus `BloqueVersion` por `eli:has_version` y sus filas de `lecturas`.
  `Comprobar` aplica las reglas a todo lo que recibe y copia `Norma` y `Bloques` del ámbito.
- Validación de los argumentos (en el applet, antes de abrir nada; FR-004): la norma dada —el argumento es un
  `*string`, así que una norma vacía se distingue de no darla— con la gramática de `boe.ValidarNorma`; un bloque vacío
  o de solo `unicode.IsSpace`. Una norma o un bloque bien formados que el grafo no
  conoce no son un error (FR-003).

## 7. Etiqueta de `version-obsoleta` (`internal/core/grafo/vocabulario.go`)

```go
// EtiquetasDeHallazgo es la etiqueta de cada clase de hallazgo que una skill traslada con forma fija:
// {ClaseVersionObsoleta: "REDACCIÓN MODIFICADA"}. Única fuente de verdad (FR-045); fuente-caducada no la tiene.
func EtiquetasDeHallazgo() map[ClaseDeHallazgo]string
```

La forma fija es `⚠` + la etiqueta + `:` en la misma línea, con las tolerancias de H5.1 (contracts/evals-y-skill.md §2).
El binario no la escribe en ninguna salida (spec, «Fuera de alcance»).

## 8. Formato común de eval (`internal/evals/formato.go`, `schemas/eval.yaml.json`)

- `Eval.Hallazgos []string` (`hallazgos` en el YAML): clases de hallazgo cuya forma fija tiene que llevar la respuesta;
  valores del enumerado `clase-de-hallazgo` = las claves de `grafo.EtiquetasDeHallazgo()` (hoy, `version-obsoleta`).
  Solo con `activa: true`, como `avisos` (FR-054).
- `ComandoEsperado.Norma` en la forma comprobación (`applet` + `verbo: check`, **y ahora `norma` opcional**): con norma,
  la cumple una invocación de `check` que consulta, termina con 0 y lleva esa norma como primer argumento (FR-050).
- `ResultadoDeEval` gana `hallazgos_encontrados` y `hallazgos_ausentes`; un ausente es motivo
  (`forma de hallazgo ausente: version-obsoleta`) y la sesión no pasa (FR-052).
- `TasaDelInforme` gana `formas`: la forma fija de cada hallazgo que la eval exige (`["⚠ REDACCIÓN MODIFICADA:"]`), y
  `[]` si ninguno (FR-055).

## 9. Lo que sale del modelo (retirada, FR-070 a FR-076)

| Pieza de H7 | Qué pasa |
|---|---|
| `Almacen.enlazar`, `publicar`, `crearElTemporal`, `retirarElTemporal`, `retirarDirectorios`, `puedeRehacerse` | se retiran con `publicar.go` (FR-071); `world.db` se crea en su sitio |
| `modoInmutable`, `comprobarEscritura`, `comprobarLectura`, `rutaDeLosAuxiliares`, `sufijoDiario`, `sufijoMemoriaCompartida`, reapertura inmutable, clasificación 776/1544/14 | se retiran (FR-072); la lectura mira solo si hay `world.db-wal` |
| `errorEsDirectorio`, `errorDeTransaccionInterrumpida`, `errorDeFicheroNoEscribible`, `errorDePublicacion` | se retiran: esos estados son la regla genérica, `errorInutilizable` (FR-070) |
| `ValidarContraGrafoVacio`, comprobación de extremos en `leerArista`, rechazos de URI no absoluto, nodo sin id o sin tipo, arista sin origen, relación o destino (`validarArista`), y el caso de la arista en `nombrar` | se retiran (FR-075); la clave ajena de `edges` rechaza un extremo ausente, y ningún rechazo que se queda lleva una arista (research V29) |
| desempate por fuente, texto de la fecha, vigencia y datos canónicos, también el de un id repetido en el lote | se retira (FR-076); queda la `url` y la observación idéntica; `Consolidar` guarda una vez un id repetido, sin comparar sus datos |
| casos no ASCII de `persona_test.go` | se retiran (FR-074); la regla y sus casos ASCII se quedan |
