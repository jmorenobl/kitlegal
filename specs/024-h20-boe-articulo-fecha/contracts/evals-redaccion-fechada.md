# Contrato: las evals con fecha y el umbral `redaccion_sin_fechar`

Lo que cambia en `internal/evals`, en `schemas/eval.yaml.json` y en `evals/boe-legislacion/`. Lo demás —el juez, su
rúbrica, sus casos y su medida, las repeticiones, la regla de 2 de 3, la definición del job— no cambia (FR-058, FR-059,
FR-071).

## 1. El formato

### 1.1 El comando de bloque con fecha

```yaml
comandos:
  - applet: boe
    norma: BOE-A-2017-12902
    bloque: a1-30
    fecha: "2019-06-01"
```

En `schemas/eval.yaml.json`, `comandos.items.oneOf` cambia `#/$defs/comando-bloque` por `#/$defs/comando-bloque-esperado`,
que es `comando-bloque` más `fecha` (`#/$defs/dia`) opcional; `grafo_previo.comandos` sigue con `comando-bloque`
(research D11). En Go, `ComandoEsperado.Fecha` (`yaml:"fecha"`); `formaDelComando` sigue dando `formaBloque`, y
`textoDelComando` añade la fecha detrás del bloque cuando la hay: `bloque boe BOE-A-2017-12902 a1-30 2019-06-01`.

**Qué casa con él** (FR-051; research D12). Los argumentos de una invocación de `articulo` o `articulos` se leen
separando `--fecha` —`--fecha=<valor>` o `--fecha <valor>`, en cualquier sitio y antes de `--`— de los de posición; la
norma es el primero de posición, los bloques los demás.

| Comando esperado | Lo cumple | No lo cumple |
|---|---|---|
| sin `fecha` | `articulo` o `articulos` de la norma con el bloque, sin `--fecha`, con código 0 | la misma lectura con `--fecha`, de cualquier fecha |
| con `fecha` *D* | `articulo` de la norma y el bloque con `--fecha` *D*, como orden o como `boe_articulo` con `fecha` *D*, con código 0 o 3 | sin `--fecha`; con otra fecha; con `articulos`; con otro código (2, 4, 5…) |

### 1.2 Lo esperado de una redacción fechada

```yaml
redaccion_fechada:
  norma: BOE-A-2017-12902
  bloque: a1-30
  fecha: "2019-06-01"
  fecha_vigencia: "20180309"
  posterior:
    fecha_vigencia: "20200206"
    norma_modificadora: BOE-A-2020-1651
```

o, sin redacción vigente en la fecha:

```yaml
redaccion_fechada:
  norma: BOE-A-2015-10565
  bloque: a21
  fecha: "2016-06-01"
  vigente_desde: "20161002"
```

Esquema: un objeto con `additionalProperties: false`, `required: [norma, bloque, fecha]` y `oneOf` de
`{required: [fecha_vigencia], not: {required: [vigente_desde]}}` y
`{required: [vigente_desde], not: {anyOf: [{required: [fecha_vigencia]}, {required: [posterior]}]}}`; `posterior` con
`required: [fecha_vigencia, norma_modificadora]`. `redaccion_fechada` entra en las dos listas que prohíben lo esperado
de una respuesta: con `activa: false` y con `sin_binario_ni_servidor` (FR-052).

### 1.3 Las dos evals

`evals/boe-legislacion/22-lcsp-contrato-menor-en-2019.yaml` (eval (a), US1):

```yaml
# Eval positiva con fecha: la pregunta nombra el artículo y el día, y se responde con la redacción vigente ese día,
# con las líneas de la redacción a fecha y de la redacción de hoy (H20 FR-050).
pregunta: "¿qué exigía el artículo 118 de la LCSP para el expediente de un contrato menor el 1 de junio de 2019?"
activa: true
comandos:
  - applet: boe
    norma: BOE-A-2017-12902
    bloque: a1-30
    fecha: "2019-06-01"
citas:
  - norma: BOE-A-2017-12902
    bloque: a1-30
redaccion_fechada:
  norma: BOE-A-2017-12902
  bloque: a1-30
  fecha: "2019-06-01"
  fecha_vigencia: "20180309"
  posterior:
    fecha_vigencia: "20200206"
    norma_modificadora: BOE-A-2020-1651
```

`evals/boe-legislacion/23-lpac-articulo-21-en-2016.yaml` (eval (b), US2):

