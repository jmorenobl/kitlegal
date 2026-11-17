# Data model: H41.3

Lo que el hito añade a lo de H41.2 (`specs/036-h41-2-por-qu-e/data-model.md`). Todo vive en `internal/evals`, que no va
en el binario, salvo la función exportada de §1, en `internal/skills`. Ningún tipo exportado nuevo en `internal/evals`.

## 1. La variante de instalación `producto-descripciones-cortas`

- **Dónde**: `laboratorio/producto-descripciones-cortas/`, un fichero `<skill>.txt` por skill de `skills/`
  (contracts/variante.md §1).
- **Texto de una `description`**: el fichero sin su salto de línea final (la regla de `textoDeLaVariante`).
- **Límites** (FR-001, FR-002): cada uno, como mucho 450 runas; todos, como mucho 1700; ninguno con una etiqueta XML, la
  primera que da `skills.PrimeraEtiquetaXML(texto) string` (`internal/skills/frontmatter.go`, la expresión `etiquetaXML`
  de hoy; `defectosDeLaEtiqueta` pasa a usarla).
- **Leída** por `descripcionesCortas(raiz) (map[string]string, []string, error)` (`laboratorio_descripciones.go`): el
  texto de cada skill de `<raiz>/skills/` por su nombre, los defectos de contracts/variante.md §3 en su orden, y el error
  de un directorio que no se puede leer (defecto `inesperado`, ADR 0023).
- Constantes, junto a la familia de `normativa` en `laboratorio_familias.go`: `varianteDelProductoDeNormativa =
  "normativa-producto"`, que la declaración de la familia usa en sus dos sitios, e `instalacionConLasDescripcionesCortas
  = "producto-descripciones-cortas"`, la única variante con la que vale; y en `laboratorio_descripciones.go`, los límites
  (450 y 1700) y la carpeta de la variante.

## 2. Los argumentos del arnés

- `ArgumentosDelLaboratorio` gana `Presupuesto string`, el décimo, detrás de `Eval`, tal como llega.
- `laboratorioComprobado` gana:
  - `instalacion string`: `INSTALACION` comprobada —vacía, `producto` o `producto-descripciones-cortas`—; `conElProducto`
    es que no esté vacía (las dos copian todas las de `skills/`);
  - `descripciones map[string]string`: con `producto-descripciones-cortas`, el texto de cada skill de `skills/`; si no,
    nil;
  - `presupuesto int`: el de `PRESUPUESTO`, o 0 sin él.
- Reglas y mensajes: contracts/arnes.md §2. `comprobar` recibe además la raíz, para leer la variante solo con
  `INSTALACION=producto-descripciones-cortas`; sus errores siguen juntos con `errors.Join` y los envuelve en `errorDeUso` el
  propio `comprobar` (`comprobarElLaboratorio` los devuelve tal cual), salvo el de una variante que no se puede leer,
  que `comprobar` devuelve sin envolver porque no es un error de uso (como el de una declaración que no se puede leer).

## 3. La instalación compuesta

`skills/normativa-producto/` del temporal: la copia de `INSTALACION=producto` (todas las de `skills/`, cada una en su
directorio) con la `description` de cada `SKILL.md` sustituida por `skillConLaDescripcion(contenido, texto)`
(`laboratorio_descripciones.go`; research D5). Qué líneas son de la `description` lo dice la clasificación que comparten
`skillConLaDescripcion` y `descripcionYCuerpo` (`laboratorio.go`). No se guarda en `DIRECTORIO`.

## 4. El entorno de cada sesión

El de hoy, `entornoDelSondeo(s.Entorno, temporal)`, con `SLASH_COMMAND_TOOL_CHAR_BUDGET=<presupuesto>` detrás si
`presupuesto > 0`; `entornoDeLaSesion` pone encima lo de cada sesión, sin cambios (research D8).

## 5. `sondeo.json`

`sondeoLeido` (`laboratorio_tabla.go`), el único tipo con que lo escribe el arnés y lo lee la tabla, gana detrás de
`SkillsInstaladas`:

| Campo | Clave | Tipo | Cuándo |
|---|---|---|---|
| `Instalacion` | `instalacion` | cadena, `omitempty` | `"producto-descripciones-cortas"` con esa opción |
| `Presupuesto` | `presupuesto` | entero, `omitempty` | el de `PRESUPUESTO` |

Sin las dos opciones, el fichero es el de hoy byte a byte (contracts/arnes.md §5).

## 6. La fila de la tabla

- `claveDeFila` gana `instalacion string` y `presupuesto int`, que solo se llenan si la tabla lleva la constancia; si no,
  van vacíos en todas las claves y la tabla es la de hoy.
- `familiaDeLaTabla` gana `conLaInstalacion bool`: alguno de los sondeos tiene `instalacion` o `presupuesto`. Con él,
  la cabecera y cada fila dan «Instalación» y «Presupuesto» detrás de «Skills instaladas» (contracts/tabla.md §3).
- La instalación de un sondeo: `instalacionDeLaFila(leido, skill)`, su `instalacion` o, sin ella, `producto` si
  `skills_instaladas` tiene alguna skill además de `skill`, y si no, `skill` (research D11). El presupuesto, el suyo; en la
  celda, vacío si es 0.
- `compararClaves` ordena por variante, modo, modelo, versión, instalación, presupuesto y skills instaladas.
