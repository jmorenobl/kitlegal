# 0063 · El cuerpo de una skill de capacidad deja sitio para las materias: la tabla de comandos, sin lo que da la llamada, y el detalle de un caso, en la referencia que se lee en él

- **Estado**: propuesta (2026-11-24); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-24
- **Hito**: H46. Se escribe con su entrada, antes del run y fuera de él, en una sesión interactiva. Sustituye el punto 3
  del ADR 0062 y la frase de sus «Consecuencias» que lo preveía («cuando el siguiente no quepa, un hito de la base
  cambia la forma de la región»). Precisa el ADR 0046 (punto 1) en lo que lleva la tabla de comandos del cuerpo, y
  sustituye la línea del sobre de la tabla de H21 (FR-030). No cambia el tope de 11 000 caracteres del cuerpo (ADR
  0046, punto 2), la forma de la línea de un paquete ni la regla de que ninguna referencia enlaza otra (§4 de
  `docs/SKILLS.md`). No cambia lo que devuelve ningún applet.

## Contexto y problema

El ADR 0062 (punto 3) dejó la línea por paquete en el cuerpo de cada skill de capacidad, con unos 70 caracteres por
paquete, y previó que «cuando el siguiente no quepa, un hito de la base cambia la forma». Ha pasado con el primero. El
prototipo de H42 sobre `main` (`7aaad57f`), sin sus tests parcheados, deja `make ci` en rojo en un solo punto: el
guion `h45-paquete-de-mas` corre `TestVerticalDeEnsayo` con un paquete de más, y el cuerpo de `normativa` llega a
11 030 caracteres. Con H42, `normativa` mide 10 821 y `plazos-y-territorio` 10 757; cada línea de paquete cuesta
52 + 3n caracteres, con n la longitud del nombre (`internal/skills/paquetes.go`, `lineas`): 70 para `fiscal`, 94 para
`administrativo`. Con el cuerpo como está, cabe uno más y no dos.

La forma de la línea se queda. Es el patrón «Domain-specific organization» de la guía de Anthropic («Skill authoring
best practices»), un enlace directo con la condición de cuándo leerlo, y la especificación pide las referencias a un
nivel de `SKILL.md` («Keep file references one level deep from SKILL.md», agentskills.io, «Specification», §«File
references»; la guía, «Avoid deeply nested references»), consultadas el 2026-11-24. Una referencia índice con la
lista de paquetes añadiría un nivel, y H40 midió que la referencia nombrada dentro de otra orden se abre en 3 a 13 de
40 sesiones (ADR 0051, punto 1). Lo que se cambia es lo demás del cuerpo: progressive disclosure, en el cuerpo solo lo
que toda respuesta necesita (ADR 0046, punto 1).

Medido sobre `main` el 2026-11-24, en runas, como el gate (`docs/SKILLS.md` §3):

| | `normativa` | `plazos-y-territorio` | `jurisprudencia` |
|---|---|---|---|
| Cuerpo | 10 750 | 10 686 | 10 842 |
| Líneas de paquete (sin las marcas) | 0 | 94 | 0 |
| Tabla de comandos generada | 2673 | 1498 | 1302 |
| … de ella, la columna «Qué devuelve en `data`» | 883 | 205 | 171 |
| … de ella, la línea del sobre y la de las banderas comunes | 318 | 318 | 318 |
| Bloques `text` (las formas fijas) | 887 | 1142 | 152 |
| Reglas | 1298 | 2168 | 1480 |
| Protocolo y demás prosa | 5817 | 5784 | 7833 |

Tres cosas del cuerpo no las necesita ninguna respuesta antes de su primera llamada, o las necesita solo una clase de
pregunta que ya lee una referencia:

- **Lo que devuelve cada orden** (la columna «Qué devuelve en `data`»), **el sobre y las banderas comunes**. El agente
  los ve en el resultado de cada llamada; con herramientas, cada una anuncia además su esquema de salida
  (`OutputSchema`, `internal/mcp/servir.go`), y con la orden lo da `--describe`. Lo que el protocolo usa de `data` lo
  nombra el protocolo (`version-obsoleta`, `data.preceptos`, `data.mensaje`, `es_la_pedida`, `data.clase`), y la única
  bandera común que usa es `--json`, que va en cada orden de ejemplo. Son 1201 caracteres en `normativa`.
- **En `plazos-y-territorio`, cómo se pide cada operación y la forma del territorio**, que el cuerpo da en «Comandos» y
  en «Cómo se presenta el territorio». `municipios.md` se manda leer sin condición antes de la primera orden o llamada
  (H41.1; ADR 0053) y ya lleva la plantilla del territorio con las mismas tres líneas de la cobertura.
- **En `jurisprudencia`, qué se pide según cómo se dio la referencia**, que repite en forma de tabla lo que
  `consulta_preparada.md` da por orden, y que se manda leer sin condición antes de `cita_preparar`; y dos reglas que
  solo valen con un documento delante —la referencia en la forma pedida y las sentencias que el documento menciona—,
  cuando `documento_aportado.md` se manda leer antes de cada cotejo.

## Opciones consideradas

1. **Subir el tope de 11 000.** Rechazada: es el de la especificación, «< 5000 tokens recommended», con la razón medida
   del ADR 0046.
2. **Una referencia con la lista de paquetes, o una línea con todos.** Rechazada: añade un nivel de referencias, que
   la especificación y la guía de Anthropic no quieren, o una lectura más en cada pregunta, en las órdenes que H41 y
   H41.1 midieron.
3. **Recortar 100 caracteres dentro de H42, o relajar `h45-paquete-de-mas`.** Rechazada: aplaza el problema al paquete
   siguiente, y el guion prueba lo que promete el molde.
