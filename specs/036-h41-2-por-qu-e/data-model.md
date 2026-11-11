# Data model: H41.2

Todo en `internal/evals`, que no va en el binario `kitlegal`. Lo que no se nombra aquí no cambia.

## 1. La activación y su lectura (`internal/evals/activaciones.go`, NUEVO; FR-001)

| Elemento | Qué es | De dónde |
|---|---|---|
| `Activacion{Skill string; Antes bool}` | Un `tool_use` de `Skill`: su `input.skill` tal cual y si va en un turno anterior al de la primera herramienta de kitlegal | nuevo; claves JSON `skill`, `antes` |
| `usoDelTranscript{turno, nombre, entrada}` | Un `tool_use` con el turno de su respuesta | movido de `laboratorio_tabla.go` |
| `leerLosUsos(directorio)` | Los usos del `sesion.jsonl`, con el turno de H39; ya no exige `message.id` (research D1) | movido, con ese cambio |
| `usosDelMensaje`, `bloqueConTipo`, `conResultado`, `activaAntes`, `skillDelUso` | Lo que usa `leerLosUsos` y la regla de «antes» | movidos, sin cambiar |
| `reconocedorDeKitlegal{herramientas, interprete}` y su `esDeKitlegal(uso, modo)` | La primera herramienta de kitlegal: en `herramienta`, una del registro; en `orden`, una orden de `Bash` que consulta con `kitlegal`; en otro modo, ninguna (research D2) | la regla de `esDeKitlegal` y `consultaKitlegal` de la tabla, en un tipo; `herramientasDelRegistro()` y `nuevoInterprete()` lo construyen |
| `activacionesDeLosUsos(usos, esDeKitlegal)` | La lista, en el orden de los usos: el turno de la primera herramienta y, por cada uso de `Skill`, su `Activacion`. Nunca nil: `[]` sin ninguna | nuevo |
| `textoDeLasActivaciones(lista)` | `normativa (antes), consulta-juridica (después)`; `ninguna` con la lista vacía | nuevo; lo usan `informe.md` y el detalle de la tabla |

Reglas (FR-001): turno, el de H39 (empieza con el primer `assistant` y con cada `assistant` que llega tras un `user` con
algún `tool_result` y lleva un `message.id` distinto del anterior); «antes», un turno anterior al de la primera
herramienta, o cualquiera si no hay ninguna; en el mismo turno, no. Cada `tool_use` de `Skill` es una entrada, también
si repite skill.

## 2. El informe del job (`internal/evals/informe.go`; FR-001 a FR-005)

- `ResultadoDeEval` gana `Activaciones []Activacion` con `json:"activaciones,omitzero"`, detrás de `Activada`.
  - Sesión que `LeerSesion` leyó: la lista de `activacionesDeLosUsos` con los usos de `leerLosUsos` y el reconocedor en
    el modo de la sesión; `[]` si no activó ninguna.
  - Sesión que `LeerSesion` no leyó, o cuya lista no se pudo leer: nil, sin clave en `informe.json`. Un error al leerla es
    un motivo `sesión ilegible: <error>` (research D5).
- `EscribirInforme` construye el reconocedor una vez; sus errores, como los de hoy al preparar el informe.
- La sección de cada sesión de `informe.md` gana el párrafo `Skills activadas: …` (research D4).
- Nada más cambia: `Activada`, la tabla de sesiones, las tasas, los umbrales, los motivos y el veredicto se calculan como
  hoy (FR-004).

## 3. El arnés (`internal/evals/laboratorio_sondeo.go`, `laboratorio_tabla.go`; FR-010 a FR-014)

- `ArgumentosDelLaboratorio` gana `Instalacion` y `Eval`, detrás de `Directorio`, tal como llegan.
- `laboratorioComprobado` gana `conElProducto bool` (`INSTALACION=producto`) y `eval string`, el fichero de la eval pedida
  (vacío sin `EVAL`); `porEval` es `SESIONES` con `EVAL` y `SESIONES/4` sin él.
- `sondeoLeido` gana `SkillsInstaladas []string` con `json:"skills_instaladas,omitempty"`, detrás de
  `ReferenciasDelProducto`.

`sondeo.json` (una línea, con su salto final):

| Sondeo | Contenido |
|---|---|
| de H39 | `{"modelo":…,"claude_code":…}`, como hoy |
| de una familia del producto, sin `<skill>-producto` | `{"modelo":…,"claude_code":…,"skills_instaladas":[…]}` |
| con `<skill>-producto` | `{"modelo":…,"claude_code":…,"referencias_del_producto":[…],"skills_instaladas":[…]}` |

`skills_instaladas`: sin la opción, `["<skill>"]`; con ella, los cuatro nombres de `skills/` en orden de nombre. De 89 a
349 bytes (research V16).

El temporal, en una familia del producto: `skills/<variante>/` tiene la copia de la variante (como hoy) y, con la opción,
la de cada otra skill de `skills/`. Cada sesión enlaza cada entrada.

## 4. La tabla (`internal/evals/laboratorio_tabla.go`; FR-020 a FR-023)

- `claveDeFila` gana `instaladas string`: los nombres de `skills_instaladas` unidos por «, », o la skill de la familia si
  el sondeo no la tiene. `compararClaves` ordena por variante, modo, modelo, versión e instalación, en ese orden.
- `sesionContada` gana `activaciones []Activacion`, la de `activacionesDeLosUsos` con el reconocedor en el modo de la
  sesión; solo de una sesión terminada, como lo demás que cuenta.
- `contadorDelLaboratorio` guarda un `reconocedorDeKitlegal` en lugar de sus `herramientas` e `interprete`.
- `familiaDeLaTabla` y `celdasDelProducto` ganan `conLasActivaciones bool`: verdadero si algún sondeo de la tabla tiene
  `skills_instaladas`. Con él, las columnas de contracts/tabla.md §3; sin él, las de hoy. La de H39 no lo usa.

Recuentos de una fila con `n` > 0 (FR-020): «Activan otra skill», las terminadas con alguna activación de una skill que
no es la de la familia, en cualquier turno; por cada otra skill, las terminadas que la activan; «No activan ninguna»,
las terminadas con la lista vacía. Una sesión cuenta una vez en cada recuento.

## 5. Fixtures (tareas `[datos]`, sin red)

| Ruta | Qué | Para |
|---|---|---|
| `internal/evals/testdata/sesiones/activaciones/` | `evals/` con una eval y `sesiones/` con diez sesiones, cinco por modo (contracts/informe.md §5) | FR-006 |
| `internal/evals/testdata/laboratorio/familias/activaciones/sondeo/{h41,h41-2-a,h41-2-b}/` y `…/activaciones/tabla.md` | tres sondeos de `normativa-producto` en modo herramienta y su tabla esperada (contracts/tabla.md §5) | FR-020 a FR-022, FR-024 |

Ningún fichero existente de `testdata/` ni de `schemas/` cambia.
