# 0045 · Lo reutilizado no cuenta en la tasa; el límite de marcadas se deriva para las respuestas medidas

- **Estado**: propuesta (2026-10-22), escrita en el run de H30; queda aceptada cuando una persona fusiona la propuesta
  de cambio del hito (ADR 0018).
- **Fecha**: 2026-10-22
- **Hito**: H30, las evals que solo miden lo que cambia (FR-024, SC-010 y SC-013 de su spec). Precisa el ADR 0038 en
  el número de respuestas de su regla (punto 3), que contaba con que una ejecución las mide todas, y mantiene su
  criterio: una ejecución sale en rojo sin motivo como mucho 1 de cada 50 veces. No toca el umbral de ninguna clase
  (`evals/*/juez/clases.yaml`; ADR 0038, puntos 2, 4 y 5), ni el modelo que decide (ADR 0031), ni el juez (ADR 0037),
  ni los guiones del workflow: `scripts/workflow/serie.sh`, `cierre.sh` e `informe.sh` siguen como están (FR-027 de
  H30).

## Contexto y problema

El ADR 0038 puso una tasa donde hay significado. El umbral `afirma_lo_no_leido:<modelo>` cuenta las respuestas del
modelo que decide con los dos modos juntos y se cumple si las marcadas, sobre el total, no pasan del `umbral` de su
clase en `clases.yaml`: 0.028 en `boe-legislacion` y 0.017 en `jurisprudencia`. Ese número se fijó para una ejecución
completa: es el más estricto con el que la ejecución, a la tasa de la skill, sale en rojo sin motivo como mucho 1 de
cada 50 veces (punto 3), y se calculó con todas sus respuestas: 108 en `boe-legislacion` entonces (120 hoy; la serie
enseña el cambio en la medición H20-1) y 60 en `jurisprudencia`. Con todas, la comparación del job da como mucho 3
marcadas de 120 y 1 de 60.

H30 cambia lo que mide una ejecución. El job mide solo las evals que una propuesta ha podido cambiar y reutiliza el
resultado de las demás de una medida de `main` con su misma huella (FR-012). En una ejecución con reutilización, el
umbral tiene menos respuestas medidas que las de la ejecución completa: *n* < *N*. Quedan dos preguntas: si las
respuestas reutilizadas cuentan, y con qué límite se juzgan *n* respuestas medidas.

**Con menos respuestas, la proporción de `clases.yaml` ya no cumple el criterio.** Aplicada tal cual, el límite es el
mayor entero que no pasa de `umbral`·*n*, y baja a saltos: con 35 respuestas de `boe-legislacion`, 0. A la tasa con la
que la ejecución completa cumple 1 de cada 50, una ejecución con 35 respuestas medidas saldría en rojo sin motivo 1 de
cada 4 veces («Validación»). Es el sorteo que el ADR 0038 quitó, devuelto por el número de respuestas.

**Contar lo reutilizado lo cuenta dos veces.** Una respuesta reutilizada se abrió, se juzgó y entró en el umbral de
la medida de `main` en que se abrió su sesión. Contada otra vez, pesaría en cada propuesta que la reutiliza, mientras
su eval no cambie: una marca de `main` dejaría a cada una de esas propuestas con un margen menor, sin que nada de lo
que cambian la haya producido. Y la serie de cierres (ADR 0038, punto 6), que `serie.sh` escribe con los campos del
informe y comprueba contra sus umbrales, no cuenta lo reutilizado: la entrada de H30 decidió que no entra en los
campos que lee el guion (FR-023).

## Opciones consideradas

1. **Contar lo reutilizado en el umbral.** Rechazada: cuenta dos veces cada respuesta reutilizada, y el umbral dejaría
   de casar con la fila de la serie, que `serie.sh` cuenta solo con lo medido y comprueba contra los umbrales del
   informe: si no casan, `serie.sh fila` sale con 1 y no escribe la fila.
