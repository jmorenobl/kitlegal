# Contrato: lo que dicen los documentos del umbral

Qué pasaje de cada documento deja falso el hito y qué tiene que decir después (FR-050 a FR-055). Las líneas son las de
`869eaf8`; el texto final lo escribe la tarea. Se corrige lo que aquí se nombra y **nada más**: no se reescribe lo que
sigue siendo cierto, no se promete ninguna release y ningún test ata este texto al código. Lo comprueba la revisión
final, por lectura contra el árbol (SC-013).

Las cifras, todas del árbol o de [umbrales-del-informe](./umbrales-del-informe.md): trece elementos y nueve que
deciden en cada una de las dos skills; 108 y 60 respuestas juzgadas; `umbral` 0.028 y 0.017, que son como mucho 3 de 108
y 1 de 60; el nombre `afirma_lo_no_leido:<modelo>`; y el motivo de §4 de ese contrato.

## `CONTRIBUTING.md`

| Líneas | Dice hoy | Tiene que decir |
|---|---|---|
| 724, 727-728 | En las dos skills, `afirma_lo_no_leido` «**decide**, con umbral 0» | Que decide con la tasa de su `clases.yaml` sobre las respuestas de los dos modos juntas: 0.028 en `boe-legislacion`, como mucho 3 de 108, y 0.017 en `jurisprudencia`, como mucho 1 de 60 (ADR 0038). Lo que cuenta en la clase no cambia |
| 1105-1108 | El motivo de una clase del juez, con el ejemplo `umbral afirma_lo_no_leido:claude-sonnet-5-5:herramienta: 1 de 54 (1,9 %), y tiene que ser ≤ 0,0 %: <sesión>: «…»` | El del umbral conjunto, con cada respuesta marcada de todos los modos y su modo: `umbral afirma_lo_no_leido:claude-sonnet-5-5: 4 de 108 (3,7 %), y tiene que ser ≤ 2,8 %: <sesión> (modo <modo>): «<frase 1>» · «<frase 2>» · «<frase 3>»`, y `; <sesión> (modo <modo>): …` por cada otra |
| 1134-1137 | Los umbrales de un modo se miden sobre las sesiones de ese modo «sin sumar las del otro»; «En `boe-legislacion` son doce, diez de ellos decidiendo»; «en `jurisprudencia`, otros doce» | Que eso vale para todos menos para el umbral conjunto de una clase del juez que decide, que suma los modos; trece, nueve de ellos decidiendo; en `jurisprudencia`, otros trece |
| 1142 | Fila `afirma_lo_no_leido:<modelo>:<modo>`, con `umbral` `0` y `decide` sí | `umbral` `0.028` y `decide` no. Y una fila nueva, detrás de la de `cuenta_su_proceso`: `afirma_lo_no_leido:<modelo>`, con las marcadas y las juzgadas de los dos modos (108), `0.028` y sí |
| 1149-1151 | El orden: los tres de cada modo, los dos de la medida, las duraciones | Detrás de los de los dos modos y delante de los de la medida, el conjunto de cada clase que decide |
| 1160 | «En `jurisprudencia` son también doce, y deciden diez» | Trece, y deciden nueve |
| 1167 | Fila `afirma_lo_no_leido:<modelo>:<modo>`, con `0` y sí | `0.017` y no. Y la fila nueva `afirma_lo_no_leido:<modelo>`, detrás de la de `cita_sin_documento`: los dos modos (60), `0.017` y sí |
| 1174-1177 | El orden; «El único que no decide es `afirma_que_existe`»; «Una respuesta marcada en `afirma_lo_no_leido` en un modo da `fallo` aunque en el otro no haya ninguna» | El orden con el conjunto; no deciden `afirma_que_existe` ni el recuento por modo de `afirma_lo_no_leido`; una marcada no da `fallo` por sí sola y dos, sí, estén en el modo que estén |

No cambia lo que dice del sondeo (1262-1280) ni el ejemplo del motivo de `cita_sin_documento` (1185).

## `docs/JURISPRUDENCIA.md`

| Líneas | Dice hoy | Tiene que decir |
|---|---|---|
| 240-243, en «Lo que H25 deja hecho» | Que lo decide el juez «con la clase `afirma_lo_no_leido` y umbral 0 en cada modo», y que la respuesta que lo hace «deja el job en rojo» | Que lo decide el juez con `afirma_lo_no_leido` y una tasa sobre los dos modos juntos: como mucho 1 marcada de las 60 respuestas, y con 2 el job sale en rojo (ADR 0038). El resto del párrafo no cambia |

