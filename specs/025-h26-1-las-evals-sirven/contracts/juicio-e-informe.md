# Contrato: lo que el juicio y el informe del job dicen de las lecturas

Quién lo consume: la persona que lee el informe del job (`informe.json` e `informe.md`), una vez por job y skill, y el
informe final del run, que lee sin modelo los umbrales con `decide: true` (ADR 0029). Requisitos: FR-007, FR-015,
FR-016, FR-019, FR-021 a FR-028, FR-033; SC-005 a SC-007, SC-012 a SC-015, SC-017.

## 1. El reparto de cada invocación

El de `data-model.md` §4. Ejemplos, con las claves que el informe escribe de cada sesión:

```json
"lecturas_sin_servir": [{"orden": "boe articulo BOE-A-2015-10565 a124 --fecha 2026-09-09 --json", "codigo": 5}],
"otras_fallidas": [
  {"orden": "boe_indice BOE-A-2015-10565", "codigo": null, "motivo": "sin_sobre"},
  {"orden": "boe indice BOE-A-2015-10565 --json", "codigo": 1, "motivo": "salida_no_escrita",
   "bash": "kitlegal boe indice BOE-A-2015-10565 --json | python3 -c \"import json,sys; d=json.load(sys.stdin); print(len(d['data']['bloques']))\""},
  {"orden": "boe indice BOE-A-2015-10565 --json", "codigo": 1}
]
```

- `lecturas_sin_servir` está en toda sesión, `[]` si no tiene ninguna. Un elemento de una sesión cortada por el tope
  lleva `codigo: null`. Una orden de salida no escrita que consulta una lectura de `sin_servir` va aquí, con su
  `codigo: 1`, su `motivo` y su `bash`: llegó a la lectura (FR-015, FR-024).
- `motivo` y `bash` no están cuando no aplican: el `1` sin enlace queda como hoy (FR-027). La llamada sin sobre lleva
  también `"codigo": null, "motivo": "sin_sobre"` en su elemento de `invocaciones` (FR-023).
- Nada de esto cambia `pasa`, `motivos`, lo que satisface o ejecuta cada invocación, la tasa de su serie ni ningún
  umbral, salvo FR-016 y los dos umbrales de §4 (FR-028).

## 2. El comando que la eval hace fallar (FR-016)

Un comando esperado de la forma bloque cuya norma y cuyo bloque están en el `sin_servir` de su eval lo satisface, además
de lo de hoy, la invocación que lee ese bloque como el comando lo pide —con su `fecha`, si la lleva— y termina con 4 o
5. En la 11: `boe articulo BOE-A-2015-10565 a124 --fecha 2026-09-09` con 5, o la llamada `boe_articulo` con
`{"norma": "BOE-A-2015-10565", "bloque": "a124", "fecha": "2026-09-09"}` y un sobre de la clase `fuente-no-disponible`.
La que lo lee sin `--fecha` o en otra fecha, y la sesión que no lo lee, dejan el comando ausente, con su motivo.

## 3. La línea de la salida no escrita (FR-024 a FR-026)

El prefijo es el que el presentador del binario pone delante de la causa cuando no puede escribir en la salida estándar,
pedido una vez a `render.Nuevo` con un escritor que falla (research D8): hoy,
`render: no se pudo escribir en la salida estándar: `. Una orden de `Sesion.OrdenesDeBash` trae la línea si alguna línea
de su salida empieza por él. Ejemplo de la salida de un uso de Bash en un runner sin Python:

```text
Exit code 127
/bin/bash: line 1: python3: command not found
render: no se pudo escribir en la salida estándar: write /dev/stdout: broken pipe
```

La invocación `boe indice BOE-A-2015-10565` de la traza, con código 1, se enlaza con ese uso si la orden contiene como
palabras `boe`, `indice` y `BOE-A-2015-10565`, y si ningún otro uso de la sesión lo cumple.

## 4. Los umbrales nuevos

Detrás de los de hoy, por cada modo del plan en su orden y, en cada modo, por cada modelo del plan —el que decide y los
informativos en su orden—, estos dos; el segundo, solo en el informe de una skill con al menos una eval con un comando
de bloque cuya lectura está en su `sin_servir` (hoy, `legal-core`), como `plazo_sin_fecha` solo en el de la que tiene
evals de plazo:

```json
{"nombre": "fuera_de_lo_grabado:claude-sonnet-5-5:orden",
 "descripcion": "Lecturas de claude-sonnet-5-5 en el modo orden que salieron con 4 o 5 sin estar en la caché, sin las de sin_servir, sobre sus sesiones medidas",
 "medida": 0, "total": 33, "comparacion": "<=", "umbral": 0, "cumple": true, "decide": false}
```

```json
{"nombre": "lectura_sin_intentar:claude-sonnet-5-5:orden",
 "descripcion": "Sesiones de claude-sonnet-5-5 en el modo orden que no intentaron la lectura que un comando esperado de su eval pide y su sin_servir no sirve, sobre sus sesiones medidas de las evals con un comando así",
 "medida": 0, "total": 3, "comparacion": "<=", "umbral": 0, "cumple": true, "decide": true}
```

