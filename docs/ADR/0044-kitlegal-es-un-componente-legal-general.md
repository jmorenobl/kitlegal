# 0044 · kitlegal es un componente legal general: la base genérica, lo administrativo y los profesionales, con las materias en el backlog

- **Estado**: aceptada (2026-10-22)
- **Fecha**: 2026-10-22
- **Hito**: H29, fuera de un run, como documentación. Sustituye al ADR 0027 en la frontera de la base (punto 3) y en el
  encuadre de la fase 3 (punto 7), y al ADR 0040 en el orden (puntos 1 a 4); precisa el ADR 0012 (qué es una
  vertical y qué es un pack); recoge el roadmap como directorio (workflow `hito` 2.10.0); constitución 2.14.0
  (principio II, principio X nuevo, el punto 4 del flujo de trabajo, las capas 1 y 3 y el punto 4 del criterio de
  decisión autónoma); workflow `hito` 2.11.0 (una fila del Constitution Check por cada principio, que comprueba
  `precheck.sh`). No cambia ningún contrato del binario.

## Contexto y problema

El 2026-10-21 una familia a la que una aseguradora renovó y cobró una póliza tras el fallecimiento de la asegurada usó
kitlegal para saber qué podía hacer (`docs/USO.md`, 2026-10-21). La norma se leyó y se citó bien: la Ley 50/1980, de
Contrato de Seguro. Le faltó todo lo que rodea a la norma —los plazos, a quién reclamar, qué vía seguir—, y nada le dijo
que faltaba. Al revisar por qué salieron cuatro cosas que no son de esa pregunta, sino de cómo está construido el kit:

- **Los motores son administrativos por construcción.** `plazos` tiene dos cómputos fijados en el código, `lpac` y
  `contencioso`, con el sábado y el domingo siempre inhábiles y los tipos de plazo enumerados en Go; `territorio` tiene
  por eje el municipio, la comunidad, el DIR3 del ayuntamiento y los boletines. El ADR 0012 dice que una vertical
  «añade criterio, no herramientas», pero con estos motores una materia nueva necesita código: el plazo general del
  Código Civil no se cuenta como el de la LPAC.
- **La cobertura se declara en negativo y respecto de la LPAC.** El ADR 0027 hizo de la base el procedimiento común y
  el acceso a la información, y de `data/cobertura.yaml` la lista de lo que la ley regula aparte de ese procedimiento.
  Una pregunta civil o de seguros no cae en ninguna fila, así que nada la declara no cubierta. La regla del principio
  IX para el territorio —fuera de lo cubierto, se declara— no tiene gemela para la materia.
- **El orden mira al despacho antes que a la base.** El ADR 0040 ordenó lo que quedaba de la fase 1 por el trabajo de
  un despacho (H9.1, H8, H27, H28) y puso la fiscalización municipal por delante de los candidatos de despacho. Cada uno
  de esos hitos se construiría sobre los motores administrativos y heredaría su forma.
- **Las skills y las evals no escalan a muchas materias.** Las skills no siguen la especificación Agent Skills —un
  cuerpo corto que guía y el detalle en referencias que se cargan cuando hacen falta—, y cualquier cambio mide las tres
  skills enteras en los dos modos.

Ese día el roadmap se reorganizó en base, administrativo y profesionales, y pasó a un fichero por hito
(`docs/ROADMAP.md`, «Punto de partida y rumbo»). Lo que lo gobierna —la constitución, los ADR 0012, 0027 y 0040, y
`CLAUDE.md`— seguía diciendo lo anterior, y es lo que lee el workflow: con «lo no especificado no se implementa» y la
lectura conservadora, un run de H30 a H38 habría construido la base dentro de la frontera administrativa.

## Opciones consideradas

1. **Mantener el marco y añadir materias como verticales sobre los motores de hoy.** Rechazada: cada materia pediría
   código (sus cómputos, sus destinatarios), contra el ADR 0012, y la cobertura en negativo seguiría sin declarar
   nada fuera del procedimiento común.
2. **Una fase por materia** (consumo y seguros, civil, familia…). Rechazada: no se sabe de antemano qué materia
   pedirá el uso, cada fase repetiría las mismas capacidades y el orden lo decide la bitácora (ADR 0013), no el
   roadmap.
3. **Primero una base genérica por capacidades; lo administrativo, como su primer paquete de materia; después, los
   profesionales; y las materias, al backlog como paquetes sobre la base.** Elegida.

## Decisión

1. **kitlegal es un componente legal general.** Ayuda a entender y resolver una situación legal de cualquier materia:
   a la ciudadanía, que quiere saber qué puede y qué no puede hacer antes de actuar, y a los profesionales, que
   necesitan más profundidad. Lo que hace —leer y citar la norma, contar un plazo, situar el territorio y el
   destinatario, cotejar el documento que trae la persona, preparar un escrito— lo hace por capacidades que no saben de
   ninguna materia; lo que es de una materia son datos y referencias.

