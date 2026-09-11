# 0004 · Frontera humana: solo fuentes públicas, y lo que exige identidad termina en fichero listo para firmar

- **Estado**: aceptada
- **Fecha**: 2026-09-10
- **Hito**: H0

## Contexto y problema

Buena parte del trabajo jurídico no termina en una consulta, sino en un acto: presentar un escrito en
una sede electrónica, recoger una notificación, solicitar un aplazamiento, firmar con certificado. Un
agente que ya sabe leer el BOE, calcular un plazo y redactar el escrito está a un paso de presentarlo.

Ese paso, técnicamente, es un POST. Jurídicamente es otra cosa: quien presenta asume la
representación, la responsabilidad del contenido y las consecuencias de haberlo presentado en un
momento concreto, y lo hace con un certificado que identifica a una persona. Un error de un binario
—un plazo mal calculado, un documento equivocado, un reintento— no es un test rojo: es un acto
jurídico con efectos.

El problema es dónde para el automatismo, y que esa frontera esté decidida **antes** de que exista el
primer adaptador que pudiera cruzarla.

## Opciones consideradas

1. **Automatizar de punta a punta con el certificado de la persona usuaria.** Es lo que más valor
   aparente entrega y lo que hacen algunos productos del sector. Pone el certificado y la
   responsabilidad en manos de un proceso desatendido. Rechazada.
2. **Automatizar hasta la sede y pedir confirmación interactiva antes de enviar.** Suena a punto
   medio, pero no lo es: el envío lo sigue haciendo el binario, la responsabilidad sigue siendo suya
   y basta un fallo de lógica, un reintento o una confirmación mal interpretada por un agente para
   presentar algo indebido. Además, buena parte de esas sedes no autoriza el acceso automatizado.
3. **Frontera dura**: leer fuentes públicas y producir el artefacto, sin ejecutar nunca el acto.

## Decisión

Se adopta la **opción 3**, como invariante del proyecto:

- **Solo se automatizan fuentes públicas.** Consultar, descargar, indexar y citar lo que es público
  y cuyos términos de uso lo permiten.
- Toda acción que exija identidad —presentar un escrito, recoger una notificación, cualquier trámite
  con certificado— **termina en un fichero listo para firmar** y en el código de salida **6**
  (*requiere identidad humana*). El applet deja el documento y la instrucción de qué hacer con él.
- **Nunca un POST a una sede.** No es una preferencia de implementación: es la línea que separa lo
  que el proyecto hace de lo que no hará.
- Los códigos de salida hacen la frontera legible para un agente sin necesidad de interpretar texto:
  **5** es rate-limited o límite de términos de uso, **6** es «esto requiere una persona». Son
  distintos de los errores de operación (2, 3, 4) a propósito.
- Cada fuente declara su semáforo de automatizabilidad —🟢 API, 🟡 scraping permitido, 🔴 requiere
  identidad humana—; las 🔴 terminan en 6 por construcción.

## Consecuencias

**A favor**

- La responsabilidad queda donde debe: ningún acto jurídico lo ejecuta un binario, y el certificado
  no sale del control de su titular.
- Un agente puede distinguir sin ambigüedad «no he podido» de «esto no me corresponde», porque son
  códigos de salida distintos; eso hace que la frontera se respete también en los flujos automáticos
  que orquesta el propio agente.
- La regla es simple de auditar en revisión: si un adaptador abre una petición de escritura contra
  una sede, está mal, sin discusión caso por caso.
- Evita construir sobre accesos que las sedes no autorizan, con el riesgo de bloqueo que eso lleva.

**En contra, y asumido**

- El flujo nunca se cierra solo: siempre queda un paso humano, y el ahorro de tiempo es menor que el
  de un producto que sí presenta.
- Quedan fuera capacidades que otros ofrecen: presentación desatendida y vigilancia automática de
  buzones de notificaciones con identidad.
- Cada fuente nueva obliga a decidir y documentar su semáforo antes de escribir el adaptador, lo que
  añade trabajo previo a cada incorporación.
- La decisión no se relaja en un hito por comodidad ni por una funcionalidad concreta: cambiarla
  exigiría un ADR que sustituya a este y una enmienda de la constitución del proyecto.
