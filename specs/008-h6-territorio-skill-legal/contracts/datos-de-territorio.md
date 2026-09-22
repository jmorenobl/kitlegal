# Contrato: los ficheros congelados de `data/territorio/` y su tarea `[datos]`

Qué ficheros hay, qué esquema los valida, quién los genera y qué evidencia deja. Nada de esto se pide por red, ni en
ejecución ni en grabación (ADR 0017, decisión 1; FR-043).

## 1. Ficheros y esquemas

| Fichero | Esquema | Origen de sus datos | Fecha que declara |
|---|---|---|---|
| `data/territorio/municipios.yaml` | `schemas/territorio-municipios.yaml.json` | relación oficial del INE (`ine.municipios`) | la de referencia del fichero del INE |
| `data/territorio/dir3.yaml` | `schemas/territorio-dir3.yaml.json` | REL (`mpt.rel`) + dígito del INE | la del volcado del REL |
| `data/territorio/estado.yaml` | `schemas/territorio-estado.yaml.json` | configuración | la de su redacción |
| `data/territorio/comunidades/<código>.yaml` (19) | `schemas/territorio-comunidad.yaml.json` | nombres del INE; régimen y boletines, configuración | la de su redacción |

La forma de cada uno está en [data-model.md](../data-model.md) §3. Todos llevan `fecha` y `source` en la raíz y una
línea por fila (research.md D4).

## 2. Validación

- **Contra su esquema**: con `skills.ValidarDocumentoYAML[T]`, que rechaza además la clave repetida (research.md V37),
  desde `TestTerritorioDelRepositorio` en `internal/skills` —donde ya viven `TestNormasDelRepositorio` y el lector
  común—, dentro de `make skills-check` y por tanto de `make ci` (FR-044). El tipo `T` de cada fichero es el que
  exporta el dominio (`territorio.FicheroDeMunicipios`…), de modo que no se repiten aquí
  ([data-model.md](../data-model.md) §2.1): `LeerTerritorio(territorio.Fuentes) (territorio.Ficheros, error)` valida
  los cuatro y devuelve lo leído.
- **Todo lo que se afirma del corpus se comprueba aquí**: este es el único paquete desde el que se pueden leer estos
  ficheros, porque el dominio tiene denegada la entrada y salida también en sus `_test.go` (research.md V7, D27). Por
  eso son subtests de `TestTerritorioDelRepositorio` la validación, la integridad, las gramáticas y los tres controles
  del pliegue sobre el corpus real.
- **Integridad entre ficheros**: la comprueba `territorio.Cargar`, y el mismo test la ejerce sobre los ficheros reales
  **delegando en ella**, sin repetir los tipos en `internal/skills` ([data-model.md](../data-model.md) §2.1). Una
  provincia sin comunidad, un DIR3 de un municipio que no existe, un DIR3 incoherente con su código INE o **una fila
  de municipio cuya `comunidad` no es la que declara su provincia** hacen fallar `make ci`. Esta última cierra la
  única redundancia del modelo —la columna que FR-040 exige en la fila— contra el camino autoritativo, que es el de
  la provincia (data-model §2.1, punto 6).
- **Gramáticas**: el subtest `gramaticas` exige que los `pattern` de los tres esquemas acepten y rechacen exactamente
  lo mismo que los analizadores de `internal/core/ids` (research.md V38).
- **Contra el binario**: los ficheros viajan embebidos (`//go:embed`, paquete `data`), así que la ausencia de uno es
  un **error de compilación** y no un modo de fallo en ejecución (FR-056, research.md V2).

## 3. El paquete `data`

```go
// data/datos.go — con su comentario de paquete, que revive exige.

// Package data embebe los ficheros congelados que el binario lleva dentro.
package data

//go:embed territorio/municipios.yaml
var Municipios []byte
//go:embed territorio/dir3.yaml
var DIR3 []byte
//go:embed territorio/estado.yaml
var Estado []byte
//go:embed territorio/comunidades
var comunidades embed.FS

func Comunidades() (map[string][]byte, error)
```

