# 0001 · Un único ejecutable multicall

- **Estado**: aceptada
- **Fecha**: 2026-09-10
- **Hito**: H0

## Contexto y problema

`kitlegal` expone dos familias de verbos que crecen a ritmos distintos: uno por fuente legal (`boe`,
`borme`, `placsp`, `bdns`, `eurlex`, `catastro`…) y uno por capacidad transversal (`cita`, `plazos`,
`competencia`, `escrito`, `graph`, `mcp`). Y tiene cuatro consumidores a la vez: la persona en la
terminal, las skills agénticas —que hoy invocan `scripts/boe articulo BOE-A-2015-10565 a21` sobre un
script Python—, el servidor MCP y la librería Go `pkg/legalkit`.

La pregunta es cómo se empaqueta y se distribuye todo eso. Las restricciones que la condicionan:

- Las skills no llevan código: su `scripts/` son enlaces a un ejecutable, y el `SKILL.md` publica la
  orden literal que el agente debe escribir. Cambiar esa orden obliga a regenerar skills.
- Las convenciones para agentes (`--json`, `--timeout`, `--offline`, `--dry-run`, `--describe`,
  `--no-graph`, `--asunto`), el sobre de salida `{ok, fuente, url, fecha_consulta, hash, data}` y los
  códigos de salida estables tienen que ser idénticos en todos los verbos, o el agente no puede
  razonar sobre ellos.
- La caché SQLite, el cliente HTTP con rate limit por host y el grafo son infraestructura compartida:
  partirlos por fuente multiplicaría ficheros de estado y ritmos contra el mismo servidor.

## Opciones consideradas

1. **Un binario por fuente** (`kitlegal-boe`, `kitlegal-placsp`…). Cada uno se versiona y se instala
   por separado, pero el usuario acaba con una docena de artefactos que actualizar, las convenciones
   se duplican y divergen, y la caché o se comparte por acuerdo tácito o se fragmenta.
2. **Un binario único con subcomandos y sin alias** (`kitlegal boe articulo …` como única forma).
   Resuelve la duplicación, pero obliga a reescribir las órdenes de todas las skills y pierde la
   compatibilidad con el patrón `scripts/<nombre>` que ya funciona.
3. **Un binario único multicall**: el mismo ejecutable decide el applet por `os.Args[0]` o, si ese
   nombre no es un applet conocido, por el primer argumento.
4. **Seguir con scripts Python independientes**, como `boe.py`. Es el punto de partida, no el
   destino: sin binario único no hay distribución sencilla, ni tipos compartidos, ni `--describe`.

## Decisión

Se adopta la **opción 3**: un solo ejecutable `kitlegal`, multicall.

- `cmd/kitlegal/main.go` es el punto de entrada; `internal/app` mantiene el registro de applets y
  hace el dispatch. Si `os.Args[0]` coincide con un applet registrado, ese es el applet; si no, lo
  decide el primer argumento. Un symlink `boe -> kitlegal` hace que `scripts/boe articulo …` siga
  funcionando exactamente igual que con el script Python.
- La CLI se construye con Kong, y `internal/cli` concentra los flags globales, el sobre de salida y
  los códigos de salida estables (0 ok · 2 args · 3 no encontrado · 4 fuente no disponible ·
  5 rate-limited/TOS · 6 requiere identidad humana).
- `internal/cli` **no** se extrae a una librería externa hasta que al menos tres applets hayan
  repetido el patrón; hasta entonces, extraerla sería adivinar la abstracción.

## Consecuencias

**A favor**

- Una instalación, una versión, un artefacto de release por plataforma: `kitlegal version` identifica
  sin ambigüedad qué está corriendo el agente.
- Las convenciones de agente se escriben una vez y las heredan todos los applets, presentes y
  futuros; un applet nuevo no puede inventarse su propio formato de salida sin salirse del patrón.
- Caché, cliente HTTP y grafo son compartidos por construcción: un solo fichero de estado por ámbito
  y un solo ritmo de peticiones por host.
- Los enlaces por nombre son gratis: dar de alta un alias es un symlink, no un binario más.

**En contra, y asumido**

- El binario crece con cada fuente y todo el mundo paga el tamaño de todas, use las que use.
- Un fallo en el dispatch o en `internal/cli` afecta a todos los applets a la vez; eso obliga a que
  esas rutas tengan la cobertura más alta del proyecto.
- Los applets no se versionan por separado: publicar un arreglo de `boe` publica también el estado
  actual de los demás. La disciplina de `CHANGELOG.md` y SemVer es la que sostiene esto.
- **El nombre del ejecutable pasa a ser interfaz pública**: renombrar un applet o un symlink rompe
  las skills que ya publican esa orden en su `SKILL.md`. Se trata como un cambio incompatible.
- Los tests de la CLI tienen que cubrir las dos vías de dispatch —por `os.Args[0]` y por primer
  argumento—, porque son dos caminos distintos hasta el mismo applet.