2. **Tres fases y un backlog.** Sustituye al ADR 0040 en el orden (puntos 1 a 4):
   - **Fase 1, la base legal genérica** (H29 a H38): este ADR; las evals que solo miden lo que cambia (H30), primero,
     porque la fase reescribe las skills una a una; las skills según la especificación Agent Skills (H31); una entrada
     genérica, skills por capacidad y las materias como referencias (H32); la cobertura en positivo (H33); los puertos
     genéricos de fuentes, con EUR-Lex (H34), de documentos (H35), de cómputo (H36) y de destinatarios, territorio y
     asunto (H37); y las skills que se actualizan con el binario (H38), que no depende de ninguno y puede adelantarse.
     Cada abstracción nueva se valida con dos implementaciones reales.
   - **Fase 2, lo administrativo sobre la base**: lo que ya hay pasa a ser el primer paquete de materia, y sobre él
     H8, H10, H11, H15, H16 y H28. El orden interno de la fase 2 del ADR 0027 (H10 y H11 detrás de los plazos) se
     mantiene.
   - **Fase 3, los profesionales: despachos y asesorías**: H9.1, H27 y las verticales, en el orden que pida
     `docs/USO.md`.
   - **Transparencia y fiscalización municipal** (H12, H13, H14, H17 y H18) va detrás de la fase 3, con sus números y
     sus entradas, en el backlog: consulta lo que hacen las administraciones y no resuelve la situación de nadie. Era la
     fase 3 del ADR 0027 (punto 7), y deja de valer el punto 3 del ADR 0040, que la ponía por delante de los
     candidatos de despacho.
   - **El backlog**: las materias, una línea cada una, las verticales y los grupos de siempre. Una materia no es nunca
     una fase. La bitácora reprioriza como siempre (ADR 0013), y lo que el uso pida puede adelantar cualquier hito.

   Sigue vigente el punto 5 del ADR 0040: el DIR3 de los órganos que no son ayuntamientos queda fuera hasta que haya
   una vía abierta y sin credenciales; H37 lo declara no cubierto en la interfaz de destinatario.

3. **Genericidad por materia** (principio X de la constitución, gemelo del IX). Ninguna herramienta, dato versionado
   ni protocolo de skill lleva un caso especial para una materia. Lo que varía por materia —las normas, los cómputos
   con sus preceptos, la cobertura, los destinatarios, las referencias de cada skill de capacidad y las evals— se
   declara como datos y referencias, y una materia nueva entra sin tocar el código ni el protocolo de ninguna skill.
   La materia de validación es lo administrativo, como la Comunidad de Madrid con Leganés lo es del territorio: se
   implementa primero y protege con sus evals que lo de hoy sigue funcionando igual. Lo que hoy es administrativo por
   construcción —los dos cómputos y los tipos de plazo en Go, el eje municipal de `territorio`, la cobertura en
   negativo y la descripción de `legal-core`— lo llevan a la base H32, H33, H36 y H37. Hasta entonces no es un caso
   especial que un run de otro hito deba corregir fuera de su alcance, y ningún hito nuevo añade otro.

4. **La cobertura, en positivo.** Sustituye al ADR 0027 en la frontera de la base (punto 3). La base no es una
   materia ni un procedimiento: kitlegal cubre lo que declara cubierto, por capacidad y materia, y todo lo demás no lo
   cubre. Es la lectura conservadora: ante la duda, no está cubierto.
   - **Hoy** se declaran cubiertas la lectura y la cita de normas consolidadas del BOE, en cualquier materia y fecha;
     la preparación de la consulta de una sentencia y el cotejo de su documento; y, en lo administrativo, el
     procedimiento común de la LPAC y el acceso a la información de la LTAIBG ante cualquier administración, con los
     plazos de `plazos` y el territorio de `territorio`.
   - **Lo que la ley regula aparte del procedimiento común** —la DA 1ª.2 LPAC, las vías que sustituyen a la alzada
     (art. 112.2) y las reclamaciones económico-administrativas (art. 112.4)— sigue declarado no cubierto dentro de lo
     administrativo, con la norma que lo rige y sin los recursos ni los plazos del común, hasta que lo cubra su materia
     o su vertical. Se mantiene todo lo demás del punto 3 del ADR 0027: la forma fija de la declaración, que las evals
     comprueban sin modelo, y la lista en `data/`, con los preceptos que la fundan.
   - **Fuera de lo declarado**, la skill lo dice con la misma forma fija en todas las materias, cita la norma que haya
     leído y no da de memoria ningún plazo, órgano ni vía, ni aplica los de otra materia o los del procedimiento común.
     Lo que el agente busque con las herramientas genéricas para completar la respuesta lo marca como no verificado.
   - **H33** hace de `data/cobertura.yaml` la lista de lo cubierto, con su esquema, y da a todas las materias un único
     `⚠ NO CUBIERTO:` que dice también con qué herramienta busca el agente lo que falta. Hasta entonces la regla obliga
     a todo lo que se construya; que las skills de hoy no declaren no cubierta una pregunta civil es el hueco conocido
     que cierra H33 (`docs/USO.md`, 2026-10-21).

