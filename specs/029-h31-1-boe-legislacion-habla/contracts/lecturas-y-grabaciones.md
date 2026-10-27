# Contrato: las lecturas de las evals por materia y sus grabaciones (FR-011 a FR-013; SC-004, SC-008)

## 1. Las lecturas (FR-011)

Nueve entradas más en `lecturas` (formato de H26.1: `{norma, bloque, motivo}`), detrás de las de hoy, cada una con el
mismo `motivo`. Nada más de esas evals ni de ninguna otra cambia, y una lectura sigue sin contar en ningún umbral.

| Eval | Hoy | Se añaden, en este orden |
|---|---|---|
| `evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml` | `a18` de `BOE-A-2006-20764` | `a7`, `a19`, `a20`, `dadecimoctava` de `BOE-A-2006-20764` |
| `evals/boe-legislacion/14-trlrhl-impuestos-por-materia.yaml` | `a15` de `BOE-A-2004-4214` | `a60`, `a92`, `a100`, `a104` de `BOE-A-2004-4214` |
| `evals/boe-legislacion/17-ltaibg-plazo-por-materia.yaml` | ninguna | `a24` de `BOE-A-2013-12887`, en una clave `lecturas` nueva al final del fichero |

Cada entrada, como las de hoy:

```yaml
  - norma: BOE-A-2006-20764
    bloque: a7
    motivo: "Lo nombran, sin haberlo leído, respuestas de los sondeos de evidencias/sondeos-20261025/ o de los jobs de las evals, y la regla de leer lo que se nombra lo pide (H31.1; ADR 0043, clase b)."
```

El «o» hace el motivo cierto para las nueve (research D6): de la eval 15, las respuestas del sondeo de control nombran
el art. 7 y la disposición adicional decimoctava (`votos.json` de `brazo-A` y `brazo-A2`: 68 y 52 de las 80
respuestas); de la 17, las del cierre de H31 y del primer run, el art. 24 (6 de 6 en cada uno); de la 14, una del
cierre de H30, el art. 60. Lo demás lo dice la entrada del hito (research S2).

Los ids son los de los índices grabados de su norma, en
`testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_<norma>_texto_indice.json`
(los nueve, con `grep -F '\"id\": \"<id>\",'` en esta sesión; research V8).

## 2. El manifiesto (FR-012): la tarea `[datos]`

`internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`, el conjunto de H9), en las dos
entradas que ya tienen las lecturas de hoy; nada más del fichero cambia:

| Entrada (`titulo_empieza_por`) | `bloques` hoy | `bloques` tras la tarea | `para` |
|---|---|---|---|
| `Ley 35/2006,` | `["a18"]` | `["a18", "a7", "a19", "a20", "dadecimoctava"]` | el de hoy, con «; y, desde H31.1, los arts. 7, 19 y 20 y la disposición adicional decimoctava, que la regla de leer lo que se nombra pide en la eval 15 (FR-011, FR-012 de H31.1)» antes del punto final, si lo hay |
| `Real Decreto Legislativo 2/2004,` | `["a15"]` | `["a15", "a60", "a92", "a100", "a104"]` | el de hoy, con «; y, desde H31.1, los arts. 60, 92, 100 y 104, que la regla de leer lo que se nombra pide en la eval 14 (FR-011, FR-012 de H31.1)» |
| `Ley 19/2013,` | `["a24"]` | sin cambios: `a24` ya está, para `data/plazos.yaml` de H9, y su grabación sirve también la lectura de la eval 17 | sin cambios |

## 3. La grabación: el paso `grabar_datos`, sin modelo

- **Fuente**: `boe.legislacion-consolidada`, con fila en `docs/SOURCES.md` de `main` y «Revisado» `2026-09-13`
  (`git show main:docs/SOURCES.md`, línea 42; research V9). Ninguna otra fuente entra en el hito.
- **Test de grabación**: `TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`), que ya existe:
  sirve cada consulta desde lo grabado y solo pide a la fuente, por `internal/httpx` (ritmo `1s`, `robots.txt`,
  User-Agent), lo que da una falta (líneas 38 a 75). Lo lanza `grabar_datos` con `KITLEGAL_RECORD=1` tras el commit de
  la tarea `[datos]`; ninguna tarea toca la red ni la variable.
- **Respuestas**: ocho ficheros nuevos en `internal/evals/testdata/boe.legislacion-consolidada/`,
  `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_<norma>_texto_bloque_<id>.json`, uno por bloque
  de §2 salvo `a24`. Grabados en la rama del primer run, ocupan de 1 766 bytes (`a60`) a 402 526 (`a7`)
  (`git ls-tree -l`, research V10). `internal/httpx` pide antes el `robots.txt`, y su grabación,
  `GET_https_www.boe.es_robots.txt.json`, puede regrabarse con otro `grabado_en`, `Age` y `Date`.
- **Material de origen en `$KITLEGAL_EVIDENCIAS`**: ninguno, porque ningún fichero de `data/` sale de lo grabado.
- Si la fuente no responde, es una causa mayor del run (dependencia externa inaccesible), no un caso del producto.

## 4. Los tests (FR-013; SC-004, SC-008)

| Test | Cambia | Qué hace cumplir |
|---|---|---|
| `TestEvalsDelRepositorio/lecturas` (`internal/evals/conjunto_test.go:2303`, `probarLasLecturasDeLasEvalsDelRepositorio`) | el mapa escrito a mano gana las nueve entradas, con una constante para el motivo de §1, y la 17 | que cada eval lleve exactamente sus lecturas: ni una de más, ni de menos, ni otro motivo (FR-011) |
| `TestCacheDeLasSesiones/lecturas` (`internal/evals/preparar_test.go:1086`, `probarLasLecturasDeLasEvals`) | la tabla gana los nueve bloques | que la caché de una sesión sirva cada uno sin red con `boe articulo … --offline`, sin fecha y con `--fecha 9999-12-31`, con código 0 (FR-013) |
| `TestEvalsDelRepositorio/grabado` | nada | que la unión de grabaciones sirva cada consulta necesaria de las evals, las lecturas entre ellas, sin red (FR-012) |
| `TestLecturaDeLaEvalSinGrabar` (`internal/evals/preparar_test.go:738`) | nada | que una lectura sin grabar dé sus dos faltas y la sesión no se prepare (US4, escenario 3) |

Con las grabaciones del primer run y las lecturas de las evals, el subtest `lecturas` del mapa sin cambiar sale en
rojo, los otros tres en verde (la tabla, con los nueve bloques, en 0,54 s); sin la grabación de `a7`, `grabado` y la
tabla salen en rojo (research V11).

Comentarios de esos tests que enumeran las lecturas («la 14, la 15 y la 16») se actualizan con la 17 en la misma
tarea.
