# 0029 · Umbrales que se cumplen: todo umbral medido sin persona tiene un control que lo hace fallar, y el informe distingue «tareas hechas» de «comprobado»

- **Estado**: aceptada; sustituida en parte por el ADR 0037 (el control de un umbral ya no es siempre «mecánico»:
  es un guion si lo medido tiene forma, y un juez con modelo si es el significado de una respuesta). Siguen vigentes
  el contrato de `umbrales`, «Controles de umbral» en el plan y el informe final sin modelo.
- **Fecha**: 2026-09-29
- **Hito**: transversal (workflow `hito` 2.2.0 y constitución 2.7.0, tras H7.2 y antes de H7.3). Enmienda el ADR 0028
  —añade una quinta regla a lo que es un defecto para un juez— y el ADR 0018 en lo que dice el informe final. Define
  el contrato de `umbrales` del informe del job de evals (ADR 0016), que implementa H7.3.

## Contexto y problema

H7.2 (run `6335565c`, PR #82, fusionada en `c4819d1`) tenía este criterio de éxito (SC-001 de
`specs/012-h7-2-la-consulta-repetida/spec.md`): como mucho el 5 % de las respuestas de cada modelo en las evals que
activan la skill llevan una expresión prohibida; con el conjunto actual, ≤ 2 de 51 con Sonnet 5 y ≤ 1 de 30 con
Haiku 4.5. La sección del hito lo pedía igual («Umbral: en el job de cierre, como mucho el 5 %…») y en su
«Aceptación».

**La medida del job de cierre**: 10 de 51 con Sonnet 5 (19,6 %) y 0 de 30 con Haiku 4.5
(`gates/evals/boe-legislacion.json`, clave `expresiones_prohibidas_por_modelo`). Ocho de las diez respuestas
empiezan por «Sin hallazgos que trasladar.»; otra, por «No hay hallazgos, así que respondo con el texto vigente.»; y la
de la eval 19 dice «hallazgo» dentro de su línea `⚠ REDACCIÓN MODIFICADA:`. Ocho son de evals informativas (13, 14,
15 y 19), que no deciden nunca; las otras dos, de las series de Sonnet 5 de las evals 03 y 06, que pasaron con 2 de 3.

**Aun así, todo salió en verde:**

- **El job dio «aprobado».** La lista de expresiones decide por serie con la regla del ADR 0016 (2 de 3 sesiones),
  que admite hasta un 33 % en cada serie y nada en las informativas. El 5 % solo se publicaba. No fue un descuido
  de la implementación: el propio spec lo dejó escrito en «Fuera de alcance», invocando la constitución —«Que el
  umbral del 5 % por modelo entre en el veredicto del job: el veredicto sigue la regla del ADR 0016 (FR-054) y el
  umbral se comprueba en el recuento que publica el informe (SC-001)», bajo «De lo que el hito no especifica
  (constitución, «Criterio de decisión autónoma», punto 2)»—. La lectura conservadora convirtió el control de un
  umbral que el hito pedía en alcance no especificado.
- **El cierre quedó en verde**, así que `reparar_cierre` no se ejecutó: el workflow no intentó arreglarlo.
- **La trazabilidad del informe final marcó SC-001 «hecho»** porque sus tareas (T005, T009) estaban marcadas, no
  porque algo lo hubiera medido.
- **El informe no mostraba el recuento**: hubo que sacarlo del JSON.
- **Ningún juez lo vio**: ni el del spec (2 rondas), ni el del plan (2), ni el de tareas (1), ni los dos de la
  revisión final (2 rondas). Ningún criterio pedía que un umbral medido en el cierre tuviera un control que lo
  hiciera cumplir; el criterio h del spec pide un test derivable, y del 5 % se podía derivar uno —el recuento
  publicado— que nadie hacía fallar.

H20 y H8 tendrán criterios de éxito medidos por evals: hay que corregir la clase de defecto, no este caso. La
corrección del producto —que el job falle por encima del 5 % y que `SKILL.md` deje de provocar el ruido— es H7.3.

