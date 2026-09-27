# 0023 · El contrato de resultados: qué devuelve cada orden en cada situación

- **Estado**: aceptada
- **Fecha**: 2026-09-27
- **Hito**: transversal (constitución 2.4.0, antes de H7). Amplía el ADR 0006 (una séptima clase de fallo y el lugar
  de los hallazgos) y precisa el ADR 0011 (el código de salida del ensayo).

## Contexto y problema

El ADR 0006 fija la forma del sobre y seis clases de fallo, cada una con su código de salida, y da por hecho que
todo lo que no es un éxito es un fallo. No dice qué es cada resultado posible, y cada hito lo ha resuelto por su
cuenta:

- H19 escribió «`skills doctor`: exit 0 sin hallazgos, 1 con ellos». El 1 es, por el ADR 0006, el fallo
  inesperado, así que los hallazgos de `doctor` salían como un defecto del programa, con `data` reducido a un texto.
- `skills install` ante una carpeta que no es suya, `list` y `doctor` ante un manifiesto ilegible y `-g` sin `HOME`
  salían también como inesperados: situaciones previstas y que la persona resuelve, confundidas con un defecto.
- H7 pedía «`graph check` … exit 0/1/2»: con el contrato de entonces, `boe-legislacion` habría leído el hallazgo que
  más importa —que una norma ha cambiado— como que la herramienta falló, y habría respondido «no pude comprobarlo».
- El ADR 0011 deja el ensayo en 0 «salvo que la invocación sea inválida por argumentos o configuración», sin decir si
  un conflicto con el estado local cuenta.

Faltaba un contrato: una sola tabla que diga, para cada resultado, qué código, qué `ok` y qué `data`. Sin ella,
cada hito vuelve a improvisar, y un agente —el consumidor principal del binario— no puede leer un resultado sin
conocer el verbo.

## Opciones consideradas

1. **Dejarlo como estaba** (hallazgos y conflictos como inesperados). Rechazada: un agente no distingue un defecto
   de un resultado, y los hallazgos pierden su estructura.
2. **El 1 para los hallazgos**, como los linters. Rechazada: el 1 ya es el fallo inesperado, y compartir código hace
   imposible distinguir «encontró algo» de «se rompió», que es justo lo que un consumidor necesita saber; además,
   con `ok` falso el `data` solo puede ser `{clase, mensaje}` (ADR 0006).
3. **Un código propio para los hallazgos.** Rechazada: un hallazgo no es un fallo —la verificación hizo lo que se le
   pidió—, y darle un código distinto de 0 obliga a presentar como fallo lo que es la respuesta.
4. **Un contrato por clase de resultado**: lo que se hizo sale con 0 y sus datos, también cuando es una verificación
   que encuentra algo; lo que no se pudo hacer sale con la clase de su causa, y la causa «el estado local lo impide»
   tiene clase y código propios.

## Decisión

Opción 4. **Todo verbo de todo applet devuelve uno de estos resultados, y solo uno:**

| Resultado | Código | `ok` | `data` |
|---|---|---|---|
| **Hecho**: la orden hizo lo que se pidió —consulta, cálculo, instalación, o una verificación, con hallazgos o sin ellos— | 0 | `true` | el resultado del verbo; los hallazgos, como lista estructurada |
| **Argumentos**: la invocación no es válida | 2 | `false` | `{clase: "argumentos", mensaje}` |
| **No encontrado**: lo pedido no existe en la fuente | 3 | `false` | `{clase: "no-encontrado", mensaje}` |
| **Fuente no disponible**: la fuente no responde, o sin red y sin caché | 4 | `false` | `{clase: "fuente-no-disponible", mensaje}` |
| **Límite o TOS** | 5 | `false` | `{clase: "limite-o-tos", mensaje}` |
| **Identidad humana**: la acción exige identidad y no se ha hecho | 6 | `false` | `{clase: "identidad-humana", mensaje}` |
| **Conflicto**: la orden reconoce y nombra algo del estado que gestiona la persona que le impide actuar —una entrada que no es suya, un fichero cuyo contenido no tiene la forma que debe, el entorno sin lo que la orden pide— | 7 | `false` | `{clase: "conflicto", mensaje}` |
| **Inesperado**: un defecto del programa o del entorno —un error del sistema que la orden no interpreta, lo empotrado ilegible, un almacén interno inutilizable, una clase no declarada— | 1 | `false` | `{clase: "inesperado", mensaje}` |

Reglas que se siguen de la tabla:

- **Un hallazgo es un resultado, no un fallo.** Un verbo que verifica —`skills doctor`, `graph check` desde H7, las
  reglas de anomalías después— sale con 0 aunque encuentre algo, con cada hallazgo en `data` como objeto (su clase,
  lo que identifica —ruta, id— y lo que lo explica o lo arregla). Solo sale distinto de 0 cuando **no ha podido
  verificar**, con la clase de la causa. No hay un modo que convierta los hallazgos en un código distinto de 0:
  ningún consumidor lo necesita hoy (constitución, principio V); si llega a necesitarlo, será una bandera explícita
  y un ADR.
