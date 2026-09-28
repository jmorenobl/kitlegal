# 0028 · Rúbrica de los jueces: materialidad, uso de fuera adentro, proporcionalidad y convergencia; informe final por impacto

- **Estado**: aceptada
- **Fecha**: 2026-09-28
- **Hito**: transversal (workflow `hito` 2.1.0 y constitución 2.6.0, tras H7 y antes de H7.1). Enmienda el ADR 0018
  en lo que juzgan los jueces y en lo que dice el informe final; no cambia la forma del run, las causas mayores ni
  quién decide.

## Contexto y problema

H7 (run `c378ced4`, PR #77) salió limpio según las medidas del ADR 0018: cero paradas, 29 de 29 tareas y la revisión
final aprobada en 2 rondas (H6: 13). La revisión posterior encontró tres defectos del proceso, no del run.

**A · Inflado por falta de materialidad.** Spec y plan agotaron sus cuatro rondas. El juez del spec dio 13, 9, 6 y 3
motivos; el del plan, 7, 3, 3 y 3. Casi todos los del criterio h del spec («sin huecos de comportamiento que el spec
implique y no defina») y de los criterios i y a del plan hablaban de estados a los que el producto no llega o cuyas
lecturas nadie ve: si «letra», «cifra» y «espacio» del patrón de NIF de `Persona` son ASCII o Unicode, cuando ningún
applet emite `Persona` hasta H17; un `world.db` de 0 bytes con un `-wal` no vacío; un `-shm` suelto; el diario de otra
aplicación; un `world.db` con `chmod a-w`; una base escrita por otra versión de SQLite con `auto_vacuum`. Tres piezas
hacen de trinquete: el criterio h, que da por defecto cualquier par de lecturas que un test pueda distinguir; la orden
de «enumerar TODOS los incumplimientos», sin umbral; y los correctores, que generalizan cada motivo «a su clase de
defecto» sin límite. Cada corrección añade texto y el texto nuevo abre huecos nuevos: de las 17 líneas de
`gates/supuestos.md` que escribieron los correctores de spec y plan, 8 fijan comportamiento para estados a los que el
producto no llega (`Persona` sin emisor, lotes que solo construye un test, bases y ficheros tocados desde fuera) y
otras 2 en parte. El
resultado: 62 FR y 13 SC, 3 177 líneas de artefactos de spec y plan, 112 supuestos más 5 respuestas de clarify, y
5 221 líneas de Go de producción frente a 17 649 de tests (`internal/graph`: 10 384 entre las dos).

**B · Nadie mira el uso de fuera adentro.** La constitución manda planificar de fuera adentro (skill → herramientas),
pero ningún criterio de ningún juez comprobaba que el diseño sirve a quien lo consume. Un defecto real de
`graph check` pasó por cinco jueces (spec, plan, tasks y los dos de la revisión): recorre el grafo entero sin filtro
(FR-060 prohíbe argumentos); da `fuente-caducada` por cada nodo consultado hace más de 7 días, a unos 680 bytes por
hallazgo (FR-066); da `version-obsoleta` para siempre, porque no depende del tiempo (FR-063), y la skill dirá «la
redacción ha cambiado» en cada consulta; y `boe-legislacion` lo llama dos veces por pregunta (FR-080). Medido: un
`world.db` con dos artículos envejecido 18 días da 5 hallazgos y 3,4 KB; con unos cien artículos consultados salen
unos 140 KB por llamada, dos veces por pregunta. Lo arregla H7.1; aquí es el caso de prueba.

**C · El informe final no dice a la persona qué mirar.** `scripts/workflow/informe.sh` volcaba los 117 supuestos en
el orden en que se escribieron: las decisiones que cambian el producto (qué verbos emiten, dos comprobaciones por
pregunta, `check` sin filtro) quedaban entre detalles de SQLite o de RFC 8785. No daba la tasa de ninguna eval: la de
la eval informativa nueva (3 de 3) hubo que sacarla del registro del job. Y su sección 5 decía «2 fixtures o esquemas
nuevos» sin nombrarlos; eran cuatro fixtures y un esquema, porque el patrón no veía `internal/app/testdata/`.

## Opciones consideradas

Para A y B:

1. **Menos rondas** (dos en lugar de cuatro). Rechazada: trata el síntoma. Un juez sin umbral da en dos rondas los
   mismos motivos inmateriales, y los defectos reales que hoy caen en la ronda 3 dejarían de corregirse.
2. **Quitar la revisión exhaustiva** («enumera todos»). Rechazada: la exhaustividad es lo que cortó el goteo de H6
   (13 rondas). Lo que falla no es enumerar todo, sino que «todo» no tenga borde.
3. **Un tope mecánico de tamaño** (FR por hito, líneas de spec). Rechazada: el tamaño no es el defecto sino el
   contenido que no sirve; un hito grande tiene muchos requisitos legítimos, y un tope haría recortar al azar.
4. **Que el corrector pueda rechazar motivos.** Rechazada: rompe la separación juez/corrector de la constitución (el
   corrector aplica, no discute) y convierte al corrector en un segundo juez sin rúbrica ni evidencia.
5. **Pedir al juez «proporcionalidad» en abstracto.** Rechazada: es lo que había de hecho —el juez elegía qué era
   importante— y no se puede comprobar. Una regla con dos condiciones que el motivo tiene que nombrar («vía: …;
   diferencia: …») sí se puede auditar veredicto a veredicto.
6. **Un juez de uso aparte**, en su propio paso. Rechazada: el uso es una propiedad de los mismos requisitos que
   juzga el juez del spec, y dos jueces sobre el mismo artefacto producen dos listas de motivos que el corrector
   aplica sin que ninguno vea la otra; la rúbrica del juez de cada artefacto es donde se ve todo junto.
7. **Umbral, uso, proporcionalidad y convergencia dentro de los jueces existentes, con el diff de cada corrección.**
   Elegida.

Para C:

8. **Clasificar los supuestos en el informe con palabras clave** (SQLite, RFC, test…). Rechazada: frágil y silenciosa
   cuando se equivoca, que es justo cuando esconde algo importante.
9. **Clasificarlos con un modelo al escribir el informe.** Rechazada: el informe se construye sin modelo (ADR 0018),
   y es lo que la persona lee para decidir.
10. **Que cada escritor etiquete su supuesto**: quien toma la decisión sabe qué cambia. Elegida. Para las tasas de
    evals, subir el informe del job como artefacto de GitHub cambiaría `.github/workflows/evals.yml` y su contrato
    (H5); el job ya imprime `informe.json` entero entre dos marcas en su registro, y de ahí se lee.

## Decisión

1. **Umbral de materialidad** (constitución, «Gates»; todos los jueces; correctores y `resolver_clarify`). Una
   ambigüedad, un hueco o un estado solo es defecto si (a) el producto puede llegar a él por una vía real —un emisor
   que existe en el repositorio o que entrega el hito; un argumento, una bandera o una variable de entorno de quien
   usa el kit; o un estado que crea el propio binario, incluidas sus invocaciones concurrentes y la interrupción de
   sus propias escrituras cuando el hito promete atomicidad— y (b) sus lecturas dan un resultado observable distinto
   para una persona o una skill: código de salida, `data`, salida legible, ficheros creados o cambiados, respuesta de
   la skill. Que un test pueda distinguirlas no basta. El motivo nombra las dos («vía: …; diferencia: …»). Lo que llega
   de fuera manipulado lo cubre una regla genérica —defecto `inesperado`, código 1 (ADR 0023)—, sin especificar bytes,
   ficheros ni recuperación; si un artefacto ya lo especifica caso a caso, el motivo es retirarlo (criterio a del spec,
   n del plan). Los correctores generalizan un motivo a su clase solo dentro del umbral y, entre dos arreglos, eligen
   el que deja el artefacto más corto. `resolver_clarify` responde «no material» (criterio `e`) a lo que no lo pasa, y
   la tarea `[aceptacion]` no afirma nada que el spec deje abierto por eso.
2. **Uso, de fuera adentro** (criterio i de `juez_spec`, m de `juez_plan`, j de los dos jueces finales). Cada salida
   que consume una skill o una persona dice quién la pide y cuántas veces por pregunta, qué tamaño tiene con un uso
   realista sostenido (meses de uso diario: cientos de normas y miles de bloques consultados) y cuándo deja de darse
   cada señal. La evidencia es un cálculo (spec y plan) o una medida en el binario (revisión). El corrector del plan
   puede tocar `spec.md` solo para un motivo de uso cuya causa esté en el spec: un defecto de fuera adentro se
   arregla en el requisito, no se arrastra como supuesto hasta la implementación.
3. **Proporcionalidad** (criterio n de `juez_plan`, k del juez B de la revisión; constitución V). Cada mecanismo se
   traza a un FR/SC que lo exige; lo que no se traza, o solo sirve a un estado por debajo del umbral, se retira.
4. **Convergencia** (todos los gates). Desde la segunda ronda, un motivo nuevo solo cuenta si lo introdujo la
   corrección anterior o si es material. `gate.sh leer` guarda una instantánea de lo juzgado en el git-dir (fuera del
   historial) y el paso nuevo `cambios_<fase>`, al empezar la ronda siguiente, deja en
   `gates/<fase>-correccion-r<n>.diff` exactamente lo que cambió el corrector; en la revisión final, la corrección es
   su commit. Lo que no llega a motivo va a `observaciones` del veredicto: se archiva, no cuenta y no llega al
   corrector.
5. **Informe por impacto.** Cada línea de `gates/supuestos.md` empieza por su impacto —`[comportamiento]`,
   `[alcance]`, `[skill]`, `[interno]`, o `[proceso]` si la escribe un paso shell— y cada respuesta de clarify lleva
   `impacto`. El informe lista enteros los que cambian el comportamiento visible, el alcance o una skill; después los
   del propio run (gates agotados y sus motivos, observaciones de los jueces); después los que no llevan etiqueta,
   enteros, porque nunca se esconde lo que no se ha clasificado (runs anteriores a la 2.1.0; de clarify, lo que dejó
   algo fuera cuenta como alcance); y los internos solo contados por autor. `cierre.sh medir` copia el `informe.json`
   de cada trabajo de evals a `gates/evals/<skill>.json` (`cierre.sh evals <hito>` lo hace para un run anterior) y el
   informe da una tabla por skill con la tasa de cada eval y modelo, con las nuevas, las cambiadas y las informativas
   marcadas. Los fixtures y esquemas nuevos van nombrados, y la suite de aceptación activada aparte. Las secciones
   se reordenan para que lo que puede cambiar la decisión vaya antes que la trazabilidad.

## Validación sobre H7

Con los jueces nuevos, ejecutados con `scripts/paso.sh` en clones fuera del repositorio sobre el estado exacto que
juzgó el run: `f2dc11c` (spec y plan tal como los vio la ronda 4) y `8471054` (el código que vio la revisión final),
cada juez como en su primera ronda, con los permisos de `.claude/settings.json` del run.

| Juez | Ejecución | Veredicto | Motivos de uso | Piden tratar estados inalcanzables | Piden retirarlos |
|---|---|---|---|---|---|
| `juez_spec` | 1 | rechazado | 2: tamaño (~5 600 hallazgos, ~650 B cada uno, dos veces por pregunta) y señales que no caducan | 0 | 3 |
| `juez_spec` | 2 | rechazado | 2: los mismos | 0 | 2 |
| `juez_plan` | 1 | rechazado | 1: tamaño (~2,4 MB por invocación, ~4,9 MB por pregunta) y permanencia; la causa, en FR-060, FR-063, FR-066 y FR-080 | 0 | 3 |
| `juez_plan` | 2 | rechazado | 2: tamaño y permanencia de `version-obsoleta` | 0 | 6 |
| `revision_juez_b` | 1 | rechazado | 1, medido en el binario: 300 normas sembradas → 4 812 hallazgos y 3 045 524 bytes por `check`, ~6,1 MB por pregunta | 0 | 0 (5 observaciones: el código se traza al spec aprobado) |

Ninguno dio un motivo que pidiera especificar, medir o declarar un estado como los de `spec-pendiente.md` y
`plan-pendiente.md`; los que los nombran piden quitarlos (criterio a del spec, n del plan). Los hallazgos menores
—referencias de línea, marcas de checklist, notas— quedaron en `observaciones` (7 a 10 por ejecución).

**Control.** En una copia del spec se cambió FR-053 para que `graph show` con un id ausente saliera con 0 y `data` nulo
(contra FR-004, US4.3 y los casos límite, que dicen 3) y se quitó de FR-060 que `graph check` rechaza argumentos. El
juez detectó los dos: la contradicción como motivo [e][h][f] y el hueco como [h], con «vía: la skill o una persona
escribe `kitlegal graph check BOE-A-2015-10565`; diferencia: 2 frente a 0 con todo el grafo». La rúbrica no es solo
más permisiva.

**Motivos antiguos con la regla nueva.** De los 31 del spec, 18 seguirían contando, 6 en parte y 7 no (el patrón de
NIF de `Persona` sin emisor, empates de fecha que solo produce el reloj fijado de los tests, lotes que solo construye
un test). De los 16 del plan, 4 sí, 2 en parte y 10 no: todos los de las rondas 2 a 4 (bases de otra aplicación,
`chmod a-w`, `-wal` junto a un `world.db` de 0 bytes, `-shm` suelto, una referencia de línea). Corregidos los de la
ronda 1, el plan no habría tenido motivos en la 2. El spec seguiría teniendo trabajo real, y los motivos de uso lo habrían orientado al problema de
`graph check` desde la ronda 1.

**Informe.** `informe.sh H7 --solo-ver` sobre la rama del run da las tasas de las 22 evals de las dos skills (la nueva
de H7, 3 de 3, marcada nueva e informativa; 7 informativas en total), los cuatro fixtures y el esquema nuevos por su
nombre, y los 113 supuestos del run como «sin clasificar», enteros, porque son anteriores a la etiqueta. Con esos
supuestos etiquetados según la regla, el informe baja de 79 a 50 KB y deja arriba los 50 que cambian el
comportamiento, el alcance o una skill, y las 59 decisiones internas solo contadas.

## Consecuencias

**A favor**

- Cada corrección acerca el artefacto a lo que el hito pide: un motivo o es material o no se corrige, y la ronda
  siguiente ve exactamente qué cambió.
- Un defecto de uso se ve en el spec, que es donde cuesta menos arreglarlo, y la revisión final lo mide en el binario.
- La persona lee primero lo que puede hacerle cambiar la entrada, ve las tasas de evals sin abrir el registro del job
  y sabe qué fixtures nuevos hay.

**En contra, y asumido**

- El umbral depende de que el juez nombre bien la vía y la diferencia; un juez puede equivocarse al negar una vía.
  Lo acota que el motivo y la observación quedan escritos: una observación mal clasificada se ve en el veredicto.
- Un estado de fuera que antes tenía comportamiento especificado (un `world.db` con permisos cambiados) pasa a ser un
  defecto `inesperado` con código 1, sin promesas sobre sus ficheros. Es deliberado: prometer bytes sobre lo que el
  binario no controla era la fuente del inflado.
- El informe depende de que cada escritor etiquete. Lo que llega sin etiqueta se enseña entero, así que el fallo es
  de orden, no de ocultación.
- Las tasas de evals salen del registro del job, que GitHub conserva un tiempo limitado; el cierre las copia en el
  momento de medir, cuando el registro existe.
