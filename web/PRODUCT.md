# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- **Juristas y despachos pequeños (la audiencia de la web).** Abogados y asesorías de 1 a 3 personas (78,5 % del sector según el Libro Blanco CGAE 2026). Buscan «IA para abogados» y «Claude para abogados»; los frenan la formación y la fiabilidad, y el miedo concreto es la jurisprudencia inventada. La portada `/` es suya desde la #148; `/despachos` y `/ciudadania` redirigen a `/`.
- **Ciudadanía.** Personas con un asunto ante una administración: buscan su problema en primera persona, el plazo o el modelo de escrito, nunca el nombre de la ley. La portada las deriva a Ventanilla Legal (ventanillalegal.es, el proyecto cívico construido sobre kitlegal) y las páginas de `/consultas/` responden a sus preguntas con la norma citada. El objeto `ciudadania` de `src/data/audiencias.ts` ya no se publica.
- **Fuera de público:** administraciones públicas (no pueden usarlo por el Esquema Nacional de Seguridad) y perfiles técnicos o integradores. La web no se dirige a ellos.

Cuando las dos audiencias chocan en una decisión de diseño, ganan los despachos (decisión del 2026-11-04).

## Product Purpose

kitlegal son skills agénticas, apoyadas en un binario local, para consultar fuentes legales públicas españolas (BOE y otras) y llevar un asunto ante cualquier administración. La web https://kitlegal.es explica qué hace, demuestra el método con citas reales y lleva a instalarlo.

Una visita sale bien cuando la persona:

1. **Confía en el método**: entiende que nada se afirma sin una cita resuelta contra el BOE (`/como-se-evalua/`).
2. **Instala kitlegal** (`/instalar/`). La barrera conocida es la terminal.
3. **Escribe a `info@kitlegal.es`**, sobre todo un despacho, para contar lo que le falta: no hay telemetría, así que es la única forma de saberlo.

## Positioning

- Cada afirmación lleva una cita resuelta contra el BOE en el momento de leerla: identificador, bloque, fecha de consulta y hash. Sin cita no hay respuesta.
- Todo corre en el equipo de quien lo usa: software libre, gratuito, sin telemetría.
- Nunca inventa jurisprudencia ni cita una sentencia sin su documento delante. El CENDOJ no se consulta de forma automática: la persona trae el documento y kitlegal lo coteja.

## Operating Context

- Se usa dentro de un agente (Claude Code, Claude Desktop y otros hosts de skills o MCP local), no en la web: la web no ejecuta consultas.
- Páginas: la portada para juristas y despachos (`/`, con la comparativa de mercado, el bloque de Ventanilla Legal y la invitación a escribir a `info@kitlegal.es`), una página por consulta en `/consultas/<slug>/`, `/instalar/` paso a paso, `/como-se-evalua/`, `/bot/` (identificación del rastreador), `/schemas/`, `llms.txt`.
- La visibilidad se sigue en Search Console (`docs/SEO.md`).

## Capabilities and Constraints

- Astro 7, HTML estático, Tailwind 4, Fontsource; desplegado en GitHub Pages. Fuera de `make ci` (`make web`).
- CSP con `script-src 'none'`: ninguna interacción con JavaScript; lo interactivo va con CSS (radios, `<details>`).
- **Ninguna cita escrita a mano.** Los fragmentos de `src/data/citas.yaml` se comprueban al construir contra los sobres del binario (`make web-citas`); si el BOE trae otra versión del bloque, la construcción falla. Los textos de las páginas nombran citas por su id y nunca copian la norma.
- Cobertura en positivo: no prometer lo que no está cubierto (p. ej. recurrir multas de tráfico queda fuera) ni dar a entender que comprueba jurisprudencia.
- Lo que el usuario debe percibir como «actuar en su municipio» queda implícito, nunca como mensaje explícito.
- Ningún contenido del negocio ni del modelo cloud aparece en la web.

## Brand Commitments

- Nombre siempre en minúsculas: **kitlegal**. Dominio kitlegal.es; buzón info@kitlegal.es.
- Idioma: español de España, claro y sin jerga técnica para la ciudadanía; preciso y sin promesas para los despachos.
- Fotografías generadas con IA, avisado en el pie. Ninguna enseña texto legible, marcas ni logotipos de una institución; cada imagen se revisa una a una.
- Logo: `public/favicon.svg`, redibujado a mano; el SVG original está pendiente.

## Evidence on Hand

- Citas reales verificadas en `src/data/citas.yaml`; consultas en `src/data/consultas.ts`; textos de las portadas en `src/data/audiencias.ts`.
- Fotografías en `src/assets/ilustraciones/`.
- Método de evaluación público en `/como-se-evalua/`.
- **No hay** testimonios, clientes, casos de éxito ni cifras de uso (no se recogen datos de uso). No se inventan.

## Product Principles

1. **Nada sin cita.** Lo que la web afirma sobre la ley se puede comprobar, y se comprueba al construir.
2. **Honestidad sobre los límites.** Decir qué no hace y por qué pesa tanto como decir qué hace.
3. **El profesional decide.** El diseño sirve primero a quien firma y responde de lo que presenta.
4. **Local y sin rastreo.** La web no recoge nada; la conversación con el usuario es el correo.
5. **Quitar la barrera de instalar** antes que añadir mensajes.
