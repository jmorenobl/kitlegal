# Contrato: `boe-legislacion` v0.1.2

FR-040 a FR-048, FR-086; SC-001, SC-008. La causa de raíz y la traza de cada cambio, en research («Causa de raíz del
ruido») y D16. Los textos de abajo son los que entran en `skills/boe-legislacion/SKILL.md`; la redacción final puede
ajustar el orden de las palabras, no lo que piden.

## 1. Lo que cambia

**Paso 5, primera viñeta** (C4): la orden de `graph check` ya no manda «trasladar lo que encuentre»:

````markdown
- Cuando ya no quede nada por leer, y antes de redactar la respuesta, comprueba la memoria de consultas una vez por
  cada norma cuyos bloques vas a citar, con esa norma y los bloques de ella que has leído:

  ```bash
  kitlegal graph check BOE-A-2015-10565 a21 --json
  ```

  No la pidas nunca sin argumentos ni antes de leer. Si da `version-obsoleta`, dilo con la forma fija de «Memoria de
  consultas»; si no, no digas nada de ella.
````

**Paso 5, dos viñetas nuevas, justo detrás** (C1, C2, C3, C5):

```markdown
- **La respuesta empieza por lo que se pregunta.** Quien pregunta no ve las órdenes que ejecutas ni lo que devuelven:
  le sirven la norma, su texto y su cita. No cuentes lo que has hecho ni lo que ha devuelto ninguna orden, tampoco para
  decir que no hay nada que decir ni para anunciar que vas a responder. Salvo la forma `⚠ REDACCIÓN MODIFICADA:`, la
  respuesta no nombra la memoria de consultas, `kitlegal graph` ni ninguno de sus verbos, los códigos de salida, los
  hallazgos, las clases del binario (`version-obsoleta`, `fuente-caducada`), el JSON ni el sobre.
- **Nada de otra conversación.** No sabes qué se preguntó ni qué se respondió en otra conversación: no hables de ello,
  ni para afirmarlo, ni para confirmarlo, ni para desmentirlo. Sin `version-obsoleta`, no digas nada de lo consultado
  antes, ni que ha cambiado ni que no: la comprobación sin hallazgos no distingue un bloque leído antes y sin cambios de
  uno que nunca se leyó.
```

**«Memoria de consultas», viñeta de `version-obsoleta`** (C6): las fechas, tal como las da el hallazgo y en la misma
línea que la forma:

```markdown
- `version-obsoleta`: la redacción del bloque ha cambiado desde la lectura anterior. Trasládalo con su forma fija,
  `⚠ REDACCIÓN MODIFICADA:` —`⚠`, la etiqueta `REDACCIÓN MODIFICADA` y dos puntos—, y detrás, en la misma línea, las dos
  fechas de vigencia tal como las da el hallazgo (`AAAAMMDD`): la de la redacción superada (`fecha_vigencia`) y la de
  la que acabas de leer (`fecha_vigencia_reciente`). Por ejemplo:
```

(el ejemplo `text` que sigue no cambia).

**«Memoria de consultas», última viñeta** (C9), en lugar de «Con código 0 y sin `version-obsoleta`, no digas nada de la
memoria de consultas; con otro código, la regla 7.»:

```markdown
- Sin `version-obsoleta`, la respuesta no dice nada de la memoria de consultas (paso 5); si `kitlegal graph check`
  termina con otro código, la regla 7.
```

**Regla 2** (C8), la primera frase:

```markdown
2. **Nunca inventar contenido legal.** Si `kitlegal boe` falla —código 3 (no encontrado), 4 (fuente no disponible) o 5
   (límite de ritmo), o sin caché con `--offline`— o no está disponible, di qué no se pudo consultar y por qué con lo
   que significa para quien pregunta —que el artículo no está en la norma, que la fuente no estaba disponible, que la
   fuente limitó las consultas—, sin el código, y no suplas el texto con conocimiento propio.
```

(lo demás de la regla 2 —los bloques por separado, cuáles faltan, `kitlegal` fuera del `PATH`— no cambia).

**Regla 7** (C7), entera:

````markdown
7. **Una comprobación con hallazgos no es un fallo.** `kitlegal graph check` con código 0 es un resultado, con
   hallazgos o sin ellos, y nunca un fallo de la herramienta: trasládalos como dice «Memoria de consultas». Si termina
   con otro código, responde igual con el texto de `kitlegal boe` y di que no se ha podido comprobar si la redacción ha
   cambiado desde una consulta anterior, sin afirmar que ha cambiado ni que no, y sin nombrar la memoria de consultas,
   `kitlegal graph` ni el código:

   ```text
   No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.
   ```
````

## 2. Lo que no cambia (FR-046, FR-047)

