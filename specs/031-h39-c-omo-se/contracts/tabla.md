# Contrato: el guion que cuenta el sondeo

FR-040 a FR-043; SC-006. Research D9 a D11; data-model §5 a §7.

## 1. Cómo se lanza

```sh
make laboratorio-tabla DIRECTORIO=<ruta>     # go run ./cmd/laboratorio tabla <ruta>
```

Sin red ni modelo ni credencial. `<ruta>` es un sondeo si tiene `sondeo.json`; si no, se leen como sondeos los
subdirectorios que lo tienen, en orden de nombre (así se juntan los dos lanzamientos de FR-050). Escribe la tabla y el
detalle en la salida estándar y sale con 0. Sin ningún sondeo, o con una sesión, un `sondeo.json` o un transcript que no
sabe leer, lo dice en la salida de error con la ruta y sale con 1, sin tabla (defecto `inesperado`, ADR 0023).

## 2. Qué cuenta, por sesión

Lo de FR-041, con estas reglas (research D10):

- **Turno**: se numeran desde 1. Empieza uno nuevo con el primer mensaje `assistant` y con cada `assistant` que llega
  después de un `user` con algún `tool_result` y lleva un `message.id` distinto del del `assistant` anterior.
- **Activa**: un `tool_use` de `Skill` con `input.skill` igual a `articulos-del-boe` (o, para la de entrada,
  `consulta-legal`).
- **Primera herramienta, modo herramienta**: el primer `tool_use` cuyo `name`, o lo que sigue a su último `__`, está en
  `app.NombresDeHerramientas(RegistroDeProduccion)`. `ToolSearch` no lo es.
- **Primera herramienta, modo orden**: el primer `tool_use` de `Bash` cuya `input.command`, partida en órdenes simples
  por `|`, `||`, `&&`, `;`, `&` y saltos de línea, y en palabras con las comillas simples y dobles y la barra invertida
  de sh, tiene una orden simple cuya primera palabra que no es `NOMBRE=valor` es `kitlegal` o termina en `/kitlegal`, y
  cuyo argv, para `InterpretarInvocacion`, es de un applet, consulta, y su `<applet>_<verbo>` es una herramienta del
  registro. No lo son `which kitlegal`, `kitlegal --help`, `kitlegal --version`, `kitlegal boe articulo … --describe`,
  `kitlegal skills list`, ni `grep kitlegal …`.
- **Activa antes**: activa en un turno anterior al de la primera herramienta, o activa y sin ninguna herramienta.
- **Abre la referencia** (variantes con `referencia` en `cambia`): un `tool_use` de `Read` cuyo `input.file_path`
  termina en `articulos-del-boe/references/cierre.md`, o de `Bash` cuya `input.command` contiene
  `references/cierre.md`. Su turno es el del primero.
- **Cumple**: `reglaTestigo(Respuesta)` (contracts/laboratorio.md §3).
- **Terminada**: `LeerSesion` da `Terminada` y una `Respuesta` no vacía; si no, la sesión va a «Sin terminar» y no
  entra en ningún `n` (FR-042).

## 3. La tabla

Una fila por variante, modo, modelo y versión de Claude Code, en el orden de `variantes.yaml`, `orden` antes que
`herramienta`, y modelo y versión por orden de texto. Cada recuento, «`<k> de <n> · <p> % [<lo>–<hi> %]`», con un
decimal y coma, y el intervalo de Wilson al 95 % (z = 1,96) recortado a [0, 100] antes de redondear. El turno de la
primera apertura, «`<mediana> (<mínimo>–<máximo>)`», sobre las sesiones que la abren, con la mediana entera o con «,5».
Celdas vacías: las de la entrada fuera de `entrada`; las de la referencia fuera de las tres de referencia; el turno si
ninguna la abre; y con `n` = 0, todas salvo «Sin terminar» y `n`.

```text
# Sondeo de la skill de laboratorio

| Variante | Modo | Modelo | Claude Code | Sin terminar | n | Activan antes | Entrada antes | Las dos antes | Abren la referencia | Turno de la primera apertura | Cumplen la regla |
|---|---|---|---|---|---|---|---|---|---|---|---|
| control | orden | claude-sonnet-5-5 | 2.1.317 | 0 | 40 | 38 de 40 · 95,0 % [83,5–98,6 %] |  |  |  |  | 31 de 40 · 77,5 % [62,5–87,7 %] |
| referencia-paso | herramienta | claude-sonnet-5-5 | 2.1.317 | 0 | 40 | 40 de 40 · 100,0 % [91,2–100,0 %] |  |  | 3 de 40 · 7,5 % [2,6–19,9 %] | 2 (2–4) | 0 de 40 · 0,0 % [0,0–8,8 %] |
| parametro | herramienta | claude-sonnet-5-5 | 2.1.317 | 2 | 0 |  |  |  |  |  |  |
```

## 4. El detalle

Detrás, «## Detalle» y una tabla con una fila por sesión, en el orden de las filas y, dentro, por ruta:

```text
| Sesión | Terminada | Activa antes | Entrada antes | Abre la referencia | Cumple |
|---|---|---|---|---|---|
| h39-a/control/orden/01-constitucion-articulo-140-01 | sí | sí |  |  | sí |
| h39-a/referencia-paso/herramienta/02-lgt-prescripcion-03 | sí | sí |  | turno 2 | no |
| h39-b/parametro/herramienta/04-et-vacaciones-10 | no: <motivo de LeerSesion> |  |  |  |  |
```

La ruta es relativa al `<ruta>` dado; «Abre la referencia», «no» o «turno <t>» en las tres de referencia.

## 5. Uso, de fuera adentro

La lee la persona una vez, para escribir el ADR (FR-051): 14 filas de unos 200 bytes (unos 3 KB) y el detalle de las
560 sesiones de FR-050, a unos 110 bytes cada una (unos 62 KB). No crece con el uso del producto ni con lo que se
acumula: un sondeo nuevo da otra tabla, y deja de leerse cuando el ADR está escrito. No es una señal: no avisa de nada.

## 6. Control

`TestTablaDelLaboratorio` (`internal/evals`, en `make ci`): sobre las sesiones fijas de
`internal/evals/testdata/laboratorio/sondeo/` —los dos modos; `Skill` y la primera herramienta en la misma respuesta
en dos mensajes `assistant`; `Skill` y `ToolSearch` en un turno y `boe_articulo` en el siguiente; una orden `kitlegal`
en una tubería y otra en una secuencia; `which kitlegal`, `--describe` y `skills list` antes; una sesión sin ninguna
herramienta que activa; la referencia abierta dos veces, con `Read` por tramos y con `Bash`; `Grep` sobre ella; una
variante con referencia que ninguna sesión abre; una sesión sin terminar; y una fila sin ninguna sesión terminada— la
salida es, byte a byte, `internal/evals/testdata/laboratorio/tabla.md`. Y sobre sondeos que el test monta en un
temporal con 40 copias de transcripciones fijas, las celdas de 0, 20 y 40 de 40: `[0,0–8,8 %]`, `[35,2–64,8 %]` y
`[91,2–100,0 %]` (FR-043, SC-006).
