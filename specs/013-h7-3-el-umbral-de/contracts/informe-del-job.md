# Contrato: el informe del job, con umbrales, sesiones sin medir, reintentos y duración

FR-001 a FR-008, FR-033, FR-040 a FR-044, FR-050, FR-051, FR-092, FR-093; SC-001, SC-006, SC-007. Decisiones en
[research.md](../research.md) D5, D6 y D14; tipos en [data-model.md](../data-model.md) §1-§4. Todo lo que el informe
publicaba sigue igual (contrato job-de-evals de H5 §5, lista-y-juicio de H7.2 §5); esto es lo que se añade.

## 1. `umbrales` (contrato del ADR 0029)

Siempre presente en `informe.json`, detrás de `expresiones_prohibidas_por_modelo`; `[]` sin ninguno (`legal-core`).
Ejemplo de `boe-legislacion` en un cierre que cumple (≈ 1 KB; ≈ 350 B por elemento de expresiones, ≈ 250 B el de
duración):

```json
  "umbrales": [
    {
      "nombre": "expresiones_prohibidas:claude-sonnet-5",
      "descripcion": "Respuestas de claude-sonnet-5 con alguna expresión prohibida, sobre sus respuestas medidas en las evals que activan la skill",
      "medida": 2,
      "total": 51,
      "comparacion": "<=",
      "umbral": 0.05,
      "cumple": true,
      "decide": true
    },
    {
      "nombre": "expresiones_prohibidas:claude-haiku-4-5-20251001",
      "descripcion": "Respuestas de claude-haiku-4-5-20251001 con alguna expresión prohibida, sobre sus respuestas medidas en las evals que activan la skill",
      "medida": 0,
      "total": 30,
      "comparacion": "<=",
      "umbral": 0.05,
      "cumple": true,
      "decide": false
    },
    {
      "nombre": "duracion_de_las_sesiones",
      "descripcion": "Segundos desde que se prepara la primera sesión hasta que termina la última",
      "medida": 544,
      "comparacion": "<=",
      "umbral": 900,
      "cumple": true,
      "decide": true
    }
  ],
```

Reglas: uno de expresiones por modelo del job —el que decide y después los informativos, en su orden— si la skill tiene
lista, con `medida` = `con_alguna` y `total` = `respuestas` del recuento de ese modelo; el de duración si el job da
objetivo (> 0). `cumple` = `medida/total <= umbral` (0 si `total` es 0) o `medida <= umbral`, en `float64` y sin
redondeos. Con 3 de 51: `3/51 = 0,0588… > 0,05`, `cumple: false`; con 2 de 51: `0,0392…`, `cumple: true`.

## 2. Motivos nuevos y veredicto

Van detrás de los de siempre (series, sesiones ilegibles, sin evals, ficheros mal formados, red), en este orden:

1. por cada umbral con `decide: true` y `cumple: false` que no es el de duración:
   `umbral <nombre>: <medida> de <total> (<p> %), y tiene que ser ≤ <u> %` —con un decimal y coma, como el informe
   final—, p. ej. `umbral expresiones_prohibidas:claude-sonnet-5: 3 de 51 (5,9 %), y tiene que ser ≤ 5,0 %` (FR-003);
2. con alguna sesión sin medir, uno solo:
   `de la ejecución, no de la skill: límite de uso de la cuenta: <n> sesiones sin medir: <sesión> (<motivo>), …`, p. ej.
   `de la ejecución, no de la skill: límite de uso de la cuenta: 2 sesiones sin medir: 03-lrbrl-atribuciones-del-pleno-claude-sonnet-5-02 (mensaje del límite de uso: You've hit your session limit · resets 5pm), 03-lrbrl-atribuciones-del-pleno-claude-sonnet-5-03 (sin abrir tras el límite de uso)`
   (FR-043);
3. con el de duración sin cumplir:
   `de la ejecución, no de la skill: duracion_de_las_sesiones: 901 s, y tiene que ser ≤ 900 s` (FR-051).

