# Contrato: la lista de expresiones prohibidas, por clases

`evals/boe-legislacion/expresiones-prohibidas.yaml`, su esquema `schemas/expresiones-prohibidas.yaml.json` y cómo la
aplica `internal/evals` (research D6, D7, D8). Sustituye, en lo que dicen, al contrato de H7.3 del mismo nombre (§4, el
calibrado) y a H7.3 FR 022 (las formas de la skill). La comparación de cada expresión es la de H7.2 FR 051 y no cambia.

## 1. El fichero

Cinco claves, las cinco obligatorias. Las tres de hoy (`maquinaria`, `otra_conversacion`, `anuncio`) son de la clase A;
`redaccion_no_leida`, de la clase B; `formas_fijas` no es una familia: son las formas que se quitan antes de buscar.

```yaml
# Expresiones que no lleva la respuesta de boe-legislacion en una eval que activa la skill (H7.2, FR-050; H7.3, FR-020;
# H7.4, FR-030), en cuatro familias. Clase A: la maquinaria interna, lo dicho en otra conversación y el anuncio de la
# respuesta o del estado de lo comprobado (de la comprobación o de una lectura anterior, o de lo que el agente tiene,
# necesita o va a hacer). Clase B, redaccion_no_leida: una redacción que ninguna orden devolvió (qué decía, hasta cuándo
# rigió o qué cambió respecto de ella). Se comparan por la forma, sin distinguir mayúsculas y con blancos y énfasis de
# Markdown entre palabras, delimitadas como palabras, después de quitar las formas fijas que enseña la skill;
# lo dicho con otras palabras no se detecta.
maquinaria: [ …las 22 de hoy, sin cambios… ]
otra_conversacion: [ …las 16 de hoy, sin cambios… ]
anuncio:
  # …las 14 de hoy, sin cambios, y detrás:
  - lectura anterior
  - lecturas anteriores
  - consulta anterior
  - consultas anteriores
  - cambio de redacción
  - cambios de redacción
  - cambio en la redacción
  - cambios en la redacción
  - la comprobación de redacción
  - la comprobación de la redacción
  - la comprobación de cambios
  - la comprobación no
  - redacción posterior
  - posterior a la consultada
  - haya cambiado desde
  - se consultó antes
  - tengo suficiente
  - no necesito
  - i have enough
  - i don't need
  - i don’t need
  - respond now
  - respondo de memoria
  - contesto de memoria
  - extracto fiel
redaccion_no_leida:
  - ya no exige
  - ya no se exige
  - se eliminó
  - vigente hasta
  - aplicable hasta
  - el cambio relevante
  - el cambio más relevante
  - artículo cambió
  - no ha cambiado por
  - esa versión ya no
formas_fijas:
  - "⚠ REDACCIÓN MODIFICADA: <cita>: la redacción con fecha de vigencia <fecha>, la que se consultó antes, ha sido sustituida por la de <fecha>, que es la que se cita."
  - "No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior"
```

87 expresiones (22 + 16 + 39 + 10), ≈ 2,3 KB con el comentario; fijas, no crecen con el uso.

## 2. El esquema (tarea `[datos]`)

`schemas/expresiones-prohibidas.yaml.json`: `required` pasa a `["maquinaria", "otra_conversacion", "anuncio",
"redaccion_no_leida", "formas_fijas"]`; `redaccion_no_leida` con `$ref: #/$defs/expresiones` (la de hoy);
`formas_fijas`, `{"type": "array", "minItems": 1, "items": {"type": "string", "pattern": "^[^\\n]*[^\\s<>][^\\n]*$"}}`
(una línea con algo más que marcadores). `additionalProperties: false` se queda. Los casos de `formato_test.go` que lo
fijan: la lista sin `redaccion_no_leida`, sin `formas_fijas`, con `formas_fijas: []`, con una forma de dos líneas, y la
lista del repositorio, que lo cumple.

## 3. Las formas fijas y la búsqueda

`ExtraerExpresionesProhibidas(texto, lista)`:

