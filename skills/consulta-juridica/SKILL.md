---
name: consulta-juridica
description: >-
  Punto de entrada de cualquier pregunta sobre una situación legal en España, de cualquier materia —un seguro, un
  contrato, una multa, un recurso contra una administración, una herencia, una sentencia o un artículo de una ley—:
  dice qué tipo de pregunta es, qué capacidad de kitlegal la resuelve —normativa, jurisprudencia, plazos y
  territorio— y con qué skill, qué cubre la base y qué no, y lo que no cubre lo dice sin dar de memoria ningún plazo,
  órgano ni vía. Se usa al empezar cualquier pregunta legal, antes que las demás skills de kitlegal, también cuando el
  agente cree saber la respuesta, cuando la pregunta no nombra ninguna norma o cuando pide qué hacer, a quién reclamar
  o qué escrito presentar.
---

# Consulta jurídica: qué la resuelve y qué cubre la base

Esta skill no consulta nada ni responde por su cuenta: decide qué capacidad de kitlegal resuelve la pregunta, con qué
skill, y qué cubre la base. Lo que responde cada capacidad —su protocolo, sus reglas, sus líneas `⚠` y sus órdenes de
`kitlegal`, desde el `PATH` o con su herramienta— lo lleva la skill de esa capacidad: actívala y síguela.

## Método

### 1. Qué tipo de pregunta es

Qué pide la persona, con sus palabras: lo que dice o fija una norma —un artículo, un requisito, dónde se regula algo—,
hoy o en una fecha; una sentencia o lo que dice la jurisprudencia; hasta cuándo puede hacer algo —un recurso, un
plazo, un silencio—; el territorio de un municipio; o qué hacer ante una situación —a quién reclamar, qué vía, qué
escrito—. Una pregunta puede pedir varias cosas: cada una va a su capacidad.

### 2. Qué capacidad la resuelve

| Lo que pide | Capacidad | Skill |
|---|---|---|
| Lo que dice o fija una norma, hoy o en una fecha | normativa | `normativa` |
| Una sentencia, si existe o qué dice, o la jurisprudencia sobre algo | jurisprudencia | `jurisprudencia` |
| Hasta cuándo se recurre, cuándo vence un plazo o se produce el silencio | plazos | `plazos-y-territorio` |
| La comunidad, la provincia, los boletines o el DIR3 de un municipio | territorio | `plazos-y-territorio` |

Activa la skill de cada capacidad que pida la pregunta antes de pedir nada a kitlegal —ninguna orden ni herramienta
suya sin su skill activada, aunque la herramienta o la forma de la cita parezcan bastar— y responde con su protocolo:
sus formas y sus citas son las de la respuesta.

### 3. Qué cubre la base

- **Normativa**: leer y citar cualquier norma consolidada del BOE, estatal o autonómica, de cualquier materia y en
  cualquier fecha.
- **Jurisprudencia**: preparar la consulta exacta de una sentencia, que la persona busca con su navegador, y cotejar
  el documento que trae. No consulta ningún buscador ni dice si una sentencia existe.
- **Plazos**: solo los del procedimiento administrativo común y la interposición del recurso contencioso-
  administrativo, con los festivos de los municipios que cuentan. Lo que la ley regula aparte de ese procedimiento lo
  declara `plazos-y-territorio` con su línea.
- **Territorio**: la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un municipio.
- **Paquetes de materia**: lo que una materia trae a las skills de capacidad —sus normas, una referencia en cada skill
  que la necesita y lo que la ley regula aparte—. Los de hoy, con las skills que los llevan:
<!-- paquetes de la base: lo genera make skills-sync -->
- lo administrativo: `plazos-y-territorio`.
<!-- fin de los paquetes de la base -->
- **Sin herramienta, no cubiertos**: los documentos —escritos, recursos, reclamaciones, solicitudes— y los
  destinatarios —a quién se reclama, qué órgano resuelve, ante quién se presenta—.

### 4. Lo que no cubre, en prosa

Si la pregunta pide algo que la base no cubre —el plazo de una materia que no es la del procedimiento administrativo
común, a quién reclamar, qué vía seguir, un escrito—, dilo en la respuesta con tus palabras: qué parte no está
cubierta. Apóyate solo en la norma que hayas leído con kitlegal en esta conversación, citada como la cita la skill de
normativa, `<forma legible> [<identificador>, bloque <id>]`. No des de memoria ningún plazo, órgano ni vía, ni
apliques los de otra materia o los del procedimiento común, ni pidas `kitlegal plazos calcular` para un asunto de
otra materia: ese plazo no es el suyo.

Por qué: quien pregunta va a actuar con lo que lee; un plazo o un órgano dicho de memoria puede ser de otra materia o
haber cambiado, y lo que la base no cubre nadie lo ha comprobado aquí. Decirlo le deja buscarlo donde sí está.
