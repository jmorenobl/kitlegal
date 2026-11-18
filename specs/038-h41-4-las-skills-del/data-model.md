# Data model: H41.4

El hito no añade tipos, tablas, ficheros de `data/` ni salidas de applet. Estas son las entidades que el plan cambia o
compara, con sus reglas; los detalles, en contracts/.

## 1. Descripción de una skill

- **Qué es**: el valor de la clave `description` del frontmatter de `skills/<skill>/SKILL.md`, tal como lo lee
  `skills.LeerFrontmatter` (`internal/skills/frontmatter.go:181`).
- **En el fichero**: la línea `description: >-`, detrás de `name:`, y dos o más líneas sangradas con dos espacios; el
  bloque `>-` las lee unidas con un espacio y sin el salto final. Las líneas de la clave son las que da
  `lineasDelSkill` (`internal/evals/laboratorio.go:321`): la que empieza por `description:` y las sangradas que la
  siguen dentro del frontmatter.
- **Reglas** (FR-001, FR-002, FR-004): igual, byte a byte, al texto de su skill en la variante; en un bloque `>-` de más
  de una línea; como mucho 450 runas, 1700 entre las de todas las skills de `skills/`, 900 y ninguna etiqueta XML.
- **Valores** de la rama: contracts/skills.md §2 y §3.

## 2. Variante `producto-descripciones-cortas`

- **Qué es**: `laboratorio/producto-descripciones-cortas/<skill>.txt`, una línea y su salto por skill de `skills/`; el
  texto es el fichero sin el salto final (`textoDeLaVariante`), y lo lee `descripcionesCortas`
  (`internal/evals/laboratorio_descripciones.go:51`).
- **Regla** (FR-006): no cambia; la fija `TestDescripcionesCortas`.
- **Relación**: es lo esperado de la descripción de §1, por el nombre de la skill.

## 3. Base de la rama y huellas de `consulta-juridica`

- **Base**: `ab951c227e1fc8acac8bd6b23534e275e4a36a76`, el `main` del que sale la rama (research V3).
- **Huellas**: `ruta relativa a la skill → SHA-256 en hexadecimal` de cada fichero de `skills/consulta-juridica/` en la
  base; la del `SKILL.md`, de su contenido sin las líneas de la clave (contracts/controles.md §2.1). Un solo fichero.
- **Regla** (FR-003): los ficheros de hoy, con el `SKILL.md` sin las líneas de la clave, dan las mismas rutas y las
  mismas huellas.

## 4. Declaración del producto (sin cambios)

- **Qué es**: `declaracionesDelProducto` de `internal/evals/laboratorio_familias_test.go` para `normativa`,
  `jurisprudencia` y `plazos-y-territorio`: su base (su control o `plazos-y-territorio-sin-condicion`), los cambios de
  su cuerpo y sus movimientos (H41 y H41.1).
- **Regla nueva** (FR-005): lo esperado de `skills/<skill>/SKILL.md` es lo de su base con su declaración, y su
  frontmatter el de su control, en los dos casos fuera de las líneas de la clave; la clave lee el texto de la variante
  (contracts/controles.md §3).

## 5. Defecto de una comprobación

- **Qué es**: una línea de texto, `<ruta o skill>: <qué falla>` o `<ruta> falta|sobra|difiere …`, la forma de los
  defectos de `descripcionesCortas` y de `compararLosFicheros`.
- **Valores**: contracts/controles.md §1.1 y §2.2.
- **Ciclo**: aparece con `make ci` cuando `skills/` deja de ser lo declarado y desaparece cuando vuelve a serlo; no se
  guarda en ningún sitio.
