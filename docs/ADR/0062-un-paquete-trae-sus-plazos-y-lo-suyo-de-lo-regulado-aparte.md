# 0062 · Un paquete de materia trae sus plazos y lo suyo de lo regulado aparte, y el molde lo admite sin código

- **Estado**: propuesta (2026-11-22); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-22
- **Hito**: H45 (puntos 1 a 5, 7 y 8) y H42 (punto 6). Se escribe con sus entradas, antes de los runs y fuera de
  ellos, en una sesión interactiva. Precisa el orden del ADR 0059 (punto 1): H45 va entre H36 y H42. Precisa el molde
  del paquete de materia de H32 (`specs/030-h32-arquitectura-de-skills/contracts/paquetes.md`, §1 y §3) y de H33
  (`contracts/datos-y-regiones.md`), el ADR 0044 (punto 5) y el ADR 0058 (punto 4) en qué referencia lleva cada tipo
  de plazo y cada fila de lo regulado aparte. Cambia la concurrencia de `plazos-y-territorio` en el trabajo de evals
  que fijó H32 (`contracts/evals.md` §6) y sustituye el tope de once sesiones a la vez de H30 (FR-020 y SC-011). Cambia
  la descripción de `plazos-y-territorio`, que nombra una materia. No cambia lo que `kitlegal plazos` calcula.

## Contexto y problema

El ADR 0059 puso H42 justo detrás de H36 para que los defectos del molde aparecieran antes de diseñar los otros
puertos: hasta hoy, el único paquete real es el administrativo, y el segundo, la vertical de ensayo de
`internal/skills/testdata/vertical-de-ensayo/`, no trae tipos de plazo ni normas que rijan lo regulado aparte.

El 2026-11-22 se montó el paquete tributario como prototipo, sobre `main` (`c5cf9a1b`) en una copia fuera del
repositorio: la LGT y la Ley del IRPF con `vertical: fiscal`, dos tipos de plazo nuevos en `data/plazos.yaml`, una fila
de `cubierto` y dos filas de `regulados_aparte` estrechadas en `data/cobertura.yaml`, y una referencia `fiscal.md` en
`plazos-y-territorio` y otra en `normativa`. Con eso, sin tocar Go:

1. **`make skills-sync` no termina.** La región `recursos_y_plazos` escribe todos los tipos de `data/plazos.yaml` en
   la referencia de cada paquete que la lleva (`filasDeLosRecursosYPlazos`, `internal/skills/regiones.go`): los dos
   tipos tributarios entran también en `administrativo.md`, que pasa de 100 líneas sin índice y la sincronía lo
   rechaza. Un tipo no sabe de qué paquete es.
2. **Lo tributario que sigue aparte se queda en la referencia de otro paquete.** Una fila de `regulados_aparte` va a
   la referencia del paquete de la norma de su precepto (`filasDeLaCobertura`), y el precepto de todas es de la LPAC.
   Una pregunta tributaria lee `fiscal.md`, y la fila que dice que la aplicación de los tributos se rige por la LGT
   está en `administrativo.md`.
3. **No hay sitio en el cuerpo de `normativa`.** Su cuerpo mide hoy 10 927 caracteres, con el tope en 11 000 (ADR
   0046). La línea del paquete fiscal lo lleva a 10 997, y la vertical de ensayo, que añade un paquete más,
   a 11 066: `TestVerticalDeEnsayo`, que prueba desde H32 que una materia entra sin Go, sale en rojo.
4. **Las grabaciones de un hito nuevo piden Go.** `grabar_datos` deja lo grabado en `evidencias/<hito>/`, y
   `UnionDeGrabaciones` (`internal/evals/grabaciones.go`) enumera las raíces de H4, H5, H9 y H36 en una constante por
   hito: H36 tuvo que añadir la suya.
5. **Una veintena de tests fijan con una lista literal lo que hoy hay en `data/` o en `evals/`** (la lista, en la
   entrada de H45): las filas de `cubierto` y de `regulados_aparte`, la tabla de tipos, el texto de una fila, el
   mensaje de un tipo desconocido, el peor caso del trabajo de evals, el plan de sesiones por eval y las evals de
   cobertura. Y el conjunto de `normativa` está cerrado en Go: como mucho 23 evals y exactamente 10 positivas
   (`internal/evals/conjunto.go`).
6. **El trabajo de evals de `plazos-y-territorio` no admite ninguna eval más.** Su peor caso, con 12 evals a
   concurrencia 2, es de 5 h 45 min, y el tope es de 352 min (`TestDefinicionDelJob`): cada eval suma unos 27 min.
