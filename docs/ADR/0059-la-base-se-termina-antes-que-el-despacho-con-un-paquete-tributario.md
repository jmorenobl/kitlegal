# 0059 · La base se termina antes que el despacho, con un paquete tributario como segunda materia

- **Estado**: propuesta (2026-11-21); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-21
- **Hito**: ninguno; se escribe fuera de un run, como documentación. Precisa el ADR 0044 en el orden de la fase 1 y en
  el contenido de las fases 2 y 3 (punto 2), y en qué entra de la vertical fiscal antes de la fase 3 (punto 5). No
  cambia ningún principio, ningún contrato del binario ni la constitución.

## Contexto y problema

Con H33 fusionado (2026-11-20), la fase 1 tiene hechas las skills por capacidad, el molde del paquete de materia y la
cobertura en positivo, pero ninguno de sus cuatro puertos: fuentes (H34), documentos (H35), cómputo (H36) y
destinatarios (H37). Una revisión del roadmap (2026-11-21) encontró que la fase 1, tal como está, no deja la base
terminada:

- **El molde no tiene una segunda materia real.** H32 lo probó con una vertical de ensayo en `testdata/`. En
  `data/normas.yaml` solo hay normas con `vertical: administrativo`, y la única referencia de materia es
  `skills/plazos-y-territorio/references/administrativo.md`. La regla de las dos implementaciones reales (ADR 0044,
  punto 2) no se cumple para el paquete. Tampoco se ha probado nunca que un paquete saque sus filas de
  `regulados_aparte` (`docs/roadmap/backlog/verticales/README.md`).
- **La aceptación de la fase no la entrega ningún hito.** `docs/ROADMAP.md` pide evals de cinco o seis materias, y
  `evals/consulta-juridica/` tiene cuatro.
- **H36 se valida con el caso fácil.** Su segunda implementación es el art. 5 del Código Civil. Los ejes que no tiene el
  cómputo de la LPAC —agosto y del 24 de diciembre al 6 de enero inhábiles (art. 130.2 LEC), el vencimiento a las
  veinticuatro horas y los meses de fecha a fecha (art. 133 LEC)— llegan con H9.1, que reabriría el motor.
- **H37 no tiene un segundo destinatario.** Su propia entrada lo nombra como riesgo: «interfaces sin segundo uso».
- **El orden de la cola no es el de las dependencias.** H38 no depende de ninguno y va el quinto. El paquete que
  valida el molde no existe. Las condiciones de EUR-Lex, BORME, PETETE, DYCTEA y REGCON, de las que dependen H34 y
  las verticales, siguen sin revisar.

Lo que viene después de la base —los plazos procesales, la revisión de escritos, el Constitucional por el BOE, el
asunto y los escritos— es lo que un despacho usa a diario (ADR 0040; `docs/USO.md`, 2026-10-19). Está repartido entre
la fase 2 y la fase 3, y H9.1 y H27 van los últimos de la cola.

## Opciones consideradas

1. **Adelantar lo del despacho** (H9.1, H8 y H27) por delante de los puertos. Rechazada: es lo que el ADR 0044 deshizo
   del 0040. Cada uno de esos hitos se construiría sobre los motores de hoy y heredaría su forma.
2. **Mantener la cola y añadir solo el cierre de la fase.** Rechazada: el molde seguiría sin segunda materia real hasta
   la fase 3, y H36 y H37 se validarían con el caso que no exige nada al diseño.
3. **Terminar la base con segundas implementaciones que la exijan, una de ellas un paquete de materia real; y después,
   lo del despacho en entregas cortas, cada una usable al cerrar.** Elegida.

## Decisión

1. **La fase 1 termina la base, en este orden**: H38, H36, H42, H34, H35, H37 y H43.
   - **H38** va primero porque no depende de nada y es corto.
   - **H36** valida el motor con el cómputo civil de la LEC (arts. 130.2 y 133) como segunda implementación, con sus
     goldens. El art. 5 del Código Civil queda como opcional. La modalidad procesal de la LRJS como dato de entrada,
     los tipos de plazo procesales, la skill y sus evals siguen en H9.1. Los plazos de fecha fija, que no se cuentan
     desde un día inicial (el calendario del contribuyente), los declara fuera el ADR de H36.
   - **H42**, el paquete de la materia tributaria, va justo detrás de H36. Así los defectos del molde aparecen antes de
     diseñar los otros puertos, y H37 tiene su segundo caso real (punto 2).
   - **H34** y **H35** no cambian de alcance, salvo que H35 incluye el PDF con texto. Un PDF escaneado se declara no
     legible, sin OCR.
   - **H37** toma como segunda implementación del destinatario el órgano tributario de H42, si tiene un esquema de
     identificador abierto. Si no lo tiene, H37 se recorta al diseño del asunto y va junto a H10. El punto 5 del ADR
     0040 sigue vigente.
   - **H43** cierra la fase con las evals de varias materias que pide la aceptación de la fase 1.