2. **Aplicar a lo medido la proporción de `clases.yaml`.** Rechazada: con menos respuestas, el rojo sin motivo pasa de
   1 de cada 50 a 1 de cada 4 en el peor caso de `boe-legislacion` y a casi 1 de cada 5 en el de `jurisprudencia`
   («Validación»).
3. **Derivar la tasa de referencia de la tasa de la serie.** Rechazada: la tasa acumulada que pide el punto 3 del ADR
   0038 es la de la skill como está sobre 500 respuestas al menos, y hoy tiene 120 («La serie de `boe-legislacion`»);
   y movería el límite sin que una persona tocara el umbral (punto 4).
4. **Publicar siempre k(*n*)/*n* como `umbral`** (research D12 de H30). Rechazada: con 120 respuestas daría 0,025 en
   lugar de 0,028 y cambiaría el informe de la suite completa (SC-012).
5. **Aplicar la regla nueva solo con algo reutilizado** (research D12 de H30). Rechazada: dejaría dos reglas para el
   mismo umbral.
6. **Lo reutilizado no cuenta, y el límite de las respuestas medidas sale de una tasa de referencia derivada del
   umbral vigente.** Elegida.

## Decisión

1. **Lo reutilizado no cuenta en la tasa.** El umbral `<clase>:<modelo>` y el recuento `<clase>:<modelo>:<modo>` se
   aplican solo a las respuestas medidas del modelo que decide: el primero, con los dos modos juntos, decide; el
   segundo, uno por modo, se publica sin decidir, como hasta ahora. De lo reutilizado, el informe da aparte las
   marcadas y el total, en `<clase>:<modelo>:reutilizadas`, que no decide (FR-015; contracts/informe-del-job.md §3 de
   H30).
2. **La tasa de referencia se deriva del umbral vigente.** De cada skill con juez y de su clase que decide:
   - *N* son las respuestas del modelo que decide que juzga una ejecución completa en sus dos modos: las de las evals
     que activan la skill y tienen binario o servidor, por las repeticiones del job y por los dos modos;
   - *K*, el límite de hoy, es el mayor *m* de 0 a *N* con *m*/*N* ≤ `umbral`, con la comparación del job;
   - *p*ref es la mayor tasa, redondeada hacia abajo a centésimas de punto porcentual, con
     P(Bin(*N*, *p*ref) > *K*) ≤ 0,02: con ella el umbral de la ejecución completa cumple aún 1 de cada 50, y con la
     centésima siguiente ya no.

   El ADR 0038 fijaba el umbral desde la tasa acumulada; esta regla va al revés para lo que la suya no cubría: dado el
   umbral que decidió una persona, deriva la mayor tasa que tolera y la lleva a cualquier número de respuestas.
3. **El límite para *n* respuestas medidas** es k(*n*), el *k* más estricto con P(Bin(*n*, *p*ref) > *k*) ≤ 0,02. El
   umbral `<clase>:<modelo>` se cumple si las marcadas medidas no pasan de k(*n*). Con todas las respuestas,
   k(*N*) = *K*: el umbral de hoy, y la suite completa da el mismo informe que antes de H30 (FR-026).
