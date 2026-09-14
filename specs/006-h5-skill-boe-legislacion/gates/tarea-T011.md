# T011 · intento 1 · en verde

## De dónde sale cada valor de `data/normas.yaml`

Identificador, título y rango (`rango.texto`) se copiaron, tal cual, del resultado que cada entrada del manifiesto
resuelve en su búsqueda grabada, leído con `jq`. Los diez identificadores coinciden con el registro de la grabación
(mensaje de `1f4feea`). Cada título se contrastó además con los metadatos grabados de la norma.

| Abreviatura | Identificador | Rango | Búsqueda grabada | Metadatos |
|---|---|---|---|---|
| LPAC | `BOE-A-2015-10565` | Ley | H4, `…-8b1e1998` | H4 |
| LCSP | `BOE-A-2017-12902` | Ley | H5, `…-dd1a9c42` | H4 |
| LRBRL | `BOE-A-1985-5392` | Ley | H5, `…-fc01fd10` | H4 |
| LGT | `BOE-A-2003-23186` | Ley | H5, `…-6f2a4a94` | H5 |
| TRLRHL | `BOE-A-2004-4214` | Real Decreto Legislativo | H5, `…-d6c87d55` | H5 |
| LIRPF | `BOE-A-2006-20764` | Ley | H5, `…-238812af` | H5 |
| LRJSP | `BOE-A-2015-10566` | Ley | H5, `…-034b1f50` | H5 |
| LTAIBG | `BOE-A-2013-12887` | Ley | H5, `…-b066f008` | H5 |
| CE | `BOE-A-1978-31229` | Constitución | H5, `…-05b220a9` | H5 |
| ET | `BOE-A-2015-11430` | Real Decreto Legislativo | H5, `…-5cc376c5` | H5 |

Los títulos solo llevan caracteres latinos precompuestos (é, í, ñ, ó, ú), sin espacios al final ni nada que escapar
entre comillas dobles. Abreviaturas y materias son las de la tabla §4 del contrato normas-y-referencias.

## El título de la LRBRL

La búsqueda grabada y los metadatos de H4 dan «Ley 7/1985, de 2 de abril, Reguladora de las Bases del Régimen Local»:
«Reguladora» con mayúscula y sin punto final. Dos textos escritos antes de la grabación decían «reguladora … Local.»:

- el ejemplo de §5 del contrato normas-y-referencias;
- la constante sintética `tituloDeLaLRBRL` de `internal/skills/normas_test.go` (T010), cuyo comentario dice «tal como
  los dan los metadatos grabados en H4».

`data/normas.yaml` lleva el título grabado. La constante se corrigió a ese mismo título: el fichero está declarado en
esta tarea y ningún caso de `TestLeerNormas` depende de esos bytes. El ejemplo de §5 solo ilustra el formato de
`references/normas.md` y no se ha tocado. La tarea de `TestRenderizarNormas` no debe tomar de él el título real.

## Verificación

- **Rojo.** `TestNormasDelRepositorio` y `TestIdentificadoresDeLasNormas` fallaban al abrir `../../data/normas.yaml`.
- **Verde.** La verificación pasa sobre la tabla y el manifiesto reales, con las diez normas. Los cuatro subtests
  negativos exigen el mensaje exacto con `EqualError`:
  - `identificador-cambiado` usa `BOE-A-2099-99999`;
  - `titulo-cambiado` quita el punto final del título de la LPAC;
  - `norma-sin-entrada` retira del manifiesto la entrada de la LPAC;
  - `norma-en-la-busqueda-de-otra-entrada` saca la norma de la búsqueda reproducida de la entrada de la LPAC. Toma el
    primer resultado que no está en la tabla; en la respuesta grabada es `BOE-A-1992-26318`, la Ley 30/1992.
- **Lint y CI.** `golangci-lint` de `tools/` sobre `internal/evals` e `internal/skills`, con las etiquetas de la
  configuración: 0 incidencias. `go vet -tags grabacion,integration,fuentes` compila junto al arnés de grabación.
  `make ci` sale con 0: «ci: todos los controles en verde».
