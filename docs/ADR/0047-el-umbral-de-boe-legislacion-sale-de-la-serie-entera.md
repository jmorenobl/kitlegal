# 0047 · El umbral de `afirma_lo_no_leido` en `boe-legislacion` sale de la serie entera: como mucho 4 de 120

- **Estado**: propuesta (2026-10-25); queda aceptada cuando una persona fusiona su propuesta de cambio.
- **Fecha**: 2026-10-25
- **Hito**: transversal, después de las fusiones de la deuda tras H31 (#192 a #196) y antes de lanzar H32. Sube un
  umbral, que el ADR 0038 (punto 4) y el ADR 0045 (punto 6) reservan a otro ADR. Mantiene el criterio del ADR 0038,
  que una ejecución salga en rojo sin motivo como mucho 1 de cada 50 veces, y la regla del límite del ADR 0045. No
  toca `jurisprudencia`, el juez, la rúbrica ni el modelo que decide.

## Contexto y problema

El umbral de `afirma_lo_no_leido:<modelo>` en `boe-legislacion` es `0.028` en `evals/boe-legislacion/juez/clases.yaml`:
con las 120 respuestas de una ejecución completa, como mucho 3 marcadas. El ADR 0045 deriva de él la tasa que tolera,
*p*ref = 0,85 %: con ella, una ejecución completa sale en rojo sin motivo 1 de cada 50 veces.

**La skill falla más que eso, y siempre en el mismo sitio.** La serie de cierres (`scripts/workflow/serie.sh mostrar
boe-legislacion`, 2026-10-25) suma 19 marcadas de 1.664 respuestas, de H24-1 a H31-1: el 1,14 %. Las marcas son casi
siempre la materia de un precepto que la respuesta nombra y no leyó, en las evals por materia: «el art. 7, que regula
las rentas exentas» en las del IRPF, «el art. 27 sobre tipicidad» en la del principio de legalidad. Con el 1,14 %, una
ejecución completa sale en rojo sin motivo 1 de cada 20 veces, y un cierre, que mide hasta tres, lo nota más.

**Ninguna versión de la skill ha cambiado esa tasa.** Por versión:

| Árbol de la skill | Marcadas | Respuestas | Tasa |
|---|---|---|---|
| `6d65aa3` | 1 | 108 | 0,93 % |
| `75e1d04` | 3 | 108 | 2,78 % |
| `6063317` | 3 | 428 | 0,70 % |
| `eed129f` | 8 | 540 | 1,48 % |
| `49b59ac` | 2 | 120 | 1,67 % |
| `27aeccc` | 1 | 240 | 0,42 % |
| `92525ec` | 1 | 120 | 0,83 % |
| **Todas** | **19** | **1.664** | **1,14 %** |

Las diferencias entre versiones caben en el azar (χ² de homogeneidad = 5,41 con 6 grados de libertad, p ≈ 0,5).
Tampoco lo cambiaron los arreglos probados fuera de la serie (2026-10-11 y 2026-10-12; `evidencias/adr-0038/`): la
rúbrica de cada artículo en el índice, Opus en lugar de Sonnet y un paso de repaso antes de responder.

**La regla del ADR 0038 no llega a aplicarse.** Su punto 3 toma la tasa acumulada «con la skill como está, sobre 500
respuestas al menos». La skill cambia en casi cada hito que la toca y cada cambio empieza la cuenta de cero, así que la
versión vigente casi nunca tiene 500 respuestas. Solo `eed129f` las tuvo. La regla no da escalón, y el umbral se queda
en el que se fijó con 0,9 % de tasa supuesta.

**El caso que lo destapa.** La #195 (`golang.org/x/text` 0.43.0) midió la suite completa porque toca `go.mod`, y
`boe-legislacion` dio 5 de 120: cuatro respuestas sobre el art. 7 de la Ley del IRPF y una sobre el art. 17. El cambio
no toca la skill ni lo que el binario devuelve. Se fusionó en rojo, con un comentario que lo explica, por decisión de
quien mantiene el proyecto. Las dos medidas siguientes del mismo día, de la #193 y la #196, dieron 3 y 2 de 120.

## Opciones consideradas

1. **Volver a medir hasta que salga en verde.** Rechazada. Es lo que el ADR 0038 quitó, cuesta una ejecución completa
   cada vez, y un verde así no dice nada de la skill.
2. **Dejar el umbral y fusionar en rojo cuando el rojo se explique.** Rechazada como regla. Un rojo que se explica a
   mano en cada propuesta deja de ser un control (ADR 0029), y la explicación la escribe quien quiere fusionar.
3. **Arreglar antes la skill.** No se decide aquí. Tres arreglos probados no bajaron la tasa, y H32 reorganiza las
   skills por capacidad: lo que se arregle hoy en `boe-legislacion` se reescribe en el hito siguiente. Sigue siendo
   trabajo normal (ADR 0038, punto 8), que se mide con la serie.
4. **La tasa de la única versión con 500 respuestas, `eed129f` (1,48 %), que daría 5 de 120.** Rechazada. Es una de
   siete versiones que no se distinguen, dejó de ser la vigente hace tres versiones, y elegirla sería elegir la más
   alta.
5. **La serie más las medidas de las propuestas del 2026-10-25 (29 de 2.024, 1,43 %), que daría 5 de 120.**
   Rechazada. La serie cuenta los cierres, una vez cada respuesta (ADR 0038, punto 6; ADR 0045, punto 5). Añadirle las
   medidas de propuestas que se conocen por haber salido en rojo es elegir los datos por su resultado.
6. **La serie entera de la skill, 1,14 %, con la regla del ADR 0038: como mucho 4 de 120.** Elegida.

## Decisión

1. **`afirma_lo_no_leido` en `boe-legislacion` lleva `umbral: 0.034`.** Con las 120 respuestas de una ejecución
   completa son como mucho 4 marcadas: 4/120 cabe en 0,034 y 5/120 no. Con 108 siguen siendo 3. La regla del ADR 0045
   deriva de él *p*ref = 1,28 %, por encima del 1,14 % de la serie: a la tasa de la skill, una ejecución completa sale
   en rojo sin motivo un 1,2 % de las veces.
2. **La tasa de la que sale es la de la serie entera**, de todas las versiones, porque ninguna se distingue de las
   demás. Es una lectura de las cifras de hoy, no una regla nueva. El punto 3 del ADR 0038 sigue como está para
   estrechar.
3. **El umbral sigue sin subir solo** (ADR 0038, punto 4). Lo estrecha una persona cuando la serie lo permita. Subirlo
   otra vez pediría otro ADR, con sus cifras.
4. **El defecto sigue siendo trabajo de la skill.** Se mide con la serie después de H32, sobre las skills ya
   reorganizadas. Los candidatos sin probar siguen siendo dos: leer un precepto antes de nombrarlo y una forma fija
   para lo que la respuesta no ha leído.

## Validación

Probabilidad de que una ejecución completa, 120 respuestas, salga en rojo según la tasa real de la skill:

| Tasa real | Con 3 de 120 (0.028) | Con 4 de 120 (0.034) |
|---|---|---|
| 0,85 % (*p*ref de 0.028) | 2,0 % | 0,4 % |
| **1,14 % (la serie)** | **4,9 %** | **1,2 %** |
| 1,43 % (serie y propuestas de hoy) | 9,4 % | 3,0 % |
| 3 % | 48,7 % | 29,2 % |
| 5,6 % (la de H22) | 90,9 % | 80,8 % |

Con 4 de 120 vuelve el criterio de 1 de cada 50 a la tasa medida. Una vuelta a la tasa de H22 se sigue viendo 4 de
cada 5 veces, y que dos ejecuciones seguidas no la vean pasa 1 de cada 27.

La #195 sigue en rojo con este umbral: 5 de 120 no cabe en 4. Este ADR no la absuelve. A la tasa de la serie, 5 o más
salen un 1,2 % de las veces.

`TestLimiteDeLaTasa`, en `make ci`, da con el `clases.yaml` nuevo: `boe-legislacion, afirma_lo_no_leido (umbral
0.034): pref = 1,28 %, N = 120, k(N) = 4`. Con *n* respuestas medidas, el límite es k(*n*) del ADR 0045 con esa
*p*ref: con 60, k(60) = 3 y el informe publica 3/60.

## Consecuencias

**A favor**

- El rojo de `boe-legislacion` vuelve a decir algo: sale cuando la skill empeora, no 1 de cada 20 ejecuciones por un
  defecto que ya se conoce.
- Un hito que no toca la skill deja de gastar sus reparaciones del cierre en ella. Es lo que pasó con H25.
- El umbral queda escrito con la tasa de la que sale, y la serie dirá si se equivoca.

**En contra**

- Se detecta menos una regresión pequeña. Una skill que pasara del 1,1 % al 3 % se vería el 29 % de las veces en lugar
  del 49 %. Una grande, como la de H22, casi igual que antes.
- Da por buena una skill que falla en el 1 % de sus respuestas por materia. Es la tasa que mide desde H24, ahora
  escrita en el umbral en lugar de esconderse detrás de un rojo a ratos.

## Enmiendas que provoca

- `evals/boe-legislacion/juez/clases.yaml`: `umbral: 0.034` y su comentario.
- Los tests que fijan a mano el valor del repositorio, para que otro no pase: `TestEvalsDelRepositorio`,
  `TestUmbralConjunto`, `TestUmbralConjuntoConMenosRespuestas`, `TestUmbralesDeBoeLegislacion` y
  `TestUmbralesDeLasTresSkills`. En `TestUmbralConjunto`, el caso de 104 respuestas publica ahora el umbral de la clase,
  porque con 0,034 la proporción ya da k(104) = 3. La rama que publica k(*n*)/*n* la sigue fijando
  `TestUmbralConjuntoConMenosRespuestas`.
- `CONTRIBUTING.md`: las cifras de `boe-legislacion`.
- `CLAUDE.md`: la línea del umbral cita este ADR.
- Los estados de los ADR 0038 y 0045 dicen que este los precisa.
- `docs/USO.md`: la entrada del 2026-10-25.
- `CHANGELOG.md`: la entrada bajo *Unreleased*.
