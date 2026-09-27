# 0005 · Contrato del applet: qué declara, qué hereda y dónde vive cada pieza

- **Estado**: aceptada
- **Fecha**: 2026-09-11
- **Hito**: H1

> Nota (2026-09-13): el ADR 0014 añade a `schema.Resultado` las operaciones de grafo que el applet
> observa, aplicadas por el kernel con la `Procedencia` del propio resultado como `source`. `Applet`,
> `Verbo` y `Argumentos` no cambian; lo que crece es el valor de dominio que el applet ya devolvía.
>
> Nota (2026-09-27): el ADR 0026 añade a `schema.Resultado` el campo `Legible`, el contenido contado para
> una persona, que el kernel escribe sin `--json` en lugar de la tabla mínima. Por el mismo camino y con la
> misma regla: el contrato no cambia, y `render` sigue sin conocer applets.

## Contexto y problema

`docs/ROADMAP.md` prevé más de veinte applets: uno por fuente legal (`boe`, `borme`, `placsp`, `bdns`,
`eurlex`…) y uno por capacidad transversal (`cita`, `plazos`, `competencia`, `graph`, `mcp`). El
[ADR 0001](0001-multicall.md) decidió que todos viven en un único ejecutable y que heredan las mismas
convenciones de agente; lo que no decidió es **qué tiene que escribir quien añade el vigésimo sexto**.

Si eso no queda fijado ahora, cada applet acabará repitiendo —y variando— el sobre de salida, la
traducción de errores a código de salida, la forma de presentación y la lectura de las banderas globales.
Un agente que razona sobre la salida del binario dejaría de poder hacerlo en cuanto dos applets
discrepasen en un detalle.

Hay además una restricción estructural. `docs/ROADMAP.md` §2 sitúa el registro de applets en
`internal/app` y las banderas, el mapeo error→código y el sobre en `internal/cli`, y la constitución §IV
exige un dominio puro sin entrada ni salida. Respetar las tres cosas a la vez obliga a decidir **en qué
dirección apunta cada dependencia**: si el kernel conociera el tipo `Applet` y el registro conociera el
kernel, habría un ciclo, y la única salida sería mover una de las dos piezas fuera del sitio que el
roadmap le dio.

## Opciones consideradas

1. **Que el applet reciba el contexto del analizador de la línea de órdenes** (`*kong.Context`) y se sirva
   de él. Es lo más directo mientras hay un applet, y filtra la biblioteca de análisis a los veinticinco
   siguientes y a cada `internal/source/<fuente>`, contra §IV. Rechazada.
2. **Que el applet monte su sobre y lo escriba**. Da libertad a cada applet, que es exactamente el
   problema: la forma del sobre pasaría a depender de quién lo emitió, y ningún control podría impedir que
   el applet número doce añadiera una séptima clave. Rechazada.
3. **Contrato declarativo mínimo**: el applet declara nombre, verbos, gramática y el tipo de su contenido;
   el kernel se ocupa de todo lo demás y el applet no puede emitir otra forma porque no tiene por dónde.
4. **Declarar `Applet` y `Registro` en `internal/cli`**, o añadir un paquete `internal/kernel` que
   orqueste, para evitar el ciclo por la vía de mover piezas. Rechazadas: la primera contradice la
   asignación literal del roadmap sin ganar nada; la segunda añade un paquete que ningún requisito pide.

## Decisión

Se adopta la **opción 3**, con la dirección de dependencia que la hace posible.

### El contrato

```go
// internal/app
type Applet interface {
    Nombre() string        // con lo que se le invoca: primer argumento o enlace simbólico
    Descripcion() string   // la línea que aparece en la ayuda del binario
    Verbos() []Verbo       // al menos uno
}

type Verbo struct {
    Nombre      string
    Descripcion string
    Argumentos  func() Argumentos // fábrica: un valor nuevo por invocación, nunca compartido
    Salida      any               // valor cero del tipo de `data`, solo para reflejarlo en --describe
    PorOmision  bool              // como máximo uno por applet
}

type Argumentos interface {
    Ejecutar(ctx context.Context, ec schema.Contexto, log *slog.Logger) (schema.Resultado, error)
}

// internal/core/schema
type Resultado struct {
    Procedencia Procedencia // {Fuente, URL}
    Datos       any
}
```

