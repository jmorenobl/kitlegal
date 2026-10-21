# Contrato: `legal-core` v1.2.3

Quién lo consume: el agente que sigue el protocolo de `skills/legal-core/SKILL.md` cuando una lectura del BOE falla, y
la persona que recibe la respuesta. Requisitos: FR-029 a FR-031; SC-012 a SC-014, SC-021.

## 1. Lo que cambia

| Paso | Hoy (`skills/legal-core/SKILL.md`) | Con v1.2.3 | Requisito |
|---|---|---|---|
| 7 | la línea `⚠ NO CUBIERTO:` nombra solo lo que trae la entrada, no pide `plazos calcular` y va aunque falle una lectura, que se dice aparte (líneas 104-112) | lo mismo, y tras esa línea la respuesta **no expone la norma que rige el procedimiento** —sus plazos, sus recursos—: la nombra la línea, con lo que trae `references/cobertura_por_materia.md`, y llega con su vertical | FR-030 |
| 10 | cada precepto de `data.preceptos` se lee en la fecha de inicio y va su línea `⚠ REDACCIÓN A FECHA:` (líneas 148-160) | lo mismo, y **si esa lectura sale con 4 o 5**: la línea `⚠ PLAZO:` se mantiene, con la fecha de `plazos` y la cita de cada precepto de `data.preceptos`; en lugar de la línea `⚠ REDACCIÓN A FECHA:` de ese precepto, la respuesta dice aparte que no pudo leer esa redacción, como `boe-legislacion` dice cualquier lectura fallida, y no afirma lo que dice el precepto | FR-029 |

Ningún otro paso, regla, referencia ni fila de la tabla de comandos cambia; ninguna herramienta nueva. Las líneas de
forma fija («Cómo se presenta un plazo») no cambian.

## 2. Por qué la fecha se mantiene

La fecha y el precepto salen de `plazos`, que no consulta la red y que ya comprobó contra el BOE grabado que su apartado
regía el día de inicio (ADR 0042, ADR 0043 punto 6): con el BOE caído, quien pregunta sigue necesitando la fecha. Lo que
la respuesta no tiene es el texto de esa redacción, y por eso no lo afirma.

## 3. Longitud y versión

`SKILL.md` tiene hoy 299 líneas, el máximo (`internal/skills/sincronia.go:27`): lo que entra se compensa reescribiendo
con menos líneas los dos párrafos que ya tratan la lectura fallida, el del paso 7 y el primer punto del paso 10, sin
quitar ninguna regla ni ningún caso. Con 300 o más, `make skills-check` sale en rojo (FR-031). La versión, 1.2.3, la
nombra la entrada de `CHANGELOG.md` (*Unreleased*); la 1.3 es la de H9.1.

## 4. Qué lo comprueba

| Qué | Cómo | Dónde |
|---|---|---|
| que el texto de los pasos 7 y 10 lo dice | la revisión final, por lectura (capa 2) | — |
| frontmatter, referencias, tabla de comandos y longitud | la validación de las skills | `make skills-check` (`TestSkillsDelRepositorio`) |
| la línea `⚠ PLAZO:` con `2026-10-13` en la 11 | `plazo_sin_fecha:<modelo>:<modo>`, con 0 | job de `legal-core` |
| la lectura de `a124` intentada en la 11 | `lectura_sin_intentar:<modelo>:<modo>`, con 0; y el comando de bloque, satisfecho con 4 o 5 (FR-016) | job de `legal-core` |
| la línea `⚠ NO CUBIERTO:` con su norma y ninguna línea del plazo en la 07, la 08 y la 12; `plazos calcular` prohibido | `cobertura_sin_declarar:<modelo>:<modo>`, con 0; `prohibidos` en cada sesión | job de `legal-core` |

No lo comprueba ningún control del job, y el informe final lo enseña como límite conocido (spec, «Fuera de alcance»):
que la respuesta de la 11 diga aparte que no pudo leer la redacción y no afirme el precepto; que no ponga de memoria la
línea `⚠ REDACCIÓN A FECHA:`; y que tras `⚠ NO CUBIERTO:` no exponga los plazos, los recursos ni el procedimiento de la
norma que lo rige.