**Un defecto menor del mismo informe.** La corrección de la primera ronda de la revisión final de H7.2 solo tocó
`quickstart.md` y `research.md`, dentro del directorio del feature. `cerrar_correccion` (`scripts/workflow/global.sh
cerrar`) no la commiteó: solo commiteaba si había cambios fuera de ese directorio, para no commitear con un mensaje
«fix» los registros de `gates/`. `cerrar_revision` commiteó después solo los veredictos, y la corrección la arrastró
`publicar` en «docs(H7.2): registros del run» (`fbdab2e`), detrás de los veredictos. El informe la listó entre los
commits «que ningún juez vio», aunque los jueces de la segunda ronda la leyeron en el árbol; y esos jueces, que
buscan la corrección en «el último commit `fix(H7.2): motivos de la revisión final`», no la encontraron como commit.

## Opciones consideradas

Para que un umbral se cumpla:

1. **Que el cierre compare cada criterio de éxito con la medida** (un paso sin modelo que lee el spec y el
   `informe.json`). Rechazada: exige leer umbrales escritos en prosa, con patrones o con un modelo (el ADR 0028
   rechazó ya clasificar con palabras clave), y pondría el control fuera del producto: solo lo tendrían los runs del
   workflow, no cada propuesta de cambio que ejecuta el job.
2. **Endurecer la regla del ADR 0016** (3 de 3 por serie en las listas). Rechazada: cambia la regla de decisión de
   todas las evals por un defecto que no es suyo, y sigue sin tocar las series informativas, donde estaban 8 de las
   10 respuestas. Una regla por serie y un umbral agregado miden cosas distintas; hacen falta los dos.
3. **Que el informe final muestre la medida y nada más.** Rechazada como solución (se hace además): publicar es
   justo lo que falló. La persona lo vería al final, pero el run no habría iterado, y la siguiente propuesta de cambio
   que rompiera el umbral tampoco fallaría.
4. **Un juez de umbrales aparte.** Rechazada por lo mismo que la opción 6 del ADR 0028: el umbral es una propiedad de
   los requisitos que ya juzga cada juez, y dos listas de motivos sobre el mismo artefacto no se ven entre sí.
5. **Un precheck que detecte umbrales en el spec** (`≤`, `%`, «como mucho») y exija su control. Rechazada: frágil y
   silenciosa cuando falla. El precheck sí comprueba la forma de lo que el plan declara; qué umbrales faltan lo
   juzga un juez.
6. **Criterio nuevo en los jueces de spec, plan y revisión final, con el control declarado en el plan en una forma
   que el informe lee sin modelo, y un contrato para que el job publique sus umbrales.** Elegida.

Para que la trazabilidad no diga «hecho» de lo que nada ha comprobado:

7. **«Comprobado» si un guion de aceptación cita el requisito.** Rechazada: que un guion mencione un id no prueba el
   umbral; en H7.2 ningún guion citaba SC-001, y en H7.1 el de 40 000 bytes es un test de integración, no un guion.
8. **Decidirlo con un modelo al escribir el informe.** Rechazada: el informe se construye sin modelo (ADR 0018).
9. **Que lo declare quien diseña el control**, como el impacto de cada supuesto en el ADR 0028: el plan nombra dónde
   vive cada control con una gramática cerrada y el informe comprueba que está. Elegida.

Para las letras de la rúbrica, **renumerar** para que el criterio común de la revisión final fuese la k (y la
proporcionalidad del juez B, la l) se rechazó: los veredictos de runs distintos dejarían de poder compararse por
letra. Cada criterio nuevo toma la letra siguiente a la última usada: j en el spec, o en el plan, l en la revisión.

Para el defecto del informe: **commitear los artefactos junto a los veredictos** en `cerrar_revision` se rechazó,
porque la corrección seguiría sin su commit y los jueces de la ronda siguiente seguirían sin encontrarla. Se corrige
en su origen.

## Decisión

