# 0046 · Las skills siguen la especificación Agent Skills: un cuerpo corto, el detalle en referencias y la descripción en tercera persona

- **Estado**: aceptada (2026-10-23)
- **Fecha**: 2026-10-23
- **Hito**: H31, escrito antes del run, en la propuesta de cambio de su entrada, como el 0036 en H23: el run lo cumple y
  no lo decide. Sustituye en el principio VIII de la constitución el límite de menos de 300 líneas de `SKILL.md` y la
  regla de que todo `references/` se genera desde `data/` (ADR 0044, punto 7, deja las dos a H31). Extiende el juez
  con modelo del ADR 0037 a un texto que no es una respuesta: la descripción de una skill. Constitución 2.15.0. No
  cambia ningún contrato del binario.

## Contexto y problema

La especificación Agent Skills (agentskills.io, «Specification», consultada el 2026-10-23) carga cada skill en tres
niveles: el nombre y la descripción de todas desde el principio, el cuerpo de `SKILL.md` entero cuando se activa
—«< 5000 tokens recommended», en menos de 500 líneas— y los ficheros de `references/` solo cuando hacen falta, a un
nivel de `SKILL.md`. La guía de autoría de Anthropic («Skill authoring best practices», misma fecha) añade la
descripción en tercera persona —«The description is injected into the system prompt, and inconsistent point-of-view
can cause discovery problems»—, el índice en las referencias de más de 100 líneas y las referencias enlazadas
directamente desde `SKILL.md`.

Las skills de kitlegal no siguen nada de eso, por cómo está escrita la regla del principio VIII. Todo `references/`
se genera desde `data/`, así que el protocolo no tiene otro sitio que `SKILL.md`, y el único límite es el de menos de
300 líneas, que se cumple con líneas de 120 columnas. Medido el 2026-10-23 con `claude-sonnet-5-5` (la diferencia de
tokens de entrada de una sesión con el cuerpo y sin él), el español de las tres skills da 2,26 caracteres por token,
y los cuerpos tienen unos 10 760 tokens (`boe-legislacion`), 8520 (`legal-core`) y 6000 (`jurisprudencia`). Las tres
descripciones le hablan al agente en segunda persona: «Úsala», «Actívala», «también cuando creas conocer la
respuesta». La de `boe-legislacion` tiene 1023 caracteres de los 1024 que admite la especificación.

Que una descripción esté en tercera persona es una cuestión de significado, no de forma. En español un verbo como
«Consulta», «Lee» o «Calcula» es a la vez la tercera persona que dice lo que hace la skill y el imperativo de tú; y la
primera o la segunda persona de una pregunta citada como ejemplo («¿hasta cuándo puedo recurrir?») es correcta. Una
lista de palabras marcaría lo correcto o dejaría pasar lo que no lo es, y el ADR 0037 ya decidió que el significado
lo juzga un modelo medido.

## Opciones consideradas

1. **Mantener las 300 líneas y las referencias generadas.** Rechazada: es lo que deja el protocolo entero en el
   cuerpo, por encima del presupuesto de la especificación.
2. **Un presupuesto de líneas.** Rechazada: las 500 líneas de la especificación ya se cumplen hoy; lo que carga el
   contexto son los tokens.
3. **Un presupuesto en tokens sin red.** Rechazada: no hay tokenizador sin red, y la suscripción no da la API de
   recuento.
4. **Un presupuesto en caracteres derivado de una razón medida, y un juez con modelo para la tercera persona.**
   Elegida.

## Decisión

1. **El reparto.** El cuerpo de `SKILL.md` lleva lo que toda respuesta de la skill necesita: el protocolo, las reglas
   invariantes, las formas fijas que las evals comprueban sin modelo (las líneas `⚠`), la tabla de comandos y, por
   cada referencia, cuándo leerla. Una referencia lleva el detalle que solo pide una clase de pregunta. El protocolo
   de una referencia se escribe a mano, sin cabecera; lo que sale de `data/*.yaml` sigue generado, con su cabecera
   «generado desde» y declarado en `kitlegal-referencias`, y `data/` sigue siendo su única fuente de verdad.
   `references/` es plano: ficheros `.md`, sin subdirectorios, cada uno enlazado desde `SKILL.md` por su ruta
   relativa y ninguno enlazado desde otra referencia; el de más de 100 líneas abre con un índice de sus apartados.
