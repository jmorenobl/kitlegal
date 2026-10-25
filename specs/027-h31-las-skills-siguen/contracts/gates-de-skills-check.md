# Contrato: los gates de `skills-check` y la regla de `references/` (FR-002 a FR-015)

Quién lo consume: quien ejecuta `make ci`, `make skills-check` o `make skills-sync` —una persona o una tarea de un
run—, que lee los defectos para arreglar la skill. Cuántas veces: una por ejecución, una línea por regla incumplida y
skill. Cuándo deja de darse cada línea: cuando la skill se corrige. No crece con el uso de kitlegal: lo acotan las
skills del repositorio (tres hoy) y sus referencias.

## 1. Dónde se comprueba

Todo vive en `internal/skills` y lo ejecuta lo que ya ejecuta `skills-check` y `skills-sync`:
`TestSkillsDelRepositorio` (`internal/app/skills_test.go`) llama a `skills.Regenerar` y `skills.Comparar`
(`fallosDeLasSkills`), y `scripts/skills-sync.sh` a `skills.Regenerar` y `skills.Escribir` (`escribirSkills`). Un
defecto (`DefectoDeSkill`) impide escribir: `Escribir` responde «no se escribe nada porque hay defectos: …», como hoy.
Una deriva (`Deriva`) la rehace `Escribir`. Sin red, sin modelo.

## 2. Los defectos

`DefectoDeSkill.Error()` es `<skill>: <defecto>`; cada defecto nombra su fichero y su regla, y da la medida y el
límite cuando los hay. Formas, con un ejemplo de bytes cada una (la skill del ejemplo, `boe-legislacion`):

| Regla | Requisito | Defecto |
|---|---|---|
| Las de hoy de `ValidarFrontmatter` | FR-002 | `SKILL.md: ` y el texto de hoy, p. ej. `boe-legislacion: SKILL.md: name vacío` |
| `name` con `anthropic` o `claude` | FR-002 | `boe-legislacion: SKILL.md: name "claude-legal" contiene «claude»` |
| Etiqueta XML en `name` | FR-002, FR-003 | `boe-legislacion: SKILL.md: name con la etiqueta XML «<x>»` |
| Etiqueta XML en `description` | FR-002, FR-003 | `boe-legislacion: SKILL.md: description con la etiqueta XML «<ejemplo>»` |
| `description` de más de 900 | FR-005 | `boe-legislacion: SKILL.md: description de 901 caracteres (máximo 900)` |
| Cuerpo de 500 líneas o más | FR-004 | `boe-legislacion: SKILL.md: cuerpo de 500 líneas (máximo 499)` |
| Cuerpo de más de 11 000 caracteres | FR-004, FR-010 | `boe-legislacion: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)` |
| Entrada de `references/` que no es un fichero `.md` | FR-006 | `boe-legislacion: references/notas.txt: no es un fichero .md` · `boe-legislacion: references/anexos: no es un fichero .md` |
| Referencia sin enlace desde `SKILL.md` | FR-007 | `boe-legislacion: references/como_se_cita.md: SKILL.md no la enlaza con [](references/como_se_cita.md)` |
| Referencia que enlaza otra | FR-008 | `boe-legislacion: references/como_se_cita.md: enlaza otra referencia, references/varios_bloques.md` |
| Más de 100 líneas, sin índice | FR-009 | `boe-legislacion: references/como_se_cita.md: 101 líneas sin un índice que nombre sus apartados ## antes del primero (más de 100)` |
| Más de 100 líneas, sin `## ` | FR-009 | `boe-legislacion: references/como_se_cita.md: 101 líneas sin ningún apartado ## (más de 100)` |

«Caracteres» son runas (`utf8.RuneCount`); las líneas, las de `ContarLineas`, que es como contaba el límite de hoy
(FR-010). Desaparecen `maximoDeLineas` y el defecto «SKILL.md tiene N líneas (máximo 299)»; desaparecen también el
máximo de 1024 (lo cubre el de 900) y la regla «description con < o >», que sustituye la de la etiqueta XML.

### 2.1 Etiqueta XML (FR-003)

