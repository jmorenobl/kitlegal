# Contrato: `boe-legislacion` v0.1.3

FR-010 a FR-018, FR-091; SC-001, SC-005, SC-010. Causa en [research.md](../research.md) («Causa de raíz del ruido en
v0.1.2»); decisión en D1 y, para la comprobación de la prosa, D4. Las líneas son las de `skills/boe-legislacion/SKILL.md`
v0.1.2 en `main`.

## 1. Cambios, cada uno con su causa

El texto nuevo es el que se escribe; donde dice «…», el de v0.1.2 sin cambios.

| Cambio | v0.1.2 (líneas) | v0.1.3 | Causa (research) | Requisito |
|---|---|---|---|---|
| C1 | 73-74: «…una segunda lectura del mismo bloque apagaría lo que la memoria de consultas tiene que decirte (más en «Memoria de consultas»).» | «…una segunda lectura del mismo bloque apagaría el aviso de que su redacción ha cambiado desde una lectura anterior (más en «Redacción modificada desde una lectura anterior»).» | «memoria de consultas» | FR-010 |
| C2 | 76: «…termina con el código 4 o 5…» | «…termina con `4` o `5`…» | «código 4» | FR-010 |
| C3 | 89: «…termina con el código 3 (no encontrado)…» | «…termina con `3` (no encontrado)…» | «código 3» | FR-010 |
| C4 | 106-114: «Cuando ya no quede nada por leer, y antes de redactar la respuesta, comprueba la memoria de consultas una vez por cada norma… Si da `version-obsoleta`, dilo con la forma fija de «Memoria de consultas»; si no, no digas nada de ella.» | «La comprobación de la redacción va después de la última lectura y antes de la respuesta: una vez por cada norma cuyos bloques vas a citar, con esa norma y los bloques de ella que has leído, comprueba si su redacción ha cambiado desde una lectura anterior:» (el bloque `bash` de siempre) «No la pidas nunca sin argumentos ni antes de leer. Por cada entrada de `data.hallazgos` con la clase `version-obsoleta`, la respuesta lleva la línea `⚠ REDACCIÓN MODIFICADA:` de «Redacción modificada desde una lectura anterior»; si termina con otro código que `0`, la regla 7.» | «antes de redactar la respuesta» (→ «Redacto la respuesta.», 2 de 10); «memoria de consultas»; el caso vacío nombrado («si no, no digas nada») | FR-010, FR-011, FR-015 |
| C5 | 115-119: «**La respuesta empieza por lo que se pregunta.** … tampoco para decir que no hay nada que decir ni para anunciar que vas a responder. Salvo la forma…, la respuesta no nombra la memoria de consultas, `kitlegal graph`…, los códigos de salida, los hallazgos, las clases del binario…, el JSON ni el sobre.» | «**La respuesta empieza por lo que se pregunta.** Quien pregunta no ve las órdenes que ejecutas ni lo que devuelven: le sirven la norma, su texto, su cita, sus avisos de vigencia y, si la redacción cambió, la línea `⚠ REDACCIÓN MODIFICADA:`. La respuesta está hecha de eso: no cuentes lo que has hecho ni lo que ha devuelto ninguna orden, ni anuncies que vas a responder.» | la enumeración enseña cada palabra; el caso vacío nombrado | FR-010, FR-011, FR-014 |
| C6 | 120-123: «…Sin `version-obsoleta`, no digas nada de lo consultado antes, ni que ha cambiado ni que no: la comprobación sin hallazgos no distingue…» | «**Nada de otra conversación.** No sabes qué se preguntó ni qué se respondió en otra conversación: no hables de ello, ni para afirmarlo, ni para confirmarlo, ni para desmentirlo. Lo único que la respuesta dice de una lectura anterior es la línea `⚠ REDACCIÓN MODIFICADA:` de un bloque cuya redacción ha cambiado.» | «sin hallazgos» y el caso vacío | FR-011, FR-016 |
| C7 | 178-200, sección «Memoria de consultas» | Sección «Redacción modificada desde una lectura anterior» (§2) | «memoria de consultas», «hallazgo», «trasládalo», «no se traslada», el ejemplo con 20161002 y 20250101, el caso vacío | FR-010, FR-011, FR-013 |
| C8 | 204-206: «Códigos de salida: 0 correcto, 2 argumentos inválidos, …» | «`kitlegal` se invoca desde el `PATH`. Cada orden termina con `0` si todo fue bien, `2` con argumentos inválidos, `3` si no encuentra lo pedido, `4` si la fuente no está disponible, `5` si la fuente limita el ritmo, `6` si hace falta la identidad de una persona y `1` ante un fallo inesperado (por ejemplo, un `world.db` que no se puede leer en `kitlegal graph`).» | «códigos de salida» | FR-010 |
| C9 | 239-240 (regla 2): «…falla —código 3 (no encontrado), 4 (fuente no disponible) o 5 (límite de ritmo)…» | «…falla —termina con `3` (no encontrado), `4` (fuente no disponible) o `5` (límite de ritmo)…»; el resto de la regla, igual | «código 3» | FR-010, FR-016 |
| C10 | 256-257 (regla 6): «…: la memoria de consultas dice qué hay que volver a comprobar, no qué dice el artículo.» | «…: lo que devuelve dice qué hay que volver a comprobar, no qué dice el artículo.» | «memoria de consultas» | FR-010, FR-016 |
| C11 | 258-266 (regla 7) | §3 | «Una comprobación con hallazgos no es un fallo… con hallazgos o sin ellos… trasládalos» (→ «Sin hallazgos que trasladar.», 8 de 10); «responde igual con el texto» (→ «así que respondo con el texto vigente») | FR-012 |

