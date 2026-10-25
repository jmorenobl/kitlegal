# Contrato: `docs/SKILLS.md` §1, `CHANGELOG.md` y la frase de `CONTRIBUTING.md`

FR-012, FR-014 y la consecuencia de la fila 1 de contracts/skill-boe-legislacion-v0-3-1.md §3.

## 1. `docs/SKILLS.md` §1 (FR-012, SC-013)

Un párrafo nuevo en «## 1. Qué va en el cuerpo y qué en una referencia», detrás del que empieza «**En una referencia,
el detalle que solo pide una clase de pregunta**» y delante de «Cuándo leer cada referencia lo dice el cuerpo…»:

```markdown
**Lo que se lee de una referencia, medido** (`evidencias/sondeos-20261025/`, `LEEME.md`). Las sesiones abren las
referencias de las que necesitan un dato para seguir y no las que solo traen guía, aunque el cuerpo mande leerlas: en
los sondeos de la eval 15 de `boe-legislacion`, ninguna de las 40 sesiones de `brazo-D` abrió la referencia que el
cuerpo mandaba leer antes de responder, y en los seis brazos se abrieron `normas` (de 23 a 40 sesiones por brazo) y,
en los brazos con la skill de H31 —todos menos `brazo-E`, que no tiene esa referencia—, `con_herramientas` (de 6 a 9).
Por eso lo que toda respuesta necesita va en el cuerpo, y una referencia nueva lleva el paso del protocolo que la pide.
```

Las cifras son las de `evidencias/sondeos-20261025/LEEME.md`, «Los sondeos» (research V8); la revisión final las
coteja con él. Nada más de `docs/SKILLS.md` cambia.

## 2. `CHANGELOG.md`, *Unreleased* (FR-014, SC-006)

Una entrada en «### Cambiado», detrás de la de `boe-legislacion` v0.3, como la v0.2.2 sigue a la v0.2.1 en 0.8.0. La
cifra del cuerpo es la que dé el gate al cerrar la tarea (con el contrato, 10 993):

```markdown
- **`boe-legislacion` v0.3.1: la respuesta habla de la norma, no de la consulta, y lo que nombra lo ha leído antes**
  (H31.1). Sustituye a `boe-legislacion` v0.3. En el sondeo de la eval 15 (`evidencias/sondeos-20261025/`), las
  respuestas que cuentan lo que hizo el agente pasaron de 0 de 40 con la skill de antes de H31 a 17 de 80 con la v0.3.
  El paso «Responder» vuelve a decir que la respuesta es todo lo que se escribe tras la última orden o llamada, desde su
  primera palabra; que quien la pregunta la lee entera y no ve las órdenes, las llamadas ni lo que devuelven; que lo
  que se decide al completar la lectura no va en ella; y que lo hecho, lo que se va a hacer y lo que devolvió una orden
  no le sirven, con sus dos ejemplos, y el paso 5 vuelve a llevar sus cuatro. El paso 3 dice que la respuesta nombra un
  precepto solo si `kitlegal boe articulo` o `articulos` ha devuelto su texto, y manda leer antes, con los pasos 1 y 2,
  el que remite un bloque leído o cuyo id solo salió en un índice; leído, la respuesta dice de qué trata, con su cita, y
  esa lectura es un salto: lo que remite, como lo que no se pudo leer, se nombra solo con las palabras del texto leído.
  El cuerpo queda en 10 993 caracteres: salen a sus referencias textos que solo pide una clase de pregunta —la regla 7,
  la de la comprobación que no termina con `0`, que ya decía `references/lecturas_que_fallan.md` (la 8 pasa a ser la
  7); el detalle de la lectura con fecha, cuyo enlace pasa al paso 2; lo demás de `kitlegal graph check` y la segunda
  lectura de un bloque, a `redaccion_modificada.md`; y lo que se dice de una norma que la búsqueda no da, a
  `buscar_la_norma.md`—. La descripción no cambia. Las evals 14, 15 y 17 llevan en `lecturas` los preceptos que sus
  respuestas nombraban sin haberlos leído —`a60`, `a92`, `a100` y `a104` del texto refundido de la Ley reguladora de
  las Haciendas Locales; `a7`, `a19`, `a20` y `dadecimoctava` de la Ley 35/2006; y `a24` de la Ley 19/2013—, grabados
  en `internal/evals/testdata/`; no cuentan en ningún umbral. `docs/SKILLS.md` §1 dice qué referencias se leen y cuáles
  no, con lo medido.
```

## 3. `CONTRIBUTING.md` (consecuencia de FR-006)

Con la regla 7 de hoy fuera del cuerpo, la 7 de v0.3.1 es la de sin herramienta ni binario, y `CONTRIBUTING.md`
(«`formas_fijas` tampoco es una familia…», línea 743) dejaría de decir verdad. El paréntesis pasa de «(en
`boe-legislacion`, la línea `⚠ REDACCIÓN MODIFICADA:` y la frase de la regla 7 sin su punto final)» a:

```markdown
(en `boe-legislacion`, la línea `⚠ REDACCIÓN MODIFICADA:` y la frase de `references/lecturas_que_fallan.md` sin su
punto final)
```

reflujando el párrafo a 120 columnas. Nada más de `CONTRIBUTING.md` cambia. Los comentarios de
`internal/evals/prohibidas_test.go` que dicen «la frase de la regla 7» citan el contrato de H7.4, en el que lo era, y no
se tocan (research D7).
