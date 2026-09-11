# 0003 · De CENDOJ, solo ECLI y metadatos: ni ingesta masiva ni scraping del buscador

- **Estado**: aceptada
- **Fecha**: 2026-09-10
- **Hito**: H0

## Contexto y problema

La jurisprudencia del Tribunal Supremo, la Audiencia Nacional, los TSJ y las Audiencias Provinciales
se consulta en España a través del buscador del CENDOJ (CGPJ). Es la fuente natural para cualquier
herramienta jurídica, y a la vez la única del mapa de fuentes de `kitlegal` marcada en rojo por un
motivo que no es técnico: es una web con CAPTCHA cuyos términos de uso **prohíben expresamente el uso
masivo o automatizado**.

`kitlegal` necesita, aun así, poder citar sentencias: una respuesta sin fuente, URL, fecha de consulta
y hash no es una cita, y una skill de jurisprudencia que no pueda identificar la resolución que cita
no sirve para nada. El problema es cómo dar esa capacidad sin apoyarse en un acceso que la fuente no
autoriza.

## Opciones consideradas

1. **Scrapear el buscador del CENDOJ.** Es lo que hace media herramienta del sector. Infringe los
   términos de uso y exige sortear el CAPTCHA, es decir, exactamente la clase de evasión que el
   proyecto no hace. Rechazada sin matices.
2. **Descarga masiva y espejo local de las resoluciones.** Al problema anterior le suma el de la
   redistribución de un fondo documental que no es nuestro. Rechazada.
3. **Resolver identificadores y metadatos por vías legítimas, y dejar el texto íntegro fuera del
   automatismo.**
4. **No ofrecer nada de jurisprudencia.** Coherente pero innecesariamente pobre: hay caminos
   legítimos que cubren la parte que de verdad importa para citar.

## Decisión

Se adopta la **opción 3**. De CENDOJ solo se resuelven **ECLI y metadatos**; nunca ingesta masiva ni
scraping del buscador. En concreto:

- El applet `ecli resolver` obtiene la URL canónica y los metadatos de una resolución (fecha, órgano,
  sala, ponente) a través del **resolutor ECLI europeo** (portal e-Justice), no del buscador.
- Los ECLI entran en el sistema por vías públicas: referencias en el BOE, notas de prensa y
  resoluciones publicadas del CGPJ, o el propio identificador que aporte la persona usuaria.
- El Tribunal Constitucional se trata aparte, con sus propios términos revisados: buscador HJ y
  sección TC del sumario del BOE.
- En el grafo, una `Resolucion` entra con ECLI, metadatos y URL. **Nunca con el texto íntegro
  descargado automáticamente.**
- Nunca se resume ni se caracteriza una sentencia que no se ha leído: sin texto, la respuesta es la
  cita y los metadatos, no una síntesis inventada.
- Para búsqueda a texto completo se usan bases de datos con licencia: el agente prepara la consulta y
  la ejecuta una persona, o la persona pega el texto de la resolución.

Esta decisión es previa a que exista el primer adaptador de fuentes, y por eso se registra en H0:
llega antes que el código que podría infringirla.

## Consecuencias

**A favor**

- El proyecto no depende de un acceso prohibido ni de rutinas de evasión de CAPTCHA, que además son
  frágiles y se rompen sin aviso.
- Las citas siguen siendo resolubles y verificables: identificador, URL canónica, fecha y órgano.
- La frontera queda escrita y auditable en `docs/SOURCES.md`, fuente a fuente, con sus términos y la
  fecha en que se revisaron.

**En contra, y asumido**

- La cobertura de jurisprudencia es menor que la de un producto comercial que sí licencia el fondo.
  Es una limitación de producto, no un defecto de implementación, y se comunica como tal.
- La búsqueda a texto completo de sentencias queda fuera del alcance automatizado.
- Cualquier skill que necesite el texto íntegro termina en un paso manual, coherente con la frontera
  humana del [ADR 0004](0004-frontera-humana.md).
- La decisión no se reabre por conveniencia ni por presión de una funcionalidad concreta. Solo la
  revisaría un cambio en los términos de uso o la aparición de una API pública del CGPJ, y entonces
  con un ADR nuevo que sustituya a este, no con una excepción puntual en el código.
