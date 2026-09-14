# Contrato: `make skills-sync` y `make skills-check`

FR-030 a FR-036, FR-040 a FR-044, US5, SC-005. Decisiones: research.md D2 (dónde vive el código), D3 (frontmatter),
D4 (declaración), D6 (tabla), D7 (sincronización como test).

## 1. Órdenes

| Orden | Receta | En `make ci` | Escribe |
|---|---|---|---|
| `make skills-sync` | `scripts/skills-sync.sh` | no | sí: `references/`, región de `SKILL.md`, enlaces de `scripts/` |
| `make skills-check` | ver §1.1 | **sí** | nunca |

### 1.1 Recetas

```make
## skills-sync: regenera references/, la tabla de comandos de SKILL.md y los enlaces de scripts/ de cada skill
skills-sync: check-tools
	scripts/skills-sync.sh

## skills-check: comprueba skills, datos y evals sin red, sin modelo y sin escribir nada
skills-check: check-tools
	go test -count=1 -run '^(TestSkillsDelRepositorio|TestNormasDelRepositorio|TestEvalsDelRepositorio|TestIdentificadoresDeLasNormas)$$' ./internal/app/ ./internal/skills/ ./internal/evals/

ci: fmt-check lint test test-integration vuln schema-check skills-check secrets mod-verify mod-tidy-check
```

`scripts/skills-sync.sh`:

```bash
#!/usr/bin/env bash
# Regenera references/, la tabla de comandos de SKILL.md y los enlaces de scripts/ de cada skill de skills/,
# desde data/*.yaml y desde --describe del binario (contracts/sincronizacion-y-comprobacion.md de H5).
set -euo pipefail
cd "$(dirname "$0")/.."

go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills
```

`make skills-check` es **una sola** invocación de `go test` sobre los tres paquetes: así un fallo en uno no oculta los de
los otros (una receta de dos líneas se detendría en la primera que falla). Cada nombre de test existe en un único
paquete, y `-run` se aplica a cada uno.

`make skills-sync` deja de anunciar que las skills llegan en H5 (FR-033). Ni la receta ni el guion invocan Python
(FR-044, SC-007).

## 2. `TestSkillsDelRepositorio` (`internal/app/skills_test.go`, paquete `app`)

Vive en `internal/app` por la misma razón que `TestEsquemasPublicados`: es donde está el registro de producción y la
función `describir` que emite `--describe`, de modo que la tabla sale **de la misma función** que atiende
`kitlegal <applet> <verbo> --describe`, sin exportar nada nuevo del kernel (research.md D2).

Flujo:

1. `app.RegistroDeProduccion()`; para cada applet declarado por alguna skill, `describir` de cada verbo en un
   `render.Nuevo` sobre un búfer; y `cli.Describir` con `Argumentos` nulo para las banderas globales.
2. `skills.Regenerar(raiz, descripciones)` construye en memoria, para cada skill, las referencias, `SKILL.md` con la
   región regenerada y el conjunto de enlaces esperados, y devuelve también los defectos de la skill.
3. Con `-regenerar-skills`, antes de comparar, `skills.Escribir(raiz, regenerado)` escribe lo que difiere, crea los
   enlaces que faltan, corrige los que apuntan a otro sitio y retira sobrantes de `references/` y `scripts/`; no escribe
   nada si hay defectos de skill o de datos, que se presentan igual.
4. `skills.Comparar(raiz, regenerado)` devuelve derivas (data-model §5); cualquier defecto o deriva hace fallar el
   subtest nombrando skill y fichero o enlace.

| Subtest | Qué fija | Requisito |
|---|---|---|
| `skills` | el árbol real: sin defectos ni derivas | FR-040 a FR-042, US5-1 |
| `referencia-editada` | copia temporal con una línea añadida a `references/normas.md` → `contenido-distinto` nombrando `boe-legislacion` y el fichero | US5-2, SC-005 |
| `datos-sin-regenerar` | copia con un título cambiado en `data/normas.yaml` → `contenido-distinto` en `references/normas.md` | US5-2, SC-005 |
| `describe-cambiado` | descripciones de verbo con la ayuda de `articulo` cambiada → `contenido-distinto` en `SKILL.md` | US5-3, SC-005 |
| `trescientas-lineas` | `SKILL.md` de 300 líneas → defecto `SKILL.md tiene 300 líneas (máximo 299)`; con 299 no hay defecto | US5-4, FR-041 |
| `frontmatter` | una fila por defecto: sin `name`, `name` distinto del directorio, `name` con mayúscula, con `--`, empezando por `-`, de 65 caracteres; sin `description`, `description` vacía, de 1025 caracteres, con `<`; clave no admitida; `kitlegal-applets` con un applet que no existe; `kitlegal-referencias` sin `data/<n>.yaml` — cada una nombrando skill y defecto; y los límites exactos válidos (64 y 1024) | US5-4, FR-040, SC-005 |
| `enlaces` | enlace ausente, sobrante, con otro destino y fichero regular en lugar de enlace → cada clase de deriva nombrando skill y enlace | US5-5, FR-036, SC-005 |
| `region` | sin marcas, con dos inicios, con el fin antes del inicio → defecto nombrando la skill | FR-032 |
| `regenerar-dos-veces` | sobre una copia temporal con derivas, `Escribir` + `Comparar` da cero derivas, y un segundo `Escribir` no cambia ningún byte ni ningún enlace (tiempos de modificación incluidos) | FR-034, FR-036, FR-055 |
| `normas-nombradas` | cada «Ley N/AAAA», «Ley Orgánica N/AAAA», «Real Decreto N/AAAA», «Real Decreto-ley N/AAAA» o «Real Decreto Legislativo N/AAAA» de `SKILL.md` es el comienzo del título de una norma de `data/normas.yaml` | FR-020 |
| `sin-instrucciones-de-evals` | contrato de la skill §2.5 | FR-077, SC-012 |