4. **Llevar la tabla de comandos entera a una referencia.** Rechazada: la orden y el nombre de la herramienta los
   necesita toda respuesta antes de su primera llamada, y en `jurisprudencia` ninguna referencia se lee siempre.
5. **Quitar de la tabla lo que da la llamada, y llevar el detalle de una clase de pregunta a la referencia que esa
   clase ya lee sin condición, con un sitio reservado y comprobado para las materias.** Elegida.

## Decisión

1. **El sitio de las materias se reserva y se comprueba.** El cuerpo de cada `SKILL.md` con las marcas de los paquetes,
   sin las líneas que la sincronía escribe entre ellas, tiene como mucho **9500 caracteres**, contados como el
   presupuesto del cuerpo (runas, con la tabla de comandos generada y las marcas). Quedan 1500 para las líneas de los
   paquetes: unas veinte materias de nombre corto, a 52 + 3n cada una. Es una regla de `skills-check`, con su defecto
   (`normativa: SKILL.md: cuerpo sin las líneas de los paquetes de 9501 caracteres (máximo 9500)`), su mutante en rojo y
   su límite en verde. El tope de 11 000 sigue siendo el del cuerpo entero, con los paquetes. `TestVerticalDeEnsayo`
   deja de comprobar el sitio para «otra línea de su misma longitud»: lo comprueba esta regla para todos los paquetes.
   Una regla nueva del protocolo cabe en ese presupuesto o saca a una referencia lo que solo pide una clase de
   pregunta (`docs/SKILLS.md` §3); no toma el sitio de las materias.
2. **La tabla de comandos lleva la orden, la herramienta y lo que hace**, una fila por verbo, y nada más: sin la
   columna «Qué devuelve en `data`», sin la línea del sobre y sin la de las banderas comunes. La genera `make
   skills-sync` desde `--describe`, como hoy; `kitlegal-sin-fila` no cambia. `--describe` y el esquema de salida de
   cada herramienta siguen declarando lo mismo, y `LeerDescripcionDeVerbo` sigue leyéndolo y dando sus defectos.
3. **El detalle de una clase de pregunta va a la referencia que esa clase lee sin condición**, con la regla de §1 de
   `docs/SKILLS.md`: lo movido va entero al final de su referencia, y la orden de leerla dice que lo trae. Lo que se
   mueve, y la eval que llega a su caso (ADR 0052, punto 4):
   - `plazos-y-territorio`: el párrafo de «Comandos» —la herramienta o la orden, y los códigos— a `municipios.md`, que
     se lee antes de la primera orden o llamada (todas las evals que llaman a `kitlegal`); y «Cómo se presenta el
     territorio» a la misma, para las evals 01 y 02. En «Comandos» queda que `kitlegal` se invoca desde el `PATH`.
   - `jurisprudencia`: la tabla «Cómo se dio / Qué se pide» y la viñeta del número de resolución, a
     `consulta_preparada.md`, que se lee antes de pedir `cita_preparar` (evals 01, 02 y 07); en el cuerpo queda que un
     número sin fecha no se prepara, porque ninguna eval llega a ese caso. La referencia en la forma pedida (eval 06) y
     las sentencias que menciona el documento (evals 04 y 08), a `documento_aportado.md`, que se lee antes de cada
     cotejo, con la salida 0 de un documento que no es el pedido.
   - `normativa`: nada se mueve. Su frase «o, si la tienes, con su herramienta» pierde la condición, que
     `con_herramientas.md` ya da, como pide el ADR 0051 (punto 2).

   No sale del cuerpo ninguna línea `⚠`, ninguna forma fija, ninguna regla invariante ni ninguna orden de leer, y los
   controles de contenido (`docs/SKILLS.md` §1) no cambian.
4. **Lo deciden las evals, con una medida previa.** Antes de fusionar la entrada, un despacho del job de evals sobre un
   prototipo con estos cuerpos, con un criterio fijado antes de mirar: ninguna clase que decide por debajo de su
   umbral frente al cierre de H45.1 (`specs/043-h45-1-los-tests-del/gates/informe-final.md`, sección 3), una muestra
   de respuestas leída, y decide solo `claude-sonnet-5-5` (ADR 0031). Como mucho un segundo despacho por pregunta.

## Consecuencias

**A favor**

- Caben unas veinte materias en cada skill de capacidad sin tocar su protocolo, y H42 entra sin recortar nada.
- El sitio de las materias no se lo come el protocolo: lo comprueba `make ci`.
- El cuerpo pierde lo que el agente ya ve en cada llamada: en `normativa`, unos 530 tokens.

**En contra, y asumido**

- Lo que va a una referencia el modelo puede no leerlo (ADR 0046). Por eso solo se mueve a las tres referencias que se
  mandan leer sin condición antes de la llamada que las necesita, la forma que H40 y H41.1 midieron en 38 a 40 de 40
  sesiones (ADR 0051 y 0053), y cada movimiento tiene una eval que llega a su caso.
- Sin la columna, el agente no sabe antes de la primera llamada qué claves trae `data`. Lo que el protocolo necesita lo
  nombra el protocolo; lo demás lo ve en el resultado.
- El protocolo de cada skill de capacidad tiene 9500 caracteres, no 11 000. Una regla nueva que no quepa obliga a sacar
  detalle a una referencia, con una eval que llegue a su caso.
- El hito cambia `skills/` e `internal/skills`, así que su cierre mide todas las skills sin reutilizar nada de `main`,
  y la serie de cierres de cada skill empieza de cero (ADR 0038).
- Veinte es una cifra de nombres cortos. Con nombres largos caben menos, y el tope de 11 000 sigue mandando.
