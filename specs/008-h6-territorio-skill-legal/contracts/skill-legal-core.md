# Contrato: la skill `legal-core` v0

El producto del hito (constitución, principio VIII). Sigue el molde de `skills/boe-legislacion/`: `SKILL.md` con el
protocolo, `references/` generadas desde `data/` y `scripts/` como enlaces al binario (FR-060).

## 1. `SKILL.md`

### 1.1 Frontmatter

```yaml
---
name: legal-core
description: >-
  … (qué resuelve y cuándo activarse; menos de 1024 caracteres)
metadata:
  kitlegal-applets: territorio
  kitlegal-referencias: leyes_vertebrales jerarquia_normativa
---
```

- `name` igual al directorio, en minúsculas y con guiones; `description` no vacía: lo exige `ValidarFrontmatter`.
- `kitlegal-applets` lleva **solo `territorio`** (FR-069): `legal-core` no invoca `boe`.
- `kitlegal-referencias` lleva las dos referencias generadas, separadas por un espacio.

### 1.2 Estructura

| Sección | Contenido |
|---|---|
| Título y párrafo | Qué resuelve la skill y qué no: identifica el territorio y razona con identificadores y jerarquía; no da el texto de ningún artículo |
| Protocolo | Los pasos, **empezando por identificar el territorio** (FR-061) |
| Tabla de comandos | Región generada desde `--describe`; solo los verbos de `territorio` (FR-064, FR-069) |
| Reglas | Las invariantes (§1.4) |

Menos de 300 líneas, contadas sobre el `SKILL.md` **regenerado** (research.md V25).

### 1.3 Protocolo

1. **Identificar el territorio.** Antes de razonar sobre normas, averigua de qué municipio se habla. Si la
   conversación no lo dice, **pregúntalo**; no lo supongas (FR-102: `.kitlegal/config.yaml` llega en H10).
2. **Resolverlo con el binario**, siempre con `--json`:
   `scripts/territorio resolver <nombre o código INE> --json`. Ningún dato de territorio —comunidad, provincia,
   boletines, DIR3, régimen— se da por sabido ni se escribe de memoria (FR-062).
3. **Leer `cobertura` y trasladarla a la respuesta.** Lo que la salida declare no configurado o no verificado se dice
   explícitamente, y **no se nombra ningún boletín que el applet no haya devuelto** (FR-062, FR-101).
4. **Ambigüedad y ausencia.** Con código 2 y una lista de candidatos, se ofrecen los candidatos y se pregunta; con
   código 3 se dice que ese municipio no está en la relación, **sin concluir que no exista** lo que no se ha podido
   comprobar.
5. **Razonar con las referencias.** `references/leyes_vertebrales.md` da el identificador de cada norma vertebral;
   `references/jerarquia_normativa.md`, qué nivel regula qué, dónde se publica y las reglas de interpretación.
6. **Delegar el texto.** Citar el identificador de una norma basta con la referencia; **afirmar lo que dice un
   artículo exige consultarlo con `boe-legislacion` en esa misma conversación** (FR-069). La delegación va en un solo
   sentido: `legal-core` no lleva los verbos de `boe` en su tabla ni su enlace en `scripts/`.

### 1.4 Reglas invariantes

Las de todas las skills (FR-063), escritas en la skill:

1. Nunca inventar contenido legal ni citar de memoria.
2. Cada afirmación sobre una norma, con su identificador; el texto de un artículo, solo desde `boe-legislacion`.
3. Distinguir ley de reglamento.
4. Señalar la variación autonómica: lo que una comunidad puede haber regulado de otro modo.
5. No concluir «no existe» a partir de un resultado sin cobertura completa.
6. Ninguna acción con identidad: ni presentar, ni notificar, ni firmar, ni tramitar, ni simularlo.

Además, y como en `boe-legislacion` (FR-077 de H5): `SKILL.md` no menciona evals, el job, modelos ni
`KITLEGAL_CACHE_DIR`, y la comprobación mecánica lo vigila en `SKILL.md` **y en las dos referencias generadas**
(research.md V25).

## 2. Referencias generadas