1. **Umbrales que se cumplen** (constitución, «Gates», quinta regla; «Criterio de decisión autónoma», punto 2). Todo
   requisito o criterio de éxito con un umbral numérico que se mide sin persona —en `make ci` o en el cierre: el job
   de evals u otra comprobación de GitHub Actions que cuenta `medir_cierre`— tiene un control mecánico que pone en
   rojo esa comprobación cuando la medida pasa del umbral. Publicar la medida no basta, y la regla por serie del
   ADR 0016 no sustituye a un umbral agregado. Hacer cumplir un umbral que pide el hito no es alcance añadido: el
   control lo exige el propio umbral, y dejarlo fuera de alcance no es una lectura conservadora. El umbral de
   materialidad y la convergencia del ADR 0028 se aplican igual (el motivo nombra la vía por la que la medida puede
   pasar del umbral y la diferencia: la comprobación en verde con la medida por encima). No entran un umbral que solo
   mide una persona (un escenario del quickstart) ni uno del repositorio cuyo control vive en configuración que el
   hito no toca (ver «Consecuencias»).
   - **Spec** (`juez_spec`, criterio j `umbral_con_control`): cada umbral nombra una comprobación que sale en rojo.
     Que mida cada forma que promete el requisito (con y sin argumentos, cada modelo, cada ámbito) es diseño: en el
     spec es a lo sumo una observación.
   - **Plan** (`juez_plan`, criterio o `controles_de_umbral`): la sección `## Controles de umbral` dice dónde vive
     cada control, y el control mide cada forma del requisito. El corrector del plan puede tocar `spec.md` por un motivo de umbral cuya causa esté en él (el caso
     de H7.2), como ya podía por uno de uso.
   - **Tareas** (`juez_tasks`, criterio b): cada fila tiene la tarea que construye su control, con un test que lo ve
     fallar.
   - **Revisión final** (criterio común l `umbrales_que_se_cumplen`): el control existe en el código y falla por
     encima del umbral, con un mutante si hace falta.
   - **Correctores** de spec, plan y revisión, y `reparar_cierre`: nunca arreglan un motivo así rebajando el umbral,
     retirándolo o dejándolo solo publicado.
2. **`## Controles de umbral` en `plan.md`**, con la tabla `| Requisito | Umbral | Control | Dónde |`: una fila por
   umbral, o «Ninguno.» con el motivo. «Dónde» es `ci:<ruta>`, `ci:<ruta>:<Test>` (un test o una comprobación que
   ejecuta `make ci`; `<Test>` es una función Go o un objetivo del Makefile) o `evals:<skill>:<nombre>` (un umbral del
   contrato de abajo). `precheck.sh plan` comprueba la forma: que la sección existe, que cada fila nombra requisitos
   que el spec define y que dice dónde vive su control.
3. **Contrato de umbrales del job de evals** (abajo). Lo implementa H7.3 en el job; aquí se escribe el lado del
   informe final.
4. **Informe final.** En la sección 3, bajo la tabla de tasas de cada skill, el recuento de
   `expresiones_prohibidas_por_modelo` (respuestas con alguna expresión, total y porcentaje con un decimal, por
   modelo) y la tabla de `umbrales`, con los que no se cumplen, los que solo se publican y los incoherentes marcados;
   un informe del job sin `umbrales` lo dice («lo de arriba solo se publica»). En la sección 1, una línea de umbrales
   por skill. En la trazabilidad (sección 6), una columna «Control de umbral» y un estado que distingue:
   - **tareas hechas**: las tareas que citan el requisito están marcadas; nada del run lo ha medido;
   - **comprobado por su control**: su fila de «Controles de umbral» nombra un control y el control está: `ci:` con la
     ruta (y la función u objetivo) presentes en la cabeza y `make ci` en verde, o `evals:` con el umbral publicado,
     cumplido, coherente con su medida y con `decide: true`;
   - **UMBRAL NO CUMPLIDO**: el job lo publica sin cumplir;
   - **CONTROL SIN VERIFICAR**: declarado pero ausente, sin medir, solo publicado o incoherente.

   Todo sin modelo. La sección 7 dice de cada commit posterior a los veredictos lo que toca fuera de `gates/`.
5. **`global.sh cerrar`** solo deja sin commitear los cambios que se limitan a `gates/`: una corrección de los
   artefactos del feature se commitea con su mensaje, antes de los veredictos.
6. `H7.3` entra en el `enum` del input `hito`.

## Contrato de umbrales del informe del job de evals

`informe.json` (el `Informe` de `internal/evals`) lleva la clave `umbrales`: una lista, siempre presente (`[]` si la
skill no tiene ninguno). Cada elemento:

| Campo | Tipo | Significado |
|---|---|---|
| `nombre` | cadena, `^[a-z0-9_.:-]+$`, única en el informe | Lo que cita el plan en `evals:<skill>:<nombre>`. Un recuento que ya publica el informe, por modelo, se nombra `<recuento>:<modelo>`: `expresiones_prohibidas:claude-sonnet-5`. |
| `descripcion` | cadena | Una línea en español: qué se mide y sobre qué respuestas. |
| `medida` | número ≥ 0 | Lo medido: un recuento, unos bytes, unos segundos. |
| `total` | entero ≥ 0, opcional | Si está, lo que se compara es la proporción `medida / total` (0 si `total` es 0) y `umbral` es una proporción entre 0 y 1; si no, la propia `medida`. |
| `comparacion` | `"<="`, `"<"`, `">="` o `">"` | `valor <comparacion> umbral`. |
| `umbral` | número | En la unidad de lo que se compara. |
| `cumple` | booleano | El resultado de la comparación, calculado por el job en coma flotante de doble precisión y sin redondeos. |
| `decide` | booleano | `true`: incumplirlo pone el veredicto en `fallo`, con un motivo que nombra el umbral, y el job sale en rojo. `false`: solo se publica, y entonces no es un control. |

Invariantes que el informe final comprueba sin modelo y marca si no se dan: `cumple` es el resultado de rehacer la
comparación; y un umbral con `decide: true` y `cumple: false` implica `veredicto: "fallo"`. Un informe sin la clave
`umbrales` es anterior a este contrato: el informe final lo dice, y no da ningún requisito por comprobado desde él.

El umbral de SC-001 de H7.2, con este contrato, serían dos elementos —`expresiones_prohibidas:claude-sonnet-5` con
`medida` 10, `total` 51, `comparacion` `"<="`, `umbral` 0.05, `cumple` `false`, `decide` `true`, y el de Haiku 4.5—, y
el veredicto habría sido `fallo`.

## Validación

Con `scripts/paso.sh`, en clones fuera del repositorio, con la rama aplicada encima del estado exacto que juzgó cada
run y los permisos de `.claude/settings.json` del run. Cada juez juzga como en su primera ronda (sin veredictos ni
diffs de corrección anteriores), después del precheck de su fase, como en el workflow. Spec y plan de H7.2 en
`31726a7`; spec y plan de H7.1 en `9df6c82`; la revisión final sobre el código de H7.2 que se fusionó (`828df2b`,
con `main` en la base del run). Los clones **no** llevan este ADR: una primera tanda sí lo llevaba, un juez citó
«el ADR 0029 registra justo ese caso» y se descartó entera, porque el ADR describe el defecto de H7.2.

| Juez | Sobre | Ejecución | Veredicto | Motivos del criterio nuevo |
|---|---|---|---|---|
| `juez_spec` | H7.2 | 1 | rechazado (solo j) | [j] SC-001: se mide en el job sin persona y ninguna comprobación sale en rojo; «Fuera de alcance» lo excluye y FR-054 deja la regla por serie como única |
| `juez_spec` | H7.2 | 2 | rechazado (f, j) | [j][f] SC-001, lo mismo; el checklist lo daba por medible |
| `juez_plan` | H7.2 | 1 | rechazado (solo o) | [o] SC-001 solo se publica (research D7, contrato lista-y-juicio §5); con 3 de 51 el veredicto sigue `aprobado`; la causa, en el spec |
| `juez_plan` | H7.2 | 2 | rechazado (solo o) | [o] SC-001, lo mismo; cita la viñeta de «Fuera de alcance» que el corrector del plan tiene que quitar |
| `revision_juez_a` | H7.2, código | 1 | rechazado (f, j, l) | [l] SC-001 solo se publica: `veredictoDelInforme` (`internal/evals/informe.go`) no mira el recuento, y el informe de cierre da `aprobado` con 10 de 51 (19,6 %) |
| `juez_spec` | H7.1 | 1 | rechazado (f, h, j) | [j] SC-005 (≤ 3 800 bytes y ≤ k hallazgos con cinco bloques): ningún requisito nombra la comprobación. Ninguno por la cota de 40 000 |
| `juez_spec` | H7.1 | 2 | rechazado (f, h, j) | [h][j] SC-005, lo mismo. Ninguno por la cota de 40 000 |
| `juez_plan` | H7.1 | 1 | rechazado (m, o) | [m][o] SC-005; [o] la cota de 50 hallazgos de FR-010 «en ninguna de sus formas» solo se prueba sin argumentos. Ninguno por la cota de 40 000 |
| `juez_plan` | H7.1 | 2 | rechazado (h, j, m, o) | [o] SC-005; [o] la cota de 50 de FR-010, lo mismo. Ninguno por la cota de 40 000 |