El prefijo fijo `de la ejecución, no de la skill: ` distingue sin modelo los dos últimos de los de la skill. El
veredicto sigue siendo `fallo` si hay algún motivo y `aprobado` si no hay ninguno, así que un umbral que decide y no se
cumple implica `fallo` (invariante del ADR 0029) y uno que se cumple, o que no decide, no cambia nada (FR-003, FR-004).
El job sale en rojo con `fallo`, como hoy (`TestEjecucionDelJob`, [ejecucion-del-job.md](./ejecucion-del-job.md) §5).

La regla por serie no cambia (ADR 0016; H7.2 FR 054; FR-007): una sesión con alguna expresión no pasa, su serie decide
con el umbral de siempre si es de las que deciden, y las informativas no deciden por su serie; lo que entra en el
umbral agregado es cada respuesta, las de las evals informativas incluidas (FR-002).

## 3. Sesiones sin medir, reintentos y duración

En la raíz de `informe.json`, detrás de `umbrales`:

```json
  "duracion_de_las_sesiones": 544,
  "reintentos_por_limite_de_ritmo": 3,
  "sesiones_sin_medir": [
    {
      "sesion": "03-lrbrl-atribuciones-del-pleno-claude-sonnet-5-02",
      "eval": "03-lrbrl-atribuciones-del-pleno.yaml",
      "modelo": "claude-sonnet-5",
      "motivo": "mensaje del límite de uso: You've hit your session limit · resets 5pm"
    }
  ],
```

En cada elemento de `evals`: `"reintentos_por_limite_de_ritmo": <n>` y `"sin_medir": "<clase>"` (`""` si se midió). En
cada elemento de `tasas`: `"sin_medir": <n>`. Una sesión sin medir no pasa, lleva como único motivo
`sin medir por límite de uso: <clase>`, no entra en `respuestas` ni en `con_alguna` del recuento y deja su serie con
`pasa: false` sin el motivo «pasan N de M». Las sesiones que el repartidor no abrió tras un límite de uso
(`InformeAEscribir.SinAbrir`) cuentan en su serie como sin medir, con la clase `sin abrir tras el límite de uso`, y no dan
el motivo «hay N sesiones y el plan pide M» (data-model §3, §4).

`duracion_de_las_sesiones`: los segundos que mide el repartidor, redondeados hacia arriba (research D14). La sesión de
la prueba de red entra en la duración y en los reintentos; no en el recuento de expresiones (H7.2 FR 053).

## 4. `informe.md`

- **Cabecera**: dos líneas más, `Duración de las sesiones: <s> s` y `Reintentos por límite de ritmo: <n>`.
- **Tasas por eval**: la columna «Resultado» dice `sin medir (<n>)` en una serie con sesiones sin medir.
- **Umbrales**, sección nueva detrás de «Expresiones prohibidas por modelo» (FR-001):

  | Umbral | Medida | Condición | Cumple | Hace fallar el veredicto |
  |---|---|---|---|---|
  | `expresiones_prohibidas:claude-sonnet-5` | 2 de 51 (3,9 %) | ≤ 5,0 % | sí | sí |
  | `expresiones_prohibidas:claude-haiku-4-5-20251001` | 0 de 30 (0,0 %) | ≤ 5,0 % | sí | no: solo se publica |
  | `duracion_de_las_sesiones` | 544 | ≤ 900 | sí | sí |

  Sin umbrales, el párrafo `ninguno`.
- **Sesiones sin medir**, sección nueva detrás de «Umbrales»: tabla `Sesión | Eval | Modelo | Motivo`, o `ninguna`.
- **Sesiones**: dos columnas más, «Reintentos por límite de ritmo» y «Sin medir» (`no` o la clase).
- **Sección de cada sesión**: el motivo `sin medir por límite de uso: …` se lista con los de sesión ilegible o sin
  terminar.

## 5. Entradas

`InformeAEscribir` gana `DuracionDeLasSesiones int` (segundos), `ObjetivoDeDuracion int` (segundos; 0, sin objetivo) y
`SinAbrir []SesionPlanificada`. Un objetivo negativo es un error que impide escribir el informe, como un umbral de
series fuera de rango; nada más cambia en los errores de `EscribirInforme`.