## `docs/WORKFLOW.md`

| Línea | Dice hoy | Tiene que decir |
|---|---|---|
| 125, en `medir_cierre` | «en `boe-legislacion` deciden diez: por cada modo, `sin_activar:<modelo>:<modo>`, `afirma_lo_no_leido:<modelo>:<modo>`, `duracion_de_las_sesiones:<modo>` y `duracion_del_juez:<modo>`, y los dos de la medida versionada […]; `cuenta_su_proceso:<modelo>:<modo>` solo se publica» | Que deciden nueve: por cada modo, `sin_activar`, `duracion_de_las_sesiones` y `duracion_del_juez`; `afirma_lo_no_leido:<modelo>`, sobre los dos modos juntos; y los dos de la medida. Que `afirma_lo_no_leido:<modelo>:<modo>` y `cuenta_su_proceso:<modelo>:<modo>` solo se publican |

## `CLAUDE.md`

| Línea | Dice hoy | Tiene que decir |
|---|---|---|
| 21, viñeta de H24 | «`afirma_lo_no_leido` decide con umbral 0 en cada modo» | Que decidía así hasta H26 y que desde H26 decide con la tasa del ADR 0038 sobre los dos modos juntos. Las cifras de su cierre (0 de 54 en cada modo, 212 de 212…) no cambian |
| 74, «Umbral de una clase del juez» | «**El job sigue con el umbral 0 hasta H26, que lo implementa.**» | Que el job la aplica desde H26: el umbral `<clase>:<modelo>` decide sobre las respuestas de los dos modos juntas, y el recuento de cada modo se publica |

No se tocan la viñeta de H25 (23), que cuenta su cierre —«dentro de los umbrales del ADR 0038, que el job no aplica
hasta H26» sigue siendo cierto de entonces—, la frase del siguiente hito ni una viñeta de H26: las escribe la persona
al fusionar.

## `CHANGELOG.md`, sección *Unreleased*

Una entrada nueva, en «Cambiado», para quien lee un informe del job (FR-050): que `afirma_lo_no_leido` decide con una
tasa sobre las respuestas de los dos modos juntas, con el umbral `afirma_lo_no_leido:<modelo>` y su valor en cada skill
(0.028, como mucho 3 de 108; 0.017, como mucho 1 de 60); que `afirma_lo_no_leido:<modelo>:<modo>` se sigue publicando,
con `decide: false`; que `umbrales` pasa de doce elementos con diez que deciden a trece con nueve en cada una de las dos
skills; y la forma del motivo.

Y, en las entradas de H24 y de H25, que siguen en *Unreleased*, lo que deja de valer:

| Líneas | Dice hoy | Tiene que decir |
|---|---|---|
| 129 | «`afirma_lo_no_leido` **decide** con 0» | Que decide con la tasa de su `clases.yaml` sobre los dos modos juntos |
| 140, 146 | «`umbrales` gana `afirma_lo_no_leido:<modelo>:<modo>` (0, decide)»; «`boe-legislacion` publica doce, diez de ellos decidiendo» | Que gana el recuento por modo, que se publica, y `afirma_lo_no_leido:<modelo>`, que decide; trece, nueve de ellos decidiendo |
| 214-215 | `afirma_lo_no_leido` de `jurisprudencia` «**decide** con 0» | Lo mismo, con 0.017 |
| 224-225, 231-233 | «**Doce umbrales, diez de ellos decidiendo.**», con `afirma_lo_no_leido:<modelo>:<modo>` «(0, decide)»; «Una respuesta marcada en `afirma_lo_no_leido` en un modo da `fallo`»; «Los doce de `boe-legislacion`» | Trece, nueve de ellos decidiendo, con el recuento por modo publicado y el conjunto decidiendo; una marcada de 60 no da `fallo` y dos, sí; los trece de `boe-legislacion` |

Las secciones de las versiones publicadas, de `[0.5.0]` hacia abajo, no cambian.

## Lo que no se toca

`docs/ROADMAP.md` —tampoco sus secciones de H24 y de H25—, ningún ADR, `docs/USO.md`, `README.md`, la web, y el spec,
el plan o los informes de un hito anterior. Tampoco el comentario de `.github/workflows/evals.yml` (líneas 152 a 154)
que dice «doce umbrales, diez que deciden»: queda falso a sabiendas y es un pendiente para la persona (spec, «Fuera de
alcance»).