1. **Quita cada forma fija de la lista**, en su orden, en todas sus apariciones. Una forma compila a una expresión
   regular: su texto con `regexp.QuoteMeta`, cada blanco como `[ \t]+`, `<fecha>` como `[0-9]{8}`, y `<cita>` como un
   grupo **opcional** con la cita, `[^\[\n]*\[[^\]\n]*,[ \t]*bloque[ \t]+[^\]\n]+\]` (de lo que haya tras el texto
   anterior al corchete de cierre de una cita, en la misma línea), y lo que la sigue en la forma hasta el blanco
   siguiente, ese blanco incluido: en la línea de la redacción, `(?:<cita>:[ \t]+)?`. Así se quita la línea con su
   cita, la que enseña v0.1.4, y también sin ella, la que escribían v0.1.1 a v0.1.3 y la que llevan las respuestas de
   los tres informes del calibrado (§4). La lista no mide que la línea lleve su cita (FR-023): eso lo juzga la eval 20
   (FR-052). Un `⚠` al principio admite los blancos y el énfasis de Markdown de la forma fija de los avisos (H5.1). La
   frase de la regla 7 va sin su punto final: la respuesta puede seguirla de un paréntesis (14-01 de `196ee05`).
2. **Quita la marca, la etiqueta y los dos puntos de cada aviso de vigencia**, con las expresiones de `ExtraerAvisos`
   (`boe.EtiquetasDeAviso`). Lo que sigue a la etiqueta en la línea se busca.
3. Cada tramo quitado se cambia por un salto de línea (una expresión no casa a través de un salto: sus palabras de los
   dos lados no forman una).
4. **Busca** como hoy, en el orden `maquinaria`, `otra_conversacion`, `anuncio`, `redaccion_no_leida`, sin repetir.

Así, «la que se consultó antes» de la línea y «desde una consulta anterior» de la regla 7 no cuentan dentro de sus
formas; las mismas palabras fuera de ellas, o en otra redacción de la línea, sí: `se consultó antes` y `consulta
anterior`, de `anuncio`. `ExpresionesProhibidas.esDeLaClaseB` (sin exportar) dice si una expresión encontrada es de
`redaccion_no_leida`. Una lista sin familias (la de una skill sin lista) no quita nada ni encuentra nada.

## 4. El calibrado (`expresiones-calibradas`, FR-032, FR-093; SC-002)

Sobre las 93 respuestas de cada informe, por las dos cifras de la eval y por columna —`maquinaria`,
`otra_conversacion`, `anuncio`, `redaccion_no_leida` y `alguna`—, exactamente (las demás evals, 0 en todas):

| Eval | H7.1 (36) | H7.2 (11) | H7.3, `196ee05` (9) |
|---|---|---|---|
| 01 | — | — | anu 1, alg 1 |
| 02 | maq 1, alg 1 | — | — |
| 03 | maq 3, anu 3, alg 3 | maq 1, anu 1, alg 1 | — |
| 04 | maq 3, anu 3, alg 3 | — | — |
| 05 | maq 3, anu 1, alg 3 | — | anu 1, alg 1 |
| 06 | maq 3, anu 2, alg 3 | maq 1, anu 1, alg 1 | — |
| 07 | maq 3, anu 2, alg 3 | — | — |
| 08 | maq 2, anu 1, alg 2 | — | — |
| 09 | maq 2, anu 1, alg 2 | — | — |
| 13 | maq 3, anu 2, alg 3 | maq 2, anu 2, alg 2 | anu 2, alg 2 |
| 14 | maq 3, anu 2, alg 3 | maq 2, anu 2, alg 2 | maq 1, anu 3, alg 3 |
| 15 | maq 3, anu 3, alg 3 | maq 3, anu 3, alg 3 | — |
| 16 | maq 2, anu 2, alg 2 | — | — |
| 17 | maq 3, anu 2, alg 3 | — | — |
| 19 | otra 1, anu 1, red 1, alg 2 | maq 1, anu 1, red 2, alg 2 | anu 1, red 2, alg 2 |

Informes: `specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json`,
`specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json` y
`specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json`. Cada diferencia nombra el informe, la eval, la columna,
lo contado y lo calibrado, como hoy. Medido con el prototipo (research V20). El `anu 1` de la 19 es `se consultó antes`
fuera de la línea: la 19-01 de H7.1 («lo que se consultó antes era la versión previa»), la 19-02 de H7.2 y la 19-02 de
`196ee05` («es distinta de la que se consultó antes»), las tres ya marcadas por otra familia, así que `alguna` no
cambia. La línea de la redacción de esos informes va sin cita (v0.1.1 a v0.1.3) y se quita porque `<cita>` es opcional
(§3); con la cita obligatoria, `se consultó antes` marcaría en cada informe una respuesta más de la 19, que no lleva
otra expresión que la de esa línea: 37, 12 y 10.

## 5. Las formas de la skill y los bloques (FR-033; SC-003)