- El texto citado sale de `kitlegal boe articulo`/`articulos`, nunca de `graph` (regla 6; H7 FR 081).
- `graph check` con código 0 es un resultado (regla 7; H7 FR 082).
- Una comprobación por norma citada, después de leer; cada bloque una vez por pregunta; `fuente-caducada` no se
  traslada (H7.1 FR 040, 041, 043).
- La etiqueta de la forma la da el binario, con su comprobación en `make ci` (subtest `hallazgos-de-la-skill`; H7.1
  FR 045).
- La forma de la cita, la de los avisos de vigencia, la tabla de comandos (generada; el binario no cambia), el
  frontmatter, los pasos 1 a 4 y las reglas 1, 3, 4, 5 y 6.
- No nombra evals, el job ni modelos (H5 FR 077).

## 3. Controles

| Control | Dónde |
|---|---|
| Frontmatter válido, región generada sin drift, < 300 líneas (hoy 251; v0.1.2, ≈ 265) | `make skills-check` (`TestSkillsDelRepositorio`) |
| La forma de cada aviso y de `version-obsoleta` sigue enseñada | subtests `avisos-de-la-skill` y `hallazgos-de-la-skill` |
| Ningún bloque `text` de `SKILL.md` —la cita, el aviso, la línea `⚠ REDACCIÓN MODIFICADA:` y la frase de la regla 7— lleva una expresión de la lista | subtest `expresiones-de-la-skill` (contracts/lista-y-juicio.md §6) |
| El efecto en las respuestas: ≤ 2 de 51 con Sonnet 5 y ≤ 1 de 30 con Haiku 4.5 con alguna expresión; la eval 19 con la forma | el job de evals de cierre (SC-001) |
| La consulta repetida en Claude Code | quickstart §6 (SC-002) |

## 4. Uso, de fuera adentro

| Salida | Quién y cuántas veces | Tamaño con meses de uso | Cuándo deja de darse |
|---|---|---|---|
| La respuesta de `boe-legislacion` | La persona que pregunta; una por pregunta | Lo que ocupa la norma citada; 0 líneas sobre la comprobación (hoy, ≈ 85 B de «Sin hallazgos en la memoria de consultas. Ya tengo todo lo necesario para responder.» en 2 de cada 3 respuestas de Sonnet 5); una línea `⚠ REDACCIÓN MODIFICADA:` de ≈ 161 B por bloque cuya redacción cambió (la de §1 con las fechas de la LCSP), 0 en la mayoría de las preguntas y como mucho k × 161 B con k bloques leídos; ≈ 83 B de la frase de la regla 7 si `graph check` falla. No crece con lo acumulado: con cientos de normas y miles de bloques consultados es la misma | La línea `⚠ REDACCIÓN MODIFICADA:` sale en la respuesta cuya lectura ve la redacción nueva y no en la siguiente sobre ese artículo (H7.1 FR 024; SC-002); la frase de la regla 7, solo en la respuesta en que falló; nada sobre consultas anteriores sin `version-obsoleta` |
| Lo que la skill lee de `graph check` | La skill; una invocación por norma citada, como en v0.1.1 | ≈ 300 B sin cambios y ≤ 3 800 B con cinco bloques cambiados, con cualquier volumen (H7.1 SC 005); v0.1.2 no añade ninguna invocación | Como en H7.1 |

## 5. `CHANGELOG.md` (*Unreleased*)

- La entrada «`boe-legislacion` v0.1.1 dice que la redacción ha cambiado desde la consulta anterior» pasa a v0.1.2 y
  dice, para quien la usa: que la respuesta empieza por lo que se pregunta y no cuenta la comprobación —ni la memoria de
  consultas, `graph`, códigos, hallazgos, clases, JSON ni sobre—, salvo la línea `⚠ REDACCIÓN MODIFICADA:` con sus dos
  fechas tal como las da el hallazgo; que no habla de lo dicho en otra conversación ni, sin cambio, de lo consultado
  antes; que un fallo de `graph check` se dice como «no se ha podido comprobar si la redacción ha cambiado desde una
  consulta anterior» y uno de `kitlegal boe`, por lo que significa, sin el código; y que sustituye a la v0.1.1 de H7.1,
  que no llegó a publicarse (FR-048).
- La entrada de la eval de la consulta repetida pasa a la eval nueva sobre el art. 118 de la LCSP, con la redacción
  original derivada de la grabación en el grafo y la vigente en la caché, y dice que retira la del art. 21 LPAC de H7,
  cuyo grafo previo sembraba una redacción escrita a mano (FR-086).
- Una entrada nueva: la lista de expresiones prohibidas del formato común de eval, que el informe comprueba sin modelo
  en las evals de `boe-legislacion` que activan la skill, por sesión y con el recuento por modelo, y que decide en las
  que deciden (FR-086).