7. **La descripción de `plazos-y-territorio` nombra una materia**: cuenta los plazos «del procedimiento común (LPAC) o
   el del contencioso» y «no cuenta plazos de otras materias». Con un segundo paquete deja de ser verdad, y la skill
   podría no activarse ante un plazo tributario.

La revisión tributaria se cuenta con la LPAC. Los arts. 223.1 y 235.1 LGT fijan un mes desde el día siguiente al de la
notificación y no traen regla de cómputo, el art. 7.2 LGT hace supletorio el derecho administrativo, y la DA 1ª.2.a
LPAC rige esa revisión «supletoriamente por lo dispuesto en esta Ley». En el prototipo, los dos tipos con el cómputo
`lpac` dan con el binario las fechas esperadas, y `TestPreceptosDeLosPlazos` confirma sus `vigente_desde` contra lo
grabado.

## Opciones consideradas

1. **Que el hito cambie a mano lo que choca**: los tests de listas, una constante de grabaciones, un `if` por paquete en
   la sincronía. Rechazada: es el «sin tocar código» que deja de ser verdad, y el tercer paquete pagaría lo mismo.
2. **Un campo `paquete:` en cada tipo de `data/plazos.yaml` y en cada fila de lo regulado aparte.** Rechazada: el
   paquete ya lo dice el `vertical` de la norma, y dos fuentes del mismo dato pueden contradecirse.
3. **Que la referencia del paquete salga de la norma que funda cada cosa, y quitar las listas cerradas.** Elegida.

Para el sitio del cuerpo había dos caminos:

- **Una referencia generada con la lista de paquetes**, y en el cuerpo una sola orden de leerla. Rechazada por ahora:
  añade un salto de lectura a toda pregunta de plazos y de normativa, en las órdenes que H41 y H41.1 midieron (ADR 0051
  a 0053), para un problema que hoy solo tiene `normativa`.
- **Dejar la línea por paquete como está y hacer sitio en `normativa`.** Elegida, con el límite escrito en
  «Consecuencias».

## Decisión

1. **El paquete de un tipo de plazo es el de la norma de sus preceptos.** Los preceptos propios de un tipo, los de
   `preceptos` y los de `unidades`, nombran normas de `data/normas.yaml`, y las que llevan `vertical` comparten uno: ese
   es su paquete. La región `recursos_y_plazos` de la referencia de un paquete lista solo sus tipos. Un tipo sin
   paquete, o con dos, es un defecto de la sincronía, con el tipo: un plazo que necesite normas de dos paquetes pide
   decidir antes a qué paquete pertenecen esas normas, no una excepción en el tipo. Los preceptos del cómputo no
   cuentan: el cómputo `lpac` sirve a los dos paquetes. Con los datos de hoy, los siete tipos son de lo administrativo, y
   `administrativo.md` sale igual byte a byte.

2. **Una fila de lo regulado aparte es del paquete de la norma que la rige**, si la lleva y esa norma tiene `vertical`;
   si no, del paquete de la norma de su precepto, como hoy. La misma regla vale para generar la fila
   (`filasDeLaCobertura`) y para exigir que alguna región la genere (`filasSinRegion`): una fila con la norma de un
   paquete cuya referencia no lleva la región es un defecto. La línea de la fila no cambia: sigue nombrando la norma que
   la rige y el precepto que dice que va aparte, sin exponer lo que la norma dice (ADR 0058, puntos 1 y 2).
   Con los datos de hoy, ninguna fila cambia de referencia: la única norma con `vertical` que rige una fila es la LGT,
   que no lo tiene hasta H42, y la del tráfico no lo tiene.

3. **Lo que un paquete añade a un `SKILL.md` cabe en su tope**, con la línea por paquete que genera la sincronía, sin
   cambiar su forma. El cuerpo de cada skill de capacidad, con sus paquetes y con la vertical de ensayo, queda bajo
   las 11 000 del ADR 0046: es lo que comprueba `TestVerticalDeEnsayo`. Con unos 70 caracteres por paquete y al menos
   100 de margen, `normativa` admite un paquete real más. En H45, `normativa` hace sitio moviendo a una de
   sus referencias texto que no es una orden de Bash ni una línea `⚠`: la forma para PowerShell 5.1 de la orden de
   lectura del paso 2, que va a `references/con_herramientas.md`, y lo dice su declaración en el laboratorio de H41.
   Para eso, el control `ordenes-para-powershell` admite que la forma para PowerShell de una orden del `SKILL.md` esté
   en una referencia a mano de la skill (las de una referencia la siguen llevando detrás, en su fichero). Y el
   laboratorio de H41 deja de exigir en su declaración las `references/<vertical>.md` de un paquete que su base no
   tiene: un paquete entra con datos y su referencia, sin cambiar el laboratorio.

