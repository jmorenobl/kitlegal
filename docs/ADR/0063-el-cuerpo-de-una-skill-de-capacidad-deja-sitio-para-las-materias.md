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
paquete de H42 sin ningún `.go` (`git diff 5153857c^ abfd73ca -- . ':!*.go' ':!.github'`), aplicado sobre la
cabeza de H45.1 (`4eb02776`, fusionada como `7aaad57f`), deja `make ci` en rojo en un solo punto: el guion
`h45-paquete-de-mas` corre `TestVerticalDeEnsayo` con un paquete de más, y el cuerpo de `normativa` llega a
11 030 caracteres. Con H42, `normativa` mide 10 820 y `plazos-y-territorio` 10 756; cada línea de paquete cuesta
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
| … de ella, la columna «Qué devuelve en `data`», la línea del sobre y la de las banderas comunes | 1237 | 569 | 506 |
| Bloques `text` (las formas fijas) | 887 | 1142 | 152 |
| Reglas | 1298 | 2168 | 1480 |

Tres cosas del cuerpo no las necesita ninguna respuesta antes de su primera llamada, o las necesita solo una clase de
pregunta que ya lee una referencia:

- **Lo que devuelve cada orden** (la columna «Qué devuelve en `data`»), **el sobre y las banderas comunes**. El agente
  los ve en el resultado de cada llamada; con la orden lo da además `--describe`, y cada herramienta anuncia su esquema
  de salida (`OutputSchema`, `internal/mcp/servir.go`), que el agente ve si su host lo pasa. Lo que el protocolo usa de
  `data` lo nombra el protocolo (`version-obsoleta`, `data.preceptos`, `data.mensaje`, `es_la_pedida`, `data.clase`),
  y las banderas que usa van escritas en sus órdenes (`--json` en todas, `--offline` en `lecturas_que_fallan.md`). Son
  1237 caracteres en `normativa`.
- **En `plazos-y-territorio`, cómo se pide cada operación**, que el cuerpo da en «Comandos», y cómo se escribe cada
  dato del territorio. `municipios.md` se manda leer sin condición antes de la primera orden o llamada (H41.1; ADR
  0053) y ya lleva la plantilla del territorio. Las tres líneas de la cobertura, que la eval 02 comprueba sin modelo,
  son una forma fija y se quedan.
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

1. **El sitio de las materias se reserva y se comprueba.** El cuerpo de cada `SKILL.md` que lleva las marcas de los
   paquetes —hoy, las tres skills de capacidad—, sin las líneas que la sincronía escribe entre ellas, tiene como mucho
   **9600 caracteres**, contados como el presupuesto del cuerpo (runas, con la tabla de comandos generada y las
   marcas). Quedan 1400 para las líneas de los paquetes: veinte materias de nombre como `fiscal`, a 52 + 3n cada una, y
   menos con nombres largos. Es una regla de `skills-check`, con su defecto
   (`normativa: SKILL.md: cuerpo sin las líneas de los paquetes de 9601 caracteres (máximo 9600)`), su mutante en rojo y
   su límite en verde en `TestSkillsDelRepositorio`. Un `SKILL.md` sin esas marcas no la lleva. Sustituye la
   comprobación de `TestVerticalDeEnsayo` de que cabe «otra línea de paquete de su misma longitud», que se quita.
   El tope de 11 000 sigue siendo el del cuerpo entero, con los paquetes. Una regla nueva del protocolo cabe en ese
   presupuesto o saca a una
   referencia lo que solo pide una clase de pregunta (`docs/SKILLS.md` §3); no toma el sitio de las materias.
2. **La tabla de comandos lleva la orden, la herramienta y lo que hace**, una fila por verbo, y nada más: sin la
   columna «Qué devuelve en `data`», sin la línea del sobre y sin la de las banderas comunes. La genera `make
   skills-sync` desde `--describe`, como hoy; `kitlegal-sin-fila` no cambia. `--describe` y el esquema de salida de
   cada herramienta siguen declarando lo mismo, y `LeerDescripcionDeVerbo` sigue leyéndolo y dando sus defectos.
