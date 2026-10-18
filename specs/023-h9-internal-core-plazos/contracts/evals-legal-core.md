# Contrato de las evals de `legal-core` v1

Formato de eval, juicio sin modelo de las tres formas fijas, umbrales y las cinco evals (FR-048, FR-050 a FR-057).
Decisiones: research D18 a D20, D26.

## 1. El formato gana dos esperados y una forma de comando (FR-051)

| Clave | Tipo | Quién la admite | Qué espera |
|---|---|---|---|
| `plazo.fecha` | cadena `AAAA-MM-DD` | una eval activa que no es sin binario ni servidor | la línea `⚠ PLAZO:` con esa fecha (§2) |
| `plazo.avisos` | lista de códigos de `plazos.CodigosDeAviso`, sin repetir, puede ir vacía | la misma | cada uno, con su línea `⚠ <ETIQUETA>:` en la respuesta |
| `plazo.razonamiento` | cadena no vacía | la misma | nada: es el cálculo a mano, independiente del código; no lo lee el juicio |
| `no_cubierto.norma` | `BOE-A-…` | la misma, nunca con `plazo` | la línea `⚠ NO CUBIERTO:` con ese identificador |

Comando `calcular`: `{applet: plazos, verbo: calcular, tipo: <código>}`, en el `oneOf` de los comandos. Lo satisface
una orden `plazos calcular <tipo> …` que termina con 0 sin `--dry-run`, `--describe` ni la ayuda, o una llamada a
`plazos_calcular` con ese `tipo` cuyo resultado no es un error. No pide ninguna consulta grabada
(`ConsultasNecesarias`: el applet no consulta nada). `prohibidos` admite ya `{applet: plazos, verbo: calcular}`, que
cubre la orden y la llamada (research V24).

`schemas/eval.yaml.json` gana las dos claves y la forma, con el enumerado de los códigos de `plazo.avisos`; un test los
ata a `plazos.CodigosDeAviso` como `ComprobarCodigosDeAviso` ata los de vigencia. **Eval de plazo**: la que lleva
`plazo`; **eval de cobertura**: la que lleva `no_cubierto`.

## 2. Las formas fijas y su juicio (FR-048, FR-057)

| Forma | Texto en `SKILL.md` | Qué lee el juicio |
|---|---|---|
| Línea del plazo | `⚠ PLAZO: <fecha AAAA-MM-DD> — <tipo, como lo nombra references/recursos_y_plazos.md> [<norma legible>, <precepto>, <identificador>, bloque <bloque>]` | la etiqueta `PLAZO` con la tolerancia de `patronDeEtiqueta`, la fecha justo detrás de los dos puntos y al menos una cita con `formaDeCita` en la misma línea |
| Declaración de no cubierto | `⚠ NO CUBIERTO: <procedimiento, como lo nombra references/cobertura_por_materia.md> se rige por <norma> (<identificador>), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (<precepto de la LPAC>).` y, sin norma, `⚠ NO CUBIERTO: <procedimiento> lo regula aparte la <precepto de la LPAC> (BOE-A-2015-10565): kitlegal no calcula sus plazos ni sus recursos.` | la etiqueta `NO CUBIERTO` y cada `BOE-A-…` de la misma línea |
| Avisos | la línea `texto` de cada aviso de `data`, tal cual (contracts/applet-plazos.md §5) | la etiqueta de cada código |

Juicio de una sesión (`cumpleLoEsperado`; cuenta en la regla de 2 de 3 de su serie):

- eval de plazo: falta la línea del plazo con `plazo.fecha` → motivo `plazo ausente: <fecha>`; falta la línea de un
  código de `plazo.avisos` → motivo `aviso del plazo ausente: <código>`;
- eval de cobertura: falta la declaración con `no_cubierto.norma` → motivo `declaración de no cubierto ausente:
  <norma>`; hay una línea del plazo en su forma fija → motivo `línea del plazo en un asunto no cubierto: <fecha>`;
  la ejecución de `plazos calcular` o de `plazos_calcular` está en `prohibidos` y da su motivo de siempre.

## 3. Los umbrales (FR-053 a FR-056; research D20)

| Nombre | Total | Medida | Umbral |
|---|---|---|---|
| `plazo_sin_fecha:<modelo>:<modo>` | las respuestas medidas del modelo que decide en ese modo —las del total de `sin_activar` de ese modo— de las evals de plazo | las que no llevan la línea del plazo con `plazo.fecha`, una vez cada sesión | `<=` 0, `decide: true` |
| `cobertura_sin_declarar:<modelo>:<modo>` | las mismas, de las evals de cobertura | las que no llevan la declaración con `no_cubierto.norma` o llevan una línea del plazo con cualquier fecha, una vez cada sesión | `<=` 0, `decide: true` |

- Descripciones: «Respuestas de %s en el modo %s sin la línea ⚠ PLAZO: con la fecha esperada, sobre sus respuestas
  medidas en las evals que esperan un plazo» y «Respuestas de %s en el modo %s sin la línea ⚠ NO CUBIERTO: con la norma
  esperada, o con una línea ⚠ PLAZO:, sobre sus respuestas medidas en las evals que esperan una declaración de no
  cubierto».
- Motivo de uno que no se cumple: `umbral <nombre>: <medida> de <total> (<p> %), y tiene que ser ≤ 0,0 %: <sesión>;
  <sesión>…`, cada sesión que cuenta por su nombre, en orden de sesión. Uno que se cumple no da motivo.
- Fuera de su total y de su medida: las sesiones sin medir (que dan su motivo de la ejecución), las de un modelo
  informativo y las de la eval sin binario ni servidor. Los avisos no entran en `plazo_sin_fecha` (FR-055).