## 2. La sección «Redacción modificada desde una lectura anterior» (C7)

```markdown
## Redacción modificada desde una lectura anterior

`kitlegal` recuerda en local los bloques que ha leído con `kitlegal boe articulo` o `articulos` y qué redacción vio
cada lectura. `kitlegal graph check <norma> <bloques>... --json` compara, para esa norma y esos bloques, la redacción de
la última lectura con la de la anterior, y devuelve en `data.hallazgos` una entrada por cada bloque en que encuentra
algo, con su `clase`. De `kitlegal graph`, el protocolo solo usa `check`: una vez por norma citada, después de la última
lectura y antes de la respuesta (paso 5).

- Una entrada de `clase` `version-obsoleta` dice que la redacción de ese bloque ha cambiado desde la lectura anterior.
  La respuesta lo dice con su forma fija, `⚠ REDACCIÓN MODIFICADA:` —`⚠`, la etiqueta `REDACCIÓN MODIFICADA` y dos
  puntos—, y detrás, en la misma línea, las dos fechas de vigencia tal como las da esa entrada (`AAAAMMDD`): la de la
  redacción superada (`fecha_vigencia`) y la de la que acabas de leer (`fecha_vigencia_reciente`). La forma, con
  `AAAAMMDD` en lugar de cada fecha:

  ```text
  ⚠ REDACCIÓN MODIFICADA: la redacción con fecha de vigencia AAAAMMDD, la que se consultó antes, ha sido sustituida por la de AAAAMMDD, que es la que se cita.
  ```

  Decirlo con otras palabras no vale: la línea va con su forma fija.
- Una entrada de `clase` `fuente-caducada` no va en la respuesta: la respuesta cita el texto que acabas de leer, que la
  caché no sirve pasada su vigencia.
- Lo único que la respuesta dice de una lectura anterior es esa línea. Si `kitlegal graph check` termina con otro
  código que `0`, la regla 7.

La etiqueta `REDACCIÓN MODIFICADA` no es la de ningún aviso de vigencia.
```

## 3. La regla 7 (C11)

```markdown
7. **Si la comprobación de la redacción no termina con `0`.** Si `kitlegal graph check` termina con otro código, la
   respuesta cita igual el texto leído con `kitlegal boe` y lleva esta frase, sin afirmar que la redacción ha cambiado
   ni que no:

   ```text
   No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior.
   ```

   La línea `⚠ REDACCIÓN MODIFICADA:` solo va cuando la comprobación termina con `0` y trae `version-obsoleta`
   («Redacción modificada desde una lectura anterior»).
```

(a) y (b) de FR-012; no nombra el resultado vacío ni lo empareja con ninguna acción.

## 4. Lo que se queda (FR-016, FR-017)

El frontmatter; los cinco pasos y su orden; la forma de la cita y «Cómo se cita» entero; la forma de los avisos y
«traslada» donde es de los avisos (paso 4, paso 5, «Cómo se cita», regla 3); la forma `⚠ REDACCIÓN MODIFICADA:` con sus
dos fechas en la misma línea; la frase fija de la regla 7; las reglas 1, 3, 4 y 5 enteras, y lo que dicen la 2 y la 6;
la región generada, sin tocar (`make skills-check` sin drift). Menos de 300 líneas (v0.1.2 tiene 266; v0.1.3, del mismo
orden). Ni evals, ni el job, ni modelos (H5 FR 077). Siguen pasando las subpruebas `avisos-de-la-skill`,
`hallazgos-de-la-skill` (la forma `⚠ REDACCIÓN MODIFICADA:` sigue en un bloque `text`) y `expresiones-de-la-skill`.