```yaml
# Eval positiva con fecha: en el día que pregunta, el art. 21 de la Ley 39/2015 no tenía redacción vigente, y la
# respuesta lo dice con su línea y su primera fecha de vigencia (H20 FR-050).
pregunta: "¿qué decía el artículo 21 de la Ley 39/2015 el 1 de junio de 2016?"
activa: true
comandos:
  - applet: boe
    norma: BOE-A-2015-10565
    bloque: a21
    fecha: "2016-06-01"
citas:
  - norma: BOE-A-2015-10565
    bloque: a21
redaccion_fechada:
  norma: BOE-A-2015-10565
  bloque: a21
  fecha: "2016-06-01"
  vigente_desde: "20161002"
```

Ninguna lleva `avisos`: las dos normas no tienen ninguno hoy. Su serie de 2 de 3 es la de hoy: comandos, citas y avisos
(FR-057). La cita de la (b) la lleva la propia línea `⚠ SIN REDACCIÓN VIGENTE:`.

### 1.4 Las reglas del conjunto

Las evals con `redaccion_fechada` se apartan como la eval sin binario ni servidor: no son `positivas` ni entran en
`citadaPor`, así que no las miran «positivas», «materias distintas» ni «normas del hito»; sí «normas conocidas» y el
esquema. `maximoDeEvals` pasa de 21 a 23 (research D15). `normasEsperadas` no lee `redaccion_fechada`.

## 2. Las líneas de la respuesta

Tres formas fijas (research D13), con la gramática de las etiquetas de los avisos (`patronDeEtiqueta`), en
`internal/evals/fechadas.go`:

| Forma | Etiqueta | Lo que se lee de la línea, detrás de su primera cita |
|---|---|---|
| A, la redacción citada (FR-041) | `REDACCIÓN A FECHA` | la primera fecha AAAA-MM-DD y la primera AAAAMMDD |
| B, la redacción de hoy (FR-041) | `REDACCIÓN DE HOY` | el primer identificador `BOE-A-…` y la primera fecha AAAAMMDD |
| C, sin redacción vigente (FR-043) | `SIN REDACCIÓN VIGENTE` | la primera fecha AAAA-MM-DD y la primera AAAAMMDD |

`ExtraerRedaccionesFechadas(respuesta) []LineaFechada` da, por cada línea con una forma y al menos una cita
(`formaDeCita`), su forma, la norma y el bloque de su primera cita y lo leído detrás de ella; una línea sin cita no da
nada. Las fechas AAAAMMDD se leen con `fechasDeVigencia` (`redacciones.go`); las AAAA-MM-DD, como la de `⚠ PLAZO:`.

**Lo que espera una eval con fecha**, sobre las líneas de su norma y su bloque:

- con `fecha_vigencia`: al menos una A, y todas con la fecha y la vigencia esperadas; con `posterior`, al menos una B,
  y todas con su norma y su vigencia; sin `posterior`, ninguna B; ninguna C;
- con `vigente_desde`: al menos una C, y todas con la fecha y la vigencia esperadas; ninguna A ni B.

Las líneas de otros bloques no cuentan.

## 3. El umbral

| Campo | Valor |
|---|---|
| `nombre` | `redaccion_sin_fechar:<modelo>:<modo>`, hoy `redaccion_sin_fechar:claude-sonnet-5-5:orden` y `…:herramienta` |
| `descripcion` | `Respuestas de <modelo> en el modo <modo> cuya sesión no leyó con boe articulo el bloque de la eval en su fecha, o sin las líneas ⚠ REDACCIÓN A FECHA:, ⚠ REDACCIÓN DE HOY: y ⚠ SIN REDACCIÓN VIGENTE: que la eval espera, sobre sus respuestas medidas en las evals que esperan una redacción fechada` |
| `total` | las respuestas del grupo del modo —las del `total` de `sin_activar:<modelo>:<modo>`— de las evals con `redaccion_fechada`: 6 por modo, 2 evals × 3 repeticiones |
| `medida` | las de ellas cuya sesión no tiene ninguna invocación que cumpla el comando `{boe, norma, bloque, fecha}` de su `redaccion_fechada` (§1.1), o cuya respuesta no lleva las líneas que espera (§2); cada sesión, una vez |
| `comparacion`, `umbral`, `decide` | `"<="`, `0`, `true` |

- Fuera del `total` y de la `medida`: las sesiones sin medir, que siguen dando su motivo de la ejecución; las de un
  modelo informativo; y las de la eval sin binario ni servidor (las mismas que `respuestasQueSeJuzgan` deja fuera hoy).