- `expresiones-en-los-bloques`: sin cambios; con la eval 20 lee también `da-3` y las dos derivadas de su grafo previo.
  0 expresiones (research V21).
- `expresiones-de-la-skill`: compone una respuesta, una línea por pieza, con la forma escrita de cada aviso de vigencia
  y de cada clase de hallazgo, seguida cada una de una frase de ejemplo (`esta norma ha sido derogada.`), y con cada
  bloque `text` de `SKILL.md` con sus marcadores sustituidos (`AAAAMMDD` → `20180309` y, en su orden, `20200206`;
  `<forma legible> [<identificador>, bloque <id>]` → `art. 118 de la Ley 9/2017 [BOE-A-2017-12902, bloque a1-30]`;
  `<identificador>, bloque <id>]` → `BOE-A-2015-10565, bloque a21]`). Exige que `ExtraerExpresionesProhibidas` no
  encuentre ninguna y que **cada** forma fija de la lista quite algo de esa respuesta (si la skill y la lista dejaran de
  decir lo mismo, la forma no quitaría nada). La premisa: sin quitar las formas fijas, la respuesta sí lleva alguna
  (`se consultó antes`, de la línea de la redacción, y `consulta anterior`, de la regla 7).

## 6. Los tests de `Juzgar` (FR-034, FR-094; SC-007)

`TestJuzgarLasClasesDeLaRespuesta` (en `juzgar_test.go`), con la lista del repositorio y una eval que activa la skill;
cada frase, entera, en una respuesta con una cita esperada; para cada una, la clase de lo que marca:
«A» = alguna expresión y ninguna de `redaccion_no_leida`; «B» = alguna de `redaccion_no_leida`.

| Origen | Sesión | Clase | Frase |
|---|---|---|---|
| `196ee05` | 01-02 | A | No hay avisos de vigencia sobre este bloque ni cambios de redacción respecto a una consulta anterior. |
| `196ee05` | 05-02 | A | No se encontraron avisos de vigencia sobre este bloque, y no consta que su redacción haya cambiado desde una consulta anterior. |
| `196ee05` | 13-02 | A | Este artículo ha sido modificado en varias ocasiones (última modificación por el Real Decreto Legislativo 7/2015, con vigencia desde el 31/10/2015); no presenta avisos de vigencia adicionales ni indicios de redacción posterior a la consultada. |
| `196ee05` | 13-03 | A | No se ha detectado ningún cambio de redacción respecto a una lectura anterior de este bloque. |
| `196ee05` | 14-01 | A | Article 59 already answers the question fully, so I don't need art. 60. I have enough to respond now. |
| `196ee05` | 14-02 | A | Con el artículo 59 tengo suficiente para responder a la pregunta completa; no necesito el artículo 60 para esto. |
| `196ee05` | 14-03 | A | El artículo 59 responde directamente a la pregunta. La comprobación de redacción terminó sin hallazgos, así que no hay cambios que señalar. |
| `196ee05` | 19-01 | B | El artículo cambió porque el Real Decreto-ley 3/2020, de 4 de febrero, modificó el art. 118 de la LCSP (entrada en vigor el 20180206... corrección: el 20200206). Esto elevó el umbral de los contratos menores de obras de 40.000 € (ya estaba así) pero, sobre todo, cambió el apartado 2: ya no exige tres informes separados, sino que ahora basta con un único informe del órgano de contratación. |
| `196ee05` | 19-02 | B | Esta redacción procede de la modificación del Real Decreto-ley 3/2020, de 4 de febrero (disposición final 1.1) [BOE-A-2020-1651], vigente desde el 6 de febrero de 2020, y es distinta de la que se consultó antes (vigente hasta el 9 de marzo de 2018): esa versión anterior no distinguía umbrales entre obras y suministros/servicios del mismo modo ni incluía la excepción del apartado 5 sobre anticipos de caja fija. |
| `6ab3add` | 03-03 | A | Sin cambios de redacción respecto a lecturas anteriores. Aquí está la respuesta. |
| `6ab3add` | 13-03 | A | Sin cambios desde una lectura anterior. Respondo. |
| `6ab3add` | 19-01 | B | Esta es la redacción vigente hoy; la versión anterior, aplicable hasta el 5 de febrero de 2020, no incluía la excepción del apartado 5 ni la obligación de motivar que no se fracciona el objeto para eludir el umbral, entre otras diferencias de redacción. |
| `eb6b4c8` | 14-01 | A | Este artículo ya da la respuesta completa. Sin avisos de vigencia ni cambios desde una lectura anterior. |
| `eb6b4c8` | 19-01 | B | Este artículo cambió respecto a su redacción original de 2017: hasta el 5 de febrero de 2020, el apartado 2 exigía justificar solo la necesidad del contrato (sin el requisito de no fraccionamiento) y no existía la excepción del apartado 5 para anticipos de caja fija; ambos elementos los introdujo el Real Decreto-ley 3/2020. |
| `eb6b4c8` | 19-02 | B | Esta redacción proviene de la modificación introducida por el Real Decreto-ley 3/2020, de 4 de febrero, vigente desde el 6 de febrero de 2020 — distinta de la versión (vigente hasta el 9 de marzo de 2018) que pudiste haber consultado anteriormente. El cambio más relevante respecto a versiones anteriores es que ya no exige un informe del "órgano de contratación" motivando la elección del contratista entre varios presupuestos (esa exigencia de comparar ofertas se eliminó); ahora el foco está en justificar la necesidad y la no alteración del objeto para evitar los umbrales. |
| H7.1 | 19-01 | B | Es decir, lo que se consultó antes era la versión previa a su entrada en vigor, y lo que ahora se cita es la versión ya en vigor — el contenido normativo del artículo 21 no ha cambiado por una reforma posterior, simplemente la versión indexada pasó a ser la vigente desde la entrada en vigor de la ley. |
| H7.2 | 19-01 | B | Si lo que tenías en mente era la redacción original de 2018 (por ejemplo, sin el requisito específico de motivar la no alteración del objeto contra el fraccionamiento en los términos actuales, y con otro régimen sobre la publicación/aprobación previa del gasto), esa versión ya no está vigente. |
| H7.2 | 19-02 | B | ⚠ REDACCIÓN MODIFICADA: la redacción con fecha de vigencia 20180309, que es la que probablemente se consultó en aquella ocasión anterior, ha sido sustituida por la de fecha de vigencia 20200206, que es la que se cita aquí. El cambio relevante fue que se eliminó la exigencia de un informe adicional del órgano de contratación acreditando que el contratista no había suscrito otros contratos menores por importe superior, individual o en conjunto, al umbral —ese control de acumulación desapareció de la ley en esta redacción; ahora el informe del apartado 2 solo debe justificar la necesidad y la no alteración del objeto para eludir umbrales. |
| línea de base | 07-03 | A | La comprobación de redacción no detecta cambios desde una lectura anterior. |
| línea de base | 01-03 | A | Es un extracto fiel del texto consolidado. Aquí va apartado por apartado. |
| línea de base | 04-01 | A | Respondo de memoria, sin haber consultado el texto consolidado del BOE. |

