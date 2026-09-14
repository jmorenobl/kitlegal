# 0015 · El puerto `core.Source` y la fecha de consulta que declara quien consulta

- **Estado**: aceptada
- **Fecha**: 2026-09-13
- **Hito**: H4; amplía `schema.Procedencia` sin tocar `Applet` (ADR 0005) ni `Resultado` (ADR 0011, ADR 0014)

## Contexto y problema

H4 trae el primer adaptador de fuente pública, `internal/source/boe`, y con él dos preguntas que los hitos
anteriores dejaron abiertas a propósito.

**La firma del puerto.** `CLAUDE.md` y el §2 del roadmap nombran el puerto de fuente como
`Source{Name, Fetch(ctx,req), TTL, Terms}`, sin fijar parámetros, y `internal/core/doc.go` anunciaba que
nacería con el hito que lo estrenase. Al bajar a `boe` aparecen tres cosas que esa frase no resuelve:

- la fuente tiene que honrar `--offline` y `--dry-run` (FR-092, FR-094), que llegan en el contexto de
  ejecución que el kernel ya entrega a los applets y a `httpx.Pedir`;
- la vigencia de lo guardado no es una por fuente sino una por verbo —300 s para `buscar` y `metadatos`,
  7 días para `indice`, `articulo`, `articulos` y `analisis`— (FR-091);
- los términos de uso y el día en que una persona los revisó tienen que poder compararse con la fila de la
  fuente en `docs/SOURCES.md` (FR-121).

**La fecha de consulta.** FR-096 exige que la `fecha_consulta` de un sobre de éxito sea el instante en que
se consultó la fuente para obtener lo que `data` contiene, **no el de la invocación**, también cuando `data`
sale, entero o en parte, de la caché, con `--offline` o sin él; y que, cuando `data` se sostiene sobre
consultas hechas en instantes distintos, lleve **el más antiguo**, para que la cita nunca aparente más
frescura que su parte más vieja. Un sobre de fallo se fecha igual que uno de éxito (ADR 0006): con el
instante de la petición que falló, la misma cuya `url` lleva, o —si no llegó a construirse ninguna
(argumentos inválidos, código 2) o no hacía falta (`--offline` sin entrada vigente, código 4)— con el
instante en que el kernel monta el sobre.

Hasta H3, `cli.Montador` fechaba **todo** sobre con su reloj al montarlo. Era correcto mientras ningún
applet consultaba nada, y deja de serlo en cuanto la caché de H3 sirve un artículo guardado hace seis días
como si se acabara de pedir. La huella no lo remedia, porque no depende de la fecha (ADR 0006):
`fecha_consulta` es el único campo del sobre que declara la antigüedad de lo citado, y sin ella no hay
cita (constitución §II).

El kernel no puede saber ese instante: lo sabe quien consultó. Y el applet no puede fechar el sobre,
porque no lo monta (ADR 0005). El spec deja el mecanismo al plan y pide que, si amplía `Procedencia` o
`Resultado`, lo haga de forma retrocompatible y con un ADR, como hizo el ADR 0011 con `Ensayo`, sin
cambiar la interfaz `Applet`.

## Opciones consideradas

### La firma del puerto (research.md D2)

1. **`Fetch(ctx, req)` con el contexto de ejecución dentro de `req`**, la letra de `CLAUDE.md`. Rechazada:
   mezcla lo que se pide con cómo se invoca, y rompe la simetría con `httpx.Pedir(ctx, ec, peticion)`, que
   ya recibe el contexto de ejecución aparte.
2. **`TTL()` sin argumentos**: una sola vigencia por fuente. Rechazada: contradice FR-091.
3. **La consulta como `any` o como `map[string]string`.** Rechazada: pierde el tipo y obliga a validar
   formas en tiempo de ejecución.
4. **El registrador de eventos como parámetro de `Fetch`**, como en `Argumentos.Ejecutar`. Rechazada:
   `internal/core` no puede importar `log/slog` (R1, que vigilan la lista `core` de `depguard` y
   `compruebaDominioPuro`).
5. **`Fetch(ctx, ec, consulta)` y `TTL(consulta)`, con una `Consulta` que cada fuente tipa.**

### La fecha de consulta (research.md D3)

- **a. Un campo en `Resultado`** (`Resultado.FechaConsulta`). Rechazada: la fecha califica a la
  procedencia —de dónde y cuándo—, y en H7 la `Procedencia` es el `source` de cada nodo y arista del grafo
  (ADR 0014), que necesitará el instante para la regla `fuente-caducada`; separarlas obligaría a casar dos
  valores.
- **b. La fecha dentro de `data`.** Rechazada: duplica un campo del sobre y hace depender la huella del
  instante (ADR 0006).
- **c. El reloj del kernel siempre**, como hasta H3. Rechazada: hace pasar por recién consultado lo servido
  desde la caché, que es justo lo que FR-096 prohíbe.
- **d. Que el applet monte el sobre** con su fecha. Rechazada: lo prohíbe el ADR 0005.
- **e. Un campo en `Procedencia`** cuyo valor cero conserve el comportamiento de hoy.

## Decisión

Se adoptan la **opción 5** para el puerto y la **opción e** para la fecha.

### El puerto

```go
// internal/core/source.go — dominio puro: importa context, time e internal/core/schema.
type Source interface {
    Name() string
    Fetch(ctx context.Context, ec schema.Contexto, consulta Consulta) (schema.Resultado, error)
    TTL(consulta Consulta) time.Duration
    Terms() Terminos
}

type Consulta interface {
    Verbo() string
}

type Terminos struct {
    URL       string    // dirección de los términos de uso
    Revisados time.Time // día de la revisión, 00:00 UTC
}
```