**Control de falsos positivos (H7.1).** La cota de ≤ 40 000 bytes la hace cumplir `TestMedidaDelGrafo` en
`make ci`. Ninguna de las cuatro ejecuciones da un motivo por ella: los jueces la citan como ejemplo de umbral con su
control («FR-013/SC-001/SC-002 → TestMedidaDelGrafo en test-integration»). Lo que sí señalan son aciertos sobre el
artefacto que juzgaron:

- **SC-005** no tenía control en el spec ni en el plan de `9df6c82`: el test que lo hace cumplir,
  `internal/app/medida_test.go`, llegó después, en la implementación.
- **La cota de 50 hallazgos con la norma** sigue sin test en `main`: `TestComprobarConLaCota` solo usa `Ambito{}`,
  aunque el código aplica la cota en la misma función para todas las formas.

Una ejecución intermedia del juez del spec pidió que el spec nombrase también el control de la forma «solo con la
norma» de la cota de 40 000. Es diseño, y la rúbrica se ajustó: en el spec basta con nombrar una comprobación que
falla, y la cobertura de cada forma la juzga el plan (criterio o). Con ese texto, las dos ejecuciones de arriba la
dejan en observaciones («Cubrir cada forma es diseño del plan»).

**La revisión final** señala, además del [l], la causa en `SKILL.md` que H7.3 tiene que corregir: la regla 7 empareja
«sin hallazgos» con «trasladar», que es lo que repiten ocho de las diez respuestas.

**Informe.** `scripts/workflow/informe.sh H7.2 --solo-ver` sobre la rama del run (`828df2b`) da:

- en la sección 1, «boe-legislacion: el job no publica umbrales; publica recuentos que ningún control hace cumplir
  (expresiones prohibidas: 10/51 `claude-sonnet-5`, 0/30 `claude-haiku-4-5-20251001`)»;
- en la sección 3, la tabla de `expresiones_prohibidas_por_modelo`: 10 de 51 (19,6 %) y 0 de 30 (0,0 %), seguida de
  «el informe del job no trae `umbrales` […] lo de arriba solo se publica»;
- en la trazabilidad, SC-001 como «tareas hechas», con la nota de que el plan no tiene «Controles de umbral»;
- y en la sección 7, que «registros del run» (`fbdab2e`) tocaba `plan.md`, `quickstart.md` y `research.md`.

Con informes sintéticos, las demás ramas dan lo esperado: cumple y decide («comprobado por su control»), no cumple
(«UMBRAL NO CUMPLIDO», e «INCOHERENTE» si el veredicto es `aprobado`), `decide: false` y umbral ausente («CONTROL SIN
VERIFICAR»), `cumple` que no casa con la medida, y controles `ci:` presentes y ausentes. `precheck.sh plan` marca la
falta de la sección en los planes de H7.1 y H7.2, una fila sin control y un requisito que el spec no define.

## Consecuencias

**A favor**

- Un umbral que el hito pide se cumple o el run lo ve en rojo e itera: el job falla, `medir_cierre` sale en rojo y
  `reparar_cierre` actúa. Y lo mismo en cada propuesta de cambio posterior, porque el control vive en el producto.
- La lectura conservadora deja de poder quitarle el control a un umbral del hito.
- La persona ve en el informe el recuento y el umbral juntos, y la trazabilidad solo dice «comprobado» de lo que un
  control comprueba.
- Una corrección de la revisión final que solo toca artefactos lleva su commit y la ven los jueces de la ronda
  siguiente como tal.

**En contra, y asumido**

- La completitud de «Controles de umbral» depende de los jueces: un umbral sin fila figura como «tareas hechas», no
  como comprobado. El fallo es de omisión en la tabla, no de un «comprobado» falso.
- Para un control `ci:` el informe solo comprueba que existe y que `make ci` está en verde; que falla por encima del
  umbral lo comprueba el juez de la revisión final, con un mutante si hace falta.
- **La cobertura es de esta clase y queda fuera.** Los umbrales de `codecov.yml` (70 % global, 85 % `internal/core`)
  solo los mide Codecov, cuyo estado `medir_cierre` trata como informativo y la protección de `main` no exige. Su
  control vive en configuración que ningún hito toca (el guardián global la protege), así que el criterio no la cuenta
  como motivo de un hito; arreglarlo es otro cambio del proceso, pendiente.
- H7.2 sigue sin cumplir su SC-001 en `main` hasta H7.3, y no hay release hasta entonces.