Sin más lógica que leer el subárbol embebido de comunidades. No lo importa `internal/core` (research.md D3): lo
importan `internal/app` y el binario de e2e.

## 4. La tarea `[datos]` que los genera

Fuera del bucle de implementación y **con pausa humana** (FR-045). La pausa se dispara porque la tarea añade también
el esquema bajo `schemas/` (research.md V9, D17): un fichero nuevo bajo `data/` no pausa por sí solo.

Lo que hace la persona en la pausa:

1. Descarga la relación del INE y el volcado del REL por las direcciones de sus filas de `docs/SOURCES.md`, fuera del
   repositorio. **El ejecutor desatendido no descarga nada.**
2. Deriva el DIR3 de cada ayuntamiento a partir del número de inscripción del REL y comprueba que es coherente con el
   código INE y su dígito de control oficial.
3. **Verifica la derivación contra DIR3 real** en una muestra que incluye al menos (FR-046):
   - un municipio fusionado o renombrado,
   - uno de régimen foral,
   - uno con entidades locales menores.
4. Escribe los ficheros de `data/territorio/` con su `fecha` y su `source`, dejando **fuera** todo municipio cuyo
   número de inscripción no sea coherente (FR-048).
5. Registra la verificación (§5).
6. Actualiza la fila de `mpt.rel` de `docs/SOURCES.md` con la fecha del volcado (FR-049). **No se añade ningún caso a
   `scripts/verify-sources.sh`**: estas fuentes no se piden en red y el CI nocturno no puede vigilarlas (ADR 0017).

**Criterio de entrada de una fila** (spec, *Clarifications* Q1; FR-048): entra todo municipio con fila en el REL cuyo
número de inscripción es coherente con su código INE y su dígito de control oficial. La muestra verifica **la regla**,
no cada fila. Si la muestra revela discrepancias más allá de casos aislados, la derivación deja de ser regla y pasa a
ser tabla, y **eso lo decide la persona en esta pausa**, no el ejecutor.

## 5. Registro de la verificación

`specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md`, con una fila por municipio de la muestra (FR-047):

| Municipio | Código INE | DIR3 derivado | DIR3 real | De dónde salió el real | Coincide |
|---|---|---|---|---|---|

Y, al final, la conclusión: regla confirmada (y cuántas filas entran y cuántas quedan fuera) o derivación convertida
en tabla.

## 6. Qué lo vigila

| Control | Test / forma | En `make ci` |
|---|---|---|
| Cada fichero contra su esquema | `TestTerritorioDelRepositorio/esquema` | sí |
| Integridad entre ficheros, incluida la coherencia municipio→comunidad→provincia | `TestTerritorioDelRepositorio/integridad` | sí |
| Gramáticas de los esquemas contra los analizadores | `TestTerritorioDelRepositorio/gramaticas` | sí |
| Que el `source` de cada dato existe | `TestTerritorioDelRepositorio/fuentes` (los identificadores, contra `docs/SOURCES.md`; las rutas, contra el árbol) | sí |
| Que la Comunidad de Madrid está configurada, sin prohibir otras (SC-005) | `TestTerritorioDelRepositorio/madrid-configurada` | sí |
| Que las 19 comunidades tienen régimen | `TestTerritorioDelRepositorio/regimen-de-todas` | sí |
| Que el pliegue de nombres cubre el corpus, no deja municipios inalcanzables y no deja ningún nombre que sea solo cifras | `TestTerritorioDelRepositorio/pliegue-cubre-el-corpus`, `/nombres-alcanzables`, `/ningun-nombre-es-solo-cifras` (research.md D27) | sí |
| Que ningún municipio concreto aparece en código ni en la skill | `TestSkillsDelRepositorio` (normas nombradas y revisión final); los municipios solo en fixtures, e2e y evals | sí / revisión |
| Que el binario no pide estas fuentes | `TestArquitectura` R2 y la ausencia de fila «consultada en ejecución» en `docs/SOURCES.md` | sí |