- **La frontera entre `conflicto` e `inesperado` es mecánica**, para que ningún hito tenga que interpretarla:
  - es **conflicto** lo que la propia orden **reconoce y nombra**, con su ruta o su variable, en el estado que
    gestiona la persona: su proyecto, lo que tiene instalado, su entorno. Se arregla mirando el mensaje: quitar una
    carpeta, definir `HOME`, rehacer un manifiesto que no tiene su forma;
  - es **inesperado** todo **error del sistema que la orden no interpreta** —un permiso denegado, un fallo de
    entrada y salida, una ruta que el sistema rechaza al examinarla—, venga de donde venga, y todo fallo de los
    **almacenes internos** de kitlegal (`cache.db`, y `world.db` desde H7). Esos almacenes son del programa, no de
    la persona: si uno no se puede abrir o es de otro esquema, lo que corresponde es que el programa lo resuelva
    —migrarlo o rehacerlo—, y si no puede, es un defecto suyo.
  - Así, `skills doctor` sale con 7 si `kitlegal.json` es un directorio o no respeta su forma, y con 1 si el
    sistema no deja leerlo; con 7 si la ruta de `--dir` es un fichero, y con 1 si el sistema no deja examinarla.
- **Una fuente pública nunca produce `conflicto`**: el único estado local que tiene es la caché, que es interna.
  `boe` y `territorio` siguen con sus clases (y `boe` nunca con 6 ni con 7).
- **El ensayo sale con la clase de cualquier fallo que se conozca sin efectos**, igual que la orden real: 2 ante
  argumentos inválidos, 7 ante un conflicto, 4 si pide la fuente sin red y sin caché, 1 ante un defecto; y 0 si
  nada lo impide. Lo demás del ADR 0011 no cambia.
- **`data` de fallo sigue siendo exactamente `{clase, mensaje}`** (ADR 0006). La lista de un conflicto va en el
  mensaje, una entrada por línea y en orden determinista.
- **La correspondencia clase ↔ código sigue siendo cerrada y única** (`internal/cli`, un `switch` sin rama por
  defecto que el linter `exhaustive` vigila): añadir una clase exige su código y un ADR.

Lo que cambia con esta decisión:

- `schema.Clases()` gana `conflicto`, que va al `enum` del sobre de fallo de todo `--describe` y de `schemas/*.json`;
  `internal/cli` gana `ErrConflicto` y el código 7.
- `skills doctor` sale con 0 y con los hallazgos en `data.hallazgos` (FR-065 a FR-067 de H19: cada uno con `clase`,
  `ruta` y `orden`); desaparece el error que los llevaba.
- `skills install` con conflictos, `list` y `doctor` con el ámbito ilegible (una guarda que no es un directorio, un
  manifiesto que no es un fichero regular o no respeta su forma, o que declara entradas de host con `--dir`), y `-g`
  sin `HOME` salen con 7 y clase `conflicto`; con `--dry-run`, igual. Un error del sistema al leer el manifiesto
  deja de ser un manifiesto ilegible: sale tal cual, con 1, como ya salía un error al examinarlo.
- La suite activa de H19 (`internal/app/testdata/script/h19-*.txtar`) se adapta a este contrato. La copia congelada
  en `specs/009-h19-instalar-sin-clonar/aceptacion/` queda como registro de lo que se congeló en H19.
- H7 se reescribe en `docs/ROADMAP.md`: `graph check` da `version-obsoleta` y `fuente-caducada` como hallazgos en
  `data`, con código 0, y un `world.db` inutilizable es un defecto (1), como la caché. Donde otros documentos dicen
  otra cosa —la sección de H19 del roadmap («exit 1 con hallazgos»), `refs/kitlegal-grafo.md` («`graph check`: 0
  limpio, 1 avisos, 2 errores»)—, manda este ADR, y quedan anotados.

## Consecuencias

**A favor**

- Un consumidor lee cualquier resultado con una sola regla: 0 es la respuesta, y sus datos dicen qué hay; distinto
  de 0 es que no se pudo, y la clase dice por qué y si está en su mano arreglarlo.
- Una skill puede actuar sobre un hallazgo —«esta norma ha cambiado desde que la consultaste»— en lugar de leerlo
  como un fallo de la herramienta.
- Los hitos siguientes no deciden códigos: eligen la fila de la tabla.

**En contra, y asumido**

- Es un cambio incompatible para quien dependiera de que `doctor` saliera con 1 al encontrar algo o de que un
  conflicto saliera con 1. El proyecto está en `0.x`, y el cambio sube el número menor (v0.2.0).
- Un guion que quiera fallar cuando `doctor` encuentra algo tiene que mirar `data.hallazgos` (por ejemplo, con
  `jq -e '.data.hallazgos == []'`) en lugar del código de salida.