2. **El segundo paquete de materia es el tributario, mínimo, y no la vertical fiscal.** Lo tributario ocupa hoy tres
   filas de `regulados_aparte` (DA 1ª.2.a, DA 1ª.2.c y art. 112.4 LPAC), así que es la única materia que prueba, además
   del molde, que un paquete saca lo suyo de lo que la ley regula aparte. Es también la primera vertical que el ADR
   0044 nombra (punto 5), y las asesorías la piden.
   - **Lo que trae H42**: sus normas con `vertical: fiscal`, una referencia en `plazos-y-territorio` y otra en
     `normativa`, sus filas de `cubierto` y los plazos de la revisión en vía administrativa como datos del motor de
     H36, con cada precepto leído antes de fijarlo.
   - **Lo que deja fuera**: la doctrina de la DGT y del TEAC, el calendario del contribuyente, los modelos, el
     sancionador tributario, las aduanas, las haciendas forales y los escritos.
   - **La vertical fiscal**, con todo eso, sigue en la fase 3 y entra cuando lo pida `docs/USO.md`. H42 es su paquete
     de materia, no la vertical.
   - Deja de valer, en `docs/ROADMAP.md` §5, «ni campo `vertical` en `data/normas.yaml` antes de la primera vertical»,
     superado desde H32.

3. **Las condiciones de las fuentes se revisan en una sesión interactiva, no en un run.** Son las de BORME, PETETE,
   DYCTEA y REGCON. Las de EUR-Lex van en H34, que ya lo pedía. Cada fila de `docs/SOURCES.md` la revisa una persona,
   y las sesiones de un run no tienen red. Se hace en paralelo a la fase 1 y no bloquea ninguno de sus hitos. Decide
   qué verticales pueden leer su doctrina por programa y cuáles siguen el patrón de `jurisprudencia`, en el que la
   persona trae el documento (ADR 0036).

4. **Fase 2, los despachos y lo administrativo, en entregas cortas.** Cada hito entrega algo que un despacho o una
   persona usa al cerrarlo, sin tocar la base. El orden es: H9.1, H8, H27, H10, H11, H15, H16 y H28.
   - **H9.1** solo añade el cómputo social, los tipos de plazo, la skill y las evals, porque el cómputo civil llega con
     H36.
   - **H8** revisa también un escrito en PDF, con lo que trae H35.
   - **H11** saca el escrito también en DOCX.
   - H9.1 y H27 pasan de la fase 3 a la 2.
   - **La fase 3** queda para las verticales —la fiscal completa primero—, en el orden de `docs/USO.md`, con los packs a
     partir de la segunda.

Se mantiene todo lo demás de los ADR 0040 y 0044. La genericidad por materia y por territorio, la cobertura en positivo,
las verticales como paquetes, el CENDOJ y la frontera humana siguen igual.

## Consecuencias

**A favor**

- Cada puerto se valida con el caso que exige al diseño, y lo de la fase 2 entra como datos sobre la base, sin
  reabrirla.
- El molde se prueba con una materia real que saca lo suyo de lo regulado aparte, antes de que lleguen las verticales.
- Desde que cierra la fase 1, cada hito es una entrega que un despacho puede usar.

**En contra, y asumido**

- La fase 1 gana dos hitos, H42 y H43, y H36 crece. Lo del despacho llega más tarde que con la cola de antes para H8,
  y antes para H9.1 y H27.
- H42 corre el riesgo de crecer hacia la vertical. Su entrada lleva la lista cerrada de lo que deja fuera.
- Etiquetar normas cambia `data/` y dispara las evals de las skills que las leen, entre ellas la 04 de `normativa`.
  H42 lleva su tope de sondeos en la entrada.