Las copias temporales se hacen con un auxiliar que copia ficheros y **recrea los enlaces simbólicos** con su destino
literal (`os.Lstat`, `os.Readlink`, `os.Symlink`).

## 3. Región generada de `SKILL.md`

Entre las marcas de data-model §2, exactamente:

```markdown

### `scripts/boe`

| Orden | Qué hace | Qué devuelve en `data` |
|---|---|---|
| `scripts/boe buscar <texto>...` | Busca normas consolidadas por las palabras de su título o con una consulta de la fuente. | lista de objetos con `…` |
| `scripts/boe indice <norma>` | Devuelve los bloques de una norma consolidada, en el orden de la fuente. | objeto con `…` |
| `scripts/boe articulo <norma> <bloque>` | Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia. | objeto con `norma`, `bloque`, `titulo`, `tipo`, `fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, `url_eli` |
| `scripts/boe articulos <norma> <bloques>...` | Devuelve el texto vigente de varios bloques de una norma, en el orden pedido. | lista de objetos con `…` |
| `scripts/boe metadatos <norma>` | Devuelve los datos de una norma y los avisos de su vigencia. | objeto con `…` |
| `scripts/boe analisis <norma>` | Devuelve las materias, las notas y las referencias de una norma. | objeto con `…` |

Todas devuelven el sobre `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

```

- Los `…` son las claves reales que da `--describe` en el orden de su `$defs` (data-model §2.1); la fila de `articulo`
  muestra las de hoy (`kitlegal boe articulo BOE-A-2015-10565 a21 --describe`, comprobado en local).
- Una sección `###` por applet declarado, en el orden de `kitlegal-applets`; filas en el orden del registro.
- `|` en la ayuda se escribe `\|`.
- `TestRenderizarTabla` (`internal/skills`) fija los bytes exactos sobre documentos de `--describe` de prueba (objeto,
  lista, sin `$ref`, argumento de varios valores, argumento opcional, ayuda con `|`).
- `TestTablaDeComandosCoincideConLaGramatica` (`internal/app`): para cada verbo del registro de producción, la sintaxis
  generada se convierte en una invocación (`x` por argumento, `x y` por argumento de varios valores, opcionales
  omitidos) seguida de `--describe`, y `invocar` sobre el registro de producción termina en 0. Si un día un verbo
  tuviera una bandera obligatoria, la sintaxis `<nombre>` no la aceptaría la gramática y el test fallaría (research.md
  D6).

## 4. Tests de `internal/skills` (unitarios)

| Test | Fichero | Qué fija |
|---|---|---|
| `TestLeerFrontmatter` | `frontmatter.go` | sin delimitadores, sin cierre, YAML inválido, mapa de `metadata` con valor no cadena, y clave repetida (`name` dos veces; `kitlegal-applets` dos veces dentro de `metadata`) nombrando la clave y sus dos líneas, con el lector común de data-model («Lectura de documentos YAML») |
| `TestValidarFrontmatter` | `frontmatter.go` | las reglas de data-model §1.1 y §1.3 con sus límites |
| `TestContarLineas` | `skill.go` | vacío = 0; `a` = 1; `a\n` = 1; `a\nb` = 2 |
| `TestListarYCargar` | `skill.go` | todo directorio de `skills/` es una skill; sin `SKILL.md` es un defecto; `skills/` ausente da cero skills |
| `TestDescripcionDeVerbo` | `comandos.go` | argumentos sin las banderas globales, obligatoriedad, varios valores, claves de `data` de objeto y de lista, sin `$ref` |
| `TestSustituirRegion` | `comandos.go` | sustituye solo entre marcas; defectos de marcas |
| `TestRenderizarTabla` | `comandos.go` | §3 |
| `TestEnlacesEsperados` | `enlaces.go` | un enlace por applet, destino literal |
| `TestRegenerarYComparar` | `sincronia.go` | cada clase de deriva de data-model §5 sobre árboles temporales |
| `TestValidarDocumentoYAML` | `esquemas.go` | lector común de data-model («Lectura de documentos YAML») y validación contra un esquema en línea: válido, clave desconocida, tipo, patrón; clave repetida en la raíz y en un mapa anidado, rechazada con la ruta, la clave y sus dos líneas antes de validar, aunque el esquema aceptara cualquiera de los dos valores; mapa con una clave que no es texto; errores con la ruta dentro del documento y su línea |
| `TestLeerNormas`, `TestEsquemaDeNormas`, `TestNormasDelRepositorio` | `normas.go` | contrato de normas y referencias §2-§4 |
| `TestRenderizarNormas` | `referencias.go` | contrato de normas y referencias §5 |
| `TestInstalacion` | `instalacion_test.go` | contrato de instalación §4 |
