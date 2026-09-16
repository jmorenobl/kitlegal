# T012 · intento 1 · en verde

## De dónde sale cada valor de las evals

Preguntas, `activa`, `reproduce`, comandos esperados y citas esperadas son los de la tabla del contrato
evals-y-grabaciones §2, tal cual. Cada identificador es el de la norma con esa abreviatura en `data/normas.yaml` (T011).
Cada bloque es el que el manifiesto grabó para su eval, confirmado en la pausa de T008. Además, el título del bloque
grabado se comprobó contra el artículo que pide la eval:

| Eval | Norma | Comandos esperados | Cita | Bloque grabado | Título en la grabación |
|---|---|---|---|---|---|
| 01 | LPAC `BOE-A-2015-10565` | bloque | `a21` | H4 | Artículo 21 |
| 02 | LCSP `BOE-A-2017-12902` | `indice`, bloque | `a1-30` | H4 | Artículo 118 |
| 03 | LRBRL `BOE-A-1985-5392` | `indice`, bloque | `a22` | H4 | Artículo 22 |
| 04 | LGT `BOE-A-2003-23186` | `indice`, bloque | `a66` | H5 | Artículo 66 |
| 05 | TRLRHL `BOE-A-2004-4214` | `indice`, bloque | `a59` | H5 | Artículo 59 |
| 06 | LIRPF `BOE-A-2006-20764` (`reproduce: boe-fiscal`) | `indice`, bloque | `a17` | H5 | Artículo 17 |
| 07 | LRJSP `BOE-A-2015-10566` | `indice`, bloque | `a25` | H5 | Artículo 25 |
| 08 | LTAIBG `BOE-A-2013-12887` | `indice`, bloque | `a20` | H5 | Artículo 20 |
| 09 | CE `BOE-A-1978-31229` | bloque | `a140` | H5 | Artículo 140 |
| 10 | ET `BOE-A-2015-11430` | `indice`, bloque | `a38` | H5 | Artículo 38 |
| 11 | — (`activa: false`) | — | — | — | — |
| 12 | — (`activa: false`) | — | — | — | — |

La 06 reproduce los usos `indice BOE-A-2006-20764` y `articulo BOE-A-2006-20764 a17` que documenta `refs/boe.py` en
sus líneas 8-9. Ninguna eval nombra un municipio ni declara un comportamiento propio de uno: la 03 y la 05 hablan de
«ayuntamiento» en general (FR-066). Cada fichero lleva una línea de comentario con su tipo, igual que los ejemplos del
contrato §1.

No faltó nada por grabar: toda consulta necesaria de las doce evals tiene su respuesta en la unión de H4 y H5.

## `TestEvalsDelRepositorio`

- La prueba lee `../../evals/boe-legislacion` y `../../data/normas.yaml`. No escribe el nombre de ningún fichero de eval.
- Una sola llamada a `ComprobarConjuntoDeBoeLegislacion` sirve a dos subtests: `conjunto` presenta los defectos de
  todas las reglas salvo `normas conocidas`, y `normas-conocidas` presenta solo los de esa regla.
- `formato` exige además al menos las doce evals bien formadas (plan, obligación 12: sin pasar en vacío).
- `normas-conocidas` comprueba que la regla existe con ese nombre en `reglasDelConjunto`. Si se renombrara, el subtest
  no vería sus defectos y pasaría en vacío.
- `grabado` exige que haya alguna consulta necesaria antes de llamar a `Preparar` y a `ComprobarSinRed`.

## Verificación

- **Rojo.** Sin el directorio, fallaban `formato` (no se puede listar), `conjunto` (los defectos de un conjunto
  vacío) y `grabado` (ninguna consulta).
- **Verde.** Con las doce evals pasan los cuatro subtests, también con `-race`.
- **Sonda positiva, ya deshecha.** Se cambió la norma de la cita de la eval 09 a `BOE-A-2099-99999`. `formato` y
  `conjunto` siguieron en verde. `normas-conocidas` y `grabado` fallaron, nombrando
  `09-constitucion-articulo-140.yaml` y la cita `BOE-A-2099-99999 a140` sin su bloque.
- **Lint y CI.** `golangci-lint` de `tools/` sobre `internal/evals`: 0 incidencias. `make ci` sale con 0.
