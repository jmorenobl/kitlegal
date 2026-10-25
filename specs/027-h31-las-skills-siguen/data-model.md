# Data model: H31 · Las skills siguen la especificación Agent Skills

Entidades del spec («Key Entities») con sus campos, sus reglas y dónde viven. Ninguna se guarda en disco por el
binario: las de §1 a §3 son lecturas de `skills/`; las de §4 a §7, del job de evals y su informe.

## 1. Cuerpo de `SKILL.md`

- **Qué es**: el `SKILL.md` con la región de la tabla de comandos regenerada (`contenidoRegenerado`), desde el byte
  que sigue a la línea `---` que cierra el frontmatter (`bloqueDelFrontmatter` da lo anterior a esa línea).
- **Medidas**: `lineas = ContarLineas(cuerpo)`, `caracteres = utf8.RuneCount(cuerpo)`.
- **Reglas** (FR-004, FR-010): `lineas < 500` y `caracteres <= 11000`; si no, un `DefectoDeSkill` por medida.
- **Vive en**: `internal/skills/sincronia.go` (`regenerarSkill`), en lugar de `defectoDeLasLineas` y `maximoDeLineas`.

## 2. Descripción

- **Qué es**: `Frontmatter.Description`, tal como lo lee `LeerFrontmatter` (el escalar plegado de YAML).
- **Reglas de forma** (FR-002, FR-003, FR-005): no vacía; `utf8.RuneCountInString <= 900`; sin etiqueta XML
  (`</?\p{L}[^<>]*>`). En `name`: las de hoy, sin `anthropic` ni `claude` y sin etiqueta XML.
- **Regla de significado** (FR-022): tercera persona; la juzga el juez de §5 en el job, nunca `make ci`.
- **Vive en**: `internal/skills/frontmatter.go` (`ValidarFrontmatter`); cada defecto empieza por `SKILL.md: `.

## 3. Referencias

| Entidad | Qué es | Regla |
|---|---|---|
| Referencia generada | `.md` de `references/` declarado en `metadata.kitlegal-referencias` | Es lo que genera su fila desde `data/`, con la cabecera `<!-- generado desde data/<f>, no editar -->`; si no, deriva (`contenido-distinto`, `fichero-ausente`) |
| Sobrante | `.md` no declarado que abre con la cabecera | Deriva (`fichero-sobrante`); `skills-sync` lo retira |
| Referencia a mano | `.md` no declarado que no abre con la cabecera | Ni deriva ni retirada; protocolo de detalle de la skill (FR-012, FR-013) |
| Entrada no válida | Lo que no es un fichero regular `.md` | Defecto FR-006 |

- **Para todas las `.md`** (FR-007 a FR-009): un enlace desde `SKILL.md` fuera de bloques de código con destino
  `references/<f>.md` (con o sin `#…`); ningún enlace a otra referencia de la skill; con más de 100 líneas, algún
  `## ` y, antes del primero, una lista que los nombra por orden.
- **Función**: `skills.ReferenciasAMano(arbol fs.FS, skill string) ([]skills.Referencia, error)` —`Referencia` es el
  tipo de hoy, `{Fichero, Contenido}`—, en orden de nombre de fichero. El error, para lo que no se puede leer (regla
  genérica: defecto `inesperado`).
- **Consumidores**: la sincronía (`sobrantes`), las dependencias de las evals (`datosDeLasReferencias`,
  `skillsDeLaEval`) y los controles de contenido de los tests.

## 4. Juicio de una descripción

| Campo | Tipo | Qué es |
|---|---|---|
| `skill` | cadena | La skill de la descripción, la del trabajo |
| `clase` | cadena | `habla_en_primera_o_segunda_persona` |
| `huella` | cadena, 64 hex | SHA-256 del texto de contracts/juez-de-las-descripciones.md §2 |
| `juzgada` | booleano | Votada en esta ejecución (`true`) o reutilizada (`false`) |
| `procedencia` | objeto, si se reutilizó | `ejecucion`, `trabajo`, `commit`, `propuesta`, `fecha`, como H30 |
| `marcada` | booleano | Tres «si» unánimes con frase comprobada |
| `frase`, `motivo` | cadenas | Las del primer voto; vacías sin marca salvo el motivo |
| `votos` | lista | Cada voto dado: `respuesta` (`si`/`no`), `frase`, `motivo`, en orden, con las repeticiones |
| `sin_juzgar` | cadena, si no se pudo | El motivo del voto que no llegó a darse (`voto <n>: <causa>`) |

- **Transiciones**: sin bloque en la decisión → no existe. Con bloque y `juzgada: false` → copia del juicio de `main`.
  Con `juzgada: true` → se vota: «no» o «si» que no llega a tres → `marcada: false`; tres «si» → `marcada: true`; un
  voto que no llega a darse → `sin_juzgar`. Un juicio `sin_juzgar` no se reutiliza (FR-032).
- **Vive en**: el bloque `descripcion` del fichero de la decisión y de `informe.json` (`internal/evals`).

## 5. El juez de las descripciones

Un `Juez` (`internal/evals/juez.go`) con una sola `ClaseDelJuez` que decide y el contenido de los cuatro ficheros de
`evidencias/adr-0046/` (research D2). Lo votan `nuevaVotacion` y `nuevoVotanteDelGuion`, como al de las respuestas.

## 6. Umbral de la clase

Un `Umbral` (`internal/evals/umbrales.go`) con `Nombre` `habla_en_primera_o_segunda_persona`, `Medida` las marcadas
entre las juzgadas en la ejecución, `Total` las juzgadas, `Comparacion` `<=`, `Umbral` 0, `Decide` true. Solo existe si
la decisión trae el bloque `descripcion` (FR-030, FR-031).

## 7. Medida del juez de las descripciones

- **La versionada**: `evidencias/adr-0046/medida.json`, que lee `leerMedidaDelJuez` (`MedidaDelJuez`): `clase`,
  `fecha`, `modelo_del_juez`, `version_de_claude_code`, `rubrica.sha256`, `casos.sha256`, `defectos.{casos,
  sin_marcar}`, `correctos.{casos, marcados}`. Corresponde si las huellas son las de los ficheros, el modelo y la
  versión los fijados y la clase la del juez; se cumple con `sin_marcar` 0 y `marcados` 0 (FR-026).
- **Los casos**: `evidencias/adr-0046/casos.yaml`, `casos:` con 99 elementos de `id`, `skill`, `origen`, `clase`,
  `etiqueta` (`defecto`/`correcto`), `forma` y `descripcion`. Un caso cuenta como mal etiquetado si es un defecto sin
  marcar o un correcto marcado.
- **La de la ejecución** (FR-034): las mismas claves y `origen` (contracts/juez-de-las-descripciones.md §7).
