# Contrato: las `lecturas` de las evals 14, 15 y 17 y sus grabaciones

FR-009 a FR-011, SC-004 y SC-008. El formato de `lecturas` es el de H26.1 (`internal/evals/formato.go`,
`LecturaDeLaEval`: `norma`, `bloque` y `motivo` no vacío ni solo de blancos; ninguna puede ser un comando de bloque ni
una cita de su eval) y no cambia, ni el código que prepara la caché de una sesión (`prepararLaCache`,
`internal/evals/preparar.go`: las consultas necesarias de **todas** las evals de la skill, lecturas incluidas, más lo
que lleva a leer `data/`, menos el `sin_servir` de la eval).

## 1. Las nueve lecturas (FR-009)

Cada una detrás de las de hoy, en este orden y con este `motivo`, el mismo en las nueve (clase (b) del ADR 0043: la
lectura que haría una respuesta mejor, de la norma de la pregunta, vista en sesiones del modelo que decide):

```yaml
    motivo: "Lo nombran sin haberlo leído las respuestas de los sondeos de evidencias/sondeos-20261025/ y de los jobs, y la regla de leer lo que se nombra lo pide (H31.1; ADR 0043, clase b)."
```

| Eval | Norma | Bloques nuevos, en orden | Ya tenía | Comando y cita de la eval |
|---|---|---|---|---|
| `evals/boe-legislacion/14-trlrhl-impuestos-por-materia.yaml` | `BOE-A-2004-4214` | `a60`, `a92`, `a100`, `a104` | `a15` | `a59` |
| `evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml` | `BOE-A-2006-20764` | `a7`, `a19`, `a20`, `dadecimoctava` | `a18` | `a17` |
| `evals/boe-legislacion/17-ltaibg-plazo-por-materia.yaml` | `BOE-A-2013-12887` | `a24` (la eval no tenía `lecturas`: la clave es nueva) | — | `a20` |

Cada id es el de la entrada de su índice grabado (`testdata/evals/boe.legislacion-consolidada/…_texto_indice.json`,
de H5) cuyo `titulo` es el precepto: «Artículo 7», «Artículo 19», «Artículo 20», «Disposición adicional
decimoctava», «Artículo 60», «Artículo 92», «Artículo 100», «Artículo 104» y «Artículo 24» (research V4). Nada más de
esas evals cambia, ni ninguna otra eval: el diff de `evals/` de la rama son esas nueve entradas y la clave `lecturas:`
de la 17.

## 2. El manifiesto (FR-010, tarea `[datos]`)

`internal/evals/testdata/grabaciones.json`, `"fuente": "boe.legislacion-consolidada"` (fila de `docs/SOURCES.md` de
`main` con «Revisado» 2026-09-13). Dos entradas ganan bloques, en el orden de §1, y su `para` lo dice; ninguna entrada
nueva, y la de la Ley 19/2013 no cambia (`a24` está grabado para `data/plazos.yaml` desde H9):

| `titulo_empieza_por` | `bloques` tras la tarea | `para`: lo que se añade al final |
|---|---|---|
| `Ley 35/2006,` | `a18`, `a7`, `a19`, `a20`, `dadecimoctava` | «Desde H31.1 (FR-009, FR-010), los arts. 7, 19 y 20 y la disposición adicional decimoctava, que las respuestas de la eval 15 nombran sin haberlos leído y que la regla de leer lo que se nombra pide, con los ids de su índice grabado.» |
| `Real Decreto Legislativo 2/2004,` | `a15`, `a60`, `a92`, `a100`, `a104` | «Desde H31.1 (FR-009, FR-010), los arts. 60, 92, 100 y 104, que las respuestas de la eval 14 nombran sin haberlos leído y que la regla de leer lo que se nombra pide, con los ids de su índice grabado.» |

La tarea `[datos]` toca solo ese fichero. `make ci` sigue en verde tras ella: en un build normal el manifiesto solo se
lee (`TestManifiestoDeGrabaciones`), y ninguna eval pide aún los bloques nuevos.