- Orden en `umbrales`: por cada modo, `sin_activar`, las clases del juez, `cita_sin_documento`,
  `documento_sin_cotejar`, `plazo_sin_fecha` y `cobertura_sin_declarar`, cada uno si la skill lo tiene. Una skill los
  tiene si alguna eval declara `plazo` o `no_cubierto`; los de las respuestas existen si tiene juez o alguna eval
  declara `sentencias`, `plazo` o `no_cubierto`.

Con el plan del job (FR-056; SC-011):

| Skill | `umbrales` | Deciden |
|---|---|---|
| `legal-core` | `sin_activar`, `plazo_sin_fecha` y `cobertura_sin_declarar` del modo orden; los mismos del modo herramienta | los seis |
| `boe-legislacion` | los trece de hoy | nueve |
| `jurisprudencia` | los quince de hoy | once |

Ejemplo del elemento, con `--json` del informe (FR-053; el total del modo con tres repeticiones):

```json
{"nombre":"plazo_sin_fecha:claude-sonnet-5-5:orden","descripcion":"Respuestas de claude-sonnet-5-5 en el modo orden sin la línea ⚠ PLAZO: con la fecha esperada, sobre sus respuestas medidas en las evals que esperan un plazo","medida":0,"total":9,"comparacion":"<=","umbral":0,"cumple":true,"decide":true}
```

## 4. Las reglas del conjunto de `legal-core` (FR-051, FR-088)

- «esperado verificable»: una eval activa que no es sin binario ni servidor lleva `citas`, `territorio`, `plazo` o
  `no_cubierto`.
- «cobertura por materia», nueva: al menos una eval de plazo y una de cobertura (Definition of Done §1.13; constitución,
  «Gates», capa 1).
- Las demás, como hoy. `ConsultasNecesarias` de las nueve evals las sirve la unión de grabaciones sin red
  (`TestEvalsDelRepositorio`, subtest de `legal-core`; FR-052).

## 5. Las cinco evals (FR-050)

Las cuatro de hoy no cambian. Las razones a mano, en contracts/casos-de-plazos.md.

| Fichero | Clase | Pregunta | Comandos | Citas | Esperado | Prohibidos |
|---|---|---|---|---|---|---|
| `05-plazo-reposicion-leganes.yaml` | plazo (a) | «Me notificaron el 9 de septiembre de 2026 una resolución del Ayuntamiento de Leganés y vivo en Leganés, ¿hasta cuándo puedo presentar el recurso de reposición?» | `territorio resolver Leganés`; `plazos calcular` `reposicion`; bloque `BOE-A-2015-10565` `a124` | `BOE-A-2015-10565` `a124` | `plazo: {fecha: "2026-10-13", avisos: []}` | — |
| `06-plazo-reposicion-tordesillas.yaml` | plazo (b) | «Me notificaron el 9 de septiembre de 2026 una resolución del Ayuntamiento de Tordesillas y vivo en Tordesillas, ¿hasta cuándo puedo presentar el recurso de reposición?» | `territorio resolver Tordesillas`; `plazos calcular` `reposicion`; bloque `a124` | `a124` | `plazo: {fecha: "2026-10-09", avisos: [faltan-festivos-locales]}` | — |
| `07-cobertura-liquidacion-tributaria.yaml` | cobertura (c) | «Vivo en Leganés y el 9 de septiembre de 2026 me notificaron una liquidación de la Agencia Tributaria. ¿Hasta cuándo puedo recurrirla?» | ninguno | ninguna | `no_cubierto: {norma: BOE-A-2003-23186}` | `plazos calcular` |
| `08-cobertura-multa-de-trafico.yaml` | cobertura (d) | «Vivo en Leganés y el 9 de septiembre de 2026 me notificaron una multa de tráfico de la Dirección General de Tráfico. ¿Hasta cuándo puedo recurrirla?» | ninguno | ninguna | `no_cubierto: {norma: <el identificador del texto refundido de Tráfico de data/cobertura.yaml>}` | `plazos calcular` |
| `09-plazo-contencioso-agosto.yaml` | plazo (e) | «El 15 de junio de 2026 el Ayuntamiento de Leganés me notificó la desestimación de mi recurso de reposición. Quiero interponer recurso contencioso-administrativo ante los juzgados de lo contencioso-administrativo de Madrid. ¿Hasta qué día puedo hacerlo?» | `territorio resolver Madrid`; `plazos calcular` `contencioso`; bloques `BOE-A-1998-16718` `a46` y `a128` | `a46` y `a128` | `plazo: {fecha: "2026-09-15", avisos: []}` | — |

Ninguna declara `no_se_activan`: el precepto lo lee la skill con `boe-legislacion` (research D21). Ninguna es
informativa ni por materia.

Con nueve evals, el plan del job abre en `legal-core` 48, 48 y 6 sesiones (dos modelos y tres repeticiones), más la
prueba de red en el modo orden, y su peor caso a dos sesiones es de 14 629 s, dentro de los 352 minutos (research V15,
M5). Por modo,
el modelo que decide da 21 respuestas medidas en el total de `sin_activar`, 9 de las evals de plazo y 6 de las de
cobertura.

## 6. Uso, de fuera adentro

| Salida | Quién, cuántas veces | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| Los seis umbrales de `legal-core` en `umbrales` | el job, una vez; el informe final, sin modelo; la persona | unos 300 bytes cada elemento; un motivo, unos 90 bytes más unos 55 por sesión, acotado por las 9 o 6 sesiones del modo: menos de 0,6 KB | cada job mide de nuevo; el `cumple: false` y su motivo dejan de darse en el primer job en que ninguna sesión cuenta |
| Los motivos de una sesión en su resultado | el informe, una vez por sesión | una línea por esperado que falta | se miden en cada job |
