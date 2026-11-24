# Informe del hito H45 · El molde admite un paquete con sus plazos y su norma, sin código (ADR 0062)

Generado por el workflow `hito` el 2026-11-24T02:57:31Z, sobre `23db28c9` de `042-h45-el-molde-admite`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/042-h45-el-molde-admite/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige el fichero del hito en docs/roadmap/in-progress/ y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `23db28c` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: consulta-juridica aprobado, jurisprudencia aprobado, normativa aprobado, plazos-y-territorio aprobado (tasas en la sección 3).
- **Umbrales del job**: consulta-juridica: 9 umbrales, los 5 que deciden se cumplen y 4 solo se publican (4 por encima de su umbral: `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); jurisprudencia: 20 umbrales, los 12 que deciden se cumplen y 8 solo se publican (2 por encima de su umbral: `afirma_que_existe:claude-sonnet-5-5:orden`, `afirma_que_existe:claude-sonnet-5-5:herramienta`); normativa: 20 umbrales, los 12 que deciden se cumplen y 8 solo se publican (6 por encima de su umbral: `cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-sonnet-5-5:orden`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden`, `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta`, `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta`); plazos-y-territorio: 19 umbrales, los 13 que deciden se cumplen y 6 solo se publican (1 por encima de su umbral: `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta`) (sección 3).
- **Revisión final**: juez A aprobado (con `claude-opus-5-5`), juez B aprobado (con `claude-fable-5-2`), 2 rondas.
- **Sesiones con otro modelo que el de su rol**: ninguna (42 sesiones registradas, todas con el modelo que `workflow.yml` da a su rol).
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 18 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 26 commits; 109 files changed, 8316 insertions(+), 1899 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Alcance (lo que queda fuera o dentro del hito)**

- corrector_spec: [f][j] el umbral de SC-004 (0 tests de §C.5 en rojo con la copia de US4.1, aceptación 3 del hito) se medía solo en el run y quedaba en `mutantes.md` → lo hace cumplir una comprobación de la suite de aceptación, en `make ci`, que ejecuta los tests de §C.5 sobre la copia de `data/` y `evals/` de US4.1 y sale en rojo si alguno falla; `mutantes.md` sigue registrando su mutante (la lista literal devuelta a uno). Cómo corren sobre la copia los guiones de `internal/app` queda para el pl… (entera en `specs/042-h45-el-molde-admite/gates/supuestos.md` o `clarify-respuestas.json`)

**Skill (lo que pide, dice o comprueba una skill)**

