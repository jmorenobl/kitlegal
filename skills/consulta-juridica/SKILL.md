---
name: consulta-juridica
description: >-
  Punto de entrada de toda pregunta legal en España, de cualquier materia: dice qué capacidad de kitlegal la resuelve
  —normativa, jurisprudencia, plazos y territorio— y con qué skill, y lo que no cubre lo dice sin dar de memoria plazos,
  órganos ni vías. Se usa al empezar toda pregunta legal, antes que las demás skills de kitlegal, también si no nombra
  norma o pide qué hacer, a quién reclamar o qué escrito presentar.
---

# Consulta jurídica: qué la resuelve y qué cubre la base

Esta skill no consulta nada: decide qué capacidad de kitlegal resuelve la pregunta, con qué skill, y qué cubre la
base, y declara lo que no cubre con su línea (paso 4). Lo que responde cada capacidad —su protocolo, sus reglas,
sus demás líneas `⚠` y sus órdenes de `kitlegal`, desde el `PATH` o con su herramienta— lo lleva la skill de esa
capacidad: actívala y síguela.

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

<!-- lo que cubre la base: lo genera make skills-sync -->
- **Normativa**: leer y citar cualquier norma consolidada del BOE, estatal o autonómica, de cualquier materia y en cualquier fecha.
- **Jurisprudencia**: preparar la consulta exacta de una sentencia, que la persona busca con su navegador, y cotejar el documento que trae. No consulta ningún buscador ni dice si una sentencia existe.
- **Plazos**: solo los del procedimiento administrativo común y la interposición del recurso contencioso-administrativo, con los festivos de los municipios que cuentan. Lo que la ley regula aparte de ese procedimiento lo declara `plazos-y-territorio` con su línea.
- **Territorio**: la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un municipio.
<!-- fin de lo que cubre la base -->
- **Paquetes de materia**: lo que una materia trae a las skills de capacidad —sus normas, una referencia en cada skill
  que la necesita, lo que cubre y lo que la ley regula aparte—. Los de hoy, con las skills que los llevan:
<!-- paquetes de la base: lo genera make skills-sync -->
- lo administrativo: `plazos-y-territorio`.
<!-- fin de los paquetes de la base -->
- **Sin herramienta, no cubiertos**: los documentos —escritos, recursos, reclamaciones, solicitudes— y los
  destinatarios —a quién se reclama, qué órgano resuelve, ante quién se presenta—.

### 4. Lo que no cubre, con su línea

Toda parte de la pregunta que no resuelve ninguna capacidad de la lista del 3 —el plazo de una materia que no es la del
procedimiento administrativo común, a quién reclamar, qué vía seguir, un escrito— lleva su línea, con el motivo «no está
en lo que kitlegal cubre». Esta línea no es para lo que la ley regula aparte del procedimiento común: eso lo declara,
con su propia línea, la skill de la capacidad que lo pide, y tras ella no leas su norma. Lo que leas con la skill de
normativa para completarla va detrás, citado, con su marca: el texto citado es el de la norma; lo que no está verificado
es que se aplique a este asunto. No des de memoria ningún plazo, órgano ni vía, ni apliques los de otra materia o los
del procedimiento común, ni pidas `kitlegal plazos calcular` para un asunto de otra materia: ese plazo no es el suyo.

```text
⚠ NO CUBIERTO: <asunto>: <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos.
⚠ NO VERIFICADO: <lo encontrado>: <forma legible> [<identificador>, bloque <id>]; kitlegal no ha comprobado que se aplique a este asunto.
```

Por qué: quien pregunta va a actuar con lo que lee; un plazo o un órgano dicho de memoria puede ser de otra materia o
haber cambiado, y lo que la base no cubre nadie lo ha comprobado aquí. Decirlo le deja buscarlo donde sí está.