2. **Los presupuestos.** El cuerpo de `SKILL.md` —sin el frontmatter y con la tabla de comandos generada— tiene
   menos de 500 líneas y como mucho 11 000 caracteres: con 2,26 caracteres por token, unos 4870 tokens, por debajo de
   los 5000 de la especificación. La descripción tiene como mucho 900 caracteres, el 88 % de los 1024. Los
   caracteres se cuentan como runas. La razón se vuelve a medir, y la cifra se revisa con una propuesta de cambio,
   cuando cambie el modelo que decide (ADR 0031).
3. **La validación de la especificación**, en Go y sin red, en `make ci`: la de `ValidarFrontmatter` y, además,
   ninguna etiqueta XML en `name` ni en `description` y ni `anthropic` ni `claude` en `name`, como pide la guía de
   Anthropic. `skills-ref` no entra: es Python, y `make ci` no lo usa.
4. **La descripción en tercera persona la juzga un modelo.** Es una clase del juez con modelo del ADR 0037,
   `habla_en_primera_o_segunda_persona`, con una pregunta cerrada sobre la descripción de cada skill. Igual que en el
   ADR 0037: el modelo del juez y la versión de Claude Code de sus votos son los fijados para el juez de las evals; la
   frase citada se comprueba sin modelo; marca con tres votos unánimes, pedidos por orden; y solo decide medido contra
   casos etiquetados, con la medida versionada y comprobada sin modelo en cada ejecución. Su umbral es 0: una
   descripción es un texto fijo, no una muestra. Corre en el job de evals cuando la elección del ADR 0045 incluye el
   cambio de una descripción, y reutiliza por huella el juicio de una descripción que no ha cambiado; nunca en
   `make ci` ni en un paso de un run (ADR 0032). La rúbrica, los casos y la medida están en `evidencias/adr-0046/`.
5. **Los controles de contenido leen la skill entera.** Lo que hoy se comprueba sobre `SKILL.md` —la prosa sin
   expresiones prohibidas, la línea sin consulta, las órdenes para PowerShell, las formas fechadas— se comprueba sobre
   `SKILL.md` y sus referencias escritas a mano: mover un texto a una referencia no lo saca de ningún control. Las
   órdenes de ejemplo de una referencia son de lo que la caché de las sesiones sirve sin red (ADR 0043).
6. **La guía de autoría** del proyecto (`docs/SKILLS.md`) recoge este reparto, los presupuestos con su porqué y la
   forma de la descripción, con las dos fuentes y su fecha de consulta.

## Validación del juez

Medido el 2026-10-23, antes del run, con `claude-opus-5-5` y Claude Code 2.1.289, los fijados para el juez
(`evidencias/adr-0046/LEEME.md`). 99 casos: las ocho descripciones que han tenido las skills del repositorio, todas
con «Úsala», y 91 derivados por construcción, con la etiqueta puesta antes del primer voto. Resultado, en una
pasada de 225 votos sin nulos ni errores: 63 de 63 defectos marcados con los tres votos y 0 de 36 correctos, ni con
un voto. Entre los correctos están los dos casos que una lista no distingue: la primera y la segunda persona dentro
de una pregunta citada y el verbo que se escribe como un imperativo con la skill de sujeto. Ningún defecto sutil es
natural, y la medida no se hizo con el código del job: una persona la lanza con la etiqueta `evals-medir-juez` sobre
la propuesta de H31, antes de fusionarla.

## Consecuencias

**A favor**

- Lo que el agente carga al activar una skill baja a menos de la mitad en `boe-legislacion`, y una materia nueva
  (ADR 0044) puede entrar como referencia sin agrandar el cuerpo.
- La descripción la juzga lo que entiende el español: un verbo de tercera persona o una pregunta citada no la tumban,
  y una segunda persona escondida no pasa.

**En contra, y asumido**

- Reescribir el cuerpo cambia lo que lee el modelo, y la deciden las evals: con las tres descripciones cambiadas, H30
  elige las 45 en los dos modos.
- La serie de cierres de cada skill empieza de cero, porque su clave es el árbol de `skills/<skill>` (ADR 0038).
- Lo que va a una referencia el modelo puede no leerlo. Por eso las formas fijas y las reglas que toda respuesta
  necesita se quedan en el cuerpo.
- La regla de tercera persona es más estricta que los ejemplos de la guía de Anthropic, que escriben «Use when…»: en
  español «Úsala» es un imperativo de tú, y la entrada de H31 pide la tercera persona.
- El presupuesto en caracteres es una aproximación: la razón depende del texto, y se mide de nuevo con el modelo.