Un texto que casa con `</?\p{L}[^<>]*>`. Casan: `<ejemplo>`, `</b>`, `<br/>`, `<descripcion skill="x">`. No casan:
`< 5000`, `a<b`, `a < b > c`, `<>`, `<1>`. El defecto cita la primera que encuentra.

### 2.2 El cuerpo (FR-004)

El `SKILL.md` con la región de la tabla de comandos regenerada (`contenidoRegenerado`), desde el byte que sigue a la
línea `---` que cierra el frontmatter hasta el final. Medido en `main` (research V1): 283 líneas y 24 439 caracteres,
283 y 19 225, 193 y 13 637. En el límite: 499 líneas, en verde; 500, en rojo; 11 000 caracteres con más de 11 000
bytes, en verde; 11 001, en rojo.

### 2.3 Enlaces, bloques de código e índice (FR-007 a FR-009)

- **Bloque de código**: de una línea que abre una valla —hasta tres espacios y tres o más `` ` `` o `~`— a la que la
  cierra con la misma marca, o al final del fichero.
- **Enlace**: `[texto](destino)` en una línea fuera de un bloque; el destino llega hasta el primer `)` o espacio. Se
  quita lo que sigue a `#`.
- **FR-007**: alguno de los enlaces de `SKILL.md` tiene por destino `references/<fichero>.md`. Ejemplo que cumple:
  `- [como_se_cita](references/como_se_cita.md): para otra forma de cita.`; también
  `[ver](references/como_se_cita.md#los-avisos)`. No cumple: `` `references/como_se_cita.md` `` ni el mismo enlace
  dentro de un bloque `text`.
- **FR-008**: el destino de un enlace de una referencia, resuelto desde `references/` con `path.Join`, es otro fichero
  de `references/` de la skill: `varios_bloques.md`, `./varios_bloques.md`, `../references/varios_bloques.md`. No
  cuentan `#apartado`, `como_se_cita.md#x` desde `como_se_cita.md` ni un destino fuera de `references/`.
- **FR-009**: con más de 100 líneas, los apartados son las líneas fuera de bloques que empiezan por `## `; tiene que
  haber alguno, y antes del primero una lista —líneas que empiezan, tras su sangría, por `- `, `* ` o `<n>. `— en la
  que, por orden, cada apartado tenga un elemento que contenga su texto (sin `[`, `](…)`, `*`, `_` ni `` ` ``). Con 100
  líneas o menos no se exige nada.

Ejemplo de índice que cumple:

```markdown
# Cómo se cita

- [Lo que hace cita](#lo-que-hace-cita)
- [Los avisos](#los-avisos)

## Lo que hace cita
…
## Los avisos
…
```

## 3. La regla de `references/` (FR-012, FR-013)

`ReferenciasAMano(arbol fs.FS, skill string) ([]Referencia, error)` es la única que clasifica, y la usan la sincronía,
las dependencias de las evals y los controles de contenido. De cada entrada de `skills/<skill>/references/`, en este
orden:

| La entrada | Es | `skills-check` | `skills-sync` |
|---|---|---|---|
| No es un fichero regular `.md` (subdirectorio, enlace, otra extensión) | defecto FR-006 | rojo | no escribe nada |
| `.md` declarado en `kitlegal-referencias` | referencia generada | deriva `contenido-distinto` si no es lo que genera su fila, con o sin cabecera | la rehace |
| Declarada y ausente | — | deriva `fichero-ausente` | la escribe |
| `.md` no declarado que abre con `<!-- generado desde data/<f>, no editar -->` | sobrante | deriva `fichero-sobrante` | lo retira |
| `.md` no declarado que no abre con la cabecera | referencia a mano | nada (solo los gates de §2) | no la toca, byte a byte |

El contenido de una referencia a mano es el protocolo de detalle de su skill y no repite lo que sale de `data/`
(FR-013, lo contrasta la revisión). Las generadas no cambian: ni su contenido ni sus generadores (FR-024; Clarificación
Q4).

## 4. Lo empotrado (FR-015)

Sin cambio: `//go:embed skills/*/SKILL.md skills/*/references/*` empotra todo fichero de `references/`, y
`TestSkillsEmpotradas` (raíz del módulo) falla si uno del árbol no está en el binario con sus bytes.
`kitlegal skills install` instala lo empotrado.