4. **Cómo se publica.** El `umbral` del informe es el de `clases.yaml` si la comparación del job con él da k(*n*), y
   si no, k(*n*)/*n*. Así `scripts/workflow/informe.sh`, que rehace `medida/total ≤ umbral`, llega al mismo resultado
   que el job sin cambiar (research D12 y contracts/informe-del-job.md §4 de H30).
5. **La serie de cierres cuenta solo lo medido.** Los campos del informe que lee `serie.sh` cuentan solo lo medido
   (FR-023): la fila de una ejecución con reutilización es la de sus respuestas medidas, cada respuesta entra en la
   serie una vez, en la medición que abrió su sesión, y `serie.sh` no cambia.
6. **Ningún umbral sube.** `clases.yaml` no cambia. Si una persona estrecha el `umbral` de una clase (ADR 0038,
   punto 4), *p*ref baja con él, y con ella el límite: ningún k(*n*) sube y k(*N*) baja. Subir un umbral pediría otro
   ADR.

`internal/evals/tasa.go` hace la cuenta con enteros, sin redondeo, y `TestLimiteDeLaTasa`, en `make ci`, comprueba para
cada skill del job con juez y cada *n* de 1 a *N*, con racionales exactos, que P(Bin(*n*, *p*ref) > k(*n*)) ≤ 0,02 y
que con k(*n*) − 1 ya pasaría; que k(*N*) es el límite de hoy; y, con el `umbral` de `boe-legislacion` estrechado en
memoria, que el límite se estrecha.

### Las cifras de hoy

Son el resultado de la derivación, no parte de la decisión: si cambian el plan de las evals o el `umbral`, el test da
otras, y esas son las que rigen. Salida de `go test -count=1 -v -run '^TestLimiteDeLaTasa$' ./internal/evals/`,
ejecutada el 2026-10-22 en la rama de H30:

```text
tasa_test.go:76: jurisprudencia, afirma_lo_no_leido (umbral 0.017): pref = 0,36 %, N = 60, k(N) = 1
tasa_test.go:76: boe-legislacion, afirma_lo_no_leido (umbral 0.028): pref = 0,85 %, N = 120, k(N) = 3
tasa_test.go:88: boe-legislacion con umbral 0.017: pref = 0,47 %, N = 120, k(N) = 2
```

| Skill | Clase que decide | `umbral` | *N* | k(*N*) | *p*ref |
|---|---|---|---|---|---|
| `boe-legislacion` | `afirma_lo_no_leido` | 0.028 | 120 | 3 | 0,85 % |
| `jurisprudencia` | `afirma_lo_no_leido` | 0.017 | 60 | 1 | 0,36 % |

La tercera línea es el estrechamiento: con `umbral: 0.017` en `boe-legislacion`, *p*ref baja a 0,47 % y k(120) a 2.

### La serie de `boe-legislacion`

Salida de `scripts/workflow/serie.sh mostrar boe-legislacion`, el guion del workflow sin cambios, ejecutada el
2026-10-22:

- La clave de la cabeza: la skill `27aeccc` (árbol de `skills/boe-legislacion`), decide `claude-sonnet-5-5`, juez
  `claude-opus-5-5` con Claude Code 2.1.289.
- «Tasa acumulada con la skill como está: 1 de 120 (0,83 %), de H26.1-1.» Y «Insuficiente: 120 de 500 respuestas»:
  con menos de 500, la regla 3 del ADR 0038 no da escalón.
- Las 13 filas de la serie, con todas las versiones de la skill, de H24-1 a H26.1-1, suman 18 marcadas de 1.424
  respuestas (1,26 %).

*p*ref no es una tasa medida: es la mayor con la que el umbral vigente cumple aún el criterio. La serie entera está por
encima, y a esa tasa una ejecución completa saldría en rojo sin motivo más de 1 de cada 50 veces, porque la
probabilidad crece con la tasa. La tasa acumulada con la skill como está, 1 de 120, queda justo por debajo, y con menos
de 500 respuestas no cuenta para la regla 3. Ninguna de las dos mueve el umbral: solo se estrecha (punto 4 del ADR
0038), y subirlo pediría otro ADR.

## Validación

Probabilidad de que una ejecución salga en rojo sin motivo si la tasa de la skill es su *p*ref, con la proporción de
`clases.yaml` aplicada a lo medido (opción 2) y con k(*n*), y el `umbral` que publica el informe con k(*n*):

| Skill, respuestas medidas | Límite con la proporción | Rojo sin motivo | k(*n*) | Rojo sin motivo | `umbral` publicado |
|---|---|---|---|---|---|
| `boe-legislacion`, 30 | 0 | 22,6 % | 2 | 0,21 % | 0,0667 (2/30) |
| `boe-legislacion`, 60 | 1 | 9,3 % | 2 | 1,47 % | 0,0333 (2/60) |
| `boe-legislacion`, 90 | 2 | 4,2 % | 3 | 0,75 % | 0,0333 (3/90) |
| `boe-legislacion`, 120 | 3 | 1,97 % | 3 | 1,97 % | 0.028 |
| `jurisprudencia`, 30 | 0 | 10,3 % | 1 | 0,53 % | 0,0333 (1/30) |
| `jurisprudencia`, 60 | 1 | 1,998 % | 1 | 1,998 % | 0.017 |

El peor caso con la proporción es *n* = 35 en `boe-legislacion` (25,8 %) y *n* = 58 en `jurisprudencia` (18,9 %). Con
k(*n*), ningún *n* de 1 a *N* pasa de 1 de cada 50: es lo que comprueba `TestLimiteDeLaTasa`. El caso de 60 respuestas
de `boe-legislacion` (con 1 y con 2 marcadas se cumple y con 3 no, con `umbral` 2/60) lo fijan además los tests del
umbral conjunto (`TestUmbralConjuntoConMenosRespuestas`).

Calculadas el 2026-10-22 con las funciones de `internal/evals` —`nuevoLimiteDeLaTasa` y su `para` de `tasa.go`,
`compararConLaTasa` de `umbrales.go`, y `mayorMarcadasDentro` y `colaDeLaBinomial` de `tasa_test.go`, en racionales
exactos— desde un test de un solo uso que no se versiona; los porcentajes van redondeados.

## Consecuencias

**A favor**

- El criterio del ADR 0038 se cumple con cualquier número de respuestas medidas, no solo con la ejecución completa, y
  lo comprueba `make ci`.
- Con todas las respuestas no cambia nada: el mismo límite, el mismo `umbral` publicado y el mismo informe
  (`TestInformeDeLaSuiteCompletaNoCambia`). La release exige esa suite completa del commit etiquetado (FR-025).
- `serie.sh`, `cierre.sh` e `informe.sh` no cambian: la fila de una ejecución con reutilización es la de lo medido
  (`TestFilaConReutilizadas`, SC-009), y cada respuesta entra en la serie una vez.
- *p*ref no es una cifra más que mantener: sale de `clases.yaml` y del plan, y estrechar el `umbral` estrecha el
  límite para todo *n*.

**En contra, y asumido**

- **Con pocas respuestas medidas, el límite es una proporción mayor que la de `clases.yaml`** (2 de 30 en
  `boe-legislacion`, 0,0667 publicado): una ejecución parcial ve una regresión peor que la completa. Un cambio de la
  skill —su `SKILL.md`, sus referencias o su juez— está en el conjunto de todas sus evals y las mide todas
  (data-model §3 de H30), así que la ejecución parcial solo se da cuando la skill y su juez no han cambiado.
- **Las marcas de una eval reutilizada no deciden en la propuesta que la reutiliza.** Decidieron en la medida de
  `main` que abrió sus sesiones, y el informe las enseña aparte.
- ***p*ref queda por debajo de la tasa de la serie entera** (1,26 %). Este ADR no lo corrige: mejorar la skill es
  trabajo normal, que se mide con la serie (ADR 0038, punto 8).

## Lo que este ADR no decide

- El umbral de ninguna clase, ni cuándo se estrecha (ADR 0038, punto 4).
- La tasa acumulada de la regla 3 del ADR 0038 ni su mínimo de 500 respuestas.
- Ningún cambio de `scripts/workflow/`: la decisión no lo pide.

## Enmiendas que provoca

Van en la propuesta de cambio de H30:

- **`CHANGELOG.md`**: el límite k(*n*) de este ADR, con la documentación del job.

Quedan fuera de las rutas del hito, para quien acepte este ADR:

- **ADR 0038**: su estado dice que este lo precisa en el número de respuestas de su regla (punto 3).
- **`CLAUDE.md`**: la decisión cerrada «Umbral de una clase del juez», con el límite de las respuestas medidas.
