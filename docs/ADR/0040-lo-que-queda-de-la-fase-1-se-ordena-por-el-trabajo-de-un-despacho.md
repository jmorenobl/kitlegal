# 0040 · Lo que queda de la fase 1 se ordena por el trabajo de un despacho, y la fase 3 conserva su sitio

- **Estado**: aceptada (2026-10-19); sustituida en parte por el ADR 0044, que fija otro orden: la base legal
  genérica, lo administrativo sobre la base (con H8 y H28) y los despachos y asesorías (con H9.1, H27 y las
  verticales), y la fiscalización municipal detrás (puntos 1 a 4). Sigue vigente el punto 5.
- **Fecha**: 2026-10-19
- **Hito**: ninguno; reordena `docs/ROADMAP.md` detrás de H9. Da número a tres candidatos —H9.1, H27 y H28— y
  completa el ADR 0013 (el orden por tiempo hasta el uso) con una regla para que la fase 3 no se quede detrás del
  backlog. No toca el ADR 0027 (la fase 2 genérica), ni el 0036 (el CENDOJ), ni ningún contrato del binario.

## Contexto y problema

La web habla a dos audiencias (ADR 0034): la ciudadanía y los despachos. Las dos usan lo mismo hasta H9 —consultar y
citar la ley vigente, situar un municipio, contar un plazo administrativo—, y la fase 2 (el asunto y los escritos)
también les sirve a las dos. A partir de ahí piden cosas distintas. La ciudadanía, lo que hacen sus administraciones:
la fase 3 (contratos, subvenciones, boletines, ordenanzas, presupuestos). Un despacho, lo que hace falta para trabajar
un asunto en serio, y eso hoy no está en ninguna fase:

- **Los plazos procesales.** H9 cuenta el plazo administrativo y la interposición del contencioso. Un despacho
  pequeño, que es el despacho típico (`docs/USO.md`, 2026-10-01), litiga sobre todo en lo civil y en lo social, y ahí
  el cómputo es otro: la LEC hace inhábiles agosto y del 24 de diciembre al 6 de enero (art. 130.2) y cuenta el día
  del vencimiento hasta las veinticuatro horas (art. 133.1); la LRJS declara inhábiles los mismos días, salvo en las
  modalidades procesales que enumera (art. 43.4). Comprobado con `kitlegal boe articulo` el 2026-10-19. Es cálculo
  de fechas: va al binario (principio VIII).
- **La jurisprudencia que se puede traer sin la persona.** `jurisprudencia` coteja lo que la persona trae del CENDOJ
  (ADR 0036), pero todas las sentencias del Tribunal Constitucional y las del Supremo que anulan disposiciones se
  publican en el BOE, con reutilización autorizada, y un ECLI del Constitucional se resuelve contra el propio BOE
  (`docs/JURISPRUDENCIA.md`, §1). Esa vía está verificada desde el 2026-10-02 y sigue en el backlog.
- **La doctrina.** El buscador del BOE aloja tres bases de doctrina: los dictámenes del Consejo de Estado (desde 1987,
  con el texto completo en HTML), la doctrina de la Fiscalía General del Estado (circulares, instrucciones y consultas)
  y la de la Abogacía del Estado (comprobado en `boe.es/buscar/` el 2026-10-19). La ficha de ayuda del Consejo de
  Estado dice que el contenido es informativo y sin valor jurídico. Están sin revisar la vía de acceso para un
  programa (`robots.txt`, API de datos abiertos) y si el aviso de reutilización del BOE las cubre. La doctrina
  administrativa sectorial —DGT, TEAC, AEPD, TACRC, Consejo de Transparencia— está fuera del BOE, cada una con sus
  términos, y la de DGT y TEAC sigue pendiente de revisar (`refs/00-README.md`).

No hay todavía entradas de despachos en la bitácora que pidan cada una de estas piezas por su nombre: lo que hay es un
abogado al que la instalación en Windows se le atascó (2026-10-04), la búsqueda de los despachos y las multas por
jurisprudencia inventada (2026-10-01) y el análisis del 2026-10-19. El orden que sigue es, por tanto, una hipótesis de
uso, y la bitácora puede deshacerlo como cualquier otro (ADR 0013).

## Decisión

1. **Detrás de H9, la fase 1 sigue este orden: H20 → H9.1 → H8 → H27 → H28.**
   - H20 (la redacción vigente en una fecha) no cambia de sitio: es pequeña, H8 la necesita y su entrada ya está
     preparada.
   - **H9.1, los plazos procesales civiles y sociales**, entra detrás de H20: reutiliza `internal/core/plazos` y
     `data/festivos/` recién hechos, y es lo que un despacho cuenta a diario.
   - H8 (`cita validar`, también sobre textos ajenos) no cambia.
   - **H27, el Tribunal Constitucional por el BOE**: `jurisprudencia` resuelve y lee una sentencia del TC, y una del
     Supremo que anula una disposición, sin que la persona traiga el documento. Sale del candidato `tc` del backlog de
     fuentes.
   - **H28, la doctrina publicada en el BOE**, con una skill genérica `doctrina`. Va la última porque su vía de acceso
     está sin verificar: su primera tarea la verifica y, si un programa no puede buscar, el hito sigue el patrón de
     `jurisprudencia` (la persona busca y trae; el binario resuelve el identificador y coteja).
2. **La fase 1 pasa a decir «y las fuentes de un jurista»**: al terminar, además de lo que ya decía, un despacho
   cuenta un plazo civil o social con el binario y cita la jurisprudencia constitucional y la doctrina que publica el
   BOE con su texto delante. Las fases no se renumeran.
3. **La fase 3 conserva su sitio.** Lo que sirve a las dos audiencias (H20, H8, la fase 2) no se mueve. Al terminar la
   fase 2, H12 y H13 van por delante de cualquier candidato de despacho del backlog —la primera vertical, EUR-Lex y el
   TJUE, la doctrina fuera del BOE— salvo que una entrada de `docs/USO.md` pida otra cosa. Un despacho de derecho
   público también usa H13, y H13 es el corazón de la fiscalización.
4. **Los candidatos de despacho del backlog, en este orden cuando les toque**: la vertical fiscal (`boe-fiscal`, con
   la DGT y el TEAC si sus términos lo permiten), EUR-Lex con el TJUE y la doctrina sectorial (AEPD, TACRC, Consejo de
   Transparencia). Siguen sin número (ADR 0013).
5. **El DIR3 de los órganos que no son ayuntamientos sigue fuera**, como un candidato del backlog de fuentes con su
   condición de entrada: una vía abierta y sin credenciales, que se vuelve a buscar porque la revisión del issue #37
   es de septiembre. Hasta entonces, el órgano lo da la notificación (art. 40.2 LPAC; H10) y PLACSP trae el DIR3 del
   órgano de contratación (H13).

## Consecuencias

- El siguiente hito sigue siendo H20. Después vienen H9.1, H8, H27 y H28, y luego H10 y H11.
- La ciudadanía recibe en la fase 1 lo que ya tenía previsto (H20, H8); lo que se añade antes de la fase 2 son tres
  hitos que sobre todo usa un despacho. La fase 3 empieza más tarde, pero con su sitio fijado por delante del backlog
  de despachos.
- H27 y H28 añaden evals a `jurisprudencia` y crean las de `doctrina`: hay que recalcular el tope del job de evals
  antes de lanzarlos.
- Si la bitácora no confirma con uso real alguna de las tres piezas nuevas, se devuelve al backlog en la relectura
  del ritual (§6 del roadmap).
