# Informe del hito H30 · Las evals miden solo lo que cambia: dependencias de cada eval y resultados por huella

Generado por el workflow `hito` el 2026-10-22T22:13:23Z, sobre `2c8dda6` de `026-h30-las-evals-miden`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/026-h30-las-evals-miden/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `2c8dda6` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (4 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 19 umbrales, los 11 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); legal-core: 18 umbrales, los 12 que deciden se cumplen y 6 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `lectura_sin_intentar:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`, `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta`) (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-1`), 2 rondas.
- **Sesiones con otro modelo que el de su rol**: ninguna (47 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 18 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 24 commits; 107 files changed, 20919 insertions(+), 531 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_spec: [c][h] FR-024 no decía si el límite sale de la cifra del ADR o del `umbral` de `clases.yaml` → la tasa de referencia se deriva, con la regla del ADR, del `umbral` vigente de `clases.yaml` y de las respuestas de la ejecución completa; el ADR escribe las cifras de hoy (≈0,85 % y ≈0,36 %) como resultado; estrechar `clases.yaml` estrecha el límite para todo *n* (ADR 0038, punto 4), y el test de SC-010 usa la tasa derivada, con un mutante que baja el umbral de `boe-legislacion` a 2 de… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: [i] la búsqueda de medidas de `main` no tenía cota y buscaba también para las evals elegidas → las elegidas se miden sin buscar (entrega (5), «mide lo elegido»), y la búsqueda recorre como mucho las 10 medidas de `main` más recientes, solo para las no elegidas; al llegar a la cota, lo que queda se mide. Peor caso: 10 medidas × 3 informes, unos 18 MiB con los informes de H26.1; lo normal, una medida. Con esto desaparece la reutilización de una eval elegida por una reversión.
- T007: contracts/informe-del-job.md §2 da `evals[].eval` como «el mismo nombre que `eval` en `tasas`» y su ejemplo lo escribe sin `.yaml`, pero en `tasas` va el fichero con `.yaml` → `origen.evals[].eval` es el fichero con `.yaml`, el de `tasas`, para que la búsqueda de T009 case cada eval con sus series sin traducir nombres; la decisión de T010 tiene que nombrar así sus evals.
- T007: sin decisión (el sondeo, y `TestEjecucionDelJob` sin `-decision` hasta T010 y T013) el informe no sabe huellas ni motivos → su `origen` es el de la suite completa: `modo: completa`, cada eval bien formada medida con `suite_completa`, en orden de fichero y sin `huella` (EscribirInforme no calcula conjuntos); los ficheros mal formados siguen solo en `ficheros_mal_formados`. En el del instrumento sin medir, el `origen` recibido sin ningún aporte y con 0 sesiones abiertas (sin decisión, sin ev… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: el contrato no fija dónde van ni cómo se escriben → `origen` va en informe.json detrás de `motivos`; la sección «Medidas y reutilizadas» va en informe.md delante de «Tasas por eval», con «Evals medidas: n», «Evals reutilizadas: n» y «Sesiones abiertas: n», el error de la plataforma, la lista y «Series reutilizadas:» con los encabezados de la tabla de las tasas (o «Series reutilizadas: ninguna»); el «commit corto» son sus 7 primeras cifras, las de git; la línea de `sin_conjunto` es «medida:… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: contracts/informe-del-job.md §3 solo da la descripción de `sin_activar:<modelo>:<modo>:reutilizadas` → la de todo `:reutilizadas` se deriva de la del original sin ningún caso por umbral: su primera palabra seguida de «reutilizadas» y el resto, sin lo que sigue a su último «, sobre » (para sin_activar, la del ejemplo); su `umbral` es el que publica el original (data-model §9: «el mismo `umbral`»), también en el conjunto de una clase, que no decide.
- T007: §5 pide un motivo por serie reutilizada que decide y no pasa sin fijar su texto → «<eval> con <modelo>, reutilizada: pasan <p> de <s>, y el umbral es <u>», el de una serie medida con «reutilizada» detrás del modelo, detrás de los motivos de las series medidas; el de un `:reutilizadas` que decide sin cumplirse es el de cualquier umbral, sin respuestas nombradas (sus sesiones no están en el informe).
- T009: la tarea pide la marca «El informe del trabajo está escrito» en el trabajo «de la skill», y FR-010 y data-model §7 en «cada trabajo `evals (<skill>)`» («una a la que le falta el informe de alguna skill» no es medida) → es medida de `main` la ejecución que tiene el trabajo de la skill buscada y en la que todos sus trabajos `evals (<skill>)` terminaron en `success` o `failure` con la marca en `success` (la lectura de FR-010, la que menos reutiliza); de ella solo se lee el informe de la skill… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T009: «el de la primera medida con su misma huella» no dice qué pasa si el resultado de esa medida está incompleto o su procedencia es de 90 días o más → la búsqueda sigue con la medida anterior, y la eval se mide solo si ninguna de las recorridas le da un resultado reutilizable (FR-013: «sin ninguna medida de `main` con su huella … cuyo resultado se pueda leer y esté completo y cuya procedencia tenga menos de 90 días»; data-model §8: «la búsqueda sigue con la medida anterior»). Una respuesta si… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T009: contracts/informe-del-job.md §2 da `error_de_la_plataforma` como «la orden» y «su salida de error» sin decir si la orden lleva sus valores ni qué se escribe si la salida de error está vacía o si gh terminó con 0 y su respuesta no se pudo leer → `orden` es el texto constante de la orden seguido de «, con NOMBRE=valor, …», como la nombra el error de T008; `error` es su salida de error sin los blancos de los extremos o, si no escribió nada, cómo terminó (`exit status 1`); y una respuesta que … (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T010: data-model §5 y contracts/job-de-evals.md §2 dicen que un error del cálculo de los conjuntos elige todas con `sin_conjunto` (`elige=si`), sin distinguir la causa de FR-005 de una eval (un applet sin fichero, una referencia sin cabecera, un paquete con error…) de un fallo del propio checkout (`go list`, `git ls-files`, `git merge-base`, `go.mod` ilegible) → solo la causa de FR-005 de una eval elige todas con `sin_conjunto`, con la causa en `origen.sin_conjunto`; un fallo del checkout hace f… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T013: el contrato (§3) dice que el guion de las evals pasa `DECISION_DE_EVALS` a `TestEjecucionDelJob` como `-decision`, sin fijar qué pasa con una ruta relativa, que el guion comprueba desde la raíz del repositorio (hace `cd` a ella al empezar) y que go test leería desde `internal/evals/` → el guion la comprueba como la ve y pasa a `-decision` esa misma ruta hecha absoluta (`pwd -P` de su directorio, como ya hace con la carpeta de salida), de modo que la comprobación y la lectura miran el mismo… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T014: contracts/release.md §2 fija la línea de un trabajo en `failure` y la de un informe con reutilización, y no el texto de los demás motivos ni el de una ejecución sin terminar (data-model §11 la da como filtro, «terminadas») → una ejecución sin terminar sobre el commit se mira y su línea dice `no ha terminado (status <status>)`, en lugar de quedar fuera, para que el mensaje no diga «ninguna ejecución del flujo evals sobre el commit» con una en curso; y los demás motivos, en el orden de §11, … (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T014: el contrato no dice qué hace la comprobación si una lectura de la plataforma falla → la lista de ejecuciones, los trabajos de una ejecución o el registro de un trabajo con su orden de gh fallida, o con el contexto agotado, hacen fallar la comprobación con ese error envuelto (`release: la comprobación de la suite completa del commit <c> no pudo leer la plataforma: …`) y sin la orden que lanza la suite completa, porque lo que falta es la lectura y no la ejecución (como en la búsqueda de medi… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_spec: [h] FR-001(d) no decía qué paquetes son «los de un applet» y `go list -deps ./internal/app` arrastraba el paquete raíz con todas las skills → los paquetes de un applet son los del repositorio que importa su código de la composición (hoy su fichero de `internal/app`) fuera de lo común, con sus `go list -deps`; lo común no entra por (d) y el paquete raíz solo llega por (b), así que el cuerpo de `skills/jurisprudencia/SKILL.md` elige sus 10 evals, como piden FR-003 y la Aceptación d… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: [h] `_test.go` y `testdata/` de los paquetes de un applet → no son de ningún conjunto ni rutas que las evals miden, porque no entran en el binario ni en las sesiones; salvo las grabaciones de FR-001(e), que el arnés sirve (p. ej., `internal/source/boe/testdata/`). Los de lo común (el arnés de `internal/evals`) siguen siendo común.
- corrector_spec: [a][h] `go.mod` y `go.sum` quedaban fuera, contra las entregas (1) y (4) del hito → la versión de cada módulo externo de los paquetes de un conjunto y de lo común, y la de Go (`go` y `toolchain`), entran en el conjunto, en lo común y en la huella; `go.mod` y `go.sum` pasan a ser rutas que las evals miden (FR-004) y salen de «Fuera de alcance»; los `tools/*/go.mod`, que no entran en el binario, siguen fuera.
- corrector_plan: [f][n] el plan sacaba de las rutas que miden los `_test.go` y `testdata/` de todos los paquetes de `cmd/` e `internal/`, contra FR-001(d), FR-002 y FR-004 → solo se sacan los de los paquetes que entran en algún conjunto por FR-001(d), salvo las raíces de grabaciones; los de un paquete de lo común son rutas de lo común y los de un paquete fuera de todo conjunto, rutas sin asignar con su línea en la huella: un cambio solo en un test de `internal/cli` o en un guion de `internal/app/… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_plan: [m] un resultado reutilizado pasaba de una medida de `main` a la siguiente sin límite, con una procedencia cuyo registro ya no existe → solo es reutilizable si la ejecución en que se abrieron sus sesiones tiene menos de 90 días antes de la decisión (la vida del registro que da el spec); si no, la eval se mide con `sin_resultado_de_main`, y una medida de `main` de 90 días o más no se recorre. Escrito en FR-012 (con su criterio), FR-013, el caso límite de FR-011 y «Uso» del spec; c… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_plan: [g] un error de la plataforma o el plazo agotado se callaban, y no se decía qué hacer sin los ficheros de la propuesta → la orden de `gh` y su salida de error (o el plazo agotado, con la orden en curso) van al registro y a `origen.error_de_la_plataforma`, que el informe copia; sin los ficheros de la propuesta se eligen todas con el motivo `plataforma`, sin rutas (lectura conservadora, como FR-005), y `cambios` da `elige=si`; en la búsqueda, la primera orden que falla la para y lo… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T004: FR-001(d) y research D3 no dicen si el cierre de un applet atraviesa los paquetes de lo común («del cierre entran los de cada paquete que no es de lo común») → el cierre no entra en ningún paquete de lo común ni en el raíz, como el de lo común no entra en uno de un conjunto (research D4): un paquete del módulo al que solo se llega a través de lo común no es de ningún conjunto y es ruta sin asignar, que elige todas (se mide de más, nunca de menos). Hoy no cambia ningún conjunto: lo común no… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: la tarea dice que el modelo del juez cambia la huella de «todas», y FR-009 («la huella del juez si su skill lo tiene») y contracts/eleccion-y-huella.md §4 («`juez`: solo si la skill de la eval tiene `evals/<skill>/juez/`») la ponen solo en las evals de una skill con juez → la línea `juez` va solo en esas: el modelo del juez y su versión de Claude Code cambian las de `boe-legislacion` y `jurisprudencia` y no las de `legal-core`, que el juez no juzga («si y solo si cambia algo que la elige»)… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T008: `informe_del_trabajo` de `comun.sh` descarta además, entre las marcas, las líneas que no empiezan por espacio o llave (las de otra salida) y comprueba que el informe es JSON; la tarea solo enumera quitar trabajo, paso, hora y marca de orden de bytes y tomar lo que hay entre las marcas → no se descartan líneas ni se lee el JSON aquí: un informe con una línea ajena no se puede leer y su medida no aporta nada (se mide, FR-013), y el paso «Ejecutar las evals» lleva ya `2>&1`, que es lo que evi… (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- T016: el ADR 0045 precisa el punto 3 del ADR 0038, cuyo estado, como el de los ADR precisados por otro, lo diría, y la decisión cerrada «Umbral de una clase del juez» de `CLAUDE.md` lo nombraría; ni `docs/ADR/0038-…` ni `CLAUDE.md` están en las rutas de la tarea → no se tocan: el ADR 0045 los deja en «Enmiendas que provoca», fuera de las rutas del hito, para quien lo acepte.
- clarify Q2 (criterio d, conservadora): Basta la granularidad por skill. En FR-001(b), «la referencia de su materia» son hoy todas las `references/` de las skills de la eval, porque ninguna eval ni ninguna referencia es de una materia. H30 no define ninguna marca de materia, ni en las evals, ni en las referencias, ni en `data/`. Tampoco deja un punto de extensión. Cómo se sabe la materia de una eval y de una referencia lo fija el molde de vertical de H32 («qué trae un paquete de materia: normas, … (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q4 (criterio c): La tasa de referencia de cada skill con juez es la mayor tasa con la que el umbral vigente de su ejecución completa cumple aún 1/50. Se deriva de ese umbral y del número de respuestas de la ejecución completa de hoy, y el ADR la escribe como cifra fija, con su derivación y redondeada hacia abajo. El test de SC-010 usa esa cifra (SC-013). Con las cifras de hoy, calculadas en esta sesión con la binomial exacta: `boe-legislacion`, 3 marcadas de 120 respuestas (umbral 0,028 … (entera en `specs/026-h30-las-evals-miden/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run

- 2026-10-22T20:04:05Z · «docs(H30): barrido de artefactos y documentación contra el producto»: make ci quedó en rojo; el cambio se apartó en specs/026-h30-las-evals-miden/gates/aparcado-1.patch y el hito vuelve a f660303.
- 2026-10-22T21:35:45Z · La ronda 2 de la revisión final (ciclo 1) aprobó con los dos jueces lo que quedaba en revision-pendiente.md; se conserva como specs/026-h30-las-evals-miden/gates/revision-pendiente-resuelto-r2.md.

- Observaciones de los jueces, por debajo del umbral y sin corregir: 8 en `spec-r2.json`, 10 en `plan-r2.json`, 3 en `tasks-r3.json`, 7 en `revision-a-r2.json`, 7 en `revision-b-r2.json`.

### Internos (31)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (1), T003 (1), T004 (3), T005 (3), T006 (4), T007 (1), T008 (2), T010 (3), T011 (2), T012 (2), T013 (2), T015 (1), T016 (3), T017 (1), barrido (1), corrector_spec (1). Enteras en `specs/026-h30-las-evals-miden/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `2c8dda6`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 2/3 | — | — |  |
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 2 de 60 (3,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 120 (0,0 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 472 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 482 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 88 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 89 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 66 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 2 de 42 (4,8 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 3 de 66 (4,5 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 16 de 42 (38,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

**jurisprudencia**: aprobado sobre `2c8dda6`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `03-por-materia.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:orden` | 3 de 30 (10,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 2 de 30 (6,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 37 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 39 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |

**legal-core**: aprobado sobre `2c8dda6`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 0/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `vigencia_sin_declarar:claude-sonnet-5-5:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `plazo_sin_fecha:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `vigencia_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 9 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 50 de 33 (151,5 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:orden` | 2 de 3 (66,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 38 de 33 (115,2 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta` | 2 de 3 (66,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de boe-legislacion** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `27aeccc` (árbol de `skills/boe-legislacion`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H24-1 | 2026-10-06 | `a5bda45` | `6d65aa3` | 1/54 | 0/54 | 1/78 | 0/30 | 1/108 | no: otra skill |
| H24-2 | 2026-10-06 | `b0a7a9c` | `75e1d04` | 0/54 | 3/54 | 0/78 | 3/30 | 3/108 | no: 1 respuesta sin juzgar |
| H24-3 | 2026-10-06 | `3107ae2` | `6063317` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | no: otra skill |
| H23-1 | 2026-10-07 | `bf4befd` | `6063317` | 0/54 | 0/50 | 0/76 | 0/28 | 0/104 | no: otra skill |
| H25-1 | 2026-10-10 | `3f99446` | `6063317` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `6063317` | 1/54 | 1/54 | 0/78 | 2/30 | 2/108 | no: otra skill |
| H25-3 | 2026-10-10 | `d51f3d7` | `eed129f` | 0/54 | 2/54 | 0/78 | 2/30 | 2/108 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `eed129f` | 0/54 | 2/54 | 1/78 | 1/30 | 2/108 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `eed129f` | 0/54 | 3/54 | 1/78 | 2/30 | 3/108 | no: otra skill |
| H9-1 | 2026-10-18 | `3f14e3a` | `eed129f` | 0/54 | 1/54 | 0/78 | 1/30 | 1/108 | no: otra skill |
| H9-2 | 2026-10-18 | `92fc232` | `eed129f` | 0/54 | 0/54 | 0/78 | 0/30 | 0/108 | no: otra skill |
| H20-1 | 2026-10-19 | `77211a9` | `49b59ac` | 0/60 | 2/60 | 0/90 | 2/30 | 2/120 | no: otra skill |
| H26.1-1 | 2026-10-21 | `444e674` | `27aeccc` | 0/60 | 1/60 | 0/90 | 1/30 | 1/120 | sí |
| H30-1 | 2026-10-22 | `2c8dda6` | `27aeccc` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |

Tasa acumulada con la skill como está: 1 de 240 (0,42 %), de H26.1-1 y H30-1.
Insuficiente: 240 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `3bbd9cd` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H25-1 | 2026-10-10 | `3f99446` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H25-2 | 2026-10-10 | `e380f0a` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: 1 respuesta sin juzgar |
| H25-3 | 2026-10-10 | `d51f3d7` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26-1 | 2026-10-13 | `b29fe6c` | `9480bbc` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H23.1-1 | 2026-10-16 | `60e8678` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H9-1 | 2026-10-18 | `3f14e3a` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H9-2 | 2026-10-18 | `92fc232` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H20-1 | 2026-10-19 | `77211a9` | `4759aa4` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H26.1-1 | 2026-10-21 | `444e674` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H30-1 | 2026-10-22 | `2c8dda6` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 120 (0,00 %), de H26.1-1 y H30-1.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

Nada: el hito no toca fuentes, anomalías, adaptadores, fixtures, esquemas, grabaciones ni evidencias.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T003(hecha) T004(hecha) | — | comprobado por su control | `ci:internal/evals/dependencias_test.go:TestConjuntoDeCadaEval`, en make ci (verde) |
| FR-002 | T004(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde) |
| FR-003 | T004(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde) |
| FR-004 | T004(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde) |
| FR-005 | T003(hecha) T004(hecha) T005(hecha) | — | comprobado por su control | `ci:internal/evals/dependencias_test.go:TestConjuntoDeCadaEval`, en make ci (verde) |
| FR-006 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde) |
| FR-007 | T005(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde) |
| FR-008 | T005(hecha) T011(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestMatrizDeLasPropuestas`, en make ci (verde) |
| FR-009 | T006(hecha) | — | comprobado por su control | `ci:internal/evals/huella_test.go:TestHuellaDeLasEvals`, en make ci (verde) |
| FR-010 | T008(hecha) T009(hecha) T013(hecha) | — | comprobado por su control | `ci:internal/evals/reutilizacion_test.go:TestBusquedaDeMedidasDeMain`, en make ci (verde) |
| FR-011 | T007(hecha) T009(hecha) | — | comprobado por su control | `ci:internal/evals/reutilizacion_test.go:TestBusquedaDeMedidasDeMain`, en make ci (verde) |
| FR-012 | T009(hecha) T010(hecha) T013(hecha) | — | comprobado por su control | `ci:internal/evals/reutilizacion_test.go:TestBusquedaDeMedidasDeMain`, en make ci (verde) |
| FR-013 | T009(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/reutilizacion_test.go:TestBusquedaDeMedidasDeMain`, en make ci (verde); `ci:internal/evals/decision_test.go:TestDecisionDelJob`, en make ci (verde) |
| FR-014 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/decision_test.go:TestDecisionDelJob`, en make ci (verde) |
| FR-015 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConReutilizadas`, en make ci (verde) |
| FR-016 | T010(hecha) T013(hecha) T017(hecha) | — | comprobado por su control | `ci:internal/evals/decision_test.go:TestDecisionDelJob`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-017 | T010(hecha) T013(hecha) T017(hecha) | — | comprobado por su control | `ci:internal/evals/decision_test.go:TestDecisionDelJob`, en make ci (verde) |
| FR-018 | T012(hecha) T013(hecha) T017(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-019 | T010(hecha) T011(hecha) T013(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestMatrizDeLasPropuestas`, en make ci (verde) |
| FR-020 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-021 | T012(hecha) | — | comprobado por su control | `ci:internal/evals/tanda_test.go:TestTandaConLaSuiteCompleta`, en make ci (verde) |
| FR-022 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeConReutilizadas`, en make ci (verde) |
| FR-023 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/serie_test.go:TestFilaConReutilizadas`, en make ci (verde) |
| FR-024 | T002(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/evals/tasa_test.go:TestLimiteDeLaTasa`, en make ci (verde) |
| FR-025 | T008(hecha) T014(hecha) T015(hecha) T017(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde); `ci:internal/evals/suite_completa_test.go:TestComprobarLaSuiteCompleta`, en make ci (verde) |
| FR-026 | T001(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeDeLaSuiteCompletaNoCambia`, en make ci (verde) |
| FR-027 | T018(hecha) | — | tareas hechas | — |
| SC-001 | T018(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde) |
| SC-002 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/dependencias_test.go:TestConjuntoDeCadaEval`, en make ci (verde) |
| SC-003 | T005(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes`, en make ci (verde) |
| SC-004 | T011(hecha) | — | comprobado por su control | `ci:internal/evals/eleccion_test.go:TestMatrizDeLasPropuestas`, en make ci (verde) |
| SC-005 | T014(hecha) T015(hecha) | — | comprobado por su control | `ci:release_test.go:TestConfiguracionDeLaRelease`, en make ci (verde); `ci:internal/evals/suite_completa_test.go:TestComprobarLaSuiteCompleta`, en make ci (verde) |
| SC-006 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/decision_test.go:TestDecisionDelJob`, en make ci (verde) |
| SC-007 | T010(hecha) | — | comprobado por su control | `ci:internal/evals/decision_test.go:TestDecisionDelJob`, en make ci (verde) |
| SC-008 | T006(hecha) | — | comprobado por su control | `ci:internal/evals/huella_test.go:TestHuellaDeLasEvals`, en make ci (verde) |
| SC-009 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/serie_test.go:TestFilaConReutilizadas`, en make ci (verde) |
| SC-010 | T002(hecha) | — | comprobado por su control | `ci:internal/evals/tasa_test.go:TestLimiteDeLaTasa`, en make ci (verde) |
| SC-011 | T013(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| SC-012 | T001(hecha) T007(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestInformeDeLaSuiteCompletaNoCambia`, en make ci (verde) |
| SC-013 | T002(hecha) T016(hecha) | — | comprobado por su control | `ci:internal/evals/tasa_test.go:TestLimiteDeLaTasa`, en make ci (verde) |
| SC-014 | T018(hecha) | — | tareas hechas | — |
| SC-015 | T018(hecha) | — | tareas hechas | — |
| SC-016 | T013(hecha) T017(hecha) | — | tareas hechas | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`f660303`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `cc977b6` fix(H30): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/ADR/0045-lo-reutilizado-no-cuenta-en-la-tasa.md`, `docs/WORKFLOW.md`, `internal/evals/decision_test.go`, `internal/evals/reutilizacion_test.go`, `cierre.md`, `contracts/eleccion-y-huella.md`, `data-model.md`, `plan.md`, `quickstart.md`.
- `8c7d0ff` docs(H30): veredictos de la revisión final: solo registros de `gates/`.
- `2c8dda6` docs(H30): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/026-h30-las-evals-miden/quickstart.md`. Suite de aceptación congelada: `specs/026-h30-las-evals-miden/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `2aff4d2e`: 14 h 35 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 2aff4d2e`.