- `Name` es el nombre de la fuente tal como va en `fuente` del sobre, y nunca empieza por `kitlegal.`
  (ADR 0006).
- `Fetch` recibe el contexto de ejecución como segundo parámetro, igual que `httpx.Pedir`, y la fuente
  honra `--offline` y `--dry-run`. En fallo devuelve **también** un `Resultado` cuya procedencia nombra la
  petición que falló (FR-101), o el valor cero si no llegó a construirse ninguna; sus errores declaran su
  clase (`schema.ConClase`).
- `TTL` recibe la consulta porque la vigencia es por verbo (FR-091).
- `Terms` declara la dirección de los términos de uso y el día en que una persona los revisó (FR-121); cada
  adaptador los ata con un test a su fila de `docs/SOURCES.md`.
- Cada fuente declara sus tipos de consulta, uno por verbo. Una consulta de un tipo ajeno es un fallo
  inesperado: defecto de quien compone, nunca de quien invoca.
- El registrador de eventos no viaja por el puerto: el adaptador lo recibe al construirse, uno por
  invocación.

El applet de la fuente vive en `internal/app`, construye la fuente con sus dependencias y devuelve lo que
responde `Fetch` tal cual. `Applet`, `Verbo`, `Argumentos` y la firma de `Main` no cambian.

### La fecha de consulta

```go
// internal/core/schema
type Procedencia struct {
    Fuente        string
    URL           string
    FechaConsulta time.Time // valor cero: «no la declara quien consulta»
}
```

- **Valor cero = el kernel fecha con su reloj al montar**, que es exactamente lo que hacía hasta H3.
  `Procedencia.Validar` no cambia: la fecha no es obligatoria, y tampoco hace citable una procedencia que
  no lo es.
- `cli.Montador` fecha el sobre así, en éxito y en fallo por el mismo camino (ADR 0006: `fecha_consulta` y
  `hash` «se calculan igual que en un sobre de éxito»):

  | Procedencia recibida | `fuente`, `url` del sobre | `fecha_consulta` del sobre |
  |---|---|---|
  | válida, `FechaConsulta` no cero | las recibidas | `FechaConsulta` |
  | válida, `FechaConsulta` cero | las recibidas | reloj del montador al montar |
  | inválida o cero (fallo del kernel, código 2 del applet) | `kitlegal.cli`, `kitlegal:cli` | reloj del montador al montar |

  Una procedencia inválida se sustituye por la del kernel, que no trae fecha, de modo que el sobre se fecha
  con el reloj aunque la rechazada trajera una.
- **La huella no cambia**: se sigue calculando sobre `data`, que no lleva la fecha.
- **Quien consulta declara cuándo.** `boe` rellena la fecha con el instante de emisión de la petición que
  le da `internal/httpx` —el de su último intento, o el del abandono si no llegó a salir— y, cuando el
  contenido sale de la caché, con el instante guardado en la entrada. Si `data` se sostiene sobre varias
  consultas, lleva el más antiguo; en un fallo, el de la petición que falló (FR-096).

## Consecuencias

**A favor**

- Una cita servida desde la caché declara la antigüedad real de lo citado: quien la lee ve cuándo se
  consultó la fuente, no cuándo se invocó el binario.
- Retrocompatible: los applets de ejemplo, el adaptador de prueba de H3 y todo sobre de fallo anterior a
  una petición no declaran fecha, y el kernel los fecha como antes sin tocar una línea suya.
- El contrato del ADR 0005 queda intacto y `Resultado` sigue siendo `{Procedencia, Datos, Ensayo}`
  (ADR 0011). Lo que crece es `Procedencia`, con un campo cuyo valor cero es válido: la misma forma de
  crecer que el ADR 0014 prevé para `Resultado`.
- La fecha viaja con la procedencia a la que califica. Cuando H7 aplique las operaciones de grafo con la
  `Procedencia` del resultado como `source` (ADR 0014), el instante ya estará ahí.
- `internal/core` sigue siendo puro: `source.go` solo importa `context`, `time` y `schema`, y no tiene
  sentencias.
- Todo adaptador futuro implementa la misma interfaz y fecha sus sobres por el mismo camino: la regla de
  FR-096 no hay que volver a decidirla por fuente.

**En contra, y asumido**

- Que `fecha_consulta` sea correcta pasa a depender de cada adaptador. Uno que no declare la fecha de lo
  que sirve desde su caché no rompe el sobre —el kernel lo fecha—, pero lo fecha mal. El control es de cada
  adaptador (en `boe`, `TestFechaDeConsultaDeArticulo` y los tests de caché de sus seis verbos);
  `TestMontadorFechaDeConsulta` fija solo la regla del kernel.
- El valor cero no distingue «la fuente no conoce la fecha» de «la fuente olvidó declararla»: por contrato
  significa lo primero.
- La firma de `Source` se fija con un solo adaptador. El riesgo es bajo porque lo que la concreta —contexto
  de ejecución aparte, vigencia por consulta y consulta tipada— no depende de nada propio del BOE; si una
  fuente futura necesitara otra cosa, se decidirá con su propio ADR.
- `CLAUDE.md` resume el puerto como `Source{Name, Fetch(ctx,req), TTL, Terms}`. Es un resumen; la firma
  concreta es la de este ADR.