5. **Una vertical es un paquete de materia; un pack, un conjunto de verticales para un profesional.** Precisa el ADR
   0012, cuyo principio —base antes que vertical— no cambia:
   - Un **paquete de materia** trae normas con `vertical: <nombre>` en `data/normas.yaml`, un fichero de referencia por
     skill de capacidad, sus filas de cobertura y sus evals, y nada más. El molde lo fija H32, probado con una vertical
     de ensayo que entra sin tocar Go ni ningún `SKILL.md`.
   - Una **vertical** (fiscal, laboral, mercantil…) es una materia empaquetada para un tipo de profesional. No es una
     skill: es su paquete en las skills de capacidad. La vertical fiscal, con la migración de `boe-fiscal` y sus
     consultas actuales como evals, es la primera candidata; las verticales entran en la fase 3 en el orden de
     `docs/USO.md`.
   - Lo que un paquete necesita y la base no tiene —una fuente propia con sus condiciones revisadas, un cómputo propio,
     doctrina, escritos— es un hito de capacidad de la base, genérico y con dos implementaciones, no código de la
     vertical. Una materia o una vertical son un hito de paquete y uno por cada hueco de capacidad.
   - Un **pack** (`packs/<pack>/pack.yaml`) agrupa verticales para un tipo de profesional: una gestoría pide fiscal,
     laboral y mercantil juntas. El mecanismo entra con la segunda vertical (backlog · distribución), no con la
     primera.

6. **El roadmap es un directorio.** Cada hito es un fichero en `docs/roadmap/` y su directorio dice su estado:
   `done/`, `in-progress/` (como mucho uno), `todo/` (la cola, en el orden del prefijo `NNN-` del nombre) y `backlog/`
   (sin número ni especificación). El identificador `H<n>` no cambia nunca, porque lo citan los `specs/`, los ADR, el
   CHANGELOG, los commits y la serie de cierres; reordenar es renombrar el prefijo. La entrada de un hito —la que la
   persona escribe (ADR 0018) y el workflow lee— es su fichero. Lanzarlo es moverlo a `in-progress/` en `main`, y
   `scripts/hito.sh` solo lanza ese; el paso `hito_a_done` lo pasa a `done/` en el primer commit de la rama del hito, y
   un hito que no se lanza con el workflow, como este, se mueve a mano en su propuesta de cambio. La lista que admite
   el input `hito` la genera `make roadmap-sync`, y `make ci` comprueba el directorio entero (`TestRoadmap`).
   `docs/ROADMAP.md` conserva las fases, la Definition of Done, la arquitectura, los controles y el ritual.

7. **Lo que no se decide aquí.** Lo que H30 a H38 cambien de un ADR vigente lo decide cada uno con el suyo: cómo
   cuenta lo reutilizado en el umbral en tasa (H30), la regla de `references/` y el límite de líneas de `SKILL.md` del
   principio VIII (H31), los nombres de las skills por capacidad (H32), el conector de terceros (H34), el motor de
   cómputo (H36), el asunto genérico (H37) y el contrato del aviso de versión de H19 (H38). Este ADR fija el marco.

No se reabre ninguna decisión cerrada: skills primero (ADR 0008), la genericidad territorial (ADR 0009), el grafo en
tres piezas (ADR 0014), la frontera humana, el contrato de resultados (ADR 0023), sin telemetría, la fase 2 ante
cualquier administración y con los municipios en el asunto (ADR 0027, puntos 1, 2, 4 a 6 y 8 a 10) y el CENDOJ (ADR
0036) siguen igual.

## Consecuencias

**A favor**

- Ninguna regla vigente contradice el roadmap: los runs de H30 a H38 construyen la base genérica, no la
  administrativa.
- Una pregunta de cualquier materia tiene una respuesta honesta desde que H33 entra: la norma citada, lo que kitlegal
  cubre y lo que no, y lo que el agente buscó, marcado como tal.
- Una materia nueva cuesta un paquete de ficheros y sus evals; lo que pide de nuevo se queda en la base para todas las
  demás.
- Lo administrativo no se pierde: es el primer paquete y el de validación, y sus evals lo protegen mientras la fase 1
  reescribe las skills.

**En contra, y asumido**

- H9.1, H8, H27 y H28 llegan más tarde que con el ADR 0040, detrás de diez hitos de base, y la fiscalización municipal
  más tarde todavía. Lo que el uso pida los puede adelantar (ADR 0013).
- Hasta H33 la cobertura en positivo es una regla sin sus datos: las skills de hoy no declaran no cubierta una pregunta
  civil o de seguros.
- La regla de las dos implementaciones alarga cada hito de la base; es lo que evita una interfaz con un solo uso y una
  promesa.
- `refs/` sigue presentando `boe-fiscal` como una skill y la base como municipal. Son documentos semilla y no se
  reescriben; ante conflicto, prevalecen la constitución y el roadmap.