## 3. La grabación (paso `grabar_datos`)

`scripts/workflow/grabar-datos.sh grabar H31.1` ejecuta el test de grabación que ya existe,
`TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`), con `KITLEGAL_RECORD=1`: por cada
entrada del manifiesto, `boe buscar`, `metadatos`, `indice` y `articulo` de cada bloque, sirviendo primero de las
grabaciones de H4, H5 y H9 y pidiendo por `internal/httpx` (robots.txt, User-Agent y `1s` de ritmo) solo lo que falta.
Lo que falta son los ocho bloques, `…/id/BOE-A-2006-20764/texto/bloque/{a7,a19,a20,dadecimoctava}` y
`…/id/BOE-A-2004-4214/texto/bloque/{a60,a92,a100,a104}`: la búsqueda, los metadatos y el índice de las dos normas están
grabados desde H5, como dice el `para` de H26.1 (supuesto, research S2: se comprueba al grabar). Las respuestas quedan en
`internal/evals/testdata/boe.legislacion-consolidada/` (ocho ficheros `GET_https_www.boe.es_…_texto_bloque_<id>.json`).
Una grabación de bloque sirve la lectura sin fecha y la de `--fecha 9999-12-31`: el BOE da todas las versiones del
bloque en la misma respuesta, como ya hacen `a18` y `a15` (H26.1). No hay material de origen: ningún fichero de
`data/` sale de lo grabado, y `$KITLEGAL_EVIDENCIAS` (`evidencias/h31-1/`) no recibe nada de esta fuente. Si el BOE no
responde tras un reintento, es la causa mayor «dependencia externa inaccesible» del run, no un caso del producto.

## 4. Los tests de `lecturas` de H26.1 con los bloques nuevos (FR-011, SC-004, SC-008)

Van en la tarea que añade las lecturas a las evals, detrás de la grabación:

- **`TestEvalsDelRepositorio/lecturas`** (`internal/evals/conjunto_test.go`,
  `probarLasLecturasDeLasEvalsDelRepositorio`): el mapa escrito a mano gana las nueve entradas con el `motivo` de §1
  —la 14 con `a15` y las cuatro suyas, la 15 con `a18` y las cuatro suyas, y la 17 con `a24`—, y su comentario dice
  que llevan lecturas la 14, la 15, la 16 y la 17. Falla si a una eval le falta o le sobra una lectura, o cambia su
  motivo.
- **`TestEvalsDelRepositorio/grabado`** (sin cambios): `ConsultasNecesarias` incluye las lecturas, sin fecha y con
  `9999-12-31`, y el subtest exige que la unión de grabaciones las sirva y que la caché preparada las dé sin red. Con
  una lectura sin grabar sale en rojo con la falta («la lectura BOE-A-2006-20764 a7 sin su bloque…», comprobado en un
  prototipo, research V7).
- **`TestCacheDeLasSesiones/lecturas`** (`internal/evals/preparar_test.go`, `probarLasLecturasDeLasEvals`): la tabla
  pasa a llevar, por fila, la eval de cuya sesión se prepara la caché. Las cuatro filas de hoy siguen leyéndose desde
  la sesión de la 01, que no lleva ninguna; las nueve nuevas, desde la sesión de su propia eval (US3, escenarios 1 y
  2): `a60`, `a92`, `a100` y `a104` de `BOE-A-2004-4214` desde la 14; `a7`, `a19`, `a20` y `dadecimoctava` de
  `BOE-A-2006-20764` desde la 15; y `a24` de `BOE-A-2013-12887` desde la 17. Cada una, `boe articulo … --offline` sin
  fecha y con `--fecha 9999-12-31`, tiene que salir con 0; se prepara una caché por eval distinta (cuatro).
- **`TestLecturaDeLaEvalSinGrabar`** (sin cambios): una lectura sin grabar da la falta y la sesión no se abre
  (US3, escenario 3).

Ninguna de las dos tareas cambia código de producción ni el arnés: solo datos, evals y los tests que los enumeran.
