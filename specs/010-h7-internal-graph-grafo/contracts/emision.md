# Contrato: qué emite cada applet

FR-040 a FR-047, FR-065, FR-071. Decisiones: [../research.md](../research.md) D3, D20, D21. Vocabulario:
[../data-model.md](../data-model.md) §7.

## 1. `boe articulo` y `boe articulos` (`internal/source/boe/grafo.go`)

Para cada `Articulo` del resultado de éxito (en `articulos`, cada bloque distinto) cuyo `url_eli` tiene ELI:

| Operación | Id | Datos |
|---|---|---|
| Nodo `Norma` | `<eli>`: de la ruta de `url_eli` (`net/url`), desde el primer segmento que es exactamente `eli` hasta el final, sin consulta ni fragmento; `https://www.boe.es/eli/es/l/2015/10/01/39` → `eli/es/l/2015/10/01/39` | `{"identificador": "BOE-A-2015-10565"}` |
| Nodo `Bloque` | `<eli>#<bloque>` | `{"bloque": "a21"}` |
| Nodo `BloqueVersion` | `<eli>#<bloque>@<fecha_vigencia>:<hash_texto>` | `{"fecha_vigencia", "fecha_version", "norma_modificadora", "hash_texto"}` tal como las lleva el `Articulo` |
| Arista | `Norma` —`eli:has_part`→ `Bloque` | |
| Arista | `Bloque` —`eli:has_version`→ `BloqueVersion` | |
| Texto | huella `hash_texto`, cuerpo `Texto` | |

- `Observado.Vigencia` = la vigencia con la que la fuente guarda la consulta (`vigenciaLarga`, 604 800 s).
- Sin ELI —`url_eli` vacío, que no se puede analizar o sin ningún segmento `eli`— el artículo no emite **nada**, ni
  con un id alternativo (FR-040); un resultado sin ninguna operación no abre `world.db` (FR-030).
- `indice`, `buscar`, `metadatos` y `analisis` no emiten (FR-040). Los fallos y `--dry-run` no emiten.
- La emisión no cambia ni un byte de `data`, de la salida, del código, de `schemas/norma.json` ni de
  `schemas/bloque.json`, ni lo que se pide a la fuente o a la caché (FR-042).

## 2. `territorio resolver` (`internal/core/territorio/grafo.go`)

`func (t Territorio) Observado() schema.Observado`, que `internal/app/territorio.go` pone en el resultado de éxito:

| Operación | Id | Datos | Cuándo |
|---|---|---|---|
| Nodo `Municipio` | `ine:<código INE de cinco cifras>` | `{"codigo_ine": "28074", "nombre": "Leganés"}` | siempre |
| Nodo `Organo` | el DIR3 del ayuntamiento (Leganés: `L01280745`, `data/territorio/dir3.yaml`) | `{"dir3": "L01280745"}` | si la respuesta trae DIR3 |
| Arista | `Organo` —`lb:pertenece_a`→ `Municipio` | | si la respuesta trae DIR3 |

Sin vigencia (`territorio` no consulta ninguna fuente en ejecución). La emisión es la misma para un municipio cubierto y
para uno no cubierto (FR-045). Sin DIR3, solo el `Municipio` (FR-044): lo prueba el dominio con fuentes sintéticas,
porque ningún municipio de los datos congelados carece de DIR3 (research V28).

## 3. Los que no emiten

`skills` (los tres verbos), los tres verbos de `graph` y los applets de ejemplo devuelven `Resultado.Grafo` vacío sin
implementar nada (FR-046, FR-047).

## 4. Qué lo vigila

| Control | Test |
|---|---|
| Ids, datos, aristas y texto de `articulo` y de `articulos` (bloque repetido una vez); sin ELI y ELI sin `eli` → nada; vigencia 604 800 s; `indice`, `buscar`, `metadatos` y `analisis` sin operaciones | `TestObservadoDeBoe` (`internal/source/boe/grafo_test.go`) y `TestFetch…` existentes |
| La salida de `boe articulo` byte a byte igual con y sin `--no-graph`, con y sin `--json`, para cada respuesta grabada (SC-003) | `TestLaSalidaDeBoeNoCambiaConElGrafo` (`internal/app/grafo_test.go`) y `h7-grafo-no-interferencia` |
| `Municipio`, `Organo` y arista; sin DIR3 solo `Municipio`; igual con y sin cobertura | `TestObservadoDeTerritorio` (`internal/core/territorio/grafo_test.go`) y `h7-grafo-matriz-territorial` |
| Nadie más emite; `world.db` no se crea | `h7-grafo-no-emiten` |
