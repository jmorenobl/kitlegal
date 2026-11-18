# Contrato: las cuatro skills del producto con las descripciones de la variante

## 1. Qué cambia y qué no

En cada `skills/<skill>/SKILL.md` de `consulta-juridica`, `jurisprudencia`, `normativa` y `plazos-y-territorio`
cambian **solo las líneas de la clave `description`**: la que empieza por `description:` y las sangradas que la siguen
(la regla de `lineasDelSkill`, `internal/evals/laboratorio.go:321`). Las de hoy dejan su sitio, detrás de `name:`, a
las de §2. Todo lo demás de las cuatro skills —el resto del frontmatter (`name`, `metadata`), el cuerpo, cada
referencia escrita a mano o generada y los nombres de los ficheros— queda byte a byte como en la base de la rama
(`ab951c22`), y no se añade ni se quita ningún fichero (FR-003). `laboratorio/`, `evals/`, `web/`, `mcp/`,
`internal/empaquetado/`, `.goreleaser.yaml` y `.github/workflows/evals.yml` no cambian (FR-006, FR-007).

## 2. Los cuatro bloques, byte a byte

Cada bloque es la línea `description: >-` y cuatro líneas sangradas con dos espacios, con el salto de línea al final
de cada una. Regla del corte (research D4): cada línea, con su sangría, de como mucho 120 caracteres, con tantas
palabras como quepan, cortando solo en un espacio, que desaparece. `skills.LeerFrontmatter` lee cada bloque como el
texto de `laboratorio/producto-descripciones-cortas/<skill>.txt` sin su salto final (research V2).

`skills/consulta-juridica/SKILL.md` (detrás, la línea `---` que cierra el frontmatter):

```text
description: >-
  Punto de entrada de toda pregunta legal en España, de cualquier materia: dice qué capacidad de kitlegal la resuelve
  —normativa, jurisprudencia, plazos y territorio— y con qué skill, y lo que no cubre lo dice sin dar de memoria plazos,
  órganos ni vías. Se usa al empezar toda pregunta legal, antes que las demás skills de kitlegal, también si no nombra
  norma o pide qué hacer, a quién reclamar o qué escrito presentar.
```

`skills/jurisprudencia/SKILL.md` (detrás, `metadata:`):

```text
description: >-
  Sentencias y jurisprudencia de los tribunales españoles, sin citar ninguna de memoria. Se usa siempre que la pregunta
  nombre, pida o cite una sentencia (STS, ECLI, ROJ), pregunte si existe o qué dice, pida jurisprudencia sobre una
  materia o traiga una resolución, aunque el agente crea conocerla. kitlegal no consulta el CENDOJ: prepara la búsqueda
  y coteja el documento. Se activa antes de cita_preparar o cita_cotejar.
```

`skills/normativa/SKILL.md` (detrás, `metadata:`):

```text
description: >-
  Consulta y cita normativa consolidada del BOE de cualquier materia: lee con kitlegal índice y artículos y cita
  identificador y bloque. Se usa siempre que la respuesta dependa de lo que dice una norma —un artículo, ley o real
  decreto, el plazo o requisito que fija— o de lo que decía en una fecha, aunque el agente crea saberlo. Se activa antes
  de boe_articulo u otra herramienta boe_…: qué pedir y cómo citar lo dice la skill.
```

`skills/plazos-y-territorio/SKILL.md` (detrás, `metadata:`):

```text
description: >-
  Calcula con kitlegal hasta qué día se puede recurrir una resolución administrativa, cuándo vence un plazo del
  procedimiento común (LPAC) o el del contencioso y cuándo hay silencio, con los festivos que cuentan, e identifica el
  territorio de un municipio: INE, provincia, comunidad, DIR3 y boletines. No cuenta plazos de otras materias. Se usa al
  preguntar cuándo vence un plazo o un recurso, o qué boletines tiene un municipio.
```

## 3. Medidas (research V1)

| Skill | Runas | Bytes | Hoy (runas) | Líneas del bloque |
|---|---|---|---|---|
| `consulta-juridica` | 417 | 431 | 705 | 1 + 4 |
| `jurisprudencia` | 420 | 424 | 842 | 1 + 4 |
| `normativa` | 426 | 438 | 887 | 1 + 4 |
| `plazos-y-territorio` | 427 | 435 | 879 | 1 + 4 |
| **Total** | **1690** | **1728** | **3313** | |

Límites (FR-004): como mucho 450 cada una y 1700 entre las cuatro (los de la variante, en runas como en H41.3), como
mucho 900 y ninguna etiqueta XML (ADR 0046, puntos 2 y 3). La tercera persona la juzga el job (punto 4).

## 4. `CHANGELOG.md` (FR-008)

Una viñeta al final de «[Unreleased]» → «### Cambiado», tal cual:

```markdown
- **Las cuatro skills del producto llevan las descripciones cortas medidas en H41.3** (ADR 0055): la `description` de
  `consulta-juridica`, `jurisprudencia`, `normativa` y `plazos-y-territorio` pasa a ser la de la variante
  `producto-descripciones-cortas` del laboratorio, 1690 caracteres entre las cuatro en lugar de 3313, para quitar la
  caída de la eval 04 de `normativa` en modo herramienta, en la que la sesión respondía sin activar ninguna skill.
  Nada más de las skills cambia: ni sus cuerpos, ni sus referencias, ni sus nombres. El plugin, el `.mcpb` y el
  `llms.txt` de la web las llevan con la release siguiente.
```

## 5. Uso, de fuera adentro

- **Quién las lee y cuántas veces.** El host (Claude Code u otro) pone las cuatro en el listado de skills al empezar
  cada sesión: una vez por sesión, no por pregunta ni por norma. Ninguna invocación de `kitlegal` las lee; el cuerpo de
  cada `SKILL.md` no cambia, así que las órdenes y llamadas por pregunta son las de la base.
- **Tamaño con meses de uso.** 1690 caracteres (1728 bytes) fijos: con cientos de normas y miles de bloques consultados
  el listado sigue siendo de cuatro textos, porque ni la caché ni el grafo entran en él. Un ejemplo con sus bytes, el de
  `normativa`: 438 bytes de texto; en el fichero, 463: 16 de `description: >-` con su salto y cuatro líneas de 115,
  119, 124 y 89 bytes con su sangría y su salto (112, 115, 120 y 84 caracteres; medido en la copia de research).
- **Señales.** Ninguna: el texto no lleva líneas `⚠` ni nada que caduque; cambia solo con una release.