- Existe si alguna eval de la skill lleva `redaccion_fechada`, sin ningún caso por skill (FR-055): solo
  `boe-legislacion` lo tiene.
- **Motivo** cuando no se cumple, el de todo umbral que decide, con cada sesión que cuenta por su nombre, en orden de
  sesión (`nombradasPorSuSesion`):
  `umbral redaccion_sin_fechar:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: <sesión>`. Con él,
  el veredicto es `fallo` y el trabajo de `boe-legislacion` sale en rojo.

**Los quince de `boe-legislacion`**, en este orden; deciden los once marcados:

| # | Nombre | Decide |
|---|---|---|
| 1 | `sin_activar:claude-sonnet-5-5:orden` | sí |
| 2 | `afirma_lo_no_leido:claude-sonnet-5-5:orden` | no |
| 3 | `cuenta_su_proceso:claude-sonnet-5-5:orden` | no |
| 4 | `redaccion_sin_fechar:claude-sonnet-5-5:orden` | sí (nuevo) |
| 5 | `sin_activar:claude-sonnet-5-5:herramienta` | sí |
| 6 | `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | no |
| 7 | `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | no |
| 8 | `redaccion_sin_fechar:claude-sonnet-5-5:herramienta` | sí (nuevo) |
| 9 | `afirma_lo_no_leido:claude-sonnet-5-5` | sí |
| 10 | `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | sí |
| 11 | `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | sí |
| 12 | `duracion_de_las_sesiones:orden` | sí |
| 13 | `duracion_de_las_sesiones:herramienta` | sí |
| 14 | `duracion_del_juez:orden` | sí |
| 15 | `duracion_del_juez:herramienta` | sí |

`legal-core`, sus seis, los seis que deciden; `jurisprudencia`, sus quince, once que deciden: sin cambios (FR-056).
Con las veintitrés evals, `sin_activar` cuenta sobre 60 respuestas por modo y `afirma_lo_no_leido:<modelo>`, sobre 120,
con su umbral de 0,028 (como mucho 3) (FR-058).

## 4. Los tests

**`TestUmbralesDeBoeLegislacion`** (`internal/evals/informe_test.go`, nuevo), con sesiones sintéticas del plan del job
sobre las evals del repositorio (`escribirSesionSinteticaDelPlan`, `informeDelPlanDelJob`, y una lectura fechada con
orden y con herramienta junto a `leerElArticulo`). Los nueve casos de FR-089 que son del umbral, cada uno con los quince
umbrales, sus medidas, el veredicto y los motivos escritos a mano:

| Caso | Sesión | Cuenta |
|---|---|---|
| 1 `lee-sin-la-fecha`, `lee-otra-fecha` | (a), modo orden: lee `a1-30` sin `--fecha`, o con `2019-06-02`, y lleva las dos líneas | sí: 1 de 6, `fallo`, su motivo |
| 2 `lee-con-la-fecha` | (a) en los dos modos, como orden y como `boe_articulo` con `fecha`, con A y B esperadas | no: 0 de 6 en los dos |
| 3 `sin-la-linea`, `con-otra-fecha` | (a): lee con la fecha y no lleva A, o la lleva con `2019-06-02` o con `20200206` | sí |
| 4 `sin-la-de-hoy`, `la-de-hoy-con-otra-fecha`, `la-de-hoy-con-otra-norma` | (a): A esperada y B ausente, con `20200205` o con otro `BOE-A-…` | sí |
| 5 `sin-redaccion-vigente`, `redaccion-a-fecha-en-la-b` | (b): C con `2016-06-01` y `20161002`; o A en su lugar | no; sí |
| 6 `una-eval-sin-fecha` | la 01, que no lleva `redaccion_fechada`, sin ninguna línea | fuera del `total`: sigue en 6 |
| 7 `sin-medir` | (a), sin medir | fuera; su motivo de la ejecución |
| 8 `de-haiku` | (a) con `claude-haiku-4-5-20251001`, sin leer con fecha | fuera |
| 9 (del juicio) | ver `TestJuzgarLecturasFechadas` | — |

**`TestUmbralesDeLasTresSkills`** (`informe_test.go`, cambia): `boe-legislacion`, con sus evals del repositorio, lleva
los quince de §3 con once que deciden (un ayudante nuevo, `umbralesDeBoeLegislacion`, que añade los dos a los de
`umbralesDeBoeLegislacionComoEnH26`); `legal-core` y `jurisprudencia`, como hoy (FR-056, SC-009).