- T005: contracts/sincronia-de-los-paquetes.md §3 y research D2 no dicen qué pasa con una fila cuya `norma` da el paquete y cuyo precepto es de una norma que no está en `data/normas.yaml` (con la regla anterior se saltaba y daba el defecto de `filasSinRegion`; ningún otro control cruza los preceptos de la cobertura con la tabla de normas) → como en `main`, corregido en la revisión final (línea de `corrector_revision` de abajo): la fila no es de ningún paquete, `filasDeLaCobertura` no la escribe y … (entera en `specs/042-h45-el-molde-admite/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_revision: [i][b] una fila de `regulados_aparte` con la `norma` de un paquete y su precepto en una norma que no está en `data/normas.yaml` salía en la referencia del paquete, sin defecto y con la cita sin el nombre de su norma, cuando en `main` daba el defecto de `filasSinRegion` → lectura conservadora, la de `main`: `paqueteDeLaFila` no da ningún paquete a esa fila antes de mirar la norma que la rige, con lo que `filasDeLaCobertura` la salta y `filasSinRegion` da el defecto de hoy con … (entera en `specs/042-h45-el-molde-admite/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 9 en `spec-r2.json`, 8 en `plan-r1.json`, 6 en `tasks-r2.json`, 6 en `revision-a-r2.json`, 8 en `revision-b-r2.json`.

### Internos (27)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (2), T003 (1), T004 (2), T005 (2), T006 (2), T008 (2), T009 (3), T010 (2), T011 (1), T012 (3), T013 (1), T015 (2), T016 (1), T018 (1), corrector_revision (1), corrector_spec (1). Enteras en `specs/042-h45-el-molde-admite/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**consulta-juridica**: aprobado sobre `23db28c`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-poliza-tras-el-fallecimiento.yaml` | 2/3 | 0/3 | 3/3 | 0/3 |  |
| `02-una-sola-capacidad.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `03-seguro-prorroga.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `04-despido-plazo.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 22 de 12 (183,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 14 de 12 (116,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 15 de 12 (125,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 16 de 12 (133,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `habla_en_primera_o_segunda_persona` | 0 de 1 (0,0 %) | ≤ 0,0 % | sí | sí |

**jurisprudencia**: aprobado sobre `23db28c`; decide `claude-sonnet-5-5` con 2 de 3; 10 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-resumen-sin-documento.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-por-materia.yaml` | 3/3 | 1/3 | 3/3 | 3/3 |  |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `07-resumen-de-una-conocida.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `08-doctrina-con-el-fallo-delante.yaml` | 3/3 | 2/3 | 3/3 | 3/3 |  |
| `09-de-que-trata-con-la-ficha-sola.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `10-doctrina-dada-por-hecha.yaml` | 3/3 | 0/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:orden` | 2 de 30 (6,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:orden` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 1,7 % | dentro | no: solo se publica, no es un control |
| `afirma_que_existe:claude-sonnet-5-5:herramienta` | 1 de 30 (3,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | sí | sí |
| `documento_sin_cotejar:claude-sonnet-5-5:herramienta` | 0 de 12 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 60 (0,0 %) | ≤ 1,7 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 125 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 124 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_del_juez:orden` | 62 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 58 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 30 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `habla_en_primera_o_segunda_persona` | 0 de 1 (0,0 %) | ≤ 0,0 % | sí | sí |

**normativa**: aprobado sobre `23db28c`; decide `claude-sonnet-5-5` con 2 de 3; 23 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 1/3 | 3/3 | 1/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 0/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 2/3 | 3/3 | 0/3 |  |
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
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `22-lcsp-contrato-menor-en-2019.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `23-lpac-articulo-21-en-2016.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 2 de 60 (3,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:orden` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 60 (0,0 %) | ≤ 2,8 % | dentro | no: solo se publica, no es un control |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 7 de 60 (11,7 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | 0 de 6 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5` | 0 de 120 (0,0 %) | ≤ 2,8 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 449 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 425 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 102 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 97 | ≤ 900 | sí | sí |
| `fuera_de_lo_grabado:claude-sonnet-5-5:orden` | 4 de 66 (6,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 22 de 42 (52,4 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 4 de 66 (6,1 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 34 de 42 (81,0 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `habla_en_primera_o_segunda_persona` | 0 de 1 (0,0 %) | ≤ 0,0 % | sí | sí |

**plazos-y-territorio**: aprobado sobre `23db28c`; decide `claude-sonnet-5-5` con 2 de 3; 12 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |
| `05-plazo-reposicion-leganes.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `06-plazo-reposicion-tordesillas.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `07-cobertura-liquidacion-tributaria.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `08-cobertura-multa-de-trafico.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-plazo-contencioso-agosto.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `10-plazo-reposicion-anterior-a-la-lpac.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `11-plazo-reposicion-sin-leer-el-precepto.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | 3/3 | 2/3 | 3/3 | 2/3 |  |

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
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:orden` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:orden` | 0 de 3 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `fuera_de_lo_grabado:claude-sonnet-5-5:herramienta` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-sonnet-5-5:herramienta` | 0 de 3 (0,0 %) | ≤ 0,0 % | sí | sí |
| `fuera_de_lo_grabado:claude-haiku-4-5-20251001:herramienta` | 0 de 33 (0,0 %) | ≤ 0,0 % | dentro | no: solo se publica, no es un control |
| `lectura_sin_intentar:claude-haiku-4-5-20251001:herramienta` | 1 de 3 (33,3 %) | ≤ 0,0 % | por encima | no: solo se publica, no es un control |
| `habla_en_primera_o_segunda_persona` | 0 de 1 (0,0 %) | ≤ 0,0 % | sí | sí |

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

**Serie de cierres de jurisprudencia** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `779f453` (árbol de `skills/jurisprudencia`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

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
| H26.1-1 | 2026-10-21 | `444e674` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H30-1 | 2026-10-22 | `2c8dda6` | `3bbd9cd` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H31-1 | 2026-10-25 | `03cf8d9` | `5a9198f` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H31.1-1 | 2026-10-27 | `a66e9d2` | `5a9198f` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H32-1 | 2026-10-30 | `7f8ef89` | `7dbf0c5` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H39-1 | 2026-11-03 | `a593631` | `7dbf0c5` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H40-1 | 2026-11-04 | `179bd21` | `7dbf0c5` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41-1 | 2026-11-06 | `69f5ecd` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41-2 | 2026-11-06 | `d5e6619` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.1-1 | 2026-11-10 | `e4915f6` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.2-1 | 2026-11-11 | `a27c276` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.2-2 | 2026-11-11 | `899708c` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.2-3 | 2026-11-11 | `e8715a9` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.3-1 | 2026-11-16 | `6701b62` | `b5b62e3` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | no: otra skill |
| H41.4-1 | 2026-11-17 | `4b073f8` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H33-1 | 2026-11-20 | `85f986f` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H38-1 | 2026-11-21 | `3473dd4` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H38-2 | 2026-11-21 | `c392ef1` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H36-1 | 2026-11-22 | `0c8f8de` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |
| H45-1 | 2026-11-24 | `23db28c` | `779f453` | 0/30 | 0/30 | 0/48 | 0/12 | 0/60 | sí |

Tasa acumulada con la skill como está: 0 de 360 (0,00 %), de H41.4-1, H33-1, H38-1, H38-2, H36-1 y H45-1.
Insuficiente: 360 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

**Serie de cierres de normativa** (ADR 0038, punto 6): respuestas de las evals que activan la skill que el juez marca con sus tres votos, por modo y por clase de pregunta.

Clave de la cabeza: skill `f0a4da8` (árbol de `skills/normativa`), decide `claude-sonnet-5-5`, juez `claude-opus-5-5` con Claude Code 2.1.289.

`afirma_lo_no_leido`:

| Medición | Fecha | Commit | Skill | Orden | Herramienta | Nombra | Por materia | Total | En la tasa |
|---|---|---|---|---|---|---|---|---|---|
| H32-1 | 2026-10-30 | `7f8ef89` | `8566858` | 0/60 | 2/60 | 2/90 | 0/30 | 2/120 | no: otra skill |
| H39-1 | 2026-11-03 | `a593631` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H40-1 | 2026-11-04 | `179bd21` | `afbcae6` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41-1 | 2026-11-06 | `69f5ecd` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41-2 | 2026-11-06 | `d5e6619` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41.1-1 | 2026-11-10 | `e4915f6` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41.2-1 | 2026-11-11 | `a27c276` | `26368b3` | 0/60 | 1/60 | 1/90 | 0/30 | 1/120 | no: otra skill |
| H41.2-2 | 2026-11-11 | `899708c` | `26368b3` | 1/60 | 0/60 | 1/90 | 0/30 | 1/120 | no: otra skill |
| H41.2-3 | 2026-11-11 | `e8715a9` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41.3-1 | 2026-11-16 | `6701b62` | `26368b3` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H41.4-1 | 2026-11-17 | `4b073f8` | `3f4260c` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H33-1 | 2026-11-20 | `85f986f` | `3f4260c` | 1/60 | 1/60 | 0/90 | 2/30 | 2/120 | no: otra skill |
| H38-1 | 2026-11-21 | `3473dd4` | `3f4260c` | 1/60 | 3/60 | 0/90 | 4/30 | 4/120 | no: otra skill |
| H38-2 | 2026-11-21 | `c392ef1` | `8358dd9` | 3/60 | 0/60 | 0/90 | 3/30 | 3/120 | no: otra skill |
| H36-1 | 2026-11-22 | `0c8f8de` | `8358dd9` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | no: otra skill |
| H45-1 | 2026-11-24 | `23db28c` | `f0a4da8` | 0/60 | 0/60 | 0/90 | 0/30 | 0/120 | sí |

Tasa acumulada con la skill como está: 0 de 120 (0,00 %), de H45-1.
Insuficiente: 120 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h32-paquetes.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h33-cobertura-instalada.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h36-plazos-de-hoy.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h41-4-skills-instaladas.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/skills/testdata/vertical-de-ensayo/data/cobertura.yaml`.
- Fixture o esquema EXISTENTE modificado: `internal/skills/testdata/vertical-de-ensayo/skills/plazos-y-territorio/references/ensayo.md`.
- Fixtures nuevos (2), por directorio:
  - `internal/skills/testdata/regiones/`: `plazos.yaml`
  - `internal/skills/testdata/vertical-de-ensayo/data/`: `plazos.yaml`
- Suite de aceptación activada y congelada (2 guiones, escritos desde el spec en la primera tarea): `h45-paquete-de-mas.txtar`, `h45-skills-instaladas.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T003(hecha) | paquete-de-mas.txtar  | tareas hechas | — |
| FR-002 | T001(hecha) T003(hecha) T006(hecha) | paquete-de-mas.txtar  | comprobado por su control | `ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo`, en make ci (verde) |
| FR-003 | T003(hecha) | — | comprobado por su control | `ci:internal/skills/regiones_test.go:TestRegiones`, en make ci (verde) |
| FR-004 | T003(hecha) | — | comprobado por su control | `ci:internal/skills/regiones_test.go:TestRegiones`, en make ci (verde) |
| FR-005 | T001(hecha) T006(hecha) | paquete-de-mas.txtar  | comprobado por su control | `ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo`, en make ci (verde) |
| FR-006 | T001(hecha) T006(hecha) | paquete-de-mas.txtar  | comprobado por su control | `ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo`, en make ci (verde) |
| FR-007 | T006(hecha) | — | tareas hechas | — |
| FR-008 | T003(hecha) T009(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/skills/regiones_test.go:TestRegiones`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| FR-009 | T001(hecha) T004(hecha) | skills-instaladas.txtar  | tareas hechas | — |
| FR-010 | T004(hecha) | — | tareas hechas | — |
| FR-011 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano`, en make ci (verde) |
| FR-012 | T004(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano`, en make ci (verde) |
| FR-013 | T004(hecha) | — | tareas hechas | — |
| FR-014 | T009(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| FR-015 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/grabaciones_test.go:TestGrabacionesSinSolape`, en make ci (verde); `ci:internal/evals/grabaciones_test.go:TestManifiestoDeGrabaciones`, en make ci (verde) |
| FR-016 | T001(hecha) T002(hecha) T003(hecha) T006(hecha) T008(hecha) T009(hecha) T010(hecha) T011(hecha) T012(hecha) T013(hecha) T014(hecha) T015(hecha) T016(hecha) | paquete-de-mas.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-017 | T001(hecha) T015(hecha) | paquete-de-mas.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-018 | T011(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-019 | T005(hecha) | — | tareas hechas | — |
| FR-020 | T006(hecha) | — | comprobado por su control | `ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo`, en make ci (verde) |
| FR-021 | T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde); `ci:internal/evals/plan_test.go:TestPlanEnDosModos`, en make ci (verde) |
| FR-022 | T008(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde); `ci:internal/evals/plan_test.go:TestPlanEnDosModos`, en make ci (verde) |
| FR-023 | T001(hecha) T016(hecha) T017(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_descripciones_test.go:TestDescripcionesCortas`, en make ci (verde); `ci:internal/evals/descripciones_del_producto_test.go:TestDescripcionesDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-024 | T008(hecha) T017(hecha) | — | tareas hechas | — |
| FR-025 | T018(hecha) | — | comprobado por su control | `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 449, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 425, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 102, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 97, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 62, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 58, ≤ 900; `evals:consulta-juridica:sin_activar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 % |
| FR-026 | T014(hecha) T015(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos`, en make ci (verde); `ci:internal/app/plazos_test.go:TestAppletPlazos`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| FR-027 | T018(hecha) | — | tareas hechas | — |
| SC-001 | T003(hecha) T009(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/skills/regiones_test.go:TestRegiones`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| SC-002 | T006(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo`, en make ci (verde) |
| SC-003 | T002(hecha) T003(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/skills/regiones_test.go:TestRegiones`, en make ci (verde) |
| SC-004 | T001(hecha) T010(hecha) T011(hecha) T012(hecha) T013(hecha) T014(hecha) T015(hecha) T018(hecha) | paquete-de-mas.txtar  | comprobado por su control | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde) |
| SC-005 | T007(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/evals/grabaciones_test.go:TestGrabacionesSinSolape`, en make ci (verde); `ci:internal/evals/grabaciones_test.go:TestManifiestoDeGrabaciones`, en make ci (verde) |
| SC-006 | T008(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde); `ci:internal/evals/plan_test.go:TestPlanEnDosModos`, en make ci (verde) |
| SC-007 | T001(hecha) T016(hecha) T017(hecha) T018(hecha) | skills-instaladas.txtar  | comprobado por su control | `ci:internal/evals/laboratorio_descripciones_test.go:TestDescripcionesCortas`, en make ci (verde); `ci:internal/evals/descripciones_del_producto_test.go:TestDescripcionesDelProducto`, en make ci (verde); `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde); `ci:internal/app/e2e_test.go:TestEntregaDelHito`, en make ci (verde); `evals:plazos-y-territorio:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:normativa:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 % |
| SC-008 | T004(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde); `ci:internal/evals/conjunto_test.go:TestControlesSobreLasReferenciasAMano`, en make ci (verde) |
| SC-009 | T009(hecha) T018(hecha) | — | comprobado por su control | `ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl`, en make ci (verde) |
| SC-010 | — | — | SIN TAREA | `evals:plazos-y-territorio:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:normativa:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:habla_en_primera_o_segunda_persona`: 0 de 1 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 9 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta`: 0 de 3 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:orden`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 60 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 120 (0,0 %), ≤ 2,8 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:normativa:duracion_de_las_sesiones:orden`: 449, ≤ 900; `evals:normativa:duracion_de_las_sesiones:herramienta`: 425, ≤ 900; `evals:normativa:duracion_del_juez:orden`: 102, ≤ 900; `evals:normativa:duracion_del_juez:herramienta`: 97, ≤ 900; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 30 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`: 0 de 60 (0,0 %), ≤ 1,7 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 125 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 124 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:duracion_del_juez:orden`: 62, ≤ 900; `evals:jurisprudencia:duracion_del_juez:herramienta`: 58, ≤ 900; `evals:consulta-juridica:sin_activar:claude-sonnet-5-5:orden`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 12 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:orden`: 0 de 6 (0,0 %), ≤ 0,0 %; `evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:herramienta`: 0 de 6 (0,0 %), ≤ 0,0 % |
| SC-011 | T018(hecha) | — | comprobado por su control | `ci:Makefile:ci`, en make ci (verde) |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`0d7cd86`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `5080c9cb` fix(H45): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `docs/SKILLS.md`, `internal/app/plazos_casos_test.go`, `internal/skills/regiones.go`, `internal/skills/regiones_test.go`, `contracts/sincronia-de-los-paquetes.md`, `mutantes.md`, `research.md`.
- `ddf90fb4` docs(H45): veredictos de la revisión final: solo registros de `gates/`.
- `23db28c9` docs(H45): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/042-h45-el-molde-admite/quickstart.md`. Suite de aceptación congelada: `specs/042-h45-el-molde-admite/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `63ba1dee`: 10 h 58 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh 63ba1dee`.