## 6. Tests (en `make ci`)

| Test | Casos | Requisito |
|---|---|---|
| `TestUmbralesDelInforme` (`umbrales_test.go`, tabla, sesiones sintéticas escritas con `EscribirInforme`) | 3 de 51 de Sonnet 5 con todas las series que deciden pasando → `cumple: false`, `decide: true`, `fallo` con el motivo de §2.1; 2 de 51 → `cumple: true` y el mismo veredicto y motivos que sin umbral; 2 de 30 de Haiku 4.5 → `cumple: false`, `decide: false`, veredicto sin cambiar; skill sin lista y sin objetivo → `umbrales: []`; duración 901 con objetivo 900 → `fallo` con el motivo de §2.3; 900 → sin cambiar; con todas las de un modelo sin medir → `total: 0` y `cumple: true`; en todos, rehacer la comparación de cada elemento da su `cumple` | FR-001 a FR-006, FR-051, FR-092; SC-006; US2-1 a US2-5, US3-5 |
| `TestInformeMarkdownDeLosUmbrales` (`informe_test.go`) | la sección «Umbrales» junto al recuento, con las filas de §4; «ninguno» sin umbrales | FR-001; US2-6 |
| `TestInformeConSesionesSinMedir` (`informe_test.go`) | sesión (a) → sin medir, su serie sin medir y sin el motivo de la tasa, fuera del recuento, `fallo` con el motivo de §2.2 que nombra el límite y las sesiones; `SinAbrir` → listadas, contadas en su serie, sin el motivo de sesiones que faltan; (b) y (c) → sin medir con su clase; una sesión con 429 recuperado → medida, con sus reintentos en la sesión y en el total | FR-033, FR-040 a FR-043, FR-093; SC-007; US3-2 a US3-4 |
| `TestLeerSesionConReintentos` (`sesion_test.go`) | los `api_retry` en orden con `attempt`, `max_retries` y `error`; uno sin esos campos → ilegible nombrando la línea; `TerminaEnReintento`; el texto de un último `result` con `is_error` en `ErrorDelResultado` y en el motivo sin terminar, con código 0 (`result con is_error: <texto>`) y con código 1 (`código 1: result con is_error: Failed to authenticate. API Error: 401 OAuth access token is invalid.`, el caso de research V18); sin `result` con `is_error`, `código 1` como hoy | FR-033, FR-040, FR-063, FR-065 |
| `TestClasificarElLimite` (`limites_test.go`, tabla) | (a) con cada uno de los tres principios; `API Error: Server is temporarily limiting requests (not your usage limit)` → no es (a); (b) con `attempt = max_retries`, y no con `attempt < max_retries`; (c) cortada con el último mensaje `api_retry` `rate_limit`, y no con `overloaded`; una sesión terminada tras reintentos → medida | FR-040, FR-041, FR-093; SC-007 |

## 7. Uso, de fuera adentro

| Salida | Quién y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| `umbrales` | el job (decide el veredicto), el informe final del workflow (lo lee sin modelo), la persona; una vez por job y skill | ≈ 1 KB en `boe-legislacion`, `[]` en `legal-core`; fijo, no crece con el uso del kit | cada job los mide de nuevo sobre su commit; uno incumplido deja de darse en el primer job que lo cumple |
| `sesiones_sin_medir` y su motivo | la persona y el cierre, que no deben cambiar la skill por él; una vez por job | `[]` lo habitual; como mucho las 94 sesiones del plan, ≈ 230 B cada una, ≤ 22 KB, solo en un job que tocó el límite | solo en ese job |
| reintentos | quien ajusta la concurrencia en la definición del job | un entero por sesión (≤ 94, ≈ 40 B con su clave) y uno en la raíz, ≤ 4 KB | cada job los mide de nuevo |
| `duracion_de_las_sesiones` | el job (umbral) y la persona | un número | cada job la mide de nuevo |

Nada de esto depende de las normas o los bloques consultados con el kit: cada sesión empieza de un estado preparado.
