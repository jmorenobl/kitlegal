# 0006 · El sobre de salida: forma canónica de `data`, huella con prefijo y espacio reservado

- **Estado**: aceptada
- **Fecha**: 2026-09-11
- **Hito**: H1

## Contexto y problema

`{ok, fuente, url, fecha_consulta, hash, data}` es, desde `CLAUDE.md`, la forma obligatoria de salida de
todo applet: sin `fuente`, `url`, `fecha_consulta` y `hash` no hay cita. Era una frase; H1 es el hito que
la convierte en código, y al hacerlo aparecen cuatro preguntas que la frase no contesta y que, una vez
respondidas, ya no se podrán cambiar sin romper a todo consumidor:

- **Qué se hashea exactamente y cómo.** Dos serializaciones del mismo contenido no tienen por qué producir
  los mismos bytes: el orden de las claves de un `struct` es el de declaración, un entero grande que pase
  por coma flotante pierde precisión y un `<` puede salir escapado o no. Si la huella depende de eso, no
  sirve para lo único para lo que existe: detectar que un contenido cambió.
- **Qué ponen en `fuente` y `url` los applets que no consultan ninguna fuente externa.** `cita`, `plazos`
  y `competencia` calculan; no tienen una URL que citar. Dejarlas vacías rompería la regla de la cita, e
  inventar una URL `https://…` que no resuelve sería peor: un consumidor la trataría como comprobable.
- **Qué forma tiene el sobre cuando algo falla.** Un fallo también tiene que ser legible por máquina, y
  quien falla suele ser el kernel —bandera desconocida, applet no registrado— antes de que exista applet
  alguno que pudiera describirlo.
- **Cuándo se emite.** Un `data` que no se puede serializar no puede dejar medio sobre escrito en la salida
  estándar.

Todo esto se decide en H1 **antes** de que exista el primer adaptador de fuente, en H4: después sería
negociar con hechos consumados.

## Opciones consideradas

**Sobre la huella**:

1. **Implementar JCS (RFC 8785) completo.** Es la norma, y exige una dependencia nueva fuera de la lista de
   la constitución o reimplementar su normalización de números, para conseguir lo mismo que ya consigue la
   biblioteca estándar en este caso. Rechazada por §V.
2. **Serializar y hashear directamente, sin viaje de ida y vuelta.** Es lo más corto y da huellas distintas
   para el mismo contenido según cómo se declarase el tipo. Rechazada.
3. **Hashear el sobre entero** en lugar del contenido de `data`: la huella cambiaría en cada consulta,
   porque `fecha_consulta` cambia. Rechazada: inutiliza el campo.
4. **Publicar solo el hexadecimal, sin nombrar el algoritmo.** Rechazada: obligaría a un cambio
   incompatible el día que el algoritmo cambie.
5. **Ida y vuelta con forma canónica explícita**, con el prefijo del algoritmo delante.

**Sobre la procedencia de lo calculado**: dejar `fuente` y `url` vacías (rechazada: contradice la regla de
la cita); una URL `https://ventanillalegal.es/kitlegal/applet/<nombre>` (rechazada: sería una URL que no
resuelve y que un consumidor tomaría por comprobable); `urn:kitlegal:…` (rechazada: más ceremonia por el
mismo efecto); un **espacio de nombres reservado propio**.

**Sobre el sobre de fallo**: envolver en `{"error": {...}}` (rechazada: `data` ya está condicionado a `ok`,
y un nivel más no desambigua nada); añadir una clave `codigo` (rechazada: duplicaría el código de salida
dentro del JSON y crearía una segunda fuente de verdad que podría discrepar); **dos claves, `clase` y
`mensaje`**.

## Decisión

### Forma

Un único documento JSON con **exactamente** seis claves en el nivel superior, ni una más ni una menos, en
éxito y en fallo. Ninguna se omite cuando vale el valor cero: ningún consumidor debe contar con que falte
una clave ni con que aparezca una séptima.

### La huella

```
hash = "sha256:" + hex( sha256( canónico(data) ) )
```

`canónico` son tres pasos:

1. Serializar `data` a JSON.
2. Volver a leerlo a una representación genérica **conservando los números como literales**, sin pasar por
   coma flotante.
3. Volver a serializarlo **con las claves de todo objeto ordenadas**, sin espacios ni saltos de línea y sin
   escapar caracteres HTML.

El paso 2 convierte todo objeto en un mapa, y el paso 3 emite las claves de un mapa ordenadas: de ahí sale
la propiedad que se busca —el mismo contenido produce la misma huella con independencia del orden en que
se hubiera declarado— sin ninguna dependencia nueva. La huella **no** depende de `fecha_consulta`, y por
eso sirve para detectar que un contenido cambió; un contenido que difiere en un byte produce otra
distinta.

**El prefijo del algoritmo es obligatorio.** Va delante del hexadecimal para que se pueda cambiar de
algoritmo sin romper a quien lee, y quien lee debe comprobar el prefijo antes de interpretar el resto. Una
huella sin prefijo no es una huella de este proyecto.

### El espacio de nombres reservado `kitlegal.` / `kitlegal:`

| Emisor | `fuente` | `url` |
|---|---|---|
| Applet calculado (`echo` en H1; previsiblemente `cita` y `plazos`) | `kitlegal.<nombre>` | `kitlegal:applet/<nombre>` |
| El propio kernel, cuando falla antes de llegar al applet | `kitlegal.cli` | `kitlegal:cli` |
| Adaptador de fuente pública (H4 en adelante) | el nombre de la fuente, p. ej. `boe.legislacion-consolidada` | URL `http(s)` comprobable |