| Fichero | Se genera desde | Contenido |
|---|---|---|
| `references/leyes_vertebrales.md` | `data/normas.yaml`, normas con `vertebral: true` | Tabla con norma, abreviatura, identificador `BOE-A-…`, rango y materias |
| `references/jerarquia_normativa.md` | `data/jerarquia.yaml` | Tabla de niveles (UE, Estado, comunidad autónoma, provincia, municipio) con su boletín y sus tipos de norma, y la lista de reglas de interpretación |

- Cabecera obligatoria y literal en la primera línea: `<!-- generado desde data/<fichero>.yaml, no editar -->`
  (FR-065, FR-066).
- **`data/` es la única fuente de verdad**: `make skills-sync` las regenera y `make ci` falla si difieren de lo
  commiteado (FR-067).
- Ninguna referencia lleva el texto de una norma (FR-069).

## 3. Enlaces

`skills/legal-core/scripts/territorio` → `../../../bin/instalado/kitlegal`, el mismo destino que usa
`boe-legislacion` (`internal/skills/enlaces.go`). Lo crea y lo comprueba la sincronización; `make install` lo deja
apuntando al binario instalado.

## 4. Generación: lo que cambia en `internal/skills`

La tabla de generadores sustituye al `switch` por nombre y rompe el acoplamiento «nombre de la referencia = nombre del
fichero de datos» (research.md D20, V22):

| Referencia | Fichero de datos | Generador |
|---|---|---|
| `normas` | `data/normas.yaml` | todas las normas |
| `leyes_vertebrales` | `data/normas.yaml` | solo `vertebral: true` |
| `jerarquia_normativa` | `data/jerarquia.yaml` | niveles y reglas |

- La cabecera, el título y las columnas dejan de ser constantes de normas y se derivan de la tabla (V23).
- `defectosDeLasReferencias` consulta la tabla en vez de componer `data/<nombre>.yaml`, de modo que una referencia
  declarada sin generador sigue siendo un defecto con su mensaje.
- `Norma` gana el campo `Vertebral bool` y `schemas/normas.yaml.json`, la propiedad `vertebral` (tarea `[datos]`,
  FR-085); el cambio es **compatible hacia atrás**: ninguna norma ya válida deja de serlo.
- En esa misma tarea —el paso 16 del plan— se regenera `skills/boe-legislacion/references/normas.md`, que sale del
  fichero entero y por tanto gana las siete normas nuevas (FR-074). Es lo **único** que cambia de `boe-legislacion`, y
  cambia solo por regeneración. Las dos referencias de `legal-core` no se generan aquí: nacen con la skill, en el paso
  21, porque es su frontmatter el que las declara.

## 5. Qué lo vigila

| Control | Test | En `make ci` |
|---|---|---|
| Frontmatter, 300 líneas, deriva de `references/`, de la tabla y de los enlaces | `TestSkillsDelRepositorio` (recorre todas las skills) | sí |
| Que la skill existe, es decir que los controles anteriores no pasan en vacío | `TestSkillsDelRepositorio/skills`: `legal-core` entra en `skillsExigidas` en el paso 21, **antes** que sus ficheros (D26; plan, obligación 12) | sí |
| Que los casos negativos también se ejercen sobre `legal-core` | `TestSkillsDelRepositorio`, casos parametrizados por skill, en `internal/app/skills_test.go` (V25; plan, *Complexity Tracking* y orden de implementación) | sí |
| Las dos referencias regeneradas sin deriva | `TestRegenerarYComparar`, `make skills-sync` idempotente | sí |
| Que la tabla de comandos solo lleva verbos de `territorio` | `TestSkillsDelRepositorio/skills` + `TestTablaDeComandosCoincideConLaGramatica` | sí |
| Que ninguna norma nombrada en `SKILL.md` falta en `data/normas.yaml` | `TestSkillsDelRepositorio/normas-nombradas` | sí |
| Que no hay instrucciones de evals en `SKILL.md` ni en las referencias | `TestSkillsDelRepositorio/sin-instrucciones-de-evals` | sí |
| Que el protocolo empieza por el territorio y delega el texto | evals (§ contrato de evals) y los dos jueces de la revisión final (constitución, capa 2) | evals / revisión |