4. **Las grabaciones de un hito se suman solas.** `UnionDeGrabaciones` junta las de H4, H5 y H9 y, sin una constante
   por hito, la carpeta `boe.legislacion-consolidada/` de cada `evidencias/<hito>/` que tiene `manifiesto.json` y esa
   carpeta, con la ruta relativa al paquete como hoy (`../../evidencias/`) y en el orden del número del hito, no el
   alfabético (h36 antes que h42). `evidencias/h9/` no entra: sus carpetas son de festivos.
   `TestGrabacionesSinSolape` sigue exigiendo que dos raíces no compartan nada más que el `robots.txt`.

5. **Ningún test de un mecanismo fija con una lista literal lo que hay en `data/` o en `evals/`.** El que comprueba el
   repositorio deriva lo esperado de `data/` y de `evals/`; el que comprueba una regla usa sus propios datos de
   `testdata/`. Así, añadir un tipo, una fila, una norma con `vertical` o una eval no rompe un test que no es de ese
   dato. Las reglas del conjunto de `normativa` (23 evals y 10 positivas) no cambian: H42 no le añade evals.
   Queda fuera el test que declara a propósito qué normas tiene un paquete, como
   `TestNormasDelRepositorio/paquete-de-lo-administrativo`: esa lista es el paquete, y cambiarla es cambiarlo.

6. **La revisión tributaria se cuenta con el cómputo `lpac`, sin uno nuevo.** Sus tipos llevan el precepto que fija la
   duración (arts. 223.1 y 235.1 LGT) y el que hace supletorio el derecho administrativo (art. 7.2 LGT). El cómputo
   sigue citando el art. 30 LPAC: meses de fecha a fecha, la prórroga, los dos municipios del art. 30.6 y agosto
   hábil. Como los preceptos del art. 30 rigen así desde el 2016-10-02, un plazo tributario que empieza antes no se
   calcula (ADR 0042), igual que los de lo administrativo.

7. **`plazos-y-territorio` pasa a tres sesiones a la vez en el trabajo de evals.** Con las 16 evals de H42, a dos su
   peor caso es de 7 h 34 min, y a tres, de 5 h 14 min, bajo el tope (cuenta de `TestDefinicionDelJob` sobre el
   prototipo). Las cuatro skills pasan de once a doce sesiones a la vez contra la misma suscripción: este punto
   sustituye el tope de once de H30 (FR-020 y SC-011; `sesionesALaVezComoMucho` en `internal/evals/definicion_test.go`)
   por doce. Ni las once ni las doce se han medido (H25 y H32).

8. **La descripción de una skill de capacidad nombra su capacidad, no una materia.** La de `plazos-y-territorio` deja
   de nombrar la LPAC y el contencioso y de decir que no cuenta plazos de otras materias: cuenta los plazos que cubren
   sus paquetes, y lo demás lo declara. Se escribe con el criterio fijado antes de medir del ADR 0055, bajo su tope de
   900, y la mide su juez de descripciones.

9. **El molde se arregla en su propio hito, H45, antes que el paquete.** H45 hace los puntos 1 a 5, 7 y 8, con lo
   administrativo y la vertical de ensayo como sus dos implementaciones, y H42 entra después solo con datos: sin un
   fichero `.go`, de `.github/` ni un `SKILL.md` escrito a mano. Así, si un cierre cae, se sabe si es del molde o del
   paquete, y H42 prueba lo que promete el principio X en lugar de suponerlo.

## Consecuencias

**A favor**

- Un paquete entra con sus normas, sus tipos, lo suyo de lo regulado aparte y sus evals sin tocar Go ni el protocolo
  de ninguna skill, y la vertical de ensayo lo prueba con un tipo y una fila con norma, además de lo que ya probaba.
- Lo que sigue aparte de una materia está en la referencia que lee la pregunta de esa materia, con su norma.
- Ningún dato nuevo rompe un test ajeno: el tercer paquete no paga lo que pagó el segundo.

**En contra, y asumido**

- El cuerpo de una skill de capacidad crece unos 70 caracteres por paquete. Con el tope de 11 000, `normativa` y
  `jurisprudencia` admiten pocos más. Cuando el siguiente no quepa, un hito de la base cambia la forma de la región
  —una referencia con la lista, o una línea con todos los paquetes— y lo mide.
- `normativa` queda en 335 min de su tope de 352 sin evals nuevas; con una más pasaría a 349,7, y su conjunto,
  cerrado en 23, no la admite sin cambiar sus reglas.
- Doce sesiones a la vez pueden toparse con el límite de la suscripción. Si pasa, el cierre lo enseña como un fallo de
  la fuente, no de la skill, y se decide con la serie.
- El hito toca `internal/skills` e `internal/evals`, así que el cierre mide todas las skills sin reutilizar nada de
  `main`.
