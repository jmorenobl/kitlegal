# Contrato: `jurisprudencia` v0.1

Dos pasajes de `skills/jurisprudencia/SKILL.md` cambian; todo lo demás se queda (FR-080 a FR-084). El cambio entero
es [skill-jurisprudencia-v0.1.diff](./skill-jurisprudencia-v0.1.diff), que se aplica sobre la cabeza con `git apply`
(research M11) y deja 197 líneas.

## 1. La causa, medida

| Qué pasa | Cuánto | Pasaje de v0 que lleva a ello |
|---|---|---|
| La respuesta nombra un CAPTCHA, y a veces se lo anuncia a la persona o le pide que lo resuelva | 8 de las 35 respuestas del modelo que decide en el cierre de H23 y 6 de las 18 del sondeo con la skill | Párrafo inicial, líneas 18-19: «el buscador pide un CAPTCHA a los programas, y no se sortea» |
| La respuesta da el ECLI que `cita preparar --roj` deduce como si fuera un dato de la sentencia | 5 de las 6 respuestas del cierre de H23 que lo dan | Viñeta del equivalente del paso 2, líneas 76-77: «puedes nombrarlo junto a la referencia» |

Las sesiones y sus frases, una a una, están en research.md, «Traza de las dos frases».

## 2. Los dos cambios

**C1 · El CAPTCHA es de los programas, y la respuesta no lo anuncia** (FR-081).

v0:

> kitlegal **no consulta el buscador de jurisprudencia del CENDOJ**: el buscador pide un CAPTCHA a los programas, y
> no se sortea. La búsqueda la hace la persona con su navegador, y la skill le quita el trabajo de alrededor: […]

v0.1:

> kitlegal **no consulta el buscador de jurisprudencia del CENDOJ**: el buscador cierra el paso a los programas con
> un CAPTCHA, y no se sortea. Ese obstáculo es de los programas y no de la persona: a ella, con su navegador, no le
> sale. Por eso la respuesta no le anuncia un CAPTCHA ni le pide que resuelva ninguno. La búsqueda la hace la persona
> con su navegador, y la skill le quita el trabajo de alrededor: […]

Dice las dos cosas que pide FR-081: de quién es el obstáculo y que la respuesta no lo anuncia. Lo dice con su razón,
no como una prohibición: v0 daba el hecho sin decir a quién afecta, y el modelo lo repetía.

**C2 · El equivalente se da como deducido y sin comprobar** (FR-082).

v0:

> - Si la operación da además un **equivalente** —el ROJ de un ECLI o el ECLI de un ROJ—, puedes nombrarlo junto a la
>   referencia. No lo compongas tú: solo se deduce para el Tribunal Supremo, y cuando no viene no lo hay.

v0.1:

> - Si la operación da además un **equivalente** —el ROJ de un ECLI o el ECLI de un ROJ—, puedes nombrarlo junto a la
>   referencia, y entonces dilo como lo que es: deducido de esa referencia, sin que nadie lo haya comprobado. No lo
>   compongas tú: solo se deduce para el Tribunal Supremo, y cuando no viene no lo hay.

## 3. Lo que se queda (FR-083)

La descripción del frontmatter, el protocolo, la forma de la cita y la de la línea `⚠ SENTENCIA NO COMPROBADA:`, las
reglas y la tabla de comandos, que sigue generada sin drift. `SKILL.md` no lleva un campo de versión: la v0.1 se
registra en `CHANGELOG.md` (FR-084).

Hasta la primera medición del cierre, las diferencias de `SKILL.md` con el de `main` son solo C1 y C2. Después, solo
una reparación del cierre con la traza de FR-027 puede tocar otra cosa de la prosa del protocolo o de las reglas, y
nunca la descripción, las dos formas fijas ni la tabla de comandos.

## 4. Uso, de fuera adentro

Lo que la persona recibe por pregunta no cambia de forma: las órdenes por pregunta son las de v0.

| Pregunta | Órdenes o llamadas | Lo que lleva la respuesta |
|---|---|---|
| Una sentencia que no está delante | 1 `cita preparar` por sentencia | La línea (unos 60 bytes) y la consulta: la dirección y hasta tres casillas, unos 300 bytes. Si nombra el equivalente, una frase de unos 100 bytes que dice que es deducido |
| Un documento traído | 1 `cita cotejar` por documento | La cita, los datos de la ficha y lo que dice el texto que trae |
| Una pregunta por materia | 1 `cita preparar` con texto | La dirección de la búsqueda |

- En el sondeo, las 18 respuestas de la skill ocupan de 768 a 2 695 bytes, 1 604 de media (research M10). Con tres
  sentencias pedidas sin documento, las líneas y sus consultas suman alrededor de 1,4 KB.
- Nada depende de lo consultado antes: el applet `cita` no guarda nada, y no hay ninguna señal acumulada.
- La línea y la consulta salen mientras la sentencia no esté delante; con su documento cotejado, las sustituye la
  cita. El equivalente deducido solo sale junto a una referencia sin documento. La respuesta no lleva ninguna frase
  sobre un CAPTCHA.
- El cuerpo de `SKILL.md` lo lee el modelo una vez por conversación en que se activa la skill: 197 líneas.

## 5. Controles

| Qué | Control | Requisitos |
|---|---|---|
| Menos de 300 líneas, frontmatter válido y la tabla de comandos sin drift | `skills-check`, en `make ci` (`TestSkillsDelRepositorio`) | FR-083; SC-011 |
| Los dos pasajes dicen lo que piden FR-081 y FR-082, y lo demás no cambia | La revisión final, por lectura del diff contra `main` | FR-081 a FR-083 |
| El efecto en las respuestas: ninguna anuncia un CAPTCHA, y el equivalente se da como deducido | Una persona, que lee las respuestas del cierre y lo anota en `docs/USO.md` | SC-014 |

No hay un control automático del efecto: la rúbrica no lo pregunta, y una respuesta no se juzga con una lista de
palabras (ADR 0037). Que quite el CAPTCHA de las respuestas no se ha medido (research S5).

## 6. Reparaciones del cierre (FR-027)

Ninguna todavía. Una reparación que toque `SKILL.md` añade aquí su cambio —C3 en adelante—, con las respuestas que lo
provocan por su sesión y con sus frases, y el pasaje que lleva a ellas.