Eso es **todo** lo que declara un applet. No declara banderas, ni códigos de salida, ni forma de
presentación, ni sobre: los hereda. Y no puede emitirlos por su cuenta, porque devuelve un `Resultado`
—procedencia y contenido— y nunca un sobre ya montado, ni texto escrito en un descriptor, ni un código.

Tres detalles del contrato que no son accidentales:

- **`Argumentos` es una fábrica y no un valor.** Cada invocación necesita un `struct` vacío sobre el que
  el analizador escriba; compartir la instancia sería estado global y rompería cualquier test en paralelo.
- **`Salida` se refleja, no se escribe.** Es el valor cero del tipo del contenido de `data`, y `--describe`
  deriva de él el esquema. Pedirle al applet un esquema escrito a mano garantizaría que algún día
  describiera algo distinto de lo que emite.
- **El registrador de eventos viaja como parámetro explícito**, no dentro del contexto de ejecución: ese
  contexto es dominio puro y no importa `log/slog`. Llega montado y con el nivel ya resuelto, de modo que
  el applet no interpreta `--verbose` ni ninguna otra bandera.

### El verbo por omisión

`PorOmision` es lo único que el contrato añade sobre la lectura literal de los requisitos, y lo añade
porque la entrega del hito es literal: `kitlegal echo hola` no nombra verbo alguno, mientras que
`kitlegal boe articulo BOE-A-2015-10565 a21` sí. Ambas conviven así: resuelto el applet, si el primer
argumento anterior al terminador `--` nombra un verbo suyo, es el verbo; en otro caso el kernel inserta el
verbo por omisión, si el applet declara uno; y si no lo declara, no inserta nada y el verbo pasa a ser
obligatorio.

La ambigüedad queda resuelta por escrito: `kitlegal echo repetir` invoca el verbo `repetir`, no el verbo
por omisión con el mensaje `repetir`. Quien quiera lo segundo escribe `kitlegal echo repetir repetir`.

**No se asume implícitamente** que un applet con un solo verbo lo tome por omisión: esa regla cambiaría el
comportamiento sola, sin que nadie tocara nada, el día que el applet declarase el segundo. Las invariantes
—`Verbos()` no vacío, nombres únicos y como mucho un verbo por omisión— se comprueban **al construir el
registro**, no en la invocación: un defecto del catálogo es un defecto de programación y no un error de
quien invoca.

### El reparto, y la dirección de dependencia que lo sostiene

```
cmd/kitlegal → internal/app → internal/cli → internal/core/schema
                                internal/render → internal/core/schema
```

| Capa | Paquete | Qué le corresponde |
|---|---|---|
| Composición | `cmd/kitlegal` | Solo la raíz: leer los argumentos, montar el registro de producción, inyectar los descriptores y propagar el código de salida |
| Composición | `internal/app` | El contrato anterior, el registro y sus invariantes, el despacho multicall, la normalización del verbo por omisión y la función que lo monta todo y devuelve un `int` |
| Kernel | `internal/cli` | Las ocho banderas globales, el pre-escaneo acotado de los argumentos, la gramática, los errores tipados y su traducción a código de salida, el montaje del sobre —de éxito y de fallo— y `--describe` |
| Dominio | `internal/core/schema` | El sobre, la procedencia, el resultado, el contexto de ejecución, las clases de error y la huella. Sin entrada ni salida |
| Presentación | `internal/render` | Las dos formas de presentación y el **único** escritor de la salida estándar |

Dos inversiones deliberadas hacen que esa dirección no tenga ciclos ni obligue a mover nada de sitio:

- **El kernel no conoce el tipo `Applet`.** La función de `internal/cli` que analiza la gramática recibe el
  `struct` como `any` y devuelve lo analizado, sin saber de dónde salió. Por eso `app` puede importar `cli`
  sin que `cli` importe `app`.
- **El kernel no importa `internal/render`.** Necesita escribir, así que declara la interfaz mínima que
  consume —`Presentador`— y la raíz de composición le inyecta la implementación. Es «accept interfaces,
  return structs», y es lo que permite escribir sin invertir la flecha.

Las cinco reglas de dependencia que esto implica no se quedan en este documento: se comprueban en cada
`make ci`, por dos capas independientes en el caso de las reglas de importación —configuración del lint y
un test que recorre el grafo transitivo real—, y el fallo nombra la regla violada.

## Consecuencias

**A favor**

- Añadir un applet es escribir un `struct` con etiquetas y un método. Todo lo que un agente necesita para
  razonar sobre la salida —sobre, huella, códigos de salida, banderas, autodescripción— viene dado.
- Ningún applet puede emitir una forma de salida distinta de la de los demás, porque no tiene por dónde
  hacerlo: no escribe, no monta el sobre y no elige el código de salida.
- El dominio se mantiene puro por construcción y no por disciplina: lo que el applet recibe del kernel ya
  está interpretado, y lo que devuelve es dominio.
- La **API** de la biblioteca de análisis de la línea de órdenes queda confinada a `internal/cli`: ningún
  applet importa Kong ni recibe su contexto. Lo que sí forma parte del contrato del applet es el
  **vocabulario de etiquetas** con el que declara sus argumentos —`arg`, `optional`, `required`,
  `default`, `help`, `name`—, que es el de Kong: sustituir la biblioteca exigiría o bien que la nueva
  leyera esas mismas etiquetas, o bien migrar las etiquetas de cada applet. Es un coste acotado a un
  `struct` por verbo y sin lógica, pero no es cero, y decirlo evita suponer lo contrario.
- `--describe` no puede mentir sobre la forma **declarada** de `data`: la deriva del tipo que el verbo
  pone en `Salida`, y no de una descripción escrita a mano que pudiera envejecer aparte. Lo que sí puede
  divergir es la declaración de lo que `Ejecutar` devuelve —ver el último punto de la lista siguiente—, y
  eso lo vigila un test por applet, no el contrato.

**En contra, y asumido**

- `Salida` y `Resultado.Datos` son dos valores distintos que nada relaciona en tiempo de compilación: un
  applet que declarase un tipo y devolviera otro publicaría con `--describe` un esquema que rechaza su
  propio sobre, sin que el compilador ni el kernel dijeran nada. Es el precio de que `Salida` sea un valor
  reflejado y no un parámetro de tipo del contrato. Se paga con el control de la Definition of Done §1.4:
  cada applet valida en test su salida real contra el esquema que él mismo emite, con las aserciones de
  formato activadas (`TestSalidaContraSuEsquema` para los dos de ejemplo, en `internal/app`).

- La firma de análisis del kernel recibe un `any` y pierde la comprobación de tipos en ese punto: es el
  precio de no mover el registro fuera de donde el roadmap lo puso, y se paga con tests.
- El contrato es difícil de cambiar: `Applet`, `Verbo` y `Argumentos` los implementará todo applet
  posterior, así que ampliarlos es un cambio incompatible y exige un ADR que sustituya a este.
- `PorOmision` es una capacidad que la mayoría de los applets no usará, y que quien lea el contrato tiene
  que entender igualmente.
- El registro valida sus invariantes al construirse, de modo que un catálogo mal declarado se manifiesta
  como un fallo al arrancar el binario y no como un error de quien invoca.
- Una interfaz `Presentador` declarada por el consumidor significa que añadir una forma de presentación
  obliga a tocar esa declaración, no solo `internal/render`.