- **Las sesiones medidas** de un modelo en un modo son las de research D11: las de las series que pide el plan con ese
  modelo y ese modo, legibles y sin `sin_medir`, de cualquier eval de ese modo. Las evals sin binario ni servidor no
  tienen modo y no cuentan. En `legal-core`, con las doce evals, 33 por modelo y modo (11 evals de modo × 3).
- **`fuera_de_lo_grabado`**: medida, la suma de las lecturas de su `fuera_de_lo_grabado`, que ya no lleva las de
  `sin_servir`; total, esas sesiones; `decide: false` siempre, con su `cumple` como el de cualquier umbral: nunca da
  motivo ni cambia el veredicto (FR-021). En el ejemplo de US4: medida 2, total 3, `cumple: false`, veredicto igual.
- **`lectura_sin_intentar`**: total, las sesiones medidas de las evals con un comando de bloque cuya lectura está en su
  `sin_servir` (hoy la 11: 3 por modelo y modo); medida, las de esas que no tienen en `lecturas_sin_servir` ninguna
  invocación que consulte el bloque de cada uno de esos comandos, con o sin `--fecha` y termine como termine —también
  la que el tope cortó antes de intentarla—; `decide: true` solo con el modelo que decide. La 12 no entra: no tiene
  comandos. Va también con medida 0; el informe de una skill sin una eval así, como `boe-legislacion` y
  `jurisprudencia`, no lo lleva (FR-033).
- **Su motivo**, si decide y no se cumple, en la forma del de `plazo_sin_fecha`:
  `umbral lectura_sin_intentar:claude-sonnet-5-5:orden: 1 de 3 (33,3 %), y tiene que ser ≤ 0,0 %: 11-plazo-reposicion-sin-leer-el-precepto-claude-sonnet-5-5-02`.
  Con él, el veredicto es `fallo` y el trabajo de `legal-core`, rojo.

Con el plan del job, dos modos y dos modelos (`claude-sonnet-5-5` y `claude-haiku-4-5-20251001`), son 8 en
`legal-core` y 4, los de `fuera_de_lo_grabado`, en las otras dos:

| Skill | Umbrales hoy (deciden) | Con H26.1 (deciden) |
|---|---|---|
| `legal-core` | 10 (10) | 18 (12) |
| `boe-legislacion` | 15 (11) | 19 (11) |
| `jurisprudencia` | 15 (11) | 19 (11) |

## 5. La serie

Cada elemento de `tasas` lleva `sesiones_con_lectura_sin_servir`: en la serie de una eval con `sin_servir`, cuántas de
sus sesiones legibles tienen algo en `lecturas_sin_servir`, también 0; `null` en las demás (FR-015). Ejemplo de la de la
11 con el modelo que decide en el modo orden: `"sesiones_con_lectura_sin_servir": 3`. No decide: lo que decide es
`lectura_sin_intentar` y la regla de 2 de 3 de la serie, como hoy.

## 6. `informe.md`

Los umbrales nuevos salen en su tabla como los demás. En la tabla de invocaciones de una sesión, la llamada sin sobre
dice «sin sobre» en la columna del código, en lugar de «sin código (sesión cortada)».

## 7. Tamaño y cuándo deja de darse cada señal

| Señal | Por job | Bytes | Deja de darse |
|---|---|---|---|
| `fuera_de_lo_grabado:<modelo>:<modo>` | 4 por skill, 12 en los tres trabajos | unos 330 cada uno | en el job siguiente al que cierra el hueco (la lectura entra en `data/` o en `lecturas` y se graba, o la skill deja de pedirla); nunca se arrastra |
| `lectura_sin_intentar:<modelo>:<modo>` | 4 en el de `legal-core`, ninguno en los otros dos | unos 365 cada uno | en el job en que las sesiones intentan la lectura; sale del informe cuando la skill deja de tener una eval que lleve en `sin_servir` la lectura de un comando |
| su motivo | como mucho uno por modo, con las sesiones de las evals con un comando en `sin_servir` (hoy, como mucho las 3 de la 11) | menos de 300 | ídem |
| `lecturas_sin_servir` | en cada sesión; con algo, solo las 24 de la 11 y la 12 (2 evals × 2 modelos × 2 modos × 3) | `[]` son 25; cada elemento, unos 85; de 1 a 3 por sesión | cuando la eval deja de llevar `sin_servir` |
| `sesiones_con_lectura_sin_servir` | en cada serie: un entero en las 8 de la 11 y la 12, `null` en las demás | unos 40 por serie | ídem |
| `motivo` y `bash` | unas diez en `boe-legislacion`, lo visto en la #175 y la #178 | unos 75 sin sobre y unos 240 con su orden, que es la que escribió el modelo | cuando el modelo deja de llamar a una herramienta que su sesión no tiene o de filtrar la salida con un programa que el runner no tiene |

Cada job lo calcula desde cero con sus sesiones: ninguna señal pasa al siguiente ni crece con lo acumulado.
