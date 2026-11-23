# Data model: H45 · El molde admite un paquete sin código

Ningún esquema ni fichero de `data/` cambia de forma: el hito cambia cómo se reparte lo que ya hay. Los nombres de Go son
los de research.md.

## 1. Paquete

- **Qué es**: un valor de `vertical:` de una norma de `data/normas.yaml` (hoy solo `administrativo`, diez normas).
- **Su referencia** en una skill de capacidad: `skills/<skill>/references/<paquete>.md`. El generador de una región toma
  el paquete del nombre del fichero sin `.md` (`contextoDeLaRegion.paquete`, `regiones.go:628`), como hoy.
- **Regla**: ninguna. Un nombre de fichero que no es un paquete no recibe filas (lo de hoy).

## 2. Tipo de plazo y su paquete (FR-001 a FR-004)

Entrada de `tipos:` de `data/plazos.yaml` (`plazos.Tipo`, `internal/core/plazos/tipos.go:72-93`).

| Campo que cuenta | Origen | Cuenta para el paquete |
|---|---|---|
| `preceptos[].norma` | el YAML | sí |
| `unidades.<unidad>.norma` | el YAML | sí |
| `PreceptosDelComputo` | copiado por `LeerTipos` del cómputo | no |

- `paquetesDelTipo(tipo, normas)`: el conjunto, en orden alfabético, de los `vertical` no vacíos de esas normas; una
  norma que no está en `data/normas.yaml` no aporta ninguno.
- **Estados**: 0 paquetes → defecto «sin paquete» (§C.1); 1 → el tipo es de ese paquete; 2 o más → defecto «de más de
  uno», con los paquetes separados por «, ».
- **Con los datos de hoy**: los siete tipos, 1 paquete cada uno, `administrativo` (LPAC, LJCA y Ley 19/2013 llevan
  `vertical: administrativo`).

## 3. Fila de lo regulado aparte y su paquete (FR-005 a FR-007)

Entrada de `regulados_aparte:` de `data/cobertura.yaml` (`ReguladoAparte`, `internal/skills/cobertura.go:49-60`).

- `paqueteDeLaFila(entrada, normas)`: el `vertical` de `norma.identificador` si la fila lleva `norma` y esa norma está en
  `data/normas.yaml` con `vertical`; si no, el de la norma de `precepto.norma`; si tampoco, ninguno.
- **Estados**: con paquete y su referencia con la región `cobertura_por_materia` → la fila sale en ella; sin paquete o
  con un paquete sin la región → defecto de `filasSinRegion` con el texto de hoy.
- **Línea**: la de hoy, con la norma del precepto y, si la lleva, la que la rige (FR-007).
- **Con los datos de hoy**: las nueve filas, de `administrativo` (la LGT y el texto de tráfico no llevan `vertical`: la
  primera está sin él y la segunda no está en `data/normas.yaml`).

## 4. Raíz de grabaciones (FR-015, §C.4)

| Raíz | Ruta relativa al paquete | Origen |
|---|---|---|
| H4 | `../../internal/source/boe/testdata/boe.legislacion-consolidada` | constante |
| H5 | `../../testdata/evals/boe.legislacion-consolidada` | constante |
| H9 | `../../internal/evals/testdata/boe.legislacion-consolidada` | constante |
| cada hito | `../../evidencias/<carpeta>/boe.legislacion-consolidada` | `evidencias/` |

- **Carpeta de un hito**: `h<n>` o `h<n>-<m>`, con `n` y `m` enteros, que es como las nombra `grabar-datos.sh:100`.
- **Entra** si la carpeta tiene `manifiesto.json` y `boe.legislacion-consolidada/`. **Orden**: `n` y después `m`, en
  número (sin subhito antes que con él).
- **Hoy**: entra `h36`; no entran `h6`, `h9` (sin la carpeta del BOE), `h23` (vacía), `adr-*` ni `sondeos-*`.
- **Regla que se mantiene**: dos raíces comparten solo `GET_https_www.boe.es_robots.txt.json`.

## 5. Declaración del laboratorio (FR-010, FR-014)

`declaracionDelProducto` (`internal/evals/laboratorio_familias_test.go:343`): `desde`, `cambios`, `movimientos`,
`enLasReferencias`. Lo nuevo:

- `normativa` gana un `cambio` (las líneas 38-41 de su cuerpo de `main`, con su sangría, por nada) y un `movimiento`
  (esas líneas, sin sangría, a `con_herramientas.md`, tras una línea en blanco).
- **Exento de declarar**: `skills/<skill>/references/<v>.md` del producto que su base no tiene, con `<v>` un paquete de
  §1. Lo que la base tiene se compara siempre.

## 6. La copia de US4.1 (SC-004)

El paquete `prueba` que añade el guion `paquete-de-mas.txtar` a una copia del módulo (research D3). Solo añade:

| Dónde | Qué |
|---|---|
| `data/normas.yaml`, `normas:` | `BOE-A-2015-11722` con `vertical: prueba` |
| `data/plazos.yaml`, `tipos:` | `prueba-de-un-paquete-mas`, cómputo `lpac`, 1 mes, precepto `BOE-A-2015-11722` `a1` |
| `data/cobertura.yaml`, `cubierto:` | capacidad `plazos`, `materia: prueba` |
| `data/cobertura.yaml`, `regulados_aparte:` | precepto DA 1ª.2.c de la LPAC, `norma` `BOE-A-2015-11722` |
| `skills/plazos-y-territorio/references/prueba.md` | los dos motivos y las dos regiones, vacías |
| `evals/plazos-y-territorio/` | `99-` + el nombre de su primer fichero, copia de él |

Tras sincronizar: `prueba.md` lleva el tipo, la fila nueva y la de tráfico de hoy; `administrativo.md`, ocho filas de lo
regulado aparte y los siete tipos.

## 7. La vertical de ensayo (FR-019)

`internal/skills/testdata/vertical-de-ensayo/` gana:

| Fichero | Qué |
|---|---|
| `data/plazos.yaml` (nuevo) | `tipos:` con `ensayo`: cómputo `lpac`, 10 días hábiles, `da: vencimiento`, precepto `BOE-A-1900-1` `a2` `"2"` `fija: duracion` |
| `data/cobertura.yaml` | una fila más de `regulados_aparte` con su precepto en la DA 1ª.2.c de la LPAC y `norma` `BOE-A-1900-2` |
| `skills/plazos-y-territorio/references/ensayo.md` | la sección y las marcas de `recursos_y_plazos`, vacías |