## 5. La comprobación de la prosa (subprueba `prosa-de-la-skill`; FR-091, SC-005)

Sobre `skills/boe-legislacion/SKILL.md` entero, frontmatter incluido:

1. quita cada bloque delimitado por una línea que empieza por ```` ``` ```` (tras sangría) y su cierre, y la región
   entre `<!-- inicio de la tabla de comandos` y `<!-- fin de la tabla de comandos -->`;
2. sustituye cada tramo de código en línea (`` `…` ``) por un espacio;
3. parte en párrafos por las líneas en blanco, por esas marcas y por cada línea que abre un elemento de lista (`- `,
   `* ` o `<n>. ` tras la sangría), y junta con un espacio las líneas de cada párrafo;
4. aplica `ExtraerExpresionesProhibidas` con la lista del repositorio a cada párrafo.

Falla con una línea por párrafo con expresiones —el número de su primera línea y las expresiones—. Además, sobre el
fichero entero sin quitar nada, falla con cada línea que casa
`(?:^|[^0-9])[0-9]{4}(?:0[1-9]|1[0-2])(?:0[1-9]|[12][0-9]|3[01])(?:$|[^0-9])`. Medido con un prototipo de este algoritmo
(research V17): v0.1.2 da 14 párrafos (los que empiezan en las líneas 73, 75, 88, 106, 113, 115, 120, 178, 185, 197, 204,
239, 254 y 258) y la fecha de la línea 191; `SKILL.md` con los cambios C1-C11 tal como los escribe este contrato, 0
párrafos, 0 fechas y 270 líneas.

El test unitario de la extracción (`TestProsaDeLaSkill`, tabla) fija: una expresión en un tramo de código no cuenta;
dentro de un bloque delimitado o de la región generada, tampoco; partida por un salto de línea dentro de un párrafo,
sí; en el frontmatter, sí; dos párrafos no se juntan; `20250101` cuenta como fecha, `AAAAMMDD`, `BOE-A-2015-10565`
y `123456789` no.

## 6. Uso, de fuera adentro

- **Quién pide y cuántas veces**: el modelo carga `SKILL.md` una vez por conversación en que se activa la skill. Las
  órdenes por pregunta no cambian: cada bloque una vez con `kitlegal boe articulo` o `articulos`, y una
  `kitlegal graph check <norma> <bloques>... --json` por norma citada, después de leer.
- **Tamaño**: `SKILL.md` < 300 líneas. La salida de la comprobación la acota H7.1 a la pregunta —≈ 300 B sin nada que
  decir, ≤ 3 800 B con cinco bloques cambiados, ≤ 50 entradas—, con cientos de normas y miles de bloques consultados
  igual que con uno: no crece con lo acumulado.
- **La respuesta**: lo que ocupa la norma, su cita y sus avisos, y 0 líneas sobre la comprobación; ≈ 160 B por bloque
  cuya redacción cambió (0 en la mayoría de las preguntas, como mucho k × 160 B con k bloques leídos); ≈ 83 B de la frase
  de la regla 7 si la comprobación falla. Ejemplo de la línea, 161 bytes:
  `⚠ REDACCIÓN MODIFICADA: la redacción con fecha de vigencia 20180309, la que se consultó antes, ha sido sustituida por la de 20200206, que es la que se cita.`
- **Cuándo deja de darse**: la línea sale en la respuesta de la lectura que ve la redacción nueva y no en la siguiente
  sobre ese bloque (H7.1 FR 024); pasado un mes, solo si el BOE publica otra redacción. La frase de la regla 7, solo en
  la respuesta cuya comprobación falló.

## 7. `CHANGELOG.md`

En *Unreleased*, «Cambiado»: «**`boe-legislacion` v0.1.3**: la respuesta empieza por la norma sin el estado de la
comprobación delante; la skill ya no enseña con su prosa el vocabulario que la respuesta no puede decir, la regla 7
dice solo qué hacer si la comprobación de la redacción falla y cuándo va `⚠ REDACCIÓN MODIFICADA:`, y el ejemplo de esa
línea lleva `AAAAMMDD` en lugar de fechas que copiar» (FR-018).