`kitlegal:applet/echo` es un URI absoluto —esquema `kitlegal`, parte opaca `applet/echo`—, así que pasa la
validación formal de `url` que el sobre exige, mientras que la cadena vacía no. La regla para el consumidor
es de una línea: **un sobre cuya `url` empieza por `kitlegal:` es un resultado calculado, no una cita de
fuente pública**, y no debe presentarse como fuente ni usarse para fundamentar una afirmación legal.

**Prohibición que nace en H4.** Ningún adaptador de `internal/source/<fuente>` puede usar el prefijo
`kitlegal.` ni el esquema `kitlegal:`: si lo hiciera, un resultado obtenido de una fuente externa quedaría
marcado como calculado —o, peor, un resultado calculado se presentaría como cita comprobable—, y la regla
anterior dejaría de ser cierta justo donde importa. En H1 **no hay nada que comprobar**, porque
`internal/source/` no existe todavía. **La comprobación mecánica se escribe con el primer adaptador, en
H4**, y queda anotada aquí para que no se pierda entre hitos: es la forma que toma en este proyecto la
regla de la constitución de que toda fuente externa se declara antes de escribir su adaptador.

### El sobre de fallo

Lo emite **el kernel**, no el applet, desde el único punto que traduce el error a código de salida; por eso
es idéntico para todos los applets y también para los fallos anteriores a la ejecución de cualquiera.

```json
{
  "ok": false,
  "fuente": "kitlegal.cli",
  "url": "kitlegal:cli",
  "fecha_consulta": "2026-09-11T10:12:00+02:00",
  "hash": "sha256:…",
  "data": { "clase": "argumentos", "mensaje": "bandera desconocida: --jsno" }
}
```

- `data` es **exactamente** `{clase, mensaje}`: sin claves adicionales, sin envoltorio y sin traza. El
  detalle técnico va al registro de eventos de la salida de error, donde `--verbose` lo hace visible.
- `clase` es una de seis —`argumentos`, `no-encontrado`, `fuente-no-disponible`, `limite-o-tos`,
  `identidad-humana`, `inesperado`— y **corresponde al código de salida emitido**. La correspondencia se
  comprueba entre la clase y el código del proceso, no dentro del JSON, que es la razón de que no haya una
  clave `codigo`.
- `ok` es falso **si y solo si** el código de salida no es 0. La invariante se escribe así, y no como un
  booleano que se pasa a mano: el sobre se monta a partir del código, de modo que no hay forma de emitir
  uno que diga lo contrario de lo que el proceso dirá al terminar.
- `fuente` y `url` son las de la fuente que se estaba consultando cuando se conocen —una URL que devolvió
  «no encontrado» es una cita negativa útil— y las del espacio reservado del kernel cuando no.
- `fecha_consulta` y `hash` se calculan igual que en un sobre de éxito.
- El sobre de fallo **no sustituye al código de salida**: lo duplica en forma estructurada. El código sigue
  siendo la vía primaria de clasificación y el mensaje sigue yendo además a la salida de error.
- Sin la forma legible por máquina no hay sobre de fallo: el mensaje va a la salida de error y la estándar
  queda vacía. No hay una tercera forma de presentar un fallo.

### Cuándo se emite

Un `data` no serializable hace fallar el montaje **antes** de que nada llegue a la salida estándar: el
fallo se clasifica como inesperado y sale con su propio sobre, porque el descriptor sigue sano. Si lo que
falla es la escritura del sobre ya montado —una tubería cerrada—, el error se propaga y termina también
como inesperado, y **no se intenta un segundo sobre por el descriptor roto**. Nunca se emite un sobre a
medias.

## Consecuencias

**A favor**

- Un consumidor —una skill, el servidor MCP, otro agente— puede escribir una sola rutina de lectura para
  todo el binario, presente y futuro, y distinguir éxito de fallo sin interpretar texto.
- La huella es comparable entre ejecuciones y entre máquinas, que es lo que permite detectar que una norma
  consolidada cambió sin volver a leerla entera.
- El espacio reservado hace visible, en el propio dato, la diferencia entre lo que se cita y lo que se
  calcula: la regla de «cada afirmación con su cita» deja de depender de que alguien la recuerde.
- El sobre de fallo hace que un error sea tan citable como un acierto, incluso cuando el fallo es anterior
  a que exista applet.
- Fijar todo esto sin adaptadores evita el sesgo de diseñar el contrato alrededor del primero que llegue.

**En contra, y asumido**

- La huella pasa por dos serializaciones en lugar de una: es trabajo que se paga en cada sobre, y se acepta
  a cambio de que la huella signifique algo.
- No es JCS. Cumple la propiedad que se busca, pero no es intercambiable con implementaciones de la norma,
  y decirlo aquí evita que alguien lo suponga.
- El espacio de nombres reservado es una convención del proyecto: fuera de él, `kitlegal:applet/echo` no
  significa nada para nadie.
- **La prohibición del espacio reservado queda escrita y sin control mecánico hasta H4.** Es deuda
  reconocida, no un olvido: mientras no exista `internal/source/`, un control no tendría nada que vigilar.
- Seis claves exactas significa que añadir una séptima es un cambio incompatible, aunque parezca aditivo:
  el contrato dice explícitamente que no habrá más.
- `{clase, mensaje}` deja fuera el detalle técnico del sobre; quien lo necesite tiene que mirar la salida
  de error, que es donde se decidió que viviera.
