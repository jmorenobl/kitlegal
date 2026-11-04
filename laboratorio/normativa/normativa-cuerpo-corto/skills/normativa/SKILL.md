---
name: normativa
description: >-
  Consulta y cita normativa consolidada del Boletín Oficial del Estado (BOE) de cualquier materia —procedimiento
  administrativo, contratación pública, régimen local, tributos, transparencia, relaciones laborales, seguros…—: lee
  con kitlegal índice y artículos y cita identificador y bloque. Se usa siempre que la respuesta dependa de lo que dice
  una norma —qué dice un artículo, ley o real decreto, qué plazo, requisito o procedimiento fija, dónde se regula una
  materia— o de lo que decía en una fecha, la del acto o de los hechos, aunque el agente crea saberlo: sin leer la
  norma con kitlegal no hay cita. Se activa antes de boe_articulo u otra herramienta boe_…: qué pedir y cómo citar lo
  dice la skill. La norma, estatal o autonómica, se nombra por número y año (Ley 39/2015, Real Decreto Legislativo
  2/2004), abreviatura (LPAC, LCSP, LRBRL, LGT, TRLRHL, LRJSP) o identificador BOE-A-….
metadata:
  kitlegal-applets: boe graph
  kitlegal-referencias: normas
  kitlegal-sin-fila: graph_show graph_stats
---

## Protocolo

Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la tienes, con su
herramienta, tras leer [con_herramientas](references/con_herramientas.md).

### 1. Identificar la norma

Toma el `BOE-A-…` de la tabla de [normas](references/normas.md); si no está, búscala con
[buscar_la_norma](references/buscar_la_norma.md).

<!-- paquetes: lo genera make skills-sync -->
<!-- fin de los paquetes -->

### 2. Leer

- El id de cada bloque sale del índice, con fecha o sin ella: `kitlegal boe indice BOE-A-1978-31229 --json`. Copia el
  id de la entrada cuyo `titulo` es el artículo; nunca lo compongas con su número ni lo tomes de la pregunta: no
  siempre es su id.
- Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, lee cada bloque una vez, con
  su comprobación detrás, en la misma orden:
  ```bash
  kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check BOE-A-1978-31229 a140 --json
  ```
  En PowerShell 5.1:
  ```powershell
  kitlegal boe articulo BOE-A-1978-31229 a140 --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check BOE-A-1978-31229 a140 --json }
  ```
  La comprobación, nunca sin argumentos, ni antes, ni con fecha; sin fecha, si faltó, pídela ya. Habla de otras
  conversaciones, que quien pregunta no conoce: un bloque sin entrada `version-obsoleta` no da nada a la respuesta, ni
  una línea, ni una frase junto a los avisos o a la fecha de vigencia. Ni «no consta otra redacción». Las líneas de una
  fecha van siempre.
- Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no termina con `0`, antes de
  seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).

### 3. Responder

- Cada afirmación sobre el contenido de una norma lleva su cita, de lo que `kitlegal boe` devolvió en esta
  conversación.

## Cómo se cita

La forma legible y, en la misma línea, entre corchetes, el identificador y el id como los da `kitlegal boe`, y nada
más: `art. 140 de la Constitución Española [BOE-A-1978-31229, bloque a140]`. Una cita por bloque, también del remitido;
repásalas, y las dudas, en [como_se_cita](references/como_se_cita.md). Cada línea `⚠` va entera y con los datos de la lectura.

## Redacción en una fecha

Con una fecha en la pregunta —un día, un mes, un año o un hecho fechado—, todo texto citado lleva A FECHA; DE HOY, nunca
sola:

```text
⚠ REDACCIÓN A FECHA: <forma legible> [<identificador>, bloque <id>]: se cita la redacción vigente el AAAA-MM-DD, la de fecha de vigencia AAAAMMDD.
⚠ REDACCIÓN DE HOY: <forma legible> [<identificador>, bloque <id>]: hoy la redacción es otra, la de <identificador>, con fecha de vigencia AAAAMMDD.
⚠ SIN REDACCIÓN VIGENTE: <forma legible> [<identificador>, bloque <id>]: el AAAA-MM-DD no tenía redacción vigente; la tiene desde el AAAAMMDD, la fecha de vigencia de su primera redacción.
```

## Redacción modificada

Por cada entrada `version-obsoleta` de `kitlegal graph check`, y si se pregunta qué cambió, lee
[redaccion_modificada](references/redaccion_modificada.md):

```text
⚠ REDACCIÓN MODIFICADA: <forma legible> [<identificador>, bloque <id>]: la redacción con fecha de vigencia AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.
```