3. **El detalle de una clase de pregunta va a la referencia que esa clase lee sin condición**, con la regla de §1 de
   `docs/SKILLS.md`: lo movido va entero a su referencia, y la orden de leerla dice que lo trae. Va al final, salvo
   cuando encaja en un apartado que ya tiene, y con lo mínimo cambiado para leerse fuera del cuerpo: lo que nombraba
   el cuerpo («esta forma», «de su fila») nombra `SKILL.md`. Lo que se mueve, y la eval que llega a su caso (ADR 0052,
   punto 4):
   - `plazos-y-territorio`: el párrafo de «Comandos» —la herramienta o la orden, y los códigos— a `municipios.md`, que
     se lee antes de la primera orden o llamada (todas las evals que llaman a `kitlegal`); y, de «Cómo se presenta el
     territorio», cómo se escribe cada dato y que la cobertura se puede explicar detrás, a la misma, para las evals 01
     y 02. En el cuerpo quedan las tres líneas de la cobertura con su forma, y en «Comandos», que `kitlegal` se invoca
     desde el `PATH`.
   - `jurisprudencia`: la tabla «Cómo se dio / Qué se pide» y la viñeta del número de resolución, a
     `consulta_preparada.md`, que se lee antes de pedir `cita_preparar` (evals 01, 02 y 07); en el cuerpo queda que un
     número sin fecha no se prepara, porque ninguna eval llega a ese caso. La referencia en la forma pedida (eval 06) y
     las sentencias que menciona el documento (evals 04 y 08), a `documento_aportado.md`, que se lee antes de cada
     cotejo, con la salida 0 de un documento que no es el pedido.
   - `normativa`: nada se mueve. Su frase «o, si la tienes, con su herramienta» pierde la condición, que
     `con_herramientas.md` ya da, como pide el ADR 0051 (punto 2).

   No sale del cuerpo ninguna línea `⚠`, ninguna forma fija —las líneas `⚠` y los bloques `text` que una eval
   comprueba sin modelo—, ninguna regla invariante ni ninguna orden de leer, y los controles de contenido
   (`docs/SKILLS.md` §1) no cambian.
4. **Lo deciden las evals, con una medida previa.** Antes de fusionar la entrada, un despacho del job de evals sobre un
   prototipo con estos cuerpos, con un criterio fijado antes de mirar: ninguna clase que decide por debajo de su
   umbral frente al cierre de H45.1 (`specs/043-h45-1-los-tests-del/gates/informe-final.md`, sección 3), una muestra
   de respuestas leída, y decide solo `claude-sonnet-5-5` (ADR 0031). Como mucho un segundo despacho por pregunta. Un
   retoque de los textos posterior al despacho que no cambia lo que se mide no pide otro: lo mide el cierre del hito.

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
- El protocolo de cada skill de capacidad tiene 9600 caracteres, no 11 000, y nace con poco margen: al cerrar el
  prototipo, 102 en `normativa`, 117 en `plazos-y-territorio` y 127 en `jurisprudencia`. La primera regla nueva que no
  quepa, en una revisión final o en un hito de capacidad (H34, H35, H44), obliga a sacar detalle a una referencia, con
  una eval que llegue a su caso.
- `consulta-juridica` crece más por paquete que las de capacidad: su región de lo cubierto y la de los paquetes de la
  base suman 317 caracteres con el paquete fiscal (4481 a 4798). A ese ritmo también caben unas veinte bajo 11 000,
  sin reserva propia: si la necesita, la pide un hito de la base.
- Lo que se mueve depende de una lectura que se mide en 38 a 40 de 40, no en 40 de 40. Por eso no se mueve ninguna
  forma fija con umbral 0.
- El hito cambia `skills/` e `internal/skills`, así que su cierre mide todas las skills sin reutilizar nada de `main`,
  y la serie de cierres de cada skill empieza de cero (ADR 0038).
- Veinte es una cifra de nombres cortos. Con nombres largos caben menos, y el tope de 11 000 sigue mandando.
