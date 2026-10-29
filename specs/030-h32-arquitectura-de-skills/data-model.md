# Data model · H32

Las entidades del hito, sus campos y sus reglas. El detalle de cada una, en el contrato que se cita.

## 1. Skill de entrada y skill de capacidad ([contracts/reparto.md](contracts/reparto.md))

| Campo | Entrada (`consulta-juridica`) | Capacidad (`normativa`, `jurisprudencia`, `plazos-y-territorio`) |
|---|---|---|
| `name` | igual al directorio; sin `boe`, `cendoj`, `eur`, `anthropic`, `claude` | ídem |
| `description` | ≤ 900 caracteres, tercera persona, cualquier materia, sin «derecho público» | ídem; la de plazos dice qué plazos cuenta hoy |
| `metadata.kitlegal-applets` | ninguno | los de su capacidad; cada applet, en una sola skill |
| `metadata.kitlegal-referencias` | ninguna | las generadas enteras de la base (`normas`, `jerarquia_normativa`) |
| cuerpo | ≤ 11 000 caracteres; método, capacidades, cobertura y la lista de paquetes entre sus marcas | ≤ 11 000 con lo generado; protocolo, invariantes, formas fijas, tabla de comandos y las marcas de los paquetes |

Regla: el cuerpo de una skill de capacidad no nombra una norma con `vertical` ni remite a otra skill de la entrega.

## 2. Norma (`data/normas.yaml`; `schemas/normas.yaml.json`)

`titulo`, `rango`, `materias`, `abreviatura?`, `vertebral?` (hoy) y **`vertical?`** (nuevo): `^[a-z][a-z0-9-]*$`, el
nombre del paquete que la trae. Entra la Ley 50/1980 (`BOE-A-1980-22501`, `materias: [seguros]`, sin `vertical`), con
su identificador y su título de lo grabado.

## 3. Paquete de materia ([contracts/paquetes.md](contracts/paquetes.md))

No es un fichero: es el conjunto de

- su **nombre**: un valor de `vertical`;
- sus **normas**: las de `data/normas.yaml` con ese `vertical`;
- sus **referencias**: `skills/<skill>/references/<nombre>.md`, una por skill de capacidad que la necesita, escritas a
  mano, con regiones generadas;
- sus **filas de cobertura**: las de `data/cobertura.yaml` cuyo `precepto.norma` es una de sus normas;
- sus **evals**: en la carpeta de la skill que las hereda.

El de lo administrativo: diez normas, una referencia (`plazos-y-territorio/references/administrativo.md`), nueve filas
y las evals de cobertura y de plazos de `plazos-y-territorio`. La vertical de ensayo: dos normas, tres referencias, una
fila y una eval, en `internal/skills/testdata/vertical-de-ensayo/`.

## 4. Región generada

| Clase | Dónde | Marcas | Contenido |
|---|---|---|---|
| tabla de comandos (hoy) | `SKILL.md` de capacidad | `<!-- inicio de la tabla de comandos: … -->` · `<!-- fin de la tabla de comandos -->` | desde `--describe` |
| paquetes de la skill | `SKILL.md` de capacidad | `<!-- paquetes: lo genera make skills-sync -->` · `<!-- fin de los paquetes -->` | una línea por paquete con referencia en la skill |
| paquetes de la base | `SKILL.md` de entrada | `<!-- paquetes de la base: lo genera make skills-sync -->` · `<!-- fin de los paquetes de la base -->` | una línea por paquete, con las skills que lo llevan |
| tabla de datos | referencia a mano | `<!-- inicio de <generador>: generada desde data/<f>[ y data/<f>] con make skills-sync, no editar -->` · `<!-- fin de <generador> -->` | la tabla del generador; `cobertura_por_materia`, filtrada por el paquete del fichero |

## 5. Línea `⚠ NO CUBIERTO:` de una fila

Una de las dos formas del cuerpo, elegida por si la fila trae `norma`, con `<procedimiento>`, `<norma>`,
`<identificador de la norma>`, `<precepto>`, `<identificador>` e `<id>` sustituidos; una por fila, en la columna
«Línea que da la respuesta» de la región `cobertura_por_materia`.

## 6. Nombre retirado ([contracts/instalacion-retiradas.md](contracts/instalacion-retiradas.md))

`instalacion.Retiradas`, `map[string][]string`: `boe-legislacion → [normativa]`, `legal-core → [plazos-y-territorio]`.
Regla: ningún nombre retirado se empotra; cada heredera, sí.

## 7. Resultado de `install` (`instalacion.SkillInstalada`; `schemas/instalacion.json`)

`{nombre, ruta, estado, enlaces, aviso?}`; `estado` ∈ {`instalada`, `actualizada`, `sin cambios`, **`retirada`**,
**`sin retirar`**}; `aviso`, solo con `sin retirar`.

Transiciones de una skill de nombre retirado declarada en el manifiesto de un ámbito:

```text
declarada, huellas que coinciden ──install──▶ retirada (directorio, hosts y entrada fuera) ──▶ no declarada
declarada, huellas que no coinciden ──install──▶ sin retirar (nada cambia, aviso) ──persona quita sus ficheros──▶ declarada, sin ficheros
declarada, sin ficheros ni hosts ──install──▶ no declarada (en silencio)
```

## 8. Eval ([contracts/evals.md](contracts/evals.md); `schemas/eval.yaml.json`)

Campos nuevos: **`se_activan`** (lista de nombres de skill) y **`no_cubierto.precepto`** (`{norma, bloque}`,
obligatorio en una eval de cobertura). Reglas nuevas: cada nombre de `se_activan` y `no_se_activan` es una skill de
`skills/`; cada skill de capacidad está en el `no_se_activan` de una eval de otra carpeta; `no_cubierto.precepto` es el
de una fila de `data/cobertura.yaml`.