**`TestJuzgarLecturasFechadas`** (`internal/evals/juzgar_test.go`, nuevo): la tabla de §1.1 —comando con fecha y orden
con `--fecha` delante y detrás, con `=` y separada, y con la herramienta; con 0 y con 3; sin fecha, con otra, con
`articulos` y con 4 no—, y el caso 9 de FR-089: el comando sin fecha de la 02 no lo cumple una lectura fechada del
`a1-30`, que deja su motivo `comando ausente: bloque boe BOE-A-2017-12902 a1-30`.

**`TestExtraerRedaccionesFechadas`** (`internal/evals/fechadas_test.go`, nuevo): cada forma, con mayúsculas y énfasis
de Markdown; una línea sin cita; dos líneas; lo leído detrás de la cita y no delante; la etiqueta de un aviso de hoy,
que no da nada.

**`TestFormatoDeLasRedaccionesFechadas`** (`internal/evals/formato_test.go`, nuevo): el esquema admite las dos formas de
§1.2 y `fecha` en `comandos`; rechaza `fecha_vigencia` con `vigente_desde`, `posterior` sin `fecha_vigencia`, una fecha
que no es AAAA-MM-DD, `fecha` en `grafo_previo`, y `redaccion_fechada` en una eval con `activa: false` o sin binario ni
servidor.

**Cambian**, sin dejar de comprobar nada: `TestConjuntoDeEvals` (`conjunto_test.go`), el máximo de 23 y una eval con
fecha que no es positiva; `TestEvalsDelRepositorio`, con las veintitrés evals, su grabado y el subtest nuevo
`formas-de-la-redaccion-fechada` —el `SKILL.md` de `boe-legislacion` lleva, en sus bloques `text`, las tres etiquetas
de §2, como `hallazgos-de-la-skill` con la suya—; `TestConsultasNecesarias` (`consultas_test.go`), con la lectura de
`--fecha 9999-12-31` de cada bloque (research D18); y `TestDefinicionDelJob` (`definicion_test.go`), solo el literal
`peorCasoDeBoeLegislacion` —18 977 s, con 109, 108 y 6 sesiones y 60, 60 y 3 respuestas— y el caso de
`probarElPeorCasoDelTrabajo` que lo lleva (research D16).

## 5. Los mutantes de FR-090

Cada uno en una copia fuera del árbol, de una línea, y retirado al terminar; la tarea apunta en
`specs/024-h20-boe-articulo-fecha/mutantes.md` qué test falla con cada uno.

| Mutante | Dónde | Falla |
|---|---|---|
| La versión elegida por `fecha_publicacion` | la elección de `fecha.go` | `TestArticuloEnUnaFecha` (la víspera del 2008-04-30, el 1985-04-22 y el `a21` el 2016-06-01) |
| Siempre la última | ídem | `TestArticuloEnUnaFecha` (la intermedia, el día y la víspera) |
| Con una fecha anterior al bloque, la primera versión | ídem | `TestArticuloEnUnaFecha` (el 1985-04-22 y el `a21`) |
| La lectura fechada contada como lectura: las operaciones de la lectura sin fecha, con `eli:has_version` | `observadoDeLaLecturaFechada` | `TestObservadoDeBoe` y `TestLecturaFechadaEnElGrafo` |
| El `decide` del umbral en falso | `umbralDeRedaccionSinFechar` | `TestUmbralesDeBoeLegislacion` y `TestUmbralesDeLasTresSkills` |

## 6. Uso, de fuera adentro

| Salida | Quién la lee | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| Los dos elementos nuevos de `umbrales` | el job, que decide con ellos; el informe final, sin modelo (SC-002); la persona y la reparación del cierre | unos 450 bytes cada uno con su descripción; parten de las evals del repositorio y no crecen con el uso | se miden de nuevo en cada job; el `cumple: false` de un modo deja de darse en el primer job en que ninguna sesión cuenta |
| Su motivo de `fallo` | la persona y la reparación | unos 90 bytes y unos 55 por sesión, como mucho las 6 del modo: menos de 0,5 KB | con el umbral cumplido |
| Las líneas A, B y C en la respuesta | la persona; el job, que las compara sin modelo | una línea cada una | solo en respuestas a preguntas con fecha; ninguna sin fecha (FR-046) |
