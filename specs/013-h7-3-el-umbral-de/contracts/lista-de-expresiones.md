# Contrato: la lista de expresiones prohibidas con la tercera familia

FR-020 a FR-024, FR-095; SC-003, SC-004. Decisiones en [research.md](../research.md) D2 y D3.

## 1. El fichero

`evals/boe-legislacion/expresiones-prohibidas.yaml` gana la clave `anuncio` detrás de las dos que tiene, y su comentario
de cabecera pasa a nombrar tres familias:

```yaml
# … la maquinaria interna, lo dicho en otra conversación y el anuncio de la respuesta o del estado de lo comprobado …
maquinaria:
  # (las 22 de H7.2, sin cambios)
otra_conversacion:
  # (las 16 de H7.2, sin cambios)
anuncio:
  - que trasladar
  - hace falta trasladar
  - ya puedo responder
  - con esto puedo responder
  - y puedo responder
  - tengo todo lo necesario
  - tengo lo necesario
  - redacto la respuesta
  - respondo con el texto
  - respondo con el contenido
  - ya tengo la respuesta
  - ya tengo el texto
  - así que respondo
  - sin redacciones cambiadas
```

52 expresiones en total (≈ 1,3 KB). Ninguna de la familia dice a quien lee que la norma no está derogada o que no tiene
avisos (FR-020).

## 2. El esquema (tarea `[datos]`)

`schemas/expresiones-prohibidas.yaml.json`: `anuncio` en `required` y en `properties`, con el mismo `$ref` que las otras
dos (`#/$defs/expresiones`: lista no vacía de expresiones de una o más palabras, sin blancos en los extremos ni `*` o
`_`). `additionalProperties: false` se queda. Una lista sin `anuncio`, con `anuncio` vacío o con una expresión mal
escrita es un fichero mal formado (FR-023; H7.2 FR 055).

## 3. Tipo y comparación

`ExpresionesProhibidas` gana `Anuncio []string` (`yaml:"anuncio"`). `ExtraerExpresionesProhibidas` recorre maquinaria,
otra conversación y anuncio, en ese orden, sin repetir; la comparación de cada expresión no cambia (H7.2 FR 051).
`recontarExpresiones` trata la lista como vacía solo si lo están las tres familias. Todo lo demás que usa la lista
—`Juzgar`, la regla por serie, las expresiones por sesión, el recuento, los umbrales y el sondeo— la toma entera
(FR-024).

## 4. Calibrado (subprueba `expresiones-calibradas`)

Sobre cada informe versionado, respuesta a respuesta, con `ExtraerExpresionesProhibidas` y la lista del repositorio:

| Eval | H7.1 maquinaria | H7.1 otra | H7.1 anuncio | H7.1 alguna | H7.2 maquinaria | H7.2 otra | H7.2 anuncio | H7.2 alguna |
|---|---|---|---|---|---|---|---|---|
| 02 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 |
| 03 | 3 | 0 | 3 | 3 | 1 | 0 | 1 | 1 |
| 04 | 3 | 0 | 3 | 3 | 0 | 0 | 0 | 0 |
| 05 | 3 | 0 | 1 | 3 | 0 | 0 | 0 | 0 |
| 06 | 3 | 0 | 2 | 3 | 1 | 0 | 1 | 1 |
| 07 | 3 | 0 | 2 | 3 | 0 | 0 | 0 | 0 |
| 08 | 2 | 0 | 1 | 2 | 0 | 0 | 0 | 0 |
| 09 | 2 | 0 | 1 | 2 | 0 | 0 | 0 | 0 |
| 13 | 3 | 0 | 2 | 3 | 2 | 0 | 2 | 2 |
| 14 | 3 | 0 | 2 | 3 | 2 | 0 | 2 | 2 |
| 15 | 3 | 0 | 3 | 3 | 3 | 0 | 3 | 3 |
| 16 | 2 | 0 | 2 | 2 | 0 | 0 | 0 | 0 |
| 17 | 3 | 0 | 2 | 3 | 0 | 0 | 0 | 0 |
| 19 | 0 | 1 | 0 | 1 | 1 | 0 | 0 | 1 |
| **Total** | 34 | 1 | 24 | **35** | 10 | 0 | 9 | **10** |

Las demás evals, 0 en todas las columnas. Los informes: `specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json`
y `specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json`, de 93 respuestas cada uno; no se editan
(FR-080). La subprueba falla nombrando el informe, la eval, la columna, lo contado y lo calibrado. Las evals se nombran
por sus dos cifras (FR-020 de H7.2).

## 5. Comprobaciones que se extienden solas (FR-022)

`expresiones-en-los-bloques` (el texto de cada bloque que leen las evals y sus grafos previos) y
`expresiones-de-la-skill` (las formas escritas de avisos y de `⚠ REDACCIÓN MODIFICADA:` y cada bloque `text` de
`SKILL.md`: la cita, el aviso, la línea con sus dos `AAAAMMDD` y la frase fija de la regla 7) aplican la lista entera: con
`anuncio` en ella, la miran sin cambiar su código. `listaDelRepositorio` exige además que `anuncio` no esté vacía, para
que ninguna pase en vacío.

## 6. Tests

| Test | Caso | Requisito |
|---|---|---|
| `formato_test.go`, casos del esquema | una lista con las tres familias valida; sin `anuncio`, con `anuncio: []` o con una expresión con `*`, no | FR-023 |
| `TestLeerConjunto` | la lista con `anuncio` queda en `Conjunto.Prohibidas.Anuncio` y en cada `Eval.Prohibidas` | FR-024 |
| `TestExtraerExpresionesProhibidas` | orden maquinaria → otra conversación → anuncio; «Nada que trasladar. Ya puedo responder.» da `que trasladar` y `ya puedo responder`; «No hay avisos de vigencia sobre este bloque» no da nada | FR-020, FR-024; US1-3, US1-4 |
| `TestJuzgarLasExpresionesProhibidas` | una respuesta con una expresión de `anuncio` no pasa, con su motivo `expresión prohibida: <expresión>` | FR-024; US1-3 |
| `TestEvalsDelRepositorio/expresiones-calibradas` | §4 | FR-021, FR-095; SC-003; US1-5 |
| `TestEvalsDelRepositorio/expresiones-en-los-bloques`, `/expresiones-de-la-skill` | 0 expresiones de las tres familias | FR-022, FR-095; SC-004 |