Sin marcar (0 expresiones): la respuesta compuesta de §5; «No hay avisos de vigencia sobre este bloque.»; la 19-02 de la
línea de base, «No puedo decirte qué contenía la redacción anterior.»; y la frase que enseña v0.1.4, en la forma en que
la diría la respuesta: «Cito la redacción vigente; la que había antes no la he leído, así que no puedo decir qué ha
cambiado.». Y las mismas palabras fuera de sus formas, marcadas en la clase A: «Esta redacción es distinta de la que se
consultó antes.» (`se consultó antes`, sola) y «La redacción es la misma que la que se consultó antes y no ha cambiado
desde una consulta anterior.» (`se consultó antes` y `consulta anterior`).

`TestExtraerExpresionesProhibidas` gana los casos de §3: la forma de la línea con su cita, con énfasis, con otra cita y
sin cita (la de v0.1.3), que se quita en los cuatro; la frase de la regla 7 con punto, sin él y seguida de un
paréntesis; una etiqueta de aviso seguida de una expresión en la misma línea (se encuentra); la línea de la redacción
escrita con otras palabras, que se busca entera («⚠ REDACCIÓN MODIFICADA: la redacción que se consultó antes ha sido
sustituida.» da `se consultó antes`); «la que se consultó antes» sola, fuera de la forma (da `se consultó antes`); y una
lista sin formas fijas, que no quita nada.

## 7. Dónde se aplica (FR-036)

En `Juzgar` a cada sesión de una eval que activa la skill; en la regla por serie (una sesión con una expresión no
pasa); en `expresiones_prohibidas` de cada sesión y en el recuento por modelo; en los umbrales
([informe-del-job.md](./informe-del-job.md)); en el sondeo; y en las subpruebas de §4 y §5 y en `prosa-de-la-skill`.
